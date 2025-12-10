# Registry Configuration Guide

This guide explains how to configure Freightliner to work with various container registries, including AWS ECR, Google GCR, and custom/third-party registries.

## Supported Registry Combinations

Freightliner supports replication between **ANY** combination of registries:

| Source → Destination | Status |
|---------------------|--------|
| ECR ↔ GCR | ✅ Supported |
| Local ↔ ECR | ✅ Supported |
| Local ↔ GCR | ✅ Supported |
| Custom ↔ ECR | ✅ Supported |
| Custom ↔ GCR | ✅ Supported |
| Custom ↔ Custom | ✅ Supported |
| Local ↔ Local | ✅ Supported |
| ECR ↔ Local | ✅ Supported |
| GCR ↔ Local | ✅ Supported |

**Custom registries include**: Harbor, Quay.io, GitLab Registry, GitHub Container Registry (GHCR), Azure ACR, Artifactory, Docker Hub, and any Docker Registry v2 compatible registry.

## Table of Contents

- [Configuration Methods](#configuration-methods)
- [Registry Types](#registry-types)
- [Authentication Methods](#authentication-methods)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)

## Configuration Methods

### Method 1: Configuration File (config.yaml)

The primary configuration method uses a YAML file:

```yaml
# General configuration
logLevel: info

# AWS ECR configuration (built-in)
ecr:
  region: us-east-1
  accountID: "123456789012"

# GCP GCR configuration (built-in)
gcr:
  project: my-gcp-project
  location: us

# Custom registries configuration
registries:
  # Default source registry for pulling images
  defaultSource: "ecr-prod"

  # Default destination registry for pushing images
  defaultDestination: "harbor-prod"

  # List of configured registries
  registries:
    # AWS ECR Production
    - name: ecr-prod
      type: ecr
      region: us-east-1
      accountID: "123456789012"
      auth:
        type: aws
      metadata:
        environment: production

    # Harbor Private Registry
    - name: harbor-prod
      type: harbor
      endpoint: harbor.example.com
      auth:
        type: basic
        username: admin
        password: ${HARBOR_PASSWORD}
      tls:
        insecureSkipVerify: false
        caFile: /etc/ssl/certs/harbor-ca.pem

    # Local Development Registry
    - name: local-dev
      type: generic
      endpoint: localhost:5000
      auth:
        type: anonymous
      insecure: true

    # Docker Hub
    - name: dockerhub
      type: dockerhub
      endpoint: registry-1.docker.io
      auth:
        type: basic
        username: myusername
        password: ${DOCKERHUB_TOKEN}

    # GitHub Container Registry
    - name: ghcr
      type: github
      endpoint: ghcr.io
      auth:
        type: token
        token: ${GITHUB_TOKEN}
```

### Method 2: Environment Variables

Override configuration with environment variables:

```bash
# General
export FREIGHTLINER_LOG_LEVEL=debug

# ECR
export FREIGHTLINER_ECR_REGION=us-west-2
export FREIGHTLINER_ECR_ACCOUNT_ID=123456789012

# GCR
export FREIGHTLINER_GCR_PROJECT=my-project
export FREIGHTLINER_GCR_LOCATION=us

# Workers
export FREIGHTLINER_REPLICATE_WORKERS=8
```

### Method 3: Command Line Flags

Pass configuration via CLI flags:

```bash
freightliner replicate ecr/my-repo gcr/my-repo \
  --ecr-region us-east-1 \
  --gcr-project my-project \
  --replicate-workers 4
```

## Registry Types

### Built-in Registry Types

#### 1. AWS ECR (Elastic Container Registry)

```yaml
- name: ecr-main
  type: ecr
  region: us-east-1
  accountID: "123456789012"
  auth:
    type: aws
    # Optional: Assume role
    roleARN: "arn:aws:iam::123456789012:role/FreightlinerRole"
  metadata:
    environment: production
```

#### 2. Google GCR (Container Registry)

```yaml
- name: gcr-main
  type: gcr
  project: my-gcp-project
  region: us-central1
  auth:
    type: gcp
    credentialsFile: /etc/gcp/service-account.json
  metadata:
    environment: production
```

### Generic/Custom Registry Types

All of these use Docker Registry v2 API and are configured the same way:

#### 3. Generic Docker Registry

```yaml
- name: my-registry
  type: generic
  endpoint: registry.example.com
  auth:
    type: basic
    username: admin
    password: ${REGISTRY_PASSWORD}
  tls:
    insecureSkipVerify: false
    caFile: /etc/ssl/certs/registry-ca.pem
```

#### 4. Harbor Registry

```yaml
- name: harbor
  type: harbor
  endpoint: harbor.example.com
  auth:
    type: basic
    username: admin
    password: ${HARBOR_PASSWORD}
  tls:
    insecureSkipVerify: false
```

#### 5. Quay.io

```yaml
- name: quay
  type: quay
  endpoint: quay.io
  auth:
    type: token
    token: ${QUAY_TOKEN}
  metadata:
    organization: myorg
```

#### 6. GitLab Container Registry

```yaml
- name: gitlab
  type: gitlab
  endpoint: registry.gitlab.com
  auth:
    type: token
    token: ${GITLAB_TOKEN}
  metadata:
    project: mygroup/myproject
```

#### 7. GitHub Container Registry (GHCR)

```yaml
- name: ghcr
  type: github
  endpoint: ghcr.io
  auth:
    type: basic
    username: myusername
    token: ${GITHUB_TOKEN}
```

#### 8. Docker Hub

```yaml
- name: dockerhub
  type: dockerhub
  endpoint: registry-1.docker.io
  auth:
    type: basic
    username: myusername
    password: ${DOCKERHUB_ACCESS_TOKEN}
```

#### 9. Azure Container Registry (ACR)

```yaml
- name: acr
  type: azure
  endpoint: myregistry.azurecr.io
  auth:
    type: basic
    username: myregistry
    password: ${ACR_PASSWORD}
  metadata:
    subscription: azure-subscription-id
```

#### 10. Artifactory

```yaml
- name: artifactory
  type: generic
  endpoint: artifactory.example.com
  auth:
    type: basic
    username: admin
    password: ${ARTIFACTORY_PASSWORD}
  tls:
    certFile: /etc/ssl/certs/artifactory-client.pem
    keyFile: /etc/ssl/private/artifactory-client-key.pem
    caFile: /etc/ssl/certs/artifactory-ca.pem
```

#### 11. Local Development Registry

```yaml
- name: local-dev
  type: generic
  endpoint: localhost:5000
  auth:
    type: anonymous
  insecure: true
```

## Authentication Methods

### 1. Anonymous (No Authentication)

```yaml
auth:
  type: anonymous
```

### 2. Basic Authentication (Username/Password)

```yaml
auth:
  type: basic
  username: myuser
  password: mypassword
  # Or use environment variables
  password: ${REGISTRY_PASSWORD}
```

### 3. Token/Bearer Authentication

```yaml
auth:
  type: token
  token: ${REGISTRY_TOKEN}
```

### 4. AWS IAM Authentication

```yaml
auth:
  type: aws
  # Optional: Assume a specific role
  roleARN: "arn:aws:iam::123456789012:role/FreightlinerRole"
```

### 5. GCP Service Account

```yaml
auth:
  type: gcp
  credentialsFile: /path/to/service-account.json
```

## Usage Examples

### Example 1: ECR to GCR Replication

```bash
# Using built-in registry names
freightliner replicate ecr/my-app gcr/my-app

# Or with tags
freightliner replicate ecr/my-app:v1.0.0 gcr/my-app:v1.0.0
```

### Example 2: Local to ECR

```yaml
# Config file
registries:
  registries:
    - name: local-dev
      type: generic
      endpoint: localhost:5000
      auth:
        type: anonymous
      insecure: true
```

```bash
# Replicate
freightliner replicate local-dev/test-app ecr/test-app
```

### Example 3: Harbor to Quay

```yaml
registries:
  registries:
    - name: harbor-prod
      type: harbor
      endpoint: harbor.example.com
      auth:
        type: basic
        username: admin
        password: ${HARBOR_PASSWORD}

    - name: quay-backup
      type: quay
      endpoint: quay.io
      auth:
        type: token
        token: ${QUAY_TOKEN}
```

```bash
# Replicate entire repository
freightliner replicate harbor-prod/myapp quay-backup/myapp

# Tree replication (all repos)
freightliner tree-replicate harbor-prod quay-backup
```

### Example 4: GitLab to Azure ACR

```yaml
registries:
  registries:
    - name: gitlab
      type: gitlab
      endpoint: registry.gitlab.com
      auth:
        type: token
        token: ${GITLAB_TOKEN}

    - name: azure
      type: azure
      endpoint: myregistry.azurecr.io
      auth:
        type: basic
        username: myregistry
        password: ${ACR_PASSWORD}
```

```bash
freightliner replicate gitlab/mygroup/myapp azure/myapp
```

### Example 5: Local to Local (Different Ports)

```yaml
registries:
  registries:
    - name: local-dev
      type: generic
      endpoint: localhost:5000
      auth:
        type: anonymous
      insecure: true

    - name: local-staging
      type: generic
      endpoint: localhost:5001
      auth:
        type: anonymous
      insecure: true
```

```bash
freightliner replicate local-dev/app local-staging/app
```

### Example 6: Docker Hub to GCR

```yaml
registries:
  registries:
    - name: dockerhub
      type: dockerhub
      endpoint: registry-1.docker.io
      auth:
        type: basic
        username: myuser
        password: ${DOCKERHUB_TOKEN}
```

```bash
freightliner replicate dockerhub/library/nginx gcr/nginx
```

### Example 7: Custom to Custom (Harbor to Artifactory)

```yaml
registries:
  registries:
    - name: harbor
      type: harbor
      endpoint: harbor.corp.com
      auth:
        type: basic
        username: admin
        password: ${HARBOR_PASS}

    - name: artifactory
      type: generic
      endpoint: artifactory.corp.com
      auth:
        type: basic
        username: admin
        password: ${ARTIFACTORY_PASS}
```

```bash
freightliner replicate harbor/production/app artifactory/backup/app
```

## TLS Configuration

### Secure Registry with Custom CA

```yaml
- name: secure-registry
  type: generic
  endpoint: registry.secure.com
  auth:
    type: basic
    username: admin
    password: ${PASSWORD}
  tls:
    insecureSkipVerify: false
    caFile: /etc/ssl/certs/custom-ca.pem
```

### Mutual TLS (mTLS)

```yaml
- name: mtls-registry
  type: generic
  endpoint: registry.secure.com
  auth:
    type: basic
    username: admin
    password: ${PASSWORD}
  tls:
    certFile: /etc/ssl/certs/client.pem
    keyFile: /etc/ssl/private/client-key.pem
    caFile: /etc/ssl/certs/ca.pem
```

### Insecure Registry (Development Only)

```yaml
- name: insecure-dev
  type: generic
  endpoint: localhost:5000
  auth:
    type: anonymous
  insecure: true
```

## Environment Variable Expansion

Use environment variables in your configuration:

```yaml
registries:
  registries:
    - name: secure-registry
      type: generic
      endpoint: ${REGISTRY_ENDPOINT}
      auth:
        type: basic
        username: ${REGISTRY_USER}
        password: ${REGISTRY_PASS}
```

Supported formats:
- `${VAR_NAME}` - Standard format
- Environment variables are expanded at runtime
- Useful for secrets and dynamic configuration

## Troubleshooting

### Connection Issues

**Problem**: Cannot connect to registry

```bash
# Test connectivity
curl https://registry.example.com/v2/

# Check DNS resolution
nslookup registry.example.com

# Test with insecure flag (dev only)
freightliner replicate source dest --config /path/to/config.yaml
```

**Solution**:
- Verify endpoint URL is correct
- Check firewall rules
- For self-signed certificates, add CA certificate or use `insecure: true` (dev only)

### Authentication Failures

**Problem**: 401 Unauthorized errors

```yaml
# Verify credentials are correct
auth:
  type: basic
  username: ${USER}
  password: ${PASS}
```

**Solution**:
- Verify environment variables are set: `echo $REGISTRY_PASSWORD`
- Check token hasn't expired
- For cloud registries, ensure IAM permissions are correct

### TLS/Certificate Issues

**Problem**: x509: certificate signed by unknown authority

**Solution**:
```yaml
# Option 1: Add custom CA (recommended)
tls:
  caFile: /etc/ssl/certs/custom-ca.pem

# Option 2: Skip verification (dev only)
insecure: true
```

### Registry Not Found

**Problem**: `registry 'xyz' not found in configuration`

**Solution**:
- Verify registry name in `config.yaml` matches exactly
- Registry names are case-sensitive
- Check YAML indentation is correct

## Security Best Practices

### 1. Never Hardcode Credentials

❌ **Bad:**
```yaml
auth:
  username: admin
  password: supersecret123
```

✅ **Good:**
```yaml
auth:
  username: ${REGISTRY_USER}
  password: ${REGISTRY_PASS}
```

### 2. Use Secrets Manager (Production)

```yaml
secrets:
  useSecretsManager: true
  secretsManagerType: aws
  registryCredsSecret: freightliner-registry-credentials
```

### 3. Limit Registry Permissions

- Use read-only credentials for source registries
- Use write-only credentials for destination registries
- Create dedicated service accounts with minimal permissions

### 4. Enable TLS in Production

```yaml
tls:
  insecureSkipVerify: false  # Always false in production
  caFile: /etc/ssl/certs/ca.pem
```

### 5. Rotate Credentials Regularly

- Set up credential rotation policies
- Use short-lived tokens when possible
- Monitor for unauthorized access

### 6. Network Security

- Use private networks when possible
- Configure firewall rules to restrict access
- Consider using VPN or private endpoints

## Advanced Configuration

### Default Source and Destination

```yaml
registries:
  defaultSource: "dockerhub"
  defaultDestination: "harbor-prod"
```

This allows shorter commands:
```bash
# Without defaults
freightliner replicate dockerhub/nginx harbor-prod/nginx

# With defaults (if source/dest match defaults)
freightliner replicate nginx nginx
```

### Metadata Tags

Add metadata to track registries:

```yaml
- name: prod-registry
  type: harbor
  endpoint: harbor.prod.com
  metadata:
    environment: production
    team: platform
    cost-center: engineering
```

### Multiple Regions

Configure the same registry type for multiple regions:

```yaml
registries:
  registries:
    - name: ecr-us-east-1
      type: ecr
      region: us-east-1
      accountID: "123456789012"

    - name: ecr-eu-west-1
      type: ecr
      region: eu-west-1
      accountID: "123456789012"
```

## Complete Configuration Example

See `examples/config-with-registries.yaml` for a comprehensive configuration file demonstrating all registry types and authentication methods.

## Further Reading

- [AWS ECR Documentation](https://docs.aws.amazon.com/ecr/)
- [Google GCR Documentation](https://cloud.google.com/container-registry/docs)
- [Docker Registry V2 API](https://docs.docker.com/registry/spec/api/)
- [Harbor Documentation](https://goharbor.io/docs/)
- [Quay.io Documentation](https://docs.quay.io/)
