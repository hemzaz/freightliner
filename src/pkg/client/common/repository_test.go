package common

import (
	"errors"
	"testing"
)

// TestMockRepository tests that our mock repository correctly implements the Repository interface
func TestMockRepository(t *testing.T) {
	repo := &mockRepository{
		name:      "test-repo",
		tags:      []string{"latest", "v1.0"},
		manifests: map[string][]byte{
			"latest": []byte("manifest-latest"),
			"v1.0":   []byte("manifest-v1.0"),
		},
		mediaType: map[string]string{
			"latest": "application/vnd.docker.distribution.manifest.v2+json",
			"v1.0":   "application/vnd.oci.image.manifest.v1+json",
		},
	}

	// Test GetRepositoryName
	if name := repo.GetRepositoryName(); name != "test-repo" {
		t.Errorf("Expected repository name to be 'test-repo', got '%s'", name)
	}

	// Test ListTags
	tags, err := repo.ListTags()
	if err != nil {
		t.Errorf("ListTags returned unexpected error: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
	for _, tag := range []string{"latest", "v1.0"} {
		found := false
		for _, t := range tags {
			if t == tag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tag '%s' not found in tags list", tag)
		}
	}

	// Test GetManifest
	manifest, mediaType, err := repo.GetManifest("latest")
	if err != nil {
		t.Errorf("GetManifest returned unexpected error: %v", err)
	}
	if string(manifest) != "manifest-latest" {
		t.Errorf("Expected manifest content to be 'manifest-latest', got '%s'", string(manifest))
	}
	if mediaType != "application/vnd.docker.distribution.manifest.v2+json" {
		t.Errorf("Expected media type to be 'application/vnd.docker.distribution.manifest.v2+json', got '%s'", mediaType)
	}

	// Test PutManifest
	err = repo.PutManifest("v2.0", []byte("manifest-v2.0"), "application/vnd.oci.image.manifest.v1+json")
	if err != nil {
		t.Errorf("PutManifest returned unexpected error: %v", err)
	}
	manifest, mediaType, err = repo.GetManifest("v2.0")
	if err != nil {
		t.Errorf("GetManifest returned unexpected error: %v", err)
	}
	if string(manifest) != "manifest-v2.0" {
		t.Errorf("Expected manifest content to be 'manifest-v2.0', got '%s'", string(manifest))
	}
	if mediaType != "application/vnd.oci.image.manifest.v1+json" {
		t.Errorf("Expected media type to be 'application/vnd.oci.image.manifest.v1+json', got '%s'", mediaType)
	}

	// Test DeleteManifest
	err = repo.DeleteManifest("v1.0")
	if err != nil {
		t.Errorf("DeleteManifest returned unexpected error: %v", err)
	}
	_, _, err = repo.GetManifest("v1.0")
	if err == nil {
		t.Error("GetManifest should have returned an error for deleted tag")
	}
}

// TestMockRegistry tests that our mock registry correctly implements the RegistryClient interface
func TestMockRegistry(t *testing.T) {
	repo1 := &mockRepository{
		name: "repo1",
		tags: []string{"latest"},
	}
	repo2 := &mockRepository{
		name: "repo2",
		tags: []string{"v1.0", "v2.0"},
	}

	registry := &mockRegistry{
		repositories: map[string]Repository{
			"repo1": repo1,
			"repo2": repo2,
		},
	}

	// Test GetRepository
	repo, err := registry.GetRepository("repo1")
	if err != nil {
		t.Errorf("GetRepository returned unexpected error: %v", err)
	}
	if repo.GetRepositoryName() != "repo1" {
		t.Errorf("Expected repository name to be 'repo1', got '%s'", repo.GetRepositoryName())
	}

	// Test GetRepository for non-existent repo
	_, err = registry.GetRepository("repo3")
	if err == nil {
		t.Error("GetRepository should have returned an error for non-existent repository")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got different error: %v", err)
	}

	// Test ListRepositories
	repos, err := registry.ListRepositories()
	if err != nil {
		t.Errorf("ListRepositories returned unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(repos))
	}
	for _, repoName := range []string{"repo1", "repo2"} {
		found := false
		for _, r := range repos {
			if r == repoName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected repository '%s' not found in repositories list", repoName)
		}
	}
}

// Test error handling
func TestRepositoryErrorHandling(t *testing.T) {
	// Test list error
	listErrRepo := &mockRepository{
		name:      "error-repo",
		listError: &RegistryError{Registry: "list error", Original: ErrUnauthorized},
	}
	_, err := listErrRepo.ListTags()
	if err == nil {
		t.Error("ListTags should have returned an error")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("Expected ErrUnauthorized, got different error: %v", err)
	}

	// Test get error
	getErrRepo := &mockRepository{
		name:     "error-repo",
		getError: &RegistryError{Registry: "get error", Original: ErrUnauthorized},
	}
	_, _, err = getErrRepo.GetManifest("tag")
	if err == nil {
		t.Error("GetManifest should have returned an error")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("Expected ErrUnauthorized, got different error: %v", err)
	}

	// Test non-existent tag
	noTagRepo := &mockRepository{
		name:      "no-tag-repo",
		manifests: map[string][]byte{},
	}
	_, _, err = noTagRepo.GetManifest("non-existent")
	if err == nil {
		t.Error("GetManifest should have returned an error for non-existent tag")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got different error: %v", err)
	}
}