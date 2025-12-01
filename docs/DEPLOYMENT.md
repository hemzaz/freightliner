# Deployment Guide

Complete deployment automation for Freightliner container registry replication tool.

## Quick Start

### Deploy to Development
```bash
./scripts/deployment/setup-env.sh dev
./scripts/deployment/deploy.sh dev latest
./scripts/deployment/health-check.sh dev
```

### Deploy to Production
```bash
./scripts/deployment/setup-env.sh production
./scripts/deployment/deploy.sh production v1.2.3
./scripts/deployment/health-check.sh production
```

## Prerequisites

### Required Tools
- Docker (>= 20.10)
- kubectl (>= 1.25)
- Helm (>= 3.10)
- Trivy (for security scanning)

### Installation
```bash
# Docker
curl -fsSL https://get.docker.com | sh

# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl && sudo mv kubectl /usr/local/bin/

# Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Trivy
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
```

## Docker Setup

### Build Optimized Image
```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f Dockerfile.optimized \
  -t ghcr.io/hemzaz/freightliner:v1.0.0 \
  --push .
```

### Local Development
```bash
# Start all services with hot reload
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f freightliner-dev

# Stop services
docker-compose -f docker-compose.dev.yml down
```

## Kubernetes Deployment

### Quick Deploy
```bash
# Create namespace
kubectl create namespace freightliner-production

# Apply manifests
kubectl apply -f deployments/kubernetes/serviceaccount.yaml -n freightliner-production
kubectl apply -f deployments/kubernetes/configmap-new.yaml -n freightliner-production
kubectl apply -f deployments/kubernetes/deployment-new.yaml -n freightliner-production
kubectl apply -f deployments/kubernetes/service-new.yaml -n freightliner-production
kubectl apply -f deployments/kubernetes/hpa.yaml -n freightliner-production

# Verify
kubectl rollout status deployment/freightliner -n freightliner-production
kubectl get pods -n freightliner-production
```

## Rollback

### Automatic Rollback
Triggered when health checks fail after deployment.

### Manual Rollback
```bash
# Rollback to previous version
./scripts/deployment/rollback.sh production

# Using kubectl
kubectl rollout undo deployment/freightliner -n freightliner-production
```

## Monitoring

### Health Checks
```bash
./scripts/deployment/health-check.sh production
```

### View Logs
```bash
kubectl logs -f deployment/freightliner -n freightliner-production
```

### Metrics
- Prometheus: http://service:2112/metrics
- Grafana: http://grafana.example.com

## Troubleshooting

### Common Issues

**Image Pull Errors**
```bash
docker pull ghcr.io/hemzaz/freightliner:v1.2.3
kubectl describe pod <pod-name> -n freightliner-production
```

**CrashLoopBackOff**
```bash
kubectl logs <pod-name> -n freightliner-production --previous
kubectl describe pod <pod-name> -n freightliner-production
```

**Service Not Accessible**
```bash
kubectl get endpoints freightliner -n freightliner-production
kubectl describe service freightliner -n freightliner-production
```

## CI/CD Integration

GitHub Actions workflow automatically:
- Builds Docker image on push
- Deploys to dev on merge to main
- Requires approval for staging/production
- Runs health checks and rolls back on failure

Trigger manual deployment:
```bash
gh workflow run deploy.yml \
  -f environment=staging \
  -f version=v1.2.3
```

## Security Best Practices

1. Use specific image tags (never `latest` in production)
2. Run as non-root user (UID 1001)
3. Enable Pod Security Standards
4. Use Network Policies
5. Scan images for vulnerabilities
6. Rotate secrets regularly
7. Use read-only root filesystem

## Support

- Documentation: See full DEPLOYMENT.md
- GitHub Issues: https://github.com/hemzaz/freightliner/issues
- Contact: DevOps team
