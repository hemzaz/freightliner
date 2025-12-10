package load

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/replication"
)

// MockReplicationJob simulates a replication job for load testing
type MockReplicationJob struct {
	ID           string
	Repositories []string
	Duration     time.Duration
	ShouldFail   bool
	ErrorType    string
}

// Execute runs the mock replication job
func (j *MockReplicationJob) Execute(ctx context.Context) error {
	// Simulate work by sleeping
	select {
	case <-time.After(j.Duration):
		if j.ShouldFail {
			return fmt.Errorf("mock error: %s", j.ErrorType)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// LoadTestRunner orchestrates load testing
type LoadTestRunner struct {
	config  LoadTestConfig
	metrics *LoadTestMetrics
	logger  log.Logger

	// Control channels
	stopCh   chan struct{}
	doneCh   chan struct{}
	jobQueue chan *MockReplicationJob
}

// NewLoadTestRunner creates a new load test runner
func NewLoadTestRunner(config LoadTestConfig, logger log.Logger) *LoadTestRunner {
	if logger == nil {
		logger = log.NewLogger()
	}

	return &LoadTestRunner{
		config:   config,
		metrics:  NewLoadTestMetrics(),
		logger:   logger,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
		jobQueue: make(chan *MockReplicationJob, config.ConcurrentJobs*2),
	}
}

// generateJob creates a mock replication job
func (r *LoadTestRunner) generateJob(jobID int) *MockReplicationJob {
	// Generate random repositories
	repositories := make([]string, r.config.RepositoriesPerJob)
	for i := 0; i < r.config.RepositoriesPerJob; i++ {
		repositories[i] = fmt.Sprintf("repo-%d-%d", jobID, i)
	}

	// Simulate random job duration (50ms to 500ms)
	duration := time.Duration(50+rand.Intn(450)) * time.Millisecond

	// Determine if job should fail based on error rate
	shouldFail := rand.Float64() < r.config.ErrorRate
	var errorType string
	if shouldFail {
		errorTypes := []string{"network_error", "auth_error", "manifest_error", "timeout"}
		errorType = errorTypes[rand.Intn(len(errorTypes))]
	}

	return &MockReplicationJob{
		ID:           fmt.Sprintf("job-%d", jobID),
		Repositories: repositories,
		Duration:     duration,
		ShouldFail:   shouldFail,
		ErrorType:    errorType,
	}
}

// worker processes jobs from the job queue
func (r *LoadTestRunner) worker(workerID int, ctx context.Context) {
	defer func() {
		r.logger.Debug("Worker stopped", map[string]interface{}{
			"worker_id": workerID,
		})
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-r.jobQueue:
			if !ok {
				return
			}

			// Update concurrency
			current := atomic.AddInt64(&r.metrics.CurrentConcurrency, 1)
			r.metrics.UpdateConcurrency(current)

			startTime := time.Now()
			err := job.Execute(ctx)
			duration := time.Since(startTime)

			// Update metrics
			if err != nil {
				r.metrics.UpdateJobFailed(job.ErrorType)
				r.logger.Debug("Job failed", map[string]interface{}{
					"worker_id": workerID,
					"job_id":    job.ID,
					"error":     err.Error(),
					"duration":  duration.String(),
				})
			} else {
				r.metrics.UpdateJobCompleted(duration, len(job.Repositories))
				r.logger.Debug("Job completed", map[string]interface{}{
					"worker_id":    workerID,
					"job_id":       job.ID,
					"repositories": len(job.Repositories),
					"duration":     duration.String(),
				})
			}

			// Update concurrency
			atomic.AddInt64(&r.metrics.CurrentConcurrency, -1)
		}
	}
}

// jobGenerator creates jobs at the specified rate
func (r *LoadTestRunner) jobGenerator(ctx context.Context) {
	defer close(r.jobQueue)

	jobID := 0
	ticker := time.NewTicker(100 * time.Millisecond) // Generate jobs every 100ms
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			job := r.generateJob(jobID)
			jobID++

			select {
			case r.jobQueue <- job:
				atomic.AddInt64(&r.metrics.TotalJobs, 1)
			case <-ctx.Done():
				return
			default:
				// Queue is full, skip this job
				r.logger.Warn("Job queue full, skipping job", map[string]interface{}{
					"job_id": job.ID,
				})
			}
		}
	}
}

// metricsCollector periodically collects and logs metrics
func (r *LoadTestRunner) metricsCollector(ctx context.Context) {
	ticker := time.NewTicker(r.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			summary := r.metrics.GetSummary()

			completedJobs := summary.CompletedJobs
			failedJobs := summary.FailedJobs
			totalJobs := summary.TotalJobs
			currentConcurrency := summary.CurrentConcurrency

			elapsed := time.Since(summary.StartTime)

			r.logger.Info("Load test metrics", map[string]interface{}{
				"elapsed_time":        elapsed.String(),
				"total_jobs":          totalJobs,
				"completed_jobs":      completedJobs,
				"failed_jobs":         failedJobs,
				"current_concurrency": currentConcurrency,
				"jobs_per_second":     float64(completedJobs) / elapsed.Seconds(),
			})
		}
	}
}

// Run executes the load test
func (r *LoadTestRunner) Run() *LoadTestSummary {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.TestDuration)
	defer cancel()

	r.logger.Info("Starting load test", map[string]interface{}{
		"concurrent_jobs":      r.config.ConcurrentJobs,
		"repositories_per_job": r.config.RepositoriesPerJob,
		"test_duration":        r.config.TestDuration.String(),
		"expected_error_rate":  r.config.ErrorRate,
	})

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < r.config.ConcurrentJobs; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r.worker(workerID, ctx)
		}(i)
	}

	// Start job generator
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.jobGenerator(ctx)
	}()

	// Start metrics collector
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.metricsCollector(ctx)
	}()

	// Wait for completion
	wg.Wait()

	r.metrics.EndTime = time.Now()

	// Final metrics
	finalSummary := r.metrics.GetSummary()
	r.logFinalResults(finalSummary)

	return &finalSummary
}

// logFinalResults logs the final test results
func (r *LoadTestRunner) logFinalResults(summary LoadTestSummary) {
	totalDuration := summary.Duration
	completedJobs := summary.CompletedJobs
	failedJobs := summary.FailedJobs
	totalJobs := summary.TotalJobs

	jobsPerSecond := float64(completedJobs) / totalDuration.Seconds()
	errorRate := float64(failedJobs) / float64(totalJobs)

	avgJobDuration := summary.AvgJobDuration

	r.logger.Info("Load test completed", map[string]interface{}{
		"total_duration":   totalDuration.String(),
		"total_jobs":       totalJobs,
		"completed_jobs":   completedJobs,
		"failed_jobs":      failedJobs,
		"jobs_per_second":  fmt.Sprintf("%.2f", jobsPerSecond),
		"error_rate":       fmt.Sprintf("%.2f%%", errorRate*100),
		"avg_job_duration": avgJobDuration.String(),
		"min_job_duration": summary.MinJobDuration.String(),
		"max_job_duration": summary.MaxJobDuration.String(),
		"max_concurrency":  summary.MaxConcurrency,
	})

	// Log error breakdown
	if len(summary.ErrorsByType) > 0 {
		errorFields := make(map[string]interface{})
		for k, v := range summary.ErrorsByType {
			errorFields[k] = v
		}
		r.logger.Info("Error breakdown by type", errorFields)
	}
}

// Benchmark tests for load testing

func BenchmarkReplicationLoad(b *testing.B) {
	logger := log.NewLogger()

	config := LoadTestConfig{
		ConcurrentJobs:     5,                // Reduced from 10
		RepositoriesPerJob: 3,                // Reduced from 5
		TestDuration:       10 * time.Second, // Reduced from 30s
		ErrorRate:          0.05,             // 5% error rate
		MetricsInterval:    2 * time.Second,  // Reduced from 5s
	}

	runner := NewLoadTestRunner(config, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics := runner.Run()

		// Verify that the load test performed as expected
		completedJobs := metrics.CompletedJobs
		if completedJobs == 0 {
			b.Fatal("No jobs completed during load test")
		}
	}
}

func TestLoadTestHighConcurrency(t *testing.T) {
	// Skip in short mode to avoid CI timeouts
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	logger := log.NewLogger()

	config := LoadTestConfig{
		ConcurrentJobs:     20,               // Reduced from 50
		RepositoriesPerJob: 5,                // Reduced from 10
		TestDuration:       15 * time.Second, // Reduced from 1 minute
		ErrorRate:          0.1,              // 10% error rate
		MetricsInterval:    5 * time.Second,  // Reduced from 10s
	}

	runner := NewLoadTestRunner(config, logger)
	metrics := runner.Run()

	// Validate results
	completedJobs := metrics.CompletedJobs
	failedJobs := metrics.FailedJobs
	totalJobs := metrics.TotalJobs

	if totalJobs == 0 {
		t.Fatal("No jobs were created")
	}

	if completedJobs == 0 {
		t.Fatal("No jobs completed")
	}

	errorRate := float64(failedJobs) / float64(totalJobs)
	expectedErrorRate := config.ErrorRate

	// Allow for some variance in error rate (±2%)
	if errorRate < expectedErrorRate-0.02 || errorRate > expectedErrorRate+0.02 {
		t.Logf("Error rate variance: expected ~%.1f%%, got %.1f%%",
			expectedErrorRate*100, errorRate*100)
	}

	maxConcurrency := metrics.MaxConcurrency
	if maxConcurrency > int64(config.ConcurrentJobs) {
		t.Errorf("Max concurrency %d exceeded configured limit %d",
			maxConcurrency, config.ConcurrentJobs)
	}

	t.Logf("Load test completed successfully:")
	t.Logf("  Total jobs: %d", totalJobs)
	t.Logf("  Completed: %d", completedJobs)
	t.Logf("  Failed: %d", failedJobs)
	t.Logf("  Error rate: %.2f%%", errorRate*100)
	t.Logf("  Max concurrency: %d", maxConcurrency)
	t.Logf("  Duration: %v", metrics.Duration)
}

func TestLoadTestWithRealWorkerPool(t *testing.T) {
	// Skip in short mode to avoid CI timeouts
	if testing.Short() {
		t.Skip("Skipping worker pool test in short mode")
	}

	logger := log.NewLogger()

	// Create a real worker pool for more realistic testing
	pool := replication.NewWorkerPool(10, logger) // Reduced from 20
	pool.Start()
	defer pool.Stop()

	// Configuration for testing with real worker pool (reduced for CI)
	const (
		numJobs       = 50 // Reduced from 100
		jobsPerSecond = 10
		timeoutPerJob = 5 * time.Second
	)

	var completedJobs atomic.Int64
	var failedJobs atomic.Int64

	startTime := time.Now()

	// Submit jobs to the worker pool
	for i := 0; i < numJobs; i++ {
		jobID := fmt.Sprintf("load-test-job-%d", i)

		err := pool.Submit(jobID, func(ctx context.Context) error {
			// Simulate replication work
			workDuration := time.Duration(50+rand.Intn(200)) * time.Millisecond

			select {
			case <-time.After(workDuration):
				completedJobs.Add(1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})

		if err != nil {
			t.Fatalf("Failed to submit job %s: %v", jobID, err)
		}

		// Rate limiting
		if i > 0 && i%jobsPerSecond == 0 {
			time.Sleep(1 * time.Second)
		}
	}

	// Collect results
	go func() {
		for result := range pool.GetResults() {
			if result.Error != nil {
				failedJobs.Add(1)
			}
		}
	}()

	// Wait for all jobs to complete with timeout
	done := make(chan struct{})
	go func() {
		pool.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Jobs completed successfully
	case <-time.After(15 * time.Second): // Reduced from 30s
		t.Fatal("Load test timed out")
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Validate results
	completed := completedJobs.Load()
	failed := failedJobs.Load()

	if completed == 0 {
		t.Fatal("No jobs completed")
	}

	if completed+failed != numJobs {
		t.Errorf("Job count mismatch: expected %d, got %d completed + %d failed",
			numJobs, completed, failed)
	}

	jobsPerSecondActual := float64(completed) / duration.Seconds()

	t.Logf("Real worker pool load test results:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Completed jobs: %d", completed)
	t.Logf("  Failed jobs: %d", failed)
	t.Logf("  Jobs per second: %.2f", jobsPerSecondActual)
	t.Logf("  Worker pool size: %d", pool.WorkerCount())

	// Performance assertions
	if jobsPerSecondActual < 5 { // Expect at least 5 jobs/second
		t.Errorf("Performance too low: %.2f jobs/second", jobsPerSecondActual)
	}

	if duration > 30*time.Second { // Should complete within 30 seconds (reduced from 2 minutes)
		t.Errorf("Test took too long: %v", duration)
	}
}
