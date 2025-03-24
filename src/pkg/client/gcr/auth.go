package gcr

import (
	"context"
	"fmt"
	"strings"
	
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GCRAuthenticator implements the go-containerregistry authn.Authenticator interface for GCR
type GCRAuthenticator struct {
	ts oauth2.TokenSource
}

// NewGCRAuthenticator creates a new authenticator for GCR
func NewGCRAuthenticator(ts oauth2.TokenSource) *GCRAuthenticator {
	return &GCRAuthenticator{
		ts: ts,
	}
}

// Authorization returns the authorization for GCR
func (a *GCRAuthenticator) Authorization() (*authn.AuthConfig, error) {
	ctx := context.Background()
	
	// Get a token from the token source
	token, err := a.ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP token: %w", err)
	}
	
	// OAuth2 tokens are used as the username for GCR
	return &authn.AuthConfig{
		Username: "oauth2accesstoken",
		Password: token.AccessToken,
	}, nil
}

// GCRKeychain implements the go-containerregistry authn.Keychain interface for GCR
type GCRKeychain struct {
	ts oauth2.TokenSource
}

// NewGCRKeychain creates a new keychain for GCR
func NewGCRKeychain() (*GCRKeychain, error) {
	ctx := context.Background()
	
	// Get default token source for GCP
	ts, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to get token source: %w", err)
	}
	
	return &GCRKeychain{
		ts: ts,
	}, nil
}

// Resolve returns an authenticator for the given resource or an error
func (k *GCRKeychain) Resolve(target authn.Resource) (authn.Authenticator, error) {
	registry := target.RegistryStr()
	
	// Check if this is a GCR registry
	if !isGCRRegistry(registry) {
		return authn.Anonymous, nil
	}
	
	return NewGCRAuthenticator(k.ts), nil
}

// isGCRRegistry checks if the registry is a GCR registry
func isGCRRegistry(registry string) bool {
	gcrDomains := []string{
		"gcr.io",
		"us.gcr.io",
		"eu.gcr.io",
		"asia.gcr.io",
		"pkg.dev",  // Artifact Registry domains
	}
	
	for _, domain := range gcrDomains {
		if strings.HasSuffix(registry, domain) || registry == domain {
			return true
		}
	}
	
	return false
}

// ParseGCRRepository parses a GCR repository name into registry and repository components
func ParseGCRRepository(repoName string) (name.Registry, name.Repository, error) {
	// If it doesn't have a registry prefix, assume it's gcr.io
	if !strings.Contains(repoName, "/") || !strings.Contains(repoName, ".") {
		fullRepo := fmt.Sprintf("gcr.io/%s", repoName)
		repo, err := name.NewRepository(fullRepo)
		if err != nil {
			return name.Registry{}, name.Repository{}, fmt.Errorf("failed to parse GCR repository name: %w", err)
		}
		
		return repo.Registry, repo, nil
	}
	
	// If it already has the registry, just parse it
	repo, err := name.NewRepository(repoName)
	if err != nil {
		return name.Registry{}, name.Repository{}, fmt.Errorf("failed to parse GCR repository name: %w", err)
	}
	
	return repo.Registry, repo, nil
}