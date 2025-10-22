# CI/CD Management Command

Manage and troubleshoot GitHub Actions CI/CD pipelines for Freightliner.

## What This Command Does

1. Checks workflow status and health
2. Troubleshoots failed pipeline runs
3. Triggers manual workflow runs
4. Views and downloads artifacts
5. Manages workflow configurations

## Usage

```bash
/ci-cd [action] [workflow]
```

## Actions

- `status` - Check status of recent workflow runs
- `debug` - Debug failed workflow runs
- `run` - Trigger manual workflow run
- `artifacts` - Download workflow artifacts
- `logs` - View workflow logs
- `update` - Update workflow configuration

## Examples

```bash
# Check status
/ci-cd status main-ci

# Debug failed run
/ci-cd debug main-ci

# Trigger manual run
/ci-cd run release-pipeline

# Download artifacts
/ci-cd artifacts latest

# View logs
/ci-cd logs failed
```

## Available Workflows

### Main CI Pipeline
- **File**: `.github/workflows/main-ci.yml`
- **Triggers**: Push to main/develop, PRs
- **Jobs**: Build, Test, Lint, Security, Docker
- **Duration**: ~15 minutes

### Release Pipeline
- **File**: `.github/workflows/release-pipeline.yml`
- **Triggers**: Version tags (v*.*.*)
- **Jobs**: Multi-platform builds, Docker, Release
- **Duration**: ~30 minutes

## Status Check Tasks

When checking status:
- List recent workflow runs
- Show success/failure counts
- Display duration statistics
- Identify failing jobs
- Show error messages

```bash
# Using GitHub CLI
gh run list --workflow=main-ci.yml --limit=10

# View specific run
gh run view <run-id>

# Watch live run
gh run watch
```

## Debug Tasks

When debugging failures:

### 1. Identify Failure Type
```bash
# Get run details
gh run view <run-id> --log-failed

# Common failure types:
# - Test failures
# - Lint errors
# - Security vulnerabilities
# - Build errors
# - Timeout issues
```

### 2. Analyze Logs
```bash
# Download logs
gh run download <run-id>

# View specific job logs
gh run view <run-id> --job=<job-id> --log
```

### 3. Common Issues

#### Test Failures
```bash
# Reproduce locally
go test -v -short -race ./...

# Check specific package
go test -v ./pkg/problematic/package/

# Run with coverage
go test -v -coverprofile=coverage.out ./...
```

#### Lint Failures
```bash
# Run locally
golangci-lint run --config=.golangci.yml

# Fix formatting
make fmt

# Check specific linter
golangci-lint run --enable-only=errcheck
```

#### Security Scan Failures
```bash
# Run gosec locally
make security

# Check vulnerabilities
govulncheck ./...

# Review dependencies
go list -m all
```

#### Docker Build Failures
```bash
# Build locally
docker build -t freightliner:test .

# Check specific stage
docker build --target builder -t test .

# Verbose output
docker build --progress=plain .
```

#### Timeout Issues
```bash
# Check workflow timeout settings
# In workflow file:
timeout-minutes: 15

# Increase if needed for specific jobs
# Monitor execution time:
gh run list --workflow=main-ci.yml | awk '{print $6}'
```

### 4. Fix Implementation

Based on failure type:

**Code Issues**:
```bash
# Fix code
vim pkg/problematic/file.go

# Test locally
make test-ci

# Verify fix
make quality
```

**Configuration Issues**:
```bash
# Update workflow
vim .github/workflows/main-ci.yml

# Validate YAML
yamllint .github/workflows/main-ci.yml

# Test with workflow dispatch
gh workflow run main-ci.yml
```

**Dependency Issues**:
```bash
# Update dependencies
go get -u ./...
go mod tidy

# Verify
go mod verify
```

## Trigger Manual Run

```bash
# Trigger main CI
gh workflow run main-ci.yml

# Trigger release with version
gh workflow run release-pipeline.yml -f tag=v1.0.0

# Watch progress
gh run watch
```

## Artifacts Management

```bash
# List artifacts from run
gh run view <run-id> --artifacts

# Download all artifacts
gh run download <run-id>

# Download specific artifact
gh run download <run-id> -n coverage-report

# Latest successful run
gh run download --name=binary-linux-amd64
```

### Common Artifacts

| Artifact | Content | Retention |
|----------|---------|-----------|
| `freightliner-{sha}` | Compiled binary | 7 days |
| `coverage-report` | Test coverage | 30 days |
| `benchmark-results` | Performance data | 30 days |
| `docker-image` | Docker image | 7 days |
| `binary-{os}-{arch}` | Release binaries | 90 days |

## Workflow Configuration

### Environment Variables

Edit `.github/workflows/main-ci.yml`:
```yaml
env:
  GO_VERSION: '1.24.5'
  GOLANGCI_LINT_VERSION: 'v1.62.2'
```

### Workflow Triggers

```yaml
on:
  push:
    branches: [main, develop, 'claude/**']
  pull_request:
    branches: [main, develop]
  workflow_dispatch:  # Manual trigger
```

### Job Timeouts

```yaml
jobs:
  test:
    timeout-minutes: 15  # Adjust as needed
```

### Caching Configuration

```yaml
- uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

## Performance Optimization

### Current Metrics
- **Main CI**: ~15 minutes average
- **Release**: ~30 minutes average
- **Cache Hit Rate**: ~70%

### Optimization Tips

1. **Enable Caching**:
```yaml
- uses: actions/setup-go@v5
  with:
    cache: true
```

2. **Parallel Jobs**:
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest]
```

3. **Conditional Execution**:
```yaml
if: github.event_name == 'pull_request'
```

4. **Artifact Retention**:
```yaml
retention-days: 7  # Reduce storage
```

## Monitoring Dashboard

View workflow metrics:
```bash
# Success rate
gh run list --workflow=main-ci.yml | grep -c "completed"

# Average duration
gh run list --workflow=main-ci.yml --json durationMs | jq '.[] | .durationMs' | awk '{sum+=$1; count++} END {print sum/count/1000/60 " minutes"}'

# Recent failures
gh run list --workflow=main-ci.yml --status=failure --limit=5
```

## Troubleshooting Guide

### Workflow Not Triggering

1. Check branch protection rules
2. Verify workflow file syntax
3. Check GitHub Actions enabled
4. Review workflow triggers

### Permission Errors

```yaml
permissions:
  contents: read
  packages: write
  security-events: write
```

### Cache Issues

```bash
# Clear cache
# Go to Actions > Caches > Delete

# Verify cache key
echo "${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}"
```

### Secret Issues

```bash
# Check secret availability
# Settings > Secrets and variables > Actions

# Required secrets (all optional):
# - CODECOV_TOKEN (for coverage)
```

## Best Practices

1. **Monitor regularly**: Check workflow status daily
2. **Fix fast**: Address failures within 24 hours
3. **Keep updated**: Update dependencies monthly
4. **Document issues**: Add comments in PRs
5. **Test locally**: Always test before pushing

## Documentation

- [CI/CD Pipeline Guide](../docs/CI_CD_PIPELINE.md)
- [GitHub Actions Skill](.claude/skills/github-actions.md)
- [GitHub Actions Docs](https://docs.github.com/en/actions)

## Integration with Claude Code

This command integrates with:
- `/fix-ci`: Automated CI fixing
- `/security-audit`: Security scanning
- `/performance-test`: Benchmarking

## Example Workflow

```bash
# 1. Check status
/ci-cd status main-ci

# 2. If failed, debug
/ci-cd debug latest

# 3. Fix issues locally
make test-ci
make quality

# 4. Commit and push
git commit -m "fix: resolve CI failures"
git push

# 5. Monitor run
/ci-cd status main-ci
```
