# Test Execution Guide

This guide provides instructions for running tests, interpreting results, and troubleshooting test failures in the Freightliner project.

## Quick Start

```bash
# Run all unit tests
make test

# Run with coverage
make test-coverage

# Run with race detection
make test-race

# Run integration tests
make test-integration

# Run benchmarks
make test-benchmark
```

## Test Categories

### 1. Unit Tests

**Purpose**: Test individual functions and methods in isolation

**Command**:
```bash
go test ./...
```

**With Coverage**:
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Coverage Report**:
```bash
# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# View coverage by function
go tool cover -func=coverage.out

# View coverage by package
go tool cover -func=coverage.out | grep -E '(pkg|total)'
```

### 2. Integration Tests

**Purpose**: Test interactions between components and external systems

**Prerequisites**:
- Docker installed and running
- Test registries available (LocalStack, Harbor, etc.)
- Valid cloud credentials (for ECR/GCR tests)

**Command**:
```bash
# Run all integration tests
go test -tags=integration ./tests/integration/...

# Run specific integration test
go test -tags=integration -run TestSingleRepositoryReplication ./tests/integration/...

# With verbose output
go test -v -tags=integration ./tests/integration/...

# With timeout
go test -timeout 30m -tags=integration ./tests/integration/...
```

**Environment Setup**:
```bash
# Start local test registries
docker run -d -p 5000:5000 --name registry-source registry:2
docker run -d -p 5001:5000 --name registry-dest registry:2

# Set environment variables
export SOURCE_REGISTRY=localhost:5000
export DEST_REGISTRY=localhost:5001
export INSECURE_REGISTRIES=true

# For cloud tests
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-east-1
export GCP_PROJECT=your-project
```

**Cleanup**:
```bash
# Stop and remove test registries
docker stop registry-source registry-dest
docker rm registry-source registry-dest
```

### 3. Race Detection Tests

**Purpose**: Detect data races and concurrent access issues

**Command**:
```bash
# Run with race detector
go test -race ./...

# Race tests specifically tagged
go test -race -tags=race ./tests/race/...

# With timeout (race detector is slower)
go test -race -timeout 10m ./...
```

**Important Notes**:
- Race detector increases memory usage (~10x)
- Tests run slower with race detection (~5-10x)
- Use `-timeout` flag to prevent hangs
- CI/CD runs race detection automatically

### 4. Performance Benchmarks

**Purpose**: Measure performance and identify regressions

**Command**:
```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkReplicationThroughput ./tests/performance/...

# Multiple iterations
go test -bench=. -benchmem -count=5 ./tests/performance/...

# Longer benchmark time
go test -bench=. -benchmem -benchtime=10s ./tests/performance/...

# Save benchmark results
go test -bench=. -benchmem ./... > benchmarks.txt
```

**Benchmark Analysis**:
```bash
# Compare benchmarks
go test -bench=. -benchmem ./... > new.txt
benchstat old.txt new.txt

# View only significant changes
benchstat -delta-test=none old.txt new.txt
```

**CPU and Memory Profiling**:
```bash
# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./tests/performance/...
go tool pprof cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof ./tests/performance/...
go tool pprof mem.prof

# Trace profiling
go test -bench=. -trace=trace.out ./tests/performance/...
go tool trace trace.out
```

### 5. E2E Tests

**Purpose**: Test complete workflows from end to end

**Command**:
```bash
# Run E2E tests
go test -tags=e2e -timeout 30m ./tests/e2e/...

# With verbose output
go test -v -tags=e2e ./tests/e2e/...
```

## Test Filtering

### Run Specific Tests

```bash
# By test name
go test -run TestWorkerPool ./pkg/replication/...

# By package
go test ./pkg/replication/...

# Multiple packages
go test ./pkg/replication/... ./pkg/copy/... ./pkg/client/...

# Exclude vendor
go test ./...
```

### Skip Tests

```bash
# Skip slow tests
go test -short ./...

# Skip integration tests
go test -tags='!integration' ./...

# Custom skip logic
# In test file:
if testing.Short() {
    t.Skip("Skipping in short mode")
}
```

## Test Output

### Verbose Output

```bash
# Verbose mode
go test -v ./...

# Show test names only
go test -v ./... | grep -E '(PASS|FAIL)'

# JSON output
go test -json ./...
```

### Parallel Execution

```bash
# Run tests in parallel (default)
go test -parallel 8 ./...

# Disable parallelism
go test -p 1 ./...

# Control parallelism per test
func TestExample(t *testing.T) {
    t.Parallel() // Mark test as parallel-safe
}
```

### Test Caching

```bash
# Disable test caching
go test -count=1 ./...

# Clear test cache
go clean -testcache

# Force rerun all tests
go test -count=1 ./...
```

## CI/CD Integration

### GitHub Actions

The project uses GitHub Actions for automated testing:

**Workflows**:
- `.github/workflows/ci.yml` - Unit and race tests
- `.github/workflows/integration.yml` - Integration tests
- `.github/workflows/benchmark.yml` - Performance benchmarks

**Triggering Tests**:
```bash
# Push triggers all workflows
git push origin feature-branch

# Manual trigger
gh workflow run integration.yml

# View workflow status
gh run list

# View workflow logs
gh run view <run-id>
```

**Status Checks**:
- Unit tests must pass
- Race detection must pass
- Coverage must be ≥85%
- Integration tests must pass (main branch)

### Local CI Simulation

```bash
# Run all CI checks locally
make ci

# Individual checks
make lint
make test
make test-race
make test-integration
make test-coverage
```

## Troubleshooting

### Test Failures

**Timeout Issues**:
```bash
# Increase timeout
go test -timeout 60m ./...

# Debug hanging tests
go test -v -timeout 5m ./... 2>&1 | tee test.log
# Check test.log for last running test
```

**Race Condition Failures**:
```bash
# Run with race detector
go test -race ./pkg/replication/...

# Increase iterations to reproduce
go test -race -count=100 ./pkg/replication/...

# Add delays to expose races
time.Sleep(100 * time.Millisecond)
```

**Flaky Tests**:
```bash
# Run multiple times to reproduce
go test -count=20 -run TestFlakyTest ./...

# Stress test
go test -count=100 -parallel 10 -run TestFlakyTest ./...

# With race detector
go test -race -count=50 -run TestFlakyTest ./...
```

**Memory Issues**:
```bash
# Check memory usage
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Reduce parallel execution
go test -parallel 2 ./...

# Run tests sequentially
go test -p 1 ./...
```

### Integration Test Issues

**Registry Connection Failed**:
```bash
# Check registry is running
docker ps | grep registry

# Test registry connectivity
curl http://localhost:5000/v2/

# Check logs
docker logs registry-source
```

**Authentication Failed**:
```bash
# Verify credentials
echo $AWS_ACCESS_KEY_ID
aws sts get-caller-identity

# For GCR
gcloud auth list
gcloud projects list
```

**Image Not Found**:
```bash
# Check source registry
curl http://localhost:5000/v2/_catalog
curl http://localhost:5000/v2/test/alpine/tags/list

# Repopulate test images
make setup-test-images
```

### Coverage Issues

**Low Coverage**:
```bash
# Identify uncovered code
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v 100.0%

# Generate coverage report
go tool cover -html=coverage.out

# Check specific package
go test -cover ./pkg/replication/...
```

**Coverage Not Updated**:
```bash
# Clear cache and rerun
go clean -testcache
go test -coverprofile=coverage.out ./...
```

## Best Practices

### Writing Tests

1. **Use table-driven tests** for multiple scenarios
2. **Isolate dependencies** with mocks and interfaces
3. **Test error paths** as thoroughly as happy paths
4. **Clean up resources** in `defer` or `t.Cleanup()`
5. **Use descriptive test names** that explain what is being tested
6. **Mark tests as parallel** when safe: `t.Parallel()`
7. **Skip expensive tests** in short mode: `testing.Short()`

### Running Tests

1. **Run tests frequently** during development
2. **Use race detector** regularly: `go test -race`
3. **Check coverage** before committing: `make test-coverage`
4. **Run integration tests** before pushing: `make test-integration`
5. **Profile performance** for optimization: `go test -bench`

### Test Maintenance

1. **Keep tests fast** - unit tests < 1s, integration < 30s
2. **Fix flaky tests** immediately
3. **Update tests** when changing functionality
4. **Remove obsolete tests** to reduce maintenance burden
5. **Review test failures** in CI before merging

## Test Metrics

### Coverage Targets

| Package | Target | Current |
|---------|--------|---------|
| pkg/replication | 90% | TBD |
| pkg/security | 90% | TBD |
| pkg/client | 85% | TBD |
| pkg/copy | 85% | TBD |
| pkg/server | 80% | TBD |
| Overall | 85% | TBD |

### Performance Baselines

| Operation | Target | Unit |
|-----------|--------|------|
| Image Copy (100MB) | < 2s | latency |
| Worker Pool (100 jobs) | < 5s | completion |
| Concurrent Replication | > 50MB/s | throughput |
| Memory Usage | < 500MB | per worker |

## Makefile Targets

```makefile
# Test targets
test:              # Run unit tests
test-coverage:     # Run with coverage
test-race:         # Run with race detector
test-integration:  # Run integration tests
test-benchmark:    # Run benchmarks
test-all:          # Run all tests

# Coverage targets
coverage-html:     # Generate HTML coverage report
coverage-func:     # Show function coverage

# CI targets
ci:               # Run all CI checks
lint:             # Run linters
fmt:              # Format code
vet:              # Run go vet
```

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Go Race Detector](https://golang.org/doc/articles/race_detector)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Project Test Strategy](/docs/testing/test-strategy.md)

## Getting Help

- Check test logs for detailed error messages
- Review GitHub Actions workflow runs
- Search existing issues for similar problems
- Ask in team chat or open a discussion

## Summary

This guide covers:
- ✅ Running different types of tests
- ✅ Using race detection
- ✅ Performance benchmarking
- ✅ Troubleshooting test failures
- ✅ CI/CD integration
- ✅ Best practices

For comprehensive testing strategy, see [test-strategy.md](./test-strategy.md).
