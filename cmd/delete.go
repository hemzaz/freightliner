package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var (
	deleteForce  bool
	deleteAll    bool
	deleteDryRun bool
)

// DeleteResult represents the result of a delete operation
type DeleteResult struct {
	Image   string
	Success bool
	Error   string
}

// newDeleteCmd creates the delete command
func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete IMAGE",
		Short: "Delete image from registry",
		Long: `Delete a container image from a registry by tag or digest.

IMAGE format: [TRANSPORT://]IMAGE[:TAG|@DIGEST]

Supported transports:
  docker://     Docker registry (default)

IMPORTANT: This operation is irreversible. Use with caution.

Examples:
  # Delete a specific tag
  freightliner delete docker://registry.io/repo:old-tag

  # Delete by digest
  freightliner delete docker://registry.io/repo@sha256:abc123...

  # Delete with confirmation prompt
  freightliner delete docker://registry.io/repo:v1.0

  # Force delete without confirmation
  freightliner delete --force docker://registry.io/repo:v1.0

  # Delete all tags in a repository (DANGEROUS!)
  freightliner delete --all --force docker://registry.io/repo

  # Dry run to see what would be deleted
  freightliner delete --dry-run docker://registry.io/repo:v1.0

Notes:
  - Most registries require authentication for delete operations
  - Some registries may require special permissions for deletion
  - Docker Hub does not support tag deletion via API
  - --all flag requires --force for safety
`,
		Args: cobra.ExactArgs(1),
		RunE: runDelete,
	}

	cmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&deleteAll, "all", false, "Delete all tags in repository (requires --force)")
	cmd.Flags().BoolVar(&deleteDryRun, "dry-run", false, "Show what would be deleted without actually deleting")

	return cmd
}

// runDelete executes the delete command
func runDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger, ctx, cancel := setupCommand(ctx)
	defer cancel()

	imageRef := args[0]

	// Safety check: --all requires --force
	if deleteAll && !deleteForce && !deleteDryRun {
		return fmt.Errorf("--all requires --force flag for safety")
	}

	// Parse the image reference
	transport, imgRef, err := parseImageReference(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference: %w", err)
	}

	if transport != "docker" && transport != "" {
		return fmt.Errorf("delete only supports docker:// transport")
	}

	logger.WithFields(map[string]interface{}{
		"image":   imgRef,
		"dry-run": deleteDryRun,
	}).Info("Preparing to delete image")

	// Parse reference
	ref, err := name.ParseReference(imgRef)
	if err != nil {
		return fmt.Errorf("invalid image reference: %w", err)
	}

	// Get authentication
	auth, err := getAuthForRegistry(ref.Context().RegistryStr())
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Using anonymous authentication")
		auth = authn.Anonymous
	}

	// Handle delete all
	if deleteAll {
		return deleteAllTags(ctx, logger, ref.Context().Name(), auth)
	}

	// Single image delete
	return deleteSingleImage(ctx, logger, ref, auth)
}

// deleteSingleImage deletes a single image by reference
func deleteSingleImage(ctx context.Context, logger log.Logger, ref name.Reference, auth authn.Authenticator) error {
	// Get the descriptor to find the digest
	desc, err := remote.Get(ref, remote.WithAuth(auth), remote.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to get image descriptor: %w", err)
	}

	digest := desc.Digest.String()
	logger.WithFields(map[string]interface{}{"digest": digest}).Info("Found image")

	// Confirm deletion
	if !deleteForce && !deleteDryRun {
		fmt.Printf("Are you sure you want to delete %s (%s)? [y/N]: ", ref.String(), digest)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Delete cancelled")
			return nil
		}
	}

	if deleteDryRun {
		fmt.Printf("[DRY RUN] Would delete: %s (%s)\n", ref.String(), digest)
		return nil
	}

	// Perform deletion
	logger.WithFields(map[string]interface{}{"reference": ref.String()}).Info("Deleting image")
	err = remote.Delete(ref, remote.WithAuth(auth), remote.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	fmt.Printf("Successfully deleted: %s\n", ref.String())
	logger.WithFields(map[string]interface{}{"reference": ref.String()}).Info("Image deleted successfully")

	return nil
}

// deleteAllTags deletes all tags in a repository
func deleteAllTags(ctx context.Context, logger log.Logger, repoName string, auth authn.Authenticator) error {
	repo, err := name.NewRepository(repoName)
	if err != nil {
		return fmt.Errorf("invalid repository: %w", err)
	}

	// List all tags
	tags, err := remote.List(repo, remote.WithAuth(auth), remote.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		fmt.Println("No tags found in repository")
		return nil
	}

	logger.WithFields(map[string]interface{}{"count": len(tags)}).Info("Found tags to delete")

	// Confirm deletion
	if !deleteForce && !deleteDryRun {
		fmt.Printf("Are you sure you want to delete ALL %d tags in %s? [y/N]: ", len(tags), repoName)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Delete cancelled")
			return nil
		}
	}

	// Delete each tag
	var results []DeleteResult
	successCount := 0
	failCount := 0

	for _, tag := range tags {
		tagRef := fmt.Sprintf("%s:%s", repoName, tag)
		ref, err := name.ParseReference(tagRef)
		if err != nil {
			logger.WithFields(map[string]interface{}{"tag": tag, "error": err.Error()}).Warn("Failed to parse tag reference")
			results = append(results, DeleteResult{
				Image:   tagRef,
				Success: false,
				Error:   err.Error(),
			})
			failCount++
			continue
		}

		if deleteDryRun {
			fmt.Printf("[DRY RUN] Would delete: %s\n", tagRef)
			results = append(results, DeleteResult{
				Image:   tagRef,
				Success: true,
			})
			successCount++
			continue
		}

		err = remote.Delete(ref, remote.WithAuth(auth), remote.WithContext(ctx))
		if err != nil {
			logger.WithFields(map[string]interface{}{"tag": tag, "error": err.Error()}).Warn("Failed to delete tag")
			fmt.Printf("Failed to delete %s: %v\n", tagRef, err)
			results = append(results, DeleteResult{
				Image:   tagRef,
				Success: false,
				Error:   err.Error(),
			})
			failCount++
		} else {
			logger.WithFields(map[string]interface{}{"tag": tag}).Info("Deleted tag")
			fmt.Printf("Deleted: %s\n", tagRef)
			results = append(results, DeleteResult{
				Image:   tagRef,
				Success: true,
			})
			successCount++
		}
	}

	// Summary
	fmt.Printf("\nDelete Summary:\n")
	fmt.Printf("  Total: %d\n", len(tags))
	fmt.Printf("  Success: %d\n", successCount)
	fmt.Printf("  Failed: %d\n", failCount)

	if failCount > 0 {
		fmt.Println("\nFailed deletions:")
		for _, result := range results {
			if !result.Success {
				fmt.Printf("  %s: %s\n", result.Image, result.Error)
			}
		}
		return fmt.Errorf("failed to delete %d tags", failCount)
	}

	return nil
}
