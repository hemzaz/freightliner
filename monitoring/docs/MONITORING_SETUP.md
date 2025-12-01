# Freightliner Monitoring Setup Guide

## Overview

This monitoring stack provides comprehensive observability for the Freightliner container replication service using:

- **Prometheus**: Metrics collection and alerting
- **Grafana**: Visualization and dashboards
- **AlertManager**: Alert routing and notification management

## Architecture

```
┌─────────────────┐
│  Freightliner   │ :2112/metrics
│   Application   │────────┐
└─────────────────┘        │
                           │
┌─────────────────┐        │        ┌─────────────────┐
│   Kubernetes    │        ├───────▶│   Prometheus    │
│   Pods/Nodes    │────────┤        │    :9090        │
└─────────────────┘        │        └────────┬────────┘
                           │                 │
┌─────────────────┐        │                 │ alerts
│  kube-state-    │────────┘                 │
│    metrics      │                          ▼
└─────────────────┘                 ┌─────────────────┐
                                    │  AlertManager   │
┌─────────────────┐                 │    :9093        │
│    Grafana      │◀────────────────┴─────────────────┘
│     :3000       │                   notifications
└─────────────────┘                          │
        │                                    ▼
        │                          ┌──────────────────┐
        └─────────────────────────▶│ Slack / PagerDuty│
           dashboards              └──────────────────┘
```

## Quick Start

### 1. Deploy to Kubernetes

```bash
# Create monitoring namespace
kubectl create namespace monitoring

# Deploy Prometheus
kubectl apply -f monitoring/kubernetes/prometheus-deployment.yaml

# Deploy Grafana
kubectl apply -f monitoring/kubernetes/grafana-deployment.yaml

# Deploy AlertManager
kubectl apply -f monitoring/kubernetes/alertmanager-deployment.yaml

# Deploy ServiceMonitors for Freightliner
kubectl apply -f monitoring/kubernetes/servicemonitor-freightliner.yaml
```

### 2. Configure Secrets

```bash
# Create Grafana admin credentials
kubectl create secret generic grafana-admin \
  --from-literal=username=admin \
  --from-literal=password=YOUR_SECURE_PASSWORD \
  -n monitoring

# Create AlertManager secrets (if using external services)
kubectl create secret generic alertmanager-secrets \
  --from-literal=slack-webhook=YOUR_SLACK_WEBHOOK \
  --from-literal=pagerduty-key=YOUR_PAGERDUTY_KEY \
  -n monitoring
```

### 3. Access Services

```bash
# Port-forward Prometheus
kubectl port-forward -n monitoring svc/prometheus 9090:9090

# Port-forward Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000

# Port-forward AlertManager
kubectl port-forward -n monitoring svc/alertmanager 9093:9093
```

Access URLs:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/YOUR_SECURE_PASSWORD)
- AlertManager: http://localhost:9093

### 4. Import Grafana Dashboards

1. Log into Grafana at http://localhost:3000
2. Navigate to Dashboards → Import
3. Upload or paste the JSON from:
   - `monitoring/grafana/dashboards/replication-overview.json`
   - `monitoring/grafana/dashboards/infrastructure-metrics.json`
   - `monitoring/grafana/dashboards/business-metrics.json`
   - `monitoring/grafana/dashboards/error-latency.json`

## Available Dashboards

### 1. Replication Overview
**Purpose**: Monitor container replication operations

**Key Metrics**:
- Replication success rate
- Active replication operations
- Data transfer rates
- P95/P99 latency
- Error rates by type
- Top replication pairs

**Use Cases**:
- Operational monitoring
- Performance analysis
- Capacity planning

### 2. Infrastructure Metrics
**Purpose**: Monitor system health and resource utilization

**Key Metrics**:
- Service uptime
- Memory usage
- Goroutine count
- Worker pool utilization
- HTTP request rates
- Job execution rates
- Authentication failures

**Use Cases**:
- Resource planning
- Performance tuning
- Capacity management

### 3. Business Metrics
**Purpose**: Track business KPIs and trends

**Key Metrics**:
- Total replications per day/hour
- Data transferred (GB)
- Success rates
- Estimated transfer costs
- Top repositories by usage
- Traffic distribution

**Use Cases**:
- Business reporting
- Cost analysis
- Usage patterns

### 4. Error Rate & Latency Analysis
**Purpose**: SRE-focused error budget and latency tracking

**Key Metrics**:
- Error budget remaining (99.9% SLO)
- Current error rates
- Latency SLO compliance
- Error rate by registry pair
- P50/P90/P95/P99 latency
- Error budget burn rate
- Slowest operations

**Use Cases**:
- SLO monitoring
- Error budget management
- Performance optimization

## Alert Rules

### Critical Alerts (PagerDuty + Slack)

1. **HighReplicationFailureRate**
   - Threshold: >10% failure rate for 5 minutes
   - Impact: Container images not replicating
   - Action: Check registry connectivity and authentication

2. **ReplicationStopped**
   - Threshold: No replication activity for 10 minutes
   - Impact: Complete service outage
   - Action: Check service health and job scheduler

3. **FreightlinerDown**
   - Threshold: Service unreachable for 1 minute
   - Impact: Complete service unavailability
   - Action: Check pod status and restart if needed

4. **HighHTTP5xxRate**
   - Threshold: >1 error/sec for 5 minutes
   - Impact: API experiencing server errors
   - Action: Check application logs and database

5. **HighMemoryUsage**
   - Threshold: >4GB for 5 minutes
   - Impact: Potential OOM kills
   - Action: Check for memory leaks

6. **ApplicationPanics**
   - Threshold: Any panic detected
   - Impact: Service instability
   - Action: Review panic stack traces

### Warning Alerts (Slack)

1. **HighReplicationDuration**
   - Threshold: P95 >5 minutes for 10 minutes
   - Impact: Slow replication
   - Action: Check network and registry performance

2. **WorkerPoolSaturated**
   - Threshold: >90% utilization for 5 minutes
   - Impact: Job queuing and delays
   - Action: Consider scaling up workers

3. **HighHTTPLatency**
   - Threshold: P95 >5s for 10 minutes
   - Impact: Slow API responses
   - Action: Check for slow queries

## Configuration

### Prometheus Configuration

Edit `monitoring/prometheus/prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'freightliner'
    static_configs:
      - targets:
          - 'freightliner.freightliner.svc.cluster.local:2112'
```

### Alert Rules

Alert rules are in `monitoring/prometheus/alert-rules.yml`:
- Organized by component (replication, http, worker_pool, system, auth)
- Include runbook links
- Configured with appropriate severity levels

### Recording Rules

Recording rules in `monitoring/prometheus/recording-rules.yml`:
- Pre-compute complex queries
- Improve dashboard performance
- Calculate SLIs and SLOs

### AlertManager Configuration

Edit `monitoring/alertmanager/alertmanager.yml`:

1. Configure Slack webhook:
```yaml
slack_api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
```

2. Configure PagerDuty:
```yaml
pagerduty_configs:
  - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
```

3. Set up routing rules for different alert types

## SLO Definitions

### Availability SLO: 99.9% (Three Nines)
- **Error Budget**: 0.1% (43.2 minutes/month)
- **Measurement**: `freightliner_replications_total{status="success"} / freightliner_replications_total`
- **Window**: 30 days

### Latency SLO: 99% under 60 seconds
- **Threshold**: P99 < 60s
- **Measurement**: `histogram_quantile(0.99, freightliner_replication_duration_seconds_bucket)`
- **Window**: 7 days

## Runbooks

All alerts include runbook links pointing to: `https://wiki.company.com/runbooks/freightliner/<alert-name>`

Create runbooks for each alert with:
1. **Symptoms**: What the alert indicates
2. **Impact**: Business impact
3. **Diagnosis**: How to investigate
4. **Resolution**: Steps to fix
5. **Prevention**: How to avoid in future

## Retention and Storage

### Prometheus
- **Retention Time**: 30 days
- **Retention Size**: 90GB
- **Storage**: 100GB PVC
- **Scrape Interval**: 15s

### Grafana
- **Storage**: 10GB PVC
- **Backup**: Configure snapshot backups

### AlertManager
- **Data Retention**: 120 hours
- **Storage**: 5GB PVC

## Scaling Considerations

### Prometheus Scaling
- For >100k samples/sec, consider Prometheus federation
- Use recording rules to reduce query load
- Consider remote storage (Thanos, Cortex) for long-term retention

### Grafana Scaling
- Single instance sufficient for <100 users
- Configure external database (PostgreSQL) for HA
- Use provisioning for dashboard management

### AlertManager Scaling
- Deploy 3 replicas for HA
- Configure clustering for gossip protocol

## Troubleshooting

### Metrics Not Appearing

1. Check Prometheus targets:
```bash
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# Visit http://localhost:9090/targets
```

2. Verify ServiceMonitor:
```bash
kubectl get servicemonitor -n freightliner
kubectl describe servicemonitor freightliner -n freightliner
```

3. Check Freightliner metrics endpoint:
```bash
kubectl port-forward -n freightliner svc/freightliner 2112:2112
curl http://localhost:2112/metrics
```

### Alerts Not Firing

1. Check alert rules:
```bash
# Visit http://localhost:9090/alerts in Prometheus
```

2. Verify AlertManager config:
```bash
kubectl logs -n monitoring deployment/alertmanager
```

3. Test notification channels:
```bash
kubectl port-forward -n monitoring svc/alertmanager 9093:9093
# Visit http://localhost:9093/#/status
```

### Dashboard Issues

1. Check Grafana logs:
```bash
kubectl logs -n monitoring deployment/grafana
```

2. Verify datasource connection:
   - Grafana UI → Configuration → Data Sources
   - Test connection to Prometheus

3. Re-import dashboards if needed

## Maintenance

### Backup Prometheus Data
```bash
kubectl exec -n monitoring prometheus-0 -- tar czf /tmp/prometheus-backup.tar.gz /prometheus
kubectl cp monitoring/prometheus-0:/tmp/prometheus-backup.tar.gz ./prometheus-backup.tar.gz
```

### Backup Grafana Dashboards
```bash
# Export all dashboards via API
curl -H "Authorization: Bearer $GRAFANA_API_KEY" \
  http://localhost:3000/api/search | \
  jq -r '.[] | .uri' | \
  xargs -I {} curl -H "Authorization: Bearer $GRAFANA_API_KEY" \
  http://localhost:3000/api/dashboards/{} > dashboards-backup.json
```

### Update Alert Rules
```bash
# Edit alert-rules.yml
kubectl create configmap prometheus-config \
  --from-file=monitoring/prometheus/ \
  -n monitoring \
  --dry-run=client -o yaml | kubectl apply -f -

# Reload Prometheus config
kubectl exec -n monitoring prometheus-0 -- curl -X POST http://localhost:9090/-/reload
```

## Best Practices

1. **Alert Fatigue Prevention**
   - Set appropriate thresholds
   - Use inhibition rules
   - Group related alerts
   - Regular alert tuning

2. **Dashboard Organization**
   - Separate operational vs business dashboards
   - Use variables for filtering
   - Keep dashboards focused and simple

3. **SLO Management**
   - Review SLOs quarterly
   - Track error budget burn rate
   - Use error budgets for prioritization

4. **Capacity Planning**
   - Monitor storage growth
   - Track query performance
   - Scale before hitting limits

5. **Security**
   - Use authentication for all services
   - Encrypt sensitive data (Slack webhooks, PagerDuty keys)
   - Restrict network access
   - Regular security updates

## Support

For issues or questions:
- Internal Wiki: https://wiki.company.com/freightliner/monitoring
- Slack: #freightliner-monitoring
- On-call: PagerDuty escalation policy
