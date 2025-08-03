#!/bin/bash
# Security Validation Script
# Validates all Kubernetes security fixes and compliance

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

# Validation counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

# Function to run a validation check
run_check() {
    local check_name="$1"
    local check_command="$2"
    local expected_result="${3:-0}"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    log_info "Running check: $check_name"
    
    if eval "$check_command"; then
        if [ $? -eq $expected_result ]; then
            log_success "$check_name - PASSED"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
            return 0
        else
            log_error "$check_name - FAILED (unexpected result)"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            return 1
        fi
    else
        log_error "$check_name - FAILED"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        return 1
    fi
}

# Function to check if a file contains specific content
check_file_content() {
    local file="$1"
    local pattern="$2"
    local description="$3"
    
    if [ -f "$file" ]; then
        if grep -q "$pattern" "$file"; then
            log_success "$description - Found in $file"
            return 0
        else
            log_error "$description - NOT found in $file"
            return 1
        fi
    else
        log_error "$file - File not found"
        return 1
    fi
}

# Main validation function
main() {
    log_info "Starting Kubernetes Security Validation"
    log_info "============================================"
    
    # Change to the repository root
    cd "$(dirname "$0")/.."
    
    # Validate Kubernetes deployment security fixes
    log_info "\n1. Validating Kubernetes Deployment Security"
    log_info "--------------------------------------------"
    
    # Check 1: Image uses digest instead of tag
    run_check "Image uses digest (CKV_K8S_43)" \
        "check_file_content 'deployments/kubernetes/deployment.yaml' '@sha256:' 'Container image with digest'"
    
    # Check 2: Image pull policy is Always
    run_check "Image pull policy is Always (CKV_K8S_15)" \
        "check_file_content 'deployments/kubernetes/deployment.yaml' 'imagePullPolicy: Always' 'ImagePullPolicy set to Always'"
    
    # Check 3: Service account token automount disabled
    run_check "Service account token automount disabled (CKV_K8S_38)" \
        "check_file_content 'deployments/kubernetes/deployment.yaml' 'automountServiceAccountToken: false' 'Service account token automount disabled'"
    
    # Check 4: Resource limits and requests are set
    run_check "CPU and Memory limits set (CKV_K8S_11, CKV_K8S_13)" \
        "check_file_content 'deployments/kubernetes/deployment.yaml' 'limits:' 'Resource limits configured'"
    
    run_check "CPU and Memory requests set (CKV_K8S_10, CKV_K8S_12)" \
        "check_file_content 'deployments/kubernetes/deployment.yaml' 'requests:' 'Resource requests configured'"
    
    # Check 5: Security context configured
    run_check "Security context configured" \
        "check_file_content 'deployments/kubernetes/deployment.yaml' 'runAsNonRoot: true' 'runAsNonRoot security context'"
    
    run_check "Read-only root filesystem" \
        "check_file_content 'deployments/kubernetes/deployment.yaml' 'readOnlyRootFilesystem: true' 'Read-only root filesystem'"
    
    run_check "Capabilities dropped" \
        "check_file_content 'deployments/kubernetes/deployment.yaml' 'drop:' 'Capabilities dropped'"
    
    # Validate Ingress security fixes
    log_info "\n2. Validating Ingress Security"
    log_info "------------------------------"
    
    # Check 6: No configuration-snippet annotation (CVE-2021-25742)
    run_check "No dangerous configuration-snippet (CKV_K8S_153)" \
        "! grep -q 'configuration-snippet:' deployments/kubernetes/ingress.yaml" \
        "0"
    
    # Check 7: TLS configured
    run_check "TLS configured in Ingress" \
        "check_file_content 'deployments/kubernetes/ingress.yaml' 'tls:' 'TLS configuration'"
    
    # Check 8: Security headers configured (using safe annotations)
    run_check "Security headers configured" \
        "check_file_content 'deployments/kubernetes/ingress.yaml' 'auth-response-headers' 'Security headers via safe annotations'"
    
    # Validate Security Policies
    log_info "\n3. Validating Security Policies"
    log_info "-------------------------------"
    
    # Check 9: PodSecurityPolicy exists
    run_check "PodSecurityPolicy exists" \
        "[ -f 'deployments/kubernetes/pod-security-policy.yaml' ]"
    
    # Check 10: PodSecurityStandards exists
    run_check "PodSecurityStandards exists" \
        "[ -f 'deployments/kubernetes/pod-security-standards.yaml' ]"
    
    # Check 11: RBAC configuration exists
    run_check "RBAC configuration exists" \
        "[ -f 'deployments/kubernetes/rbac.yaml' ]"
    
    # Check 12: NetworkPolicy configured
    run_check "NetworkPolicy configured" \
        "check_file_content 'deployments/kubernetes/rbac.yaml' 'NetworkPolicy' 'NetworkPolicy configuration'"
    
    # Validate GitHub Actions security fixes
    log_info "\n4. Validating GitHub Actions Security"
    log_info "------------------------------------"
    
    # Check 13: Workflow inputs security (either required=true or no inputs)
    local workflow_files=(
        ".github/workflows/security-monitoring.yml"
        ".github/workflows/scheduled-comprehensive.yml"
        ".github/workflows/security-monitoring-enhanced.yml"
        ".github/workflows/oidc-authentication.yml"
    )
    
    for workflow in "${workflow_files[@]}"; do
        if [ -f "$workflow" ]; then
            if grep -q "required: true" "$workflow" || ! grep -q "inputs:" "$workflow"; then
                log_success "Workflow security - $(basename "$workflow") - SECURE (required inputs or no inputs)"
                PASSED_CHECKS=$((PASSED_CHECKS + 1))
            else
                log_error "Workflow security - $(basename "$workflow") - INSECURE (optional inputs found)"
                FAILED_CHECKS=$((FAILED_CHECKS + 1))
            fi
            TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
        else
            log_warning "Workflow file not found: $workflow"
            WARNINGS=$((WARNINGS + 1))
        fi
    done
    
    # Run Checkov if available
    log_info "\n5. Running Checkov Security Scan"
    log_info "--------------------------------"
    
    if command -v checkov &> /dev/null; then
        log_info "Running Checkov scan on Kubernetes files..."
        if checkov -d deployments/kubernetes/ --framework kubernetes --check CKV_K8S_* --output cli; then
            log_success "Checkov scan completed successfully"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
        else
            log_error "Checkov scan found security issues"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        fi
        TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    else
        log_warning "Checkov not installed - skipping automated security scan"
        log_info "Install with: pip install checkov"
        WARNINGS=$((WARNINGS + 1))
    fi
    
    # Validate Docker security if Dockerfile exists
    if [ -f "Dockerfile" ] || [ -f "Dockerfile.secure" ]; then
        log_info "\n6. Validating Docker Security"
        log_info "-----------------------------"
        
        dockerfile="Dockerfile"
        if [ -f "Dockerfile.secure" ]; then
            dockerfile="Dockerfile.secure"
        fi
        
        # Check for non-root user
        run_check "Docker runs as non-root user" \
            "check_file_content '$dockerfile' 'USER' 'Non-root user configured'"
        
        if command -v checkov &> /dev/null; then
            log_info "Running Checkov scan on Docker files..."
            if checkov -f "$dockerfile" --framework dockerfile --check CKV_DOCKER_* --output cli; then
                log_success "Docker Checkov scan completed successfully"
                PASSED_CHECKS=$((PASSED_CHECKS + 1))
            else
                log_error "Docker Checkov scan found security issues"
                FAILED_CHECKS=$((FAILED_CHECKS + 1))
            fi
            TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
        fi
    fi
    
    # Final results
    log_info "\n============================================"
    log_info "Security Validation Results"
    log_info "============================================"
    
    log_info "Total Checks: $TOTAL_CHECKS"
    log_success "Passed: $PASSED_CHECKS"
    log_error "Failed: $FAILED_CHECKS"
    log_warning "Warnings: $WARNINGS"
    
    # Calculate success rate
    if [ $TOTAL_CHECKS -gt 0 ]; then
        SUCCESS_RATE=$(( (PASSED_CHECKS * 100) / TOTAL_CHECKS ))
        log_info "Success Rate: ${SUCCESS_RATE}%"
        
        if [ $SUCCESS_RATE -ge 95 ]; then
            log_success "🎉 EXCELLENT: Security validation passed with ${SUCCESS_RATE}% success rate!"
        elif [ $SUCCESS_RATE -ge 80 ]; then
            log_info "✅ GOOD: Security validation passed with ${SUCCESS_RATE}% success rate"
        elif [ $SUCCESS_RATE -ge 60 ]; then
            log_warning "⚠️ FAIR: Security validation needs improvement - ${SUCCESS_RATE}% success rate"
        else
            log_error "❌ POOR: Security validation failed - ${SUCCESS_RATE}% success rate"
        fi
    fi
    
    # Exit with appropriate code
    if [ $FAILED_CHECKS -eq 0 ]; then
        log_success "All security validations passed!"
        exit 0
    else
        log_error "Security validation failed with $FAILED_CHECKS failures"
        exit 1
    fi
}

# Run main function
main "$@"