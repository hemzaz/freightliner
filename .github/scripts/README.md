# GitHub Actions Scripts

This directory contains utility scripts for managing and monitoring GitHub Actions workflows.

## Available Scripts

### cicd-status.sh

**Purpose:** Real-time status dashboard for CICD pipeline health

**Usage:**
```bash
# One-time status check
./cicd-status.sh

# Watch mode (auto-refresh every 30 seconds)
./cicd-status.sh --watch
```

**What It Shows:**
- 🎯 Recent workflow runs (last 10)
- 📊 24-hour statistics (total, success, failure, cancelled)
- ✅ Success rate with color-coded health indicators
- 🔍 Critical workflow status (CI, Tests, Deploy, Docker)
- ⚡ Quick health checks (timeouts, failures, deployments)
- 📋 Quick action commands

**Output Example:**
```
╔════════════════════════════════════════════════════════════════╗
║  Freightliner CICD Pipeline Status                             ║
╚════════════════════════════════════════════════════════════════╝

Repository: owner/freightliner
Last Updated: 2025-12-10 14:30:00

Recent Workflow Runs (Last 10):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 ✓ [12/10 14:25] CI Pipeline
 ✓ [12/10 14:20] Integration Tests
 ✗ [12/10 14:15] Deployment
...

Pipeline Health (Last 24 Hours):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Total Runs:      45
  ✓ Successful:    43
  ✗ Failed:        2
  ⊘ Cancelled:     0
  ⟳ In Progress:   0
  Success Rate:    95.6% ✓
```

**Use Cases:**
- Quick morning health check
- Monitoring during deployments
- Investigating issues (see which workflows are failing)
- Week 1 monitoring during optimization validation

**Requirements:**
- GitHub CLI (`gh`)
- `jq`
- Terminal with color support

---

### validate-optimizations.sh

**Purpose:** Automated validation of CICD optimization changes from Session 4

**Usage:**
```bash
# Basic validation
./validate-optimizations.sh

# Verbose output with details
./validate-optimizations.sh --verbose
```

**What It Checks:**
1. ✅ Modified workflow files exist
2. ✅ Timeout configurations are correct
3. ✅ Permission blocks are present
4. ✅ Go version consistency (1.25.4)
5. ✅ No deprecated actions
6. ✅ Recent workflow runs and success rate
7. ✅ Permission errors in logs
8. ✅ Timeout failures
9. ✅ Workflow-specific validations
10. ✅ Documentation completeness

**Exit Codes:**
- `0` - All checks passed (or passed with warnings)
- `1` - One or more checks failed

**Requirements:**
- GitHub CLI (`gh`) - Install: https://cli.github.com/
- `jq` - Install: `brew install jq` (macOS)
- Repository access with appropriate permissions

**Example Output:**
```
========================================
CICD Optimization Validation
Session 4 - Week 1 Monitoring
========================================

Repository: owner/repo
Date: 2025-12-10 14:30:00

1. Checking Modified Workflow Files
---
✓ Workflow file exists: benchmark.yml
✓ Workflow file exists: comprehensive-validation.yml
...

Overall Status: ✓ ALL CHECKS PASSED
```

**Scheduling:**

Run daily during Week 1 monitoring:

```bash
# Add to crontab for daily execution at 9 AM
0 9 * * * cd /Users/elad/PROJ/freightliner && ./.github/scripts/validate-optimizations.sh >> .github/logs/validation-$(date +\%Y\%m\%d).log 2>&1
```

Or use GitHub Actions scheduled workflow:

```yaml
name: Daily Validation Check

on:
  schedule:
    - cron: '0 9 * * *'  # 9 AM daily

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run validation
        run: ./.github/scripts/validate-optimizations.sh
```

## Troubleshooting

### GitHub CLI Not Authenticated

```bash
gh auth login
```

### Permission Denied

```bash
chmod +x ./.github/scripts/validate-optimizations.sh
```

### jq Not Found

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

### Script Returns Warnings

Review the warning messages. Common warnings:
- No recent workflow runs (workflows may need to be triggered)
- Workflow hasn't run recently (manual trigger workflows like deploy.yml)
- Success rate below target (investigate failed runs)

These are typically informational and don't require immediate action unless they persist.

### Script Returns Failures

Review the failure messages and:
1. Check workflow files for syntax errors
2. Review recent workflow run logs
3. Consult `WEEK1_MONITORING_GUIDE.md`
4. Consider selective rollback if issues persist

## Related Documentation

- `../WEEK1_MONITORING_GUIDE.md` - Comprehensive Week 1 monitoring procedures
- `../SESSION_SUMMARY.md` - Complete session overview
- `../PERMISSIONS_OPTIMIZATION_SUMMARY.md` - Permission changes details
- `../TIMEOUT_OPTIMIZATION_SUMMARY.md` - Timeout changes details
- `../WORKFLOW_VALIDATION_REPORT.md` - Validation results

## Support

For issues or questions:
1. Check the monitoring guide: `WEEK1_MONITORING_GUIDE.md`
2. Review session documentation in `.github/` directory
3. Check GitHub Actions workflow logs: `gh run list --status failure`
4. Contact the DevOps team

## Future Script Ideas

Potential future additions:
- `analyze-timeouts.sh` - Deep analysis of job execution times vs configured timeouts
- `audit-permissions.sh` - Automated quarterly permission audit with recommendations
- `generate-week1-report.sh` - Automatically generate Week 1 summary report from metrics
- `cost-tracker.sh` - Track GitHub Actions minutes usage and cost trends
- `workflow-diff.sh` - Compare workflow configurations between branches/commits
