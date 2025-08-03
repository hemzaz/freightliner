#!/bin/bash
# CI/CD Reliability Enhancement System
# Provides automated error detection, recovery, and pipeline reliability improvements

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
RELIABILITY_LOG="${PROJECT_ROOT}/ci-reliability.log"
FAILURE_PATTERNS_FILE="${PROJECT_ROOT}/.github/failure-patterns.json"
RECOVERY_ACTIONS_DIR="${PROJECT_ROOT}/.github/recovery-actions"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${timestamp} [${level}] ${message}" | tee -a "${RELIABILITY_LOG}"
}

info() { log "INFO" "$@"; }
warn() { log "WARN" "${YELLOW}$*${NC}"; }
error() { log "ERROR" "${RED}$*${NC}"; }
success() { log "SUCCESS" "${GREEN}$*${NC}"; }
debug() { log "DEBUG" "${PURPLE}$*${NC}"; }

# Initialize reliability system
init_reliability_system() {
    info "Initializing CI/CD reliability system..."
    
    mkdir -p "${RECOVERY_ACTIONS_DIR}"
    
    # Create failure patterns database
    if [[ ! -f "${FAILURE_PATTERNS_FILE}" ]]; then
        cat > "${FAILURE_PATTERNS_FILE}" << 'EOF'
{
  "patterns": {
    "golangci_lint_schema": {
      "pattern": "unknown configuration key.*colored-line-number",
      "category": "configuration",
      "severity": "high",
      "description": "golangci-lint deprecated configuration option",
      "auto_fix": true,
      "recovery_action": "fix_golangci_config"
    },
    "docker_build_failure": {
      "pattern": "(docker: command not found|Cannot connect to the Docker daemon)",
      "category": "infrastructure",
      "severity": "high",
      "description": "Docker daemon or installation issues",
      "auto_fix": false,
      "recovery_action": "setup_docker"
    },
    "go_mod_tidy_required": {
      "pattern": "go.mod file indicates go .* but maximum supported version is",
      "category": "dependency",
      "severity": "medium",
      "description": "Go module version mismatch",
      "auto_fix": true,
      "recovery_action": "fix_go_mod"
    },
    "test_timeout": {
      "pattern": "test timed out after .*",
      "category": "performance",
      "severity": "medium",
      "description": "Test execution timeout",
      "auto_fix": true,
      "recovery_action": "optimize_test_timeouts"
    },
    "registry_connection_failure": {
      "pattern": "connection refused.*registry",
      "category": "network",
      "severity": "high",
      "description": "Container registry connection failure",
      "auto_fix": true,
      "recovery_action": "fix_registry_connection"
    },
    "memory_exhaustion": {
      "pattern": "(out of memory|cannot allocate memory)",
      "category": "resource",
      "severity": "high",
      "description": "Memory exhaustion during build/test",
      "auto_fix": true,
      "recovery_action": "optimize_memory_usage"
    },
    "disk_space_full": {
      "pattern": "no space left on device",
      "category": "resource",
      "severity": "critical",
      "description": "Disk space exhaustion",
      "auto_fix": true,
      "recovery_action": "cleanup_disk_space"
    }
  },
  "recovery_strategies": {
    "retry_with_backoff": {
      "max_attempts": 3,
      "backoff_multiplier": 2,
      "initial_delay": 5
    },
    "resource_optimization": {
      "memory_limit": "4Gi",
      "cpu_limit": "2",
      "timeout_multiplier": 1.5
    }
  }
}
EOF
        info "Created failure patterns database"
    fi
    
    success "Reliability system initialized"
}

# Analyze logs for failure patterns
analyze_failure_patterns() {
    local log_file="$1"
    local detected_patterns=()
    
    info "Analyzing failure patterns in: ${log_file}"
    
    if [[ ! -f "${log_file}" ]]; then
        warn "Log file not found: ${log_file}"
        return 1
    fi
    
    # Parse failure patterns from JSON
    local patterns
    if ! patterns=$(jq -r '.patterns | to_entries[] | "\(.key)|\(.value.pattern)|\(.value.category)|\(.value.severity)|\(.value.auto_fix)|\(.value.recovery_action)"' "${FAILURE_PATTERNS_FILE}" 2>/dev/null); then
        error "Failed to parse failure patterns"
        return 1
    fi
    
    # Check each pattern against the log
    while IFS='|' read -r name pattern category severity auto_fix recovery_action; do
        if grep -qE "${pattern}" "${log_file}"; then
            detected_patterns+=("${name}")
            warn "Detected failure pattern: ${name} (${category}/${severity})"
            
            if [[ "${auto_fix}" == "true" ]]; then
                info "Auto-fix available for ${name}: ${recovery_action}"
                execute_recovery_action "${recovery_action}"
            else
                warn "Manual intervention required for ${name}"
            fi
        fi
    done <<< "${patterns}"
    
    if [[ ${#detected_patterns[@]} -eq 0 ]]; then
        success "No known failure patterns detected"
        return 0
    else
        error "Detected ${#detected_patterns[@]} failure patterns: ${detected_patterns[*]}"
        return 1
    fi
}

# Execute recovery action
execute_recovery_action() {
    local action="$1"
    
    info "Executing recovery action: ${action}"
    
    case "${action}" in
        "fix_golangci_config")
            fix_golangci_config
            ;;
        "setup_docker")
            setup_docker_environment
            ;;
        "fix_go_mod")
            fix_go_mod_issues
            ;;
        "optimize_test_timeouts")
            optimize_test_timeouts
            ;;
        "fix_registry_connection")
            fix_registry_connection
            ;;
        "optimize_memory_usage")
            optimize_memory_usage
            ;;
        "cleanup_disk_space")
            cleanup_disk_space
            ;;
        *)
            warn "Unknown recovery action: ${action}"
            return 1
            ;;
    esac
}

# Fix golangci-lint configuration
fix_golangci_config() {
    info "Fixing golangci-lint configuration..."
    
    local config_file="${PROJECT_ROOT}/.golangci.yml"
    
    if [[ ! -f "${config_file}" ]]; then
        error "golangci-lint config file not found"
        return 1
    fi
    
    # Backup original config
    cp "${config_file}" "${config_file}.backup.$(date +%s)"
    
    # Check if colored-line-number is present
    if grep -q "colored-line-number:" "${config_file}"; then
        info "Updating deprecated colored-line-number configuration..."
        
        # Replace deprecated option with new format
        sed -i.tmp 's/colored-line-number: true/formats:\
    - format: colored-line-number\
      path: stdout/' "${config_file}"
        
        rm -f "${config_file}.tmp"
        success "Fixed golangci-lint configuration"
    else
        info "No deprecated options found in golangci-lint config"
    fi
    
    # Validate configuration
    if command -v golangci-lint &> /dev/null; then
        if golangci-lint config verify; then
            success "golangci-lint configuration validated successfully"
            return 0
        else
            error "Configuration validation failed"
            # Restore backup
            mv "${config_file}.backup."* "${config_file}"
            return 1
        fi
    else
        warn "golangci-lint not available for validation"
        return 0
    fi
}

# Setup Docker environment
setup_docker_environment() {
    info "Setting up Docker environment..."
    
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed. Please install Docker first."
        return 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        warn "Docker daemon is not running"
        
        # Try to start Docker (works on some systems)
        if command -v systemctl &> /dev/null; then
            info "Attempting to start Docker daemon..."
            sudo systemctl start docker || warn "Failed to start Docker daemon"
        elif command -v service &> /dev/null; then
            info "Attempting to start Docker daemon..."
            sudo service docker start || warn "Failed to start Docker daemon"
        else
            error "Cannot start Docker daemon automatically. Please start it manually."
            return 1
        fi
        
        # Wait a moment and check again
        sleep 5
        if docker info &> /dev/null; then
            success "Docker daemon started successfully"
        else
            error "Docker daemon failed to start"
            return 1
        fi
    else
        success "Docker daemon is running"
    fi
    
    # Test Docker functionality
    if docker run --rm hello-world &> /dev/null; then
        success "Docker environment is working correctly"
        return 0
    else
        error "Docker environment test failed"
        return 1
    fi
}

# Fix Go module issues
fix_go_mod_issues() {
    info "Fixing Go module issues..."
    
    cd "${PROJECT_ROOT}"
    
    # Clean module cache
    go clean -modcache
    
    # Tidy modules
    if go mod tidy; then
        success "Go modules tidied successfully"
    else
        error "Failed to tidy Go modules"
        return 1
    fi
    
    # Download dependencies
    if go mod download; then
        success "Dependencies downloaded successfully"
    else
        error "Failed to download dependencies"
        return 1
    fi
    
    # Verify modules
    if go mod verify; then
        success "Go modules verified successfully"
        return 0
    else
        error "Go module verification failed"
        return 1
    fi
}

# Optimize test timeouts
optimize_test_timeouts() {
    info "Optimizing test timeouts..."
    
    cd "${PROJECT_ROOT}"
    
    # Find test files with long-running tests
    local test_files
    if test_files=$(find . -name "*_test.go" -type f); then
        info "Found test files: $(echo "${test_files}" | wc -l) files"
        
        # Check for existing timeout patterns
        local files_with_long_tests=()
        while IFS= read -r file; do
            if grep -q "time\.Minute\|time\.Hour\|30.*time\.Second" "${file}"; then
                files_with_long_tests+=("${file}")
            fi
        done <<< "${test_files}"
        
        if [[ ${#files_with_long_tests[@]} -gt 0 ]]; then
            info "Found ${#files_with_long_tests[@]} files with potentially long-running tests"
            for file in "${files_with_long_tests[@]}"; do
                info "  - ${file}"
            done
            
            # Suggest optimizations
            warn "Consider optimizing these tests:"
            warn "1. Use testing.Short() to skip long tests in CI"
            warn "2. Reduce test data sizes for CI environments"
            warn "3. Use context.WithTimeout for better control"
            warn "4. Mock external dependencies"
        else
            info "No obviously long-running tests found"
        fi
    else
        warn "No test files found"
    fi
    
    success "Test timeout analysis completed"
    return 0
}

# Fix registry connection issues
fix_registry_connection() {
    info "Fixing registry connection issues..."
    
    # Test common registries
    local registries=("docker.io" "registry-1.docker.io" "index.docker.io")
    local working_registries=()
    
    for registry in "${registries[@]}"; do
        info "Testing connection to ${registry}..."
        if timeout 10 curl -sSf "https://${registry}/v2/" &> /dev/null; then
            working_registries+=("${registry}")
            success "Connection to ${registry} successful"
        else
            warn "Connection to ${registry} failed"
        fi
    done
    
    if [[ ${#working_registries[@]} -gt 0 ]]; then
        success "Found ${#working_registries[@]} working registries"
        return 0
    else
        error "No registry connections available"
        return 1
    fi
}

# Optimize memory usage
optimize_memory_usage() {
    info "Optimizing memory usage..."
    
    # Check current memory usage
    if command -v free &> /dev/null; then
        local mem_info
        mem_info=$(free -h)
        info "Current memory usage:"
        echo "${mem_info}" | while IFS= read -r line; do
            info "  ${line}"
        done
    fi
    
    # Clean Go build cache
    info "Cleaning Go build cache..."
    go clean -cache
    go clean -testcache
    go clean -modcache
    
    # Clean Docker if available
    if command -v docker &> /dev/null && docker info &> /dev/null; then
        info "Cleaning Docker resources..."
        docker system prune -f || warn "Docker cleanup failed"
    fi
    
    # Set memory-optimized Go environment
    export GOMEMLIMIT=2GiB
    export GOGC=100
    
    success "Memory optimization completed"
    return 0
}

# Cleanup disk space
cleanup_disk_space() {
    info "Cleaning up disk space..."
    
    # Show current disk usage
    local disk_usage
    disk_usage=$(df -h . | tail -1)
    info "Current disk usage: ${disk_usage}"
    
    cd "${PROJECT_ROOT}"
    
    # Clean temporary files
    info "Cleaning temporary files..."
    find . -name "*.tmp" -type f -delete 2>/dev/null || true
    find . -name ".DS_Store" -type f -delete 2>/dev/null || true
    
    # Clean Go caches
    info "Cleaning Go caches..."
    go clean -cache -testcache -modcache
    
    # Clean build artifacts
    info "Cleaning build artifacts..."
    rm -rf coverage.out gosec-results.sarif *.prof 2>/dev/null || true
    
    # Clean Docker if available
    if command -v docker &> /dev/null && docker info &> /dev/null; then
        info "Cleaning Docker resources..."
        docker system prune -af --volumes || warn "Docker cleanup failed"
    fi
    
    # Show updated disk usage
    local new_disk_usage
    new_disk_usage=$(df -h . | tail -1)
    info "Updated disk usage: ${new_disk_usage}"
    
    success "Disk cleanup completed"
    return 0
}

# Pre-flight checks
run_preflight_checks() {
    info "Running pre-flight reliability checks..."
    
    local issues=0
    
    # Check Go installation and version
    if ! command -v go &> /dev/null; then
        error "Go is not installed"
        ((issues++))
    else
        local go_version
        go_version=$(go version | grep -o 'go[0-9]*\.[0-9]*\.[0-9]*')
        info "Go version: ${go_version}"
        
        # Check if version matches CI expectations
        local expected_version="go1.24.5"
        if [[ "${go_version}" != "${expected_version}" ]]; then
            warn "Go version mismatch: expected ${expected_version}, got ${go_version}"
        fi
    fi
    
    # Check golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        error "golangci-lint is not installed"
        ((issues++))
    else
        local lint_version
        lint_version=$(golangci-lint --version | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*')
        info "golangci-lint version: ${lint_version}"
        
        # Validate configuration
        if [[ -f "${PROJECT_ROOT}/.golangci.yml" ]]; then
            if golangci-lint config verify; then
                success "golangci-lint configuration is valid"
            else
                error "golangci-lint configuration is invalid"
                ((issues++))
            fi
        else
            warn "No golangci-lint configuration found"
        fi
    fi
    
    # Check disk space
    local disk_usage
    disk_usage=$(df . | tail -1 | awk '{print $5}' | sed 's/%//')
    if [[ ${disk_usage} -gt 90 ]]; then
        error "Disk usage is critically high: ${disk_usage}%"
        ((issues++))
    elif [[ ${disk_usage} -gt 80 ]]; then
        warn "Disk usage is high: ${disk_usage}%"
    else
        info "Disk usage: ${disk_usage}%"
    fi
    
    # Check memory
    if command -v free &> /dev/null; then
        local mem_usage
        mem_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
        if (( $(echo "${mem_usage} > 90" | bc -l) )); then
            error "Memory usage is critically high: ${mem_usage}%"
            ((issues++))
        else
            info "Memory usage: ${mem_usage}%"
        fi
    fi
    
    # Check network connectivity
    if ! timeout 10 curl -sSf https://proxy.golang.org &> /dev/null; then
        error "Cannot reach Go module proxy"
        ((issues++))
    else
        success "Go module proxy is accessible"
    fi
    
    if [[ ${issues} -eq 0 ]]; then
        success "All pre-flight checks passed"
        return 0
    else
        error "Pre-flight checks failed with ${issues} issues"
        return 1
    fi
}

# Continuous monitoring
continuous_monitoring() {
    info "Starting continuous monitoring mode..."
    
    local monitor_interval=30  # seconds
    local failure_count=0
    local max_failures=5
    
    while true; do
        info "Running health check cycle..."
        
        if run_preflight_checks; then
            if [[ ${failure_count} -gt 0 ]]; then
                info "System recovered from previous failures"
                failure_count=0
            fi
        else
            ((failure_count++))
            warn "Health check failed (${failure_count}/${max_failures})"
            
            if [[ ${failure_count} -ge ${max_failures} ]]; then
                error "Maximum consecutive failures reached. Exiting continuous monitoring."
                break
            fi
        fi
        
        info "Sleeping for ${monitor_interval} seconds..."
        sleep ${monitor_interval}
    done
}

# Generate reliability report
generate_reliability_report() {
    info "Generating reliability report..."
    
    local report_file="${PROJECT_ROOT}/ci-reliability-report-$(date +%Y%m%d-%H%M%S).json"
    
    # Collect system information
    local go_version
    go_version=$(go version 2>/dev/null | grep -o 'go[0-9]*\.[0-9]*\.[0-9]*' || echo "unknown")
    
    local lint_version
    lint_version=$(golangci-lint --version 2>/dev/null | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*' || echo "unknown")
    
    local disk_usage
    disk_usage=$(df . | tail -1 | awk '{print $5}' | sed 's/%//')
    
    local mem_usage
    if command -v free &> /dev/null; then
        mem_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
    else
        mem_usage="unknown"
    fi
    
    cat > "${report_file}" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "reliability_assessment": {
    "overall_status": "$(run_preflight_checks &>/dev/null && echo "healthy" || echo "issues_detected")",
    "last_check": "$(date -Iseconds)"
  },
  "system_info": {
    "go_version": "${go_version}",
    "golangci_lint_version": "${lint_version}",
    "disk_usage_percent": ${disk_usage},
    "memory_usage_percent": "${mem_usage}",
    "docker_available": $(command -v docker &> /dev/null && echo "true" || echo "false"),
    "docker_running": $(docker info &> /dev/null && echo "true" || echo "false")
  },
  "configuration_health": {
    "golangci_config_valid": $(golangci-lint config verify &>/dev/null && echo "true" || echo "false"),
    "go_mod_valid": $(cd "${PROJECT_ROOT}" && go mod verify &>/dev/null && echo "true" || echo "false"),
    "dockerfile_exists": $([[ -f "${PROJECT_ROOT}/Dockerfile" ]] && echo "true" || echo "false")
  },
  "recommendations": [
    $(if [[ ${disk_usage} -gt 85 ]]; then echo '"Consider disk cleanup"'; fi)
    $(if ! command -v docker &> /dev/null; then echo '"Install Docker for full CI functionality"'; fi)
    $(if ! golangci-lint config verify &>/dev/null; then echo '"Fix golangci-lint configuration"'; fi)
  ]
}
EOF
    
    info "Reliability report generated: ${report_file}"
    
    # Display summary
    echo
    echo -e "${BLUE}=== CI/CD RELIABILITY REPORT ===${NC}"
    echo "Timestamp: $(date)"
    echo "Report: ${report_file}"
    echo
    
    if run_preflight_checks &>/dev/null; then
        echo -e "${GREEN}✓${NC} Overall Status: HEALTHY"
    else
        echo -e "${RED}✗${NC} Overall Status: ISSUES DETECTED"
    fi
    
    echo -e "${BLUE}System Information:${NC}"
    echo "  Go Version: ${go_version}"
    echo "  golangci-lint Version: ${lint_version}"
    echo "  Disk Usage: ${disk_usage}%"
    echo "  Memory Usage: ${mem_usage}%"
    echo
    
    return 0
}

# Main execution
main() {
    local command="${1:-check}"
    
    echo -e "${BLUE}=== FREIGHTLINER CI/CD RELIABILITY SYSTEM ===${NC}"
    echo "Command: ${command}"
    echo
    
    init_reliability_system
    
    case "${command}" in
        "check"|"preflight")
            run_preflight_checks
            ;;
        "analyze")
            local log_file="${2:-${RELIABILITY_LOG}}"
            analyze_failure_patterns "${log_file}"
            ;;
        "fix")
            local action="${2:-all}"
            if [[ "${action}" == "all" ]]; then
                execute_recovery_action "fix_golangci_config"
                execute_recovery_action "fix_go_mod"
                execute_recovery_action "optimize_memory_usage"
            else
                execute_recovery_action "${action}"
            fi
            ;;
        "monitor")
            continuous_monitoring
            ;;
        "report")
            generate_reliability_report
            ;;
        *)
            echo "Usage: $0 [check|analyze|fix|monitor|report]"
            echo
            echo "Commands:"
            echo "  check    - Run pre-flight reliability checks (default)"
            echo "  analyze  - Analyze logs for failure patterns"
            echo "  fix      - Execute recovery actions"
            echo "  monitor  - Continuous monitoring mode"
            echo "  report   - Generate reliability report"
            echo
            echo "Examples:"
            echo "  $0 check                    # Run health checks"
            echo "  $0 analyze build.log        # Analyze specific log file"
            echo "  $0 fix golangci_config      # Fix specific issue"
            echo "  $0 fix all                  # Fix all known issues"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"