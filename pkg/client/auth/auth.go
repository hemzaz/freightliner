package auth

import (
	"context"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
)

// TokenCache provides caching for registry tokens
type TokenCache struct {
	tokens     map[string]CachedToken
	mutex      sync.RWMutex
	defaultTTL time.Duration
}

// CachedToken holds token data with expiration
type CachedToken struct {
	Value     string
	Username  string
	ExpiresAt time.Time
}

// NewTokenCache creates a new token cache with default TTL
func NewTokenCache(defaultTTL time.Duration) *TokenCache {
	if defaultTTL <= 0 {
		defaultTTL = 30 * time.Minute
	}

	return &TokenCache{
		tokens:     make(map[string]CachedToken),
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a token from cache
func (c *TokenCache) Get(key string) (string, string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	token, found := c.tokens[key]
	if !found || time.Now().After(token.ExpiresAt) {
		return "", "", false
	}

	return token.Value, token.Username, true
}

// Put stores a token in cache
func (c *TokenCache) Put(key, value, username string, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	c.tokens[key] = CachedToken{
		Value:     value,
		Username:  username,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// AuthProvider is a unified interface for registry authentication
type AuthProvider interface {
	GetToken(ctx context.Context, registry string) (string, string, error)
	GetAuthenticator(ctx context.Context, registry string) (authn.Authenticator, error)
}

// BaseAuthProvider implements common auth provider functionality
type BaseAuthProvider struct {
	Logger     *log.Logger
	TokenCache *TokenCache
	Registry   string
}

// NewBaseAuthProvider creates a new base auth provider
func NewBaseAuthProvider(logger *log.Logger, registry string) *BaseAuthProvider {
	return &BaseAuthProvider{
		Logger:     logger,
		TokenCache: NewTokenCache(30 * time.Minute),
		Registry:   registry,
	}
}

// GetToken retrieves an authentication token for the specified registry
func (p *BaseAuthProvider) GetToken(ctx context.Context, registry string) (string, string, error) {
	// Check cache first
	if token, username, found := p.TokenCache.Get(registry); found {
		p.Logger.Debug("Using cached token", map[string]interface{}{
			"registry": registry,
		})
		return token, username, nil
	}

	// Implement in specific providers
	return "", "", errors.NotImplementedf("authentication not implemented in base provider")
}

// GetAuthenticator returns an authenticator for the registry
func (p *BaseAuthProvider) GetAuthenticator(ctx context.Context, registry string) (authn.Authenticator, error) {
	if ctx == nil {
		return nil, errors.InvalidInputf("context cannot be nil")
	}

	if registry == "" {
		return nil, errors.InvalidInputf("registry cannot be empty")
	}

	token, username, err := p.GetToken(ctx, registry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token")
	}

	return &BasicAuthenticator{
		Username: username,
		Password: token,
	}, nil
}

// BasicAuthenticator implements authn.Authenticator
type BasicAuthenticator struct {
	Username string
	Password string
}

// Authorization returns the auth config
func (a *BasicAuthenticator) Authorization() (*authn.AuthConfig, error) {
	return &authn.AuthConfig{
		Username: a.Username,
		Password: a.Password,
	}, nil
}
