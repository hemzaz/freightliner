#!/bin/bash

# Test Infrastructure Teardown Script
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🧹 Tearing down test infrastructure..."

# Stop and remove Docker Compose services
docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" down -v

echo "✅ Test infrastructure cleaned up!"
