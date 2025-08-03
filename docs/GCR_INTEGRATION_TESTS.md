# GCR Integration Tests

## Overview

The GCR (Google Container Registry) client tests include both unit tests and integration tests. Integration tests require valid GCP credentials and are disabled by default to prevent CI pipeline failures.

## Running Tests

### Unit Tests (Default)
```bash
go test ./pkg/client/gcr/... -v
```

Unit tests use mocks and do not require GCP credentials. These tests run by default in CI pipelines.

### Integration Tests
```bash
ENABLE_GCR_INTEGRATION_TESTS=true go test ./pkg/client/gcr/... -v
```

Integration tests make real API calls to GCR and require:
- Valid GCP credentials configured via Application Default Credentials (ADC)
- Access to a GCP project with Container Registry API enabled
- Appropriate IAM permissions

## Test Types

### Unit Tests (Always Run)
- `TestGCRAuthenticatorAuthorization` - Tests authentication logic with mocks
- `TestGCRKeychainResolve` - Tests keychain resolution with mocks
- `TestGCRTransport` - Tests HTTP transport with mocks
- `TestNewClient` - Tests client creation (may show credential warnings but passes)
- `TestClientGetRepository` - Tests repository reference creation
- All parser and utility function tests

### Integration Tests (Require `ENABLE_GCR_INTEGRATION_TESTS=true`)
- `TestClientListRepositories` - Makes real API calls to list repositories
- `TestGCRClientWithDifferentRegistryTypes` - Creates real GCP clients

## Skip Messages

When integration tests are disabled, you'll see skip messages like:
```
GCR integration tests disabled. Set ENABLE_GCR_INTEGRATION_TESTS=true to enable
```

This is expected behavior and indicates the integration tests are properly configured.

## Credential Warnings

You may see warnings like:
```
Failed to create Artifact Registry client, some functionality may be limited
error=credentials: could not find default credentials
```

These warnings are expected when running without GCP credentials and do not cause test failures.