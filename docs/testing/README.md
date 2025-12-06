# Freightliner Testing Documentation

Complete testing documentation for the Freightliner container registry replication system.

## 📋 Overview

This directory contains comprehensive testing documentation, strategies, and execution guides to ensure high quality and reliability of the Freightliner system.

## 📚 Documentation Structure

### [test-strategy.md](./test-strategy.md)
**Comprehensive Testing Strategy**

Defines the complete testing approach including:
- Unit testing patterns and coverage requirements (85%+)
- Integration test scenarios for multi-registry support
- Performance benchmarking methodology
- Race detection and concurrency testing
- CI/CD integration workflows
- Test fixtures and mock services

**Target Audience**: Developers, QA Engineers, DevOps

### [test-execution.md](./test-execution.md)
**Test Execution Guide**

Practical guide for running tests:
- Quick start commands
- Test category descriptions
- Filtering and running specific tests
- Troubleshooting common issues
- CI/CD workflow integration
- Performance profiling

**Target Audience**: All team members

## 🎯 Testing Objectives

### 1. Coverage Goals
- **Overall**: 85%+ code coverage
- **Critical paths**: 90%+ coverage (replication, security, authentication)
- **Minimum per package**: 80% coverage

### 2. Quality Metrics
- **Zero data races** detected with `-race` flag
- **All integration tests pass** with real registries
- **Performance benchmarks** meet baseline requirements
- **Security tests** validate authentication and encryption

### 3. Test Categories

```
Test Pyramid:
         /\
        /E2E\        <- 10% - End-to-end workflows
       /------\
      /Integr.\     <- 20% - Multi-registry tests
     /----------\
    /   Unit     \  <- 70% - Fast, focused tests
   /--------------\
```

## 🚀 Quick Start

### Prerequisites

```bash
# Install dependencies
go mod download

# Install testing tools
go install github.com/stretchr/testify
go install golang.org/x/tools/cmd/benchcmp
```

### Run All Tests

```bash
# Unit tests
make test

# With coverage
make test-coverage

# With race detection
make test-race

# Integration tests
make test-integration

# Performance benchmarks
make test-benchmark
```

### View Coverage

```bash
# Generate HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

## 📁 Test Organization

```
freightliner/
├── pkg/                          # Package tests
│   ├── client/
│   │   └── *_test.go            # Client unit tests
│   ├── copy/
│   │   └── *_test.go            # Copy operation tests
│   ├── replication/
│   │   └── *_test.go            # Replication logic tests
│   └── security/
│       └── *_test.go            # Security tests
│
├── tests/                        # Dedicated test directory
│   ├── integration/              # Integration tests
│   │   ├── registry_test.go     # Multi-registry harness
│   │   ├── ecr_test.go          # AWS ECR tests
│   │   ├── gcr_test.go          # Google GCR tests
│   │   ├── dockerhub_test.go    # Docker Hub tests
│   │   ├── harbor_test.go       # Harbor tests
│   │   ├── ghcr_test.go         # GitHub CR tests
│   │   ├── quay_test.go         # Quay.io tests
│   │   └── acr_test.go          # Azure CR tests
│   │
│   ├── performance/              # Performance benchmarks
│   │   └── benchmark_test.go    # Benchmark suite
│   │
│   ├── race/                     # Race detection tests
│   │   └── race_test.go         # Concurrent operation tests
│   │
│   ├── e2e/                      # End-to-end tests
│   │   └── e2e_test.go          # Complete workflow tests
│   │
│   └── helpers/                  # Test utilities
│       └── test_fixtures.go     # Test fixtures and mocks
│
├── docs/testing/                 # Testing documentation
│   ├── README.md                # This file
│   ├── test-strategy.md         # Testing strategy
│   └── test-execution.md        # Execution guide
│
└── .github/workflows/            # CI/CD workflows
    ├── ci.yml                   # Unit and race tests
    ├── integration.yml          # Integration tests
    └── benchmark.yml            # Performance benchmarks
```

## 🧪 Test Types

### Unit Tests
**Location**: `pkg/*/` alongside source code
**Tags**: None (run by default)
**Purpose**: Test individual functions in isolation
**Run**: `go test ./...`

**Example**:
```go
func TestWorkerPoolSubmit(t *testing.T) {
    pool := NewWorkerPool(5, nil)
    pool.Start()
    defer pool.Stop()

    err := pool.Submit("job-1", func(ctx context.Context) error {
        return nil
    })
    assert.NoError(t, err)
}
```

### Integration Tests
**Location**: `tests/integration/`
**Tags**: `integration`
**Purpose**: Test interactions with real registries
**Run**: `go test -tags=integration ./tests/integration/...`

**Registries Tested**:
- AWS ECR
- Google GCR
- Docker Hub
- Harbor (self-hosted)
- GitHub Container Registry
- Quay.io
- Azure Container Registry

### Race Detection Tests
**Location**: `tests/race/`
**Tags**: `race`
**Purpose**: Detect data races and concurrent issues
**Run**: `go test -race -tags=race ./tests/race/...`

**What's Tested**:
- Worker pool concurrency
- Scheduler race conditions
- Map concurrent access
- Channel operations
- Context cancellation

### Performance Benchmarks
**Location**: `tests/performance/`
**Tags**: `benchmark`
**Purpose**: Measure and track performance
**Run**: `go test -bench=. -benchmem ./tests/performance/...`

**Metrics**:
- Throughput (MB/s)
- Latency (ms)
- Memory usage
- CPU utilization
- Concurrency scaling

### E2E Tests
**Location**: `tests/e2e/`
**Tags**: `e2e`
**Purpose**: Test complete workflows
**Run**: `go test -tags=e2e ./tests/e2e/...`

## 🔧 Test Utilities

### Test Fixtures (`tests/helpers/test_fixtures.go`)

```go
// Generate test image
image := helpers.GenerateTestImage(t, 100*MB, 5)

// Create mock registry
registry := helpers.NewMockRegistryServer()
registry.AddImage(image)

// Generate test data
data := helpers.GenerateRandomData(t, 1*MB)
```

### Mock Services

```go
// Mock registry client
type MockRegistryClient struct {
    mock.Mock
}

func (m *MockRegistryClient) Repository(name string) Repository {
    args := m.Called(name)
    return args.Get(0).(Repository)
}
```

## 📊 Coverage Requirements

| Package | Minimum | Target | Critical |
|---------|---------|--------|----------|
| pkg/replication | 85% | 90% | Yes |
| pkg/security | 85% | 90% | Yes |
| pkg/client | 80% | 85% | Yes |
| pkg/copy | 80% | 85% | Yes |
| pkg/server | 75% | 80% | No |
| pkg/helper | 75% | 80% | No |
| cmd | 60% | 70% | No |

### Checking Coverage

```bash
# Overall coverage
go test -cover ./...

# Detailed coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Coverage for specific package
go test -cover ./pkg/replication/...

# Generate HTML report
go tool cover -html=coverage.out
```

## 🚦 CI/CD Workflows

### Continuous Integration (`.github/workflows/ci.yml`)
**Triggers**: Push, Pull Request
**Tests**:
- Unit tests
- Race detection
- Code coverage
- Linting

**Status**: ✅ Required to pass

### Integration Tests (`.github/workflows/integration.yml`)
**Triggers**: Push to main, Pull Request, Nightly
**Tests**:
- Local registry integration
- Harbor integration
- Cloud registry integration (ECR, GCR)
- E2E workflows

**Status**: ✅ Required on main

### Performance Benchmarks (`.github/workflows/benchmark.yml`)
**Triggers**: Push, Pull Request, Weekly
**Tests**:
- Micro benchmarks
- Copy performance
- Compression benchmarks
- Memory profiling
- Network performance

**Status**: ℹ️ Informational

## 🐛 Troubleshooting

### Common Issues

#### Test Timeout
```bash
# Increase timeout
go test -timeout 30m ./...
```

#### Race Condition Detected
```bash
# Run with race detector
go test -race ./pkg/replication/...

# Increase iterations
go test -race -count=100 ./...
```

#### Flaky Tests
```bash
# Run multiple times
go test -count=20 -run TestFlakyTest ./...

# With race detector
go test -race -count=50 -run TestFlakyTest ./...
```

#### Integration Test Failed
```bash
# Check registry status
docker ps | grep registry

# View logs
docker logs registry-source

# Restart registries
make test-registry-restart
```

### Getting Help

1. Check test logs for error details
2. Review [test-execution.md](./test-execution.md) for troubleshooting guide
3. Check GitHub Actions workflow logs
4. Search existing issues
5. Ask in team chat

## 📈 Performance Baselines

| Metric | Baseline | Target |
|--------|----------|--------|
| Image Copy (100MB) | 2s | < 1.5s |
| Worker Pool (100 jobs) | 5s | < 3s |
| Concurrent Replication | 50MB/s | > 100MB/s |
| Memory per Worker | 100MB | < 50MB |

## ✅ Best Practices

### Writing Tests
1. ✅ Use table-driven tests
2. ✅ Test error paths
3. ✅ Clean up resources
4. ✅ Use descriptive names
5. ✅ Isolate dependencies
6. ✅ Mark parallel-safe tests
7. ✅ Skip expensive tests in short mode

### Running Tests
1. ✅ Run tests frequently
2. ✅ Use race detector regularly
3. ✅ Check coverage before commit
4. ✅ Run integration tests before push
5. ✅ Profile performance regularly

### Test Maintenance
1. ✅ Keep tests fast
2. ✅ Fix flaky tests immediately
3. ✅ Update tests with code changes
4. ✅ Remove obsolete tests
5. ✅ Review CI failures

## 🔗 Related Documentation

- [Test Strategy](./test-strategy.md) - Comprehensive testing strategy
- [Test Execution](./test-execution.md) - Running and troubleshooting tests
- [Go Testing](https://golang.org/pkg/testing/) - Go testing package
- [Testify](https://github.com/stretchr/testify) - Testing toolkit

## 📅 Test Schedule

### Daily (Automated)
- Unit tests on every commit
- Race detection on every commit
- Coverage checks on every PR

### Weekly (Automated)
- Performance benchmarks (Sunday 3 AM UTC)
- Full integration test suite
- Coverage trend analysis

### Monthly (Manual)
- Performance baseline review
- Test suite optimization
- Flaky test analysis

## 🎯 Success Metrics

### Test Quality
- ✅ 85%+ code coverage
- ✅ Zero race conditions
- ✅ < 1% flaky test rate
- ✅ All integration tests pass

### Test Performance
- ✅ Unit tests < 5 minutes
- ✅ Integration tests < 30 minutes
- ✅ Benchmarks < 20 minutes
- ✅ E2E tests < 15 minutes

### Test Reliability
- ✅ 99%+ CI success rate
- ✅ < 5% test retry rate
- ✅ Zero false positives
- ✅ Fast feedback (< 10 min)

## 📝 Summary

This testing documentation provides:

1. **Comprehensive Strategy** - Detailed testing approach for all components
2. **Practical Execution Guide** - Step-by-step instructions for running tests
3. **Test Organization** - Clear structure for test files and utilities
4. **CI/CD Integration** - Automated testing workflows
5. **Best Practices** - Guidelines for writing and maintaining tests
6. **Troubleshooting** - Solutions for common test issues

**Goal**: Achieve 85%+ coverage with robust, reliable, and fast tests that catch bugs early and provide confidence in code quality.

---

**Last Updated**: 2025-12-05
**Maintained By**: Test Engineering Team
**Version**: 1.0.0
