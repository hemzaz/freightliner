# Performance Testing and Optimization Guide

## Table of Contents
1. [Performance Testing Framework Overview](#performance-testing-framework-overview)
2. [Running Performance Tests](#running-performance-tests)
3. [Baseline Metrics](#baseline-metrics)
4. [Establishing Baselines](#establishing-baselines)
5. [Monitoring and Alerting](#monitoring-and-alerting)
6. [Performance Tuning Guide](#performance-tuning-guide)

---

## Performance Testing Framework Overview

### Framework Architecture

Our performance testing stack uses industry-standard tools for comprehensive load and stress testing:

```
┌─────────────────────────────────────────────────────────┐
│                   Performance Testing Stack              │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │     k6       │  │   Locust     │  │   JMeter     │ │
│  │  (Protocol)  │  │  (Python)    │  │   (Java)     │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘ │
│         │                  │                  │          │
│         └──────────────────┼──────────────────┘          │
│                            │                             │
│                     ┌──────▼───────┐                    │
│                     │  Application  │                    │
│                     │   Under Test  │                    │
│                     └──────┬───────┘                    │
│                            │                             │
│         ┌──────────────────┼──────────────────┐         │
│         │                  │                  │          │
│  ┌──────▼───────┐  ┌──────▼───────┐  ┌──────▼───────┐ │
│  │  Prometheus  │  │   Grafana    │  │  CloudWatch  │ │
│  │  (Metrics)   │  │  (Dashboards)│  │   (AWS)      │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Testing Layers

1. **Protocol Level (k6)**: Fast, resource-efficient, ideal for API testing
2. **User Behavior (Locust)**: Python-based, excellent for complex user journeys
3. **Enterprise (JMeter)**: GUI-based, comprehensive reporting, CI/CD integration

### Key Metrics Tracked

- **Response Time**: P50, P95, P99 latencies
- **Throughput**: Requests per second (RPS)
- **Error Rate**: Percentage of failed requests
- **Resource Utilization**: CPU, Memory, Network I/O
- **Database Performance**: Query execution time, connection pool usage
- **Cache Hit Ratio**: Redis/CDN effectiveness

---

## Running Performance Tests

### Prerequisites

```bash
# Install k6
brew install k6  # macOS
# or
curl -fsSL https://k6.io/install | sh  # Linux

# Install Locust
pip3 install locust

# Install JMeter (optional)
brew install jmeter  # macOS
```

### Directory Structure

```
performance/
├── k6/
│   ├── api-load-test.js
│   ├── stress-test.js
│   └── spike-test.js
├── locust/
│   ├── user-journey.py
│   └── locustfile.py
├── jmeter/
│   └── comprehensive-test.jmx
├── results/
│   └── .gitkeep
└── scripts/
    ├── run-all-tests.sh
    └── analyze-results.sh
```

### Quick Start Examples

#### 1. k6 API Load Test

```bash
# Basic load test - 10 VUs for 30 seconds
k6 run --vus 10 --duration 30s performance/k6/api-load-test.js

# Example k6 script (api-load-test.js):
```

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const apiDuration = new Trend('api_duration');

export const options = {
  stages: [
    { duration: '2m', target: 50 },   // Ramp up to 50 users
    { duration: '5m', target: 50 },   // Stay at 50 users
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% < 500ms, 99% < 1s
    http_req_failed: ['rate<0.01'], // Error rate < 1%
    errors: ['rate<0.1'],
  },
};

export default function () {
  // Test GET endpoint
  const getRes = http.get('https://api.example.com/api/v1/resources');
  check(getRes, {
    'GET status is 200': (r) => r.status === 200,
    'GET response time < 200ms': (r) => r.timings.duration < 200,
  });
  errorRate.add(getRes.status !== 200);
  apiDuration.add(getRes.timings.duration);

  sleep(1);

  // Test POST endpoint
  const payload = JSON.stringify({
    name: 'Test Resource',
    type: 'performance-test',
  });
  const postRes = http.post('https://api.example.com/api/v1/resources', payload, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(postRes, {
    'POST status is 201': (r) => r.status === 201,
    'POST response time < 300ms': (r) => r.timings.duration < 300,
  });

  sleep(2);
}
```

**Run with output:**
```bash
k6 run --out json=performance/results/load-test-$(date +%Y%m%d-%H%M%S).json \
       performance/k6/api-load-test.js
```

**Expected Output:**
```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: api-load-test.js
     output: json (performance/results/load-test-20250101-143000.json)

  scenarios: (100.00%) 1 scenario, 100 max VUs, 16m30s max duration

     ✓ GET status is 200
     ✓ GET response time < 200ms
     ✓ POST status is 201
     ✓ POST response time < 300ms

     checks.........................: 100.00% ✓ 48000  ✗ 0
     data_received..................: 120 MB  750 kB/s
     data_sent......................: 24 MB   150 kB/s
     http_req_blocked...............: avg=1.2ms   min=0.8ms   med=1.1ms   max=15.2ms  p(95)=2.1ms   p(99)=3.5ms
     http_req_connecting............: avg=0.8ms   min=0.5ms   med=0.7ms   max=12.1ms  p(95)=1.5ms   p(99)=2.8ms
     http_req_duration..............: avg=145ms   min=82ms    med=138ms   max=487ms   p(95)=234ms   p(99)=356ms
     http_req_failed................: 0.00%   ✓ 0      ✗ 24000
     http_req_receiving.............: avg=1.5ms   min=0.5ms   med=1.2ms   max=18.5ms  p(95)=3.2ms   p(99)=5.8ms
     http_req_sending...............: avg=0.8ms   min=0.3ms   med=0.7ms   max=8.2ms   p(95)=1.5ms   p(99)=2.8ms
     http_req_tls_handshaking.......: avg=0ms     min=0ms     med=0ms     max=0ms     p(95)=0ms     p(99)=0ms
     http_req_waiting...............: avg=142.7ms min=80ms    med=135.8ms max=480ms   p(95)=230ms   p(99)=350ms
     http_reqs......................: 24000   150/s
     iteration_duration.............: avg=3.2s    min=3.1s    med=3.15s   max=4.8s    p(95)=3.5s    p(99)=3.9s
     iterations.....................: 12000   75/s
     vus............................: 100     min=0    max=100
     vus_max........................: 100     min=100  max=100
```

#### 2. Locust User Behavior Test

```bash
# Start Locust web UI
locust -f performance/locust/user-journey.py --host=https://api.example.com

# Or headless mode
locust -f performance/locust/user-journey.py \
       --host=https://api.example.com \
       --users 100 \
       --spawn-rate 10 \
       --run-time 5m \
       --headless \
       --html performance/results/locust-report-$(date +%Y%m%d-%H%M%S).html
```

**Example Locust script (user-journey.py):**

```python
from locust import HttpUser, task, between
import random
import json

class APIUser(HttpUser):
    wait_time = between(1, 3)  # Wait 1-3 seconds between tasks

    def on_start(self):
        """Login and setup session"""
        response = self.client.post("/auth/login", json={
            "username": f"user{random.randint(1, 1000)}@example.com",
            "password": "test123"
        })
        if response.status_code == 200:
            self.token = response.json()["token"]

    @task(3)  # Weight: 3x more likely than other tasks
    def get_resources(self):
        """Fetch resource list"""
        headers = {"Authorization": f"Bearer {self.token}"}
        with self.client.get("/api/v1/resources",
                            headers=headers,
                            catch_response=True) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")

    @task(2)
    def get_single_resource(self):
        """Fetch single resource detail"""
        resource_id = random.randint(1, 100)
        headers = {"Authorization": f"Bearer {self.token}"}
        self.client.get(f"/api/v1/resources/{resource_id}",
                       headers=headers,
                       name="/api/v1/resources/[id]")

    @task(1)
    def create_resource(self):
        """Create new resource"""
        headers = {
            "Authorization": f"Bearer {self.token}",
            "Content-Type": "application/json"
        }
        payload = {
            "name": f"Resource {random.randint(1, 10000)}",
            "type": random.choice(["type_a", "type_b", "type_c"]),
            "metadata": {"test": True}
        }
        self.client.post("/api/v1/resources",
                        headers=headers,
                        json=payload)

    @task(1)
    def search_resources(self):
        """Search resources"""
        headers = {"Authorization": f"Bearer {self.token}"}
        query = random.choice(["test", "prod", "staging"])
        self.client.get(f"/api/v1/resources/search?q={query}",
                       headers=headers,
                       name="/api/v1/resources/search")
```

**Expected Locust Output:**
```
Type     Name                              # reqs      # fails |    Avg     Min     Max    Med |   req/s  failures/s
--------|-------------------------------|---------|-------------|-------|-------|-------|-------|--------|-----------
GET      /api/v1/resources                  3456         0     |     95      45     850     80 |   23.04        0.00
GET      /api/v1/resources/[id]             2304         0     |     82      38     720     70 |   15.36        0.00
POST     /api/v1/resources                  1152         2     |    128      55     980    105 |    7.68        0.01
GET      /api/v1/resources/search           1152         1     |    145      62    1200    120 |    7.68        0.01
POST     /auth/login                         100         0     |    210     120     650    195 |    0.67        0.00
--------|-------------------------------|---------|-------------|-------|-------|-------|-------|--------|-----------
         Aggregated                         8164         3     |    104      38    1200     85 |   54.43        0.02

Response time percentiles (approximated):
Type     Name                              50%    66%    75%    80%    90%    95%    98%    99%  99.9% 99.99%   100% # reqs
--------|-------------------------------|--------|------|------|------|------|------|------|------|------|------|------|------
GET      /api/v1/resources                  80     95    110    125    175    220    320    450    850    850    850   3456
GET      /api/v1/resources/[id]             70     82     95    105    145    180    265    380    720    720    720   2304
POST     /api/v1/resources                 105    125    145    160    220    285    420    580    980    980    980   1152
GET      /api/v1/resources/search          120    140    165    180    250    320    480    650   1200   1200   1200   1152
POST     /auth/login                       195    215    235    255    325    395    520    610    650    650    650    100
--------|-------------------------------|--------|------|------|------|------|------|------|------|------|------|------|------
         Aggregated                         85    105    125    140    195    250    370    520   1100   1200   1200   8164
```

#### 3. Stress Test

```bash
# k6 stress test - gradually increase load to breaking point
k6 run performance/k6/stress-test.js
```

**Example stress-test.js:**

```javascript
export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Normal load
    { duration: '5m', target: 100 },
    { duration: '2m', target: 200 },   // Around breaking point
    { duration: '5m', target: 200 },
    { duration: '2m', target: 300 },   // Beyond breaking point
    { duration: '5m', target: 300 },
    { duration: '2m', target: 400 },   // Push to failure
    { duration: '5m', target: 400 },
    { duration: '5m', target: 0 },     // Recovery
  ],
};
```

---

## Baseline Metrics

### API Performance Targets

| Metric | Target | Acceptable | Critical |
|--------|--------|------------|----------|
| **P50 Latency** | < 100ms | < 200ms | > 500ms |
| **P95 Latency** | < 250ms | < 500ms | > 1000ms |
| **P99 Latency** | < 500ms | < 1000ms | > 2000ms |
| **Throughput** | > 1000 RPS | > 500 RPS | < 100 RPS |
| **Error Rate** | < 0.1% | < 1% | > 5% |
| **Availability** | 99.9% | 99.5% | < 99% |

### Database Performance Targets

| Metric | Target | Acceptable | Critical |
|--------|--------|------------|----------|
| **Query Execution (P95)** | < 50ms | < 100ms | > 500ms |
| **Connection Pool Usage** | < 70% | < 85% | > 95% |
| **Slow Queries** | 0 | < 5/min | > 20/min |
| **Deadlocks** | 0 | < 1/hour | > 5/hour |

### Cache Performance Targets

| Metric | Target | Acceptable | Critical |
|--------|--------|------------|----------|
| **Hit Ratio** | > 95% | > 80% | < 60% |
| **Read Latency (P95)** | < 5ms | < 10ms | > 50ms |
| **Write Latency (P95)** | < 10ms | < 20ms | > 100ms |

### Resource Utilization Targets

| Resource | Target | Acceptable | Critical |
|----------|--------|------------|----------|
| **CPU Usage** | < 60% | < 75% | > 90% |
| **Memory Usage** | < 70% | < 85% | > 95% |
| **Disk I/O** | < 50% | < 70% | > 90% |
| **Network I/O** | < 60% | < 80% | > 95% |

### Frontend Performance (Core Web Vitals)

| Metric | Good | Needs Improvement | Poor |
|--------|------|-------------------|------|
| **LCP (Largest Contentful Paint)** | < 2.5s | 2.5s - 4s | > 4s |
| **FID (First Input Delay)** | < 100ms | 100ms - 300ms | > 300ms |
| **CLS (Cumulative Layout Shift)** | < 0.1 | 0.1 - 0.25 | > 0.25 |
| **TTFB (Time to First Byte)** | < 800ms | 800ms - 1800ms | > 1800ms |

---

## Establishing Baselines

### Phase 1: Staging Environment Baseline

#### Step 1: Deploy to Staging

```bash
# Deploy application to staging
./scripts/deploy-staging.sh

# Verify deployment
curl -s https://staging-api.example.com/health | jq
```

#### Step 2: Run Baseline Test Suite

```bash
# Create baseline test script
cat > performance/scripts/baseline-test.sh << 'EOF'
#!/bin/bash
set -e

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
RESULTS_DIR="performance/results/baseline-${TIMESTAMP}"
mkdir -p "${RESULTS_DIR}"

echo "Starting baseline performance tests - ${TIMESTAMP}"

# 1. Warm-up (not recorded)
echo "Warming up application..."
k6 run --vus 10 --duration 2m \
   --quiet \
   performance/k6/api-load-test.js

# 2. Light load test
echo "Running light load test..."
k6 run --vus 25 --duration 5m \
   --out json="${RESULTS_DIR}/light-load.json" \
   performance/k6/api-load-test.js

# 3. Normal load test
echo "Running normal load test..."
k6 run --vus 50 --duration 10m \
   --out json="${RESULTS_DIR}/normal-load.json" \
   performance/k6/api-load-test.js

# 4. Peak load test
echo "Running peak load test..."
k6 run --vus 100 --duration 10m \
   --out json="${RESULTS_DIR}/peak-load.json" \
   performance/k6/api-load-test.js

# 5. Endurance test
echo "Running endurance test..."
k6 run --vus 50 --duration 30m \
   --out json="${RESULTS_DIR}/endurance.json" \
   performance/k6/api-load-test.js

# 6. Analyze results
echo "Analyzing results..."
./performance/scripts/analyze-results.sh "${RESULTS_DIR}"

echo "Baseline tests complete. Results in: ${RESULTS_DIR}"
EOF

chmod +x performance/scripts/baseline-test.sh

# Run baseline tests
./performance/scripts/baseline-test.sh
```

#### Step 3: Analyze and Document Baseline

```bash
# Create analysis script
cat > performance/scripts/analyze-results.sh << 'EOF'
#!/bin/bash
RESULTS_DIR=$1

echo "Analyzing performance test results..."

# Extract key metrics from k6 JSON output
for test in light-load normal-load peak-load endurance; do
    echo "=== ${test} ==="
    jq -r '
        .metrics |
        {
            "http_req_duration_p95": .http_req_duration.values["p(95)"],
            "http_req_duration_p99": .http_req_duration.values["p(99)"],
            "http_req_failed_rate": .http_req_failed.values.rate,
            "http_reqs_rate": .http_reqs.values.rate
        }
    ' "${RESULTS_DIR}/${test}.json"
    echo ""
done

# Generate summary report
cat > "${RESULTS_DIR}/BASELINE_REPORT.md" << 'REPORT'
# Performance Baseline Report

Generated: $(date)

## Test Environment
- Environment: Staging
- Infrastructure: [EC2 t3.large, RDS db.t3.medium, ElastiCache cache.t3.small]
- Application Version: $(git rev-parse --short HEAD)

## Results Summary

### Light Load (25 VUs)
- P95 Latency: XXms
- P99 Latency: XXms
- Throughput: XX RPS
- Error Rate: X.XX%

### Normal Load (50 VUs)
- P95 Latency: XXms
- P99 Latency: XXms
- Throughput: XX RPS
- Error Rate: X.XX%

### Peak Load (100 VUs)
- P95 Latency: XXms
- P99 Latency: XXms
- Throughput: XX RPS
- Error Rate: X.XX%

### Endurance (50 VUs, 30min)
- P95 Latency: XXms
- P99 Latency: XXms
- Throughput: XX RPS
- Error Rate: X.XX%
- Memory Leak Detected: No

## Bottlenecks Identified
1. Database query performance on `/api/v1/resources` endpoint
2. High memory usage during peak load
3. Connection pool saturation at 100+ concurrent users

## Recommendations
1. Add database indexes on frequently queried columns
2. Implement Redis caching for read-heavy endpoints
3. Increase connection pool size from 20 to 50
4. Enable HTTP/2 for API endpoints

REPORT

echo "Report generated: ${RESULTS_DIR}/BASELINE_REPORT.md"
EOF

chmod +x performance/scripts/analyze-results.sh
```

### Phase 2: Production Baseline

#### Step 1: Shadow Testing

```bash
# Use production traffic replay for realistic testing
# Tool: GoReplay (gor) or AWS Application Load Balancer request mirroring

# Example with gor (requires deployment)
gor --input-raw :8080 \
    --output-http "https://staging-api.example.com" \
    --output-http-workers 10
```

#### Step 2: Gradual Production Rollout

```bash
# Blue-green deployment with monitoring
# 1. Deploy to 10% of production traffic
# 2. Monitor for 1 hour
# 3. Gradually increase to 50%, then 100%

# Monitor key metrics during rollout
watch -n 5 '
  echo "=== API Response Time ==="
  curl -s "https://api.example.com/metrics" | grep http_request_duration

  echo -e "\n=== Error Rate ==="
  curl -s "https://api.example.com/metrics" | grep http_requests_failed

  echo -e "\n=== CPU/Memory ==="
  aws cloudwatch get-metric-statistics \
    --namespace AWS/ECS \
    --metric-name CPUUtilization \
    --dimensions Name=ServiceName,Value=api-service \
    --statistics Average \
    --start-time $(date -u -d "5 minutes ago" +%Y-%m-%dT%H:%M:%S) \
    --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
    --period 300
'
```

#### Step 3: Document Production Baseline

```bash
# Query production metrics from last 7 days
aws cloudwatch get-metric-statistics \
  --namespace AWS/ApplicationELB \
  --metric-name TargetResponseTime \
  --dimensions Name=LoadBalancer,Value=app/api-lb/xxx \
  --statistics Average,p95,p99 \
  --start-time $(date -u -d "7 days ago" +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 3600 \
  > performance/baselines/production-baseline-$(date +%Y%m%d).json
```

---

## Monitoring and Alerting

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093

rule_files:
  - "alerts/performance-alerts.yml"

scrape_configs:
  - job_name: 'api-service'
    static_configs:
      - targets: ['api:8080']
    metrics_path: '/metrics'

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

### Alert Rules

```yaml
# alerts/performance-alerts.yml
groups:
  - name: performance
    interval: 30s
    rules:
      # High latency alert
      - alert: HighAPILatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
        for: 5m
        labels:
          severity: warning
          category: performance
        annotations:
          summary: "High API latency detected"
          description: "P95 latency is {{ $value }}s (threshold: 0.5s)"
          runbook_url: "https://wiki.example.com/runbooks/high-latency"

      # Critical latency alert
      - alert: CriticalAPILatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1.0
        for: 2m
        labels:
          severity: critical
          category: performance
        annotations:
          summary: "CRITICAL: API latency exceeds 1s"
          description: "P95 latency is {{ $value }}s"

      # High error rate
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.01
        for: 3m
        labels:
          severity: warning
          category: reliability
        annotations:
          summary: "Error rate exceeds 1%"
          description: "Current error rate: {{ $value | humanizePercentage }}"

      # Database slow queries
      - alert: DatabaseSlowQueries
        expr: rate(pg_stat_statements_mean_exec_time[5m]) > 100
        for: 5m
        labels:
          severity: warning
          category: database
        annotations:
          summary: "Database queries are slow"
          description: "Average query time: {{ $value }}ms"

      # Low cache hit ratio
      - alert: LowCacheHitRatio
        expr: rate(redis_keyspace_hits_total[5m]) / (rate(redis_keyspace_hits_total[5m]) + rate(redis_keyspace_misses_total[5m])) < 0.8
        for: 10m
        labels:
          severity: warning
          category: caching
        annotations:
          summary: "Redis cache hit ratio is low"
          description: "Cache hit ratio: {{ $value | humanizePercentage }}"

      # High CPU usage
      - alert: HighCPUUsage
        expr: rate(process_cpu_seconds_total[5m]) > 0.8
        for: 10m
        labels:
          severity: warning
          category: resources
        annotations:
          summary: "High CPU usage detected"
          description: "CPU usage: {{ $value | humanizePercentage }}"

      # High memory usage
      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes / node_memory_MemTotal_bytes > 0.85
        for: 5m
        labels:
          severity: warning
          category: resources
        annotations:
          summary: "High memory usage detected"
          description: "Memory usage: {{ $value | humanizePercentage }}"
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Performance Monitoring Dashboard",
    "panels": [
      {
        "title": "API Response Time (P95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ],
        "thresholds": [
          {"value": 0.25, "color": "green"},
          {"value": 0.5, "color": "yellow"},
          {"value": 1.0, "color": "red"}
        ]
      },
      {
        "title": "Requests Per Second",
        "targets": [
          {
            "expr": "rate(http_requests_total[1m])"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Database Query Time (P95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(pg_stat_statements_exec_time_bucket[5m]))"
          }
        ]
      },
      {
        "title": "Cache Hit Ratio",
        "targets": [
          {
            "expr": "rate(redis_keyspace_hits_total[5m]) / (rate(redis_keyspace_hits_total[5m]) + rate(redis_keyspace_misses_total[5m]))"
          }
        ]
      }
    ]
  }
}
```

### CloudWatch Alarms (AWS)

```bash
# Create CloudWatch alarms for production
aws cloudwatch put-metric-alarm \
  --alarm-name api-high-latency \
  --alarm-description "Alert when API P95 latency exceeds 500ms" \
  --metric-name TargetResponseTime \
  --namespace AWS/ApplicationELB \
  --statistic Average \
  --period 300 \
  --evaluation-periods 2 \
  --threshold 0.5 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=LoadBalancer,Value=app/api-lb/xxx \
  --alarm-actions arn:aws:sns:us-east-1:123456789:performance-alerts

aws cloudwatch put-metric-alarm \
  --alarm-name api-high-error-rate \
  --alarm-description "Alert when error rate exceeds 1%" \
  --metric-name HTTPCode_Target_5XX_Count \
  --namespace AWS/ApplicationELB \
  --statistic Sum \
  --period 300 \
  --evaluation-periods 2 \
  --threshold 10 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=LoadBalancer,Value=app/api-lb/xxx \
  --alarm-actions arn:aws:sns:us-east-1:123456789:performance-alerts

aws cloudwatch put-metric-alarm \
  --alarm-name rds-high-cpu \
  --alarm-description "Alert when RDS CPU exceeds 80%" \
  --metric-name CPUUtilization \
  --namespace AWS/RDS \
  --statistic Average \
  --period 300 \
  --evaluation-periods 2 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=DBInstanceIdentifier,Value=prod-db \
  --alarm-actions arn:aws:sns:us-east-1:123456789:performance-alerts
```

### Alert Severity Levels

| Severity | Response Time | Escalation | Examples |
|----------|---------------|------------|----------|
| **Critical** | Immediate | Page on-call engineer | P95 > 1s, Error rate > 5%, Complete outage |
| **Warning** | 15 minutes | Slack notification | P95 > 500ms, Error rate > 1%, CPU > 80% |
| **Info** | Next business day | Email | P95 > 250ms, Cache hit < 90% |

---

## Performance Tuning Guide

### 1. Application-Level Optimization

#### Database Query Optimization

```sql
-- Identify slow queries
SELECT
  query,
  calls,
  mean_exec_time,
  total_exec_time,
  stddev_exec_time
FROM pg_stat_statements
WHERE mean_exec_time > 100  -- queries slower than 100ms
ORDER BY mean_exec_time DESC
LIMIT 20;

-- Add indexes for frequently queried columns
CREATE INDEX CONCURRENTLY idx_resources_created_at ON resources(created_at);
CREATE INDEX CONCURRENTLY idx_resources_type_status ON resources(type, status);

-- Analyze query execution plan
EXPLAIN ANALYZE
SELECT * FROM resources WHERE type = 'active' AND status = 'approved';
```

#### Connection Pool Tuning

```javascript
// PostgreSQL connection pool configuration
const pool = new Pool({
  max: 50,                    // Maximum pool size
  min: 10,                    // Minimum pool size
  idleTimeoutMillis: 30000,   // Close idle connections after 30s
  connectionTimeoutMillis: 5000, // Fail fast if can't connect in 5s
  maxUses: 7500,              // Close connection after 7500 queries
});

// Monitor pool health
setInterval(() => {
  console.log({
    total: pool.totalCount,
    idle: pool.idleCount,
    waiting: pool.waitingCount,
  });
}, 60000);
```

#### Caching Strategy

```javascript
// Multi-layer caching strategy
const NodeCache = require('node-cache');
const Redis = require('ioredis');

// Layer 1: In-memory cache (fastest, smallest)
const memoryCache = new NodeCache({ stdTTL: 60, checkperiod: 10 });

// Layer 2: Redis cache (fast, larger)
const redisCache = new Redis({
  host: 'redis.example.com',
  port: 6379,
  retryStrategy: (times) => Math.min(times * 50, 2000),
});

async function getCachedResource(id) {
  // Try L1 cache first
  let resource = memoryCache.get(`resource:${id}`);
  if (resource) {
    return { data: resource, source: 'memory' };
  }

  // Try L2 cache
  resource = await redisCache.get(`resource:${id}`);
  if (resource) {
    resource = JSON.parse(resource);
    memoryCache.set(`resource:${id}`, resource);
    return { data: resource, source: 'redis' };
  }

  // Fetch from database
  resource = await db.query('SELECT * FROM resources WHERE id = $1', [id]);

  // Populate caches
  await redisCache.setex(`resource:${id}`, 3600, JSON.stringify(resource));
  memoryCache.set(`resource:${id}`, resource);

  return { data: resource, source: 'database' };
}
```

#### API Response Compression

```javascript
// Enable compression middleware
const compression = require('compression');

app.use(compression({
  filter: (req, res) => {
    if (req.headers['x-no-compression']) {
      return false;
    }
    return compression.filter(req, res);
  },
  level: 6,  // Balance between speed and compression ratio
}));
```

### 2. Infrastructure-Level Optimization

#### Auto-scaling Configuration

```yaml
# ECS Auto-scaling
Resources:
  ServiceScalingTarget:
    Type: AWS::ApplicationAutoScaling::ScalableTarget
    Properties:
      MaxCapacity: 20
      MinCapacity: 2
      ResourceId: !Sub service/${ECSCluster}/${ServiceName}
      ScalableDimension: ecs:service:DesiredCount
      ServiceNamespace: ecs

  ServiceScalingPolicy:
    Type: AWS::ApplicationAutoScaling::ScalingPolicy
    Properties:
      PolicyName: cpu-scaling
      PolicyType: TargetTrackingScaling
      ScalingTargetId: !Ref ServiceScalingTarget
      TargetTrackingScalingPolicyConfiguration:
        TargetValue: 70.0
        PredefinedMetricSpecification:
          PredefinedMetricType: ECSServiceAverageCPUUtilization
        ScaleInCooldown: 300
        ScaleOutCooldown: 60
```

#### Load Balancer Optimization

```yaml
# Application Load Balancer settings
Properties:
  LoadBalancerAttributes:
    - Key: idle_timeout.timeout_seconds
      Value: '60'
    - Key: routing.http2.enabled
      Value: 'true'
    - Key: routing.http.drop_invalid_header_fields.enabled
      Value: 'true'
```

#### CDN Configuration

```javascript
// CloudFront distribution for static assets
{
  "DistributionConfig": {
    "CacheBehaviors": [
      {
        "PathPattern": "/static/*",
        "TargetOriginId": "S3-static",
        "ViewerProtocolPolicy": "redirect-to-https",
        "CachePolicyId": "658327ea-f89d-4fab-a63d-7e88639e58f6",  // CachingOptimized
        "Compress": true,
        "DefaultTTL": 86400,
        "MaxTTL": 31536000
      },
      {
        "PathPattern": "/api/*",
        "TargetOriginId": "ALB-api",
        "ViewerProtocolPolicy": "https-only",
        "CachePolicyId": "4135ea2d-6df8-44a3-9df3-4b5a84be39ad",  // CachingDisabled
        "OriginRequestPolicyId": "216adef6-5c7f-47e4-b989-5492eafa07d3"  // AllViewer
      }
    ]
  }
}
```

### 3. Database Optimization

#### Read Replicas

```bash
# Create read replica
aws rds create-db-instance-read-replica \
  --db-instance-identifier prod-db-replica-1 \
  --source-db-instance-identifier prod-db \
  --db-instance-class db.t3.medium \
  --publicly-accessible false
```

```javascript
// Route read queries to replica
const { Pool } = require('pg');

const writePool = new Pool({ host: 'prod-db.xxx.rds.amazonaws.com' });
const readPool = new Pool({ host: 'prod-db-replica-1.xxx.rds.amazonaws.com' });

async function getResource(id) {
  // Use read replica for queries
  return readPool.query('SELECT * FROM resources WHERE id = $1', [id]);
}

async function createResource(data) {
  // Use primary for writes
  return writePool.query('INSERT INTO resources (name, type) VALUES ($1, $2)', [data.name, data.type]);
}
```

#### Query Result Caching

```sql
-- Enable query result caching in PostgreSQL
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
ALTER SYSTEM SET pg_stat_statements.track = all;

-- Create materialized view for expensive queries
CREATE MATERIALIZED VIEW resource_stats AS
SELECT
  type,
  status,
  COUNT(*) as count,
  AVG(size) as avg_size
FROM resources
GROUP BY type, status;

-- Refresh periodically (via cron job)
REFRESH MATERIALIZED VIEW CONCURRENTLY resource_stats;
```

### 4. Frontend Optimization

#### Code Splitting

```javascript
// React lazy loading
import React, { lazy, Suspense } from 'react';

const Dashboard = lazy(() => import('./pages/Dashboard'));
const Reports = lazy(() => import('./pages/Reports'));

function App() {
  return (
    <Suspense fallback={<Loading />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/reports" element={<Reports />} />
      </Routes>
    </Suspense>
  );
}
```

#### Image Optimization

```bash
# Optimize images during build
npm install --save-dev imagemin imagemin-mozjpeg imagemin-pngquant

# Build script
const imagemin = require('imagemin');
const imageminMozjpeg = require('imagemin-mozjpeg');
const imageminPngquant = require('imagemin-pngquant');

await imagemin(['images/*.{jpg,png}'], {
  destination: 'dist/images',
  plugins: [
    imageminMozjpeg({ quality: 75 }),
    imageminPngquant({ quality: [0.6, 0.8] })
  ]
});
```

#### Service Worker Caching

```javascript
// service-worker.js
const CACHE_NAME = 'app-cache-v1';
const urlsToCache = [
  '/',
  '/static/css/main.css',
  '/static/js/main.js',
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(urlsToCache))
  );
});

self.addEventListener('fetch', (event) => {
  event.respondWith(
    caches.match(event.request)
      .then((response) => response || fetch(event.request))
  );
});
```

### 5. Continuous Performance Testing

```yaml
# .github/workflows/performance-test.yml
name: Performance Tests

on:
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM

jobs:
  performance-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install k6
        run: |
          sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6

      - name: Run performance tests
        run: |
          k6 run --out json=results.json performance/k6/api-load-test.js

      - name: Check performance thresholds
        run: |
          # Fail if P95 latency > 500ms
          P95=$(jq '.metrics.http_req_duration.values["p(95)"]' results.json)
          if (( $(echo "$P95 > 500" | bc -l) )); then
            echo "Performance regression detected: P95 latency $P95ms > 500ms"
            exit 1
          fi

      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: performance-results
          path: results.json
```

---

## Quick Reference Commands

```bash
# Run basic load test
k6 run --vus 50 --duration 5m performance/k6/api-load-test.js

# Run stress test to find breaking point
k6 run performance/k6/stress-test.js

# Run Locust with web UI
locust -f performance/locust/user-journey.py --host=https://api.example.com

# Run Locust headless
locust -f performance/locust/user-journey.py --host=https://api.example.com \
       --users 100 --spawn-rate 10 --run-time 10m --headless

# Monitor database performance
psql -U postgres -c "SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# Monitor Redis performance
redis-cli INFO stats | grep keyspace

# Monitor API metrics
curl -s https://api.example.com/metrics | grep http_request_duration

# Check CloudWatch metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/ApplicationELB \
  --metric-name TargetResponseTime \
  --dimensions Name=LoadBalancer,Value=app/api-lb/xxx \
  --statistics Average \
  --start-time $(date -u -d "1 hour ago" +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300
```

---

## Troubleshooting Performance Issues

### High Latency Checklist

1. ✅ Check database query performance (`pg_stat_statements`)
2. ✅ Verify cache hit ratio (should be > 80%)
3. ✅ Check connection pool utilization
4. ✅ Review slow API endpoints (APM traces)
5. ✅ Verify CDN is serving static assets
6. ✅ Check for N+1 query problems
7. ✅ Review database indexes
8. ✅ Monitor CPU/Memory on application servers

### High Error Rate Checklist

1. ✅ Check application logs for exceptions
2. ✅ Verify database connectivity
3. ✅ Check dependency service status
4. ✅ Review recent deployments
5. ✅ Verify authentication service
6. ✅ Check rate limiting thresholds
7. ✅ Monitor circuit breaker status

### Resource Exhaustion Checklist

1. ✅ Check memory leaks (heap dumps)
2. ✅ Verify connection pool isn't exhausted
3. ✅ Review file descriptor limits
4. ✅ Check for goroutine/thread leaks
5. ✅ Monitor disk space
6. ✅ Review log retention policies

---

## Additional Resources

- **k6 Documentation**: https://k6.io/docs/
- **Locust Documentation**: https://docs.locust.io/
- **Grafana Dashboards**: https://grafana.com/grafana/dashboards/
- **Prometheus Best Practices**: https://prometheus.io/docs/practices/
- **Web Performance Working Group**: https://www.w3.org/webperf/
- **High Performance Browser Networking**: https://hpbn.co/

---

**Last Updated**: 2025-01-01
**Maintained By**: Performance Engineering Team
**Review Cycle**: Quarterly
