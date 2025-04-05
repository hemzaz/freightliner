package interfaces

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
)

// RegistryAuthenticator defines the interface for registry authentication
type RegistryAuthenticator interface {
	// GetAuthToken returns an authentication token for the registry
	GetAuthToken(ctx context.Context, registry string) (string, error)

	// GetAuthHeader returns an authentication header for the registry
	GetAuthHeader(ctx context.Context, registry string) (string, error)

	// GetAuthenticator returns an authn.Authenticator for the registry
	GetAuthenticator(ctx context.Context, registry string) (authn.Authenticator, error)
}
