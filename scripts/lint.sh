#!/bin/bash
# Script to run golangci-lint with standard configuration

set -e

echo "Running golangci-lint..."

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint is not installed. Installing..."
    # Use the official installation script
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.0.2
fi

# Define color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Default timeout
TIMEOUT="3m"

# Parse command line options
EXTRA_ARGS=""
PATHS="./..."

while [[ $# -gt 0 ]]; do
  case $1 in
    --fast)
      EXTRA_ARGS="$EXTRA_ARGS --fast"
      shift
      ;;
    --timeout=*)
      TIMEOUT="${1#*=}"
      shift
      ;;
    --fix)
      EXTRA_ARGS="$EXTRA_ARGS --fix"
      shift
      ;;
    --*=*)
      EXTRA_ARGS="$EXTRA_ARGS $1"
      shift
      ;;
    *)
      PATHS="$1"
      shift
      ;;
  esac
done

# Run golangci-lint
echo -e "${YELLOW}Running linters with configuration from .golangci.yml${NC}"
echo -e "${YELLOW}Timeout: ${TIMEOUT}${NC}"

RESULT=0
golangci-lint run --timeout=$TIMEOUT $EXTRA_ARGS $PATHS || RESULT=$?

if [ $RESULT -eq 0 ]; then
    echo -e "${GREEN}Linting completed successfully with no issues!${NC}"
else
    echo -e "${RED}Linting found issues. Please fix them before committing.${NC}"
    exit $RESULT
fi
