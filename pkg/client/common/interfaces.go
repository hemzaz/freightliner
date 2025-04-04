package common

import (
	"context"
	"fmt"
	"freightliner/pkg/helper/errors"
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// RegistryClient defines the interface for registry clients
type RegistryClient interface {
	// ListRepositories lists all repositories in a registry with the given prefix
	ListRepositories(ctx context.Context, prefix string) ([]string, error)

	// GetRepository returns a repository reference for the given name
	GetRepository(ctx context.Context, name string) (Repository, error)

	// GetRegistryName returns the name of the registry
	GetRegistryName() string
}

// RegistryProvider defines the interface for registry client providers
type RegistryProvider interface {
	// GetRegistryClient returns a registry client for the given registry type
	GetRegistryClient(registryType string) (RegistryClient, error)
}

// RepositoryCreator is an interface for client types that can create repositories
type RepositoryCreator interface {
	// CreateRepository creates a new repository with the given name and tags
	CreateRepository(ctx context.Context, name string, tags map[string]string) (Repository, error)
}

// Repository represents a container repository in a registry
type Repository interface {
	// GetRepositoryName returns the name of the repository
	GetRepositoryName() string

	// GetName is an alias for GetRepositoryName for backward compatibility
	GetName() string

	// ListTags returns all tags for the repository
	ListTags(ctx context.Context) ([]string, error)

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

	// GetImage retrieves the v1.Image for the given tag
	GetImage(ctx context.Context, tag string) (v1.Image, error)
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

// RegistryAuthenticator defines the interface for registry authentication
type RegistryAuthenticator interface {
	// GetAuthToken returns an authentication token for the registry
	GetAuthToken(ctx context.Context, registry string) (string, error)

	// GetAuthHeader returns an authentication header for the registry
	GetAuthHeader(ctx context.Context, registry string) (string, error)

	// GetAuthenticator returns an authn.Authenticator for the registry
	GetAuthenticator(ctx context.Context, registry string) (authn.Authenticator, error)
}

// FormRegistryPath creates a properly formatted registry path
func FormRegistryPath(registry, name string) string {
	return fmt.Sprintf("%s/%s", registry, name)
}

// ParseRegistryPath parses a registry path into registry type and repository name
func ParseRegistryPath(path string) (registry, repo string, err error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return "", "", errors.InvalidInputf("invalid format. Use [registry]/[repository]")
	}
	return parts[0], parts[1], nil
}
