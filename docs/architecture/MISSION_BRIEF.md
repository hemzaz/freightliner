# Mission Brief — Freightliner
(HIGH-THROUGHPUT CONTAINER REGISTRY REPLICATION — PRODUCTION-READY MULTI-CLOUD SYNC)

Pure-Go **container registry replication** tool achieving **production-ready parity** with enterprise container registry features across **AWS ECR**, **Google GCR**, and **generic registries** (Docker Hub, Harbor, Quay, GitLab, etc.).

- Repository: <https://github.com/hemzaz/freightliner>
- Go Version: **1.21+**
- License: **Apache-2.0** (or MIT)

You are an implementation agent. Your job is to **close feature gaps fast**, with **large, coherent changes** backed by tests, not tiny edits.

Always use these commands to validate code:
```bash
make lint && make test && make security && make build
```

Address resulting issues immediately.

Implement missing components and functionalities to achieve production-ready container registry synchronization across all major cloud providers.

---

## 0. Sources of Truth (Ranked)

1. **Observed registry behavior**
   Real registry APIs (ECR, GCR, Docker Hub, Harbor) on production systems: API contracts, authentication flows, manifest formats, layer transfer protocols.

2. **Upstream specifications & docs**
   Especially:
   - Docker Registry HTTP API V2 specification
   - OCI Distribution Specification
   - AWS ECR API documentation
   - Google Artifact Registry API documentation
   - go-containerregistry library patterns

3. **This Mission Brief + AGENTS.md**
   - Single **`freightliner`** binary (CLI + HTTP server + worker pool).
   - Error trailers (structured logging with trace IDs).
   - Concurrent worker pools and batching rules.
   - CI & production deployment contracts.
   - **No external tools / no wrapper scripts.**

4. **Current Go code**
   Implementation to be corrected. If it conflicts with registry API semantics, treat it as a bug unless this brief explicitly says otherwise.

5. **Go-specific optimizations**
   Goroutine pools, channel patterns, sync primitives, buffer pools: allowed if they **do not change** observable behavior.

If upstream registry behavior contradicts anything else, **upstream wins** unless explicitly overridden here.

---

## 0.5 No External Tools / No Wrapper Scripts

This is a hard constraint:

- **Do NOT** spawn or delegate to external tools (docker, skopeo, crane) for any normal operation mode.
- **Do NOT** rely on wrapper scripts, environment variable hacks, or system binaries to perform registry operations.
- All roles (client, server, worker, scheduler) must be implemented **natively in Go** inside the **single `freightliner` binary**.
- Any existing delegation/bridge/wrapper logic must be:
  - Removed from the codebase.
  - Removed from docs and `--help`.
  - Replaced with **native Go registry client adapters** that mirror registry API behavior.

The only acceptable uses of external programs are:

- Transport utilities (SSH for remote execution, if explicitly required).
- User-specified custom scripts for hooks/notifications.

---

## 1. Binary, Branding & Error Trailers

### 1.1 Single binary, multiple modes

- Workspace builds **one** binary: `freightliner`.
- Modes:
  - CLI mode (default): Execute replication commands from terminal.
  - HTTP server mode (`serve`): RESTful API for programmatic access.
  - Worker mode: Execute background replication jobs.
- Downstream packaging may create service files or containers, but you **do not** build separate binaries per mode.
- All modes must be fully functional in Go (no delegation to external tools).

### 1.2 Canonical metadata (do not hard-code)

From build system or `version.go`:

- Binary name: `freightliner`
- Version string: `v1.0.0` (follows semantic versioning)
- Configuration directory: `/etc/freightliner/` or `~/.config/freightliner/`
  - Example: `/etc/freightliner/config.yaml`
  - Example: `/etc/freightliner/registries.yaml`
- Project URL: <https://github.com/hemzaz/freightliner>

Any user-facing surface — CLI help, `version`, docs, packaging metadata, logs — must derive these values via **shared metadata helpers**, not literals.

### 1.3 Error trailers (structured logging)

All errors include structured context:

```text
... [trace_id=<uuid>] [component=<component>] [operation=<operation>] [code=<error_code>]
```

Implement via `pkg/helper/logger.go`:
- Component enum: Client, Server, Replication, Worker, Cache, Security, etc.
- Trace ID generation using UUID v4.
- Operation context (e.g., "replicate_repository", "list_tags", "transfer_layer").

Use a central constructor:

```go
logger.Error("operation failed",
    zap.String("trace_id", traceID),
    zap.String("component", component),
    zap.String("operation", operation),
    zap.Int("error_code", code),
    zap.Error(err))
```

Components must mirror registry operation semantics. Only the trace format is freightliner-specific.

---

## 2. Behavioral Parity (What Must Match)

You are aiming for "swap container replication tools for freightliner and nothing breaks", aside from names/paths and structured logging.

### 2.1 CLI & option semantics

Implement and align:

**Core flags:**
- `--source`: Source registry URL (ecr://, gcr://, https://, etc.)
- `--dest` / `--destination`: Destination registry URL
- `--workers`: Number of concurrent workers (default: auto-detect CPUs)
- `--dry-run`: Preview changes without executing
- `--force`: Overwrite existing images
- `--log-level`: Logging verbosity (debug, info, warn, error)

**Registry-specific:**
- `--source-profile` / `--dest-profile`: AWS credential profiles
- `--source-project` / `--dest-project`: GCP project IDs
- `--source-region` / `--dest-region`: AWS regions

**Filtering:**
- `--include-tags`: Tag patterns to include (regex)
- `--exclude-tags`: Tag patterns to exclude (regex)
- `--include-repos`: Repository patterns to include
- `--exclude-repos`: Repository patterns to exclude

**Security:**
- `--encrypt`: Enable AES-256-GCM encryption
- `--kms-key`: AWS KMS or GCP KMS key ID
- `--insecure`: Skip TLS verification (for internal registries)

**Operational:**
- `--checkpoint-id`: Resume from checkpoint
- `--batch-size`: Batch size for tree replication
- `--timeout`: Operation timeout duration
- `--retry-attempts`: Number of retry attempts
- `--retry-delay`: Initial retry delay with exponential backoff

**Server mode:**
- `serve --port 8080`: Start HTTP server
- `--config`: Path to configuration file

Help and misuse:
- `freightliner --help` should be clear, comprehensive, and production-ready.
- Misuse errors (missing operands, invalid URLs) must have appropriate exit codes and clear messages.

### 2.2 Replication semantics (critical)

Go replication handling must mirror production registry behavior:

**Manifest replication:**
- Support for Docker Manifest V2 Schema 1 & 2
- Support for OCI Image Manifest
- Support for manifest lists (multi-architecture)
- Preserve manifest digests and media types

**Layer transfer:**
- Concurrent layer downloads with worker pool
- Resume capability for interrupted transfers
- Checksum verification (SHA256)
- Deduplication: Skip layers already present in destination

**Tag management:**
- Preserve all tags from source
- Support tag filtering and exclusion
- Handle tag deletion in destination when source is removed (optional)

**Interactions:**
- `--force` with existing manifests
- `--dry-run` with layer counting
- `--checkpoint` with partial transfers
- `--encrypt` with layer data

Existing tests:
- Unit tests in `pkg/service/`
- Integration tests in `tests/integration/`
- Must pass on Linux, macOS, and Windows (where applicable)

### 2.3 Authentication, registry types, encryption

**Authentication:**
- AWS ECR: IAM credentials, STS assume-role, cross-account
- GCP GCR: OAuth2, service accounts, Application Default Credentials
- Generic registries: Anonymous, Basic auth, Bearer token
- Docker config.json credential store integration

**Registry types:**
- AWS ECR: Full API support with pagination
- Google GCR/Artifact Registry: Dual-mode with fallback
- Docker Hub: Rate limiting awareness
- Harbor: Project-based access
- Quay.io: Robot accounts
- GitHub Container Registry: PAT tokens
- Azure ACR: Managed identity support
- Artifactory: Token-based auth

**Encryption:**
- `--encrypt`: AES-256-GCM encryption in transit
- Customer-managed keys: AWS KMS and GCP KMS
- Envelope encryption for data protection
- Key rotation support

### 2.4 HTTP Server & API

Server mode (`freightliner serve`) must provide:

**Health endpoints:**
- `GET /health`: Basic health check
- `GET /ready`: Readiness probe
- `GET /live`: Liveness probe

**API endpoints:**
- `POST /api/v1/replicate`: Single repository replication
- `POST /api/v1/replicate-tree`: Multi-repository batch replication
- `GET /api/v1/jobs`: List active jobs
- `GET /api/v1/jobs/{id}`: Get job status
- `DELETE /api/v1/jobs/{id}`: Cancel job
- `POST /api/v1/checkpoints/{id}/resume`: Resume from checkpoint

**Metrics:**
- `GET /metrics`: Prometheus metrics exposition
- Comprehensive metrics (success rate, latency, throughput)

**Security:**
- Optional mTLS for client authentication
- API key authentication
- Rate limiting per client

### 2.5 Worker Pool & Concurrency

Worker pool must be implemented natively in Go:

**Architecture:**
- Configurable worker count (default: runtime.NumCPU())
- Job queue with priority support
- Graceful shutdown with in-flight job completion
- Worker health monitoring and auto-recovery

**Job management:**
- Job scheduling with dependencies
- Checkpoint creation for resumability
- Progress tracking and reporting
- Error aggregation and reporting

**Concurrency patterns:**
- Channel-based work distribution
- sync.WaitGroup for synchronization
- context.Context for cancellation
- Semaphore for resource limiting

---

## 3. Architecture & Performance (Constraints You Must Respect)

### 3.1 Worker Pool & Goroutines

Existing concurrency patterns are first-class. Preserve them:

**pkg/replication/:**
- Worker pool implementation with configurable workers
- Goroutine-based concurrent layer transfers
- Channel-based job distribution
- Semaphore-based resource limiting

**Rules:**
1. Always implement robust error handling in goroutines.
2. Use context.Context for cancellation propagation.
3. Ensure proper cleanup with defer statements.
4. Test concurrent paths with race detector (`go test -race`).

### 3.2 Layering (AGENTS.md contract)

Obey the layering from AGENTS.md:

**cmd/:**
- Parses CLI flags.
- Decides mode (CLI/server/worker).
- Calls service layer with configuration.

**pkg/service/:**
- Orchestrates replication workflows.
- Manages checkpoints and state.
- Coordinates worker pools.

**pkg/client/:**
- Registry client adapters (ECR, GCR, Generic).
- Factory pattern for client creation.
- Authentication handling.

**pkg/replication/:**
- Worker pool implementation.
- Job scheduling and execution.
- Progress tracking.

**Lower packages:**
- `pkg/security/`: Encryption, mTLS, secrets.
- `pkg/cache/`: LRU and high-performance caching.
- `pkg/network/`: Transfer manager, compression, buffering.
- `pkg/metrics/`: Prometheus metrics collection.
- `pkg/helper/`: Logging, utilities, banner.

Do not introduce cross-layer shortcuts.

---

## 4. CI & Interop (Non-Negotiable)

### 4.1 CI workflows

Keep these green:

**`.github/workflows/ci.yml`:**
```bash
make lint          # golangci-lint with strict rules
make test          # Unit tests with race detection
make test-ci       # CI-optimized tests with coverage
make security      # gosec security scanning
make build         # Multi-arch builds
```

**`.github/workflows/integration.yml`:**
- Integration tests against real registries (using test accounts)
- Docker Hub, Harbor (self-hosted), LocalStack for ECR
- Tests for all authentication methods
- Protocol version compatibility tests

If you add packages, features, or build targets, you must:
- Update Makefile accordingly.
- Update CI workflows to include new tests.
- Ensure compatibility across Linux, macOS, Windows.

### 4.2 Interop harness

You must rely on an interop harness that:

**Tests against production registries:**
- AWS ECR (test account with limited permissions)
- Google GCR (test project)
- Docker Hub (anonymous and authenticated)
- Harbor (self-hosted in CI)

**For each scenario:**
- Execute freightliner with specific flags and configuration.
- Verify:
  - Exit codes (0 for success, non-zero for failures).
  - Manifest integrity in destination.
  - Layer checksums and sizes.
  - Tag preservation and correctness.
  - Authentication and authorization behavior.

**Key scenarios (minimum):**
- Single repository replication (various architectures).
- Multi-repository batch replication with filters.
- Encrypted replication with KMS.
- Cross-cloud replication (ECR → GCR, GCR → Docker Hub).
- Resume from checkpoint after interruption.
- Concurrent replication with multiple workers.
- Rate limiting and retry logic.
- Authentication failures and error handling.

The harness is part of the production-readiness contract: do not break it.

---

## 5. High-Throughput Behavior (How You Work)

You are optimized for throughput. That means:

- **Batch work:** Prefer patches that fix whole feature clusters (e.g., all encryption interactions) over single-line tweaks.
- **Minimize passes:** When you open a module (e.g., transfer manager), fix all obvious issues there, not just the one that tripped a test.
- **Exploit structure:** Move logic into small, well-named helpers and reuse them; avoid repeating similar code across packages.
- **Keep tests running:** Every substantial change should be accompanied by tests that pin the new behavior.

**Operational rules:**
1. Never leave placeholders (TODO comments, panic stubs, unimplemented features).
2. Each time you touch a feature:
   - Check registry API behavior.
   - Update Go implementation.
   - Add/extend tests.
   - Ensure CI assumptions still hold.
3. Prefer fewer, larger, coherent commits over many tiny ones.

---

## 6. Call to Action — High-Throughput Plan

Follow these phases in order, doing as much as possible per phase.

### Phase 0 — Remove external tool dependencies

1. Identify all code paths that:
   - Shell out to docker, skopeo, crane, or other tools.
   - Rely on wrapper scripts for registry operations.
   - Use environment variables for tool delegation.

2. Remove those paths and replace them with:
   - Native Go registry client implementations.
   - Clear diagnostics when a feature is not yet implemented.

3. Update tests:
   - Delete or refactor tests that assume external tools.
   - Add tests that assert no external tools are invoked.

4. Update docs:
   - Mission brief, AGENTS.md, README, and CLI --help.

### Phase 1 — Map and log feature gaps

5. For each feature group:
   - Authentication, replication, encryption, caching, monitoring, API.

6. For each group, design multiple representative scenarios (3–10):
   - Commands, configurations, input data patterns.

7. For every scenario:
   - Run against real registries.
   - Record results, errors, and performance metrics.

8. For each gap or issue:
   - Classify: bug, missing feature, performance issue, documentation gap.
   - Log it in `docs/gaps.md` with:
     - Scenario description.
     - Expected vs actual behavior.
     - Priority and category.

Do this once per group, not per individual flag.

### Phase 2 — Fix replication as a complete cluster

9. From registry API specs and observed behavior, derive exact replication semantics:
   - Manifest formats and media types.
   - Layer transfer protocols.
   - Tag management and deletion.
   - Multi-architecture support.

10. Update Go replication implementation in one pass:
    - Maintain concurrent worker pool efficiency.
    - Produce registry-equivalent behavior.
    - Keep checkpointing and resumability.

11. Extend tests:
    - Service-level replication tests for multiple patterns.
    - CLI-level integration tests including edge cases.
    - Interop tests comparing freightliner vs native tools.

### Phase 3 — Sweep option groups in large batches

For each group, run measure → diff → fix → test → docs in one batch:

12. **Authentication group:**
    - AWS IAM, GCP OAuth2, generic Basic/Bearer.
    - Credential providers and caching.
    - STS assume-role and cross-account access.

13. **Encryption & security group:**
    - AES-256-GCM encryption.
    - AWS KMS and GCP KMS integration.
    - Envelope encryption and key rotation.
    - mTLS for registry communication.

14. **Performance & caching group:**
    - Worker pool optimization.
    - LRU cache and high-performance cache.
    - Streaming buffer pools.
    - Compression and transfer optimization.

15. **API & server group:**
    - HTTP server implementation.
    - RESTful endpoints with OpenAPI specs.
    - Job management and lifecycle.
    - Metrics exposition and monitoring.

16. **Monitoring & observability group:**
    - Prometheus metrics (90%+ coverage).
    - Structured logging with trace IDs.
    - Health checks and readiness probes.
    - Performance dashboards.

For each group, do all fixes you can see before moving on, and add tests to pin the behavior.

### Phase 4 — Messages, exit codes, CI, interop

17. **Messages:**
    - Ensure all user-visible text is clear and consistent.
    - Use structured logging for all errors.
    - Attach trace IDs and context everywhere.

18. **Exit codes:**
    - Build a table of exit codes and their conditions.
    - Map error scenarios to appropriate codes consistently.

19. **CI & interop:**
    - Keep all workflows green.
    - Expand interop harness to cover all feature groups.
    - Add performance benchmarks.

20. **Documentation:**
    - Update AGENTS.md and docs/gaps.md to reflect:
      - What is production-ready.
      - What is still in progress.
      - Any known limitations or design decisions.

---

## 7. Production Readiness Criteria

Before declaring a feature production-ready, ensure:

1. **Functionality:**
   - Feature works as designed across all supported registries.
   - Edge cases are handled gracefully.

2. **Testing:**
   - Unit tests with 85%+ coverage.
   - Integration tests against real registries.
   - Performance tests with benchmarks.

3. **Documentation:**
   - Feature documented in README and docs/.
   - CLI help updated.
   - API endpoints documented in OpenAPI spec.

4. **Observability:**
   - Prometheus metrics exposed.
   - Structured logging with appropriate levels.
   - Error scenarios produce actionable messages.

5. **Security:**
   - Security review completed.
   - Secrets not leaked in logs or errors.
   - Authentication and authorization tested.

6. **Performance:**
   - Meets throughput targets.
   - Resource usage is reasonable.
   - Concurrent operations scale linearly.

---

Never introduce new visible behavior without:
- Checking registry API behavior first.
- Updating Go implementation + tests.
- Updating this brief and AGENTS.md if contracts change.

---

## 8. Key Design Patterns

1. **Factory Pattern:** Client factory for registry adapter creation.
2. **Strategy Pattern:** Authentication, encryption, compression strategies.
3. **Adapter Pattern:** Wrapping external registry APIs.
4. **Worker Pool Pattern:** Concurrent job execution.
5. **Repository Pattern:** Registry and repository abstractions.
6. **Singleton Pattern:** Global buffer pools and caches.
7. **Observer Pattern:** Progress tracking and metrics.
8. **Command Pattern:** Job scheduling and execution.

---

## 9. Summary

- **Goal:** Production-ready Go container registry replication tool.
- **Binary:** One `freightliner` binary with native CLI/server/worker modes.
- **Forbidden:** External tool dependencies or wrapper scripts.
- **Guardrails:** Mission brief + AGENTS.md + registry API behavior.
- **Method:** Large, coherent, well-tested changes that keep CI green.

Always align your work with the agents' responsibilities and this mission brief, and always keep production-readiness and native Go implementation at the center of your decisions.
