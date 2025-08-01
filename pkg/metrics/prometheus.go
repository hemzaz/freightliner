package metrics

import (
	"sync"
	"time"
)

// PrometheusMetrics provides a metrics collector that can be used with Prometheus
// This is a simpler version that doesn't actually depend on Prometheus client libraries
// so that we can avoid adding extra dependencies. In a real implementation,
// this would use the Prometheus client libraries directly.
type PrometheusMetrics struct {
	// Use a mutex to protect the metrics
	mutex sync.Mutex

	// Counters for replication operations
	replicationCount        int64
	replicationErrors       int64
	layersCopied            int64
	bytesCopied             int64
	replicationLatencies    []time.Duration
	sourceRepositories      map[string]int64
	destinationRepositories map[string]int64
}

// NewPrometheusMetrics creates a new metrics collector
func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		sourceRepositories:      make(map[string]int64),
		destinationRepositories: make(map[string]int64),
	}
}

// ReplicationStarted records the start of a replication operation
func (p *PrometheusMetrics) ReplicationStarted(source, destination string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.replicationCount++
	p.sourceRepositories[source]++
	p.destinationRepositories[destination]++
}

// ReplicationCompleted records the completion of a replication operation
func (p *PrometheusMetrics) ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.replicationLatencies = append(p.replicationLatencies, duration)
	p.layersCopied += int64(layerCount)
	p.bytesCopied += byteCount
}

// ReplicationFailed records a failed replication operation
func (p *PrometheusMetrics) ReplicationFailed() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.replicationErrors++
}

// GetReplicationCount returns the total number of replication operations
func (p *PrometheusMetrics) GetReplicationCount() int64 {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.replicationCount
}

// GetReplicationErrors returns the total number of failed replication operations
func (p *PrometheusMetrics) GetReplicationErrors() int64 {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.replicationErrors
}

// GetLayersCopied returns the total number of layers copied
func (p *PrometheusMetrics) GetLayersCopied() int64 {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.layersCopied
}

// GetBytesCopied returns the total number of bytes copied
func (p *PrometheusMetrics) GetBytesCopied() int64 {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.bytesCopied
}

// GetAverageLatency returns the average latency of replication operations
func (p *PrometheusMetrics) GetAverageLatency() time.Duration {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(p.replicationLatencies) == 0 {
		return 0
	}

	var sum time.Duration
	for _, latency := range p.replicationLatencies {
		sum += latency
	}

	return sum / time.Duration(len(p.replicationLatencies))
}

// GetTopSourceRepositories returns the top source repositories by replication count
func (p *PrometheusMetrics) GetTopSourceRepositories(n int) map[string]int64 {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Make a copy to avoid concurrent map access
	result := make(map[string]int64, len(p.sourceRepositories))
	for repo, count := range p.sourceRepositories {
		result[repo] = count
	}

	return result
}

// GetTopDestinationRepositories returns the top destination repositories by replication count
func (p *PrometheusMetrics) GetTopDestinationRepositories(n int) map[string]int64 {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Make a copy to avoid concurrent map access
	result := make(map[string]int64, len(p.destinationRepositories))
	for repo, count := range p.destinationRepositories {
		result[repo] = count
	}

	return result
}

// TagCopyStarted records the start of copying a specific tag
func (p *PrometheusMetrics) TagCopyStarted(sourceRepo, destRepo, tag string) {
	// No-op for now - could track tag-level metrics in the future
}

// TagCopyCompleted records the completion of copying a specific tag
func (p *PrometheusMetrics) TagCopyCompleted(sourceRepo, destRepo, tag string, byteCount int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.bytesCopied += byteCount
}

// TagCopyFailed records a failure to copy a specific tag
func (p *PrometheusMetrics) TagCopyFailed(sourceRepo, destRepo, tag string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.replicationErrors++
}

// RepositoryCopyCompleted records the completion of copying an entire repository
func (p *PrometheusMetrics) RepositoryCopyCompleted(sourceRepo, destRepo string, totalTags, copiedTags, skippedTags, failedTags int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.sourceRepositories[sourceRepo]++
	p.destinationRepositories[destRepo]++
}
