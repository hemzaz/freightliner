#!/bin/bash
# Workflow Validation Script
# Validates GitHub Actions workflow files for syntax and structure

set -e

echo "🔍 GitHub Actions Workflow Validation"
echo "======================================"
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

WORKFLOW_DIR=".github/workflows"
ERRORS=0
WARNINGS=0

# Check if workflow directory exists
if [ ! -d "$WORKFLOW_DIR" ]; then
    echo -e "${RED}❌ Workflow directory not found: $WORKFLOW_DIR${NC}"
    exit 1
fi

echo -e "${BLUE}📂 Checking workflow directory: $WORKFLOW_DIR${NC}"
echo ""

# Function to validate YAML syntax
validate_yaml() {
    local file="$1"
    echo -n "  Checking $file... "

    if python3 -c "import yaml; yaml.safe_load(open('$file'))" 2>/dev/null; then
        echo -e "${GREEN}✅ Valid YAML${NC}"
        return 0
    else
        echo -e "${RED}❌ Invalid YAML${NC}"
        ERRORS=$((ERRORS + 1))
        return 1
    fi
}

# Function to check workflow structure
check_workflow_structure() {
    local file="$1"
    local basename=$(basename "$file")

    # Skip non-workflow files
    if [[ "$basename" == "README.md" ]] || \
       [[ "$basename" == "MIGRATION-CHECKLIST.md" ]] || \
       [[ "$basename" == "WORKFLOW-ARCHITECTURE.md" ]] || \
       [[ "$basename" == "IMPLEMENTATION-SUMMARY.md" ]] || \
       [[ "$basename" == "validate-workflows.sh" ]]; then
        return 0
    fi

    echo "  Validating structure of $basename..."

    # Check for required fields
    if ! grep -q "^name:" "$file"; then
        echo -e "    ${YELLOW}⚠️  Warning: Missing 'name' field${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi

    if ! grep -q "^on:" "$file"; then
        echo -e "    ${RED}❌ Error: Missing 'on' trigger field${NC}"
        ERRORS=$((ERRORS + 1))
    fi

    if ! grep -q "^jobs:" "$file"; then
        echo -e "    ${RED}❌ Error: Missing 'jobs' field${NC}"
        ERRORS=$((ERRORS + 1))
    fi

    # Check for permissions (security best practice)
    if ! grep -q "^permissions:" "$file"; then
        echo -e "    ${YELLOW}⚠️  Warning: No explicit permissions defined${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi

    # Check for timeout-minutes (good practice)
    if ! grep -q "timeout-minutes:" "$file"; then
        echo -e "    ${YELLOW}⚠️  Note: No timeout defined for jobs${NC}"
    fi

    echo -e "    ${GREEN}✅ Structure check complete${NC}"
}

# New workflow files
NEW_WORKFLOWS=(
    "ci-cd-main.yml"
    "release-v2.yml"
    "reusable-security-scan.yml"
    "reusable-docker-publish.yml"
)

echo -e "${BLUE}🆕 Validating New Workflows${NC}"
echo "======================================"
echo ""

for workflow in "${NEW_WORKFLOWS[@]}"; do
    file="$WORKFLOW_DIR/$workflow"
    if [ -f "$file" ]; then
        echo -e "${BLUE}Validating: $workflow${NC}"
        if validate_yaml "$file"; then
            check_workflow_structure "$file"
        fi
        echo ""
    else
        echo -e "${RED}❌ File not found: $workflow${NC}"
        ERRORS=$((ERRORS + 1))
        echo ""
    fi
done

# Check existing workflows (optional)
echo -e "${BLUE}📋 Checking Existing Workflows${NC}"
echo "======================================"
echo ""

EXISTING_COUNT=0
for file in "$WORKFLOW_DIR"/*.yml; do
    if [ -f "$file" ]; then
        basename=$(basename "$file")
        # Skip new workflows (already checked)
        if [[ ! " ${NEW_WORKFLOWS[@]} " =~ " ${basename} " ]]; then
            EXISTING_COUNT=$((EXISTING_COUNT + 1))
        fi
    fi
done

echo "Found $EXISTING_COUNT existing workflow files"
echo ""

# Documentation check
echo -e "${BLUE}📚 Checking Documentation${NC}"
echo "======================================"
echo ""

DOCS=(
    "MIGRATION-CHECKLIST.md"
    "WORKFLOW-ARCHITECTURE.md"
    "IMPLEMENTATION-SUMMARY.md"
    "README.md"
)

for doc in "${DOCS[@]}"; do
    file="$WORKFLOW_DIR/$doc"
    if [ -f "$file" ]; then
        echo -e "  ${GREEN}✅ $doc exists${NC}"
    else
        echo -e "  ${YELLOW}⚠️  $doc not found${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi
done
echo ""

# Summary
echo -e "${BLUE}📊 Validation Summary${NC}"
echo "======================================"
echo ""
echo "New workflows validated: ${#NEW_WORKFLOWS[@]}"
echo "Existing workflows found: $EXISTING_COUNT"
echo -e "Errors: ${RED}$ERRORS${NC}"
echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"
echo ""

if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ All validations passed!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Review MIGRATION-CHECKLIST.md"
    echo "  2. Test workflows on a feature branch"
    echo "  3. Follow migration plan"
    exit 0
else
    echo -e "${RED}❌ Validation failed with $ERRORS error(s)${NC}"
    echo ""
    echo "Please fix the errors before proceeding."
    exit 1
fi
