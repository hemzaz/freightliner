package replication

import (
	"context"
	"fmt"
	"sync"
	
	"github.com/hemzaz/freightliner/internal/log"
	"github.com/hemzaz/freightliner/pkg/client/common"
	"github.com/hemzaz/freightliner/pkg/copy"
)

// Reconciler handles reconciling repositories between registries
type Reconciler struct {
	logger     *log.Logger
	copier     *copy.Copier
	workerPool *WorkerPool
}

// NewReconciler creates a new reconciler
func NewReconciler(logger *log.Logger, copier *copy.Copier, workerPool *WorkerPool) *Reconciler {
	return &Reconciler{
		logger:     logger,
		copier:     copier,
		workerPool: workerPool,
	}
}

// ReconcileRepository reconciles a source repository with a destination repository
func (r *Reconciler) ReconcileRepository(
	ctx context.Context,
	rule ReplicationRule,
	sourceClient common.RegistryClient,
	destClient common.RegistryClient) error {
	
	// Get the source repository
	sourceRepo, err := sourceClient.GetRepository(rule.SourceRepository)
	if err != nil {
		return fmt.Errorf("failed to get source repository: %w", err)
	}
	
	// Get the destination repository
	destRepo, err := destClient.GetRepository(rule.DestinationRepository)
	if err != nil {
		return fmt.Errorf("failed to get destination repository: %w", err)
	}
	
	// List all tags in the source repository
	sourceTags, err := sourceRepo.ListTags()
	if err != nil {
		return fmt.Errorf("failed to list source tags: %w", err)
	}
	
	// List all tags in the destination repository
	destTags, err := destRepo.ListTags()
	if err != nil {
		return fmt.Errorf("failed to list destination tags: %w", err)
	}
	
	// Create a map of destination tags for faster lookup
	destTagMap := make(map[string]bool)
	for _, tag := range destTags {
		destTagMap[tag] = true
	}
	
	// Process each source tag that matches the rule
	var wg sync.WaitGroup
	for _, tag := range sourceTags {
		if !ShouldReplicate(rule, rule.SourceRepository, tag) {
			continue
		}
		
		// Skip if the tag already exists in the destination
		if _, exists := destTagMap[tag]; exists {
			// TODO: Implement SHA verification to check if the image is the same
			continue
		}
		
		// Copy the tag to the destination
		wg.Add(1)
		finalTag := tag // Create a copy for the closure
		
		r.workerPool.Submit(func(ctx context.Context) error {
			defer wg.Done()
			
			r.logger.Info("Copying tag", map[string]interface{}{
				"source_repository": rule.SourceRepository,
				"destination_repository": rule.DestinationRepository,
				"tag": finalTag,
			})
			
			options := copy.CopyOptions{
				SourceTag:      finalTag,
				DestinationTag: finalTag,
				ForceOverwrite: false,
			}
			
			err := r.copier.CopyImage(ctx, sourceRepo, destRepo, options)
			if err != nil {
				r.logger.Error("Failed to copy tag", err, map[string]interface{}{
					"source_repository": rule.SourceRepository,
					"destination_repository": rule.DestinationRepository,
					"tag": finalTag,
				})
			}
			
			return err
		})
	}
	
	// Wait for all copy operations to complete
	r.workerPool.Wait()
	
	return nil
}

// ReconcileAllRepositories reconciles all repositories based on the given rules
func (r *Reconciler) ReconcileAllRepositories(
	ctx context.Context,
	rules []ReplicationRule,
	registryClients map[string]common.RegistryClient) error {
	
	for _, rule := range rules {
		// Get the source client
		sourceClient, ok := registryClients[rule.SourceRegistry]
		if !ok {
			r.logger.Error("Source registry client not found", nil, map[string]interface{}{
				"registry": rule.SourceRegistry,
			})
			continue
		}
		
		// Get the destination client
		destClient, ok := registryClients[rule.DestinationRegistry]
		if !ok {
			r.logger.Error("Destination registry client not found", nil, map[string]interface{}{
				"registry": rule.DestinationRegistry,
			})
			continue
		}
		
		// Reconcile the repository
		err := r.ReconcileRepository(ctx, rule, sourceClient, destClient)
		if err != nil {
			r.logger.Error("Failed to reconcile repository", err, map[string]interface{}{
				"source_registry": rule.SourceRegistry,
				"source_repository": rule.SourceRepository,
				"destination_registry": rule.DestinationRegistry,
				"destination_repository": rule.DestinationRepository,
			})
		}
	}
	
	return nil
}
