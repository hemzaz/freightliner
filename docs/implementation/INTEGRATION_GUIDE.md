# Integration Guide - Production Features

## Overview

This guide explains how to integrate the newly implemented production features into the existing Freightliner codebase.

## Files to Modify

### 1. Update `pkg/server/server.go`

Add these fields to the `Server` struct:

```go
type Server struct {
	// ... existing fields ...

	// New fields for production features
	autoScaler  *replication.AutoScaler
	rateLimiter *RateLimiter
}
```

### 2. Update `pkg/server/server.go` - NewServer function

Replace the worker pool creation:

```go
// OLD:
workerPool := replication.NewWorkerPool(workerCount, logger)

// NEW:
workerPool := server.createEnhancedWorkerPool(workerCount)
```

### 3. Update `pkg/server/server.go` - registerEndpoints function

Add enhanced endpoints after existing ones:

```go
func (s *Server) registerEndpoints() {
	// ... existing endpoints ...

	// Add enhanced endpoints
	s.registerEnhancedEndpoints()
}
```

### 4. Update `pkg/replication/worker_pool.go`

Add stats field to WorkerPool struct:

```go
type WorkerPool struct {
	// ... existing fields ...

	stats *statsCollector
}
```

Initialize stats in `NewWorkerPool`:

```go
func NewWorkerPool(workerCount int, logger log.Logger) *WorkerPool {
	// ... existing initialization ...

	pool := &WorkerPool{
		// ... existing fields ...
		stats: newStatsCollector(),
	}

	return pool
}
```

### 5. Update `pkg/server/jobs.go`

Add missing methods to Job interface implementations:

```go
// Ensure ReplicateJob and ReplicateTreeJob have these fields:
type ReplicateJob struct {
	// ... existing fields ...

	// Add if missing:
	source      string
	destination string
	tags        []string
	force       bool
	dryRun      bool
}

type ReplicateTreeJob struct {
	// ... existing fields ...

	// Add if missing:
	source      string
	destination string
	options     map[string]interface{}
}
```

### 6. Update `pkg/config/config.go`

Add new configuration fields:

```go
type ServerConfig struct {
	// ... existing fields ...

	// New fields
	RateLimit int `yaml:"rate_limit"` // Requests per minute
}

type WorkersConfig struct {
	// ... existing fields ...

	// New fields for auto-scaling
	AutoScale          bool          `yaml:"auto_scale"`
	MinWorkers         int           `yaml:"min_workers"`
	MaxWorkers         int           `yaml:"max_workers"`
	ScaleCheckInterval time.Duration `yaml:"scale_check_interval"`
}
```

## Testing the Integration

### 1. Build the Project

```bash
cd /Users/elad/PROJ/freightliner
go mod tidy
go build -o freightliner ./cmd/freightliner
```

### 2. Run Unit Tests

```bash
# Test rate limiter
go test ./pkg/server -run TestRateLimiter -v

# Test priority queue
go test ./pkg/replication -run TestPriorityQueue -v

# Test autoscaler
go test ./pkg/replication -run TestAutoScaler -v
```

### 3. Start Server with New Features

```bash
# Create config file
cat > config.yaml <<EOF
server:
  port: 8080
  host: "0.0.0.0"
  rate_limit: 100
  api_key_auth: true
  api_key: "your-secure-api-key"
  enable_cors: true

workers:
  serve_workers: 10
  auto_scale: true
  min_workers: 2
  max_workers: 50
  scale_check_interval: 30s

registries:
  registries:
    - name: dockerhub
      type: dockerhub
      endpoint: "https://registry-1.docker.io"
      auth:
        type: anonymous
EOF

# Start server
./freightliner serve -c config.yaml
```

### 4. Test API Endpoints

```bash
# Set API key
API_KEY="your-secure-api-key"
BASE_URL="http://localhost:8080/api/v1"

# Test system health
curl -H "X-API-Key: $API_KEY" $BASE_URL/system/health

# Test worker pool stats
curl -H "X-API-Key: $API_KEY" $BASE_URL/system/stats

# List registries
curl -H "X-API-Key: $API_KEY" $BASE_URL/registries

# Check registry health
curl -H "X-API-Key: $API_KEY" $BASE_URL/registries/dockerhub/health

# Create a replication job
curl -X POST -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "source_registry": "docker.io",
    "source_repo": "library/nginx",
    "dest_registry": "ecr",
    "dest_repo": "my-project/nginx",
    "tags": ["latest"],
    "dry_run": true
  }' \
  $BASE_URL/replicate

# List jobs
curl -H "X-API-Key: $API_KEY" $BASE_URL/jobs

# Get job details
curl -H "X-API-Key: $API_KEY" $BASE_URL/jobs/{job-id}

# Cancel a job
curl -X POST -H "X-API-Key: $API_KEY" $BASE_URL/jobs/{job-id}/cancel

# Retry a failed job
curl -X POST -H "X-API-Key: $API_KEY" $BASE_URL/jobs/{job-id}/retry
```

### 5. Test Rate Limiting

```bash
# Rapid fire requests to trigger rate limit
for i in {1..150}; do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -H "X-API-Key: $API_KEY" \
    $BASE_URL/system/health
done
# Should see 200s followed by 429s
```

### 6. Test Auto-Scaling

```bash
# Generate load to trigger auto-scaling
for i in {1..100}; do
  curl -X POST -H "X-API-Key: $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{
      "source_registry": "docker.io",
      "source_repo": "library/nginx",
      "dest_registry": "ecr",
      "dest_repo": "test/nginx-'$i'",
      "tags": ["latest"]
    }' \
    $BASE_URL/replicate &
done

# Watch worker pool stats
watch -n 1 "curl -s -H 'X-API-Key: $API_KEY' $BASE_URL/system/stats | jq '.workers'"
```

## Prometheus Integration

### Scrape Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'freightliner'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Example Queries

```promql
# Worker pool utilization
rate(worker_pool_completed_jobs_total[5m])

# API request rate
rate(http_requests_total[5m])

# Error rate
rate(worker_pool_failed_jobs_total[5m]) / rate(worker_pool_completed_jobs_total[5m])

# Queue depth
worker_pool_queued_jobs

# Worker count
worker_pool_active_workers
```

## Grafana Dashboard

### Example Dashboard JSON

```json
{
  "dashboard": {
    "title": "Freightliner Replication",
    "panels": [
      {
        "title": "Worker Pool Utilization",
        "targets": [
          {
            "expr": "worker_pool_active_workers / worker_pool_total_workers",
            "legendFormat": "Utilization"
          }
        ]
      },
      {
        "title": "Job Throughput",
        "targets": [
          {
            "expr": "rate(worker_pool_completed_jobs_total[5m])",
            "legendFormat": "Jobs/sec"
          }
        ]
      },
      {
        "title": "Queue Depth",
        "targets": [
          {
            "expr": "worker_pool_queued_jobs",
            "legendFormat": "Queued Jobs"
          }
        ]
      }
    ]
  }
}
```

## Docker Deployment

### Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o freightliner ./cmd/freightliner

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/freightliner .
COPY config.yaml .

EXPOSE 8080

CMD ["./freightliner", "serve", "-c", "config.yaml"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  freightliner:
    build: .
    ports:
      - "8080:8080"
    environment:
      - FREIGHTLINER_SERVER_API_KEY=${API_KEY}
      - FREIGHTLINER_WORKERS_AUTO_SCALE=true
      - FREIGHTLINER_WORKERS_MIN=2
      - FREIGHTLINER_WORKERS_MAX=50
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - ./checkpoints:/app/checkpoints
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
    depends_on:
      - freightliner

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana
    depends_on:
      - prometheus

volumes:
  grafana-storage:
```

## Kubernetes Deployment

### Deployment YAML

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: freightliner
spec:
  replicas: 3
  selector:
    matchLabels:
      app: freightliner
  template:
    metadata:
      labels:
        app: freightliner
    spec:
      containers:
      - name: freightliner
        image: freightliner:latest
        ports:
        - containerPort: 8080
        env:
        - name: FREIGHTLINER_SERVER_API_KEY
          valueFrom:
            secretKeyRef:
              name: freightliner-secrets
              key: api-key
        - name: FREIGHTLINER_WORKERS_AUTO_SCALE
          value: "true"
        - name: FREIGHTLINER_WORKERS_MIN
          value: "2"
        - name: FREIGHTLINER_WORKERS_MAX
          value: "50"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: freightliner
spec:
  selector:
    app: freightliner
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer

---
apiVersion: v1
kind: Secret
metadata:
  name: freightliner-secrets
type: Opaque
stringData:
  api-key: "your-secure-api-key"
```

## Troubleshooting

### Rate Limiter Not Working

Check configuration:
```yaml
server:
  rate_limit: 100  # Must be > 0 to enable
```

### Auto-Scaler Not Scaling

1. Check configuration:
```yaml
workers:
  auto_scale: true
  min_workers: 2
  max_workers: 50
```

2. Check logs:
```bash
# Look for auto-scaler messages
grep "autoscaler" freightliner.log
```

3. Monitor metrics:
```bash
curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/v1/system/stats
```

### Jobs Not Running

1. Check worker pool status:
```bash
curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/v1/system/stats
```

2. Check job status:
```bash
curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/v1/jobs
```

3. Check server logs for errors

## Performance Tuning

### High Throughput

```yaml
server:
  rate_limit: 1000  # Increase limit

workers:
  serve_workers: 50  # Start with more workers
  auto_scale: true
  min_workers: 20    # Higher minimum
  max_workers: 100   # Higher maximum
```

### Low Resource Usage

```yaml
server:
  rate_limit: 50

workers:
  serve_workers: 5
  auto_scale: true
  min_workers: 1
  max_workers: 10
```

## Next Steps

1. **Add Unit Tests**: Test new components
2. **Integration Tests**: End-to-end API testing
3. **Load Testing**: Validate performance targets
4. **Documentation**: Update user guides
5. **Monitoring**: Set up alerts in Prometheus

## Support

For issues or questions:
- Check logs: `./freightliner serve -v` (verbose mode)
- Review metrics: `http://localhost:8080/metrics`
- System health: `http://localhost:8080/api/v1/system/health`
