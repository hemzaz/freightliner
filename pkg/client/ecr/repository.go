package ecr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/interfaces"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// Repository implements the common.Repository interface for AWS ECR
type Repository struct {
	client     *Client
	name       string
	repository name.Repository
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
	var nextToken *string

	for {
		input := &awsecr.ListImagesInput{
			RepositoryName: aws.String(repo.name),
			NextToken:      nextToken,
		}

		// Apply account ID if specified
		if repo.client.accountID != "" {
			input.RegistryId = aws.String(repo.client.accountID)
		}

		// Debug logging
		repo.client.logger.WithFields(map[string]interface{}{
			"repository_name": repo.name,
			"registry_id":     repo.client.accountID,
			"has_next_token":  nextToken != nil,
		}).Debug("Calling ECR ListImages API")

		resp, err := repo.client.ecr.ListImages(ctx, input)
		if err != nil {
			repo.client.logger.WithFields(map[string]interface{}{
				"repository_name": repo.name,
				"registry_id":     repo.client.accountID,
				"error":           err.Error(),
			}).Warn("ECR ListImages API failed")
			return nil, errors.Wrap(err, "failed to list images")
		}

		// Extract tags from the response
		for _, img := range resp.ImageIds {
			if img.ImageTag != nil {
				tags = append(tags, *img.ImageTag)
			}
		}

		// Check for more pages
		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}

	return tags, nil
}

// GetImage retrieves an image by tag - implements common.Repository
func (repo *Repository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	if tag == "" {
		return nil, errors.InvalidInputf("tag cannot be empty")
	}

	// Create a reference for the tag
	ref, err := name.NewTag(fmt.Sprintf("%s:%s", repo.repository.String(), tag))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tag reference")
	}

	// Get the image from the registry
	img, err := remote.Image(ref, repo.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get image from registry")
	}

	return img, nil
}

// GetManifest returns the manifest for the given tag - implements interfaces.Repository
func (repo *Repository) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	if tag == "" {
		return nil, errors.InvalidInputf("tag cannot be empty")
	}

	// Create a reference for the tag
	ref, err := name.NewTag(fmt.Sprintf("%s:%s", repo.repository.String(), tag))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tag reference")
	}

	// Get the image from the registry
	desc, err := remote.Get(ref, repo.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get image from registry")
	}

	// Get the raw manifest
	manifestBytes, err := desc.RawManifest()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get raw manifest")
	}

	// Get the digest
	digest := desc.Digest

	manifest := &interfaces.Manifest{
		Content:   manifestBytes,
		MediaType: string(desc.MediaType),
		Digest:    digest.String(),
	}

	return manifest, nil
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
	}, repo.client.GetRemoteOptions()...)

	if err != nil {
		return errors.Wrap(err, "failed to push manifest")
	}

	return nil
}

// DeleteManifest deletes a manifest with the given tag - implements common.Repository
func (repo *Repository) DeleteManifest(ctx context.Context, tag string) error {
	if tag == "" {
		return errors.InvalidInputf("tag cannot be empty")
	}

	// First, we need to get the image digest for the tag
	input := &awsecr.BatchGetImageInput{
		RepositoryName: aws.String(repo.name),
		ImageIds: []ecrtypes.ImageIdentifier{
			{
				ImageTag: aws.String(tag),
			},
		},
	}

	// Apply account ID if specified
	if repo.client.accountID != "" {
		input.RegistryId = aws.String(repo.client.accountID)
	}

	// Get the image details
	resp, err := repo.client.ecr.BatchGetImage(ctx, input)
	if err != nil {
		return errors.Wrap(err, "failed to get image details")
	}

	if len(resp.Images) == 0 {
		return errors.NotFoundf("image not found: %s", tag)
	}

	// Get the image digest
	imageDigest := resp.Images[0].ImageId.ImageDigest
	if imageDigest == nil {
		return errors.InvalidInputf("image digest is nil")
	}

	// Delete the image by digest
	deleteInput := &awsecr.BatchDeleteImageInput{
		RepositoryName: aws.String(repo.name),
		ImageIds: []ecrtypes.ImageIdentifier{
			{
				ImageDigest: imageDigest,
			},
		},
	}

	// Apply account ID if specified
	if repo.client.accountID != "" {
		deleteInput.RegistryId = aws.String(repo.client.accountID)
	}

	// Delete the image
	_, err = repo.client.ecr.BatchDeleteImage(ctx, deleteInput)
	if err != nil {
		return errors.Wrap(err, "failed to delete image")
	}

	return nil
}

// GetLayerReader returns a reader for a layer with the given digest - implements common.Repository
func (repo *Repository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	if digest == "" {
		return nil, errors.InvalidInputf("digest cannot be empty")
	}

	// Create a digest reference
	ref, err := name.NewDigest(fmt.Sprintf("%s@%s", repo.repository.String(), digest))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digest reference")
	}

	// Get the layer
	layer, err := remote.Layer(ref, repo.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get layer")
	}

	// Get the layer content
	return layer.Compressed()
}

// GetImageReference returns a name.Reference for a tag - implements common.Repository
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
	return repo.client.GetRemoteOptions(), nil
}

// PutImage uploads an image with the given tag - implements interfaces.Repository
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

	// Push the image using go-containerregistry
	if err := remote.Write(taggedRef, img, repo.client.GetRemoteOptions()...); err != nil {
		return errors.Wrap(err, "failed to write image to ECR")
	}

	return nil
}

// mockRemoteImage is a complete implementation of the v1.Image interface for pushing manifests
// It's primarily used for manifest operations and doesn't need to implement all
// image operations fully, but it should return reasonable values for all methods
type mockRemoteImage struct {
	manifestBytes []byte
	mediaType     types.MediaType
	// Cached values
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

	// Cache the config file
	_ = configFile
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

	// Cache the digest
	_ = hash
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
