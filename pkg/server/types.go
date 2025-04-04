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
