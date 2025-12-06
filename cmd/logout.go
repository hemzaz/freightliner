package cmd

import (
	"fmt"

	"freightliner/pkg/auth"
	"freightliner/pkg/helper/log"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout [REGISTRY]",
	Short: "Log out from a container registry",
	Long: `Remove stored credentials for a container registry.

Removes credentials from ~/.docker/config.json and credential helpers.

Examples:
  # Logout from Docker Hub
  freightliner logout docker.io

  # Logout from private registry
  freightliner logout registry.company.com

  # Logout from all registries
  freightliner logout --all
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogout,
}

var logoutAll bool

func init() {
	rootCmd.AddCommand(logoutCmd)

	logoutCmd.Flags().BoolVar(&logoutAll, "all", false, "Logout from all registries")
}

func runLogout(cmd *cobra.Command, args []string) error {
	logger := log.NewBasicLogger(log.InfoLevel)
	store := auth.NewCredentialStore()

	if logoutAll {
		// Get list of all registries
		registries, err := store.List()
		if err != nil {
			return fmt.Errorf("failed to list registries: %w", err)
		}

		if len(registries) == 0 {
			fmt.Println("Not logged in to any registries")
			return nil
		}

		// Logout from each registry
		for _, registry := range registries {
			if err := store.Delete(registry); err != nil {
				logger.WithFields(map[string]interface{}{
					"registry": registry,
					"error":    err.Error(),
				}).Warn("Failed to logout from registry")
				continue
			}
			fmt.Printf("Logged out from %s\n", registry)
		}

		logger.WithFields(map[string]interface{}{
			"count": len(registries),
		}).Info("Logged out from all registries")
		return nil
	}

	// Single registry logout
	if len(args) == 0 {
		return fmt.Errorf("registry argument is required (or use --all)")
	}

	registry := args[0]

	// Check if credentials exist
	_, _, err := store.Get(registry)
	if err != nil {
		return fmt.Errorf("not logged in to %s", registry)
	}

	// Delete credentials
	if err := store.Delete(registry); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	fmt.Printf("Removed credentials for %s\n", registry)
	logger.WithFields(map[string]interface{}{
		"registry": registry,
	}).Info("Logged out successfully")

	return nil
}
