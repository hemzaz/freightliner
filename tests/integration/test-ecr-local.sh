#!/bin/bash

# Comprehensive ECR → Local Replication Test Script
# Date: December 7, 2025
# Purpose: Find bugs in ECR to local replication functionality

set -e  # Exit on error (we'll handle errors manually)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
BUGS_FOUND=0

# Test output directory
TEST_DIR="/tmp/freightliner-test-$(date +%s)"
BUG_REPORT="${TEST_DIR}/bug-report.md"
TEST_LOG="${TEST_DIR}/test-log.txt"

# Freightliner binary
FREIGHTLINER="./freightliner"

# Create test directory
mkdir -p "${TEST_DIR}"

# Initialize bug report
cat > "${BUG_REPORT}" <<EOF
# ECR → Local Replication Bug Report
**Date**: $(date)
**Test Run ID**: $(date +%s)
**Environment**: $(uname -a)

---

## Bugs Found

EOF

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "${TEST_LOG}"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1" | tee -a "${TEST_LOG}"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1" | tee -a "${TEST_LOG}"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "${TEST_LOG}"
}

log_bug() {
    BUGS_FOUND=$((BUGS_FOUND + 1))
    echo -e "${RED}[BUG #${BUGS_FOUND}]${NC} $1" | tee -a "${TEST_LOG}"

    cat >> "${BUG_REPORT}" <<EOF

### Bug #${BUGS_FOUND}: $1
- **Severity**: ${2:-Medium}
- **Category**: ${3:-General}
- **Details**: ${4:-See log for details}

---

EOF
}

start_test() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_info "Test ${TOTAL_TESTS}: $1"
}

pass_test() {
    PASSED_TESTS=$((PASSED_TESTS + 1))
    log_success "$1"
}

fail_test() {
    FAILED_TESTS=$((FAILED_TESTS + 1))
    log_error "$1"
}

# =============================================================================
# TEST SUITE 1: CLI INTERFACE TESTS
# =============================================================================

log_info "========================================="
log_info "TEST SUITE 1: CLI INTERFACE TESTS"
log_info "========================================="

# Test 1.1: Basic help command
start_test "CLI Help Command"
if ${FREIGHTLINER} replicate --help > /dev/null 2>&1; then
    pass_test "Help command works"
else
    fail_test "Help command failed"
    log_bug "Help command returns non-zero exit code" "Low" "CLI"
fi

# Test 1.2: No arguments
start_test "CLI with no arguments"
if ${FREIGHTLINER} replicate 2>&1 | grep -q "requires"; then
    pass_test "Correctly requires arguments"
else
    fail_test "Should require arguments"
    log_bug "Replicate command doesn't validate required arguments" "High" "CLI" "Should show error when no source/dest provided"
fi

# Test 1.3: Single argument only
start_test "CLI with single argument only"
if ${FREIGHTLINER} replicate docker.io/library/alpine 2>&1 | grep -q "requires\|accepts"; then
    pass_test "Correctly requires both source and destination"
else
    fail_test "Should require both arguments"
    log_bug "Replicate command doesn't validate argument count" "High" "CLI"
fi

# Test 1.4: Invalid source format
start_test "CLI with invalid source format"
ERROR_OUTPUT=$(${FREIGHTLINER} replicate "not-a-valid-url" "${TEST_DIR}/dest" 2>&1 || true)
if echo "${ERROR_OUTPUT}" | grep -qi "error\|invalid\|failed"; then
    pass_test "Correctly rejects invalid source"
else
    log_warning "May not validate source format properly"
    log_bug "Weak validation of source URL format" "Medium" "CLI" "Should validate registry URL format"
fi

# Test 1.5: Destination doesn't exist
start_test "CLI with non-existent destination parent"
NON_EXIST_DIR="${TEST_DIR}/does/not/exist/at/all"
ERROR_OUTPUT=$(${FREIGHTLINER} replicate docker.io/library/alpine:latest "${NON_EXIST_DIR}" 2>&1 || true)
if echo "${ERROR_OUTPUT}" | grep -qi "error\|directory\|not found"; then
    pass_test "Correctly handles non-existent destination"
else
    log_warning "May not validate destination directory"
    log_bug "Doesn't validate destination directory existence" "Medium" "Error Handling"
fi

# Test 1.6: Dry-run flag
start_test "CLI with --dry-run flag"
DEST_DIR="${TEST_DIR}/dry-run-test"
mkdir -p "${DEST_DIR}"
${FREIGHTLINER} replicate --dry-run --log-level=debug docker.io/library/alpine:latest "${DEST_DIR}" > "${TEST_DIR}/dry-run.log" 2>&1 || true
if [ -f "${TEST_DIR}/dry-run.log" ]; then
    if grep -qi "dry.run\|dry_run" "${TEST_DIR}/dry-run.log"; then
        pass_test "Dry-run flag works"
    else
        log_warning "Dry-run flag may not be respected"
        log_bug "Dry-run flag doesn't appear in logs" "Medium" "CLI"
    fi
else
    fail_test "Dry-run execution failed"
fi

# =============================================================================
# TEST SUITE 2: AUTHENTICATION TESTS
# =============================================================================

log_info "========================================="
log_info "TEST SUITE 2: AUTHENTICATION TESTS"
log_info "========================================="

# Test 2.1: Public registry access (no auth)
start_test "Public registry access without authentication"
DEST_DIR="${TEST_DIR}/public-test"
mkdir -p "${DEST_DIR}"
log_info "Attempting to replicate from public Docker Hub..."
${FREIGHTLINER} replicate --log-level=debug docker.io/library/alpine:3.18 "${DEST_DIR}" > "${TEST_DIR}/public-test.log" 2>&1 &
PID=$!
sleep 5
if ps -p ${PID} > /dev/null 2>&1; then
    log_info "Replication in progress (PID: ${PID})"
    wait ${PID} || true
    if [ -d "${DEST_DIR}" ] && [ "$(ls -A ${DEST_DIR} 2>/dev/null | wc -l)" -gt 0 ]; then
        pass_test "Public registry replication succeeded"
    else
        fail_test "Public registry replication failed or produced no output"
        log_bug "Cannot replicate from public Docker Hub" "Critical" "Replication" "Check docker.io/library/alpine:3.18 replication"
    fi
else
    fail_test "Process terminated immediately"
    log_bug "Replicate command exits immediately without doing work" "Critical" "Replication"
fi

# Test 2.2: ECR without credentials (should fail gracefully)
start_test "ECR access without credentials"
DEST_DIR="${TEST_DIR}/ecr-no-creds"
mkdir -p "${DEST_DIR}"
# Temporarily unset AWS credentials
export AWS_ACCESS_KEY_ID_BACKUP="${AWS_ACCESS_KEY_ID:-}"
export AWS_SECRET_ACCESS_KEY_BACKUP="${AWS_SECRET_ACCESS_KEY:-}"
unset AWS_ACCESS_KEY_ID
unset AWS_SECRET_ACCESS_KEY
unset AWS_SESSION_TOKEN

ERROR_OUTPUT=$(${FREIGHTLINER} replicate 123456789012.dkr.ecr.us-east-1.amazonaws.com/test:latest "${DEST_DIR}" 2>&1 || true)

# Restore credentials
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID_BACKUP}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY_BACKUP}"

if echo "${ERROR_OUTPUT}" | grep -qi "credentials\|auth\|permission\|unauthorized"; then
    pass_test "Correctly reports authentication failure"
else
    log_warning "Authentication error message could be clearer"
    log_bug "Unclear authentication error messages" "Low" "Error Handling" "Should explicitly mention AWS credentials"
fi

# =============================================================================
# TEST SUITE 3: TAG HANDLING TESTS
# =============================================================================

log_info "========================================="
log_info "TEST SUITE 3: TAG HANDLING TESTS"
log_info "========================================="

# Test 3.1: Specific tag
start_test "Replicate specific tag"
DEST_DIR="${TEST_DIR}/specific-tag"
mkdir -p "${DEST_DIR}"
${FREIGHTLINER} replicate --tags latest --log-level=debug docker.io/library/alpine "${DEST_DIR}" > "${TEST_DIR}/specific-tag.log" 2>&1 &
PID=$!
sleep 3
if ps -p ${PID} > /dev/null 2>&1; then
    log_info "Tag-specific replication in progress"
    wait ${PID} || true
    pass_test "Tag-specific replication executed"
else
    fail_test "Tag-specific replication failed immediately"
    log_bug "Cannot replicate specific tags" "High" "Tag Filtering"
fi

# Test 3.2: Multiple tags
start_test "Replicate multiple tags"
DEST_DIR="${TEST_DIR}/multi-tag"
mkdir -p "${DEST_DIR}"
${FREIGHTLINER} replicate --tags latest,3.18 --log-level=debug docker.io/library/alpine "${DEST_DIR}" > "${TEST_DIR}/multi-tag.log" 2>&1 &
PID=$!
sleep 3
if ps -p ${PID} > /dev/null 2>&1; then
    log_info "Multi-tag replication in progress"
    wait ${PID} || true
    pass_test "Multi-tag replication executed"
else
    fail_test "Multi-tag replication failed"
    log_bug "Cannot replicate multiple tags" "Medium" "Tag Filtering"
fi

# Test 3.3: Non-existent tag
start_test "Replicate non-existent tag"
DEST_DIR="${TEST_DIR}/nonexist-tag"
mkdir -p "${DEST_DIR}"
ERROR_OUTPUT=$(${FREIGHTLINER} replicate --tags this-tag-does-not-exist-12345 docker.io/library/alpine "${DEST_DIR}" 2>&1 || true)
if echo "${ERROR_OUTPUT}" | grep -qi "not found\|doesn't exist\|no such tag"; then
    pass_test "Correctly handles non-existent tag"
else
    log_warning "Error message for non-existent tag could be clearer"
    log_bug "Unclear error for non-existent tags" "Low" "Error Handling"
fi

# =============================================================================
# TEST SUITE 4: CONFIG FILE TESTS
# =============================================================================

log_info "========================================="
log_info "TEST SUITE 4: CONFIG FILE TESTS"
log_info "========================================="

# Test 4.1: Load config from file
start_test "Load configuration from file"
TEST_CONFIG="${TEST_DIR}/test-config.yaml"
cat > "${TEST_CONFIG}" <<CONFEOF
logLevel: debug
replicate:
  force: false
  dryRun: true
  tags:
    - latest
ecr:
  region: us-east-1
workers:
  replicateWorkers: 2
CONFEOF

DEST_DIR="${TEST_DIR}/config-test"
mkdir -p "${DEST_DIR}"
${FREIGHTLINER} --config "${TEST_CONFIG}" replicate docker.io/library/alpine "${DEST_DIR}" > "${TEST_DIR}/config-test.log" 2>&1 &
PID=$!
sleep 2
if ps -p ${PID} > /dev/null 2>&1; then
    wait ${PID} || true
    if grep -qi "config\|yaml" "${TEST_DIR}/config-test.log"; then
        pass_test "Config file loaded successfully"
    else
        log_warning "Config file may not be loaded"
        log_bug "Config file not being loaded or logged" "Medium" "Config"
    fi
else
    fail_test "Failed to run with config file"
    log_bug "Cannot use config file" "High" "Config"
fi

# Test 4.2: Invalid config file
start_test "Invalid config file path"
ERROR_OUTPUT=$(${FREIGHTLINER} --config /nonexistent/config.yaml replicate docker.io/library/alpine "${TEST_DIR}/dest" 2>&1 || true)
if echo "${ERROR_OUTPUT}" | grep -qi "config\|not found\|no such file"; then
    pass_test "Correctly handles missing config file"
else
    log_warning "Missing config file error could be clearer"
    log_bug "Unclear error for missing config file" "Low" "Config"
fi

# Test 4.3: Malformed YAML
start_test "Malformed YAML config"
BAD_CONFIG="${TEST_DIR}/bad-config.yaml"
cat > "${BAD_CONFIG}" <<CONFEOF
logLevel: debug
replicate:
  force: false
  - this is invalid yaml
  dryRun: true
CONFEOF

ERROR_OUTPUT=$(${FREIGHTLINER} --config "${BAD_CONFIG}" replicate docker.io/library/alpine "${TEST_DIR}/dest" 2>&1 || true)
if echo "${ERROR_OUTPUT}" | grep -qi "yaml\|parse\|invalid\|format"; then
    pass_test "Correctly handles malformed YAML"
else
    log_warning "YAML parse error could be clearer"
    log_bug "Unclear error for malformed YAML" "Low" "Config"
fi

# =============================================================================
# TEST SUITE 5: ERROR HANDLING TESTS
# =============================================================================

log_info "========================================="
log_info "TEST SUITE 5: ERROR HANDLING TESTS"
log_info "========================================="

# Test 5.1: Read-only destination
start_test "Read-only destination directory"
READONLY_DIR="${TEST_DIR}/readonly"
mkdir -p "${READONLY_DIR}"
chmod 444 "${READONLY_DIR}"
ERROR_OUTPUT=$(${FREIGHTLINER} replicate docker.io/library/alpine:latest "${READONLY_DIR}" 2>&1 || true)
chmod 755 "${READONLY_DIR}"  # Restore permissions
if echo "${ERROR_OUTPUT}" | grep -qi "permission\|denied\|read-only\|write"; then
    pass_test "Correctly handles read-only destination"
else
    log_warning "Permission error could be clearer"
    log_bug "Unclear error for read-only destination" "Medium" "Error Handling"
fi

# Test 5.2: Destination is a file, not directory
start_test "Destination is a file"
FILE_DEST="${TEST_DIR}/file-not-dir"
touch "${FILE_DEST}"
ERROR_OUTPUT=$(${FREIGHTLINER} replicate docker.io/library/alpine:latest "${FILE_DEST}" 2>&1 || true)
if echo "${ERROR_OUTPUT}" | grep -qi "directory\|not a directory\|invalid"; then
    pass_test "Correctly rejects file as destination"
else
    log_warning "File vs directory error could be clearer"
    log_bug "Doesn't validate destination is a directory" "Medium" "Error Handling"
fi

# Test 5.3: Empty source string
start_test "Empty source string"
ERROR_OUTPUT=$(${FREIGHTLINER} replicate "" "${TEST_DIR}/dest" 2>&1 || true)
if echo "${ERROR_OUTPUT}" | grep -qi "empty\|invalid\|required"; then
    pass_test "Correctly rejects empty source"
else
    log_warning "Empty source validation could be better"
    log_bug "Weak validation for empty source" "Medium" "CLI"
fi

# =============================================================================
# TEST SUITE 6: WORKER/CONCURRENCY TESTS
# =============================================================================

log_info "========================================="
log_info "TEST SUITE 6: WORKER/CONCURRENCY TESTS"
log_info "========================================="

# Test 6.1: Default worker count
start_test "Default worker count (auto-detect)"
DEST_DIR="${TEST_DIR}/workers-default"
mkdir -p "${DEST_DIR}"
${FREIGHTLINER} --log-level=debug replicate docker.io/library/alpine:latest "${DEST_DIR}" > "${TEST_DIR}/workers-default.log" 2>&1 &
PID=$!
sleep 2
if ps -p ${PID} > /dev/null 2>&1; then
    wait ${PID} || true
    if grep -qi "worker" "${TEST_DIR}/workers-default.log"; then
        pass_test "Worker count logged"
    else
        log_warning "Worker count not visible in logs"
        log_bug "Worker information not logged" "Low" "Logging"
    fi
else
    fail_test "Failed with default workers"
fi

# Test 6.2: Specific worker count
start_test "Specific worker count (2 workers)"
DEST_DIR="${TEST_DIR}/workers-2"
mkdir -p "${DEST_DIR}"
${FREIGHTLINER} --replicate-workers=2 --log-level=debug replicate docker.io/library/alpine:latest "${DEST_DIR}" > "${TEST_DIR}/workers-2.log" 2>&1 &
PID=$!
sleep 2
if ps -p ${PID} > /dev/null 2>&1; then
    wait ${PID} || true
    pass_test "Explicit worker count accepted"
else
    fail_test "Failed with explicit worker count"
    log_bug "Cannot set explicit worker count" "Medium" "Workers"
fi

# Test 6.3: Invalid worker count (negative)
start_test "Invalid worker count (negative)"
ERROR_OUTPUT=$(${FREIGHTLINER} --replicate-workers=-1 replicate docker.io/library/alpine:latest "${TEST_DIR}/dest" 2>&1 || true)
if echo "${ERROR_OUTPUT}" | grep -qi "invalid\|positive\|greater"; then
    pass_test "Correctly rejects negative worker count"
else
    log_warning "Negative worker count not validated"
    log_bug "Doesn't validate worker count range" "Low" "CLI"
fi

# Test 6.4: Excessive worker count
start_test "Excessive worker count (1000)"
DEST_DIR="${TEST_DIR}/workers-1000"
mkdir -p "${DEST_DIR}"
${FREIGHTLINER} --replicate-workers=1000 --log-level=debug replicate docker.io/library/alpine:latest "${DEST_DIR}" > "${TEST_DIR}/workers-1000.log" 2>&1 &
PID=$!
sleep 2
if ps -p ${PID} > /dev/null 2>&1; then
    kill ${PID} 2>/dev/null || true
    log_warning "Accepts unreasonably high worker count"
    log_bug "No upper limit on worker count" "Low" "Workers" "Should cap at reasonable maximum (e.g., 100)"
else
    pass_test "System handled excessive workers"
fi

# =============================================================================
# TEST SUMMARY
# =============================================================================

log_info "========================================="
log_info "TEST EXECUTION COMPLETE"
log_info "========================================="

# Append summary to bug report
cat >> "${BUG_REPORT}" <<EOF

## Test Summary

- **Total Tests**: ${TOTAL_TESTS}
- **Passed**: ${PASSED_TESTS}
- **Failed**: ${FAILED_TESTS}
- **Bugs Found**: ${BUGS_FOUND}

## Test Environment

- **Test Directory**: ${TEST_DIR}
- **Log File**: ${TEST_LOG}
- **Freightliner Binary**: ${FREIGHTLINER}
- **Date**: $(date)

## Next Steps

1. Review each bug for severity and priority
2. Create GitHub issues for critical/high bugs
3. Implement fixes with test coverage
4. Re-run test suite to verify fixes

---

**End of Report**
EOF

# Print summary
echo ""
echo "========================================="
echo "TEST SUMMARY"
echo "========================================="
echo "Total Tests:  ${TOTAL_TESTS}"
echo "Passed:       ${GREEN}${PASSED_TESTS}${NC}"
echo "Failed:       ${RED}${FAILED_TESTS}${NC}"
echo "Bugs Found:   ${RED}${BUGS_FOUND}${NC}"
echo ""
echo "Bug Report:   ${BUG_REPORT}"
echo "Test Log:     ${TEST_LOG}"
echo "Test Dir:     ${TEST_DIR}"
echo "========================================="

# Exit with error if tests failed
if [ ${FAILED_TESTS} -gt 0 ] || [ ${BUGS_FOUND} -gt 5 ]; then
    exit 1
else
    exit 0
fi
