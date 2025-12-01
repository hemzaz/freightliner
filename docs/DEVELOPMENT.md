# Development Guide

## Setup

```bash
# Clone
git clone <repo>
cd freightliner

# Install dependencies
go mod download

# Run locally
go run cmd/server/main.go
```

## Development Environment

```bash
# Start services (Docker Compose)
docker-compose -f docker-compose.dev.yml up -d

# Services running:
# - Source registry: localhost:5000
# - Dest registry: localhost:5001
# - Redis: localhost:6380
# - MinIO: localhost:9002
# - Prometheus: localhost:9091
```

## Project Structure

```
freightliner/
├── cmd/                # CLI entry points
├── pkg/                # Core packages (35 packages)
│   ├── client/        # Registry clients (ECR, GCR)
│   ├── replication/   # Replication logic
│   ├── server/        # HTTP server
│   └── metrics/       # Prometheus metrics
├── tests/             # Integration tests
├── deployments/       # Kubernetes manifests
└── scripts/           # Build/deployment scripts
```

## Running Tests

```bash
# All tests
go test ./...

# Unit tests only
go test -short ./...

# Integration tests
go test -tags=integration ./tests/integration/...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Style

```bash
# Format
gofmt -s -w .
goimports -w .

# Lint
golangci-lint run ./...

# Pre-commit
make lint test
```

## Building

```bash
# Binary
make build

# Docker image
make docker

# Multi-platform
docker buildx build --platform linux/amd64,linux/arm64 -t freightliner:latest .
```

## Debugging

```bash
# Enable debug logs
export LOG_LEVEL=debug
go run cmd/server/main.go

# Delve debugger
dlv debug cmd/server/main.go

# Remote debugging in Kubernetes
kubectl port-forward pod/<pod-name> 2345:2345
```

## Adding Features

1. **Create spec**: `.claude/specs/<feature>/requirements.md`
2. **Write tests**: `pkg/<package>/<feature>_test.go`
3. **Implement**: `pkg/<package>/<feature>.go`
4. **Document**: Update relevant docs
5. **PR**: Open pull request with tests

## Architecture Patterns

- **Interfaces**: Consumer-defined, small focused interfaces
- **Composition**: Embed base types, extend with provider-specific
- **Context**: All operations accept `context.Context`
- **Error handling**: Wrap errors with context
- **Testing**: Co-locate tests with source

## Performance

```bash
# Benchmarks
go test -bench=. ./...

# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## Contributing

See CONTRIBUTING.md for:
- Code review process
- Branch naming
- Commit message format
- Release process
