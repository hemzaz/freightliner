# Freightliner Monitoring Stack - Quick Start Guide

Complete Docker Compose monitoring setup for local development with Prometheus, Grafana, and Freightliner API.

## 🚀 Quick Start (30 seconds)

```bash
# 1. Copy environment configuration
cp .env.monitoring .env

# 2. Start the stack
./scripts/monitoring-stack.sh start

# 3. Access services
open http://localhost:3000  # Grafana (admin/admin)
open http://localhost:9090  # Prometheus
open http://localhost:8080  # Freightliner API
```

## 📋 What's Included

### Core Services (Basic Stack)

| Service | Port | Purpose | URL |
|---------|------|---------|-----|
| **Grafana** | 3000 | Metrics visualization | http://localhost:3000 |
| **Prometheus** | 9090 | Metrics collection | http://localhost:9090 |
| **Freightliner API** | 8080 | Application | http://localhost:8080 |
| **Metrics Endpoint** | 2112 | Prometheus scrape | http://localhost:2112/metrics |

### Optional Services (Full Stack)

Start with: `./scripts/monitoring-stack.sh start-full`

| Service | Port | Purpose | URL |
|---------|------|---------|-----|
| **AlertManager** | 9093 | Alert management | http://localhost:9093 |
| **Node Exporter** | 9100 | System metrics | http://localhost:9100 |
| **cAdvisor** | 8081 | Container metrics | http://localhost:8081 |
| **Redis** | 6379 | Cache | localhost:6379 |

## 🎯 Common Commands

### Using the Management Script

```bash
# Start basic stack (recommended for development)
./scripts/monitoring-stack.sh start

# Start full stack with all services
./scripts/monitoring-stack.sh start-full

# Check status
./scripts/monitoring-stack.sh status

# View logs
./scripts/monitoring-stack.sh logs
./scripts/monitoring-stack.sh logs prometheus

# Check health
./scripts/monitoring-stack.sh health

# Show current metrics
./scripts/monitoring-stack.sh metrics

# Stop stack
./scripts/monitoring-stack.sh stop

# Restart stack
./scripts/monitoring-stack.sh restart

# Backup data
./scripts/monitoring-stack.sh backup

# Clean all data (destructive!)
./scripts/monitoring-stack.sh clean
```

### Using Docker Compose Directly

```bash
# Start services
docker-compose -f docker-compose.monitoring.yml up -d

# Start with full profile
docker-compose -f docker-compose.monitoring.yml --profile full up -d

# View logs
docker-compose -f docker-compose.monitoring.yml logs -f

# Check status
docker-compose -f docker-compose.monitoring.yml ps

# Stop services
docker-compose -f docker-compose.monitoring.yml down

# Remove all data
docker-compose -f docker-compose.monitoring.yml down -v
```

## 📊 Accessing Dashboards

### Grafana Setup

1. **First Login**
   - URL: http://localhost:3000
   - Username: `admin`
   - Password: `admin` (change on first login)

2. **View Dashboards**
   - Navigate to **Dashboards** → **Browse**
   - Select **Freightliner** folder
   - Available dashboards:
     - Replication Overview
     - Error & Latency Analysis
     - Infrastructure Metrics
     - Business Metrics

3. **Explore Metrics**
   - Go to **Explore** (compass icon)
   - Select **Prometheus** datasource
   - Try example queries:
     ```promql
     # Replication rate
     rate(freightliner_replications_total[5m])

     # Error rate
     rate(freightliner_replications_failed_total[5m])

     # API latency (95th percentile)
     histogram_quantile(0.95, freightliner_api_request_duration_seconds_bucket)

     # Active workers
     freightliner_worker_pool_active
     ```

### Prometheus Queries

Access Prometheus UI at http://localhost:9090

**Useful Queries:**

```promql
# Current replication rate (per second)
rate(freightliner_replications_total[5m])

# Error percentage
sum(rate(freightliner_replications_failed_total[5m])) / sum(rate(freightliner_replications_total[5m])) * 100

# Average replication duration
rate(freightliner_replication_duration_seconds_sum[5m]) / rate(freightliner_replication_duration_seconds_count[5m])

# Worker pool utilization
freightliner_worker_pool_active / freightliner_worker_pool_capacity * 100

# Memory usage
process_resident_memory_bytes / 1024 / 1024

# Goroutines
go_goroutines
```

## 🔧 Configuration

### Environment Variables

Edit `.env` file to customize:

```bash
# Core settings
ENVIRONMENT=development
LOG_LEVEL=info

# Ports
GRAFANA_PORT=3000
PROMETHEUS_PORT=9090
API_PORT=8080
METRICS_PORT=2112

# Security
GRAFANA_ADMIN_PASSWORD=your_secure_password

# Performance
WORKER_POOL_SIZE=10
MAX_CONCURRENT_REPLICATIONS=5

# Retention
PROMETHEUS_RETENTION=30d
PROMETHEUS_RETENTION_SIZE=10GB
```

### Grafana Customization

1. **Change Admin Password**
   ```bash
   # Edit .env
   GRAFANA_ADMIN_PASSWORD=secure_password_here

   # Restart Grafana
   docker-compose -f docker-compose.monitoring.yml restart grafana
   ```

2. **Add Custom Datasources**
   - Add YAML files to `monitoring/grafana/datasources/`
   - Restart Grafana

3. **Add Custom Dashboards**
   - Export JSON from Grafana UI
   - Place in `monitoring/grafana/dashboards/`
   - Or update `monitoring/grafana-dashboard.json`

### Prometheus Customization

1. **Modify Scrape Config**
   - Edit `monitoring/prometheus/prometheus-local.yml`
   - Reload config:
     ```bash
     docker exec freightliner-prometheus kill -HUP 1
     # Or
     curl -X POST http://localhost:9090/-/reload
     ```

2. **Add Alert Rules**
   - Edit `monitoring/prometheus/alert-rules.yml`
   - Reload config (same as above)

## 🚨 Troubleshooting

### Services Won't Start

```bash
# Check Docker is running
docker info

# Check logs
./scripts/monitoring-stack.sh logs

# Check specific service
docker-compose -f docker-compose.monitoring.yml logs prometheus
```

### Grafana Shows "No Data"

1. **Check Prometheus is running:**
   ```bash
   curl http://localhost:9090/-/healthy
   ```

2. **Verify datasource in Grafana:**
   - Go to Configuration → Data Sources
   - Click "Prometheus"
   - Click "Save & Test"

3. **Check if API is exposing metrics:**
   ```bash
   curl http://localhost:2112/metrics | grep freightliner_
   ```

### Prometheus Can't Scrape Targets

1. **Check targets in Prometheus UI:**
   - Open http://localhost:9090/targets
   - Look for "DOWN" status

2. **Test connectivity:**
   ```bash
   docker exec freightliner-prometheus wget -O- http://freightliner-api:2112/metrics
   ```

3. **Verify network:**
   ```bash
   docker network inspect freightliner-monitoring
   ```

### Port Conflicts

If ports are already in use, modify in `.env`:

```bash
GRAFANA_PORT=3001
PROMETHEUS_PORT=9091
API_PORT=8081
```

### API Not Building

```bash
# Build manually
./scripts/monitoring-stack.sh build

# Or
docker-compose -f docker-compose.monitoring.yml build freightliner-api

# Check build logs
docker-compose -f docker-compose.monitoring.yml build --no-cache freightliner-api
```

## 💾 Data Management

### Backup Data

```bash
# Using script (recommended)
./scripts/monitoring-stack.sh backup

# Manual backup
docker run --rm -v freightliner-prometheus-data:/data \
  -v $(pwd)/backup:/backup alpine \
  tar czf /backup/prometheus-$(date +%Y%m%d).tar.gz /data
```

### Restore Data

```bash
docker run --rm -v freightliner-prometheus-data:/data \
  -v $(pwd)/backup:/backup alpine \
  tar xzf /backup/prometheus-20231201.tar.gz -C /
```

### Clean All Data

```bash
# Interactive prompt
./scripts/monitoring-stack.sh clean

# Force delete
docker-compose -f docker-compose.monitoring.yml down -v
```

## 📈 Performance Tuning

### For Development (Minimal Resources)

```env
PROMETHEUS_RETENTION=7d
PROMETHEUS_RETENTION_SIZE=2GB
WORKER_POOL_SIZE=5
MAX_CONCURRENT_REPLICATIONS=3
REDIS_MAX_MEMORY=256mb
```

### For Production-Like Testing

```env
PROMETHEUS_RETENTION=90d
PROMETHEUS_RETENTION_SIZE=50GB
WORKER_POOL_SIZE=50
MAX_CONCURRENT_REPLICATIONS=20
REDIS_MAX_MEMORY=2gb
```

## 🔐 Security Checklist

- [ ] Change default Grafana password
- [ ] Disable anonymous access: `GRAFANA_ANONYMOUS_ENABLED=false`
- [ ] Use strong passwords in `.env`
- [ ] Never commit `.env` file
- [ ] Restrict network access in production
- [ ] Enable HTTPS with reverse proxy
- [ ] Use secrets management for cloud credentials

## 🎓 Next Steps

1. **Explore Metrics**
   - Open Grafana dashboards
   - Run example Prometheus queries
   - View real-time data

2. **Customize Dashboards**
   - Edit existing dashboards
   - Create new visualizations
   - Export and share

3. **Set Up Alerts**
   - Configure AlertManager
   - Add notification channels (Slack, Email, PagerDuty)
   - Test alert routing

4. **Add More Services**
   - Enable Redis for caching
   - Add distributed tracing with Jaeger
   - Integrate log aggregation with Loki

5. **Production Deployment**
   - Review security settings
   - Configure backup strategy
   - Set up long-term storage
   - Enable authentication

## 📚 Additional Resources

- **Full Documentation**: [monitoring/README-DOCKER-MONITORING.md](../monitoring/README-DOCKER-MONITORING.md)
- **Architecture**: [monitoring/ARCHITECTURE.md](../monitoring/ARCHITECTURE.md)
- **Alert Rules**: [monitoring/prometheus-alerts.yml](../monitoring/prometheus-alerts.yml)
- **Deployment Checklist**: [monitoring/DEPLOYMENT_CHECKLIST.md](../monitoring/DEPLOYMENT_CHECKLIST.md)

## 🆘 Getting Help

1. Check service logs: `./scripts/monitoring-stack.sh logs [service]`
2. Verify health: `./scripts/monitoring-stack.sh health`
3. Review documentation in `monitoring/` directory
4. Check existing dashboards for examples
5. Open GitHub issue with logs and configuration

---

**Last Updated**: 2024-12-01
**Version**: 1.0.0
**Maintainer**: Freightliner DevOps Team
