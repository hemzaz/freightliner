package tree

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/hemzaz/freightliner/src/internal/log"
	"github.com/hemzaz/freightliner/src/internal/util"
	"github.com/hemzaz/freightliner/src/pkg/client/common"
	"github.com/hemzaz/freightliner/src/pkg/copy"
	"github.com/hemzaz/freightliner/src/pkg/metrics"
)

// TreeReplicator handles replicating entire repository trees between registries
type TreeReplicator struct {
	logger       *log.Logger
	copier       *copy.Copier
	workerPool   int
	metrics      metrics.Metrics
	excludeRepos []string
	excludeTags  []string
	includeTags  []string
	dryRun       bool
}

// TreeReplicatorOptions configures the tree replicator
type TreeReplicatorOptions struct {
	// WorkerCount is the number of concurrent repository replications
	WorkerCount int
	
	// ExcludeRepositories is a list of repository patterns to exclude
	ExcludeRepositories []string
	
	// ExcludeTags is a list of tag patterns to exclude
	ExcludeTags []string
	
	// IncludeTags is a list of tag patterns to include (if empty, all tags are included)
	IncludeTags []string
	
	// DryRun performs a dry run without actually copying images
	DryRun bool
}

// ReplicationResult contains the results of a tree replication
type ReplicationResult struct {
	// Repositories is the number of repositories processed
	Repositories int
	
	// ImagesReplicated is the number of images successfully replicated
	ImagesReplicated int
	
	// ImagesSkipped is the number of images skipped (already exists or excluded)
	ImagesSkipped int
	
	// ImagesFailed is the number of images that failed to replicate
	ImagesFailed int
	
	// Duration is the total duration of the replication
	Duration time.Duration
}

// NewTreeReplicator creates a new tree replicator
func NewTreeReplicator(logger *log.Logger, copier *copy.Copier, opts TreeReplicatorOptions) *TreeReplicator {
	workerCount := opts.WorkerCount
	if workerCount <= 0 {
		workerCount = 5 // Default to 5 concurrent replications
	}
	
	return &TreeReplicator{
		logger:       logger,
		copier:       copier,
		workerPool:   workerCount,
		metrics:      &metrics.NoopMetrics{}, // Default to no-op metrics
		excludeRepos: opts.ExcludeRepositories,
		excludeTags:  opts.ExcludeTags,
		includeTags:  opts.IncludeTags,
		dryRun:       opts.DryRun,
	}
}

// WithMetrics sets the metrics collector
func (t *TreeReplicator) WithMetrics(m metrics.Metrics) *TreeReplicator {
	t.metrics = m
	t.copier.WithMetrics(m) // Propagate metrics to the copier
	return t
}

// ReplicateTree replicates an entire tree of repositories from source to destination
func (t *TreeReplicator) ReplicateTree(
	ctx context.Context,
	sourceClient common.RegistryClient,
	destClient common.RegistryClient,
	sourcePrefix string,
	destPrefix string,
	forceOverwrite bool,
) (*ReplicationResult, error) {
	start := time.Now()
	
	// Get all repositories from the source registry
	t.logger.Info("Listing repositories in source registry", map[string]interface{}{
		"source_prefix": sourcePrefix,
	})
	
	sourceRepos, err := sourceClient.ListRepositories()
	if err != nil {
		return nil, fmt.Errorf("failed to list source repositories: %w", err)
	}
	
	// Filter repositories based on prefix and exclusions
	var filteredRepos []string
	for _, repo := range sourceRepos {
		// Check if the repository matches the source prefix
		if sourcePrefix != "" && !strings.HasPrefix(repo, sourcePrefix) {
			continue
		}
		
		// Check if the repository should be excluded
		excluded := false
		for _, excludePattern := range t.excludeRepos {
			if matchPattern(excludePattern, repo) {
				t.logger.Debug("Excluding repository", map[string]interface{}{
					"repository": repo,
					"pattern":    excludePattern,
				})
				excluded = true
				break
			}
		}
		
		if !excluded {
			filteredRepos = append(filteredRepos, repo)
		}
	}
	
	t.logger.Info("Found repositories to replicate", map[string]interface{}{
		"count": len(filteredRepos),
	})
	
	// Set up a worker pool for repository replication
	sem := make(chan struct{}, t.workerPool)
	var wg sync.WaitGroup
	
	// Create a channel for collecting results
	results := make(chan struct {
		repo           string
		imagesReplicated int
		imagesSkipped  int
		imagesFailed   int
		err            error
	}, len(filteredRepos))
	
	// Process each repository
	for _, repo := range filteredRepos {
		wg.Add(1)
		
		// Create a copy of the loop variable for the goroutine
		repoName := repo
		
		go func() {
			defer wg.Done()
			sem <- struct{}{} // Acquire token
			defer func() { <-sem }() // Release token
			
			// Calculate the destination repository name
			destRepo := repoName
			if sourcePrefix != "" && destPrefix != "" {
				// Replace the source prefix with the destination prefix
				destRepo = path.Join(destPrefix, strings.TrimPrefix(repoName, sourcePrefix))
			} else if destPrefix != "" {
				// Prepend the destination prefix
				destRepo = path.Join(destPrefix, repoName)
			}
			
			// Replicate the repository
			imagesReplicated, imagesSkipped, imagesFailed, err := t.replicateRepository(
				ctx, sourceClient, destClient, repoName, destRepo, forceOverwrite,
			)
			
			// Send results back
			results <- struct {
				repo           string
				imagesReplicated int
				imagesSkipped  int
				imagesFailed   int
				err            error
			}{
				repo:           repoName,
				imagesReplicated: imagesReplicated,
				imagesSkipped:  imagesSkipped,
				imagesFailed:   imagesFailed,
				err:            err,
			}
		}()
	}
	
	// Wait for all replications to complete
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	result := &ReplicationResult{}
	var errors []error
	
	for res := range results {
		result.Repositories++
		result.ImagesReplicated += res.imagesReplicated
		result.ImagesSkipped += res.imagesSkipped
		result.ImagesFailed += res.imagesFailed
		
		if res.err != nil {
			errors = append(errors, fmt.Errorf("failed to replicate repository %s: %w", res.repo, res.err))
		}
	}
	
	result.Duration = time.Since(start)
	
	t.logger.Info("Tree replication completed", map[string]interface{}{
		"repositories":      result.Repositories,
		"images_replicated": result.ImagesReplicated,
		"images_skipped":    result.ImagesSkipped,
		"images_failed":     result.ImagesFailed,
		"duration_ms":       result.Duration.Milliseconds(),
	})
	
	var finalErr error
	if len(errors) > 0 {
		// Combine all errors into a single error message
		finalErr = fmt.Errorf("failed to replicate %d repositories: %v", len(errors), errors)
	}
	
	return result, finalErr
}

// replicateRepository replicates a single repository from source to destination
func (t *TreeReplicator) replicateRepository(
	ctx context.Context,
	sourceClient common.RegistryClient,
	destClient common.RegistryClient,
	sourceRepo string,
	destRepo string,
	forceOverwrite bool,
) (int, int, int, error) {
	t.logger.Info("Replicating repository", map[string]interface{}{
		"source":      sourceRepo,
		"destination": destRepo,
	})
	
	// Get the source repository
	sourceRepository, err := sourceClient.GetRepository(sourceRepo)
	if err != nil {
		return 0, 0, 1, fmt.Errorf("failed to get source repository: %w", err)
	}
	
	// Get the destination repository
	destRepository, err := destClient.GetRepository(destRepo)
	if err != nil {
		return 0, 0, 1, fmt.Errorf("failed to get destination repository: %w", err)
	}
	
	// Get all tags from the source repository
	tags, err := sourceRepository.ListTags()
	if err != nil {
		return 0, 0, 1, fmt.Errorf("failed to list tags: %w", err)
	}
	
	// Get existing tags from destination repository for comparison
	existingTags := make(map[string]bool)
	destTags, err := destRepository.ListTags()
	if err != nil {
		t.logger.Warn("Failed to list destination tags", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		for _, tag := range destTags {
			existingTags[tag] = true
		}
	}
	
	// Filter tags based on include/exclude patterns
	var filteredTags []string
	for _, tag := range tags {
		// Check if the tag should be included
		if len(t.includeTags) > 0 {
			included := false
			for _, includePattern := range t.includeTags {
				if matchPattern(includePattern, tag) {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}
		
		// Check if the tag should be excluded
		excluded := false
		for _, excludePattern := range t.excludeTags {
			if matchPattern(excludePattern, tag) {
				excluded = true
				break
			}
		}
		
		if !excluded {
			filteredTags = append(filteredTags, tag)
		}
	}
	
	t.logger.Info("Replicating tags", map[string]interface{}{
		"repository": sourceRepo,
		"tag_count":  len(filteredTags),
	})
	
	// Copy each tag
	var imagesReplicated, imagesSkipped, imagesFailed int
	
	for _, tag := range filteredTags {
		// Check if tag already exists in destination
		if !forceOverwrite && existingTags[tag] {
			t.logger.Debug("Tag already exists, skipping", map[string]interface{}{
				"repository": sourceRepo,
				"tag":        tag,
			})
			imagesSkipped++
			continue
		}
		
		if t.dryRun {
			t.logger.Info("Dry run, would copy tag", map[string]interface{}{
				"source_repository":      sourceRepo,
				"destination_repository": destRepo,
				"tag":                    tag,
			})
			imagesReplicated++
			continue
		}
		
		// Copy the image
		err := t.copier.CopyImage(ctx, sourceRepository, destRepository, copy.CopyOptions{
			SourceTag:      tag,
			DestinationTag: tag,
			ForceOverwrite: forceOverwrite,
		})
		
		if err != nil {
			t.logger.Error("Failed to copy tag", err, map[string]interface{}{
				"source_repository":      sourceRepo,
				"destination_repository": destRepo,
				"tag":                    tag,
			})
			imagesFailed++
		} else {
			imagesReplicated++
		}
	}
	
	return imagesReplicated, imagesSkipped, imagesFailed, nil
}

// matchPattern checks if a string matches a simple glob pattern
func matchPattern(pattern, str string) bool {
	// Handle simple cases first
	if pattern == str {
		return true
	}
	
	if pattern == "*" {
		return true
	}
	
	// Handle prefix match (e.g., "foo*")
	if strings.HasSuffix(pattern, "*") && !strings.Contains(pattern[:len(pattern)-1], "*") {
		return strings.HasPrefix(str, pattern[:len(pattern)-1])
	}
	
	// Handle suffix match (e.g., "*foo")
	if strings.HasPrefix(pattern, "*") && !strings.Contains(pattern[1:], "*") {
		return strings.HasSuffix(str, pattern[1:])
	}
	
	// Handle contains match (e.g., "*foo*")
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") && len(pattern) > 2 {
		middle := pattern[1 : len(pattern)-1]
		if !strings.Contains(middle, "*") {
			return strings.Contains(str, middle)
		}
	}
	
	// Fall back to path.Match for more complex patterns
	matched, _ := path.Match(pattern, str)
	return matched
}