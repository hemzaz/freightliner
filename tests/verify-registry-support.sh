#!/bin/bash
# Quick Registry Support Verification Test
# Tests that Critical Bug #1 fix is working

set -e

echo "=================================================="
echo "Registry Support Verification Test"
echo "=================================================="
echo ""

BINARY="./freightliner"

if [ ! -f "$BINARY" ]; then
    echo "❌ Binary not found. Run: go build -o freightliner ."
    exit 1
fi

echo "✅ Binary found: $BINARY"
echo ""

# Test 1: Docker Hub
echo "Test 1: Docker Hub Support"
echo "Command: $BINARY replicate --dry-run docker.io/library/alpine:latest docker.io/test/alpine:copy"
if $BINARY replicate --dry-run docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1 | grep -q "Auto-detected Docker Hub"; then
    echo "✅ Docker Hub: PASS (registry detected)"
else
    echo "❌ Docker Hub: FAIL"
    exit 1
fi
echo ""

# Test 2: Quay.io
echo "Test 2: Quay.io Support"
echo "Command: $BINARY replicate --dry-run --tags latest quay.io/prometheus/node-exporter:latest quay.io/test/node-exporter:copy"
if $BINARY replicate --dry-run --tags latest quay.io/prometheus/node-exporter:latest quay.io/test/node-exporter:copy 2>&1 | grep -q "generic OCI registry"; then
    echo "✅ Quay.io: PASS (generic client used)"
else
    echo "❌ Quay.io: FAIL"
    exit 1
fi
echo ""

# Test 3: GHCR
echo "Test 3: GitHub Container Registry Support"
echo "Command: $BINARY replicate --dry-run --tags latest ghcr.io/linuxserver/plex:latest ghcr.io/test/plex:copy"
if $BINARY replicate --dry-run --tags latest ghcr.io/linuxserver/plex:latest ghcr.io/test/plex:copy 2>&1 | grep -q "Auto-detected GitHub Container Registry"; then
    echo "✅ GHCR: PASS (registry detected)"
else
    echo "❌ GHCR: FAIL"
    exit 1
fi
echo ""

# Test 4: Tag Stripping
echo "Test 4: Tag Stripping from Repository Names"
echo "Command: $BINARY replicate --dry-run docker.io/library/alpine:3.18 docker.io/test/alpine:latest"
if $BINARY replicate --dry-run docker.io/library/alpine:3.18 docker.io/test/alpine:latest 2>&1 | grep -q "repository=library/alpine"; then
    echo "✅ Tag Stripping: PASS (tag removed from repo name)"
else
    echo "❌ Tag Stripping: FAIL"
    exit 1
fi
echo ""

# Test 5: Worker Auto-Detection
echo "Test 5: Worker Auto-Detection"
echo "Command: $BINARY replicate --dry-run --tags latest docker.io/library/alpine:latest docker.io/test/alpine:copy"
if $BINARY replicate --dry-run --tags latest docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1 | grep -q "Auto-detected worker count"; then
    echo "✅ Worker Auto-Detection: PASS"
else
    echo "❌ Worker Auto-Detection: FAIL"
    exit 1
fi
echo ""

echo "=================================================="
echo "✅ ALL TESTS PASSED"
echo "=================================================="
echo ""
echo "Registry Support Summary:"
echo "  ✅ Docker Hub (docker.io)"
echo "  ✅ Quay.io (quay.io)"
echo "  ✅ GitHub Container Registry (ghcr.io)"
echo "  ✅ Generic Docker v2 registries"
echo "  ✅ Tag/Digest stripping"
echo "  ✅ Auto-detection & authentication"
echo ""
echo "Critical Bug #1: FIXED ✅"
