# Freightliner

Container registry replication tool for AWS ECR, Google GCR, Docker Hub, Harbor, Quay, GitLab, GitHub, Azure.

```
    _______________________________________________
   |  ___________________________________________  |
   | |     FREIGHTLINER                          | |
   | |     Container Registry Replication        | |
   | |___________________________________________| |
   |_______________________________________________|
```

[![Go](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org) [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Install

```bash
git clone https://github.com/hemzaz/freightliner.git
cd freightliner
make build
```

## Quick Usage

```bash
# Single image
freightliner replicate docker.io/library/alpine:latest gcr.io/my-project/alpine:latest

# Entire repository tree
freightliner replicate-tree ecr/my-company gcr.io/my-project --workers 10

# YAML-based sync
freightliner sync --config sync.yaml

# Run API server
freightliner serve --port 8080 --config config.yaml
```

**→ [Quick Start Guide](QUICKSTART.md)** | **→ [CLI Reference](docs/CLI_COMMANDS.md)**

## Core Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `replicate` | Copy single image | `freightliner replicate SOURCE DEST` |
| `replicate-tree` | Copy repository tree | `freightliner replicate-tree SOURCE DEST --workers 10` |
| `sync` | YAML-based batch sync | `freightliner sync --config sync.yaml` |
| `inspect` | View image details | `freightliner inspect IMAGE` |
| `scan` | Vulnerability scan | `freightliner scan IMAGE --fail-on critical` |
| `sbom` | Generate SBOM | `freightliner sbom IMAGE --format spdx` |
| `serve` | Run HTTP API server | `freightliner serve --port 8080` |
| `list-tags` | List repository tags | `freightliner list-tags REPO` |
| `delete` | Delete image | `freightliner delete IMAGE --force` |
| `login/logout` | Registry auth | `freightliner login REGISTRY` |
| `checkpoint` | Manage checkpoints | `freightliner checkpoint list` |
| `version` | Show version | `freightliner version --banner` |

## Configuration

### File (YAML/JSON)

```yaml
# config.yaml
log_level: info

ecr:
  region: us-west-2
  account_id: "123456789012"

gcr:
  project: my-project
  location: us

workers:
  replicate_workers: 10
  auto_detect: true

encryption:
  enabled: true
  aws_kms_key_id: arn:aws:kms:...

server:
  port: 8080
  tls_enabled: true
  api_key_auth: true

metrics:
  enabled: true
  port: 2112
```

### HTTP/HTTPS URL

```bash
freightliner serve --config https://config.example.com/freightliner.yaml
```

### Environment Variables

```bash
export FREIGHTLINER_ECR_REGION=us-west-2
export FREIGHTLINER_LOG_LEVEL=debug
export FREIGHTLINER_REPLICATE_WORKERS=20
freightliner serve
```

## Key Flags

```bash
# Logging
--log-level debug|info|warn|error

# Workers
--replicate-workers 10
--auto-detect-workers

# Encryption
--encrypt
--aws-kms-key ARN
--gcp-kms-key KEY_ID

# Secrets
--use-secrets-manager
--secrets-manager-type aws|gcp

# Server
--port 8080
--tls --tls-cert CERT --tls-key KEY
--api-key-auth --api-key KEY

# Checkpoint
--enable-checkpoint
--resume-id ID
--skip-completed
--retry-failed

# Filtering
--exclude-tag "dev-*,test-*"
--tags "v1.0,v1.1,latest"
--dry-run
--force
```

## Common Operations

### Migrate Repository

```bash
freightliner replicate-tree \
  docker.io/myorg \
  gcr.io/my-project \
  --enable-checkpoint \
  --workers 15 \
  --exclude-tag "dev-*"
```

### Resume Interrupted Migration

```bash
freightliner checkpoint list
freightliner replicate-tree \
  SOURCE DEST \
  --resume-id <ID> \
  --skip-completed \
  --retry-failed
```

### Security Scan

```bash
freightliner scan IMAGE --fail-on critical --format sarif --output results.sarif
```

### Generate SBOM

```bash
freightliner sbom IMAGE --format spdx --output sbom.json
```

### Run API Server

```bash
freightliner serve \
  --port 8080 \
  --config config.yaml \
  --tls \
  --api-key-auth
```

## Health Checks

```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl http://localhost:2112/metrics
```

## Deploy

### Kubernetes

```bash
kubectl apply -k deployments/kubernetes/overlays/prod
kubectl get pods -n freightliner
```

### Docker Compose

```bash
docker-compose -f docker-compose.prod.yml up -d
```

## Troubleshooting

```bash
# Test registry auth
freightliner auth test REGISTRY

# Enable debug logging
freightliner COMMAND --log-level debug

# AWS ECR login
aws ecr get-login-password --region REGION | docker login --username AWS --password-stdin ECR_URL

# GCP GCR login
gcloud auth configure-docker
```

## Features

- ✅ Multi-registry sync (ECR, GCR, Docker Hub, Harbor, Quay, GitLab, GitHub, Azure)
- ✅ Multi-arch support (amd64, arm64)
- ✅ Checkpoint/resume for large migrations
- ✅ AES-256-GCM encryption with KMS
- ✅ Vulnerability scanning & SBOM generation
- ✅ HTTP API with Prometheus metrics
- ✅ YAML-based batch operations
- ✅ Worker auto-scaling
- ✅ Tag filtering (regex, semver)
- ✅ Dry-run mode

## Docs

- [Quick Start](QUICKSTART.md) - Get started in 5 minutes
- [CLI Reference](docs/CLI_COMMANDS.md) - All commands and flags
- [Configuration](docs/server-configuration.md) - Config file reference
- [HTTP API](docs/API.md) - REST API endpoints
- [Deployment](docs/DEPLOYMENT.md) - Kubernetes/Docker deployment
- [Operations](docs/RUNBOOK.md) - Troubleshooting guide
- [Security](docs/SECURITY.md) - Security practices
- [Development](docs/DEVELOPMENT.md) - Dev setup

## Status

**88% Production Ready** | Code 85% | Security 90% | Tests 85% | Deploy 95% | Monitoring 90%

## License

MIT
