# Freightliner Docker Monitoring Stack

Complete local development monitoring stack with Prometheus, Grafana, and Freightliner API.

## Quick Start

### 1. Basic Monitoring Stack (Recommended)

Start Prometheus, Grafana, and Freightliner API:

```bash
# Start the monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d

# View logs
docker-compose -f docker-compose.monitoring.yml logs -f

# Stop the stack
docker-compose -f docker-compose.monitoring.yml down
```

**Services Started:**
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Freightliner API: http://localhost:8080
- Metrics Endpoint: http://localhost:2112/metrics

### 2. Full Stack with System Metrics

Start all services including Node Exporter, cAdvisor, AlertManager, and Redis:

```bash
# Start with full profile
docker-compose -f docker-compose.monitoring.yml --profile full up -d

# Or using environment variable
COMPOSE_PROFILES=full docker-compose -f docker-compose.monitoring.yml up -d
```

**Additional Services:**
- AlertManager: http://localhost:9093
- Node Exporter: http://localhost:9100
- cAdvisor: http://localhost:8081
- Redis: localhost:6379

## Configuration

### Environment Variables

1. Copy the example environment file:
```bash
cp .env.monitoring .env
```

2. Edit `.env` with your settings:
```bash
# Example customizations
GRAFANA_ADMIN_PASSWORD=secure_password
PROMETHEUS_RETENTION=60d
WORKER_POOL_SIZE=20
LOG_LEVEL=debug
```

### Key Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `GRAFANA_PORT` | 3000 | Grafana UI port |
| `PROMETHEUS_PORT` | 9090 | Prometheus UI port |
| `API_PORT` | 8080 | Freightliner API port |
| `METRICS_PORT` | 2112 | Prometheus metrics endpoint |
| `GRAFANA_ADMIN_PASSWORD` | admin | Grafana admin password |
| `PROMETHEUS_RETENTION` | 30d | Metrics retention period |
| `LOG_LEVEL` | info | Application log level |
| `WORKER_POOL_SIZE` | 10 | API worker pool size |

## Accessing Services

### Grafana Dashboard

1. Open http://localhost:3000
2. Login with `admin` / `admin` (change on first login)
3. Navigate to **Dashboards** → **Freightliner**
4. Pre-configured dashboard shows:
   - Replication metrics
   - API performance
   - Error rates
   - Worker pool utilization
   - System resources

### Prometheus Queries

1. Open http://localhost:9090
2. Example queries:
```promql
# Replication rate
rate(freightliner_replications_total[5m])

# Error rate
rate(freightliner_replications_failed_total[5m])

# API latency
histogram_quantile(0.95, freightliner_api_request_duration_seconds_bucket)

# Active workers
freightliner_worker_pool_active
```

### API Endpoints

**Health Check:**
```bash
curl http://localhost:8080/health
```

**Metrics:**
```bash
curl http://localhost:2112/metrics
```

**Example API Usage:**
```bash
# Trigger replication
curl -X POST http://localhost:8080/api/v1/replicate \
  -H "Content-Type: application/json" \
  -d '{
    "source": "gcr.io/project/image:tag",
    "destination": "123456789012.dkr.ecr.us-east-1.amazonaws.com/image:tag"
  }'
```

## Monitoring Features

### Built-in Dashboards

1. **Freightliner Overview**
   - Real-time replication metrics
   - Success/failure rates
   - Performance trends
   - Resource utilization

2. **System Metrics** (Full profile)
   - CPU, memory, disk usage
   - Network I/O
   - Container metrics
   - Host statistics

### Alert Rules

Pre-configured alerts in Prometheus:

- **High Error Rate**: >5% failures over 5 minutes
- **Slow Replications**: P95 latency >5 minutes
- **API Unavailable**: Health check failures
- **High Memory Usage**: >90% for 5 minutes
- **Disk Space Low**: <10% available

View alerts: http://localhost:9090/alerts

### Metrics Collected

**Application Metrics:**
- `freightliner_replications_total` - Total replications
- `freightliner_replications_failed_total` - Failed replications
- `freightliner_replication_duration_seconds` - Replication duration
- `freightliner_api_requests_total` - API request count
- `freightliner_worker_pool_active` - Active workers
- `freightliner_worker_pool_capacity` - Total worker capacity

**System Metrics:** (Full profile)
- CPU usage per core
- Memory usage and swap
- Disk I/O and space
- Network traffic
- Container stats

## Data Persistence

All data is stored in named Docker volumes:

```bash
# List volumes
docker volume ls | grep freightliner

# Inspect volume
docker volume inspect freightliner-prometheus-data

# Backup Grafana data
docker run --rm -v freightliner-grafana-data:/data \
  -v $(pwd):/backup alpine tar czf /backup/grafana-backup.tar.gz /data

# Restore Grafana data
docker run --rm -v freightliner-grafana-data:/data \
  -v $(pwd):/backup alpine tar xzf /backup/grafana-backup.tar.gz -C /
```

### Volume Locations

- `freightliner-prometheus-data` - Prometheus TSDB
- `freightliner-grafana-data` - Grafana dashboards and settings
- `freightliner-alertmanager-data` - AlertManager state
- `freightliner-redis-data` - Redis cache

## Troubleshooting

### Services Won't Start

```bash
# Check logs
docker-compose -f docker-compose.monitoring.yml logs

# Check specific service
docker-compose -f docker-compose.monitoring.yml logs prometheus

# Check health status
docker-compose -f docker-compose.monitoring.yml ps
```

### Grafana Shows "No Data"

1. Check Prometheus is running:
```bash
curl http://localhost:9090/-/healthy
```

2. Verify datasource connection in Grafana:
   - Go to **Configuration** → **Data Sources**
   - Click **Prometheus**
   - Click **Save & Test**

3. Check if Freightliner API is exposing metrics:
```bash
curl http://localhost:2112/metrics
```

### Prometheus Can't Scrape Targets

1. Check Prometheus targets:
   - Open http://localhost:9090/targets
   - Look for "DOWN" targets

2. Verify network connectivity:
```bash
docker exec freightliner-prometheus wget -O- http://freightliner-api:2112/metrics
```

3. Check Prometheus config:
```bash
docker exec freightliner-prometheus cat /etc/prometheus/prometheus.yml
```

### High Memory Usage

Adjust Prometheus retention:

```bash
# Edit .env
PROMETHEUS_RETENTION=7d
PROMETHEUS_RETENTION_SIZE=5GB

# Restart services
docker-compose -f docker-compose.monitoring.yml restart prometheus
```

### Reset Everything

```bash
# Stop and remove containers, networks
docker-compose -f docker-compose.monitoring.yml down

# Remove volumes (WARNING: Deletes all data)
docker-compose -f docker-compose.monitoring.yml down -v

# Clean up completely
docker system prune -a --volumes
```

## Advanced Configuration

### Custom Prometheus Configuration

Edit `monitoring/prometheus/prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'my-custom-target'
    static_configs:
      - targets: ['my-service:9090']
```

Reload configuration:
```bash
docker exec freightliner-prometheus kill -HUP 1
# Or
curl -X POST http://localhost:9090/-/reload
```

### Custom Grafana Dashboards

1. Create/edit dashboard in Grafana UI
2. Export JSON: **Share** → **Export** → **Save to file**
3. Copy to `monitoring/grafana-dashboard.json`
4. Restart Grafana to load changes

### AlertManager Configuration

Edit `monitoring/alertmanager/config.yml`:

```yaml
route:
  receiver: 'slack'
  routes:
    - match:
        severity: critical
      receiver: pagerduty

receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/...'
        channel: '#alerts'
```

## Performance Tuning

### For Development

```env
PROMETHEUS_RETENTION=7d
PROMETHEUS_RETENTION_SIZE=2GB
WORKER_POOL_SIZE=5
MAX_CONCURRENT_REPLICATIONS=3
```

### For Production-Like Testing

```env
PROMETHEUS_RETENTION=90d
PROMETHEUS_RETENTION_SIZE=50GB
WORKER_POOL_SIZE=50
MAX_CONCURRENT_REPLICATIONS=20
REDIS_MAX_MEMORY=2gb
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
- name: Start Monitoring Stack
  run: |
    docker-compose -f docker-compose.monitoring.yml up -d
    sleep 30  # Wait for services to be ready

- name: Run Tests with Metrics
  run: |
    make test

- name: Check Metrics
  run: |
    curl http://localhost:2112/metrics | grep freightliner_replications_total

- name: Tear Down
  run: docker-compose -f docker-compose.monitoring.yml down
```

## Security Considerations

### Production Deployment

1. **Change default passwords:**
```env
GRAFANA_ADMIN_PASSWORD=strong_random_password
```

2. **Enable authentication:**
```env
GRAFANA_ANONYMOUS_ENABLED=false
```

3. **Use secrets management:**
```bash
# Use Docker secrets instead of environment variables
docker secret create grafana_admin_password password.txt
```

4. **Restrict network access:**
```yaml
networks:
  freightliner-monitoring:
    internal: true  # No external access
```

5. **Enable HTTPS:**
   - Configure reverse proxy (nginx, traefik)
   - Use Let's Encrypt certificates
   - Update `GRAFANA_ROOT_URL` to use https://

## Next Steps

1. **Customize dashboards** for your specific needs
2. **Set up alerts** for critical metrics
3. **Configure long-term storage** (Thanos, Cortex, Mimir)
4. **Enable distributed tracing** with Jaeger
5. **Add log aggregation** with Loki

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Freightliner Monitoring Guide](./ARCHITECTURE.md)
- [Alert Rules Reference](./prometheus-alerts.yml)

## Support

For issues or questions:
1. Check logs: `docker-compose -f docker-compose.monitoring.yml logs`
2. Review [Troubleshooting](#troubleshooting) section
3. Open issue on GitHub
4. Check existing monitoring documentation in `/monitoring/docs/`
