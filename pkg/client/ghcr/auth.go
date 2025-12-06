package ghcr

import (
	"encoding/base64"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
)

// Authenticator implements GHCR authentication using GitHub tokens
type Authenticator struct {
	token    string
	username string
}

// NewAuthenticator creates a new GHCR authenticator
func NewAuthenticator(token string, username string) *Authenticator {
	// If username is empty, use a default
	if username == "" {
		username = "USERNAME" // GHCR accepts any username with valid token
	}

	return &Authenticator{
		token:    token,
		username: username,
	}
}

// NewAuthenticatorFromEnv creates a new GHCR authenticator from environment variables
func NewAuthenticatorFromEnv() *Authenticator {
	token := ""
	username := ""

	// Try various environment variables
	if ghToken := os.Getenv("GITHUB_TOKEN"); ghToken != "" {
		token = ghToken
	} else if ghToken := os.Getenv("GH_TOKEN"); ghToken != "" {
		token = ghToken
	} else if ghcrToken := os.Getenv("GHCR_TOKEN"); ghcrToken != "" {
		token = ghcrToken
	}

	// Try to get username from environment
	if ghUser := os.Getenv("GITHUB_USERNAME"); ghUser != "" {
		username = ghUser
	} else if ghUser := os.Getenv("GITHUB_ACTOR"); ghUser != "" {
		// GitHub Actions sets GITHUB_ACTOR
		username = ghUser
	}

	return NewAuthenticator(token, username)
}

// Authorization returns the authorization configuration
func (a *Authenticator) Authorization() (*authn.AuthConfig, error) {
	if a.token == "" {
		// Anonymous access
		return &authn.AuthConfig{}, nil
	}

	// GHCR uses the token as password with any username
	return &authn.AuthConfig{
		Username: a.username,
		Password: a.token,
	}, nil
}

// GetToken returns the GitHub token
func (a *Authenticator) GetToken() string {
	return a.token
}

// GetAuthHeader returns the base64-encoded auth header value
func (a *Authenticator) GetAuthHeader() string {
	if a.token == "" {
		return ""
	}

	auth := a.username + ":" + a.token
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// IsAuthenticated returns true if a token is configured
func (a *Authenticator) IsAuthenticated() bool {
	return a.token != ""
}
