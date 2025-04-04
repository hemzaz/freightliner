// Package auth provides authentication utilities for container registry clients
package auth

// Authenticator defines the interface for authenticating with container registries
type Authenticator interface {
	// GetToken returns an authentication token for the specified registry
	GetToken() (string, error)
}
