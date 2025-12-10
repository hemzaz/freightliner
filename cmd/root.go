package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// Configuration
	cfg *config.Config

	// Path to configuration file
	configFile string

	// Root command
	rootCmd = &cobra.Command{
		Use:   "freightliner",
		Short: "Freightliner is a container image replication tool",
		Long:  `A tool for replicating container images between registries like AWS ECR and Google GCR`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip for version and help commands
			if cmd.Name() == "version" || cmd.Name() == "help" {
				return nil
			}

			// Load configuration from file if specified, otherwise use defaults
			if configFile != "" {
				var err error
				cfg, err = config.LoadFromFile(configFile)
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}
				// Log successful config load
				fmt.Printf("✅ Loaded configuration from: %s\n", configFile)
			} else if cfg == nil {
				// Initialize with default config if no config file and cfg not set
				cfg = config.NewDefaultConfig()
			}

			// Apply command line flags to override config file values or defaults
			cmd.Flags().Visit(func(f *pflag.Flag) {
				// Flags present on command line take precedence over config file and defaults
				switch f.Name {
				case "log-level":
					cfg.LogLevel = f.Value.String()
				case "ecr-region":
					cfg.ECR.Region = f.Value.String()
				case "ecr-account":
					cfg.ECR.AccountID = f.Value.String()
				case "gcr-project":
					cfg.GCR.Project = f.Value.String()
				case "gcr-location":
					cfg.GCR.Location = f.Value.String()
				case "replicate-workers":
					if val, err := strconv.Atoi(f.Value.String()); err == nil {
						cfg.Workers.ReplicateWorkers = val
					}
				case "serve-workers":
					if val, err := strconv.Atoi(f.Value.String()); err == nil {
						cfg.Workers.ServeWorkers = val
					}
				case "auto-detect-workers":
					if val, err := strconv.ParseBool(f.Value.String()); err == nil {
						cfg.Workers.AutoDetect = val
					}
				case "encrypt":
					if val, err := strconv.ParseBool(f.Value.String()); err == nil {
						cfg.Encryption.Enabled = val
					}
				case "customer-key":
					if val, err := strconv.ParseBool(f.Value.String()); err == nil {
						cfg.Encryption.CustomerManagedKeys = val
					}
				case "aws-kms-key":
					cfg.Encryption.AWSKMSKeyID = f.Value.String()
				case "gcp-kms-key":
					cfg.Encryption.GCPKMSKeyID = f.Value.String()
				case "gcp-key-ring":
					cfg.Encryption.GCPKeyRing = f.Value.String()
				case "gcp-key-name":
					cfg.Encryption.GCPKeyName = f.Value.String()
				case "envelope-encryption":
					if val, err := strconv.ParseBool(f.Value.String()); err == nil {
						cfg.Encryption.EnvelopeEncryption = val
					}
				case "use-secrets-manager":
					if val, err := strconv.ParseBool(f.Value.String()); err == nil {
						cfg.Secrets.UseSecretsManager = val
					}
				case "secrets-manager-type":
					cfg.Secrets.SecretsManagerType = f.Value.String()
				case "aws-secret-region":
					cfg.Secrets.AWSSecretRegion = f.Value.String()
				case "gcp-secret-project":
					cfg.Secrets.GCPSecretProject = f.Value.String()
				case "gcp-credentials-file":
					cfg.Secrets.GCPCredentialsFile = f.Value.String()
				case "registry-creds-secret":
					cfg.Secrets.RegistryCredsSecret = f.Value.String()
				case "encryption-keys-secret":
					cfg.Secrets.EncryptionKeysSecret = f.Value.String()
				case "checkpoint-dir":
					cfg.Checkpoint.Directory = f.Value.String()
				case "force":
					if val, err := strconv.ParseBool(f.Value.String()); err == nil {
						cfg.Replicate.Force = val
					}
				case "dry-run":
					if val, err := strconv.ParseBool(f.Value.String()); err == nil {
						cfg.Replicate.DryRun = val
					}
				case "tags":
					// Cobra's StringSliceVar has already parsed the comma-separated values
					// Just get the parsed slice from the flag (CLI overrides config)
					if tags, err := cmd.Flags().GetStringSlice("tags"); err == nil {
						cfg.Replicate.Tags = tags
					}
				}
			})

			return nil
		},
	}
)

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// init initializes the command structure
func init() {
	// Initialize configuration
	cfg = config.NewDefaultConfig()

	// Add configuration file flag
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to configuration file")

	// Add configuration flags to root command
	cfg.AddFlagsToCommand(rootCmd)

	// Add existing commands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newHealthCheckCmd())
	rootCmd.AddCommand(newReplicateCmd())
	rootCmd.AddCommand(newReplicateTreeCmd())
	rootCmd.AddCommand(newCheckpointCmd())
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newSBOMCmd())
	rootCmd.AddCommand(newScanCmd())

	// Add new advanced CLI commands (Skopeo-like functionality)
	rootCmd.AddCommand(newInspectCmd())
	rootCmd.AddCommand(newListTagsCmd())
	rootCmd.AddCommand(newDeleteCmd())
	rootCmd.AddCommand(newSyncCmd())

	// Add manifest operations
	rootCmd.AddCommand(newManifestCmd())

	// Add layers command
	rootCmd.AddCommand(newLayersCmd())

	// Add auth management
	rootCmd.AddCommand(newAuthCmd())
}

// setupCommand creates a logger and a cancellable context
func setupCommand(ctx context.Context) (log.Logger, context.Context, context.CancelFunc) {
	logger := createLogger(cfg.LogLevel)
	ctx, cancel := context.WithCancel(ctx)

	// Set up signal handling
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigCh:
			logger.Info("Received termination signal, shutting down")
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	return logger, ctx, cancel
}

// createLogger creates a new logger with the specified level
func createLogger(level string) log.Logger {
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
	return log.NewBasicLogger(logLevel)
}
