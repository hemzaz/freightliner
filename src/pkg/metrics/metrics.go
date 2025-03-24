package metrics

import "time"

// Metrics is an interface for collecting metrics about image replication
type Metrics interface {
	// ReplicationStarted records the start of a replication operation
	ReplicationStarted(source, destination string)
	
	// ReplicationCompleted records the completion of a replication operation
	ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64)
	
	// ReplicationFailed records a failed replication operation
	ReplicationFailed()
}

// NoopMetrics is a no-op implementation of the Metrics interface
type NoopMetrics struct{}

// ReplicationStarted is a no-op implementation
func (n *NoopMetrics) ReplicationStarted(source, destination string) {}

// ReplicationCompleted is a no-op implementation
func (n *NoopMetrics) ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64) {}

// ReplicationFailed is a no-op implementation
func (n *NoopMetrics) ReplicationFailed() {}