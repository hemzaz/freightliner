package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"freightliner/pkg/auth"
	"freightliner/pkg/client/factory"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/spf13/cobra"
)

// newAuthCmd creates the auth command
func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication management",
		Long: `Manage container registry authentication.

Available subcommands:
  list  - List stored credentials
  test  - Test registry credentials`,
	}

	cmd.AddCommand(newAuthListCmd())
	cmd.AddCommand(newAuthTestCmd())

	return cmd
}

// newAuthListCmd creates the auth list command
func newAuthListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stored credentials",
		Long: `List all stored container registry credentials.

Shows the registries for which credentials are stored, along with
the username and last used timestamp (if available).

Examples:
  # List all stored credentials
  freightliner auth list`,
		Args: cobra.NoArgs,
		RunE: runAuthList,
	}

	return cmd
}

// newAuthTestCmd creates the auth test command
func newAuthTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test REGISTRY",
		Short: "Test registry credentials",
		Long: `Test stored credentials by attempting authentication.

Validates that the stored credentials work by attempting to connect
to the registry and perform a basic operation.

Examples:
  # Test Docker Hub credentials
  freightliner auth test docker.io

  # Test private registry credentials
  freightliner auth test registry.company.com`,
		Args: cobra.ExactArgs(1),
		RunE: runAuthTest,
	}

	return cmd
}

// runAuthList executes the auth list command
func runAuthList(cmd *cobra.Command, args []string) error {
	logger := log.NewBasicLogger(log.InfoLevel)
	store := auth.NewCredentialStore()

	// Get list of registries
	registries, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to list registries: %w", err)
	}

	if len(registries) == 0 {
		fmt.Println("No stored credentials found")
		logger.Info("No credentials stored")
		return nil
	}

	// Print table header
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "REGISTRY\tUSERNAME\tSTORED\n")
	fmt.Fprintf(w, "--------\t--------\t------\n")

	// List each registry
	for _, registry := range registries {
		username, _, err := store.Get(registry)
		if err != nil {
			logger.WithFields(map[string]interface{}{
				"registry": registry,
				"error":    err.Error(),
			}).Warn("Failed to get credentials")
			continue
		}

		// Show registry and username
		fmt.Fprintf(w, "%s\t%s\t%s\n", registry, username, "✓")
	}

	logger.WithFields(map[string]interface{}{
		"count": len(registries),
	}).Info("Listed credentials")

	return nil
}

// runAuthTest executes the auth test command
func runAuthTest(cmd *cobra.Command, args []string) error {
	registry := args[0]
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger := log.NewBasicLogger(log.InfoLevel)
	store := auth.NewCredentialStore()

	// Get credentials
	username, password, err := store.Get(registry)
	if err != nil {
		return fmt.Errorf("no credentials found for %s: %w", registry, err)
	}

	fmt.Printf("Testing authentication to %s...\n", registry)
	logger.WithFields(map[string]interface{}{
		"registry": registry,
		"username": username,
	}).Info("Testing credentials")

	// Create registry config
	registryConfig := config.RegistryConfig{
		Name:     registry,
		Type:     config.RegistryTypeGeneric,
		Endpoint: registry,
		Auth: config.AuthConfig{
			Type:     config.AuthTypeBasic,
			Username: username,
			Password: password,
		},
		Insecure: false,
	}

	// Create client factory
	clientFactory := factory.NewRegistryClientFactory(logger)

	// Create client
	client, err := clientFactory.CreateClient(ctx, &registryConfig)
	if err != nil {
		fmt.Printf("✗ Authentication failed: %v\n", err)
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Test connection by listing repositories
	_, err = client.ListRepositories(ctx, "")
	if err != nil {
		fmt.Printf("✗ Authentication failed: %v\n", err)
		return fmt.Errorf("authentication test failed: %w", err)
	}

	// Success
	fmt.Printf("✓ Authentication successful\n")
	fmt.Printf("Registry: %s\n", registry)
	fmt.Printf("Username: %s\n", username)

	logger.WithFields(map[string]interface{}{
		"registry": registry,
		"username": username,
	}).Info("Authentication test passed")

	return nil
}
