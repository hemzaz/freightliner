# Version Mismatch Analysis - Freightliner Project

## Overview
This analysis identifies version inconsistencies across the codebase, development tools, and CI environment that may cause build failures and import resolution issues.

## Critical Version Mismatches Found

### 1. **Go Version Inconsistencies** 🔴 CRITICAL
- **Local Development**: Go 1.24.5 (go1.24.5 darwin/arm64)
- **CI Environment**: Go 1.24 (.github/workflows/ci.yml:20,71,134)
- **go.mod Declaration**: `go 1.23.0` with `toolchain go1.24.5`
- **Impact**: Import path resolution failures in CI ("package freightliner/pkg/secrets is not in std")

### 2. **golangci-lint Version Mismatch** 🔴 CRITICAL
- **Local Installation**: v2.0.2 (built with go1.24.1)
- **CI Action**: v1.61.0 (.github/workflows/ci.yml:39)
- **CI Fallback**: v1.61.0 (.github/workflows/ci.yml:47) 
- **Impact**: Schema validation errors, configuration incompatibility

### 3. **Tool Installation Strategy Issues** 🟡 MEDIUM
- **Makefile**: Uses `@latest` for all tool installations (lines 42,57,58,59,60,61)
- **CI**: Pins specific versions for some tools
- **Impact**: Non-deterministic builds, potential breakage with tool updates

## Detailed Version Inventory

### Core Go Environment
```
Local Go:        1.24.5 darwin/arm64
CI Go:           1.24 (GitHub Actions)
go.mod:          1.23.0 + toolchain go1.24.5
```

### Development Tools
```
golangci-lint:   Local v2.0.2 vs CI v1.61.0
staticcheck:     2025.1.1 (0.6.1)
goimports:       From golang.org/x/tools v0.29.0
shadow:          From golang.org/x/tools @latest
Docker:          28.3.2 build 578ccf6
```

### GitHub Actions Versions
```
checkout:        v4 (latest)
setup-go:        v5 (latest)  
cache:           v4 (latest)
upload-artifact: v4 (latest)
golangci-lint-action: v6 (latest)
```

## Root Cause Analysis

### Primary Issue: Module Resolution Failure
The error `package freightliner/pkg/secrets is not in std` indicates Go is not recognizing the local module properly in CI. This is likely caused by:

1. **Go Version Mismatch**: Different Go versions handle module resolution differently
2. **Module Mode Configuration**: Missing or inconsistent GO111MODULE settings
3. **Toolchain Declaration**: go.mod declares toolchain 1.24.5 but CI uses 1.24

### Secondary Issues: Configuration Drift
- golangci-lint v2.0.2 uses different configuration schema than v1.61.0
- Tool versions installed with @latest create non-reproducible builds
- Local development environment differs significantly from CI

## Recommended Solutions

### Phase 1: Immediate Fixes (High Priority)

#### 1. Align Go Versions
**Option A: Standardize on Go 1.23** (Conservative)
```yaml
# .github/workflows/ci.yml
go-version: '1.23'
```
```go
// go.mod  
go 1.23
// Remove toolchain directive
```

**Option B: Standardize on Go 1.24** (Current)
```yaml  
# .github/workflows/ci.yml
go-version: '1.24.5'  # Match local exactly
```
```go
// go.mod
go 1.24
toolchain go1.24.5
```

#### 2. Pin golangci-lint Version
**Local Installation:**
```bash
# Install specific version locally
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
```

**CI Configuration:**
```yaml
# Keep CI at v1.61.0 or upgrade both to v2.0.2
version: v1.61.0  # or v2.0.2 for both
```

#### 3. Fix Module Resolution in CI
```yaml
# Add explicit module configuration
env:
  GO111MODULE: on
  GOFLAGS: -mod=readonly
  GOROOT: ${{ steps.setup-go.outputs.go-root }}
  GOPATH: ${{ steps.setup-go.outputs.go-path }}
```

### Phase 2: Tool Version Pinning (Medium Priority)

#### Update Makefile to Pin Versions
```makefile
# Replace @latest with specific versions
GOIMPORTS_VERSION = v0.29.0
SHADOW_VERSION = v0.29.0  
STATICCHECK_VERSION = 2025.1.1
GOLANGCI_LINT_VERSION = v1.61.0

setup:
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@$(SHADOW_VERSION)
	go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
```

#### Create Tool Version Lock File
```json
// tools.json
{
  "go": "1.24.5",
  "golangci-lint": "v1.61.0", 
  "staticcheck": "2025.1.1",
  "goimports": "v0.29.0",
  "shadow": "v0.29.0"
}
```

### Phase 3: Development Environment Standardization (Low Priority)

#### 1. Add Development Container Support
```dockerfile
# .devcontainer/Dockerfile
FROM golang:1.24.5-alpine
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
# ... other tools
```

#### 2. Create Version Check Script
```bash
#!/bin/bash
# scripts/check-versions.sh
./scripts/version-check.sh --strict
```

## Implementation Priority

### Immediate (Today)
1. ✅ **Fix go.mod Go version** - Choose 1.23 or 1.24 consistently
2. ✅ **Update CI Go version** - Match local environment  
3. ✅ **Add module environment variables** - Fix import resolution

### This Week  
4. **Pin golangci-lint version** - Eliminate schema conflicts
5. **Update tool installations** - Remove @latest dependencies
6. **Test version alignment** - Verify CI passes consistently

### Next Sprint
7. **Create version lock file** - Centralize version management
8. **Add version checking** - Prevent future drift
9. **Document version policy** - Establish update procedures

## Testing Strategy

### Verification Steps
1. **Local Build**: `make check` should pass with pinned versions
2. **CI Consistency**: All CI jobs should use identical tool versions  
3. **Import Resolution**: Module imports should resolve correctly in all environments
4. **Schema Validation**: golangci-lint configuration should validate in all environments

### Success Criteria
- ✅ CI builds complete without import resolution errors
- ✅ golangci-lint runs with consistent configuration schema
- ✅ Local and CI environments produce identical results
- ✅ Tool versions are pinned and reproducible

## Long-term Recommendations

1. **Adopt Renovate/Dependabot** for automated dependency updates
2. **Implement version matrix testing** for Go version compatibility
3. **Create development environment documentation** with exact version requirements
4. **Establish version update policy** with testing procedures

---

**Next Action**: Choose Go version standardization strategy (1.23 vs 1.24) and implement Phase 1 fixes.