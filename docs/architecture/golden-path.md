# Golden Path Architecture — Freightliner

**Production-Ready Container Registry Replication System**

**Document Version:** 1.0
**Date:** 2025-12-05
**Status:** Architecture Design Document

---

## Executive Summary

This document defines the **Golden Path Architecture** for Freightliner, a high-throughput container registry replication tool built entirely in Go. The architecture follows **Mission Brief Section 3.2** layering principles and implements production-ready patterns for multi-cloud container synchronization across AWS ECR, Google GCR, and generic OCI-compliant registries.

### Key Architectural Principles

1. **Native Go Implementation** — No external tool dependencies (no skopeo, crane, docker CLI)
2. **Single Binary** — One `freightliner` executable with CLI, server, and worker modes
3. **Clear Layering** — Strict separation: cmd → service → client → replication → infrastructure
4. **Goroutine-First Concurrency** — Worker pools, channels, context cancellation
5. **Production Patterns** — Factory, Strategy, Observer, Repository, Worker Pool
6. **Observability-Native** — Prometheus metrics, structured logging, trace IDs

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Architectural Layers](#2-architectural-layers)
3. [Component Architecture](#3-component-architecture)
4. [Design Patterns](#4-design-patterns)
5. [Data Flow Diagrams](#5-data-flow-diagrams)
6. [Concurrency Architecture](#6-concurrency-architecture)
7. [Production Deployment](#7-production-deployment)
8. [Security Architecture](#8-security-architecture)
9. [Monitoring & Observability](#9-monitoring--observability)
10. [Development Guidelines](#10-development-guidelines)

---

## 1. System Overview

### 1.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Freightliner Binary                        │
│                   (Single Executable, Multiple Modes)              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐               │
│  │  CLI Mode   │  │ Server Mode │  │ Worker Mode │               │
│  │   (cobra)   │  │   (HTTP)    │  │  (daemon)   │               │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘               │
│         │                 │                 │                       │
│         └─────────────────┼─────────────────┘                       │
│                           ▼                                         │
│         ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓                   │
│         ┃     SERVICE LAYER (Orchestration)    ┃                   │
│         ┃  - ReplicationService                 ┃                   │
│         ┃  - TreeReplicationService             ┃                   │
│         ┃  - CheckpointService                  ┃                   │
│         ┗━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━┛                   │
│                           ▼                                         │
│         ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓                   │
│         ┃     CLIENT LAYER (Registry Adapters) ┃                   │
│         ┃  - Factory (creates clients)          ┃                   │
│         ┃  - ECR Client (AWS native)            ┃                   │
│         ┃  - GCR Client (GCP native)            ┃                   │
│         ┃  - Generic Client (OCI v2)            ┃                   │
│         ┃  - Harbor/Quay/GHCR/ACR Clients       ┃                   │
│         ┗━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━┛                   │
│                           ▼                                         │
│         ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓                   │
│         ┃   REPLICATION LAYER (Concurrency)    ┃                   │
│         ┃  - WorkerPool (goroutine management) ┃                   │
│         ┃  - Scheduler (cron jobs)              ┃                   │
│         ┃  - PriorityQueue (job ordering)       ┃                   │
│         ┃  - Autoscaler (dynamic sizing)        ┃                   │
│         ┗━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━┛                   │
│                           ▼                                         │
│         ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓                   │
│         ┃    INFRASTRUCTURE LAYER              ┃                   │
│         ┃  - Network (transfer, compression)   ┃                   │
│         ┃  - Security (encryption, KMS, mTLS)  ┃                   │
│         ┃  - Cache (LRU, buffer pools)         ┃                   │
│         ┃  - Metrics (Prometheus)               ┃                   │
│         ┃  - Helper (logging, errors, banner)  ┃                   │
│         ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛                   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.2 Core Capabilities

| Capability | Implementation | Status |
|------------|----------------|--------|
| **Multi-Registry Support** | Native Go clients for ECR, GCR, Harbor, Quay, GHCR, ACR | ✅ Implemented |
| **Concurrent Replication** | Worker pool with configurable goroutines | ✅ Implemented |
| **Resumable Transfers** | Checkpoint-based state management | ✅ Implemented |
| **Encryption** | AES-256-GCM with KMS integration | ✅ Implemented |
| **HTTP API** | RESTful server with job management | ✅ Implemented |
| **Observability** | Prometheus metrics + structured logging | ✅ Implemented |
| **Production Deployment** | Kubernetes-ready with health checks | ✅ Implemented |

---

## 2. Architectural Layers

Following **Mission Brief Section 3.2**, the architecture enforces strict layering:

### 2.1 Layer Hierarchy

```
┌─────────────────────────────────────────────────────────┐
│ LAYER 1: PRESENTATION (cmd/)                           │
│ - Parses CLI flags and configuration                    │
│ - Decides execution mode (CLI/server/worker)            │
│ - Marshals requests to service layer                    │
│ - NO business logic, NO registry API calls              │
└───────────────────────┬─────────────────────────────────┘
                        │ calls
                        ▼
┌─────────────────────────────────────────────────────────┐
│ LAYER 2: SERVICE (pkg/service/)                        │
│ - Orchestrates replication workflows                    │
│ - Manages checkpoints and state                         │
│ - Coordinates worker pools                              │
│ - Handles job lifecycle                                 │
│ - NO direct registry API calls                          │
└───────────────────────┬─────────────────────────────────┘
                        │ uses
                        ▼
┌─────────────────────────────────────────────────────────┐
│ LAYER 3: CLIENT (pkg/client/)                          │
│ - Registry-specific adapters (ECR, GCR, Generic)        │
│ - Factory pattern for client creation                   │
│ - Authentication handling (IAM, OAuth2, Bearer)         │
│ - Manifest and layer operations                         │
│ - Implements RegistryClient interface                   │
└───────────────────────┬─────────────────────────────────┘
                        │ executes via
                        ▼
┌─────────────────────────────────────────────────────────┐
│ LAYER 4: REPLICATION (pkg/replication/)                │
│ - WorkerPool: goroutine management                      │
│ - Scheduler: cron-based job scheduling                  │
│ - Job queueing and priority handling                    │
│ - Progress tracking and metrics                         │
└───────────────────────┬─────────────────────────────────┘
                        │ depends on
                        ▼
┌─────────────────────────────────────────────────────────┐
│ LAYER 5: INFRASTRUCTURE (pkg/*)                        │
│ - network/: Transfer manager, compression, pooling      │
│ - security/: Encryption, KMS, mTLS, signatures          │
│ - cache/: LRU cache, buffer pools                       │
│ - metrics/: Prometheus registry                         │
│ - helper/: Logging, errors, utilities                   │
└─────────────────────────────────────────────────────────┘
```

### 2.2 Layer Responsibilities

**LAYER 1: PRESENTATION (cmd/)**

```go
// cmd/replicate.go
func newReplicateCmd() *cobra.Command {
    // Parse flags → Create config → Call service
    // NO business logic here
    service := service.NewReplicationService(cfg, logger)
    return service.Replicate(ctx, rule)
}
```

**LAYER 2: SERVICE (pkg/service/)**

```go
// pkg/service/replication_service.go
type ReplicationService struct {
    clientFactory *client.Factory
    workerPool    *replication.WorkerPool
    checkpointSvc *CheckpointService
}

func (s *ReplicationService) Replicate(ctx context.Context, rule ReplicationRule) error {
    // 1. Create source/dest clients via factory
    // 2. Submit job to worker pool
    // 3. Track progress and checkpoints
    // 4. Return aggregated results
}
```

**LAYER 3: CLIENT (pkg/client/)**

```go
// pkg/client/factory.go
type Factory struct {
    config *config.Config
    logger log.Logger
}

func (f *Factory) CreateClientForRegistry(ctx context.Context, url string) (interfaces.RegistryClient, error) {
    // Auto-detect registry type from URL
    // Return appropriate native Go client
    // ECR → ecr.NewClient()
    // GCR → gcr.NewClient()
    // Generic → generic.NewClient()
}
```

**LAYER 4: REPLICATION (pkg/replication/)**

```go
// pkg/replication/worker_pool.go
type WorkerPool struct {
    workers   int
    jobQueue  chan WorkerJob
    results   chan JobResult
    stopCtx   context.Context
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        go p.worker(i) // Goroutine per worker
    }
}
```

**LAYER 5: INFRASTRUCTURE (pkg/)**

```go
// pkg/network/transfer_manager.go
type TransferManager struct {
    bufferPool *util.BufferPool
    compressor Compressor
}

// pkg/security/encryption/manager.go
type Manager struct {
    kmsClient KMSClient
    cipher    cipher.AEAD
}
```

### 2.3 Dependency Rules

**✅ ALLOWED:**
- Lower layers MAY NOT depend on higher layers
- Layers communicate via interfaces
- Cross-cutting concerns (logging, metrics) accessible from all layers

**❌ FORBIDDEN:**
- cmd/ calling registry APIs directly
- service/ implementing registry protocols
- client/ managing worker pools
- Circular dependencies between layers

---

## 3. Component Architecture

### 3.1 Client Layer Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Factory                         │
│  (Creates registry-specific clients based on URL/config)   │
└───────────────────┬─────────────────────────────────────────┘
                    │
        ┌───────────┼───────────┬───────────┐
        ▼           ▼           ▼           ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ ECR Client  │ │ GCR Client  │ │Generic Client│ │Harbor Client│
│             │ │             │ │             │ │             │
│ AWS SDK v2  │ │ GCP SDK     │ │ OCI Spec    │ │ Harbor API  │
│ IAM Auth    │ │ OAuth2      │ │ Docker V2   │ │ Robot Auth  │
│ Pagination  │ │ ADC         │ │ Anonymous   │ │ Projects    │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘
        │           │           │           │
        └───────────┴───────────┴───────────┘
                    │
                    ▼
        ┌─────────────────────────────┐
        │  RegistryClient Interface   │
        │  - ListTags()               │
        │  - GetManifest()            │
        │  - PutManifest()            │
        │  - GetBlob()                │
        │  - PutBlob()                │
        │  - DeleteTag()              │
        └─────────────────────────────┘
```

#### 3.1.1 Client Factory Pattern

**Purpose:** Encapsulate client creation logic and provide auto-detection

```go
// pkg/client/factory.go

type Factory struct {
    config *config.Config
    logger log.Logger
}

func (f *Factory) CreateClientForRegistry(ctx context.Context, url string) (interfaces.RegistryClient, error) {
    // Auto-detect based on URL patterns
    switch {
    case strings.Contains(url, ".dkr.ecr."):
        return f.CreateECRClient()
    case strings.Contains(url, "gcr.io") || strings.Contains(url, "pkg.dev"):
        return f.CreateGCRClient()
    case strings.Contains(url, "ghcr.io"):
        return f.CreateGHCRClient()
    case strings.Contains(url, ".azurecr.io"):
        return f.CreateACRClient(registryName, opts)
    default:
        return f.CreateGenericClient(url)
    }
}
```

**Benefits:**
- Single point of client creation
- Consistent authentication handling
- Easy to add new registry types
- Testable via mocks

#### 3.1.2 Registry Client Interface

```go
// pkg/interfaces/registry_client.go

type RegistryClient interface {
    // Manifest operations
    ListTags(ctx context.Context, repository string) ([]string, error)
    GetManifest(ctx context.Context, ref string) (*Manifest, error)
    PutManifest(ctx context.Context, ref string, manifest *Manifest) error
    DeleteManifest(ctx context.Context, ref string) error

    // Blob operations
    GetBlob(ctx context.Context, ref string, digest string) (io.ReadCloser, error)
    PutBlob(ctx context.Context, ref string, digest string, content io.Reader) error
    BlobExists(ctx context.Context, ref string, digest string) (bool, error)

    // Repository operations
    ListRepositories(ctx context.Context) ([]string, error)

    // Lifecycle
    Close() error
}
```

### 3.2 Replication Layer Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Scheduler                              │
│  (Cron-based job scheduling with priorities)               │
└───────────────────┬─────────────────────────────────────────┘
                    │ submits to
                    ▼
┌─────────────────────────────────────────────────────────────┐
│                   Worker Pool                              │
│  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐           │
│  │Worker 1│  │Worker 2│  │Worker 3│  │Worker N│           │
│  │goroutine│ │goroutine│ │goroutine│ │goroutine│          │
│  └───┬────┘  └───┬────┘  └───┬────┘  └───┬────┘           │
│      │           │           │           │                  │
│      └───────────┴───────────┴───────────┘                  │
│                  │                                           │
│                  ▼                                           │
│         ┌─────────────────┐                                 │
│         │   Job Queue     │                                 │
│         │  (buffered chan)│                                 │
│         └─────────────────┘                                 │
│                  │                                           │
│                  ▼                                           │
│         ┌─────────────────┐                                 │
│         │ Results Channel │                                 │
│         │  (buffered chan)│                                 │
│         └─────────────────┘                                 │
└─────────────────────────────────────────────────────────────┘
```

#### 3.2.1 Worker Pool Implementation

```go
// pkg/replication/worker_pool.go

type WorkerPool struct {
    workers       int
    jobQueue      chan WorkerJob
    results       chan JobResult
    waitGroup     sync.WaitGroup
    stopContext   context.Context
    stopFunc      context.CancelFunc
    logger        log.Logger
    closed        atomic.Bool
    stats         *statsCollector
}

// Start spawns N goroutines
func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        workerID := i
        p.waitGroup.Add(1)
        go func() {
            defer p.waitGroup.Done()
            p.worker(workerID)
        }()
    }
}

// Worker goroutine processes jobs from queue
func (p *WorkerPool) worker(id int) {
    for {
        select {
        case <-p.stopContext.Done():
            return
        case job, ok := <-p.jobQueue:
            if !ok {
                return
            }
            p.processJob(id, job)
        }
    }
}
```

**Key Features:**
- Configurable worker count (default: `runtime.NumCPU()`)
- Graceful shutdown via context cancellation
- Job result aggregation
- Statistics collection (success rate, duration, throughput)

#### 3.2.2 Scheduler Component

```go
// pkg/replication/scheduler.go

type Scheduler struct {
    jobs              map[string]*Job
    mutex             sync.RWMutex
    workerPool        *WorkerPool
    replicationSvc    ReplicationService
    cronParser        cron.Parser
}

func (s *Scheduler) AddJob(rule ReplicationRule) error {
    // Parse cron expression
    schedule, err := s.cronParser.Parse(rule.Schedule)

    // Calculate next run time
    nextRun := schedule.Next(time.Now())

    // Register job
    s.jobs[id] = &Job{
        Rule:    rule,
        NextRun: nextRun,
        Running: false,
    }
}

func (s *Scheduler) run() {
    ticker := time.NewTicker(1 * time.Minute)
    for {
        select {
        case <-ticker.C:
            s.checkJobs()
        case <-s.ctx.Done():
            return
        }
    }
}
```

### 3.3 Security Components

```
┌─────────────────────────────────────────────────────────────┐
│                   Security Manager                         │
└───────┬────────────────┬────────────────┬───────────────────┘
        │                │                │
        ▼                ▼                ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ Encryption   │ │    mTLS      │ │   Cosign     │
│  Manager     │ │   Manager    │ │  Verifier    │
│              │ │              │ │              │
│ AES-256-GCM  │ │ Cert Store   │ │ Signatures   │
│ KMS (AWS/GCP)│ │ Client Auth  │ │ SBOM         │
│ Key Rotation │ │ Server Auth  │ │ Attestation  │
└──────────────┘ └──────────────┘ └──────────────┘
```

#### 3.3.1 Encryption Strategy Pattern

```go
// pkg/security/encryption/manager.go

type Manager struct {
    kmsClient    KMSClient
    cipher       cipher.AEAD
    keyCache     *cache.LRUCache
    logger       log.Logger
}

// Strategy interface for different encryption backends
type EncryptionStrategy interface {
    Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
    Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
    RotateKey(ctx context.Context) error
}

// AWS KMS strategy
type AWSKMSStrategy struct {
    client *kms.Client
    keyID  string
}

// GCP KMS strategy
type GCPKMSStrategy struct {
    client *kmspb.KeyManagementServiceClient
    keyName string
}
```

---

## 4. Design Patterns

### 4.1 Factory Pattern (Client Creation)

**Purpose:** Encapsulate complex client instantiation logic

```go
// pkg/client/factory/factory.go

type ClientFactory interface {
    CreateClient(registryType string, config *Config) (RegistryClient, error)
}

type StandardClientFactory struct {
    logger log.Logger
    cache  *cache.LRUCache
}

func (f *StandardClientFactory) CreateClient(registryType string, config *Config) (RegistryClient, error) {
    switch registryType {
    case "ecr":
        return ecr.NewClient(ecr.ClientOptions{
            Region:    config.Region,
            AccountID: config.AccountID,
            Logger:    f.logger,
        })
    case "gcr":
        return gcr.NewClient(gcr.ClientOptions{
            Project:  config.Project,
            Location: config.Location,
            Logger:   f.logger,
        })
    default:
        return generic.NewClient(generic.ClientOptions{
            RegistryURL: config.Endpoint,
            Logger:      f.logger,
        })
    }
}
```

**Benefits:**
- Single Responsibility: Factory only creates clients
- Open/Closed: Easy to add new registry types
- Dependency Injection: Logger and config injected
- Testability: Mock factory for unit tests

### 4.2 Strategy Pattern (Authentication)

**Purpose:** Interchangeable authentication methods per registry

```go
// pkg/client/auth/strategy.go

type AuthStrategy interface {
    Authenticate(ctx context.Context) (*Credentials, error)
    RefreshToken(ctx context.Context) error
}

// AWS IAM authentication
type IAMAuthStrategy struct {
    region    string
    accountID string
    stsClient *sts.Client
}

// GCP OAuth2 authentication
type OAuth2Strategy struct {
    credentials *google.Credentials
    tokenSource oauth2.TokenSource
}

// Basic authentication
type BasicAuthStrategy struct {
    username string
    password string
}

// Registry client uses strategy
type GenericClient struct {
    authStrategy AuthStrategy
    httpClient   *http.Client
}

func (c *GenericClient) request(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Apply authentication via strategy
    creds, err := c.authStrategy.Authenticate(ctx)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", creds.Token)
    return c.httpClient.Do(req)
}
```

### 4.3 Worker Pool Pattern (Concurrency)

**Purpose:** Manage goroutine lifecycle and job distribution

```go
// pkg/replication/worker_pool.go

type WorkerPool struct {
    workers   int
    jobQueue  chan WorkerJob
    results   chan JobResult
    wg        sync.WaitGroup
    ctx       context.Context
    cancel    context.CancelFunc
}

// Submit adds job to queue (non-blocking with timeout)
func (p *WorkerPool) Submit(id string, task TaskFunc) error {
    select {
    case p.jobQueue <- WorkerJob{ID: id, Task: task}:
        return nil
    case <-time.After(30 * time.Second):
        return errors.New("job queue full")
    }
}

// Worker processes jobs from queue
func (p *WorkerPool) worker(id int) {
    defer p.wg.Done()
    for {
        select {
        case <-p.ctx.Done():
            return
        case job := <-p.jobQueue:
            result := p.processJob(job)
            p.results <- result
        }
    }
}
```

### 4.4 Repository Pattern (Data Abstraction)

**Purpose:** Abstract checkpoint and state persistence

```go
// pkg/tree/checkpoint/repository.go

type CheckpointRepository interface {
    Save(ctx context.Context, checkpoint *Checkpoint) error
    Load(ctx context.Context, id string) (*Checkpoint, error)
    List(ctx context.Context) ([]*Checkpoint, error)
    Delete(ctx context.Context, id string) error
}

// File-based implementation
type FileCheckpointRepository struct {
    baseDir string
    logger  log.Logger
}

// S3-based implementation (future)
type S3CheckpointRepository struct {
    bucket  string
    s3Client *s3.Client
}
```

### 4.5 Observer Pattern (Progress Tracking)

**Purpose:** Notify subscribers of replication progress

```go
// pkg/replication/observer.go

type ProgressObserver interface {
    OnProgress(event *ProgressEvent)
    OnComplete(event *CompletionEvent)
    OnError(event *ErrorEvent)
}

type MetricsObserver struct {
    registry *metrics.Registry
}

func (o *MetricsObserver) OnProgress(event *ProgressEvent) {
    o.registry.RecordTransfer(event.BytesTransferred)
}

type ReplicationEngine struct {
    observers []ProgressObserver
}

func (e *ReplicationEngine) notifyProgress(event *ProgressEvent) {
    for _, observer := range e.observers {
        observer.OnProgress(event)
    }
}
```

---

## 5. Data Flow Diagrams

### 5.1 Replication Request Flow

```
┌─────────┐
│  User   │
│ (CLI)   │
└────┬────┘
     │ freightliner replicate \
     │   --source ecr://account.dkr.ecr.us-east-1.amazonaws.com/app \
     │   --dest gcr://project/app \
     │   --workers 10
     ▼
┌─────────────────────────────────────────────────────────┐
│  cmd/replicate.go (Presentation Layer)                 │
│  1. Parse flags                                         │
│  2. Load config from file + env vars                    │
│  3. Create logger with structured output                │
│  4. Instantiate ReplicationService                      │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  pkg/service/replication_service.go (Service Layer)    │
│  1. Create source client via factory                    │
│  2. Create destination client via factory               │
│  3. Validate connectivity and permissions               │
│  4. Create ReplicationRule from config                  │
│  5. Submit job to WorkerPool                            │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  pkg/client/factory.go (Client Layer)                  │
│  1. Auto-detect registry type from URL                  │
│  2. Load authentication credentials                     │
│  3. Create registry-specific client:                    │
│     - ECR: AWS SDK v2 with IAM                          │
│     - GCR: GCP SDK with OAuth2                          │
│  4. Return RegistryClient interface                     │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  pkg/replication/worker_pool.go (Replication Layer)    │
│  1. Receive job from queue                              │
│  2. Assign to available worker goroutine                │
│  3. Worker executes replication task:                   │
│     a. List tags from source                            │
│     b. For each tag:                                    │
│        - Get manifest                                   │
│        - Get layer blobs                                │
│        - Put blobs to destination                       │
│        - Put manifest to destination                    │
│  4. Track progress and metrics                          │
│  5. Send result to results channel                      │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  Infrastructure Layer (pkg/network, pkg/security)      │
│  - Transfer manager: Stream blobs with compression      │
│  - Encryption manager: Encrypt layers with KMS          │
│  - Cache: Deduplicate layers via SHA256 lookup          │
│  - Metrics: Record bytes transferred, duration          │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  Results Aggregation                                    │
│  1. Collect results from workers                        │
│  2. Update checkpoint state                             │
│  3. Log completion with trace ID                        │
│  4. Return success/failure to user                      │
└─────────────────────────────────────────────────────────┘
```

### 5.2 HTTP API Request Flow

```
┌─────────┐
│  HTTP   │
│ Client  │
└────┬────┘
     │ POST /api/v1/replicate
     │ {
     │   "source": "ecr://...",
     │   "destination": "gcr://...",
     │   "workers": 10
     │ }
     ▼
┌─────────────────────────────────────────────────────────┐
│  pkg/server/handlers.go (HTTP Server)                  │
│  Middleware Stack:                                      │
│  1. Logging middleware (request ID)                     │
│  2. Metrics middleware (increment counters)             │
│  3. Recovery middleware (panic handler)                 │
│  4. CORS middleware (origin validation)                 │
│  5. Auth middleware (API key check)                     │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  replicateHandler                                       │
│  1. Parse JSON request body                             │
│  2. Validate request (source, dest, tags)               │
│  3. Create ReplicateJob                                 │
│  4. Add job to JobManager                               │
│  5. Submit job to WorkerPool                            │
│  6. Return job ID + status (202 Accepted)               │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  pkg/server/jobs.go (Job Manager)                      │
│  1. Register job with unique ID                         │
│  2. Track job state (pending → running → completed)     │
│  3. Store job metadata (timestamps, params)             │
│  4. Provide job query API                               │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  Background Processing (same as CLI flow)               │
│  - Worker pool executes job                             │
│  - Progress tracked via JobManager                      │
│  - Results available via GET /api/v1/jobs/{id}          │
└─────────────────────────────────────────────────────────┘
```

### 5.3 Authentication Flow

```
┌─────────────────────────────────────────────────────────┐
│  Client Factory: CreateClientForRegistry()             │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
         ┌────────────┴────────────┐
         │  Registry Type?          │
         └────────────┬────────────┘
                      │
      ┌───────────────┼───────────────┐
      │               │               │
      ▼               ▼               ▼
┌──────────┐   ┌──────────┐   ┌──────────┐
│   ECR    │   │   GCR    │   │ Generic  │
└────┬─────┘   └────┬─────┘   └────┬─────┘
     │              │              │
     ▼              ▼              ▼
┌──────────────────────────────────────────┐
│ AWS IAM Authentication                   │
│ 1. Load credentials from:                │
│    - Environment (AWS_ACCESS_KEY_ID)     │
│    - Config file (~/.aws/credentials)    │
│    - IAM role (EC2/ECS instance)         │
│ 2. Assume role if cross-account          │
│ 3. Get ECR authorization token           │
│ 4. Token valid for 12 hours              │
└──────────────────────────────────────────┘

┌──────────────────────────────────────────┐
│ GCP OAuth2 Authentication                │
│ 1. Load credentials from:                │
│    - Service account key file            │
│    - Application Default Credentials     │
│    - Compute Engine metadata             │
│ 2. Create OAuth2 token source            │
│ 3. Get access token                      │
│ 4. Token auto-refreshed on expiry        │
└──────────────────────────────────────────┘

┌──────────────────────────────────────────┐
│ Generic Registry Authentication          │
│ 1. Try Docker config.json first          │
│ 2. Fall back to Basic auth               │
│ 3. Or Bearer token if provided           │
│ 4. Or anonymous if public registry       │
└──────────────────────────────────────────┘
```

---

## 6. Concurrency Architecture

### 6.1 Goroutine Management

```
┌─────────────────────────────────────────────────────────┐
│               Main Goroutine                            │
│  - CLI/Server initialization                            │
│  - Configuration loading                                │
│  - Signal handling (SIGTERM, SIGINT)                    │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│           WorkerPool.Start()                            │
│  Spawns N worker goroutines                             │
└─────────────────────┬───────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┬─────────────┐
        ▼             ▼             ▼             ▼
┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐
│  Worker 1  │  │  Worker 2  │  │  Worker 3  │  │  Worker N  │
│ goroutine  │  │ goroutine  │  │ goroutine  │  │ goroutine  │
└─────┬──────┘  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘
      │               │               │               │
      │ Each worker reads from job queue              │
      │ Processes job (transfer layers)               │
      │ Writes result to results channel              │
      │                                                 │
      └─────────────────┬───────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│           Results Aggregator Goroutine                  │
│  - Reads from results channel                           │
│  - Updates job status                                   │
│  - Records metrics                                      │
│  - Logs completion                                      │
└─────────────────────────────────────────────────────────┘

Additional goroutines:
┌─────────────────────────────────────────────────────────┐
│  - HTTP server goroutine (if serve mode)                │
│  - Scheduler goroutine (if cron jobs enabled)           │
│  - Metrics collector goroutine (periodic scrape)        │
│  - Checkpoint saver goroutine (periodic flush)          │
└─────────────────────────────────────────────────────────┘
```

### 6.2 Channel Communication

```go
// pkg/replication/worker_pool.go

// Job submission channel (buffered for burst handling)
jobQueue := make(chan WorkerJob, workerCount * 20)

// Results channel (buffered to prevent blocking)
results := make(chan JobResult, workerCount * 20)

// Submit job (with timeout to prevent deadlock)
func (p *WorkerPool) Submit(id string, task TaskFunc) error {
    select {
    case p.jobQueue <- WorkerJob{ID: id, Task: task}:
        return nil
    case <-time.After(30 * time.Second):
        return errors.New("job queue full")
    case <-p.ctx.Done():
        return errors.New("worker pool stopped")
    }
}

// Worker reads from job queue
func (p *WorkerPool) worker(id int) {
    for {
        select {
        case <-p.ctx.Done():
            return // Graceful shutdown
        case job, ok := <-p.jobQueue:
            if !ok {
                return // Queue closed
            }
            result := p.processJob(job)

            select {
            case p.results <- result:
                // Result sent
            case <-time.After(5 * time.Second):
                // Results channel full, log and discard
                p.logger.Warn("Results channel timeout, discarding result")
            }
        }
    }
}
```

### 6.3 Context Cancellation

```go
// Hierarchical context propagation

// Root context (from main)
rootCtx, cancel := context.WithCancel(context.Background())
defer cancel()

// Service context (with timeout)
serviceCtx, serviceCancel := context.WithTimeout(rootCtx, 30*time.Minute)
defer serviceCancel()

// Worker pool context (inherits from service)
poolCtx, poolCancel := context.WithCancel(serviceCtx)
defer poolCancel()

// Job context (per-job cancellation)
jobCtx, jobCancel := context.WithTimeout(poolCtx, 5*time.Minute)
defer jobCancel()

// Propagation:
// cancel() → serviceCancel() → poolCancel() → jobCancel()
// Any parent cancellation propagates to children
```

### 6.4 Synchronization Primitives

```go
// sync.WaitGroup for worker lifecycle
var wg sync.WaitGroup
for i := 0; i < workers; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        worker(id)
    }(i)
}
wg.Wait() // Block until all workers complete

// sync.Mutex for shared state
var mu sync.Mutex
func updateJobStatus(id string, status JobStatus) {
    mu.Lock()
    defer mu.Unlock()
    jobs[id].Status = status
}

// sync.RWMutex for read-heavy workloads
var rwmu sync.RWMutex
func getJobStatus(id string) JobStatus {
    rwmu.RLock()
    defer rwmu.RUnlock()
    return jobs[id].Status
}

// atomic operations for counters
var completed atomic.Int64
completed.Add(1)
count := completed.Load()
```

---

## 7. Production Deployment

### 7.1 Kubernetes Deployment

```yaml
# deployments/kubernetes/freightliner-deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: freightliner
  namespace: replication
spec:
  replicas: 3
  selector:
    matchLabels:
      app: freightliner
  template:
    metadata:
      labels:
        app: freightliner
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "2112"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: freightliner
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000

      containers:
      - name: freightliner
        image: freightliner:v1.0.0
        imagePullPolicy: IfNotPresent

        command: ["freightliner", "serve"]
        args:
          - --port=8080
          - --workers=20
          - --log-level=info
          - --metrics-port=2112

        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: metrics
          containerPort: 2112
          protocol: TCP

        env:
        - name: LOG_LEVEL
          value: "info"
        - name: METRICS_ENABLED
          value: "true"
        - name: AWS_REGION
          value: "us-east-1"

        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 2Gi

        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2

        volumeMounts:
        - name: config
          mountPath: /etc/freightliner
          readOnly: true
        - name: checkpoints
          mountPath: /var/lib/freightliner/checkpoints

      volumes:
      - name: config
        configMap:
          name: freightliner-config
      - name: checkpoints
        persistentVolumeClaim:
          claimName: freightliner-checkpoints
```

### 7.2 Service Configuration

```yaml
# deployments/kubernetes/freightliner-service.yaml

apiVersion: v1
kind: Service
metadata:
  name: freightliner
  namespace: replication
  labels:
    app: freightliner
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  - name: metrics
    port: 2112
    targetPort: 2112
    protocol: TCP
  selector:
    app: freightliner

---
apiVersion: v1
kind: Service
metadata:
  name: freightliner-metrics
  namespace: replication
  labels:
    app: freightliner
spec:
  type: ClusterIP
  ports:
  - name: metrics
    port: 2112
    targetPort: 2112
  selector:
    app: freightliner
```

### 7.3 ConfigMap

```yaml
# deployments/kubernetes/freightliner-configmap.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: freightliner-config
  namespace: replication
data:
  config.yaml: |
    log_level: info
    server:
      port: 8080
      read_timeout: 30s
      write_timeout: 30s
      max_header_bytes: 1048576

    workers:
      count: 20
      queue_size: 1000

    metrics:
      enabled: true
      port: 2112
      path: /metrics

    registries:
      default_source: ecr
      default_destination: gcr

      registries:
        - name: prod-ecr
          type: ecr
          region: us-east-1
          account_id: "123456789012"

        - name: prod-gcr
          type: gcr
          project: my-gcp-project
          location: us-central1
```

### 7.4 RBAC Configuration

```yaml
# deployments/kubernetes/freightliner-rbac.yaml

apiVersion: v1
kind: ServiceAccount
metadata:
  name: freightliner
  namespace: replication

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: freightliner
  namespace: replication
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create", "patch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: freightliner
  namespace: replication
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: freightliner
subjects:
- kind: ServiceAccount
  name: freightliner
  namespace: replication
```

---

## 8. Security Architecture

### 8.1 Authentication & Authorization

```
┌─────────────────────────────────────────────────────────┐
│             Authentication Flow                         │
└─────────────────────┬───────────────────────────────────┘
                      │
         ┌────────────┴────────────┐
         │  Client Request         │
         └────────────┬────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │  Auth Middleware           │
         │  1. Extract API key        │
         │  2. Validate signature     │
         │  3. Check rate limits      │
         └────────────┬────────────────┘
                      │
              ┌───────┴───────┐
              │  Valid?        │
              └───────┬───────┘
                 Yes  │   No
         ┌────────────┴────────────┐
         ▼                         ▼
┌────────────────┐      ┌────────────────┐
│  Proceed to    │      │  Return 401    │
│  Handler       │      │  Unauthorized  │
└────────────────┘      └────────────────┘
```

### 8.2 Encryption Architecture

```
┌─────────────────────────────────────────────────────────┐
│          Layer Encryption with KMS                      │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │  1. Generate data key      │
         │     via KMS                │
         └────────────┬────────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │  2. Encrypt layer data     │
         │     with data key          │
         │     (AES-256-GCM)          │
         └────────────┬────────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │  3. Encrypt data key       │
         │     with KMS master key    │
         └────────────┬────────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │  4. Store encrypted layer  │
         │     + encrypted data key   │
         └────────────────────────────┘
```

**Envelope Encryption:**
1. **Data Key Generation**: Request data key from KMS (AWS/GCP)
2. **Data Encryption**: Encrypt layer blob with data key (AES-256-GCM)
3. **Key Encryption**: Encrypt data key with KMS master key
4. **Storage**: Store encrypted blob + encrypted data key

**Decryption:**
1. Retrieve encrypted blob + encrypted data key
2. Decrypt data key via KMS
3. Decrypt blob with data key
4. Cache decrypted data key for performance

### 8.3 Secrets Management

```go
// pkg/secrets/manager.go

type SecretsManager interface {
    GetSecret(ctx context.Context, key string) (string, error)
    PutSecret(ctx context.Context, key, value string) error
    DeleteSecret(ctx context.Context, key string) error
}

// AWS Secrets Manager implementation
type AWSSecretsManager struct {
    client *secretsmanager.Client
    cache  *cache.LRUCache
}

// GCP Secret Manager implementation
type GCPSecretsManager struct {
    client *secretmanagerpb.SecretManagerServiceClient
    cache  *cache.LRUCache
}

// Kubernetes Secrets implementation
type K8sSecretsManager struct {
    clientset kubernetes.Interface
    namespace string
}
```

---

## 9. Monitoring & Observability

### 9.1 Prometheus Metrics

```go
// pkg/metrics/registry.go

type Registry struct {
    // HTTP metrics
    httpRequestsTotal    *prometheus.CounterVec      // Total HTTP requests
    httpRequestDuration  *prometheus.HistogramVec    // Request duration
    httpRequestsInFlight prometheus.Gauge            // Active requests

    // Replication metrics
    replicationTotal       *prometheus.CounterVec    // Total replications
    replicationDuration    *prometheus.HistogramVec  // Replication duration
    replicationBytesTotal  *prometheus.CounterVec    // Bytes transferred
    replicationLayersTotal *prometheus.CounterVec    // Layers transferred

    // Worker pool metrics
    workerPoolSize        prometheus.Gauge           // Active workers
    workerPoolQueueDepth  prometheus.Gauge           // Jobs in queue
    workerPoolJobDuration *prometheus.HistogramVec   // Job duration

    // System metrics
    memoryUsage       prometheus.Gauge              // Memory usage (bytes)
    goroutineCount    prometheus.Gauge              // Active goroutines
    gcDuration        prometheus.Summary             // GC pause time
}

// Metric labels
// - http_method: GET, POST, PUT, DELETE
// - http_status: 200, 404, 500, etc.
// - registry_type: ecr, gcr, docker, harbor
// - operation: list_tags, get_manifest, put_blob
// - status: success, failure
```

### 9.2 Structured Logging

```go
// pkg/helper/log/structured_logger.go

type StructuredLogger struct {
    level      Level
    fields     map[string]interface{}
    traceID    string
    spanID     string
    component  string
    output     io.Writer
}

// Log format (JSON)
{
  "timestamp": "2025-12-05T10:15:30.123Z",
  "level": "info",
  "message": "replication completed",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "span_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "component": "replication",
  "operation": "replicate_repository",
  "source": "ecr://123456789012.dkr.ecr.us-east-1.amazonaws.com/app",
  "destination": "gcr://my-project/app",
  "tags_replicated": 5,
  "bytes_transferred": 1073741824,
  "duration_ms": 45123,
  "worker_id": 3,
  "error": null
}
```

### 9.3 Health Check Endpoints

```go
// pkg/server/health.go

// GET /health - Basic health check
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
        "version": version.Version,
    })
}

// GET /ready - Readiness probe (checks dependencies)
func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
    checks := []HealthCheck{
        s.checkWorkerPool(),
        s.checkRegistryConnectivity(),
        s.checkKMSConnectivity(),
    }

    allHealthy := true
    for _, check := range checks {
        if !check.Healthy {
            allHealthy = false
        }
    }

    status := http.StatusOK
    if !allHealthy {
        status = http.StatusServiceUnavailable
    }

    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": allHealthy,
        "checks": checks,
    })
}

// GET /live - Liveness probe (application responsive)
func (s *Server) liveHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "alive",
    })
}
```

---

## 10. Development Guidelines

### 10.1 Code Organization Principles

**DO:**
- ✅ Place code in appropriate layer (cmd → service → client → replication → infra)
- ✅ Use interfaces for abstraction and testability
- ✅ Inject dependencies via constructors
- ✅ Return errors, don't panic
- ✅ Use context.Context for cancellation
- ✅ Add structured logging with trace IDs
- ✅ Expose Prometheus metrics
- ✅ Write unit tests with table-driven tests
- ✅ Use go-containerregistry for registry operations

**DON'T:**
- ❌ Call registry APIs directly from cmd/
- ❌ Implement business logic in client/
- ❌ Use global variables
- ❌ Ignore errors
- ❌ Block indefinitely without context
- ❌ Shell out to external tools (docker, skopeo, crane)
- ❌ Hard-code configuration values
- ❌ Use naked returns
- ❌ Create circular dependencies

### 10.2 Adding a New Registry Client

```go
// 1. Define client structure
// pkg/client/newregistry/client.go

package newregistry

type Client struct {
    registryURL string
    httpClient  *http.Client
    authToken   string
    logger      log.Logger
}

// 2. Implement RegistryClient interface
func (c *Client) ListTags(ctx context.Context, repository string) ([]string, error) {
    // Implementation
}

func (c *Client) GetManifest(ctx context.Context, ref string) (*Manifest, error) {
    // Implementation
}

// ... implement all interface methods

// 3. Add to factory
// pkg/client/factory.go

func (f *Factory) CreateNewRegistryClient(opts NewRegistryClientOptions) (interfaces.RegistryClient, error) {
    return newregistry.NewClient(opts)
}

// 4. Update factory auto-detection
func (f *Factory) CreateClientForRegistry(ctx context.Context, url string) (interfaces.RegistryClient, error) {
    if strings.Contains(url, "newregistry.io") {
        return f.CreateNewRegistryClient(opts)
    }
    // ...
}

// 5. Add tests
// pkg/client/newregistry/client_test.go

func TestClient_ListTags(t *testing.T) {
    // Table-driven tests
}
```

### 10.3 Testing Strategy

```go
// Unit tests: Test individual components in isolation
// tests/pkg/client/ecr/client_test.go

func TestECRClient_ListTags(t *testing.T) {
    tests := []struct {
        name       string
        repository string
        want       []string
        wantErr    bool
    }{
        {
            name:       "successful listing",
            repository: "test-repo",
            want:       []string{"v1.0.0", "v1.0.1"},
            wantErr:    false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := newMockECRClient(t)
            got, err := client.ListTags(context.Background(), tt.repository)

            if (err != nil) != tt.wantErr {
                t.Errorf("ListTags() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ListTags() = %v, want %v", got, tt.want)
            }
        })
    }
}

// Integration tests: Test against real registries
// tests/integration/ecr_test.go

func TestECRIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup
    client := setupECRClient(t)
    defer client.Close()

    // Test
    tags, err := client.ListTags(context.Background(), "test-repo")
    require.NoError(t, err)
    assert.NotEmpty(t, tags)
}
```

### 10.4 Performance Optimization

**Buffer Pools:**
```go
// pkg/helper/util/buffer_pool.go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func GetBuffer() *bytes.Buffer {
    buf := bufferPool.Get().(*bytes.Buffer)
    buf.Reset()
    return buf
}

func PutBuffer(buf *bytes.Buffer) {
    bufferPool.Put(buf)
}
```

**Connection Pooling:**
```go
// pkg/network/connection_pool.go
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
```

**Caching:**
```go
// pkg/cache/lru_cache.go
cache := cache.NewLRUCache(1000) // 1000 entries
cache.Set(key, value, 5*time.Minute)
if value, ok := cache.Get(key); ok {
    // Use cached value
}
```

---

## Conclusion

This Golden Path Architecture provides a **production-ready foundation** for Freightliner's multi-cloud container registry replication system. Key takeaways:

1. **Native Go Implementation** — No external tool dependencies
2. **Clear Layering** — Strict separation of concerns
3. **Production Patterns** — Factory, Strategy, Worker Pool, Observer
4. **Observability First** — Metrics, logging, tracing built-in
5. **Kubernetes-Ready** — Health checks, graceful shutdown, scalability

**Next Steps:**
1. Review component implementations against this architecture
2. Identify architectural violations and refactor
3. Add missing components (security, monitoring enhancements)
4. Validate production deployment patterns
5. Document operational procedures

---

**Document Metadata:**
- **Version:** 1.0
- **Date:** 2025-12-05
- **Author:** System Architecture Agent
- **Status:** Active
- **Review Cycle:** Quarterly
