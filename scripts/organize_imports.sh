#!/bin/bash
# Script to organize imports in Go files

set -e

echo "Organizing imports in Go files..."

# Check if goimports is installed
if ! command -v goimports &> /dev/null; then
    echo "goimports is not installed. Installing..."
    go install golang.org/x/tools/cmd/goimports@v0.29.0
fi

# Define patterns to organize imports
# 1. Standard library packages
# 2. Third-party packages
# 3. Local packages (freightliner)

# Run goimports with verbose output
find . -type f -name "*.go" ! -path "./vendor/*" ! -path "./.git/*" | xargs goimports -w -local freightliner

echo "Import organization complete."
