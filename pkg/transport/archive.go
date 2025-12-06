package transport

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DockerArchiveTransport implements the docker-archive: transport for tar archives
// Compatible with Docker's save/load format and Skopeo's docker-archive: transport
type DockerArchiveTransport struct{}

// NewDockerArchiveTransport creates a new docker-archive transport
func NewDockerArchiveTransport() *DockerArchiveTransport {
	return &DockerArchiveTransport{}
}

// Name returns the transport name
func (t *DockerArchiveTransport) Name() string {
	return "docker-archive"
}

// ValidateReference validates an archive reference
func (t *DockerArchiveTransport) ValidateReference(ref string) error {
	if ref == "" {
		return fmt.Errorf("archive path cannot be empty")
	}
	return nil
}

// ParseReference parses an archive reference
// Format: /path/to/archive.tar[:reference]
func (t *DockerArchiveTransport) ParseReference(ref string) (Reference, error) {
	if err := t.ValidateReference(ref); err != nil {
		return nil, err
	}

	// Split path and reference (tag or digest)
	path := ref
	reference := ""

	// Check for :tag or @digest suffix
	if idx := strings.LastIndex(ref, ":"); idx > 0 {
		// Make sure it's not part of a Windows path (C:\...)
		if !(idx == 1 && len(ref) > 2 && ref[idx+1] == '\\') {
			path = ref[:idx]
			reference = ref[idx+1:]
		}
	} else if idx := strings.LastIndex(ref, "@"); idx > 0 {
		path = ref[:idx]
		reference = ref[idx+1:]
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &DockerArchiveReference{
		transport: t,
		path:      absPath,
		reference: reference,
	}, nil
}

// DockerArchiveReference represents a docker-archive image reference
type DockerArchiveReference struct {
	transport *DockerArchiveTransport
	path      string
	reference string // optional tag or digest
}

// Transport returns the transport
func (r *DockerArchiveReference) Transport() Transport {
	return r.transport
}

// StringWithinTransport returns the path
func (r *DockerArchiveReference) StringWithinTransport() string {
	if r.reference != "" {
		return r.path + ":" + r.reference
	}
	return r.path
}

// DockerReference returns equivalent docker:// reference
func (r *DockerArchiveReference) DockerReference() string {
	// Archive doesn't have a direct docker:// equivalent
	return ""
}

// PolicyConfigurationIdentity returns identity for policy configuration
func (r *DockerArchiveReference) PolicyConfigurationIdentity() string {
	return r.path
}

// PolicyConfigurationNamespaces returns namespaces for policy configuration
func (r *DockerArchiveReference) PolicyConfigurationNamespaces() []string {
	return []string{r.path}
}

// NewImage returns an Image for this reference
func (r *DockerArchiveReference) NewImage(ctx context.Context) (Image, error) {
	return &DockerArchiveImage{
		ref:  r,
		path: r.path,
	}, nil
}

// NewImageSource returns an ImageSource for reading
func (r *DockerArchiveReference) NewImageSource(ctx context.Context) (ImageSource, error) {
	// Check if archive exists
	if _, err := os.Stat(r.path); os.IsNotExist(err) {
		return nil, fmt.Errorf("archive does not exist: %s", r.path)
	}

	return &DockerArchiveImageSource{
		ref:  r,
		path: r.path,
	}, nil
}

// NewImageDestination returns an ImageDestination for writing
func (r *DockerArchiveReference) NewImageDestination(ctx context.Context) (ImageDestination, error) {
	return &DockerArchiveImageDestination{
		ref:       r,
		path:      r.path,
		layers:    make(map[string]string),
		tempFiles: make([]string, 0),
	}, nil
}

// DeleteImage deletes the archive file
func (r *DockerArchiveReference) DeleteImage(ctx context.Context) error {
	return os.Remove(r.path)
}

// DockerArchiveManifest represents the manifest.json structure in Docker archives
type DockerArchiveManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// DockerArchiveImage represents an image stored in a docker archive
type DockerArchiveImage struct {
	ref  *DockerArchiveReference
	path string
}

// Reference returns the image reference
func (i *DockerArchiveImage) Reference() Reference {
	return i.ref
}

// Close releases resources
func (i *DockerArchiveImage) Close() error {
	return nil
}

// Manifest returns the image manifest
func (i *DockerArchiveImage) Manifest(ctx context.Context) ([]byte, string, error) {
	file, err := os.Open(i.path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)

	// Find manifest.json
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Name == "manifest.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read manifest: %w", err)
			}

			// Parse manifest to get config file name
			var manifests []DockerArchiveManifest
			if err := json.Unmarshal(data, &manifests); err != nil {
				return nil, "", fmt.Errorf("failed to parse manifest: %w", err)
			}

			if len(manifests) == 0 {
				return nil, "", fmt.Errorf("no manifests found in archive")
			}

			// Return first manifest
			manifest := manifests[0]
			manifestData, _ := json.Marshal(map[string]interface{}{
				"config": map[string]interface{}{
					"digest": manifest.Config,
				},
				"layers": manifest.Layers,
			})

			return manifestData, "application/vnd.docker.distribution.manifest.v2+json", nil
		}
	}

	return nil, "", fmt.Errorf("manifest.json not found in archive")
}

// Inspect returns image metadata
func (i *DockerArchiveImage) Inspect(ctx context.Context) (*ImageInspectInfo, error) {
	file, err := os.Open(i.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)

	var manifest *DockerArchiveManifest
	var configData []byte

	// Read manifest and config
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Name == "manifest.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read manifest: %w", err)
			}

			var manifests []DockerArchiveManifest
			if err := json.Unmarshal(data, &manifests); err != nil {
				return nil, fmt.Errorf("failed to parse manifest: %w", err)
			}

			if len(manifests) > 0 {
				manifest = &manifests[0]
			}
		} else if manifest != nil && header.Name == manifest.Config {
			configData, err = io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read config: %w", err)
			}
			break
		}
	}

	if manifest == nil {
		return nil, fmt.Errorf("manifest.json not found in archive")
	}

	info := &ImageInspectInfo{
		Name:     filepath.Base(i.path),
		RepoTags: manifest.RepoTags,
		Layers:   manifest.Layers,
	}

	if len(configData) > 0 {
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

	return info, nil
}

// LayerInfos returns information about image layers
func (i *DockerArchiveImage) LayerInfos() []LayerInfo {
	file, err := os.Open(i.path)
	if err != nil {
		return nil
	}
	defer file.Close()

	tr := tar.NewReader(file)

	// Find manifest.json
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil
		}

		if header.Name == "manifest.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil
			}

			var manifests []DockerArchiveManifest
			if err := json.Unmarshal(data, &manifests); err != nil {
				return nil
			}

			if len(manifests) == 0 {
				return nil
			}

			manifest := manifests[0]
			infos := make([]LayerInfo, len(manifest.Layers))
			for idx, layer := range manifest.Layers {
				infos[idx] = LayerInfo{
					Digest:    layer,
					MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				}
			}
			return infos
		}
	}

	return nil
}

// Size returns the archive size in bytes
func (i *DockerArchiveImage) Size() (int64, error) {
	stat, err := os.Stat(i.path)
	if err != nil {
		return 0, fmt.Errorf("failed to stat archive: %w", err)
	}
	return stat.Size(), nil
}

// DockerArchiveImageSource implements ImageSource for docker-archive transport
type DockerArchiveImageSource struct {
	ref  *DockerArchiveReference
	path string
	mu   sync.Mutex
}

// Reference returns the image reference
func (s *DockerArchiveImageSource) Reference() Reference {
	return s.ref
}

// Close releases resources
func (s *DockerArchiveImageSource) Close() error {
	return nil
}

// GetManifest returns the image manifest
func (s *DockerArchiveImageSource) GetManifest(ctx context.Context, instanceDigest *string) ([]byte, string, error) {
	file, err := os.Open(s.path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Name == "manifest.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read manifest: %w", err)
			}
			return data, "application/vnd.docker.distribution.manifest.v2+json", nil
		}
	}

	return nil, "", fmt.Errorf("manifest.json not found in archive")
}

// GetBlob returns a blob (layer or config)
func (s *DockerArchiveImageSource) GetBlob(ctx context.Context, info LayerInfo, cache BlobInfoCache) (io.ReadCloser, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open archive: %w", err)
	}

	tr := tar.NewReader(file)

	// Find the blob in the archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			file.Close()
			return nil, 0, fmt.Errorf("blob not found in archive: %s", info.Digest)
		}
		if err != nil {
			file.Close()
			return nil, 0, fmt.Errorf("failed to read tar: %w", err)
		}

		// Match by digest or filename
		if header.Name == info.Digest || strings.Contains(header.Name, info.Digest) {
			// Read the blob data into memory
			data, err := io.ReadAll(tr)
			if err != nil {
				file.Close()
				return nil, 0, fmt.Errorf("failed to read blob: %w", err)
			}
			file.Close()

			// Return a reader from the in-memory data
			return io.NopCloser(strings.NewReader(string(data))), int64(len(data)), nil
		}
	}
}

// HasThreadSafeGetBlob returns false because tar reading must be sequential
func (s *DockerArchiveImageSource) HasThreadSafeGetBlob() bool {
	return false
}

// GetSignatures returns image signatures
func (s *DockerArchiveImageSource) GetSignatures(ctx context.Context, instanceDigest *string) ([][]byte, error) {
	// Docker archives don't typically contain signatures
	return nil, nil
}

// LayerInfosForCopy returns layer infos optimized for copying
func (s *DockerArchiveImageSource) LayerInfosForCopy(ctx context.Context) ([]LayerInfo, error) {
	file, err := os.Open(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Name == "manifest.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read manifest: %w", err)
			}

			var manifests []DockerArchiveManifest
			if err := json.Unmarshal(data, &manifests); err != nil {
				return nil, fmt.Errorf("failed to parse manifest: %w", err)
			}

			if len(manifests) == 0 {
				return nil, fmt.Errorf("no manifests found in archive")
			}

			manifest := manifests[0]
			infos := make([]LayerInfo, len(manifest.Layers))
			for idx, layer := range manifest.Layers {
				infos[idx] = LayerInfo{
					Digest:    layer,
					MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				}
			}
			return infos, nil
		}
	}

	return nil, fmt.Errorf("manifest.json not found in archive")
}

// DockerArchiveImageDestination implements ImageDestination for docker-archive transport
type DockerArchiveImageDestination struct {
	ref       *DockerArchiveReference
	path      string
	config    []byte
	manifest  []byte
	layers    map[string]string // digest -> temp file path
	tempFiles []string
	mu        sync.Mutex
}

// Reference returns the image reference
func (d *DockerArchiveImageDestination) Reference() Reference {
	return d.ref
}

// Close releases resources and cleans up temp files
func (d *DockerArchiveImageDestination) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, tempFile := range d.tempFiles {
		os.Remove(tempFile)
	}
	d.tempFiles = nil
	return nil
}

// SupportedManifestMIMETypes returns supported manifest MIME types
func (d *DockerArchiveImageDestination) SupportedManifestMIMETypes() []string {
	return []string{
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
	}
}

// SupportsSignatures returns error because archives don't support signatures
func (d *DockerArchiveImageDestination) SupportsSignatures(ctx context.Context) error {
	return fmt.Errorf("docker-archive transport does not support signatures")
}

// DesiredLayerCompression returns desired layer compression
func (d *DockerArchiveImageDestination) DesiredLayerCompression() LayerCompression {
	return PreserveOriginal
}

// AcceptsForeignLayerURLs returns false
func (d *DockerArchiveImageDestination) AcceptsForeignLayerURLs() bool {
	return false
}

// MustMatchRuntimeOS returns false
func (d *DockerArchiveImageDestination) MustMatchRuntimeOS() bool {
	return false
}

// IgnoresEmbeddedDockerReference returns false
func (d *DockerArchiveImageDestination) IgnoresEmbeddedDockerReference() bool {
	return false
}

// HasThreadSafePutBlob returns true
func (d *DockerArchiveImageDestination) HasThreadSafePutBlob() bool {
	return true
}

// PutBlob writes a blob (layer or config)
func (d *DockerArchiveImageDestination) PutBlob(ctx context.Context, stream io.Reader, inputInfo LayerInfo, cache BlobInfoCache, isConfig bool) (LayerInfo, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create temp file for this blob
	tempFile, err := os.CreateTemp("", "freightliner-archive-*.tmp")
	if err != nil {
		return LayerInfo{}, fmt.Errorf("failed to create temp file: %w", err)
	}
	d.tempFiles = append(d.tempFiles, tempFile.Name())

	// Write blob to temp file
	written, err := io.Copy(tempFile, stream)
	if err != nil {
		tempFile.Close()
		return LayerInfo{}, fmt.Errorf("failed to write blob: %w", err)
	}
	tempFile.Close()

	if isConfig {
		// Read config data
		data, err := os.ReadFile(tempFile.Name())
		if err != nil {
			return LayerInfo{}, fmt.Errorf("failed to read config: %w", err)
		}
		d.config = data
	} else {
		// Store layer temp file path
		d.layers[inputInfo.Digest] = tempFile.Name()
	}

	outputInfo := inputInfo
	outputInfo.Size = written
	return outputInfo, nil
}

// TryReusingBlob checks if a blob can be reused
func (d *DockerArchiveImageDestination) TryReusingBlob(ctx context.Context, info LayerInfo, cache BlobInfoCache, canSubstitute bool) (bool, LayerInfo, error) {
	// Cannot reuse blobs from archives
	return false, LayerInfo{}, nil
}

// PutManifest writes the image manifest
func (d *DockerArchiveImageDestination) PutManifest(ctx context.Context, manifest []byte, instanceDigest *string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.manifest = manifest
	return nil
}

// PutSignatures writes image signatures
func (d *DockerArchiveImageDestination) PutSignatures(ctx context.Context, signatures [][]byte, instanceDigest *string) error {
	return fmt.Errorf("docker-archive transport does not support signatures")
}

// Commit commits the image by creating the tar archive
func (d *DockerArchiveImageDestination) Commit(ctx context.Context, unparsedToplevel interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.manifest == nil {
		return fmt.Errorf("manifest not set")
	}

	// Create the archive file
	file, err := os.Create(d.path)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	// Write manifest.json
	manifestData, err := json.Marshal([]DockerArchiveManifest{
		{
			Config:   "config.json",
			RepoTags: []string{d.ref.reference},
			Layers:   d.getLayerFilenames(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := tw.WriteHeader(&tar.Header{
		Name: "manifest.json",
		Mode: 0644,
		Size: int64(len(manifestData)),
	}); err != nil {
		return fmt.Errorf("failed to write manifest header: %w", err)
	}

	if _, err := tw.Write(manifestData); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	// Write config.json
	if d.config != nil {
		if err := tw.WriteHeader(&tar.Header{
			Name: "config.json",
			Mode: 0644,
			Size: int64(len(d.config)),
		}); err != nil {
			return fmt.Errorf("failed to write config header: %w", err)
		}

		if _, err := tw.Write(d.config); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
	}

	// Write layers
	for digest, tempPath := range d.layers {
		layerData, err := os.ReadFile(tempPath)
		if err != nil {
			return fmt.Errorf("failed to read layer: %w", err)
		}

		// Use digest as filename
		filename := digest
		if strings.Contains(digest, ":") {
			filename = strings.ReplaceAll(digest, ":", "-") + ".tar"
		}

		if err := tw.WriteHeader(&tar.Header{
			Name: filename,
			Mode: 0644,
			Size: int64(len(layerData)),
		}); err != nil {
			return fmt.Errorf("failed to write layer header: %w", err)
		}

		if _, err := tw.Write(layerData); err != nil {
			return fmt.Errorf("failed to write layer: %w", err)
		}
	}

	return nil
}

// getLayerFilenames returns filenames for all layers
func (d *DockerArchiveImageDestination) getLayerFilenames() []string {
	filenames := make([]string, 0, len(d.layers))
	for digest := range d.layers {
		filename := digest
		if strings.Contains(digest, ":") {
			filename = strings.ReplaceAll(digest, ":", "-") + ".tar"
		}
		filenames = append(filenames, filename)
	}
	return filenames
}

func init() {
	// Register docker-archive transport
	RegisterTransport(NewDockerArchiveTransport())
}
