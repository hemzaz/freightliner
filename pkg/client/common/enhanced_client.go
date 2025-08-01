package common

import (
	"context"
	"net/http"
	"sync"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// EnhancedClientOptions provides options for creating an enhanced client
type EnhancedClientOptions struct {
	// Basic options
	RegistryName string
	Logger       log.Logger

	// Authentication options
	Authenticator            authn.Authenticator
	InsecureSkipTLSVerify    bool
	CredentialRefreshTimeout time.Duration

	// HTTP options
	Transport             http.RoundTripper
	EnableLogging         bool
	EnableRetries         bool
	MaxRetries            int
	RequestTimeout        time.Duration
	ConnectionIdleTimeout time.Duration
	ConnectionKeepAlive   time.Duration
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
	ExpectContinueTimeout time.Duration
}

// EnhancedClient extends the base client with additional functionality
type EnhancedClient struct {
	*BaseClient

	// Authentication
	authenticator   authn.Authenticator
	authRefreshTime time.Duration

	// HTTP transport
	baseTransport  *BaseTransport
	defaultOptions []remote.Option
	plainHTTP      bool

	// Configuration
	options EnhancedClientOptions

	// Advanced configuration
	retryPolicy   func(*http.Response, error) bool
	transportOpts []TransportOption

	// Cache
	transportCache      map[string]http.RoundTripper
	transportCacheMutex sync.RWMutex
}

// TransportOption is a function that configures an HTTP transport
type TransportOption func(*http.Transport)

// NewEnhancedClient creates a new enhanced client
func NewEnhancedClient(opts EnhancedClientOptions) *EnhancedClient {
	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Create base client
	baseClient := NewBaseClient(BaseClientOptions{
		RegistryName: opts.RegistryName,
		Logger:       opts.Logger,
	})

	// Create base transport
	baseTransport := NewBaseTransport(opts.Logger)

	// Set default options
	if opts.MaxRetries == 0 {
		opts.MaxRetries = 3
	}

	if opts.RequestTimeout == 0 {
		opts.RequestTimeout = 30 * time.Second
	}

	if opts.CredentialRefreshTimeout == 0 {
		opts.CredentialRefreshTimeout = 15 * time.Minute
	}

	// Create enhanced client
	client := &EnhancedClient{
		BaseClient:      baseClient,
		authenticator:   opts.Authenticator,
		authRefreshTime: opts.CredentialRefreshTimeout,
		baseTransport:   baseTransport,
		options:         opts,
		transportCache:  make(map[string]http.RoundTripper),
	}

	// Default retry policy
	client.retryPolicy = func(resp *http.Response, err error) bool {
		if err != nil {
			// Retry network errors
			return true
		}

		if resp != nil {
			// Retry 5xx server errors
			return resp.StatusCode >= 500
		}

		return false
	}

	return client
}

// GetAuthenticator returns the client's authenticator
func (c *EnhancedClient) GetAuthenticator() authn.Authenticator {
	return c.authenticator
}

// SetAuthenticator sets the client's authenticator
func (c *EnhancedClient) SetAuthenticator(auth authn.Authenticator) {
	c.authenticator = auth
}

// GetTransport gets a transport for a repository with the client's configuration
func (c *EnhancedClient) GetTransport(repositoryName string) (http.RoundTripper, error) {
	// Try to get from cache first
	var transport http.RoundTripper
	var ok bool
	func() {
		c.transportCacheMutex.RLock()
		defer c.transportCacheMutex.RUnlock()
		transport, ok = c.transportCache[repositoryName]
	}()

	if ok {
		return transport, nil
	}

	// Create a proper repository reference
	registry := c.GetRegistryName()
	repository, err := c.util.CreateRepositoryReference(registry, repositoryName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Create a base transport
	baseTransport := c.baseTransport.CreateDefaultTransport()

	// Apply transport options
	for _, opt := range c.transportOpts {
		opt(baseTransport)
	}

	// Create authentication if needed
	var authTransport http.RoundTripper = baseTransport
	if c.authenticator != nil {
		authTransport = TransportWithAuth(baseTransport, c.authenticator, repository)
	}

	// Add logging if enabled
	loggingTransport := authTransport
	if c.options.EnableLogging {
		loggingTransport = c.baseTransport.LoggingTransport(authTransport)
	}

	// Add retries if enabled
	retryTransport := loggingTransport
	if c.options.EnableRetries {
		retryTransport = c.baseTransport.RetryTransport(loggingTransport, c.options.MaxRetries, c.retryPolicy)
	}

	// Add timeout if specified
	timeoutTransport := retryTransport
	if c.options.RequestTimeout > 0 {
		timeoutTransport = c.baseTransport.TimeoutTransport(retryTransport, c.options.RequestTimeout)
	}

	// Store in cache
	func() {
		c.transportCacheMutex.Lock()
		defer c.transportCacheMutex.Unlock()
		c.transportCache[repositoryName] = timeoutTransport
	}()

	return timeoutTransport, nil
}

// ClearTransportCache clears the transport cache
func (c *EnhancedClient) ClearTransportCache() {
	func() {
		c.transportCacheMutex.Lock()
		defer c.transportCacheMutex.Unlock()
		c.transportCache = make(map[string]http.RoundTripper)
	}()
}

// SetRetryPolicy sets a custom retry policy
func (c *EnhancedClient) SetRetryPolicy(policy func(*http.Response, error) bool) {
	c.retryPolicy = policy
	// Clear cache to recreate transports with new policy
	c.ClearTransportCache()
}

// AddTransportOption adds a transport option
func (c *EnhancedClient) AddTransportOption(opt TransportOption) {
	c.transportOpts = append(c.transportOpts, opt)
	// Clear cache to recreate transports with new options
	c.ClearTransportCache()
}

// GetEnhancedRemoteOptions returns options for the go-containerregistry remote package
func (c *EnhancedClient) GetEnhancedRemoteOptions(ctx context.Context, repoName string) ([]remote.Option, error) {
	var options []remote.Option

	// Get transport
	transport, err := c.GetTransport(repoName)
	if err != nil {
		return nil, err
	}

	// Add transport option
	options = append(options, remote.WithTransport(transport))

	// Add context
	options = append(options, remote.WithContext(ctx))

	// Add other options from configuration
	if c.options.InsecureSkipTLSVerify {
		// WithInsecure is defined in the remote package, but we'll reference it indirectly
		// This comment acknowledges that we need to handle this separately
	}

	return options, nil
}
