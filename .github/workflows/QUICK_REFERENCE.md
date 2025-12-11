# GitHub Actions Quick Reference Card

## 🚀 Core Workflows

### 1. CI Pipeline
```bash
Workflow: consolidated-ci.yml
Trigger:  Push, PR, Manual
Duration: 15-20 minutes
Purpose:  Build → Test → Lint → Security → Docker
```

### 2. Security Scan
```bash
Workflow: security-scan.yml
Trigger:  PR, Push, Call, Manual
Duration: 10-30 minutes (scope dependent)
Purpose:  Secrets → SAST → Dependencies → Container → IaC
```

### 3. Deploy
```bash
Workflow: deploy-unified.yml
Trigger:  Manual, Push to main (dev only)
Duration: 15-25 minutes
Purpose:  Build → Deploy → Health Check → Rollback
```

### 4. Release
```bash
Workflow: release-pipeline.yml
Trigger:  Tag push (v*.*.*)
Duration: 20-30 minutes
Purpose:  Build Binaries → Docker → Release Notes
```

### 5. Monitoring
```bash
Workflow: monitoring.yml
Trigger:  Daily 2 AM UTC, Manual
Duration: 30-40 minutes
Purpose:  Security → Health → Dependencies → Alerts
```

## 📋 Common Commands

### Trigger Workflow Manually
```bash
# Security scan (quick)
gh workflow run security-scan.yml -f scan_scope=quick

# Security scan (full)
gh workflow run security-scan.yml -f scan_scope=full

# Deploy to staging
gh workflow run deploy-unified.yml -f environment=staging -f version=v1.0.0

# Deploy to production (dry-run)
gh workflow run deploy-unified.yml -f environment=production -f dry_run=true

# Run monitoring
gh workflow run monitoring.yml -f scan_type=security
```

### Check Workflow Status
```bash
# List recent runs
gh run list --limit 10

# View specific run
gh run view <run-id>

# Watch live
gh run watch

# View logs
gh run view <run-id> --log
```

### Cancel Workflow
```bash
# Cancel specific run
gh run cancel <run-id>

# Cancel all runs for workflow
gh run list --workflow=ci.yml --json databaseId -q '.[].databaseId' | xargs -n1 gh run cancel
```

## 🎯 Decision Tree

```
What do you want to do?

├─ Test my code changes
│  └─ Push to PR → consolidated-ci.yml (auto)
│
├─ Check for security issues
│  └─ gh workflow run security-scan.yml -f scan_scope=full
│
├─ Deploy to environment
│  ├─ Dev: push to main (auto) or manual
│  ├─ Staging: gh workflow run deploy-unified.yml -f environment=staging
│  └─ Production: gh workflow run deploy-unified.yml -f environment=production
│
├─ Create a release
│  └─ git tag v1.0.0 && git push --tags → release-pipeline.yml (auto)
│
└─ Check system health
   └─ Wait for daily monitoring or trigger manually
```

## 🔧 Troubleshooting

### Workflow Failed?
```bash
1. Check logs: gh run view <run-id> --log
2. Check job: gh run view <run-id> --job=<job-id>
3. Re-run: gh run rerun <run-id>
4. Re-run failed: gh run rerun <run-id> --failed
```

### Security Issues?
```bash
1. View alerts: gh browse /security/code-scanning
2. Check SARIF: .github/workflows/security-scan.yml logs
3. Review exceptions: .gitleaks.toml
```

### Deployment Issues?
```bash
1. Check environment: gh browse /settings/environments
2. Verify secrets: gh secret list
3. Check kubeconfig: kubectl config view
4. Rollback: gh workflow run rollback.yml -f environment=<env>
```

## 📊 Workflow Matrix

| Event | Workflow | Jobs | Time |
|-------|----------|------|------|
| PR | CI | Build, Test, Lint, Security (quick), Docker | 15-20m |
| Push main | CI + Deploy | CI + Deploy to dev | 25-30m |
| Push tag | Release | Build all platforms + Docker + Release | 20-30m |
| Manual deploy | Deploy | Build + Deploy + Health check | 15-25m |
| Daily 2AM | Monitoring | Security + Health + Dependencies | 30-40m |

## 🎨 Environment Variables

### CI Pipeline
```yaml
GO_VERSION: '1.25.4'
GOLANGCI_LINT_VERSION: 'v1.62.2'
```

### Security Scan
```yaml
SEVERITY_THRESHOLD: 'HIGH'  # or 'CRITICAL', 'MEDIUM', 'LOW'
SCAN_SCOPE: 'quick'         # or 'full'
```

### Deploy
```yaml
REGISTRY: 'ghcr.io'
IMAGE_NAME: '${{ github.repository }}'
```

## 🔐 Required Secrets

```bash
# Repository secrets
GITHUB_TOKEN          # (automatic)
CODECOV_TOKEN         # (optional)
SEMGREP_APP_TOKEN     # (optional)

# Environment secrets
KUBE_CONFIG_DEV       # Kubernetes config for dev
KUBE_CONFIG_STAGING   # Kubernetes config for staging
KUBE_CONFIG_PROD      # Kubernetes config for production

# Optional
SLACK_WEBHOOK_URL     # Slack notifications
```

## 🏃 Quick Start

### New Feature
```bash
1. Create feature branch: git checkout -b feature/my-feature
2. Make changes and commit
3. Push: git push origin feature/my-feature
4. Create PR → CI runs automatically
5. Fix any issues
6. Merge → Auto-deploy to dev
```

### Security Check
```bash
1. Run: gh workflow run security-scan.yml -f scan_scope=full
2. Wait for completion (~30 min)
3. Check results in Security tab
4. Fix any issues
5. Re-run to verify
```

### Deploy to Production
```bash
1. Ensure all tests pass on main
2. Create tag: git tag v1.0.0
3. Push tag: git push --tags
4. Wait for release build (~25 min)
5. Deploy: gh workflow run deploy-unified.yml -f environment=production -f version=v1.0.0
6. Approve deployment in GitHub UI
7. Monitor health checks
```

## 📈 Performance Tips

### Speed up CI
- ✅ Use cache effectively (automatic)
- ✅ Run jobs in parallel (automatic)
- ✅ Skip unchanged paths (automatic)
- ✅ Use matrix strategy (automatic)

### Speed up Security Scans
- ✅ Use quick scope for PRs (default)
- ✅ Use full scope for releases only
- ✅ Run in parallel where possible (automatic)

### Speed up Deployments
- ✅ Build once, deploy many times
- ✅ Use pre-built images
- ✅ Enable health checks
- ✅ Use rollback on failure (automatic)

## 🎓 Learning Resources

### Workflow Files
- `consolidated-ci.yml` - Main CI pipeline
- `security-scan.yml` - Security scanning
- `deploy-unified.yml` - Deployment
- `release-pipeline.yml` - Release process
- `monitoring.yml` - Scheduled monitoring

### Documentation
- `README.md` - Comprehensive guide
- `OPTIMIZATION_PLAN.md` - Strategy and planning
- `IMPLEMENTATION_SUMMARY.md` - What was built

### GitHub Actions Docs
- Workflow syntax: https://docs.github.com/en/actions/using-workflows
- Reusable workflows: https://docs.github.com/en/actions/using-workflows/reusing-workflows
- Security: https://docs.github.com/en/actions/security-guides

## 💡 Pro Tips

1. **Use dry-run** for production deploys first
2. **Monitor first runs** after workflow changes
3. **Check cache hit rates** to optimize
4. **Review security alerts** daily
5. **Keep workflows DRY** (Don't Repeat Yourself)
6. **Test in feature branches** before main
7. **Use workflow_call** for reusability
8. **Enable auto-merge** for passing PRs
9. **Set up branch protection** properly
10. **Document custom behavior**

---

**Version**: 1.0
**Last Updated**: 2025-12-11
**Print this page for quick reference**
