# Freightliner Monitoring Setup Guide

Complete setup guide for deploying and configuring the Freightliner monitoring stack in local development and remote production environments.

---

## Table of Contents

1. [Local Development Setup](#1-local-development-setup)
2. [Remote Deployment Setup](#2-remote-deployment-setup)
3. [Configuration Reference](#3-configuration-reference)
4. [Troubleshooting](#4-troubleshooting)
5. [Quick Reference Commands](#5-quick-reference-commands)

---

## 1. Local Development Setup

### 1.1 Docker Compose Quick Start

The fastest way to get the monitoring stack running locally.

**Prerequisites:**
- Docker Engine 20.10+
- Docker Compose v2.0+
- 4GB available RAM

**Start the full stack:**

```bash
# Navigate to project root
cd /path/to/freightliner

# Start all services including monitoring
docker-compose up -d

# Or start only monitoring services
docker-compose up -d prometheus grafana redis
```

**Verify services are running:**

```bash
docker-compose ps
```

Expected output:
```
NAME                           STATUS    PORTS
freightliner-dev               running   0.0.0.0:8080->8080/tcp
freightliner-prometheus-dev    running   0.0.0.0:9090->9090/tcp
freightliner-grafana-dev       running   0.0.0.0:3000->3000/tcp
freightliner-redis-dev         running   0.0.0.0:6379->6379/tcp
freightliner-jaeger-dev        running   0.0.0.0:16686->16686/tcp
```

### 1.2 Accessing Services

| Service | URL | Description |
|---------|-----|-------------|
| **Grafana** | http://localhost:3000 | Dashboards and visualization |
| **Prometheus** | http://localhost:9090 | Metrics and alerting |
| **API** | http://localhost:8080 | Freightliner application |
| **Metrics Endpoint** | http://localhost:2112/metrics | Raw Prometheus metrics |
| **Jaeger** | http://localhost:16686 | Distributed tracing |
| **MinIO Console** | http://localhost:9001 | S3-compatible storage |

### 1.3 Default Credentials

| Service | Username | Password | Notes |
|---------|----------|----------|-------|
| **Grafana** | `admin` | `admin` | Change on first login |
| **MinIO** | `minioadmin` | `minioadmin123` | Development only |

**Change Grafana password:**
1. Log into http://localhost:3000
2. Navigate to Profile (bottom left) > Change Password
3. Or set via environment variable:
   ```yaml
   environment:
     - GF_SECURITY_ADMIN_PASSWORD=your_secure_password
   ```

### 1.4 Import Dashboards

**Option A: Manual Import**

1. Open Grafana at http://localhost:3000
2. Log in with `admin`/`admin`
3. Navigate to **Dashboards** (left sidebar) > **Import**
4. Click **Upload JSON file**
5. Import each dashboard:
   - `monitoring/grafana/dashboards/replication-overview.json`
   - `monitoring/grafana/dashboards/infrastructure-metrics.json`
   - `monitoring/grafana/dashboards/business-metrics.json`
   - `monitoring/grafana/dashboards/error-latency.json`
6. Select **Prometheus** as the data source when prompted
7. Click **Import**

**Option B: Automatic Provisioning**

Create the provisioning directory structure:

```bash
mkdir -p config/grafana/dashboards config/grafana/datasources
```

Create datasource configuration:

```bash
cat > config/grafana/datasources/prometheus.yaml << 'EOF'
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF
```

Create dashboard provider:

```bash
cat > config/grafana/dashboards/dashboards.yaml << 'EOF'
apiVersion: 1
providers:
  - name: 'freightliner'
    orgId: 1
    folder: 'Freightliner'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 30
    options:
      path: /var/lib/grafana/dashboards
EOF
```

Copy dashboards:

```bash
cp monitoring/grafana/dashboards/*.json config/grafana/dashboards/
```

Restart Grafana:

```bash
docker-compose restart grafana
```

### 1.5 Verify Prometheus Targets

1. Open http://localhost:9090/targets
2. Confirm all targets show **UP** status:
   - `freightliner` - Application metrics
   - `prometheus` - Self-monitoring

**If targets show DOWN:**
```bash
# Check if Freightliner is running
curl http://localhost:2112/metrics

# Check Prometheus logs
docker-compose logs prometheus
```

---

## 2. Remote Deployment Setup

### 2.1 Kubernetes with Custom IP/Domain

**Prerequisites:**
- Kubernetes cluster 1.21+
- kubectl configured
- Helm 3.x (optional)
- Storage class available

**Deploy the monitoring stack:**

```bash
# Create namespace
kubectl create namespace monitoring

# Deploy Prometheus
kubectl apply -f monitoring/kubernetes/prometheus-deployment.yaml

# Deploy Grafana
kubectl apply -f monitoring/kubernetes/grafana-deployment.yaml

# Deploy AlertManager
kubectl apply -f monitoring/kubernetes/alertmanager-deployment.yaml

# Deploy ServiceMonitor for Freightliner
kubectl apply -f monitoring/kubernetes/servicemonitor-freightliner.yaml
```

**Create required secrets:**

```bash
# Grafana admin credentials
kubectl create secret generic grafana-admin \
  --from-literal=username=admin \
  --from-literal=password=YOUR_SECURE_PASSWORD \
  -n monitoring

# AlertManager notification secrets (optional)
kubectl create secret generic alertmanager-secrets \
  --from-literal=slack-webhook='https://hooks.slack.com/services/YOUR/WEBHOOK' \
  --from-literal=pagerduty-key='YOUR_PAGERDUTY_KEY' \
  -n monitoring
```

### 2.2 Setting External URLs

**Option A: Using External IPs**

Patch services to use your server's IP address:

```bash
# Set your server IP
export SERVER_IP=192.168.1.100

# Patch Grafana service
kubectl patch svc grafana -n monitoring -p "{\"spec\":{\"externalIPs\":[\"${SERVER_IP}\"]}}"

# Patch Prometheus service
kubectl patch svc prometheus -n monitoring -p "{\"spec\":{\"externalIPs\":[\"${SERVER_IP}\"]}}"

# Patch AlertManager service
kubectl patch svc alertmanager -n monitoring -p "{\"spec\":{\"externalIPs\":[\"${SERVER_IP}\"]}}"
```

Access URLs with external IPs:
- Grafana: http://192.168.1.100:3000
- Prometheus: http://192.168.1.100:9090
- AlertManager: http://192.168.1.100:9093

**Option B: Using LoadBalancer (Cloud providers)**

The default configuration includes LoadBalancer services:

```bash
# Check LoadBalancer external IPs
kubectl get svc -n monitoring

# Example output:
# NAME                   TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)
# grafana-external       LoadBalancer   10.96.123.45    34.123.45.67    80:31234/TCP
# prometheus-external    LoadBalancer   10.96.123.46    34.123.45.68    9090:31235/TCP
```

**Option C: Using NodePort**

Convert services to NodePort:

```bash
# Create NodePort service for Grafana
kubectl patch svc grafana -n monitoring -p '{"spec":{"type":"NodePort","ports":[{"port":3000,"nodePort":30300,"name":"http"}]}}'

# Create NodePort service for Prometheus
kubectl patch svc prometheus -n monitoring -p '{"spec":{"type":"NodePort","ports":[{"port":9090,"nodePort":30909,"name":"http"}]}}'

# Create NodePort service for AlertManager
kubectl patch svc alertmanager -n monitoring -p '{"spec":{"type":"NodePort","ports":[{"port":9093,"nodePort":30093,"name":"http"}]}}'
```

Access via any node IP:
- Grafana: http://<NODE_IP>:30300
- Prometheus: http://<NODE_IP>:30909
- AlertManager: http://<NODE_IP>:30093

### 2.3 Ingress Configuration

**Create Ingress for all monitoring services:**

```bash
cat << 'EOF' | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: monitoring-ingress
  namespace: monitoring
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  rules:
    - host: grafana.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: grafana
                port:
                  number: 3000
    - host: prometheus.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: prometheus
                port:
                  number: 9090
    - host: alertmanager.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: alertmanager
                port:
                  number: 9093
EOF
```

**Update Grafana root URL:**

```bash
kubectl set env deployment/grafana -n monitoring \
  GF_SERVER_ROOT_URL=https://grafana.example.com
```

### 2.4 TLS/SSL Setup

**Option A: cert-manager with Let's Encrypt**

```bash
# Install cert-manager (if not already installed)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create ClusterIssuer for Let's Encrypt
cat << 'EOF' | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: nginx
EOF

# Update Ingress with TLS
cat << 'EOF' | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: monitoring-ingress
  namespace: monitoring
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
    - hosts:
        - grafana.example.com
        - prometheus.example.com
        - alertmanager.example.com
      secretName: monitoring-tls
  rules:
    - host: grafana.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: grafana
                port:
                  number: 3000
    - host: prometheus.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: prometheus
                port:
                  number: 9090
    - host: alertmanager.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: alertmanager
                port:
                  number: 9093
EOF
```

**Option B: Manual TLS with existing certificates**

```bash
# Create TLS secret from existing certificates
kubectl create secret tls monitoring-tls \
  --cert=/path/to/tls.crt \
  --key=/path/to/tls.key \
  -n monitoring
```

### 2.5 Firewall Considerations

**Required ports:**

| Port | Protocol | Service | Direction |
|------|----------|---------|-----------|
| 3000 | TCP | Grafana | Inbound |
| 9090 | TCP | Prometheus | Inbound |
| 9093 | TCP | AlertManager | Inbound |
| 2112 | TCP | Freightliner metrics | Internal |
| 443 | TCP | HTTPS (Ingress) | Inbound |
| 80 | TCP | HTTP (Ingress) | Inbound |

**iptables example:**

```bash
# Allow Grafana
iptables -A INPUT -p tcp --dport 3000 -j ACCEPT

# Allow Prometheus
iptables -A INPUT -p tcp --dport 9090 -j ACCEPT

# Allow AlertManager
iptables -A INPUT -p tcp --dport 9093 -j ACCEPT

# Save rules
iptables-save > /etc/iptables/rules.v4
```

**UFW example:**

```bash
ufw allow 3000/tcp comment 'Grafana'
ufw allow 9090/tcp comment 'Prometheus'
ufw allow 9093/tcp comment 'AlertManager'
ufw reload
```

**AWS Security Group:**

```bash
# Allow from specific CIDR
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxxxxx \
  --protocol tcp \
  --port 3000 \
  --cidr 10.0.0.0/8
```

---

## 3. Configuration Reference

### 3.1 Environment Variables

**Grafana Configuration:**

| Variable | Default | Description |
|----------|---------|-------------|
| `GF_SECURITY_ADMIN_USER` | `admin` | Admin username |
| `GF_SECURITY_ADMIN_PASSWORD` | `admin` | Admin password |
| `GF_USERS_ALLOW_SIGN_UP` | `false` | Allow user registration |
| `GF_SERVER_ROOT_URL` | `http://localhost:3000` | External URL for Grafana |
| `GF_INSTALL_PLUGINS` | (empty) | Comma-separated plugin list |
| `GF_ANALYTICS_REPORTING_ENABLED` | `true` | Anonymous usage reporting |
| `GF_AUTH_ANONYMOUS_ENABLED` | `false` | Allow anonymous access |

**Prometheus Configuration:**

| Variable | Default | Description |
|----------|---------|-------------|
| `--storage.tsdb.retention.time` | `15d` | Data retention period |
| `--storage.tsdb.retention.size` | (empty) | Max storage size |
| `--web.enable-lifecycle` | `false` | Enable reload API |
| `--web.enable-admin-api` | `false` | Enable admin API |
| `--web.external-url` | (empty) | External URL for links |

**AlertManager Configuration:**

| Variable | Default | Description |
|----------|---------|-------------|
| `--data.retention` | `120h` | Alert data retention |
| `--web.external-url` | (empty) | External URL for links |
| `--cluster.listen-address` | `:9094` | Cluster peer address |

### 3.2 Kustomize Overlays

Create environment-specific overlays for monitoring:

**Directory structure:**

```
monitoring/kubernetes/
├── base/
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── prometheus-deployment.yaml
│   ├── grafana-deployment.yaml
│   └── alertmanager-deployment.yaml
└── overlays/
    ├── local/
    │   ├── kustomization.yaml
    │   └── patches/
    └── remote/
        ├── kustomization.yaml
        └── patches/
```

**Base kustomization.yaml:**

```yaml
# monitoring/kubernetes/base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: monitoring

resources:
  - namespace.yaml
  - prometheus-deployment.yaml
  - grafana-deployment.yaml
  - alertmanager-deployment.yaml

commonLabels:
  app.kubernetes.io/part-of: freightliner-monitoring
```

**Local overlay:**

```yaml
# monitoring/kubernetes/overlays/local/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - patch: |-
      - op: replace
        path: /spec/type
        value: NodePort
    target:
      kind: Service
      name: grafana
  - patch: |-
      - op: replace
        path: /spec/type
        value: NodePort
    target:
      kind: Service
      name: prometheus

configMapGenerator:
  - name: grafana-config
    behavior: merge
    literals:
      - GF_SERVER_ROOT_URL=http://localhost:3000
```

**Remote overlay:**

```yaml
# monitoring/kubernetes/overlays/remote/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base
  - ingress.yaml

patches:
  - patch: |-
      - op: add
        path: /spec/externalIPs
        value:
          - "192.168.1.100"
    target:
      kind: Service
      name: grafana-external
  - patch: |-
      - op: add
        path: /spec/externalIPs
        value:
          - "192.168.1.100"
    target:
      kind: Service
      name: prometheus-external

configMapGenerator:
  - name: grafana-config
    behavior: merge
    literals:
      - GF_SERVER_ROOT_URL=https://grafana.example.com
```

**Deploy with Kustomize:**

```bash
# Local deployment
kubectl apply -k monitoring/kubernetes/overlays/local

# Remote deployment
kubectl apply -k monitoring/kubernetes/overlays/remote
```

### 3.3 ConfigMap Options

**Prometheus ConfigMap:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
      external_labels:
        cluster: 'production'

    scrape_configs:
      - job_name: 'freightliner'
        static_configs:
          - targets: ['freightliner.freightliner.svc:2112']

      - job_name: 'prometheus'
        static_configs:
          - targets: ['localhost:9090']
```

**Grafana Datasource ConfigMap:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
  namespace: monitoring
data:
  prometheus.yaml: |
    apiVersion: 1
    datasources:
      - name: Prometheus
        type: prometheus
        access: proxy
        url: http://prometheus:9090
        isDefault: true
        jsonData:
          timeInterval: "30s"
```

### 3.4 Service Types Comparison

| Type | Use Case | External Access | Load Balancing |
|------|----------|-----------------|----------------|
| **ClusterIP** | Internal only | Port-forward required | Internal only |
| **NodePort** | Development | `<NodeIP>:<NodePort>` | None |
| **LoadBalancer** | Cloud production | Cloud LB IP | Cloud provider |
| **ExternalIP** | On-premise | Fixed IP | None |
| **Ingress** | Production | Domain-based | L7 |

**Choosing the right type:**

```
Local development:
  └── NodePort or port-forward

On-premise Kubernetes:
  └── NodePort + external load balancer
  └── Or ExternalIP with direct access

Cloud Kubernetes (EKS/GKE/AKS):
  └── LoadBalancer for internal use
  └── Ingress + LoadBalancer for external access

Production with domain:
  └── Ingress with TLS
```

---

## 4. Troubleshooting

### 4.1 Cannot Access Grafana

**Symptom:** Browser shows "Connection refused" or times out.

**Check 1: Service is running**
```bash
# Docker Compose
docker-compose ps grafana
docker-compose logs grafana

# Kubernetes
kubectl get pods -n monitoring -l app=grafana
kubectl logs -n monitoring -l app=grafana
```

**Check 2: Port is exposed**
```bash
# Docker
docker port freightliner-grafana-dev

# Kubernetes - check service
kubectl get svc grafana -n monitoring
kubectl describe svc grafana -n monitoring
```

**Check 3: Network connectivity**
```bash
# From local machine
curl -v http://localhost:3000/api/health

# From within cluster
kubectl run curl --image=curlimages/curl -i --tty --rm -- \
  curl http://grafana.monitoring.svc:3000/api/health
```

**Check 4: Firewall/Security Group**
```bash
# Check if port is blocked
nc -zv localhost 3000

# Check iptables
iptables -L -n | grep 3000
```

**Solutions:**

```bash
# Solution 1: Restart service
docker-compose restart grafana
# or
kubectl rollout restart deployment/grafana -n monitoring

# Solution 2: Port forward (temporary access)
kubectl port-forward -n monitoring svc/grafana 3000:3000

# Solution 3: Check and fix service type
kubectl patch svc grafana -n monitoring -p '{"spec":{"type":"NodePort"}}'
```

### 4.2 Prometheus Not Scraping Metrics

**Symptom:** Targets show as DOWN in Prometheus UI.

**Check 1: View target status**
```bash
# Access Prometheus targets page
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets'
```

**Check 2: Verify metrics endpoint**
```bash
# Direct curl to metrics endpoint
curl http://localhost:2112/metrics

# From Prometheus pod
kubectl exec -n monitoring deploy/prometheus -- \
  wget -qO- http://freightliner.freightliner.svc:2112/metrics
```

**Check 3: Prometheus configuration**
```bash
# View current config
kubectl get configmap prometheus-config -n monitoring -o yaml

# Check scrape config
curl http://localhost:9090/api/v1/status/config | jq '.data.yaml' -r
```

**Check 4: Network policies**
```bash
# Check if NetworkPolicy blocks traffic
kubectl get networkpolicy -A

# Test connectivity
kubectl run test --image=busybox -i --tty --rm -- \
  wget -qO- http://freightliner.freightliner.svc:2112/metrics
```

**Solutions:**

```bash
# Solution 1: Fix target address in prometheus.yml
# Ensure correct service name and port

# Solution 2: Add prometheus.io annotations to service
kubectl annotate svc freightliner -n freightliner \
  prometheus.io/scrape=true \
  prometheus.io/port=2112 \
  prometheus.io/path=/metrics

# Solution 3: Reload Prometheus config
curl -X POST http://localhost:9090/-/reload
```

### 4.3 CORS Errors

**Symptom:** Browser console shows CORS-related errors.

**Check 1: Identify the origin**
```bash
# Check browser console for specific error
# Example: "Access to XMLHttpRequest at 'http://prometheus:9090'
#          from origin 'http://grafana:3000' has been blocked"
```

**Solutions:**

```bash
# Solution 1: Configure Grafana datasource with proxy mode
# In Grafana UI: Configuration > Data Sources > Prometheus
# Set Access: Server (proxy)

# Solution 2: Enable CORS in Prometheus
# Add to prometheus args:
--web.cors.origin=".*"

# Solution 3: Use ingress with CORS headers
kubectl annotate ingress monitoring-ingress -n monitoring \
  nginx.ingress.kubernetes.io/enable-cors="true" \
  nginx.ingress.kubernetes.io/cors-allow-origin="*"
```

**For Docker Compose:**

```yaml
prometheus:
  command:
    - '--web.cors.origin=.*'
```

### 4.4 Network Connectivity Issues

**Symptom:** Services cannot communicate with each other.

**Diagnostic commands:**

```bash
# Check pod networking
kubectl get pods -n monitoring -o wide

# Check service endpoints
kubectl get endpoints -n monitoring

# DNS resolution test
kubectl run dns-test --image=busybox -i --tty --rm -- \
  nslookup grafana.monitoring.svc.cluster.local

# Connectivity test
kubectl run net-test --image=nicolaka/netshoot -i --tty --rm -- \
  curl -v http://prometheus.monitoring.svc:9090/-/healthy
```

**Check 1: DNS resolution**
```bash
# From within the cluster
kubectl exec -n monitoring deploy/grafana -- \
  nslookup prometheus.monitoring.svc.cluster.local
```

**Check 2: Service discovery**
```bash
# Verify endpoints exist
kubectl get endpoints prometheus -n monitoring
kubectl describe endpoints prometheus -n monitoring
```

**Check 3: Network policies**
```bash
# List all network policies
kubectl get networkpolicy -n monitoring

# Check policy details
kubectl describe networkpolicy -n monitoring
```

**Solutions:**

```bash
# Solution 1: Create NetworkPolicy allowing monitoring traffic
cat << 'EOF' | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-monitoring
  namespace: monitoring
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
  egress:
    - to:
        - namespaceSelector: {}
EOF

# Solution 2: Verify service selectors match pod labels
kubectl get pods -n monitoring --show-labels
kubectl get svc -n monitoring -o wide

# Solution 3: Check CoreDNS is running
kubectl get pods -n kube-system -l k8s-app=kube-dns
```

### 4.5 Common Error Messages

**"connection refused"**
- Service not running or wrong port
- Firewall blocking connection
- Wrong service name/IP

**"no such host"**
- DNS resolution failing
- Wrong service name
- CoreDNS not running

**"context deadline exceeded"**
- Network timeout
- Service overloaded
- Network policy blocking

**"permission denied"**
- RBAC issues
- Pod security policy
- File permissions in container

---

## 5. Quick Reference Commands

### Local Development

```bash
# Start monitoring stack
docker-compose up -d prometheus grafana

# View logs
docker-compose logs -f grafana
docker-compose logs -f prometheus

# Stop all services
docker-compose down

# Reset and start fresh
docker-compose down -v
docker-compose up -d

# Access services
open http://localhost:3000   # Grafana
open http://localhost:9090   # Prometheus
open http://localhost:8080   # API
```

### Local Kubernetes (minikube/kind/k3s)

```bash
# Deploy monitoring stack
kubectl apply -f monitoring/kubernetes/prometheus-deployment.yaml
kubectl apply -f monitoring/kubernetes/grafana-deployment.yaml
kubectl apply -f monitoring/kubernetes/alertmanager-deployment.yaml

# Or with Kustomize
kubectl apply -k monitoring/kubernetes/overlays/local

# Port forward for access
kubectl port-forward -n monitoring svc/grafana 3000:3000 &
kubectl port-forward -n monitoring svc/prometheus 9090:9090 &
kubectl port-forward -n monitoring svc/alertmanager 9093:9093 &

# Check status
kubectl get pods -n monitoring
kubectl get svc -n monitoring

# View logs
kubectl logs -n monitoring -l app=grafana -f
kubectl logs -n monitoring -l app=prometheus -f
```

### Remote Kubernetes

```bash
# Deploy with remote overlay
kubectl apply -k monitoring/kubernetes/overlays/remote

# Set external IP
kubectl patch svc grafana -n monitoring \
  -p '{"spec":{"externalIPs":["192.168.1.100"]}}'

kubectl patch svc prometheus -n monitoring \
  -p '{"spec":{"externalIPs":["192.168.1.100"]}}'

kubectl patch svc alertmanager -n monitoring \
  -p '{"spec":{"externalIPs":["192.168.1.100"]}}'

# Or use LoadBalancer and get external IP
kubectl get svc -n monitoring -w

# Create Grafana admin secret
kubectl create secret generic grafana-admin \
  --from-literal=username=admin \
  --from-literal=password=SecureP@ssw0rd! \
  -n monitoring

# Reload Prometheus config without restart
kubectl exec -n monitoring deploy/prometheus -- \
  wget -qO- --post-data='' http://localhost:9090/-/reload
```

### Debugging

```bash
# Get all resources in monitoring namespace
kubectl get all -n monitoring

# Describe pods for events
kubectl describe pods -n monitoring

# Check PVC status
kubectl get pvc -n monitoring

# Exec into pod for debugging
kubectl exec -it -n monitoring deploy/prometheus -- sh
kubectl exec -it -n monitoring deploy/grafana -- bash

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets | jq

# Check Prometheus alerts
curl http://localhost:9090/api/v1/alerts | jq

# Test AlertManager
curl http://localhost:9093/api/v2/status | jq

# Check Grafana health
curl http://localhost:3000/api/health
```

### Backup and Restore

```bash
# Backup Prometheus data
kubectl exec -n monitoring deploy/prometheus -- \
  tar czf /tmp/prometheus-backup.tar.gz /prometheus
kubectl cp monitoring/prometheus-xxx:/tmp/prometheus-backup.tar.gz ./backup/

# Backup Grafana dashboards
curl -H "Authorization: Bearer $API_KEY" \
  http://localhost:3000/api/search | \
  jq -r '.[].uid' | \
  xargs -I {} curl -H "Authorization: Bearer $API_KEY" \
  http://localhost:3000/api/dashboards/uid/{} > dashboards-backup.json

# Export Grafana datasources
curl -H "Authorization: Bearer $API_KEY" \
  http://localhost:3000/api/datasources > datasources-backup.json
```

---

## Related Documentation

- [Monitoring Architecture](./ARCHITECTURE.md)
- [Monitoring Setup Details](./docs/MONITORING_SETUP.md)
- [SLO Runbooks](./docs/SLO_RUNBOOKS.md)
- [Deployment Checklist](./DEPLOYMENT_CHECKLIST.md)
- [Kubernetes Base Deployment](../deployments/kubernetes/base/)
