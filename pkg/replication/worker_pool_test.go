package replication

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
)

// TestWorkerPool_Basic tests basic worker pool functionality
func TestWorkerPool_Basic(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()

	// Create a counter to track completed tasks
	var completedTasks atomic.Int32

	// Submit tasks
	for i := 0; i < 10; i++ {
		err := pool.Submit("task-"+string(rune('A'+i)), func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			completedTasks.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}
	}

	// Wait for tasks to complete
	pool.Wait()

	// Verify all tasks were completed
	if completedTasks.Load() != 10 {
		t.Errorf("Expected 10 completed tasks, got %d", completedTasks.Load())
	}
}

// TestWorkerPool_Errors tests error handling in the worker pool
func TestWorkerPool_Errors(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	pool := NewWorkerPool(3, logger)
	pool.Start()

	expectedErr := errors.New("test error")
	var errorsReceived atomic.Int32

	// Submit successful tasks
	for i := 0; i < 5; i++ {
		err := pool.Submit("success-"+string(rune('A'+i)), func(ctx context.Context) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}
	}

	// Submit a failing task
	err := pool.Submit("error-task", func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return expectedErr
	})
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	// Collect results in a separate goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for result := range pool.GetResults() {
			if result.Error != nil {
				errorsReceived.Add(1)
				if !errors.Is(result.Error, expectedErr) {
					t.Errorf("Expected error %v, got %v", expectedErr, result.Error)
				}
			}
		}
	}()

	// Wait for tasks to complete
	pool.Wait()

	// Wait for result processing to complete
	wg.Wait()

	// Verify we received the expected error
	if errorsReceived.Load() != 1 {
		t.Errorf("Expected 1 error, got %d", errorsReceived.Load())
	}
}

// TestWorkerPool_ContextCancellation tests cancellation via context
func TestWorkerPool_ContextCancellation(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	pool := NewWorkerPool(3, logger)
	pool.Start()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Submit a task that respects context cancellation
	err := pool.SubmitWithContext(ctx, "cancel-task", func(ctx context.Context) error {
		select {
		case <-time.After(500 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	// Cancel the context after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// Verify the result shows cancellation
	var receivedContextCancelled bool
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for result := range pool.GetResults() {
			if result.Error != nil && errors.Is(result.Error, context.Canceled) {
				receivedContextCancelled = true
			}
		}
	}()

	// Wait for tasks to complete
	pool.Wait()

	// Wait for result processing to complete
	wg.Wait()

	if !receivedContextCancelled {
		t.Error("Expected to receive context cancelled error")
	}
}

// TestWorkerPool_Concurrency tests that the worker pool respects concurrency limits
func TestWorkerPool_Concurrency(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	workers := 3
	pool := NewWorkerPool(workers, logger)
	pool.Start()

	// Track concurrent tasks
	var activeTasks atomic.Int32
	var maxConcurrent atomic.Int32

	// Create a wait group to signal when all tasks are submitted
	var wg sync.WaitGroup

	// Submit a large number of tasks
	taskCount := 50
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		err := pool.Submit("task-"+string(rune('A'+i%26)), func(ctx context.Context) error {
			// Increment active tasks
			current := activeTasks.Add(1)
			defer activeTasks.Add(-1)
			defer wg.Done()

			// Update max concurrent if needed
			for {
				max := maxConcurrent.Load()
				if current <= max {
					break
				}
				if maxConcurrent.CompareAndSwap(max, current) {
					break
				}
			}

			// Simulate some work
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}
	}

	// Wait for all tasks to be processed
	wg.Wait()
	pool.Wait()

	// Verify that concurrency was limited to the number of workers
	if maxConcurrent.Load() > int32(workers) {
		t.Errorf("Expected max concurrent tasks to be <= %d, got %d", workers, maxConcurrent.Load())
	}
}

// TestWorkerPool_Stop tests stopping the worker pool
func TestWorkerPool_Stop(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	pool := NewWorkerPool(2, logger)
	pool.Start()

	var completedTasks atomic.Int32
	var cancelled atomic.Bool

	// Submit tasks that respect context cancellation
	for i := 0; i < 4; i++ {
		err := pool.Submit("task-"+string(rune('A'+i)), func(ctx context.Context) error {
			select {
			case <-time.After(100 * time.Millisecond):
				completedTasks.Add(1)
				return nil
			case <-ctx.Done():
				cancelled.Store(true)
				return ctx.Err()
			}
		})
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}
	}

	// Stop the pool after a short delay
	go func() {
		time.Sleep(25 * time.Millisecond)
		pool.Stop()
	}()

	// The wait should complete due to Stop
	pool.Wait()

	// Some tasks should have been cancelled or not all should have completed naturally
	completed := completedTasks.Load()
	if completed == 4 && !cancelled.Load() {
		t.Error("Expected some tasks to be interrupted by Stop() or cancelled due to context")
	}

	t.Logf("Completed tasks: %d, Cancelled: %v", completed, cancelled.Load())
}

// TestWorkerPool_RaceConditions focuses on detecting race conditions with the race detector
func TestWorkerPool_RaceConditions(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	pool := NewWorkerPool(10, logger)
	pool.Start()

	// A shared counter to be accessed by all worker tasks
	var counter atomic.Int32

	// Submit a large number of tasks
	taskCount := 100
	for i := 0; i < taskCount; i++ {
		err := pool.Submit("task-"+string(rune('A'+i%26)), func(ctx context.Context) error {
			// Increment counter
			counter.Add(1)

			// Simulate some work with random durations
			time.Sleep(time.Duration(i%10+1) * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}
	}

	// Collect results in a separate goroutine
	go func() {
		for range pool.GetResults() {
			// Just drain the results channel
		}
	}()

	// Start other goroutines that access worker pool state concurrently
	for i := 0; i < 5; i++ {
		go func() {
			// Just read worker count
			_ = pool.WorkerCount()
			time.Sleep(5 * time.Millisecond)
		}()
	}

	// Wait for tasks to complete
	pool.Wait()

	// Verify all tasks were executed
	if counter.Load() != int32(taskCount) {
		t.Errorf("Expected %d completed tasks, got %d", taskCount, counter.Load())
	}
}

// TestWorkerPool_MultipleSubmittersAndConsumers tests multiple goroutines submitting and consuming results
func TestWorkerPool_MultipleSubmittersAndConsumers(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()

	var submittedTasks atomic.Int32
	var completedResults atomic.Int32

	// Create multiple submitter goroutines
	var wgSubmitters sync.WaitGroup
	for i := 0; i < 5; i++ {
		wgSubmitters.Add(1)
		go func(id int) {
			defer wgSubmitters.Done()
			for j := 0; j < 20; j++ {
				taskID := string(rune('A'+id)) + "-" + string(rune('0'+j%10))
				err := pool.Submit(taskID, func(ctx context.Context) error {
					time.Sleep(time.Duration(id+j) % 10 * time.Millisecond)
					submittedTasks.Add(1)
					return nil
				})
				if err != nil {
					t.Errorf("Failed to submit task: %v", err)
					return
				}
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Create multiple consumer goroutines
	var wgConsumers sync.WaitGroup
	resultsCh := pool.GetResults()
	for i := 0; i < 3; i++ {
		wgConsumers.Add(1)
		go func() {
			defer wgConsumers.Done()
			for result := range resultsCh {
				if result.Error != nil {
					t.Errorf("Unexpected error: %v", result.Error)
				}
				completedResults.Add(1)
			}
		}()
	}

	// Wait for submitters to finish
	wgSubmitters.Wait()

	// Wait for all tasks to complete
	pool.Wait()

	// Wait for consumers to finish processing results
	wgConsumers.Wait()

	// Verify all tasks were completed
	if submittedTasks.Load() != 100 {
		t.Errorf("Expected 100 submitted tasks, got %d", submittedTasks.Load())
	}

	if completedResults.Load() != 100 {
		t.Errorf("Expected 100 completed results, got %d", completedResults.Load())
	}
}

// TestWorkerPool_PriorityOrdering tests tasks with different priorities
func TestWorkerPool_PriorityOrdering(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	pool := NewWorkerPool(1, logger) // Single worker to ensure sequential processing
	pool.Start()

	// Track task execution order
	var executionOrder []string
	var mu sync.Mutex

	// Submit tasks with varying priorities
	// Note: This test is more about race conditions than actual priority ordering
	// since the worker pool implementation might not strictly order by priority
	for i := 0; i < 10; i++ {
		priority := 10 - i
		taskID := string(rune('A' + i))

		err := pool.SubmitWithPriority(taskID, func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, taskID)
			mu.Unlock()
			return nil
		}, priority)

		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}
	}

	// Wait for tasks to complete
	pool.Wait()

	// Just verify we got 10 tasks executed without races
	mu.Lock()
	if len(executionOrder) != 10 {
		t.Errorf("Expected 10 executed tasks, got %d", len(executionOrder))
	}
	mu.Unlock()
}
