#!/bin/bash
# Freightliner Secrets Manager
# Simplifies secret management for Kubernetes and local development

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="${FREIGHTLINER_NAMESPACE:-freightliner}"
SECRET_NAME="freightliner-secrets"
ENV_FILE=".env.secrets"

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# Help message
show_help() {
    cat << EOF
Freightliner Secrets Manager

USAGE:
    $(basename "$0") [COMMAND] [OPTIONS]

COMMANDS:
    setup               Interactive setup wizard for all secrets
    create              Create secrets in Kubernetes
    update              Update existing secrets
    view                View current secrets (redacted)
    delete              Delete secrets
    export              Export secrets to .env file
    import              Import secrets from .env file
    validate            Validate required secrets are present
    rotate              Rotate secrets (generate new values)
    help                Show this help message

OPTIONS:
    --namespace NAME    Kubernetes namespace (default: freightliner)
    --secret-name NAME  Secret resource name (default: freightliner-secrets)
    --env-file PATH     Environment file path (default: .env.secrets)
    --dry-run           Show what would be done without executing
    --force             Skip confirmation prompts
    --context CONTEXT   Kubernetes context to use

EXAMPLES:
    # Interactive setup (recommended for first-time setup)
    $(basename "$0") setup

    # Create secrets from environment variables
    export AWS_ACCESS_KEY_ID=your-key
    export AWS_SECRET_ACCESS_KEY=your-secret
    $(basename "$0") create

    # Import secrets from file
    $(basename "$0") import --env-file ~/.freightliner/secrets.env

    # View current secrets
    $(basename "$0") view

    # Rotate encryption keys
    $(basename "$0") rotate --type encryption

    # Validate all required secrets
    $(basename "$0") validate

ENVIRONMENT VARIABLES:
    AWS_ACCESS_KEY_ID              AWS access key for ECR
    AWS_SECRET_ACCESS_KEY          AWS secret key for ECR
    GCP_SERVICE_ACCOUNT_KEY        GCP service account JSON (base64 or path)
    API_KEY                        Freightliner API key
    ENCRYPTION_KEY                 Data encryption key (32 bytes, base64)
    REGISTRY_AUTH_TOKEN            Container registry authentication token
    FREIGHTLINER_NAMESPACE         Kubernetes namespace

NOTES:
    - Secrets are never displayed in plain text (use --show-plaintext to override)
    - Always verify secrets after creation with: $(basename "$0") validate
    - Store .env.secrets files securely (add to .gitignore)
    - Use separate secrets for dev/staging/production

EOF
}

# Check if kubectl is available
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found. Please install kubectl first."
        return 1
    fi
}

# Check if namespace exists
check_namespace() {
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_warn "Namespace '$NAMESPACE' doesn't exist. Creating..."
        kubectl create namespace "$NAMESPACE"
        log_success "Namespace '$NAMESPACE' created"
    fi
}

# Generate secure random value
generate_random() {
    local length="${1:-32}"
    openssl rand -base64 "$length" | tr -d '\n'
}

# Prompt for secret with validation
prompt_secret() {
    local var_name="$1"
    local prompt_text="$2"
    local default_value="${3:-}"
    local allow_generate="${4:-false}"

    echo ""
    log_info "$prompt_text"

    if [ "$allow_generate" = "true" ]; then
        echo "  (Press Enter for random generation, or provide value)"
    fi

    if [ -n "$default_value" ]; then
        echo "  (Current: ${default_value:0:8}...)"
    fi

    read -rsp "  Value: " value
    echo ""

    if [ -z "$value" ] && [ "$allow_generate" = "true" ]; then
        value=$(generate_random 32)
        log_success "Generated secure random value"
    elif [ -z "$value" ] && [ -n "$default_value" ]; then
        value="$default_value"
        log_info "Using existing value"
    elif [ -z "$value" ]; then
        log_warn "Empty value provided"
        return 1
    fi

    # Store in variable
    eval "$var_name='$value'"
    return 0
}

# Interactive setup wizard
setup_wizard() {
    log_info "=== Freightliner Secrets Setup Wizard ==="
    echo ""

    # AWS ECR credentials
    log_info "📦 AWS ECR Credentials (for pulling/pushing container images)"
    prompt_secret AWS_ACCESS_KEY_ID "Enter AWS Access Key ID" "${AWS_ACCESS_KEY_ID:-}"
    prompt_secret AWS_SECRET_ACCESS_KEY "Enter AWS Secret Access Key" "${AWS_SECRET_ACCESS_KEY:-}"

    # GCP credentials
    echo ""
    log_info "☁️ GCP Credentials (for Google Container Registry)"
    echo "  Options:"
    echo "    1) Path to service account JSON file"
    echo "    2) Paste JSON content directly"
    echo "    3) Base64-encoded JSON"
    read -rp "  Choice (1/2/3): " gcp_choice

    case $gcp_choice in
        1)
            read -rp "  File path: " gcp_file
            if [ -f "$gcp_file" ]; then
                GCP_SERVICE_ACCOUNT_KEY=$(base64 < "$gcp_file" | tr -d '\n')
                log_success "Service account loaded from file"
            else
                log_error "File not found: $gcp_file"
                return 1
            fi
            ;;
        2)
            log_info "Paste JSON content (Ctrl+D when done):"
            gcp_json=$(cat)
            GCP_SERVICE_ACCOUNT_KEY=$(echo "$gcp_json" | base64 | tr -d '\n')
            log_success "Service account loaded from input"
            ;;
        3)
            prompt_secret GCP_SERVICE_ACCOUNT_KEY "Enter base64-encoded service account"
            ;;
        *)
            log_warn "Skipping GCP credentials"
            GCP_SERVICE_ACCOUNT_KEY=""
            ;;
    esac

    # API Key
    echo ""
    log_info "🔑 Freightliner API Key"
    prompt_secret API_KEY "Enter API key (or generate random)" "" true

    # Encryption Key
    echo ""
    log_info "🔒 Encryption Key (for data-at-rest encryption)"
    prompt_secret ENCRYPTION_KEY "Enter encryption key (or generate random)" "" true

    # Registry Auth Token
    echo ""
    log_info "🎫 Registry Authentication Token (optional)"
    read -rp "  Do you need a registry auth token? (y/N): " need_registry_auth
    if [[ "$need_registry_auth" =~ ^[Yy]$ ]]; then
        prompt_secret REGISTRY_AUTH_TOKEN "Enter registry auth token"
    else
        REGISTRY_AUTH_TOKEN=""
    fi

    # Summary
    echo ""
    log_info "=== Setup Summary ==="
    echo "  AWS Access Key ID: ${AWS_ACCESS_KEY_ID:+✓ Provided} ${AWS_ACCESS_KEY_ID:-✗ Missing}"
    echo "  AWS Secret Access Key: ${AWS_SECRET_ACCESS_KEY:+✓ Provided} ${AWS_SECRET_ACCESS_KEY:-✗ Missing}"
    echo "  GCP Service Account: ${GCP_SERVICE_ACCOUNT_KEY:+✓ Provided} ${GCP_SERVICE_ACCOUNT_KEY:-✗ Missing}"
    echo "  API Key: ${API_KEY:+✓ Provided} ${API_KEY:-✗ Missing}"
    echo "  Encryption Key: ${ENCRYPTION_KEY:+✓ Provided} ${ENCRYPTION_KEY:-✗ Missing}"
    echo "  Registry Auth Token: ${REGISTRY_AUTH_TOKEN:+✓ Provided} ${REGISTRY_AUTH_TOKEN:-○ Optional}"
    echo ""

    read -rp "Create secrets with these values? (y/N): " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        log_warn "Setup cancelled"
        return 1
    fi

    create_secrets
}

# Create secrets in Kubernetes
create_secrets() {
    check_kubectl
    check_namespace

    log_info "Creating secrets in namespace '$NAMESPACE'..."

    # Build kubectl command
    local cmd=(kubectl create secret generic "$SECRET_NAME" -n "$NAMESPACE")

    [ -n "${AWS_ACCESS_KEY_ID:-}" ] && cmd+=(--from-literal=aws-access-key-id="$AWS_ACCESS_KEY_ID")
    [ -n "${AWS_SECRET_ACCESS_KEY:-}" ] && cmd+=(--from-literal=aws-secret-access-key="$AWS_SECRET_ACCESS_KEY")
    [ -n "${GCP_SERVICE_ACCOUNT_KEY:-}" ] && cmd+=(--from-literal=gcp-service-account-key="$GCP_SERVICE_ACCOUNT_KEY")
    [ -n "${API_KEY:-}" ] && cmd+=(--from-literal=api-key="$API_KEY")
    [ -n "${ENCRYPTION_KEY:-}" ] && cmd+=(--from-literal=encryption-key="$ENCRYPTION_KEY")
    [ -n "${REGISTRY_AUTH_TOKEN:-}" ] && cmd+=(--from-literal=registry-auth="$REGISTRY_AUTH_TOKEN")

    # Check if secret already exists
    if kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" &> /dev/null; then
        log_warn "Secret '$SECRET_NAME' already exists"
        read -rp "Delete and recreate? (y/N): " confirm
        if [[ "$confirm" =~ ^[Yy]$ ]]; then
            kubectl delete secret "$SECRET_NAME" -n "$NAMESPACE"
            log_success "Existing secret deleted"
        else
            log_info "Use 'update' command to modify existing secrets"
            return 1
        fi
    fi

    # Create secret
    if "${cmd[@]}"; then
        log_success "Secrets created successfully"

        # Add labels
        kubectl label secret "$SECRET_NAME" -n "$NAMESPACE" \
            app.kubernetes.io/name=freightliner \
            app.kubernetes.io/component=secrets \
            --overwrite

        log_success "Labels applied"
    else
        log_error "Failed to create secrets"
        return 1
    fi
}

# View secrets (redacted)
view_secrets() {
    check_kubectl

    if ! kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" &> /dev/null; then
        log_error "Secret '$SECRET_NAME' not found in namespace '$NAMESPACE'"
        return 1
    fi

    log_info "Secrets in '$SECRET_NAME' (namespace: $NAMESPACE):"
    echo ""

    # Get secret keys and show redacted values
    kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" -o json | \
        jq -r '.data | keys[]' | while read -r key; do
            value=$(kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" -o jsonpath="{.data.$key}" | base64 -d)
            value_len=${#value}

            if [ "$value_len" -gt 0 ]; then
                echo "  $key: [REDACTED] ($value_len characters)"
            else
                echo "  $key: [EMPTY]"
            fi
        done

    echo ""
    log_info "To view plaintext values, use: kubectl get secret $SECRET_NAME -n $NAMESPACE -o yaml"
}

# Validate secrets
validate_secrets() {
    check_kubectl

    log_info "Validating secrets..."

    if ! kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" &> /dev/null; then
        log_error "Secret '$SECRET_NAME' not found"
        return 1
    fi

    local required_keys=(
        "aws-access-key-id"
        "aws-secret-access-key"
        "gcp-service-account-key"
        "api-key"
        "encryption-key"
    )

    local missing=0
    for key in "${required_keys[@]}"; do
        if kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" -o jsonpath="{.data.$key}" | base64 -d | grep -q '^$'; then
            log_warn "Missing or empty: $key"
            ((missing++))
        else
            log_success "Valid: $key"
        fi
    done

    echo ""
    if [ "$missing" -eq 0 ]; then
        log_success "All required secrets are present and non-empty"
        return 0
    else
        log_error "$missing required secret(s) missing or empty"
        return 1
    fi
}

# Export secrets to file
export_secrets() {
    check_kubectl

    local output_file="${1:-$ENV_FILE}"

    log_info "Exporting secrets to $output_file..."

    if [ -f "$output_file" ]; then
        log_warn "File already exists: $output_file"
        read -rp "Overwrite? (y/N): " confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            log_info "Export cancelled"
            return 1
        fi
    fi

    # Create secure env file
    cat > "$output_file" << EOF
# Freightliner Secrets
# Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
# WARNING: Keep this file secure! Add to .gitignore

EOF

    kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" -o json | \
        jq -r '.data | to_entries[] | "\(.key)=\(.value)"' | while read -r line; do
            key=$(echo "$line" | cut -d= -f1 | tr '[:lower:]' '[:upper:]' | tr '-' '_')
            value=$(echo "$line" | cut -d= -f2 | base64 -d)
            echo "$key=\"$value\"" >> "$output_file"
        done

    chmod 600 "$output_file"
    log_success "Secrets exported to $output_file (mode 600)"
    log_warn "Keep this file secure! Add to .gitignore"
}

# Import secrets from file
import_secrets() {
    local input_file="${1:-$ENV_FILE}"

    if [ ! -f "$input_file" ]; then
        log_error "File not found: $input_file"
        return 1
    fi

    log_info "Importing secrets from $input_file..."

    # Source the env file
    set -a
    source "$input_file"
    set +a

    create_secrets
}

# Rotate secrets
rotate_secrets() {
    local secret_type="${1:-all}"

    log_info "Rotating secrets: $secret_type"

    case "$secret_type" in
        encryption)
            ENCRYPTION_KEY=$(generate_random 32)
            log_success "Generated new encryption key"
            ;;
        api)
            API_KEY=$(generate_random 32)
            log_success "Generated new API key"
            ;;
        all)
            ENCRYPTION_KEY=$(generate_random 32)
            API_KEY=$(generate_random 32)
            log_success "Generated new encryption key and API key"
            ;;
        *)
            log_error "Unknown secret type: $secret_type"
            log_info "Valid types: encryption, api, all"
            return 1
            ;;
    esac

    # Export current secrets to backup
    local backup_file=".env.secrets.backup.$(date +%Y%m%d_%H%M%S)"
    export_secrets "$backup_file"
    log_success "Current secrets backed up to: $backup_file"

    # Update secrets
    read -rp "Apply rotated secrets to Kubernetes? (y/N): " confirm
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        kubectl delete secret "$SECRET_NAME" -n "$NAMESPACE" 2>/dev/null || true
        create_secrets
        log_success "Secrets rotated successfully"
    else
        log_info "Rotation cancelled. Backup saved to: $backup_file"
    fi
}

# Delete secrets
delete_secrets() {
    check_kubectl

    log_warn "This will delete secret '$SECRET_NAME' from namespace '$NAMESPACE'"
    read -rp "Are you sure? (y/N): " confirm

    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        log_info "Deletion cancelled"
        return 0
    fi

    if kubectl delete secret "$SECRET_NAME" -n "$NAMESPACE"; then
        log_success "Secret deleted"
    else
        log_error "Failed to delete secret"
        return 1
    fi
}

# Main command dispatcher
main() {
    local command="${1:-help}"
    shift || true

    case "$command" in
        setup)
            setup_wizard
            ;;
        create)
            create_secrets
            ;;
        view)
            view_secrets
            ;;
        validate)
            validate_secrets
            ;;
        export)
            export_secrets "$@"
            ;;
        import)
            import_secrets "$@"
            ;;
        rotate)
            rotate_secrets "$@"
            ;;
        delete)
            delete_secrets
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $command"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
