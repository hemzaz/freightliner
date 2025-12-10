package monitoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/log"
)

// PerformanceBenchmarking provides comprehensive benchmarking for container registry operations
type PerformanceBenchmarking struct {
	logger log.Logger
	config BenchmarkingConfig

	// Performance targets from industry benchmarks
	targets IndustryBenchmarkTargets

	// Real-time metrics collection
	metrics *RealTimeMetrics

	// Benchmark execution
	runners map[string]*BenchmarkRunner
	mutex   sync.RWMutex

	// Lifecycle management
	started atomic.Bool
	stopped atomic.Bool
}

// BenchmarkingConfig configures the performance benchmarking system
type BenchmarkingConfig struct {
	// Benchmark execution settings
	WarmupDuration    time.Duration
	BenchmarkDuration time.Duration
	CooldownDuration  time.Duration

	// Concurrency settings
	MaxConcurrentBenchmarks int
	MaxConcurrentOperations int

	// Target thresholds
	ThroughputTargetMBps int64
	LatencyTargetMs      int64

	// Reporting settings
	ReportInterval   time.Duration
	EnableContinuous bool
	EnableRegression bool
}

// DefaultBenchmarkingConfig returns optimized defaults for container registry benchmarking
func DefaultBenchmarkingConfig() BenchmarkingConfig {
	return BenchmarkingConfig{
		WarmupDuration:          30 * time.Second,
		BenchmarkDuration:       5 * time.Minute,
		CooldownDuration:        10 * time.Second,
		MaxConcurrentBenchmarks: 5,
		MaxConcurrentOperations: 100,
		ThroughputTargetMBps:    125, // Target: 125 MB/s (middle of 100-150 range)
		LatencyTargetMs:         50,  // Target: <50ms latency
		ReportInterval:          30 * time.Second,
		EnableContinuous:        true,
		EnableRegression:        true,
	}
}

// IndustryBenchmarkTargets defines performance targets based on industry leaders
type IndustryBenchmarkTargets struct {
	// Throughput targets (MB/s)
	DockerHubThroughput int64 // 150+ MB/s
	AWSECRThroughput    int64 // 100-200 MB/s
	GCPGCRThroughput    int64 // 80-150 MB/s

	// Latency targets (milliseconds)
	DockerHubLatency int64 // <50ms
	AWSECRLatency    int64 // <100ms
	GCPGCRLatency    int64 // <100ms

	// Concurrency targets
	MaxConcurrentConnections int64 // 50+ concurrent

	// Reliability targets
	UptimeTarget    float64 // 99.9%
	ErrorRateTarget float64 // <0.1%
}

// GetIndustryBenchmarkTargets returns current industry benchmark targets
func GetIndustryBenchmarkTargets() IndustryBenchmarkTargets {
	return IndustryBenchmarkTargets{
		DockerHubThroughput:      150,
		AWSECRThroughput:         125,
		GCPGCRThroughput:         115,
		DockerHubLatency:         50,
		AWSECRLatency:            75,
		GCPGCRLatency:            85,
		MaxConcurrentConnections: 50,
		UptimeTarget:             99.9,
		ErrorRateTarget:          0.1,
	}
}

// RealTimeMetrics tracks real-time performance metrics
type RealTimeMetrics struct {
	// Throughput metrics
	CurrentThroughputMBps atomic.Int64
	PeakThroughputMBps    atomic.Int64
	AvgThroughputMBps     atomic.Int64

	// Latency metrics
	CurrentLatencyMs atomic.Int64
	MinLatencyMs     atomic.Int64
	MaxLatencyMs     atomic.Int64
	P99LatencyMs     atomic.Int64
	P95LatencyMs     atomic.Int64
	P50LatencyMs     atomic.Int64

	// Operation metrics
	TotalOperations atomic.Int64
	SuccessfulOps   atomic.Int64
	FailedOps       atomic.Int64
	ConcurrentOps   atomic.Int64
	QueuedOps       atomic.Int64

	// Resource metrics
	CPUUsagePercent    atomic.Int64
	MemoryUsageMB      atomic.Int64
	NetworkBytesPerSec atomic.Int64

	// Latency distribution
	latencyHistogram *LatencyHistogram

	// Performance snapshots for trend analysis
	snapshots []PerformanceSnapshot
	mutex     sync.RWMutex
}

// PerformanceSnapshot captures performance metrics at a point in time
type PerformanceSnapshot struct {
	Timestamp      time.Time
	ThroughputMBps int64
	LatencyMs      int64
	ConcurrentOps  int64
	CPUUsage       int64
	MemoryUsage    int64
	ErrorRate      float64
}

// LatencyHistogram tracks latency distribution with high precision
type LatencyHistogram struct {
	buckets []atomic.Int64
	bounds  []time.Duration
	mutex   sync.RWMutex
}

// NewLatencyHistogram creates a high-precision latency histogram
func NewLatencyHistogram() *LatencyHistogram {
	// Define latency buckets with high precision for container registry operations
	bounds := []time.Duration{
		1 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		25 * time.Millisecond,
		50 * time.Millisecond, // Target threshold
		100 * time.Millisecond,
		250 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
		5 * time.Second,
	}

	return &LatencyHistogram{
		buckets: make([]atomic.Int64, len(bounds)+1),
		bounds:  bounds,
	}
}

// Record records a latency measurement
func (lh *LatencyHistogram) Record(latency time.Duration) {
	for i, bound := range lh.bounds {
		if latency <= bound {
			lh.buckets[i].Add(1)
			return
		}
	}
	// Latency is greater than all bounds
	lh.buckets[len(lh.bounds)].Add(1)
}

// GetPercentile calculates the specified percentile from the histogram
func (lh *LatencyHistogram) GetPercentile(percentile float64) time.Duration {
	lh.mutex.RLock()
	defer lh.mutex.RUnlock()

	total := int64(0)
	for i := range lh.buckets {
		total += lh.buckets[i].Load()
	}

	if total == 0 {
		return 0
	}

	target := int64(float64(total) * percentile / 100.0)
	cumulative := int64(0)

	for i, bound := range lh.bounds {
		cumulative += lh.buckets[i].Load()
		if cumulative >= target {
			return bound
		}
	}

	// Return the maximum bound if not found
	return lh.bounds[len(lh.bounds)-1]
}

// BenchmarkRunner executes specific benchmark scenarios
type BenchmarkRunner struct {
	name     string
	scenario BenchmarkScenario
	metrics  *RealTimeMetrics
	logger   log.Logger

	// Execution state
	running   atomic.Bool
	cancelled atomic.Bool

	// Performance tracking
	startTime time.Time
	endTime   time.Time
	results   *BenchmarkResults
}

// BenchmarkScenario defines a specific benchmark scenario
type BenchmarkScenario struct {
	Name             string
	Description      string
	TargetThroughput int64 // MB/s
	TargetLatency    int64 // ms
	ConcurrencyLevel int
	Duration         time.Duration

	// Scenario-specific configuration
	OperationType  string // "replication", "manifest-fetch", "blob-transfer"
	PayloadSizeMB  int64
	NumberOfAssets int64

	// Execution function
	ExecuteFunc func(ctx context.Context, config ScenarioConfig) (*BenchmarkResults, error)
}

// ScenarioConfig provides runtime configuration for benchmark scenarios
type ScenarioConfig struct {
	ConcurrencyLevel int
	Duration         time.Duration
	PayloadSizeMB    int64
	NumberOfAssets   int64
	TargetRegistry   string
	MetricsCollector *RealTimeMetrics
}

// BenchmarkResults contains the results of a benchmark execution
type BenchmarkResults struct {
	ScenarioName string
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration

	// Performance metrics
	ThroughputMBps float64
	LatencyMs      int64
	P99LatencyMs   int64
	P95LatencyMs   int64
	P50LatencyMs   int64

	// Operation metrics
	TotalOperations int64
	SuccessfulOps   int64
	FailedOps       int64
	ErrorRate       float64

	// Resource utilization
	PeakCPUUsage            int64
	PeakMemoryUsage         int64
	NetworkBytesTransferred int64

	// Comparison with targets
	ThroughputTarget      int64
	LatencyTarget         int64
	MeetsThroughputTarget bool
	MeetsLatencyTarget    bool
	OverallScore          float64 // 0-100 score against industry benchmarks

	// Detailed metrics
	LatencyDistribution  map[string]int64
	PerformanceSnapshots []PerformanceSnapshot
}

// NewPerformanceBenchmarking creates a new performance benchmarking system
func NewPerformanceBenchmarking(config BenchmarkingConfig, logger log.Logger) *PerformanceBenchmarking {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	pb := &PerformanceBenchmarking{
		logger:  logger,
		config:  config,
		targets: GetIndustryBenchmarkTargets(),
		metrics: &RealTimeMetrics{
			latencyHistogram: NewLatencyHistogram(),
		},
		runners: make(map[string]*BenchmarkRunner),
	}

	// Initialize minimum latency to max value
	pb.metrics.MinLatencyMs.Store(math.MaxInt64)

	return pb
}

// Start starts the performance benchmarking system
func (pb *PerformanceBenchmarking) Start() error {
	if !pb.started.CompareAndSwap(false, true) {
		return nil // Already started
	}

	pb.logger.WithFields(map[string]interface{}{
		"throughput_target_mbps": pb.config.ThroughputTargetMBps,
		"latency_target_ms":      pb.config.LatencyTargetMs,
		"max_concurrency":        pb.config.MaxConcurrentOperations,
	}).Info("Starting performance benchmarking system")

	// Start continuous monitoring if enabled
	if pb.config.EnableContinuous {
		go pb.continuousMonitoring()
	}

	return nil
}

// Stop stops the performance benchmarking system
func (pb *PerformanceBenchmarking) Stop() {
	if !pb.stopped.CompareAndSwap(false, true) {
		return
	}

	// Cancel all running benchmarks
	pb.mutex.RLock()
	for _, runner := range pb.runners {
		runner.cancelled.Store(true)
	}
	pb.mutex.RUnlock()

	pb.logger.Info("Performance benchmarking system stopped")
}

// RunBenchmark executes a specific benchmark scenario
func (pb *PerformanceBenchmarking) RunBenchmark(scenario BenchmarkScenario) (*BenchmarkResults, error) {
	pb.logger.WithFields(map[string]interface{}{
		"scenario":          scenario.Name,
		"duration":          scenario.Duration,
		"concurrency":       scenario.ConcurrencyLevel,
		"target_throughput": scenario.TargetThroughput,
		"target_latency":    scenario.TargetLatency,
	}).Info("Starting benchmark scenario")

	runner := &BenchmarkRunner{
		name:     scenario.Name,
		scenario: scenario,
		metrics:  pb.metrics,
		logger:   pb.logger,
	}

	// Register runner
	pb.mutex.Lock()
	pb.runners[scenario.Name] = runner
	pb.mutex.Unlock()

	// Execute benchmark
	ctx, cancel := context.WithTimeout(context.Background(), scenario.Duration+pb.config.WarmupDuration+pb.config.CooldownDuration)
	defer cancel()

	results, err := pb.executeBenchmarkWithPhases(ctx, runner)

	// Unregister runner
	pb.mutex.Lock()
	delete(pb.runners, scenario.Name)
	pb.mutex.Unlock()

	if err != nil {
		pb.logger.WithFields(map[string]interface{}{
			"scenario": scenario.Name,
			"error":    err.Error(),
		}).Error("Benchmark scenario failed", err)
		return nil, err
	}

	// Log results
	pb.logBenchmarkResults(results)

	return results, nil
}

// executeBenchmarkWithPhases executes a benchmark with warmup, execution, and cooldown phases
func (pb *PerformanceBenchmarking) executeBenchmarkWithPhases(ctx context.Context, runner *BenchmarkRunner) (*BenchmarkResults, error) {
	// Phase 1: Warmup
	pb.logger.WithFields(map[string]interface{}{
		"scenario": runner.name,
		"duration": pb.config.WarmupDuration,
	}).Info("Starting warmup phase")

	warmupCtx, warmupCancel := context.WithTimeout(ctx, pb.config.WarmupDuration)
	defer warmupCancel()

	// Execute warmup with reduced load
	warmupConfig := ScenarioConfig{
		ConcurrencyLevel: runner.scenario.ConcurrencyLevel / 2,
		Duration:         pb.config.WarmupDuration,
		PayloadSizeMB:    runner.scenario.PayloadSizeMB,
		NumberOfAssets:   runner.scenario.NumberOfAssets / 2,
		MetricsCollector: runner.metrics,
	}

	_, err := runner.scenario.ExecuteFunc(warmupCtx, warmupConfig)
	if err != nil {
		return nil, fmt.Errorf("warmup phase failed: %w", err)
	}

	// Phase 2: Actual benchmark execution
	pb.logger.WithFields(map[string]interface{}{
		"scenario": runner.name,
		"duration": runner.scenario.Duration,
	}).Info("Starting benchmark execution phase")

	runner.startTime = time.Now()
	runner.running.Store(true)
	defer runner.running.Store(false)

	execCtx, execCancel := context.WithTimeout(ctx, runner.scenario.Duration)
	defer execCancel()

	// Reset metrics for clean measurement
	pb.resetMetrics()

	execConfig := ScenarioConfig{
		ConcurrencyLevel: runner.scenario.ConcurrencyLevel,
		Duration:         runner.scenario.Duration,
		PayloadSizeMB:    runner.scenario.PayloadSizeMB,
		NumberOfAssets:   runner.scenario.NumberOfAssets,
		MetricsCollector: runner.metrics,
	}

	results, err := runner.scenario.ExecuteFunc(execCtx, execConfig)
	if err != nil {
		return nil, fmt.Errorf("benchmark execution failed: %w", err)
	}

	runner.endTime = time.Now()

	// Phase 3: Cooldown
	pb.logger.WithFields(map[string]interface{}{
		"scenario": runner.name,
		"duration": pb.config.CooldownDuration,
	}).Debug("Starting cooldown phase")

	time.Sleep(pb.config.CooldownDuration)

	// Finalize results
	results = pb.finalizeResults(runner, results)

	return results, nil
}

// resetMetrics resets metrics for a clean benchmark measurement
func (pb *PerformanceBenchmarking) resetMetrics() {
	pb.metrics.CurrentThroughputMBps.Store(0)
	pb.metrics.CurrentLatencyMs.Store(0)
	pb.metrics.TotalOperations.Store(0)
	pb.metrics.SuccessfulOps.Store(0)
	pb.metrics.FailedOps.Store(0)
	pb.metrics.ConcurrentOps.Store(0)
	pb.metrics.QueuedOps.Store(0)
	pb.metrics.MinLatencyMs.Store(math.MaxInt64)
	pb.metrics.MaxLatencyMs.Store(0)

	// Clear latency histogram
	pb.metrics.latencyHistogram = NewLatencyHistogram()

	// Clear snapshots
	pb.metrics.mutex.Lock()
	pb.metrics.snapshots = nil
	pb.metrics.mutex.Unlock()
}

// finalizeResults calculates final metrics and comparisons
func (pb *PerformanceBenchmarking) finalizeResults(runner *BenchmarkRunner, results *BenchmarkResults) *BenchmarkResults {
	// Calculate performance percentiles
	results.P99LatencyMs = pb.metrics.latencyHistogram.GetPercentile(99).Milliseconds()
	results.P95LatencyMs = pb.metrics.latencyHistogram.GetPercentile(95).Milliseconds()
	results.P50LatencyMs = pb.metrics.latencyHistogram.GetPercentile(50).Milliseconds()

	// Set targets
	results.ThroughputTarget = runner.scenario.TargetThroughput
	results.LatencyTarget = runner.scenario.TargetLatency

	// Check if targets are met
	results.MeetsThroughputTarget = results.ThroughputMBps >= float64(results.ThroughputTarget)
	results.MeetsLatencyTarget = results.LatencyMs <= results.LatencyTarget

	// Calculate overall score against industry benchmarks
	results.OverallScore = pb.calculateOverallScore(results)

	// Get latency distribution
	results.LatencyDistribution = pb.getLatencyDistribution()

	// Get performance snapshots
	pb.metrics.mutex.RLock()
	results.PerformanceSnapshots = make([]PerformanceSnapshot, len(pb.metrics.snapshots))
	copy(results.PerformanceSnapshots, pb.metrics.snapshots)
	pb.metrics.mutex.RUnlock()

	return results
}

// calculateOverallScore calculates a 0-100 score against industry benchmarks
func (pb *PerformanceBenchmarking) calculateOverallScore(results *BenchmarkResults) float64 {
	// Weight different metrics
	const (
		throughputWeight  = 0.4
		latencyWeight     = 0.3
		reliabilityWeight = 0.2
		resourceWeight    = 0.1
	)

	// Throughput score (compared to Docker Hub target)
	throughputScore := math.Min(100.0, (results.ThroughputMBps/float64(pb.targets.DockerHubThroughput))*100.0)

	// Latency score (inverse - lower is better)
	latencyScore := math.Max(0.0, 100.0-(float64(results.LatencyMs)/float64(pb.targets.DockerHubLatency))*100.0)

	// Reliability score (based on error rate)
	reliabilityScore := math.Max(0.0, 100.0-(results.ErrorRate/pb.targets.ErrorRateTarget)*100.0)

	// Resource efficiency score (lower CPU/memory usage is better - simplified)
	resourceScore := math.Max(0.0, 100.0-float64(results.PeakCPUUsage))

	// Calculate weighted score
	overallScore := (throughputScore * throughputWeight) +
		(latencyScore * latencyWeight) +
		(reliabilityScore * reliabilityWeight) +
		(resourceScore * resourceWeight)

	return math.Min(100.0, math.Max(0.0, overallScore))
}

// getLatencyDistribution returns the current latency distribution
func (pb *PerformanceBenchmarking) getLatencyDistribution() map[string]int64 {
	distribution := make(map[string]int64)

	labels := []string{"<1ms", "<5ms", "<10ms", "<25ms", "<50ms", "<100ms", "<250ms", "<500ms", "<1s", "<5s", ">=5s"}

	for i, label := range labels {
		if i < len(pb.metrics.latencyHistogram.buckets) {
			distribution[label] = pb.metrics.latencyHistogram.buckets[i].Load()
		}
	}

	return distribution
}

// continuousMonitoring runs continuous performance monitoring
func (pb *PerformanceBenchmarking) continuousMonitoring() {
	ticker := time.NewTicker(pb.config.ReportInterval)
	defer ticker.Stop()

	for !pb.stopped.Load() {
		<-ticker.C
		pb.collectPerformanceSnapshot()
		pb.reportCurrentMetrics()
	}
}

// collectPerformanceSnapshot collects a performance snapshot
func (pb *PerformanceBenchmarking) collectPerformanceSnapshot() {
	snapshot := PerformanceSnapshot{
		Timestamp:      time.Now(),
		ThroughputMBps: pb.metrics.CurrentThroughputMBps.Load(),
		LatencyMs:      pb.metrics.CurrentLatencyMs.Load(),
		ConcurrentOps:  pb.metrics.ConcurrentOps.Load(),
		CPUUsage:       pb.metrics.CPUUsagePercent.Load(),
		MemoryUsage:    pb.metrics.MemoryUsageMB.Load(),
	}

	// Calculate error rate
	totalOps := pb.metrics.TotalOperations.Load()
	failedOps := pb.metrics.FailedOps.Load()
	if totalOps > 0 {
		snapshot.ErrorRate = float64(failedOps) / float64(totalOps) * 100.0
	}

	// Store snapshot
	pb.metrics.mutex.Lock()
	pb.metrics.snapshots = append(pb.metrics.snapshots, snapshot)

	// Keep only last 1000 snapshots
	if len(pb.metrics.snapshots) > 1000 {
		pb.metrics.snapshots = pb.metrics.snapshots[1:]
	}
	pb.metrics.mutex.Unlock()
}

// reportCurrentMetrics logs current performance metrics
func (pb *PerformanceBenchmarking) reportCurrentMetrics() {
	pb.logger.WithFields(map[string]interface{}{
		"throughput_mbps":         pb.metrics.CurrentThroughputMBps.Load(),
		"latency_ms":              pb.metrics.CurrentLatencyMs.Load(),
		"concurrent_operations":   pb.metrics.ConcurrentOps.Load(),
		"total_operations":        pb.metrics.TotalOperations.Load(),
		"successful_operations":   pb.metrics.SuccessfulOps.Load(),
		"failed_operations":       pb.metrics.FailedOps.Load(),
		"cpu_usage_percent":       pb.metrics.CPUUsagePercent.Load(),
		"memory_usage_mb":         pb.metrics.MemoryUsageMB.Load(),
		"meets_throughput_target": pb.metrics.CurrentThroughputMBps.Load() >= pb.config.ThroughputTargetMBps,
		"meets_latency_target":    pb.metrics.CurrentLatencyMs.Load() <= pb.config.LatencyTargetMs,
	}).Info("Performance monitoring report")
}

// logBenchmarkResults logs comprehensive benchmark results
func (pb *PerformanceBenchmarking) logBenchmarkResults(results *BenchmarkResults) {
	pb.logger.WithFields(map[string]interface{}{
		"scenario":                  results.ScenarioName,
		"duration_seconds":          results.Duration.Seconds(),
		"throughput_mbps":           results.ThroughputMBps,
		"throughput_target_mbps":    results.ThroughputTarget,
		"meets_throughput_target":   results.MeetsThroughputTarget,
		"latency_ms":                results.LatencyMs,
		"p99_latency_ms":            results.P99LatencyMs,
		"p95_latency_ms":            results.P95LatencyMs,
		"p50_latency_ms":            results.P50LatencyMs,
		"latency_target_ms":         results.LatencyTarget,
		"meets_latency_target":      results.MeetsLatencyTarget,
		"total_operations":          results.TotalOperations,
		"successful_operations":     results.SuccessfulOps,
		"failed_operations":         results.FailedOps,
		"error_rate_percent":        results.ErrorRate,
		"peak_cpu_usage_percent":    results.PeakCPUUsage,
		"peak_memory_usage_mb":      results.PeakMemoryUsage,
		"network_bytes_transferred": results.NetworkBytesTransferred,
		"overall_score":             results.OverallScore,
		"docker_hub_comparison":     fmt.Sprintf("%.1f%% of Docker Hub performance", (results.ThroughputMBps/float64(pb.targets.DockerHubThroughput))*100),
		"aws_ecr_comparison":        fmt.Sprintf("%.1f%% of AWS ECR performance", (results.ThroughputMBps/float64(pb.targets.AWSECRThroughput))*100),
		"gcp_gcr_comparison":        fmt.Sprintf("%.1f%% of GCP GCR performance", (results.ThroughputMBps/float64(pb.targets.GCPGCRThroughput))*100),
	}).Info("Benchmark results summary")

	// Log performance verdict
	if results.MeetsThroughputTarget && results.MeetsLatencyTarget {
		pb.logger.WithFields(map[string]interface{}{
			"scenario":      results.ScenarioName,
			"overall_score": results.OverallScore,
		}).Info("✅ BENCHMARK PASSED - Meets industry performance targets")
	} else {
		pb.logger.WithFields(map[string]interface{}{
			"scenario":            results.ScenarioName,
			"throughput_gap_mbps": results.ThroughputTarget - int64(results.ThroughputMBps),
			"latency_excess_ms":   results.LatencyMs - results.LatencyTarget,
			"overall_score":       results.OverallScore,
		}).Warn("❌ BENCHMARK FAILED - Does not meet industry performance targets")
	}
}
