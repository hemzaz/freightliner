package load

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
)

// BaselineEstablishmentSuite handles establishing performance baselines and scalability limits
//
// IMPORTANT: Baseline establishment should NOT be performed during automated test runs.
// Baselines must be established manually in production or staging environments with:
//   - Stable, representative hardware matching production
//   - No other load or background processes
//   - Multiple runs to ensure statistical validity
//   - Proper warmup and cooldown periods
//
// To establish baselines:
//  1. Deploy application to production/staging environment
//  2. Run: go run cmd/establish-baselines/main.go (or equivalent tool)
//  3. Review baseline results in the output directory
//  4. Commit baseline JSON files to the repository
//  5. Use committed baselines for regression testing
//
// Regression tests will skip gracefully if baseline data is not available.
type BaselineEstablishmentSuite struct {
	logger          log.Logger
	resultsDir      string
	baselineManager *BaselineManager

	// Configuration
	config BaselineEstablishmentConfig

	// Results storage
	establishedBaselines map[string]EstablishedBaseline
	scalabilityLimits    ScalabilityLimits
	resultsMutex         sync.RWMutex

	// Integration components
	scenarioRunner *ScenarioRunner
	benchmarkSuite *BenchmarkSuite
}

// BaselineEstablishmentConfig defines configuration for baseline establishment
type BaselineEstablishmentConfig struct {
	// Test execution configuration
	RunsPerScenario     int           `json:"runs_per_scenario"`
	WarmupRuns          int           `json:"warmup_runs"`
	CooldownPeriod      time.Duration `json:"cooldown_period"`
	SystemStabilization time.Duration `json:"system_stabilization"`

	// Statistical analysis
	OutlierThreshold   float64 `json:"outlier_threshold"`   // Standard deviations for outlier detection
	ConfidenceLevel    float64 `json:"confidence_level"`    // Statistical confidence level
	AcceptableVariance float64 `json:"acceptable_variance"` // Maximum acceptable variance percentage

	// Scalability testing
	ScalabilitySteps          []int         `json:"scalability_steps"`            // Concurrency levels to test
	ScalabilityStepDuration   time.Duration `json:"scalability_step_duration"`    // Duration for each step
	ScalabilityMaxFailureRate float64       `json:"scalability_max_failure_rate"` // Max failure rate before limit

	// Baseline validation
	ValidationRuns      int     `json:"validation_runs"`      // Runs to validate established baselines
	ValidationTolerance float64 `json:"validation_tolerance"` // Acceptable deviation from baseline

	// Reporting
	GenerateReport         bool `json:"generate_report"`
	IncludeDetailedMetrics bool `json:"include_detailed_metrics"`
	SaveRawData            bool `json:"save_raw_data"`
}

// EstablishedBaseline represents a validated performance baseline
type EstablishedBaseline struct {
	ScenarioName  string     `json:"scenario_name"`
	EstablishedAt time.Time  `json:"established_at"`
	SystemInfo    SystemInfo `json:"system_info"`

	// Performance metrics with statistical analysis
	ThroughputStats PerformanceStats    `json:"throughput_stats"`
	LatencyStats    LatencyStats        `json:"latency_stats"`
	MemoryStats     PerformanceStats    `json:"memory_stats"`
	ConnectionStats ConnectionPoolStats `json:"connection_stats"`

	// Reliability metrics
	FailureRateStats PerformanceStats `json:"failure_rate_stats"`
	RetryStats       RetryStats       `json:"retry_stats"`

	// Test execution details
	RunsCompleted       int     `json:"runs_completed"`
	OutliersRemoved     int     `json:"outliers_removed"`
	StatisticalValidity bool    `json:"statistical_validity"`
	ConfidenceLevel     float64 `json:"confidence_level"`

	// Raw measurements (if configured to save)
	RawMeasurements []LoadTestResults `json:"raw_measurements,omitempty"`

	// Baseline validation
	ValidationResults []ValidationResult `json:"validation_results"`
	BaselineStatus    string             `json:"baseline_status"` // "provisional", "validated", "deprecated"
}

// PerformanceStats contains statistical analysis of performance metrics
type PerformanceStats struct {
	Mean                   float64            `json:"mean"`
	Median                 float64            `json:"median"`
	StandardDeviation      float64            `json:"standard_deviation"`
	Min                    float64            `json:"min"`
	Max                    float64            `json:"max"`
	Percentile95           float64            `json:"percentile_95"`
	Percentile99           float64            `json:"percentile_99"`
	CoefficientOfVariation float64            `json:"coefficient_of_variation"`
	ConfidenceInterval     ConfidenceInterval `json:"confidence_interval"`
}

// LatencyStats contains detailed latency statistics
type LatencyStats struct {
	P50Stats  PerformanceStats `json:"p50_stats"`
	P90Stats  PerformanceStats `json:"p90_stats"`
	P95Stats  PerformanceStats `json:"p95_stats"`
	P99Stats  PerformanceStats `json:"p99_stats"`
	P999Stats PerformanceStats `json:"p999_stats"`
	MaxStats  PerformanceStats `json:"max_stats"`
}

// ConnectionPoolStats contains connection pool performance statistics
type ConnectionPoolStats struct {
	ReuseRateStats       PerformanceStats `json:"reuse_rate_stats"`
	MaxConnectionsStats  PerformanceStats `json:"max_connections_stats"`
	PoolUtilizationStats PerformanceStats `json:"pool_utilization_stats"`
}

// RetryStats contains retry mechanism statistics
type RetryStats struct {
	RetryRateStats    PerformanceStats `json:"retry_rate_stats"`
	RetrySuccessStats PerformanceStats `json:"retry_success_stats"`
	AvgRetriesStats   PerformanceStats `json:"avg_retries_stats"`
}

// ConfidenceInterval represents statistical confidence interval
type ConfidenceInterval struct {
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Level      float64 `json:"level"`
}

// ValidationResult represents baseline validation outcome
type ValidationResult struct {
	ValidationRun    int       `json:"validation_run"`
	Timestamp        time.Time `json:"timestamp"`
	ThroughputMatch  bool      `json:"throughput_match"`
	LatencyMatch     bool      `json:"latency_match"`
	MemoryMatch      bool      `json:"memory_match"`
	OverallMatch     bool      `json:"overall_match"`
	DeviationPercent float64   `json:"deviation_percent"`
}

// ScalabilityLimits defines the scalability characteristics of the system
type ScalabilityLimits struct {
	EstablishedAt time.Time  `json:"established_at"`
	SystemInfo    SystemInfo `json:"system_info"`

	// Concurrency limits per scenario
	ScenarioLimits map[string]ScenarioScalabilityLimit `json:"scenario_limits"`

	// Overall system limits
	MaxTotalConcurrency    int     `json:"max_total_concurrency"`
	MaxSustainedThroughput float64 `json:"max_sustained_throughput_mbps"`
	MaxMemoryUtilization   int64   `json:"max_memory_utilization_mb"`

	// Breaking points
	ThroughputBreakingPoint ConcurrencyBreakingPoint `json:"throughput_breaking_point"`
	LatencyBreakingPoint    ConcurrencyBreakingPoint `json:"latency_breaking_point"`
	MemoryBreakingPoint     ConcurrencyBreakingPoint `json:"memory_breaking_point"`

	// Resource utilization limits
	CPUUtilizationLimit     float64 `json:"cpu_utilization_limit"`
	NetworkUtilizationLimit float64 `json:"network_utilization_limit"`
	DiskIOPSLimit           float64 `json:"disk_iops_limit"`

	// Recommendations
	RecommendedOperatingPoint OperatingPoint `json:"recommended_operating_point"`
	SafetyMargins             SafetyMargins  `json:"safety_margins"`
}

// ScenarioScalabilityLimit defines scalability limits for a specific scenario
type ScenarioScalabilityLimit struct {
	ScenarioName        string  `json:"scenario_name"`
	MaxConcurrency      int     `json:"max_concurrency"`
	OptimalConcurrency  int     `json:"optimal_concurrency"`
	MaxThroughput       float64 `json:"max_throughput_mbps"`
	ThroughputAtOptimal float64 `json:"throughput_at_optimal_mbps"`
	LatencyAtOptimal    int64   `json:"latency_at_optimal_ms"`
	MemoryAtOptimal     int64   `json:"memory_at_optimal_mb"`
	BreakingPointReason string  `json:"breaking_point_reason"`
}

// ConcurrencyBreakingPoint identifies where performance degrades
type ConcurrencyBreakingPoint struct {
	ConcurrencyLevel   int     `json:"concurrency_level"`
	MetricValue        float64 `json:"metric_value"`
	DegradationPercent float64 `json:"degradation_percent"`
	Reason             string  `json:"reason"`
}

// OperatingPoint defines recommended operating parameters
type OperatingPoint struct {
	Concurrency        int     `json:"concurrency"`
	ExpectedThroughput float64 `json:"expected_throughput_mbps"`
	ExpectedLatency    int64   `json:"expected_latency_ms"`
	ExpectedMemory     int64   `json:"expected_memory_mb"`
	UtilizationPercent float64 `json:"utilization_percent"`
}

// SafetyMargins defines safety margins for production operation
type SafetyMargins struct {
	ConcurrencyMargin float64 `json:"concurrency_margin"` // Percentage below max
	ThroughputMargin  float64 `json:"throughput_margin"`  // Percentage below max
	MemoryMargin      float64 `json:"memory_margin"`      // Percentage below max
	LatencyMargin     float64 `json:"latency_margin"`     // Percentage above baseline
}

// SystemInfo captures system information during baseline establishment
type SystemInfo struct {
	Timestamp        time.Time `json:"timestamp"`
	HostName         string    `json:"hostname"`
	OS               string    `json:"os"`
	Architecture     string    `json:"architecture"`
	CPUCores         int       `json:"cpu_cores"`
	MemoryGB         int       `json:"memory_gb"`
	NetworkBandwidth string    `json:"network_bandwidth"`
	GoVersion        string    `json:"go_version"`
	GitCommit        string    `json:"git_commit,omitempty"`
	GitBranch        string    `json:"git_branch,omitempty"`
}

// NewBaselineEstablishmentSuite creates a new baseline establishment suite
func NewBaselineEstablishmentSuite(resultsDir string, logger log.Logger) *BaselineEstablishmentSuite {
	if logger == nil {
		logger = log.NewLoggerWithLevel(log.InfoLevel)
	}

	config := BaselineEstablishmentConfig{
		RunsPerScenario:           10, // 10 runs for statistical validity
		WarmupRuns:                3,  // 3 warmup runs
		CooldownPeriod:            30 * time.Second,
		SystemStabilization:       60 * time.Second,
		OutlierThreshold:          2.0,  // 2 standard deviations
		ConfidenceLevel:           0.95, // 95% confidence
		AcceptableVariance:        15.0, // 15% variance
		ScalabilitySteps:          []int{1, 5, 10, 20, 50, 100, 200, 500},
		ScalabilityStepDuration:   2 * time.Minute,
		ScalabilityMaxFailureRate: 0.05, // 5% failure rate
		ValidationRuns:            5,    // 5 validation runs
		ValidationTolerance:       10.0, // 10% tolerance
		GenerateReport:            true,
		IncludeDetailedMetrics:    true,
		SaveRawData:               true,
	}

	return &BaselineEstablishmentSuite{
		logger:               logger,
		resultsDir:           resultsDir,
		baselineManager:      &BaselineManager{baselineDir: resultsDir, logger: logger},
		config:               config,
		establishedBaselines: make(map[string]EstablishedBaseline),
	}
}

// EstablishBaselines runs the complete baseline establishment process
func (bes *BaselineEstablishmentSuite) EstablishBaselines(ctx context.Context) (*BaselineEstablishmentReport, error) {
	bes.logger.Info(fmt.Sprintf("Starting baseline establishment process with runs_per_scenario=%d warmup_runs=%d confidence_level=%.2f",
		bes.config.RunsPerScenario, bes.config.WarmupRuns, bes.config.ConfidenceLevel))

	report := &BaselineEstablishmentReport{
		StartTime:     time.Now(),
		SystemInfo:    bes.getSystemInfo(),
		Configuration: bes.config,
	}

	// Define scenarios for baseline establishment
	scenarios := []ScenarioConfig{
		CreateHighVolumeReplicationScenario(),
		CreateLargeImageStressScenario(),
		CreateNetworkResilienceScenario(),
		CreateBurstReplicationScenario(),
		CreateSustainedThroughputScenario(),
		CreateMixedContainerSizesScenario(),
	}

	// Establish baselines for each scenario
	for _, scenario := range scenarios {
		bes.logger.Info(fmt.Sprintf("Establishing baseline for scenario: %s", scenario.Name))

		baseline, err := bes.establishScenarioBaseline(ctx, scenario)
		if err != nil {
			bes.logger.Error(fmt.Sprintf("Failed to establish baseline for scenario: %s", scenario.Name), err)
			continue
		}

		bes.resultsMutex.Lock()
		bes.establishedBaselines[scenario.Name] = baseline
		bes.resultsMutex.Unlock()

		// Save individual baseline
		if err := bes.saveBaseline(baseline); err != nil {
			bes.logger.Error(fmt.Sprintf("Failed to save baseline for scenario: %s", scenario.Name), err)
		}
	}

	// Establish scalability limits
	bes.logger.Info("Establishing scalability limits")
	scalabilityLimits, err := bes.establishScalabilityLimits(ctx, scenarios)
	if err != nil {
		bes.logger.Error("Failed to establish scalability limits", err)
	} else {
		bes.scalabilityLimits = scalabilityLimits
		if err := bes.saveScalabilityLimits(scalabilityLimits); err != nil {
			bes.logger.Error("Failed to save scalability limits", err)
		}
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)
	report.EstablishedBaselines = bes.establishedBaselines
	report.ScalabilityLimits = bes.scalabilityLimits

	// Generate comprehensive report
	if bes.config.GenerateReport {
		if err := bes.generateBaselineReport(report); err != nil {
			bes.logger.Error("Failed to generate baseline report", err)
		}
	}

	bes.logger.Info(fmt.Sprintf("Baseline establishment completed duration=%v baselines=%d limits=%d", report.Duration, len(report.EstablishedBaselines), report.ScalabilityLimits.MaxTotalConcurrency))

	return report, nil
}

// establishScenarioBaseline establishes a baseline for a single scenario
func (bes *BaselineEstablishmentSuite) establishScenarioBaseline(ctx context.Context, scenario ScenarioConfig) (EstablishedBaseline, error) {
	baseline := EstablishedBaseline{
		ScenarioName:      scenario.Name,
		EstablishedAt:     time.Now(),
		SystemInfo:        bes.getSystemInfo(),
		ConfidenceLevel:   bes.config.ConfidenceLevel,
		ValidationResults: make([]ValidationResult, 0),
	}

	var measurements []LoadTestResults

	// System stabilization period
	bes.logger.Info(fmt.Sprintf("System stabilization period duration=%v", bes.config.SystemStabilization))
	time.Sleep(bes.config.SystemStabilization)

	// Warmup runs
	bes.logger.Info(fmt.Sprintf("Running warmup tests warmup_runs=%d", bes.config.WarmupRuns))
	for i := 0; i < bes.config.WarmupRuns; i++ {
		runner := NewScenarioRunner(scenario, bes.logger)
		_, err := runner.Run()
		if err != nil {
			bes.logger.Warn(fmt.Sprintf("Warmup run failed run=%d error=%v", i+1, err))
		}

		// Cooldown between runs
		time.Sleep(bes.config.CooldownPeriod)
	}

	// Baseline measurement runs
	bes.logger.Info(fmt.Sprintf("Running baseline measurement tests measurement_runs=%d", bes.config.RunsPerScenario))

	for i := 0; i < bes.config.RunsPerScenario; i++ {
		bes.logger.Debug(fmt.Sprintf("Baseline measurement run=%d scenario=%s", i+1, scenario.Name))

		runner := NewScenarioRunner(scenario, bes.logger)
		result, err := runner.Run()
		if err != nil {
			bes.logger.Error(fmt.Sprintf("Baseline measurement run failed run=%d", i+1), err)
			continue
		}

		measurements = append(measurements, *result)

		// Cooldown between runs
		time.Sleep(bes.config.CooldownPeriod)
	}

	baseline.RunsCompleted = len(measurements)

	if len(measurements) < 3 {
		return baseline, fmt.Errorf("insufficient valid measurements: %d", len(measurements))
	}

	// Remove outliers
	measurements, outliersRemoved := bes.removeOutliers(measurements)
	baseline.OutliersRemoved = outliersRemoved

	// Calculate statistics
	baseline.ThroughputStats = bes.calculatePerformanceStats(bes.extractThroughput(measurements))
	baseline.MemoryStats = bes.calculatePerformanceStats(bes.extractMemory(measurements))
	baseline.FailureRateStats = bes.calculatePerformanceStats(bes.extractFailureRate(measurements))
	baseline.ConnectionStats = bes.calculateConnectionStats(measurements)
	baseline.LatencyStats = bes.calculateLatencyStats(measurements)

	// Check statistical validity
	baseline.StatisticalValidity = bes.checkStatisticalValidity(baseline)

	// Save raw measurements if configured
	if bes.config.SaveRawData {
		baseline.RawMeasurements = measurements
	}

	// Validate baseline
	if err := bes.validateBaseline(ctx, scenario, &baseline); err != nil {
		bes.logger.Warn(fmt.Sprintf("Baseline validation failed scenario=%s error=%v", scenario.Name, err))
		baseline.BaselineStatus = "provisional"
	} else {
		baseline.BaselineStatus = "validated"
	}

	return baseline, nil
}

// establishScalabilityLimits determines the scalability limits of the system
func (bes *BaselineEstablishmentSuite) establishScalabilityLimits(ctx context.Context, scenarios []ScenarioConfig) (ScalabilityLimits, error) {
	limits := ScalabilityLimits{
		EstablishedAt:  time.Now(),
		SystemInfo:     bes.getSystemInfo(),
		ScenarioLimits: make(map[string]ScenarioScalabilityLimit),
	}

	// Test each scenario at different concurrency levels
	for _, scenario := range scenarios {
		bes.logger.Info(fmt.Sprintf("Testing scalability limits scenario=%s steps=%d", scenario.Name, bes.config.ScalabilitySteps))

		scenarioLimit, err := bes.findScenarioScalabilityLimit(ctx, scenario)
		if err != nil {
			bes.logger.Error(fmt.Sprintf("Failed to find scalability limit scenario=%s", scenario.Name), err)
			continue
		}

		limits.ScenarioLimits[scenario.Name] = scenarioLimit

		// Update overall limits
		if scenarioLimit.MaxConcurrency > limits.MaxTotalConcurrency {
			limits.MaxTotalConcurrency = scenarioLimit.MaxConcurrency
		}
		if scenarioLimit.MaxThroughput > limits.MaxSustainedThroughput {
			limits.MaxSustainedThroughput = scenarioLimit.MaxThroughput
		}
	}

	// Calculate recommended operating point and safety margins
	limits.RecommendedOperatingPoint = bes.calculateRecommendedOperatingPoint(limits)
	limits.SafetyMargins = bes.calculateSafetyMargins(limits)

	return limits, nil
}

// findScenarioScalabilityLimit finds the scalability limit for a specific scenario
func (bes *BaselineEstablishmentSuite) findScenarioScalabilityLimit(ctx context.Context, scenario ScenarioConfig) (ScenarioScalabilityLimit, error) {
	limit := ScenarioScalabilityLimit{
		ScenarioName: scenario.Name,
	}

	var bestThroughput float64
	_ = 0 // placeholder for optimalConcurrency tracking
	var results []ScalabilityTestResult

	// Test at each concurrency level
	for _, concurrency := range bes.config.ScalabilitySteps {
		bes.logger.Debug(fmt.Sprintf("Testing concurrency level scenario=%s concurrency=%d", scenario.Name, concurrency))

		// Modify scenario for this concurrency level
		testScenario := scenario
		testScenario.ConcurrentWorkers = concurrency
		testScenario.Duration = bes.config.ScalabilityStepDuration

		runner := NewScenarioRunner(testScenario, bes.logger)
		result, err := runner.Run()
		if err != nil {
			bes.logger.Error(fmt.Sprintf("Scalability test failed scenario=%s concurrency=%d", scenario.Name, concurrency), err)
			continue
		}

		scalabilityResult := ScalabilityTestResult{
			Concurrency:         concurrency,
			ThroughputMBps:      result.AverageThroughputMBps,
			LatencyP99Ms:        result.P99LatencyMs,
			MemoryUsageMB:       result.MemoryUsageMB,
			FailureRate:         result.FailureRate,
			ConnectionReuseRate: result.ConnectionReuseRate,
		}

		results = append(results, scalabilityResult)

		// Check if this is the optimal point (best throughput with acceptable failure rate)
		if result.FailureRate <= bes.config.ScalabilityMaxFailureRate && result.AverageThroughputMBps > bestThroughput {
			bestThroughput = result.AverageThroughputMBps
			// optimalConcurrency = concurrency
			limit.OptimalConcurrency = concurrency
			limit.ThroughputAtOptimal = result.AverageThroughputMBps
			limit.LatencyAtOptimal = result.P99LatencyMs
			limit.MemoryAtOptimal = result.MemoryUsageMB
		}

		// Check for breaking point
		if result.FailureRate > bes.config.ScalabilityMaxFailureRate {
			limit.MaxConcurrency = concurrency - 1 // Previous level was the limit
			limit.BreakingPointReason = fmt.Sprintf("Failure rate exceeded %.1f%% at concurrency %d",
				bes.config.ScalabilityMaxFailureRate*100, concurrency)
			break
		}

		// Check for throughput degradation (indicating resource exhaustion)
		if len(results) >= 2 {
			previousResult := results[len(results)-2]
			if result.AverageThroughputMBps < previousResult.ThroughputMBps*0.9 { // 10% degradation
				limit.MaxConcurrency = previousResult.Concurrency
				limit.BreakingPointReason = fmt.Sprintf("Throughput degraded by %.1f%% at concurrency %d",
					(previousResult.ThroughputMBps-result.AverageThroughputMBps)/previousResult.ThroughputMBps*100,
					concurrency)
				break
			}
		}

		limit.MaxConcurrency = concurrency
		limit.MaxThroughput = math.Max(limit.MaxThroughput, result.AverageThroughputMBps)
	}

	if limit.BreakingPointReason == "" {
		limit.BreakingPointReason = "Reached maximum tested concurrency without breaking point"
	}

	return limit, nil
}

// ScalabilityTestResult represents results from a single scalability test
type ScalabilityTestResult struct {
	Concurrency         int     `json:"concurrency"`
	ThroughputMBps      float64 `json:"throughput_mbps"`
	LatencyP99Ms        int64   `json:"latency_p99_ms"`
	MemoryUsageMB       int64   `json:"memory_usage_mb"`
	FailureRate         float64 `json:"failure_rate"`
	ConnectionReuseRate float64 `json:"connection_reuse_rate"`
}

// Statistical analysis helper methods
func (bes *BaselineEstablishmentSuite) calculatePerformanceStats(values []float64) PerformanceStats {
	if len(values) == 0 {
		return PerformanceStats{}
	}

	sort.Float64s(values)

	stats := PerformanceStats{
		Min:  values[0],
		Max:  values[len(values)-1],
		Mean: bes.calculateMean(values),
	}

	stats.Median = bes.calculatePercentile(values, 0.5)
	stats.Percentile95 = bes.calculatePercentile(values, 0.95)
	stats.Percentile99 = bes.calculatePercentile(values, 0.99)
	stats.StandardDeviation = bes.calculateStandardDeviation(values, stats.Mean)
	stats.CoefficientOfVariation = stats.StandardDeviation / stats.Mean * 100
	stats.ConfidenceInterval = bes.calculateConfidenceInterval(values, bes.config.ConfidenceLevel)

	return stats
}

func (bes *BaselineEstablishmentSuite) calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (bes *BaselineEstablishmentSuite) calculatePercentile(sortedValues []float64, percentile float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	index := percentile * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

func (bes *BaselineEstablishmentSuite) calculateStandardDeviation(values []float64, mean float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += math.Pow(v-mean, 2)
	}
	variance := sum / float64(len(values)-1)
	return math.Sqrt(variance)
}

func (bes *BaselineEstablishmentSuite) calculateConfidenceInterval(values []float64, level float64) ConfidenceInterval {
	if len(values) < 2 {
		return ConfidenceInterval{}
	}

	mean := bes.calculateMean(values)
	stdDev := bes.calculateStandardDeviation(values, mean)

	// Use t-distribution for small samples
	_ = float64(len(values) - 1) // df for t-distribution
	_ = 1.0 - level              // alpha for confidence level

	// Simplified t-value calculation (for larger samples, approaches normal distribution)
	var tValue float64
	if len(values) >= 30 {
		// Normal approximation for large samples
		tValue = 1.96 // 95% confidence
		if level == 0.99 {
			tValue = 2.576
		}
	} else {
		// Simplified t-values for common confidence levels and small samples
		if level == 0.95 {
			tValue = 2.262 // Approximation for df=9 (10 samples)
		} else {
			tValue = 3.250 // 99% confidence, df=9
		}
	}

	margin := tValue * stdDev / math.Sqrt(float64(len(values)))

	return ConfidenceInterval{
		LowerBound: mean - margin,
		UpperBound: mean + margin,
		Level:      level,
	}
}

// Data extraction helper methods
func (bes *BaselineEstablishmentSuite) extractThroughput(measurements []LoadTestResults) []float64 {
	values := make([]float64, len(measurements))
	for i, m := range measurements {
		values[i] = m.AverageThroughputMBps
	}
	return values
}

func (bes *BaselineEstablishmentSuite) extractMemory(measurements []LoadTestResults) []float64 {
	values := make([]float64, len(measurements))
	for i, m := range measurements {
		values[i] = float64(m.MemoryUsageMB)
	}
	return values
}

func (bes *BaselineEstablishmentSuite) extractFailureRate(measurements []LoadTestResults) []float64 {
	values := make([]float64, len(measurements))
	for i, m := range measurements {
		values[i] = m.FailureRate * 100 // Convert to percentage
	}
	return values
}

func (bes *BaselineEstablishmentSuite) calculateConnectionStats(measurements []LoadTestResults) ConnectionPoolStats {
	reuseRates := make([]float64, len(measurements))
	for i, m := range measurements {
		reuseRates[i] = m.ConnectionReuseRate * 100 // Convert to percentage
	}

	return ConnectionPoolStats{
		ReuseRateStats: bes.calculatePerformanceStats(reuseRates),
		// MaxConnectionsStats and PoolUtilizationStats would be populated
		// if this data was available in LoadTestResults
	}
}

func (bes *BaselineEstablishmentSuite) calculateLatencyStats(measurements []LoadTestResults) LatencyStats {
	p99Values := make([]float64, len(measurements))
	for i, m := range measurements {
		p99Values[i] = float64(m.P99LatencyMs)
	}

	return LatencyStats{
		P99Stats: bes.calculatePerformanceStats(p99Values),
		// Other percentile stats would be populated if available in LoadTestResults
	}
}

// Outlier detection and removal
func (bes *BaselineEstablishmentSuite) removeOutliers(measurements []LoadTestResults) ([]LoadTestResults, int) {
	if len(measurements) < 3 {
		return measurements, 0
	}

	// Extract key metrics for outlier detection
	throughputValues := bes.extractThroughput(measurements)

	// Calculate mean and standard deviation
	mean := bes.calculateMean(throughputValues)
	stdDev := bes.calculateStandardDeviation(throughputValues, mean)

	// Identify outliers using the threshold
	var filtered []LoadTestResults
	outliersRemoved := 0

	for i, measurement := range measurements {
		throughput := throughputValues[i]
		zScore := math.Abs(throughput-mean) / stdDev

		if zScore <= bes.config.OutlierThreshold {
			filtered = append(filtered, measurement)
		} else {
			outliersRemoved++
			bes.logger.Debug(fmt.Sprintf("Removed outlier measurement throughput=%.2f mean=%.2f z_score=%.2f", throughput, mean, zScore))
		}
	}

	return filtered, outliersRemoved
}

// Validation methods
func (bes *BaselineEstablishmentSuite) checkStatisticalValidity(baseline EstablishedBaseline) bool {
	// Check coefficient of variation for key metrics
	if baseline.ThroughputStats.CoefficientOfVariation > bes.config.AcceptableVariance {
		return false
	}

	// Check if we have enough measurements
	if baseline.RunsCompleted < 5 {
		return false
	}

	// Check confidence interval width
	ciWidth := baseline.ThroughputStats.ConfidenceInterval.UpperBound - baseline.ThroughputStats.ConfidenceInterval.LowerBound
	return ciWidth/baseline.ThroughputStats.Mean <= 0.2 // 20% of mean
}

func (bes *BaselineEstablishmentSuite) validateBaseline(ctx context.Context, scenario ScenarioConfig, baseline *EstablishedBaseline) error {
	bes.logger.Info(fmt.Sprintf("Validating established baseline scenario=%s validation_runs=%d", scenario.Name, bes.config.ValidationRuns))

	for i := 0; i < bes.config.ValidationRuns; i++ {
		runner := NewScenarioRunner(scenario, bes.logger)
		result, err := runner.Run()
		if err != nil {
			return fmt.Errorf("validation run %d failed: %w", i+1, err)
		}

		// Check if result matches baseline within tolerance
		validation := ValidationResult{
			ValidationRun: i + 1,
			Timestamp:     time.Now(),
		}

		// Throughput validation
		throughputDiff := math.Abs(result.AverageThroughputMBps-baseline.ThroughputStats.Mean) / baseline.ThroughputStats.Mean * 100
		validation.ThroughputMatch = throughputDiff <= bes.config.ValidationTolerance

		// Memory validation
		memoryDiff := math.Abs(float64(result.MemoryUsageMB)-baseline.MemoryStats.Mean) / baseline.MemoryStats.Mean * 100
		validation.MemoryMatch = memoryDiff <= bes.config.ValidationTolerance

		// Overall validation
		validation.OverallMatch = validation.ThroughputMatch && validation.MemoryMatch
		validation.DeviationPercent = math.Max(throughputDiff, memoryDiff)

		baseline.ValidationResults = append(baseline.ValidationResults, validation)
	}

	// Check if majority of validation runs passed
	passedRuns := 0
	for _, validation := range baseline.ValidationResults {
		if validation.OverallMatch {
			passedRuns++
		}
	}

	if float64(passedRuns)/float64(len(baseline.ValidationResults)) < 0.7 { // 70% pass rate
		return fmt.Errorf("validation failed: only %d/%d runs passed", passedRuns, len(baseline.ValidationResults))
	}

	return nil
}

// System information gathering
func (bes *BaselineEstablishmentSuite) getSystemInfo() SystemInfo {
	return SystemInfo{
		Timestamp:        time.Now(),
		HostName:         "localhost", // Would get actual hostname
		OS:               "linux",     // Would get actual OS
		Architecture:     "amd64",     // Would get actual arch
		CPUCores:         8,           // Would get actual CPU cores
		MemoryGB:         16,          // Would get actual memory
		NetworkBandwidth: "1Gbps",     // Would get actual network info
		GoVersion:        "1.21",      // Would get actual Go version
		// GitCommit and GitBranch would be populated from git info
	}
}

// Report generation and utility methods
func (bes *BaselineEstablishmentSuite) calculateRecommendedOperatingPoint(limits ScalabilityLimits) OperatingPoint {
	// Calculate recommended operating point based on safety margins
	totalOptimalConcurrency := 0
	totalExpectedThroughput := 0.0

	for _, scenarioLimit := range limits.ScenarioLimits {
		totalOptimalConcurrency += scenarioLimit.OptimalConcurrency
		totalExpectedThroughput += scenarioLimit.ThroughputAtOptimal
	}

	// Apply safety margin of 20%
	return OperatingPoint{
		Concurrency:        int(float64(totalOptimalConcurrency) * 0.8),
		ExpectedThroughput: totalExpectedThroughput * 0.8,
		UtilizationPercent: 80.0,
	}
}

func (bes *BaselineEstablishmentSuite) calculateSafetyMargins(limits ScalabilityLimits) SafetyMargins {
	return SafetyMargins{
		ConcurrencyMargin: 20.0, // 20% below max
		ThroughputMargin:  20.0, // 20% below max
		MemoryMargin:      15.0, // 15% below max
		LatencyMargin:     25.0, // 25% above baseline acceptable
	}
}

// File I/O methods
func (bes *BaselineEstablishmentSuite) saveBaseline(baseline EstablishedBaseline) error {
	filename := filepath.Join(bes.resultsDir, fmt.Sprintf("baseline_%s.json",
		filepath.Base(baseline.ScenarioName)))

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0600)
}

func (bes *BaselineEstablishmentSuite) saveScalabilityLimits(limits ScalabilityLimits) error {
	filename := filepath.Join(bes.resultsDir, "scalability_limits.json")

	data, err := json.MarshalIndent(limits, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0600)
}

func (bes *BaselineEstablishmentSuite) generateBaselineReport(report *BaselineEstablishmentReport) error {
	filename := filepath.Join(bes.resultsDir, "baseline_establishment_report.json")

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0600)
}

// BaselineEstablishmentReport contains the complete baseline establishment results
type BaselineEstablishmentReport struct {
	StartTime            time.Time                      `json:"start_time"`
	EndTime              time.Time                      `json:"end_time"`
	Duration             time.Duration                  `json:"duration"`
	SystemInfo           SystemInfo                     `json:"system_info"`
	Configuration        BaselineEstablishmentConfig    `json:"configuration"`
	EstablishedBaselines map[string]EstablishedBaseline `json:"established_baselines"`
	ScalabilityLimits    ScalabilityLimits              `json:"scalability_limits"`
	Summary              BaselineEstablishmentSummary   `json:"summary"`
}

// BaselineEstablishmentSummary provides a high-level summary of results
type BaselineEstablishmentSummary struct {
	TotalBaselinesEstablished int     `json:"total_baselines_established"`
	ValidatedBaselines        int     `json:"validated_baselines"`
	ProvisionalBaselines      int     `json:"provisional_baselines"`
	AverageConfidenceLevel    float64 `json:"average_confidence_level"`
	MaxThroughputAchieved     float64 `json:"max_throughput_achieved_mbps"`
	RecommendedMaxConcurrency int     `json:"recommended_max_concurrency"`
	SystemCapacityUtilization float64 `json:"system_capacity_utilization"`
	PerformanceStability      string  `json:"performance_stability"` // "excellent", "good", "acceptable", "poor"
}
