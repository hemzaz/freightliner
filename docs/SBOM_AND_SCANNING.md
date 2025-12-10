# SBOM Generation and Vulnerability Scanning

Freightliner includes comprehensive Software Bill of Materials (SBOM) generation and vulnerability scanning capabilities to help secure your container image supply chain.

## Features

### SBOM Generation
- Multiple output formats: SPDX, CycloneDX, Syft JSON, human-readable tables
- OS package detection (Debian/Ubuntu, Alpine, RHEL/CentOS)
- Language-specific package detection (npm, pip, Go modules, Maven, Ruby gems)
- Integration with Syft for advanced scanning
- File cataloging (optional)
- Secret detection (optional)

### Vulnerability Scanning
- CVE database scanning using Grype integration
- Severity-based filtering (critical, high, medium, low)
- Fix availability checking
- Policy-based enforcement
- Multiple output formats: JSON, SARIF, human-readable tables
- GitHub Security integration via SARIF format

## Installation

### Prerequisites

For full functionality, install Syft and Grype:

```bash
# macOS (Homebrew)
brew install syft grype

# Linux (curl)
curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin

# Docker
docker pull anchore/syft:latest
docker pull anchore/grype:latest
```

Note: Freightliner includes built-in SBOM generation and scanning capabilities, but Syft and Grype provide more comprehensive analysis.

## SBOM Generation

### Basic Usage

```bash
# Generate SBOM in Syft JSON format
freightliner sbom docker.io/library/alpine:latest

# Generate SBOM in SPDX format
freightliner sbom nginx:latest --format spdx

# Generate SBOM in CycloneDX format
freightliner sbom myregistry.com/app:v1.0.0 --format cyclonedx

# Generate human-readable table
freightliner sbom alpine:latest --format table
```

### Advanced Options

```bash
# Use Syft for advanced scanning
freightliner sbom gcr.io/myproject/app:latest --use-syft

# Include file catalog
freightliner sbom nginx:latest --include-files

# Scan for potential secrets
freightliner sbom myapp:latest --scan-secrets

# Save to file
freightliner sbom alpine:latest --format spdx --output sbom.json

# Exclude specific paths
freightliner sbom myapp:latest --exclude "/tmp/*" --exclude "/var/cache/*"

# Scan all layers (vs squashed image)
freightliner sbom myapp:latest --scope all-layers
```

### SBOM Output Formats

#### SPDX (Software Package Data Exchange)
```bash
freightliner sbom alpine:latest --format spdx --output alpine-sbom.spdx.json
```

SPDX is an open standard for communicating software bill of material information. It's widely used in enterprise environments and regulatory compliance.

#### CycloneDX
```bash
freightliner sbom nginx:latest --format cyclonedx --output nginx-sbom.cdx.json
```

CycloneDX is a lightweight SBOM standard designed for use in application security contexts and supply chain component analysis.

#### Syft JSON
```bash
freightliner sbom myapp:latest --format syft-json --output myapp-sbom.json
```

Syft's native format, providing the most detailed package information.

#### Table
```bash
freightliner sbom alpine:latest --format table
```

Human-readable table format for quick inspection.

## Vulnerability Scanning

### Basic Usage

```bash
# Scan image for vulnerabilities
freightliner scan docker.io/library/nginx:latest

# Fail if critical vulnerabilities found
freightliner scan nginx:latest --fail-on critical

# Fail on high or critical vulnerabilities
freightliner scan myapp:v1.0.0 --fail-on high
```

### Advanced Options

```bash
# Use Grype for scanning
freightliner scan alpine:latest --use-grype

# Only show vulnerabilities with fixes
freightliner scan nginx:latest --only-fixed

# Ignore unfixed vulnerabilities
freightliner scan myapp:latest --ignore-unfixed

# Update vulnerability database before scanning
freightliner scan alpine:latest --db-update

# Generate JSON report
freightliner scan nginx:latest --format json --output scan-results.json

# Generate SARIF report for GitHub
freightliner scan myapp:latest --format sarif --output results.sarif

# Exclude specific paths
freightliner scan myapp:latest --exclude "/test/*" --exclude "/docs/*"

# Scan specific platform
freightliner scan myapp:latest --platform linux/amd64
```

### Severity Levels

Vulnerabilities are classified by severity:
- **Critical**: Actively exploited, immediate action required
- **High**: Severe vulnerabilities that should be addressed urgently
- **Medium**: Moderate vulnerabilities to address in regular maintenance
- **Low**: Minor vulnerabilities with limited impact
- **Negligible**: Very minor issues with minimal risk
- **Unknown**: Severity not yet determined

### Output Formats

#### JSON
```bash
freightliner scan nginx:latest --format json --output results.json
```

Machine-readable format for integration with CI/CD pipelines and security tools.

#### Table (default)
```bash
freightliner scan nginx:latest --format table
```

Human-readable table showing vulnerability details:
```
Vulnerability Scan Report for: nginx:latest
Scan Time: 2024-12-05T10:30:00Z

SUMMARY:
--------------------------------------------------------------------------------
Total Vulnerabilities: 15
  Critical: 2
  High:     5
  Medium:   6
  Low:      2

Total Packages: 142
Vulnerable Packages: 8
Packages with Fixes: 6

VULNERABILITIES:
--------------------------------------------------------------------------------
ID              SEVERITY   PACKAGE                        FIX VERSION
--------------------------------------------------------------------------------
CVE-2024-1234   critical   openssl                        1.1.1w-r0
CVE-2024-5678   high       curl                           8.5.0-r0
...
```

#### SARIF
```bash
freightliner scan nginx:latest --format sarif --output results.sarif
```

SARIF (Static Analysis Results Interchange Format) is GitHub's format for security findings, enabling integration with GitHub Security.

## CI/CD Integration

### GitHub Actions

```yaml
name: Container Security Scan

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install Freightliner
        run: |
          # Install freightliner
          go install github.com/yourorg/freightliner@latest

      - name: Generate SBOM
        run: |
          freightliner sbom myregistry.io/app:${{ github.sha }} \
            --format spdx \
            --output sbom.json

      - name: Upload SBOM
        uses: actions/upload-artifact@v3
        with:
          name: sbom
          path: sbom.json

      - name: Scan for Vulnerabilities
        run: |
          freightliner scan myregistry.io/app:${{ github.sha }} \
            --fail-on high \
            --format sarif \
            --output results.sarif

      - name: Upload SARIF to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        if: always()
        with:
          sarif_file: results.sarif
```

### GitLab CI

```yaml
security:scan:
  stage: security
  image: alpine:latest
  before_script:
    - apk add --no-cache go
    - go install github.com/yourorg/freightliner@latest
  script:
    # Generate SBOM
    - freightliner sbom $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
        --format cyclonedx
        --output sbom.json

    # Scan for vulnerabilities
    - freightliner scan $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
        --fail-on high
        --format json
        --output vulnerabilities.json
  artifacts:
    reports:
      cyclonedx: sbom.json
    paths:
      - vulnerabilities.json
  allow_failure: false
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any

    stages {
        stage('Security Scan') {
            steps {
                script {
                    // Generate SBOM
                    sh """
                        freightliner sbom ${DOCKER_REGISTRY}/${IMAGE_NAME}:${BUILD_NUMBER} \
                            --format spdx \
                            --output sbom.json
                    """

                    // Scan vulnerabilities
                    sh """
                        freightliner scan ${DOCKER_REGISTRY}/${IMAGE_NAME}:${BUILD_NUMBER} \
                            --fail-on critical \
                            --format json \
                            --output scan-results.json
                    """
                }
            }
            post {
                always {
                    archiveArtifacts artifacts: '*.json', fingerprint: true
                }
            }
        }
    }
}
```

## Policy-Based Scanning

Create a vulnerability policy file (`vulnerability-policy.json`):

```json
{
  "max_critical": 0,
  "max_high": 5,
  "max_medium": 20,
  "ignore_cves": [
    "CVE-2024-1234",
    "CVE-2024-5678"
  ],
  "ignore_packages": [
    "test-package",
    "dev-dependency"
  ],
  "require_fix_available": true,
  "custom_rules": [
    {
      "name": "Block GPL licenses in production",
      "severity": "high",
      "action": "deny",
      "conditions": ["license contains 'GPL'"],
      "description": "GPL licenses not allowed in production images"
    }
  ]
}
```

Use the policy:
```bash
freightliner scan myapp:latest --policy vulnerability-policy.json
```

## Integration with Container Replication

Scan images during replication:

```bash
# Option 1: Scan source before replication
freightliner scan docker.io/library/nginx:latest --fail-on high
freightliner replicate docker.io/library/nginx:latest myregistry.io/nginx:latest

# Option 2: Generate SBOM of replicated image
freightliner replicate docker.io/library/nginx:latest myregistry.io/nginx:latest
freightliner sbom myregistry.io/nginx:latest --format spdx --output nginx-sbom.json
```

Future enhancement: Automatic scanning during replication with `--scan-vulnerabilities` flag.

## API Usage

### Programmatic SBOM Generation

```go
package main

import (
    "context"
    "freightliner/pkg/sbom"
    "github.com/google/go-containerregistry/pkg/name"
)

func main() {
    ctx := context.Background()

    // Parse image reference
    ref, _ := name.ParseReference("nginx:latest")

    // Create generator
    generator := sbom.NewGenerator(sbom.GeneratorConfig{
        Format:                  sbom.FormatSPDX,
        IncludeOSPackages:       true,
        IncludeLanguagePackages: true,
    })

    // Generate SBOM
    imageSBOM, err := generator.Generate(ctx, ref)
    if err != nil {
        panic(err)
    }

    // Export to file
    err = generator.WriteToFile(imageSBOM, "sbom.json", sbom.FormatSPDX)
}
```

### Programmatic Vulnerability Scanning

```go
package main

import (
    "context"
    "freightliner/pkg/vulnerability"
    "github.com/google/go-containerregistry/pkg/name"
)

func main() {
    ctx := context.Background()

    // Parse image reference
    ref, _ := name.ParseReference("nginx:latest")

    // Create scanner
    scanner, err := vulnerability.NewScanner(vulnerability.ScanConfig{
        FailOnSeverity: vulnerability.SeverityHigh,
        AutoUpdateDB:   true,
    })
    if err != nil {
        panic(err)
    }

    // Scan image
    report, err := scanner.Scan(ctx, ref)
    if err != nil {
        panic(err)
    }

    // Check results
    if report.Summary.Critical > 0 || report.Summary.High > 0 {
        panic("Critical or high vulnerabilities found!")
    }
}
```

## Best Practices

1. **Always scan before deployment**: Integrate scanning into your CI/CD pipeline
2. **Set appropriate severity thresholds**: Use `--fail-on` to enforce security standards
3. **Keep databases updated**: Use `--db-update` regularly to get latest vulnerability data
4. **Generate SBOMs for compliance**: Many regulations require SBOMs for software supply chain transparency
5. **Archive scan results**: Keep historical records of vulnerability scans
6. **Review unfixed vulnerabilities**: Decide on acceptance or mitigation strategies
7. **Use policy files**: Centralize security policies in configuration files
8. **Integrate with GitHub Security**: Use SARIF format for GitHub Advanced Security features

## Troubleshooting

### Syft not found
```bash
Error: syft not found in PATH

# Install Syft
brew install syft  # macOS
# or
curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
```

### Grype not found
```bash
Error: grype not found in PATH

# Install Grype
brew install grype  # macOS
# or
curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin
```

### Database update fails
```bash
# Manually update Grype database
grype db update

# Check database status
grype db status
```

### Authentication issues
```bash
# For private registries, authenticate first
docker login myregistry.io

# Or use registry config
freightliner scan myregistry.io/private/app:latest --config registry-config.yaml
```

## Related Documentation

- [Container Replication](./REPLICATION.md)
- [Registry Configuration](./REGISTRY_CONFIG.md)
- [CLI Reference](./CLI_REFERENCE.md)
- [API Documentation](./API.md)

## External Resources

- [Syft Documentation](https://github.com/anchore/syft)
- [Grype Documentation](https://github.com/anchore/grype)
- [SPDX Specification](https://spdx.dev/)
- [CycloneDX Specification](https://cyclonedx.org/)
- [SARIF Specification](https://sarifweb.azurewebsites.net/)
- [GitHub Security](https://docs.github.com/en/code-security)
