#!/bin/bash

# Test Performance Monitor - Tracks and alerts on test performance degradation
# Monitors test execution times, timeout rates, and resource usage

set -euo pipefail

# Configuration
METRICS_DIR="./test-metrics"
BASELINE_FILE="$METRICS_DIR/baseline.json"
ALERT_THRESHOLD=20  # Percentage increase in execution time to trigger alert
TIMEOUT_THRESHOLD=5 # Maximum acceptable timeout percentage
MAX_EXECUTION_TIME=900 # 15 minutes in seconds

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create metrics directory
ensure_metrics_dir() {
    mkdir -p "$METRICS_DIR"
}

# Monitor test execution
monitor_test_execution() {
    local test_type="$1"
    local start_time=$(date +%s)
    local test_log="$METRICS_DIR/test_execution_$(date +%Y%m%d_%H%M%S).log"
    
    log_info "Starting performance monitoring for $test_type tests..."
    
    # Track system resources before test
    local cpu_before=$(top -l 1 -n 0 | grep "CPU usage" | awk '{print $3}' | sed 's/%//' || echo "0")
    local mem_before=$(top -l 1 -n 0 | grep "PhysMem" | awk '{print $2}' | sed 's/M//' || echo "0")
    
    # Execute tests with monitoring
    case "$test_type" in
        "integration")
            log_info "Running integration tests with monitoring..."
            {
                echo "=== Test Execution Started at $(date) ==="
                echo "Initial CPU: ${cpu_before}%"
                echo "Initial Memory: ${mem_before}M"
                echo ""
                
                # Run tests with timeout
                if timeout $MAX_EXECUTION_TIME go test -v -timeout=15m -run Integration ./pkg/testing/load/... ./pkg/tree/... ; then
                    echo "TEST_RESULT=PASSED"
                else
                    local exit_code=$?
                    if [ $exit_code -eq 124 ]; then
                        echo "TEST_RESULT=TIMEOUT"
                        log_error "Tests timed out after $MAX_EXECUTION_TIME seconds"
                    else
                        echo "TEST_RESULT=FAILED"
                    fi
                fi
                
                echo ""
                echo "=== Test Execution Completed at $(date) ==="
            } | tee "$test_log"
            ;;
        "unit")
            log_info "Running unit tests with monitoring..."
            {
                echo "=== Unit Test Execution Started at $(date) ==="
                
                if timeout 300 go test -v -short -timeout=5m ./...; then
                    echo "TEST_RESULT=PASSED"
                else
                    echo "TEST_RESULT=FAILED"
                fi
                
                echo "=== Unit Test Execution Completed at $(date) ==="
            } | tee "$test_log"
            ;;
        *)
            log_error "Unknown test type: $test_type"
            return 1
            ;;
    esac
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Track system resources after test
    local cpu_after=$(top -l 1 -n 0 | grep "CPU usage" | awk '{print $3}' | sed 's/%//' || echo "0")
    local mem_after=$(top -l 1 -n 0 | grep "PhysMem" | awk '{print $2}' | sed 's/M//' || echo "0")
    
    # Extract test result
    local test_result=$(grep "TEST_RESULT=" "$test_log" | cut -d'=' -f2)
    
    # Record metrics
    record_metrics "$test_type" "$duration" "$test_result" "$cpu_before" "$cpu_after" "$mem_before" "$mem_after"
    
    # Analyze performance
    analyze_performance "$test_type" "$duration" "$test_result"
}

# Record test metrics
record_metrics() {
    local test_type="$1"
    local duration="$2"
    local result="$3"
    local cpu_before="$4"
    local cpu_after="$5"
    local mem_before="$6"
    local mem_after="$7"
    
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local metrics_file="$METRICS_DIR/metrics_$(date +%Y%m%d).json"
    
    # Create metrics entry
    local entry=$(cat <<EOF
{
  "timestamp": "$timestamp",
  "test_type": "$test_type",
  "duration_seconds": $duration,
  "result": "$result",
  "resources": {
    "cpu_before": "$cpu_before",
    "cpu_after": "$cpu_after",
    "memory_before": "$mem_before",
    "memory_after": "$mem_after"
  }
}
EOF
)
    
    # Append to metrics file
    if [ -f "$metrics_file" ]; then
        # Read existing array and add new entry
        local temp_file=$(mktemp)
        jq ". += [$entry]" "$metrics_file" > "$temp_file" 2>/dev/null || echo "[$entry]" > "$temp_file"
        mv "$temp_file" "$metrics_file"
    else
        echo "[$entry]" > "$metrics_file"
    fi
    
    log_info "Recorded metrics: duration=${duration}s, result=$result"
}

# Analyze test performance against baseline
analyze_performance() {
    local test_type="$1"
    local duration="$2"
    local result="$3"
    
    log_info "Analyzing performance for $test_type tests..."
    
    # Check for timeout
    if [ "$result" = "TIMEOUT" ]; then
        log_error "❌ Test execution timed out (>${MAX_EXECUTION_TIME}s)"
        send_alert "TIMEOUT" "$test_type" "$duration"
        return 1
    fi
    
    # Load baseline if it exists
    if [ -f "$BASELINE_FILE" ]; then
        local baseline_duration=$(jq -r ".${test_type}.average_duration // 300" "$BASELINE_FILE" 2>/dev/null || echo "300")
        local performance_change=$(( (duration - baseline_duration) * 100 / baseline_duration ))
        
        log_info "Baseline duration: ${baseline_duration}s, Current: ${duration}s, Change: ${performance_change}%"
        
        if [ $performance_change -gt $ALERT_THRESHOLD ]; then
            log_warning "⚠️  Performance regression detected: ${performance_change}% slower than baseline"
            send_alert "REGRESSION" "$test_type" "$duration" "$baseline_duration" "$performance_change"
        elif [ $performance_change -lt -10 ]; then
            log_success "🚀 Performance improvement: ${performance_change}% faster than baseline"
        else
            log_success "✅ Performance within acceptable range"
        fi
    else
        log_info "No baseline found, creating initial baseline..."
        create_baseline "$test_type" "$duration"
    fi
}

# Create performance baseline
create_baseline() {
    local test_type="$1"
    local duration="$2"
    
    local baseline_entry=$(cat <<EOF
{
  "$test_type": {
    "average_duration": $duration,
    "last_updated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "sample_count": 1
  }
}
EOF
)
    
    if [ -f "$BASELINE_FILE" ]; then
        local temp_file=$(mktemp)
        jq ". * $baseline_entry" "$BASELINE_FILE" > "$temp_file"
        mv "$temp_file" "$BASELINE_FILE"
    else
        echo "$baseline_entry" > "$BASELINE_FILE"
    fi
    
    log_success "Created baseline for $test_type: ${duration}s"
}

# Send performance alert
send_alert() {
    local alert_type="$1"
    local test_type="$2"
    local duration="$3"
    local baseline_duration="${4:-0}"
    local change_percent="${5:-0}"
    
    local alert_file="$METRICS_DIR/alerts_$(date +%Y%m%d).log"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    case "$alert_type" in
        "TIMEOUT")
            echo "[$timestamp] TIMEOUT: $test_type tests timed out after ${duration}s" >> "$alert_file"
            log_error "ALERT: Test timeout detected"
            ;;
        "REGRESSION")
            echo "[$timestamp] REGRESSION: $test_type tests ${change_percent}% slower (${duration}s vs ${baseline_duration}s baseline)" >> "$alert_file"
            log_warning "ALERT: Performance regression detected"
            ;;
    esac
    
    # In a real CI environment, this could send to Slack, email, or other alerting systems
    if [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
        send_slack_alert "$alert_type" "$test_type" "$duration" "$baseline_duration" "$change_percent"
    fi
}

# Send Slack alert (optional)
send_slack_alert() {
    local alert_type="$1"
    local test_type="$2"
    local duration="$3"
    local baseline_duration="${4:-0}"
    local change_percent="${5:-0}"
    
    local message
    case "$alert_type" in
        "TIMEOUT")
            message="🚨 Test Timeout Alert: $test_type tests timed out after ${duration}s"
            ;;
        "REGRESSION")
            message="⚠️ Performance Regression: $test_type tests are ${change_percent}% slower (${duration}s vs ${baseline_duration}s baseline)"
            ;;
    esac
    
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"$message\"}" \
        "$SLACK_WEBHOOK_URL" || log_warning "Failed to send Slack alert"
}

# Generate performance report
generate_report() {
    local report_file="$METRICS_DIR/performance_report_$(date +%Y%m%d).md"
    
    log_info "Generating performance report..."
    
    cat > "$report_file" <<EOF
# Test Performance Report - $(date +"%Y-%m-%d")

## Summary

Generated at: $(date)

## Recent Test Executions

EOF
    
    # Add recent metrics if available
    local recent_metrics="$METRICS_DIR/metrics_$(date +%Y%m%d).json"
    if [ -f "$recent_metrics" ]; then
        echo "### Today's Executions" >> "$report_file"
        echo "" >> "$report_file"
        
        jq -r '.[] | "- **\(.test_type)**: \(.duration_seconds)s (\(.result))"' "$recent_metrics" >> "$report_file" 2>/dev/null || echo "No metrics available" >> "$report_file"
    fi
    
    # Add alerts if any
    local alerts_file="$METRICS_DIR/alerts_$(date +%Y%m%d).log"
    if [ -f "$alerts_file" ] && [ -s "$alerts_file" ]; then
        echo "" >> "$report_file"
        echo "## Alerts" >> "$report_file"
        echo "" >> "$report_file"
        echo '```' >> "$report_file"
        cat "$alerts_file" >> "$report_file"
        echo '```' >> "$report_file"
    fi
    
    log_success "Performance report generated: $report_file"
}

# Clean up old metrics (keep last 30 days)
cleanup_old_metrics() {
    log_info "Cleaning up old metrics files..."
    
    find "$METRICS_DIR" -name "metrics_*.json" -mtime +30 -delete 2>/dev/null || true
    find "$METRICS_DIR" -name "alerts_*.log" -mtime +30 -delete 2>/dev/null || true
    find "$METRICS_DIR" -name "test_execution_*.log" -mtime +7 -delete 2>/dev/null || true
    
    log_success "Old metrics cleaned up"
}

# Main execution
main() {
    local test_type="${1:-integration}"
    
    log_info "Test Performance Monitor - Starting monitoring for $test_type tests"
    
    ensure_metrics_dir
    cleanup_old_metrics
    
    # Run test monitoring
    if monitor_test_execution "$test_type"; then
        log_success "Test monitoring completed successfully"
    else
        log_error "Test monitoring completed with issues"
    fi
    
    # Generate report
    generate_report
    
    log_info "Performance monitoring complete. Check $METRICS_DIR for detailed metrics."
}

# Parse command line arguments
case "${1:-integration}" in
    integration|unit)
        main "$1"
        ;;
    --report-only)
        ensure_metrics_dir
        generate_report
        ;;
    --cleanup)
        ensure_metrics_dir
        cleanup_old_metrics
        ;;
    --help|-h)
        echo "Usage: $0 [integration|unit|--report-only|--cleanup|--help]"
        echo ""
        echo "Options:"
        echo "  integration    Monitor integration test execution (default)"
        echo "  unit          Monitor unit test execution" 
        echo "  --report-only Generate performance report only"
        echo "  --cleanup     Clean up old metrics files"
        echo "  --help        Show this help"
        echo ""
        echo "Environment Variables:"
        echo "  SLACK_WEBHOOK_URL  Slack webhook for alerts (optional)"
        ;;
    *)
        log_error "Unknown option: $1"
        echo "Use $0 --help for usage information"
        exit 1
        ;;
esac