# GitHub Actions Workflows

This directory contains all CI/CD workflows for the Freightliner project.

## 🚀 Active Production Workflows

### Continuous Integration

#### `consolidated-ci.yml` ⭐ **PRIMARY CI PIPELINE**
**Purpose:** Complete CI pipeline with build, test, lint, security, and Docker
**Trigger:** Push to main/develop, pull requests
**Duration:** ~10-12 minutes

**Jobs:**
- Setup & caching
- Build (Go binary + static)
- Unit tests (matrix: 2 OS × 2 Go versions)
- Integration tests (with registry service)
- Linting & formatting
- Security scanning (gosec, govulncheck, Trivy)
- Docker build & scan
- Benchmark tests
- Status reporting & PR comments

**Use this instead of:** `main-ci.yml`, `ci-optimized.yml`, `security-hardened-ci.yml`

### Continuous Deployment

#### `docker-publish.yml`
**Purpose:** Build, sign, and publish multi-platform Docker images
**Trigger:** Push to main, version tags, manual dispatch
**Duration:** ~8-10 minutes

**Features:**
- Multi-platform builds (linux/amd64, linux/arm64)
- Image signing with Cosign
- SBOM generation (SPDX format)
- Trivy vulnerability scanning
- Multiple tagging strategies

#### `kubernetes-deploy.yml`
**Purpose:** Deploy to Kubernetes clusters
**Trigger:** Manual dispatch
**Duration:** ~8-10 minutes
**Environments:** dev, staging, production

**Features:**
- Auto-generated manifests
- Zero-downtime rolling updates
- Health check verification
- HPA configuration
- Dry-run support

#### `helm-deploy.yml`
**Purpose:** Deploy using Helm charts
**Trigger:** Manual dispatch
**Duration:** ~10-12 minutes
**Environments:** dev, staging, production

**Features:**
- Automatic chart generation
- Environment-specific values
- Atomic deployments (auto-rollback)
- Release history tracking

#### `rollback.yml`
**Purpose:** Emergency rollback with verification
**Trigger:** Manual dispatch
**Duration:** ~12-15 minutes
**RTO:** < 15 minutes

**Features:**
- Pre-rollback backup
- Kubernetes & Helm support
- Post-rollback verification
- Smoke test execution
- Incident creation on failure

---

## 📋 Legacy Workflows (Deprecated)

The following workflows are deprecated and will be removed:

### ⚠️ `main-ci.yml` - DEPRECATED
**Replaced by:** `consolidated-ci.yml`
**Reason:** Consolidated into single optimized pipeline

### ⚠️ `ci-optimized.yml` - DEPRECATED
**Replaced by:** `consolidated-ci.yml`
**Reason:** Merged with main CI for consistency

### ⚠️ `security-hardened-ci.yml` - DEPRECATED
**Replaced by:** `consolidated-ci.yml`
**Reason:** Security features integrated into main pipeline

---

## 🔧 Composite Actions

Reusable workflow components in `.github/actions/`:

### `setup-go/`
Go environment setup with dependency caching

**Usage:**
```yaml
- uses: ./.github/actions/setup-go
  with:
    go-version: '1.25.4'
```

### `run-tests/`
Comprehensive test execution with coverage

**Usage:**
```yaml
- uses: ./.github/actions/run-tests
  with:
    test-type: unit
    coverage: 'true'
    coverage-threshold: '70'
```

---

## 🎯 Quick Commands

### Run CI Pipeline
```bash
# Automatic on push
git push origin main

# Manual trigger
gh workflow run consolidated-ci.yml
```

### Deploy to Development
```bash
gh workflow run kubernetes-deploy.yml \
  -f environment=dev \
  -f dry-run=false
```

### Deploy to Production
```bash
gh workflow run helm-deploy.yml \
  -f environment=production \
  -f image-tag=v1.2.3
```

### Emergency Rollback
```bash
gh workflow run rollback.yml \
  -f environment=production \
  -f deployment-type=helm \
  -f reason="Critical bug detected"
```

### Publish Docker Image
```bash
# Tag-based
git tag v1.2.3
git push origin v1.2.3

# Manual
gh workflow run docker-publish.yml \
  -f environment=production
```

---

## 📊 Workflow Performance

| Workflow | Duration | Parallel Jobs | Cache Hit |
|----------|----------|---------------|-----------|
| consolidated-ci | 10-12 min | 6-8 jobs | 90% |
| docker-publish | 8-10 min | 1 job | 85% |
| kubernetes-deploy | 8-10 min | 1 job | - |
| helm-deploy | 10-12 min | 1 job | - |
| rollback | 12-15 min | 1 job | - |

---

## 🔐 Required Secrets

### GitHub (Automatic)
- `GITHUB_TOKEN` - Provided by GitHub Actions

### Optional
- `CODECOV_TOKEN` - For coverage uploads

### Cloud Providers (Choose one)

**AWS EKS:**
```bash
gh secret set AWS_ROLE_ARN
gh secret set AWS_REGION
gh secret set EKS_CLUSTER_NAME
```

**GCP GKE:**
```bash
gh secret set GCP_WORKLOAD_IDENTITY_PROVIDER
gh secret set GCP_SERVICE_ACCOUNT
gh secret set GKE_CLUSTER_NAME
gh secret set GCP_REGION
```

**Azure AKS:**
```bash
gh secret set AZURE_CREDENTIALS
gh secret set AKS_RESOURCE_GROUP
gh secret set AKS_CLUSTER_NAME
```

---

## 📚 Documentation

- **Complete Guide:** `/docs/CI-CD-VALIDATION-REPORT.md`
- **Quick Start:** `/docs/DEPLOYMENT-QUICK-START.md`
- **Composite Actions:** `/.github/actions/*/action.yml`

---

## 🆘 Troubleshooting

### Workflow Fails
```bash
# View workflow runs
gh run list --limit 10

# View specific run
gh run view <run-id>

# View logs
gh run view <run-id> --log
```

### Check Workflow Status
```bash
# Watch running workflow
gh run watch

# List workflows
gh workflow list

# View workflow details
gh workflow view consolidated-ci.yml
```

### Re-run Failed Jobs
```bash
# Re-run failed jobs only
gh run rerun <run-id> --failed

# Re-run entire workflow
gh run rerun <run-id>
```

---

## 🔄 Workflow Dependencies

```
┌─────────────────────────────────────────────┐
│         consolidated-ci.yml                  │
│                                             │
│  setup → [build, test, lint, security]     │
│            ↓                                │
│          docker                             │
│            ↓                                │
│          status                             │
└─────────────────────────────────────────────┘
              ↓ (on success)
┌─────────────────────────────────────────────┐
│       docker-publish.yml                     │
│                                             │
│  build-and-push → sign-image → notify      │
└─────────────────────────────────────────────┘
              ↓ (manual trigger)
┌─────────────────────────────────────────────┐
│    kubernetes-deploy.yml / helm-deploy.yml  │
│                                             │
│  validate → deploy → verify → notify        │
└─────────────────────────────────────────────┘
```

---

## 🎓 Best Practices

1. **Always use dry-run first** when deploying to production
2. **Monitor workflow execution** during deployments
3. **Test rollback procedures** regularly in dev/staging
4. **Keep secrets updated** in GitHub repository settings
5. **Review workflow logs** for optimization opportunities
6. **Use consolidated-ci.yml** for all CI needs
7. **Tag releases properly** for versioning

---

## 📈 Metrics & Monitoring

### GitHub Actions Insights
Navigate to: `Repository → Actions → Insights`

**Key Metrics:**
- Workflow success rate
- Average execution time
- Most used workflows
- Cache performance

### Deployment Metrics

**Track:**
- Deployment frequency
- Lead time for changes
- Mean time to recovery (MTTR)
- Change failure rate

---

## 🚀 Future Enhancements

### Planned
- GitOps integration (ArgoCD/FluxCD)
- Progressive delivery (Flagger)
- Multi-cluster deployments
- Blue-green deployments
- Canary releases

### Under Consideration
- Self-hosted runners
- Workflow templates
- Custom GitHub App
- Advanced caching strategies

---

**For detailed information, see:**
- `/docs/CI-CD-VALIDATION-REPORT.md`
- `/docs/DEPLOYMENT-QUICK-START.md`
