#!/bin/bash
# Automated rollback script with validation
set -euo pipefail

# ============================================
# Configuration
# ============================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ROLLBACK_TIMESTAMP=$(date +%Y%m%d-%H%M%S)
ROLLBACK_LOG="${PROJECT_ROOT}/logs/rollback-${ROLLBACK_TIMESTAMP}.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============================================
# Functions
# ============================================
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*" | tee -a "$ROLLBACK_LOG"
}

error() {
    echo -e "${RED}[ERROR] $*${NC}" | tee -a "$ROLLBACK_LOG" >&2
}

warn() {
    echo -e "${YELLOW}[WARN] $*${NC}" | tee -a "$ROLLBACK_LOG"
}

info() {
    echo -e "${BLUE}[INFO] $*${NC}" | tee -a "$ROLLBACK_LOG"
}

# Create log directory
mkdir -p "${PROJECT_ROOT}/logs"

# ============================================
# Parse arguments
# ============================================
ENVIRONMENT="${1:-dev}"
ROLLBACK_TO="${2:-previous}"

log "Starting rollback for ${ENVIRONMENT} environment"
log "Rollback target: ${ROLLBACK_TO}"

# ============================================
# Pre-rollback validation
# ============================================
log "Running pre-rollback validation..."

# Check required tools
for cmd in kubectl helm; do
    if ! command -v "$cmd" &> /dev/null; then
        error "$cmd is not installed"
        exit 1
    fi
done

# Validate environment
case "$ENVIRONMENT" in
    dev|staging|production)
        log "Environment validated: $ENVIRONMENT"
        ;;
    *)
        error "Invalid environment: $ENVIRONMENT"
        exit 1
        ;;
esac

KUBE_NAMESPACE="freightliner-${ENVIRONMENT}"

# Check if namespace exists
if ! kubectl get namespace "$KUBE_NAMESPACE" &> /dev/null; then
    error "Namespace $KUBE_NAMESPACE does not exist"
    exit 1
fi

# ============================================
# Determine rollback target
# ============================================
log "Determining rollback target..."

if [ "$ROLLBACK_TO" = "previous" ]; then
    # Get previous image from backup
    BACKUP_DIR="${PROJECT_ROOT}/backups/${ENVIRONMENT}"

    if [ -f "${BACKUP_DIR}/previous-image.txt" ]; then
        PREVIOUS_IMAGE=$(cat "${BACKUP_DIR}/previous-image.txt")

        if [ "$PREVIOUS_IMAGE" = "no-previous-deployment" ]; then
            error "No previous deployment found"
            exit 1
        fi

        log "Rolling back to previous image: $PREVIOUS_IMAGE"
    else
        # Use Kubernetes rollback history
        log "No backup found, using Kubernetes rollout history"
        ROLLBACK_TO="revision"
    fi
fi

# ============================================
# Create backup of current state
# ============================================
log "Backing up current state before rollback..."

BACKUP_DIR="${PROJECT_ROOT}/backups/${ENVIRONMENT}"
mkdir -p "$BACKUP_DIR"

kubectl get all -n "$KUBE_NAMESPACE" -o yaml > "${BACKUP_DIR}/pre-rollback-${ROLLBACK_TIMESTAMP}.yaml" 2>/dev/null || true

# ============================================
# Execute rollback
# ============================================
log "Executing rollback..."

if [ "$ROLLBACK_TO" = "revision" ]; then
    # Rollback using Kubernetes revision history
    log "Rolling back deployment using revision history..."

    # Show rollout history
    kubectl rollout history deployment/freightliner -n "$KUBE_NAMESPACE"

    # Rollback to previous revision
    kubectl rollout undo deployment/freightliner -n "$KUBE_NAMESPACE"

    # Wait for rollback to complete
    kubectl rollout status deployment/freightliner -n "$KUBE_NAMESPACE" --timeout=5m

elif [ -n "${PREVIOUS_IMAGE:-}" ]; then
    # Rollback to specific image
    log "Rolling back to specific image: $PREVIOUS_IMAGE"

    if helm list -n "$KUBE_NAMESPACE" | grep -q freightliner; then
        # Use Helm rollback
        PREVIOUS_REVISION=$(helm history freightliner -n "$KUBE_NAMESPACE" --max 2 -o json | jq -r '.[1].revision' 2>/dev/null || echo "")

        if [ -n "$PREVIOUS_REVISION" ]; then
            log "Rolling back Helm release to revision $PREVIOUS_REVISION"
            helm rollback freightliner "$PREVIOUS_REVISION" -n "$KUBE_NAMESPACE" --wait --timeout 5m
        else
            error "Cannot determine Helm revision to rollback to"
            exit 1
        fi
    else
        # Update image directly
        kubectl set image deployment/freightliner freightliner="$PREVIOUS_IMAGE" -n "$KUBE_NAMESPACE"
        kubectl rollout status deployment/freightliner -n "$KUBE_NAMESPACE" --timeout=5m
    fi
else
    error "No valid rollback target found"
    exit 1
fi

# ============================================
# Post-rollback validation
# ============================================
log "Running post-rollback validation..."

# Wait for pods to be ready
log "Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app=freightliner -n "$KUBE_NAMESPACE" --timeout=300s || {
    error "Pods did not become ready after rollback"

    # Show pod status
    kubectl get pods -n "$KUBE_NAMESPACE" -l app=freightliner
    kubectl describe pods -n "$KUBE_NAMESPACE" -l app=freightliner

    exit 1
}

# Verify pods are running
READY_PODS=$(kubectl get pods -n "$KUBE_NAMESPACE" -l app=freightliner --field-selector=status.phase=Running --no-headers | wc -l)
log "Ready pods after rollback: $READY_PODS"

if [ "$READY_PODS" -lt 1 ]; then
    error "No ready pods found after rollback"
    exit 1
fi

# Run health check
if [ -f "${SCRIPT_DIR}/health-check.sh" ]; then
    log "Running health check..."
    bash "${SCRIPT_DIR}/health-check.sh" "$ENVIRONMENT" || {
        error "Health check failed after rollback"
        exit 1
    }
fi

# ============================================
# Get current state
# ============================================
log "Current deployment state:"

CURRENT_IMAGE=$(kubectl get deployment freightliner -n "$KUBE_NAMESPACE" -o jsonpath='{.spec.template.spec.containers[0].image}')
log "Current image: $CURRENT_IMAGE"

kubectl get pods -n "$KUBE_NAMESPACE" -l app=freightliner

# ============================================
# Success notification
# ============================================
log "Rollback completed successfully!"
log "Environment: $ENVIRONMENT"
log "Current image: $CURRENT_IMAGE"
log "Timestamp: $ROLLBACK_TIMESTAMP"

# Send notification
if [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"⚠️ Rollback completed: freightliner ${ENVIRONMENT} to ${CURRENT_IMAGE}\"}" \
        "$SLACK_WEBHOOK_URL" || true
fi

log "Rollback log saved to: $ROLLBACK_LOG"

exit 0
