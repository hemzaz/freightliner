# Enhanced Client and Repository Implementations

This document outlines the enhanced client and repository implementations that build upon the base implementations in Freightliner.

## Enhanced Client

The `EnhancedClient` class in `pkg/client/common/enhanced_client.go` extends the base client with additional functionality:

1. **Authentication Management**: Integrated authenticator handling with caching
2. **Transport Configuration**: Customizable HTTP transport settings
3. **Retry Policy**: Configurable retry logic for failed requests
4. **Logging**: Built-in request/response logging
5. **Timeouts**: Configurable timeouts for all operations
6. **TLS Configuration**: Supports insecure TLS (for testing) or strict TLS

### Usage Example

```go
// Create an enhanced client with custom options
client := common.NewEnhancedClient(common.EnhancedClientOptions{
    RegistryName:         "registry.example.com",
    Logger:               logger,
    Authenticator:        authenticator,
    EnableLogging:        true,
    EnableRetries:        true,
    MaxRetries:           5,
    RequestTimeout:       30 * time.Second,
    InsecureSkipTLSVerify: false,
})

// Get enhanced remote options for go-containerregistry operations
options, err := client.GetEnhancedRemoteOptions(ctx, "my-repository")
if err != nil {
    return err
}

// Use the custom retry policy
client.SetRetryPolicy(func(resp *http.Response, err error) bool {
    // Custom retry logic
    return err != nil || (resp != nil && resp.StatusCode >= 500)
})

// Add transport option
client.AddTransportOption(func(t *http.Transport) {
    t.MaxIdleConns = 100
    t.MaxIdleConnsPerHost = 10
})
```

### Key Features

#### Transport Management

The enhanced client manages HTTP transports intelligently:

1. Transports are created on-demand and cached for reuse
2. Each repository gets its own optimized transport
3. Transports can be customized with options
4. The cache can be cleared to force recreation of transports

#### Authentication Integration

The enhanced client integrates with authenticators:

1. Authentication is applied to all registry operations
2. Authenticator tokens are cached when possible
3. Supports automatic token refresh

#### Configuration Options

The enhanced client supports numerous configuration options:

- Authentication settings
- HTTP transport settings
- Retry and logging settings
- Timeout settings for different operation types

## Enhanced Repository

The `EnhancedRepository` class in `pkg/client/common/enhanced_repository.go` extends the base repository with additional functionality:

1. **Image Summaries**: Quick access to image metadata without downloading the full image
2. **Tag Comparison**: Compare two tags to find differences
3. **Image Export**: Export image data to various formats
4. **Tag Copying**: Copy tags between repositories
5. **Repository Analysis**: Analyze repository contents
6. **Caching Management**: Enhanced caching for repository operations

### Usage Example

```go
// Create an enhanced repository
repo := common.NewEnhancedRepository(common.EnhancedRepositoryOptions{
    Name:              "my-repository",
    Repository:        repoRef,
    Logger:            logger,
    Client:            client,
    CacheExpiration:   10 * time.Minute,
    EnableSummaryCache: true,
})

// Get a summary of an image without downloading the whole image
summary, err := repo.GetImageSummary(ctx, "latest")
if err != nil {
    return err
}

fmt.Printf("Image digest: %s\n", summary.Digest)
fmt.Printf("Image size: %d bytes\n", summary.Size)
fmt.Printf("Layers: %d\n", summary.Layers)

// Compare two tags
diff, err := repo.CompareTags(ctx, "v1", "v2")
if err != nil {
    return err
}

if diff["identical"].(bool) {
    fmt.Println("Tags are identical")
} else {
    fmt.Printf("Differences found. Size difference: %d bytes\n", 
        diff["size2"].(int64) - diff["size1"].(int64))
}

// Export image data
file, err := os.Create("image-data.json")
if err != nil {
    return err
}
defer file.Close()

err = repo.ExportImage(ctx, "latest", "json", file)
if err != nil {
    return err
}
```

### Key Features

#### Image Summaries

The enhanced repository can provide quick summaries of images:

1. Digest, size, and layer count
2. Media type and creation time
3. Summaries are cached for quick access

#### Tag Comparison

The repository can compare two tags to find differences:

1. Compare digests to quickly identify identical images
2. Compare layer counts and sizes
3. Compare layer digests to identify specific changes
4. Compare configurations to identify build differences

#### Image Export

The repository can export image data in various formats:

1. JSON export of manifest and configuration
2. Additional formats can be added as needed

#### Tag Operations

The repository supports enhanced tag operations:

1. Copy tags between repositories
2. Bulk tag management
3. Tag refreshing

## Base Authenticator

The `BaseAuthenticator` class in `pkg/client/common/base_authenticator.go` provides a foundation for authentication implementations:

1. **Credential Caching**: Cache authentication tokens to reduce authentication requests
2. **Expiry Management**: Track token expiry for automatic renewal
3. **Transport Integration**: Easily create authenticated transports

### Usage Example

```go
// Create a custom authenticator extending the base authenticator
type CustomAuth struct {
    *common.BaseAuthenticator
    // Custom fields
}

// Implement the Authorization method
func (a *CustomAuth) Authorization(ctx context.Context, resource authn.Resource) (*authn.AuthConfig, error) {
    // Check if we have cached credentials
    if auth, valid := a.GetCachedAuth(time.Now().Unix()); valid {
        return auth.Authorization(ctx, resource)
    }
    
    // Obtain new credentials
    // ...
    
    // Cache the new credentials
    a.SetCachedAuth(newAuth, expiryTime)
    
    return newAuth.Authorization(ctx, resource)
}

// Create an HTTP transport with authentication
transport := common.TransportWithAuth(http.DefaultTransport, authenticator, resource)
```

## Base Transport

The `BaseTransport` class in `pkg/client/common/base_transport.go` provides enhanced HTTP transport functionality:

1. **Default Configuration**: Reasonable defaults for timeouts and connections
2. **Request Logging**: Log HTTP requests and responses
3. **Automatic Retries**: Retry failed requests with configurable policy
4. **Request Timeouts**: Add timeouts to requests

### Usage Example

```go
// Create a base transport
transport := common.NewBaseTransport(logger)

// Create a default transport
defaultTransport := transport.CreateDefaultTransport()

// Create logging transport
loggingTransport := transport.LoggingTransport(defaultTransport)

// Create retry transport
retryTransport := transport.RetryTransport(loggingTransport, 3, func(resp *http.Response, err error) bool {
    return err != nil || (resp != nil && resp.StatusCode >= 500)
})

// Create timeout transport
timeoutTransport := transport.TimeoutTransport(retryTransport, 30*time.Second)

// Use the final transport
client := &http.Client{
    Transport: timeoutTransport,
}
```

## Benefits of Enhanced Implementations

1. **Advanced Features**: More capabilities for client code without reimplementing
2. **Performance Optimizations**: Caching and transport optimization
3. **Developer Experience**: Simplified API for common operations
4. **Flexibility**: Customizable behavior without modifying base code
5. **Robustness**: Retry logic, timeouts, and error handling
6. **Maintainability**: Centralized implementation of complex functionality
