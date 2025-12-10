// Package acr provides Azure Container Registry client functionality.
package acr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// Client implements the registry client interface for Azure Container Registry
type Client struct {
	registryName string
	auth         *ACRAuthenticator
	logger       log.Logger
	transportOpt remote.Option
}

// ClientOptions provides configuration for connecting to ACR
type ClientOptions struct {
	// RegistryName is the ACR registry name (without .azurecr.io)
	RegistryName string

	// AuthConfig contains authentication configuration
	AuthConfig *AuthConfig

	// Logger is the logger to use
	Logger log.Logger

	// TenantID for service principal auth
	TenantID string

	// ClientID for service principal auth
	ClientID string

	// ClientSecret for service principal auth
	ClientSecret string

	// UseManagedIdentity enables managed identity authentication
	UseManagedIdentity bool
}

// NewClient creates a new ACR client
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.RegistryName == "" {
		return nil, errors.InvalidInputf("registry name is required")
	}

	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Create auth config if not provided
	var authConfig *AuthConfig
	if opts.AuthConfig != nil {
		authConfig = opts.AuthConfig
	} else {
		authConfig = &AuthConfig{
			RegistryName:       opts.RegistryName,
			TenantID:           opts.TenantID,
			ClientID:           opts.ClientID,
			ClientSecret:       opts.ClientSecret,
			UseManagedIdentity: opts.UseManagedIdentity,
		}
	}

	// Create authenticator
	auth, err := NewACRAuthenticator(authConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ACR authenticator")
	}

	return &Client{
		registryName: opts.RegistryName,
		auth:         auth,
		logger:       opts.Logger,
		transportOpt: remote.WithAuth(auth),
	}, nil
}

// GetRegistryName returns the ACR registry endpoint
func (c *Client) GetRegistryName() string {
	return fmt.Sprintf("%s.azurecr.io", c.registryName)
}

// ListRepositories lists all repositories in the registry
func (c *Client) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	registryURL := c.GetRegistryName()
	registry, err := name.NewRegistry(registryURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create registry reference")
	}

	// Use the catalog API to list repositories
	catalogURL := fmt.Sprintf("https://%s/v2/_catalog", registryURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, catalogURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create catalog request")
	}

	// Get authenticated transport
	tr, err := transport.NewWithContext(
		ctx,
		registry,
		c.auth,
		http.DefaultTransport,
		[]string{registry.Scope("")},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transport")
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch catalog")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.InvalidInputf("catalog request failed with status: %s", resp.Status)
	}

	// Parse catalog response
	var catalog struct {
		Repositories []string `json:"repositories"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read catalog response")
	}

	if err := json.Unmarshal(body, &catalog); err != nil {
		return nil, errors.Wrap(err, "failed to parse catalog response")
	}

	// Filter by prefix if provided
	var repositories []string
	for _, repo := range catalog.Repositories {
		if prefix == "" || strings.HasPrefix(repo, prefix) {
			repositories = append(repositories, repo)
		}
	}

	return repositories, nil
}

// GetRepository returns a repository by name
func (c *Client) GetRepository(ctx context.Context, repoName string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Create a proper repository reference
	registry := c.GetRegistryName()
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", registry, repoName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	return &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}, nil
}

// CreateRepository creates a new repository in ACR
// Note: ACR automatically creates repositories when the first image is pushed
func (c *Client) CreateRepository(ctx context.Context, repoName string, tags map[string]string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	c.logger.WithFields(map[string]interface{}{
		"repository": repoName,
		"registry":   c.GetRegistryName(),
	}).Info("Creating repository reference (ACR auto-creates on first push)")

	// Create the repository reference
	registry := c.GetRegistryName()
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", registry, repoName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Note: ACR repositories are created automatically on first push
	// Tags are not supported at repository creation time in ACR

	return &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}, nil
}

// GetTransport returns an authenticated HTTP transport for ACR
func (c *Client) GetTransport(repositoryName string) (http.RoundTripper, error) {
	// Create a proper repository reference
	registry := c.GetRegistryName()
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", registry, repositoryName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Create transport with authentication
	rt, err := transport.NewWithContext(
		context.Background(),
		repository.Registry,
		c.auth,
		http.DefaultTransport,
		[]string{repository.Scope(transport.PushScope)},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ACR transport")
	}

	return rt, nil
}

// GetRemoteOptions returns options for the go-containerregistry remote package
func (c *Client) GetRemoteOptions() []remote.Option {
	return []remote.Option{
		c.transportOpt,
	}
}

// RefreshAuth refreshes the authentication token
func (c *Client) RefreshAuth() error {
	return c.auth.RefreshToken()
}

// GetRegistryURL returns the full ACR registry URL
func (c *Client) GetRegistryURL() string {
	return c.auth.GetRegistryURL()
}
