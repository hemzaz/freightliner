package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DirectoryTransport implements the dir: transport for local directories
// Compatible with Skopeo's dir: transport
type DirectoryTransport struct{}

// NewDirectoryTransport creates a new directory transport
func NewDirectoryTransport() *DirectoryTransport {
	return &DirectoryTransport{}
}

// Name returns the transport name
func (t *DirectoryTransport) Name() string {
	return "dir"
}

// ValidateReference validates a directory reference
func (t *DirectoryTransport) ValidateReference(ref string) error {
	if ref == "" {
		return fmt.Errorf("directory path cannot be empty")
	}
	return nil
}

// ParseReference parses a directory reference
func (t *DirectoryTransport) ParseReference(ref string) (Reference, error) {
	if err := t.ValidateReference(ref); err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &DirectoryReference{
		transport: t,
		path:      absPath,
	}, nil
}

// DirectoryReference represents a directory-based image reference
type DirectoryReference struct {
	transport *DirectoryTransport
	path      string
}

// Transport returns the transport
func (r *DirectoryReference) Transport() Transport {
	return r.transport
}

// StringWithinTransport returns the path
func (r *DirectoryReference) StringWithinTransport() string {
	return r.path
}

// DockerReference returns equivalent docker:// reference
func (r *DirectoryReference) DockerReference() string {
	// Directory transport doesn't have a docker equivalent
	return ""
}

// PolicyConfigurationIdentity returns identity for policy configuration
func (r *DirectoryReference) PolicyConfigurationIdentity() string {
	return r.path
}

// PolicyConfigurationNamespaces returns namespaces for policy configuration
func (r *DirectoryReference) PolicyConfigurationNamespaces() []string {
	return []string{r.path}
}

// NewImage returns an Image for this reference
func (r *DirectoryReference) NewImage(ctx context.Context) (Image, error) {
	return &DirectoryImage{
		ref:  r,
		path: r.path,
	}, nil
}

// NewImageSource returns an ImageSource for reading
func (r *DirectoryReference) NewImageSource(ctx context.Context) (ImageSource, error) {
	// Check if directory exists
	if _, err := os.Stat(r.path); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", r.path)
	}

	// Check if manifest exists
	manifestPath := filepath.Join(r.path, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest.json not found in directory: %s", r.path)
	}

	return &DirectoryImageSource{
		ref:  r,
		path: r.path,
	}, nil
}

// NewImageDestination returns an ImageDestination for writing
func (r *DirectoryReference) NewImageDestination(ctx context.Context) (ImageDestination, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(r.path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &DirectoryImageDestination{
		ref:  r,
		path: r.path,
	}, nil
}

// DeleteImage deletes the image directory
func (r *DirectoryReference) DeleteImage(ctx context.Context) error {
	return os.RemoveAll(r.path)
}

// DirectoryImage represents an image stored in a directory
type DirectoryImage struct {
	ref  *DirectoryReference
	path string
}

// Reference returns the image reference
func (i *DirectoryImage) Reference() Reference {
	return i.ref
}

// Close releases resources
func (i *DirectoryImage) Close() error {
	return nil
}

// Manifest returns the image manifest
func (i *DirectoryImage) Manifest(ctx context.Context) ([]byte, string, error) {
	manifestPath := filepath.Join(i.path, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read manifest: %w", err)
	}

	// Detect manifest type
	var manifestMap map[string]interface{}
	if err := json.Unmarshal(data, &manifestMap); err != nil {
		return nil, "", fmt.Errorf("failed to parse manifest: %w", err)
	}

	mediaType := "application/vnd.docker.distribution.manifest.v2+json"
	if mt, ok := manifestMap["mediaType"].(string); ok {
		mediaType = mt
	}

	return data, mediaType, nil
}

// Inspect returns image metadata
func (i *DirectoryImage) Inspect(ctx context.Context) (*ImageInspectInfo, error) {
	manifest, _, err := i.Manifest(ctx)
	if err != nil {
		return nil, err
	}

	var manifestMap map[string]interface{}
	if err := json.Unmarshal(manifest, &manifestMap); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	info := &ImageInspectInfo{
		Name: filepath.Base(i.path),
	}

	// Extract config digest
	if config, ok := manifestMap["config"].(map[string]interface{}); ok {
		if digest, ok := config["digest"].(string); ok {
			info.Digest = digest

			// Try to read config blob
			configPath := filepath.Join(i.path, digest+".json")
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
func (i *DirectoryImage) LayerInfos() []LayerInfo {
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
func (i *DirectoryImage) Size() (int64, error) {
	var totalSize int64

	err := filepath.Walk(i.path, func(path string, info os.FileInfo, err error) error {
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

// DirectoryImageSource implements ImageSource for directory transport
type DirectoryImageSource struct {
	ref  *DirectoryReference
	path string
}

// Reference returns the image reference
func (s *DirectoryImageSource) Reference() Reference {
	return s.ref
}

// Close releases resources
func (s *DirectoryImageSource) Close() error {
	return nil
}

// GetManifest returns the image manifest
func (s *DirectoryImageSource) GetManifest(ctx context.Context, instanceDigest *string) ([]byte, string, error) {
	manifestPath := filepath.Join(s.path, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read manifest: %w", err)
	}

	mediaType := "application/vnd.docker.distribution.manifest.v2+json"
	var manifestMap map[string]interface{}
	if err := json.Unmarshal(data, &manifestMap); err == nil {
		if mt, ok := manifestMap["mediaType"].(string); ok {
			mediaType = mt
		}
	}

	return data, mediaType, nil
}

// GetBlob returns a blob (layer or config)
func (s *DirectoryImageSource) GetBlob(ctx context.Context, info LayerInfo, cache BlobInfoCache) (io.ReadCloser, int64, error) {
	// Try digest with and without algorithm prefix
	blobPath := filepath.Join(s.path, info.Digest)
	if _, err := os.Stat(blobPath); os.IsNotExist(err) {
		// Try without algorithm prefix (sha256:xxx -> xxx)
		if len(info.Digest) > 7 && info.Digest[6] == ':' {
			blobPath = filepath.Join(s.path, info.Digest[7:])
		}
	}

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
func (s *DirectoryImageSource) HasThreadSafeGetBlob() bool {
	return true
}

// GetSignatures returns image signatures
func (s *DirectoryImageSource) GetSignatures(ctx context.Context, instanceDigest *string) ([][]byte, error) {
	// Directory transport doesn't support signatures by default
	return nil, nil
}

// LayerInfosForCopy returns layer infos optimized for copying
func (s *DirectoryImageSource) LayerInfosForCopy(ctx context.Context) ([]LayerInfo, error) {
	manifest, _, err := s.GetManifest(ctx, nil)
	if err != nil {
		return nil, err
	}

	var manifestMap map[string]interface{}
	if err := json.Unmarshal(manifest, &manifestMap); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
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

	return infos, nil
}

// DirectoryImageDestination implements ImageDestination for directory transport
type DirectoryImageDestination struct {
	ref     *DirectoryReference
	path    string
	written map[string]bool
}

// Reference returns the image reference
func (d *DirectoryImageDestination) Reference() Reference {
	return d.ref
}

// Close releases resources
func (d *DirectoryImageDestination) Close() error {
	return nil
}

// SupportedManifestMIMETypes returns supported manifest MIME types
func (d *DirectoryImageDestination) SupportedManifestMIMETypes() []string {
	return []string{
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.oci.image.manifest.v1+json",
		"application/vnd.oci.image.index.v1+json",
	}
}

// SupportsSignatures returns true if signatures are supported
func (d *DirectoryImageDestination) SupportsSignatures(ctx context.Context) error {
	return fmt.Errorf("directory transport does not support signatures")
}

// DesiredLayerCompression returns desired layer compression
func (d *DirectoryImageDestination) DesiredLayerCompression() LayerCompression {
	return PreserveOriginal
}

// AcceptsForeignLayerURLs returns true if foreign layer URLs are accepted
func (d *DirectoryImageDestination) AcceptsForeignLayerURLs() bool {
	return false
}

// MustMatchRuntimeOS returns true if runtime OS must match
func (d *DirectoryImageDestination) MustMatchRuntimeOS() bool {
	return false
}

// IgnoresEmbeddedDockerReference returns true if embedded docker reference is ignored
func (d *DirectoryImageDestination) IgnoresEmbeddedDockerReference() bool {
	return true
}

// HasThreadSafePutBlob returns true if PutBlob can be called concurrently
func (d *DirectoryImageDestination) HasThreadSafePutBlob() bool {
	return true
}

// PutBlob writes a blob (layer or config)
func (d *DirectoryImageDestination) PutBlob(ctx context.Context, stream io.Reader, inputInfo LayerInfo, cache BlobInfoCache, isConfig bool) (LayerInfo, error) {
	if d.written == nil {
		d.written = make(map[string]bool)
	}

	// Use digest as filename
	filename := inputInfo.Digest
	if len(filename) > 7 && filename[6] == ':' {
		filename = filename[7:] // Remove algorithm prefix
	}

	blobPath := filepath.Join(d.path, filename)

	// Check if already written
	if d.written[blobPath] {
		return inputInfo, nil
	}

	// Write blob to file
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
func (d *DirectoryImageDestination) TryReusingBlob(ctx context.Context, info LayerInfo, cache BlobInfoCache, canSubstitute bool) (bool, LayerInfo, error) {
	filename := info.Digest
	if len(filename) > 7 && filename[6] == ':' {
		filename = filename[7:]
	}

	blobPath := filepath.Join(d.path, filename)
	if _, err := os.Stat(blobPath); err == nil {
		// Blob already exists
		return true, info, nil
	}

	return false, LayerInfo{}, nil
}

// PutManifest writes the image manifest
func (d *DirectoryImageDestination) PutManifest(ctx context.Context, manifest []byte, instanceDigest *string) error {
	manifestPath := filepath.Join(d.path, "manifest.json")
	return os.WriteFile(manifestPath, manifest, 0644)
}

// PutSignatures writes image signatures
func (d *DirectoryImageDestination) PutSignatures(ctx context.Context, signatures [][]byte, instanceDigest *string) error {
	return fmt.Errorf("directory transport does not support signatures")
}

// Commit commits the image
func (d *DirectoryImageDestination) Commit(ctx context.Context, unparsedToplevel interface{}) error {
	// No additional commit operation needed for directory transport
	return nil
}

func init() {
	// Register directory transport
	RegisterTransport(NewDirectoryTransport())
}
