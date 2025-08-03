#!/bin/bash

# CI/CD Reliability Enhancement Script
# Provides comprehensive error recovery, retry mechanisms, and circuit breaker functionality
# for GitHub Actions CI/CD pipeline operations

set -euo pipefail

# ==============================================================================
# CONFIGURATION AND CONSTANTS
# ==============================================================================

# Retry configuration
readonly DEFAULT_MAX_RETRIES=5
readonly DEFAULT_INITIAL_WAIT=2
readonly DEFAULT_MAX_WAIT=120
readonly DEFAULT_BACKOFF_FACTOR=2

# Circuit breaker configuration
readonly CIRCUIT_BREAKER_FAILURE_THRESHOLD=3
readonly CIRCUIT_BREAKER_TIMEOUT=300  # 5 minutes
readonly CIRCUIT_BREAKER_RESET_TIMEOUT=60  # 1 minute

# Timeout configuration
readonly DEFAULT_COMMAND_TIMEOUT=600  # 10 minutes
readonly HEALTH_CHECK_TIMEOUT=30
readonly NETWORK_OPERATION_TIMEOUT=60

# Logging configuration
readonly LOG_LEVEL="${LOG_LEVEL:-INFO}"
readonly LOG_FILE="${GITHUB_WORKSPACE:-/tmp}/ci-reliability.log"

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# State directories for circuit breakers
readonly STATE_DIR="${GITHUB_WORKSPACE:-/tmp}/.ci-reliability"
mkdir -p "${STATE_DIR}"

# ==============================================================================
# LOGGING FUNCTIONS
# ==============================================================================

log() {
    local level="$1"
    local message="$2"
    local timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    echo "${timestamp} [${level}] ${message}" | tee -a "${LOG_FILE}"
    
    case "${level}" in
        "ERROR")
            echo -e "${RED}❌ ${message}${NC}" >&2
            ;;
        "WARN")
            echo -e "${YELLOW}⚠️  ${message}${NC}" >&2
            ;;
        "INFO")
            echo -e "${GREEN}ℹ️  ${message}${NC}"
            ;;
        "DEBUG")
            if [[ "${LOG_LEVEL}" == "DEBUG" ]]; then
                echo -e "${BLUE}🔍 ${message}${NC}"
            fi
            ;;
    esac
}

log_error() { log "ERROR" "$1"; }
log_warn() { log "WARN" "$1"; }
log_info() { log "INFO" "$1"; }
log_debug() { log "DEBUG" "$1"; }

# ==============================================================================
# CIRCUIT BREAKER IMPLEMENTATION
# ==============================================================================

# Get circuit breaker state file path
get_circuit_breaker_state_file() {
    local service_name="$1"
    echo "${STATE_DIR}/circuit_breaker_${service_name}"
}

# Initialize circuit breaker for a service
init_circuit_breaker() {
    local service_name="$1"
    local state_file
    state_file=$(get_circuit_breaker_state_file "${service_name}")
    
    if [[ ! -f "${state_file}" ]]; then
        echo "CLOSED:0:0" > "${state_file}"
        log_debug "Initialized circuit breaker for ${service_name}"
    fi
}

# Get circuit breaker state
get_circuit_breaker_state() {
    local service_name="$1"
    local state_file
    state_file=$(get_circuit_breaker_state_file "${service_name}")
    
    if [[ -f "${state_file}" ]]; then
        cat "${state_file}"
    else
        echo "CLOSED:0:0"
    fi
}

# Update circuit breaker state
update_circuit_breaker_state() {
    local service_name="$1"
    local state="$2"
    local failure_count="$3"
    local last_failure_time="$4"
    local state_file
    state_file=$(get_circuit_breaker_state_file "${service_name}")
    
    echo "${state}:${failure_count}:${last_failure_time}" > "${state_file}"
    log_debug "Updated circuit breaker for ${service_name}: ${state}:${failure_count}:${last_failure_time}"
}

# Check if circuit breaker allows operation
is_circuit_breaker_open() {
    local service_name="$1"
    local current_time
    current_time=$(date +%s)
    
    init_circuit_breaker "${service_name}"
    
    local state_info
    state_info=$(get_circuit_breaker_state "${service_name}")
    
    IFS=':' read -r state failure_count last_failure_time <<< "${state_info}"
    
    case "${state}" in
        "CLOSED")
            return 1  # Circuit is closed, allow operation
            ;;
        "OPEN")
            local time_since_failure=$((current_time - last_failure_time))
            if [[ ${time_since_failure} -gt ${CIRCUIT_BREAKER_TIMEOUT} ]]; then
                # Move to half-open state
                update_circuit_breaker_state "${service_name}" "HALF_OPEN" "${failure_count}" "${current_time}"
                log_info "Circuit breaker for ${service_name} moved to HALF_OPEN state"
                return 1  # Allow one test operation
            else
                log_warn "Circuit breaker for ${service_name} is OPEN (${time_since_failure}s since last failure)"
                return 0  # Circuit is open, block operation
            fi
            ;;
        "HALF_OPEN")
            return 1  # Allow test operation
            ;;
    esac
}

# Record circuit breaker success
record_circuit_breaker_success() {
    local service_name="$1"
    local current_time
    current_time=$(date +%s)
    
    local state_info
    state_info=$(get_circuit_breaker_state "${service_name}")
    IFS=':' read -r state failure_count last_failure_time <<< "${state_info}"
    
    if [[ "${state}" == "HALF_OPEN" ]]; then
        update_circuit_breaker_state "${service_name}" "CLOSED" "0" "0"
        log_info "Circuit breaker for ${service_name} reset to CLOSED state"
    elif [[ "${state}" == "CLOSED" && ${failure_count} -gt 0 ]]; then
        update_circuit_breaker_state "${service_name}" "CLOSED" "0" "0"
        log_debug "Circuit breaker for ${service_name} failure count reset"
    fi
}

# Record circuit breaker failure
record_circuit_breaker_failure() {
    local service_name="$1"
    local current_time
    current_time=$(date +%s)
    
    local state_info
    state_info=$(get_circuit_breaker_state "${service_name}")
    IFS=':' read -r state failure_count last_failure_time <<< "${state_info}"
    
    failure_count=$((failure_count + 1))
    
    if [[ ${failure_count} -ge ${CIRCUIT_BREAKER_FAILURE_THRESHOLD} ]]; then
        update_circuit_breaker_state "${service_name}" "OPEN" "${failure_count}" "${current_time}"
        log_warn "Circuit breaker for ${service_name} opened due to ${failure_count} consecutive failures"
    else
        update_circuit_breaker_state "${service_name}" "${state}" "${failure_count}" "${current_time}"
        log_debug "Circuit breaker for ${service_name} recorded failure ${failure_count}/${CIRCUIT_BREAKER_FAILURE_THRESHOLD}"
    fi
}

# ==============================================================================
# ENHANCED RETRY MECHANISM
# ==============================================================================

# Enhanced retry function with circuit breaker and jitter
retry_with_circuit_breaker() {
    local operation_name="$1"
    local service_name="$2"
    local max_retries="${3:-${DEFAULT_MAX_RETRIES}}"
    local initial_wait="${4:-${DEFAULT_INITIAL_WAIT}}"
    local max_wait="${5:-${DEFAULT_MAX_WAIT}}"
    shift 5
    local command=("$@")
    
    log_info "Starting operation: ${operation_name} (service: ${service_name})"
    
    # Check circuit breaker
    if is_circuit_breaker_open "${service_name}"; then
        log_error "Circuit breaker is open for ${service_name}, skipping operation"
        return 1
    fi
    
    local attempt=0
    local wait_time=${initial_wait}
    
    while [[ ${attempt} -le ${max_retries} ]]; do
        if [[ ${attempt} -gt 0 ]]; then
            # Add jitter to wait time (±20%)
            local jitter_range=$((wait_time / 5))
            local jitter=$((RANDOM % (jitter_range * 2) - jitter_range))
            local actual_wait=$((wait_time + jitter))
            
            log_info "Retry attempt ${attempt}/${max_retries} for ${operation_name} in ${actual_wait}s"
            sleep ${actual_wait}
            
            # Exponential backoff with cap
            wait_time=$((wait_time * DEFAULT_BACKOFF_FACTOR))
            if [[ ${wait_time} -gt ${max_wait} ]]; then
                wait_time=${max_wait}
            fi
        fi
        
        log_debug "Executing: ${command[*]}"
        
        if timeout ${DEFAULT_COMMAND_TIMEOUT} "${command[@]}"; then
            log_info "Operation ${operation_name} succeeded on attempt $((attempt + 1))"
            record_circuit_breaker_success "${service_name}"
            return 0
        else
            local exit_code=$?
            log_warn "Operation ${operation_name} failed on attempt $((attempt + 1)) with exit code ${exit_code}"
            
            # Check if error is retryable
            if ! is_retryable_error ${exit_code}; then
                log_error "Non-retryable error encountered, stopping retries"
                record_circuit_breaker_failure "${service_name}"
                return ${exit_code}
            fi
        fi
        
        attempt=$((attempt + 1))
    done
    
    log_error "Operation ${operation_name} failed after ${max_retries} retries"
    record_circuit_breaker_failure "${service_name}"
    return 1
}

# Check if error code indicates a retryable error
is_retryable_error() {
    local exit_code="$1"
    
    # Non-retryable exit codes
    case ${exit_code} in
        2|126|127|128|130)  # Syntax errors, permission denied, command not found, etc.
            return 1
            ;;
        *)
            return 0  # Most errors are retryable
            ;;
    esac
}

# ==============================================================================
# NETWORK RESILIENCE FUNCTIONS
# ==============================================================================

# Test network connectivity with retry
test_network_connectivity() {
    local target="$1"
    local port="${2:-80}"
    local service_name="${3:-network}"
    
    log_info "Testing network connectivity to ${target}:${port}"
    
    retry_with_circuit_breaker \
        "network_connectivity_test" \
        "${service_name}" \
        3 2 30 \
        timeout ${NETWORK_OPERATION_TIMEOUT} nc -z "${target}" "${port}"
}

# Wait for service to be ready
wait_for_service_ready() {
    local service_name="$1"
    local health_check_url="$2"
    local max_wait_time="${3:-300}"
    local check_interval="${4:-5}"
    
    log_info "Waiting for ${service_name} to be ready at ${health_check_url}"
    
    local start_time
    start_time=$(date +%s)
    
    while true; do
        local current_time
        current_time=$(date +%s)
        local elapsed_time=$((current_time - start_time))
        
        if [[ ${elapsed_time} -gt ${max_wait_time} ]]; then
            log_error "${service_name} did not become ready within ${max_wait_time} seconds"
            return 1
        fi
        
        if timeout ${HEALTH_CHECK_TIMEOUT} curl -sf "${health_check_url}" >/dev/null 2>&1; then
            log_info "${service_name} is ready (took ${elapsed_time}s)"
            return 0
        fi
        
        log_debug "${service_name} not ready yet, waiting ${check_interval}s..."
        sleep ${check_interval}
    done
}

# ==============================================================================
# GO-SPECIFIC RELIABILITY FUNCTIONS
# ==============================================================================

# Enhanced Go module download with retry
go_mod_download_with_retry() {
    local service_name="${1:-go-proxy}"
    
    log_info "Downloading Go modules with retry mechanism"
    
    # Set Go proxy with fallbacks
    export GOPROXY="https://proxy.golang.org,https://goproxy.cn,https://goproxy.io,direct"
    export GOSUMDB="sum.golang.org"
    export GOPRIVATE=""
    
    retry_with_circuit_breaker \
        "go_mod_download" \
        "${service_name}" \
        5 3 60 \
        go mod download -x
    
    local download_result=$?
    
    if [[ ${download_result} -eq 0 ]]; then
        log_info "Verifying Go module checksums"
        retry_with_circuit_breaker \
            "go_mod_verify" \
            "${service_name}" \
            3 2 30 \
            go mod verify
    fi
    
    return ${download_result}
}

# Enhanced Go build with retry and caching
go_build_with_retry() {
    local build_target="${1:-.\/...}"
    local service_name="${2:-go-build}"
    
    log_info "Building Go packages: ${build_target}"
    
    # Enable build cache
    export GOCACHE="${GOCACHE:-${HOME}/.cache/go-build}"
    
    retry_with_circuit_breaker \
        "go_build" \
        "${service_name}" \
        3 5 60 \
        go build -v "${build_target}"
}

# Enhanced Go test execution with isolation
go_test_with_retry() {
    local test_type="${1:-all}"
    local service_name="${2:-go-test}"
    local test_flags=("${@:3}")
    
    log_info "Running Go tests: ${test_type}"
    
    # Configure test environment
    export GOFLAGS="${GOFLAGS:-} -mod=mod"
    export CGO_ENABLED="${CGO_ENABLED:-1}"
    
    local test_command=()
    
    case "${test_type}" in
        "unit")
            test_command=(go test -short "${test_flags[@]}" ./...)
            ;;
        "integration")
            test_command=(go test -run "Integration" "${test_flags[@]}" ./...)
            ;;
        "race")
            test_command=(go test -race "${test_flags[@]}" ./...)
            ;;
        "all"|*)
            test_command=(go test "${test_flags[@]}" ./...)
            ;;
    esac
    
    retry_with_circuit_breaker \
        "go_test_${test_type}" \
        "${service_name}" \
        2 10 120 \
        "${test_command[@]}"
}

# ==============================================================================
# DOCKER-SPECIFIC RELIABILITY FUNCTIONS
# ==============================================================================

# Enhanced Docker registry health check
check_docker_registry_health() {
    local registry_host="$1"
    local service_name="${2:-docker-registry}"
    
    log_info "Checking Docker registry health: ${registry_host}"
    
    local health_url="http://${registry_host}/v2/"
    
    retry_with_circuit_breaker \
        "docker_registry_health_check" \
        "${service_name}" \
        5 2 30 \
        timeout ${HEALTH_CHECK_TIMEOUT} curl -sf "${health_url}"
}

# Enhanced Docker build with retry and cache optimization
docker_build_with_retry() {
    local image_name="$1"
    local dockerfile="${2:-Dockerfile}"
    local build_context="${3:-.}"
    local service_name="${4:-docker-build}"
    
    log_info "Building Docker image: ${image_name}"
    
    # Configure Docker buildkit
    export DOCKER_BUILDKIT=1
    export BUILDX_NO_DEFAULT_ATTESTATIONS=1
    
    local build_args=(
        --file "${dockerfile}"
        --cache-from type=gha
        --cache-to type=gha,mode=max
        --tag "${image_name}"
        "${build_context}"
    )
    
    retry_with_circuit_breaker \
        "docker_build" \
        "${service_name}" \
        3 10 180 \
        docker buildx build "${build_args[@]}"
}

# ==============================================================================
# PACKAGE MANAGER RELIABILITY FUNCTIONS
# ==============================================================================

# Enhanced package installation with retry
install_packages_with_retry() {
    local package_manager="$1"
    local service_name="${2:-package-manager}"
    shift 2
    local packages=("$@")
    
    log_info "Installing packages with ${package_manager}: ${packages[*]}"
    
    case "${package_manager}" in
        "apt")
            retry_with_circuit_breaker \
                "apt_update" \
                "${service_name}" \
                3 5 60 \
                apt-get update
            
            retry_with_circuit_breaker \
                "apt_install" \
                "${service_name}" \
                3 5 60 \
                apt-get install -y "${packages[@]}"
            ;;
        "npm")
            retry_with_circuit_breaker \
                "npm_install" \
                "${service_name}" \
                3 10 120 \
                npm install "${packages[@]}"
            ;;
        "pip")
            retry_with_circuit_breaker \
                "pip_install" \
                "${service_name}" \
                3 10 120 \
                pip install "${packages[@]}"
            ;;
        "go")
            for package in "${packages[@]}"; do
                retry_with_circuit_breaker \
                    "go_install_${package//\//_}" \
                    "${service_name}" \
                    3 5 60 \
                    go install "${package}"
            done
            ;;
        *)
            log_error "Unsupported package manager: ${package_manager}"
            return 1
            ;;
    esac
}

# ==============================================================================
# CLEANUP AND RESOURCE MANAGEMENT
# ==============================================================================

# Cleanup resources and temporary files
cleanup_resources() {
    log_info "Cleaning up resources and temporary files"
    
    # Clean up Go build cache if too large
    if command -v go >/dev/null 2>&1; then
        local cache_size
        cache_size=$(du -sm "${GOCACHE:-${HOME}/.cache/go-build}" 2>/dev/null | cut -f1 || echo "0")
        if [[ ${cache_size} -gt 1000 ]]; then  # If cache > 1GB
            log_info "Go build cache is ${cache_size}MB, cleaning up"
            go clean -cache 2>/dev/null || true
        fi
    fi
    
    # Clean up Docker images and containers
    if command -v docker >/dev/null 2>&1; then
        log_info "Cleaning up Docker resources"
        docker system prune -f --volumes 2>/dev/null || true
    fi
    
    # Clean up old circuit breaker state files (older than 1 hour)
    find "${STATE_DIR}" -name "circuit_breaker_*" -mmin +60 -delete 2>/dev/null || true
    
    log_info "Resource cleanup completed"
}

# ==============================================================================
# MONITORING AND ALERTING
# ==============================================================================

# Generate pipeline health report
generate_health_report() {
    local report_file="${GITHUB_WORKSPACE:-/tmp}/pipeline-health-report.json"
    
    log_info "Generating pipeline health report"
    
    local current_time
    current_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    cat > "${report_file}" << EOF
{
  "timestamp": "${current_time}",
  "pipeline_run": "${GITHUB_RUN_ID:-unknown}",
  "workflow": "${GITHUB_WORKFLOW:-unknown}",
  "repository": "${GITHUB_REPOSITORY:-unknown}",
  "ref": "${GITHUB_REF:-unknown}",
  "circuit_breakers": [
EOF
    
    local first=true
    for state_file in "${STATE_DIR}"/circuit_breaker_*; do
        if [[ -f "${state_file}" ]]; then
            local service_name
            service_name=$(basename "${state_file}" | sed 's/circuit_breaker_//')
            local state_info
            state_info=$(cat "${state_file}")
            
            if [[ "${first}" == "false" ]]; then
                echo "," >> "${report_file}"
            fi
            first=false
            
            IFS=':' read -r state failure_count last_failure_time <<< "${state_info}"
            
            cat >> "${report_file}" << EOF
    {
      "service": "${service_name}",
      "state": "${state}",
      "failure_count": ${failure_count},
      "last_failure_time": ${last_failure_time}
    }
EOF
        fi
    done
    
    cat >> "${report_file}" << EOF
  ],
  "log_file": "${LOG_FILE}"
}
EOF
    
    log_info "Pipeline health report generated: ${report_file}"
    
    # Output report to GitHub step summary if available
    if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
        {
            echo "## Pipeline Health Report"
            echo ""
            echo "**Timestamp:** ${current_time}"
            echo "**Pipeline Run:** ${GITHUB_RUN_ID:-unknown}"
            echo "**Workflow:** ${GITHUB_WORKFLOW:-unknown}"
            echo ""
            echo "### Circuit Breaker Status"
            echo ""
            echo "| Service | State | Failure Count | Last Failure |"
            echo "|---------|--------|---------------|--------------|"
            
            for state_file in "${STATE_DIR}"/circuit_breaker_*; do
                if [[ -f "${state_file}" ]]; then
                    local service_name
                    service_name=$(basename "${state_file}" | sed 's/circuit_breaker_//')
                    local state_info
                    state_info=$(cat "${state_file}")
                    IFS=':' read -r state failure_count last_failure_time <<< "${state_info}"
                    
                    local last_failure_display
                    if [[ ${last_failure_time} -eq 0 ]]; then
                        last_failure_display="Never"
                    else
                        last_failure_display=$(date -d "@${last_failure_time}" '+%Y-%m-%d %H:%M:%S' 2>/dev/null || echo "Unknown")
                    fi
                    
                    echo "| ${service_name} | ${state} | ${failure_count} | ${last_failure_display} |"
                fi
            done
        } >> "${GITHUB_STEP_SUMMARY}"
    fi
}

# ==============================================================================
# MAIN FUNCTION AND CLI INTERFACE
# ==============================================================================

# Show usage information
show_usage() {
    cat << EOF
CI/CD Reliability Enhancement Script

Usage: $0 <command> [options]

Commands:
  retry <name> <service> <command...>     Execute command with retry and circuit breaker
  go-download [service]                   Download Go modules with retry
  go-build [target] [service]             Build Go packages with retry
  go-test <type> [service] [flags...]     Run Go tests with retry
  docker-health <registry> [service]     Check Docker registry health
  docker-build <image> [dockerfile] [context] [service]  Build Docker image
  install <manager> [service] <packages...>  Install packages with retry
  network-test <host> [port] [service]   Test network connectivity
  wait-service <name> <url> [timeout] [interval]  Wait for service readiness
  cleanup                                 Clean up resources
  health-report                          Generate health report

Options:
  --max-retries <n>       Maximum number of retries (default: ${DEFAULT_MAX_RETRIES})
  --initial-wait <s>      Initial wait time in seconds (default: ${DEFAULT_INITIAL_WAIT})
  --max-wait <s>          Maximum wait time in seconds (default: ${DEFAULT_MAX_WAIT})
  --log-level <level>     Log level: DEBUG, INFO, WARN, ERROR (default: INFO)
  --help                  Show this help message

Examples:
  $0 retry "install_gosec" "go-tools" go install github.com/securego/gosec/v2/cmd/gosec@latest
  $0 go-download
  $0 go-test unit --race --coverprofile=coverage.out
  $0 docker-health localhost:5100
  $0 cleanup

Environment Variables:
  LOG_LEVEL              Set logging level (DEBUG, INFO, WARN, ERROR)
  GITHUB_WORKSPACE       GitHub Actions workspace (used for state storage)
EOF
}

# Main function
main() {
    local command="${1:-}"
    
    if [[ $# -eq 0 || "${command}" == "--help" || "${command}" == "-h" ]]; then
        show_usage
        exit 0
    fi
    
    # Parse global options
    while [[ $# -gt 0 ]]; do
        case $1 in
            --max-retries)
                DEFAULT_MAX_RETRIES="$2"
                shift 2
                ;;
            --initial-wait)
                DEFAULT_INITIAL_WAIT="$2"
                shift 2
                ;;
            --max-wait)
                DEFAULT_MAX_WAIT="$2"
                shift 2
                ;;
            --log-level)
                LOG_LEVEL="$2"
                shift 2
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            *)
                break
                ;;
        esac
    done
    
    # Execute command
    case "${command}" in
        "retry")
            if [[ $# -lt 4 ]]; then
                log_error "Usage: $0 retry <name> <service> <command...>"
                exit 1
            fi
            local operation_name="$2"
            local service_name="$3"
            shift 3
            retry_with_circuit_breaker "${operation_name}" "${service_name}" "${DEFAULT_MAX_RETRIES}" "${DEFAULT_INITIAL_WAIT}" "${DEFAULT_MAX_WAIT}" "$@"
            ;;
        "go-download")
            local service_name="${2:-go-proxy}"
            go_mod_download_with_retry "${service_name}"
            ;;
        "go-build")
            local target="${2:-.\/...}"
            local service_name="${3:-go-build}"
            go_build_with_retry "${target}" "${service_name}"
            ;;
        "go-test")
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 go-test <type> [service] [flags...]"
                exit 1
            fi
            local test_type="$2"
            local service_name="${3:-go-test}"
            shift 3
            go_test_with_retry "${test_type}" "${service_name}" "$@"
            ;;
        "docker-health")
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 docker-health <registry> [service]"
                exit 1
            fi
            local registry="$2"
            local service_name="${3:-docker-registry}"
            check_docker_registry_health "${registry}" "${service_name}"
            ;;
        "docker-build")
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 docker-build <image> [dockerfile] [context] [service]"
                exit 1
            fi
            local image="$2"
            local dockerfile="${3:-Dockerfile}"
            local context="${4:-.}"
            local service_name="${5:-docker-build}"
            docker_build_with_retry "${image}" "${dockerfile}" "${context}" "${service_name}"
            ;;
        "install")
            if [[ $# -lt 3 ]]; then
                log_error "Usage: $0 install <manager> [service] <packages...>"
                exit 1
            fi
            local manager="$2"
            local service_name="${3:-package-manager}"
            shift 3
            install_packages_with_retry "${manager}" "${service_name}" "$@"
            ;;
        "network-test")
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 network-test <host> [port] [service]"
                exit 1
            fi
            local host="$2"
            local port="${3:-80}"
            local service_name="${4:-network}"
            test_network_connectivity "${host}" "${port}" "${service_name}"
            ;;
        "wait-service")
            if [[ $# -lt 3 ]]; then
                log_error "Usage: $0 wait-service <name> <url> [timeout] [interval]"
                exit 1
            fi
            local name="$2"
            local url="$3"
            local timeout="${4:-300}"
            local interval="${5:-5}"
            wait_for_service_ready "${name}" "${url}" "${timeout}" "${interval}"
            ;;
        "cleanup")
            cleanup_resources
            ;;
        "health-report")
            generate_health_report
            ;;
        *)
            log_error "Unknown command: ${command}"
            show_usage
            exit 1
            ;;
    esac
}

# Trap signals for cleanup
trap cleanup_resources EXIT INT TERM

# Execute main function if script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi