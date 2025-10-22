# Go Microservice Development Skill

Expert skill for developing production-grade Go microservices following Freightliner's patterns.

## Architecture Patterns

### Interface-Driven Design
```go
// Define interfaces where they're used
package service

type RegistryClient interface {
    GetRepository(ctx context.Context, name string) (*Repository, error)
}

type ReplicationService struct {
    client RegistryClient  // Dependency injection
    logger log.Logger
}

func NewReplicationService(client RegistryClient, logger log.Logger) *ReplicationService {
    return &ReplicationService{
        client: client,
        logger: logger,
    }
}
```

### Composition Over Inheritance
```go
// Embed base functionality
type ECRClient struct {
    *BaseClient  // Embedded base
    ecrService  *ecr.Client
}

// Methods automatically available
func (c *ECRClient) DoSomething() {
    c.BaseClient.CommonMethod()  // From embedded type
}
```

### Options Pattern
```go
type ServerOptions struct {
    Port        int
    MetricsPort int
    Logger      log.Logger
    Metrics     *metrics.Registry
}

func NewServer(opts ServerOptions) (*Server, error) {
    // Set defaults
    if opts.Port == 0 {
        opts.Port = 8080
    }

    // Validate
    if opts.Logger == nil {
        return nil, errors.New("logger is required")
    }

    return &Server{
        port:    opts.Port,
        logger:  opts.Logger,
        metrics: opts.Metrics,
    }, nil
}
```

### Worker Pool Pattern
```go
type WorkerPool struct {
    workers   int
    jobs      chan Job
    results   chan Result
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    return &WorkerPool{
        workers: workers,
        jobs:    make(chan Job, workers*2),
        results: make(chan Result, workers*2),
        ctx:     ctx,
        cancel:  cancel,
    }
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker()
    }
}

func (p *WorkerPool) worker() {
    defer p.wg.Done()
    for {
        select {
        case job := <-p.jobs:
            result := job.Execute()
            p.results <- result
        case <-p.ctx.Done():
            return
        }
    }
}
```

## Error Handling

### Error Wrapping with Context
```go
import "freightliner/pkg/helper/errors"

func doSomething() error {
    data, err := fetchData()
    if err != nil {
        return errors.Wrap(err, "failed to fetch data")
    }

    if err := processData(data); err != nil {
        return errors.Wrapf(err, "failed to process data: %s", data.ID)
    }

    return nil
}
```

### Domain-Specific Errors
```go
// Use helper/errors package
return errors.NotFoundf("repository %s not found", name)
return errors.Unauthorizedf("invalid credentials for %s", registry)
return errors.InvalidArgumentf("port must be between 1 and 65535, got %d", port)
```

### Error Checking Patterns
```go
// Always check errors
result, err := doOperation()
if err != nil {
    logger.Error("operation failed", err)
    return errors.Wrap(err, "operation failed")
}

// Use errors.As for type checking
var authErr *AuthError
if errors.As(err, &authErr) {
    // Handle authentication error specifically
}
```

## Context Usage

### Always Accept Context
```go
func ReplicateImage(ctx context.Context, src, dst string) error {
    // Check context before expensive operations
    if err := ctx.Err(); err != nil {
        return err
    }

    // Pass context to all downstream calls
    img, err := fetchImage(ctx, src)
    if err != nil {
        return err
    }

    return pushImage(ctx, dst, img)
}
```

### Context with Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := longRunningOperation(ctx)
```

### Context with Values (Tracing)
```go
// Add trace ID to context
ctx = context.WithValue(ctx, "trace_id", traceID)
ctx = context.WithValue(ctx, "span_id", spanID)

// Extract in logging
logger := logger.WithContext(ctx)
```

## Structured Logging

### Using the Logger Interface
```go
import "freightliner/pkg/helper/log"

// Create logger
logger := log.NewBasicLogger(log.InfoLevel)

// Or structured logger for production
logger := log.NewStructuredLogger(log.InfoLevel, os.Stdout)

// Log with fields
logger.WithFields(map[string]interface{}{
    "operation": "replicate",
    "source":    source,
    "dest":      dest,
    "duration":  duration,
}).Info("Replication completed")

// Log errors with context
logger.WithError(err).Error("Replication failed")

// Chain field addition
logger.
    WithField("user_id", userID).
    WithField("request_id", reqID).
    Info("Processing request")
```

## Metrics Collection

### Prometheus Metrics
```go
import "freightliner/pkg/metrics"

// Create registry
registry := metrics.NewRegistry()

// Record metrics
registry.RecordReplication(source, dest, "success", duration, bytesCopied)

// Record HTTP metrics
registry.RecordHTTPRequest(method, path, statusCode, duration)

// Increment counters
registry.IncrementReplicationTotal("ecr", "gcr", "success")

// Set gauges
registry.SetMemoryUsage(float64(m.Alloc))
```

## Testing Patterns

### Table-Driven Tests
```go
func TestReplication(t *testing.T) {
    tests := []struct {
        name        string
        source      string
        dest        string
        wantErr     bool
        expectedErr error
    }{
        {
            name:    "successful replication",
            source:  "ecr/repo",
            dest:    "gcr/repo",
            wantErr: false,
        },
        {
            name:        "source not found",
            source:      "ecr/nonexistent",
            dest:        "gcr/repo",
            wantErr:     true,
            expectedErr: errors.ErrNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Replicate(context.Background(), tt.source, tt.dest)

            if tt.wantErr {
                require.Error(t, err)
                if tt.expectedErr != nil {
                    require.ErrorIs(t, err, tt.expectedErr)
                }
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Using Mocks
```go
// Use gomock or create interface mocks
mockClient := &MockRegistryClient{}
mockClient.On("GetRepository", mock.Anything, "test-repo").
    Return(&Repository{Name: "test-repo"}, nil)

service := NewReplicationService(mockClient, logger)
err := service.Replicate(ctx, "test-repo", "dest-repo")

mockClient.AssertExpectations(t)
```

## HTTP Server Best Practices

### Middleware Stack
```go
server := NewServer(opts)

// Add middleware in order
server.AddMiddleware(LoggingMiddleware)
server.AddMiddleware(MetricsMiddleware)
server.AddMiddleware(RecoveryMiddleware)
server.AddMiddleware(CORSMiddleware)
server.AddMiddleware(AuthMiddleware)
```

### Graceful Shutdown
```go
server := &http.Server{
    Addr:    ":8080",
    Handler: router,
}

// Start server
go func() {
    if err := server.ListenAndServe(); err != http.ErrServerClosed {
        logger.Error("server error", err)
    }
}()

// Wait for shutdown signal
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
<-sigCh

// Graceful shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    logger.Error("shutdown error", err)
}
```

## Resource Management

### Always Use defer for Cleanup
```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()  // Guaranteed cleanup

    // Process file
    return nil
}
```

### Channel Cleanup
```go
func worker(jobs <-chan Job, results chan<- Result, done <-chan struct{}) {
    defer close(results)  // Close when done

    for {
        select {
        case job := <-jobs:
            results <- job.Execute()
        case <-done:
            return
        }
    }
}
```

## Package Organization

Follow Freightliner conventions:
```
pkg/
├── myfeature/
│   ├── service.go          # Main service implementation
│   ├── service_test.go     # Tests
│   ├── interfaces.go       # Local interfaces if needed
│   ├── types.go            # Data types
│   ├── errors.go           # Feature-specific errors
│   └── README.md           # Package documentation
```
