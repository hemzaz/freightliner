# Freightliner

[![Go Report Card](https://goreportcard.com/badge/github.com/hemzaz/freightliner)](https://goreportcard.com/report/github.com/hemzaz/freightliner)
[![GoDoc](https://godoc.org/github.com/hemzaz/freightliner?status.svg)](https://godoc.org/github.com/hemzaz/freightliner)
[![License](https://img.shields.io/github/license/hemzaz/freightliner.svg)](https://blob/master/LICENSE)

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
- Image signing and verification using Cosign
- Customer-managed encryption keys (AWS KMS and GCP KMS)
- Network bandwidth optimization with compression and delta updates
- Checkpointing for resumable replication

## Installation

### Binary Installation

Download the latest release from the [releases page](https://releases).

### Docker

```bash
docker pull ghcr.io/hemzaz/freightliner:latest
```

### Build from Source

```bash
git clone https://github.com/hemzaz/freightliner.git
cd freightliner
go build -o bin/freightliner cmd/freightliner/main.go
```

## Quick Start

### Single Repository Replication

```bash
# Replicate a repository from ECR to GCR
freightliner replicate ecr/my-repository gcr/my-repository

# Specify AWS region and GCP project
freightliner replicate ecr/my-repository gcr/my-repository --ecr-region=us-east-1 --gcr-project=my-project
```

### Security Features

```bash
# Replicate with image signing and verification
freightliner replicate ecr/my-repository gcr/my-repository \
  --sign --sign-key=/path/to/cosign.key --sign-key-id=my-key-id

# Replicate with AWS KMS customer-managed encryption key
freightliner replicate ecr/my-repository gcr/my-repository \
  --encrypt --customer-key --aws-kms-key=alias/my-key

# Replicate with GCP KMS customer-managed encryption key
freightliner replicate ecr/my-repository gcr/my-repository \
  --encrypt --customer-key --gcp-kms-key=projects/my-project/locations/global/keyRings/freightliner/cryptoKeys/my-key
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

# Resumable replication with checkpointing
freightliner replicate-tree ecr/prod gcr/prod-mirror \
  --checkpoint --checkpoint-dir=/path/to/checkpoints

# Resume an interrupted replication
freightliner replicate-tree ecr/prod gcr/prod-mirror \
  --resume=<checkpoint-id> --skip-completed
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

## Recent Changes

### Security Features

- **Image Signing with Cosign**: 
  - Added support for signing and verifying images using Cosign
  - Integrates with both ECR and GCR registry types

- **Customer-Managed Encryption Keys**:
  - Implemented AWS KMS integration for ECR images
  - Implemented GCP KMS integration for GCR images
  - Added envelope encryption for secure data transfer

- **Cloud Secrets Manager Integration**:
  - Added support for AWS Secrets Manager and Google Secret Manager
  - Securely store and retrieve registry credentials, encryption keys, and signing materials
  - Command-line flag support for using cloud provider secrets in operations
  - JSON-based structured secret format for complex configuration

### Client Fixes

- **ECR Client**:
  - Fixed credential helper implementation to properly support the Authenticator interface
  - Resolved MediaType undefined errors by properly importing the types package
  - Added correct authorization method to the ECR credential helper
  - Fixed client configuration handling with authentication libraries

- **GCR Client**:
  - Updated Google registry list functionality to work with the latest API
  - Fixed transport and authentication handling for Google credentials
  - Corrected manifest handling for different image types
  - Improved descriptor handling for registry operations

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/freightliner cmd/freightliner/main.go
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.