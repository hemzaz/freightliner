//go:build race
// +build race

// Race detection tests - Run with: go test -race ./tests/race/...

package race

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/replication"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkerPoolRaceConditions tests worker pool for race conditions
func TestWorkerPoolRaceConditions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		workerCount int
		jobCount    int
	}{
		{"FewWorkers", 5, 100},
		{"ManyWorkers", 50, 100},
		{"HighConcurrency", 100, 1000},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := log.NewBasicLogger(log.ErrorLevel)
			pool := replication.NewWorkerPool(tc.workerCount, logger)
			pool.Start()
			defer pool.Stop()

			var counter atomic.Int64
			var wg sync.WaitGroup

			// Submit jobs concurrently from multiple goroutines
			for i := 0; i < tc.jobCount; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					jobID := fmt.Sprintf("job-%d", id)
					err := pool.Submit(jobID, func(ctx context.Context) error {
						counter.Add(1)
						time.Sleep(1 * time.Millisecond)
						return nil
					})

					assert.NoError(t, err)
				}(i)
			}

			wg.Wait()
			pool.Wait()

			// Verify counter
			assert.Equal(t, int64(tc.jobCount), counter.Load())
		})
	}
}

// TestSchedulerRaceConditions tests scheduler for race conditions
func TestSchedulerRaceConditions(t *testing.T) {
	t.Parallel()

	logger := log.NewBasicLogger(log.ErrorLevel)
	pool := replication.NewWorkerPool(10, logger)
	pool.Start()
	defer pool.Stop()

	scheduler := replication.NewScheduler(replication.SchedulerOptions{
		Logger:     logger,
		WorkerPool: pool,
	})
	defer scheduler.Stop()

	var wg sync.WaitGroup

	// Add jobs concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			rule := replication.ReplicationRule{
				SourceRegistry:        "source",
				SourceRepository:      fmt.Sprintf("repo-%d", id),
				DestinationRegistry:   "dest",
				DestinationRepository: fmt.Sprintf("repo-%d", id),
				Schedule:              "@now",
			}

			err := scheduler.AddJob(rule)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
	time.Sleep(2 * time.Second)
}

// TestConcurrentMapAccess tests concurrent map access
func TestConcurrentMapAccess(t *testing.T) {
	t.Parallel()

	data := make(map[string]int)
	var mu sync.RWMutex
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", id)
			mu.Lock()
			data[key] = id
			mu.Unlock()
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", id)
			mu.RLock()
			_ = data[key]
			mu.RUnlock()
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 100, len(data))
}

// TestAtomicOperationsRace tests atomic operations under race conditions
func TestAtomicOperationsRace(t *testing.T) {
	t.Parallel()

	var counter atomic.Int64
	var wg sync.WaitGroup

	// Increment concurrently
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Add(1)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(1000), counter.Load())
}

// TestChannelRaceConditions tests channel operations
func TestChannelRaceConditions(t *testing.T) {
	t.Parallel()

	ch := make(chan int, 100)
	var wg sync.WaitGroup

	// Producers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				ch <- start*10 + j
			}
		}(i)
	}

	// Close channel after all producers finish
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Consumer
	received := make(map[int]bool)
	var receiveMu sync.Mutex

	var consumerWg sync.WaitGroup
	for i := 0; i < 5; i++ {
		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			for val := range ch {
				receiveMu.Lock()
				received[val] = true
				receiveMu.Unlock()
			}
		}()
	}

	consumerWg.Wait()
	assert.Equal(t, 100, len(received))
}

// TestContextCancellationRace tests context cancellation under concurrent load
func TestContextCancellationRace(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Start goroutine that waits for cancellation
			done := make(chan struct{})
			go func() {
				<-ctx.Done()
				close(done)
			}()

			// Cancel immediately
			cancel()

			// Wait for done
			select {
			case <-done:
			case <-time.After(1 * time.Second):
				t.Errorf("Context cancellation timeout for goroutine %d", id)
			}
		}(i)
	}

	wg.Wait()
}

// TestWorkerPoolStopRace tests stopping worker pool under load
func TestWorkerPoolStopRace(t *testing.T) {
	t.Parallel()

	logger := log.NewBasicLogger(log.ErrorLevel)
	pool := replication.NewWorkerPool(10, logger)
	pool.Start()

	var wg sync.WaitGroup

	// Submit jobs
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			jobID := fmt.Sprintf("job-%d", id)
			_ = pool.Submit(jobID, func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}(i)
	}

	// Stop pool concurrently
	go func() {
		time.Sleep(20 * time.Millisecond)
		pool.Stop()
	}()

	wg.Wait()
}

// TestJobResultChannelRace tests job result channel operations
func TestJobResultChannelRace(t *testing.T) {
	t.Parallel()

	logger := log.NewBasicLogger(log.ErrorLevel)
	pool := replication.NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	var wg sync.WaitGroup
	results := pool.GetResults()

	// Submit jobs
	jobCount := 100
	for i := 0; i < jobCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			jobID := fmt.Sprintf("job-%d", id)
			err := pool.Submit(jobID, func(ctx context.Context) error {
				time.Sleep(1 * time.Millisecond)
				return nil
			})
			assert.NoError(t, err)
		}(i)
	}

	// Read results concurrently
	received := make(map[string]bool)
	var receiveMu sync.Mutex

	var resultWg sync.WaitGroup
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		count := 0
		for result := range results {
			receiveMu.Lock()
			received[result.JobID] = true
			receiveMu.Unlock()
			count++
			if count >= jobCount {
				break
			}
		}
	}()

	wg.Wait()
	pool.Wait()
	resultWg.Wait()

	assert.GreaterOrEqual(t, len(received), jobCount/2, "Should receive at least half the results")
}

// TestSharedStateRace tests shared state modifications
func TestSharedStateRace(t *testing.T) {
	t.Parallel()

	type SharedState struct {
		mu      sync.RWMutex
		counter int
		data    map[string]int
	}

	state := &SharedState{
		data: make(map[string]int),
	}

	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			state.mu.Lock()
			state.counter++
			state.data[fmt.Sprintf("key-%d", id)] = id
			state.mu.Unlock()
		}(i)
	}

	// Readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			state.mu.RLock()
			_ = state.counter
			_ = len(state.data)
			state.mu.RUnlock()
		}()
	}

	wg.Wait()

	state.mu.RLock()
	assert.Equal(t, 50, state.counter)
	assert.Equal(t, 50, len(state.data))
	state.mu.RUnlock()
}

// TestMemoryBarrierRace tests memory barriers and synchronization
func TestMemoryBarrierRace(t *testing.T) {
	t.Parallel()

	var ready atomic.Bool
	var data int

	var wg sync.WaitGroup

	// Writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		data = 42
		ready.Store(true)
	}()

	// Readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !ready.Load() {
				time.Sleep(1 * time.Microsecond)
			}
			assert.Equal(t, 42, data)
		}()
	}

	wg.Wait()
}

// TestDoubleCheckLockingRace tests double-checked locking pattern
func TestDoubleCheckLockingRace(t *testing.T) {
	t.Parallel()

	type Singleton struct {
		mu    sync.Mutex
		value atomic.Value
	}

	s := &Singleton{}

	getInstance := func() interface{} {
		if val := s.value.Load(); val != nil {
			return val
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		if val := s.value.Load(); val != nil {
			return val
		}

		instance := "initialized"
		s.value.Store(instance)
		return instance
	}

	var wg sync.WaitGroup

	// Multiple goroutines trying to get instance
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val := getInstance()
			assert.NotNil(t, val)
			assert.Equal(t, "initialized", val)
		}()
	}

	wg.Wait()
}

// TestTimerRace tests timer usage under concurrent conditions
func TestTimerRace(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			timer := time.NewTimer(10 * time.Millisecond)
			defer timer.Stop()

			select {
			case <-timer.C:
				// Timer fired
			case <-time.After(100 * time.Millisecond):
				t.Error("Timer should have fired")
			}
		}()
	}

	wg.Wait()
}

// TestWaitGroupRace tests WaitGroup under concurrent conditions
func TestWaitGroupRace(t *testing.T) {
	t.Parallel()

	for iteration := 0; iteration < 100; iteration++ {
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(1 * time.Millisecond)
			}()
		}

		wg.Wait()
	}
}

// TestSelectRace tests select statement under concurrent load
func TestSelectRace(t *testing.T) {
	t.Parallel()

	ch1 := make(chan int)
	ch2 := make(chan int)
	done := make(chan struct{})

	var wg sync.WaitGroup

	// Senders
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			select {
			case ch1 <- i:
			case <-done:
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			select {
			case ch2 <- i:
			case <-done:
				return
			}
		}
	}()

	// Receiver
	wg.Add(1)
	go func() {
		defer wg.Done()
		count := 0
		for count < 100 {
			select {
			case <-ch1:
				count++
			case <-ch2:
				count++
			case <-time.After(100 * time.Millisecond):
				return
			}
		}
	}()

	// Wait or timeout
	done2 := make(chan struct{})
	go func() {
		wg.Wait()
		close(done2)
	}()

	select {
	case <-done2:
		close(done)
	case <-time.After(5 * time.Second):
		close(done)
		t.Error("Test timed out")
	}
}

// TestConcurrentJobSubmission stresses concurrent job submission
func TestConcurrentJobSubmission(t *testing.T) {
	t.Parallel()

	logger := log.NewBasicLogger(log.ErrorLevel)
	pool := replication.NewWorkerPool(20, logger)
	pool.Start()
	defer pool.Stop()

	var submitted atomic.Int64
	var completed atomic.Int64

	var wg sync.WaitGroup

	// Submit jobs from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				jobID := fmt.Sprintf("routine-%d-job-%d", routineID, j)
				err := pool.Submit(jobID, func(ctx context.Context) error {
					completed.Add(1)
					time.Sleep(1 * time.Millisecond)
					return nil
				})

				if err == nil {
					submitted.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()
	pool.Wait()

	t.Logf("Submitted: %d, Completed: %d", submitted.Load(), completed.Load())
	require.Greater(t, submitted.Load(), int64(0), "Should have submitted jobs")
}
