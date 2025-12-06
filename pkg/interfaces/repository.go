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

// LayerDescriptor represents a layer in a manifest
type LayerDescriptor struct {
	// Digest is the digest of the layer
	Digest string `json:"digest"`

	// Size is the size of the layer in bytes
	Size int64 `json:"size"`

	// MediaType is the media type of the layer
	MediaType string `json:"mediaType"`
}

// Manifest represents a container image manifest
type Manifest struct {
	// Content is the raw manifest content
	Content []byte

	// MediaType is the content type of the manifest
	MediaType string

	// Digest is the SHA256 digest of the manifest
	Digest string

	// SchemaVersion is the manifest schema version
	SchemaVersion int `json:"schemaVersion,omitempty"`

	// Layers is the list of layers in the manifest
	Layers []LayerDescriptor `json:"layers,omitempty"`

	// Config is the config descriptor
	Config *LayerDescriptor `json:"config,omitempty"`
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
// NOTE: This large interface is kept for backward compatibility.
// New code should use the more focused interfaces below for better testability.
type Repository interface {
	RepositoryInfo
	TagLister
	ManifestManager
	LayerAccessor
	RemoteImageAccessor
}

// ===== SEGREGATED INTERFACES FOR SINGLE RESPONSIBILITIES =====

// Reader provides read-only access to repository data
type Reader interface {
	RepositoryInfo
	TagLister
	ManifestAccessor
	LayerAccessor
}

// Writer provides write access to repository data
type Writer interface {
	RepositoryInfo
	ManifestManager
}

// ImageProvider provides image access capabilities
type ImageProvider interface {
	RepositoryInfo
	ImageReferencer
	RemoteOptionsProvider
	ImageGetter
}

// MetadataProvider provides repository metadata access
type MetadataProvider interface {
	RepositoryInfo
	TagLister
}

// ContentProvider provides content access capabilities
type ContentProvider interface {
	ManifestAccessor
	LayerAccessor
}

// ContentManager provides content management capabilities
type ContentManager interface {
	ManifestManager
	LayerAccessor
}

// Note: RegistryClient interface is defined in client.go to avoid duplication

// RegistryProvider defines the interface for registry providers that create clients
type RegistryProvider interface {
	// GetClient returns a registry client for the given type and endpoint
	GetClient(ctx context.Context, endpoint string) (RegistryClient, error)
}

// ===== COMPOSITION PATTERNS FOR COMPLEX BEHAVIORS =====

// RepositoryComposer provides composition of repository behaviors
type RepositoryComposer interface {
	// AsReader returns a read-only view of the repository
	AsReader() Reader

	// AsWriter returns a write-only view of the repository
	AsWriter() Writer

	// AsImageProvider returns an image provider view
	AsImageProvider() ImageProvider

	// AsMetadataProvider returns a metadata provider view
	AsMetadataProvider() MetadataProvider

	// AsContentProvider returns a content provider view
	AsContentProvider() ContentProvider

	// AsContentManager returns a content manager view
	AsContentManager() ContentManager
}

// ReadWriteRepository combines read and write capabilities
type ReadWriteRepository interface {
	Reader
	Writer
}

// ImageRepository combines image and metadata access
type ImageRepository interface {
	ImageProvider
	MetadataProvider
}

// FullRepository provides all repository capabilities with composition
type FullRepository interface {
	Reader
	Writer
	ImageProvider
	RepositoryComposer
}

// ===== CONTEXT-AWARE INTERFACES =====

// ContextualTagLister extends TagLister with context-aware batch operations
type ContextualTagLister interface {
	TagLister

	// ListTagsWithLimit returns tags with pagination support
	ListTagsWithLimit(ctx context.Context, limit int, offset int) ([]string, error)

	// CountTags returns the total number of tags
	CountTags(ctx context.Context) (int, error)
}

// ContextualManifestManager extends ManifestManager with batch operations
type ContextualManifestManager interface {
	ManifestManager

	// GetManifestsBatch retrieves multiple manifests efficiently
	GetManifestsBatch(ctx context.Context, tags []string) (map[string]*Manifest, error)

	// PutManifestsBatch uploads multiple manifests efficiently
	PutManifestsBatch(ctx context.Context, manifests map[string]*Manifest) error

	// DeleteManifestsBatch deletes multiple manifests efficiently
	DeleteManifestsBatch(ctx context.Context, tags []string) error
}

// ContextualLayerAccessor extends LayerAccessor with efficient layer access
type ContextualLayerAccessor interface {
	LayerAccessor

	// GetLayerReadersBatch returns readers for multiple layers
	GetLayerReadersBatch(ctx context.Context, digests []string) (map[string]io.ReadCloser, error)

	// GetLayerInfo returns layer metadata without downloading content
	GetLayerInfo(ctx context.Context, digest string) (*LayerInfo, error)
}

// LayerInfo provides metadata about a layer
type LayerInfo struct {
	Digest     string
	Size       int64
	MediaType  string
	Compressed bool
	Exists     bool
}

// ===== HIGH-PERFORMANCE INTERFACES =====

// StreamingRepository provides streaming access to repository data
type StreamingRepository interface {
	RepositoryInfo

	// StreamTags provides a streaming interface for large tag lists
	StreamTags(ctx context.Context) (<-chan string, <-chan error)

	// StreamManifests provides streaming access to manifests
	StreamManifests(ctx context.Context, tags <-chan string) (<-chan ManifestResult, <-chan error)
}

// ManifestResult combines a manifest with its tag for streaming operations
type ManifestResult struct {
	Tag      string
	Manifest *Manifest
	Error    error
}
