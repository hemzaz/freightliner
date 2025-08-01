# Freightliner Configuration Guide

## Overview

Freightliner supports flexible configuration through multiple sources with a clear priority order. This guide covers all configuration options, environment variables, and usage patterns for production deployments.

## Configuration Sources

Configuration is loaded in the following priority order (highest to lowest):

1. **Command Line Flags** - Highest priority
2. **Environment Variables** - Medium priority  
3. **Configuration Files** - YAML/JSON format
4. **Default Values** - Lowest priority

## Configuration Structure

### Complete Configuration Example

```yaml
# config.yaml - Complete configuration example
log_level: info

# Server configuration
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 60s
  shutdown_timeout: 15s
  
  # TLS configuration
  tls_enabled: false
  tls_cert_file: "/path/to/cert.pem"
  tls_key_file: "/path/to/key.pem"
  
  # Authentication
  api_key_auth: false
  api_key: "your-secret-api-key"
  
  # CORS
  allowed_origins:
    - "http://localhost:3000"
    - "https://your-domain.com"
  
  # Endpoints
  health_check_path: "/health"
  metrics_path: "/metrics"
  replicate_path: "/api/v1/replicate"
  tree_replicate_path: "/api/v1/replicate-tree"
  status_path: "/api/v1/status"

# Metrics configuration
metrics:
  enabled: true
  port: 2112
  path: "/metrics"
  namespace: "freightliner"

# Registry configurations
ecr:
  region: "us-west-2"
  account_id: "123456789012"

gcr:
  project: "my-gcp-project"
  location: "us"

# Worker configuration
workers:
  replicate_workers: 4
  serve_workers: 8
  auto_detect: true

# Encryption configuration
encryption:
  enabled: false
  customer_managed_keys: false
  aws_kms_key_id: "alias/freightliner-key"
  gcp_kms_key_id: "projects/my-project/locations/global/keyRings/freightliner/cryptoKeys/image-encryption"
  gcp_key_ring: "freightliner"
  gcp_key_name: "image-encryption"
  envelope_encryption: true

# Secrets configuration
secrets:
  use_secrets_manager: false
  secrets_manager_type: "aws"
  aws_secret_region: "us-west-2"
  gcp_secret_project: "my-gcp-project"
  gcp_credentials_file: "/path/to/credentials.json"
  registry_creds_secret: "freightliner-registry-credentials"
  encryption_keys_secret: "freightliner-encryption-keys"

# Checkpoint configuration
checkpoint:
  directory: "${HOME}/.freightliner/checkpoints"
  id: ""

# Tree replication defaults
tree_replicate:
  workers: 0
  exclude_repos: []
  exclude_tags: []
  include_tags: []
  dry_run: false
  force: false
  enable_checkpoint: false
  checkpoint_dir: "${HOME}/.freightliner/checkpoints"
  resume_id: ""
  skip_completed: true
  retry_failed: true

# Single replication defaults
replicate:
  force: false
  dry_run: false
  tags: []
```

## Environment Variables

### Server Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `HOST` | `0.0.0.0` | Server bind address |
| `READ_TIMEOUT` | `30s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `60s` | HTTP write timeout |
| `SHUTDOWN_TIMEOUT` | `15s` | Graceful shutdown timeout |
| `TLS_ENABLED` | `false` | Enable TLS/HTTPS |
| `TLS_CERT_FILE` | - | TLS certificate file path |
| `TLS_KEY_FILE` | - | TLS private key file path |
| `API_KEY_AUTH` | `false` | Enable API key authentication |
| `API_KEY` | - | API key for authentication |
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins (comma-separated) |

### Metrics Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `METRICS_ENABLED` | `true` | Enable metrics collection |
| `METRICS_PORT` | `2112` | Metrics server port |
| `METRICS_PATH` | `/metrics` | Metrics endpoint path |
| `METRICS_NAMESPACE` | `freightliner` | Prometheus metrics namespace |

### Logging Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error, fatal) |

### Registry Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `ECR_REGION` | `us-west-2` | AWS ECR region |
| `ECR_ACCOUNT_ID` | - | AWS account ID (auto-detected if empty) |
| `GCR_PROJECT` | - | Google Cloud project ID |
| `GCR_LOCATION` | `us` | GCR location (us, eu, asia) |

### Worker Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `REPLICATE_WORKERS` | `0` | Number of replication workers (0 = auto) |
| `SERVE_WORKERS` | `0` | Number of server workers (0 = auto) |
| `AUTO_DETECT_WORKERS` | `true` | Auto-detect optimal worker count |

### Security Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `ENCRYPT_IMAGES` | `false` | Enable image encryption |
| `CUSTOMER_MANAGED_KEYS` | `false` | Use customer-managed encryption keys |
| `AWS_KMS_KEY_ID` | - | AWS KMS key ID for encryption |
| `GCP_KMS_KEY_ID` | - | GCP KMS key ID for encryption |
| `USE_SECRETS_MANAGER` | `false` | Use cloud secrets manager |
| `SECRETS_MANAGER_TYPE` | `aws` | Secrets manager type (aws, gcp) |

## Command Line Flags

### Global Flags

```bash
freightliner [command] [flags]

Global Flags:
  --log-level string          Log level (debug, info, warn, error, fatal) (default "info")
  --ecr-region string         AWS region for ECR (default "us-west-2")
  --ecr-account string        AWS account ID for ECR (empty uses default from credentials)
  --gcr-project string        GCP project for GCR
  --gcr-location string       GCR location (us, eu, asia) (default "us")
  --replicate-workers int     Number of concurrent workers for replication (0 = auto-detect)
  --serve-workers int         Number of concurrent workers for server mode (0 = auto-detect)
  --auto-detect-workers       Auto-detect optimal worker count based on system resources (default true)
  --encrypt                   Enable image encryption
  --customer-key              Use customer-managed encryption keys
  --aws-kms-key string        AWS KMS key ID for encryption
  --gcp-kms-key string        GCP KMS key ID for encryption
  --use-secrets-manager       Use cloud provider secrets manager for credentials
  --secrets-manager-type string   Type of secrets manager to use (aws, gcp) (default "aws")
```

### Server Command Flags

```bash
freightliner serve [flags]

Server Flags:
  --port int                  Server listening port (default 8080)
  --host string               Server host (default "0.0.0.0")
  --read-timeout duration     HTTP server read timeout (default 30s)
  --write-timeout duration    HTTP server write timeout (default 60s)
  --shutdown-timeout duration Server shutdown timeout (default 15s)
  --tls                       Enable TLS
  --tls-cert string           TLS certificate file
  --tls-key string            TLS key file
  --api-key-auth              Enable API key authentication
  --api-key string            API key for authentication
  --allowed-origins strings   Allowed CORS origins (default [*])
  --metrics                   Enable metrics collection (default true)
  --metrics-port int          Metrics server port (default 2112)
  --metrics-path string       Metrics endpoint path (default "/metrics")
```

## Configuration File Examples

### Development Configuration

```yaml
# config-dev.yaml
log_level: debug

server:
  port: 8080
  api_key_auth: false
  allowed_origins:
    - "http://localhost:3000"

metrics:
  enabled: true
  port: 2112

workers:
  auto_detect: true

encryption:
  enabled: false
```

### Production Configuration

```yaml
# config-prod.yaml
log_level: info

server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 60s
  shutdown_timeout: 15s
  tls_enabled: true
  tls_cert_file: "/etc/ssl/certs/freightliner.crt"
  tls_key_file: "/etc/ssl/private/freightliner.key"
  api_key_auth: true
  api_key: "${API_KEY}"  # From environment
  allowed_origins:
    - "https://my-domain.com"

metrics:
  enabled: true
  port: 2112
  namespace: "freightliner"

ecr:
  region: "us-west-2"

gcr:
  project: "my-production-project"
  location: "us"

workers:
  replicate_workers: 8
  serve_workers: 16
  auto_detect: false

encryption:
  enabled: true
  customer_managed_keys: true
  aws_kms_key_id: "alias/freightliner-prod"
  envelope_encryption: true

secrets:
  use_secrets_manager: true
  secrets_manager_type: "aws"
  aws_secret_region: "us-west-2"
  registry_creds_secret: "prod/freightliner/registry-creds"
  encryption_keys_secret: "prod/freightliner/encryption-keys"
```

### High Availability Configuration

```yaml
# config-ha.yaml
log_level: info

server:
  port: 8080
  read_timeout: 30s
  write_timeout: 120s
  shutdown_timeout: 30s
  tls_enabled: true
  api_key_auth: true

metrics:
  enabled: true
  port: 2112

workers:
  replicate_workers: 16
  serve_workers: 32

tree_replicate:
  enable_checkpoint: true
  checkpoint_dir: "/data/checkpoints"
  skip_completed: true
  retry_failed: true

encryption:
  enabled: true
  customer_managed_keys: true
  envelope_encryption: true

secrets:
  use_secrets_manager: true
```

## Docker Configuration

### Environment Variables in Docker

```bash
docker run -d \
  --name freightliner \
  -p 8080:8080 \
  -p 2112:2112 \
  -e LOG_LEVEL=info \
  -e PORT=8080 \
  -e METRICS_ENABLED=true \
  -e METRICS_PORT=2112 \
  -e API_KEY_AUTH=true \
  -e API_KEY=your-secret-key \
  -e ECR_REGION=us-west-2 \
  -e GCR_PROJECT=my-project \
  freightliner serve
```

### Docker Compose Configuration

```yaml
# docker-compose.yml
version: '3.8'

services:
  freightliner:
    image: freightliner:latest
    ports:
      - "8080:8080"
      - "2112:2112"
    environment:
      - LOG_LEVEL=info
      - PORT=8080
      - METRICS_ENABLED=true
      - METRICS_PORT=2112
      - API_KEY_AUTH=true
      - API_KEY=${API_KEY}
      - ECR_REGION=us-west-2
      - GCR_PROJECT=${GCR_PROJECT}
      - ENCRYPT_IMAGES=true
      - USE_SECRETS_MANAGER=true
    volumes:
      - ./config.yaml:/etc/freightliner/config.yaml:ro
      - checkpoints:/data/checkpoints
    command: ["serve", "--config", "/etc/freightliner/config.yaml"]
    healthcheck:
      test: ["CMD", "./freightliner", "health-check"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  checkpoints:
```

## Kubernetes Configuration

### ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: freightliner-config
data:
  config.yaml: |
    log_level: info
    server:
      port: 8080
      api_key_auth: true
      allowed_origins:
        - "https://my-domain.com"
    metrics:
      enabled: true
      port: 2112
    encryption:
      enabled: true
      customer_managed_keys: true
    secrets:
      use_secrets_manager: true
      secrets_manager_type: "aws"
```

### Secret

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: freightliner-secrets
type: Opaque
data:
  api-key: <base64-encoded-api-key>
  aws-kms-key-id: <base64-encoded-kms-key-id>
```

### Deployment

```yaml
# deployment.yaml
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
        - containerPort: 2112
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
        - name: checkpoints
          mountPath: /data/checkpoints
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: freightliner-config
      - name: checkpoints
        persistentVolumeClaim:
          claimName: freightliner-checkpoints
```

## Configuration Validation

### Runtime Validation

The application validates configuration at startup:

```go
// Configuration validation examples
if config.Server.Port < 1 || config.Server.Port > 65535 {
    log.Fatal("Invalid port number", nil)
}

if config.Server.APIKeyAuth && config.Server.APIKey == "" {
    log.Fatal("API key required when authentication is enabled", nil)
}

if config.Encryption.Enabled && config.Encryption.CustomerManagedKeys {
    if config.Encryption.AWSKMSKeyID == "" && config.Encryption.GCPKMSKeyID == "" {
        log.Fatal("KMS key required when customer-managed encryption is enabled", nil)
    }
}
```

### Environment Variable Expansion

The configuration system supports environment variable expansion:

```yaml
# Environment variable expansion examples
checkpoint:
  directory: "${HOME}/.freightliner/checkpoints"  # Expands to user home

secrets:
  registry_creds_secret: "${ENVIRONMENT}/freightliner/creds"  # e.g., "prod/freightliner/creds"

server:
  api_key: "${API_KEY}"  # Expands from environment
```

## Best Practices

### Production Configuration

1. **Use Configuration Files** - Store non-sensitive configuration in YAML files
2. **Environment Variables for Secrets** - Use env vars for sensitive data
3. **Enable Authentication** - Always use API key auth in production
4. **Configure TLS** - Use HTTPS in production environments
5. **Set Resource Limits** - Configure appropriate timeouts and worker counts
6. **Enable Metrics** - Always enable metrics collection for monitoring
7. **Use Secrets Managers** - Store sensitive data in cloud secret managers

### Security Configuration

1. **Strong API Keys** - Use cryptographically secure random keys
2. **Restrict CORS Origins** - Don't use wildcard origins in production
3. **Enable Encryption** - Use customer-managed keys for sensitive data
4. **Audit Logging** - Enable comprehensive request logging
5. **Network Security** - Use TLS and restrict network access

### Monitoring Configuration

1. **Structured Logging** - Use JSON log format for aggregation
2. **Comprehensive Metrics** - Enable all metric categories
3. **Health Checks** - Configure liveness and readiness probes
4. **Alerting Rules** - Set up alerts based on metrics
5. **Distributed Tracing** - Enable trace correlation for debugging

## Troubleshooting

### Common Configuration Issues

1. **Port Conflicts** - Ensure ports 8080 and 2112 are available
2. **Permission Issues** - Verify TLS certificate file permissions
3. **Environment Variables** - Check variable names and values
4. **YAML Syntax** - Validate YAML configuration files
5. **Resource Limits** - Ensure sufficient memory and CPU

### Configuration Debugging

Enable debug logging to see configuration loading:

```bash
freightliner serve --log-level=debug --config=config.yaml
```

This will output detailed information about:
- Configuration file loading
- Environment variable expansion
- Default value application
- Validation results

For more detailed troubleshooting, see the [Operations Guide](OPERATIONS.md).