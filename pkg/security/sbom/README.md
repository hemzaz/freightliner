# SBOM (Software Bill of Materials) Package

This package provides interfaces and types for generating, managing, and analyzing Software Bill of Materials (SBOMs) with integrated vulnerability scanning capabilities.

## Overview

The SBOM package is designed to integrate with [Syft](https://github.com/anchore/syft) for SBOM generation and [Grype](https://github.com/anchore/grype) for vulnerability scanning, supporting industry-standard formats like SPDX and CycloneDX.

## Features

- **SBOM Generation**: Create SBOMs from container images, directories, and archives
- **Multiple Formats**: Support for SPDX 2.2/2.3 and CycloneDX 1.4/1.5
- **Vulnerability Scanning**: Integrated vulnerability detection and reporting
- **SBOM Comparison**: Compare SBOMs to track changes and new vulnerabilities
- **SBOM Enrichment**: Add additional metadata, licenses, and supply chain information
- **Attestation**: Sign and verify SBOMs using various signing methods
- **Export/Import**: Convert between formats and export to various outputs

## Architecture

### Core Interfaces

1. **SBOMGenerator**: Generates SBOMs from various sources
2. **SBOMExporter**: Exports SBOMs to different formats
3. **VulnerabilityScanner**: Scans SBOMs for known vulnerabilities
4. **SBOMComparer**: Compares SBOMs to detect changes
5. **SBOMEnricher**: Adds additional metadata to SBOMs
6. **SBOMAttestor**: Creates and verifies SBOM attestations

### Main Types

- **SBOM**: Core SBOM structure following SPDX/CycloneDX standards
- **Component**: Individual software component with metadata
- **Dependency**: Dependency relationships between components
- **VulnerabilityReport**: Results of vulnerability scanning
- **Vulnerability**: Individual vulnerability with severity and fix information

## Integration with Syft

### Installation

```bash
# Install Syft
curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin

# Or using Homebrew
brew install syft
```

### Implementation Approach

The package uses Syft as the underlying SBOM generation engine:

```go
import (
    "github.com/anchore/syft/syft"
    "github.com/anchore/syft/syft/source"
    "github.com/anchore/syft/syft/sbom"
)

type syftGenerator struct {
    config SyftConfig
}

func (g *syftGenerator) Generate(ctx context.Context, opts GenerateOptions) (*SBOM, error) {
    // Create source from image/directory/archive
    src, err := source.NewFromImage(...)

    // Generate SBOM using Syft
    s := syft.CreateSBOM(ctx, src, nil)

    // Convert Syft SBOM to our SBOM type
    return convertSyftSBOM(s, opts.Format)
}
```

### Syft Configuration

```go
type SyftConfig struct {
    // Catalogers to use (empty = all)
    Catalogers []string

    // Search configuration
    Search SearchConfig

    // Package configuration
    Packages PackageConfig

    // File metadata configuration
    FileMetadata FileMetadataConfig

    // File classification
    FileClassification FileClassificationConfig

    // Secret search configuration
    Secrets SecretsConfig
}
```

## SBOM Formats

### SPDX (Software Package Data Exchange)

SPDX is an ISO standard (ISO/IEC 5962:2021) for communicating SBOM information.

**Supported Versions:**
- SPDX 2.2 (JSON)
- SPDX 2.3 (JSON)

**Key Features:**
- Standardized license identifiers
- Rich relationship types
- Extensive metadata support
- Strong tooling ecosystem

### CycloneDX

CycloneDX is an OWASP standard designed for security use cases.

**Supported Versions:**
- CycloneDX 1.4 (JSON)
- CycloneDX 1.5 (JSON)

**Key Features:**
- Security-focused design
- Vulnerability information
- Service and dependency tracking
- Lightweight and efficient

## Vulnerability Scanning

### Integration with Grype

```bash
# Install Grype
curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin

# Or using Homebrew
brew install grype
```

### Implementation

```go
import (
    "github.com/anchore/grype/grype"
    "github.com/anchore/grype/grype/db"
)

type grypeScanner struct {
    dbConfig db.Config
}

func (s *grypeScanner) Scan(ctx context.Context, sbom *SBOM, opts ScanOptions) (*VulnerabilityReport, error) {
    // Convert our SBOM to Syft SBOM
    syftSBOM := convertToSyftSBOM(sbom)

    // Load vulnerability database
    store, err := grype.LoadVulnerabilityDB(s.dbConfig)

    // Perform vulnerability scan
    matches := grype.FindVulnerabilities(store, syftSBOM)

    // Convert to our VulnerabilityReport
    return convertGrypeMatches(matches, opts)
}
```

### Vulnerability Database

Grype uses multiple vulnerability databases:
- **NVD**: NIST National Vulnerability Database
- **GitHub Advisory Database**: GitHub security advisories
- **Alpine SecDB**: Alpine Linux security database
- **RHEL/Debian/Ubuntu**: Distribution-specific databases

## Usage Examples

### Generate SBOM from Container Image

```go
generator := NewSyftGenerator()
sbom, err := generator.GenerateFromImage(ctx, "nginx:latest", GenerateOptions{
    Format: FormatSPDX23JSON,
    IncludePackages: true,
    IncludeDependencies: true,
    IncludeLicenses: true,
})
```

### Export SBOM to File

```go
exporter := NewSBOMExporter()
err := exporter.ExportToFile(ctx, sbom, FormatCycloneDX15JSON, "sbom.json")
```

### Scan for Vulnerabilities

```go
scanner := NewGrypeScanner()
report, err := scanner.Scan(ctx, sbom, ScanOptions{
    Severities: []VulnerabilitySeverity{SeverityCritical, SeverityHigh},
    OnlyFixed: true,
})

fmt.Printf("Found %d vulnerabilities\n", report.Summary.Total)
```

### Compare SBOMs

```go
comparer := NewSBOMComparer()
diff, err := comparer.Compare(ctx, baselineSBOM, currentSBOM)

fmt.Printf("Added: %d, Removed: %d, Modified: %d\n",
    diff.Summary.TotalAdded,
    diff.Summary.TotalRemoved,
    diff.Summary.TotalModified)
```

### Sign and Attest SBOM

```go
attestor := NewCosignAttestor()
attestation, err := attestor.CreateAttestation(ctx, sbom, AttestationOptions{
    SigningKey: "/path/to/cosign.key",
    SigningMethod: SigningMethodCosign,
    PredicateType: "https://spdx.dev/Document",
})
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Generate SBOM

on:
  push:
    branches: [ main ]

jobs:
  sbom:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install Syft
        run: |
          curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin

      - name: Generate SBOM
        run: |
          syft packages docker:my-image:latest -o spdx-json=sbom.json

      - name: Scan for vulnerabilities
        run: |
          grype sbom:sbom.json --fail-on critical

      - name: Upload SBOM
        uses: actions/upload-artifact@v3
        with:
          name: sbom
          path: sbom.json
```

### GitLab CI Example

```yaml
sbom-generation:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - apk add curl
    - curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
    - syft packages docker:$CI_REGISTRY_IMAGE:$CI_COMMIT_TAG -o cyclonedx-json=sbom.json
    - curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin
    - grype sbom:sbom.json --fail-on high
  artifacts:
    paths:
      - sbom.json
    reports:
      cyclonedx: sbom.json
```

## Best Practices

### 1. SBOM Generation

- **Generate Early**: Create SBOMs as part of your build process
- **Include All Layers**: Analyze all container image layers
- **Use Specific Tags**: Avoid using `latest` tag for reproducibility
- **Version Your SBOMs**: Track SBOM changes over time

### 2. Vulnerability Scanning

- **Scan Regularly**: Run vulnerability scans on every build
- **Update Database**: Keep vulnerability databases current
- **Set Severity Thresholds**: Define acceptable vulnerability levels
- **Track Fixes**: Monitor which vulnerabilities have available fixes

### 3. SBOM Storage and Management

- **Store with Artifacts**: Keep SBOMs alongside container images
- **Sign SBOMs**: Use digital signatures for authenticity
- **Version Control**: Track SBOM changes in version control
- **Automate Distribution**: Share SBOMs with stakeholders automatically

### 4. Compliance

- **Choose Standards**: Select SPDX or CycloneDX based on requirements
- **Include Licenses**: Always include license information
- **Document Dependencies**: Capture complete dependency graphs
- **Maintain Audit Trail**: Keep records of SBOM generation and scans

## Supply Chain Security

### SLSA Integration

The SBOM package supports [SLSA](https://slsa.dev/) (Supply-chain Levels for Software Artifacts) compliance:

- **Provenance**: Generate provenance attestations
- **Build Integrity**: Verify build reproducibility
- **Source Tracking**: Link components to source repositories

### Sigstore Integration

Use [Sigstore](https://www.sigstore.dev/) for keyless signing:

```go
attestor := NewCosignAttestor()
attestation, err := attestor.CreateAttestation(ctx, sbom, AttestationOptions{
    SigningMethod: SigningMethodCosign,
    // Keyless signing using OIDC
})
```

## Performance Considerations

- **Parallel Scanning**: Scan multiple components concurrently
- **Cache Results**: Cache SBOM generation results
- **Incremental Updates**: Update SBOMs incrementally when possible
- **Database Updates**: Update vulnerability databases off-peak

## Security Considerations

- **Protect Signing Keys**: Store signing keys securely
- **Validate Inputs**: Validate all source inputs before processing
- **Sanitize Outputs**: Ensure SBOM outputs don't leak sensitive data
- **Access Control**: Restrict access to SBOMs containing sensitive information

## Future Enhancements

- [ ] Support for additional SBOM formats (SWID, etc.)
- [ ] Machine learning for vulnerability prediction
- [ ] Automated remediation suggestions
- [ ] Integration with package managers for dependency updates
- [ ] Enhanced supply chain risk scoring
- [ ] Real-time vulnerability monitoring
- [ ] SBOM diff visualization tools
- [ ] Integration with container registries (Harbor, ECR, etc.)

## Resources

- [SPDX Specification](https://spdx.github.io/spdx-spec/)
- [CycloneDX Specification](https://cyclonedx.org/specification/overview/)
- [Syft Documentation](https://github.com/anchore/syft)
- [Grype Documentation](https://github.com/anchore/grype)
- [CISA SBOM Guidelines](https://www.cisa.gov/sbom)
- [NTIA Minimum Elements for SBOM](https://www.ntia.gov/report/2021/minimum-elements-software-bill-materials-sbom)

## Contributing

When implementing SBOM functionality:

1. Follow the defined interfaces
2. Add comprehensive tests
3. Document all public APIs
4. Include examples in documentation
5. Ensure compliance with SBOM standards
6. Consider performance implications
7. Add integration tests with Syft/Grype

## License

This package is part of the Freightliner project and follows the project's licensing terms.
