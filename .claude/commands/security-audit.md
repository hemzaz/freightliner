# Security Audit Command

Perform a comprehensive security audit of the Freightliner codebase and infrastructure.

## What This Command Does

1. Runs all security scanning tools (gosec, dependency checks)
2. Reviews authentication and authorization patterns
3. Checks for credential exposure and secrets
4. Audits encryption implementations
5. Reviews container security configurations
6. Provides actionable security recommendations

## Usage

```bash
/security-audit [focus-area]
```

## Focus Areas

- `code` - Source code security scan
- `deps` - Dependency vulnerability scan
- `containers` - Docker/K8s security review
- `secrets` - Credential and secrets management
- `all` - Comprehensive audit (default)

## Audit Checklist

### 1. Code Security
```bash
# Run gosec security scanner
make security

# Check for hardcoded credentials
grep -r "password\|secret\|key\|token" --include="*.go" | grep -v "test"

# Check for SQL injection risks
grep -r "Exec\|Query" --include="*.go" pkg/

# Check error handling patterns
grep -r "err != nil" --include="*.go" | grep -v "return"
```

### 2. Dependency Security
```bash
# Check for known vulnerabilities
go list -json -m all | nancy sleuth

# Audit direct dependencies
go mod graph | grep "^freightliner"

# Check for outdated dependencies
go list -u -m all
```

### 3. Authentication & Authorization
- Review AWS IAM role usage in `pkg/client/ecr/`
- Review GCP service account usage in `pkg/client/gcr/`
- Check API key handling in `pkg/server/middleware.go`
- Verify CORS configuration
- Check TLS certificate validation

### 4. Secrets Management
- Review `pkg/secrets/` implementation
- Verify no secrets in environment variables logged
- Check secrets rotation capabilities
- Verify encryption at rest for sensitive data

### 5. Encryption
- Review `pkg/security/encryption/` implementation
- Verify AES-256 usage
- Check KMS integration (AWS KMS, GCP KMS)
- Verify TLS 1.2+ for all network communications

### 6. Container Security
```bash
# Scan Docker images
docker scan freightliner:latest

# Check Dockerfile best practices
hadolint Dockerfile

# Review Kubernetes security contexts
kubectl explain pod.spec.securityContext
```

### 7. Network Security
- Review network policies in `deployments/kubernetes/`
- Check ingress TLS configuration
- Verify no plain HTTP endpoints
- Review CORS allowed origins

## Security Scoring

Rate each area as:
- ✅ **SECURE** - Meets security best practices
- ⚠️ **NEEDS IMPROVEMENT** - Has minor issues
- ❌ **CRITICAL** - Requires immediate attention

## Deliverables

1. Security audit report in `docs/security-audit-<date>.md`
2. List of findings with severity (Critical/High/Medium/Low)
3. Remediation steps for each finding
4. Updated security documentation if needed
5. GitHub security advisory if critical issues found

## Compliance Checks

- OWASP Top 10 compliance
- CIS Docker Benchmark
- CIS Kubernetes Benchmark
- SOC 2 considerations
- GDPR data handling (if applicable)
