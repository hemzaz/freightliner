package replication

import (
	"context"
	"runtime"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
)

// AutoScaler automatically adjusts worker pool size based on load
type AutoScaler struct {
	pool           *WorkerPool
	logger         log.Logger
	minWorkers     int
	maxWorkers     int
	scaleUpThresh  float64 // Queue utilization threshold to scale up (0.0-1.0)
	scaleDownTresh float64 // Queue utilization threshold to scale down (0.0-1.0)
	checkInterval  time.Duration
	stopCh         chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
	enabled        bool
}

// AutoScalerConfig configuration for the autoscaler
type AutoScalerConfig struct {
	MinWorkers      int
	MaxWorkers      int
	ScaleUpThresh   float64 // e.g., 0.7 = scale up when queue is 70% full
	ScaleDownThresh float64 // e.g., 0.3 = scale down when queue is 30% full
	CheckInterval   time.Duration
}

// NewAutoScaler creates a new autoscaler
func NewAutoScaler(pool *WorkerPool, config AutoScalerConfig, logger log.Logger) *AutoScaler {
	if config.MinWorkers <= 0 {
		config.MinWorkers = 1
	}
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = runtime.NumCPU() * 2
	}
	if config.ScaleUpThresh <= 0 {
		config.ScaleUpThresh = 0.7
	}
	if config.ScaleDownThresh <= 0 {
		config.ScaleDownThresh = 0.3
	}
	if config.CheckInterval <= 0 {
		config.CheckInterval = 30 * time.Second
	}
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &AutoScaler{
		pool:           pool,
		logger:         logger,
		minWorkers:     config.MinWorkers,
		maxWorkers:     config.MaxWorkers,
		scaleUpThresh:  config.ScaleUpThresh,
		scaleDownTresh: config.ScaleDownThresh,
		checkInterval:  config.CheckInterval,
		stopCh:         make(chan struct{}),
	}
}

// Start starts the autoscaler
func (as *AutoScaler) Start(ctx context.Context) {
	as.mu.Lock()
	if as.enabled {
		as.mu.Unlock()
		return
	}
	as.enabled = true
	as.mu.Unlock()

	as.logger.WithFields(map[string]interface{}{
		"min_workers":       as.minWorkers,
		"max_workers":       as.maxWorkers,
		"scale_up_thresh":   as.scaleUpThresh,
		"scale_down_thresh": as.scaleDownTresh,
		"check_interval":    as.checkInterval,
	}).Info("Starting worker pool autoscaler")

	as.wg.Add(1)
	go as.run(ctx)
}

// Stop stops the autoscaler
func (as *AutoScaler) Stop() {
	as.mu.Lock()
	if !as.enabled {
		as.mu.Unlock()
		return
	}
	as.enabled = false
	as.mu.Unlock()

	close(as.stopCh)
	as.wg.Wait()

	as.logger.Info("Autoscaler stopped")
}

// run runs the autoscaling loop
func (as *AutoScaler) run(ctx context.Context) {
	defer as.wg.Done()

	ticker := time.NewTicker(as.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-as.stopCh:
			return
		case <-ticker.C:
			as.evaluate()
		}
	}
}

// evaluate evaluates current load and adjusts worker count
func (as *AutoScaler) evaluate() {
	stats := as.pool.GetStats()

	// Calculate queue utilization
	queueCap := cap(as.pool.jobQueue)
	queueSize := stats.QueuedJobs
	utilization := float64(queueSize) / float64(queueCap)

	// Get system metrics
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	cpuCount := runtime.NumCPU()

	as.logger.WithFields(map[string]interface{}{
		"current_workers": stats.TotalWorkers,
		"active_workers":  stats.ActiveWorkers,
		"queued_jobs":     queueSize,
		"queue_cap":       queueCap,
		"utilization":     utilization,
		"cpu_count":       cpuCount,
		"mem_alloc_mb":    memStats.Alloc / 1024 / 1024,
	}).Debug("Evaluating autoscaler metrics")

	currentWorkers := stats.TotalWorkers

	// Scale up if queue utilization is high
	if utilization >= as.scaleUpThresh && currentWorkers < as.maxWorkers {
		// Calculate how many workers to add
		// Add 25% more workers, but respect max limit
		additionalWorkers := max(1, currentWorkers/4)
		newWorkerCount := min(currentWorkers+additionalWorkers, as.maxWorkers)

		as.logger.WithFields(map[string]interface{}{
			"from":        currentWorkers,
			"to":          newWorkerCount,
			"utilization": utilization,
		}).Info("Scaling up worker pool")

		as.scaleUp(newWorkerCount - currentWorkers)
	}

	// Scale down if queue utilization is low and we have idle workers
	if utilization <= as.scaleDownTresh && currentWorkers > as.minWorkers {
		// Only scale down if we have idle workers
		idleWorkers := stats.IdleWorkers
		if idleWorkers > 0 {
			// Remove 25% of idle workers, but respect min limit
			workersToRemove := max(1, idleWorkers/4)
			newWorkerCount := max(currentWorkers-workersToRemove, as.minWorkers)

			as.logger.WithFields(map[string]interface{}{
				"from":         currentWorkers,
				"to":           newWorkerCount,
				"idle_workers": idleWorkers,
				"utilization":  utilization,
			}).Info("Scaling down worker pool")

			as.scaleDown(currentWorkers - newWorkerCount)
		}
	}
}

// scaleUp adds workers to the pool
func (as *AutoScaler) scaleUp(count int) {
	// Add new workers to the pool
	// This would require modifying the WorkerPool to support dynamic worker addition
	// For now, this is a placeholder
	as.logger.WithFields(map[string]interface{}{
		"workers_to_add": count,
	}).Debug("Scale up operation (not yet implemented)")
}

// scaleDown removes workers from the pool
func (as *AutoScaler) scaleDown(count int) {
	// Remove workers from the pool gracefully
	// This would require modifying the WorkerPool to support dynamic worker removal
	// For now, this is a placeholder
	as.logger.WithFields(map[string]interface{}{
		"workers_to_remove": count,
	}).Debug("Scale down operation (not yet implemented)")
}

// GetCurrentWorkerCount returns the current number of workers
func (as *AutoScaler) GetCurrentWorkerCount() int {
	return as.pool.GetStats().TotalWorkers
}

// SetMinWorkers updates the minimum worker count
func (as *AutoScaler) SetMinWorkers(min int) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.minWorkers = max(1, min)
}

// SetMaxWorkers updates the maximum worker count
func (as *AutoScaler) SetMaxWorkers(max int) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.maxWorkers = max
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
