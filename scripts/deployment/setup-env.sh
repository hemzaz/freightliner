#!/bin/bash
# Environment setup script for deployment
set -euo pipefail

# ============================================
# Configuration
# ============================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*"
}

warn() {
    echo -e "${YELLOW}[WARN] $*${NC}"
}

# ============================================
# Parse arguments
# ============================================
ENVIRONMENT="${1:-}"

if [ -z "$ENVIRONMENT" ]; then
    echo "Usage: $0 <environment>"
    echo "  environment: dev, staging, or production"
    exit 1
fi

log "Setting up environment: $ENVIRONMENT"

# ============================================
# Install required tools
# ============================================
log "Checking required tools..."

# Docker
if ! command -v docker &> /dev/null; then
    warn "Docker not installed. Installing..."
    curl -fsSL https://get.docker.com | sh
fi

# kubectl
if ! command -v kubectl &> /dev/null; then
    warn "kubectl not installed. Installing..."
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    chmod +x kubectl
    sudo mv kubectl /usr/local/bin/
fi

# Helm
if ! command -v helm &> /dev/null; then
    warn "Helm not installed. Installing..."
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
fi

# Trivy
if ! command -v trivy &> /dev/null; then
    warn "Trivy not installed. Installing..."
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
fi

log "All required tools installed"

# ============================================
# Configure Kubernetes context
# ============================================
log "Configuring Kubernetes context..."

case "$ENVIRONMENT" in
    dev)
        KUBE_CONTEXT="${KUBE_CONTEXT_DEV:-minikube}"
        ;;
    staging)
        KUBE_CONTEXT="${KUBE_CONTEXT_STAGING:-staging-cluster}"
        ;;
    production)
        KUBE_CONTEXT="${KUBE_CONTEXT_PROD:-production-cluster}"
        ;;
esac

if kubectl config get-contexts "$KUBE_CONTEXT" &>/dev/null; then
    kubectl config use-context "$KUBE_CONTEXT"
    log "Using Kubernetes context: $KUBE_CONTEXT"
else
    warn "Kubernetes context $KUBE_CONTEXT not found"
    warn "Available contexts:"
    kubectl config get-contexts
fi

# ============================================
# Create necessary directories
# ============================================
log "Creating necessary directories..."

mkdir -p "${PROJECT_ROOT}/logs"
mkdir -p "${PROJECT_ROOT}/backups/${ENVIRONMENT}"

# ============================================
# Load environment variables
# ============================================
log "Loading environment variables..."

ENV_FILE="${PROJECT_ROOT}/.env.${ENVIRONMENT}"

if [ -f "$ENV_FILE" ]; then
    set -a
    source "$ENV_FILE"
    set +a
    log "Environment variables loaded from $ENV_FILE"
else
    warn "Environment file not found: $ENV_FILE"
fi

# ============================================
# Verify Docker registry access
# ============================================
log "Verifying Docker registry access..."

DOCKER_REGISTRY="${DOCKER_REGISTRY:-ghcr.io}"

if [ -n "${DOCKER_PASSWORD:-}" ]; then
    echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME:-}" --password-stdin "$DOCKER_REGISTRY"
    log "Logged into Docker registry: $DOCKER_REGISTRY"
else
    warn "DOCKER_PASSWORD not set. Skipping Docker login."
fi

# ============================================
# Verify Kubernetes cluster access
# ============================================
log "Verifying Kubernetes cluster access..."

if kubectl cluster-info &>/dev/null; then
    log "Successfully connected to Kubernetes cluster"
    kubectl cluster-info
else
    warn "Cannot connect to Kubernetes cluster"
    exit 1
fi

# ============================================
# Create Kubernetes namespace
# ============================================
KUBE_NAMESPACE="freightliner-${ENVIRONMENT}"
log "Creating Kubernetes namespace: $KUBE_NAMESPACE"

kubectl create namespace "$KUBE_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# ============================================
# Setup monitoring
# ============================================
log "Setting up monitoring..."

# Create monitoring namespace
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -

# Install Prometheus Operator (if not exists)
if ! helm list -n monitoring | grep -q prometheus; then
    log "Installing Prometheus Operator..."
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update
    helm install prometheus prometheus-community/kube-prometheus-stack \
        -n monitoring \
        --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false
fi

log "Environment setup complete!"
log "You can now run: ./scripts/deployment/deploy.sh $ENVIRONMENT"
