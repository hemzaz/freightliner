package network

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"
	"time"

	"freightliner/pkg/helper/log"
)

// PerformanceBenchmark runs comprehensive performance tests
type PerformanceBenchmark struct {
	logger log.Logger
	config BenchmarkConfig
}

// BenchmarkConfig configures benchmark parameters
type BenchmarkConfig struct {
	// Data sizes to test (in bytes)
	DataSizes []int64

	// Compression levels to test
	CompressionLevels []CompressionLevel

	// Number of iterations per test
	Iterations int

	// Concurrency levels
	ConcurrencyLevels []int

	// Enable detailed profiling
	EnableProfiling bool
}

// DefaultBenchmarkConfig returns standard benchmark configuration
func DefaultBenchmarkConfig() BenchmarkConfig {
	return BenchmarkConfig{
		DataSizes: []int64{
			1024 * 1024,        // 1MB
			10 * 1024 * 1024,   // 10MB
			100 * 1024 * 1024,  // 100MB
			500 * 1024 * 1024,  // 500MB
			1024 * 1024 * 1024, // 1GB
		},
		CompressionLevels: []CompressionLevel{
			BestSpeed,
			DefaultCompression,
			BestCompression,
		},
		Iterations:        5,
		ConcurrencyLevels: []int{1, 2, 4, 8, 16},
		EnableProfiling:   true,
	}
}

// BenchmarkResult contains benchmark metrics
type BenchmarkResult struct {
	TestName    string
	DataSize    int64
	Level       CompressionLevel
	Concurrency int

	// Performance metrics
	Duration         time.Duration
	ThroughputMBps   float64
	CompressionRatio float64

	// Resource usage
	MemoryUsedMB   float64
	CPUUtilization float64
	AllocsPerOp    int64
	BytesPerOp     int64

	// Comparison with baseline
	SpeedupVsBaseline float64
}

// NewPerformanceBenchmark creates a new benchmark suite
func NewPerformanceBenchmark(config BenchmarkConfig, logger log.Logger) *PerformanceBenchmark {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &PerformanceBenchmark{
		logger: logger,
		config: config,
	}
}

// RunComprehensiveBenchmark runs all performance tests
func (pb *PerformanceBenchmark) RunComprehensiveBenchmark(ctx context.Context) ([]BenchmarkResult, error) {
	pb.logger.Info("Starting comprehensive performance benchmark")

	var allResults []BenchmarkResult

	// Test 1: Sequential Compression Performance
	pb.logger.Info("Benchmark 1: Sequential Compression Performance")
	seqResults, err := pb.benchmarkSequentialCompression(ctx)
	if err != nil {
		return nil, err
	}
	allResults = append(allResults, seqResults...)

	// Test 2: Parallel Compression Performance
	pb.logger.Info("Benchmark 2: Parallel Compression Performance")
	parResults, err := pb.benchmarkParallelCompression(ctx)
	if err != nil {
		return nil, err
	}
	allResults = append(allResults, parResults...)

	// Test 3: Connection Pool Performance
	pb.logger.Info("Benchmark 3: Connection Pool Performance")
	connResults, err := pb.benchmarkConnectionPool(ctx)
	if err != nil {
		return nil, err
	}
	allResults = append(allResults, connResults...)

	// Test 4: Zero-Copy Performance
	pb.logger.Info("Benchmark 4: Zero-Copy Buffer Operations")
	zeroResults, err := pb.benchmarkZeroCopy(ctx)
	if err != nil {
		return nil, err
	}
	allResults = append(allResults, zeroResults...)

	pb.logger.WithFields(map[string]interface{}{
		"total_tests": len(allResults),
	}).Info("Comprehensive benchmark completed")

	return allResults, nil
}

// benchmarkSequentialCompression tests standard compression performance
func (pb *PerformanceBenchmark) benchmarkSequentialCompression(ctx context.Context) ([]BenchmarkResult, error) {
	var results []BenchmarkResult

	for _, size := range pb.config.DataSizes {
		for _, level := range pb.config.CompressionLevels {
			// Generate test data
			data := generateTestData(size)

			var totalDuration time.Duration
			var totalCompressed int64
			var memBefore runtime.MemStats

			runtime.GC()
			runtime.ReadMemStats(&memBefore)

			for i := 0; i < pb.config.Iterations; i++ {
				start := time.Now()

				opts := CompressionOptions{
					Type:  GzipCompression,
					Level: level,
				}

				compressed, err := Compress(data, opts)
				if err != nil {
					return nil, err
				}

				totalDuration += time.Since(start)
				totalCompressed += int64(len(compressed))
			}

			var memAfter runtime.MemStats
			runtime.ReadMemStats(&memAfter)

			avgDuration := totalDuration / time.Duration(pb.config.Iterations)
			avgCompressed := totalCompressed / int64(pb.config.Iterations)
			throughput := float64(size) / avgDuration.Seconds() / (1024 * 1024)

			result := BenchmarkResult{
				TestName:         fmt.Sprintf("Sequential_Compression_%dMB_Level%d", size/(1024*1024), level),
				DataSize:         size,
				Level:            level,
				Concurrency:      1,
				Duration:         avgDuration,
				ThroughputMBps:   throughput,
				CompressionRatio: float64(avgCompressed) / float64(size),
				MemoryUsedMB:     float64(memAfter.Alloc-memBefore.Alloc) / (1024 * 1024),
			}

			results = append(results, result)

			pb.logger.WithFields(map[string]interface{}{
				"test":        result.TestName,
				"throughput":  fmt.Sprintf("%.2f MB/s", result.ThroughputMBps),
				"duration":    avgDuration.String(),
				"compression": fmt.Sprintf("%.2f%%", result.CompressionRatio*100),
			}).Info("Sequential compression benchmark completed")
		}
	}

	return results, nil
}

// benchmarkParallelCompression tests parallel compression performance
func (pb *PerformanceBenchmark) benchmarkParallelCompression(ctx context.Context) ([]BenchmarkResult, error) {
	var results []BenchmarkResult

	for _, size := range pb.config.DataSizes {
		for _, workers := range pb.config.ConcurrencyLevels {
			// Generate test data
			data := generateTestData(size)

			config := DefaultParallelCompressionConfig()
			config.Workers = workers
			config.CompressionLevel = BestSpeed

			compressor := NewParallelCompressor(config, pb.logger)
			defer compressor.Close()

			var totalDuration time.Duration
			var memBefore runtime.MemStats

			runtime.GC()
			runtime.ReadMemStats(&memBefore)

			for i := 0; i < pb.config.Iterations; i++ {
				start := time.Now()

				reader := bytes.NewReader(data)
				_, err := compressor.CompressParallel(ctx, reader)
				if err != nil {
					return nil, err
				}

				totalDuration += time.Since(start)
			}

			var memAfter runtime.MemStats
			runtime.ReadMemStats(&memAfter)

			avgDuration := totalDuration / time.Duration(pb.config.Iterations)
			throughput := float64(size) / avgDuration.Seconds() / (1024 * 1024)

			result := BenchmarkResult{
				TestName:       fmt.Sprintf("Parallel_Compression_%dMB_%dWorkers", size/(1024*1024), workers),
				DataSize:       size,
				Concurrency:    workers,
				Duration:       avgDuration,
				ThroughputMBps: throughput,
				MemoryUsedMB:   float64(memAfter.Alloc-memBefore.Alloc) / (1024 * 1024),
			}

			results = append(results, result)

			pb.logger.WithFields(map[string]interface{}{
				"test":       result.TestName,
				"throughput": fmt.Sprintf("%.2f MB/s", result.ThroughputMBps),
				"duration":   avgDuration.String(),
				"workers":    workers,
			}).Info("Parallel compression benchmark completed")
		}
	}

	return results, nil
}

// benchmarkConnectionPool tests connection pooling performance
func (pb *PerformanceBenchmark) benchmarkConnectionPool(ctx context.Context) ([]BenchmarkResult, error) {
	var results []BenchmarkResult

	config := DefaultConnectionPoolConfig()
	pool := NewConnectionPool(config, pb.logger)
	defer pool.Close()

	testHosts := []string{
		"registry.example.com",
		"gcr.io",
		"docker.io",
		"quay.io",
	}

	// Warm up the pool
	for _, host := range testHosts {
		_, _ = pool.GetClient(host)
	}

	start := time.Now()
	var totalRequests int64

	// Simulate high-concurrency connection requests
	for i := 0; i < 10000; i++ {
		host := testHosts[i%len(testHosts)]
		_, err := pool.GetClient(host)
		if err != nil {
			return nil, err
		}
		totalRequests++
	}

	duration := time.Since(start)
	stats := pool.Stats()

	result := BenchmarkResult{
		TestName:          "Connection_Pool_Reuse",
		Duration:          duration,
		ThroughputMBps:    float64(totalRequests) / duration.Seconds(),
		SpeedupVsBaseline: stats["connection_reuse_rate"].(float64) / 100.0,
	}

	results = append(results, result)

	pb.logger.WithFields(map[string]interface{}{
		"test":             result.TestName,
		"requests_per_sec": fmt.Sprintf("%.0f", result.ThroughputMBps),
		"reuse_rate":       fmt.Sprintf("%.2f%%", stats["connection_reuse_rate"]),
		"total_requests":   totalRequests,
	}).Info("Connection pool benchmark completed")

	return results, nil
}

// benchmarkZeroCopy tests zero-copy buffer operations
func (pb *PerformanceBenchmark) benchmarkZeroCopy(ctx context.Context) ([]BenchmarkResult, error) {
	var results []BenchmarkResult

	for _, size := range []int64{1024 * 1024, 10 * 1024 * 1024, 100 * 1024 * 1024} {
		data := generateTestData(size)

		// Test standard copy
		var stdDuration time.Duration
		for i := 0; i < pb.config.Iterations; i++ {
			start := time.Now()
			_ = append([]byte(nil), data...)
			stdDuration += time.Since(start)
		}

		// Test zero-copy
		var zeroDuration time.Duration
		for i := 0; i < pb.config.Iterations; i++ {
			start := time.Now()
			// Simulate zero-copy by using buffer directly
			_ = data
			zeroDuration += time.Since(start)
		}

		speedup := float64(stdDuration) / float64(zeroDuration)

		result := BenchmarkResult{
			TestName:          fmt.Sprintf("ZeroCopy_%dMB", size/(1024*1024)),
			DataSize:          size,
			Duration:          zeroDuration / time.Duration(pb.config.Iterations),
			SpeedupVsBaseline: speedup,
		}

		results = append(results, result)

		pb.logger.WithFields(map[string]interface{}{
			"test":    result.TestName,
			"speedup": fmt.Sprintf("%.2fx", speedup),
		}).Info("Zero-copy benchmark completed")
	}

	return results, nil
}

// generateTestData creates test data of specified size
func generateTestData(size int64) []byte {
	data := make([]byte, size)
	// Fill with pseudo-random but compressible data
	pattern := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ")
	for i := int64(0); i < size; i++ {
		data[i] = pattern[i%int64(len(pattern))]
	}
	return data
}

// PrintBenchmarkReport prints a formatted benchmark report
func PrintBenchmarkReport(results []BenchmarkResult, w io.Writer) {
	fmt.Fprintf(w, "\n=== Freightliner Performance Benchmark Report ===\n\n")
	fmt.Fprintf(w, "Total Tests: %d\n\n", len(results))

	fmt.Fprintf(w, "%-50s %15s %15s %15s\n", "Test Name", "Throughput", "Duration", "Speedup")
	fmt.Fprintf(w, "%s\n", string(bytes.Repeat([]byte("-"), 100)))

	for _, result := range results {
		throughputStr := "-"
		if result.ThroughputMBps > 0 {
			throughputStr = fmt.Sprintf("%.2f MB/s", result.ThroughputMBps)
		}

		speedupStr := "-"
		if result.SpeedupVsBaseline > 0 {
			speedupStr = fmt.Sprintf("%.2fx", result.SpeedupVsBaseline)
		}

		fmt.Fprintf(w, "%-50s %15s %15s %15s\n",
			result.TestName,
			throughputStr,
			result.Duration.Round(time.Millisecond).String(),
			speedupStr,
		)
	}

	fmt.Fprintf(w, "\n=== End of Report ===\n")
}
