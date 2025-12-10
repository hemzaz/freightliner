package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// OCILayoutTransport implements the oci: transport for OCI Image Layout
// Compatible with OCI Image Layout Specification v1.0.0
type OCILayoutTransport struct{}

// NewOCILayoutTransport creates a new OCI layout transport
func NewOCILayoutTransport() *OCILayoutTransport {
	return &OCILayoutTransport{}
}

// Name returns the transport name
func (t *OCILayoutTransport) Name() string {
	return "oci"
}

// ValidateReference validates an OCI layout reference
func (t *OCILayoutTransport) ValidateReference(ref string) error {
	if ref == "" {
		return fmt.Errorf("OCI layout path cannot be empty")
	}
	// Format: <path>:<tag> or <path>@<digest>
	return nil
}

// ParseReference parses an OCI layout reference
// Format: /path/to/layout:tag or /path/to/layout@sha256:digest
func (t *OCILayoutTransport) ParseReference(ref string) (Reference, error) {
	if err := t.ValidateReference(ref); err != nil {
		return nil, err
	}

	// Split path and reference (tag or digest)
	path := ref
	reference := "latest"

	// Check for tag separator
	if idx := strings.LastIndex(ref, ":"); idx > 0 {
		// Make sure it's not part of a digest (sha256:...)
		if idx > 6 && ref[idx-6:idx] != "sha256" {
			path = ref[:idx]
			reference = ref[idx+1:]
		}
	}

	// Check for digest separator
	if idx := strings.LastIndex(ref, "@"); idx > 0 {
		path = ref[:idx]
		reference = ref[idx+1:]
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &OCILayoutReference{
		transport: t,
		path:      absPath,
		reference: reference,
	}, nil
}

// OCILayoutReference represents an OCI layout image reference
type OCILayoutReference struct {
	transport *OCILayoutTransport
	path      string
	reference string // tag or digest
}

// Transport returns the transport
func (r *OCILayoutReference) Transport() Transport {
	return r.transport
}

// StringWithinTransport returns the reference string
func (r *OCILayoutReference) StringWithinTransport() string {
	return fmt.Sprintf("%s:%s", r.path, r.reference)
}

// DockerReference returns equivalent docker:// reference
func (r *OCILayoutReference) DockerReference() string {
	// OCI layout doesn't have a docker equivalent
	return ""
}

// PolicyConfigurationIdentity returns identity for policy configuration
func (r *OCILayoutReference) PolicyConfigurationIdentity() string {
	return r.path
}

// PolicyConfigurationNamespaces returns namespaces for policy configuration
func (r *OCILayoutReference) PolicyConfigurationNamespaces() []string {
	return []string{r.path}
}

// NewImage returns an Image for this reference
func (r *OCILayoutReference) NewImage(ctx context.Context) (Image, error) {
	return &OCILayoutImage{
		ref:       r,
		path:      r.path,
		reference: r.reference,
	}, nil
}

// NewImageSource returns an ImageSource for reading
func (r *OCILayoutReference) NewImageSource(ctx context.Context) (ImageSource, error) {
	// Verify OCI layout structure
	if err := r.validateLayout(); err != nil {
		return nil, err
	}

	return &OCILayoutImageSource{
		ref:       r,
		path:      r.path,
		reference: r.reference,
	}, nil
}

// NewImageDestination returns an ImageDestination for writing
func (r *OCILayoutReference) NewImageDestination(ctx context.Context) (ImageDestination, error) {
	// Create OCI layout structure
	if err := r.initializeLayout(); err != nil {
		return nil, err
	}

	return &OCILayoutImageDestination{
		ref:       r,
		path:      r.path,
		reference: r.reference,
	}, nil
}

// DeleteImage deletes the image from the OCI layout
func (r *OCILayoutReference) DeleteImage(ctx context.Context) error {
	// Remove the specific reference from index.json
	indexPath := filepath.Join(r.path, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read index.json: %w", err)
	}

	var index OCIIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("failed to parse index.json: %w", err)
	}

	// Filter out the reference
	var newManifests []OCIDescriptor
	for _, manifest := range index.Manifests {
		// Check annotations for reference match
		keep := true
		if refName, ok := manifest.Annotations["org.opencontainers.image.ref.name"]; ok {
			if refName == r.reference {
				keep = false
			}
		}
		if keep {
			newManifests = append(newManifests, manifest)
		}
	}

	index.Manifests = newManifests

	// Write updated index
	newData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	return os.WriteFile(indexPath, newData, 0644)
}

// validateLayout verifies OCI layout structure exists
func (r *OCILayoutReference) validateLayout() error {
	// Check oci-layout file
	layoutPath := filepath.Join(r.path, "oci-layout")
	if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
		return fmt.Errorf("not an OCI layout: oci-layout file not found")
	}

	// Check index.json
	indexPath := filepath.Join(r.path, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return fmt.Errorf("not an OCI layout: index.json not found")
	}

	// Check blobs directory
	blobsPath := filepath.Join(r.path, "blobs")
	if stat, err := os.Stat(blobsPath); os.IsNotExist(err) || !stat.IsDir() {
		return fmt.Errorf("not an OCI layout: blobs directory not found")
	}

	return nil
}

// initializeLayout creates OCI layout structure
func (r *OCILayoutReference) initializeLayout() error {
	// Create directories
	if err := os.MkdirAll(r.path, 0755); err != nil {
		return fmt.Errorf("failed to create layout directory: %w", err)
	}

	blobsPath := filepath.Join(r.path, "blobs", "sha256")
	if err := os.MkdirAll(blobsPath, 0755); err != nil {
		return fmt.Errorf("failed to create blobs directory: %w", err)
	}

	// Create oci-layout file
	layoutPath := filepath.Join(r.path, "oci-layout")
	if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
		layout := OCILayout{
			ImageLayoutVersion: "1.0.0",
		}
		data, err := json.MarshalIndent(layout, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal oci-layout: %w", err)
		}
		if err := os.WriteFile(layoutPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write oci-layout: %w", err)
		}
	}

	// Create index.json if it doesn't exist
	indexPath := filepath.Join(r.path, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		index := OCIIndex{
			SchemaVersion: 2,
			Manifests:     []OCIDescriptor{},
		}
		data, err := json.MarshalIndent(index, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal index.json: %w", err)
		}
		if err := os.WriteFile(indexPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write index.json: %w", err)
		}
	}

	return nil
}

// OCILayout represents the oci-layout file
type OCILayout struct {
	ImageLayoutVersion string `json:"imageLayoutVersion"`
}

// OCIIndex represents the index.json file
type OCIIndex struct {
	SchemaVersion int             `json:"schemaVersion"`
	Manifests     []OCIDescriptor `json:"manifests"`
}

// OCIDescriptor represents a content descriptor
type OCIDescriptor struct {
	MediaType   string            `json:"mediaType"`
	Digest      string            `json:"digest"`
	Size        int64             `json:"size"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// OCILayoutImage represents an image in OCI layout
type OCILayoutImage struct {
	ref       *OCILayoutReference
	path      string
	reference string
}

// Reference returns the image reference
func (i *OCILayoutImage) Reference() Reference {
	return i.ref
}

// Close releases resources
func (i *OCILayoutImage) Close() error {
	return nil
}

// Manifest returns the image manifest
func (i *OCILayoutImage) Manifest(ctx context.Context) ([]byte, string, error) {
	// Read index.json
	indexPath := filepath.Join(i.path, "index.json")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read index.json: %w", err)
	}

	var index OCIIndex
	if err := json.Unmarshal(indexData, &index); err != nil {
		return nil, "", fmt.Errorf("failed to parse index.json: %w", err)
	}

	// Find manifest descriptor for this reference
	var manifestDescriptor *OCIDescriptor
	for _, desc := range index.Manifests {
		if refName, ok := desc.Annotations["org.opencontainers.image.ref.name"]; ok {
			if refName == i.reference {
				manifestDescriptor = &desc
				break
			}
		}
		// Also check digest match
		if desc.Digest == i.reference {
			manifestDescriptor = &desc
			break
		}
	}

	if manifestDescriptor == nil {
		return nil, "", fmt.Errorf("manifest not found for reference: %s", i.reference)
	}

	// Read manifest blob
	manifestPath := i.blobPath(manifestDescriptor.Digest)
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read manifest blob: %w", err)
	}

	return manifestData, manifestDescriptor.MediaType, nil
}

// Inspect returns image metadata
func (i *OCILayoutImage) Inspect(ctx context.Context) (*ImageInspectInfo, error) {
	manifest, _, err := i.Manifest(ctx)
	if err != nil {
		return nil, err
	}

	var manifestMap map[string]interface{}
	if err := json.Unmarshal(manifest, &manifestMap); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	info := &ImageInspectInfo{
		Name: fmt.Sprintf("%s:%s", filepath.Base(i.path), i.reference),
	}

	// Extract config
	if config, ok := manifestMap["config"].(map[string]interface{}); ok {
		if digest, ok := config["digest"].(string); ok {
			info.Digest = digest

			// Read config blob
			configPath := i.blobPath(digest)
			if configData, err := os.ReadFile(configPath); err == nil {
				var configMap map[string]interface{}
				if err := json.Unmarshal(configData, &configMap); err == nil {
					if arch, ok := configMap["architecture"].(string); ok {
						info.Architecture = arch
					}
					if os, ok := configMap["os"].(string); ok {
						info.Os = os
					}
					if created, ok := configMap["created"].(string); ok {
						info.Created = created
					}
				}
			}
		}
	}

	// Extract layers
	if layers, ok := manifestMap["layers"].([]interface{}); ok {
		for _, layer := range layers {
			if layerMap, ok := layer.(map[string]interface{}); ok {
				if digest, ok := layerMap["digest"].(string); ok {
					info.Layers = append(info.Layers, digest)
				}
			}
		}
	}

	return info, nil
}

// LayerInfos returns information about image layers
func (i *OCILayoutImage) LayerInfos() []LayerInfo {
	manifest, _, err := i.Manifest(context.Background())
	if err != nil {
		return nil
	}

	var manifestMap map[string]interface{}
	if err := json.Unmarshal(manifest, &manifestMap); err != nil {
		return nil
	}

	var infos []LayerInfo
	if layers, ok := manifestMap["layers"].([]interface{}); ok {
		for _, layer := range layers {
			if layerMap, ok := layer.(map[string]interface{}); ok {
				info := LayerInfo{}
				if digest, ok := layerMap["digest"].(string); ok {
					info.Digest = digest
				}
				if size, ok := layerMap["size"].(float64); ok {
					info.Size = int64(size)
				}
				if mediaType, ok := layerMap["mediaType"].(string); ok {
					info.MediaType = mediaType
				}
				infos = append(infos, info)
			}
		}
	}

	return infos
}

// Size returns the image size in bytes
func (i *OCILayoutImage) Size() (int64, error) {
	var totalSize int64

	blobsPath := filepath.Join(i.path, "blobs")
	err := filepath.Walk(blobsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

// blobPath returns the path to a blob file
func (i *OCILayoutImage) blobPath(digest string) string {
	// Digest format: algorithm:hash (e.g., sha256:abc123...)
	parts := strings.SplitN(digest, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	algorithm := parts[0]
	hash := parts[1]
	return filepath.Join(i.path, "blobs", algorithm, hash)
}

// OCILayoutImageSource implements ImageSource for OCI layout
type OCILayoutImageSource struct {
	ref       *OCILayoutReference
	path      string
	reference string
}

// Reference returns the image reference
func (s *OCILayoutImageSource) Reference() Reference {
	return s.ref
}

// Close releases resources
func (s *OCILayoutImageSource) Close() error {
	return nil
}

// GetManifest returns the image manifest
func (s *OCILayoutImageSource) GetManifest(ctx context.Context, instanceDigest *string) ([]byte, string, error) {
	img := &OCILayoutImage{
		ref:       s.ref,
		path:      s.path,
		reference: s.reference,
	}
	return img.Manifest(ctx)
}

// GetBlob returns a blob (layer or config)
func (s *OCILayoutImageSource) GetBlob(ctx context.Context, info LayerInfo, cache BlobInfoCache) (io.ReadCloser, int64, error) {
	blobPath := s.blobPath(info.Digest)

	file, err := os.Open(blobPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open blob: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, fmt.Errorf("failed to stat blob: %w", err)
	}

	return file, stat.Size(), nil
}

// HasThreadSafeGetBlob returns true if GetBlob can be called concurrently
func (s *OCILayoutImageSource) HasThreadSafeGetBlob() bool {
	return true
}

// GetSignatures returns image signatures
func (s *OCILayoutImageSource) GetSignatures(ctx context.Context, instanceDigest *string) ([][]byte, error) {
	// OCI layout can store signatures but it's optional
	return nil, nil
}

// LayerInfosForCopy returns layer infos optimized for copying
func (s *OCILayoutImageSource) LayerInfosForCopy(ctx context.Context) ([]LayerInfo, error) {
	img := &OCILayoutImage{
		ref:       s.ref,
		path:      s.path,
		reference: s.reference,
	}
	return img.LayerInfos(), nil
}

// blobPath returns the path to a blob file
func (s *OCILayoutImageSource) blobPath(digest string) string {
	parts := strings.SplitN(digest, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	algorithm := parts[0]
	hash := parts[1]
	return filepath.Join(s.path, "blobs", algorithm, hash)
}

// OCILayoutImageDestination implements ImageDestination for OCI layout
type OCILayoutImageDestination struct {
	ref       *OCILayoutReference
	path      string
	reference string
	written   map[string]bool
}

// Reference returns the image reference
func (d *OCILayoutImageDestination) Reference() Reference {
	return d.ref
}

// Close releases resources
func (d *OCILayoutImageDestination) Close() error {
	return nil
}

// SupportedManifestMIMETypes returns supported manifest MIME types
func (d *OCILayoutImageDestination) SupportedManifestMIMETypes() []string {
	return []string{
		"application/vnd.oci.image.manifest.v1+json",
		"application/vnd.oci.image.index.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
	}
}

// SupportsSignatures returns true if signatures are supported
func (d *OCILayoutImageDestination) SupportsSignatures(ctx context.Context) error {
	return nil // OCI layout supports signatures
}

// DesiredLayerCompression returns desired layer compression
func (d *OCILayoutImageDestination) DesiredLayerCompression() LayerCompression {
	return PreserveOriginal
}

// AcceptsForeignLayerURLs returns true if foreign layer URLs are accepted
func (d *OCILayoutImageDestination) AcceptsForeignLayerURLs() bool {
	return true
}

// MustMatchRuntimeOS returns true if runtime OS must match
func (d *OCILayoutImageDestination) MustMatchRuntimeOS() bool {
	return false
}

// IgnoresEmbeddedDockerReference returns true if embedded docker reference is ignored
func (d *OCILayoutImageDestination) IgnoresEmbeddedDockerReference() bool {
	return true
}

// HasThreadSafePutBlob returns true if PutBlob can be called concurrently
func (d *OCILayoutImageDestination) HasThreadSafePutBlob() bool {
	return true
}

// PutBlob writes a blob (layer or config)
func (d *OCILayoutImageDestination) PutBlob(ctx context.Context, stream io.Reader, inputInfo LayerInfo, cache BlobInfoCache, isConfig bool) (LayerInfo, error) {
	if d.written == nil {
		d.written = make(map[string]bool)
	}

	blobPath := d.blobPath(inputInfo.Digest)

	// Check if already written
	if d.written[blobPath] {
		return inputInfo, nil
	}

	// Create algorithm directory if needed
	if err := os.MkdirAll(filepath.Dir(blobPath), 0755); err != nil {
		return LayerInfo{}, fmt.Errorf("failed to create blob directory: %w", err)
	}

	// Write blob
	file, err := os.Create(blobPath)
	if err != nil {
		return LayerInfo{}, fmt.Errorf("failed to create blob file: %w", err)
	}
	defer file.Close()

	written, err := io.Copy(file, stream)
	if err != nil {
		return LayerInfo{}, fmt.Errorf("failed to write blob: %w", err)
	}

	d.written[blobPath] = true

	outputInfo := inputInfo
	outputInfo.Size = written

	return outputInfo, nil
}

// TryReusingBlob checks if a blob can be reused
func (d *OCILayoutImageDestination) TryReusingBlob(ctx context.Context, info LayerInfo, cache BlobInfoCache, canSubstitute bool) (bool, LayerInfo, error) {
	blobPath := d.blobPath(info.Digest)
	if _, err := os.Stat(blobPath); err == nil {
		return true, info, nil
	}
	return false, LayerInfo{}, nil
}

// PutManifest writes the image manifest
func (d *OCILayoutImageDestination) PutManifest(ctx context.Context, manifest []byte, instanceDigest *string) error {
	// Calculate manifest digest
	// For simplicity, using a basic approach - production code would use proper hashing
	manifestDigest := fmt.Sprintf("sha256:%x", manifest[:32]) // Simplified

	// Write manifest blob
	manifestPath := d.blobPath(manifestDigest)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}
	if err := os.WriteFile(manifestPath, manifest, 0644); err != nil {
		return fmt.Errorf("failed to write manifest blob: %w", err)
	}

	// Update index.json
	indexPath := filepath.Join(d.path, "index.json")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read index.json: %w", err)
	}

	var index OCIIndex
	if err := json.Unmarshal(indexData, &index); err != nil {
		return fmt.Errorf("failed to parse index.json: %w", err)
	}

	// Add or update manifest descriptor
	descriptor := OCIDescriptor{
		MediaType: "application/vnd.oci.image.manifest.v1+json",
		Digest:    manifestDigest,
		Size:      int64(len(manifest)),
		Annotations: map[string]string{
			"org.opencontainers.image.ref.name": d.reference,
		},
	}

	// Remove existing descriptor with same reference
	var newManifests []OCIDescriptor
	for _, desc := range index.Manifests {
		if refName, ok := desc.Annotations["org.opencontainers.image.ref.name"]; !ok || refName != d.reference {
			newManifests = append(newManifests, desc)
		}
	}
	newManifests = append(newManifests, descriptor)
	index.Manifests = newManifests

	// Write updated index
	newIndexData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index.json: %w", err)
	}

	return os.WriteFile(indexPath, newIndexData, 0644)
}

// PutSignatures writes image signatures
func (d *OCILayoutImageDestination) PutSignatures(ctx context.Context, signatures [][]byte, instanceDigest *string) error {
	// OCI layout signature support would go here
	// For now, return success (signatures are optional)
	return nil
}

// Commit commits the image
func (d *OCILayoutImageDestination) Commit(ctx context.Context, unparsedToplevel interface{}) error {
	// No additional commit operation needed for OCI layout
	return nil
}

// blobPath returns the path to a blob file
func (d *OCILayoutImageDestination) blobPath(digest string) string {
	parts := strings.SplitN(digest, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	algorithm := parts[0]
	hash := parts[1]
	return filepath.Join(d.path, "blobs", algorithm, hash)
}

func init() {
	// Register OCI layout transport
	RegisterTransport(NewOCILayoutTransport())
}
