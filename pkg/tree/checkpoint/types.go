package checkpoint

import (
	"time"
)

// Status represents the status of a replication task
type Status string

// CheckpointStatus is an alias for Status for backward compatibility
type CheckpointStatus = Status

const (
	// StatusPending indicates a task is pending
	StatusPending Status = "pending"

	// StatusInProgress indicates a task is in progress
	StatusInProgress Status = "in_progress"

	// StatusCompleted indicates a task has completed successfully
	StatusCompleted Status = "completed"

	// StatusFailed indicates a task has failed
	StatusFailed Status = "failed"

	// StatusInterrupted indicates a task was interrupted
	StatusInterrupted Status = "interrupted"

	// StatusSkipped indicates a task was skipped (already exists)
	StatusSkipped Status = "skipped"
)

// RepoTask represents a repository replication task
type RepoTask struct {
	// SourceRegistry is the source registry name
	SourceRegistry string `json:"source_registry"`

	// SourceRepository is the source repository name
	SourceRepository string `json:"source_repository"`

	// DestRegistry is the destination registry name
	DestRegistry string `json:"dest_registry"`

	// DestRepository is the destination repository name
	DestRepository string `json:"dest_repository"`

	// Status is the current status of this task
	Status Status `json:"status"`

	// LastUpdated is when this task was last updated
	LastUpdated time.Time `json:"last_updated"`

	// Error is the error message if status is failed
	Error string `json:"error,omitempty"`

	// TagTasks is a list of tag tasks within this repository
	TagTasks []TagTask `json:"tag_tasks,omitempty"`
}

// TagTask represents a tag replication task
type TagTask struct {
	// SourceTag is the source tag to replicate
	SourceTag string `json:"source_tag"`

	// DestTag is the destination tag
	DestTag string `json:"dest_tag"`

	// Status is the current status of this task
	Status Status `json:"status"`

	// ManifestDigest is the digest of the manifest
	ManifestDigest string `json:"manifest_digest,omitempty"`

	// LastUpdated is when this task was last updated
	LastUpdated time.Time `json:"last_updated"`

	// Error is the error message if status is failed
	Error string `json:"error,omitempty"`

	// LayerTasks tracks each layer's replication status
	LayerTasks []LayerTask `json:"layer_tasks,omitempty"`
}

// LayerTask represents a layer replication task
type LayerTask struct {
	// Digest is the layer digest
	Digest string `json:"digest"`

	// MediaType is the layer media type
	MediaType string `json:"media_type"`

	// Status is the current status of this task
	Status Status `json:"status"`

	// LastUpdated is when this task was last updated
	LastUpdated time.Time `json:"last_updated"`

	// Size is the size of the layer in bytes
	Size int64 `json:"size,omitempty"`

	// Error is the error message if status is failed
	Error string `json:"error,omitempty"`
}

// RepoStatus represents the status of a repository in a tree replication
type RepoStatus struct {
	// Status is the current status of this repository replication
	Status Status `json:"status"`

	// SourceRepo is the source repository name
	SourceRepo string `json:"source_repo"`

	// DestRepo is the destination repository name
	DestRepo string `json:"dest_repo"`

	// LastUpdated is when this repo status was last updated
	LastUpdated time.Time `json:"last_updated"`

	// Error is the error message if status is failed
	Error string `json:"error,omitempty"`
}

type TreeCheckpoint struct {
	// ID is a unique identifier for this replication run
	ID string `json:"id"`

	// StartTime is when the replication started
	StartTime time.Time `json:"start_time"`

	// LastUpdated is when the checkpoint was last updated
	LastUpdated time.Time `json:"last_updated"`

	// SourceRegistry is the source registry name
	SourceRegistry string `json:"source_registry"`

	// SourcePrefix is the source prefix
	SourcePrefix string `json:"source_prefix"`

	// DestRegistry is the destination registry name
	DestRegistry string `json:"dest_registry"`

	// DestPrefix is the destination prefix
	DestPrefix string `json:"dest_prefix"`

	// Status is the overall status of the tree replication
	Status Status `json:"status"`

	// LastError is the last error that occurred
	LastError string `json:"last_error,omitempty"`

	// RepoTasks tracks the repository replication tasks
	RepoTasks []RepoTask `json:"repo_tasks"`

	// Repositories is a map of repository replication statuses
	Repositories map[string]RepoStatus `json:"repositories"`

	// CompletedRepositories is a list of completed repositories
	CompletedRepositories []string `json:"completed_repositories"`

	// Progress indicates overall progress as a percentage (0-100)
	Progress float64 `json:"progress"`
}

// CheckpointStore defines the interface for checkpoint storage
type CheckpointStore interface {
	// SaveCheckpoint saves a checkpoint to storage
	SaveCheckpoint(checkpoint *TreeCheckpoint) error

	// LoadCheckpoint loads a checkpoint from storage by ID
	LoadCheckpoint(id string) (*TreeCheckpoint, error)

	// ListCheckpoints lists all checkpoints
	ListCheckpoints() ([]*TreeCheckpoint, error)

	// DeleteCheckpoint deletes a checkpoint by ID
	DeleteCheckpoint(id string) error

	// CheckpointExists checks if a checkpoint with the given ID exists
	CheckpointExists(id string) (bool, error)
}
