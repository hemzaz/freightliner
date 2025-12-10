#!/bin/bash

# CICD Pipeline Status Dashboard
# Quick at-a-glance view of pipeline health
# Usage: ./cicd-status.sh [--watch]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Watch mode
WATCH_MODE=false
if [[ "$1" == "--watch" ]]; then
    WATCH_MODE=true
    WATCH_INTERVAL=30  # seconds
fi

# Check for required tools
command -v gh >/dev/null 2>&1 || { echo -e "${RED}✗${NC} GitHub CLI (gh) required"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo -e "${RED}✗${NC} jq required"; exit 1; }

# Function to display status
display_status() {
    clear

    # Header
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}  ${BOLD}Freightliner CICD Pipeline Status${NC}                           ${BLUE}║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "Repository: ${CYAN}$(gh repo view --json nameWithOwner -q .nameWithOwner)${NC}"
    echo -e "Last Updated: ${CYAN}$(date '+%Y-%m-%d %H:%M:%S')${NC}"
    echo ""

    # Recent Runs (Last 10)
    echo -e "${BOLD}Recent Workflow Runs (Last 10):${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    gh run list --limit 10 --json name,status,conclusion,createdAt,databaseId | jq -r '
        .[] |
        "\(.databaseId)|\(.name)|\(.status)|\(.conclusion // "running")|\(.createdAt)"
    ' | while IFS='|' read -r id name status conclusion created; do
        # Status symbol
        if [[ "$conclusion" == "success" ]]; then
            symbol="${GREEN}✓${NC}"
        elif [[ "$conclusion" == "failure" ]]; then
            symbol="${RED}✗${NC}"
        elif [[ "$conclusion" == "cancelled" ]]; then
            symbol="${YELLOW}⊘${NC}"
        elif [[ "$conclusion" == "timed_out" ]]; then
            symbol="${RED}⏱${NC}"
        elif [[ "$status" == "in_progress" ]]; then
            symbol="${CYAN}⟳${NC}"
        else
            symbol="${YELLOW}?${NC}"
        fi

        # Format timestamp
        timestamp=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$created" "+%m/%d %H:%M" 2>/dev/null || echo "$created")

        # Truncate workflow name if too long
        name_short=$(echo "$name" | cut -c1-40)

        echo -e " ${symbol} [${timestamp}] ${name_short}"
    done
    echo ""

    # Summary Statistics
    echo -e "${BOLD}Pipeline Health (Last 24 Hours):${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Get 24h stats
    YESTERDAY=$(date -u -v-1d +%Y-%m-%dT%H:%M:%S 2>/dev/null || date -u -d '1 day ago' +%Y-%m-%dT%H:%M:%S)

    TOTAL=$(gh run list --limit 100 --json createdAt,conclusion 2>/dev/null | \
        jq --arg date "$YESTERDAY" '[.[] | select(.createdAt > $date)] | length')

    SUCCESS=$(gh run list --status success --limit 100 --json createdAt 2>/dev/null | \
        jq --arg date "$YESTERDAY" '[.[] | select(.createdAt > $date)] | length')

    FAILURE=$(gh run list --status failure --limit 100 --json createdAt 2>/dev/null | \
        jq --arg date "$YESTERDAY" '[.[] | select(.createdAt > $date)] | length')

    CANCELLED=$(gh run list --status cancelled --limit 100 --json createdAt 2>/dev/null | \
        jq --arg date "$YESTERDAY" '[.[] | select(.createdAt > $date)] | length')

    IN_PROGRESS=$(gh run list --status in_progress --limit 100 --json createdAt 2>/dev/null | \
        jq --arg date "$YESTERDAY" '[.[] | select(.createdAt > $date)] | length')

    # Calculate success rate
    if [[ "$TOTAL" -gt 0 ]]; then
        SUCCESS_RATE=$(echo "scale=1; ($SUCCESS * 100) / $TOTAL" | bc)
    else
        SUCCESS_RATE="N/A"
    fi

    # Display stats
    echo -e "  Total Runs:      ${CYAN}${TOTAL}${NC}"
    echo -e "  ${GREEN}✓${NC} Successful:    ${GREEN}${SUCCESS}${NC}"
    echo -e "  ${RED}✗${NC} Failed:        ${RED}${FAILURE}${NC}"
    echo -e "  ${YELLOW}⊘${NC} Cancelled:     ${YELLOW}${CANCELLED}${NC}"
    echo -e "  ${CYAN}⟳${NC} In Progress:   ${CYAN}${IN_PROGRESS}${NC}"

    # Success rate with color coding
    if [[ "$SUCCESS_RATE" != "N/A" ]]; then
        if (( $(echo "$SUCCESS_RATE >= 95" | bc -l) )); then
            echo -e "  Success Rate:    ${GREEN}${SUCCESS_RATE}%${NC} ${GREEN}✓${NC}"
        elif (( $(echo "$SUCCESS_RATE >= 85" | bc -l) )); then
            echo -e "  Success Rate:    ${YELLOW}${SUCCESS_RATE}%${NC} ${YELLOW}⚠${NC}"
        else
            echo -e "  Success Rate:    ${RED}${SUCCESS_RATE}%${NC} ${RED}✗${NC}"
        fi
    else
        echo -e "  Success Rate:    ${YELLOW}N/A${NC} (no runs)"
    fi
    echo ""

    # Critical Workflows Status
    echo -e "${BOLD}Critical Workflows Status:${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    CRITICAL_WORKFLOWS=(
        "consolidated-ci.yml:CI Pipeline"
        "integration-tests.yml:Integration Tests"
        "docker-publish.yml:Docker Publish"
        "deploy.yml:Deployment"
    )

    for workflow_info in "${CRITICAL_WORKFLOWS[@]}"; do
        IFS=':' read -r workflow_file workflow_name <<< "$workflow_info"

        # Get last run for this workflow
        LAST_RUN=$(gh run list --workflow="$workflow_file" --limit 1 --json conclusion,createdAt 2>/dev/null | \
            jq -r '.[0] | "\(.conclusion // "none")|\(.createdAt // "never")"' 2>/dev/null || echo "error|never")

        IFS='|' read -r conclusion created <<< "$LAST_RUN"

        if [[ "$conclusion" == "success" ]]; then
            symbol="${GREEN}✓${NC}"
            status="${GREEN}Healthy${NC}"
        elif [[ "$conclusion" == "failure" ]]; then
            symbol="${RED}✗${NC}"
            status="${RED}Failed${NC}"
        elif [[ "$conclusion" == "none" ]]; then
            symbol="${CYAN}⟳${NC}"
            status="${CYAN}Running${NC}"
        elif [[ "$conclusion" == "error" ]]; then
            symbol="${YELLOW}?${NC}"
            status="${YELLOW}Unknown${NC}"
        else
            symbol="${YELLOW}⊘${NC}"
            status="${YELLOW}${conclusion}${NC}"
        fi

        # Format timestamp
        if [[ "$created" != "never" && "$created" != "error" ]]; then
            timestamp=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$created" "+%m/%d %H:%M" 2>/dev/null || echo "$created")
        else
            timestamp="Never run"
        fi

        printf "  %-25s %b [%s]\n" "$workflow_name" "$symbol $status" "$timestamp"
    done
    echo ""

    # Active Issues Check
    echo -e "${BOLD}Quick Health Checks:${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Check for timeout failures
    TIMEOUT_FAILURES=$(gh run list --status timed_out --limit 100 --json createdAt 2>/dev/null | \
        jq --arg date "$YESTERDAY" '[.[] | select(.createdAt > $date)] | length')

    if [[ "$TIMEOUT_FAILURES" -eq 0 ]]; then
        echo -e "  ${GREEN}✓${NC} No timeout failures (24h)"
    else
        echo -e "  ${RED}✗${NC} ${TIMEOUT_FAILURES} timeout failures (24h)"
    fi

    # Check for recent failures
    if [[ "$FAILURE" -eq 0 ]]; then
        echo -e "  ${GREEN}✓${NC} No workflow failures (24h)"
    elif [[ "$FAILURE" -le 3 ]]; then
        echo -e "  ${YELLOW}⚠${NC} ${FAILURE} workflow failures (24h)"
    else
        echo -e "  ${RED}✗${NC} ${FAILURE} workflow failures (24h)"
    fi

    # Check if deployments are healthy
    DEPLOY_FAILURES=$(gh run list --workflow=deploy.yml --status failure --limit 10 2>/dev/null | wc -l)
    if [[ "$DEPLOY_FAILURES" -eq 0 ]]; then
        echo -e "  ${GREEN}✓${NC} No deployment failures"
    else
        echo -e "  ${YELLOW}⚠${NC} ${DEPLOY_FAILURES} recent deployment failures"
    fi

    # Check optimization status
    if [[ -f ".github/SESSION_SUMMARY.md" ]]; then
        echo -e "  ${GREEN}✓${NC} Session 4 optimizations deployed"
    else
        echo -e "  ${YELLOW}⚠${NC} Session 4 optimizations not found"
    fi

    echo ""

    # Quick Actions
    echo -e "${BOLD}Quick Actions:${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "  • View failed runs:     ${CYAN}gh run list --status failure${NC}"
    echo -e "  • View logs:            ${CYAN}gh run view <run-id> --log${NC}"
    echo -e "  • Run validation:       ${CYAN}./.github/scripts/validate-optimizations.sh${NC}"
    echo -e "  • Monitoring guide:     ${CYAN}cat .github/WEEK1_MONITORING_GUIDE.md${NC}"
    echo ""

    if [[ "$WATCH_MODE" == "true" ]]; then
        echo -e "${YELLOW}[Watch Mode] Refreshing in ${WATCH_INTERVAL}s... Press Ctrl+C to exit${NC}"
    fi
}

# Main loop
if [[ "$WATCH_MODE" == "true" ]]; then
    while true; do
        display_status
        sleep "$WATCH_INTERVAL"
    done
else
    display_status
fi
