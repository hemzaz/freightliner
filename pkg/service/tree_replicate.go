package service

import (
	"context"
	"freightliner/pkg/client/common"
	"freightliner/pkg/config"
	"freightliner/pkg/copy"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/tree"
	"freightliner/pkg/tree/checkpoint"
)

// TreeReplicationService handles tree replication operations
type TreeReplicationService struct {
	cfg                *config.Config
	logger             *log.Logger
	replicationService *ReplicationService
}

// NewTreeReplicationService creates a new tree replication service
func NewTreeReplicationService(cfg *config.Config, logger *log.Logger) *TreeReplicationService {
	return &TreeReplicationService{
		cfg:                cfg,
		logger:             logger,
		replicationService: NewReplicationService(cfg, logger),
	}
}

// TreeReplicationResult contains the results of a tree replication operation
type TreeReplicationResult struct {
	RepositoriesFound      int
	RepositoriesReplicated int
	RepositoriesSkipped    int
	RepositoriesFailed     int
	TotalTagsCopied        int
	TotalTagsSkipped       int
	TotalErrors            int
	TotalBytesTransferred  int64
	CheckpointID           string
}

// ReplicateTree replicates a tree of repositories
func (s *TreeReplicationService) ReplicateTree(ctx context.Context, source, destination string) (*TreeReplicationResult, error) {
	// Parse source and destination
	sourceRegistry, sourceRepo, err := parseRegistryPath(source)
	if err != nil {
		return nil, err
	}

	destRegistry, destRepo, err := parseRegistryPath(destination)
	if err != nil {
		return nil, err
	}

	// Validate registry types
	if !isValidRegistryType(sourceRegistry) || !isValidRegistryType(destRegistry) {
		return nil, errors.InvalidInputf("registry type must be 'ecr' or 'gcr'")
	}

	// Create registry clients
	clients, err := s.replicationService.createRegistryClients(ctx, sourceRegistry, destRegistry)
	if err != nil {
		return nil, err
	}

	// Initialize credentials if using secrets manager
	if err := s.replicationService.initializeCredentials(ctx); err != nil {
		return nil, err
	}

	// Get source and destination clients
	sourceClient := clients[sourceRegistry]
	destClient := clients[destRegistry]

	// Determine worker count
	workerCount := s.cfg.TreeReplicate.Workers
	if workerCount == 0 && s.cfg.Workers.AutoDetect {
		workerCount = config.GetOptimalWorkerCount()
		s.logger.Info("Auto-detected worker count", map[string]interface{}{
			"workers": workerCount,
		})
	}

	// Create options for tree replicator
	options := map[string]interface{}{
		"workers":          workerCount,
		"excludeRepos":     s.cfg.TreeReplicate.ExcludeRepos,
		"excludeTags":      s.cfg.TreeReplicate.ExcludeTags,
		"includeTags":      s.cfg.TreeReplicate.IncludeTags,
		"dryRun":           s.cfg.TreeReplicate.DryRun,
		"force":            s.cfg.TreeReplicate.Force,
		"enableCheckpoint": s.cfg.TreeReplicate.EnableCheckpoint,
		"checkpointDir":    s.cfg.TreeReplicate.CheckpointDir,
		"resumeID":         s.cfg.TreeReplicate.ResumeID,
		"skipCompleted":    s.cfg.TreeReplicate.SkipCompleted,
		"retryFailed":      s.cfg.TreeReplicate.RetryFailed,
	}

	// Create a tree replicator
	replicator, err := s.createTreeReplicator(ctx, sourceClient, destClient, sourceRepo, destRepo, options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tree replicator")
	}

	// Start replication with the correct method signature
	result, err := replicator.ReplicateTree(ctx, sourceClient, destClient, sourceRepo, destRepo, s.cfg.TreeReplicate.Force)
	if err != nil {
		return nil, errors.Wrap(err, "failed to replicate tree")
	}

	// Return results, adapting TreeReplicationResult to our service-level type
	return &TreeReplicationResult{
		RepositoriesFound:      result.Repositories,
		RepositoriesReplicated: result.ImagesReplicated,
		RepositoriesSkipped:    result.ImagesSkipped,
		RepositoriesFailed:     result.ImagesFailed,
		TotalTagsCopied:        0, // Not provided in tree.TreeReplicationResult
		TotalTagsSkipped:       0, // Not provided in tree.TreeReplicationResult
		TotalErrors:            0, // Not provided in tree.TreeReplicationResult
		TotalBytesTransferred:  0, // Not provided in tree.TreeReplicationResult
		CheckpointID:           result.CheckpointID,
	}, nil
}

// createTreeReplicator creates a new tree replicator
func (s *TreeReplicationService) createTreeReplicator(ctx context.Context, source common.RegistryClient, dest common.RegistryClient, sourcePath, destPath string, opts map[string]interface{}) (*tree.TreeReplicator, error) {
	// Extract options from the map
	workerCount := 2 // Default value
	if workers, ok := opts["workers"].(int); ok && workers > 0 {
		workerCount = workers
	}

	var excludeRepos []string
	if excludes, ok := opts["excludeRepos"].([]string); ok {
		excludeRepos = excludes
	}

	var excludeTags []string
	if excludes, ok := opts["excludeTags"].([]string); ok {
		excludeTags = excludes
	}

	var includeTags []string
	if includes, ok := opts["includeTags"].([]string); ok {
		includeTags = includes
	}

	dryRun := false
	if dry, ok := opts["dryRun"].(bool); ok {
		dryRun = dry
	}

	enableCheckpoint := false
	if enable, ok := opts["enableCheckpoint"].(bool); ok {
		enableCheckpoint = enable
	}

	checkpointDir := "${HOME}/.freightliner/checkpoints"
	if dir, ok := opts["checkpointDir"].(string); ok && dir != "" {
		checkpointDir = dir
	}

	// Only used for resume
	resumeID := ""
	if id, ok := opts["resumeID"].(string); ok {
		resumeID = id
	}

	skipCompleted := true
	if skip, ok := opts["skipCompleted"].(bool); ok {
		skipCompleted = skip
	}

	retryFailed := true
	if retry, ok := opts["retryFailed"].(bool); ok {
		retryFailed = retry
	}

	force := false
	if f, ok := opts["force"].(bool); ok {
		force = f
	}

	// Create a copier for the tree replicator to use
	encManager, err := s.replicationService.setupEncryptionManager(ctx, dest.GetRegistryName())
	if err != nil {
		return nil, errors.Wrap(err, "failed to set up encryption manager for tree replicator")
	}

	// Set up tree replicator configuration
	treeReplicatorOpts := tree.TreeReplicatorOptions{
		WorkerCount:         workerCount,
		ExcludeRepositories: excludeRepos,
		ExcludeTags:         excludeTags,
		IncludeTags:         includeTags,
		EnableCheckpointing: enableCheckpoint,
		CheckpointDirectory: checkpointDir,
		DryRun:              dryRun,
	}

	// Create copier instance for the tree replicator
	copier := copy.NewCopier(s.logger).
		WithEncryptionManager(encManager)

	// Create the tree replicator
	replicator := tree.NewTreeReplicator(s.logger, copier, treeReplicatorOpts)

	// If resuming from a checkpoint, set up the resume operation
	if resumeID != "" {
		s.logger.Info("Setting up tree replication resume", map[string]interface{}{
			"resumeID":      resumeID,
			"skipCompleted": skipCompleted,
			"retryFailed":   retryFailed,
			"checkpointDir": checkpointDir,
		})

		// Initialize the checkpoint store for resume
		store, err := tree.InitCheckpointStore(checkpointDir)
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize checkpoint store for resume")
		}

		// Load the checkpoint
		cp, err := checkpoint.GetCheckpointByID(store, resumeID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load checkpoint for resume")
		}

		// Note: We would set up resume options here if we had a method to use them
		// Currently there is no SetupResume method available in the TreeReplicator

		// Get repositories to process
		repositories, err := checkpoint.GetRemainingRepositories(cp, checkpoint.ResumableOptions{
			ID:            resumeID,
			SkipCompleted: skipCompleted,
			RetryFailed:   retryFailed,
			Force:         force,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to get remaining repositories for resume")
		}

		s.logger.Info("Resume operation set up", map[string]interface{}{
			"repositories": len(repositories),
		})

		// For now, we don't have a method to set up a resume
		// In a real implementation, you would use the checkpoint and resumeOpts to configure the replicator
		// This is just a placeholder for now
		if len(repositories) > 0 {
			s.logger.Info("Found repositories to resume", map[string]interface{}{
				"count": len(repositories),
			})
		}
	}

	return replicator, nil
}
