# Incident Response Workflow

Workflow for diagnosing and resolving production incidents.

## Severity Levels

### P0 - Critical (Production Down)
- Production service completely unavailable
- Data loss occurring
- Security breach detected
- **Response Time:** Immediate
- **Agents:** devops-troubleshooter, security-auditor, performance-engineer

### P1 - High (Severe Degradation)
- Significant performance degradation
- Partial service outage
- Critical feature broken
- **Response Time:** < 1 hour
- **Agents:** devops-troubleshooter, performance-engineer

### P2 - Medium (Minor Impact)
- Non-critical feature broken
- Performance issues affecting some users
- Workaround available
- **Response Time:** < 4 hours
- **Agents:** devops-troubleshooter, golang-pro

### P3 - Low (Minor Issue)
- Cosmetic issues
- Documentation errors
- Enhancement requests
- **Response Time:** < 24 hours
- **Agents:** golang-pro, api-documenter

## Incident Response Steps

### 1. Initial Assessment (devops-troubleshooter)
```bash
/incident-assess [description]
```

**Tasks:**
- Determine severity level
- Identify affected services
- Check recent deployments
- Review error logs
- Assess user impact

### 2. Immediate Mitigation (devops-troubleshooter)

**For P0/P1 Incidents:**
- Consider rollback if recent deployment
- Scale resources if capacity issue
- Enable circuit breakers if downstream issue
- Implement rate limiting if traffic spike

```bash
# Quick rollback
helm rollback freightliner -n freightliner-production

# Scale up
kubectl scale deployment/freightliner --replicas=10 -n freightliner-production
```

### 3. Root Cause Analysis (Multi-Agent)

#### Application Issues (golang-pro + devops-troubleshooter)
```bash
# Check application logs
kubectl logs -f deployment/freightliner -n freightliner-production --tail=1000

# Check for panics or errors
kubectl logs deployment/freightliner -n freightliner-production | grep -E "panic|error|fatal"

# Analyze recent changes
git log --oneline --since="24 hours ago"
```

#### Performance Issues (performance-engineer)
```bash
# Check metrics
kubectl top pod -l app=freightliner -n freightliner-production

# Review Prometheus metrics
# - HTTP response times
# - Error rates
# - Memory usage
# - Goroutine count

# CPU profiling if needed
kubectl exec -it <pod> -n freightliner-production -- curl http://localhost:8080/debug/pprof/profile?seconds=30
```

#### Security Issues (security-auditor)
```bash
# Check for unauthorized access
kubectl logs deployment/freightliner -n freightliner-production | grep "401\|403"

# Review network policies
kubectl describe networkpolicy freightliner -n freightliner-production

# Check for suspicious activity
# - Unusual traffic patterns
# - Failed authentication attempts
# - Privilege escalation attempts
```

#### Infrastructure Issues (cloud-architect + devops-troubleshooter)
```bash
# Check node health
kubectl get nodes
kubectl describe node <node-name>

# Check pod events
kubectl get events -n freightliner-production --sort-by='.lastTimestamp'

# Check cluster resources
kubectl top nodes
```

### 4. Fix Implementation

#### Code Fix (golang-pro)
```bash
# Create hotfix branch
git checkout -b hotfix/incident-$(date +%Y%m%d-%H%M)

# Implement fix
# ... make changes ...

# Test locally
make test-ci
make quality

# Commit and push
git commit -m "hotfix: [description]"
git push origin hotfix/incident-$(date +%Y%m%d-%H%M)
```

#### Configuration Fix (deployment-engineer)
```bash
# Update Helm values
vim deployments/helm/freightliner/values-production.yaml

# Deploy fix
helm upgrade freightliner ./deployments/helm/freightliner \
  -f ./deployments/helm/freightliner/values-production.yaml \
  -n freightliner-production

# Verify
kubectl rollout status deployment/freightliner -n freightliner-production
```

#### Infrastructure Fix (cloud-architect)
```bash
# Update Terraform
cd infrastructure/terraform/environments/production

# Apply changes
terraform plan
terraform apply

# Verify
kubectl get all -n freightliner-production
```

### 5. Verification (test-automator + devops-troubleshooter)

```bash
# Health checks
curl -k https://freightliner.production.company.com/health
curl -k https://freightliner.production.company.com/ready

# Smoke tests
./scripts/production-smoke-test.sh

# Monitor metrics
# - Error rate back to normal
# - Response times acceptable
# - No new errors in logs

# Load test if performance issue
/performance-test baseline
```

### 6. Post-Incident Review (All Agents)

Create incident report in `docs/incidents/incident-YYYYMMDD.md`:

```markdown
# Incident Report - [Date]

## Summary
Brief description of incident

## Timeline
- HH:MM - Incident detected
- HH:MM - Initial assessment complete
- HH:MM - Mitigation implemented
- HH:MM - Root cause identified
- HH:MM - Fix deployed
- HH:MM - Incident resolved

## Impact
- Users affected: X
- Duration: X minutes
- Services impacted: [list]

## Root Cause
Detailed explanation

## Resolution
What was done to fix

## Prevention
How to prevent in future

## Action Items
- [ ] Task 1
- [ ] Task 2
```

### 7. Follow-Up Actions

- Update monitoring and alerting
- Add tests to prevent regression
- Update runbooks
- Implement additional safeguards
- Share learnings with team

## Agent Responsibilities

### devops-troubleshooter
- Initial triage and assessment
- Log analysis
- Infrastructure troubleshooting
- Coordination

### performance-engineer
- Performance metrics analysis
- Bottleneck identification
- Resource optimization
- Load testing

### security-auditor
- Security incident assessment
- Vulnerability analysis
- Access control review
- Compliance checking

### golang-pro
- Code fix implementation
- Unit test creation
- Code review

### cloud-architect
- Infrastructure changes
- Scaling decisions
- Architecture review

### deployment-engineer
- Deployment operations
- Rollback execution
- CI/CD fixes
- Configuration management

## Communication Template

### Initial Notification
```
🚨 INCIDENT DETECTED - P[0/1/2/3]

Service: Freightliner
Status: Investigating
Impact: [description]
Started: [timestamp]
Next Update: [timestamp]
```

### Update
```
📊 INCIDENT UPDATE

Service: Freightliner
Status: [Investigating/Mitigating/Resolved]
Progress: [description]
ETA: [if known]
Next Update: [timestamp]
```

### Resolution
```
✅ INCIDENT RESOLVED

Service: Freightliner
Duration: [X minutes]
Root Cause: [brief description]
Resolution: [brief description]
Post-Mortem: [link when available]
```
