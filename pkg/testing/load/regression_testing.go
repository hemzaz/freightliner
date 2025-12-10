package load

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
)

// RegressionTestSuite manages automated performance regression testing
type RegressionTestSuite struct {
	logger          log.Logger
	baselineManager *BaselineManager
	testScheduler   *TestScheduler
	alertManager    *AlertManager

	// Configuration
	config       RegressionConfig
	testResults  []RegressionTestResult
	resultsMutex sync.RWMutex

	// Integration with other components
	scenarioRunner      *ScenarioRunner
	benchmarkSuite      *BenchmarkSuite
	prometheusCollector *PrometheusLoadTestCollector
}

// RegressionConfig defines configuration for regression testing
type RegressionConfig struct {
	// Test scheduling
	ScheduleInterval    time.Duration `json:"schedule_interval"`
	TriggerOnCodeChange bool          `json:"trigger_on_code_change"`
	TriggerOnBaseline   bool          `json:"trigger_on_baseline"`

	// Regression thresholds
	ThroughputRegressionThreshold  float64 `json:"throughput_regression_threshold"`
	LatencyRegressionThreshold     float64 `json:"latency_regression_threshold"`
	MemoryRegressionThreshold      float64 `json:"memory_regression_threshold"`
	FailureRateRegressionThreshold float64 `json:"failure_rate_regression_threshold"`

	// Test execution
	ParallelExecution bool          `json:"parallel_execution"`
	MaxParallelTests  int           `json:"max_parallel_tests"`
	TestTimeout       time.Duration `json:"test_timeout"`
	RetryFailedTests  bool          `json:"retry_failed_tests"`
	MaxRetries        int           `json:"max_retries"`

	// Baseline management
	BaselineRetentionDays   int     `json:"baseline_retention_days"`
	AutoUpdateBaseline      bool    `json:"auto_update_baseline"`
	BaselineUpdateThreshold float64 `json:"baseline_update_threshold"`

	// Reporting
	GenerateReports      bool     `json:"generate_reports"`
	ReportFormats        []string `json:"report_formats"`        // "json", "html", "csv"
	NotificationChannels []string `json:"notification_channels"` // "email", "slack", "webhook"
}

// RegressionTestResult contains the results of a regression test run
type RegressionTestResult struct {
	TestID        string    `json:"test_id"`
	Timestamp     time.Time `json:"timestamp"`
	TriggerReason string    `json:"trigger_reason"`
	GitCommit     string    `json:"git_commit,omitempty"`
	GitBranch     string    `json:"git_branch,omitempty"`

	// Test execution details
	ScenariosRun    int           `json:"scenarios_run"`
	TotalDuration   time.Duration `json:"total_duration"`
	ExecutionStatus string        `json:"execution_status"` // "success", "failed", "partial"

	// Performance comparison
	BaselineVersion  string                   `json:"baseline_version"`
	RegressionIssues []RegressionIssue        `json:"regression_issues"`
	Improvements     []PerformanceImprovement `json:"improvements"`

	// Detailed results per scenario
	ScenarioResults map[string]ScenarioRegressionResult `json:"scenario_results"`

	// Overall regression assessment
	RegressionSeverity string   `json:"regression_severity"` // "none", "minor", "major", "critical"
	RecommendedActions []string `json:"recommended_actions"`
	AlertsTriggered    []string `json:"alerts_triggered"`
}

// ScenarioRegressionResult contains regression analysis for a specific scenario
type ScenarioRegressionResult struct {
	ScenarioName      string            `json:"scenario_name"`
	CurrentResult     *LoadTestResults  `json:"current_result"`
	BaselineResult    *BenchmarkResult  `json:"baseline_result"`
	RegressionMetrics RegressionMetrics `json:"regression_metrics"`
	ValidationStatus  string            `json:"validation_status"` // "pass", "fail", "warning"
	PerformanceTrend  string            `json:"performance_trend"` // "improving", "stable", "degrading"
}

// PerformanceImprovement represents a positive performance change
type PerformanceImprovement struct {
	Metric             string  `json:"metric"`
	ImprovementPercent float64 `json:"improvement_percent"`
	Description        string  `json:"description"`
}

// BaselineManager handles performance baseline storage and management
type BaselineManager struct {
	baselineDir      string
	retentionPeriod  time.Duration
	currentBaselines map[string]BenchmarkResult
	mutex            sync.RWMutex
	logger           log.Logger
}

// TestScheduler manages automated test scheduling
type TestScheduler struct {
	config    RegressionConfig
	testSuite *RegressionTestSuite
	logger    log.Logger

	// Scheduling state
	lastRun   time.Time
	isRunning bool
	stopChan  chan struct{}
	mutex     sync.Mutex
}

// AlertManager handles regression test alerts and notifications
type AlertManager struct {
	config RegressionConfig
	logger log.Logger

	// Alert channels
	emailChannel   EmailNotifier
	slackChannel   SlackNotifier
	webhookChannel WebhookNotifier
}

// Notification interfaces
type EmailNotifier interface {
	SendRegressionAlert(result RegressionTestResult) error
}

type SlackNotifier interface {
	PostRegressionAlert(result RegressionTestResult) error
}

type WebhookNotifier interface {
	SendWebhookAlert(result RegressionTestResult) error
}

// NewRegressionTestSuite creates a new regression test suite
func NewRegressionTestSuite(baselineDir string, logger log.Logger) *RegressionTestSuite {
	if logger == nil {
		logger = log.NewLogger()
	}

	config := RegressionConfig{
		ScheduleInterval:               24 * time.Hour, // Daily by default
		TriggerOnCodeChange:            true,
		TriggerOnBaseline:              true,
		ThroughputRegressionThreshold:  10.0, // 10% throughput drop
		LatencyRegressionThreshold:     20.0, // 20% latency increase
		MemoryRegressionThreshold:      15.0, // 15% memory increase
		FailureRateRegressionThreshold: 0.05, // 5% failure rate increase
		ParallelExecution:              true,
		MaxParallelTests:               3,
		TestTimeout:                    2 * time.Hour,
		RetryFailedTests:               true,
		MaxRetries:                     2,
		BaselineRetentionDays:          30,
		AutoUpdateBaseline:             false,
		BaselineUpdateThreshold:        5.0, // 5% improvement to update baseline
		GenerateReports:                true,
		ReportFormats:                  []string{"json", "html"},
		NotificationChannels:           []string{"slack"},
	}

	// Initialize BaselineManager with proper logger
	baselineManager := &BaselineManager{
		baselineDir:      baselineDir,
		retentionPeriod:  time.Duration(config.BaselineRetentionDays) * 24 * time.Hour,
		currentBaselines: make(map[string]BenchmarkResult),
		logger:           logger,
	}

	suite := &RegressionTestSuite{
		logger:          logger,
		baselineManager: baselineManager,
		config:          config,
		testResults:     make([]RegressionTestResult, 0),
	}

	suite.testScheduler = &TestScheduler{
		config:    config,
		testSuite: suite,
		logger:    logger,
		stopChan:  make(chan struct{}),
	}

	suite.alertManager = &AlertManager{
		config: config,
		logger: logger,
	}

	return suite
}

// StartAutomatedTesting begins automated regression testing
func (rts *RegressionTestSuite) StartAutomatedTesting(ctx context.Context) error {
	rts.logger.WithFields(map[string]interface{}{
		"schedule_interval": rts.config.ScheduleInterval.String(),
		"baseline_dir":      rts.baselineManager.baselineDir,
		"parallel_tests":    rts.config.MaxParallelTests,
	}).Info("Starting automated regression testing")

	// Load existing baselines
	if err := rts.baselineManager.LoadBaselines(); err != nil {
		rts.logger.WithFields(map[string]interface{}{"error": err.Error()}).Warn("Failed to load baselines")
	}

	// Start the test scheduler
	go rts.testScheduler.Start(ctx)

	// Start baseline cleanup routine
	go rts.baselineManager.StartCleanupRoutine(ctx)

	return nil
}

// StopAutomatedTesting stops automated regression testing
func (rts *RegressionTestSuite) StopAutomatedTesting() error {
	rts.logger.Info("Stopping automated regression testing")
	rts.testScheduler.Stop()
	return nil
}

// RunRegressionTest executes a single regression test run
func (rts *RegressionTestSuite) RunRegressionTest(ctx context.Context, triggerReason string) (*RegressionTestResult, error) {
	testID := fmt.Sprintf("regression_%s_%d", triggerReason, time.Now().Unix())

	rts.logger.Info("Starting regression test", map[string]interface{}{
		"test_id":        testID,
		"trigger_reason": triggerReason,
	})

	result := &RegressionTestResult{
		TestID:          testID,
		Timestamp:       time.Now(),
		TriggerReason:   triggerReason,
		ScenarioResults: make(map[string]ScenarioRegressionResult),
	}

	// Get git information if available
	result.GitCommit, result.GitBranch = rts.getGitInfo()

	// Get current baselines
	baselines := rts.baselineManager.GetCurrentBaselines()
	if len(baselines) == 0 {
		return nil, fmt.Errorf("no baselines available for regression testing - baselines must be established manually in production/staging environments. " +
			"To establish baselines: 1) Deploy to production/staging, 2) Run baseline establishment tool, 3) Commit baseline files to repository")
	}

	// Create scenarios based on baselines
	scenarios := rts.createScenariosFromBaselines(baselines)

	startTime := time.Now()

	// Execute scenarios
	if rts.config.ParallelExecution {
		result.ScenarioResults = rts.runScenariosParallel(ctx, scenarios, baselines)
	} else {
		result.ScenarioResults = rts.runScenariosSequential(ctx, scenarios, baselines)
	}

	result.TotalDuration = time.Since(startTime)
	result.ScenariosRun = len(result.ScenarioResults)

	// Analyze results for regressions
	rts.analyzeRegressions(result)

	// Determine overall status
	result.ExecutionStatus = rts.determineExecutionStatus(result)

	// Generate recommendations
	result.RecommendedActions = rts.generateRecommendations(result)

	// Store results
	rts.resultsMutex.Lock()
	rts.testResults = append(rts.testResults, *result)
	rts.resultsMutex.Unlock()

	// Save results to disk
	if err := rts.saveRegressionResults(result); err != nil {
		rts.logger.Error("Failed to save regression results", err, map[string]interface{}{})
	}

	// Send alerts if necessary
	if result.RegressionSeverity != "none" {
		rts.alertManager.SendRegressionAlerts(*result)
	}

	// Update baselines if configured and results are better
	if rts.config.AutoUpdateBaseline {
		rts.updateBaselinesIfImproved(result)
	}

	rts.logger.Info("Regression test completed", map[string]interface{}{
		"test_id":             testID,
		"duration":            result.TotalDuration.String(),
		"scenarios_run":       result.ScenariosRun,
		"regression_severity": result.RegressionSeverity,
		"execution_status":    result.ExecutionStatus,
	})

	return result, nil
}

// runScenariosParallel executes scenarios in parallel
func (rts *RegressionTestSuite) runScenariosParallel(ctx context.Context, scenarios []ScenarioConfig, baselines map[string]BenchmarkResult) map[string]ScenarioRegressionResult {
	results := make(map[string]ScenarioRegressionResult)
	resultsChan := make(chan ScenarioRegressionResult, len(scenarios))
	semaphore := make(chan struct{}, rts.config.MaxParallelTests)

	var wg sync.WaitGroup

	for _, scenario := range scenarios {
		wg.Add(1)
		go func(s ScenarioConfig) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := rts.runSingleScenarioRegression(ctx, s, baselines[s.Name])
			resultsChan <- result
		}(scenario)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		results[result.ScenarioName] = result
	}

	return results
}

// runScenariosSequential executes scenarios one at time
func (rts *RegressionTestSuite) runScenariosSequential(ctx context.Context, scenarios []ScenarioConfig, baselines map[string]BenchmarkResult) map[string]ScenarioRegressionResult {
	results := make(map[string]ScenarioRegressionResult)

	for _, scenario := range scenarios {
		result := rts.runSingleScenarioRegression(ctx, scenario, baselines[scenario.Name])
		results[result.ScenarioName] = result
	}

	return results
}

// runSingleScenarioRegression executes a single scenario for regression testing
func (rts *RegressionTestSuite) runSingleScenarioRegression(ctx context.Context, scenario ScenarioConfig, baseline BenchmarkResult) ScenarioRegressionResult {
	rts.logger.Info("Running scenario regression test", map[string]interface{}{
		"scenario": scenario.Name,
	})

	// Execute the scenario
	runner := NewScenarioRunner(scenario, rts.logger)
	currentResult, err := runner.Run()

	result := ScenarioRegressionResult{
		ScenarioName:   scenario.Name,
		CurrentResult:  currentResult,
		BaselineResult: &baseline,
	}

	if err != nil {
		rts.logger.Error("Scenario execution failed", err, map[string]interface{}{
			"scenario": scenario.Name,
		})
		result.ValidationStatus = "fail"
		return result
	}

	// Compare with baseline
	result.RegressionMetrics = rts.compareWithBaseline(currentResult, baseline)

	// Determine validation status
	result.ValidationStatus = rts.determineValidationStatus(result.RegressionMetrics)

	// Analyze performance trend
	result.PerformanceTrend = rts.analyzePerformanceTrend(scenario.Name, currentResult)

	return result
}

// compareWithBaseline compares current results with baseline
func (rts *RegressionTestSuite) compareWithBaseline(current *LoadTestResults, baseline BenchmarkResult) RegressionMetrics {
	regression := RegressionMetrics{}

	// Calculate throughput change
	if baseline.ThroughputMBps > 0 {
		regression.ThroughputChange = ((current.AverageThroughputMBps - baseline.ThroughputMBps) / baseline.ThroughputMBps) * 100
	}

	// Calculate latency change
	if baseline.P99LatencyMs > 0 && current.P99LatencyMs > 0 {
		regression.LatencyChange = ((float64(current.P99LatencyMs) - float64(baseline.P99LatencyMs)) / float64(baseline.P99LatencyMs)) * 100
	}

	// Calculate memory change
	if baseline.MemoryUsageMB > 0 {
		regression.MemoryChange = ((float64(current.MemoryUsageMB) - float64(baseline.MemoryUsageMB)) / float64(baseline.MemoryUsageMB)) * 100
	}

	// Calculate failure rate change
	regression.FailureRateChange = (current.FailureRate - baseline.FailureRate) * 100

	// Determine severity
	regression.Severity = rts.calculateRegressionSeverity(regression)

	return regression
}

// analyzeRegressions analyzes all scenario results for regressions
func (rts *RegressionTestSuite) analyzeRegressions(result *RegressionTestResult) {
	var regressions []RegressionIssue
	var improvements []PerformanceImprovement

	maxSeverity := "none"

	for _, scenarioResult := range result.ScenarioResults {
		regression := scenarioResult.RegressionMetrics

		// Check for regressions
		if rts.isRegression(regression) {
			issue := RegressionIssue{
				Benchmark: scenarioResult.ScenarioName,
				Severity:  regression.Severity,
			}

			if regression.ThroughputChange < -rts.config.ThroughputRegressionThreshold {
				issue.Metric = "throughput"
				issue.RegressionPercent = -regression.ThroughputChange
				regressions = append(regressions, issue)
			}

			if regression.LatencyChange > rts.config.LatencyRegressionThreshold {
				issue.Metric = "latency"
				issue.RegressionPercent = regression.LatencyChange
				regressions = append(regressions, issue)
			}

			if regression.MemoryChange > rts.config.MemoryRegressionThreshold {
				issue.Metric = "memory"
				issue.RegressionPercent = regression.MemoryChange
				regressions = append(regressions, issue)
			}

			// Update max severity
			if rts.compareSeverity(regression.Severity, maxSeverity) > 0 {
				maxSeverity = regression.Severity
			}
		}

		// Check for improvements
		if rts.isImprovement(regression) {
			if regression.ThroughputChange > 5.0 {
				improvements = append(improvements, PerformanceImprovement{
					Metric:             "throughput",
					ImprovementPercent: regression.ThroughputChange,
					Description:        fmt.Sprintf("Throughput improved by %.1f%% in %s", regression.ThroughputChange, scenarioResult.ScenarioName),
				})
			}

			if regression.LatencyChange < -5.0 {
				improvements = append(improvements, PerformanceImprovement{
					Metric:             "latency",
					ImprovementPercent: -regression.LatencyChange,
					Description:        fmt.Sprintf("Latency improved by %.1f%% in %s", -regression.LatencyChange, scenarioResult.ScenarioName),
				})
			}

			if regression.MemoryChange < -5.0 {
				improvements = append(improvements, PerformanceImprovement{
					Metric:             "memory",
					ImprovementPercent: -regression.MemoryChange,
					Description:        fmt.Sprintf("Memory usage improved by %.1f%% in %s", -regression.MemoryChange, scenarioResult.ScenarioName),
				})
			}
		}
	}

	result.RegressionIssues = regressions
	result.Improvements = improvements
	result.RegressionSeverity = maxSeverity
}

// Helper methods for BaselineManager
func (bm *BaselineManager) LoadBaselines() error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	if _, err := os.Stat(bm.baselineDir); os.IsNotExist(err) {
		if err := os.MkdirAll(bm.baselineDir, 0755); err != nil {
			return fmt.Errorf("failed to create baseline directory: %w", err)
		}
		return nil
	}

	files, err := filepath.Glob(filepath.Join(bm.baselineDir, "baseline_*.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		var baseline BenchmarkResult
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(data, &baseline); err != nil {
			continue
		}

		bm.currentBaselines[baseline.Scenario] = baseline
	}

	bm.logger.Info("Loaded baselines", map[string]interface{}{
		"count": len(bm.currentBaselines),
	})

	return nil
}

func (bm *BaselineManager) GetCurrentBaselines() map[string]BenchmarkResult {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	baselines := make(map[string]BenchmarkResult)
	for k, v := range bm.currentBaselines {
		baselines[k] = v
	}
	return baselines
}

func (bm *BaselineManager) UpdateBaseline(scenario string, result BenchmarkResult) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	bm.currentBaselines[scenario] = result

	// Save to disk
	filename := filepath.Join(bm.baselineDir, fmt.Sprintf("baseline_%s.json", scenario))
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0600)
}

func (bm *BaselineManager) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Daily cleanup
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bm.cleanupOldBaselines()
		}
	}
}

func (bm *BaselineManager) cleanupOldBaselines() {
	cutoff := time.Now().Add(-bm.retentionPeriod)

	files, err := filepath.Glob(filepath.Join(bm.baselineDir, "baseline_*.json"))
	if err != nil {
		return
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(file); err != nil {
				bm.logger.WithFields(map[string]interface{}{
					"file":  file,
					"error": err.Error(),
				}).Error("Failed to remove old baseline file", err)
			} else {
				bm.logger.Debug("Removed old baseline", map[string]interface{}{
					"file": file,
				})
			}
		}
	}
}

// Helper methods for test scheduling and execution
func (ts *TestScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(ts.config.ScheduleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ts.stopChan:
			return
		case <-ticker.C:
			ts.mutex.Lock()
			if !ts.isRunning {
				ts.isRunning = true
				ts.mutex.Unlock()

				go func() {
					_, err := ts.testSuite.RunRegressionTest(ctx, "scheduled")
					if err != nil {
						ts.logger.Error("Scheduled regression test failed", err, map[string]interface{}{})
					}

					ts.mutex.Lock()
					ts.isRunning = false
					ts.lastRun = time.Now()
					ts.mutex.Unlock()
				}()
			} else {
				ts.mutex.Unlock()
			}
		}
	}
}

func (ts *TestScheduler) Stop() {
	close(ts.stopChan)
}

// Additional helper methods...
func (rts *RegressionTestSuite) getGitInfo() (commit, branch string) {
	// Implementation would get git commit and branch information
	return "", ""
}

func (rts *RegressionTestSuite) createScenariosFromBaselines(baselines map[string]BenchmarkResult) []ScenarioConfig {
	// Create scenarios based on baseline configurations
	var scenarios []ScenarioConfig

	for name := range baselines {
		// Map baseline names to scenario types
		switch name {
		case "High Volume Container Replication":
			scenarios = append(scenarios, CreateHighVolumeReplicationScenario())
		case "Large Image Stress Test":
			scenarios = append(scenarios, CreateLargeImageStressScenario())
		case "Network Resilience Test":
			scenarios = append(scenarios, CreateNetworkResilienceScenario())
		case "Burst Replication Load":
			scenarios = append(scenarios, CreateBurstReplicationScenario())
		case "Sustained High-Throughput":
			scenarios = append(scenarios, CreateSustainedThroughputScenario())
		case "Mixed Container Sizes":
			scenarios = append(scenarios, CreateMixedContainerSizesScenario())
		}
	}

	return scenarios
}

func (rts *RegressionTestSuite) calculateRegressionSeverity(regression RegressionMetrics) string {
	// Use the same logic as the Prometheus collector
	if regression.ThroughputChange < -30 || regression.LatencyChange > 100 || regression.MemoryChange > 50 {
		return "critical"
	} else if regression.ThroughputChange < -20 || regression.LatencyChange > 50 || regression.MemoryChange > 30 {
		return "major"
	} else if regression.ThroughputChange < -10 || regression.LatencyChange > 25 || regression.MemoryChange > 15 {
		return "minor"
	} else if regression.ThroughputChange > 10 && regression.LatencyChange < -10 && regression.MemoryChange < -10 {
		return "improvement"
	}
	return "stable"
}

func (rts *RegressionTestSuite) isRegression(regression RegressionMetrics) bool {
	return regression.ThroughputChange < -rts.config.ThroughputRegressionThreshold ||
		regression.LatencyChange > rts.config.LatencyRegressionThreshold ||
		regression.MemoryChange > rts.config.MemoryRegressionThreshold ||
		regression.FailureRateChange > rts.config.FailureRateRegressionThreshold*100
}

func (rts *RegressionTestSuite) isImprovement(regression RegressionMetrics) bool {
	return regression.ThroughputChange > 5.0 || regression.LatencyChange < -5.0 || regression.MemoryChange < -5.0
}

func (rts *RegressionTestSuite) compareSeverity(a, b string) int {
	severityOrder := map[string]int{
		"none": 0, "stable": 1, "improvement": 1, "minor": 2, "major": 3, "critical": 4,
	}
	return severityOrder[a] - severityOrder[b]
}

func (rts *RegressionTestSuite) determineValidationStatus(regression RegressionMetrics) string {
	if rts.isRegression(regression) {
		if regression.Severity == "critical" || regression.Severity == "major" {
			return "fail"
		}
		return "warning"
	}
	return "pass"
}

func (rts *RegressionTestSuite) analyzePerformanceTrend(scenarioName string, result *LoadTestResults) string {
	// Analyze recent performance trends for this scenario
	// This would typically look at historical data
	return "stable" // Placeholder
}

func (rts *RegressionTestSuite) determineExecutionStatus(result *RegressionTestResult) string {
	failCount := 0
	for _, scenarioResult := range result.ScenarioResults {
		if scenarioResult.ValidationStatus == "fail" {
			failCount++
		}
	}

	if failCount == 0 {
		return "success"
	} else if failCount < len(result.ScenarioResults) {
		return "partial"
	}
	return "failed"
}

func (rts *RegressionTestSuite) generateRecommendations(result *RegressionTestResult) []string {
	var recommendations []string

	if result.RegressionSeverity == "critical" {
		recommendations = append(recommendations, "URGENT: Critical performance regression detected - investigate immediately")
	}

	for _, issue := range result.RegressionIssues {
		switch issue.Metric {
		case "throughput":
			recommendations = append(recommendations, fmt.Sprintf("Investigate throughput degradation in %s (%.1f%% drop)", issue.Benchmark, issue.RegressionPercent))
		case "latency":
			recommendations = append(recommendations, fmt.Sprintf("Investigate latency increase in %s (%.1f%% increase)", issue.Benchmark, issue.RegressionPercent))
		case "memory":
			recommendations = append(recommendations, fmt.Sprintf("Investigate memory usage increase in %s (%.1f%% increase)", issue.Benchmark, issue.RegressionPercent))
		}
	}

	if len(result.Improvements) > 0 {
		recommendations = append(recommendations, "Consider updating baselines to reflect performance improvements")
	}

	return recommendations
}

func (rts *RegressionTestSuite) saveRegressionResults(result *RegressionTestResult) error {
	resultsDir := filepath.Join(rts.baselineManager.baselineDir, "regression_results")
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return fmt.Errorf("failed to create regression results directory: %w", err)
	}

	filename := filepath.Join(resultsDir, fmt.Sprintf("%s.json", result.TestID))
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0600)
}

func (rts *RegressionTestSuite) updateBaselinesIfImproved(result *RegressionTestResult) {
	for _, improvement := range result.Improvements {
		if improvement.ImprovementPercent > rts.config.BaselineUpdateThreshold {
			// Update baseline if improvement is significant
			scenarioResult := result.ScenarioResults[improvement.Metric]
			if scenarioResult.CurrentResult != nil {
				baselineResult := BenchmarkResult{
					Tool:             "regression-test",
					Scenario:         scenarioResult.ScenarioName,
					Timestamp:        time.Now(),
					Duration:         scenarioResult.CurrentResult.Duration,
					ThroughputMBps:   scenarioResult.CurrentResult.AverageThroughputMBps,
					MemoryUsageMB:    scenarioResult.CurrentResult.MemoryUsageMB,
					FailureRate:      scenarioResult.CurrentResult.FailureRate,
					P99LatencyMs:     scenarioResult.CurrentResult.P99LatencyMs,
					ValidationPassed: scenarioResult.CurrentResult.ValidationPassed,
				}

				if err := rts.baselineManager.UpdateBaseline(scenarioResult.ScenarioName, baselineResult); err != nil {
					rts.logger.WithFields(map[string]interface{}{
						"scenario": scenarioResult.ScenarioName,
						"error":    err.Error(),
					}).Error("Failed to update baseline", err)
				}
				rts.logger.Info("Updated baseline due to performance improvement", map[string]interface{}{
					"scenario":            scenarioResult.ScenarioName,
					"improvement_percent": improvement.ImprovementPercent,
				})
			}
		}
	}
}

// AlertManager methods
func (am *AlertManager) SendRegressionAlerts(result RegressionTestResult) {
	for _, channel := range am.config.NotificationChannels {
		switch channel {
		case "email":
			if am.emailChannel != nil {
				if err := am.emailChannel.SendRegressionAlert(result); err != nil {
					am.logger.WithFields(map[string]interface{}{
						"error":   err.Error(),
						"channel": "email",
					}).Error("Failed to send regression alert", err)
				}
			}
		case "slack":
			if am.slackChannel != nil {
				if err := am.slackChannel.PostRegressionAlert(result); err != nil {
					am.logger.WithFields(map[string]interface{}{
						"error":   err.Error(),
						"channel": "slack",
					}).Error("Failed to send regression alert", err)
				}
			}
		case "webhook":
			if am.webhookChannel != nil {
				if err := am.webhookChannel.SendWebhookAlert(result); err != nil {
					am.logger.WithFields(map[string]interface{}{
						"error":   err.Error(),
						"channel": "webhook",
					}).Error("Failed to send regression alert", err)
				}
			}
		}
	}
}
