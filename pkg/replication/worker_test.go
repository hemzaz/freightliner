package replication

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPoolSubmit(t *testing.T) {
	// Create a worker pool with 2 workers
	pool := NewWorkerPool(2)

	// Variables to track task execution
	var completed int32
	var executions sync.Map

	// Create a task that increments the counter and records execution
	task := func() error {
		taskID := atomic.AddInt32(&completed, 1)
		executions.Store(taskID, true)
		return nil
	}

	// Submit 5 tasks
	for i := 0; i < 5; i++ {
		err := pool.Submit(task)
		if err != nil {
			t.Errorf("Failed to submit task: %v", err)
		}
	}

	// Wait for all tasks to complete
	pool.Wait()

	// Check if all tasks were executed
	if atomic.LoadInt32(&completed) != 5 {
		t.Errorf("Expected 5 tasks to be completed, got %d", completed)
	}

	// Check if each task was executed exactly once
	count := 0
	executions.Range(func(_, _ interface{}) bool {
		count++
		return true
	})

	if count != 5 {
		t.Errorf("Expected 5 unique task executions, got %d", count)
	}
}

func TestWorkerPoolStop(t *testing.T) {
	// Create a worker pool with 2 workers
	pool := NewWorkerPool(2)

	// Variables to track task execution
	var completed int32

	// Create a task that sleeps briefly and then increments the counter
	task := func() error {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&completed, 1)
		return nil
	}

	// Submit 10 tasks (more than workers to ensure queue fills)
	for i := 0; i < 10; i++ {
		err := pool.Submit(task)
		if err != nil {
			t.Errorf("Failed to submit task: %v", err)
		}
	}

	// Stop the pool immediately after submitting
	pool.Stop()

	// Wait a short time to allow some tasks to complete
	time.Sleep(200 * time.Millisecond)

	// Check that some tasks completed but not all
	completedTasks := atomic.LoadInt32(&completed)
	if completedTasks == 0 {
		t.Error("No tasks completed, expected at least some to complete")
	}

	if completedTasks == 10 {
		t.Error("All tasks completed, expected some to be cancelled")
	}

	// Try to submit another task after stop
	err := pool.Submit(task)
	if err == nil {
		t.Error("Expected error when submitting to stopped pool, got nil")
	}
}

func TestWorkerPoolWait(t *testing.T) {
	// Create a worker pool with 3 workers
	pool := NewWorkerPool(3)

	// Variables to track task execution
	var completed int32

	// Create tasks with different durations
	tasks := []struct {
		duration time.Duration
	}{
		{50 * time.Millisecond},
		{100 * time.Millisecond},
		{150 * time.Millisecond},
		{200 * time.Millisecond},
		{250 * time.Millisecond},
	}

	// Submit all tasks
	for _, task := range tasks {
		duration := task.duration
		err := pool.Submit(func() error {
			time.Sleep(duration)
			atomic.AddInt32(&completed, 1)
			return nil
		})

		if err != nil {
			t.Errorf("Failed to submit task: %v", err)
		}
	}

	// Record start time
	startWait := time.Now()

	// Wait for all tasks to complete
	pool.Wait()

	// Check wait duration
	waitDuration := time.Since(startWait)

	// Wait should take at least the duration of the longest task
	if waitDuration < 250*time.Millisecond {
		t.Errorf("Wait returned too quickly: %v", waitDuration)
	}

	// Check if all tasks were executed
	if atomic.LoadInt32(&completed) != 5 {
		t.Errorf("Expected 5 tasks to be completed, got %d", completed)
	}
}

func TestWorkerPoolErrorHandling(t *testing.T) {
	// Create a worker pool with 2 workers
	pool := NewWorkerPool(2)

	// Variables to track task execution
	var successCount int32
	var errorCount int32

	// Create tasks that sometimes return errors
	for i := 0; i < 10; i++ {
		i := i // Capture loop variable
		err := pool.Submit(func() error {
			if i%2 == 0 {
				atomic.AddInt32(&successCount, 1)
				return nil
			} else {
				atomic.AddInt32(&errorCount, 1)
				return errors.New("test error")
			}
		})

		if err != nil {
			t.Errorf("Failed to submit task: %v", err)
		}
	}

	// Wait for all tasks to complete
	pool.Wait()

	// Check counts
	if atomic.LoadInt32(&successCount) != 5 {
		t.Errorf("Expected 5 successful tasks, got %d", successCount)
	}

	if atomic.LoadInt32(&errorCount) != 5 {
		t.Errorf("Expected 5 error tasks, got %d", errorCount)
	}
}

func TestWorkerPoolConcurrency(t *testing.T) {
	// Test with different worker counts to ensure concurrency works
	workerCounts := []int{1, 2, 4, 8}

	for _, workerCount := range workerCounts {
		t.Run("WorkerCount_"+string(rune('0'+workerCount)), func(t *testing.T) {
			// Create a worker pool with specified number of workers
			pool := NewWorkerPool(workerCount)

			// Create a mutex-protected map to track concurrent execution
			var mu sync.Mutex
			executing := make(map[int]bool)
			maxConcurrent := 0

			// Submit tasks that track their concurrent execution
			taskCount := workerCount * 3
			var wg sync.WaitGroup
			wg.Add(taskCount)

			for i := 0; i < taskCount; i++ {
				i := i
				err := pool.Submit(func() error {
					// Mark this task as executing
					mu.Lock()
					executing[i] = true
					currentConcurrent := len(executing)
					if currentConcurrent > maxConcurrent {
						maxConcurrent = currentConcurrent
					}
					mu.Unlock()

					// Simulate work
					time.Sleep(50 * time.Millisecond)

					// Mark this task as done
					mu.Lock()
					delete(executing, i)
					mu.Unlock()

					wg.Done()
					return nil
				})

				if err != nil {
					t.Errorf("Failed to submit task: %v", err)
				}
			}

			// Wait for all tasks to complete
			wg.Wait()
			pool.Wait()

			// Check maximum concurrency
			if workerCount == 1 {
				// For single worker, max concurrent should be exactly 1
				if maxConcurrent != 1 {
					t.Errorf("Expected max concurrency of 1 for single worker, got %d", maxConcurrent)
				}
			} else {
				// For multiple workers, max concurrent should be at least 2 and at most workerCount
				if maxConcurrent < 2 || maxConcurrent > workerCount {
					t.Errorf("Expected max concurrency between 2 and %d, got %d", workerCount, maxConcurrent)
				}
			}
		})
	}
}

func TestWorkerPoolContextCancellation(t *testing.T) {
	// Create a worker pool with context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewWorkerPoolWithContext(ctx, 2)

	// Variables to track task execution
	var taskStarted sync.WaitGroup
	var taskCompleted int32

	// Create a task that waits for cancellation
	taskStarted.Add(1)
	err := pool.Submit(func() error {
		taskStarted.Done()
		// This task will block until context is cancelled or a long timeout
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second): // Long timeout as safeguard
			atomic.AddInt32(&taskCompleted, 1)
			return nil
		}
	})

	if err != nil {
		t.Errorf("Failed to submit task: %v", err)
	}

	// Wait for task to start
	taskStarted.Wait()

	// Cancel the context
	cancel()

	// Wait for pool to finish
	pool.Wait()

	// Check that task did not complete normally
	if atomic.LoadInt32(&taskCompleted) > 0 {
		t.Error("Task should not have completed normally after context cancellation")
	}
}

func TestWorkerPoolPanic(t *testing.T) {
	// Create a worker pool
	pool := NewWorkerPool(2)

	// Submit a task that panics
	err := pool.Submit(func() error {
		panic("test panic")
	})

	if err != nil {
		t.Errorf("Failed to submit task: %v", err)
	}

	// Submit a normal task
	var taskCompleted int32
	err = pool.Submit(func() error {
		atomic.AddInt32(&taskCompleted, 1)
		return nil
	})

	if err != nil {
		t.Errorf("Failed to submit task: %v", err)
	}

	// Wait for pool to finish
	pool.Wait()

	// Check that the normal task completed
	if atomic.LoadInt32(&taskCompleted) != 1 {
		t.Error("Normal task should have completed despite panic in other task")
	}
}
