# CI/CD Troubleshooting Runbook

## Overview

This runbook provides comprehensive troubleshooting procedures for the enhanced CI/CD pipeline in the Freightliner project. It covers common issues, diagnostic procedures, and resolution steps for all components of the reliability system.

## Table of Contents

1. [Emergency Procedures](#emergency-procedures)
2. [Component Diagnostics](#component-diagnostics)
3. [Common Issues and Solutions](#common-issues-and-solutions)
4. [Performance Troubleshooting](#performance-troubleshooting)
5. [Reliability System Issues](#reliability-system-issues)
6. [Monitoring and Alerting Issues](#monitoring-and-alerting-issues)
7. [Docker Build Problems](#docker-build-problems)
8. [Network and Connectivity Issues](#network-and-connectivity-issues)
9. [Recovery Procedures](#recovery-procedures)
10. [Escalation Procedures](#escalation-procedures)

## Emergency Procedures

### Critical Pipeline Failure

**Immediate Actions (Within 5 minutes):**

1. **Assess Impact:**
```bash
# Check current pipeline status
gh run list --limit 5
gh run view <run-id>
```

2. **Quick Disable of Reliability Features:**
```yaml
# In .github/workflows/ci.yml
env:
  PIPELINE_RELIABILITY_ENABLED: 'false'
```

3. **Emergency Rollback:**
```bash
# Revert to last known good configuration
git revert <commit-hash> --no-edit
git push origin main
```

4. **Notify Team:**
```bash
# Use configured alerting channels
# Manual notification if automated alerts fail
```

### Complete System Outage

**Actions for Total CI/CD Failure:**

1. **Switch to Manual Deployment:**
   - Disable automatic triggers
   - Use local build and test procedures
   - Manual Docker builds for urgent releases

2. **Investigate Root Cause:**
```bash
# Check GitHub Actions status
curl -s https://www.githubstatus.com/api/v2/incidents.json

# Check runner availability
gh api repos/:owner/:repo/actions/runners
```

3. **Temporary Workarounds:**
   - Skip non-critical tests: `SKIP_TESTS=true`
   - Disable security scans: `SKIP_SECURITY=true`
   - Use simple Dockerfile: `docker build -f Dockerfile.simple`

## Component Diagnostics

### Pipeline Health Check

**Comprehensive System Diagnosis:**

```bash
# 1. Initialize recovery environment
.github/scripts/pipeline-recovery.sh init

# 2. Run comprehensive health check
.github/scripts/pipeline-recovery.sh health-check

# 3. Generate detailed diagnostics
.github/scripts/pipeline-recovery.sh diagnostics

# 4. Check component status
.github/scripts/pipeline-recovery.sh status
```

**Health Check Output Analysis:**

```bash
# Look for these indicators in health check output:
✅ Component healthy - No action needed
⚠️  Component degraded - Monitor closely
❌ Component failed - Immediate attention required
🔧 Recovery in progress - Wait for completion
```

### Go Environment Diagnostics

**Check Go Environment Health:**

```bash
# 1. Version verification
go version
go env GOVERSION GOOS GOARCH

# 2. Module verification
go mod download
go mod verify

# 3. Build capability test
go build -v ./...

# 4. Environment variables
echo $GOPROXY
echo $GOSUMDB
echo $GOPRIVATE
```

**Common Go Issues:**

| Issue | Symptoms | Solution |
|-------|----------|----------|
| Version mismatch | Build errors, module conflicts | Update go.mod and Docker base images |
| Module download failure | Network timeouts, checksum errors | Clear module cache: `go clean -modcache` |
| Build cache corruption | Inconsistent builds | Clear build cache: `go clean -cache` |
| Proxy issues | Slow downloads, connection errors | Check GOPROXY settings, use fallback |

### Docker Environment Diagnostics

**Check Docker Health:**

```bash
# 1. Basic Docker functionality
docker --version
docker info

# 2. Buildx availability
docker buildx version
docker buildx ls

# 3. Registry connectivity
docker pull alpine:latest
docker push <test-image>

# 4. Build system test
docker build -t test:diagnostics - <<EOF
FROM alpine:latest
RUN echo "Docker build working"
EOF
```

**Docker Troubleshooting Commands:**

```bash
# Clean up Docker system
docker system prune -af
docker builder prune -af

# Check disk space
docker system df

# Registry health check
curl -f http://localhost:5100/v2/ || echo "Registry unhealthy"

# Container logs
docker logs <container-id>
```

### Network Connectivity Diagnostics

**Network Health Verification:**

```bash
# 1. Basic connectivity
ping -c 3 8.8.8.8

# 2. DNS resolution
nslookup proxy.golang.org
nslookup registry-1.docker.io

# 3. HTTPS connectivity
curl -I https://proxy.golang.org
curl -I https://registry-1.docker.io

# 4. GitHub API access
curl -I https://api.github.com/user
```

**Network Issue Resolution:**

```bash
# Test alternative proxies
export GOPROXY="https://goproxy.cn,https://goproxy.io,direct"

# Test with different DNS
export DNS_SERVERS="8.8.8.8,1.1.1.1"

# Check firewall/proxy settings
env | grep -i proxy
```

## Common Issues and Solutions

### Issue: Build Failures Due to Version Inconsistency

**Symptoms:**
- Go version mismatch errors
- Module compatibility issues
- Build tool version conflicts

**Diagnostic Commands:**
```bash
# Check all version configurations
grep -r "go 1\." .
grep -r "golang:" .
grep -r "GO_VERSION" .
```

**Resolution:**
```bash
# 1. Update go.mod
go mod edit -go=1.23.4

# 2. Update Dockerfiles
sed -i 's/golang:1\.[0-9]\+\.[0-9]\+/golang:1.23.4/g' Dockerfile*

# 3. Update CI configuration
# Update GO_VERSION in .github/workflows/ci.yml

# 4. Update toolchain
go mod edit -toolchain=go1.23.4
```

### Issue: Circuit Breaker False Positives

**Symptoms:**
- Services marked as "open" incorrectly
- Operations blocked unnecessarily
- High false positive rate in alerts

**Diagnostic Commands:**
```bash
# Check circuit breaker states
find .ci-reliability -name "circuit_breaker_*" -exec cat {} \;

# Review failure history
cat .ci-reliability/*.log | grep "Circuit breaker"
```

**Resolution:**
```bash
# 1. Adjust failure threshold (in ci-reliability.sh)
readonly CIRCUIT_BREAKER_FAILURE_THRESHOLD=5  # Increase from 3

# 2. Increase timeout (in ci-reliability.sh)
readonly CIRCUIT_BREAKER_TIMEOUT=600  # Increase from 300

# 3. Reset specific circuit breaker
rm .ci-reliability/circuit_breaker_<service-name>

# 4. Disable circuit breaker temporarily
export PIPELINE_RELIABILITY_ENABLED=false
```

### Issue: Test Failures in Package Isolation

**Symptoms:**
- Individual package tests fail
- Integration tests have dependency issues
- Race conditions in parallel testing

**Diagnostic Commands:**
```bash
# Run tests with verbose output
go test -v -race ./pkg/problematic/package

# Check for resource conflicts
lsof | grep <test-resources>

# Examine test logs
find . -name "*.test.log" -exec cat {} \;
```

**Resolution:**
```bash
# 1. Disable package isolation temporarily
# In run-tests action: package-isolation: 'false'

# 2. Reduce test parallelism
export GOMAXPROCS=1

# 3. Add test timeouts
go test -timeout=30s ./...

# 4. Fix race conditions
go test -race ./... > race-report.txt
```

### Issue: Docker Build Timeout or Failure

**Symptoms:**
- Docker builds hang or timeout
- Registry connectivity issues
- Out of disk space errors

**Diagnostic Commands:**
```bash
# Check Docker daemon
docker info
docker system df

# Test registry connectivity
curl -f http://localhost:5100/v2/

# Check build logs
docker buildx build --progress=plain . 2>&1 | tee build.log
```

**Resolution:**
```bash
# 1. Clean up Docker system
docker system prune -af
docker builder prune -af

# 2. Increase build timeout
timeout 1800 docker build .  # 30 minutes

# 3. Use alternative registry
export REGISTRY_HOST=alternative-registry:5000

# 4. Build with reduced parallelism
export DOCKER_BUILDKIT_INLINE_CACHE=1
docker build --build-arg BUILDKIT_INLINE_CACHE=1 .
```

## Performance Troubleshooting

### Slow Build Performance

**Performance Analysis:**

```bash
# 1. Measure baseline performance
time go build ./...

# 2. Check cache effectiveness
go clean -cache
time go build ./...  # First run
time go build ./...  # Second run (should be faster)

# 3. Analyze build bottlenecks
go build -x ./... 2>&1 | grep -E "(compile|link)" | head -20
```

**Performance Optimization:**

```bash
# 1. Enable build cache
export GOCACHE=$HOME/.cache/go-build

# 2. Increase build parallelism
export GOMAXPROCS=4

# 3. Use optimized build flags
go build -ldflags="-s -w" -trimpath ./...

# 4. Profile build performance
go build -x ./... 2>&1 | awk '/compile/ {print $NF}' | sort | uniq -c | sort -nr
```

### Cache Performance Issues

**Cache Diagnosis:**

```bash
# 1. Check cache hit rates
# Look for "CACHE HIT" vs "CACHE MISS" in CI logs

# 2. Verify cache configuration
grep -r "actions/cache" .github/

# 3. Check cache size limits
du -sh ~/.cache/go-build
du -sh ~/go/pkg/mod
```

**Cache Optimization:**

```bash
# 1. Adjust cache keys
# Include more specific hash components

# 2. Increase cache version
CACHE_VERSION: 'v3'  # Increment to invalidate

# 3. Optimize cache paths
path: |
  ~/.cache/go-build
  ~/go/pkg/mod
  ~/.cache/docker
```

### Memory and Resource Issues

**Resource Monitoring:**

```bash
# 1. Monitor memory usage during build
ps aux | grep go
docker stats

# 2. Check for memory leaks
valgrind --tool=memcheck go test ./...

# 3. Monitor disk usage
df -h
du -sh .
```

**Resource Optimization:**

```bash
# 1. Reduce parallel execution
export GOMAXPROCS=2
export TEST_PARALLELISM=1

# 2. Enable memory optimization
export GOGC=50  # More aggressive garbage collection

# 3. Clean up intermediate files
go clean -testcache -modcache -cache
docker system prune -f
```

## Reliability System Issues

### Retry Mechanism Problems

**Symptoms:**
- Operations fail without retries
- Excessive retry attempts
- Retry backoff not working

**Diagnostic Commands:**
```bash
# Check retry configuration
grep -r "MAX_RETRY_ATTEMPTS" .
grep -r "retry_with_circuit_breaker" .github/scripts/

# Review retry logs
grep "Retry attempt" .ci-reliability/*.log
```

**Resolution:**
```bash
# 1. Adjust retry parameters
readonly DEFAULT_MAX_RETRIES=5
readonly DEFAULT_INITIAL_WAIT=2
readonly DEFAULT_MAX_WAIT=120

# 2. Add operation-specific retry logic
retry_with_circuit_breaker "custom_operation" "service" 3 5 60 command args

# 3. Check for non-retryable errors
# Review is_retryable_error function in ci-reliability.sh
```

### Recovery Script Failures

**Common Recovery Issues:**

```bash
# 1. Script permissions
chmod +x .github/scripts/*.sh

# 2. Missing dependencies (macOS compatibility)
# Replace 'timeout' with 'gtimeout' or remove timeout constraints

# 3. State directory issues
mkdir -p .pipeline-recovery
chown -R $USER .pipeline-recovery

# 4. Environment variable issues
export GITHUB_WORKSPACE=$(pwd)
export PIPELINE_RELIABILITY_ENABLED=true
```

### Health Check Failures

**Health Check Troubleshooting:**

```bash
# 1. Manual health checks
.github/scripts/pipeline-recovery.sh check-go
.github/scripts/pipeline-recovery.sh check-docker
.github/scripts/pipeline-recovery.sh check-network

# 2. Component-specific diagnosis
go version && echo "Go OK" || echo "Go FAILED"
docker info && echo "Docker OK" || echo "Docker FAILED"

# 3. Reset health check state
rm -rf .pipeline-recovery/health-*
```

## Monitoring and Alerting Issues

### Missing or Incorrect Alerts

**Alert Troubleshooting:**

```bash
# 1. Check alert configuration
grep -r "SLACK_WEBHOOK_URL" .
grep -r "alert" .github/scripts/

# 2. Test webhook connectivity
curl -X POST $SLACK_WEBHOOK_URL -d '{"text":"Test alert"}'

# 3. Verify alert thresholds
grep -r "THRESHOLD" .github/scripts/pipeline-monitoring.sh
```

**Alert Configuration:**

```yaml
# Environment variables for alerting
SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
TEAMS_WEBHOOK_URL: ${{ secrets.TEAMS_WEBHOOK_URL }}
EMAIL_NOTIFICATION_ENABLED: 'true'

# Alert thresholds (in pipeline-monitoring.sh)
readonly SUCCESS_RATE_THRESHOLD=95
readonly ERROR_RATE_THRESHOLD=30
readonly DURATION_THRESHOLD=30
```

### Dashboard Generation Issues

**Dashboard Troubleshooting:**

```bash
# 1. Generate dashboard manually
.github/scripts/pipeline-monitoring.sh generate-dashboard

# 2. Check dashboard dependencies
command -v jq >/dev/null || echo "jq not installed"
command -v curl >/dev/null || echo "curl not available"

# 3. Verify data collection
ls -la .pipeline-monitoring/
cat .pipeline-monitoring/metrics.json
```

### Metrics Collection Problems

**Metrics Diagnosis:**

```bash
# 1. Check metrics files
find . -name "*.json" -path "*/.pipeline-*" -exec cat {} \;

# 2. Validate JSON format
jq . .pipeline-monitoring/metrics.json

# 3. Check metrics permissions
ls -la .pipeline-monitoring/
```

## Docker Build Problems

### Multi-stage Build Issues

**Common Docker Problems:**

```bash
# 1. Stage not found errors
# Check stage names in Dockerfile.optimized
grep "^FROM.*AS" Dockerfile.optimized

# 2. Build context issues
# Check .dockerignore configuration
cat .dockerignore

# 3. Cache mount problems
# Verify BuildKit is enabled
export DOCKER_BUILDKIT=1
```

**Docker Build Fixes:**

```dockerfile
# Fix stage naming issues
FROM golang:1.23.4-alpine AS base
FROM base AS dependencies  # Ensure correct stage reference

# Fix build path issues
WORKDIR /workspace
COPY . .
RUN go build -o /tmp/freightliner .  # Use root path, not ./cmd/freightliner

# Fix tool installation issues (Alpine linker problem)
RUN apk add --no-cache gcc musl-dev binutils-gold
```

### Registry Connectivity Issues

**Registry Troubleshooting:**

```bash
# 1. Test registry health
curl -f http://localhost:5100/v2/ || echo "Registry down"

# 2. Check registry logs
docker logs $(docker ps -q --filter ancestor=registry:2)

# 3. Test authentication
echo $REGISTRY_PASSWORD | docker login -u $REGISTRY_USER --password-stdin $REGISTRY_HOST

# 4. Use alternative registry
export REGISTRY_HOST=ghcr.io
```

## Network and Connectivity Issues

### Proxy and Firewall Problems

**Network Diagnosis:**

```bash
# 1. Check proxy settings
env | grep -i proxy

# 2. Test direct vs proxy connection
curl --noproxy "*" https://proxy.golang.org
curl https://proxy.golang.org

# 3. DNS resolution issues
nslookup proxy.golang.org
nslookup registry-1.docker.io
```

**Network Resolution:**

```bash
# 1. Configure Go proxy fallbacks
export GOPROXY="https://proxy.golang.org,https://goproxy.cn,direct"

# 2. Configure Docker registry mirrors
# Add to docker daemon.json
{
  "registry-mirrors": ["https://mirror.gcr.io"]
}

# 3. Bypass proxy for local services
export NO_PROXY="localhost,127.0.0.1,.local"
```

### DNS and Certificate Issues

**SSL/TLS Troubleshooting:**

```bash
# 1. Test certificate validation
openssl s_client -connect proxy.golang.org:443 -servername proxy.golang.org

# 2. Check certificate store
curl -v https://proxy.golang.org 2>&1 | grep -i certificate

# 3. Skip certificate validation (temporary)
export GOINSECURE="example.com/*"
```

## Recovery Procedures

### Automated Recovery

**Trigger Automatic Recovery:**

```bash
# 1. Full system recovery
.github/scripts/pipeline-recovery.sh auto-recover

# 2. Component-specific recovery
.github/scripts/pipeline-recovery.sh recover go
.github/scripts/pipeline-recovery.sh recover docker
.github/scripts/pipeline-recovery.sh recover network

# 3. Reset all circuit breakers
.github/scripts/pipeline-recovery.sh reset-circuit-breakers
```

### Manual Recovery Steps

**Step-by-Step Recovery:**

```bash
# 1. Stop current operations
# Cancel running workflows in GitHub UI

# 2. Clean local state
rm -rf .pipeline-recovery/
rm -rf .ci-reliability/

# 3. Reset environment
unset $(env | grep PIPELINE_ | cut -d= -f1)
export PIPELINE_RELIABILITY_ENABLED=true

# 4. Verify base functionality
go version
docker --version
make test

# 5. Restart reliability system
.github/scripts/pipeline-recovery.sh init
.github/scripts/pipeline-recovery.sh health-check
```

### Data Recovery

**Recover Lost Metrics:**

```bash
# 1. Export current metrics
.github/scripts/pipeline-monitoring.sh export-metrics > backup-metrics.json

# 2. Restore from backup
.github/scripts/pipeline-monitoring.sh import-metrics < backup-metrics.json

# 3. Regenerate dashboard
.github/scripts/pipeline-monitoring.sh generate-dashboard
```

## Escalation Procedures

### When to Escalate

**Critical Issues (Immediate Escalation):**
- Complete CI/CD system failure > 30 minutes
- Security vulnerabilities in pipeline
- Data loss or corruption
- Customer-impacting build failures

**Major Issues (Escalate within 2 hours):**
- Performance degradation > 50%
- Reliability system failures
- Persistent build failures
- Monitoring system outages

### Escalation Contacts

**Internal Team:**
1. **Primary DevOps Engineer** - Immediate response
2. **Lead Developer** - Technical decisions
3. **Site Reliability Team** - Infrastructure issues
4. **Security Team** - Security-related problems

**External Support:**
1. **GitHub Support** - Actions/runner issues
2. **Docker Support** - Container platform issues
3. **Cloud Provider** - Infrastructure problems

### Escalation Information to Provide

**Technical Details:**
- Exact error messages and stack traces
- Affected components and services
- Timeline of events
- Steps already attempted
- Current system status

**Business Impact:**
- Services affected
- Number of users impacted
- Duration of outage
- Estimated recovery time

**Documentation:**
- Log files and diagnostics
- Configuration changes made
- Recovery attempts performed
- Screenshots or recordings

### Post-Incident Procedures

**Immediate (Within 24 hours):**
1. Document incident timeline
2. Identify root cause
3. Implement permanent fix
4. Update monitoring/alerting

**Follow-up (Within 1 week):**
1. Post-mortem meeting
2. Update runbooks
3. Improve monitoring
4. Team training if needed

**Long-term (Within 1 month):**
1. Process improvements
2. Tool upgrades
3. Documentation updates
4. Prevention measures

## Quick Reference Commands

### Emergency Commands
```bash
# Disable reliability features
export PIPELINE_RELIABILITY_ENABLED=false

# Quick rollback
git revert HEAD --no-edit && git push

# Force clean state
rm -rf .pipeline-* .ci-reliability && docker system prune -af

# Basic health check
go version && docker --version && make test
```

### Diagnostic Commands
```bash
# Full system diagnosis
.github/scripts/pipeline-recovery.sh diagnostics

# Component health checks
.github/scripts/pipeline-recovery.sh health-check

# Performance analysis
time make test && docker system df
```

### Recovery Commands
```bash
# Auto recovery
.github/scripts/pipeline-recovery.sh auto-recover

# Manual recovery
.github/scripts/pipeline-recovery.sh init && \
.github/scripts/pipeline-recovery.sh health-check
```

This runbook should be kept updated as new issues are discovered and resolved. Regular review and updates ensure it remains an effective troubleshooting resource.