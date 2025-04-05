package replication

import (
	"context"
	"sync"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
)

// WorkerPool manages a pool of workers for parallel processing
type WorkerPool struct {
	workers     int
	jobQueue    chan Job
	results     chan JobResult
	waitGroup   sync.WaitGroup
	stopContext context.Context
	stopFunc    context.CancelFunc
	logger      *log.Logger
}

// Job represents a unit of work to be processed by a worker
type Job struct {
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
func NewWorkerPool(workerCount int, logger *log.Logger) *WorkerPool {
	if workerCount <= 0 {
		workerCount = 1
	}

	if logger == nil {
		logger = log.NewLogger(log.InfoLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:     workerCount,
		jobQueue:    make(chan Job, workerCount*10),
		results:     make(chan JobResult, workerCount*10),
		stopContext: ctx,
		stopFunc:    cancel,
		logger:      logger,
	}
}

// Start starts the worker pool
func (p *WorkerPool) Start() {
	p.logger.Info("Starting worker pool", map[string]interface{}{
		"workers": p.workers,
	})

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
func (p *WorkerPool) handleJobFromQueue(workerID int, job Job, ok bool) bool {
	if !ok {
		p.logger.Debug("Worker stopped by closed job queue", map[string]interface{}{
			"worker_id": workerID,
		})
		return false // Signal to stop the worker
	}

	p.processJob(workerID, job)
	return true // Continue processing
}

// worker processes jobs from the job queue
func (p *WorkerPool) worker(id int) {
	p.logger.Debug("Worker started", map[string]interface{}{
		"worker_id": id,
	})

	for {
		select {
		case <-p.stopContext.Done():
			p.logger.Debug("Worker stopped by context cancellation", map[string]interface{}{
				"worker_id": id,
			})
			return

		case job, ok := <-p.jobQueue:
			if !p.handleJobFromQueue(id, job, ok) {
				return
			}
		}
	}
}

// setupJobContext creates a job context that will be canceled if either the job's context
// or the pool's context is canceled
func (p *WorkerPool) setupJobContext(job Job) (context.Context, context.CancelFunc) {
	jobCtx, cancel := context.WithCancel(job.Context)

	// Setup cancellation if the pool's context is canceled
	go func() {
		select {
		case <-p.stopContext.Done():
			cancel()
		case <-jobCtx.Done():
			// Job already done
		}
	}()

	return jobCtx, cancel
}

// executeJob runs the job's task and measures execution time
func (p *WorkerPool) executeJob(ctx context.Context, job Job) (time.Duration, error) {
	startTime := time.Now()
	err := job.Task(ctx)
	duration := time.Since(startTime)
	return duration, err
}

// logJobResult logs the outcome of a job execution
func (p *WorkerPool) logJobResult(workerID int, jobID string, duration time.Duration, err error) {
	if err != nil {
		p.logger.Error("Job failed", err, map[string]interface{}{
			"worker_id": workerID,
			"job_id":    jobID,
			"duration":  duration.String(),
		})
	} else {
		p.logger.Debug("Job completed successfully", map[string]interface{}{
			"worker_id": workerID,
			"job_id":    jobID,
			"duration":  duration.String(),
		})
	}
}

// sendJobResult sends a job result to the results channel
func (p *WorkerPool) sendJobResult(result JobResult) {
	select {
	case p.results <- result:
		// Result sent successfully
	default:
		// Results channel is full, log and continue
		p.logger.Warn("Results channel is full, discarding result", map[string]interface{}{
			"job_id": result.JobID,
		})
	}
}

// processJob processes a single job
func (p *WorkerPool) processJob(workerID int, job Job) {
	p.logger.Debug("Processing job", map[string]interface{}{
		"worker_id": workerID,
		"job_id":    job.ID,
		"priority":  job.Priority,
	})

	// Set up job context with cancellation
	jobCtx, cancel := p.setupJobContext(job)
	defer cancel()

	// Execute the job and measure duration
	duration, err := p.executeJob(jobCtx, job)

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
func (p *WorkerPool) createJob(id string, task TaskFunc, priority int, ctx context.Context) Job {
	if ctx == nil {
		ctx = context.Background()
	}

	return Job{
		ID:       id,
		Task:     task,
		Priority: priority,
		Context:  ctx,
	}
}

// enqueueJob adds a job to the job queue
func (p *WorkerPool) enqueueJob(job Job) error {
	select {
	case <-p.stopContext.Done():
		return errors.New("worker pool is stopped")
	case p.jobQueue <- job:
		return nil
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
	close(p.jobQueue)
	p.waitGroup.Wait()
	close(p.results)
}

// Stop stops the worker pool
func (p *WorkerPool) Stop() {
	p.stopFunc()
	p.waitGroup.Wait()
	close(p.results)
}

// WorkerCount returns the number of workers in the pool
func (p *WorkerPool) WorkerCount() int {
	return p.workers
}
