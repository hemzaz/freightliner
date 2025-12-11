#!/bin/bash
# Minimal Workflow Verification Script
# This script verifies the minimal workflows are correctly configured

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKFLOW_DIR="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../../../.." && pwd)"

echo "=========================================="
echo "Minimal Workflow Verification"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
WARNINGS=0

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED++))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED++))
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARNINGS++))
}

# Check 1: Verify workflow files exist
echo "1. Checking workflow files..."
if [ -f "$SCRIPT_DIR/ci.yml" ]; then
    pass "ci.yml exists"
else
    fail "ci.yml not found"
fi

if [ -f "$SCRIPT_DIR/deploy.yml" ]; then
    pass "deploy.yml exists"
else
    fail "deploy.yml not found"
fi

if [ -f "$SCRIPT_DIR/scheduled.yml" ]; then
    pass "scheduled.yml exists"
else
    fail "scheduled.yml not found"
fi
echo ""

# Check 2: Verify documentation files exist
echo "2. Checking documentation files..."
if [ -f "$SCRIPT_DIR/README.md" ]; then
    pass "README.md exists"
else
    fail "README.md not found"
fi

if [ -f "$SCRIPT_DIR/MIGRATION.md" ]; then
    pass "MIGRATION.md exists"
else
    fail "MIGRATION.md not found"
fi

if [ -f "$SCRIPT_DIR/IMPLEMENTATION_CHECKLIST.md" ]; then
    pass "IMPLEMENTATION_CHECKLIST.md exists"
else
    fail "IMPLEMENTATION_CHECKLIST.md not found"
fi

if [ -f "$SCRIPT_DIR/SUMMARY.md" ]; then
    pass "SUMMARY.md exists"
else
    fail "SUMMARY.md not found"
fi
echo ""

# Check 3: Verify workflow syntax
echo "3. Checking workflow syntax..."
if command -v yamllint &> /dev/null; then
    for file in ci.yml deploy.yml scheduled.yml; do
        if yamllint -d relaxed "$SCRIPT_DIR/$file" &> /dev/null; then
            pass "$file syntax valid"
        else
            fail "$file syntax invalid"
        fi
    done
else
    warn "yamllint not installed, skipping syntax check"
fi
echo ""

# Check 4: Verify workflow structure
echo "4. Checking workflow structure..."

# Check ci.yml
if grep -q "name: CI - The One Pipeline to Rule Them All" "$SCRIPT_DIR/ci.yml"; then
    pass "ci.yml has correct name"
else
    fail "ci.yml name incorrect"
fi

if grep -q "jobs:" "$SCRIPT_DIR/ci.yml"; then
    pass "ci.yml has jobs"
else
    fail "ci.yml missing jobs"
fi

# Check deploy.yml
if grep -q "name: Deploy - Universal Multi-Environment Deployment" "$SCRIPT_DIR/deploy.yml"; then
    pass "deploy.yml has correct name"
else
    fail "deploy.yml name incorrect"
fi

if grep -q "jobs:" "$SCRIPT_DIR/deploy.yml"; then
    pass "deploy.yml has jobs"
else
    fail "deploy.yml missing jobs"
fi

# Check scheduled.yml
if grep -q "name: Scheduled - Nightly Comprehensive Tasks" "$SCRIPT_DIR/scheduled.yml"; then
    pass "scheduled.yml has correct name"
else
    fail "scheduled.yml name incorrect"
fi

if grep -q "jobs:" "$SCRIPT_DIR/scheduled.yml"; then
    pass "scheduled.yml has jobs"
else
    fail "scheduled.yml missing jobs"
fi
echo ""

# Check 5: Verify key jobs exist
echo "5. Checking job definitions..."

# ci.yml jobs
for job in lint test security-quick build docker; do
    if grep -q "$job:" "$SCRIPT_DIR/ci.yml"; then
        pass "ci.yml has $job job"
    else
        fail "ci.yml missing $job job"
    fi
done

# deploy.yml jobs
for job in validate deploy-dev deploy-staging deploy-production; do
    if grep -q "$job:" "$SCRIPT_DIR/deploy.yml"; then
        pass "deploy.yml has $job job"
    else
        fail "deploy.yml missing $job job"
    fi
done

# scheduled.yml jobs
for job in security-comprehensive dependency-updates performance-benchmarks cleanup; do
    if grep -q "$job:" "$SCRIPT_DIR/scheduled.yml"; then
        pass "scheduled.yml has $job job"
    else
        fail "scheduled.yml missing $job job"
    fi
done
echo ""

# Check 6: Verify triggers
echo "6. Checking workflow triggers..."

if grep -q "on:" "$SCRIPT_DIR/ci.yml"; then
    pass "ci.yml has triggers"
else
    fail "ci.yml missing triggers"
fi

if grep -q "workflow_dispatch:" "$SCRIPT_DIR/ci.yml"; then
    pass "ci.yml has manual trigger"
else
    fail "ci.yml missing manual trigger"
fi

if grep -q "workflow_dispatch:" "$SCRIPT_DIR/deploy.yml"; then
    pass "deploy.yml has manual trigger"
else
    fail "deploy.yml missing manual trigger"
fi

if grep -q "schedule:" "$SCRIPT_DIR/scheduled.yml"; then
    pass "scheduled.yml has schedule trigger"
else
    fail "scheduled.yml missing schedule trigger"
fi
echo ""

# Check 7: Verify environment variables
echo "7. Checking environment variables..."

for file in ci.yml deploy.yml scheduled.yml; do
    if grep -q "GO_VERSION:" "$SCRIPT_DIR/$file" || grep -q "env:" "$SCRIPT_DIR/$file"; then
        pass "$file has environment variables"
    else
        warn "$file may be missing environment variables"
    fi
done
echo ""

# Check 8: Verify permissions
echo "8. Checking permissions..."

for file in ci.yml deploy.yml scheduled.yml; do
    if grep -q "permissions:" "$SCRIPT_DIR/$file"; then
        pass "$file has permissions defined"
    else
        warn "$file may be missing permissions"
    fi
done
echo ""

# Check 9: Verify timeout settings
echo "9. Checking timeout settings..."

for file in ci.yml deploy.yml scheduled.yml; do
    if grep -q "timeout-minutes:" "$SCRIPT_DIR/$file"; then
        pass "$file has timeout settings"
    else
        warn "$file may be missing timeout settings"
    fi
done
echo ""

# Check 10: Verify concurrency settings
echo "10. Checking concurrency settings..."

for file in ci.yml deploy.yml scheduled.yml; do
    if grep -q "concurrency:" "$SCRIPT_DIR/$file"; then
        pass "$file has concurrency control"
    else
        warn "$file may be missing concurrency control"
    fi
done
echo ""

# Check 11: Verify old workflows
echo "11. Checking old workflows..."

OLD_WORKFLOWS=(
    "consolidated-ci.yml"
    "consolidated-ci-v2.yml"
    "test-matrix.yml"
    "integration-tests.yml"
    "security-scan.yml"
    "deploy.yml"
    "deploy-unified.yml"
    "kubernetes-deploy.yml"
    "security-comprehensive.yml"
)

FOUND_OLD=0
for old_workflow in "${OLD_WORKFLOWS[@]}"; do
    if [ -f "$WORKFLOW_DIR/$old_workflow" ] && [ "$WORKFLOW_DIR/$old_workflow" != "$SCRIPT_DIR/"* ]; then
        ((FOUND_OLD++))
    fi
done

if [ $FOUND_OLD -gt 0 ]; then
    warn "Found $FOUND_OLD old workflow(s) - migration not complete"
else
    pass "No conflicting old workflows found"
fi
echo ""

# Check 12: Verify GitHub CLI
echo "12. Checking GitHub CLI..."

if command -v gh &> /dev/null; then
    pass "GitHub CLI installed"

    # Check if authenticated
    if gh auth status &> /dev/null; then
        pass "GitHub CLI authenticated"
    else
        warn "GitHub CLI not authenticated"
    fi
else
    warn "GitHub CLI not installed (optional for manual testing)"
fi
echo ""

# Check 13: File sizes
echo "13. Checking file sizes..."

for file in ci.yml deploy.yml scheduled.yml; do
    size=$(wc -c < "$SCRIPT_DIR/$file")
    if [ $size -gt 0 ]; then
        pass "$file size: $size bytes"
    else
        fail "$file is empty"
    fi
done
echo ""

# Check 14: Line counts
echo "14. Checking line counts..."

for file in ci.yml deploy.yml scheduled.yml; do
    lines=$(wc -l < "$SCRIPT_DIR/$file")
    if [ $lines -gt 50 ]; then
        pass "$file has $lines lines (sufficient)"
    else
        warn "$file has only $lines lines (may be incomplete)"
    fi
done
echo ""

# Summary
echo "=========================================="
echo "Verification Summary"
echo "=========================================="
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${YELLOW}Warnings:${NC} $WARNINGS"
echo -e "${RED}Failed:${NC} $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All critical checks passed!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Review the workflow files in: $SCRIPT_DIR"
    echo "2. Read the documentation: README.md, MIGRATION.md"
    echo "3. Follow the implementation checklist: IMPLEMENTATION_CHECKLIST.md"
    echo "4. Test the workflows before deploying"
    echo ""
    exit 0
else
    echo -e "${RED}✗ Some checks failed. Please review and fix issues.${NC}"
    echo ""
    exit 1
fi
