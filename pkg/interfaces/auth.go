package interfaces

import (
	"context"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
)

// ===== AUTHENTICATION INTERFACES WITH SEGREGATION =====

// TokenProvider provides authentication token access
type TokenProvider interface {
	// GetAuthToken returns an authentication token for the registry
	GetAuthToken(ctx context.Context, registry string) (string, error)
}

// HeaderProvider provides authentication header access
type HeaderProvider interface {
	// GetAuthHeader returns an authentication header for the registry
	GetAuthHeader(ctx context.Context, registry string) (string, error)
}

// AuthenticatorProvider provides go-containerregistry authenticator access
type AuthenticatorProvider interface {
	// GetAuthenticator returns an authn.Authenticator for the registry
	GetAuthenticator(ctx context.Context, registry string) (authn.Authenticator, error)
}

// RegistryAuthenticator defines the interface for registry authentication
// NOTE: This large interface is kept for backward compatibility.
// New code should use the more focused interfaces above.
type RegistryAuthenticator interface {
	TokenProvider
	HeaderProvider
	AuthenticatorProvider
}

// ===== ENHANCED AUTHENTICATION INTERFACES =====

// TokenManager provides advanced token management capabilities
type TokenManager interface {
	TokenProvider

	// RefreshToken refreshes an authentication token for the registry
	RefreshToken(ctx context.Context, registry string) (string, error)

	// ValidateToken validates an authentication token
	ValidateToken(ctx context.Context, registry string, token string) (bool, error)

	// GetTokenExpiry returns the expiry time of a token
	GetTokenExpiry(ctx context.Context, registry string, token string) (*time.Time, error)
}

// CachingAuthenticator provides caching capabilities for authentication
type CachingAuthenticator interface {
	RegistryAuthenticator

	// ClearCache clears the authentication cache for a registry
	ClearCache(ctx context.Context, registry string) error

	// ClearAllCache clears all authentication caches
	ClearAllCache(ctx context.Context) error

	// GetCacheStats returns cache statistics
	GetCacheStats(ctx context.Context) (*AuthCacheStats, error)
}

// AuthCacheStats provides statistics about authentication cache
type AuthCacheStats struct {
	TotalEntries  int
	HitCount      int64
	MissCount     int64
	EvictionCount int64
	LastAccess    *time.Time
}

// MultiRegistryAuthenticator provides authentication for multiple registries
type MultiRegistryAuthenticator interface {
	// GetAuthenticatorForRegistry returns an authenticator for a specific registry
	GetAuthenticatorForRegistry(ctx context.Context, registry string) (RegistryAuthenticator, error)

	// RegisterAuthenticator registers an authenticator for a registry pattern
	RegisterAuthenticator(ctx context.Context, pattern string, auth RegistryAuthenticator) error

	// UnregisterAuthenticator removes an authenticator for a registry pattern
	UnregisterAuthenticator(ctx context.Context, pattern string) error

	// ListRegistryPatterns returns all registered registry patterns
	ListRegistryPatterns(ctx context.Context) ([]string, error)
}

// ===== AUTHENTICATION COMPOSITION INTERFACES =====

// BasicAuth provides basic authentication capabilities
type BasicAuth interface {
	TokenProvider
	HeaderProvider
}

// AdvancedAuth provides advanced authentication capabilities
type AdvancedAuth interface {
	BasicAuth
	TokenManager
}

// FullAuth provides all authentication capabilities
type FullAuth interface {
	AdvancedAuth
	CachingAuthenticator
}

// AuthComposer provides composition of authentication behaviors
type AuthComposer interface {
	// AsTokenProvider returns a token provider view
	AsTokenProvider() TokenProvider

	// AsHeaderProvider returns a header provider view
	AsHeaderProvider() HeaderProvider

	// AsAuthenticatorProvider returns an authenticator provider view
	AsAuthenticatorProvider() AuthenticatorProvider

	// AsTokenManager returns a token manager view
	AsTokenManager() TokenManager

	// AsCachingAuthenticator returns a caching authenticator view
	AsCachingAuthenticator() CachingAuthenticator
}
