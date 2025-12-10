# Operations Runbook

Operational procedures for Freightliner container registry replication service.

## Quick Reference

### Emergency Commands

```bash
# Check health
curl http://localhost:8080/health

# View logs
docker logs freightliner --tail 100 -f
kubectl logs -l app=freightliner --tail=100 -f

# Restart service
docker restart freightliner
kubectl rollout restart deployment/freightliner

# Emergency rollback
kubectl rollout undo deployment/freightliner
```

### Service Endpoints

| Port | Endpoint | Purpose |
|------|----------|---------|
| 8080 | `/health` | Health check |
| 8080 | `/ready` | Readiness probe |
| 8080 | `/live` | Liveness probe |
| 2112 | `/metrics` | Prometheus metrics |

## Troubleshooting

### Service Not Starting

**Symptoms**: Container exits immediately, health checks fail

**Diagnosis**:
```bash
# Check container logs
docker logs freightliner 2>&1 | head -50

# Check for port conflicts
netstat -tlnp | grep -E '8080|2112'

# Verify configuration
docker exec freightliner env | grep -E 'PORT|LOG'
```

**Common Causes**:
- Port already in use: Change PORT environment variable
- Invalid configuration: Check config file syntax
- Missing credentials: Verify AWS/GCP auth setup

### Health Check Failures

**Symptoms**: Load balancer marks instance unhealthy

**Diagnosis**:
```bash
# Test health endpoints directly
curl -v http://localhost:8080/health
curl -v http://localhost:8080/ready

# Check system resources
docker stats freightliner
```

**Resolution**:
```bash
# Restart if unresponsive
docker restart freightliner

# Check resource limits
kubectl describe pod -l app=freightliner | grep -A5 Resources
```

### High Memory Usage

**Symptoms**: OOM kills, slow response times

**Diagnosis**:
```bash
# Check memory metrics
curl http://localhost:2112/metrics | grep memory

# View container stats
docker stats freightliner --no-stream
```

**Resolution**:
```bash
# Increase memory limit (Kubernetes)
kubectl patch deployment freightliner \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"freightliner","resources":{"limits":{"memory":"2Gi"}}}]}}}}'

# Restart to clear memory
docker restart freightliner
```

### Replication Failures

**Symptoms**: Images not syncing, error logs

**Diagnosis**:
```bash
# Check replication metrics
curl http://localhost:2112/metrics | grep replication

# View error logs
docker logs freightliner 2>&1 | grep -i error

# Test registry connectivity
docker exec freightliner curl -v https://123456789012.dkr.ecr.us-west-2.amazonaws.com/v2/
```

**Common Causes**:
- Expired credentials: Refresh AWS/GCP tokens
- Network issues: Check firewall rules
- Rate limiting: Add retry delays

### Authentication Errors

**AWS ECR**:
```bash
# Verify credentials
aws sts get-caller-identity

# Refresh token
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 123456789012.dkr.ecr.us-west-2.amazonaws.com
```

**Google GCR**:
```bash
# Verify credentials
gcloud auth application-default print-access-token

# Re-authenticate
gcloud auth application-default login
```

## Maintenance Procedures

### Rolling Update

```bash
# Kubernetes
kubectl set image deployment/freightliner \
  freightliner=ghcr.io/hemzaz/freightliner:v1.1.0

# Monitor rollout
kubectl rollout status deployment/freightliner

# Docker Compose
docker-compose pull
docker-compose up -d --no-deps freightliner
```

### Scaling

```bash
# Kubernetes manual scaling
kubectl scale deployment freightliner --replicas=5

# Enable HPA
kubectl autoscale deployment freightliner \
  --min=3 --max=10 --cpu-percent=70
```

### Log Management

```bash
# View recent logs
kubectl logs -l app=freightliner --tail=1000 --since=1h

# Export logs
kubectl logs -l app=freightliner > freightliner-$(date +%Y%m%d).log

# Clear old logs (Docker)
docker system prune --volumes
```

### Backup Checkpoints

```bash
# Export checkpoints
kubectl cp freightliner-pod:/data/checkpoints ./backup/checkpoints-$(date +%Y%m%d)

# Restore checkpoints
kubectl cp ./backup/checkpoints/ freightliner-pod:/data/checkpoints
```

## Monitoring

### Key Metrics to Watch

| Metric | Warning | Critical | Action |
|--------|---------|----------|--------|
| `up{job="freightliner"}` | - | == 0 | Restart service |
| HTTP 5xx rate | > 1% | > 5% | Check logs |
| Memory usage | > 70% | > 90% | Scale or restart |
| Request latency p99 | > 1s | > 5s | Check bottlenecks |

### Grafana Dashboard Queries

```promql
# Request rate
rate(freightliner_http_requests_total[5m])

# Error rate
rate(freightliner_http_requests_total{status=~"5.."}[5m])
  / rate(freightliner_http_requests_total[5m])

# Latency p95
histogram_quantile(0.95, rate(freightliner_http_request_duration_seconds_bucket[5m]))

# Replication throughput
rate(freightliner_replication_bytes_total[5m])
```

## Incident Response

### Severity Levels

| Level | Description | Response Time | Escalation |
|-------|-------------|---------------|------------|
| P1 | Service down | 15 min | Immediate |
| P2 | Degraded performance | 1 hour | Team lead |
| P3 | Non-critical issue | 4 hours | Standard |
| P4 | Minor issue | Next business day | Backlog |

### P1 Incident Procedure

1. **Assess** (0-5 min)
   ```bash
   curl http://freightliner/health
   kubectl get pods -l app=freightliner
   kubectl logs -l app=freightliner --tail=50
   ```

2. **Mitigate** (5-15 min)
   ```bash
   # Restart pods
   kubectl rollout restart deployment/freightliner

   # Or rollback
   kubectl rollout undo deployment/freightliner
   ```

3. **Communicate**
   - Update status page
   - Notify stakeholders

4. **Resolve**
   - Identify root cause
   - Apply fix
   - Verify recovery

5. **Document**
   - Write incident report
   - Update runbook if needed

### Post-Incident Checklist

- [ ] Incident timeline documented
- [ ] Root cause identified
- [ ] Fix deployed and verified
- [ ] Monitoring updated if needed
- [ ] Runbook updated if needed
- [ ] Post-mortem scheduled (for P1/P2)

## Configuration Reference

### Environment Variables

```bash
# Core
LOG_LEVEL=info          # debug, info, warn, error
PORT=8080               # HTTP port
METRICS_PORT=2112       # Prometheus port

# Authentication
API_KEY=secret          # API authentication
TLS_ENABLED=true        # Enable HTTPS
TLS_CERT_FILE=/path     # TLS certificate
TLS_KEY_FILE=/path      # TLS key

# AWS
AWS_REGION=us-west-2
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx

# GCP
GCP_PROJECT=project-id
GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json

# Workers
REPLICATE_WORKERS=4     # Parallel workers
```

### Useful Commands

```bash
# Debug mode
LOG_LEVEL=debug ./freightliner serve

# Dry run replication
./freightliner replicate --dry-run source dest

# Version info
./freightliner version

# Health check (container)
./freightliner health-check
```

## Contact

- **On-call**: Check PagerDuty rotation
- **Slack**: #freightliner-ops
- **Documentation**: See [DEPLOYMENT.md](DEPLOYMENT.md)
