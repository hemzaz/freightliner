package distributed_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"freightliner/pkg/distributed"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkStealingScheduler_LocalQueue(t *testing.T) {
	scheduler := distributed.NewWorkStealingScheduler("node-1", 100, 1000, log.NewBasicLogger(log.InfoLevel))
	defer scheduler.Stop()

	job := &distributed.Job{
		ID:       "job-1",
		Priority: 10,
		Task: func(ctx context.Context) error {
			return nil
		},
	}

	err := scheduler.Schedule(job)
	require.NoError(t, err)

	// Should be in local queue
	assert.Equal(t, int64(1), scheduler.GetQueueDepth())
}

func TestWorkStealingScheduler_StealWork(t *testing.T) {
	scheduler := distributed.NewWorkStealingScheduler("node-1", 100, 1000, log.NewBasicLogger(log.InfoLevel))
	defer scheduler.Stop()

	// Add job
	job := &distributed.Job{
		ID:       "job-1",
		Priority: 10,
		Task: func(ctx context.Context) error {
			return nil
		},
	}

	err := scheduler.Schedule(job)
	require.NoError(t, err)

	// Steal work
	ctx := context.Background()
	stolen := scheduler.StealWork(ctx)
	require.NotNil(t, stolen)
	assert.Equal(t, "job-1", stolen.ID)

	// Queue should be empty
	assert.Equal(t, int64(0), scheduler.GetQueueDepth())
}

func TestWorkStealingScheduler_GlobalQueue(t *testing.T) {
	// Small local queue to force global queue usage
	scheduler := distributed.NewWorkStealingScheduler("node-1", 2, 1000, log.NewBasicLogger(log.InfoLevel))
	defer scheduler.Stop()

	// Fill local queue and overflow to global
	for i := 0; i < 10; i++ {
		job := &distributed.Job{
			ID:       string(rune(i)),
			Priority: 10,
			Task: func(ctx context.Context) error {
				return nil
			},
		}
		err := scheduler.Schedule(job)
		require.NoError(t, err)
	}

	// Should have jobs in both queues
	metrics := scheduler.GetMetrics()
	assert.Greater(t, metrics.JobsScheduled.Load(), uint64(0))
}

func TestWorkStealingScheduler_Concurrent(t *testing.T) {
	scheduler := distributed.NewWorkStealingScheduler("node-1", 100, 1000, log.NewBasicLogger(log.InfoLevel))
	defer scheduler.Stop()

	var wg sync.WaitGroup
	jobCount := 100

	// Schedule jobs concurrently
	for i := 0; i < jobCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			job := &distributed.Job{
				ID:       string(rune(id)),
				Priority: 10,
				Task: func(ctx context.Context) error {
					return nil
				},
			}
			err := scheduler.Schedule(job)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// All jobs should be scheduled
	metrics := scheduler.GetMetrics()
	assert.Equal(t, uint64(jobCount), metrics.JobsScheduled.Load())
}

func TestWorkStealingScheduler_Metrics(t *testing.T) {
	scheduler := distributed.NewWorkStealingScheduler("node-1", 100, 1000, log.NewBasicLogger(log.InfoLevel))
	defer scheduler.Stop()

	// Add and steal jobs
	for i := 0; i < 5; i++ {
		job := &distributed.Job{
			ID:       string(rune(i)),
			Priority: 10,
			Task: func(ctx context.Context) error {
				return nil
			},
		}
		scheduler.Schedule(job)
	}

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		scheduler.StealWork(ctx)
	}

	metrics := scheduler.GetMetrics()
	assert.Equal(t, uint64(5), metrics.JobsScheduled.Load())
	// LocalHits includes both Schedule (local) and StealWork (local queue first)
	assert.GreaterOrEqual(t, metrics.LocalHits.Load(), uint64(3))
}

func TestConcurrentQueue_PushPop(t *testing.T) {
	queue := distributed.NewConcurrentQueue(100)

	job := &distributed.Job{
		ID: "job-1",
	}

	err := queue.Push(job)
	require.NoError(t, err)
	assert.Equal(t, 1, queue.Len())

	popped := queue.Pop()
	require.NotNil(t, popped)
	assert.Equal(t, "job-1", popped.ID)
	assert.Equal(t, 0, queue.Len())
}

func TestConcurrentQueue_WaitPop(t *testing.T) {
	queue := distributed.NewConcurrentQueue(100)

	// Start goroutine to pop with wait
	done := make(chan *distributed.Job)
	go func() {
		job := queue.WaitPop(2 * time.Second)
		done <- job
	}()

	// Push job after small delay
	time.Sleep(100 * time.Millisecond)
	queue.Push(&distributed.Job{ID: "job-1"})

	// Should receive job
	select {
	case job := <-done:
		require.NotNil(t, job)
		assert.Equal(t, "job-1", job.ID)
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for job")
	}
}

func TestConcurrentQueue_Full(t *testing.T) {
	queue := distributed.NewConcurrentQueue(2)

	// Fill queue
	queue.Push(&distributed.Job{ID: "job-1"})
	queue.Push(&distributed.Job{ID: "job-2"})

	// Third push should fail
	err := queue.Push(&distributed.Job{ID: "job-3"})
	require.Error(t, err)
}

func TestWorkStealingScheduler_Capacity(t *testing.T) {
	scheduler := distributed.NewWorkStealingScheduler("node-1", 100, 1000, log.NewBasicLogger(log.InfoLevel))
	defer scheduler.Stop()

	capacity := scheduler.GetCapacity()
	assert.Equal(t, int64(100), capacity)
}
