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
	logger            *log.Logger
	workerPool        *WorkerPool
	replicationSvc    ReplicationService
	registryProviders map[string]interfaces.RegistryProvider
	cronParser        cron.Parser
	encryptionMgr     *encryption.Manager
}

// SchedulerOptions provides configuration for the scheduler
type SchedulerOptions struct {
	// Logger is the logger to use
	Logger *log.Logger

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
		logger = log.NewLogger(log.InfoLevel)
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
		s.logger.Debug("Skipping job without schedule", map[string]interface{}{
			"id": id,
		})
		return nil
	}

	// Pre-validate the schedule if it's not a special case
	if rule.Schedule != "@now" {
		_, err := s.cronParser.Parse(rule.Schedule)
		if err != nil {
			return errors.Wrap(err, "invalid cron expression: %s", rule.Schedule)
		}
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Parse the schedule as a cron expression
	var nextRun time.Time

	if rule.Schedule == "@now" {
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

			// Calculate the next run time based on the cron expression
			if job.Rule.Schedule != "@once" && job.Rule.Schedule != "@now" {
				schedule, err := s.cronParser.Parse(job.Rule.Schedule)
				if err != nil {
					s.logger.Warn("Invalid cron expression, using default schedule", map[string]interface{}{
						"id":       id,
						"schedule": job.Rule.Schedule,
						"error":    err.Error(),
						"next_run": now.Add(1 * time.Hour),
					})
					job.NextRun = now.Add(1 * time.Hour)
				} else {
					job.NextRun = schedule.Next(now)
					s.logger.Debug("Scheduled next run", map[string]interface{}{
						"id":       id,
						"next_run": job.NextRun,
					})
				}
			} else {
				// For @once and @now schedules, don't reschedule
				s.logger.Debug("One-time job, not rescheduling", map[string]interface{}{
					"id": id,
				})
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
		s.logger.Error("Invalid job submission", err, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Verify worker pool is initialized
	if s.workerPool == nil {
		err := errors.InvalidInputf("worker pool not initialized")
		s.logger.Error("Invalid worker pool", err, map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})

		// Mark the job as not running since submission will fail
		s.mutex.Lock()
		if j, exists := s.jobs[id]; exists {
			j.Running = false
		}
		s.mutex.Unlock()
		return
	}

	s.logger.Info("Running scheduled job", map[string]interface{}{
		"id": id,
	})

	// Submit the job to the worker pool
	err := s.workerPool.Submit(id, func(ctx context.Context) error {
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

		// Check if we have a replication service
		if s.replicationSvc == nil {
			return errors.InvalidInputf("replication service not configured")
		}

		// Log job start
		s.logger.Info("Starting replication job", map[string]interface{}{
			"id":                id,
			"source_registry":   job.Rule.SourceRegistry,
			"source_repository": job.Rule.SourceRepository,
			"dest_registry":     job.Rule.DestinationRegistry,
			"dest_repository":   job.Rule.DestinationRepository,
			"include_tags":      job.Rule.IncludeTags,
			"exclude_tags":      job.Rule.ExcludeTags,
			"force_overwrite":   job.Rule.ForceOverwrite,
		})

		// Execute the replication using the service
		startTime := time.Now()
		err := s.replicationSvc.ReplicateRepository(ctx, job.Rule)
		duration := time.Since(startTime)

		if err != nil {
			replicationErr := errors.Wrap(err, "replication failed")
			s.logger.Error("Replication job failed", replicationErr, map[string]interface{}{
				"id":       id,
				"duration": duration.String(),
				"error":    err.Error(),
			})
			return replicationErr
		}

		// Log job completion
		s.logger.Info("Completed replication job", map[string]interface{}{
			"id":       id,
			"duration": duration.String(),
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

	s.logger.Info("Stopping scheduler", nil)
	s.cancelFn()
	return nil
}
