# Container Registry Support Architecture

## Executive Summary

This document outlines the architecture for adding comprehensive container registry support to Freightliner, a Go-based container image replication tool. The design supports local registries, third-party registries, and multi-cloud providers (AWS ECR, GCP GCR, Azure ACR, DockerHub, Harbor, etc.) with flexible authentication and configuration patterns.

## 1. Architecture Overview

### 1.1 Current State Analysis

**Existing Components:**
- `/pkg/config/config.go` - Configuration management with ECR and GCR support
- `/pkg/client/ecr/` - AWS ECR client implementation
- `/pkg/client/gcr/` - GCP GCR client implementation
- `/pkg/client/common/registry_util.go` - Common registry utilities
- `/pkg/metrics/registry.go` - Prometheus metrics for registry operations
- `config.yaml` - Basic registry configuration (aws.region, gcp.region)

**Limitations:**
- Hard-coded support for only ECR and GCR
- Limited authentication methods (AWS/GCP credentials only)
- No support for custom/private registries
- Registry configuration scattered across multiple config sections

### 1.2 Design Principles

1. **Extensibility**: Easy to add new registry types without modifying core logic
2. **Security**: Secure credential management with multiple authentication methods
3. **Flexibility**: Support for global and per-component registry overrides
4. **Backwards Compatibility**: Existing configurations continue to work
5. **Multi-Cloud**: First-class support for AWS, GCP, Azure, and custom registries
6. **Performance**: Efficient connection pooling and credential caching

## 2. Configuration Schema Design

### 2.1 Enhanced Configuration Structure

```yaml
# Global registry configuration
registries:
  # Default registry selection strategy
  defaultStrategy: prefer-private  # Options: prefer-private, public-only, custom-priority

  # Default authentication timeout and retry settings
  defaults:
    timeout: 300s
    retryAttempts: 3
    retryDelay: 5s
    connectionPoolSize: 10
    tlsVerify: true

  # Registry definitions
  definitions:
    # AWS ECR
    - name: aws-ecr-primary
      type: ecr
      enabled: true
      priority: 1  # Lower number = higher priority
      config:
        region: us-west-2
        accountID: "123456789012"
        endpoint: ""  # Optional: custom endpoint for VPC endpoints
      authentication:
        method: iam-role  # Options: iam-role, access-key, assume-role, instance-profile
        # For access-key method
        accessKeyID: ""  # Can use env var: ${AWS_ACCESS_KEY_ID}
        secretAccessKey: ""  # Can use env var: ${AWS_SECRET_ACCESS_KEY}
        # For assume-role method
        roleARN: ""
        sessionName: "freightliner-session"
        # Secrets manager integration
        secretsManager:
          enabled: false
          secretName: "freightliner/registry/aws-ecr-primary"
          region: us-west-2
      imagePrefix: ""  # Automatically constructed: {accountID}.dkr.ecr.{region}.amazonaws.com

    # GCP GCR
    - name: gcp-gcr-primary
      type: gcr
      enabled: true
      priority: 2
      config:
        project: my-gcp-project
        location: us  # Options: us, eu, asia, or specific region
        endpoint: ""  # Optional: custom endpoint
      authentication:
        method: service-account  # Options: service-account, adc, workload-identity
        # For service-account method
        credentialsFile: /path/to/service-account.json
        credentialsJSON: ""  # Inline JSON credentials
        # Secrets manager integration
        secretsManager:
          enabled: false
          secretName: "freightliner/registry/gcp-gcr-primary"
          project: my-gcp-project
      imagePrefix: ""  # Automatically constructed: {location}.gcr.io/{project}

    # Azure ACR
    - name: azure-acr-primary
      type: acr
      enabled: true
      priority: 3
      config:
        registryName: myacr
        resourceGroup: my-resource-group
        subscriptionID: "12345678-1234-1234-1234-123456789012"
        endpoint: ""  # Optional: defaults to {registryName}.azurecr.io
      authentication:
        method: managed-identity  # Options: managed-identity, service-principal, admin-credentials
        # For service-principal method
        clientID: ""
        clientSecret: ""
        tenantID: ""
        # Secrets manager integration
        secretsManager:
          enabled: false
          secretName: "freightliner/registry/azure-acr-primary"
          vaultName: my-keyvault
      imagePrefix: ""  # Automatically constructed: {registryName}.azurecr.io

    # DockerHub
    - name: dockerhub-public
      type: dockerhub
      enabled: true
      priority: 10
      config:
        endpoint: https://registry-1.docker.io  # Official DockerHub registry
        namespace: ""  # Optional: organization/user namespace
      authentication:
        method: username-password  # Options: username-password, token, anonymous
        username: ""  # Can use env var: ${DOCKERHUB_USERNAME}
        password: ""  # Can use env var: ${DOCKERHUB_PASSWORD}
        token: ""     # Alternative to username/password
        # Secrets manager integration
        secretsManager:
          enabled: false
          provider: aws  # Options: aws, gcp, azure, vault
          secretName: "freightliner/registry/dockerhub"
      imagePrefix: ""  # No prefix for DockerHub (or use namespace if provided)

    # Harbor (Private Registry)
    - name: harbor-private
      type: harbor
      enabled: true
      priority: 1  # High priority for private registry
      config:
        endpoint: https://harbor.example.com
        project: my-project
        tlsVerify: true
        caCert: /path/to/ca.crt  # Optional: custom CA certificate
      authentication:
        method: username-password  # Options: username-password, robot-account, oidc
        username: admin
        password: "${HARBOR_PASSWORD}"
        # For robot-account method
        robotToken: ""
        # Secrets manager integration
        secretsManager:
          enabled: true
          provider: vault
          secretName: "freightliner/registry/harbor"
          vaultAddr: https://vault.example.com
          vaultToken: "${VAULT_TOKEN}"
      imagePrefix: harbor.example.com/my-project

    # Generic Private Registry (Nexus, Artifactory, etc.)
    - name: custom-registry
      type: generic
      enabled: true
      priority: 5
      config:
        endpoint: https://registry.mycompany.com
        tlsVerify: true
        caCert: /path/to/ca.crt
        insecureSkipTLSVerify: false  # Use with caution
      authentication:
        method: basic-auth  # Options: basic-auth, bearer-token, mtls
        username: "${REGISTRY_USER}"
        password: "${REGISTRY_PASS}"
        # For bearer-token method
        bearerToken: ""
        # For mTLS method
        clientCert: /path/to/client.crt
        clientKey: /path/to/client.key
        # Secrets manager integration
        secretsManager:
          enabled: false
      imagePrefix: registry.mycompany.com

    # Quay.io
    - name: quay-io
      type: quay
      enabled: true
      priority: 8
      config:
        endpoint: https://quay.io
        namespace: myorg  # Organization or user namespace
      authentication:
        method: robot-account  # Options: username-password, robot-account, oauth-token
        username: myorg+robot_name
        password: "${QUAY_ROBOT_TOKEN}"
      imagePrefix: quay.io/myorg

# Component-level registry overrides
components:
  ecs-tasks:
    registryOverrides:
      - name: aws-ecr-primary
        includeImages:
          - "nginx:*"
          - "app-*:latest"
        excludeImages:
          - "*:dev"

  kubernetes-deployments:
    registryOverrides:
      - name: harbor-private
        includeImages:
          - "internal/*"
      - name: dockerhub-public
        includeImages:
          - "public/*"

# Replication rules (for registry-to-registry sync)
replication:
  rules:
    - name: sync-ecr-to-harbor
      enabled: true
      sourceRegistry: aws-ecr-primary
      targetRegistry: harbor-private
      imageFilters:
        - "production/*:v*"
      schedule: "0 2 * * *"  # Daily at 2 AM

    - name: mirror-dockerhub-public
      enabled: true
      sourceRegistry: dockerhub-public
      targetRegistry: harbor-private
      imageFilters:
        - "nginx:*"
        - "postgres:*"
      onDemand: true  # Trigger only when needed
```

### 2.2 Configuration Schema Validation

**JSON Schema Definition:**

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "registries": {
      "type": "object",
      "properties": {
        "defaultStrategy": {
          "type": "string",
          "enum": ["prefer-private", "public-only", "custom-priority"]
        },
        "defaults": {
          "type": "object",
          "properties": {
            "timeout": {"type": "string", "pattern": "^[0-9]+(s|m|h)$"},
            "retryAttempts": {"type": "integer", "minimum": 0, "maximum": 10},
            "retryDelay": {"type": "string", "pattern": "^[0-9]+(s|m|h)$"},
            "connectionPoolSize": {"type": "integer", "minimum": 1, "maximum": 100},
            "tlsVerify": {"type": "boolean"}
          }
        },
        "definitions": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["name", "type"],
            "properties": {
              "name": {"type": "string", "pattern": "^[a-z0-9-]+$"},
              "type": {
                "type": "string",
                "enum": ["ecr", "gcr", "acr", "dockerhub", "harbor", "quay", "generic"]
              },
              "enabled": {"type": "boolean"},
              "priority": {"type": "integer", "minimum": 1, "maximum": 100},
              "config": {"type": "object"},
              "authentication": {"type": "object"},
              "imagePrefix": {"type": "string"}
            }
          }
        }
      },
      "required": ["definitions"]
    }
  }
}
```

## 3. Component Architecture

### 3.1 New Components

```
pkg/
├── registry/
│   ├── manager.go           # RegistryManager: Central registry coordination
│   ├── types.go             # Common types and interfaces
│   ├── factory.go           # Registry client factory
│   ├── auth/
│   │   ├── provider.go      # Authentication provider interface
│   │   ├── aws_auth.go      # AWS IAM authentication
│   │   ├── gcp_auth.go      # GCP service account authentication
│   │   ├── azure_auth.go    # Azure managed identity authentication
│   │   ├── basic_auth.go    # Username/password authentication
│   │   ├── token_auth.go    # Bearer token authentication
│   │   └── mtls_auth.go     # Mutual TLS authentication
│   ├── clients/
│   │   ├── ecr_client.go    # Enhanced AWS ECR client
│   │   ├── gcr_client.go    # Enhanced GCP GCR client
│   │   ├── acr_client.go    # NEW: Azure ACR client
│   │   ├── dockerhub_client.go  # NEW: DockerHub client
│   │   ├── harbor_client.go     # NEW: Harbor client
│   │   ├── quay_client.go       # NEW: Quay.io client
│   │   └── generic_client.go    # NEW: Generic OCI registry client
│   ├── secrets/
│   │   ├── provider.go      # Secrets provider interface
│   │   ├── aws_secrets.go   # AWS Secrets Manager
│   │   ├── gcp_secrets.go   # GCP Secret Manager
│   │   ├── azure_secrets.go # Azure Key Vault
│   │   └── vault_secrets.go # HashiCorp Vault
│   └── selector.go          # Registry selection logic
```

### 3.2 Core Interfaces

```go
// RegistryManager - Central coordinator for all registry operations
type RegistryManager interface {
    // Initialize loads configuration and prepares all registries
    Initialize(ctx context.Context, config *Config) error

    // GetRegistry returns a specific registry client by name
    GetRegistry(name string) (RegistryClient, error)

    // SelectRegistry chooses the best registry for an image based on policy
    SelectRegistry(ctx context.Context, imageName string, policy SelectionPolicy) (RegistryClient, error)

    // ListRegistries returns all enabled registries
    ListRegistries() []RegistryInfo

    // GetImageURI constructs the full image URI for a registry
    GetImageURI(registryName, imageName, tag string) (string, error)

    // ValidateAccess tests connectivity and authentication to a registry
    ValidateAccess(ctx context.Context, registryName string) error

    // Close gracefully shuts down all registry connections
    Close() error
}

// RegistryClient - Unified interface for all registry types
type RegistryClient interface {
    // Metadata
    Name() string
    Type() RegistryType
    Endpoint() string
    ImagePrefix() string

    // Authentication
    Authenticate(ctx context.Context) error
    RefreshCredentials(ctx context.Context) error

    // Repository operations
    ListRepositories(ctx context.Context) ([]string, error)
    GetRepository(ctx context.Context, name string) (Repository, error)
    CreateRepository(ctx context.Context, name string, opts CreateOptions) error
    DeleteRepository(ctx context.Context, name string) error

    // Image operations
    PushImage(ctx context.Context, image Image) error
    PullImage(ctx context.Context, ref ImageReference) (Image, error)
    TagImage(ctx context.Context, src, dst ImageReference) error
    DeleteImage(ctx context.Context, ref ImageReference) error

    // Manifest operations
    GetManifest(ctx context.Context, ref ImageReference) (Manifest, error)
    PutManifest(ctx context.Context, ref ImageReference, manifest Manifest) error

    // Health and metrics
    Ping(ctx context.Context) error
    GetMetrics() RegistryMetrics

    // Lifecycle
    Close() error
}

// AuthProvider - Authentication provider interface
type AuthProvider interface {
    // GetCredentials retrieves credentials for authentication
    GetCredentials(ctx context.Context) (*Credentials, error)

    // RefreshCredentials refreshes expired credentials
    RefreshCredentials(ctx context.Context) (*Credentials, error)

    // Validate checks if credentials are valid
    Validate(ctx context.Context) error

    // Type returns the authentication method type
    Type() AuthMethodType
}

// SecretsProvider - Secrets management interface
type SecretsProvider interface {
    // GetSecret retrieves a secret by name
    GetSecret(ctx context.Context, name string) ([]byte, error)

    // GetSecretString retrieves a secret as a string
    GetSecretString(ctx context.Context, name string) (string, error)

    // GetSecretJSON retrieves and unmarshals a JSON secret
    GetSecretJSON(ctx context.Context, name string, target interface{}) error

    // PutSecret stores a secret
    PutSecret(ctx context.Context, name string, value []byte) error

    // DeleteSecret removes a secret
    DeleteSecret(ctx context.Context, name string) error

    // Provider returns the secrets provider type
    Provider() SecretsProviderType
}
```

## 4. Integration Architecture

### 4.1 Configuration Loading Flow

```
1. Load config.yaml → Parse YAML
2. Validate schema → JSON Schema validation
3. Expand environment variables → ${VAR} substitution
4. Resolve secrets → Fetch from secrets managers
5. Build RegistryManager → Initialize clients
6. Validate connections → Test authentication
7. Cache credentials → Prepare for operations
```

### 4.2 Registry Selection Logic

```go
type SelectionPolicy struct {
    Strategy         SelectionStrategy  // prefer-private, public-only, custom-priority
    PreferredTypes   []RegistryType     // Preferred registry types in order
    RequirePrivate   bool               // Only use private registries
    IncludePatterns  []string           // Image patterns to include
    ExcludePatterns  []string           // Image patterns to exclude
    ComponentContext string             // Component requesting the registry
}

// Selection algorithm
func (m *RegistryManager) SelectRegistry(ctx context.Context, imageName string, policy SelectionPolicy) (RegistryClient, error) {
    // 1. Filter enabled registries
    candidates := m.filterEnabled()

    // 2. Apply component-specific overrides
    if policy.ComponentContext != "" {
        candidates = m.applyComponentOverrides(candidates, policy.ComponentContext, imageName)
    }

    // 3. Apply image pattern filters
    candidates = m.applyImageFilters(candidates, imageName, policy)

    // 4. Sort by priority and strategy
    candidates = m.sortByPriority(candidates, policy.Strategy)

    // 5. Validate connectivity and return first available
    for _, registry := range candidates {
        if err := registry.Ping(ctx); err == nil {
            return registry, nil
        }
    }

    return nil, ErrNoRegistryAvailable
}
```

### 4.3 Authentication Flow

```
1. Registry client initialization
2. Check authentication method type
3. Load credentials:
   a. From secrets manager (if configured)
   b. From environment variables
   c. From configuration file
   d. From cloud provider metadata service
4. Create authentication provider
5. Authenticate with registry
6. Cache credentials with TTL
7. Setup auto-refresh (for temporary tokens)
8. Register refresh handler
```

### 4.4 Component Integration Pattern

**Before (Hard-coded ECR):**
```go
// pkg/service/replicate.go
func (s *Service) ReplicateImage(ctx context.Context, src, dst string) error {
    // Hard-coded ECR logic
    ecrClient := ecr.NewClient(s.config.ECR.Region, s.config.ECR.AccountID)
    // ...
}
```

**After (Dynamic Registry Selection):**
```go
// pkg/service/replicate.go
func (s *Service) ReplicateImage(ctx context.Context, src, dst string) error {
    // Select source registry dynamically
    srcRegistry, err := s.registryManager.SelectRegistry(ctx, src, SelectionPolicy{
        Strategy:       PreferPrivate,
        ComponentContext: "replication-service",
    })
    if err != nil {
        return err
    }

    // Select destination registry
    dstRegistry, err := s.registryManager.SelectRegistry(ctx, dst, SelectionPolicy{
        Strategy: PreferPrivate,
        ComponentContext: "replication-service",
    })
    if err != nil {
        return err
    }

    // Use unified interface
    image, err := srcRegistry.PullImage(ctx, ParseImageReference(src))
    if err != nil {
        return err
    }

    return dstRegistry.PushImage(ctx, image)
}
```

## 5. Security Architecture

### 5.1 Credential Management

**Supported Credential Sources (in order of preference):**
1. Cloud Provider Secrets Manager (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault)
2. HashiCorp Vault
3. Environment Variables (with warnings)
4. Configuration File (encrypted, with warnings)
5. Cloud Provider Metadata Service (IAM roles, managed identities)

**Security Best Practices:**
- Never log credentials
- Encrypt credentials at rest (if stored in config)
- Use temporary credentials when possible (IAM roles, assume-role)
- Implement credential rotation
- Audit credential access
- Use least-privilege IAM policies

### 5.2 TLS/SSL Configuration

```yaml
registries:
  definitions:
    - name: secure-registry
      config:
        tlsVerify: true                    # Verify server certificates
        caCert: /path/to/ca.crt           # Custom CA certificate
        insecureSkipTLSVerify: false      # Dangerous: skip verification
        clientCert: /path/to/client.crt   # mTLS client certificate
        clientKey: /path/to/client.key    # mTLS client key
        minTLSVersion: "1.2"              # Minimum TLS version
        cipherSuites:                      # Allowed cipher suites
          - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
          - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```

### 5.3 Access Control

```go
type RegistryAccessPolicy struct {
    // IP whitelisting
    AllowedCIDRs []string

    // Rate limiting
    RateLimits RateLimitConfig

    // Audit logging
    AuditLog AuditLogConfig

    // Component-based access control
    ComponentPermissions map[string][]Permission
}
```

## 6. Migration Strategy

### 6.1 Backward Compatibility

**Existing Configuration (still works):**
```yaml
# Old format
ecr:
  region: us-west-2
  accountID: "123456789012"
gcr:
  project: my-project
```

**Automatic Migration:**
```go
func MigrateConfig(oldConfig *OldConfig) (*NewConfig, error) {
    newConfig := &NewConfig{
        Registries: RegistryConfig{
            Definitions: []RegistryDefinition{},
        },
    }

    // Migrate ECR config
    if oldConfig.ECR.AccountID != "" {
        newConfig.Registries.Definitions = append(newConfig.Registries.Definitions, RegistryDefinition{
            Name:     "ecr-migrated",
            Type:     "ecr",
            Enabled:  true,
            Priority: 1,
            Config: ECRConfig{
                Region:    oldConfig.ECR.Region,
                AccountID: oldConfig.ECR.AccountID,
            },
            Authentication: AuthConfig{
                Method: "iam-role",
            },
        })
    }

    // Similar for GCR...

    return newConfig, nil
}
```

### 6.2 Phased Rollout

**Phase 1: Internal Changes**
- Implement RegistryManager interface
- Add registry factory
- Implement authentication providers
- Add unit tests

**Phase 2: New Registry Types**
- Implement Azure ACR client
- Implement DockerHub client
- Implement Harbor client
- Implement generic OCI client
- Add integration tests

**Phase 3: Configuration Migration**
- Add new configuration schema
- Implement backward compatibility layer
- Add configuration validation
- Document migration guide

**Phase 4: Component Integration**
- Update replication service
- Update server handlers
- Update CLI commands
- Update Kubernetes manifests

**Phase 5: Advanced Features**
- Add registry-to-registry replication
- Add multi-registry failover
- Add credential rotation
- Add monitoring and alerting

## 7. Testing Strategy

### 7.1 Unit Tests
- Registry manager initialization
- Authentication provider logic
- Registry selection algorithm
- Configuration parsing and validation
- Secrets provider implementations

### 7.2 Integration Tests
- End-to-end authentication flows
- Cross-registry replication
- Failover scenarios
- Credential rotation
- TLS/mTLS connections

### 7.3 Performance Tests
- Connection pool efficiency
- Credential caching performance
- Registry selection speed
- Concurrent authentication requests

## 8. Monitoring and Observability

### 8.1 Metrics

```go
// Registry-specific metrics
registry_authentication_attempts_total{registry="harbor-private",method="username-password",status="success"} 145
registry_authentication_attempts_total{registry="harbor-private",method="username-password",status="failure"} 3
registry_authentication_duration_seconds{registry="harbor-private"} 0.234
registry_operations_total{registry="aws-ecr-primary",operation="push",status="success"} 1234
registry_operations_total{registry="aws-ecr-primary",operation="pull",status="success"} 5678
registry_connection_pool_active{registry="harbor-private"} 8
registry_connection_pool_idle{registry="harbor-private"} 2
registry_credential_refresh_total{registry="aws-ecr-primary",status="success"} 24
registry_selection_duration_seconds{component="replication-service"} 0.012
```

### 8.2 Logging

```json
{
  "level": "info",
  "timestamp": "2025-12-02T10:00:00Z",
  "component": "registry-manager",
  "registry": "harbor-private",
  "operation": "authenticate",
  "method": "username-password",
  "duration_ms": 234,
  "status": "success"
}
```

### 8.3 Health Checks

```yaml
# Registry health check endpoint
GET /api/v1/registries/health
Response:
{
  "registries": [
    {
      "name": "aws-ecr-primary",
      "type": "ecr",
      "status": "healthy",
      "lastCheck": "2025-12-02T10:00:00Z",
      "responseTime": "123ms"
    },
    {
      "name": "harbor-private",
      "type": "harbor",
      "status": "unhealthy",
      "lastCheck": "2025-12-02T10:00:00Z",
      "error": "connection timeout"
    }
  ]
}
```

## 9. Implementation Roadmap

### 9.1 Timeline

**Week 1-2: Core Infrastructure**
- [ ] Design and implement RegistryManager interface
- [ ] Create registry factory pattern
- [ ] Implement configuration schema and validation
- [ ] Add backward compatibility layer
- [ ] Write unit tests

**Week 3-4: Authentication System**
- [ ] Implement authentication provider interface
- [ ] Add AWS IAM authentication
- [ ] Add GCP service account authentication
- [ ] Add Azure managed identity authentication
- [ ] Add basic auth, token auth, mTLS
- [ ] Implement secrets providers (AWS, GCP, Azure, Vault)
- [ ] Add credential caching and rotation

**Week 5-6: Registry Clients**
- [ ] Refactor existing ECR and GCR clients
- [ ] Implement Azure ACR client
- [ ] Implement DockerHub client
- [ ] Implement Harbor client
- [ ] Implement Quay.io client
- [ ] Implement generic OCI client
- [ ] Add integration tests

**Week 7-8: Component Integration**
- [ ] Update replication service
- [ ] Update server HTTP handlers
- [ ] Update CLI commands
- [ ] Add registry selection logic
- [ ] Update Kubernetes manifests
- [ ] Add end-to-end tests

**Week 9-10: Advanced Features**
- [ ] Implement registry-to-registry replication
- [ ] Add multi-registry failover
- [ ] Add monitoring and alerting
- [ ] Performance optimization
- [ ] Documentation and examples

## 10. Success Criteria

### 10.1 Functional Requirements
- [x] Support at least 5 registry types (ECR, GCR, ACR, DockerHub, Harbor)
- [x] Support multiple authentication methods per registry type
- [x] Implement secrets manager integration
- [x] Provide component-level registry overrides
- [x] Maintain backward compatibility
- [x] Pass all unit and integration tests

### 10.2 Non-Functional Requirements
- [ ] Registry selection in < 50ms
- [ ] Authentication caching reduces auth calls by 90%
- [ ] Zero downtime during registry failover
- [ ] Support 100+ concurrent registry operations
- [ ] Comprehensive documentation with examples
- [ ] 90%+ code coverage

## 11. Related Documents

- [Configuration Schema Reference](./registry-config-schema.md)
- [Component Modification Guide](./component-modifications.md)
- [Security Best Practices](./security-guidelines.md)
- [Migration Guide](./migration-guide.md)
- [API Reference](./api-reference.md)

---

**Document Version:** 1.0
**Last Updated:** 2025-12-02
**Author:** System Architect
**Status:** Design Review
