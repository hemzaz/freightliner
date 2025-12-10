package monitoring

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDefaultBenchmarkingConfig(t *testing.T) {
	config := DefaultBenchmarkingConfig()

	tests := []struct {
		name  string
		value interface{}
		check func() bool
	}{
		{"WarmupDuration", config.WarmupDuration, func() bool { return config.WarmupDuration > 0 }},
		{"BenchmarkDuration", config.BenchmarkDuration, func() bool { return config.BenchmarkDuration > 0 }},
		{"CooldownDuration", config.CooldownDuration, func() bool { return config.CooldownDuration > 0 }},
		{"MaxConcurrentBenchmarks", config.MaxConcurrentBenchmarks, func() bool { return config.MaxConcurrentBenchmarks > 0 }},
		{"MaxConcurrentOperations", config.MaxConcurrentOperations, func() bool { return config.MaxConcurrentOperations > 0 }},
		{"ThroughputTargetMBps", config.ThroughputTargetMBps, func() bool { return config.ThroughputTargetMBps > 0 }},
		{"LatencyTargetMs", config.LatencyTargetMs, func() bool { return config.LatencyTargetMs > 0 }},
		{"ReportInterval", config.ReportInterval, func() bool { return config.ReportInterval > 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Errorf("%s has invalid value: %v", tt.name, tt.value)
			}
		})
	}

	// Verify defaults
	if config.ThroughputTargetMBps != 125 {
		t.Errorf("ThroughputTargetMBps = %d, want 125", config.ThroughputTargetMBps)
	}

	if config.LatencyTargetMs != 50 {
		t.Errorf("LatencyTargetMs = %d, want 50", config.LatencyTargetMs)
	}

	if !config.EnableContinuous {
		t.Error("EnableContinuous should be true by default")
	}

	if !config.EnableRegression {
		t.Error("EnableRegression should be true by default")
	}
}

func TestGetIndustryBenchmarkTargets(t *testing.T) {
	targets := GetIndustryBenchmarkTargets()

	tests := []struct {
		name  string
		value int64
	}{
		{"DockerHubThroughput", targets.DockerHubThroughput},
		{"AWSECRThroughput", targets.AWSECRThroughput},
		{"GCPGCRThroughput", targets.GCPGCRThroughput},
		{"DockerHubLatency", targets.DockerHubLatency},
		{"AWSECRLatency", targets.AWSECRLatency},
		{"GCPGCRLatency", targets.GCPGCRLatency},
		{"MaxConcurrentConnections", targets.MaxConcurrentConnections},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value <= 0 {
				t.Errorf("%s = %d, want > 0", tt.name, tt.value)
			}
		})
	}

	// Verify specific values
	if targets.DockerHubThroughput != 150 {
		t.Errorf("DockerHubThroughput = %d, want 150", targets.DockerHubThroughput)
	}

	if targets.UptimeTarget != 99.9 {
		t.Errorf("UptimeTarget = %f, want 99.9", targets.UptimeTarget)
	}

	if targets.ErrorRateTarget != 0.1 {
		t.Errorf("ErrorRateTarget = %f, want 0.1", targets.ErrorRateTarget)
	}
}

func TestNewLatencyHistogram(t *testing.T) {
	lh := NewLatencyHistogram()

	if lh == nil {
		t.Fatal("NewLatencyHistogram() returned nil")
	}

	if len(lh.buckets) == 0 {
		t.Error("LatencyHistogram should have buckets")
	}

	if len(lh.bounds) == 0 {
		t.Error("LatencyHistogram should have bounds")
	}

	// Verify buckets match bounds + 1 (for overflow)
	if len(lh.buckets) != len(lh.bounds)+1 {
		t.Errorf("buckets length = %d, want %d", len(lh.buckets), len(lh.bounds)+1)
	}

	// Verify bounds are in ascending order
	for i := 1; i < len(lh.bounds); i++ {
		if lh.bounds[i] <= lh.bounds[i-1] {
			t.Errorf("bounds[%d] = %v should be > bounds[%d] = %v", i, lh.bounds[i], i-1, lh.bounds[i-1])
		}
	}
}

func TestLatencyHistogramRecord(t *testing.T) {
	lh := NewLatencyHistogram()

	testLatencies := []time.Duration{
		500 * time.Microsecond, // <1ms
		3 * time.Millisecond,   // <5ms
		15 * time.Millisecond,  // <25ms
		75 * time.Millisecond,  // <100ms
		200 * time.Millisecond, // <250ms
		10 * time.Second,       // overflow bucket
	}

	for _, latency := range testLatencies {
		lh.Record(latency)
	}

	// Verify total count
	total := int64(0)
	for i := range lh.buckets {
		total += lh.buckets[i].Load()
	}

	if total != int64(len(testLatencies)) {
		t.Errorf("total count = %d, want %d", total, len(testLatencies))
	}
}

func TestLatencyHistogramGetPercentile(t *testing.T) {
	lh := NewLatencyHistogram()

	// Record some latencies
	for i := 0; i < 100; i++ {
		if i < 50 {
			lh.Record(10 * time.Millisecond)
		} else if i < 95 {
			lh.Record(30 * time.Millisecond)
		} else {
			lh.Record(100 * time.Millisecond)
		}
	}

	// Test percentiles
	p50 := lh.GetPercentile(50)
	if p50 == 0 {
		t.Error("P50 should not be 0")
	}

	p95 := lh.GetPercentile(95)
	if p95 == 0 {
		t.Error("P95 should not be 0")
	}

	p99 := lh.GetPercentile(99)
	if p99 == 0 {
		t.Error("P99 should not be 0")
	}

	// P95 should be >= P50
	if p95 < p50 {
		t.Errorf("P95 (%v) should be >= P50 (%v)", p95, p50)
	}

	// P99 should be >= P95
	if p99 < p95 {
		t.Errorf("P99 (%v) should be >= P95 (%v)", p99, p95)
	}
}

func TestLatencyHistogramGetPercentileEmpty(t *testing.T) {
	lh := NewLatencyHistogram()

	p50 := lh.GetPercentile(50)
	if p50 != 0 {
		t.Errorf("percentile of empty histogram = %v, want 0", p50)
	}
}

func TestNewPerformanceBenchmarking(t *testing.T) {
	config := DefaultBenchmarkingConfig()
	pb := NewPerformanceBenchmarking(config, nil)

	if pb == nil {
		t.Fatal("NewPerformanceBenchmarking() returned nil")
	}

	if pb.logger == nil {
		t.Error("logger should not be nil (default logger created)")
	}

	if pb.metrics == nil {
		t.Error("metrics should not be nil")
	}

	if pb.metrics.latencyHistogram == nil {
		t.Error("latency histogram should not be nil")
	}

	if pb.runners == nil {
		t.Error("runners map should not be nil")
	}

	// Verify targets are initialized
	if pb.targets.DockerHubThroughput == 0 {
		t.Error("industry benchmark targets should be initialized")
	}
}

func TestPerformanceBenchmarkingStartStop(t *testing.T) {
	config := DefaultBenchmarkingConfig()
	config.EnableContinuous = false // Disable for testing
	pb := NewPerformanceBenchmarking(config, nil)

	// Test Start
	err := pb.Start()
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if !pb.started.Load() {
		t.Error("started flag should be true after Start()")
	}

	// Test double start (should not error)
	err = pb.Start()
	if err != nil {
		t.Errorf("second Start() should not error, got %v", err)
	}

	// Test Stop
	pb.Stop()
	if !pb.stopped.Load() {
		t.Error("stopped flag should be true after Stop()")
	}

	// Test double stop (should not panic)
	pb.Stop()
}

func TestResetMetrics(t *testing.T) {
	config := DefaultBenchmarkingConfig()
	pb := NewPerformanceBenchmarking(config, nil)

	// Set some values
	pb.metrics.CurrentThroughputMBps.Store(100)
	pb.metrics.TotalOperations.Store(1000)
	pb.metrics.MaxLatencyMs.Store(500)

	// Reset
	pb.resetMetrics()

	// Verify reset
	if pb.metrics.CurrentThroughputMBps.Load() != 0 {
		t.Error("CurrentThroughputMBps should be 0 after reset")
	}

	if pb.metrics.TotalOperations.Load() != 0 {
		t.Error("TotalOperations should be 0 after reset")
	}

	if pb.metrics.MaxLatencyMs.Load() != 0 {
		t.Error("MaxLatencyMs should be 0 after reset")
	}
}

func TestCalculateOverallScore(t *testing.T) {
	config := DefaultBenchmarkingConfig()
	pb := NewPerformanceBenchmarking(config, nil)

	tests := []struct {
		name    string
		results *BenchmarkResults
		wantMin float64
		wantMax float64
	}{
		{
			name: "perfect performance",
			results: &BenchmarkResults{
				ThroughputMBps: 150,
				LatencyMs:      25,
				ErrorRate:      0,
				PeakCPUUsage:   20,
			},
			wantMin: 80,
			wantMax: 100,
		},
		{
			name: "poor performance",
			results: &BenchmarkResults{
				ThroughputMBps: 10,
				LatencyMs:      200,
				ErrorRate:      5,
				PeakCPUUsage:   95,
			},
			wantMin: 0,
			wantMax: 30,
		},
		{
			name: "moderate performance",
			results: &BenchmarkResults{
				ThroughputMBps: 75,
				LatencyMs:      75,
				ErrorRate:      0.5,
				PeakCPUUsage:   50,
			},
			wantMin: 20,
			wantMax: 70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := pb.calculateOverallScore(tt.results)

			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("calculateOverallScore() = %f, want between %f and %f", score, tt.wantMin, tt.wantMax)
			}

			// Score should be 0-100
			if score < 0 || score > 100 {
				t.Errorf("calculateOverallScore() = %f, want 0-100", score)
			}
		})
	}
}

func TestGetLatencyDistribution(t *testing.T) {
	config := DefaultBenchmarkingConfig()
	pb := NewPerformanceBenchmarking(config, nil)

	// Record some latencies
	pb.metrics.latencyHistogram.Record(5 * time.Millisecond)
	pb.metrics.latencyHistogram.Record(50 * time.Millisecond)
	pb.metrics.latencyHistogram.Record(500 * time.Millisecond)

	dist := pb.getLatencyDistribution()

	if len(dist) == 0 {
		t.Error("getLatencyDistribution() returned empty map")
	}

	// Verify expected keys exist
	expectedKeys := []string{"<1ms", "<5ms", "<10ms", "<25ms", "<50ms", "<100ms"}
	for _, key := range expectedKeys {
		if _, ok := dist[key]; !ok {
			t.Errorf("distribution missing key: %s", key)
		}
	}

	// Verify total count
	total := int64(0)
	for _, count := range dist {
		total += count
	}
	if total != 3 {
		t.Errorf("total distribution count = %d, want 3", total)
	}
}

func TestCollectPerformanceSnapshot(t *testing.T) {
	config := DefaultBenchmarkingConfig()
	pb := NewPerformanceBenchmarking(config, nil)

	// Set some metrics
	pb.metrics.CurrentThroughputMBps.Store(100)
	pb.metrics.CurrentLatencyMs.Store(50)
	pb.metrics.ConcurrentOps.Store(10)
	pb.metrics.TotalOperations.Store(1000)
	pb.metrics.FailedOps.Store(10)

	// Collect snapshot
	pb.collectPerformanceSnapshot()

	// Verify snapshot was added
	pb.metrics.mutex.RLock()
	snapshots := pb.metrics.snapshots
	pb.metrics.mutex.RUnlock()

	if len(snapshots) != 1 {
		t.Errorf("snapshots length = %d, want 1", len(snapshots))
	}

	snapshot := snapshots[0]
	if snapshot.ThroughputMBps != 100 {
		t.Errorf("snapshot.ThroughputMBps = %d, want 100", snapshot.ThroughputMBps)
	}

	if snapshot.LatencyMs != 50 {
		t.Errorf("snapshot.LatencyMs = %d, want 50", snapshot.LatencyMs)
	}

	// Verify error rate calculation
	expectedErrorRate := (10.0 / 1000.0) * 100.0
	if snapshot.ErrorRate != expectedErrorRate {
		t.Errorf("snapshot.ErrorRate = %f, want %f", snapshot.ErrorRate, expectedErrorRate)
	}
}

func TestCollectPerformanceSnapshotLimit(t *testing.T) {
	config := DefaultBenchmarkingConfig()
	pb := NewPerformanceBenchmarking(config, nil)

	// Add more than 1000 snapshots
	for i := 0; i < 1100; i++ {
		pb.collectPerformanceSnapshot()
	}

	// Verify limit is enforced
	pb.metrics.mutex.RLock()
	snapshotCount := len(pb.metrics.snapshots)
	pb.metrics.mutex.RUnlock()

	if snapshotCount > 1000 {
		t.Errorf("snapshots count = %d, want <= 1000", snapshotCount)
	}
}

func TestRunBenchmarkWithMockScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark test in short mode")
	}

	config := DefaultBenchmarkingConfig()
	config.WarmupDuration = 10 * time.Millisecond
	config.BenchmarkDuration = 50 * time.Millisecond
	config.CooldownDuration = 10 * time.Millisecond
	config.EnableContinuous = false

	pb := NewPerformanceBenchmarking(config, nil)
	err := pb.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer pb.Stop()

	// Create mock scenario
	scenario := BenchmarkScenario{
		Name:             "test_scenario",
		Description:      "Test scenario",
		TargetThroughput: 100,
		TargetLatency:    50,
		ConcurrencyLevel: 1,
		Duration:         50 * time.Millisecond,
		OperationType:    "test",
		ExecuteFunc: func(ctx context.Context, config ScenarioConfig) (*BenchmarkResults, error) {
			// Simulate work
			time.Sleep(10 * time.Millisecond)

			return &BenchmarkResults{
				ScenarioName:    "test_scenario",
				StartTime:       time.Now().Add(-50 * time.Millisecond),
				EndTime:         time.Now(),
				Duration:        50 * time.Millisecond,
				ThroughputMBps:  100,
				LatencyMs:       50,
				TotalOperations: 100,
				SuccessfulOps:   95,
				FailedOps:       5,
				ErrorRate:       5.0,
			}, nil
		},
	}

	results, err := pb.RunBenchmark(scenario)
	if err != nil {
		t.Errorf("RunBenchmark() error = %v", err)
	}

	if results == nil {
		t.Fatal("RunBenchmark() returned nil results")
	}

	// Verify results structure
	if results.ScenarioName != "test_scenario" {
		t.Errorf("results.ScenarioName = %s, want test_scenario", results.ScenarioName)
	}

	if results.ThroughputTarget != 100 {
		t.Errorf("results.ThroughputTarget = %d, want 100", results.ThroughputTarget)
	}

	if results.LatencyTarget != 50 {
		t.Errorf("results.LatencyTarget = %d, want 50", results.LatencyTarget)
	}

	if results.OverallScore < 0 || results.OverallScore > 100 {
		t.Errorf("results.OverallScore = %f, want 0-100", results.OverallScore)
	}
}

func TestFinalizeResults(t *testing.T) {
	config := DefaultBenchmarkingConfig()
	pb := NewPerformanceBenchmarking(config, nil)

	// Setup metrics
	pb.metrics.latencyHistogram.Record(10 * time.Millisecond)
	pb.metrics.latencyHistogram.Record(50 * time.Millisecond)
	pb.metrics.latencyHistogram.Record(100 * time.Millisecond)

	runner := &BenchmarkRunner{
		scenario: BenchmarkScenario{
			TargetThroughput: 100,
			TargetLatency:    50,
		},
	}

	results := &BenchmarkResults{
		ThroughputMBps: 120,
		LatencyMs:      45,
		ErrorRate:      1.0,
		PeakCPUUsage:   60,
	}

	finalized := pb.finalizeResults(runner, results)

	// Verify percentiles are calculated
	if finalized.P50LatencyMs == 0 {
		t.Error("P50LatencyMs should be calculated")
	}

	if finalized.P95LatencyMs == 0 {
		t.Error("P95LatencyMs should be calculated")
	}

	if finalized.P99LatencyMs == 0 {
		t.Error("P99LatencyMs should be calculated")
	}

	// Verify targets are set
	if finalized.ThroughputTarget != 100 {
		t.Error("ThroughputTarget should be set from scenario")
	}

	if finalized.LatencyTarget != 50 {
		t.Error("LatencyTarget should be set from scenario")
	}

	// Verify target checks
	if !finalized.MeetsThroughputTarget {
		t.Error("MeetsThroughputTarget should be true (120 >= 100)")
	}

	if !finalized.MeetsLatencyTarget {
		t.Error("MeetsLatencyTarget should be true (45 <= 50)")
	}

	// Verify score is calculated
	if finalized.OverallScore == 0 {
		t.Error("OverallScore should be calculated")
	}

	// Verify latency distribution is set
	if len(finalized.LatencyDistribution) == 0 {
		t.Error("LatencyDistribution should be populated")
	}
}

func BenchmarkNewLatencyHistogram(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLatencyHistogram()
	}
}

func BenchmarkLatencyHistogramRecord(b *testing.B) {
	lh := NewLatencyHistogram()
	latency := 50 * time.Millisecond
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lh.Record(latency)
	}
}

func BenchmarkLatencyHistogramGetPercentile(b *testing.B) {
	lh := NewLatencyHistogram()
	for i := 0; i < 1000; i++ {
		lh.Record(time.Duration(i) * time.Millisecond)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lh.GetPercentile(95)
	}
}

func TestReportCurrentMetrics(t *testing.T) {
	config := DefaultBenchmarkingConfig()
	pb := NewPerformanceBenchmarking(config, nil)

	// Set some metrics
	pb.metrics.CurrentThroughputMBps.Store(100)
	pb.metrics.CurrentLatencyMs.Store(50)
	pb.metrics.ConcurrentOps.Store(10)
	pb.metrics.TotalOperations.Store(1000)
	pb.metrics.SuccessfulOps.Store(990)
	pb.metrics.FailedOps.Store(10)

	// Call should not panic
	pb.reportCurrentMetrics()
}

func TestLogBenchmarkResults(t *testing.T) {
	config := DefaultBenchmarkingConfig()
	pb := NewPerformanceBenchmarking(config, nil)

	results := &BenchmarkResults{
		ScenarioName:            "test",
		Duration:                time.Minute,
		ThroughputMBps:          120,
		ThroughputTarget:        100,
		MeetsThroughputTarget:   true,
		LatencyMs:               45,
		LatencyTarget:           50,
		MeetsLatencyTarget:      true,
		P99LatencyMs:            95,
		P95LatencyMs:            85,
		P50LatencyMs:            40,
		TotalOperations:         1000,
		SuccessfulOps:           990,
		FailedOps:               10,
		ErrorRate:               1.0,
		PeakCPUUsage:            60,
		PeakMemoryUsage:         512,
		NetworkBytesTransferred: 1000000,
		OverallScore:            85.5,
	}

	// Call should not panic
	pb.logBenchmarkResults(results)
}

func TestContinuousMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping continuous monitoring test in short mode")
	}

	config := DefaultBenchmarkingConfig()
	config.ReportInterval = 50 * time.Millisecond
	config.EnableContinuous = true

	pb := NewPerformanceBenchmarking(config, nil)

	err := pb.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Let it run for a bit
	time.Sleep(150 * time.Millisecond)

	// Stop it
	pb.Stop()

	// Give it time to finish
	time.Sleep(50 * time.Millisecond)
}

func TestExecuteBenchmarkWithPhases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark phases test in short mode")
	}

	config := DefaultBenchmarkingConfig()
	config.WarmupDuration = 20 * time.Millisecond
	config.BenchmarkDuration = 50 * time.Millisecond
	config.CooldownDuration = 20 * time.Millisecond

	pb := NewPerformanceBenchmarking(config, nil)

	executionCount := 0
	scenario := BenchmarkScenario{
		Name:             "test_phases",
		TargetThroughput: 100,
		TargetLatency:    50,
		ConcurrencyLevel: 1,
		Duration:         50 * time.Millisecond,
		ExecuteFunc: func(ctx context.Context, cfg ScenarioConfig) (*BenchmarkResults, error) {
			executionCount++
			// Simulate some work
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(10 * time.Millisecond):
			}

			return &BenchmarkResults{
				ScenarioName:    "test_phases",
				StartTime:       time.Now(),
				EndTime:         time.Now(),
				Duration:        cfg.Duration,
				ThroughputMBps:  100,
				LatencyMs:       50,
				TotalOperations: 100,
				SuccessfulOps:   100,
				FailedOps:       0,
				ErrorRate:       0,
			}, nil
		},
	}

	runner := &BenchmarkRunner{
		name:     scenario.Name,
		scenario: scenario,
		metrics:  pb.metrics,
		logger:   pb.logger,
	}

	ctx := context.Background()
	results, err := pb.executeBenchmarkWithPhases(ctx, runner)

	if err != nil {
		t.Errorf("executeBenchmarkWithPhases() error = %v", err)
	}

	if results == nil {
		t.Fatal("executeBenchmarkWithPhases() returned nil results")
	}

	// Should have executed twice: once for warmup, once for actual benchmark
	if executionCount != 2 {
		t.Errorf("executionCount = %d, want 2 (warmup + benchmark)", executionCount)
	}
}

func TestExecuteBenchmarkWithPhasesError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark error test in short mode")
	}

	config := DefaultBenchmarkingConfig()
	config.WarmupDuration = 20 * time.Millisecond
	config.BenchmarkDuration = 50 * time.Millisecond
	config.CooldownDuration = 20 * time.Millisecond

	pb := NewPerformanceBenchmarking(config, nil)

	testErr := errors.New("execution failed")
	scenario := BenchmarkScenario{
		Name:     "test_error",
		Duration: 50 * time.Millisecond,
		ExecuteFunc: func(ctx context.Context, cfg ScenarioConfig) (*BenchmarkResults, error) {
			return nil, testErr
		},
	}

	runner := &BenchmarkRunner{
		name:     scenario.Name,
		scenario: scenario,
		metrics:  pb.metrics,
		logger:   pb.logger,
	}

	ctx := context.Background()
	_, err := pb.executeBenchmarkWithPhases(ctx, runner)

	if err == nil {
		t.Error("executeBenchmarkWithPhases() should return error")
	}

	if !strings.Contains(err.Error(), "warmup phase failed") {
		t.Errorf("error should mention warmup phase, got: %v", err)
	}
}

func TestRunBenchmarkIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full benchmark integration test in short mode")
	}

	config := DefaultBenchmarkingConfig()
	config.WarmupDuration = 10 * time.Millisecond
	config.BenchmarkDuration = 50 * time.Millisecond
	config.CooldownDuration = 10 * time.Millisecond
	config.EnableContinuous = false

	pb := NewPerformanceBenchmarking(config, nil)
	err := pb.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer pb.Stop()

	scenario := BenchmarkScenario{
		Name:             "integration_test",
		Description:      "Integration test scenario",
		TargetThroughput: 100,
		TargetLatency:    50,
		ConcurrencyLevel: 2,
		Duration:         50 * time.Millisecond,
		OperationType:    "test",
		PayloadSizeMB:    10,
		NumberOfAssets:   100,
		ExecuteFunc: func(ctx context.Context, cfg ScenarioConfig) (*BenchmarkResults, error) {
			// Record some latencies for distribution
			cfg.MetricsCollector.latencyHistogram.Record(10 * time.Millisecond)
			cfg.MetricsCollector.latencyHistogram.Record(50 * time.Millisecond)
			cfg.MetricsCollector.latencyHistogram.Record(100 * time.Millisecond)

			// Simulate work
			time.Sleep(20 * time.Millisecond)

			return &BenchmarkResults{
				ScenarioName:            "integration_test",
				StartTime:               time.Now().Add(-50 * time.Millisecond),
				EndTime:                 time.Now(),
				Duration:                50 * time.Millisecond,
				ThroughputMBps:          110,
				LatencyMs:               45,
				TotalOperations:         200,
				SuccessfulOps:           195,
				FailedOps:               5,
				ErrorRate:               2.5,
				PeakCPUUsage:            55,
				PeakMemoryUsage:         256,
				NetworkBytesTransferred: 1048576,
			}, nil
		},
	}

	results, err := pb.RunBenchmark(scenario)
	if err != nil {
		t.Errorf("RunBenchmark() error = %v", err)
	}

	if results == nil {
		t.Fatal("RunBenchmark() returned nil results")
	}

	// Verify results are properly finalized
	if results.P50LatencyMs == 0 {
		t.Error("P50LatencyMs should be calculated")
	}

	if results.ThroughputTarget != 100 {
		t.Error("ThroughputTarget should be set")
	}

	if results.LatencyTarget != 50 {
		t.Error("LatencyTarget should be set")
	}

	if results.OverallScore == 0 {
		t.Error("OverallScore should be calculated")
	}

	if len(results.LatencyDistribution) == 0 {
		t.Error("LatencyDistribution should be populated")
	}

	// Check that targets are properly evaluated
	if !results.MeetsThroughputTarget {
		t.Error("Should meet throughput target (110 >= 100)")
	}

	if !results.MeetsLatencyTarget {
		t.Error("Should meet latency target (45 <= 50)")
	}
}

func BenchmarkCalculateOverallScore(b *testing.B) {
	config := DefaultBenchmarkingConfig()
	pb := NewPerformanceBenchmarking(config, nil)
	results := &BenchmarkResults{
		ThroughputMBps: 100,
		LatencyMs:      50,
		ErrorRate:      0.5,
		PeakCPUUsage:   50,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pb.calculateOverallScore(results)
	}
}
