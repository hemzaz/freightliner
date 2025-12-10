package load

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/metrics"
)

// PrometheusLoadTestCollector integrates load testing with Prometheus metrics
type PrometheusLoadTestCollector struct {
	metricsCollector metrics.MetricsCollector
	logger           log.Logger

	// Load test specific metrics
	loadTestMetrics *LoadTestPrometheusMetrics

	// Configuration
	metricsAddr     string
	scrapeInterval  time.Duration
	retentionPeriod time.Duration

	// HTTP server for metrics endpoint
	metricsServer *http.Server
	serverMutex   sync.Mutex

	// Alert thresholds
	alertThresholds AlertThresholds
}

// LoadTestPrometheusMetrics contains Prometheus metrics for load testing
type LoadTestPrometheusMetrics struct {
	// Scenario execution metrics
	ScenarioExecutions  map[string]int64         // Counter per scenario
	ScenarioSuccessRate map[string]float64       // Success rate per scenario
	ScenarioDuration    map[string]time.Duration // Average duration per scenario

	// Performance metrics
	ThroughputMBps      map[string]float64 // Throughput per scenario
	MemoryUsageMB       map[string]int64   // Memory usage per scenario
	ConnectionReuseRate map[string]float64 // Connection reuse rate per scenario

	// Network and reliability metrics
	PacketLossRate     map[string]float64        // Simulated packet loss per scenario
	RetrySuccessRate   map[string]float64        // Retry success rate per scenario
	LatencyPercentiles map[string]LatencyMetrics // Latency percentiles per scenario

	// Resource utilization
	CPUUtilization     map[string]float64 // CPU usage during tests
	NetworkUtilization map[string]float64 // Network bandwidth utilization
	DiskIOPS           map[string]float64 // Disk IOPS during tests

	// Regression tracking
	BaselineComparison map[string]RegressionMetrics // Comparison with baseline
	PerformanceTrends  []PerformanceTrendPoint      // Historical performance data

	Mutex sync.RWMutex // Exported for testing access
}

// LatencyMetrics contains latency percentile measurements
type LatencyMetrics struct {
	P50Ms  int64 `json:"p50_ms"`
	P95Ms  int64 `json:"p95_ms"`
	P99Ms  int64 `json:"p99_ms"`
	P999Ms int64 `json:"p999_ms"`
	MaxMs  int64 `json:"max_ms"`
}

// RegressionMetrics tracks performance changes compared to baseline
type RegressionMetrics struct {
	ThroughputChange  float64 `json:"throughput_change_percent"`
	LatencyChange     float64 `json:"latency_change_percent"`
	MemoryChange      float64 `json:"memory_change_percent"`
	FailureRateChange float64 `json:"failure_rate_change_percent"`
	Severity          string  `json:"severity"` // "improvement", "minor", "major", "critical"
}

// PerformanceTrendPoint represents a point in performance trend analysis
type PerformanceTrendPoint struct {
	Timestamp        time.Time `json:"timestamp"`
	ScenarioName     string    `json:"scenario_name"`
	ThroughputMBps   float64   `json:"throughput_mbps"`
	LatencyP99Ms     int64     `json:"latency_p99_ms"`
	MemoryUsageMB    int64     `json:"memory_usage_mb"`
	FailureRate      float64   `json:"failure_rate"`
	ValidationPassed bool      `json:"validation_passed"`
}

// AlertThresholds defines when to trigger performance alerts
type AlertThresholds struct {
	ThroughputDropPercent      float64 `json:"throughput_drop_percent"`       // Alert if throughput drops by this %
	LatencyIncreasePercent     float64 `json:"latency_increase_percent"`      // Alert if latency increases by this %
	MemoryIncreasePercent      float64 `json:"memory_increase_percent"`       // Alert if memory usage increases by this %
	FailureRateThreshold       float64 `json:"failure_rate_threshold"`        // Alert if failure rate exceeds this
	ConnectionReuseDropPercent float64 `json:"connection_reuse_drop_percent"` // Alert if connection reuse drops
}

// NewPrometheusLoadTestCollector creates a new Prometheus-integrated load test collector
func NewPrometheusLoadTestCollector(metricsAddr string, logger log.Logger) *PrometheusLoadTestCollector {
	if logger == nil {
		logger = log.NewLogger()
	}

	return &PrometheusLoadTestCollector{
		metricsCollector: metrics.NewPrometheusMetrics(),
		logger:           logger,
		metricsAddr:      metricsAddr,
		scrapeInterval:   15 * time.Second,
		retentionPeriod:  7 * 24 * time.Hour, // 7 days
		loadTestMetrics: &LoadTestPrometheusMetrics{
			ScenarioExecutions:  make(map[string]int64),
			ScenarioSuccessRate: make(map[string]float64),
			ScenarioDuration:    make(map[string]time.Duration),
			ThroughputMBps:      make(map[string]float64),
			MemoryUsageMB:       make(map[string]int64),
			ConnectionReuseRate: make(map[string]float64),
			PacketLossRate:      make(map[string]float64),
			RetrySuccessRate:    make(map[string]float64),
			LatencyPercentiles:  make(map[string]LatencyMetrics),
			CPUUtilization:      make(map[string]float64),
			NetworkUtilization:  make(map[string]float64),
			DiskIOPS:            make(map[string]float64),
			BaselineComparison:  make(map[string]RegressionMetrics),
		},
		alertThresholds: AlertThresholds{
			ThroughputDropPercent:      20.0, // 20% throughput drop
			LatencyIncreasePercent:     50.0, // 50% latency increase
			MemoryIncreasePercent:      30.0, // 30% memory increase
			FailureRateThreshold:       0.05, // 5% failure rate
			ConnectionReuseDropPercent: 15.0, // 15% connection reuse drop
		},
	}
}

// GetLoadTestMetrics returns the load test metrics (exported for testing)
func (pc *PrometheusLoadTestCollector) GetLoadTestMetrics() *LoadTestPrometheusMetrics {
	return pc.loadTestMetrics
}

// StartMetricsServer starts the Prometheus metrics HTTP server
func (pc *PrometheusLoadTestCollector) StartMetricsServer(ctx context.Context) error {
	pc.serverMutex.Lock()
	defer pc.serverMutex.Unlock()

	if pc.metricsServer != nil {
		return fmt.Errorf("metrics server already running")
	}

	mux := http.NewServeMux()

	// Main Prometheus metrics endpoint
	mux.HandleFunc("/metrics", pc.handleMetrics)

	// Load test specific endpoints
	mux.HandleFunc("/metrics/scenarios", pc.handleScenarioMetrics)
	mux.HandleFunc("/metrics/performance", pc.handlePerformanceMetrics)
	mux.HandleFunc("/metrics/regression", pc.handleRegressionMetrics)
	mux.HandleFunc("/metrics/health", pc.handleHealthCheck)

	// Performance dashboard data
	mux.HandleFunc("/dashboard/data", pc.handleDashboardData)

	pc.metricsServer = &http.Server{
		Addr:    pc.metricsAddr,
		Handler: mux,
	}

	go func() {
		pc.logger.WithFields(map[string]interface{}{
			"address": pc.metricsAddr,
		}).Info("Starting Prometheus metrics server")

		// Safely access the server
		pc.serverMutex.Lock()
		server := pc.metricsServer
		pc.serverMutex.Unlock()

		if server != nil {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Metrics server error", err)
			}
		}
	}()

	// Start background metrics collection
	go pc.collectSystemMetrics(ctx)
	go pc.detectPerformanceAlerts(ctx)

	return nil
}

// StopMetricsServer stops the Prometheus metrics server
func (pc *PrometheusLoadTestCollector) StopMetricsServer() error {
	pc.serverMutex.Lock()
	defer pc.serverMutex.Unlock()

	if pc.metricsServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pc.metricsServer.Shutdown(ctx)
	pc.metricsServer = nil

	pc.logger.Info("Prometheus metrics server stopped")
	return err
}

// RecordScenarioExecution records metrics for a completed scenario
func (pc *PrometheusLoadTestCollector) RecordScenarioExecution(scenario string, result *LoadTestResults) {
	pc.loadTestMetrics.Mutex.Lock()
	defer pc.loadTestMetrics.Mutex.Unlock()

	// Update execution counters
	pc.loadTestMetrics.ScenarioExecutions[scenario]++

	// Calculate success rate
	if result.ProcessedImages > 0 {
		successRate := float64(result.ProcessedImages-result.FailedImages) / float64(result.ProcessedImages)
		pc.loadTestMetrics.ScenarioSuccessRate[scenario] = successRate
	}

	// Record performance metrics
	pc.loadTestMetrics.ScenarioDuration[scenario] = result.Duration
	pc.loadTestMetrics.ThroughputMBps[scenario] = result.AverageThroughputMBps
	pc.loadTestMetrics.MemoryUsageMB[scenario] = result.MemoryUsageMB
	pc.loadTestMetrics.ConnectionReuseRate[scenario] = result.ConnectionReuseRate

	// Record latency metrics if available
	if result.P99LatencyMs > 0 {
		pc.loadTestMetrics.LatencyPercentiles[scenario] = LatencyMetrics{
			P99Ms: result.P99LatencyMs,
			// Additional percentiles would come from detailed results
		}
	}

	// Add to performance trends
	trendPoint := PerformanceTrendPoint{
		Timestamp:        time.Now(),
		ScenarioName:     scenario,
		ThroughputMBps:   result.AverageThroughputMBps,
		LatencyP99Ms:     result.P99LatencyMs,
		MemoryUsageMB:    result.MemoryUsageMB,
		FailureRate:      result.FailureRate,
		ValidationPassed: result.ValidationPassed,
	}
	pc.loadTestMetrics.PerformanceTrends = append(pc.loadTestMetrics.PerformanceTrends, trendPoint)

	// Trim old trend data to maintain retention period
	cutoff := time.Now().Add(-pc.retentionPeriod)
	pc.trimOldTrends(cutoff)

	successRate := 1.0 - result.FailureRate
	pc.logger.WithFields(map[string]interface{}{
		"scenario":          scenario,
		"throughput_mbps":   result.AverageThroughputMBps,
		"memory_mb":         result.MemoryUsageMB,
		"success_rate":      fmt.Sprintf("%.2f%%", successRate*100),
		"validation_passed": result.ValidationPassed,
	}).Info("Recorded scenario execution")
}

// CompareWithBaseline compares current results with baseline and detects regressions
func (pc *PrometheusLoadTestCollector) CompareWithBaseline(scenario string, result *LoadTestResults, baseline BenchmarkResult) {
	pc.loadTestMetrics.Mutex.Lock()
	defer pc.loadTestMetrics.Mutex.Unlock()

	regression := RegressionMetrics{}

	// Calculate throughput change
	if baseline.ThroughputMBps > 0 {
		regression.ThroughputChange = ((result.AverageThroughputMBps - baseline.ThroughputMBps) / baseline.ThroughputMBps) * 100
	}

	// Calculate latency change
	if baseline.P99LatencyMs > 0 && result.P99LatencyMs > 0 {
		regression.LatencyChange = ((float64(result.P99LatencyMs) - float64(baseline.P99LatencyMs)) / float64(baseline.P99LatencyMs)) * 100
	}

	// Calculate memory change
	if baseline.MemoryUsageMB > 0 {
		regression.MemoryChange = ((float64(result.MemoryUsageMB) - float64(baseline.MemoryUsageMB)) / float64(baseline.MemoryUsageMB)) * 100
	}

	// Calculate failure rate change
	regression.FailureRateChange = (result.FailureRate - baseline.FailureRate) * 100

	// Determine severity
	regression.Severity = pc.calculateRegressionSeverity(regression)

	pc.loadTestMetrics.BaselineComparison[scenario] = regression

	pc.logger.WithFields(map[string]interface{}{
		"scenario":          scenario,
		"throughput_change": fmt.Sprintf("%.1f%%", regression.ThroughputChange),
		"latency_change":    fmt.Sprintf("%.1f%%", regression.LatencyChange),
		"memory_change":     fmt.Sprintf("%.1f%%", regression.MemoryChange),
		"severity":          regression.Severity,
	}).Info("Baseline comparison completed")
}

// HTTP handlers for Prometheus metrics endpoints

func (pc *PrometheusLoadTestCollector) handleMetrics(w http.ResponseWriter, r *http.Request) {
	pc.loadTestMetrics.Mutex.RLock()
	defer pc.loadTestMetrics.Mutex.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	// Generate Prometheus formatted metrics
	if _, err := fmt.Fprintf(w, "# HELP load_test_scenario_executions_total Total number of scenario executions\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write metrics help", err)
		return
	}
	if _, err := fmt.Fprintf(w, "# TYPE load_test_scenario_executions_total counter\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write metrics type", err)
		return
	}
	for scenario, count := range pc.loadTestMetrics.ScenarioExecutions {
		if _, err := fmt.Fprintf(w, "load_test_scenario_executions_total{scenario=\"%s\"} %d\n", scenario, count); err != nil {
			pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write scenario execution metrics", err)
			return
		}
	}

	if _, err := fmt.Fprintf(w, "\n# HELP load_test_throughput_mbps Current throughput in MB/s\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write throughput help", err)
		return
	}
	if _, err := fmt.Fprintf(w, "# TYPE load_test_throughput_mbps gauge\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write throughput type", err)
		return
	}
	for scenario, throughput := range pc.loadTestMetrics.ThroughputMBps {
		if _, err := fmt.Fprintf(w, "load_test_throughput_mbps{scenario=\"%s\"} %.2f\n", scenario, throughput); err != nil {
			pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write throughput metrics", err)
			return
		}
	}

	if _, err := fmt.Fprintf(w, "\n# HELP load_test_memory_usage_mb Current memory usage in MB\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write memory help", err)
		return
	}
	if _, err := fmt.Fprintf(w, "# TYPE load_test_memory_usage_mb gauge\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write memory type", err)
		return
	}
	for scenario, memory := range pc.loadTestMetrics.MemoryUsageMB {
		if _, err := fmt.Fprintf(w, "load_test_memory_usage_mb{scenario=\"%s\"} %d\n", scenario, memory); err != nil {
			pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write memory metrics", err)
			return
		}
	}

	if _, err := fmt.Fprintf(w, "\n# HELP load_test_success_rate Success rate for scenarios\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write success rate help", err)
		return
	}
	if _, err := fmt.Fprintf(w, "# TYPE load_test_success_rate gauge\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write success rate type", err)
		return
	}
	for scenario, rate := range pc.loadTestMetrics.ScenarioSuccessRate {
		if _, err := fmt.Fprintf(w, "load_test_success_rate{scenario=\"%s\"} %.4f\n", scenario, rate); err != nil {
			pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write success rate metrics", err)
			return
		}
	}

	if _, err := fmt.Fprintf(w, "\n# HELP load_test_connection_reuse_rate Connection reuse rate\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write connection reuse help", err)
		return
	}
	if _, err := fmt.Fprintf(w, "# TYPE load_test_connection_reuse_rate gauge\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write connection reuse type", err)
		return
	}
	for scenario, rate := range pc.loadTestMetrics.ConnectionReuseRate {
		if _, err := fmt.Fprintf(w, "load_test_connection_reuse_rate{scenario=\"%s\"} %.4f\n", scenario, rate); err != nil {
			pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write connection reuse metrics", err)
			return
		}
	}

	if _, err := fmt.Fprintf(w, "\n# HELP load_test_latency_p99_ms 99th percentile latency in milliseconds\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write latency help", err)
		return
	}
	if _, err := fmt.Fprintf(w, "# TYPE load_test_latency_p99_ms gauge\n"); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write latency type", err)
		return
	}
	for scenario, latency := range pc.loadTestMetrics.LatencyPercentiles {
		if _, err := fmt.Fprintf(w, "load_test_latency_p99_ms{scenario=\"%s\"} %d\n", scenario, latency.P99Ms); err != nil {
			pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to write latency metrics", err)
			return
		}
	}
}

func (pc *PrometheusLoadTestCollector) handleScenarioMetrics(w http.ResponseWriter, r *http.Request) {
	pc.loadTestMetrics.Mutex.RLock()
	defer pc.loadTestMetrics.Mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	scenarioData := map[string]interface{}{
		"executions":       pc.loadTestMetrics.ScenarioExecutions,
		"success_rates":    pc.loadTestMetrics.ScenarioSuccessRate,
		"throughput_mbps":  pc.loadTestMetrics.ThroughputMBps,
		"memory_usage_mb":  pc.loadTestMetrics.MemoryUsageMB,
		"connection_reuse": pc.loadTestMetrics.ConnectionReuseRate,
		"latency_p99_ms":   pc.loadTestMetrics.LatencyPercentiles,
	}

	if err := json.NewEncoder(w).Encode(scenarioData); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to encode scenario data", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (pc *PrometheusLoadTestCollector) handlePerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	pc.loadTestMetrics.Mutex.RLock()
	defer pc.loadTestMetrics.Mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	// Return recent performance trends (last 24 hours)
	cutoff := time.Now().Add(-24 * time.Hour)
	recentTrends := []PerformanceTrendPoint{}

	for _, trend := range pc.loadTestMetrics.PerformanceTrends {
		if trend.Timestamp.After(cutoff) {
			recentTrends = append(recentTrends, trend)
		}
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"trends":          recentTrends,
		"cpu_utilization": pc.loadTestMetrics.CPUUtilization,
		"network_usage":   pc.loadTestMetrics.NetworkUtilization,
		"disk_iops":       pc.loadTestMetrics.DiskIOPS,
	}); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to encode performance metrics", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (pc *PrometheusLoadTestCollector) handleRegressionMetrics(w http.ResponseWriter, r *http.Request) {
	pc.loadTestMetrics.Mutex.RLock()
	defer pc.loadTestMetrics.Mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pc.loadTestMetrics.BaselineComparison); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to encode baseline comparison", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (pc *PrometheusLoadTestCollector) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to encode health check response", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (pc *PrometheusLoadTestCollector) handleDashboardData(w http.ResponseWriter, r *http.Request) {
	pc.loadTestMetrics.Mutex.RLock()
	defer pc.loadTestMetrics.Mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	dashboardData := map[string]interface{}{
		"scenarios":        pc.loadTestMetrics.ScenarioExecutions,
		"performance":      pc.loadTestMetrics.ThroughputMBps,
		"memory":           pc.loadTestMetrics.MemoryUsageMB,
		"success_rates":    pc.loadTestMetrics.ScenarioSuccessRate,
		"regressions":      pc.loadTestMetrics.BaselineComparison,
		"trends":           pc.loadTestMetrics.PerformanceTrends,
		"alert_thresholds": pc.alertThresholds,
	}

	if err := json.NewEncoder(w).Encode(dashboardData); err != nil {
		pc.logger.WithFields(map[string]interface{}{"error": err.Error()}).Error("Failed to encode dashboard data", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Background monitoring functions

func (pc *PrometheusLoadTestCollector) collectSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(pc.scrapeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pc.updateSystemMetrics()
		}
	}
}

func (pc *PrometheusLoadTestCollector) detectPerformanceAlerts(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // Check for alerts every minute
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pc.checkPerformanceAlerts()
		}
	}
}

func (pc *PrometheusLoadTestCollector) updateSystemMetrics() {
	// Implementation would collect actual system metrics
	// For now, this is a placeholder
	pc.loadTestMetrics.Mutex.Lock()
	defer pc.loadTestMetrics.Mutex.Unlock()

	// Update CPU, memory, network, and disk metrics
	// This would typically use system monitoring libraries
}

func (pc *PrometheusLoadTestCollector) checkPerformanceAlerts() {
	pc.loadTestMetrics.Mutex.RLock()
	defer pc.loadTestMetrics.Mutex.RUnlock()

	for scenario, regression := range pc.loadTestMetrics.BaselineComparison {
		if pc.shouldTriggerAlert(regression) {
			pc.triggerPerformanceAlert(scenario, regression)
		}
	}
}

func (pc *PrometheusLoadTestCollector) shouldTriggerAlert(regression RegressionMetrics) bool {
	return regression.ThroughputChange < -pc.alertThresholds.ThroughputDropPercent ||
		regression.LatencyChange > pc.alertThresholds.LatencyIncreasePercent ||
		regression.MemoryChange > pc.alertThresholds.MemoryIncreasePercent ||
		regression.FailureRateChange > pc.alertThresholds.FailureRateThreshold*100
}

func (pc *PrometheusLoadTestCollector) triggerPerformanceAlert(scenario string, regression RegressionMetrics) {
	pc.logger.WithFields(map[string]interface{}{
		"scenario":          scenario,
		"throughput_change": fmt.Sprintf("%.1f%%", regression.ThroughputChange),
		"latency_change":    fmt.Sprintf("%.1f%%", regression.LatencyChange),
		"memory_change":     fmt.Sprintf("%.1f%%", regression.MemoryChange),
		"severity":          regression.Severity,
	}).Warn("Performance alert triggered")

	// Here you would integrate with alerting systems like:
	// - Slack notifications
	// - Email alerts
	// - PagerDuty
	// - Webhook notifications
}

func (pc *PrometheusLoadTestCollector) calculateRegressionSeverity(regression RegressionMetrics) string {
	// Determine severity based on multiple factors
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

func (pc *PrometheusLoadTestCollector) trimOldTrends(cutoff time.Time) {
	filtered := []PerformanceTrendPoint{}
	for _, trend := range pc.loadTestMetrics.PerformanceTrends {
		if trend.Timestamp.After(cutoff) {
			filtered = append(filtered, trend)
		}
	}
	pc.loadTestMetrics.PerformanceTrends = filtered
}
