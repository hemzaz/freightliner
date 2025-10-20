# Performance Test Command

Run comprehensive performance tests and benchmarks for Freightliner replication operations.

## What This Command Does

1. Runs Go benchmarks for critical paths
2. Executes load tests with realistic scenarios
3. Analyzes performance metrics and bottlenecks
4. Compares against performance baselines
5. Generates performance report with recommendations

## Usage

```bash
/performance-test [scenario]
```

## Scenarios

- `benchmarks` - Go benchmark tests
- `load` - Load testing with multiple workers
- `stress` - Stress testing to find limits
- `baseline` - Establish new baseline metrics
- `regression` - Compare against baseline
- `all` - Run all performance tests (default)

## Test Execution

### 1. Go Benchmarks
```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific package benchmarks
go test -bench=. -benchmem ./pkg/copy/
go test -bench=. -benchmem ./pkg/replication/

# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./pkg/copy/
go tool pprof cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof ./pkg/copy/
go tool pprof mem.prof
```

### 2. Load Testing
```bash
# High-volume replication scenario
go test -v -run TestLoadReplication ./pkg/testing/load/

# Check monitoring dashboard
open http://localhost:2112/metrics
```

### 3. Stress Testing
- Incrementally increase concurrent workers
- Monitor memory and CPU usage
- Identify breaking point and resource limits
- Test graceful degradation

### 4. Integration Performance
```bash
# Set up local registries
./scripts/setup-test-registries.sh

# Run performance integration tests
make test-integration

# Cleanup
./scripts/setup-test-registries.sh --cleanup
```

## Metrics to Collect

### Throughput Metrics
- Images replicated per second
- Bytes transferred per second
- Tags processed per minute

### Latency Metrics
- P50, P95, P99 replication latency
- Authentication latency
- Registry API response times

### Resource Metrics
- Memory usage (RSS, heap)
- CPU utilization
- Goroutine count
- Network bandwidth usage

### Error Metrics
- Error rate percentage
- Retry count
- Timeout occurrences

## Performance Targets

Based on `docs/PRODUCTION_READINESS_REPORT.md`:

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Throughput | 125 MB/s | 149.17 MB/s | ✅ |
| Response Time | <200ms | <50ms | ✅ |
| Error Rate | <5% | <2% | ✅ |
| Availability | 99.5% | 99.9% | ✅ |

## Analysis Steps

### 1. Identify Bottlenecks
- CPU-bound operations
- Memory allocations
- Network I/O wait
- Lock contention
- Registry API rate limits

### 2. Profile Hot Paths
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=BenchmarkCopyImage
go tool pprof -http=:8080 cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=BenchmarkCopyImage
go tool pprof -http=:8080 mem.prof

# Trace analysis
go test -trace=trace.out -bench=BenchmarkCopyImage
go tool trace trace.out
```

### 3. Optimization Opportunities
- Connection pooling improvements
- Batch operations
- Caching strategies
- Compression optimization
- Worker pool tuning

## Report Generation

Create report in `docs/performance-test-<date>.md`:

1. **Executive Summary**
   - Overall performance assessment
   - Key findings
   - Recommendations

2. **Benchmark Results**
   - Detailed benchmark output
   - Comparison with baseline
   - Performance regressions identified

3. **Load Test Results**
   - Throughput achieved
   - Resource utilization
   - Error rates

4. **Bottleneck Analysis**
   - Hot paths identified
   - Resource constraints
   - Optimization recommendations

5. **Recommendations**
   - Code optimizations
   - Infrastructure tuning
   - Configuration changes
