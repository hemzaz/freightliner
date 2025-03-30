package replication

import (
	"context"
	"fmt"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"sync"
	"time"
)

// Job represents a scheduled replication job
type Job struct {
	// Rule is the replication rule to apply
	Rule ReplicationRule

	// NextRun is the next time the job should run
	NextRun time.Time

	// Running indicates if the job is currently running
	Running bool
}

// Scheduler manages scheduled replication jobs
type Scheduler struct {
	jobs       map[string]*Job
	mutex      sync.RWMutex
	ctx        context.Context
	cancelFn   context.CancelFunc
	logger     *log.Logger
	workerPool *WorkerPool
}

// NewScheduler creates a new replication scheduler
func NewScheduler(logger *log.Logger, workerPool *WorkerPool) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &Scheduler{
		jobs:       make(map[string]*Job),
		ctx:        ctx,
		cancelFn:   cancel,
		logger:     logger,
		workerPool: workerPool,
	}

	// Start the scheduler loop
	go scheduler.run()

	return scheduler
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(rule ReplicationRule) error {
	if rule.SourceRegistry == "" {
		return errors.InvalidInputf("source registry cannot be empty")
	}

	if rule.SourceRepository == "" {
		return errors.InvalidInputf("source repository cannot be empty")
	}

	if rule.DestinationRegistry == "" {
		return errors.InvalidInputf("destination registry cannot be empty")
	}

	if rule.DestinationRepository == "" {
		return errors.InvalidInputf("destination repository cannot be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create a unique ID for the job
	id := rule.SourceRegistry + "/" + rule.SourceRepository + " -> " +
		rule.DestinationRegistry + "/" + rule.DestinationRepository

	// Skip jobs without a schedule
	if rule.Schedule == "" {
		s.logger.Debug("Skipping job without schedule", map[string]interface{}{
			"id": id,
		})
		return nil
	}

	// TODO: Parse the schedule as a cron expression
	// For now, just schedule it to run in 5 minutes
	nextRun := time.Now().Add(5 * time.Minute)

	// Check if job already exists
	if _, exists := s.jobs[id]; exists {
		s.logger.Debug("Updating existing job", map[string]interface{}{
			"id": id,
		})
	}

	s.jobs[id] = &Job{
		Rule:    rule,
		NextRun: nextRun,
		Running: false,
	}

	s.logger.Info("Added scheduled job", map[string]interface{}{
		"id":       id,
		"next_run": nextRun,
	})

	return nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(rule ReplicationRule) error {
	if rule.SourceRegistry == "" || rule.SourceRepository == "" ||
		rule.DestinationRegistry == "" || rule.DestinationRepository == "" {
		return errors.InvalidInputf("all registry and repository fields must be provided")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create a unique ID for the job
	id := rule.SourceRegistry + "/" + rule.SourceRepository + " -> " +
		rule.DestinationRegistry + "/" + rule.DestinationRepository

	if _, exists := s.jobs[id]; exists {
		delete(s.jobs, id)

		s.logger.Info("Removed scheduled job", map[string]interface{}{
			"id": id,
		})
		return nil
	}

	return errors.NotFoundf("job not found with ID: %s", id)
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkJobs()
		case <-s.ctx.Done():
			return
		}
	}
}

// checkJobs checks for jobs that need to run
func (s *Scheduler) checkJobs() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()

	for id, job := range s.jobs {
		if !job.Running && now.After(job.NextRun) {
			// Mark the job as running
			job.Running = true

			// TODO: Calculate the next run time based on the cron expression
			// For now, just schedule it to run again in 1 hour
			job.NextRun = now.Add(1 * time.Hour)

			// Submit the job to the worker pool
			s.submitJob(id, job)
		}
	}
}

// submitJob submits a job to the worker pool
func (s *Scheduler) submitJob(id string, job *Job) {
	if id == "" || job == nil {
		err := errors.InvalidInputf("job ID empty or job is nil")
		s.logger.Error("Invalid job submission", err, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	s.logger.Info("Running scheduled job", map[string]interface{}{
		"id": id,
	})

	// Submit the job to the worker pool
	err := s.workerPool.Submit(func(ctx context.Context) error {
		defer func() {
			// Recover from panics
			if r := recover(); r != nil {
				panicErr := errors.New(fmt.Sprintf("job panic: %v", r))
				s.logger.Error("Job panic recovered", panicErr, map[string]interface{}{
					"id":    id,
					"panic": fmt.Sprintf("%v", r),
				})
			}

			// Mark the job as not running when done
			s.mutex.Lock()
			if j, exists := s.jobs[id]; exists {
				j.Running = false
			}
			s.mutex.Unlock()
		}()

		// Check for context cancellation
		if ctx.Err() != nil {
			return errors.Wrap(ctx.Err(), "job context canceled")
		}

		// TODO: Implement the actual replication logic
		// This would call into a ReplicationService or similar

		s.logger.Info("Completed scheduled job", map[string]interface{}{
			"id": id,
		})

		return nil
	})

	if err != nil {
		submitErr := errors.Wrap(err, "failed to submit job to worker pool")
		s.logger.Error("Failed to submit job", submitErr, map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})

		// Mark the job as not running since submission failed
		s.mutex.Lock()
		if j, exists := s.jobs[id]; exists {
			j.Running = false
		}
		s.mutex.Unlock()
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.ctx.Err() != nil {
		return errors.AlreadyExistsf("scheduler already stopped")
	}

	s.logger.Info("Stopping scheduler", nil)
	s.cancelFn()
	return nil
}
