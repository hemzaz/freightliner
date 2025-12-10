// Package client provides factory functions for creating registry clients
package client

import (
	"context"
	"strings"

	"freightliner/pkg/client/acr"
	"freightliner/pkg/client/dockerhub"
	"freightliner/pkg/client/ecr"
	"freightliner/pkg/client/gcr"
	"freightliner/pkg/client/generic"
	"freightliner/pkg/client/ghcr"
	"freightliner/pkg/client/harbor"
	"freightliner/pkg/client/quay"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"
)

// Factory creates registry clients based on configuration
type Factory struct {
	config *config.Config
	logger log.Logger
}

// NewFactory creates a new registry client factory
func NewFactory(cfg *config.Config, logger log.Logger) *Factory {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &Factory{
		config: cfg,
		logger: logger,
	}
}

// CreateECRClient creates an ECR client using the factory's configuration
func (f *Factory) CreateECRClient() (interfaces.RegistryClient, error) {
	return ecr.NewClient(ecr.ClientOptions{
		Region:    f.config.ECR.Region,
		AccountID: f.config.ECR.AccountID,
		Logger:    f.logger,
	})
}

// CreateGCRClient creates a GCR client using the factory's configuration
func (f *Factory) CreateGCRClient() (interfaces.RegistryClient, error) {
	return gcr.NewClient(gcr.ClientOptions{
		Project:  f.config.GCR.Project,
		Location: f.config.GCR.Location,
		Logger:   f.logger,
	})
}

// CreateDockerHubClient creates a Docker Hub client
func (f *Factory) CreateDockerHubClient(username, password string) (interfaces.RegistryClient, error) {
	return dockerhub.NewClient(dockerhub.ClientOptions{
		Username: username,
		Password: password,
		Logger:   f.logger,
	})
}

// CreateGHCRClient creates a GitHub Container Registry client
func (f *Factory) CreateGHCRClient(token, username string) (interfaces.RegistryClient, error) {
	return ghcr.NewClient(ghcr.ClientOptions{
		Token:    token,
		Username: username,
		Logger:   f.logger,
	})
}

// CreateACRClient creates an ACR client using the factory's configuration
func (f *Factory) CreateACRClient(registryName string, opts acr.ClientOptions) (interfaces.RegistryClient, error) {
	if opts.RegistryName == "" {
		opts.RegistryName = registryName
	}
	if opts.Logger == nil {
		opts.Logger = f.logger
	}
	return acr.NewClient(opts)
}

// CreateHarborClient creates a Harbor client using the factory's configuration
func (f *Factory) CreateHarborClient(opts harbor.ClientOptions) (interfaces.RegistryClient, error) {
	if opts.Logger == nil {
		opts.Logger = f.logger
	}
	return harbor.NewClient(opts)
}

// CreateQuayClient creates a Quay.io client using the factory's configuration
func (f *Factory) CreateQuayClient(opts quay.ClientOptions) (interfaces.RegistryClient, error) {
	if opts.Logger == nil {
		opts.Logger = f.logger
	}
	return quay.NewClient(opts)
}

// CreateCustomClient creates a custom registry client by name
func (f *Factory) CreateCustomClient(name string) (interfaces.RegistryClient, error) {
	// Find the registry configuration by name
	var regConfig *config.RegistryConfig
	for i := range f.config.Registries.Registries {
		if f.config.Registries.Registries[i].Name == name {
			regConfig = &f.config.Registries.Registries[i]
			break
		}
	}

	if regConfig == nil {
		return nil, errors.NotFoundf("registry '%s' not found in configuration", name)
	}

	return f.CreateClientFromConfig(*regConfig, name)
}

// CreateClientFromConfig creates a registry client from a registry configuration
func (f *Factory) CreateClientFromConfig(regConfig config.RegistryConfig, name string) (interfaces.RegistryClient, error) {
	// Determine registry type
	regType := strings.ToLower(string(regConfig.Type))
	if regType == "" {
		regType = "generic"
	}

	switch regType {
	case "ecr":
		// Create ECR client with configuration from registry config
		return ecr.NewClient(ecr.ClientOptions{
			Region:    f.getRegionFromConfig(regConfig),
			AccountID: f.getAccountIDFromConfig(regConfig),
			Logger:    f.logger,
		})

	case "gcr":
		// Create GCR client with configuration from registry config
		return gcr.NewClient(gcr.ClientOptions{
			Project:  f.getProjectFromConfig(regConfig),
			Location: f.getLocationFromConfig(regConfig),
			Logger:   f.logger,
		})

	case "dockerhub":
		// Create Docker Hub client
		return dockerhub.NewClient(dockerhub.ClientOptions{
			RegistryConfig: regConfig,
			Logger:         f.logger,
		})

	case "github":
		// Create GHCR client
		return ghcr.NewClient(ghcr.ClientOptions{
			RegistryConfig: regConfig,
			Logger:         f.logger,
		})

	case "acr", "azure":
		// Create ACR client with configuration from registry config
		return acr.NewClient(acr.ClientOptions{
			RegistryName:       f.getRegistryNameFromConfig(regConfig),
			TenantID:           f.getMetadata(regConfig, "tenantId", "tenant_id"),
			ClientID:           f.getMetadata(regConfig, "clientId", "client_id"),
			ClientSecret:       f.getMetadata(regConfig, "clientSecret", "client_secret"),
			UseManagedIdentity: f.getMetadata(regConfig, "useManagedIdentity") == "true",
			Logger:             f.logger,
		})

	case "harbor":
		// Create Harbor client with configuration from registry config
		return harbor.NewClient(harbor.ClientOptions{
			RegistryURL: regConfig.Endpoint,
			Username:    f.getUsernameFromConfig(regConfig),
			Password:    f.getPasswordFromConfig(regConfig),
			RobotName:   f.getMetadata(regConfig, "robotName", "robot_name"),
			RobotToken:  f.getMetadata(regConfig, "robotToken", "robot_token"),
			ProjectName: f.getMetadata(regConfig, "projectName", "project_name", "project"),
			Insecure:    f.getMetadata(regConfig, "insecure") == "true",
			Logger:      f.logger,
		})

	case "quay":
		// Create Quay client with configuration from registry config
		return quay.NewClient(quay.ClientOptions{
			RegistryURL:   regConfig.Endpoint,
			Username:      f.getUsernameFromConfig(regConfig),
			Password:      f.getPasswordFromConfig(regConfig),
			RobotUsername: f.getMetadata(regConfig, "robotUsername", "robot_username"),
			RobotToken:    f.getMetadata(regConfig, "robotToken", "robot_token"),
			OAuth2Token:   f.getMetadata(regConfig, "oauth2Token", "oauth2_token"),
			Organization:  f.getMetadata(regConfig, "organization", "org"),
			Insecure:      f.getMetadata(regConfig, "insecure") == "true",
			Logger:        f.logger,
		})

	case "generic", "docker", "gitlab", "artifactory":
		// Create generic client for all Docker v2 compatible registries
		return generic.NewClient(generic.ClientOptions{
			RegistryConfig: regConfig,
			RegistryName:   name,
			Logger:         f.logger,
		})

	default:
		return nil, errors.InvalidInputf("unsupported registry type: %s", regType)
	}
}

// CreateClientForRegistry creates a client for the specified registry endpoint
// This is a convenience method that tries to auto-detect the registry type
func (f *Factory) CreateClientForRegistry(ctx context.Context, registryURL string) (interfaces.RegistryClient, error) {
	// Normalize URL for comparison
	normalizedURL := strings.ToLower(registryURL)

	// Check for Docker Hub
	if strings.Contains(normalizedURL, "docker.io") ||
		strings.Contains(normalizedURL, "registry-1.docker.io") ||
		strings.Contains(normalizedURL, "index.docker.io") {
		f.logger.Info("Auto-detected Docker Hub registry")
		return f.CreateDockerHubClient("", "")
	}

	// Check for GitHub Container Registry
	if strings.Contains(normalizedURL, "ghcr.io") {
		f.logger.Info("Auto-detected GitHub Container Registry")
		return f.CreateGHCRClient("", "")
	}

	// Check for AWS ECR
	if strings.Contains(normalizedURL, ".dkr.ecr.") && strings.Contains(normalizedURL, ".amazonaws.com") {
		f.logger.Info("Auto-detected AWS ECR registry")
		return f.CreateECRClient()
	}

	// Check for Google Container Registry
	if strings.Contains(normalizedURL, "gcr.io") || strings.Contains(normalizedURL, "pkg.dev") {
		f.logger.Info("Auto-detected Google Container Registry")
		return f.CreateGCRClient()
	}

	// Check for Azure Container Registry
	if strings.Contains(normalizedURL, ".azurecr.io") {
		// Extract registry name from URL
		parts := strings.Split(registryURL, ".")
		if len(parts) > 0 {
			registryName := strings.TrimPrefix(parts[0], "https://")
			registryName = strings.TrimPrefix(registryName, "http://")
			f.logger.Info("Auto-detected Azure Container Registry")
			return f.CreateACRClient(registryName, acr.ClientOptions{
				UseManagedIdentity: true, // Default to managed identity
			})
		}
	}

	// Check if it's a configured custom registry
	for _, reg := range f.config.Registries.Registries {
		if strings.Contains(normalizedURL, reg.Endpoint) {
			f.logger.WithFields(map[string]interface{}{
				"registryName": reg.Name,
				"registryType": reg.Type,
			}).Info("Matched configured registry")
			return f.CreateCustomClient(reg.Name)
		}
	}

	// Fall back to generic client with anonymous auth
	f.logger.WithFields(map[string]interface{}{
		"registryURL": registryURL,
	}).Info("Using generic OCI registry client")

	return generic.NewClient(generic.ClientOptions{
		RegistryConfig: config.RegistryConfig{
			Name:     registryURL,
			Type:     config.RegistryTypeGeneric,
			Endpoint: registryURL,
			Auth: config.AuthConfig{
				Type: config.AuthTypeAnonymous,
			},
		},
		RegistryName: registryURL,
		Logger:       f.logger,
	})
}

// GetDefaultSourceRegistry returns the default registry for pulling images
func (f *Factory) GetDefaultSourceRegistry() string {
	if f.config.Registries.DefaultSource != "" {
		return f.config.Registries.DefaultSource
	}
	return "" // Empty means use source registry as-is
}

// GetDefaultDestinationRegistry returns the default registry for pushing images
func (f *Factory) GetDefaultDestinationRegistry() string {
	if f.config.Registries.DefaultDestination != "" {
		return f.config.Registries.DefaultDestination
	}
	return "" // Empty means use configured destination (ECR or GCR)
}

// ListCustomRegistries returns a list of all configured custom registry names
func (f *Factory) ListCustomRegistries() []string {
	names := make([]string, 0, len(f.config.Registries.Registries))
	for _, reg := range f.config.Registries.Registries {
		names = append(names, reg.Name)
	}
	return names
}

// Helper methods to extract configuration values

func (f *Factory) getRegionFromConfig(reg config.RegistryConfig) string {
	if reg.Region != "" {
		return reg.Region
	}
	if region, ok := reg.Metadata["region"]; ok {
		return region
	}
	return f.config.ECR.Region
}

func (f *Factory) getAccountIDFromConfig(reg config.RegistryConfig) string {
	if reg.AccountID != "" {
		return reg.AccountID
	}
	if accountID, ok := reg.Metadata["accountId"]; ok {
		return accountID
	}
	if accountID, ok := reg.Metadata["account_id"]; ok {
		return accountID
	}
	return f.config.ECR.AccountID
}

func (f *Factory) getProjectFromConfig(reg config.RegistryConfig) string {
	if reg.Project != "" {
		return reg.Project
	}
	if project, ok := reg.Metadata["project"]; ok {
		return project
	}
	return f.config.GCR.Project
}

func (f *Factory) getLocationFromConfig(reg config.RegistryConfig) string {
	if location, ok := reg.Metadata["location"]; ok {
		return location
	}
	return f.config.GCR.Location
}

func (f *Factory) getRegistryNameFromConfig(reg config.RegistryConfig) string {
	if name, ok := reg.Metadata["registryName"]; ok {
		return name
	}
	if name, ok := reg.Metadata["registry_name"]; ok {
		return name
	}
	// Try to extract from endpoint
	endpoint := strings.TrimSuffix(reg.Endpoint, ".azurecr.io")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return endpoint
}

func (f *Factory) getUsernameFromConfig(reg config.RegistryConfig) string {
	if reg.Auth.Username != "" {
		return reg.Auth.Username
	}
	if username, ok := reg.Metadata["username"]; ok {
		return username
	}
	return ""
}

func (f *Factory) getPasswordFromConfig(reg config.RegistryConfig) string {
	if reg.Auth.Password != "" {
		return reg.Auth.Password
	}
	if password, ok := reg.Metadata["password"]; ok {
		return password
	}
	return ""
}

func (f *Factory) getMetadata(reg config.RegistryConfig, keys ...string) string {
	for _, key := range keys {
		if value, ok := reg.Metadata[key]; ok {
			return value
		}
	}
	return ""
}

// CreateMultiRegistryClient creates a multi-registry client that can access multiple registries
// This is useful for cross-registry operations
func (f *Factory) CreateMultiRegistryClient(ctx context.Context) (interfaces.MultiRegistryClient, error) {
	// TODO: Implement multi-registry client when needed
	return nil, errors.NotImplementedf("multi-registry client not yet implemented")
}
