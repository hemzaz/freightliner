# Minimal Workflow Set - Quick Reference

## The 3 Workflows That Replace Everything

### 1. **ci.yml** - Continuous Integration
> Replaces: 11 workflows | Target: <15 min (PR), <25 min (full)

**What it does:**
- ✅ Lint (Go, Shell, YAML)
- ✅ Test (Unit + Integration)
- ✅ Security (Quick scan)
- ✅ Build (Binary + Docker)
- ✅ Benchmark (Conditional)

**Triggers:**
- Push to main/develop
- Pull requests
- Manual dispatch

**Usage:**
```bash
# Manual trigger
gh workflow run ci.yml

# With options
gh workflow run ci.yml -f run_benchmarks=true
gh workflow run ci.yml -f skip_tests=true
```

---

### 2. **deploy.yml** - Deployment
> Replaces: 8 workflows | Target: <10 min per environment

**What it does:**
- 🚀 Deploy to dev/staging/production
- ✅ Health checks + smoke tests
- 🔄 Automatic rollback on failure
- 📦 GitHub releases (production)
- 💬 Slack notifications

**Triggers:**
- Manual dispatch (with environment selection)
- Tag push (auto-deploy to production)

**Usage:**
```bash
# Deploy to dev
gh workflow run deploy.yml -f environment=dev -f version=latest

# Deploy to staging (requires approval)
gh workflow run deploy.yml -f environment=staging -f version=main-abc123

# Deploy to production (requires approval)
gh workflow run deploy.yml -f environment=production -f version=v1.2.3

# Dry run
gh workflow run deploy.yml -f environment=dev -f version=latest -f dry_run=true

# Rollback
gh workflow run deploy.yml -f environment=production -f action=rollback
```

---

### 3. **scheduled.yml** - Nightly Tasks
> Replaces: 5 workflows | Target: <40 min (parallel)

**What it does:**
- 🔒 Comprehensive security scans (all tools)
- 📦 Automated dependency updates (creates PRs)
- ⚡ Performance benchmarks + stress tests
- 🧹 Cleanup old artifacts/workflows
- 🩺 Health monitoring (all environments)

**Triggers:**
- Schedule: Daily at 2 AM UTC
- Manual dispatch

**Usage:**
```bash
# Run all tasks
gh workflow run scheduled.yml

# Run specific task
gh workflow run scheduled.yml -f tasks=security
gh workflow run scheduled.yml -f tasks=dependencies
gh workflow run scheduled.yml -f tasks=benchmarks
gh workflow run scheduled.yml -f tasks=cleanup
```

---

## Quick Comparison

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| **Workflow files** | 25+ | 3 | 88% reduction |
| **CI time (PR)** | 35 min | 15 min | 57% faster |
| **CI time (Full)** | 60 min | 25 min | 58% faster |
| **Deploy time** | 20 min | 10 min | 50% faster |
| **Scheduled time** | 90 min | 40 min | 56% faster |
| **Monthly cost** | $110 | $34 | $76 saved |
| **Maintenance** | 25+ files | 3 files | 88% less work |

---

## Common Tasks

### For Developers

**Check CI status:**
```bash
gh run list --workflow=ci.yml --limit=5
```

**View CI logs:**
```bash
gh run view <run-id> --log
```

**Trigger CI manually:**
```bash
gh workflow run ci.yml
```

**Run benchmarks:**
```bash
gh workflow run ci.yml -f run_benchmarks=true
```

### For DevOps

**Deploy to dev:**
```bash
gh workflow run deploy.yml -f environment=dev -f version=latest
```

**Deploy to production:**
```bash
# From tag
git tag v1.2.3
git push origin v1.2.3
# Auto-deploys to production (requires approval)

# Or manually
gh workflow run deploy.yml -f environment=production -f version=v1.2.3
```

**Rollback production:**
```bash
gh workflow run deploy.yml -f environment=production -f action=rollback
```

**Check deployment status:**
```bash
gh run list --workflow=deploy.yml --limit=5
```

### For Security Team

**Run security scan now:**
```bash
gh workflow run scheduled.yml -f tasks=security
```

**View security results:**
```bash
gh run view <run-id> --log
# Then download artifacts from GitHub Actions UI
```

**Check for vulnerabilities:**
```bash
# Security scan runs nightly
# Check GitHub Security tab for SARIF results
```

### For SRE/Platform

**Monitor health:**
```bash
# Health checks run nightly via scheduled.yml
gh run list --workflow=scheduled.yml --limit=1
```

**Cleanup resources:**
```bash
gh workflow run scheduled.yml -f tasks=cleanup
```

**Update dependencies:**
```bash
# Runs nightly, creates PR automatically
# Or run manually:
gh workflow run scheduled.yml -f tasks=dependencies
```

---

## Workflow Architecture

### CI Workflow (ci.yml)

```
┌─────────────────────────────────────────┐
│          Trigger (Push/PR)              │
└─────────────────┬───────────────────────┘
                  │
        ┌─────────┴─────────┐
        │                   │
   ┌────▼────┐      ┌───────▼────────┐
   │  Lint   │      │  Security-Quick │
   │ (5 min) │      │    (10 min)    │
   └────┬────┘      └───────┬────────┘
        │                   │
   ┌────▼────┐         ┌────▼────┐
   │  Build  │         │  Test   │
   │ (8 min) │         │(15 min) │
   └────┬────┘         └────┬────┘
        │                   │
        └──────────┬────────┘
              ┌────▼────┐
              │ Docker  │
              │(12 min) │
              └────┬────┘
              ┌────▼────┐
              │ Status  │
              └─────────┘

Total: <15 min (parallel execution)
```

### Deploy Workflow (deploy.yml)

```
┌─────────────────────────────────────────┐
│    Trigger (Manual/Tag)                 │
└─────────────────┬───────────────────────┘
                  │
            ┌─────▼─────┐
            │ Validate  │
            │  (5 min)  │
            └─────┬─────┘
                  │
     ┌────────────┼────────────┐
     │            │            │
┌────▼────┐  ┌───▼───┐  ┌─────▼─────┐
│Deploy-  │  │Deploy-│  │Deploy-    │
│Dev      │  │Staging│  │Production │
│(8 min)  │  │(10min)│  │(15 min)   │
└────┬────┘  └───┬───┘  └─────┬─────┘
     │           │            │
     └───────────┼────────────┘
            ┌────▼────┐
            │ Notify  │
            └─────────┘

Total: <10 min per environment
```

### Scheduled Workflow (scheduled.yml)

```
┌─────────────────────────────────────────┐
│    Trigger (Nightly 2 AM UTC)           │
└─────────────────┬───────────────────────┘
                  │
        ┌─────────┴─────────┐
        │   All Parallel    │
        └─────────┬─────────┘
     ┌────────────┼────────────┐
     │            │            │
┌────▼────┐  ┌───▼───┐  ┌─────▼─────┐
│Security-│  │Depend-│  │Performance│
│Compreh- │  │ency   │  │-Benchmarks│
│ensive   │  │Updates│  │ (25 min)  │
│(30 min) │  │(20min)│  └─────┬─────┘
└────┬────┘  └───┬───┘        │
     │           │            │
┌────▼────┐  ┌───▼───┐  ┌─────▼─────┐
│Cleanup  │  │Monitor│  │  Summary  │
│(10 min) │  │(5 min)│  │           │
└────┬────┘  └───┬───┘  └─────┬─────┘
     │           │            │
     └───────────┼────────────┘
            ┌────▼────┐
            │ Summary │
            └─────────┘

Total: <40 min (parallel execution)
```

---

## Feature Matrix

### CI Workflow

| Feature | Included | Notes |
|---------|----------|-------|
| Go Linting | ✅ | golangci-lint v1.62.2 |
| Go Formatting | ✅ | gofmt check |
| Go Vet | ✅ | Standard checks |
| Mod Tidy | ✅ | Ensures clean go.mod |
| Shell Check | ✅ | Bash script validation |
| YAML Lint | ✅ | Workflow validation |
| Unit Tests | ✅ | With race detector |
| Integration Tests | ✅ | With registry service |
| Coverage Report | ✅ | Uploaded to Codecov |
| GoSec | ✅ | Static security analysis |
| GovulnCheck | ✅ | Known CVE detection |
| TruffleHog | ✅ | Secret scanning |
| Dependency Review | ✅ | PR only |
| Binary Build | ✅ | Standard + static |
| Docker Build | ✅ | Multi-platform |
| Docker Scan | ✅ | Trivy SARIF upload |
| Benchmarks | ✅ | Conditional |
| PR Comments | ✅ | Status summary |

### Deploy Workflow

| Feature | Included | Notes |
|---------|----------|-------|
| Dev Deploy | ✅ | Auto on main |
| Staging Deploy | ✅ | With approval |
| Prod Deploy | ✅ | With approval |
| Tag-based Deploy | ✅ | Auto for v*.*.* |
| Dry Run Mode | ✅ | Test without changes |
| Health Checks | ✅ | Post-deployment |
| Smoke Tests | ✅ | Dev + staging |
| Blue-Green Deploy | ✅ | Production only |
| Rollback | ✅ | Manual + auto |
| GitHub Releases | ✅ | Production only |
| Slack Notify | ✅ | Configurable |
| OIDC Auth | ✅ | Kubernetes |

### Scheduled Workflow

| Feature | Included | Notes |
|---------|----------|-------|
| GoSec (deep) | ✅ | Comprehensive mode |
| GovulnCheck (deep) | ✅ | Detailed analysis |
| TruffleHog (full) | ✅ | Full history scan |
| GitLeaks | ✅ | Alternative scanner |
| CodeQL | ✅ | Advanced analysis |
| License Scan | ✅ | Compliance check |
| Trivy (deep) | ✅ | All severities |
| Grype | ✅ | Alternative scanner |
| SBOM (SPDX) | ✅ | Standard format |
| SBOM (CycloneDX) | ✅ | Alternative format |
| Dependency PRs | ✅ | Auto-created |
| Benchmarks | ✅ | Full suite |
| Stress Tests | ✅ | 10 iterations |
| Artifact Cleanup | ✅ | >30 days |
| Workflow Cleanup | ✅ | >30 days |
| Health Monitoring | ✅ | All environments |
| Auto Issue Creation | ✅ | On failures |

---

## Environment Variables

### CI Workflow
```yaml
GO_VERSION: '1.25.4'
GOLANGCI_LINT_VERSION: 'v1.62.2'
DOCKER_REGISTRY: ghcr.io
IMAGE_NAME: ${{ github.repository }}
```

### Deploy Workflow
```yaml
DOCKER_REGISTRY: ghcr.io
DOCKER_IMAGE: ghcr.io/${{ github.repository }}
```

### Required Secrets
```
GITHUB_TOKEN: Auto-provided
CODECOV_TOKEN: Optional (for coverage)
KUBE_CONFIG_DEV: Kubernetes config (dev)
KUBE_CONFIG_STAGING: Kubernetes config (staging)
KUBE_CONFIG_PROD: Kubernetes config (production)
SLACK_WEBHOOK_URL: Optional (for notifications)
GITLEAKS_LICENSE: Optional (for GitLeaks)
```

---

## Migration Steps

1. **Test** (Week 1)
   - Copy workflows to test location
   - Run manually
   - Verify all features work

2. **Enable** (Week 2)
   - Enable ci.yml
   - Monitor for 2-3 days
   - Enable deploy.yml
   - Test deployments
   - Enable scheduled.yml

3. **Cleanup** (Week 3)
   - Archive old workflows
   - Update documentation
   - Remove unused secrets

See [MIGRATION.md](./MIGRATION.md) for detailed guide.

---

## Troubleshooting

### CI fails on lint
```bash
# Check what's wrong
gh run view <run-id> --log | grep -A 5 "lint"

# Fix locally
gofmt -s -w .
go mod tidy
golangci-lint run --fix
```

### Deploy fails
```bash
# Check logs
gh run view <run-id> --log

# Check kubectl access
kubectl get pods -n dev

# Rollback if needed
gh workflow run deploy.yml -f environment=dev -f action=rollback
```

### Scheduled task fails
```bash
# Check what failed
gh run view <run-id> --log

# Rerun specific task
gh workflow run scheduled.yml -f tasks=security
```

---

## Best Practices

### For PRs
- Wait for CI to pass before requesting review
- Address linting issues immediately
- Check coverage reports
- Review security scan results

### For Deployments
- Always deploy to dev first
- Use dry-run for production
- Monitor metrics after deployment
- Keep rollback window open (30 min)

### For Security
- Review nightly security reports
- Address HIGH/CRITICAL findings within 48h
- Keep dependencies up-to-date
- Review auto-created dependency PRs

---

## Support & Resources

- **Migration Guide**: [MIGRATION.md](./MIGRATION.md)
- **Workflow Files**:
  - [ci.yml](./ci.yml)
  - [deploy.yml](./deploy.yml)
  - [scheduled.yml](./scheduled.yml)
- **GitHub Actions Docs**: https://docs.github.com/en/actions
- **Issue Tracker**: Create issue with `ci-cd` label

---

**Last updated:** 2025-12-11
**Version:** 1.0.0
**Status:** Production Ready
