# Security Workflows Guide

**Purpose:** Quick reference for understanding Freightliner's security workflows
**Last Updated:** 2025-12-11
**Status:** ✅ Active

---

## Quick Reference

### Four Security Workflows

| Workflow | Purpose | When It Runs | Duration | Focus |
|----------|---------|--------------|----------|-------|
| **security-gates.yml** | Policy Enforcement | All PRs + push to main/master/develop | ~10 min | Fast policy checks |
| **security-gates-enhanced.yml** | Vulnerability Scanning | PRs/push to main/master only | ~30-40 min | Deep security scans |
| **security-comprehensive.yml** | Comprehensive Testing | Scheduled (periodic) | ~45 min | Complete analysis |
| **security-monitoring-enhanced.yml** | Continuous Monitoring | Scheduled + manual | ~20 min | Ongoing surveillance |

---

## Workflow Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     PULL REQUEST CREATED                         │
│                  (to main/master/develop)                        │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
        ┌────────────────────────────────────────┐
        │  security-gates.yml (POLICY)           │
        │  ⚡ Fast: ~10 minutes                   │
        │                                         │
        │  ✓ Required files exist?               │
        │  ✓ Workflows use hardened runner?      │
        │  ✓ Actions SHA-pinned?                 │
        │  ✓ Dockerfiles secure?                 │
        │  ✓ Branch protection enabled?          │
        └────────────────┬───────────────────────┘
                         │
                    ✅ PASS │ ❌ FAIL
                         │       └──> PR BLOCKED
                         │            (fix policy violations)
                         │
            ┌────────────┴──────────────┐
            │                            │
      To develop branch?          To main/master?
            │                            │
            ▼                            ▼
      PR can merge          ┌─────────────────────────────────────┐
      (policy check OK)     │ security-gates-enhanced.yml (SCANS) │
                            │ 🔍 Deep: ~30-40 minutes              │
                            │                                      │
                            │ 🔐 Secret scanning (TruffleHog)      │
                            │ 🔍 SAST analysis                      │
                            │ 📦 Dependency CVE scanning            │
                            │ 🐳 Container vulnerability scanning   │
                            │ ☁️ IaC security scanning             │
                            │ ✅ Compliance check                   │
                            └────────────┬─────────────────────────┘
                                         │
                                    ✅ PASS │ ❌ FAIL
                                         │       └──> PR BLOCKED
                                         │            (fix vulnerabilities)
                                         │
                                         ▼
                                  PR can merge
                              (all checks passed)


┌─────────────────────────────────────────────────────────────────┐
│                    SCHEDULED / PERIODIC                          │
└────────────────────────┬────────────────────────────────────────┘
                         │
            ┌────────────┴──────────────┐
            │                            │
            ▼                            ▼
┌──────────────────────────┐  ┌───────────────────────────────┐
│ security-comprehensive   │  │ security-monitoring-enhanced  │
│ 📅 Weekly/Monthly        │  │ 📅 Daily/Continuous           │
│ ⏱️ ~45 minutes           │  │ ⏱️ ~20 minutes                │
│                          │  │                               │
│ Complete security audit  │  │ Runtime security monitoring   │
│ All tools + analysis     │  │ Threat detection              │
│ Detailed reporting       │  │ Anomaly detection             │
└──────────────────────────┘  └───────────────────────────────┘
            │                            │
            │                            │
            └────────────┬───────────────┘
                         │
                         ▼
              📊 Security Dashboard
                 (reports generated)
```

---

## Workflow Details

### 1. security-gates.yml - Policy & Compliance Gates ⚡

**Purpose:** Fast validation that security policies and configurations are correct

**Trigger Events:**
```yaml
✅ Pull requests (any branch → main/master/develop)
✅ Push to main/master/develop
```

**What It Checks:**
- ✅ Required security files exist
  - `.github/security.yml`
  - `.gitleaks.toml`
  - `SECURITY.md`
  - `.github/dependabot.yml`
- ✅ Workflows use `step-security/harden-runner`
- ✅ Security workflows use SHA-pinned actions (not tag versions)
- ✅ Dockerfiles have non-root user configuration
- ✅ Dockerfiles have security labels
- ✅ Branch protection rules are enabled

**Execution Time:** ~10 minutes (fast)

**Fails When:**
- Required security files are missing
- Workflows don't use hardened runner
- Security workflows use non-SHA-pinned actions
- Dockerfiles run as root
- Branch protection not configured

**Why It's Fast:**
- Only checks configuration files
- No scanning or analysis
- Simple file existence and pattern matching

**Use Cases:**
- ✅ Pre-commit validation
- ✅ Quick feedback on policy compliance
- ✅ Enforce security standards across all branches
- ✅ Block PRs early if policies violated

---

### 2. security-gates-enhanced.yml - Vulnerability Scanning 🔍

**Purpose:** Comprehensive vulnerability scanning and threat detection

**Trigger Events:**
```yaml
✅ Pull requests to main/master (production branches only)
✅ Push to main/master
✅ Workflow call (reusable - can be called from other workflows)
```

**Inputs (when called as reusable):**
- `severity_threshold` - Minimum severity to fail build (default: HIGH)
- `skip_container_scan` - Skip container scanning (default: false)

**What It Scans:**

**🔐 Secret Scanning (TruffleHog)**
- Scans entire git history for leaked credentials
- Detects API keys, passwords, tokens
- Prevents credential leaks

**🔍 SAST (Static Application Security Testing)**
- Analyzes source code for security vulnerabilities
- Identifies common security flaws (injection, XSS, etc.)
- Code-level threat detection

**📦 Dependency Scanning**
- Checks Go modules for known CVEs
- Identifies outdated packages with vulnerabilities
- Dependency supply chain security

**🐳 Container Scanning**
- Scans Docker images for vulnerabilities
- Checks base image security
- Layer-by-layer vulnerability analysis

**☁️ IaC Scanning**
- Terraform/CloudFormation security checks
- Identifies misconfigurations
- Infrastructure security validation

**✅ Compliance Check**
- Aggregates all scan results
- Determines overall security posture
- Pass/fail based on severity threshold

**Execution Time:** ~30-40 minutes (comprehensive)

**Fails When:**
- Secrets found in code/history
- Critical/High vulnerabilities in code
- Vulnerable dependencies (CVEs)
- Container image has critical vulnerabilities
- IaC misconfigurations detected
- Severity threshold exceeded

**Why It's Slower:**
- Scans full git history
- Analyzes all source code
- Checks all dependencies
- Scans container images
- Multiple security tools run in parallel

**Use Cases:**
- ✅ Pre-production security validation
- ✅ Comprehensive threat detection
- ✅ Production readiness check
- ✅ Reusable from deployment workflows

**Configuration:**
```yaml
# Call from another workflow
jobs:
  security:
    uses: ./.github/workflows/security-gates-enhanced.yml
    with:
      severity_threshold: 'CRITICAL'  # Only fail on CRITICAL
      skip_container_scan: false       # Run all scans
```

---

### 3. security-comprehensive.yml - Complete Security Audit 📊

**Purpose:** Deep periodic security analysis and reporting

**Trigger Events:**
```yaml
✅ Scheduled (weekly/monthly via cron)
✅ Manual trigger (workflow_dispatch)
```

**What It Does:**
- Runs all security tools
- Comprehensive analysis
- Detailed reporting
- Historical trend analysis
- Security posture assessment

**Execution Time:** ~45 minutes

**Use Cases:**
- ✅ Weekly/monthly security audits
- ✅ Compliance reporting
- ✅ Security posture tracking
- ✅ Deep dive security analysis

---

### 4. security-monitoring-enhanced.yml - Continuous Monitoring 👁️

**Purpose:** Ongoing security monitoring and threat detection

**Trigger Events:**
```yaml
✅ Scheduled (daily via cron)
✅ Manual trigger (workflow_dispatch)
✅ On security events
```

**What It Does:**
- Runtime security monitoring
- Anomaly detection
- Threat intelligence integration
- Security dashboard updates
- Alert generation

**Execution Time:** ~20 minutes

**Use Cases:**
- ✅ Continuous security surveillance
- ✅ Early threat detection
- ✅ Real-time security posture
- ✅ Compliance monitoring

---

## Decision Tree: Which Workflow Runs When?

```
┌─ Pull Request Created ─────────────────────────────────────┐
│                                                             │
├─ To develop branch?                                         │
│  └─ YES → security-gates.yml (policy only)                 │
│      ├─ PASS → PR can merge                                │
│      └─ FAIL → PR blocked (fix policy violations)          │
│                                                             │
├─ To main/master branch?                                     │
│  └─ YES → security-gates.yml (policy)                      │
│      ├─ PASS → security-gates-enhanced.yml (scans)         │
│      │   ├─ PASS → PR can merge                            │
│      │   └─ FAIL → PR blocked (fix vulnerabilities)        │
│      └─ FAIL → PR blocked (fix policy violations)          │
│                                                             │
└─────────────────────────────────────────────────────────────┘

┌─ Push to Branch ───────────────────────────────────────────┐
│                                                             │
├─ develop branch?                                            │
│  └─ YES → security-gates.yml (policy check)                │
│                                                             │
├─ main/master branch?                                        │
│  └─ YES → security-gates.yml + security-gates-enhanced.yml │
│                                                             │
└─────────────────────────────────────────────────────────────┘

┌─ Scheduled Execution ──────────────────────────────────────┐
│                                                             │
├─ Daily: security-monitoring-enhanced.yml                    │
├─ Weekly: security-comprehensive.yml                         │
├─ Monthly: Full security audit with reporting                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Comparison Matrix

| Feature | Policy Gates | Vulnerability Scanning | Comprehensive | Monitoring |
|---------|--------------|----------------------|---------------|------------|
| **File** | security-gates.yml | security-gates-enhanced.yml | security-comprehensive.yml | security-monitoring-enhanced.yml |
| **Speed** | ⚡ Fast (~10 min) | 🐢 Slow (~30-40 min) | 🐌 Slowest (~45 min) | 🏃 Medium (~20 min) |
| **Triggers** | All PRs/pushes | main/master only | Scheduled | Scheduled |
| **Reusable** | ❌ No | ✅ Yes | ❌ No | ❌ No |
| **Secret Scan** | ❌ | ✅ TruffleHog | ✅ Multiple tools | ✅ Continuous |
| **SAST** | ❌ | ✅ Yes | ✅ Yes | ✅ Runtime |
| **Dependency** | ❌ | ✅ CVE scan | ✅ CVE scan | ✅ Monitoring |
| **Container** | ❌ | ✅ Image scan | ✅ Image scan | ✅ Runtime scan |
| **IaC** | ❌ | ✅ Terraform scan | ✅ Complete scan | ✅ Drift detection |
| **Policy Check** | ✅ Yes | ❌ | ✅ Yes | ✅ Compliance |
| **Branch Protection** | ✅ Validates | ❌ | ✅ Validates | ✅ Monitors |
| **Blocks PRs** | ✅ Yes | ✅ Yes | ❌ No | ❌ No |
| **Reporting** | ❌ Basic | ✅ Detailed | ✅ Comprehensive | ✅ Dashboards |

---

## Common Scenarios

### Scenario 1: Creating Feature PR to develop
```
1. Create PR: feature-branch → develop
2. Runs: security-gates.yml (policy check)
3. Duration: ~10 minutes
4. Result: PASS → Can merge
```

### Scenario 2: Creating Release PR to main
```
1. Create PR: release-branch → main
2. Runs: security-gates.yml (policy check) → ~10 min
3. If PASS: security-gates-enhanced.yml (vulnerability scan) → ~30-40 min
4. Result: Both PASS → Can merge to production
```

### Scenario 3: Direct Push to main (CI/CD)
```
1. Push to main branch
2. Runs: Both workflows in parallel
   - security-gates.yml (~10 min)
   - security-gates-enhanced.yml (~30-40 min)
3. Result: Deployment proceeds if both pass
```

### Scenario 4: Weekly Security Audit
```
1. Scheduled: Every Monday 2 AM UTC
2. Runs: security-comprehensive.yml (~45 min)
3. Result: Report generated, metrics updated
```

### Scenario 5: Daily Monitoring
```
1. Scheduled: Every day 6 AM UTC
2. Runs: security-monitoring-enhanced.yml (~20 min)
3. Result: Security dashboard updated
```

---

## FAQ

### Q: Why do we have two "gates" workflows?

**A:** They serve different purposes:
- **security-gates.yml** = Fast policy enforcement (10 min)
- **security-gates-enhanced.yml** = Deep vulnerability scanning (30-40 min)

Think of it like airport security:
- Policy gates = checking you have a ticket and ID (fast)
- Vulnerability scanning = full body scan and baggage X-ray (thorough)

### Q: Which workflow blocks my PR?

**A:** Depends on target branch:
- **To develop:** Only security-gates.yml (policy check)
- **To main/master:** Both workflows must pass

### Q: Can I skip vulnerability scanning?

**A:** Only for develop branch PRs (they don't run enhanced scanning).
For main/master PRs, both checks are required.

However, you can call security-gates-enhanced.yml with `skip_container_scan: true` if needed.

### Q: Why is my PR taking 40+ minutes?

**A:** You're merging to main/master, so comprehensive vulnerability scanning runs.
This is intentional for production branches to ensure security.

**Tip:** Merge to develop first for faster iteration, then create release PR to main.

### Q: Can I run security scans manually?

**A:** Yes! All workflows support manual trigger:
```bash
# Using GitHub CLI
gh workflow run security-gates.yml
gh workflow run security-gates-enhanced.yml
gh workflow run security-comprehensive.yml
gh workflow run security-monitoring-enhanced.yml
```

### Q: How do I fix policy violations?

**A:** Check the workflow run for specific violations. Common fixes:
- Add missing security files
- Update workflows to use hardened-runner
- Use SHA-pinned actions in security workflows
- Configure Dockerfiles with non-root users
- Enable branch protection

### Q: How do I fix vulnerability findings?

**A:** Depends on the finding type:
- **Secrets:** Remove from code/history, rotate credentials
- **Code vulns:** Fix the security flaw in code
- **Dependencies:** Update vulnerable packages
- **Container:** Update base image or patch vulnerabilities
- **IaC:** Fix misconfigurations in Terraform/CloudFormation

---

## Best Practices

### For Developers

**When Creating PRs:**
1. ✅ Target develop first for faster feedback
2. ✅ Fix policy violations quickly (they're fast to check)
3. ✅ For main/master PRs, expect 40+ min wait for scans
4. ✅ Don't commit secrets (secret scanning checks full history)

**When Security Checks Fail:**
1. ✅ Read the workflow logs carefully
2. ✅ Fix the specific violation/vulnerability cited
3. ✅ Push fix and let workflows re-run
4. ✅ Don't try to disable security checks

### For Maintainers

**Workflow Maintenance:**
1. ✅ Keep both workflows separate (they serve different purposes)
2. ✅ Update security tool versions regularly
3. ✅ Review security policies quarterly
4. ✅ Adjust severity thresholds as needed

**Monitoring:**
1. ✅ Review security-comprehensive.yml reports weekly
2. ✅ Monitor security-monitoring-enhanced.yml alerts daily
3. ✅ Track trends in vulnerabilities over time
4. ✅ Update policies based on findings

---

## Related Documentation

- `SECURITY_WORKFLOWS_ANALYSIS.md` - Detailed comparison and analysis
- `SESSION_SUMMARY.md` - CICD optimization overview
- `WORKFLOW_VALIDATION_REPORT.md` - Validation results
- `PERMISSIONS_OPTIMIZATION_SUMMARY.md` - Security permission details

---

## Quick Commands

```bash
# View workflow files
cat .github/workflows/security-gates.yml
cat .github/workflows/security-gates-enhanced.yml

# Run workflows manually
gh workflow run security-gates.yml
gh workflow run security-gates-enhanced.yml

# View recent runs
gh run list --workflow=security-gates.yml --limit 5
gh run list --workflow=security-gates-enhanced.yml --limit 5

# Check run status
gh run view <run-id>

# View logs
gh run view <run-id> --log
```

---

**Last Updated:** 2025-12-11
**Maintained By:** DevOps Team
**Review Schedule:** Quarterly
