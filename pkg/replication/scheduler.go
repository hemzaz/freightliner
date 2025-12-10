package replication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"
	"freightliner/pkg/security/encryption"

	"github.com/robfig/cron/v3"
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

// ReplicationService handles the actual replication operations
type ReplicationService interface {
	// ReplicateRepository replicates a repository according to the given rule
	ReplicateRepository(ctx context.Context, rule ReplicationRule) error
}

// Scheduler manages scheduled replication jobs
type Scheduler struct {
	jobs              map[string]*Job
	mutex             sync.RWMutex
	ctx               context.Context
	cancelFn          context.CancelFunc
	logger            log.Logger
	workerPool        *WorkerPool
	replicationSvc    ReplicationService
	registryProviders map[string]interfaces.RegistryProvider
	cronParser        cron.Parser
	encryptionMgr     *encryption.Manager
}

// SchedulerOptions provides configuration for the scheduler
type SchedulerOptions struct {
	// Logger is the logger to use
	Logger log.Logger

	// WorkerPool is the worker pool for executing jobs
	WorkerPool *WorkerPool

	// RegistryProviders is a map of registry providers by type (e.g., "ecr", "gcr")
	RegistryProviders map[string]interfaces.RegistryProvider

	// ReplicationService is the service that handles actual replication
	ReplicationService ReplicationService

	// EncryptionManager is the manager for encryption operations (optional)
	EncryptionManager *encryption.Manager
}

// NewScheduler creates a new replication scheduler
func NewScheduler(opts SchedulerOptions) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	// Create default logger if not provided
	logger := opts.Logger
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Configure cron parser with seconds field
	cronParser := cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	scheduler := &Scheduler{
		jobs:              make(map[string]*Job),
		ctx:               ctx,
		cancelFn:          cancel,
		logger:            logger,
		workerPool:        opts.WorkerPool,
		replicationSvc:    opts.ReplicationService,
		registryProviders: opts.RegistryProviders,
		cronParser:        cronParser,
		encryptionMgr:     opts.EncryptionManager,
	}

	// Start the scheduler loop
	go scheduler.run()

	return scheduler
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(rule ReplicationRule) error {
	// Validate input before locking to fail fast
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

	// Create a unique ID for the job
	id := rule.SourceRegistry + "/" + rule.SourceRepository + " -> " +
		rule.DestinationRegistry + "/" + rule.DestinationRepository

	// Skip jobs without a schedule
	if rule.Schedule == "" {
		s.logger.WithFields(map[string]interface{}{
			"id": id,
		}).Debug("Skipping job without schedule")
		return nil
	}

	// Pre-validate the schedule if it's not a special case
	if rule.Schedule != "@now" && rule.Schedule != "@once" {
		_, err := s.cronParser.Parse(rule.Schedule)
		if err != nil {
			return errors.Wrap(err, "invalid cron expression: %s", rule.Schedule)
		}
	}

	s.mutex.Lock()
	// Note: Unlock is done explicitly before returning (no defer)

	// Parse the schedule as a cron expression
	var nextRun time.Time

	if rule.Schedule == "@now" || rule.Schedule == "@once" {
		// Special case for immediate execution
		nextRun = time.Now()
	} else {
		schedule, err := s.cronParser.Parse(rule.Schedule)
		if err != nil {
			// This should never happen as we've already validated the schedule
			return errors.Wrap(err, "invalid cron expression: %s", rule.Schedule)
		}

		// Calculate the next run time based on the schedule
		nextRun = schedule.Next(time.Now())
	}

	// Check if job already exists
	if _, exists := s.jobs[id]; exists {
		s.logger.WithFields(map[string]interface{}{
			"id": id,
		}).Debug("Updating existing job")
	}

	s.jobs[id] = &Job{
		Rule:    rule,
		NextRun: nextRun,
		Running: false,
	}

	s.logger.WithFields(map[string]interface{}{
		"id":       id,
		"next_run": nextRun,
	}).Info("Added scheduled job")

	// If this is an immediate execution job, trigger a check after releasing lock
	triggerImmediate := rule.Schedule == "@now" || rule.Schedule == "@once"

	// Release lock before spawning goroutine
	s.mutex.Unlock()

	// Trigger immediate check if needed (outside critical section)
	if triggerImmediate {
		go func() {
			time.Sleep(10 * time.Millisecond) // Small delay to ensure job is registered
			s.checkJobs()
		}()
	}

	return nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(rule ReplicationRule) error {
	// Validate input before locking to fail fast
	if rule.SourceRegistry == "" || rule.SourceRepository == "" ||
		rule.DestinationRegistry == "" || rule.DestinationRepository == "" {
		return errors.InvalidInputf("all registry and repository fields must be provided")
	}

	// Create a unique ID for the job
	id := rule.SourceRegistry + "/" + rule.SourceRepository + " -> " +
		rule.DestinationRegistry + "/" + rule.DestinationRepository

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.jobs[id]; exists {
		delete(s.jobs, id)

		s.logger.WithFields(map[string]interface{}{
			"id": id,
		}).Info("Removed scheduled job")
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
		// Check if job should run (not running and next run time has passed or is now)
		if !job.Running && (now.After(job.NextRun) || now.Equal(job.NextRun)) {
			// Mark the job as running
			job.Running = true

			// Calculate the next run time based on the cron expression
			if job.Rule.Schedule != "@once" && job.Rule.Schedule != "@now" {
				schedule, err := s.cronParser.Parse(job.Rule.Schedule)
				if err != nil {
					s.logger.WithFields(map[string]interface{}{
						"id":       id,
						"schedule": job.Rule.Schedule,
						"error":    err.Error(),
						"next_run": now.Add(1 * time.Hour),
					}).Warn("Invalid cron expression, using default schedule")
					job.NextRun = now.Add(1 * time.Hour)
				} else {
					job.NextRun = schedule.Next(now)
					s.logger.WithFields(map[string]interface{}{
						"id":       id,
						"next_run": job.NextRun,
					}).Debug("Scheduled next run")
				}
			} else {
				// For @once and @now schedules, don't reschedule
				s.logger.WithFields(map[string]interface{}{
					"id": id,
				}).Debug("One-time job, not rescheduling")
			}

			// Submit the job to the worker pool
			s.submitJob(id, job)
		}
	}
}

// submitJob submits a job to the worker pool
func (s *Scheduler) submitJob(id string, job *Job) {
	// Validate input before any processing or locking
	if id == "" || job == nil {
		err := errors.InvalidInputf("job ID empty or job is nil")
		s.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Invalid job submission", err)
		return
	}

	// Verify worker pool is initialized
	if s.workerPool == nil {
		err := errors.InvalidInputf("worker pool not initialized")
		s.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}).Error("Invalid worker pool", err)

		// Mark the job as not running since submission will fail
		s.mutex.Lock()
		if j, exists := s.jobs[id]; exists {
			j.Running = false
		}
		s.mutex.Unlock()
		return
	}

	s.logger.WithFields(map[string]interface{}{
		"id": id,
	}).Info("Running scheduled job")

	// Submit the job to the worker pool
	err := s.workerPool.Submit(id, func(ctx context.Context) error {
		defer func() {
			// Recover from panics
			if r := recover(); r != nil {
				panicErr := errors.New(fmt.Sprintf("job panic: %v", r))
				s.logger.WithFields(map[string]interface{}{
					"id":    id,
					"panic": fmt.Sprintf("%v", r),
				}).Error("Job panic recovered", panicErr)
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

		// Check if we have a replication service
		if s.replicationSvc == nil {
			return errors.InvalidInputf("replication service not configured")
		}

		// Log job start
		s.logger.WithFields(map[string]interface{}{
			"id":                id,
			"source_registry":   job.Rule.SourceRegistry,
			"source_repository": job.Rule.SourceRepository,
			"dest_registry":     job.Rule.DestinationRegistry,
			"dest_repository":   job.Rule.DestinationRepository,
			"include_tags":      job.Rule.IncludeTags,
			"exclude_tags":      job.Rule.ExcludeTags,
			"force_overwrite":   job.Rule.ForceOverwrite,
		}).Info("Starting replication job")

		// Execute the replication using the service
		startTime := time.Now()
		err := s.replicationSvc.ReplicateRepository(ctx, job.Rule)
		duration := time.Since(startTime)

		if err != nil {
			replicationErr := errors.Wrap(err, "replication failed")
			s.logger.WithFields(map[string]interface{}{
				"id":       id,
				"duration": duration.String(),
				"error":    err.Error(),
			}).Error("Replication job failed", replicationErr)
			return replicationErr
		}

		// Log job completion
		s.logger.WithFields(map[string]interface{}{
			"id":       id,
			"duration": duration.String(),
		}).Info("Completed replication job")

		return nil
	})

	if err != nil {
		submitErr := errors.Wrap(err, "failed to submit job to worker pool")
		s.logger.WithFields(map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		}).Error("Failed to submit job", submitErr)

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
	// Check context before locking
	if s.ctx.Err() != nil {
		return errors.AlreadyExistsf("scheduler already stopped")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Double-check context after acquiring lock to prevent race condition
	if s.ctx.Err() != nil {
		return errors.AlreadyExistsf("scheduler already stopped")
	}

	s.logger.Info("Stopping scheduler")
	s.cancelFn()
	return nil
}
