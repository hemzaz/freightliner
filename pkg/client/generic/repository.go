package generic

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

// Repository implements the repository interface for generic registries
type Repository struct {
	*common.BaseRepository
	client     *Client
	name       string
	repository name.Repository
}

// GetName returns the repository name
func (r *Repository) GetName() string {
	return r.name
}

// GetFullName returns the full repository name including registry
func (r *Repository) GetFullName() string {
	return r.repository.Name()
}

// GetTag retrieves an image by tag
func (r *Repository) GetTag(ctx context.Context, tag string) (v1.Image, error) {
	if r.BaseRepository != nil {
		return r.BaseRepository.GetTag(ctx, tag)
	}
	return nil, errors.NotImplementedf("base repository not initialized")
}

// DeleteManifest deletes a manifest (required by Repository interface)
func (r *Repository) DeleteManifest(ctx context.Context, digest string) error {
	// Generic registries may not support manifest deletion
	return errors.NotImplementedf("manifest deletion not supported for generic registries")
}

// GetImageReference returns a name.Reference for the given tag
func (r *Repository) GetImageReference(tag string) (name.Reference, error) {
	return name.ParseReference(fmt.Sprintf("%s:%s", r.repository.Name(), tag))
}

// GetLayerReader returns a layer reader for a specific digest (required by Repository interface)
func (r *Repository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	// This is a placeholder - actual implementation would use go-containerregistry to get layer
	return nil, errors.NotImplementedf("layer reader not yet implemented for generic registries")
}

// GetManifest returns a manifest for the given tag or digest
func (r *Repository) GetManifest(ctx context.Context, ref string) (*interfaces.Manifest, error) {
	// Delegate to base repository if available
	if r.BaseRepository != nil {
		// Base repository doesn't have GetManifest, so we need to implement it
		// For now, return not implemented
		return nil, errors.NotImplementedf("manifest retrieval not yet fully implemented for generic registries")
	}
	return nil, errors.NotImplementedf("base repository not initialized")
}

// GetRemoteOptions returns remote options for registry operations
func (r *Repository) GetRemoteOptions() ([]remote.Option, error) {
	// Return the client's remote options
	if r.client != nil {
		return r.client.GetRemoteOptions(), nil
	}
	return []remote.Option{}, nil
}

// GetRepositoryName returns the repository name (alias for GetName)
func (r *Repository) GetRepositoryName() string {
	return r.GetName()
}

// PutManifest uploads a manifest to the repository
func (r *Repository) PutManifest(ctx context.Context, ref string, manifest *interfaces.Manifest) error {
	return errors.NotImplementedf("manifest upload not yet implemented for generic registries")
}

