package service

import (
	"context"

	"freightliner/pkg/interfaces"
)

// Import types from the shared interfaces package for compatibility
type (
	Repository            = interfaces.Repository
	RegistryClient        = interfaces.RegistryClient
	Manifest              = interfaces.Manifest
	InterfaceRepoInfo     = interfaces.RepositoryInfo // Renamed to avoid conflict
	TagLister             = interfaces.TagLister
	ManifestAccessor      = interfaces.ManifestAccessor
	ManifestManager       = interfaces.ManifestManager
	LayerAccessor         = interfaces.LayerAccessor
	RemoteImageAccessor   = interfaces.RemoteImageAccessor
	RegistryAuthenticator = interfaces.RegistryAuthenticator
)

// RepositoryCreator is an interface for client types that can create repositories
type RepositoryCreator interface {
	// CreateRepository creates a new repository with the given name and tags
	CreateRepository(ctx context.Context, name string, tags map[string]string) (Repository, error)
}
