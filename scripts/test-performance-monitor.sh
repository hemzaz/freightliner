#!/bin/bash
# Test Performance Monitoring and Optimization Script
# Monitors test execution times and provides optimization recommendations

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PERFORMANCE_LOG="${PROJECT_ROOT}/test-performance.log"
METRICS_DIR="${PROJECT_ROOT}/.test-metrics"
BASELINE_FILE="${METRICS_DIR}/test-baselines.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${timestamp} [${level}] ${message}" | tee -a "${PERFORMANCE_LOG}"
}

info() { log "INFO" "$@"; }
warn() { log "WARN" "${YELLOW}$*${NC}"; }
error() { log "ERROR" "${RED}$*${NC}"; }
success() { log "SUCCESS" "${GREEN}$*${NC}"; }
debug() { log "DEBUG" "${CYAN}$*${NC}"; }

# Initialize performance monitoring
init_performance_monitoring() {
    info "Initializing test performance monitoring..."
    
    mkdir -p "${METRICS_DIR}"
    
    # Create baseline file if it doesn't exist
    if [[ ! -f "${BASELINE_FILE}" ]]; then
        cat > "${BASELINE_FILE}" << 'EOF'
{
  "version": "1.0",
  "baselines": {
    "unit_tests": {
      "max_duration_seconds": 30,
      "expected_duration_seconds": 10,
      "timeout_seconds": 60
    },
    "integration_tests": {
      "max_duration_seconds": 120,
      "expected_duration_seconds": 45,
      "timeout_seconds": 300
    },
    "load_tests": {
      "max_duration_seconds": 180,
      "expected_duration_seconds": 90,
      "timeout_seconds": 600
    },
    "benchmark_tests": {
      "max_duration_seconds": 60,
      "expected_duration_seconds": 30,
      "timeout_seconds": 120
    }
  },
  "thresholds": {
    "performance_regression": 1.5,
    "timeout_buffer": 1.2,
    "critical_slowdown": 3.0
  }
}
EOF
        info "Created baseline configuration"
    fi
    
    success "Performance monitoring initialized"
}

# Run tests with performance monitoring
run_monitored_tests() {
    local test_type="${1:-all}"
    local test_pattern="${2:-./...}"
    
    info "Running monitored tests: ${test_type}"
    
    cd "${PROJECT_ROOT}"
    
    case "${test_type}" in
        "unit")
            run_unit_tests "${test_pattern}"
            ;;
        "integration")
            run_integration_tests "${test_pattern}"
            ;;
        "load")
            run_load_tests "${test_pattern}"
            ;;
        "benchmark")
            run_benchmark_tests "${test_pattern}"
            ;;
        "all")
            run_unit_tests "${test_pattern}"
            run_integration_tests "${test_pattern}"
            run_benchmark_tests "${test_pattern}"
            # Load tests are optional for 'all' due to their duration
            if [[ "${CI:-false}" != "true" ]]; then
                run_load_tests "${test_pattern}"
            else
                info "Skipping load tests in CI environment"
            fi
            ;;
        *)
            error "Unknown test type: ${test_type}"
            return 1
            ;;
    esac
}

# Run unit tests with monitoring
run_unit_tests() {
    local pattern="$1"
    info "Running unit tests with performance monitoring..."
    
    local start_time=$(date +%s)
    local test_output
    local exit_code=0
    
    # Run tests with timeout and capture output
    if test_output=$(timeout 60s go test -v -race -short -coverprofile=coverage.out "${pattern}" 2>&1); then
        success "Unit tests passed"
    else
        exit_code=$?
        if [[ ${exit_code} -eq 124 ]]; then
            error "Unit tests timed out after 60 seconds"
        else
            error "Unit tests failed with exit code ${exit_code}"
        fi
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Record metrics
    record_test_metrics "unit_tests" "${duration}" "${exit_code}"
    
    # Analyze performance
    analyze_test_performance "unit_tests" "${duration}"
    
    info "Unit tests completed in ${duration} seconds"
    echo "${test_output}"
    
    return ${exit_code}
}

# Run integration tests with monitoring
run_integration_tests() {
    local pattern="$1"
    info "Running integration tests with performance monitoring..."
    
    # Check if integration tests exist
    if ! find . -name "*integration*test.go" -o -name "*_integration_test.go" | grep -q .; then
        info "No integration tests found"
        return 0
    fi
    
    local start_time=$(date +%s)
    local test_output
    local exit_code=0
    
    # Set environment for integration tests
    export INTEGRATION_TESTS=true
    
    # Run integration tests with extended timeout
    if test_output=$(timeout 300s go test -v -race -tags=integration "${pattern}" 2>&1); then
        success "Integration tests passed"
    else
        exit_code=$?
        if [[ ${exit_code} -eq 124 ]]; then
            error "Integration tests timed out after 300 seconds"
        else
            error "Integration tests failed with exit code ${exit_code}"
        fi
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Record metrics
    record_test_metrics "integration_tests" "${duration}" "${exit_code}"
    
    # Analyze performance
    analyze_test_performance "integration_tests" "${duration}"
    
    info "Integration tests completed in ${duration} seconds"
    echo "${test_output}"
    
    return ${exit_code}
}

# Run load tests with monitoring
run_load_tests() {
    local pattern="$1"
    info "Running load tests with performance monitoring..."
    
    # Check if load tests exist
    if ! find . -path "*/load/*test.go" | grep -q .; then
        info "No load tests found"
        return 0
    fi
    
    local start_time=$(date +%s)
    local test_output
    local exit_code=0
    
    # Set optimized timeouts for load tests
    export LOAD_TEST_TIMEOUT="120s"  # Reduced from original 122s
    export LOAD_TEST_DURATION="30s"  # Reasonable duration for CI
    
    # Run load tests with optimized timeout
    if test_output=$(timeout 600s go test -v "./pkg/testing/load" -timeout=120s 2>&1); then
        success "Load tests passed"
    else
        exit_code=$?
        if [[ ${exit_code} -eq 124 ]]; then
            error "Load tests timed out after 600 seconds"
        else
            error "Load tests failed with exit code ${exit_code}"
        fi
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Record metrics
    record_test_metrics "load_tests" "${duration}" "${exit_code}"
    
    # Analyze performance
    analyze_test_performance "load_tests" "${duration}"
    
    info "Load tests completed in ${duration} seconds"
    echo "${test_output}"
    
    return ${exit_code}
}

# Run benchmark tests with monitoring
run_benchmark_tests() {
    local pattern="$1"
    info "Running benchmark tests with performance monitoring..."
    
    # Check if benchmark tests exist
    if ! find . -name "*_test.go" -exec grep -l "func Benchmark" {} \; | grep -q .; then
        info "No benchmark tests found"
        return 0
    fi
    
    local start_time=$(date +%s)
    local test_output
    local exit_code=0
    
    # Run benchmarks with reasonable settings for CI
    if test_output=$(timeout 120s go test -bench=. -benchmem -benchtime=1s -count=1 "${pattern}" 2>&1); then
        success "Benchmark tests passed"
    else
        exit_code=$?
        if [[ ${exit_code} -eq 124 ]]; then
            error "Benchmark tests timed out after 120 seconds"
        else
            error "Benchmark tests failed with exit code ${exit_code}"
        fi
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Record metrics
    record_test_metrics "benchmark_tests" "${duration}" "${exit_code}"
    
    # Analyze performance
    analyze_test_performance "benchmark_tests" "${duration}"
    
    info "Benchmark tests completed in ${duration} seconds"
    echo "${test_output}"
    
    return ${exit_code}
}\n\n# Record test metrics\nrecord_test_metrics() {\n    local test_type=\"$1\"\n    local duration=\"$2\"\n    local exit_code=\"$3\"\n    \n    local metrics_file=\"${METRICS_DIR}/${test_type}-$(date +%Y%m%d).json\"\n    local timestamp=$(date -Iseconds)\n    \n    # Create or update metrics file\n    if [[ ! -f \"${metrics_file}\" ]]; then\n        echo '{\"runs\": []}' > \"${metrics_file}\"\n    fi\n    \n    # Add new run data\n    local new_run=$(cat << EOF\n{\n  \"timestamp\": \"${timestamp}\",\n  \"duration_seconds\": ${duration},\n  \"exit_code\": ${exit_code},\n  \"success\": $(if [[ ${exit_code} -eq 0 ]]; then echo \"true\"; else echo \"false\"; fi),\n  \"environment\": \"${CI:-local}\"\n}\nEOF\n)\n    \n    # Update metrics file using jq if available, otherwise append\n    if command -v jq &> /dev/null; then\n        local temp_file=$(mktemp)\n        jq \".runs += [${new_run}]\" \"${metrics_file}\" > \"${temp_file}\"\n        mv \"${temp_file}\" \"${metrics_file}\"\n    else\n        debug \"jq not available, metrics not recorded\"\n    fi\n    \n    debug \"Recorded metrics for ${test_type}: ${duration}s (exit: ${exit_code})\"\n}\n\n# Analyze test performance\nanalyze_test_performance() {\n    local test_type=\"$1\"\n    local duration=\"$2\"\n    \n    if ! command -v jq &> /dev/null; then\n        warn \"jq not available, skipping performance analysis\"\n        return 0\n    fi\n    \n    # Get baseline expectations\n    local expected_duration\n    local max_duration\n    local timeout_duration\n    \n    expected_duration=$(jq -r \".baselines.${test_type}.expected_duration_seconds\" \"${BASELINE_FILE}\" 2>/dev/null || echo \"0\")\n    max_duration=$(jq -r \".baselines.${test_type}.max_duration_seconds\" \"${BASELINE_FILE}\" 2>/dev/null || echo \"0\")\n    timeout_duration=$(jq -r \".baselines.${test_type}.timeout_seconds\" \"${BASELINE_FILE}\" 2>/dev/null || echo \"0\")\n    \n    if [[ \"${expected_duration}\" == \"null\" ]] || [[ \"${expected_duration}\" == \"0\" ]]; then\n        info \"No baseline found for ${test_type}, establishing new baseline\"\n        return 0\n    fi\n    \n    # Performance analysis\n    local performance_ratio\n    performance_ratio=$(echo \"scale=2; ${duration} / ${expected_duration}\" | bc -l 2>/dev/null || echo \"1.0\")\n    \n    if (( $(echo \"${duration} <= ${expected_duration}\" | bc -l) )); then\n        success \"${test_type} performance: EXCELLENT (${duration}s <= ${expected_duration}s)\"\n    elif (( $(echo \"${duration} <= ${max_duration}\" | bc -l) )); then\n        info \"${test_type} performance: ACCEPTABLE (${duration}s <= ${max_duration}s)\"\n    elif (( $(echo \"${duration} <= ${timeout_duration}\" | bc -l) )); then\n        warn \"${test_type} performance: SLOW (${duration}s > ${max_duration}s, ratio: ${performance_ratio}x)\"\n        suggest_optimizations \"${test_type}\" \"${performance_ratio}\"\n    else\n        error \"${test_type} performance: CRITICAL (${duration}s approaching timeout ${timeout_duration}s)\"\n        suggest_optimizations \"${test_type}\" \"${performance_ratio}\"\n    fi\n}\n\n# Suggest performance optimizations\nsuggest_optimizations() {\n    local test_type=\"$1\"\n    local performance_ratio=\"$2\"\n    \n    warn \"Performance optimization suggestions for ${test_type}:\"\n    \n    case \"${test_type}\" in\n        \"unit_tests\")\n            warn \"  - Use testing.Short() to skip expensive operations in CI\"\n            warn \"  - Mock external dependencies and I/O operations\"\n            warn \"  - Reduce test data sizes\"\n            warn \"  - Consider parallel test execution with t.Parallel()\"\n            ;;            \n        \"integration_tests\")\n            warn \"  - Use test containers or in-memory databases\"\n            warn \"  - Implement proper test isolation and cleanup\"\n            warn \"  - Use context.WithTimeout for better timeout control\"\n            warn \"  - Consider running integration tests in parallel\"\n            ;;            \n        \"load_tests\")\n            warn \"  - Reduce test duration in CI environments\"\n            warn \"  - Use smaller dataset sizes for CI\"\n            warn \"  - Implement adaptive timeout strategies\"\n            warn \"  - Consider using testing.Short() for quick CI runs\"\n            ;;            \n        \"benchmark_tests\")\n            warn \"  - Use shorter benchtime in CI (e.g., -benchtime=1s)\"\n            warn \"  - Reduce benchmark iteration counts\"\n            warn \"  - Focus on relative performance rather than absolute numbers\"\n            warn \"  - Consider running benchmarks separately from other tests\"\n            ;;            \n    esac\n    \n    if (( $(echo \"${performance_ratio} > 2.0\" | bc -l) )); then\n        error \"  - URGENT: Test is >2x slower than expected, requires immediate attention\"\n    fi\n}\n\n# Generate performance report\ngenerate_performance_report() {\n    info \"Generating test performance report...\"\n    \n    local report_file=\"${METRICS_DIR}/performance-report-$(date +%Y%m%d-%H%M%S).json\"\n    \n    if ! command -v jq &> /dev/null; then\n        warn \"jq not available, generating basic report\"\n        \n        cat > \"${report_file}\" << EOF\n{\n  \"timestamp\": \"$(date -Iseconds)\",\n  \"summary\": \"Performance report generated without detailed metrics (jq not available)\",\n  \"baselines\": $(cat \"${BASELINE_FILE}\" 2>/dev/null || echo '{}')\n}\nEOF\n        info \"Basic report generated: ${report_file}\"\n        return 0\n    fi\n    \n    # Collect metrics from all test types\n    local report_data='{\n      \"timestamp\": \"'$(date -Iseconds)'\",\n      \"test_types\": {}\n    }'\n    \n    for test_type in \"unit_tests\" \"integration_tests\" \"load_tests\" \"benchmark_tests\"; do\n        local latest_metrics_file\n        latest_metrics_file=$(ls \"${METRICS_DIR}/${test_type}-\"*.json 2>/dev/null | tail -1 || echo \"\")\n        \n        if [[ -n \"${latest_metrics_file}\" ]] && [[ -f \"${latest_metrics_file}\" ]]; then\n            local latest_run\n            latest_run=$(jq '.runs[-1]' \"${latest_metrics_file}\" 2>/dev/null || echo 'null')\n            \n            if [[ \"${latest_run}\" != \"null\" ]]; then\n                report_data=$(echo \"${report_data}\" | jq \".test_types.${test_type} = ${latest_run}\")\n            fi\n        fi\n    done\n    \n    # Add baseline data\n    local baselines\n    baselines=$(cat \"${BASELINE_FILE}\" 2>/dev/null || echo '{}')\n    report_data=$(echo \"${report_data}\" | jq \".baselines = ${baselines}\")\n    \n    # Write report\n    echo \"${report_data}\" > \"${report_file}\"\n    \n    info \"Performance report generated: ${report_file}\"\n    \n    # Display summary\n    echo\n    echo -e \"${BLUE}=== TEST PERFORMANCE SUMMARY ===${NC}\"\n    echo \"Report: ${report_file}\"\n    echo \"Timestamp: $(date)\"\n    echo\n    \n    for test_type in \"unit_tests\" \"integration_tests\" \"load_tests\" \"benchmark_tests\"; do\n        local duration\n        local success\n        \n        duration=$(echo \"${report_data}\" | jq -r \".test_types.${test_type}.duration_seconds // \\\"N/A\\\"\" 2>/dev/null || echo \"N/A\")\n        success=$(echo \"${report_data}\" | jq -r \".test_types.${test_type}.success // false\" 2>/dev/null || echo \"false\")\n        \n        if [[ \"${duration}\" != \"N/A\" ]]; then\n            if [[ \"${success}\" == \"true\" ]]; then\n                echo -e \"${GREEN}✓${NC} ${test_type}: ${duration}s\"\n            else\n                echo -e \"${RED}✗${NC} ${test_type}: ${duration}s (failed)\"\n            fi\n        else\n            echo -e \"${YELLOW}-${NC} ${test_type}: Not run\"\n        fi\n    done\n    \n    echo\n}\n\n# Optimize test configurations based on performance data\noptimize_test_configurations() {\n    info \"Optimizing test configurations based on performance data...\"\n    \n    cd \"${PROJECT_ROOT}\"\n    \n    # Check for load test configuration files\n    local load_test_files\n    if load_test_files=$(find . -path \"*/load/*test.go\" -type f); then\n        info \"Found load test files, checking for optimization opportunities...\"\n        \n        while IFS= read -r file; do\n            if [[ -n \"${file}\" ]]; then\n                info \"Analyzing: ${file}\"\n                \n                # Check for hardcoded timeouts that are too long\n                if grep -q \"[2-9][0-9].*time\\.Minute\\|[1-9][0-9][0-9].*time\\.Second\" \"${file}\"; then\n                    warn \"Found potentially long timeouts in ${file}\"\n                    warn \"Consider using environment variables or testing.Short() checks\"\n                fi\n                \n                # Check for CI-specific optimizations\n                if ! grep -q \"testing\\.Short()\" \"${file}\"; then\n                    warn \"${file} doesn't use testing.Short() - consider adding CI optimizations\"\n                fi\n            fi\n        done <<< \"${load_test_files}\"\n    fi\n    \n    success \"Test configuration analysis completed\"\n}\n\n# Main execution function\nmain() {\n    local command=\"${1:-run}\"\n    local test_type=\"${2:-all}\"\n    local test_pattern=\"${3:-./...}\"\n    \n    echo -e \"${BLUE}=== FREIGHTLINER TEST PERFORMANCE MONITOR ===${NC}\"\n    echo \"Command: ${command}\"\n    echo \"Test Type: ${test_type}\"\n    echo \"Pattern: ${test_pattern}\"\n    echo\n    \n    init_performance_monitoring\n    \n    case \"${command}\" in\n        \"run\")\n            run_monitored_tests \"${test_type}\" \"${test_pattern}\"\n            ;;\n        \"analyze\")\n            analyze_test_performance \"${test_type}\" \"${test_pattern}\"\n            ;;\n        \"report\")\n            generate_performance_report\n            ;;\n        \"optimize\")\n            optimize_test_configurations\n            ;;\n        *)\n            echo \"Usage: $0 [run|analyze|report|optimize] [test_type] [pattern]\"\n            echo\n            echo \"Commands:\"\n            echo \"  run      - Run tests with performance monitoring (default)\"\n            echo \"  analyze  - Analyze test performance\"\n            echo \"  report   - Generate performance report\"\n            echo \"  optimize - Optimize test configurations\"\n            echo\n            echo \"Test Types:\"\n            echo \"  unit         - Unit tests only\"\n            echo \"  integration  - Integration tests only\"\n            echo \"  load         - Load tests only\"\n            echo \"  benchmark    - Benchmark tests only\"\n            echo \"  all          - All test types (default)\"\n            echo\n            echo \"Examples:\"\n            echo \"  $0 run unit                    # Run unit tests with monitoring\"\n            echo \"  $0 run load ./pkg/testing/load # Run load tests for specific package\"\n            echo \"  $0 report                      # Generate performance report\"\n            echo \"  $0 optimize                    # Analyze and suggest optimizations\"\n            exit 1\n            ;;\n    esac\n}\n\n# Ensure bc is available for calculations\nif ! command -v bc &> /dev/null; then\n    warn \"bc (calculator) not available - some performance calculations may be skipped\"\nfi\n\n# Run main function with all arguments\nmain \"$@\"