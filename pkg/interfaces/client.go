package interfaces

import (
	"context"
	"time"
)

// ===== CLIENT INTERFACES WITH SEGREGATION =====

// RepositoryLister provides repository listing capabilities
type RepositoryLister interface {
	// ListRepositories lists all repositories in a registry with the given prefix
	ListRepositories(ctx context.Context, prefix string) ([]string, error)
}

// RepositoryProvider provides repository access capabilities
type RepositoryProvider interface {
	// GetRepository returns a repository reference for the given name
	GetRepository(ctx context.Context, name string) (Repository, error)
}

// RegistryInfo provides registry information
type RegistryInfo interface {
	// GetRegistryName returns the name of the registry
	GetRegistryName() string
}

// RegistryClient defines the interface for registry clients
// NOTE: This large interface is kept for backward compatibility.
// New code should use the more focused interfaces above.
type RegistryClient interface {
	RepositoryLister
	RepositoryProvider
	RegistryInfo
}

// ===== ENHANCED CLIENT INTERFACES =====

// PaginatedRepositoryLister provides paginated repository listing
type PaginatedRepositoryLister interface {
	RepositoryLister

	// ListRepositoriesWithPagination lists repositories with pagination support
	ListRepositoriesWithPagination(ctx context.Context, prefix string, limit, offset int) (*RepositoryPage, error)

	// CountRepositories returns the total number of repositories matching the prefix
	CountRepositories(ctx context.Context, prefix string) (int, error)
}

// RepositoryPage represents a page of repositories
type RepositoryPage struct {
	Repositories []string
	TotalCount   int
	Limit        int
	Offset       int
	HasNext      bool
	HasPrevious  bool
}

// CachingRepositoryProvider provides caching for repository access
type CachingRepositoryProvider interface {
	RepositoryProvider

	// GetRepositoryWithCache returns a repository with caching support
	GetRepositoryWithCache(ctx context.Context, name string, ttl time.Duration) (Repository, error)

	// InvalidateCache invalidates the cache for a specific repository
	InvalidateCache(ctx context.Context, name string) error

	// ClearCache clears all cached repositories
	ClearCache(ctx context.Context) error

	// GetCacheStats returns cache statistics
	GetCacheStats(ctx context.Context) (*RepositoryCacheStats, error)
}

// RepositoryCacheStats provides statistics about repository cache
type RepositoryCacheStats struct {
	TotalEntries  int
	HitCount      int64
	MissCount     int64
	EvictionCount int64
	LastAccess    *time.Time
}

// BatchRepositoryProvider provides batch repository operations
type BatchRepositoryProvider interface {
	// GetRepositoriesBatch retrieves multiple repositories efficiently
	GetRepositoriesBatch(ctx context.Context, names []string) (map[string]Repository, error)

	// ExistRepositoriesBatch checks existence of multiple repositories
	ExistRepositoriesBatch(ctx context.Context, names []string) (map[string]bool, error)
}

// HealthChecker provides registry health checking capabilities
type HealthChecker interface {
	// CheckHealth performs a health check on the registry
	CheckHealth(ctx context.Context) (*RegistryHealth, error)

	// CheckConnectivity tests connectivity to the registry
	CheckConnectivity(ctx context.Context) error
}

// RegistryHealth represents the health status of a registry
type RegistryHealth struct {
	Status        HealthStatus
	ResponseTime  time.Duration
	AvailableAPIs []string
	Version       string
	LastChecked   time.Time
	ErrorMessage  string
}

// HealthStatus represents the health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ===== CLIENT COMPOSITION INTERFACES =====

// BasicClient provides basic client functionality
type BasicClient interface {
	RepositoryLister
	RepositoryProvider
	RegistryInfo
}

// EnhancedClient provides enhanced client functionality
type EnhancedClient interface {
	BasicClient
	PaginatedRepositoryLister
	BatchRepositoryProvider
}

// CachingClient provides caching client functionality
type CachingClient interface {
	EnhancedClient
	CachingRepositoryProvider
}

// FullClient provides all client functionality
type FullClient interface {
	CachingClient
	HealthChecker
}

// ClientComposer provides composition of client behaviors
type ClientComposer interface {
	// AsRepositoryLister returns a repository lister view
	AsRepositoryLister() RepositoryLister

	// AsRepositoryProvider returns a repository provider view
	AsRepositoryProvider() RepositoryProvider

	// AsRegistryInfo returns a registry info view
	AsRegistryInfo() RegistryInfo

	// AsPaginatedRepositoryLister returns a paginated repository lister view
	AsPaginatedRepositoryLister() PaginatedRepositoryLister

	// AsCachingRepositoryProvider returns a caching repository provider view
	AsCachingRepositoryProvider() CachingRepositoryProvider

	// AsBatchRepositoryProvider returns a batch repository provider view
	AsBatchRepositoryProvider() BatchRepositoryProvider

	// AsHealthChecker returns a health checker view
	AsHealthChecker() HealthChecker
}

// ===== STREAMING CLIENT INTERFACES =====

// StreamingRepositoryLister provides streaming repository listing for large registries
type StreamingRepositoryLister interface {
	// StreamRepositories provides a streaming interface for large repository lists
	StreamRepositories(ctx context.Context, prefix string) (<-chan string, <-chan error)

	// StreamRepositoriesWithFilter streams repositories with filtering
	StreamRepositoriesWithFilter(ctx context.Context, prefix string, filter RepositoryFilter) (<-chan string, <-chan error)
}

// RepositoryFilter defines filtering criteria for repository streams
type RepositoryFilter struct {
	NamePattern   string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	MinSize       int64
	MaxSize       int64
	Tags          []string
}

// ===== MULTI-REGISTRY CLIENT INTERFACES =====

// MultiRegistryClient provides access to multiple registries
type MultiRegistryClient interface {
	// GetClientForRegistry returns a client for a specific registry
	GetClientForRegistry(ctx context.Context, registry string) (RegistryClient, error)

	// RegisterRegistry registers a new registry client
	RegisterRegistry(ctx context.Context, name string, client RegistryClient) error

	// UnregisterRegistry removes a registry client
	UnregisterRegistry(ctx context.Context, name string) error

	// ListRegistries returns all registered registry names
	ListRegistries(ctx context.Context) ([]string, error)

	// GetRegistryHealth checks health of all registered registries
	GetRegistryHealth(ctx context.Context) (map[string]*RegistryHealth, error)
}

// FederatedClient provides federated access across multiple registries
type FederatedClient interface {
	MultiRegistryClient

	// SearchRepositoriesAcrossRegistries searches repositories across all registries
	SearchRepositoriesAcrossRegistries(ctx context.Context, query string) (map[string][]string, error)

	// GetRepositoryFromAnyRegistry finds a repository in any registered registry
	GetRepositoryFromAnyRegistry(ctx context.Context, name string) (Repository, string, error)
}
