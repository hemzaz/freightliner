package service

import (
	"context"

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
	logger             log.Logger
	replicationService ReplicationService
}

// NewTreeReplicationService creates a new tree replication service
func NewTreeReplicationService(cfg *config.Config, logger log.Logger) *TreeReplicationService {
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

// TreeReplicationOptions contains options for tree replication
type TreeReplicationOptions struct {
	// Source and destination paths
	Source      string
	Destination string

	// Worker configuration
	WorkerCount int

	// Filtering options
	ExcludeRepos []string
	ExcludeTags  []string
	IncludeTags  []string

	// Operation behavior
	DryRun bool
	Force  bool

	// Checkpoint configuration
	EnableCheckpoint bool
	CheckpointDir    string

	// Resume options
	ResumeID      string
	SkipCompleted bool
	RetryFailed   bool
}

// ReplicateTree replicates a tree of repositories
func (s *TreeReplicationService) ReplicateTree(ctx context.Context, source, destination string) (*TreeReplicationResult, error) {
	// Create options struct with values from config
	options := TreeReplicationOptions{
		Source:           source,
		Destination:      destination,
		WorkerCount:      s.cfg.TreeReplicate.Workers,
		ExcludeRepos:     s.cfg.TreeReplicate.ExcludeRepos,
		ExcludeTags:      s.cfg.TreeReplicate.ExcludeTags,
		IncludeTags:      s.cfg.TreeReplicate.IncludeTags,
		DryRun:           s.cfg.TreeReplicate.DryRun,
		Force:            s.cfg.TreeReplicate.Force,
		EnableCheckpoint: s.cfg.TreeReplicate.EnableCheckpoint,
		CheckpointDir:    s.cfg.TreeReplicate.CheckpointDir,
		ResumeID:         s.cfg.TreeReplicate.ResumeID,
		SkipCompleted:    s.cfg.TreeReplicate.SkipCompleted,
		RetryFailed:      s.cfg.TreeReplicate.RetryFailed,
	}

	// Parse source and destination
	sourceRegistry, sourceRepo, err := parseRegistryPath(options.Source)
	if err != nil {
		return nil, err
	}

	destRegistry, destRepo, err := parseRegistryPath(options.Destination)
	if err != nil {
		return nil, err
	}

	// Create registry clients - need to access implementation methods
	replicationSvc, ok := s.replicationService.(*replicationService)
	if !ok {
		return nil, errors.InvalidInputf("replication service must be concrete implementation for tree replication")
	}

	// Validate registry types (now supports ALL Docker v2 registries)
	if !replicationSvc.isValidRegistryType(sourceRegistry) {
		return nil, errors.InvalidInputf("invalid source registry '%s'. Registry cannot be empty", sourceRegistry)
	}
	if !replicationSvc.isValidRegistryType(destRegistry) {
		return nil, errors.InvalidInputf("invalid destination registry '%s'. Registry cannot be empty", destRegistry)
	}

	clients, err := replicationSvc.createRegistryClients(ctx, sourceRegistry, destRegistry)
	if err != nil {
		return nil, err
	}

	// Initialize credentials if using secrets manager
	if initErr := replicationSvc.initializeCredentials(ctx); initErr != nil {
		return nil, initErr
	}

	// Get source and destination clients
	sourceClient := clients[sourceRegistry]
	destClient := clients[destRegistry]

	// Auto-detect worker count if configured
	if options.WorkerCount == 0 && s.cfg.Workers.AutoDetect {
		options.WorkerCount = config.GetOptimalWorkerCount()
		s.logger.WithFields(map[string]interface{}{
			"workers": options.WorkerCount,
		}).Info("Auto-detected worker count")
	}

	// Create options map for tree replicator (for backward compatibility)
	optionsMap := map[string]interface{}{
		"workers":          options.WorkerCount,
		"excludeRepos":     options.ExcludeRepos,
		"excludeTags":      options.ExcludeTags,
		"includeTags":      options.IncludeTags,
		"dryRun":           options.DryRun,
		"force":            options.Force,
		"enableCheckpoint": options.EnableCheckpoint,
		"checkpointDir":    options.CheckpointDir,
		"resumeID":         options.ResumeID,
		"skipCompleted":    options.SkipCompleted,
		"retryFailed":      options.RetryFailed,
	}

	// Create a tree replicator
	replicator, err := s.createTreeReplicator(ctx, sourceClient, destClient, sourceRepo, destRepo, optionsMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tree replicator")
	}

	// Set up the replicate tree options
	replicateOpts := tree.ReplicateTreeOptions{
		SourceClient:              sourceClient,
		DestClient:                destClient,
		SourcePrefix:              sourceRepo,
		DestPrefix:                destRepo,
		ForceOverwrite:            options.Force,
		ResumeFromCheckpoint:      options.ResumeID,
		SkipCompletedRepositories: options.SkipCompleted,
	}

	// Start replication with the options
	result, err := replicator.ReplicateTree(ctx, replicateOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to replicate tree")
	}

	// Return results, adapting TreeReplicationResult to our service-level type
	return &TreeReplicationResult{
		RepositoriesFound:      result.Repositories,
		RepositoriesReplicated: int(result.ImagesReplicated.Load()),
		RepositoriesSkipped:    int(result.ImagesSkipped.Load()),
		RepositoriesFailed:     int(result.ImagesFailed.Load()),
		TotalTagsCopied:        0, // Not provided in tree.TreeReplicationResult
		TotalTagsSkipped:       0, // Not provided in tree.TreeReplicationResult
		TotalErrors:            0, // Not provided in tree.TreeReplicationResult
		TotalBytesTransferred:  0, // Not provided in tree.TreeReplicationResult
		CheckpointID:           result.CheckpointID,
	}, nil
}

// TreeReplicatorCreationOptions holds all options for creating a tree replicator
type TreeReplicatorCreationOptions struct {
	// Worker configuration
	WorkerCount int

	// Filtering options
	ExcludeRepos []string
	ExcludeTags  []string
	IncludeTags  []string

	// Operation behavior
	DryRun bool
	Force  bool

	// Checkpoint configuration
	EnableCheckpoint bool
	CheckpointDir    string

	// Resume options
	ResumeID      string
	SkipCompleted bool
	RetryFailed   bool
}

// DefaultTreeReplicatorCreationOptions returns sensible defaults
func DefaultTreeReplicatorCreationOptions() TreeReplicatorCreationOptions {
	return TreeReplicatorCreationOptions{
		WorkerCount:      2,
		ExcludeRepos:     []string{},
		ExcludeTags:      []string{},
		IncludeTags:      []string{},
		DryRun:           false,
		Force:            false,
		EnableCheckpoint: false,
		CheckpointDir:    "${HOME}/.freightliner/checkpoints",
		ResumeID:         "",
		SkipCompleted:    true,
		RetryFailed:      true,
	}
}

// createTreeReplicator creates a new tree replicator
func (s *TreeReplicationService) createTreeReplicator(ctx context.Context, source RegistryClient, dest RegistryClient, sourcePath, destPath string, opts map[string]interface{}) (*tree.TreeReplicator, error) {
	// Create options with defaults
	options := DefaultTreeReplicatorCreationOptions()

	// Extract options from the map
	if workers, ok := opts["workers"].(int); ok && workers > 0 {
		options.WorkerCount = workers
	}

	if excludes, ok := opts["excludeRepos"].([]string); ok {
		options.ExcludeRepos = excludes
	}

	if excludes, ok := opts["excludeTags"].([]string); ok {
		options.ExcludeTags = excludes
	}

	if includes, ok := opts["includeTags"].([]string); ok {
		options.IncludeTags = includes
	}

	if dry, ok := opts["dryRun"].(bool); ok {
		options.DryRun = dry
	}

	if enable, ok := opts["enableCheckpoint"].(bool); ok {
		options.EnableCheckpoint = enable
	}

	if dir, ok := opts["checkpointDir"].(string); ok && dir != "" {
		options.CheckpointDir = dir
	}

	if id, ok := opts["resumeID"].(string); ok {
		options.ResumeID = id
	}

	if skip, ok := opts["skipCompleted"].(bool); ok {
		options.SkipCompleted = skip
	}

	if retry, ok := opts["retryFailed"].(bool); ok {
		options.RetryFailed = retry
	}

	if f, ok := opts["force"].(bool); ok {
		options.Force = f
	}

	// Create a copier for the tree replicator to use
	replicationSvc, ok := s.replicationService.(*replicationService)
	if !ok {
		return nil, errors.InvalidInputf("replication service must be concrete implementation for encryption setup")
	}

	encManager, err := replicationSvc.setupEncryptionManager(ctx, dest.GetRegistryName())
	if err != nil {
		return nil, errors.Wrap(err, "failed to set up encryption manager for tree replicator")
	}

	// Set up tree replicator configuration
	treeReplicatorOpts := tree.TreeReplicatorOptions{
		WorkerCount:         options.WorkerCount,
		ExcludeRepositories: options.ExcludeRepos,
		ExcludeTags:         options.ExcludeTags,
		IncludeTags:         options.IncludeTags,
		EnableCheckpointing: options.EnableCheckpoint,
		CheckpointDirectory: options.CheckpointDir,
		DryRun:              options.DryRun,
	}

	// Create copier instance for the tree replicator
	copier := copy.NewCopier(s.logger).
		WithEncryptionManager(encManager)

	// Create the tree replicator
	replicator := tree.NewTreeReplicator(s.logger, copier, treeReplicatorOpts)

	// If resuming from a checkpoint, set up the resume operation
	if options.ResumeID != "" {
		s.logger.WithFields(map[string]interface{}{
			"resumeID":      options.ResumeID,
			"skipCompleted": options.SkipCompleted,
			"retryFailed":   options.RetryFailed,
			"checkpointDir": options.CheckpointDir,
		}).Info("Setting up tree replication resume")

		// Initialize the checkpoint store for resume
		store, err := tree.InitCheckpointStore(options.CheckpointDir)
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize checkpoint store for resume")
		}

		// Load the checkpoint
		cp, err := checkpoint.GetCheckpointByID(store, options.ResumeID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load checkpoint for resume")
		}

		// Note: We would set up resume options here if we had a method to use them
		// Currently there is no SetupResume method available in the TreeReplicator

		// Get repositories to process
		repositories, err := checkpoint.GetRemainingRepositories(cp, checkpoint.ResumableOptions{
			ID:            options.ResumeID,
			SkipCompleted: options.SkipCompleted,
			RetryFailed:   options.RetryFailed,
			Force:         options.Force,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to get remaining repositories for resume")
		}

		s.logger.WithFields(map[string]interface{}{
			"repositories": len(repositories),
		}).Info("Resume operation set up")

		// For now, we don't have a method to set up a resume
		// In a real implementation, you would use the checkpoint and resumeOpts to configure the replicator
		// This is just a placeholder for now
		if len(repositories) > 0 {
			s.logger.WithFields(map[string]interface{}{
				"count": len(repositories),
			}).Info("Found repositories to resume")
		}
	}

	return replicator, nil
}
