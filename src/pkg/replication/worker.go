package replication

import (
	"context"
	"sync"

	"github.com/hemzaz/freightliner/src/internal/log"
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
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int, logger *log.Logger) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:  workers,
		jobs:     make(chan Task),
		ctx:      ctx,
		cancelFn: cancel,
		logger:   logger,
	}

	// Start the workers
	pool.start()

	return pool
}

// start launches the worker goroutines
func (p *WorkerPool) start() {
	p.logger.Info("Starting worker pool", map[string]interface{}{
		"workers": p.workers,
	})

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker runs a worker goroutine
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	p.logger.Debug("Worker started", map[string]interface{}{
		"worker_id": id,
	})

	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				// Channel closed, worker should exit
				p.logger.Debug("Worker shutting down", map[string]interface{}{
					"worker_id": id,
				})
				return
			}

			err := job(p.ctx)
			if err != nil {
				p.logger.Error("Task execution failed", err, map[string]interface{}{
					"worker_id": id,
				})
			}

		case <-p.ctx.Done():
			// Context cancelled, worker should exit
			p.logger.Debug("Worker context cancelled", map[string]interface{}{
				"worker_id": id,
			})
			return
		}
	}
}

// Submit adds a task to the worker pool
func (p *WorkerPool) Submit(task Task) {
	select {
	case p.jobs <- task:
		// Task submitted successfully
	case <-p.ctx.Done():
		// Context cancelled, can't submit more tasks
	}
}

// Stop shuts down the worker pool
func (p *WorkerPool) Stop() {
	p.logger.Info("Stopping worker pool", nil)

	// Cancel the context to signal workers to stop
	p.cancelFn()

	// Close the jobs channel
	close(p.jobs)

	// Wait for all workers to exit
	p.wg.Wait()

	p.logger.Info("Worker pool stopped", nil)
}

// Wait waits for all submitted tasks to complete
func (p *WorkerPool) Wait() {
	// Create a temporary channel to signal completion
	done := make(chan struct{})

	// Submit a final task that will close the channel
	p.Submit(func(ctx context.Context) error {
		close(done)
		return nil
	})

	// Wait for the signal
	<-done
}
