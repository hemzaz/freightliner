package copy

import (
	"freightliner/pkg/interfaces"
)

// Import types from the shared interfaces package for compatibility
type (
	RepositoryName        = interfaces.RepositoryName
	ImageReferencer       = interfaces.ImageReferencer
	RemoteOptionsProvider = interfaces.RemoteOptionsProvider
	ImageGetter           = interfaces.ImageGetter
	Manifest              = interfaces.Manifest
	ManifestAccessor      = interfaces.ManifestAccessor
	LayerAccessor         = interfaces.LayerAccessor
)

// Repository represents a container repository interface needed for copy operations
// This is a local interface that defines exactly what operations the copy package
// requires from a repository, following the Interface Segregation Principle.
// It's intentionally more limited than interfaces.Repository.
type Repository interface {
	RepositoryName
	ImageReferencer
	RemoteOptionsProvider
	ImageGetter
	ManifestAccessor
}
