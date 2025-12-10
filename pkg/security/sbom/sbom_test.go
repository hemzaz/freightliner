package sbom

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSBOM_Creation(t *testing.T) {
	now := time.Now()

	sbom := &SBOM{
		Format:      FormatSPDX23JSON,
		SpecVersion: "2.3",
		DocumentID:  "test-doc-123",
		Metadata: SBOMMetadata{
			Name:        "test-component",
			Version:     "1.0.0",
			Description: "Test component",
			Timestamp:   now,
			ToolInfo: ToolInfo{
				Vendor:  "freightliner",
				Name:    "sbom-generator",
				Version: "1.0.0",
			},
			SourceInfo: SourceInfo{
				Type:   SourceTypeImage,
				Target: "nginx:latest",
				Scheme: "docker",
			},
		},
		Components: []Component{},
		CreatedAt:  now,
	}

	assert.Equal(t, FormatSPDX23JSON, sbom.Format)
	assert.Equal(t, "test-doc-123", sbom.DocumentID)
	assert.Equal(t, "test-component", sbom.Metadata.Name)
	assert.Equal(t, now, sbom.CreatedAt)
}

func TestComponent_Creation(t *testing.T) {
	component := Component{
		ID:          "pkg:npm/express@4.18.0",
		Type:        ComponentTypeLibrary,
		Name:        "express",
		Version:     "4.18.0",
		PURL:        "pkg:npm/express@4.18.0",
		Description: "Fast, unopinionated, minimalist web framework",
		Licenses: []License{
			{
				ID:   "MIT",
				Name: "MIT License",
			},
		},
		Supplier:  "npm",
		Publisher: "npm",
		Hashes: map[string]string{
			"sha256": "abc123def456",
		},
		Properties: map[string]string{
			"language": "javascript",
		},
		FoundBy: "npm-cataloger",
		Locations: []Location{
			{
				Path: "/app/node_modules/express",
			},
		},
	}

	assert.Equal(t, "express", component.Name)
	assert.Equal(t, ComponentTypeLibrary, component.Type)
	assert.Equal(t, "4.18.0", component.Version)
	assert.Equal(t, "npm-cataloger", component.FoundBy)
	assert.Len(t, component.Licenses, 1)
	assert.Len(t, component.Locations, 1)
}

func TestDependency_Creation(t *testing.T) {
	dep := Dependency{
		Ref:               "pkg:npm/express@4.18.0",
		DependsOn:         []string{"pkg:npm/body-parser@1.20.0", "pkg:npm/cookie@0.5.0"},
		Scope:             ScopeRuntime,
		Direct:            true,
		VersionConstraint: "^4.18.0",
	}

	assert.Equal(t, "pkg:npm/express@4.18.0", dep.Ref)
	assert.Len(t, dep.DependsOn, 2)
	assert.Equal(t, ScopeRuntime, dep.Scope)
	assert.True(t, dep.Direct)
}

func TestVulnerability_Creation(t *testing.T) {
	vuln := Vulnerability{
		ID:                 "CVE-2023-12345",
		Aliases:            []string{"GHSA-xxxx-yyyy-zzzz"},
		Source:             "NVD",
		Severity:           SeverityHigh,
		CVSSScore:          7.5,
		CVSSVector:         "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N",
		Description:        "SQL injection vulnerability in component X",
		AffectedComponents: []string{"pkg:npm/vulnerable-lib@1.0.0"},
		AffectedVersions:   []string{"< 1.0.1"},
		FixedVersions:      []string{">= 1.0.1"},
		URLs: []string{
			"https://nvd.nist.gov/vuln/detail/CVE-2023-12345",
		},
		PublishedDate:    time.Now().Add(-30 * 24 * time.Hour),
		ModifiedDate:     time.Now().Add(-1 * 24 * time.Hour),
		ExploitAvailable: false,
		FixAvailable:     true,
		Remediation:      "Update to version 1.0.1 or later",
	}

	assert.Equal(t, "CVE-2023-12345", vuln.ID)
	assert.Equal(t, SeverityHigh, vuln.Severity)
	assert.Equal(t, 7.5, vuln.CVSSScore)
	assert.True(t, vuln.FixAvailable)
	assert.False(t, vuln.ExploitAvailable)
	assert.Len(t, vuln.AffectedComponents, 1)
}

func TestVulnerabilitySeverity_String(t *testing.T) {
	tests := []struct {
		severity VulnerabilitySeverity
		expected string
	}{
		{SeverityCritical, "Critical"},
		{SeverityHigh, "High"},
		{SeverityMedium, "Medium"},
		{SeverityLow, "Low"},
		{SeverityNegligible, "Negligible"},
		{SeverityUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.severity.String())
		})
	}
}

func TestVulnerabilitySeverity_Compare(t *testing.T) {
	tests := []struct {
		name     string
		s1       VulnerabilitySeverity
		s2       VulnerabilitySeverity
		expected int
	}{
		{
			name:     "critical > high",
			s1:       SeverityCritical,
			s2:       SeverityHigh,
			expected: 1,
		},
		{
			name:     "high > medium",
			s1:       SeverityHigh,
			s2:       SeverityMedium,
			expected: 1,
		},
		{
			name:     "medium > low",
			s1:       SeverityMedium,
			s2:       SeverityLow,
			expected: 1,
		},
		{
			name:     "low > negligible",
			s1:       SeverityLow,
			s2:       SeverityNegligible,
			expected: 1,
		},
		{
			name:     "equal severities",
			s1:       SeverityHigh,
			s2:       SeverityHigh,
			expected: 0,
		},
		{
			name:     "low < high",
			s1:       SeverityLow,
			s2:       SeverityHigh,
			expected: -1,
		},
		{
			name:     "unknown severity",
			s1:       SeverityUnknown,
			s2:       SeverityHigh,
			expected: -1, // Unknown has order -1, high has order 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.s1.Compare(tt.s2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVulnerabilityReport_Creation(t *testing.T) {
	now := time.Now()

	sbom := &SBOM{
		Format:      FormatSPDX23JSON,
		SpecVersion: "2.3",
		DocumentID:  "test-doc",
		Metadata: SBOMMetadata{
			Name:    "test-app",
			Version: "1.0.0",
		},
		CreatedAt: now,
	}

	vulnerabilities := []Vulnerability{
		{
			ID:       "CVE-2023-00001",
			Severity: SeverityCritical,
		},
		{
			ID:       "CVE-2023-00002",
			Severity: SeverityHigh,
		},
		{
			ID:       "CVE-2023-00003",
			Severity: SeverityMedium,
		},
	}

	report := &VulnerabilityReport{
		SBOM:            sbom,
		Vulnerabilities: vulnerabilities,
		Summary: VulnerabilitySummary{
			Total: 3,
			BySeverity: map[VulnerabilitySeverity]int{
				SeverityCritical: 1,
				SeverityHigh:     1,
				SeverityMedium:   1,
			},
			Fixable:            2,
			Unfixable:          1,
			AffectedComponents: 2,
		},
		ScanMetadata: ScanMetadata{
			Scanner:        "trivy",
			ScannerVersion: "0.45.0",
			Duration:       5 * time.Second,
			StartTime:      now.Add(-5 * time.Second),
			EndTime:        now,
		},
		Timestamp: now,
		Database: DatabaseInfo{
			Name:          "nvd",
			Version:       "2023.10",
			LastUpdated:   now.Add(-24 * time.Hour),
			RecordCount:   150000,
			SchemaVersion: "1.0",
		},
	}

	assert.NotNil(t, report.SBOM)
	assert.Len(t, report.Vulnerabilities, 3)
	assert.Equal(t, 3, report.Summary.Total)
	assert.Equal(t, 1, report.Summary.BySeverity[SeverityCritical])
	assert.Equal(t, 2, report.Summary.Fixable)
}

func TestGenerateOptions_Defaults(t *testing.T) {
	opts := GenerateOptions{
		Format:                 FormatSPDX23JSON,
		IncludePackages:        true,
		IncludeDependencies:    true,
		IncludeLicenses:        true,
		IncludeVulnerabilities: false,
		Platform:               "linux/amd64",
		Scope:                  ScopeAllLayers,
		Timeout:                5 * time.Minute,
		Author:                 "test-user",
		AuthorEmail:            "test@example.com",
	}

	assert.Equal(t, FormatSPDX23JSON, opts.Format)
	assert.True(t, opts.IncludePackages)
	assert.True(t, opts.IncludeDependencies)
	assert.True(t, opts.IncludeLicenses)
	assert.False(t, opts.IncludeVulnerabilities)
	assert.Equal(t, "linux/amd64", opts.Platform)
	assert.Equal(t, ScopeAllLayers, opts.Scope)
}

func TestScanOptions_Filtering(t *testing.T) {
	opts := ScanOptions{
		Severities: []VulnerabilitySeverity{
			SeverityCritical,
			SeverityHigh,
		},
		IgnoreCVEs:     []string{"CVE-2023-12345", "CVE-2023-67890"},
		OnlyFixed:      true,
		Database:       "nvd",
		FailOn:         SeverityHigh,
		Timeout:        10 * time.Minute,
		IncludeExpired: false,
	}

	assert.Len(t, opts.Severities, 2)
	assert.Len(t, opts.IgnoreCVEs, 2)
	assert.True(t, opts.OnlyFixed)
	assert.Equal(t, SeverityHigh, opts.FailOn)
}

func TestSBOMFormat_Values(t *testing.T) {
	formats := []SBOMFormat{
		FormatSPDX22JSON,
		FormatSPDX23JSON,
		FormatCycloneDX14JSON,
		FormatCycloneDX15JSON,
		FormatSyftJSON,
	}

	assert.Len(t, formats, 5)
	assert.Equal(t, "spdx-2.2-json", string(FormatSPDX22JSON))
	assert.Equal(t, "spdx-2.3-json", string(FormatSPDX23JSON))
	assert.Equal(t, "cyclonedx-1.4-json", string(FormatCycloneDX14JSON))
	assert.Equal(t, "cyclonedx-1.5-json", string(FormatCycloneDX15JSON))
	assert.Equal(t, "syft-json", string(FormatSyftJSON))
}

func TestComponentType_Values(t *testing.T) {
	types := []ComponentType{
		ComponentTypeApplication,
		ComponentTypeLibrary,
		ComponentTypeFramework,
		ComponentTypeContainer,
		ComponentTypeOS,
		ComponentTypeDevice,
		ComponentTypeFile,
	}

	assert.Len(t, types, 7)
	assert.Equal(t, "application", string(ComponentTypeApplication))
	assert.Equal(t, "library", string(ComponentTypeLibrary))
	assert.Equal(t, "container", string(ComponentTypeContainer))
}

func TestDependencyScope_Values(t *testing.T) {
	scopes := []DependencyScope{
		ScopeRuntime,
		ScopeDevelopment,
		ScopeTest,
		ScopeOptional,
		ScopeProvided,
	}

	assert.Len(t, scopes, 5)
	assert.Equal(t, "runtime", string(ScopeRuntime))
	assert.Equal(t, "development", string(ScopeDevelopment))
	assert.Equal(t, "test", string(ScopeTest))
}

func TestImageMetadata_Complete(t *testing.T) {
	metadata := &ImageMetadata{
		ID:           "sha256:abc123",
		RepoTags:     []string{"nginx:latest", "nginx:1.25"},
		RepoDigests:  []string{"nginx@sha256:def456"},
		Size:         142000000,
		Architecture: "amd64",
		OS:           "linux",
		Created:      time.Now().Add(-7 * 24 * time.Hour),
		Layers: []Layer{
			{
				Digest:  "sha256:layer1",
				Size:    50000000,
				Command: "ADD file:abc123 in /",
			},
			{
				Digest:  "sha256:layer2",
				Size:    92000000,
				Command: "RUN apt-get update && apt-get install -y nginx",
			},
		},
	}

	assert.Equal(t, "sha256:abc123", metadata.ID)
	assert.Len(t, metadata.RepoTags, 2)
	assert.Len(t, metadata.Layers, 2)
	assert.Equal(t, int64(142000000), metadata.Size)
	assert.Equal(t, "amd64", metadata.Architecture)
}

func TestLicense_Parsing(t *testing.T) {
	tests := []struct {
		name    string
		license License
		hasID   bool
		hasURL  bool
		hasText bool
	}{
		{
			name: "SPDX license with ID",
			license: License{
				ID:   "MIT",
				Name: "MIT License",
				URL:  "https://opensource.org/licenses/MIT",
			},
			hasID:  true,
			hasURL: true,
		},
		{
			name: "custom license with text",
			license: License{
				Name: "Custom License",
				Text: "Copyright (c) 2023...",
			},
			hasText: true,
		},
		{
			name: "license with URL only",
			license: License{
				Name: "Proprietary",
				URL:  "https://example.com/license",
			},
			hasURL: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.hasID {
				assert.NotEmpty(t, tt.license.ID)
			}
			if tt.hasURL {
				assert.NotEmpty(t, tt.license.URL)
			}
			if tt.hasText {
				assert.NotEmpty(t, tt.license.Text)
			}
			assert.NotEmpty(t, tt.license.Name)
		})
	}
}

func TestExternalReference_Types(t *testing.T) {
	references := []ExternalReference{
		{
			Type:    ReferenceTypeVCS,
			URL:     "https://github.com/org/repo",
			Comment: "Source repository",
		},
		{
			Type:    ReferenceTypeWebsite,
			URL:     "https://example.com",
			Comment: "Project website",
		},
		{
			Type:    ReferenceTypeDocumentation,
			URL:     "https://docs.example.com",
			Comment: "API documentation",
		},
		{
			Type: ReferenceTypeAdvisory,
			URL:  "https://nvd.nist.gov/vuln/detail/CVE-2023-12345",
		},
	}

	assert.Len(t, references, 4)
	assert.Equal(t, ReferenceTypeVCS, references[0].Type)
	assert.Equal(t, ReferenceTypeWebsite, references[1].Type)
	assert.Equal(t, ReferenceTypeDocumentation, references[2].Type)
	assert.Equal(t, ReferenceTypeAdvisory, references[3].Type)
}

func TestVulnerabilitySummary_Calculation(t *testing.T) {
	vulnerabilities := []Vulnerability{
		{ID: "CVE-1", Severity: SeverityCritical, FixAvailable: true},
		{ID: "CVE-2", Severity: SeverityHigh, FixAvailable: true},
		{ID: "CVE-3", Severity: SeverityHigh, FixAvailable: false},
		{ID: "CVE-4", Severity: SeverityMedium, FixAvailable: true},
		{ID: "CVE-5", Severity: SeverityLow, FixAvailable: false},
	}

	summary := VulnerabilitySummary{
		Total: len(vulnerabilities),
		BySeverity: map[VulnerabilitySeverity]int{
			SeverityCritical: 0,
			SeverityHigh:     0,
			SeverityMedium:   0,
			SeverityLow:      0,
		},
		Fixable:   0,
		Unfixable: 0,
	}

	for _, v := range vulnerabilities {
		summary.BySeverity[v.Severity]++
		if v.FixAvailable {
			summary.Fixable++
		} else {
			summary.Unfixable++
		}
	}

	assert.Equal(t, 5, summary.Total)
	assert.Equal(t, 1, summary.BySeverity[SeverityCritical])
	assert.Equal(t, 2, summary.BySeverity[SeverityHigh])
	assert.Equal(t, 1, summary.BySeverity[SeverityMedium])
	assert.Equal(t, 1, summary.BySeverity[SeverityLow])
	assert.Equal(t, 3, summary.Fixable)
	assert.Equal(t, 2, summary.Unfixable)
}

func TestSBOM_WithSignature(t *testing.T) {
	sbom := &SBOM{
		Format:      FormatSPDX23JSON,
		SpecVersion: "2.3",
		DocumentID:  "signed-doc",
		Signature: &Signature{
			Algorithm: "RSA-SHA256",
			Value:     []byte("signature-data"),
			PublicKey: []byte("public-key-data"),
			SignedAt:  time.Now(),
		},
		CreatedAt: time.Now(),
	}

	require.NotNil(t, sbom.Signature)
	assert.Equal(t, "RSA-SHA256", sbom.Signature.Algorithm)
	assert.NotEmpty(t, sbom.Signature.Value)
	assert.NotEmpty(t, sbom.Signature.PublicKey)
}

func TestMockSBOMGenerator(t *testing.T) {
	// Mock implementation for testing interfaces
	generator := &mockSBOMGenerator{}

	ctx := context.Background()
	opts := GenerateOptions{
		Format:              FormatSPDX23JSON,
		IncludePackages:     true,
		IncludeDependencies: true,
	}

	sbom, err := generator.Generate(ctx, opts)
	require.NoError(t, err)
	assert.NotNil(t, sbom)
	assert.Equal(t, FormatSPDX23JSON, sbom.Format)
}

// Mock implementations for interface testing
type mockSBOMGenerator struct{}

func (m *mockSBOMGenerator) Generate(ctx context.Context, opts GenerateOptions) (*SBOM, error) {
	return &SBOM{
		Format:      opts.Format,
		SpecVersion: "2.3",
		DocumentID:  "mock-doc",
		Metadata: SBOMMetadata{
			Name:      "mock-component",
			Version:   "1.0.0",
			Timestamp: time.Now(),
		},
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockSBOMGenerator) GenerateFromImage(ctx context.Context, imageRef string, opts GenerateOptions) (*SBOM, error) {
	sbom, _ := m.Generate(ctx, opts)
	sbom.Metadata.SourceInfo.Type = SourceTypeImage
	sbom.Metadata.SourceInfo.Target = imageRef
	return sbom, nil
}

func (m *mockSBOMGenerator) GenerateFromDirectory(ctx context.Context, path string, opts GenerateOptions) (*SBOM, error) {
	sbom, _ := m.Generate(ctx, opts)
	sbom.Metadata.SourceInfo.Type = SourceTypeDirectory
	sbom.Metadata.SourceInfo.Target = path
	return sbom, nil
}

func (m *mockSBOMGenerator) GenerateFromArchive(ctx context.Context, archivePath string, opts GenerateOptions) (*SBOM, error) {
	sbom, _ := m.Generate(ctx, opts)
	sbom.Metadata.SourceInfo.Type = SourceTypeArchive
	sbom.Metadata.SourceInfo.Target = archivePath
	return sbom, nil
}

func (m *mockSBOMGenerator) SupportedFormats() []SBOMFormat {
	return []SBOMFormat{FormatSPDX23JSON, FormatCycloneDX15JSON}
}

func (m *mockSBOMGenerator) SupportedSources() []SourceType {
	return []SourceType{SourceTypeImage, SourceTypeDirectory, SourceTypeArchive}
}
