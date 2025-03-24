package gcr

import (
	"context"
	"fmt"
	
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Repository implements the repository interface for Google GCR
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
	ctx := context.Background()
	
	// Use the go-containerregistry library to list tags
	tags, err := google.List(r.repository.String(), r.client.transportOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	
	// Extract tag names from the full image references
	var tagNames []string
	repoPrefix := r.repository.String() + ":"
	
	for _, tag := range tags {
		if tag == r.repository.String() {
			// Skip the repository itself
			continue
		}
		
		// Check if this is a tag reference (has format repo:tag)
		if len(tag) > len(repoPrefix) && tag[:len(repoPrefix)] == repoPrefix {
			tagName := tag[len(repoPrefix):]
			tagNames = append(tagNames, tagName)
		}
	}
	
	return tagNames, nil
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
	
	// Create a reference to the image with the given tag
	tagRef, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.Name(), tag))
	if err != nil {
		return fmt.Errorf("failed to create tag reference: %w", err)
	}
	
	// In GCR, we delete by tagging with "gcr.io/google/cloud-builders/gcloud" and using gcrDelete flag
	// But for go-containerregistry, there's a better way using the remote.Delete function
	if err := remote.Delete(tagRef, r.client.transportOpt); err != nil {
		return fmt.Errorf("failed to delete manifest: %w", err)
	}
	
	return nil
}
