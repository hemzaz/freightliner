#!/bin/bash
# Comprehensive health check and validation script
set -euo pipefail

# ============================================
# Configuration
# ============================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ============================================
# Functions
# ============================================
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR] $*${NC}" >&2
}

warn() {
    echo -e "${YELLOW}[WARN] $*${NC}"
}

check_passed() {
    echo -e "${GREEN}✓${NC} $*"
}

check_failed() {
    echo -e "${RED}✗${NC} $*"
    return 1
}

# ============================================
# Parse arguments
# ============================================
ENVIRONMENT="${1:-dev}"
TIMEOUT="${TIMEOUT:-300}"

log "Running health checks for ${ENVIRONMENT} environment"

KUBE_NAMESPACE="freightliner-${ENVIRONMENT}"
FAILED_CHECKS=0

# ============================================
# 1. Kubernetes Pod Health
# ============================================
log "Checking Kubernetes pods..."

POD_COUNT=$(kubectl get pods -n "$KUBE_NAMESPACE" -l app=freightliner --no-headers 2>/dev/null | wc -l)
READY_PODS=$(kubectl get pods -n "$KUBE_NAMESPACE" -l app=freightliner --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l)

if [ "$POD_COUNT" -gt 0 ] && [ "$READY_PODS" -eq "$POD_COUNT" ]; then
    check_passed "All pods are running ($READY_PODS/$POD_COUNT)"
else
    check_failed "Not all pods are running ($READY_PODS/$POD_COUNT)" || ((FAILED_CHECKS++))
    kubectl get pods -n "$KUBE_NAMESPACE" -l app=freightliner
fi

# ============================================
# 2. Container Health Checks
# ============================================
log "Checking container health..."

UNHEALTHY_PODS=$(kubectl get pods -n "$KUBE_NAMESPACE" -l app=freightliner -o json | \
    jq -r '.items[] | select(.status.containerStatuses[].ready == false) | .metadata.name' 2>/dev/null || echo "")

if [ -z "$UNHEALTHY_PODS" ]; then
    check_passed "All containers are healthy"
else
    check_failed "Unhealthy containers found: $UNHEALTHY_PODS" || ((FAILED_CHECKS++))

    for pod in $UNHEALTHY_PODS; do
        warn "Pod logs for $pod:"
        kubectl logs "$pod" -n "$KUBE_NAMESPACE" --tail=50 || true
    done
fi

# ============================================
# 3. Service Endpoint Health
# ============================================
log "Checking service endpoints..."

SERVICE_NAME="freightliner"
ENDPOINTS=$(kubectl get endpoints "$SERVICE_NAME" -n "$KUBE_NAMESPACE" -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null || echo "")

if [ -n "$ENDPOINTS" ]; then
    ENDPOINT_COUNT=$(echo "$ENDPOINTS" | wc -w)
    check_passed "Service has $ENDPOINT_COUNT endpoint(s)"
else
    check_failed "Service has no endpoints" || ((FAILED_CHECKS++))
fi

# ============================================
# 4. Application Health Endpoint
# ============================================
log "Checking application health endpoint..."

# Get service URL
SERVICE_TYPE=$(kubectl get service "$SERVICE_NAME" -n "$KUBE_NAMESPACE" -o jsonpath='{.spec.type}' 2>/dev/null || echo "")

if [ "$SERVICE_TYPE" = "LoadBalancer" ]; then
    SERVICE_URL=$(kubectl get service "$SERVICE_NAME" -n "$KUBE_NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")

    if [ -z "$SERVICE_URL" ]; then
        SERVICE_URL=$(kubectl get service "$SERVICE_NAME" -n "$KUBE_NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
    fi
elif [ "$SERVICE_TYPE" = "NodePort" ]; then
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}' 2>/dev/null || echo "")
    NODE_PORT=$(kubectl get service "$SERVICE_NAME" -n "$KUBE_NAMESPACE" -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "")
    SERVICE_URL="${NODE_IP}:${NODE_PORT}"
fi

if [ -n "$SERVICE_URL" ]; then
    # Try health endpoint
    if curl -sf "http://${SERVICE_URL}/health" &>/dev/null; then
        check_passed "Application health endpoint responding"
    elif curl -sf "http://${SERVICE_URL}/healthz" &>/dev/null; then
        check_passed "Application healthz endpoint responding"
    else
        warn "Health endpoint not accessible (might not be exposed)"
    fi
else
    warn "Cannot determine service URL (ClusterIP or not exposed)"
fi

# ============================================
# 5. Resource Usage Check
# ============================================
log "Checking resource usage..."

PODS=$(kubectl get pods -n "$KUBE_NAMESPACE" -l app=freightliner -o json 2>/dev/null || echo "{}")

# Check CPU and memory limits
CPU_LIMITS=$(echo "$PODS" | jq -r '.items[].spec.containers[].resources.limits.cpu // "not-set"' | head -1)
MEM_LIMITS=$(echo "$PODS" | jq -r '.items[].spec.containers[].resources.limits.memory // "not-set"' | head -1)

if [ "$CPU_LIMITS" != "not-set" ] && [ "$MEM_LIMITS" != "not-set" ]; then
    check_passed "Resource limits configured (CPU: $CPU_LIMITS, Memory: $MEM_LIMITS)"
else
    warn "Resource limits not configured"
fi

# ============================================
# 6. ConfigMap and Secret Health
# ============================================
log "Checking configuration..."

if kubectl get configmap freightliner-config -n "$KUBE_NAMESPACE" &>/dev/null; then
    check_passed "ConfigMap exists"
else
    warn "ConfigMap not found (might not be required)"
fi

if kubectl get secret freightliner-secrets -n "$KUBE_NAMESPACE" &>/dev/null; then
    check_passed "Secrets exist"
else
    warn "Secrets not found (might not be required)"
fi

# ============================================
# 7. Network Policy Check
# ============================================
log "Checking network policies..."

NETPOL_COUNT=$(kubectl get networkpolicies -n "$KUBE_NAMESPACE" --no-headers 2>/dev/null | wc -l)

if [ "$NETPOL_COUNT" -gt 0 ]; then
    check_passed "Network policies configured ($NETPOL_COUNT)"
else
    warn "No network policies found (consider adding for security)"
fi

# ============================================
# 8. Recent Events Check
# ============================================
log "Checking recent events..."

ERROR_EVENTS=$(kubectl get events -n "$KUBE_NAMESPACE" --field-selector type=Warning --sort-by='.lastTimestamp' 2>/dev/null | tail -5)

if [ -z "$ERROR_EVENTS" ]; then
    check_passed "No recent error events"
else
    warn "Recent warning events found:"
    echo "$ERROR_EVENTS"
fi

# ============================================
# 9. Metrics Check (if Prometheus available)
# ============================================
log "Checking metrics availability..."

if kubectl get service prometheus -n monitoring &>/dev/null 2>&1; then
    check_passed "Prometheus available for metrics"
else
    warn "Prometheus not available"
fi

# ============================================
# 10. Ingress/Route Check
# ============================================
log "Checking ingress configuration..."

if kubectl get ingress freightliner -n "$KUBE_NAMESPACE" &>/dev/null 2>&1; then
    INGRESS_HOST=$(kubectl get ingress freightliner -n "$KUBE_NAMESPACE" -o jsonpath='{.spec.rules[0].host}' 2>/dev/null || echo "")
    check_passed "Ingress configured: $INGRESS_HOST"
else
    warn "No ingress configured"
fi

# ============================================
# Summary
# ============================================
log "Health check summary:"
log "================================"

if [ "$FAILED_CHECKS" -eq 0 ]; then
    echo -e "${GREEN}✓ All health checks passed${NC}"
    exit 0
else
    echo -e "${RED}✗ $FAILED_CHECKS health check(s) failed${NC}"
    exit 1
fi
