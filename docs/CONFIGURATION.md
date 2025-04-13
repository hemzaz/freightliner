# Freightliner Configuration Guide

Freightliner offers a flexible configuration system that supports multiple sources:
1. Default values
2. YAML configuration files 
3. Environment variables (highest priority)
4. Command-line flags (overrides all other sources)

## Configuration File

You can specify a configuration file using the global flag:
```
freightliner --config /path/to/config.yaml [command]
```

See `examples/config_example.yaml` for a complete example configuration file.

## Environment Variables

All configuration settings can be overridden with environment variables. Environment variables have higher priority than configuration file values but lower priority than command-line flags.

### Naming Convention

Environment variables for Freightliner follow this pattern:
```
FREIGHTLINER_[SECTION]_[OPTION]
```

For example, to set the log level:
```
FREIGHTLINER_LOG_LEVEL=debug
```

### Variable Types

- **Strings**: Set directly as environment variable values
- **Booleans**: Use `true`, `false`, `1`, `0`, `yes`, or `no`
- **Integers**: Use numeric values
- **Lists/Slices**: Use comma-separated values
- **Durations**: Use time formats like `30s`, `5m`, `1h`

### Environment Variable Reference

#### General Settings
- `FREIGHTLINER_LOG_LEVEL`: Log level (debug, info, warn, error, fatal)

#### ECR Configuration
- `FREIGHTLINER_ECR_REGION`: AWS region for ECR
- `FREIGHTLINER_ECR_ACCOUNT_ID`: AWS account ID for ECR

#### GCR Configuration
- `FREIGHTLINER_GCR_PROJECT`: GCP project for GCR
- `FREIGHTLINER_GCR_LOCATION`: GCR location (us, eu, asia)

#### Worker Configuration
- `FREIGHTLINER_REPLICATE_WORKERS`: Number of workers for replication
- `FREIGHTLINER_SERVE_WORKERS`: Number of workers for server mode
- `FREIGHTLINER_AUTO_DETECT_WORKERS`: Whether to auto-detect worker count

#### Encryption Configuration
- `FREIGHTLINER_ENCRYPTION_ENABLED`: Enable encryption
- `FREIGHTLINER_CUSTOMER_MANAGED_KEYS`: Use customer-managed keys
- `FREIGHTLINER_AWS_KMS_KEY_ID`: AWS KMS key ID
- `FREIGHTLINER_GCP_KMS_KEY_ID`: GCP KMS key ID
- `FREIGHTLINER_GCP_KEY_RING`: GCP KMS key ring
- `FREIGHTLINER_GCP_KEY_NAME`: GCP KMS key name
- `FREIGHTLINER_ENVELOPE_ENCRYPTION`: Enable envelope encryption

#### Secrets Configuration
- `FREIGHTLINER_USE_SECRETS_MANAGER`: Whether to use secrets manager
- `FREIGHTLINER_SECRETS_MANAGER_TYPE`: Type of secrets manager (aws, gcp)
- `FREIGHTLINER_AWS_SECRET_REGION`: AWS region for Secrets Manager
- `FREIGHTLINER_GCP_SECRET_PROJECT`: GCP project for Secret Manager
- `FREIGHTLINER_GCP_CREDENTIALS_FILE`: GCP credentials file path
- `FREIGHTLINER_REGISTRY_CREDS_SECRET`: Secret name for registry credentials
- `FREIGHTLINER_ENCRYPTION_KEYS_SECRET`: Secret name for encryption keys

#### Server Configuration
- `FREIGHTLINER_SERVER_PORT`: Server port
- `FREIGHTLINER_TLS_ENABLED`: Enable TLS
- `FREIGHTLINER_TLS_CERT_FILE`: TLS certificate file
- `FREIGHTLINER_TLS_KEY_FILE`: TLS key file
- `FREIGHTLINER_API_KEY_AUTH`: Enable API key authentication
- `FREIGHTLINER_API_KEY`: API key for authentication
- `FREIGHTLINER_SERVER_ALLOWED_ORIGINS`: Allowed CORS origins (comma-separated)
- `FREIGHTLINER_SERVER_READ_TIMEOUT`: Server read timeout (e.g., 30s)
- `FREIGHTLINER_SERVER_WRITE_TIMEOUT`: Server write timeout (e.g., 60s)
- `FREIGHTLINER_SERVER_SHUTDOWN_TIMEOUT`: Server shutdown timeout (e.g., 15s)
- `FREIGHTLINER_HEALTH_CHECK_PATH`: Health check path
- `FREIGHTLINER_METRICS_PATH`: Metrics path
- `FREIGHTLINER_REPLICATE_PATH`: Replicate API path
- `FREIGHTLINER_TREE_REPLICATE_PATH`: Tree replicate API path
- `FREIGHTLINER_STATUS_PATH`: Status API path

#### Checkpoint Configuration
- `FREIGHTLINER_CHECKPOINT_DIRECTORY`: Directory for checkpoints
- `FREIGHTLINER_CHECKPOINT_ID`: Checkpoint ID

#### Tree Replication Configuration
- `FREIGHTLINER_TREE_WORKERS`: Number of workers for tree replication
- `FREIGHTLINER_TREE_EXCLUDE_REPOS`: Repository patterns to exclude (comma-separated)
- `FREIGHTLINER_TREE_EXCLUDE_TAGS`: Tag patterns to exclude (comma-separated)
- `FREIGHTLINER_TREE_INCLUDE_TAGS`: Tag patterns to include (comma-separated)
- `FREIGHTLINER_TREE_DRY_RUN`: Enable dry run mode
- `FREIGHTLINER_TREE_FORCE`: Force overwrite of existing images
- `FREIGHTLINER_TREE_ENABLE_CHECKPOINT`: Enable checkpointing
- `FREIGHTLINER_TREE_CHECKPOINT_DIR`: Checkpoint directory for tree replication
- `FREIGHTLINER_TREE_RESUME_ID`: Resume ID
- `FREIGHTLINER_TREE_SKIP_COMPLETED`: Skip completed repositories
- `FREIGHTLINER_TREE_RETRY_FAILED`: Retry failed repositories

#### Replication Configuration
- `FREIGHTLINER_REPLICATE_FORCE`: Force overwrite of existing images
- `FREIGHTLINER_REPLICATE_DRY_RUN`: Enable dry run mode
- `FREIGHTLINER_REPLICATE_TAGS`: Tags to replicate (comma-separated)

## Configuration Precedence

When multiple configuration sources provide a value for the same setting, the precedence is as follows (from highest to lowest):

1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

## Examples

### Using a configuration file with environment variable overrides

```bash
# Use a config file for most settings
export FREIGHTLINER_ECR_REGION=us-east-1  # Override the region
export FREIGHTLINER_REPLICATE_WORKERS=8    # Use 8 workers 

# Run with the config file and the overrides
freightliner --config freightliner.yaml replicate-tree ecr/src-repo gcr/dest-repo
```

### Using environment variables for CI/CD pipelines

```bash
# Set all necessary configuration via environment variables
export FREIGHTLINER_LOG_LEVEL=info
export FREIGHTLINER_ECR_REGION=us-west-2
export FREIGHTLINER_ECR_ACCOUNT_ID=123456789012
export FREIGHTLINER_GCR_PROJECT=my-project
export FREIGHTLINER_REPLICATE_WORKERS=4
export FREIGHTLINER_TREE_DRY_RUN=false
export FREIGHTLINER_TREE_FORCE=true
export FREIGHTLINER_TREE_INCLUDE_TAGS=latest,v1,v2

# Run without a config file
freightliner replicate-tree ecr/src-repo gcr/dest-repo
```

### Using Docker with environment variables

```bash
docker run -e FREIGHTLINER_LOG_LEVEL=debug \
           -e FREIGHTLINER_ECR_REGION=us-west-2 \
           -e FREIGHTLINER_ECR_ACCOUNT_ID=123456789012 \
           -e FREIGHTLINER_GCR_PROJECT=my-project \
           -e FREIGHTLINER_TREE_INCLUDE_TAGS=latest,v1.0,v2.0 \
           freightliner:latest replicate-tree ecr/src-repo gcr/dest-repo
```
