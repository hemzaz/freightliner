# GitHub Actions Workflows

This directory contains optimized GitHub Actions workflows for the Freightliner project. The workflows have been consolidated and streamlined for maximum efficiency and maintainability.

## Quick Reference

| Workflow | Purpose | Trigger | Duration |
|----------|---------|---------|----------|
| **consolidated-ci.yml** | Main CI pipeline (build, test, lint, security, docker) | Push, PR | 15-20 min |
| **security-scan.yml** | Unified security scanning (secrets, SAST, dependencies, container, IaC) | PR, Push, Call, Schedule | 10-30 min |
| **deploy.yml** | Deploy to environments (dev/staging/production) | Manual, Push to main | 15-25 min |
| **release-pipeline.yml** | Create releases with binaries and Docker images | Tag push | 20-30 min |
| **monitoring.yml** | Scheduled security and health monitoring | Daily 2 AM UTC | 30-40 min |

## Workflow Details

### 1. Consolidated CI (`consolidated-ci.yml`)

**The main continuous integration pipeline that runs on every PR and push.**

#### Jobs:
- **Build**: Compile Go application and create binary artifact
- **Test Unit**: Run unit tests on multiple OS (Ubuntu, macOS)
- **Test Integration**: Run integration tests with Docker registry service
- **Lint**: Code quality checks (gofmt, golangci-lint, go vet, go mod tidy)
- **Security**: Call security-scan workflow with quick scope
- **Docker**: Build and test Docker image
- **CI Status**: Aggregate all results and report status

#### When it runs:
- Push to `main`, `master`, `develop`, or `claude/**` branches
- Pull requests to `main`, `master`, `develop`
- Manual trigger via workflow_dispatch

#### Key features:
- ✅ Parallel job execution for speed
- ✅ Matrix strategy for multi-OS testing
- ✅ Composite actions for reusability
- ✅ Smart caching for dependencies
- ✅ GitHub Actions cache for Docker layers
- ✅ Path filtering to skip docs changes
- ✅ Concurrency control to cancel outdated runs

#### Example output:
```
✅ Build: success (3m 45s)
✅ Unit Tests (ubuntu): success (5m 12s)
✅ Unit Tests (macos): success (5m 34s)
✅ Integration Tests: success (8m 21s)
✅ Lint: success (4m 15s)
✅ Security: success (9m 48s)
✅ Docker: success (6m 30s)
✅ CI Status: success
Total: 18m 42s
```

---

### 2. Security Scan (`security-scan.yml`)

**Unified security scanning workflow that consolidates all security checks.**

#### Jobs:
- **Configure**: Determine scan scope (quick vs full)
- **Secret Scan**: TruffleHog + GitLeaks for hardcoded secrets
- **SAST Scan**: Gosec + Semgrep for code vulnerabilities
- **Dependency Scan**: govulncheck + dependency review
- **Container Scan**: Trivy for Docker image vulnerabilities (optional)
- **IaC Scan**: Checkov for infrastructure security (optional)
- **Security Gate**: Aggregate results and enforce policy

#### Scan scopes:
- **Quick** (default for PRs): Secrets + SAST + Dependencies (~10 min)
- **Full** (scheduled/manual): All scans including container + IaC (~30 min)

#### When it runs:
- Pull requests (quick scan)
- Push to main/master (quick scan)
- Called by other workflows (configurable)
- Daily schedule via monitoring.yml (full scan)
- Manual trigger (configurable scope)

#### Severity thresholds:
- **CRITICAL**: Always fails the build
- **HIGH**: Fails by default (configurable)
- **MEDIUM**: Warning only
- **LOW**: Informational

#### Example usage in other workflows:
```yaml
security:
  uses: ./.github/workflows/security-scan.yml
  permissions:
    contents: read
    security-events: write
  with:
    scan_scope: quick
    severity_threshold: HIGH
```

---

### 3. Deploy (`deploy.yml`)

**Unified deployment workflow for all environments.**

#### Jobs:
- **Build**: Build and push multi-platform Docker image
- **Deploy**: Deploy to target environment with health checks
- **Rollback**: Automatic rollback on failure

#### Environments:
- **dev**: Auto-deploy on push to main (no approval required)
- **staging**: Manual trigger (approval required)
- **production**: Manual trigger (approval required)

#### When it runs:
- Push to `main` → auto-deploy to dev
- Manual trigger with environment selection

#### Deployment flow:
```
Build Image (multi-platform)
  ↓
Security Scan (Trivy)
  ↓
Deploy to Environment
  ↓
Health Check
  ↓
Smoke Tests (staging/prod only)
  ↓
Success ✅ or Rollback ↻
```

#### Manual deployment:
1. Go to Actions → Deploy workflow
2. Click "Run workflow"
3. Select environment (dev/staging/production)
4. Enter version (or use "latest")
5. Optionally enable dry-run mode
6. Click "Run workflow"

#### Dry-run mode:
- Builds image but doesn't push
- Shows what would be deployed
- No actual changes to environment

---

### 4. Release Pipeline (`release-pipeline.yml`)

**Create GitHub releases with binaries and Docker images.**

#### Jobs:
- **Build Binaries**: Build for multiple platforms (Linux, macOS, Windows × amd64/arm64)
- **Build Docker**: Build and push multi-platform Docker images
- **Create Release**: Create GitHub release with assets and changelog
- **Notify**: Post announcement (optional)

#### When it runs:
- Push tags matching `v*.*.*` (e.g., v1.2.3)
- Manual trigger with tag name

#### Artifacts created:
- **Binaries**:
  - `freightliner-v1.2.3-linux-amd64`
  - `freightliner-v1.2.3-linux-arm64`
  - `freightliner-v1.2.3-darwin-amd64` (Intel Mac)
  - `freightliner-v1.2.3-darwin-arm64` (Apple Silicon)
  - `freightliner-v1.2.3-windows-amd64.exe`
- **Checksums**: SHA256 checksums for all binaries
- **Docker Images**:
  - `ghcr.io/hemzaz/freightliner:v1.2.3`
  - `ghcr.io/hemzaz/freightliner:1.2`
  - `ghcr.io/hemzaz/freightliner:1`
  - `ghcr.io/hemzaz/freightliner:latest`
- **SBOM**: Software Bill of Materials (CycloneDX format)

#### Release notes:
- Auto-generated changelog from commits
- Links to installation guide
- Docker pull command
- Verification instructions

---

### 5. Monitoring (`monitoring.yml`)

**Scheduled security and health monitoring.**

#### Jobs:
- **Security Monitoring**: Full security scan (calls security-scan.yml)
- **Health Monitoring**: Check endpoint health across environments
- **Dependency Monitoring**: Check for outdated dependencies
- **Monitoring Summary**: Aggregate results and create issues

#### When it runs:
- Daily at 2 AM UTC (scheduled)
- Manual trigger with scan type selection

#### Alerts:
- Creates GitHub issues for security alerts
- Reports dependency vulnerabilities
- Monitors endpoint health
- Tracks security score over time

#### Alert severity:
- **CRITICAL**: Immediate issue created
- **HIGH**: Issue created
- **MEDIUM**: Warning logged
- **LOW**: Informational

---

## Composite Actions

Reusable action components in `.github/actions/`:

### setup-go
**Purpose**: Set up Go environment with caching

**Inputs**:
- `go-version`: Go version (default: 1.25.4)
- `cache-dependency-path`: Path to go.sum
- `skip-cache`: Skip module caching

**Outputs**:
- `go-version`: Installed Go version
- `cache-hit`: Whether cache was hit

**Usage**:
```yaml
- name: Setup Go
  uses: ./.github/actions/setup-go
  with:
    go-version: '1.25.4'
```

### run-tests
**Purpose**: Run Go tests with coverage and benchmarks

**Inputs**:
- `test-type`: unit, integration, all, benchmark
- `race-detection`: Enable race detector (default: true)
- `coverage`: Generate coverage report (default: true)
- `coverage-threshold`: Minimum coverage % (default: 40)
- `timeout`: Test timeout (default: 10m)
- `packages`: Packages to test (default: ./...)

**Outputs**:
- `coverage-percentage`: Test coverage %
- `tests-passed`: Whether tests passed
- `benchmark-results`: Path to benchmark results

**Usage**:
```yaml
- name: Run tests
  uses: ./.github/actions/run-tests
  with:
    test-type: unit
    race-detection: 'true'
    coverage: 'true'
    coverage-threshold: '40'
```

---

## Reusable Workflows

### reusable-build.yml
**Purpose**: Build Go binary for specific platform

**Inputs**:
- `go-version`: Go version
- `goos`: Target OS (linux, darwin, windows)
- `goarch`: Target architecture (amd64, arm64)
- `binary-name`: Output binary name
- `version`: Version string

**Usage**:
```yaml
build-linux:
  uses: ./.github/workflows/reusable-build.yml
  with:
    go-version: '1.25.4'
    goos: linux
    goarch: amd64
    version: v1.0.0
```

### reusable-docker-publish.yml
**Purpose**: Build and publish Docker image

**Inputs**:
- `image-name`: Docker image name
- `platforms`: Target platforms
- `push`: Whether to push image
- `tags`: Image tags

### reusable-test.yml
**Purpose**: Run test suite with specific configuration

**Inputs**:
- `test-type`: Type of tests
- `go-version`: Go version
- `coverage-threshold`: Minimum coverage

---

## Workflow Triggers

### Event Triggers

| Event | Workflows | Description |
|-------|-----------|-------------|
| `push` | CI, Deploy (main only) | Code pushed to repository |
| `pull_request` | CI, Security | PR opened/updated |
| `workflow_dispatch` | All | Manual trigger from UI |
| `schedule` | Monitoring | Cron schedule (daily 2 AM UTC) |
| `push (tags)` | Release | Tag pushed (v*.*.*)  |
| `workflow_call` | Security | Called by other workflows |

### Path Filtering

Workflows automatically skip on:
- `**.md` (Markdown files)
- `docs/**` (Documentation)
- `.gitignore`, `LICENSE`
- `.github/**` (for Deploy workflow)

---

## Environment Configuration

### GitHub Environments

Required environments in GitHub settings:

1. **dev**
   - Protection rules: None (auto-deploy)
   - Secrets: `KUBE_CONFIG_DEV`
   - URL: https://dev.freightliner.example.com

2. **staging**
   - Protection rules: Required reviewers (1)
   - Secrets: `KUBE_CONFIG_STAGING`
   - URL: https://staging.freightliner.example.com

3. **production**
   - Protection rules: Required reviewers (2+)
   - Secrets: `KUBE_CONFIG_PROD`
   - URL: https://freightliner.example.com

### Required Secrets

Repository secrets needed:

- `GITHUB_TOKEN` (automatic)
- `CODECOV_TOKEN` (optional - for coverage upload)
- `SEMGREP_APP_TOKEN` (optional - for Semgrep SAST)
- `SLACK_WEBHOOK_URL` (optional - for notifications)
- `KUBE_CONFIG_DEV` (Kubernetes config for dev)
- `KUBE_CONFIG_STAGING` (Kubernetes config for staging)
- `KUBE_CONFIG_PROD` (Kubernetes config for production)

---

## Branch Protection Rules

Recommended branch protection for `main`:

- ✅ Require pull request reviews (1 reviewer)
- ✅ Require status checks:
  - `CI Status` (from consolidated-ci.yml)
  - `Security Gate` (from security-scan.yml)
- ✅ Require branches to be up to date
- ✅ Require conversation resolution
- ✅ Require signed commits (optional)
- ✅ Do not allow bypassing (except admins)

---

## Migration Guide

### From Old Workflows

If upgrading from previous workflow structure:

1. **Review OPTIMIZATION_PLAN.md** for detailed migration strategy

2. **Run migration script**:
   ```bash
   cd .github/workflows
   ./migrate-workflows.sh --dry-run    # Preview changes
   ./migrate-workflows.sh --execute    # Apply changes
   ```

3. **Update branch protection rules** to use new workflow names

4. **Test thoroughly** in feature branch before merging

5. **Monitor first few runs** for any issues

### Workflow Mapping

| Old Workflow | New Workflow | Status |
|--------------|--------------|--------|
| security-gates-enhanced.yml | security-scan.yml | ✅ Replaced |
| security-comprehensive.yml | security-scan.yml | ✅ Replaced |
| security-monitoring-enhanced.yml | monitoring.yml | ✅ Replaced |
| helm-deploy.yml | deploy.yml | ✅ Replaced |
| kubernetes-deploy.yml | deploy.yml | ✅ Replaced |
| integration-tests.yml | consolidated-ci.yml | ✅ Replaced |
| test-matrix.yml | consolidated-ci.yml | ✅ Replaced |
| reusable-security-scan.yml | security-scan.yml | ✅ Replaced |

---

## Troubleshooting

### Workflow Failed

1. **Check workflow summary** in GitHub Actions UI
2. **Review job logs** for specific failure
3. **Check if transient** (network, rate limit) → re-run
4. **Check if persistent** → create issue

### Common Issues

#### Cache Issues
```bash
# Clear GitHub Actions cache
gh cache delete --all
```

#### Docker Build Fails
- Check Dockerfile syntax
- Verify base image availability
- Check disk space on runner

#### Security Scan False Positives
- Review SARIF output in Security tab
- Add exceptions to `.gitleaks.toml` if needed
- Adjust severity threshold in workflow

#### Deployment Fails
- Verify kubeconfig secret is up to date
- Check kubectl connection
- Review deployment logs
- Check if rollback triggered

### Getting Help

1. Check workflow logs in GitHub Actions
2. Review this documentation
3. Check OPTIMIZATION_PLAN.md
4. Create issue with:
   - Workflow name
   - Job that failed
   - Error message
   - Run URL

---

## Performance Metrics

### Current Performance

| Metric | Value |
|--------|-------|
| Average PR CI time | 15-20 minutes |
| Security scan time (quick) | 10 minutes |
| Security scan time (full) | 30 minutes |
| Deployment time | 15-20 minutes |
| Release build time | 20-30 minutes |

### Optimization Improvements

| Area | Before | After | Improvement |
|------|--------|-------|-------------|
| Workflow count | 22 files | 5 files | 77% reduction |
| PR CI time | 25-30 min | 15-20 min | 33% faster |
| Security scans | 4 × 30 min | 1 × 10 min | 87% faster |
| Deployment | 20-25 min | 15-20 min | 25% faster |

---

## Best Practices

### For Developers

1. **Run tests locally** before pushing
   ```bash
   make test
   make lint
   ```

2. **Use conventional commits** for better changelogs
   ```bash
   git commit -m "feat: add new feature"
   git commit -m "fix: resolve bug"
   ```

3. **Keep PRs small** for faster CI runs

4. **Don't push WIP** to main branches

5. **Use feature branches** for workflow changes

### For Maintainers

1. **Monitor workflow execution times** weekly
2. **Review security alerts** promptly
3. **Update dependencies** regularly
4. **Test workflow changes** in feature branches
5. **Document non-obvious behavior**

---

## Contributing

When modifying workflows:

1. ✅ Read this documentation
2. ✅ Test changes in feature branch
3. ✅ Run validation script
4. ✅ Update documentation if needed
5. ✅ Get review from team
6. ✅ Monitor first runs after merge

---

## Support

- **Documentation**: This README + OPTIMIZATION_PLAN.md
- **Issues**: Create GitHub issue with `workflow` label
- **Questions**: Ask in team chat or create discussion

---

**Last Updated**: 2025-12-11
**Version**: 2.0
**Maintained by**: DevOps Team
