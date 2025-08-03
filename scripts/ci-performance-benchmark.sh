#!/bin/bash
# =============================================================================
# CI/CD Pipeline Performance Benchmarking Script
# =============================================================================
# This script measures and compares CI pipeline performance metrics
# Usage: ./scripts/ci-performance-benchmark.sh [baseline|compare|report]

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BENCHMARK_DIR="$PROJECT_ROOT/.ci-benchmarks"
RESULTS_FILE="$BENCHMARK_DIR/results.json"
BASELINE_FILE="$BENCHMARK_DIR/baseline.json"
REPORT_FILE="$BENCHMARK_DIR/performance-report.md"

# Performance thresholds (in seconds)
SETUP_THRESHOLD=120      # 2 minutes
BUILD_THRESHOLD=300      # 5 minutes  
TEST_THRESHOLD=900       # 15 minutes
LINT_THRESHOLD=180       # 3 minutes
DOCKER_THRESHOLD=900     # 15 minutes
TOTAL_THRESHOLD=1800     # 30 minutes

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Initialize benchmark directory
init_benchmark_dir() {
    mkdir -p "$BENCHMARK_DIR"
    chmod 755 "$BENCHMARK_DIR"
}

# Get current timestamp
get_timestamp() {
    date -u +"%Y-%m-%dT%H:%M:%SZ"
}

# Measure command execution time
measure_time() {
    local command="$1"
    local description="$2"
    
    log_info "Measuring: $description"
    
    local start_time=$(date +%s.%N)
    
    # Execute command and capture exit code
    local exit_code=0
    if ! eval "$command" >/dev/null 2>&1; then
        exit_code=$?
    fi
    
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc -l)
    
    # Format duration to 2 decimal places
    duration=$(printf "%.2f" "$duration")
    
    if [[ $exit_code -eq 0 ]]; then
        log_success "$description completed in ${duration}s"
    else
        log_error "$description failed in ${duration}s (exit code: $exit_code)"
    fi
    
    echo "$duration"
}

# Benchmark individual pipeline stages
benchmark_pipeline() {
    local mode="${1:-full}"
    
    log_info "Starting CI pipeline benchmark (mode: $mode)"
    
    # Ensure we're in the project root
    cd "$PROJECT_ROOT"
    
    # Results object
    local results="{}"
    local total_start=$(date +%s.%N)
    
    # 1. Setup and dependency resolution
    log_info "📦 Benchmarking dependency setup..."
    local setup_time
    setup_time=$(measure_time "go mod download && go mod verify" "Dependency setup")
    results=$(echo "$results" | jq ". + {\"setup_time\": $setup_time}")
    
    # 2. Build performance
    log_info "🔨 Benchmarking build performance..."
    local build_time
    build_time=$(measure_time "go build -v ./..." "Build")
    results=$(echo "$results" | jq ". + {\"build_time\": $build_time}")
    
    # 3. Test execution (if not quick mode)
    if [[ "$mode" != "quick" ]]; then
        log_info "🧪 Benchmarking test execution..."
        local test_time
        test_time=$(measure_time "go test -short -timeout=8m ./..." "Unit tests")
        results=$(echo "$results" | jq ". + {\"test_time\": $test_time}")
        
        # Integration tests
        local integration_time
        integration_time=$(measure_time "go test -run Integration -timeout=10m ./..." "Integration tests")
        results=$(echo "$results" | jq ". + {\"integration_time\": $integration_time}")
    else
        results=$(echo "$results" | jq ". + {\"test_time\": 0, \"integration_time\": 0}")
    fi
    
    # 4. Linting performance
    log_info "🔍 Benchmarking linting performance..."
    local lint_time
    if command -v golangci-lint >/dev/null 2>&1; then
        lint_time=$(measure_time "golangci-lint run --timeout=6m" "Linting")
    else
        lint_time=0
        log_warning "golangci-lint not available, skipping lint benchmark"
    fi
    results=$(echo "$results" | jq ". + {\"lint_time\": $lint_time}")
    
    # 5. Docker build (if dockerfile exists and not quick mode)
    if [[ "$mode" != "quick" && (-f "Dockerfile.optimized" || -f "Dockerfile") ]]; then
        log_info "🐳 Benchmarking Docker build..."
        local docker_time
        local dockerfile="Dockerfile.optimized"
        [[ -f "$dockerfile" ]] || dockerfile="Dockerfile"
        
        docker_time=$(measure_time "docker build -f $dockerfile -t freightliner:benchmark . >/dev/null" "Docker build")
        results=$(echo "$results" | jq ". + {\"docker_time\": $docker_time}")
        
        # Cleanup
        docker rmi freightliner:benchmark >/dev/null 2>&1 || true
    else
        results=$(echo "$results" | jq ". + {\"docker_time\": 0}")
    fi
    
    # Calculate total time
    local total_end=$(date +%s.%N)
    local total_time=$(echo "$total_end - $total_start" | bc -l)
    total_time=$(printf "%.2f" "$total_time")
    
    # Add metadata
    results=$(echo "$results" | jq ". + {
        \"total_time\": $total_time,
        \"timestamp\": \"$(get_timestamp)\",
        \"mode\": \"$mode\",
        \"git_commit\": \"$(git rev-parse HEAD 2>/dev/null || echo 'unknown')\",
        \"git_branch\": \"$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')\",
        \"go_version\": \"$(go version | awk '{print $3}' | sed 's/go//')\",
        \"platform\": \"$(uname -s)-$(uname -m)\"
    }")
    
    log_success "Pipeline benchmark completed in ${total_time}s"
    
    echo "$results"
}

# Save baseline performance
save_baseline() {
    log_info "🎯 Creating performance baseline..."
    
    local results
    results=$(benchmark_pipeline "full")
    
    echo "$results" | jq '.' > "$BASELINE_FILE"
    
    log_success "Baseline saved to $BASELINE_FILE"
    
    # Display baseline summary
    echo
    echo "📊 Baseline Performance Summary:"
    echo "================================"
    printf "Setup Time:       %6.2fs\n" "$(echo "$results" | jq -r '.setup_time')"
    printf "Build Time:       %6.2fs\n" "$(echo "$results" | jq -r '.build_time')"
    printf "Test Time:        %6.2fs\n" "$(echo "$results" | jq -r '.test_time')"
    printf "Integration Time: %6.2fs\n" "$(echo "$results" | jq -r '.integration_time')"
    printf "Lint Time:        %6.2fs\n" "$(echo "$results" | jq -r '.lint_time')"
    printf "Docker Time:      %6.2fs\n" "$(echo "$results" | jq -r '.docker_time')"
    printf "Total Time:       %6.2fs\n" "$(echo "$results" | jq -r '.total_time')"
}

# Compare current performance against baseline
compare_performance() {
    local mode="${1:-full}"
    
    if [[ ! -f "$BASELINE_FILE" ]]; then
        log_error "No baseline found. Run './scripts/ci-performance-benchmark.sh baseline' first."
        exit 1
    fi
    
    log_info "📈 Comparing performance against baseline..."
    
    local current_results
    current_results=$(benchmark_pipeline "$mode")
    
    local baseline_results
    baseline_results=$(cat "$BASELINE_FILE")
    
    # Save current results
    echo "$current_results" | jq '.' > "$RESULTS_FILE"
    
    # Compare results
    echo
    echo "📊 Performance Comparison:"
    echo "=========================="
    
    # Helper function to compare times
    compare_time() {
        local metric="$1"
        local current
        local baseline
        local diff
        local percentage
        
        current=$(echo "$current_results" | jq -r ".$metric")
        baseline=$(echo "$baseline_results" | jq -r ".$metric")
        
        if [[ "$baseline" == "0" ]] || [[ "$current" == "0" ]]; then
            printf "%-20s %8.2fs -> %8.2fs (skipped)\n" "$metric:" "$baseline" "$current"
            return
        fi
        
        diff=$(echo "$current - $baseline" | bc -l)
        percentage=$(echo "scale=1; ($diff / $baseline) * 100" | bc -l)
        
        local status="📊"
        if (( $(echo "$percentage > 20" | bc -l) )); then
            status="🔴"  # Significant regression
        elif (( $(echo "$percentage > 5" | bc -l) )); then
            status="🟡"  # Minor regression
        elif (( $(echo "$percentage < -5" | bc -l) )); then
            status="🟢"  # Improvement
        fi
        
        printf "%-20s %8.2fs -> %8.2fs (%+6.1f%%) %s\n" "$metric:" "$baseline" "$current" "$percentage" "$status"
    }
    
    compare_time "setup_time"
    compare_time "build_time"
    compare_time "test_time"
    compare_time "integration_time"
    compare_time "lint_time"
    compare_time "docker_time"
    compare_time "total_time"
    
    # Check for significant regressions
    local total_current total_baseline total_percentage
    total_current=$(echo "$current_results" | jq -r '.total_time')
    total_baseline=$(echo "$baseline_results" | jq -r '.total_time')
    total_percentage=$(echo "scale=1; (($total_current - $total_baseline) / $total_baseline) * 100" | bc -l)
    
    echo
    if (( $(echo "$total_percentage > 20" | bc -l) )); then
        log_error "Significant performance regression detected: +${total_percentage}%"
        echo "Consider investigating the cause of this slowdown."
        exit 1
    elif (( $(echo "$total_percentage > 5" | bc -l) )); then
        log_warning "Minor performance regression: +${total_percentage}%"
        echo "Monitor this trend to prevent further degradation."
    elif (( $(echo "$total_percentage < -5" | bc -l) )); then
        log_success "Performance improvement detected: ${total_percentage}%"
        echo "Great work on optimizing the pipeline!"
    else
        log_info "Performance is stable: ${total_percentage}%"
    fi
}

# Generate detailed performance report
generate_report() {
    if [[ ! -f "$RESULTS_FILE" ]] && [[ ! -f "$BASELINE_FILE" ]]; then
        log_error "No performance data found. Run benchmarks first."
        exit 1
    fi
    
    log_info "📋 Generating performance report..."
    
    local current_results="{}"
    local baseline_results="{}"
    
    [[ -f "$RESULTS_FILE" ]] && current_results=$(cat "$RESULTS_FILE")
    [[ -f "$BASELINE_FILE" ]] && baseline_results=$(cat "$BASELINE_FILE")
    
    # Generate markdown report
    cat > "$REPORT_FILE" << 'EOF'
# CI/CD Pipeline Performance Report

**Generated:** `$(get_timestamp)`  
**Project:** Freightliner  
**Purpose:** CI/CD pipeline performance analysis and optimization tracking

## Executive Summary

This report provides a comprehensive analysis of the CI/CD pipeline performance, comparing current execution times against established baselines and industry benchmarks.

EOF
    
    # Add current results if available
    if [[ "$current_results" != "{}" ]]; then
        cat >> "$REPORT_FILE" << EOF
### Current Performance Metrics

| Stage | Time (seconds) | Status |
|-------|---------------|--------|
| Setup & Dependencies | $(echo "$current_results" | jq -r '.setup_time') | $(if (( $(echo "$(echo "$current_results" | jq -r '.setup_time') <= $SETUP_THRESHOLD" | bc -l) )); then echo "✅ Good"; else echo "⚠️ Slow"; fi) |
| Build | $(echo "$current_results" | jq -r '.build_time') | $(if (( $(echo "$(echo "$current_results" | jq -r '.build_time') <= $BUILD_THRESHOLD" | bc -l) )); then echo "✅ Good"; else echo "⚠️ Slow"; fi) |
| Unit Tests | $(echo "$current_results" | jq -r '.test_time') | $(if (( $(echo "$(echo "$current_results" | jq -r '.test_time') <= $TEST_THRESHOLD" | bc -l) )); then echo "✅ Good"; else echo "⚠️ Slow"; fi) |
| Integration Tests | $(echo "$current_results" | jq -r '.integration_time') | $(if (( $(echo "$(echo "$current_results" | jq -r '.integration_time') <= $TEST_THRESHOLD" | bc -l) )); then echo "✅ Good"; else echo "⚠️ Slow"; fi) |
| Linting | $(echo "$current_results" | jq -r '.lint_time') | $(if (( $(echo "$(echo "$current_results" | jq -r '.lint_time') <= $LINT_THRESHOLD" | bc -l) )); then echo "✅ Good"; else echo "⚠️ Slow"; fi) |
| Docker Build | $(echo "$current_results" | jq -r '.docker_time') | $(if (( $(echo "$(echo "$current_results" | jq -r '.docker_time') <= $DOCKER_THRESHOLD" | bc -l) )); then echo "✅ Good"; else echo "⚠️ Slow"; fi) |
| **Total Pipeline** | **$(echo "$current_results" | jq -r '.total_time')** | $(if (( $(echo "$(echo "$current_results" | jq -r '.total_time') <= $TOTAL_THRESHOLD" | bc -l) )); then echo "✅ Excellent"; else echo "⚠️ Needs Optimization"; fi) |

EOF
    fi
    
    # Add baseline comparison if both exist
    if [[ "$current_results" != "{}" ]] && [[ "$baseline_results" != "{}" ]]; then
        cat >> "$REPORT_FILE" << 'EOF'
### Performance Comparison vs Baseline

| Stage | Baseline | Current | Change | Trend |
|-------|----------|---------|--------|-------|
EOF
        
        # Helper function for report comparison
        add_comparison_row() {
            local metric="$1"
            local label="$2"
            local current baseline diff percentage
            
            current=$(echo "$current_results" | jq -r ".$metric")
            baseline=$(echo "$baseline_results" | jq -r ".$metric")
            
            if [[ "$baseline" == "0" ]] || [[ "$current" == "0" ]]; then
                echo "| $label | ${baseline}s | ${current}s | N/A | - |" >> "$REPORT_FILE"
                return
            fi
            
            diff=$(echo "$current - $baseline" | bc -l)
            percentage=$(echo "scale=1; ($diff / $baseline) * 100" | bc -l)
            
            local trend="📊 Stable"
            if (( $(echo "$percentage > 20" | bc -l) )); then
                trend="🔴 Major Regression"
            elif (( $(echo "$percentage > 5" | bc -l) )); then
                trend="🟡 Minor Regression"
            elif (( $(echo "$percentage < -20" | bc -l) )); then
                trend="🟢 Major Improvement"
            elif (( $(echo "$percentage < -5" | bc -l) )); then
                trend="🟢 Improvement"
            fi
            
            printf "| %s | %.2fs | %.2fs | %+.1f%% | %s |\n" "$label" "$baseline" "$current" "$percentage" "$trend" >> "$REPORT_FILE"
        }
        
        add_comparison_row "setup_time" "Setup & Dependencies"
        add_comparison_row "build_time" "Build"  
        add_comparison_row "test_time" "Unit Tests"
        add_comparison_row "integration_time" "Integration Tests"
        add_comparison_row "lint_time" "Linting"
        add_comparison_row "docker_time" "Docker Build"
        add_comparison_row "total_time" "**Total Pipeline**"
    fi
    
    # Add recommendations
    cat >> "$REPORT_FILE" << 'EOF'

## Optimization Recommendations

### Implemented Optimizations ✅

1. **Enhanced Dependency Caching**: Go modules and build cache optimization
2. **Parallel Job Execution**: Concurrent test execution and build parallelization  
3. **Docker Layer Caching**: Multi-stage builds with efficient layer management
4. **Reduced Timeouts**: Optimized timeout configurations for faster failure detection
5. **Smart Test Selection**: Conditional test execution based on code changes

### Potential Future Improvements 🚀

1. **Build Matrix Optimization**: Further parallelize different test configurations
2. **Incremental Builds**: Implement incremental compilation for large codebases
3. **Test Sharding**: Distribute tests across multiple runners for large test suites
4. **Cache Warming**: Pre-populate caches during off-peak hours
5. **Resource Scaling**: Dynamic resource allocation based on workload

## Performance Thresholds

| Stage | Target | Threshold | Notes |
|-------|--------|-----------|-------|
| Setup | < 1 min | < 2 min | Dependency resolution and caching |
| Build | < 2 min | < 5 min | Go compilation and binary generation |
| Tests | < 5 min | < 15 min | Unit and integration test execution |
| Linting | < 1 min | < 3 min | Code quality and style checks |
| Docker | < 10 min | < 15 min | Container image build and optimization |
| **Total** | **< 15 min** | **< 30 min** | **Complete pipeline execution** |

## Next Steps

1. **Monitor Trends**: Track performance metrics over time
2. **Investigate Regressions**: Address any performance degradation quickly
3. **Optimize Bottlenecks**: Focus on stages exceeding thresholds
4. **Update Baselines**: Refresh baselines after major optimizations

---

**Report Generated By:** CI Performance Benchmark Tool  
**Last Updated:** `$(get_timestamp)`
EOF
    
    log_success "Performance report generated: $REPORT_FILE"
    
    # Display quick summary
    echo
    echo "📋 Report Summary:"
    echo "=================="
    echo "📁 Report location: $REPORT_FILE"
    
    if [[ "$current_results" != "{}" ]]; then
        local total_time
        total_time=$(echo "$current_results" | jq -r '.total_time')
        echo "⏱️  Current total time: ${total_time}s"
        
        if (( $(echo "$total_time <= $TOTAL_THRESHOLD" | bc -l) )); then
            log_success "Pipeline performance is within target thresholds"
        else
            log_warning "Pipeline performance exceeds target thresholds"
        fi
    fi
}

# Quick benchmark (setup, build, lint only)
quick_benchmark() {
    log_info "🚀 Running quick performance benchmark..."
    compare_performance "quick"
}

# Main execution
main() {
    local command="${1:-help}"
    
    init_benchmark_dir
    
    case "$command" in
        "baseline")
            save_baseline
            ;;
        "compare"|"benchmark")
            compare_performance "full"
            ;;
        "quick")
            quick_benchmark
            ;;
        "report")
            generate_report
            ;;
        "help"|"--help"|"-h")
            echo "CI/CD Pipeline Performance Benchmark Tool"
            echo
            echo "Usage: $0 [command]"
            echo
            echo "Commands:"
            echo "  baseline    Create performance baseline"
            echo "  compare     Compare current performance against baseline"
            echo "  quick       Quick benchmark (setup, build, lint only)"
            echo "  report      Generate detailed performance report"
            echo "  help        Show this help message"
            echo
            echo "Examples:"
            echo "  $0 baseline              # Create initial baseline"
            echo "  $0 compare               # Compare against baseline"
            echo "  $0 quick                 # Quick performance check"
            echo "  $0 report                # Generate markdown report"
            ;;
        *)
            log_error "Unknown command: $command"
            echo "Use '$0 help' for usage information."
            exit 1
            ;;
    esac
}

# Check dependencies
check_dependencies() {
    local missing_deps=()
    
    command -v jq >/dev/null 2>&1 || missing_deps+=("jq")
    command -v bc >/dev/null 2>&1 || missing_deps+=("bc")
    command -v go >/dev/null 2>&1 || missing_deps+=("go")
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        echo "Please install the missing dependencies and try again."
        exit 1
    fi
}

# Ensure we have required dependencies
check_dependencies

# Run main function
main "$@"