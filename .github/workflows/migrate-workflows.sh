#!/bin/bash
# ============================================================================
# WORKFLOW MIGRATION SCRIPT
# ============================================================================
# Purpose: Archive old workflows and activate new optimized workflows
# Usage: ./migrate-workflows.sh [--dry-run|--execute]
# ============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKFLOWS_DIR="$SCRIPT_DIR"
ARCHIVED_DIR="$WORKFLOWS_DIR/archived"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Workflows to archive (OLD workflows being replaced)
WORKFLOWS_TO_ARCHIVE=(
    "security-comprehensive.yml"
    "security-gates.yml"
    "security-gates-enhanced.yml"
    "security-monitoring-enhanced.yml"
    "helm-deploy.yml"
    "kubernetes-deploy.yml"
    "integration-tests.yml"
    "test-matrix.yml"
    "reusable-security-scan.yml"
)

# Workflows to keep active (NEW + essential workflows)
WORKFLOWS_TO_KEEP=(
    "consolidated-ci-v2.yml"
    "security-scan.yml"
    "deploy-unified.yml"
    "monitoring.yml"
    "release-pipeline.yml"
    "benchmark.yml"
    "rollback.yml"
    "reusable-build.yml"
    "reusable-docker-publish.yml"
    "reusable-test.yml"
    "oidc-authentication.yml"
)

# Print colored message
log() {
    local level=$1
    shift
    case $level in
        info)    echo -e "${BLUE}[INFO]${NC} $*" ;;
        success) echo -e "${GREEN}[SUCCESS]${NC} $*" ;;
        warning) echo -e "${YELLOW}[WARNING]${NC} $*" ;;
        error)   echo -e "${RED}[ERROR]${NC} $*" ;;
    esac
}

# Check if running in dry-run mode
DRY_RUN=true
if [[ "${1:-}" == "--execute" ]]; then
    DRY_RUN=false
    log warning "EXECUTE mode enabled - changes will be made"
else
    log info "DRY-RUN mode - no changes will be made (use --execute to apply changes)"
fi

# Create archived directory if it doesn't exist
if [[ ! -d "$ARCHIVED_DIR" ]]; then
    log info "Creating archived directory: $ARCHIVED_DIR"
    if [[ "$DRY_RUN" == false ]]; then
        mkdir -p "$ARCHIVED_DIR"
        log success "Archived directory created"
    fi
fi

# Archive old workflows
log info "Archiving old workflows..."
archived_count=0
for workflow in "${WORKFLOWS_TO_ARCHIVE[@]}"; do
    workflow_path="$WORKFLOWS_DIR/$workflow"
    if [[ -f "$workflow_path" ]]; then
        log info "  Archiving: $workflow"
        if [[ "$DRY_RUN" == false ]]; then
            # Add timestamp to archived file
            timestamp=$(date +%Y%m%d-%H%M%S)
            archived_name="${workflow%.yml}-${timestamp}.yml"
            mv "$workflow_path" "$ARCHIVED_DIR/$archived_name"
            log success "    Moved to: archived/$archived_name"
        else
            log info "    Would move to: archived/$workflow"
        fi
        ((archived_count++))
    else
        log warning "  Not found: $workflow (may already be archived)"
    fi
done

# Rename consolidated-ci.yml to consolidated-ci-v1.yml (backup)
if [[ -f "$WORKFLOWS_DIR/consolidated-ci.yml" ]]; then
    log info "Backing up consolidated-ci.yml to consolidated-ci-v1.yml"
    if [[ "$DRY_RUN" == false ]]; then
        mv "$WORKFLOWS_DIR/consolidated-ci.yml" "$WORKFLOWS_DIR/consolidated-ci-v1.yml"
        log success "Backup created"
    fi
fi

# Activate new consolidated-ci-v2.yml by renaming to consolidated-ci.yml
if [[ -f "$WORKFLOWS_DIR/consolidated-ci-v2.yml" ]]; then
    log info "Activating consolidated-ci-v2.yml as consolidated-ci.yml"
    if [[ "$DRY_RUN" == false ]]; then
        mv "$WORKFLOWS_DIR/consolidated-ci-v2.yml" "$WORKFLOWS_DIR/consolidated-ci.yml"
        log success "New CI workflow activated"
    fi
fi

# Backup old deploy.yml
if [[ -f "$WORKFLOWS_DIR/deploy.yml" ]]; then
    log info "Backing up deploy.yml to deploy-v1.yml"
    if [[ "$DRY_RUN" == false ]]; then
        mv "$WORKFLOWS_DIR/deploy.yml" "$WORKFLOWS_DIR/deploy-v1.yml"
        log success "Backup created"
    fi
fi

# Activate new deploy-unified.yml by renaming to deploy.yml
if [[ -f "$WORKFLOWS_DIR/deploy-unified.yml" ]]; then
    log info "Activating deploy-unified.yml as deploy.yml"
    if [[ "$DRY_RUN" == false ]]; then
        mv "$WORKFLOWS_DIR/deploy-unified.yml" "$WORKFLOWS_DIR/deploy.yml"
        log success "New deployment workflow activated"
    fi
fi

# Generate summary report
log info ""
log info "==============================================="
log info "MIGRATION SUMMARY"
log info "==============================================="
log info "Workflows archived: $archived_count"
log info "Workflows kept active: ${#WORKFLOWS_TO_KEEP[@]}"
log info ""

if [[ "$DRY_RUN" == true ]]; then
    log warning "DRY-RUN completed - no changes made"
    log info "Run with --execute to apply changes:"
    log info "  ./migrate-workflows.sh --execute"
else
    log success "Migration completed successfully!"
    log info ""
    log info "Next steps:"
    log info "1. Validate new workflows:"
    log info "   ./validate-workflows.sh"
    log info "2. Test in a feature branch before merging"
    log info "3. Update branch protection rules to use new workflows"
    log info "4. Monitor first few workflow runs"
fi

log info ""
log info "Active workflows:"
for workflow in "${WORKFLOWS_TO_KEEP[@]}"; do
    if [[ -f "$WORKFLOWS_DIR/$workflow" ]]; then
        log success "  ✓ $workflow"
    else
        log warning "  ✗ $workflow (not found)"
    fi
done

# List archived workflows
log info ""
log info "Archived workflows in $ARCHIVED_DIR:"
if [[ -d "$ARCHIVED_DIR" ]]; then
    archived_files=$(ls -1 "$ARCHIVED_DIR" | wc -l)
    log info "  Total archived files: $archived_files"
    if [[ $archived_files -gt 0 ]]; then
        ls -1 "$ARCHIVED_DIR" | head -10 | while read -r file; do
            log info "    - $file"
        done
        if [[ $archived_files -gt 10 ]]; then
            log info "    ... and $((archived_files - 10)) more"
        fi
    fi
fi

log info ""
log info "Migration script completed"
