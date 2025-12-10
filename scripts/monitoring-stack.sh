#!/bin/bash
# Freightliner Monitoring Stack Management Script
# Simplifies Docker Compose monitoring stack operations

set -euo pipefail

# Configuration
COMPOSE_FILE="docker-compose.monitoring.yml"
ENV_FILE=".env"
PROJECT_NAME="freightliner"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

check_dependencies() {
    log_info "Checking dependencies..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi

    log_success "All dependencies are installed"
}

setup_env() {
    if [ ! -f "$ENV_FILE" ]; then
        log_warning ".env file not found, copying from .env.monitoring"
        if [ -f ".env.monitoring" ]; then
            cp .env.monitoring "$ENV_FILE"
            log_success "Created .env file"
            log_info "Please review and customize $ENV_FILE before starting"
            return 1
        else
            log_error ".env.monitoring template not found"
            exit 1
        fi
    fi
    return 0
}

start_basic() {
    log_info "Starting basic monitoring stack..."
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d \
        prometheus grafana freightliner-api

    log_success "Monitoring stack started"
    show_urls
}

start_full() {
    log_info "Starting full monitoring stack with all services..."
    COMPOSE_PROFILES=full docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d

    log_success "Full monitoring stack started"
    show_urls_full
}

stop_stack() {
    log_info "Stopping monitoring stack..."
    docker-compose -f "$COMPOSE_FILE" down
    log_success "Monitoring stack stopped"
}

restart_stack() {
    log_info "Restarting monitoring stack..."
    docker-compose -f "$COMPOSE_FILE" restart
    log_success "Monitoring stack restarted"
}

show_status() {
    log_info "Monitoring stack status:"
    docker-compose -f "$COMPOSE_FILE" ps
}

show_logs() {
    local service="${1:-}"
    if [ -z "$service" ]; then
        docker-compose -f "$COMPOSE_FILE" logs -f
    else
        docker-compose -f "$COMPOSE_FILE" logs -f "$service"
    fi
}

show_urls() {
    echo ""
    log_info "Services are available at:"
    echo "  Grafana:          http://localhost:3000 (admin/admin)"
    echo "  Prometheus:       http://localhost:9090"
    echo "  Freightliner API: http://localhost:8080"
    echo "  Metrics:          http://localhost:2112/metrics"
    echo "  Health:           http://localhost:8080/health"
    echo ""
}

show_urls_full() {
    show_urls
    log_info "Additional services:"
    echo "  AlertManager:     http://localhost:9093"
    echo "  Node Exporter:    http://localhost:9100"
    echo "  cAdvisor:         http://localhost:8081"
    echo "  Redis:            localhost:6379"
    echo ""
}

check_health() {
    log_info "Checking service health..."

    local all_healthy=true

    # Check Prometheus
    if curl -sf http://localhost:9090/-/healthy > /dev/null 2>&1; then
        log_success "Prometheus is healthy"
    else
        log_error "Prometheus is not healthy"
        all_healthy=false
    fi

    # Check Grafana
    if curl -sf http://localhost:3000/api/health > /dev/null 2>&1; then
        log_success "Grafana is healthy"
    else
        log_error "Grafana is not healthy"
        all_healthy=false
    fi

    # Check Freightliner API
    if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        log_success "Freightliner API is healthy"
    else
        log_warning "Freightliner API is not responding (may not be built yet)"
    fi

    if [ "$all_healthy" = true ]; then
        log_success "All services are healthy"
    else
        log_warning "Some services are not healthy. Check logs with: $0 logs"
    fi
}

backup_data() {
    log_info "Backing up monitoring data..."

    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"

    # Backup Prometheus data
    docker run --rm -v freightliner-prometheus-data:/data \
        -v "$(pwd)/$backup_dir":/backup alpine \
        tar czf /backup/prometheus-data.tar.gz /data

    # Backup Grafana data
    docker run --rm -v freightliner-grafana-data:/data \
        -v "$(pwd)/$backup_dir":/backup alpine \
        tar czf /backup/grafana-data.tar.gz /data

    log_success "Backup created in $backup_dir"
}

clean_volumes() {
    log_warning "This will delete ALL monitoring data. Are you sure? (yes/no)"
    read -r response

    if [ "$response" != "yes" ]; then
        log_info "Cancelled"
        return
    fi

    log_info "Stopping services and removing volumes..."
    docker-compose -f "$COMPOSE_FILE" down -v
    log_success "All data cleaned"
}

build_api() {
    log_info "Building Freightliner API..."
    docker-compose -f "$COMPOSE_FILE" build freightliner-api
    log_success "API built successfully"
}

show_metrics() {
    log_info "Current metrics from Freightliner API:"
    if curl -sf http://localhost:2112/metrics > /dev/null 2>&1; then
        curl -s http://localhost:2112/metrics | grep "^freightliner_" | head -20
    else
        log_error "Metrics endpoint not available"
    fi
}

show_help() {
    cat << EOF
Freightliner Monitoring Stack Management

Usage: $0 <command> [options]

Commands:
  start         Start basic monitoring stack (Prometheus, Grafana, API)
  start-full    Start full stack with all services
  stop          Stop monitoring stack
  restart       Restart monitoring stack
  status        Show service status
  logs [svc]    Show logs (optionally for specific service)
  health        Check health of all services
  urls          Show service URLs
  build         Build Freightliner API image
  metrics       Show current metrics
  backup        Backup monitoring data
  clean         Remove all data volumes (WARNING: destructive)
  help          Show this help message

Examples:
  $0 start              # Start basic stack
  $0 start-full         # Start all services
  $0 logs prometheus    # Show Prometheus logs
  $0 health             # Check all services
  $0 backup             # Backup data

Service URLs:
  Grafana:          http://localhost:3000
  Prometheus:       http://localhost:9090
  Freightliner API: http://localhost:8080
  Metrics:          http://localhost:2112/metrics

EOF
}

# Main script
main() {
    check_dependencies

    local command="${1:-help}"
    shift || true

    case "$command" in
        start)
            if setup_env; then
                start_basic
            fi
            ;;
        start-full)
            if setup_env; then
                start_full
            fi
            ;;
        stop)
            stop_stack
            ;;
        restart)
            restart_stack
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$@"
            ;;
        health)
            check_health
            ;;
        urls)
            show_urls
            ;;
        build)
            build_api
            ;;
        metrics)
            show_metrics
            ;;
        backup)
            backup_data
            ;;
        clean)
            clean_volumes
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
