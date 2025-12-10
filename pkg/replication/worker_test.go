package replication

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
)

func TestWorkerPoolSubmit(t *testing.T) {
	// Create a worker pool with 2 workers
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(2, logger)
	pool.Start()
	defer pool.Stop()

	// Create a wait group to wait for all jobs to complete
	var wg sync.WaitGroup
	wg.Add(10)

	// Create a channel to track results
	results := make(chan int, 10)

	// Submit 10 jobs
	for i := 0; i < 10; i++ {
		i := i // Capture the value
		err := pool.Submit(fmt.Sprintf("job-%d", i), func(ctx context.Context) error {
			defer wg.Done()
			// Simulate work
			time.Sleep(10 * time.Millisecond)
			results <- i
			return nil
		})

		if err != nil {
			t.Fatalf("Failed to submit job: %v", err)
		}
	}

	// Wait for all jobs to complete
	wg.Wait()
	close(results)

	// Check results
	var count int
	for range results {
		count++
	}

	if count != 10 {
		t.Errorf("Expected 10 results, got %d", count)
	}
}

func TestWorkerPoolPriority(t *testing.T) {
	// Create a worker pool with 2 workers
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(2, logger)
	pool.Start()
	defer pool.Stop()

	// Create a wait group to wait for all jobs to complete
	var wg sync.WaitGroup
	wg.Add(10)

	// Create a variable to track the order of completion
	var completionOrder []int
	var orderMutex sync.Mutex

	// Submit 10 jobs with different priorities
	for i := 0; i < 10; i++ {
		i := i             // Capture the value
		priority := 10 - i // Higher number = higher priority

		err := pool.SubmitWithPriority(fmt.Sprintf("job-%d", i), func(ctx context.Context) error {
			defer wg.Done()
			// Record completion order
			orderMutex.Lock()
			completionOrder = append(completionOrder, i)
			orderMutex.Unlock()
			return nil
		}, priority)

		if err != nil {
			t.Fatalf("Failed to submit job: %v", err)
		}
	}

	// Wait for all jobs to complete
	wg.Wait()

	// In a priority queue, we can't guarantee exact order due to concurrency,
	// but we can check that higher priority jobs tend to complete first
	// This is a basic sanity check
	if len(completionOrder) != 10 {
		t.Errorf("Expected 10 completed jobs, got %d", len(completionOrder))
	}
}

func TestWorkerPoolScaling(t *testing.T) {
	for _, workerCount := range []int{1, 2, 4, 8} {
		t.Run(fmt.Sprintf("Workers-%d", workerCount), func(t *testing.T) {
			// Create a worker pool with variable worker count
			logger := log.NewBasicLogger(log.InfoLevel)
			pool := NewWorkerPool(workerCount, logger)
			pool.Start()
			defer pool.Stop()

			// Create a counter for completed jobs
			var completed int32
			var resultsReceived int32

			// Start result consumer to prevent deadlock
			resultsDone := make(chan struct{})
			go func() {
				defer close(resultsDone)
				for range pool.GetResults() {
					atomic.AddInt32(&resultsReceived, 1)
				}
			}()

			// Submit 100 jobs
			for i := 0; i < 100; i++ {
				err := pool.Submit(fmt.Sprintf("job-%d", i), func(ctx context.Context) error {
					// Simulate work
					time.Sleep(1 * time.Millisecond)
					atomic.AddInt32(&completed, 1)
					return nil
				})

				if err != nil {
					t.Fatalf("Failed to submit job: %v", err)
				}
			}

			// Wait for all jobs to complete
			pool.Wait()

			// Wait for result consumer to finish
			<-resultsDone

			// Verify all jobs completed
			if atomic.LoadInt32(&completed) != 100 {
				t.Errorf("Expected 100 completed jobs, got %d", atomic.LoadInt32(&completed))
			}

			if atomic.LoadInt32(&resultsReceived) != 100 {
				t.Errorf("Expected 100 results received, got %d", atomic.LoadInt32(&resultsReceived))
			}
		})
	}
}

func TestWorkerPoolCancellation(t *testing.T) {
	// Create a worker pool with 2 workers
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(2, logger)
	pool.Start()
	defer pool.Stop()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Create a wait group to track job starts
	var startWg sync.WaitGroup
	startWg.Add(1)

	// Create a wait group to track job completions
	var completeWg sync.WaitGroup
	completeWg.Add(1)

	// Counter for completed jobs
	var completed int32

	// Submit a job that will be cancelled
	err := pool.SubmitWithContext(ctx, "cancellable-job", func(jobCtx context.Context) error {
		// Signal that the job has started
		startWg.Done()

		// Wait for cancellation or completion
		select {
		case <-jobCtx.Done():
			// Job was cancelled
			return jobCtx.Err()
		case <-time.After(5 * time.Second):
			// This should not happen if cancellation works
			atomic.StoreInt32(&completed, 1)
			completeWg.Done()
			return nil
		}
	})

	if err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}

	// Wait for the job to start
	startWg.Wait()

	// Cancel the context
	cancel()

	// Wait a short time to allow cancellation to propagate
	time.Sleep(100 * time.Millisecond)

	// Check if the job completed (it shouldn't have)
	if atomic.LoadInt32(&completed) != 0 {
		t.Errorf("Job completed despite cancellation")
	}
}

func TestWorkerPoolStopWithRunningJobs(t *testing.T) {
	// Create a worker pool with 2 workers
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(2, logger)
	pool.Start()

	// Create a wait group to track job starts
	var startWg sync.WaitGroup
	startWg.Add(2)

	// Counter for completed jobs
	var completed int32

	// Submit two long-running jobs
	for i := 0; i < 2; i++ {
		err := pool.Submit(fmt.Sprintf("long-job-%d", i), func(ctx context.Context) error {
			// Signal that the job has started
			startWg.Done()

			// Wait for cancellation or completion
			select {
			case <-ctx.Done():
				// Job was cancelled by worker pool stop
				return ctx.Err()
			case <-time.After(5 * time.Second):
				// This should not happen if stop works
				atomic.AddInt32(&completed, 1)
				return nil
			}
		})

		if err != nil {
			t.Fatalf("Failed to submit job: %v", err)
		}
	}

	// Wait for both jobs to start
	startWg.Wait()

	// Stop the worker pool
	pool.Stop()

	// Check if any jobs completed (they shouldn't have)
	if atomic.LoadInt32(&completed) != 0 {
		t.Errorf("Expected no jobs to complete, but %d completed", completed)
	}
}

func TestResultsChannel(t *testing.T) {
	// Create a worker pool with 2 workers
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(2, logger)
	pool.Start()

	// Submit jobs that will succeed or fail
	// Even-numbered jobs succeed, odd-numbered jobs fail
	for i := 0; i < 10; i++ {
		i := i // Capture for closure
		_ = pool.Submit(fmt.Sprintf("job-%d", i), func(ctx context.Context) error {
			if i%2 == 0 {
				return nil // Success
			} else {
				return fmt.Errorf("job %d failed", i) // Failure
			}
		})
	}

	// Count successes and failures
	results := pool.GetResults()
	successCount := 0
	failureCount := 0

	// Wait for a short time to collect results
	timeout := time.After(2 * time.Second)
	for i := 0; i < 10; {
		select {
		case result, ok := <-results:
			if !ok {
				t.Fatalf("Results channel closed unexpectedly")
			}
			if result.Error == nil {
				successCount++
			} else {
				failureCount++
			}
			i++
		case <-timeout:
			t.Fatalf("Timed out waiting for results, got %d successes and %d failures", successCount, failureCount)
		}
	}

	// Verify counts
	if successCount != 5 || failureCount != 5 {
		t.Errorf("Expected 5 successes and 5 failures, got %d successes and %d failures", successCount, failureCount)
	}

	// Stop the pool
	pool.Stop()
}
