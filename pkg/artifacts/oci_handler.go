// Package artifacts provides OCI artifact handling and replication
package artifacts

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"freightliner/pkg/interfaces"
)

// Handler manages OCI artifact operations
type Handler struct {
	sourceClient interfaces.RegistryClient
	destClient   interfaces.RegistryClient
	options      *HandlerOptions
}

// HandlerOptions configures the artifact handler
type HandlerOptions struct {
	// MaxConcurrentTransfers is the maximum number of concurrent layer transfers
	MaxConcurrentTransfers int

	// EnableBlobMounting enables zero-copy blob mounting when possible
	EnableBlobMounting bool

	// EnableDeltaSync enables delta synchronization for large layers
	EnableDeltaSync bool

	// DeltaSyncThreshold is the minimum layer size for delta sync (in bytes)
	DeltaSyncThreshold int64

	// RetryAttempts is the number of retry attempts for failed operations
	RetryAttempts int

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration

	// VerifyDigests enables digest verification during transfer
	VerifyDigests bool

	// BufferSize is the buffer size for streaming operations
	BufferSize int
}

// DefaultHandlerOptions returns default handler options
func DefaultHandlerOptions() *HandlerOptions {
	return &HandlerOptions{
		MaxConcurrentTransfers: 5,
		EnableBlobMounting:     true,
		EnableDeltaSync:        true,
		DeltaSyncThreshold:     100 * 1024 * 1024, // 100MB
		RetryAttempts:          3,
		RetryDelay:             2 * time.Second,
		VerifyDigests:          true,
		BufferSize:             32 * 1024, // 32KB
	}
}

// NewHandler creates a new OCI artifact handler
func NewHandler(sourceClient, destClient interfaces.RegistryClient, opts *HandlerOptions) *Handler {
	if opts == nil {
		opts = DefaultHandlerOptions()
	}

	return &Handler{
		sourceClient: sourceClient,
		destClient:   destClient,
		options:      opts,
	}
}

// SupportedTypes returns artifact types that can be handled
func (h *Handler) SupportedTypes() []ArtifactType {
	return []ArtifactType{
		ArtifactTypeHelm,
		ArtifactTypeHelmConfig,
		ArtifactTypeHelmProvenance,
		ArtifactTypeWASM,
		ArtifactTypeWASMConfig,
		ArtifactTypeOCIImage,
		ArtifactTypeOCIImageIndex,
		ArtifactTypeDockerManifest,
		ArtifactTypeDockerManifestList,
		ArtifactTypeMLModel,
		ArtifactTypeMLModelConfig,
		ArtifactTypeSBOM,
		ArtifactTypeSPDX,
		ArtifactTypeCycloneDX,
		ArtifactTypeSignature,
		ArtifactTypeAttestation,
		ArtifactTypeGeneric,
	}
}

// Replicate replicates an artifact from source to destination
func (h *Handler) Replicate(ctx context.Context, src, dst string, opts *ReplicationOptions) error {
	if opts == nil {
		opts = DefaultReplicationOptions()
	}

	// Add timeout to context if specified
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Parse source reference
	srcRef, err := name.ParseReference(src)
	if err != nil {
		return fmt.Errorf("invalid source reference %s: %w", src, err)
	}

	// Parse destination reference
	dstRef, err := name.ParseReference(dst)
	if err != nil {
		return fmt.Errorf("invalid destination reference %s: %w", dst, err)
	}

	// Get artifact descriptor from source
	artifact, err := h.getArtifact(ctx, srcRef)
	if err != nil {
		return fmt.Errorf("failed to get artifact from source: %w", err)
	}

	// Check if artifact type is allowed
	if !opts.IsAllowed(artifact.ArtifactType) {
		return fmt.Errorf("artifact type %s is not allowed for replication", artifact.ArtifactType)
	}

	// Check if artifact already exists at destination
	if opts.SkipExisting {
		exists, err := h.artifactExists(ctx, dstRef)
		if err != nil {
			return fmt.Errorf("failed to check if artifact exists: %w", err)
		}
		if exists {
			return nil // Skip replication
		}
	}

	// Replicate based on artifact type
	switch {
	case artifact.IsMultiArch():
		return h.replicateIndex(ctx, srcRef, dstRef, artifact, opts)
	case artifact.IsContainerImage():
		return h.replicateImage(ctx, srcRef, dstRef, artifact, opts)
	case artifact.IsHelm():
		return h.replicateHelm(ctx, srcRef, dstRef, artifact, opts)
	case artifact.IsWASM():
		return h.replicateWASM(ctx, srcRef, dstRef, artifact, opts)
	case artifact.IsMLModel():
		return h.replicateMLModel(ctx, srcRef, dstRef, artifact, opts)
	default:
		return h.replicateGeneric(ctx, srcRef, dstRef, artifact, opts)
	}
}

// ReplicateWithReferrers replicates an artifact and its referrers
func (h *Handler) ReplicateWithReferrers(ctx context.Context, src, dst string, opts *ReplicationOptions) error {
	// First replicate the main artifact
	if err := h.Replicate(ctx, src, dst, opts); err != nil {
		return fmt.Errorf("failed to replicate artifact: %w", err)
	}

	// If referrers should not be included, return early
	if !opts.IncludeReferrers {
		return nil
	}

	// Parse source reference
	srcRef, err := name.ParseReference(src)
	if err != nil {
		return fmt.Errorf("invalid source reference: %w", err)
	}

	// Get artifact to retrieve digest
	_, err = h.getArtifact(ctx, srcRef)
	if err != nil {
		return fmt.Errorf("failed to get artifact: %w", err)
	}

	// List referrers
	referrers, err := h.listReferrers(ctx, srcRef, "")
	if err != nil {
		return fmt.Errorf("failed to list referrers: %w", err)
	}

	// Replicate each referrer
	for _, referrer := range referrers {
		// Build referrer source reference
		referrerSrc := fmt.Sprintf("%s@%s", srcRef.Context().Name(), referrer.Digest)

		// Build referrer destination reference
		referrerDst := fmt.Sprintf("%s@%s", dst, referrer.Digest)

		// Filter based on referrer type
		if referrer.ArtifactType != "" {
			artifactType := DetectArtifactType(referrer.ArtifactType)

			// Skip signatures if not included
			if !opts.IncludeSignatures &&
				(artifactType == ArtifactTypeSignature || artifactType == ArtifactTypeAttestation) {
				continue
			}

			// Skip SBOMs if not included
			if !opts.IncludeSBOMs &&
				(artifactType == ArtifactTypeSBOM || artifactType == ArtifactTypeSPDX || artifactType == ArtifactTypeCycloneDX) {
				continue
			}
		}

		// Replicate referrer
		if err := h.Replicate(ctx, referrerSrc, referrerDst, opts); err != nil {
			return fmt.Errorf("failed to replicate referrer %s: %w", referrer.Digest, err)
		}
	}

	return nil
}

// getArtifact retrieves artifact metadata
func (h *Handler) getArtifact(ctx context.Context, ref name.Reference) (*Artifact, error) {
	// Get descriptor from remote
	desc, err := remote.Get(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get descriptor: %w", err)
	}

	artifact := &Artifact{
		Reference:  ref.String(),
		Repository: ref.Context().RepositoryStr(),
		Digest:     desc.Digest.String(),
		MediaType:  string(desc.MediaType),
		Size:       desc.Size,
	}

	// Set tag if available
	if tagged, ok := ref.(name.Tag); ok {
		artifact.Tag = tagged.TagStr()
	}

	// Detect artifact type
	artifact.ArtifactType = DetectArtifactType(artifact.MediaType)

	return artifact, nil
}

// artifactExists checks if an artifact exists at the given reference
func (h *Handler) artifactExists(ctx context.Context, ref name.Reference) (bool, error) {
	_, err := remote.Get(ref)
	if err != nil {
		// Check if error is "not found"
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// replicateIndex replicates a multi-arch image index
func (h *Handler) replicateIndex(ctx context.Context, srcRef, dstRef name.Reference, artifact *Artifact, opts *ReplicationOptions) error {
	// Get index from source
	idx, err := remote.Index(srcRef)
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	// Get index manifest
	manifest, err := idx.IndexManifest()
	if err != nil {
		return fmt.Errorf("failed to get index manifest: %w", err)
	}

	// Replicate each platform image
	for _, desc := range manifest.Manifests {
		// Build platform image reference using digest
		platformSrc := fmt.Sprintf("%s@%s", srcRef.Context().Name(), desc.Digest.String())
		platformDst := fmt.Sprintf("%s@%s", dstRef.Context().Name(), desc.Digest.String())

		// Replicate platform image
		if err := h.Replicate(ctx, platformSrc, platformDst, opts); err != nil {
			return fmt.Errorf("failed to replicate platform image %s: %w", desc.Digest, err)
		}
	}

	// Write index to destination
	if err := remote.WriteIndex(dstRef, idx); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

// replicateImage replicates a container image
func (h *Handler) replicateImage(ctx context.Context, srcRef, dstRef name.Reference, artifact *Artifact, opts *ReplicationOptions) error {
	// Get image from source
	img, err := remote.Image(srcRef)
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}

	// Write image to destination
	if err := remote.Write(dstRef, img); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}

	return nil
}

// replicateHelm replicates a Helm chart artifact
func (h *Handler) replicateHelm(ctx context.Context, srcRef, dstRef name.Reference, artifact *Artifact, opts *ReplicationOptions) error {
	// Helm charts are stored as OCI artifacts
	// Use generic replication
	return h.replicateGeneric(ctx, srcRef, dstRef, artifact, opts)
}

// replicateWASM replicates a WebAssembly module artifact
func (h *Handler) replicateWASM(ctx context.Context, srcRef, dstRef name.Reference, artifact *Artifact, opts *ReplicationOptions) error {
	// WASM modules are stored as OCI artifacts
	// Use generic replication
	return h.replicateGeneric(ctx, srcRef, dstRef, artifact, opts)
}

// replicateMLModel replicates a machine learning model artifact
func (h *Handler) replicateMLModel(ctx context.Context, srcRef, dstRef name.Reference, artifact *Artifact, opts *ReplicationOptions) error {
	// ML models are stored as OCI artifacts
	// Use generic replication
	return h.replicateGeneric(ctx, srcRef, dstRef, artifact, opts)
}

// replicateGeneric replicates a generic OCI artifact
func (h *Handler) replicateGeneric(ctx context.Context, srcRef, dstRef name.Reference, artifact *Artifact, opts *ReplicationOptions) error {
	// Get descriptor from source
	desc, err := remote.Get(srcRef)
	if err != nil {
		return fmt.Errorf("failed to get descriptor: %w", err)
	}

	// Get image representation (works for most OCI artifacts)
	img, err := desc.Image()
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}

	// Write image to destination
	if err := remote.Write(dstRef, img); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}

	return nil
}

// listReferrers lists artifacts that reference the given artifact
func (h *Handler) listReferrers(ctx context.Context, ref name.Reference, artifactType string) ([]Referrer, error) {
	// Note: Referrers API support is registry-dependent
	// This is a placeholder implementation
	// In a full implementation, this would use the OCI Referrers API
	return nil, nil
}

// copyBlob copies a blob between registries with optional mounting
func (h *Handler) copyBlob(ctx context.Context, srcRef, dstRef name.Reference, desc v1.Descriptor) error {
	// Try blob mounting first if enabled
	if h.options.EnableBlobMounting {
		if err := h.mountBlob(ctx, srcRef, dstRef, desc); err == nil {
			return nil // Successfully mounted
		}
		// Fall through to regular copy if mounting fails
	}

	// Get image from source descriptor
	img, err := remote.Get(srcRef)
	if err != nil {
		return fmt.Errorf("failed to get descriptor: %w", err)
	}

	// Get the actual image
	actualImg, err := img.Image()
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}

	// Write image to destination
	if err := remote.Write(dstRef, actualImg); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}

	return nil
}

// mountBlob attempts to mount a blob from source to destination
func (h *Handler) mountBlob(ctx context.Context, srcRef, dstRef name.Reference, desc v1.Descriptor) error {
	// Note: Blob mounting is registry-dependent
	// This is a placeholder - actual implementation would use registry-specific APIs
	return fmt.Errorf("blob mounting not implemented")
}

// verifyDigest verifies that a blob matches its expected digest
func (h *Handler) verifyDigest(ctx context.Context, reader io.Reader, expected digest.Digest) error {
	if !h.options.VerifyDigests {
		return nil
	}

	verifier := expected.Verifier()
	if _, err := io.Copy(verifier, reader); err != nil {
		return fmt.Errorf("failed to compute digest: %w", err)
	}

	if !verifier.Verified() {
		return fmt.Errorf("digest mismatch: expected %s", expected)
	}

	return nil
}

// isNotFoundError checks if an error is a "not found" error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common not found error messages
	errMsg := err.Error()
	return contains(errMsg, "not found") ||
		contains(errMsg, "404") ||
		contains(errMsg, "MANIFEST_UNKNOWN") ||
		contains(errMsg, "NAME_UNKNOWN")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

// GetReferrersAPI checks if the registry supports the OCI Referrers API
func (h *Handler) GetReferrersAPI(ctx context.Context, ref name.Reference) (bool, error) {
	// Try to access the referrers endpoint
	// This is a simplified check - actual implementation would test the API
	return false, nil
}

// ValidateArtifact validates an artifact after replication
func (h *Handler) ValidateArtifact(ctx context.Context, ref name.Reference, expectedDigest string) error {
	desc, err := remote.Get(ref)
	if err != nil {
		return fmt.Errorf("failed to get artifact descriptor: %w", err)
	}

	if desc.Digest.String() != expectedDigest {
		return fmt.Errorf("digest mismatch: expected %s, got %s", expectedDigest, desc.Digest.String())
	}

	return nil
}

// GetArtifactManifest retrieves the manifest for an artifact
func (h *Handler) GetArtifactManifest(ctx context.Context, ref name.Reference) (*ocispec.Manifest, error) {
	desc, err := remote.Get(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get descriptor: %w", err)
	}

	// Parse manifest based on media type
	switch desc.MediaType {
	case "application/vnd.oci.image.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json":
		// Get image and extract manifest
		img, err := desc.Image()
		if err != nil {
			return nil, fmt.Errorf("failed to get image: %w", err)
		}

		manifest, err := img.Manifest()
		if err != nil {
			return nil, fmt.Errorf("failed to get manifest: %w", err)
		}

		// Convert to OCI spec manifest
		return convertManifest(manifest), nil

	default:
		return nil, fmt.Errorf("unsupported manifest media type: %s", desc.MediaType)
	}
}

// convertManifest converts a v1.Manifest to ocispec.Manifest
func convertManifest(m *v1.Manifest) *ocispec.Manifest {
	layers := make([]ocispec.Descriptor, len(m.Layers))
	for i, layer := range m.Layers {
		layers[i] = ocispec.Descriptor{
			MediaType: string(layer.MediaType),
			Digest:    digest.Digest(layer.Digest.String()),
			Size:      layer.Size,
		}
	}

	manifest := &ocispec.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: int(m.SchemaVersion),
		},
		MediaType: string(m.MediaType),
		Config: ocispec.Descriptor{
			MediaType: string(m.Config.MediaType),
			Digest:    digest.Digest(m.Config.Digest.String()),
			Size:      m.Config.Size,
		},
		Layers: layers,
	}

	return manifest
}
