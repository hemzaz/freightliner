# Code Reuse Patterns

This document outlines the code reuse patterns used in the Freightliner project. It provides guidelines and examples for how to effectively reuse code across the codebase.

## Table of Contents

1. [Introduction](#introduction)
2. [Composition over Inheritance](#composition-over-inheritance)
3. [Interface-Based Design](#interface-based-design)
4. [Utility Functions](#utility-functions)
5. [Middleware Pattern](#middleware-pattern)
6. [Options Pattern](#options-pattern)
7. [Factory Functions](#factory-functions)
8. [Caching Strategies](#caching-strategies)
9. [Code Generation](#code-generation)
10. [Best Practices](#best-practices)

## Introduction

Code reuse is a fundamental software engineering practice that helps reduce duplication, improve maintainability, and ensure consistency across a codebase. The Freightliner project employs several patterns for effective code reuse, each with its own use cases and benefits.

## Composition over Inheritance

Freightliner favors composition over inheritance for code reuse. This pattern involves embedding structs rather than creating deep inheritance hierarchies.

### Pattern

```go
// Base struct with common functionality
type BaseClient struct {
    // Common fields
    logger *log.Logger
}

// Specific implementation using composition
type ECRClient struct {
    *BaseClient            // Embed the base client
    ecrClient *ecr.Service // ECR-specific field
}
```

### When to Use

- When shared functionality is needed across multiple types
- When behavior needs to be extended without modifying existing code
- When a type needs functionality from multiple sources

### Examples in Freightliner

- `pkg/client/ecr/client.go`: ECR client embeds the base client
- `pkg/client/gcr/client.go`: GCR client embeds the base client
- `pkg/client/common/enhanced_client.go`: Enhanced client embeds the base client

### Benefits

- Simplifies type hierarchies
- Provides flexibility in combining behaviors
- Avoids the "diamond problem" of multiple inheritance
- Makes dependencies explicit

## Interface-Based Design

Freightliner uses interfaces extensively to define contracts and enable pluggable implementations.

### Pattern

```go
// Define interface in the package that uses it
type Repository interface {
    ListTags(ctx context.Context) ([]string, error)
    GetTag(ctx context.Context, tag string) (v1.Image, error)
    // Other methods...
}

// Implement interface in specific packages
type ECRRepository struct {
    // Implementation fields...
}

// Implement interface methods
func (r *ECRRepository) ListTags(ctx context.Context) ([]string, error) {
    // ECR-specific implementation...
}
```

### When to Use

- When multiple implementations of the same functionality are needed
- When testing with mocks is required
- When dependencies need to be decoupled
- When functionality may change or be extended in the future

### Examples in Freightliner

- `pkg/client/common`: Repository and Client interfaces
- `pkg/service`: Service interfaces
- `pkg/copy`: Copier interfaces

### Benefits

- Enables dependency injection for testing
- Provides a clear contract for implementations
- Decouples consumers from concrete implementations
- Allows for easy swapping of implementations

## Utility Functions

Freightliner provides utility packages with reusable functions for common operations.

### Pattern

```go
// Define utility functions in a dedicated package
package util

// ParseRegistryPath parses a registry path into components
func ParseRegistryPath(path string) (registry, repo string, err error) {
    parts := strings.SplitN(path, "/", 2)
    if len(parts) != 2 {
        return "", "", fmt.Errorf("invalid registry path: %s", path)
    }
    return parts[0], parts[1], nil
}
```

### When to Use

- For stateless operations that are used across multiple packages
- For commonly repeated algorithms or operations
- For format conversions, parsing, and validation

### Examples in Freightliner

- `pkg/helper/util`: Common utility functions
- `pkg/client/common/registry_util.go`: Registry-specific utilities
- `pkg/helper/errors`: Error handling utilities

### Benefits

- Centralizes common operations
- Ensures consistent behavior across the codebase
- Simplifies testing of common functionality
- Reduces code duplication

## Middleware Pattern

Freightliner uses the middleware pattern for HTTP transport customization, similar to how web frameworks use middleware.

### Pattern

```go
// Create a middleware that wraps a handler
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Pre-processing
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        
        // Call the next handler
        next.ServeHTTP(w, r)
        
        // Post-processing
        log.Printf("Completed request: %s %s", r.Method, r.URL.Path)
    })
}

// Apply middleware
handler = LoggingMiddleware(handler)
```

### When to Use

- For cross-cutting concerns like logging, authentication, and metrics
- For operations that need to wrap around core functionality
- For adding behavior to HTTP handlers or transports

### Examples in Freightliner

- `pkg/client/common/base_transport.go`: Transport middleware for logging and retries
- `pkg/server/handlers.go`: HTTP handler middleware
- `pkg/client/common/base_authenticator.go`: Authentication middleware

### Benefits

- Separates cross-cutting concerns from core logic
- Enables composable behavior
- Allows behavior to be added or removed without modifying core code
- Creates a pipeline of operations

## Options Pattern

Freightliner uses the options pattern for configurable function and constructor parameters.

### Pattern

```go
// Define options struct
type ClientOptions struct {
    Region    string
    Logger    *log.Logger
    Transport http.RoundTripper
    Timeout   time.Duration
    // Other options...
}

// Use options in constructors
func NewClient(opts ClientOptions) (*Client, error) {
    // Set defaults for unspecified options
    if opts.Logger == nil {
        opts.Logger = log.NewLogger(log.InfoLevel)
    }
    
    if opts.Timeout == 0 {
        opts.Timeout = 30 * time.Second
    }
    
    // Create client with options
    return &Client{
        region:    opts.Region,
        logger:    opts.Logger,
        transport: opts.Transport,
        timeout:   opts.Timeout,
    }, nil
}
```

### When to Use

- For constructors with many optional parameters
- For functions with configuration options
- When backward compatibility is needed when adding parameters
- When defaults should be provided for unspecified options

### Examples in Freightliner

- `pkg/client/ecr/client.go`: `ClientOptions` for ECR client
- `pkg/client/gcr/client.go`: `ClientOptions` for GCR client
- `pkg/copy/copier.go`: `CopyOptions` for copying images

### Benefits

- Makes function signatures cleaner
- Provides self-documenting parameter names
- Allows for default values
- Enables future extension without breaking changes

## Factory Functions

Freightliner uses factory functions to create instances of interfaces with specific configurations.

### Pattern

```go
// Factory function for creating repositories
func NewRepository(ctx context.Context, repoType, name string, opts Options) (Repository, error) {
    switch repoType {
    case "ecr":
        return ecr.NewRepository(ctx, name, opts.ECR)
    case "gcr":
        return gcr.NewRepository(ctx, name, opts.GCR)
    default:
        return nil, fmt.Errorf("unsupported repository type: %s", repoType)
    }
}
```

### When to Use

- When creating instances of interfaces with various implementations
- When creation logic depends on runtime parameters
- When hiding implementation details from callers
- When instance creation involves complex logic

### Examples in Freightliner

- `pkg/client/common/base_client.go`: Repository factory
- `pkg/service/replicate.go`: Client factory
- `pkg/replication/worker_pool.go`: Worker factory

### Benefits

- Hides implementation details from callers
- Centralizes creation logic
- Enables dependency injection
- Provides a consistent way to create objects

## Caching Strategies

Freightliner uses several caching strategies to improve performance and reduce duplication.

### Pattern

```go
// Simple caching with lazy initialization
type Client struct {
    cacheMutex sync.RWMutex
    cache      map[string]interface{}
}

// Get or create cached item
func (c *Client) GetOrCreate(key string, creator func() interface{}) interface{} {
    // Check cache first with read lock
    c.cacheMutex.RLock()
    if item, found := c.cache[key]; found {
        c.cacheMutex.RUnlock()
        return item
    }
    c.cacheMutex.RUnlock()
    
    // Create item
    item := creator()
    
    // Store in cache with write lock
    c.cacheMutex.Lock()
    c.cache[key] = item
    c.cacheMutex.Unlock()
    
    return item
}
```

### When to Use

- For expensive operations that are called repeatedly
- For objects that are used frequently
- When creating objects is more expensive than caching them
- For resources like connections and clients

### Examples in Freightliner

- `pkg/client/common/base_client.go`: Repository caching
- `pkg/client/common/base_repository.go`: Image caching
- `pkg/client/common/enhanced_client.go`: Transport caching

### Benefits

- Improves performance by avoiding redundant operations
- Reduces resource usage
- Ensures consistent instances for the same keys
- Simplifies resource management

## Code Generation

Freightliner uses code generation for repetitive code patterns.

### Pattern

```go
//go:generate mockgen -source=interfaces.go -destination=../test/mocks/mock_client.go
```

### When to Use

- For generating boilerplate code
- For creating mock implementations for testing
- For generating code from schemas or definitions
- For creating repetitive implementation patterns

### Examples in Freightliner

- `go:generate` directives in interface files
- Mock generation for testing
- Generated client code

### Benefits

- Reduces manual coding errors
- Ensures consistency across generated code
- Makes updates easier when definitions change
- Reduces maintenance burden

## Best Practices

### 1. Use Composition for Shared Behavior

Prefer embedding structs to share behavior rather than creating deep inheritance hierarchies.

### 2. Define Interfaces in Consumer Packages

Define interfaces where they are used, not where they are implemented, to adhere to the Dependency Inversion Principle.

### 3. Keep Utility Functions Focused

Utility functions should have a single responsibility and be well-tested.

### 4. Consider Performance Impact of Caching

Be aware of memory usage and cache invalidation when implementing caching strategies.

### 5. Document Reuse Patterns

Clearly document how and when to use reusable components to encourage proper usage.

### 6. Test Reusable Components Thoroughly

Since reusable components are used widely, ensure they are well-tested with comprehensive test cases.

### 7. Use Consistent Naming Conventions

Follow consistent naming conventions for reusable components to make them easy to discover and understand.

### 8. Consider Thread Safety

Ensure reusable components are thread-safe when they might be used concurrently.

### 9. Avoid Premature Abstraction

Only create reusable components when there's a clear need for reuse, not based on speculation.

### 10. Balance Flexibility and Complexity

More flexible components tend to be more complex. Find the right balance for your use case.
