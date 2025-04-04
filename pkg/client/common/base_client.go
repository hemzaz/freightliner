package common

import (
	"context"
	"fmt"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// ClientOptions provides common options for creating registry clients
type ClientOptions struct {
	// Logger is the logger to use
	Logger *log.Logger

	// Registry is the registry hostname
	Registry string

	// Region is the AWS region/GCP location
	Region string

	// Project is the GCP project
	Project string

	// Account is the AWS account ID
	Account string

	// CustomTransport is an optional custom HTTP transport
	CustomTransport http.RoundTripper

	// AuthenticatorOverride is an optional authenticator override
	AuthenticatorOverride RegistryAuthenticator
}

// BaseClient provides common functionality for registry clients
type BaseClient struct {
	// Logger is the logger to use
	Logger *log.Logger

	// Registry is the registry hostname
	Registry string

	// Options contains client options
	Options ClientOptions

	// HTTPClient is the HTTP client to use
	HTTPClient *http.Client

	// Authenticator is the registry authenticator
	Authenticator RegistryAuthenticator

	// Auth cache for tokens
	authCache      map[string]authCacheEntry
	authCacheMutex sync.RWMutex
}

// authCacheEntry represents a cached authentication token
type authCacheEntry struct {
	Token     string
	ExpiresAt time.Time
}

// NewBaseClient creates a new base client
func NewBaseClient(options ClientOptions) *BaseClient {
	// Default options
	if options.Logger == nil {
		options.Logger = log.NewLogger(log.InfoLevel)
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Use custom transport if provided
	if options.CustomTransport != nil {
		httpClient.Transport = options.CustomTransport
	}

	return &BaseClient{
		Logger:     options.Logger,
		Registry:   options.Registry,
		Options:    options,
		HTTPClient: httpClient,
		authCache:  make(map[string]authCacheEntry),
	}
}

// SetAuthenticator sets the registry authenticator
func (c *BaseClient) SetAuthenticator(auth RegistryAuthenticator) {
	c.Authenticator = auth
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

	// This is a placeholder implementation that specific clients should override
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

	// This is a placeholder implementation that specific clients should override
	return nil, errors.NotImplementedf("method not implemented in base client")
}

// GetRegistryName returns the name of the registry
func (c *BaseClient) GetRegistryName() string {
	return c.Registry
}

// GetAuthToken gets an authentication token for the registry
func (c *BaseClient) GetAuthToken(ctx context.Context) (string, error) {
	if c.Authenticator == nil {
		return "", errors.NotImplementedf("authenticator not set")
	}

	// Check cache first
	c.authCacheMutex.RLock()
	if entry, ok := c.authCache[c.Registry]; ok {
		if time.Now().Before(entry.ExpiresAt) {
			token := entry.Token
			c.authCacheMutex.RUnlock()
			return token, nil
		}
	}
	c.authCacheMutex.RUnlock()

	// Get new token
	token, err := c.Authenticator.GetAuthToken(ctx, c.Registry)
	if err != nil {
		return "", errors.Wrap(err, "failed to get authentication token")
	}

	// Cache token with expiry (default 10 minutes)
	c.authCacheMutex.Lock()
	c.authCache[c.Registry] = authCacheEntry{
		Token:     token,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	c.authCacheMutex.Unlock()

	return token, nil
}

// CreateTransport creates a transport for registry operations
func (c *BaseClient) CreateTransport(ctx context.Context, repository string) (http.RoundTripper, error) {
	if c.Authenticator == nil {
		return nil, errors.NotImplementedf("authenticator not set")
	}

	// Get authenticator for registry
	auth, err := c.Authenticator.GetAuthenticator(ctx, c.Registry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get authenticator")
	}

	// This ensures the authn import is used
	var _ authn.Authenticator = auth

	// Create registry reference
	registry, err := name.NewRegistry(c.Registry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create registry reference")
	}

	// Create scope
	scope := fmt.Sprintf("repository:%s:pull,push", repository)

	// Create transport
	rt, err := transport.New(
		registry,
		auth,
		http.DefaultTransport,
		[]string{scope},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transport")
	}

	return rt, nil
}

// BaseRepository provides common functionality for repositories
type BaseRepository struct {
	// Client is the registry client
	Client *BaseClient

	// Name is the repository name
	Name string

	// ManifestMap is a cache of manifests
	ManifestMap map[string]*Manifest

	// ManifestMutex protects ManifestMap
	ManifestMutex sync.RWMutex
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(client *BaseClient, name string) *BaseRepository {
	return &BaseRepository{
		Client:      client,
		Name:        name,
		ManifestMap: make(map[string]*Manifest),
	}
}

// GetRepositoryName returns the name of the repository
func (r *BaseRepository) GetRepositoryName() string {
	return r.Name
}

// GetName is an alias for GetRepositoryName for backward compatibility
func (r *BaseRepository) GetName() string {
	return r.GetRepositoryName()
}

// ListTags returns all tags for the repository
func (r *BaseRepository) ListTags() ([]string, error) {
	// This is a placeholder implementation that specific clients should override
	return nil, errors.NotImplementedf("method not implemented in base repository")
}

// GetManifest returns the manifest for the given tag
func (r *BaseRepository) GetManifest(ctx context.Context, tag string) (*Manifest, error) {
	// This is a placeholder implementation that specific clients should override
	return nil, errors.NotImplementedf("method not implemented in base repository")
}

// PutManifest uploads a manifest with the given tag
func (r *BaseRepository) PutManifest(ctx context.Context, tag string, manifest *Manifest) error {
	// This is a placeholder implementation that specific clients should override
	return errors.NotImplementedf("method not implemented in base repository")
}

// DeleteManifest deletes the manifest for the given tag
func (r *BaseRepository) DeleteManifest(ctx context.Context, tag string) error {
	// This is a placeholder implementation that specific clients should override
	return errors.NotImplementedf("method not implemented in base repository")
}

// GetLayerReader returns a reader for the layer with the given digest
func (r *BaseRepository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	// This is a placeholder implementation that specific clients should override
	return nil, errors.NotImplementedf("method not implemented in base repository")
}

// GetImageReference returns a name.Reference for the given tag
func (r *BaseRepository) GetImageReference(tag string) (name.Reference, error) {
	// Create the full reference based on whether a tag is provided
	var fullRef string
	if tag != "" {
		fullRef = fmt.Sprintf("%s/%s:%s", r.Client.Registry, r.Name, tag)
		tagRef, err := name.NewTag(fullRef)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create tag reference")
		}
		return tagRef, nil
	} else {
		fullRef = fmt.Sprintf("%s/%s", r.Client.Registry, r.Name)
		refObj, err := name.ParseReference(fullRef)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create repository reference")
		}
		return refObj, nil
	}
}

// GetRemoteOptions returns options for remote operations
func (r *BaseRepository) GetRemoteOptions() ([]remote.Option, error) {
	// Create transport
	rt, err := r.Client.CreateTransport(context.Background(), r.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transport")
	}

	// Create options
	return []remote.Option{
		remote.WithTransport(rt),
	}, nil
}

// GetImage retrieves the v1.Image for the given tag
func (r *BaseRepository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	// This is a placeholder implementation that specific clients should override
	return nil, errors.NotImplementedf("method not implemented in base repository")
}

// PutImage uploads a v1.Image with the given tag
func (r *BaseRepository) PutImage(ctx context.Context, tag string, img v1.Image) error {
	// This is a placeholder implementation that specific clients should override
	return errors.NotImplementedf("method not implemented in base repository")
}
