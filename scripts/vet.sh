#!/bin/bash
# Script to run go vet with standard configuration

set -e

echo "Running go vet checks..."

# Define color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Parse command line options
PACKAGES="./..."
VERBOSE=false
SHOW_ALL=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --verbose)
      VERBOSE=true
      shift
      ;;
    --all)
      SHOW_ALL=true
      shift
      ;;
    *)
      PACKAGES="$1"
      shift
      ;;
  esac
done

# Run standard go vet
echo -e "${YELLOW}Running standard go vet checks...${NC}"
if [ "$VERBOSE" = true ]; then
  go vet -v $PACKAGES
else
  go vet $PACKAGES
fi
VET_RESULT=$?

# Additional checks - interfaces
echo -e "${YELLOW}Running interface check...${NC}"
IFACE_RESULT=0
go vet -vettool=$(which ifacecheck) $PACKAGES 2>/dev/null || IFACE_RESULT=$?

# Additional checks - shadow variables
echo -e "${YELLOW}Running shadow check...${NC}"
SHADOW_RESULT=0
go vet -vettool=$(which shadow) $PACKAGES 2>/dev/null || SHADOW_RESULT=$?

# Install missing tools if needed
if [ "${IFACE_RESULT}" -eq 127 ] || [ "${SHADOW_RESULT}" -eq 127 ]; then
  echo -e "${YELLOW}Installing additional vet tools...${NC}"
  go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
  go install github.com/mvdan/interfacer/cmd/interfacer@latest
  
  # Try again
  echo -e "${YELLOW}Running interface check...${NC}"
  go vet -vettool=$(which interfacer) $PACKAGES 2>/dev/null || IFACE_RESULT=$?
  
  echo -e "${YELLOW}Running shadow check...${NC}"
  go vet -vettool=$(which shadow) $PACKAGES 2>/dev/null || SHADOW_RESULT=$?
fi

# Simplify results for reporting
if [ "${IFACE_RESULT}" -eq 127 ]; then IFACE_RESULT=0; fi
if [ "${SHADOW_RESULT}" -eq 127 ]; then SHADOW_RESULT=0; fi

# Determine final result
RESULT=$((VET_RESULT + IFACE_RESULT + SHADOW_RESULT))

if [ $RESULT -eq 0 ]; then
  echo -e "${GREEN}All go vet checks passed!${NC}"
else
  echo -e "${RED}go vet found issues. Please fix them before committing.${NC}"
  exit 1
fi
