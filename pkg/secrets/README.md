# Secrets Management Package

The secrets management package provides integration with cloud provider secrets managers for securely storing and retrieving sensitive credentials and configuration.

## Supported Providers

- **AWS Secrets Manager**: For storing and retrieving secrets in AWS
- **Google Secret Manager**: For storing and retrieving secrets in Google Cloud

## Key Features

- Unified interface for multiple secret manager providers
- JSON serialization/deserialization for structured secrets
- Secure handling of registry credentials, encryption keys, and signing materials
- Support for listing and managing secrets

## Usage Examples

### Initializing a Secrets Provider

```go
import (
    "context"
    "helper/log"
    "pkg/secrets"
)

// Create a logger
logger := log.NewLogger(log.InfoLevel)

// Create a context
ctx := context.Background()

// Initialize an AWS Secrets Manager provider
awsOpts := secrets.ManagerOptions{
    Provider:  secrets.AWSProvider,
    Logger:    logger,
    AWSRegion: "us-west-2",
}

awsProvider, err := secrets.GetProvider(ctx, awsOpts)
if err != nil {
    // Handle error
}

// Initialize a Google Secret Manager provider
gcpOpts := secrets.ManagerOptions{
    Provider:           secrets.GCPProvider,
    Logger:             logger,
    GCPProject:         "my-project",
    GCPCredentialsFile: "/path/to/credentials.json",
}

gcpProvider, err := secrets.GetProvider(ctx, gcpOpts)
if err != nil {
    // Handle error
}
```

### Storing and Retrieving Simple Secrets

```go
// Store a simple secret
err := provider.PutSecret(ctx, "api-key", "secret-value")
if err != nil {
    // Handle error
}

// Retrieve a simple secret
value, err := provider.GetSecret(ctx, "api-key")
if err != nil {
    // Handle error
}
```

### Working with Structured JSON Secrets

```go
// Define a struct for your credentials
type DBCredentials struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Host     string `json:"host"`
    Port     int    `json:"port"`
}

// Create the credentials
creds := DBCredentials{
    Username: "admin",
    Password: "secure-password",
    Host:     "db.example.com",
    Port:     5432,
}

// Store the credentials as a JSON secret
err := provider.PutJSONSecret(ctx, "database-credentials", creds)
if err != nil {
    // Handle error
}

// Retrieve and unmarshal the credentials
var retrievedCreds DBCredentials
err = provider.GetJSONSecret(ctx, "database-credentials", &retrievedCreds)
if err != nil {
    // Handle error
}
```

### Deleting Secrets

```go
// Delete a secret
err := provider.DeleteSecret(ctx, "api-key")
if err != nil {
    // Handle error
}
```

## Command-Line Usage

Freightliner supports using cloud provider secrets managers directly from the command line:

```bash
# Using AWS Secrets Manager
freightliner replicate ecr/my-repository gcr/my-repository \
  --use-secrets-manager \
  --secrets-manager-type=aws \
  --aws-secret-region=us-west-2 \
  --registry-creds-secret=freightliner-registry-credentials

# Using Google Secret Manager
freightliner replicate ecr/my-repository gcr/my-repository \
  --use-secrets-manager \
  --secrets-manager-type=gcp \
  --gcp-secret-project=my-project \
  --registry-creds-secret=freightliner-registry-credentials
```

## Secret Format

### Registry Credentials

Registry credentials should be stored in the following JSON format:

```json
{
  "ecr": {
    "access_key": "AWS_ACCESS_KEY_ID",
    "secret_key": "AWS_SECRET_ACCESS_KEY",
    "account_id": "012345678901",
    "region": "us-west-2",
    "session_token": "OPTIONAL_SESSION_TOKEN"
  },
  "gcr": {
    "project": "my-project",
    "location": "us",
    "credentials": "BASE64_ENCODED_SERVICE_ACCOUNT_JSON"
  }
}
```

### Encryption Keys

Encryption keys should be stored in the following JSON format:

```json
{
  "aws": {
    "kms_key_id": "alias/my-key-alias",
    "region": "us-west-2"
  },
  "gcp": {
    "kms_key_id": "projects/my-project/locations/global/keyRings/freightliner/cryptoKeys/image-encryption",
    "project": "my-project",
    "location": "global",
    "key_ring": "freightliner",
    "key": "image-encryption"
  }
}
```

### Signing Keys

Signing keys should be stored in the following JSON format:

```json
{
  "key_path": "/path/to/key/file",
  "key_id": "key-identifier",
  "key_data": "BASE64_ENCODED_KEY_DATA"
}
```

If `key_data` is provided, it will be decoded and written to a temporary file for use during the operation.