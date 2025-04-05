package common

import (
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
)

// BaseAuthenticator provides a foundation for registry authentication implementations
type BaseAuthenticator struct {
	// Cached credentials
	cachedAuth   authn.Authenticator
	cachedExpiry int64
}

// NewBaseAuthenticator creates a new base authenticator
func NewBaseAuthenticator() *BaseAuthenticator {
	return &BaseAuthenticator{}
}

// Authorization returns an auth header to authenticate a request
func (a *BaseAuthenticator) Authorization() (*authn.AuthConfig, error) {
	// This must be implemented by derived authenticators
	return nil, nil
}

// SetCachedAuth sets the cached authenticator and expiry time
func (a *BaseAuthenticator) SetCachedAuth(auth authn.Authenticator, expiryTime int64) {
	a.cachedAuth = auth
	a.cachedExpiry = expiryTime
}

// GetCachedAuth returns the cached authenticator and whether it's still valid
func (a *BaseAuthenticator) GetCachedAuth(currentTime int64) (authn.Authenticator, bool) {
	if a.cachedAuth == nil || (a.cachedExpiry > 0 && currentTime > a.cachedExpiry) {
		return nil, false
	}
	return a.cachedAuth, true
}

// ClearCachedAuth clears the cached authenticator
func (a *BaseAuthenticator) ClearCachedAuth() {
	a.cachedAuth = nil
	a.cachedExpiry = 0
}

// TransportWithAuth creates an HTTP transport with authentication
func TransportWithAuth(baseTransport http.RoundTripper, auth authn.Authenticator, resource authn.Resource) http.RoundTripper {
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	return &authnTransport{
		inner:    baseTransport,
		auth:     auth,
		resource: resource,
	}
}

// authnTransport is an HTTP transport that adds authentication to requests
type authnTransport struct {
	inner    http.RoundTripper
	auth     authn.Authenticator
	resource authn.Resource
}

// RoundTrip implements http.RoundTripper
func (t *authnTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	authConfig, err := t.auth.Authorization()
	if err != nil {
		return nil, err
	}

	if authConfig.Username != "" && authConfig.Password != "" {
		req.SetBasicAuth(authConfig.Username, authConfig.Password)
	} else if authConfig.Auth != "" {
		req.Header.Set("Authorization", "Basic "+authConfig.Auth)
	} else if authConfig.IdentityToken != "" {
		req.Header.Set("Authorization", "Bearer "+authConfig.IdentityToken)
	} else if authConfig.RegistryToken != "" {
		req.Header.Set("Authorization", "Bearer "+authConfig.RegistryToken)
	}

	return t.inner.RoundTrip(req)
}
