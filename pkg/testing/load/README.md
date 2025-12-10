# Load Testing Framework

This directory contains a comprehensive load testing framework for the Freightliner container replication system.

## Components

### Core Testing
- `load_test.go` - Basic load test runner and mock job execution
- `scenarios.go` - Predefined load test scenarios
- `metrics.go` - Performance metrics collection and tracking
- `types.go` - Core types and data structures

### Advanced Testing
- `baseline_establishment.go` - Performance baseline establishment
- `regression_testing.go` - Automated regression testing against baselines
- `benchmarks.go` - Benchmark suite with multiple testing tools
- `scenario_runners.go` - Scenario execution engine
- `k6_generator.go` - k6 load test script generation
- `prometheus_integration.go` - Prometheus metrics integration

## Running Tests

### Quick Tests
```bash
# Run all tests (skips long-running tests)
go test ./pkg/testing/load

# Run with verbose output
go test -v ./pkg/testing/load

# Run specific test
go test -v ./pkg/testing/load -run TestLoadTestHighConcurrency
```

### Full Test Suite
```bash
# Run all tests including long-running ones
go test -v ./pkg/testing/load -timeout 30m

# Run benchmarks
go test -bench=. ./pkg/testing/load
```

## Baseline Establishment

**IMPORTANT**: Performance baselines should **NOT** be established during automated test runs. Baselines require stable, representative hardware and multiple runs to establish accurate performance metrics.

### Why Manual Baseline Establishment?

1. **Hardware Consistency**: Baselines must be established on production or staging hardware that matches the deployment environment
2. **Stability**: No other processes or load should be running during baseline establishment
3. **Statistical Validity**: Multiple runs with proper warmup/cooldown periods are required
4. **Reproducibility**: Baselines serve as the source of truth for regression testing

### How to Establish Baselines

#### Step 1: Deploy to Production/Staging
Deploy your application to a production or staging environment with representative hardware.

#### Step 2: Run Baseline Establishment
```bash
# Create a baseline establishment tool
cd cmd
mkdir -p establish-baselines
cat > establish-baselines/main.go << 'EOF'
package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "time"

    "freightliner/pkg/helper/log"
    "freightliner/pkg/testing/load"
)

func main() {
    resultsDir := flag.String("output", "./baseline-results", "Output directory for baseline results")
    flag.Parse()

    logger := log.NewLoggerWithLevel(log.InfoLevel)

    suite := load.NewBaselineEstablishmentSuite(*resultsDir, logger)

    fmt.Println("Starting baseline establishment...")
    fmt.Println("This will take approximately 2-4 hours depending on configuration")
    fmt.Println("Ensure no other processes are running and system is stable")

    ctx := context.Background()
    report, err := suite.EstablishBaselines(ctx)

    if err != nil {
        fmt.Fprintf(os.Stderr, "Baseline establishment failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("\nBaseline establishment completed!\n")
    fmt.Printf("Duration: %v\n", report.Duration)
    fmt.Printf("Baselines established: %d\n", len(report.EstablishedBaselines))
    fmt.Printf("Results saved to: %s\n", *resultsDir)
    fmt.Println("\nNext steps:")
    fmt.Println("1. Review baseline results in the output directory")
    fmt.Println("2. Copy baseline JSON files to your repository")
    fmt.Println("3. Commit baseline files for use in regression testing")
}
EOF

# Build and run
go build -o establish-baselines ./cmd/establish-baselines
./establish-baselines --output ./baseline-results
```

#### Step 3: Review Baseline Results
```bash
# Review the generated baseline files
ls -lh baseline-results/
cat baseline-results/baseline_establishment_report.json | jq .

# Check baseline validity
grep -r "baseline_status" baseline-results/*.json
```

#### Step 4: Commit Baselines to Repository
```bash
# Copy baseline files to your test data directory
mkdir -p testdata/baselines
cp baseline-results/baseline_*.json testdata/baselines/
cp baseline-results/scalability_limits.json testdata/baselines/

# Commit to repository
git add testdata/baselines/
git commit -m "Add performance baselines for regression testing"
git push
```

### Baseline Configuration

The baseline establishment process uses the following default configuration:

```go
RunsPerScenario:           10,              // 10 runs for statistical validity
WarmupRuns:                3,               // 3 warmup runs to stabilize system
CooldownPeriod:            30 * time.Second // 30s between runs
SystemStabilization:       60 * time.Second // 60s initial stabilization
OutlierThreshold:          2.0,             // 2 standard deviations
ConfidenceLevel:           0.95,            // 95% confidence
AcceptableVariance:        15.0,            // 15% variance acceptable
ScalabilitySteps:          []int{1, 5, 10, 20, 50, 100, 200, 500}
ScalabilityStepDuration:   2 * time.Minute  // 2 minutes per step
```

You can customize these values when creating the baseline establishment suite.

## Regression Testing

Once baselines are established and committed, regression testing will automatically compare new performance results against the baselines.

### Running Regression Tests

```bash
# Regression tests will skip if baselines are not available
go test -v ./pkg/testing/load -run TestRegressionTesting

# With baselines in place
go test -v ./pkg/testing/load -run TestLoadTestFrameworkIntegration
```

### Understanding Regression Test Results

Regression tests will:
1. Load baseline performance data
2. Execute the same scenarios
3. Compare results against baselines
4. Report any performance regressions
5. Generate detailed comparison reports

### Regression Severity Levels

- **None/Stable**: Performance matches baseline within tolerance
- **Minor**: Small degradation (10-20% throughput drop or 20-50% latency increase)
- **Major**: Significant degradation (20-30% throughput drop or 50-100% latency increase)
- **Critical**: Severe degradation (>30% throughput drop or >100% latency increase)

## Test Scenarios

The framework includes several predefined scenarios:

1. **High Volume Replication** - Tests high concurrent replication load
2. **Large Image Stress** - Tests handling of large container images
3. **Network Resilience** - Tests performance under network issues
4. **Burst Replication** - Tests handling of traffic bursts
5. **Sustained Throughput** - Tests long-term sustained performance
6. **Mixed Container Sizes** - Tests with varying image sizes

## Continuous Integration

The test suite is designed to work in CI environments:

- Short tests run in CI by default (`go test ./pkg/testing/load`)
- Long-running tests are skipped in `-short` mode
- Baseline establishment is always skipped in automated tests
- Regression tests skip gracefully if baselines aren't available

### CI Configuration Example

```yaml
# .github/workflows/test.yml
- name: Run load tests
  run: |
    go test -v -short ./pkg/testing/load

- name: Run regression tests (if baselines exist)
  run: |
    if [ -f testdata/baselines/baseline_High\ Volume*.json ]; then
      go test -v ./pkg/testing/load -run TestRegressionTesting
    else
      echo "Baselines not found - skipping regression tests"
    fi
```

## Troubleshooting

### "No baselines available" Error

This is expected if you haven't established baselines yet. Follow the baseline establishment process above.

### High Variance in Results

If baseline establishment reports high variance:
1. Ensure no other processes are running
2. Wait for system to stabilize
3. Increase warmup runs
4. Check for background system tasks (updates, backups, etc.)

### Tests Timing Out in CI

- Use `-short` flag to skip long-running tests
- Adjust timeouts in CI configuration
- Consider running full tests only on specific branches

## Further Reading

- [Load Testing Best Practices](docs/load-testing-best-practices.md)
- [Performance Baseline Guide](docs/baseline-establishment-guide.md)
- [Regression Testing Guide](docs/regression-testing-guide.md)
