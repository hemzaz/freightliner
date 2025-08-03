package server

import (
	"sync"
	"sync/atomic"
	"time"
)

// ReplicateRequest represents a request to replicate a repository
type ReplicateRequest struct {
	SourceRegistry string   `json:"source_registry"`
	SourceRepo     string   `json:"source_repo"`
	DestRegistry   string   `json:"dest_registry"`
	DestRepo       string   `json:"dest_repo"`
	Tags           []string `json:"tags,omitempty"`
	Force          bool     `json:"force"`
	DryRun         bool     `json:"dry_run"`
}

// ReplicateTreeRequest represents a request to replicate a tree of repositories
type ReplicateTreeRequest struct {
	SourceRegistry   string   `json:"source_registry"`
	SourceRepo       string   `json:"source_repo"`
	DestRegistry     string   `json:"dest_registry"`
	DestRepo         string   `json:"dest_repo"`
	ExcludeRepos     []string `json:"exclude_repos,omitempty"`
	ExcludeTags      []string `json:"exclude_tags,omitempty"`
	IncludeTags      []string `json:"include_tags,omitempty"`
	Force            bool     `json:"force"`
	DryRun           bool     `json:"dry_run"`
	EnableCheckpoint bool     `json:"enable_checkpoint"`
	CheckpointDir    string   `json:"checkpoint_dir,omitempty"`
	ResumeID         string   `json:"resume_id,omitempty"`
}

// JobResponse represents a job response
type JobResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// MetricsRegistry handles HTTP metrics recording
type MetricsRegistry struct {
	mu sync.RWMutex

	// HTTP request metrics
	httpRequests  map[string]*uint64  // key: method:route:status
	httpDurations map[string]*float64 // key: method:route
	totalRequests uint64

	// Error metrics
	panicCount   uint64
	authFailures map[string]*uint64 // key: auth_type

	// Timestamps for rate calculations
	startTime time.Time
}

// NewMetricsRegistry creates a new metrics registry
func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		httpRequests:  make(map[string]*uint64),
		httpDurations: make(map[string]*float64),
		authFailures:  make(map[string]*uint64),
		startTime:     time.Now(),
	}
}

// RecordHTTPRequest records HTTP request metrics
func (m *MetricsRegistry) RecordHTTPRequest(method, route, status string, duration float64) {
	atomic.AddUint64(&m.totalRequests, 1)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Record request count by method:route:status
	requestKey := method + ":" + route + ":" + status
	if counter, exists := m.httpRequests[requestKey]; exists {
		atomic.AddUint64(counter, 1)
	} else {
		var counter uint64 = 1
		m.httpRequests[requestKey] = &counter
	}

	// Record average duration by method:route
	durationKey := method + ":" + route
	if avgDuration, exists := m.httpDurations[durationKey]; exists {
		// Simple moving average (this is a simplified implementation)
		*avgDuration = (*avgDuration + duration) / 2
	} else {
		m.httpDurations[durationKey] = &duration
	}
}

// RecordPanic records panic metrics
func (m *MetricsRegistry) RecordPanic(component string) {
	atomic.AddUint64(&m.panicCount, 1)
}

// RecordAuthFailure records authentication failure metrics
func (m *MetricsRegistry) RecordAuthFailure(authType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if counter, exists := m.authFailures[authType]; exists {
		atomic.AddUint64(counter, 1)
	} else {
		var counter uint64 = 1
		m.authFailures[authType] = &counter
	}
}

// GetTotalRequests returns the total number of HTTP requests processed
func (m *MetricsRegistry) GetTotalRequests() uint64 {
	return atomic.LoadUint64(&m.totalRequests)
}

// GetPanicCount returns the total number of panics recorded
func (m *MetricsRegistry) GetPanicCount() uint64 {
	return atomic.LoadUint64(&m.panicCount)
}

// GetAuthFailures returns authentication failure counts by type
func (m *MetricsRegistry) GetAuthFailures() map[string]uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]uint64)
	for authType, counter := range m.authFailures {
		result[authType] = atomic.LoadUint64(counter)
	}
	return result
}
