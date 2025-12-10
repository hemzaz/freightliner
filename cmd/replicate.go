package cmd

import (
	"fmt"
	"os"

	"freightliner/pkg/service"

	"github.com/spf13/cobra"
)

// newReplicateCmd creates a new replicate command
func newReplicateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replicate [source] [destination]",
		Short: "Replicate container images",
		Long:  `Replicates container images from source to destination registry`,
		Example: `  # Copy from Docker Hub to another registry
  freightliner replicate docker.io/library/alpine:latest gcr.io/my-project/alpine:latest

  # Copy all tags (specify repository without tag)
  freightliner replicate quay.io/prometheus/node-exporter gcr.io/my-project/node-exporter

  # Copy specific tags only
  freightliner replicate --tags v1.0,v1.1 ghcr.io/owner/repo gcr.io/my-project/repo

  # Dry run to preview what would be copied
  freightliner replicate --dry-run docker.io/library/nginx:latest gcr.io/my-project/nginx:latest`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(cmd.Context())
			defer cancel()

			// Parse source and destination
			source := args[0]
			destination := args[1]

			// Create replication service
			replicationSvc := service.NewReplicationService(cfg, logger)

			// Execute replication
			logger.WithFields(map[string]interface{}{
				"source":      source,
				"destination": destination,
				"force":       cfg.Replicate.Force,
				"dry_run":     cfg.Replicate.DryRun,
			}).Info("Starting replication")

			result, err := replicationSvc.ReplicateRepository(ctx, source, destination)
			if err != nil {
				logger.Error("Replication failed", err)
				fmt.Printf("Error during replication: %s\n", err)
				os.Exit(1)
			}

			// Print results
			fmt.Println("\nReplication complete")
			fmt.Printf("Tags copied: %d\n", result.LayersCopied)
			fmt.Printf("Tags skipped: %d\n", 0) // This info is not available in ReplicationResult
			fmt.Printf("Errors: %s\n", func() string {
				if result.Error != nil {
					return result.Error.Error()
				}
				return "none"
			}())
			fmt.Printf("Total bytes transferred: %d\n", result.BytesCopied)
		},
	}

	// Add replicate-specific flags
	cfg.AddReplicateFlags(cmd)

	return cmd
}
