// Package generic provides a generic Docker Registry v2 compatible client.
package generic

import (
	"context"
	"crypto/tls"
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

// Client implements the registry client interface for generic Docker v2 registries
type Client struct {
	registry      string
	registryConf  config.RegistryConfig
	logger        log.Logger
	authenticator authn.Authenticator
	transportOpt  remote.Option
}

// ClientOptions provides configuration for connecting to a generic registry
type ClientOptions struct {
	// RegistryConfig contains the registry configuration
	RegistryConfig config.RegistryConfig

	// RegistryName is a friendly name for the registry
	RegistryName string

	// Logger is the logger to use
	Logger log.Logger
}

// NewClient creates a new generic registry client
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.RegistryConfig.Endpoint == "" {
		return nil, errors.InvalidInputf("registry endpoint is required")
	}

	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Normalize registry URL (remove http/https scheme if present)
	registry := normalizeRegistryURL(opts.RegistryConfig.Endpoint)

	// Create authenticator based on auth type
	auth, err := createAuthenticator(opts.RegistryConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create authenticator")
	}

	// Create HTTP transport with optional insecure TLS
	insecure := opts.RegistryConfig.Insecure
	if opts.RegistryConfig.TLS.InsecureSkipVerify {
		insecure = true
	}
	httpTransport := createHTTPTransport(insecure)

	// Create transport option
	transportOpt := remote.WithAuth(auth)
	if insecure {
		transportOpt = remote.WithTransport(httpTransport)
	}

	return &Client{
		registry:      registry,
		registryConf:  opts.RegistryConfig,
		logger:        opts.Logger,
		authenticator: auth,
		transportOpt:  transportOpt,
	}, nil
}

// GetRegistryName returns the registry endpoint
func (c *Client) GetRegistryName() string {
	return c.registry
}

// ListRepositories lists all repositories in the registry
func (c *Client) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	// Parse registry
	reg, err := name.NewRegistry(c.registry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse registry")
	}

	// List repositories using the catalog API
	repos, err := remote.Catalog(ctx, reg, c.transportOpt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list repositories")
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

	return filtered, nil
}

// GetRepository returns a repository by name
func (c *Client) GetRepository(ctx context.Context, repoName string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Create a proper repository reference
	fullRepoName := fmt.Sprintf("%s/%s", c.registry, repoName)
	repository, err := name.NewRepository(fullRepoName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Create base repository with common functionality
	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       repoName,
		Repository: repository,
		Logger:     c.logger,
	})

	return &Repository{
		BaseRepository: baseRepo,
		client:         c,
		name:           repoName,
		repository:     repository,
	}, nil
}

// GetTransport returns an authenticated HTTP transport for the registry
func (c *Client) GetTransport(repositoryName string) (http.RoundTripper, error) {
	// Create a proper repository reference
	fullRepoName := fmt.Sprintf("%s/%s", c.registry, repositoryName)
	repository, err := name.NewRepository(fullRepoName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Create transport with authentication
	insecure := c.registryConf.Insecure || c.registryConf.TLS.InsecureSkipVerify
	httpTransport := createHTTPTransport(insecure)

	rt, err := transport.NewWithContext(
		context.Background(),
		repository.Registry,
		c.authenticator,
		httpTransport,
		[]string{repository.Scope(transport.PullScope)},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transport")
	}

	return rt, nil
}

// GetRemoteOptions returns options for the go-containerregistry remote package
func (c *Client) GetRemoteOptions() []remote.Option {
	opts := []remote.Option{
		remote.WithAuth(c.authenticator),
	}

	insecure := c.registryConf.Insecure || c.registryConf.TLS.InsecureSkipVerify
	if insecure {
		httpTransport := createHTTPTransport(true)
		opts = append(opts, remote.WithTransport(httpTransport))
	}

	return opts
}

// normalizeRegistryURL removes http/https schemes from registry URLs
func normalizeRegistryURL(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, "/")
	return url
}

// createAuthenticator creates an authenticator based on the registry configuration
func createAuthenticator(conf config.RegistryConfig) (authn.Authenticator, error) {
	authType := strings.ToLower(string(conf.Auth.Type))

	switch authType {
	case "anonymous", "":
		return authn.Anonymous, nil

	case "basic":
		username := conf.Auth.Username
		password := conf.Auth.Password

		// Expand environment variables if present
		username = expandEnvVars(username)
		password = expandEnvVars(password)

		if username == "" || password == "" {
			return nil, errors.InvalidInputf("username and password required for basic auth")
		}

		return &authn.Basic{
			Username: username,
			Password: password,
		}, nil

	case "token", "bearer":
		token := conf.Auth.Token

		// Expand environment variables if present
		token = expandEnvVars(token)

		if token == "" {
			return nil, errors.InvalidInputf("token required for token auth")
		}

		return &authn.Bearer{
			Token: token,
		}, nil

	default:
		return nil, errors.InvalidInputf("unsupported auth type: %s", conf.Auth.Type)
	}
}

// createHTTPTransport creates an HTTP transport with optional insecure TLS
func createHTTPTransport(insecureSkipVerify bool) *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if insecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return transport
}

// expandEnvVars expands environment variable references in the format ${VAR_NAME}
func expandEnvVars(s string) string {
	if !strings.Contains(s, "${") {
		return s
	}

	// Simple environment variable expansion
	result := s
	start := strings.Index(result, "${")
	for start >= 0 {
		end := strings.Index(result[start:], "}")
		if end < 0 {
			break
		}
		end += start

		// Extract variable name
		varName := result[start+2 : end]

		// Get environment variable value
		varValue := ""
		if val, exists := lookupEnv(varName); exists {
			varValue = val
		}

		// Replace ${VAR} with value
		result = result[:start] + varValue + result[end+1:]

		// Look for next variable
		start = strings.Index(result, "${")
	}

	return result
}

// lookupEnv is a wrapper around os.LookupEnv for easier testing
var lookupEnv = os.LookupEnv
