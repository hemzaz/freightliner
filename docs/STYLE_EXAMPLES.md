# Freightliner Style Guide Examples

This document provides concrete examples of Freightliner's coding style conventions to supplement the general guidelines in GUIDELINES.md.

## Table of Contents
1. [Naming Conventions](#naming-conventions)
2. [Function Design](#function-design)
3. [Error Handling](#error-handling)
4. [Code Organization](#code-organization)
5. [Comments and Documentation](#comments-and-documentation)
6. [Import Organization](#import-organization)
7. [Concurrency Patterns](#concurrency-patterns)
8. [Performance Considerations](#performance-considerations)
9. [Testing Patterns](#testing-patterns)
10. [Interfaces and Abstractions](#interfaces-and-abstractions)

## Naming Conventions

### Consistent Package Names

```go
// Good - Simple, clear package names
package client
package network
package config

// Avoid - Generic or ambiguous package names
package utils
package common
package helpers
```

### Variable Naming

```go
// Good - Descriptive names based on purpose
func ListRepositories(ctx context.Context, registryName string, maxResults int) ([]Repository, error) {
    var repositories []Repository
    pageToken := ""
    
    // Good - Short but clear names in short-lived scopes
    for i, repo := range repositories {
        // ...
    }
    
    // Good - Consistent variable naming for common types
    var (
        client  = newClient(ctx)
        cfg     = loadConfig()
        logger  = cfg.Logger
    )
}

// Avoid - Single letter variables outside of very short scopes
func ListRepositories(ctx context.Context, r string, m int) ([]Repository, error) {
    var rs []Repository
    p := ""
    // ...
}
```

### Constant Naming

```go
// Good - Exported constants use PascalCase
const (
    MaxRetries      = 3
    DefaultTimeout  = 30 * time.Second
    APIVersion      = "v1"
)

// Good - Unexported constants use UPPERCASE
const (
    defaultWorkerCount = 5
    minCacheSize       = 100
)

// Good - Related constants grouped together with a type
type HTTPMethod string
const (
    HTTPGet     HTTPMethod = "GET"
    HTTPPost    HTTPMethod = "POST"
    HTTPPut     HTTPMethod = "PUT"
    HTTPDelete  HTTPMethod = "DELETE"
)
```

### Function and Method Naming

```go
// Good - Verb-noun naming for actions
func FetchRepository(name string) (*Repository, error)
func CopyImage(source, dest string) error
func DeleteTag(tag string) error

// Good - Getter methods without "Get" prefix
func (r *Repository) Name() string
func (r *Repository) Tags() []string

// Good - Boolean methods with predicate names
func (r *Repository) IsEmpty() bool
func (r *Repository) HasTag(tag string) bool
func IsValidImageName(name string) bool

// Avoid - Inconsistent naming patterns
func GetName() string          // Should be Name()
func repository_tags() []string // Should be RepositoryTags() or Tags()
func is_valid(name string) bool // Should be IsValid(name string)
```

## Function Design

### Single Responsibility Functions

```go
// Good - Functions with single responsibilities
func ParseRepositoryPath(path string) (registry, repository string, err error) {
    // Only parses a path into components
    // ...
}

func CreateRegistryClient(ctx context.Context, registryType, region string) (Client, error) {
    // Only creates a client
    // ...
}

// Avoid - Functions that do multiple things
func ParseAndCreateClient(ctx context.Context, path string) (Client, error) {
    // Parses path AND creates client
    // ...
}
```

### Parameter Grouping

```go
// Good - Using option struct for many parameters
type ReplicateOptions struct {
    DryRun        bool
    Force         bool
    Concurrency   int
    Timeout       time.Duration
    IncludeTags   []string
    ExcludeTags   []string
    CheckpointDir string
}

func ReplicateRepository(ctx context.Context, source, dest string, opts ReplicateOptions) error {
    // ...
}

// Instead of:
// func ReplicateRepository(ctx context.Context, source, dest string, dryRun, force bool,
//     concurrency int, timeout time.Duration, includeTags, excludeTags []string,
//     checkpointDir string) error
```

### Return Values

```go
// Good - Return relevant values, use direct returns
func GetRepositoryTags(ctx context.Context, repoName string) ([]string, error) {
    if repoName == "" {
        return nil, errors.InvalidInputf("repository name cannot be empty")
    }
    
    tags, err := client.ListTags(ctx, repoName)
    if err != nil {
        return nil, errors.Wrap(err, "failed to list tags")
    }
    
    return tags, nil
}

// Avoid - Named returns with naked returns
func GetRepositoryTags(ctx context.Context, repoName string) (tags []string, err error) {
    if repoName == "" {
        err = errors.InvalidInputf("repository name cannot be empty")
        return
    }
    
    tags, err = client.ListTags(ctx, repoName)
    if err != nil {
        err = errors.Wrap(err, "failed to list tags")
        return
    }
    
    return
}
```

## Error Handling

### Contextual Errors

```go
// Good - Adding context to errors
func (c *Client) GetRepository(ctx context.Context, name string) (*Repository, error) {
    repo, err := c.client.DescribeRepository(ctx, name)
    if err != nil {
        return nil, errors.Wrap(err, "failed to describe repository")
    }
    // ...
}

// Good - Creating new errors with context
func ParseConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, errors.NotFoundf("config file %s not found", path)
        }
        return nil, errors.Wrap(err, "failed to read config file")
    }
    // ...
}

// Avoid - Returning raw errors without context
func GetRepository(ctx context.Context, name string) (*Repository, error) {
    return c.client.DescribeRepository(ctx, name)
}
```

### Error Types

```go
// Good - Using typed errors for better error handling
var (
    ErrInvalidInput    = errors.New("invalid input provided")
    ErrNotFound        = errors.New("resource not found")
    ErrAuthFailed      = errors.New("authentication failed")
    ErrPermissionDenied = errors.New("permission denied")
)

// Good - Using helper functions for creating specific error types
func (c *Client) GetRepository(ctx context.Context, name string) (*Repository, error) {
    if name == "" {
        return nil, errors.InvalidInputf("repository name cannot be empty")
    }
    
    repo, err := c.client.DescribeRepository(ctx, name)
    if err != nil {
        if isNotFoundError(err) {
            return nil, errors.NotFoundf("repository %s not found", name)
        }
        if isPermissionError(err) {
            return nil, errors.PermissionDenied("insufficient permissions to access repository %s", name)
        }
        return nil, errors.Wrap(err, "failed to describe repository")
    }
    // ...
}
```

### Error Handling Flow

```go
// Good - Clear error handling flow
func ProcessRepository(ctx context.Context, repoName string) error {
    // Early validation
    if repoName == "" {
        return errors.InvalidInputf("repository name cannot be empty")
    }
    
    // Main operations with proper error handling
    repo, err := client.GetRepository(ctx, repoName)
    if err != nil {
        return errors.Wrap(err, "failed to get repository")
    }
    
    tags, err := repo.ListTags(ctx)
    if err != nil {
        return errors.Wrap(err, "failed to list tags")
    }
    
    // Process tags
    for _, tag := range tags {
        if err := processTag(ctx, repo, tag); err != nil {
            return errors.Wrapf(err, "failed to process tag %s", tag)
        }
    }
    
    return nil
}
```

## Code Organization

### Package Structure

```go
// Good - Packages organized by domain
package client    // Handles registry client implementations
package network   // Handles network operations
package config    // Handles configuration
package server    // Handles HTTP server implementation

// Avoid - Packages organized by technical function
package models    // Contains various data models
package utils     // Contains various utilities
package helpers   // Contains various helper functions
```

### File Organization

```go
// Good - File structure for a typical package
client/
├── client.go         // Main client interface and factory functions
├── client_test.go    // Tests for client.go
├── ecr_client.go     // ECR-specific client implementation
├── ecr_client_test.go
├── gcr_client.go     // GCR-specific client implementation
├── gcr_client_test.go
├── auth.go           // Authentication-related functions
├── auth_test.go
├── interfaces.go     // Public interfaces for the package
├── repository.go     // Repository implementation
└── repository_test.go
```

### Method Organization within Types

```go
// Good - Methods organized by purpose
type Repository struct {
    // fields...
}

// Constructor
func NewRepository(name string, client Client) *Repository {
    // ...
}

// Core interface methods
func (r *Repository) Name() string {
    // ...
}

func (r *Repository) ListTags(ctx context.Context) ([]string, error) {
    // ...
}

// Public methods
func (r *Repository) DeleteTag(ctx context.Context, tag string) error {
    // ...
}

func (r *Repository) UpdateMetadata(ctx context.Context, metadata map[string]string) error {
    // ...
}

// Internal helper methods
func (r *Repository) validateTag(tag string) error {
    // ...
}

func (r *Repository) createTagReference(tag string) (name.Tag, error) {
    // ...
}
```

## Comments and Documentation

### Package Documentation

```go
// Package client provides registry client implementations for accessing
// container registries like AWS ECR and Google GCR. It handles authentication,
// repository operations, and image management across different registry types.
package client

// Import statements...
```

### Interface Documentation

```go
// RegistryClient represents a client for container registries.
// It provides methods for accessing repositories and performing
// registry-specific operations.
type RegistryClient interface {
    // GetRepository returns a repository by name.
    // If the repository doesn't exist, it returns an error.
    GetRepository(ctx context.Context, name string) (Repository, error)
    
    // ListRepositories returns a list of repository names in the registry.
    // The prefix parameter can be used to filter repositories by name prefix.
    ListRepositories(ctx context.Context, prefix string) ([]string, error)
    
    // GetRegistryName returns the registry endpoint name.
    GetRegistryName() string
}
```

### Function Documentation

```go
// NewClient creates a new ECR client with the provided options.
// It handles authentication and configures the client for the specified
// AWS region and account.
//
// If accountID is empty, it attempts to detect the AWS account ID
// using the current AWS credentials.
//
// Example:
//
//     client, err := NewClient(ClientOptions{
//         Region: "us-west-2",
//         Logger: logger,
//     })
func NewClient(opts ClientOptions) (*Client, error) {
    // Implementation...
}
```

### Code Comments

```go
// Good - Comments explaining why, not what
func getAuthToken(ctx context.Context) (string, error) {
    // Token is valid for 12 hours, but we refresh every 6 hours to be safe
    if time.Since(c.tokenTimestamp) < 6*time.Hour && c.authToken != "" {
        return c.authToken, nil
    }
    
    // Try AWS SDK auth first, fall back to credential helper
    token, err := c.getAWSAuthToken(ctx)
    if err != nil {
        c.logger.Debug("AWS SDK auth failed, trying credential helper", map[string]interface{}{
            "error": err.Error(),
        })
        return c.getCredentialHelperToken(ctx)
    }
    
    return token, nil
}
```

## Import Organization

```go
// Good - Imports organized into groups
package main

import (
    // Standard library imports
    "context"
    "fmt"
    "os"
    "time"
    
    // Third-party imports
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/google/go-containerregistry/pkg/name"
    "github.com/spf13/cobra"
    
    // Internal project imports
    "freightliner/pkg/client"
    "freightliner/pkg/config"
    "freightliner/pkg/helper/log"
)
```

## Concurrency Patterns

### Worker Pool

```go
// Good - Worker pool implementation
func ReplicateRepositories(ctx context.Context, repos []string, destRegistry string, workerCount int) error {
    // Create a worker pool with the specified number of workers
    pool := workerpool.New(workerCount)
    
    // Create a channel to collect errors
    errCh := make(chan error, len(repos))
    
    // Submit jobs to the worker pool
    for _, repo := range repos {
        repoName := repo // Capture variable for closure
        pool.Submit(func() {
            err := replicateRepository(ctx, repoName, destRegistry)
            if err != nil {
                errCh <- fmt.Errorf("failed to replicate %s: %w", repoName, err)
            }
        })
    }
    
    // Wait for all jobs to complete
    pool.Wait()
    close(errCh)
    
    // Collect any errors
    var errs []string
    for err := range errCh {
        errs = append(errs, err.Error())
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("encountered %d errors during replication: %s", len(errs), strings.Join(errs, "; "))
    }
    
    return nil
}
```

### Context Handling

```go
// Good - Proper context usage
func ProcessImage(ctx context.Context, imageName string) error {
    // Create a child context with timeout
    processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel() // Always defer context cancellation
    
    // Use the context in all operations
    img, err := client.GetImage(processCtx, imageName)
    if err != nil {
        return err
    }
    
    // Check for context cancellation between operations
    select {
    case <-processCtx.Done():
        return processCtx.Err()
    default:
        // Continue processing
    }
    
    // Use the context for the next operation
    return processImage(processCtx, img)
}
```

### Mutex Usage

```go
// Good - Proper mutex handling
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}

func (c *Cache) Get(key string) (Item, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    item, found := c.items[key]
    return item, found
}

func (c *Cache) Set(key string, item Item) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.items[key] = item
}

func (c *Cache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    delete(c.items, key)
}
```

## Performance Considerations

### Memory Allocation

```go
// Good - Efficient memory allocation
func ProcessItems(items []Item) []Result {
    // Pre-allocate slice with known capacity
    results := make([]Result, 0, len(items))
    
    for _, item := range items {
        result := processItem(item)
        results = append(results, result)
    }
    
    return results
}

// Avoid - Inefficient memory allocation
func ProcessItems(items []Item) []Result {
    var results []Result // Will cause multiple reallocations
    
    for _, item := range items {
        result := processItem(item)
        results = append(results, result)
    }
    
    return results
}
```

### String Concatenation

```go
// Good - Efficient string building
func BuildReport(items []Item) string {
    var builder strings.Builder
    
    // Pre-allocate for rough estimated size
    builder.Grow(len(items) * 20)
    
    builder.WriteString("Report:\n")
    for _, item := range items {
        builder.WriteString(fmt.Sprintf("- %s: %d\n", item.Name, item.Count))
    }
    
    return builder.String()
}

// Avoid - Inefficient string concatenation
func BuildReport(items []Item) string {
    result := "Report:\n"
    
    for _, item := range items {
        // Creates a new string on each iteration
        result += fmt.Sprintf("- %s: %d\n", item.Name, item.Count)
    }
    
    return result
}
```

### Batching Operations

```go
// Good - Batched API calls
func GetRepositoryTags(ctx context.Context, repo string) ([]string, error) {
    const batchSize = 100
    var allTags []string
    var nextToken string
    
    for {
        tags, newToken, err := client.ListTags(ctx, repo, nextToken, batchSize)
        if err != nil {
            return nil, err
        }
        
        allTags = append(allTags, tags...)
        
        if newToken == "" {
            break
        }
        nextToken = newToken
    }
    
    return allTags, nil
}
```

## Testing Patterns

### Table-Driven Tests

```go
// Good - Table-driven testing pattern
func TestParseRepositoryPath(t *testing.T) {
    tests := []struct {
        name           string
        input          string
        wantRegistry   string
        wantRepository string
        wantErr        bool
    }{
        {
            name:           "Valid ECR path",
            input:          "ecr/my-repo",
            wantRegistry:   "ecr",
            wantRepository: "my-repo",
            wantErr:        false,
        },
        {
            name:           "Valid GCR path",
            input:          "gcr/project/repo",
            wantRegistry:   "gcr",
            wantRepository: "project/repo",
            wantErr:        false,
        },
        {
            name:           "Empty path",
            input:          "",
            wantRegistry:   "",
            wantRepository: "",
            wantErr:        true,
        },
        {
            name:           "Missing repository",
            input:          "ecr/",
            wantRegistry:   "",
            wantRepository: "",
            wantErr:        true,
        },
        {
            name:           "Invalid format",
            input:          "ecr-my-repo",
            wantRegistry:   "",
            wantRepository: "",
            wantErr:        true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            gotRegistry, gotRepository, err := ParseRepositoryPath(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.wantRegistry, gotRegistry)
            assert.Equal(t, tt.wantRepository, gotRepository)
        })
    }
}
```

### Mocking Dependencies

```go
// Good - Using interfaces and mocks for testing
func TestRegistryService_ListRepositories(t *testing.T) {
    // Create mock controller
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    // Create a mock client
    mockClient := mocks.NewMockRegistryClient(ctrl)
    
    // Set expectations
    mockClient.EXPECT().
        ListRepositories(gomock.Any(), "prefix").
        Return([]string{"repo1", "repo2"}, nil)
    
    // Create the service with the mock client
    service := NewRegistryService(mockClient)
    
    // Call the method
    repos, err := service.ListRepositories(context.Background(), "prefix")
    
    // Assert results
    assert.NoError(t, err)
    assert.Equal(t, []string{"repo1", "repo2"}, repos)
}
```

### Subtest Organization

```go
// Good - Organized subtests
func TestRepository(t *testing.T) {
    // Setup common test fixtures
    client := newTestClient(t)
    repo := NewRepository("test-repo", client)
    
    // Test ListTags method
    t.Run("ListTags", func(t *testing.T) {
        t.Run("Empty repository", func(t *testing.T) {
            // Test specific case
        })
        
        t.Run("With tags", func(t *testing.T) {
            // Test specific case
        })
        
        t.Run("With pagination", func(t *testing.T) {
            // Test specific case
        })
    })
    
    // Test DeleteTag method
    t.Run("DeleteTag", func(t *testing.T) {
        t.Run("Valid tag", func(t *testing.T) {
            // Test specific case
        })
        
        t.Run("Nonexistent tag", func(t *testing.T) {
            // Test specific case
        })
    })
}
```

## Interfaces and Abstractions

### Interface Definition

```go
// Good - Focused interfaces with clear purpose
// Reader represents a basic reader interface with a single method
type Reader interface {
    Read(ctx context.Context, path string) ([]byte, error)
}

// Writer represents a basic writer interface with a single method
type Writer interface {
    Write(ctx context.Context, path string, data []byte) error
}

// ReadWriter combines reading and writing capabilities
type ReadWriter interface {
    Reader
    Writer
}

// Repository represents a container image repository
type Repository interface {
    // Basic information methods
    GetName() string
    GetURI() string
    
    // Tag operations
    ListTags(ctx context.Context) ([]string, error)
    GetTag(ctx context.Context, tag string) (Image, error)
    DeleteTag(ctx context.Context, tag string) error
    
    // Image operations
    GetImage(ctx context.Context, digest string) (Image, error)
    PutImage(ctx context.Context, img Image) error
}
```

### Dependency Injection

```go
// Good - Dependency injection through interfaces
type ReplicationService struct {
    sourceClient RegistryClient
    destClient   RegistryClient
    logger       Logger
}

// NewReplicationService creates a new replication service with the provided dependencies
func NewReplicationService(sourceClient, destClient RegistryClient, logger Logger) *ReplicationService {
    return &ReplicationService{
        sourceClient: sourceClient,
        destClient:   destClient,
        logger:       logger,
    }
}

// Instead of:
// type ReplicationService struct {}
// func NewReplicationService() *ReplicationService {
//     sourceClient := createSourceClient()
//     destClient := createDestClient()
//     logger := createLogger()
//     // ...
// }
```

### Interface Segregation

```go
// Good - Small, focused interfaces
// ImageFetcher only handles fetching images
type ImageFetcher interface {
    FetchImage(ctx context.Context, ref string) (Image, error)
}

// ImagePusher only handles pushing images
type ImagePusher interface {
    PushImage(ctx context.Context, img Image, ref string) error
}

// TagLister only handles listing tags
type TagLister interface {
    ListTags(ctx context.Context, repo string) ([]string, error)
}

// Combining interfaces for services that need multiple capabilities
type ImageTransferService struct {
    fetcher ImageFetcher
    pusher  ImagePusher
    lister  TagLister
}

// Instead of:
// type RegistryClient interface {
//     FetchImage(ctx context.Context, ref string) (Image, error)
//     PushImage(ctx context.Context, img Image, ref string) error
//     ListTags(ctx context.Context, repo string) ([]string, error)
//     // And many other methods...
// }
```
