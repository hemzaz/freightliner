# Freightliner Operations Guide

## Overview

This guide covers deployment, monitoring, maintenance, and troubleshooting of Freightliner in production environments. The application is designed for enterprise-grade reliability with comprehensive observability and operational features.

## 🚀 Deployment Readiness Status (January 2025)

**✅ READY FOR IMMEDIATE PRODUCTION DEPLOYMENT**

All critical blockers have been resolved and the application is production-ready:

### Pre-Deployment Checklist

| Component | Status | Notes |
|---|---|---|
| **Core Application** | ✅ **READY** | All P0 blockers resolved, stable build |
| **HTTP Server** | ✅ **READY** | No duplicate methods, health checks operational |
| **Worker Pool** | ✅ **READY** | Health monitoring implemented |
| **ECR Integration** | ✅ **READY** | Authentication and client fully functional |
| **GCR Integration** | ✅ **READY** | Google Container Registry operations stable |
| **Logging System** | ✅ **READY** | Structured JSON logging with trace support |
| **Metrics Collection** | ✅ **READY** | Prometheus endpoints active |
| **Configuration** | ✅ **READY** | Environment variables and CLI flags supported |

### Deployment Confidence Level: **HIGH** 🟢

- Zero compilation errors in core packages
- All critical runtime dependencies resolved
- Health check endpoints responding correctly
- Container registry clients authenticated and operational

## Production Deployment

### System Requirements

**Minimum Requirements:**
- CPU: 2 cores
- Memory: 512MB RAM
- Storage: 1GB (for checkpoints and logs)
- Network: HTTPS connectivity to registry endpoints

**Recommended Production:**
- CPU: 4+ cores
- Memory: 2GB+ RAM
- Storage: 10GB+ SSD (for checkpoints and persistent data)
- Network: High bandwidth for image transfers

**Supported Platforms:**
- Linux (x86_64, ARM64)
- macOS (x86_64, ARM64)
- Windows (x86_64)
- Docker containers
- Kubernetes clusters

### Container Deployment

#### Docker Production Setup

```bash
# Build production image with version info
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  --build-arg GIT_COMMIT=$(git rev-parse HEAD) \
  -t freightliner:v1.0.0 .

# Run with production configuration
docker run -d \
  --name freightliner-prod \
  --restart unless-stopped \
  -p 8080:8080 \
  -p 2112:2112 \
  -e LOG_LEVEL=info \
  -e API_KEY_AUTH=true \
  -e API_KEY=${FREIGHTLINER_API_KEY} \
  -e METRICS_ENABLED=true \
  -v /etc/freightliner:/etc/freightliner:ro \
  -v /data/freightliner:/data/checkpoints \
  --health-cmd="./freightliner health-check" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  freightliner:v1.0.0 serve --config=/etc/freightliner/config.yaml
```

#### Docker Compose Production

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  freightliner:
    image: freightliner:v1.0.0
    container_name: freightliner-prod
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "2112:2112"
    environment:
      - LOG_LEVEL=info
      - API_KEY_AUTH=true
      - API_KEY=${FREIGHTLINER_API_KEY}
      - METRICS_ENABLED=true
      - TLS_ENABLED=true
      - ECR_REGION=${AWS_REGION}
      - GCR_PROJECT=${GCP_PROJECT}
    volumes:
      - ./config:/etc/freightliner:ro
      - ./certs:/etc/ssl/freightliner:ro
      - freightliner-data:/data/checkpoints
      - freightliner-logs:/var/log/freightliner
    healthcheck:
      test: ["CMD", "./freightliner", "health-check"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
    networks:
      - freightliner-net
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '0.5'
        reservations:
          memory: 512M
          cpus: '0.25'

  # Optional: Prometheus for metrics collection
  prometheus:
    image: prometheus/prometheus:latest
    container_name: freightliner-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
    networks:
      - freightliner-net

volumes:
  freightliner-data:
    driver: local
  freightliner-logs:
    driver: local
  prometheus-data:
    driver: local

networks:
  freightliner-net:
    driver: bridge
```

### Kubernetes Deployment

#### Production Kubernetes Manifests

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: freightliner
  labels:
    app.kubernetes.io/name: freightliner
    app.kubernetes.io/instance: production

---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: freightliner-config
  namespace: freightliner
data:
  config.yaml: |
    log_level: info
    server:
      port: 8080
      api_key_auth: true
      tls_enabled: true
      tls_cert_file: "/etc/ssl/certs/tls.crt"
      tls_key_file: "/etc/ssl/private/tls.key"
      allowed_origins:
        - "https://freightliner.company.com"
      read_timeout: 30s
      write_timeout: 60s
      shutdown_timeout: 30s
    metrics:
      enabled: true
      port: 2112
      namespace: "freightliner"
    encryption:
      enabled: true
      customer_managed_keys: true
    secrets:
      use_secrets_manager: true
      secrets_manager_type: "aws"

---
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: freightliner-secrets
  namespace: freightliner
type: Opaque
data:
  api-key: <base64-encoded-api-key>
  aws-kms-key-id: <base64-encoded-kms-key>

---
# tls-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: freightliner-tls
  namespace: freightliner
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-certificate>
  tls.key: <base64-encoded-private-key>

---
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: freightliner-checkpoints
  namespace: freightliner
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 10Gi

---
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: freightliner
  namespace: freightliner
  labels:
    app.kubernetes.io/name: freightliner
    app.kubernetes.io/instance: production
    app.kubernetes.io/version: "v1.0.0"
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: freightliner
      app.kubernetes.io/instance: production
  template:
    metadata:
      labels:
        app.kubernetes.io/name: freightliner
        app.kubernetes.io/instance: production
        app.kubernetes.io/version: "v1.0.0"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "2112"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: freightliner
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: freightliner
        image: freightliner:v1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        - containerPort: 2112
          name: metrics
          protocol: TCP
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: PORT
          value: "8080"
        - name: METRICS_PORT
          value: "2112"
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: freightliner-secrets
              key: api-key
        - name: AWS_KMS_KEY_ID
          valueFrom:
            secretKeyRef:
              name: freightliner-secrets
              key: aws-kms-key-id
        volumeMounts:
        - name: config
          mountPath: /etc/freightliner
          readOnly: true
        - name: tls-certs
          mountPath: /etc/ssl/certs
          readOnly: true
        - name: tls-keys
          mountPath: /etc/ssl/private
          readOnly: true
        - name: checkpoints
          mountPath: /data/checkpoints
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
            scheme: HTTPS
          initialDelaySeconds: 60
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
            scheme: HTTPS
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 5
          failureThreshold: 3
        startupProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTPS
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 30
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
      volumes:
      - name: config
        configMap:
          name: freightliner-config
      - name: tls-certs
        secret:
          secretName: freightliner-tls
          items:
          - key: tls.crt
            path: tls.crt
      - name: tls-keys
        secret:
          secretName: freightliner-tls
          items:
          - key: tls.key
            path: tls.key
            mode: 0600
      - name: checkpoints
        persistentVolumeClaim:
          claimName: freightliner-checkpoints
      restartPolicy: Always
      terminationGracePeriodSeconds: 60

---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: freightliner
  namespace: freightliner
  labels:
    app.kubernetes.io/name: freightliner
    app.kubernetes.io/instance: production
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert: "arn:aws:acm:region:account:certificate/cert-id"
spec:
  type: LoadBalancer
  ports:
  - port: 443
    targetPort: 8080
    protocol: TCP
    name: https
  selector:
    app.kubernetes.io/name: freightliner
    app.kubernetes.io/instance: production

---
# service-metrics.yaml
apiVersion: v1
kind: Service
metadata:
  name: freightliner-metrics
  namespace: freightliner
  labels:
    app.kubernetes.io/name: freightliner
    app.kubernetes.io/instance: production
    app.kubernetes.io/component: metrics
spec:
  type: ClusterIP
  ports:
  - port: 2112
    targetPort: 2112
    protocol: TCP
    name: metrics
  selector:
    app.kubernetes.io/name: freightliner
    app.kubernetes.io/instance: production

---
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: freightliner
  namespace: freightliner
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - freightliner.company.com
    secretName: freightliner-ingress-tls
  rules:
  - host: freightliner.company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: freightliner
            port:
              number: 443

---
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: freightliner
  namespace: freightliner
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: freightliner
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60

---
# serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: freightliner
  namespace: freightliner
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT:role/FreightlinerServiceRole

---
# networkpolicy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: freightliner
  namespace: freightliner
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
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 2112
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 443  # HTTPS to registries
    - protocol: TCP
      port: 53   # DNS
    - protocol: UDP
      port: 53   # DNS
```

## Monitoring and Observability

### Health Checks

The application provides multiple health check endpoints for different use cases:

```bash
# Basic health check (load balancer)
curl -f https://freightliner.company.com/health

# Readiness check (Kubernetes)
curl -f https://freightliner.company.com/ready

# Liveness check (container orchestration)
curl -f https://freightliner.company.com/live

# Detailed system information
curl https://freightliner.company.com/health/system
```

### Prometheus Metrics

#### Key Metrics to Monitor

**HTTP Metrics:**
```
freightliner_http_requests_total{method,path,status}
freightliner_http_request_duration_seconds{method,path,status}
freightliner_http_requests_in_flight
```

**Application Metrics:**
```
freightliner_replication_total{source_registry,dest_registry,status}
freightliner_replication_duration_seconds{source_registry,dest_registry}
freightliner_replication_bytes_total{source_registry,dest_registry}
freightliner_jobs_active
freightliner_worker_pool_active
freightliner_worker_pool_queued
```

**System Metrics:**
```
freightliner_memory_usage_bytes
freightliner_goroutines_count
freightliner_panics_total{component}
```

#### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "freightliner_alerts.yml"

scrape_configs:
  - job_name: 'freightliner'
    static_configs:
      - targets: ['freightliner:2112']
    scrape_interval: 10s
    metrics_path: /metrics
    scheme: http

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

#### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Freightliner Operations",
    "panels": [
      {
        "title": "HTTP Requests Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(freightliner_http_requests_total[5m])",
            "legendFormat": "{{method}} {{path}} {{status}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(freightliner_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Active Replications",
        "type": "singlestat",
        "targets": [
          {
            "expr": "freightliner_jobs_active",
            "legendFormat": "Active Jobs"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "freightliner_memory_usage_bytes / 1024 / 1024",
            "legendFormat": "Memory (MB)"
          }
        ]
      }
    ]
  }
}
```

### Alerting Rules

```yaml
# freightliner_alerts.yml
groups:
  - name: freightliner
    rules:
      - alert: FreightlinerDown
        expr: up{job="freightliner"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Freightliner instance is down"
          description: "Freightliner instance {{ $labels.instance }} has been down for more than 1 minute."
      
      - alert: HighErrorRate
        expr: rate(freightliner_http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} requests per second"
      
      - alert: HighMemoryUsage
        expr: freightliner_memory_usage_bytes / 1024 / 1024 > 800
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is {{ $value }}MB"
      
      - alert: ReplicationFailures
        expr: rate(freightliner_replication_total{status="failed"}[10m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High replication failure rate"
          description: "Replication failure rate is {{ $value }} per second"
      
      - alert: LongRunningReplications
        expr: freightliner_replication_duration_seconds > 3600
        for: 0m
        labels:
          severity: warning
        annotations:
          summary: "Long-running replication detected"
          description: "Replication has been running for {{ $value }} seconds"
```

### Logging

#### Log Format

All logs are output in structured JSON format:

```json
{
  "timestamp": "2024-01-15T10:30:00.123456Z",
  "level": "info",
  "message": "HTTP request completed",
  "fields": {
    "method": "POST",
    "path": "/api/v1/replicate",
    "status": 200,
    "duration_ms": 1250.5,
    "user_agent": "curl/7.68.0",
    "request_id": "req_abc123",
    "source_registry": "ecr",
    "dest_registry": "gcr",
    "repository": "my-app"
  },
  "caller": {
    "file": "middleware.go",
    "line": 45,
    "function": "loggingMiddleware"
  },
  "trace_id": "abc123def456",
  "span_id": "789ghi012jkl"
}
```

#### Log Aggregation Setup

**ELK Stack Configuration:**

```yaml
# filebeat.yml
filebeat.inputs:
  - type: docker
    containers.ids:
      - "*freightliner*"
    json.keys_under_root: true
    json.message_key: message

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "freightliner-logs-%{+yyyy.MM.dd}"

setup.template:
  name: "freightliner"
  pattern: "freightliner-logs-*"
  settings:
    index.number_of_shards: 1
    index.number_of_replicas: 1
```

**Fluentd Configuration:**

```xml
<source>
  @type tail
  path /var/log/containers/*freightliner*.log
  pos_file /var/log/fluentd-freightliner.log.pos
  tag freightliner.*
  format json
  time_key timestamp
  time_format %Y-%m-%dT%H:%M:%S.%NZ
</source>

<match freightliner.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  index_name freightliner-logs
  type_name _doc
  include_tag_key true
  tag_key @log_name
  flush_interval 10s
</match>
```

## Backup and Recovery

### Checkpoint Management

```bash
# Backup checkpoints
tar -czf freightliner-checkpoints-$(date +%Y%m%d).tar.gz /data/checkpoints/

# Restore checkpoints
tar -xzf freightliner-checkpoints-20240115.tar.gz -C /

# List available checkpoints
ls -la /data/checkpoints/
```

### Configuration Backup

```bash
# Backup configuration
kubectl get configmap freightliner-config -n freightliner -o yaml > freightliner-config-backup.yaml
kubectl get secret freightliner-secrets -n freightliner -o yaml > freightliner-secrets-backup.yaml

# Backup persistent volumes
kubectl get pvc freightliner-checkpoints -n freightliner -o yaml > freightliner-pvc-backup.yaml
```

### Disaster Recovery

```bash
# Restore from backup
kubectl apply -f freightliner-config-backup.yaml
kubectl apply -f freightliner-secrets-backup.yaml
kubectl apply -f freightliner-pvc-backup.yaml

# Restart deployment
kubectl rollout restart deployment/freightliner -n freightliner

# Verify recovery
kubectl get pods -n freightliner
curl -f https://freightliner.company.com/health
```

## Performance Tuning

### Resource Optimization

**CPU Optimization:**
```yaml
# Adjust worker counts based on CPU cores
env:
  - name: REPLICATE_WORKERS
    value: "8"  # 2x CPU cores
  - name: SERVE_WORKERS
    value: "16"  # 4x CPU cores
```

**Memory Optimization:**
```yaml
resources:
  requests:
    memory: "1Gi"   # Base memory requirement
  limits:
    memory: "2Gi"   # Allow for peak usage
```

**Storage Optimization:**
```yaml
# Use SSD storage for checkpoints
storageClassName: fast-ssd
resources:
  requests:
    storage: 20Gi  # Size based on replication volume
```

### Network Optimization

```yaml
# Increase timeouts for large transfers
server:
  read_timeout: 300s   # 5 minutes for large images
  write_timeout: 600s  # 10 minutes for uploads
  
# Configure connection pooling
transport:
  max_idle_conns: 100
  max_idle_conns_per_host: 10
  idle_conn_timeout: 90s
```

## Security Operations

### TLS Certificate Management

```bash
# Generate self-signed certificate (development)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout freightliner.key \
  -out freightliner.crt \
  -subj "/CN=freightliner.company.com"

# Create Kubernetes TLS secret
kubectl create secret tls freightliner-tls \
  --cert=freightliner.crt \
  --key=freightliner.key \
  -n freightliner

# Renew certificates (Let's Encrypt)
certbot renew --deploy-hook "kubectl rollout restart deployment/freightliner -n freightliner"
```

### API Key Management

```bash
# Generate secure API key
openssl rand -base64 32

# Update API key in Kubernetes
kubectl create secret generic freightliner-secrets \
  --from-literal=api-key="$(openssl rand -base64 32)" \
  -n freightliner \
  --dry-run=client -o yaml | kubectl apply -f -

# Rotate API key
kubectl patch secret freightliner-secrets -n freightliner \
  -p='{"data":{"api-key":"'$(openssl rand -base64 32 | base64 -w 0)'"}}'
```

### Security Scanning

```bash
# Container security scanning
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image freightliner:v1.0.0

# Kubernetes security scanning
kube-score score freightliner-deployment.yaml

# Network policy validation
kubectl auth can-i --list --as=system:serviceaccount:freightliner:freightliner -n freightliner
```

## Troubleshooting

### Common Issues

#### Application Won't Start

```bash
# Check logs
kubectl logs -f deployment/freightliner -n freightliner

# Common issues:
# 1. Invalid configuration
freightliner serve --config=/etc/freightliner/config.yaml --log-level=debug

# 2. Missing API key
kubectl get secret freightliner-secrets -n freightliner -o yaml

# 3. Port conflicts
netstat -tulpn | grep :8080
```

#### Health Check Failures

```bash
# Test health endpoints directly
curl -v http://localhost:8080/health
curl -v http://localhost:8080/ready
curl -v http://localhost:8080/live

# Check certificate issues
openssl s_client -connect freightliner.company.com:443 -servername freightliner.company.com

# Verify DNS resolution
nslookup freightliner.company.com
```

#### High Memory Usage

```bash
# Check memory metrics
curl http://localhost:2112/metrics | grep memory

# Force garbage collection (temporary)
kubectl exec -it deployment/freightliner -n freightliner -- kill -USR1 1

# Scale up resources
kubectl patch deployment freightliner -n freightliner -p='{"spec":{"template":{"spec":{"containers":[{"name":"freightliner","resources":{"limits":{"memory":"2Gi"}}}]}}}}'
```

#### Replication Failures

```bash
# Check replication metrics
curl http://localhost:2112/metrics | grep replication

# Verify registry connectivity
kubectl exec -it deployment/freightliner -n freightliner -- curl -v https://123456789012.dkr.ecr.us-west-2.amazonaws.com/v2/

# Check authentication
aws ecr get-login-token --region us-west-2
gcloud auth print-access-token
```

### Debug Mode

```bash
# Enable debug logging
kubectl set env deployment/freightliner LOG_LEVEL=debug -n freightliner

# Port forward for local debugging
kubectl port-forward deployment/freightliner 8080:8080 2112:2112 -n freightliner

# Access debug endpoints
curl http://localhost:8080/health/system
curl http://localhost:2112/metrics
```

### Performance Debugging

```bash
# CPU profiling
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Memory profiling
curl http://localhost:8080/debug/pprof/heap > mem.prof
go tool pprof mem.prof

# Goroutine analysis
curl http://localhost:8080/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof
```

## Maintenance

### Rolling Updates

```bash
# Update image version
kubectl set image deployment/freightliner freightliner=freightliner:v1.1.0 -n freightliner

# Monitor rollout
kubectl rollout status deployment/freightliner -n freightliner

# Rollback if needed
kubectl rollout undo deployment/freightliner -n freightliner
```

### Scaling Operations

```bash
# Manual scaling
kubectl scale deployment freightliner --replicas=5 -n freightliner

# Update HPA limits
kubectl patch hpa freightliner -n freightliner -p='{"spec":{"maxReplicas":20}}'

# Verify scaling
kubectl get hpa freightliner -n freightliner
```

### Log Rotation

```bash
# Container log rotation (Docker)
docker run --log-driver=json-file --log-opt max-size=10m --log-opt max-file=3

# Kubernetes log rotation
kubectl patch daemonset fluentd -n kube-system -p='{"spec":{"template":{"spec":{"containers":[{"name":"fluentd","env":[{"name":"FLUENT_CONF","value":"fluent.conf"},{"name":"LOG_ROTATE_SIZE","value":"100MB"}]}]}}}}'
```

This operations guide provides comprehensive coverage of production deployment, monitoring, and maintenance procedures for Freightliner. Regular review and updates of these procedures ensure smooth operations in production environments.