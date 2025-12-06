# Component Modification Guide for Registry Support

## Overview

This document outlines the specific modifications required for each component in the Freightliner codebase to support the new multi-registry architecture. It provides detailed implementation guidance, code examples, and migration strategies.

## Table of Contents

1. [Component Overview](#component-overview)
2. [Core Components](#core-components)
3. [Service Layer](#service-layer)
4. [Client Layer](#client-layer)
5. [Server/API Layer](#server-api-layer)
6. [CLI Commands](#cli-commands)
7. [Configuration Layer](#configuration-layer)
8. [Testing Components](#testing-components)
9. [Migration Checklist](#migration-checklist)

## 1. Component Overview

### 1.1 Affected Components

| Component | Priority | Complexity | Impact |
|-----------|----------|------------|--------|
| pkg/config/config.go | High | Medium | Breaking change to config structure |
| pkg/registry/* (NEW) | High | High | New package with core logic |
| pkg/client/ecr/* | High | Medium | Refactor to use new interfaces |
| pkg/client/gcr/* | High | Medium | Refactor to use new interfaces |
| pkg/client/common/* | Medium | Low | Extend utility functions |
| pkg/service/replicate.go | High | Medium | Update to use RegistryManager |
| pkg/service/tree_replicate.go | High | Medium | Update to use RegistryManager |
| pkg/server/handlers.go | Medium | Low | Update API endpoints |
| pkg/server/types.go | Medium | Low | Add new request/response types |
| cmd/replicate.go | High | Low | Update CLI flags and logic |
| cmd/tree-replicate.go | High | Low | Update CLI flags and logic |
| cmd/serve.go | Medium | Low | Initialize RegistryManager |
| deployments/kubernetes/* | Low | Low | Update config examples |
| docs/* | Low | Low | Update documentation |

### 1.2 Dependency Graph

```
                    ┌─────────────────┐
                    │  pkg/registry   │
                    │  (NEW)          │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
    ┌─────────▼────────┐ ┌──▼──────────┐ ┌─▼─────────────┐
    │ pkg/service/*    │ │ pkg/server/*│ │ cmd/*         │
    │ (MODIFIED)       │ │ (MODIFIED)  │ │ (MODIFIED)    │
    └──────────────────┘ └─────────────┘ └───────────────┘
              │
    ┌─────────▼────────┐
    │ pkg/client/*     │
    │ (REFACTORED)     │
    └──────────────────┘
```

## 2. Core Components

### 2.1 New Package: pkg/registry

**Location:** `/Users/elad/PROJ/freightliner/pkg/registry/`

#### 2.1.1 Registry Manager (manager.go)

```go
package registry

import (
    "context"
    "fmt"
    "sync"
    "time"

    "freightliner/pkg/config"
    "freightliner/pkg/helper/log"
)

// RegistryManager coordinates all registry operations
type RegistryManager struct {
    config      *config.RegistryConfig
    registries  map[string]RegistryClient
    selector    *RegistrySelector
    authCache   *AuthCache
    logger      log.Logger
    mu          sync.RWMutex
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager(cfg *config.RegistryConfig, logger log.Logger) (*RegistryManager, error) {
    if cfg == nil {
        return nil, fmt.Errorf("registry config is required")
    }

    if logger == nil {
        logger = log.NewBasicLogger(log.InfoLevel)
    }

    rm := &RegistryManager{
        config:     cfg,
        registries: make(map[string]RegistryClient),
        authCache:  NewAuthCache(15 * time.Minute), // 15-minute TTL
        logger:     logger,
    }

    // Initialize registry selector
    rm.selector = NewRegistrySelector(cfg, logger)

    return rm, nil
}

// Initialize sets up all configured registries
func (rm *RegistryManager) Initialize(ctx context.Context) error {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    factory := NewRegistryFactory(rm.logger)

    for _, regDef := range rm.config.Definitions {
        if !regDef.Enabled {
            rm.logger.WithField("registry", regDef.Name).Debug("Skipping disabled registry")
            continue
        }

        // Create registry client
        client, err := factory.CreateClient(ctx, &regDef)
        if err != nil {
            rm.logger.WithFields(map[string]interface{}{
                "registry": regDef.Name,
                "type":     regDef.Type,
                "error":    err,
            }).Warn("Failed to create registry client")

            // Continue with other registries instead of failing
            continue
        }

        rm.registries[regDef.Name] = client
        rm.logger.WithField("registry", regDef.Name).Info("Registry initialized")
    }

    if len(rm.registries) == 0 {
        return fmt.Errorf("no registries were successfully initialized")
    }

    return nil
}

// GetRegistry returns a specific registry by name
func (rm *RegistryManager) GetRegistry(name string) (RegistryClient, error) {
    rm.mu.RLock()
    defer rm.mu.RUnlock()

    client, ok := rm.registries[name]
    if !ok {
        return nil, fmt.Errorf("registry not found: %s", name)
    }

    return client, nil
}

// SelectRegistry chooses the best registry based on policy
func (rm *RegistryManager) SelectRegistry(ctx context.Context, imageName string, policy SelectionPolicy) (RegistryClient, error) {
    rm.mu.RLock()
    defer rm.mu.RUnlock()

    return rm.selector.Select(ctx, imageName, policy, rm.registries)
}

// GetImageURI constructs the full image URI
func (rm *RegistryManager) GetImageURI(registryName, imageName, tag string) (string, error) {
    client, err := rm.GetRegistry(registryName)
    if err != nil {
        return "", err
    }

    prefix := client.ImagePrefix()
    if prefix == "" {
        return fmt.Sprintf("%s:%s", imageName, tag), nil
    }

    return fmt.Sprintf("%s/%s:%s", prefix, imageName, tag), nil
}

// ValidateAccess tests connectivity to a registry
func (rm *RegistryManager) ValidateAccess(ctx context.Context, registryName string) error {
    client, err := rm.GetRegistry(registryName)
    if err != nil {
        return err
    }

    if err := client.Authenticate(ctx); err != nil {
        return fmt.Errorf("authentication failed: %w", err)
    }

    if err := client.Ping(ctx); err != nil {
        return fmt.Errorf("connectivity check failed: %w", err)
    }

    return nil
}

// ListRegistries returns all enabled registries
func (rm *RegistryManager) ListRegistries() []RegistryInfo {
    rm.mu.RLock()
    defer rm.mu.RUnlock()

    var registries []RegistryInfo
    for name, client := range rm.registries {
        registries = append(registries, RegistryInfo{
            Name:        name,
            Type:        client.Type(),
            Endpoint:    client.Endpoint(),
            ImagePrefix: client.ImagePrefix(),
        })
    }

    return registries
}

// Close shuts down all registry connections
func (rm *RegistryManager) Close() error {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    var errors []error
    for name, client := range rm.registries {
        if err := client.Close(); err != nil {
            rm.logger.WithFields(map[string]interface{}{
                "registry": name,
                "error":    err,
            }).Warn("Failed to close registry client")
            errors = append(errors, err)
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("failed to close %d registry clients", len(errors))
    }

    return nil
}
```

#### 2.1.2 Registry Types (types.go)

```go
package registry

import (
    "context"
    "time"
)

// RegistryType represents the type of registry
type RegistryType string

const (
    RegistryTypeECR       RegistryType = "ecr"
    RegistryTypeGCR       RegistryType = "gcr"
    RegistryTypeACR       RegistryType = "acr"
    RegistryTypeDockerHub RegistryType = "dockerhub"
    RegistryTypeHarbor    RegistryType = "harbor"
    RegistryTypeQuay      RegistryType = "quay"
    RegistryTypeGeneric   RegistryType = "generic"
)

// RegistryClient is the unified interface for all registry types
type RegistryClient interface {
    // Metadata
    Name() string
    Type() RegistryType
    Endpoint() string
    ImagePrefix() string

    // Authentication
    Authenticate(ctx context.Context) error
    RefreshCredentials(ctx context.Context) error

    // Repository operations
    ListRepositories(ctx context.Context) ([]string, error)
    GetRepository(ctx context.Context, name string) (Repository, error)
    CreateRepository(ctx context.Context, name string, opts CreateOptions) error
    DeleteRepository(ctx context.Context, name string) error

    // Image operations
    PushImage(ctx context.Context, image Image) error
    PullImage(ctx context.Context, ref ImageReference) (Image, error)
    TagImage(ctx context.Context, src, dst ImageReference) error
    DeleteImage(ctx context.Context, ref ImageReference) error

    // Manifest operations
    GetManifest(ctx context.Context, ref ImageReference) (Manifest, error)
    PutManifest(ctx context.Context, ref ImageReference, manifest Manifest) error

    // Health
    Ping(ctx context.Context) error
    GetMetrics() RegistryMetrics

    // Lifecycle
    Close() error
}

// ImageReference represents an image reference
type ImageReference struct {
    Registry   string
    Repository string
    Tag        string
    Digest     string
}

// Image represents a container image
type Image struct {
    Reference  ImageReference
    Manifest   Manifest
    Layers     []Layer
    Config     ImageConfig
}

// Repository represents a container repository
type Repository interface {
    Name() string
    ListTags(ctx context.Context) ([]string, error)
    GetTag(ctx context.Context, tag string) (Image, error)
}

// Manifest represents an image manifest
type Manifest interface {
    MediaType() string
    SchemaVersion() int
    Config() ConfigDescriptor
    Layers() []LayerDescriptor
}

// Layer represents an image layer
type Layer interface {
    Digest() string
    Size() int64
    MediaType() string
}

// CreateOptions contains options for creating repositories
type CreateOptions struct {
    Public      bool
    ScanOnPush  bool
    Immutable   bool
    Description string
    Labels      map[string]string
}

// RegistryMetrics contains registry performance metrics
type RegistryMetrics struct {
    TotalRequests      int64
    SuccessfulRequests int64
    FailedRequests     int64
    AverageLatency     time.Duration
    ActiveConnections  int
    IdleConnections    int
}

// RegistryInfo contains basic registry information
type RegistryInfo struct {
    Name        string
    Type        RegistryType
    Endpoint    string
    ImagePrefix string
}

// SelectionPolicy defines how to select a registry
type SelectionPolicy struct {
    Strategy         SelectionStrategy
    PreferredTypes   []RegistryType
    RequirePrivate   bool
    IncludePatterns  []string
    ExcludePatterns  []string
    ComponentContext string
}

// SelectionStrategy defines the registry selection strategy
type SelectionStrategy string

const (
    StrategyPreferPrivate  SelectionStrategy = "prefer-private"
    StrategyPublicOnly     SelectionStrategy = "public-only"
    StrategyCustomPriority SelectionStrategy = "custom-priority"
    StrategyFastestResponse SelectionStrategy = "fastest-response"
)
```

#### 2.1.3 Registry Factory (factory.go)

```go
package registry

import (
    "context"
    "fmt"

    "freightliner/pkg/config"
    "freightliner/pkg/helper/log"
    "freightliner/pkg/registry/clients"
)

// RegistryFactory creates registry clients
type RegistryFactory struct {
    logger log.Logger
}

// NewRegistryFactory creates a new registry factory
func NewRegistryFactory(logger log.Logger) *RegistryFactory {
    return &RegistryFactory{
        logger: logger,
    }
}

// CreateClient creates a registry client based on configuration
func (f *RegistryFactory) CreateClient(ctx context.Context, def *config.RegistryDefinition) (RegistryClient, error) {
    switch def.Type {
    case "ecr":
        return clients.NewECRClient(ctx, def, f.logger)
    case "gcr":
        return clients.NewGCRClient(ctx, def, f.logger)
    case "acr":
        return clients.NewACRClient(ctx, def, f.logger)
    case "dockerhub":
        return clients.NewDockerHubClient(ctx, def, f.logger)
    case "harbor":
        return clients.NewHarborClient(ctx, def, f.logger)
    case "quay":
        return clients.NewQuayClient(ctx, def, f.logger)
    case "generic":
        return clients.NewGenericClient(ctx, def, f.logger)
    default:
        return nil, fmt.Errorf("unsupported registry type: %s", def.Type)
    }
}
```

## 3. Service Layer

### 3.1 pkg/service/replicate.go

**Current Implementation:**
```go
// OLD: Hard-coded ECR/GCR logic
func (s *Service) Replicate(ctx context.Context, src, dst string) error {
    srcClient := ecr.NewClient(...)  // Hard-coded
    dstClient := gcr.NewClient(...)  // Hard-coded
    // ...
}
```

**New Implementation:**
```go
// NEW: Dynamic registry selection
type Service struct {
    config          *config.Config
    registryManager *registry.RegistryManager  // NEW
    logger          log.Logger
    metrics         *metrics.Registry
}

// NewService creates a new service with registry manager
func NewService(cfg *config.Config, logger log.Logger) (*Service, error) {
    // Initialize registry manager
    rm, err := registry.NewRegistryManager(&cfg.Registries, logger)
    if err != nil {
        return nil, fmt.Errorf("failed to create registry manager: %w", err)
    }

    if err := rm.Initialize(context.Background()); err != nil {
        return nil, fmt.Errorf("failed to initialize registries: %w", err)
    }

    return &Service{
        config:          cfg,
        registryManager: rm,
        logger:          logger,
        metrics:         metrics.NewRegistry(),
    }, nil
}

// Replicate replicates an image between registries
func (s *Service) Replicate(ctx context.Context, srcImage, dstImage string) error {
    // Parse image references
    srcRef, err := registry.ParseImageReference(srcImage)
    if err != nil {
        return fmt.Errorf("invalid source image: %w", err)
    }

    dstRef, err := registry.ParseImageReference(dstImage)
    if err != nil {
        return fmt.Errorf("invalid destination image: %w", err)
    }

    // Select source registry
    srcRegistry, err := s.registryManager.SelectRegistry(ctx, srcImage, registry.SelectionPolicy{
        Strategy:         registry.StrategyPreferPrivate,
        ComponentContext: "replication-service",
    })
    if err != nil {
        return fmt.Errorf("failed to select source registry: %w", err)
    }

    // Select destination registry
    dstRegistry, err := s.registryManager.SelectRegistry(ctx, dstImage, registry.SelectionPolicy{
        Strategy:         registry.StrategyPreferPrivate,
        ComponentContext: "replication-service",
    })
    if err != nil {
        return fmt.Errorf("failed to select destination registry: %w", err)
    }

    // Pull image from source
    s.logger.WithFields(map[string]interface{}{
        "source": srcImage,
        "registry": srcRegistry.Name(),
    }).Info("Pulling image from source registry")

    image, err := srcRegistry.PullImage(ctx, srcRef)
    if err != nil {
        s.metrics.RecordReplicationError(srcRegistry.Name(), dstRegistry.Name(), "pull_failed")
        return fmt.Errorf("failed to pull image: %w", err)
    }

    // Push image to destination
    s.logger.WithFields(map[string]interface{}{
        "destination": dstImage,
        "registry": dstRegistry.Name(),
    }).Info("Pushing image to destination registry")

    if err := dstRegistry.PushImage(ctx, image); err != nil {
        s.metrics.RecordReplicationError(srcRegistry.Name(), dstRegistry.Name(), "push_failed")
        return fmt.Errorf("failed to push image: %w", err)
    }

    // Record success metrics
    s.metrics.RecordReplication(
        srcRegistry.Name(),
        dstRegistry.Name(),
        "success",
        time.Since(startTime),
        image.Size(),
        len(image.Layers),
    )

    return nil
}

// Close gracefully shuts down the service
func (s *Service) Close() error {
    return s.registryManager.Close()
}
```

### 3.2 pkg/service/tree_replicate.go

```go
// TreeReplicateService handles bulk replication
type TreeReplicateService struct {
    service         *Service
    registryManager *registry.RegistryManager  // NEW
    logger          log.Logger
}

// ReplicateTree replicates entire repository trees
func (s *TreeReplicateService) ReplicateTree(ctx context.Context, srcRegistry, dstRegistry string, opts TreeReplicateOptions) error {
    // Get source registry client
    src, err := s.registryManager.GetRegistry(srcRegistry)
    if err != nil {
        return fmt.Errorf("source registry not found: %w", err)
    }

    // Get destination registry client
    dst, err := s.registryManager.GetRegistry(dstRegistry)
    if err != nil {
        return fmt.Errorf("destination registry not found: %w", err)
    }

    // List all repositories in source
    repos, err := src.ListRepositories(ctx)
    if err != nil {
        return fmt.Errorf("failed to list repositories: %w", err)
    }

    // Filter repositories based on options
    repos = s.filterRepositories(repos, opts)

    // Replicate each repository
    for _, repo := range repos {
        if err := s.replicateRepository(ctx, src, dst, repo, opts); err != nil {
            s.logger.WithFields(map[string]interface{}{
                "repository": repo,
                "error":      err,
            }).Error("Failed to replicate repository")

            if !opts.ContinueOnError {
                return err
            }
        }
    }

    return nil
}
```

## 4. Client Layer

### 4.1 pkg/client/ecr/client.go

**Refactor to implement RegistryClient interface:**

```go
package ecr

import (
    "context"

    "freightliner/pkg/registry"
    "freightliner/pkg/config"
)

// Client implements registry.RegistryClient for AWS ECR
type Client struct {
    name        string
    config      *config.ECRConfig
    authConfig  *config.AuthConfig
    authProvider registry.AuthProvider
    endpoint    string
    imagePrefix string
    logger      log.Logger
}

// NewClient creates a new ECR client
func NewClient(ctx context.Context, def *config.RegistryDefinition, logger log.Logger) (*Client, error) {
    // Extract ECR-specific config
    ecrConfig, err := extractECRConfig(def)
    if err != nil {
        return nil, err
    }

    // Create authentication provider
    authProvider, err := createAuthProvider(ctx, def.Authentication)
    if err != nil {
        return nil, err
    }

    // Construct image prefix
    imagePrefix := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com",
        ecrConfig.AccountID,
        ecrConfig.Region)

    client := &Client{
        name:         def.Name,
        config:       ecrConfig,
        authConfig:   &def.Authentication,
        authProvider: authProvider,
        imagePrefix:  imagePrefix,
        logger:       logger,
    }

    // Initial authentication
    if err := client.Authenticate(ctx); err != nil {
        return nil, fmt.Errorf("initial authentication failed: %w", err)
    }

    return client, nil
}

// Implement RegistryClient interface methods
func (c *Client) Name() string {
    return c.name
}

func (c *Client) Type() registry.RegistryType {
    return registry.RegistryTypeECR
}

func (c *Client) Endpoint() string {
    return c.imagePrefix
}

func (c *Client) ImagePrefix() string {
    return c.imagePrefix
}

// ... implement all other interface methods
```

### 4.2 pkg/client/common/registry_util.go

**Extend with new utility functions:**

```go
// ParseImageReference parses a full image URI into components
func (u *RegistryUtil) ParseImageReference(uri string) (*registry.ImageReference, error) {
    // Parse: registry.example.com/repo/image:tag@sha256:digest
    parts := strings.Split(uri, "/")
    // ... parsing logic

    return &registry.ImageReference{
        Registry:   registryHost,
        Repository: repoName,
        Tag:        tag,
        Digest:     digest,
    }, nil
}

// DetectRegistryType attempts to detect the registry type from the URI
func (u *RegistryUtil) DetectRegistryType(uri string) (registry.RegistryType, error) {
    switch {
    case strings.Contains(uri, ".dkr.ecr.") && strings.Contains(uri, ".amazonaws.com"):
        return registry.RegistryTypeECR, nil
    case strings.Contains(uri, ".gcr.io"):
        return registry.RegistryTypeGCR, nil
    case strings.Contains(uri, ".azurecr.io"):
        return registry.RegistryTypeACR, nil
    case strings.Contains(uri, "docker.io") || strings.Contains(uri, "index.docker.io"):
        return registry.RegistryTypeDockerHub, nil
    default:
        return registry.RegistryTypeGeneric, nil
    }
}
```

## 5. Server/API Layer

### 5.1 pkg/server/handlers.go

```go
// Add new endpoint for registry management
func (s *Server) handleListRegistries(w http.ResponseWriter, r *http.Request) {
    registries := s.registryManager.ListRegistries()

    response := ListRegistriesResponse{
        Registries: registries,
        Count:      len(registries),
    }

    s.writeJSON(w, http.StatusOK, response)
}

// Add endpoint to validate registry access
func (s *Server) handleValidateRegistry(w http.ResponseWriter, r *http.Request) {
    var req ValidateRegistryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    ctx := r.Context()
    if err := s.registryManager.ValidateAccess(ctx, req.RegistryName); err != nil {
        s.writeError(w, http.StatusBadGateway, fmt.Sprintf("validation failed: %v", err))
        return
    }

    s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Update existing replicate endpoint
func (s *Server) handleReplicate(w http.ResponseWriter, r *http.Request) {
    var req ReplicateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    // Validate input
    if req.Source == "" || req.Destination == "" {
        s.writeError(w, http.StatusBadRequest, "source and destination are required")
        return
    }

    // NEW: Support explicit registry selection
    if req.SourceRegistry != "" {
        // Use specific source registry
        srcClient, err := s.registryManager.GetRegistry(req.SourceRegistry)
        if err != nil {
            s.writeError(w, http.StatusBadRequest, fmt.Sprintf("source registry not found: %v", err))
            return
        }
        // ... use srcClient
    } else {
        // Use automatic selection
        // ... existing logic with SelectRegistry
    }

    // Call service
    if err := s.service.Replicate(r.Context(), req.Source, req.Destination); err != nil {
        s.writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    s.writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
```

### 5.2 pkg/server/types.go

```go
// Add new request/response types
type ReplicateRequest struct {
    Source          string `json:"source"`           // Image URI
    Destination     string `json:"destination"`      // Image URI
    SourceRegistry  string `json:"sourceRegistry"`   // NEW: Optional explicit registry
    DestRegistry    string `json:"destRegistry"`     // NEW: Optional explicit registry
    Force           bool   `json:"force"`
    Tags            []string `json:"tags"`
}

type ListRegistriesResponse struct {
    Registries []registry.RegistryInfo `json:"registries"`
    Count      int                     `json:"count"`
}

type ValidateRegistryRequest struct {
    RegistryName string `json:"registryName"`
}

type RegistryHealthResponse struct {
    Registries []RegistryHealth `json:"registries"`
}

type RegistryHealth struct {
    Name         string    `json:"name"`
    Type         string    `json:"type"`
    Status       string    `json:"status"` // healthy, unhealthy, degraded
    LastCheck    time.Time `json:"lastCheck"`
    ResponseTime string    `json:"responseTime"`
    Error        string    `json:"error,omitempty"`
}
```

## 6. CLI Commands

### 6.1 cmd/replicate.go

```go
var replicateCmd = &cobra.Command{
    Use:   "replicate SOURCE DESTINATION",
    Short: "Replicate an image between registries",
    Long: `Replicate a container image from a source registry to a destination registry.

Examples:
  # Automatic registry selection
  freightliner replicate my-image:v1.0 my-image:v1.0

  # Explicit source registry
  freightliner replicate --source-registry=aws-ecr-prod my-image:v1.0 my-image:v1.0

  # Explicit source and destination registries
  freightliner replicate \
    --source-registry=dockerhub \
    --dest-registry=harbor-prod \
    nginx:latest internal/nginx:latest
`,
    Args: cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Load configuration
        cfg, err := config.LoadConfig(configFile)
        if err != nil {
            return fmt.Errorf("failed to load config: %w", err)
        }

        // Create service with registry manager
        svc, err := service.NewService(cfg, logger)
        if err != nil {
            return fmt.Errorf("failed to create service: %w", err)
        }
        defer svc.Close()

        // NEW: Support explicit registry selection
        if sourceRegistry != "" || destRegistry != "" {
            return replicateExplicit(cmd.Context(), svc, args[0], args[1])
        }

        // Use automatic selection
        return svc.Replicate(cmd.Context(), args[0], args[1])
    },
}

func init() {
    // NEW flags
    replicateCmd.Flags().StringVar(&sourceRegistry, "source-registry", "",
        "Explicit source registry name from configuration")
    replicateCmd.Flags().StringVar(&destRegistry, "dest-registry", "",
        "Explicit destination registry name from configuration")
    replicateCmd.Flags().StringSliceVar(&tags, "tags", []string{},
        "Specific tags to replicate (default: all tags)")

    // Existing flags
    replicateCmd.Flags().BoolVar(&force, "force", false, "Force overwrite")
    replicateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run")
}
```

### 6.2 cmd/serve.go

```go
var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the Freightliner API server",
    RunE: func(cmd *cobra.Command, args []error) error {
        // Load configuration
        cfg, err := config.LoadConfig(configFile)
        if err != nil {
            return err
        }

        // Create service with registry manager
        svc, err := service.NewService(cfg, logger)
        if err != nil {
            return err
        }
        defer svc.Close()

        // Validate all configured registries on startup
        if validateOnStartup {
            logger.Info("Validating registry connections...")
            for _, reg := range svc.GetRegistryManager().ListRegistries() {
                if err := svc.GetRegistryManager().ValidateAccess(cmd.Context(), reg.Name); err != nil {
                    logger.WithFields(map[string]interface{}{
                        "registry": reg.Name,
                        "error":    err,
                    }).Warn("Registry validation failed")
                } else {
                    logger.WithField("registry", reg.Name).Info("Registry validated successfully")
                }
            }
        }

        // Create and start server
        srv := server.NewServer(cfg, svc, logger)
        return srv.ListenAndServe()
    },
}

func init() {
    serveCmd.Flags().BoolVar(&validateOnStartup, "validate-registries", true,
        "Validate all registry connections on startup")
}
```

## 7. Configuration Layer

### 7.1 pkg/config/config.go

```go
// Add new configuration structures
type Config struct {
    LogLevel   string
    Registries RegistryConfig        // NEW
    Workers    WorkerConfig
    Encryption EncryptionConfig
    Secrets    SecretsConfig
    Server     ServerConfig

    // Legacy fields (for backward compatibility)
    ECR ECRConfig
    GCR GCRConfig
}

type RegistryConfig struct {
    DefaultStrategy string                 `yaml:"defaultStrategy" json:"defaultStrategy"`
    Defaults        RegistryDefaults       `yaml:"defaults" json:"defaults"`
    Definitions     []RegistryDefinition   `yaml:"definitions" json:"definitions"`
}

type RegistryDefaults struct {
    Timeout            time.Duration `yaml:"timeout" json:"timeout"`
    RetryAttempts      int           `yaml:"retryAttempts" json:"retryAttempts"`
    RetryDelay         time.Duration `yaml:"retryDelay" json:"retryDelay"`
    ConnectionPoolSize int           `yaml:"connectionPoolSize" json:"connectionPoolSize"`
    TLSVerify          bool          `yaml:"tlsVerify" json:"tlsVerify"`
}

type RegistryDefinition struct {
    Name           string         `yaml:"name" json:"name"`
    Type           string         `yaml:"type" json:"type"`
    Enabled        bool           `yaml:"enabled" json:"enabled"`
    Priority       int            `yaml:"priority" json:"priority"`
    Config         interface{}    `yaml:"config" json:"config"`       // Type-specific config
    Authentication AuthConfig     `yaml:"authentication" json:"authentication"`
    ImagePrefix    string         `yaml:"imagePrefix" json:"imagePrefix"`
    Tags           map[string]string `yaml:"tags" json:"tags"`
}

type AuthConfig struct {
    Method        string              `yaml:"method" json:"method"`
    // Generic fields (used by multiple methods)
    Username      string              `yaml:"username" json:"username"`
    Password      string              `yaml:"password" json:"password"`
    Token         string              `yaml:"token" json:"token"`
    // AWS-specific
    AccessKeyID     string            `yaml:"accessKeyID" json:"accessKeyID"`
    SecretAccessKey string            `yaml:"secretAccessKey" json:"secretAccessKey"`
    RoleARN         string            `yaml:"roleARN" json:"roleARN"`
    // GCP-specific
    CredentialsFile string            `yaml:"credentialsFile" json:"credentialsFile"`
    CredentialsJSON string            `yaml:"credentialsJSON" json:"credentialsJSON"`
    // Azure-specific
    ClientID        string            `yaml:"clientID" json:"clientID"`
    ClientSecret    string            `yaml:"clientSecret" json:"clientSecret"`
    TenantID        string            `yaml:"tenantID" json:"tenantID"`
    // Secrets manager
    SecretsManager  SecretsManagerConfig `yaml:"secretsManager" json:"secretsManager"`
}

type SecretsManagerConfig struct {
    Enabled    bool   `yaml:"enabled" json:"enabled"`
    Provider   string `yaml:"provider" json:"provider"`  // aws, gcp, azure, vault
    SecretName string `yaml:"secretName" json:"secretName"`
    Region     string `yaml:"region" json:"region"`       // AWS-specific
    Project    string `yaml:"project" json:"project"`     // GCP-specific
    VaultName  string `yaml:"vaultName" json:"vaultName"` // Azure-specific
    VaultAddr  string `yaml:"vaultAddr" json:"vaultAddr"` // Vault-specific
}

// LoadConfig loads configuration from file with backward compatibility
func LoadConfig(path string) (*Config, error) {
    cfg := NewDefaultConfig()

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, err
    }

    // Migrate old configuration format
    if err := cfg.MigrateLegacyConfig(); err != nil {
        return nil, err
    }

    // Expand environment variables
    if err := cfg.ExpandEnvVars(); err != nil {
        return nil, err
    }

    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    return cfg, nil
}

// MigrateLegacyConfig converts old ECR/GCR config to new format
func (c *Config) MigrateLegacyConfig() error {
    // If new format already exists, skip migration
    if len(c.Registries.Definitions) > 0 {
        return nil
    }

    // Migrate ECR config
    if c.ECR.AccountID != "" {
        c.Registries.Definitions = append(c.Registries.Definitions, RegistryDefinition{
            Name:     "ecr-migrated",
            Type:     "ecr",
            Enabled:  true,
            Priority: 1,
            Config: map[string]interface{}{
                "region":    c.ECR.Region,
                "accountID": c.ECR.AccountID,
            },
            Authentication: AuthConfig{
                Method: "iam-role",
            },
        })
    }

    // Migrate GCR config
    if c.GCR.Project != "" {
        c.Registries.Definitions = append(c.Registries.Definitions, RegistryDefinition{
            Name:     "gcr-migrated",
            Type:     "gcr",
            Enabled:  true,
            Priority: 2,
            Config: map[string]interface{}{
                "project":  c.GCR.Project,
                "location": c.GCR.Location,
            },
            Authentication: AuthConfig{
                Method: "adc",
            },
        })
    }

    return nil
}
```

## 8. Testing Components

### 8.1 Unit Tests

Create comprehensive unit tests for new components:

**pkg/registry/manager_test.go:**
```go
func TestRegistryManager_Initialize(t *testing.T) {
    tests := []struct {
        name    string
        config  *config.RegistryConfig
        wantErr bool
    }{
        {
            name: "successful initialization",
            config: &config.RegistryConfig{
                Definitions: []config.RegistryDefinition{
                    {
                        Name:    "test-ecr",
                        Type:    "ecr",
                        Enabled: true,
                        // ... config
                    },
                },
            },
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            rm, err := registry.NewRegistryManager(tt.config, testLogger)
            if err != nil {
                t.Fatalf("failed to create manager: %v", err)
            }

            err = rm.Initialize(context.Background())
            if (err != nil) != tt.wantErr {
                t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 8.2 Integration Tests

**tests/integration/registry_test.go:**
```go
func TestRegistryIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup
    cfg := loadTestConfig(t)
    svc, err := service.NewService(cfg, testLogger)
    require.NoError(t, err)
    defer svc.Close()

    // Test replication
    t.Run("replicate_ecr_to_harbor", func(t *testing.T) {
        ctx := context.Background()
        err := svc.Replicate(ctx, "test-image:v1", "test-image:v1")
        require.NoError(t, err)
    })
}
```

## 9. Migration Checklist

### 9.1 Development Phase

- [ ] Implement `pkg/registry` package
  - [ ] manager.go
  - [ ] types.go
  - [ ] factory.go
  - [ ] selector.go
  - [ ] auth providers
- [ ] Implement registry clients
  - [ ] Refactor ECR client
  - [ ] Refactor GCR client
  - [ ] New ACR client
  - [ ] New DockerHub client
  - [ ] New Harbor client
  - [ ] New Generic client
- [ ] Update service layer
  - [ ] service/replicate.go
  - [ ] service/tree_replicate.go
- [ ] Update server layer
  - [ ] server/handlers.go
  - [ ] server/types.go
- [ ] Update CLI commands
  - [ ] cmd/replicate.go
  - [ ] cmd/tree-replicate.go
  - [ ] cmd/serve.go
- [ ] Update configuration
  - [ ] config/config.go
  - [ ] Add validation
  - [ ] Add migration logic
- [ ] Write tests
  - [ ] Unit tests (80%+ coverage)
  - [ ] Integration tests
  - [ ] End-to-end tests

### 9.2 Testing Phase

- [ ] Test backward compatibility
- [ ] Test each registry type
- [ ] Test authentication methods
- [ ] Test secrets manager integration
- [ ] Performance testing
- [ ] Load testing
- [ ] Security testing

### 9.3 Documentation Phase

- [ ] Update README.md
- [ ] Update configuration examples
- [ ] Create migration guide
- [ ] Update API documentation
- [ ] Create troubleshooting guide
- [ ] Update Kubernetes manifests

### 9.4 Deployment Phase

- [ ] Staging deployment
- [ ] Production canary deployment
- [ ] Monitor metrics
- [ ] Gradual rollout
- [ ] Full production deployment

---

**Document Version:** 1.0
**Last Updated:** 2025-12-02
**Related:** [Architecture Overview](./registry-support-architecture.md) | [Configuration Schema](./registry-config-schema.md)
