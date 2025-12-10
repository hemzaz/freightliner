package gcr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/interfaces"

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
func (repo *Repository) GetName() string {
	return repo.name
}

// GetRepositoryName returns the name of the repository - implements common.Repository
func (repo *Repository) GetRepositoryName() string {
	return repo.name
}

// ListTags returns all tags for the repository - implements common.Repository
func (repo *Repository) ListTags(ctx context.Context) ([]string, error) {
	var tags []string

	// In a real implementation, this would use google.List or the GCR API
	// For now, using a simulated implementation
	registry := fmt.Sprintf("gcr.io/%s", repo.client.project)
	repoName := fmt.Sprintf("%s/%s", registry, repo.name)

	// Create a full repository reference
	repoRef, err := name.NewRepository(repoName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Get tags
	gTags, err := google.List(repoRef, repo.client.googleAuthOpts...)
	if err != nil {
		// Handle 404 error specifically
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil, errors.NotFoundf("repository %s not found", repo.name)
		}
		return nil, errors.Wrap(err, "failed to list tags")
	}

	// Extract tag names
	for _, info := range gTags.Manifests {
		tags = append(tags, info.Tags...)
	}

	return tags, nil
}

// GetImage retrieves an image by tag
func (repo *Repository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	if tag == "" {
		return nil, errors.InvalidInputf("tag cannot be empty")
	}

	// Create a tagged reference
	taggedRef, err := name.NewTag(fmt.Sprintf("%s:%s", repo.repository.String(), tag))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tag reference")
	}

	// Get the image
	img, err := remote.Image(taggedRef, repo.client.transportOpt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil, errors.NotFoundf("image %s:%s not found", repo.name, tag)
		}
		return nil, errors.Wrap(err, "failed to get image")
	}

	return img, nil
}

// GetManifest retrieves a manifest by tag - implements interfaces.Repository
func (repo *Repository) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	if tag == "" {
		return nil, errors.InvalidInputf("tag cannot be empty")
	}

	// Create a tagged reference
	taggedRef, err := name.NewTag(fmt.Sprintf("%s:%s", repo.repository.String(), tag))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tag reference")
	}

	// Get the descriptor
	desc, err := remote.Get(taggedRef, repo.client.transportOpt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil, errors.NotFoundf("image %s:%s not found", repo.name, tag)
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
	manifest := &interfaces.Manifest{
		Content:   rawManifest,
		MediaType: string(desc.MediaType),
		Digest:    digest.String(),
	}

	return manifest, nil
}

// GetMediaType returns the media type of the manifest
func (repo *Repository) GetMediaType(ctx context.Context, tag string) (types.MediaType, error) {
	if tag == "" {
		return "", errors.InvalidInputf("tag cannot be empty")
	}

	// Create a tagged reference
	taggedRef, err := name.NewTag(fmt.Sprintf("%s:%s", repo.repository.String(), tag))
	if err != nil {
		return "", errors.Wrap(err, "failed to create tag reference")
	}

	// Get the descriptor
	desc, err := remote.Get(taggedRef, repo.client.transportOpt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return "", errors.NotFoundf("image %s:%s not found", repo.name, tag)
		}
		return "", errors.Wrap(err, "failed to get image descriptor")
	}

	return desc.MediaType, nil
}

// PutImage uploads an image with the given tag
func (repo *Repository) PutImage(ctx context.Context, tag string, img v1.Image) error {
	if tag == "" {
		return errors.InvalidInputf("tag cannot be empty")
	}

	if img == nil {
		return errors.InvalidInputf("image cannot be nil")
	}

	// Create a tagged reference
	taggedRef, err := name.NewTag(fmt.Sprintf("%s:%s", repo.repository.String(), tag))
	if err != nil {
		return errors.Wrap(err, "failed to create tag reference")
	}

	// Push the image
	if err := remote.Write(taggedRef, img, repo.client.transportOpt); err != nil {
		return errors.Wrap(err, "failed to write image")
	}

	return nil
}

// PutLayer uploads a layer to the repository
func (repo *Repository) PutLayer(ctx context.Context, layer v1.Layer) error {
	if layer == nil {
		return errors.InvalidInputf("layer cannot be nil")
	}

	// Get the layer digest (unused but validate it exists)
	_, err := layer.Digest()
	if err != nil {
		return errors.Wrap(err, "failed to get layer digest")
	}

	// Upload the layer
	if err := remote.WriteLayer(repo.repository, layer, repo.client.transportOpt); err != nil {
		return errors.Wrap(err, "failed to write layer")
	}

	return nil
}

// GetLayerReader gets a reader for a layer by digest - implements common.Repository
func (repo *Repository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	if digest == "" {
		return nil, errors.InvalidInputf("digest cannot be empty")
	}

	// Create a digest reference for the layer
	digestRef := repo.repository.Digest(digest)

	// Get the layer
	layer, err := remote.Layer(digestRef, repo.client.transportOpt)
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
func (repo *Repository) DeleteImage(ctx context.Context, tag string) error {
	if tag == "" {
		return errors.InvalidInputf("tag cannot be empty")
	}

	// First, get the manifest to extract the digest
	manifest, err := repo.GetManifest(ctx, tag)
	if err != nil {
		return errors.Wrap(err, "failed to get manifest for deletion")
	}

	// Google Container Registry uses the Artifact Registry API for deletion operations
	// If AR client is not available, try HTTP-based approach as fallback
	if repo.client.arClient != nil {
		// Construct the resource name
		// Format: projects/{project}/locations/{location}/repositories/{repository}/packages/{package}/versions/{version}
		location := repo.client.location
		if location == "us" || location == "eu" || location == "asia" {
			location = "us-central1" // Map legacy locations to GCP regions
		}

		digestRef := strings.TrimPrefix(manifest.Digest, "sha256:")
		resourceName := fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s/versions/%s",
			repo.client.project, location, repo.name, repo.name, digestRef)

		// Delete the version using Artifact Registry API
		deleteReq := repo.client.arClient.Projects.Locations.Repositories.Packages.Versions.Delete(resourceName)
		resp, deleteErr := deleteReq.Context(ctx).Do()
		if deleteErr != nil {
			// Check specific error messages to provide better diagnostics
			if strings.Contains(deleteErr.Error(), "404") {
				return errors.NotFoundf("image %s:%s not found or already deleted", repo.name, tag)
			}
			return errors.Wrap(deleteErr, "failed to delete image via Artifact Registry API")
		}

		// Check response
		if resp.HTTPStatusCode != http.StatusOK && resp.HTTPStatusCode != http.StatusAccepted {
			return errors.InvalidInputf("failed to delete image, status: %d", resp.HTTPStatusCode)
		}

		return nil
	}

	// Fallback approach using gcrane or container registry HTTP API
	// Create a reference for the image digest
	digestRef, err := name.NewDigest(fmt.Sprintf("%s@%s", repo.repository.String(), manifest.Digest))
	if err != nil {
		return errors.Wrap(err, "failed to create digest reference")
	}

	// Use HTTP DELETE request to the GCR registry API
	transport, err := repo.client.GetTransport(repo.name)
	if err != nil {
		return errors.Wrap(err, "failed to get authenticated transport")
	}

	// Create authenticated HTTP client
	client := &http.Client{
		Transport: transport,
	}

	// GCR supports DELETE operations on manifest endpoints
	// URL format: https://gcr.io/v2/{repository}/manifests/{digest}
	deleteURL := fmt.Sprintf("https://%s/v2/%s/%s/manifests/%s",
		digestRef.Context().RegistryStr(),
		repo.client.project,
		repo.name,
		strings.TrimPrefix(manifest.Digest, "sha256:"))

	// Create a DELETE request
	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create delete request")
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send delete request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check response
	if resp.StatusCode == http.StatusNotFound {
		return errors.NotFoundf("image %s:%s not found or already deleted", repo.name, tag)
	} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return errors.InvalidInputf("failed to delete image, status: %d, response: %s",
			resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeleteManifest deletes the manifest for the given tag - implements common.Repository
func (repo *Repository) DeleteManifest(ctx context.Context, tag string) error {
	// This is a wrapper around DeleteImage to match the common.Repository interface
	return repo.DeleteImage(ctx, tag)
}

// PutManifest uploads a manifest with the given tag - implements interfaces.Repository
func (repo *Repository) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	if tag == "" {
		return errors.InvalidInputf("tag cannot be empty")
	}

	if manifest == nil {
		return errors.InvalidInputf("manifest cannot be nil")
	}

	// Create a tag reference
	ref, err := name.NewTag(fmt.Sprintf("%s:%s", repo.repository.String(), tag))
	if err != nil {
		return errors.Wrap(err, "failed to create tag reference")
	}

	// Use the go-containerregistry library to push the manifest
	err = remote.Put(ref, mockRemoteImage{
		manifestBytes: manifest.Content,
		mediaType:     types.MediaType(manifest.MediaType),
	}, repo.client.transportOpt)

	if err != nil {
		return errors.Wrap(err, "failed to push manifest")
	}

	return nil
}

// GetRepositoryReference returns the name.Repository reference
func (repo *Repository) GetRepositoryReference() name.Repository {
	return repo.repository
}

// GetImageReference returns a name.Reference for the given tag - implements common.Repository
func (repo *Repository) GetImageReference(tag string) (name.Reference, error) {
	if tag == "" {
		return nil, errors.InvalidInputf("tag cannot be empty")
	}

	// Check if it's already a digest
	if strings.Contains(tag, "@") {
		return name.NewDigest(fmt.Sprintf("%s@%s", repo.repository.String(), tag))
	}

	// Create a tag reference
	return name.NewTag(fmt.Sprintf("%s:%s", repo.repository.String(), tag))
}

// GetRemoteOptions returns the remote options for this repository - implements common.Repository
func (repo *Repository) GetRemoteOptions() ([]remote.Option, error) {
	return []remote.Option{repo.client.transportOpt}, nil
}

// mockRemoteImage is a complete implementation of the v1.Image interface for pushing manifests
// It's primarily used for manifest operations and doesn't need to implement all
// image operations fully, but it should return reasonable values for all methods
type mockRemoteImage struct {
	manifestBytes []byte
	mediaType     types.MediaType
	// Cached values for performance
	manifest   *v1.Manifest
	configFile *v1.ConfigFile
	layers     []v1.Layer
	digest     v1.Hash
}

// Layers returns the ordered collection of filesystem layers that comprise this image
func (mockImg mockRemoteImage) Layers() ([]v1.Layer, error) {
	// If we've already parsed the manifest and have layers, return them
	if mockImg.manifest != nil && mockImg.layers != nil {
		return mockImg.layers, nil
	}

	// If not, return an empty layer slice, which is valid for schema1 manifests
	return []v1.Layer{}, nil
}

// MediaType returns the media type of the image's manifest
func (mockImg mockRemoteImage) MediaType() (types.MediaType, error) {
	return mockImg.mediaType, nil
}

// Size returns the size of the manifest
func (mockImg mockRemoteImage) Size() (int64, error) {
	return int64(len(mockImg.manifestBytes)), nil
}

// ConfigName returns the hash of the image's config file (Image ID)
func (mockImg mockRemoteImage) ConfigName() (v1.Hash, error) {
	// If we have a manifest, return the config digest
	manifest, err := mockImg.Manifest()
	if err == nil && manifest != nil {
		return manifest.Config.Digest, nil
	}

	// For OCI and Docker manifests, this is required
	// Return an empty hash as we don't have a real config
	emptyHash, err := v1.NewHash("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	if err != nil {
		return v1.Hash{}, err
	}
	return emptyHash, nil
}

// ConfigFile returns the image's config file
func (mockImg mockRemoteImage) ConfigFile() (*v1.ConfigFile, error) {
	// If we have a cached config file, return it
	if mockImg.configFile != nil {
		return mockImg.configFile, nil
	}

	// Create a minimal valid config file
	configFile := &v1.ConfigFile{
		Architecture: "amd64",
		OS:           "linux",
		RootFS: v1.RootFS{
			Type:    "layers",
			DiffIDs: []v1.Hash{},
		},
		History: []v1.History{},
		// Other fields left at defaults
	}

	mockImg.configFile = configFile
	return configFile, nil
}

// RawConfigFile returns the serialized bytes of ConfigFile()
func (mockImg mockRemoteImage) RawConfigFile() ([]byte, error) {
	configFile, err := mockImg.ConfigFile()
	if err != nil {
		return nil, err
	}

	// Marshal the config to JSON
	configBytes, err := json.Marshal(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal config file")
	}

	return configBytes, nil
}

// Digest returns the sha256 of this image's manifest
func (mockImg mockRemoteImage) Digest() (v1.Hash, error) {
	// If we have a cached digest, return it
	if (mockImg.digest != v1.Hash{}) {
		return mockImg.digest, nil
	}

	// Calculate the digest of the manifest bytes
	bytes := mockImg.manifestBytes
	if len(bytes) == 0 {
		// If the manifest is empty, return the empty string hash
		emptyHash, err := v1.NewHash("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
		if err != nil {
			return v1.Hash{}, err
		}
		mockImg.digest = emptyHash
		return mockImg.digest, nil
	}

	// Otherwise, compute the digest
	hash, _, err := v1.SHA256(bytes2Reader(bytes))
	if err != nil {
		return v1.Hash{}, err
	}

	mockImg.digest = hash
	return hash, nil
}

// Manifest returns this image's Manifest object
func (mockImg mockRemoteImage) Manifest() (*v1.Manifest, error) {
	// If we have a cached manifest, return it
	if mockImg.manifest != nil {
		return mockImg.manifest, nil
	}

	// If we have manifest bytes, try to parse them
	if len(mockImg.manifestBytes) > 0 {
		manifest := &v1.Manifest{}
		if err := json.Unmarshal(mockImg.manifestBytes, manifest); err == nil {
			mockImg.manifest = manifest
			return manifest, nil
		}

		// If we can't parse them, create a minimal valid manifest
		configDigest, err := mockImg.ConfigName()
		if err != nil {
			return nil, err
		}

		// Create a minimal valid manifest
		manifest = &v1.Manifest{
			SchemaVersion: 2,
			MediaType:     mockImg.mediaType,
			Config: v1.Descriptor{
				MediaType: types.DockerConfigJSON,
				Digest:    configDigest,
				Size:      0,
			},
			Layers: []v1.Descriptor{},
		}

		mockImg.manifest = manifest
		return manifest, nil
	}

	// Create a minimal valid manifest
	configDigest, err := mockImg.ConfigName()
	if err != nil {
		return nil, err
	}

	// Create a minimal valid manifest
	manifest := &v1.Manifest{
		SchemaVersion: 2,
		MediaType:     mockImg.mediaType,
		Config: v1.Descriptor{
			MediaType: types.DockerConfigJSON,
			Digest:    configDigest,
			Size:      0,
		},
		Layers: []v1.Descriptor{},
	}

	mockImg.manifest = manifest
	return manifest, nil
}

// RawManifest returns the serialized bytes of Manifest()
func (mockImg mockRemoteImage) RawManifest() ([]byte, error) {
	// If we have existing bytes, return them
	if len(mockImg.manifestBytes) > 0 {
		return mockImg.manifestBytes, nil
	}

	// Otherwise, build a manifest from our data
	manifest, err := mockImg.Manifest()
	if err != nil {
		return nil, err
	}

	// Marshal the manifest to JSON
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal manifest")
	}

	return manifestBytes, nil
}

// LayerByDigest returns a Layer for interacting with a particular layer
func (mockImg mockRemoteImage) LayerByDigest(h v1.Hash) (v1.Layer, error) {
	// We don't have real layers, so we can't return one by digest
	return nil, errors.NotFoundf("layer with digest %s not found in mock image", h)
}

// LayerByDiffID is an analog to LayerByDigest, looking up by uncompressed hash
func (mockImg mockRemoteImage) LayerByDiffID(h v1.Hash) (v1.Layer, error) {
	// We don't have real layers, so we can't return one by diff ID
	return nil, errors.NotFoundf("layer with diff ID %s not found in mock image", h)
}

// bytes2Reader converts a byte slice to an io.Reader
func bytes2Reader(b []byte) io.Reader {
	return bytes.NewReader(b)
}
