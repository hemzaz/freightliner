package checkpoint

import (
	"time"

	"freightliner/pkg/helper/errors"
)

// ResumableCheckpoint contains information about a checkpoint that can be resumed
type ResumableCheckpoint struct {
	// ID is the checkpoint identifier
	ID string

	// SourceRegistry is the source registry name
	SourceRegistry string

	// SourcePrefix is the source prefix
	SourcePrefix string

	// DestRegistry is the destination registry name
	DestRegistry string

	// DestPrefix is the destination prefix
	DestPrefix string

	// Status is the current status
	Status Status

	// Progress is the progress percentage (0-100)
	Progress float64

	// LastUpdated is when the checkpoint was last updated
	LastUpdated time.Time

	// TotalRepositories is the total repositories in the replication
	TotalRepositories int

	// CompletedRepositories is the number of completed repositories
	CompletedRepositories int

	// FailedRepositories is the number of failed repositories
	FailedRepositories int

	// Duration is how long the replication has been running
	Duration time.Duration
}

// ResumableOptions contains options for resuming a replication
type ResumableOptions struct {
	// ID is the checkpoint ID to resume
	ID string

	// SkipCompleted skips repositories that have already been completed
	SkipCompleted bool

	// RetryFailed retries repositories that previously failed
	RetryFailed bool

	// Force forces overwriting existing tags
	Force bool
}

// GetResumableCheckpoints returns a list of checkpoints that can be resumed
func GetResumableCheckpoints(store CheckpointStore) ([]ResumableCheckpoint, error) {
	if store == nil {
		return nil, errors.InvalidInputf("checkpoint store cannot be nil")
	}

	checkpoints, err := store.ListCheckpoints()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list checkpoints")
	}

	var resumable []ResumableCheckpoint

	for _, cp := range checkpoints {
		// Only include checkpoints that can be resumed
		if cp.Status == StatusPending || cp.Status == StatusInterrupted || cp.Status == StatusFailed || cp.Status == StatusInProgress {
			// Count completed and failed repositories
			var completed, failed int

			// Count directly from Repositories map
			for _, repoStatus := range cp.Repositories {
				if repoStatus.Status == StatusCompleted {
					completed++
				} else if repoStatus.Status == StatusFailed {
					failed++
				}
			}

			// Create a resumable checkpoint
			resumable = append(resumable, ResumableCheckpoint{
				ID:                    cp.ID,
				SourceRegistry:        cp.SourceRegistry,
				SourcePrefix:          cp.SourcePrefix,
				DestRegistry:          cp.DestRegistry,
				DestPrefix:            cp.DestPrefix,
				Status:                cp.Status,
				Progress:              cp.Progress,
				LastUpdated:           cp.LastUpdated,
				TotalRepositories:     len(cp.Repositories),
				CompletedRepositories: completed,
				FailedRepositories:    failed,
				Duration:              cp.LastUpdated.Sub(cp.StartTime),
			})
		}
	}

	return resumable, nil
}

// GetCheckpointByID retrieves a specific checkpoint by ID
func GetCheckpointByID(store CheckpointStore, id string) (*TreeCheckpoint, error) {
	if store == nil {
		return nil, errors.InvalidInputf("checkpoint store cannot be nil")
	}

	if id == "" {
		return nil, errors.InvalidInputf("checkpoint ID cannot be empty")
	}

	checkpoint, err := store.LoadCheckpoint(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load checkpoint")
	}

	if checkpoint == nil {
		return nil, errors.NotFoundf("checkpoint not found with ID: %s", id)
	}

	return checkpoint, nil
}

// GetRemainingRepositories returns repositories that still need to be processed
func GetRemainingRepositories(cp *TreeCheckpoint, opts ResumableOptions) ([]string, error) {
	if cp == nil {
		return nil, errors.InvalidInputf("checkpoint cannot be nil")
	}

	var remaining []string

	// Check RepoTasks for repository status
	for _, repoTask := range cp.RepoTasks {
		// Skip completed repositories if requested
		if opts.SkipCompleted && repoTask.Status == StatusCompleted {
			continue
		}

		// Skip failed repositories if not retrying
		if !opts.RetryFailed && repoTask.Status == StatusFailed {
			continue
		}

		// Include repositories that are not completed or if we don't skip completed
		// Include failed repositories if we retry failed
		shouldInclude := false

		if repoTask.Status == StatusCompleted && !opts.SkipCompleted {
			shouldInclude = true
		} else if repoTask.Status == StatusFailed && opts.RetryFailed {
			shouldInclude = true
		} else if repoTask.Status != StatusCompleted && repoTask.Status != StatusFailed {
			// Include pending, interrupted, in-progress, etc.
			shouldInclude = true
		}

		if shouldInclude {
			remaining = append(remaining, repoTask.SourceRepository)
		}
	}

	return remaining, nil
}
