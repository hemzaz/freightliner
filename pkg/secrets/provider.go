package secrets

import (
	"context"
	"fmt"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/secrets/aws"
	"freightliner/pkg/secrets/gcp"
)

// Provider defines the interface for secret management across different providers.
// This interface should be used by any package that needs secrets management
// functionality, following the Dependency Inversion Principle.
type Provider interface {
	// GetSecret retrieves a secret value by name
	GetSecret(ctx context.Context, secretName string) (string, error)

	// GetJSONSecret retrieves a JSON-formatted secret and unmarshal it into the provided struct
	GetJSONSecret(ctx context.Context, secretName string, v interface{}) error

	// PutSecret creates or updates a secret value
	PutSecret(ctx context.Context, secretName, secretValue string) error

	// PutJSONSecret marshals a struct to JSON and stores it as a secret
	PutJSONSecret(ctx context.Context, secretName string, v interface{}) error

	// DeleteSecret deletes a secret
	DeleteSecret(ctx context.Context, secretName string) error
}

// ProviderType defines the supported secret manager providers
type ProviderType string

const (
	// AWSProvider is AWS Secrets Manager
	AWSProvider ProviderType = "aws"

	// GCPProvider is Google Secret Manager
	GCPProvider ProviderType = "gcp"
)

// ManagerOptions contains configuration for creating a secret manager
type ManagerOptions struct {
	// Provider is the type of secret manager to use
	Provider ProviderType

	// Logger is the logger instance to use
	Logger *log.Logger

	// AWS-specific options
	AWSRegion string

	// GCP-specific options
	GCPProject         string
	GCPCredentialsFile string
}

// GetProvider creates and returns a secret provider based on the specified type
func GetProvider(ctx context.Context, opts ManagerOptions) (Provider, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	switch opts.Provider {
	case AWSProvider:
		return aws.NewProvider(ctx, aws.ProviderOptions{
			Region: opts.AWSRegion,
			Logger: opts.Logger,
		})
	case GCPProvider:
		return gcp.NewProvider(ctx, gcp.ProviderOptions{
			Project:         opts.GCPProject,
			CredentialsFile: opts.GCPCredentialsFile,
			Logger:          opts.Logger,
		})
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", opts.Provider)
	}
}
