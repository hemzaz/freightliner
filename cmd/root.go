package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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

			// Load configuration from file if specified
			if configFile != "" {
				var err error
				cfg, err = config.LoadFromFile(configFile)
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

				// Re-apply command line flags to override config file and env vars
				cmd.Flags().Visit(func(f *pflag.Flag) {
					// Flags present on command line take precedence
				})
			}

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

	// Add commands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newReplicateCmd())
	rootCmd.AddCommand(newReplicateTreeCmd())
	rootCmd.AddCommand(newCheckpointCmd())
	rootCmd.AddCommand(newServeCmd())
}

// setupCommand creates a logger and a cancellable context
func setupCommand(ctx context.Context) (*log.Logger, context.Context, context.CancelFunc) {
	logger := createLogger(cfg.LogLevel)
	ctx, cancel := context.WithCancel(ctx)

	// Set up signal handling
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigCh:
			logger.Info("Received termination signal, shutting down", nil)
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	return logger, ctx, cancel
}

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
