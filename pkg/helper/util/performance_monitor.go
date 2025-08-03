package util

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/log"
)

// PerformanceMonitor provides comprehensive performance monitoring capabilities
type PerformanceMonitor struct {
	logger    log.Logger
	metrics   *PerformanceMetrics
	enabled   atomic.Bool
	startTime time.Time

	// Operation tracking
	operations map[string]*OperationMetrics
	opMutex    sync.RWMutex

	// Resource monitoring
	resourceMonitor *ResourceMonitor
	gcOptimizer     *GCOptimizer
}

// PerformanceMetrics holds comprehensive performance metrics
type PerformanceMetrics struct {
	// Memory metrics
	HeapAllocated atomic.Int64 // Current heap allocation
	HeapInUse     atomic.Int64 // Current heap in use
	HeapObjects   atomic.Int64 // Number of heap objects
	StackInUse    atomic.Int64 // Current stack in use

	// GC metrics
	GCCycles     atomic.Uint32 // Number of GC cycles
	GCPauseTotal atomic.Int64  // Total GC pause time (ns)
	GCPauseLast  atomic.Int64  // Last GC pause time (ns)

	// Goroutine metrics
	NumGoroutines atomic.Int32 // Current number of goroutines
	MaxGoroutines atomic.Int32 // Peak number of goroutines

	// CPU metrics
	CPUCores int          // Number of CPU cores
	CPUUsage atomic.Int64 // CPU usage percentage (scaled by 1000)

	// Network metrics (for container registry operations)
	BytesTransferred atomic.Int64 // Total bytes transferred
	TransferRate     atomic.Int64 // Current transfer rate (bytes/sec)

	// Operation metrics
	OperationsTotal  atomic.Int64 // Total operations performed
	OperationsFailed atomic.Int64 // Total failed operations
	AvgOperationTime atomic.Int64 // Average operation time (ns)
}

// OperationMetrics tracks metrics for specific operations
type OperationMetrics struct {
	Name          string
	Count         atomic.Int64
	TotalTime     atomic.Int64
	MinTime       atomic.Int64
	MaxTime       atomic.Int64
	ErrorCount    atomic.Int64
	LastExecution atomic.Int64 // Unix timestamp

	// Operation-specific metrics
	BytesProcessed atomic.Int64
	ItemsProcessed atomic.Int64

	// Histogram for latency distribution
	latencyHistogram *LatencyHistogram
}

// LatencyHistogram tracks latency distribution
type LatencyHistogram struct {
	buckets []atomic.Int64  // Buckets for different latency ranges
	bounds  []time.Duration // Bucket boundaries
	mutex   sync.RWMutex
}

// NewLatencyHistogram creates a new latency histogram
func NewLatencyHistogram() *LatencyHistogram {
	// Define latency buckets: <1ms, <10ms, <100ms, <1s, <10s, >=10s
	bounds := []time.Duration{
		1 * time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
		1 * time.Second,
		10 * time.Second,
	}

	return &LatencyHistogram{
		buckets: make([]atomic.Int64, len(bounds)+1), // +1 for >=10s bucket
		bounds:  bounds,
	}
}

// Record records a latency measurement
func (lh *LatencyHistogram) Record(latency time.Duration) {
	for i, bound := range lh.bounds {
		if latency < bound {
			lh.buckets[i].Add(1)
			return
		}
	}
	// Latency is >= largest bound
	lh.buckets[len(lh.bounds)].Add(1)
}

// GetDistribution returns the current latency distribution
func (lh *LatencyHistogram) GetDistribution() map[string]int64 {
	lh.mutex.RLock()
	defer lh.mutex.RUnlock()

	distribution := make(map[string]int64)
	labels := []string{"<1ms", "<10ms", "<100ms", "<1s", "<10s", ">=10s"}

	for i, label := range labels {
		distribution[label] = lh.buckets[i].Load()
	}

	return distribution
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger log.Logger) *PerformanceMonitor {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	pm := &PerformanceMonitor{
		logger:     logger,
		metrics:    &PerformanceMetrics{},
		operations: make(map[string]*OperationMetrics),
		startTime:  time.Now(),
	}

	// Initialize CPU cores
	pm.metrics.CPUCores = runtime.NumCPU()

	// Initialize GC optimizer for comprehensive monitoring
	pm.gcOptimizer = NewGCOptimizer(DefaultGCOptimizerConfig(), logger)

	return pm
}

// Start enables performance monitoring
func (pm *PerformanceMonitor) Start() {
	if !pm.enabled.CompareAndSwap(false, true) {
		return // Already started
	}

	pm.logger.WithFields(map[string]interface{}{
		"cpu_cores":  pm.metrics.CPUCores,
		"start_time": pm.startTime.Format(time.RFC3339),
	}).Info("Starting performance monitor")

	// Start background monitoring
	go pm.collectMetrics()

	// Start GC optimizer
	pm.gcOptimizer.Start()
}

// Stop disables performance monitoring
func (pm *PerformanceMonitor) Stop() {
	if !pm.enabled.CompareAndSwap(true, false) {
		return // Already stopped
	}

	// Stop GC optimizer
	pm.gcOptimizer.Stop()

	pm.logger.WithFields(map[string]interface{}{
		"total_runtime": time.Since(pm.startTime).String(),
	}).Info("Performance monitor stopped")
}

// collectMetrics continuously collects system metrics
func (pm *PerformanceMonitor) collectMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for pm.enabled.Load() {
		<-ticker.C
		pm.updateSystemMetrics()
	}
}

// updateSystemMetrics updates system-level metrics
func (pm *PerformanceMonitor) updateSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Update memory metrics
	pm.metrics.HeapAllocated.Store(int64(m.Alloc))
	pm.metrics.HeapInUse.Store(int64(m.HeapInuse))
	pm.metrics.HeapObjects.Store(int64(m.HeapObjects))
	pm.metrics.StackInUse.Store(int64(m.StackInuse))

	// Update GC metrics
	pm.metrics.GCCycles.Store(m.NumGC)
	pm.metrics.GCPauseTotal.Store(int64(m.PauseTotalNs))
	if m.NumGC > 0 {
		pm.metrics.GCPauseLast.Store(int64(m.PauseNs[(m.NumGC+255)%256]))
	}

	// Update goroutine metrics
	numGoroutines := int32(runtime.NumGoroutine())
	pm.metrics.NumGoroutines.Store(numGoroutines)

	// Update max goroutines if needed
	for {
		current := pm.metrics.MaxGoroutines.Load()
		if numGoroutines <= current {
			break
		}
		if pm.metrics.MaxGoroutines.CompareAndSwap(current, numGoroutines) {
			break
		}
	}
}

// StartOperation begins tracking an operation
func (pm *PerformanceMonitor) StartOperation(name string) *OperationTracker {
	pm.opMutex.RLock()
	opMetrics, exists := pm.operations[name]
	pm.opMutex.RUnlock()

	if !exists {
		pm.opMutex.Lock()
		// Double-check after acquiring write lock
		if opMetrics, exists = pm.operations[name]; !exists {
			opMetrics = &OperationMetrics{
				Name:             name,
				latencyHistogram: NewLatencyHistogram(),
			}
			// Initialize min time to max value
			opMetrics.MinTime.Store(int64(^uint64(0) >> 1)) // Max int64
			pm.operations[name] = opMetrics
		}
		pm.opMutex.Unlock()
	}

	return &OperationTracker{
		operation: opMetrics,
		startTime: time.Now(),
		monitor:   pm,
	}
}

// OperationTracker tracks individual operation execution
type OperationTracker struct {
	operation *OperationMetrics
	startTime time.Time
	monitor   *PerformanceMonitor

	// Operation-specific data
	bytesProcessed atomic.Int64
	itemsProcessed atomic.Int64
}

// AddBytes records bytes processed in this operation
func (ot *OperationTracker) AddBytes(bytes int64) {
	ot.bytesProcessed.Add(bytes)
}

// AddItems records items processed in this operation
func (ot *OperationTracker) AddItems(items int64) {
	ot.itemsProcessed.Add(items)
}

// Finish completes the operation tracking
func (ot *OperationTracker) Finish(err error) {
	duration := time.Since(ot.startTime)
	durationNs := duration.Nanoseconds()

	// Update operation metrics
	ot.operation.Count.Add(1)
	ot.operation.TotalTime.Add(durationNs)
	ot.operation.LastExecution.Store(time.Now().Unix())

	// Update bytes and items processed
	bytes := ot.bytesProcessed.Load()
	items := ot.itemsProcessed.Load()
	ot.operation.BytesProcessed.Add(bytes)
	ot.operation.ItemsProcessed.Add(items)

	// Update min/max time
	for {
		current := ot.operation.MinTime.Load()
		if durationNs >= current {
			break
		}
		if ot.operation.MinTime.CompareAndSwap(current, durationNs) {
			break
		}
	}

	for {
		current := ot.operation.MaxTime.Load()
		if durationNs <= current {
			break
		}
		if ot.operation.MaxTime.CompareAndSwap(current, durationNs) {
			break
		}
	}

	// Record error if any
	if err != nil {
		ot.operation.ErrorCount.Add(1)
		ot.monitor.metrics.OperationsFailed.Add(1)
	}

	// Update global metrics
	ot.monitor.metrics.OperationsTotal.Add(1)
	ot.monitor.metrics.BytesTransferred.Add(bytes)

	// Record latency in histogram
	ot.operation.latencyHistogram.Record(duration)

	// Update average operation time
	totalOps := ot.monitor.metrics.OperationsTotal.Load()
	if totalOps > 0 {
		// Simple moving average approximation
		currentAvg := ot.monitor.metrics.AvgOperationTime.Load()
		newAvg := (currentAvg*9 + durationNs) / 10 // Weighted average
		ot.monitor.metrics.AvgOperationTime.Store(newAvg)
	}

	// Log operation completion
	ot.monitor.logger.WithFields(map[string]interface{}{
		"operation":       ot.operation.Name,
		"duration_ms":     duration.Nanoseconds() / 1000000,
		"bytes_processed": bytes,
		"items_processed": items,
		"error":           err != nil,
	}).Debug("Operation completed")
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() *PerformanceMetrics {
	return pm.metrics
}

// GetOperationMetrics returns metrics for a specific operation
func (pm *PerformanceMonitor) GetOperationMetrics(name string) (*OperationMetrics, bool) {
	pm.opMutex.RLock()
	defer pm.opMutex.RUnlock()

	metrics, exists := pm.operations[name]
	return metrics, exists
}

// GetAllOperationMetrics returns metrics for all operations
func (pm *PerformanceMonitor) GetAllOperationMetrics() map[string]*OperationMetrics {
	pm.opMutex.RLock()
	defer pm.opMutex.RUnlock()

	result := make(map[string]*OperationMetrics)
	for name, metrics := range pm.operations {
		result[name] = metrics
	}
	return result
}

// GenerateReport generates a comprehensive performance report
func (pm *PerformanceMonitor) GenerateReport() *PerformanceReport {
	report := &PerformanceReport{
		Timestamp:      time.Now(),
		Uptime:         time.Since(pm.startTime),
		SystemMetrics:  pm.getSystemMetricsSnapshot(),
		OperationStats: pm.getOperationStatsSnapshot(),
	}

	// Add GC optimizer stats
	if pm.gcOptimizer != nil {
		report.GCOptimizerStats = pm.gcOptimizer.GetStats()
	}

	return report
}

// PerformanceReport contains a comprehensive performance report
type PerformanceReport struct {
	Timestamp        time.Time
	Uptime           time.Duration
	SystemMetrics    SystemMetricsSnapshot
	OperationStats   []OperationStatsSnapshot
	GCOptimizerStats *GCOptimizerStats
}

// SystemMetricsSnapshot contains a snapshot of system metrics
type SystemMetricsSnapshot struct {
	HeapAllocatedMB    float64
	HeapInUseMB        float64
	HeapObjects        int64
	StackInUseMB       float64
	GCCycles           uint32
	GCPauseTotalMS     float64
	GCPauseLastMS      float64
	NumGoroutines      int32
	MaxGoroutines      int32
	CPUCores           int
	BytesTransferredMB float64
	OperationsTotal    int64
	OperationsFailed   int64
	AvgOperationTimeMS float64
}

// OperationStatsSnapshot contains a snapshot of operation statistics
type OperationStatsSnapshot struct {
	Name                string
	Count               int64
	TotalTimeMS         float64
	AvgTimeMS           float64
	MinTimeMS           float64
	MaxTimeMS           float64
	ErrorCount          int64
	ErrorRate           float64
	BytesProcessedMB    float64
	ItemsProcessed      int64
	LatencyDistribution map[string]int64
}

// getSystemMetricsSnapshot creates a snapshot of current system metrics
func (pm *PerformanceMonitor) getSystemMetricsSnapshot() SystemMetricsSnapshot {
	const MB = 1024 * 1024

	return SystemMetricsSnapshot{
		HeapAllocatedMB:    float64(pm.metrics.HeapAllocated.Load()) / MB,
		HeapInUseMB:        float64(pm.metrics.HeapInUse.Load()) / MB,
		HeapObjects:        pm.metrics.HeapObjects.Load(),
		StackInUseMB:       float64(pm.metrics.StackInUse.Load()) / MB,
		GCCycles:           pm.metrics.GCCycles.Load(),
		GCPauseTotalMS:     float64(pm.metrics.GCPauseTotal.Load()) / 1000000,
		GCPauseLastMS:      float64(pm.metrics.GCPauseLast.Load()) / 1000000,
		NumGoroutines:      pm.metrics.NumGoroutines.Load(),
		MaxGoroutines:      pm.metrics.MaxGoroutines.Load(),
		CPUCores:           pm.metrics.CPUCores,
		BytesTransferredMB: float64(pm.metrics.BytesTransferred.Load()) / MB,
		OperationsTotal:    pm.metrics.OperationsTotal.Load(),
		OperationsFailed:   pm.metrics.OperationsFailed.Load(),
		AvgOperationTimeMS: float64(pm.metrics.AvgOperationTime.Load()) / 1000000,
	}
}

// getOperationStatsSnapshot creates snapshots of all operation statistics
func (pm *PerformanceMonitor) getOperationStatsSnapshot() []OperationStatsSnapshot {
	pm.opMutex.RLock()
	defer pm.opMutex.RUnlock()

	const MB = 1024 * 1024
	var stats []OperationStatsSnapshot

	for name, metrics := range pm.operations {
		count := metrics.Count.Load()
		totalTimeNs := metrics.TotalTime.Load()

		var avgTimeMS, errorRate float64
		if count > 0 {
			avgTimeMS = float64(totalTimeNs) / float64(count) / 1000000
			errorRate = float64(metrics.ErrorCount.Load()) / float64(count) * 100
		}

		stats = append(stats, OperationStatsSnapshot{
			Name:                name,
			Count:               count,
			TotalTimeMS:         float64(totalTimeNs) / 1000000,
			AvgTimeMS:           avgTimeMS,
			MinTimeMS:           float64(metrics.MinTime.Load()) / 1000000,
			MaxTimeMS:           float64(metrics.MaxTime.Load()) / 1000000,
			ErrorCount:          metrics.ErrorCount.Load(),
			ErrorRate:           errorRate,
			BytesProcessedMB:    float64(metrics.BytesProcessed.Load()) / MB,
			ItemsProcessed:      metrics.ItemsProcessed.Load(),
			LatencyDistribution: metrics.latencyHistogram.GetDistribution(),
		})
	}

	return stats
}

// LogReport logs a performance report
func (pm *PerformanceMonitor) LogReport() {
	report := pm.GenerateReport()

	pm.logger.WithFields(map[string]interface{}{
		"uptime_hours":         report.Uptime.Hours(),
		"heap_allocated_mb":    report.SystemMetrics.HeapAllocatedMB,
		"heap_inuse_mb":        report.SystemMetrics.HeapInUseMB,
		"heap_objects":         report.SystemMetrics.HeapObjects,
		"gc_cycles":            report.SystemMetrics.GCCycles,
		"gc_pause_total_ms":    report.SystemMetrics.GCPauseTotalMS,
		"num_goroutines":       report.SystemMetrics.NumGoroutines,
		"max_goroutines":       report.SystemMetrics.MaxGoroutines,
		"operations_total":     report.SystemMetrics.OperationsTotal,
		"operations_failed":    report.SystemMetrics.OperationsFailed,
		"avg_operation_ms":     report.SystemMetrics.AvgOperationTimeMS,
		"bytes_transferred_mb": report.SystemMetrics.BytesTransferredMB,
	}).Info("Performance Report")

	// Log top operations
	for i, opStats := range report.OperationStats {
		if i >= 5 { // Limit to top 5 operations
			break
		}

		pm.logger.WithFields(map[string]interface{}{
			"operation":          opStats.Name,
			"count":              opStats.Count,
			"avg_time_ms":        opStats.AvgTimeMS,
			"error_rate_percent": opStats.ErrorRate,
			"bytes_processed_mb": opStats.BytesProcessedMB,
		}).Info("Operation Stats")
	}
}

// Global performance monitor for convenience
var GlobalPerformanceMonitor = NewPerformanceMonitor(nil)

// BenchmarkSuite provides benchmarking capabilities for performance testing
type BenchmarkSuite struct {
	monitor    *PerformanceMonitor
	benchmarks map[string]*BenchmarkResult
	mutex      sync.RWMutex
	logger     log.Logger
}

// BenchmarkResult holds the results of a benchmark
type BenchmarkResult struct {
	Name            string
	Iterations      int64
	TotalDuration   time.Duration
	AvgDuration     time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	MemoryAllocated int64
	MemoryFreed     int64
	BytesProcessed  int64
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(logger log.Logger) *BenchmarkSuite {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}
	return &BenchmarkSuite{
		monitor:    NewPerformanceMonitor(logger),
		benchmarks: make(map[string]*BenchmarkResult),
		logger:     logger,
	}
}

// RunBenchmark runs a benchmark and records the results
func (bs *BenchmarkSuite) RunBenchmark(name string, iterations int64, benchmark func() error) *BenchmarkResult {
	bs.logger.WithFields(map[string]interface{}{
		"name":       name,
		"iterations": iterations,
	}).Info("Starting benchmark")

	// Start monitoring
	bs.monitor.Start()

	var totalDuration time.Duration
	var minDuration = time.Duration(^uint64(0) >> 1) // Max duration
	var maxDuration time.Duration
	var memBefore, memAfter runtime.MemStats

	runtime.ReadMemStats(&memBefore)
	runtime.GC() // Clean GC state before benchmark

	startTime := time.Now()

	// Run benchmark iterations
	for i := int64(0); i < iterations; i++ {
		iterStart := time.Now()
		err := benchmark()
		iterDuration := time.Since(iterStart)

		if err != nil {
			bs.logger.WithError(err).WithFields(map[string]interface{}{
				"name":      name,
				"iteration": i,
			}).Error("Benchmark iteration failed", err)
			continue
		}

		totalDuration += iterDuration
		if iterDuration < minDuration {
			minDuration = iterDuration
		}
		if iterDuration > maxDuration {
			maxDuration = iterDuration
		}
	}

	runtime.GC() // Clean up after benchmark
	runtime.ReadMemStats(&memAfter)

	// Create result
	result := &BenchmarkResult{
		Name:            name,
		Iterations:      iterations,
		TotalDuration:   time.Since(startTime),
		AvgDuration:     time.Duration(int64(totalDuration) / iterations),
		MinDuration:     minDuration,
		MaxDuration:     maxDuration,
		MemoryAllocated: int64(memAfter.TotalAlloc - memBefore.TotalAlloc),
		MemoryFreed:     int64(memBefore.HeapInuse - memAfter.HeapInuse),
	}

	// Store result
	bs.mutex.Lock()
	bs.benchmarks[name] = result
	bs.mutex.Unlock()

	// Log result
	bs.logger.WithFields(map[string]interface{}{
		"name":                name,
		"iterations":          iterations,
		"total_duration_ms":   result.TotalDuration.Nanoseconds() / 1000000,
		"avg_duration_ms":     result.AvgDuration.Nanoseconds() / 1000000,
		"min_duration_ms":     result.MinDuration.Nanoseconds() / 1000000,
		"max_duration_ms":     result.MaxDuration.Nanoseconds() / 1000000,
		"memory_allocated_mb": float64(result.MemoryAllocated) / (1024 * 1024),
	}).Info("Benchmark completed")

	return result
}

// GetBenchmarkResults returns all benchmark results
func (bs *BenchmarkSuite) GetBenchmarkResults() map[string]*BenchmarkResult {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	results := make(map[string]*BenchmarkResult)
	for name, result := range bs.benchmarks {
		results[name] = result
	}
	return results
}
