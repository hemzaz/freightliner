package cmd

import (
	"fmt"
	"os"

	"freightliner/pkg/config"
	"freightliner/pkg/server"
	"freightliner/pkg/service"

	"github.com/spf13/cobra"
)

// newServeCmd creates a new serve command
func newServeCmd() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the replication server",
		Long:  `Starts a server that listens for replication requests`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(cmd.Context())
			defer cancel()

			// Load configuration from file if specified
			if configFile != "" {
				logger.Info("Loading configuration from file", map[string]interface{}{
					"file": configFile,
				})

				loadedCfg, err := config.LoadFromFile(configFile)
				if err != nil {
					logger.Error("Failed to load configuration", err, nil)
					fmt.Printf("Error loading configuration: %s\n", err)
					os.Exit(1)
				}

				// Replace our global configuration
				cfg = loadedCfg
			}

			logger.Info("Starting replication server", map[string]interface{}{
				"port":    cfg.Server.Port,
				"workers": cfg.Workers.ServeWorkers,
			})

			// Create services
			replicationSvc := service.NewReplicationService(cfg, logger)
			treeReplicationSvc := service.NewTreeReplicationService(cfg, logger)
			checkpointSvc := service.NewCheckpointService(cfg, logger)

			// Create server
			srv, err := server.NewServer(ctx, cfg, logger, replicationSvc, treeReplicationSvc, checkpointSvc)
			if err != nil {
				logger.Error("Failed to create server", err, nil)
				fmt.Printf("Error creating server: %s\n", err)
				os.Exit(1)
			}

			// Start server (this will block until the server is shut down)
			if err := srv.Start(); err != nil {
				logger.Error("Server failed", err, nil)
				fmt.Printf("Server error: %s\n", err)
				os.Exit(1)
			}
		},
	}

	// Add server-specific flags
	cfg.AddServerFlags(cmd)

	// Add config file flag
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path")

	return cmd
}
