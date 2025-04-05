package replication

import (
	"context"
	"sync"
	"time"

	"freightliner/pkg/client/common"
	"freightliner/pkg/copy"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/metrics"
)

// Reconciler handles reconciling repositories between registries
type Reconciler struct {
	logger      *log.Logger
	copier      *copy.Copier
	workerPool  *WorkerPool
	metrics     metrics.MetricsCollector
	dryRun      bool
	forceUpdate bool
}

// ReconcilerOptions configures the reconciler behavior
type ReconcilerOptions struct {
	Logger      *log.Logger
	Copier      *copy.Copier
	WorkerPool  *WorkerPool
	Metrics     metrics.MetricsCollector
	DryRun      bool
	ForceUpdate bool
}

// NewReconciler creates a new reconciler
func NewReconciler(opts ReconcilerOptions) *Reconciler {
	return &Reconciler{
		logger:      opts.Logger,
		copier:      opts.Copier,
		workerPool:  opts.WorkerPool,
		metrics:     opts.Metrics,
		dryRun:      opts.DryRun,
		forceUpdate: opts.ForceUpdate,
	}
}

// ReconcileRepository reconciles a source repository with a destination repository
func (r *Reconciler) ReconcileRepository(
	ctx context.Context,
	rule ReplicationRule,
	sourceClient common.RegistryClient,
	destClient common.RegistryClient) error {

	// Validate input parameters
	if err := r.validateReconcileParams(rule, sourceClient, destClient); err != nil {
		return err
	}

	// Get repositories and tags
	sourceRepo, destRepo, sourceTags, destTagMap, err := r.getRepositoriesAndTags(ctx, rule, sourceClient, destClient)
	if err != nil {
		return err
	}

	// Process and replicate tags
	return r.processAndReplicateTags(ctx, rule, sourceRepo, destRepo, sourceTags, destTagMap)
}

// validateReconcileParams validates the input parameters for reconciliation
func (r *Reconciler) validateReconcileParams(
	rule ReplicationRule,
	sourceClient common.RegistryClient,
	destClient common.RegistryClient) error {

	if rule.SourceRepository == "" {
		return errors.InvalidInputf("source repository cannot be empty")
	}

	if rule.DestinationRepository == "" {
		return errors.InvalidInputf("destination repository cannot be empty")
	}

	if sourceClient == nil {
		return errors.InvalidInputf("source client cannot be nil")
	}

	if destClient == nil {
		return errors.InvalidInputf("destination client cannot be nil")
	}

	return nil
}

// getRepositoriesAndTags retrieves the source and destination repositories and tags
func (r *Reconciler) getRepositoriesAndTags(
	ctx context.Context,
	rule ReplicationRule,
	sourceClient common.RegistryClient,
	destClient common.RegistryClient) (common.Repository, common.Repository, []string, map[string]bool, error) {

	// Get the source repository
	sourceRepo, err := sourceClient.GetRepository(ctx, rule.SourceRepository)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to get source repository")
	}

	// Get the destination repository
	destRepo, err := destClient.GetRepository(ctx, rule.DestinationRepository)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to get destination repository")
	}

	// List all tags in the source repository
	sourceTags, err := sourceRepo.ListTags(ctx)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to list source tags")
	}

	// List all tags in the destination repository
	destTags, err := destRepo.ListTags(ctx)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to list destination tags")
	}

	// Create a map of destination tags for faster lookup
	destTagMap := make(map[string]bool)
	for _, tag := range destTags {
		destTagMap[tag] = true
	}

	return sourceRepo, destRepo, sourceTags, destTagMap, nil
}

// tagsNeedReplica checks if a tag needs to be replicated based on manifests
func (r *Reconciler) tagNeedsReplica(
	ctx context.Context,
	rule ReplicationRule,
	tag string,
	sourceRepo common.Repository,
	destRepo common.Repository,
	destTagMap map[string]bool) (bool, error) {

	// If the tag is not in destination, it needs replication
	if _, exists := destTagMap[tag]; !exists || r.forceUpdate {
		return true, nil
	}

	// Get the source manifest
	sourceManifest, err := sourceRepo.GetManifest(ctx, tag)
	if err != nil {
		r.logger.Warn("Failed to get source manifest, skipping tag", map[string]interface{}{
			"tag":   tag,
			"error": err.Error(),
		})
		return false, err
	}

	// Get the destination manifest
	destManifest, err := destRepo.GetManifest(ctx, tag)
	if err != nil {
		r.logger.Warn("Failed to get destination manifest, will re-copy", map[string]interface{}{
			"tag":   tag,
			"error": err.Error(),
		})
		return true, nil
	}

	// Compare the manifests
	sourceDigest := sourceManifest.Digest
	destDigest := destManifest.Digest

	if sourceDigest == destDigest {
		r.logger.Debug("Skipping tag, already exists with same digest", map[string]interface{}{
			"tag":           tag,
			"source_digest": sourceDigest,
			"dest_digest":   destDigest,
			"source":        rule.SourceRepository,
			"destination":   rule.DestinationRepository,
		})
		return false, nil
	}

	r.logger.Info("Tag exists but has different digest, will re-copy", map[string]interface{}{
		"tag":           tag,
		"source_digest": sourceDigest,
		"dest_digest":   destDigest,
		"source":        rule.SourceRepository,
		"destination":   rule.DestinationRepository,
	})
	return true, nil
}

// processTagFn returns a function to process a single tag replication task
func (r *Reconciler) createTagProcessorFunc(
	ctx context.Context,
	rule ReplicationRule,
	sourceRepo common.Repository,
	destRepo common.Repository,
	tag string,
	wg *sync.WaitGroup,
	copiedTags *int,
	failedTags *int) func(context.Context) error {

	return func(ctx context.Context) error {
		defer wg.Done()

		r.logger.Info("Copying tag", map[string]interface{}{
			"source_repository":      rule.SourceRepository,
			"destination_repository": rule.DestinationRepository,
			"tag":                    tag,
		})

		// Track metrics for this tag copy operation
		if r.metrics != nil {
			r.metrics.TagCopyStarted(rule.SourceRepository, rule.DestinationRepository, tag)
		}

		startTime := time.Now()

		// Get references and options
		srcRef, destRef, srcOpts, destOpts, err := r.prepareReferences(ctx, rule, sourceRepo, destRepo, tag)
		if err != nil {
			*failedTags++
			return err
		}

		// Skip the actual copy in dry run mode
		if r.dryRun {
			r.logger.Info("Dry run - would copy image", map[string]interface{}{
				"source_repo": sourceRepo.GetRepositoryName(),
				"dest_repo":   destRepo.GetRepositoryName(),
				"tag":         tag,
			})
			*copiedTags++

			if r.metrics != nil {
				r.metrics.TagCopyCompleted(rule.SourceRepository, rule.DestinationRepository, tag, 0)
			}

			return nil
		}

		// Perform the copy
		copyResult, err := r.performCopy(ctx, rule, srcRef, destRef, srcOpts, destOpts, tag)
		if err != nil {
			*failedTags++
			return err
		}

		// Log success
		r.logger.Info("Successfully copied tag", map[string]interface{}{
			"source_repository":      rule.SourceRepository,
			"destination_repository": rule.DestinationRepository,
			"tag":                    tag,
			"bytes_transferred":      copyResult.Stats.BytesTransferred,
			"layers":                 copyResult.Stats.Layers,
			"duration":               time.Since(startTime),
		})

		*copiedTags++

		// Update metrics
		if r.metrics != nil {
			r.metrics.TagCopyCompleted(
				rule.SourceRepository,
				rule.DestinationRepository,
				tag,
				copyResult.Stats.BytesTransferred,
			)
		}

		return nil
	}
}

// prepareReferences prepares the source and destination references and options
func (r *Reconciler) prepareReferences(
	ctx context.Context,
	rule ReplicationRule,
	sourceRepo common.Repository,
	destRepo common.Repository,
	tag string) (srcRef, destRef interface{}, srcOpts, destOpts []interface{}, err error) {

	// Get source image reference
	srcRef, err = sourceRepo.GetImageReference(tag)
	if err != nil {
		r.logger.Error("Failed to get source image reference", err, map[string]interface{}{
			"source_repository": rule.SourceRepository,
			"tag":               tag,
		})

		if r.metrics != nil {
			r.metrics.TagCopyFailed(rule.SourceRepository, rule.DestinationRepository, tag)
		}

		return nil, nil, nil, nil, err
	}

	// Get destination image reference
	destRef, err = destRepo.GetImageReference(tag)
	if err != nil {
		r.logger.Error("Failed to get destination image reference", err, map[string]interface{}{
			"dest_repository": rule.DestinationRepository,
			"tag":             tag,
		})

		if r.metrics != nil {
			r.metrics.TagCopyFailed(rule.SourceRepository, rule.DestinationRepository, tag)
		}

		return nil, nil, nil, nil, err
	}

	// Get source and destination options
	srcOpts, err = sourceRepo.GetRemoteOptions()
	if err != nil {
		r.logger.Error("Failed to get source remote options", err, map[string]interface{}{
			"source_repository": rule.SourceRepository,
			"tag":               tag,
		})

		if r.metrics != nil {
			r.metrics.TagCopyFailed(rule.SourceRepository, rule.DestinationRepository, tag)
		}

		return nil, nil, nil, nil, err
	}

	destOpts, err = destRepo.GetRemoteOptions()
	if err != nil {
		r.logger.Error("Failed to get destination remote options", err, map[string]interface{}{
			"dest_repository": rule.DestinationRepository,
			"tag":             tag,
		})

		if r.metrics != nil {
			r.metrics.TagCopyFailed(rule.SourceRepository, rule.DestinationRepository, tag)
		}

		return nil, nil, nil, nil, err
	}

	return srcRef, destRef, srcOpts, destOpts, nil
}

// performCopy performs the actual copy operation
func (r *Reconciler) performCopy(
	ctx context.Context,
	rule ReplicationRule,
	srcRef, destRef interface{},
	srcOpts, destOpts []interface{},
	tag string) (*copy.CopyResult, error) {

	// Set up copy options
	copyOpts := copy.CopyOptions{
		DryRun:         r.dryRun,
		ForceOverwrite: r.forceUpdate,
		Source:         srcRef,
		Destination:    destRef,
	}

	// Perform the copy
	result, err := r.copier.CopyImage(
		ctx,
		srcRef,
		destRef,
		srcOpts,
		destOpts,
		copyOpts,
	)

	if err != nil {
		r.logger.Error("Failed to copy tag", err, map[string]interface{}{
			"source_repository":      rule.SourceRepository,
			"destination_repository": rule.DestinationRepository,
			"tag":                    tag,
		})

		if r.metrics != nil {
			r.metrics.TagCopyFailed(rule.SourceRepository, rule.DestinationRepository, tag)
		}

		return nil, err
	}

	return result, nil
}

// processAndReplicateTags processes and replicates tags from source to destination
func (r *Reconciler) processAndReplicateTags(
	ctx context.Context,
	rule ReplicationRule,
	sourceRepo common.Repository,
	destRepo common.Repository,
	sourceTags []string,
	destTagMap map[string]bool) error {

	var wg sync.WaitGroup
	var totalTags, copiedTags, skippedTags, failedTags int

	// Process each source tag that matches the rule
	for _, tag := range sourceTags {
		if !ShouldReplicate(rule, rule.SourceRepository, tag) {
			skippedTags++
			continue
		}

		totalTags++

		// Check if the tag needs replication
		needsReplica, err := r.tagNeedsReplica(ctx, rule, tag, sourceRepo, destRepo, destTagMap)
		if err != nil || !needsReplica {
			if err != nil {
				r.logger.Debug("Error checking if tag needs replica", map[string]interface{}{
					"tag":   tag,
					"error": err.Error(),
				})
			}
			skippedTags++
			continue
		}

		// Copy the tag to the destination
		wg.Add(1)
		finalTag := tag // Create a copy for the closure

		// Create the tag processor function
		processTagFn := r.createTagProcessorFunc(ctx, rule, sourceRepo, destRepo, finalTag, &wg, &copiedTags, &failedTags)

		// Run the tag processor
		if r.workerPool == nil {
			// Run synchronously if no worker pool
			r.logger.Warn("WorkerPool is nil, running task synchronously", map[string]interface{}{
				"tag": finalTag,
			})
			// Execute the function directly
			go processTagFn(ctx)
		} else {
			// Submit to worker pool if available
			r.workerPool.Submit(processTagFn)
		}
	}

	// Wait for all copy operations to complete
	wg.Wait()

	// Log final summary and update metrics
	r.logCompletionAndUpdateMetrics(rule, totalTags, copiedTags, skippedTags, failedTags)

	return nil
}

// logCompletionAndUpdateMetrics logs the reconciliation summary and updates metrics
func (r *Reconciler) logCompletionAndUpdateMetrics(
	rule ReplicationRule,
	totalTags, copiedTags, skippedTags, failedTags int) {

	// Log final summary
	r.logger.Info("Reconciliation complete", map[string]interface{}{
		"source_repository":      rule.SourceRepository,
		"destination_repository": rule.DestinationRepository,
		"total_tags":             totalTags,
		"copied_tags":            copiedTags,
		"skipped_tags":           skippedTags,
		"failed_tags":            failedTags,
	})

	// Update metrics
	if r.metrics != nil {
		r.metrics.RepositoryCopyCompleted(
			rule.SourceRepository,
			rule.DestinationRepository,
			totalTags,
			copiedTags,
			skippedTags,
			failedTags,
		)
	}
}

// ReconcileAllRepositories reconciles all repositories based on the given rules
func (r *Reconciler) ReconcileAllRepositories(
	ctx context.Context,
	rules []ReplicationRule,
	registryClients map[string]common.RegistryClient) error {

	if ctx == nil {
		return errors.InvalidInputf("context cannot be nil")
	}

	if len(rules) == 0 {
		return errors.InvalidInputf("replication rules cannot be empty")
	}

	if len(registryClients) == 0 {
		return errors.InvalidInputf("registry clients cannot be empty")
	}

	var reconcileErrors []error

	for _, rule := range rules {
		// Get the source client
		sourceClient, ok := registryClients[rule.SourceRegistry]
		if !ok {
			err := errors.NotFoundf("source registry client not found: %s", rule.SourceRegistry)
			reconcileErrors = append(reconcileErrors, err)
			r.logger.Error("Source registry client not found", err, map[string]interface{}{
				"registry": rule.SourceRegistry,
			})
			continue
		}

		// Get the destination client
		destClient, ok := registryClients[rule.DestinationRegistry]
		if !ok {
			err := errors.NotFoundf("destination registry client not found: %s", rule.DestinationRegistry)
			reconcileErrors = append(reconcileErrors, err)
			r.logger.Error("Destination registry client not found", err, map[string]interface{}{
				"registry": rule.DestinationRegistry,
			})
			continue
		}

		// Reconcile the repository
		err := r.ReconcileRepository(ctx, rule, sourceClient, destClient)
		if err != nil {
			reconcileErrors = append(reconcileErrors, err)
			r.logger.Error("Failed to reconcile repository", err, map[string]interface{}{
				"source_registry":        rule.SourceRegistry,
				"source_repository":      rule.SourceRepository,
				"destination_registry":   rule.DestinationRegistry,
				"destination_repository": rule.DestinationRepository,
			})
		}
	}

	if len(reconcileErrors) > 0 {
		return errors.Wrap(reconcileErrors[0], "failed to reconcile one or more repositories")
	}

	return nil
}
