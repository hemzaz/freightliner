package replication

import (
	"sync/atomic"
	"time"
)

// WorkerPoolStats represents statistics about the worker pool
type WorkerPoolStats struct {
	TotalWorkers   int
	ActiveWorkers  int
	IdleWorkers    int
	QueuedJobs     int
	RunningJobs    int
	CompletedJobs  int64
	FailedJobs     int64
	AvgJobDuration time.Duration
	Throughput     float64 // Jobs per minute
}

// statsCollector collects statistics about worker pool operations
type statsCollector struct {
	completedJobs atomic.Int64
	failedJobs    atomic.Int64
	totalDuration atomic.Int64 // Sum of all job durations in nanoseconds
	startTime     time.Time
}

// newStatsCollector creates a new stats collector
func newStatsCollector() *statsCollector {
	return &statsCollector{
		startTime: time.Now(),
	}
}

// recordJobCompletion records a completed job
func (s *statsCollector) recordJobCompletion(duration time.Duration) {
	s.completedJobs.Add(1)
	s.totalDuration.Add(int64(duration))
}

// recordJobFailure records a failed job
func (s *statsCollector) recordJobFailure(duration time.Duration) {
	s.failedJobs.Add(1)
	s.totalDuration.Add(int64(duration))
}

// getAvgDuration returns the average job duration
func (s *statsCollector) getAvgDuration() time.Duration {
	completed := s.completedJobs.Load()
	failed := s.failedJobs.Load()
	total := completed + failed

	if total == 0 {
		return 0
	}

	totalDur := s.totalDuration.Load()
	return time.Duration(totalDur / total)
}

// getThroughput returns jobs per minute
func (s *statsCollector) getThroughput() float64 {
	elapsed := time.Since(s.startTime).Minutes()
	if elapsed == 0 {
		return 0
	}

	completed := s.completedJobs.Load()
	return float64(completed) / elapsed
}

// GetStats returns current worker pool statistics
func (p *WorkerPool) GetStats() WorkerPoolStats {
	// Count active workers
	activeWorkers := 0
	// This is a simplified implementation
	// In production, we would track this more accurately

	// Get queue sizes
	queuedJobs := len(p.jobQueue)

	// Get stats from collector
	completed := p.stats.completedJobs.Load()
	failed := p.stats.failedJobs.Load()
	avgDuration := p.stats.getAvgDuration()
	throughput := p.stats.getThroughput()

	return WorkerPoolStats{
		TotalWorkers:   p.workers,
		ActiveWorkers:  activeWorkers,
		IdleWorkers:    p.workers - activeWorkers,
		QueuedJobs:     queuedJobs,
		RunningJobs:    activeWorkers,
		CompletedJobs:  completed,
		FailedJobs:     failed,
		AvgJobDuration: avgDuration,
		Throughput:     throughput,
	}
}

// Add stats collector to WorkerPool initialization
func (p *WorkerPool) initStats() {
	if p.stats == nil {
		p.stats = newStatsCollector()
	}
}

// Add stats field to WorkerPool struct (needs to be added to worker_pool.go)
// stats *statsCollector
