package common

import (
	"errors"
	"testing"
)

func TestRegistryErrorWrap(t *testing.T) {
	origErr := errors.New("original error")
	regErr := &RegistryError{Registry: "error description", Original: origErr}

	if regErr.Error() != "error description: original error" {
		t.Errorf("Expected error message to be 'error description: original error', got '%s'", regErr.Error())
	}

	unwrapped := errors.Unwrap(regErr)
	if unwrapped != origErr {
		t.Errorf("Unwrapping registry error did not return original error")
	}
}

func TestRegistryErrorIs(t *testing.T) {
	// Test ErrNotFound
	notFoundErr := &RegistryError{Registry: "not found", Original: ErrNotFound}
	if !errors.Is(notFoundErr, ErrNotFound) {
		t.Errorf("errors.Is should return true for ErrNotFound")
	}

	// Test ErrUnauthorized
	unauthorizedErr := &RegistryError{Registry: "unauthorized", Original: ErrUnauthorized}
	if !errors.Is(unauthorizedErr, ErrUnauthorized) {
		t.Errorf("errors.Is should return true for ErrUnauthorized")
	}

	// Test ErrRateLimit
	rateLimitErr := &RegistryError{Registry: "rate limit", Original: ErrRateLimit}
	if !errors.Is(rateLimitErr, ErrRateLimit) {
		t.Errorf("errors.Is should return true for ErrRateLimit")
	}

	// Test custom error
	customErr := errors.New("custom error")
	customWrappedErr := &RegistryError{Registry: "wrapped custom", Original: customErr}
	if !errors.Is(customWrappedErr, customErr) {
		t.Errorf("errors.Is should return true for wrapped custom error")
	}
}
