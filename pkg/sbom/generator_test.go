package sbom

import (
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name   string
		config GeneratorConfig
		want   SBOMFormat
	}{
		{
			name:   "default format",
			config: GeneratorConfig{},
			want:   FormatSyftJSON,
		},
		{
			name: "custom format",
			config: GeneratorConfig{
				Format: FormatSPDX,
			},
			want: FormatSPDX,
		},
		{
			name: "cyclonedx format",
			config: GeneratorConfig{
				Format: FormatCycloneDX,
			},
			want: FormatCycloneDX,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(tt.config)
			assert.NotNil(t, gen)
			assert.Equal(t, tt.want, gen.config.Format)
		})
	}
}

func TestParseDebianStatus(t *testing.T) {
	content := `Package: base-files
Version: 11.1+deb11u2
Architecture: amd64

Package: bash
Version: 5.1-2+deb11u1
Architecture: amd64

Package: libc6
Version: 2.31-13+deb11u3
Architecture: amd64
`

	gen := NewGenerator(GeneratorConfig{})
	packages := gen.parseDebianStatus(content, "layer-123")

	assert.Len(t, packages, 3)

	// Check first package
	assert.Equal(t, "base-files", packages[0].Name)
	assert.Equal(t, "11.1+deb11u2", packages[0].Version)
	assert.Equal(t, "deb", packages[0].Type)
	assert.Equal(t, "layer-123", packages[0].Locations[0].LayerID)

	// Check second package
	assert.Equal(t, "bash", packages[1].Name)
	assert.Equal(t, "5.1-2+deb11u1", packages[1].Version)

	// Check third package
	assert.Equal(t, "libc6", packages[2].Name)
	assert.Equal(t, "2.31-13+deb11u3", packages[2].Version)
}

func TestParseAlpineInstalled(t *testing.T) {
	content := `P:alpine-baselayout
V:3.2.0-r23

P:busybox
V:1.34.1-r3

P:musl
V:1.2.2-r7
`

	gen := NewGenerator(GeneratorConfig{})
	packages := gen.parseAlpineInstalled(content, "layer-456")

	assert.Len(t, packages, 3)

	// Check first package
	assert.Equal(t, "alpine-baselayout", packages[0].Name)
	assert.Equal(t, "3.2.0-r23", packages[0].Version)
	assert.Equal(t, "apk", packages[0].Type)
	assert.Equal(t, "layer-456", packages[0].Locations[0].LayerID)

	// Check second package
	assert.Equal(t, "busybox", packages[1].Name)
	assert.Equal(t, "1.34.1-r3", packages[1].Version)

	// Check third package
	assert.Equal(t, "musl", packages[2].Name)
	assert.Equal(t, "1.2.2-r7", packages[2].Version)
}

func TestExportSyftJSON(t *testing.T) {
	sbom := &SBOM{
		ImageRef: "nginx:latest",
		Format:   FormatSyftJSON,
		Packages: []Package{
			{
				Name:    "nginx",
				Version: "1.21.0",
				Type:    "deb",
			},
			{
				Name:    "openssl",
				Version: "1.1.1k",
				Type:    "deb",
			},
		},
	}

	gen := NewGenerator(GeneratorConfig{})
	data, err := gen.exportSyftJSON(sbom)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify JSON structure
	jsonStr := string(data)
	assert.Contains(t, jsonStr, "nginx:latest")
	assert.Contains(t, jsonStr, "nginx")
	assert.Contains(t, jsonStr, "openssl")
	assert.Contains(t, jsonStr, "1.21.0")
	assert.Contains(t, jsonStr, "1.1.1k")
}

func TestExportSPDX(t *testing.T) {
	sbom := &SBOM{
		ImageRef: "alpine:latest",
		Format:   FormatSPDX,
		Packages: []Package{
			{
				Name:    "musl",
				Version: "1.2.2",
				Type:    "apk",
			},
		},
	}

	gen := NewGenerator(GeneratorConfig{})
	data, err := gen.exportSPDX(sbom)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify SPDX structure
	jsonStr := string(data)
	assert.Contains(t, jsonStr, "SPDX-2.3")
	assert.Contains(t, jsonStr, "alpine:latest")
	assert.Contains(t, jsonStr, "musl")
	assert.Contains(t, jsonStr, "1.2.2")
	assert.Contains(t, jsonStr, "SPDXRef-DOCUMENT")
}

func TestExportCycloneDX(t *testing.T) {
	sbom := &SBOM{
		ImageRef: "node:16",
		Format:   FormatCycloneDX,
		Packages: []Package{
			{
				Name:    "express",
				Version: "4.18.0",
				Type:    "npm",
				PURL:    "pkg:npm/express@4.18.0",
			},
		},
	}

	gen := NewGenerator(GeneratorConfig{})
	data, err := gen.exportCycloneDX(sbom)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify CycloneDX structure
	jsonStr := string(data)
	assert.Contains(t, jsonStr, "CycloneDX")
	assert.Contains(t, jsonStr, "1.4")
	assert.Contains(t, jsonStr, "node:16")
	assert.Contains(t, jsonStr, "express")
	assert.Contains(t, jsonStr, "4.18.0")
}

func TestExportTable(t *testing.T) {
	sbom := &SBOM{
		ImageRef: "ubuntu:20.04",
		Format:   FormatTable,
		Packages: []Package{
			{
				Name:    "apt",
				Version: "2.0.2",
				Type:    "deb",
			},
			{
				Name:    "curl",
				Version: "7.68.0",
				Type:    "deb",
			},
		},
	}

	gen := NewGenerator(GeneratorConfig{})
	data, err := gen.exportTable(sbom)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify table structure
	tableStr := string(data)
	assert.Contains(t, tableStr, "ubuntu:20.04")
	assert.Contains(t, tableStr, "PACKAGES:")
	assert.Contains(t, tableStr, "apt")
	assert.Contains(t, tableStr, "curl")
	assert.Contains(t, tableStr, "2.0.2")
	assert.Contains(t, tableStr, "7.68.0")
	assert.Contains(t, tableStr, "Total Packages: 2")
}

func TestExportUnsupportedFormat(t *testing.T) {
	sbom := &SBOM{
		ImageRef: "test:latest",
	}

	gen := NewGenerator(GeneratorConfig{})
	_, err := gen.Export(sbom, SBOMFormat("invalid-format"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestParseNPMPackage(t *testing.T) {
	// This would require creating a temporary package.json file
	// Skipping for now as it involves filesystem operations
	t.Skip("Requires filesystem operations")
}

func TestSBOMMetadata(t *testing.T) {
	sbom := &SBOM{
		ImageRef: "test:latest",
		Metadata: map[string]string{
			"architecture": "amd64",
			"os":           "linux",
		},
	}

	assert.Equal(t, "test:latest", sbom.ImageRef)
	assert.Equal(t, "amd64", sbom.Metadata["architecture"])
	assert.Equal(t, "linux", sbom.Metadata["os"])
}

func TestPackageLocations(t *testing.T) {
	pkg := Package{
		Name:    "test-package",
		Version: "1.0.0",
		Type:    "npm",
		Locations: []Location{
			{
				Path:    "/app/node_modules/test-package",
				LayerID: "sha256:abc123",
			},
			{
				Path:    "/app/package.json",
				LayerID: "sha256:def456",
			},
		},
	}

	assert.Len(t, pkg.Locations, 2)
	assert.Equal(t, "/app/node_modules/test-package", pkg.Locations[0].Path)
	assert.Equal(t, "sha256:abc123", pkg.Locations[0].LayerID)
}

func TestPackageDependencies(t *testing.T) {
	pkg := Package{
		Name:    "express",
		Version: "4.18.0",
		Type:    "npm",
		Dependencies: []string{
			"body-parser",
			"cookie",
			"debug",
		},
	}

	assert.Len(t, pkg.Dependencies, 3)
	assert.Contains(t, pkg.Dependencies, "body-parser")
	assert.Contains(t, pkg.Dependencies, "cookie")
	assert.Contains(t, pkg.Dependencies, "debug")
}

func TestRelationships(t *testing.T) {
	sbom := &SBOM{
		ImageRef: "test:latest",
		Relationships: []Relationship{
			{
				From: "pkg:npm/express@4.18.0",
				To:   "pkg:npm/body-parser@1.20.0",
				Type: "depends_on",
			},
			{
				From: "image",
				To:   "pkg:npm/express@4.18.0",
				Type: "contains",
			},
		},
	}

	assert.Len(t, sbom.Relationships, 2)
	assert.Equal(t, "depends_on", sbom.Relationships[0].Type)
	assert.Equal(t, "contains", sbom.Relationships[1].Type)
}

func TestSecretDetection(t *testing.T) {
	secret := Secret{
		Type: "api_key",
		Location: Location{
			Path:    "/app/config.json",
			LayerID: "sha256:abc123",
		},
		Match:   "api_key=sk_test_1234567890",
		Context: "Found in configuration file",
	}

	assert.Equal(t, "api_key", secret.Type)
	assert.Contains(t, secret.Match, "sk_test")
	assert.Equal(t, "/app/config.json", secret.Location.Path)
}

func TestSBOMFormats(t *testing.T) {
	formats := []SBOMFormat{
		FormatSPDX,
		FormatCycloneDX,
		FormatSyftJSON,
		FormatTable,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			config := GeneratorConfig{
				Format: format,
			}
			gen := NewGenerator(config)
			assert.Equal(t, format, gen.config.Format)
		})
	}
}

func TestConvertToSPDXPackages(t *testing.T) {
	sbom := &SBOM{
		Packages: []Package{
			{Name: "pkg1", Version: "1.0.0"},
			{Name: "pkg2", Version: "2.0.0"},
		},
	}

	gen := NewGenerator(GeneratorConfig{})
	spdxPkgs := gen.convertToSPDXPackages(sbom)

	assert.Len(t, spdxPkgs, 2)
	assert.Contains(t, spdxPkgs[0], "SPDXID")
	assert.Contains(t, spdxPkgs[0], "name")
	assert.Contains(t, spdxPkgs[0], "versionInfo")
}

func TestConvertToCycloneDXComponents(t *testing.T) {
	sbom := &SBOM{
		Packages: []Package{
			{
				Name:    "express",
				Version: "4.18.0",
				PURL:    "pkg:npm/express@4.18.0",
			},
		},
	}

	gen := NewGenerator(GeneratorConfig{})
	components := gen.convertToCycloneDXComponents(sbom)

	assert.Len(t, components, 1)
	assert.Equal(t, "library", components[0]["type"])
	assert.Equal(t, "express", components[0]["name"])
	assert.Equal(t, "4.18.0", components[0]["version"])
}

func TestGeneratorPackageTypes(t *testing.T) {
	config := GeneratorConfig{}
	gen := NewGenerator(config)

	// Default package types should be set
	assert.NotEmpty(t, gen.config.PackageTypes)
	assert.Contains(t, gen.config.PackageTypes, "deb")
	assert.Contains(t, gen.config.PackageTypes, "rpm")
	assert.Contains(t, gen.config.PackageTypes, "apk")
	assert.Contains(t, gen.config.PackageTypes, "npm")
	assert.Contains(t, gen.config.PackageTypes, "pip")
	assert.Contains(t, gen.config.PackageTypes, "go")
}

func TestParseReferenceError(t *testing.T) {
	// Test with invalid reference
	_, err := name.ParseReference("invalid::reference")
	assert.Error(t, err)
}

func TestEmptyPackageList(t *testing.T) {
	sbom := &SBOM{
		ImageRef: "empty:latest",
		Packages: []Package{},
	}

	gen := NewGenerator(GeneratorConfig{})
	data, err := gen.Export(sbom, FormatTable)
	require.NoError(t, err)

	tableStr := string(data)
	assert.Contains(t, tableStr, "Total Packages: 0")
}

func TestPackageLicenses(t *testing.T) {
	pkg := Package{
		Name:     "express",
		Version:  "4.18.0",
		Licenses: []string{"MIT", "Apache-2.0"},
	}

	assert.Len(t, pkg.Licenses, 2)
	assert.Contains(t, pkg.Licenses, "MIT")
	assert.Contains(t, pkg.Licenses, "Apache-2.0")
}

func TestPackagePURL(t *testing.T) {
	tests := []struct {
		name     string
		pkg      Package
		wantPURL string
	}{
		{
			name: "npm package",
			pkg: Package{
				Name:    "express",
				Version: "4.18.0",
				Type:    "npm",
				PURL:    "pkg:npm/express@4.18.0",
			},
			wantPURL: "pkg:npm/express@4.18.0",
		},
		{
			name: "pip package",
			pkg: Package{
				Name:    "django",
				Version: "4.0.0",
				Type:    "pip",
				PURL:    "pkg:pypi/django@4.0.0",
			},
			wantPURL: "pkg:pypi/django@4.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantPURL, tt.pkg.PURL)
		})
	}
}

func TestTableOutput(t *testing.T) {
	sbom := &SBOM{
		ImageRef: "test:latest",
		Packages: []Package{
			{Name: "pkg1", Version: "1.0", Type: "deb"},
			{Name: "pkg2", Version: "2.0", Type: "rpm"},
		},
	}

	gen := NewGenerator(GeneratorConfig{})
	data, err := gen.exportTable(sbom)
	require.NoError(t, err)

	output := string(data)
	lines := strings.Split(output, "\n")

	// Should have header, separator, column headers, separator, and data rows
	assert.True(t, len(lines) >= 5)
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "VERSION")
	assert.Contains(t, output, "TYPE")
}
