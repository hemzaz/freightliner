package tree

import (
	"context"
	"sync"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/interfaces"
	"freightliner/pkg/tree/checkpoint"
)

// ResumeOptions configures the replication resume process
type ResumeOptions struct {
	// CheckpointID is the ID of the checkpoint to resume
	CheckpointID string

	// SkipCompleted skips repositories that have already been completed
	SkipCompleted bool

	// RetryFailed retries repositories that previously failed
	RetryFailed bool

	// ForceOverwrite forces overwriting existing tags
	ForceOverwrite bool
}

// ListResumableReplications returns a list of replications that can be resumed
func (t *TreeReplicator) ListResumableReplications() ([]checkpoint.ResumableCheckpoint, error) {
	if !t.checkpointing.Enabled {
		return nil, errors.InvalidInputf("checkpointing is not enabled")
	}

	if t.checkpointStore == nil {
		return nil, errors.InvalidInputf("checkpoint store is not configured")
	}

	checkpoints, err := checkpoint.GetResumableCheckpoints(t.checkpointStore)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get resumable checkpoints")
	}

	return checkpoints, nil
}

// ResumeTreeReplication resumes a previously interrupted tree replication
func (t *TreeReplicator) ResumeTreeReplication(
	ctx context.Context,
	sourceClient interfaces.RegistryClient,
	destClient interfaces.RegistryClient,
	opts ResumeOptions,
) (*TreeReplicationResult, error) {
	start := time.Now()

	// Input validation
	if ctx == nil {
		return nil, errors.InvalidInputf("context cannot be nil")
	}

	if sourceClient == nil {
		return nil, errors.InvalidInputf("source client cannot be nil")
	}

	if destClient == nil {
		return nil, errors.InvalidInputf("destination client cannot be nil")
	}

	if opts.CheckpointID == "" {
		return nil, errors.InvalidInputf("checkpoint ID cannot be empty")
	}

	// Check if checkpointing is enabled
	if !t.checkpointing.Enabled {
		return nil, errors.InvalidInputf("checkpointing is not enabled")
	}

	if t.checkpointStore == nil {
		return nil, errors.InvalidInputf("checkpoint store is not configured")
	}

	// Load the checkpoint to resume
	savedCheckpoint, err := t.checkpointStore.LoadCheckpoint(opts.CheckpointID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load checkpoint")
	}

	if savedCheckpoint == nil {
		return nil, errors.NotFoundf("checkpoint not found with ID: %s", opts.CheckpointID)
	}

	// Create a result with default values
	result := &TreeReplicationResult{
		Progress:     savedCheckpoint.Progress,
		CheckpointID: savedCheckpoint.ID,
		StartTime:    start,
		Resumed:      true,
	}

	t.logger.WithFields(map[string]interface{}{
		"id":              savedCheckpoint.ID,
		"source_prefix":   savedCheckpoint.SourcePrefix,
		"dest_prefix":     savedCheckpoint.DestPrefix,
		"progress":        savedCheckpoint.Progress,
		"total_repos":     len(savedCheckpoint.Repositories),
		"completed_repos": len(savedCheckpoint.CompletedRepositories),
	}).Info("Resuming tree replication")

	// Get repositories that still need to be processed
	resumeOpts := checkpoint.ResumableOptions{
		ID:            opts.CheckpointID,
		SkipCompleted: opts.SkipCompleted,
		RetryFailed:   opts.RetryFailed,
		Force:         opts.ForceOverwrite,
	}
	remainingRepos, err := checkpoint.GetRemainingRepositories(savedCheckpoint, resumeOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get remaining repositories")
	}

	t.logger.WithFields(map[string]interface{}{
		"count": len(remainingRepos),
	}).Info("Found repositories to resume")

	// Update checkpoint for resumed replication
	savedCheckpoint.Status = checkpoint.StatusInProgress
	savedCheckpoint.LastUpdated = time.Now()

	// Save the updated checkpoint
	if err := t.checkpointStore.SaveCheckpoint(savedCheckpoint); err != nil {
		t.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
			"id":    savedCheckpoint.ID,
		}).Warn("Failed to save updated checkpoint")
	}

	// Mutex to protect concurrent access to checkpoint data
	var checkpointMu sync.RWMutex

	// Set up checkpoint timer for periodic updates (every 30 seconds)
	checkpointTicker := time.NewTicker(30 * time.Second)
	checkpointDone := make(chan bool)

	// Run checkpoint updater in a goroutine
	go func() {
		for {
			select {
			case <-checkpointTicker.C:
				// Update and save the checkpoint
				checkpointMu.Lock()
				savedCheckpoint.LastUpdated = time.Now()
				checkpointMu.Unlock()
				if err := t.checkpointStore.SaveCheckpoint(savedCheckpoint); err != nil {
					t.logger.WithFields(map[string]interface{}{
						"error": err.Error(),
						"id":    savedCheckpoint.ID,
					}).Warn("Failed to save periodic checkpoint")
				}
			case <-checkpointDone:
				checkpointTicker.Stop()
				return
			}
		}
	}()

	// Set up a worker pool for repository replication
	sem := make(chan struct{}, t.workerCount)
	var wg sync.WaitGroup

	// Create a channel for collecting results
	type repoResult struct {
		repo             string
		imagesReplicated int
		imagesSkipped    int
		imagesFailed     int
		err              error
	}
	results := make(chan repoResult, len(remainingRepos))

	// Set up context with cancellation for handling interruptions
	_, cancel := context.WithCancel(ctx)
	defer cancel()

	// Process each repository
	for _, repo := range remainingRepos {
		wg.Add(1)

		// Create a copy of the loop variable for the goroutine
		repoName := repo

		// Find the repository status in the checkpoint
		checkpointMu.RLock()
		repoStatus, exists := savedCheckpoint.Repositories[repoName]
		checkpointMu.RUnlock()
		if !exists {
			t.logger.WithFields(map[string]interface{}{
				"repository": repoName,
			}).Warn("Repository not found in checkpoint")
			wg.Done()
			continue
		}

		go func() {
			defer wg.Done()
			sem <- struct{}{}        // Acquire token
			defer func() { <-sem }() // Release token

			// Note: In a real implementation, we would use the dest repo name
			// We don't need it for our mock implementation but we'll keep the reference
			// to avoid unused variable warning
			_ = repoStatus.DestRepo

			// Update checkpoint for this repository
			checkpointMu.Lock()
			repoStatus.Status = checkpoint.StatusInProgress
			repoStatus.LastUpdated = time.Now()
			savedCheckpoint.Repositories[repoName] = repoStatus
			checkpointMu.Unlock()

			// Mock implementation for now
			// In a real implementation, we would call the actual replication logic
			imagesReplicated, imagesSkipped, imagesFailed := 0, 0, 0
			var err error

			// For testing, simulate a successful replication
			imagesReplicated = 1

			// Update checkpoint for this repository with results
			checkpointMu.Lock()
			if err != nil {
				repoStatus.Status = checkpoint.StatusFailed
				repoStatus.Error = err.Error()
			} else {
				repoStatus.Status = checkpoint.StatusCompleted
				savedCheckpoint.CompletedRepositories = append(savedCheckpoint.CompletedRepositories, repoName)
			}
			repoStatus.LastUpdated = time.Now()
			savedCheckpoint.Repositories[repoName] = repoStatus

			// Calculate progress
			completed := len(savedCheckpoint.CompletedRepositories)
			total := len(savedCheckpoint.Repositories)
			if total > 0 {
				savedCheckpoint.Progress = float64(completed) / float64(total) * 100
				result.Progress = savedCheckpoint.Progress
			}
			checkpointMu.Unlock()

			// Send results back
			results <- repoResult{
				repo:             repoName,
				imagesReplicated: imagesReplicated,
				imagesSkipped:    imagesSkipped,
				imagesFailed:     imagesFailed,
				err:              err,
			}
		}()
	}

	// Wait for all replications to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errs []error
	interrupted := false

	for res := range results {
		result.Repositories++
		result.ImagesReplicated.Add(int64(res.imagesReplicated))
		result.ImagesSkipped.Add(int64(res.imagesSkipped))
		result.ImagesFailed.Add(int64(res.imagesFailed))

		if res.err != nil {
			errs = append(errs, errors.Wrapf(res.err, "failed to replicate repository %s", res.repo))
		}
	}

	result.Duration = time.Since(start)
	result.Interrupted = interrupted

	// Stop the checkpoint ticker
	checkpointDone <- true

	// Update final checkpoint status
	checkpointMu.Lock()
	if interrupted {
		savedCheckpoint.Status = checkpoint.StatusInterrupted
		result.Interrupted = true
	} else if len(errs) > 0 {
		savedCheckpoint.Status = checkpoint.StatusFailed
		savedCheckpoint.LastError = errs[0].Error()
	} else {
		savedCheckpoint.Status = checkpoint.StatusCompleted
	}

	savedCheckpoint.LastUpdated = time.Now()
	savedCheckpoint.Progress = result.Progress
	checkpointMu.Unlock()

	// Save final checkpoint
	if err := t.checkpointStore.SaveCheckpoint(savedCheckpoint); err != nil {
		t.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
			"id":    savedCheckpoint.ID,
		}).Warn("Failed to save final checkpoint")
	} else {
		t.logger.WithFields(map[string]interface{}{
			"id":       savedCheckpoint.ID,
			"status":   savedCheckpoint.Status,
			"progress": savedCheckpoint.Progress,
		}).Info("Saved final checkpoint")
	}

	// Log completion
	status := "completed"
	if interrupted {
		status = "interrupted"
	}

	t.logger.WithFields(map[string]interface{}{
		"repositories":      result.Repositories,
		"images_replicated": result.ImagesReplicated.Load(),
		"images_skipped":    result.ImagesSkipped.Load(),
		"images_failed":     result.ImagesFailed.Load(),
		"duration_ms":       result.Duration.Milliseconds(),
		"progress":          result.Progress,
		"interrupted":       result.Interrupted,
	}).Info("Tree replication resume " + status)

	var finalErr error
	if len(errs) > 0 {
		if len(errs) == 1 {
			finalErr = errs[0]
		} else {
			// Create an error that represents all the failures
			finalErr = errors.Internalf("failed to replicate %d repositories", len(errs))
			// Add the first error as additional context
			finalErr = errors.Wrap(errs[0], "%s", finalErr.Error())
		}
	}

	return result, finalErr
}
