package server

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
	// Currently a stub implementation
}

// NewMetricsRegistry creates a new metrics registry
func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{}
}

// RecordHTTPRequest records HTTP request metrics
func (m *MetricsRegistry) RecordHTTPRequest(method, route, status string, duration float64) {
	// TODO: Implement actual metrics recording
	// For now, this is a no-op to prevent compilation errors
}

// RecordPanic records panic metrics
func (m *MetricsRegistry) RecordPanic(component string) {
	// TODO: Implement actual panic metrics recording
	// For now, this is a no-op to prevent compilation errors
}

// RecordAuthFailure records authentication failure metrics
func (m *MetricsRegistry) RecordAuthFailure(authType string) {
	// TODO: Implement actual auth failure metrics recording
	// For now, this is a no-op to prevent compilation errors
}
