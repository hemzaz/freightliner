# Freightliner Usage Guide

Freightliner is a tool for replicating container images between different container registries. It supports AWS ECR and Google GCR.

## Installation

### Binary Installation

Download the latest release from the [releases page](https://github.com/elad/freightliner/releases).

### Docker

```bash
docker pull ghcr.io/elad/freightliner:latest
```

### Homebrew

```bash
brew tap elad/tools
brew install freightliner
```

## Basic Usage

### One-time Replication

To replicate a repository from one registry to another:

```bash
freightliner replicate ecr/my-repository gcr/my-repository
```

### Server Mode

Start the replication server that will periodically replicate based on configuration:

```bash
freightliner serve --config config.yaml
```

## Configuration

Freightliner can be configured using a YAML configuration file:

```yaml
registries:
  ecr:
    type: ecr
    region: us-west-2
    # Additional AWS-specific settings
  gcr:
    type: gcr
    project: my-project
    # Additional GCP-specific settings

rules:
  - source_registry: ecr
    source_repository: my-repository
    destination_registry: gcr
    destination_repository: my-repository
    tag_filter: "v*"
    schedule: "*/30 * * * *"  # Every 30 minutes

settings:
  max_concurrent_replications: 5
  retry_count: 3
```

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

## Advanced Features

### Bidirectional Replication

To set up bidirectional replication, configure rules in both directions:

```yaml
rules:
  - source_registry: ecr
    source_repository: my-repository
    destination_registry: gcr
    destination_repository: my-repository
    tag_filter: "*"
    schedule: "*/30 * * * *"
  
  - source_registry: gcr
    source_repository: my-repository
    destination_registry: ecr
    destination_repository: my-repository
    tag_filter: "*"
    schedule: "*/30 * * * *"
```

### Repository Pattern Matching

You can use wildcards in repository patterns to match multiple repositories:

```yaml
rules:
  - source_registry: ecr
    source_repository: "app-*"
    destination_registry: gcr
    destination_repository: "app-*"
    tag_filter: "prod-*"
    schedule: "0 * * * *"
```

### Monitoring

Freightliner exposes Prometheus metrics on port 9090 by default. Available metrics include:

- `freightliner_replication_count`: Number of replications performed
- `freightliner_replication_errors`: Number of replication errors
- `freightliner_replication_duration_seconds`: Duration of replication operations
