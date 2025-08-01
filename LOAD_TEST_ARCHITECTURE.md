# Load Testing Framework Architecture

## Overview

The Freightliner Load Testing Framework is a comprehensive performance testing solution designed to validate the container registry replication system's capability to achieve 100-150 MB/s throughput targets while maintaining memory efficiency below 1GB peak usage.

## Architecture Components

### 1. Load Test Scenarios (`scenarios.go`)

**Purpose**: Defines realistic, production-like test scenarios that validate different aspects of the container registry system.

**Key Features**:
- **High-Volume Replication**: 1000+ repositories with mixed container sizes
- **Large Image Stress Testing**: 50 repositories with 5GB+ images
- **Network Resilience Testing**: Simulated packet loss and service interruptions
- **Burst Replication**: Sudden high-volume transfer bursts
- **Sustained Throughput**: Continuous high-throughput operations
- **Mixed Container Sizes**: Realistic production-like size distributions

**Validation Targets**:
- Throughput: 100-150 MB/s sustained
- Memory: <1GB peak usage
- Concurrency: 200-500 concurrent operations  
- Reliability: >99% success rate
- Error Rate: <0.1% failure rate

### 2. Scenario Runners (`scenario_runners.go`)

**Purpose**: Executes specific load test scenarios with detailed performance tracking and validation.

**Key Capabilities**:
- **Parallel Processing**: Simulates 5-concurrent tag replication
- **Connection Pooling**: Validates 200 max connections with 80%+ reuse
- **Memory Monitoring**: Tracks streaming buffer efficiency (50MB chunks)
- **Network Simulation**: Packet loss, latency, and bandwidth constraints
- **Retry Logic**: Tests resilience with intelligent backoff

**Performance Tracking**:
```go
type LoadTestResults struct {
    ThroughputMBps      float64
    MemoryUsageMB       int64
    ConnectionReuseRate float64
    FailureRate         float64
    P99LatencyMs        int64
    ValidationPassed    bool
}
```

### 3. Benchmark Suite (`benchmarks.go`)

**Purpose**: Integrates multiple performance testing tools for comprehensive validation.

**Supported Tools**:
- **k6**: JavaScript-based load testing for realistic HTTP scenarios
- **Apache Bench**: HTTP server stress testing  
- **Go Benchmarks**: Native Go performance testing

**Features**:
- **Cross-Platform**: Automatic tool detection and execution
- **Parallel Execution**: Concurrent benchmark runs
- **Result Aggregation**: Unified reporting across tools
- **Baseline Comparison**: Automated regression detection

### 4. K6 Script Generation (`k6_generator.go`)

**Purpose**: Automatically generates k6 JavaScript test scripts for container registry scenarios.

**Generated Scripts Include**:
- Container registry authentication simulation
- Manifest and layer transfer operations
- Parallel tag processing validation
- Network failure simulation
- Memory pressure testing

**Example k6 Script Structure**:
```javascript
export let options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '5m', target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    'replication_throughput_mbps': ['avg>100'],
    'replication_failures': ['rate<0.01'],
  },
};
```

### 5. Prometheus Integration (`prometheus_integration.go`)

**Purpose**: Provides real-time metrics collection, monitoring, and alerting for load tests.

**Metrics Collected**:
- **Throughput**: MB/s per scenario
- **Memory Usage**: Peak and average memory consumption
- **Connection Statistics**: Reuse rates and pool utilization
- **Latency Percentiles**: P50, P95, P99 response times
- **Failure Rates**: Success/failure tracking per scenario

**HTTP Endpoints**:
- `/metrics`: Prometheus-formatted metrics
- `/metrics/scenarios`: Scenario-specific data
- `/metrics/performance`: Performance trends
- `/metrics/regression`: Regression analysis
- `/dashboard/data`: Dashboard integration data

**Alert Thresholds**:
```go
type AlertThresholds struct {
    ThroughputDropPercent    float64 // 20% throughput drop
    LatencyIncreasePercent   float64 // 50% latency increase
    MemoryIncreasePercent    float64 // 30% memory increase
    FailureRateThreshold     float64 // 5% failure rate
}
```

### 6. Regression Testing (`regression_testing.go`)

**Purpose**: Automated performance regression detection and alerting.

**Automation Features**:
- **Scheduled Testing**: Daily/weekly regression runs
- **Git Integration**: Trigger tests on code changes
- **Baseline Comparison**: Statistical regression analysis
- **Alert Management**: Multi-channel notifications (Slack, email, webhook)

**Regression Detection**:
```go
type RegressionMetrics struct {
    ThroughputChange    float64 // Percentage change
    LatencyChange       float64 // Percentage change  
    MemoryChange        float64 // Percentage change
    Severity           string  // "minor", "major", "critical"
}
```

**Configuration Options**:
- Parallel vs sequential test execution
- Automatic baseline updates
- Retention policies
- Notification channels

### 7. Baseline Establishment (`baseline_establishment.go`)

**Purpose**: Establishes statistically valid performance baselines and scalability limits.

**Statistical Analysis**:
- **Multiple Runs**: 10 measurement runs with 3 warmup runs
- **Outlier Detection**: 2 standard deviation threshold
- **Confidence Intervals**: 95% statistical confidence
- **Variance Analysis**: <15% acceptable variance

**Scalability Testing**:
- **Concurrency Steps**: Tests at 1, 5, 10, 20, 50, 100, 200, 500 concurrent operations
- **Breaking Point Detection**: Identifies performance degradation points
- **Operating Point Recommendation**: Safe operational parameters

**Baseline Validation**:
```go
type EstablishedBaseline struct {
    ThroughputStats     PerformanceStats
    LatencyStats        LatencyStats  
    MemoryStats         PerformanceStats
    StatisticalValidity bool
    ConfidenceLevel     float64
}
```

## Integration with Existing Optimizations

### Memory-Profiler Coordination
- **50MB Buffer Chunks**: Aligned with memory-profiler streaming optimization
- **GC Coordination**: Reduced pressure through aligned buffer lifecycle
- **LRU Cache Integration**: 100-entry transport cache with 1-hour TTL

### Network-Optimizer Integration  
- **Parallel Tag Processing**: Validates 5-concurrent vs sequential performance
- **Connection Pool Testing**: Tests 200 max connections with HTTP/2
- **Compression Validation**: Streaming compression with bandwidth monitoring
- **Retry Logic Testing**: Registry-specific retry patterns with jitter

## Performance Validation Targets

### Throughput Validation
- **Target**: 100-150 MB/s sustained transfer rate
- **Validation**: 5x improvement over 20 MB/s baseline
- **Measurement**: Real-time throughput monitoring per scenario

### Memory Efficiency Validation
- **Target**: <1GB peak usage during high-volume operations
- **Validation**: 85-90% reduction from 4-8GB baseline
- **Measurement**: Memory tracking with 50MB buffer alignment

### Concurrency Validation  
- **Target**: 200-500 concurrent operations
- **Validation**: Parallel processing vs sequential baseline
- **Measurement**: Connection pool utilization and reuse rates

### Reliability Validation
- **Target**: >99% success rate under load
- **Validation**: <0.1% failure rate with retry mechanisms
- **Measurement**: Error tracking and retry success rates

## Usage Examples

### Running Individual Scenarios
```go
// Create high-volume replication scenario
scenario := CreateHighVolumeReplicationScenario()

// Execute scenario
runner := NewScenarioRunner(scenario, logger)
result, err := runner.Run()

// Validate results
if result.AverageThroughputMBps < 100.0 {
    log.Error("Throughput below target")
}
```

### Running Complete Benchmark Suite
```go
// Initialize benchmark suite
suite := NewBenchmarkSuite(resultsDir, logger)

// Run all benchmarks
report, err := suite.RunFullBenchmarkSuite(scenarios)

// Check validation results
if report.ValidationSummary.PassRate < 0.95 {
    log.Warn("Low validation pass rate")
}
```

### Establishing Baselines
```go
// Create baseline establishment suite
suite := NewBaselineEstablishmentSuite(resultsDir, logger)

// Establish baselines for all scenarios
report, err := suite.EstablishBaselines(ctx)

// Review scalability limits
maxConcurrency := report.ScalabilityLimits.MaxTotalConcurrency
recommendedOp := report.ScalabilityLimits.RecommendedOperatingPoint
```

### Automated Regression Testing
```go
// Initialize regression testing
regSuite := NewRegressionTestSuite(baselineDir, logger)

// Start automated testing
err := regSuite.StartAutomatedTesting(ctx)

// Manual regression test
result, err := regSuite.RunRegressionTest(ctx, "code-change")
```

## Monitoring and Alerting

### Prometheus Metrics Integration
```bash
# View metrics
curl http://localhost:8080/metrics

# Scenario-specific metrics
curl http://localhost:8080/metrics/scenarios

# Performance trends
curl http://localhost:8080/metrics/performance
```

### Alert Configuration
```go
alertThresholds := AlertThresholds{
    ThroughputDropPercent:      20.0, // Alert on 20% drop
    LatencyIncreasePercent:     50.0, // Alert on 50% increase
    MemoryIncreasePercent:      30.0, // Alert on 30% increase
    FailureRateThreshold:       0.05, // Alert above 5% failures
}
```

## File Structure

```
pkg/testing/load/
├── scenarios.go                 # Test scenario definitions
├── scenario_runners.go          # Scenario execution logic
├── benchmarks.go               # Multi-tool benchmark suite
├── k6_generator.go             # k6 script generation
├── prometheus_integration.go   # Metrics and monitoring
├── regression_testing.go       # Automated regression testing
├── baseline_establishment.go   # Statistical baseline establishment
└── integration_test.go         # Comprehensive integration tests
```

## Configuration Files

### Test Configuration
```yaml
scenarios:
  high_volume_replication:
    duration: 2h
    concurrent_workers: 50
    expected_throughput: 125.0
    memory_limit_mb: 1024
    
  large_image_stress:
    duration: 90m
    concurrent_workers: 10
    expected_throughput: 140.0
    memory_limit_mb: 1024
```

### Regression Testing Configuration
```json
{
  "schedule_interval": "24h",
  "trigger_on_code_change": true,
  "throughput_regression_threshold": 10.0,
  "latency_regression_threshold": 20.0,
  "memory_regression_threshold": 15.0,
  "notification_channels": ["slack", "email"]
}
```

## Testing and Validation

### Integration Tests
- **Framework Integration**: End-to-end testing of all components
- **Concurrency Testing**: Validates thread-safe operations
- **Stress Testing**: High-frequency metrics collection
- **Resource Cleanup**: Proper cleanup and resource management

### Benchmark Tests
- **Scenario Creation**: Performance of scenario generation
- **Statistical Calculation**: Mathematical computation efficiency
- **Metrics Collection**: High-throughput data recording

### Validation Criteria
- **Statistical Validity**: 95% confidence intervals
- **Performance Consistency**: <15% variance across runs
- **Resource Efficiency**: Memory and CPU utilization
- **Error Handling**: Graceful failure recovery

## Deployment and Operations

### Prerequisites
- Go 1.21+
- k6 (optional, for k6 benchmarks)
- Apache Bench (optional, for ab benchmarks)
- Prometheus (for metrics storage)

### Installation
```bash
# Run load tests
go test ./pkg/testing/load -v

# Run with specific scenarios
go test ./pkg/testing/load -run TestHighVolumeReplication

# Run benchmark suite
go test ./pkg/testing/load -bench=.
```

### Production Deployment
- Deploy Prometheus metrics endpoint
- Configure automated regression testing
- Set up alerting channels
- Establish baseline performance data

## Performance Targets Summary

| Metric | Target | Validation Method |
|--------|--------|------------------|
| **Throughput** | 100-150 MB/s | Real-time monitoring |
| **Memory Usage** | <1GB peak | Streaming buffer tracking |
| **Concurrency** | 200-500 operations | Connection pool metrics |
| **Success Rate** | >99% | Error rate monitoring |
| **Latency P99** | <5000ms | Percentile tracking |
| **Connection Reuse** | >80% | Pool utilization metrics |

This comprehensive load testing framework validates the memory and network optimizations implemented by previous agents, ensuring the Freightliner container registry system meets its performance objectives while maintaining reliability and efficiency.