# Contributing to Freightliner

Thank you for your interest in contributing to Freightliner! This guide will help you get started with development using our AI-enhanced workflow.

## 🚀 Quick Start

### Prerequisites

- Go 1.24.5 or later
- Docker
- kubectl (for Kubernetes deployments)
- Helm 3.x
- AWS CLI (for ECR testing)
- gcloud CLI (for GCR testing)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/hemzaz/freightliner.git
cd freightliner

# Install development tools
make setup

# Verify setup
make quality
```

## 🤖 AI-Assisted Development

Freightliner uses Claude Code for AI-assisted development. The `.claude/` directory contains:

- **Slash commands** for common workflows
- **Specialized agents** for different tasks
- **Skills** for domain expertise
- **Workflows** for complex processes

### Using Slash Commands

```bash
# Add a new feature
/add-feature "Your feature description"

# Fix CI failures
/fix-ci

# Run security audit
/security-audit

# Deploy to staging
/deploy staging v1.2.0
```

See [.claude/README.md](.claude/README.md) for complete documentation.

## 📝 Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Implement Your Changes

Follow the project's architecture patterns:

#### Interface-Driven Design
```go
// Define interfaces where used
type MyService interface {
    DoSomething(ctx context.Context, input string) (*Result, error)
}
```

#### Error Handling
```go
import "freightliner/pkg/helper/errors"

if err != nil {
    return errors.Wrap(err, "failed to do something")
}
```

#### Structured Logging
```go
logger.WithFields(map[string]interface{}{
    "operation": "replicate",
    "duration":  duration,
}).Info("Operation completed")
```

#### Context Usage
```go
func DoOperation(ctx context.Context, input string) error {
    // Always accept and pass context
    if err := ctx.Err(); err != nil {
        return err
    }

    return downstreamOperation(ctx, input)
}
```

See [.claude/skills/go-microservice.md](.claude/skills/go-microservice.md) for detailed patterns.

### 3. Write Tests

```go
func TestMyFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Result
        wantErr bool
    }{
        {
            name:  "successful case",
            input: "test",
            want:  &Result{Value: "test"},
        },
        {
            name:    "error case",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFeature(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            require.Equal(t, tt.want, got)
        })
    }
}
```

**Requirements:**
- Minimum 80% code coverage
- Table-driven tests with subtests
- Test error conditions and edge cases

### 4. Run Quality Checks

```bash
# Format code
make fmt

# Run linting
make lint

# Run tests
make test-ci

# Run security scan
make security

# All quality checks
make quality
```

### 5. Update Documentation

- Add GoDoc comments to all exported symbols
- Update README.md if user-facing changes
- Update architecture docs if structural changes
- Add examples if new features

### 6. Commit Changes

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git commit -m "feat: add support for Azure Container Registry"
git commit -m "fix: resolve authentication timeout in ECR client"
git commit -m "docs: update deployment guide"
git commit -m "test: add integration tests for tree replication"
```

### 7. Push and Create PR

```bash
git push origin feature/your-feature-name
gh pr create --title "Your PR title" --body "Description"
```

## 🎯 Contribution Guidelines

### Code Quality Standards

#### Must Have
- ✅ All tests passing
- ✅ 80%+ code coverage for new code
- ✅ Linting passes (golangci-lint)
- ✅ Security scan clean (gosec)
- ✅ GoDoc comments on all exports
- ✅ Follows project architecture patterns

#### Best Practices
- Use interface-driven design
- Implement proper error handling
- Add structured logging
- Use context.Context for cancellation
- Implement graceful cleanup
- Follow Go best practices

### Architecture Constraints

1. **Interfaces defined where used**, not where implemented
2. **Composition over inheritance** (embedding)
3. **Options pattern** for configuration
4. **Worker pools** for concurrency
5. **Layered architecture** (server → service → client)

### Security Requirements

- ❌ No hardcoded credentials
- ✅ Use cloud-native credential providers
- ✅ Sanitize all inputs
- ✅ Encrypt sensitive data at rest
- ✅ Use TLS for all network communication
- ✅ Follow least privilege principle

### Testing Requirements

#### Unit Tests
- Table-driven tests with subtests
- Test error conditions
- Mock external dependencies
- Use gomock for interface mocking

#### Integration Tests
- Test with real registries when possible
- Use local Docker registries for testing
- Clean up resources after tests
- Handle timeouts and retries

#### Performance Tests
- Benchmark critical paths
- Profile for bottlenecks
- Test with realistic data sizes
- Measure memory allocations

## 🔧 Development Tools

### Makefile Targets

```bash
make build          # Build the application
make test           # Run all tests
make test-unit      # Run unit tests only
make test-ci        # Run CI test suite
make lint           # Run golangci-lint
make fmt            # Format code
make quality        # Run all quality checks
make security       # Run security scan
make dev            # Full dev cycle
make clean          # Clean build artifacts
```

### Testing with Local Registries

```bash
# Set up local Docker registries
./scripts/setup-test-registries.sh

# Run integration tests
make test-integration

# Clean up
./scripts/setup-test-registries.sh --cleanup
```

## 🐛 Debugging

### Common Issues

#### Import Errors
```bash
go mod tidy
go mod verify
```

#### Test Failures
```bash
# Run specific test
go test -v -run TestName ./pkg/...

# Run with race detection
make test-race

# Verbose output
go test -v ./...
```

#### Linting Errors
```bash
# Auto-fix formatting
make fmt

# See lint issues
make lint

# Check specific linters
golangci-lint run --enable-only=errcheck
```

## 📚 Resources

### Project Documentation
- [Architecture](docs/ARCHITECTURE.md)
- [Production Readiness](docs/PRODUCTION_READINESS_REPORT.md)
- [Security Guide](docs/SECURITY.md)
- [Operations Guide](docs/OPERATIONS.md)

### Development Guides
- [Go Microservice Patterns](.claude/skills/go-microservice.md)
- [Container Registry Operations](.claude/skills/container-registry.md)
- [Kubernetes Operations](.claude/skills/kubernetes-ops.md)

### AI Development
- [Claude Code Setup](.claude/README.md)
- [Agent Coordination](.claude/agents/COORDINATION.md)
- [Workflows](.claude/agents/WORKFLOWS.md)

## 🎓 Learning Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Container Registry Spec](https://github.com/opencontainers/distribution-spec)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)

## 🤝 Getting Help

- **Questions**: Open a GitHub Discussion
- **Bugs**: Create a GitHub Issue
- **Features**: Create a GitHub Issue with feature template
- **Security**: Email security@freightliner.dev

## 📋 PR Checklist

Before submitting a PR, ensure:

- [ ] Code follows project architecture patterns
- [ ] All tests pass (`make test-ci`)
- [ ] Code coverage >= 80% for new code
- [ ] Linting passes (`make lint`)
- [ ] Security scan clean (`make security`)
- [ ] Documentation updated (GoDoc, README, etc.)
- [ ] Commit messages follow Conventional Commits
- [ ] PR description explains changes and motivation
- [ ] Screenshots/examples for UI changes
- [ ] Breaking changes documented

## 🏆 Recognition

Contributors are recognized in:
- GitHub contributors list
- Release notes for significant contributions
- Project documentation

## 📜 License

By contributing to Freightliner, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to Freightliner! 🚂✨
