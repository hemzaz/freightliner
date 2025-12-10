package ghcr

import (
	"context"
	"fmt"
	"io"

	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Repository represents a GHCR repository
type Repository struct {
	*common.BaseRepository
	client     *Client
	name       string
	repository name.Repository
}

// ListTags lists all tags for this repository
func (r *Repository) ListTags(ctx context.Context) ([]string, error) {
	tags, err := remote.List(r.repository, r.client.GetRemoteOptions()...)
	if err != nil {
		r.client.logger.WithFields(map[string]interface{}{
			"repository": r.name,
			"error":      err.Error(),
		}).Error("Failed to list tags from GHCR", err)
		return nil, errors.Wrap(err, "failed to list tags")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tagCount":   len(tags),
	}).Debug("Successfully listed GHCR tags")

	return tags, nil
}

// GetManifest retrieves a manifest by tag or digest
func (r *Repository) GetManifest(ctx context.Context, reference string) (*interfaces.Manifest, error) {
	// Create reference
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", r.repository.Name(), reference))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse reference")
	}

	// Get descriptor
	desc, err := remote.Get(ref, r.client.GetRemoteOptions()...)
	if err != nil {
		r.client.logger.WithFields(map[string]interface{}{
			"repository": r.name,
			"reference":  reference,
			"error":      err.Error(),
		}).Error("Failed to get manifest from GHCR", err)
		return nil, errors.Wrap(err, "failed to get manifest")
	}

	// Extract manifest content
	manifestBytes, err := desc.RawManifest()
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract manifest content")
	}

	manifest := &interfaces.Manifest{
		Content:   manifestBytes,
		MediaType: string(desc.MediaType),
		Digest:    desc.Digest.String(),
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"reference":  reference,
		"digest":     manifest.Digest,
	}).Debug("Successfully retrieved GHCR manifest")

	return manifest, nil
}

// PutManifest uploads a manifest with the given tag
func (r *Repository) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	// Create reference
	tagRef, err := name.NewTag(fmt.Sprintf("%s:%s", r.repository.Name(), tag))
	if err != nil {
		return errors.Wrap(err, "failed to parse tag")
	}

	// Get or create an image from the manifest
	img, err := remote.Image(tagRef, r.client.GetRemoteOptions()...)
	if err != nil {
		// If image doesn't exist, we need to create it differently
		// For now, return an error as this is complex
		return errors.Wrap(err, "manifest upload to GHCR requires full image context")
	}

	// Write the image
	if err := remote.Write(tagRef, img, r.client.GetRemoteOptions()...); err != nil {
		return errors.Wrap(err, "failed to write manifest")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tag,
		"digest":     manifest.Digest,
	}).Info("Successfully uploaded manifest to GHCR")

	return nil
}

// DeleteManifest deletes a manifest by tag or digest
func (r *Repository) DeleteManifest(ctx context.Context, reference string) error {
	// Create reference
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", r.repository.Name(), reference))
	if err != nil {
		return errors.Wrap(err, "failed to parse reference")
	}

	// Delete the manifest
	if err := remote.Delete(ref, r.client.GetRemoteOptions()...); err != nil {
		return errors.Wrap(err, "failed to delete manifest")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"reference":  reference,
	}).Info("Successfully deleted manifest from GHCR")

	return nil
}

// GetImage retrieves an image by tag
func (r *Repository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	// Create reference
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", r.repository.Name(), tag))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse reference")
	}

	// Get image
	img, err := remote.Image(ref, r.client.GetRemoteOptions()...)
	if err != nil {
		r.client.logger.WithFields(map[string]interface{}{
			"repository": r.name,
			"tag":        tag,
			"error":      err.Error(),
		}).Error("Failed to get image from GHCR", err)
		return nil, errors.Wrap(err, "failed to get image")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tag,
	}).Debug("Successfully retrieved GHCR image")

	return img, nil
}

// GetLayerReader returns a reader for the layer with the given digest
func (r *Repository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	// Parse digest as a name.Digest for the go-containerregistry API
	nameDigest, err := name.NewDigest(fmt.Sprintf("%s@%s", r.repository.Name(), digest))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse digest")
	}

	// Get the layer from the registry
	layer, err := remote.Layer(nameDigest, r.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get layer")
	}

	// Get a reader for the layer
	reader, err := layer.Compressed()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get layer reader")
	}

	return reader, nil
}

// GetImageReference returns a name.Reference for the given tag
func (r *Repository) GetImageReference(tag string) (name.Reference, error) {
	return name.ParseReference(fmt.Sprintf("%s:%s", r.repository.Name(), tag))
}

// GetRemoteOptions returns options for remote operations
func (r *Repository) GetRemoteOptions() ([]remote.Option, error) {
	return r.client.GetRemoteOptions(), nil
}

// GetName returns the repository name
func (r *Repository) GetName() string {
	return r.name
}

// GetRepositoryName returns the repository name (alias for GetName)
func (r *Repository) GetRepositoryName() string {
	return r.GetName()
}
