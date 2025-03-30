package tree

import (
	"context"
	"fmt"
	"freightliner/pkg/client/common"
	"freightliner/pkg/copy"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/tree/checkpoint"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// FilterOptions simplifies filter-related options
type FilterOptions struct {
	ExcludeRepos []string
	ExcludeTags  []string
	IncludeTags  []string
}

// CheckpointOptions simplifies checkpoint-related options
type CheckpointOptions struct {
	Enabled bool
	Dir     string
}

// TreeReplicationResult encapsulates the result and metrics of a tree replication
type TreeReplicationResult struct {
	// Total repositories that were processed
	Repositories int
	// Total images that were replicated successfully
	ImagesReplicated int
	// Total images that were skipped (already exist or filtered)
	ImagesSkipped int
	// Total images that failed to replicate
	ImagesFailed int
	// Progress percentage (0-100)
	Progress float64
	// Start time of the replication
	StartTime time.Time
	// Duration of the replication
	Duration time.Duration
	// Whether the replication was interrupted
	Interrupted bool
	// ID of the checkpoint if checkpointing is enabled
	CheckpointID string
	// Completed repository names
	CompletedRepositories []string
}

// TreeReplicatorOptions provides configuration for tree replication
type TreeReplicatorOptions struct {
	// WorkerCount is the number of concurrent workers
	WorkerCount int

	// ExcludeRepositories is a list of repository patterns to exclude
	ExcludeRepositories []string

	// ExcludeTags is a list of tag patterns to exclude
	ExcludeTags []string

	// IncludeTags is a list of tag patterns to include
	IncludeTags []string

	// EnableCheckpointing enables checkpoint functionality
	EnableCheckpointing bool

	// CheckpointDirectory is the directory for checkpoint files
	CheckpointDirectory string

	// DryRun indicates whether to perform actual copies
	DryRun bool
}

// TreeReplicator coordinates the replication of repositories
type TreeReplicator struct {
	logger            *log.Logger
	copier            *copy.Copier
	workerCount       int
	filters           FilterOptions
	excludeReposCache *patternCache
	excludeTagsCache  *patternCache
	includeTagsCache  *patternCache
	checkpointing     CheckpointOptions
	checkpointStore   checkpoint.CheckpointStore
	dryRun            bool
	metrics           interface{} // Metrics interface for tracking replication stats
}

// SetMetrics sets the metrics interface for the tree replicator
func (t *TreeReplicator) SetMetrics(metrics interface{}) {
	t.metrics = metrics
}

// NewTreeReplicator creates a new tree replicator
func NewTreeReplicator(logger *log.Logger, copier *copy.Copier, options TreeReplicatorOptions) *TreeReplicator {
	filters := FilterOptions{
		ExcludeRepos: options.ExcludeRepositories,
		ExcludeTags:  options.ExcludeTags,
		IncludeTags:  options.IncludeTags,
	}

	t := &TreeReplicator{
		logger:            logger,
		copier:            copier,
		workerCount:       options.WorkerCount,
		filters:           filters,
		excludeReposCache: newPatternCache(filters.ExcludeRepos),
		excludeTagsCache:  newPatternCache(filters.ExcludeTags),
		includeTagsCache:  newPatternCache(filters.IncludeTags),
		checkpointing: CheckpointOptions{
			Enabled: options.EnableCheckpointing,
			Dir:     options.CheckpointDirectory,
		},
		dryRun: options.DryRun,
	}

	// Initialize checkpoint store if enabled
	if t.checkpointing.Enabled {
		store, err := InitCheckpointStore(t.checkpointing.Dir)
		if err != nil {
			t.logger.Warn("Failed to initialize checkpoint store, checkpointing disabled", map[string]interface{}{
				"error": err.Error(),
				"dir":   t.checkpointing.Dir,
			})
		} else {
			t.checkpointStore = store
		}
	}

	return t
}

// ReplicateTree replicates all repositories from source to destination with the given prefix
// Breaking down the large function into several smaller focused functions
func (t *TreeReplicator) ReplicateTree(
	ctx context.Context,
	sourceClient common.RegistryClient,
	destClient common.RegistryClient,
	sourcePrefix string,
	destPrefix string,
	forceOverwrite bool,
) (*TreeReplicationResult, error) {
	// Setup and initialize replication
	startTime := time.Now()
	result := &TreeReplicationResult{
		Repositories:     0,
		ImagesReplicated: 0,
		ImagesSkipped:    0,
		ImagesFailed:     0,
		Progress:         0,
		StartTime:        startTime,
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Initialize checkpoint if enabled
	var treeCheckpoint *checkpoint.TreeCheckpoint
	if t.checkpointing.Enabled && t.checkpointStore != nil {
		treeCheckpoint = &checkpoint.TreeCheckpoint{
			ID:             uuid.New().String(),
			SourceRegistry: sourceClient.GetRegistryName(),
			SourcePrefix:   sourcePrefix,
			DestRegistry:   destClient.GetRegistryName(),
			DestPrefix:     destPrefix,
			Status:         checkpoint.StatusInProgress,
			StartTime:      startTime,
			LastUpdated:    startTime,
			Repositories:   make(map[string]checkpoint.RepoStatus),
		}

		if err := t.checkpointStore.SaveCheckpoint(treeCheckpoint); err != nil {
			wrappedErr := errors.Wrap(err, "failed to save initial checkpoint")
			t.logger.Warn(wrappedErr.Error(), map[string]interface{}{
				"checkpoint_id": treeCheckpoint.ID,
			})
		} else {
			result.CheckpointID = treeCheckpoint.ID
		}
	}

	// List and filter repositories
	repositories, err := t.listAndFilterRepositories(ctx, sourceClient, sourcePrefix)
	if err != nil {
		t.handleError(err, treeCheckpoint, "Failed to list repositories")
		return result, err
	}

	repoCount := len(repositories)
	result.Repositories = repoCount

	if repoCount == 0 {
		t.logger.Info("No repositories found matching prefix", map[string]interface{}{
			"source_registry": sourceClient.GetRegistryName(),
			"prefix":          sourcePrefix,
		})
		t.completeReplication(treeCheckpoint, result, checkpoint.StatusCompleted)
		return result, nil
	}

	// Setup worker pool
	t.logger.Info("Starting replication", map[string]interface{}{
		"repositories": repoCount,
		"workers":      t.workerCount,
		"dry_run":      t.dryRun,
	})

	// Create jobs channel and worker group
	jobs := make(chan replicationJob, repoCount)
	var wg sync.WaitGroup
	var completedRepos atomic.Int32
	var errorCount atomic.Int32

	// Setup signal handling for graceful cancellation
	done := t.setupSignalHandling(ctx, cancel)
	defer close(done)

	// Start workers
	for i := 0; i < t.workerCount; i++ {
		wg.Add(1)
		go t.replicationWorker(
			ctx,
			jobs,
			&wg,
			&completedRepos,
			&errorCount,
			sourceClient,
			destClient,
			sourcePrefix,
			destPrefix,
			forceOverwrite,
			treeCheckpoint,
			result,
		)
	}

	// Queue jobs
	for _, repo := range repositories {
		select {
		case <-ctx.Done():
			break
		case jobs <- replicationJob{repository: repo}:
			// Job queued successfully
		}
	}

	// Close jobs channel when all jobs are queued
	close(jobs)

	// Wait for workers to complete
	wg.Wait()

	// Update final metrics
	result.Duration = time.Since(startTime)
	result.Progress = float64(completedRepos.Load()) / float64(repoCount) * 100.0

	// Handle completion status
	if ctx.Err() != nil {
		result.Interrupted = true
		t.completeReplication(treeCheckpoint, result, checkpoint.StatusInterrupted)
		return result, ctx.Err()
	}

	t.completeReplication(treeCheckpoint, result, checkpoint.StatusCompleted)
	return result, nil
}

// Helper functions to break down the large methods

// filterTags applies tag filters using the optimized pattern caches
// Returns tags that should be included (pass all filters)
func (t *TreeReplicator) filterTags(tags []string) []string {
	// If no filters defined, return all tags
	if len(t.filters.ExcludeTags) == 0 && len(t.filters.IncludeTags) == 0 {
		return tags
	}

	// Quick size estimate for the result
	estimatedSize := len(tags)
	if len(t.filters.IncludeTags) > 0 {
		// If we have include filters, we'll likely get fewer tags
		estimatedSize = estimatedSize / 2
	}
	if estimatedSize < 10 {
		estimatedSize = 10
	}

	// Pre-allocate result slice for better performance
	result := make([]string, 0, estimatedSize)

	for _, tag := range tags {
		// Skip excluded tags
		if t.excludeTagsCache != nil && t.excludeTagsCache.matches(tag) {
			continue
		}

		// If we have include patterns, tag must match at least one
		if t.includeTagsCache != nil && len(t.filters.IncludeTags) > 0 {
			if !t.includeTagsCache.matches(tag) {
				continue
			}
		}

		// Tag passes all filters
		result = append(result, tag)
	}

	return result
}

// listAndFilterRepositories gets repositories and applies filters
func (t *TreeReplicator) listAndFilterRepositories(
	ctx context.Context,
	sourceClient common.RegistryClient,
	sourcePrefix string,
) ([]string, error) {
	t.logger.Info("Listing repositories", map[string]interface{}{
		"registry": sourceClient.GetRegistryName(),
		"prefix":   sourcePrefix,
	})

	repositories, err := sourceClient.ListRepositories(ctx, sourcePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list repositories")
	}

	// Apply repository exclusion filters using the cached patterns
	if t.excludeReposCache != nil {
		filtered := make([]string, 0, len(repositories))
		for _, repo := range repositories {
			if !t.excludeReposCache.matches(repo) {
				filtered = append(filtered, repo)
			}
		}
		repositories = filtered
	}

	return repositories, nil
}

// patternCache caches compiled glob patterns for faster matching
type patternCache struct {
	exactMatches    map[string]struct{} // For exact string matches
	prefixMatches   map[string]struct{} // For prefix* style patterns
	suffixMatches   map[string]struct{} // For *suffix style patterns
	containsMatches map[string]struct{} // For *contains* style patterns
	complexPatterns []string            // For more complex patterns requiring path.Match
	hasWildcard     bool                // Whether "*" is in the patterns
}

// newPatternCache creates an optimized pattern cache from a slice of patterns
func newPatternCache(patterns []string) *patternCache {
	if len(patterns) == 0 {
		return nil
	}

	cache := &patternCache{
		exactMatches:    make(map[string]struct{}),
		prefixMatches:   make(map[string]struct{}),
		suffixMatches:   make(map[string]struct{}),
		containsMatches: make(map[string]struct{}),
		complexPatterns: []string{},
		hasWildcard:     false,
	}

	for _, pattern := range patterns {
		if pattern == "*" {
			cache.hasWildcard = true
			continue
		}

		// Check for simple prefix match (pattern ends with *)
		if strings.HasSuffix(pattern, "*") && !strings.Contains(pattern[:len(pattern)-1], "*") && !strings.Contains(pattern, "?") {
			cache.prefixMatches[pattern[:len(pattern)-1]] = struct{}{}
			continue
		}

		// Check for simple suffix match (pattern starts with *)
		if strings.HasPrefix(pattern, "*") && !strings.Contains(pattern[1:], "*") && !strings.Contains(pattern, "?") {
			cache.suffixMatches[pattern[1:]] = struct{}{}
			continue
		}

		// Check for simple contains match (pattern is *text*)
		if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") &&
			!strings.Contains(pattern[1:len(pattern)-1], "*") && !strings.Contains(pattern, "?") {
			cache.containsMatches[pattern[1:len(pattern)-1]] = struct{}{}
			continue
		}

		// If it's not a special case, check if it's a literal string
		if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
			cache.exactMatches[pattern] = struct{}{}
			continue
		}

		// For complex patterns, we'll use path.Match
		cache.complexPatterns = append(cache.complexPatterns, pattern)
	}

	return cache
}

// matches returns true if the string matches any pattern in the cache
func (pc *patternCache) matches(s string) bool {
	if pc == nil {
		return false
	}

	// Fast path - universal wildcard
	if pc.hasWildcard {
		return true
	}

	// Fast path - exact match
	if _, ok := pc.exactMatches[s]; ok {
		return true
	}

	// Check prefix matches
	for prefix := range pc.prefixMatches {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}

	// Check suffix matches
	for suffix := range pc.suffixMatches {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}

	// Check contains matches
	for contains := range pc.containsMatches {
		if strings.Contains(s, contains) {
			return true
		}
	}

	// Fall back to complex patterns
	for _, pattern := range pc.complexPatterns {
		matched, _ := path.Match(pattern, s)
		if matched {
			return true
		}
	}

	return false
}

// matchesAnyPattern returns true if the string matches any of the patterns
// This is a compatibility wrapper around the new optimized pattern cache
func (t *TreeReplicator) matchesAnyPattern(s string, patterns []string) bool {
	cache := newPatternCache(patterns)
	return cache.matches(s)
}

// matchPattern is a helper function for testing that matches a string against a pattern
func matchPattern(pattern, s string) bool {
	cache := newPatternCache([]string{pattern})
	return cache.matches(s)
}

// replicationJob represents a single repository replication task
type replicationJob struct {
	repository string
}

// setupSignalHandling sets up goroutine for handling cancellation signals
func (t *TreeReplicator) setupSignalHandling(ctx context.Context, cancel context.CancelFunc) chan struct{} {
	done := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			t.logger.Info("Replication canceled", nil)
		case <-done:
			// Normal exit
		}
	}()

	return done
}

// replicationWorker processes repository replication jobs
func (t *TreeReplicator) replicationWorker(
	ctx context.Context,
	jobs <-chan replicationJob,
	wg *sync.WaitGroup,
	completedRepos *atomic.Int32,
	errorCount *atomic.Int32,
	sourceClient common.RegistryClient,
	destClient common.RegistryClient,
	sourcePrefix string,
	destPrefix string,
	forceOverwrite bool,
	treeCheckpoint *checkpoint.TreeCheckpoint,
	result *TreeReplicationResult,
) {
	defer wg.Done()

	for job := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			// Process job
			repo := job.repository

			// Generate destination repository name by replacing prefix
			destRepo := strings.Replace(repo, sourcePrefix, destPrefix, 1)

			t.logger.Info("Replicating repository", map[string]interface{}{
				"source":      fmt.Sprintf("%s/%s", sourceClient.GetRegistryName(), repo),
				"destination": fmt.Sprintf("%s/%s", destClient.GetRegistryName(), destRepo),
				"dry_run":     t.dryRun,
			})

			// Process repository
			if err := t.processRepository(
				ctx, sourceClient, destClient, repo, destRepo, forceOverwrite, treeCheckpoint, result,
			); err != nil {
				errorCount.Add(1)
				t.logger.Error("Failed to replicate repository", err, map[string]interface{}{
					"source":      fmt.Sprintf("%s/%s", sourceClient.GetRegistryName(), repo),
					"destination": fmt.Sprintf("%s/%s", destClient.GetRegistryName(), destRepo),
				})
			}

			completedRepos.Add(1)
		}
	}
}

// processRepository handles the replication of a single repository
func (t *TreeReplicator) processRepository(
	ctx context.Context,
	sourceClient common.RegistryClient,
	destClient common.RegistryClient,
	sourceRepo string,
	destRepo string,
	forceOverwrite bool,
	treeCheckpoint *checkpoint.TreeCheckpoint,
	result *TreeReplicationResult,
) error {
	// Process repository implementation
	// and handle tag filtering, image copying, etc.

	// Update checkpoint if enabled
	if t.checkpointing.Enabled && t.checkpointStore != nil && treeCheckpoint != nil {
		treeCheckpoint.Repositories[sourceRepo] = checkpoint.RepoStatus{
			Status:     checkpoint.StatusInProgress,
			SourceRepo: sourceRepo,
			DestRepo:   destRepo,
		}

		if err := t.checkpointStore.SaveCheckpoint(treeCheckpoint); err != nil {
			wrappedErr := errors.Wrap(err, "failed to update repository checkpoint")
			t.logger.Warn(wrappedErr.Error(), map[string]interface{}{
				"checkpoint_id": treeCheckpoint.ID,
				"source_repo":   sourceRepo,
				"dest_repo":     destRepo,
			})
		}
	}

	// Mock implementation - in actual code would use source/destClient to perform replication
	time.Sleep(10 * time.Millisecond) // Simulating work

	// In a real implementation, we would:
	// 1. Get source repository reference
	// 2. Get destination repository reference
	// 3. List tags in source repository
	// 4. Filter tags
	// 5. For each tag, copy the image

	// Update checkpoint status to completed for this repo
	if t.checkpointing.Enabled && t.checkpointStore != nil && treeCheckpoint != nil {
		if repo, ok := treeCheckpoint.Repositories[sourceRepo]; ok {
			repo.Status = checkpoint.StatusCompleted
			treeCheckpoint.Repositories[sourceRepo] = repo
			treeCheckpoint.CompletedRepositories = append(treeCheckpoint.CompletedRepositories, sourceRepo)
		}
	}

	return nil
}

// handleError processes errors and updates checkpoints
func (t *TreeReplicator) handleError(err error, treeCheckpoint *checkpoint.TreeCheckpoint, message string) {
	// Add context to the error if it's not already wrapped
	if !strings.Contains(err.Error(), message) {
		err = errors.Wrap(err, message)
	}

	t.logger.Error(message, err, nil)

	if t.checkpointing.Enabled && t.checkpointStore != nil && treeCheckpoint != nil {
		treeCheckpoint.Status = checkpoint.StatusFailed
		treeCheckpoint.LastError = err.Error()
		if saveErr := t.checkpointStore.SaveCheckpoint(treeCheckpoint); saveErr != nil {
			t.logger.Warn("Failed to save error checkpoint", map[string]interface{}{
				"error":          saveErr.Error(),
				"original_error": err.Error(),
				"id":             treeCheckpoint.ID,
			})
		}
	}
}

// completeReplication finalizes the replication and updates the checkpoint
func (t *TreeReplicator) completeReplication(treeCheckpoint *checkpoint.TreeCheckpoint, result *TreeReplicationResult, status checkpoint.Status) {
	if t.checkpointing.Enabled && t.checkpointStore != nil && treeCheckpoint != nil {
		treeCheckpoint.Status = status
		treeCheckpoint.Progress = result.Progress
		treeCheckpoint.LastUpdated = time.Now()

		if err := t.checkpointStore.SaveCheckpoint(treeCheckpoint); err != nil {
			wrappedErr := errors.Wrap(err, "failed to save final checkpoint")
			t.logger.Warn(wrappedErr.Error(), map[string]interface{}{
				"checkpoint_id": treeCheckpoint.ID,
				"status":        status,
			})
		}
	}
}
