package common

import (
	"context"
	"fmt"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// CreateTransport creates a transport for registry operations with authentication
// This centralizes the common transport creation logic used in both ECR and GCR clients
func CreateTransport(registry name.Registry, auth authn.Authenticator, logger *log.Logger) (http.RoundTripper, error) {
	scopes := []string{
		fmt.Sprintf("repository:%s:pull,push", registry.String()),
	}

	// Create transport with authentication and scopes
	rt, err := transport.New(
		registry,
		auth,
		http.DefaultTransport,
		scopes,
	)
	if err != nil {
		logger.Error("Failed to create transport", err, map[string]interface{}{
			"registry": registry.String(),
		})
		return nil, errors.Wrap(err, "failed to create transport")
	}

	return rt, nil
}

// RegistryClientOptions provides common options for creating registry clients
type RegistryClientOptions struct {
	Logger   *log.Logger
	Registry string // Registry hostname
	Region   string // AWS region/GCP location
	Project  string // GCP project
	Account  string // AWS account ID
}

// BaseClient provides common functionality for registry clients
type BaseClient struct {
	Logger     *log.Logger
	Registry   string
	Options    RegistryClientOptions
	HTTPClient *http.Client
}

// ListRepositories lists all repositories in a registry with the given prefix
func (c *BaseClient) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	c.Logger.Debug("Listing repositories", map[string]interface{}{
		"registry": c.Registry,
		"prefix":   prefix,
	})

	// Implement in specific clients as required interface method.
	// This is a placeholder to satisfy the interface.
	return nil, errors.NotImplementedf("method not implemented in base client")
}

// GetRepository returns a repository reference for the given name
func (c *BaseClient) GetRepository(ctx context.Context, name string) (Repository, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	c.Logger.Debug("Getting repository", map[string]interface{}{
		"registry": c.Registry,
		"name":     name,
	})

	// Implement in specific clients as required interface method.
	// This is a placeholder to satisfy the interface.
	return nil, errors.NotImplementedf("method not implemented in base client")
}

// GetRegistryName returns the name of the registry
func (c *BaseClient) GetRegistryName() string {
	return c.Registry
}

// CommonAuthOptions provides shared authentication options
type CommonAuthOptions struct {
	Registry string
	Region   string
	Project  string
	Account  string
}

// FormRegistryPath creates a properly formatted registry path
// This replaces the duplicated path construction logic in both client implementations
func FormRegistryPath(registry, name string) string {
	return fmt.Sprintf("%s/%s", registry, name)
}

// RegistryClient defines the interface for registry clients
type RegistryClient interface {
	// ListRepositories lists all repositories in a registry with the given prefix
	ListRepositories(ctx context.Context, prefix string) ([]string, error)

	// GetRepository returns a repository reference for the given name
	GetRepository(ctx context.Context, name string) (Repository, error)

	// GetRegistryName returns the name of the registry
	GetRegistryName() string
}
