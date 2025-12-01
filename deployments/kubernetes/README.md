# Kubernetes Deployments

This directory contains production-ready Kubernetes manifests for deploying Freightliner container registry replication service.

## Directory Structure

```
kubernetes/
├── base/                           # Base Kustomize configuration
│   ├── kustomization.yaml         # Base Kustomize file
│   ├── namespace.yaml             # Namespace definition
│   ├── serviceaccount.yaml        # Service account
│   ├── configmap.yaml             # Application configuration
│   ├── secret.yaml                # Secrets (templates only)
│   ├── deployment.yaml            # Deployment manifest
│   ├── service.yaml               # Service definitions
│   ├── hpa.yaml                   # Horizontal Pod Autoscaler
│   ├── pdb.yaml                   # Pod Disruption Budget
│   ├── networkpolicy.yaml         # Network policies
│   ├── config.yaml                # Base configuration file
│   └── secrets.env                # Secret templates
│
└── overlays/                       # Environment-specific overlays
    ├── dev/                        # Development environment
    │   ├── kustomization.yaml
    │   ├── deployment-patch.yaml
    │   ├── service-patch.yaml
    │   ├── hpa-patch.yaml
    │   ├── config.yaml
    │   └── secrets.env
    │
    ├── staging/                    # Staging environment
    │   ├── kustomization.yaml
    │   ├── deployment-patch.yaml
    │   ├── hpa-patch.yaml
    │   ├── pdb-patch.yaml
    │   ├── ingress.yaml
    │   ├── config.yaml
    │   └── secrets.env
    │
    └── prod/                       # Production environment
        ├── kustomization.yaml
        ├── deployment-patch.yaml
        ├── service-patch.yaml
        ├── hpa-patch.yaml
        ├── pdb-patch.yaml
        ├── networkpolicy-patch.yaml
        ├── ingress.yaml
        ├── servicemonitor.yaml
        ├── vpa.yaml
        ├── config.yaml
        └── secrets.env
```

## Quick Start

### Deploy to Development

```bash
kubectl apply -k overlays/dev
```

### Deploy to Staging

```bash
# Create secrets first
kubectl create secret generic freightliner-secrets-staging \
  --from-literal=api-key='your-staging-api-key' \
  --namespace=freightliner-staging

# Deploy
kubectl apply -k overlays/staging
```

### Deploy to Production

```bash
# IMPORTANT: Use proper secret management (External Secrets, Vault, etc.)
kubectl create secret generic freightliner-secrets-prod \
  --from-literal=api-key='your-production-api-key' \
  --namespace=freightliner-prod

# Deploy
kubectl apply -k overlays/prod
```

## Environment Comparison

| Feature | Dev | Staging | Production |
|---------|-----|---------|------------|
| Replicas (min-max) | 1-3 | 2-5 | 5-20 |
| CPU Request | 250m | 500m | 1000m |
| CPU Limit | 1000m | 2000m | 4000m |
| Memory Request | 512Mi | 1Gi | 2Gi |
| Memory Limit | 2Gi | 4Gi | 8Gi |
| Service Type | NodePort | ClusterIP | LoadBalancer |
| Ingress | ❌ | ✅ | ✅ |
| TLS | ❌ | ✅ | ✅ |
| API Auth | ❌ | ✅ | ✅ |
| Monitoring | Basic | Standard | Advanced |
| NetworkPolicy | Permissive | Standard | Strict |
| PDB | minAvailable: 0 | minAvailable: 1 | minAvailable: 3 |
| VPA | ❌ | ❌ | ✅ |

## Configuration Management

### Base Configuration

Base configuration is defined in `base/config.yaml` and includes:
- Server settings (port, timeouts, health checks)
- Logging configuration
- Worker pool settings
- Replication settings
- Metrics configuration
- Registry settings

### Environment Overrides

Each environment overlay can override base settings:

**Dev**: Debug logging, fewer workers, disabled auth
**Staging**: Info logging, moderate workers, staging endpoints
**Production**: Warn logging, max workers, production endpoints

### Secret Management

**CRITICAL**: Never commit actual secrets to version control!

Recommended approaches:
1. **External Secrets Operator** (Recommended for AWS/GCP)
2. **Sealed Secrets** (For GitOps workflows)
3. **HashiCorp Vault** (For multi-cloud)
4. **Manual kubectl create secret** (Simple deployments)

See [Security Guide](../../docs/kubernetes-deployment-guide.md#secret-management) for details.

## Deployment Methods

### Method 1: Kustomize (Recommended)

```bash
# Preview changes
kubectl kustomize overlays/prod

# Apply directly
kubectl apply -k overlays/prod

# With server-side apply
kubectl apply -k overlays/prod --server-side

# Diff before applying (requires kubectl-diff plugin)
kubectl diff -k overlays/prod
```

### Method 2: Generate and Apply

```bash
# Generate manifests
kubectl kustomize overlays/prod > prod-manifests.yaml

# Review generated manifests
cat prod-manifests.yaml

# Apply
kubectl apply -f prod-manifests.yaml
```

### Method 3: CI/CD Pipeline

Example GitLab CI pipeline:

```yaml
deploy:production:
  stage: deploy
  script:
    - kubectl apply -k overlays/prod
    - kubectl rollout status deployment/freightliner-prod -n freightliner-prod
  only:
    - main
  environment:
    name: production
    url: https://freightliner-prod.company.com
```

## Health Checks

The application provides three health endpoints:

1. **Liveness** (`/live`): Is the application alive?
2. **Readiness** (`/ready`): Can the application serve traffic?
3. **Startup** (`/health`): Has the application started?

Configuration:

```yaml
livenessProbe:
  httpGet:
    path: /live
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 30
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 15
  periodSeconds: 10
  failureThreshold: 3

startupProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  failureThreshold: 30
```

## Auto-Scaling

### Horizontal Pod Autoscaler (HPA)

Automatically scales pods based on CPU and memory utilization:

**Dev**: 1-3 replicas, 80% CPU threshold
**Staging**: 2-5 replicas, 75% CPU threshold
**Production**: 5-20 replicas, 60% CPU threshold

View HPA status:
```bash
kubectl get hpa -n freightliner-prod
kubectl describe hpa freightliner-prod -n freightliner-prod
```

### Vertical Pod Autoscaler (VPA)

Production only: Automatically adjusts resource requests/limits

View VPA recommendations:
```bash
kubectl get vpa -n freightliner-prod
kubectl describe vpa freightliner-prod -n freightliner-prod
```

## High Availability

### Pod Disruption Budget (PDB)

Ensures minimum availability during voluntary disruptions:

**Dev**: No PDB (single replica)
**Staging**: minAvailable: 1
**Production**: minAvailable: 3

### Pod Anti-Affinity

Ensures pods are spread across different nodes:

**Dev/Staging**: Preferred (soft)
**Production**: Required (hard)

### Multi-Zone Deployment

For production, ensure:
- Cluster spans multiple availability zones
- Node pools in each zone
- PDB configured for zone-level failures

## Security

### Pod Security Standards

All deployments enforce **Restricted** Pod Security Standards:
- Non-root user (UID 1000)
- Read-only root filesystem
- No privilege escalation
- Dropped capabilities (ALL)
- seccomp profile (RuntimeDefault)

### Network Policies

Control ingress and egress traffic:

**Ingress**: Only from ingress-nginx and monitoring namespaces
**Egress**: DNS (53), HTTPS (443) for registry access

### RBAC

Minimal permissions:
- Service account with no token auto-mount
- Role for reading ConfigMaps only
- No cluster-level permissions

## Monitoring

### Metrics

Prometheus metrics exposed at `/metrics` on port 8080:
- `freightliner_replication_total`
- `freightliner_replication_duration_seconds`
- `freightliner_replication_errors_total`
- `freightliner_worker_pool_*`
- `http_request_*`

### ServiceMonitor

Production includes ServiceMonitor for Prometheus Operator:

```bash
kubectl get servicemonitor freightliner-prod -n freightliner-prod
```

### Logging

Structured JSON logs to stdout:
- `level`: Log level (debug, info, warn, error)
- `timestamp`: ISO 8601 timestamp
- `message`: Log message
- `context`: Additional structured data

## Troubleshooting

### Common Issues

**Pods not starting:**
```bash
kubectl describe pod <pod-name> -n freightliner-prod
kubectl logs <pod-name> -n freightliner-prod
```

**Health checks failing:**
```bash
kubectl get events -n freightliner-prod
kubectl logs deployment/freightliner-prod -n freightliner-prod
```

**High memory usage:**
```bash
kubectl top pods -n freightliner-prod
kubectl describe vpa freightliner-prod -n freightliner-prod
```

### Debugging Commands

```bash
# View all resources
kubectl get all -n freightliner-prod

# Stream logs
kubectl logs -f deployment/freightliner-prod -n freightliner-prod

# Execute shell in pod
kubectl exec -it <pod-name> -n freightliner-prod -- /bin/sh

# Port forward for local access
kubectl port-forward svc/freightliner-prod 8080:80 -n freightliner-prod
```

## Validation

Before deploying to production:

1. **Test in staging**: Deploy and validate in staging first
2. **Run validation checklist**: See [Validation Checklist](../../docs/kubernetes-validation-checklist.md)
3. **Load test**: Validate performance under load
4. **Disaster recovery**: Test backup and restore
5. **Document changes**: Update runbooks and documentation

## Upgrading

### Zero-Downtime Updates

```bash
# Update image tag
cd overlays/prod
kustomize edit set image freightliner=gcr.io/company-prod/freightliner:v1.1.0

# Apply update
kubectl apply -k .

# Monitor rollout
kubectl rollout status deployment/freightliner-prod -n freightliner-prod

# Verify health
kubectl get pods -n freightliner-prod
```

### Rollback

```bash
# View history
kubectl rollout history deployment/freightliner-prod -n freightliner-prod

# Rollback to previous version
kubectl rollout undo deployment/freightliner-prod -n freightliner-prod

# Rollback to specific revision
kubectl rollout undo deployment/freightliner-prod --to-revision=2 -n freightliner-prod
```

## Documentation

- [Quick Start Guide](../../docs/kubernetes-quick-start.md)
- [Deployment Guide](../../docs/kubernetes-deployment-guide.md)
- [Validation Checklist](../../docs/kubernetes-validation-checklist.md)
- [Application README](../../README.md)

## Support

- GitHub Issues: https://github.com/company/freightliner/issues
- Slack: #freightliner-support
- Email: platform@company.com
