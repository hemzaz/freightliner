// Package generic provides a generic Docker Registry v2 compatible client.
package generic

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"freightliner/pkg/client/common"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/helper/util"
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
	httpTransport *http.Transport // Reusable HTTP transport with connection pooling
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

	// Log security warning if insecure mode is requested
	if insecure {
		allowInsecure := os.Getenv("FREIGHTLINER_ALLOW_INSECURE_TLS")
		if allowInsecure == "true" || allowInsecure == "1" {
			opts.Logger.WithFields(map[string]interface{}{
				"registry": registry,
			}).Warn("SECURITY WARNING: Insecure TLS mode enabled - certificate verification disabled")
		} else {
			opts.Logger.WithFields(map[string]interface{}{
				"registry": registry,
			}).Info("Insecure TLS requested but blocked - set FREIGHTLINER_ALLOW_INSECURE_TLS=true to enable")
		}
	}

	// Create and store HTTP transport for connection pooling
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
		httpTransport: httpTransport, // Store for reuse in GetTransport/GetRemoteOptions
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

	// Log security warning if insecure mode is enabled
	if insecure {
		allowInsecure := os.Getenv("FREIGHTLINER_ALLOW_INSECURE_TLS")
		if allowInsecure == "true" || allowInsecure == "1" {
			c.logger.WithFields(map[string]interface{}{
				"registry":   c.registry,
				"repository": repositoryName,
			}).Warn("SECURITY WARNING: Creating insecure transport - certificate verification disabled")
		}
	}

	// Reuse stored HTTP transport for connection pooling
	rt, err := transport.NewWithContext(
		context.Background(),
		repository.Registry,
		c.authenticator,
		c.httpTransport,
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
		// Log security warning if insecure mode is enabled
		allowInsecure := os.Getenv("FREIGHTLINER_ALLOW_INSECURE_TLS")
		if allowInsecure == "true" || allowInsecure == "1" {
			c.logger.WithFields(map[string]interface{}{
				"registry": c.registry,
			}).Warn("SECURITY WARNING: Using insecure remote options - certificate verification disabled")
		}

		// Reuse stored HTTP transport for connection pooling
		opts = append(opts, remote.WithTransport(c.httpTransport))
	}

	return opts
}

// GetManifest is a convenience method to get a manifest from a repository
func (c *Client) GetManifest(ctx context.Context, repoName string, tag string) (*interfaces.Manifest, error) {
	repo, err := c.GetRepository(ctx, repoName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get repository")
	}

	return repo.GetManifest(ctx, tag)
}

// ListTags is a convenience method to list tags from a repository
func (c *Client) ListTags(ctx context.Context, repoName string) ([]string, error) {
	repo, err := c.GetRepository(ctx, repoName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get repository")
	}

	return repo.ListTags(ctx)
}

// DownloadLayer is a convenience method to download a layer from a repository
func (c *Client) DownloadLayer(ctx context.Context, repoName string, digest string) ([]byte, error) {
	repo, err := c.GetRepository(ctx, repoName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get repository")
	}

	reader, err := repo.GetLayerReader(ctx, digest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get layer reader")
	}
	defer reader.Close()

	// Use pooled buffer to read layer data (reduces GC pressure)
	// Layers can be large, use 256KB buffer
	buf := util.GetZeroCopyBuffer(256 * 1024)
	defer buf.Release()

	// Read layer into pooled buffer
	var data []byte
	for {
		n, readErr := reader.Read(buf.Bytes())
		if n > 0 {
			data = append(data, buf.Bytes()[:n]...)
		}
		if readErr != nil {
			if readErr != io.EOF {
				return nil, errors.Wrap(readErr, "failed to read layer data")
			}
			break
		}
	}

	return data, nil
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

// createHTTPTransport creates an HTTP transport with secure TLS configuration
// insecureSkipVerify should only be used for testing/development
func createHTTPTransport(insecureSkipVerify bool) *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	// Create TLS config with system cert pool
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Enforce TLS 1.2+
	}

	// Load system cert pool for certificate verification
	if certPool, err := x509.SystemCertPool(); err == nil {
		tlsConfig.RootCAs = certPool
	}

	// SECURITY WARNING: InsecureSkipVerify disables certificate validation
	// This should ONLY be used for testing/development, NEVER in production
	if insecureSkipVerify {
		// Check if explicitly allowed via environment variable
		allowInsecure := os.Getenv("FREIGHTLINER_ALLOW_INSECURE_TLS")
		if allowInsecure != "true" && allowInsecure != "1" {
			// Block insecure TLS unless explicitly enabled
			// This is a breaking change but necessary for security
			// Set FREIGHTLINER_ALLOW_INSECURE_TLS=true to allow (NOT recommended)
			tlsConfig.InsecureSkipVerify = false
			// Certificate verification will use system cert pool
		} else {
			// Insecure mode explicitly enabled - log warning
			tlsConfig.InsecureSkipVerify = true
			// Note: We can't log here as we don't have a logger,
			// but the caller should log this warning
		}
	}

	transport.TLSClientConfig = tlsConfig
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
