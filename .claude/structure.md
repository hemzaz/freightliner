# Project Structure and Organization - Freightliner

## Directory Structure Overview

```
freightliner/
├── cmd/                          # Command-line applications and CLI logic
│   ├── checkpoint.go             # Checkpoint management commands
│   ├── checkpoint/               # Checkpoint-specific command handlers
│   ├── replicate.go              # Single repository replication commands
│   ├── replicate/                # Replication command handlers
│   ├── replicate_tree.go         # Tree replication commands
│   ├── root.go                   # Root command and CLI setup
│   ├── serve.go                  # Server mode commands
│   ├── serve/                    # Server command handlers
│   └── version.go                # Version command
├── pkg/                          # Library code and core functionality
│   ├── client/                   # Registry client implementations
│   ├── config/                   # Configuration management
│   ├── copy/                     # Image copying logic
│   ├── helper/                   # Utility packages
│   ├── interfaces/               # Central interface definitions
│   ├── metrics/                  # Metrics collection and reporting
│   ├── network/                  # Network optimization (compression, delta)
│   ├── replication/              # Replication orchestration and rules
│   ├── secrets/                  # Cloud secrets manager integration
│   ├── security/                 # Security features (encryption, signing)
│   ├── server/                   # HTTP server implementation
│   ├── service/                  # Business logic services
│   └── tree/                     # Tree replication logic
├── test/                         # Test utilities and fixtures
│   ├── fixtures/                 # Test data and fixtures
│   ├── integration/              # Integration test suites
│   └── mocks/                    # Generated mock implementations
├── docs/                         # Technical documentation
├── examples/                     # Configuration examples
├── scripts/                      # Build and development scripts
├── main.go                       # Application entry point
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── Makefile                      # Build automation
├── README.md                     # Project documentation
├── GUIDELINES.md                 # Development guidelines
└── CLAUDE.md                     # Claude-specific instructions
```

## Package Organization Principles

### 1. Domain-Driven Structure
Packages are organized by business domain and functionality rather than technical layers:

- **client/**: Registry interaction and client implementations
- **replication/**: Replication orchestration and business logic  
- **security/**: Security-related functionality (encryption, signing)
- **service/**: High-level business services

### 2. Shared Infrastructure
Common utilities and shared code organized by purpose:

- **helper/**: General utilities (errors, logging, throttling)
- **interfaces/**: Central interface definitions to prevent circular dependencies
- **config/**: Configuration loading and management

### 3. Separation of Concerns
Clear boundaries between different types of code:

- **cmd/**: CLI interface and command handling (presentation layer)
- **pkg/**: Core business logic and domain functionality
- **test/**: Testing utilities separate from production code

## File Naming Conventions

### Source Files
- **Base functionality**: `client.go`, `repository.go`, `copier.go`
- **Provider-specific**: `ecr_client.go`, `gcr_client.go`, `aws_provider.go`
- **Feature-specific**: `encryption.go`, `compression.go`, `delta.go`
- **Base implementations**: `base_client.go`, `base_repository.go`

### Test Files
- **Unit tests**: `client_test.go`, `repository_test.go`
- **Integration tests**: `client_integration_test.go`
- **Benchmark tests**: `client_bench_test.go`

### Configuration Files
- **Interface definitions**: `interfaces.go`
- **Type definitions**: `types.go`
- **Error definitions**: `errors.go`
- **Utility functions**: `util.go` or `registry_util.go`

## Package Dependencies

### Dependency Direction Rules
```
cmd/ → pkg/service/ → pkg/replication/ → pkg/client/ → pkg/helper/
  ↓         ↓              ↓                ↓
pkg/config/  ↓         pkg/copy/      pkg/interfaces/
             ↓              ↓
        pkg/metrics/   pkg/network/
```

**Rules:**
- Dependencies flow from outer layers (cmd) to inner layers (helper)
- No circular dependencies between packages
- `interfaces/` package can be imported by any package
- `helper/` packages should not import business logic packages

### Import Organization Within Files

```go
import (
    // Standard library imports
    "context"
    "fmt"
    "strings"
    
    // Third-party imports (alphabetical)
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/google/go-containerregistry/pkg/name"
    "github.com/spf13/cobra"
    
    // Internal imports (by dependency depth)
    "freightliner/pkg/interfaces"
    "freightliner/pkg/helper/log"
    "freightliner/pkg/client/common"
    "freightliner/pkg/service"
)
```

## Code Organization Within Packages

### Client Package Structure
```
pkg/client/
├── common/                       # Shared client functionality
│   ├── base_client.go            # Base client implementation
│   ├── base_repository.go        # Base repository implementation
│   ├── base_authenticator.go     # Base authentication
│   ├── base_transport.go         # Base HTTP transport
│   ├── enhanced_client.go        # Enhanced client with additional features
│   ├── enhanced_repository.go    # Enhanced repository with extra capabilities
│   ├── interfaces.go             # Client-specific interfaces
│   ├── errors.go                 # Client-specific errors
│   └── registry_util.go          # Registry utility functions
├── ecr/                          # AWS ECR specific implementation
│   ├── client.go                 # ECR client implementation
│   ├── repository.go             # ECR repository implementation
│   ├── auth.go                   # ECR authentication
│   └── *_test.go                 # ECR tests
└── gcr/                          # Google GCR specific implementation
    ├── client.go                 # GCR client implementation
    ├── repository.go             # GCR repository implementation
    ├── auth.go                   # GCR authentication
    ├── parser.go                 # GCR-specific parsing
    └── *_test.go                 # GCR tests
```

### Service Package Structure
```
pkg/service/
├── interfaces.go                 # Service interfaces
├── replicate.go                  # Repository replication service
├── checkpoint.go                 # Checkpoint management service
└── tree_replicate.go             # Tree replication service
```

### Helper Package Structure
```
pkg/helper/
├── errors/                       # Error handling utilities
│   └── errors.go                 # Custom error types and functions
├── log/                          # Logging utilities
│   ├── logger.go                 # Logger implementation
│   └── log_test.go               # Logger tests
├── throttle/                     # Rate limiting and throttling
│   ├── limiter.go                # Rate limiter implementation
│   └── throttle_test.go          # Throttling tests
└── util/                         # General utilities
    ├── digest.go                 # Digest utilities
    ├── retry.go                  # Retry logic
    ├── errgroup_helpers.go       # Error group utilities
    └── mutex_helpers.go          # Mutex utilities
```

## Configuration File Organization

### Main Configuration
- **config_example.yaml**: Complete configuration example with all options
- **Environment-specific configs**: Separate files for different environments

### Documentation Structure
```
docs/
├── CODE_REUSE_PATTERNS.md        # Code reuse guidelines
├── CONCURRENCY_PATTERNS.md       # Concurrency best practices  
├── CONFIGURATION.md               # Configuration documentation
├── ENHANCED_IMPLEMENTATIONS.md    # Advanced implementations guide
├── FORMATTING_LINTING.md          # Code quality standards
├── GO_VET_GUIDE.md               # Static analysis guide
├── IMPORT_ORGANIZATION.md         # Import organization rules
├── LINTING_GUIDE.md              # Linting tool usage
├── METHOD_ORGANIZATION.md         # Method organization patterns
├── RETURN_STYLE_GUIDE.md         # Return value conventions
├── SHARED_IMPLEMENTATIONS.md      # Shared code patterns
├── STATICCHECK_GUIDE.md          # Advanced static analysis
└── STYLE_EXAMPLES.md             # Concrete style examples
```

## Testing Structure

### Test Organization
```
test/
├── fixtures/                     # Test data and fixtures
│   ├── configs/                  # Test configuration files
│   ├── images/                   # Test container images
│   └── manifests/                # Test manifests
├── integration/                  # Integration tests
│   ├── ecr_integration_test.go   # ECR integration tests
│   ├── gcr_integration_test.go   # GCR integration tests
│   └── replication_test.go       # End-to-end replication tests
└── mocks/                        # Generated mocks
    ├── mock_client.go            # Mock client implementations
    ├── mock_repository.go        # Mock repository implementations
    └── mock_service.go           # Mock service implementations
```

### Test File Placement
- **Unit tests**: Co-located with source files (`client_test.go` next to `client.go`)
- **Integration tests**: In `test/integration/` directory
- **Mocks**: Generated in `test/mocks/` directory
- **Fixtures**: Shared test data in `test/fixtures/`

## Build and Development Structure

### Scripts Directory
```
scripts/
├── lint.sh                       # Linting script
├── organize_imports.sh           # Import organization script
├── pre-commit                    # Pre-commit hook
├── setup-test-registries.sh      # Test registry setup script
├── test-registry-setup.sh        # Alternative test registry setup
├── test-with-manifest.sh         # Test execution with manifest filtering
└── vet.sh                        # Go vet script
```

### Root Level Files
- **Makefile**: Build automation and common tasks
- **go.mod/go.sum**: Go module management
- **main.go**: Application entry point (minimal, delegates to cmd/)
- **.gitignore**: Git ignore patterns
- **staticcheck.conf**: Static analysis configuration

## Interface Definition Strategy

### Central Interfaces (`pkg/interfaces/`)
For interfaces used across multiple packages to prevent circular dependencies:

```go
// Core registry interfaces
type RegistryClient interface { ... }
type Repository interface { ... }
type Authenticator interface { ... }
```

### Package-Specific Interfaces
For interfaces used within a single domain:

```go
// In pkg/copy/interfaces.go
type ImageCopier interface { ... }
type CompressionProvider interface { ... }
```

### Consumer-Defined Interfaces
For specific use cases that don't need the full interface:

```go
// In service package, define only what's needed
type TagLister interface {
    ListTags(ctx context.Context) ([]string, error)
}
```

## Naming Conventions Summary

### Package Names
- Lowercase, single word when possible
- Descriptive of package purpose: `client`, `replication`, `security`
- Avoid generic names like `common`, `util` unless truly cross-cutting

### File Names
- snake_case with descriptive nouns
- Provider-specific: `aws_provider.go`, `gcp_provider.go`  
- Feature-specific: `compression.go`, `encryption.go`
- Base implementations: `base_client.go`

### Type Names
- PascalCase for exported types: `RegistryClient`, `ReplicationService`
- camelCase for unexported types: `registryAuth`, `imageCache`
- Interface names with -er suffix when appropriate: `Authenticator`, `Copier`

### Function and Method Names
- PascalCase for exported: `NewClient`, `ListRepositories`
- camelCase for unexported: `parseRegistryPath`, `shouldSkipImage`
- Constructor pattern: `New*` for constructors

This structure ensures maintainable, scalable code organization that supports the project's growth while maintaining clear boundaries and dependencies.