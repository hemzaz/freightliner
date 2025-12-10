#!/bin/bash

# test-registry-setup.sh
# Validates that the test registry setup is working correctly

set -euo pipefail

# Configuration
SOURCE_REGISTRY_PORT=5100
DEST_REGISTRY_PORT=5101

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

# Test registry connectivity
test_registry_connectivity() {
    log_info "Testing registry connectivity..."
    
    # Test source registry
    if curl -f -s "http://localhost:${SOURCE_REGISTRY_PORT}/v2/" >/dev/null; then
        log_success "Source registry (localhost:${SOURCE_REGISTRY_PORT}) is accessible"
    else
        log_error "Source registry (localhost:${SOURCE_REGISTRY_PORT}) is not accessible"
        return 1
    fi
    
    # Test destination registry
    if curl -f -s "http://localhost:${DEST_REGISTRY_PORT}/v2/" >/dev/null; then
        log_success "Destination registry (localhost:${DEST_REGISTRY_PORT}) is accessible"
    else
        log_error "Destination registry (localhost:${DEST_REGISTRY_PORT}) is not accessible"
        return 1
    fi
}

# Test repository listing
test_repository_listing() {
    log_info "Testing repository listing..."
    
    local repos
    repos=$(curl -s "http://localhost:${SOURCE_REGISTRY_PORT}/v2/_catalog" | grep -o '"repositories":\[[^]]*\]' || true)
    
    if [[ -n "$repos" ]]; then
        log_success "Source registry contains repositories"
        log_info "Repositories found: $repos"
    else
        log_warning "Source registry appears to be empty"
        return 1
    fi
}

# Test specific repositories
test_specific_repositories() {
    log_info "Testing specific test repositories..."
    
    local expected_repos=(
        "project-a/service-1"
        "project-a/service-2"
        "project-b/service-3"
        "project-c/service-4"
    )
    
    for repo in "${expected_repos[@]}"; do
        log_info "Checking repository: $repo"
        
        # Test tags endpoint
        local url="http://localhost:${SOURCE_REGISTRY_PORT}/v2/${repo}/tags/list"
        if curl -f -s "$url" >/dev/null; then
            local tags
            tags=$(curl -s "$url" | grep -o '"tags":\[[^]]*\]' || echo "no tags found")
            log_success "Repository $repo exists with tags: $tags"
        else
            log_warning "Repository $repo not found or inaccessible"
        fi
    done
}

# Test image manifest retrieval
test_manifest_retrieval() {
    log_info "Testing image manifest retrieval..."
    
    local test_image="project-a/service-1"
    local test_tag="v1.0"
    local url="http://localhost:${SOURCE_REGISTRY_PORT}/v2/${test_image}/manifests/${test_tag}"
    
    if curl -f -s -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "$url" >/dev/null; then
        log_success "Successfully retrieved manifest for ${test_image}:${test_tag}"
    else
        log_warning "Could not retrieve manifest for ${test_image}:${test_tag}"
    fi
}

# Test Docker container status
test_container_status() {
    log_info "Testing Docker container status..."
    
    local containers=("source-registry" "dest-registry")
    
    for container in "${containers[@]}"; do
        if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "$container"; then
            local status
            status=$(docker ps --format "table {{.Names}}\t{{.Status}}" | grep "$container" | awk '{print $2}')
            log_success "Container $container is running ($status)"
        else
            log_error "Container $container is not running"
            return 1
        fi
    done
}

# Test network connectivity
test_network_connectivity() {
    log_info "Testing Docker network connectivity..."
    
    if docker network ls --format "table {{.Name}}" | grep -q "freightliner-test"; then
        log_success "Docker network 'freightliner-test' exists"
        
        # Test if containers can reach each other
        if docker exec source-registry ping -c 1 dest-registry >/dev/null 2>&1; then
            log_success "Network connectivity between registries works"
        else
            log_warning "Network connectivity test failed"
        fi
    else
        log_error "Docker network 'freightliner-test' not found"
        return 1
    fi
}

# Test a complete image pull/push cycle
test_image_operations() {
    log_info "Testing complete image operations..."
    
    local test_repo="project-a/service-1"
    local test_tag="v1.0"
    local source_image="localhost:${SOURCE_REGISTRY_PORT}/${test_repo}:${test_tag}"
    local dest_image="localhost:${DEST_REGISTRY_PORT}/${test_repo}:${test_tag}"
    
    # Pull from source
    log_info "Pulling test image from source registry..."
    if docker pull "$source_image" >/dev/null 2>&1; then
        log_success "Successfully pulled $source_image"
        
        # Tag for destination
        docker tag "$source_image" "$dest_image"
        
        # Push to destination
        log_info "Pushing test image to destination registry..."
        if docker push "$dest_image" >/dev/null 2>&1; then
            log_success "Successfully pushed $dest_image"
            
            # Verify it appears in destination catalog
            sleep 2
            if curl -s "http://localhost:${DEST_REGISTRY_PORT}/v2/_catalog" | grep -q "$test_repo"; then
                log_success "Image appears in destination registry catalog"
            else
                log_warning "Image not found in destination registry catalog"
            fi
        else
            log_warning "Failed to push $dest_image"
        fi
        
        # Clean up test images
        docker rmi "$source_image" "$dest_image" >/dev/null 2>&1 || true
    else
        log_warning "Failed to pull $source_image"
    fi
}

# Generate test report
generate_report() {
    echo ""
    log_info "=== TEST REGISTRY SETUP VALIDATION REPORT ==="
    echo ""
    echo "Registry URLs:"
    echo "  Source:      http://localhost:${SOURCE_REGISTRY_PORT}"
    echo "  Destination: http://localhost:${DEST_REGISTRY_PORT}"
    echo ""
    echo "Quick Commands:"
    echo "  List repositories: curl http://localhost:${SOURCE_REGISTRY_PORT}/v2/_catalog"
    echo "  List tags:         curl http://localhost:${SOURCE_REGISTRY_PORT}/v2/project-a/service-1/tags/list"
    echo "  Run tests:         go test ./pkg/tree/ -v"
    echo ""
    
    if [[ $overall_success == true ]]; then
        log_success "All tests passed! Setup is ready for Freightliner testing."
    else
        log_warning "Some tests failed. Check the output above for details."
        echo ""
        echo "Troubleshooting:"
        echo "  1. Make sure Docker is running"
        echo "  2. Run: ./scripts/setup-test-registries.sh"
        echo "  3. Check for port conflicts (5000, 5001)"
        echo "  4. Review container logs: docker logs source-registry"
    fi
}

# Main execution
main() {
    log_info "Validating Freightliner test registry setup..."
    echo ""
    
    local overall_success=true
    
    # Run all tests
    test_container_status || overall_success=false
    echo ""
    
    test_network_connectivity || overall_success=false
    echo ""
    
    test_registry_connectivity || overall_success=false
    echo ""
    
    test_repository_listing || overall_success=false
    echo ""
    
    test_specific_repositories || overall_success=false
    echo ""
    
    test_manifest_retrieval || overall_success=false
    echo ""
    
    test_image_operations || overall_success=false
    
    # Generate report
    generate_report
}

# Run main function
main