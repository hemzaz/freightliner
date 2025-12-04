# Freightliner Architecture Documentation

This directory contains the core architecture documentation for the Freightliner container registry replication tool, adapted from the oc-rsync mission brief pattern.

## 📚 Document Overview

### 1. MISSION_BRIEF.md - Production-Readiness Contract

**Purpose:** The primary source of truth for production-readiness standards and behavioral parity requirements.

**Key Sections:**
- Sources of truth hierarchy (registry APIs first)
- No external tools constraint (no docker/skopeo/crane)
- Binary structure and branding
- Behavioral parity requirements
- Architecture and performance constraints
- CI/CD requirements
- High-throughput development methodology
- Phase-based implementation roadmap

**When to consult:**
- Before starting any major feature
- When making architectural decisions
- To understand production-readiness criteria
- To resolve conflicting approaches

### 2. AGENTS.md - Agent Roster (100 Agents)

**Purpose:** Defines 100 specialized agents with clear responsibilities and boundaries.

**Agent Categories:**
- **A01-A10:** Governance & Meta Agents
- **A11-A20:** CLI & Frontend Agents
- **A21-A30:** Core Orchestration Agents
- **A31-A40:** Registry Client & Adapter Agents
- **A41-A50:** Replication Engine Agents
- **A51-A60:** Authentication & Security Agents
- **A61-A70:** Encryption & Key Management Agents
- **A71-A80:** Caching & Performance Agents
- **A81-A90:** Network & Transfer Agents
- **A91-A100:** Observability & CI Agents

**When to consult:**
- Before implementing any feature
- To understand component ownership
- To identify which agents are affected by changes
- For commit message agent mapping

**Example agent mapping in commits:**
```
feat(encryption): add GCP KMS support

Agents: A63 (GcpKmsAgent), A61 (EncryptionManagerAgent),
        A64 (EnvelopeEncryptionAgent), A97 (IntegrationTestAgent)
```

### 3. CLAUDE_IMPLEMENTATION.md - Implementation Guide

**Purpose:** Practical guide for AI implementation agents and human developers.

**Key Sections:**
- Role and objectives
- Core constraints (repeated for emphasis)
- Document hierarchy and conflict resolution
- Change proposal process with agent mapping
- Common implementation patterns (code examples)
- Testing requirements (unit, integration, benchmark)
- Security considerations
- Commit message format
- Quick reference guide

**When to consult:**
- When implementing a feature
- To find code patterns and examples
- For testing best practices
- For commit message formatting

## 🎯 How to Use These Documents

### For New Features

1. **Read MISSION_BRIEF.md** - Understand constraints and requirements
2. **Consult AGENTS.md** - Identify affected agents
3. **Follow CLAUDE_IMPLEMENTATION.md** - Use patterns and examples
4. **Map agents in commits** - Include agent IDs in commit messages

### For Bug Fixes

1. **Identify component** in AGENTS.md
2. **Check behavioral requirements** in MISSION_BRIEF.md
3. **Follow implementation patterns** from CLAUDE_IMPLEMENTATION.md
4. **Include agent IDs** in commit message

### For Architecture Decisions

1. **MISSION_BRIEF.md** - Check production-readiness criteria
2. **AGENTS.md** - Identify affected agents and dependencies
3. **Agent interaction matrix** - Understand component relationships
4. **Document decision** - Update docs with rationale

## 🔧 Agent-Based Development Workflow

### Step 1: Understand the Change

```bash
# Read mission brief for context
cat docs/architecture/MISSION_BRIEF.md

# Identify affected agents
grep -A 5 "ComponentName" docs/architecture/AGENTS.md
```

### Step 2: Map to Agents

Identify which agents are involved:
- Primary agents (directly modified)
- Secondary agents (dependencies)
- Testing agents (validation required)

### Step 3: Implement with Patterns

Use patterns from CLAUDE_IMPLEMENTATION.md:
- Client adapter pattern
- Worker pool pattern
- Retry with backoff
- Structured logging
- Context-aware operations

### Step 4: Test All Affected Agents

```bash
# Unit tests
make test

# Integration tests
make test-integration

# Security scan
make security
```

### Step 5: Commit with Agent Mapping

```bash
git commit -m "feat(scope): description

Agents: A## (AgentName), A## (AgentName), ...

Detailed description of changes..."
```

## 📊 Agent Interaction Examples

### Example 1: Adding ECR Support

**Affected Agents:**
- A32 (EcrClientAgent) - Primary implementation
- A52 (EcrAuthAgent) - Authentication
- A31 (ClientFactoryAgent) - Factory integration
- A35 (RegistryInterfaceAgent) - Interface compliance
- A60 (AuthInteropAgent) - Auth testing
- A97 (IntegrationTestAgent) - Integration testing

**Files Modified:**
- `pkg/client/ecr/client.go`
- `pkg/client/ecr/auth.go`
- `pkg/client/factory.go`
- `tests/integration/ecr_test.go`

### Example 2: Adding Encryption

**Affected Agents:**
- A61 (EncryptionManagerAgent) - Orchestration
- A62 (AwsKmsAgent) or A63 (GcpKmsAgent) - Provider
- A64 (EnvelopeEncryptionAgent) - Pattern implementation
- A65 (StreamEncryptionAgent) - Streaming support
- A81 (TransferManagerAgent) - Integration
- A97 (IntegrationTestAgent) - Testing

**Files Modified:**
- `pkg/security/encryption/manager.go`
- `pkg/security/encryption/aws_kms.go` or `gcp_kms.go`
- `pkg/network/transfer.go`
- `tests/integration/encryption_test.go`

### Example 3: Performance Optimization

**Affected Agents:**
- A71-A80 (Cache agents) - Caching layer
- A43 (WorkerPoolAgent) - Concurrency
- A81-A90 (Network agents) - Transfer optimization
- A78 (PerformanceBenchmarkAgent) - Benchmarking
- A91 (PrometheusMetricsAgent) - Metrics

**Files Modified:**
- `pkg/cache/high_performance_cache.go`
- `pkg/replication/worker_pool.go`
- `pkg/network/transfer.go`
- `benchmarks/`

## 🔍 Quick Agent Lookup Table

| Component | Primary Agents | Files |
|-----------|----------------|-------|
| **CLI** | A11-A20 | `cmd/` |
| **Core Orchestration** | A21-A30 | `pkg/service/` |
| **ECR Client** | A32, A52 | `pkg/client/ecr/` |
| **GCR Client** | A33, A53 | `pkg/client/gcr/` |
| **Generic Client** | A34, A54-A55 | `pkg/client/generic/` |
| **Replication** | A41-A50 | `pkg/service/`, `pkg/replication/` |
| **Encryption** | A61-A70 | `pkg/security/encryption/`, `pkg/secrets/` |
| **Caching** | A71-A80 | `pkg/cache/` |
| **Network** | A81-A90 | `pkg/network/` |
| **Observability** | A91-A100 | `pkg/metrics/`, `pkg/helper/` |

## 📖 Document Relationships

```
MISSION_BRIEF.md
    ├─ Defines production-readiness standards
    ├─ Sets architectural constraints
    └─ Establishes phase-based roadmap
         │
         ├─> AGENTS.md
         │    ├─ Maps features to agent responsibilities
         │    ├─ Defines agent boundaries
         │    └─ Provides agent interaction matrix
         │
         └─> CLAUDE_IMPLEMENTATION.md
              ├─ Implements mission brief requirements
              ├─ Uses agent-based architecture
              └─ Provides practical patterns
```

## 🚀 Getting Started

### For Developers

1. **Read all three documents** in order (takes ~30 minutes)
2. **Bookmark this README** for quick reference
3. **Use agent IDs in commits** from day one
4. **Check agent responsibilities** before making changes

### For AI Implementation Agents

1. **ALWAYS consult MISSION_BRIEF.md** for constraints
2. **ALWAYS map changes to agents** in AGENTS.md
3. **ALWAYS use patterns** from CLAUDE_IMPLEMENTATION.md
4. **ALWAYS include agent IDs** in commit messages

### For Code Reviewers

1. **Verify agent mapping** in commit messages
2. **Check against mission brief** requirements
3. **Validate agent boundaries** aren't violated
4. **Ensure test coverage** for affected agents

## 🔄 Maintenance

These documents should be updated when:

- New features add agent responsibilities
- Agent boundaries change
- New architectural patterns emerge
- Production-readiness criteria evolve

**Owner:** A03 (AgentsRosterAgent) and A02 (MissionBriefAgent)

## 📚 Additional Resources

- **GitHub Repository:** https://github.com/hemzaz/freightliner
- **Docker Registry API V2:** https://docs.docker.com/registry/spec/api/
- **OCI Distribution Spec:** https://github.com/opencontainers/distribution-spec
- **AWS ECR API:** https://docs.aws.amazon.com/ecr/
- **Google Artifact Registry:** https://cloud.google.com/artifact-registry/docs

---

**Remember:** Agent-based architecture provides clarity, accountability, and maintainability. Use it consistently for best results! 🚂✨
