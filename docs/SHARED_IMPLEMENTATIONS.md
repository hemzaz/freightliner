# Shared Utilities and Base Implementations

This document explains the shared utilities and base implementations used across the Freightliner codebase.

## Registry Utilities

The `RegistryUtil` class in `pkg/client/common/registry_util.go` provides common functionality for registry operations, including:

1. Parsing registry paths (e.g., `ecr/repository` → `ecr`, `repository`)
2. Validating repository names
3. Creating repository references
4. Getting remote options for the go-containerregistry package
5. Validating registry types
6. Formatting repository URIs
7. Logging registry operations with consistent format

Example usage:

```go
// Create a registry utility
util := common.NewRegistryUtil(logger)

// Parse a registry path
registryType, repoName, err := util.ParseRegistryPath("ecr/my-repository")
if err != nil {
    return err
}

// Validate repository name
if err := util.ValidateRepositoryName(repoName); err != nil {
    return err
}

// Format repository URI
uri := util.FormatRepositoryURI("ecr", "123456789012", "us-west-2", repoName)
```

## Base Client Implementation

The `BaseClient` class in `pkg/client/common/base_client.go` provides a foundation for registry client implementations with common functionality:

1. Registry name access
2. Repository caching and retrieval
3. Remote options generation
4. Repository name validation
5. Consistent operation logging

Registry-specific clients (e.g., ECR, GCR) should extend this base client and implement registry-specific operations.

Example usage:

```go
// Create a base client
baseClient := common.NewBaseClient(common.BaseClientOptions{
    RegistryName: "registry.example.com",
    Logger:       logger,
})

// Get registry name
registryName := baseClient.GetRegistryName()

// Create a repository factory function
repoFactory := func(repo name.Repository) common.Repository {
    return NewSpecificRepository(repo)
}

// Get or create a repository with caching
repo, err := baseClient.GetCachedRepository(ctx, "my-repository", repoFactory)
if err != nil {
    return err
}

// Log an operation
baseClient.LogOperation(ctx, "ListTags", "my-repository", map[string]interface{}{
    "extra_field": "value",
})
```

## Base Repository Implementation

The `BaseRepository` class in `pkg/client/common/base_repository.go` provides a foundation for repository implementations with common functionality:

1. Repository name and URI access
2. Tag and image caching
3. Image retrieval using the go-containerregistry package
4. Tag reference creation
5. Cache management

Registry-specific repository implementations should extend this base repository and implement registry-specific operations.

Example usage:

```go
// Create a base repository
baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
    Name:       "my-repository",
    Repository: repo,
    Logger:     logger,
})

// Get repository name and URI
name := baseRepo.GetName()
uri := baseRepo.GetURI()

// Create a tag reference
tagRef, err := baseRepo.CreateTagReference("latest")
if err != nil {
    return err
}

// Get an image with remote options
img, err := baseRepo.GetRemoteImage(ctx, tagRef, remoteOptions...)
if err != nil {
    return err
}

// Cache the image for future use
baseRepo.CacheImage("latest", img)
```

## Worker Pool Implementation

The `WorkerPool` class in `pkg/replication/worker_pool.go` provides a unified worker pool implementation for parallel processing:

1. Configurable number of workers
2. Job submission with optional priority
3. Context-aware job processing
4. Result collection
5. Graceful shutdown

Example usage:

```go
// Create a worker pool with 5 workers
pool := replication.NewWorkerPool(5, logger)

// Start the pool
pool.Start()

// Submit a job
err := pool.Submit("job-1", func(ctx context.Context) error {
    // Perform work
    return nil
})

// Submit a job with context
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
err = pool.SubmitWithContext(ctx, "job-2", func(ctx context.Context) error {
    // Perform work with context awareness
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue working
    }
    return nil
})

// Collect results
go func() {
    for result := range pool.GetResults() {
        if result.Error != nil {
            log.Printf("Job %s failed: %v", result.JobID, result.Error)
        } else {
            log.Printf("Job %s completed successfully", result.JobID)
        }
    }
}()

// Wait for all jobs to complete
pool.Wait()

// Or stop the pool immediately
// pool.Stop()
```

## Extending the Base Implementations

When creating registry-specific implementations, extend the base classes and implement the required methods:

### For Registry Clients

```go
type ECRClient struct {
    *common.BaseClient
    // ECR-specific fields
}

func NewECRClient(opts ECRClientOptions) (*ECRClient, error) {
    baseClient := common.NewBaseClient(common.BaseClientOptions{
        RegistryName: formatECRRegistryName(opts.AccountID, opts.Region),
        Logger:       opts.Logger,
    })
    
    return &ECRClient{
        BaseClient: baseClient,
        // Initialize ECR-specific fields
    }, nil
}

// Implement or override required methods
func (c *ECRClient) GetRepository(ctx context.Context, repoName string) (common.Repository, error) {
    // Use base client's caching functionality
    return c.GetCachedRepository(ctx, repoName, func(repo name.Repository) common.Repository {
        return NewECRRepository(c, repo, repoName)
    })
}
```

### For Repositories

```go
type ECRRepository struct {
    *common.BaseRepository
    client *ECRClient
    // ECR-specific fields
}

func NewECRRepository(client *ECRClient, repo name.Repository, name string) *ECRRepository {
    baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
        Name:       name,
        Repository: repo,
        Logger:     client.Logger,
    })
    
    return &ECRRepository{
        BaseRepository: baseRepo,
        client:         client,
        // Initialize ECR-specific fields
    }
}

// Implement or override required methods
func (r *ECRRepository) ListTags(ctx context.Context) ([]string, error) {
    // ECR-specific implementation
}
```

## Benefits of Shared Implementations

1. **Reduced Duplication**: Common code is centralized, reducing duplication and maintenance overhead
2. **Consistent Behavior**: All clients and repositories behave consistently for common operations
3. **Easier Maintenance**: Changes to common functionality only need to be made in one place
4. **Simplified Implementation**: New registry types can focus on registry-specific functionality
5. **Better Testing**: Base functionality can be tested independently of specific implementations

## Enhanced Implementations

For more advanced functionality, see the [Enhanced Implementations](ENHANCED_IMPLEMENTATIONS.md) documentation, which covers:

1. **Enhanced Client**: Advanced client with authentication, retry logic, and transport customization
2. **Enhanced Repository**: Repository with image summary, comparison, and export capabilities
3. **Base Authenticator**: Foundation for registry authentication implementations
4. **Base Transport**: HTTP transport with logging, retries, and timeout functionality
