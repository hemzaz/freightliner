# Freightliner SLO Runbooks

## Overview

This document contains runbooks for all Freightliner alerts, organized by component and severity.

## SLO Definitions

### Service Level Objectives

1. **Availability**: 99.9% successful replications (30-day window)
   - Error Budget: 43.2 minutes/month of failed replications
   - Measurement: Ratio of successful to total replication operations

2. **Latency**: 99% of replications complete in < 60 seconds
   - Measurement: P99 of replication_duration_seconds

3. **HTTP API**: 99.5% availability (5-minute window)
   - Error Budget: 0.5% requests can fail
   - Measurement: HTTP 2xx/3xx vs total requests

## Critical Alerts

### HighReplicationFailureRate

**Alert**: More than 10% of replications are failing

**Symptoms**:
- Replication failure rate > 10% for 5+ minutes
- Errors visible in logs
- Error metrics increasing

**Impact**:
- Container images not reaching destination registries
- Deployment pipelines may be blocked
- Error budget burning rapidly

**Diagnosis**:

1. Check current error rate:
```promql
sum by (source_registry, dest_registry) (
  rate(freightliner_replications_total{status="failed"}[5m])
) /
sum by (source_registry, dest_registry) (
  rate(freightliner_replications_total[5m])
)
```

2. Identify error types:
```bash
kubectl logs -n freightliner deployment/freightliner --tail=100 | grep -i error
```

3. Check specific registry pair:
```promql
freightliner_replication_errors_total{
  source_registry="SOURCE",
  dest_registry="DEST"
}
```

**Resolution**:

Common causes and fixes:

1. **Authentication Failures**
   - Check registry credentials:
   ```bash
   kubectl get secret registry-credentials -n freightliner -o yaml
   ```
   - Verify credentials haven't expired
   - Test manual login to registries

2. **Network Connectivity**
   - Test registry reachability:
   ```bash
   kubectl exec -n freightliner deployment/freightliner -- \
     curl -I https://registry.example.com/v2/
   ```
   - Check DNS resolution
   - Verify firewall rules

3. **Rate Limiting**
   - Check if hitting registry rate limits
   - Review error messages for "429 Too Many Requests"
   - Reduce worker pool size or add delays

4. **Registry Unavailability**
   - Check registry status pages
   - Verify upstream registry health
   - Consider temporary failover to alternate registry

**Prevention**:
- Implement credential rotation monitoring
- Add circuit breakers for failing registries
- Set up pre-expiry alerts for credentials
- Regular connectivity testing

---

### ReplicationStopped

**Alert**: No replication activity for 10 minutes

**Symptoms**:
- Zero replication operations recorded
- No metrics being updated
- Possible service crash or hang

**Impact**:
- Complete replication outage
- No new images being distributed
- Critical for deployment pipelines

**Diagnosis**:

1. Check service health:
```bash
kubectl get pods -n freightliner
kubectl describe pod <pod-name> -n freightliner
kubectl logs -n freightliner deployment/freightliner --tail=200
```

2. Verify job scheduler:
```bash
# Check if jobs are being scheduled
kubectl logs -n freightliner deployment/freightliner | grep "scheduled"
```

3. Check worker pool:
```promql
freightliner_worker_pool_active
freightliner_worker_pool_queued
freightliner_jobs_active
```

**Resolution**:

1. **Service Not Running**
   ```bash
   kubectl rollout restart deployment/freightliner -n freightliner
   ```

2. **Job Scheduler Issue**
   - Check scheduler configuration
   - Verify cron patterns
   - Review job definitions

3. **Database Connection Issue**
   - Check database connectivity
   - Verify connection pool settings
   - Review database logs

4. **Deadlock or Hang**
   - Get goroutine dump:
   ```bash
   kubectl exec -n freightliner deployment/freightliner -- \
     curl http://localhost:8080/debug/pprof/goroutine?debug=1
   ```
   - Analyze for blocked goroutines
   - Restart service if confirmed deadlock

**Prevention**:
- Implement health check endpoints
- Add deadlock detection
- Configure appropriate timeouts
- Regular load testing

---

### FreightlinerDown

**Alert**: Freightliner service is unreachable

**Symptoms**:
- Prometheus target showing as down
- Health check endpoint not responding
- Pod in CrashLoopBackOff or ImagePullBackOff

**Impact**:
- Complete service outage
- All replication stopped
- Immediate customer impact

**Diagnosis**:

1. Check pod status:
```bash
kubectl get pods -n freightliner -o wide
kubectl describe pod <pod-name> -n freightliner
```

2. Check recent events:
```bash
kubectl get events -n freightliner --sort-by='.lastTimestamp'
```

3. Review logs:
```bash
kubectl logs -n freightliner deployment/freightliner --previous
```

4. Check resource limits:
```promql
container_memory_usage_bytes{namespace="freightliner"}
container_cpu_usage_seconds_total{namespace="freightliner"}
```

**Resolution**:

1. **OOMKilled**
   - Increase memory limits
   - Check for memory leaks
   - Restart pod

2. **CrashLoopBackOff**
   - Check logs for panic/error
   - Verify configuration
   - Review recent deployments
   - Rollback if needed:
   ```bash
   kubectl rollout undo deployment/freightliner -n freightliner
   ```

3. **ImagePullBackOff**
   - Verify image exists
   - Check image pull secrets
   - Review registry credentials

4. **Node Issues**
   - Check node health:
   ```bash
   kubectl describe node <node-name>
   ```
   - Cordon and drain if necessary

**Prevention**:
- Set appropriate resource requests/limits
- Implement proper health checks
- Use pod disruption budgets
- Regular chaos testing

---

### HighHTTP5xxRate

**Alert**: High rate of HTTP 5xx errors

**Symptoms**:
- HTTP 500/502/503/504 errors > 1/sec
- API endpoints returning server errors
- Potential database or dependency issues

**Impact**:
- Degraded API functionality
- Failed replication triggers
- User-facing errors

**Diagnosis**:

1. Identify failing endpoints:
```promql
topk(10, sum by (path, method) (
  rate(freightliner_http_requests_total{status=~"5.."}[5m])
))
```

2. Check application logs:
```bash
kubectl logs -n freightliner deployment/freightliner | grep "HTTP 5"
```

3. Review error traces:
```bash
kubectl logs -n freightliner deployment/freightliner | grep -A 10 "panic\|fatal\|error"
```

4. Check dependencies:
   - Database connectivity
   - External registry APIs
   - Kubernetes API server

**Resolution**:

1. **Database Connection Issues**
   - Check connection pool exhaustion
   - Verify database health
   - Review slow query logs
   - Increase connection limits if needed

2. **Resource Exhaustion**
   - Check goroutine count
   - Review memory usage
   - Look for resource leaks
   - Scale up if necessary

3. **Code Errors**
   - Identify panic/error in logs
   - Check recent code changes
   - Deploy hotfix or rollback

4. **Dependency Failures**
   - Identify failing external service
   - Implement fallback/retry logic
   - Consider circuit breaker

**Prevention**:
- Comprehensive error handling
- Circuit breakers for external calls
- Connection pool monitoring
- Load testing before deployment

---

### HighMemoryUsage

**Alert**: Memory usage exceeds 4GB

**Symptoms**:
- Memory usage trending upward
- Potential memory leak
- Risk of OOMKill

**Impact**:
- Service instability
- Possible crashes
- Performance degradation

**Diagnosis**:

1. Check current memory usage:
```promql
freightliner_memory_usage_bytes / 1024 / 1024
```

2. Get memory profile:
```bash
kubectl exec -n freightliner deployment/freightliner -- \
  curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof -http=:8081 heap.prof
```

3. Check for goroutine leaks:
```bash
kubectl exec -n freightliner deployment/freightliner -- \
  curl http://localhost:8080/debug/pprof/goroutine?debug=1
```

4. Review metrics:
```promql
freightliner_goroutines_count
freightliner_worker_pool_queued
```

**Resolution**:

1. **Memory Leak**
   - Analyze heap profile for growing allocations
   - Identify leaking goroutines or objects
   - Deploy fix with proper cleanup
   - Restart service temporarily

2. **Legitimate High Usage**
   - Increase memory limits:
   ```bash
   kubectl set resources deployment freightliner \
     --limits=memory=8Gi -n freightliner
   ```
   - Scale horizontally if possible

3. **Large Jobs**
   - Reduce batch sizes
   - Implement streaming for large data
   - Add backpressure mechanisms

**Prevention**:
- Regular memory profiling
- Proper resource cleanup
- Memory limit monitoring
- Load testing with realistic data

---

## Warning Alerts

### HighReplicationDuration

**Alert**: P95 replication duration > 5 minutes

**Impact**: Slow replication affecting deployment speed

**Quick Actions**:
1. Check network latency between registries
2. Review image sizes being replicated
3. Check registry performance metrics
4. Consider increasing parallelism

**Investigation**:
```promql
histogram_quantile(0.95,
  sum by (source_registry, dest_registry, le) (
    rate(freightliner_replication_duration_seconds_bucket[10m])
  )
)
```

---

### WorkerPoolSaturated

**Alert**: Worker pool utilization > 90%

**Impact**: Job queuing and processing delays

**Quick Actions**:
1. Check queue depth
2. Review job execution times
3. Scale up worker pool:
```bash
kubectl scale deployment freightliner --replicas=5 -n freightliner
```

---

### HighHTTPLatency

**Alert**: P95 HTTP latency > 5 seconds

**Impact**: Slow API responses

**Quick Actions**:
1. Identify slow endpoints
2. Check for slow database queries
3. Review application profiling
4. Consider caching improvements

---

### HighAuthFailureRate

**Alert**: High rate of authentication failures

**Impact**: Potential security issue or credential problems

**Quick Actions**:
1. Check for credential expiry
2. Review for brute force attacks
3. Verify registry authentication configuration
4. Check IP allowlists

---

## Error Budget Management

### Calculating Error Budget Burn

```promql
# 1-hour burn rate (should be ~1x)
(
  sum(rate(freightliner_replications_total{status="failed"}[1h]))
  /
  sum(rate(freightliner_replications_total[1h]))
) / 0.001

# 5-minute burn rate (for fast detection)
(
  sum(rate(freightliner_replications_total{status="failed"}[5m]))
  /
  sum(rate(freightliner_replications_total[5m]))
) / 0.001
```

### Burn Rate Thresholds

| Window | Threshold | Action |
|--------|-----------|--------|
| 5m     | > 14x     | Page immediately |
| 1h     | > 6x      | Page during business hours |
| 6h     | > 1x      | Create ticket |

### When Error Budget is Exhausted

1. **Immediate Actions**:
   - Freeze non-critical deployments
   - Focus on reliability fixes
   - Increase monitoring frequency

2. **Root Cause Analysis**:
   - Identify failure patterns
   - Review recent changes
   - Update runbooks

3. **Recovery**:
   - Implement fixes
   - Verify improvements
   - Resume normal operations when budget recovered

## Escalation Paths

### Severity Levels

**P0 - Critical** (< 15 min response):
- FreightlinerDown
- HighReplicationFailureRate
- ApplicationPanics

**P1 - High** (< 1 hour response):
- ReplicationStopped
- HighHTTP5xxRate
- HighMemoryUsage

**P2 - Medium** (< 4 hours response):
- HighReplicationDuration
- WorkerPoolSaturated
- HighHTTPLatency

**P3 - Low** (< 24 hours response):
- HighAuthFailureRate
- High queue depth warnings

### On-Call Contacts

1. **Primary**: Freightliner team on-call (PagerDuty)
2. **Secondary**: Platform engineering team
3. **Escalation**: Engineering manager

## Additional Resources

- Metrics Dashboard: http://grafana.example.com/d/freightliner
- Logs: Kibana/CloudWatch logs
- Service Status: http://status.example.com
- Incident Response: Internal wiki
