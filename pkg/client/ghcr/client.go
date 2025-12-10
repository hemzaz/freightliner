// Package ghcr provides GitHub Container Registry client functionality.
package ghcr

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"freightliner/pkg/client/common"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

const (
	// GHCRRegistry is the GitHub Container Registry endpoint
	GHCRRegistry = "ghcr.io"

	// GHCRAPIEndpoint is the GitHub API endpoint
	GHCRAPIEndpoint = "https://api.github.com"

	// DefaultGHCRTimeout is the default timeout for GHCR operations
	DefaultGHCRTimeout = 30
)

// Client implements the registry client interface for GitHub Container Registry
type Client struct {
	registry      string
	logger        log.Logger
	authenticator authn.Authenticator
	token         string
	username      string
	transportOpt  remote.Option
}

// ClientOptions provides configuration for connecting to GHCR
type ClientOptions struct {
	// Token is the GitHub Personal Access Token or GitHub Actions token
	Token string

	// Username is the GitHub username (optional, can be extracted from token)
	Username string

	// RegistryConfig contains the registry configuration
	RegistryConfig config.RegistryConfig

	// Logger is the logger to use
	Logger log.Logger
}

// NewClient creates a new GHCR client
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Determine authentication method
	token := opts.Token
	username := opts.Username

	// Try to get token from config if not provided
	if token == "" && opts.RegistryConfig.Auth.Type == config.AuthTypeToken {
		token = opts.RegistryConfig.Auth.Token
	}

	// Try to get token from environment variables
	if token == "" {
		if ghToken := os.Getenv("GITHUB_TOKEN"); ghToken != "" {
			token = ghToken
			opts.Logger.Info("Using GITHUB_TOKEN from environment")
		} else if ghToken := os.Getenv("GH_TOKEN"); ghToken != "" {
			token = ghToken
			opts.Logger.Info("Using GH_TOKEN from environment")
		} else if ghcrToken := os.Getenv("GHCR_TOKEN"); ghcrToken != "" {
			token = ghcrToken
			opts.Logger.Info("Using GHCR_TOKEN from environment")
		}
	}

	// Try to get username from config if not provided
	if username == "" && opts.RegistryConfig.Auth.Username != "" {
		username = opts.RegistryConfig.Auth.Username
	}

	// Create authenticator
	var auth authn.Authenticator
	if token != "" {
		// GHCR uses the token as the password with any username (often "USERNAME" or the actual username)
		if username == "" {
			username = "USERNAME" // GHCR accepts any username when using PAT
		}
		auth = &authn.Basic{
			Username: username,
			Password: token,
		}
		opts.Logger.WithFields(map[string]interface{}{
			"username": username,
		}).Info("Using authenticated GHCR access")
	} else {
		auth = authn.Anonymous
		opts.Logger.Warn("Using anonymous GHCR access (only public repositories accessible)")
	}

	return &Client{
		registry:      GHCRRegistry,
		logger:        opts.Logger,
		authenticator: auth,
		token:         token,
		username:      username,
		transportOpt:  remote.WithAuth(auth),
	}, nil
}

// GetRegistryName returns the GHCR registry endpoint
func (c *Client) GetRegistryName() string {
	return c.registry
}

// ListRepositories lists all repositories accessible with the current authentication
// Note: GHCR's catalog API is limited
func (c *Client) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	// Parse registry
	reg, err := name.NewRegistry(c.registry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse GHCR registry")
	}

	// Try to list repositories using the catalog API
	repos, err := remote.Catalog(ctx, reg, c.transportOpt)
	if err != nil {
		// GHCR catalog API is limited and may not work for all use cases
		c.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("GHCR catalog API failed, use GetRepository for known repositories")
		return nil, errors.Wrap(err, "failed to list GHCR repositories")
	}

	// Filter by prefix if specified
	if prefix == "" {
		return repos, nil
	}

	filtered := make([]string, 0)
	for _, repo := range repos {
		if strings.HasPrefix(repo, prefix) {
			filtered = append(filtered, repo)
		}
	}

	c.logger.WithFields(map[string]interface{}{
		"total":    len(repos),
		"filtered": len(filtered),
		"prefix":   prefix,
	}).Debug("Listed GHCR repositories")

	return filtered, nil
}

// GetRepository returns a repository by name
func (c *Client) GetRepository(ctx context.Context, repoName string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Normalize repository name for GHCR
	normalizedName := c.normalizeRepositoryName(repoName)

	// Create repository reference
	fullRepoName := fmt.Sprintf("%s/%s", c.registry, normalizedName)
	repository, err := name.NewRepository(fullRepoName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	c.logger.WithFields(map[string]interface{}{
		"repository":   repoName,
		"normalized":   normalizedName,
		"fullRepoName": fullRepoName,
	}).Debug("Created GHCR repository reference")

	// Create base repository with common functionality
	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       normalizedName,
		Repository: repository,
		Logger:     c.logger,
	})

	return &Repository{
		BaseRepository: baseRepo,
		client:         c,
		name:           normalizedName,
		repository:     repository,
	}, nil
}

// GetTransport returns an authenticated HTTP transport for GHCR
func (c *Client) GetTransport(repositoryName string) (http.RoundTripper, error) {
	// Normalize repository name
	normalizedName := c.normalizeRepositoryName(repositoryName)

	// Create repository reference
	fullRepoName := fmt.Sprintf("%s/%s", c.registry, normalizedName)
	repository, err := name.NewRepository(fullRepoName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Create transport with authentication
	rt, err := transport.NewWithContext(
		context.Background(),
		repository.Registry,
		c.authenticator,
		http.DefaultTransport,
		[]string{repository.Scope(transport.PullScope)},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GHCR transport")
	}

	return rt, nil
}

// GetRemoteOptions returns options for the go-containerregistry remote package
func (c *Client) GetRemoteOptions() []remote.Option {
	return []remote.Option{
		c.transportOpt,
	}
}

// normalizeRepositoryName normalizes GHCR repository names
// GHCR repositories follow the pattern: owner/repo or org/repo
func (c *Client) normalizeRepositoryName(repoName string) string {
	// Remove leading slashes
	repoName = strings.TrimPrefix(repoName, "/")

	// Remove ghcr.io prefix if present
	repoName = strings.TrimPrefix(repoName, GHCRRegistry+"/")

	// Ensure lowercase (GHCR requires lowercase)
	repoName = strings.ToLower(repoName)

	return repoName
}

// IsPublicRepository checks if a repository is publicly accessible
func (c *Client) IsPublicRepository(ctx context.Context, repoName string) (bool, error) {
	// Try to list tags without authentication
	normalizedName := c.normalizeRepositoryName(repoName)
	fullRepoName := fmt.Sprintf("%s/%s", c.registry, normalizedName)

	repository, err := name.NewRepository(fullRepoName)
	if err != nil {
		return false, errors.Wrap(err, "failed to create repository reference")
	}

	// Try with anonymous auth
	_, err = remote.List(repository, remote.WithAuth(authn.Anonymous))
	if err != nil {
		// If we get an auth error, it's likely private
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
			return false, nil
		}
		// Other errors are ambiguous
		return false, errors.Wrap(err, "failed to check repository visibility")
	}

	return true, nil
}

// GetPackageVisibility returns the visibility of a package (public/private)
// This requires the GitHub API
func (c *Client) GetPackageVisibility(ctx context.Context, owner, packageName string) (string, error) {
	if c.token == "" {
		return "", errors.InvalidInputf("GitHub token required to check package visibility")
	}

	// This would require implementing GitHub API calls
	// For now, we return a not implemented error
	c.logger.Warn("GetPackageVisibility requires GitHub API integration (not yet implemented)")
	return "", errors.NotImplementedf("GetPackageVisibility requires GitHub API integration")
}
