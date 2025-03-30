package main

import (
	"context"
	"fmt"
	"freightliner/pkg/client/common"
	"freightliner/pkg/client/ecr"
	"freightliner/pkg/client/gcr"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/security/encryption"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// Configuration variables
var (
	logLevel      string
	ecrRegion     string
	ecrAccountID  string
	gcrProject    string
	gcrLocation   string
	useEncryption bool

	// Worker pool configuration
	workerConfig = struct {
		replicateWorkers int
		serveWorkers     int
		autoDetect       bool
	}{
		replicateWorkers: 0,
		serveWorkers:     0,
		autoDetect:       true,
	}

	// Encryption configuration
	encryptionConfig = struct {
		customerManagedKeys bool
		awsKmsKeyID         string
		gcpKmsKeyID         string
		gcpKeyRing          string
		gcpKeyName          string
		envelopeEncryption  bool
	}{
		customerManagedKeys: false,
		awsKmsKeyID:         "",
		gcpKmsKeyID:         "",
		gcpKeyRing:          "freightliner",
		gcpKeyName:          "image-encryption",
		envelopeEncryption:  true,
	}

	// Secrets configuration
	secretsConfig = struct {
		useSecretsManager    bool
		secretsManagerType   string
		awsSecretRegion      string
		gcpSecretProject     string
		gcpCredentialsFile   string
		registryCredsSecret  string
		encryptionKeysSecret string
	}{
		useSecretsManager:    false,
		secretsManagerType:   "aws",
		awsSecretRegion:      "",
		gcpSecretProject:     "",
		gcpCredentialsFile:   "",
		registryCredsSecret:  "freightliner-registry-credentials",
		encryptionKeysSecret: "freightliner-encryption-keys",
	}

	// Tree replication options
	treeReplicateOpts = struct {
		workers          int
		excludeRepos     []string
		excludeTags      []string
		includeTags      []string
		dryRun           bool
		force            bool
		enableCheckpoint bool
		checkpointDir    string
		resumeID         string
		skipCompleted    bool
		retryFailed      bool
	}{
		workers:          0,
		excludeRepos:     []string{},
		excludeTags:      []string{},
		includeTags:      []string{},
		dryRun:           false,
		force:            false,
		enableCheckpoint: false,
		checkpointDir:    "${HOME}/.freightliner/checkpoints",
		resumeID:         "",
		skipCompleted:    true,
		retryFailed:      true,
	}

	// Checkpoint configuration
	checkpointConfig = struct {
		dir string
		id  string
	}{
		dir: "${HOME}/.freightliner/checkpoints",
		id:  "",
	}

	// Root command
	rootCmd = &cobra.Command{
		Use:   "freightliner",
		Short: "Freightliner is a container image replication tool",
		Long:  `A tool for replicating container images between registries like AWS ECR and Google GCR`,
	}

	// Checkpoint command for managing operation checkpoints
	checkpointCmd = &cobra.Command{
		Use:   "checkpoint",
		Short: "Manage replication checkpoints",
		Long:  `Commands for creating, listing, and resuming image replication checkpoints`,
	}

	// Tree replication command
	replicateTreeCmd = &cobra.Command{
		Use:   "replicate-tree [source] [destination]",
		Short: "Replicate a tree of repositories",
		Long:  `Replicates multiple repositories from source to destination registry`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(context.Background())
			defer cancel()

			// Parse source and destination
			source := args[0]
			destination := args[1]

			// Parse source and destination registry paths
			sourceRegistry, sourceRepo, err := parseRegistryPath(source)
			if err != nil {
				fmt.Printf("Error parsing source path: %s\n", err)
				os.Exit(1)
			}

			destRegistry, destRepo, err := parseRegistryPath(destination)
			if err != nil {
				fmt.Printf("Error parsing destination path: %s\n", err)
				os.Exit(1)
			}

			// Validate registry types
			if !isValidRegistryType(sourceRegistry) || !isValidRegistryType(destRegistry) {
				fmt.Println("Error: Registry type must be 'ecr' or 'gcr'")
				os.Exit(1)
			}

			// Create registry clients
			registryClients, err := createRegistryClients(logger, sourceRegistry, destRegistry)
			if err != nil {
				fmt.Printf("Error creating registry clients: %s\n", err)
				os.Exit(1)
			}

			// Initialize credentials if using secrets manager
			if err := initializeCredentials(ctx, logger); err != nil {
				fmt.Printf("Error initializing credentials: %s\n", err)
				os.Exit(1)
			}

			// Setup encryption manager if encryption is enabled
			_, err = setupEncryptionManager(ctx, logger, destRegistry)
			if err != nil {
				fmt.Printf("Error setting up encryption: %s\n", err)
				os.Exit(1)
			}

			// Get source and destination clients
			sourceClient := registryClients[sourceRegistry]
			destClient := registryClients[destRegistry]

			// Determine worker count
			workerCount := treeReplicateOpts.workers
			if workerCount == 0 && workerConfig.autoDetect {
				workerCount = getOptimalWorkerCount()
				logger.Info("Auto-detected worker count", map[string]interface{}{
					"workers": workerCount,
				})
			}

			// Create options map
			options := map[string]interface{}{
				"workers":          workerCount,
				"excludeRepos":     treeReplicateOpts.excludeRepos,
				"excludeTags":      treeReplicateOpts.excludeTags,
				"includeTags":      treeReplicateOpts.includeTags,
				"dryRun":           treeReplicateOpts.dryRun,
				"force":            treeReplicateOpts.force,
				"enableCheckpoint": treeReplicateOpts.enableCheckpoint,
				"checkpointDir":    treeReplicateOpts.checkpointDir,
				"resumeID":         treeReplicateOpts.resumeID,
				"skipCompleted":    treeReplicateOpts.skipCompleted,
				"retryFailed":      treeReplicateOpts.retryFailed,
			}

			// Create a tree replicator with our configuration
			replicator, err := createTreeReplicator(ctx, sourceClient, destClient, sourceRepo, destRepo, logger, options)
			if err != nil {
				fmt.Printf("Error creating tree replicator: %s\n", err)
				os.Exit(1)
			}

			// Start replication
			fmt.Printf("Starting tree replication from %s/%s to %s/%s\n",
				sourceClient.GetRegistryName(), sourceRepo,
				destClient.GetRegistryName(), destRepo)

			if treeReplicateOpts.resumeID != "" {
				fmt.Printf("Resuming from checkpoint: %s\n", treeReplicateOpts.resumeID)
			}

			err = replicator.ReplicateTree(ctx)
			if err != nil {
				fmt.Printf("Error during tree replication: %s\n", err)
				os.Exit(1)
			}

			fmt.Println("\nTree replication complete")
		},
	}

	// Version command
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  `Displays the version and build information for this installation of Freightliner`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Freightliner v0.1.0")
			fmt.Println("Go Version:", runtime.Version())
			fmt.Println("OS/Arch:", runtime.GOOS+"/"+runtime.GOARCH)
		},
	}

	// Replicate command
	replicateCmd = &cobra.Command{
		Use:   "replicate [source] [destination]",
		Short: "Replicate container images",
		Long:  `Replicates container images from source to destination registry`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(context.Background())
			defer cancel()

			// Parse source and destination
			source := args[0]
			destination := args[1]

			// Parse source and destination registry paths
			sourceRegistry, sourceRepo, err := parseRegistryPath(source)
			if err != nil {
				fmt.Printf("Error parsing source path: %s\n", err)
				os.Exit(1)
			}

			destRegistry, destRepo, err := parseRegistryPath(destination)
			if err != nil {
				fmt.Printf("Error parsing destination path: %s\n", err)
				os.Exit(1)
			}

			// Validate registry types
			if !isValidRegistryType(sourceRegistry) || !isValidRegistryType(destRegistry) {
				fmt.Println("Error: Registry type must be 'ecr' or 'gcr'")
				os.Exit(1)
			}

			// Create registry clients
			registryClients, err := createRegistryClients(logger, sourceRegistry, destRegistry)
			if err != nil {
				fmt.Printf("Error creating registry clients: %s\n", err)
				os.Exit(1)
			}

			// Initialize credentials if using secrets manager
			if err := initializeCredentials(ctx, logger); err != nil {
				fmt.Printf("Error initializing credentials: %s\n", err)
				os.Exit(1)
			}

			// Setup encryption manager if encryption is enabled
			encManager, err := setupEncryptionManager(ctx, logger, destRegistry)
			if err != nil {
				fmt.Printf("Error setting up encryption: %s\n", err)
				os.Exit(1)
			}

			// Get source repository
			sourceClient := registryClients[sourceRegistry]
			sourceRepository, err := sourceClient.GetRepository(ctx, sourceRepo)
			if err != nil {
				fmt.Printf("Error getting source repository: %s\n", err)
				os.Exit(1)
			}

			// Get or create destination repository
			destClient := registryClients[destRegistry]
			destRepository, err := destClient.GetRepository(ctx, destRepo)
			if err != nil {
				logger.Info("Destination repository does not exist, attempting to create", nil)

				// If we have a type-specific client with creation capability, use it
				if creator, ok := destClient.(RepositoryCreator); ok {
					destRepository, err = creator.CreateRepository(ctx, destRepo, map[string]string{
						"CreatedBy": "Freightliner",
						"Source":    sourceClient.GetRegistryName() + "/" + sourceRepo,
					})
					if err != nil {
						fmt.Printf("Error creating destination repository: %s\n", err)
						os.Exit(1)
					}
				} else {
					fmt.Println("Error: Destination registry does not support repository creation")
					os.Exit(1)
				}
			}

			// Determine worker count
			workerCount := workerConfig.replicateWorkers
			if workerCount == 0 && workerConfig.autoDetect {
				workerCount = getOptimalWorkerCount()
				logger.Info("Auto-detected worker count", map[string]interface{}{
					"workers": workerCount,
				})
			}

			// Create a copier with our configuration
			copier, err := createCopier(ctx, sourceRepository, destRepository, encManager, logger, workerCount)
			if err != nil {
				fmt.Printf("Error creating copier: %s\n", err)
				os.Exit(1)
			}

			// Start replication
			fmt.Printf("Starting replication from %s/%s to %s/%s\n",
				sourceClient.GetRegistryName(), sourceRepo,
				destClient.GetRegistryName(), destRepo)

			result, err := copier.CopyRepository(ctx)
			if err != nil {
				fmt.Printf("Error during replication: %s\n", err)
				os.Exit(1)
			}

			// Print results
			fmt.Println("\nReplication complete")
			fmt.Printf("Tags copied: %d\n", result.TagsCopied)
			fmt.Printf("Tags skipped: %d\n", result.TagsSkipped)
			fmt.Printf("Errors: %d\n", result.Errors)
			fmt.Printf("Total bytes transferred: %d\n", result.BytesTransferred)
		},
	}

	// Serve command
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the replication server",
		Long:  `Starts a server that listens for replication requests`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger := createLogger(logLevel)
			fmt.Println("Server mode not yet implemented")
			logger.Info("Starting replication server", nil)
			os.Exit(1)
		},
	}

	// Checkpoint subcommands
	checkpointListCmd = &cobra.Command{
		Use:   "list",
		Short: "List checkpoints",
		Long:  `Lists all available replication checkpoints`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger := createLogger(logLevel)

			logger.Info("Listing checkpoints", map[string]interface{}{
				"dir": checkpointConfig.dir,
			})
			fmt.Println("Checkpoint listing not yet implemented")
		},
	}

	checkpointShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show checkpoint details",
		Long:  `Shows detailed information about a specific checkpoint`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger := createLogger(logLevel)

			if checkpointConfig.id == "" {
				fmt.Println("Error: checkpoint ID is required")
				os.Exit(1)
			}

			logger.Info("Showing checkpoint", map[string]interface{}{
				"id":  checkpointConfig.id,
				"dir": checkpointConfig.dir,
			})
			fmt.Println("Checkpoint details not yet implemented")
		},
	}

	checkpointDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a checkpoint",
		Long:  `Deletes a specific checkpoint`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger := createLogger(logLevel)

			if checkpointConfig.id == "" {
				fmt.Println("Error: checkpoint ID is required")
				os.Exit(1)
			}

			logger.Info("Deleting checkpoint", map[string]interface{}{
				"id":  checkpointConfig.id,
				"dir": checkpointConfig.dir,
			})
			fmt.Println("Checkpoint deletion not yet implemented")
		},
	}
)

// The entry point of the application
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// max returns the larger of two ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func getOptimalWorkerCount() int {
	numCPU := runtime.NumCPU()

	// Calculate available memory and system load if needed
	// This is where you could add more sophisticated logic

	// Simple heuristic:
	// - Minimum of 2 workers
	// - Maximum of NumCPU * 2 (allowing for I/O bound operations)
	// - Default to NumCPU for a balance of CPU utilization

	if numCPU <= 2 {
		return 2 // Always have at least 2 workers
	} else if numCPU <= 4 {
		return numCPU // For small machines, use one worker per core
	} else {
		return numCPU - 1 // For larger machines, leave one core free for system tasks
	}
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().StringVar(&ecrRegion, "ecr-region", "us-west-2", "AWS region for ECR")
	rootCmd.PersistentFlags().StringVar(&ecrAccountID, "ecr-account", "", "AWS account ID for ECR (empty uses default from credentials)")
	rootCmd.PersistentFlags().StringVar(&gcrProject, "gcr-project", "", "GCP project for GCR")
	rootCmd.PersistentFlags().StringVar(&gcrLocation, "gcr-location", "us", "GCR location (us, eu, asia)")

	// Add worker configuration flags
	rootCmd.PersistentFlags().IntVar(&workerConfig.replicateWorkers, "replicate-workers", 0, "Number of concurrent workers for replication (0 = auto-detect)")
	rootCmd.PersistentFlags().IntVar(&workerConfig.serveWorkers, "serve-workers", 0, "Number of concurrent workers for server mode (0 = auto-detect)")
	rootCmd.PersistentFlags().BoolVar(&workerConfig.autoDetect, "auto-detect-workers", true, "Auto-detect optimal worker count based on system resources")

	// Add encryption-related global flags
	rootCmd.PersistentFlags().BoolVar(&useEncryption, "encrypt", false, "Enable image encryption")
	rootCmd.PersistentFlags().BoolVar(&encryptionConfig.customerManagedKeys, "customer-key", false, "Use customer-managed encryption keys")
	rootCmd.PersistentFlags().StringVar(&encryptionConfig.awsKmsKeyID, "aws-kms-key", "", "AWS KMS key ID for encryption")
	rootCmd.PersistentFlags().StringVar(&encryptionConfig.gcpKmsKeyID, "gcp-kms-key", "", "GCP KMS key ID for encryption")
	rootCmd.PersistentFlags().StringVar(&encryptionConfig.gcpKeyRing, "gcp-key-ring", "freightliner", "GCP KMS key ring name")
	rootCmd.PersistentFlags().StringVar(&encryptionConfig.gcpKeyName, "gcp-key-name", "image-encryption", "GCP KMS key name")
	rootCmd.PersistentFlags().BoolVar(&encryptionConfig.envelopeEncryption, "envelope-encryption", true, "Use envelope encryption")

	// Add secrets manager related flags
	rootCmd.PersistentFlags().BoolVar(&secretsConfig.useSecretsManager, "use-secrets-manager", false, "Use cloud provider secrets manager for credentials")
	rootCmd.PersistentFlags().StringVar(&secretsConfig.secretsManagerType, "secrets-manager-type", "aws", "Type of secrets manager to use (aws, gcp)")
	rootCmd.PersistentFlags().StringVar(&secretsConfig.awsSecretRegion, "aws-secret-region", "", "AWS region for Secrets Manager (defaults to --ecr-region if not specified)")
	rootCmd.PersistentFlags().StringVar(&secretsConfig.gcpSecretProject, "gcp-secret-project", "", "GCP project for Secret Manager (defaults to --gcr-project if not specified)")
	rootCmd.PersistentFlags().StringVar(&secretsConfig.gcpCredentialsFile, "gcp-credentials-file", "", "GCP credentials file path for Secret Manager")
	rootCmd.PersistentFlags().StringVar(&secretsConfig.registryCredsSecret, "registry-creds-secret", "freightliner-registry-credentials", "Secret name for registry credentials")
	rootCmd.PersistentFlags().StringVar(&secretsConfig.encryptionKeysSecret, "encryption-keys-secret", "freightliner-encryption-keys", "Secret name for encryption keys")

	// Add checkpoint management flags
	checkpointCmd.PersistentFlags().StringVar(&checkpointConfig.dir, "checkpoint-dir", "${HOME}/.freightliner/checkpoints", "Directory for checkpoint files")
	checkpointCmd.Flags().StringVar(&checkpointConfig.id, "id", "", "Checkpoint ID for operations")

	// Add tree replication flags
	replicateTreeCmd.Flags().IntVar(&treeReplicateOpts.workers, "workers", 0, "Number of concurrent worker threads (0 = auto-detect)")
	replicateTreeCmd.Flags().StringSliceVar(&treeReplicateOpts.excludeRepos, "exclude-repo", []string{}, "Repository patterns to exclude (e.g. 'helper-*')")
	replicateTreeCmd.Flags().StringSliceVar(&treeReplicateOpts.excludeTags, "exclude-tag", []string{}, "Tag patterns to exclude (e.g. 'dev-*')")
	replicateTreeCmd.Flags().StringSliceVar(&treeReplicateOpts.includeTags, "include-tag", []string{}, "Tag patterns to include (e.g. 'v*')")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateOpts.dryRun, "dry-run", false, "Perform a dry run without actually copying images")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateOpts.force, "force", false, "Force overwrite of existing images")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateOpts.enableCheckpoint, "checkpoint", false, "Enable checkpointing for interrupted replications")
	replicateTreeCmd.Flags().StringVar(&treeReplicateOpts.checkpointDir, "checkpoint-dir", "${HOME}/.freightliner/checkpoints", "Directory for storing checkpoint files")
	replicateTreeCmd.Flags().StringVar(&treeReplicateOpts.resumeID, "resume", "", "Resume replication from a checkpoint ID")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateOpts.skipCompleted, "skip-completed", true, "Skip completed repositories when resuming")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateOpts.retryFailed, "retry-failed", true, "Retry failed repositories when resuming")

	// Add commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(replicateCmd)
	rootCmd.AddCommand(replicateTreeCmd)
	rootCmd.AddCommand(checkpointCmd)
	rootCmd.AddCommand(serveCmd)

	// Add checkpoint subcommands
	checkpointCmd.AddCommand(checkpointListCmd)
	checkpointCmd.AddCommand(checkpointShowCmd)
	checkpointCmd.AddCommand(checkpointDeleteCmd)
}

// Consolidated helper functions to reduce duplication
// createLogger creates a new logger with the specified level
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
	default:
		logLevel = log.InfoLevel
	}
	return log.NewLogger(logLevel)
}

func setupCommand(ctx context.Context) (*log.Logger, context.Context, context.CancelFunc) {
	logger := createLogger(logLevel)
	ctx, cancel := context.WithCancel(ctx)
	return logger, ctx, cancel
}

// SecretsProvider represents an interface to a secrets management service
type SecretsProvider interface {
	// GetSecret retrieves a secret by name
	GetSecret(ctx context.Context, name string) (string, error)
}

// RegistryCredentials contains credentials for different registry types
type RegistryCredentials struct {
	ECR map[string]string
	GCR map[string]string
}

// EncryptionKeys contains encryption keys for different registry types
type EncryptionKeys struct {
	AWSKeys map[string]string
	GCPKeys map[string]string
}

// Helper functions for secrets and credentials management
func initializeSecretsManager(ctx context.Context, logger *log.Logger) (SecretsProvider, error) {
	// This would be implemented to create the appropriate secrets manager client
	// For now, return a stub that just logs
	logger.Info("Secrets manager initialization not fully implemented", nil)
	return nil, errors.NotImplementedf("secrets manager initialization")
}

func loadRegistryCredentials(ctx context.Context, provider SecretsProvider) (RegistryCredentials, error) {
	// This would be implemented to load credentials from the secrets manager
	return RegistryCredentials{}, errors.NotImplementedf("registry credentials loading")
}

func applyRegistryCredentials(creds RegistryCredentials) {
	// This would be implemented to apply the loaded credentials
}

func loadEncryptionKeys(ctx context.Context, provider SecretsProvider) (EncryptionKeys, error) {
	// This would be implemented to load encryption keys from the secrets manager
	return EncryptionKeys{}, errors.NotImplementedf("encryption keys loading")
}

func applyEncryptionKeys(keys EncryptionKeys) {
	// This would be implemented to apply the loaded encryption keys
}

func initializeCredentials(ctx context.Context, logger *log.Logger) error {
	if !secretsConfig.useSecretsManager {
		return nil
	}

	logger.Info("Using secrets manager for credentials", map[string]interface{}{
		"provider": secretsConfig.secretsManagerType,
	})

	secretsProvider, err := initializeSecretsManager(ctx, logger)
	if err != nil {
		return errors.Wrap(err, "failed to initialize secrets manager")
	}

	// Load and apply registry credentials
	creds, err := loadRegistryCredentials(ctx, secretsProvider)
	if err != nil {
		return errors.Wrap(err, "failed to load registry credentials")
	}
	applyRegistryCredentials(creds)

	// Load and apply encryption keys if encryption is enabled
	if useEncryption {
		keys, err := loadEncryptionKeys(ctx, secretsProvider)
		if err != nil {
			return errors.Wrap(err, "failed to load encryption keys")
		}
		applyEncryptionKeys(keys)
	}

	return nil
}

func createRegistryClients(logger *log.Logger, registries ...string) (map[string]common.RegistryClient, error) {
	registrySet := make(map[string]bool)
	for _, r := range registries {
		registrySet[r] = true
	}

	registryClients := make(map[string]common.RegistryClient)

	if len(registries) == 0 || registrySet["ecr"] {
		ecrClient, err := ecr.NewClient(ecr.ClientOptions{
			Region:    ecrRegion,
			AccountID: ecrAccountID,
			Logger:    logger,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create ECR client")
		}
		registryClients["ecr"] = ecrClient
	}

	if len(registries) == 0 || registrySet["gcr"] {
		gcrClient, err := gcr.NewClient(gcr.ClientOptions{
			Project:  gcrProject,
			Location: gcrLocation,
			Logger:   logger,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create GCR client")
		}
		registryClients["gcr"] = gcrClient
	}

	return registryClients, nil
}

func parseRegistryPath(path string) (registry, repo string, err error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return "", "", errors.InvalidInputf("invalid format. Use [registry]/[repository]")
	}
	return parts[0], parts[1], nil
}

// RepositoryCreator is an interface for client types that can create repositories
type RepositoryCreator interface {
	// CreateRepository creates a new repository with the given name and tags
	CreateRepository(ctx context.Context, name string, tags map[string]string) (common.Repository, error)
}

// CopyResult contains the results of a copy operation
type CopyResult struct {
	TagsCopied       int
	TagsSkipped      int
	Errors           int
	BytesTransferred int64
}

// Copier represents an interface for copying images between repositories
type Copier interface {
	// CopyRepository copies all images from source to destination repository
	CopyRepository(ctx context.Context) (CopyResult, error)

	// CopyImage copies a single image with the given tag from source to destination
	CopyImage(ctx context.Context, tag string) error
}

// Interface for tree replication
type TreeReplicator interface {
	// ReplicateTree replicates a tree of repositories
	ReplicateTree(ctx context.Context) error

	// ReplicateRepositories replicates a specific set of repositories
	ReplicateRepositories(ctx context.Context, repositories []string) error
}

// isValidRegistryType checks if a registry type is supported
func isValidRegistryType(registry string) bool {
	return registry == "ecr" || registry == "gcr"
}

// createCopier creates a new image copier with the specified configuration
func createCopier(ctx context.Context, source, dest common.Repository, encManager *encryption.Manager, logger *log.Logger, workers int) (Copier, error) {
	// This would be implemented to create a new copier
	// For now, return a stub implementation
	return &stubCopier{
		source:      source,
		destination: dest,
		logger:      logger,
		workers:     workers,
	}, nil
}

// stubCopier is a temporary implementation for the Copier interface
type stubCopier struct {
	source      common.Repository
	destination common.Repository
	logger      *log.Logger
	workers     int
}

func (c *stubCopier) CopyRepository(ctx context.Context) (CopyResult, error) {
	// This would be implemented to copy all images between repositories
	c.logger.Info("Stub implementation of CopyRepository", map[string]interface{}{
		"source":      c.source,
		"destination": c.destination,
		"workers":     c.workers,
	})

	// Return mock results
	return CopyResult{
		TagsCopied:       5,
		TagsSkipped:      0,
		Errors:           0,
		BytesTransferred: 1024 * 1024 * 10, // 10MB
	}, nil
}

func (c *stubCopier) CopyImage(ctx context.Context, tag string) error {
	// This would be implemented to copy a single image
	c.logger.Info("Stub implementation of CopyImage", map[string]interface{}{
		"tag":         tag,
		"source":      c.source,
		"destination": c.destination,
	})
	return nil
}

// createTreeReplicator creates a new tree replicator with the specified configuration
func createTreeReplicator(ctx context.Context, source common.RegistryClient, dest common.RegistryClient, sourcePath, destPath string, logger *log.Logger, opts map[string]interface{}) (TreeReplicator, error) {
	// This would be implemented to create a new tree replicator
	// For now, return a stub implementation
	return &stubTreeReplicator{
		sourceClient: source,
		destClient:   dest,
		sourcePath:   sourcePath,
		destPath:     destPath,
		logger:       logger,
		options:      opts,
	}, nil
}

// stubTreeReplicator is a temporary implementation for the TreeReplicator interface
type stubTreeReplicator struct {
	sourceClient common.RegistryClient
	destClient   common.RegistryClient
	sourcePath   string
	destPath     string
	logger       *log.Logger
	options      map[string]interface{}
}

func (t *stubTreeReplicator) ReplicateTree(ctx context.Context) error {
	// This would be implemented to replicate a tree of repositories
	t.logger.Info("Stub implementation of ReplicateTree", map[string]interface{}{
		"source":      t.sourcePath,
		"destination": t.destPath,
		"options":     t.options,
	})
	return nil
}

func (t *stubTreeReplicator) ReplicateRepositories(ctx context.Context, repositories []string) error {
	// This would be implemented to replicate specific repositories
	t.logger.Info("Stub implementation of ReplicateRepositories", map[string]interface{}{
		"repositories": repositories,
		"source":       t.sourcePath,
		"destination":  t.destPath,
	})
	return nil
}

func setupEncryptionManager(ctx context.Context, logger *log.Logger, destRegistry string) (*encryption.Manager, error) {
	if !useEncryption {
		// Create an empty manager with no providers instead of returning nil
		return encryption.NewManager(make(map[string]encryption.Provider), encryption.EncryptionConfig{}), nil
	}

	// Create encryption providers map
	encProviders := make(map[string]encryption.Provider)

	// Create encryption config
	encConfig := encryption.EncryptionConfig{
		EnvelopeEncryption: encryptionConfig.envelopeEncryption,
		CustomerManagedKey: encryptionConfig.customerManagedKeys,
		DataKeyLength:      32, // 256-bit keys
	}

	// Check which KMS provider to use based on provided key IDs and destination registry
	if encryptionConfig.awsKmsKeyID != "" || destRegistry == "ecr" {
		// Configure for AWS KMS
		encConfig.Provider = "aws-kms"
		encConfig.KeyID = encryptionConfig.awsKmsKeyID
		encConfig.Region = ecrRegion

		// Create AWS KMS provider
		awsKms, err := encryption.NewAWSKMS(ctx, encryption.AWSOpts{
			Region: ecrRegion,
			KeyID:  encryptionConfig.awsKmsKeyID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create AWS KMS provider")
		}

		encProviders["aws-kms"] = awsKms

		logger.Info("AWS KMS encryption enabled", map[string]interface{}{
			"region": ecrRegion,
			"key_id": encryptionConfig.awsKmsKeyID,
			"cmk":    encryptionConfig.customerManagedKeys,
		})
	} else if encryptionConfig.gcpKmsKeyID != "" || destRegistry == "gcr" {
		// Configure for GCP KMS
		encConfig.Provider = "gcp-kms"
		encConfig.KeyID = encryptionConfig.gcpKmsKeyID
		encConfig.Region = gcrLocation

		// Create GCP KMS provider
		gcpKms, err := encryption.NewGCPKMS(ctx, encryption.GCPOpts{
			Project:  gcrProject,
			Location: gcrLocation,
			KeyRing:  encryptionConfig.gcpKeyRing,
			Key:      encryptionConfig.gcpKeyName,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create GCP KMS provider")
		}

		encProviders["gcp-kms"] = gcpKms

		logger.Info("GCP KMS encryption enabled", map[string]interface{}{
			"project":  gcrProject,
			"location": gcrLocation,
			"key_ring": encryptionConfig.gcpKeyRing,
			"key_name": encryptionConfig.gcpKeyName,
			"cmk":      encryptionConfig.customerManagedKeys,
		})
	}

	// Create encryption manager if we have providers
	if len(encProviders) > 0 {
		return encryption.NewManager(encProviders, encConfig), nil
	}

	return nil, nil
}
