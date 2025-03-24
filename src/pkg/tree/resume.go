package tree

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/hemzaz/freightliner/src/pkg/client/common"
	"github.com/hemzaz/freightliner/src/pkg/tree/checkpoint"
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
	if !t.enableCheckpoints || t.checkpointStore == nil {
		return nil, fmt.Errorf("checkpointing is not enabled")
	}

	return checkpoint.GetResumableCheckpoints(t.checkpointStore)
}

// ResumeTreeReplication resumes a previously interrupted tree replication
func (t *TreeReplicator) ResumeTreeReplication(
	ctx context.Context,
	sourceClient common.RegistryClient,
	destClient common.RegistryClient,
	opts ResumeOptions,
) (*ReplicationResult, error) {
	start := time.Now()

	// Check if checkpointing is enabled
	if !t.enableCheckpoints || t.checkpointStore == nil {
		return nil, fmt.Errorf("checkpointing is not enabled")
	}

	// Load the checkpoint to resume
	savedCheckpoint, err := t.checkpointStore.LoadCheckpoint(opts.CheckpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	// Create a result with default values
	result := &ReplicationResult{
		Progress: savedCheckpoint.Progress,
		Resumed:  true,
		CheckpointID: savedCheckpoint.ID,
	}

	t.logger.Info("Resuming tree replication", map[string]interface{}{
		"id":              savedCheckpoint.ID,
		"source_prefix":   savedCheckpoint.SourcePrefix,
		"dest_prefix":     savedCheckpoint.DestPrefix,
		"progress":        savedCheckpoint.Progress,
		"total_repos":     len(savedCheckpoint.Repositories),
		"completed_repos": len(savedCheckpoint.CompletedRepositories),
	})

	// Get repositories that still need to be processed
	resumeOpts := checkpoint.ResumableOptions{
		ID:            opts.CheckpointID,
		SkipCompleted: opts.SkipCompleted,
		RetryFailed:   opts.RetryFailed,
		Force:         opts.ForceOverwrite,
	}
	remainingRepos := checkpoint.GetRemainingRepositories(savedCheckpoint, resumeOpts)

	t.logger.Info("Found repositories to resume", map[string]interface{}{
		"count": len(remainingRepos),
	})

	// Update checkpoint for resumed replication
	savedCheckpoint.Status = checkpoint.StatusInProgress
	savedCheckpoint.LastUpdated = time.Now()

	// Save the updated checkpoint
	if err := t.checkpointStore.SaveCheckpoint(savedCheckpoint); err != nil {
		t.logger.Warn("Failed to save updated checkpoint", map[string]interface{}{
			"error": err.Error(),
			"id":    savedCheckpoint.ID,
		})
	}

	// Set up checkpoint timer for periodic updates
	checkpointTicker := time.NewTicker(t.checkpointInterval)
	checkpointDone := make(chan bool)

	// Run checkpoint updater in a goroutine
	go func() {
		for {
			select {
			case <-checkpointTicker.C:
				// Update and save the checkpoint
				savedCheckpoint.LastUpdated = time.Now()
				if err := t.checkpointStore.SaveCheckpoint(savedCheckpoint); err != nil {
					t.logger.Warn("Failed to save periodic checkpoint", map[string]interface{}{
						"error": err.Error(),
						"id":    savedCheckpoint.ID,
					})
				}
			case <-checkpointDone:
				checkpointTicker.Stop()
				return
			}
		}
	}()

	// Set up a worker pool for repository replication
	sem := make(chan struct{}, t.workerPool)
	var wg sync.WaitGroup

	// Create a channel for collecting results
	results := make(chan struct {
		repo            string
		imagesReplicated int
		imagesSkipped   int
		imagesFailed    int
		err             error
	}, len(remainingRepos))

	// Set up context with cancellation for handling interruptions
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Process each repository
	for i, repo := range remainingRepos {
		wg.Add(1)

		// Create a copy of the loop variable for the goroutine
		repoName := repo
		repoIndex := -1

		// Find the repository task index in the checkpoint
		for j, task := range savedCheckpoint.RepoTasks {
			if task.SourceRepository == repoName {
				repoIndex = j
				break
			}
		}

		// Skip if we couldn't find the repository in the tasks
		if repoIndex == -1 {
			t.logger.Warn("Repository not found in checkpoint tasks", map[string]interface{}{
				"repository": repoName,
			})
			wg.Done()
			continue
		}

		go func() {
			defer wg.Done()
			sem <- struct{}{} // Acquire token
			defer func() { <-sem }() // Release token

			// Get destination repository name from the checkpoint
			destRepo := savedCheckpoint.RepoTasks[repoIndex].DestRepository

			// Update checkpoint for this repository
			savedCheckpoint.RepoTasks[repoIndex].Status = checkpoint.StatusInProgress
			savedCheckpoint.RepoTasks[repoIndex].LastUpdated = time.Now()

			// Replicate the repository
			imagesReplicated, imagesSkipped, imagesFailed, err := t.replicateRepository(
				ctx, sourceClient, destClient, repoName, destRepo, opts.ForceOverwrite,
			)

			// Update checkpoint for this repository with results
			if err != nil {
				savedCheckpoint.RepoTasks[repoIndex].Status = checkpoint.StatusFailed
				savedCheckpoint.RepoTasks[repoIndex].Error = err.Error()
			} else {
				savedCheckpoint.RepoTasks[repoIndex].Status = checkpoint.StatusCompleted
				savedCheckpoint.CompletedRepositories = append(savedCheckpoint.CompletedRepositories, repoName)
			}
			savedCheckpoint.RepoTasks[repoIndex].LastUpdated = time.Now()

			// Calculate progress
			completed := len(savedCheckpoint.CompletedRepositories)
			total := len(savedCheckpoint.Repositories)
			savedCheckpoint.Progress = float64(completed) / float64(total) * 100
			result.Progress = savedCheckpoint.Progress

			// Send results back
			results <- struct {
				repo            string
				imagesReplicated int
				imagesSkipped   int
				imagesFailed    int
				err             error
			}{
				repo:            repoName,
				imagesReplicated: imagesReplicated,
				imagesSkipped:   imagesSkipped,
				imagesFailed:    imagesFailed,
				err:             err,
			}
		}()
	}

	// Wait for all replications to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errors []error
	interrupted := false

	for res := range results {
		result.Repositories++
		result.ImagesReplicated += res.imagesReplicated
		result.ImagesSkipped += res.imagesSkipped
		result.ImagesFailed += res.imagesFailed

		if res.err != nil {
			errors = append(errors, fmt.Errorf("failed to replicate repository %s: %w", res.repo, res.err))
		}

		// Include previously completed results in the total
		if opts.SkipCompleted {
			// Count previously completed repositories in the checkpoint
			for _, task := range savedCheckpoint.RepoTasks {
				if task.Status == checkpoint.StatusCompleted && task.SourceRepository != res.repo {
					// Check if this task was already counted in the current results
					found := false
					for _, repo := range remainingRepos {
						if repo == task.SourceRepository {
							found = true
							break
						}
					}
					if !found {
						result.Repositories++
						// We can't exactly know how many images were handled, but we can count them as skipped
						result.ImagesSkipped++
					}
				}
			}
		}
	}

	result.Duration = time.Since(start)
	result.Interrupted = interrupted

	// Stop the checkpoint ticker
	checkpointDone <- true

	// Update final checkpoint status
	if interrupted {
		savedCheckpoint.Status = checkpoint.StatusInterrupted
		result.Interrupted = true
	} else if len(errors) > 0 {
		savedCheckpoint.Status = checkpoint.StatusFailed
		savedCheckpoint.LastError = errors[0].Error()
	} else {
		savedCheckpoint.Status = checkpoint.StatusCompleted
	}

	savedCheckpoint.LastUpdated = time.Now()
	savedCheckpoint.Progress = result.Progress

	// Save final checkpoint
	if err := t.checkpointStore.SaveCheckpoint(savedCheckpoint); err != nil {
		t.logger.Warn("Failed to save final checkpoint", map[string]interface{}{
			"error": err.Error(),
			"id":    savedCheckpoint.ID,
		})
	} else {
		t.logger.Info("Saved final checkpoint", map[string]interface{}{
			"id":       savedCheckpoint.ID,
			"status":   savedCheckpoint.Status,
			"progress": savedCheckpoint.Progress,
		})
	}

	// Log completion
	status := "completed"
	if interrupted {
		status = "interrupted"
	}

	t.logger.Info("Tree replication resume "+status, map[string]interface{}{
		"repositories":      result.Repositories,
		"images_replicated": result.ImagesReplicated,
		"images_skipped":    result.ImagesSkipped,
		"images_failed":     result.ImagesFailed,
		"duration_ms":       result.Duration.Milliseconds(),
		"progress":          result.Progress,
		"interrupted":       result.Interrupted,
	})

	var finalErr error
	if len(errors) > 0 {
		// Combine all errors into a single error message
		finalErr = fmt.Errorf("failed to replicate %d repositories: %v", len(errors), errors)
	}

	return result, finalErr
}