// Package interfaces provides shared interface definitions for Freightliner
// This package is designed to be a central location for interfaces that are used
// across multiple packages to prevent circular dependencies.
//
// NOTE ON DEPENDENCY INVERSION PRINCIPLE:
// Normally, interfaces should be defined in the package that uses them, not in
// the package that implements them. In a perfect DIP implementation, each consumer
// would define exactly the interface it needs.
//
// However, having a central interfaces package is a pragmatic compromise when:
// 1. Multiple packages need the same or similar interfaces, which would create duplication
// 2. Circular dependencies would occur if each package defined its own interfaces
// 3. The codebase needs a common "language" of interfaces as a public API
//
// When using interfaces from this package:
//   - Consider defining your own package-specific interfaces if you don't need
//     the full interface definition (see pkg/copy/interfaces.go for an example)
//   - Prefer type aliases to reference these interfaces rather than
//     copying interface definitions
//   - Add new interfaces here only if they're used by multiple packages
package interfaces

import (
	"context"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// RepositoryName provides repository naming functionality
type RepositoryName interface {
	// GetName returns the name of the repository
	GetName() string
}

// RepositoryInfo extends RepositoryName with additional repository information
type RepositoryInfo interface {
	RepositoryName

	// GetRepositoryName returns the name of the repository
	GetRepositoryName() string
}

// TagLister provides tag listing functionality
type TagLister interface {
	// ListTags returns all tags for the repository
	ListTags(ctx context.Context) ([]string, error)
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

// ManifestAccessor provides access to manifests
type ManifestAccessor interface {
	// GetManifest returns the manifest for the given tag
	GetManifest(ctx context.Context, tag string) (*Manifest, error)
}

// ManifestManager extends ManifestAccessor with additional manifest operations
type ManifestManager interface {
	ManifestAccessor

	// PutManifest uploads a manifest with the given tag
	PutManifest(ctx context.Context, tag string, manifest *Manifest) error

	// DeleteManifest deletes the manifest for the given tag
	DeleteManifest(ctx context.Context, tag string) error
}

// LayerAccessor provides access to image layers
type LayerAccessor interface {
	// GetLayerReader returns a reader for the layer with the given digest
	GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error)
}

// ImageReferencer provides image reference capabilities
type ImageReferencer interface {
	// GetImageReference returns a name.Reference for the given tag
	GetImageReference(tag string) (name.Reference, error)
}

// RemoteOptionsProvider provides remote access options
type RemoteOptionsProvider interface {
	// GetRemoteOptions returns options for remote operations
	GetRemoteOptions() ([]remote.Option, error)
}

// ImageGetter provides ability to fetch container images
type ImageGetter interface {
	// GetImage retrieves the v1.Image for the given tag
	GetImage(ctx context.Context, tag string) (v1.Image, error)
}

// RemoteImageAccessor combines capabilities for remote image access
type RemoteImageAccessor interface {
	ImageReferencer
	RemoteOptionsProvider
	ImageGetter
}

// BasicRepository combines the core repository interfaces used across packages
type BasicRepository interface {
	RepositoryInfo
	TagLister
	ManifestAccessor
}

// Repository represents a complete container repository with all capabilities
type Repository interface {
	RepositoryInfo
	TagLister
	ManifestManager
	LayerAccessor
	RemoteImageAccessor
}

// RegistryClient defines the interface for registry clients
type RegistryClient interface {
	// ListRepositories lists all repositories in a registry with the given prefix
	ListRepositories(ctx context.Context, prefix string) ([]string, error)

	// GetRepository returns a repository reference for the given name
	GetRepository(ctx context.Context, name string) (Repository, error)

	// GetRegistryName returns the name of the registry
	GetRegistryName() string
}

// RegistryProvider defines the interface for registry providers that create clients
type RegistryProvider interface {
	// GetClient returns a registry client for the given type and endpoint
	GetClient(ctx context.Context, endpoint string) (RegistryClient, error)
}
