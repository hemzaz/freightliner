#!/bin/bash

# Week 1 CICD Optimization Validation Script
# Purpose: Automated validation of Session 4 optimizations
# Usage: ./validate-optimizations.sh [--verbose]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
VERBOSE=false
if [[ "$1" == "--verbose" ]]; then
    VERBOSE=true
fi

# Check for required tools
command -v gh >/dev/null 2>&1 || { echo -e "${RED}✗${NC} GitHub CLI (gh) is required but not installed. Install: https://cli.github.com/"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo -e "${RED}✗${NC} jq is required but not installed. Install: brew install jq"; exit 1; }

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}CICD Optimization Validation${NC}"
echo -e "${BLUE}Session 4 - Week 1 Monitoring${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Get repository info
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
echo -e "Repository: ${GREEN}${REPO}${NC}"
echo -e "Date: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

# Function to print check result
check_result() {
    local status=$1
    local message=$2
    local details=$3

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    if [[ "$status" == "pass" ]]; then
        echo -e "${GREEN}✓${NC} ${message}"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        if [[ "$VERBOSE" == "true" && -n "$details" ]]; then
            echo -e "  ${details}"
        fi
    elif [[ "$status" == "fail" ]]; then
        echo -e "${RED}✗${NC} ${message}"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        if [[ -n "$details" ]]; then
            echo -e "  ${details}"
        fi
    elif [[ "$status" == "warn" ]]; then
        echo -e "${YELLOW}⚠${NC} ${message}"
        WARNINGS=$((WARNINGS + 1))
        if [[ -n "$details" ]]; then
            echo -e "  ${details}"
        fi
    fi
}

echo -e "${BLUE}1. Checking Modified Workflow Files${NC}"
echo "---"

# Check if modified workflow files exist
WORKFLOWS=(
    "benchmark.yml"
    "comprehensive-validation.yml"
    "integration-tests.yml"
    "test-matrix.yml"
    "scheduled-comprehensive.yml"
    "reusable-build.yml"
    "reusable-test.yml"
    "reusable-security-scan.yml"
    "reusable-docker-publish.yml"
    "deploy.yml"
)

for workflow in "${WORKFLOWS[@]}"; do
    if [[ -f ".github/workflows/${workflow}" ]]; then
        check_result "pass" "Workflow file exists: ${workflow}"
    else
        check_result "fail" "Workflow file missing: ${workflow}"
    fi
done
echo ""

echo -e "${BLUE}2. Validating Timeout Configurations${NC}"
echo "---"

# Check timeout changes in benchmark.yml
BENCHMARK_TIMEOUT=$(grep -A 0 "timeout-minutes:" .github/workflows/benchmark.yml | tail -1 | awk '{print $2}')
if [[ "$BENCHMARK_TIMEOUT" == "35" ]]; then
    check_result "pass" "benchmark.yml timeout optimized (35 min)"
else
    check_result "fail" "benchmark.yml timeout incorrect (expected 35, got ${BENCHMARK_TIMEOUT})"
fi

# Check timeout changes in integration-tests.yml
INT_TEST_TIMEOUT=$(grep -m 1 "timeout-minutes:" .github/workflows/integration-tests.yml | awk '{print $2}')
if [[ "$INT_TEST_TIMEOUT" == "25" ]]; then
    check_result "pass" "integration-tests.yml integration timeout optimized (25 min)"
else
    check_result "fail" "integration-tests.yml timeout incorrect (expected 25, got ${INT_TEST_TIMEOUT})"
fi

PERF_TEST_TIMEOUT=$(grep "timeout-minutes:" .github/workflows/integration-tests.yml | tail -1 | awk '{print $2}')
if [[ "$PERF_TEST_TIMEOUT" == "25" ]]; then
    check_result "pass" "integration-tests.yml performance timeout optimized (25 min)"
else
    check_result "fail" "integration-tests.yml performance timeout incorrect (expected 25, got ${PERF_TEST_TIMEOUT})"
fi
echo ""

echo -e "${BLUE}3. Validating Permission Configurations${NC}"
echo "---"

# Check for permissions blocks in modified workflows
for workflow in "${WORKFLOWS[@]}"; do
    if grep -q "^permissions:" ".github/workflows/${workflow}"; then
        check_result "pass" "Explicit permissions found in ${workflow}"
    else
        check_result "fail" "Missing explicit permissions in ${workflow}"
    fi
done
echo ""

echo -e "${BLUE}4. Validating Go Version Consistency${NC}"
echo "---"

# Check Go version in integration-tests.yml
GO_VERSION=$(grep "GO_VERSION:" .github/workflows/integration-tests.yml | head -1 | awk -F"'" '{print $2}')
if [[ "$GO_VERSION" == "1.25.4" ]]; then
    check_result "pass" "integration-tests.yml Go version standardized (1.25.4)"
else
    check_result "fail" "integration-tests.yml Go version incorrect (expected 1.25.4, got ${GO_VERSION})"
fi
echo ""

echo -e "${BLUE}5. Checking for Deprecated Actions${NC}"
echo "---"

# Check that actions/create-release is not used
if grep -r "actions/create-release@" .github/workflows/ >/dev/null 2>&1; then
    check_result "fail" "Deprecated actions/create-release found in workflows"
else
    check_result "pass" "No deprecated actions/create-release found"
fi

# Verify gh release create is used in deploy.yml
if grep -q "gh release create" .github/workflows/deploy.yml; then
    check_result "pass" "deploy.yml uses modern GitHub CLI for releases"
else
    check_result "warn" "deploy.yml may not be using GitHub CLI for releases"
fi
echo ""

echo -e "${BLUE}6. Checking Recent Workflow Runs (Last 24 Hours)${NC}"
echo "---"

# Get recent runs
RECENT_RUNS=$(gh run list --limit 100 --json createdAt,conclusion 2>/dev/null | \
    jq --arg date "$(date -u -v-1d +%Y-%m-%dT%H:%M:%S)" \
    '[.[] | select(.createdAt > $date)] | length' 2>/dev/null || echo "0")

if [[ "$RECENT_RUNS" -gt 0 ]]; then
    check_result "pass" "Found ${RECENT_RUNS} workflow runs in last 24 hours"

    # Calculate success rate
    SUCCESS_RUNS=$(gh run list --status success --limit 100 --json createdAt 2>/dev/null | \
        jq --arg date "$(date -u -v-1d +%Y-%m-%dT%H:%M:%S)" \
        '[.[] | select(.createdAt > $date)] | length' 2>/dev/null || echo "0")

    if [[ "$RECENT_RUNS" -gt 0 ]]; then
        SUCCESS_RATE=$(echo "scale=1; ($SUCCESS_RUNS * 100) / $RECENT_RUNS" | bc)
        if (( $(echo "$SUCCESS_RATE >= 95" | bc -l) )); then
            check_result "pass" "Success rate: ${SUCCESS_RATE}% (${SUCCESS_RUNS}/${RECENT_RUNS})"
        elif (( $(echo "$SUCCESS_RATE >= 85" | bc -l) )); then
            check_result "warn" "Success rate: ${SUCCESS_RATE}% (${SUCCESS_RUNS}/${RECENT_RUNS}) - Below target"
        else
            check_result "fail" "Success rate: ${SUCCESS_RATE}% (${SUCCESS_RUNS}/${RECENT_RUNS}) - Critical"
        fi
    fi
else
    check_result "warn" "No workflow runs found in last 24 hours - may need to trigger workflows"
fi
echo ""

echo -e "${BLUE}7. Checking for Permission Errors${NC}"
echo "---"

# Check recent runs for permission errors (last 10 runs)
PERMISSION_ERRORS=0
RUN_IDS=$(gh run list --limit 10 --json databaseId -q '.[].databaseId' 2>/dev/null || echo "")

if [[ -n "$RUN_IDS" ]]; then
    while IFS= read -r run_id; do
        if gh run view "$run_id" --log 2>/dev/null | grep -qi "permission\|forbidden\|403"; then
            PERMISSION_ERRORS=$((PERMISSION_ERRORS + 1))
            if [[ "$VERBOSE" == "true" ]]; then
                echo "  Run ID with permission error: $run_id"
            fi
        fi
    done <<< "$RUN_IDS"

    if [[ "$PERMISSION_ERRORS" -eq 0 ]]; then
        check_result "pass" "No permission errors found in last 10 runs"
    else
        check_result "fail" "Found ${PERMISSION_ERRORS} runs with permission errors"
    fi
else
    check_result "warn" "Could not check permission errors - no recent runs"
fi
echo ""

echo -e "${BLUE}8. Checking for Timeout Failures${NC}"
echo "---"

# Check for timeout failures in last 24 hours
TIMEOUT_FAILURES=$(gh run list --status timed_out --limit 100 --json createdAt 2>/dev/null | \
    jq --arg date "$(date -u -v-1d +%Y-%m-%dT%H:%M:%S)" \
    '[.[] | select(.createdAt > $date)] | length' 2>/dev/null || echo "0")

if [[ "$TIMEOUT_FAILURES" -eq 0 ]]; then
    check_result "pass" "No timeout failures in last 24 hours"
else
    check_result "fail" "Found ${TIMEOUT_FAILURES} timeout failures in last 24 hours"
fi
echo ""

echo -e "${BLUE}9. Checking Workflow-Specific Validations${NC}"
echo "---"

# Check if benchmark workflow has run recently
BENCHMARK_RUNS=$(gh run list --workflow=benchmark.yml --limit 5 --json conclusion 2>/dev/null | jq 'length' 2>/dev/null || echo "0")
if [[ "$BENCHMARK_RUNS" -gt 0 ]]; then
    BENCHMARK_SUCCESS=$(gh run list --workflow=benchmark.yml --status success --limit 5 2>/dev/null | wc -l)
    if [[ "$BENCHMARK_SUCCESS" -gt 0 ]]; then
        check_result "pass" "benchmark.yml has successful runs (${BENCHMARK_SUCCESS})"
    else
        check_result "warn" "benchmark.yml has runs but none successful"
    fi
else
    check_result "warn" "benchmark.yml has not run recently"
fi

# Check if integration tests have run
INT_TEST_RUNS=$(gh run list --workflow="Integration Tests" --limit 5 --json conclusion 2>/dev/null | jq 'length' 2>/dev/null || echo "0")
if [[ "$INT_TEST_RUNS" -gt 0 ]]; then
    INT_TEST_SUCCESS=$(gh run list --workflow="Integration Tests" --status success --limit 5 2>/dev/null | wc -l)
    if [[ "$INT_TEST_SUCCESS" -gt 0 ]]; then
        check_result "pass" "integration-tests.yml has successful runs (${INT_TEST_SUCCESS})"
    else
        check_result "warn" "integration-tests.yml has runs but none successful"
    fi
else
    check_result "warn" "integration-tests.yml has not run recently"
fi

# Check if deployment workflow has run
DEPLOY_RUNS=$(gh run list --workflow=deploy.yml --limit 5 --json conclusion 2>/dev/null | jq 'length' 2>/dev/null || echo "0")
if [[ "$DEPLOY_RUNS" -gt 0 ]]; then
    check_result "pass" "deploy.yml has recent runs"
else
    check_result "warn" "deploy.yml has not run recently (manual trigger workflow)"
fi
echo ""

echo -e "${BLUE}10. Documentation Validation${NC}"
echo "---"

# Check if documentation files exist
DOCS=(
    "SESSION_SUMMARY.md"
    "TIMEOUT_OPTIMIZATION_SUMMARY.md"
    "PERMISSIONS_OPTIMIZATION_SUMMARY.md"
    "WORKFLOW_VALIDATION_REPORT.md"
    "WEEK1_MONITORING_GUIDE.md"
)

for doc in "${DOCS[@]}"; do
    if [[ -f ".github/${doc}" ]]; then
        check_result "pass" "Documentation exists: ${doc}"
    else
        check_result "warn" "Documentation missing: ${doc}"
    fi
done
echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Validation Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Total Checks:    ${TOTAL_CHECKS}"
echo -e "${GREEN}Passed:          ${PASSED_CHECKS}${NC}"
echo -e "${RED}Failed:          ${FAILED_CHECKS}${NC}"
echo -e "${YELLOW}Warnings:        ${WARNINGS}${NC}"
echo ""

# Calculate pass rate
if [[ "$TOTAL_CHECKS" -gt 0 ]]; then
    PASS_RATE=$(echo "scale=1; ($PASSED_CHECKS * 100) / $TOTAL_CHECKS" | bc)
    echo -e "Pass Rate:       ${PASS_RATE}%"
    echo ""
fi

# Overall status
if [[ "$FAILED_CHECKS" -eq 0 ]]; then
    if [[ "$WARNINGS" -eq 0 ]]; then
        echo -e "Overall Status:  ${GREEN}✓ ALL CHECKS PASSED${NC}"
        exit 0
    else
        echo -e "Overall Status:  ${YELLOW}⚠ PASSED WITH WARNINGS${NC}"
        echo ""
        echo "Review warnings above. Optimizations are working but some workflows may need attention."
        exit 0
    fi
else
    echo -e "Overall Status:  ${RED}✗ VALIDATION FAILED${NC}"
    echo ""
    echo "Review failed checks above. Some optimizations may need adjustment."
    echo ""
    echo "Next steps:"
    echo "1. Review failed checks in detail"
    echo "2. Check workflow logs: gh run list --status failure"
    echo "3. Consult WEEK1_MONITORING_GUIDE.md for troubleshooting"
    echo "4. Consider selective rollback if critical issues persist"
    exit 1
fi
