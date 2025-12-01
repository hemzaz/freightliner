# Freightliner Quick Start

**Get running in 5 minutes** ⚡

## What is Freightliner?

Container registry replication tool: **AWS ECR ↔️ Google GCR**

## Install & Run

```bash
# Clone and build
git clone <repo>
cd freightliner
go build -o freightliner cmd/server/main.go

# Run
./freightliner --source-registry ecr://123456789.dkr.ecr.us-east-1.amazonaws.com \
               --dest-registry gcr://my-project/my-repo
```

## Deploy to Kubernetes

```bash
# Dev environment
kubectl apply -k deployments/kubernetes/overlays/dev

# Production
kubectl apply -k deployments/kubernetes/overlays/prod
```

## Configuration

```yaml
# config.yaml
source:
  type: ecr
  region: us-east-1

destination:
  type: gcr
  project: my-project

workers: 10
encryption: true
```

## Common Commands

```bash
# Run tests
go test ./...

# Build Docker image
docker build -t freightliner:latest .

# Deploy
./scripts/deployment/deploy.sh production v1.0.0

# Rollback
./scripts/deployment/rollback.sh production
```

## Health Checks

```bash
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

## Troubleshooting

**Authentication failed?**
```bash
aws ecr get-login-password | docker login --username AWS --password-stdin <ecr-url>
gcloud auth configure-docker
```

**Build fails?**
```bash
go mod download
go mod tidy
```

## Next Steps

- [Full Documentation](docs/ARCHITECTURE.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Development Guide](docs/DEVELOPMENT.md)
