package replication

import (
	"context"
	"runtime"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
)

// WorkerTask represents a task to be executed by the worker pool
type WorkerTask func(ctx context.Context) error

// WorkerPoolOptions contains options for creating a worker pool
type WorkerPoolOptions struct {
	// Workers is the number of worker goroutines to create
	Workers int

	// Logger is the logger to use
	Logger *log.Logger

	// QueueSize is the size of the task queue (0 = unbuffered)
	QueueSize int

	// TaskTimeout is the maximum time a task can run (0 = no timeout)
	TaskTimeout time.Duration
}

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	workers int
	tasks   chan WorkerTask
	logger  *log.Logger
	wg      sync.WaitGroup
	cancel  context.CancelFunc
	ctx     context.Context
	running bool
	mu      sync.Mutex
	timeout time.Duration
	metrics *WorkerPoolMetrics
}

// WorkerPoolMetrics tracks metrics for the worker pool
type WorkerPoolMetrics struct {
	TasksSubmitted     int64
	TasksCompleted     int64
	TasksFailed        int64
	TasksInProgress    int64
	TotalExecutionTime time.Duration
	mu                 sync.Mutex
}

// NewWorkerPool creates a new worker pool with the specified options
func NewWorkerPool(opts WorkerPoolOptions) *WorkerPool {
	// Default options
	if opts.Workers <= 0 {
		opts.Workers = runtime.NumCPU()
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogger(log.InfoLevel)
	}
	if opts.QueueSize < 0 {
		opts.QueueSize = 0
	}

	// Create worker pool
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers: opts.Workers,
		tasks:   make(chan WorkerTask, opts.QueueSize),
		logger:  opts.Logger,
		ctx:     ctx,
		cancel:  cancel,
		timeout: opts.TaskTimeout,
		metrics: &WorkerPoolMetrics{},
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.running {
		return
	}

	wp.logger.Info("Starting worker pool", map[string]interface{}{
		"workers": wp.workers,
	})

	// Reset context and cancel function
	wp.ctx, wp.cancel = context.WithCancel(context.Background())

	// Start worker goroutines
	wp.wg.Add(wp.workers)
	for i := 0; i < wp.workers; i++ {
		go wp.worker(i)
	}

	wp.running = true
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.running {
		return
	}

	wp.logger.Info("Stopping worker pool", nil)

	// Cancel context to signal workers to stop
	if wp.cancel != nil {
		wp.cancel()
	}

	// Close task channel to prevent new tasks
	close(wp.tasks)

	// Wait for all workers to finish
	wp.wg.Wait()

	wp.running = false
}

// Submit submits a task to the worker pool
func (wp *WorkerPool) Submit(task WorkerTask) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.running {
		// If not running, execute task synchronously
		wp.logger.Warn("Worker pool not running, executing task synchronously", nil)

		// Increment task count
		wp.metrics.mu.Lock()
		wp.metrics.TasksSubmitted++
		wp.metrics.TasksInProgress++
		wp.metrics.mu.Unlock()

		// Execute task
		start := time.Now()
		err := task(context.Background())
		duration := time.Since(start)

		// Update metrics
		wp.metrics.mu.Lock()
		wp.metrics.TasksInProgress--
		wp.metrics.TotalExecutionTime += duration
		if err != nil {
			wp.metrics.TasksFailed++
		} else {
			wp.metrics.TasksCompleted++
		}
		wp.metrics.mu.Unlock()

		return err
	}

	// Increment tasks submitted
	wp.metrics.mu.Lock()
	wp.metrics.TasksSubmitted++
	wp.metrics.mu.Unlock()

	// Submit task to channel
	select {
	case wp.tasks <- task:
		return nil
	case <-wp.ctx.Done():
		return wp.ctx.Err()
	}
}

// worker is the worker goroutine
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	wp.logger.Debug("Worker started", map[string]interface{}{
		"worker_id": id,
	})

	for {
		select {
		case <-wp.ctx.Done():
			wp.logger.Debug("Worker stopping due to context cancellation", map[string]interface{}{
				"worker_id": id,
			})
			return
		case task, ok := <-wp.tasks:
			if !ok {
				wp.logger.Debug("Worker stopping due to closed task channel", map[string]interface{}{
					"worker_id": id,
				})
				return
			}

			// Increment in-progress count
			wp.metrics.mu.Lock()
			wp.metrics.TasksInProgress++
			wp.metrics.mu.Unlock()

			// Execute task with timeout if specified
			var taskCtx context.Context
			var taskCancel context.CancelFunc

			if wp.timeout > 0 {
				taskCtx, taskCancel = context.WithTimeout(wp.ctx, wp.timeout)
			} else {
				taskCtx, taskCancel = context.WithCancel(wp.ctx)
			}

			start := time.Now()
			err := task(taskCtx)
			duration := time.Since(start)
			taskCancel()

			// Update metrics
			wp.metrics.mu.Lock()
			wp.metrics.TasksInProgress--
			wp.metrics.TotalExecutionTime += duration
			if err != nil {
				wp.metrics.TasksFailed++
				wp.logger.Error("Task execution failed", err, map[string]interface{}{
					"worker_id": id,
					"duration":  duration.String(),
				})
			} else {
				wp.metrics.TasksCompleted++
				wp.logger.Debug("Task execution completed", map[string]interface{}{
					"worker_id": id,
					"duration":  duration.String(),
				})
			}
			wp.metrics.mu.Unlock()
		}
	}
}

// GetMetrics returns the current metrics for the worker pool
func (wp *WorkerPool) GetMetrics() WorkerPoolMetrics {
	wp.metrics.mu.Lock()
	defer wp.metrics.mu.Unlock()

	// Return a copy of the metrics
	return WorkerPoolMetrics{
		TasksSubmitted:     wp.metrics.TasksSubmitted,
		TasksCompleted:     wp.metrics.TasksCompleted,
		TasksFailed:        wp.metrics.TasksFailed,
		TasksInProgress:    wp.metrics.TasksInProgress,
		TotalExecutionTime: wp.metrics.TotalExecutionTime,
	}
}

// SetWorkerCount updates the number of worker goroutines
func (wp *WorkerPool) SetWorkerCount(count int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if count <= 0 {
		count = runtime.NumCPU()
	}

	if !wp.running || wp.workers == count {
		// Just update the count if not running
		wp.workers = count
		return
	}

	// If running, we need to restart the pool with the new count
	wp.logger.Info("Updating worker count", map[string]interface{}{
		"old_count": wp.workers,
		"new_count": count,
	})

	// Stop the pool
	wp.running = false
	if wp.cancel != nil {
		wp.cancel()
	}
	close(wp.tasks)
	wp.wg.Wait()

	// Update worker count
	wp.workers = count

	// Create new channels
	wp.tasks = make(chan WorkerTask, cap(wp.tasks))

	// Restart the pool
	wp.ctx, wp.cancel = context.WithCancel(context.Background())
	wp.wg.Add(wp.workers)
	for i := 0; i < wp.workers; i++ {
		go wp.worker(i)
	}
	wp.running = true
}

// IsRunning returns whether the worker pool is running
func (wp *WorkerPool) IsRunning() bool {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.running
}

// GetWorkerCount returns the number of worker goroutines
func (wp *WorkerPool) GetWorkerCount() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.workers
}
