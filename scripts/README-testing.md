# Freightliner Testing Setup

This directory contains scripts and tools for setting up a complete testing environment for Freightliner.

## Prerequisites

- Docker installed and running
- `curl` command available
- `jq` (optional, for pretty JSON output)

## Quick Start

1. **Set up test registries:**
   ```bash
   ./scripts/setup-test-registries.sh
   ```

2. **Run tests:**
   ```bash
   go test ./pkg/tree/ -v
   go test ./pkg/service/ -v
   go test ./pkg/copy/ -v
   ```

3. **Clean up when done:**
   ```bash
   ./scripts/setup-test-registries.sh --cleanup
   ```

## What the Setup Script Does

### 1. Creates Local Docker Registries
- **Source Registry**: `localhost:5100` (mimics source container registry)
- **Destination Registry**: `localhost:5101` (mimics destination container registry)
- **Docker Network**: `freightliner-test` (for container communication)

### 2. Populates Test Data
The script creates a realistic test environment with multiple repositories and tags:

```
Source Registry (localhost:5100):
├── project-a/
│   ├── service-1:v1.0, v1.1, latest (Alpine images)
│   └── service-2:v2.0, latest (Nginx images)
├── project-b/
│   └── service-3:v3.0, latest (Busybox images)
└── project-c/
    └── service-4:v1.0 (Hello-world image)
```

### 3. Test Image Mapping
- **Alpine 3.18** → `project-a/service-1:v1.0`
- **Alpine 3.19** → `project-a/service-1:v1.1`
- **Alpine latest** → `project-a/service-1:latest`
- **Nginx 1.24** → `project-a/service-2:v2.0`
- **Nginx latest** → `project-a/service-2:latest`
- **Busybox 1.36** → `project-b/service-3:v3.0`
- **Busybox latest** → `project-b/service-3:latest`
- **Hello-world** → `project-c/service-4:v1.0`

## Manual Testing Commands

### Registry Inspection
```bash
# List all repositories in source registry
curl http://localhost:5100/v2/_catalog

# List tags for a specific repository
curl http://localhost:5100/v2/project-a/service-1/tags/list

# Get manifest for a specific tag
curl http://localhost:5100/v2/project-a/service-1/manifests/v1.0

# Check destination registry (should be empty initially)
curl http://localhost:5101/v2/_catalog
```

### Docker Registry Management
```bash
# View running containers
docker ps

# Check registry logs
docker logs source-registry
docker logs dest-registry

# Inspect network
docker network inspect freightliner-test
```

### Manual Image Operations
```bash
# Pull an image from source registry
docker pull localhost:5100/project-a/service-1:v1.0

# Tag and push to destination registry
docker tag localhost:5100/project-a/service-1:v1.0 localhost:5101/project-a/service-1:v1.0
docker push localhost:5101/project-a/service-1:v1.0

# Verify push worked
curl http://localhost:5101/v2/_catalog
```

## Freightliner Test Configuration

When running Freightliner tests, the registries will be accessible at:
- **Source**: `localhost:5100`
- **Destination**: `localhost:5101`

Example Freightliner configuration for testing:
```yaml
source:
  type: "generic"
  endpoint: "localhost:5100"
  
destination:
  type: "generic"
  endpoint: "localhost:5101"
```

## Troubleshooting

### Registry Not Starting
```bash
# Check if ports are already in use
netstat -an | grep :5100
netstat -an | grep :5101

# Check Docker daemon
docker info
```

### Images Not Appearing
```bash
# Verify images were pushed
docker images | grep localhost:5100

# Check registry storage
docker exec source-registry find /var/lib/registry -name "*.json"
```

### Network Issues
```bash
# Test registry connectivity
curl -v http://localhost:5100/v2/
curl -v http://localhost:5101/v2/

# Check containers can communicate
docker exec source-registry ping dest-registry
```

### Permission Issues
```bash
# Check registry permissions
docker exec source-registry ls -la /var/lib/registry

# Restart with debug logging
docker logs source-registry --follow
```

## Test Data Structure for Different Scenarios

### Basic Replication Test
- Source: `project-a/service-1` with tags `v1.0`, `v1.1`, `latest`
- Expected: All tags replicated to destination

### Selective Replication Test
- Source: Multiple repositories with various tags
- Expected: Only specified repositories/tags replicated

### Error Handling Test
- Simulate network issues by stopping registries
- Expected: Graceful error handling and retry logic

### Resume Functionality Test
- Start replication, interrupt, then resume
- Expected: Replication continues from checkpoint

## Cleanup

The cleanup command will:
1. Stop and remove registry containers
2. Remove the Docker network
3. Optionally remove pulled test images
4. Clean up any tagged images

```bash
./scripts/setup-test-registries.sh --cleanup
```

## Advanced Usage

### Custom Test Data
To add your own test repositories, modify the `repositories` array in the script:

```bash
declare -A repositories=(
    ["your-project/your-service"]="image:tag1,image:tag2"
)
```

### Different Registry Ports
Modify the script configuration:
```bash
SOURCE_REGISTRY_PORT=5000
DEST_REGISTRY_PORT=5001
```

### Registry with Authentication
To test with authentication, modify the Docker run command to include auth configuration.

This setup provides a complete, realistic testing environment for developing and testing Freightliner's container registry replication functionality.