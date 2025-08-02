package load

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
)

// TestLoadTestFrameworkIntegration tests the complete load testing framework
func TestLoadTestFrameworkIntegration(t *testing.T) {
	// Create temporary directory for test results
	tempDir, err := os.MkdirTemp("", "load_test_integration")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	logger := log.NewLogger()

	t.Run("ScenarioExecution", func(t *testing.T) {
		testScenarioExecution(t, tempDir, logger)
	})

	t.Run("BenchmarkSuite", func(t *testing.T) {
		testBenchmarkSuite(t, tempDir, logger)
	})

	t.Run("PrometheusIntegration", func(t *testing.T) {
		testPrometheusIntegration(t, tempDir, logger)
	})

	t.Run("RegressionTesting", func(t *testing.T) {
		testRegressionTesting(t, tempDir, logger)
	})

	t.Run("BaselineEstablishment", func(t *testing.T) {
		testBaselineEstablishment(t, tempDir, logger)
	})
}

func testScenarioExecution(t *testing.T, tempDir string, logger log.Logger) {
	// Test high-volume replication scenario
	scenario := CreateHighVolumeReplicationScenario()
	// Reduce duration for testing
	scenario.Duration = 30 * time.Second
	scenario.Images = scenario.Images[:10] // Limit to 10 images for testing

	runner := NewScenarioRunner(scenario, logger)
	result, err := runner.Run()

	if err != nil {
		t.Fatalf("Scenario execution failed: %v", err)
	}

	// Validate results
	if result.ProcessedImages == 0 {
		t.Error("No images were processed")
	}

	if result.AverageThroughputMBps <= 0 {
		t.Error("Throughput should be positive")
	}

	if result.MemoryUsageMB <= 0 {
		t.Error("Memory usage should be tracked")
	}

	t.Logf("Scenario completed: %d images processed, %.2f MB/s throughput",
		result.ProcessedImages, result.AverageThroughputMBps)
}

func testBenchmarkSuite(t *testing.T, tempDir string, logger log.Logger) {
	suite := NewBenchmarkSuite(tempDir, logger)

	// Create minimal scenarios for testing
	scenarios := []ScenarioConfig{
		CreateHighVolumeReplicationScenario(),
		CreateLargeImageStressScenario(),
	}

	// Reduce test duration
	for i := range scenarios {
		scenarios[i].Duration = 15 * time.Second
		if len(scenarios[i].Images) > 5 {
			scenarios[i].Images = scenarios[i].Images[:5]
		}
	}

	// Test Go benchmarks only (k6 and Apache Bench require external tools)
	results, err := suite.runGoBenchmarks()
	if err != nil {
		t.Fatalf("Go benchmarks failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("No benchmark results returned")
	}

	for _, result := range results {
		if result.Tool != "go-benchmark" {
			t.Errorf("Expected go-benchmark tool, got %s", result.Tool)
		}

		if result.Duration <= 0 {
			t.Error("Benchmark duration should be positive")
		}

		t.Logf("Benchmark completed: %s, duration: %v", result.Scenario, result.Duration)
	}
}

func testPrometheusIntegration(t *testing.T, tempDir string, logger log.Logger) {
	collector := NewPrometheusLoadTestCollector(":0", logger) // Use port 0 for testing

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start metrics server
	err := collector.StartMetricsServer(ctx)
	if err != nil {
		t.Fatalf("Failed to start metrics server: %v", err)
	}
	defer func() { _ = collector.StopMetricsServer() }()

	// Create test scenario result
	testResult := &LoadTestResults{
		ScenarioName:          "Test Scenario",
		Duration:              10 * time.Second,
		TotalImages:           100,
		ProcessedImages:       95,
		FailedImages:          5,
		AverageThroughputMBps: 125.5,
		PeakThroughputMBps:    150.0,
		MemoryUsageMB:         512,
		ConnectionReuseRate:   0.85,
		FailureRate:           0.05,
		ValidationPassed:      true,
	}

	// Record scenario execution
	collector.RecordScenarioExecution("Test Scenario", testResult)

	// Verify metrics were recorded
	collector.loadTestMetrics.mutex.RLock()
	executions := collector.loadTestMetrics.ScenarioExecutions["Test Scenario"]
	throughput := collector.loadTestMetrics.ThroughputMBps["Test Scenario"]
	collector.loadTestMetrics.mutex.RUnlock()

	if executions != 1 {
		t.Errorf("Expected 1 execution, got %d", executions)
	}

	if throughput != 125.5 {
		t.Errorf("Expected throughput 125.5, got %.1f", throughput)
	}

	t.Logf("Prometheus integration test completed successfully")
}

func testRegressionTesting(t *testing.T, tempDir string, logger log.Logger) {
	suite := NewRegressionTestSuite(tempDir, logger)

	// Create and save a baseline for testing
	baseline := BenchmarkResult{
		Tool:             "test",
		Scenario:         "Test Scenario",
		Timestamp:        time.Now(),
		Duration:         10 * time.Second,
		ThroughputMBps:   100.0,
		MemoryUsageMB:    400,
		FailureRate:      0.02,
		P99LatencyMs:     1000,
		ValidationPassed: true,
	}

	err := suite.baselineManager.UpdateBaseline("Test Scenario", baseline)
	if err != nil {
		t.Fatalf("Failed to save baseline: %v", err)
	}

	// Load baselines
	err = suite.baselineManager.LoadBaselines()
	if err != nil {
		t.Fatalf("Failed to load baselines: %v", err)
	}

	baselines := suite.baselineManager.GetCurrentBaselines()
	if len(baselines) == 0 {
		t.Error("No baselines loaded")
	}

	if _, exists := baselines["Test Scenario"]; !exists {
		t.Error("Test baseline not found")
	}

	t.Logf("Regression testing setup completed successfully")
}

func testBaselineEstablishment(t *testing.T, tempDir string, logger log.Logger) {
	suite := NewBaselineEstablishmentSuite(tempDir, logger)

	// Reduce configuration for testing
	suite.config.RunsPerScenario = 3
	suite.config.WarmupRuns = 1
	suite.config.CooldownPeriod = 1 * time.Second
	suite.config.SystemStabilization = 1 * time.Second
	suite.config.ValidationRuns = 2

	// Test statistical calculation methods
	testValues := []float64{100.0, 105.0, 98.0, 102.0, 99.0}
	stats := suite.calculatePerformanceStats(testValues)

	if stats.Mean <= 0 {
		t.Error("Mean should be positive")
	}

	if stats.StandardDeviation <= 0 {
		t.Error("Standard deviation should be positive")
	}

	if stats.Min > stats.Max {
		t.Error("Min should be less than max")
	}

	expectedMean := 100.8 // (100+105+98+102+99)/5
	if abs(stats.Mean-expectedMean) > 0.1 {
		t.Errorf("Expected mean ~%.1f, got %.1f", expectedMean, stats.Mean)
	}

	// Test outlier removal
	outlierValues := []LoadTestResults{
		{AverageThroughputMBps: 100.0},
		{AverageThroughputMBps: 105.0},
		{AverageThroughputMBps: 98.0},
		{AverageThroughputMBps: 200.0}, // Outlier
		{AverageThroughputMBps: 102.0},
	}

	filtered, outliersRemoved := suite.removeOutliers(outlierValues)

	if outliersRemoved != 1 {
		t.Errorf("Expected 1 outlier removed, got %d", outliersRemoved)
	}

	if len(filtered) != 4 {
		t.Errorf("Expected 4 filtered results, got %d", len(filtered))
	}

	t.Logf("Baseline establishment tests completed successfully")
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// BenchmarkLoadTestFramework benchmarks the load testing framework itself
func BenchmarkLoadTestFramework(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "load_test_benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	logger := log.NewLogger()

	b.Run("ScenarioCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			scenario := CreateHighVolumeReplicationScenario()
			if len(scenario.Images) == 0 {
				b.Error("No images in scenario")
			}
		}
	})

	b.Run("StatisticalCalculation", func(b *testing.B) {
		suite := NewBaselineEstablishmentSuite(tempDir, logger)
		values := make([]float64, 1000)
		for i := range values {
			values[i] = float64(i) + 100.0
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stats := suite.calculatePerformanceStats(values)
			if stats.Mean <= 0 {
				b.Error("Invalid mean calculated")
			}
		}
	})

	b.Run("MetricsCollection", func(b *testing.B) {
		collector := NewPrometheusLoadTestCollector(":0", logger)
		testResult := &LoadTestResults{
			ScenarioName:          "Benchmark Test",
			AverageThroughputMBps: 100.0,
			MemoryUsageMB:         400,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			collector.RecordScenarioExecution("Benchmark Test", testResult)
		}
	})
}

// TestLoadTestFrameworkStress performs stress testing of the framework
func TestLoadTestFrameworkStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "load_test_stress")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	logger := log.NewLogger()
	collector := NewPrometheusLoadTestCollector(":0", logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = collector.StartMetricsServer(ctx)
	if err != nil {
		t.Fatalf("Failed to start metrics server: %v", err)
	}
	defer func() { _ = collector.StopMetricsServer() }()

	// Simulate high-frequency metrics collection
	const numScenarios = 10
	const numRecords = 1000

	for scenario := 0; scenario < numScenarios; scenario++ {
		scenarioName := fmt.Sprintf("Stress Test Scenario %d", scenario)

		for record := 0; record < numRecords; record++ {
			testResult := &LoadTestResults{
				ScenarioName:          scenarioName,
				Duration:              time.Duration(record) * time.Millisecond,
				ProcessedImages:       int64(record),
				AverageThroughputMBps: float64(100 + record%50),
				MemoryUsageMB:         int64(400 + record%200),
				FailureRate:           float64(record%10) / 1000.0,
			}

			collector.RecordScenarioExecution(scenarioName, testResult)
		}
	}

	// Verify all metrics were recorded
	collector.loadTestMetrics.mutex.RLock()
	totalExecutions := int64(0)
	for _, executions := range collector.loadTestMetrics.ScenarioExecutions {
		totalExecutions += executions
	}
	collector.loadTestMetrics.mutex.RUnlock()

	expectedExecutions := int64(numScenarios * numRecords)
	if totalExecutions != expectedExecutions {
		t.Errorf("Expected %d total executions, got %d", expectedExecutions, totalExecutions)
	}

	t.Logf("Stress test completed: %d executions across %d scenarios", totalExecutions, numScenarios)
}

// TestLoadTestFrameworkConcurrency tests concurrent operations
func TestLoadTestFrameworkConcurrency(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "load_test_concurrency")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	logger := log.NewLogger()
	collector := NewPrometheusLoadTestCollector(":0", logger)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = collector.StartMetricsServer(ctx)
	if err != nil {
		t.Fatalf("Failed to start metrics server: %v", err)
	}
	defer func() { _ = collector.StopMetricsServer() }()

	// Test concurrent metrics recording
	const numGoroutines = 50
	const recordsPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			scenarioName := fmt.Sprintf("Concurrent Scenario %d", goroutineID)

			for j := 0; j < recordsPerGoroutine; j++ {
				testResult := &LoadTestResults{
					ScenarioName:          scenarioName,
					ProcessedImages:       int64(j),
					AverageThroughputMBps: float64(100 + j%20),
					MemoryUsageMB:         int64(400 + j%100),
				}

				collector.RecordScenarioExecution(scenarioName, testResult)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Goroutine completed
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for goroutines to complete")
		}
	}

	// Verify metrics consistency
	collector.loadTestMetrics.mutex.RLock()
	totalScenarios := len(collector.loadTestMetrics.ScenarioExecutions)
	collector.loadTestMetrics.mutex.RUnlock()

	if totalScenarios != numGoroutines {
		t.Errorf("Expected %d scenarios, got %d", numGoroutines, totalScenarios)
	}

	t.Logf("Concurrency test completed: %d goroutines, %d scenarios", numGoroutines, totalScenarios)
}

// TestLoadTestFrameworkResourceCleanup tests proper resource cleanup
func TestLoadTestFrameworkResourceCleanup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "load_test_cleanup")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	// Note: Not deferring cleanup to test manual cleanup

	logger := log.NewLogger()

	// Test baseline manager cleanup
	baselineManager := &BaselineManager{
		baselineDir:      tempDir,
		retentionPeriod:  1 * time.Hour,
		currentBaselines: make(map[string]BenchmarkResult),
		logger:           logger,
	}

	// Create test baseline files with different ages
	oldBaseline := BenchmarkResult{
		Tool:      "test",
		Scenario:  "Old Scenario",
		Timestamp: time.Now().Add(-2 * time.Hour), // 2 hours old
	}

	newBaseline := BenchmarkResult{
		Tool:      "test",
		Scenario:  "New Scenario",
		Timestamp: time.Now().Add(-30 * time.Minute), // 30 minutes old
	}

	err = baselineManager.UpdateBaseline("Old Scenario", oldBaseline)
	if err != nil {
		t.Fatalf("Failed to save old baseline: %v", err)
	}

	err = baselineManager.UpdateBaseline("New Scenario", newBaseline)
	if err != nil {
		t.Fatalf("Failed to save new baseline: %v", err)
	}

	// Verify files were created
	files, err := filepath.Glob(filepath.Join(tempDir, "baseline_*.json"))
	if err != nil {
		t.Fatalf("Failed to list baseline files: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 baseline files, got %d", len(files))
	}

	// Run cleanup
	baselineManager.cleanupOldBaselines()

	// Verify old file was removed
	files, err = filepath.Glob(filepath.Join(tempDir, "baseline_*.json"))
	if err != nil {
		t.Fatalf("Failed to list baseline files after cleanup: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 baseline file after cleanup, got %d", len(files))
	}

	// Manual cleanup
	_ = os.RemoveAll(tempDir)

	t.Logf("Resource cleanup test completed successfully")
}
