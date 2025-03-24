# Freightliner

[![Go Report Card](https://goreportcard.com/badge/github.com/hemzaz/freightliner)](https://goreportcard.com/report/github.com/hemzaz/freightliner)
[![GoDoc](https://godoc.org/github.com/hemzaz/freightliner?status.svg)](https://godoc.org/github.com/hemzaz/freightliner)
[![License](https://img.shields.io/github/license/hemzaz/freightliner.svg)](https://github.com/hemzaz/freightliner/blob/master/LICENSE)

Freightliner is a container registry replication tool that supports cross-registry replication between AWS ECR and Google Container Registry (GCR).

## Features

- Cross-registry replication between ECR and GCR
- Bidirectional replication capabilities
- Single image replication
- Complete tree replication across registries
- Support for multi-architecture images and manifest lists
- Repository and tag filtering with pattern matching
- Parallel replication with configurable worker counts
- Metrics collection for monitoring
- Dry-run capability for validation

## Installation

### Binary Installation

Download the latest release from the [releases page](https://github.com/hemzaz/freightliner/releases).

### Docker

```bash
docker pull ghcr.io/hemzaz/freightliner:latest
```

### Build from Source

```bash
git clone https://github.com/hemzaz/freightliner.git
cd freightliner
go build -o bin/freightliner src/cmd/freightliner/main.go
```

## Quick Start

### Single Repository Replication

```bash
# Replicate a repository from ECR to GCR
freightliner replicate ecr/my-repository gcr/my-repository

# Specify AWS region and GCP project
freightliner replicate ecr/my-repository gcr/my-repository --ecr-region=us-east-1 --gcr-project=my-project
```

### Tree Replication

```bash
# Replicate all repositories with prefix "prod/" from ECR to GCR
freightliner replicate-tree ecr/prod gcr/prod-mirror

# With filtering
freightliner replicate-tree ecr/staging gcr/staging-mirror \
  --exclude-repo="internal-*" \
  --include-tag="v*" \
  --workers=10
```

### Server Mode

```bash
# Start the replication server with a configuration file
freightliner serve --config config.yaml
```

## Configuration

See the [usage documentation](docs/usage.md) for detailed configuration options.

## Authentication

### AWS ECR

Freightliner uses the standard AWS SDK authentication methods. You can configure authentication using:

- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- AWS configuration files (`~/.aws/credentials`)
- IAM roles (when running on EC2 or ECS)

### Google GCR

Freightliner uses the standard Google Cloud authentication methods. You can configure authentication using:

- Service account JSON key file (specified with `GOOGLE_APPLICATION_CREDENTIALS`)
- Application Default Credentials
- GKE Workload Identity (when running on GKE)

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/freightliner src/cmd/freightliner/main.go
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.