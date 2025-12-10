package cmd

import (
	"fmt"
	"os"

	"freightliner/pkg/service"

	"github.com/spf13/cobra"
)

// newReplicateTreeCmd creates a new replicate-tree command
func newReplicateTreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replicate-tree [source] [destination]",
		Short: "Replicate a tree of repositories",
		Long:  `Replicates multiple repositories from source to destination registry`,
		Example: `  # Copy all repositories under a prefix
  freightliner replicate-tree ecr/my-company gcr.io/my-project

  # Copy with checkpointing for resume capability
  freightliner replicate-tree --enable-checkpoint ecr/prod gcr.io/prod-backup

  # Resume interrupted replication
  freightliner replicate-tree --resume-id abc123 ecr/prod gcr.io/prod-backup

  # Dry run to preview what repositories would be copied
  freightliner replicate-tree --dry-run quay.io/myorg gcr.io/my-project`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(cmd.Context())
			defer cancel()

			// Parse source and destination
			source := args[0]
			destination := args[1]

			// Create tree replication service
			treeReplicationSvc := service.NewTreeReplicationService(cfg, logger)

			// Execute tree replication
			logger.WithFields(map[string]interface{}{
				"source":      source,
				"destination": destination,
				"force":       cfg.TreeReplicate.Force,
				"dry_run":     cfg.TreeReplicate.DryRun,
				"checkpoint":  cfg.TreeReplicate.EnableCheckpoint,
				"resume_id":   cfg.TreeReplicate.ResumeID,
			}).Info("Starting tree replication")

			result, err := treeReplicationSvc.ReplicateTree(ctx, source, destination)
			if err != nil {
				logger.Error("Tree replication failed", err)
				fmt.Printf("Error during tree replication: %s\n", err)
				os.Exit(1)
			}

			// Print results
			fmt.Println("\nTree replication complete")
			fmt.Printf("Repositories found: %d\n", result.RepositoriesFound)
			fmt.Printf("Repositories replicated: %d\n", result.RepositoriesReplicated)
			fmt.Printf("Repositories skipped: %d\n", result.RepositoriesSkipped)
			fmt.Printf("Repositories failed: %d\n", result.RepositoriesFailed)
			fmt.Printf("Total tags copied: %d\n", result.TotalTagsCopied)
			fmt.Printf("Total tags skipped: %d\n", result.TotalTagsSkipped)
			fmt.Printf("Total errors: %d\n", result.TotalErrors)
			fmt.Printf("Total bytes transferred: %d\n", result.TotalBytesTransferred)

			if cfg.TreeReplicate.EnableCheckpoint && result.CheckpointID != "" {
				fmt.Printf("Checkpoint ID: %s\n", result.CheckpointID)
			}
		},
	}

	// Add tree replicate-specific flags
	cfg.AddTreeReplicateFlags(cmd)

	return cmd
}
