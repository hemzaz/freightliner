#!/bin/bash
# Freightliner Monitoring Stack Validation Script
# Tests all monitoring components are working correctly

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
PASSED=0
FAILED=0
WARNINGS=0

# Test results
log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    ((WARNINGS++))
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Test functions
test_docker_running() {
    log_test "Checking Docker daemon"
    if docker info &> /dev/null; then
        log_pass "Docker daemon is running"
        return 0
    else
        log_fail "Docker daemon is not running"
        return 1
    fi
}

test_compose_file() {
    log_test "Validating docker-compose file"
    if [ -f "docker-compose.monitoring.yml" ]; then
        if docker-compose -f docker-compose.monitoring.yml config > /dev/null 2>&1; then
            log_pass "docker-compose.monitoring.yml is valid"
            return 0
        else
            log_fail "docker-compose.monitoring.yml has syntax errors"
            return 1
        fi
    else
        log_fail "docker-compose.monitoring.yml not found"
        return 1
    fi
}

test_env_file() {
    log_test "Checking environment configuration"
    if [ -f ".env" ]; then
        log_pass ".env file exists"
        return 0
    else
        log_warn ".env file not found (will use defaults)"
        return 0
    fi
}

test_prometheus_running() {
    log_test "Checking Prometheus service"
    if docker ps | grep -q freightliner-prometheus; then
        log_pass "Prometheus container is running"
        return 0
    else
        log_fail "Prometheus container is not running"
        return 1
    fi
}

test_prometheus_health() {
    log_test "Checking Prometheus health endpoint"
    if curl -sf http://localhost:9090/-/healthy > /dev/null 2>&1; then
        log_pass "Prometheus is healthy"
        return 0
    else
        log_fail "Prometheus health check failed"
        return 1
    fi
}

test_prometheus_targets() {
    log_test "Checking Prometheus targets"
    local response=$(curl -sf http://localhost:9090/api/v1/targets 2>/dev/null)
    if [ -n "$response" ]; then
        local up_count=$(echo "$response" | grep -o '"health":"up"' | wc -l)
        local down_count=$(echo "$response" | grep -o '"health":"down"' | wc -l)

        if [ "$down_count" -eq 0 ]; then
            log_pass "All Prometheus targets are up ($up_count targets)"
            return 0
        else
            log_warn "$down_count Prometheus targets are down, $up_count are up"
            return 0
        fi
    else
        log_fail "Cannot query Prometheus targets"
        return 1
    fi
}

test_grafana_running() {
    log_test "Checking Grafana service"
    if docker ps | grep -q freightliner-grafana; then
        log_pass "Grafana container is running"
        return 0
    else
        log_fail "Grafana container is not running"
        return 1
    fi
}

test_grafana_health() {
    log_test "Checking Grafana health endpoint"
    if curl -sf http://localhost:3000/api/health > /dev/null 2>&1; then
        log_pass "Grafana is healthy"
        return 0
    else
        log_fail "Grafana health check failed"
        return 1
    fi
}

test_grafana_datasource() {
    log_test "Checking Grafana datasources"
    local grafana_user="${GRAFANA_ADMIN_USER:-admin}"
    local grafana_pass="${GRAFANA_ADMIN_PASSWORD:-admin}"
    local response=$(curl -sf -u "${grafana_user}:${grafana_pass}" http://localhost:3000/api/datasources 2>/dev/null)
    if [ -n "$response" ]; then
        if echo "$response" | grep -q "Prometheus"; then
            log_pass "Grafana Prometheus datasource is configured"
            return 0
        else
            log_warn "Grafana Prometheus datasource not found (may need manual setup)"
            return 0
        fi
    else
        log_warn "Cannot query Grafana datasources (authentication may be required)"
        return 0
    fi
}

test_api_running() {
    log_test "Checking Freightliner API service"
    if docker ps | grep -q freightliner-api; then
        log_pass "Freightliner API container is running"
        return 0
    else
        log_warn "Freightliner API container is not running (may need to be built)"
        return 0
    fi
}

test_api_health() {
    log_test "Checking Freightliner API health endpoint"
    if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        log_pass "Freightliner API is healthy"
        return 0
    else
        log_warn "Freightliner API health check failed (may not be implemented yet)"
        return 0
    fi
}

test_metrics_endpoint() {
    log_test "Checking metrics endpoint"
    if curl -sf http://localhost:2112/metrics > /dev/null 2>&1; then
        local metric_count=$(curl -sf http://localhost:2112/metrics | grep -c "^freightliner_" || echo "0")
        log_pass "Metrics endpoint is accessible ($metric_count freightliner metrics)"
        return 0
    else
        log_warn "Metrics endpoint not accessible (API may not be running)"
        return 0
    fi
}

test_network() {
    log_test "Checking Docker network"
    if docker network ls | grep -q freightliner-monitoring; then
        log_pass "Docker network 'freightliner-monitoring' exists"
        return 0
    else
        log_fail "Docker network 'freightliner-monitoring' not found"
        return 1
    fi
}

test_volumes() {
    log_test "Checking Docker volumes"
    local volume_count=$(docker volume ls | grep -c freightliner || echo "0")
    if [ "$volume_count" -gt 0 ]; then
        log_pass "Found $volume_count freightliner volumes"
        return 0
    else
        log_warn "No freightliner volumes found (will be created on first run)"
        return 0
    fi
}

test_config_files() {
    log_test "Checking configuration files"
    local all_present=true

    declare -a config_files=(
        "monitoring/prometheus/prometheus-local.yml"
        "monitoring/prometheus/alert-rules.yml"
        "monitoring/grafana/datasources/prometheus.yml"
        "monitoring/grafana/dashboards/dashboards.yml"
    )

    for file in "${config_files[@]}"; do
        if [ ! -f "$file" ]; then
            log_warn "Config file missing: $file"
            all_present=false
        fi
    done

    if [ "$all_present" = true ]; then
        log_pass "All configuration files present"
        return 0
    else
        return 0
    fi
}

test_port_availability() {
    log_test "Checking port availability"

    declare -A ports=(
        [3000]="Grafana"
        [8080]="API"
        [9090]="Prometheus"
        [2112]="Metrics"
    )

    local conflicts=0
    for port in "${!ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_info "Port $port (${ports[$port]}) is in use"
        fi
    done

    log_pass "Port availability check complete"
    return 0
}

# Summary
print_summary() {
    echo ""
    echo "=================================="
    echo "Validation Summary"
    echo "=================================="
    echo -e "${GREEN}Passed:${NC}   $PASSED"
    echo -e "${RED}Failed:${NC}   $FAILED"
    echo -e "${YELLOW}Warnings:${NC} $WARNINGS"
    echo "=================================="
    echo ""

    if [ $FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ All critical tests passed${NC}"
        if [ $WARNINGS -gt 0 ]; then
            echo -e "${YELLOW}⚠ Some warnings detected - review above${NC}"
        fi
        return 0
    else
        echo -e "${RED}✗ Some tests failed - review above${NC}"
        return 1
    fi
}

# Main execution
main() {
    echo "=========================================="
    echo "Freightliner Monitoring Stack Validation"
    echo "=========================================="
    echo ""

    # Prerequisites
    test_docker_running || exit 1
    test_compose_file || exit 1
    test_env_file

    echo ""
    log_info "Testing infrastructure..."
    test_network
    test_volumes
    test_config_files
    test_port_availability

    echo ""
    log_info "Testing Prometheus..."
    test_prometheus_running
    test_prometheus_health
    test_prometheus_targets

    echo ""
    log_info "Testing Grafana..."
    test_grafana_running
    test_grafana_health
    test_grafana_datasource

    echo ""
    log_info "Testing Freightliner API..."
    test_api_running
    test_api_health
    test_metrics_endpoint

    print_summary
}

# Run main
main
exit $?
