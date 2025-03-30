# Placeholder Implementations in Freightliner

This document lists all incomplete and placeholder implementations found in the codebase, organized by priority level.

## Critical Priority

1. **GCR Repository Image Deletion** 
   - **File**: `/pkg/client/gcr/repository.go`
   - **Function**: `DeleteImage`
   - **Issue**: Image deletion is not implemented for GCR
   ```go
   return errors.NotImplementedf("image deletion not implemented for GCR")
   ```
   - **Impact**: Critical functionality missing for cleaning up repositories

2. **ECR Cross-Region Authentication**
   - **File**: `/pkg/client/ecr/auth.go`
   - **Function**: `RegistryAuthenticator`
   - **Issue**: Authentication for cross-region ECR registries not implemented
   ```go
   return nil, errors.NotImplementedf("authentication for cross-region ECR registries not implemented")
   ```
   - **Impact**: Prevents cross-region replication for ECR

3. **Scheduled Replication**
   - **File**: `/pkg/replication/scheduler.go`
   - **Functions**: `checkJobs` and `submitJob` 
   - **Issue**: Missing cron expression parsing and actual replication logic implementation
   ```go
   // TODO: Parse the schedule as a cron expression
   // For now, just schedule it to run in 5 minutes

   // TODO: Calculate the next run time based on the cron expression
   // For now, just schedule it to run again in 1 hour

   // TODO: Implement the actual replication logic
   // This would call into a ReplicationService or similar
   ```
   - **Impact**: Scheduled replication is non-functional

## High Priority

1. **GCR Repository Listing**
   - **File**: `/pkg/client/gcr/client.go`
   - **Function**: `ListRepositories`
   - **Issue**: Currently returns hardcoded mock repositories
   ```go
   // For testing purposes, we'll just create a mock list of repositories
   var mockRepos = []string{"repo1", "repo2", "testing/repo3", "testing/repo4"}
   var repositories []string

   // In a real implementation, we would call google.List, but the API has changed
   // so for this test we're using a mock
   ```
   - **Impact**: Prevents real repository discovery in GCR

2. **Secrets Management**
   - **File**: `/main.go`
   - **Functions**: `initializeSecretsManager`, `loadRegistryCredentials`, `loadEncryptionKeys` 
   - **Issue**: Not implemented, preventing use of cloud provider secrets management
   ```go
   func initializeSecretsManager(ctx context.Context, logger *log.Logger) (SecretsProvider, error) {
       // This would be implemented to create the appropriate secrets manager client
       // For now, return a stub that just logs
       logger.Info("Secrets manager initialization not fully implemented", nil)
       return nil, errors.NotImplementedf("secrets manager initialization")
   }

   func loadRegistryCredentials(ctx context.Context, provider SecretsProvider) (RegistryCredentials, error) {
       // This would be implemented to load credentials from the secrets manager
       return RegistryCredentials{}, errors.NotImplementedf("registry credentials loading")
   }

   func loadEncryptionKeys(ctx context.Context, provider SecretsProvider) (EncryptionKeys, error) {
       // This would be implemented to load encryption keys from the secrets manager
       return EncryptionKeys{}, errors.NotImplementedf("encryption keys loading")
   }
   ```
   - **Impact**: Forces use of less secure credential management

3. **Copier and TreeReplicator Implementation**
   - **File**: `/main.go`
   - **Functions**: `createCopier`, `createTreeReplicator`, and the stub implementation classes
   - **Issue**: Using temporary stub implementations instead of real functionality
   ```go
   func createCopier(ctx context.Context, source, dest common.Repository, encManager *encryption.Manager, logger *log.Logger, workers int) (Copier, error) {
       // This would be implemented to create a new copier
       // For now, return a stub implementation
       return &stubCopier{
           source:      source,
           destination: dest,
           logger:      logger,
           workers:     workers,
       }, nil
   }

   func createTreeReplicator(ctx context.Context, source common.RegistryClient, dest common.RegistryClient, sourcePath, destPath string, logger *log.Logger, opts map[string]interface{}) (TreeReplicator, error) {
       // This would be implemented to create a new tree replicator
       // For now, return a stub implementation
       return &stubTreeReplicator{
           sourceClient: source,
           destClient:   dest,
           sourcePath:   sourcePath,
           destPath:     destPath,
           logger:       logger,
           options:      opts,
       }, nil
   }
   ```
   - **Impact**: Core replication functionality may be incomplete or suboptimal

## Medium Priority

1. **Server Mode**
   - **File**: `/main.go`
   - **Function**: `serveCmd`
   - **Issue**: Server mode not implemented, prevents API-based usage
   ```go
   fmt.Println("Server mode not yet implemented")
   ```
   - **Impact**: Limits deployment options to CLI-only

2. **Checkpoint Management**
   - **File**: `/main.go`
   - **Functions**: `checkpointListCmd`, `checkpointShowCmd`, `checkpointDeleteCmd`
   - **Issue**: Checkpoint management commands not implemented
   ```go
   fmt.Println("Checkpoint listing not yet implemented")
   fmt.Println("Checkpoint details not yet implemented")
   fmt.Println("Checkpoint deletion not yet implemented")
   ```
   - **Impact**: Users can't inspect or manage replication checkpoints

3. **Base Client Methods**
   - **File**: `/pkg/client/common/client.go`
   - **Functions**: `ListRepositories`, `GetRepository`
   - **Issue**: Base client methods not implemented
   ```go
   return nil, errors.NotImplementedf("method not implemented in base client")
   ```
   - **Impact**: Prevents fallback implementations

## Low Priority

1. **GCR Repository ListTags Implementation**
   - **File**: `/pkg/client/gcr/repository.go`
   - **Function**: `ListTags`
   - **Comment**: Uses basic implementation but could be enhanced for better pagination and error handling
   ```go
   // In a real implementation, this would use google.List or the GCR API
   // For now, using a simulated implementation
   ```
   - **Impact**: May not handle all edge cases efficiently

2. **Mock Image Implementations**
   - **File**: `/pkg/client/gcr/repository.go`
   - **Type**: `mockRemoteImage`
   - **Issue**: Minimal implementation of the v1.Image interface, may not support all operations
   ```go
   type mockRemoteImage struct {
       manifestBytes []byte
       mediaType     types.MediaType
   }
   // ... with minimal method implementations
   ```
   - **Impact**: May cause unexpected behavior in edge cases

3. **Network Delta Optimization**
   - **File**: `/pkg/network/delta.go`
   - **Function**: `GenerateManifest`
   - **Issue**: Missing real-world optimizations for better performance
   ```go
   // This is just a sample implementation - real code would:
   // 1. List tags in source and destination
   // 2. For each tag in source that needs to be copied:
   //    a. Check if it exists in destination
   //    b. If not, mark for full copy
   //    c. If yes, compare layers and mark changed ones for delta copy
   // 3. Generate the manifest with this information
   ```
   - **Impact**: Suboptimal network transfer efficiency

## Implementation Status Matrix

| Feature Area            | Implementation Status | Priority | Files Affected                         |
|-------------------------|----------------------|----------|---------------------------------------|
| GCR Repository Deletion | ❌ Not Implemented   | Critical | `pkg/client/gcr/repository.go`        |
| Cross-Region ECR Auth   | ❌ Not Implemented   | Critical | `pkg/client/ecr/auth.go`              |
| Scheduled Replication   | ❌ Not Implemented   | Critical | `pkg/replication/scheduler.go`        |
| GCR Repository Listing  | ❌ Stub Only         | High     | `pkg/client/gcr/client.go`            |
| Secrets Management      | ❌ Not Implemented   | High     | `main.go`                             |
| Core Replication Logic  | ❌ Stub Only         | High     | `main.go`                             |
| Server Mode             | ❌ Not Implemented   | Medium   | `main.go`                             |
| Checkpoint Management   | ❌ Not Implemented   | Medium   | `main.go`                             |
| Base Client Methods     | ❌ Not Implemented   | Medium   | `pkg/client/common/client.go`         |
| GCR Tag Listing         | ⚠️ Basic Only        | Low      | `pkg/client/gcr/repository.go`        |
| Mock Image Interface    | ⚠️ Minimal Only      | Low      | `pkg/client/gcr/repository.go`        |
| Network Delta Optimize  | ⚠️ Basic Only        | Low      | `pkg/network/delta.go`                |

## Next Steps

To complete these placeholder implementations, the following approach is recommended:

1. Start with the Critical priority items:
   - Implement GCR image deletion using the GCR API
   - Add support for cross-region ECR authentication
   - Complete the scheduled replication implementation with proper cron parsing

2. Proceed to High priority items:
   - Replace mock GCR repository listing with real API calls
   - Implement the secrets management functionality
   - Complete the copier and tree replicator implementations

3. Follow with Medium and Low priority items as resources allow

Each implementation should include proper error handling, logging, metrics collection, and tests to ensure robustness.