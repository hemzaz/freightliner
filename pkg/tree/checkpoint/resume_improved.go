package checkpoint

import (
	"time"

	"freightliner/pkg/helper/errors"
)

// GetResumableCheckpointsImproved returns a list of checkpoints that can be resumed with improved filtering
func GetResumableCheckpointsImproved(store CheckpointStore) ([]ResumableCheckpoint, error) {
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
		// Exclude completed checkpoints and include pending ones
		if isResumableStatus(cp.Status) {
			// Count completed and failed repositories more accurately
			completed, failed, total := countRepositoryStatuses(cp)

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
				TotalRepositories:     total,
				CompletedRepositories: completed,
				FailedRepositories:    failed,
				Duration:              calculateDuration(cp.StartTime, cp.LastUpdated),
			})
		}
	}

	return resumable, nil
}

// isResumableStatus determines if a checkpoint status allows for resumption
func isResumableStatus(status CheckpointStatus) bool {
	switch status {
	case StatusPending, StatusInProgress, StatusInterrupted, StatusFailed:
		return true
	case StatusCompleted:
		return false
	default:
		// Unknown statuses are considered resumable to be safe
		return true
	}
}

// countRepositoryStatuses accurately counts repository statuses from multiple sources
func countRepositoryStatuses(cp *TreeCheckpoint) (completed, failed, total int) {
	// Use a map to track all unique repositories
	repoStatuses := make(map[string]CheckpointStatus)

	// First, collect from the Repositories map (most accurate)
	for repoName, repoStatus := range cp.Repositories {
		repoStatuses[repoName] = repoStatus.Status
	}

	// Also check RepoTasks for any additional repositories
	for _, task := range cp.RepoTasks {
		if task.SourceRepository != "" {
			// Only override if we don't already have this repo or if task status is more recent
			if _, exists := repoStatuses[task.SourceRepository]; !exists {
				repoStatuses[task.SourceRepository] = task.Status
			}
		}
	}

	// Count statuses
	for _, status := range repoStatuses {
		switch status {
		case StatusCompleted:
			completed++
		case StatusFailed:
			failed++
		}
		total++
	}

	return completed, failed, total
}

// calculateDuration safely calculates duration, handling zero times
func calculateDuration(startTime, lastUpdated time.Time) time.Duration {
	if startTime.IsZero() || lastUpdated.IsZero() {
		return 0
	}
	if lastUpdated.Before(startTime) {
		return 0
	}
	return lastUpdated.Sub(startTime)
}

// GetRemainingRepositoriesImproved returns repositories that still need to be processed with improved logic
func GetRemainingRepositoriesImproved(cp *TreeCheckpoint, opts ResumableOptions) ([]string, error) {
	if cp == nil {
		return nil, errors.InvalidInputf("checkpoint cannot be nil")
	}

	if opts.ID == "" {
		return nil, errors.InvalidInputf("resume options ID cannot be empty")
	}

	var remaining []string

	// Build a comprehensive view of repository statuses
	repoStatuses := buildRepositoryStatusMap(cp)

	// Process each repository according to the options
	for repoName, status := range repoStatuses {
		shouldInclude := shouldIncludeRepository(status, opts)
		if shouldInclude {
			remaining = append(remaining, repoName)
		}
	}

	return remaining, nil
}

// buildRepositoryStatusMap creates a comprehensive map of repository statuses
func buildRepositoryStatusMap(cp *TreeCheckpoint) map[string]CheckpointStatus {
	repoStatuses := make(map[string]CheckpointStatus)

	// Start with the Repositories map (most authoritative)
	for repoName, repoStatus := range cp.Repositories {
		repoStatuses[repoName] = repoStatus.Status
	}

	// Cross-reference with RepoTasks for any missing repositories
	for _, task := range cp.RepoTasks {
		if task.SourceRepository != "" {
			// Only add if not already present (Repositories map takes precedence)
			if _, exists := repoStatuses[task.SourceRepository]; !exists {
				repoStatuses[task.SourceRepository] = task.Status
			}
		}
	}

	// Also consider repositories from CompletedRepositories list
	// These should be marked as completed if not already tracked
	for _, repoName := range cp.CompletedRepositories {
		if _, exists := repoStatuses[repoName]; !exists {
			repoStatuses[repoName] = StatusCompleted
		}
	}

	return repoStatuses
}

// shouldIncludeRepository determines if a repository should be included based on its status and options
func shouldIncludeRepository(status CheckpointStatus, opts ResumableOptions) bool {
	switch status {
	case StatusCompleted:
		// Include completed repositories only if SkipCompleted is false
		return !opts.SkipCompleted

	case StatusFailed:
		// Include failed repositories only if RetryFailed is true
		return opts.RetryFailed

	case StatusPending, StatusInProgress, StatusInterrupted:
		// Always include these statuses as they need to be processed
		return true

	default:
		// For unknown statuses, include them to be safe
		return true
	}
}

// ValidateResumableOptions validates the resumable options
func ValidateResumableOptions(opts ResumableOptions) error {
	if opts.ID == "" {
		return errors.InvalidInputf("resume options ID cannot be empty")
	}

	// Additional validation can be added here
	return nil
}

// GetResumableCheckpointSummary provides a summary of a resumable checkpoint
func GetResumableCheckpointSummary(cp *TreeCheckpoint) ResumableCheckpoint {
	completed, failed, total := countRepositoryStatuses(cp)

	return ResumableCheckpoint{
		ID:                    cp.ID,
		SourceRegistry:        cp.SourceRegistry,
		SourcePrefix:          cp.SourcePrefix,
		DestRegistry:          cp.DestRegistry,
		DestPrefix:            cp.DestPrefix,
		Status:                cp.Status,
		Progress:              cp.Progress,
		LastUpdated:           cp.LastUpdated,
		TotalRepositories:     total,
		CompletedRepositories: completed,
		FailedRepositories:    failed,
		Duration:              calculateDuration(cp.StartTime, cp.LastUpdated),
	}
}
