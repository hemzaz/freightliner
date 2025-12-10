#!/bin/bash
# CI/CD Pipeline Recovery System
# Provides automated recovery mechanisms for common pipeline failures

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
RECOVERY_LOG="${PROJECT_ROOT}/pipeline-recovery.log"
BACKUP_DIR="${PROJECT_ROOT}/.github/recovery-backups"
TEMP_DIR="${PROJECT_ROOT}/.github/recovery-temp"

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
    echo -e "${timestamp} [${level}] ${message}" | tee -a "${RECOVERY_LOG}"
}

info() { log "INFO" "$@"; }
warn() { log "WARN" "${YELLOW}$*${NC}"; }
error() { log "ERROR" "${RED}$*${NC}"; }
success() { log "SUCCESS" "${GREEN}$*${NC}"; }
debug() { log "DEBUG" "${CYAN}$*${NC}"; }

# Initialize recovery system
init_recovery_system() {
    info "Initializing pipeline recovery system..."
    
    mkdir -p "${BACKUP_DIR}"
    mkdir -p "${TEMP_DIR}"
    
    success "Recovery system initialized"
}

# Create backup before recovery
create_backup() {
    local backup_name="${1:-$(date +%Y%m%d-%H%M%S)}"
    local backup_path="${BACKUP_DIR}/${backup_name}"
    
    info "Creating backup: ${backup_name}"
    mkdir -p "${backup_path}"
    
    # Backup critical files
    local files_to_backup=(
        ".golangci.yml"
        "go.mod"
        "go.sum"
        "Dockerfile"
        ".github/workflows/ci.yml"
    )
    
    for file in "${files_to_backup[@]}"; do
        if [[ -f "${PROJECT_ROOT}/${file}" ]]; then
            cp "${PROJECT_ROOT}/${file}" "${backup_path}/"
            debug "Backed up: ${file}"
        fi
    done
    
    success "Backup created: ${backup_path}"
    echo "${backup_path}"
}

# Restore from backup
restore_backup() {
    local backup_name="$1"
    local backup_path="${BACKUP_DIR}/${backup_name}"
    
    if [[ ! -d "${backup_path}" ]]; then
        error "Backup not found: ${backup_name}"
        return 1
    fi
    
    info "Restoring from backup: ${backup_name}"
    
    # Restore files
    for file in "${backup_path}"/*; do
        if [[ -f "${file}" ]]; then
            local filename=$(basename "${file}")
            cp "${file}" "${PROJECT_ROOT}/${filename}"
            debug "Restored: ${filename}"
        fi
    done
    
    success "Backup restored: ${backup_name}"
}

# List available backups
list_backups() {
    info "Available backups:"
    
    if [[ ! -d "${BACKUP_DIR}" ]] || [[ -z "$(ls -A "${BACKUP_DIR}" 2>/dev/null)" ]]; then
        warn "No backups found"
        return 0
    fi
    
    for backup in "${BACKUP_DIR}"/*; do
        if [[ -d "${backup}" ]]; then
            local backup_name=$(basename "${backup}")
            local backup_date=$(stat -c %y "${backup}" 2>/dev/null || stat -f %Sm "${backup}" 2>/dev/null || echo "unknown")
            info "  ${backup_name} (${backup_date})"
        fi
    done
}

# Emergency recovery - restore to known good state
emergency_recovery() {
    warn "Initiating emergency recovery..."
    
    # Create emergency backup first
    local emergency_backup
    emergency_backup=$(create_backup "emergency-$(date +%Y%m%d-%H%M%S)")
    
    # Reset to clean state
    cd "${PROJECT_ROOT}"
    
    # Clean all caches and temporary files
    info "Cleaning caches and temporary files..."
    go clean -cache -testcache -modcache || true
    
    # Remove potential problematic files
    rm -f coverage.out gosec-results.sarif *.prof || true
    
    # Restore Go modules
    info "Restoring Go modules..."
    if [[ -f "go.mod" ]]; then
        go mod tidy
        go mod download
        go mod verify
    else
        error "go.mod not found - cannot restore modules"
        return 1
    fi
    
    # Fix golangci-lint configuration
    info "Fixing golangci-lint configuration..."
    fix_golangci_lint_config
    
    # Test basic functionality
    info "Testing basic functionality..."
    if go build -v ./...; then
        success "Build test passed"
    else
        error "Build test failed"
        return 1
    fi
    
    success "Emergency recovery completed"
    info "Emergency backup created: ${emergency_backup}"
}

# Fix golangci-lint configuration issues
fix_golangci_lint_config() {
    local config_file="${PROJECT_ROOT}/.golangci.yml"
    
    if [[ ! -f "${config_file}" ]]; then
        warn "golangci-lint config not found, creating default..."
        create_default_golangci_config
        return 0
    fi
    
    info "Fixing golangci-lint configuration..."
    
    # Create backup of current config
    cp "${config_file}" "${config_file}.recovery-backup"
    
    # Fix deprecated options
    if grep -q "colored-line-number:" "${config_file}"; then
        info "Fixing deprecated colored-line-number option..."
        
        # Create temporary file with fixed configuration
        cat > "${TEMP_DIR}/golangci-fix.yml" << 'EOF'
# golangci-lint configuration for Freightliner
# Optimized for CI/CD performance and comprehensive analysis

run:
  # Timeout for analysis
  timeout: 8m
  
  # Exit code when at least one issue was found
  issues-exit-code: 1
  
  # Include test files in analysis
  tests: true
  
  # Allow parallel runners
  concurrency: 4
  
  # Define the Go version to target
  go: '1.24.5'

# Output configuration
output:
  # Format configuration (replaces deprecated colored-line-number)
  formats:
    - format: colored-line-number
      path: stdout
  print-issued-lines: true
  print-linter-name: true
  uniq-by-line: true
  sort-results: true

# Linters configuration
linters:
  # Enable specific linters
  enable:
    # Essential linters (always enabled)
    - errcheck      # Check for unchecked errors
    - gosimple      # Simplify code
    - govet         # Examine Go source code and report suspicious constructs
    - ineffassign   # Detect ineffectual assignments
    - staticcheck   # State-of-the-art Go linter
    - typecheck     # Parse and type-check Go code
    - unused        # Check for unused constants, variables, functions and types
    
    # Security and reliability
    - gosec         # Inspect source code for security problems
    - bodyclose     # Check whether HTTP response body is closed successfully
    - contextcheck  # Check whether the function uses a non-inherited context
    - errorlint     # Find code that will cause problems with the error wrapping scheme
    - rowserrcheck  # Check whether Err of rows is checked successfully
    - sqlclosecheck # Check that sql.Rows and sql.Stmt are closed
    
    # Code quality
    - dupl          # Tool for code clone detection
    - gocognit      # Compute and check the cognitive complexity of functions
    - goconst       # Find repeated strings that could be replaced by a constant
    - gocyclo       # Compute and check the cyclomatic complexity of functions
    - revive        # Fast, configurable, extensible, flexible, and beautiful linter for Go
    - stylecheck    # Replacement for golint
    
    # Formatting and imports
    - gofmt         # Check whether code was gofmt-ed
    - goimports     # Check import statements are formatted according to the 'goimport' command
    - misspell      # Find commonly misspelled English words in comments
    
    # Performance and best practices
    - prealloc      # Find slice declarations that could potentially be preallocated
    - unconvert     # Remove unnecessary type conversions
    - unparam       # Report unused function parameters

# Issues configuration
issues:
  # Fix found issues (if it's supported by the linter)
  fix: false
  
  # Maximum issues count per one linter
  max-issues-per-linter: 50
  
  # Maximum count of issues with the same text
  max-same-issues: 3
  
  # Exclude some patterns
  exclude:
    # errcheck: Almost all programs ignore errors on these functions and in most cases it's ok
    - Error return value of .((os\.)?std(out|err)\..*|.*Close|.*Flush|os\.Remove(All)?|.*print.*|os\.(Un)?Setenv). is not checked
    # govet: Common false positives
    - (possible misuse of unsafe.Pointer|should have signature)
  
  # Excluding configuration per-path and per-linter
  exclude-rules:
    # Exclude some linters from running on tests files
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec
        - goconst
        - gocognit
    
    # Exclude some linters from main.go
    - path: main\.go
      linters:
        - gochecknoglobals
EOF
        
        # Replace the current config with the fixed version
        cp "${TEMP_DIR}/golangci-fix.yml" "${config_file}"
        success "Fixed golangci-lint configuration"
    fi
    
    # Validate the configuration
    if command -v golangci-lint &> /dev/null; then
        if golangci-lint config verify; then
            success "golangci-lint configuration validated"
            rm -f "${config_file}.recovery-backup"
        else
            error "Configuration validation failed, restoring backup"
            mv "${config_file}.recovery-backup" "${config_file}"
            return 1
        fi
    else
        warn "golangci-lint not available for validation"
    fi
}

# Create default golangci-lint configuration
create_default_golangci_config() {
    local config_file="${PROJECT_ROOT}/.golangci.yml"
    
    info "Creating default golangci-lint configuration..."
    
    cat > "${config_file}" << 'EOF'
# golangci-lint configuration for Freightliner
# Optimized for CI/CD performance and comprehensive analysis

run:
  timeout: 8m
  issues-exit-code: 1
  tests: true
  concurrency: 4
  go: '1.24.5'

output:
  formats:
    - format: colored-line-number
      path: stdout
  print-issued-lines: true
  print-linter-name: true
  uniq-by-line: true
  sort-results: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gosec
    - bodyclose
    - contextcheck
    - errorlint
    - rowserrcheck
    - sqlclosecheck
    - dupl
    - gocognit
    - goconst
    - gocyclo
    - revive
    - stylecheck
    - gofmt
    - goimports
    - misspell
    - prealloc
    - unconvert
    - unparam

issues:
  fix: false
  max-issues-per-linter: 50
  max-same-issues: 3
  exclude:
    - Error return value of .((os\.)?std(out|err)\..*|.*Close|.*Flush|os\.Remove(All)?|.*print.*|os\.(Un)?Setenv). is not checked
    - (possible misuse of unsafe.Pointer|should have signature)
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec
        - goconst
        - gocognit
    - path: main\.go
      linters:
        - gochecknoglobals
EOF
    
    success "Default golangci-lint configuration created"
}

# Fix Docker-related issues
fix_docker_issues() {
    info "Fixing Docker-related issues..."
    
    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed"
        return 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        warn "Docker daemon is not running"
        return 1
    fi
    
    # Clean up Docker resources
    info "Cleaning up Docker resources..."
    docker system prune -f || warn "Docker system prune failed"
    
    # Test Docker build
    local dockerfile="${PROJECT_ROOT}/Dockerfile"
    if [[ -f "${dockerfile}" ]]; then
        info "Testing Docker build..."
        if docker build -t freightliner:recovery-test "${PROJECT_ROOT}"; then
            success "Docker build test passed"
            
            # Test the image
            if docker run --rm freightliner:recovery-test --version; then
                success "Docker image test passed"
            else
                warn "Docker image test failed"
            fi
            
            # Clean up test image
            docker rmi freightliner:recovery-test || true
        else
            error "Docker build test failed"
            return 1
        fi
    else
        warn "Dockerfile not found"
    fi
    
    success "Docker issues fixed"
}

# Fix Go module issues
fix_go_module_issues() {
    info "Fixing Go module issues..."
    
    cd "${PROJECT_ROOT}"
    
    # Check if go.mod exists
    if [[ ! -f "go.mod" ]]; then
        error "go.mod not found"
        return 1
    fi
    
    # Clean module cache
    info "Cleaning module cache..."
    go clean -modcache
    
    # Tidy modules
    info "Tidying modules..."
    if ! go mod tidy; then
        error "Failed to tidy modules"
        return 1
    fi
    
    # Download dependencies
    info "Downloading dependencies..."
    if ! go mod download; then
        error "Failed to download dependencies"
        return 1
    fi
    
    # Verify modules
    info "Verifying modules..."
    if ! go mod verify; then
        error "Module verification failed"
        return 1
    fi
    
    success "Go module issues fixed"
}

# Fix test-related issues
fix_test_issues() {
    info "Fixing test-related issues..."
    
    cd "${PROJECT_ROOT}"
    
    # Clean test cache
    info "Cleaning test cache..."
    go clean -testcache
    
    # Find and analyze test files
    local test_files
    if test_files=$(find . -name "*_test.go" -type f); then
        local test_count
        test_count=$(echo "${test_files}" | wc -l)
        info "Found ${test_count} test files"
        
        # Check for common test issues
        info "Analyzing test files for common issues..."
        
        local files_with_issues=()
        while IFS= read -r file; do
            # Check for potential timeout issues
            if grep -q "time\.Hour\|[5-9][0-9].*time\.Minute" "${file}"; then
                files_with_issues+=("${file} (long timeouts)")
                warn "Found potentially long timeout in: ${file}"
            fi
            
            # Check for potential race conditions
            if grep -q "go func\|goroutine" "${file}" && ! grep -q "sync\." "${file}"; then
                files_with_issues+=("${file} (potential race)")
                warn "Found potential race condition in: ${file}"
            fi
        done <<< "${test_files}"
        
        if [[ ${#files_with_issues[@]} -gt 0 ]]; then
            warn "Found ${#files_with_issues[@]} test files with potential issues:"
            for issue in "${files_with_issues[@]}"; do
                warn "  - ${issue}"
            done
        fi
    else
        warn "No test files found"
    fi
    
    # Test basic functionality
    info "Running basic test functionality check..."
    if go test -v -short ./...; then
        success "Basic test functionality works"
    else
        warn "Basic test functionality has issues"
    fi
    
    success "Test issue analysis completed"
}

# Automated recovery based on detected issues
auto_recovery() {
    info "Starting automated recovery process..."
    
    local recovery_actions=()
    local issues_detected=0
    
    # Check golangci-lint configuration
    if [[ -f "${PROJECT_ROOT}/.golangci.yml" ]]; then
        if command -v golangci-lint &> /dev/null; then
            if ! golangci-lint config verify &> /dev/null; then
                warn "golangci-lint configuration issues detected"
                recovery_actions+=("fix_golangci_config")
                ((issues_detected++))
            fi
        fi
    fi
    
    # Check Go modules
    cd "${PROJECT_ROOT}"
    if [[ -f "go.mod" ]]; then
        if ! go mod verify &> /dev/null; then
            warn "Go module issues detected"
            recovery_actions+=("fix_go_modules")
            ((issues_detected++))
        fi
    fi
    
    # Check Docker (if available)
    if command -v docker &> /dev/null; then
        if docker info &> /dev/null; then
            if [[ -f "Dockerfile" ]] && ! docker build -q -t freightliner:health-check . &> /dev/null; then
                warn "Docker build issues detected"
                recovery_actions+=("fix_docker")
                ((issues_detected++))
                docker rmi freightliner:health-check &> /dev/null || true
            fi
        fi
    fi
    
    if [[ ${issues_detected} -eq 0 ]]; then
        success "No issues detected - system appears healthy"
        return 0
    fi
    
    info "Detected ${issues_detected} issues, executing recovery actions..."
    
    # Create backup before making changes
    local backup_path
    backup_path=$(create_backup "auto-recovery-$(date +%Y%m%d-%H%M%S)")
    
    # Execute recovery actions
    local recovery_success=0
    for action in "${recovery_actions[@]}"; do
        case "${action}" in
            "fix_golangci_config")
                if fix_golangci_lint_config; then
                    success "Fixed golangci-lint configuration"
                else
                    error "Failed to fix golangci-lint configuration"
                    ((recovery_success++))
                fi
                ;;
            "fix_go_modules")
                if fix_go_module_issues; then
                    success "Fixed Go module issues"
                else
                    error "Failed to fix Go module issues"
                    ((recovery_success++))
                fi
                ;;
            "fix_docker")
                if fix_docker_issues; then
                    success "Fixed Docker issues"
                else
                    error "Failed to fix Docker issues"
                    ((recovery_success++))
                fi
                ;;
        esac
    done
    
    if [[ ${recovery_success} -eq 0 ]]; then
        success "All recovery actions completed successfully"
        info "Backup created at: ${backup_path}"
        return 0
    else
        error "${recovery_success} recovery actions failed"
        warn "You may need to restore from backup: ${backup_path}"
        return 1
    fi
}

# Health check
health_check() {
    info "Running pipeline health check..."
    
    local issues=0
    
    # Check Go
    if ! command -v go &> /dev/null; then
        error "Go is not installed"
        ((issues++))
    else
        local go_version
        go_version=$(go version | grep -o 'go[0-9]*\.[0-9]*\.[0-9]*')
        info "Go version: ${go_version}"
    fi
    
    # Check golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        error "golangci-lint is not installed"
        ((issues++))
    else
        local lint_version
        lint_version=$(golangci-lint --version | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*')
        info "golangci-lint version: ${lint_version}"
        
        if [[ -f "${PROJECT_ROOT}/.golangci.yml" ]]; then
            if golangci-lint config verify; then
                success "golangci-lint configuration is valid"
            else
                error "golangci-lint configuration is invalid"
                ((issues++))
            fi
        fi
    fi
    
    # Check Go modules
    cd "${PROJECT_ROOT}"
    if [[ -f "go.mod" ]]; then
        if go mod verify &> /dev/null; then
            success "Go modules are valid"
        else
            error "Go module issues detected"
            ((issues++))
        fi
    else
        error "go.mod not found"
        ((issues++))
    fi
    
    # Check Docker (optional)
    if command -v docker &> /dev/null; then
        if docker info &> /dev/null; then
            success "Docker is available and running"
        else
            warn "Docker daemon is not running"
        fi
    else
        info "Docker is not installed (optional)"
    fi
    
    if [[ ${issues} -eq 0 ]]; then
        success "Health check passed - system is healthy"
        return 0
    else
        error "Health check failed with ${issues} issues"
        return 1
    fi
}

# Main execution function
main() {
    local command="${1:-help}"
    
    echo -e "${BLUE}=== FREIGHTLINER PIPELINE RECOVERY SYSTEM ===${NC}"
    echo "Command: ${command}"
    echo
    
    init_recovery_system
    
    case "${command}" in
        "health"|"check")
            health_check
            ;;
        "auto"|"auto-recovery")
            auto_recovery
            ;;
        "emergency")
            emergency_recovery
            ;;
        "backup")
            local backup_name="${2:-$(date +%Y%m%d-%H%M%S)}"
            create_backup "${backup_name}"
            ;;
        "restore")
            if [[ -z "${2:-}" ]]; then
                error "Backup name required for restore"
                list_backups
                exit 1
            fi
            restore_backup "$2"
            ;;
        "list-backups")
            list_backups
            ;;
        "fix-golangci")
            fix_golangci_lint_config
            ;;
        "fix-docker")
            fix_docker_issues
            ;;
        "fix-modules")
            fix_go_module_issues
            ;;
        "fix-tests")
            fix_test_issues
            ;;
        "clean")
            info "Performing deep clean..."
            go clean -cache -testcache -modcache
            docker system prune -af || true
            rm -rf "${TEMP_DIR}"/* || true
            success "Deep clean completed"
            ;;
        *)
            echo "Usage: $0 [command] [options]"
            echo
            echo "Commands:"
            echo "  health              - Run health check"
            echo "  auto                - Automatic recovery based on detected issues"
            echo "  emergency           - Emergency recovery to known good state"
            echo "  backup [name]       - Create backup of critical files"
            echo "  restore <name>      - Restore from backup"
            echo "  list-backups        - List available backups"
            echo "  fix-golangci        - Fix golangci-lint configuration"
            echo "  fix-docker          - Fix Docker-related issues"
            echo "  fix-modules         - Fix Go module issues"
            echo "  fix-tests           - Analyze and fix test issues"
            echo "  clean               - Deep clean caches and temporary files"
            echo
            echo "Examples:"
            echo "  $0 health                      # Check system health"
            echo "  $0 auto                        # Run automatic recovery"
            echo "  $0 backup my-backup            # Create named backup"
            echo "  $0 restore my-backup           # Restore from backup"
            echo "  $0 emergency                   # Emergency recovery"
            exit 1
            ;;
    esac
}

# Handle script termination
cleanup() {
    info "Cleaning up temporary files..."
    rm -rf "${TEMP_DIR}"/* 2>/dev/null || true
}

trap cleanup EXIT

# Run main function with all arguments
main "$@"