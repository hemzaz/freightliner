package load

import (
	"sync"
	"sync/atomic"
	"time"
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
	StartTime        time.Time
	EndTime          time.Time
	MinJobDuration   time.Duration
	MaxJobDuration   time.Duration
	TotalJobTimeNano int64 // Total job time in nanoseconds for atomic operations

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
	atomic.AddInt64(&m.TotalJobTimeNano, int64(duration))

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
		oldMax := atomic.LoadInt64(&m.MaxConcurrency)
		if current <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt64(&m.MaxConcurrency, oldMax, current) {
			break
		}
	}
}

// GetSummary returns a summary of the metrics
func (m *LoadTestMetrics) GetSummary() LoadTestSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := LoadTestSummary{
		TotalJobs:          atomic.LoadInt64(&m.TotalJobs),
		CompletedJobs:      atomic.LoadInt64(&m.CompletedJobs),
		FailedJobs:         atomic.LoadInt64(&m.FailedJobs),
		TotalRepositories:  atomic.LoadInt64(&m.TotalRepositories),
		StartTime:          m.StartTime,
		EndTime:            m.EndTime,
		MinJobDuration:     m.MinJobDuration,
		MaxJobDuration:     m.MaxJobDuration,
		CurrentConcurrency: atomic.LoadInt64(&m.CurrentConcurrency),
		MaxConcurrency:     atomic.LoadInt64(&m.MaxConcurrency),
		ErrorsByType:       make(map[string]int64),
		PeakMemoryMB:       atomic.LoadInt64(&m.PeakMemoryMB),
		PeakGoroutines:     atomic.LoadInt64(&m.PeakGoroutines),
	}

	// Copy error map
	for k, v := range m.ErrorsByType {
		summary.ErrorsByType[k] = v
	}

	// Calculate derived metrics
	if summary.CompletedJobs > 0 {
		avgJobTime := time.Duration(atomic.LoadInt64(&m.TotalJobTimeNano) / summary.CompletedJobs)
		summary.AvgJobDuration = avgJobTime
	}

	if summary.EndTime.IsZero() {
		summary.Duration = time.Since(m.StartTime)
	} else {
		summary.Duration = summary.EndTime.Sub(m.StartTime)
	}

	return summary
}

// LoadTestSummary provides a snapshot of load test metrics
type LoadTestSummary struct {
	TotalJobs          int64
	CompletedJobs      int64
	FailedJobs         int64
	TotalRepositories  int64
	StartTime          time.Time
	EndTime            time.Time
	Duration           time.Duration
	MinJobDuration     time.Duration
	MaxJobDuration     time.Duration
	AvgJobDuration     time.Duration
	CurrentConcurrency int64
	MaxConcurrency     int64
	ErrorsByType       map[string]int64
	PeakMemoryMB       int64
	PeakGoroutines     int64
}

// SuccessRate returns the success rate as a percentage
func (s LoadTestSummary) SuccessRate() float64 {
	if s.TotalJobs == 0 {
		return 0
	}
	return float64(s.CompletedJobs) / float64(s.TotalJobs) * 100
}

// ErrorRate returns the error rate as a percentage
func (s LoadTestSummary) ErrorRate() float64 {
	if s.TotalJobs == 0 {
		return 0
	}
	return float64(s.FailedJobs) / float64(s.TotalJobs) * 100
}

// ThroughputReposPerSecond returns repositories processed per second
func (s LoadTestSummary) ThroughputReposPerSecond() float64 {
	if s.Duration.Seconds() == 0 {
		return 0
	}
	return float64(s.TotalRepositories) / s.Duration.Seconds()
}
