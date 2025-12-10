package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
)

// TestPerformanceMonitor tracks and analyzes test execution performance
type TestPerformanceMonitor struct {
	logger             log.Logger
	metricsDir         string
	performanceHistory map[string][]*PerformanceSnapshot
	currentSession     *TestSession
	thresholds         *PerformanceThresholds
	alerting           *AlertingConfig
	mu                 sync.RWMutex
}

// PerformanceSnapshot captures performance metrics at a point in time
type PerformanceSnapshot struct {
	Timestamp       time.Time              `json:"timestamp"`
	PackagePath     string                 `json:"package_path"`
	TestName        string                 `json:"test_name"`
	Duration        time.Duration          `json:"duration"`
	MemoryUsageMB   int64                  `json:"memory_usage_mb"`
	PeakMemoryMB    int64                  `json:"peak_memory_mb"`
	CPUUsagePercent float64                `json:"cpu_usage_percent"`
	GoroutineCount  int                    `json:"goroutine_count"`
	GCPauseTime     time.Duration          `json:"gc_pause_time"`
	AllocObjects    uint64                 `json:"alloc_objects"`
	HeapSize        uint64                 `json:"heap_size"`
	Success         bool                   `json:"success"`
	RetryCount      int                    `json:"retry_count"`
	CacheHit        bool                   `json:"cache_hit"`
	ExternalCalls   int                    `json:"external_calls"`
	NetworkLatency  time.Duration          `json:"network_latency"`
	DiskIO          int64                  `json:"disk_io_bytes"`
	CustomMetrics   map[string]interface{} `json:"custom_metrics"`
}

// TestSession represents a complete test execution session
type TestSession struct {
	SessionID       string                `json:"session_id"`
	StartTime       time.Time             `json:"start_time"`
	EndTime         time.Time             `json:"end_time"`
	TotalDuration   time.Duration         `json:"total_duration"`
	TestType        string                `json:"test_type"`
	Environment     string                `json:"environment"`
	GoVersion       string                `json:"go_version"`
	TotalPackages   int                   `json:"total_packages"`
	PackageResults  []*PackageResult      `json:"package_results"`
	SystemMetrics   *SystemPerformance    `json:"system_metrics"`
	Anomalies       []*PerformanceAnomaly `json:"anomalies"`
	Recommendations []string              `json:"recommendations"`
}

// PackageResult holds performance results for a single package
type PackageResult struct {
	PackagePath      string                 `json:"package_path"`
	Duration         time.Duration          `json:"duration"`
	TestCount        int                    `json:"test_count"`
	Success          bool                   `json:"success"`
	MemoryProfile    *MemoryProfile         `json:"memory_profile"`
	PerformanceScore float64                `json:"performance_score"`
	Snapshots        []*PerformanceSnapshot `json:"snapshots"`
}

// MemoryProfile provides detailed memory usage analysis
type MemoryProfile struct {
	PeakUsageMB    int64         `json:"peak_usage_mb"`
	AverageUsageMB int64         `json:"average_usage_mb"`
	AllocationRate float64       `json:"allocation_rate"`
	GCFrequency    float64       `json:"gc_frequency"`
	GCPauseTotal   time.Duration `json:"gc_pause_total"`
	HeapGrowthRate float64       `json:"heap_growth_rate"`
	StackUsageMB   int64         `json:"stack_usage_mb"`
}

// SystemPerformance captures overall system performance during tests
type SystemPerformance struct {
	AverageCPUUsage  float64 `json:"average_cpu_usage"`
	PeakCPUUsage     float64 `json:"peak_cpu_usage"`
	AverageMemoryMB  int64   `json:"average_memory_mb"`
	PeakMemoryMB     int64   `json:"peak_memory_mb"`
	DiskUsageMB      int64   `json:"disk_usage_mb"`
	NetworkTrafficMB int64   `json:"network_traffic_mb"`
	LoadAverage      float64 `json:"load_average"`
	OpenFileHandles  int     `json:"open_file_handles"`
	Concurrency      int     `json:"concurrency"`
}

// PerformanceThresholds defines acceptable performance limits
type PerformanceThresholds struct {
	MaxDurationPerTest   time.Duration `json:"max_duration_per_test"`
	MaxMemoryUsageMB     int64         `json:"max_memory_usage_mb"`
	MaxCPUUsagePercent   float64       `json:"max_cpu_usage_percent"`
	MaxGCPauseTime       time.Duration `json:"max_gc_pause_time"`
	MaxRetryCount        int           `json:"max_retry_count"`
	MinSuccessRate       float64       `json:"min_success_rate"`
	MaxRegressionPercent float64       `json:"max_regression_percent"`
}

// PerformanceAnomaly represents detected performance issues
type PerformanceAnomaly struct {
	Type        string      `json:"type"`
	Severity    string      `json:"severity"`
	Package     string      `json:"package"`
	TestName    string      `json:"test_name"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
	Threshold   interface{} `json:"threshold"`
	Timestamp   time.Time   `json:"timestamp"`
	Suggestions []string    `json:"suggestions"`
}

// AlertingConfig defines when and how to alert on performance issues
type AlertingConfig struct {
	Enabled           bool          `json:"enabled"`
	SlackWebhook      string        `json:"slack_webhook"`
	EmailRecipients   []string      `json:"email_recipients"`
	CriticalThreshold float64       `json:"critical_threshold"`
	WarningThreshold  float64       `json:"warning_threshold"`
	CooldownPeriod    time.Duration `json:"cooldown_period"`
}

// NewTestPerformanceMonitor creates a new performance monitoring system
func NewTestPerformanceMonitor(logger log.Logger, metricsDir string) *TestPerformanceMonitor {
	if logger == nil {
		logger = log.NewLogger()
	}

	if metricsDir == "" {
		metricsDir = filepath.Join(".", ".test-metrics")
	}

	_ = os.MkdirAll(metricsDir, 0755)

	monitor := &TestPerformanceMonitor{
		logger:             logger,
		metricsDir:         metricsDir,
		performanceHistory: make(map[string][]*PerformanceSnapshot),
		thresholds: &PerformanceThresholds{
			MaxDurationPerTest:   10 * time.Minute,
			MaxMemoryUsageMB:     2048,
			MaxCPUUsagePercent:   80.0,
			MaxGCPauseTime:       100 * time.Millisecond,
			MaxRetryCount:        3,
			MinSuccessRate:       95.0,
			MaxRegressionPercent: 20.0,
		},
		alerting: &AlertingConfig{
			Enabled:           false,
			CriticalThreshold: 50.0,
			WarningThreshold:  20.0,
			CooldownPeriod:    1 * time.Hour,
		},
	}

	// Load historical data
	if err := monitor.loadPerformanceHistory(); err != nil {
		monitor.logger.Error("Failed to load performance history", err)
	}

	return monitor
}

// StartSession begins a new test session monitoring
func (tpm *TestPerformanceMonitor) StartSession(testType, environment string) *TestSession {
	sessionID := fmt.Sprintf("%s_%s_%d", testType, environment, time.Now().Unix())

	session := &TestSession{
		SessionID:       sessionID,
		StartTime:       time.Now(),
		TestType:        testType,
		Environment:     environment,
		GoVersion:       runtime.Version(),
		PackageResults:  make([]*PackageResult, 0),
		SystemMetrics:   &SystemPerformance{},
		Anomalies:       make([]*PerformanceAnomaly, 0),
		Recommendations: make([]string, 0),
	}

	tpm.mu.Lock()
	tpm.currentSession = session
	tpm.mu.Unlock()

	tpm.logger.Info(fmt.Sprintf("Started performance monitoring session: %s", sessionID))

	// Start system metrics collection
	go tpm.collectSystemMetrics(context.Background())

	return session
}

// RecordTestExecution records performance metrics for a test execution
func (tpm *TestPerformanceMonitor) RecordTestExecution(packagePath, testName string, duration time.Duration, success bool) *PerformanceSnapshot {
	snapshot := &PerformanceSnapshot{
		Timestamp:     time.Now(),
		PackagePath:   packagePath,
		TestName:      testName,
		Duration:      duration,
		Success:       success,
		CustomMetrics: make(map[string]interface{}),
	}

	// Capture runtime metrics
	tpm.captureRuntimeMetrics(snapshot)

	// Store snapshot
	tpm.mu.Lock()
	if tpm.performanceHistory[packagePath] == nil {
		tpm.performanceHistory[packagePath] = make([]*PerformanceSnapshot, 0)
	}
	tpm.performanceHistory[packagePath] = append(tpm.performanceHistory[packagePath], snapshot)

	// Add to current session
	if tpm.currentSession != nil {
		tpm.addSnapshotToSession(snapshot)
	}
	tpm.mu.Unlock()

	// Check for anomalies
	tpm.detectAnomalies(snapshot)

	return snapshot
}

// EndSession completes the current test session and generates analysis
func (tpm *TestPerformanceMonitor) EndSession() (*TestSession, error) {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()

	if tpm.currentSession == nil {
		return nil, fmt.Errorf("no active session")
	}

	session := tpm.currentSession
	session.EndTime = time.Now()
	session.TotalDuration = session.EndTime.Sub(session.StartTime)

	// Generate performance analysis
	tpm.analyzeSessionPerformance(session)

	// Generate recommendations
	tpm.generateRecommendations(session)

	// Save session data
	if err := tpm.saveSessionData(session); err != nil {
		tpm.logger.Error("Failed to save session data", err)
	}

	tpm.logger.Info(fmt.Sprintf("Completed performance monitoring session: %s (duration: %v, packages: %d)",
		session.SessionID, session.TotalDuration, session.TotalPackages))

	tpm.currentSession = nil

	return session, nil
}

// GetPerformanceTrends analyzes performance trends over time
func (tpm *TestPerformanceMonitor) GetPerformanceTrends(packagePath string, days int) (*PerformanceTrends, error) {
	tpm.mu.RLock()
	snapshots, exists := tpm.performanceHistory[packagePath]
	tpm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no performance history for package: %s", packagePath)
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	recentSnapshots := make([]*PerformanceSnapshot, 0)

	for _, snapshot := range snapshots {
		if snapshot.Timestamp.After(cutoff) {
			recentSnapshots = append(recentSnapshots, snapshot)
		}
	}

	trends := &PerformanceTrends{
		PackagePath:   packagePath,
		TimeRange:     days,
		DataPoints:    len(recentSnapshots),
		TrendAnalysis: make(map[string]*TrendData),
	}

	// Analyze trends for key metrics
	trends.TrendAnalysis["duration"] = tpm.analyzeTrend(recentSnapshots, func(s *PerformanceSnapshot) float64 {
		return float64(s.Duration.Milliseconds())
	})

	trends.TrendAnalysis["memory"] = tpm.analyzeTrend(recentSnapshots, func(s *PerformanceSnapshot) float64 {
		return float64(s.MemoryUsageMB)
	})

	trends.TrendAnalysis["success_rate"] = tpm.analyzeTrend(recentSnapshots, func(s *PerformanceSnapshot) float64 {
		if s.Success {
			return 1.0
		}
		return 0.0
	})

	return trends, nil
}

// PerformanceTrends represents performance trend analysis
type PerformanceTrends struct {
	PackagePath   string                `json:"package_path"`
	TimeRange     int                   `json:"time_range_days"`
	DataPoints    int                   `json:"data_points"`
	TrendAnalysis map[string]*TrendData `json:"trend_analysis"`
}

// TrendData represents trend analysis for a specific metric
type TrendData struct {
	Current       float64 `json:"current"`
	Previous      float64 `json:"previous"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Direction     string  `json:"direction"` // "improving", "degrading", "stable"
	Confidence    float64 `json:"confidence"`
}

// GeneratePerformanceReport creates a comprehensive performance report
func (tpm *TestPerformanceMonitor) GeneratePerformanceReport() (*PerformanceReport, error) {
	tpm.mu.RLock()
	defer tpm.mu.RUnlock()

	report := &PerformanceReport{
		GeneratedAt:     time.Now(),
		TotalPackages:   len(tpm.performanceHistory),
		PackageReports:  make([]*PackagePerformanceReport, 0),
		GlobalMetrics:   &GlobalPerformanceMetrics{},
		Recommendations: make([]string, 0),
	}

	// Generate package-level reports
	for packagePath, snapshots := range tpm.performanceHistory {
		packageReport := tpm.generatePackageReport(packagePath, snapshots)
		report.PackageReports = append(report.PackageReports, packageReport)
	}

	// Sort packages by performance score
	sort.Slice(report.PackageReports, func(i, j int) bool {
		return report.PackageReports[i].PerformanceScore > report.PackageReports[j].PerformanceScore
	})

	// Generate global metrics
	tpm.calculateGlobalMetrics(report)

	// Generate system-wide recommendations
	tpm.generateGlobalRecommendations(report)

	return report, nil
}

// PerformanceReport represents a comprehensive performance analysis
type PerformanceReport struct {
	GeneratedAt     time.Time                   `json:"generated_at"`
	TotalPackages   int                         `json:"total_packages"`
	PackageReports  []*PackagePerformanceReport `json:"package_reports"`
	GlobalMetrics   *GlobalPerformanceMetrics   `json:"global_metrics"`
	Recommendations []string                    `json:"recommendations"`
}

// PackagePerformanceReport provides detailed analysis for a single package
type PackagePerformanceReport struct {
	PackagePath        string             `json:"package_path"`
	TotalExecutions    int                `json:"total_executions"`
	SuccessRate        float64            `json:"success_rate"`
	AverageDuration    time.Duration      `json:"average_duration"`
	MedianDuration     time.Duration      `json:"median_duration"`
	P95Duration        time.Duration      `json:"p95_duration"`
	AverageMemoryUsage int64              `json:"average_memory_usage"`
	PeakMemoryUsage    int64              `json:"peak_memory_usage"`
	PerformanceScore   float64            `json:"performance_score"`
	StabilityScore     float64            `json:"stability_score"`
	EfficiencyScore    float64            `json:"efficiency_score"`
	RecentTrends       *PerformanceTrends `json:"recent_trends"`
	IdentifiedIssues   []string           `json:"identified_issues"`
	Recommendations    []string           `json:"recommendations"`
}

// GlobalPerformanceMetrics provides system-wide performance insights
type GlobalPerformanceMetrics struct {
	OverallSuccessRate      float64                   `json:"overall_success_rate"`
	AverageTestDuration     time.Duration             `json:"average_test_duration"`
	TotalTestTime           time.Duration             `json:"total_test_time"`
	SystemEfficiencyScore   float64                   `json:"system_efficiency_score"`
	MostProblematicPackages []string                  `json:"most_problematic_packages"`
	BestPerformingPackages  []string                  `json:"best_performing_packages"`
	ResourceUtilization     *ResourceUtilizationStats `json:"resource_utilization"`
}

// ResourceUtilizationStats tracks resource usage patterns
type ResourceUtilizationStats struct {
	AverageCPUUsage    float64 `json:"average_cpu_usage"`
	PeakCPUUsage       float64 `json:"peak_cpu_usage"`
	AverageMemoryUsage int64   `json:"average_memory_usage"`
	PeakMemoryUsage    int64   `json:"peak_memory_usage"`
	TotalDiskIO        int64   `json:"total_disk_io"`
	TotalNetworkIO     int64   `json:"total_network_io"`
}

// Helper methods implementation

func (tpm *TestPerformanceMonitor) captureRuntimeMetrics(snapshot *PerformanceSnapshot) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snapshot.MemoryUsageMB = int64(m.Alloc / 1024 / 1024)
	snapshot.PeakMemoryMB = int64(m.Sys / 1024 / 1024)
	snapshot.GoroutineCount = runtime.NumGoroutine()
	snapshot.AllocObjects = m.Mallocs - m.Frees
	snapshot.HeapSize = m.HeapAlloc

	// Calculate GC pause time (simplified)
	if len(m.PauseNs) > 0 {
		snapshot.GCPauseTime = time.Duration(m.PauseNs[(m.NumGC+255)%256])
	}
}

func (tpm *TestPerformanceMonitor) collectSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Collect system-wide metrics
			tpm.updateSystemMetrics()
		}
	}
}

func (tpm *TestPerformanceMonitor) updateSystemMetrics() {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()

	if tpm.currentSession == nil {
		return
	}

	// Update system metrics (simplified implementation)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	tpm.currentSession.SystemMetrics.AverageMemoryMB = int64(m.Alloc / 1024 / 1024)
	tpm.currentSession.SystemMetrics.OpenFileHandles = runtime.NumGoroutine()
}

func (tpm *TestPerformanceMonitor) addSnapshotToSession(snapshot *PerformanceSnapshot) {
	// Find or create package result
	var packageResult *PackageResult
	for _, pr := range tpm.currentSession.PackageResults {
		if pr.PackagePath == snapshot.PackagePath {
			packageResult = pr
			break
		}
	}

	if packageResult == nil {
		packageResult = &PackageResult{
			PackagePath: snapshot.PackagePath,
			Snapshots:   make([]*PerformanceSnapshot, 0),
		}
		tpm.currentSession.PackageResults = append(tpm.currentSession.PackageResults, packageResult)
		tpm.currentSession.TotalPackages++
	}

	packageResult.Snapshots = append(packageResult.Snapshots, snapshot)
	packageResult.TestCount++
	packageResult.Duration += snapshot.Duration

	if snapshot.Success {
		// Update success tracking
	}
}

func (tpm *TestPerformanceMonitor) detectAnomalies(snapshot *PerformanceSnapshot) {
	anomalies := make([]*PerformanceAnomaly, 0)

	// Check duration threshold
	if snapshot.Duration > tpm.thresholds.MaxDurationPerTest {
		anomaly := &PerformanceAnomaly{
			Type:        "duration",
			Severity:    "warning",
			Package:     snapshot.PackagePath,
			TestName:    snapshot.TestName,
			Description: "Test execution time exceeds threshold",
			Value:       snapshot.Duration,
			Threshold:   tpm.thresholds.MaxDurationPerTest,
			Timestamp:   snapshot.Timestamp,
			Suggestions: []string{"Consider test optimization", "Check for resource contention"},
		}
		anomalies = append(anomalies, anomaly)
	}

	// Check memory threshold
	if snapshot.MemoryUsageMB > tpm.thresholds.MaxMemoryUsageMB {
		anomaly := &PerformanceAnomaly{
			Type:        "memory",
			Severity:    "critical",
			Package:     snapshot.PackagePath,
			TestName:    snapshot.TestName,
			Description: "Memory usage exceeds threshold",
			Value:       snapshot.MemoryUsageMB,
			Threshold:   tpm.thresholds.MaxMemoryUsageMB,
			Timestamp:   snapshot.Timestamp,
			Suggestions: []string{"Review memory allocations", "Check for memory leaks"},
		}
		anomalies = append(anomalies, anomaly)
	}

	// Store anomalies
	if len(anomalies) > 0 {
		tpm.mu.Lock()
		if tpm.currentSession != nil {
			tpm.currentSession.Anomalies = append(tpm.currentSession.Anomalies, anomalies...)
		}
		tpm.mu.Unlock()

		// Log anomalies
		for _, anomaly := range anomalies {
			tpm.logger.Warn(fmt.Sprintf("Performance anomaly detected: %s in %s - %s",
				anomaly.Type, anomaly.Package, anomaly.Description))
		}
	}
}

func (tpm *TestPerformanceMonitor) analyzeSessionPerformance(session *TestSession) {
	// Analyze each package's performance
	for _, packageResult := range session.PackageResults {
		tpm.calculatePackageScores(packageResult)
	}
}

func (tpm *TestPerformanceMonitor) calculatePackageScores(packageResult *PackageResult) {
	if len(packageResult.Snapshots) == 0 {
		return
	}

	// Calculate success rate
	successCount := 0
	totalDuration := time.Duration(0)
	totalMemory := int64(0)

	for _, snapshot := range packageResult.Snapshots {
		if snapshot.Success {
			successCount++
		}
		totalDuration += snapshot.Duration
		totalMemory += snapshot.MemoryUsageMB
	}

	successRate := float64(successCount) / float64(len(packageResult.Snapshots))
	avgDuration := totalDuration / time.Duration(len(packageResult.Snapshots))
	avgMemory := totalMemory / int64(len(packageResult.Snapshots))

	// Calculate performance score (0-100)
	performanceScore := 100.0

	// Penalize for low success rate
	performanceScore *= successRate

	// Penalize for high duration (relative to threshold)
	durationPenalty := float64(avgDuration) / float64(tpm.thresholds.MaxDurationPerTest)
	if durationPenalty > 1.0 {
		performanceScore /= durationPenalty
	}

	// Penalize for high memory usage
	memoryPenalty := float64(avgMemory) / float64(tpm.thresholds.MaxMemoryUsageMB)
	if memoryPenalty > 1.0 {
		performanceScore /= memoryPenalty
	}

	packageResult.PerformanceScore = performanceScore
}

func (tpm *TestPerformanceMonitor) generateRecommendations(session *TestSession) {
	recommendations := make([]string, 0)

	// Analyze anomalies for patterns
	if len(session.Anomalies) > 0 {
		memoryIssues := 0
		durationIssues := 0

		for _, anomaly := range session.Anomalies {
			switch anomaly.Type {
			case "memory":
				memoryIssues++
			case "duration":
				durationIssues++
			}
		}

		if memoryIssues > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Consider memory optimization - %d memory issues detected", memoryIssues))
		}

		if durationIssues > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Consider test parallelization - %d duration issues detected", durationIssues))
		}
	}

	// System-level recommendations
	if session.SystemMetrics.AverageCPUUsage < 50.0 {
		recommendations = append(recommendations, "CPU utilization is low - consider increasing test concurrency")
	}

	session.Recommendations = recommendations
}

func (tpm *TestPerformanceMonitor) analyzeTrend(snapshots []*PerformanceSnapshot, extractor func(*PerformanceSnapshot) float64) *TrendData {
	if len(snapshots) < 2 {
		return &TrendData{Direction: "insufficient_data"}
	}

	// Sort by timestamp
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.Before(snapshots[j].Timestamp)
	})

	values := make([]float64, len(snapshots))
	for i, snapshot := range snapshots {
		values[i] = extractor(snapshot)
	}

	// Simple trend analysis
	firstHalf := values[:len(values)/2]
	secondHalf := values[len(values)/2:]

	firstAvg := average(firstHalf)
	secondAvg := average(secondHalf)

	change := secondAvg - firstAvg
	changePercent := 0.0
	if firstAvg != 0 {
		changePercent = (change / firstAvg) * 100
	}

	direction := "stable"
	if changePercent > 5 {
		direction = "increasing"
	} else if changePercent < -5 {
		direction = "decreasing"
	}

	return &TrendData{
		Current:       secondAvg,
		Previous:      firstAvg,
		Change:        change,
		ChangePercent: changePercent,
		Direction:     direction,
		Confidence:    0.8, // Simplified confidence calculation
	}
}

func (tpm *TestPerformanceMonitor) generatePackageReport(packagePath string, snapshots []*PerformanceSnapshot) *PackagePerformanceReport {
	report := &PackagePerformanceReport{
		PackagePath:      packagePath,
		TotalExecutions:  len(snapshots),
		IdentifiedIssues: make([]string, 0),
		Recommendations:  make([]string, 0),
	}

	if len(snapshots) == 0 {
		return report
	}

	// Calculate basic metrics
	successCount := 0
	durations := make([]time.Duration, 0)
	memoryUsages := make([]int64, 0)

	for _, snapshot := range snapshots {
		if snapshot.Success {
			successCount++
		}
		durations = append(durations, snapshot.Duration)
		memoryUsages = append(memoryUsages, snapshot.MemoryUsageMB)
	}

	report.SuccessRate = float64(successCount) / float64(len(snapshots)) * 100

	// Calculate duration statistics
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	report.MedianDuration = durations[len(durations)/2]
	report.P95Duration = durations[int(float64(len(durations))*0.95)]

	// Calculate average duration
	totalDuration := time.Duration(0)
	for _, d := range durations {
		totalDuration += d
	}
	report.AverageDuration = totalDuration / time.Duration(len(durations))

	// Calculate memory statistics
	sort.Slice(memoryUsages, func(i, j int) bool { return memoryUsages[i] < memoryUsages[j] })
	report.PeakMemoryUsage = memoryUsages[len(memoryUsages)-1]

	totalMemory := int64(0)
	for _, m := range memoryUsages {
		totalMemory += m
	}
	report.AverageMemoryUsage = totalMemory / int64(len(memoryUsages))

	// Calculate scores (simplified)
	report.PerformanceScore = calculateScore(report.AverageDuration, tpm.thresholds.MaxDurationPerTest)
	report.StabilityScore = report.SuccessRate
	report.EfficiencyScore = calculateScore(time.Duration(report.AverageMemoryUsage)*time.Millisecond,
		time.Duration(tpm.thresholds.MaxMemoryUsageMB)*time.Millisecond)

	return report
}

func (tpm *TestPerformanceMonitor) calculateGlobalMetrics(report *PerformanceReport) {
	if len(report.PackageReports) == 0 {
		return
	}

	totalSuccess := 0.0
	totalTests := 0
	totalDuration := time.Duration(0)

	for _, pkgReport := range report.PackageReports {
		totalSuccess += pkgReport.SuccessRate * float64(pkgReport.TotalExecutions)
		totalTests += pkgReport.TotalExecutions
		totalDuration += pkgReport.AverageDuration * time.Duration(pkgReport.TotalExecutions)
	}

	report.GlobalMetrics.OverallSuccessRate = totalSuccess / float64(totalTests)
	report.GlobalMetrics.TotalTestTime = totalDuration
	if totalTests > 0 {
		report.GlobalMetrics.AverageTestDuration = totalDuration / time.Duration(totalTests)
	}

	// Identify best and worst performing packages
	if len(report.PackageReports) > 0 {
		report.GlobalMetrics.BestPerformingPackages = []string{report.PackageReports[0].PackagePath}
		report.GlobalMetrics.MostProblematicPackages = []string{
			report.PackageReports[len(report.PackageReports)-1].PackagePath,
		}
	}
}

func (tpm *TestPerformanceMonitor) generateGlobalRecommendations(report *PerformanceReport) {
	recommendations := make([]string, 0)

	if report.GlobalMetrics.OverallSuccessRate < tpm.thresholds.MinSuccessRate {
		recommendations = append(recommendations, "Overall test success rate is below threshold - investigate failing tests")
	}

	if len(report.PackageReports) > 0 {
		avgScore := 0.0
		for _, pkgReport := range report.PackageReports {
			avgScore += pkgReport.PerformanceScore
		}
		avgScore /= float64(len(report.PackageReports))

		if avgScore < 70.0 {
			recommendations = append(recommendations, "Average package performance score is low - consider optimization")
		}
	}

	report.Recommendations = recommendations
}

func (tpm *TestPerformanceMonitor) loadPerformanceHistory() error {
	// Implementation would load historical data from files
	return nil
}

func (tpm *TestPerformanceMonitor) saveSessionData(session *TestSession) error {
	filename := filepath.Join(tpm.metricsDir, fmt.Sprintf("session_%s.json", session.SessionID))
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0600)
}

// Utility functions
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateScore(actual, threshold time.Duration) float64 {
	if threshold == 0 {
		return 100.0
	}
	ratio := float64(actual) / float64(threshold)
	if ratio <= 1.0 {
		return 100.0
	}
	return 100.0 / ratio
}
