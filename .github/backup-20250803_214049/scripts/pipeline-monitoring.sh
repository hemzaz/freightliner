#!/bin/bash

# Pipeline Health Monitoring and Alerting Script
# Provides comprehensive monitoring, alerting, and SLA tracking
# for GitHub Actions CI/CD pipeline operations

set -euo pipefail

# ==============================================================================
# CONFIGURATION AND CONSTANTS
# ==============================================================================

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly WORKSPACE_DIR="${GITHUB_WORKSPACE:-$(pwd)}"
readonly MONITORING_STATE_DIR="${WORKSPACE_DIR}/.pipeline-monitoring"
readonly LOG_FILE="${MONITORING_STATE_DIR}/monitoring.log"

# Monitoring configuration
readonly HEALTH_CHECK_INTERVAL=30
readonly ALERT_THRESHOLD_ERROR_RATE=0.3  # 30% error rate triggers alert
readonly ALERT_THRESHOLD_DURATION=1800   # 30 minutes duration triggers alert
readonly SLA_SUCCESS_RATE_TARGET=0.95    # 95% success rate SLA
readonly SLA_DURATION_TARGET=1200        # 20 minutes duration SLA

# Webhook and notification configuration
readonly SLACK_WEBHOOK_URL="${SLACK_WEBHOOK_URL:-}"
readonly TEAMS_WEBHOOK_URL="${TEAMS_WEBHOOK_URL:-}"
readonly EMAIL_NOTIFICATION_ENABLED="${EMAIL_NOTIFICATION_ENABLED:-false}"

# Pipeline metrics storage
readonly METRICS_FILE="${MONITORING_STATE_DIR}/pipeline-metrics.json"
readonly ALERTS_FILE="${MONITORING_STATE_DIR}/alerts.json"
readonly SLA_TRACKING_FILE="${MONITORING_STATE_DIR}/sla-tracking.json"

# Colors and formatting
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# ==============================================================================
# INITIALIZATION AND UTILITIES
# ==============================================================================

# Initialize monitoring environment
init_monitoring_environment() {
    mkdir -p "${MONITORING_STATE_DIR}"
    touch "${LOG_FILE}"
    
    # Initialize metrics file if not exists
    if [[ ! -f "${METRICS_FILE}" ]]; then
        cat > "${METRICS_FILE}" << EOF
{
  "pipeline_runs": [],
  "job_statistics": {},
  "performance_metrics": {
    "average_duration": 0,
    "success_rate": 0,
    "error_rate": 0,
    "last_updated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  },
  "health_scores": {
    "overall": 100,
    "components": {}
  }
}
EOF
    fi
    
    # Initialize alerts file if not exists
    if [[ ! -f "${ALERTS_FILE}" ]]; then
        cat > "${ALERTS_FILE}" << EOF
{
  "active_alerts": [],
  "alert_history": [],
  "alert_rules": {
    "error_rate_threshold": $ALERT_THRESHOLD_ERROR_RATE,
    "duration_threshold": $ALERT_THRESHOLD_DURATION,
    "success_rate_threshold": $SLA_SUCCESS_RATE_TARGET
  }
}
EOF
    fi
    
    # Initialize SLA tracking file if not exists
    if [[ ! -f "${SLA_TRACKING_FILE}" ]]; then
        cat > "${SLA_TRACKING_FILE}" << EOF
{
  "sla_targets": {
    "success_rate": $SLA_SUCCESS_RATE_TARGET,
    "duration_target": $SLA_DURATION_TARGET
  },
  "current_period": {
    "start_date": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "total_runs": 0,
    "successful_runs": 0,
    "total_duration": 0,
    "success_rate": 0,
    "average_duration": 0
  },
  "historical_periods": []
}
EOF
    fi
}

# Enhanced logging with monitoring context
log_monitoring() {
    local level="$1"
    local component="$2"
    local message="$3"
    local metrics="${4:-}"
    
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    local log_entry
    if [[ -n "$metrics" ]]; then
        log_entry="{\"timestamp\":\"$timestamp\",\"level\":\"$level\",\"component\":\"$component\",\"message\":\"$message\",\"metrics\":$metrics}"
    else
        log_entry="{\"timestamp\":\"$timestamp\",\"level\":\"$level\",\"component\":\"$component\",\"message\":\"$message\"}"
    fi
    
    echo "$log_entry" >> "$LOG_FILE"
    
    # Also output to console with formatting
    local color=""
    local icon=""
    
    case "$level" in
        "ALERT")
            color="$RED"
            icon="🚨"
            ;;
        "WARNING")
            color="$YELLOW"
            icon="⚠️"
            ;;
        "INFO")
            color="$GREEN"
            icon="📊"
            ;;
        "DEBUG")
            color="$BLUE"
            icon="🔍"
            ;;
        "METRIC")
            color="$CYAN"
            icon="📈"
            ;;
    esac
    
    echo -e "${color}${icon} [${component}] ${message}${NC}"
}

log_alert() { log_monitoring "ALERT" "${2:-monitoring}" "$1" "${3:-}"; }
log_warning() { log_monitoring "WARNING" "${2:-monitoring}" "$1" "${3:-}"; }
log_info() { log_monitoring "INFO" "${2:-monitoring}" "$1" "${3:-}"; }
log_debug() { log_monitoring "DEBUG" "${2:-monitoring}" "$1" "${3:-}"; }
log_metric() { log_monitoring "METRIC" "${2:-monitoring}" "$1" "${3:-}"; }

# ==============================================================================
# METRICS COLLECTION
# ==============================================================================

# Record pipeline run metrics
record_pipeline_run() {
    local pipeline_id="${1:-${GITHUB_RUN_ID:-unknown}}"
    local status="$2"
    local duration="$3"
    local job_results="$4"  # JSON object with job results
    
    log_info "Recording pipeline run metrics" "metrics"
    
    local current_time
    current_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    local run_record
    run_record=$(cat << EOF
{
  "pipeline_id": "$pipeline_id",
  "timestamp": "$current_time",
  "status": "$status",
  "duration": $duration,
  "job_results": $job_results,
  "workflow": "${GITHUB_WORKFLOW:-unknown}",
  "repository": "${GITHUB_REPOSITORY:-unknown}",
  "ref": "${GITHUB_REF:-unknown}",
  "actor": "${GITHUB_ACTOR:-unknown}",
  "event_name": "${GITHUB_EVENT_NAME:-unknown}"
}
EOF
)
    
    # Update metrics file using jq if available
    if command -v jq >/dev/null 2>&1; then
        local temp_file
        temp_file=$(mktemp)
        
        # Add run record to pipeline_runs array
        jq --argjson run "$run_record" '.pipeline_runs += [$run]' "$METRICS_FILE" > "$temp_file"
        mv "$temp_file" "$METRICS_FILE"
        
        # Update performance metrics
        update_performance_metrics
        
        log_metric "Pipeline run recorded successfully" "metrics" "$run_record"
    else
        log_warning "jq not available, metrics recording limited" "metrics"
    fi
}

# Update performance metrics based on recorded runs
update_performance_metrics() {
    if ! command -v jq >/dev/null 2>&1; then
        return
    fi
    
    log_info "Updating performance metrics" "metrics"
    
    local temp_file
    temp_file=$(mktemp)
    
    # Calculate metrics from pipeline runs
    jq '
    {
      pipeline_runs: .pipeline_runs,
      job_statistics: (
        .pipeline_runs | 
        group_by(.workflow) | 
        map({
          key: .[0].workflow,
          value: {
            total_runs: length,
            successful_runs: map(select(.status == "success")) | length,
            failed_runs: map(select(.status == "failure")) | length,
            average_duration: (map(.duration) | add / length),
            success_rate: ((map(select(.status == "success")) | length) / length)
          }
        }) | 
        from_entries
      ),
      performance_metrics: {
        average_duration: (.pipeline_runs | map(.duration) | add / length),
        success_rate: ((.pipeline_runs | map(select(.status == "success")) | length) / (.pipeline_runs | length)),
        error_rate: ((.pipeline_runs | map(select(.status == "failure")) | length) / (.pipeline_runs | length)),
        last_updated: "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'"
      },
      health_scores: .health_scores
    }
    ' "$METRICS_FILE" > "$temp_file"
    
    mv "$temp_file" "$METRICS_FILE"
    
    log_metric "Performance metrics updated" "metrics"
}

# Calculate component health scores
calculate_health_scores() {
    if ! command -v jq >/dev/null 2>&1; then
        return
    fi
    
    log_info "Calculating component health scores" "health"
    
    local temp_file
    temp_file=$(mktemp)
    
    # Calculate health scores based on recent performance
    jq '
    .health_scores = {
      overall: (
        if .performance_metrics.success_rate > 0.95 then 100
        elif .performance_metrics.success_rate > 0.90 then 90
        elif .performance_metrics.success_rate > 0.80 then 75
        elif .performance_metrics.success_rate > 0.70 then 60
        elif .performance_metrics.success_rate > 0.50 then 40
        else 20
        end
      ),
      components: {
        "pipeline": (
          if .performance_metrics.success_rate > 0.95 then 100
          elif .performance_metrics.success_rate > 0.90 then 85
          else 60
          end
        ),
        "performance": (
          if .performance_metrics.average_duration < 1200 then 100
          elif .performance_metrics.average_duration < 1800 then 80
          elif .performance_metrics.average_duration < 2400 then 60
          else 40
          end
        ),
        "reliability": (
          if .performance_metrics.error_rate < 0.05 then 100
          elif .performance_metrics.error_rate < 0.10 then 85
          elif .performance_metrics.error_rate < 0.20 then 70
          else 50
          end
        )
      }
    }
    ' "$METRICS_FILE" > "$temp_file"
    
    mv "$temp_file" "$METRICS_FILE"
    
    log_metric "Health scores calculated" "health"
}

# ==============================================================================
# ALERT MANAGEMENT
# ==============================================================================

# Check alert conditions and trigger alerts if necessary
check_alert_conditions() {
    if ! command -v jq >/dev/null 2>&1; then
        log_warning "jq not available, alert checking disabled" "alerts"
        return
    fi
    
    log_info "Checking alert conditions" "alerts"
    
    # Get current metrics
    local success_rate
    local error_rate
    local avg_duration
    
    success_rate=$(jq -r '.performance_metrics.success_rate // 0' "$METRICS_FILE")
    error_rate=$(jq -r '.performance_metrics.error_rate // 0' "$METRICS_FILE")
    avg_duration=$(jq -r '.performance_metrics.average_duration // 0' "$METRICS_FILE")
    
    local alerts_triggered=()
    
    # Check success rate threshold
    if (( $(echo "$success_rate < $SLA_SUCCESS_RATE_TARGET" | bc -l) )); then
        alerts_triggered+=("LOW_SUCCESS_RATE")
        log_alert "Success rate below SLA target" "alerts" "{\"success_rate\":$success_rate,\"target\":$SLA_SUCCESS_RATE_TARGET}"
    fi
    
    # Check error rate threshold
    if (( $(echo "$error_rate > $ALERT_THRESHOLD_ERROR_RATE" | bc -l) )); then
        alerts_triggered+=("HIGH_ERROR_RATE")
        log_alert "Error rate above threshold" "alerts" "{\"error_rate\":$error_rate,\"threshold\":$ALERT_THRESHOLD_ERROR_RATE}"
    fi
    
    # Check average duration threshold
    if (( $(echo "$avg_duration > $ALERT_THRESHOLD_DURATION" | bc -l) )); then
        alerts_triggered+=("HIGH_DURATION")
        log_alert "Average duration above threshold" "alerts" "{\"duration\":$avg_duration,\"threshold\":$ALERT_THRESHOLD_DURATION}"
    fi
    
    # Trigger alerts if any conditions are met
    if [[ ${#alerts_triggered[@]} -gt 0 ]]; then
        trigger_alerts "${alerts_triggered[@]}"
    else
        log_info "All alert conditions are within normal ranges" "alerts"
    fi
}

# Trigger alerts for the given alert types
trigger_alerts() {
    local alert_types=("$@")
    
    log_alert "Triggering ${#alert_types[@]} alerts: ${alert_types[*]}" "alerts"
    
    for alert_type in "${alert_types[@]}"; do
        # Record alert
        record_alert "$alert_type"
        
        # Send notifications
        send_alert_notifications "$alert_type"
    done
}

# Record alert in alerts file
record_alert() {
    local alert_type="$1"
    local current_time
    current_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    if ! command -v jq >/dev/null 2>&1; then
        return
    fi
    
    local alert_record
    alert_record=$(cat << EOF
{
  "type": "$alert_type",
  "timestamp": "$current_time",
  "pipeline_id": "${GITHUB_RUN_ID:-unknown}",
  "workflow": "${GITHUB_WORKFLOW:-unknown}",
  "repository": "${GITHUB_REPOSITORY:-unknown}",
  "severity": "$(get_alert_severity "$alert_type")"
}
EOF
)
    
    local temp_file
    temp_file=$(mktemp)
    
    # Add to active alerts and alert history
    jq --argjson alert "$alert_record" '
    .active_alerts += [$alert] |
    .alert_history += [$alert]
    ' "$ALERTS_FILE" > "$temp_file"
    
    mv "$temp_file" "$ALERTS_FILE"
    
    log_metric "Alert recorded" "alerts" "$alert_record"
}

# Get alert severity based on alert type
get_alert_severity() {
    local alert_type="$1"
    
    case "$alert_type" in
        "LOW_SUCCESS_RATE"|"HIGH_ERROR_RATE")
            echo "HIGH"
            ;;
        "HIGH_DURATION")
            echo "MEDIUM"
            ;;
        *)
            echo "LOW"
            ;;
    esac
}

# Send alert notifications
send_alert_notifications() {
    local alert_type="$1"
    local severity
    severity=$(get_alert_severity "$alert_type")
    
    log_info "Sending alert notifications for $alert_type (severity: $severity)" "notifications"
    
    # Prepare alert message
    local alert_message
    alert_message=$(generate_alert_message "$alert_type" "$severity")
    
    # Send to Slack if configured
    if [[ -n "$SLACK_WEBHOOK_URL" ]]; then
        send_slack_notification "$alert_message" "$severity"
    fi
    
    # Send to Teams if configured
    if [[ -n "$TEAMS_WEBHOOK_URL" ]]; then
        send_teams_notification "$alert_message" "$severity"
    fi
    
    # Update GitHub step summary
    add_alert_to_github_summary "$alert_type" "$severity" "$alert_message"
}

# Generate alert message
generate_alert_message() {
    local alert_type="$1"
    local severity="$2"
    
    local base_message="Pipeline Alert: $alert_type"
    local details=""
    
    case "$alert_type" in
        "LOW_SUCCESS_RATE")
            details="Pipeline success rate has fallen below the SLA target of $(echo "$SLA_SUCCESS_RATE_TARGET * 100" | bc)%"
            ;;
        "HIGH_ERROR_RATE")
            details="Pipeline error rate has exceeded the threshold of $(echo "$ALERT_THRESHOLD_ERROR_RATE * 100" | bc)%"
            ;;
        "HIGH_DURATION")
            details="Pipeline average duration has exceeded the threshold of $((ALERT_THRESHOLD_DURATION / 60)) minutes"
            ;;
    esac
    
    cat << EOF
$base_message

Severity: $severity
Repository: ${GITHUB_REPOSITORY:-unknown}
Workflow: ${GITHUB_WORKFLOW:-unknown}
Pipeline ID: ${GITHUB_RUN_ID:-unknown}

Details: $details

Please investigate the pipeline health and take corrective action.

Time: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
EOF
}

# Send Slack notification
send_slack_notification() {
    local message="$1"
    local severity="$2"
    
    local color
    case "$severity" in
        "HIGH") color="#FF0000" ;;
        "MEDIUM") color="#FFA500" ;;
        *) color="#FFFF00" ;;
    esac
    
    local slack_payload
    slack_payload=$(cat << EOF
{
  "text": "Pipeline Alert",
  "attachments": [
    {
      "color": "$color",
      "title": "CI/CD Pipeline Alert - $severity Severity",
      "text": "$message",
      "footer": "Pipeline Monitoring",
      "ts": $(date +%s)
    }
  ]
}
EOF
)
    
    if curl -X POST -H 'Content-type: application/json' \
        --data "$slack_payload" \
        --connect-timeout 30 \
        --max-time 60 \
        "$SLACK_WEBHOOK_URL" >/dev/null 2>&1; then
        log_info "Slack notification sent successfully" "notifications"
    else
        log_warning "Failed to send Slack notification" "notifications"
    fi
}

# Send Teams notification
send_teams_notification() {
    local message="$1"
    local severity="$2"
    
    local theme_color
    case "$severity" in
        "HIGH") theme_color="FF0000" ;;
        "MEDIUM") theme_color="FFA500" ;;
        *) theme_color="FFFF00" ;;
    esac
    
    local teams_payload
    teams_payload=$(cat << EOF
{
  "@type": "MessageCard",
  "@context": "http://schema.org/extensions",
  "themeColor": "$theme_color",
  "summary": "Pipeline Alert",
  "sections": [{
    "activityTitle": "CI/CD Pipeline Alert",
    "activitySubtitle": "$severity Severity",
    "activityImage": "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
    "text": "$message",
    "markdown": true
  }]
}
EOF
)
    
    if curl -X POST -H 'Content-type: application/json' \
        --data "$teams_payload" \
        --connect-timeout 30 \
        --max-time 60 \
        "$TEAMS_WEBHOOK_URL" >/dev/null 2>&1; then
        log_info "Teams notification sent successfully" "notifications"
    else
        log_warning "Failed to send Teams notification" "notifications"
    fi
}

# Add alert to GitHub step summary
add_alert_to_github_summary() {
    local alert_type="$1"
    local severity="$2"
    local message="$3"
    
    if [[ -z "${GITHUB_STEP_SUMMARY:-}" ]]; then
        return
    fi
    
    local severity_icon
    case "$severity" in
        "HIGH") severity_icon="🚨" ;;
        "MEDIUM") severity_icon="⚠️" ;;
        *) severity_icon="⚠️" ;;
    esac
    
    cat >> "$GITHUB_STEP_SUMMARY" << EOF

## $severity_icon Pipeline Alert - $severity Severity

**Alert Type**: $alert_type  
**Timestamp**: $(date -u +"%Y-%m-%d %H:%M:%S UTC")  

### Details
$message

### Recommended Actions
1. Review recent pipeline runs for patterns
2. Check system resource availability
3. Investigate external dependency health
4. Consider scaling or optimization measures

EOF
    
    log_info "Alert added to GitHub step summary" "notifications"
}

# ==============================================================================
# SLA TRACKING
# ==============================================================================

# Update SLA tracking metrics
update_sla_tracking() {
    local pipeline_status="$1"
    local pipeline_duration="$2"
    
    if ! command -v jq >/dev/null 2>&1; then
        log_warning "jq not available, SLA tracking disabled" "sla"
        return
    fi
    
    log_info "Updating SLA tracking metrics" "sla"
    
    local temp_file
    temp_file=$(mktemp)
    
    # Update current period metrics
    jq --arg status "$pipeline_status" --argjson duration "$pipeline_duration" '
    .current_period.total_runs += 1 |
    .current_period.total_duration += $duration |
    if $status == "success" then
      .current_period.successful_runs += 1
    else
      .
    end |
    .current_period.success_rate = (.current_period.successful_runs / .current_period.total_runs) |
    .current_period.average_duration = (.current_period.total_duration / .current_period.total_runs)
    ' "$SLA_TRACKING_FILE" > "$temp_file"
    
    mv "$temp_file" "$SLA_TRACKING_FILE"
    
    # Check SLA compliance
    check_sla_compliance
    
    log_metric "SLA tracking updated" "sla"
}

# Check SLA compliance
check_sla_compliance() {
    if ! command -v jq >/dev/null 2>&1; then
        return
    fi
    
    local current_success_rate
    local current_avg_duration
    local success_sla_met
    local duration_sla_met
    
    current_success_rate=$(jq -r '.current_period.success_rate' "$SLA_TRACKING_FILE")
    current_avg_duration=$(jq -r '.current_period.average_duration' "$SLA_TRACKING_FILE")
    
    success_sla_met=$(echo "$current_success_rate >= $SLA_SUCCESS_RATE_TARGET" | bc -l)
    duration_sla_met=$(echo "$current_avg_duration <= $SLA_DURATION_TARGET" | bc -l)
    
    if [[ "$success_sla_met" -eq 1 ]] && [[ "$duration_sla_met" -eq 1 ]]; then
        log_info "SLA compliance: MEETING targets" "sla" "{\"success_rate\":$current_success_rate,\"duration\":$current_avg_duration}"
    else
        log_warning "SLA compliance: NOT MEETING targets" "sla" "{\"success_rate\":$current_success_rate,\"duration\":$current_avg_duration,\"success_target\":$SLA_SUCCESS_RATE_TARGET,\"duration_target\":$SLA_DURATION_TARGET}"
        
        # Trigger SLA breach alert
        if [[ "$success_sla_met" -eq 0 ]]; then
            trigger_alerts "SLA_SUCCESS_BREACH"
        fi
        if [[ "$duration_sla_met" -eq 0 ]]; then
            trigger_alerts "SLA_DURATION_BREACH"
        fi
    fi
}

# Generate SLA report
generate_sla_report() {
    if ! command -v jq >/dev/null 2>&1; then
        log_warning "jq not available, SLA report generation disabled" "sla"
        return
    fi
    
    log_info "Generating SLA compliance report" "sla"
    
    local report_file="${MONITORING_STATE_DIR}/sla-report.json"
    local current_time
    current_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Generate comprehensive SLA report
    jq --arg timestamp "$current_time" '
    {
      "report_timestamp": $timestamp,
      "sla_targets": .sla_targets,
      "current_period": .current_period,
      "compliance_status": {
        "success_rate_compliant": (.current_period.success_rate >= .sla_targets.success_rate),
        "duration_compliant": (.current_period.average_duration <= .sla_targets.duration_target),
        "overall_compliant": (
          (.current_period.success_rate >= .sla_targets.success_rate) and
          (.current_period.average_duration <= .sla_targets.duration_target)
        )
      },
      "performance_summary": {
        "success_rate_percentage": (.current_period.success_rate * 100),
        "average_duration_minutes": (.current_period.average_duration / 60),
        "total_runs": .current_period.total_runs,
        "successful_runs": .current_period.successful_runs,
        "failed_runs": (.current_period.total_runs - .current_period.successful_runs)
      }
    }
    ' "$SLA_TRACKING_FILE" > "$report_file"
    
    log_metric "SLA report generated" "sla" "$(cat "$report_file")"
    
    # Add to GitHub step summary if available
    if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
        add_sla_report_to_github_summary "$report_file"
    fi
}

# Add SLA report to GitHub step summary
add_sla_report_to_github_summary() {
    local report_file="$1"
    
    if ! command -v jq >/dev/null 2>&1; then
        return
    fi
    
    local overall_compliant
    local success_rate_pct
    local avg_duration_min
    local total_runs
    local success_runs
    local failed_runs
    
    overall_compliant=$(jq -r '.compliance_status.overall_compliant' "$report_file")
    success_rate_pct=$(jq -r '.performance_summary.success_rate_percentage' "$report_file")
    avg_duration_min=$(jq -r '.performance_summary.average_duration_minutes' "$report_file")
    total_runs=$(jq -r '.performance_summary.total_runs' "$report_file")
    success_runs=$(jq -r '.performance_summary.successful_runs' "$report_file")
    failed_runs=$(jq -r '.performance_summary.failed_runs' "$report_file")
    
    local compliance_icon
    if [[ "$overall_compliant" == "true" ]]; then
        compliance_icon="✅"
    else
        compliance_icon="❌"
    fi
    
    cat >> "$GITHUB_STEP_SUMMARY" << EOF

## $compliance_icon SLA Compliance Report

**Overall Compliance**: $(if [[ "$overall_compliant" == "true" ]]; then echo "MEETING SLA"; else echo "NOT MEETING SLA"; fi)  
**Report Generated**: $(date -u +"%Y-%m-%d %H:%M:%S UTC")  

### Performance Metrics

| Metric | Current | Target | Status |
|--------|---------|---------|--------|
| Success Rate | $(printf "%.1f%%" "$success_rate_pct") | $(echo "$SLA_SUCCESS_RATE_TARGET * 100" | bc)% | $(if (( $(echo "$success_rate_pct >= $(echo "$SLA_SUCCESS_RATE_TARGET * 100" | bc)" | bc -l) )); then echo "✅"; else echo "❌"; fi) |
| Avg Duration | $(printf "%.1f min" "$avg_duration_min") | $((SLA_DURATION_TARGET / 60)) min | $(if (( $(echo "$avg_duration_min <= $((SLA_DURATION_TARGET / 60))" | bc -l) )); then echo "✅"; else echo "❌"; fi) |

### Run Statistics

- **Total Runs**: $total_runs
- **Successful**: $success_runs
- **Failed**: $failed_runs

EOF
}

# ==============================================================================
# REPORTING AND DASHBOARDS
# ==============================================================================

# Generate comprehensive monitoring dashboard
generate_monitoring_dashboard() {
    log_info "Generating comprehensive monitoring dashboard" "dashboard"
    
    local dashboard_file="${MONITORING_STATE_DIR}/monitoring-dashboard.html"
    local current_time
    current_time=$(date -u +"%Y-%m-%d %H:%M:%S UTC")
    
    cat > "$dashboard_file" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CI/CD Pipeline Monitoring Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2196F3; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .card { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric { display: inline-block; margin: 10px 20px 10px 0; }
        .metric-value { font-size: 2em; font-weight: bold; color: #2196F3; }
        .metric-label { font-size: 0.9em; color: #666; }
        .alert { padding: 15px; margin: 10px 0; border-radius: 4px; }
        .alert-high { background: #ffebee; border-left: 4px solid #f44336; }
        .alert-medium { background: #fff3e0; border-left: 4px solid #ff9800; }
        .alert-low { background: #f3e5f5; border-left: 4px solid #9c27b0; }
        .status-good { color: #4caf50; }
        .status-warning { color: #ff9800; }
        .status-error { color: #f44336; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f5f5f5; }
        .progress-bar { width: 100%; height: 20px; background: #eee; border-radius: 10px; overflow: hidden; }
        .progress-fill { height: 100%; background: #4caf50; transition: width 0.3s ease; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 CI/CD Pipeline Monitoring Dashboard</h1>
            <p>Real-time monitoring and alerting for pipeline health and performance</p>
            <p><strong>Last Updated:</strong> TIMESTAMP_PLACEHOLDER</p>
        </div>
        
        <div class="card">
            <h2>📊 Key Performance Indicators</h2>
            <div id="kpi-metrics">
                <!-- KPI metrics will be injected here -->
            </div>
        </div>
        
        <div class="card">
            <h2>🏥 Health Scores</h2>
            <div id="health-scores">
                <!-- Health scores will be injected here -->
            </div>
        </div>
        
        <div class="card">
            <h2>🚨 Active Alerts</h2>
            <div id="active-alerts">
                <!-- Active alerts will be injected here -->
            </div>
        </div>
        
        <div class="card">
            <h2>📈 SLA Compliance</h2>
            <div id="sla-compliance">
                <!-- SLA compliance will be injected here -->
            </div>
        </div>
        
        <div class="card">
            <h2>📋 Recent Pipeline Runs</h2>
            <div id="recent-runs">
                <!-- Recent runs will be injected here -->
            </div>
        </div>
    </div>
</body>
</html>
EOF
    
    # Replace timestamp placeholder
    sed -i "s/TIMESTAMP_PLACEHOLDER/$current_time/g" "$dashboard_file"
    
    # Inject data if jq is available
    if command -v jq >/dev/null 2>&1; then
        inject_dashboard_data "$dashboard_file"
    fi
    
    log_info "Monitoring dashboard generated: $dashboard_file" "dashboard"
}

# Inject data into dashboard HTML
inject_dashboard_data() {
    local dashboard_file="$1"
    
    # This is a simplified version - in a real implementation,
    # you would inject actual data from the JSON files
    log_info "Dashboard data injection would occur here" "dashboard"
}

# ==============================================================================
# MAIN FUNCTIONS AND CLI INTERFACE
# ==============================================================================

# Show usage information
show_usage() {
    cat << EOF
Pipeline Health Monitoring and Alerting Script

Usage: $0 <command> [options]

Commands:
  init                           Initialize monitoring environment
  record-run <status> <duration> [job_results]  Record pipeline run metrics
  check-alerts                   Check alert conditions and trigger if needed
  update-sla <status> <duration> Update SLA tracking metrics
  generate-report                Generate comprehensive monitoring report
  generate-dashboard             Generate HTML monitoring dashboard
  send-test-alert <type>         Send test alert notification

Monitoring Commands:
  health-check                   Perform comprehensive health assessment
  metrics-summary                Display current metrics summary
  sla-status                     Show current SLA compliance status
  alert-history                  Show recent alert history

Reporting Commands:
  dashboard                      Generate monitoring dashboard
  export-metrics                 Export metrics to external format
  cleanup                        Clean up old monitoring data

Options:
  --alert-threshold-error <rate>     Set error rate alert threshold (default: $ALERT_THRESHOLD_ERROR_RATE)
  --alert-threshold-duration <sec>   Set duration alert threshold (default: $ALERT_THRESHOLD_DURATION)
  --sla-success-rate <rate>          Set SLA success rate target (default: $SLA_SUCCESS_RATE_TARGET)
  --sla-duration <sec>               Set SLA duration target (default: $SLA_DURATION_TARGET)
  --help                             Show this help message

Examples:
  $0 init
  $0 record-run success 1200 '{"quick-checks":"success","test":"success"}'
  $0 check-alerts
  $0 update-sla success 1200
  $0 generate-dashboard
  $0 send-test-alert HIGH_ERROR_RATE

Environment Variables:
  SLACK_WEBHOOK_URL              Slack webhook URL for notifications
  TEAMS_WEBHOOK_URL              Microsoft Teams webhook URL
  EMAIL_NOTIFICATION_ENABLED     Enable email notifications (true/false)
EOF
}

# Main function
main() {
    local command="${1:-}"
    
    if [[ $# -eq 0 || "$command" == "--help" || "$command" == "-h" ]]; then
        show_usage
        exit 0
    fi
    
    # Initialize monitoring environment
    init_monitoring_environment
    
    case "$command" in
        "init")
            log_info "Monitoring environment initialized" "monitoring"
            ;;
        "record-run")
            if [[ $# -lt 3 ]]; then
                log_alert "Usage: $0 record-run <status> <duration> [job_results]" "monitoring"
                exit 1
            fi
            local status="$2"
            local duration="$3"
            local job_results="${4:-{}}"
            record_pipeline_run "${GITHUB_RUN_ID:-unknown}" "$status" "$duration" "$job_results"
            ;;
        "check-alerts")
            check_alert_conditions
            ;;
        "update-sla")
            if [[ $# -lt 3 ]]; then
                log_alert "Usage: $0 update-sla <status> <duration>" "monitoring"
                exit 1
            fi
            update_sla_tracking "$2" "$3"
            ;;
        "generate-report")
            update_performance_metrics
            calculate_health_scores
            generate_sla_report
            ;;
        "generate-dashboard")
            generate_monitoring_dashboard
            ;;
        "send-test-alert")
            if [[ $# -lt 2 ]]; then
                log_alert "Usage: $0 send-test-alert <alert_type>" "monitoring"
                exit 1
            fi
            trigger_alerts "$2"
            ;;
        "health-check")
            update_performance_metrics
            calculate_health_scores
            check_alert_conditions
            log_info "Health check completed" "monitoring"
            ;;
        "metrics-summary")
            if command -v jq >/dev/null 2>&1 && [[ -f "$METRICS_FILE" ]]; then
                echo "Pipeline Metrics Summary:"
                jq -r '
                "Success Rate: " + (.performance_metrics.success_rate * 100 | tostring) + "%\n" +
                "Error Rate: " + (.performance_metrics.error_rate * 100 | tostring) + "%\n" +
                "Average Duration: " + ((.performance_metrics.average_duration / 60) | tostring) + " minutes\n" +
                "Overall Health Score: " + (.health_scores.overall | tostring)
                ' "$METRICS_FILE"
            else
                log_warning "Metrics file not available or jq not installed" "monitoring"
            fi
            ;;
        "sla-status")
            if command -v jq >/dev/null 2>&1 && [[ -f "$SLA_TRACKING_FILE" ]]; then
                echo "SLA Compliance Status:"
                jq -r '
                "Success Rate: " + (.current_period.success_rate * 100 | tostring) + "% (Target: " + (.sla_targets.success_rate * 100 | tostring) + "%)\n" +
                "Average Duration: " + ((.current_period.average_duration / 60) | tostring) + " min (Target: " + ((.sla_targets.duration_target / 60) | tostring) + " min)\n" +
                "Total Runs: " + (.current_period.total_runs | tostring) + "\n" +
                "Successful Runs: " + (.current_period.successful_runs | tostring)
                ' "$SLA_TRACKING_FILE"
            else
                log_warning "SLA tracking file not available or jq not installed" "monitoring"
            fi
            ;;
        "alert-history")
            if command -v jq >/dev/null 2>&1 && [[ -f "$ALERTS_FILE" ]]; then
                echo "Recent Alert History:"
                jq -r '.alert_history[-10:] | .[] | .timestamp + " - " + .type + " (" + .severity + ")"' "$ALERTS_FILE"
            else
                log_warning "Alerts file not available or jq not installed" "monitoring"
            fi
            ;;
        "dashboard")
            generate_monitoring_dashboard
            ;;
        "cleanup")
            log_info "Cleaning up old monitoring data" "monitoring"
            find "$MONITORING_STATE_DIR" -name "*.log" -mtime +7 -delete 2>/dev/null || true
            find "$MONITORING_STATE_DIR" -name "*.html" -mtime +1 -delete 2>/dev/null || true
            log_info "Monitoring data cleanup completed" "monitoring"
            ;;
        *)
            log_alert "Unknown command: $command" "monitoring"
            show_usage
            exit 1
            ;;
    esac
}

# Execute main function if script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi