package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"freightliner/pkg/auth"
	"freightliner/pkg/client/factory"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login [REGISTRY]",
	Short: "Log in to a container registry",
	Long: `Authenticate with a container registry and store credentials securely.

Credentials are stored in ~/.docker/config.json by default, compatible with Docker and other tools.
Supports credential helpers (keychain, pass, secretservice) for secure storage.

Examples:
  # Login to Docker Hub
  freightliner login docker.io

  # Login to private registry
  freightliner login registry.company.com

  # Login with username (will prompt for password)
  freightliner login --username myuser registry.io

  # Login with environment variables
  export REGISTRY_USERNAME=myuser
  export REGISTRY_PASSWORD=mypass
  freightliner login registry.io
`,
	Args: cobra.ExactArgs(1),
	RunE: runLogin,
}

var (
	loginUsername string
	loginPassword string
	loginInsecure bool
)

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Username for authentication")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Password for authentication (insecure, use stdin or prompt)")
	loginCmd.Flags().BoolVar(&loginInsecure, "insecure", false, "Allow insecure connections (skip TLS verification)")
}

func runLogin(cmd *cobra.Command, args []string) error {
	registry := args[0]
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger := log.NewBasicLogger(log.InfoLevel)

	// Get username
	username := loginUsername
	if username == "" {
		username = os.Getenv("REGISTRY_USERNAME")
	}
	if username == "" {
		fmt.Print("Username: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read username: %w", err)
		}
		username = strings.TrimSpace(input)
	}

	if username == "" {
		return fmt.Errorf("username is required")
	}

	// Get password
	password := loginPassword
	if password == "" {
		password = os.Getenv("REGISTRY_PASSWORD")
	}
	if password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // New line after password input
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = string(passwordBytes)
	}

	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Test authentication by creating a client
	logger.WithFields(map[string]interface{}{
		"registry": registry,
	}).Info("Testing authentication...")

	registryConfig := config.RegistryConfig{
		Name:     registry,
		Type:     config.RegistryTypeGeneric,
		Endpoint: registry,
		Auth: config.AuthConfig{
			Type:     config.AuthTypeBasic,
			Username: username,
			Password: password,
		},
		Insecure: loginInsecure,
	}

	// Create factory and client to test credentials
	clientFactory := factory.NewRegistryClientFactory(logger)
	client, err := clientFactory.CreateClient(ctx, &registryConfig)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	// Test connection by listing repositories
	_, err = client.ListRepositories(ctx, "")
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"registry": registry,
		"username": username,
	}).Info("Authentication successful")

	// Store credentials
	store := auth.NewCredentialStore()
	err = store.Store(registry, username, password)
	if err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	fmt.Printf("Login Succeeded\n")
	logger.WithFields(map[string]interface{}{
		"registry": registry,
	}).Info("Credentials stored in Docker config")

	return nil
}
