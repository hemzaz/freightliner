//go:build benchmark
// +build benchmark

package performance

import (
	"context"
	"crypto/sha256"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/replication"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

// BenchmarkReplicationThroughput measures throughput for different image sizes
func BenchmarkReplicationThroughput(b *testing.B) {
	sizes := []int64{
		1 * MB,   // Small image
		10 * MB,  // Medium-small image
		100 * MB, // Medium image
		500 * MB, // Large image
		1 * GB,   // Very large image
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("ImageSize_%dMB", size/MB), func(b *testing.B) {
			b.SetBytes(size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate image replication
				data := make([]byte, size)
				hash := sha256.New()
				hash.Write(data)
				_ = hash.Sum(nil)
			}

			// Report throughput
			mbps := float64(b.N*int(size)) / b.Elapsed().Seconds() / float64(MB)
			b.ReportMetric(mbps, "MB/s")
		})
	}
}

// BenchmarkWorkerPoolScaling measures scaling characteristics
func BenchmarkWorkerPoolScaling(b *testing.B) {
	workerCounts := []int{1, 2, 4, 8, 16, 32, 64}

	for _, count := range workerCounts {
		b.Run(fmt.Sprintf("Workers_%d", count), func(b *testing.B) {
			logger := log.NewBasicLogger(log.ErrorLevel)
			pool := replication.NewWorkerPool(count, logger)
			pool.Start()
			defer pool.Stop()

			b.ResetTimer()

			var wg sync.WaitGroup
			for i := 0; i < b.N; i++ {
				wg.Add(1)
				jobID := fmt.Sprintf("job-%d", i)

				err := pool.Submit(jobID, func(ctx context.Context) error {
					defer wg.Done()
					// Simulate work
					time.Sleep(10 * time.Millisecond)
					return nil
				})

				if err != nil {
					b.Fatalf("Failed to submit job: %v", err)
				}
			}

			wg.Wait()

			// Report jobs per second
			jobsPerSec := float64(b.N) / b.Elapsed().Seconds()
			b.ReportMetric(jobsPerSec, "jobs/sec")
		})
	}
}

// BenchmarkWorkerPoolParallelism tests parallel job execution
func BenchmarkWorkerPoolParallelism(b *testing.B) {
	workerCount := runtime.NumCPU()
	logger := log.NewBasicLogger(log.ErrorLevel)
	pool := replication.NewWorkerPool(workerCount, logger)
	pool.Start()
	defer pool.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			jobID := fmt.Sprintf("parallel-job-%d", i)
			i++

			done := make(chan struct{})
			err := pool.Submit(jobID, func(ctx context.Context) error {
				defer close(done)
				// Simulate CPU-intensive work
				data := make([]byte, 1*MB)
				hash := sha256.New()
				hash.Write(data)
				_ = hash.Sum(nil)
				return nil
			})

			if err != nil {
				b.Fatalf("Failed to submit job: %v", err)
			}

			// Wait for job completion
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				b.Fatal("Job timed out")
			}
		}
	})
}

// BenchmarkMemoryUsage measures memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	dataSizes := []int64{
		1 * MB,
		10 * MB,
		100 * MB,
	}

	for _, size := range dataSizes {
		b.Run(fmt.Sprintf("DataSize_%dMB", size/MB), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Allocate and process data
				data := make([]byte, size)
				hash := sha256.New()
				hash.Write(data)
				_ = hash.Sum(nil)
			}

			// Memory metrics automatically reported
		})
	}
}

// BenchmarkCPUUtilization measures CPU utilization
func BenchmarkCPUUtilization(b *testing.B) {
	workerCount := runtime.NumCPU()
	logger := log.NewBasicLogger(log.ErrorLevel)
	pool := replication.NewWorkerPool(workerCount, logger)
	pool.Start()
	defer pool.Stop()

	b.ResetTimer()

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		jobID := fmt.Sprintf("cpu-job-%d", i)

		err := pool.Submit(jobID, func(ctx context.Context) error {
			defer wg.Done()
			// CPU-intensive work
			hash := sha256.New()
			data := make([]byte, 1*MB)
			for j := 0; j < 10; j++ {
				hash.Write(data)
			}
			_ = hash.Sum(nil)
			return nil
		})

		if err != nil {
			b.Fatalf("Failed to submit job: %v", err)
		}
	}

	wg.Wait()
}

// BenchmarkConcurrentOperations measures concurrent operation overhead
func BenchmarkConcurrentOperations(b *testing.B) {
	concurrencyLevels := []int{1, 10, 50, 100, 500, 1000}

	for _, level := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", level), func(b *testing.B) {
			logger := log.NewBasicLogger(log.ErrorLevel)
			pool := replication.NewWorkerPool(runtime.NumCPU()*2, logger)
			pool.Start()
			defer pool.Stop()

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				for j := 0; j < level; j++ {
					wg.Add(1)
					jobID := fmt.Sprintf("concurrent-job-%d-%d", i, j)

					err := pool.Submit(jobID, func(ctx context.Context) error {
						defer wg.Done()
						// Light work
						time.Sleep(1 * time.Millisecond)
						return nil
					})

					if err != nil {
						b.Fatalf("Failed to submit job: %v", err)
					}
				}
				wg.Wait()
			}

			opsPerSec := float64(b.N*level) / b.Elapsed().Seconds()
			b.ReportMetric(opsPerSec, "ops/sec")
		})
	}
}

// BenchmarkChannelThroughput measures channel throughput
func BenchmarkChannelThroughput(b *testing.B) {
	bufferSizes := []int{0, 1, 10, 100, 1000}

	for _, size := range bufferSizes {
		b.Run(fmt.Sprintf("BufferSize_%d", size), func(b *testing.B) {
			ch := make(chan int, size)
			var sent, received atomic.Int64

			b.ResetTimer()

			go func() {
				for i := 0; i < b.N; i++ {
					ch <- i
					sent.Add(1)
				}
				close(ch)
			}()

			for range ch {
				received.Add(1)
			}

			msgsPerSec := float64(b.N) / b.Elapsed().Seconds()
			b.ReportMetric(msgsPerSec, "msgs/sec")
		})
	}
}

// BenchmarkMutexContention measures mutex contention overhead
func BenchmarkMutexContention(b *testing.B) {
	goroutineCounts := []int{1, 2, 4, 8, 16, 32}

	for _, count := range goroutineCounts {
		b.Run(fmt.Sprintf("Goroutines_%d", count), func(b *testing.B) {
			var mu sync.Mutex
			counter := 0

			b.ResetTimer()

			var wg sync.WaitGroup
			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < b.N/count; j++ {
						mu.Lock()
						counter++
						mu.Unlock()
					}
				}()
			}

			wg.Wait()

			locksPerSec := float64(b.N) / b.Elapsed().Seconds()
			b.ReportMetric(locksPerSec, "locks/sec")
		})
	}
}

// BenchmarkAtomicOperations measures atomic operation performance
func BenchmarkAtomicOperations(b *testing.B) {
	operations := []struct {
		name string
		fn   func(*atomic.Int64)
	}{
		{
			name: "Load",
			fn:   func(v *atomic.Int64) { _ = v.Load() },
		},
		{
			name: "Store",
			fn:   func(v *atomic.Int64) { v.Store(42) },
		},
		{
			name: "Add",
			fn:   func(v *atomic.Int64) { v.Add(1) },
		},
		{
			name: "CompareAndSwap",
			fn:   func(v *atomic.Int64) { v.CompareAndSwap(0, 1) },
		},
	}

	for _, op := range operations {
		b.Run(op.name, func(b *testing.B) {
			var value atomic.Int64

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				op.fn(&value)
			}

			opsPerSec := float64(b.N) / b.Elapsed().Seconds()
			b.ReportMetric(opsPerSec, "ops/sec")
		})
	}
}

// BenchmarkSchedulerPerformance measures scheduler performance
func BenchmarkSchedulerPerformance(b *testing.B) {
	jobCounts := []int{10, 50, 100, 500, 1000}

	for _, count := range jobCounts {
		b.Run(fmt.Sprintf("Jobs_%d", count), func(b *testing.B) {
			logger := log.NewBasicLogger(log.ErrorLevel)
			pool := replication.NewWorkerPool(runtime.NumCPU(), logger)
			pool.Start()
			defer pool.Stop()

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				for j := 0; j < count; j++ {
					wg.Add(1)
					jobID := fmt.Sprintf("sched-job-%d-%d", i, j)

					err := pool.Submit(jobID, func(ctx context.Context) error {
						defer wg.Done()
						// Minimal work
						time.Sleep(100 * time.Microsecond)
						return nil
					})

					if err != nil {
						b.Fatalf("Failed to submit job: %v", err)
					}
				}
				wg.Wait()
			}

			jobsPerSec := float64(b.N*count) / b.Elapsed().Seconds()
			b.ReportMetric(jobsPerSec, "jobs/sec")
		})
	}
}

// BenchmarkContextCancellation measures context cancellation overhead
func BenchmarkContextCancellation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			<-ctx.Done()
		}()

		cancel()
	}

	cancellationsPerSec := float64(b.N) / b.Elapsed().Seconds()
	b.ReportMetric(cancellationsPerSec, "cancel/sec")
}

// BenchmarkGoroutineCreation measures goroutine creation overhead
func BenchmarkGoroutineCreation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		done := make(chan struct{})
		go func() {
			close(done)
		}()
		<-done
	}

	goroutinesPerSec := float64(b.N) / b.Elapsed().Seconds()
	b.ReportMetric(goroutinesPerSec, "goroutines/sec")
}

// BenchmarkWaitGroupOverhead measures sync.WaitGroup overhead
func BenchmarkWaitGroupOverhead(b *testing.B) {
	goroutineCounts := []int{1, 10, 100, 1000}

	for _, count := range goroutineCounts {
		b.Run(fmt.Sprintf("Goroutines_%d", count), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				for j := 0; j < count; j++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						// Minimal work
						runtime.Gosched()
					}()
				}
				wg.Wait()
			}

			operationsPerSec := float64(b.N*count) / b.Elapsed().Seconds()
			b.ReportMetric(operationsPerSec, "ops/sec")
		})
	}
}
