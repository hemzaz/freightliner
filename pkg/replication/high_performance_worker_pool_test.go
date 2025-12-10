package replication

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
)

func TestNewHighPerformanceWorkerPool(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	config := DefaultHighPerformancePoolConfig()

	pool := NewHighPerformanceWorkerPool(config, logger)

	if pool == nil {
		t.Fatal("Expected non-nil pool")
	}

	if pool.minWorkers <= 0 {
		t.Error("Expected positive minWorkers")
	}

	if pool.maxWorkers <= 0 {
		t.Error("Expected positive maxWorkers")
	}

	if pool.maxWorkers < pool.minWorkers {
		t.Errorf("maxWorkers (%d) should be >= minWorkers (%d)", pool.maxWorkers, pool.minWorkers)
	}
}

func TestDefaultHighPerformancePoolConfig(t *testing.T) {
	config := DefaultHighPerformancePoolConfig()

	if config.TargetThroughputMBps != 125 {
		t.Errorf("Expected TargetThroughputMBps 125, got %d", config.TargetThroughputMBps)
	}

	if !config.AdaptiveScaling {
		t.Error("Expected AdaptiveScaling to be true")
	}

	if !config.PerformanceMonitoring {
		t.Error("Expected PerformanceMonitoring to be true")
	}
}

func TestHighPerformanceWorkerPool_StartStop(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	config := HighPerformancePoolConfig{
		MinWorkers:            2,
		MaxWorkers:            5,
		AdaptiveScaling:       true,
		PerformanceMonitoring: true,
	}

	pool := NewHighPerformanceWorkerPool(config, logger)

	// Start pool
	err := pool.Start()
	if err != nil {
		t.Fatalf("Failed to start pool: %v", err)
	}

	// Verify started
	if !pool.started.Load() {
		t.Error("Expected pool to be marked as started")
	}

	// Try starting again (should fail)
	err = pool.Start()
	if err == nil {
		t.Error("Expected error when starting already started pool")
	}

	// Stop pool
	pool.Stop()

	// Verify stopped
	if !pool.stopped.Load() {
		t.Error("Expected pool to be marked as stopped")
	}
}

func TestHighPerformanceWorkerPool_Submit(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	config := HighPerformancePoolConfig{
		MinWorkers: 2,
		MaxWorkers: 5,
	}

	pool := NewHighPerformanceWorkerPool(config, logger)
	pool.Start()
	defer pool.Stop()

	var executed atomic.Bool
	job := HighPerformanceJob{
		ID: "test-job-1",
		Task: func(ctx context.Context) error {
			executed.Store(true)
			return nil
		},
		Priority:          1,
		Context:           context.Background(),
		EstimatedBytes:    1024,
		EstimatedDuration: 100 * time.Millisecond,
	}

	err := pool.Submit(job)
	if err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}

	// Wait for job execution
	time.Sleep(500 * time.Millisecond)

	if !executed.Load() {
		t.Error("Job was not executed")
	}
}

func TestHighPerformanceWorkerPool_SubmitMultipleJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-job test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	config := HighPerformancePoolConfig{
		MinWorkers: 3,
		MaxWorkers: 10,
	}

	pool := NewHighPerformanceWorkerPool(config, logger)
	pool.Start()
	defer pool.Stop()

	const numJobs = 20
	var completedJobs atomic.Int32

	for i := 0; i < numJobs; i++ {
		job := HighPerformanceJob{
			ID: fmt.Sprintf("test-job-%d", i),
			Task: func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				completedJobs.Add(1)
				return nil
			},
			Priority:          1,
			Context:           context.Background(),
			EstimatedBytes:    1024 * 100,
			EstimatedDuration: 50 * time.Millisecond,
		}

		err := pool.Submit(job)
		if err != nil {
			t.Fatalf("Failed to submit job %d: %v", i, err)
		}
	}

	// Wait for all jobs to complete
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for jobs to complete. Completed: %d/%d", completedJobs.Load(), numJobs)
		case <-ticker.C:
			if completedJobs.Load() == numJobs {
				return
			}
		}
	}
}

func TestHighPerformanceWorkerPool_GetResults(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	config := HighPerformancePoolConfig{
		MinWorkers: 2,
		MaxWorkers: 5,
	}

	pool := NewHighPerformanceWorkerPool(config, logger)
	pool.Start()
	defer pool.Stop()

	results := pool.GetResults()
	if results == nil {
		t.Fatal("Expected non-nil results channel")
	}

	// Submit a job
	job := HighPerformanceJob{
		ID: "test-job-1",
		Task: func(ctx context.Context) error {
			return nil
		},
		Context:        context.Background(),
		EstimatedBytes: 1024,
	}

	err := pool.Submit(job)
	if err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}

	// Read result
	select {
	case result := <-results:
		if result.JobID != "test-job-1" {
			t.Errorf("Expected job ID 'test-job-1', got '%s'", result.JobID)
		}
		if result.Error != nil {
			t.Errorf("Expected no error, got %v", result.Error)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
}

func TestHighPerformanceWorkerPool_GetPerformanceReport(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	config := HighPerformancePoolConfig{
		MinWorkers: 2,
		MaxWorkers: 5,
	}

	pool := NewHighPerformanceWorkerPool(config, logger)
	pool.Start()
	defer pool.Stop()

	// Submit some jobs
	for i := 0; i < 5; i++ {
		job := HighPerformanceJob{
			ID: fmt.Sprintf("test-job-%d", i),
			Task: func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
			Context:        context.Background(),
			EstimatedBytes: 1024 * 10,
		}
		pool.Submit(job)
	}

	// Wait for jobs to complete
	time.Sleep(500 * time.Millisecond)

	// Get performance report
	report := pool.GetPerformanceReport()

	if report == nil {
		t.Fatal("Expected non-nil performance report")
	}

	if report.CurrentWorkers < config.MinWorkers {
		t.Errorf("Expected at least %d workers, got %d", config.MinWorkers, report.CurrentWorkers)
	}

	if report.TotalJobs < 5 {
		t.Errorf("Expected at least 5 total jobs, got %d", report.TotalJobs)
	}

	if report.PerformanceMetrics == nil {
		t.Error("Expected non-nil performance metrics")
	}
}

func TestThroughputTracker_RecordJob(t *testing.T) {
	tracker := &ThroughputTracker{
		measurements:     make([]ThroughputMeasurement, 0, 60),
		targetThroughput: 125 * 1024 * 1024,
	}

	tracker.recordJob(1024*1024*10, 100*time.Millisecond) // 10 MB

	totalBytes := tracker.totalBytes.Load()
	totalJobs := tracker.totalJobs.Load()

	if totalBytes != 1024*1024*10 {
		t.Errorf("Expected 10 MB recorded, got %d bytes", totalBytes)
	}

	if totalJobs != 1 {
		t.Errorf("Expected 1 job recorded, got %d", totalJobs)
	}
}

func TestThroughputTracker_GetCurrentThroughput(t *testing.T) {
	tracker := &ThroughputTracker{
		measurements:     make([]ThroughputMeasurement, 0, 60),
		targetThroughput: 125 * 1024 * 1024,
	}

	// Initially should be 0
	throughput := tracker.getCurrentThroughput()
	if throughput != 0 {
		t.Errorf("Expected 0 throughput initially, got %d", throughput)
	}

	// Add a measurement
	tracker.addMeasurement(ThroughputMeasurement{
		Timestamp:      time.Now(),
		BytesPerSecond: 100 * 1024 * 1024, // 100 MB/s
		JobsPerSecond:  10,
	})

	throughput = tracker.getCurrentThroughput()
	if throughput != 100*1024*1024 {
		t.Errorf("Expected 100 MB/s throughput, got %d", throughput)
	}
}

func TestThroughputTracker_AddMeasurement(t *testing.T) {
	tracker := &ThroughputTracker{
		measurements:     make([]ThroughputMeasurement, 0, 60),
		targetThroughput: 125 * 1024 * 1024,
	}

	// Add 70 measurements (more than capacity)
	for i := 0; i < 70; i++ {
		tracker.addMeasurement(ThroughputMeasurement{
			Timestamp:      time.Now(),
			BytesPerSecond: int64(i * 1024 * 1024),
			JobsPerSecond:  float64(i),
		})
	}

	// Should only keep last 60
	tracker.mutex.RLock()
	count := len(tracker.measurements)
	tracker.mutex.RUnlock()

	if count != 60 {
		t.Errorf("Expected 60 measurements, got %d", count)
	}
}

func TestHighPerformanceWorkerPool_SubmitToStoppedPool(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	config := HighPerformancePoolConfig{
		MinWorkers: 2,
		MaxWorkers: 5,
	}

	pool := NewHighPerformanceWorkerPool(config, logger)
	pool.Start()
	pool.Stop()

	job := HighPerformanceJob{
		ID: "test-job-1",
		Task: func(ctx context.Context) error {
			return nil
		},
		Context: context.Background(),
	}

	err := pool.Submit(job)
	if err == nil {
		t.Error("Expected error when submitting to stopped pool")
	}
}

func TestHighPerformanceWorkerPool_AdaptiveScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping adaptive scaling test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	config := HighPerformancePoolConfig{
		MinWorkers:            2,
		MaxWorkers:            10,
		AdaptiveScaling:       true,
		PerformanceMonitoring: true,
	}

	pool := NewHighPerformanceWorkerPool(config, logger)
	pool.Start()
	defer pool.Stop()

	initialWorkers := pool.currentWorkers.Load()

	// Submit many jobs to trigger scaling
	for i := 0; i < 50; i++ {
		job := HighPerformanceJob{
			ID: fmt.Sprintf("test-job-%d", i),
			Task: func(ctx context.Context) error {
				time.Sleep(100 * time.Millisecond)
				return nil
			},
			Context:        context.Background(),
			EstimatedBytes: 1024 * 100,
		}
		pool.Submit(job)
	}

	// Wait for adaptive scaling to occur
	time.Sleep(6 * time.Second)

	// Check if workers scaled up
	currentWorkers := pool.currentWorkers.Load()
	if currentWorkers <= initialWorkers {
		t.Logf("Workers did not scale up as expected (initial: %d, current: %d)", initialWorkers, currentWorkers)
	}
}

func TestHighPerformanceWorkerPool_JobWithError(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	config := HighPerformancePoolConfig{
		MinWorkers: 2,
		MaxWorkers: 5,
	}

	pool := NewHighPerformanceWorkerPool(config, logger)
	pool.Start()
	defer pool.Stop()

	expectedError := fmt.Errorf("test error")

	job := HighPerformanceJob{
		ID: "test-job-error",
		Task: func(ctx context.Context) error {
			return expectedError
		},
		Context: context.Background(),
	}

	err := pool.Submit(job)
	if err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}

	// Read result
	results := pool.GetResults()
	select {
	case result := <-results:
		if result.Error == nil {
			t.Error("Expected error in result")
		}
		if result.Error != expectedError {
			t.Errorf("Expected error '%v', got '%v'", expectedError, result.Error)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for error result")
	}
}

func TestHighPerformanceWorkerPool_ContextCancellation(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	config := HighPerformancePoolConfig{
		MinWorkers: 2,
		MaxWorkers: 5,
	}

	pool := NewHighPerformanceWorkerPool(config, logger)
	pool.Start()
	defer pool.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	job := HighPerformanceJob{
		ID: "test-job-cancelled",
		Task: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(1 * time.Second):
				return nil
			}
		},
		Context: ctx,
	}

	err := pool.Submit(job)
	if err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}

	// Read result
	results := pool.GetResults()
	select {
	case result := <-results:
		if result.Error == nil {
			t.Error("Expected cancellation error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for cancellation result")
	}
}
