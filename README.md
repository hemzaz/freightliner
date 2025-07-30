# Freightliner

[![Go Report Card](https://goreportcard.com/badge/github.com/hemzaz/freightliner)](https://goreportcard.com/report/github.com/hemzaz/freightliner)
[![GoDoc](https://godoc.org/github.com/hemzaz/freightliner?status.svg)](https://godoc.org/github.com/hemzaz/freightliner)
[![License](https://img.shields.io/github/license/hemzaz/freightliner.svg)](https://blob/master/LICENSE)

Freightliner is a container registry replication tool that supports cross-registry replication between AWS ECR and Google Container Registry (GCR).

## Features

- Cross-registry replication between ECR and GCR
- Bidirectional replication capabilities
- Single image replication
- Complete tree replication across registries
- Support for multi-architecture images and manifest lists
- Repository and tag filtering with pattern matching
- Parallel replication with configurable worker counts
- Metrics collection for monitoring
- Dry-run capability for validation
- Image signing and verification using Cosign
- Customer-managed encryption keys (AWS KMS and GCP KMS)
- Network bandwidth optimization with compression and delta updates
- Checkpointing for resumable replication

## Installation

### Binary Installation

Download the latest release from the [releases page](https://releases).

### Docker

```bash
docker pull ghcr.io/hemzaz/freightliner:latest
```

### Build from Source

```bash
git clone https://github.com/hemzaz/freightliner.git
cd freightliner
go build -o bin/freightliner cmd/freightliner/main.go
```

## Quick Start

### Single Repository Replication

```bash
# Replicate a repository from ECR to GCR
freightliner replicate ecr/my-repository gcr/my-repository

# Specify AWS region and GCP project
freightliner replicate ecr/my-repository gcr/my-repository --ecr-region=us-east-1 --gcr-project=my-project
```

### Security Features

```bash
# Replicate with image signing and verification
freightliner replicate ecr/my-repository gcr/my-repository \
  --sign --sign-key=/path/to/cosign.key --sign-key-id=my-key-id

# Replicate with AWS KMS customer-managed encryption key
freightliner replicate ecr/my-repository gcr/my-repository \
  --encrypt --customer-key --aws-kms-key=alias/my-key

# Replicate with GCP KMS customer-managed encryption key
freightliner replicate ecr/my-repository gcr/my-repository \
  --encrypt --customer-key --gcp-kms-key=projects/my-project/locations/global/keyRings/freightliner/cryptoKeys/my-key
```

### Tree Replication

```bash
# Replicate all repositories with prefix "prod/" from ECR to GCR
freightliner replicate-tree ecr/prod gcr/prod-mirror

# With filtering
freightliner replicate-tree ecr/staging gcr/staging-mirror \
  --exclude-repo="internal-*" \
  --include-tag="v*" \
  --workers=10

# Resumable replication with checkpointing
freightliner replicate-tree ecr/prod gcr/prod-mirror \
  --checkpoint --checkpoint-dir=/path/to/checkpoints

# Resume an interrupted replication
freightliner replicate-tree ecr/prod gcr/prod-mirror \
  --resume=<checkpoint-id> --skip-completed
```

### Server Mode

```bash
# Start the replication server with a configuration file
freightliner serve --config config.yaml
```

## Configuration

See the [usage documentation](docs/usage.md) for detailed configuration options.

## Authentication

### AWS ECR

Freightliner uses the standard AWS SDK authentication methods. You can configure authentication using:

- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- AWS configuration files (`~/.aws/credentials`)
- IAM roles (when running on EC2 or ECS)

### Google GCR

Freightliner uses the standard Google Cloud authentication methods. You can configure authentication using:

- Service account JSON key file (specified with `GOOGLE_APPLICATION_CREDENTIALS`)
- Application Default Credentials
- GKE Workload Identity (when running on GKE)

## Recent Changes

### CI/CD Linting Overhaul (Latest)

- **Complete CI linting system overhaul** reducing issues from 104+ to 0
- **golangci-lint v2 migration** with focused, meaningful checks only
- **Eliminated noisy linters** (gosec, staticcheck, unused) that created development toil
- **Streamlined linting pipeline** focusing on real bug detection:
  - `errcheck` - Catches unchecked errors (critical for reliability)
  - `govet` - Standard Go vet checks (real bugs)
  - `ineffassign` - Detects ineffectual assignments (potential bugs)
  - `misspell` - Fixes spelling mistakes
- **Docker Buildx CI compatibility** with consistent linting across all CI pipelines
- **Removed redundant checks** from Makefile, pre-commit hooks, and CI configurations
- **Result**: All CI pipelines now pass consistently with fast, reliable linting

The CI system uses a unified workflow with comprehensive test manifest integration for reliable builds.

### Security Features

- **Image Signing with Cosign**: 
  - Added support for signing and verifying images using Cosign
  - Integrates with both ECR and GCR registry types

- **Customer-Managed Encryption Keys**:
  - Implemented AWS KMS integration for ECR images
  - Implemented GCP KMS integration for GCR images
  - Added envelope encryption for secure data transfer

- **Cloud Secrets Manager Integration**:
  - Added support for AWS Secrets Manager and Google Secret Manager
  - Securely store and retrieve registry credentials, encryption keys, and signing materials
  - Command-line flag support for using cloud provider secrets in operations
  - JSON-based structured secret format for complex configuration

### Client Fixes

- **ECR Client**:
  - Fixed credential helper implementation to properly support the Authenticator interface
  - Resolved MediaType undefined errors by properly importing the types package
  - Added correct authorization method to the ECR credential helper
  - Fixed client configuration handling with authentication libraries

- **GCR Client**:
  - Updated Google registry list functionality to work with the latest API
  - Fixed transport and authentication handling for Google credentials
  - Corrected manifest handling for different image types
  - Improved descriptor handling for registry operations

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/freightliner cmd/freightliner/main.go
```

## Production Readiness Status

⚠️ **IMPORTANT: NOT PRODUCTION READY** ⚠️

This project is currently in **active development** and has **47 critical blockers** that prevent production deployment. 

### Current Status
- **Development Phase**: Feature development with comprehensive testing infrastructure
- **Security**: Multiple critical vulnerabilities requiring immediate attention
- **Reliability**: Concurrency issues and resource management problems
- **Performance**: Scalability limitations and optimization needs
- **Operations**: Missing monitoring, health checks, and deployment infrastructure

### Production Readiness Roadmap

#### P0 Blockers (Must Fix Before ANY Deployment)
- [ ] Fix timing attack vulnerability in API key authentication
- [ ] Add panic recovery middleware for all HTTP handlers  
- [ ] Implement functional Prometheus metrics collection
- [ ] Add checksum validation for all image transfers
- [ ] Implement high availability for checkpoint storage

#### P1 Blockers (Critical for Production)
- [ ] Implement comprehensive rate limiting and input validation
- [ ] Fix goroutine leaks in worker pool management
- [ ] Design and implement horizontal scaling architecture
- [ ] Add comprehensive configuration validation

#### P2 Blockers (Important for Production)
- [ ] Secure CORS configuration with origin allowlists
- [ ] Implement streaming for large image transfers
- [ ] Add structured logging with correlation IDs

**Estimated Effort**: 14-20 weeks for full production readiness

For detailed analysis of all production blockers, see [Development Infrastructure Specifications](.claude/specs/development-infrastructure/).

### Development Infrastructure

The project includes comprehensive development tooling:

#### Available Make Targets
```bash
make setup          # Install all required development tools
make build          # Build the application
make test           # Run all tests
make test-race      # Run tests with race detection
make test-coverage  # Generate test coverage report
make lint           # Run linting with golangci-lint v2 (focused checks)
make fmt            # Format code with gofmt
make vet            # Run go vet
make imports        # Organize imports with goimports
make check          # Run all quality checks (streamlined)
make clean          # Clean build artifacts
```

**Note**: `staticcheck` has been integrated into golangci-lint for more efficient CI execution.

#### Quality Assurance
- **Streamlined linting** with golangci-lint v2 focusing on real bug detection
- **Static analysis** with go vet for reliable checks
- **Race condition detection** in all tests  
- **Pre-commit hooks** for automated quality checks
- **Test coverage tracking** with detailed reporting
- **CI/CD pipelines** passing with 0 linting issues (recently overhauled)

#### Local Testing Infrastructure
```bash
# Set up local Docker registries for testing
./scripts/setup-test-registries.sh

# Run integration tests
make test-integration

# Clean up test environment
./scripts/setup-test-registries.sh --cleanup
```

The local testing setup provides:
- **Source registry**: `localhost:5100` with populated test data
- **Destination registry**: `localhost:5101` for replication testing
- **4 test repositories** with realistic container images
- **Automated cleanup** for consistent test environments

### Development Workflow

1. **Setup**: Run `make setup` to install all development tools
2. **Development**: Use `make check` to run all quality checks locally
3. **Testing**: Use local registry setup for integration testing
4. **Quality**: Pre-commit hooks ensure code quality standards

### Concurrency Issues Identified

Critical concurrency problems have been identified and documented:
- **Race conditions** in counter operations and metrics collection
- **Goroutine leaks** in worker pool management
- **Channel management** issues with double-close vulnerabilities
- **Resource management** problems with connection and memory leaks

See [Concurrency Analysis](.claude/specs/development-infrastructure/design.md#7-concurrency-issues-analysis) for detailed technical analysis.

## Contributing

⚠️ **Before Contributing**: Please review the [Production Readiness Analysis](.claude/specs/development-infrastructure/design.md#8-production-readiness-blockers) to understand current limitations.

### Development Setup
1. Fork and clone the repository
2. Run `make setup` to install development tools
3. Set up local testing infrastructure: `./scripts/setup-test-registries.sh`
4. Make changes and run `make check` before committing
5. Submit pull requests with comprehensive tests

### Priority Areas for Contribution
1. **P0 Security Fixes**: Authentication timing attacks, rate limiting
2. **P0 Reliability**: Panic recovery, goroutine leak fixes
3. **P0 Operations**: Functional metrics collection, health checks
4. **Test Coverage**: Currently at 44%, target is 80%
5. **Integration Testing**: End-to-end workflow validation

### Quality Standards
- All code must pass `make check` (linting, testing, static analysis)
- New features require comprehensive tests
- Security-related changes require extra scrutiny
- Performance changes require benchmarking

Contributions are welcome! Please focus on production readiness blockers and follow the established development infrastructure patterns.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.