package replication

import (
	"context"
	"freightliner/pkg/client/common"
	"freightliner/pkg/copy"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/metrics"
	"sync"
	"time"
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

	// Input validation
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

	// Get the source repository
	sourceRepo, err := sourceClient.GetRepository(ctx, rule.SourceRepository)
	if err != nil {
		return errors.Wrap(err, "failed to get source repository")
	}

	// Get the destination repository
	destRepo, err := destClient.GetRepository(ctx, rule.DestinationRepository)
	if err != nil {
		return errors.Wrap(err, "failed to get destination repository")
	}

	// List all tags in the source repository
	sourceTags, err := sourceRepo.ListTags()
	if err != nil {
		return errors.Wrap(err, "failed to list source tags")
	}

	// List all tags in the destination repository
	destTags, err := destRepo.ListTags()
	if err != nil {
		return errors.Wrap(err, "failed to list destination tags")
	}

	// Create a map of destination tags for faster lookup
	destTagMap := make(map[string]bool)
	for _, tag := range destTags {
		destTagMap[tag] = true
	}

	// Process each source tag that matches the rule
	var wg sync.WaitGroup
	var totalTags, copiedTags, skippedTags, failedTags int

	// Process each source tag that matches the rule
	for _, tag := range sourceTags {
		if !ShouldReplicate(rule, rule.SourceRepository, tag) {
			skippedTags++
			continue
		}

		totalTags++

		// If the tag already exists in the destination, check if it's the same image
		if _, exists := destTagMap[tag]; exists && !r.forceUpdate {
			// Get the source manifest
			sourceManifest, err := sourceRepo.GetManifest(ctx, tag)
			if err != nil {
				r.logger.Warn("Failed to get source manifest, skipping tag", map[string]interface{}{
					"tag":   tag,
					"error": err.Error(),
				})
				skippedTags++
				continue
			}

			// Get the destination manifest
			destManifest, err := destRepo.GetManifest(ctx, tag)
			if err != nil {
				r.logger.Warn("Failed to get destination manifest, will re-copy", map[string]interface{}{
					"tag":   tag,
					"error": err.Error(),
				})
			} else {
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
					skippedTags++
					continue
				}

				r.logger.Info("Tag exists but has different digest, will re-copy", map[string]interface{}{
					"tag":           tag,
					"source_digest": sourceDigest,
					"dest_digest":   destDigest,
					"source":        rule.SourceRepository,
					"destination":   rule.DestinationRepository,
				})
			}
		}

		// Copy the tag to the destination
		wg.Add(1)
		finalTag := tag // Create a copy for the closure

		r.workerPool.Submit(func(ctx context.Context) error {
			defer wg.Done()

			r.logger.Info("Copying tag", map[string]interface{}{
				"source_repository":      rule.SourceRepository,
				"destination_repository": rule.DestinationRepository,
				"tag":                    finalTag,
			})

			// Track metrics for this tag copy operation
			if r.metrics != nil {
				r.metrics.TagCopyStarted(rule.SourceRepository, rule.DestinationRepository, finalTag)
			}

			startTime := time.Now()

			// Get source image reference
			srcRef, err := sourceRepo.GetImageReference(finalTag)
			if err != nil {
				r.logger.Error("Failed to get source image reference", err, map[string]interface{}{
					"source_repository": rule.SourceRepository,
					"tag":               finalTag,
				})
				failedTags++

				if r.metrics != nil {
					r.metrics.TagCopyFailed(rule.SourceRepository, rule.DestinationRepository, finalTag)
				}

				return err
			}

			// Get destination image reference
			destRef, err := destRepo.GetImageReference(finalTag)
			if err != nil {
				r.logger.Error("Failed to get destination image reference", err, map[string]interface{}{
					"dest_repository": rule.DestinationRepository,
					"tag":             finalTag,
				})
				failedTags++

				if r.metrics != nil {
					r.metrics.TagCopyFailed(rule.SourceRepository, rule.DestinationRepository, finalTag)
				}

				return err
			}

			// Skip the actual copy in dry run mode
			if r.dryRun {
				r.logger.Info("Dry run - would copy image", map[string]interface{}{
					"source_repo": sourceRepo.GetRepositoryName(),
					"dest_repo":   destRepo.GetRepositoryName(),
					"tag":         finalTag,
				})
				copiedTags++

				if r.metrics != nil {
					r.metrics.TagCopyCompleted(rule.SourceRepository, rule.DestinationRepository, finalTag, 0)
				}

				return nil
			}

			// Get source and destination options
			srcOpts, err := sourceRepo.GetRemoteOptions()
			if err != nil {
				r.logger.Error("Failed to get source remote options", err, map[string]interface{}{
					"source_repository": rule.SourceRepository,
					"tag":               finalTag,
				})
				failedTags++

				if r.metrics != nil {
					r.metrics.TagCopyFailed(rule.SourceRepository, rule.DestinationRepository, finalTag)
				}

				return err
			}

			destOpts, err := destRepo.GetRemoteOptions()
			if err != nil {
				r.logger.Error("Failed to get destination remote options", err, map[string]interface{}{
					"dest_repository": rule.DestinationRepository,
					"tag":             finalTag,
				})
				failedTags++

				if r.metrics != nil {
					r.metrics.TagCopyFailed(rule.SourceRepository, rule.DestinationRepository, finalTag)
				}

				return err
			}

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
					"tag":                    finalTag,
				})
				failedTags++

				if r.metrics != nil {
					r.metrics.TagCopyFailed(rule.SourceRepository, rule.DestinationRepository, finalTag)
				}

				return err
			}

			// Log success
			r.logger.Info("Successfully copied tag", map[string]interface{}{
				"source_repository":      rule.SourceRepository,
				"destination_repository": rule.DestinationRepository,
				"tag":                    finalTag,
				"bytes_transferred":      result.Stats.BytesTransferred,
				"layers":                 result.Stats.Layers,
				"duration":               time.Since(startTime),
			})

			copiedTags++

			// Update metrics
			if r.metrics != nil {
				r.metrics.TagCopyCompleted(
					rule.SourceRepository,
					rule.DestinationRepository,
					finalTag,
					result.Stats.BytesTransferred,
				)
			}

			return nil
		})
	}

	// Wait for all copy operations to complete
	r.workerPool.Wait()

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

	return nil
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
