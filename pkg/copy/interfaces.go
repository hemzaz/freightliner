package copy

import (
	"context"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Repository represents a container repository interface needed for copy operations
type Repository interface {
	// GetImageReference returns a name.Reference for the given tag
	GetImageReference(tag string) (name.Reference, error)

	// GetName returns the name of the repository
	GetName() string

	// GetRemoteOptions returns options for remote operations
	GetRemoteOptions() ([]remote.Option, error)

	// GetImage retrieves the v1.Image for the given tag
	GetImage(ctx context.Context, tag string) (v1.Image, error)
}

// ManifestAccessor provides access to image manifests
type ManifestAccessor interface {
	// GetManifest returns the manifest for the given tag
	GetManifest(ctx context.Context, tag string) (*Manifest, error)

	// PutManifest uploads a manifest with the given tag
	PutManifest(ctx context.Context, tag string, manifest *Manifest) error
}

// Manifest represents a container image manifest
type Manifest struct {
	// Content is the raw manifest content
	Content []byte

	// MediaType is the content type of the manifest
	MediaType string

	// Digest is the SHA256 digest of the manifest
	Digest string
}

// LayerAccessor provides access to image layers
type LayerAccessor interface {
	// GetLayerReader returns a reader for the layer with the given digest
	GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error)
}
