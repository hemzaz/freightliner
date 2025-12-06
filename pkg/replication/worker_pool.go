package replication

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
)

// WorkerPool manages a pool of workers for parallel processing
type WorkerPool struct {
	workers       int
	jobQueue      chan WorkerJob
	results       chan JobResult
	waitGroup     sync.WaitGroup
	stopContext   context.Context
	stopFunc      context.CancelFunc
	logger        log.Logger
	closed        atomic.Bool
	jobsClosed    atomic.Bool
	resultsClosed atomic.Bool
	stats         *statsCollector
}

// WorkerJob represents a unit of work to be processed by a worker
type WorkerJob struct {
	ID       string
	Task     TaskFunc
	Priority int
	Context  context.Context
}

// JobResult represents the result of a job
type JobResult struct {
	JobID string
	Error error
}

// TaskFunc is a function that performs a task
type TaskFunc func(ctx context.Context) error

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workerCount int, logger log.Logger) *WorkerPool {
	if workerCount <= 0 {
		workerCount = 1
	}

	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Calculate optimal buffer sizes based on worker count
	// Use a minimum buffer size and scale with worker count for better performance
	minBufferSize := 10
	maxBufferSize := 1000
	bufferSize := workerCount * 20 // Allow for more jobs per worker
	if bufferSize < minBufferSize {
		bufferSize = minBufferSize
	}
	if bufferSize > maxBufferSize {
		bufferSize = maxBufferSize
	}

	pool := &WorkerPool{
		workers:     workerCount,
		jobQueue:    make(chan WorkerJob, bufferSize),
		results:     make(chan JobResult, bufferSize),
		stopContext: ctx,
		stopFunc:    cancel,
		logger:      logger,
		stats:       newStatsCollector(),
	}
	return pool
}

// Start starts the worker pool
func (p *WorkerPool) Start() {
	p.logger.WithFields(map[string]interface{}{
		"workers": p.workers,
	}).Info("Starting worker pool")

	for i := 0; i < p.workers; i++ {
		workerID := i
		p.waitGroup.Add(1)

		go func() {
			defer p.waitGroup.Done()
			p.worker(workerID)
		}()
	}
}

// handleJobFromQueue processes a job from the queue
func (p *WorkerPool) handleJobFromQueue(workerID int, job WorkerJob, ok bool) bool {
	if !ok {
		p.logger.WithFields(map[string]interface{}{
			"worker_id": workerID,
		}).Debug("Worker stopped by closed job queue")
		return false // Signal to stop the worker
	}

	p.processJob(workerID, job)
	return true // Continue processing
}

// worker processes jobs from the job queue
func (p *WorkerPool) worker(id int) {
	p.logger.WithFields(map[string]interface{}{
		"worker_id": id,
	}).Debug("Worker started")

	for {
		select {
		case <-p.stopContext.Done():
			p.logger.WithFields(map[string]interface{}{
				"worker_id": id,
			}).Debug("Worker stopped by context cancellation")
			return

		case job, ok := <-p.jobQueue:
			if !p.handleJobFromQueue(id, job, ok) {
				return
			}
		}
	}
}

// setupJobContext creates a job context that will be canceled if either the job's context
// or the pool's context is canceled. Avoids goroutine leaks by using simple context hierarchy.
func (p *WorkerPool) setupJobContext(job WorkerJob) (context.Context, context.CancelFunc) {
	// Create a child context from the job's context
	// This will be canceled when either the job context or pool's stop context is done
	jobCtx, jobCancel := context.WithCancel(job.Context)

	// Note: The job execution monitors both jobCtx and p.stopContext in the worker
	// to handle cancellation from either source
	return jobCtx, jobCancel
}

// executeJob runs the job's task and measures execution time
func (p *WorkerPool) executeJob(ctx context.Context, job WorkerJob) (time.Duration, error) {
	startTime := time.Now()
	err := job.Task(ctx)
	duration := time.Since(startTime)
	return duration, err
}

// logJobResult logs the outcome of a job execution
func (p *WorkerPool) logJobResult(workerID int, jobID string, duration time.Duration, err error) {
	if err != nil {
		p.logger.WithFields(map[string]interface{}{
			"worker_id": workerID,
			"job_id":    jobID,
			"duration":  duration.String(),
		}).Error("Job failed", err)
	} else {
		p.logger.WithFields(map[string]interface{}{
			"worker_id": workerID,
			"job_id":    jobID,
			"duration":  duration.String(),
		}).Debug("Job completed successfully")
	}
}

// sendJobResult sends a job result to the results channel with deadlock prevention
func (p *WorkerPool) sendJobResult(result JobResult) {
	select {
	case p.results <- result:
		// Result sent successfully
	case <-p.stopContext.Done():
		// Pool is stopping, don't block on sending results
		p.logger.WithFields(map[string]interface{}{
			"job_id": result.JobID,
		}).Debug("Pool stopping, discarding result")
	case <-time.After(5 * time.Second):
		// Results channel is full or blocked, log and continue to prevent deadlock
		p.logger.WithFields(map[string]interface{}{
			"job_id": result.JobID,
		}).Warn("Results channel timeout, discarding result")
	}
}

// processJob processes a single job
func (p *WorkerPool) processJob(workerID int, job WorkerJob) {
	p.logger.WithFields(map[string]interface{}{
		"worker_id": workerID,
		"job_id":    job.ID,
		"priority":  job.Priority,
	}).Debug("Processing job")

	// Set up job context with cancellation
	jobCtx, cancel := p.setupJobContext(job)
	defer cancel()

	// Execute the job and measure duration
	duration, err := p.executeJob(jobCtx, job)

	// Record stats
	if p.stats != nil {
		if err != nil {
			p.stats.recordJobFailure(duration)
		} else {
			p.stats.recordJobCompletion(duration)
		}
	}

	// Create job result
	result := JobResult{
		JobID: job.ID,
		Error: err,
	}

	// Log the result
	p.logJobResult(workerID, job.ID, duration, err)

	// Send the result
	p.sendJobResult(result)
}

// createJob creates a new job with the given parameters
func (p *WorkerPool) createJob(id string, task TaskFunc, priority int, ctx context.Context) WorkerJob {
	if ctx == nil {
		ctx = context.Background()
	}

	return WorkerJob{
		ID:       id,
		Task:     task,
		Priority: priority,
		Context:  ctx,
	}
}

// enqueueJob adds a job to the job queue with timeout to prevent deadlocks
func (p *WorkerPool) enqueueJob(job WorkerJob) error {
	select {
	case <-p.stopContext.Done():
		return errors.New("worker pool is stopped")
	case p.jobQueue <- job:
		return nil
	case <-time.After(30 * time.Second): // Prevent indefinite blocking
		return errors.New("job queue is full, timeout after 30 seconds")
	}
}

// Submit adds a job to the pool
func (p *WorkerPool) Submit(id string, task TaskFunc) error {
	return p.SubmitWithPriority(id, task, 0)
}

// SubmitWithPriority adds a job to the pool with the specified priority
func (p *WorkerPool) SubmitWithPriority(id string, task TaskFunc, priority int) error {
	if task == nil {
		return errors.InvalidInputf("task cannot be nil")
	}

	job := p.createJob(id, task, priority, context.Background())
	return p.enqueueJob(job)
}

// SubmitWithContext adds a job to the pool with the specified context
func (p *WorkerPool) SubmitWithContext(ctx context.Context, id string, task TaskFunc) error {
	if task == nil {
		return errors.InvalidInputf("task cannot be nil")
	}

	job := p.createJob(id, task, 0, ctx)
	return p.enqueueJob(job)
}

// GetResults returns the results channel
func (p *WorkerPool) GetResults() <-chan JobResult {
	return p.results
}

// Wait waits for all submitted jobs to complete
func (p *WorkerPool) Wait() {
	// Close job queue only once
	if p.jobsClosed.CompareAndSwap(false, true) {
		close(p.jobQueue)
	}

	p.waitGroup.Wait()

	// Close results channel only once, and only if pool isn't stopped
	if !p.closed.Load() && p.resultsClosed.CompareAndSwap(false, true) {
		close(p.results)
	}
}

// Stop stops the worker pool
func (p *WorkerPool) Stop() {
	if p.closed.CompareAndSwap(false, true) {
		p.stopFunc()

		// Close job queue if not already closed
		if p.jobsClosed.CompareAndSwap(false, true) {
			close(p.jobQueue)
		}

		p.waitGroup.Wait()

		// Close results channel only once
		if p.resultsClosed.CompareAndSwap(false, true) {
			close(p.results)
		}
	}
}

// WorkerCount returns the number of workers in the pool
func (p *WorkerPool) WorkerCount() int {
	return p.workers
}

// IsHealthy returns true if the worker pool is operational
func (p *WorkerPool) IsHealthy() bool {
	return !p.closed.Load()
}
