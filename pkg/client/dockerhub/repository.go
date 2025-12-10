package dockerhub

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

// Repository represents a Docker Hub repository
type Repository struct {
	*common.BaseRepository
	client     *Client
	name       string
	repository name.Repository
}

// ListTags lists all tags for this repository with retry logic
func (r *Repository) ListTags(ctx context.Context) ([]string, error) {
	var tags []string
	var listErr error

	// Use retry logic for rate-limited operations
	err := r.client.executeWithRetry(ctx, "ListTags", func() error {
		var err error
		tags, err = remote.List(r.repository, r.client.GetRemoteOptions()...)
		if err != nil {
			listErr = err
			return err
		}
		return nil
	})

	if err != nil {
		r.client.logger.WithFields(map[string]interface{}{
			"repository": r.name,
			"error":      err.Error(),
		}).Error("Failed to list tags from Docker Hub", err)
		return nil, errors.Wrap(listErr, "failed to list tags")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tagCount":   len(tags),
	}).Debug("Successfully listed Docker Hub tags")

	return tags, nil
}

// GetManifest retrieves a manifest by tag or digest with retry logic
func (r *Repository) GetManifest(ctx context.Context, reference string) (*interfaces.Manifest, error) {
	var manifest *interfaces.Manifest
	var getErr error

	// Create reference
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", r.repository.Name(), reference))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse reference")
	}

	// Use retry logic for rate-limited operations
	err = r.client.executeWithRetry(ctx, "GetManifest", func() error {
		desc, err := remote.Get(ref, r.client.GetRemoteOptions()...)
		if err != nil {
			getErr = err
			return err
		}

		// Extract manifest content
		manifestBytes, err := desc.RawManifest()
		if err != nil {
			getErr = err
			return err
		}

		manifest = &interfaces.Manifest{
			Content:   manifestBytes,
			MediaType: string(desc.MediaType),
			Digest:    desc.Digest.String(),
		}
		return nil
	})

	if err != nil {
		r.client.logger.WithFields(map[string]interface{}{
			"repository": r.name,
			"reference":  reference,
			"error":      err.Error(),
		}).Error("Failed to get manifest from Docker Hub", err)
		return nil, errors.Wrap(getErr, "failed to get manifest")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"reference":  reference,
		"digest":     manifest.Digest,
	}).Debug("Successfully retrieved Docker Hub manifest")

	return manifest, nil
}

// PutManifest uploads a manifest with the given tag
func (r *Repository) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	return errors.NotImplementedf("manifest upload not supported for Docker Hub public registry")
}

// DeleteManifest deletes a manifest by tag or digest
func (r *Repository) DeleteManifest(ctx context.Context, reference string) error {
	return errors.NotImplementedf("manifest deletion not supported for Docker Hub public registry")
}

// GetImage retrieves an image by tag with retry logic
func (r *Repository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	var img v1.Image
	var getErr error

	// Create reference
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", r.repository.Name(), tag))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse reference")
	}

	// Use retry logic for rate-limited operations
	err = r.client.executeWithRetry(ctx, "GetImage", func() error {
		remoteImg, err := remote.Image(ref, r.client.GetRemoteOptions()...)
		if err != nil {
			getErr = err
			return err
		}
		img = remoteImg
		return nil
	})

	if err != nil {
		r.client.logger.WithFields(map[string]interface{}{
			"repository": r.name,
			"tag":        tag,
			"error":      err.Error(),
		}).Error("Failed to get image from Docker Hub", err)
		return nil, errors.Wrap(getErr, "failed to get image")
	}

	r.client.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tag,
	}).Debug("Successfully retrieved Docker Hub image")

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
