# Container Registry Configuration Schema Reference

## Overview

This document provides a comprehensive reference for the container registry configuration schema, including detailed examples for each supported registry type, authentication method, and configuration pattern.

## Table of Contents

1. [Root Configuration Structure](#root-configuration-structure)
2. [Registry Types](#registry-types)
3. [Authentication Methods](#authentication-methods)
4. [Secrets Manager Integration](#secrets-manager-integration)
5. [Component Overrides](#component-overrides)
6. [Replication Rules](#replication-rules)
7. [Validation Rules](#validation-rules)
8. [Complete Examples](#complete-examples)

## 1. Root Configuration Structure

### 1.1 Top-Level Schema

```yaml
# config.yaml
registries:
  defaultStrategy: string           # Default registry selection strategy
  defaults: object                  # Global defaults for all registries
  definitions: array                # List of registry definitions

components: object                  # Component-specific overrides

replication: object                 # Registry-to-registry replication rules
```

### 1.2 Default Strategy Options

```yaml
registries:
  defaultStrategy: prefer-private   # Options:
                                     # - prefer-private: Prefer private registries over public
                                     # - public-only: Only use public registries
                                     # - custom-priority: Use explicit priority values
                                     # - fastest-response: Select based on response time
                                     # - least-loaded: Select based on connection pool utilization
```

### 1.3 Global Defaults

```yaml
registries:
  defaults:
    timeout: 300s                   # Connection timeout (default: 300s)
    retryAttempts: 3                # Number of retry attempts (default: 3)
    retryDelay: 5s                  # Delay between retries (default: 5s)
    connectionPoolSize: 10          # Size of connection pool (default: 10)
    tlsVerify: true                 # Verify TLS certificates (default: true)
    keepAlive: 30s                  # Keep-alive interval (default: 30s)
    idleConnTimeout: 90s            # Idle connection timeout (default: 90s)
    maxIdleConns: 100               # Max idle connections (default: 100)
    maxIdleConnsPerHost: 10         # Max idle connections per host (default: 10)
    responseHeaderTimeout: 10s      # Response header timeout (default: 10s)
```

## 2. Registry Types

### 2.1 AWS ECR (Elastic Container Registry)

```yaml
- name: aws-ecr-prod
  type: ecr
  enabled: true
  priority: 1
  config:
    region: us-west-2               # AWS region (required)
    accountID: "123456789012"       # AWS account ID (required)
    endpoint: ""                    # Custom endpoint (optional, for VPC endpoints)
    crossAccountRoleARN: ""         # Cross-account access role (optional)
  authentication:
    method: iam-role                # Options: iam-role, access-key, assume-role, instance-profile
    # ... see Authentication Methods section
  imagePrefix: ""                   # Auto-generated: {accountID}.dkr.ecr.{region}.amazonaws.com
  tags:                             # Optional metadata tags
    environment: production
    team: platform
```

**Supported ECR Features:**
- Private repositories
- Public repositories (public.ecr.aws)
- Cross-region replication
- Image scanning
- Lifecycle policies
- Repository policies

**Authentication Methods:**
- `iam-role`: Use IAM role attached to EC2/ECS/EKS
- `access-key`: Use AWS access key ID and secret access key
- `assume-role`: Assume a different IAM role
- `instance-profile`: Use EC2 instance profile

### 2.2 GCP GCR (Google Container Registry)

```yaml
- name: gcp-gcr-prod
  type: gcr
  enabled: true
  priority: 2
  config:
    project: my-gcp-project         # GCP project ID (required)
    location: us                    # Location: us, eu, asia, or specific region (required)
    endpoint: ""                    # Custom endpoint (optional)
    legacy: false                   # Use legacy GCR (gcr.io) vs Artifact Registry (optional)
  authentication:
    method: service-account         # Options: service-account, adc, workload-identity
    # ... see Authentication Methods section
  imagePrefix: ""                   # Auto-generated: {location}.gcr.io/{project}
  tags:
    environment: production
    cost-center: engineering
```

**Supported GCR Features:**
- Standard storage classes
- Multi-regional repositories
- Artifact Registry (next-gen GCR)
- Vulnerability scanning
- Binary authorization
- Pub/Sub notifications

**Authentication Methods:**
- `service-account`: Use service account JSON key
- `adc`: Use Application Default Credentials
- `workload-identity`: Use GKE Workload Identity

### 2.3 Azure ACR (Azure Container Registry)

```yaml
- name: azure-acr-prod
  type: acr
  enabled: true
  priority: 3
  config:
    registryName: myregistry         # ACR registry name (required)
    resourceGroup: my-rg             # Resource group (optional)
    subscriptionID: "uuid"           # Subscription ID (optional)
    endpoint: ""                     # Custom endpoint (optional, defaults to {name}.azurecr.io)
    sku: Premium                     # SKU: Basic, Standard, Premium (optional)
    loginServer: ""                  # Login server (auto-detected if empty)
  authentication:
    method: managed-identity         # Options: managed-identity, service-principal, admin-credentials
    # ... see Authentication Methods section
  imagePrefix: ""                    # Auto-generated: {registryName}.azurecr.io
  tags:
    environment: production
```

**Supported ACR Features:**
- Geo-replication
- Content trust
- Webhook notifications
- Helm chart storage
- OCI artifacts
- Azure AD integration

**Authentication Methods:**
- `managed-identity`: Use Azure Managed Identity
- `service-principal`: Use Azure Service Principal
- `admin-credentials`: Use admin username/password (not recommended for production)

### 2.4 DockerHub

```yaml
- name: dockerhub-public
  type: dockerhub
  enabled: true
  priority: 10
  config:
    endpoint: https://registry-1.docker.io  # DockerHub registry endpoint
    namespace: myorg                 # Organization or username (optional)
    useHubProxy: false              # Use Docker Hub pull-through cache (optional)
  authentication:
    method: username-password        # Options: username-password, token, anonymous
    # ... see Authentication Methods section
  imagePrefix: ""                    # No prefix (or use namespace: myorg/)
  rateLimits:                        # DockerHub rate limits
    pullsPerHour: 200                # Authenticated pulls per hour
    anonymousPullsPerHour: 100       # Anonymous pulls per hour
```

**DockerHub Features:**
- Public repositories
- Private repositories
- Organizations
- Teams and access control
- Automated builds
- Webhooks

**Rate Limiting:**
- Anonymous: 100 pulls per 6 hours
- Authenticated (Free): 200 pulls per 6 hours
- Authenticated (Pro): 5000 pulls per day
- Authenticated (Team): 5000 pulls per day per user

### 2.5 Harbor (Private Registry)

```yaml
- name: harbor-enterprise
  type: harbor
  enabled: true
  priority: 1
  config:
    endpoint: https://harbor.company.com  # Harbor endpoint (required)
    project: production              # Harbor project name (required)
    apiVersion: v2.0                 # Harbor API version (optional, default: v2.0)
    tlsVerify: true                  # Verify TLS certificate
    caCert: /etc/ssl/certs/harbor-ca.crt  # Custom CA certificate path (optional)
    chartRepository: true            # Enable Helm chart storage (optional)
  authentication:
    method: robot-account            # Options: username-password, robot-account, oidc
    # ... see Authentication Methods section
  imagePrefix: harbor.company.com/production
  features:
    replication: true                # Harbor-to-Harbor replication
    contentTrust: true               # Notary content trust
    vulnerability-scanning: true     # Trivy/Clair scanning
    immutableTags: true              # Tag immutability
```

**Harbor Features:**
- Multi-project support
- Role-based access control (RBAC)
- Replication policies
- Vulnerability scanning (Trivy/Clair)
- Content trust (Notary)
- Helm chart repository
- Garbage collection
- Audit logs

### 2.6 Quay.io

```yaml
- name: quay-enterprise
  type: quay
  enabled: true
  priority: 5
  config:
    endpoint: https://quay.io        # Quay.io endpoint (required)
    namespace: myorganization        # Organization namespace (required)
    useProxy: false                  # Use pull-through cache (optional)
  authentication:
    method: robot-account            # Options: username-password, robot-account, oauth-token
    # ... see Authentication Methods section
  imagePrefix: quay.io/myorganization
  features:
    buildTriggers: true              # Automated container builds
    notifications: true              # Webhook notifications
    securityScanning: true           # Clair vulnerability scanning
```

**Quay Features:**
- Robot accounts
- Teams and organizations
- Repository mirroring
- Vulnerability scanning (Clair)
- Geo-replication
- Time machine (image history)
- Build triggers
- Application repositories

### 2.7 Generic OCI Registry

```yaml
- name: custom-registry
  type: generic
  enabled: true
  priority: 5
  config:
    endpoint: https://registry.company.com  # Registry endpoint (required)
    ociVersion: v1                   # OCI distribution spec version (optional)
    apiVersion: v2                   # Docker registry API version (optional)
    tlsVerify: true                  # Verify TLS certificates
    caCert: /etc/ssl/certs/custom-ca.crt  # Custom CA certificate
    insecureSkipTLSVerify: false     # Skip TLS verification (not recommended)
    checkPing: true                  # Check /v2/ ping endpoint on init
  authentication:
    method: basic-auth               # Options: basic-auth, bearer-token, mtls
    # ... see Authentication Methods section
  imagePrefix: registry.company.com
```

**Compatible with:**
- Nexus Repository Manager
- JFrog Artifactory
- GitLab Container Registry
- GitHub Container Registry (ghcr.io)
- DigitalOcean Container Registry
- Any OCI-compliant registry

### 2.8 GitHub Container Registry (ghcr.io)

```yaml
- name: github-packages
  type: generic  # Use generic type for ghcr.io
  enabled: true
  priority: 7
  config:
    endpoint: https://ghcr.io
    namespace: myorganization        # GitHub organization or username
  authentication:
    method: bearer-token
    bearerToken: "${GITHUB_TOKEN}"   # GitHub Personal Access Token
  imagePrefix: ghcr.io/myorganization
```

## 3. Authentication Methods

### 3.1 AWS IAM Role

```yaml
authentication:
  method: iam-role
  # No additional configuration needed
  # Uses attached IAM role (EC2, ECS, EKS, Lambda)
```

### 3.2 AWS Access Key

```yaml
authentication:
  method: access-key
  accessKeyID: "${AWS_ACCESS_KEY_ID}"       # From environment variable
  secretAccessKey: "${AWS_SECRET_ACCESS_KEY}"  # From environment variable
  sessionToken: "${AWS_SESSION_TOKEN}"      # Optional, for temporary credentials
  region: us-west-2                          # Optional, defaults to registry region
```

### 3.3 AWS Assume Role

```yaml
authentication:
  method: assume-role
  roleARN: arn:aws:iam::123456789012:role/ECRAccessRole
  sessionName: freightliner-session
  externalID: unique-external-id             # Optional, for cross-account access
  duration: 3600                             # Session duration in seconds
  policy: ""                                 # Optional inline session policy
```

### 3.4 GCP Service Account

```yaml
authentication:
  method: service-account
  credentialsFile: /path/to/service-account.json  # Path to JSON key file
  # OR
  credentialsJSON: '{"type":"service_account",...}'  # Inline JSON (not recommended)
  # OR use environment variable
  credentialsFile: "${GOOGLE_APPLICATION_CREDENTIALS}"
```

### 3.5 GCP Application Default Credentials

```yaml
authentication:
  method: adc  # Uses gcloud CLI credentials or metadata service
  # No additional configuration needed
```

### 3.6 GCP Workload Identity

```yaml
authentication:
  method: workload-identity
  serviceAccount: freightliner@project.iam.gserviceaccount.com
  # Used in GKE with Workload Identity enabled
```

### 3.7 Azure Managed Identity

```yaml
authentication:
  method: managed-identity
  clientID: ""  # Optional, for user-assigned managed identity
  # Uses system-assigned or user-assigned managed identity
```

### 3.8 Azure Service Principal

```yaml
authentication:
  method: service-principal
  clientID: "${AZURE_CLIENT_ID}"
  clientSecret: "${AZURE_CLIENT_SECRET}"
  tenantID: "${AZURE_TENANT_ID}"
```

### 3.9 Username/Password (Basic Auth)

```yaml
authentication:
  method: username-password
  username: "${REGISTRY_USERNAME}"
  password: "${REGISTRY_PASSWORD}"
  # Used for: DockerHub, Harbor, Nexus, Artifactory, etc.
```

### 3.10 Bearer Token

```yaml
authentication:
  method: bearer-token
  bearerToken: "${REGISTRY_TOKEN}"
  # Used for: GitHub Container Registry, GitLab, etc.
```

### 3.11 Robot Account (Harbor/Quay)

```yaml
authentication:
  method: robot-account
  robotName: robot$freightliner       # Harbor format
  robotToken: "${ROBOT_SECRET}"
  # OR
  username: myorg+robot_name          # Quay format
  password: "${QUAY_ROBOT_TOKEN}"
```

### 3.12 OAuth Token

```yaml
authentication:
  method: oauth-token
  oauthToken: "${OAUTH_TOKEN}"
  tokenEndpoint: https://auth.example.com/token  # Optional
  clientID: freightliner
  clientSecret: "${CLIENT_SECRET}"
  scope: "registry:pull,push"
```

### 3.13 Mutual TLS (mTLS)

```yaml
authentication:
  method: mtls
  clientCert: /path/to/client.crt
  clientKey: /path/to/client.key
  caCert: /path/to/ca.crt
```

## 4. Secrets Manager Integration

### 4.1 AWS Secrets Manager

```yaml
authentication:
  method: username-password
  secretsManager:
    enabled: true
    provider: aws
    secretName: freightliner/registry/harbor-prod
    region: us-west-2
    versionId: ""                    # Optional, specific version
    versionStage: AWSCURRENT         # Optional, default: AWSCURRENT

# Secret format in AWS Secrets Manager:
# {
#   "username": "admin",
#   "password": "secure-password",
#   "endpoint": "https://harbor.example.com"  // Optional overrides
# }
```

### 4.2 GCP Secret Manager

```yaml
authentication:
  method: service-account
  secretsManager:
    enabled: true
    provider: gcp
    secretName: freightliner-gcr-credentials
    project: my-gcp-project
    version: latest                  # Optional, default: latest

# Secret format in GCP Secret Manager:
# Store the entire service account JSON as a secret
```

### 4.3 Azure Key Vault

```yaml
authentication:
  method: service-principal
  secretsManager:
    enabled: true
    provider: azure
    secretName: freightliner-acr-credentials
    vaultName: my-keyvault
    vaultURL: https://my-keyvault.vault.azure.net/  # Optional, auto-constructed
    version: ""                      # Optional, specific version

# Secret format in Azure Key Vault:
# Store as JSON: {"clientId":"...","clientSecret":"...","tenantId":"..."}
```

### 4.4 HashiCorp Vault

```yaml
authentication:
  method: username-password
  secretsManager:
    enabled: true
    provider: vault
    secretName: freightliner/registry/harbor
    vaultAddr: https://vault.example.com
    vaultToken: "${VAULT_TOKEN}"
    vaultNamespace: ""               # Optional, for Vault Enterprise
    secretPath: secret/data/         # Optional, custom mount path
    kvVersion: 2                     # KV secrets engine version (1 or 2)

# Secret format in Vault:
# vault kv put secret/freightliner/registry/harbor \
#   username=admin \
#   password=secure-password
```

### 4.5 Kubernetes Secrets (for in-cluster deployments)

```yaml
authentication:
  method: username-password
  secretsManager:
    enabled: true
    provider: kubernetes
    secretName: harbor-credentials
    namespace: freightliner           # Optional, defaults to pod namespace
    key: .dockerconfigjson           # Optional, for Docker config secrets

# Kubernetes Secret format:
# apiVersion: v1
# kind: Secret
# metadata:
#   name: harbor-credentials
# type: kubernetes.io/dockerconfigjson
# data:
#   .dockerconfigjson: <base64-encoded-config>
```

## 5. Component Overrides

### 5.1 Component-Specific Registry Selection

```yaml
components:
  # Override for ECS tasks
  ecs-tasks:
    registryOverrides:
      - name: aws-ecr-prod
        priority: 1
        includeImages:
          - "app-*:*"                 # Application images
          - "backend/*:v*"            # Backend services with version tags
        excludeImages:
          - "*:dev"                   # Exclude dev tags
          - "*:test"                  # Exclude test tags

      - name: dockerhub-public
        priority: 2
        includeImages:
          - "nginx:*"                 # Public NGINX images
          - "postgres:*"              # Public Postgres images

  # Override for Kubernetes deployments
  kubernetes-deployments:
    registryOverrides:
      - name: harbor-private
        priority: 1
        includeImages:
          - "internal/*"              # All internal images
          - "microservices/*"         # All microservices

      - name: quay-enterprise
        priority: 2
        includeImages:
          - "operators/*"             # Kubernetes operators

      - name: ghcr-public
        priority: 3
        includeImages:
          - "opensource/*"            # Open source images

  # Override for CI/CD pipelines
  cicd-pipeline:
    registryOverrides:
      - name: harbor-private
        priority: 1
        includeImages:
          - "build-cache/*"           # Build cache images

  # Override for Lambda container images
  lambda-functions:
    registryOverrides:
      - name: aws-ecr-prod
        priority: 1  # Lambda only supports ECR
```

### 5.2 Environment-Based Overrides

```yaml
components:
  # Development environment
  development:
    registryOverrides:
      - name: harbor-dev
        priority: 1
        includeImages:
          - "*"  # All images from dev Harbor

  # Staging environment
  staging:
    registryOverrides:
      - name: harbor-staging
        priority: 1
      - name: dockerhub-public
        priority: 2

  # Production environment
  production:
    registryOverrides:
      - name: aws-ecr-prod
        priority: 1
      - name: azure-acr-prod
        priority: 2
      - name: harbor-prod
        priority: 3
      # No public registries in production
```

## 6. Replication Rules

### 6.1 Scheduled Replication

```yaml
replication:
  rules:
    - name: ecr-to-harbor-nightly
      enabled: true
      description: "Replicate production images from ECR to Harbor nightly"
      sourceRegistry: aws-ecr-prod
      targetRegistry: harbor-prod
      schedule: "0 2 * * *"           # Cron schedule: daily at 2 AM
      imageFilters:
        includeRepositories:
          - "production/*"
          - "apps/*"
        includeTags:
          - "v*"                       # Version tags
          - "latest"
        excludeTags:
          - "*-dev"
          - "*-test"
      options:
        overwrite: false               # Don't overwrite existing images
        deleteRemote: false            # Keep source images
        skipScanResult: false          # Include scan results
        triggerOnChange: true          # Trigger on source change
```

### 6.2 On-Demand Replication

```yaml
replication:
  rules:
    - name: dockerhub-to-harbor-mirror
      enabled: true
      description: "Mirror public images from DockerHub to Harbor on-demand"
      sourceRegistry: dockerhub-public
      targetRegistry: harbor-prod
      trigger: on-demand               # Manual or API-triggered
      imageFilters:
        includeRepositories:
          - "nginx"
          - "postgres"
          - "redis"
        includeTags:
          - "latest"
          - "stable"
          - "*-alpine"
      options:
        overwrite: true                # Update existing images
        verifyCertificate: true
        bandwidth: 100                 # Limit to 100 MB/s
```

### 6.3 Event-Driven Replication

```yaml
replication:
  rules:
    - name: webhook-triggered-sync
      enabled: true
      description: "Replicate on webhook events"
      sourceRegistry: harbor-dev
      targetRegistry: harbor-staging
      trigger: webhook                 # Triggered by webhook
      webhookURL: https://freightliner.example.com/api/v1/replicate/webhook
      webhookSecret: "${WEBHOOK_SECRET}"
      imageFilters:
        includeRepositories:
          - "apps/*"
        includeTags:
          - "release-*"
      options:
        async: true                    # Replicate asynchronously
        priority: high                 # High priority queue
```

### 6.4 Bidirectional Replication

```yaml
replication:
  rules:
    - name: harbor-us-to-eu
      enabled: true
      direction: bidirectional         # Sync in both directions
      registry1: harbor-us
      registry2: harbor-eu
      schedule: "*/30 * * * *"         # Every 30 minutes
      conflictResolution: latest-timestamp  # Resolve conflicts by timestamp
      imageFilters:
        includeRepositories:
          - "global/*"                  # Only replicate global images
```

## 7. Validation Rules

### 7.1 Configuration Validation

```yaml
# Enable strict validation
validation:
  enabled: true
  strict: true                         # Fail on warnings
  checkConnectivity: true              # Test registry connections on startup
  validateCredentials: true            # Validate authentication on startup
  allowUnknownFields: false            # Reject unknown configuration fields
```

### 7.2 Schema Validation Rules

**Registry Name:**
- Pattern: `^[a-z0-9-]+$`
- Min length: 1
- Max length: 64
- Unique within configuration

**Priority:**
- Type: integer
- Min: 1
- Max: 100
- Unique per registry type (recommended)

**Timeout:**
- Pattern: `^[0-9]+(s|m|h)$`
- Examples: `30s`, `5m`, `1h`
- Min: 1s
- Max: 1h

**Registry Type:**
- Enum: `ecr`, `gcr`, `acr`, `dockerhub`, `harbor`, `quay`, `generic`

**Authentication Method (by registry type):**
- ECR: `iam-role`, `access-key`, `assume-role`, `instance-profile`
- GCR: `service-account`, `adc`, `workload-identity`
- ACR: `managed-identity`, `service-principal`, `admin-credentials`
- Harbor/Quay: `username-password`, `robot-account`, `oidc`
- Generic: `basic-auth`, `bearer-token`, `mtls`, `anonymous`

## 8. Complete Examples

### 8.1 Multi-Cloud Production Setup

```yaml
registries:
  defaultStrategy: prefer-private
  defaults:
    timeout: 300s
    retryAttempts: 3
    connectionPoolSize: 20
    tlsVerify: true

  definitions:
    # Primary: AWS ECR
    - name: aws-ecr-us-west
      type: ecr
      enabled: true
      priority: 1
      config:
        region: us-west-2
        accountID: "123456789012"
      authentication:
        method: iam-role

    # Secondary: GCP GCR
    - name: gcp-gcr-us
      type: gcr
      enabled: true
      priority: 2
      config:
        project: my-prod-project
        location: us
      authentication:
        method: service-account
        credentialsFile: "${GOOGLE_APPLICATION_CREDENTIALS}"

    # Tertiary: Azure ACR
    - name: azure-acr-east
      type: acr
      enabled: true
      priority: 3
      config:
        registryName: mycompanyprod
        subscriptionID: "${AZURE_SUBSCRIPTION_ID}"
      authentication:
        method: managed-identity

    # Private Harbor
    - name: harbor-prod
      type: harbor
      enabled: true
      priority: 1  # Highest priority for private registry
      config:
        endpoint: https://harbor.company.com
        project: production
        tlsVerify: true
      authentication:
        method: robot-account
        secretsManager:
          enabled: true
          provider: vault
          secretName: freightliner/harbor-prod
          vaultAddr: https://vault.company.com

    # Public fallback
    - name: dockerhub
      type: dockerhub
      enabled: true
      priority: 100
      config:
        endpoint: https://registry-1.docker.io
      authentication:
        method: username-password
        username: "${DOCKERHUB_USER}"
        password: "${DOCKERHUB_TOKEN}"

components:
  kubernetes-prod:
    registryOverrides:
      - name: harbor-prod
        priority: 1
        includeImages: ["internal/*", "microservices/*"]
      - name: aws-ecr-us-west
        priority: 2
        includeImages: ["aws/*", "lambda/*"]

  ecs-prod:
    registryOverrides:
      - name: aws-ecr-us-west
        priority: 1

replication:
  rules:
    - name: ecr-to-harbor-sync
      enabled: true
      sourceRegistry: aws-ecr-us-west
      targetRegistry: harbor-prod
      schedule: "0 */6 * * *"  # Every 6 hours
      imageFilters:
        includeRepositories: ["production/*"]
        includeTags: ["v*", "latest"]
```

### 8.2 Development Environment

```yaml
registries:
  defaultStrategy: fastest-response
  defaults:
    timeout: 60s
    retryAttempts: 2
    connectionPoolSize: 5

  definitions:
    - name: local-harbor
      type: harbor
      enabled: true
      priority: 1
      config:
        endpoint: https://harbor.dev.local
        project: development
        tlsVerify: false  # Dev environment with self-signed cert
      authentication:
        method: username-password
        username: admin
        password: Harbor12345

    - name: dockerhub
      type: dockerhub
      enabled: true
      priority: 10
      config:
        endpoint: https://registry-1.docker.io
      authentication:
        method: anonymous

components:
  local-kubernetes:
    registryOverrides:
      - name: local-harbor
        priority: 1
```

### 8.3 Air-Gapped Environment

```yaml
registries:
  defaultStrategy: custom-priority
  defaults:
    timeout: 600s
    retryAttempts: 1
    connectionPoolSize: 5

  definitions:
    - name: internal-registry
      type: harbor
      enabled: true
      priority: 1
      config:
        endpoint: https://registry.internal.corp
        project: airgap
        tlsVerify: true
        caCert: /etc/ssl/certs/internal-ca.crt
      authentication:
        method: mtls
        clientCert: /etc/ssl/certs/freightliner.crt
        clientKey: /etc/ssl/private/freightliner.key

components:
  all:
    registryOverrides:
      - name: internal-registry
        priority: 1
        includeImages: ["*"]  # All images must come from internal registry
```

---

**Document Version:** 1.0
**Last Updated:** 2025-12-02
**Related:** [Architecture Overview](./registry-support-architecture.md)
