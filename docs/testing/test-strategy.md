# Freightliner Test Strategy

## Executive Summary

This document defines the comprehensive testing strategy for the Freightliner container registry replication system. Our goal is to achieve **85%+ code coverage** with a focus on concurrent operations, multi-registry integration, and performance validation.

## Testing Objectives

1. **Unit Test Coverage**: 85%+ coverage across all packages
2. **Race Detection**: All concurrent operations tested with `-race` flag
3. **Integration Testing**: Multi-registry support with real and simulated registries
4. **Performance Benchmarks**: Throughput, concurrency, memory, and CPU profiling
5. **Security Testing**: Authentication, authorization, and encryption validation

## Test Pyramid

```
         /\
        /E2E\        <- 10% - Full workflow tests
       /------\
      /Integr.\     <- 20% - Multi-registry tests
     /----------\
    /   Unit     \  <- 70% - Fast, focused tests
   /--------------\
```

## 1. Unit Testing Strategy

### Coverage Requirements

- **Target**: 85% overall coverage
- **Minimum per package**: 80%
- **Critical paths**: 95%+ (replication, authentication, encryption)

### Test Organization

```
pkg/
├── client/          # Registry client tests
├── copy/            # Copy operation tests
├── replication/     # Replication logic tests
├── security/        # Security and encryption tests
├── server/          # API server tests
└── helper/          # Utility function tests
```

### Unit Test Patterns

#### 1. Table-Driven Tests

```go
func TestSchedulerAddJob(t *testing.T) {
    tests := []struct {
        name    string
        rule    ReplicationRule
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid job with cron schedule",
            rule: ReplicationRule{
                SourceRegistry:      "ecr",
                SourceRepository:    "app",
                DestinationRegistry: "gcr",
                DestinationRepository: "app",
                Schedule:            "*/5 * * * *",
            },
            wantErr: false,
        },
        {
            name: "missing source registry",
            rule: ReplicationRule{
                SourceRepository: "app",
            },
            wantErr: true,
            errMsg:  "source registry cannot be empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### 2. Mock-Based Testing

```go
func TestWorkerPoolSubmit(t *testing.T) {
    logger := log.NewBasicLogger(log.InfoLevel)
    pool := NewWorkerPool(5, logger)
    pool.Start()
    defer pool.Stop()

    var executed atomic.Bool
    err := pool.Submit("test-job", func(ctx context.Context) error {
        executed.Store(true)
        return nil
    })

    assert.NoError(t, err)
    assert.Eventually(t, func() bool {
        return executed.Load()
    }, 5*time.Second, 100*time.Millisecond)
}
```

#### 3. Race Detection Tests

```go
func TestWorkerPoolConcurrency(t *testing.T) {
    // Run with: go test -race ./...
    pool := NewWorkerPool(10, nil)
    pool.Start()
    defer pool.Stop()

    var counter atomic.Int64
    var wg sync.WaitGroup

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            pool.Submit("job", func(ctx context.Context) error {
                counter.Add(1)
                return nil
            })
        }()
    }

    wg.Wait()
    assert.Equal(t, int64(100), counter.Load())
}
```

## 2. Integration Testing Strategy

### Test Registries

| Registry | Purpose | Implementation |
|----------|---------|----------------|
| AWS ECR | Production testing | Test AWS account |
| Google GCR | Production testing | Test GCP project |
| Docker Hub | Public registry | Anonymous + authenticated |
| Harbor | Self-hosted | Docker Compose in CI |
| LocalStack | ECR simulation | Mock AWS services |
| Registry:2 | Generic OCI | Docker registry container |

### Integration Test Structure

```
tests/integration/
├── registry_integration_test.go  # Main integration harness
├── ecr_test.go                   # AWS ECR specific tests
├── gcr_test.go                   # Google GCR specific tests
├── dockerhub_test.go             # Docker Hub tests
├── harbor_test.go                # Harbor tests
├── ghcr_test.go                  # GitHub Container Registry
├── quay_test.go                  # Quay.io tests
├── acr_test.go                   # Azure Container Registry
├── generic_test.go               # Generic OCI registry
├── oci_artifacts_test.go         # OCI artifacts (SBOM, signatures)
└── cosign_test.go                # Cosign signature tests
```

### Test Scenarios

#### Scenario 1: Single Repository Replication

```go
func TestSingleRepositoryReplication(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup source and destination registries
    source := setupECRRegistry(t)
    dest := setupGCRRegistry(t)

    // Push test image to source
    testImage := "test-app:v1.0.0"
    pushTestImage(t, source, testImage)

    // Replicate
    rule := ReplicationRule{
        SourceRegistry:        source.Name,
        SourceRepository:      "test-app",
        DestinationRegistry:   dest.Name,
        DestinationRepository: "test-app",
    }

    err := replicationService.ReplicateRepository(context.Background(), rule)
    assert.NoError(t, err)

    // Verify image in destination
    verifyImageExists(t, dest, testImage)
}
```

#### Scenario 2: Multi-Repository Batch Replication

```go
func TestBatchReplication(t *testing.T) {
    repos := []string{"app1", "app2", "app3", "app4", "app5"}

    rules := make([]ReplicationRule, len(repos))
    for i, repo := range repos {
        rules[i] = ReplicationRule{
            SourceRegistry:        "ecr",
            SourceRepository:      repo,
            DestinationRegistry:   "gcr",
            DestinationRepository: repo,
            IncludeTags:           []string{"latest", "v*"},
        }
    }

    // Execute batch replication
    results := scheduler.BatchReplicate(context.Background(), rules)

    // Verify all succeeded
    for _, result := range results {
        assert.NoError(t, result.Error)
    }
}
```

#### Scenario 3: Encrypted Replication with KMS

```go
func TestEncryptedReplication(t *testing.T) {
    // Setup encryption manager
    encMgr, err := encryption.NewManager(encryption.ManagerConfig{
        Provider: "aws-kms",
        KeyID:    "arn:aws:kms:us-east-1:123456789012:key/test",
    })
    require.NoError(t, err)

    // Configure replication with encryption
    rule := ReplicationRule{
        SourceRegistry:      "ecr",
        SourceRepository:    "sensitive-app",
        DestinationRegistry: "gcr",
        DestinationRepository: "sensitive-app",
        Encryption: &EncryptionConfig{
            Enabled: true,
            KeyID:   "arn:aws:kms:us-east-1:123456789012:key/test",
        },
    }

    // Execute with encryption
    err = replicationService.ReplicateRepository(context.Background(), rule)
    assert.NoError(t, err)

    // Verify encryption metadata
    verifyEncryptionMetadata(t, "gcr", "sensitive-app")
}
```

#### Scenario 4: Cross-Cloud Replication

```go
func TestCrossCloudReplication(t *testing.T) {
    scenarios := []struct {
        name   string
        source string
        dest   string
    }{
        {"ECR to GCR", "ecr", "gcr"},
        {"GCR to ECR", "gcr", "ecr"},
        {"DockerHub to Harbor", "dockerhub", "harbor"},
        {"GHCR to Quay", "ghcr", "quay"},
    }

    for _, sc := range scenarios {
        t.Run(sc.name, func(t *testing.T) {
            // Test cross-cloud replication
        })
    }
}
```

#### Scenario 5: Resume from Checkpoint

```go
func TestReplicationResume(t *testing.T) {
    // Start replication
    ctx, cancel := context.WithCancel(context.Background())

    go func() {
        time.Sleep(2 * time.Second)
        cancel() // Simulate interruption
    }()

    err := replicationService.ReplicateRepository(ctx, rule)
    assert.Error(t, err) // Expect context cancellation

    // Resume from checkpoint
    err = replicationService.ResumeReplication(context.Background(), rule)
    assert.NoError(t, err)

    // Verify completion
    verifyReplicationComplete(t, rule)
}
```

#### Scenario 6: Concurrent Replication

```go
func TestConcurrentReplication(t *testing.T) {
    workerPool := NewWorkerPool(10, nil)
    workerPool.Start()
    defer workerPool.Stop()

    // Submit 50 concurrent replication jobs
    for i := 0; i < 50; i++ {
        repo := fmt.Sprintf("app-%d", i)
        workerPool.Submit(repo, func(ctx context.Context) error {
            return replicationService.ReplicateRepository(ctx, ReplicationRule{
                SourceRegistry:      "ecr",
                SourceRepository:    repo,
                DestinationRegistry: "gcr",
                DestinationRepository: repo,
            })
        })
    }

    // Wait for completion
    workerPool.Wait()

    // Verify all replications succeeded
    verifyAllReplicationsSucceeded(t, 50)
}
```

#### Scenario 7: Rate Limiting and Retries

```go
func TestRateLimitingAndRetries(t *testing.T) {
    // Configure aggressive rate limiting
    config := &Config{
        RateLimit: RateLimitConfig{
            RequestsPerSecond: 10,
            Burst:             5,
        },
        Retry: RetryConfig{
            MaxAttempts:    3,
            InitialBackoff: 1 * time.Second,
            MaxBackoff:     10 * time.Second,
        },
    }

    // Create service with rate limiting
    svc := NewReplicationService(config, logger)

    // Simulate 100 rapid requests
    var errors atomic.Int32
    var wg sync.WaitGroup

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            if err := svc.ReplicateTag(ctx, rule, "latest"); err != nil {
                errors.Add(1)
            }
        }()
    }

    wg.Wait()

    // All should eventually succeed with retries
    assert.Equal(t, int32(0), errors.Load())
}
```

#### Scenario 8: Authentication Failures

```go
func TestAuthenticationFailures(t *testing.T) {
    tests := []struct {
        name     string
        authType string
        config   AuthConfig
        wantErr  string
    }{
        {
            name:     "invalid ECR credentials",
            authType: "ecr",
            config:   AuthConfig{/* invalid */},
            wantErr:  "authentication failed",
        },
        {
            name:     "expired GCR token",
            authType: "gcr",
            config:   AuthConfig{/* expired */},
            wantErr:  "token expired",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test authentication failure handling
        })
    }
}
```

## 3. Performance Benchmarks

### Benchmark Organization

```
tests/performance/
├── benchmark_test.go           # Main benchmark suite
├── throughput_test.go          # Throughput benchmarks
├── concurrency_test.go         # Concurrent worker scaling
├── memory_test.go              # Memory usage profiling
└── cpu_test.go                 # CPU utilization
```

### Benchmark Patterns

#### Throughput Testing

```go
func BenchmarkReplicationThroughput(b *testing.B) {
    sizes := []int64{
        1 * MB,      // Small image
        100 * MB,    // Medium image
        1 * GB,      // Large image
        5 * GB,      // Very large image
    }

    for _, size := range sizes {
        b.Run(fmt.Sprintf("ImageSize_%dMB", size/MB), func(b *testing.B) {
            b.SetBytes(size)
            b.ResetTimer()

            for i := 0; i < b.N; i++ {
                replicateImage(size)
            }

            // Report throughput
            mbps := float64(b.N*size) / b.Elapsed().Seconds() / MB
            b.ReportMetric(mbps, "MB/s")
        })
    }
}
```

#### Concurrent Worker Scaling

```go
func BenchmarkWorkerPoolScaling(b *testing.B) {
    workerCounts := []int{1, 2, 4, 8, 16, 32, 64}

    for _, count := range workerCounts {
        b.Run(fmt.Sprintf("Workers_%d", count), func(b *testing.B) {
            pool := NewWorkerPool(count, nil)
            pool.Start()
            defer pool.Stop()

            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                pool.Submit(fmt.Sprintf("job-%d", i), func(ctx context.Context) error {
                    // Simulate work
                    time.Sleep(10 * time.Millisecond)
                    return nil
                })
            }

            pool.Wait()

            // Report jobs per second
            jobsPerSec := float64(b.N) / b.Elapsed().Seconds()
            b.ReportMetric(jobsPerSec, "jobs/sec")
        })
    }
}
```

#### Memory Usage Profiling

```go
func BenchmarkMemoryUsage(b *testing.B) {
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        // Allocate and use resources
        data := make([]byte, 100*MB)
        processData(data)
    }

    // Memory metrics automatically reported
}
```

#### CPU Utilization

```go
func BenchmarkCPUUtilization(b *testing.B) {
    pool := NewWorkerPool(runtime.NumCPU(), nil)
    pool.Start()
    defer pool.Stop()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        pool.Submit(fmt.Sprintf("cpu-job-%d", i), func(ctx context.Context) error {
            // CPU-intensive work
            hash := sha256.New()
            data := make([]byte, 1*MB)
            hash.Write(data)
            hash.Sum(nil)
            return nil
        })
    }

    pool.Wait()
}
```

## 4. Test Fixtures and Mocks

### Mock Registry Server

```go
type MockRegistry struct {
    images     map[string][]byte
    manifests  map[string]*Manifest
    blobs      map[string][]byte
    mutex      sync.RWMutex
}

func NewMockRegistry() *MockRegistry {
    return &MockRegistry{
        images:    make(map[string][]byte),
        manifests: make(map[string]*Manifest),
        blobs:     make(map[string][]byte),
    }
}
```

### Test Image Generators

```go
func GenerateTestImage(size int64) (*TestImage, error) {
    layers := []Layer{
        {Size: size / 3, Data: randomData(size / 3)},
        {Size: size / 3, Data: randomData(size / 3)},
        {Size: size / 3, Data: randomData(size / 3)},
    }

    return &TestImage{
        Repository: "test-repo",
        Tag:        "test-tag",
        Layers:     layers,
    }, nil
}
```

## 5. CI/CD Integration

### GitHub Actions Workflows

#### Integration Test Workflow

```yaml
name: Integration Tests
on: [push, pull_request]

jobs:
  integration:
    runs-on: ubuntu-latest
    services:
      localstack:
        image: localstack/localstack:latest
        ports:
          - 4566:4566
      registry:
        image: registry:2
        ports:
          - 5000:5000
      harbor:
        image: goharbor/harbor-core:latest
        ports:
          - 8080:8080

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Run integration tests
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.TEST_AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.TEST_AWS_SECRET_ACCESS_KEY }}
          GCP_PROJECT: ${{ secrets.TEST_GCP_PROJECT }}
        run: |
          go test -v -tags=integration ./tests/integration/...
```

#### Benchmark Workflow

```yaml
name: Performance Benchmarks
on:
  push:
    branches: [main]
  pull_request:

jobs:
  benchmark:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem -benchtime=10s \
            ./tests/performance/... | tee benchmark.txt

      - name: Store benchmark results
        uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: 'go'
          output-file-path: benchmark.txt
```

## 6. Test Execution

### Running Tests Locally

```bash
# Unit tests
go test ./...

# Unit tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Race detection
go test -race ./...

# Integration tests
go test -tags=integration ./tests/integration/...

# Benchmarks
go test -bench=. -benchmem ./tests/performance/...

# Specific package
go test -v ./pkg/replication/...

# Specific test
go test -run TestSchedulerAddJob ./pkg/replication/...
```

### Test Tags

```go
//go:build integration
// +build integration

func TestECRIntegration(t *testing.T) {
    // Integration test
}
```

## 7. Coverage Requirements

### Package-Level Coverage Targets

| Package | Target Coverage | Critical |
|---------|----------------|----------|
| pkg/replication | 90% | Yes |
| pkg/security | 90% | Yes |
| pkg/client | 85% | Yes |
| pkg/copy | 85% | Yes |
| pkg/server | 80% | No |
| pkg/helper | 80% | No |
| cmd | 70% | No |

### Coverage Enforcement

```yaml
# .github/workflows/coverage.yml
- name: Check coverage
  run: |
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//' > coverage.txt
    COVERAGE=$(cat coverage.txt)
    if (( $(echo "$COVERAGE < 85.0" | bc -l) )); then
      echo "Coverage $COVERAGE% is below 85%"
      exit 1
    fi
```

## 8. Test Maintenance

### Test Review Checklist

- [ ] All tests have descriptive names
- [ ] Table-driven tests used where appropriate
- [ ] Mocks used for external dependencies
- [ ] Race conditions tested with `-race`
- [ ] Error paths tested
- [ ] Edge cases covered
- [ ] Performance benchmarks included
- [ ] Integration tests tagged properly
- [ ] Documentation updated

### Continuous Improvement

1. **Weekly**: Review failed tests and flaky tests
2. **Monthly**: Analyze coverage trends
3. **Quarterly**: Review test performance and optimize slow tests
4. **Per Release**: Full integration test suite execution

## 9. Test Reporting

### Coverage Dashboard

```bash
# Generate coverage badge
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total | \
  awk '{print "![Coverage](https://img.shields.io/badge/coverage-"$3"-brightgreen.svg)"}'
```

### Benchmark Dashboard

```bash
# Compare benchmarks
go test -bench=. -benchmem ./... > new.txt
benchcmp old.txt new.txt
```

## 10. Security Testing

### Security Test Scenarios

1. **Authentication bypass attempts**
2. **Token expiration handling**
3. **Permission boundary validation**
4. **Encryption key rotation**
5. **Secrets exposure prevention**
6. **Rate limiting enforcement**
7. **Input validation**
8. **RBAC enforcement**

## Appendix A: Test Data

### Sample Registries

```yaml
test_registries:
  - name: test-ecr
    type: ecr
    region: us-east-1
    account_id: "123456789012"

  - name: test-gcr
    type: gcr
    project: test-project
    location: us

  - name: test-harbor
    type: harbor
    endpoint: http://localhost:8080
```

### Sample Images

```yaml
test_images:
  - name: alpine-small
    size: 5MB
    layers: 1

  - name: ubuntu-medium
    size: 100MB
    layers: 5

  - name: application-large
    size: 1GB
    layers: 20
```

## Appendix B: Common Test Patterns

### Setup and Teardown

```go
func setupTest(t *testing.T) (*TestContext, func()) {
    ctx := &TestContext{
        // Initialize test context
    }

    cleanup := func() {
        // Cleanup resources
    }

    return ctx, cleanup
}

func TestExample(t *testing.T) {
    ctx, cleanup := setupTest(t)
    defer cleanup()

    // Test implementation
}
```

### Retry Logic Testing

```go
func TestWithRetry(t *testing.T) {
    var attempts atomic.Int32

    fn := func() error {
        if attempts.Add(1) < 3 {
            return errors.New("temporary failure")
        }
        return nil
    }

    err := retryWithBackoff(fn, 3, 100*time.Millisecond)
    assert.NoError(t, err)
    assert.Equal(t, int32(3), attempts.Load())
}
```

## Summary

This test strategy ensures comprehensive coverage of the Freightliner system with:

- **85%+ unit test coverage**
- **Multi-registry integration testing**
- **Performance benchmarks and profiling**
- **Race detection for concurrent operations**
- **Automated CI/CD integration**
- **Security and authentication validation**

By following this strategy, we maintain high code quality and catch issues early in the development cycle.
