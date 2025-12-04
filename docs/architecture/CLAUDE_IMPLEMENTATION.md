# Claude Implementation Guide — Freightliner (Optimized)

## 1. Role & Objective

You are an implementation agent working on **freightliner**, a production-ready Go container registry replication tool, housed at:
- https://github.com/hemzaz/freightliner

Your job is to:
- Achieve production-ready container registry synchronization across AWS ECR, Google GCR, and generic registries.
- Eliminate delegation to external tools (docker, skopeo, crane).
- Produce large, coherent changes with matching tests and documentation.
- Keep CI green and all integrations working.

---

## 2. Core Constraints

1. **Single Go binary:**
   Only `freightliner` is built. CLI, HTTP server, and worker modes are all implemented natively in Go.

2. **No external tool dependencies:**
   - Do not spawn or depend on external container tools (docker, skopeo, crane).
   - Remove or ignore any legacy delegation logic (wrapper scripts, tool execs).
   - All registry operations must use native Go clients.

3. **Production-first:**
   - Registry API behavior is the primary reference.
   - If in doubt, follow official API specifications, then optimize.

4. **No placeholders:**
   - No TODO comments or unimplemented stubs.
   - Every change must be production-grade and compilable.

5. **CI & tests:**
   - Always keep `make lint`, `make test`, `make security` passing.
   - Extend tests for any behavior you change.

---

## 3. Documents You Must Obey

1. **Mission Brief** (docs/architecture/MISSION_BRIEF.md) — Main production-readiness contract.
2. **AGENTS.md** (docs/architecture/AGENTS.md) — Agent roster defining ownership and contracts.
3. **Registry API specifications:**
   - Docker Registry HTTP API V2
   - OCI Distribution Specification
   - AWS ECR API documentation
   - Google Artifact Registry API documentation
4. **Current Go code** (to be corrected; conflicts with API specs are bugs unless explicitly justified).

**If there is a conflict:**
1. Registry API behavior wins.
2. Then mission brief.
3. Then AGENTS.md.
4. Then existing Go code.

---

## 4. How to Propose and Implement Changes

### 4.1 Map changes to agents

For every change, identify which agents you are affecting. For example:

**Adding encryption support:**
- A61 EncryptionManagerAgent
- A62 AwsKmsAgent
- A63 GcpKmsAgent
- A64 EnvelopeEncryptionAgent
- A97 IntegrationTestAgent

Include this agent mapping in your explanation and commit messages.

### 4.2 Preserve structure, then optimize

1. **Mirror API patterns:**
   - Keep client adapters aligned with API specifications.
   - Use analogous package and function boundaries.

2. **Re-express in Go idioms:**
   - Use interfaces, structs, and error handling patterns.
   - Keep behavior identical to API specifications.

3. **Optimize only when safe:**
   - Add goroutine pools or caching only after behavior is pinned by tests.
   - Maintain reference implementations for validation.

### 4.3 Tests and CI alignment

Every substantive change must include:
- Unit tests in relevant packages.
- Integration tests where behavior spans layers.
- Interop tests if affecting registry interactions.

Make sure:
- Tests run on all CI targets (Linux, macOS, Windows where applicable).
- New features are covered, not just happy paths.

---

## 5. Specific Guidance: Removing External Tools & Implementing Native Clients

### 5.1 External tool removal

When you encounter external tool dependencies:

**Remove code paths that:**
- Exec docker, skopeo, crane, or other tools.
- Invoke wrapper scripts only to reach external tools.
- Rely on environment variables solely for tool delegation.

**Replace them with:**
- Direct Go calls to registry client implementations.
- Clear errors if a mode is not yet fully implemented.

**Update:**
- Docs and `--help` to avoid promising external tool usage.
- Tests to assert that tools are not required on system PATH.

### 5.2 Native registry client flow

Implement registry clients in pure Go:

1. **Factory pattern:**
   - Parse registry URL to determine type (ECR, GCR, generic).
   - Create appropriate client adapter.
   - Cache client instances per registry.

2. **Authentication:**
   - Use AWS SDK Go v2 for ECR (IAM, STS).
   - Use Google Cloud Go SDK for GCR (OAuth2, service accounts).
   - Implement Docker Registry v2 auth for generic registries.

3. **API operations:**
   - List repositories with pagination.
   - List tags per repository.
   - Get manifests by digest or tag.
   - Put manifests and layers.
   - Delete tags/manifests (where supported).

4. **Interoperability:**
   - Ensure: ECR ↔ GCR, ECR ↔ Generic, GCR ↔ Generic transfers work.
   - Handle manifest format conversions (Docker ↔ OCI).
   - Test all registry pair combinations.

5. **Testing:**
   - Add unit tests in pkg/client/ packages.
   - Add integration tests in tests/integration/.
   - Extend interop harness to validate native clients without external tools.

---

## 6. Style and Code Quality

Follow idiomatic Go:
- Use `error` returns for error propagation.
- Prefer interfaces and composition over inheritance.
- Keep packages small, focused, and testable.
- Use `context.Context` for cancellation and timeouts.

**Do not:**
- Introduce unsafe code unless absolutely necessary and well-audited.
- Add unbounded goroutines or global mutable state.
- Add new dependencies without strong justification.

**Naming conventions:**
- Packages: lowercase, single word (e.g., `client`, `replication`, `cache`).
- Interfaces: descriptive names ending in relevant suffix (e.g., `RegistryClient`, `Authenticator`, `Provider`).
- Structs: PascalCase (e.g., `EcrClient`, `WorkerPool`).
- Functions: camelCase for unexported, PascalCase for exported.

---

## 7. How to Reason About Changes

When responding to tasks:

1. **Restate the goal** in terms of production-readiness and agents.

2. **Identify the layers:**
   - CLI (`cmd/`), service (`pkg/service/`), client (`pkg/client/`), etc.

3. **Plan in coherent chunks:**
   - Prefer multi-file, feature-complete patches.

4. **Show concrete code:**
   - Provide full functions or structs, not fragments.
   - Ensure code compiles (no placeholders).

5. **Attach tests:**
   - Provide test code alongside implementations.
   - Explain what behavior each test validates.

---

## 8. When You Encounter Gaps or Ambiguity

If behavior is unclear:

1. **Assume registry API specifications are correct.**

2. **Infer intention from:**
   - Official API documentation.
   - go-containerregistry library patterns.
   - Existing observed behavior.

3. **Propose:**
   - Conservative implementation that matches API assumptions.
   - Tests that demonstrate and lock in that behavior.

**If you must choose between cleverness and production-readiness:**
- Choose production-readiness.

---

## 9. Common Implementation Patterns

### 9.1 Client adapter pattern

```go
// Factory creates appropriate client based on URL scheme
func (f *Factory) CreateClientForRegistry(ctx context.Context, registryURL string) (interfaces.RegistryClient, error) {
    u, err := url.Parse(registryURL)
    if err != nil {
        return nil, err
    }

    switch u.Scheme {
    case "ecr":
        return f.CreateECRClient()
    case "gcr":
        return f.CreateGCRClient()
    default:
        return f.CreateGenericClient(registryURL)
    }
}
```

### 9.2 Worker pool pattern

```go
// WorkerPool manages concurrent job execution
type WorkerPool struct {
    workers   int
    jobQueue  chan Job
    results   chan Result
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
}

func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()
    for {
        select {
        case job := <-wp.jobQueue:
            result := job.Execute(wp.ctx)
            wp.results <- result
        case <-wp.ctx.Done():
            return
        }
    }
}
```

### 9.3 Retry with exponential backoff

```go
func retryWithBackoff(ctx context.Context, operation func() error, maxAttempts int) error {
    var err error
    delay := 1 * time.Second

    for attempt := 0; attempt < maxAttempts; attempt++ {
        err = operation()
        if err == nil {
            return nil
        }

        if !isRetryable(err) {
            return err
        }

        select {
        case <-time.After(delay):
            delay *= 2
            if delay > 30*time.Second {
                delay = 30 * time.Second
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }

    return fmt.Errorf("max retry attempts reached: %w", err)
}
```

### 9.4 Structured logging with trace IDs

```go
func (s *ReplicationService) ReplicateRepository(ctx context.Context, source, dest string) error {
    traceID := uuid.New().String()
    logger := s.logger.With(
        zap.String("trace_id", traceID),
        zap.String("component", "replication"),
        zap.String("source", source),
        zap.String("destination", dest),
    )

    logger.Info("starting repository replication")

    err := s.doReplicate(ctx, source, dest)
    if err != nil {
        logger.Error("replication failed",
            zap.Error(err),
            zap.Int("error_code", ErrorCodeReplicationFailed))
        return err
    }

    logger.Info("repository replication completed successfully")
    return nil
}
```

### 9.5 Context-aware operations

```go
func (c *EcrClient) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
    var repos []string
    var nextToken *string

    for {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }

        input := &ecr.DescribeRepositoriesInput{
            NextToken: nextToken,
        }

        output, err := c.client.DescribeRepositories(ctx, input)
        if err != nil {
            return nil, err
        }

        for _, repo := range output.Repositories {
            if prefix == "" || strings.HasPrefix(*repo.RepositoryName, prefix) {
                repos = append(repos, *repo.RepositoryName)
            }
        }

        if output.NextToken == nil {
            break
        }
        nextToken = output.NextToken
    }

    return repos, nil
}
```

---

## 10. Testing Requirements

### 10.1 Unit tests

Every package must have comprehensive unit tests:

```go
func TestEcrClient_ListRepositories(t *testing.T) {
    tests := []struct {
        name    string
        prefix  string
        want    []string
        wantErr bool
    }{
        {
            name:   "list all repositories",
            prefix: "",
            want:   []string{"repo1", "repo2", "repo3"},
        },
        {
            name:   "list with prefix",
            prefix: "repo",
            want:   []string{"repo1", "repo2", "repo3"},
        },
        {
            name:   "list with specific prefix",
            prefix: "repo1",
            want:   []string{"repo1"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := newMockEcrClient(tt.want)
            got, err := client.ListRepositories(context.Background(), tt.prefix)
            if (err != nil) != tt.wantErr {
                t.Errorf("ListRepositories() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ListRepositories() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 10.2 Integration tests

Test against real registries using test accounts:

```go
// +build integration

func TestIntegration_EcrToGcr(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    ctx := context.Background()

    // Setup
    sourceClient := setupEcrTestClient(t)
    destClient := setupGcrTestClient(t)

    // Create test repository and push test image
    testRepo := "test-" + uuid.New().String()
    setupTestImage(t, sourceClient, testRepo)

    // Execute replication
    service := NewReplicationService(sourceClient, destClient, logger)
    err := service.ReplicateRepository(ctx, testRepo, testRepo)

    // Verify
    assert.NoError(t, err)
    assertImageExists(t, destClient, testRepo, "latest")
    assertManifestMatches(t, sourceClient, destClient, testRepo, "latest")

    // Cleanup
    cleanup(t, sourceClient, destClient, testRepo)
}
```

### 10.3 Benchmark tests

Performance-critical paths must have benchmarks:

```go
func BenchmarkWorkerPool_Concurrent(b *testing.B) {
    pool := NewWorkerPool(runtime.NumCPU())
    pool.Start()
    defer pool.Stop()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        job := &mockJob{duration: 1 * time.Millisecond}
        pool.Submit(job)
    }
    pool.Wait()
}
```

---

## 11. Documentation Requirements

Every exported type, function, and package must have documentation:

```go
// Package replication provides concurrent container image replication
// across multiple registry types including AWS ECR, Google GCR, and
// generic Docker v2 registries.
//
// The package implements a worker pool pattern for concurrent transfers,
// checkpoint-based resumability, and comprehensive error handling.
package replication

// WorkerPool manages a pool of worker goroutines for concurrent job execution.
// It provides graceful shutdown, progress tracking, and error aggregation.
//
// Example usage:
//   pool := NewWorkerPool(WorkerPoolConfig{
//       Workers: runtime.NumCPU(),
//       QueueSize: 1000,
//   })
//   pool.Start()
//   defer pool.Stop()
//
//   job := &ReplicationJob{...}
//   pool.Submit(job)
type WorkerPool struct {
    // workers is the number of concurrent worker goroutines
    workers int

    // jobQueue is the channel for incoming jobs
    jobQueue chan Job

    // ... other fields
}

// Submit adds a job to the worker pool queue for execution.
// It blocks if the queue is full. Use SubmitAsync for non-blocking submission.
//
// Returns an error if the pool is stopped or the context is cancelled.
func (wp *WorkerPool) Submit(ctx context.Context, job Job) error {
    select {
    case wp.jobQueue <- job:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    case <-wp.stopChan:
        return ErrPoolStopped
    }
}
```

---

## 12. Security Considerations

### 12.1 Secrets handling

Never log or expose sensitive data:

```go
// ✅ GOOD: Redact credentials in logs
logger.Info("authenticating to registry",
    zap.String("registry", registryURL),
    zap.String("username", "***REDACTED***"))

// ❌ BAD: Never log actual credentials
logger.Info("authenticating",
    zap.String("password", actualPassword))  // NEVER DO THIS
```

### 12.2 Input validation

Always validate external input:

```go
func ValidateRegistryURL(urlStr string) error {
    u, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    if u.Scheme != "ecr" && u.Scheme != "gcr" &&
       u.Scheme != "http" && u.Scheme != "https" {
        return fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
    }

    if u.Host == "" {
        return fmt.Errorf("registry host is required")
    }

    return nil
}
```

### 12.3 Resource limits

Prevent resource exhaustion:

```go
// Limit concurrent operations
semaphore := make(chan struct{}, maxConcurrent)

for _, item := range items {
    semaphore <- struct{}{} // Acquire

    go func(item Item) {
        defer func() { <-semaphore }() // Release
        process(item)
    }(item)
}
```

---

## 13. Commit Message Format

Use conventional commits with agent mapping:

```
<type>(<scope>): <subject>

Agents: <agent-ids>

<body>

<footer>
```

**Example:**

```
feat(encryption): add Google Cloud KMS support

Agents: A63 (GcpKmsAgent), A61 (EncryptionManagerAgent),
        A64 (EnvelopeEncryptionAgent), A97 (IntegrationTestAgent)

Implement Google Cloud KMS encryption provider with:
- KMS key encryption/decryption
- Envelope encryption pattern
- Data key generation
- Integration tests with test project

Closes #123
```

**Types:** feat, fix, docs, style, refactor, test, chore

---

## 14. Summary

- **Goal:** Production-ready Go container registry replication.
- **Binary:** One `freightliner` binary with native CLI/server/worker.
- **Forbidden:** External tool dependencies (docker, skopeo, crane).
- **Guardrails:** Mission brief + AGENTS.md + registry API specs.
- **Method:** Large, coherent, well-tested changes that keep CI green.

Always align your work with agent responsibilities and this implementation guide, and always keep production-readiness and native Go implementation at the center of your decisions.

---

## 15. Quick Reference

### Build and Test Commands

```bash
# Format code
make fmt

# Lint code
make lint

# Run unit tests
make test

# Run integration tests (requires test accounts)
make test-integration

# Run with race detector
make test-race

# Security scan
make security

# Build binary
make build

# Build for all platforms
make release-build

# Generate documentation
make docs
```

### Common Tasks

| Task | Command | Notes |
|------|---------|-------|
| Add new registry type | Implement `interfaces.RegistryClient` | See pkg/client/ |
| Add new auth method | Implement auth interface | See pkg/interfaces/auth.go |
| Add new encryption provider | Implement `Provider` interface | See pkg/security/encryption/ |
| Add new metrics | Use prometheus client | See pkg/metrics/ |
| Add new CLI command | Add cobra command | See cmd/ |
| Update API | Modify server handlers | See pkg/server/ |

### Agent Quick Lookup

| Component | Primary Agents | Secondary Agents |
|-----------|----------------|------------------|
| ECR Client | A32, A52 | A31, A35, A60 |
| GCR Client | A33, A53 | A31, A35, A60 |
| Generic Client | A34, A54-A55 | A31, A35, A60 |
| Replication | A41-A50 | A71-A80, A81-A90 |
| Encryption | A61-A70 | A09, A98 |
| Caching | A71-A80 | A43-A50 |
| Server | A15, A95 | A91, A96 |
| Testing | A96-A100 | All agents |

---

**Remember:** When in doubt, check the Mission Brief and AGENTS.md, follow registry API specifications, write tests, and keep production-readiness as the primary goal.
