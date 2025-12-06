# Freightliner Native Client Architecture

## System Architecture Overview

```
┌───────────────────────────────────────────────────────────────────────┐
│                         Freightliner CLI                              │
│                      (Pure Go Application)                            │
└─────────────────────────────────┬─────────────────────────────────────┘
                                  │
                                  ▼
┌───────────────────────────────────────────────────────────────────────┐
│                    Client Factory (factory.go)                        │
│              Auto-detection & Configuration-based                     │
│                    Client Instantiation                               │
└─────────────┬─────────────────────────────────────────────────────────┘
              │
              ├─────────────────┬──────────────────┬──────────────┐
              ▼                 ▼                  ▼              ▼
    ┌─────────────────┐ ┌─────────────┐ ┌──────────────┐ ┌────────────┐
    │   AWS ECR       │ │ Google GCR  │ │  Azure ACR   │ │   Others   │
    │   Client        │ │   Client    │ │    Client    │ │  (5 more)  │
    └────────┬────────┘ └──────┬──────┘ └──────┬───────┘ └─────┬──────┘
             │                 │                │               │
             ▼                 ▼                ▼               ▼
    ┌──────────────────────────────────────────────────────────────────┐
    │               interfaces.RegistryClient                          │
    │    (Common Interface for All Registry Operations)               │
    │                                                                  │
    │  • ListRepositories(prefix)                                     │
    │  • GetRepository(name)                                          │
    │  • CreateRepository(name, tags)                                 │
    │  • GetTransport(repo)                                           │
    │  • GetRemoteOptions()                                           │
    └─────────────────────────┬────────────────────────────────────────┘
                              │
                              ▼
    ┌──────────────────────────────────────────────────────────────────┐
    │              Common Client Infrastructure                        │
    │                  (pkg/client/common/)                            │
    │                                                                  │
    │  • BaseClient - HTTP transport, retry logic                     │
    │  • BaseRepository - Manifest operations                         │
    │  • EnhancedClient - Advanced features                           │
    │  • Registry utilities                                           │
    └─────────────────────────┬────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
    ┌─────────────┐  ┌──────────────┐  ┌─────────────┐
    │  Manifest   │  │   Layer      │  │    Auth     │
    │  Operations │  │   Transfer   │  │  Handling   │
    └──────┬──────┘  └──────┬───────┘  └──────┬──────┘
           │                │                  │
           └────────────────┼──────────────────┘
                            ▼
    ┌──────────────────────────────────────────────────────────────────┐
    │            go-containerregistry Library                          │
    │         (github.com/google/go-containerregistry)                 │
    │                                                                  │
    │  • OCI/Docker v2 protocol implementation                        │
    │  • Manifest parsing and validation                              │
    │  • Layer streaming and verification                             │
    │  • Authentication transport                                     │
    └─────────────────────────┬────────────────────────────────────────┘
                              │
                              ▼
    ┌──────────────────────────────────────────────────────────────────┐
    │                  Native Go SDKs                                  │
    │                                                                  │
    │  ┌────────────┐  ┌──────────────┐  ┌───────────────┐          │
    │  │  AWS SDK   │  │   GCP SDK    │  │   Azure SDK   │          │
    │  │  (Go v2)   │  │   (Go)       │  │   (Go)        │          │
    │  └──────┬─────┘  └──────┬───────┘  └───────┬───────┘          │
    │         │                │                  │                   │
    │         └────────────────┼──────────────────┘                   │
    └─────────────────────────┼────────────────────────────────────────┘
                              │
                              ▼
    ┌──────────────────────────────────────────────────────────────────┐
    │              Container Registry HTTP APIs                        │
    │           (Docker Registry v2 / OCI Distribution)                │
    │                                                                  │
    │  • ECR: *.dkr.ecr.*.amazonaws.com                              │
    │  • GCR: gcr.io, *.gcr.io, *.pkg.dev                            │
    │  • ACR: *.azurecr.io                                           │
    │  • Docker Hub: registry-1.docker.io                            │
    │  • GHCR: ghcr.io                                               │
    │  • Harbor: Custom deployment                                    │
    │  • Quay: quay.io                                               │
    │  • Generic: Any OCI-compliant registry                         │
    └──────────────────────────────────────────────────────────────────┘
```

## Authentication Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Registry Type Detection                      │
│              (Auto-detect from URL or config)                   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│     ECR      │    │     GCR      │    │     ACR      │
│  IAM/STS     │    │  OAuth2/ADC  │    │ MI/SP/Azure  │
│ Credentials  │    │  Service Acc │    │     Auth     │
└──────┬───────┘    └──────┬───────┘    └──────┬───────┘
       │                   │                    │
       └───────────────────┼────────────────────┘
                           ▼
              ┌─────────────────────────┐
              │  Token/Credentials      │
              │    Acquisition          │
              │   (Native SDK)          │
              └────────────┬────────────┘
                           │
                           ▼
              ┌─────────────────────────┐
              │   HTTP Transport        │
              │  with Auth Headers      │
              └────────────┬────────────┘
                           │
                           ▼
              ┌─────────────────────────┐
              │    Registry API Call    │
              │  (Authenticated)        │
              └─────────────────────────┘
```

## Manifest and Layer Operations

```
┌─────────────────────────────────────────────────────────────────┐
│                    Replication Request                          │
│               (Source Registry → Destination)                   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Step 1: Fetch Manifest                        │
│                                                                 │
│  Source Registry ────── GET /v2/{repo}/manifests/{ref} ───────┐│
│                                                                 ││
│  ┌─────────────────────────────────────────────────────────┐  ││
│  │  Supported Formats:                                     │  ││
│  │  • Docker Manifest V2 Schema 2                         │  ││
│  │  • OCI Image Manifest                                  │  ││
│  │  • OCI Image Index (multi-arch)                        │  ││
│  │  • Docker Manifest List (multi-arch)                   │  ││
│  └─────────────────────────────────────────────────────────┘  ││
│                                                                 ││
└─────────────────────────────────────────────────────────────────┘│
                            │                                      │
                            ▼                                      │
┌─────────────────────────────────────────────────────────────────┐│
│                   Step 2: Parse Manifest                        ││
│                                                                 ││
│  ┌─────────────────────────────────────────────────────────┐  ││
│  │  Extract:                                               │  ││
│  │  • Config digest                                        │  ││
│  │  • Layer digests (sha256:...)                          │  ││
│  │  • Media types                                          │  ││
│  │  • Platform information (arch/os)                      │  ││
│  └─────────────────────────────────────────────────────────┘  ││
└─────────────────────────────────────────────────────────────────┘│
                            │                                      │
                            ▼                                      │
┌─────────────────────────────────────────────────────────────────┐│
│             Step 3: Check Destination for Existing Blobs        ││
│                                                                 ││
│  Destination ─── HEAD /v2/{repo}/blobs/{digest} ──────────┐   ││
│                                                             │   ││
│  ┌──────────────────────────────────────────────────────┐  │   ││
│  │  Skip transfer if:                                   │  │   ││
│  │  • Blob already exists (200 OK)                      │  │   ││
│  │  • Digest matches                                    │  │   ││
│  │  • Attempt blob mount from other repo                │  │   ││
│  └──────────────────────────────────────────────────────┘  │   ││
└─────────────────────────────────────────────────────────────────┘│
                            │                                      │
                            ▼                                      │
┌─────────────────────────────────────────────────────────────────┐│
│                Step 4: Transfer Layers                          ││
│                    (Concurrent)                                 ││
│                                                                 ││
│  ┌────────────┐  ┌────────────┐  ┌────────────┐              ││
│  │  Layer 1   │  │  Layer 2   │  │  Layer N   │              ││
│  │  Transfer  │  │  Transfer  │  │  Transfer  │              ││
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘              ││
│        │                │                │                     ││
│        └────────────────┼────────────────┘                     ││
│                         │                                      ││
│  For each layer:                                              ││
│  1. Stream from source: GET /v2/{repo}/blobs/{digest}        ││
│  2. Verify checksum during transfer                          ││
│  3. Upload to destination: POST/PUT /v2/{repo}/blobs/uploads ││
│  4. Commit with digest                                       ││
└─────────────────────────────────────────────────────────────────┘│
                            │                                      │
                            ▼                                      │
┌─────────────────────────────────────────────────────────────────┐│
│                   Step 5: Push Manifest                         ││
│                                                                 ││
│  Destination ─── PUT /v2/{repo}/manifests/{ref} ───────────┐  ││
│                                                              │  ││
│  • Include all layer references                             │  ││
│  • Set correct Content-Type                                 │  ││
│  • Sign if required (Cosign integration)                    │  ││
└─────────────────────────────────────────────────────────────────┘│
                            │                                      │
                            ▼                                      │
┌─────────────────────────────────────────────────────────────────┐│
│                    Replication Complete                         ││
│                                                                 ││
│  ✓ All layers transferred                                      ││
│  ✓ Manifest pushed                                             ││
│  ✓ Checksums verified                                          ││
│  ✓ Image available at destination                             ││
└─────────────────────────────────────────────────────────────────┘│
```

## Client Type Hierarchy

```
┌────────────────────────────────────────────────────────────────┐
│                    interfaces.RegistryClient                   │
│                      (Core Interface)                          │
└──────────────────────────┬─────────────────────────────────────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
          ▼                ▼                ▼
    ┌──────────┐    ┌───────────┐    ┌──────────┐
    │  Cloud   │    │  Public   │    │ Private  │
    │ Provider │    │ Registry  │    │ Registry │
    │ Specific │    │ Services  │    │ Services │
    └────┬─────┘    └─────┬─────┘    └─────┬────┘
         │                │                 │
    ┌────┼────┐      ┌────┼────┐      ┌────┼────┐
    ▼    ▼    ▼      ▼    ▼    ▼      ▼    ▼    ▼
  ┌───┐┌───┐┌───┐  ┌───┐┌───┐┌───┐  ┌───┐┌───┐┌───┐
  │ECR││GCR││ACR│  │Hub││GHR││Any│  │Hrb││Quy││Reg│
  └───┘└───┘└───┘  └───┘└───┘└───┘  └───┘└───┘└───┘
   AWS  GCP Azure   Dkr  Git  OCI   Hrbr Quay Gnrc
```

**Legend**:
- ECR: Amazon Elastic Container Registry
- GCR: Google Container Registry
- ACR: Azure Container Registry
- Hub: Docker Hub
- GHR: GitHub Container Registry
- Any: Any OCI-compliant registry
- Hrb: Harbor
- Quy: Quay.io
- Reg: Generic Docker v2 registry

## Data Flow: Image Replication

```
┌──────────┐
│   CLI    │
│ Command  │
└─────┬────┘
      │
      ▼
┌──────────────────────┐
│  Config Loader       │
│  • Parse YAML        │
│  • Validate          │
│  • Expand env vars   │
└─────────┬────────────┘
          │
          ▼
┌──────────────────────┐
│  Client Factory      │
│  • Auto-detect type  │
│  • Create clients    │
│  • Setup auth        │
└─────────┬────────────┘
          │
          ├─────────────────┬─────────────────┐
          ▼                 ▼                 ▼
    ┌──────────┐      ┌──────────┐     ┌──────────┐
    │  Source  │      │   Work   │     │   Dest   │
    │  Client  │─────▶│  Queue   │────▶│  Client  │
    └──────────┘      └──────────┘     └──────────┘
          │                 │                 │
          │        ┌────────┴────────┐        │
          │        ▼                 ▼        │
          │   ┌─────────┐      ┌─────────┐   │
          │   │Worker 1 │      │Worker N │   │
          │   └─────────┘      └─────────┘   │
          │        │                 │        │
          └────────┼─────────────────┼────────┘
                   │                 │
                   ▼                 ▼
              ┌─────────────────────────┐
              │   Progress Tracking     │
              │   • Metrics             │
              │   • Logging             │
              │   • Status updates      │
              └─────────────────────────┘
```

## Technology Stack

```
┌───────────────────────────────────────────────────────────┐
│                    Application Layer                      │
│                                                           │
│  • CLI (cobra)                                           │
│  • Configuration (yaml)                                  │
│  • Logging (structured)                                  │
│  • Metrics (prometheus)                                  │
└────────────────────────┬──────────────────────────────────┘
                         │
┌────────────────────────┴──────────────────────────────────┐
│                    Client Layer                           │
│                                                           │
│  • Registry clients (custom + sdks)                      │
│  • Factory pattern                                       │
│  • Interface abstraction                                 │
└────────────────────────┬──────────────────────────────────┘
                         │
┌────────────────────────┴──────────────────────────────────┐
│                 Container Library Layer                   │
│                                                           │
│  • go-containerregistry (Google)                         │
│  • Manifest operations                                   │
│  • Layer streaming                                       │
│  • OCI specs                                             │
└────────────────────────┬──────────────────────────────────┘
                         │
┌────────────────────────┴──────────────────────────────────┐
│                   Cloud SDK Layer                         │
│                                                           │
│  • AWS SDK Go v2                                         │
│  • Google Cloud SDK                                      │
│  • Azure SDK                                             │
│  • Native authentication                                 │
└────────────────────────┬──────────────────────────────────┘
                         │
┌────────────────────────┴──────────────────────────────────┐
│                  HTTP/Network Layer                       │
│                                                           │
│  • net/http (Go standard)                                │
│  • TLS/SSL                                               │
│  • HTTP/2 support                                        │
│  • Connection pooling                                    │
└────────────────────────┬──────────────────────────────────┘
                         │
┌────────────────────────┴──────────────────────────────────┐
│               Container Registry APIs                     │
│                                                           │
│  • Docker Registry v2                                    │
│  • OCI Distribution Spec                                 │
│  • Cloud-specific extensions                             │
└───────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Pure Go Implementation
- ✅ No external binary dependencies
- ✅ Cross-platform compatibility
- ✅ Type-safe error handling
- ✅ Easy testing and mocking

### 2. Interface-Based Design
- ✅ `interfaces.RegistryClient` abstraction
- ✅ Easy to add new registry types
- ✅ Testable with mocks
- ✅ Consistent API across all registries

### 3. Factory Pattern
- ✅ Centralized client creation
- ✅ Auto-detection of registry types
- ✅ Configuration-driven instantiation
- ✅ Credential management

### 4. Native SDK Integration
- ✅ Direct cloud provider SDKs
- ✅ Native authentication flows
- ✅ Automatic token management
- ✅ Regional awareness

### 5. go-containerregistry Library
- ✅ Industry-standard implementation
- ✅ Full OCI/Docker v2 support
- ✅ Active Google maintenance
- ✅ Battle-tested in production

## Security Architecture

```
┌────────────────────────────────────────────────────────────┐
│                     Credential Sources                     │
│                                                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │   IAM    │  │  OAuth2  │  │ Managed  │  │  Config  │ │
│  │  Role    │  │   ADC    │  │ Identity │  │   File   │ │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘ │
└───────┼─────────────┼─────────────┼─────────────┼────────┘
        │             │             │             │
        └─────────────┼─────────────┼─────────────┘
                      ▼             ▼
          ┌────────────────────────────────┐
          │    Native SDK Auth Handler     │
          │  • Token acquisition          │
          │  • Automatic refresh          │
          │  • No disk I/O                │
          └────────────┬───────────────────┘
                       │
                       ▼
          ┌────────────────────────────────┐
          │   In-Memory Token Storage      │
          │  • No temp files              │
          │  • No env var exposure        │
          │  • Process memory only        │
          └────────────┬───────────────────┘
                       │
                       ▼
          ┌────────────────────────────────┐
          │   HTTP Transport with Auth     │
          │  • Authorization headers      │
          │  • TLS encryption             │
          │  • No credential logging      │
          └────────────────────────────────┘
```

## Error Handling Flow

```
┌────────────────────────────────────────────┐
│         Operation Attempt                  │
└──────────────┬─────────────────────────────┘
               │
               ▼
         ┌──────────┐
         │ Success? │────Yes───▶ Return result
         └────┬─────┘
              │ No
              ▼
    ┌────────────────────┐
    │  Transient Error?  │────Yes───┐
    │  (4xx/5xx, timeout)│          │
    └─────┬──────────────┘          │
          │ No                       │
          │                          ▼
          │              ┌─────────────────────┐
          │              │   Retry with        │
          │              │   Exponential       │
          │              │   Backoff           │
          │              └──────────┬──────────┘
          │                         │
          │                         └──────▶ Loop
          │
          ▼
    ┌──────────────────────┐
    │  Wrap Error with     │
    │  Context             │
    │  • Operation type    │
    │  • Registry          │
    │  • Image reference   │
    └─────┬────────────────┘
          │
          ▼
    ┌──────────────────────┐
    │  Return Typed Error  │
    │  • NotFound          │
    │  • Unauthorized      │
    │  • InvalidInput      │
    │  • Internal          │
    └──────────────────────┘
```

## Conclusion

The Freightliner architecture demonstrates a **mature, production-ready implementation** using:

- ✅ 100% native Go code (no external tools)
- ✅ Clean interfaces and abstractions
- ✅ Comprehensive error handling
- ✅ Strong security practices
- ✅ High-performance concurrent operations
- ✅ Extensive test coverage

**No architectural changes required.**

---

**Document Version**: 1.0
**Date**: 2025-12-05
**Status**: Complete ✅
