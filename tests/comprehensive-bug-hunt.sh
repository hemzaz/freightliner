#!/bin/bash
# Comprehensive Bug Hunting Test Suite
# Tests edge cases, error scenarios, and command combinations

set +e  # Don't exit on error - we want to capture all failures

BINARY="./freightliner"
FAILURES=0
TESTS_RUN=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_test() {
    echo -e "\n${YELLOW}[TEST $((TESTS_RUN+1))]${NC} $1"
    TESTS_RUN=$((TESTS_RUN+1))
}

log_pass() {
    echo -e "${GREEN}✅ PASS${NC}: $1"
}

log_fail() {
    echo -e "${RED}❌ FAIL${NC}: $1"
    FAILURES=$((FAILURES+1))
}

log_warn() {
    echo -e "${YELLOW}⚠️  WARN${NC}: $1"
}

echo "======================================================"
echo "Comprehensive Bug Hunting Test Suite"
echo "======================================================"
echo ""

# Test 1: Empty/Invalid Registry Names
log_test "Empty registry name"
if $BINARY replicate "" docker.io/test/alpine:latest 2>&1 | grep -q "invalid format"; then
    log_pass "Rejected empty registry name"
else
    log_fail "Did not reject empty registry name"
fi

# Test 2: Very long repository names
log_test "Very long repository name (255 chars)"
LONG_REPO=$(python3 -c "print('a' * 255)")
if $BINARY replicate --dry-run --tags latest docker.io/library/alpine:latest "docker.io/$LONG_REPO:latest" 2>&1 | grep -qE "(invalid|error|Error)"; then
    log_pass "Handled very long repository name"
else
    log_warn "Very long repository name not validated"
fi

# Test 3: Special characters in repository names
log_test "Special characters in repository name"
if $BINARY replicate --dry-run --tags latest docker.io/library/alpine:latest "docker.io/test/image@#$%:latest" 2>&1 | grep -qE "(invalid|error|Error)"; then
    log_pass "Rejected special characters"
else
    log_warn "Special characters not validated"
fi

# Test 4: Multiple tags with some invalid
log_test "Multiple tags with mix of valid and invalid"
OUTPUT=$($BINARY --tags latest,nonexistent,3.18 replicate --dry-run docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1)
if echo "$OUTPUT" | grep -q "Available tags"; then
    log_pass "Shows tag suggestions for invalid tags"
else
    log_fail "Does not show tag suggestions"
fi

# Test 5: Worker count boundary conditions
log_test "Worker count = 0 (should auto-detect)"
if $BINARY --replicate-workers 0 replicate --dry-run --tags latest docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1 | grep -q "Auto-detected worker count"; then
    log_pass "Auto-detects workers when set to 0"
else
    log_fail "Does not auto-detect with workers=0"
fi

# Test 6: Worker count = 1
log_test "Worker count = 1 (minimum)"
if $BINARY --replicate-workers 1 replicate --dry-run --tags latest docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1 | grep -q "Using configured worker count workers=1"; then
    log_pass "Accepts worker count = 1"
else
    log_fail "Does not accept worker count = 1"
fi

# Test 7: Worker count = 1000 (should cap)
log_test "Worker count = 1000 (should cap at 100)"
if $BINARY --replicate-workers 1000 replicate --dry-run --tags latest docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1 | grep -q "capping at maximum"; then
    log_pass "Caps excessive worker count at 100"
else
    log_fail "Does not cap excessive worker count"
fi

# Test 8: Negative worker count
log_test "Negative worker count (should error)"
if $BINARY --replicate-workers -5 replicate --dry-run --tags latest docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1 | grep -qE "(invalid|error|Error)"; then
    log_pass "Rejects negative worker count"
else
    log_fail "Accepts negative worker count"
fi

# Test 9: Invalid config file
log_test "Non-existent config file"
if $BINARY --config /nonexistent/config.yaml replicate --dry-run docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1 | grep -qE "(failed to load|not found|no such file)"; then
    log_pass "Reports missing config file"
else
    log_fail "Does not report missing config file"
fi

# Test 10: Filesystem path as source
log_test "Filesystem path as source"
if $BINARY replicate /tmp/source docker.io/test/alpine:copy 2>&1 | grep -q "filesystem paths are not supported"; then
    log_pass "Rejects filesystem path as source"
else
    log_fail "Does not reject filesystem path as source"
fi

# Test 11: Filesystem path as destination
log_test "Filesystem path as destination"
if $BINARY replicate docker.io/library/alpine:latest /tmp/dest 2>&1 | grep -q "filesystem paths are not supported"; then
    log_pass "Rejects filesystem path as destination"
else
    log_fail "Does not reject filesystem path as destination"
fi

# Test 12: Missing source argument
log_test "Missing source argument"
if $BINARY replicate 2>&1 | grep -qE "(accepts 2 arg|Example)"; then
    log_pass "Shows usage with examples when args missing"
else
    log_fail "Does not show helpful usage message"
fi

# Test 13: Missing destination argument
log_test "Missing destination argument"
if $BINARY replicate docker.io/library/alpine:latest 2>&1 | grep -qE "(accepts 2 arg|Example)"; then
    log_pass "Shows usage with examples when destination missing"
else
    log_fail "Does not show helpful usage message"
fi

# Test 14: Empty tags flag
log_test "Empty --tags flag"
OUTPUT=$($BINARY --tags "" replicate --dry-run docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1)
if echo "$OUTPUT" | grep -qE "(Listing all tags|tag_count)"; then
    log_pass "Falls back to all tags when --tags is empty"
else
    log_warn "Behavior unclear with empty --tags"
fi

# Test 15: Duplicate tags
log_test "Duplicate tags in --tags flag"
OUTPUT=$($BINARY --tags latest,latest,latest replicate --dry-run docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1)
if echo "$OUTPUT" | grep -q "Copying image"; then
    log_pass "Handles duplicate tags"
else
    log_warn "Duplicate tags behavior unclear"
fi

# Test 16: Registry with port number
log_test "Registry with port number (localhost:5000)"
if $BINARY replicate --dry-run --tags latest localhost:5000/test/image:latest docker.io/test/copy:latest 2>&1 | grep -qE "(Error|error|invalid)"; then
    log_warn "Port numbers may not be fully supported"
else
    log_pass "Handles registry with port number"
fi

# Test 17: Tag with digest instead of tag name
log_test "Image reference with digest"
if $BINARY replicate --dry-run docker.io/library/alpine@sha256:abc123 docker.io/test/alpine:copy 2>&1; then
    log_pass "Handles digest references"
else
    log_warn "Digest references may have issues"
fi

# Test 18: Very long tag name
log_test "Very long tag name (128 chars)"
LONG_TAG=$(python3 -c "print('v' + '1' * 127)")
OUTPUT=$($BINARY --tags "$LONG_TAG" replicate --dry-run docker.io/library/alpine:latest docker.io/test/alpine:copy 2>&1)
if echo "$OUTPUT" | grep -qE "(invalid|error)"; then
    log_pass "Validates very long tag names"
else
    log_warn "Long tag names not validated"
fi

# Test 19: Unicode in repository names
log_test "Unicode characters in repository name"
if $BINARY replicate --dry-run docker.io/library/alpine:latest "docker.io/测试/image:latest" 2>&1 | grep -qE "(invalid|error|Error)"; then
    log_pass "Rejects unicode in repository names"
else
    log_warn "Unicode handling unclear"
fi

# Test 20: Help command works
log_test "Help command displays correctly"
if $BINARY replicate --help 2>&1 | grep -q "Example"; then
    log_pass "Help shows examples"
else
    log_fail "Help does not show examples"
fi

echo ""
echo "======================================================"
echo "Bug Hunting Summary"
echo "======================================================"
echo "Tests run: $TESTS_RUN"
echo "Failures: $FAILURES"

if [ $FAILURES -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ $FAILURES test(s) failed${NC}"
    exit 1
fi
