package generic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/util"
	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// Repository implements the repository interface for generic registries
type Repository struct {
	*common.BaseRepository
	client     *Client
	name       string
	repository name.Repository
}

// GetName returns the repository name
func (r *Repository) GetName() string {
	return r.name
}

// GetFullName returns the full repository name including registry
func (r *Repository) GetFullName() string {
	return r.repository.Name()
}

// GetTag retrieves an image by tag
func (r *Repository) GetTag(ctx context.Context, tag string) (v1.Image, error) {
	if r.BaseRepository != nil {
		return r.BaseRepository.GetTag(ctx, tag)
	}
	return nil, errors.NotImplementedf("base repository not initialized")
}

// DeleteManifest deletes a manifest (required by Repository interface)
func (r *Repository) DeleteManifest(ctx context.Context, digest string) error {
	// Generic registries may not support manifest deletion
	return errors.NotImplementedf("manifest deletion not supported for generic registries")
}

// GetImageReference returns a name.Reference for the given tag
func (r *Repository) GetImageReference(tag string) (name.Reference, error) {
	return name.ParseReference(fmt.Sprintf("%s:%s", r.repository.Name(), tag))
}

// GetLayerReader returns a layer reader for a specific digest
func (r *Repository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	if digest == "" {
		return nil, errors.InvalidInputf("digest cannot be empty")
	}

	// Parse digest as a name.Digest for the go-containerregistry API
	digestRef, err := name.NewDigest(fmt.Sprintf("%s@%s", r.repository.Name(), digest))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse digest")
	}

	// Get the layer from the registry
	layer, err := remote.Layer(digestRef, r.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get layer from registry")
	}

	// Get a reader for the layer (compressed)
	reader, err := layer.Compressed()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get layer reader")
	}

	return reader, nil
}

// GetManifest returns a manifest for the given tag or digest
func (r *Repository) GetManifest(ctx context.Context, ref string) (*interfaces.Manifest, error) {
	if ref == "" {
		return nil, errors.InvalidInputf("reference cannot be empty")
	}

	// Create reference (can be tag or digest)
	reference, err := name.ParseReference(fmt.Sprintf("%s:%s", r.repository.Name(), ref))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse reference")
	}

	// Get the descriptor using go-containerregistry
	desc, err := remote.Get(reference, r.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get manifest from registry")
	}

	// Extract manifest content
	manifestBytes, err := desc.RawManifest()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get raw manifest")
	}

	// Create the manifest object
	manifest := &interfaces.Manifest{
		Content:   manifestBytes,
		MediaType: string(desc.MediaType),
		Digest:    desc.Digest.String(),
	}

	// Parse the manifest to extract schema version and layers
	var parsedManifest struct {
		SchemaVersion int `json:"schemaVersion"`
		Config        *struct {
			Digest    string `json:"digest"`
			Size      int64  `json:"size"`
			MediaType string `json:"mediaType"`
		} `json:"config,omitempty"`
		Layers []struct {
			Digest    string `json:"digest"`
			Size      int64  `json:"size"`
			MediaType string `json:"mediaType"`
		} `json:"layers,omitempty"`
	}

	if err := json.Unmarshal(manifestBytes, &parsedManifest); err == nil {
		// Successfully parsed, populate the fields
		manifest.SchemaVersion = parsedManifest.SchemaVersion

		// Convert layers
		if len(parsedManifest.Layers) > 0 {
			manifest.Layers = make([]interfaces.LayerDescriptor, len(parsedManifest.Layers))
			for i, layer := range parsedManifest.Layers {
				manifest.Layers[i] = interfaces.LayerDescriptor{
					Digest:    layer.Digest,
					Size:      layer.Size,
					MediaType: layer.MediaType,
				}
			}
		}

		// Convert config
		if parsedManifest.Config != nil {
			manifest.Config = &interfaces.LayerDescriptor{
				Digest:    parsedManifest.Config.Digest,
				Size:      parsedManifest.Config.Size,
				MediaType: parsedManifest.Config.MediaType,
			}
		}
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository":    r.name,
		"reference":     ref,
		"digest":        manifest.Digest,
		"mediaType":     manifest.MediaType,
		"schemaVersion": manifest.SchemaVersion,
		"layerCount":    len(manifest.Layers),
	}).Debug("Successfully retrieved manifest from generic registry")

	return manifest, nil
}

// GetConfigBlob fetches a config blob by digest from the repository
// This is needed to determine architecture for single-arch manifests
func (r *Repository) GetConfigBlob(ctx context.Context, digest string) ([]byte, error) {
	if digest == "" {
		return nil, errors.InvalidInputf("digest cannot be empty")
	}

	// Parse digest as a name.Digest for the go-containerregistry API
	digestRef, err := name.NewDigest(fmt.Sprintf("%s@%s", r.repository.Name(), digest))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse digest")
	}

	// Get the layer/blob from the registry
	layer, err := remote.Layer(digestRef, r.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config blob from registry")
	}

	// Get the uncompressed content (config blobs are typically uncompressed)
	reader, err := layer.Uncompressed()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config blob reader")
	}
	defer reader.Close()

	// Use pooled buffer to read config blob (reduces GC pressure)
	// Config blobs are typically small (< 16KB)
	buf := util.GetZeroCopyBuffer(16 * 1024)
	defer buf.Release()

	// Read config blob into pooled buffer
	var data []byte
	for {
		n, readErr := reader.Read(buf.Bytes())
		if n > 0 {
			data = append(data, buf.Bytes()[:n]...)
		}
		if readErr != nil {
			if readErr != io.EOF {
				return nil, errors.Wrap(readErr, "failed to read config blob data")
			}
			break
		}
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"digest":     digest,
		"size":       len(data),
	}).Debug("Successfully retrieved config blob from generic registry")

	return data, nil
}

// GetRemoteOptions returns remote options for registry operations
func (r *Repository) GetRemoteOptions() ([]remote.Option, error) {
	// Return the client's remote options
	if r.client != nil {
		return r.client.GetRemoteOptions(), nil
	}
	return []remote.Option{}, nil
}

// GetRepositoryName returns the repository name (alias for GetName)
func (r *Repository) GetRepositoryName() string {
	return r.GetName()
}

// PutManifest uploads a manifest to the repository
func (r *Repository) PutManifest(ctx context.Context, ref string, manifest *interfaces.Manifest) error {
	if ref == "" {
		return errors.InvalidInputf("reference cannot be empty")
	}

	if manifest == nil {
		return errors.InvalidInputf("manifest cannot be nil")
	}

	// Create a tag reference
	tagRef, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.Name(), ref))
	if err != nil {
		return errors.Wrap(err, "failed to create tag reference")
	}

	// Use remote.Put to upload the manifest
	// We need to create a minimal image implementation that returns our manifest
	img := &manifestImage{
		manifestBytes: manifest.Content,
		mediaType:     manifest.MediaType,
		digest:        manifest.Digest,
	}

	err = remote.Put(tagRef, img, r.client.GetRemoteOptions()...)
	if err != nil {
		return errors.Wrap(err, "failed to push manifest to registry")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"reference":  ref,
		"digest":     manifest.Digest,
	}).Info("Successfully pushed manifest to generic registry")

	return nil
}

// PushLayer uploads a blob/layer to the repository
func (r *Repository) PushLayer(ctx context.Context, digest string, data io.Reader) error {
	if digest == "" {
		return errors.InvalidInputf("digest cannot be empty")
	}

	if data == nil {
		return errors.InvalidInputf("data reader cannot be nil")
	}

	// For now, return not implemented as pushing individual layers
	// requires more complex handling with the blob upload API
	return errors.NotImplementedf("direct layer push not yet implemented for generic registries")
}

// ListTags lists all tags for this repository
func (r *Repository) ListTags(ctx context.Context) ([]string, error) {
	// Delegate to base repository if available
	if r.BaseRepository != nil {
		return r.BaseRepository.ListTags(ctx)
	}

	// Fallback implementation
	tags, err := remote.List(r.repository, r.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list tags from registry")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tagCount":   len(tags),
	}).Debug("Successfully listed tags from generic registry")

	return tags, nil
}

// manifestImage is a minimal implementation of v1.Image interface for pushing manifests
// It only implements the bare minimum required by remote.Put
type manifestImage struct {
	manifestBytes []byte
	mediaType     string
	digest        string
}

// Layers returns the ordered collection of filesystem layers
func (m *manifestImage) Layers() ([]v1.Layer, error) {
	return []v1.Layer{}, nil
}

// MediaType returns the media type of the manifest
func (m *manifestImage) MediaType() (types.MediaType, error) {
	if m.mediaType != "" {
		return types.MediaType(m.mediaType), nil
	}
	return types.MediaType("application/vnd.docker.distribution.manifest.v2+json"), nil
}

// Size returns the size of the manifest
func (m *manifestImage) Size() (int64, error) {
	return int64(len(m.manifestBytes)), nil
}

// ConfigName returns the hash of the config
func (m *manifestImage) ConfigName() (v1.Hash, error) {
	// Return a placeholder hash for the config
	return v1.NewHash("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
}

// ConfigFile returns the image's config file
func (m *manifestImage) ConfigFile() (*v1.ConfigFile, error) {
	return &v1.ConfigFile{
		Architecture: "amd64",
		OS:           "linux",
		RootFS: v1.RootFS{
			Type:    "layers",
			DiffIDs: []v1.Hash{},
		},
	}, nil
}

// RawConfigFile returns the serialized bytes of ConfigFile()
func (m *manifestImage) RawConfigFile() ([]byte, error) {
	return []byte("{}"), nil
}

// Digest returns the sha256 of the manifest
func (m *manifestImage) Digest() (v1.Hash, error) {
	if m.digest != "" {
		return v1.NewHash(m.digest)
	}
	// Calculate digest from manifest bytes - need to convert bytes to reader
	hash, _, err := v1.SHA256(bytes.NewReader(m.manifestBytes))
	return hash, err
}

// Manifest returns the image's Manifest object
func (m *manifestImage) Manifest() (*v1.Manifest, error) {
	// Parse the manifest bytes if possible
	// For now, return a minimal manifest
	configName, _ := m.ConfigName()
	return &v1.Manifest{
		SchemaVersion: 2,
		MediaType:     types.MediaType(m.mediaType),
		Config: v1.Descriptor{
			MediaType: types.MediaType("application/vnd.docker.container.image.v1+json"),
			Digest:    configName,
			Size:      2,
		},
		Layers: []v1.Descriptor{},
	}, nil
}

// RawManifest returns the serialized bytes of Manifest()
func (m *manifestImage) RawManifest() ([]byte, error) {
	return m.manifestBytes, nil
}

// LayerByDigest returns a Layer for interacting with a particular layer
func (m *manifestImage) LayerByDigest(h v1.Hash) (v1.Layer, error) {
	return nil, errors.NotFoundf("layer not found in manifest image")
}

// LayerByDiffID is an analog to LayerByDigest
func (m *manifestImage) LayerByDiffID(h v1.Hash) (v1.Layer, error) {
	return nil, errors.NotFoundf("layer not found in manifest image")
}
