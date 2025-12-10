package quay

import (
	"context"
	"fmt"
	"io"
	"strings"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Repository implements the interfaces.Repository interface for Quay
type Repository struct {
	client     *Client
	name       string
	repository name.Repository
}

// Name returns the repository name
func (r *Repository) Name() string {
	return r.name
}

// GetName returns the repository name (alias for Name)
func (r *Repository) GetName() string {
	return r.name
}

// GetRepositoryName returns the repository name
func (r *Repository) GetRepositoryName() string {
	return r.name
}

// GetReference returns a reference for the given tag
func (r *Repository) GetReference(tag string) (name.Reference, error) {
	if tag == "" {
		return nil, errors.InvalidInputf("tag cannot be empty")
	}

	ref := fmt.Sprintf("%s:%s", r.repository.Name(), tag)
	return name.ParseReference(ref)
}

// ListTags lists all tags for this repository
func (r *Repository) ListTags(ctx context.Context) ([]string, error) {
	tags, err := remote.List(r.repository, r.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list tags")
	}
	return tags, nil
}

// GetManifest retrieves the manifest for a specific tag
func (r *Repository) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	ref, err := r.GetReference(tag)
	if err != nil {
		return nil, err
	}

	desc, err := remote.Get(ref, r.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get manifest")
	}

	manifest := &interfaces.Manifest{
		Digest:    desc.Digest.String(),
		MediaType: string(desc.MediaType),
	}

	return manifest, nil
}

// PutManifest uploads a manifest with the given tag
func (r *Repository) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	return errors.NotImplementedf("PutManifest not yet implemented for Quay")
}

// DeleteManifest deletes a manifest by tag or digest
func (r *Repository) DeleteManifest(ctx context.Context, tagOrDigest string) error {
	ref, err := r.GetReference(tagOrDigest)
	if err != nil {
		return err
	}

	if err := remote.Delete(ref, r.client.GetRemoteOptions()...); err != nil {
		return errors.Wrap(err, "failed to delete manifest")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"reference":  tagOrDigest,
		"registry":   r.client.GetRegistryName(),
	}).Info("Successfully deleted manifest from Quay")

	return nil
}

// GetImage retrieves an image by tag
func (r *Repository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	ref, err := r.GetReference(tag)
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(ref, r.client.GetRemoteOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get image")
	}

	return img, nil
}

// PutImage pushes an image with the given tag
func (r *Repository) PutImage(ctx context.Context, tag string, img v1.Image) error {
	ref, err := r.GetReference(tag)
	if err != nil {
		return err
	}

	if err := remote.Write(ref, img, r.client.GetRemoteOptions()...); err != nil {
		return errors.Wrap(err, "failed to push image")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tag,
		"registry":   r.client.GetRegistryName(),
	}).Info("Successfully pushed image to Quay")

	return nil
}

// DeleteTag deletes a tag from the repository
func (r *Repository) DeleteTag(ctx context.Context, tag string) error {
	ref, err := r.GetReference(tag)
	if err != nil {
		return err
	}

	if err := remote.Delete(ref, r.client.GetRemoteOptions()...); err != nil {
		return errors.Wrap(err, "failed to delete tag")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tag,
		"registry":   r.client.GetRegistryName(),
	}).Info("Successfully deleted tag from Quay")

	return nil
}

// GetLayers retrieves all layers for a given image tag
func (r *Repository) GetLayers(ctx context.Context, tag string) ([]v1.Layer, error) {
	img, err := r.GetImage(ctx, tag)
	if err != nil {
		return nil, err
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get image layers")
	}

	return layers, nil
}

// GetLayerReader returns a reader for the layer with the given digest
func (r *Repository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	return nil, errors.NotImplementedf("GetLayerReader not yet implemented for Quay")
}

// Exists checks if a tag exists in the repository
func (r *Repository) Exists(ctx context.Context, tag string) (bool, error) {
	ref, err := r.GetReference(tag)
	if err != nil {
		return false, err
	}

	_, err = remote.Head(ref, r.client.GetRemoteOptions()...)
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to check tag existence")
	}

	return true, nil
}

// GetImageReference returns a name.Reference for the given tag
func (r *Repository) GetImageReference(tag string) (name.Reference, error) {
	return r.GetReference(tag)
}

// GetRemoteOptions returns options for remote operations
func (r *Repository) GetRemoteOptions() ([]remote.Option, error) {
	return r.client.GetRemoteOptions(), nil
}

// GetInfo returns repository information (placeholder for interface compatibility)
func (r *Repository) GetInfo(ctx context.Context) (interface{}, error) {
	tags, err := r.ListTags(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":      r.name,
		"registry":  r.client.GetRegistryName(),
		"tag_count": len(tags),
		"tags":      tags,
	}, nil
}

// isNotFoundError checks if an error is a "not found" error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "MANIFEST_UNKNOWN")
}
