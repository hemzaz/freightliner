package common

import (
	"context"
	"net/http"
	"sync"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// BaseClient implements common functionality for registry clients
type BaseClient struct {
	registryName string
	util         *RegistryUtil
	logger       *log.Logger

	// Cache for repositories to avoid recreating them
	repositoriesMutex sync.RWMutex
	repositories      map[string]interface{}
}

// BaseClientOptions provides options for creating a base client
type BaseClientOptions struct {
	RegistryName string
	Logger       *log.Logger
}

// NewBaseClient creates a new base client for registry operations
func NewBaseClient(opts BaseClientOptions) *BaseClient {
	if opts.Logger == nil {
		opts.Logger = log.NewLogger(log.InfoLevel)
	}

	return &BaseClient{
		registryName: opts.RegistryName,
		util:         NewRegistryUtil(opts.Logger),
		logger:       opts.Logger,
		repositories: make(map[string]interface{}),
	}
}

// GetRegistryName returns the registry endpoint
func (c *BaseClient) GetRegistryName() string {
	return c.registryName
}

// GetRepository returns a repository by name with caching
func (c *BaseClient) GetRepository(ctx context.Context, repoName string) (interface{}, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Check the cache first
	var repo interface{}
	var ok bool
	func() {
		c.repositoriesMutex.RLock()
		defer c.repositoriesMutex.RUnlock()
		repo, ok = c.repositories[repoName]
	}()

	if ok {
		return repo, nil
	}

	// Create a proper repository reference
	_, err := c.util.CreateRepositoryReference(c.registryName, repoName)
	if err != nil {
		return nil, err
	}

	// Repository creation would depend on the specific implementation
	// This is a placeholder that should be overridden by specific implementations
	return nil, errors.NotImplementedf("GetRepository must be implemented by specific registry clients")
}

// GetCachedRepository gets a repository from the cache or creates a new one
func (c *BaseClient) GetCachedRepository(ctx context.Context, repoName string, factory func(name.Repository) interface{}) (interface{}, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Check the cache first
	var repo interface{}
	var ok bool
	func() {
		c.repositoriesMutex.RLock()
		defer c.repositoriesMutex.RUnlock()
		repo, ok = c.repositories[repoName]
	}()

	if ok {
		return repo, nil
	}

	// Create a proper repository reference
	repo_ref, err := c.util.CreateRepositoryReference(c.registryName, repoName)
	if err != nil {
		return nil, err
	}

	// Create the repository using the factory function
	repo = factory(repo_ref)

	// Cache the repository
	func() {
		c.repositoriesMutex.Lock()
		defer c.repositoriesMutex.Unlock()
		c.repositories[repoName] = repo
	}()

	return repo, nil
}

// GetRemoteOptions returns common options for the remote package
func (c *BaseClient) GetRemoteOptions(transport http.RoundTripper) []remote.Option {
	return c.util.GetRemoteOptions(transport)
}

// ValidateRepositoryName checks if a repository name is valid
func (c *BaseClient) ValidateRepositoryName(repoName string) error {
	return c.util.ValidateRepositoryName(repoName)
}

// LogOperation logs a registry operation with consistent format
func (c *BaseClient) LogOperation(ctx context.Context, operation, repository string, extraFields map[string]interface{}) {
	c.util.LogRegistryOperation(ctx, operation, c.registryName, repository, extraFields)
}
