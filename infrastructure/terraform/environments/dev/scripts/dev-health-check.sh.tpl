#!/bin/bash

# Development Environment Health Check Script
# Generated automatically by Terraform for the freightliner project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
HEALTH_CHECK_URL="${health_check_url}"
PIPELINE_DASHBOARD_URL="${pipeline_dashboard_url}"
PERFORMANCE_DASHBOARD_URL="${performance_dashboard_url}"
COST_DASHBOARD_URL="${cost_dashboard_url}"
GRAFANA_URL="${grafana_url}"

# Logging function
log() {
    echo -e "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

log_success() {
    log "${GREEN}✓ $1${NC}"
}

log_warning() {
    log "${YELLOW}⚠ $1${NC}"
}

log_error() {
    log "${RED}✗ $1${NC}"
}

log_info() {
    log "${BLUE}ℹ $1${NC}"
}

# Header
echo -e "${BLUE}"
echo "=============================================="
echo "  Freightliner CI/CD Monitoring Health Check"
echo "  Environment: Development"
echo "  Generated: $(date)"
echo "=============================================="
echo -e "${NC}"

# Function to check HTTP endpoint
check_endpoint() {
    local url="$1"
    local name="$2"
    local timeout="${3:-10}"
    
    log_info "Checking $name..."
    
    if curl -s --fail --max-time $timeout "$url" > /dev/null 2>&1; then
        log_success "$name is responding"
        return 0
    else
        log_error "$name is not responding or returned an error"
        return 1
    fi
}

# Function to check detailed health
check_detailed_health() {
    log_info "Performing detailed health check..."
    
    local response=$(curl -s --fail --max-time 30 "$HEALTH_CHECK_URL" 2>/dev/null)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        echo "$response" | jq . > /dev/null 2>&1
        if [ $? -eq 0 ]; then
            local overall_status=$(echo "$response" | jq -r '.overall')
            local healthy_components=$(echo "$response" | jq -r '.summary.healthy_components // 0')
            local total_components=$(echo "$response" | jq -r '.summary.total_components // 0')
            
            case "$overall_status" in
                "healthy")
                    log_success "System is healthy ($healthy_components/$total_components components)"
                    ;;
                "degraded"|"warning")
                    log_warning "System is $overall_status ($healthy_components/$total_components components)"
                    ;;
                "unhealthy"|"error")
                    log_error "System is $overall_status ($healthy_components/$total_components components)"
                    ;;
                *)
                    log_warning "System status is $overall_status"
                    ;;
            esac
            
            # Show component details
            echo "$response" | jq -r '.components | to_entries[] | select(.value.status != "healthy") | "\(.key): \(.value.status)"' | while read line; do
                if [ -n "$line" ]; then
                    log_warning "Component issue: $line"
                fi
            done
            
        else
            log_warning "Health check returned invalid JSON"
        fi
    else
        log_error "Health check endpoint is not accessible"
    fi
}

# Function to check AWS CLI
check_aws_cli() {
    log_info "Checking AWS CLI access..."
    
    if command -v aws > /dev/null 2>&1; then
        if aws sts get-caller-identity > /dev/null 2>&1; then
            local account_id=$(aws sts get-caller-identity --query Account --output text 2>/dev/null)
            local region=$(aws configure get region 2>/dev/null || echo "default")
            log_success "AWS CLI is configured (Account: $account_id, Region: $region)"
        else
            log_warning "AWS CLI is installed but not configured or lacks permissions"
        fi
    else
        log_warning "AWS CLI is not installed"
    fi
}

# Function to check recent deployments
check_recent_deployments() {
    log_info "Checking recent GitHub Actions runs..."
    
    if command -v gh > /dev/null 2>&1; then
        local recent_runs=$(gh api repos/{owner}/{repo}/actions/runs --jq '.workflow_runs[:5] | .[] | "\(.conclusion) \(.created_at) \(.head_branch)"' 2>/dev/null)
        
        if [ $? -eq 0 ] && [ -n "$recent_runs" ]; then
            echo "$recent_runs" | while IFS= read -r line; do
                local status=$(echo "$line" | cut -d' ' -f1)
                local date=$(echo "$line" | cut -d' ' -f2)
                local branch=$(echo "$line" | cut -d' ' -f3)
                
                case "$status" in
                    "success")
                        log_success "Recent run: $status on $branch ($date)"
                        ;;
                    "failure")
                        log_error "Recent run: $status on $branch ($date)"
                        ;;
                    *)
                        log_info "Recent run: $status on $branch ($date)"
                        ;;
                esac
            done
        else
            log_warning "Could not fetch recent GitHub Actions runs"
        fi
    else
        log_info "GitHub CLI not installed, skipping recent deployments check"
    fi
}

# Function to show resource usage
show_resource_usage() {
    log_info "Checking AWS resource usage..."
    
    if command -v aws > /dev/null 2>&1 && aws sts get-caller-identity > /dev/null 2>&1; then
        # Check Lambda functions
        local lambda_count=$(aws lambda list-functions --query 'Functions[?contains(FunctionName, `freightliner`) || contains(FunctionName, `ci-monitoring`)] | length(@)' --output text 2>/dev/null || echo "0")
        log_info "Lambda functions: $lambda_count"
        
        # Check S3 buckets
        local s3_count=$(aws s3api list-buckets --query 'Buckets[?contains(Name, `freightliner`) || contains(Name, `ci-monitoring`)] | length(@)' --output text 2>/dev/null || echo "0")
        log_info "S3 buckets: $s3_count"
        
        # Check DynamoDB tables
        local dynamo_count=$(aws dynamodb list-tables --query 'TableNames[?contains(@, `freightliner`) || contains(@, `ci-monitoring`)] | length(@)' --output text 2>/dev/null || echo "0")
        log_info "DynamoDB tables: $dynamo_count"
        
        # Check CloudWatch alarms
        local alarm_count=$(aws cloudwatch describe-alarms --query 'MetricAlarms[?contains(AlarmName, `freightliner`) || contains(AlarmName, `ci-monitoring`)] | length(@)' --output text 2>/dev/null || echo "0")
        log_info "CloudWatch alarms: $alarm_count"
    fi
}

# Main health check sequence
main() {
    local overall_status="healthy"
    
    # Basic connectivity checks
    if ! check_endpoint "$HEALTH_CHECK_URL" "Health Check API"; then
        overall_status="unhealthy"
    fi
    
    # Detailed health check
    check_detailed_health
    
    # Additional checks
    check_aws_cli
    check_recent_deployments
    show_resource_usage
    
    echo ""
    log_info "Dashboard URLs:"
    echo "  • Pipeline Overview: $PIPELINE_DASHBOARD_URL"
    echo "  • Performance Analysis: $PERFORMANCE_DASHBOARD_URL"
    echo "  • Cost Analysis: $COST_DASHBOARD_URL"
    %{ if grafana_url != null && grafana_url != "" }
    echo "  • Grafana Workspace: $GRAFANA_URL"
    %{ endif }
    
    echo ""
    log_info "Quick Actions:"
    echo "  • View logs: aws logs describe-log-groups --log-group-name-prefix '/aws/lambda/freightliner'"
    echo "  • Check metrics: aws cloudwatch list-metrics --namespace 'CI-CD/Pipeline'"
    echo "  • Test health endpoint: curl -s '$HEALTH_CHECK_URL' | jq ."
    
    echo ""
    case "$overall_status" in
        "healthy")
            log_success "Overall system status: HEALTHY"
            exit 0
            ;;
        "degraded")
            log_warning "Overall system status: DEGRADED"
            exit 1
            ;;
        "unhealthy")
            log_error "Overall system status: UNHEALTHY"
            exit 2
            ;;
        *)
            log_warning "Overall system status: UNKNOWN"
            exit 3
            ;;
    esac
}

# Function to show help
show_help() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -v, --verbose  Enable verbose output"
    echo "  -q, --quiet    Suppress non-error output"
    echo "  --json         Output results in JSON format"
    echo ""
    echo "Examples:"
    echo "  $0                 # Run standard health check"
    echo "  $0 --verbose       # Run with detailed output"
    echo "  $0 --json          # Output JSON results"
    echo ""
}

# Parse command line arguments
VERBOSE=false
QUIET=false
JSON_OUTPUT=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        --json)
            JSON_OUTPUT=true
            QUIET=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Check for required tools
if ! command -v curl > /dev/null 2>&1; then
    log_error "curl is required but not installed"
    exit 1
fi

if [ "$JSON_OUTPUT" = true ] && ! command -v jq > /dev/null 2>&1; then
    log_error "jq is required for JSON output but not installed"
    exit 1
fi

# Adjust output based on options
if [ "$QUIET" = true ]; then
    exec > /dev/null 2>&1
fi

# Run main function
main