package replication

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
)

// ImprovedWorkerPool manages a pool of workers with better concurrency handling
type ImprovedWorkerPool struct {
	workers       int
	jobQueue      chan WorkerJob
	results       chan JobResult
	waitGroup     sync.WaitGroup
	stopContext   context.Context
	stopFunc      context.CancelFunc
	logger        log.Logger
	stopped       atomic.Bool
	resultsMu     sync.Mutex
	resultsClosed atomic.Bool
}

// NewImprovedWorkerPool creates a new improved worker pool
func NewImprovedWorkerPool(workerCount int, logger log.Logger) *ImprovedWorkerPool {
	if workerCount <= 0 {
		workerCount = 1
	}

	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ImprovedWorkerPool{
		workers:     workerCount,
		jobQueue:    make(chan WorkerJob, workerCount*10),
		results:     make(chan JobResult, workerCount*10),
		stopContext: ctx,
		stopFunc:    cancel,
		logger:      logger,
	}
}

// Start starts the improved worker pool
func (p *ImprovedWorkerPool) Start() {
	p.logger.WithFields(map[string]interface{}{
		"workers": p.workers,
	}).Info("Starting improved worker pool")

	for i := 0; i < p.workers; i++ {
		workerID := i
		p.waitGroup.Add(1)

		go func() {
			defer p.waitGroup.Done()
			p.worker(workerID)
		}()
	}
}

// worker processes jobs with improved error handling
func (p *ImprovedWorkerPool) worker(id int) {
	p.logger.WithFields(map[string]interface{}{
		"worker_id": id,
	}).Debug("Improved worker started")

	defer func() {
		p.logger.WithFields(map[string]interface{}{
			"worker_id": id,
		}).Debug("Improved worker stopped")
	}()

	for {
		select {
		case <-p.stopContext.Done():
			return

		case job, ok := <-p.jobQueue:
			if !ok {
				return // Job queue closed
			}
			p.processJobSafely(id, job)
		}
	}
}

// processJobSafely processes a job with improved error handling and result reporting
func (p *ImprovedWorkerPool) processJobSafely(workerID int, job WorkerJob) {
	p.logger.WithFields(map[string]interface{}{
		"worker_id": workerID,
		"job_id":    job.ID,
		"priority":  job.Priority,
	}).Debug("Processing job safely")

	// Set up job context with cancellation
	jobCtx, cancel := p.setupJobContext(job)
	defer cancel()

	// Execute the job and measure duration
	startTime := time.Now()
	err := job.Task(jobCtx)
	duration := time.Since(startTime)

	// Create job result
	result := JobResult{
		JobID: job.ID,
		Error: err,
	}

	// Log the result
	if err != nil {
		p.logger.WithFields(map[string]interface{}{
			"worker_id": workerID,
			"job_id":    job.ID,
			"duration":  duration.String(),
		}).Error("Job failed", err)
	} else {
		p.logger.WithFields(map[string]interface{}{
			"worker_id": workerID,
			"job_id":    job.ID,
			"duration":  duration.String(),
		}).Debug("Job completed successfully")
	}

	// Send result safely
	p.sendResultSafely(result)
}

// setupJobContext creates a job context that will be canceled properly
func (p *ImprovedWorkerPool) setupJobContext(job WorkerJob) (context.Context, context.CancelFunc) {
	// Create a context that inherits from both the job context and pool context
	if job.Context == nil {
		job.Context = context.Background()
	}

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

// sendResultSafely sends a result without blocking and handles closed channels
func (p *ImprovedWorkerPool) sendResultSafely(result JobResult) {
	// Check if results channel is closed
	if p.resultsClosed.Load() {
		return
	}

	select {
	case p.results <- result:
		// Result sent successfully
	case <-p.stopContext.Done():
		// Pool is stopping, don't block
		return
	default:
		// Results channel is full, log and discard
		p.logger.WithFields(map[string]interface{}{
			"job_id": result.JobID,
		}).Warn("Results channel is full, discarding result")
	}
}

// Submit adds a job to the pool
func (p *ImprovedWorkerPool) Submit(id string, task TaskFunc) error {
	return p.SubmitWithPriority(id, task, 0)
}

// SubmitWithPriority adds a job to the pool with the specified priority
func (p *ImprovedWorkerPool) SubmitWithPriority(id string, task TaskFunc, priority int) error {
	if task == nil {
		return errors.InvalidInputf("task cannot be nil")
	}

	job := WorkerJob{
		ID:       id,
		Task:     task,
		Priority: priority,
		Context:  context.Background(),
	}

	return p.enqueueJobSafely(job)
}

// SubmitWithContext adds a job to the pool with the specified context
func (p *ImprovedWorkerPool) SubmitWithContext(ctx context.Context, id string, task TaskFunc) error {
	if task == nil {
		return errors.InvalidInputf("task cannot be nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	job := WorkerJob{
		ID:       id,
		Task:     task,
		Priority: 0,
		Context:  ctx,
	}

	return p.enqueueJobSafely(job)
}

// enqueueJobSafely adds a job to the queue with safety checks
func (p *ImprovedWorkerPool) enqueueJobSafely(job WorkerJob) error {
	if p.stopped.Load() {
		return errors.New("worker pool is stopped")
	}

	select {
	case <-p.stopContext.Done():
		return errors.New("worker pool is stopping")
	case p.jobQueue <- job:
		return nil
	default:
		return errors.New("job queue is full")
	}
}

// GetResults returns the results channel
func (p *ImprovedWorkerPool) GetResults() <-chan JobResult {
	return p.results
}

// Wait waits for all submitted jobs to complete and closes results channel
func (p *ImprovedWorkerPool) Wait() {
	// Close job queue to signal no more jobs
	close(p.jobQueue)

	// Wait for all workers to finish
	p.waitGroup.Wait()

	// Close results channel safely
	p.closeResultsSafely()
}

// Stop stops the worker pool immediately
func (p *ImprovedWorkerPool) Stop() {
	if p.stopped.CompareAndSwap(false, true) {
		// Cancel context to signal workers to stop
		p.stopFunc()

		// Wait for workers to stop
		p.waitGroup.Wait()

		// Close results channel
		p.closeResultsSafely()
	}
}

// closeResultsSafely closes the results channel only once
func (p *ImprovedWorkerPool) closeResultsSafely() {
	p.resultsMu.Lock()
	defer p.resultsMu.Unlock()

	if !p.resultsClosed.Load() {
		close(p.results)
		p.resultsClosed.Store(true)
	}
}

// WorkerCount returns the number of workers in the pool
func (p *ImprovedWorkerPool) WorkerCount() int {
	return p.workers
}

// IsRunning returns true if the worker pool is currently running
func (p *ImprovedWorkerPool) IsRunning() bool {
	return !p.stopped.Load()
}
