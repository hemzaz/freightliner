#!/bin/bash

# Comprehensive Integration Test Runner
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_THRESHOLD=85
PARALLEL_JOBS=${PARALLEL_JOBS:-4}
TEST_TIMEOUT=${TEST_TIMEOUT:-20m}

echo "🧪 Freightliner Integration Test Suite"
echo "========================================"
echo ""

# Parse arguments
RUN_UNIT=true
RUN_INTEGRATION=true
RUN_E2E=false
RUN_PERFORMANCE=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --unit-only)
            RUN_INTEGRATION=false
            RUN_E2E=false
            RUN_PERFORMANCE=false
            shift
            ;;
        --integration-only)
            RUN_UNIT=false
            RUN_E2E=false
            RUN_PERFORMANCE=false
            shift
            ;;
        --e2e)
            RUN_E2E=true
            shift
            ;;
        --performance)
            RUN_PERFORMANCE=true
            shift
            ;;
        --all)
            RUN_UNIT=true
            RUN_INTEGRATION=true
            RUN_E2E=true
            RUN_PERFORMANCE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--unit-only|--integration-only|--e2e|--performance|--all] [-v|--verbose]"
            exit 1
            ;;
    esac
done

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "success")
            echo -e "${GREEN}✅ ${message}${NC}"
            ;;
        "error")
            echo -e "${RED}❌ ${message}${NC}"
            ;;
        "warning")
            echo -e "${YELLOW}⚠️  ${message}${NC}"
            ;;
        "info")
            echo -e "ℹ️  ${message}"
            ;;
    esac
}

# Function to run tests with proper error handling
run_test_suite() {
    local test_name=$1
    local test_path=$2
    local test_flags=$3

    print_status "info" "Running ${test_name}..."

    if $VERBOSE; then
        TEST_VERBOSITY="-v"
    else
        TEST_VERBOSITY=""
    fi

    if go test $TEST_VERBOSITY -timeout=$TEST_TIMEOUT -parallel=$PARALLEL_JOBS $test_flags $test_path; then
        print_status "success" "${test_name} passed"
        return 0
    else
        print_status "error" "${test_name} failed"
        return 1
    fi
}

# Track overall success
OVERALL_SUCCESS=true

# 1. Run Unit Tests
if $RUN_UNIT; then
    echo ""
    print_status "info" "Phase 1: Unit Tests"
    echo "-------------------"

    if run_test_suite "Unit Tests" "./..." "-short -coverprofile=coverage.out -covermode=atomic"; then
        # Check coverage
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        COVERAGE_INT=${COVERAGE%.*}

        echo ""
        print_status "info" "Code Coverage: ${COVERAGE}%"

        if [ $COVERAGE_INT -lt $COVERAGE_THRESHOLD ]; then
            print_status "warning" "Coverage ${COVERAGE}% is below threshold ${COVERAGE_THRESHOLD}%"
        else
            print_status "success" "Coverage meets threshold"
        fi

        # Generate HTML coverage report
        go tool cover -html=coverage.out -o coverage.html
        print_status "info" "Coverage report: coverage.html"
    else
        OVERALL_SUCCESS=false
    fi
fi

# 2. Run Integration Tests
if $RUN_INTEGRATION; then
    echo ""
    print_status "info" "Phase 2: Integration Tests"
    echo "---------------------------"

    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        print_status "warning" "Docker not found, skipping integration tests"
    else
        # Setup test infrastructure
        print_status "info" "Setting up test infrastructure..."

        if [ -f "${PROJECT_ROOT}/tests/integration/setup_test.sh" ]; then
            bash "${PROJECT_ROOT}/tests/integration/setup_test.sh" || {
                print_status "error" "Failed to setup test infrastructure"
                OVERALL_SUCCESS=false
            }
        fi

        # Run integration tests
        if ! run_test_suite "Integration Tests" "./tests/integration/..." "-tags=integration"; then
            OVERALL_SUCCESS=false
        fi

        # Cleanup test infrastructure
        if [ -f "${PROJECT_ROOT}/tests/integration/teardown_test.sh" ]; then
            print_status "info" "Cleaning up test infrastructure..."
            bash "${PROJECT_ROOT}/tests/integration/teardown_test.sh" || {
                print_status "warning" "Failed to cleanup test infrastructure"
            }
        fi
    fi
fi

# 3. Run E2E Tests
if $RUN_E2E; then
    echo ""
    print_status "info" "Phase 3: End-to-End Tests"
    echo "--------------------------"

    # Check if binary exists
    if [ ! -f "${PROJECT_ROOT}/bin/freightliner" ]; then
        print_status "info" "Building freightliner binary..."
        make -C "${PROJECT_ROOT}" build || {
            print_status "error" "Failed to build binary"
            OVERALL_SUCCESS=false
        }
    fi

    if [ -f "${PROJECT_ROOT}/bin/freightliner" ]; then
        if ! run_test_suite "E2E Tests" "./tests/e2e/..." "-tags=e2e"; then
            OVERALL_SUCCESS=false
        fi
    else
        print_status "error" "Binary not found, skipping E2E tests"
        OVERALL_SUCCESS=false
    fi
fi

# 4. Run Performance Tests
if $RUN_PERFORMANCE; then
    echo ""
    print_status "info" "Phase 4: Performance Tests"
    echo "---------------------------"

    BENCH_OUTPUT="${PROJECT_ROOT}/bench-results.txt"

    if go test -v -bench=. -benchmem -timeout=20m ./tests/performance/... | tee "$BENCH_OUTPUT"; then
        print_status "success" "Performance tests completed"
        print_status "info" "Results saved to: $BENCH_OUTPUT"

        # Show summary
        echo ""
        echo "Performance Summary:"
        grep -E "Benchmark.*-[0-9]+" "$BENCH_OUTPUT" | tail -10 || true
    else
        print_status "error" "Performance tests failed"
        OVERALL_SUCCESS=false
    fi
fi

# Final Summary
echo ""
echo "========================================"
if $OVERALL_SUCCESS; then
    print_status "success" "All test suites passed! 🎉"
    exit 0
else
    print_status "error" "Some test suites failed"
    exit 1
fi
