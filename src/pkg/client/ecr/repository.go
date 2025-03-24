package ecr

import (
	"context"
	"fmt"
	"io"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/elad/freightliner/pkg/client/common"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Repository implements the repository interface for AWS ECR
type Repository struct {
	client     *Client
	name       string
	repository name.Repository
}

// ListTags returns all tags for the repository
func (r *Repository) ListTags() ([]string, error) {
	ctx := context.Background()
	var tags []string
	var nextToken *string
	
	// Get repository name without registry prefix
	repoName := r.repository.RepositoryStr()
	
	for {
		// Call the API to list tags
		resp, err := r.client.ecrClient.ListImages(ctx, &ecr.ListImagesInput{
			RepositoryName: aws.String(repoName),
			NextToken:      nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list tags: %w", err)
		}
		
		// Add tags to the list
		for _, image := range resp.ImageIds {
			if image.ImageTag != nil {
				tags = append(tags, aws.ToString(image.ImageTag))
			}
		}
		
		// Check if there are more tags
		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}
	
	return tags, nil
}

// GetManifest returns the manifest for the given tag
func (r *Repository) GetManifest(tag string) ([]byte, string, error) {
	ctx := context.Background()
	
	// Create a reference to the image with the given tag
	ref, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.Name(), tag))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create tag reference: %w", err)
	}
	
	// Get the image from the registry
	img, err := remote.Image(ref, r.client.transportOpt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get image: %w", err)
	}
	
	// Get the manifest
	manifest, err := img.RawManifest()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get manifest: %w", err)
	}
	
	// Get the manifest media type
	digest, err := img.Digest()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get digest: %w", err)
	}
	
	// Get media type from the descriptor
	desc, err := r.getDescriptor(ref, digest)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get descriptor: %w", err)
	}
	
	return manifest, string(desc.MediaType), nil
}

// getDescriptor returns the descriptor for an image
func (r *Repository) getDescriptor(ref name.Reference, digest v1.Hash) (*v1.Descriptor, error) {
	ctx := context.Background()
	
	// Get the descriptor from the registry
	desc, err := remote.Get(ref, r.client.transportOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to get descriptor: %w", err)
	}
	
	return &desc, nil
}

// PutManifest uploads a manifest with the given tag
func (r *Repository) PutManifest(tag string, manifest []byte, mediaType string) error {
	ctx := context.Background()
	
	// Create a reference to the image with the given tag
	tagRef, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.Name(), tag))
	if err != nil {
		return fmt.Errorf("failed to create tag reference: %w", err)
	}
	
	// Create a transport for uploads
	transport, err := r.client.GetTransport(r.name)
	if err != nil {
		return fmt.Errorf("failed to get transport: %w", err)
	}
	
	// Upload the manifest
	if err := remote.Put(tagRef, remote.Descriptor{
		Size:      int64(len(manifest)),
		Digest:    v1.Hash{}, // Will be calculated by the remote library
		MediaType: mediaType,
		Data:      manifest,
	}, r.client.transportOpt); err != nil {
		return fmt.Errorf("failed to upload manifest: %w", err)
	}
	
	return nil
}

// DeleteManifest deletes the manifest for the given tag
func (r *Repository) DeleteManifest(tag string) error {
	ctx := context.Background()
	
	// Get repository name without registry prefix
	repoName := r.repository.RepositoryStr()
	
	// Find the image ID for the given tag
	resp, err := r.client.ecrClient.ListImages(ctx, &ecr.ListImagesInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}
	
	var imageID *types.ImageIdentifier
	for _, image := range resp.ImageIds {
		if image.ImageTag != nil && aws.ToString(image.ImageTag) == tag {
			imageID = &image
			break
		}
	}
	
	if imageID == nil {
		return fmt.Errorf("image with tag %s not found", tag)
	}
	
	// Delete the image
	_, err = r.client.ecrClient.BatchDeleteImage(ctx, &ecr.BatchDeleteImageInput{
		RepositoryName: aws.String(repoName),
		ImageIds:       []types.ImageIdentifier{*imageID},
	})
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	
	return nil
}
