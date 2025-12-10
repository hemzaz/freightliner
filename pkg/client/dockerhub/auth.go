package dockerhub

import (
	"encoding/base64"

	"github.com/google/go-containerregistry/pkg/authn"
)

// Authenticator implements Docker Hub authentication
type Authenticator struct {
	username string
	password string
}

// NewAuthenticator creates a new Docker Hub authenticator
func NewAuthenticator(username, password string) *Authenticator {
	return &Authenticator{
		username: username,
		password: password,
	}
}

// Authorization returns the authorization configuration
func (a *Authenticator) Authorization() (*authn.AuthConfig, error) {
	if a.username == "" {
		// Anonymous access
		return &authn.AuthConfig{}, nil
	}

	return &authn.AuthConfig{
		Username: a.username,
		Password: a.password,
	}, nil
}

// GetToken returns a base64-encoded auth token
func (a *Authenticator) GetToken() string {
	if a.username == "" {
		return ""
	}

	auth := a.username + ":" + a.password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
