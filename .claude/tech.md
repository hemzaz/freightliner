# Technical Standards and Architecture - Freightliner

## Technology Stack

### Core Language and Runtime
- **Go 1.24.1**: Primary programming language for performance and cloud-native capabilities
- **Cobra CLI Framework**: Command-line interface framework for consistent UX
- **Context-Based Architecture**: All operations use context.Context for cancellation and timeouts

### Container Registry Integration
- **go-containerregistry**: Primary library for OCI-compliant registry operations
- **AWS SDK Go v2**: Native AWS ECR integration with modern SDK patterns
- **Google Cloud Client Libraries**: Native GCP integration for GCR and Artifact Registry

### Security and Encryption
- **AWS KMS & GCP KMS**: Customer-managed encryption keys for data at rest
- **Cosign Integration**: Image signing and verification capabilities
- **Cloud Secrets Manager**: AWS Secrets Manager and Google Secret Manager integration
- **TLS/HTTPS**: All network communications encrypted in transit

### Observability and Monitoring
- **Prometheus Metrics**: Built-in metrics collection and exposure
- **Structured Logging**: JSON-based logging with contextual information
- **OpenTelemetry**: Distributed tracing and observability (planned)

## Architectural Patterns

### 1. Interface-Driven Design
```go
// Central interfaces in pkg/interfaces/
type RegistryClient interface {
    ListRepositories(ctx context.Context, prefix string) ([]string, error)
    GetRepository(ctx context.Context, name string) (Repository, error)
    GetRegistryName() string
}
```

**Principles:**
- Interfaces defined where used, not where implemented
- Small, focused interfaces following Single Responsibility Principle
- Enables dependency injection and testing with mocks

### 2. Composition over Inheritance
```go
// Base functionality through embedding
type ECRClient struct {
    *common.BaseClient  // Embedded base functionality
    ecrService *ecr.Service  // ECR-specific functionality
}
```

**Benefits:**
- Shared behavior without deep inheritance hierarchies
- Explicit dependencies and clear separation of concerns
- Flexible composition of behaviors

### 3. Options Pattern for Configuration
```go
type ClientOptions struct {
    Region    string
    Logger    *log.Logger
    Transport http.RoundTripper
    Timeout   time.Duration
}

func NewClient(opts ClientOptions) (*Client, error) {
    // Set defaults and create client
}
```

**Advantages:**
- Self-documenting parameter names
- Backward compatibility when adding options
- Clear defaults and validation

### 4. Worker Pool Pattern for Concurrency
```go
// Configurable parallel processing
pool := replication.NewWorkerPool(workerCount, logger)
pool.Submit(jobID, func(ctx context.Context) error {
    // Work function
})
```

**Design Goals:**
- Bounded concurrency to prevent resource exhaustion
- Context-aware cancellation and timeout handling
- Result collection and error aggregation

## Code Organization Standards

### Package Structure
```
pkg/
├── client/           # Registry client implementations
│   ├── common/       # Shared client interfaces and base implementations
│   ├── ecr/          # AWS ECR specific client
│   └── gcr/          # Google GCR specific client
├── config/           # Configuration management
├── copy/             # Image copying logic
├── helper/           # Utility packages
│   ├── errors/       # Error handling utilities
│   ├── log/          # Logging utilities
│   └── util/         # General utilities
├── interfaces/       # Central interface definitions
├── replication/      # Replication orchestration
├── security/         # Security-related functionality
└── service/          # Business logic services
```

### File Naming Conventions
- **Source files**: `snake_case.go` with descriptive nouns
- **Test files**: `source_file_test.go`
- **Interface files**: `interfaces.go` for central interfaces
- **Implementation files**: `provider_type.go` (e.g., `aws_provider.go`)

### Import Organization
```go
import (
    // Standard library
    "context"
    "fmt"
    
    // Third-party packages
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/spf13/cobra"
    
    // Internal packages
    "freightliner/pkg/client/common"
    "freightliner/pkg/config"
)
```

## Error Handling Standards

### Error Types and Wrapping
```go
// Use custom error types for different categories
return nil, errors.NotFoundf("repository %s not found", name)
return nil, errors.Wrap(err, "failed to authenticate with registry")
```

### Error Context
- Always add context to errors when wrapping
- Use domain-specific error types for different failure modes
- Include relevant metadata in error messages

## Security Standards

### 1. Credential Management
- Never log or expose credentials in plain text
- Use environment variables or secure credential providers
- Support for cloud-native credential sources (IAM roles, service accounts)

### 2. Encryption Standards
- AES-256 encryption for data at rest
- Customer-managed keys (CMK) for enterprise security
- Envelope encryption for performance with large data

### 3. Network Security
- All HTTP communications over TLS
- Certificate validation and proper TLS configuration
- Support for corporate proxy environments

## Performance Standards

### 1. Concurrency Management
- Configurable worker pools with sensible defaults
- Context-aware cancellation for all long-running operations
- Proper resource cleanup with defer statements

### 2. Memory Efficiency
- Stream processing for large images to avoid memory exhaustion
- Bounded channel sizes to prevent memory leaks
- Proper context cleanup in goroutines

### 3. Network Optimization
- HTTP connection reuse and pooling
- Compression for network transfers
- Delta updates to minimize bandwidth usage

## Testing Standards

### 1. Test Structure
```go
func TestFunction(t *testing.T) {
    testCases := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        // Test cases
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2. Mock Usage
- Use gomock for interface mocking
- Create focused unit tests with minimal dependencies
- Integration tests for end-to-end scenarios

### 3. Test Coverage
- Minimum 80% code coverage for new features
- Focus on edge cases and error conditions
- Performance tests for critical paths

## Configuration Management

### 1. Configuration Hierarchy
1. Command-line flags (highest priority)
2. Environment variables
3. Configuration files (YAML)
4. Default values (lowest priority)

### 2. Configuration Validation
- Validate configuration at startup
- Provide clear error messages for invalid configuration
- Support for configuration file schemas

## Logging and Observability

### 1. Structured Logging
```go
logger.Info("Operation completed", map[string]interface{}{
    "operation": "replicate",
    "source":    sourceRepo,
    "dest":      destRepo,
    "duration":  duration,
})
```

### 2. Metrics Collection
- Prometheus metrics for operational visibility
- Business metrics (replications, errors, performance)
- System metrics (memory, CPU, network)

### 3. Tracing (Future)
- OpenTelemetry integration for distributed tracing
- Request correlation across service boundaries

## Deployment and Operations

### 1. Container Support
- Multi-architecture container images (AMD64, ARM64)
- Minimal base images for security
- Non-root user execution

### 2. Configuration Management
- Environment variable support for containerized deployments
- Configuration file mounting for Kubernetes
- Secret management integration

### 3. Health Checks and Monitoring
- Health check endpoints for load balancers
- Readiness and liveness probes
- Metrics endpoints for monitoring systems

## Code Quality Standards

### 1. Static Analysis (Recently Overhauled)
- **golangci-lint v2** with focused, meaningful checks only
  - `errcheck` - Unchecked error detection (critical for reliability)
  - `govet` - Standard Go vet checks (real bug detection)
  - `ineffassign` - Ineffectual assignment detection (potential bugs)
  - `misspell` - Spelling mistake detection
- **Eliminated noisy linters** (gosec, staticcheck, unused) to reduce development toil
- **Result**: 0 linting issues across all CI pipelines with fast, reliable checks

### 2. Code Formatting
- gofmt for consistent formatting
- goimports for import organization
- Pre-commit hooks for automated checks

### 3. Documentation
- GoDoc comments for all exported functions and types
- README files for package-level documentation
- Architecture decision records (ADRs) for significant decisions

## Dependencies Management

### 1. Dependency Selection
- Prefer standard library when possible
- Choose well-maintained, widely-used third-party libraries
- Minimize dependency tree depth and conflicts

### 2. Dependency Updates
- Regular security updates for all dependencies
- Compatibility testing before major version updates
- Vendor critical dependencies when appropriate

### 3. License Compliance
- Only use dependencies with compatible licenses
- Maintain license inventory for compliance
- Avoid GPL and other copyleft licenses in core functionality