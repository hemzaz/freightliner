package metrics

import "time"

// TagCopyStatus represents the status of a tag copy operation
type TagCopyStatus string

// Copy status constants
const (
	TagCopySuccess   TagCopyStatus = "success"
	TagCopySkipped   TagCopyStatus = "skipped"
	TagCopyFailed    TagCopyStatus = "failed"
	TagCopyDuplicate TagCopyStatus = "duplicate"
)

// MetricsCollector is an interface for collecting metrics about image replication
type MetricsCollector interface {
	// ReplicationStarted records the start of a replication operation
	ReplicationStarted(source, destination string)

	// ReplicationCompleted records the completion of a replication operation
	ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64)

	// ReplicationFailed records a failed replication operation
	ReplicationFailed()

	// TagCopyStarted records the start of copying a specific tag
	TagCopyStarted(sourceRepo, destRepo, tag string)

	// TagCopyCompleted records the completion of copying a specific tag
	TagCopyCompleted(sourceRepo, destRepo, tag string, byteCount int64)

	// TagCopyFailed records a failure to copy a specific tag
	TagCopyFailed(sourceRepo, destRepo, tag string)

	// RepositoryCopyCompleted records the completion of copying an entire repository
	RepositoryCopyCompleted(sourceRepo, destRepo string, totalTags, copiedTags, skippedTags, failedTags int)
}

// NoopMetrics is a no-op implementation of the MetricsCollector interface
type NoopMetrics struct{}

// ReplicationStarted is a no-op implementation
func (n *NoopMetrics) ReplicationStarted(source, destination string) {}

// ReplicationCompleted is a no-op implementation
func (n *NoopMetrics) ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64) {}

// ReplicationFailed is a no-op implementation
func (n *NoopMetrics) ReplicationFailed() {}

// TagCopyStarted is a no-op implementation
func (n *NoopMetrics) TagCopyStarted(sourceRepo, destRepo, tag string) {}

// TagCopyCompleted is a no-op implementation
func (n *NoopMetrics) TagCopyCompleted(sourceRepo, destRepo, tag string, byteCount int64) {}

// TagCopyFailed is a no-op implementation
func (n *NoopMetrics) TagCopyFailed(sourceRepo, destRepo, tag string) {}

// RepositoryCopyCompleted is a no-op implementation
func (n *NoopMetrics) RepositoryCopyCompleted(sourceRepo, destRepo string, totalTags, copiedTags, skippedTags, failedTags int) {
}

// NewNoopMetrics returns a new instance of NoopMetrics
func NewNoopMetrics() MetricsCollector {
	return &NoopMetrics{}
}
