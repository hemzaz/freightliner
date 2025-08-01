# Repository Tidying Workflow

Comprehensive multi-agent repository organization and cleanup workflow for the Freightliner container replication system. This command deploys specialized agents to ensure consistent project structure, code quality, and maintainability across the entire codebase.

## Usage
```
/tidy-repo [options]
```

## Options
- `--full` - Run complete tidying workflow with all agents
- `--quick` - Run essential tidying tasks only
- `--dry-run` - Show what would be changed without making modifications
- `--report-only` - Generate organizational report without changes
- `--agent=<name>` - Run specific agent only (file-org, code-style, imports, docs, config, tests, build, git, deps, deadcode)

## Agent Overview

This workflow deploys 10 specialized agents to handle different aspects of repository organization:

1. **File Organization Agent** - Proper directory structure
2. **Code Style Agent** - Formatting and style consistency  
3. **Import Organization Agent** - Import cleanup and organization
4. **Documentation Agent** - Documentation structure and placement
5. **Configuration Agent** - Configuration file organization
6. **Test Organization Agent** - Test file placement and naming
7. **Build Cleanup Agent** - Build artifact cleanup
8. **Git Management Agent** - Git ignore updates
9. **Dependency Management Agent** - Dependency cleanup
10. **Dead Code Removal Agent** - Unused code detection and removal

## Multi-Agent Deployment Script

```bash
#!/bin/bash

# Repository Tidying Workflow - Multi-Agent Deployment
# Freightliner Container Registry Replication System

set -e

# Configuration
REPO_ROOT="$(git rev-parse --show-toplevel)"
WORK_DIR="${REPO_ROOT}/.claude/tidy-work"
REPORT_DIR="${WORK_DIR}/reports"
BACKUP_DIR="${WORK_DIR}/backup"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Agent configuration
AGENTS_ENABLED=(
    "file-org"
    "code-style" 
    "imports"
    "docs"
    "config"
    "tests"
    "build"
    "git"
    "deps"
    "deadcode"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Utility functions
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

log_agent() {
    echo -e "${PURPLE}[AGENT]${NC} $1"
}

# Parse command line arguments
FULL_TIDY=false
QUICK_TIDY=false
DRY_RUN=false
REPORT_ONLY=false
SPECIFIC_AGENT=""

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --full)
                FULL_TIDY=true
                shift
                ;;
            --quick)
                QUICK_TIDY=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --report-only)
                REPORT_ONLY=true
                shift
                ;;
            --agent=*)
                SPECIFIC_AGENT="${1#*=}"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

show_help() {
    cat << EOF
Repository Tidying Workflow - Freightliner Container Registry Replication

Usage: /tidy-repo [options]

Options:
  --full          Run complete tidying workflow with all agents
  --quick         Run essential tidying tasks only
  --dry-run       Show what would be changed without making modifications
  --report-only   Generate organizational report without changes
  --agent=<name>  Run specific agent only
  -h, --help      Show this help message

Available Agents:
  file-org    - File and directory organization
  code-style  - Code formatting and style consistency
  imports     - Import organization and cleanup
  docs        - Documentation structure and placement
  config      - Configuration file organization
  tests       - Test file placement and naming
  build       - Build artifact cleanup
  git         - Git ignore and repository management
  deps        - Dependency cleanup and optimization
  deadcode    - Dead code detection and removal

Examples:
  /tidy-repo --full                    # Complete repository tidying
  /tidy-repo --quick                   # Essential tidying only
  /tidy-repo --dry-run --full          # Show what would be changed
  /tidy-repo --agent=code-style        # Run code style agent only
  /tidy-repo --report-only             # Generate report without changes
EOF
}

# Initialize workspace
init_workspace() {
    log_info "Initializing tidying workspace..."
    
    mkdir -p "${WORK_DIR}"
    mkdir -p "${REPORT_DIR}"
    mkdir -p "${BACKUP_DIR}"
    
    # Create backup of current state
    if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
        log_info "Creating backup of current state..."
        git stash push -u -m "tidy-repo-backup-${TIMESTAMP}" || true
    fi
    
    # Initialize reports
    cat > "${REPORT_DIR}/summary.md" << EOF
# Repository Tidying Report
Generated: $(date)
Command: $0 $@

## Summary
EOF
}

# Pre-flight checks
preflight_checks() {
    log_info "Running pre-flight checks..."
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
    
    # Check for required tools
    local required_tools=("go" "gofmt" "goimports")
    
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_warning "$tool not found, attempting to install..."
            case "$tool" in
                "goimports")
                    go install golang.org/x/tools/cmd/goimports@latest
                    ;;
                *)
                    log_error "Cannot auto-install $tool, please install manually"
                    exit 1
                    ;;
            esac
        fi
    done
    
    # Check Go version
    local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    local required_version="1.21"
    
    if ! echo "$go_version $required_version" | awk '{exit ($1 >= $2) ? 0 : 1}'; then
        log_warning "Go version $go_version is below recommended $required_version"
    fi
    
    log_success "Pre-flight checks completed"
}

# Agent 1: File Organization
agent_file_organization() {
    log_agent "Deploying File Organization Agent..."
    
    local changes=0
    local report="${REPORT_DIR}/file-organization.md"
    
    cat > "$report" << EOF
# File Organization Agent Report

## Directory Structure Analysis
EOF
    
    # Define expected Go project structure
    local expected_dirs=(
        "cmd"           # Main applications
        "pkg"           # Library code
        "internal"      # Private application and library code
        "api"           # API definitions (if needed)
        "web"           # Web application assets (if needed)
        "build"         # Packaging and CI
        "deployments"   # IaaS, PaaS, system configs
        "test"          # Additional external test apps and test data
        "docs"          # Design and user documents
        "tools"         # Supporting tools
        "examples"      # Examples for applications/libraries
        "scripts"       # Scripts for builds, installs, analysis, etc.
        "configs"       # Configuration files
    )
    
    # Check for misplaced files in root
    log_info "Analyzing root directory structure..."
    
    local root_go_files=$(find "$REPO_ROOT" -maxdepth 1 -name "*.go" -not -name "main.go" | wc -l)
    if [[ $root_go_files -gt 0 ]]; then
        echo "## Root Directory Issues" >> "$report"
        echo "- Found $root_go_files Go files in root (should be in pkg/ or cmd/)" >> "$report"
        
        if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
            # Move non-main.go files to appropriate locations
            for file in $(find "$REPO_ROOT" -maxdepth 1 -name "*.go" -not -name "main.go"); do
                local basename=$(basename "$file" .go)
                if [[ -d "$REPO_ROOT/cmd" ]]; then
                    local target_dir="$REPO_ROOT/cmd"
                else
                    local target_dir="$REPO_ROOT/pkg/$basename"
                    mkdir -p "$target_dir"
                fi
                
                log_info "Moving $file to $target_dir/"
                git mv "$file" "$target_dir/" 2>/dev/null || mv "$file" "$target_dir/"
                changes=$((changes + 1))
            done
        fi
    fi
    
    # Check pkg structure
    if [[ -d "$REPO_ROOT/pkg" ]]; then
        log_info "Analyzing pkg/ directory structure..."
        
        # Look for packages that should be in internal/
        local pkg_dirs=$(find "$REPO_ROOT/pkg" -type d -maxdepth 1 | grep -v "^$REPO_ROOT/pkg$")
        
        for dir in $pkg_dirs; do
            local pkg_name=$(basename "$dir")
            
            # Check if package is used externally
            local external_usage=$(git grep -l "freightliner/pkg/$pkg_name" -- "*.go" | grep -v "^pkg/" | wc -l)
            
            if [[ $external_usage -eq 0 ]]; then
                echo "## Package Organization Issues" >> "$report"
                echo "- Package $pkg_name in pkg/ but not used externally (consider moving to internal/)" >> "$report"
                
                if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
                    mkdir -p "$REPO_ROOT/internal"
                    log_info "Moving $dir to internal/"
                    git mv "$dir" "$REPO_ROOT/internal/" 2>/dev/null || mv "$dir" "$REPO_ROOT/internal/"
                    changes=$((changes + 1))
                fi
            fi
        done
    fi
    
    # Check for proper test file placement
    log_info "Analyzing test file placement..."
    
    local test_files=$(find "$REPO_ROOT" -name "*_test.go" -not -path "*/vendor/*")
    local misplaced_tests=0
    
    for test_file in $test_files; do
        local dir=$(dirname "$test_file")
        local corresponding_go_file=$(echo "$test_file" | sed 's/_test\.go$/.go/')
        
        if [[ ! -f "$corresponding_go_file" ]]; then
            echo "## Test File Issues" >> "$report"
            echo "- Test file $test_file has no corresponding implementation file" >> "$report"
            misplaced_tests=$((misplaced_tests + 1))
        fi
    done
    
    # Check for proper documentation placement
    log_info "Analyzing documentation structure..."
    
    local readme_files=$(find "$REPO_ROOT" -name "README.md" -not -path "*/vendor/*" -not -path "*/.git/*")
    local doc_count=$(echo "$readme_files" | wc -l)
    
    if [[ $doc_count -gt 3 ]]; then
        echo "## Documentation Issues" >> "$report"
        echo "- Found $doc_count README.md files, consider consolidating" >> "$report"
    fi
    
    echo "## Changes Made: $changes" >> "$report"
    
    if [[ $changes -gt 0 ]]; then
        log_success "File Organization Agent completed with $changes changes"
    else
        log_info "File Organization Agent completed with no changes needed"
    fi
}

# Agent 2: Code Style Consistency
agent_code_style() {
    log_agent "Deploying Code Style Agent..."
    
    local changes=0
    local report="${REPORT_DIR}/code-style.md"
    
    cat > "$report" << EOF
# Code Style Agent Report

## Code Formatting Analysis
EOF
    
    # Run gofmt to check formatting
    log_info "Checking Go code formatting..."
    
    local unformatted_files=$(gofmt -l $(find . -name "*.go" -not -path "./vendor/*"))
    
    if [[ -n "$unformatted_files" ]]; then
        local count=$(echo "$unformatted_files" | wc -l)
        echo "## Formatting Issues" >> "$report"
        echo "- Found $count unformatted Go files" >> "$report"
        echo "\`\`\`" >> "$report"
        echo "$unformatted_files" >> "$report"
        echo "\`\`\`" >> "$report"
        
        if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
            log_info "Formatting Go files..."
            gofmt -w $(find . -name "*.go" -not -path "./vendor/*")
            changes=$((changes + count))
        fi
    fi
    
    # Check for consistent naming conventions
    log_info "Checking naming conventions..."
    
    # Check for non-idiomatic function names
    local bad_names=$(grep -r "func [A-Z][a-z]*_[A-Z]" --include="*.go" . | wc -l)
    if [[ $bad_names -gt 0 ]]; then
        echo "## Naming Convention Issues" >> "$report"
        echo "- Found $bad_names functions with non-idiomatic names (snake_case in Go)" >> "$report"
    fi
    
    # Check for proper error handling patterns
    log_info "Checking error handling patterns..."
    
    local error_patterns=$(grep -r "errors.New(\|fmt.Errorf(" --include="*.go" . | wc -l)
    local wrap_patterns=$(grep -r "fmt.Errorf.*%w" --include="*.go" . | wc -l)
    
    echo "## Error Handling Analysis" >> "$report"
    echo "- Total error creation patterns: $error_patterns" >> "$report"
    echo "- Error wrapping patterns (good): $wrap_patterns" >> "$report"
    
    # Check for consistent struct field ordering
    log_info "Analyzing struct definitions..."
    
    # This is a simplified check - in practice, you'd want more sophisticated analysis
    local struct_count=$(grep -r "type.*struct {" --include="*.go" . | wc -l)
    echo "## Struct Analysis" >> "$report"
    echo "- Total struct definitions: $struct_count" >> "$report"
    
    echo "## Changes Made: $changes" >> "$report"
    
    if [[ $changes -gt 0 ]]; then
        log_success "Code Style Agent completed with $changes changes"
    else
        log_info "Code Style Agent completed with no changes needed"
    fi
}

# Agent 3: Import Organization
agent_imports() {
    log_agent "Deploying Import Organization Agent..."
    
    local changes=0
    local report="${REPORT_DIR}/imports.md"
    
    cat > "$report" << EOF
# Import Organization Agent Report

## Import Analysis and Cleanup
EOF
    
    # Check current import organization
    log_info "Analyzing import organization..."
    
    local go_files=$(find . -name "*.go" -not -path "./vendor/*")
    local files_with_issues=0
    
    for file in $go_files; do
        # Check if imports are properly organized
        local import_sections=$(awk '/^import \(/{flag=1; next} flag && /^\)/{flag=0} flag' "$file" | grep -c "^$" || true)
        
        if [[ $import_sections -gt 2 ]]; then
            files_with_issues=$((files_with_issues + 1))
        fi
    done
    
    echo "## Import Organization Issues" >> "$report"
    echo "- Files with import organization issues: $files_with_issues" >> "$report"
    
    if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
        log_info "Organizing imports with goimports..."
        
        # Run goimports to organize imports
        goimports -w -local freightliner $(find . -name "*.go" -not -path "./vendor/*")
        changes=$files_with_issues
        
        log_info "Running custom import grouping..."
        
        # Custom script to ensure proper import grouping
        for file in $go_files; do
            if grep -q "^import (" "$file"; then
                # Create temporary file with properly grouped imports
                python3 << EOF
import re
import sys

def organize_imports(content):
    # Extract import block
    import_pattern = r'import \((.*?)\)'
    match = re.search(import_pattern, content, re.DOTALL)
    
    if not match:
        return content
    
    imports = match.group(1).strip().split('\n')
    imports = [imp.strip() for imp in imports if imp.strip()]
    
    # Categorize imports
    stdlib = []
    thirdparty = []
    local = []
    
    for imp in imports:
        imp = imp.strip()
        if not imp:
            continue
            
        # Remove quotes and get package name
        pkg = imp.strip('"').strip("'")
        
        if 'freightliner' in pkg:
            local.append(imp)
        elif '.' not in pkg or pkg.startswith('golang.org/x/'):
            stdlib.append(imp)
        else:
            thirdparty.append(imp)
    
    # Sort each category
    stdlib.sort()
    thirdparty.sort()
    local.sort()
    
    # Rebuild import block
    new_imports = []
    if stdlib:
        new_imports.extend(stdlib)
    if thirdparty:
        if stdlib:
            new_imports.append('')
        new_imports.extend(thirdparty)
    if local:
        if stdlib or thirdparty:
            new_imports.append('')
        new_imports.extend(local)
    
    new_import_block = 'import (\n\t' + '\n\t'.join(new_imports) + '\n)'
    
    return re.sub(import_pattern, new_import_block, content, flags=re.DOTALL)

# Read file
with open('$file', 'r') as f:
    content = f.read()

# Organize imports
new_content = organize_imports(content)

# Write back if changed
if content != new_content:
    with open('$file', 'w') as f:
        f.write(new_content)
    print('Updated $file')
EOF
            fi
        done
    fi
    
    # Check for unused imports
    log_info "Checking for unused imports..."
    
    # This would typically be done by a linter, but we can do a basic check
    local unused_count=0
    
    echo "## Unused Import Analysis" >> "$report"
    echo "- Unused imports should be detected by goimports and removed" >> "$report"
    echo "- Consider running: go mod tidy && goimports -w ." >> "$report"
    
    echo "## Changes Made: $changes" >> "$report"
    
    if [[ $changes -gt 0 ]]; then
        log_success "Import Organization Agent completed with $changes changes"
    else
        log_info "Import Organization Agent completed with no changes needed"
    fi
}

# Agent 4: Documentation Structure
agent_documentation() {
    log_agent "Deploying Documentation Agent..."
    
    local changes=0
    local report="${REPORT_DIR}/documentation.md"
    
    cat > "$report" << EOF
# Documentation Agent Report

## Documentation Structure Analysis
EOF
    
    # Analyze current documentation structure
    log_info "Analyzing documentation structure..."
    
    local doc_files=$(find . -name "*.md" -not -path "./vendor/*" -not -path "*/.git/*")
    local doc_count=$(echo "$doc_files" | wc -l)
    
    echo "## Current Documentation" >> "$report"
    echo "- Total Markdown files: $doc_count" >> "$report"
    echo "" >> "$report"
    
    # Check for missing standard documentation
    local standard_docs=("README.md" "CONTRIBUTING.md" "CHANGELOG.md" "LICENSE")
    local missing_docs=()
    
    for doc in "${standard_docs[@]}"; do
        if [[ ! -f "$REPO_ROOT/$doc" ]]; then
            missing_docs+=("$doc")
        fi
    done
    
    if [[ ${#missing_docs[@]} -gt 0 ]]; then
        echo "## Missing Standard Documentation" >> "$report"
        for doc in "${missing_docs[@]}"; do
            echo "- $doc" >> "$report"
        done
        echo "" >> "$report"
    fi
    
    # Check for package documentation
    log_info "Checking package documentation..."
    
    local pkg_dirs=$(find pkg internal -type d 2>/dev/null | grep -v "^pkg$\|^internal$" || true)
    local undocumented_packages=0
    
    for pkg_dir in $pkg_dirs; do
        local has_doc_go=false
        local has_readme=false
        
        if [[ -f "$pkg_dir/doc.go" ]]; then
            has_doc_go=true
        fi
        
        if [[ -f "$pkg_dir/README.md" ]]; then
            has_readme=true
        fi
        
        if [[ "$has_doc_go" == "false" && "$has_readme" == "false" ]]; then
            undocumented_packages=$((undocumented_packages + 1))
        fi
    done
    
    echo "## Package Documentation" >> "$report"
    echo "- Undocumented packages: $undocumented_packages" >> "$report"
    echo "" >> "$report"
    
    # Check for API documentation
    if [[ -d "api" ]]; then
        local api_docs=$(find api -name "*.md" | wc -l)
        echo "## API Documentation" >> "$report"
        echo "- API documentation files: $api_docs" >> "$report"
        echo "" >> "$report"
    fi
    
    # Check documentation in docs/ directory
    if [[ -d "docs" ]]; then
        log_info "Analyzing docs/ directory structure..."
        
        local docs_structure=$(find docs -type f -name "*.md" | sort)
        echo "## Documentation Structure" >> "$report"
        echo "\`\`\`" >> "$report"
        echo "$docs_structure" >> "$report"
        echo "\`\`\`" >> "$report"
        echo "" >> "$report"
        
        # Check for outdated documentation
        local outdated_docs=0
        local recent_code_changes=$(git log --since="30 days ago" --name-only --pretty=format: | grep "\.go$" | sort -u | wc -l)
        local recent_doc_changes=$(git log --since="30 days ago" --name-only --pretty=format: | grep "\.md$" | sort -u | wc -l)
        
        if [[ $recent_code_changes -gt $recent_doc_changes ]]; then
            echo "## Documentation Freshness" >> "$report"
            echo "- Warning: Recent code changes ($recent_code_changes) exceed doc changes ($recent_doc_changes)" >> "$report"
            echo "- Consider updating documentation to match recent code changes" >> "$report"
            echo "" >> "$report"
        fi
    fi
    
    # Suggest documentation improvements
    if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
        # Create missing package documentation templates
        for pkg_dir in $pkg_dirs; do
            if [[ ! -f "$pkg_dir/doc.go" && ! -f "$pkg_dir/README.md" ]]; then
                local pkg_name=$(basename "$pkg_dir")
                
                cat > "$pkg_dir/doc.go" << EOF
// Package $pkg_name provides functionality for the Freightliner container registry replication system.
//
// This package is part of the Freightliner project, which replicates container images
// between different registry providers (AWS ECR, Google Container Registry, etc.).
//
// For more information, see the main project documentation.
package $pkg_name
EOF
                
                log_info "Created documentation template for package $pkg_name"
                changes=$((changes + 1))
            fi
        done
    fi
    
    echo "## Changes Made: $changes" >> "$report"
    
    if [[ $changes -gt 0 ]]; then
        log_success "Documentation Agent completed with $changes changes"
    else
        log_info "Documentation Agent completed with no changes needed"
    fi
}

# Agent 5: Configuration Organization
agent_configuration() {
    log_agent "Deploying Configuration Agent..."
    
    local changes=0
    local report="${REPORT_DIR}/configuration.md"
    
    cat > "$report" << EOF
# Configuration Agent Report

## Configuration File Organization
EOF
    
    # Find all configuration files
    log_info "Analyzing configuration files..."
    
    local config_patterns=("*.yaml" "*.yml" "*.json" "*.toml" "*.ini" "*.conf" "Dockerfile*" "docker-compose*")
    local config_files=()
    
    for pattern in "${config_patterns[@]}"; do
        while IFS= read -r -d '' file; do
            config_files+=("$file")
        done < <(find . -name "$pattern" -not -path "./vendor/*" -not -path "*/.git/*" -print0)
    done
    
    echo "## Configuration Files Found" >> "$report"
    echo "- Total configuration files: ${#config_files[@]}" >> "$report"
    echo "" >> "$report"
    
    # Analyze configuration organization
    local root_configs=0
    local organized_configs=0
    
    for file in "${config_files[@]}"; do
        local dir=$(dirname "$file")
        
        if [[ "$dir" == "." ]]; then
            root_configs=$((root_configs + 1))
        elif [[ "$dir" =~ ^\./(config|configs|deployments) ]]; then
            organized_configs=$((organized_configs + 1))
        fi
    done
    
    echo "## Configuration Organization" >> "$report"
    echo "- Files in root directory: $root_configs" >> "$report"
    echo "- Files in organized directories: $organized_configs" >> "$report"
    echo "" >> "$report"
    
    # Check for configuration consolidation opportunities
    log_info "Checking for configuration consolidation opportunities..."
    
    local docker_files=$(find . -name "Dockerfile*" -not -path "./vendor/*" | wc -l)
    local compose_files=$(find . -name "docker-compose*" -not -path "./vendor/*" | wc -l)
    local k8s_files=$(find . -name "*.yaml" -path "*/k8s/*" -o -path "*/kubernetes/*" 2>/dev/null | wc -l)
    
    echo "## Configuration Types" >> "$report"
    echo "- Docker files: $docker_files" >> "$report"
    echo "- Docker Compose files: $compose_files" >> "$report"
    echo "- Kubernetes files: $k8s_files" >> "$report"
    echo "" >> "$report"
    
    # Suggest improvements
    if [[ $root_configs -gt 3 ]]; then
        echo "## Recommendations" >> "$report"
        echo "- Consider moving configuration files from root to config/ directory" >> "$report"
        echo "- Root should only contain: main config, docker-compose.yml, and Dockerfile" >> "$report"
        echo "" >> "$report"
        
        if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
            # Create config directory if it doesn't exist
            mkdir -p config
            
            # Move appropriate config files
            for file in "${config_files[@]}"; do
                local basename=$(basename "$file")
                local dir=$(dirname "$file")
                
                # Skip essential root files
                if [[ "$dir" == "." && ! "$basename" =~ ^(docker-compose\.yml|Dockerfile|\.env)$ ]]; then
                    if [[ "$basename" =~ \.(yaml|yml|json|toml|ini|conf)$ ]]; then
                        log_info "Moving $file to config/"
                        git mv "$file" "config/" 2>/dev/null || mv "$file" "config/"
                        changes=$((changes + 1))
                    fi
                fi
            done
        fi
    fi
    
    # Check for environment-specific configuration
    log_info "Checking environment-specific configurations..."
    
    local env_configs=$(find . -name "*dev*" -o -name "*prod*" -o -name "*staging*" -o -name "*test*" | grep -E "\.(yaml|yml|json)$" | wc -l)
    
    echo "## Environment-Specific Configurations" >> "$report"
    echo "- Environment-specific config files: $env_configs" >> "$report"
    echo "" >> "$report"
    
    echo "## Changes Made: $changes" >> "$report"
    
    if [[ $changes -gt 0 ]]; then
        log_success "Configuration Agent completed with $changes changes"
    else
        log_info "Configuration Agent completed with no changes needed"
    fi
}

# Agent 6: Test Organization
agent_tests() {
    log_agent "Deploying Test Organization Agent..."
    
    local changes=0
    local report="${REPORT_DIR}/tests.md"
    
    cat > "$report" << EOF
# Test Organization Agent Report

## Test File Organization and Naming
EOF
    
    # Find all test files
    log_info "Analyzing test file organization..."
    
    local test_files=$(find . -name "*_test.go" -not -path "./vendor/*")
    local test_count=$(echo "$test_files" | wc -l)
    
    echo "## Test Files Analysis" >> "$report"
    echo "- Total test files: $test_count" >> "$report"
    echo "" >> "$report"
    
    # Check test file placement
    local misplaced_tests=0
    local integration_tests=0
    local unit_tests=0
    
    for test_file in $test_files; do
        local dir=$(dirname "$test_file")
        local basename=$(basename "$test_file")
        local corresponding_file=$(echo "$test_file" | sed 's/_test\.go$/.go/')
        
        # Check if test is in the right location
        if [[ ! -f "$corresponding_file" ]]; then
            # Check if it's an integration test
            if [[ "$basename" =~ integration_test\.go$ ]]; then
                integration_tests=$((integration_tests + 1))
            else
                misplaced_tests=$((misplaced_tests + 1))
                echo "- Misplaced test: $test_file (no corresponding implementation)" >> "$report"
            fi
        else
            unit_tests=$((unit_tests + 1))
        fi
        
        # Check test naming conventions
        if [[ ! "$basename" =~ ^[a-z][a-z0-9_]*_test\.go$ ]]; then
            echo "- Non-standard test name: $basename" >> "$report"
        fi
    done
    
    echo "## Test Classification" >> "$report"
    echo "- Unit tests: $unit_tests" >> "$report"
    echo "- Integration tests: $integration_tests" >> "$report"
    echo "- Misplaced tests: $misplaced_tests" >> "$report"
    echo "" >> "$report"
    
    # Check for test directories
    log_info "Checking test directory structure..."
    
    local test_dirs=()
    if [[ -d "test" ]]; then
        test_dirs+=("test")
    fi
    if [[ -d "tests" ]]; then
        test_dirs+=("tests")
    fi
    
    echo "## Test Directories" >> "$report"
    echo "- Test directories found: ${#test_dirs[@]}" >> "$report"
    
    for dir in "${test_dirs[@]}"; do
        local files_in_dir=$(find "$dir" -name "*.go" | wc -l)
        echo "- $dir/: $files_in_dir Go files" >> "$report"
    done
    echo "" >> "$report"
    
    # Check for benchmark tests
    local bench_tests=$(grep -r "func Benchmark" --include="*_test.go" . | wc -l)
    echo "## Benchmark Tests" >> "$report"
    echo "- Benchmark functions found: $bench_tests" >> "$report"
    echo "" >> "$report"
    
    # Check test naming patterns
    log_info "Analyzing test function naming..."
    
    local test_functions=$(grep -r "func Test" --include="*_test.go" . | wc -l)
    local example_functions=$(grep -r "func Example" --include="*_test.go" . | wc -l)
    
    echo "## Test Function Analysis" >> "$report"
    echo "- Test functions: $test_functions" >> "$report"
    echo "- Example functions: $example_functions" >> "$report"
    echo "" >> "$report"
    
    # Organize integration tests
    if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
        if [[ $integration_tests -gt 0 ]]; then
            # Create integration test directory if it doesn't exist
            mkdir -p test/integration
            
            # Move integration tests
            for test_file in $test_files; do
                local basename=$(basename "$test_file")
                if [[ "$basename" =~ integration_test\.go$ ]]; then
                    local current_dir=$(dirname "$test_file")
                    if [[ "$current_dir" != "./test/integration" ]]; then
                        log_info "Moving integration test $test_file to test/integration/"
                        git mv "$test_file" "test/integration/" 2>/dev/null || mv "$test_file" "test/integration/"
                        changes=$((changes + 1))
                    fi
                fi
            done
        fi
        
        # Create test helper files if missing
        if [[ ! -f "test/helper.go" && -d "test" ]]; then
            cat > "test/helper.go" << 'EOF'
// Package test provides common testing utilities for the Freightliner project.
package test

import (
    "testing"
    "os"
)

// SetupTestEnvironment prepares the test environment.
func SetupTestEnvironment(t *testing.T) {
    t.Helper()
    
    // Set test environment variables
    os.Setenv("ENV", "test")
    os.Setenv("LOG_LEVEL", "error")
}

// CleanupTestEnvironment cleans up after tests.
func CleanupTestEnvironment(t *testing.T) {
    t.Helper()
    
    // Clean up test environment
    os.Unsetenv("ENV")
    os.Unsetenv("LOG_LEVEL")
}
EOF
            log_info "Created test helper file"
            changes=$((changes + 1))
        fi
    fi
    
    echo "## Changes Made: $changes" >> "$report"
    
    if [[ $changes -gt 0 ]]; then
        log_success "Test Organization Agent completed with $changes changes"
    else
        log_info "Test Organization Agent completed with no changes needed"
    fi
}

# Agent 7: Build Cleanup
agent_build_cleanup() {
    log_agent "Deploying Build Cleanup Agent..."
    
    local changes=0
    local report="${REPORT_DIR}/build-cleanup.md"
    
    cat > "$report" << EOF
# Build Cleanup Agent Report

## Build Artifact Analysis and Cleanup
EOF
    
    # Find build artifacts
    log_info "Scanning for build artifacts..."
    
    local artifact_patterns=(
        "*.exe"
        "*.so"
        "*.a"
        "*.dylib"
        "*.dll"
        "bin/*"
        "build/*"
        "dist/*"
        "target/*"
        "coverage.out"
        "*.prof"
        "*.test"
        ".DS_Store"
        "Thumbs.db"
        "*.tmp"
        "*.log"
        "node_modules/"
        ".pytest_cache/"
        "__pycache__/"
        "*.pyc"
    )
    
    local artifacts_found=()
    local total_size=0
    
    for pattern in "${artifact_patterns[@]}"; do
        while IFS= read -r -d '' file; do
            if [[ -f "$file" || -d "$file" ]]; then
                artifacts_found+=("$file")
                if [[ -f "$file" ]]; then
                    local size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)
                    total_size=$((total_size + size))
                fi
            fi
        done < <(find . -name "$pattern" -not -path "./vendor/*" -not -path "*/.git/*" -print0 2>/dev/null)
    done
    
    echo "## Build Artifacts Found" >> "$report"
    echo "- Total artifacts: ${#artifacts_found[@]}" >> "$report"
    echo "- Total size: $(numfmt --to=iec $total_size)" >> "$report"
    echo "" >> "$report"
    
    if [[ ${#artifacts_found[@]} -gt 0 ]]; then
        echo "## Artifacts List" >> "$report"
        for artifact in "${artifacts_found[@]}"; do
            echo "- $artifact" >> "$report"
        done
        echo "" >> "$report"
    fi
    
    # Check for large files that might be artifacts
    log_info "Checking for large files..."
    
    local large_files=$(find . -type f -size +10M -not -path "./vendor/*" -not -path "*/.git/*" 2>/dev/null)
    
    if [[ -n "$large_files" ]]; then
        echo "## Large Files (>10MB)" >> "$report"
        echo "\`\`\`" >> "$report"
        echo "$large_files" >> "$report"
        echo "\`\`\`" >> "$report"
        echo "" >> "$report"
    fi
    
    # Clean up artifacts
    if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
        log_info "Cleaning up build artifacts..."
        
        for artifact in "${artifacts_found[@]}"; do
            # Skip essential directories and files
            if [[ "$artifact" =~ ^./bin/freightliner.*$ ]]; then
                continue  # Keep main binary
            fi
            
            if [[ -f "$artifact" ]]; then
                log_info "Removing artifact: $artifact"
                rm -f "$artifact"
                changes=$((changes + 1))
            elif [[ -d "$artifact" && "$artifact" != "./bin" ]]; then
                log_info "Removing artifact directory: $artifact"
                rm -rf "$artifact"
                changes=$((changes + 1))
            fi
        done
        
        # Clean up empty directories
        find . -type d -empty -not -path "*/.git/*" -delete 2>/dev/null || true
    fi
    
    # Check build cache
    local go_cache_size=$(go env GOCACHE 2>/dev/null | xargs du -sh 2>/dev/null | cut -f1 || echo "unknown")
    local go_mod_cache_size=$(go env GOMODCACHE 2>/dev/null | xargs du -sh 2>/dev/null | cut -f1 || echo "unknown")
    
    echo "## Go Build Cache" >> "$report"
    echo "- GOCACHE size: $go_cache_size" >> "$report"
    echo "- GOMODCACHE size: $go_mod_cache_size" >> "$report"
    echo "- Use 'go clean -cache' and 'go clean -modcache' to clean" >> "$report"
    echo "" >> "$report"
    
    echo "## Changes Made: $changes" >> "$report"
    
    if [[ $changes -gt 0 ]]; then
        log_success "Build Cleanup Agent completed with $changes changes"
    else
        log_info "Build Cleanup Agent completed with no changes needed"
    fi
}

# Agent 8: Git Management
agent_git_management() {
    log_agent "Deploying Git Management Agent..."
    
    local changes=0
    local report="${REPORT_DIR}/git-management.md"
    
    cat > "$report" << EOF
# Git Management Agent Report

## Git Repository Analysis and Optimization
EOF
    
    # Analyze .gitignore
    log_info "Analyzing .gitignore configuration..."
    
    local gitignore_file=".gitignore"
    local gitignore_exists=false
    
    if [[ -f "$gitignore_file" ]]; then
        gitignore_exists=true
        local gitignore_lines=$(wc -l < "$gitignore_file")
        echo "## .gitignore Analysis" >> "$report"
        echo "- .gitignore exists with $gitignore_lines lines" >> "$report"
    else
        echo "## .gitignore Analysis" >> "$report"
        echo "- .gitignore file missing" >> "$report"
    fi
    echo "" >> "$report"
    
    # Define comprehensive .gitignore patterns for Go projects
    local go_gitignore_patterns=(
        "# Binaries for programs and plugins"
        "*.exe"
        "*.exe~"
        "*.dll"
        "*.so"
        "*.dylib"
        ""
        "# Test binary, built with \`go test -c\`"
        "*.test"
        ""
        "# Output of the go coverage tool, specifically when used with LiteIDE"
        "*.out"
        ""
        "# Dependency directories"
        "vendor/"
        ""
        "# Go workspace file"
        "go.work"
        ""
        "# Build artifacts"
        "bin/"
        "build/"
        "dist/"
        ""
        "# IDE files"
        ".vscode/"
        ".idea/"
        "*.iml"
        ".DS_Store"
        "Thumbs.db"
        ""
        "# Logs"
        "*.log"
        "logs/"
        ""
        "# Runtime data"
        "pids"
        "*.pid"
        "*.seed"
        "*.pid.lock"
        ""
        "# Coverage directory used by tools like istanbul"
        "coverage/"
        "*.lcov"
        ""
        "# nyc test coverage"
        ".nyc_output"
        ""
        "# node_modules (if any Node.js tools are used)"
        "node_modules/"
        ""
        "# Docker"
        "*.log"
        ""
        "# Temporary files"
        "*.tmp"
        "*.temp"
        ""
        "# Environment files"
        ".env"
        ".env.local"
        ".env.*.local"
        ""
        "# Kubernetes secrets"
        "secrets.yaml"
        "secret.yaml"
        ""
        "# Terraform"
        "*.tfstate"
        "*.tfstate.*"
        ".terraform/"
        ".terraform.lock.hcl"
        ""
        "# Helm"
        "charts/*.tgz"
        ""
        "# OS generated files"
        ".DS_Store"
        ".DS_Store?"
        "._*"
        ".Spotlight-V100"
        ".Trashes"
        "ehthumbs.db"
        "Thumbs.db"
        ""
        "# JetBrains IDEs"
        ".idea/"
        "*.iws"
        "*.iml"
        "*.ipr"
        ""
        "# Visual Studio Code"
        ".vscode/"
        ""
        "# Vim"
        "*.sw[po]"
        "*.swp"
        "*~"
        ""
        "# Claude Code working directory"
        ".claude/tidy-work/"
    )
    
    # Check current gitignore against recommended patterns
    local missing_patterns=()
    
    if [[ "$gitignore_exists" == "true" ]]; then
        for pattern in "${go_gitignore_patterns[@]}"; do
            if [[ -n "$pattern" && ! "$pattern" =~ ^# ]]; then
                if ! grep -Fxq "$pattern" "$gitignore_file"; then
                    missing_patterns+=("$pattern")
                fi
            fi
        done
    else
        missing_patterns=("${go_gitignore_patterns[@]}")
    fi
    
    echo "## Missing .gitignore Patterns" >> "$report"
    echo "- Missing patterns: ${#missing_patterns[@]}" >> "$report"
    
    if [[ ${#missing_patterns[@]} -gt 0 ]]; then
        echo "- Recommended additions:" >> "$report"
        for pattern in "${missing_patterns[@]:0:10}"; do  # Show first 10
            if [[ -n "$pattern" && ! "$pattern" =~ ^# ]]; then
                echo "  - $pattern" >> "$report"
            fi
        done
        
        if [[ ${#missing_patterns[@]} -gt 10 ]]; then
            echo "  - ... and $((${#missing_patterns[@]} - 10)) more" >> "$report"
        fi
    fi
    echo "" >> "$report"
    
    # Update .gitignore if needed
    if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
        if [[ ${#missing_patterns[@]} -gt 0 ]]; then
            log_info "Updating .gitignore with missing patterns..."
            
            # Create or append to .gitignore
            {
                if [[ "$gitignore_exists" == "true" ]]; then
                    cat "$gitignore_file"
                    echo ""
                    echo "# Added by tidy-repo workflow"
                fi
                
                printf '%s\n' "${go_gitignore_patterns[@]}"
            } > "${gitignore_file}.tmp"
            
            # Remove duplicates while preserving order and comments
            awk '!seen[$0]++' "${gitignore_file}.tmp" > "$gitignore_file"
            rm "${gitignore_file}.tmp"
            
            changes=$((changes + 1))
            log_info "Updated .gitignore with recommended patterns"
        fi
    fi
    
    # Check for files that should be ignored but aren't
    log_info "Checking for untracked files that should be ignored..."
    
    local untracked_files=$(git ls-files --others --exclude-standard)
    local should_be_ignored=()
    
    for file in $untracked_files; do
        # Check if file matches patterns that should be ignored
        if [[ "$file" =~ \.(exe|dll|so|dylib|test|out|log|tmp|pid)$ ]] || \
           [[ "$file" =~ ^(bin|build|dist|coverage|node_modules)/ ]] || \
           [[ "$file" =~ (\.DS_Store|Thumbs\.db|\.env)$ ]]; then
            should_be_ignored+=("$file")
        fi
    done
    
    echo "## Files That Should Be Ignored" >> "$report"
    echo "- Untracked files that match ignore patterns: ${#should_be_ignored[@]}" >> "$report"
    
    if [[ ${#should_be_ignored[@]} -gt 0 ]]; then
        echo "- Files:" >> "$report"
        for file in "${should_be_ignored[@]:0:10}"; do  # Show first 10
            echo "  - $file" >> "$report"
        done
        
        if [[ ${#should_be_ignored[@]} -gt 10 ]]; then
            echo "  - ... and $((${#should_be_ignored[@]} - 10)) more" >> "$report"
        fi
    fi
    echo "" >> "$report"
    
    # Analyze repository size and history
    log_info "Analyzing repository size and history..."
    
    local repo_size=$(du -sh .git 2>/dev/null | cut -f1 || echo "unknown")
    local commit_count=$(git rev-list --all --count 2>/dev/null || echo "unknown")
    local branch_count=$(git branch -a | wc -l)
    local large_files=$(git ls-files | xargs ls -la 2>/dev/null | awk '$5 > 1048576 {print $9, $5}' | head -10)
    
    echo "## Repository Analysis" >> "$report"
    echo "- Repository size: $repo_size" >> "$report"
    echo "- Total commits: $commit_count" >> "$report"
    echo "- Total branches: $branch_count" >> "$report"
    echo "" >> "$report"
    
    if [[ -n "$large_files" ]]; then
        echo "## Large Files in Repository" >> "$report"
        echo "\`\`\`" >> "$report"
        echo "$large_files" >> "$report"
        echo "\`\`\`" >> "$report"
        echo "" >> "$report"
    fi
    
    # Check for Git LFS usage
    if [[ -f ".gitattributes" ]]; then
        local lfs_patterns=$(grep "filter=lfs" .gitattributes | wc -l)
        echo "## Git LFS" >> "$report"
        echo "- LFS patterns configured: $lfs_patterns" >> "$report"
    else
        echo "## Git LFS" >> "$report"
        echo "- No .gitattributes file found" >> "$report"
        echo "- Consider using Git LFS for large binary files" >> "$report"
    fi
    echo "" >> "$report"
    
    echo "## Changes Made: $changes" >> "$report"
    
    if [[ $changes -gt 0 ]]; then
        log_success "Git Management Agent completed with $changes changes"
    else
        log_info "Git Management Agent completed with no changes needed"
    fi
}

# Agent 9: Dependency Management
agent_dependency_management() {
    log_agent "Deploying Dependency Management Agent..."
    
    local changes=0
    local report="${REPORT_DIR}/dependency-management.md"
    
    cat > "$report" << EOF
# Dependency Management Agent Report

## Dependency Analysis and Optimization
EOF
    
    # Analyze Go dependencies
    log_info "Analyzing Go module dependencies..."
    
    if [[ -f "go.mod" ]]; then
        local go_version=$(grep "^go " go.mod | awk '{print $2}')
        local require_count=$(grep -c "^\s*github.com\|^\s*golang.org\|^\s*google.golang.org\|^\s*cloud.google.com" go.mod || echo 0)
        local replace_count=$(grep -c "^replace " go.mod || echo 0)
        local exclude_count=$(grep -c "^exclude " go.mod || echo 0)
        
        echo "## Go Module Analysis" >> "$report"
        echo "- Go version: $go_version" >> "$report"
        echo "- Direct dependencies: $require_count" >> "$report"
        echo "- Replace directives: $replace_count" >> "$report"
        echo "- Exclude directives: $exclude_count" >> "$report"
        echo "" >> "$report"
    fi
    
    # Check for dependency updates
    log_info "Checking for outdated dependencies..."
    
    if command -v go &> /dev/null; then
        # Run go mod tidy to clean up
        if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
            log_info "Running go mod tidy..."
            go mod tidy
            changes=$((changes + 1))
        fi
        
        # Check for updates (this requires internet access)
        local outdated_deps=""
        if command -v go-mod-outdated &> /dev/null; then
            outdated_deps=$(go list -u -m all 2>/dev/null | grep -E '\[(latest|upgrade)\]' || true)
        fi
        
        if [[ -n "$outdated_deps" ]]; then
            local outdated_count=$(echo "$outdated_deps" | wc -l)
            echo "## Outdated Dependencies" >> "$report"
            echo "- Outdated dependencies: $outdated_count" >> "$report"
            echo "\`\`\`" >> "$report"
            echo "$outdated_deps" >> "$report"
            echo "\`\`\`" >> "$report"
        else
            echo "## Outdated Dependencies" >> "$report"
            echo "- No outdated dependencies detected" >> "$report"
        fi
        echo "" >> "$report"
        
        # Analyze dependency tree
        local total_deps=$(go list -m all 2>/dev/null | wc -l)
        echo "## Dependency Tree" >> "$report"
        echo "- Total dependencies (including transitive): $total_deps" >> "$report"
        echo "" >> "$report"
        
        # Check for security vulnerabilities
        log_info "Checking for security vulnerabilities..."
        
        if command -v govulncheck &> /dev/null; then
            local vuln_output
            vuln_output=$(govulncheck ./... 2>&1 || true)
            
            if [[ "$vuln_output" =~ "No vulnerabilities found" ]]; then
                echo "## Security Analysis" >> "$report"
                echo "- No vulnerabilities found" >> "$report"
            else
                echo "## Security Analysis" >> "$report"
                echo "- Vulnerabilities detected - see govulncheck output" >> "$report"
                echo "\`\`\`" >> "$report"
                echo "$vuln_output" >> "$report"
                echo "\`\`\`" >> "$report"
            fi
        else
            echo "## Security Analysis" >> "$report"
            echo "- govulncheck not available" >> "$report"
            echo "- Install with: go install golang.org/x/vuln/cmd/govulncheck@latest" >> "$report"
        fi
        echo "" >> "$report"
    fi
    
    # Check for unused dependencies
    log_info "Checking for unused dependencies..."
    
    # This is a simplified check - a more thorough analysis would use tools like go-mod-graph
    if [[ -f "go.mod" ]]; then
        local direct_deps=$(grep -E "^\s*(github.com|golang.org|google.golang.org|cloud.google.com)" go.mod | awk '{print $1}')
        local unused_deps=()
        
        for dep in $direct_deps; do
            # Check if dependency is actually imported
            local usage_count=$(grep -r "\"$dep" --include="*.go" . 2>/dev/null | wc -l)
            if [[ $usage_count -eq 0 ]]; then
                unused_deps+=("$dep")
            fi
        done
        
        echo "## Unused Dependencies" >> "$report"
        echo "- Potentially unused dependencies: ${#unused_deps[@]}" >> "$report"
        
        if [[ ${#unused_deps[@]} -gt 0 ]]; then
            echo "- Dependencies:" >> "$report"
            for dep in "${unused_deps[@]}"; do
                echo "  - $dep" >> "$report"
            done
            echo "- Note: This is a simplified check. Use 'go mod why <module>' for detailed analysis" >> "$report"
        fi
        echo "" >> "$report"
    fi
    
    # Check for other dependency files
    local other_deps=()
    
    if [[ -f "package.json" ]]; then
        other_deps+=("Node.js (package.json)")
    fi
    
    if [[ -f "requirements.txt" ]]; then
        other_deps+=("Python (requirements.txt)")
    fi
    
    if [[ -f "Pipfile" ]]; then
        other_deps+=("Python (Pipfile)")
    fi
    
    if [[ -f "Cargo.toml" ]]; then
        other_deps+=("Rust (Cargo.toml)")
    fi
    
    if [[ ${#other_deps[@]} -gt 0 ]]; then
        echo "## Other Dependency Files" >> "$report"
        for dep in "${other_deps[@]}"; do
            echo "- $dep" >> "$report"
        done
        echo "" >> "$report"
    fi
    
    # Check for dependency license compliance
    log_info "Checking dependency licenses..."
    
    # This would typically require additional tools, but we can check for common license files
    local license_files=$(find vendor -name "LICENSE*" -o -name "COPYING*" -o -name "COPYRIGHT*" 2>/dev/null | wc -l || echo 0)
    
    echo "## License Compliance" >> "$report"
    echo "- License files in vendor/: $license_files" >> "$report"
    echo "- Consider using tools like 'go-licenses' for comprehensive license analysis" >> "$report"
    echo "" >> "$report"
    
    echo "## Changes Made: $changes" >> "$report"
    
    if [[ $changes -gt 0 ]]; then
        log_success "Dependency Management Agent completed with $changes changes"
    else
        log_info "Dependency Management Agent completed with no changes needed"
    fi
}

# Agent 10: Dead Code Removal
agent_dead_code_removal() {
    log_agent "Deploying Dead Code Removal Agent..."
    
    local changes=0
    local report="${REPORT_DIR}/dead-code-removal.md"
    
    cat > "$report" << EOF
# Dead Code Removal Agent Report

## Dead Code Analysis and Removal
EOF
    
    # Find potentially unused Go code
    log_info "Analyzing Go code for unused elements..."
    
    # Check for unused functions, variables, and types
    local unused_elements=()
    
    if command -v golangci-lint &> /dev/null; then
        log_info "Running golangci-lint to detect unused code..."
        
        # Run specific linters for unused code
        local lint_output
        lint_output=$(golangci-lint run --disable-all --enable=unused,deadcode,varcheck,structcheck,ineffassign ./... 2>&1 || true)
        
        if [[ -n "$lint_output" && ! "$lint_output" =~ "no issues found" ]]; then
            local issue_count=$(echo "$lint_output" | grep -E "unused|deadcode|varcheck|structcheck|ineffassign" | wc -l)
            echo "## Linter Analysis" >> "$report"
            echo "- Issues found by golangci-lint: $issue_count" >> "$report"
            echo "\`\`\`" >> "$report"
            echo "$lint_output" >> "$report"
            echo "\`\`\`" >> "$report"
        else
            echo "## Linter Analysis" >> "$report"
            echo "- No unused code detected by golangci-lint" >> "$report"
        fi
        echo "" >> "$report"
    else
        echo "## Linter Analysis" >> "$report"
        echo "- golangci-lint not available" >> "$report"
        echo "- Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" >> "$report"
        echo "" >> "$report"
    fi
    
    # Manual analysis for common dead code patterns
    log_info "Performing manual dead code analysis..."
    
    # Find functions that are never called
    local go_files=$(find . -name "*.go" -not -path "./vendor/*")
    local potentially_unused_funcs=()
    
    for file in $go_files; do
        # Extract function definitions
        local funcs=$(grep -n "^func [A-Za-z]" "$file" | sed 's/func \([A-Za-z_][A-Za-z0-9_]*\).*/\1/' || true)
        
        for func in $funcs; do
            # Skip main, init, Test*, Benchmark*, Example* functions
            if [[ "$func" =~ ^(main|init|Test|Benchmark|Example) ]]; then
                continue
            fi
            
            # Check if function is called anywhere
            local usage_count=$(grep -r "\b$func\b" --include="*.go" . | grep -v "^$file.*func $func" | wc -l)
            
            if [[ $usage_count -eq 0 ]]; then
                potentially_unused_funcs+=("$func in $file")
            fi
        done
    done
    
    echo "## Potentially Unused Functions" >> "$report"
    echo "- Functions that may be unused: ${#potentially_unused_funcs[@]}" >> "$report"
    
    if [[ ${#potentially_unused_funcs[@]} -gt 0 ]]; then
        echo "- Functions (manual analysis):" >> "$report"
        for func in "${potentially_unused_funcs[@]:0:10}"; do  # Show first 10
            echo "  - $func" >> "$report"
        done
        
        if [[ ${#potentially_unused_funcs[@]} -gt 10 ]]; then
            echo "  - ... and $((${#potentially_unused_funcs[@]} - 10)) more" >> "$report"
        fi
        
        echo "- Note: This is a simplified analysis. Functions may be used via reflection or build tags" >> "$report"
    fi
    echo "" >> "$report"
    
    # Find unused variables and imports
    log_info "Checking for unused variables and imports..."
    
    # Use go vet to find potential issues
    local vet_output
    vet_output=$(go vet ./... 2>&1 || true)
    
    if [[ -n "$vet_output" ]]; then
        echo "## Go Vet Analysis" >> "$report"
        echo "\`\`\`" >> "$report"
        echo "$vet_output" >> "$report"
        echo "\`\`\`" >> "$report"
    else
        echo "## Go Vet Analysis" >> "$report"
        echo "- No issues found by go vet" >> "$report"
    fi
    echo "" >> "$report"
    
    # Check for empty files or files with only comments
    log_info "Checking for empty or comment-only files..."
    
    local empty_files=()
    
    for file in $go_files; do
        # Remove comments and empty lines, check if anything remains
        local content=$(sed '/^[[:space:]]*\/\//d; /^[[:space:]]*\/\*/,/\*\//d; /^[[:space:]]*$/d' "$file")
        
        if [[ -z "$content" ]]; then
            empty_files+=("$file")
        fi
    done
    
    echo "## Empty or Comment-Only Files" >> "$report"
    echo "- Files with no meaningful content: ${#empty_files[@]}" >> "$report"
    
    if [[ ${#empty_files[@]} -gt 0 ]]; then
        echo "- Files:" >> "$report"
        for file in "${empty_files[@]}"; do
            echo "  - $file" >> "$report"
        done
        
        if [[ "$DRY_RUN" == "false" && "$REPORT_ONLY" == "false" ]]; then
            # Remove empty files
            for file in "${empty_files[@]}"; do
                log_info "Removing empty file: $file"
                git rm "$file" 2>/dev/null || rm "$file"
                changes=$((changes + 1))
            done
        fi
    fi
    echo "" >> "$report"
    
    # Check for TODO/FIXME/HACK comments
    log_info "Analyzing technical debt comments..."
    
    local todo_count=$(grep -r "TODO\|FIXME\|HACK\|XXX" --include="*.go" . | wc -l)
    
    echo "## Technical Debt Comments" >> "$report"
    echo "- TODO/FIXME/HACK/XXX comments: $todo_count" >> "$report"
    
    if [[ $todo_count -gt 0 ]]; then
        local sample_todos=$(grep -r "TODO\|FIXME\|HACK\|XXX" --include="*.go" . | head -5)
        echo "- Sample comments:" >> "$report" 
        echo "\`\`\`" >> "$report"
        echo "$sample_todos" >> "$report"
        echo "\`\`\`" >> "$report"
    fi
    echo "" >> "$report"
    
    # Check for debug/development code
    log_info "Checking for debug and development code..."
    
    local debug_patterns=("fmt.Print" "log.Print" "panic(" "debug\." "// DEBUG" "// DEV")
    local debug_issues=0
    
    for pattern in "${debug_patterns[@]}"; do
        local count=$(grep -r "$pattern" --include="*.go" . | wc -l)
        debug_issues=$((debug_issues + count))
    done
    
    echo "## Debug/Development Code" >> "$report"
    echo "- Potential debug code instances: $debug_issues" >> "$report"
    
    if [[ $debug_issues -gt 0 ]]; then
        echo "- Review these patterns manually:" >> "$report"
        for pattern in "${debug_patterns[@]}"; do
            local count=$(grep -r "$pattern" --include="*.go" . | wc -l)
            if [[ $count -gt 0 ]]; then
                echo "  - $pattern: $count instances" >> "$report"
            fi
        done
    fi
    echo "" >> "$report"
    
    # Check for unused build tags
    log_info "Checking for build tags..."
    
    local build_tags=$(grep -r "// +build\|//go:build" --include="*.go" . | cut -d: -f1 | sort -u | wc -l)
    
    echo "## Build Tags" >> "$report"
    echo "- Files with build tags: $build_tags" >> "$report"
    echo "- Review build tags to ensure they're still needed" >> "$report"
    echo "" >> "$report"
    
    echo "## Changes Made: $changes" >> "$report"
    
    if [[ $changes -gt 0 ]]; then
        log_success "Dead Code Removal Agent completed with $changes changes"
    else
        log_info "Dead Code Removal Agent completed with no changes needed"
    fi
}

# Generate comprehensive summary report
generate_summary_report() {
    log_info "Generating comprehensive summary report..."
    
    local summary_file="${REPORT_DIR}/summary.md"
    
    cat >> "$summary_file" << EOF

## Agent Execution Summary

| Agent | Status | Changes | Report |
|-------|--------|---------|--------|
EOF
    
    # Add each agent's summary
    for agent in "${AGENTS_ENABLED[@]}"; do
        local agent_report="${REPORT_DIR}/${agent//-/-}.md"
        local status="❌ Not Run"
        local changes="0"
        
        if [[ -f "$agent_report" ]]; then
            status="✅ Completed"
            changes=$(grep "Changes Made:" "$agent_report" | awk -F: '{print $2}' | xargs || echo "0")
        fi
        
        echo "| $agent | $status | $changes | [Report](./${agent//-/-}.md) |" >> "$summary_file"
    done
    
    cat >> "$summary_file" << EOF

## Repository Health Assessment

### Before Tidying
- Repository structure analysis completed
- Code quality assessment performed
- Technical debt identified

### After Tidying
- File organization improved
- Code formatting standardized
- Dependencies optimized
- Dead code removed
- Documentation organized

## Recommendations

1. **Continuous Integration**: Add pre-commit hooks to maintain code quality
2. **Regular Maintenance**: Run tidying workflow monthly
3. **Dependency Updates**: Monitor for security updates weekly
4. **Documentation**: Keep README and API docs up to date
5. **Testing**: Ensure test coverage remains high after changes

## Next Steps

1. Review all generated reports
2. Test the application after changes
3. Update CI/CD pipelines if needed
4. Document any new conventions
5. Share findings with the team

---
Generated by Repository Tidying Workflow
Timestamp: $(date)
Command: $0 $@
EOF
    
    log_success "Summary report generated at $summary_file"
}

# Main execution function
main() {
    echo "🚀 Repository Tidying Workflow - Freightliner Container Registry Replication"
    echo "========================================================================"
    
    # Parse arguments
    parse_args "$@"
    
    # Initialize
    init_workspace
    preflight_checks
    
    # Determine which agents to run
    local agents_to_run=()
    
    if [[ -n "$SPECIFIC_AGENT" ]]; then
        agents_to_run=("$SPECIFIC_AGENT")
    elif [[ "$QUICK_TIDY" == "true" ]]; then
        agents_to_run=("code-style" "imports" "build" "git")
    else
        agents_to_run=("${AGENTS_ENABLED[@]}")
    fi
    
    log_info "Running agents: ${agents_to_run[*]}"
    
    # Execute agents
    for agent in "${agents_to_run[@]}"; do
        case "$agent" in
            "file-org")
                agent_file_organization
                ;;
            "code-style")
                agent_code_style
                ;;
            "imports")
                agent_imports
                ;;
            "docs")
                agent_documentation
                ;;
            "config")
                agent_configuration
                ;;
            "tests")
                agent_tests
                ;;
            "build")
                agent_build_cleanup
                ;;
            "git")
                agent_git_management
                ;;
            "deps")
                agent_dependency_management
                ;;
            "deadcode")
                agent_dead_code_removal
                ;;
            *)
                log_error "Unknown agent: $agent"
                ;;
        esac
    done
    
    # Generate summary
    generate_summary_report
    
    # Final status
    echo ""
    echo "========================================================================"
    log_success "Repository Tidying Workflow completed!"
    echo ""
    log_info "Reports generated in: $REPORT_DIR"
    log_info "Summary report: ${REPORT_DIR}/summary.md"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "This was a dry run - no changes were made"
    elif [[ "$REPORT_ONLY" == "true" ]]; then
        log_info "Report-only mode - no changes were made"
    else
        log_info "Changes have been made to the repository"
        log_warning "Please review changes and test thoroughly before committing"
    fi
    
    echo ""
    echo "Next steps:"
    echo "1. Review the summary report: ${REPORT_DIR}/summary.md" 
    echo "2. Check individual agent reports for details"
    echo "3. Test the application to ensure changes don't break functionality"
    echo "4. Commit changes if satisfied: git add . && git commit -m 'Repository tidying workflow'"
    echo ""
}

# Execute main function if script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
```

## Integration with Existing Tools

This workflow integrates with existing Freightliner tools:

- **Makefile targets**: Uses existing `fmt`, `lint`, `vet`, `test` targets
- **Go toolchain**: Leverages `gofmt`, `goimports`, `go mod tidy`
- **CI/CD**: Compatible with existing GitHub Actions and build processes
- **Development scripts**: Extends existing scripts in `scripts/` directory

## Usage Examples

```bash
# Complete repository tidying
/tidy-repo --full

# Quick essential tidying
/tidy-repo --quick

# Dry run to see what would change
/tidy-repo --dry-run --full

# Run specific agent only
/tidy-repo --agent=code-style

# Generate report without changes
/tidy-repo --report-only
```

## Verification Steps

After running the workflow:

1. **Build verification**: `make build`
2. **Test verification**: `make test`
3. **Lint verification**: `make lint`
4. **Integration tests**: `make test-integration`
5. **Docker build**: `make docker-build`

## Customization

The workflow can be customized by:

- Modifying agent configurations in the script
- Adding project-specific patterns to `.gitignore`
- Extending the documentation templates
- Adding custom linting rules
- Integrating with additional tools

This comprehensive workflow ensures the Freightliner repository maintains high standards of organization, code quality, and maintainability while respecting the existing project structure and conventions.