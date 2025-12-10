# Freightliner Quick Start

**Get running in 5 minutes** ⚡

## What is Freightliner?

Freightliner is a production-ready container registry replication and management tool that enables:
- **Multi-registry sync**: AWS ECR ↔ GCR ↔ Docker Hub ↔ Harbor ↔ Quay ↔ GitLab ↔ GitHub ↔ Azure
- **Enterprise security**: AES-256-GCM encryption, KMS integration, vulnerability scanning
- **Operational excellence**: Prometheus metrics, health checks, checkpoint/resume
- **Developer-friendly**: Comprehensive CLI, HTTP API, YAML configuration

## Prerequisites

- **Go 1.25+** (for building from source)
- **Docker** (optional, for container deployment)
- **AWS CLI** (for ECR authentication)
- **gcloud CLI** (for GCR authentication)

## Installation

### Option 1: Build from Source

```bash
# Clone repository
git clone https://github.com/hemzaz/freightliner.git
cd freightliner

# Build binary
make build

# Verify installation
./bin/freightliner version --banner
```

### Option 2: Docker

```bash
# Pull image
docker pull freightliner:latest

# Run container
docker run -d -p 8080:8080 \
  -e FREIGHTLINER_ECR_REGION=us-east-1 \
  -e FREIGHTLINER_GCR_PROJECT=my-project \
  freightliner:latest
```

### Option 3: Pre-built Binary

```bash
# Download latest release
wget https://github.com/hemzaz/freightliner/releases/latest/download/freightliner-linux-amd64

# Make executable
chmod +x freightliner-linux-amd64
mv freightliner-linux-amd64 /usr/local/bin/freightliner

# Verify
freightliner version
```

## 1. Authentication Setup (2 minutes)

### AWS ECR

```bash
# Configure AWS credentials
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-west-2

# Or use AWS CLI login
aws ecr get-login-password --region us-west-2 | \
  docker login --username AWS --password-stdin \
  123456789012.dkr.ecr.us-west-2.amazonaws.com
```

### Google Container Registry

```bash
# Configure GCP credentials
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json

# Or use gcloud
gcloud auth configure-docker
```

### Docker Hub

```bash
# Login to Docker Hub
freightliner login docker.io
# Enter username and password when prompted
```

## 2. Basic Operations (3 minutes)

### Single Image Replication

```bash
# Replicate single image
freightliner replicate \
  docker.io/library/alpine:latest \
  gcr.io/my-project/alpine:latest

# With specific tags
freightliner replicate \
  docker.io/library/nginx \
  gcr.io/my-project/nginx \
  --tags "1.20,1.21,latest"

# Dry-run (preview without copying)
freightliner replicate \
  docker.io/library/redis:latest \
  gcr.io/my-project/redis:latest \
  --dry-run
```

### Bulk Tree Replication

```bash
# Replicate entire repository tree
freightliner replicate-tree \
  ecr/my-company \
  gcr.io/my-project \
  --workers 10

# With checkpoint for large migrations
freightliner replicate-tree \
  ecr/my-company \
  gcr.io/my-project \
  --enable-checkpoint \
  --workers 15 \
  --exclude-tag "dev-*,test-*"

# Resume interrupted replication
freightliner checkpoint list
freightliner replicate-tree \
  ecr/my-company \
  gcr.io/my-project \
  --resume-id abc123 \
  --skip-completed \
  --retry-failed
```

### Image Inspection

```bash
# Inspect image without pulling
freightliner inspect docker.io/library/nginx:latest

# Show raw manifest
freightliner inspect docker.io/library/nginx:latest --raw

# Show container configuration
freightliner inspect docker.io/library/nginx:latest --config

# List all tags
freightliner list-tags docker.io/library/nginx --limit 10

# Show layer information
freightliner layers docker.io/library/nginx:latest

# Calculate manifest digest
freightliner manifest digest docker.io/library/nginx:latest
```

## 3. Configuration-Based Sync (5 minutes)

Create a sync configuration file:

```yaml
# sync-config.yaml
images:
  # Example 1: Sync latest 5 tags with semver filtering
  - source: docker.io/library/nginx
    destination: gcr.io/my-project/nginx
    tags:
      latest_n: 5
      semver_constraint: ">=1.20.0"

  # Example 2: Sync specific tag patterns
  - source: ecr/my-app
    destination: gcr.io/my-project/my-app
    tags:
      patterns: ["v*", "prod-*", "release-*"]
      exclude_patterns: ["*-dev", "*-test"]

  # Example 3: Sync all tags matching regex
  - source: docker.io/library/redis
    destination: gcr.io/my-project/redis
    tags:
      regex_patterns: ["^[0-9]+\\.[0-9]+\\.[0-9]+$"]

# Global settings
deduplication_enabled: true
parallel_workers: 8
retry_failed: true
batch_size: 10
```

Run the sync:

```bash
# Execute sync configuration
freightliner sync --config sync-config.yaml

# Dry-run mode
freightliner sync --config sync-config.yaml --dry-run

# Override parallel workers
freightliner sync --config sync-config.yaml --parallel 12
```

## 4. Security Operations (5 minutes)

### Vulnerability Scanning

```bash
# Basic vulnerability scan
freightliner scan gcr.io/my-project/app:v1.2.0

# Fail on critical vulnerabilities
freightliner scan gcr.io/my-project/app:v1.2.0 \
  --fail-on critical

# Generate SARIF report for GitHub
freightliner scan gcr.io/my-project/app:v1.2.0 \
  --format sarif \
  --output scan-results.sarif

# Use Grype scanner
freightliner scan gcr.io/my-project/app:v1.2.0 \
  --use-grype \
  --only-fixed \
  --db-update
```

### SBOM Generation

```bash
# Generate SPDX format SBOM
freightliner sbom gcr.io/my-project/app:v1.2.0 \
  --format spdx \
  --output sbom.json

# Generate CycloneDX format
freightliner sbom gcr.io/my-project/app:v1.2.0 \
  --format cyclonedx \
  --output sbom-cyclonedx.json

# Include file catalog and scan secrets
freightliner sbom gcr.io/my-project/app:v1.2.0 \
  --format syft-json \
  --output sbom-full.json \
  --include-files \
  --scan-secrets
```

## 5. HTTP Server Mode (10 minutes)

### Create Server Configuration

```yaml
# server-config.yaml
log_level: info

# Registry configuration
ecr:
  region: us-west-2
  account_id: "123456789012"

gcr:
  project: my-gcp-project
  location: us

# Worker configuration
workers:
  replicate_workers: 10
  serve_workers: 5
  auto_detect: true

# Server settings
server:
  host: localhost
  port: 8080
  tls_enabled: true
  tls_cert_file: /etc/ssl/certs/server.crt
  tls_key_file: /etc/ssl/private/server.key
  api_key_auth: true
  api_key: your-secret-api-key
  enable_cors: true
  allowed_origins:
    - "https://example.com"
    - "https://app.example.com"

# Metrics configuration
metrics:
  enabled: true
  port: 2112
  path: /metrics
```

### Start Server

```bash
# Start server with configuration file
freightliner serve --config server-config.yaml

# Start with environment variables
export FREIGHTLINER_SERVER_PORT=8080
export FREIGHTLINER_API_KEY=your-secret-key
freightliner serve

# Start with command-line flags
freightliner serve \
  --port 8080 \
  --tls \
  --tls-cert /etc/ssl/certs/server.crt \
  --tls-key /etc/ssl/private/server.key \
  --api-key-auth \
  --api-key $API_KEY
```

### Test API Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready

# Liveness check
curl http://localhost:8080/live

# System status
curl -H "X-API-Key: your-secret-key" \
  http://localhost:8080/api/v1/status

# Trigger replication via API
curl -X POST \
  -H "X-API-Key: your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{
    "source": "docker.io/library/nginx:latest",
    "destination": "gcr.io/my-project/nginx:latest"
  }' \
  http://localhost:8080/api/v1/replicate

# Prometheus metrics
curl http://localhost:2112/metrics
```

## 6. Load Configuration from URL

Freightliner supports loading configuration from HTTP/HTTPS URLs:

```bash
# Load config from remote URL
freightliner serve --config https://config.example.com/freightliner.yaml

# With authentication (using environment variable)
export CONFIG_URL=https://user:pass@config.example.com/freightliner.yaml
freightliner serve --config $CONFIG_URL

# Override config with environment variables
FREIGHTLINER_LOG_LEVEL=debug \
FREIGHTLINER_REPLICATE_WORKERS=20 \
freightliner serve --config https://config.example.com/freightliner.yaml
```

## 7. Checkpoint Management

For large-scale migrations, use checkpoints to resume interrupted operations:

```bash
# List all checkpoints
freightliner checkpoint list

# Show checkpoint details
freightliner checkpoint show <checkpoint-id>

# Delete checkpoint
freightliner checkpoint delete <checkpoint-id>

# Export checkpoint to file
freightliner checkpoint export <checkpoint-id> checkpoint-backup.json

# Import checkpoint from file
freightliner checkpoint import checkpoint-backup.json

# Resume replication from checkpoint
freightliner replicate-tree \
  ecr/my-company \
  gcr.io/my-project \
  --resume-id <checkpoint-id> \
  --skip-completed \
  --retry-failed
```

## 8. Credential Management

```bash
# Login to registry
freightliner login docker.io
freightliner login gcr.io

# Login with credentials
freightliner login docker.io \
  --username myuser \
  --password mypassword

# List stored credentials
freightliner auth list

# Test registry credentials
freightliner auth test docker.io
freightliner auth test gcr.io/my-project

# Logout from registry
freightliner logout docker.io

# Logout from all registries
freightliner logout --all
```

## 9. Advanced Features

### Encryption

```bash
# Enable encryption with AWS KMS
freightliner replicate \
  docker.io/library/alpine:latest \
  ecr/my-encrypted-repo:latest \
  --encrypt \
  --customer-key \
  --aws-kms-key arn:aws:kms:us-west-2:123456789012:key/abc123

# Enable encryption with GCP KMS
freightliner replicate \
  docker.io/library/alpine:latest \
  gcr.io/my-project/alpine:latest \
  --encrypt \
  --customer-key \
  --gcp-kms-key projects/my-project/locations/us/keyRings/my-ring/cryptoKeys/my-key
```

### Secrets Management

```bash
# Use AWS Secrets Manager
freightliner serve \
  --use-secrets-manager \
  --secrets-manager-type aws \
  --aws-secret-region us-west-2 \
  --registry-creds-secret freightliner-registry-creds

# Use Google Secret Manager
freightliner serve \
  --use-secrets-manager \
  --secrets-manager-type gcp \
  --gcp-secret-project my-project \
  --registry-creds-secret freightliner-registry-creds
```

### Worker Auto-Scaling

```bash
# Auto-detect optimal worker count
freightliner replicate-tree \
  ecr/my-company \
  gcr.io/my-project \
  --auto-detect-workers

# Manual worker count
freightliner replicate-tree \
  ecr/my-company \
  gcr.io/my-project \
  --workers 20
```

## 10. Deployment to Kubernetes

```bash
# Deploy to development environment
kubectl apply -k deployments/kubernetes/overlays/dev

# Deploy to production environment
kubectl apply -k deployments/kubernetes/overlays/prod

# Verify deployment
kubectl get pods -n freightliner
kubectl get svc -n freightliner

# Check logs
kubectl logs -n freightliner deployment/freightliner -f

# Port-forward for local access
kubectl port-forward -n freightliner svc/freightliner 8080:8080

# Test deployed service
curl https://freightliner.example.com/health
```

## Common Use Cases

### Use Case 1: Docker Hub to ECR Migration

```bash
# Migrate Docker Hub organization to ECR
freightliner replicate-tree \
  docker.io/myorganization \
  123456789012.dkr.ecr.us-west-2.amazonaws.com/myorganization \
  --enable-checkpoint \
  --workers 15 \
  --exclude-tag "latest"
```

### Use Case 2: Multi-Cloud Registry Sync

```yaml
# multi-cloud-sync.yaml
images:
  - source: docker.io/myorg/app
    destination: ecr/myorg/app
    tags:
      patterns: ["v*"]

  - source: docker.io/myorg/app
    destination: gcr.io/my-project/myorg/app
    tags:
      patterns: ["v*"]

  - source: docker.io/myorg/app
    destination: quay.io/myorg/app
    tags:
      patterns: ["v*"]

parallel_workers: 12
```

```bash
freightliner sync --config multi-cloud-sync.yaml
```

### Use Case 3: Continuous Security Scanning

```bash
#!/bin/bash
# scan-latest-images.sh

REGISTRIES=(
  "gcr.io/my-project/app-1:latest"
  "gcr.io/my-project/app-2:latest"
  "gcr.io/my-project/app-3:latest"
)

for image in "${REGISTRIES[@]}"; do
  echo "Scanning $image..."
  freightliner scan "$image" \
    --fail-on high \
    --format sarif \
    --output "scan-$(echo $image | tr '/:' '-').sarif"
done
```

## Troubleshooting

### Authentication Errors

```bash
# AWS ECR
aws ecr get-login-password --region us-west-2 | \
  docker login --username AWS --password-stdin \
  123456789012.dkr.ecr.us-west-2.amazonaws.com

# GCP GCR
gcloud auth configure-docker
gcloud auth print-access-token | \
  docker login -u oauth2accesstoken --password-stdin gcr.io
```

### Connection Issues

```bash
# Test registry connectivity
freightliner auth test docker.io
freightliner auth test gcr.io/my-project

# Enable debug logging
freightliner replicate \
  docker.io/library/alpine:latest \
  gcr.io/my-project/alpine:latest \
  --log-level debug
```

### Performance Tuning

```bash
# Increase workers for faster replication
freightliner replicate-tree \
  ecr/my-company \
  gcr.io/my-project \
  --workers 25 \
  --auto-detect-workers=false

# Enable checkpoint for large operations
freightliner replicate-tree \
  ecr/my-company \
  gcr.io/my-project \
  --enable-checkpoint \
  --checkpoint-dir /var/freightliner/checkpoints
```

### Interrupted Operations

```bash
# List checkpoints
freightliner checkpoint list

# Resume from checkpoint
freightliner replicate-tree \
  ecr/my-company \
  gcr.io/my-project \
  --resume-id <checkpoint-id> \
  --skip-completed \
  --retry-failed
```

## Next Steps

### Learn More
- **[Complete CLI Reference](docs/CLI_COMMANDS.md)** - All commands and flags
- **[Configuration Guide](docs/server-configuration.md)** - Detailed config options
- **[HTTP API Documentation](docs/API.md)** - REST API reference
- **[Architecture Guide](docs/ARCHITECTURE.md)** - System design and components

### Operations
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Kubernetes and Docker deployment
- **[Operations Runbook](docs/RUNBOOK.md)** - Troubleshooting and maintenance
- **[Monitoring Guide](docs/MONITORING-QUICKSTART.md)** - Prometheus and metrics

### Security
- **[Security Guide](docs/SECURITY.md)** - Security best practices
- **[Secrets Management](docs/SECRETS_MANAGEMENT.md)** - Managing credentials
- **[SBOM and Scanning](docs/SBOM_AND_SCANNING.md)** - Vulnerability scanning

### Development
- **[Development Guide](docs/DEVELOPMENT.md)** - Local development setup
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute
- **[GitHub Repository](https://github.com/hemzaz/freightliner)** - Source code

## Support

- **Issues**: [GitHub Issues](https://github.com/hemzaz/freightliner/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hemzaz/freightliner/discussions)

---

**Ready to replicate containers at scale?** 🚂✨
