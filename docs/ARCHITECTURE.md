# Freightliner Architecture Documentation

## Overview

Freightliner is a production-ready container registry replication service built with enterprise-grade architecture patterns. The application follows a layered architecture with clear separation of concerns, comprehensive observability, and production-ready operational features.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Freightliner Application                    │
├─────────────────────────────────────────────────────────────────┤
│  HTTP Server Layer                                             │
│  ├── Middleware Stack (Logging, Metrics, Recovery, CORS)       │
│  ├── Health Check Endpoints (/health, /ready, /live)           │
│  ├── API Endpoints (Future: /api/v1/*)                         │
│  └── Metrics Endpoint (/metrics)                               │
├─────────────────────────────────────────────────────────────────┤
│  Application Core                                              │
│  ├── Command Structure (CLI + Server)                          │
│  ├── Configuration Management                                  │
│  ├── Logging System (Structured + Basic)                       │
│  └── Metrics Registry (Prometheus)                             │
├─────────────────────────────────────────────────────────────────┤
│  Infrastructure Layer                                          │
│  ├── Registry Clients (ECR, GCR) [In Development]              │
│  ├── Replication Engine [In Development]                       │
│  ├── Security Services [In Development]                        │
│  └── Persistence Layer [In Development]                        │
└─────────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. HTTP Server Layer (`pkg/server/`)

The HTTP server provides the main application interface with production-ready capabilities.

#### Server Structure
```go
type Server struct {
    config     *config.Config
    logger     log.Logger
    metrics    *metrics.Registry
    router     *http.ServeMux
    server     *http.Server
    middleware []Middleware
}
```

#### Middleware Stack
1. **Logging Middleware** - Request/response logging with structured output
2. **Metrics Middleware** - Prometheus metrics collection
3. **Recovery Middleware** - Panic recovery and error handling
4. **CORS Middleware** - Cross-origin request support
5. **Authentication Middleware** - API key validation

#### Health Check System
- **Basic Health** (`/health`) - Simple alive check
- **Readiness** (`/ready`) - Dependency health validation
- **Liveness** (`/live`) - Application responsiveness check
- **System Info** (`/health/system`) - Detailed system metrics

### 2. Logging System (`pkg/helper/log/`)

Enterprise-grade logging with multiple implementations and structured output.

#### Logger Interface
```go
type Logger interface {
    Debug(message string)
    Info(message string)
    Warn(message string)
    Error(message string, err error)
    Fatal(message string, err error)
    Panic(message string, err error)
    WithField(key string, value interface{}) Logger
    WithFields(fields map[string]interface{}) Logger
    WithError(err error) Logger
    WithContext(ctx context.Context) Logger
}
```

#### Implementations
- **BasicLogger** - Simple text-based logging for development
- **StructuredLogger** - JSON logging with trace/span support for production

#### Features
- **Structured Output** - JSON format for log aggregation
- **Context Support** - Trace and span ID extraction
- **Caller Information** - File, line, and function details
- **Stack Traces** - Automatic stack trace on errors/panics
- **Global Management** - Thread-safe global logger access

### 3. Metrics System (`pkg/metrics/`)

Comprehensive Prometheus metrics collection for observability.

#### Registry Structure
```go
type Registry struct {
    // HTTP metrics
    httpRequestsTotal    *prometheus.CounterVec
    httpRequestDuration  *prometheus.HistogramVec
    httpRequestsInFlight prometheus.Gauge
    
    // Replication metrics
    replicationTotal       *prometheus.CounterVec
    replicationDuration    *prometheus.HistogramVec
    replicationBytesTotal  *prometheus.CounterVec
    
    // System metrics
    memoryUsage    prometheus.Gauge
    goroutineCount prometheus.Gauge
    panicTotal     *prometheus.CounterVec
}
```

#### Metric Categories
1. **HTTP Metrics** - Request counts, durations, status codes
2. **Replication Metrics** - Transfer statistics, success rates
3. **Job Metrics** - Background task monitoring
4. **Worker Pool Metrics** - Concurrency and queue monitoring
5. **System Metrics** - Memory, goroutines, GC statistics
6. **Authentication Metrics** - Security event tracking

### 4. Configuration System (`pkg/config/`)

Flexible configuration management with multiple sources and validation.

#### Configuration Structure
```go
type Config struct {
    // Application settings
    LogLevel string
    
    // Server configuration
    Server ServerConfig
    
    // Metrics configuration
    Metrics MetricsConfig
    
    // Registry configurations
    ECR ECRConfig
    GCR GCRConfig
    
    // Feature configurations
    Workers    WorkerConfig
    Encryption EncryptionConfig
    Secrets    SecretsConfig
}
```

#### Configuration Sources (Priority Order)
1. **Command Line Flags** - Highest priority
2. **Environment Variables** - Medium priority
3. **Configuration Files** - YAML/JSON support
4. **Default Values** - Lowest priority

### 5. Command Structure (`cmd/`)

Clean command-line interface with extensible command structure.

#### Command Hierarchy
```
freightliner
├── version       # Version and build information
├── health-check  # Container health validation
├── serve         # HTTP server mode
├── replicate     # Single repository replication [Future]
└── replicate-tree # Tree replication [Future]
```

## Data Flow Architecture

### 1. HTTP Request Flow

```
Client Request
    ↓
Middleware Stack
    ├── Logging (request start)
    ├── Metrics (increment counters)
    ├── Recovery (panic protection)
    ├── CORS (origin validation)
    └── Auth (API key validation)
    ↓
Router/Handler
    ├── Health Checks
    ├── API Endpoints
    └── Metrics Endpoint
    ↓
Response Processing
    ├── Metrics (record duration)
    ├── Logging (request completion)
    └── Client Response
```

### 2. Logging Flow

```
Application Event
    ↓
Logger Interface
    ├── Context Extraction (trace/span)
    ├── Caller Information
    └── Field Aggregation
    ↓
Logger Implementation
    ├── BasicLogger → Text Output
    └── StructuredLogger → JSON Output
    ↓
Output Destination
    ├── STDOUT (container logs)
    ├── File (local development)
    └── Log Aggregation System
```

### 3. Metrics Flow

```
Application Metric
    ↓
Metrics Registry
    ├── Counter Increment
    ├── Histogram Observation
    └── Gauge Setting
    ↓
Prometheus Registry
    ↓
HTTP Metrics Endpoint
    ↓
Monitoring System
    ├── Prometheus Scraping
    ├── Grafana Dashboards
    └── AlertManager Rules
```

## Security Architecture

### 1. Authentication Layer

```go
type AuthMiddleware struct {
    apiKey    string
    enabled   bool
    logger    log.Logger
    metrics   *metrics.Registry
}
```

- **API Key Authentication** - Bearer token validation
- **Rate Limiting** - Request throttling (configurable)
- **CORS Protection** - Origin-based access control
- **Audit Logging** - Security event tracking

### 2. Error Handling

- **Panic Recovery** - Graceful error handling without crashes
- **Structured Errors** - Consistent error response format
- **Security Headers** - Protection against common attacks
- **Input Validation** - Request parameter sanitization

## Deployment Architecture

### 1. Container Deployment

```yaml
# Production container configuration
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: freightliner
        image: freightliner:latest
        ports:
        - containerPort: 8080  # HTTP API
        - containerPort: 2112  # Metrics
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: METRICS_ENABLED
          value: "true"
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
```

### 2. Service Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Monitoring    │    │   Log Aggregation│
│   (Health Check)│    │   (Prometheus)  │    │   (ELK/Fluentd) │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │                      │                      │
    ┌─────▼─────────────────────────▼─────────────────────▼─────┐
    │              Freightliner Instances                      │
    │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐     │
    │  │   Pod   │  │   Pod   │  │   Pod   │  │   Pod   │     │
    │  │  :8080  │  │  :8080  │  │  :8080  │  │  :8080  │     │
    │  │  :2112  │  │  :2112  │  │  :2112  │  │  :2112  │     │
    │  └─────────┘  └─────────┘  └─────────┘  └─────────┘     │
    └───────────────────────────────────────────────────────────┘
```

## Scalability Architecture

### 1. Horizontal Scaling

- **Stateless Design** - No local state dependencies
- **Health Checks** - Kubernetes-ready probes
- **Graceful Shutdown** - Clean connection termination
- **Resource Management** - Configurable limits and requests

### 2. Performance Optimization

- **Connection Pooling** - Reusable HTTP connections
- **Request Batching** - Efficient bulk operations
- **Caching Strategies** - In-memory and distributed caching
- **Asynchronous Processing** - Non-blocking operations

## Observability Architecture

### 1. Three Pillars of Observability

**Metrics (Prometheus)**
- Application performance metrics
- Business logic metrics
- Infrastructure metrics
- Custom metric collection

**Logs (Structured JSON)**
- Request tracing
- Error tracking
- Security events
- Performance insights

**Traces (OpenTelemetry Ready)**
- Request flow tracking
- Service dependency mapping
- Performance bottleneck identification
- Cross-service correlation

### 2. Monitoring Stack Integration

```
Application Metrics → Prometheus → Grafana → AlertManager
Application Logs → Fluentd → Elasticsearch → Kibana
Application Traces → Jaeger → Service Maps → Performance Analysis
```

## Extension Points

### 1. Middleware Extensions

```go
type Middleware func(http.Handler) http.Handler

// Custom middleware can be added
server.AddMiddleware(customAuthMiddleware)
server.AddMiddleware(customRateLimitMiddleware)
```

### 2. Plugin Architecture (Future)

- **Registry Plugins** - Additional registry support
- **Authentication Plugins** - Custom auth providers
- **Storage Plugins** - Different persistence backends
- **Notification Plugins** - Event publishing

## Development Architecture

### 1. Package Organization

```
freightliner/
├── cmd/                    # Command line interfaces
├── pkg/
│   ├── server/            # HTTP server implementation
│   ├── config/            # Configuration management
│   ├── helper/
│   │   └── log/           # Logging system
│   ├── metrics/           # Metrics collection
│   ├── client/            # Registry clients [Future]
│   └── security/          # Security features [Future]
├── docs/                  # Documentation
├── examples/              # Example configurations
└── scripts/               # Development scripts
```

### 2. Testing Architecture

- **Unit Tests** - Individual component testing
- **Integration Tests** - Multi-component interactions
- **Performance Tests** - Load and stress testing
- **Contract Tests** - API contract validation

## Future Architecture Enhancements

### 1. Microservices Evolution

```
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   API Gateway   │  │  Replication    │  │   Security      │
│   Service       │  │   Service       │  │   Service       │
└─────────────────┘  └─────────────────┘  └─────────────────┘
         │                    │                    │
         └────────────────────┼────────────────────┘
                              │
              ┌─────────────────────────────┐
              │     Message Queue           │
              │   (Event Distribution)      │
              └─────────────────────────────┘
```

### 2. Event-Driven Architecture

- **Event Sourcing** - Audit trail and replay capability
- **CQRS Pattern** - Command and Query separation
- **Event Streaming** - Real-time replication updates
- **Distributed State Management** - Cross-service coordination

## Conclusion

The Freightliner architecture provides a solid foundation for production container registry replication with:

- **Production-Ready Components** - Enterprise-grade HTTP server, logging, metrics
- **Scalable Design** - Horizontal scaling and performance optimization
- **Comprehensive Observability** - Metrics, logs, and tracing ready
- **Security First** - Authentication, authorization, and audit capabilities
- **Extensible Framework** - Plugin architecture for future enhancements

The architecture supports both immediate production deployment and future expansion into a comprehensive container registry management platform.