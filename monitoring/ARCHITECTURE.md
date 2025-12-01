# Freightliner Monitoring Architecture

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Kubernetes Cluster                              │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                    Freightliner Namespace                         │  │
│  │                                                                   │  │
│  │  ┌──────────────────┐     ┌──────────────────┐                  │  │
│  │  │  Freightliner    │     │  Freightliner    │                  │  │
│  │  │     Pod 1        │     │     Pod 2        │                  │  │
│  │  │                  │     │                  │                  │  │
│  │  │  :8080 (HTTP)    │     │  :8080 (HTTP)    │                  │  │
│  │  │  :2112 (Metrics) │     │  :2112 (Metrics) │                  │  │
│  │  └────────┬─────────┘     └────────┬─────────┘                  │  │
│  │           │                        │                            │  │
│  │           └────────────┬───────────┘                            │  │
│  │                        │                                        │  │
│  │                        │ metrics scrape                         │  │
│  │                        │ (every 15s)                           │  │
│  └────────────────────────┼───────────────────────────────────────┘  │
│                           │                                           │
│  ┌────────────────────────┼───────────────────────────────────────┐  │
│  │                        │     Monitoring Namespace               │  │
│  │                        ▼                                        │  │
│  │              ┌──────────────────┐                               │  │
│  │              │   Prometheus     │                               │  │
│  │              │     :9090        │                               │  │
│  │              │                  │                               │  │
│  │              │ - Time series DB │                               │  │
│  │              │ - Rule engine    │                               │  │
│  │              │ - Alert manager  │                               │  │
│  │              │                  │                               │  │
│  │              │ Storage: 100GB   │                               │  │
│  │              │ Retention: 30d   │                               │  │
│  │              └─────┬────────┬───┘                               │  │
│  │                    │        │                                   │  │
│  │        ┌───────────┘        └──────────┐                        │  │
│  │        │ query                alerts   │                        │  │
│  │        ▼                               ▼                        │  │
│  │  ┌─────────────┐              ┌──────────────┐                 │  │
│  │  │   Grafana   │              │ AlertManager │                 │  │
│  │  │    :3000    │              │    :9093     │                 │  │
│  │  │             │              │              │                 │  │
│  │  │ - Dashboards│              │ - Routing    │                 │  │
│  │  │ - Viz engine│              │ - Grouping   │                 │  │
│  │  │ - Auth      │              │ - Throttling │                 │  │
│  │  │             │              │ - Silences   │                 │  │
│  │  │ Storage:10GB│              │ Storage: 5GB │                 │  │
│  │  └─────────────┘              └──────┬───────┘                 │  │
│  │                                      │                         │  │
│  └──────────────────────────────────────┼─────────────────────────┘  │
│                                         │                            │
└─────────────────────────────────────────┼────────────────────────────┘
                                          │
                                          │ notifications
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
                    ▼                     ▼                     ▼
          ┌─────────────────┐   ┌──────────────┐    ┌──────────────┐
          │      Slack      │   │  PagerDuty   │    │    Email     │
          │                 │   │              │    │              │
          │ - #critical     │   │ - On-call    │    │ - Reports    │
          │ - #warnings     │   │ - Escalation │    │ - Summaries  │
          │ - #replication  │   │ - Ack/Resolve│    │              │
          │ - #api          │   │              │    │              │
          │ - #system       │   │              │    │              │
          └─────────────────┘   └──────────────┘    └──────────────┘
```

## Component Relationships

### Data Flow

1. **Metrics Collection** (Prometheus ← Freightliner)
   - Prometheus scrapes `/metrics` endpoint every 15 seconds
   - ServiceMonitor uses Kubernetes service discovery
   - Metrics stored in time-series database with labels
   - Retention: 30 days (configurable)

2. **Recording Rules** (Prometheus)
   - Pre-compute expensive queries every 30-60 seconds
   - Store as new time series for dashboard performance
   - Examples: success_rate:5m, duration:p95:5m

3. **Alert Evaluation** (Prometheus)
   - Evaluate alert rules every 30 seconds
   - Track alert state (Pending → Firing → Resolved)
   - Send firing alerts to AlertManager

4. **Alert Routing** (AlertManager)
   - Receive alerts from Prometheus
   - Group related alerts (by alertname, cluster, service)
   - Apply inhibition rules (suppress redundant alerts)
   - Route to appropriate receivers (Slack/PagerDuty)
   - Throttle repeat notifications

5. **Visualization** (Grafana ← Prometheus)
   - Query Prometheus for metrics
   - Render dashboards with real-time data
   - Use recording rules for performance
   - Provide interactive exploration

6. **Notification** (Slack/PagerDuty ← AlertManager)
   - Receive formatted alert messages
   - Display in appropriate channels
   - Enable acknowledgment and resolution
   - Track escalation

## Metrics Architecture

### Metric Types Used

1. **Counters** (monotonically increasing)
   - `freightliner_replications_total`
   - `freightliner_http_requests_total`
   - `freightliner_replication_errors_total`
   - Usage: Rates and increases

2. **Histograms** (distribution of observations)
   - `freightliner_replication_duration_seconds`
   - `freightliner_http_request_duration_seconds`
   - `freightliner_tag_copy_duration_seconds`
   - Usage: Percentiles (P50, P95, P99)

3. **Gauges** (can go up or down)
   - `freightliner_memory_usage_bytes`
   - `freightliner_goroutines_count`
   - `freightliner_worker_pool_active`
   - Usage: Current values and trends

### Label Strategy

**Good Labels** (low cardinality):
- `source_registry`: ~10 unique values
- `dest_registry`: ~10 unique values
- `status`: 3-5 values (success, failed, skipped)
- `method`: 5-10 HTTP methods
- `component`: 10-20 components

**Avoid** (high cardinality):
- User IDs or session IDs
- Image digests or tags
- Timestamps
- Dynamic paths with IDs

### Cardinality Management

```
Total Time Series = 
  Metric Family × 
  Label Value Combinations

Example:
freightliner_replications_total{
  source_registry(10) × 
  dest_registry(10) × 
  status(3)
} = 300 time series
```

**Best Practices**:
- Keep label values < 100 per label
- Monitor cardinality with `count by (__name__)`
- Use recording rules to reduce high-cardinality queries
- Drop unnecessary labels with `metric_relabel_configs`

## Alert Architecture

### Alert Severity Levels

```
Critical (P0)
    │
    ├─ PagerDuty (immediate page)
    │   - 24/7 on-call response
    │   - <15 minute SLA
    │   - Escalation after 30 min
    │
    └─ Slack #freightliner-critical
        - Immediate notification
        - Red/rotating light emoji
        - Full context + runbook

Warning (P2)
    │
    └─ Slack #freightliner-warnings
        - 5-minute group wait
        - Yellow warning emoji
        - Context + action items
```

### Alert State Machine

```
      ┌──────────────┐
      │   Inactive   │  (no alert)
      └──────┬───────┘
             │ expr becomes true
             ▼
      ┌──────────────┐
      │   Pending    │  (waiting for duration)
      └──────┬───────┘
             │ duration exceeded
             ▼
      ┌──────────────┐
      │    Firing    │─────► AlertManager ─────► Notifications
      └──────┬───────┘
             │ expr becomes false
             ▼
      ┌──────────────┐
      │   Resolved   │─────► Resolution notification
      └──────────────┘
```

### Inhibition Rules

Prevent alert storms by suppressing child alerts when parent alerts fire:

```
FreightlinerDown (firing)
    │
    └─ inhibits ─► All component alerts
                   - HighReplicationFailureRate
                   - HighHTTP5xxRate
                   - WorkerPoolSaturated
                   - etc.

ReplicationStopped (firing)
    │
    └─ inhibits ─► Replication component alerts
                   - HighReplicationDuration
                   - HighReplicationErrorRate

Critical (firing)
    │
    └─ inhibits ─► Warnings for same component
```

## Dashboard Architecture

### Dashboard Hierarchy

```
Dashboards (4 total)
│
├─ Replication Overview (Operational)
│  ├─ For: DevOps, SRE, On-call
│  ├─ Focus: Real-time operations
│  └─ Refresh: 30s
│
├─ Infrastructure Metrics (System Health)
│  ├─ For: Platform engineers, SRE
│  ├─ Focus: Resource utilization
│  └─ Refresh: 30s
│
├─ Business Metrics (KPIs)
│  ├─ For: Management, Product
│  ├─ Focus: Business outcomes
│  └─ Refresh: 1m
│
└─ Error Rate & Latency (SRE)
   ├─ For: SRE, Engineering leads
   ├─ Focus: SLO compliance
   └─ Refresh: 30s
```

### Panel Types Used

1. **Stat Panels** - Single value with threshold
   - Success rate
   - Current error rate
   - Memory usage
   - SLO compliance

2. **Time Series** - Trends over time
   - Replication rate
   - Latency percentiles
   - Error rate
   - Resource usage

3. **Tables** - Sorted/filtered data
   - Top repositories
   - Slowest operations
   - Error breakdown

4. **Heatmaps** - Distribution visualization
   - Latency distribution
   - Request patterns

5. **Pie Charts** - Proportional distribution
   - Status distribution
   - Traffic by registry
   - Error types

6. **Gauges** - Progress towards threshold
   - Active jobs
   - Queue depth
   - Pool utilization

## Deployment Architecture

### Kubernetes Resources

```
monitoring namespace
│
├─ ConfigMaps (3)
│  ├─ prometheus-config
│  │  ├─ prometheus.yml
│  │  ├─ recording-rules.yml
│  │  └─ alert-rules.yml
│  │
│  ├─ alertmanager-config
│  │  └─ alertmanager.yml
│  │
│  └─ grafana-dashboards
│     ├─ replication-overview.json
│     ├─ infrastructure-metrics.json
│     ├─ business-metrics.json
│     └─ error-latency.json
│
├─ Secrets (2)
│  ├─ grafana-admin
│  │  ├─ username
│  │  └─ password
│  │
│  └─ alertmanager-secrets
│     ├─ slack-webhook
│     └─ pagerduty-key
│
├─ PersistentVolumeClaims (3)
│  ├─ prometheus-data (100GB)
│  ├─ grafana-data (10GB)
│  └─ alertmanager-data (5GB)
│
├─ Deployments (3)
│  ├─ prometheus (1 replica)
│  ├─ grafana (1 replica)
│  └─ alertmanager (1 replica)
│
├─ Services (6)
│  ├─ prometheus (ClusterIP :9090)
│  ├─ prometheus-external (LoadBalancer)
│  ├─ grafana (ClusterIP :3000)
│  ├─ grafana-external (LoadBalancer)
│  ├─ alertmanager (ClusterIP :9093)
│  └─ alertmanager-external (LoadBalancer)
│
└─ RBAC (3)
   ├─ ServiceAccount: prometheus
   ├─ ClusterRole: prometheus
   └─ ClusterRoleBinding: prometheus
```

### High Availability Options

For production HA (not implemented by default):

```
Prometheus HA
├─ 2+ replicas with same config
├─ Load balancer across replicas
├─ Each replica scrapes all targets
└─ Deduplicated in Grafana or downstream

AlertManager HA
├─ 3+ replicas in cluster mode
├─ Gossip protocol for sync
├─ Any replica can receive alerts
└─ Coordinated deduplication

Grafana HA
├─ 2+ replicas
├─ Shared PostgreSQL backend
├─ Session stored in database
└─ Load balanced
```

## Security Architecture

### Network Security

```
Internet
    │
    ├─ Ingress/LoadBalancer (optional)
    │     │
    │     └─ TLS termination
    │
    ▼
Grafana (3000)
    │
    ├─ Basic Auth / OAuth
    │
    └──► Prometheus (9090)
          │
          ├─ Internal only (ClusterIP)
          │
          └──► Freightliner (:2112)
                  │
                  └─ Internal only

AlertManager (9093)
    │
    ├─ Internal only (ClusterIP)
    │
    └──► External APIs (HTTPS)
          ├─ Slack webhooks
          └─ PagerDuty API
```

### Authentication & Authorization

1. **Grafana**
   - Basic auth (default)
   - OAuth2/LDAP (recommended for production)
   - Role-based access control
   - API keys for automation

2. **Prometheus**
   - No auth by default
   - Recommend: Reverse proxy with basic auth
   - Or: NetworkPolicy to restrict access
   - OAuth2 Proxy for production

3. **AlertManager**
   - No auth by default
   - Internal-only access via ClusterIP
   - API auth via reverse proxy

### Secret Management

**Current Approach**:
- Kubernetes Secrets (base64 encoded)
- Mounted as files or env vars

**Production Recommendations**:
- Sealed Secrets for GitOps
- External Secrets Operator
- Vault integration
- Cloud provider secret managers (AWS Secrets Manager, etc.)

## Scaling Architecture

### Vertical Scaling

```
Current → Recommended for growth

Prometheus:
  CPU: 500m → 2000m → 4000m
  Memory: 2Gi → 8Gi → 16Gi
  Storage: 100GB → 500GB → 1TB

Grafana:
  CPU: 250m → 1000m → 2000m
  Memory: 512Mi → 2Gi → 4Gi

AlertManager:
  CPU: 100m → 500m → 1000m
  Memory: 128Mi → 512Mi → 1Gi
```

### Horizontal Scaling

**When to scale horizontally**:
- Prometheus ingestion > 1M samples/sec
- Grafana concurrent users > 100
- AlertManager alert rate > 1000/min

**Approaches**:

1. **Prometheus Federation**
   ```
   Central Prometheus
        │
        ├─ scrapes from ─► Regional Prometheus 1
        │                  (scrapes local services)
        │
        └─ scrapes from ─► Regional Prometheus 2
                           (scrapes local services)
   ```

2. **Prometheus Sharding**
   - Shard by namespace
   - Shard by service
   - Use Thanos for global view

3. **Grafana Scaling**
   - Multiple replicas behind load balancer
   - Shared PostgreSQL backend
   - Redis for session storage

## Storage Architecture

### Prometheus Storage

```
Time Series Database (TSDB)
│
├─ Write-Ahead Log (WAL)
│  └─ Recent data (last 2 hours)
│
└─ Blocks (immutable)
   ├─ 2h blocks (recent)
   └─ Compacted blocks (older)
       └─ Retention: 30 days
```

**Storage Formula**:
```
size = retention × ingestion_rate × 2 bytes/sample

Example:
  30 days × 50k samples/sec × 2 bytes
  = 30 × 86400 × 50000 × 2
  = 259.2 GB
  + 30% buffer = 337 GB recommended
```

### Long-term Storage Options

For >30 days retention:

1. **Thanos** (recommended)
   - S3-compatible object storage
   - Unlimited retention
   - Global query view
   - Compaction for efficiency

2. **Cortex**
   - Multi-tenant capable
   - Horizontal scalability
   - Cloud-native architecture

3. **Victoria Metrics**
   - High compression
   - Fast queries
   - Low resource usage

## Disaster Recovery

### Backup Strategy

```
Daily Backups
│
├─ Prometheus Data
│  ├─ Snapshot via API
│  ├─ Copy to S3/GCS
│  └─ Retention: 7 days
│
├─ Grafana
│  ├─ Dashboard export via API
│  ├─ Database backup (if external DB)
│  └─ Version control in Git
│
└─ AlertManager
   ├─ Config in Git (IaC)
   └─ Silence/notification state (optional)
```

### Recovery Procedures

**RTO (Recovery Time Objective)**: < 1 hour
**RPO (Recovery Point Objective)**: < 24 hours

1. **Prometheus Failure**
   - New pod starts automatically (StatefulSet)
   - Data recovered from PVC
   - Loss: Only WAL data (< 2 hours)

2. **Grafana Failure**
   - New pod starts automatically
   - Dashboards re-imported from ConfigMap
   - Loss: None (stateless)

3. **PVC Failure**
   - Restore from S3/GCS backup
   - Re-attach to pod
   - Loss: Data since last backup

## Performance Optimization

### Query Optimization

1. **Use Recording Rules**
   ```promql
   # Before (expensive):
   histogram_quantile(0.95,
     sum by (le) (
       rate(freightliner_replication_duration_seconds_bucket[5m])
     )
   )

   # After (pre-computed):
   freightliner:replication:duration:p95:5m
   ```

2. **Limit Time Range**
   - Dashboard default: 1 hour
   - Increase only when needed
   - Use aggregated data for long ranges

3. **Reduce Cardinality**
   - Drop unnecessary labels
   - Aggregate high-cardinality metrics
   - Use metric_relabel_configs

### Dashboard Optimization

1. **Panel Query Optimization**
   - Use recording rules
   - Minimize time range
   - Limit series with filters

2. **Refresh Rate**
   - Operational: 30s
   - Business: 1-5m
   - Historical: Manual

3. **Panel Count**
   - Max 12-15 panels per dashboard
   - Use rows to organize
   - Lazy-load panels

## Monitoring the Monitoring

### Meta-Metrics

Monitor the monitoring stack itself:

```promql
# Prometheus health
up{job="prometheus"} == 1
prometheus_tsdb_storage_blocks_bytes

# Grafana health
up{job="grafana"} == 1

# AlertManager health
up{job="alertmanager"} == 1
alertmanager_notifications_failed_total

# Scrape failures
up == 0
prometheus_target_scrapes_exceeded_sample_limit_total > 0
```

### Meta-Alerts

```yaml
- alert: PrometheusScrapeErrors
  expr: increase(prometheus_target_scrapes_exceeded_sample_limit_total[5m]) > 0
  
- alert: AlertManagerNotificationFailed
  expr: increase(alertmanager_notifications_failed_total[5m]) > 0
  
- alert: MonitoringStackDown
  expr: up{job=~"prometheus|grafana|alertmanager"} == 0
```

## Conclusion

This architecture provides:

- **Reliability**: Persistent storage, health checks, auto-recovery
- **Scalability**: Vertical and horizontal scaling paths
- **Performance**: Recording rules, query optimization, caching
- **Security**: RBAC, network policies, secret management
- **Observability**: Meta-monitoring, comprehensive metrics
- **Maintainability**: IaC, documentation, automation

The stack is production-ready and follows SRE best practices for managing observability at scale.
