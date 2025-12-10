#!/bin/bash

# SECURITY COMPLIANCE VALIDATION SCRIPT
# Validates all security implementations and compliance standards
# Usage: ./scripts/security-compliance-check.sh [--fix] [--report]

set -euo pipefail

# SECURITY: Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPLIANCE_REPORT="${PROJECT_ROOT}/security-compliance-report.json"
FIX_ISSUES=false
GENERATE_REPORT=false

# SECURITY: Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# SECURITY: Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --fix)
            FIX_ISSUES=true
            shift
            ;;
        --report)
            GENERATE_REPORT=true
            shift
            ;;
        --help)
            echo "Usage: $0 [--fix] [--report]"
            echo "  --fix      Attempt to fix compliance issues automatically"
            echo "  --report   Generate detailed compliance report"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# SECURITY: Logging functions
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

# SECURITY: Initialize compliance tracking
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
COMPLIANCE_ISSUES=()

# SECURITY: Function to track compliance check
track_check() {
    local check_name="$1"
    local status="$2"
    local details="${3:-}"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    if [[ "$status" == "PASS" ]]; then
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        log_success "$check_name: PASSED"
    else
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        log_error "$check_name: FAILED - $details"
        COMPLIANCE_ISSUES+=("$check_name: $details")
    fi
}

# SECURITY: Check for security workflow files
check_security_workflows() {
    log_info "Checking security workflow implementations..."
    
    local required_workflows=(
        ".github/workflows/ci-secure.yml"
        ".github/workflows/security-gates-enhanced.yml"
        ".github/workflows/security-monitoring-enhanced.yml"
        ".github/workflows/oidc-authentication.yml"
    )
    
    for workflow in "${required_workflows[@]}"; do
        if [[ -f "$PROJECT_ROOT/$workflow" ]]; then
            track_check "Security workflow: $(basename "$workflow")" "PASS"
        else
            track_check "Security workflow: $(basename "$workflow")" "FAIL" "Workflow file missing"
            
            if [[ "$FIX_ISSUES" == "true" ]]; then
                log_warning "Auto-fix not available for missing workflows"
            fi
        fi
    done
}

# SECURITY: Check for shell injection vulnerabilities
check_shell_injection() {
    log_info "Checking for shell injection vulnerabilities..."
    
    # SECURITY: Search for dangerous patterns in workflows
    local dangerous_patterns=(
        "github\.event\.head_commit\.message"
        "github\.event\.pull_request\.title"
        "github\.event\.pull_request\.body" 
        "github\.event\.comment\.body"
        "\$\{\{.*github\.event\..*\}\}.*run:"
    )
    
    local injection_found=false
    
    for pattern in "${dangerous_patterns[@]}"; do
        if find "$PROJECT_ROOT/.github/workflows" -name "*.yml" -exec grep -l "$pattern" {} \; 2>/dev/null | grep -v "ci-secure.yml\|security-"; then
            injection_found=true
            track_check "Shell injection check: $pattern" "FAIL" "Dangerous pattern found in workflows"
        fi
    done
    
    if [[ "$injection_found" == "false" ]]; then
        track_check "Shell injection vulnerabilities" "PASS"
    fi
}

# SECURITY: Check secret scanning configuration
check_secret_scanning() {
    log_info "Checking secret scanning configuration..."
    
    # SECURITY: Check for GitLeaks configuration
    if [[ -f "$PROJECT_ROOT/.gitleaks.toml" ]]; then
        track_check "GitLeaks configuration" "PASS"
    else
        track_check "GitLeaks configuration" "FAIL" "GitLeaks config file missing"
    fi
    
    # SECURITY: Check for secrets in codebase
    if command -v gitleaks >/dev/null 2>&1; then
        cd "$PROJECT_ROOT"
        if gitleaks detect --source . --config .gitleaks.toml --no-git --quiet; then
            track_check "Secret detection scan" "PASS"
        else
            track_check "Secret detection scan" "FAIL" "Potential secrets detected"
        fi
    else
        log_warning "GitLeaks not installed - skipping secret detection scan"
    fi
}

# SECURITY: Check container security
check_container_security() {
    log_info "Checking container security configuration..."
    
    # SECURITY: Check for security-hardened Dockerfile
    if [[ -f "$PROJECT_ROOT/Dockerfile.secure" ]]; then
        track_check "Security-hardened Dockerfile" "PASS"
        
        # SECURITY: Validate Dockerfile security practices
        local dockerfile_issues=0
        
        # Check for root user
        if grep -q "^USER root" "$PROJECT_ROOT/Dockerfile.secure"; then
            dockerfile_issues=$((dockerfile_issues + 1))
            track_check "Dockerfile: Non-root user" "FAIL" "Running as root user"
        else
            track_check "Dockerfile: Non-root user" "PASS"
        fi
        
        # Check for latest tags
        if grep -q ":latest" "$PROJECT_ROOT/Dockerfile.secure"; then
            dockerfile_issues=$((dockerfile_issues + 1)) 
            track_check "Dockerfile: Pinned versions" "FAIL" "Using latest tags"
        else
            track_check "Dockerfile: Pinned versions" "PASS"
        fi
        
    else
        track_check "Security-hardened Dockerfile" "FAIL" "Dockerfile.secure missing"
    fi
    
    # SECURITY: Check for container scanning in workflows
    if grep -q "trivy\|grype\|anchore" "$PROJECT_ROOT/.github/workflows/"*.yml 2>/dev/null; then
        track_check "Container vulnerability scanning" "PASS"
    else
        track_check "Container vulnerability scanning" "FAIL" "No container scanning configured"
    fi
}

# SECURITY: Check OIDC authentication setup
check_oidc_authentication() {
    log_info "Checking OIDC authentication configuration..."
    
    # SECURITY: Check for OIDC workflow
    if [[ -f "$PROJECT_ROOT/.github/workflows/oidc-authentication.yml" ]]; then
        track_check "OIDC authentication workflow" "PASS"
        
        # SECURITY: Check for proper permissions
        if grep -q "id-token: write" "$PROJECT_ROOT/.github/workflows/oidc-authentication.yml"; then
            track_check "OIDC: Token permissions" "PASS"
        else
            track_check "OIDC: Token permissions" "FAIL" "Missing id-token: write permission"
        fi
        
        # SECURITY: Check for multi-cloud support
        if grep -q "aws-actions/configure-aws-credentials\|google-github-actions/auth\|azure/login" "$PROJECT_ROOT/.github/workflows/oidc-authentication.yml"; then
            track_check "OIDC: Multi-cloud support" "PASS"
        else
            track_check "OIDC: Multi-cloud support" "FAIL" "Limited cloud provider support"
        fi
        
    else
        track_check "OIDC authentication workflow" "FAIL" "OIDC workflow missing"
    fi
}

# SECURITY: Check dependency security
check_dependency_security() {
    log_info "Checking dependency security configuration..."
    
    # SECURITY: Check for Go vulnerability scanning
    if command -v govulncheck >/dev/null 2>&1; then
        cd "$PROJECT_ROOT"
        if govulncheck ./...; then
            track_check "Go vulnerability check" "PASS"
        else
            track_check "Go vulnerability check" "FAIL" "Vulnerabilities found in dependencies"
        fi
    else
        log_warning "govulncheck not installed - skipping dependency vulnerability check"
    fi
    
    # SECURITY: Check for dependency scanning in workflows
    if grep -q "govulncheck\|nancy\|snyk" "$PROJECT_ROOT/.github/workflows/"*.yml 2>/dev/null; then
        track_check "Dependency scanning workflows" "PASS"
    else
        track_check "Dependency scanning workflows" "FAIL" "No dependency scanning in workflows"
    fi
}

# SECURITY: Check access controls and permissions
check_access_controls() {
    log_info "Checking access controls and permissions..."
    
    # SECURITY: Check workflow permissions
    local workflows_with_minimal_perms=0
    local total_workflows=0
    
    for workflow in "$PROJECT_ROOT/.github/workflows/"*.yml; do
        if [[ -f "$workflow" ]]; then
            total_workflows=$((total_workflows + 1))
            
            # SECURITY: Check for explicit permissions
            if grep -q "permissions:" "$workflow"; then
                workflows_with_minimal_perms=$((workflows_with_minimal_perms + 1))
            fi
        fi
    done
    
    if [[ $workflows_with_minimal_perms -eq $total_workflows ]] && [[ $total_workflows -gt 0 ]]; then
        track_check "Workflow permissions" "PASS"
    else
        track_check "Workflow permissions" "FAIL" "Some workflows missing explicit permissions"
    fi
}

# SECURITY: Check security monitoring and alerting
check_security_monitoring() {
    log_info "Checking security monitoring and alerting..."
    
    # SECURITY: Check for security monitoring workflow
    if [[ -f "$PROJECT_ROOT/.github/workflows/security-monitoring-enhanced.yml" ]]; then
        track_check "Security monitoring workflow" "PASS"
        
        # SECURITY: Check for scheduled scans
        if grep -q "schedule:" "$PROJECT_ROOT/.github/workflows/security-monitoring-enhanced.yml"; then
            track_check "Scheduled security scans" "PASS"
        else
            track_check "Scheduled security scans" "FAIL" "No scheduled scans configured"
        fi
        
        # SECURITY: Check for alerting configuration
        if grep -q "SLACK_WEBHOOK\|TEAMS_WEBHOOK\|EMAIL_ENDPOINT" "$PROJECT_ROOT/.github/workflows/security-monitoring-enhanced.yml"; then
            track_check "Security alerting configuration" "PASS"
        else
            track_check "Security alerting configuration" "FAIL" "No alerting endpoints configured"
        fi
        
    else
        track_check "Security monitoring workflow" "FAIL" "Security monitoring workflow missing"
    fi
}

# SECURITY: Check compliance with security standards
check_compliance_standards() {
    log_info "Checking compliance with security standards..."
    
    # SECURITY: OWASP CI/CD Security Top 10 compliance
    local owasp_controls=(
        "Flow Control:security-gates"
        "Identity and Access:oidc-authentication"
        "Dependency Chain:dependency-scanning" 
        "Pipeline Execution:shell-injection-prevention"
        "Access Controls:minimal-permissions"
        "Credential Hygiene:secret-scanning"
        "System Configuration:container-hardening"
        "Third Party Services:vendor-validation"
        "Artifact Integrity:checksum-validation"
        "Logging and Visibility:security-monitoring"
    )
    
    for control in "${owasp_controls[@]}"; do
        local control_name="${control%%:*}"
        local check_pattern="${control##*:}"
        
        if grep -r "$check_pattern" "$PROJECT_ROOT/.github/workflows/" >/dev/null 2>&1; then
            track_check "OWASP CI/CD: $control_name" "PASS"
        else
            track_check "OWASP CI/CD: $control_name" "FAIL" "Control not implemented"
        fi
    done
}

# SECURITY: Generate compliance report
generate_compliance_report() {
    log_info "Generating compliance report..."
    
    local compliance_percentage=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
    local compliance_status="NON_COMPLIANT"
    
    if [[ $compliance_percentage -eq 100 ]]; then
        compliance_status="FULLY_COMPLIANT"
    elif [[ $compliance_percentage -ge 90 ]]; then
        compliance_status="SUBSTANTIALLY_COMPLIANT"
    elif [[ $compliance_percentage -ge 70 ]]; then
        compliance_status="PARTIALLY_COMPLIANT"
    fi
    
    # SECURITY: Create JSON report
    cat > "$COMPLIANCE_REPORT" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "compliance_status": "$compliance_status",
  "compliance_percentage": $compliance_percentage,
  "total_checks": $TOTAL_CHECKS,
  "passed_checks": $PASSED_CHECKS,
  "failed_checks": $FAILED_CHECKS,
  "security_score": $compliance_percentage,
  "issues": [
$(IFS=$'\n'; for issue in "${COMPLIANCE_ISSUES[@]}"; do echo "    \"$issue\","; done | sed '$ s/,$//')
  ],
  "recommendations": [
    "Fix all failed compliance checks",
    "Implement missing security controls",
    "Regular security compliance validation",
    "Continuous security monitoring"
  ]
}
EOF
    
    log_success "Compliance report generated: $COMPLIANCE_REPORT"
}

# SECURITY: Main execution function
main() {
    echo -e "${BLUE}🛡️ SECURITY COMPLIANCE VALIDATION${NC}"
    echo "================================================"
    echo "Project: Freightliner Container Registry Replication"
    echo "Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "Fix Issues: $FIX_ISSUES"
    echo "Generate Report: $GENERATE_REPORT"
    echo "================================================"
    
    # SECURITY: Run all compliance checks
    check_security_workflows
    check_shell_injection  
    check_secret_scanning
    check_container_security
    check_oidc_authentication
    check_dependency_security
    check_access_controls
    check_security_monitoring
    check_compliance_standards
    
    # SECURITY: Calculate final compliance score
    local compliance_percentage=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
    
    echo ""
    echo "================================================"
    echo -e "${BLUE}📊 COMPLIANCE SUMMARY${NC}"
    echo "================================================"
    echo "Total Checks: $TOTAL_CHECKS"
    echo "Passed: $PASSED_CHECKS"
    echo "Failed: $FAILED_CHECKS"
    echo "Compliance: $compliance_percentage%"
    echo ""
    
    # SECURITY: Display compliance status
    if [[ $compliance_percentage -eq 100 ]]; then
        echo -e "${GREEN}✅ FULLY COMPLIANT - PRODUCTION READY${NC}"
        echo "🎉 All security requirements met!"
    elif [[ $compliance_percentage -ge 90 ]]; then
        echo -e "${YELLOW}⚠️ SUBSTANTIALLY COMPLIANT${NC}"
        echo "🔍 Minor issues detected - review and fix"
    elif [[ $compliance_percentage -ge 70 ]]; then
        echo -e "${YELLOW}⚠️ PARTIALLY COMPLIANT${NC}"
        echo "🚨 Multiple issues detected - remediation required"
    else
        echo -e "${RED}❌ NON-COMPLIANT${NC}"
        echo "🚨 Critical security issues - immediate action required"
    fi
    
    # SECURITY: Display failed checks
    if [[ $FAILED_CHECKS -gt 0 ]]; then
        echo ""
        echo -e "${RED}📋 COMPLIANCE ISSUES:${NC}"
        for issue in "${COMPLIANCE_ISSUES[@]}"; do
            echo "  • $issue"
        done
    fi
    
    # SECURITY: Generate report if requested
    if [[ "$GENERATE_REPORT" == "true" ]]; then
        generate_compliance_report
    fi
    
    echo ""
    echo "================================================"
    
    # SECURITY: Exit with appropriate code
    if [[ $FAILED_CHECKS -eq 0 ]]; then
        log_success "Security compliance validation completed successfully"
        exit 0
    else
        log_error "Security compliance validation failed - $FAILED_CHECKS issues found"
        exit 1
    fi
}

# SECURITY: Execute main function
main "$@"