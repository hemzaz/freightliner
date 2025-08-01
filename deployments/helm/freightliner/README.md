# Freightliner Helm Chart

A Helm chart for deploying the Freightliner Container Registry Replication application on Kubernetes.

## Overview

This Helm chart deploys a production-ready Freightliner application with:
- High availability deployment with configurable replicas
- Auto-scaling based on CPU/memory metrics
- TLS-enabled ingress with certificate management
- Comprehensive monitoring and observability
- Security hardening with network policies and security contexts
- Multi-environment support (development, staging, production)

## Prerequisites

- Kubernetes 1.20+
- Helm 3.2.0+
- (Optional) Prometheus Operator for ServiceMonitor CRDs
- (Optional) cert-manager for automatic TLS certificate management
- (Optional) NGINX Ingress Controller

## Installation

### Quick Start

```bash
# Add the repository (if using a Helm repository)
helm repo add freightliner https://charts.company.com/freightliner
helm repo update

# Install with default values
helm install freightliner freightliner/freightliner

# Or install from local chart
helm install freightliner ./deployments/helm/freightliner
```

### Environment-Specific Deployments

#### Production
```bash
helm install freightliner-prod ./deployments/helm/freightliner \
  --namespace freightliner-prod \
  --create-namespace \
  --values ./deployments/helm/freightliner/values-production.yaml \
  --set config.aws.region=us-west-2 \
  --set config.gcp.projectId=company-prod-123456
```

#### Staging
```bash
helm install freightliner-staging ./deployments/helm/freightliner \
  --namespace freightliner-staging \
  --create-namespace \
  --values ./deployments/helm/freightliner/values-staging.yaml \
  --set config.aws.region=us-west-2 \
  --set config.gcp.projectId=company-staging-123456
```

#### Development
```bash
helm install freightliner-dev ./deployments/helm/freightliner \
  --namespace freightliner-dev \
  --create-namespace \
  --set replicaCount=1 \
  --set resources.requests.cpu=100m \
  --set resources.requests.memory=256Mi \
  --set config.logLevel=debug
```

## Configuration

### Core Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `3` |
| `image.registry` | Image registry | `docker.io` |
| `image.repository` | Image repository | `freightliner/app` |
| `image.tag` | Image tag | `1.0.0` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### Service Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Service type | `LoadBalancer` |
| `service.port` | Service port | `443` |
| `service.targetPort` | Container port | `8080` |

### Ingress Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `true` |
| `ingress.className` | Ingress class | `nginx` |
| `ingress.hosts` | Ingress hosts | See values.yaml |
| `ingress.tls` | TLS configuration | See values.yaml |

### Application Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.logLevel` | Log level | `info` |
| `config.port` | Application port | `8080` |
| `config.aws.region` | AWS region | `us-west-2` |
| `config.gcp.projectId` | GCP project ID | `""` |
| `config.workerPoolSize` | Worker pool size | `10` |
| `config.maxConcurrentReplications` | Max concurrent replications | `5` |

### Resource Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `2000m` |
| `resources.limits.memory` | Memory limit | `4Gi` |
| `resources.requests.cpu` | CPU request | `500m` |
| `resources.requests.memory` | Memory request | `1Gi` |

### Auto-scaling Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `autoscaling.enabled` | Enable HPA | `true` |
| `autoscaling.minReplicas` | Minimum replicas | `3` |
| `autoscaling.maxReplicas` | Maximum replicas | `10` |
| `autoscaling.targetCPUUtilizationPercentage` | CPU target | `70` |
| `autoscaling.targetMemoryUtilizationPercentage` | Memory target | `80` |

### Security Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `securityContext.runAsNonRoot` | Run as non-root | `true` |
| `securityContext.runAsUser` | User ID | `1000` |
| `securityContext.readOnlyRootFilesystem` | Read-only root | `true` |
| `networkPolicy.enabled` | Enable network policies | `true` |

### Monitoring Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.metricsEnabled` | Enable metrics | `true` |
| `config.metricsPort` | Metrics port | `2112` |
| `monitoring.serviceMonitor.enabled` | Enable ServiceMonitor | `true` |

## Secrets Management

The chart expects secrets to be provided externally. Create secrets before deployment:

### AWS Credentials
```bash
kubectl create secret generic freightliner-secrets \
  --from-literal=aws-access-key-id=YOUR_ACCESS_KEY \
  --from-literal=aws-secret-access-key=YOUR_SECRET_KEY \
  --namespace freightliner
```

### GCP Service Account
```bash
kubectl create secret generic freightliner-secrets \
  --from-file=gcp-service-account-key=path/to/service-account.json \
  --namespace freightliner
```

### Using External Secret Management

For production deployments, consider using:
- [External Secrets Operator](https://external-secrets.io/)
- [Sealed Secrets](https://sealed-secrets.netlify.app/)
- [Vault](https://www.vaultproject.io/)

Example with External Secrets Operator:
```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-west-2
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: freightliner-secrets
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: freightliner-secrets
    creationPolicy: Owner
  data:
  - secretKey: aws-access-key-id
    remoteRef:
      key: freightliner/aws-credentials
      property: access_key_id
  - secretKey: aws-secret-access-key
    remoteRef:
      key: freightliner/aws-credentials
      property: secret_access_key
```

## Monitoring and Observability

### Prometheus Metrics

The application exposes Prometheus metrics at `/metrics` endpoint. The chart includes:
- ServiceMonitor for Prometheus Operator
- Proper metric port configuration
- Labels for metric identification

### Health Checks

The application provides health check endpoints:
- `/health` - Application health
- `/ready` - Readiness probe

### Logging

Configure log level through `config.logLevel`:
- `debug` - Detailed debugging information
- `info` - General operational information
- `warn` - Warning messages only
- `error` - Error messages only

## Networking

### Network Policies

When `networkPolicy.enabled` is true, the chart creates network policies that:
- Allow ingress from ingress controller namespace
- Allow ingress from monitoring namespace
- Allow egress for DNS resolution
- Allow egress for HTTPS (container registries)

### TLS Configuration

The chart supports TLS termination at:
1. **Ingress Level** (recommended)
   - Uses cert-manager for automatic certificate provisioning
   - Configurable through `ingress.tls`

2. **Load Balancer Level**
   - For cloud provider load balancers
   - Configure through service annotations

## Persistence

Optional persistent storage for:
- Checkpoint data
- Temporary files
- Application state

Configure through `persistence` values:
```yaml
persistence:
  enabled: true
  storageClass: "fast-ssd"
  size: 50Gi
  accessMode: ReadWriteOnce
```

## High Availability

The chart provides high availability through:

### Pod Anti-Affinity
Spreads pods across different nodes to avoid single points of failure.

### Pod Disruption Budget
Ensures minimum number of pods remain available during:
- Cluster maintenance
- Node upgrades
- Voluntary pod evictions

### Health Checks
- Liveness probes detect and restart unhealthy pods
- Readiness probes prevent traffic to non-ready pods

### Resource Management
- Resource requests ensure proper scheduling
- Resource limits prevent resource starvation

## Upgrading

### Rolling Updates
```bash
# Update image tag
helm upgrade freightliner ./deployments/helm/freightliner \
  --set image.tag=1.1.0

# Update configuration
helm upgrade freightliner ./deployments/helm/freightliner \
  --values updated-values.yaml
```

### Blue-Green Deployments
For zero-downtime updates:
```bash
# Deploy new version alongside current
helm install freightliner-new ./deployments/helm/freightliner \
  --set fullnameOverride=freightliner-new \
  --set image.tag=1.1.0

# Switch traffic (update ingress or service)
# Remove old deployment
helm uninstall freightliner-old
```

## Troubleshooting

### Common Issues

1. **Pods not starting**
   ```bash
   kubectl describe pod <pod-name>
   kubectl logs <pod-name>
   ```

2. **Service not accessible**
   ```bash
   kubectl get svc
   kubectl describe ingress
   ```

3. **Certificate issues**
   ```bash
   kubectl describe certificate
   kubectl describe certificaterequest
   ```

4. **Resource issues**
   ```bash
   kubectl top nodes
   kubectl top pods
   ```

### Debug Mode

Enable debug logging:
```bash
helm upgrade freightliner ./deployments/helm/freightliner \
  --set config.logLevel=debug
```

### Health Checks

Test application health:
```bash
kubectl port-forward svc/freightliner 8080:80
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

## Contributing

1. Make changes to the chart
2. Update version in `Chart.yaml`
3. Test with different value configurations
4. Update documentation

### Testing

```bash
# Lint the chart
helm lint ./deployments/helm/freightliner

# Test template rendering
helm template freightliner ./deployments/helm/freightliner

# Dry run installation
helm install freightliner ./deployments/helm/freightliner --dry-run
```

## Support

For support and questions:
- Create an issue in the repository
- Contact the Platform Team at platform@company.com
- Check the troubleshooting guide above