package replication

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/helper/util"
)

// HighPerformanceWorkerPool provides adaptive concurrency optimized for container registry operations
type HighPerformanceWorkerPool struct {
	// Core configuration
	minWorkers     int
	maxWorkers     int
	currentWorkers atomic.Int32

	// Channels for work distribution
	jobQueue      chan HighPerformanceJob
	results       chan HighPerformanceJobResult
	workerControl chan workerControlMessage

	// Lifecycle management
	waitGroup   sync.WaitGroup
	stopContext context.Context
	stopFunc    context.CancelFunc
	started     atomic.Bool
	stopped     atomic.Bool

	// Performance monitoring
	performanceMonitor *util.PerformanceMonitor
	logger             log.Logger

	// Adaptive concurrency metrics
	throughputTracker *ThroughputTracker
	lastAdjustment    atomic.Int64 // Unix timestamp

	// Resource management
	resultsMu     sync.Mutex
	resultsClosed atomic.Bool
}

// HighPerformanceJob represents a job with performance tracking
type HighPerformanceJob struct {
	ID       string
	Task     TaskFunc
	Priority int
	Context  context.Context

	// Performance tracking
	EstimatedBytes    int64
	EstimatedDuration time.Duration
	SubmissionTime    time.Time
}

// HighPerformanceJobResult contains job results with performance metrics
type HighPerformanceJobResult struct {
	JobID string
	Error error

	// Performance metrics
	ExecutionTime  time.Duration
	BytesProcessed int64
	QueueTime      time.Duration
	WorkerID       int
}

// ThroughputTracker monitors system throughput for adaptive scaling
type ThroughputTracker struct {
	// Sliding window of throughput measurements
	measurements []ThroughputMeasurement
	mutex        sync.RWMutex

	// Target throughput (bytes per second)
	targetThroughput int64

	// Performance history
	lastMeasurement time.Time
	totalBytes      atomic.Int64
	totalJobs       atomic.Int64
}

// ThroughputMeasurement represents a single throughput measurement
type ThroughputMeasurement struct {
	Timestamp      time.Time
	BytesPerSecond int64
	JobsPerSecond  float64
	ActiveWorkers  int
	QueueDepth     int
}

// workerControlMessage controls worker lifecycle
type workerControlMessage struct {
	action   string // "start", "stop"
	workerID int
}

// NewHighPerformanceWorkerPool creates a new high-performance worker pool
func NewHighPerformanceWorkerPool(config HighPerformancePoolConfig, logger log.Logger) *HighPerformanceWorkerPool {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Set intelligent defaults based on system capabilities
	minWorkers := config.MinWorkers
	maxWorkers := config.MaxWorkers

	if minWorkers <= 0 {
		// Base minimum on CPU cores for I/O bound operations
		minWorkers = runtime.NumCPU() * 4
		if minWorkers < 10 {
			minWorkers = 10
		}
	}

	if maxWorkers <= 0 {
		// For container registry operations (I/O bound), higher concurrency is beneficial
		maxWorkers = runtime.NumCPU() * 20
		if maxWorkers < 100 {
			maxWorkers = 100
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Target throughput: 125 MB/s (middle of 100-150 MB/s range)
	targetThroughput := int64(125 * 1024 * 1024)

	pool := &HighPerformanceWorkerPool{
		minWorkers:    minWorkers,
		maxWorkers:    maxWorkers,
		jobQueue:      make(chan HighPerformanceJob, maxWorkers*4), // Larger buffer for high throughput
		results:       make(chan HighPerformanceJobResult, maxWorkers*4),
		workerControl: make(chan workerControlMessage, maxWorkers),
		stopContext:   ctx,
		stopFunc:      cancel,
		logger:        logger,
		throughputTracker: &ThroughputTracker{
			measurements:     make([]ThroughputMeasurement, 0, 60), // 1 minute of history at 1s intervals
			targetThroughput: targetThroughput,
		},
		performanceMonitor: util.NewPerformanceMonitor(logger),
	}

	// Set initial worker count
	pool.currentWorkers.Store(int32(minWorkers))

	return pool
}

// HighPerformancePoolConfig configures the high-performance worker pool
type HighPerformancePoolConfig struct {
	MinWorkers            int
	MaxWorkers            int
	TargetThroughputMBps  int
	AdaptiveScaling       bool
	PerformanceMonitoring bool
}

// DefaultHighPerformancePoolConfig returns optimized defaults for container registry operations
func DefaultHighPerformancePoolConfig() HighPerformancePoolConfig {
	return HighPerformancePoolConfig{
		MinWorkers:            0, // Will be calculated based on system
		MaxWorkers:            0, // Will be calculated based on system
		TargetThroughputMBps:  125,
		AdaptiveScaling:       true,
		PerformanceMonitoring: true,
	}
}

// Start starts the high-performance worker pool
func (hp *HighPerformanceWorkerPool) Start() error {
	if !hp.started.CompareAndSwap(false, true) {
		return errors.New("worker pool already started")
	}

	hp.logger.WithFields(map[string]interface{}{
		"min_workers":            hp.minWorkers,
		"max_workers":            hp.maxWorkers,
		"target_throughput_mbps": hp.throughputTracker.targetThroughput / (1024 * 1024),
	}).Info("Starting high-performance worker pool")

	// Start performance monitoring
	hp.performanceMonitor.Start()

	// Start initial workers
	for i := 0; i < hp.minWorkers; i++ {
		hp.startWorker(i)
	}

	// Start adaptive scaling monitor
	go hp.adaptiveScalingMonitor()

	// Start throughput tracker
	go hp.throughputTracker.start(hp.stopContext)

	return nil
}

// startWorker starts a new worker with the given ID
func (hp *HighPerformanceWorkerPool) startWorker(workerID int) {
	hp.waitGroup.Add(1)

	go func() {
		defer hp.waitGroup.Done()
		hp.worker(workerID)
	}()

	hp.logger.WithFields(map[string]interface{}{
		"worker_id":     workerID,
		"total_workers": hp.currentWorkers.Load(),
	}).Debug("Started worker")
}

// worker processes jobs with performance tracking
func (hp *HighPerformanceWorkerPool) worker(workerID int) {
	hp.logger.WithFields(map[string]interface{}{
		"worker_id": workerID,
	}).Debug("Worker started")

	defer func() {
		hp.logger.WithFields(map[string]interface{}{
			"worker_id": workerID,
		}).Debug("Worker stopped")
	}()

	for {
		select {
		case <-hp.stopContext.Done():
			return

		case job, ok := <-hp.jobQueue:
			if !ok {
				return // Job queue closed
			}
			hp.processJobWithMetrics(workerID, job)
		}
	}
}

// processJobWithMetrics processes a job with comprehensive performance tracking
func (hp *HighPerformanceWorkerPool) processJobWithMetrics(workerID int, job HighPerformanceJob) {
	startTime := time.Now()
	queueTime := startTime.Sub(job.SubmissionTime)

	// Start operation tracking
	tracker := hp.performanceMonitor.StartOperation("job_execution")
	if job.EstimatedBytes > 0 {
		tracker.AddBytes(job.EstimatedBytes)
	}
	tracker.AddItems(1)

	// Execute the job
	err := job.Task(job.Context)

	// Calculate metrics
	executionTime := time.Since(startTime)

	// Finish operation tracking
	tracker.Finish(err)

	// Create result
	result := HighPerformanceJobResult{
		JobID:          job.ID,
		Error:          err,
		ExecutionTime:  executionTime,
		BytesProcessed: job.EstimatedBytes,
		QueueTime:      queueTime,
		WorkerID:       workerID,
	}

	// Update throughput tracking
	hp.throughputTracker.recordJob(job.EstimatedBytes, executionTime)

	// Send result safely
	hp.sendResultSafely(result)

	// Log performance metrics
	hp.logger.WithFields(map[string]interface{}{
		"worker_id":       workerID,
		"job_id":          job.ID,
		"execution_ms":    executionTime.Milliseconds(),
		"queue_ms":        queueTime.Milliseconds(),
		"bytes_processed": job.EstimatedBytes,
		"error":           err != nil,
	}).Debug("Job completed with metrics")
}

// sendResultSafely sends a result without blocking
func (hp *HighPerformanceWorkerPool) sendResultSafely(result HighPerformanceJobResult) {
	if hp.resultsClosed.Load() {
		return
	}

	select {
	case hp.results <- result:
		// Result sent successfully
	case <-hp.stopContext.Done():
		// Pool is stopping
		return
	default:
		// Results channel is full, log and discard
		hp.logger.WithFields(map[string]interface{}{
			"job_id": result.JobID,
		}).Warn("Results channel is full, discarding result")
	}
}

// Submit submits a high-performance job
func (hp *HighPerformanceWorkerPool) Submit(job HighPerformanceJob) error {
	if hp.stopped.Load() {
		return errors.New("worker pool is stopped")
	}

	job.SubmissionTime = time.Now()

	select {
	case <-hp.stopContext.Done():
		return errors.New("worker pool is stopping")
	case hp.jobQueue <- job:
		return nil
	default:
		return errors.New("job queue is full")
	}
}

// GetResults returns the results channel
func (hp *HighPerformanceWorkerPool) GetResults() <-chan HighPerformanceJobResult {
	return hp.results
}

// adaptiveScalingMonitor monitors performance and adjusts worker count
func (hp *HighPerformanceWorkerPool) adaptiveScalingMonitor() {
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-hp.stopContext.Done():
			return
		case <-ticker.C:
			hp.adjustWorkerCount()
		}
	}
}

// adjustWorkerCount dynamically adjusts the number of workers based on performance
func (hp *HighPerformanceWorkerPool) adjustWorkerCount() {
	currentWorkers := int(hp.currentWorkers.Load())
	queueDepth := len(hp.jobQueue)

	// Get current throughput
	currentThroughput := hp.throughputTracker.getCurrentThroughput()
	targetThroughput := hp.throughputTracker.targetThroughput

	// Scaling decisions based on multiple factors
	shouldScaleUp := false
	shouldScaleDown := false

	// Scale up conditions
	if queueDepth > currentWorkers*2 { // Queue depth indicates bottleneck
		shouldScaleUp = true
	}
	if currentThroughput < targetThroughput*8/10 && currentWorkers < hp.maxWorkers { // Below 80% of target
		shouldScaleUp = true
	}

	// Scale down conditions (more conservative)
	if queueDepth == 0 && currentWorkers > hp.minWorkers { // No queued work
		shouldScaleDown = true
	}
	if currentThroughput > targetThroughput*12/10 && currentWorkers > hp.minWorkers { // Above 120% of target
		shouldScaleDown = true
	}

	// Apply scaling with rate limiting
	now := time.Now().Unix()
	lastAdjustment := hp.lastAdjustment.Load()

	if now-lastAdjustment < 10 { // Don't adjust more than once every 10 seconds
		return
	}

	if shouldScaleUp && currentWorkers < hp.maxWorkers {
		newWorkerID := currentWorkers
		hp.startWorker(newWorkerID)
		hp.currentWorkers.Add(1)
		hp.lastAdjustment.Store(now)

		hp.logger.WithFields(map[string]interface{}{
			"new_worker_count":       currentWorkers + 1,
			"queue_depth":            queueDepth,
			"throughput_mbps":        currentThroughput / (1024 * 1024),
			"target_throughput_mbps": targetThroughput / (1024 * 1024),
		}).Info("Scaled up worker pool")

	} else if shouldScaleDown && currentWorkers > hp.minWorkers {
		// Note: In this implementation, we don't actively stop workers
		// They will naturally finish and stop when the pool is stopped
		// A more sophisticated implementation could implement worker shutdown
		hp.logger.WithFields(map[string]interface{}{
			"current_workers":        currentWorkers,
			"queue_depth":            queueDepth,
			"throughput_mbps":        currentThroughput / (1024 * 1024),
			"target_throughput_mbps": targetThroughput / (1024 * 1024),
		}).Debug("Worker pool could be scaled down (natural attrition)")
	}
}

// recordJob records job completion for throughput tracking
func (tt *ThroughputTracker) recordJob(bytes int64, duration time.Duration) {
	tt.totalBytes.Add(bytes)
	tt.totalJobs.Add(1)
}

// getCurrentThroughput returns the current throughput in bytes per second
func (tt *ThroughputTracker) getCurrentThroughput() int64 {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	if len(tt.measurements) == 0 {
		return 0
	}

	// Return the most recent measurement
	latest := tt.measurements[len(tt.measurements)-1]
	return latest.BytesPerSecond
}

// start begins throughput tracking
func (tt *ThroughputTracker) start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastBytes, lastJobs int64
	lastTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			currentBytes := tt.totalBytes.Load()
			currentJobs := tt.totalJobs.Load()

			deltaTime := now.Sub(lastTime).Seconds()
			deltaBytes := currentBytes - lastBytes
			deltaJobs := currentJobs - lastJobs

			if deltaTime > 0 {
				bytesPerSecond := int64(float64(deltaBytes) / deltaTime)
				jobsPerSecond := float64(deltaJobs) / deltaTime

				measurement := ThroughputMeasurement{
					Timestamp:      now,
					BytesPerSecond: bytesPerSecond,
					JobsPerSecond:  jobsPerSecond,
				}

				tt.addMeasurement(measurement)
			}

			lastBytes = currentBytes
			lastJobs = currentJobs
			lastTime = now
		}
	}
}

// addMeasurement adds a new throughput measurement
func (tt *ThroughputTracker) addMeasurement(measurement ThroughputMeasurement) {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	tt.measurements = append(tt.measurements, measurement)

	// Keep only last 60 measurements (1 minute of history)
	if len(tt.measurements) > 60 {
		tt.measurements = tt.measurements[1:]
	}
}

// Stop stops the high-performance worker pool
func (hp *HighPerformanceWorkerPool) Stop() {
	if !hp.stopped.CompareAndSwap(false, true) {
		return // Already stopped
	}

	// Cancel context to signal workers to stop
	hp.stopFunc()

	// Close job queue
	close(hp.jobQueue)

	// Wait for workers to finish
	hp.waitGroup.Wait()

	// Stop performance monitoring
	hp.performanceMonitor.Stop()

	// Close results channel safely
	hp.closeResultsSafely()

	hp.logger.Info("High-performance worker pool stopped")
}

// closeResultsSafely closes the results channel only once
func (hp *HighPerformanceWorkerPool) closeResultsSafely() {
	hp.resultsMu.Lock()
	defer hp.resultsMu.Unlock()

	if !hp.resultsClosed.Load() {
		close(hp.results)
		hp.resultsClosed.Store(true)
	}
}

// GetPerformanceReport returns a comprehensive performance report
func (hp *HighPerformanceWorkerPool) GetPerformanceReport() *HighPerformancePoolReport {
	report := &HighPerformancePoolReport{
		Timestamp:          time.Now(),
		CurrentWorkers:     int(hp.currentWorkers.Load()),
		MinWorkers:         hp.minWorkers,
		MaxWorkers:         hp.maxWorkers,
		QueueDepth:         len(hp.jobQueue),
		TotalJobs:          hp.throughputTracker.totalJobs.Load(),
		TotalBytes:         hp.throughputTracker.totalBytes.Load(),
		PerformanceMetrics: hp.performanceMonitor.GenerateReport(),
	}

	// Add throughput history
	hp.throughputTracker.mutex.RLock()
	report.ThroughputHistory = make([]ThroughputMeasurement, len(hp.throughputTracker.measurements))
	copy(report.ThroughputHistory, hp.throughputTracker.measurements)
	hp.throughputTracker.mutex.RUnlock()

	return report
}

// HighPerformancePoolReport contains comprehensive pool performance metrics
type HighPerformancePoolReport struct {
	Timestamp          time.Time
	CurrentWorkers     int
	MinWorkers         int
	MaxWorkers         int
	QueueDepth         int
	TotalJobs          int64
	TotalBytes         int64
	ThroughputHistory  []ThroughputMeasurement
	PerformanceMetrics *util.PerformanceReport
}
