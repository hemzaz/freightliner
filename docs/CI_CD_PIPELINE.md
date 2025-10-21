# CI/CD Pipeline Documentation

Complete guide to Freightliner's continuous integration and deployment pipelines.

## Overview

Freightliner uses GitHub Actions for CI/CD with **cloud-account-free** workflows that work without AWS or GCP credentials.

### Pipeline Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Trigger (Push/PR/Tag)                   │
└──────────────────┬──────────────────────────────────────────┘
                   │
         ┌─────────┴─────────┐
         │                   │
         ▼                   ▼
    ┌────────┐          ┌────────┐
    │ Main   │          │Release │
    │   CI   │          │Pipeline│
    └────────┘          └────────┘
         │                   │
    ┌────┴────┐         ┌────┴────┐
    ▼    ▼    ▼         ▼    ▼    ▼
  Build Test Docker  Binaries Multi Release
  Lint Security     Docker  Platform
```

## Main CI Pipeline

**File**: `.github/workflows/main-ci.yml`

**Triggers**:
- Push to `main`, `develop`, or `claude/**` branches
- Pull requests to `main` or `develop`
- Manual workflow dispatch

### Jobs Overview

| Job | Purpose | Runs On | Timeout |
|-----|---------|---------|---------|
| **build** | Compile application | ubuntu-latest | 10 min |
| **test** | Run unit tests | ubuntu, macos (matrix) | 15 min |
| **lint** | Code quality checks | ubuntu-latest | 10 min |
| **security** | Security scanning | ubuntu-latest | 15 min |
| **docker** | Build Docker image | ubuntu-latest | 20 min |
| **benchmark** | Performance tests | ubuntu-latest | 20 min |
| **status** | Overall CI status | ubuntu-latest | 5 min |

### 1. Build Job

Compiles the Go application and generates binaries.

```yaml
Steps:
1. Checkout code
2. Set up Go (with caching)
3. Download dependencies
4. Build application (make build)
5. Build static binary
6. Upload binary artifact
```

**Artifacts**:
- `freightliner-{sha}`: Compiled binary (7 days retention)

### 2. Test Job

Runs unit tests across multiple OS and Go versions.

**Matrix**:
- OS: `ubuntu-latest`, `macos-latest`
- Go: `1.23.4`, `1.24.5`

```bash
# Test command
go test -v -short -race -coverprofile=coverage.out -covermode=atomic ./...
```

**Environment Variables**:
- `SKIP_INTEGRATION=true`: Skips integration tests
- `SKIP_CLOUD_TESTS=true`: Skips cloud-dependent tests

**Coverage**:
- Uploaded to Codecov (ubuntu + Go 1.24.5 only)
- HTML report generated
- 80%+ coverage target

**Artifacts**:
- `coverage-report`: Coverage files (30 days)

### 3. Lint Job

Code quality and formatting checks.

```yaml
Checks:
1. golangci-lint (v1.62.2)
   - errcheck, govet, ineffassign, misspell
2. go vet
3. gofmt formatting
4. go mod tidy verification
```

**Linters Enabled**:
- `errcheck`: Unchecked errors
- `govet`: Standard Go vet checks
- `ineffassign`: Ineffectual assignments
- `misspell`: Spelling mistakes

### 4. Security Job

Security scanning and vulnerability detection.

```yaml
Scans:
1. gosec - SAST security scanner
2. govulncheck - Known vulnerability checker
3. Dependency review (PRs only)
```

**Output**:
- SARIF results uploaded to GitHub Security
- Vulnerability reports
- Dependency analysis

**Artifacts**:
- `gosec-results.sarif`: Security scan results

### 5. Docker Job

Builds and tests Docker image.

```yaml
Steps:
1. Set up Docker Buildx
2. Build multi-stage Docker image
3. Test image (version, non-root user)
4. Scan with Trivy (vulnerability scanner)
5. Upload scan results
```

**Image Tags**:
- `ghcr.io/hemzaz/freightliner:{sha}`
- `ghcr.io/hemzaz/freightliner:latest`

**Caching**: GitHub Actions cache for faster builds

**Artifacts**:
- `docker-image`: Saved image (main branch only, 7 days)
- `trivy-results.sarif`: Vulnerability scan

### 6. Benchmark Job

Performance benchmarking (PRs and main branch).

```bash
# Benchmark command
go test -bench=. -benchmem -count=3 -run=^$ ./...
```

**Artifacts**:
- `benchmark-results`: Performance data (30 days)

### 7. Status Job

Aggregates results and reports overall status.

```yaml
Checks:
- All jobs completed successfully
- Posts PR comment with status table
- Fails if any critical job failed
```

**PR Comment Example**:
```markdown
## ✅ CI Pipeline Status

| Job | Status |
|-----|--------|
| Build | ✅ success |
| Test | ✅ success |
| Lint | ✅ success |
| Security | ✅ success |
| Docker | ✅ success |
```

## Release Pipeline

**File**: `.github/workflows/release-pipeline.yml`

**Triggers**:
- Push tags matching `v*.*.*` (e.g., v1.0.0)
- Manual workflow dispatch

### Jobs Overview

| Job | Purpose | Platform |
|-----|---------|----------|
| **build-binaries** | Multi-platform binaries | Matrix |
| **build-docker** | Multi-arch Docker images | ubuntu |
| **create-release** | GitHub Release | ubuntu |
| **notify** | Release announcements | ubuntu |

### 1. Build Binaries Job

Creates binaries for multiple platforms.

**Matrix**:
```
- linux/amd64
- linux/arm64
- darwin/amd64 (Intel Mac)
- darwin/arm64 (Apple Silicon)
- windows/amd64
```

**Build Command**:
```bash
CGO_ENABLED=0 go build \
  -ldflags="-w -s \
    -X main.version=${VERSION} \
    -X main.buildTime=${BUILD_TIME} \
    -X main.gitCommit=${GIT_COMMIT}" \
  -o freightliner-${VERSION}-${OS}-${ARCH}
```

**Features**:
- Static binaries (CGO_ENABLED=0)
- Version information embedded
- SHA256 checksums generated

**Artifacts**:
- Binaries for each platform
- Checksums for verification

### 2. Build Docker Job

Multi-platform Docker images.

**Platforms**:
- `linux/amd64`
- `linux/arm64`

**Tags Generated**:
```
ghcr.io/hemzaz/freightliner:{version}
ghcr.io/hemzaz/freightliner:{major}.{minor}
ghcr.io/hemzaz/freightliner:{major}
ghcr.io/hemzaz/freightliner:latest
```

**Additional Features**:
- SBOM generation (Trivy)
- Multi-arch manifest
- GitHub Container Registry push

### 3. Create Release Job

Creates GitHub Release with all assets.

**Release Assets**:
- All platform binaries
- Combined checksums file
- SBOM (Software Bill of Materials)

**Release Notes Include**:
- Changelog since previous version
- Installation instructions
- Docker pull command
- Verification instructions

### 4. Notify Job

Post-release notifications.

**Actions**:
- Creates discussion announcement (if enabled)
- Posts release status
- Logs download links

## Configuration

### Environment Variables

```yaml
GO_VERSION: '1.24.5'
GOLANGCI_LINT_VERSION: 'v1.62.2'
REGISTRY: ghcr.io
IMAGE_NAME: hemzaz/freightliner
```

### Secrets Required

| Secret | Purpose | Required For |
|--------|---------|--------------|
| `GITHUB_TOKEN` | Automatic | All workflows (auto-provided) |
| `CODECOV_TOKEN` | Coverage upload | Optional (test job) |

**Note**: No cloud credentials required! All workflows run without AWS/GCP access.

## Skipping Cloud Tests

Tests are automatically configured to skip cloud integration tests:

```bash
# In test job
go test -v -short -race ./...

# Environment variables set
SKIP_INTEGRATION=true
SKIP_CLOUD_TESTS=true
```

**Test Tags**:
```go
// +build !integration

// Skip this test if -short flag is used
func TestCloudIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    // ... test code
}
```

## Workflow Triggers

### Automatic Triggers

```yaml
# Push to specific branches
push:
  branches: [main, develop, 'claude/**']

# Pull requests
pull_request:
  branches: [main, develop]

# Tag pushes
push:
  tags: ['v*.*.*']

# Scheduled runs
schedule:
  - cron: '0 2 * * *'  # Daily at 2 AM UTC
```

### Manual Triggers

```bash
# Trigger workflow manually
gh workflow run main-ci.yml

# Trigger release workflow
gh workflow run release-pipeline.yml -f tag=v1.0.0
```

## Concurrency Control

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

**Behavior**: Cancels in-progress runs when new commit is pushed to same branch.

## Caching Strategy

### Go Module Cache
```yaml
- uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

### Docker Build Cache
```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

**Benefits**:
- Faster dependency downloads
- Faster builds (up to 70% reduction)
- Lower GitHub Actions minutes usage

## Troubleshooting

### Test Failures

```bash
# Run tests locally with same flags
go test -v -short -race -coverprofile=coverage.out ./...

# Check for race conditions
go test -race ./...

# Verbose output
go test -v ./pkg/specific/package/
```

### Lint Failures

```bash
# Run locally
golangci-lint run --config=.golangci.yml

# Auto-fix formatting
make fmt

# Check specific linter
golangci-lint run --enable-only=errcheck
```

### Docker Build Failures

```bash
# Build locally
docker build -t freightliner:test .

# Test specific stage
docker build --target builder -t freightliner:builder .

# Check logs
docker build --progress=plain -t freightliner:test .
```

### Security Scan Failures

```bash
# Run gosec locally
make security

# Run govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Performance Optimization

### Workflow Execution Time

**Current Targets**:
- Main CI: < 15 minutes
- Release Pipeline: < 30 minutes

**Optimization Techniques**:
1. **Parallel Jobs**: Independent jobs run simultaneously
2. **Caching**: Go modules and Docker layers cached
3. **Matrix Strategy**: Tests run in parallel across OS/versions
4. **Artifact Sharing**: Binaries built once, shared across jobs

### Resource Usage

**GitHub Actions Minutes**:
- Main CI (per run): ~20-30 minutes (including matrix)
- Release Pipeline: ~45-60 minutes

**Cost Optimization**:
- Caching reduces by ~70%
- Concurrency cancellation prevents duplicate runs
- Conditional jobs (benchmarks only on PRs/main)

## Best Practices

### 1. Always Use Caching
```yaml
- uses: actions/setup-go@v5
  with:
    cache: true  # Enable automatic caching
```

### 2. Set Timeouts
```yaml
jobs:
  test:
    timeout-minutes: 15  # Prevent hanging jobs
```

### 3. Use Continue-on-Error Wisely
```yaml
- name: Upload coverage
  continue-on-error: true  # Don't fail on optional steps
```

### 4. Fail Fast When Appropriate
```yaml
strategy:
  fail-fast: false  # Don't stop matrix on first failure
```

### 5. Use Artifacts Efficiently
```yaml
retention-days: 7  # Keep only as long as needed
```

## Maintenance

### Updating Go Version

1. Update `.github/workflows/main-ci.yml`:
```yaml
env:
  GO_VERSION: '1.25.0'  # New version
```

2. Update `.github/workflows/release-pipeline.yml`
3. Update `go.mod`:
```go
go 1.25
```

### Updating Dependencies

```bash
# Update Go dependencies
go get -u ./...
go mod tidy

# Update GitHub Actions
# Check: https://github.com/actions/checkout/releases
# Update version in workflow files
```

### Adding New Checks

1. Add job to `main-ci.yml`
2. Add to `status` job dependencies
3. Test with workflow dispatch
4. Monitor execution time

## Monitoring

### Workflow Status

```bash
# List recent runs
gh run list --workflow=main-ci.yml

# View specific run
gh run view <run-id>

# Download artifacts
gh run download <run-id>
```

### Success Metrics

**Target SLAs**:
- Success Rate: > 95%
- Average Duration: < 15 min
- Build Time: < 5 min
- Test Time: < 10 min

### Dashboard

GitHub Actions provides built-in dashboards:
- Actions tab: Workflow runs
- Insights tab: Usage statistics
- Security tab: Scan results

## Future Enhancements

### Planned Improvements

1. **Code Coverage Enforcement**: Block PRs < 80%
2. **Performance Regression Detection**: Compare benchmarks
3. **Automated Dependency Updates**: Dependabot PRs
4. **Enhanced Security Scans**: SAST/DAST integration
5. **Cloud Integration Tests**: Optional with credentials

### Optional Features

- **Slack Notifications**: Release announcements
- **Auto-merge**: Dependabot PRs
- **Nightly Builds**: Latest development builds
- **Performance Dashboards**: Historical benchmark tracking

---

**Need Help?**
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Go Testing Guide](https://golang.org/doc/tutorial/add-a-test)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
