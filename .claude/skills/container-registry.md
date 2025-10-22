# Container Registry Operations Skill

Expert skill for working with container registries (ECR, GCR, ACR) in the Freightliner project.

## Capabilities

- Authenticate with AWS ECR and Google GCR
- List and inspect images and repositories
- Perform replication operations
- Debug registry connectivity issues
- Manage image signatures and encryption

## Authentication Patterns

### AWS ECR
```go
// Standard authentication flow
cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
ecrClient := ecr.NewFromConfig(cfg)

// With assumed role
cfg, err := config.LoadDefaultConfig(ctx,
    config.WithRegion(region),
    config.WithAssumeRoleCredentialOptions(func(o *stscreds.AssumeRoleOptions) {
        o.RoleARN = roleARN
    }),
)
```

### Google GCR
```go
// Using application default credentials
ctx := context.Background()
creds, err := google.FindDefaultCredentials(ctx, containerregistry.CloudPlatformScope)

// Using service account key
creds, err := google.CredentialsFromJSON(ctx, keyJSON, containerregistry.CloudPlatformScope)
```

## Common Operations

### List Repositories
```bash
# ECR
aws ecr describe-repositories --region us-east-1

# GCR
gcloud container images list --repository=gcr.io/my-project
```

### List Image Tags
```bash
# ECR
aws ecr list-images --repository-name my-repo --region us-east-1

# GCR
gcloud container images list-tags gcr.io/my-project/my-repo
```

### Pull and Inspect Image
```bash
# Pull image
docker pull 123456789.dkr.ecr.us-east-1.amazonaws.com/my-repo:latest

# Inspect image
docker inspect 123456789.dkr.ecr.us-east-1.amazonaws.com/my-repo:latest

# Check manifest
docker manifest inspect 123456789.dkr.ecr.us-east-1.amazonaws.com/my-repo:latest
```

## Debugging Common Issues

### ECR Authentication Failures
1. Check AWS credentials: `aws sts get-caller-identity`
2. Verify IAM permissions for ECR operations
3. Check ECR repository policy
4. Verify region configuration

### GCR Authentication Failures
1. Check credentials: `gcloud auth list`
2. Verify service account permissions
3. Check project ID configuration
4. Verify API is enabled: `gcloud services list --enabled`

### Network Connectivity
```bash
# Test ECR endpoint
curl -I https://123456789.dkr.ecr.us-east-1.amazonaws.com/v2/

# Test GCR endpoint
curl -I https://gcr.io/v2/
```

## Freightliner-Specific

### Registry Client Interface
Located in `pkg/interfaces/interfaces.go`:
```go
type RegistryClient interface {
    ListRepositories(ctx context.Context, prefix string) ([]string, error)
    GetRepository(ctx context.Context, name string) (Repository, error)
    GetRegistryName() string
}
```

### ECR Client Implementation
Located in `pkg/client/ecr/client.go`:
- Authentication with credential helper
- Repository listing and filtering
- Image descriptor handling
- Multi-architecture support

### GCR Client Implementation
Located in `pkg/client/gcr/client.go`:
- Google Cloud authentication
- Artifact Registry API integration
- Service account credential handling
- Project-based repository access

## Best Practices

1. **Always use context.Context** for cancellation and timeouts
2. **Implement retry logic** with exponential backoff for transient failures
3. **Use credential providers** instead of hardcoded credentials
4. **Log operations** with structured logging for debugging
5. **Handle rate limits** with proper backoff strategies
6. **Verify checksums** for image integrity
7. **Use streaming** for large image transfers to avoid memory issues

## Error Handling

```go
// Wrap errors with context
if err != nil {
    return errors.Wrap(err, "failed to list repositories")
}

// Use domain-specific errors
if notFound {
    return errors.NotFoundf("repository %s not found", name)
}

// Check for specific error types
var authErr *AuthenticationError
if errors.As(err, &authErr) {
    // Handle authentication error
}
```
