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

// LoadTestConfig defines configuration for load testing
type LoadTestConfig struct {
	// ConcurrentJobs is the number of concurrent replication jobs
	ConcurrentJobs int

	// RepositoriesPerJob is the number of repositories per job
	RepositoriesPerJob int

	// TestDuration is how long to run the load test
	TestDuration time.Duration

	// RampUpTime is the time to gradually increase load
	RampUpTime time.Duration

	// RampDownTime is the time to gradually decrease load
	RampDownTime time.Duration

	// ErrorRate is the expected error rate (0.0-1.0)
	ErrorRate float64

	// MetricsInterval is how often to collect metrics
	MetricsInterval time.Duration
}

// LoadTestMetrics tracks metrics during load testing
type LoadTestMetrics struct {
	mu sync.RWMutex

	// Counters
	TotalJobs         int64
	CompletedJobs     int64
	FailedJobs        int64
	TotalRepositories int64

	// Timing
	StartTime      time.Time
	EndTime        time.Time
	MinJobDuration time.Duration
	MaxJobDuration time.Duration
	TotalJobTime   time.Duration

	// Concurrency
	CurrentConcurrency int64
	MaxConcurrency     int64

	// Error tracking
	ErrorsByType map[string]int64

	// Memory and resource tracking
	PeakMemoryMB   int64
	PeakGoroutines int64
}

// NewLoadTestMetrics creates a new metrics tracker
func NewLoadTestMetrics() *LoadTestMetrics {
	return &LoadTestMetrics{
		StartTime:      time.Now(),
		ErrorsByType:   make(map[string]int64),
		MinJobDuration: time.Hour, // Initialize to a large value
	}
}

// UpdateJobCompleted updates metrics when a job completes
func (m *LoadTestMetrics) UpdateJobCompleted(duration time.Duration, repositories int) {
	atomic.AddInt64(&m.CompletedJobs, 1)
	atomic.AddInt64(&m.TotalRepositories, int64(repositories))
	atomic.AddInt64(&m.TotalJobTime, int64(duration))

	m.mu.Lock()
	defer m.mu.Unlock()

	if duration < m.MinJobDuration {
		m.MinJobDuration = duration
	}
	if duration > m.MaxJobDuration {
		m.MaxJobDuration = duration
	}
}

// UpdateJobFailed updates metrics when a job fails
func (m *LoadTestMetrics) UpdateJobFailed(errorType string) {
	atomic.AddInt64(&m.FailedJobs, 1)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorsByType[errorType]++
}

// UpdateConcurrency updates concurrency tracking
func (m *LoadTestMetrics) UpdateConcurrency(current int64) {
	atomic.StoreInt64(&m.CurrentConcurrency, current)

	for {
		max := atomic.LoadInt64(&m.MaxConcurrency)
		if current <= max || atomic.CompareAndSwapInt64(&m.MaxConcurrency, max, current) {
			break
		}
	}
}

// GetSnapshot returns a snapshot of current metrics
func (m *LoadTestMetrics) GetSnapshot() LoadTestMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := *m
	snapshot.ErrorsByType = make(map[string]int64)
	for k, v := range m.ErrorsByType {
		snapshot.ErrorsByType[k] = v
	}

	return snapshot
}

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
	logger  *log.Logger

	// Control channels
	stopCh   chan struct{}
	doneCh   chan struct{}
	jobQueue chan *MockReplicationJob
}

// NewLoadTestRunner creates a new load test runner
func NewLoadTestRunner(config LoadTestConfig, logger *log.Logger) *LoadTestRunner {
	if logger == nil {
		logger = log.NewLogger(log.InfoLevel)
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
			snapshot := r.metrics.GetSnapshot()

			completedJobs := atomic.LoadInt64(&snapshot.CompletedJobs)
			failedJobs := atomic.LoadInt64(&snapshot.FailedJobs)
			totalJobs := atomic.LoadInt64(&snapshot.TotalJobs)
			currentConcurrency := atomic.LoadInt64(&snapshot.CurrentConcurrency)

			elapsed := time.Since(snapshot.StartTime)

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
func (r *LoadTestRunner) Run() *LoadTestMetrics {
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
	finalMetrics := r.metrics.GetSnapshot()
	r.logFinalResults(finalMetrics)

	return &finalMetrics
}

// logFinalResults logs the final test results
func (r *LoadTestRunner) logFinalResults(metrics LoadTestMetrics) {
	totalDuration := metrics.EndTime.Sub(metrics.StartTime)
	completedJobs := atomic.LoadInt64(&metrics.CompletedJobs)
	failedJobs := atomic.LoadInt64(&metrics.FailedJobs)
	totalJobs := atomic.LoadInt64(&metrics.TotalJobs)

	jobsPerSecond := float64(completedJobs) / totalDuration.Seconds()
	errorRate := float64(failedJobs) / float64(totalJobs)

	avgJobDuration := time.Duration(0)
	if completedJobs > 0 {
		avgJobDuration = time.Duration(atomic.LoadInt64(&metrics.TotalJobTime) / completedJobs)
	}

	r.logger.Info("Load test completed", map[string]interface{}{
		"total_duration":   totalDuration.String(),
		"total_jobs":       totalJobs,
		"completed_jobs":   completedJobs,
		"failed_jobs":      failedJobs,
		"jobs_per_second":  fmt.Sprintf("%.2f", jobsPerSecond),
		"error_rate":       fmt.Sprintf("%.2f%%", errorRate*100),
		"avg_job_duration": avgJobDuration.String(),
		"min_job_duration": metrics.MinJobDuration.String(),
		"max_job_duration": metrics.MaxJobDuration.String(),
		"max_concurrency":  atomic.LoadInt64(&metrics.MaxConcurrency),
	})

	// Log error breakdown
	if len(metrics.ErrorsByType) > 0 {
		r.logger.Info("Error breakdown by type", map[string]interface{}(metrics.ErrorsByType))
	}
}

// Benchmark tests for load testing

func BenchmarkReplicationLoad(b *testing.B) {
	logger := log.NewLogger(log.InfoLevel)

	config := LoadTestConfig{
		ConcurrentJobs:     10,
		RepositoriesPerJob: 5,
		TestDuration:       30 * time.Second,
		ErrorRate:          0.05, // 5% error rate
		MetricsInterval:    5 * time.Second,
	}

	runner := NewLoadTestRunner(config, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics := runner.Run()

		// Verify that the load test performed as expected
		completedJobs := atomic.LoadInt64(&metrics.CompletedJobs)
		if completedJobs == 0 {
			b.Fatal("No jobs completed during load test")
		}
	}
}

func TestLoadTestHighConcurrency(t *testing.T) {
	logger := log.NewLogger(log.InfoLevel)

	config := LoadTestConfig{
		ConcurrentJobs:     50,
		RepositoriesPerJob: 10,
		TestDuration:       1 * time.Minute,
		ErrorRate:          0.1, // 10% error rate
		MetricsInterval:    10 * time.Second,
	}

	runner := NewLoadTestRunner(config, logger)
	metrics := runner.Run()

	// Validate results
	completedJobs := atomic.LoadInt64(&metrics.CompletedJobs)
	failedJobs := atomic.LoadInt64(&metrics.FailedJobs)
	totalJobs := atomic.LoadInt64(&metrics.TotalJobs)

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

	maxConcurrency := atomic.LoadInt64(&metrics.MaxConcurrency)
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
	t.Logf("  Duration: %v", metrics.EndTime.Sub(metrics.StartTime))
}

func TestLoadTestWithRealWorkerPool(t *testing.T) {
	logger := log.NewLogger(log.InfoLevel)

	// Create a real worker pool for more realistic testing
	pool := replication.NewWorkerPool(20, logger)
	pool.Start()
	defer pool.Stop()

	// Configuration for testing with real worker pool
	const (
		numJobs       = 100
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
	case <-time.After(30 * time.Second):
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

	if duration > 2*time.Minute { // Should complete within 2 minutes
		t.Errorf("Test took too long: %v", duration)
	}
}
