#!/bin/bash
# Simple version checker for GitHub Actions

echo "🔍 Checking GitHub Actions versions..."

# Check for specific action versions
echo ""
echo "📋 Current action versions found:"

# Find all action usages
find .github -name "*.yml" -o -name "*.yaml" | xargs grep -h "uses:" | sort -u | while read line; do
    if [[ $line =~ uses:[[:space:]]*([^@]+)@([^[:space:]]+) ]]; then
        action="${BASH_REMATCH[1]}"
        version="${BASH_REMATCH[2]}"
        echo "  $action@$version"
    fi
done

echo ""
echo "⚡ Go versions found:"
find .github -name "*.yml" -o -name "*.yaml" | xargs grep -h "GO_VERSION\|go-version:" | grep -v "#" | sort -u

echo ""
echo "🎯 Tool versions found:"
find .github -name "*.yml" -o -name "*.yaml" | xargs grep -h "GOLANGCI_LINT_VERSION" | grep -v "#" | sort -u

echo ""
echo "✅ Version check complete!"