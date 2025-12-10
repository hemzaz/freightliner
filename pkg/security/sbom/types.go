package sbom

import (
	"time"
)

// SBOM represents a Software Bill of Materials following SPDX and CycloneDX standards
type SBOM struct {
	// Format specifies the SBOM format (SPDX, CycloneDX)
	Format SBOMFormat `json:"format"`

	// Version of the SBOM specification
	SpecVersion string `json:"specVersion"`

	// DocumentID uniquely identifies this SBOM document
	DocumentID string `json:"documentId"`

	// Metadata about the SBOM document
	Metadata SBOMMetadata `json:"metadata"`

	// Components are the software components in the SBOM
	Components []Component `json:"components"`

	// Dependencies represents the dependency graph
	Dependencies []Dependency `json:"dependencies,omitempty"`

	// ExternalReferences are references to external resources
	ExternalReferences []ExternalReference `json:"externalReferences,omitempty"`

	// Licenses are the licenses found in the components
	Licenses []License `json:"licenses,omitempty"`

	// Vulnerabilities are known vulnerabilities (if included)
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`

	// Signature contains the SBOM signature (if signed)
	Signature *Signature `json:"signature,omitempty"`

	// CreatedAt timestamp
	CreatedAt time.Time `json:"createdAt"`
}

// SBOMMetadata contains metadata about the SBOM document
type SBOMMetadata struct {
	// Name of the component being described
	Name string `json:"name"`

	// Version of the component
	Version string `json:"version"`

	// Description of the component
	Description string `json:"description,omitempty"`

	// Authors of the SBOM
	Authors []Author `json:"authors,omitempty"`

	// Supplier of the component
	Supplier string `json:"supplier,omitempty"`

	// Manufacturer of the component
	Manufacturer string `json:"manufacturer,omitempty"`

	// Timestamp when the SBOM was created
	Timestamp time.Time `json:"timestamp"`

	// ToolInfo describes the tool that generated the SBOM
	ToolInfo ToolInfo `json:"toolInfo"`

	// SourceInfo describes the source of the analysis
	SourceInfo SourceInfo `json:"sourceInfo"`

	// Properties contains additional custom properties
	Properties map[string]string `json:"properties,omitempty"`
}

// Component represents a software component in the SBOM
type Component struct {
	// ID uniquely identifies the component
	ID string `json:"id"`

	// Type of component (library, application, container, etc.)
	Type ComponentType `json:"type"`

	// Name of the component
	Name string `json:"name"`

	// Version of the component
	Version string `json:"version"`

	// PackageURL (PURL) identifier
	PURL string `json:"purl,omitempty"`

	// CPE (Common Platform Enumeration) identifier
	CPE string `json:"cpe,omitempty"`

	// Description of the component
	Description string `json:"description,omitempty"`

	// Licenses associated with the component
	Licenses []License `json:"licenses,omitempty"`

	// Supplier of the component
	Supplier string `json:"supplier,omitempty"`

	// Publisher of the component
	Publisher string `json:"publisher,omitempty"`

	// Hashes of the component (SHA256, SHA1, etc.)
	Hashes map[string]string `json:"hashes,omitempty"`

	// ExternalReferences to external resources
	ExternalReferences []ExternalReference `json:"externalReferences,omitempty"`

	// Properties contains additional metadata
	Properties map[string]string `json:"properties,omitempty"`

	// FoundBy indicates which cataloger found this component
	FoundBy string `json:"foundBy,omitempty"`

	// Locations where the component was found
	Locations []Location `json:"locations,omitempty"`

	// Metadata contains component-specific metadata
	Metadata interface{} `json:"metadata,omitempty"`
}

// Dependency represents a dependency relationship between components
type Dependency struct {
	// Ref identifies the dependent component
	Ref string `json:"ref"`

	// DependsOn lists the component IDs this component depends on
	DependsOn []string `json:"dependsOn"`

	// Scope of the dependency (runtime, development, etc.)
	Scope DependencyScope `json:"scope,omitempty"`

	// Direct indicates if this is a direct dependency
	Direct bool `json:"direct"`

	// VersionConstraint specifies version requirements
	VersionConstraint string `json:"versionConstraint,omitempty"`
}

// VulnerabilityReport contains the results of a vulnerability scan
type VulnerabilityReport struct {
	// SBOM that was scanned
	SBOM *SBOM `json:"sbom"`

	// Vulnerabilities found
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`

	// Summary statistics
	Summary VulnerabilitySummary `json:"summary"`

	// ScanMetadata contains information about the scan
	ScanMetadata ScanMetadata `json:"scanMetadata"`

	// Timestamp when the scan was performed
	Timestamp time.Time `json:"timestamp"`

	// Database information
	Database DatabaseInfo `json:"database"`
}

// Vulnerability represents a known security vulnerability
type Vulnerability struct {
	// ID is the vulnerability identifier (CVE, GHSA, etc.)
	ID string `json:"id"`

	// Aliases are alternative identifiers
	Aliases []string `json:"aliases,omitempty"`

	// Source identifies where the vulnerability was found
	Source string `json:"source"`

	// Severity of the vulnerability
	Severity VulnerabilitySeverity `json:"severity"`

	// CVSSScore if available
	CVSSScore float64 `json:"cvssScore,omitempty"`

	// CVSSVector string representation
	CVSSVector string `json:"cvssVector,omitempty"`

	// Description of the vulnerability
	Description string `json:"description"`

	// AffectedComponents lists affected components
	AffectedComponents []string `json:"affectedComponents"`

	// AffectedVersions specifies which versions are affected
	AffectedVersions []string `json:"affectedVersions,omitempty"`

	// FixedVersions specifies versions with the fix
	FixedVersions []string `json:"fixedVersions,omitempty"`

	// URLs to more information
	URLs []string `json:"urls,omitempty"`

	// PublishedDate when the vulnerability was published
	PublishedDate time.Time `json:"publishedDate,omitempty"`

	// ModifiedDate when the vulnerability was last modified
	ModifiedDate time.Time `json:"modifiedDate,omitempty"`

	// ExploitAvailable indicates if exploits are known
	ExploitAvailable bool `json:"exploitAvailable"`

	// FixAvailable indicates if a fix is available
	FixAvailable bool `json:"fixAvailable"`

	// Remediation guidance
	Remediation string `json:"remediation,omitempty"`

	// References to external resources
	References []Reference `json:"references,omitempty"`
}

// VulnerabilitySummary provides summary statistics for vulnerabilities
type VulnerabilitySummary struct {
	// Total number of vulnerabilities
	Total int `json:"total"`

	// BySeverity breaks down counts by severity
	BySeverity map[VulnerabilitySeverity]int `json:"bySeverity"`

	// ByType breaks down counts by vulnerability type
	ByType map[string]int `json:"byType"`

	// Fixable number of vulnerabilities with available fixes
	Fixable int `json:"fixable"`

	// Unfixable number of vulnerabilities without fixes
	Unfixable int `json:"unfixable"`

	// AffectedComponents number of components with vulnerabilities
	AffectedComponents int `json:"affectedComponents"`
}

// ScanMetadata contains metadata about the vulnerability scan
type ScanMetadata struct {
	// Scanner name and version
	Scanner string `json:"scanner"`

	// ScannerVersion version of the scanner
	ScannerVersion string `json:"scannerVersion"`

	// Duration of the scan
	Duration time.Duration `json:"duration"`

	// StartTime when the scan started
	StartTime time.Time `json:"startTime"`

	// EndTime when the scan completed
	EndTime time.Time `json:"endTime"`

	// Options used for scanning
	Options interface{} `json:"options,omitempty"`
}

// Author represents an SBOM author
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// ToolInfo describes the tool that generated the SBOM
type ToolInfo struct {
	// Vendor of the tool
	Vendor string `json:"vendor"`

	// Name of the tool
	Name string `json:"name"`

	// Version of the tool
	Version string `json:"version"`

	// HashAlgorithm used by the tool
	HashAlgorithm string `json:"hashAlgorithm,omitempty"`
}

// SourceInfo describes the source that was analyzed
type SourceInfo struct {
	// Type of source (image, directory, archive, etc.)
	Type SourceType `json:"type"`

	// Target is the identifier for the source
	Target string `json:"target"`

	// Scheme specifies how to interpret the target
	Scheme string `json:"scheme,omitempty"`

	// Platform for multi-platform images
	Platform string `json:"platform,omitempty"`

	// ImageMetadata for container images
	ImageMetadata *ImageMetadata `json:"imageMetadata,omitempty"`
}

// ImageMetadata contains metadata about a container image
type ImageMetadata struct {
	// ID of the image
	ID string `json:"id"`

	// RepoTags are the repository tags
	RepoTags []string `json:"repoTags,omitempty"`

	// RepoDigests are the repository digests
	RepoDigests []string `json:"repoDigests,omitempty"`

	// Size of the image in bytes
	Size int64 `json:"size"`

	// Architecture of the image
	Architecture string `json:"architecture"`

	// OS of the image
	OS string `json:"os"`

	// Created timestamp
	Created time.Time `json:"created,omitempty"`

	// Layers in the image
	Layers []Layer `json:"layers,omitempty"`
}

// Layer represents a layer in a container image
type Layer struct {
	// Digest of the layer
	Digest string `json:"digest"`

	// Size of the layer in bytes
	Size int64 `json:"size"`

	// Command that created the layer
	Command string `json:"command,omitempty"`
}

// License represents a software license
type License struct {
	// ID is the SPDX license identifier
	ID string `json:"id,omitempty"`

	// Name of the license
	Name string `json:"name"`

	// URL to the license text
	URL string `json:"url,omitempty"`

	// Text of the license
	Text string `json:"text,omitempty"`
}

// ExternalReference represents a reference to an external resource
type ExternalReference struct {
	// Type of reference
	Type ReferenceType `json:"type"`

	// URL to the resource
	URL string `json:"url"`

	// Comment about the reference
	Comment string `json:"comment,omitempty"`

	// Hashes of the external resource
	Hashes map[string]string `json:"hashes,omitempty"`
}

// Reference represents a vulnerability reference
type Reference struct {
	// URL to the reference
	URL string `json:"url"`

	// Source of the reference
	Source string `json:"source,omitempty"`

	// Type of reference
	Type string `json:"type,omitempty"`
}

// Location represents where a component was found
type Location struct {
	// Path to the file or directory
	Path string `json:"path"`

	// LayerID for container images
	LayerID string `json:"layerId,omitempty"`

	// AccessPath shows how to access this location
	AccessPath string `json:"accessPath,omitempty"`
}

// Signature contains SBOM signature information
type Signature struct {
	// Algorithm used for signing
	Algorithm string `json:"algorithm"`

	// Value is the signature value
	Value []byte `json:"value"`

	// PublicKey used for verification
	PublicKey []byte `json:"publicKey,omitempty"`

	// Certificate chain
	Certificates [][]byte `json:"certificates,omitempty"`

	// SignedAt timestamp
	SignedAt time.Time `json:"signedAt"`
}

// SBOMFormat represents the SBOM format
type SBOMFormat string

const (
	// FormatSPDX22JSON is SPDX 2.2 in JSON format
	FormatSPDX22JSON SBOMFormat = "spdx-2.2-json"

	// FormatSPDX23JSON is SPDX 2.3 in JSON format
	FormatSPDX23JSON SBOMFormat = "spdx-2.3-json"

	// FormatCycloneDX14JSON is CycloneDX 1.4 in JSON format
	FormatCycloneDX14JSON SBOMFormat = "cyclonedx-1.4-json"

	// FormatCycloneDX15JSON is CycloneDX 1.5 in JSON format
	FormatCycloneDX15JSON SBOMFormat = "cyclonedx-1.5-json"

	// FormatSyftJSON is Syft's native JSON format
	FormatSyftJSON SBOMFormat = "syft-json"
)

// SourceType represents the type of source being analyzed
type SourceType string

const (
	// SourceTypeImage is a container image
	SourceTypeImage SourceType = "image"

	// SourceTypeDirectory is a filesystem directory
	SourceTypeDirectory SourceType = "directory"

	// SourceTypeArchive is an archive file
	SourceTypeArchive SourceType = "archive"

	// SourceTypeFile is a single file
	SourceTypeFile SourceType = "file"
)

// ComponentType represents the type of component
type ComponentType string

const (
	// ComponentTypeApplication is an application
	ComponentTypeApplication ComponentType = "application"

	// ComponentTypeLibrary is a library
	ComponentTypeLibrary ComponentType = "library"

	// ComponentTypeFramework is a framework
	ComponentTypeFramework ComponentType = "framework"

	// ComponentTypeContainer is a container
	ComponentTypeContainer ComponentType = "container"

	// ComponentTypeOS is an operating system
	ComponentTypeOS ComponentType = "operating-system"

	// ComponentTypeDevice is a device
	ComponentTypeDevice ComponentType = "device"

	// ComponentTypeFile is a file
	ComponentTypeFile ComponentType = "file"
)

// DependencyScope represents the scope of a dependency
type DependencyScope string

const (
	// ScopeRuntime is a runtime dependency
	ScopeRuntime DependencyScope = "runtime"

	// ScopeDevelopment is a development dependency
	ScopeDevelopment DependencyScope = "development"

	// ScopeTest is a test dependency
	ScopeTest DependencyScope = "test"

	// ScopeOptional is an optional dependency
	ScopeOptional DependencyScope = "optional"

	// ScopeProvided is a provided dependency
	ScopeProvided DependencyScope = "provided"
)

// VulnerabilitySeverity represents the severity of a vulnerability
type VulnerabilitySeverity string

const (
	// SeverityCritical is critical severity
	SeverityCritical VulnerabilitySeverity = "Critical"

	// SeverityHigh is high severity
	SeverityHigh VulnerabilitySeverity = "High"

	// SeverityMedium is medium severity
	SeverityMedium VulnerabilitySeverity = "Medium"

	// SeverityLow is low severity
	SeverityLow VulnerabilitySeverity = "Low"

	// SeverityNegligible is negligible severity
	SeverityNegligible VulnerabilitySeverity = "Negligible"

	// SeverityUnknown is unknown severity
	SeverityUnknown VulnerabilitySeverity = "Unknown"
)

// ReferenceType represents the type of external reference
type ReferenceType string

const (
	// ReferenceTypeVCS is a version control system reference
	ReferenceTypeVCS ReferenceType = "vcs"

	// ReferenceTypeWebsite is a website reference
	ReferenceTypeWebsite ReferenceType = "website"

	// ReferenceTypeIssueTracker is an issue tracker reference
	ReferenceTypeIssueTracker ReferenceType = "issue-tracker"

	// ReferenceTypeMailingList is a mailing list reference
	ReferenceTypeMailingList ReferenceType = "mailing-list"

	// ReferenceTypeDocumentation is a documentation reference
	ReferenceTypeDocumentation ReferenceType = "documentation"

	// ReferenceTypeDistribution is a distribution reference
	ReferenceTypeDistribution ReferenceType = "distribution"

	// ReferenceTypeLicense is a license reference
	ReferenceTypeLicense ReferenceType = "license"

	// ReferenceTypeAdvisory is a security advisory reference
	ReferenceTypeAdvisory ReferenceType = "advisory"
)

// String returns the string representation of VulnerabilitySeverity
func (s VulnerabilitySeverity) String() string {
	return string(s)
}

// Compare compares two vulnerability severities and returns:
// -1 if s is less severe than other
// 0 if s equals other
// 1 if s is more severe than other
func (s VulnerabilitySeverity) Compare(other VulnerabilitySeverity) int {
	severityOrder := map[VulnerabilitySeverity]int{
		SeverityNegligible: 0,
		SeverityLow:        1,
		SeverityMedium:     2,
		SeverityHigh:       3,
		SeverityCritical:   4,
		SeverityUnknown:    -1,
	}

	sOrder, sExists := severityOrder[s]
	otherOrder, otherExists := severityOrder[other]

	if !sExists || !otherExists {
		return 0
	}

	if sOrder < otherOrder {
		return -1
	} else if sOrder > otherOrder {
		return 1
	}
	return 0
}
