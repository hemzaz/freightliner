# Performance Baseline Establishment Guide

## Overview

This guide explains how to establish performance baselines for the Freightliner container replication system. Baselines are critical reference points for automated regression testing and performance monitoring.

## Why Manual Baseline Establishment?

Performance baselines **must not** be established during automated test runs because:

1. **Hardware Consistency**: Baselines must reflect production/staging hardware capabilities
2. **Environmental Stability**: Requires isolation from other processes and load
3. **Statistical Validity**: Multiple runs with proper warmup/cooldown are essential
4. **Reproducibility**: Baselines serve as the source of truth for all future comparisons

## Prerequisites

### System Requirements
- Production or staging environment with representative hardware
- No competing processes or background load
- Stable network connectivity
- Sufficient disk space for baseline data (~100MB)
- At least 2-4 hours of dedicated testing time

### Software Requirements
- Go 1.21 or later
- Access to container registries
- Sufficient system resources (CPU, memory, network bandwidth)

## Baseline Establishment Process

### Step 1: Prepare the Environment

```bash
# Ensure system is stable and idle
# Stop unnecessary services
# Clear system caches if needed
sync; echo 3 > /proc/sys/vm/drop_caches  # Linux only

# Verify no competing load
top
htop
```

### Step 2: Create Baseline Establishment Tool

Create a dedicated tool for running baseline establishment:

```bash
mkdir -p cmd/establish-baselines
```

Create `cmd/establish-baselines/main.go`:

```go
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
    // Configuration flags
    resultsDir := flag.String("output", "./baseline-results", "Output directory for baseline results")
    runsPerScenario := flag.Int("runs", 10, "Number of runs per scenario")
    warmupRuns := flag.Int("warmup", 3, "Number of warmup runs")
    flag.Parse()

    // Create logger
    logger := log.NewLoggerWithLevel(log.InfoLevel)

    // Create baseline establishment suite
    suite := load.NewBaselineEstablishmentSuite(*resultsDir, logger)

    // Customize configuration if needed
    suite.config.RunsPerScenario = *runsPerScenario
    suite.config.WarmupRuns = *warmupRuns

    fmt.Println("===== Performance Baseline Establishment =====")
    fmt.Printf("Output Directory: %s\n", *resultsDir)
    fmt.Printf("Runs per Scenario: %d\n", *runsPerScenario)
    fmt.Printf("Warmup Runs: %d\n", *warmupRuns)
    fmt.Println("\nThis process will take approximately 2-4 hours.")
    fmt.Println("Ensure:")
    fmt.Println("  - No other processes are running")
    fmt.Println("  - System is stable and idle")
    fmt.Println("  - Network connectivity is stable")
    fmt.Println("\nPress Ctrl+C to cancel, or wait 10 seconds to start...")

    time.Sleep(10 * time.Second)

    fmt.Println("\nStarting baseline establishment...")
    startTime := time.Now()

    ctx := context.Background()
    report, err := suite.EstablishBaselines(ctx)

    if err != nil {
        fmt.Fprintf(os.Stderr, "\nERROR: Baseline establishment failed: %v\n", err)
        os.Exit(1)
    }

    duration := time.Since(startTime)

    // Print summary
    fmt.Println("\n===== Baseline Establishment Completed =====")
    fmt.Printf("Total Duration: %v\n", duration)
    fmt.Printf("Baselines Established: %d\n", len(report.EstablishedBaselines))
    fmt.Printf("Max Concurrency: %d\n", report.ScalabilityLimits.MaxTotalConcurrency)
    fmt.Printf("Max Throughput: %.2f MB/s\n", report.ScalabilityLimits.MaxSustainedThroughput)

    // Print individual baseline summaries
    fmt.Println("\nBaseline Summary:")
    for name, baseline := range report.EstablishedBaselines {
        fmt.Printf("\n  %s:\n", name)
        fmt.Printf("    Status: %s\n", baseline.BaselineStatus)
        fmt.Printf("    Runs: %d (outliers removed: %d)\n", baseline.RunsCompleted, baseline.OutliersRemoved)
        fmt.Printf("    Throughput: %.2f ± %.2f MB/s\n",
            baseline.ThroughputStats.Mean,
            baseline.ThroughputStats.StandardDeviation)
        fmt.Printf("    Latency P99: %.2f ± %.2f ms\n",
            baseline.LatencyStats.P99Stats.Mean,
            baseline.LatencyStats.P99Stats.StandardDeviation)
        fmt.Printf("    Memory: %.2f ± %.2f MB\n",
            baseline.MemoryStats.Mean,
            baseline.MemoryStats.StandardDeviation)
        fmt.Printf("    Statistical Validity: %v\n", baseline.StatisticalValidity)
    }

    fmt.Printf("\nResults saved to: %s\n", *resultsDir)
    fmt.Println("\nNext Steps:")
    fmt.Println("1. Review baseline results in the output directory")
    fmt.Println("2. Verify all baselines have 'validated' status")
    fmt.Println("3. Copy baseline files to testdata/baselines/ directory")
    fmt.Println("4. Commit baseline files to repository")
    fmt.Println("\nExample commands:")
    fmt.Printf("  mkdir -p testdata/baselines\n")
    fmt.Printf("  cp %s/baseline_*.json testdata/baselines/\n", *resultsDir)
    fmt.Printf("  cp %s/scalability_limits.json testdata/baselines/\n", *resultsDir)
    fmt.Println("  git add testdata/baselines/")
    fmt.Println("  git commit -m 'Add performance baselines'")
}
```

### Step 3: Build and Run

```bash
# Build the baseline establishment tool
cd cmd/establish-baselines
go build -o establish-baselines

# Run with default configuration (recommended for first run)
./establish-baselines --output ../../baseline-results

# Or with custom configuration
./establish-baselines \
  --output ../../baseline-results \
  --runs 15 \
  --warmup 5
```

### Step 4: Monitor Execution

The baseline establishment process will:

1. **System Stabilization** (60 seconds)
   - Allows system to settle before measurements

2. **Warmup Runs** (3 runs per scenario by default)
   - Warms up caches and connection pools
   - Results are not included in baselines

3. **Measurement Runs** (10 runs per scenario by default)
   - Actual performance measurements
   - Outliers are automatically detected and removed

4. **Validation Runs** (5 runs per scenario by default)
   - Validates baseline stability
   - Ensures reproducibility

5. **Scalability Testing**
   - Tests at multiple concurrency levels
   - Identifies breaking points and optimal operating points

### Step 5: Review Results

After completion, review the generated files:

```bash
cd baseline-results

# View the comprehensive report
cat baseline_establishment_report.json | jq .

# View individual baseline files
ls -lh baseline_*.json

# Check baseline status
grep -h "baseline_status" baseline_*.json

# View scalability limits
cat scalability_limits.json | jq .
```

### Step 6: Validate Baselines

Check that baselines meet quality criteria:

```bash
# Check statistical validity
jq '.statistical_validity' baseline_*.json

# Check confidence intervals
jq '.throughput_stats.confidence_interval' baseline_*.json

# Check coefficient of variation (should be < 15%)
jq '.throughput_stats.coefficient_of_variation' baseline_*.json

# Check baseline status (should be "validated")
jq '.baseline_status' baseline_*.json
```

Expected output for validated baselines:
- `statistical_validity: true`
- `coefficient_of_variation: < 15.0`
- `baseline_status: "validated"`

### Step 7: Commit Baselines to Repository

```bash
# Create baselines directory in testdata
mkdir -p testdata/baselines

# Copy baseline files
cp baseline-results/baseline_*.json testdata/baselines/
cp baseline-results/scalability_limits.json testdata/baselines/

# Verify files were copied
ls -lh testdata/baselines/

# Add to git
git add testdata/baselines/

# Commit with descriptive message
git commit -m "Add performance baselines established on $(date +%Y-%m-%d)

- Established baselines for 6 load test scenarios
- Hardware: [describe hardware configuration]
- Environment: [production/staging]
- Baseline status: All validated
- Next regression test will use these baselines as reference"

# Push to repository
git push origin main
```

## Configuration Options

The baseline establishment suite can be customized:

```go
suite.config.RunsPerScenario = 15           // More runs = better statistical validity
suite.config.WarmupRuns = 5                 // More warmup = more stable results
suite.config.CooldownPeriod = 60 * time.Second  // Longer cooldown = better isolation
suite.config.OutlierThreshold = 2.5         // Stricter outlier detection
suite.config.ConfidenceLevel = 0.99         // Higher confidence level (99%)
```

## Troubleshooting

### High Variance in Results

**Symptom**: Coefficient of variation > 15%, statistical validity = false

**Solutions**:
1. Increase warmup runs: `--warmup 5`
2. Increase runs per scenario: `--runs 15`
3. Check for background processes interfering
4. Verify network stability
5. Ensure disk I/O is not saturated

### Baseline Status = "provisional"

**Symptom**: Baseline saved with "provisional" status instead of "validated"

**Solutions**:
1. Review validation results in the baseline JSON
2. Check if validation runs matched baseline within tolerance
3. Increase validation runs
4. Ensure system stability between measurement and validation

### Scalability Limits Not Detected

**Symptom**: Breaking point reason = "Reached maximum tested concurrency without breaking point"

**Solutions**:
1. Increase max concurrency levels in configuration
2. System may handle load better than expected (good!)
3. Review scalability test results to understand system capacity

### Tests Taking Too Long

**Symptom**: Process exceeds expected 2-4 hour window

**Solutions**:
1. Reduce runs per scenario (minimum 5 for validity)
2. Reduce scalability steps
3. Verify no network latency issues
4. Check if warmup period is appropriate

## Baseline Maintenance

### When to Re-establish Baselines

Re-establish baselines when:
1. **Hardware changes**: New servers, CPU/memory upgrades
2. **Major code changes**: Performance optimizations, algorithmic changes
3. **Infrastructure changes**: Network upgrades, storage changes
4. **Significant time passed**: Quarterly or semi-annually
5. **Baseline becomes invalid**: Consistent regression test failures

### Baseline Versioning

Use descriptive commit messages and tags:

```bash
# Tag baseline version
git tag -a baseline-v1.0.0 -m "Initial performance baselines - Q4 2024"
git push origin baseline-v1.0.0
```

### Comparing Baselines Over Time

Keep historical baselines for trend analysis:

```bash
# Archive old baselines before updating
mkdir -p testdata/baselines/archive/2024-Q3
mv testdata/baselines/baseline_*.json testdata/baselines/archive/2024-Q3/

# Commit new baselines
cp baseline-results/baseline_*.json testdata/baselines/
git add testdata/baselines/
git commit -m "Update baselines for Q4 2024"
```

## Integration with CI/CD

### CI Configuration

Baselines should be committed to the repository and used by CI:

```yaml
# .github/workflows/regression-test.yml
name: Performance Regression Tests

on:
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM

jobs:
  regression-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Check for baselines
        id: check-baselines
        run: |
          if [ -d testdata/baselines ] && [ "$(ls -A testdata/baselines)" ]; then
            echo "baselines_exist=true" >> $GITHUB_OUTPUT
          else
            echo "baselines_exist=false" >> $GITHUB_OUTPUT
          fi

      - name: Run regression tests
        if: steps.check-baselines.outputs.baselines_exist == 'true'
        run: |
          go test -v ./pkg/testing/load -run TestRegressionTesting

      - name: Skip regression tests
        if: steps.check-baselines.outputs.baselines_exist == 'false'
        run: |
          echo "No baselines found - skipping regression tests"
          echo "To enable regression testing, establish baselines following the guide"
```

## Best Practices

1. **Consistency**: Always establish baselines on the same hardware/environment
2. **Documentation**: Document hardware specs and conditions in commit messages
3. **Validation**: Always verify baselines have "validated" status
4. **Isolation**: Ensure no other load during baseline establishment
5. **Review**: Have team review baseline results before committing
6. **Versioning**: Tag baseline commits for easy reference
7. **Monitoring**: Track baseline trends over time
8. **Updates**: Re-establish baselines after significant changes

## Further Reading

- [Load Testing Framework README](README.md)
- [Regression Testing Guide](REGRESSION_TESTING_GUIDE.md)
- [Performance Optimization Guide](../../../docs/performance-optimization.md)
