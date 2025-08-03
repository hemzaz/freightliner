#!/bin/bash

# Pipeline Error Recovery Script
# Provides comprehensive error recovery, diagnostics, and state management
# for GitHub Actions CI/CD pipeline operations

set -euo pipefail

# ==============================================================================
# CONFIGURATION AND CONSTANTS
# ==============================================================================

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly WORKSPACE_DIR="${GITHUB_WORKSPACE:-$(pwd)}"
readonly RECOVERY_STATE_DIR="${WORKSPACE_DIR}/.pipeline-recovery"
readonly LOG_FILE="${RECOVERY_STATE_DIR}/recovery.log"

# Recovery timeouts and limits
readonly MAX_RECOVERY_ATTEMPTS=3
readonly RECOVERY_TIMEOUT=300  # 5 minutes
readonly HEALTH_CHECK_INTERVAL=10
readonly COMPONENT_RESTART_DELAY=30

# Component health check timeouts
readonly GO_HEALTH_TIMEOUT=60
readonly DOCKER_HEALTH_TIMEOUT=120
readonly REGISTRY_HEALTH_TIMEOUT=60
readonly NETWORK_HEALTH_TIMEOUT=30

# Pipeline state management
readonly PIPELINE_STATE_FILE="${RECOVERY_STATE_DIR}/pipeline-state.json"
readonly COMPONENT_STATUS_FILE="${RECOVERY_STATE_DIR}/component-status.json"
readonly RECOVERY_HISTORY_FILE="${RECOVERY_STATE_DIR}/recovery-history.log"

# Colors and formatting
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# ==============================================================================
# LOGGING AND UTILITIES
# ==============================================================================

# Initialize recovery environment
init_recovery_environment() {
    mkdir -p "${RECOVERY_STATE_DIR}"
    touch "${LOG_FILE}"
    
    # Initialize pipeline state if not exists
    if [[ ! -f "${PIPELINE_STATE_FILE}" ]]; then
        cat > "${PIPELINE_STATE_FILE}" << EOF
{
  "pipeline_id": "${GITHUB_RUN_ID:-unknown}",
  "workflow": "${GITHUB_WORKFLOW:-unknown}",
  "repository": "${GITHUB_REPOSITORY:-unknown}",
  "ref": "${GITHUB_REF:-unknown}",
  "started_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "recovery_enabled": true,
  "recovery_attempts": 0
}
EOF
    fi
    
    # Initialize component status
    if [[ ! -f "${COMPONENT_STATUS_FILE}" ]]; then
        cat > "${COMPONENT_STATUS_FILE}" << EOF
{
  "go": {"status": "unknown", "last_check": "", "failures": 0},
  "docker": {"status": "unknown", "last_check": "", "failures": 0},
  "registry": {"status": "unknown", "last_check": "", "failures": 0},
  "network": {"status": "unknown", "last_check": "", "failures": 0},
  "cache": {"status": "unknown", "last_check": "", "failures": 0}
}
EOF
    fi
}

# Enhanced logging with structured format
log_structured() {
    local level="$1"
    local component="$2"
    local message="$3"
    local extra_data="${4:-}"
    
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    local log_entry
    if [[ -n "$extra_data" ]]; then
        log_entry="{\"timestamp\":\"$timestamp\",\"level\":\"$level\",\"component\":\"$component\",\"message\":\"$message\",\"data\":$extra_data}"
    else
        log_entry="{\"timestamp\":\"$timestamp\",\"level\":\"$level\",\"component\":\"$component\",\"message\":\"$message\"}"
    fi
    
    echo "$log_entry" >> "$LOG_FILE"
    
    # Also output to console with formatting
    local color=""
    local icon=""
    
    case "$level" in
        "ERROR")
            color="$RED"
            icon="❌"
            ;;
        "WARN")
            color="$YELLOW"
            icon="⚠️"
            ;;
        "INFO")
            color="$GREEN"
            icon="ℹ️"
            ;;
        "DEBUG")
            color="$BLUE"
            icon="🔍"
            ;;
        "RECOVERY")
            color="$PURPLE"
            icon="🔄"
            ;;
    esac
    
    echo -e "${color}${icon} [${component}] ${message}${NC}"
}

log_error() { log_structured "ERROR" "${2:-recovery}" "$1" "${3:-}"; }
log_warn() { log_structured "WARN" "${2:-recovery}" "$1" "${3:-}"; }
log_info() { log_structured "INFO" "${2:-recovery}" "$1" "${3:-}"; }
log_debug() { log_structured "DEBUG" "${2:-recovery}" "$1" "${3:-}"; }
log_recovery() { log_structured "RECOVERY" "${2:-recovery}" "$1" "${3:-}"; }

# ==============================================================================
# COMPONENT HEALTH CHECKS
# ==============================================================================

# Update component status in JSON file
update_component_status() {
    local component="$1"
    local status="$2"
    local failure_increment="${3:-0}"
    
    local current_time
    current_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Use jq if available, otherwise use sed (less reliable but more portable)
    if command -v jq >/dev/null 2>&1; then
        local temp_file
        temp_file=$(mktemp)
        
        if [[ "$failure_increment" -gt 0 ]]; then
            jq --arg comp "$component" --arg status "$status" --arg time "$current_time" \
               '.[$comp].status = $status | .[$comp].last_check = $time | .[$comp].failures = (.[$comp].failures + 1)' \
               "$COMPONENT_STATUS_FILE" > "$temp_file"
        else
            jq --arg comp "$component" --arg status "$status" --arg time "$current_time" \
               '.[$comp].status = $status | .[$comp].last_check = $time' \
               "$COMPONENT_STATUS_FILE" > "$temp_file"
        fi
        
        mv "$temp_file" "$COMPONENT_STATUS_FILE"
    else
        log_warn "jq not available, component status tracking limited" "recovery"
    fi
}

# Check Go environment health
check_go_health() {
    log_info "Checking Go environment health" "go"
    
    local health_status="healthy"
    local issues=()
    
    # Check Go installation
    if ! timeout "$GO_HEALTH_TIMEOUT" go version >/dev/null 2>&1; then
        health_status="unhealthy"
        issues+=("go_not_installed")
    fi
    
    # Check Go environment variables
    local required_vars=("GOPATH" "GOPROXY" "GOSUMDB")
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var:-}" ]]; then
            issues+=("missing_${var,,}")
        fi
    done
    
    # Check if we can list packages
    if ! timeout 30 go list ./... >/dev/null 2>&1; then
        health_status="degraded"
        issues+=("package_listing_failed")
    fi
    
    # Check module status
    if [[ -f "go.mod" ]]; then
        if ! go mod verify >/dev/null 2>&1; then
            health_status="degraded"
            issues+=("module_verification_failed")
        fi
    else
        issues+=("no_go_mod")
    fi
    
    update_component_status "go" "$health_status" $([[ "$health_status" != "healthy" ]] && echo 1 || echo 0)
    
    if [[ ${#issues[@]} -gt 0 ]]; then
        log_warn "Go health issues detected: ${issues[*]}" "go"
    else
        log_info "Go environment is healthy" "go"
    fi
    
    [[ "$health_status" != "unhealthy" ]]
}

# Check Docker environment health
check_docker_health() {
    log_info "Checking Docker environment health" "docker"
    
    local health_status="healthy"
    local issues=()
    
    # Check Docker installation
    if ! command -v docker >/dev/null 2>&1; then
        health_status="unhealthy"
        issues+=("docker_not_installed")
        update_component_status "docker" "$health_status" 1
        log_error "Docker is not installed" "docker"
        return 1
    fi
    
    # Check Docker daemon
    if ! timeout "$DOCKER_HEALTH_TIMEOUT" docker info >/dev/null 2>&1; then
        health_status="unhealthy"
        issues+=("docker_daemon_unreachable")
    fi
    
    # Check Docker Buildx
    if ! docker buildx version >/dev/null 2>&1; then
        health_status="degraded"
        issues+=("buildx_unavailable")
    fi
    
    # Test basic Docker functionality
    if ! echo "FROM alpine:latest" | timeout 60 docker build -q -t health-test - >/dev/null 2>&1; then
        health_status="degraded"
        issues+=("docker_build_failed")
    else
        docker rmi health-test >/dev/null 2>&1 || true
    fi
    
    update_component_status "docker" "$health_status" $([[ "$health_status" == "unhealthy" ]] && echo 1 || echo 0)
    
    if [[ ${#issues[@]} -gt 0 ]]; then
        log_warn "Docker health issues detected: ${issues[*]}" "docker"
    else
        log_info "Docker environment is healthy" "docker"
    fi
    
    [[ "$health_status" != "unhealthy" ]]
}

# Check registry health
check_registry_health() {
    local registry_host="${1:-${REGISTRY_HOST:-localhost:5100}}"
    
    log_info "Checking registry health: $registry_host" "registry"
    
    if [[ -z "$registry_host" ]] || [[ "$registry_host" == "offline" ]]; then
        update_component_status "registry" "offline" 0
        log_info "Registry is in offline mode" "registry"
        return 0
    fi
    
    local health_status="healthy"
    local health_url
    
    # Determine health check URL
    if [[ "$registry_host" == "localhost"* ]] || [[ "$registry_host" == "127.0.0.1"* ]]; then
        health_url="http://$registry_host/v2/"
    else
        health_url="https://$registry_host/v2/"
    fi
    
    # Perform health check with retry
    local attempt=1
    local max_attempts=3
    
    while [[ $attempt -le $max_attempts ]]; do
        if timeout "$REGISTRY_HEALTH_TIMEOUT" curl -sf "$health_url" >/dev/null 2>&1; then
            update_component_status "registry" "healthy" 0
            log_info "Registry is healthy: $registry_host" "registry"
            return 0
        fi
        
        log_debug "Registry health check attempt $attempt failed" "registry"
        attempt=$((attempt + 1))
        
        if [[ $attempt -le $max_attempts ]]; then
            sleep 5
        fi
    done
    
    update_component_status "registry" "unhealthy" 1
    log_error "Registry health check failed after $max_attempts attempts: $registry_host" "registry"
    return 1
}

# Check network connectivity
check_network_health() {
    log_info "Checking network connectivity" "network"
    
    local health_status="healthy"
    local issues=()
    
    # Test basic internet connectivity
    local test_hosts=("8.8.8.8" "1.1.1.1" "google.com")
    local connectivity_ok=false
    
    for host in "${test_hosts[@]}"; do
        if timeout "$NETWORK_HEALTH_TIMEOUT" ping -c 3 "$host" >/dev/null 2>&1; then
            connectivity_ok=true
            break
        fi
    done
    
    if [[ "$connectivity_ok" != "true" ]]; then
        health_status="unhealthy"
        issues+=("no_internet_connectivity")
    fi
    
    # Test DNS resolution
    if ! timeout 10 nslookup google.com >/dev/null 2>&1; then
        health_status="degraded"
        issues+=("dns_resolution_failed")
    fi
    
    # Test proxy connectivity if configured
    if [[ -n "${GOPROXY:-}" ]] && [[ "$GOPROXY" != "direct" ]] && [[ "$GOPROXY" != "off" ]]; then
        local proxy_host
        proxy_host=$(echo "$GOPROXY" | cut -d',' -f1 | sed 's|https\?://||' | cut -d'/' -f1)
        
        if ! timeout 15 curl -sf "https://$proxy_host" >/dev/null 2>&1; then
            health_status="degraded"
            issues+=("proxy_unreachable")
        fi
    fi
    
    update_component_status "network" "$health_status" $([[ "$health_status" == "unhealthy" ]] && echo 1 || echo 0)
    
    if [[ ${#issues[@]} -gt 0 ]]; then
        log_warn "Network health issues detected: ${issues[*]}" "network"
    else
        log_info "Network connectivity is healthy" "network"
    fi
    
    [[ "$health_status" != "unhealthy" ]]
}

# Comprehensive health check
perform_comprehensive_health_check() {
    log_info "Performing comprehensive pipeline health check" "recovery"
    
    local overall_health="healthy"
    local component_results=()
    
    # Check all components
    if check_go_health; then
        component_results+=("go:healthy")
    else
        component_results+=("go:unhealthy")
        overall_health="degraded"
    fi
    
    if check_docker_health; then
        component_results+=("docker:healthy")
    else
        component_results+=("docker:unhealthy")
        overall_health="unhealthy"
    fi
    
    if check_registry_health; then
        component_results+=("registry:healthy")
    else
        component_results+=("registry:unhealthy")
        # Registry failure doesn't make overall health unhealthy
        if [[ "$overall_health" == "healthy" ]]; then
            overall_health="degraded"
        fi
    fi
    
    if check_network_health; then
        component_results+=("network:healthy")
    else
        component_results+=("network:unhealthy")
        overall_health="unhealthy"
    fi
    
    log_info "Health check results: ${component_results[*]}" "recovery"
    log_info "Overall pipeline health: $overall_health" "recovery"
    
    [[ "$overall_health" != "unhealthy" ]]
}

# ==============================================================================
# COMPONENT RECOVERY FUNCTIONS
# ==============================================================================

# Recover Go environment
recover_go_environment() {
    log_recovery "Attempting Go environment recovery" "go"
    
    # Clear Go module cache
    log_info "Clearing Go module cache" "go"
    go clean -modcache 2>/dev/null || true
    
    # Reset Go environment variables
    export GO111MODULE=on
    export GOFLAGS=-mod=mod
    export GOPROXY=https://proxy.golang.org,direct
    export GOSUMDB=sum.golang.org
    export CGO_ENABLED=1
    
    # Recreate cache directories
    mkdir -p "${HOME}/.cache/go-build"
    mkdir -p "${HOME}/go/pkg/mod"
    
    # Attempt module download with fallback proxies
    local proxies=("https://proxy.golang.org,direct" "https://goproxy.cn,direct" "direct")
    
    for proxy in "${proxies[@]}"; do
        log_info "Trying Go proxy: $proxy" "go"
        export GOPROXY="$proxy"
        
        if timeout 300 go mod download 2>/dev/null; then
            log_recovery "Go module download successful with proxy: $proxy" "go"
            return 0
        fi
    done
    
    log_error "Go environment recovery failed" "go"
    return 1
}

# Recover Docker environment
recover_docker_environment() {
    log_recovery "Attempting Docker environment recovery" "docker"
    
    # Check if Docker daemon is running
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker daemon is not running - cannot recover" "docker"
        return 1
    fi
    
    # Clean up Docker resources
    log_info "Cleaning Docker resources" "docker"
    docker system prune -f 2>/dev/null || true
    
    # Reset Docker Buildx
    log_info "Resetting Docker Buildx" "docker"
    docker buildx rm --all-inactive 2>/dev/null || true
    
    # Recreate Buildx builder
    if docker buildx create --name recovery-builder --use --driver docker-container 2>/dev/null; then
        log_recovery "Docker Buildx recovery successful" "docker"
    else
        log_warn "Docker Buildx recovery failed, using default builder" "docker"
        docker buildx use default 2>/dev/null || true
    fi
    
    # Test Docker functionality
    if echo "FROM alpine:latest" | docker build -q -t recovery-test - >/dev/null 2>&1; then
        docker rmi recovery-test >/dev/null 2>&1 || true
        log_recovery "Docker environment recovery successful" "docker"
        return 0
    else
        log_error "Docker environment recovery failed" "docker"
        return 1
    fi
}

# Recover network connectivity
recover_network_connectivity() {
    log_recovery "Attempting network connectivity recovery" "network"
    
    # Flush DNS cache (Linux)
    if command -v systemd-resolve >/dev/null 2>&1; then
        sudo systemd-resolve --flush-caches 2>/dev/null || true
    fi
    
    # Reset network interfaces (if in container with capabilities)
    if [[ -f /.dockerenv ]]; then
        log_info "Running in container, skipping network interface reset" "network"
    else
        # This might not work in GitHub Actions, but worth trying
        sudo ip route flush cache 2>/dev/null || true
    fi
    
    # Test connectivity recovery
    sleep 5
    if timeout 30 ping -c 3 8.8.8.8 >/dev/null 2>&1; then
        log_recovery "Network connectivity recovery successful" "network"
        return 0
    else
        log_error "Network connectivity recovery failed" "network"
        return 1
    fi
}

# ==============================================================================
# PIPELINE RECOVERY ORCHESTRATION
# ==============================================================================

# Execute component recovery
execute_component_recovery() {
    local component="$1"
    local recovery_attempt="$2"
    
    log_recovery "Starting recovery for component: $component (attempt $recovery_attempt)" "recovery"
    
    case "$component" in
        "go")
            recover_go_environment
            ;;
        "docker")
            recover_docker_environment
            ;;
        "network")
            recover_network_connectivity
            ;;
        "registry")
            # Registry recovery is usually handled by fallback mechanisms
            log_info "Registry recovery handled by fallback mechanisms" "registry"
            return 0
            ;;
        *)
            log_error "Unknown component for recovery: $component" "recovery"
            return 1
            ;;
    esac
}

# Main pipeline recovery function
execute_pipeline_recovery() {
    local failed_components=("$@")
    
    if [[ ${#failed_components[@]} -eq 0 ]]; then
        log_info "No components require recovery" "recovery"
        return 0
    fi
    
    log_recovery "Starting pipeline recovery for components: ${failed_components[*]}" "recovery"
    
    # Update recovery attempt count
    local current_attempts
    current_attempts=$(jq -r '.recovery_attempts // 0' "$PIPELINE_STATE_FILE" 2>/dev/null || echo "0")
    current_attempts=$((current_attempts + 1))
    
    if [[ $current_attempts -gt $MAX_RECOVERY_ATTEMPTS ]]; then
        log_error "Maximum recovery attempts ($MAX_RECOVERY_ATTEMPTS) exceeded" "recovery"
        return 1
    fi
    
    # Update pipeline state
    if command -v jq >/dev/null 2>&1; then
        local temp_file
        temp_file=$(mktemp)
        jq --argjson attempts "$current_attempts" '.recovery_attempts = $attempts' "$PIPELINE_STATE_FILE" > "$temp_file"
        mv "$temp_file" "$PIPELINE_STATE_FILE"
    fi
    
    # Record recovery attempt
    echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ") - Recovery attempt $current_attempts for: ${failed_components[*]}" >> "$RECOVERY_HISTORY_FILE"
    
    local recovery_success=true
    
    # Attempt recovery for each failed component
    for component in "${failed_components[@]}"; do
        if ! execute_component_recovery "$component" "$current_attempts"; then
            log_error "Recovery failed for component: $component" "recovery"
            recovery_success=false
        else
            # Wait before checking health
            sleep "$COMPONENT_RESTART_DELAY"
            
            # Verify recovery
            case "$component" in
                "go")
                    if ! check_go_health; then
                        log_error "Go health check failed after recovery" "recovery"
                        recovery_success=false
                    fi
                    ;;
                "docker")
                    if ! check_docker_health; then
                        log_error "Docker health check failed after recovery" "recovery"
                        recovery_success=false
                    fi
                    ;;
                "network")
                    if ! check_network_health; then
                        log_error "Network health check failed after recovery" "recovery"
                        recovery_success=false
                    fi
                    ;;
            esac
        fi
    done
    
    if [[ "$recovery_success" == "true" ]]; then
        log_recovery "Pipeline recovery completed successfully" "recovery"
        return 0
    else
        log_error "Pipeline recovery failed for some components" "recovery"
        return 1
    fi
}

# ==============================================================================
# DIAGNOSTICS AND REPORTING
# ==============================================================================

# Generate comprehensive diagnostics report
generate_diagnostics_report() {
    log_info "Generating comprehensive diagnostics report" "diagnostics"
    
    local report_file="${RECOVERY_STATE_DIR}/diagnostics-report.json"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Collect system information
    local system_info
    system_info=$(cat << EOF
{
  "timestamp": "$timestamp",
  "system": {
    "os": "$(uname -s)",
    "arch": "$(uname -m)",
    "kernel": "$(uname -r)",
    "hostname": "$(hostname)"
  },
  "environment": {
    "github": {
      "workflow": "${GITHUB_WORKFLOW:-unknown}",
      "run_id": "${GITHUB_RUN_ID:-unknown}",
      "repository": "${GITHUB_REPOSITORY:-unknown}",
      "ref": "${GITHUB_REF:-unknown}",
      "actor": "${GITHUB_ACTOR:-unknown}"
    },
    "runner": {
      "os": "${RUNNER_OS:-unknown}",
      "arch": "${RUNNER_ARCH:-unknown}",
      "temp": "${RUNNER_TEMP:-unknown}"
    }
  }
}
EOF
)
    
    # Collect Go environment if available
    local go_info="{}"
    if command -v go >/dev/null 2>&1; then
        go_info=$(cat << EOF
{
  "version": "$(go version 2>/dev/null || echo 'unknown')",
  "env": {
    "GOVERSION": "$(go env GOVERSION 2>/dev/null || echo 'unknown')",
    "GOOS": "$(go env GOOS 2>/dev/null || echo 'unknown')",
    "GOARCH": "$(go env GOARCH 2>/dev/null || echo 'unknown')",
    "GOROOT": "$(go env GOROOT 2>/dev/null || echo 'unknown')",
    "GOPATH": "$(go env GOPATH 2>/dev/null || echo 'unknown')",
    "GOPROXY": "$(go env GOPROXY 2>/dev/null || echo 'unknown')",
    "GOSUMDB": "$(go env GOSUMDB 2>/dev/null || echo 'unknown')"
  }
}
EOF
)
    fi
    
    # Collect Docker information if available
    local docker_info="{}"
    if command -v docker >/dev/null 2>&1; then
        local docker_version
        docker_version=$(docker --version 2>/dev/null | head -1 || echo "unknown")
        
        local buildx_version="unknown"
        if docker buildx version >/dev/null 2>&1; then
            buildx_version=$(docker buildx version 2>/dev/null | head -1 || echo "unknown")
        fi
        
        docker_info=$(cat << EOF
{
  "version": "$docker_version",
  "buildx_version": "$buildx_version",
  "daemon_accessible": $(docker info >/dev/null 2>&1 && echo "true" || echo "false")
}
EOF
)
    fi
    
    # Collect network information
    local network_info
    network_info=$(cat << EOF
{
  "interfaces": $(ip addr show 2>/dev/null | grep -E "^[0-9]+:" | awk '{print $2}' | sed 's/:$//' | jq -R . | jq -s . 2>/dev/null || echo '[]'),
  "dns_servers": $(cat /etc/resolv.conf 2>/dev/null | grep nameserver | awk '{print $2}' | jq -R . | jq -s . 2>/dev/null || echo '[]'),
  "default_route": "$(ip route show default 2>/dev/null | head -1 || echo 'unknown')"
}
EOF
)
    
    # Combine all information
    local full_report
    if command -v jq >/dev/null 2>&1; then
        full_report=$(echo "$system_info" | jq --argjson go "$go_info" --argjson docker "$docker_info" --argjson network "$network_info" \
            '. + {"go": $go, "docker": $docker, "network": $network}')
    else
        # Fallback without jq
        full_report="$system_info"
    fi
    
    echo "$full_report" > "$report_file"
    
    log_info "Diagnostics report generated: $report_file" "diagnostics"
    
    # Also generate a human-readable summary
    local summary_file="${RECOVERY_STATE_DIR}/diagnostics-summary.txt"
    cat > "$summary_file" << EOF
Pipeline Diagnostics Summary
===========================
Generated: $timestamp

System Information:
- OS: $(uname -s) $(uname -r)
- Architecture: $(uname -m)
- Hostname: $(hostname)

GitHub Environment:
- Workflow: ${GITHUB_WORKFLOW:-unknown}
- Run ID: ${GITHUB_RUN_ID:-unknown}
- Repository: ${GITHUB_REPOSITORY:-unknown}
- Ref: ${GITHUB_REF:-unknown}

Go Environment:
- Version: $(go version 2>/dev/null || echo 'Not available')
- GOPROXY: ${GOPROXY:-not set}
- GOSUMDB: ${GOSUMDB:-not set}

Docker Environment:
- Version: $(docker --version 2>/dev/null || echo 'Not available')
- Daemon: $(docker info >/dev/null 2>&1 && echo 'Accessible' || echo 'Not accessible')
- Buildx: $(docker buildx version >/dev/null 2>&1 && echo 'Available' || echo 'Not available')

Network Information:
- Default Route: $(ip route show default 2>/dev/null | head -1 || echo 'unknown')
- DNS Servers: $(cat /etc/resolv.conf 2>/dev/null | grep nameserver | awk '{print $2}' | tr '\n' ' ' || echo 'unknown')

Component Status:
$(if [[ -f "$COMPONENT_STATUS_FILE" ]]; then
    if command -v jq >/dev/null 2>&1; then
        jq -r 'to_entries[] | "- \(.key): \(.value.status) (last check: \(.value.last_check), failures: \(.value.failures))"' "$COMPONENT_STATUS_FILE"
    else
        echo "- Status file exists but jq not available for parsing"
    fi
else
    echo "- No component status available"
fi)

Recent Recovery History:
$(if [[ -f "$RECOVERY_HISTORY_FILE" ]]; then
    tail -10 "$RECOVERY_HISTORY_FILE" | sed 's/^/- /'
else
    echo "- No recovery history available"
fi)
EOF
    
    log_info "Diagnostics summary generated: $summary_file" "diagnostics"
}

# Generate GitHub Actions step summary
generate_github_summary() {
    if [[ -z "${GITHUB_STEP_SUMMARY:-}" ]]; then
        log_warn "GITHUB_STEP_SUMMARY not available, skipping GitHub summary" "reporting"
        return
    fi
    
    log_info "Generating GitHub Actions step summary" "reporting"
    
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S UTC")
    
    cat >> "$GITHUB_STEP_SUMMARY" << EOF
## Pipeline Recovery Report

**Generated:** $timestamp  
**Pipeline:** ${GITHUB_WORKFLOW:-unknown} #${GITHUB_RUN_ID:-unknown}  
**Repository:** ${GITHUB_REPOSITORY:-unknown}  

### Component Health Status

EOF
    
    # Add component status table
    if [[ -f "$COMPONENT_STATUS_FILE" ]] && command -v jq >/dev/null 2>&1; then
        cat >> "$GITHUB_STEP_SUMMARY" << EOF
| Component | Status | Last Check | Failures |
|-----------|--------|------------|----------|
EOF
        
        jq -r 'to_entries[] | "| \(.key) | \(.value.status) | \(.value.last_check) | \(.value.failures) |"' "$COMPONENT_STATUS_FILE" >> "$GITHUB_STEP_SUMMARY"
    else
        echo "Component status information not available." >> "$GITHUB_STEP_SUMMARY"
    fi
    
    # Add recovery history if available
    if [[ -f "$RECOVERY_HISTORY_FILE" ]] && [[ -s "$RECOVERY_HISTORY_FILE" ]]; then
        cat >> "$GITHUB_STEP_SUMMARY" << EOF

### Recent Recovery Activities

EOF
        tail -5 "$RECOVERY_HISTORY_FILE" | while read -r line; do
            echo "- $line" >> "$GITHUB_STEP_SUMMARY"
        done
    fi
    
    # Add diagnostics information
    cat >> "$GITHUB_STEP_SUMMARY" << EOF

### System Information

- **OS:** $(uname -s) $(uname -r)
- **Architecture:** $(uname -m)
- **Go Version:** $(go version 2>/dev/null | cut -d' ' -f3 || echo 'Not available')
- **Docker Version:** $(docker --version 2>/dev/null | cut -d' ' -f3 | sed 's/,//' || echo 'Not available')

### Recovery Status

Pipeline recovery mechanisms are active and monitoring component health.
EOF
    
    log_info "GitHub Actions step summary updated" "reporting"
}

# ==============================================================================
# MAIN FUNCTIONS AND CLI INTERFACE
# ==============================================================================

# Show usage information
show_usage() {
    cat << EOF
Pipeline Error Recovery Script

Usage: $0 <command> [options]

Commands:
  health-check                    Perform comprehensive health check
  recover <component>             Recover specific component (go, docker, network)
  auto-recover                    Automatically recover failed components
  diagnostics                     Generate diagnostics report
  status                          Show current pipeline status
  init                           Initialize recovery environment
  cleanup                         Clean up recovery state files

Component Health Commands:
  check-go                        Check Go environment health
  check-docker                    Check Docker environment health
  check-registry [host]           Check registry health
  check-network                   Check network connectivity

Recovery Commands:
  recover-go                      Recover Go environment
  recover-docker                  Recover Docker environment
  recover-network                 Recover network connectivity

Reporting Commands:
  generate-report                 Generate comprehensive diagnostics report
  github-summary                  Generate GitHub Actions step summary

Options:
  --max-attempts <n>              Maximum recovery attempts (default: $MAX_RECOVERY_ATTEMPTS)
  --timeout <s>                   Recovery timeout in seconds (default: $RECOVERY_TIMEOUT)
  --help                          Show this help message

Examples:
  $0 health-check
  $0 recover go
  $0 auto-recover
  $0 diagnostics
  $0 check-registry localhost:5100

Environment Variables:
  RECOVERY_STATE_DIR              Directory for recovery state files
  MAX_RECOVERY_ATTEMPTS           Maximum number of recovery attempts
  RECOVERY_TIMEOUT                Timeout for recovery operations
EOF
}

# Main function
main() {
    local command="${1:-}"
    
    if [[ $# -eq 0 || "$command" == "--help" || "$command" == "-h" ]]; then
        show_usage
        exit 0
    fi
    
    # Initialize recovery environment
    init_recovery_environment
    
    case "$command" in
        "health-check")
            perform_comprehensive_health_check
            ;;
        "check-go")
            check_go_health
            ;;
        "check-docker")
            check_docker_health
            ;;
        "check-registry")
            check_registry_health "${2:-}"
            ;;
        "check-network")
            check_network_health
            ;;
        "recover")
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 recover <component>" "recovery"
                exit 1
            fi
            execute_component_recovery "$2" 1
            ;;
        "recover-go")
            recover_go_environment
            ;;
        "recover-docker")
            recover_docker_environment
            ;;
        "recover-network")
            recover_network_connectivity
            ;;
        "auto-recover")
            # Perform health check and recover failed components
            failed_components=()
            
            if ! check_go_health; then
                failed_components+=("go")
            fi
            
            if ! check_docker_health; then
                failed_components+=("docker")
            fi
            
            if ! check_network_health; then
                failed_components+=("network")
            fi
            
            if [[ ${#failed_components[@]} -gt 0 ]]; then
                execute_pipeline_recovery "${failed_components[@]}"
            else
                log_info "All components are healthy, no recovery needed" "recovery"
            fi
            ;;
        "diagnostics"|"generate-report")
            generate_diagnostics_report
            ;;
        "github-summary")
            generate_github_summary
            ;;
        "status")
            echo "Pipeline Recovery Status:"
            echo "========================"
            if [[ -f "$PIPELINE_STATE_FILE" ]]; then
                if command -v jq >/dev/null 2>&1; then
                    jq -r '"Pipeline ID: " + .pipeline_id + "\nWorkflow: " + .workflow + "\nRepository: " + .repository + "\nRecovery Attempts: " + (.recovery_attempts | tostring)' "$PIPELINE_STATE_FILE"
                else
                    echo "Pipeline state file exists but jq not available for parsing"
                fi
            else
                echo "No pipeline state available"
            fi
            echo ""
            echo "Component Status:"
            if [[ -f "$COMPONENT_STATUS_FILE" ]]; then
                if command -v jq >/dev/null 2>&1; then
                    jq -r 'to_entries[] | "  " + .key + ": " + .value.status + " (failures: " + (.value.failures | tostring) + ")"' "$COMPONENT_STATUS_FILE"
                else
                    echo "  Component status file exists but jq not available for parsing"
                fi
            else
                echo "  No component status available"
            fi
            ;;
        "init")
            log_info "Recovery environment initialized" "recovery"
            ;;
        "cleanup")
            if [[ -d "$RECOVERY_STATE_DIR" ]]; then
                rm -rf "$RECOVERY_STATE_DIR"
                log_info "Recovery state files cleaned up" "recovery"
            else
                log_info "No recovery state files to clean up" "recovery"
            fi
            ;;
        *)
            log_error "Unknown command: $command" "recovery"
            show_usage
            exit 1
            ;;
    esac
}

# Trap signals for cleanup
trap 'log_info "Pipeline recovery script interrupted" "recovery"' INT TERM

# Execute main function if script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi