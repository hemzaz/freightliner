#!/bin/bash
# Script to run staticcheck static analysis

set -e

echo "Running staticcheck analysis..."

# Check if staticcheck is installed
if ! command -v staticcheck &> /dev/null; then
    echo "staticcheck is not installed. Installing..."
    go install honnef.co/go/tools/cmd/staticcheck@latest
fi

# Define color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Default checks
CHECKS="all"

# Parse command line options
PACKAGES="./..."
VERBOSE=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --checks=*)
      CHECKS="${1#*=}"
      shift
      ;;
    --verbose)
      VERBOSE=true
      shift
      ;;
    *)
      PACKAGES="$1"
      shift
      ;;
  esac
done

# Run staticcheck
echo -e "${YELLOW}Running staticcheck with checks: ${CHECKS}${NC}"

if [ "$VERBOSE" = true ]; then
  staticcheck -checks="${CHECKS}" -f=text -explain $PACKAGES
else
  staticcheck -checks="${CHECKS}" $PACKAGES
fi

RESULT=$?

if [ $RESULT -eq 0 ]; then
  echo -e "${GREEN}Staticcheck completed successfully with no issues!${NC}"
else
  echo -e "${RED}Staticcheck found issues. Please fix them before committing.${NC}"
  echo -e "${YELLOW}Tip: Run with --verbose to see detailed explanations of issues.${NC}"
  exit $RESULT
fi
