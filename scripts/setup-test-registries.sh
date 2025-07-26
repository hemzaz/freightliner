#!/bin/bash

# setup-test-registries.sh
# Sets up local Docker registries for Freightliner testing
# Requires Docker to be installed and running

set -euo pipefail

# Configuration
SOURCE_REGISTRY_PORT=5100
DEST_REGISTRY_PORT=5101
NETWORK_NAME="freightliner-test"
SOURCE_REGISTRY_NAME="source-registry"
DEST_REGISTRY_NAME="dest-registry"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
check_docker() {
    log_info "Checking if Docker is running..."
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    log_success "Docker is running"
}

# Create Docker network
create_network() {
    log_info "Creating Docker network: $NETWORK_NAME"
    if docker network ls | grep -q "$NETWORK_NAME"; then
        log_warning "Network $NETWORK_NAME already exists, skipping creation"
    else
        docker network create "$NETWORK_NAME"
        log_success "Created network: $NETWORK_NAME"
    fi
}

# Start registry container
start_registry() {
    local name=$1
    local port=$2
    
    log_info "Starting registry: $name on port $port"
    
    # Stop existing container if it exists
    if docker ps -a --format "table {{.Names}}" | grep -q "^${name}$"; then
        log_info "Stopping existing $name container"
        docker stop "$name" >/dev/null 2>&1 || true
        docker rm "$name" >/dev/null 2>&1 || true
    fi
    
    # Start new registry container
    docker run -d \
        --name "$name" \
        --network "$NETWORK_NAME" \
        -p "${port}:5000" \
        -e REGISTRY_STORAGE_DELETE_ENABLED=true \
        -e REGISTRY_HTTP_ADDR=0.0.0.0:5000 \
        -e REGISTRY_LOG_LEVEL=debug \
        registry:2
    
    log_success "Started registry: $name on port $port"
}

# Wait for registry to be ready
wait_for_registry() {
    local port=$1
    local name=$2
    
    log_info "Waiting for $name to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f "http://localhost:${port}/v2/" >/dev/null 2>&1; then
            log_success "$name is ready"
            return 0
        fi
        
        log_info "Attempt $attempt/$max_attempts: $name not ready yet, waiting..."
        sleep 2
        ((attempt++))
    done
    
    log_error "$name failed to start after $max_attempts attempts"
    return 1
}

# Pull test images from Docker Hub
pull_test_images() {
    log_info "Pulling test images from Docker Hub..."
    
    local images=(
        "alpine:3.18"
        "alpine:3.19"
        "alpine:latest"
        "nginx:1.24"
        "nginx:1.25"
        "nginx:latest"
        "busybox:1.36"
        "busybox:latest"
        "hello-world:latest"
    )
    
    for image in "${images[@]}"; do
        log_info "Pulling $image"
        docker pull "$image"
    done
    
    log_success "All test images pulled"
}

# Tag and push images to source registry
populate_source_registry() {
    log_info "Populating source registry with test images..."
    
    # Populate project-a/service-1 with Alpine images
    log_info "Populating repository: project-a/service-1"
    docker tag alpine:3.18 localhost:${SOURCE_REGISTRY_PORT}/project-a/service-1:v1.0
    docker tag alpine:3.19 localhost:${SOURCE_REGISTRY_PORT}/project-a/service-1:v1.1
    docker tag alpine:latest localhost:${SOURCE_REGISTRY_PORT}/project-a/service-1:latest
    
    docker push localhost:${SOURCE_REGISTRY_PORT}/project-a/service-1:v1.0
    docker push localhost:${SOURCE_REGISTRY_PORT}/project-a/service-1:v1.1
    docker push localhost:${SOURCE_REGISTRY_PORT}/project-a/service-1:latest
    log_success "Populated repository: project-a/service-1"
    
    # Populate project-a/service-2 with Nginx images
    log_info "Populating repository: project-a/service-2"
    docker tag nginx:1.24 localhost:${SOURCE_REGISTRY_PORT}/project-a/service-2:v2.0
    docker tag nginx:latest localhost:${SOURCE_REGISTRY_PORT}/project-a/service-2:latest
    
    docker push localhost:${SOURCE_REGISTRY_PORT}/project-a/service-2:v2.0
    docker push localhost:${SOURCE_REGISTRY_PORT}/project-a/service-2:latest
    log_success "Populated repository: project-a/service-2"
    
    # Populate project-b/service-3 with Busybox images
    log_info "Populating repository: project-b/service-3"
    docker tag busybox:1.36 localhost:${SOURCE_REGISTRY_PORT}/project-b/service-3:v3.0
    docker tag busybox:latest localhost:${SOURCE_REGISTRY_PORT}/project-b/service-3:latest
    
    docker push localhost:${SOURCE_REGISTRY_PORT}/project-b/service-3:v3.0
    docker push localhost:${SOURCE_REGISTRY_PORT}/project-b/service-3:latest
    log_success "Populated repository: project-b/service-3"
    
    # Populate project-c/service-4 with Hello-world image
    log_info "Populating repository: project-c/service-4"
    docker tag hello-world:latest localhost:${SOURCE_REGISTRY_PORT}/project-c/service-4:v1.0
    
    docker push localhost:${SOURCE_REGISTRY_PORT}/project-c/service-4:v1.0
    log_success "Populated repository: project-c/service-4"
    
    log_success "Source registry populated with all test repositories"
}

# Verify registry contents
verify_registries() {
    log_info "Verifying registry contents..."
    
    log_info "Source registry (localhost:${SOURCE_REGISTRY_PORT}) contents:"
    echo "Repositories:"
    curl -s "http://localhost:${SOURCE_REGISTRY_PORT}/v2/_catalog" | jq -r '.repositories[]' 2>/dev/null || {
        log_warning "jq not available, showing raw output:"
        curl -s "http://localhost:${SOURCE_REGISTRY_PORT}/v2/_catalog"
    }
    
    echo ""
    log_info "Destination registry (localhost:${DEST_REGISTRY_PORT}) contents:"
    echo "Repositories:"
    curl -s "http://localhost:${DEST_REGISTRY_PORT}/v2/_catalog" | jq -r '.repositories[]?' 2>/dev/null || {
        log_warning "jq not available or no repositories, showing raw output:"
        curl -s "http://localhost:${DEST_REGISTRY_PORT}/v2/_catalog"
    }
}

# Show registry URLs and test commands
show_usage() {
    log_info "Registry setup complete!"
    echo ""
    echo "Registry URLs:"
    echo "  Source Registry:      http://localhost:${SOURCE_REGISTRY_PORT}"
    echo "  Destination Registry: http://localhost:${DEST_REGISTRY_PORT}"
    echo ""
    echo "Container Names:"
    echo "  Source:      $SOURCE_REGISTRY_NAME"
    echo "  Destination: $DEST_REGISTRY_NAME"
    echo "  Network:     $NETWORK_NAME"
    echo ""
    echo "Test Commands:"
    echo "  List source repositories:"
    echo "    curl http://localhost:${SOURCE_REGISTRY_PORT}/v2/_catalog"
    echo ""
    echo "  List tags for a repository:"
    echo "    curl http://localhost:${SOURCE_REGISTRY_PORT}/v2/project-a/service-1/tags/list"
    echo ""
    echo "  Run Freightliner tests:"
    echo "    go test ./pkg/tree/ -v"
    echo ""
    echo "  Clean up (when done):"
    echo "    $0 --cleanup"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test registries..."
    
    # Stop and remove containers
    for container in "$SOURCE_REGISTRY_NAME" "$DEST_REGISTRY_NAME"; do
        if docker ps -a --format "table {{.Names}}" | grep -q "^${container}$"; then
            log_info "Stopping and removing $container"
            docker stop "$container" >/dev/null 2>&1 || true
            docker rm "$container" >/dev/null 2>&1 || true
        fi
    done
    
    # Remove network
    if docker network ls | grep -q "$NETWORK_NAME"; then
        log_info "Removing network $NETWORK_NAME"
        docker network rm "$NETWORK_NAME" >/dev/null 2>&1 || true
    fi
    
    # Remove test images (optional)
    read -p "Remove pulled test images? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Removing test images..."
        docker rmi alpine:3.18 alpine:3.19 alpine:latest >/dev/null 2>&1 || true
        docker rmi nginx:1.24 nginx:1.25 nginx:latest >/dev/null 2>&1 || true
        docker rmi busybox:1.36 busybox:latest >/dev/null 2>&1 || true
        docker rmi hello-world:latest >/dev/null 2>&1 || true
        
        # Remove tagged images
        docker images --format "table {{.Repository}}:{{.Tag}}" | grep "localhost:${SOURCE_REGISTRY_PORT}" | while read -r image; do
            docker rmi "$image" >/dev/null 2>&1 || true
        done
        
        log_success "Test images removed"
    fi
    
    log_success "Cleanup complete"
}

# Main function
main() {
    log_info "Setting up Freightliner test registries..."
    
    check_docker
    create_network
    
    # Start registries
    start_registry "$SOURCE_REGISTRY_NAME" "$SOURCE_REGISTRY_PORT"
    start_registry "$DEST_REGISTRY_NAME" "$DEST_REGISTRY_PORT"
    
    # Wait for registries to be ready
    wait_for_registry "$SOURCE_REGISTRY_PORT" "$SOURCE_REGISTRY_NAME"
    wait_for_registry "$DEST_REGISTRY_PORT" "$DEST_REGISTRY_NAME"
    
    # Pull and populate images
    pull_test_images
    populate_source_registry
    
    # Verify setup
    verify_registries
    show_usage
}

# Parse command line arguments
case "${1:-}" in
    --cleanup|-c)
        cleanup
        exit 0
        ;;
    --help|-h)
        echo "Usage: $0 [--cleanup|--help]"
        echo ""
        echo "Options:"
        echo "  --cleanup, -c    Clean up test registries and networks"
        echo "  --help, -h       Show this help message"
        echo ""
        echo "Default: Set up test registries for Freightliner testing"
        exit 0
        ;;
    "")
        main
        ;;
    *)
        log_error "Unknown option: $1"
        echo "Use $0 --help for usage information"
        exit 1
        ;;
esac