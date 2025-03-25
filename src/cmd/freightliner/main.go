package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"src/internal/log"
	"src/pkg/client/common"
	"src/pkg/client/ecr"
	"src/pkg/client/gcr"
	"src/pkg/copy"
	"src/pkg/replication"
	"src/pkg/secrets"
	"src/pkg/security/encryption"
	"src/pkg/security/signing"
	"src/pkg/tree"
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

	// Checkpoint management flags
	checkpointDir string
	checkpointID  string

	// Tree replication flags
	treeReplicateWorkers       int
	treeReplicateExcludeRepos  []string
	treeReplicateExcludeTags   []string
	treeReplicateIncludeTags   []string
	treeReplicateDryRun        bool
	treeReplicateForce         bool
	treeReplicateCheckpoint    bool
	treeReplicateCheckpointDir string
	treeReplicateResumeID      string
	treeReplicateSkipCompleted bool
	treeReplicateRetryFailed   bool
	ecrRegion                  string
	ecrAccountID               string
	gcrProject                 string
	gcrLocation                string

	// Security flags
	signImages             bool
	verifySignatures       bool
	signKeyPath            string
	signKeyID              string
	signatureStorePath     string
	strictVerification     bool
	useEncryption          bool
	useCustomerManagedKeys bool
	awsKmsKeyID            string
	gcpKmsKeyID            string
	envelopeEncryption     bool

	// Secrets Manager flags
	useSecretsManager    bool
	secretsManagerType   string
	awsSecretRegion      string
	gcpSecretProject     string
	gcpCredentialsFile   string
	registryCredsSecret  string
	encryptionKeysSecret string
	signingKeysSecret    string
)

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().StringVar(&ecrRegion, "ecr-region", "us-west-2", "AWS region for ECR")
	rootCmd.PersistentFlags().StringVar(&ecrAccountID, "ecr-account", "", "AWS account ID for ECR (empty uses default from credentials)")
	rootCmd.PersistentFlags().StringVar(&gcrProject, "gcr-project", "", "GCP project for GCR")
	rootCmd.PersistentFlags().StringVar(&gcrLocation, "gcr-location", "us", "GCR location (us, eu, asia)")

	// Add security-related global flags
	rootCmd.PersistentFlags().BoolVar(&signImages, "sign", false, "Enable image signing")
	rootCmd.PersistentFlags().BoolVar(&verifySignatures, "verify", false, "Verify image signatures")
	rootCmd.PersistentFlags().StringVar(&signKeyPath, "sign-key", "", "Path to the signing key file")
	rootCmd.PersistentFlags().StringVar(&signKeyID, "sign-key-id", "", "ID of the signing key")
	rootCmd.PersistentFlags().StringVar(&signatureStorePath, "signature-store", "/tmp/freightliner-signatures", "Path to store image signatures")
	rootCmd.PersistentFlags().BoolVar(&strictVerification, "strict-verify", false, "Fail if signature verification isn't possible")

	// Add encryption-related global flags
	rootCmd.PersistentFlags().BoolVar(&useEncryption, "encrypt", false, "Enable image encryption")
	rootCmd.PersistentFlags().BoolVar(&useCustomerManagedKeys, "customer-key", false, "Use customer-managed encryption keys")
	rootCmd.PersistentFlags().StringVar(&awsKmsKeyID, "aws-kms-key", "", "AWS KMS key ID for encryption")
	rootCmd.PersistentFlags().StringVar(&gcpKmsKeyID, "gcp-kms-key", "", "GCP KMS key ID for encryption")
	rootCmd.PersistentFlags().BoolVar(&envelopeEncryption, "envelope-encryption", true, "Use envelope encryption")

	// Add secrets manager related flags
	rootCmd.PersistentFlags().BoolVar(&useSecretsManager, "use-secrets-manager", false, "Use cloud provider secrets manager for credentials")
	rootCmd.PersistentFlags().StringVar(&secretsManagerType, "secrets-manager-type", "aws", "Type of secrets manager to use (aws, gcp)")
	rootCmd.PersistentFlags().StringVar(&awsSecretRegion, "aws-secret-region", "", "AWS region for Secrets Manager (defaults to --ecr-region if not specified)")
	rootCmd.PersistentFlags().StringVar(&gcpSecretProject, "gcp-secret-project", "", "GCP project for Secret Manager (defaults to --gcr-project if not specified)")
	rootCmd.PersistentFlags().StringVar(&gcpCredentialsFile, "gcp-credentials-file", "", "GCP credentials file path for Secret Manager")
	rootCmd.PersistentFlags().StringVar(&registryCredsSecret, "registry-creds-secret", "freightliner-registry-credentials", "Secret name for registry credentials")
	rootCmd.PersistentFlags().StringVar(&encryptionKeysSecret, "encryption-keys-secret", "freightliner-encryption-keys", "Secret name for encryption keys")
	rootCmd.PersistentFlags().StringVar(&signingKeysSecret, "signing-keys-secret", "freightliner-signing-keys", "Secret name for signing keys")

	// Add checkpoint management flags
	checkpointCmd.PersistentFlags().StringVar(&checkpointDir, "checkpoint-dir", "/tmp/freightliner-checkpoints", "Directory for checkpoint files")
	checkpointCmd.Flags().StringVar(&checkpointID, "id", "", "Checkpoint ID for operations")

	// Add tree replication flags
	replicateTreeCmd.Flags().IntVar(&treeReplicateWorkers, "workers", 5, "Number of concurrent worker threads")
	replicateTreeCmd.Flags().StringSliceVar(&treeReplicateExcludeRepos, "exclude-repo", []string{}, "Repository patterns to exclude (e.g. 'internal-*')")
	replicateTreeCmd.Flags().StringSliceVar(&treeReplicateExcludeTags, "exclude-tag", []string{}, "Tag patterns to exclude (e.g. 'dev-*')")
	replicateTreeCmd.Flags().StringSliceVar(&treeReplicateIncludeTags, "include-tag", []string{}, "Tag patterns to include (e.g. 'v*')")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateDryRun, "dry-run", false, "Perform a dry run without actually copying images")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateForce, "force", false, "Force overwrite of existing images")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateCheckpoint, "checkpoint", false, "Enable checkpointing for interrupted replications")
	replicateTreeCmd.Flags().StringVar(&treeReplicateCheckpointDir, "checkpoint-dir", "/tmp/freightliner-checkpoints", "Directory for storing checkpoint files")
	replicateTreeCmd.Flags().StringVar(&treeReplicateResumeID, "resume", "", "Resume replication from a checkpoint ID")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateSkipCompleted, "skip-completed", true, "Skip completed repositories when resuming")
	replicateTreeCmd.Flags().BoolVar(&treeReplicateRetryFailed, "retry-failed", true, "Retry failed repositories when resuming")

	// Add commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(replicateCmd)
	rootCmd.AddCommand(replicateTreeCmd)
	rootCmd.AddCommand(checkpointCmd)
	rootCmd.AddCommand(serveCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of freightliner",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Freightliner v0.1.0") // TODO: Use build version from goreleaser
	},
}

var replicateCmd = &cobra.Command{
	Use:   "replicate [source-registry]/[source-repository] [destination-registry]/[destination-repository]",
	Short: "Replicate a repository from one registry to another",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger := createLogger(logLevel)

		// Create a context for operations
		ctx := context.Background()

		// Load credentials from secrets manager if enabled
		if useSecretsManager {
			logger.Info("Using secrets manager for credentials", map[string]interface{}{
				"provider": secretsManagerType,
			})

			secretsProvider, err := initializeSecretsManager(ctx, logger)
			if err != nil {
				logger.Fatal("Failed to initialize secrets manager", err, nil)
			}

			// Load and apply registry credentials
			creds, err := loadRegistryCredentials(ctx, secretsProvider)
			if err != nil {
				logger.Fatal("Failed to load registry credentials", err, nil)
			}
			applyRegistryCredentials(creds)

			// Load and apply encryption keys if encryption is enabled
			if useEncryption {
				keys, err := loadEncryptionKeys(ctx, secretsProvider)
				if err != nil {
					logger.Fatal("Failed to load encryption keys", err, nil)
				}
				applyEncryptionKeys(keys)
			}

			// Load and apply signing keys if signing is enabled
			if signImages {
				keys, err := loadSigningKeys(ctx, secretsProvider)
				if err != nil {
					logger.Fatal("Failed to load signing keys", err, nil)
				}
				applySigningKeys(keys)
			}
		}

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
			"signing":     signImages,
			"encryption":  useEncryption,
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

		// Set up security managers if enabled
		if signImages {
			if signKeyPath == "" {
				logger.Fatal("Signing key path is required when signing is enabled", nil, nil)
			}

			// Create signing manager
			signOpts := signing.SignManagerOptions{
				DefaultProvider:    "cosign",
				SignImages:         signImages,
				VerifyImages:       verifySignatures,
				StrictVerification: strictVerification,
				SignatureStorePath: signatureStorePath,
				KeyPath:            signKeyPath,
				KeyID:              signKeyID,
			}

			signManager := signing.NewManager(signOpts)

			// Create and register a Cosign signer
			signer, err := signing.NewCosignSigner(signing.SignOptions{
				KeyPath: signKeyPath,
				KeyID:   signKeyID,
			})
			if err != nil {
				logger.Fatal("Failed to create Cosign signer", err, nil)
			}

			signManager.RegisterSigner("cosign", signer)
			copier.WithSigningManager(signManager)

			logger.Info("Image signing enabled", map[string]interface{}{
				"provider": "cosign",
				"key_id":   signKeyID,
			})
		}

		// Set up encryption if enabled
		if useEncryption {
			ctx := context.Background()

			// Create encryption providers map
			encProviders := make(map[string]encryption.Provider)

			// Create encryption config
			encConfig := encryption.EncryptionConfig{
				EnvelopeEncryption: envelopeEncryption,
				CustomerManagedKey: useCustomerManagedKeys,
				DataKeyLength:      32, // 256-bit keys
			}

			// Check which KMS provider to use based on provided key IDs and destination registry
			if awsKmsKeyID != "" || destRegistry == "ecr" {
				// Configure for AWS KMS
				encConfig.Provider = "aws-kms"
				encConfig.KeyID = awsKmsKeyID
				encConfig.Region = ecrRegion

				// Create AWS KMS provider
				awsKms, err := encryption.NewAWSKMS(ctx, encryption.AWSOpts{
					Region: ecrRegion,
					KeyID:  awsKmsKeyID,
				})
				if err != nil {
					logger.Fatal("Failed to create AWS KMS provider", err, nil)
				}

				encProviders["aws-kms"] = awsKms

				logger.Info("AWS KMS encryption enabled", map[string]interface{}{
					"region": ecrRegion,
					"key_id": awsKmsKeyID,
					"cmk":    useCustomerManagedKeys,
				})
			} else if gcpKmsKeyID != "" || destRegistry == "gcr" {
				// Configure for GCP KMS
				encConfig.Provider = "gcp-kms"
				encConfig.KeyID = gcpKmsKeyID
				encConfig.Region = gcrLocation

				// Create GCP KMS provider
				gcpKms, err := encryption.NewGCPKMS(ctx, encryption.GCPOpts{
					Project:  gcrProject,
					Location: gcrLocation,
					KeyRing:  "freightliner",
					Key:      "image-encryption",
				})
				if err != nil {
					logger.Fatal("Failed to create GCP KMS provider", err, nil)
				}

				encProviders["gcp-kms"] = gcpKms

				logger.Info("GCP KMS encryption enabled", map[string]interface{}{
					"project":  gcrProject,
					"location": gcrLocation,
					"cmk":      useCustomerManagedKeys,
				})
			}

			// Create encryption manager if we have providers
			if len(encProviders) > 0 {
				encManager := encryption.NewManager(encProviders, encConfig)
				copier.WithEncryptionManager(encManager)
			}
		}

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
	Long: `Replicates an entire tree of repositories from a source registry to a destination registry.
	
This command allows you to replicate multiple repositories at once based on a prefix.
You can filter which repositories and tags to include or exclude using pattern matching.

Example usage:
  freightliner replicate-tree ecr/prod gcr/prod-mirror
  freightliner replicate-tree ecr/staging gcr/staging-mirror --exclude-repo="internal-*" --include-tag="v*"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger := createLogger(logLevel)

		// Create a context for operations
		ctx := context.Background()

		// Load credentials from secrets manager if enabled
		if useSecretsManager {
			logger.Info("Using secrets manager for credentials", map[string]interface{}{
				"provider": secretsManagerType,
			})

			secretsProvider, err := initializeSecretsManager(ctx, logger)
			if err != nil {
				logger.Fatal("Failed to initialize secrets manager", err, nil)
			}

			// Load and apply registry credentials
			creds, err := loadRegistryCredentials(ctx, secretsProvider)
			if err != nil {
				logger.Fatal("Failed to load registry credentials", err, nil)
			}
			applyRegistryCredentials(creds)

			// Load and apply encryption keys if encryption is enabled
			if useEncryption {
				keys, err := loadEncryptionKeys(ctx, secretsProvider)
				if err != nil {
					logger.Fatal("Failed to load encryption keys", err, nil)
				}
				applyEncryptionKeys(keys)
			}

			// Load and apply signing keys if signing is enabled
			if signImages {
				keys, err := loadSigningKeys(ctx, secretsProvider)
				if err != nil {
					logger.Fatal("Failed to load signing keys", err, nil)
				}
				applySigningKeys(keys)
			}
		}

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
			EnableCheckpoints:   treeReplicateCheckpoint,
			CheckpointDir:       treeReplicateCheckpointDir,
		})

		// Create context
		ctx := context.Background()

		var result *tree.ReplicationResult
		var err error

		// Check if we're resuming a previous replication
		if treeReplicateResumeID != "" {
			logger.Info("Resuming replication from checkpoint", map[string]interface{}{
				"checkpoint_id": treeReplicateResumeID,
			})

			// Resume the replication
			result, err = replicator.ResumeTreeReplication(
				ctx,
				sourceClient,
				destClient,
				tree.ResumeOptions{
					CheckpointID:   treeReplicateResumeID,
					SkipCompleted:  treeReplicateSkipCompleted,
					RetryFailed:    treeReplicateRetryFailed,
					ForceOverwrite: treeReplicateForce,
				},
			)
		} else {
			// Run a new replication
			result, err = replicator.ReplicateTree(
				ctx,
				sourceClient,
				destClient,
				sourcePrefix,
				destPrefix,
				treeReplicateForce,
			)
		}

		if err != nil {
			logger.Fatal("Tree replication failed", err, nil)
		}

		// Print summary
		var status string
		if result.Interrupted {
			status = "interrupted"
		} else if result.Resumed {
			status = "resumed and completed"
		} else {
			status = "completed successfully"
		}

		summaryInfo := map[string]interface{}{
			"repositories":      result.Repositories,
			"images_replicated": result.ImagesReplicated,
			"images_skipped":    result.ImagesSkipped,
			"images_failed":     result.ImagesFailed,
			"duration_sec":      result.Duration.Seconds(),
			"progress":          fmt.Sprintf("%.1f%%", result.Progress),
		}

		if treeReplicateCheckpoint && result.CheckpointID != "" {
			summaryInfo["checkpoint_id"] = result.CheckpointID
		}

		logger.Info("Tree replication "+status, summaryInfo)
	},
}

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "Manage replication checkpoints",
	Long: `Manage replication checkpoints for interrupted replications.

This command provides subcommands to list, show, and delete replication checkpoints.`,
}

var checkpointListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available checkpoints",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger := createLogger(logLevel)

		// Initialize checkpoint store
		store, err := tree.InitCheckpointStore(checkpointDir)
		if err != nil {
			logger.Fatal("Failed to initialize checkpoint store", err, map[string]interface{}{
				"checkpoint_dir": checkpointDir,
			})
		}

		// Get resumable checkpoints
		checkpoints, err := tree.ListResumableCheckpoints(store)
		if err != nil {
			logger.Fatal("Failed to list checkpoints", err, nil)
		}

		if len(checkpoints) == 0 {
			fmt.Println("No checkpoints found.")
			return
		}

		// Print checkpoints
		fmt.Printf("Found %d checkpoint(s):\n\n", len(checkpoints))
		fmt.Printf("%-36s %-15s %-15s %-10s %-25s\n", "ID", "SOURCE", "PROGRESS", "STATUS", "LAST UPDATED")
		fmt.Println(strings.Repeat("-", 100))

		for _, cp := range checkpoints {
			fmt.Printf("%-36s %-15s %-15s %-10s %-25s\n",
				cp.ID,
				fmt.Sprintf("%s/%s", cp.SourceRegistry, cp.SourcePrefix),
				fmt.Sprintf("%.1f%%", cp.Progress),
				cp.Status,
				cp.LastUpdated.Format(time.RFC3339),
			)
		}
	},
}

var checkpointShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show details of a specific checkpoint",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger := createLogger(logLevel)

		if checkpointID == "" {
			logger.Fatal("Checkpoint ID is required", nil, nil)
		}

		// Initialize checkpoint store
		store, err := tree.InitCheckpointStore(checkpointDir)
		if err != nil {
			logger.Fatal("Failed to initialize checkpoint store", err, map[string]interface{}{
				"checkpoint_dir": checkpointDir,
			})
		}

		// Get checkpoint
		checkpoint, err := store.LoadCheckpoint(checkpointID)
		if err != nil {
			logger.Fatal("Failed to load checkpoint", err, map[string]interface{}{
				"id": checkpointID,
			})
		}

		// Print checkpoint details
		fmt.Println("Checkpoint Details:")
		fmt.Printf("  ID:               %s\n", checkpoint.ID)
		fmt.Printf("  Source Registry:  %s\n", checkpoint.SourceRegistry)
		fmt.Printf("  Source Prefix:    %s\n", checkpoint.SourcePrefix)
		fmt.Printf("  Destination:      %s/%s\n", checkpoint.DestRegistry, checkpoint.DestPrefix)
		fmt.Printf("  Status:           %s\n", checkpoint.Status)
		fmt.Printf("  Progress:         %.1f%%\n", checkpoint.Progress)
		fmt.Printf("  Start Time:       %s\n", checkpoint.StartTime.Format(time.RFC3339))
		fmt.Printf("  Last Updated:     %s\n", checkpoint.LastUpdated.Format(time.RFC3339))
		fmt.Printf("  Duration:         %s\n", checkpoint.LastUpdated.Sub(checkpoint.StartTime))
		fmt.Printf("  Repositories:     %d total, %d completed\n", len(checkpoint.Repositories), len(checkpoint.CompletedRepositories))

		if checkpoint.LastError != "" {
			fmt.Printf("  Last Error:       %s\n", checkpoint.LastError)
		}
	},
}

var checkpointDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a checkpoint",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger := createLogger(logLevel)

		if checkpointID == "" {
			logger.Fatal("Checkpoint ID is required", nil, nil)
		}

		// Initialize checkpoint store
		store, err := tree.InitCheckpointStore(checkpointDir)
		if err != nil {
			logger.Fatal("Failed to initialize checkpoint store", err, map[string]interface{}{
				"checkpoint_dir": checkpointDir,
			})
		}

		// Delete checkpoint
		err = store.DeleteCheckpoint(checkpointID)
		if err != nil {
			logger.Fatal("Failed to delete checkpoint", err, map[string]interface{}{
				"id": checkpointID,
			})
		}

		fmt.Printf("Checkpoint %s deleted successfully.\n", checkpointID)
	},
}

func init() {
	// Add checkpoint subcommands
	checkpointCmd.AddCommand(checkpointListCmd)
	checkpointCmd.AddCommand(checkpointShowCmd)
	checkpointCmd.AddCommand(checkpointDeleteCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the replication server",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger := createLogger(logLevel)

		// Create a context for operations
		ctx := context.Background()

		// Load credentials from secrets manager if enabled
		if useSecretsManager {
			logger.Info("Using secrets manager for credentials", map[string]interface{}{
				"provider": secretsManagerType,
			})

			secretsProvider, err := initializeSecretsManager(ctx, logger)
			if err != nil {
				logger.Fatal("Failed to initialize secrets manager", err, nil)
			}

			// Load and apply registry credentials
			creds, err := loadRegistryCredentials(ctx, secretsProvider)
			if err != nil {
				logger.Fatal("Failed to load registry credentials", err, nil)
			}
			applyRegistryCredentials(creds)

			// Load and apply encryption keys if encryption is enabled
			if useEncryption {
				keys, err := loadEncryptionKeys(ctx, secretsProvider)
				if err != nil {
					logger.Fatal("Failed to load encryption keys", err, nil)
				}
				applyEncryptionKeys(keys)
			}

			// Load and apply signing keys if signing is enabled
			if signImages {
				keys, err := loadSigningKeys(ctx, secretsProvider)
				if err != nil {
					logger.Fatal("Failed to load signing keys", err, nil)
				}
				applySigningKeys(keys)
			}
		}

		// Create the worker pool
		workerPool := replication.NewWorkerPool(10, logger)
		defer workerPool.Stop()

		// Create the copier
		copier := copy.NewCopier(logger)

		// Create the scheduler
		scheduler := replication.NewScheduler(logger, workerPool)
		defer scheduler.Stop()

		// Create the reconciler
		// Create reconciler but ignore for now (will be used in future functionality)
		_ = replication.NewReconciler(logger, copier, workerPool)

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
