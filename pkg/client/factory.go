// Package client provides factory functions for creating registry clients
package client

import (
	"context"
	"strings"

	"freightliner/pkg/client/ecr"
	"freightliner/pkg/client/gcr"
	"freightliner/pkg/client/generic"
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

	case "generic", "docker", "harbor", "quay", "gitlab", "github", "azure", "artifactory", "dockerhub":
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
	// Check if it's a known cloud provider registry
	if strings.Contains(registryURL, ".dkr.ecr.") && strings.Contains(registryURL, ".amazonaws.com") {
		return f.CreateECRClient()
	}

	if strings.Contains(registryURL, "gcr.io") || strings.Contains(registryURL, "pkg.dev") {
		return f.CreateGCRClient()
	}

	// Check if it's a configured custom registry
	for _, reg := range f.config.Registries.Registries {
		if strings.Contains(registryURL, reg.Endpoint) {
			return f.CreateCustomClient(reg.Name)
		}
	}

	// Fall back to generic client with anonymous auth
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

// CreateMultiRegistryClient creates a multi-registry client that can access multiple registries
// This is useful for cross-registry operations
func (f *Factory) CreateMultiRegistryClient(ctx context.Context) (interfaces.MultiRegistryClient, error) {
	// TODO: Implement multi-registry client when needed
	return nil, errors.NotImplementedf("multi-registry client not yet implemented")
}
