package replication

import (
	"context"
	"sync"
	"time"

	"github.com/hemzaz/freightliner/src/internal/log"
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
func (s *Scheduler) AddJob(rule ReplicationRule) {
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
		return
	}

	// TODO: Parse the schedule as a cron expression
	// For now, just schedule it to run in 5 minutes
	nextRun := time.Now().Add(5 * time.Minute)

	s.jobs[id] = &Job{
		Rule:    rule,
		NextRun: nextRun,
		Running: false,
	}

	s.logger.Info("Added scheduled job", map[string]interface{}{
		"id":       id,
		"next_run": nextRun,
	})
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(rule ReplicationRule) {
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
	}
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
	s.logger.Info("Running scheduled job", map[string]interface{}{
		"id": id,
	})

	// Create a copy of the job's rule for the worker
	rule := job.Rule

	// Submit the job to the worker pool
	s.workerPool.Submit(func(ctx context.Context) error {
		defer func() {
			// Mark the job as not running when done
			s.mutex.Lock()
			if j, exists := s.jobs[id]; exists {
				j.Running = false
			}
			s.mutex.Unlock()
		}()

		// TODO: Implement the actual replication logic
		// This would call into a ReplicationService or similar

		s.logger.Info("Completed scheduled job", map[string]interface{}{
			"id": id,
		})

		return nil
	})
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.logger.Info("Stopping scheduler", nil)
	s.cancelFn()
}
