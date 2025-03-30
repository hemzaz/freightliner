package common

import (
	"context"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Manifest represents a container image manifest
type Manifest struct {
	// Content is the raw manifest content
	Content []byte

	// MediaType is the content type of the manifest
	MediaType string

	// Digest is the SHA256 digest of the manifest
	Digest string
}

// Repository represents a container repository in a registry
type Repository interface {
	// GetRepositoryName returns the name of the repository
	GetRepositoryName() string

	// GetName is an alias for GetRepositoryName for backward compatibility
	GetName() string

	// ListTags returns all tags for the repository
	ListTags() ([]string, error)

	// GetManifest returns the manifest for the given tag
	GetManifest(ctx context.Context, tag string) (*Manifest, error)

	// PutManifest uploads a manifest with the given tag
	PutManifest(ctx context.Context, tag string, manifest *Manifest) error

	// DeleteManifest deletes the manifest for the given tag
	DeleteManifest(ctx context.Context, tag string) error

	// GetLayerReader returns a reader for the layer with the given digest
	GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error)

	// GetImageReference returns a name.Reference for the given tag
	GetImageReference(tag string) (name.Reference, error)

	// GetRemoteOptions returns options for remote operations
	GetRemoteOptions() ([]remote.Option, error)
}
