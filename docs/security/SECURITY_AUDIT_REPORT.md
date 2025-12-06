# FREIGHTLINER SECURITY AUDIT REPORT
**Date:** 2025-12-06
**Auditor:** Security Agent (Claude Code)
**Severity Classification:** 🔴 CRITICAL | 🟠 HIGH | 🟡 MEDIUM | 🔵 LOW

---

## EXECUTIVE SUMMARY

Comprehensive security audit of Freightliner container registry replication tool identified **23 security issues** across multiple categories. The codebase demonstrates good security practices in some areas (credential storage, encryption) but has critical vulnerabilities in command execution, TLS verification, and potential information leakage.

### VULNERABILITY BREAKDOWN
- 🔴 **CRITICAL:** 3 issues
- 🟠 **HIGH:** 7 issues
- 🟡 **MEDIUM:** 8 issues
- 🔵 **LOW:** 5 issues

**IMMEDIATE ACTION REQUIRED:** Address all CRITICAL and HIGH severity issues before production deployment.

---

## 🔴 CRITICAL SECURITY ISSUES

### C1: Command Injection via Credential Helpers
**File:** `/pkg/auth/credential_store.go` (Lines 221, 260, 282, 307, 332)
**OWASP:** A03:2021 - Injection

**Description:**
The credential helper execution uses user-controlled input (`helper` variable from config) to construct command names without validation:

```go
// Line 221: Vulnerable to command injection
cmdName := "docker-credential-" + helper
cmd := exec.Command(cmdName, "get")
```

**Attack Vector:**
1. Attacker modifies `~/.docker/config.json` with malicious helper: `{"credsStore": "../../../../../../tmp/malicious"}`
2. System executes `docker-credential-../../../../../../tmp/malicious get`
3. Path traversal leads to arbitrary command execution

**Proof of Concept:**
```bash
# Attacker creates malicious binary
echo '#!/bin/bash\nwhoami > /tmp/pwned' > /tmp/malicious
chmod +x /tmp/malicious

# Modify Docker config
echo '{"credsStore":"../../../../../../tmp/malicious"}' > ~/.docker/config.json

# Trigger execution
freightliner login registry.example.com
```

**Impact:**
- **RCE (Remote Code Execution)** on systems with compromised Docker config
- Credential theft from all registries
- Lateral movement within infrastructure

**Remediation:**
```go
// Validate helper name (whitelist approach)
func validateHelperName(helper string) error {
    if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(helper) {
        return fmt.Errorf("invalid credential helper name: %s", helper)
    }

    // Check for path traversal
    if strings.Contains(helper, "..") || strings.Contains(helper, "/") {
        return fmt.Errorf("credential helper name contains invalid characters")
    }

    return nil
}

// Apply validation before command execution
if err := validateHelperName(helper); err != nil {
    return "", "", err
}
```

**References:**
- CWE-78: OS Command Injection
- OWASP Command Injection Prevention Cheat Sheet

---

### C2: Plaintext Credentials in Memory
**Files:**
- `/pkg/auth/credential_store.go` (Lines 64, 103-123)
- `/cmd/login.go` (Lines 89-100, 119)

**Description:**
Credentials are stored as plaintext strings in memory without secure memory handling. Base64 encoding provides **no security** (it's encoding, not encryption).

```go
// Line 64: Plaintext concatenation before encoding
auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

// Line 123: Plaintext password stored in variable
password = string(parts[colonIndex+1:])
```

**Attack Vectors:**
1. Memory dumps (crash dumps, core dumps)
2. Swap file exposure
3. Process memory inspection via `/proc/<pid>/mem`
4. Heap spraying attacks

**Impact:**
- Credential exposure in memory dumps
- Forensic recovery of credentials
- Cross-process credential theft (if privileges escalated)

**Remediation:**
1. Use `mlock()` to prevent swapping of credential pages
2. Zero memory immediately after use
3. Use secure string types (e.g., `golang.org/x/crypto/ssh/terminal`)

```go
import "runtime"

// Secure credential handling
func (cs *CredentialStore) Store(registry, username, password string) error {
    // Convert to byte slices for secure handling
    passBytes := []byte(password)
    defer func() {
        // Zero memory
        for i := range passBytes {
            passBytes[i] = 0
        }
        runtime.GC() // Force garbage collection
    }()

    // Continue with secure storage...
}
```

---

### C3: Insecure TLS Configuration in Multiple Locations
**Files:**
- `/pkg/network/connection_pool.go` (Line 230)
- `/pkg/client/generic/client.go` (Lines 290-291)
- `/pkg/client/harbor/client.go` (Line 131)
- `/pkg/client/quay/client.go` (Line 139)

**Description:**
Multiple locations allow `InsecureSkipVerify: true` without proper warnings or restrictions, exposing the system to MITM attacks.

```go
// Line 291: Allows complete TLS bypass
transport.TLSClientConfig = &tls.Config{
    InsecureSkipVerify: true,
}
```

**Attack Scenario:**
1. Attacker performs ARP spoofing or DNS hijacking
2. Freightliner connects to attacker's registry (no certificate validation)
3. Attacker intercepts credentials and manipulates container images
4. Compromised images deployed to production

**Impact:**
- **Man-in-the-Middle attacks**
- Credential theft over network
- Malicious image injection
- Supply chain compromise

**Remediation:**
1. **Remove all `InsecureSkipVerify: true` from production code**
2. Add strict certificate validation
3. Implement certificate pinning for known registries
4. Add audit logging when insecure mode is attempted

```go
// Secure TLS configuration
func createSecureTLSConfig(registry string) (*tls.Config, error) {
    certPool, err := x509.SystemCertPool()
    if err != nil {
        return nil, err
    }

    return &tls.Config{
        RootCAs:            certPool,
        MinVersion:         tls.VersionTLS13,
        InsecureSkipVerify: false, // NEVER allow this in production
        ServerName:         registry,
    }, nil
}

// Add audit logging
if config.InsecureSkipVerify {
    logger.Warn("SECURITY RISK: TLS certificate verification disabled",
        map[string]interface{}{
            "registry": registry,
            "caller": "GetCallerInfo()",
        })
}
```

---

## 🟠 HIGH SEVERITY ISSUES

### H1: Password Logging Risk in Login Flow
**File:** `/cmd/login.go` (Lines 108-110, 138-140)

**Description:**
Logger with structured fields may inadvertently log sensitive information:

```go
logger.WithFields(map[string]interface{}{
    "registry": registry,
    "username": username, // Username logged (PII)
}).Info("Authentication successful")
```

**Issue:** If logger configuration changes or debug mode is enabled, this could log passwords/tokens.

**Remediation:**
```go
// Redact sensitive fields
logger.WithFields(map[string]interface{}{
    "registry": registry,
    "username": "[REDACTED]", // Never log usernames in production
    "auth_method": "basic",
}).Info("Authentication successful")
```

---

### H2: API Key in Environment Variables
**Files:**
- `/pkg/server/middleware.go` (Line 214)
- `/docker-compose.yml` (Lines 31-32)

**Description:**
API keys stored in environment variables can be exposed via:
- Process listings (`ps auxe`)
- Container inspection (`docker inspect`)
- CI/CD logs
- Error messages

**Example from code:**
```go
// Line 214: API key from environment
if apiKey == "" || apiKey != s.cfg.Server.APIKey {
```

**Remediation:**
1. Use secrets management (AWS Secrets Manager, HashiCorp Vault)
2. Load API keys from encrypted files with restricted permissions
3. Rotate API keys regularly
4. Implement API key hashing (store bcrypt hash, compare during auth)

---

### H3: Insufficient Input Validation in Server Handlers
**File:** `/pkg/server/handlers.go`

**Description:**
HTTP handlers accept user input without comprehensive validation:

```go
// Line 29: No validation of registry/repo format
source := fmt.Sprintf("%s/%s", req.SourceRegistry, req.SourceRepo)
```

**Attack Vectors:**
1. Path traversal via malicious registry names: `../../etc/passwd`
2. SSRF via internal endpoints: `http://169.254.169.254/metadata`
3. Command injection via special characters

**Remediation:**
```go
func validateRegistryPath(registry, repo string) error {
    // Check for path traversal
    if strings.Contains(registry, "..") || strings.Contains(repo, "..") {
        return errors.New("path traversal detected")
    }

    // Validate format (RFC 3986 for URI)
    registryPattern := regexp.MustCompile(`^[a-z0-9]([a-z0-9-\.]*[a-z0-9])?(\:[0-9]+)?$`)
    if !registryPattern.MatchString(registry) {
        return fmt.Errorf("invalid registry format: %s", registry)
    }

    // Validate repository name (Docker naming rules)
    repoPattern := regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*(/[a-z0-9]+([._-][a-z0-9]+)*)*$`)
    if !repoPattern.MatchString(repo) {
        return fmt.Errorf("invalid repository format: %s", repo)
    }

    return nil
}
```

---

### H4: Weak Rate Limiting Implementation
**File:** `/pkg/server/middleware.go` (Lines 238-282)

**Description:**
In-memory rate limiter is vulnerable to:
1. **IP spoofing** via `X-Forwarded-For` header manipulation
2. **Memory exhaustion** (unbounded `clients` map)
3. **Race conditions** in token calculation

```go
// Line 298: Vulnerable to header injection
if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
    ips := strings.Split(xff, ",")
    if len(ips) > 0 {
        return strings.TrimSpace(ips[0]) // Attacker controls this
    }
}
```

**Remediation:**
1. Validate and sanitize IP addresses
2. Implement maximum map size with LRU eviction
3. Use atomic operations for token counting
4. Add per-endpoint rate limits (not just per-IP)

---

### H5: Secrets in Configuration Files
**Files:**
- `/docker-compose.prod.yml` (Line 148): `GF_SECURITY_ADMIN_PASSWORD=admin123`
- `/docker-compose.yml` (Line 90): `GF_SECURITY_ADMIN_PASSWORD=admin`
- `/docker-compose.yml` (Line 123): `MINIO_ROOT_PASSWORD=minioadmin123`

**Description:**
Default/weak passwords committed to version control.

**Impact:**
- Unauthorized access to Grafana dashboards
- MinIO object storage compromise
- Credential stuffing attacks

**Remediation:**
1. **Remove all passwords from git history** (use `git filter-branch`)
2. Use environment variable substitution: `${GRAFANA_PASSWORD:?Password required}`
3. Implement secrets rotation policy
4. Add pre-commit hooks to detect secrets

```bash
# Pre-commit hook to detect secrets
git secrets --scan --no-verify
```

---

### H6: No Authentication on Health/Metrics Endpoints
**File:** `/pkg/server/middleware.go` (Lines 195-197)

**Description:**
Health and metrics endpoints bypass authentication, potentially exposing sensitive operational data.

```go
// Line 195: Bypasses auth middleware
if strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/metrics") {
    next.ServeHTTP(w, r)
    return
}
```

**Exposed Information:**
- System resource usage
- Active job counts
- Error rates
- Registry endpoints
- Version information

**Remediation:**
1. Require authentication for `/metrics` endpoint
2. Implement separate `/health/ready` and `/health/live` endpoints
3. Use IP whitelisting for monitoring systems
4. Sanitize metrics to remove sensitive labels

---

### H7: Excessive Error Information Disclosure
**File:** `/pkg/server/handlers.go` (Multiple locations)

**Description:**
Error messages leak internal implementation details:

```go
// Line 18: Exposes internal error messages
s.writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %s", err))
```

**Information Leaked:**
- File paths
- Database schema
- Stack traces
- Internal service names

**Remediation:**
```go
// Generic error for external users
s.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")

// Detailed error for internal logging only
s.logger.WithFields(map[string]interface{}{
    "error": err.Error(),
    "request_id": requestID,
}).Error("Request validation failed")
```

---

## 🟡 MEDIUM SEVERITY ISSUES

### M1: Timing Attack Vulnerability in API Key Comparison
**File:** `/pkg/server/middleware.go` (Line 214)

**Description:**
String comparison for API keys is vulnerable to timing attacks:

```go
if apiKey == "" || apiKey != s.cfg.Server.APIKey {
```

**Remediation:**
```go
import "crypto/subtle"

if subtle.ConstantTimeCompare([]byte(apiKey), []byte(s.cfg.Server.APIKey)) != 1 {
    // Unauthorized
}
```

---

### M2: Unbounded Memory Growth in Rate Limiter
**File:** `/pkg/server/middleware.go` (Line 74)

**Description:**
Cleanup goroutine only runs every minute, allowing memory growth from malicious clients.

**Remediation:**
1. Implement maximum client map size (e.g., 10,000 entries)
2. Use LRU cache instead of plain map
3. Trigger cleanup more frequently under high load

---

### M3: Missing CSRF Protection
**File:** `/pkg/server/middleware.go`

**Description:**
No CSRF token validation for state-changing operations (POST, PUT, DELETE).

**Remediation:**
Implement CSRF middleware using `gorilla/csrf`:
```go
import "github.com/gorilla/csrf"

csrfMiddleware := csrf.Protect(
    []byte(csrfKey),
    csrf.Secure(true),
    csrf.SameSite(csrf.SameSiteStrictMode),
)
```

---

### M4: Weak CORS Configuration
**File:** `/pkg/server/middleware.go` (Lines 172-174)

**Description:**
Allows `Access-Control-Allow-Origin: *` by default, enabling CSRF from any origin.

**Remediation:**
1. Never use `*` in production
2. Implement strict origin whitelist
3. Add `Access-Control-Allow-Credentials` validation

---

### M5: No Security Headers
**File:** `/pkg/server/middleware.go`

**Description:**
Missing critical security headers:
- `X-Frame-Options`
- `X-Content-Type-Options`
- `Content-Security-Policy`
- `Strict-Transport-Security`

**Remediation:**
```go
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

        next.ServeHTTP(w, r)
    })
}
```

---

### M6: Insufficient Logging of Security Events
**Files:** Multiple

**Description:**
Security events not consistently logged:
- Failed authentication attempts
- Authorization failures
- Suspicious request patterns
- Configuration changes

**Remediation:**
Implement centralized security audit logging:
```go
type SecurityAuditLog struct {
    Timestamp   time.Time
    EventType   string // AUTH_FAILURE, AUTHZ_FAILURE, SUSPICIOUS_REQUEST
    Username    string
    SourceIP    string
    RequestPath string
    Details     map[string]interface{}
}
```

---

### M7: No Request Size Limits
**File:** `/pkg/server/handlers.go`

**Description:**
No limits on request body size could enable DoS attacks.

**Remediation:**
```go
// Add middleware
func (s *Server) requestSizeLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024) // 10MB limit
        next.ServeHTTP(w, r)
    })
}
```

---

### M8: Predictable Job IDs
**File:** `/pkg/server/jobs.go` (inferred)

**Description:**
If job IDs are sequential or predictable, attackers can enumerate jobs.

**Remediation:**
Use cryptographically secure random UUIDs:
```go
import "github.com/google/uuid"

jobID := uuid.New().String()
```

---

## 🔵 LOW SEVERITY ISSUES

### L1: Overly Permissive File Permissions
**File:** `/pkg/testing/validation/quality_gates.go` (Line 457)

```go
os.WriteFile(filepath, data, 0644) // World-readable
```

**Remediation:**
```go
os.WriteFile(filepath, data, 0600) // Owner only
```

---

### L2: Missing Input Sanitization in Logs
**File:** Multiple locations

**Description:**
User input logged without sanitization could inject malicious content into log parsers.

**Remediation:**
```go
func sanitizeForLog(input string) string {
    // Remove control characters
    return strings.Map(func(r rune) rune {
        if r < 32 || r == 127 {
            return -1
        }
        return r
    }, input)
}
```

---

### L3: No Timeout on Context Operations
**File:** `/pkg/sync/batch.go` (Line 210)

**Description:**
Some operations don't enforce timeouts, potentially causing goroutine leaks.

**Remediation:**
Always use context with timeout:
```go
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
```

---

### L4: Potential Integer Overflow in Backoff Calculation
**File:** `/pkg/client/common/base_transport.go` (Line 197)

```go
baseDelay := time.Duration(1<<uint(backoffFactor)) * 200 * time.Millisecond
```

**Note:** Marked with `#nosec G115` but still risky if cap is increased.

**Remediation:**
Add explicit overflow check:
```go
if backoffFactor > 20 { // Prevent overflow
    backoffFactor = 20
}
```

---

### L5: Deprecated Cryptographic Hash (MD5)
**Status:** Not found in main codebase (✅ GOOD)

**Note:** Codebase correctly uses SHA256/SHA512 for digests.

---

## DEPENDENCY VULNERABILITIES

### Go Module Analysis

**Critical Dependencies to Monitor:**
```
github.com/sigstore/cosign/v2 v2.2.2
github.com/aws/aws-sdk-go-v2 v1.38.1
golang.org/x/crypto (indirect)
golang.org/x/net (indirect)
```

**Recommendations:**
1. Run `govulncheck ./...` regularly (install via `go install golang.org/x/vuln/cmd/govulncheck@latest`)
2. Enable Dependabot alerts in GitHub
3. Pin indirect dependencies with `go mod tidy`
4. Monitor advisories: https://github.com/advisories

### Example Vulnerability Check:
```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan for vulnerabilities
govulncheck ./...

# Update vulnerable packages
go get -u <vulnerable-package>
go mod tidy
```

---

## OWASP TOP 10 COMPLIANCE

| OWASP 2021 Category | Status | Issues Found |
|---------------------|--------|--------------|
| A01: Broken Access Control | 🟠 | H3, M3, M8 |
| A02: Cryptographic Failures | 🔴 | C2, C3, M1 |
| A03: Injection | 🔴 | C1, H3 |
| A04: Insecure Design | 🟡 | H4, M2 |
| A05: Security Misconfiguration | 🟠 | C3, H5, H6, M4, M5 |
| A06: Vulnerable Components | 🟡 | (Dependency scan needed) |
| A07: Authentication Failures | 🟠 | H1, H6 |
| A08: Software & Data Integrity | 🟡 | (Requires runtime analysis) |
| A09: Logging Failures | 🟡 | M6, L2 |
| A10: SSRF | 🟡 | H3 (partial) |

---

## SECURE CODING CHECKLIST

### ✅ GOOD PRACTICES OBSERVED
- ✅ Uses prepared statements-equivalent for Docker Registry API (no SQL injection)
- ✅ Implements rate limiting (though needs improvement)
- ✅ Uses HTTPS by default
- ✅ Credential storage follows Docker config format
- ✅ Structured logging with appropriate levels
- ✅ Context-based cancellation
- ✅ Input validation in most places
- ✅ Secrets validator with comprehensive checks

### ❌ IMPROVEMENTS NEEDED
- ❌ No secrets scanning in CI/CD
- ❌ Missing security headers
- ❌ Insufficient authentication audit logging
- ❌ No CSRF protection
- ❌ Weak CORS policy
- ❌ Plain HTTP allowed in some configurations
- ❌ No certificate pinning

---

## REMEDIATION PRIORITY

### PHASE 1: IMMEDIATE (Week 1)
1. **C1**: Fix command injection in credential helpers
2. **C3**: Remove all `InsecureSkipVerify: true` from non-test code
3. **H5**: Remove hardcoded passwords from repository

### PHASE 2: CRITICAL (Week 2-3)
4. **C2**: Implement secure memory handling for credentials
5. **H1**: Sanitize all logging of authentication data
6. **H2**: Move API keys to secrets management
7. **H3**: Add comprehensive input validation

### PHASE 3: HIGH PRIORITY (Week 4-6)
8. **H4**: Improve rate limiting with proper IP validation
9. **H6**: Add authentication to metrics endpoint
10. **H7**: Implement generic error messages for users

### PHASE 4: MEDIUM PRIORITY (Week 7-10)
11. **M1-M8**: Address all medium severity issues
12. Add security headers middleware
13. Implement CSRF protection
14. Fix CORS configuration

### PHASE 5: HARDENING (Ongoing)
15. **L1-L5**: Address low severity issues
16. Run regular dependency scans
17. Implement security testing in CI/CD
18. Conduct penetration testing

---

## SECURITY TESTING RECOMMENDATIONS

### 1. Automated Security Testing
```yaml
# .github/workflows/security.yml
- name: Run GoSec
  run: |
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    gosec -exclude=G104 ./...

- name: Run Trivy
  run: |
    trivy fs --security-checks vuln,config .

- name: Check for secrets
  run: |
    docker run --rm -v ${PWD}:/path zricethezav/gitleaks:latest detect --source="/path" -v
```

### 2. Manual Penetration Testing
- API fuzzing with OWASP ZAP
- Credential stuffing simulation
- SSRF testing against internal endpoints
- Path traversal testing in all file operations

### 3. Code Review Checklist
- [ ] No plaintext passwords in code/config
- [ ] All user input validated
- [ ] TLS certificate verification enabled
- [ ] Security headers present
- [ ] Rate limiting implemented
- [ ] Authentication on all sensitive endpoints
- [ ] Audit logging for security events

---

## COMPLIANCE CONSIDERATIONS

### PCI-DSS (If handling payment data)
- Requires encryption at rest and in transit ✅ (KMS integration present)
- Credential storage needs HSM backing (not implemented)

### GDPR (If handling EU user data)
- Username logging may violate privacy (see H1)
- Implement right to erasure for logs
- Add consent management

### SOC 2
- Audit logging insufficient (see M6)
- Access control needs improvement (see H6)
- Incident response documentation needed

---

## INCIDENT RESPONSE RECOMMENDATIONS

### Detection
1. Monitor for repeated authentication failures
2. Alert on TLS errors (potential MITM)
3. Track anomalous request patterns
4. Monitor rate limit violations

### Response Playbook
```bash
# Suspected credential compromise
1. Rotate all API keys immediately
2. Force re-authentication for all users
3. Review audit logs for unauthorized access
4. Check Docker config files for tampering

# Suspected MITM attack
1. Enable TLS verification immediately
2. Regenerate all session tokens
3. Audit recent image pulls/pushes
4. Scan container images for malware
```

---

## TOOLS & RESOURCES

### Security Scanning Tools
- **Static Analysis**: GoSec, Semgrep, SonarQube
- **Dependency Scanning**: govulncheck, Snyk, Dependabot
- **Secrets Detection**: GitLeaks, TruffleHog
- **Container Scanning**: Trivy, Grype, Clair
- **Dynamic Testing**: OWASP ZAP, Burp Suite

### References
- OWASP Go Security Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Go_SCP.cheat_sheet.html
- CWE Top 25: https://cwe.mitre.org/top25/
- NIST Secure Software Development Framework: https://csrc.nist.gov/projects/ssdf

---

## CONCLUSION

Freightliner has a solid foundation but requires **immediate attention** to 3 critical vulnerabilities before production deployment. The development team has implemented good practices in areas like encryption and structured logging, but security hardening is incomplete.

**Primary Risk:** Command injection and TLS bypass create immediate attack vectors for credential theft and supply chain compromise.

**Recommendation:** Complete Phase 1-2 remediation (critical fixes) before any production release. Implement automated security testing in CI/CD pipeline. Consider third-party security audit for production readiness certification.

---

**Report Status:** DRAFT - For Internal Review
**Classification:** CONFIDENTIAL
**Next Review:** 2025-12-20 (after remediation)

*This report was generated by an automated security audit. Manual verification of findings is recommended.*
