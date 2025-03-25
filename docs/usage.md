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

### Command-line Options

Freightliner supports the following global command-line options:

| Option | Description |
|--------|-------------|
| `--log-level` | Log level (debug, info, warn, error, fatal) |
| `--ecr-region` | AWS region for ECR (default: us-west-2) |
| `--ecr-account` | AWS account ID for ECR (empty uses default from credentials) |
| `--gcr-project` | GCP project for GCR |
| `--gcr-location` | GCR location (us, eu, asia) (default: us) |

Example:
```bash
freightliner replicate ecr/my-repo gcr/my-repo --ecr-region=us-east-1 --gcr-project=my-gcp-project
```

#### Security Options

| Option | Description |
|--------|-------------|
| `--sign` | Enable image signing |
| `--verify` | Verify image signatures |
| `--sign-key` | Path to the signing key file |
| `--sign-key-id` | ID of the signing key |
| `--signature-store` | Path to store image signatures |
| `--strict-verify` | Fail if signature verification isn't possible |
| `--encrypt` | Enable image encryption |
| `--customer-key` | Use customer-managed encryption keys |
| `--aws-kms-key` | AWS KMS key ID for encryption |
| `--gcp-kms-key` | GCP KMS key ID for encryption |
| `--envelope-encryption` | Use envelope encryption (default: true) |

#### Secrets Manager Options

| Option | Description |
|--------|-------------|
| `--use-secrets-manager` | Enable using cloud provider secrets manager |
| `--secrets-manager-type` | Type of secrets manager to use (aws, gcp) |
| `--aws-secret-region` | AWS region for Secrets Manager |
| `--gcp-secret-project` | GCP project for Secret Manager |
| `--gcp-credentials-file` | GCP credentials file path |
| `--registry-creds-secret` | Secret name for registry credentials |
| `--encryption-keys-secret` | Secret name for encryption keys |
| `--signing-keys-secret` | Secret name for signing keys |

### YAML Configuration

Freightliner can be configured using a YAML configuration file for the server mode:

```yaml
registries:
  ecr:
    type: ecr
    region: us-west-2
    account_id: "123456789012"  # Optional, uses AWS credentials if empty
  gcr:
    type: gcr
    project: my-project
    location: us  # Optional, defaults to "us"

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
  metrics_port: 9090
```

## Authentication

### AWS ECR

Freightliner uses the standard AWS SDK authentication methods. You can configure authentication using:

- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- AWS configuration files (`~/.aws/credentials`)
- IAM roles (when running on EC2 or ECS)
- AWS Secrets Manager (see Secrets Manager Integration section)

### Google GCR

Freightliner uses the standard Google Cloud authentication methods. You can configure authentication using:

- Service account JSON key file (specified with `GOOGLE_APPLICATION_CREDENTIALS`)
- Application Default Credentials
- GKE Workload Identity (when running on GKE)
- Google Secret Manager (see Secrets Manager Integration section)

### Secrets Manager Integration

Freightliner supports storing and retrieving credentials from cloud provider secrets managers:

```bash
# Using AWS Secrets Manager
freightliner replicate ecr/my-repository gcr/my-repository \
  --use-secrets-manager \
  --secrets-manager-type=aws \
  --aws-secret-region=us-west-2 \
  --registry-creds-secret=freightliner-registry-credentials

# Using Google Secret Manager
freightliner replicate ecr/my-repository gcr/my-repository \
  --use-secrets-manager \
  --secrets-manager-type=gcp \
  --gcp-secret-project=my-project \
  --registry-creds-secret=freightliner-registry-credentials
```

Available options for secrets manager:

| Option | Description |
|--------|-------------|
| `--use-secrets-manager` | Enable using cloud provider secrets manager |
| `--secrets-manager-type` | Type of secrets manager to use (aws, gcp) |
| `--aws-secret-region` | AWS region for Secrets Manager |
| `--gcp-secret-project` | GCP project for Secret Manager |
| `--gcp-credentials-file` | GCP credentials file path |
| `--registry-creds-secret` | Secret name for registry credentials |
| `--encryption-keys-secret` | Secret name for encryption keys |
| `--signing-keys-secret` | Secret name for signing keys |

#### Registry Credentials Format

Registry credentials should be stored in the following JSON format:

```json
{
  "ecr": {
    "access_key": "AWS_ACCESS_KEY_ID",
    "secret_key": "AWS_SECRET_ACCESS_KEY",
    "account_id": "012345678901",
    "region": "us-west-2",
    "session_token": "OPTIONAL_SESSION_TOKEN"
  },
  "gcr": {
    "project": "my-project",
    "location": "us",
    "credentials": "BASE64_ENCODED_SERVICE_ACCOUNT_JSON"
  }
}
```

#### Encryption Keys Format

Encryption keys should be stored in the following JSON format:

```json
{
  "aws": {
    "kms_key_id": "alias/my-key-alias",
    "region": "us-west-2"
  },
  "gcp": {
    "kms_key_id": "projects/my-project/locations/global/keyRings/freightliner/cryptoKeys/image-encryption",
    "project": "my-project",
    "location": "global",
    "key_ring": "freightliner",
    "key": "image-encryption"
  }
}
```

#### Signing Keys Format

Signing keys should be stored in the following JSON format:

```json
{
  "key_path": "/path/to/key/file",
  "key_id": "key-identifier",
  "key_data": "BASE64_ENCODED_KEY_DATA"
}
```

If `key_data` is provided, it will be decoded and written to a temporary file for use during the operation.

## Advanced Features

### Tree Replication

Tree replication allows you to replicate entire repository trees from one registry to another, matching a specific prefix pattern:

```bash
# Replicate all repositories with prefix "prod/" from ECR to GCR
freightliner replicate-tree ecr/prod gcr/prod-mirror

# Replicate all repositories with prefix "staging/" excluding internal ones
freightliner replicate-tree ecr/staging gcr/staging-mirror --exclude-repo="internal-*"

# Only replicate versioned tags
freightliner replicate-tree ecr/prod gcr/prod-mirror --include-tag="v*"

# Perform a dry run without actually copying images
freightliner replicate-tree ecr/prod gcr/prod-mirror --dry-run

# Use 10 concurrent worker threads
freightliner replicate-tree ecr/prod gcr/prod-mirror --workers=10

# Enable checkpointing for interrupted replications
freightliner replicate-tree ecr/prod gcr/prod-mirror --checkpoint

# Resume an interrupted replication
freightliner replicate-tree ecr/prod gcr/prod-mirror --resume=<checkpoint-id>
```

Available options for tree replication:

| Option | Description |
|--------|-------------|
| `--workers` | Number of concurrent worker threads (default: 5) |
| `--exclude-repo` | Repository patterns to exclude (e.g. 'internal-*') |
| `--exclude-tag` | Tag patterns to exclude (e.g. 'dev-*') |
| `--include-tag` | Tag patterns to include (e.g. 'v*') |
| `--dry-run` | Perform a dry run without actually copying images |
| `--force` | Force overwrite of existing images |
| `--checkpoint` | Enable checkpointing for interrupted replications |
| `--checkpoint-dir` | Directory to store checkpoint files |
| `--resume` | Resume replication from a checkpoint ID |
| `--skip-completed` | Skip completed repositories when resuming (default: true) |
| `--retry-failed` | Retry failed repositories when resuming (default: true) |

### Network Optimization

Freightliner optimizes network transfers when replicating images:

1. **Compression**: Data is compressed during transfer to reduce bandwidth usage
2. **Delta Updates**: Only transfers the differences between image versions

The network optimization is enabled by default and helps:

- Reduce bandwidth usage by up to 90% for similar images
- Speed up replication of large images
- Improve performance over slow or high-latency connections

### Checkpoint Management

Freightliner provides commands to manage checkpoints:

```bash
# List available checkpoints
freightliner checkpoint list

# Show details of a specific checkpoint
freightliner checkpoint show --id=<checkpoint-id>

# Delete a checkpoint
freightliner checkpoint delete --id=<checkpoint-id>
```

Checkpoints allow you to:

- Resume interrupted replications
- Track progress of long-running operations
- Recover from failures without starting over

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
