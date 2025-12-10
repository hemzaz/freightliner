package util

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLimitedErrGroup_NoLimit(t *testing.T) {
	// Create a group with no concurrency limit
	ctx := context.Background()
	g := NewLimitedErrGroup(ctx, 0)

	// Number of tasks to run
	const tasks = 50
	counter := atomic.Int32{}

	// Add a bunch of tasks
	for i := 0; i < tasks; i++ {
		g.Go(func() error {
			// Increment the counter atomically
			counter.Add(1)
			return nil
		})
	}

	// Wait for all tasks to complete
	err := g.Wait()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// All tasks should have completed
	if counter.Load() != tasks {
		t.Errorf("Expected %d tasks to run, got %d", tasks, counter.Load())
	}
}

func TestLimitedErrGroup_WithLimit(t *testing.T) {
	// Create a group with concurrency limit of 5
	ctx := context.Background()
	g := NewLimitedErrGroup(ctx, 5)

	// Number of tasks to run
	const tasks = 100
	counter := atomic.Int32{}
	concurrent := atomic.Int32{}
	maxConcurrent := atomic.Int32{}

	// Add a bunch of tasks
	for i := 0; i < tasks; i++ {
		g.Go(func() error {
			// Track concurrency level
			current := concurrent.Add(1)
			// Update max concurrent tasks if current is greater
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
			time.Sleep(1 * time.Millisecond)
			concurrent.Add(-1) // Decrement concurrency counter
			counter.Add(1)     // Increment completed counter
			return nil
		})
	}

	// Wait for all tasks to complete
	err := g.Wait()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// All tasks should have completed
	if counter.Load() != tasks {
		t.Errorf("Expected %d tasks to run, got %d", tasks, counter.Load())
	}

	// Maximum concurrency should not exceed the limit
	if maxConcurrent.Load() > 5 {
		t.Errorf("Expected max concurrency of 5, got %d", maxConcurrent.Load())
	}
}

func TestLimitedErrGroup_ContextCancellation(t *testing.T) {
	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	g := NewLimitedErrGroup(ctx, 5)

	// Create a notification channel
	done := make(chan struct{})

	// Task will block until cancelled
	g.Go(func() error {
		// Signal that we're inside the goroutine
		close(done)
		<-ctx.Done() // Block until context is cancelled
		return ctx.Err()
	})

	// Wait for goroutine to start
	<-done

	// Cancel the context
	cancel()

	// Wait for all tasks to complete
	err := g.Wait()
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context cancelled error, got %v", err)
	}
}

func TestLimitedErrGroup_ErrorPropagation(t *testing.T) {
	ctx := context.Background()
	g := NewLimitedErrGroup(ctx, 10)

	expectedErr := errors.New("test error")

	// Add tasks that succeed
	for i := 0; i < 5; i++ {
		g.Go(func() error {
			time.Sleep(1 * time.Millisecond)
			return nil
		})
	}

	// Add a task that fails
	g.Go(func() error {
		time.Sleep(2 * time.Millisecond)
		return expectedErr
	})

	// Add more tasks that succeed
	for i := 0; i < 5; i++ {
		g.Go(func() error {
			time.Sleep(3 * time.Millisecond)
			return nil
		})
	}

	// Wait for all tasks to complete
	err := g.Wait()
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected %v, got %v", expectedErr, err)
	}
}

func TestResults_ThreadSafety(t *testing.T) {
	// Create a new Results collector
	results := NewResults()

	// Number of goroutines and operations per goroutine
	const goroutines = 10
	const opsPerGoroutine = 1000

	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // Double because we're testing Add and AddMetric separately

	// Test Add method concurrently
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				results.Add(fmt.Sprintf("item-%d-%d", id, j))
			}
		}(i)
	}

	// Test AddMetric method concurrently
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				// Use different metric names to test map operations
				metricName := fmt.Sprintf("metric-%d", j%10)
				results.AddMetric(metricName, 1)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify results
	items := results.GetItems()
	if len(items) != goroutines*opsPerGoroutine {
		t.Errorf("Expected %d items, got %d", goroutines*opsPerGoroutine, len(items))
	}

	// Verify metrics - each metric should have been incremented by (goroutines * opsPerGoroutine / 10)
	metrics := results.GetAllMetrics()
	expectedMetricCount := goroutines * opsPerGoroutine / 10
	for i := 0; i < 10; i++ {
		metricName := fmt.Sprintf("metric-%d", i)
		if metrics[metricName] != int64(expectedMetricCount) {
			t.Errorf("Expected metric %s to be %d, got %d", metricName, expectedMetricCount, metrics[metricName])
		}
	}
}

func TestResults_GetItemsConcurrent(t *testing.T) {
	// Create a new Results collector
	results := NewResults()

	// Add some initial items
	for i := 0; i < 100; i++ {
		results.Add(fmt.Sprintf("item-%d", i))
	}

	// Test concurrent reads and writes
	var wg sync.WaitGroup
	wg.Add(2)

	// Writer goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			results.Add(fmt.Sprintf("new-item-%d", i))
			time.Sleep(time.Microsecond)
		}
	}()

	// Reader goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			items := results.GetItems()
			// Just verify we can read without crashing
			if len(items) < 100 {
				t.Errorf("Expected at least 100 items, got %d", len(items))
			}
			time.Sleep(time.Millisecond)
		}
	}()

	wg.Wait()
}

func TestResults_GetMetricConcurrent(t *testing.T) {
	// Create a new Results collector
	results := NewResults()

	// Add some initial metrics
	for i := 0; i < 10; i++ {
		metricName := fmt.Sprintf("metric-%d", i)
		results.AddMetric(metricName, 100)
	}

	// Test concurrent metric reads and writes
	var wg sync.WaitGroup
	wg.Add(2)

	// Writer goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			metricName := fmt.Sprintf("metric-%d", i%10)
			results.AddMetric(metricName, 1)
			time.Sleep(time.Microsecond)
		}
	}()

	// Reader goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			metricName := fmt.Sprintf("metric-%d", i%10)
			value := results.GetMetric(metricName)
			// Just verify we can read without crashing
			if value < 100 {
				t.Errorf("Expected metric %s to be at least 100, got %d", metricName, value)
			}
			time.Sleep(time.Millisecond)
		}
	}()

	wg.Wait()

	// Test GetAllMetrics
	allMetrics := results.GetAllMetrics()
	if len(allMetrics) != 10 {
		t.Errorf("Expected 10 metrics, got %d", len(allMetrics))
	}
}

// This race condition test runs with -race flag to detect race conditions
func TestLimitedErrGroup_RaceConditions(t *testing.T) {
	ctx := context.Background()
	g := NewLimitedErrGroup(ctx, 10)

	// Counter for tasks
	counter := atomic.Int32{}

	// Add a bunch of tasks
	for i := 0; i < 100; i++ {
		g.Go(func() error {
			counter.Add(1)
			return nil
		})
	}

	// Wait for all tasks to complete
	_ = g.Wait()

	// Make sure all tasks ran
	if counter.Load() != 100 {
		t.Errorf("Expected 100 tasks to run, got %d", counter.Load())
	}
}

// This explicitly tests for concurrent access to results
func TestResults_RaceDetection(t *testing.T) {
	results := NewResults()

	// Add from multiple goroutines simultaneously
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				results.Add(fmt.Sprintf("item-%d-%d", id, j))
				results.AddMetric("totalItems", 1)
				results.AddMetric(fmt.Sprintf("group-%d", id), 1)
			}
		}(i)
	}

	// Read while writes are happening
	for i := 0; i < 50; i++ {
		go func() {
			results.GetItems()
			results.GetMetric("totalItems")
			results.GetAllMetrics()
			time.Sleep(time.Millisecond)
		}()
	}

	wg.Wait()

	// Verify final counts
	if len(results.GetItems()) != 1000 {
		t.Errorf("Expected 1000 items, got %d", len(results.GetItems()))
	}

	if results.GetMetric("totalItems") != 1000 {
		t.Errorf("Expected totalItems to be 1000, got %d", results.GetMetric("totalItems"))
	}
}

// This test is focused specifically on running with the race detector
func TestLimitedErrGroup_ContextCancellationRace(t *testing.T) {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a group with a concurrency limit
	g := NewLimitedErrGroup(ctx, 5)

	// Add more tasks than the concurrency limit
	for i := 0; i < 20; i++ {
		i := i
		g.Go(func() error {
			// Some tasks will sleep longer, some shorter
			select {
			case <-time.After(time.Duration(i) * time.Millisecond):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}

	// Cancel after a small delay
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	// Wait for completion or cancellation
	err := g.Wait()

	// Either we get context cancellation or nil if tasks completed before cancellation
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("Expected nil or context.Canceled, got %v", err)
	}
}
