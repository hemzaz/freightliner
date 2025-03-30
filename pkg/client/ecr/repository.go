package ecr

import (
	"context"
	"fmt"
	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/errors"
	"io"
	"strings"

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
func (r *Repository) GetName() string {
	return r.name
}

// GetRepositoryName returns the name of the repository - implements common.Repository
func (r *Repository) GetRepositoryName() string {
	return r.name
}

// ListTags returns all tags for the repository - implements common.Repository
func (r *Repository) ListTags() ([]string, error) {
	ctx := context.Background()
	var tags []string
	var nextToken *string

	for {
		input := &awsecr.ListImagesInput{
			RepositoryName: aws.String(r.name),
			NextToken:      nextToken,
		}

		// Apply account ID if specified
		if r.client.accountID != "" {
			input.RegistryId = aws.String(r.client.accountID)
		}

		resp, err := r.client.ecr.ListImages(ctx, input)
		if err != nil {
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

// GetManifest returns the manifest for the given tag - implements common.Repository
func (r *Repository) GetManifest(ctx context.Context, tag string) (*common.Manifest, error) {
	if tag == "" {
		return nil, errors.InvalidInputf("tag cannot be empty")
	}

	// Create a reference for the tag
	ref, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.String(), tag))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tag reference")
	}

	// Get the image from the registry
	desc, err := remote.Get(ref, r.client.GetRemoteOptions()...)
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

	manifest := &common.Manifest{
		Content:   manifestBytes,
		MediaType: string(desc.MediaType),
		Digest:    digest.String(),
	}

	return manifest, nil
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
	}, r.client.GetRemoteOptions()...)

	if err != nil {
		return errors.Wrap(err, "failed to push manifest")
	}

	return nil
}

// DeleteManifest deletes a manifest with the given tag - implements common.Repository
func (r *Repository) DeleteManifest(ctx context.Context, tag string) error {
	if tag == "" {
		return errors.InvalidInputf("tag cannot be empty")
	}

	// First, we need to get the image digest for the tag
	input := &awsecr.BatchGetImageInput{
		RepositoryName: aws.String(r.name),
		ImageIds: []ecrtypes.ImageIdentifier{
			{
				ImageTag: aws.String(tag),
			},
		},
	}

	// Apply account ID if specified
	if r.client.accountID != "" {
		input.RegistryId = aws.String(r.client.accountID)
	}

	// Get the image details
	resp, err := r.client.ecr.BatchGetImage(ctx, input)
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
		RepositoryName: aws.String(r.name),
		ImageIds: []ecrtypes.ImageIdentifier{
			{
				ImageDigest: imageDigest,
			},
		},
	}

	// Apply account ID if specified
	if r.client.accountID != "" {
		deleteInput.RegistryId = aws.String(r.client.accountID)
	}

	// Delete the image
	_, err = r.client.ecr.BatchDeleteImage(ctx, deleteInput)
	if err != nil {
		return errors.Wrap(err, "failed to delete image")
	}

	return nil
}

// GetLayerReader returns a reader for a layer with the given digest - implements common.Repository
func (r *Repository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	if digest == "" {
		return nil, errors.InvalidInputf("digest cannot be empty")
	}

	// Create a digest reference
	ref, err := name.NewDigest(fmt.Sprintf("%s@%s", r.repository.String(), digest))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digest reference")
	}

	// Get the layer
	layer, err := remote.Layer(ref, r.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get layer")
	}

	// Get the layer content
	return layer.Compressed()
}

// GetImageReference returns a name.Reference for a tag - implements common.Repository
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
	return r.client.GetRemoteOptions(), nil
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
