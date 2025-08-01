package service

import (
	"context"
	"time"

	"freightliner/pkg/interfaces"
)

// Import types from the shared interfaces package for compatibility
type (
	// Legacy interfaces (kept for backward compatibility)
	Repository            = interfaces.Repository
	RegistryClient        = interfaces.RegistryClient
	Manifest              = interfaces.Manifest
	InterfaceRepoInfo     = interfaces.RepositoryInfo // Renamed to avoid conflict
	TagLister             = interfaces.TagLister
	ManifestAccessor      = interfaces.ManifestAccessor
	ManifestManager       = interfaces.ManifestManager
	LayerAccessor         = interfaces.LayerAccessor
	RemoteImageAccessor   = interfaces.RemoteImageAccessor
	RegistryAuthenticator = interfaces.RegistryAuthenticator

	// Segregated interfaces (preferred for new code)
	Reader           = interfaces.Reader
	Writer           = interfaces.Writer
	ImageProvider    = interfaces.ImageProvider
	MetadataProvider = interfaces.MetadataProvider
	ContentProvider  = interfaces.ContentProvider
	ContentManager   = interfaces.ContentManager

	// Client interfaces
	RepositoryLister          = interfaces.RepositoryLister
	RepositoryProvider        = interfaces.RepositoryProvider
	RegistryInfo              = interfaces.RegistryInfo
	PaginatedRepositoryLister = interfaces.PaginatedRepositoryLister
	CachingRepositoryProvider = interfaces.CachingRepositoryProvider
	BatchRepositoryProvider   = interfaces.BatchRepositoryProvider
	HealthChecker             = interfaces.HealthChecker

	// Auth interfaces
	TokenProvider         = interfaces.TokenProvider
	HeaderProvider        = interfaces.HeaderProvider
	AuthenticatorProvider = interfaces.AuthenticatorProvider
	TokenManager          = interfaces.TokenManager
	CachingAuthenticator  = interfaces.CachingAuthenticator
)

// ===== SERVICE-SPECIFIC INTERFACES =====

// RepositoryCreator is an interface for client types that can create repositories
type RepositoryCreator interface {
	// CreateRepository creates a new repository with the given name and tags
	CreateRepository(ctx context.Context, name string, tags map[string]string) (Repository, error)
}

// ReplicationService provides image replication capabilities
type ReplicationService interface {
	// ReplicateRepository replicates a repository from source to destination
	ReplicateRepository(ctx context.Context, source, destination string) (*ReplicationResult, error)

	// ReplicateImage replicates a single image between registries
	ReplicateImage(ctx context.Context, request *ReplicationRequest) (*ReplicationResult, error)

	// ReplicateImagesBatch replicates multiple images in a batch
	ReplicateImagesBatch(ctx context.Context, requests []*ReplicationRequest) ([]*ReplicationResult, error)

	// StreamReplication provides streaming replication for large operations
	StreamReplication(ctx context.Context, requests <-chan *ReplicationRequest) (<-chan *ReplicationResult, <-chan error)
}

// ReplicationRequest represents a replication request
type ReplicationRequest struct {
	SourceRegistry        string
	SourceRepository      string
	SourceTags            []string
	DestinationRegistry   string
	DestinationRepository string
	DestinationTags       []string
	Options               *ReplicationOptions
	Priority              int
}

// ReplicationOptions provides options for replication
type ReplicationOptions struct {
	DryRun           bool
	ForceOverwrite   bool
	IncludeManifests bool
	IncludeLayers    bool
	ParallelCopies   int
	RetryAttempts    int
	RetryDelay       time.Duration
	ProgressCallback func(progress *ReplicationProgress)
}

// ReplicationResult represents the result of a replication
type ReplicationResult struct {
	Request      *ReplicationRequest
	Success      bool
	Error        error
	Duration     time.Duration
	BytesCopied  int64
	LayersCopied int
	StartTime    time.Time
	EndTime      time.Time
}

// ReplicationProgress represents replication progress
type ReplicationProgress struct {
	Request          *ReplicationRequest
	Stage            string
	Completed        int
	Total            int
	BytesTransferred int64
	TotalBytes       int64
	CurrentImage     string
}

// ===== MONITORING AND OBSERVABILITY =====

// MonitoringService provides monitoring capabilities for services
type MonitoringService interface {
	// RecordMetric records a metric value
	RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) error

	// IncrementCounter increments a counter metric
	IncrementCounter(ctx context.Context, name string, tags map[string]string) error

	// RecordHistogram records a histogram value
	RecordHistogram(ctx context.Context, name string, value float64, tags map[string]string) error

	// SetGauge sets a gauge metric value
	SetGauge(ctx context.Context, name string, value float64, tags map[string]string) error
}

// HealthService provides health checking capabilities
type HealthService interface {
	// CheckHealth performs a comprehensive health check
	CheckHealth(ctx context.Context) (*ServiceHealth, error)

	// CheckDependencies checks the health of service dependencies
	CheckDependencies(ctx context.Context) (map[string]*DependencyHealth, error)

	// GetServiceInfo returns service information
	GetServiceInfo(ctx context.Context) (*ServiceInfo, error)
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Status       HealthStatus
	Version      string
	Uptime       time.Duration
	Dependencies map[string]*DependencyHealth
	Checks       map[string]*HealthCheck
	Timestamp    time.Time
}

// DependencyHealth represents the health of a dependency
type DependencyHealth struct {
	Name         string
	Status       HealthStatus
	ResponseTime time.Duration
	ErrorMessage string
	LastChecked  time.Time
}

// HealthCheck represents a single health check
type HealthCheck struct {
	Name      string
	Status    HealthStatus
	Message   string
	Duration  time.Duration
	Timestamp time.Time
}

// ServiceInfo provides information about the service
type ServiceInfo struct {
	Name        string
	Version     string
	BuildTime   time.Time
	GitCommit   string
	Environment string
	Features    []string
}

// HealthStatus represents health status
type HealthStatus string

const (
	HealthyStatus   HealthStatus = "healthy"
	DegradedStatus  HealthStatus = "degraded"
	UnhealthyStatus HealthStatus = "unhealthy"
	UnknownStatus   HealthStatus = "unknown"
)

// ===== COMPOSITION INTERFACES FOR SERVICES =====

// BasicService provides basic service functionality
type BasicService interface {
	Reader
	Writer
	HealthService
}

// EnhancedService provides enhanced service functionality
type EnhancedService interface {
	BasicService
	ReplicationService
	MonitoringService
}

// FullService provides all service functionality
type FullService interface {
	EnhancedService
	// Additional service-specific interfaces can be added here
}

// ServiceComposer provides composition of service behaviors
type ServiceComposer interface {
	// AsReader returns a reader view
	AsReader() Reader

	// AsWriter returns a writer view
	AsWriter() Writer

	// AsReplicationService returns a replication service view
	AsReplicationService() ReplicationService

	// AsMonitoringService returns a monitoring service view
	AsMonitoringService() MonitoringService

	// AsHealthService returns a health service view
	AsHealthService() HealthService
}
