# GitHub Actions Version Matrix

This document tracks the standardized versions of all GitHub Actions and tools used across our CI/CD pipelines.

## 📅 Last Updated: 2025-08-03 12:00:00 UTC

## 🚀 Core GitHub Actions

| Action | Current Version | Purpose | Notes |
|--------|----------------|---------|-------|
| `actions/checkout` | **v4** | Repository checkout | Latest stable, supports Node.js 20 |
| `actions/setup-go` | **v5** | Go environment setup | Latest with Node.js 20 runtime |
| `actions/cache` | **v4** | Dependency caching | Required for artifact compatibility |
| `actions/upload-artifact` | **v4** | Artifact uploads | v3 deprecated Jan 30, 2025 |
| `actions/download-artifact` | **v4** | Artifact downloads | Matches upload-artifact version |

## 🐳 Docker Actions

| Action | Current Version | Purpose | Notes |
|--------|----------------|---------|-------|
| `docker/setup-buildx-action` | **v3** | Docker Buildx setup | Multi-platform builds |
| `docker/setup-docker-action` | **v3** | Docker CE setup | Alternative to buildx |
| `docker/login-action` | **v3** | Registry authentication | Supports OIDC |
| `docker/build-push-action` | **v6** | Build and push images | Job summaries, improved performance |
| `docker/metadata-action` | **v5** | Extract image metadata | Tag and label generation |

## 🔒 Security Actions

| Action | Current Version | Purpose | Notes |
|--------|----------------|---------|-------|
| `github/codeql-action/upload-sarif` | **v3** | SARIF results upload | v2 deprecated |
| `aquasecurity/trivy-action` | **0.30.0** | Container vulnerability scanning | Pinned stable version |
| `anchore/scan-action` | **v4** | Alternative vulnerability scanner | SARIF output support |
| `returntocorp/semgrep-action` | **v1** | Static analysis | SAST scanning |
| `trufflesecurity/trufflehog` | **main** | Secret detection | Use main branch |
| `gitleaks/gitleaks-action` | **v2** | Alternative secret detection | Stable version |

## 🔧 Tool Actions

| Action | Current Version | Purpose | Notes |
|--------|----------------|---------|-------|
| `golangci/golangci-lint-action` | **v6** | Go linting | Fast with caching |
| `dominikh/staticcheck-action` | **v1.3.0** | Go static analysis | Version locked |
| `codecov/codecov-action` | **v4** | Coverage uploads | Requires token |
| `anchore/sbom-action` | **v0.17.7** | SBOM generation | Security compliance |
| `bridgecrewio/checkov-action` | **master** | IaC security scanning | Use latest |
| `tenable/terrascan-action` | **main** | Terraform scanning | Use main branch |

## 🏗️ Build Tools

| Tool | Current Version | Purpose | Notes |
|------|----------------|---------|-------|
| **Go** | `1.24.5` | Runtime version | Latest stable |
| **golangci-lint** | `latest` | Linter version | Use latest to support Go 1.24.5 |
| **gosec** | `latest` | Security scanner | `github.com/securego/gosec/v2/cmd/gosec` |
| **staticcheck** | `latest` | Static analyzer | `honnef.co/go/tools/cmd/staticcheck` |
| **Node.js** | `20` | GitHub Actions runtime | Default in latest actions |

## 🎯 Version Selection Strategy

### 🔄 Auto-Update (Use Latest)
- Security scanners (Trivy, Gosec, TruffleHog)
- Static analysis tools (Checkov, Terrascan)
- Tools with frequent security updates

### 📌 Pinned Versions
- Core language runtimes (Go 1.24.5)
- Specific tool versions for consistency (golangci-lint v2.3.0)
- Stable workflow actions (staticcheck v1.3.0)

### 🎯 Major Version Tracking
- GitHub Actions (use latest major: v4, v5, v6)
- Docker actions (track latest stable)
- Security actions (use latest for patches)

## 🔄 Update Schedule

### 🗓️ Monthly Updates
- Review and update pinned tool versions
- Check for new major versions of GitHub Actions
- Update Go version if new stable release

### 🚨 Security Updates
- Immediate updates for security-related actions
- Weekly review of vulnerability scanner versions
- Patch updates for critical issues

### 📋 Quarterly Reviews
- Comprehensive version matrix review
- Deprecation notice monitoring
- Performance impact assessment

## 📝 Change Log

### 2025-08-03
- **Initial version matrix created**
- **Updated gosec repository**: `securecodewarrior/gosec` → `securego/gosec`  
- **Standardized action versions** across all workflows:
  - Upgraded `actions/setup-go` from v4 → **v5**
  - Upgraded `actions/cache` from v3 → **v4**
  - Upgraded `actions/upload-artifact` from v3 → **v4**
  - Upgraded `docker/build-push-action` from v5 → **v6**
  - Upgraded `github/codeql-action/upload-sarif` from v2 → **v3**
  - Upgraded `golangci/golangci-lint-action` from v4 → **v6**
  - Upgraded `anchore/scan-action` from v3 → **v4**
  - Upgraded `codecov/codecov-action` from v3 → **v4**
  - Pinned `aquasecurity/trivy-action` to **0.30.0**
  - Updated `golangci-lint` version to **v1.62.2**
- **Added version validation scripts** for continuous compliance
- **Added OIDC authentication examples** for future implementation

### Future Changes
- Track deprecation notices from GitHub
- Monitor performance improvements in new versions
- Assess security updates and patches

## 🔍 Verification Commands

```bash
# Check latest versions
gh api repos/actions/checkout/releases/latest --jq '.tag_name'
gh api repos/actions/setup-go/releases/latest --jq '.tag_name'
gh api repos/docker/build-push-action/releases/latest --jq '.tag_name'

# Verify tool versions
go version
gosec --version
golangci-lint --version
```

## 🎯 Compatibility Matrix

| GitHub Runner | Ubuntu Version | Go Version | Docker Version | Node.js Version |
|---------------|----------------|------------|----------------|-----------------|
| `ubuntu-latest` | 24.04.2 LTS | 1.24.5 | 27.x | 20.x |
| `ubuntu-24.04` | 24.04.2 LTS | 1.24.5 | 27.x | 20.x |
| `ubuntu-22.04` | 22.04.5 LTS | 1.24.5 | 25.x | 20.x |

---

**Note**: This matrix should be reviewed and updated monthly, with immediate updates for security-related changes.