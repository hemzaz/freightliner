# CI/CD Validation Checklist

Quick reference checklist for validating workflow changes and ensuring CI/CD infrastructure health.

## Pre-Deployment Checklist

### 1. Workflow Syntax
- [ ] All YAML files validate with `yamllint` or `python -c "import yaml; yaml.safe_load(open('file.yml'))"`
- [ ] No syntax errors in workflow files
- [ ] All action versions are pinned (use @v4, @v5, not @main)
- [ ] No trailing whitespace or formatting issues

### 2. Security Validation
- [ ] No hardcoded secrets in workflow files
- [ ] All secrets use `${{ secrets.SECRET_NAME }}` syntax
- [ ] Minimal permissions defined (no unnecessary write access)
- [ ] SARIF results uploaded for security scans
- [ ] Secret scanning enabled (TruffleHog, GitLeaks)
- [ ] SAST scanning configured (Gosec, Semgrep)
- [ ] Container scanning enabled (Trivy, Grype)
- [ ] IaC scanning active (Checkov, TFSec)

### 3. Performance Configuration
- [ ] Timeout set on all jobs (timeout-minutes)
- [ ] Concurrency control configured
- [ ] Caching strategy implemented
- [ ] Parallel execution where possible
- [ ] Appropriate job dependencies

### 4. Error Handling
- [ ] Critical paths fail fast (no continue-on-error)
- [ ] Optional features use continue-on-error
- [ ] Rollback mechanisms in place
- [ ] Health checks after deployments
- [ ] Proper error messages in logs

### 5. Testing & Coverage
- [ ] Test coverage threshold configured
- [ ] Race detection enabled for Go tests
- [ ] Integration tests have service dependencies
- [ ] Benchmark tests available
- [ ] Matrix testing across platforms

### 6. Deployment Safety
- [ ] Environment protection configured
- [ ] Manual approval gates for production
- [ ] Dry-run option available
- [ ] Rollback workflow exists
- [ ] Health check validation
- [ ] Smoke tests after deployment

### 7. Monitoring & Observability
- [ ] GitHub Step Summaries generated
- [ ] PR comments for status updates
- [ ] Artifact retention configured
- [ ] Metrics collection (optional)
- [ ] Notification systems (Slack, etc.)

### 8. Documentation
- [ ] Workflow purpose documented
- [ ] Required secrets documented
- [ ] Trigger conditions clear
- [ ] Dependencies listed
- [ ] Troubleshooting guide available

## Critical Security Checks

### Zero-Tolerance Security Gates
- [ ] Critical vulnerabilities = IMMEDIATE FAILURE
- [ ] High-severity secrets = IMMEDIATE FAILURE
- [ ] License violations = IMMEDIATE FAILURE
- [ ] Container vulnerabilities = IMMEDIATE FAILURE
- [ ] Infrastructure misconfigurations = IMMEDIATE FAILURE

### Permission Audit
```bash
# Check for unnecessary write permissions
grep -r "permissions:" .github/workflows/ -A10 | grep "write"

# Verify minimal permissions
grep -A5 "^permissions:" .github/workflows/*.yml
```

### Secret Safety
```bash
# Check for hardcoded secrets (should return nothing)
grep -rE "(password|token|key).*[:=].*['\"][a-zA-Z0-9]{20,}['\"]" .github/workflows/

# Verify secret references
grep -r "secrets\." .github/workflows/ | wc -l
```

## Workflow-Specific Validation

### Consolidated CI (`consolidated-ci.yml`)
- [ ] Setup job runs first
- [ ] Build, test, lint run in parallel
- [ ] Integration tests depend on build
- [ ] Docker build depends on tests
- [ ] Status job runs last with `if: always()`
- [ ] Coverage threshold set appropriately

### Security Gates (`security-gates-enhanced.yml`)
- [ ] All 5 security layers active:
  - [ ] Secret scanning
  - [ ] SAST scanning
  - [ ] Dependency scanning
  - [ ] Container scanning
  - [ ] IaC scanning
- [ ] Compliance check runs last
- [ ] Zero-tolerance policy enforced
- [ ] Timeout sufficient (20 minutes recommended)

### Deploy (`deploy.yml`)
- [ ] Environments properly configured (dev/staging/production)
- [ ] Manual approval required for staging/production
- [ ] Health checks after deployment
- [ ] Rollback job triggers on failure
- [ ] Dry-run option available

### Release Pipeline (`release-pipeline.yml`)
- [ ] Multi-platform builds (linux, darwin, windows)
- [ ] Multi-architecture (amd64, arm64)
- [ ] Checksums generated
- [ ] SBOM created
- [ ] Release notes auto-generated
- [ ] Docker images multi-platform

## Quick Validation Commands

### Validate YAML Syntax
```bash
python3 -c "
import yaml
import glob
for f in glob.glob('.github/workflows/*.yml'):
    try:
        yaml.safe_load(open(f))
        print(f'✅ {f}')
    except Exception as e:
        print(f'❌ {f}: {e}')
"
```

### Check Action Versions
```bash
# Find unpinned actions
grep -r "uses:" .github/workflows/ | grep -v "@" | grep -v "#"
```

### Verify Timeouts
```bash
# Find workflows without timeouts
grep -L "timeout-minutes:" .github/workflows/*.yml
```

### Check Concurrency
```bash
# Find workflows without concurrency control
grep -L "concurrency:" .github/workflows/*.yml | grep -v reusable
```

### Audit Permissions
```bash
# Check for dangerous permissions
grep -A10 "permissions:" .github/workflows/*.yml | grep -E "contents:.*write|packages:.*write"
```

## Issue Priority Matrix

### Critical (Fix Immediately)
- Hardcoded secrets in workflows
- Missing security scans on production paths
- No rollback mechanism
- Write permissions in CI workflows
- No timeout on jobs

### High (Fix Within 1 Week)
- Coverage threshold below 60%
- Missing manual approval for production
- No health checks after deployment
- Secret pattern false positives blocking builds

### Medium (Fix Within 1 Month)
- Timeout too tight for security scans
- Missing retry logic for external dependencies
- Incomplete path filtering
- Workflow duplication

### Low (Improve Over Time)
- Single Go version in test matrix
- No centralized version management
- Missing workflow metrics
- Documentation gaps

## Common Issues & Solutions

### Issue: Workflow Fails Due to Timeout
**Check:**
```bash
grep "timeout-minutes:" .github/workflows/your-workflow.yml
```
**Fix:** Increase timeout or optimize job
```yaml
timeout-minutes: 20  # Increase from 15
```

### Issue: Security Scan Blocking Legitimate Code
**Check:** Review scan output for false positives
**Fix:** Add exclusions for test data
```yaml
--exclude-dir=test \
--exclude-dir=testdata \
--exclude="*_test.go"
```

### Issue: Excessive CI Runs for Docs Changes
**Check:** Path filtering configuration
**Fix:** Add comprehensive exclusions
```yaml
paths-ignore:
  - '**.md'
  - 'docs/**'
  - '.github/**/*.md'
```

### Issue: Secrets Not Available
**Check:** Secret is configured in repository settings
**Fix:** Add to GitHub Secrets or use continue-on-error for optional secrets

### Issue: Cache Not Working
**Check:** Cache key configuration
**Fix:** Ensure cache key includes version and checksums
```yaml
key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

## Regular Maintenance Tasks

### Weekly
- [ ] Review failed workflow runs
- [ ] Check for security alerts
- [ ] Monitor workflow execution times
- [ ] Review artifact storage usage

### Monthly
- [ ] Update action versions
- [ ] Review and update timeouts
- [ ] Audit secret usage
- [ ] Check for deprecated features
- [ ] Review coverage trends

### Quarterly
- [ ] Comprehensive security audit
- [ ] Performance optimization review
- [ ] Documentation update
- [ ] Workflow consolidation review
- [ ] Dependency updates (actions, tools)

## Validation Report

When all checks pass, your CI/CD infrastructure meets:
- ✅ GitHub Actions best practices
- ✅ OWASP CI/CD Security Top 10
- ✅ CIS Docker Benchmark
- ✅ NIST Cybersecurity Framework
- ✅ SOC2/ISO27001 compliance requirements

## Getting Help

### Resources
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Security Best Practices](https://docs.github.com/en/actions/security-guides)
- [Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)

### Internal Documentation
- `.github/CICD_VALIDATION_REPORT.md` - Comprehensive validation report
- `.github/WORKFLOWS.md` - Workflow overview
- `.github/SECURITY_WORKFLOWS_GUIDE.md` - Security configuration

### Validation Script
```bash
# Run automated validation
bash .github/scripts/validate-cicd.sh
```

---

**Last Updated:** 2025-12-11
**Maintained By:** CI/CD Team
**Review Frequency:** Monthly
