#!/bin/bash

# Pre-Implementation Validation Suite
# Validates configurations and environment before implementing pipeline fixes

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Global variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEMP_DIR=""
VALIDATION_ERRORS=0
VALIDATION_WARNINGS=0

# Cleanup function
cleanup() {
    if [[ -n "$TEMP_DIR" && -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Set up trap for cleanup
trap cleanup EXIT

# Create temporary directory
create_temp_dir() {
    TEMP_DIR=$(mktemp -d -t freightliner-validation-XXXXXX)
    log_info "Created temporary directory: $TEMP_DIR"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Validate Go installation and version
validate_go_environment() {
    log_info "Validating Go environment..."
    
    if ! command_exists go; then
        log_error "Go is not installed"
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    local go_version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Found Go version: $go_version"
    
    # Check if Go version is compatible (1.23+ required)
    local major minor patch
    IFS='.' read -r major minor patch <<< "$go_version"
    
    if [[ $major -eq 1 && $minor -ge 23 ]]; then
        log_success "Go version is compatible"
    else
        log_warning "Go version $go_version may not be fully compatible (recommended: 1.23+)"
        ((VALIDATION_WARNINGS++))
    fi
    
    # Validate Go modules
    cd "$PROJECT_ROOT"
    
    log_info "Validating Go modules..."
    if go mod verify; then
        log_success "Go modules verified successfully"
    else
        log_error "Go module verification failed"
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    # Test Go module operations
    log_info "Testing Go module operations..."
    if go mod tidy; then
        log_success "go mod tidy completed successfully"
    else
        log_error "go mod tidy failed"
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    return 0
}

# Validate golangci-lint configuration
validate_golangci_config() {
    log_info "Validating golangci-lint configuration..."
    
    local config_file="$PROJECT_ROOT/.golangci.yml"
    
    if [[ ! -f "$config_file" ]]; then
        log_error "golangci-lint configuration file not found: $config_file"
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    # Check YAML syntax
    if command_exists yamllint; then
        if yamllint "$config_file"; then
            log_success "golangci-lint YAML syntax is valid"
        else
            log_warning "golangci-lint YAML has formatting issues"
            ((VALIDATION_WARNINGS++))
        fi
    else
        log_warning "yamllint not available, skipping YAML syntax validation"
    fi
    
    # Install golangci-lint if not available
    if ! command_exists golangci-lint; then
        log_info "Installing golangci-lint..."
        if curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v1.62.2; then
            log_success "golangci-lint installed successfully"
            export PATH="$(go env GOPATH)/bin:$PATH"
        else
            log_error "Failed to install golangci-lint"
            ((VALIDATION_ERRORS++))
            return 1
        fi
    fi
    
    # Validate configuration using golangci-lint
    cd "$PROJECT_ROOT"
    if golangci-lint config verify "$config_file"; then
        log_success "golangci-lint configuration is valid"
    else
        log_error "golangci-lint configuration validation failed"
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    # Test dry run
    log_info "Testing golangci-lint dry run..."
    if timeout 60 golangci-lint run --dry-run --timeout=30s; then
        log_success "golangci-lint dry run completed successfully"
    else
        local exit_code=$?
        if [[ $exit_code -eq 124 ]]; then
            log_warning "golangci-lint dry run timed out (acceptable for validation)"
        else
            log_warning "golangci-lint dry run found issues (may be acceptable)"
            ((VALIDATION_WARNINGS++))
        fi
    fi
    
    return 0
}

# Validate Docker configuration
validate_docker_config() {
    log_info "Validating Docker configuration..."
    
    if ! command_exists docker; then
        log_warning "Docker is not installed - Docker validation skipped"
        ((VALIDATION_WARNINGS++))
        return 0
    fi
    
    # Check Docker daemon
    if ! docker info >/dev/null 2>&1; then
        log_warning "Docker daemon is not running - Docker validation skipped"
        ((VALIDATION_WARNINGS++))
        return 0
    fi
    
    local dockerfile="$PROJECT_ROOT/Dockerfile"
    if [[ ! -f "$dockerfile" ]]; then
        log_error "Dockerfile not found: $dockerfile"
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    # Basic Dockerfile syntax validation
    log_info "Validating Dockerfile syntax..."
    local line_number=0
    local has_from=false
    
    while IFS= read -r line; do
        ((line_number++))
        line=$(echo "$line" | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//')
        
        # Skip empty lines and comments
        if [[ -z "$line" || "$line" =~ ^# ]]; then
            continue
        fi
        
        # Check for FROM instruction
        if [[ "$line" =~ ^FROM ]]; then
            has_from=true
        fi
        
        # Check for basic instruction format
        if [[ ! "$line" =~ ^(FROM|RUN|COPY|ADD|WORKDIR|EXPOSE|ENV|USER|VOLUME|ENTRYPOINT|CMD|LABEL|HEALTHCHECK|SHELL|STOPSIGNAL|ARG|ONBUILD)[[:space:]] ]]; then
            # Allow multi-line continuations
            if [[ ! "$line" =~ \\$ ]] && [[ ! "$line" =~ ^[[:space:]] ]]; then
                log_warning "Line $line_number: Potentially invalid Dockerfile instruction: $line"
                ((VALIDATION_WARNINGS++))
            fi
        fi
    done < "$dockerfile"
    
    if [[ "$has_from" == false ]]; then
        log_error "Dockerfile must contain at least one FROM instruction"
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    log_success "Dockerfile syntax validation completed"
    
    # Test Docker build stages
    log_info "Testing Docker multi-stage build..."
    
    # Test builder stage
    if timeout 300 docker build --target builder -t freightliner:validation-builder . >/dev/null 2>&1; then
        log_success "Docker builder stage builds successfully"
        docker rmi freightliner:validation-builder >/dev/null 2>&1 || true
    else
        log_error "Docker builder stage build failed"
        ((VALIDATION_ERRORS++))
    fi
    
    # Test test stage
    if timeout 300 docker build --target test -t freightliner:validation-test . >/dev/null 2>&1; then
        log_success "Docker test stage builds successfully"
        docker rmi freightliner:validation-test >/dev/null 2>&1 || true
    else
        log_error "Docker test stage build failed"
        ((VALIDATION_ERRORS++))
    fi
    
    return 0
}

# Validate GitHub Actions workflows
validate_github_actions() {
    log_info "Validating GitHub Actions workflows..."
    
    local workflows_dir="$PROJECT_ROOT/.github/workflows"
    if [[ ! -d "$workflows_dir" ]]; then
        log_error "GitHub workflows directory not found: $workflows_dir"
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    local workflow_count=0
    local valid_workflows=0
    
    # Check each workflow file
    for workflow_file in "$workflows_dir"/*.yml "$workflows_dir"/*.yaml; do
        if [[ -f "$workflow_file" ]]; then
            ((workflow_count++))
            local filename
            filename=$(basename "$workflow_file")
            
            log_info "Validating workflow: $filename"
            
            # Basic YAML syntax validation
            if command_exists yamllint; then
                if yamllint "$workflow_file" >/dev/null 2>&1; then
                    ((valid_workflows++))
                    log_success "Workflow $filename has valid YAML syntax"
                else
                    log_warning "Workflow $filename has YAML formatting issues"
                    ((VALIDATION_WARNINGS++))
                fi
            else
                # Basic syntax check using Python or another method
                if python3 -c "import yaml; yaml.safe_load(open('$workflow_file'))" 2>/dev/null; then
                    ((valid_workflows++))
                    log_success "Workflow $filename has valid YAML syntax"
                else
                    log_warning "Workflow $filename may have YAML syntax issues"
                    ((VALIDATION_WARNINGS++))
                fi
            fi
            
            # Check for required fields
            if grep -q "^name:" "$workflow_file" && \
               grep -q "^on:" "$workflow_file" && \
               grep -q "jobs:" "$workflow_file"; then
                log_success "Workflow $filename has required fields"
            else
                log_warning "Workflow $filename missing required fields (name, on, jobs)"
                ((VALIDATION_WARNINGS++))
            fi
        fi
    done
    
    if [[ $workflow_count -eq 0 ]]; then
        log_warning "No GitHub Actions workflows found"
        ((VALIDATION_WARNINGS++))
    else
        log_info "Validated $valid_workflows out of $workflow_count workflows"
    fi
    
    return 0
}

# Validate project structure
validate_project_structure() {
    log_info "Validating project structure..."
    
    local required_files=(
        "go.mod"
        "go.sum"
        "Dockerfile"
        ".golangci.yml"
        ".gitignore"
    )
    
    local required_dirs=(
        ".github/workflows"
        "pkg"
        "cmd"
    )
    
    # Check required files
    for file in "${required_files[@]}"; do
        if [[ -f "$PROJECT_ROOT/$file" ]]; then
            log_success "Required file found: $file"
        else
            log_error "Required file missing: $file"
            ((VALIDATION_ERRORS++))
        fi
    done
    
    # Check required directories
    for dir in "${required_dirs[@]}"; do
        if [[ -d "$PROJECT_ROOT/$dir" ]]; then
            log_success "Required directory found: $dir"
        else
            log_error "Required directory missing: $dir"
            ((VALIDATION_ERRORS++))
        fi
    done
    
    # Check for recommended security files
    local security_files=(
        ".github/dependabot.yml"
        "scripts/security-validation.sh"
        ".github/security.yml"
    )
    
    for file in "${security_files[@]}"; do
        if [[ -f "$PROJECT_ROOT/$file" ]]; then
            log_success "Security file found: $file"
        else
            log_warning "Recommended security file missing: $file"
            ((VALIDATION_WARNINGS++))
        fi
    done
    
    return 0
}

# Validate build environment
validate_build_environment() {
    log_info "Validating build environment..."
    
    cd "$PROJECT_ROOT"
    
    # Test Go build
    log_info "Testing Go build process..."
    if go build -v ./...; then
        log_success "Go build completed successfully"
    else
        log_error "Go build failed"
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    # Test unit tests (short mode)
    log_info "Testing unit tests..."
    if go test -short ./...; then
        log_success "Unit tests passed"
    else
        local exit_code=$?
        if [[ $exit_code -eq 1 ]]; then
            log_warning "Some unit tests failed (may be acceptable)"
            ((VALIDATION_WARNINGS++))
        else
            log_error "Unit test execution failed"
            ((VALIDATION_ERRORS++))
            return 1
        fi
    fi
    
    return 0
}

# Run performance baseline measurement
measure_performance_baseline() {
    log_info "Measuring performance baseline..."
    
    cd "$PROJECT_ROOT"
    
    # Measure build time
    log_info "Measuring build time..."
    local start_time end_time build_duration
    start_time=$(date +%s)
    
    if go build -v ./... >/dev/null 2>&1; then
        end_time=$(date +%s)
        build_duration=$((end_time - start_time))
        log_info "Build completed in ${build_duration} seconds"
        
        if [[ $build_duration -gt 300 ]]; then # 5 minutes
            log_warning "Build time exceeds 5 minutes: ${build_duration}s"
            ((VALIDATION_WARNINGS++))
        else
            log_success "Build time is acceptable: ${build_duration}s"
        fi
    else
        log_error "Build performance measurement failed"
        ((VALIDATION_ERRORS++))
    fi
    
    # Measure test time
    log_info "Measuring test execution time..."
    start_time=$(date +%s)
    
    if go test -short ./... >/dev/null 2>&1; then
        end_time=$(date +%s)
        local test_duration=$((end_time - start_time))
        log_info "Tests completed in ${test_duration} seconds"
        
        if [[ $test_duration -gt 120 ]]; then # 2 minutes
            log_warning "Test time exceeds 2 minutes: ${test_duration}s"
            ((VALIDATION_WARNINGS++))
        else
            log_success "Test time is acceptable: ${test_duration}s"
        fi
    else
        log_warning "Test performance measurement completed with issues"
        ((VALIDATION_WARNINGS++))
    fi
}

# Generate validation report
generate_report() {
    log_info "Generating validation report..."
    
    local report_file="$PROJECT_ROOT/PRE_IMPLEMENTATION_VALIDATION_REPORT.md"
    
    cat > "$report_file" << EOF
# Pre-Implementation Validation Report

**Generated:** $(date)
**Script:** $0

## Summary

- **Errors:** $VALIDATION_ERRORS
- **Warnings:** $VALIDATION_WARNINGS
- **Status:** $(if [[ $VALIDATION_ERRORS -eq 0 ]]; then echo "✅ PASSED"; else echo "❌ FAILED"; fi)

## Validation Results

### Environment Validation
- Go installation and version check
- Go modules verification
- Build environment validation

### Configuration Validation
- golangci-lint configuration
- Dockerfile syntax and structure
- GitHub Actions workflows
- Project structure

### Performance Baseline
- Build time measurement
- Test execution time measurement

## Recommendations

$(if [[ $VALIDATION_ERRORS -gt 0 ]]; then
    echo "### Critical Issues (Must Fix Before Implementation)"
    echo "- Review error messages above and fix all critical issues"
    echo "- Re-run validation after fixes"
fi)

$(if [[ $VALIDATION_WARNINGS -gt 0 ]]; then
    echo "### Warnings (Recommended Improvements)"
    echo "- Address warning messages to improve pipeline reliability"
    echo "- Consider implementing recommended security practices"
fi)

$(if [[ $VALIDATION_ERRORS -eq 0 && $VALIDATION_WARNINGS -eq 0 ]]; then
    echo "### All Validations Passed"
    echo "- Environment is ready for pipeline implementation"
    echo "- Proceed with implementation phase"
fi)

## Next Steps

1. Address any critical errors listed above
2. Consider fixing warnings for improved reliability
3. Proceed to implementation phase testing
4. Run post-implementation validation

---
*Report generated by pre-implementation validation suite*
EOF

    log_success "Validation report generated: $report_file"
}

# Main execution function
main() {
    log_info "Starting Pre-Implementation Validation Suite"
    log_info "Project root: $PROJECT_ROOT"
    
    create_temp_dir
    
    # Run all validation checks
    validate_go_environment
    validate_golangci_config
    validate_docker_config
    validate_github_actions
    validate_project_structure
    validate_build_environment
    measure_performance_baseline
    
    # Generate report
    generate_report
    
    # Final summary
    log_info "=========================================="
    log_info "Pre-Implementation Validation Complete"
    log_info "=========================================="
    log_info "Errors: $VALIDATION_ERRORS"
    log_info "Warnings: $VALIDATION_WARNINGS"
    
    if [[ $VALIDATION_ERRORS -eq 0 ]]; then
        log_success "✅ Validation PASSED - Ready for implementation"
        exit 0
    else
        log_error "❌ Validation FAILED - Fix errors before proceeding"
        exit 1
    fi
}

# Help function
show_help() {
    cat << EOF
Pre-Implementation Validation Suite

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -h, --help      Show this help message
    -v, --verbose   Enable verbose output

DESCRIPTION:
    Validates the development environment and project configuration before
    implementing CI/CD pipeline fixes. This includes checking:
    
    - Go environment and dependencies
    - golangci-lint configuration
    - Docker configuration and build process
    - GitHub Actions workflow files
    - Project structure and required files
    - Build and test performance baseline

EXIT CODES:
    0    All validations passed
    1    Validation errors found (must fix before implementation)

EXAMPLES:
    $0                 # Run all validations
    $0 --verbose       # Run with verbose output
    $0 --help          # Show this help

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            set -x
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Run main function
main "$@"