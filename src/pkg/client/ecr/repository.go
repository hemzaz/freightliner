package ecr

import (
	"context"
	"fmt"
	"io"
	"strings"

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

// GetRepositoryName returns the repository name
func (r *Repository) GetRepositoryName() string {
	return r.name
}

// ListTags returns all tags for the repository
func (r *Repository) ListTags() ([]string, error) {
	// Context for the request
	ctx := context.Background()
	var tags []string
	var nextToken *string

	// Repository name in the format registry.amazonaws.com/repo
	fullName := r.repository.String()
	// Extract just the repository name without the registry prefix
	parts := strings.Split(fullName, "/")
	repoName := parts[len(parts)-1]

	// Get repository name without registry prefix
	input := &awsecr.DescribeImagesInput{
		RepositoryName: &repoName,
	}

	// Use the ECR client to list images in the repository
	for {
		// Set the pagination token if we have one
		if nextToken != nil {
			input.NextToken = nextToken
		}

		// Make the request
		result, err := r.client.ecr.DescribeImages(ctx, input)

		if err != nil {
			// If the repository doesn't exist or we don't have access, return an empty list
			if strings.Contains(err.Error(), "RepositoryNotFoundException") {
				return []string{}, nil
			}
			return nil, fmt.Errorf("failed to list tags: %w", err)
		}

		// Extract tags from each image
		for _, image := range result.ImageDetails {
			for _, tag := range image.ImageTags {
				tags = append(tags, string(tag))
			}
		}

		// Check if there are more results
		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}

	return tags, nil
}

// GetManifest returns the manifest for the given tag or digest
func (r *Repository) GetManifest(tag string) ([]byte, string, error) {
	// Create a reference to the image with the given tag or digest
	var ref name.Reference
	var err error

	if strings.HasPrefix(tag, "@") {
		// This is a digest reference
		digest := strings.TrimPrefix(tag, "@")
		ref, err = name.NewDigest(fmt.Sprintf("%s@%s", r.repository.String(), digest))
	} else {
		// This is a tag reference
		ref, err = name.NewTag(fmt.Sprintf("%s:%s", r.repository.String(), tag))
	}

	if err != nil {
		return nil, "", fmt.Errorf("failed to create reference: %w", err)
	}

	// Get the image from the registry
	desc, err := remote.Get(ref, r.client.transportOpt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get image: %w", err)
	}

	// Get the raw manifest
	rawManifest, err := desc.RawManifest()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get raw manifest: %w", err)
	}

	// Get the media type from the descriptor
	mediaType := desc.MediaType
	if mediaType == "" {
		// If we can't determine media type, default to Docker Manifest Schema 2
		mediaType = types.DockerManifestSchema2
	}

	return rawManifest, string(mediaType), nil
}

// PutManifest uploads a manifest with the given tag
func (r *Repository) PutManifest(tag string, manifestBytes []byte, mediaTypeStr string) error {
	// Create a reference to the image with the given tag or digest
	var ref name.Reference
	var err error

	if strings.HasPrefix(tag, "@") {
		// This is a digest reference
		digest := strings.TrimPrefix(tag, "@")
		ref, err = name.NewDigest(fmt.Sprintf("%s@%s", r.repository.String(), digest))
	} else {
		// This is a tag reference
		ref, err = name.NewTag(fmt.Sprintf("%s:%s", r.repository.String(), tag))
	}

	if err != nil {
		return fmt.Errorf("failed to create reference: %w", err)
	}

	// Parse the media type
	var mediaType types.MediaType
	switch mediaTypeStr {
	case string(types.DockerManifestSchema1):
		mediaType = types.DockerManifestSchema1
	case string(types.DockerManifestSchema1Signed):
		mediaType = types.DockerManifestSchema1Signed
	case string(types.DockerManifestSchema2):
		mediaType = types.DockerManifestSchema2
	case string(types.OCIManifestSchema1):
		mediaType = types.OCIManifestSchema1
	case string(types.DockerManifestList):
		mediaType = types.DockerManifestList
	case string(types.OCIImageIndex):
		mediaType = types.OCIImageIndex
	default:
		// Default to DockerManifestSchema2 if unknown
		mediaType = types.DockerManifestSchema2
	}

	// Create a descriptor
	hash, size, err := v1.SHA256(io.NewSectionReader(strings.NewReader(string(manifestBytes)), 0, int64(len(manifestBytes))))
	if err != nil {
		return fmt.Errorf("failed to calculate manifest hash: %w", err)
	}

	img := &remoteImage{
		manifest: &v1.Manifest{
			MediaType: mediaType,
		},
		rawManifest: manifestBytes,
		digest:      hash,
		mediaType:   mediaType,
		size:        size,
	}

	// Upload the manifest
	if err := remote.Put(ref, img, r.client.transportOpt); err != nil {
		return fmt.Errorf("failed to upload manifest: %w", err)
	}

	return nil
}

// DeleteManifest deletes the manifest for the given tag
func (r *Repository) DeleteManifest(tag string) error {
	// Get the manifest to get the digest
	manifest, _, err := r.GetManifest(tag)
	if err != nil {
		return fmt.Errorf("failed to get manifest for deletion: %w", err)
	}

	// Calculate the digest
	hash, _, err := v1.SHA256(io.NewSectionReader(strings.NewReader(string(manifest)), 0, int64(len(manifest))))
	if err != nil {
		return fmt.Errorf("failed to calculate manifest hash: %w", err)
	}

	// Repository name in the format registry.amazonaws.com/repo
	fullName := r.repository.String()
	// Extract just the repository name without the registry prefix
	parts := strings.Split(fullName, "/")
	repoName := parts[len(parts)-1]

	// Delete the image using the digest
	digestStr := hash.String()
	_, err = r.client.ecr.BatchDeleteImage(context.Background(), &awsecr.BatchDeleteImageInput{
		RepositoryName: &repoName,
		ImageIds: []ecrtypes.ImageIdentifier{
			{
				ImageDigest: &digestStr,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete manifest: %w", err)
	}

	return nil
}

// remoteImage implements the v1.Image interface for use with remote.Put
type remoteImage struct {
	manifest    *v1.Manifest
	rawManifest []byte
	digest      v1.Hash
	mediaType   types.MediaType
	size        int64
}

// Digest implements v1.Image
func (i *remoteImage) Digest() (v1.Hash, error) {
	return i.digest, nil
}

// Manifest implements v1.Image
func (i *remoteImage) Manifest() (*v1.Manifest, error) {
	return i.manifest, nil
}

// RawManifest implements v1.Image
func (i *remoteImage) RawManifest() ([]byte, error) {
	return i.rawManifest, nil
}

// MediaType implements v1.Image
func (i *remoteImage) MediaType() (types.MediaType, error) {
	return i.mediaType, nil
}

// Size implements v1.Image
func (i *remoteImage) Size() (int64, error) {
	return i.size, nil
}

// ConfigName implements v1.Image
func (i *remoteImage) ConfigName() (v1.Hash, error) {
	return v1.Hash{}, fmt.Errorf("not implemented")
}

// ConfigFile implements v1.Image
func (i *remoteImage) ConfigFile() (*v1.ConfigFile, error) {
	return nil, fmt.Errorf("not implemented")
}

// Layers implements v1.Image
func (i *remoteImage) Layers() ([]v1.Layer, error) {
	return nil, fmt.Errorf("not implemented")
}

// LayerByDigest implements v1.Image
func (i *remoteImage) LayerByDigest(v1.Hash) (v1.Layer, error) {
	return nil, fmt.Errorf("not implemented")
}

// LayerByDiffID implements v1.Image
func (i *remoteImage) LayerByDiffID(v1.Hash) (v1.Layer, error) {
	return nil, fmt.Errorf("not implemented")
}
