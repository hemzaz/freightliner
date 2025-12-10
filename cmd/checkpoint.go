package cmd

import (
	"fmt"
	"os"

	"freightliner/pkg/service"

	"github.com/spf13/cobra"
)

// newCheckpointCmd creates a new checkpoint command
func newCheckpointCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkpoint",
		Short: "Manage replication checkpoints",
		Long:  `Commands for creating, listing, and resuming image replication checkpoints`,
	}

	// Add checkpoint-specific flags
	cfg.AddCheckpointFlagsToCommand(cmd)

	// Add checkpoint subcommands
	cmd.AddCommand(newCheckpointListCmd())
	cmd.AddCommand(newCheckpointShowCmd())
	cmd.AddCommand(newCheckpointDeleteCmd())
	cmd.AddCommand(newCheckpointExportCmd())
	cmd.AddCommand(newCheckpointImportCmd())

	return cmd
}

// newCheckpointListCmd creates a new checkpoint list command
func newCheckpointListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List checkpoints",
		Long:  `Lists all available replication checkpoints`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(cmd.Context())
			defer cancel()

			logger.WithFields(map[string]interface{}{
				"dir": cfg.Checkpoint.Directory,
			}).Info("Listing checkpoints")

			// Create checkpoint service
			checkpointSvc := service.NewCheckpointService(cfg, logger)

			// List checkpoints
			checkpoints, err := checkpointSvc.ListCheckpoints(ctx)
			if err != nil {
				logger.Error("Failed to list checkpoints", err)
				fmt.Printf("Error listing checkpoints: %s\n", err)
				os.Exit(1)
			}

			// Print checkpoints
			if len(checkpoints) == 0 {
				fmt.Println("No checkpoints found")
				return
			}

			fmt.Printf("Found %d checkpoints:\n", len(checkpoints))
			fmt.Println("\nID                                   | Created               | Source -> Destination                | Repos | Status")
			fmt.Println("--------------------------------------|----------------------|---------------------------------------|-------|-------")
			for _, cp := range checkpoints {
				fmt.Printf("%-36s | %-20s | %-37s | %5d | %s\n",
					cp.ID,
					cp.CreatedAt.Format("2006-01-02 15:04:05"),
					cp.Source+" -> "+cp.Destination,
					cp.TotalRepositories,
					cp.Status)
			}
		},
	}
}

// newCheckpointShowCmd creates a new checkpoint show command
func newCheckpointShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show checkpoint details",
		Long:  `Shows detailed information about a specific checkpoint`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(cmd.Context())
			defer cancel()

			if cfg.Checkpoint.ID == "" {
				fmt.Println("Error: checkpoint ID is required")
				os.Exit(1)
			}

			logger.WithFields(map[string]interface{}{
				"id":  cfg.Checkpoint.ID,
				"dir": cfg.Checkpoint.Directory,
			}).Info("Showing checkpoint")

			// Create checkpoint service
			checkpointSvc := service.NewCheckpointService(cfg, logger)

			// Get checkpoint details
			checkpoint, err := checkpointSvc.GetCheckpoint(ctx, cfg.Checkpoint.ID)
			if err != nil {
				logger.Error("Failed to get checkpoint", err)
				fmt.Printf("Error getting checkpoint: %s\n", err)
				os.Exit(1)
			}

			// Print checkpoint details
			fmt.Printf("Checkpoint ID: %s\n", checkpoint.ID)
			fmt.Printf("Created At: %s\n", checkpoint.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Source: %s\n", checkpoint.Source)
			fmt.Printf("Destination: %s\n", checkpoint.Destination)
			fmt.Printf("Status: %s\n", checkpoint.Status)
			fmt.Printf("Total Repositories: %d\n", checkpoint.TotalRepositories)
			fmt.Printf("Completed Repositories: %d\n", checkpoint.CompletedRepositories)
			fmt.Printf("Failed Repositories: %d\n", checkpoint.FailedRepositories)
			fmt.Printf("Total Tags Copied: %d\n", checkpoint.TotalTagsCopied)
			fmt.Printf("Total Tags Skipped: %d\n", checkpoint.TotalTagsSkipped)
			fmt.Printf("Total Errors: %d\n", checkpoint.TotalErrors)
			fmt.Printf("Total Bytes Transferred: %d\n", checkpoint.TotalBytesTransferred)

			// Print repository details
			if len(checkpoint.Repositories) > 0 {
				fmt.Println("\nRepositories:")
				fmt.Println("Name                                  | Status    | Tags Copied | Tags Skipped | Errors")
				fmt.Println("--------------------------------------|-----------|-------------|--------------|-------")
				for _, repo := range checkpoint.Repositories {
					fmt.Printf("%-36s | %-9s | %11d | %12d | %6d\n",
						repo.Name,
						repo.Status,
						repo.TagsCopied,
						repo.TagsSkipped,
						repo.Errors)
				}
			}
		},
	}
}

// newCheckpointDeleteCmd creates a new checkpoint delete command
func newCheckpointDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a checkpoint",
		Long:  `Deletes a specific checkpoint`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(cmd.Context())
			defer cancel()

			if cfg.Checkpoint.ID == "" {
				fmt.Println("Error: checkpoint ID is required")
				os.Exit(1)
			}

			logger.WithFields(map[string]interface{}{
				"id":  cfg.Checkpoint.ID,
				"dir": cfg.Checkpoint.Directory,
			}).Info("Deleting checkpoint")

			// Create checkpoint service
			checkpointSvc := service.NewCheckpointService(cfg, logger)

			// Delete checkpoint
			err := checkpointSvc.DeleteCheckpoint(ctx, cfg.Checkpoint.ID)
			if err != nil {
				logger.Error("Failed to delete checkpoint", err)
				fmt.Printf("Error deleting checkpoint: %s\n", err)
				os.Exit(1)
			}

			fmt.Printf("Checkpoint '%s' deleted successfully\n", cfg.Checkpoint.ID)
		},
	}
}

// newCheckpointExportCmd creates a new checkpoint export command
func newCheckpointExportCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a checkpoint to a file",
		Long:  `Exports a checkpoint to a JSON file for backup or sharing`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(cmd.Context())
			defer cancel()

			if cfg.Checkpoint.ID == "" {
				fmt.Println("Error: checkpoint ID is required")
				os.Exit(1)
			}

			if outputPath == "" {
				outputPath = fmt.Sprintf("checkpoint-%s.json", cfg.Checkpoint.ID)
			}

			logger.WithFields(map[string]interface{}{
				"id":     cfg.Checkpoint.ID,
				"dir":    cfg.Checkpoint.Directory,
				"output": outputPath,
			}).Info("Exporting checkpoint")

			// Create checkpoint service
			checkpointSvc := service.NewCheckpointService(cfg, logger)

			// Export checkpoint
			err := checkpointSvc.ExportCheckpoint(ctx, cfg.Checkpoint.ID, outputPath)
			if err != nil {
				logger.Error("Failed to export checkpoint", err)
				fmt.Printf("Error exporting checkpoint: %s\n", err)
				os.Exit(1)
			}

			fmt.Printf("Checkpoint '%s' exported to %s\n", cfg.Checkpoint.ID, outputPath)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path")
	return cmd
}

// newCheckpointImportCmd creates a new checkpoint import command
func newCheckpointImportCmd() *cobra.Command {
	var inputPath string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a checkpoint from a file",
		Long:  `Imports a checkpoint from a JSON file previously exported`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(cmd.Context())
			defer cancel()

			if inputPath == "" {
				fmt.Println("Error: input file path is required")
				os.Exit(1)
			}

			logger.WithFields(map[string]interface{}{
				"dir":   cfg.Checkpoint.Directory,
				"input": inputPath,
			}).Info("Importing checkpoint")

			// Create checkpoint service
			checkpointSvc := service.NewCheckpointService(cfg, logger)

			// Import checkpoint
			checkpoint, err := checkpointSvc.ImportCheckpoint(ctx, inputPath)
			if err != nil {
				logger.Error("Failed to import checkpoint", err)
				fmt.Printf("Error importing checkpoint: %s\n", err)
				os.Exit(1)
			}

			fmt.Printf("Checkpoint '%s' imported successfully\n", checkpoint.ID)
			fmt.Printf("Source: %s\n", checkpoint.Source)
			fmt.Printf("Destination: %s\n", checkpoint.Destination)
			fmt.Printf("Repositories: %d\n", checkpoint.TotalRepositories)
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "Input file path")
	if err := cmd.MarkFlagRequired("input"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}
	return cmd
}
