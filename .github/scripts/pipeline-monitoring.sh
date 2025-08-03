#!/bin/bash
# Advanced CI/CD Pipeline Monitoring System
# Provides comprehensive monitoring, alerting, and diagnostics

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MONITORING_CONFIG="${PROJECT_ROOT}/.github/monitoring-config.yml"
LOG_FILE="${PROJECT_ROOT}/pipeline-monitoring.log"
METRICS_DIR="${PROJECT_ROOT}/.github/metrics"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${timestamp} [${level}] ${message}" | tee -a "${LOG_FILE}"
}

info() { log "INFO" "$@"; }
warn() { log "WARN" "${YELLOW}$*${NC}"; }
error() { log "ERROR" "${RED}$*${NC}"; }
success() { log "SUCCESS" "${GREEN}$*${NC}"; }

# Initialize monitoring directories
init_monitoring() {
    info "Initializing pipeline monitoring system..."
    
    mkdir -p "${METRICS_DIR}/build"
    mkdir -p "${METRICS_DIR}/test"
    mkdir -p "${METRICS_DIR}/lint"
    mkdir -p "${METRICS_DIR}/security"
    mkdir -p "${METRICS_DIR}/docker"
    
    # Create monitoring config if it doesn't exist
    if [[ ! -f "${MONITORING_CONFIG}" ]]; then
        cat > "${MONITORING_CONFIG}" << 'EOF'
monitoring:
  enabled: true
  alerts:
    enabled: true
    channels:
      - slack
      - email
  thresholds:
    build_timeout: 600  # 10 minutes
    test_timeout: 1800  # 30 minutes
    lint_timeout: 480   # 8 minutes
    security_timeout: 300  # 5 minutes
    docker_timeout: 900    # 15 minutes
  performance:
    baseline_build_time: 180  # 3 minutes
    baseline_test_time: 300   # 5 minutes
    performance_degradation_threshold: 50  # 50% increase
EOF
        info "Created monitoring configuration at ${MONITORING_CONFIG}"
    fi
    
    success "Monitoring system initialized"
}

# System health check
check_system_health() {
    info "Performing system health check..."
    
    local issues=0
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        error "Go is not installed or not in PATH"
        ((issues++))
    else
        local go_version=$(go version | grep -o 'go[0-9]*\.[0-9]*\.[0-9]*')
        info "Go version: ${go_version}"
    fi
    
    # Check golangci-lint installation
    if ! command -v golangci-lint &> /dev/null; then
        error "golangci-lint is not installed or not in PATH"
        ((issues++))
    else
        local lint_version=$(golangci-lint --version | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*')
        info "golangci-lint version: ${lint_version}"
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        warn "Docker is not installed or not in PATH"
    else
        if ! docker info &> /dev/null; then
            warn "Docker daemon is not running"
        else
            local docker_version=$(docker --version | grep -o '[0-9]*\.[0-9]*\.[0-9]*')
            info "Docker version: ${docker_version}"
        fi
    fi
    
    # Check disk space
    local disk_usage=$(df . | tail -1 | awk '{print $5}' | sed 's/%//')
    if [[ ${disk_usage} -gt 85 ]]; then
        error "Disk usage is ${disk_usage}% - consider cleanup"
        ((issues++))
    else
        info "Disk usage: ${disk_usage}%"
    fi
    
    # Check memory
    if command -v free &> /dev/null; then
        local mem_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
        info "Memory usage: ${mem_usage}%"
    fi
    
    if [[ ${issues} -eq 0 ]]; then
        success "System health check passed"
        return 0
    else
        error "System health check failed with ${issues} issues"
        return 1
    fi
}

# Monitor build stage
monitor_build() {
    info "Monitoring build stage..."
    local start_time=$(date +%s)
    
    # Run build with monitoring
    if go build -v ./...; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        # Record metrics
        echo "${duration}" > "${METRICS_DIR}/build/last_duration"
        echo "$(date -Iseconds)" > "${METRICS_DIR}/build/last_success"
        
        info "Build completed in ${duration} seconds"
        
        # Check performance regression
        if [[ -f "${METRICS_DIR}/build/baseline_duration" ]]; then
            local baseline=$(cat "${METRICS_DIR}/build/baseline_duration")
            local threshold=$((baseline * 150 / 100))  # 50% increase threshold
            
            if [[ ${duration} -gt ${threshold} ]]; then
                warn "Build performance regression detected: ${duration}s vs baseline ${baseline}s"
            fi
        else
            echo "${duration}" > "${METRICS_DIR}/build/baseline_duration"
            info "Established build baseline: ${duration}s"
        fi
        
        success "Build monitoring completed successfully"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo "$(date -Iseconds)" > "${METRICS_DIR}/build/last_failure"
        error "Build failed after ${duration} seconds"
        return 1
    fi
}

# Monitor test stage
monitor_tests() {
    info "Monitoring test stage..."
    local start_time=$(date +%s)
    
    # Run tests with coverage and race detection
    if go test -v -race -coverprofile=coverage.out -timeout=30m ./...; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        # Record metrics
        echo "${duration}" > "${METRICS_DIR}/test/last_duration"
        echo "$(date -Iseconds)" > "${METRICS_DIR}/test/last_success"
        
        # Extract coverage information
        if [[ -f "coverage.out" ]]; then
            local coverage=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
            echo "${coverage}" > "${METRICS_DIR}/test/last_coverage"
            info "Test coverage: ${coverage}%"
        fi
        
        info "Tests completed in ${duration} seconds"
        success "Test monitoring completed successfully"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo "$(date -Iseconds)" > "${METRICS_DIR}/test/last_failure"
        error "Tests failed after ${duration} seconds"
        return 1
    fi
}

# Monitor lint stage
monitor_lint() {
    info "Monitoring lint stage..."
    local start_time=$(date +%s)
    
    # Validate golangci-lint configuration first
    if ! golangci-lint config verify; then
        error "golangci-lint configuration is invalid"
        return 1
    fi
    
    # Run linting
    if golangci-lint run --timeout=8m; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo "${duration}" > "${METRICS_DIR}/lint/last_duration"
        echo "$(date -Iseconds)" > "${METRICS_DIR}/lint/last_success"
        
        info "Linting completed in ${duration} seconds"
        success "Lint monitoring completed successfully"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo "$(date -Iseconds)" > "${METRICS_DIR}/lint/last_failure"
        error "Linting failed after ${duration} seconds"
        return 1
    fi
}

# Monitor security stage
monitor_security() {
    info "Monitoring security stage..."
    local start_time=$(date +%s)
    
    # Install gosec if not available
    if ! command -v gosec &> /dev/null; then
        info "Installing gosec..."
        go install github.com/securego/gosec/v2/cmd/gosec@latest
    fi
    
    # Run security scan
    if gosec -no-fail -fmt sarif -out gosec-results.sarif ./...; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo "${duration}" > "${METRICS_DIR}/security/last_duration"
        echo "$(date -Iseconds)" > "${METRICS_DIR}/security/last_success"
        
        # Count security issues
        if [[ -f "gosec-results.sarif" ]]; then
            local issue_count=$(jq '.runs[0].results | length' gosec-results.sarif 2>/dev/null || echo "0")
            echo "${issue_count}" > "${METRICS_DIR}/security/last_issue_count"
            info "Security scan found ${issue_count} issues"
        fi
        
        info "Security scan completed in ${duration} seconds"
        success "Security monitoring completed successfully"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo "$(date -Iseconds)" > "${METRICS_DIR}/security/last_failure"
        error "Security scan failed after ${duration} seconds"
        return 1
    fi
}

# Monitor Docker stage
monitor_docker() {
    info "Monitoring Docker stage..."
    local start_time=$(date +%s)
    
    if ! command -v docker &> /dev/null; then
        warn "Docker not available, skipping Docker monitoring"
        return 0
    fi
    
    if ! docker info &> /dev/null; then
        error "Docker daemon not running"
        return 1
    fi
    
    # Build Docker image
    if docker build -t freightliner:test .; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo "${duration}" > "${METRICS_DIR}/docker/last_duration"
        echo "$(date -Iseconds)" > "${METRICS_DIR}/docker/last_success"
        
        # Test the built image
        if docker run --rm freightliner:test --version; then
            info "Docker image test passed"
        else
            warn "Docker image test failed"
        fi
        
        # Get image size
        local image_size=$(docker images freightliner:test --format "{{.Size}}")
        echo "${image_size}" > "${METRICS_DIR}/docker/last_size"
        info "Docker image size: ${image_size}"
        
        info "Docker build completed in ${duration} seconds"
        success "Docker monitoring completed successfully"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo "$(date -Iseconds)" > "${METRICS_DIR}/docker/last_failure"
        error "Docker build failed after ${duration} seconds"
        return 1
    fi
}

# Generate monitoring report
generate_report() {
    info "Generating monitoring report..."
    
    local report_file="${METRICS_DIR}/pipeline-report-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "${report_file}" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "system_health": {
    "go_version": "$(go version | grep -o 'go[0-9]*\.[0-9]*\.[0-9]*' || echo 'unknown')",
    "golangci_lint_version": "$(golangci-lint --version | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*' || echo 'unknown')",
    "docker_version": "$(docker --version | grep -o '[0-9]*\.[0-9]*\.[0-9]*' || echo 'unknown')"
  },
  "build": {
    "last_duration": $(cat "${METRICS_DIR}/build/last_duration" 2>/dev/null || echo 0),
    "last_success": "$(cat "${METRICS_DIR}/build/last_success" 2>/dev/null || echo 'never')",
    "baseline_duration": $(cat "${METRICS_DIR}/build/baseline_duration" 2>/dev/null || echo 0)
  },
  "test": {
    "last_duration": $(cat "${METRICS_DIR}/test/last_duration" 2>/dev/null || echo 0),
    "last_success": "$(cat "${METRICS_DIR}/test/last_success" 2>/dev/null || echo 'never')",
    "last_coverage": $(cat "${METRICS_DIR}/test/last_coverage" 2>/dev/null || echo 0)
  },
  "lint": {
    "last_duration": $(cat "${METRICS_DIR}/lint/last_duration" 2>/dev/null || echo 0),
    "last_success": "$(cat "${METRICS_DIR}/lint/last_success" 2>/dev/null || echo 'never')"
  },
  "security": {
    "last_duration": $(cat "${METRICS_DIR}/security/last_duration" 2>/dev/null || echo 0),
    "last_success": "$(cat "${METRICS_DIR}/security/last_success" 2>/dev/null || echo 'never')",
    "last_issue_count": $(cat "${METRICS_DIR}/security/last_issue_count" 2>/dev/null || echo 0)
  },
  "docker": {
    "last_duration": $(cat "${METRICS_DIR}/docker/last_duration" 2>/dev/null || echo 0),
    "last_success": "$(cat "${METRICS_DIR}/docker/last_success" 2>/dev/null || echo 'never')",
    "last_size": "$(cat "${METRICS_DIR}/docker/last_size" 2>/dev/null || echo 'unknown')"
  }
}
EOF
    
    info "Report generated: ${report_file}"
    
    # Display summary
    echo
    echo -e "${BLUE}=== PIPELINE MONITORING SUMMARY ===${NC}"
    echo "Timestamp: $(date)"
    echo
    
    if [[ -f "${METRICS_DIR}/build/last_success" ]]; then
        echo -e "${GREEN}✓${NC} Build: Last success $(cat "${METRICS_DIR}/build/last_success")"
    else
        echo -e "${RED}✗${NC} Build: No successful builds recorded"
    fi
    
    if [[ -f "${METRICS_DIR}/test/last_success" ]]; then
        local coverage=$(cat "${METRICS_DIR}/test/last_coverage" 2>/dev/null || echo "unknown")
        echo -e "${GREEN}✓${NC} Test: Last success $(cat "${METRICS_DIR}/test/last_success") (Coverage: ${coverage}%)"
    else
        echo -e "${RED}✗${NC} Test: No successful test runs recorded"
    fi
    
    if [[ -f "${METRICS_DIR}/lint/last_success" ]]; then
        echo -e "${GREEN}✓${NC} Lint: Last success $(cat "${METRICS_DIR}/lint/last_success")"
    else
        echo -e "${RED}✗${NC} Lint: No successful lint runs recorded"
    fi
    
    if [[ -f "${METRICS_DIR}/security/last_success" ]]; then
        local issues=$(cat "${METRICS_DIR}/security/last_issue_count" 2>/dev/null || echo "unknown")
        echo -e "${GREEN}✓${NC} Security: Last success $(cat "${METRICS_DIR}/security/last_success") (Issues: ${issues})"
    else
        echo -e "${RED}✗${NC} Security: No successful security scans recorded"
    fi
    
    if [[ -f "${METRICS_DIR}/docker/last_success" ]]; then
        local size=$(cat "${METRICS_DIR}/docker/last_size" 2>/dev/null || echo "unknown")
        echo -e "${GREEN}✓${NC} Docker: Last success $(cat "${METRICS_DIR}/docker/last_success") (Size: ${size})"
    else
        echo -e "${YELLOW}!${NC} Docker: No successful builds recorded"
    fi
    
    echo
}

# Main execution
main() {
    local command="${1:-full}"
    
    echo -e "${BLUE}=== FREIGHTLINER CI/CD PIPELINE MONITOR ===${NC}"
    echo "Starting monitoring for: ${command}"
    echo
    
    init_monitoring
    
    case "${command}" in
        "health")
            check_system_health
            ;;
        "build")
            monitor_build
            ;;
        "test")
            monitor_tests
            ;;
        "lint")
            monitor_lint
            ;;
        "security")
            monitor_security
            ;;
        "docker")
            monitor_docker
            ;;
        "report")
            generate_report
            ;;
        "full")
            local overall_status=0
            
            check_system_health || overall_status=1
            monitor_build || overall_status=1
            monitor_tests || overall_status=1
            monitor_lint || overall_status=1
            monitor_security || overall_status=1
            monitor_docker || overall_status=1
            
            generate_report
            
            if [[ ${overall_status} -eq 0 ]]; then
                success "All pipeline stages completed successfully"
            else
                error "One or more pipeline stages failed"
            fi
            
            exit ${overall_status}
            ;;
        *)
            echo "Usage: $0 [health|build|test|lint|security|docker|report|full]"
            echo
            echo "Commands:"
            echo "  health   - Check system health"
            echo "  build    - Monitor build stage"
            echo "  test     - Monitor test stage"
            echo "  lint     - Monitor lint stage"
            echo "  security - Monitor security stage"
            echo "  docker   - Monitor docker stage"
            echo "  report   - Generate monitoring report"
            echo "  full     - Run complete pipeline monitoring (default)"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"