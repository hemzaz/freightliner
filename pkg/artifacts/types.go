// Package artifacts provides support for OCI artifact types beyond container images
package artifacts

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// ArtifactType represents the type of OCI artifact
type ArtifactType string

const (
	// ArtifactTypeHelm represents Helm chart artifacts
	ArtifactTypeHelm ArtifactType = "application/vnd.cncf.helm.chart.content.v1.tar+gzip"

	// ArtifactTypeHelmConfig represents Helm chart config
	ArtifactTypeHelmConfig ArtifactType = "application/vnd.cncf.helm.config.v1+json"

	// ArtifactTypeHelmProvenance represents Helm chart provenance
	ArtifactTypeHelmProvenance ArtifactType = "application/vnd.cncf.helm.chart.provenance.v1.prov"

	// ArtifactTypeWASM represents WebAssembly module artifacts
	ArtifactTypeWASM ArtifactType = "application/vnd.wasm.content.layer.v1+wasm"

	// ArtifactTypeWASMConfig represents WASM module config
	ArtifactTypeWASMConfig ArtifactType = "application/vnd.wasm.config.v1+json"

	// ArtifactTypeOCIImage represents standard OCI container image
	ArtifactTypeOCIImage ArtifactType = "application/vnd.oci.image.manifest.v1+json"

	// ArtifactTypeOCIImageIndex represents OCI image index (multi-arch)
	ArtifactTypeOCIImageIndex ArtifactType = "application/vnd.oci.image.index.v1+json"

	// ArtifactTypeDockerManifest represents Docker manifest v2 schema 2
	ArtifactTypeDockerManifest ArtifactType = "application/vnd.docker.distribution.manifest.v2+json"

	// ArtifactTypeDockerManifestList represents Docker manifest list
	ArtifactTypeDockerManifestList ArtifactType = "application/vnd.docker.distribution.manifest.list.v2+json"

	// ArtifactTypeMLModel represents machine learning model artifacts
	ArtifactTypeMLModel ArtifactType = "application/vnd.ml.model+tar+gzip"

	// ArtifactTypeMLModelConfig represents ML model configuration
	ArtifactTypeMLModelConfig ArtifactType = "application/vnd.ml.model.config.v1+json"

	// ArtifactTypeSBOM represents Software Bill of Materials (SBOM)
	ArtifactTypeSBOM ArtifactType = "application/vnd.dev.sbom.v1+json"

	// ArtifactTypeSPDX represents SPDX SBOM format
	ArtifactTypeSPDX ArtifactType = "application/spdx+json"

	// ArtifactTypeCycloneDX represents CycloneDX SBOM format
	ArtifactTypeCycloneDX ArtifactType = "application/vnd.cyclonedx+json"

	// ArtifactTypeSignature represents Cosign signatures
	ArtifactTypeSignature ArtifactType = "application/vnd.dev.cosign.simplesigning.v1+json"

	// ArtifactTypeAttestation represents in-toto attestations
	ArtifactTypeAttestation ArtifactType = "application/vnd.in-toto+json"

	// ArtifactTypeGeneric represents generic OCI artifacts
	ArtifactTypeGeneric ArtifactType = "application/vnd.oci.artifact.manifest.v1+json"
)

// Artifact represents an OCI artifact with metadata
type Artifact struct {
	// Reference is the full artifact reference (registry/repo:tag or @digest)
	Reference string

	// Repository is the artifact repository path
	Repository string

	// Tag is the artifact tag (if tagged)
	Tag string

	// Digest is the artifact digest
	Digest string

	// MediaType is the artifact media type
	MediaType string

	// ArtifactType is the classified artifact type
	ArtifactType ArtifactType

	// Size is the artifact size in bytes
	Size int64

	// Manifest is the OCI manifest descriptor
	Manifest *v1.Manifest

	// Index is the OCI index (for multi-arch artifacts)
	Index *v1.Index

	// Config is the artifact configuration blob
	Config []byte

	// Layers are the artifact layers
	Layers []Layer

	// Referrers are artifacts that reference this artifact
	Referrers []Referrer

	// Annotations are custom key-value pairs
	Annotations map[string]string

	// CreatedAt is when the artifact was created
	CreatedAt time.Time

	// UpdatedAt is when the artifact was last updated
	UpdatedAt time.Time
}

// Layer represents a layer in an OCI artifact
type Layer struct {
	// MediaType is the layer media type
	MediaType string

	// Digest is the layer digest
	Digest string

	// Size is the layer size in bytes
	Size int64

	// URLs are distribution URLs for the layer
	URLs []string

	// Annotations are layer-specific annotations
	Annotations map[string]string
}

// Referrer represents an artifact that references another artifact
type Referrer struct {
	// Digest is the referrer artifact digest
	Digest string

	// MediaType is the referrer media type
	MediaType string

	// ArtifactType is the referrer artifact type
	ArtifactType string

	// Annotations are referrer-specific annotations
	Annotations map[string]string

	// Size is the referrer size in bytes
	Size int64
}

// ReplicationOptions contains options for artifact replication
type ReplicationOptions struct {
	// IncludeLayers determines if layers should be replicated
	IncludeLayers bool

	// IncludeReferrers determines if referrers should be replicated
	IncludeReferrers bool

	// IncludeSignatures determines if signatures should be replicated
	IncludeSignatures bool

	// IncludeSBOMs determines if SBOMs should be replicated
	IncludeSBOMs bool

	// VerifySignatures determines if signatures should be verified before replication
	VerifySignatures bool

	// AllowedTypes restricts replication to specific artifact types
	AllowedTypes []ArtifactType

	// DeniedTypes excludes specific artifact types from replication
	DeniedTypes []ArtifactType

	// PreserveAnnotations determines if annotations should be preserved
	PreserveAnnotations bool

	// AddAnnotations adds additional annotations during replication
	AddAnnotations map[string]string

	// SkipExisting skips replication if artifact already exists at destination
	SkipExisting bool

	// Timeout is the replication timeout per artifact
	Timeout time.Duration
}

// DefaultReplicationOptions returns default replication options
func DefaultReplicationOptions() *ReplicationOptions {
	return &ReplicationOptions{
		IncludeLayers:       true,
		IncludeReferrers:    true,
		IncludeSignatures:   true,
		IncludeSBOMs:        true,
		VerifySignatures:    false,
		AllowedTypes:        nil, // nil means allow all
		DeniedTypes:         nil,
		PreserveAnnotations: true,
		AddAnnotations:      nil,
		SkipExisting:        false,
		Timeout:             5 * time.Minute,
	}
}

// IsAllowed checks if an artifact type is allowed for replication
func (opts *ReplicationOptions) IsAllowed(artifactType ArtifactType) bool {
	// Check denied list first
	for _, denied := range opts.DeniedTypes {
		if denied == artifactType {
			return false
		}
	}

	// If allowed list is empty, allow all (except denied)
	if len(opts.AllowedTypes) == 0 {
		return true
	}

	// Check allowed list
	for _, allowed := range opts.AllowedTypes {
		if allowed == artifactType {
			return true
		}
	}

	return false
}

// Client interface defines methods for working with OCI artifacts
type Client interface {
	// GetArtifact retrieves artifact metadata
	GetArtifact(ctx context.Context, ref string) (*Artifact, error)

	// PushArtifact pushes an artifact to a registry
	PushArtifact(ctx context.Context, artifact *Artifact) error

	// PullArtifact pulls an artifact from a registry
	PullArtifact(ctx context.Context, ref string) (*Artifact, error)

	// CopyArtifact copies an artifact from source to destination
	CopyArtifact(ctx context.Context, src, dst string, opts *ReplicationOptions) error

	// ListReferrers lists artifacts that reference the given artifact
	ListReferrers(ctx context.Context, ref string, artifactType string) ([]Referrer, error)

	// GetManifest retrieves the manifest for an artifact
	GetManifest(ctx context.Context, ref string) (*v1.Manifest, error)

	// GetIndex retrieves the index for a multi-arch artifact
	GetIndex(ctx context.Context, ref string) (*v1.Index, error)

	// SupportedTypes returns the list of supported artifact types
	SupportedTypes() []ArtifactType
}

// DetectArtifactType detects the artifact type from media type
func DetectArtifactType(mediaType string) ArtifactType {
	switch {
	case strings.Contains(mediaType, "helm.chart"):
		return ArtifactTypeHelm
	case strings.Contains(mediaType, "helm.config"):
		return ArtifactTypeHelmConfig
	case strings.Contains(mediaType, "helm.provenance"):
		return ArtifactTypeHelmProvenance
	case strings.Contains(mediaType, "wasm.content"):
		return ArtifactTypeWASM
	case strings.Contains(mediaType, "wasm.config"):
		return ArtifactTypeWASMConfig
	case strings.Contains(mediaType, "ml.model") && !strings.Contains(mediaType, "config"):
		return ArtifactTypeMLModel
	case strings.Contains(mediaType, "ml.model.config"):
		return ArtifactTypeMLModelConfig
	case strings.Contains(mediaType, "sbom"):
		return ArtifactTypeSBOM
	case strings.Contains(mediaType, "spdx"):
		return ArtifactTypeSPDX
	case strings.Contains(mediaType, "cyclonedx"):
		return ArtifactTypeCycloneDX
	case strings.Contains(mediaType, "cosign"):
		return ArtifactTypeSignature
	case strings.Contains(mediaType, "in-toto"):
		return ArtifactTypeAttestation
	case mediaType == string(ArtifactTypeOCIImage):
		return ArtifactTypeOCIImage
	case mediaType == string(ArtifactTypeOCIImageIndex):
		return ArtifactTypeOCIImageIndex
	case mediaType == string(ArtifactTypeDockerManifest):
		return ArtifactTypeDockerManifest
	case mediaType == string(ArtifactTypeDockerManifestList):
		return ArtifactTypeDockerManifestList
	default:
		return ArtifactTypeGeneric
	}
}

// IsContainerImage checks if the artifact is a container image
func (a *Artifact) IsContainerImage() bool {
	return a.ArtifactType == ArtifactTypeOCIImage ||
		a.ArtifactType == ArtifactTypeDockerManifest
}

// IsMultiArch checks if the artifact is a multi-architecture image
func (a *Artifact) IsMultiArch() bool {
	return a.ArtifactType == ArtifactTypeOCIImageIndex ||
		a.ArtifactType == ArtifactTypeDockerManifestList
}

// IsHelm checks if the artifact is a Helm chart
func (a *Artifact) IsHelm() bool {
	return a.ArtifactType == ArtifactTypeHelm
}

// IsWASM checks if the artifact is a WebAssembly module
func (a *Artifact) IsWASM() bool {
	return a.ArtifactType == ArtifactTypeWASM
}

// IsMLModel checks if the artifact is a machine learning model
func (a *Artifact) IsMLModel() bool {
	return a.ArtifactType == ArtifactTypeMLModel
}

// IsSBOM checks if the artifact is a Software Bill of Materials
func (a *Artifact) IsSBOM() bool {
	return a.ArtifactType == ArtifactTypeSBOM ||
		a.ArtifactType == ArtifactTypeSPDX ||
		a.ArtifactType == ArtifactTypeCycloneDX
}

// IsSignature checks if the artifact is a signature
func (a *Artifact) IsSignature() bool {
	return a.ArtifactType == ArtifactTypeSignature ||
		a.ArtifactType == ArtifactTypeAttestation
}

// GetAnnotation retrieves an annotation value by key
func (a *Artifact) GetAnnotation(key string) (string, bool) {
	if a.Annotations == nil {
		return "", false
	}
	val, ok := a.Annotations[key]
	return val, ok
}

// SetAnnotation sets an annotation value
func (a *Artifact) SetAnnotation(key, value string) {
	if a.Annotations == nil {
		a.Annotations = make(map[string]string)
	}
	a.Annotations[key] = value
}

// Validate validates the artifact structure
func (a *Artifact) Validate() error {
	if a.Reference == "" && a.Digest == "" {
		return fmt.Errorf("artifact must have either reference or digest")
	}

	if a.Repository == "" {
		return fmt.Errorf("artifact repository is required")
	}

	if a.MediaType == "" {
		return fmt.Errorf("artifact media type is required")
	}

	if a.Manifest == nil && a.Index == nil {
		return fmt.Errorf("artifact must have either manifest or index")
	}

	return nil
}

// String returns a string representation of the artifact
func (a *Artifact) String() string {
	if a.Reference != "" {
		return a.Reference
	}
	if a.Digest != "" {
		return fmt.Sprintf("%s@%s", a.Repository, a.Digest)
	}
	if a.Tag != "" {
		return fmt.Sprintf("%s:%s", a.Repository, a.Tag)
	}
	return a.Repository
}
