# CLAUDE.md - Freightliner Project Guide

This file provides guidance to Claude Code when working with the Freightliner container registry replication tool.

## Project Overview

Freightliner is a Go CLI application for container registry replication between AWS ECR and Google Container Registry (GCR). It supports bidirectional replication, multi-architecture images, encryption, signing, and resumable transfers.

## Technology Stack
- **Go 1.24.5** - Primary language
- **Cobra CLI** - Command-line interface framework
- **AWS SDK Go v2** - ECR integration
- **Google Cloud Client Libraries** - GCR/Artifact Registry integration
- **go-containerregistry** - OCI registry operations
- **Prometheus** - Metrics collection
- **Docker** - Container builds and testing

## Common Commands

### Build & Test
```bash
make build              # Build the application
make test               # Run all tests
make test-unit          # Run unit tests only
make test-ci            # Run CI-optimized tests with coverage
make lint               # Run golangci-lint
make quality            # Run all quality checks (fmt, vet, lint, security)
```

### Development
```bash
make dev                # Full dev cycle: clean, deps, build, test
make deps               # Download and verify dependencies
make fmt                # Format code
make security           # Run security scan (gosec)
```

### Docker
```bash
make docker-build       # Build Docker image
make docker-test        # Test Docker image
make docker-run         # Run container
```

## Code Quality Standards

### Go Best Practices
- Use `context.Context` for all operations with cancellation/timeouts
- Implement interfaces where used, not where implemented (interface-driven design)
- Prefer composition over inheritance (embedding)
- Use Options pattern for configuration structs
- Follow standard Go project layout (pkg/, cmd/, internal/)

### Error Handling
```go
// Wrap errors with context
return nil, errors.Wrap(err, "failed to authenticate with registry")

// Use domain-specific error types
return nil, errors.NotFoundf("repository %s not found", name)
```

### Testing Standards
- Table-driven tests with subtests
- Minimum 80% code coverage for new features
- Use gomock for interface mocking
- Test edge cases and error conditions
- Integration tests use `TestMain` for setup/teardown

### Code Structure
```
pkg/
├── client/         # Registry client implementations (ECR, GCR)
├── config/         # Configuration management
├── copy/           # Image copying logic
├── helper/         # Utilities (errors, log, util)
├── interfaces/     # Central interface definitions
├── replication/    # Replication orchestration
├── security/       # Encryption, signing
└── service/        # Business logic services
```

## Development Workflow

### Before Starting Work
1. Pull latest changes from master branch
2. Run `make deps` to ensure dependencies are up-to-date
3. Review existing code and architecture (see tech.md)
4. Plan implementation approach

### During Development
1. Run `make test` frequently to catch issues early
2. Follow interface-driven design patterns
3. Add structured logging with context
4. Update tests alongside code changes

### Before Committing
1. `make quality` - Run all quality checks
2. `make test-ci` - Ensure CI tests pass
3. Update documentation if needed
4. Follow conventional commits format

## Security Standards

### Credential Management
- Never log credentials in plain text
- Use cloud-native credential sources (IAM roles, service accounts)
- Support environment variables for configuration

### Encryption
- AES-256 for data at rest
- Customer-managed keys (AWS KMS, GCP KMS)
- All network communications over TLS

### Dependencies
- Run `make security` (gosec) regularly
- Keep dependencies updated (`go get -u && go mod tidy`)
- Audit third-party libraries for vulnerabilities

## Logging & Observability

### Structured Logging
```go
logger.Info("Operation completed", map[string]interface{}{
    "operation": "replicate",
    "source": sourceRepo,
    "dest": destRepo,
    "duration": duration,
})
```

### Metrics
- Prometheus metrics exposed on `/metrics` endpoint
- Track replications, errors, performance
- Monitor worker pool utilization

## CI/CD Integration

### GitHub Actions Workflows
- **comprehensive-validation.yml** - Main CI pipeline
- **security.yml** - Security scanning
- **release.yml** - Release automation

### Quality Gates
- All tests must pass (unit + integration)
- golangci-lint must pass (errcheck, govet, ineffassign, misspell)
- Code coverage reported
- Security scans clean (gosec)

## Review Checklist

Before marking any task as complete:
- [ ] Code follows Go best practices and project patterns
- [ ] Tests written and passing (`make test-ci`)
- [ ] Documentation updated (GoDoc, README if needed)
- [ ] Security considerations addressed
- [ ] No secrets or credentials in code
- [ ] Code quality checks pass (`make quality`)
- [ ] Performance impact considered (especially for replication logic)