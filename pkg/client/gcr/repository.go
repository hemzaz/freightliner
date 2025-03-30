package gcr

import (
	"context"
	"fmt"
	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/errors"
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// Repository implements the repository interface for Google GCR
type Repository struct {
	client     *Client
	name       string          // Repository name (without registry prefix)
	repository name.Repository // Full repository reference

	// Test-specific fields
	ref        name.Repository                                                                         // Used in tests
	registry   name.Registry                                                                           // Used in tests
	keychain   authn.Keychain                                                                          // Used in tests
	tagsFunc   func(ctx context.Context, repo name.Repository, opt ...google.Option) ([]string, error) // Used in tests
	remoteFunc func(ref name.Reference, options ...remote.Option) (*remote.Descriptor, error)          // Used in tests
}

// GetName returns the repository name - internal method
func (r *Repository) GetName() string {
	return r.name
}

// GetRepositoryName returns the name of the repository - implements common.Repository
func (r *Repository) GetRepositoryName() string {
	return r.name
}

// ListTags returns all tags for the repository - implements common.Repository
func (r *Repository) ListTags() ([]string, error) {
	var tags []string

	// In a real implementation, this would use google.List or the GCR API
	// For now, using a simulated implementation
	registry := fmt.Sprintf("gcr.io/%s", r.client.project)
	repoName := fmt.Sprintf("%s/%s", registry, r.name)

	// Create a full repository reference
	repo, err := name.NewRepository(repoName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Get tags
	gTags, err := google.List(repo, r.client.googleAuthOpts...)
	if err != nil {
		// Handle 404 error specifically
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil, errors.NotFoundf("repository %s not found", r.name)
		}
		return nil, errors.Wrap(err, "failed to list tags")
	}

	// Extract tag names
	for _, info := range gTags.Manifests {
		for _, tag := range info.Tags {
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// GetImage retrieves an image by tag
func (r *Repository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	if tag == "" {
		return nil, errors.InvalidInputf("tag cannot be empty")
	}

	// Create a tagged reference
	taggedRef, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.String(), tag))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tag reference")
	}

	// Get the image
	img, err := remote.Image(taggedRef, r.client.transportOpt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil, errors.NotFoundf("image %s:%s not found", r.name, tag)
		}
		return nil, errors.Wrap(err, "failed to get image")
	}

	return img, nil
}

// GetManifest retrieves a manifest by tag - implements common.Repository
func (r *Repository) GetManifest(ctx context.Context, tag string) (*common.Manifest, error) {
	if tag == "" {
		return nil, errors.InvalidInputf("tag cannot be empty")
	}

	// Create a tagged reference
	taggedRef, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.String(), tag))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tag reference")
	}

	// Get the descriptor
	desc, err := remote.Get(taggedRef, r.client.transportOpt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil, errors.NotFoundf("image %s:%s not found", r.name, tag)
		}
		return nil, errors.Wrap(err, "failed to get image descriptor")
	}

	// Get raw manifest
	rawManifest, err := desc.RawManifest()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get raw manifest")
	}

	// Get digest
	digest := desc.Digest

	// Create the manifest
	manifest := &common.Manifest{
		Content:   rawManifest,
		MediaType: string(desc.MediaType),
		Digest:    digest.String(),
	}

	return manifest, nil
}

// GetMediaType returns the media type of the manifest
func (r *Repository) GetMediaType(ctx context.Context, tag string) (types.MediaType, error) {
	if tag == "" {
		return "", errors.InvalidInputf("tag cannot be empty")
	}

	// Create a tagged reference
	taggedRef, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.String(), tag))
	if err != nil {
		return "", errors.Wrap(err, "failed to create tag reference")
	}

	// Get the descriptor
	desc, err := remote.Get(taggedRef, r.client.transportOpt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return "", errors.NotFoundf("image %s:%s not found", r.name, tag)
		}
		return "", errors.Wrap(err, "failed to get image descriptor")
	}

	return desc.MediaType, nil
}

// PutImage uploads an image with the given tag
func (r *Repository) PutImage(ctx context.Context, tag string, img v1.Image) error {
	if tag == "" {
		return errors.InvalidInputf("tag cannot be empty")
	}

	if img == nil {
		return errors.InvalidInputf("image cannot be nil")
	}

	// Create a tagged reference
	taggedRef, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.String(), tag))
	if err != nil {
		return errors.Wrap(err, "failed to create tag reference")
	}

	// Push the image
	if err := remote.Write(taggedRef, img, r.client.transportOpt); err != nil {
		return errors.Wrap(err, "failed to write image")
	}

	return nil
}

// PutLayer uploads a layer to the repository
func (r *Repository) PutLayer(ctx context.Context, layer v1.Layer) error {
	if layer == nil {
		return errors.InvalidInputf("layer cannot be nil")
	}

	// Get the layer digest (unused but validate it exists)
	_, err := layer.Digest()
	if err != nil {
		return errors.Wrap(err, "failed to get layer digest")
	}

	// Upload the layer
	if err := remote.WriteLayer(r.repository, layer, r.client.transportOpt); err != nil {
		return errors.Wrap(err, "failed to write layer")
	}

	return nil
}

// GetLayerReader gets a reader for a layer by digest - implements common.Repository
func (r *Repository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	if digest == "" {
		return nil, errors.InvalidInputf("digest cannot be empty")
	}

	// Create a digest reference for the layer
	digestRef := r.repository.Digest(digest)

	// Get the layer
	layer, err := remote.Layer(digestRef, r.client.transportOpt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil, errors.NotFoundf("layer %s not found", digest)
		}
		return nil, errors.Wrap(err, "failed to get layer")
	}

	// Get the compressed reader
	reader, err := layer.Compressed()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get layer reader")
	}

	return reader, nil
}

// DeleteImage deletes an image by tag
func (r *Repository) DeleteImage(ctx context.Context, tag string) error {
	// Currently, the go-containerregistry library doesn't have a direct way to delete an image
	// In a real implementation, this would use the GCR API directly
	// For now, we'll return an error
	return errors.NotImplementedf("image deletion not implemented for GCR")
}

// DeleteManifest deletes the manifest for the given tag - implements common.Repository
func (r *Repository) DeleteManifest(ctx context.Context, tag string) error {
	// This is a wrapper around DeleteImage to match the common.Repository interface
	return r.DeleteImage(ctx, tag)
}

// PutManifest uploads a manifest with the given tag - implements common.Repository
func (r *Repository) PutManifest(ctx context.Context, tag string, manifest *common.Manifest) error {
	if tag == "" {
		return errors.InvalidInputf("tag cannot be empty")
	}

	if manifest == nil {
		return errors.InvalidInputf("manifest cannot be nil")
	}

	// Create a tag reference
	ref, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.String(), tag))
	if err != nil {
		return errors.Wrap(err, "failed to create tag reference")
	}

	// Use the go-containerregistry library to push the manifest
	err = remote.Put(ref, mockRemoteImage{
		manifestBytes: manifest.Content,
		mediaType:     types.MediaType(manifest.MediaType),
	}, r.client.transportOpt)

	if err != nil {
		return errors.Wrap(err, "failed to push manifest")
	}

	return nil
}

// GetRepositoryReference returns the name.Repository reference
func (r *Repository) GetRepositoryReference() name.Repository {
	return r.repository
}

// GetImageReference returns a name.Reference for the given tag - implements common.Repository
func (r *Repository) GetImageReference(tag string) (name.Reference, error) {
	if tag == "" {
		return nil, errors.InvalidInputf("tag cannot be empty")
	}

	// Check if it's already a digest
	if strings.Contains(tag, "@") {
		return name.NewDigest(fmt.Sprintf("%s@%s", r.repository.String(), tag))
	}

	// Create a tag reference
	return name.NewTag(fmt.Sprintf("%s:%s", r.repository.String(), tag))
}

// GetRemoteOptions returns the remote options for this repository - implements common.Repository
func (r *Repository) GetRemoteOptions() ([]remote.Option, error) {
	return []remote.Option{r.client.transportOpt}, nil
}

// mockRemoteImage is a stub implementation of the v1.Image interface for pushing manifests
type mockRemoteImage struct {
	manifestBytes []byte
	mediaType     types.MediaType
}

// Layers returns the layers of the image
func (m mockRemoteImage) Layers() ([]v1.Layer, error) {
	return nil, nil
}

// MediaType returns the media type of the image
func (m mockRemoteImage) MediaType() (types.MediaType, error) {
	return m.mediaType, nil
}

// Size returns the size of the image
func (m mockRemoteImage) Size() (int64, error) {
	return int64(len(m.manifestBytes)), nil
}

// ConfigName returns the hash of the image config
func (m mockRemoteImage) ConfigName() (v1.Hash, error) {
	return v1.Hash{}, nil
}

// ConfigFile returns the image config file
func (m mockRemoteImage) ConfigFile() (*v1.ConfigFile, error) {
	return nil, nil
}

// RawConfigFile returns the raw image config
func (m mockRemoteImage) RawConfigFile() ([]byte, error) {
	return nil, nil
}

// Digest returns the digest of the image
func (m mockRemoteImage) Digest() (v1.Hash, error) {
	return v1.Hash{}, nil
}

// Manifest returns the manifest of the image
func (m mockRemoteImage) Manifest() (*v1.Manifest, error) {
	return nil, nil
}

// RawManifest returns the raw manifest of the image
func (m mockRemoteImage) RawManifest() ([]byte, error) {
	return m.manifestBytes, nil
}

// LayerByDigest returns a layer by digest
func (m mockRemoteImage) LayerByDigest(v1.Hash) (v1.Layer, error) {
	return nil, nil
}

// LayerByDiffID returns a layer by diff ID
func (m mockRemoteImage) LayerByDiffID(v1.Hash) (v1.Layer, error) {
	return nil, nil
}
