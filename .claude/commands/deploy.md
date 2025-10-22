# Deploy Command

Deploy Freightliner to Kubernetes using Helm charts with proper validation and rollback capabilities.

## What This Command Does

1. Validates deployment prerequisites
2. Runs pre-deployment checks
3. Deploys using Helm with environment-specific values
4. Performs post-deployment validation
5. Sets up monitoring and alerting
6. Provides rollback instructions

## Usage

```bash
/deploy [environment] [version]
```

## Environments

- `dev` - Development environment
- `staging` - Staging environment
- `production` - Production environment

## Example

```bash
/deploy production v1.2.0
```

## Pre-Deployment Checklist

### 1. Prerequisites Validation
```bash
# Verify kubectl context
kubectl config current-context

# Verify namespace exists
kubectl get namespace freightliner-production

# Verify Helm is installed
helm version

# Check image availability
docker pull ghcr.io/hemzaz/freightliner:v1.2.0
```

### 2. Configuration Review
- [ ] Secrets are created and encrypted
- [ ] ConfigMaps are up to date
- [ ] Resource limits are appropriate
- [ ] Ingress TLS certificates are valid
- [ ] Service accounts have proper permissions

### 3. Backup Current State
```bash
# Backup current deployment
helm get values freightliner -n freightliner-production > backup-values.yaml
kubectl get all -n freightliner-production -o yaml > backup-resources.yaml
```

## Deployment Steps

### 1. Deploy with Helm
```bash
# Development
helm upgrade --install freightliner \
  ./deployments/helm/freightliner \
  -f ./deployments/helm/freightliner/values-dev.yaml \
  --namespace freightliner-dev \
  --create-namespace \
  --wait \
  --timeout 10m

# Staging
helm upgrade --install freightliner \
  ./deployments/helm/freightliner \
  -f ./deployments/helm/freightliner/values-staging.yaml \
  --namespace freightliner-staging \
  --create-namespace \
  --wait \
  --timeout 10m

# Production
helm upgrade --install freightliner \
  ./deployments/helm/freightliner \
  -f ./deployments/helm/freightliner/values-production.yaml \
  --namespace freightliner-production \
  --create-namespace \
  --wait \
  --timeout 10m \
  --set image.tag=v1.2.0
```

### 2. Verify Deployment
```bash
# Check rollout status
kubectl rollout status deployment/freightliner -n freightliner-production

# Verify pods are running
kubectl get pods -n freightliner-production

# Check pod logs
kubectl logs -f deployment/freightliner -n freightliner-production

# Verify services
kubectl get svc -n freightliner-production
```

### 3. Health Checks
```bash
# Get service endpoint
export ENDPOINT=$(kubectl get svc freightliner -n freightliner-production -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

# Health check
curl -k https://$ENDPOINT/health

# Readiness check
curl -k https://$ENDPOINT/ready

# Liveness check
curl -k https://$ENDPOINT/live

# Metrics endpoint
curl -k https://$ENDPOINT/metrics
```

### 4. Smoke Tests
```bash
# Run production smoke test
./scripts/production-smoke-test.sh $ENDPOINT

# Test replication functionality
./bin/freightliner replicate ecr/test-repo gcr/test-repo --dry-run
```

### 5. Monitoring Setup
```bash
# Verify Prometheus ServiceMonitor
kubectl get servicemonitor freightliner -n freightliner-production

# Check metrics in Prometheus
# Open Prometheus UI and verify metrics are being scraped

# Verify Grafana dashboards
# Import dashboard from deployments/monitoring/grafana-dashboard.json
```

## Post-Deployment Validation

### 1. Functional Tests
- [ ] Health endpoints responding correctly
- [ ] Metrics being collected
- [ ] Authentication working
- [ ] Replication operations functional

### 2. Performance Validation
- [ ] Response times within SLA (<200ms)
- [ ] Throughput meets targets (>125 MB/s)
- [ ] Resource usage within limits
- [ ] No memory leaks

### 3. Security Validation
- [ ] TLS certificates valid
- [ ] Network policies enforced
- [ ] RBAC permissions correct
- [ ] Secrets properly mounted

## Rollback Procedure

If deployment fails or issues are detected:

```bash
# List Helm releases
helm list -n freightliner-production

# Rollback to previous revision
helm rollback freightliner -n freightliner-production

# Or rollback to specific revision
helm rollback freightliner 5 -n freightliner-production

# Verify rollback
kubectl get pods -n freightliner-production
kubectl rollout status deployment/freightliner -n freightliner-production
```

## Monitoring & Alerts

### Key Metrics to Monitor
- Pod restart count
- HTTP error rates
- Replication success/failure rates
- Memory and CPU usage
- Network throughput

### Alert Conditions
- Pod crash loop backoff
- High error rate (>5%)
- High response time (>500ms)
- Memory usage >80%
- Replication failures

## Documentation

Update deployment documentation:
- Record deployment time and version
- Document any issues encountered
- Update runbooks if new procedures discovered
- Note any configuration changes made
