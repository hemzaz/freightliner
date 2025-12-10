package load

import (
	"sync"
	"testing"
	"time"
)

func TestNewLoadTestMetrics(t *testing.T) {
	metrics := NewLoadTestMetrics()

	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}

	if metrics.ErrorsByType == nil {
		t.Error("Expected initialized ErrorsByType map")
	}

	if metrics.MinJobDuration != time.Hour {
		t.Errorf("Expected MinJobDuration initialized to 1 hour, got %v", metrics.MinJobDuration)
	}

	if metrics.StartTime.IsZero() {
		t.Error("Expected non-zero StartTime")
	}
}

func TestLoadTestMetrics_UpdateJobCompleted(t *testing.T) {
	metrics := NewLoadTestMetrics()

	duration := 100 * time.Millisecond
	repositories := 5

	metrics.UpdateJobCompleted(duration, repositories)

	if metrics.CompletedJobs != 1 {
		t.Errorf("Expected 1 completed job, got %d", metrics.CompletedJobs)
	}

	if metrics.TotalRepositories != 5 {
		t.Errorf("Expected 5 total repositories, got %d", metrics.TotalRepositories)
	}

	if metrics.MinJobDuration != duration {
		t.Errorf("Expected MinJobDuration %v, got %v", duration, metrics.MinJobDuration)
	}

	if metrics.MaxJobDuration != duration {
		t.Errorf("Expected MaxJobDuration %v, got %v", duration, metrics.MaxJobDuration)
	}

	// Add another job with different duration
	longerDuration := 200 * time.Millisecond
	metrics.UpdateJobCompleted(longerDuration, 3)

	if metrics.CompletedJobs != 2 {
		t.Errorf("Expected 2 completed jobs, got %d", metrics.CompletedJobs)
	}

	if metrics.MinJobDuration != duration {
		t.Errorf("Expected MinJobDuration %v, got %v", duration, metrics.MinJobDuration)
	}

	if metrics.MaxJobDuration != longerDuration {
		t.Errorf("Expected MaxJobDuration %v, got %v", longerDuration, metrics.MaxJobDuration)
	}
}

func TestLoadTestMetrics_UpdateJobFailed(t *testing.T) {
	metrics := NewLoadTestMetrics()

	metrics.UpdateJobFailed("network_error")
	metrics.UpdateJobFailed("network_error")
	metrics.UpdateJobFailed("auth_error")

	if metrics.FailedJobs != 3 {
		t.Errorf("Expected 3 failed jobs, got %d", metrics.FailedJobs)
	}

	if metrics.ErrorsByType["network_error"] != 2 {
		t.Errorf("Expected 2 network errors, got %d", metrics.ErrorsByType["network_error"])
	}

	if metrics.ErrorsByType["auth_error"] != 1 {
		t.Errorf("Expected 1 auth error, got %d", metrics.ErrorsByType["auth_error"])
	}
}

func TestLoadTestMetrics_UpdateConcurrency(t *testing.T) {
	metrics := NewLoadTestMetrics()

	metrics.UpdateConcurrency(10)
	if metrics.CurrentConcurrency != 10 {
		t.Errorf("Expected CurrentConcurrency 10, got %d", metrics.CurrentConcurrency)
	}
	if metrics.MaxConcurrency != 10 {
		t.Errorf("Expected MaxConcurrency 10, got %d", metrics.MaxConcurrency)
	}

	// Update to higher value
	metrics.UpdateConcurrency(25)
	if metrics.CurrentConcurrency != 25 {
		t.Errorf("Expected CurrentConcurrency 25, got %d", metrics.CurrentConcurrency)
	}
	if metrics.MaxConcurrency != 25 {
		t.Errorf("Expected MaxConcurrency 25, got %d", metrics.MaxConcurrency)
	}

	// Update to lower value (max should not decrease)
	metrics.UpdateConcurrency(15)
	if metrics.CurrentConcurrency != 15 {
		t.Errorf("Expected CurrentConcurrency 15, got %d", metrics.CurrentConcurrency)
	}
	if metrics.MaxConcurrency != 25 {
		t.Errorf("Expected MaxConcurrency to remain 25, got %d", metrics.MaxConcurrency)
	}
}

func TestLoadTestMetrics_GetSummary(t *testing.T) {
	metrics := NewLoadTestMetrics()

	// Add some data
	metrics.UpdateJobCompleted(100*time.Millisecond, 5)
	metrics.UpdateJobCompleted(200*time.Millisecond, 3)
	metrics.UpdateJobFailed("network_error")
	metrics.UpdateConcurrency(10)
	metrics.TotalJobs = 3
	metrics.EndTime = time.Now()

	summary := metrics.GetSummary()

	if summary.TotalJobs != 3 {
		t.Errorf("Expected 3 total jobs, got %d", summary.TotalJobs)
	}

	if summary.CompletedJobs != 2 {
		t.Errorf("Expected 2 completed jobs, got %d", summary.CompletedJobs)
	}

	if summary.FailedJobs != 1 {
		t.Errorf("Expected 1 failed job, got %d", summary.FailedJobs)
	}

	if summary.TotalRepositories != 8 {
		t.Errorf("Expected 8 total repositories, got %d", summary.TotalRepositories)
	}

	if summary.MaxConcurrency != 10 {
		t.Errorf("Expected MaxConcurrency 10, got %d", summary.MaxConcurrency)
	}

	if summary.MinJobDuration != 100*time.Millisecond {
		t.Errorf("Expected MinJobDuration 100ms, got %v", summary.MinJobDuration)
	}

	if summary.MaxJobDuration != 200*time.Millisecond {
		t.Errorf("Expected MaxJobDuration 200ms, got %v", summary.MaxJobDuration)
	}

	if summary.AvgJobDuration != 150*time.Millisecond {
		t.Errorf("Expected AvgJobDuration 150ms, got %v", summary.AvgJobDuration)
	}

	if summary.ErrorsByType["network_error"] != 1 {
		t.Errorf("Expected 1 network error in summary, got %d", summary.ErrorsByType["network_error"])
	}
}

func TestLoadTestMetrics_ConcurrentUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	metrics := NewLoadTestMetrics()

	const numGoroutines = 100
	const updatesPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent job completions
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < updatesPerGoroutine; j++ {
				metrics.UpdateJobCompleted(50*time.Millisecond, 1)
			}
		}()
	}

	wg.Wait()

	expectedJobs := int64(numGoroutines * updatesPerGoroutine)
	if metrics.CompletedJobs != expectedJobs {
		t.Errorf("Expected %d completed jobs, got %d", expectedJobs, metrics.CompletedJobs)
	}

	if metrics.TotalRepositories != expectedJobs {
		t.Errorf("Expected %d total repositories, got %d", expectedJobs, metrics.TotalRepositories)
	}
}

func TestLoadTestMetrics_ConcurrentFailures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	metrics := NewLoadTestMetrics()

	const numGoroutines = 50
	const failuresPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent failures
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			errorType := "error_type_" + string(rune('A'+id%5))
			for j := 0; j < failuresPerGoroutine; j++ {
				metrics.UpdateJobFailed(errorType)
			}
		}(i)
	}

	wg.Wait()

	expectedFailures := int64(numGoroutines * failuresPerGoroutine)
	if metrics.FailedJobs != expectedFailures {
		t.Errorf("Expected %d failed jobs, got %d", expectedFailures, metrics.FailedJobs)
	}

	// Verify all error types were recorded
	totalErrors := int64(0)
	for _, count := range metrics.ErrorsByType {
		totalErrors += count
	}

	if totalErrors != expectedFailures {
		t.Errorf("Expected %d total errors, got %d", expectedFailures, totalErrors)
	}
}

func TestLoadTestSummary_SuccessRate(t *testing.T) {
	tests := []struct {
		name          string
		totalJobs     int64
		completedJobs int64
		want          float64
	}{
		{"No jobs", 0, 0, 0},
		{"All success", 100, 100, 100.0},
		{"Half success", 100, 50, 50.0},
		{"25% success", 100, 25, 25.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := LoadTestSummary{
				TotalJobs:     tt.totalJobs,
				CompletedJobs: tt.completedJobs,
			}

			got := summary.SuccessRate()
			if got != tt.want {
				t.Errorf("SuccessRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadTestSummary_ErrorRate(t *testing.T) {
	tests := []struct {
		name       string
		totalJobs  int64
		failedJobs int64
		want       float64
	}{
		{"No jobs", 0, 0, 0},
		{"No errors", 100, 0, 0.0},
		{"10% errors", 100, 10, 10.0},
		{"50% errors", 100, 50, 50.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := LoadTestSummary{
				TotalJobs:  tt.totalJobs,
				FailedJobs: tt.failedJobs,
			}

			got := summary.ErrorRate()
			if got != tt.want {
				t.Errorf("ErrorRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadTestSummary_ThroughputReposPerSecond(t *testing.T) {
	tests := []struct {
		name              string
		totalRepositories int64
		duration          time.Duration
		want              float64
	}{
		{"Zero duration", 100, 0, 0},
		{"1 second", 100, 1 * time.Second, 100.0},
		{"10 seconds", 100, 10 * time.Second, 10.0},
		{"1 minute", 600, 1 * time.Minute, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := LoadTestSummary{
				TotalRepositories: tt.totalRepositories,
				Duration:          tt.duration,
			}

			got := summary.ThroughputReposPerSecond()
			if got != tt.want {
				t.Errorf("ThroughputReposPerSecond() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadTestMetrics_ZeroDuration(t *testing.T) {
	metrics := NewLoadTestMetrics()
	metrics.EndTime = metrics.StartTime // Zero duration

	summary := metrics.GetSummary()

	if summary.Duration != 0 {
		t.Errorf("Expected zero duration, got %v", summary.Duration)
	}
}

func TestLoadTestMetrics_NoEndTime(t *testing.T) {
	metrics := NewLoadTestMetrics()
	time.Sleep(100 * time.Millisecond)

	summary := metrics.GetSummary()

	if summary.Duration <= 0 {
		t.Error("Expected positive duration when EndTime is not set")
	}

	if summary.Duration < 100*time.Millisecond {
		t.Errorf("Expected duration >= 100ms, got %v", summary.Duration)
	}
}

func TestLoadTestMetrics_MultipleErrorTypes(t *testing.T) {
	metrics := NewLoadTestMetrics()

	errorTypes := []string{
		"network_error",
		"auth_error",
		"timeout_error",
		"manifest_error",
		"blob_error",
	}

	for i := 0; i < 100; i++ {
		errorType := errorTypes[i%len(errorTypes)]
		metrics.UpdateJobFailed(errorType)
	}

	summary := metrics.GetSummary()

	for _, errorType := range errorTypes {
		if summary.ErrorsByType[errorType] != 20 {
			t.Errorf("Expected 20 errors of type %s, got %d", errorType, summary.ErrorsByType[errorType])
		}
	}
}

func TestLoadTestMetrics_PeakMemoryAndGoroutines(t *testing.T) {
	metrics := NewLoadTestMetrics()

	metrics.PeakMemoryMB = 512
	metrics.PeakGoroutines = 100

	summary := metrics.GetSummary()

	if summary.PeakMemoryMB != 512 {
		t.Errorf("Expected PeakMemoryMB 512, got %d", summary.PeakMemoryMB)
	}

	if summary.PeakGoroutines != 100 {
		t.Errorf("Expected PeakGoroutines 100, got %d", summary.PeakGoroutines)
	}
}
