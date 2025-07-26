# Freightliner Local Registry Testing Setup - Complete

## 🎯 Overview

This directory provides a complete testing infrastructure for Freightliner container registry replication, including local Docker registries populated with realistic test data.

## ✅ What's Included

### 1. **Main Setup Script** 
**File**: `setup-test-registries.sh`
- Creates two local Docker registries (source: 5100, destination: 5101)
- Pulls real container images from Docker Hub
- Populates source registry with test repositories
- Creates Docker network for container communication

### 2. **Validation Script**
**File**: `test-registry-setup.sh` 
- Validates registry connectivity and health
- Tests repository listing and tag enumeration
- Performs complete image pull/push cycle
- Generates comprehensive test report

### 3. **Documentation**
**File**: `README-testing.md`
- Complete usage guide and troubleshooting
- Manual testing commands
- Configuration examples

### 4. **Makefile Integration**
New make targets for easy testing:
- `make test-setup` - Set up registries
- `make test-validate` - Validate setup
- `make test-full` - Run complete test suite
- `make test-cleanup` - Remove registries

## 🚀 Quick Start

```bash
# 1. Set up test registries
make test-setup

# 2. Validate setup (optional)
make test-validate

# 3. Run tests
make test-full

# 4. Clean up when done
make test-cleanup
```

## 📊 Test Data Structure

The setup creates a realistic test environment with:

```
Source Registry (localhost:5100):
├── project-a/
│   ├── service-1:v1.0, v1.1, latest (Alpine 3.18, 3.19, latest)
│   └── service-2:v2.0, latest (Nginx 1.24, latest)
├── project-b/
│   └── service-3:v3.0, latest (Busybox 1.36, latest)  
└── project-c/
    └── service-4:v1.0 (Hello-world)

Destination Registry (localhost:5101):
└── (empty, ready for replication testing)
```

## 🧪 What Tests Are Enabled

### 1. **Tree Replication Tests**
- Multi-repository replication
- Tag filtering and selection
- Parallel processing validation
- Error handling and recovery
- Progress tracking and checkpoints

### 2. **Service Integration Tests**
- Registry client functionality  
- Secrets management operations
- Configuration validation
- Authentication workflows

### 3. **Copy Package Tests**
- Image copying mechanics
- Manifest and layer transfer
- Error handling for network issues
- Performance and statistics

## 🔧 Technical Details

### Registry Configuration
- **Source Port**: 5100 (avoids macOS Control Center on 5000)
- **Destination Port**: 5101
- **Network**: `freightliner-test` Docker bridge network
- **Storage**: In-memory (deleted on container removal)

### Container Images Used
- **Alpine**: 3.18, 3.19, latest (small, fast for testing)
- **Nginx**: 1.24, latest (multi-layer images)
- **Busybox**: 1.36, latest (minimal images)
- **Hello-world**: latest (tiny test image)

### Test Scenarios Covered
1. **Basic replication**: Single repository, multiple tags
2. **Selective replication**: Filter by repository or tag patterns
3. **Error handling**: Network timeouts, permission issues
4. **Resume functionality**: Checkpoint and restart scenarios
5. **Performance**: Large repositories, concurrent operations

## 📋 Usage Examples

### Manual Registry Inspection
```bash
# List all repositories
curl http://localhost:5100/v2/_catalog

# List tags for a repository
curl http://localhost:5100/v2/project-a/service-1/tags/list

# Get image manifest
curl -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
     http://localhost:5100/v2/project-a/service-1/manifests/v1.0
```

### Test Individual Components
```bash
# Test tree replication only
go test ./pkg/tree/ -v -timeout=300s

# Test with specific verbosity
go test ./pkg/tree/ -v -run TestReplicateTree

# Test service layer
go test ./pkg/service/ -v -timeout=300s
```

### Registry Management
```bash
# View container logs
docker logs source-registry
docker logs dest-registry

# Check container status
docker ps --filter "name=registry"

# Monitor network traffic
docker exec source-registry ping dest-registry
```

## ⚡ Performance Notes

### Startup Time
- **Registry containers**: ~5 seconds to be ready
- **Image pulling**: ~30 seconds (cached after first run)
- **Registry population**: ~15 seconds
- **Total setup time**: ~60 seconds

### Test Execution
- **Basic tests**: 2-3 minutes
- **Full test suite**: 5-10 minutes  
- **Cleanup**: ~10 seconds

### Resource Usage
- **Memory**: ~100MB per registry container
- **Storage**: ~200MB for test images (temporary)
- **Network**: Local bridge, no external bandwidth

## 🐛 Troubleshooting

### Common Issues

1. **Port conflicts**:
   ```bash
   lsof -i :5100 :5101
   # Change ports in setup script if needed
   ```

2. **Docker not running**:
   ```bash
   docker info
   # Start Docker Desktop or daemon
   ```

3. **Registry not responding**:
   ```bash
   docker logs source-registry
   curl -v http://localhost:5100/v2/
   ```

4. **Network issues**:
   ```bash
   docker network ls
   docker network inspect freightliner-test
   ```

### Advanced Debugging

```bash
# Enter registry container
docker exec -it source-registry sh

# Check registry storage
docker exec source-registry find /var/lib/registry -type f

# Monitor registry requests
docker logs -f source-registry

# Test with curl verbose
curl -v http://localhost:5100/v2/_catalog
```

## 🎯 Testing Best Practices

### Before Running Tests
1. Ensure Docker has sufficient resources (2GB+ RAM)
2. Close other applications using ports 5100-5101
3. Run setup validation before tests
4. Check registry logs for any startup issues

### During Tests  
1. Monitor container health: `docker ps`
2. Watch for network timeouts in test output
3. Check destination registry gets populated
4. Verify cleanup between test runs

### After Tests
1. Review test logs for any errors
2. Check registry logs for issues
3. Validate no data leaked between tests
4. Clean up to free system resources

## 📈 Extending the Test Environment

### Adding New Test Repositories
Edit `setup-test-registries.sh`:
```bash
# Add new repository in populate_source_registry()
log_info "Populating repository: my-project/my-service"
docker tag some-image:tag localhost:${SOURCE_REGISTRY_PORT}/my-project/my-service:v1.0
docker push localhost:${SOURCE_REGISTRY_PORT}/my-project/my-service:v1.0
```

### Testing with Authentication
1. Add auth config to registry containers
2. Create test credentials
3. Update client configurations
4. Test auth workflows

### Performance Testing
1. Add more repositories and tags
2. Use larger container images
3. Test concurrent replication scenarios  
4. Monitor resource usage during tests

## ✅ Success Validation

A successful setup should show:
- ✅ Two registry containers running
- ✅ 4 test repositories in source registry
- ✅ Empty destination registry ready for testing
- ✅ Network connectivity between containers
- ✅ All registry endpoints accessible via curl
- ✅ Test images tagged and pushed successfully

## 🎉 Ready for Development

With this setup, you can now:
- Test Freightliner's core replication functionality
- Debug registry client implementations  
- Validate error handling and recovery
- Develop new features with realistic test data
- Ensure production readiness with comprehensive testing

The local registry environment provides a complete, isolated testing infrastructure that mirrors production registry scenarios while being fast and reliable for development workflows.