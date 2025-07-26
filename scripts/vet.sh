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

# Ensure we're in module root directory
if [ ! -f "go.mod" ]; then
  echo -e "${RED}Error: go.mod not found in current directory. Make sure you're in the module root.${NC}"
  exit 1
fi

# Debug module setup
echo -e "${YELLOW}Debugging module setup...${NC}"
echo "Working directory: $(pwd)"
echo "GO111MODULE: ${GO111MODULE:-default}"
echo "GOFLAGS: ${GOFLAGS:-none}"
echo "Go module: $(go list -m 2>/dev/null || echo 'No module found')"
echo "Go version: $(go version)"
echo "GOMOD path: $(go env GOMOD 2>/dev/null || echo 'No go.mod found')"

# Ensure module mode is enabled
export GO111MODULE=on

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
  go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@v0.29.0
  go install github.com/mvdan/interfacer/cmd/interfacer@v0.0.0-20180902061238-70be1b28218b
  
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
