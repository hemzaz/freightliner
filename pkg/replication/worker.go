package replication

import (
	"context"
	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"sync"
	"sync/atomic"
)

// Task represents a task to be executed by a worker
type Task func(context.Context) error

// WorkerPool manages a pool of workers for parallel execution
type WorkerPool struct {
	workers  int
	jobs     chan Task
	wg       sync.WaitGroup
	ctx      context.Context
	cancelFn context.CancelFunc
	logger   *log.Logger
	running  atomic.Bool
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int, logger *log.Logger) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}

	if logger == nil {
		logger = log.NewLogger(log.InfoLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:  workers,
		jobs:     make(chan Task),
		ctx:      ctx,
		cancelFn: cancel,
		logger:   logger,
	}

	return pool
}

// Start starts the worker pool
func (p *WorkerPool) Start() {
	if p.running.Load() {
		p.logger.Warn("Worker pool already running", nil)
		return
	}

	p.running.Store(true)
	p.logger.Info("Starting worker pool", map[string]interface{}{
		"workers": p.workers,
	})

	// Start workers
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// Submit adds a task to the worker pool
func (p *WorkerPool) Submit(task Task) error {
	if !p.running.Load() {
		return errors.InvalidInputf("worker pool not running")
	}

	select {
	case p.jobs <- task:
		return nil
	case <-p.ctx.Done():
		return errors.Wrap(p.ctx.Err(), "worker pool was stopped")
	}
}

// Stop stops the worker pool and waits for all workers to finish
func (p *WorkerPool) Stop() {
	if !p.running.Load() {
		return
	}

	// First, cancel the context to signal stop to all workers
	p.cancelFn()

	// Next close the jobs channel to prevent new submissions
	close(p.jobs)

	// Wait for all workers to finish
	p.wg.Wait()

	// Mark as not running
	p.running.Store(false)

	p.logger.Info("Worker pool stopped", nil)
}

// worker is the helper worker goroutine
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	p.logger.Debug("Worker started", map[string]interface{}{
		"worker_id": id,
	})

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Debug("Worker stopping due to cancellation", map[string]interface{}{
				"worker_id": id,
			})
			return
		case task, ok := <-p.jobs:
			if !ok {
				// Channel closed
				p.logger.Debug("Worker stopping due to closed job channel", map[string]interface{}{
					"worker_id": id,
				})
				return
			}

			// Execute the task
			err := task(p.ctx)
			if err != nil {
				if p.ctx.Err() != nil {
					// Cancelled context, expected error
					p.logger.Debug("Task cancelled", map[string]interface{}{
						"worker_id": id,
						"error":     err.Error(),
					})
				} else {
					// Actual task error
					p.logger.Error("Task failed", err, map[string]interface{}{
						"worker_id": id,
					})
				}
			}
		}
	}
}

// IsRunning returns true if the worker pool is running
func (p *WorkerPool) IsRunning() bool {
	return p.running.Load()
}

// Wait waits for all submitted tasks to complete
// This does not stop the worker pool, it just waits for the current queue to drain
func (p *WorkerPool) Wait() {
	// We simply wait for all jobs to be processed
	// Since we have access to the WaitGroup (wg) we can just wait on it
	p.wg.Wait()
}

// WithContext returns a new worker pool with the given context
func (p *WorkerPool) WithContext(ctx context.Context) *WorkerPool {
	childCtx, cancel := context.WithCancel(ctx)

	// Cancel the previous context
	p.cancelFn()

	// Update the context and cancel function
	p.ctx = childCtx
	p.cancelFn = cancel

	return p
}
