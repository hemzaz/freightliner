package common

import (
	"errors"
	"testing"
)

type mockRegistry struct {
	repositories map[string]Repository
	listError    error
}

func (m *mockRegistry) GetRepository(name string) (Repository, error) {
	if m.repositories == nil {
		return nil, errors.New("mock registry not initialized")
	}

	repo, exists := m.repositories[name]
	if !exists {
		return nil, &RegistryError{Registry: "mock", Original: ErrNotFound}
	}

	return repo, nil
}

func (m *mockRegistry) ListRepositories() ([]string, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	if m.repositories == nil {
		return []string{}, nil
	}

	repos := make([]string, 0, len(m.repositories))
	for name := range m.repositories {
		repos = append(repos, name)
	}
	return repos, nil
}

type mockRepository struct {
	name      string
	tags      []string
	manifests map[string][]byte
	mediaType map[string]string
	listError error
	getError  error
	putError  error
	delError  error
}

func (m *mockRepository) GetRepositoryName() string {
	return m.name
}

func (m *mockRepository) ListTags() ([]string, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.tags, nil
}

func (m *mockRepository) GetManifest(tag string) ([]byte, string, error) {
	if m.getError != nil {
		return nil, "", m.getError
	}

	manifest, exists := m.manifests[tag]
	if !exists {
		return nil, "", &RegistryError{Registry: "mock", Original: ErrNotFound}
	}

	mediaType, exists := m.mediaType[tag]
	if !exists {
		mediaType = "application/vnd.docker.distribution.manifest.v2+json"
	}

	return manifest, mediaType, nil
}

func (m *mockRepository) PutManifest(tag string, manifest []byte, mediaType string) error {
	if m.putError != nil {
		return m.putError
	}

	if m.manifests == nil {
		m.manifests = make(map[string][]byte)
	}
	if m.mediaType == nil {
		m.mediaType = make(map[string]string)
	}

	m.manifests[tag] = manifest
	m.mediaType[tag] = mediaType
	
	// Add tag to tags list if not already there
	found := false
	for _, t := range m.tags {
		if t == tag {
			found = true
			break
		}
	}
	if !found {
		m.tags = append(m.tags, tag)
	}
	
	return nil
}

func (m *mockRepository) DeleteManifest(tag string) error {
	if m.delError != nil {
		return m.delError
	}

	if m.manifests == nil {
		return &RegistryError{Registry: "mock", Original: ErrNotFound}
	}

	_, exists := m.manifests[tag]
	if !exists {
		return &RegistryError{Registry: "mock", Original: ErrNotFound}
	}

	delete(m.manifests, tag)
	delete(m.mediaType, tag)
	
	// Remove tag from tags list
	for i, t := range m.tags {
		if t == tag {
			m.tags = append(m.tags[:i], m.tags[i+1:]...)
			break
		}
	}
	
	return nil
}

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