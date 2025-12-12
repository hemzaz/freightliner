# CI/CD Pipeline Guide

This document provides a comprehensive overview of the Freightliner CI/CD pipeline, including workflows, best practices, and troubleshooting guidance.

## Table of Contents

- [Overview](#overview)
- [Workflows](#workflows)
- [Getting Started](#getting-started)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Advanced Configuration](#advanced-configuration)

## Overview

The Freightliner CI/CD pipeline is built on GitHub Actions and provides comprehensive automation for:

- ✅ Continuous Integration (build, test, lint)
- 🚀 Continuous Deployment (multi-environment)
- 🔒 Security scanning and monitoring
- ⚡ Performance testing
- 📦 Release management
- 🤖 PR automation

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     CI/CD Pipeline                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  PR Created                                                  │
│      │                                                       │
│      ├─> Consolidated CI                                    │
│      │   ├─> Build                                          │
│      │   ├─> Unit Tests (Ubuntu, macOS)                     │
│      │   ├─> Integration Tests                              │
│      │   ├─> Lint & Format                                  │
│      │   ├─> Security Scan                                  │
│      │   └─> Docker Build & Scan                            │
│      │                                                       │
│      └─> PR Automation                                      │
│          ├─> Validate PR Title                              │
│          ├─> Auto-label                                     │
│          ├─> Code Review                                    │
│          └─> Size Check                                     │
│                                                              │
│  Tag Pushed (v*.*.*)                                         │
│      │                                                       │
│      └─> Release Workflow                                   │
│          ├─> Validate Version                               │
│          ├─> Build Binaries (multi-platform)                │
│          ├─> Build Docker Images (multi-arch)               │
│          ├─> Generate SBOM                                  │
│          ├─> Security Scan                                  │
│          └─> Create GitHub Release                          │
│                                                              │
│  Manual Trigger                                              │
│      │                                                       │
│      ├─> Deploy Workflow                                    │
│      │   ├─> Prepare (dev/staging/production)               │
│      │   ├─> Validate                                       │
│      │   ├─> Deploy to Kubernetes                           │
│      │   └─> Verify                                         │
│      │                                                       │
│      └─> Performance Testing                                │
│          ├─> Go Benchmarks                                  │
│          ├─> Load Testing (k6)                              │
│          ├─> Memory Profiling                               │
│          └─> CPU Profiling                                  │
│                                                              │
│  Scheduled (Daily)                                           │
│      │                                                       │
│      └─> Security Monitoring                                │
│          ├─> Secret Scanning                                │
│          ├─> Dependency Vulnerabilities                     │
│          ├─> Container Security                             │
│          └─> Compliance Checks                              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Workflows

### 1. Consolidated CI (`consolidated-ci.yml`)

**Trigger**: Push to main/master/develop, Pull Requests

**Purpose**: Comprehensive continuous integration pipeline

**Jobs**:
- **Setup**: Dependency caching and preparation
- **Build**: Compile application and static binaries
- **Test Unit**: Run unit tests on Ubuntu and macOS
- **Test Integration**: Run integration tests with Docker registry
- **Lint**: Code formatting and linting checks
- **Security**: Security scanning (gosec, govulncheck)
- **Docker**: Build and scan Docker images
- **Benchmark**: Performance benchmarks
- **Status**: Overall CI status reporting

**Key Features**:
- ✅ Multi-OS testing (Ubuntu, macOS)
- ✅ Parallel job execution
- ✅ Docker layer caching
- ✅ Coverage reporting
- ✅ PR status comments

**Usage**:
```bash
# Automatically runs on push and PR
git push origin feature-branch

# Or trigger manually
gh workflow run consolidated-ci.yml
```

### 2. Release Workflow (`release.yml`)

**Trigger**: Tag push (v*.*.*), Manual dispatch

**Purpose**: Automated release management and distribution

**Jobs**:
- **Validate**: Pre-release validation and version checks
- **Build Binaries**: Multi-platform binary builds (Linux, macOS, Windows, amd64/arm64)
- **Build Docker**: Multi-architecture Docker images
- **Create Release**: GitHub release with assets
- **Notify**: Post-release notifications

**Key Features**:
- ✅ Semantic versioning
- ✅ Multi-platform builds
- ✅ SBOM generation
- ✅ Automated changelog
- ✅ Docker multi-arch support

**Usage**:
```bash
# Create and push a tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Or trigger manually
gh workflow run release.yml -f version=v1.0.0 -f prerelease=false
```

**Release Process**:
1. Validate version format (vX.Y.Z)
2. Run validation tests
3. Build binaries for all platforms
4. Build Docker images (amd64, arm64)
5. Generate checksums and SBOM
6. Create GitHub release with artifacts
7. Push Docker images to registry

### 3. Deployment Workflow (`deploy.yml`)

**Trigger**: Manual dispatch

**Purpose**: Deploy to Kubernetes environments

**Jobs**:
- **Build and Push**: Build and push Docker image
- **Deploy Dev**: Automatic deployment to development
- **Deploy Staging**: Manual approval for staging
- **Deploy Production**: Manual approval for production
- **Rollback**: Automatic rollback on failure

**Environments**:
- **Development**: `https://dev.freightliner.example.com`
- **Staging**: `https://staging.freightliner.example.com`
- **Production**: `https://freightliner.example.com`

**Key Features**:
- ✅ Environment-specific configurations
- ✅ Manual approval gates
- ✅ Automatic rollback
- ✅ Health checks
- ✅ Smoke tests

**Usage**:
```bash
# Deploy to development
gh workflow run deploy.yml -f environment=dev -f version=v1.0.0

# Deploy to staging (requires approval)
gh workflow run deploy.yml -f environment=staging -f version=v1.0.0

# Deploy to production (requires approval)
gh workflow run deploy.yml -f environment=production -f version=v1.0.0

# Dry run
gh workflow run deploy.yml -f environment=production -f version=v1.0.0 -f dry_run=true
```

### 4. Security Monitoring (`security-monitoring-enhanced.yml`)

**Trigger**: Daily schedule (2 AM UTC), Manual dispatch

**Purpose**: Continuous security monitoring and threat detection

**Jobs**:
- **Monitoring Init**: Configuration and initialization
- **Continuous Secret Monitoring**: TruffleHog, GitLeaks
- **Dependency Vulnerability Monitoring**: govulncheck, license compliance
- **Container Security Monitoring**: Trivy scanning, Docker best practices
- **Security Baseline Monitoring**: Security scoring and posture
- **Security Alerting**: Notifications and issue creation

**Scan Types**:
- **Full**: Complete security stack
- **Quick**: Secrets and SAST only
- **Dependencies**: Dependency vulnerabilities
- **Containers**: Container security
- **Secrets**: Secret detection

**Key Features**:
- ✅ Automated security scoring
- ✅ Trend analysis
- ✅ Multi-channel notifications
- ✅ Issue creation for critical findings
- ✅ License compliance checking

**Usage**:
```bash
# Run full security scan
gh workflow run security-monitoring-enhanced.yml -f scan_type=full

# Run quick scan
gh workflow run security-monitoring-enhanced.yml -f scan_type=quick

# Notify on success
gh workflow run security-monitoring-enhanced.yml -f notify_on_success=true
```

### 5. Performance Testing (`performance.yml`)

**Trigger**: Weekly schedule (Sunday 3 AM UTC), Manual dispatch

**Purpose**: Performance testing and profiling

**Jobs**:
- **Benchmark**: Go benchmark tests
- **Load Test**: k6 load testing
- **Stress Test**: System limits testing
- **Memory Profile**: Memory usage analysis
- **CPU Profile**: CPU usage analysis
- **Report**: Performance summary

**Test Types**:
- **Load**: Sustained load testing
- **Stress**: System breaking point
- **Spike**: Sudden traffic spikes
- **Endurance**: Long-duration testing

**Key Features**:
- ✅ k6 load testing integration
- ✅ Go profiling (CPU, memory)
- ✅ Performance baseline comparison
- ✅ Automated reporting
- ✅ Issue creation on failures

**Usage**:
```bash
# Run load test
gh workflow run performance.yml -f test-type=load -f duration=10 -f users=100

# Run all tests
gh workflow run performance.yml -f test-type=all

# Run stress test
gh workflow run performance.yml -f test-type=stress
```

### 6. PR Automation (`pr-automation.yml`)

**Trigger**: Pull request events

**Purpose**: Automated PR management and quality gates

**Jobs**:
- **Validate PR**: Title validation, size checks, breaking change detection
- **Auto Label**: Automatic labeling based on file changes
- **Code Review**: Automated code review suggestions
- **Check Dependencies**: Dependency update detection
- **Changelog**: Auto-generate changelog entries
- **PR Summary**: Comprehensive PR summary

**Key Features**:
- ✅ Semantic PR titles
- ✅ Automatic size labeling
- ✅ Breaking change detection
- ✅ Code quality suggestions
- ✅ Smart labeling system

**PR Title Format**:
```
type(scope): description

Examples:
- feat(registry): add support for OCI artifacts
- fix(cli): correct version display bug
- docs(readme): update installation instructions
- perf(copy): optimize layer transfer
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style changes
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `test`: Test additions/changes
- `build`: Build system changes
- `ci`: CI/CD changes
- `chore`: Maintenance tasks
- `revert`: Revert previous changes

## Getting Started

### Prerequisites

1. **GitHub Repository Secrets**:
   ```
   GITHUB_TOKEN              # Automatically provided
   CODECOV_TOKEN             # Optional: Codecov integration
   SLACK_SECURITY_WEBHOOK    # Optional: Slack notifications
   KUBE_CONFIG_DEV           # Kubernetes config for dev
   KUBE_CONFIG_STAGING       # Kubernetes config for staging
   KUBE_CONFIG_PROD          # Kubernetes config for production
   ```

2. **GitHub Environments**:
   - `dev`: Development environment (no approval required)
   - `staging`: Staging environment (optional approval)
   - `production`: Production environment (required approval)

3. **Local Tools**:
   ```bash
   # Install GitHub CLI
   brew install gh

   # Install required tools
   make tools
   ```

### First-Time Setup

1. **Configure Environments**:
   ```bash
   # Go to repository Settings > Environments
   # Create: dev, staging, production
   # Set protection rules for production
   ```

2. **Add Secrets**:
   ```bash
   # Go to repository Settings > Secrets and variables > Actions
   # Add required secrets
   ```

3. **Test Workflows**:
   ```bash
   # Trigger CI workflow
   git push origin feature-branch

   # Run manual workflow
   gh workflow run consolidated-ci.yml
   ```

## Best Practices

### Development Workflow

1. **Create Feature Branch**:
   ```bash
   git checkout -b feat/my-feature
   ```

2. **Make Changes and Test Locally**:
   ```bash
   make test
   make lint
   make build
   ```

3. **Create Pull Request**:
   ```bash
   # Use semantic title
   gh pr create --title "feat(component): add new functionality" --body "..."
   ```

4. **Wait for CI Checks**:
   - All tests must pass
   - Code coverage threshold met
   - Security scans clean
   - Linting passed

5. **Request Reviews**:
   - Assign reviewers
   - Address feedback
   - Update PR as needed

6. **Merge**:
   - Squash and merge preferred
   - Delete branch after merge

### Release Workflow

1. **Prepare Release**:
   ```bash
   # Update version and changelog
   # Ensure all tests pass
   make test-ci
   ```

2. **Create Tag**:
   ```bash
   # Use semantic versioning
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```

3. **Monitor Release**:
   ```bash
   # Watch workflow progress
   gh run watch

   # Check release
   gh release view v1.2.3
   ```

4. **Deploy**:
   ```bash
   # Deploy to staging first
   gh workflow run deploy.yml -f environment=staging -f version=v1.2.3

   # After validation, deploy to production
   gh workflow run deploy.yml -f environment=production -f version=v1.2.3
   ```

### Deployment Best Practices

1. **Always Deploy to Staging First**:
   - Test in staging environment
   - Run smoke tests
   - Verify functionality

2. **Use Dry Runs**:
   ```bash
   gh workflow run deploy.yml -f environment=production -f version=v1.2.3 -f dry_run=true
   ```

3. **Monitor Deployments**:
   - Watch logs
   - Check metrics
   - Verify health checks

4. **Rollback if Needed**:
   ```bash
   # Deploy previous version
   gh workflow run deploy.yml -f environment=production -f version=v1.2.2
   ```

### Security Best Practices

1. **Regular Security Scans**:
   - Daily automated scans
   - Review security reports
   - Address critical issues immediately

2. **Dependency Management**:
   - Keep dependencies up to date
   - Review vulnerability reports
   - Use `go mod tidy` regularly

3. **Secret Management**:
   - Never commit secrets
   - Use GitHub Secrets
   - Rotate secrets regularly

4. **Container Security**:
   - Use minimal base images
   - Run as non-root user
   - Keep images updated

## Troubleshooting

### Common Issues

#### 1. CI Failing on Dependencies

**Symptom**: `go.mod` or `go.sum` errors

**Solution**:
```bash
# Locally
go mod tidy
go mod verify

# Commit changes
git add go.mod go.sum
git commit -m "chore(deps): update go modules"
```

#### 2. Test Failures

**Symptom**: Tests failing in CI but passing locally

**Solution**:
```bash
# Run tests with race detection
go test -race ./...

# Run tests with verbose output
go test -v ./...

# Check for environment-specific issues
export SKIP_INTEGRATION=true
go test ./...
```

#### 3. Docker Build Failures

**Symptom**: Docker build failing in CI

**Solution**:
```bash
# Test locally
docker build -f Dockerfile.optimized -t freightliner:test .

# Check Dockerfile syntax
docker build --no-cache -f Dockerfile.optimized .

# Review build logs
gh run view --log-failed
```

#### 4. Security Scan Failures

**Symptom**: Security scans reporting vulnerabilities

**Solution**:
```bash
# Run gosec locally
gosec ./...

# Run govulncheck
govulncheck ./...

# Update vulnerable dependencies
go get -u ./...
go mod tidy
```

#### 5. Deployment Failures

**Symptom**: Deployment workflow failing

**Solution**:
```bash
# Check Kubernetes config
kubectl get pods -n <environment>

# Review deployment logs
kubectl logs -n <environment> deployment/freightliner

# Verify image availability
docker pull ghcr.io/$GITHUB_REPOSITORY:$VERSION

# Check environment secrets
gh secret list
```

### Getting Help

1. **View Workflow Logs**:
   ```bash
   gh run list
   gh run view <run-id> --log
   ```

2. **Check Workflow Status**:
   ```bash
   gh workflow list
   gh workflow view consolidated-ci.yml
   ```

3. **Debug Locally**:
   ```bash
   # Use act to run GitHub Actions locally
   brew install act
   act -l
   act push
   ```

4. **Contact Team**:
   - Create an issue with `ci/cd` label
   - Include workflow run URL
   - Provide relevant logs

## Advanced Configuration

### Custom Workflows

Create custom workflows in `.github/workflows/`:

```yaml
name: Custom Workflow

on:
  workflow_dispatch:

jobs:
  custom-job:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Custom step
        run: echo "Custom logic here"
```

### Composite Actions

Reusable actions in `.github/actions/`:

```yaml
name: 'My Custom Action'
description: 'Reusable action'

inputs:
  my-input:
    description: 'Input parameter'
    required: true

runs:
  using: 'composite'
  steps:
    - name: Do something
      shell: bash
      run: echo "${{ inputs.my-input }}"
```

### Environment-Specific Configuration

1. **Create environment configs**:
   ```yaml
   # config/dev.yaml
   environment: development
   replicas: 1
   resources:
     limits:
       cpu: 500m
       memory: 512Mi
   ```

2. **Use in deployment**:
   ```bash
   kubectl apply -f config/${ENVIRONMENT}.yaml
   ```

### Caching Strategy

Optimize build times with strategic caching:

```yaml
- uses: actions/cache@v4
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

### Matrix Builds

Test across multiple configurations:

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go-version: ['1.24', '1.25']
```

## Metrics and Monitoring

### Key Metrics

Track these CI/CD metrics:

- **Build Success Rate**: Percentage of successful builds
- **Build Duration**: Average time for CI pipeline
- **Deployment Frequency**: How often code is deployed
- **Mean Time to Recovery (MTTR)**: Time to recover from failures
- **Change Failure Rate**: Percentage of deployments causing failures

### Dashboards

View metrics at:
- GitHub Actions dashboard
- Workflow insights
- Security overview

### Alerts

Configure alerts for:
- Security findings (critical/high)
- Build failures on main branch
- Deployment failures
- Performance degradation

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing Guide](https://go.dev/doc/tutorial/add-a-test)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

---

**Last Updated**: 2025-12-12
**Maintained by**: Freightliner Team
