package common

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound indicates a resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthorized indicates an authentication failure
	ErrUnauthorized = errors.New("unauthorized")

	// ErrRateLimit indicates a rate limit was hit
	ErrRateLimit = errors.New("rate limit exceeded")
)

// RegistryError wraps an error with registry-specific information
type RegistryError struct {
	Registry string
	Original error
}

func (e *RegistryError) Error() string {
	return fmt.Sprintf("%s: %v", e.Registry, e.Original)
}

func (e *RegistryError) Unwrap() error {
	return e.Original
}

// NewRegistryError creates a new RegistryError with the given message and original error
func NewRegistryError(message string, originalErr error) error {
	return &RegistryError{
		Registry: message,
		Original: originalErr,
	}
}
