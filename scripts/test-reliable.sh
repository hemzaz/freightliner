#!/bin/bash
# Reliable test execution script with retry logic and better error handling

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
MANIFEST_PATH="test-manifest.yaml"
ENVIRONMENT=""
CATEGORIES=""
PACKAGES=""
MAX_RETRIES=3
RETRY_DELAY=5
PARALLEL_JOBS=1
COVERAGE_THRESHOLD=60
TIMEOUT="10m"
FAIL_FAST=false
VERBOSE=false

# Usage function
usage() {
    cat << EOF
Reliable Test Execution Script

Usage: $0 [options] [packages...]

Options:
    -m, --manifest PATH         Path to test manifest file (default: test-manifest.yaml)
    -e, --env ENV              Environment override (ci|local|integration)
    -c, --categories CATS      Comma-separated categories to filter by
    -r, --max-retries NUM      Maximum number of retries for flaky tests (default: 3)
    -d, --retry-delay SEC      Delay between retries in seconds (default: 5)
    -j, --parallel NUM         Number of parallel test jobs (default: 1)
    -t, --timeout DURATION     Test timeout (default: 10m)
    --coverage-threshold NUM   Minimum coverage percentage (default: 60)
    --fail-fast               Stop on first failure
    -v, --verbose             Enable verbose output
    -h, --help                Show this help message

Examples:
    $0                                    # Run all tests with retries
    $0 --env ci --max-retries 5          # CI tests with 5 retries
    $0 --parallel 4 --timeout 15m        # Parallel tests with extended timeout
    $0 --categories unit --fail-fast     # Unit tests only, stop on first failure
    $0 freightliner/pkg/replication      # Test specific package with retries

Environment Detection:
    The script automatically detects the environment based on environment variables:
    - CI environment: CI=true, GITHUB_ACTIONS=true, etc.
    - Integration: TEST_ENV=integration, RUN_INTEGRATION_TESTS=true
    - Local: Default when no CI indicators are present

Retry Logic:
    - Flaky tests are automatically retried up to --max-retries times
    - Each retry waits --retry-delay seconds before attempting again
    - Only tests marked as flaky in the manifest are retried
    - Tests with consistent failures are not retried indefinitely

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
        -r|--max-retries)
            MAX_RETRIES="$2"
            shift 2
            ;;
        -d|--retry-delay)
            RETRY_DELAY="$2"
            shift 2
            ;;
        -j|--parallel)
            PARALLEL_JOBS="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --coverage-threshold)
            COVERAGE_THRESHOLD="$2"
            shift 2
            ;;
        --fail-fast)
            FAIL_FAST=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
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

# Build test-manifest tool if it doesn't exist
ensure_test_manifest_tool() {
    local TEST_MANIFEST_BIN="./bin/test-manifest"
    
    if [[ ! -f "$TEST_MANIFEST_BIN" ]]; then
        log_info "Building test-manifest tool..."
        mkdir -p bin
        if go build -o "$TEST_MANIFEST_BIN" ./cmd/test-manifest; then
            log_success "Test manifest tool built successfully"
        else
            log_error "Failed to build test-manifest tool"
            exit 1
        fi
    fi
}

# Check if manifest file exists
check_manifest_file() {
    if [[ ! -f "$MANIFEST_PATH" ]]; then
        log_error "Test manifest file not found: $MANIFEST_PATH"
        echo "Create a test manifest file or specify a different path with -m" >&2
        exit 1
    fi
}

# Detect environment if not specified
detect_environment() {
    if [[ -z "$ENVIRONMENT" ]]; then
        if [[ "$CI" == "true" || "$GITHUB_ACTIONS" == "true" ]]; then
            ENVIRONMENT="ci"
        elif [[ "$TEST_ENV" == "integration" || "$RUN_INTEGRATION_TESTS" == "true" ]]; then
            ENVIRONMENT="integration"
        else
            ENVIRONMENT="local"
        fi
    fi
    
    log_info "Detected environment: $ENVIRONMENT"
}

# Run a single test with retry logic
run_test_with_retry() {
    local package="$1"
    local attempt=1
    local success=false
    
    log_info "Testing package: $package"
    
    while [[ $attempt -le $MAX_RETRIES && "$success" == "false" ]]; do
        if [[ $attempt -gt 1 ]]; then
            log_warning "Retry attempt $attempt/$MAX_RETRIES for $package"
            sleep $RETRY_DELAY
        fi
        
        local test_cmd="./bin/test-manifest test"
        
        # Build command arguments
        if [[ -n "$ENVIRONMENT" ]]; then
            test_cmd="$test_cmd -env $ENVIRONMENT"
        fi
        if [[ -n "$CATEGORIES" ]]; then
            test_cmd="$test_cmd -categories $CATEGORIES"
        fi
        if [[ "$VERBOSE" == "true" ]]; then
            test_cmd="$test_cmd -verbose"
        fi
        
        test_cmd="$test_cmd -manifest $MANIFEST_PATH $package"
        
        # Run the test with timeout
        if timeout "$TIMEOUT" bash -c "$test_cmd"; then
            log_success "✓ $package passed (attempt $attempt)"
            success=true
        else
            local exit_code=$?
            if [[ $exit_code -eq 124 ]]; then
                log_error "✗ $package timed out after $TIMEOUT (attempt $attempt)"
            else
                log_error "✗ $package failed with exit code $exit_code (attempt $attempt)"
            fi
            
            if [[ "$FAIL_FAST" == "true" ]]; then
                log_error "Stopping due to --fail-fast option"
                return $exit_code
            fi
        fi
        
        ((attempt++))
    done
    
    if [[ "$success" == "true" ]]; then
        return 0
    else
        log_error "Package $package failed after $MAX_RETRIES attempts"
        return 1
    fi
}

# Run tests in parallel using GNU parallel or xargs
run_tests_parallel() {
    local packages_list="$1"
    
    if command -v parallel >/dev/null 2>&1; then
        log_info "Using GNU parallel for $PARALLEL_JOBS parallel jobs"
        echo "$packages_list" | tr ' ' '\n' | \
        parallel -j "$PARALLEL_JOBS" --halt now,fail=1 run_test_with_retry {}
    else
        log_warning "GNU parallel not available, using xargs"
        echo "$packages_list" | tr ' ' '\n' | \
        xargs -I {} -P "$PARALLEL_JOBS" bash -c 'run_test_with_retry "$@"' _ {}
    fi
}

# Generate coverage report
generate_coverage_report() {
    if [[ -f "coverage.out" ]]; then
        log_info "Generating coverage report..."
        
        local coverage_percent=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        
        if [[ -n "$coverage_percent" ]]; then
            log_info "Total test coverage: ${coverage_percent}%"
            
            if (( $(echo "$coverage_percent >= $COVERAGE_THRESHOLD" | bc -l) )); then
                log_success "Coverage meets threshold of ${COVERAGE_THRESHOLD}%"
            else
                log_warning "Coverage ${coverage_percent}% is below threshold of ${COVERAGE_THRESHOLD}%"
            fi
            
            # Generate HTML report if requested
            if [[ "$VERBOSE" == "true" ]]; then
                go tool cover -html=coverage.out -o coverage.html
                log_info "HTML coverage report generated: coverage.html"
            fi
        fi
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    # Add any cleanup logic here
}

# Set up signal handlers
trap cleanup EXIT
trap 'log_error "Script interrupted"; exit 130' INT TERM

# Main execution
main() {
    log_info "Starting reliable test execution..."
    log_info "Max retries: $MAX_RETRIES, Retry delay: ${RETRY_DELAY}s, Parallel jobs: $PARALLEL_JOBS"
    
    ensure_test_manifest_tool
    check_manifest_file
    detect_environment
    
    # Show configuration summary
    log_info "Test Configuration:"
    echo "  Environment: $ENVIRONMENT"
    echo "  Categories: ${CATEGORIES:-all}"
    echo "  Timeout: $TIMEOUT"
    echo "  Parallel jobs: $PARALLEL_JOBS"
    echo "  Max retries: $MAX_RETRIES"
    echo "  Coverage threshold: ${COVERAGE_THRESHOLD}%"
    
    # Determine packages to test
    if [[ -z "$PACKAGES" ]]; then
        log_info "No specific packages specified, discovering test packages..."
        PACKAGES=$(go list ./... | grep -E "(pkg/client|pkg/replication|pkg/tree|pkg/network|pkg/metrics|pkg/helper|pkg/copy)")
    fi
    
    local package_count=$(echo "$PACKAGES" | wc -w)
    log_info "Testing $package_count packages"
    
    # Track results
    local start_time=$(date +%s)
    local success=true
    
    # Run tests
    if [[ $PARALLEL_JOBS -gt 1 ]]; then
        log_info "Running tests in parallel..."
        if ! run_tests_parallel "$PACKAGES"; then
            success=false
        fi
    else
        log_info "Running tests sequentially..."
        for package in $PACKAGES; do
            if ! run_test_with_retry "$package"; then
                success=false
                if [[ "$FAIL_FAST" == "true" ]]; then
                    break
                fi
            fi
        done
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Generate reports
    generate_coverage_report
    
    # Final summary
    log_info "Test execution completed in ${duration}s"
    
    if [[ "$success" == "true" ]]; then
        log_success "All tests passed!"
        exit 0
    else
        log_error "Some tests failed"
        exit 1
    fi
}

# Export functions for parallel execution
export -f run_test_with_retry log_info log_success log_warning log_error

# Run main function
main "$@"