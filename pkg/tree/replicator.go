package tree

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/client/common"
	"freightliner/pkg/copy"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/tree/checkpoint"

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
	// Whether this is a resumed replication
	Resumed bool
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

// ReplicateTreeOptions provides options for the ReplicateTree method
type ReplicateTreeOptions struct {
	// SourceClient is the client for the source registry
	SourceClient common.RegistryClient

	// DestClient is the client for the destination registry
	DestClient common.RegistryClient

	// SourcePrefix is the prefix for source repositories
	SourcePrefix string

	// DestPrefix is the prefix for destination repositories
	DestPrefix string

	// ForceOverwrite determines whether to overwrite existing images
	ForceOverwrite bool

	// ResumeFromCheckpoint is the ID of a checkpoint to resume from
	ResumeFromCheckpoint string

	// SkipCompletedRepositories skips repositories marked as completed in the checkpoint
	SkipCompletedRepositories bool
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
func (t *TreeReplicator) ReplicateTree(
	ctx context.Context,
	opts ReplicateTreeOptions,
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
			SourceRegistry: opts.SourceClient.GetRegistryName(),
			SourcePrefix:   opts.SourcePrefix,
			DestRegistry:   opts.DestClient.GetRegistryName(),
			DestPrefix:     opts.DestPrefix,
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
	repositories, err := t.listAndFilterRepositories(ctx, opts.SourceClient, opts.SourcePrefix)
	if err != nil {
		t.handleError(err, treeCheckpoint, "Failed to list repositories")
		return result, err
	}

	repoCount := len(repositories)
	result.Repositories = repoCount

	if repoCount == 0 {
		t.logger.Info("No repositories found matching prefix", map[string]interface{}{
			"source_registry": opts.SourceClient.GetRegistryName(),
			"prefix":          opts.SourcePrefix,
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
	workerOpts := replicationWorkerOptions{
		Context:        ctx,
		Jobs:           jobs,
		WaitGroup:      &wg,
		CompletedRepos: &completedRepos,
		ErrorCount:     &errorCount,
		SourceClient:   opts.SourceClient,
		DestClient:     opts.DestClient,
		SourcePrefix:   opts.SourcePrefix,
		DestPrefix:     opts.DestPrefix,
		ForceOverwrite: opts.ForceOverwrite,
		TreeCheckpoint: treeCheckpoint,
		Result:         result,
	}

	for i := 0; i < t.workerCount; i++ {
		wg.Add(1)
		go t.replicationWorker(workerOpts)
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

// replicationWorkerOptions holds all parameters for replication workers
type replicationWorkerOptions struct {
	Context        context.Context
	Jobs           <-chan replicationJob
	WaitGroup      *sync.WaitGroup
	CompletedRepos *atomic.Int32
	ErrorCount     *atomic.Int32
	SourceClient   common.RegistryClient
	DestClient     common.RegistryClient
	SourcePrefix   string
	DestPrefix     string
	ForceOverwrite bool
	TreeCheckpoint *checkpoint.TreeCheckpoint
	Result         *TreeReplicationResult
}

// replicationWorker processes repository replication jobs
func (t *TreeReplicator) replicationWorker(opts replicationWorkerOptions) {
	defer opts.WaitGroup.Done()

	for job := range opts.Jobs {
		select {
		case <-opts.Context.Done():
			return
		default:
			// Process job
			repo := job.repository

			// Generate destination repository name by replacing prefix
			destRepo := strings.Replace(repo, opts.SourcePrefix, opts.DestPrefix, 1)

			t.logger.Info("Replicating repository", map[string]interface{}{
				"source":      fmt.Sprintf("%s/%s", opts.SourceClient.GetRegistryName(), repo),
				"destination": fmt.Sprintf("%s/%s", opts.DestClient.GetRegistryName(), destRepo),
				"dry_run":     t.dryRun,
			})

			// Create options for processing the repository
			processOpts := repositoryProcessOptions{
				Context:        opts.Context,
				SourceClient:   opts.SourceClient,
				DestClient:     opts.DestClient,
				SourceRepo:     repo,
				DestRepo:       destRepo,
				ForceOverwrite: opts.ForceOverwrite,
				TreeCheckpoint: opts.TreeCheckpoint,
				Result:         opts.Result,
			}

			// Process repository
			if err := t.processRepository(processOpts); err != nil {
				opts.ErrorCount.Add(1)
				t.logger.Error("Failed to replicate repository", err, map[string]interface{}{
					"source":      fmt.Sprintf("%s/%s", opts.SourceClient.GetRegistryName(), repo),
					"destination": fmt.Sprintf("%s/%s", opts.DestClient.GetRegistryName(), destRepo),
				})
			}

			opts.CompletedRepos.Add(1)
		}
	}
}

// repositoryProcessOptions holds options for processing a single repository
type repositoryProcessOptions struct {
	Context        context.Context
	SourceClient   common.RegistryClient
	DestClient     common.RegistryClient
	SourceRepo     string
	DestRepo       string
	ForceOverwrite bool
	TreeCheckpoint *checkpoint.TreeCheckpoint
	Result         *TreeReplicationResult
}

// processRepository handles the replication of a single repository
func (t *TreeReplicator) processRepository(opts repositoryProcessOptions) error {
	// Process repository implementation
	// and handle tag filtering, image copying, etc.

	// Update checkpoint if enabled
	if t.checkpointing.Enabled && t.checkpointStore != nil && opts.TreeCheckpoint != nil {
		opts.TreeCheckpoint.Repositories[opts.SourceRepo] = checkpoint.RepoStatus{
			Status:     checkpoint.StatusInProgress,
			SourceRepo: opts.SourceRepo,
			DestRepo:   opts.DestRepo,
		}

		if err := t.checkpointStore.SaveCheckpoint(opts.TreeCheckpoint); err != nil {
			wrappedErr := errors.Wrap(err, "failed to update repository checkpoint")
			t.logger.Warn(wrappedErr.Error(), map[string]interface{}{
				"checkpoint_id": opts.TreeCheckpoint.ID,
				"source_repo":   opts.SourceRepo,
				"dest_repo":     opts.DestRepo,
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
	if t.checkpointing.Enabled && t.checkpointStore != nil && opts.TreeCheckpoint != nil {
		if repo, ok := opts.TreeCheckpoint.Repositories[opts.SourceRepo]; ok {
			repo.Status = checkpoint.StatusCompleted
			opts.TreeCheckpoint.Repositories[opts.SourceRepo] = repo
			opts.TreeCheckpoint.CompletedRepositories = append(opts.TreeCheckpoint.CompletedRepositories, opts.SourceRepo)
		}
	}

	return nil
}

// Helper functions to break down the large methods

// filterTags applies tag filters using the optimized pattern caches
// Returns tags that should be included (pass all filters)
func (t *TreeReplicator) filterTags(tags []string) []string {
	// Skip filtering if no filters are defined
	if len(t.filters.ExcludeTags) == 0 && len(t.filters.IncludeTags) == 0 {
		return tags
	}

	// Pre-allocate result slice for better performance
	estimatedSize := estimateFilteredSize(tags, len(t.filters.IncludeTags) > 0)
	result := make([]string, 0, estimatedSize)

	// Filter each tag
	for _, tag := range tags {
		if isTagIncluded(tag, t.excludeTagsCache, t.includeTagsCache, t.filters.IncludeTags) {
			result = append(result, tag)
		}
	}

	return result
}

// estimateFilteredSize estimates how many tags will pass filtering
func estimateFilteredSize(tags []string, hasIncludeFilters bool) int {
	estimatedSize := len(tags)
	if hasIncludeFilters {
		// If we have include filters, we'll likely get fewer tags
		estimatedSize = estimatedSize / 2
	}
	if estimatedSize < 10 {
		estimatedSize = 10
	}
	return estimatedSize
}

// isTagIncluded determines if a tag should be included based on filters
func isTagIncluded(tag string, excludeCache, includeCache *patternCache, includePatterns []string) bool {
	// Check exclusion filters first (excluded tags are always removed)
	if excludeCache != nil && excludeCache.matches(tag) {
		return false
	}

	// If there are no include patterns, all non-excluded tags are included
	if len(includePatterns) == 0 {
		return true
	}

	// If there are include patterns, tag must match at least one
	return includeCache != nil && includeCache.matches(tag)
}

// listAndFilterRepositoriesOptions holds options for listing and filtering repositories
type listAndFilterRepositoriesOptions struct {
	Context      context.Context
	SourceClient common.RegistryClient
	SourcePrefix string
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

// handleErrorOptions contains options for handling errors
type handleErrorOptions struct {
	Error          error
	TreeCheckpoint *checkpoint.TreeCheckpoint
	Message        string
}

// handleError processes errors and updates checkpoints
func (t *TreeReplicator) handleError(err error, treeCheckpoint *checkpoint.TreeCheckpoint, message string) {
	// Add context to the error if it's not already wrapped
	if !strings.Contains(err.Error(), message) {
		err = errors.Wrap(err, "%s", message)
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

// completeReplicationOptions contains options for completing replication
type completeReplicationOptions struct {
	TreeCheckpoint *checkpoint.TreeCheckpoint
	Result         *TreeReplicationResult
	Status         checkpoint.Status
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

// patternCache caches compiled glob patterns for faster matching
type patternCache struct {
	exactMatches    map[string]struct{} // For exact string matches
	prefixMatches   map[string]struct{} // For prefix* style patterns
	suffixMatches   map[string]struct{} // For *suffix style patterns
	containsMatches map[string]struct{} // For *contains* style patterns
	complexPatterns []string            // For more complex patterns requiring path.Match
	hasWildcard     bool                // Whether "*" is in the patterns
}

// newPatternCacheOptions contains options for creating a pattern cache
type newPatternCacheOptions struct {
	Patterns []string
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
	// Handle empty case
	if pc == nil {
		return false
	}

	// Check patterns in order of evaluation cost

	// Universal wildcard - fastest check
	if pc.hasWildcard {
		return true
	}

	// Exact match - very fast
	if _, ok := pc.exactMatches[s]; ok {
		return true
	}

	// Prefix, suffix, and contains checks - still relatively fast
	if pc.matchesPrefix(s) || pc.matchesSuffix(s) || pc.matchesContains(s) {
		return true
	}

	// Complex pattern matching - most expensive, do last
	return pc.matchesComplex(s)
}

// matchesPrefix checks if the string matches any prefix pattern
func (pc *patternCache) matchesPrefix(s string) bool {
	for prefix := range pc.prefixMatches {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// matchesSuffix checks if the string matches any suffix pattern
func (pc *patternCache) matchesSuffix(s string) bool {
	for suffix := range pc.suffixMatches {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// matchesContains checks if the string matches any contains pattern
func (pc *patternCache) matchesContains(s string) bool {
	for contains := range pc.containsMatches {
		if strings.Contains(s, contains) {
			return true
		}
	}
	return false
}

// matchesComplex checks if the string matches any complex pattern
func (pc *patternCache) matchesComplex(s string) bool {
	for _, pattern := range pc.complexPatterns {
		matched, _ := path.Match(pattern, s)
		if matched {
			return true
		}
	}
	return false
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

// InitCheckpointStore initializes the checkpoint store in the specified directory
func InitCheckpointStore(dir string) (checkpoint.CheckpointStore, error) {
	// Implementation would be here
	return nil, nil
}
