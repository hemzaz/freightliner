#!/bin/bash
# Production deployment script with validation and rollback capabilities
set -euo pipefail

# ============================================
# Configuration
# ============================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DEPLOYMENT_TIMESTAMP=$(date +%Y%m%d-%H%M%S)
DEPLOYMENT_LOG="${PROJECT_ROOT}/logs/deployment-${DEPLOYMENT_TIMESTAMP}.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ============================================
# Functions
# ============================================
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*" | tee -a "$DEPLOYMENT_LOG"
}

error() {
    echo -e "${RED}[ERROR] $*${NC}" | tee -a "$DEPLOYMENT_LOG" >&2
}

warn() {
    echo -e "${YELLOW}[WARN] $*${NC}" | tee -a "$DEPLOYMENT_LOG"
}

info() {
    echo -e "${BLUE}[INFO] $*${NC}" | tee -a "$DEPLOYMENT_LOG"
}

# Create log directory
mkdir -p "${PROJECT_ROOT}/logs"

# ============================================
# Parse arguments
# ============================================
ENVIRONMENT="${1:-dev}"
VERSION="${2:-latest}"
SKIP_TESTS="${SKIP_TESTS:-false}"
DRY_RUN="${DRY_RUN:-false}"
AUTO_ROLLBACK="${AUTO_ROLLBACK:-true}"

log "Starting deployment to ${ENVIRONMENT} environment"
log "Version: ${VERSION}"
log "Dry run: ${DRY_RUN}"

# ============================================
# Pre-deployment validation
# ============================================
log "Running pre-deployment validation..."

# Check required tools
for cmd in docker kubectl helm; do
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
        error "Invalid environment: $ENVIRONMENT. Must be dev, staging, or production"
        exit 1
        ;;
esac

# Check if running in CI
if [ -n "${CI:-}" ]; then
    log "Running in CI environment"
    # Authenticate with registries
    if [ -n "${DOCKER_PASSWORD:-}" ]; then
        echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME:-}" --password-stdin "${DOCKER_REGISTRY:-ghcr.io}"
    fi
fi

# ============================================
# Build Docker image
# ============================================
log "Building Docker image..."

DOCKER_IMAGE="${DOCKER_REGISTRY:-ghcr.io/hemzaz}/freightliner:${VERSION}"
DOCKER_CACHE_IMAGE="${DOCKER_REGISTRY:-ghcr.io/hemzaz}/freightliner:cache"

if [ "$DRY_RUN" = "false" ]; then
    docker buildx build \
        --platform linux/amd64,linux/arm64 \
        --tag "$DOCKER_IMAGE" \
        --tag "${DOCKER_REGISTRY:-ghcr.io/hemzaz}/freightliner:latest" \
        --cache-from "type=registry,ref=${DOCKER_CACHE_IMAGE}" \
        --cache-to "type=registry,ref=${DOCKER_CACHE_IMAGE},mode=max" \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --file Dockerfile.optimized \
        --push \
        "${PROJECT_ROOT}"

    log "Docker image built and pushed: $DOCKER_IMAGE"
else
    info "Dry run: Skipping Docker build"
fi

# ============================================
# Security scanning
# ============================================
log "Running security scan on Docker image..."

if [ "$DRY_RUN" = "false" ]; then
    if command -v trivy &> /dev/null; then
        trivy image --severity HIGH,CRITICAL "$DOCKER_IMAGE" || warn "Security scan found issues"
    else
        warn "Trivy not installed, skipping security scan"
    fi
fi

# ============================================
# Run tests
# ============================================
if [ "$SKIP_TESTS" = "false" ]; then
    log "Running tests in container..."

    if [ "$DRY_RUN" = "false" ]; then
        docker run --rm "$DOCKER_IMAGE" --version || {
            error "Container health check failed"
            exit 1
        }
    fi
fi

# ============================================
# Backup current deployment
# ============================================
log "Backing up current deployment..."

BACKUP_DIR="${PROJECT_ROOT}/backups/${ENVIRONMENT}"
mkdir -p "$BACKUP_DIR"

if [ "$DRY_RUN" = "false" ]; then
    # Save current Kubernetes state
    kubectl get all -n "freightliner-${ENVIRONMENT}" -o yaml > "${BACKUP_DIR}/backup-${DEPLOYMENT_TIMESTAMP}.yaml" 2>/dev/null || true

    # Save current image tag
    kubectl get deployment freightliner -n "freightliner-${ENVIRONMENT}" -o jsonpath='{.spec.template.spec.containers[0].image}' > "${BACKUP_DIR}/previous-image.txt" 2>/dev/null || echo "no-previous-deployment" > "${BACKUP_DIR}/previous-image.txt"

    log "Backup saved to ${BACKUP_DIR}/backup-${DEPLOYMENT_TIMESTAMP}.yaml"
fi

# ============================================
# Deploy to Kubernetes
# ============================================
log "Deploying to Kubernetes cluster..."

KUBE_NAMESPACE="freightliner-${ENVIRONMENT}"
KUBE_CONTEXT="${KUBE_CONTEXT:-$(kubectl config current-context)}"

log "Using context: $KUBE_CONTEXT"
log "Namespace: $KUBE_NAMESPACE"

if [ "$DRY_RUN" = "false" ]; then
    # Create namespace if it doesn't exist
    kubectl create namespace "$KUBE_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

    # Deploy using Helm (preferred) or kubectl
    if [ -f "${PROJECT_ROOT}/deployments/helm/freightliner/Chart.yaml" ]; then
        log "Deploying with Helm..."
        helm upgrade --install freightliner \
            "${PROJECT_ROOT}/deployments/helm/freightliner" \
            --namespace "$KUBE_NAMESPACE" \
            --set image.tag="$VERSION" \
            --set environment="$ENVIRONMENT" \
            --values "${PROJECT_ROOT}/deployments/helm/freightliner/values-${ENVIRONMENT}.yaml" \
            --wait \
            --timeout 5m
    elif [ -f "${PROJECT_ROOT}/deployments/kubernetes/deployment.yaml" ]; then
        log "Deploying with kubectl..."

        # Update image in manifests
        sed "s|image:.*freightliner:.*|image: $DOCKER_IMAGE|g" \
            "${PROJECT_ROOT}/deployments/kubernetes/deployment.yaml" | \
            kubectl apply -n "$KUBE_NAMESPACE" -f -

        # Apply other manifests
        for manifest in service configmap ingress; do
            if [ -f "${PROJECT_ROOT}/deployments/kubernetes/${manifest}.yaml" ]; then
                kubectl apply -n "$KUBE_NAMESPACE" -f "${PROJECT_ROOT}/deployments/kubernetes/${manifest}.yaml"
            fi
        done

        # Wait for rollout
        kubectl rollout status deployment/freightliner -n "$KUBE_NAMESPACE" --timeout=5m
    else
        error "No deployment manifests found"
        exit 1
    fi

    log "Deployment completed"
else
    info "Dry run: Skipping Kubernetes deployment"
fi

# ============================================
# Post-deployment validation
# ============================================
log "Running post-deployment validation..."

if [ "$DRY_RUN" = "false" ]; then
    # Run health check script
    if [ -f "${SCRIPT_DIR}/health-check.sh" ]; then
        bash "${SCRIPT_DIR}/health-check.sh" "$ENVIRONMENT" || {
            error "Health check failed"

            if [ "$AUTO_ROLLBACK" = "true" ]; then
                warn "Initiating automatic rollback..."
                bash "${SCRIPT_DIR}/rollback.sh" "$ENVIRONMENT"
                exit 1
            fi
        }
    fi

    # Verify pods are running
    READY_PODS=$(kubectl get pods -n "$KUBE_NAMESPACE" -l app=freightliner --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l)
    log "Ready pods: $READY_PODS"

    if [ "$READY_PODS" -lt 1 ]; then
        error "No ready pods found"

        if [ "$AUTO_ROLLBACK" = "true" ]; then
            warn "Initiating automatic rollback..."
            bash "${SCRIPT_DIR}/rollback.sh" "$ENVIRONMENT"
            exit 1
        fi
    fi
fi

# ============================================
# Success notification
# ============================================
log "Deployment successful!"
log "Environment: $ENVIRONMENT"
log "Version: $VERSION"
log "Image: $DOCKER_IMAGE"
log "Timestamp: $DEPLOYMENT_TIMESTAMP"

# Send notification (optional)
if [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"✅ Deployment successful: freightliner ${VERSION} to ${ENVIRONMENT}\"}" \
        "$SLACK_WEBHOOK_URL" || true
fi

log "Deployment log saved to: $DEPLOYMENT_LOG"

exit 0
