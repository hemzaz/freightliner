package cmd

import (
	"context"
	"fmt"
	"time"

	"freightliner/pkg/client"
	"freightliner/pkg/client/generic"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/sync"

	"github.com/spf13/cobra"
)

var (
	syncConfigFile string
	syncDryRun     bool
	syncParallel   int
)

// newSyncCmd creates the sync command
func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync --config FILE",
		Short: "Sync images between registries using YAML configuration",
		Long: `Sync multiple images between registries based on a YAML configuration file.

This command enables bulk operations with advanced filtering and parallel execution.

Configuration file format:
  source:
    registry: "registry-1.docker.io"
    type: "docker"  # Auto-detected if not specified
    auth:
      username: "user"
      password: "pass"

  destination:
    registry: "my-registry.io"
    type: "generic"
    auth:
      username: "user"
      password: "pass"

  parallel: 5
  batch_size: 10
  retry_attempts: 3
  retry_backoff: 2
  enable_deduplication: true
  enable_http3: true

  images:
    - repository: "library/nginx"
      tags: ["latest", "1.21", "1.22"]

    - repository: "library/redis"
      tag_regex: "^7\\..*"
      destination_repository: "cache/redis"

    - repository: "library/postgres"
      all_tags: true
      limit: 10

    - repository: "library/ubuntu"
      semver_constraint: ">=20.04"
      latest_n: 5

Examples:
  # Sync using configuration file
  freightliner sync --config sync.yaml

  # Dry run to see what would be synced
  freightliner sync --config sync.yaml --dry-run

  # Override parallelism
  freightliner sync --config sync.yaml --parallel 10
`,
		RunE: runSync,
	}

	cmd.Flags().StringVar(&syncConfigFile, "config", "", "Path to sync configuration file (required)")
	cmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Show what would be synced without actually syncing")
	cmd.Flags().IntVar(&syncParallel, "parallel", 0, "Override parallel workers from config (default: from config or 3)")

	cmd.MarkFlagRequired("config")

	return cmd
}

// runSync executes the sync command
func runSync(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger, ctx, cancel := setupCommand(ctx)
	defer cancel()

	// Load sync configuration using pkg/sync
	syncConfig, err := sync.LoadConfig(syncConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := syncConfig.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Override parallel workers if specified
	if syncParallel > 0 {
		syncConfig.Parallel = syncParallel
	}

	logger.WithFields(map[string]interface{}{
		"source":      syncConfig.Source.Registry,
		"destination": syncConfig.Destination.Registry,
		"parallel":    syncConfig.Parallel,
		"dry-run":     syncDryRun,
	}).Info("Starting sync operation")

	// Build list of sync tasks
	syncTasks, err := buildSyncTasks(ctx, logger, syncConfig)
	if err != nil {
		return fmt.Errorf("failed to build sync tasks: %w", err)
	}

	if len(syncTasks) == 0 {
		fmt.Println("No images to sync")
		return nil
	}

	logger.WithFields(map[string]interface{}{
		"count": len(syncTasks),
	}).Info("Found images to sync")
	fmt.Printf("Found %d images to sync\n\n", len(syncTasks))

	// Display tasks if dry run
	if syncDryRun {
		fmt.Println("Dry run - would sync the following images:")
		for _, task := range syncTasks {
			fmt.Printf("  %s/%s:%s -> %s/%s:%s\n",
				task.SourceRegistry, task.SourceRepository, task.SourceTag,
				task.DestRegistry, task.DestRepository, task.DestTag)
		}
		return nil
	}

	// Create client factory (use global cfg or create minimal one)
	var factoryCfg *config.Config
	if cfg != nil {
		factoryCfg = cfg
	} else {
		// Create minimal config for factory if global config not loaded
		factoryCfg = &config.Config{
			Registries: config.RegistriesConfig{
				Registries: []config.RegistryConfig{},
			},
		}
	}
	factory := client.NewFactory(factoryCfg, logger)

	// Execute sync tasks using batch executor with factory
	executor := sync.NewBatchExecutorWithFactory(syncConfig, logger, factory)
	results, err := executor.Execute(ctx, syncTasks)
	if err != nil {
		return fmt.Errorf("batch execution failed: %w", err)
	}

	// Display results
	displaySyncResults(results)

	// Check for failures
	failCount := 0
	for _, result := range results {
		if !result.Success {
			failCount++
		}
	}

	if failCount > 0 {
		return fmt.Errorf("sync failed for %d images", failCount)
	}

	return nil
}

// buildSyncTasks builds a list of sync tasks from the configuration
func buildSyncTasks(ctx context.Context, logger log.Logger, config *sync.Config) ([]sync.SyncTask, error) {
	var tasks []sync.SyncTask

	for _, imageSync := range config.Images {
		// Resolve tags using the appropriate filter
		tags, err := resolveTags(ctx, logger, &config.Source, imageSync)
		if err != nil {
			logger.WithFields(map[string]interface{}{
				"repository": imageSync.Repository,
			}).Error("Failed to resolve tags", err)
			continue
		}

		// Apply limit if specified
		if imageSync.LatestN > 0 && len(tags) > imageSync.LatestN {
			tags = tags[:imageSync.LatestN]
		}

		// Create sync tasks
		for _, tag := range tags {
			destRepo := imageSync.Repository
			if imageSync.DestinationRepository != "" {
				destRepo = imageSync.DestinationRepository
			}

			destTag := tag
			if imageSync.DestinationPrefix != "" {
				destTag = imageSync.DestinationPrefix + tag
			}

			tasks = append(tasks, sync.SyncTask{
				SourceRegistry:   config.Source.Registry,
				SourceRepository: imageSync.Repository,
				SourceTag:        tag,
				DestRegistry:     config.Destination.Registry,
				DestRepository:   destRepo,
				DestTag:          destTag,
			})
		}
	}

	return tasks, nil
}

// resolveTags resolves the list of tags to sync based on the ImageSync configuration
func resolveTags(ctx context.Context, logger log.Logger, source *sync.RegistryConfig, imageSync sync.ImageSync) ([]string, error) {
	logger.WithFields(map[string]interface{}{
		"repository": imageSync.Repository,
	}).Info("Resolving tags")

	// If specific tags are listed, return them
	if len(imageSync.Tags) > 0 {
		return imageSync.Tags, nil
	}

	// Otherwise, we need to list tags from the registry
	// Convert sync.RegistryConfig to config.RegistryConfig
	registryConfig := convertToConfigRegistryConfig(source)

	// Create a generic client
	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: registryConfig,
		RegistryName:   source.Registry,
		Logger:         logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create registry client: %w", err)
	}

	// Get repository
	repo, err := client.GetRepository(ctx, imageSync.Repository)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// List all tags
	allTags, err := repo.ListTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"repository": imageSync.Repository,
		"totalTags":  len(allTags),
	}).Info("Listed tags from registry")

	// Apply filters
	filter, err := sync.NewTagFilter(imageSync)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag filter: %w", err)
	}

	// Filter tags
	filteredTags := filter.Filter(allTags)

	logger.WithFields(map[string]interface{}{
		"repository":   imageSync.Repository,
		"totalTags":    len(allTags),
		"filteredTags": len(filteredTags),
	}).Info("Applied tag filters")

	// Apply limit if specified
	if imageSync.Limit > 0 {
		filteredTags = sync.ApplyLimit(filteredTags, imageSync.Limit)
		logger.WithFields(map[string]interface{}{
			"repository": imageSync.Repository,
			"limit":      imageSync.Limit,
			"finalCount": len(filteredTags),
		}).Info("Applied tag limit")
	}

	return filteredTags, nil
}

// convertToConfigRegistryConfig converts sync.RegistryConfig to config.RegistryConfig
func convertToConfigRegistryConfig(src *sync.RegistryConfig) config.RegistryConfig {
	cfg := config.RegistryConfig{
		Name:      src.Registry,
		Type:      mapRegistryType(src.Type),
		Endpoint:  src.Registry,
		Insecure:  src.Insecure,
		Region:    src.Region,
		Project:   src.Project,
		AccountID: src.Account,
	}

	// Convert authentication
	if src.Auth != nil {
		cfg.Auth = config.AuthConfig{}

		// Determine auth type
		if src.Auth.Username != "" && src.Auth.Password != "" {
			cfg.Auth.Type = config.AuthTypeBasic
			cfg.Auth.Username = src.Auth.Username
			cfg.Auth.Password = src.Auth.Password
		} else if src.Auth.Token != "" {
			cfg.Auth.Type = config.AuthTypeToken
			cfg.Auth.Token = src.Auth.Token
		} else if src.Auth.UseDockerConfig {
			cfg.Auth.Type = config.AuthTypeBasic
			// Generic client will handle docker config
		} else {
			cfg.Auth.Type = config.AuthTypeAnonymous
		}

		// AWS/GCP specific
		if src.Auth.AWSProfile != "" {
			cfg.Auth.Type = config.AuthTypeAWS
			cfg.Auth.Profile = src.Auth.AWSProfile
		}
		if src.Auth.GCPCredentials != "" {
			cfg.Auth.Type = config.AuthTypeGCP
			cfg.Auth.CredentialsFile = src.Auth.GCPCredentials
		}
	} else {
		cfg.Auth = config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		}
	}

	return cfg
}

// mapRegistryType maps sync registry type string to config.RegistryType
func mapRegistryType(typeStr string) config.RegistryType {
	switch typeStr {
	case "ecr":
		return config.RegistryTypeECR
	case "gcr":
		return config.RegistryTypeGCR
	case "docker", "dockerhub":
		return config.RegistryTypeDockerHub
	case "harbor":
		return config.RegistryTypeHarbor
	case "quay":
		return config.RegistryTypeQuay
	case "gitlab":
		return config.RegistryTypeGitLab
	case "github", "ghcr":
		return config.RegistryTypeGitHub
	case "acr", "azure":
		return config.RegistryTypeAzure
	default:
		return config.RegistryTypeGeneric
	}
}

// displaySyncResults displays the sync results summary
func displaySyncResults(results []sync.SyncResult) {
	successCount := 0
	failCount := 0
	var totalDuration int64
	var totalBytes int64

	for _, result := range results {
		totalDuration += result.Duration
		totalBytes += result.BytesCopied
		if result.Success {
			successCount++
		} else {
			failCount++
		}
	}

	fmt.Printf("\nSync Summary:\n")
	fmt.Printf("  Total: %d\n", len(results))
	fmt.Printf("  Success: %d\n", successCount)
	fmt.Printf("  Failed: %d\n", failCount)
	fmt.Printf("  Total Duration: %s\n", time.Duration(totalDuration)*time.Millisecond)
	fmt.Printf("  Total Bytes: %s\n", formatBytes(totalBytes))

	if failCount > 0 {
		fmt.Println("\nFailed syncs:")
		for _, result := range results {
			if !result.Success {
				srcRef := fmt.Sprintf("%s/%s:%s", result.Task.SourceRegistry, result.Task.SourceRepository, result.Task.SourceTag)
				dstRef := fmt.Sprintf("%s/%s:%s", result.Task.DestRegistry, result.Task.DestRepository, result.Task.DestTag)
				errMsg := "unknown error"
				if result.Error != nil {
					errMsg = result.Error.Error()
				}
				fmt.Printf("  %s -> %s: %s\n", srcRef, dstRef, errMsg)
			}
		}
	}
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
