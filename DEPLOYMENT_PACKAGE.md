# Production Deployment Package
**Freightliner Container Registry Replication System**

**Version:** 1.0.0  
**Release Date:** August 3, 2025  
**Deployment Status:** ✅ **READY FOR PRODUCTION**

## Package Contents

This deployment package contains all necessary artifacts, configurations, and procedures for production deployment of the Freightliner container registry replication system.

### 📦 Package Structure

```
freightliner-production-v1.0.0/
├── artifacts/
│   ├── container-images/
│   │   ├── freightliner-1.0.0.tar
│   │   ├── freightliner-1.0.0-secure.tar
│   │   └── image-manifests.yaml
│   └── binaries/
│       ├── freightliner-linux-amd64
│       ├── freightliner-linux-arm64
│       └── freightliner-darwin-amd64
├── kubernetes/
│   ├── base/
│   │   ├── namespace.yaml
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── configmap.yaml
│   │   ├── secrets.yaml.template
│   │   ├── ingress.yaml
│   │   └── networkpolicy.yaml
│   └── overlays/
│       ├── production/
│       │   ├── kustomization.yaml
│       │   └── production-overrides.yaml
│       └── staging/
│           ├── kustomization.yaml
│           └── staging-overrides.yaml
├── helm/
│   ├── freightliner/
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   ├── values-production.yaml
│   │   ├── values-staging.yaml
│   │   └── templates/
│   └── dependencies/
│       ├── prometheus-stack/
│       ├── ingress-nginx/
│       └── cert-manager/
├── infrastructure/
│   ├── terraform/
│   │   ├── environments/
│   │   │   ├── production/
│   │   │   └── staging/
│   │   └── modules/
│   │       ├── monitoring/
│   │       ├── security/
│   │       └── networking/
│   └── scripts/
│       ├── deploy-infrastructure.sh
│       ├── configure-monitoring.sh
│       └── setup-security.sh
├── scripts/
│   ├── deploy.sh
│   ├── rollback.sh
│   ├── health-check.sh
│   ├── smoke-test.sh
│   └── troubleshoot.sh
├── documentation/
│   ├── DEPLOYMENT_GUIDE.md
│   ├── TROUBLESHOOTING.md
│   ├── SECURITY_GUIDE.md
│   ├── MONITORING_GUIDE.md
│   └── OPERATIONS_RUNBOOK.md
└── tests/
    ├── integration/
    ├── performance/
    ├── security/
    └── smoke-tests/
```

## 🚀 Quick Deployment Guide

### Prerequisites Checklist

Before deploying, ensure the following prerequisites are met:

#### Infrastructure Requirements
- [ ] Kubernetes cluster v1.24+ with at least 3 nodes
- [ ] Minimum 8 CPU cores and 16GB RAM available
- [ ] Persistent storage class configured (minimum 100GB)
- [ ] Load balancer or ingress controller installed
- [ ] DNS configuration for application domains

#### Access Requirements
- [ ] kubectl configured with cluster admin access
- [ ] Helm 3.8+ installed and configured
- [ ] Docker registry access for image pulls
- [ ] AWS/GCP credentials configured (if using cloud registries)

#### Security Requirements
- [ ] TLS certificates available (or cert-manager configured)
- [ ] Service account roles and permissions configured
- [ ] Network policies and security groups configured
- [ ] Secrets management system available (e.g., Vault, K8s secrets)

### 1. Quick Deployment (Production)

```bash
#!/bin/bash
# Quick production deployment script

set -euo pipefail

echo "🚀 Starting Freightliner production deployment..."

# Deploy namespace and RBAC
kubectl apply -f kubernetes/base/namespace.yaml

# Deploy secrets (update secrets.yaml with actual values)
kubectl apply -f kubernetes/base/secrets.yaml

# Deploy with Helm
helm upgrade --install freightliner \
  ./helm/freightliner \
  -f ./helm/freightliner/values-production.yaml \
  --namespace freightliner-production \
  --create-namespace \
  --wait --timeout=10m

# Verify deployment
./scripts/health-check.sh production

echo "✅ Production deployment completed successfully!"
```

### 2. Step-by-Step Deployment

#### Step 1: Prepare Environment
```bash
# Create namespace
kubectl create namespace freightliner-production

# Label namespace for monitoring
kubectl label namespace freightliner-production \
  monitoring=enabled \
  security-policy=strict
```

#### Step 2: Configure Secrets
```bash
# Copy secrets template
cp kubernetes/base/secrets.yaml.template kubernetes/base/secrets.yaml

# Edit secrets with actual values
# Required secrets:
# - registry-credentials: Container registry access
# - api-keys: Application API keys
# - tls-certificates: HTTPS certificates
# - encryption-keys: Data encryption keys

kubectl apply -f kubernetes/base/secrets.yaml
```

#### Step 3: Deploy Infrastructure Dependencies
```bash
# Deploy monitoring stack (if not already installed)
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --values helm/dependencies/prometheus-stack/values.yaml

# Deploy ingress controller (if not already installed)
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --values helm/dependencies/ingress-nginx/values.yaml
```

#### Step 4: Deploy Application
```bash
# Deploy with Helm
helm install freightliner ./helm/freightliner \
  -f ./helm/freightliner/values-production.yaml \
  --namespace freightliner-production \
  --create-namespace \
  --wait --timeout=15m

# Verify pods are running
kubectl get pods -n freightliner-production

# Check service status
kubectl get svc -n freightliner-production

# Verify ingress
kubectl get ingress -n freightliner-production
```

#### Step 5: Post-Deployment Validation
```bash
# Run health checks
./scripts/health-check.sh production

# Run smoke tests
./scripts/smoke-test.sh production

# Verify monitoring
./scripts/check-monitoring.sh production

echo "🎉 Deployment validation completed!"
```

## 🔧 Configuration Reference

### Production Values (values-production.yaml)

```yaml
# Production configuration for Freightliner
global:
  environment: production
  imageRegistry: "ghcr.io/hemzaz"
  imagePullSecrets:
    - name: registry-credentials

image:
  repository: freightliner
  tag: "1.0.0"
  pullPolicy: IfNotPresent

# High availability configuration
replicaCount: 3
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxUnavailable: 1
    maxSurge: 1

# Production resource allocation
resources:
  requests:
    cpu: 500m
    memory: 1Gi
  limits:
    cpu: 2
    memory: 4Gi

# Auto-scaling configuration
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

# Production service configuration
service:
  type: ClusterIP
  port: 80
  targetPort: 8080
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
    service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing

# Production ingress with TLS
ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: freightliner.company.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: freightliner-tls-cert
      hosts:
        - freightliner.company.com

# Production monitoring
monitoring:
  serviceMonitor:
    enabled: true
    interval: 30s
    path: /metrics
    labels:
      team: platform
      environment: production

# Production persistence
persistence:
  enabled: true
  storageClass: "fast-ssd"
  accessMode: ReadWriteOnce
  size: 50Gi

# Production security
securityContext:
  runAsNonRoot: true
  runAsUser: 10001
  runAsGroup: 10001
  fsGroup: 10001
  seccompProfile:
    type: RuntimeDefault

podSecurityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: true

# Production node scheduling
nodeSelector:
  node-type: production
  workload: stateful

tolerations:
  - key: "production"
    operator: "Equal"
    value: "true"
    effect: "NoSchedule"

affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/name
              operator: In
              values:
                - freightliner
        topologyKey: kubernetes.io/hostname

# Production network policies
networkPolicy:
  enabled: true
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 8080
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to: []
      ports:
        - protocol: TCP
          port: 443
        - protocol: TCP
          port: 80
        - protocol: UDP
          port: 53
```

## 🛠️ Deployment Scripts

### deploy.sh
```bash
#!/bin/bash
# Production deployment script

set -euo pipefail

ENVIRONMENT=${1:-production}
NAMESPACE="freightliner-${ENVIRONMENT}"
HELM_RELEASE="freightliner"

echo "🚀 Deploying Freightliner to ${ENVIRONMENT}..."

# Validate prerequisites
echo "📋 Validating prerequisites..."
if ! kubectl cluster-info &>/dev/null; then
    echo "❌ kubectl not configured or cluster unreachable"
    exit 1
fi

if ! helm version &>/dev/null; then
    echo "❌ Helm not installed or not in PATH"
    exit 1
fi

# Create namespace if it doesn't exist
kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# Label namespace
kubectl label namespace "${NAMESPACE}" \
  environment="${ENVIRONMENT}" \
  monitoring=enabled \
  security-policy=strict \
  --overwrite

# Deploy secrets
echo "🔐 Applying secrets..."
if [[ -f "kubernetes/base/secrets.yaml" ]]; then
    kubectl apply -f kubernetes/base/secrets.yaml -n "${NAMESPACE}"
else
    echo "⚠️  Warning: secrets.yaml not found. Please configure secrets manually."
fi

# Deploy with Helm
echo "📦 Deploying with Helm..."
helm upgrade --install "${HELM_RELEASE}" ./helm/freightliner \
  -f "./helm/freightliner/values-${ENVIRONMENT}.yaml" \
  --namespace "${NAMESPACE}" \
  --wait --timeout=15m \
  --create-namespace

# Wait for rollout
echo "⏳ Waiting for deployment rollout..."
kubectl rollout status deployment/freightliner -n "${NAMESPACE}" --timeout=600s

# Post-deployment validation
echo "✅ Running post-deployment validation..."
./scripts/health-check.sh "${ENVIRONMENT}"

echo "🎉 Deployment to ${ENVIRONMENT} completed successfully!"
```

### health-check.sh
```bash
#!/bin/bash
# Health check script

set -euo pipefail

ENVIRONMENT=${1:-production}
NAMESPACE="freightliner-${ENVIRONMENT}"

echo "🏥 Running health checks for ${ENVIRONMENT}..."

# Check pod status
echo "📊 Checking pod status..."
if ! kubectl get pods -n "${NAMESPACE}" -l app.kubernetes.io/name=freightliner | grep -q "Running"; then
    echo "❌ No running pods found"
    kubectl get pods -n "${NAMESPACE}" -l app.kubernetes.io/name=freightliner
    exit 1
fi

echo "✅ Pods are running"

# Check service endpoints
echo "🔗 Checking service endpoints..."
SERVICE_IP=$(kubectl get svc freightliner -n "${NAMESPACE}" -o jsonpath='{.spec.clusterIP}')
if [[ -z "${SERVICE_IP}" ]]; then
    echo "❌ Service IP not found"
    exit 1
fi

echo "✅ Service endpoints available: ${SERVICE_IP}"

# Check ingress
echo "🌐 Checking ingress..."
if kubectl get ingress -n "${NAMESPACE}" | grep -q freightliner; then
    INGRESS_HOST=$(kubectl get ingress freightliner -n "${NAMESPACE}" -o jsonpath='{.spec.rules[0].host}')
    echo "✅ Ingress configured: ${INGRESS_HOST}"
else
    echo "⚠️  Warning: No ingress found"
fi

# Health endpoint check
echo "❤️  Checking health endpoint..."
kubectl port-forward -n "${NAMESPACE}" svc/freightliner 8080:80 &
PF_PID=$!
sleep 5

if curl -f http://localhost:8080/health &>/dev/null; then
    echo "✅ Health endpoint responding"
else
    echo "❌ Health endpoint not responding"
    kill $PF_PID
    exit 1
fi

kill $PF_PID

# Metrics endpoint check
echo "📈 Checking metrics endpoint..."
kubectl port-forward -n "${NAMESPACE}" svc/freightliner 8080:80 &
PF_PID=$!
sleep 5

if curl -f http://localhost:8080/metrics &>/dev/null; then
    echo "✅ Metrics endpoint responding"
else
    echo "❌ Metrics endpoint not responding"
    kill $PF_PID
    exit 1
fi

kill $PF_PID

echo "🎉 All health checks passed!"
```

### rollback.sh
```bash
#!/bin/bash
# Rollback script

set -euo pipefail

ENVIRONMENT=${1:-production}
NAMESPACE="freightliner-${ENVIRONMENT}"
HELM_RELEASE="freightliner"

echo "🔄 Rolling back Freightliner in ${ENVIRONMENT}..."

# Check Helm history
echo "📜 Checking deployment history..."
helm history "${HELM_RELEASE}" -n "${NAMESPACE}"

# Get previous revision
PREVIOUS_REVISION=$(helm history "${HELM_RELEASE}" -n "${NAMESPACE}" --max 2 -o json | jq -r '.[1].revision')

if [[ "${PREVIOUS_REVISION}" == "null" || -z "${PREVIOUS_REVISION}" ]]; then
    echo "❌ No previous revision found for rollback"
    exit 1
fi

echo "🔄 Rolling back to revision ${PREVIOUS_REVISION}..."

# Perform rollback
helm rollback "${HELM_RELEASE}" "${PREVIOUS_REVISION}" -n "${NAMESPACE}" --wait --timeout=10m

# Wait for rollout
echo "⏳ Waiting for rollback to complete..."
kubectl rollout status deployment/freightliner -n "${NAMESPACE}" --timeout=600s

# Post-rollback validation
echo "✅ Running post-rollback validation..."
./scripts/health-check.sh "${ENVIRONMENT}"

echo "🎉 Rollback completed successfully!"
```

## 📊 Monitoring and Observability

### Prometheus Metrics
The application exposes the following metrics at `/metrics`:

```
# System metrics
freightliner_build_info - Build information
freightliner_uptime_seconds - Application uptime
freightliner_memory_usage_bytes - Memory usage
freightliner_cpu_usage_percent - CPU utilization

# Business metrics
freightliner_replications_total - Total replications
freightliner_replication_duration_seconds - Replication duration
freightliner_replication_errors_total - Replication errors
freightliner_throughput_mbps - Replication throughput

# Infrastructure metrics
freightliner_registry_connections - Registry connections
freightliner_worker_pool_size - Worker pool size
freightliner_queue_depth - Job queue depth
```

### Grafana Dashboards
Pre-configured dashboards available in `monitoring/grafana/`:
- Application Overview Dashboard
- Performance Metrics Dashboard
- Error Tracking Dashboard
- Infrastructure Metrics Dashboard

### Alerting Rules
Critical alerting rules in `monitoring/prometheus/`:
- High error rate (>5%)
- Performance degradation (>20% slower)
- Memory/CPU exhaustion
- Service unavailability

## 🔒 Security Configuration

### Network Policies
```yaml
# Restrict network traffic to essential connections only
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: freightliner-network-policy
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: freightliner
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 80
```

### Pod Security Standards
```yaml
# Enforce restricted pod security standards
apiVersion: v1
kind: Namespace
metadata:
  name: freightliner-production
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

### Secret Management
```bash
# Create secrets using kubectl (example)
kubectl create secret generic freightliner-registry-credentials \
  --from-literal=username=your-username \
  --from-literal=password=your-password \
  --namespace freightliner-production

kubectl create secret tls freightliner-tls-cert \
  --cert=tls.crt \
  --key=tls.key \
  --namespace freightliner-production
```

## 🚨 Troubleshooting Guide

### Common Issues and Solutions

#### 1. Pods Not Starting
```bash
# Check pod status and events
kubectl describe pods -n freightliner-production -l app.kubernetes.io/name=freightliner

# Check logs
kubectl logs -n freightliner-production -l app.kubernetes.io/name=freightliner --tail=100

# Common causes:
# - Image pull errors: Check registry credentials
# - Resource constraints: Verify cluster resources
# - Configuration errors: Validate ConfigMap and Secrets
```

#### 2. Service Not Accessible
```bash
# Check service configuration
kubectl get svc -n freightliner-production -o wide

# Check endpoints
kubectl get endpoints -n freightliner-production

# Test internal connectivity
kubectl exec -it deployment/freightliner -n freightliner-production -- wget -qO- http://localhost:8080/health
```

#### 3. Ingress Issues
```bash
# Check ingress status
kubectl describe ingress -n freightliner-production

# Check ingress controller logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx

# Verify DNS resolution
nslookup freightliner.company.com
```

#### 4. Performance Issues
```bash
# Check resource usage
kubectl top pods -n freightliner-production

# Check HPA status
kubectl get hpa -n freightliner-production

# Review performance metrics
kubectl port-forward -n freightliner-production svc/freightliner 8080:80
curl http://localhost:8080/metrics | grep freightliner_performance
```

### Emergency Procedures

#### Immediate Rollback
```bash
# Quick rollback to previous version
./scripts/rollback.sh production
```

#### Scale Down (Emergency)
```bash
# Scale to zero replicas (emergency stop)
kubectl scale deployment freightliner --replicas=0 -n freightliner-production

# Scale back up
kubectl scale deployment freightliner --replicas=3 -n freightliner-production
```

#### Debug Mode
```bash
# Enable debug logging
kubectl patch deployment freightliner -n freightliner-production -p '{"spec":{"template":{"spec":{"containers":[{"name":"freightliner","env":[{"name":"LOG_LEVEL","value":"debug"}]}]}}}}'
```

## 📞 Support and Escalation

### Support Contacts
- **Platform Team:** platform@company.com
- **On-Call:** +1-555-PLATFORM
- **Slack Channel:** #freightliner-support

### Escalation Matrix
1. **Level 1:** Platform Engineering Team
2. **Level 2:** Infrastructure Team Lead
3. **Level 3:** Engineering Director
4. **Level 4:** CTO

### SLA Commitments
- **Critical Issues:** 15 minutes response, 2 hours resolution
- **High Issues:** 1 hour response, 8 hours resolution
- **Medium Issues:** 4 hours response, 24 hours resolution
- **Low Issues:** 24 hours response, 72 hours resolution

---

## ✅ Deployment Checklist

### Pre-Deployment
- [ ] Prerequisites verified (Kubernetes, Helm, access)
- [ ] Infrastructure dependencies deployed
- [ ] Secrets configured and validated
- [ ] DNS and TLS certificates prepared
- [ ] Monitoring systems operational
- [ ] Backup procedures tested

### Deployment
- [ ] Namespace created and labeled
- [ ] Secrets applied
- [ ] Helm deployment successful
- [ ] Pods running and healthy
- [ ] Services accessible
- [ ] Ingress configured correctly

### Post-Deployment
- [ ] Health checks passing
- [ ] Smoke tests successful
- [ ] Monitoring dashboards operational
- [ ] Alerts configured and tested
- [ ] Documentation updated
- [ ] Team notified of deployment

### Production Validation
- [ ] Performance benchmarks met
- [ ] Security scans clean
- [ ] Load testing completed
- [ ] Disaster recovery tested
- [ ] Runbooks validated

---

**Deployment Package Version:** 1.0.0  
**Last Updated:** August 3, 2025  
**Next Review:** September 3, 2025  

**Package Prepared By:** Production Engineering Team  
**Approved By:** Platform Architecture Team  

## 🎯 Quick Reference

### Essential Commands
```bash
# Deploy production
./scripts/deploy.sh production

# Health check
./scripts/health-check.sh production

# Rollback
./scripts/rollback.sh production

# View logs
kubectl logs -n freightliner-production -l app.kubernetes.io/name=freightliner -f

# Scale replicas
kubectl scale deployment freightliner --replicas=5 -n freightliner-production

# Port forward for debugging
kubectl port-forward -n freightliner-production svc/freightliner 8080:80
```

### Important URLs
- **Application:** https://freightliner.company.com
- **Health:** https://freightliner.company.com/health
- **Metrics:** https://freightliner.company.com/metrics
- **Grafana:** https://grafana.company.com/d/freightliner
- **Prometheus:** https://prometheus.company.com/targets

### Emergency Contacts
- **Platform Team:** Slack #platform-team
- **On-Call:** PagerDuty escalation
- **Security:** security@company.com

---

**🚀 Ready for Production Deployment!**