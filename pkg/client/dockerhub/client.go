// Package dockerhub provides Docker Hub registry client functionality.
package dockerhub

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

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
	// DockerHubRegistry is the Docker Hub registry endpoint
	DockerHubRegistry = "registry-1.docker.io"

	// DockerHubIndexServer is the Docker Hub index server
	DockerHubIndexServer = "index.docker.io"

	// DefaultDockerHubLibrary is the library namespace for official images
	DefaultDockerHubLibrary = "library"

	// MaxRetries is the maximum number of retry attempts
	MaxRetries = 3

	// BaseRetryDelay is the base delay for exponential backoff
	BaseRetryDelay = 1 * time.Second

	// RateLimitHeader is the header for rate limit information
	RateLimitHeader = "RateLimit-Remaining"
)

// Client implements the registry client interface for Docker Hub
type Client struct {
	registry      string
	logger        log.Logger
	authenticator authn.Authenticator
	httpClient    *http.Client
	retryConfig   RetryConfig
}

// ClientOptions provides configuration for connecting to Docker Hub
type ClientOptions struct {
	// Username for Docker Hub authentication (optional for public images)
	Username string

	// Password for Docker Hub authentication (optional for public images)
	Password string

	// RegistryConfig contains the registry configuration
	RegistryConfig config.RegistryConfig

	// Logger is the logger to use
	Logger log.Logger

	// HTTPClient is an optional custom HTTP client
	HTTPClient *http.Client

	// RetryConfig configures retry behavior
	RetryConfig *RetryConfig
}

// RetryConfig configures retry behavior for rate limiting
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// BaseDelay is the base delay for exponential backoff
	BaseDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// EnableExponentialBackoff enables exponential backoff
	EnableExponentialBackoff bool
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:               MaxRetries,
		BaseDelay:                BaseRetryDelay,
		MaxDelay:                 30 * time.Second,
		EnableExponentialBackoff: true,
	}
}

// NewClient creates a new Docker Hub client
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Create authenticator based on credentials
	var auth authn.Authenticator
	if opts.Username != "" && opts.Password != "" {
		auth = &authn.Basic{
			Username: opts.Username,
			Password: opts.Password,
		}
		opts.Logger.Info("Using authenticated Docker Hub access")
	} else if opts.RegistryConfig.Auth.Type == config.AuthTypeBasic {
		auth = &authn.Basic{
			Username: opts.RegistryConfig.Auth.Username,
			Password: opts.RegistryConfig.Auth.Password,
		}
		opts.Logger.Info("Using authenticated Docker Hub access from config")
	} else {
		auth = authn.Anonymous
		opts.Logger.Warn("Using anonymous Docker Hub access (strict rate limits apply)")
	}

	// Use custom HTTP client or default
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &rateLimitTransport{
				base:   http.DefaultTransport,
				logger: opts.Logger,
			},
		}
	}

	// Set retry config
	retryConfig := DefaultRetryConfig()
	if opts.RetryConfig != nil {
		retryConfig = *opts.RetryConfig
	}

	return &Client{
		registry:      DockerHubRegistry,
		logger:        opts.Logger,
		authenticator: auth,
		httpClient:    httpClient,
		retryConfig:   retryConfig,
	}, nil
}

// GetRegistryName returns the Docker Hub registry endpoint
func (c *Client) GetRegistryName() string {
	return c.registry
}

// ListRepositories lists all repositories accessible to the authenticated user
// Note: Docker Hub API has limitations and rate limits
func (c *Client) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	// Docker Hub's catalog API is limited and requires authentication
	// For now, we return an error indicating this limitation
	c.logger.Warn("Docker Hub catalog API is limited, consider using GetRepository for known repositories")
	return nil, errors.NotImplementedf("Docker Hub catalog API has significant limitations, use GetRepository instead")
}

// GetRepository returns a repository by name
func (c *Client) GetRepository(ctx context.Context, repoName string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Normalize repository name for Docker Hub
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
	}).Debug("Created Docker Hub repository reference")

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

// GetTransport returns an authenticated HTTP transport for Docker Hub
func (c *Client) GetTransport(repositoryName string) (http.RoundTripper, error) {
	// Normalize repository name
	normalizedName := c.normalizeRepositoryName(repositoryName)

	// Create repository reference
	fullRepoName := fmt.Sprintf("%s/%s", c.registry, normalizedName)
	repository, err := name.NewRepository(fullRepoName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Create transport with authentication and rate limiting
	rt, err := transport.NewWithContext(
		context.Background(),
		repository.Registry,
		c.authenticator,
		c.httpClient.Transport,
		[]string{repository.Scope(transport.PullScope)},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Docker Hub transport")
	}

	return rt, nil
}

// GetRemoteOptions returns options for the go-containerregistry remote package
func (c *Client) GetRemoteOptions() []remote.Option {
	return []remote.Option{
		remote.WithAuth(c.authenticator),
		remote.WithTransport(c.httpClient.Transport),
	}
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

	// Read all data from the layer
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read layer data")
	}

	return data, nil
}

// normalizeRepositoryName normalizes Docker Hub repository names
// Docker Hub official images are in the "library" namespace
func (c *Client) normalizeRepositoryName(repoName string) string {
	// Remove leading slashes
	repoName = strings.TrimPrefix(repoName, "/")

	// If there's no slash, it's an official image and should be prefixed with "library/"
	if !strings.Contains(repoName, "/") {
		return fmt.Sprintf("%s/%s", DefaultDockerHubLibrary, repoName)
	}

	// Handle docker.io prefix
	repoName = strings.TrimPrefix(repoName, "docker.io/")
	repoName = strings.TrimPrefix(repoName, DockerHubRegistry+"/")

	return repoName
}

// executeWithRetry executes a function with exponential backoff retry logic
func (c *Client) executeWithRetry(ctx context.Context, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Check context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Execute the operation
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if !c.shouldRetry(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt < c.retryConfig.MaxRetries {
			delay := c.calculateBackoff(attempt)
			c.logger.WithFields(map[string]interface{}{
				"operation":  operation,
				"attempt":    attempt + 1,
				"maxRetries": c.retryConfig.MaxRetries,
				"delay":      delay,
				"error":      err.Error(),
			}).Warn("Operation failed, retrying with backoff")

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	return errors.Wrap(lastErr, "operation failed after retries")
}

// shouldRetry determines if an error is retryable
func (c *Client) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Retry on rate limit errors
	if strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit") {
		return true
	}

	// Retry on temporary network errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection refused") {
		return true
	}

	// Retry on 5xx server errors
	if strings.Contains(errStr, "500") || strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") || strings.Contains(errStr, "504") {
		return true
	}

	return false
}

// calculateBackoff calculates the backoff delay for a given attempt
func (c *Client) calculateBackoff(attempt int) time.Duration {
	if !c.retryConfig.EnableExponentialBackoff {
		return c.retryConfig.BaseDelay
	}

	// Exponential backoff: baseDelay * 2^attempt
	delay := time.Duration(float64(c.retryConfig.BaseDelay) * math.Pow(2, float64(attempt)))

	// Cap at max delay
	if delay > c.retryConfig.MaxDelay {
		delay = c.retryConfig.MaxDelay
	}

	return delay
}

// rateLimitTransport wraps an HTTP transport to add rate limit awareness
type rateLimitTransport struct {
	base   http.RoundTripper
	logger log.Logger
}

// RoundTrip implements http.RoundTripper with rate limit logging
func (t *rateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Log rate limit information if available
	if remaining := resp.Header.Get(RateLimitHeader); remaining != "" {
		t.logger.WithFields(map[string]interface{}{
			"rateLimit": remaining,
			"endpoint":  req.URL.String(),
		}).Debug("Docker Hub rate limit status")
	}

	// Check for rate limit exceeded
	if resp.StatusCode == http.StatusTooManyRequests {
		t.logger.Warn("Docker Hub rate limit exceeded, request will be retried")
	}

	return resp, nil
}
