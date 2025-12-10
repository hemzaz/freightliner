#!/bin/bash
# Test execution with manifest support
# This script integrates the test manifest system with existing test workflows

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Default values
MANIFEST_PATH="test-manifest.yaml"
ENVIRONMENT=""
CATEGORIES=""
DRY_RUN=false
VERBOSE=false
PACKAGES=""

# Usage function
usage() {
    cat << EOF
Test Manifest Integration Script

Usage: $0 [options] [packages...]

Options:
    -m, --manifest PATH     Path to test manifest file (default: test-manifest.yaml)
    -e, --env ENV          Environment override (ci|local|integration)
    -c, --categories CATS  Comma-separated categories to filter by
    -d, --dry-run          Show what would be executed without running
    -v, --verbose          Enable verbose output
    -s, --summary          Show test manifest summary and exit
    -h, --help             Show this help message

Examples:
    $0                                    # Run all enabled tests for current environment
    $0 --env ci                          # Run tests as if in CI environment
    $0 --categories unit                 # Run only unit tests
    $0 freightliner/pkg/client/gcr       # Run tests for specific package
    $0 --summary                         # Show test manifest summary
    $0 --dry-run ./...                   # Show what tests would run for all packages

Environment Detection:
    The script automatically detects the environment based on environment variables:
    - CI environment: CI=true, GITHUB_ACTIONS=true, etc.
    - Integration: TEST_ENV=integration, RUN_INTEGRATION_TESTS=true
    - Local: Default when no CI indicators are present

Test Categories:
    - unit: Pure unit tests with no external dependencies
    - integration: Tests requiring real external services
    - external_deps: Tests requiring AWS, GCP, or other external dependencies
    - flaky: Tests that are intermittently failing
    - incomplete: Tests for incomplete functionality
    - timing_sensitive: Tests sensitive to timing and concurrency
    - metrics: Tests related to metrics collection
    - worker_pool: Tests related to worker pool functionality

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--manifest)
            MANIFEST_PATH="$2"
            shift 2
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -c|--categories)
            CATEGORIES="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -s|--summary)
            SHOW_SUMMARY=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        -*)
            echo -e "${RED}Unknown option: $1${NC}" >&2
            usage >&2
            exit 1
            ;;
        *)
            PACKAGES="$PACKAGES $1"
            shift
            ;;
    esac
done

# Build test-manifest tool if it doesn't exist
TEST_MANIFEST_BIN="./bin/test-manifest"
if [[ ! -f "$TEST_MANIFEST_BIN" ]]; then
    echo -e "${YELLOW}Building test-manifest tool...${NC}"
    mkdir -p bin
    go build -o "$TEST_MANIFEST_BIN" ./cmd/test-manifest
    echo -e "${GREEN}Test manifest tool built successfully${NC}"
fi

# Check if manifest file exists
if [[ ! -f "$MANIFEST_PATH" ]]; then
    echo -e "${RED}Error: Test manifest file not found: $MANIFEST_PATH${NC}" >&2
    echo "Create a test manifest file or specify a different path with -m" >&2
    exit 1
fi

# Build command arguments
CMD_ARGS=()
if [[ -n "$ENVIRONMENT" ]]; then
    CMD_ARGS+=("-env" "$ENVIRONMENT")
fi
if [[ -n "$CATEGORIES" ]]; then
    CMD_ARGS+=("-categories" "$CATEGORIES")
fi
if [[ "$DRY_RUN" == "true" ]]; then
    CMD_ARGS+=("-dry-run")
fi
if [[ "$VERBOSE" == "true" ]]; then
    CMD_ARGS+=("-verbose")
fi

# Add manifest path
CMD_ARGS+=("-manifest" "$MANIFEST_PATH")

# Show summary if requested
if [[ "$SHOW_SUMMARY" == "true" ]]; then
    echo -e "${GREEN}Test Manifest Summary${NC}"
    "$TEST_MANIFEST_BIN" summary "${CMD_ARGS[@]}"
    exit 0
fi

# Determine packages to test
if [[ -z "$PACKAGES" ]]; then
    # No specific packages specified, run all packages in manifest
    echo -e "${YELLOW}Running tests for all packages in manifest...${NC}"
    
    # First show summary
    echo -e "${GREEN}Current Test Configuration:${NC}"
    "$TEST_MANIFEST_BIN" summary "${CMD_ARGS[@]}"
    echo ""
    
    # Get all packages from manifest and run them
    PACKAGES=$(go list ./... | grep -E "(pkg/client|pkg/replication|pkg/tree|pkg/network|pkg/metrics|pkg/helper|pkg/copy)")
fi

# Track results
TOTAL_PACKAGES=0
PASSED_PACKAGES=0
FAILED_PACKAGES=0
FAILED_PACKAGE_LIST=""

# Run tests for each package
for PACKAGE in $PACKAGES; do
    TOTAL_PACKAGES=$((TOTAL_PACKAGES + 1))
    
    echo -e "${YELLOW}Testing package: $PACKAGE${NC}"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        "$TEST_MANIFEST_BIN" test "${CMD_ARGS[@]}" "$PACKAGE"
    else
        if "$TEST_MANIFEST_BIN" test "${CMD_ARGS[@]}" "$PACKAGE"; then
            echo -e "${GREEN}✓ $PACKAGE passed${NC}"
            PASSED_PACKAGES=$((PASSED_PACKAGES + 1))
        else
            echo -e "${RED}✗ $PACKAGE failed${NC}"
            FAILED_PACKAGES=$((FAILED_PACKAGES + 1))
            FAILED_PACKAGE_LIST="$FAILED_PACKAGE_LIST $PACKAGE"
        fi
        echo ""
    fi
done

# Print summary
if [[ "$DRY_RUN" != "true" ]]; then
    echo -e "${GREEN}Test Execution Summary:${NC}"
    echo "  Total packages: $TOTAL_PACKAGES"
    echo "  Passed: $PASSED_PACKAGES"
    echo "  Failed: $FAILED_PACKAGES"
    
    if [[ $FAILED_PACKAGES -gt 0 ]]; then
        echo -e "${RED}Failed packages:${NC}$FAILED_PACKAGE_LIST"
        exit 1
    else
        echo -e "${GREEN}All tests passed!${NC}"
    fi
fi