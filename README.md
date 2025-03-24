# Freightliner

Freightliner is a high-performance container registry replication tool that supports cross-registry replication between AWS ECR and Google GCR.

## Features

- Cross-registry replication between ECR and GCR
- Bidirectional replication support
- Rule-based filtering using patterns
- Scheduled replication with cron syntax
- Parallel image processing for high throughput
- API rate limiting to comply with registry limits
- Comprehensive monitoring with Prometheus metrics
- CLI and server modes

## Installation

### Binary Installation

```bash
# Download the latest release (replace with the appropriate OS/arch)
curl -LO https://github.com/elad/freightliner/releases/download/v0.1.0/freightliner_0.1.0_linux_x86_64.tar.gz
tar xzf freightliner_0.1.0_linux_x86_64.tar.gz
chmod +x freightliner
mv freightliner /usr/local/bin/
```

### Docker

```bash
docker pull ghcr.io/elad/freightliner:latest
```

## Quick Start

### One-time Replication

```bash
# Replicate a repository from ECR to GCR
freightliner replicate ecr/my-repository gcr/my-repository
```

### Server Mode

```bash
# Start the replication server with a configuration file
freightliner serve --config config.yaml
```

## Configuration

Create a `config.yaml` file:

```yaml
registries:
  ecr:
    type: ecr
    region: us-west-2
  gcr:
    type: gcr
    project: my-project

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

## Documentation

See the [usage documentation](docs/usage.md) for more detailed information.

## Development

### Building from Source

```bash
# Build the binary
go build -o freightliner ./cmd/freightliner

# Run tests
go test -v ./...

# Create a release (using goreleaser)
goreleaser release --snapshot --clean
```

## License

MIT
