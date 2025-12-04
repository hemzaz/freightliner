# AGENTS.md — Freightliner Agent Roster (100 Agents)

## Overview

This document defines 100 specialized "agents" — conceptual responsibilities that structure how changes are designed and reviewed. When you modify code, identify which agents are involved and keep their contracts intact.

Each agent has:
- **ID** — Stable identifier.
- **Name** — Short label.
- **Scope** — What part of the system it owns.
- **Key Duties** — What it must guarantee.
- **No External Tools Rule** — All agents must respect the global rule: no delegation to external tools (docker, skopeo, crane).

---

## A. Governance & Meta Agents (A01–A10)

### A01 — ProductionReadinessAgent
- **Scope:** Global production-readiness and quality standards.
- **Key Duties:**
  - Enforce "swap registry tools for freightliner and nothing breaks".
  - Block incomplete features from reaching production.
  - Maintain 88%+ production-ready status.
- **Notes:** Approves feature completeness and deployment readiness.

### A02 — MissionBriefAgent
- **Scope:** Mission Brief document.
- **Key Duties:**
  - Keep mission brief in sync with implementation.
  - Reflect removal of external tool dependencies.
  - Document all architectural decisions.
- **Notes:** Updates this document when contracts change.

### A03 — AgentsRosterAgent
- **Scope:** AGENTS.md (this document).
- **Key Duties:**
  - Maintain accurate agent list and responsibilities.
  - Ensure each new feature maps to one or more agents.
  - Prevent overlapping/ambiguous ownership.
- **Notes:** Central registry of agent responsibilities.

### A04 — DocsAgent
- **Scope:** Documentation (README, docs/*, API specs).
- **Key Duties:**
  - Keep docs aligned with CLI and API behavior.
  - Remove references to external tools.
  - Maintain OpenAPI/Swagger specifications.
- **Notes:** Ensures documentation accuracy and completeness.

### A05 — UXTextAgent
- **Scope:** User-visible strings and messages.
- **Key Duties:**
  - Keep error messages clear and actionable.
  - Provide consistent terminology across CLI/API.
  - Maintain user-friendly help text.
- **Notes:** Coordinates with structured logging.

### A06 — StructuredLoggingAgent
- **Scope:** Logging format and trace IDs.
- **Key Duties:**
  - Enforce structured logging with trace IDs.
  - Ensure consistent log levels and formats.
  - Prevent secrets leakage in logs.
- **Notes:** Works with observability stack.

### A07 — BrandingAgent
- **Scope:** Names, versions, URLs.
- **Key Duties:**
  - Centralize branding in version package.
  - Avoid hard-coded metadata throughout code.
  - Maintain consistent version strings.
- **Notes:** Keeps `--version` and banner consistent.

### A08 — RoadmapAgent
- **Scope:** Development roadmap and priorities.
- **Key Duties:**
  - Maintain feature gap analysis in docs/gaps.md.
  - Prioritize large, coherent feature batches.
  - Track production-readiness metrics.
- **Notes:** Informs which agent groups are active next.

### A09 — SecurityReviewAgent
- **Scope:** Security posture and compliance.
- **Key Duties:**
  - Guard against injection attacks and leaks.
  - Review authentication and encryption flows.
  - Ensure secrets management best practices.
- **Notes:** Reviews all security-sensitive changes.

### A10 — ObservabilityAgent
- **Scope:** Metrics, tracing, and monitoring.
- **Key Duties:**
  - Ensure 90%+ Prometheus metrics coverage.
  - Provide structured logging for debugging.
  - Avoid noisy or redundant metrics.
- **Notes:** Works with Grafana dashboards.

---

## B. CLI & Frontend Agents (A11–A20)

### A11 — CliParserAgent
- **Scope:** Command-line argument parsing.
- **Key Duties:**
  - Parse flags, options, and operands consistently.
  - Provide clear error messages for invalid input.
  - Support environment variable overrides.
- **Notes:** Uses cobra/viper patterns.

### A12 — CliModeSelectionAgent
- **Scope:** Mode detection (CLI/server/worker).
- **Key Duties:**
  - Decide mode based on command and flags.
  - Route to appropriate service layer entry point.
  - Handle graceful mode transitions.
- **Notes:** No external tool fallback.

### A13 — CliRegistryUrlAgent
- **Scope:** Registry URL parsing (ecr://, gcr://, https://).
- **Key Duties:**
  - Parse and validate registry URLs.
  - Extract registry type, region, project info.
  - Provide clear errors for invalid URLs.
- **Notes:** Drives client factory selection.

### A14 — CliHelpUsageAgent
- **Scope:** --help and usage text.
- **Key Duties:**
  - Render comprehensive, well-organized help.
  - Provide examples for common use cases.
  - Match help text to actual behavior.
- **Notes:** Keep help user-friendly and accurate.

### A15 — CliServerFlagsAgent
- **Scope:** Server mode flags (--port, --config).
- **Key Duties:**
  - Parse server-specific configuration.
  - Validate port numbers and addresses.
  - Handle TLS certificate paths.
- **Notes:** Does not start external servers.

### A16 — CliLoggingFlagsAgent
- **Scope:** Logging flags (--log-level, --log-format).
- **Key Duties:**
  - Configure logging level and format.
  - Support JSON and text formats.
  - Handle log file rotation settings.
- **Notes:** Integrates with zap/logrus.

### A17 — CliDryRunAgent
- **Scope:** --dry-run mode.
- **Key Duties:**
  - Preview operations without executing.
  - Count manifests and layers to transfer.
  - Provide accurate size estimates.
- **Notes:** Must not modify any registry state.

### A18 — CliValidationAgent
- **Scope:** Flag validation and conflicts.
- **Key Duties:**
  - Reject incompatible flag combinations.
  - Validate URLs, paths, and identifiers.
  - Provide actionable error messages.
- **Notes:** Fails fast on invalid input.

### A19 — CliEnvVarsAgent
- **Scope:** Environment variable support.
- **Key Duties:**
  - Honor standard registry env vars (AWS_*, GOOGLE_*).
  - Support custom freightliner env vars.
  - Document all supported variables.
- **Notes:** Env vars override config file but not CLI flags.

### A20 — CliConfigFileAgent
- **Scope:** Configuration file loading (YAML).
- **Key Duties:**
  - Load and parse config.yaml files.
  - Support nested registry configurations.
  - Validate configuration schema.
- **Notes:** Uses viper for multi-format support.

---

## C. Core Orchestration Agents (A21–A30)

### A21 — CoreConfigAgent
- **Scope:** Core configuration structures.
- **Key Duties:**
  - Translate CLI/config into service config.
  - Maintain stable configuration API.
  - Support configuration validation.
- **Notes:** Bridge between frontend and services.

### A22 — CoreModeRouterAgent
- **Scope:** Mode routing and dispatch.
- **Key Duties:**
  - Route to replication/server/worker modes.
  - Handle mode-specific initialization.
  - Ensure clean resource lifecycle.
- **Notes:** Clear separation of mode concerns.

### A23 — CoreSessionLifecycleAgent
- **Scope:** Session and job lifecycle.
- **Key Duties:**
  - Start, execute, and clean up jobs.
  - Manage resource allocation and cleanup.
  - Handle graceful shutdown.
- **Notes:** Works with context cancellation.

### A24 — CoreExitCodeAgent
- **Scope:** Exit code mapping.
- **Key Duties:**
  - Map errors to appropriate exit codes.
  - Maintain exit code documentation.
  - Ensure CLI and API consistency.
- **Notes:** Standard Unix exit code conventions.

### A25 — CoreProgressAgent
- **Scope:** Progress tracking and reporting.
- **Key Duties:**
  - Track transfer progress (bytes, layers).
  - Provide real-time progress updates.
  - Support progress bars in CLI.
- **Notes:** Thread-safe progress aggregation.

### A26 — CoreErrorMappingAgent
- **Scope:** Error categorization and mapping.
- **Key Duties:**
  - Categorize errors (network, auth, registry).
  - Provide actionable error messages.
  - Attach trace IDs and context.
- **Notes:** Works with structured logging.

### A27 — CoreFeatureFlagsAgent
- **Scope:** Feature toggles and experimental features.
- **Key Duties:**
  - Manage feature flag system.
  - Document experimental features.
  - Ensure defaults are production-safe.
- **Notes:** Supports gradual rollouts.

### A28 — CoreCheckpointAgent
- **Scope:** Checkpoint creation and resumption.
- **Key Duties:**
  - Save replication state to checkpoints.
  - Resume from saved checkpoints.
  - Clean up stale checkpoints.
- **Notes:** Enables resumable transfers.

### A29 — CoreSignalHandlingAgent
- **Scope:** Signal handling (SIGINT, SIGTERM).
- **Key Duties:**
  - Handle graceful shutdown on signals.
  - Complete in-flight operations if possible.
  - Clean up temporary resources.
- **Notes:** Uses context cancellation.

### A30 — CoreResourceLimitsAgent
- **Scope:** Resource limits and quotas.
- **Key Duties:**
  - Enforce memory and goroutine limits.
  - Prevent resource exhaustion.
  - Fail gracefully on limit violations.
- **Notes:** Uses semaphores and rate limiters.

---

## D. Registry Client & Adapter Agents (A31–A40)

### A31 — ClientFactoryAgent
- **Scope:** Registry client factory pattern.
- **Key Duties:**
  - Create appropriate client based on URL scheme.
  - Support ECR, GCR, and generic registries.
  - Cache client instances appropriately.
- **Notes:** Central point for client creation.

### A32 — EcrClientAgent
- **Scope:** AWS ECR client adapter.
- **Key Duties:**
  - Implement ECR API interactions.
  - Handle pagination and rate limiting.
  - Support cross-account and cross-region.
- **Notes:** Uses AWS SDK Go v2.

### A33 — GcrClientAgent
- **Scope:** Google GCR/Artifact Registry client.
- **Key Duties:**
  - Implement dual-mode (GCR and Artifact Registry).
  - Handle GCP authentication flows.
  - Support multi-location registries.
- **Notes:** Uses Google Cloud Go SDK.

### A34 — GenericClientAgent
- **Scope:** Generic Docker v2 registry client.
- **Key Duties:**
  - Implement Docker Registry HTTP API V2.
  - Support OCI Distribution Specification.
  - Handle various authentication methods.
- **Notes:** Works with Docker Hub, Harbor, Quay, etc.

### A35 — RegistryInterfaceAgent
- **Scope:** RegistryClient interface definition.
- **Key Duties:**
  - Define common registry operations.
  - Ensure consistent behavior across adapters.
  - Support interface segregation.
- **Notes:** pkg/interfaces/client.go

### A36 — RepositoryInterfaceAgent
- **Scope:** Repository interface definition.
- **Key Duties:**
  - Define repository-level operations.
  - Support manifest and tag operations.
  - Enable testing with mocks.
- **Notes:** pkg/interfaces/repository.go

### A37 — BaseClientAgent
- **Scope:** Common client functionality.
- **Key Duties:**
  - Provide shared client utilities.
  - Implement caching and retry logic.
  - Handle common error scenarios.
- **Notes:** pkg/client/common/base_client.go

### A38 — BaseRepositoryAgent
- **Scope:** Common repository functionality.
- **Key Duties:**
  - Implement shared repository operations.
  - Provide caching for tags and manifests.
  - Handle concurrent access safely.
- **Notes:** pkg/client/common/base_repository.go

### A39 — RegistryAutoDetectionAgent
- **Scope:** Automatic registry type detection.
- **Key Duties:**
  - Detect registry type from URL patterns.
  - Probe registry capabilities.
  - Select appropriate client adapter.
- **Notes:** Fallback to generic for unknown types.

### A40 — RegistryInteropAgent
- **Scope:** Cross-registry interoperability.
- **Key Duties:**
  - Ensure ECR ↔ GCR ↔ Generic transfers work.
  - Handle manifest format conversions.
  - Test all registry pair combinations.
- **Notes:** Critical for production readiness.

---

## E. Replication Engine Agents (A41–A50)

### A41 — ReplicationServiceAgent
- **Scope:** Core replication orchestration.
- **Key Duties:**
  - Orchestrate single repository replication.
  - Coordinate source and destination clients.
  - Handle errors and retries.
- **Notes:** pkg/service/replication.go

### A42 — TreeReplicationServiceAgent
- **Scope:** Multi-repository batch replication.
- **Key Duties:**
  - List and filter repositories.
  - Batch replication with checkpoints.
  - Aggregate results and errors.
- **Notes:** pkg/service/tree_replication.go

### A43 — WorkerPoolAgent
- **Scope:** Concurrent worker pool.
- **Key Duties:**
  - Manage configurable worker goroutines.
  - Distribute jobs via channels.
  - Monitor worker health.
- **Notes:** pkg/replication/worker_pool.go

### A44 — JobSchedulerAgent
- **Scope:** Job scheduling and prioritization.
- **Key Duties:**
  - Queue jobs with priorities.
  - Schedule based on dependencies.
  - Balance load across workers.
- **Notes:** pkg/replication/scheduler.go

### A45 — ManifestReplicationAgent
- **Scope:** Manifest transfer and verification.
- **Key Duties:**
  - Copy manifests between registries.
  - Preserve manifest digests.
  - Handle manifest lists (multi-arch).
- **Notes:** Support Docker and OCI formats.

### A46 — LayerReplicationAgent
- **Scope:** Layer transfer and deduplication.
- **Key Duties:**
  - Transfer layers with verification.
  - Skip already-present layers.
  - Handle layer compression.
- **Notes:** Uses streaming for large layers.

### A47 — TagReplicationAgent
- **Scope:** Tag management and filtering.
- **Key Duties:**
  - Copy tags based on filters.
  - Handle tag deletion in sync mode.
  - Preserve tag timestamps where possible.
- **Notes:** Respects include/exclude patterns.

### A48 — DeltaTransferAgent
- **Scope:** Differential transfer optimization.
- **Key Duties:**
  - Compare source and destination states.
  - Transfer only missing/changed artifacts.
  - Minimize bandwidth usage.
- **Notes:** Works with caching layer.

### A49 — ChecksumVerificationAgent
- **Scope:** Integrity verification.
- **Key Duties:**
  - Verify SHA256 checksums for layers.
  - Detect corrupted transfers.
  - Retry failed verifications.
- **Notes:** Critical for data integrity.

### A50 — ProgressReportingAgent
- **Scope:** Transfer progress tracking.
- **Key Duties:**
  - Track bytes transferred per layer.
  - Calculate transfer rates.
  - Estimate time remaining.
- **Notes:** Thread-safe aggregation.

---

## F. Authentication & Security Agents (A51–A60)

### A51 — AuthenticationFactoryAgent
- **Scope:** Authentication provider factory.
- **Key Duties:**
  - Select auth provider based on registry type.
  - Support multiple auth methods per registry.
  - Cache authentication tokens.
- **Notes:** pkg/client/auth/factory.go

### A52 — EcrAuthAgent
- **Scope:** AWS ECR authentication.
- **Key Duties:**
  - Handle IAM credentials and STS tokens.
  - Support assume-role for cross-account.
  - Refresh tokens automatically.
- **Notes:** pkg/client/ecr/auth.go

### A53 — GcrAuthAgent
- **Scope:** GCP OAuth2 authentication.
- **Key Duties:**
  - Handle service account credentials.
  - Support Application Default Credentials.
  - Token refresh with OAuth2.
- **Notes:** pkg/client/gcr/auth.go

### A54 — BasicAuthAgent
- **Scope:** HTTP Basic authentication.
- **Key Duties:**
  - Handle username/password auth.
  - Support Docker config.json credentials.
  - Encode credentials properly.
- **Notes:** Used by generic registries.

### A55 — BearerTokenAuthAgent
- **Scope:** Bearer token authentication.
- **Key Duties:**
  - Handle OAuth2 bearer tokens.
  - Implement token challenge/response.
  - Cache tokens with expiry.
- **Notes:** Docker Registry v2 auth.

### A56 — TokenCacheAgent
- **Scope:** Authentication token caching.
- **Key Duties:**
  - Cache tokens with TTL.
  - Refresh tokens before expiry.
  - Invalidate on auth errors.
- **Notes:** Thread-safe cache operations.

### A57 — CredentialProviderAgent
- **Scope:** Credential storage and retrieval.
- **Key Duties:**
  - Load credentials from config files.
  - Support environment variables.
  - Integrate with credential helpers.
- **Notes:** Docker config.json compatibility.

### A58 — MtlsAgent
- **Scope:** Mutual TLS configuration.
- **Key Duties:**
  - Load client certificates and keys.
  - Configure TLS for registries.
  - Handle certificate validation.
- **Notes:** pkg/security/mtls.go

### A59 — InsecureTlsAgent
- **Scope:** Insecure TLS mode.
- **Key Duties:**
  - Support --insecure flag for internal registries.
  - Log warnings for insecure connections.
  - Document security implications.
- **Notes:** Development/testing only.

### A60 — AuthInteropAgent
- **Scope:** Cross-registry authentication.
- **Key Duties:**
  - Test auth with all registry types.
  - Handle auth failures gracefully.
  - Provide actionable error messages.
- **Notes:** Integration testing focus.

---

## G. Encryption & Key Management Agents (A61–A70)

### A61 — EncryptionManagerAgent
- **Scope:** Encryption orchestration.
- **Key Duties:**
  - Coordinate encryption providers.
  - Support envelope encryption.
  - Handle streaming encryption.
- **Notes:** pkg/security/encryption/manager.go

### A62 — AwsKmsAgent
- **Scope:** AWS KMS integration.
- **Key Duties:**
  - Encrypt/decrypt with KMS keys.
  - Generate data keys for envelope encryption.
  - Support customer-managed keys.
- **Notes:** pkg/security/encryption/aws_kms.go

### A63 — GcpKmsAgent
- **Scope:** Google Cloud KMS integration.
- **Key Duties:**
  - Encrypt/decrypt with Cloud KMS.
  - Generate and manage data keys.
  - Support key versioning.
- **Notes:** pkg/security/encryption/gcp_kms.go

### A64 — EnvelopeEncryptionAgent
- **Scope:** Envelope encryption pattern.
- **Key Duties:**
  - Encrypt data with data keys.
  - Encrypt data keys with master keys.
  - Embed encrypted data key with ciphertext.
- **Notes:** Industry-standard pattern.

### A65 — StreamEncryptionAgent
- **Scope:** Streaming encryption for large data.
- **Key Duties:**
  - Encrypt data in chunks.
  - Maintain encryption context.
  - Handle chunked decryption.
- **Notes:** Memory-efficient for large layers.

### A66 — KeyRotationAgent
- **Scope:** Encryption key rotation.
- **Key Duties:**
  - Support multiple key versions.
  - Re-encrypt with new keys.
  - Track key usage metadata.
- **Notes:** Supports compliance requirements.

### A67 — EncryptionProviderAgent
- **Scope:** Encryption provider interface.
- **Key Duties:**
  - Define common encryption operations.
  - Support multiple KMS backends.
  - Enable testing with mocks.
- **Notes:** pkg/security/encryption/types.go

### A68 — SecretsManagerAgent
- **Scope:** Secrets management integration.
- **Key Duties:**
  - Load secrets from AWS Secrets Manager.
  - Load secrets from GCP Secret Manager.
  - Cache secrets with TTL.
- **Notes:** pkg/secrets/provider.go

### A69 — AwsSecretsAgent
- **Scope:** AWS Secrets Manager.
- **Key Duties:**
  - Get/put secrets in AWS.
  - Handle JSON secrets.
  - Support secret rotation.
- **Notes:** pkg/secrets/aws/provider.go

### A70 — GcpSecretsAgent
- **Scope:** Google Secret Manager.
- **Key Duties:**
  - Get/put secrets in GCP.
  - Handle versioned secrets.
  - Support automatic replication.
- **Notes:** pkg/secrets/gcp/provider.go

---

## H. Caching & Performance Agents (A71–A80)

### A71 — LruCacheAgent
- **Scope:** Generic LRU cache implementation.
- **Key Duties:**
  - Provide thread-safe LRU cache.
  - Support generic key/value types.
  - Efficient O(1) operations.
- **Notes:** pkg/cache/lru_cache.go

### A72 — HighPerformanceCacheAgent
- **Scope:** Multi-tier high-performance cache.
- **Key Duties:**
  - Cache manifests, blobs, and tags separately.
  - TTL-based expiration.
  - Memory limit enforcement.
- **Notes:** pkg/cache/high_performance_cache.go

### A73 — ManifestCacheAgent
- **Scope:** Manifest caching.
- **Key Duties:**
  - Cache manifest content by digest.
  - Support TTL and size limits.
  - Thread-safe concurrent access.
- **Notes:** 10K default capacity.

### A74 — BlobCacheAgent
- **Scope:** Blob metadata caching.
- **Key Duties:**
  - Cache blob existence and metadata.
  - Skip redundant existence checks.
  - Support cross-registry deduplication.
- **Notes:** 50K default capacity.

### A75 — TagCacheAgent
- **Scope:** Tag list caching.
- **Key Duties:**
  - Cache tag lists per repository.
  - Shorter TTL (15 min default).
  - Invalidate on tag operations.
- **Notes:** 5K default capacity.

### A76 — CacheMetricsAgent
- **Scope:** Cache performance metrics.
- **Key Duties:**
  - Track hit/miss rates.
  - Measure cache latencies.
  - Report memory usage.
- **Notes:** Prometheus integration.

### A77 — CacheCleanupAgent
- **Scope:** Cache maintenance.
- **Key Duties:**
  - Background cleanup of expired entries.
  - Memory limit enforcement.
  - Periodic metrics reporting.
- **Notes:** Runs on fixed interval.

### A78 — PerformanceBenchmarkAgent
- **Scope:** Performance benchmarking.
- **Key Duties:**
  - Benchmark critical paths.
  - Detect performance regressions.
  - Track throughput metrics.
- **Notes:** CI performance tests.

### A79 — ResourcePoolAgent
- **Scope:** Resource pooling (buffers, connections).
- **Key Duties:**
  - Pool reusable resources.
  - Manage pool size limits.
  - Handle pool exhaustion gracefully.
- **Notes:** sync.Pool and custom pools.

### A80 — MemoryManagementAgent
- **Scope:** Memory efficiency.
- **Key Duties:**
  - Minimize allocations in hot paths.
  - Reuse buffers and objects.
  - Monitor memory usage.
- **Notes:** Profiling-driven optimization.

---

## I. Network & Transfer Agents (A81–A90)

### A81 — TransferManagerAgent
- **Scope:** Transfer orchestration.
- **Key Duties:**
  - Manage blob and image transfers.
  - Apply compression and encryption.
  - Retry failed transfers.
- **Notes:** pkg/network/transfer.go

### A82 — CompressionAgent
- **Scope:** Data compression.
- **Key Duties:**
  - Support gzip and zlib compression.
  - Adaptive compression thresholds.
  - Streaming compression/decompression.
- **Notes:** pkg/network/compression.go

### A83 — StreamingBufferPoolAgent
- **Scope:** Buffer pool management.
- **Key Duties:**
  - Provide optimized readers/writers (64KB).
  - Memory-efficient streaming copy.
  - Chunked stream processing.
- **Notes:** pkg/network/stream_pool.go

### A84 — RetryLogicAgent
- **Scope:** Exponential backoff retry.
- **Key Duties:**
  - Implement exponential backoff.
  - Configure max attempts and delays.
  - Log retry attempts.
- **Notes:** Handles transient failures.

### A85 — RateLimitingAgent
- **Scope:** Rate limiting.
- **Key Duties:**
  - Enforce bandwidth limits (--bwlimit).
  - Implement token bucket algorithm.
  - Per-worker rate limiting.
- **Notes:** Respects registry rate limits.

### A86 — TimeoutManagementAgent
- **Scope:** Operation timeouts.
- **Key Duties:**
  - Configure context timeouts.
  - Handle timeout errors gracefully.
  - Allow timeout configuration per operation.
- **Notes:** Uses context.WithTimeout.

### A87 — ConnectionPoolAgent
- **Scope:** HTTP connection pooling.
- **Key Duties:**
  - Configure http.Client transport.
  - Manage connection pool size.
  - Handle connection reuse.
- **Notes:** Improves performance.

### A88 — TlsConfigAgent
- **Scope:** TLS configuration.
- **Key Duties:**
  - Configure TLS versions and ciphers.
  - Load CA certificates.
  - Support custom TLS configs.
- **Notes:** Security-hardened defaults.

### A89 — NetworkErrorHandlingAgent
- **Scope:** Network error classification.
- **Key Duties:**
  - Classify network errors (timeout, refused, DNS).
  - Provide actionable error messages.
  - Determine if errors are retryable.
- **Notes:** Improves user experience.

### A90 — BandwidthMonitoringAgent
- **Scope:** Bandwidth tracking.
- **Key Duties:**
  - Measure transfer rates.
  - Track total bytes transferred.
  - Report bandwidth metrics.
- **Notes:** Real-time monitoring.

---

## J. Observability & CI Agents (A91–A100)

### A91 — PrometheusMetricsAgent
- **Scope:** Prometheus metrics exposition.
- **Key Duties:**
  - Expose metrics on /metrics endpoint.
  - Maintain 90%+ coverage.
  - Use standard metric types.
- **Notes:** pkg/metrics/prometheus.go

### A92 — GrafanaDashboardAgent
- **Scope:** Grafana dashboard integration.
- **Key Duties:**
  - Provide dashboard JSON definitions.
  - Visualize key metrics.
  - Alert on anomalies.
- **Notes:** monitoring/grafana-dashboard.json

### A93 — StructuredLoggerAgent
- **Scope:** Structured logging implementation.
- **Key Duties:**
  - Use zap for structured logging.
  - Attach trace IDs and context.
  - Support log level filtering.
- **Notes:** pkg/helper/logger.go

### A94 — TraceIdAgent
- **Scope:** Distributed tracing.
- **Key Duties:**
  - Generate and propagate trace IDs.
  - Attach trace context to logs.
  - Support OpenTelemetry integration.
- **Notes:** UUID v4 for trace IDs.

### A95 — HealthCheckAgent
- **Scope:** Health check endpoints.
- **Key Duties:**
  - Implement /health, /ready, /live.
  - Check dependencies (registries, KMS).
  - Return appropriate HTTP codes.
- **Notes:** Kubernetes-compatible.

### A96 — CiPipelineAgent
- **Scope:** CI/CD pipeline.
- **Key Duties:**
  - Maintain GitHub Actions workflows.
  - Ensure all checks pass.
  - Run lint, test, security scans.
- **Notes:** .github/workflows/ci.yml

### A97 — IntegrationTestAgent
- **Scope:** Integration testing.
- **Key Duties:**
  - Test against real registries.
  - Validate all registry types.
  - Use test accounts safely.
- **Notes:** tests/integration/

### A98 — SecurityScanAgent
- **Scope:** Security scanning.
- **Key Duties:**
  - Run gosec on every commit.
  - Scan for vulnerabilities.
  - Enforce security policies.
- **Notes:** Uses gosec and trivy.

### A99 — CodeCoverageAgent
- **Scope:** Test coverage tracking.
- **Key Duties:**
  - Maintain 85%+ unit test coverage.
  - Track coverage trends.
  - Fail builds on coverage drops.
- **Notes:** Uses go test -coverprofile.

### A100 — InteropHarnessAgent
- **Scope:** Interoperability testing.
- **Key Duties:**
  - Test ECR ↔ GCR ↔ Generic transfers.
  - Validate authentication across types.
  - Ensure manifest compatibility.
- **Notes:** Critical for production-readiness.

---

## Agent Interaction Matrix

| Category | Primary Dependencies | Critical Paths |
|----------|---------------------|----------------|
| **Governance (A01-A10)** | All agents | Documentation, security review |
| **CLI (A11-A20)** | Core (A21-A30), Config (A20) | User input validation |
| **Core (A21-A30)** | Replication (A41-A50), Client (A31-A40) | Job orchestration |
| **Client (A31-A40)** | Auth (A51-A60), Network (A81-A90) | Registry interaction |
| **Replication (A41-A50)** | Client (A31-A40), Cache (A71-A80) | Transfer execution |
| **Auth (A51-A60)** | Client (A31-A40), Secrets (A68-A70) | Credential management |
| **Encryption (A61-A70)** | Network (A81-A90), Secrets (A68-A70) | Data protection |
| **Cache (A71-A80)** | Replication (A41-A50), Metrics (A91) | Performance optimization |
| **Network (A81-A90)** | Client (A31-A40), Encryption (A61-A70) | Data transfer |
| **Observability (A91-A100)** | All agents | Monitoring and testing |

---

## Using This Document

When making changes:

1. **Identify affected agents** from the list above.
2. **Check agent dependencies** in the interaction matrix.
3. **Review agent duties** to ensure compliance.
4. **Update tests** for all affected agents.
5. **Document changes** in commit messages with agent IDs.

Example commit message:
```
feat: add GCP KMS encryption support

Agents: A63 (GcpKmsAgent), A61 (EncryptionManagerAgent),
        A64 (EnvelopeEncryptionAgent), A97 (IntegrationTestAgent)

- Implement Google Cloud KMS provider
- Add envelope encryption with Cloud KMS
- Integration tests with test project
- Update encryption manager to support GCP KMS
```

---

## Maintenance

This document should be updated when:
- New features add responsibilities.
- Agent boundaries change.
- New agent categories emerge.
- Interaction patterns evolve.

The AgentsRosterAgent (A03) owns these updates.
