#!/bin/bash

# Test Infrastructure Setup Script
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "🚀 Setting up test infrastructure..."

# Start Docker Compose services
echo "📦 Starting test services..."
docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
timeout=60
elapsed=0

while [ $elapsed -lt $timeout ]; do
    if docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" ps | grep -q "(healthy)"; then
        echo "✅ All services are healthy"
        break
    fi

    echo "Waiting for services... ($elapsed/$timeout seconds)"
    sleep 5
    elapsed=$((elapsed + 5))
done

if [ $elapsed -ge $timeout ]; then
    echo "❌ Timeout waiting for services to become healthy"
    docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" ps
    exit 1
fi

# Populate source registry with test images
echo "📥 Populating source registry with test data..."

# Pull some small public images for testing
test_images=(
    "alpine:3.18"
    "alpine:3.19"
    "busybox:1.36"
    "busybox:latest"
)

for image in "${test_images[@]}"; do
    echo "Pulling ${image}..."
    docker pull "${image}" || echo "Warning: Failed to pull ${image}"

    # Tag for source registry
    local_tag="localhost:5000/test/${image}"
    docker tag "${image}" "${local_tag}" || echo "Warning: Failed to tag ${image}"

    # Push to source registry
    docker push "${local_tag}" || echo "Warning: Failed to push ${image}"
done

echo "✅ Test infrastructure is ready!"
echo ""
echo "Test Endpoints:"
echo "  Source Registry: http://localhost:5000"
echo "  Dest Registry:   http://localhost:5001"
echo "  Redis:           localhost:6380"
echo "  MinIO:           http://localhost:9002"
echo "  Prometheus:      http://localhost:9091"
echo ""
echo "Run tests with: go test -v ./tests/integration/..."
