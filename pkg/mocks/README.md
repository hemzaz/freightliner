# Interface Mocks

This directory contains automatically generated mocks for all interfaces in the Freightliner project.

## Generating Mocks

To generate all mocks, run from the project root:

```bash
make generate-mocks
```

Or generate mocks for specific packages:

```bash
# Generate interface mocks
cd pkg/interfaces && go generate ./...

# Generate copy mocks
cd pkg/copy && go generate ./...

# Generate service mocks
cd pkg/service && go generate ./...
```

## Mock Usage in Tests

The mocks are designed following the Interface Segregation Principle, making tests more focused and maintainable:

```go
package mytest

import (
    "testing"
    "freightliner/pkg/mocks"
    "github.com/golang/mock/gomock"
)

func TestMyFunction(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Use focused mock instead of large Repository mock
    mockReader := mocks.NewMockReader(ctrl)
    mockReader.EXPECT().GetRepositoryName().Return("test-repo")
    
    // Test your function with the mock
    result := myFunction(mockReader)
    // ... assertions
}
```

## Interface Segregation Benefits

The segregated interfaces provide several testing benefits:

1. **Focused Mocks**: Test only the methods your code actually uses
2. **Better Error Messages**: Mock failures are more specific
3. **Easier Setup**: Less boilerplate for mock expectations
4. **Composition Testing**: Test different combinations of behaviors

## Mock Categories

### Repository Mocks
- `Reader`, `Writer`, `ImageProvider`, `MetadataProvider`
- `ContentProvider`, `ContentManager`
- Composition mocks: `ReadWriteRepository`, `ImageRepository`, `FullRepository`

### Authentication Mocks
- `TokenProvider`, `HeaderProvider`, `AuthenticatorProvider`
- `TokenManager`, `CachingAuthenticator`
- `MultiRegistryAuthenticator`

### Client Mocks
- `RepositoryLister`, `RepositoryProvider`, `RegistryInfo`
- `PaginatedRepositoryLister`, `CachingRepositoryProvider`
- `BatchRepositoryProvider`, `HealthChecker`

### Copy Mocks
- `SourceReader`, `DestinationWriter`
- `ProgressReporter`, `LayerProcessor`, `ManifestProcessor`
- `ContextualCopier`, `StreamingCopier`

### Service Mocks
- `ReplicationService`, `MonitoringService`, `HealthService`
- Service composition mocks

## Legacy Interface Support

Legacy interfaces (large, monolithic) are still supported for backward compatibility:
- `Repository`, `RegistryClient`, `RegistryAuthenticator`

However, new tests should prefer the segregated interfaces for better maintainability.