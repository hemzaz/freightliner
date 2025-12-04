# Freightliner Architecture Configuration

## Architecture Documents

Freightliner follows a structured agent-based architecture with clear responsibilities and production-readiness standards.

### Core Architecture Documents

1. **Mission Brief** - `/docs/architecture/MISSION_BRIEF.md`
   - Production-readiness contract
   - Core constraints and behavioral parity requirements
   - High-throughput implementation guidelines
   - Phase-based development roadmap

2. **Agent Roster** - `/docs/architecture/AGENTS.md`
   - 100 specialized agents with clear responsibilities
   - Agent categories:
     - A01-A10: Governance & Meta
     - A11-A20: CLI & Frontend
     - A21-A30: Core Orchestration
     - A31-A40: Registry Client & Adapter
     - A41-A50: Replication Engine
     - A51-A60: Authentication & Security
     - A61-A70: Encryption & Key Management
     - A71-A80: Caching & Performance
     - A81-A90: Network & Transfer
     - A91-A100: Observability & CI
   - Agent interaction matrix
   - Usage guidelines for commit messages

3. **Implementation Guide** - `/docs/architecture/CLAUDE_IMPLEMENTATION.md`
   - Role and objectives for implementation agents
   - Core constraints (no external tools)
   - Document hierarchy and conflict resolution
   - Common implementation patterns
   - Testing requirements
   - Security considerations
   - Commit message format
   - Quick reference guide

### Key Principles

#### 1. No External Tool Dependencies
- **FORBIDDEN:** Spawning docker, skopeo, crane, or other external tools
- **REQUIRED:** Native Go implementations for all registry operations
- All registry interactions must use native client adapters

#### 2. Production-Ready Standards
- 88%+ production-ready status
- 85%+ unit test coverage
- 90%+ Prometheus metrics coverage
- Comprehensive integration testing
- Security scanning on every commit

#### 3. Agent-Based Architecture
Every change must:
- Identify affected agents (e.g., "Agents: A32, A52, A60")
- Map to agent responsibilities
- Include agent IDs in commit messages
- Update tests for all affected agents

#### 4. Layered Design
```
cmd/              → CLI parsing and mode selection
pkg/service/      → Business logic orchestration
pkg/client/       → Registry client adapters (ECR, GCR, Generic)
pkg/replication/  → Worker pools and job execution
pkg/security/     → Encryption, mTLS, secrets
pkg/cache/        → LRU and high-performance caching
pkg/network/      → Transfer, compression, buffering
pkg/metrics/      → Prometheus metrics
pkg/helper/       → Logging, utilities, banner
```

#### 5. Testing Strategy
- **Unit Tests:** 85%+ coverage in each package
- **Integration Tests:** Against real registries (test accounts)
- **Interop Tests:** ECR ↔ GCR ↔ Generic transfers
- **Benchmark Tests:** Performance-critical paths
- **Security Tests:** gosec on every commit

### When Making Changes

1. **Read the Mission Brief** - Understand production-readiness criteria
2. **Identify Agents** - Map your changes to agent responsibilities
3. **Follow Patterns** - Use established patterns from implementation guide
4. **Write Tests** - Unit, integration, and interop tests
5. **Document Changes** - Update docs and include agent IDs in commits

### Example Workflow

```bash
# 1. Understand the change requirements
cat docs/architecture/MISSION_BRIEF.md
cat docs/architecture/AGENTS.md

# 2. Implement with agent awareness
# Example: Adding GCP KMS encryption
# Affected agents: A63, A61, A64, A97

# 3. Write tests
make test

# 4. Validate
make lint
make security
make test-integration

# 5. Commit with agent mapping
git commit -m "feat(encryption): add GCP KMS support

Agents: A63 (GcpKmsAgent), A61 (EncryptionManagerAgent),
        A64 (EnvelopeEncryptionAgent), A97 (IntegrationTestAgent)

Implement Google Cloud KMS encryption provider with envelope encryption..."
```

### Quick Agent Lookup

| Task | Primary Agents | Files |
|------|----------------|-------|
| Add ECR support | A32, A52 | pkg/client/ecr/ |
| Add GCR support | A33, A53 | pkg/client/gcr/ |
| Add encryption | A61-A70 | pkg/security/encryption/ |
| Add caching | A71-A80 | pkg/cache/ |
| Add metrics | A91 | pkg/metrics/ |
| CLI changes | A11-A20 | cmd/ |
| Replication logic | A41-A50 | pkg/service/, pkg/replication/ |

### CI/CD Integration

All changes must pass:
```bash
make fmt          # Code formatting
make lint         # golangci-lint
make test         # Unit tests with race detection
make test-ci      # CI-optimized tests
make security     # gosec security scanning
make build        # Binary build
```

### Resources

- **GitHub:** https://github.com/hemzaz/freightliner
- **Mission Brief:** [docs/architecture/MISSION_BRIEF.md](../../docs/architecture/MISSION_BRIEF.md)
- **Agents Roster:** [docs/architecture/AGENTS.md](../../docs/architecture/AGENTS.md)
- **Implementation Guide:** [docs/architecture/CLAUDE_IMPLEMENTATION.md](../../docs/architecture/CLAUDE_IMPLEMENTATION.md)

---

**IMPORTANT:** Always consult the Mission Brief and Agents roster before making significant changes. When in doubt, follow registry API specifications and prioritize production-readiness over clever solutions.
