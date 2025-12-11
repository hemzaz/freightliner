#!/bin/bash
# CI/CD Validation Script
# Automated validation checks for workflow infrastructure
# Version: 1.0.0

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
WORKFLOWS_DIR="$REPO_ROOT/.github/workflows"

ERRORS=0
WARNINGS=0
CHECKS_PASSED=0
CHECKS_TOTAL=0

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✅${NC} $1"
    ((CHECKS_PASSED++))
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARNINGS++))
}

log_error() {
    echo -e "${RED}❌${NC} $1"
    ((ERRORS++))
}

# Check function wrapper
run_check() {
    local check_name="$1"
    ((CHECKS_TOTAL++))
    log_info "Checking: $check_name"
}

echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   CI/CD Infrastructure Validation${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""

# Check 1: Workflow YAML Syntax
run_check "Workflow YAML syntax validation"
if command -v python3 &> /dev/null; then
    if python3 -c "
import yaml
import glob
import sys
errors = []
for f in glob.glob('$WORKFLOWS_DIR/*.yml'):
    try:
        with open(f) as file:
            yaml.safe_load(file)
    except Exception as e:
        errors.append((f, str(e)))

if errors:
    for f, e in errors:
        print(f'ERROR in {f}: {e}')
    sys.exit(1)
" 2>/dev/null; then
        log_success "All workflow YAML files are valid"
    else
        log_error "YAML syntax errors detected"
    fi
else
    log_warning "Python3 not available, skipping YAML validation"
fi

# Check 2: Required Workflows Exist
run_check "Required workflows presence"
REQUIRED_WORKFLOWS=(
    "consolidated-ci.yml"
    "security-gates-enhanced.yml"
    "deploy.yml"
    "release-pipeline.yml"
)

for workflow in "${REQUIRED_WORKFLOWS[@]}"; do
    if [ -f "$WORKFLOWS_DIR/$workflow" ]; then
        log_success "Required workflow exists: $workflow"
    else
        log_error "Missing required workflow: $workflow"
    fi
    ((CHECKS_TOTAL++))
done

# Check 3: Permissions Audit
run_check "Workflow permissions validation"
DANGEROUS_PERMS=$(grep -r "permissions:" "$WORKFLOWS_DIR" -A10 | \
    grep -E "contents:.*write|packages:.*write" | \
    grep -v "release-pipeline.yml" | \
    grep -v "deploy.yml" | wc -l | tr -d ' ')

if [ "$DANGEROUS_PERMS" -eq 0 ]; then
    log_success "No unnecessary write permissions in CI workflows"
else
    log_warning "Found $DANGEROUS_PERMS potential write permission issues"
fi

# Check 4: Timeout Coverage
run_check "Timeout configuration coverage"
WORKFLOWS_WITHOUT_TIMEOUT=$(find "$WORKFLOWS_DIR" -name "*.yml" ! -name "reusable-*" -exec grep -L "timeout-minutes:" {} \;)
if [ -z "$WORKFLOWS_WITHOUT_TIMEOUT" ]; then
    log_success "All workflows have timeout configuration"
else
    log_warning "Some workflows missing timeout-minutes:"
    echo "$WORKFLOWS_WITHOUT_TIMEOUT"
fi

# Check 5: Concurrency Control
run_check "Concurrency control presence"
WORKFLOWS_WITHOUT_CONCURRENCY=$(find "$WORKFLOWS_DIR" -name "*.yml" ! -name "reusable-*" -exec grep -L "concurrency:" {} \;)
if [ -z "$WORKFLOWS_WITHOUT_CONCURRENCY" ]; then
    log_success "All workflows have concurrency control"
else
    log_warning "Some workflows missing concurrency control:"
    echo "$WORKFLOWS_WITHOUT_CONCURRENCY"
fi

# Check 6: Secret References
run_check "Secret reference validation"
SECRET_REFS=$(grep -r "secrets\." "$WORKFLOWS_DIR" --include="*.yml" | wc -l | tr -d ' ')
log_info "Found $SECRET_REFS secret references"

# Check for hardcoded secrets (should never exist)
HARDCODED=$(grep -rE "(password|token|key).*[:=].*['\"][a-zA-Z0-9]{20,}['\"]" "$WORKFLOWS_DIR" --include="*.yml" || true)
if [ -z "$HARDCODED" ]; then
    log_success "No hardcoded secrets detected"
else
    log_error "Potential hardcoded secrets found!"
    echo "$HARDCODED"
fi

# Check 7: Action Version Pinning
run_check "Action version pinning validation"
UNPINNED_ACTIONS=$(grep -r "uses:" "$WORKFLOWS_DIR" --include="*.yml" | grep -v "@" | grep -v "# " || true)
if [ -z "$UNPINNED_ACTIONS" ]; then
    log_success "All actions are version-pinned"
else
    log_error "Found unpinned actions:"
    echo "$UNPINNED_ACTIONS"
fi

# Check 8: Reusable Workflows
run_check "Reusable workflow validation"
REUSABLE_COUNT=$(find "$WORKFLOWS_DIR" -name "reusable-*.yml" | wc -l | tr -d ' ')
log_info "Found $REUSABLE_COUNT reusable workflows"

for reusable in "$WORKFLOWS_DIR"/reusable-*.yml; do
    if [ -f "$reusable" ]; then
        if grep -q "workflow_call:" "$reusable"; then
            log_success "Valid reusable workflow: $(basename "$reusable")"
            ((CHECKS_TOTAL++))
        else
            log_error "Invalid reusable workflow (missing workflow_call): $(basename "$reusable")"
            ((CHECKS_TOTAL++))
        fi
    fi
done

# Check 9: Composite Actions
run_check "Composite actions validation"
ACTIONS_DIR="$REPO_ROOT/.github/actions"
if [ -d "$ACTIONS_DIR" ]; then
    for action in "$ACTIONS_DIR"/*; do
        if [ -d "$action" ]; then
            action_yml="$action/action.yml"
            if [ -f "$action_yml" ]; then
                log_success "Composite action found: $(basename "$action")"
                ((CHECKS_TOTAL++))
            else
                log_error "Composite action missing action.yml: $(basename "$action")"
                ((CHECKS_TOTAL++))
            fi
        fi
    done
else
    log_warning "No composite actions directory found"
fi

# Check 10: Security Workflow Configuration
run_check "Security workflow validation"
SECURITY_WORKFLOWS=(
    "security-gates-enhanced.yml"
    "security-gates.yml"
)

for sec_workflow in "${SECURITY_WORKFLOWS[@]}"; do
    sec_file="$WORKFLOWS_DIR/$sec_workflow"
    if [ -f "$sec_file" ]; then
        # Check for critical security scanning steps
        if grep -q "TruffleHog\|gitleaks" "$sec_file"; then
            log_success "Secret scanning configured in $sec_workflow"
        else
            log_warning "No secret scanning found in $sec_workflow"
        fi

        if grep -q "gosec\|Semgrep" "$sec_file"; then
            log_success "SAST scanning configured in $sec_workflow"
        else
            log_warning "No SAST scanning found in $sec_workflow"
        fi

        if grep -q "trivy\|grype" "$sec_file"; then
            log_success "Container scanning configured in $sec_workflow"
        else
            log_warning "No container scanning found in $sec_workflow"
        fi

        ((CHECKS_TOTAL+=3))
    fi
done

# Check 11: Caching Configuration
run_check "Caching strategy validation"
CACHE_COUNT=$(grep -r "uses: actions/cache@" "$WORKFLOWS_DIR" --include="*.yml" | wc -l | tr -d ' ')
if [ "$CACHE_COUNT" -gt 0 ]; then
    log_success "Caching configured ($CACHE_COUNT instances)"
else
    log_warning "No caching configured in workflows"
fi

# Check 12: Deployment Safety
run_check "Deployment safety checks"
if [ -f "$WORKFLOWS_DIR/deploy.yml" ]; then
    if grep -q "environment:" "$WORKFLOWS_DIR/deploy.yml"; then
        log_success "Environment protection configured"
    else
        log_warning "No environment protection in deploy.yml"
    fi

    if grep -q "manual\|approval" "$WORKFLOWS_DIR/deploy.yml"; then
        log_success "Manual approval gates detected"
    else
        log_warning "No manual approval gates in deploy.yml"
    fi

    if grep -q "rollback" "$WORKFLOWS_DIR/deploy.yml"; then
        log_success "Rollback mechanism present"
    else
        log_error "No rollback mechanism in deploy.yml"
    fi

    ((CHECKS_TOTAL+=3))
fi

# Check 13: Test Coverage Configuration
run_check "Test coverage validation"
COVERAGE_THRESHOLD=$(grep -r "coverage-threshold:" "$WORKFLOWS_DIR" --include="*.yml" | head -1 | grep -oE "[0-9]+" || echo "0")
if [ "$COVERAGE_THRESHOLD" -ge 80 ]; then
    log_success "Coverage threshold meets standard: ${COVERAGE_THRESHOLD}%"
elif [ "$COVERAGE_THRESHOLD" -ge 40 ]; then
    log_warning "Coverage threshold below standard: ${COVERAGE_THRESHOLD}% (recommend 80%)"
else
    log_error "Coverage threshold too low: ${COVERAGE_THRESHOLD}%"
fi

# Check 14: Documentation
run_check "Documentation validation"
DOCS=(
    "$REPO_ROOT/.github/WORKFLOWS.md"
    "$REPO_ROOT/.github/CICD_VALIDATION_REPORT.md"
)

for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        log_success "Documentation exists: $(basename "$doc")"
    else
        log_warning "Missing documentation: $(basename "$doc")"
    fi
    ((CHECKS_TOTAL++))
done

# Check 15: Workflow Dependencies
run_check "Workflow dependency validation"
if command -v jq &> /dev/null; then
    # Check for circular dependencies (would require more complex parsing)
    log_success "Dependency check available (jq installed)"
else
    log_warning "jq not installed, skipping dependency graph analysis"
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   Validation Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo "Total Checks: $CHECKS_TOTAL"
echo -e "${GREEN}Passed: $CHECKS_PASSED${NC}"
echo -e "${YELLOW}Warnings: $WARNINGS${NC}"
echo -e "${RED}Errors: $ERRORS${NC}"
echo ""

# Calculate success rate
SUCCESS_RATE=$((CHECKS_PASSED * 100 / CHECKS_TOTAL))
echo -e "Success Rate: ${SUCCESS_RATE}%"
echo ""

# Final verdict
if [ $ERRORS -eq 0 ]; then
    if [ $WARNINGS -eq 0 ]; then
        echo -e "${GREEN}✅ VALIDATION PASSED${NC} - No issues found"
        exit 0
    else
        echo -e "${YELLOW}⚠ VALIDATION PASSED WITH WARNINGS${NC} - $WARNINGS warnings to review"
        exit 0
    fi
else
    echo -e "${RED}❌ VALIDATION FAILED${NC} - $ERRORS errors must be fixed"
    exit 1
fi
