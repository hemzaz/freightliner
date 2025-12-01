# Freightliner Monitoring Stack

Production-ready observability stack for container image replication monitoring with Prometheus, Grafana, and AlertManager.

## Quick Start

```bash
# Deploy complete monitoring stack to Kubernetes
kubectl apply -f monitoring/kubernetes/

# Access services locally
kubectl port-forward -n monitoring svc/grafana 3000:3000
kubectl port-forward -n monitoring svc/prometheus 9090:9090
kubectl port-forward -n monitoring svc/alertmanager 9093:9093
```

## Stack Components

### Prometheus (Port 9090)
- **Purpose**: Metrics collection and alerting
- **Scrape Interval**: 15 seconds
- **Retention**: 30 days (90GB)
- **Features**:
  - Kubernetes service discovery
  - Recording rules for performance
  - Alert rules with runbook links
  - Support for remote storage

### Grafana (Port 3000)
- **Purpose**: Metrics visualization
- **Default Credentials**: admin/changeme (change in production!)
- **Features**:
  - 4 pre-built dashboards
  - Automatic datasource provisioning
  - Dashboard version control
  - User management

### AlertManager (Port 9093)
- **Purpose**: Alert routing and notification
- **Features**:
  - Multi-channel notifications (Slack, PagerDuty)
  - Alert grouping and throttling
  - Inhibition rules
  - Silence management

## Directory Structure

```
monitoring/
├── prometheus/
│   ├── prometheus.yml           # Main Prometheus configuration
│   ├── recording-rules.yml      # Pre-computed aggregations
│   └── alert-rules.yml          # Alert definitions with runbooks
├── grafana/
│   └── dashboards/
│       ├── replication-overview.json      # Operational dashboard
│       ├── infrastructure-metrics.json    # System health dashboard
│       ├── business-metrics.json          # Business KPI dashboard
│       └── error-latency.json             # SRE error budget dashboard
├── alertmanager/
│   └── alertmanager.yml         # Alert routing configuration
├── kubernetes/
│   ├── prometheus-deployment.yaml
│   ├── grafana-deployment.yaml
│   ├── alertmanager-deployment.yaml
│   └── servicemonitor-freightliner.yaml
└── docs/
    ├── MONITORING_SETUP.md      # Detailed setup guide
    └── SLO_RUNBOOKS.md          # Alert runbooks and procedures
```

## Dashboards

### 1. Replication Overview
**Focus**: Real-time replication operations

**Key Panels**:
- Success rate gauge (target: >99%)
- Active replication rate
- Data transfer throughput
- P95/P99 latency charts
- Error breakdown by type
- Top replication pairs table

**Use Case**: Daily operations monitoring

### 2. Infrastructure Metrics
**Focus**: System health and resources

**Key Panels**:
- Service uptime (target: >99.9%)
- Memory usage trend
- Goroutine count
- Worker pool utilization
- HTTP request metrics
- Job execution rates

**Use Case**: Resource planning and capacity management

### 3. Business Metrics
**Focus**: Business KPIs and trends

**Key Panels**:
- Total replications (24h)
- Data transferred (GB)
- Success rate trend
- Estimated costs
- Top repositories by volume
- Traffic distribution

**Use Case**: Business reporting and cost analysis

### 4. Error Rate & Latency Analysis
**Focus**: SRE error budgets and SLO compliance

**Key Panels**:
- Error budget remaining (99.9% SLO)
- Current error rate
- Latency SLO compliance (<60s target)
- Error rate by registry
- Latency heatmap
- Slowest operations

**Use Case**: SLO management and reliability engineering

## Alert Rules

### Critical (PagerDuty + Slack)
- `HighReplicationFailureRate`: >10% failures for 5m
- `ReplicationStopped`: No activity for 10m
- `FreightlinerDown`: Service unreachable for 1m
- `HighHTTP5xxRate`: >1 error/sec for 5m
- `HighMemoryUsage`: >4GB for 5m
- `ApplicationPanics`: Any panic detected

### Warning (Slack)
- `HighReplicationDuration`: P95 >5min for 10m
- `WorkerPoolSaturated`: >90% utilization for 5m
- `HighHTTPLatency`: P95 >5s for 10m
- `HighAuthFailureRate`: >5 failures/sec for 5m

All alerts include:
- Severity classification
- Business impact description
- Actionable remediation steps
- Runbook documentation links

## Metrics Collected

### Replication Metrics
```
freightliner_replications_total{source_registry, dest_registry, status}
freightliner_replication_duration_seconds{source_registry, dest_registry}
freightliner_replication_bytes_total{source_registry, dest_registry}
freightliner_replication_layers_total{source_registry, dest_registry}
freightliner_replication_errors_total{source_registry, dest_registry, error_type}
```

### HTTP Metrics
```
freightliner_http_requests_total{method, path, status}
freightliner_http_request_duration_seconds{method, path, status}
freightliner_http_requests_in_flight
```

### Worker Pool Metrics
```
freightliner_worker_pool_size
freightliner_worker_pool_active
freightliner_worker_pool_queued
```

### System Metrics
```
freightliner_memory_usage_bytes
freightliner_goroutines_count
freightliner_panics_total{component}
```

### Job Metrics
```
freightliner_jobs_total{type, status}
freightliner_job_duration_seconds{type}
freightliner_jobs_active
```

### Authentication Metrics
```
freightliner_auth_failures_total{type}
```

## SLO Definitions

### Availability SLO: 99.9%
- **Error Budget**: 0.1% (43.2 minutes/month)
- **Measurement**: Success ratio of replications
- **Window**: 30 days rolling
- **Alert Threshold**: >10% error rate for 5 minutes

### Latency SLO: 99% < 60s
- **Target**: 99% of replications complete in under 60 seconds
- **Measurement**: P99 latency from histogram
- **Window**: 7 days rolling
- **Alert Threshold**: P95 >300s for 10 minutes

### HTTP API SLO: 99.5%
- **Error Budget**: 0.5% request failures
- **Measurement**: Non-5xx response ratio
- **Window**: 5 minutes
- **Alert Threshold**: >1 error/sec for 5 minutes

## Configuration

### Prometheus Configuration

The main configuration includes:
- **15s scrape interval** for timely metrics
- **Kubernetes service discovery** for dynamic pod scraping
- **Recording rules** for expensive query optimization
- **Alert rules** with severity and runbook links

Key scrape configs:
```yaml
- job_name: 'freightliner'
  kubernetes_sd_configs:
    - role: pod
      namespaces:
        names: [freightliner]
```

### AlertManager Routing

Alerts are routed based on severity:
- **Critical** → PagerDuty (10s wait) + Slack (10s wait)
- **Warning** → Slack (5m wait)
- Component-specific → Dedicated Slack channels

Inhibition rules prevent alert storms:
- Critical alerts suppress warnings
- Service down suppresses component alerts
- Stopped replication suppresses replication errors

### Grafana Provisioning

Automatic setup of:
- Prometheus datasource
- Dashboard folder structure
- Dashboard import from JSON
- User authentication

## Deployment

### Prerequisites

```bash
# Kubernetes cluster with:
- kubectl configured
- 100GB+ storage available
- LoadBalancer or Ingress support (optional)

# Tools required:
- kubectl 1.24+
- curl
- jq (for dashboard management)
```

### Step-by-Step Deployment

1. **Create namespace**:
```bash
kubectl create namespace monitoring
```

2. **Configure secrets**:
```bash
# Grafana admin password
kubectl create secret generic grafana-admin \
  --from-literal=username=admin \
  --from-literal=password=$(openssl rand -base64 32) \
  -n monitoring

# AlertManager secrets
kubectl create secret generic alertmanager-secrets \
  --from-literal=slack-webhook=YOUR_SLACK_WEBHOOK_URL \
  --from-literal=pagerduty-key=YOUR_PAGERDUTY_KEY \
  -n monitoring
```

3. **Update ConfigMaps**:
```bash
# Create Prometheus ConfigMap
kubectl create configmap prometheus-config \
  --from-file=prometheus.yml=monitoring/prometheus/prometheus.yml \
  --from-file=recording-rules.yml=monitoring/prometheus/recording-rules.yml \
  --from-file=alert-rules.yml=monitoring/prometheus/alert-rules.yml \
  -n monitoring

# Create AlertManager ConfigMap
kubectl create configmap alertmanager-config \
  --from-file=alertmanager.yml=monitoring/alertmanager/alertmanager.yml \
  -n monitoring

# Create Grafana dashboards ConfigMap
kubectl create configmap grafana-dashboards \
  --from-file=monitoring/grafana/dashboards/ \
  -n monitoring
```

4. **Deploy services**:
```bash
kubectl apply -f monitoring/kubernetes/prometheus-deployment.yaml
kubectl apply -f monitoring/kubernetes/grafana-deployment.yaml
kubectl apply -f monitoring/kubernetes/alertmanager-deployment.yaml
kubectl apply -f monitoring/kubernetes/servicemonitor-freightliner.yaml
```

5. **Verify deployment**:
```bash
kubectl get pods -n monitoring
kubectl get svc -n monitoring
```

6. **Access dashboards**:
```bash
# Get Grafana password
kubectl get secret grafana-admin -n monitoring -o jsonpath='{.data.password}' | base64 -d

# Port forward
kubectl port-forward -n monitoring svc/grafana 3000:3000
# Open http://localhost:3000
```

## Customization

### Adding Custom Metrics

1. Instrument code with Prometheus client:
```go
myMetric := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "freightliner_custom_metric_total",
        Help: "Description of metric",
    },
    []string{"label1", "label2"},
)
```

2. Register with registry:
```go
registry.MustRegister(myMetric)
```

3. Use in code:
```go
myMetric.WithLabelValues("value1", "value2").Inc()
```

### Adding Custom Alerts

1. Edit `monitoring/prometheus/alert-rules.yml`:
```yaml
- alert: MyCustomAlert
  expr: my_metric > threshold
  for: 5m
  labels:
    severity: warning
    component: my-component
  annotations:
    summary: "Custom alert fired"
    description: "Detailed description"
```

2. Update ConfigMap:
```bash
kubectl create configmap prometheus-config \
  --from-file=monitoring/prometheus/ \
  -n monitoring \
  --dry-run=client -o yaml | kubectl apply -f -
```

3. Reload Prometheus:
```bash
kubectl exec -n monitoring prometheus-0 -- curl -X POST http://localhost:9090/-/reload
```

### Creating Custom Dashboards

1. Design in Grafana UI
2. Export JSON via Share → Export
3. Save to `monitoring/grafana/dashboards/`
4. Update ConfigMap:
```bash
kubectl create configmap grafana-dashboards \
  --from-file=monitoring/grafana/dashboards/ \
  -n monitoring \
  --dry-run=client -o yaml | kubectl apply -f -
```

## Scaling Considerations

### Prometheus Scaling

**Current capacity**: ~100k samples/sec

For larger deployments:
- **Federation**: Shard Prometheus by namespace or service
- **Remote Storage**: Thanos or Cortex for long-term retention
- **Recording Rules**: Pre-compute expensive queries
- **Metric Relabeling**: Drop unnecessary metrics

### High Availability

Deploy multiple replicas:

```yaml
# Prometheus StatefulSet with 2 replicas
replicas: 2

# AlertManager with clustering
replicas: 3
args:
  - '--cluster.peer=alertmanager-0.alertmanager:9094'
  - '--cluster.peer=alertmanager-1.alertmanager:9094'
  - '--cluster.peer=alertmanager-2.alertmanager:9094'
```

### Storage Planning

**Prometheus storage formula**:
```
disk_space = retention_time * ingestion_rate * 2 bytes/sample
```

Example:
- 30 days retention
- 50k samples/sec
- = 30 * 86400 * 50000 * 2 = 259GB

Add 30% buffer = **337GB recommended**

## Troubleshooting

### Metrics Not Appearing

1. Check Prometheus targets:
   - Visit http://localhost:9090/targets
   - Look for red/down targets
   - Check error messages

2. Verify ServiceMonitor:
```bash
kubectl get servicemonitor -A
kubectl describe servicemonitor freightliner -n freightliner
```

3. Test metrics endpoint directly:
```bash
kubectl port-forward -n freightliner svc/freightliner 2112:2112
curl http://localhost:2112/metrics | grep freightliner
```

### Alerts Not Firing

1. Check alert rules syntax:
   - Visit http://localhost:9090/alerts
   - Look for "invalid" status
   - Review error messages

2. Test alert expressions:
   - Run query in Prometheus UI
   - Verify data exists
   - Check for duration requirement

3. Verify AlertManager configuration:
```bash
kubectl logs -n monitoring deployment/alertmanager
# Look for config errors
```

### Dashboard Issues

1. Check Grafana logs:
```bash
kubectl logs -n monitoring deployment/grafana
```

2. Test Prometheus datasource:
   - Grafana → Configuration → Data Sources
   - Click "Test" button
   - Verify connection succeeds

3. Re-import dashboard:
   - Dashboards → Import
   - Upload JSON from `monitoring/grafana/dashboards/`

### High Prometheus Memory Usage

1. Check cardinality:
```promql
count by (__name__) ({__name__=~".+"})
```

2. Identify high-cardinality metrics:
```bash
curl http://localhost:9090/api/v1/status/tsdb | jq '.data.seriesCountByMetricName'
```

3. Solutions:
   - Drop unnecessary labels
   - Increase memory limits
   - Reduce retention time
   - Use recording rules

## Maintenance

### Backup Procedures

**Prometheus data**:
```bash
kubectl exec -n monitoring prometheus-0 -- \
  tar czf /tmp/backup.tar.gz /prometheus
kubectl cp monitoring/prometheus-0:/tmp/backup.tar.gz ./prometheus-backup.tar.gz
```

**Grafana dashboards**:
```bash
# Export via API
GRAFANA_API_KEY="your-api-key"
curl -H "Authorization: Bearer $GRAFANA_API_KEY" \
  http://localhost:3000/api/search | \
  jq -r '.[] | .uid' | \
  xargs -I {} curl -H "Authorization: Bearer $GRAFANA_API_KEY" \
  http://localhost:3000/api/dashboards/uid/{} > backup-{}.json
```

### Updating Alert Rules

```bash
# 1. Edit alert-rules.yml locally
vim monitoring/prometheus/alert-rules.yml

# 2. Update ConfigMap
kubectl create configmap prometheus-config \
  --from-file=monitoring/prometheus/ \
  -n monitoring \
  --dry-run=client -o yaml | kubectl apply -f -

# 3. Reload Prometheus (if lifecycle API enabled)
kubectl exec -n monitoring prometheus-0 -- \
  curl -X POST http://localhost:9090/-/reload
```

### Upgrading Components

```bash
# Update image versions in deployment files
vim monitoring/kubernetes/prometheus-deployment.yaml

# Apply changes
kubectl apply -f monitoring/kubernetes/prometheus-deployment.yaml

# Watch rollout
kubectl rollout status deployment/prometheus -n monitoring
```

## Best Practices

### Alert Design
- Set appropriate thresholds based on SLOs
- Include actionable remediation steps
- Add runbook links to all alerts
- Use inhibition to prevent alert storms
- Regular alert tuning based on feedback

### Dashboard Design
- Keep dashboards focused (one purpose per dashboard)
- Use variables for filtering
- Show both current state and trends
- Include SLO targets as reference lines
- Optimize queries with recording rules

### SLO Management
- Review SLOs quarterly with stakeholders
- Track error budget burn rate
- Use error budgets for prioritization decisions
- Document SLO rationale and measurement

### Capacity Planning
- Monitor Prometheus storage growth weekly
- Track query performance monthly
- Plan scaling 3 months ahead
- Review retention policies quarterly

### Security
- Use strong passwords (not 'changeme'!)
- Enable authentication for all services
- Encrypt sensitive configuration (Sealed Secrets)
- Restrict network access with NetworkPolicies
- Regular security updates

## Documentation

- **Setup Guide**: [monitoring/docs/MONITORING_SETUP.md](docs/MONITORING_SETUP.md)
- **SLO Runbooks**: [monitoring/docs/SLO_RUNBOOKS.md](docs/SLO_RUNBOOKS.md)
- **Prometheus Docs**: https://prometheus.io/docs/
- **Grafana Docs**: https://grafana.com/docs/
- **AlertManager Docs**: https://prometheus.io/docs/alerting/alertmanager/

## Support

- **Slack**: #freightliner-monitoring
- **PagerDuty**: Freightliner on-call rotation
- **Wiki**: https://wiki.company.com/freightliner/monitoring
- **Issues**: File tickets in Jira (FREIGHT project)

## License

Internal use only. See company licensing policy.
