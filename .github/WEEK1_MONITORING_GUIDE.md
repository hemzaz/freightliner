# Week 1 Monitoring & Validation Guide

**Purpose:** Validate CICD optimizations implemented in Session 4
**Duration:** 7 days from deployment
**Status:** 🟡 In Progress
**Last Updated:** 2025-12-10

---

## Quick Reference

### Changes to Monitor
- ✅ **3 workflows** with reduced timeouts (15 min saved per execution)
- ✅ **9 workflows** with new explicit permissions
- ✅ **1 deprecated action** replaced (deploy.yml)
- ✅ **1 Go version** standardized (integration-tests.yml)

### Success Criteria
- Zero permission-related errors
- Zero timeout failures
- All PR comments post successfully
- All security scans upload successfully
- All Docker publishes complete successfully

---

## Daily Monitoring Tasks

### Day 1: Initial Validation

#### 1. Check Workflow Runs
```bash
# Navigate to repository
cd /Users/elad/PROJ/freightliner

# View recent workflow runs (requires gh CLI)
gh run list --limit 20

# Check for failed runs
gh run list --status failure --limit 10

# View specific run details
gh run view <run-id>
```

#### 2. Validate Permission Changes
**Workflows to Watch:**
- `benchmark.yml` - Should post PR comments
- `comprehensive-validation.yml` - Should upload SARIF and post comments
- `integration-tests.yml` - Should post test results
- `reusable-docker-publish.yml` - Should publish containers
- `reusable-security-scan.yml` - Should upload security results

**What to Check:**
```bash
# Check for permission errors in logs
gh run view <run-id> --log | grep -i "permission\|forbidden\|403"

# Verify PR comments are being posted
gh pr view <pr-number> --comments

# Check GitHub Security tab for uploaded scans
open https://github.com/<org>/<repo>/security/code-scanning
```

#### 3. Validate Timeout Changes
**Workflows with Reduced Timeouts:**
- `benchmark.yml` copy-benchmarks: 35 min (was 40)
- `integration-tests.yml` integration-tests: 25 min (was 30)
- `integration-tests.yml` performance-tests: 25 min (was 30)

**What to Check:**
```bash
# Check if any jobs are timing out
gh run list --status timed_out --limit 10

# View job execution times
gh run view <run-id> --log | grep "Job .* completed"

# Compare with previous execution times
gh api /repos/:owner/:repo/actions/runs/<run-id>/timing --jq '.run_started_at, .run_completed_at'
```

---

### Days 2-7: Ongoing Monitoring

#### Daily Checklist
- [ ] Review failed workflow runs: `gh run list --status failure`
- [ ] Check for permission errors in logs
- [ ] Verify PR comments are posting correctly
- [ ] Monitor job execution times vs timeouts
- [ ] Check security scan uploads to Security tab
- [ ] Review GitHub Actions insights dashboard

#### GitHub Actions Insights
Navigate to: `https://github.com/<org>/<repo>/actions`

**Metrics to Track:**
1. **Workflow Success Rate**
   - Target: >95%
   - Check: Workflows → Select workflow → View metrics

2. **Average Execution Time**
   - Compare with pre-optimization baseline
   - Should see 15+ min improvement per full execution

3. **Failure Reasons**
   - Look for patterns in failures
   - Check if timeout-related or permission-related

4. **Cost Metrics**
   - Settings → Actions → Billing
   - Track minutes used vs previous period

---

## Specific Validation Procedures

### 1. Benchmark Workflow (benchmark.yml)

**Expected Behavior:**
- Runs on PR events
- Completes within 35 minutes
- Posts benchmark results as PR comment

**Validation Steps:**
```bash
# Trigger manually or wait for PR event
# Check run status
gh run list --workflow=benchmark.yml --limit 5

# Verify PR comment was posted
gh pr view <pr-number> --json comments --jq '.comments[] | select(.body | contains("Benchmark Results"))'

# Check execution time
gh run view <run-id> --json jobs --jq '.jobs[] | select(.name == "copy-benchmarks") | .completed_at - .started_at'
```

**Red Flags:**
- Job times out at 35 minutes
- No PR comment posted (permission issue)
- "Permission denied" errors in logs

---

### 2. Integration Tests (integration-tests.yml)

**Expected Behavior:**
- Runs on push/PR to main/master/develop
- Integration tests complete within 25 minutes
- Performance tests complete within 25 minutes
- Posts test results as PR comment

**Validation Steps:**
```bash
# Check recent runs
gh run list --workflow="Integration Tests" --limit 5

# Verify test completion times
gh run view <run-id> --json jobs --jq '.jobs[] | {name: .name, duration: (.completed_at - .started_at), status: .conclusion}'

# Check for test result comments on PRs
gh pr view <pr-number> --json comments --jq '.comments[] | select(.body | contains("Performance Test Results"))'
```

**Red Flags:**
- Jobs timeout at 25 minutes (need to increase timeout)
- Tests fail due to insufficient time
- No PR comments (permission issue)

---

### 3. Comprehensive Validation (comprehensive-validation.yml)

**Expected Behavior:**
- Runs validation tests
- Uploads SARIF results to Security tab
- Posts validation results as PR comment

**Validation Steps:**
```bash
# Check recent runs
gh run list --workflow="comprehensive-validation.yml" --limit 5

# Verify SARIF upload succeeded
gh api /repos/:owner/:repo/code-scanning/alerts --jq 'map(select(.created_at > "2025-12-10")) | length'

# Check PR comments
gh pr view <pr-number> --json comments --jq '.comments[] | select(.body | contains("validation"))'
```

**Red Flags:**
- SARIF upload fails (security-events permission issue)
- No validation results in Security tab
- Missing PR comments

---

### 4. Docker Publishing (reusable-docker-publish.yml)

**Expected Behavior:**
- Builds and publishes Docker images
- Signs images with Cosign (OIDC)
- Uploads vulnerability scans
- Pushes to GitHub Container Registry

**Validation Steps:**
```bash
# Check recent Docker publish runs
gh run list --workflow="reusable-docker-publish.yml" --limit 5

# Verify packages are published
gh api /users/<owner>/packages/container/<package>/versions --jq '.[0] | {name: .name, created_at: .created_at}'

# Check for signing attestations
gh api /repos/:owner/:repo/attestations/<package> --jq '.attestations[] | {predicate_type: .predicate_type}'

# Verify vulnerability scans uploaded
# Check Security tab → Code scanning
```

**Red Flags:**
- Package publish fails (packages: write permission issue)
- Image signing fails (id-token: write permission issue)
- SARIF upload fails (security-events: write permission issue)
- 403 Forbidden errors in logs

---

### 5. Deployment Workflow (deploy.yml)

**Expected Behavior:**
- Creates GitHub releases using `gh release create`
- No longer uses deprecated `actions/create-release@v1`
- Release creation succeeds

**Validation Steps:**
```bash
# Check recent deployment runs
gh run list --workflow=deploy.yml --limit 5

# Verify release was created
gh release list --limit 5

# Check release creation step in logs
gh run view <run-id> --log | grep -A 10 "Create GitHub Release"
```

**Red Flags:**
- Release creation fails
- "gh: command not found" errors
- GitHub CLI authentication issues

---

### 6. Reusable Workflows

**Workflows to Monitor:**
- `reusable-build.yml` - Binary builds
- `reusable-test.yml` - Test execution
- `reusable-security-scan.yml` - Security scanning

**Expected Behavior:**
- All complete with read-only permissions
- No write operations attempted
- Security scans upload successfully

**Validation Steps:**
```bash
# Check for permission errors in any reusable workflow
gh run list --limit 20 --json name,conclusion,workflowName | jq '.[] | select(.workflowName | contains("reusable"))'

# View logs for specific reusable workflow run
gh run view <run-id> --log | grep -i "permission\|forbidden"
```

---

## Common Issues and Solutions

### Issue 1: Permission Denied Errors

**Symptoms:**
```
Error: Resource not accessible by integration
Error: 403 Forbidden
```

**Diagnosis:**
```bash
# Check workflow permissions in YAML
grep -A 5 "^permissions:" .github/workflows/<workflow>.yml

# View run attempt with permission details
gh run view <run-id> --log | grep -B 5 -A 5 "permission"
```

**Solutions:**
1. **Missing permission** - Add required permission to workflow
2. **Repository settings** - Check repository settings → Actions → General → Workflow permissions
3. **Branch protection** - Verify token has access to protected branches

**How to Fix:**
```yaml
# Add missing permission to workflow file
permissions:
  contents: read
  missing-permission: write  # Add as needed
```

---

### Issue 2: Timeout Failures

**Symptoms:**
```
Error: The operation was canceled.
Job timed out after 25 minutes
```

**Diagnosis:**
```bash
# Check job execution times
gh run view <run-id> --json jobs --jq '.jobs[] | {name: .name, duration: ((.completed_at // now) - .started_at), status: .conclusion}'

# Compare with timeout setting
grep "timeout-minutes:" .github/workflows/<workflow>.yml
```

**Solutions:**
1. **Insufficient timeout** - Increase timeout by 5-10 minutes
2. **Actual performance issue** - Investigate slow steps
3. **Resource contention** - Check if multiple jobs running

**How to Fix:**
```yaml
# Increase timeout conservatively
jobs:
  job-name:
    timeout-minutes: 30  # Increase from 25
```

---

### Issue 3: PR Comments Not Posting

**Symptoms:**
- Workflow succeeds but no PR comment
- "Could not create comment" errors

**Diagnosis:**
```bash
# Check if pull-requests: write permission exists
grep -A 5 "^permissions:" .github/workflows/<workflow>.yml | grep "pull-requests"

# Check PR comment step logs
gh run view <run-id> --log | grep -A 10 "Comment PR"
```

**Solutions:**
1. **Missing permission** - Add `pull-requests: write`
2. **Wrong context** - Verify using `github.event.pull_request.number`
3. **Fork PRs** - External fork PRs have limited permissions

**How to Fix:**
```yaml
# Ensure permission exists
permissions:
  contents: read
  pull-requests: write  # Required for PR comments
```

---

### Issue 4: Security Scan Upload Failures

**Symptoms:**
```
Error: Upload SARIF failed
403 Forbidden - security-events
```

**Diagnosis:**
```bash
# Check for security-events permission
grep -A 5 "^permissions:" .github/workflows/<workflow>.yml | grep "security-events"

# View upload step logs
gh run view <run-id> --log | grep -A 10 "Upload.*SARIF"
```

**Solutions:**
1. **Missing permission** - Add `security-events: write`
2. **Invalid SARIF** - Validate SARIF format
3. **Repository settings** - Enable code scanning in settings

**How to Fix:**
```yaml
# Add security-events permission
permissions:
  contents: read
  security-events: write  # Required for SARIF upload
```

---

## Rollback Procedures

### Quick Rollback (All Changes)

If major issues are discovered:

```bash
cd .github/workflows

# Revert all modified files to previous commit
git checkout HEAD~1 benchmark.yml
git checkout HEAD~1 comprehensive-validation.yml
git checkout HEAD~1 integration-tests.yml
git checkout HEAD~1 test-matrix.yml
git checkout HEAD~1 scheduled-comprehensive.yml
git checkout HEAD~1 reusable-build.yml
git checkout HEAD~1 reusable-test.yml
git checkout HEAD~1 reusable-security-scan.yml
git checkout HEAD~1 reusable-docker-publish.yml
git checkout HEAD~1 deploy.yml

# Commit rollback
git commit -m "Rollback: CICD optimizations - identified issues during Week 1 monitoring"
git push
```

### Selective Rollback (By Category)

**Revert Only Timeout Changes:**
```bash
git show HEAD~1:.github/workflows/benchmark.yml | \
  sed -n '/timeout-minutes:/p' > /tmp/timeout.txt
# Manually restore timeout values
```

**Revert Only Permission Changes:**
```bash
git show HEAD~1:.github/workflows/benchmark.yml | \
  sed -n '/^permissions:/,/^[^ ]/p'
# Manually restore or remove permissions block
```

**Revert Only deploy.yml Changes:**
```bash
git checkout HEAD~1 .github/workflows/deploy.yml
git commit -m "Revert deploy.yml to previous release creation method"
git push
```

---

## Metrics Tracking

### Week 1 Metrics Spreadsheet

Track these daily:

| Metric | Day 1 | Day 2 | Day 3 | Day 4 | Day 5 | Day 6 | Day 7 | Target |
|--------|-------|-------|-------|-------|-------|-------|-------|--------|
| Total Runs | | | | | | | | - |
| Success Rate | | | | | | | | >95% |
| Failed Runs | | | | | | | | <5 |
| Permission Errors | | | | | | | | 0 |
| Timeout Failures | | | | | | | | 0 |
| PR Comments Posted | | | | | | | | 100% |
| SARIF Uploads | | | | | | | | 100% |
| Docker Publishes | | | | | | | | 100% |
| Avg Execution Time | | | | | | | | <baseline |

### Automated Metrics Collection

```bash
#!/bin/bash
# Save as monitor-week1.sh

echo "Date: $(date)"
echo "---"

# Total runs today
echo "Total runs (last 24h):"
gh run list --created "$(date -v-1d +%Y-%m-%d)" --limit 100 --json conclusion | jq 'length'

# Success rate
echo "Success rate:"
gh run list --status success --created "$(date -v-1d +%Y-%m-%d)" --limit 100 | wc -l

# Failed runs
echo "Failed runs:"
gh run list --status failure --created "$(date -v-1d +%Y-%m-%d)" --limit 100

# Check for permission errors
echo "Permission errors:"
gh run list --created "$(date -v-1d +%Y-%m-%d)" --limit 100 --json databaseId | \
  jq -r '.[].databaseId' | \
  while read id; do
    gh run view $id --log 2>/dev/null | grep -i "permission\|403" && echo "  Run ID: $id"
  done

# Check for timeout failures
echo "Timeout failures:"
gh run list --status timed_out --created "$(date -v-1d +%Y-%m-%d)" --limit 100
```

Run daily:
```bash
chmod +x monitor-week1.sh
./monitor-week1.sh >> week1-metrics.log
```

---

## End of Week 1 Review

### Review Checklist

- [ ] All workflows executed at least once
- [ ] Zero permission-related errors observed
- [ ] Zero timeout failures observed
- [ ] All PR comments posted successfully
- [ ] All security scans uploaded successfully
- [ ] All Docker publishes completed successfully
- [ ] Metrics show expected improvements
- [ ] No rollbacks were required
- [ ] Team feedback collected

### Success Criteria Met?

**If YES:**
- ✅ Mark optimizations as validated
- ✅ Update SESSION_SUMMARY.md status to "Deployed & Validated"
- ✅ Proceed to Short Term recommendations (Weeks 2-4)
- ✅ Share success metrics with stakeholders

**If NO:**
- ⚠️ Document specific issues encountered
- ⚠️ Apply targeted fixes or rollbacks
- ⚠️ Extend monitoring period to Week 2
- ⚠️ Update optimization documentation with lessons learned

### Week 1 Report Template

```markdown
# Week 1 CICD Optimization Validation Report

**Date Range:** YYYY-MM-DD to YYYY-MM-DD
**Status:** ✅ Validated / ⚠️ Issues Found / ❌ Rollback Required

## Summary
[Brief overview of Week 1 monitoring results]

## Metrics
- Total workflow runs: XXX
- Success rate: XX%
- Permission errors: X
- Timeout failures: X
- Average execution time improvement: XX minutes

## Issues Encountered
1. [Issue description and resolution]
2. [Issue description and resolution]

## Recommendations
- [Action item 1]
- [Action item 2]

## Status
[Ready for next phase / Requires additional monitoring / Requires rollback]
```

---

## Support and Escalation

### When to Escalate

**Immediate Escalation (Critical):**
- Production deployments failing
- Multiple workflows consistently failing
- Security scan uploads failing (compliance risk)
- Docker publishes failing (deployment blocker)

**Standard Escalation (Non-Critical):**
- PR comments not posting (convenience feature)
- Intermittent timeout issues (may need adjustment)
- Single workflow having issues (isolated problem)

### Debug Commands

```bash
# Get detailed run information
gh run view <run-id> --log-failed

# Download full logs
gh run download <run-id>

# View workflow file syntax
gh workflow view <workflow-name>

# Check repository permissions
gh api /repos/:owner/:repo --jq '.permissions'

# View Actions settings
gh api /repos/:owner/:repo/actions/permissions
```

### Useful Resources

- [GitHub Actions Permissions Reference](https://docs.github.com/en/actions/security-guides/automatic-token-authentication)
- [Troubleshooting GitHub Actions](https://docs.github.com/en/actions/monitoring-and-troubleshooting-workflows)
- Session Documentation: `SESSION_SUMMARY.md`, `PERMISSIONS_OPTIMIZATION_SUMMARY.md`, `WORKFLOW_VALIDATION_REPORT.md`

---

## Appendix: Quick Command Reference

```bash
# View all recent workflow runs
gh run list --limit 20

# Watch live workflow
gh run watch <run-id>

# View specific workflow runs
gh run list --workflow=<workflow-name>

# Check for failures in last 24 hours
gh run list --status failure --created "$(date -v-1d +%Y-%m-%d)"

# View logs for failed run
gh run view <run-id> --log-failed

# Re-run failed jobs
gh run rerun <run-id> --failed

# Cancel running workflow
gh run cancel <run-id>

# View workflow file
cat .github/workflows/<workflow>.yml

# Check git status
git status

# View recent commits
git log --oneline -5

# Revert specific file
git checkout HEAD~1 <file>
```

---

**Status:** 🟡 Monitoring in Progress
**Next Review:** After 7 days
**Contact:** [Team lead or responsible party]
