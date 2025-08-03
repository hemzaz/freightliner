#!/bin/bash
# Version Validation Script
# Validates that all workflows use consistent and latest versions

set -e

echo "🔍 Validating GitHub Actions versions across workflows..."

WORKFLOW_DIR=".github/workflows"
ACTIONS_DIR=".github/actions"

# Define expected versions
declare -A EXPECTED_VERSIONS=(
    ["actions/checkout"]="v4"
    ["actions/setup-go"]="v5"
    ["actions/cache"]="v4"
    ["actions/upload-artifact"]="v4"
    ["actions/download-artifact"]="v4"
    ["docker/setup-buildx-action"]="v3"
    ["docker/login-action"]="v3"
    ["docker/build-push-action"]="v6"
    ["docker/metadata-action"]="v5"
    ["github/codeql-action/upload-sarif"]="v3"
    ["golangci/golangci-lint-action"]="v6"
    ["codecov/codecov-action"]="v4"
    ["aquasecurity/trivy-action"]="0.30.0"
    ["anchore/scan-action"]="v4"
    ["softprops/action-gh-release"]="v2"
)

# Function to check version in file
check_versions() {
    local file="$1"
    local issues=0
    
    echo "📄 Checking $file..."
    
    for action in "${!EXPECTED_VERSIONS[@]}"; do
        expected="${EXPECTED_VERSIONS[$action]}"
        
        # Find all uses of this action
        while IFS= read -r line; do
            if [[ -n "$line" ]]; then
                # Extract version from the line
                current_version=$(echo "$line" | sed -n "s/.*${action}@\([^[:space:]]*\).*/\1/p")
                
                if [[ "$current_version" != "$expected" ]]; then
                    echo "❌ $file: $action@$current_version (expected: $expected)"
                    issues=$((issues + 1))
                else
                    echo "✅ $file: $action@$current_version"
                fi
            fi
        done < <(grep -n "uses:.*$action@" "$file" 2>/dev/null || true)
    done
    
    return $issues
}

# Function to check Go version consistency
check_go_versions() {
    echo "🐹 Checking Go version consistency..."
    
    local go_version_files=(
        ".github/actions/setup-go/action.yml"
        ".github/workflows/ci.yml"
        ".github/workflows/release.yml"
        ".github/workflows/security.yml"
        ".github/workflows/scheduled-comprehensive.yml"
    )
    
    local expected_go_version="1.24.5"
    local issues=0
    
    for file in "${go_version_files[@]}"; do
        if [[ -f "$file" ]]; then
            while IFS= read -r line; do
                if [[ -n "$line" ]]; then
                    current_version=$(echo "$line" | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1)
                    if [[ "$current_version" != "$expected_go_version" ]]; then
                        echo "❌ $file: Go $current_version (expected: $expected_go_version)"
                        issues=$((issues + 1))
                    else
                        echo "✅ $file: Go $current_version"
                    fi
                fi
            done < <(grep -n "GO_VERSION\|go-version\|default:" "$file" 2>/dev/null | grep -v "#" || true)
        fi
    done
    
    return $issues
}

# Main validation
total_issues=0

# Check workflow files
if [[ -d "$WORKFLOW_DIR" ]]; then
    for file in "$WORKFLOW_DIR"/*.yml "$WORKFLOW_DIR"/*.yaml; do
        if [[ -f "$file" ]]; then
            check_versions "$file"
            total_issues=$((total_issues + $?))
        fi
    done
fi

# Check action files
if [[ -d "$ACTIONS_DIR" ]]; then
    for file in $(find "$ACTIONS_DIR" -name "*.yml" -o -name "*.yaml"); do
        if [[ -f "$file" ]]; then
            check_versions "$file"
            total_issues=$((total_issues + $?))
        fi
    done
fi

# Check Go versions
check_go_versions
total_issues=$((total_issues + $?))

# Check for deprecated actions
echo "🚨 Checking for deprecated actions..."
deprecated_patterns=(
    "actions/setup-go@v[123]"
    "actions/checkout@v[123]"
    "actions/cache@v[123]"
    "actions/upload-artifact@v[123]"
    "github/codeql-action.*@v[12]"
    "codecov/codecov-action@v[123]"
    "docker/build-push-action@v[12345]"
)

for pattern in "${deprecated_patterns[@]}"; do
    if grep -r "$pattern" "$WORKFLOW_DIR" "$ACTIONS_DIR" 2>/dev/null; then
        echo "⚠️  Found deprecated action: $pattern"
        total_issues=$((total_issues + 1))
    fi
done

# Summary
echo ""
echo "📊 Validation Summary:"
if [[ $total_issues -eq 0 ]]; then
    echo "✅ All versions are up to date and consistent!"
    exit 0
else
    echo "❌ Found $total_issues version issues"
    echo ""
    echo "💡 To fix these issues:"
    echo "   1. Update versions in workflow files"
    echo "   2. Check the VERSION_MATRIX.md for latest versions"
    echo "   3. Test workflows after updates"
    exit 1
fi