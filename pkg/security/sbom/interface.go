// Package sbom provides interfaces and types for Software Bill of Materials (SBOM) generation
// and vulnerability scanning integration.
package sbom

import (
	"context"
	"io"
	"time"
)

// SBOMGenerator defines the interface for generating Software Bill of Materials
type SBOMGenerator interface {
	// Generate creates an SBOM for the given source
	Generate(ctx context.Context, opts GenerateOptions) (*SBOM, error)

	// GenerateFromImage creates an SBOM from a container image
	GenerateFromImage(ctx context.Context, imageRef string, opts GenerateOptions) (*SBOM, error)

	// GenerateFromDirectory creates an SBOM from a filesystem directory
	GenerateFromDirectory(ctx context.Context, path string, opts GenerateOptions) (*SBOM, error)

	// GenerateFromArchive creates an SBOM from an archive file (tar, zip, etc.)
	GenerateFromArchive(ctx context.Context, archivePath string, opts GenerateOptions) (*SBOM, error)

	// SupportedFormats returns the list of supported SBOM output formats
	SupportedFormats() []SBOMFormat

	// SupportedSources returns the list of supported source types
	SupportedSources() []SourceType
}

// SBOMExporter defines the interface for exporting SBOMs to various formats
type SBOMExporter interface {
	// Export writes the SBOM to the provided writer in the specified format
	Export(ctx context.Context, sbom *SBOM, format SBOMFormat, writer io.Writer) error

	// ExportToFile writes the SBOM to a file in the specified format
	ExportToFile(ctx context.Context, sbom *SBOM, format SBOMFormat, filePath string) error

	// Validate validates the SBOM against the format specification
	Validate(ctx context.Context, sbom *SBOM, format SBOMFormat) error

	// Convert converts an SBOM from one format to another
	Convert(ctx context.Context, sbom *SBOM, fromFormat, toFormat SBOMFormat) (*SBOM, error)

	// SupportedFormats returns the list of supported export formats
	SupportedFormats() []SBOMFormat
}

// VulnerabilityScanner defines the interface for scanning SBOMs for vulnerabilities
type VulnerabilityScanner interface {
	// Scan scans the SBOM for known vulnerabilities
	Scan(ctx context.Context, sbom *SBOM, opts ScanOptions) (*VulnerabilityReport, error)

	// ScanComponent scans a specific component for vulnerabilities
	ScanComponent(ctx context.Context, component *Component, opts ScanOptions) ([]Vulnerability, error)

	// UpdateDatabase updates the vulnerability database
	UpdateDatabase(ctx context.Context) error

	// GetDatabaseInfo returns information about the vulnerability database
	GetDatabaseInfo(ctx context.Context) (*DatabaseInfo, error)

	// SupportedSeverities returns the list of supported severity levels
	SupportedSeverities() []VulnerabilitySeverity

	// SupportedVulnDBs returns the list of supported vulnerability databases
	SupportedVulnDBs() []string
}

// SBOMComparer defines the interface for comparing SBOMs
type SBOMComparer interface {
	// Compare compares two SBOMs and returns the differences
	Compare(ctx context.Context, sbom1, sbom2 *SBOM) (*SBOMDiff, error)

	// DetectChanges detects changes between two SBOMs (added, removed, updated components)
	DetectChanges(ctx context.Context, baseline, current *SBOM) (*ChangeReport, error)

	// FindNewVulnerabilities compares vulnerability reports and identifies new vulnerabilities
	FindNewVulnerabilities(ctx context.Context, baseline, current *VulnerabilityReport) ([]Vulnerability, error)
}

// SBOMEnricher defines the interface for enriching SBOMs with additional metadata
type SBOMEnricher interface {
	// Enrich adds additional metadata to the SBOM
	Enrich(ctx context.Context, sbom *SBOM, opts EnrichOptions) error

	// EnrichComponent adds additional metadata to a component
	EnrichComponent(ctx context.Context, component *Component, opts EnrichOptions) error

	// AddLicenseInfo adds license information to components
	AddLicenseInfo(ctx context.Context, sbom *SBOM) error

	// AddSupplyChainInfo adds supply chain metadata
	AddSupplyChainInfo(ctx context.Context, sbom *SBOM) error
}

// SBOMAttestor defines the interface for creating and verifying SBOM attestations
type SBOMAttestor interface {
	// CreateAttestation creates a signed attestation for the SBOM
	CreateAttestation(ctx context.Context, sbom *SBOM, opts AttestationOptions) (*Attestation, error)

	// VerifyAttestation verifies a signed SBOM attestation
	VerifyAttestation(ctx context.Context, attestation *Attestation, opts AttestationOptions) (bool, error)

	// SignSBOM signs an SBOM using the specified signing method
	SignSBOM(ctx context.Context, sbom *SBOM, opts SigningOptions) ([]byte, error)

	// VerifySignature verifies an SBOM signature
	VerifySignature(ctx context.Context, sbom *SBOM, signature []byte, opts SigningOptions) (bool, error)
}

// GenerateOptions contains options for SBOM generation
type GenerateOptions struct {
	// Format specifies the output SBOM format
	Format SBOMFormat

	// IncludePackages specifies whether to include package information
	IncludePackages bool

	// IncludeDependencies specifies whether to include dependency graph
	IncludeDependencies bool

	// IncludeLicenses specifies whether to include license information
	IncludeLicenses bool

	// IncludeVulnerabilities specifies whether to include vulnerability scan results
	IncludeVulnerabilities bool

	// Platform specifies the target platform (for multi-arch images)
	Platform string

	// Catalogers specifies which catalogers to use (empty = all)
	Catalogers []string

	// Scope defines the scope of analysis
	Scope AnalysisScope

	// Timeout specifies the maximum time for generation
	Timeout time.Duration

	// Author information
	Author      string
	AuthorEmail string
}

// ScanOptions contains options for vulnerability scanning
type ScanOptions struct {
	// Severities filters vulnerabilities by severity
	Severities []VulnerabilitySeverity

	// IgnoreCVEs is a list of CVE IDs to ignore
	IgnoreCVEs []string

	// OnlyFixed only reports vulnerabilities with available fixes
	OnlyFixed bool

	// Database specifies which vulnerability database to use
	Database string

	// FailOn specifies the severity level at which to fail
	FailOn VulnerabilitySeverity

	// Timeout specifies the maximum time for scanning
	Timeout time.Duration

	// IncludeExpired includes expired/rejected vulnerabilities
	IncludeExpired bool
}

// EnrichOptions contains options for SBOM enrichment
type EnrichOptions struct {
	// AddLicenses adds license information
	AddLicenses bool

	// AddSupplyChain adds supply chain metadata
	AddSupplyChain bool

	// AddSecurityInfo adds security-related metadata
	AddSecurityInfo bool

	// AddBuildInfo adds build-time information
	AddBuildInfo bool

	// CustomFields allows adding custom metadata fields
	CustomFields map[string]interface{}
}

// AttestationOptions contains options for SBOM attestation
type AttestationOptions struct {
	// SigningKey is the path to the signing key
	SigningKey string

	// SigningMethod specifies the signing method (cosign, gpg, etc.)
	SigningMethod SigningMethod

	// PredicateType specifies the attestation predicate type
	PredicateType string

	// RegistryURL is the container registry URL for attestation storage
	RegistryURL string

	// RegistryAuth contains authentication for the registry
	RegistryAuth *RegistryAuth
}

// SigningOptions contains options for SBOM signing
type SigningOptions struct {
	// KeyPath is the path to the signing key
	KeyPath string

	// Method specifies the signing method
	Method SigningMethod

	// Password for encrypted keys
	Password string

	// PublicKeyPath for verification
	PublicKeyPath string
}

// RegistryAuth contains authentication information for container registries
type RegistryAuth struct {
	Username string
	Password string
	Token    string
}

// DatabaseInfo contains information about the vulnerability database
type DatabaseInfo struct {
	// Name of the vulnerability database
	Name string

	// Version of the database
	Version string

	// LastUpdated timestamp
	LastUpdated time.Time

	// Source URL
	Source string

	// RecordCount number of vulnerability records
	RecordCount int

	// SchemaVersion database schema version
	SchemaVersion string
}

// SBOMDiff represents the differences between two SBOMs
type SBOMDiff struct {
	// AddedComponents are components present in sbom2 but not sbom1
	AddedComponents []Component

	// RemovedComponents are components present in sbom1 but not sbom2
	RemovedComponents []Component

	// ModifiedComponents are components present in both but with different versions/metadata
	ModifiedComponents []ComponentChange

	// Summary provides a high-level summary of changes
	Summary DiffSummary
}

// ComponentChange represents a change to a component
type ComponentChange struct {
	// Before is the component state in the baseline SBOM
	Before Component

	// After is the component state in the current SBOM
	After Component

	// ChangeType describes the type of change
	ChangeType ChangeType

	// ChangeDetails provides details about what changed
	ChangeDetails map[string]interface{}
}

// ChangeReport contains a detailed report of changes between SBOMs
type ChangeReport struct {
	// BaselineSBOM is the reference SBOM
	BaselineSBOM *SBOM

	// CurrentSBOM is the SBOM being compared
	CurrentSBOM *SBOM

	// Diff contains the differences
	Diff *SBOMDiff

	// VulnerabilityChanges tracks changes in vulnerabilities
	VulnerabilityChanges *VulnerabilityChanges

	// RiskScore assesses the risk of the changes
	RiskScore float64

	// Timestamp when the comparison was performed
	Timestamp time.Time
}

// VulnerabilityChanges tracks changes in vulnerability status
type VulnerabilityChanges struct {
	// NewVulnerabilities are newly discovered vulnerabilities
	NewVulnerabilities []Vulnerability

	// FixedVulnerabilities are vulnerabilities that were fixed
	FixedVulnerabilities []Vulnerability

	// SeverityChanges tracks vulnerabilities with changed severity
	SeverityChanges []VulnerabilitySeverityChange
}

// VulnerabilitySeverityChange represents a change in vulnerability severity
type VulnerabilitySeverityChange struct {
	VulnerabilityID string
	OldSeverity     VulnerabilitySeverity
	NewSeverity     VulnerabilitySeverity
	Reason          string
}

// DiffSummary provides a high-level summary of SBOM differences
type DiffSummary struct {
	TotalAdded    int
	TotalRemoved  int
	TotalModified int
	RiskLevel     RiskLevel
}

// Attestation represents a signed SBOM attestation
type Attestation struct {
	// SBOM is the attested SBOM
	SBOM *SBOM

	// Signature is the cryptographic signature
	Signature []byte

	// PredicateType specifies the type of attestation
	PredicateType string

	// Predicate contains the attestation predicate
	Predicate interface{}

	// Subject identifies what is being attested
	Subject AttestationSubject

	// Metadata contains attestation metadata
	Metadata AttestationMetadata
}

// AttestationSubject identifies what is being attested
type AttestationSubject struct {
	Name   string
	Digest map[string]string
}

// AttestationMetadata contains metadata about the attestation
type AttestationMetadata struct {
	BuildInvocationID string
	Timestamp         time.Time
	BuilderID         string
	Materials         []Material
}

// Material represents a build material
type Material struct {
	URI    string
	Digest map[string]string
}

// AnalysisScope defines the scope of SBOM analysis
type AnalysisScope string

const (
	// ScopeAllLayers analyzes all layers in a container image
	ScopeAllLayers AnalysisScope = "all-layers"

	// ScopeSquashed analyzes the squashed filesystem
	ScopeSquashed AnalysisScope = "squashed"

	// ScopeDirectory analyzes a directory tree
	ScopeDirectory AnalysisScope = "directory"
)

// SigningMethod specifies the method used for signing
type SigningMethod string

const (
	// SigningMethodCosign uses Sigstore Cosign
	SigningMethodCosign SigningMethod = "cosign"

	// SigningMethodGPG uses GPG signing
	SigningMethodGPG SigningMethod = "gpg"

	// SigningMethodX509 uses X.509 certificates
	SigningMethodX509 SigningMethod = "x509"
)

// ChangeType describes the type of change to a component
type ChangeType string

const (
	// ChangeTypeVersionUpdate indicates a version change
	ChangeTypeVersionUpdate ChangeType = "version-update"

	// ChangeTypeMetadataUpdate indicates metadata change
	ChangeTypeMetadataUpdate ChangeType = "metadata-update"

	// ChangeTypeLicenseUpdate indicates license change
	ChangeTypeLicenseUpdate ChangeType = "license-update"
)

// RiskLevel represents the risk level of changes
type RiskLevel string

const (
	// RiskLevelCritical indicates critical risk
	RiskLevelCritical RiskLevel = "critical"

	// RiskLevelHigh indicates high risk
	RiskLevelHigh RiskLevel = "high"

	// RiskLevelMedium indicates medium risk
	RiskLevelMedium RiskLevel = "medium"

	// RiskLevelLow indicates low risk
	RiskLevelLow RiskLevel = "low"

	// RiskLevelNone indicates no risk
	RiskLevelNone RiskLevel = "none"
)
