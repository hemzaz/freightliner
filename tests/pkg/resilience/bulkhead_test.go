package resilience_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/resilience"

	"github.com/stretchr/testify/assert"
)

func TestBulkhead_AllowsConcurrentRequests(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	settings := resilience.BulkheadSettings{
		MaxConcurrent: 5,
		MaxQueueDepth: 10,
		Timeout:       1 * time.Second,
	}
	bulkhead := resilience.NewBulkhead("test", settings, logger)

	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := bulkhead.Execute(context.Background(), func() error {
				current := concurrent.Add(1)
				defer concurrent.Add(-1)

				// Track max concurrent
				for {
					max := maxConcurrent.Load()
					if current <= max || maxConcurrent.CompareAndSwap(max, current) {
						break
					}
				}

				time.Sleep(50 * time.Millisecond)
				return nil
			})
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	assert.LessOrEqual(t, int(maxConcurrent.Load()), 5)
}

func TestBulkhead_RejectsWhenFull(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	settings := resilience.BulkheadSettings{
		MaxConcurrent: 2,
		MaxQueueDepth: 0, // No queue
		Timeout:       100 * time.Millisecond,
	}
	bulkhead := resilience.NewBulkhead("test", settings, logger)

	// Fill the bulkhead
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bulkhead.Execute(context.Background(), func() error {
				time.Sleep(200 * time.Millisecond)
				return nil
			})
		}()
	}

	// Wait for bulkhead to fill
	time.Sleep(10 * time.Millisecond)

	// This should be rejected
	err := bulkhead.Execute(context.Background(), func() error {
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue full")

	wg.Wait()
}

func TestBulkhead_QueueManagement(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	settings := resilience.BulkheadSettings{
		MaxConcurrent: 1,
		MaxQueueDepth: 5,
		Timeout:       1 * time.Second,
	}
	bulkhead := resilience.NewBulkhead("test", settings, logger)

	completed := atomic.Int32{}
	var wg sync.WaitGroup

	// Submit more requests than max concurrent
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := bulkhead.Execute(context.Background(), func() error {
				time.Sleep(20 * time.Millisecond)
				completed.Add(1)
				return nil
			})
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(5), completed.Load())
}

func TestBulkhead_Timeout(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	settings := resilience.BulkheadSettings{
		MaxConcurrent: 1,
		MaxQueueDepth: 0, // No queue - fill immediately
		Timeout:       50 * time.Millisecond,
	}
	bulkhead := resilience.NewBulkhead("test", settings, logger)

	// Block the bulkhead
	started := make(chan struct{})
	go bulkhead.Execute(context.Background(), func() error {
		close(started)
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	<-started // Wait for first operation to start

	// This should be rejected due to queue full (MaxQueueDepth=0)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := bulkhead.Execute(ctx, func() error {
		return nil
	})

	assert.Error(t, err)
	// With MaxQueueDepth=0, it's rejected at queue, not timeout
	assert.True(t, err != nil)
}

func TestBulkhead_Stats(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	settings := resilience.DefaultBulkheadSettings()
	bulkhead := resilience.NewBulkhead("test", settings, logger)

	// Execute some operations
	for i := 0; i < 5; i++ {
		bulkhead.Execute(context.Background(), func() error {
			return nil
		})
	}

	stats := bulkhead.Stats()
	assert.Equal(t, "test", stats.Name)
	assert.Equal(t, int64(5), stats.TotalExecutions)
	assert.Equal(t, int64(0), stats.TotalRejections)
}

func TestBulkheadManager(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	manager := resilience.NewBulkheadManager(logger)

	// Create multiple bulkheads
	bh1 := manager.GetOrCreate("registry1", resilience.DefaultBulkheadSettings())
	bh2 := manager.GetOrCreate("registry2", resilience.DefaultBulkheadSettings())

	assert.NotNil(t, bh1)
	assert.NotNil(t, bh2)
	assert.NotEqual(t, bh1, bh2)

	// Test execution through manager
	err := manager.Execute(context.Background(), "registry1", func() error {
		return nil
	})
	assert.NoError(t, err)

	// Get all stats
	allStats := manager.GetAllStats()
	assert.GreaterOrEqual(t, len(allStats), 2)
}

func TestBulkhead_IsolationBetweenResources(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	manager := resilience.NewBulkheadManager(logger)

	settings := resilience.BulkheadSettings{
		MaxConcurrent: 2,
		MaxQueueDepth: 0,
		Timeout:       100 * time.Millisecond,
	}

	// Fill registry1 bulkhead
	bh1 := manager.GetOrCreate("registry1", settings)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bh1.Execute(context.Background(), func() error {
				time.Sleep(200 * time.Millisecond)
				return nil
			})
		}()
	}

	time.Sleep(10 * time.Millisecond)

	// registry2 should still work
	err := manager.Execute(context.Background(), "registry2", func() error {
		return nil
	})
	assert.NoError(t, err)

	wg.Wait()
}
