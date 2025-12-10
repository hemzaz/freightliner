package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestTagCopyStatus tests the tag copy status constants
func TestTagCopyStatus(t *testing.T) {
	tests := []struct {
		name   string
		status TagCopyStatus
	}{
		{"success status", TagCopySuccess},
		{"skipped status", TagCopySkipped},
		{"failed status", TagCopyFailed},
		{"duplicate status", TagCopyDuplicate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, string(tt.status))
		})
	}
}

// TestNoopMetrics_ReplicationStarted tests noop implementation
func TestNoopMetrics_ReplicationStarted(t *testing.T) {
	metrics := NewNoopMetrics()

	// Should not panic or error
	metrics.ReplicationStarted("source", "destination")

	// Cast to verify it's the right type
	noop, ok := metrics.(*NoopMetrics)
	assert.True(t, ok)
	assert.NotNil(t, noop)
}

// TestNoopMetrics_ReplicationCompleted tests noop implementation
func TestNoopMetrics_ReplicationCompleted(t *testing.T) {
	metrics := NewNoopMetrics()

	duration := 5 * time.Second
	layerCount := 10
	byteCount := int64(1024 * 1024)

	// Should not panic or error
	metrics.ReplicationCompleted(duration, layerCount, byteCount)
}

// TestNoopMetrics_ReplicationFailed tests noop implementation
func TestNoopMetrics_ReplicationFailed(t *testing.T) {
	metrics := NewNoopMetrics()

	// Should not panic or error
	metrics.ReplicationFailed()
}

// TestNoopMetrics_TagCopyStarted tests noop implementation
func TestNoopMetrics_TagCopyStarted(t *testing.T) {
	metrics := NewNoopMetrics()

	// Should not panic or error
	metrics.TagCopyStarted("source-repo", "dest-repo", "v1.0")
}

// TestNoopMetrics_TagCopyCompleted tests noop implementation
func TestNoopMetrics_TagCopyCompleted(t *testing.T) {
	metrics := NewNoopMetrics()

	byteCount := int64(2048)

	// Should not panic or error
	metrics.TagCopyCompleted("source-repo", "dest-repo", "v1.0", byteCount)
}

// TestNoopMetrics_TagCopyFailed tests noop implementation
func TestNoopMetrics_TagCopyFailed(t *testing.T) {
	metrics := NewNoopMetrics()

	// Should not panic or error
	metrics.TagCopyFailed("source-repo", "dest-repo", "v1.0")
}

// TestNoopMetrics_RepositoryCopyCompleted tests noop implementation
func TestNoopMetrics_RepositoryCopyCompleted(t *testing.T) {
	metrics := NewNoopMetrics()

	totalTags := 100
	copiedTags := 95
	skippedTags := 3
	failedTags := 2

	// Should not panic or error
	metrics.RepositoryCopyCompleted("source-repo", "dest-repo", totalTags, copiedTags, skippedTags, failedTags)
}

// TestNoopMetrics_AllMethods tests all noop methods in sequence
func TestNoopMetrics_AllMethods(t *testing.T) {
	metrics := NewNoopMetrics()

	// Simulate a complete replication workflow
	metrics.ReplicationStarted("ecr/source", "gcr/dest")

	metrics.TagCopyStarted("source-repo", "dest-repo", "v1.0")
	metrics.TagCopyCompleted("source-repo", "dest-repo", "v1.0", 1024)

	metrics.TagCopyStarted("source-repo", "dest-repo", "v1.1")
	metrics.TagCopyFailed("source-repo", "dest-repo", "v1.1")

	metrics.RepositoryCopyCompleted("source-repo", "dest-repo", 10, 8, 1, 1)

	metrics.ReplicationCompleted(10*time.Second, 10, 10240)

	// All methods should complete without panic
	assert.NotNil(t, metrics)
}

// TestNoopMetrics_Interface tests that NoopMetrics implements MetricsCollector
func TestNoopMetrics_Interface(t *testing.T) {
	var metrics MetricsCollector = NewNoopMetrics()
	assert.NotNil(t, metrics)

	// Verify it implements all interface methods
	metrics.ReplicationStarted("test", "test")
	metrics.ReplicationCompleted(time.Second, 1, 100)
	metrics.ReplicationFailed()
	metrics.TagCopyStarted("a", "b", "c")
	metrics.TagCopyCompleted("a", "b", "c", 100)
	metrics.TagCopyFailed("a", "b", "c")
	metrics.RepositoryCopyCompleted("a", "b", 10, 8, 1, 1)
}

// TestNewNoopMetrics tests the constructor
func TestNewNoopMetrics(t *testing.T) {
	metrics := NewNoopMetrics()
	assert.NotNil(t, metrics)

	// Should return NoopMetrics type
	_, ok := metrics.(*NoopMetrics)
	assert.True(t, ok)
}

// TestMetricsCollector_ConcurrentAccess tests concurrent access to noop metrics
func TestMetricsCollector_ConcurrentAccess(t *testing.T) {
	metrics := NewNoopMetrics()

	// Run multiple goroutines concurrently
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			metrics.ReplicationStarted("source", "dest")
			metrics.TagCopyStarted("repo", "repo", "tag")
			metrics.TagCopyCompleted("repo", "repo", "tag", 1024)
			metrics.ReplicationCompleted(time.Second, 1, 1024)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic with concurrent access
	assert.NotNil(t, metrics)
}

// MockMetricsCollector is a mock implementation for testing
type MockMetricsCollector struct {
	ReplicationStartedCalls      int
	ReplicationCompletedCalls    int
	ReplicationFailedCalls       int
	TagCopyStartedCalls          int
	TagCopyCompletedCalls        int
	TagCopyFailedCalls           int
	RepositoryCopyCompletedCalls int

	LastSourceRepo  string
	LastDestRepo    string
	LastTag         string
	LastDuration    time.Duration
	LastLayerCount  int
	LastByteCount   int64
	LastTotalTags   int
	LastCopiedTags  int
	LastSkippedTags int
	LastFailedTags  int
}

func (m *MockMetricsCollector) ReplicationStarted(source, destination string) {
	m.ReplicationStartedCalls++
	m.LastSourceRepo = source
	m.LastDestRepo = destination
}

func (m *MockMetricsCollector) ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64) {
	m.ReplicationCompletedCalls++
	m.LastDuration = duration
	m.LastLayerCount = layerCount
	m.LastByteCount = byteCount
}

func (m *MockMetricsCollector) ReplicationFailed() {
	m.ReplicationFailedCalls++
}

func (m *MockMetricsCollector) TagCopyStarted(sourceRepo, destRepo, tag string) {
	m.TagCopyStartedCalls++
	m.LastSourceRepo = sourceRepo
	m.LastDestRepo = destRepo
	m.LastTag = tag
}

func (m *MockMetricsCollector) TagCopyCompleted(sourceRepo, destRepo, tag string, byteCount int64) {
	m.TagCopyCompletedCalls++
	m.LastSourceRepo = sourceRepo
	m.LastDestRepo = destRepo
	m.LastTag = tag
	m.LastByteCount = byteCount
}

func (m *MockMetricsCollector) TagCopyFailed(sourceRepo, destRepo, tag string) {
	m.TagCopyFailedCalls++
	m.LastSourceRepo = sourceRepo
	m.LastDestRepo = destRepo
	m.LastTag = tag
}

func (m *MockMetricsCollector) RepositoryCopyCompleted(sourceRepo, destRepo string, totalTags, copiedTags, skippedTags, failedTags int) {
	m.RepositoryCopyCompletedCalls++
	m.LastSourceRepo = sourceRepo
	m.LastDestRepo = destRepo
	m.LastTotalTags = totalTags
	m.LastCopiedTags = copiedTags
	m.LastSkippedTags = skippedTags
	m.LastFailedTags = failedTags
}

// TestMockMetricsCollector tests the mock implementation
func TestMockMetricsCollector(t *testing.T) {
	mock := &MockMetricsCollector{}

	// Test ReplicationStarted
	mock.ReplicationStarted("source", "dest")
	assert.Equal(t, 1, mock.ReplicationStartedCalls)
	assert.Equal(t, "source", mock.LastSourceRepo)
	assert.Equal(t, "dest", mock.LastDestRepo)

	// Test TagCopyStarted
	mock.TagCopyStarted("repo1", "repo2", "v1.0")
	assert.Equal(t, 1, mock.TagCopyStartedCalls)
	assert.Equal(t, "repo1", mock.LastSourceRepo)
	assert.Equal(t, "repo2", mock.LastDestRepo)
	assert.Equal(t, "v1.0", mock.LastTag)

	// Test TagCopyCompleted
	mock.TagCopyCompleted("repo1", "repo2", "v1.0", 2048)
	assert.Equal(t, 1, mock.TagCopyCompletedCalls)
	assert.Equal(t, int64(2048), mock.LastByteCount)

	// Test TagCopyFailed
	mock.TagCopyFailed("repo1", "repo2", "v1.1")
	assert.Equal(t, 1, mock.TagCopyFailedCalls)
	assert.Equal(t, "v1.1", mock.LastTag)

	// Test RepositoryCopyCompleted
	mock.RepositoryCopyCompleted("repo1", "repo2", 100, 95, 3, 2)
	assert.Equal(t, 1, mock.RepositoryCopyCompletedCalls)
	assert.Equal(t, 100, mock.LastTotalTags)
	assert.Equal(t, 95, mock.LastCopiedTags)
	assert.Equal(t, 3, mock.LastSkippedTags)
	assert.Equal(t, 2, mock.LastFailedTags)

	// Test ReplicationCompleted
	mock.ReplicationCompleted(5*time.Second, 10, 10240)
	assert.Equal(t, 1, mock.ReplicationCompletedCalls)
	assert.Equal(t, 5*time.Second, mock.LastDuration)
	assert.Equal(t, 10, mock.LastLayerCount)
	assert.Equal(t, int64(10240), mock.LastByteCount)

	// Test ReplicationFailed
	mock.ReplicationFailed()
	assert.Equal(t, 1, mock.ReplicationFailedCalls)
}

// TestMetricsCollector_WorkflowSimulation tests a complete workflow
func TestMetricsCollector_WorkflowSimulation(t *testing.T) {
	mock := &MockMetricsCollector{}

	// Simulate a complete replication workflow
	mock.ReplicationStarted("ecr/source-repo", "gcr/dest-repo")

	// Copy multiple tags
	tags := []string{"v1.0", "v1.1", "v1.2", "latest"}
	for _, tag := range tags {
		mock.TagCopyStarted("source-repo", "dest-repo", tag)
		if tag == "v1.1" {
			mock.TagCopyFailed("source-repo", "dest-repo", tag)
		} else {
			mock.TagCopyCompleted("source-repo", "dest-repo", tag, 1024)
		}
	}

	// Complete repository copy
	mock.RepositoryCopyCompleted("source-repo", "dest-repo", 4, 3, 0, 1)

	// Complete replication
	mock.ReplicationCompleted(30*time.Second, 3, 3072)

	// Verify call counts
	assert.Equal(t, 1, mock.ReplicationStartedCalls)
	assert.Equal(t, 4, mock.TagCopyStartedCalls)
	assert.Equal(t, 3, mock.TagCopyCompletedCalls)
	assert.Equal(t, 1, mock.TagCopyFailedCalls)
	assert.Equal(t, 1, mock.RepositoryCopyCompletedCalls)
	assert.Equal(t, 1, mock.ReplicationCompletedCalls)

	// Verify last values
	assert.Equal(t, 30*time.Second, mock.LastDuration)
	assert.Equal(t, 3, mock.LastCopiedTags)
	assert.Equal(t, 1, mock.LastFailedTags)
}

// TestMetricsCollector_MultipleRepositories tests metrics for multiple repositories
func TestMetricsCollector_MultipleRepositories(t *testing.T) {
	mock := &MockMetricsCollector{}

	repositories := []struct {
		source string
		dest   string
		tags   int
	}{
		{"repo1", "repo1-dest", 10},
		{"repo2", "repo2-dest", 15},
		{"repo3", "repo3-dest", 20},
	}

	for _, repo := range repositories {
		mock.RepositoryCopyCompleted(repo.source, repo.dest, repo.tags, repo.tags, 0, 0)
	}

	assert.Equal(t, 3, mock.RepositoryCopyCompletedCalls)
	assert.Equal(t, 20, mock.LastTotalTags) // Last call had 20 tags
}

// TestTagCopyStatus_StringValues tests status string representations
func TestTagCopyStatus_StringValues(t *testing.T) {
	tests := []struct {
		status   TagCopyStatus
		expected string
	}{
		{TagCopySuccess, "success"},
		{TagCopySkipped, "skipped"},
		{TagCopyFailed, "failed"},
		{TagCopyDuplicate, "duplicate"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

// TestMetricsCollector_LargeValues tests handling of large metric values
func TestMetricsCollector_LargeValues(t *testing.T) {
	mock := &MockMetricsCollector{}

	// Test with large byte counts (10GB)
	largeByteCount := int64(10 * 1024 * 1024 * 1024)
	mock.TagCopyCompleted("source", "dest", "tag", largeByteCount)
	assert.Equal(t, largeByteCount, mock.LastByteCount)

	// Test with large layer counts
	largeLayers := 1000
	mock.ReplicationCompleted(time.Hour, largeLayers, largeByteCount)
	assert.Equal(t, largeLayers, mock.LastLayerCount)

	// Test with many tags
	manyTags := 10000
	mock.RepositoryCopyCompleted("source", "dest", manyTags, manyTags-10, 5, 5)
	assert.Equal(t, manyTags, mock.LastTotalTags)
}
