package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hemzaz/freightliner/src/internal/log"
	"github.com/hemzaz/freightliner/src/pkg/client/common"
	"github.com/hemzaz/freightliner/src/pkg/client/ecr"
	"github.com/hemzaz/freightliner/src/pkg/client/gcr"
	"github.com/hemzaz/freightliner/src/pkg/copy"
	"github.com/hemzaz/freightliner/src/pkg/replication"
	"github.com/hemzaz/freightliner/src/pkg/tree"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "freightliner",
		Short: "Container registry replication tool",
		Long:  `Freightliner is a tool for replicating container images between different container registries.`,
	}

	// Global flags
	logLevel string

	// Tree replication flags
	treeReplicateWorkers     int
	treeReplicateExcludeRepos []string
	treeReplicateExcludeTags []string
	treeReplicateIncludeTags []string
	treeReplicateDryRun      bool
	treeReplicateForce       bool
	ecrRegion                string
	ecrAccountID             string
	gcrProject               string
	gcrLocation              string
)

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().StringVar(&ecrRegion, "ecr-region", "us-west-2", "AWS region for ECR")
	rootCmd.PersistentFlags().StringVar(&ecrAccountID, "ecr-account", "", "AWS account ID for ECR (empty uses default from credentials)")
	rootCmd.PersistentFlags().StringVar(&gcrProject, "gcr-project", "", "GCP project for GCR")
	rootCmd.PersistentFlags().StringVar(&gcrLocation, "gcr-location", "us", "GCR location (us, eu, asia)")

	// Add tree replication flags
	replicateTreeCmd.Flags().IntVar(&treeReplicateWorkers, "workers", 5, "Number of concurrent worker threads")
	replicateTreeCmd.Flags().StringSliceVar(&treeReplicateExcludeRepos, "exclude-repo", []string{}, "Repository patterns to exclude (e.g. 'internal-*')")
	replicateTreeCmd.Flags().StringSliceVar(&treeReplicateExcludeTags, "exclude-tag", []string{}, "Tag patterns to exclude (e.g. 'dev-*')")
	replicateTreeCmd.Flags().StringSliceVar(&treeReplicateIncludeTags, "include-tag", []string{}, "Tag patterns to include (e.g. 'v*')")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateDryRun, "dry-run", false, "Perform a dry run without actually copying images")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateForce, "force", false, "Force overwrite of existing images")

	// Add commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(replicateCmd)
	rootCmd.AddCommand(replicateTreeCmd)
	rootCmd.AddCommand(serveCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of freightliner",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Freightliner v0.1.0") // TODO: Use build version from goreleaser
	},
}

var replicateCmd = &cobra.Command{
	Use:   "replicate [source-registry]/[source-repository] [destination-registry]/[destination-repository]",
	Short: "Replicate a repository from one registry to another",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger := createLogger(logLevel)

		// Parse source and destination
		sourceParts := strings.SplitN(args[0], "/", 2)
		destParts := strings.SplitN(args[1], "/", 2)

		if len(sourceParts) != 2 || len(destParts) != 2 {
			logger.Fatal("Invalid repository format. Use [registry]/[repository]", nil, nil)
		}

		sourceRegistry := sourceParts[0]
		sourceRepo := sourceParts[1]
		destRegistry := destParts[0]
		destRepo := destParts[1]

		logger.Info("Starting replication", map[string]interface{}{
			"source":      fmt.Sprintf("%s/%s", sourceRegistry, sourceRepo),
			"destination": fmt.Sprintf("%s/%s", destRegistry, destRepo),
		})

		// Create the registry clients
		registryClients := make(map[string]common.RegistryClient)

		// Create clients based on the source and destination registries
		if sourceRegistry == "ecr" || destRegistry == "ecr" {
			// Create ECR client
			ecrClient, err := ecr.NewClient(ecr.ClientOptions{
				Region:    ecrRegion,
				AccountID: ecrAccountID,
				Logger:    logger,
			})
			if err != nil {
				logger.Fatal("Failed to create ECR client", err, nil)
			}
			registryClients["ecr"] = ecrClient
		}

		if sourceRegistry == "gcr" || destRegistry == "gcr" {
			// Create GCR client
			gcrClient, err := gcr.NewClient(gcr.ClientOptions{
				Project:  gcrProject,
				Location: gcrLocation,
				Logger:   logger,
			})
			if err != nil {
				logger.Fatal("Failed to create GCR client", err, nil)
			}
			registryClients["gcr"] = gcrClient
		}

		// Create worker pool for parallelism
		workerPool := replication.NewWorkerPool(10, logger)
		defer workerPool.Stop()

		// Create the copier
		copier := copy.NewCopier(logger)

		// Create the reconciler
		reconciler := replication.NewReconciler(logger, copier, workerPool)

		// Set up the replication rule
		rule := replication.ReplicationRule{
			SourceRegistry:        sourceRegistry,
			SourceRepository:      sourceRepo,
			DestinationRegistry:   destRegistry,
			DestinationRepository: destRepo,
			TagFilter:             "*", // Replicate all tags
		}

		// Get the source and destination clients
		sourceClient, ok := registryClients[sourceRegistry]
		if !ok {
			logger.Fatal("Unsupported source registry", nil, map[string]interface{}{
				"registry": sourceRegistry,
			})
		}

		destClient, ok := registryClients[destRegistry]
		if !ok {
			logger.Fatal("Unsupported destination registry", nil, map[string]interface{}{
				"registry": destRegistry,
			})
		}

		// Run the replication
		ctx := context.Background()
		err := reconciler.ReconcileRepository(ctx, rule, sourceClient, destClient)
		if err != nil {
			logger.Fatal("Replication failed", err, nil)
		}

		logger.Info("Replication completed successfully", nil)
	},
}

var replicateTreeCmd = &cobra.Command{
	Use:   "replicate-tree [source-registry]/[source-prefix] [destination-registry]/[destination-prefix]",
	Short: "Replicate a repository tree from one registry to another",
	Long:  `Replicates an entire tree of repositories from a source registry to a destination registry.
	
This command allows you to replicate multiple repositories at once based on a prefix.
You can filter which repositories and tags to include or exclude using pattern matching.

Example usage:
  freightliner replicate-tree ecr/prod gcr/prod-mirror
  freightliner replicate-tree ecr/staging gcr/staging-mirror --exclude-repo="internal-*" --include-tag="v*"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger := createLogger(logLevel)

		// Parse source and destination
		sourceParts := strings.SplitN(args[0], "/", 2)
		destParts := strings.SplitN(args[1], "/", 2)

		if len(sourceParts) != 2 || len(destParts) != 2 {
			logger.Fatal("Invalid format. Use [registry]/[prefix]", nil, nil)
		}

		sourceRegistry := sourceParts[0]
		sourcePrefix := sourceParts[1]
		destRegistry := destParts[0]
		destPrefix := destParts[1]

		logger.Info("Starting tree replication", map[string]interface{}{
			"source":      fmt.Sprintf("%s/%s", sourceRegistry, sourcePrefix),
			"destination": fmt.Sprintf("%s/%s", destRegistry, destPrefix),
			"workers":     treeReplicateWorkers,
			"dry_run":     treeReplicateDryRun,
		})

		// Create the registry clients
		registryClients := make(map[string]common.RegistryClient)

		// Create clients based on the source and destination registries
		if sourceRegistry == "ecr" || destRegistry == "ecr" {
			// Create ECR client
			ecrClient, err := ecr.NewClient(ecr.ClientOptions{
				Region:    ecrRegion,
				AccountID: ecrAccountID, // Empty uses the default from AWS credentials
				Logger:    logger,
			})
			if err != nil {
				logger.Fatal("Failed to create ECR client", err, nil)
			}
			registryClients["ecr"] = ecrClient
		}

		if sourceRegistry == "gcr" || destRegistry == "gcr" {
			// Create GCR client
			gcrClient, err := gcr.NewClient(gcr.ClientOptions{
				Project:  gcrProject,
				Location: gcrLocation,
				Logger:   logger,
			})
			if err != nil {
				logger.Fatal("Failed to create GCR client", err, nil)
			}
			registryClients["gcr"] = gcrClient
		}

		// Get the source and destination clients
		sourceClient, ok := registryClients[sourceRegistry]
		if !ok {
			logger.Fatal("Unsupported source registry", nil, map[string]interface{}{
				"registry": sourceRegistry,
			})
		}

		destClient, ok := registryClients[destRegistry]
		if !ok {
			logger.Fatal("Unsupported destination registry", nil, map[string]interface{}{
				"registry": destRegistry,
			})
		}

		// Create the copier
		copier := copy.NewCopier(logger)

		// Create the tree replicator
		replicator := tree.NewTreeReplicator(logger, copier, tree.TreeReplicatorOptions{
			WorkerCount:         treeReplicateWorkers,
			ExcludeRepositories: treeReplicateExcludeRepos,
			ExcludeTags:         treeReplicateExcludeTags,
			IncludeTags:         treeReplicateIncludeTags,
			DryRun:              treeReplicateDryRun,
		})

		// Run the replication
		ctx := context.Background()
		result, err := replicator.ReplicateTree(
			ctx,
			sourceClient,
			destClient,
			sourcePrefix,
			destPrefix,
			treeReplicateForce,
		)

		if err != nil {
			logger.Fatal("Tree replication failed", err, nil)
		}

		// Print summary
		logger.Info("Tree replication completed successfully", map[string]interface{}{
			"repositories":      result.Repositories,
			"images_replicated": result.ImagesReplicated,
			"images_skipped":    result.ImagesSkipped,
			"images_failed":     result.ImagesFailed,
			"duration_sec":      result.Duration.Seconds(),
		})
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the replication server",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger := createLogger(logLevel)

		// Create the worker pool
		workerPool := replication.NewWorkerPool(10, logger)
		defer workerPool.Stop()

		// Create the copier
		copier := copy.NewCopier(logger)

		// Create the scheduler
		scheduler := replication.NewScheduler(logger, workerPool)
		defer scheduler.Stop()

		// Create the reconciler
		reconciler := replication.NewReconciler(logger, copier, workerPool)

		// Create registry clients
		registryClients := make(map[string]common.RegistryClient)

		// Create ECR client
		ecrClient, err := ecr.NewClient(ecr.ClientOptions{
			Region:    ecrRegion,
			AccountID: ecrAccountID,
			Logger:    logger,
		})
		if err != nil {
			logger.Fatal("Failed to create ECR client", err, nil)
		}
		registryClients["ecr"] = ecrClient

		// Create GCR client
		gcrClient, err := gcr.NewClient(gcr.ClientOptions{
			Project:  gcrProject,
			Location: gcrLocation,
			Logger:   logger,
		})
		if err != nil {
			logger.Fatal("Failed to create GCR client", err, nil)
		}
		registryClients["gcr"] = gcrClient

		// Example replication rules
		rules := []replication.ReplicationRule{
			{
				SourceRegistry:        "ecr",
				SourceRepository:      "my-repository",
				DestinationRegistry:   "gcr",
				DestinationRepository: "my-repository",
				TagFilter:             "v*",
				Schedule:              "*/30 * * * *", // Every 30 minutes
			},
		}

		// Add jobs to the scheduler
		for _, rule := range rules {
			scheduler.AddJob(rule)
		}

		// Set up graceful shutdown
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

		logger.Info("Freightliner server started", nil)

		// Wait for termination signal
		<-signals

		logger.Info("Shutting down", nil)
	},
}

func createLogger(level string) *log.Logger {
	var logLevel log.Level

	switch level {
	case "debug":
		logLevel = log.DebugLevel
	case "info":
		logLevel = log.InfoLevel
	case "warn":
		logLevel = log.WarnLevel
	case "error":
		logLevel = log.ErrorLevel
	case "fatal":
		logLevel = log.FatalLevel
	default:
		logLevel = log.InfoLevel
	}

	return log.NewLogger(logLevel)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
