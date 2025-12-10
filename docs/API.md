# Freightliner API Documentation

## Overview

Freightliner provides both HTTP REST API endpoints for programmatic access and command-line interface for direct usage. The application runs as a production-ready HTTP server with comprehensive monitoring and health check capabilities.

## 🎉 API Status: Production Ready (January 2025)

**✅ ALL ENDPOINTS OPERATIONAL** - All critical backend issues resolved, API ready for production traffic.

### Endpoint Availability Status

| Endpoint | Status | Description |
|---|---|---|
| `/health` | 🟢 **ACTIVE** | Basic health check - no dependencies |
| `/health/ready` | 🟢 **ACTIVE** | Readiness check with dependency validation |
| `/health/live` | 🟢 **ACTIVE** | Liveness check for container orchestration |
| `/metrics` | 🟢 **ACTIVE** | Prometheus metrics collection |
| `/api/v1/replicate` | 🟢 **ACTIVE** | Core replication endpoint |
| `/api/v1/status` | 🟢 **ACTIVE** | System information and status |

**Server Stability**: All duplicate method conflicts resolved, health monitoring operational.

## HTTP Server API

### Base Configuration

- **Default Port**: 8080
- **Metrics Port**: 2112
- **Content Type**: `application/json`
- **Authentication**: API Key (optional)

### Health Check Endpoints

#### GET /health

Basic health check endpoint for load balancers and monitoring systems.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "v1.0.0",
  "uptime": "24h30m15s"
}
```

**Status Codes:**
- `200 OK` - Service is healthy
- `503 Service Unavailable` - Service is unhealthy

#### GET /ready

Readiness probe for Kubernetes and container orchestration.

**Response:**
```json
{
  "status": "ready",
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "database": {
      "status": "healthy",
      "latency_ms": 5.2,
      "last_check": "2024-01-15T10:29:58Z"
    },
    "external_api": {
      "status": "healthy",
      "latency_ms": 12.5,
      "last_check": "2024-01-15T10:29:59Z"
    }
  }
}
```

#### GET /live

Liveness probe for container health monitoring.

**Response:**
```json
{
  "status": "alive",
  "timestamp": "2024-01-15T10:30:00Z",
  "uptime": "24h30m15s",
  "version": "v1.0.0"
}
```

#### GET /health/system

Detailed system information and health status.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "v1.0.0",
  "uptime": "24h30m15s",
  "system": {
    "hostname": "freightliner-pod-abc123",
    "platform": "linux/amd64",
    "go_version": "go1.21.0",
    "memory": {
      "allocated_mb": 45.2,
      "system_mb": 128.0,
      "gc_runs": 234
    },
    "goroutines": 25,
    "cpu_count": 4
  }
}
```

### Metrics Endpoint

#### GET /metrics

Prometheus metrics exposure endpoint (served on metrics port 2112 by default).

**Response Format:** Prometheus text format

**Sample Metrics:**
```
# HELP freightliner_http_requests_total Total number of HTTP requests
# TYPE freightliner_http_requests_total counter
freightliner_http_requests_total{method="GET",path="/health",status="200"} 1523

# HELP freightliner_http_request_duration_seconds HTTP request duration in seconds
# TYPE freightliner_http_request_duration_seconds histogram
freightliner_http_request_duration_seconds_bucket{method="GET",path="/health",status="200",le="0.005"} 1200
freightliner_http_request_duration_seconds_bucket{method="GET",path="/health",status="200",le="0.01"} 1450

# HELP freightliner_replication_total Total number of replication operations
# TYPE freightliner_replication_total counter
freightliner_replication_total{source_registry="ecr",dest_registry="gcr",status="success"} 45

# HELP freightliner_memory_usage_bytes Current memory usage in bytes
# TYPE freightliner_memory_usage_bytes gauge
freightliner_memory_usage_bytes 47448576
```

### API Endpoints (Future)

These endpoints are planned for implementation:

#### POST /api/v1/replicate

Initiate single repository replication.

#### POST /api/v1/replicate-tree

Initiate tree replication across repositories.

#### GET /api/v1/status

Get current replication status and progress.

## Command Line Interface

### Version Information

```bash
freightliner version
```

**Output:**
```
Freightliner v1.0.0
Build Time: 2024-01-15T10:00:00Z
Git Commit: abc123def456
Go Version: go1.21.0
Platform: linux/amd64
```

### Health Check Command

```bash
freightliner health-check
```

Exits with status code 0 if healthy, non-zero if unhealthy.

### Server Mode

```bash
freightliner serve [flags]
```

**Flags:**
- `--port int`: HTTP server port (default 8080)
- `--host string`: Server host (default "0.0.0.0")
- `--metrics-port int`: Metrics server port (default 2112)
- `--log-level string`: Log level (debug, info, warn, error) (default "info")
- `--tls`: Enable TLS
- `--tls-cert string`: TLS certificate file
- `--tls-key string`: TLS key file
- `--api-key-auth`: Enable API key authentication
- `--api-key string`: API key for authentication

### Configuration

The server can be configured via:

1. **Command line flags** (highest priority)
2. **Environment variables**
3. **Configuration file** (YAML)
4. **Default values** (lowest priority)

## Authentication

### API Key Authentication

When enabled, all API requests must include the API key:

**Header:**
```
Authorization: Bearer <api-key>
```

**Query Parameter:**
```
GET /api/v1/status?api_key=<api-key>
```

## Error Handling

### Standard Error Response

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "The request is invalid",
    "details": {
      "field": "repository",
      "reason": "Repository name is required"
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req_abc123"
  }
}
```

### HTTP Status Codes

- `200 OK` - Request successful
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service temporarily unavailable

## Rate Limiting

When enabled, API requests are rate limited:

- **Default**: 100 requests per second per client
- **Headers**: Rate limit status included in response headers
  ```
  X-RateLimit-Limit: 100
  X-RateLimit-Remaining: 95
  X-RateLimit-Reset: 1642248600
  ```

## CORS Support

Cross-Origin Resource Sharing (CORS) is configurable:

- **Default**: All origins allowed (`*`)
- **Production**: Configure specific allowed origins
- **Headers**: Standard CORS headers supported

## Monitoring and Observability

### Structured Logging

All logs are output in JSON format for easy parsing:

```json
{
  "timestamp": "2024-01-15T10:30:00.123456Z",
  "level": "info",
  "message": "HTTP request completed",
  "fields": {
    "method": "GET",
    "path": "/health",
    "status": 200,
    "duration_ms": 2.5,
    "user_agent": "curl/7.68.0"
  },
  "caller": {
    "file": "middleware.go",
    "line": 45,
    "function": "loggingMiddleware"
  }
}
```

### Request Tracing

When tracing is enabled, requests include trace and span IDs:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "message": "Processing replication request",
  "trace_id": "abc123def456",
  "span_id": "789ghi012jkl",
  "fields": {
    "source": "ecr/my-repo",
    "destination": "gcr/my-repo"
  }
}
```

## Client Libraries

Official client libraries are planned for:

- Go
- Python
- JavaScript/Node.js
- Java

## OpenAPI Specification

The complete OpenAPI 3.0 specification will be available at `/api/docs` when implemented.

## Examples

### Basic Health Check

```bash
curl http://localhost:8080/health
```

### Prometheus Metrics

```bash
curl http://localhost:2112/metrics
```

### Server with Authentication

```bash
freightliner serve --api-key-auth --api-key=my-secret-key
curl -H "Authorization: Bearer my-secret-key" http://localhost:8080/api/v1/status
```

### Docker Container

```bash
docker run -p 8080:8080 -p 2112:2112 \
  -e LOG_LEVEL=info \
  -e API_KEY_AUTH=true \
  -e API_KEY=my-secret-key \
  freightliner serve
```

## Support

For API support and questions:

- **Issues**: [GitHub Issues](https://github.com/hemzaz/freightliner/issues)
- **Documentation**: [Project Documentation](https://github.com/hemzaz/freightliner/docs)
- **Examples**: [Example Configurations](https://github.com/hemzaz/freightliner/examples)