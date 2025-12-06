# Freightliner Security Remediation Checklist

**Date Created:** 2025-12-06
**Priority:** CRITICAL - Complete before production deployment

---

## 🔴 CRITICAL - IMMEDIATE ACTION (Week 1)

### [ ] C1: Fix Command Injection in Credential Helpers
**File:** `pkg/auth/credential_store.go`
**Lines:** 221, 260, 282, 307, 332

**Actions:**
- [ ] Add `validateHelperName()` function with strict regex validation
- [ ] Block path traversal characters (`..`, `/`, `\`)
- [ ] Implement whitelist of allowed credential helpers
- [ ] Add security logging for helper execution attempts
- [ ] Write unit tests for validation bypass attempts

**Verification:**
```bash
# Test command injection protection
go test -v ./pkg/auth -run TestCredentialHelper_Validation
```

---

### [ ] C2: Implement Secure Memory Handling
**Files:** `pkg/auth/credential_store.go`, `cmd/login.go`

**Actions:**
- [ ] Replace string passwords with `[]byte` for secure zeroing
- [ ] Add `defer` cleanup to zero memory after use
- [ ] Force GC after credential operations
- [ ] Consider using `github.com/awnumar/memguard` for sensitive data
- [ ] Audit all locations where credentials are held in memory

**Implementation:**
```go
import (
    "runtime"
    "crypto/subtle"
)

func secureZeroMemory(b []byte) {
    for i := range b {
        b[i] = 0
    }
    runtime.GC()
}
```

---

### [ ] C3: Remove All Insecure TLS Configurations
**Files:** Multiple (see audit report)

**Actions:**
- [ ] Remove `InsecureSkipVerify: true` from production code
- [ ] Keep ONLY in test files with explicit `// For testing only` comment
- [ ] Add TLS version enforcement (minimum TLS 1.3)
- [ ] Implement certificate pinning for known registries
- [ ] Add audit log when insecure mode is attempted
- [ ] Update documentation to warn against insecure mode

**Code Review:**
```bash
# Find all instances
grep -rn "InsecureSkipVerify.*true" pkg/ cmd/

# Ensure only test files remain
git grep -n "InsecureSkipVerify.*true" | grep -v "_test.go"
```

---

### [ ] H5: Remove Hardcoded Secrets from Repository
**Files:** `docker-compose.prod.yml`, `docker-compose.yml`

**URGENT ACTIONS:**
- [ ] Change all production passwords IMMEDIATELY
- [ ] Remove from git history using `git filter-branch` or BFG Repo-Cleaner
- [ ] Rotate Grafana admin password
- [ ] Rotate MinIO root password
- [ ] Update all deployed instances
- [ ] Add `.env` to `.gitignore` if not already present
- [ ] Implement secrets management solution

**Commands:**
```bash
# Remove from git history
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch docker-compose.prod.yml' \
  --prune-empty --tag-name-filter cat -- --all

# Force push (coordinate with team first!)
git push origin --force --all
git push origin --force --tags

# Update references
echo "See https://wiki.example.com/secrets-management for current credentials" > docker-compose.prod.yml
```

---

## 🟠 HIGH PRIORITY (Weeks 2-3)

### [ ] H1: Sanitize Authentication Logging
**File:** `cmd/login.go`

**Actions:**
- [ ] Remove username from success logs
- [ ] Add `[REDACTED]` for all PII
- [ ] Implement separate audit log for security events
- [ ] Review all logger.WithFields() calls project-wide
- [ ] Add linting rule to detect logging of sensitive fields

---

### [ ] H2: Move API Keys to Secrets Management
**Files:** `pkg/server/middleware.go`, `docker-compose.yml`

**Actions:**
- [ ] Implement secrets provider interface
- [ ] Add AWS Secrets Manager integration
- [ ] Add HashiCorp Vault integration (optional)
- [ ] Update configuration loading to use secrets provider
- [ ] Document secrets rotation procedure
- [ ] Implement API key rotation mechanism

**Architecture:**
```go
type SecretsProvider interface {
    GetAPIKey(context.Context) (string, error)
    RotateAPIKey(context.Context) error
}
```

---

### [ ] H3: Comprehensive Input Validation
**File:** `pkg/server/handlers.go`

**Actions:**
- [ ] Add `validateRegistryPath()` function
- [ ] Validate all user input before processing
- [ ] Add regex patterns for registry/repository names
- [ ] Implement SSRF protection for registry endpoints
- [ ] Add request sanitization middleware
- [ ] Write property-based tests for validation bypass

---

### [ ] H4: Improve Rate Limiting
**File:** `pkg/server/middleware.go`

**Actions:**
- [ ] Validate IP addresses before use
- [ ] Implement maximum client map size (10,000 entries)
- [ ] Add LRU eviction policy
- [ ] Use atomic operations for token counting
- [ ] Add per-endpoint rate limits
- [ ] Implement distributed rate limiting (Redis) for multi-instance deployments

---

### [ ] H6: Secure Health/Metrics Endpoints
**File:** `pkg/server/middleware.go`

**Actions:**
- [ ] Require authentication for `/metrics`
- [ ] Implement IP whitelist for monitoring systems
- [ ] Separate `/health/ready` and `/health/live`
- [ ] Sanitize sensitive labels from metrics
- [ ] Add configuration option for metrics authentication

---

### [ ] H7: Generic Error Messages
**File:** `pkg/server/handlers.go`

**Actions:**
- [ ] Create error mapping table (internal → external)
- [ ] Return generic messages to users
- [ ] Log detailed errors internally with request IDs
- [ ] Add error code system (e.g., ERR-1001)
- [ ] Update API documentation with error codes

---

## 🟡 MEDIUM PRIORITY (Weeks 4-6)

### [ ] M1: Fix Timing Attack in API Key Comparison
**File:** `pkg/server/middleware.go`

**Actions:**
- [ ] Replace `==` with `subtle.ConstantTimeCompare()`
- [ ] Apply to all credential comparisons
- [ ] Add unit tests for timing consistency

---

### [ ] M2: Fix Memory Leak in Rate Limiter
**File:** `pkg/server/middleware.go`

**Actions:**
- [ ] Implement bounded client map
- [ ] Add LRU cache (e.g., `github.com/hashicorp/golang-lru`)
- [ ] Trigger cleanup under high load
- [ ] Add metrics for map size

---

### [ ] M3: Implement CSRF Protection
**File:** `pkg/server/middleware.go`

**Actions:**
- [ ] Add `gorilla/csrf` middleware
- [ ] Generate CSRF tokens per session
- [ ] Validate tokens on state-changing operations
- [ ] Add CSRF token to API responses

```go
import "github.com/gorilla/csrf"

csrfMiddleware := csrf.Protect(
    csrfKey,
    csrf.Secure(true),
    csrf.HttpOnly(true),
    csrf.SameSite(csrf.SameSiteStrictMode),
)
```

---

### [ ] M4: Fix CORS Configuration
**File:** `pkg/server/middleware.go`

**Actions:**
- [ ] Remove `Access-Control-Allow-Origin: *`
- [ ] Implement strict origin whitelist
- [ ] Add configuration for allowed origins
- [ ] Validate credentials with CORS

---

### [ ] M5: Add Security Headers
**File:** `pkg/server/middleware.go`

**Actions:**
- [ ] Create `securityHeadersMiddleware`
- [ ] Add all OWASP recommended headers
- [ ] Configure CSP policy
- [ ] Add HSTS with preload

**Headers to add:**
```go
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'
Strict-Transport-Security: max-age=63072000; includeSubDomains; preload
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

---

### [ ] M6: Comprehensive Security Audit Logging
**Files:** Multiple

**Actions:**
- [ ] Create `SecurityAuditLog` struct
- [ ] Log all authentication attempts (success/failure)
- [ ] Log authorization failures
- [ ] Log configuration changes
- [ ] Log suspicious request patterns
- [ ] Implement log aggregation (e.g., ELK stack)
- [ ] Add alerting for security events

---

### [ ] M7: Request Size Limits
**File:** `pkg/server/handlers.go`

**Actions:**
- [ ] Add `requestSizeLimitMiddleware`
- [ ] Set maximum request body size (10MB)
- [ ] Add per-endpoint size limits
- [ ] Handle size limit errors gracefully

---

### [ ] M8: Cryptographically Secure Job IDs
**File:** `pkg/server/jobs.go`

**Actions:**
- [ ] Replace sequential IDs with UUIDs
- [ ] Use `github.com/google/uuid`
- [ ] Add uniqueness validation

---

## 🔵 LOW PRIORITY (Ongoing)

### [ ] L1: Fix File Permissions
**Files:** Multiple test/validation files

**Actions:**
- [ ] Change `0644` → `0600` for sensitive files
- [ ] Audit all `os.WriteFile()` calls
- [ ] Add linting rule for file permissions

---

### [ ] L2: Sanitize Log Input
**Files:** Multiple

**Actions:**
- [ ] Create `sanitizeForLog()` function
- [ ] Remove control characters
- [ ] Apply to all user input logging
- [ ] Add to logging wrapper

---

### [ ] L3: Add Context Timeouts
**File:** `pkg/sync/batch.go`

**Actions:**
- [ ] Review all context usage
- [ ] Add timeouts to all operations
- [ ] Use `context.WithTimeout()` consistently
- [ ] Add configuration for timeout values

---

### [ ] L4: Prevent Integer Overflow
**File:** `pkg/client/common/base_transport.go`

**Actions:**
- [ ] Add explicit overflow check
- [ ] Lower maximum backoff factor cap
- [ ] Add unit tests for edge cases

---

### [ ] L5: Verify Cryptographic Algorithms
**Actions:**
- [ ] Audit all crypto usage
- [ ] Ensure SHA256+ for hashing
- [ ] Ensure AES-256-GCM for encryption
- [ ] Document cryptographic choices

---

## 🛠️ INFRASTRUCTURE & TOOLING

### [ ] CI/CD Security Integration
**Priority:** HIGH

**Actions:**
- [ ] Add GoSec to CI pipeline
- [ ] Add Trivy container scanning
- [ ] Add GitLeaks secret scanning
- [ ] Add dependency vulnerability scanning (govulncheck)
- [ ] Fail builds on critical vulnerabilities
- [ ] Add security testing stage to GitHub Actions

**Example `.github/workflows/security.yml`:**
```yaml
name: Security Scan

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run GoSec
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec -exclude=G104 -out=gosec-report.json -fmt=json ./...

      - name: Run Trivy
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          security-checks: 'vuln,config,secret'
          severity: 'CRITICAL,HIGH'

      - name: Check for secrets
        uses: trufflesecurity/trufflehog@main
        with:
          path: ./

      - name: Dependency scan
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
```

---

### [ ] Pre-commit Hooks
**Priority:** MEDIUM

**Actions:**
- [ ] Install pre-commit framework
- [ ] Add secret detection hook
- [ ] Add GoSec hook
- [ ] Add gofmt/goimports hook
- [ ] Document setup in CONTRIBUTING.md

**.pre-commit-config.yaml:**
```yaml
repos:
  - repo: https://github.com/zricethezav/gitleaks
    rev: v8.18.0
    hooks:
      - id: gitleaks

  - repo: https://github.com/securego/gosec
    rev: v2.18.0
    hooks:
      - id: gosec
```

---

### [ ] Security Documentation
**Priority:** MEDIUM

**Actions:**
- [ ] Create SECURITY.md with vulnerability reporting process
- [ ] Document secure configuration guide
- [ ] Create deployment security checklist
- [ ] Document incident response playbook
- [ ] Create security training materials

---

## 📊 TESTING & VERIFICATION

### [ ] Security Test Suite
**Priority:** HIGH

**Actions:**
- [ ] Create security test package
- [ ] Add authentication bypass tests
- [ ] Add authorization tests
- [ ] Add input validation tests
- [ ] Add CSRF tests
- [ ] Add rate limiting tests
- [ ] Add TLS configuration tests

**Example test structure:**
```
tests/
└── security/
    ├── auth_test.go
    ├── injection_test.go
    ├── csrf_test.go
    ├── rate_limit_test.go
    └── tls_test.go
```

---

### [ ] Penetration Testing
**Priority:** HIGH (before production)

**Actions:**
- [ ] Conduct internal penetration test
- [ ] Test all critical vulnerabilities fixed
- [ ] Test authentication/authorization
- [ ] Test API endpoints
- [ ] Test container security
- [ ] Document findings
- [ ] Consider third-party security audit

---

### [ ] Compliance Verification
**Priority:** MEDIUM

**Actions:**
- [ ] Document compliance requirements (PCI-DSS, GDPR, SOC 2)
- [ ] Create compliance checklist
- [ ] Conduct compliance gap analysis
- [ ] Implement missing controls
- [ ] Document audit trail

---

## 📅 TIMELINE & MILESTONES

### Week 1: Critical Fixes
- Complete C1, C2, C3
- Remove hardcoded secrets (H5)
- **Milestone:** No CRITICAL vulnerabilities remain

### Week 2-3: High Priority
- Complete H1, H2, H3, H4, H6, H7
- **Milestone:** All HIGH vulnerabilities addressed

### Week 4-6: Medium Priority
- Complete M1-M8
- Add security headers
- Implement CSRF protection
- **Milestone:** All MEDIUM vulnerabilities addressed

### Week 7-10: Low Priority & Hardening
- Complete L1-L5
- Add comprehensive security testing
- Conduct penetration testing
- **Milestone:** Production-ready security posture

---

## ✅ VERIFICATION CHECKLIST

### Before Production Deployment:
- [ ] All CRITICAL issues resolved (C1-C3)
- [ ] All HIGH issues resolved (H1-H7)
- [ ] Security CI/CD pipeline active
- [ ] Security tests passing
- [ ] Secrets removed from repository
- [ ] TLS properly configured
- [ ] Authentication/authorization tested
- [ ] Rate limiting tested
- [ ] Input validation tested
- [ ] Security documentation complete
- [ ] Incident response plan documented
- [ ] Monitoring and alerting configured

---

## 🚨 EMERGENCY CONTACTS

**Security Team Lead:** [Name]
**DevOps Lead:** [Name]
**Incident Response:** [Contact]

---

## 📝 NOTES

- This checklist is based on the security audit conducted on 2025-12-06
- All team members should have access to this checklist
- Update checklist as issues are resolved
- Conduct weekly security review meetings during remediation
- Re-audit after completing each phase

---

**Status:** IN PROGRESS
**Last Updated:** 2025-12-06
**Next Review:** Weekly until complete
