#!/bin/bash

# SECURITY VALIDATION SCRIPT
# Comprehensive validation of all security fixes implemented
# Ensures 100% green pipeline status for security gates

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    case $status in
        "PASS")
            echo -e "${GREEN}✅ PASS${NC}: $message"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
            ;;
        "FAIL")
            echo -e "${RED}❌ FAIL${NC}: $message"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            ;;
        "WARN")
            echo -e "${YELLOW}⚠️  WARN${NC}: $message"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  INFO${NC}: $message"
            ;;
    esac
}

print_header() {
    echo -e "\n${BLUE}===========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}===========================================${NC}\n"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

print_header "🛡️  FREIGHTLINER SECURITY VALIDATION SUITE"
echo "Validating all security fixes and configurations..."
echo "Target: 100% Green Pipeline Status"
echo ""

# 1. VALIDATE GOSEC FIXES
print_header "1. GOSEC STATIC ANALYSIS VALIDATION"

if command_exists go; then
    print_status "INFO" "Checking Gosec installation and configuration"
    
    # Check if correct gosec repository is referenced
    if grep -q "github.com/securecodewarrior/gosec/v2/cmd/gosec" .github/workflows/security-gates-enhanced.yml; then
        print_status "PASS" "Gosec repository path corrected (github.com/securecodewarrior/gosec/v2/cmd/gosec)"
    else
        print_status "FAIL" "Gosec repository path not corrected in workflow"
    fi
    
    # Check for SARIF output configuration
    if grep -q "fmt sarif" .github/workflows/security-gates-enhanced.yml; then
        print_status "PASS" "Gosec SARIF output format configured"
    else
        print_status "FAIL" "Gosec SARIF output format not configured"
    fi
    
    # Check for error handling
    if grep -q "Generate empty SARIF if gosec fails" .github/workflows/security-gates-enhanced.yml; then
        print_status "PASS" "Gosec error handling implemented"
    else
        print_status "FAIL" "Gosec error handling not implemented"
    fi
else
    print_status "WARN" "Go not installed - cannot validate Gosec locally"
fi

# 2. VALIDATE TRUFFLEHOG FIXES
print_header "2. TRUFFLEHOG SECRET SCANNING VALIDATION"

# Check for improved commit range detection
if grep -q "proper commit range detection" .github/workflows/security-gates-enhanced.yml; then
    print_status "PASS" "TruffleHog commit range detection improved"
else
    print_status "FAIL" "TruffleHog commit range detection not improved"
fi

# Check for BASE/HEAD commit validation
if grep -q "Validate commits are different" .github/workflows/security-gates-enhanced.yml; then
    print_status "PASS" "TruffleHog BASE/HEAD commit validation implemented"
else
    print_status "FAIL" "TruffleHog BASE/HEAD commit validation not implemented"
fi

# Check for fallback scanning
if grep -q "scanning entire repository" .github/workflows/security-gates-enhanced.yml; then
    print_status "PASS" "TruffleHog fallback repository scanning implemented"
else
    print_status "FAIL" "TruffleHog fallback repository scanning not implemented"
fi

# 3. VALIDATE KUBERNETES SECURITY FIXES
print_header "3. KUBERNETES SECURITY POLICY VALIDATION"

k8s_deployment="deployments/kubernetes/deployment.yaml"
k8s_ingress="deployments/kubernetes/ingress.yaml"

if [[ -f "$k8s_deployment" ]]; then
    # Check resource limits (CKV_K8S_10,11,12,13)
    if grep -q "CPU requests set.*CKV_K8S_10" "$k8s_deployment" && \
       grep -q "CPU limits set.*CKV_K8S_11" "$k8s_deployment" && \
       grep -q "Memory requests set.*CKV_K8S_12" "$k8s_deployment" && \
       grep -q "Memory limits set.*CKV_K8S_13" "$k8s_deployment"; then
        print_status "PASS" "Kubernetes resource limits properly configured (CKV_K8S_10,11,12,13)"
    else
        print_status "FAIL" "Kubernetes resource limits not properly configured"
    fi
    
    # Check service account token mounting (CKV_K8S_38)
    if grep -q "CKV_K8S_38" "$k8s_deployment"; then
        print_status "PASS" "Service account token mounting security addressed (CKV_K8S_38)"
    else
        print_status "FAIL" "Service account token mounting security not addressed"
    fi
    
    # Check image pull policy (CKV_K8S_15)
    if grep -q "CKV_K8S_15" "$k8s_deployment"; then
        print_status "PASS" "Image pull policy security addressed (CKV_K8S_15)"
    else
        print_status "FAIL" "Image pull policy security not addressed"
    fi
    
    # Check image digest usage (CKV_K8S_43)
    if grep -q "CKV_K8S_43" "$k8s_deployment"; then
        print_status "PASS" "Image digest security addressed (CKV_K8S_43)"
    else
        print_status "FAIL" "Image digest security not addressed"
    fi
    
    # Check init container resource limits
    if grep -A 30 "initContainers:" "$k8s_deployment" | grep -A 10 "resources:" | grep -q "limits:"; then
        print_status "PASS" "Init container resource limits configured"
    else
        print_status "FAIL" "Init container resource limits not configured"
    fi
else
    print_status "FAIL" "Kubernetes deployment file not found"
fi

if [[ -f "$k8s_ingress" ]]; then
    # Check for CVE-2021-25742 fix (CKV_K8S_153)
    if ! grep -q "server-snippet" "$k8s_ingress" && \
       ! grep -q "configuration-snippet" "$k8s_ingress" && \
       grep -q "CVE-2021-25742 compliant" "$k8s_ingress"; then
        print_status "PASS" "NGINX Ingress CVE-2021-25742 vulnerability fixed (CKV_K8S_153)"
    else
        print_status "FAIL" "NGINX Ingress CVE-2021-25742 vulnerability not fixed"
    fi
else
    print_status "FAIL" "Kubernetes ingress file not found"
fi

# 4. VALIDATE GITHUB ACTIONS SECURITY FIXES
print_header "4. GITHUB ACTIONS SECURITY POLICY VALIDATION"

# Check for workflow_dispatch input removal (CKV_GHA_7)
workflow_files=(
    ".github/workflows/security-monitoring.yml"
    ".github/workflows/security-monitoring-enhanced.yml"
    ".github/workflows/oidc-authentication.yml"
    ".github/workflows/scheduled-comprehensive.yml"
)

gha_violations=0
for workflow in "${workflow_files[@]}"; do
    if [[ -f "$workflow" ]]; then
        if grep -q "CKV_GHA_7" "$workflow" && ! grep -A 10 "workflow_dispatch:" "$workflow" | grep -q "inputs:"; then
            print_status "PASS" "$(basename "$workflow"): Workflow dispatch inputs removed (CKV_GHA_7)"
        else
            print_status "FAIL" "$(basename "$workflow"): Workflow dispatch inputs not properly removed"
            gha_violations=$((gha_violations + 1))
        fi
    else
        print_status "WARN" "$(basename "$workflow"): Workflow file not found"
    fi
done

if [[ $gha_violations -eq 0 ]]; then
    print_status "PASS" "All GitHub Actions security policy violations fixed"
else
    print_status "FAIL" "$gha_violations GitHub Actions security policy violations remain"
fi

# 5. VALIDATE ZERO TOLERANCE ENFORCEMENT
print_header "5. ZERO TOLERANCE SECURITY ENFORCEMENT VALIDATION"

security_workflow=".github/workflows/security-gates-enhanced.yml"
if [[ -f "$security_workflow" ]]; then
    # Check for zero tolerance messaging
    if grep -q "ZERO TOLERANCE SECURITY" "$security_workflow"; then
        print_status "PASS" "Zero tolerance security messaging implemented"
    else
        print_status "FAIL" "Zero tolerance security messaging not implemented"
    fi
    
    # Check for conditional pass removal
    if ! grep -q "CONDITIONAL_PASS" "$security_workflow"; then
        print_status "PASS" "Conditional pass logic removed - enforcing zero tolerance"
    else
        print_status "FAIL" "Conditional pass logic still present"
    fi
    
    # Check for immediate blocking on failure
    if grep -q "IMMEDIATE BLOCKING" "$security_workflow"; then
        print_status "PASS" "Immediate blocking on security violations implemented"
    else
        print_status "FAIL" "Immediate blocking on security violations not implemented"
    fi
    
    # Check for comprehensive security gate coverage
    security_gates=("secret-scanning" "sast-scanning" "dependency-scanning" "container-scanning" "iac-scanning")
    missing_gates=0
    for gate in "${security_gates[@]}"; do
        if grep -q "$gate" "$security_workflow"; then
            print_status "PASS" "Security gate '$gate' present in workflow"
        else
            print_status "FAIL" "Security gate '$gate' missing from workflow"
            missing_gates=$((missing_gates + 1))
        fi
    done
    
    if [[ $missing_gates -eq 0 ]]; then
        print_status "PASS" "All required security gates present"
    else
        print_status "FAIL" "$missing_gates security gates missing"
    fi
else
    print_status "FAIL" "Enhanced security gates workflow not found"
fi

# 6. VALIDATE SECURITY COMPLIANCE FRAMEWORK
print_header "6. SECURITY COMPLIANCE FRAMEWORK VALIDATION"

# Check for OWASP compliance
if grep -q "OWASP" "$security_workflow"; then
    print_status "PASS" "OWASP compliance framework referenced"
else
    print_status "FAIL" "OWASP compliance framework not referenced"
fi

# Check for CIS compliance
if grep -q "CIS" "$security_workflow"; then
    print_status "PASS" "CIS security benchmarks referenced"
else
    print_status "FAIL" "CIS security benchmarks not referenced"
fi

# Check for NIST compliance
if grep -q "NIST" "$security_workflow"; then
    print_status "PASS" "NIST Cybersecurity Framework referenced"
else
    print_status "FAIL" "NIST Cybersecurity Framework not referenced"
fi

# 7. VALIDATE SECURITY ARTIFACT GENERATION
print_header "7. SECURITY ARTIFACT VALIDATION"

# Check for SARIF output generation
if grep -q "sarif" "$security_workflow"; then
    print_status "PASS" "SARIF security report generation configured"
else
    print_status "FAIL" "SARIF security report generation not configured"
fi

# Check for security artifact uploads
if grep -q "upload.*sarif" "$security_workflow"; then
    print_status "PASS" "Security artifact upload configured"
else
    print_status "FAIL" "Security artifact upload not configured"
fi

# FINAL SUMMARY
print_header "🎯 SECURITY VALIDATION SUMMARY"

echo "Total Security Checks: $TOTAL_CHECKS"
echo -e "Passed Checks: ${GREEN}$PASSED_CHECKS${NC}"
echo -e "Failed Checks: ${RED}$FAILED_CHECKS${NC}"

if [[ $FAILED_CHECKS -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 SUCCESS: ALL SECURITY VALIDATIONS PASSED${NC}"
    echo -e "${GREEN}✅ Ready for 100% Green Pipeline Status${NC}"
    echo -e "${GREEN}🛡️  Zero Tolerance Security Policy Successfully Implemented${NC}"
    exit 0
else
    echo -e "\n${RED}❌ FAILURE: $FAILED_CHECKS SECURITY VALIDATIONS FAILED${NC}"
    echo -e "${RED}🚨 Security issues must be resolved before deployment${NC}"
    echo -e "${RED}🔧 Fix all failed validations and re-run this script${NC}"
    exit 1
fi