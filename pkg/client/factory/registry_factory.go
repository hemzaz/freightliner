// Package factory provides a factory for creating registry clients
package factory

import (
	"context"
	"fmt"

	"freightliner/pkg/client/ecr"
	"freightliner/pkg/client/gcr"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"
)

// RegistryClientFactory creates registry clients based on configuration
type RegistryClientFactory struct {
	logger log.Logger
}

// NewRegistryClientFactory creates a new registry client factory
func NewRegistryClientFactory(logger log.Logger) *RegistryClientFactory {
	if logger == nil {
		logger = log.NewLogger()
	}
	return &RegistryClientFactory{
		logger: logger,
	}
}

// CreateClient creates a registry client based on the provided configuration
func (f *RegistryClientFactory) CreateClient(ctx context.Context, regConfig *config.RegistryConfig) (interfaces.RegistryClient, error) {
	if regConfig == nil {
		return nil, fmt.Errorf("registry configuration is nil")
	}

	// Validate the configuration
	if err := regConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid registry configuration: %w", err)
	}

	f.logger.WithFields(map[string]interface{}{
		"registry": regConfig.Name,
		"type":     regConfig.Type,
	}).Info("Creating registry client")

	// Create client based on registry type
	switch regConfig.Type {
	case config.RegistryTypeECR:
		return f.createECRClient(ctx, regConfig)
	case config.RegistryTypeGCR:
		return f.createGCRClient(ctx, regConfig)
	case config.RegistryTypeDockerHub, config.RegistryTypeHarbor, config.RegistryTypeQuay,
		config.RegistryTypeGitLab, config.RegistryTypeGitHub, config.RegistryTypeGeneric:
		return f.createGenericClient(ctx, regConfig)
	case config.RegistryTypeAzure:
		return nil, fmt.Errorf("Azure Container Registry support not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", regConfig.Type)
	}
}

// createECRClient creates an AWS ECR client
func (f *RegistryClientFactory) createECRClient(ctx context.Context, regConfig *config.RegistryConfig) (interfaces.RegistryClient, error) {
	opts := ecr.ClientOptions{
		Region:    regConfig.Region,
		AccountID: regConfig.AccountID,
		Logger:    f.logger,
	}

	// Configure authentication
	if regConfig.Auth.Profile != "" {
		opts.Profile = regConfig.Auth.Profile
	}
	if regConfig.Auth.RoleARN != "" {
		opts.RoleARN = regConfig.Auth.RoleARN
	}
	if regConfig.Auth.CredentialsFile != "" {
		opts.CredentialsFile = regConfig.Auth.CredentialsFile
	}

	client, err := ecr.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECR client: %w", err)
	}

	return client, nil
}

// createGCRClient creates a Google Container Registry client
func (f *RegistryClientFactory) createGCRClient(ctx context.Context, regConfig *config.RegistryConfig) (interfaces.RegistryClient, error) {
	opts := gcr.ClientOptions{
		Project:  regConfig.Project,
		Location: regConfig.Region,
		Logger:   f.logger,
	}

	// Configure authentication
	if regConfig.Auth.CredentialsFile != "" {
		opts.CredentialsFile = regConfig.Auth.CredentialsFile
	}

	client, err := gcr.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCR client: %w", err)
	}

	return client, nil
}

// createGenericClient creates a generic OCI registry client
// This handles DockerHub, Harbor, Quay, GitLab, GitHub, and other OCI-compliant registries
func (f *RegistryClientFactory) createGenericClient(ctx context.Context, regConfig *config.RegistryConfig) (interfaces.RegistryClient, error) {
	// For now, we'll return an error as generic client implementation is not yet complete
	// This would require implementing a new generic client that uses go-containerregistry
	// with configurable authentication
	return nil, fmt.Errorf("generic registry support for %s not yet implemented", regConfig.Type)
}

// CreateClientByName creates a registry client by looking up the registry name in the configuration
func (f *RegistryClientFactory) CreateClientByName(ctx context.Context, registryName string, registriesConfig *config.RegistriesConfig) (interfaces.RegistryClient, error) {
	if registriesConfig == nil {
		return nil, fmt.Errorf("registries configuration is nil")
	}

	regConfig, err := registriesConfig.GetByName(registryName)
	if err != nil {
		return nil, fmt.Errorf("failed to find registry %s: %w", registryName, err)
	}

	return f.CreateClient(ctx, regConfig)
}

// CreateSourceAndDestClients creates both source and destination clients
func (f *RegistryClientFactory) CreateSourceAndDestClients(
	ctx context.Context,
	sourceRegistry, destRegistry string,
	registriesConfig *config.RegistriesConfig,
) (source, dest interfaces.RegistryClient, err error) {
	// Use default source if not specified
	if sourceRegistry == "" && registriesConfig.DefaultSource != "" {
		sourceRegistry = registriesConfig.DefaultSource
		f.logger.WithField("registry", sourceRegistry).Info("Using default source registry")
	}

	// Use default destination if not specified
	if destRegistry == "" && registriesConfig.DefaultDestination != "" {
		destRegistry = registriesConfig.DefaultDestination
		f.logger.WithField("registry", destRegistry).Info("Using default destination registry")
	}

	if sourceRegistry == "" {
		return nil, nil, fmt.Errorf("source registry not specified and no default configured")
	}
	if destRegistry == "" {
		return nil, nil, fmt.Errorf("destination registry not specified and no default configured")
	}

	// Create source client
	source, err = f.CreateClientByName(ctx, sourceRegistry, registriesConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create source client: %w", err)
	}

	// Create destination client
	dest, err = f.CreateClientByName(ctx, destRegistry, registriesConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create destination client: %w", err)
	}

	return source, dest, nil
}

// ValidateRegistryConnection validates that a connection can be established to the registry
func (f *RegistryClientFactory) ValidateRegistryConnection(ctx context.Context, regConfig *config.RegistryConfig) error {
	client, err := f.CreateClient(ctx, regConfig)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Try to list repositories as a connection test
	// Note: This assumes the client implements RegistryClient interface
	// with a ListRepositories method
	if lister, ok := client.(interfaces.RepositoryLister); ok {
		_, err := lister.ListRepositories(ctx, "")
		if err != nil {
			return fmt.Errorf("connection validation failed: %w", err)
		}
		f.logger.WithField("registry", regConfig.Name).Info("Registry connection validated successfully")
		return nil
	}

	// If ListRepositories is not available, consider the client creation as sufficient validation
	f.logger.WithField("registry", regConfig.Name).Info("Registry client created successfully (connection not validated)")
	return nil
}

// GetSupportedRegistryTypes returns a list of supported registry types
func (f *RegistryClientFactory) GetSupportedRegistryTypes() []config.RegistryType {
	return []config.RegistryType{
		config.RegistryTypeECR,
		config.RegistryTypeGCR,
		// Generic types are listed but return not implemented error
		config.RegistryTypeDockerHub,
		config.RegistryTypeHarbor,
		config.RegistryTypeQuay,
		config.RegistryTypeGitLab,
		config.RegistryTypeGitHub,
		config.RegistryTypeGeneric,
	}
}

// IsRegistryTypeSupported checks if a registry type is supported
func (f *RegistryClientFactory) IsRegistryTypeSupported(registryType config.RegistryType) bool {
	supported := f.GetSupportedRegistryTypes()
	for _, t := range supported {
		if t == registryType {
			return true
		}
	}
	return false
}
