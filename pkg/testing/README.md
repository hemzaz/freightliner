# Testing Strategy and Mock Implementation Guide

This document outlines the comprehensive testing strategy for the Freightliner container replication project, including mock implementations, testing patterns, and best practices.

## Table of Contents

1. [Overview](#overview)
2. [Test Architecture](#test-architecture)
3. [Mock Strategy](#mock-strategy)
4. [Test Categories](#test-categories)
5. [CI/CD Integration](#cicd-integration)
6. [Load Testing](#load-testing)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

## Overview

The Freightliner testing strategy focuses on:

- **Reliability**: Tests should be deterministic and not depend on external services
- **Performance**: Load testing ensures the system can handle high-volume replication
- **Coverage**: Comprehensive test coverage across all components
- **Maintainability**: Clear, readable tests that serve as documentation
- **CI/CD Integration**: Automated testing with proper feedback loops

## Test Architecture

### Directory Structure

```
pkg/
├── testing/
│   ├── mocks/           # Mock implementations
│   │   ├── aws_mocks.go
│   │   ├── gcp_mocks.go
│   │   └── registry_mocks.go
│   ├── load/            # Load testing utilities
│   │   └── load_test.go
│   ├── helper.go        # Test utilities
│   └── manifest.go      # Test manifest system
├── client/
│   ├── ecr/
│   │   ├── client_test.go          # Original tests
│   │   └── client_improved_test.go # Tests with mocks
│   └── gcr/
│       ├── client_test.go
│       └── client_improved_test.go
├── replication/
│   ├── worker_pool_test.go          # Original tests
│   └── worker_pool_improved_test.go # Race-condition fixes
└── ...
```

### Test Manifest System

The test manifest system (`test-manifest.yaml`) provides fine-grained control over test execution:

- **Environment-based filtering**: Different test sets for CI, local, and integration environments
- **Category-based filtering**: Group tests by type (unit, integration, flaky, etc.)
- **Individual test control**: Enable/disable specific tests with reasons
- **Reporting**: Detailed reporting on skipped tests and reasons

## Mock Strategy

### Design Principles

1. **Interface Segregation**: Mocks implement only the interfaces they need
2. **Behavioral Testing**: Mocks verify interactions, not just state
3. **Realistic Responses**: Mock responses should match real API behavior
4. **Error Scenarios**: Comprehensive error condition testing
5. **Performance**: Mocks should be fast and not introduce delays

### AWS Service Mocks

#### ECR Client Mock (`pkg/testing/mocks/aws_mocks.go`)

```go
// Example usage
mockECR := mocks.NewMockECRClient().
    ExpectDescribeRepositories(repos, nil).
    ExpectGetAuthorizationToken(token, nil).
    Build()

// Test with mock
result, err := mockECR.DescribeRepositories(ctx, input)
```

**Features:**
- Full ECR API coverage (DescribeRepositories, GetAuthorizationToken, etc.)
- Realistic AWS error simulation
- Builder pattern for easy test setup
- Pre-built common scenarios

#### STS Client Mock

```go
mockSTS := mocks.NewMockSTSClient().
    ExpectGetCallerIdentity(identity, nil).
    Build()
```

### GCP Service Mocks

#### Google Container Registry Mock (`pkg/testing/mocks/gcp_mocks.go`)

```go
scenarios := &mocks.MockGCRTestScenarios{}
catalogClient, authTransport := scenarios.SuccessfulGCRListRepositories("project-id", 5)
```

**Features:**
- GCR Catalog API mocking
- Artifact Registry API mocking
- Authentication flow simulation
- Realistic Google API responses

### Mock Builders and Scenarios

#### Builder Pattern

```go
// Fluent interface for setting up expectations
mockClient := mocks.NewMockECRClient().
    ExpectDescribeRepositories(repos, nil).
    ExpectGetAuthorizationToken(token, nil).
    Build()
```

#### Common Scenarios

```go
// Pre-built scenarios for common test cases
scenarios := &mocks.MockGCRTestScenarios{}

// Success scenarios
catalogClient, authTransport := scenarios.SuccessfulGCRListRepositories("project", 10)

// Error scenarios  
catalogClient, authTransport := scenarios.FailedGCRAuthentication()
```

## Test Categories

### Unit Tests
- **Purpose**: Test individual components in isolation
- **Characteristics**: Fast, deterministic, no external dependencies
- **Mocking**: Heavy use of mocks for dependencies
- **Examples**: Helper functions, utility classes, business logic

### Integration Tests
- **Purpose**: Test component interactions
- **Characteristics**: Use real services when possible, longer execution time
- **Environment**: Requires docker registries and network access
- **Examples**: End-to-end replication workflows, client integration

### Load Tests
- **Purpose**: Validate performance under high load
- **Characteristics**: Long-running, resource-intensive
- **Metrics**: Throughput, latency, error rates, resource usage
- **Examples**: High-volume replication, concurrent operations

### Flaky Tests
- **Purpose**: Tests that are inherently unstable
- **Handling**: Retry logic, conditional execution, improved reliability
- **Categories**: Timing-sensitive, external dependency, concurrency tests

## CI/CD Integration

### Pipeline Structure

The enhanced CI pipeline (`/.github/workflows/ci-enhanced.yml`) provides:

1. **Quick Checks**: Fast feedback (formatting, basic validation)
2. **Unit Tests**: Parallel execution by test group
3. **Integration Tests**: With proper service setup
4. **Load Tests**: Scheduled and on-demand
5. **Security Scanning**: Vulnerability and code analysis

### Test Execution Strategy

```bash
# Fast unit tests (CI environment)
./scripts/test-reliable.sh --env ci --categories unit --parallel 4

# Integration tests with retries
./scripts/test-reliable.sh --env integration --max-retries 3

# Load tests (scheduled)
./scripts/test-reliable.sh --env integration --categories load_test
```

### Reliability Features

- **Retry Logic**: Automatic retry for flaky tests
- **Timeout Management**: Configurable timeouts per test type
- **Parallel Execution**: Faster feedback through parallelization
- **Environment Detection**: Automatic environment configuration

## Load Testing

### Load Test Framework (`pkg/testing/load/`)

#### Configuration
```go
config := LoadTestConfig{
    ConcurrentJobs:     50,              // Parallel operations
    RepositoriesPerJob: 10,              // Repos per job
    TestDuration:       5 * time.Minute, // Test duration
    ErrorRate:          0.05,            // Expected 5% error rate
    MetricsInterval:    10 * time.Second, // Metrics collection
}
```

#### Metrics Collection
- **Throughput**: Operations per second
- **Latency**: Min/max/average response times
- **Error Rates**: By error type
- **Resource Usage**: Memory, goroutines, etc.
- **Concurrency**: Actual vs. configured limits

#### Performance Benchmarks
```go
func BenchmarkReplicationLoad(b *testing.B) {
    config := LoadTestConfig{...}
    runner := NewLoadTestRunner(config, logger)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        metrics := runner.Run()
        validateMetrics(b, metrics)
    }
}
```

### Network Performance Testing (`pkg/network/performance_test.go`)

- **Compression Performance**: Various payload sizes and compression levels
- **Concurrent Operations**: Thread safety and performance under load
- **Memory Usage**: Memory consumption patterns
- **Throughput Measurement**: MB/s for different scenarios

## Best Practices

### Writing Effective Tests

1. **Arrange-Act-Assert Pattern**
   ```go
   func TestMyFunction(t *testing.T) {
       // Arrange
       input := "test data"
       expected := "expected result"
       
       // Act
       result := MyFunction(input)
       
       // Assert
       assert.Equal(t, expected, result)
   }
   ```

2. **Table-Driven Tests**
   ```go
   tests := []struct {
       name     string
       input    string
       expected string
       wantErr  bool
   }{
       {"success case", "input", "output", false},
       {"error case", "bad input", "", true},
   }
   
   for _, tc := range tests {
       t.Run(tc.name, func(t *testing.T) {
           result, err := Function(tc.input)
           if tc.wantErr {
               assert.Error(t, err)
           } else {
               assert.NoError(t, err)
               assert.Equal(t, tc.expected, result)
           }
       })
   }
   ```

3. **Mock Verification**
   ```go
   mockClient := &MockClient{}
   mockClient.On("Method", mock.Anything).Return(result, nil)
   
   // Use mock
   service := NewService(mockClient)
   service.DoSomething()
   
   // Verify interactions
   mockClient.AssertExpectations(t)
   ```

### Test Organization

1. **One Test Package Per Source Package**
2. **Separate Test Files for Different Concerns**
   - `*_test.go`: Standard tests
   - `*_integration_test.go`: Integration tests
   - `*_load_test.go`: Load tests
   - `*_improved_test.go`: Enhanced versions with mocks

3. **Test Naming Convention**
   - `TestFunctionName_Condition_ExpectedBehavior`
   - `TestUserService_InvalidInput_ReturnsError`

### Error Testing

1. **Test All Error Paths**
   ```go
   // Test success
   result, err := Function(validInput)
   assert.NoError(t, err)
   
   // Test various error conditions
   _, err = Function(invalidInput)
   assert.Error(t, err)
   assert.Contains(t, err.Error(), "expected error message")
   ```

2. **Mock Error Scenarios**
   ```go
   mockClient.On("Method", mock.Anything).Return(nil, errors.New("network error"))
   ```

### Performance Testing

1. **Benchmark Critical Paths**
   ```go
   func BenchmarkCriticalFunction(b *testing.B) {
       b.ResetTimer()
       for i := 0; i < b.N; i++ {
           CriticalFunction(testData)
       }
   }
   ```

2. **Memory Allocation Testing**
   ```go
   func BenchmarkMemoryAllocation(b *testing.B) {
       b.ReportAllocs()  // Report memory allocations
       for i := 0; i < b.N; i++ {
           result := ExpensiveFunction()
           _ = result // Prevent optimization
       }
   }
   ```

## Troubleshooting

### Common Test Issues

1. **Race Conditions**
   - **Symptom**: Tests pass locally but fail in CI
   - **Solution**: Use proper synchronization, atomic operations
   - **Detection**: Run with `-race` flag

2. **Flaky Tests**
   - **Symptom**: Intermittent failures
   - **Solution**: Add retry logic, better synchronization
   - **Management**: Use test manifest to control execution

3. **Slow Tests**
   - **Symptom**: Tests take too long to execute
   - **Solution**: Optimize algorithms, reduce I/O, use mocks
   - **Monitoring**: Set appropriate timeouts

4. **Mock Assertion Failures**
   - **Symptom**: Mock expectations not met
   - **Solution**: Verify call order, parameters, call count
   - **Debugging**: Enable mock debugging, add logging

### Debugging Test Failures

1. **Enable Verbose Output**
   ```bash
   go test -v ./...
   ./scripts/test-reliable.sh --verbose
   ```

2. **Run Specific Tests**
   ```bash
   go test -run TestSpecificTest ./pkg/package
   ```

3. **Debug Race Conditions**
   ```bash
   go test -race ./...
   ```

4. **Profile Performance**
   ```bash
   go test -bench=. -cpuprofile cpu.prof -memprofile mem.prof
   ```

### Environment Issues

1. **Docker Registry Not Available**
   - **Check**: Registry health in CI logs
   - **Solution**: Increase health check timeout, verify port mapping

2. **Network Timeouts**
   - **Check**: Network connectivity in test environment
   - **Solution**: Increase timeouts, add retry logic

3. **Resource Constraints**
   - **Check**: Available memory, CPU in CI environment
   - **Solution**: Reduce concurrent operations, optimize resource usage

## Migration Guide

### Updating Existing Tests

1. **Add Mock Dependencies**
   ```go
   import "freightliner/pkg/testing/mocks"
   ```

2. **Replace Real Clients with Mocks**
   ```go
   // Before
   client := ecr.NewClient(options)
   
   // After (in tests)
   mockClient := mocks.NewMockECRClient().
       ExpectDescribeRepositories(repos, nil).
       Build()
   ```

3. **Use Test Manifest**
   ```bash
   # Replace direct go test calls
   go test ./...
   
   # With manifest-based execution
   ./scripts/test-with-manifest.sh --env ci
   ```

### Adding New Tests

1. **Choose Appropriate Category**
   - Unit test with mocks for isolated testing
   - Integration test for component interaction
   - Load test for performance validation

2. **Follow Naming Conventions**
3. **Add to Test Manifest** if special handling needed
4. **Include Error Scenarios**
5. **Add Performance Benchmarks** for critical paths

## Conclusion

This testing strategy provides a comprehensive framework for ensuring the reliability, performance, and maintainability of the Freightliner container replication system. By following these patterns and practices, developers can create effective tests that provide confidence in the system's behavior while maintaining fast feedback cycles.

For questions or improvements to this testing strategy, please refer to the project's contribution guidelines.