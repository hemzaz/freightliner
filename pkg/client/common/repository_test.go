package common

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// testMockRepository is a mock implementation of the Repository interface for testing
type testMockRepository struct {
	name        string
	tags        []string
	manifests   map[string]*Manifest
	layers      map[string]string // digest -> content
	listError   error
	getError    error
	putError    error
	deleteError error
}

func (r *testMockRepository) GetRepositoryName() string {
	return r.name
}

func (r *testMockRepository) GetName() string {
	return r.name
}

func (r *testMockRepository) ListTags() ([]string, error) {
	if r.listError != nil {
		return nil, r.listError
	}
	return r.tags, nil
}

func (r *testMockRepository) GetManifest(ctx context.Context, tag string) (*Manifest, error) {
	if r.getError != nil {
		return nil, r.getError
	}

	manifest, ok := r.manifests[tag]
	if !ok {
		return nil, ErrNotFound
	}

	return manifest, nil
}

func (r *testMockRepository) PutManifest(ctx context.Context, tag string, manifest *Manifest) error {
	if r.putError != nil {
		return r.putError
	}

	if r.manifests == nil {
		r.manifests = make(map[string]*Manifest)
	}

	r.manifests[tag] = manifest

	// Add tag to tags list if not already present
	tagExists := false
	for _, t := range r.tags {
		if t == tag {
			tagExists = true
			break
		}
	}

	if !tagExists {
		r.tags = append(r.tags, tag)
	}

	return nil
}

func (r *testMockRepository) DeleteManifest(ctx context.Context, tag string) error {
	if r.deleteError != nil {
		return r.deleteError
	}

	if _, ok := r.manifests[tag]; !ok {
		return ErrNotFound
	}

	delete(r.manifests, tag)

	// Remove tag from tags list
	for i, t := range r.tags {
		if t == tag {
			r.tags = append(r.tags[:i], r.tags[i+1:]...)
			break
		}
	}

	return nil
}

func (r *testMockRepository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	content, ok := r.layers[digest]
	if !ok {
		return nil, ErrNotFound
	}

	return io.NopCloser(strings.NewReader(content)), nil
}

func (r *testMockRepository) GetImageReference(tag string) (name.Reference, error) {
	return name.NewTag("example.com/repo:" + tag)
}

func (r *testMockRepository) GetRemoteOptions() ([]remote.Option, error) {
	return []remote.Option{}, nil
}

// testMockRegistry is a mock implementation of the RegistryClient interface for testing
type testMockRegistry struct {
	repositories map[string]Repository
	listError    error
	getError     error
}

func (r *testMockRegistry) ListRepositories() ([]string, error) {
	if r.listError != nil {
		return nil, r.listError
	}

	var repos []string
	for name := range r.repositories {
		repos = append(repos, name)
	}

	return repos, nil
}

func (r *testMockRegistry) GetRepository(name string) (Repository, error) {
	if r.getError != nil {
		return nil, r.getError
	}

	repo, ok := r.repositories[name]
	if !ok {
		return nil, ErrNotFound
	}

	return repo, nil
}

// TestMockRepository tests that our mock repository correctly implements the Repository interface
func TestMockRepository(t *testing.T) {
	// Create test manifests
	latestManifest := &Manifest{
		Content:   []byte("manifest-latest"),
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:    "sha256:latest",
	}

	v1Manifest := &Manifest{
		Content:   []byte("manifest-v1.0"),
		MediaType: "application/vnd.oci.image.manifest.v1+json",
		Digest:    "sha256:v1.0",
	}

	repo := &testMockRepository{
		name: "test-repo",
		tags: []string{"latest", "v1.0"},
		manifests: map[string]*Manifest{
			"latest": latestManifest,
			"v1.0":   v1Manifest,
		},
		layers: map[string]string{
			"sha256:layer1": "layer1-content",
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
	ctx := context.Background()
	manifest, err := repo.GetManifest(ctx, "latest")
	if err != nil {
		t.Errorf("GetManifest returned unexpected error: %v", err)
	}
	if string(manifest.Content) != "manifest-latest" {
		t.Errorf("Expected manifest content to be 'manifest-latest', got '%s'", string(manifest.Content))
	}
	if manifest.MediaType != "application/vnd.docker.distribution.manifest.v2+json" {
		t.Errorf("Expected media type to be 'application/vnd.docker.distribution.manifest.v2+json', got '%s'", manifest.MediaType)
	}

	// Test PutManifest
	newManifest := &Manifest{
		Content:   []byte("manifest-v2.0"),
		MediaType: "application/vnd.oci.image.manifest.v1+json",
		Digest:    "sha256:v2.0",
	}
	err = repo.PutManifest(ctx, "v2.0", newManifest)
	if err != nil {
		t.Errorf("PutManifest returned unexpected error: %v", err)
	}

	manifest, err = repo.GetManifest(ctx, "v2.0")
	if err != nil {
		t.Errorf("GetManifest returned unexpected error: %v", err)
	}
	if string(manifest.Content) != "manifest-v2.0" {
		t.Errorf("Expected manifest content to be 'manifest-v2.0', got '%s'", string(manifest.Content))
	}
	if manifest.MediaType != "application/vnd.oci.image.manifest.v1+json" {
		t.Errorf("Expected media type to be 'application/vnd.oci.image.manifest.v1+json', got '%s'", manifest.MediaType)
	}

	// Test DeleteManifest
	err = repo.DeleteManifest(ctx, "v1.0")
	if err != nil {
		t.Errorf("DeleteManifest returned unexpected error: %v", err)
	}
	_, err = repo.GetManifest(ctx, "v1.0")
	if err == nil {
		t.Error("GetManifest should have returned an error for deleted tag")
	}

	// Test GetLayerReader
	reader, err := repo.GetLayerReader(ctx, "sha256:layer1")
	if err != nil {
		t.Errorf("GetLayerReader returned unexpected error: %v", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("Failed to read layer content: %v", err)
	}
	if string(content) != "layer1-content" {
		t.Errorf("Expected layer content to be 'layer1-content', got '%s'", string(content))
	}
}

// TestMockRegistry tests that our mock registry correctly implements the RegistryClient interface
func TestMockRegistry(t *testing.T) {
	repo1 := &testMockRepository{
		name: "repo1",
		tags: []string{"latest"},
	}
	repo2 := &testMockRepository{
		name: "repo2",
		tags: []string{"v1.0", "v2.0"},
	}

	registry := &testMockRegistry{
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
	ctx := context.Background()

	// Test list error
	listErrRepo := &testMockRepository{
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
	getErrRepo := &testMockRepository{
		name:     "error-repo",
		getError: &RegistryError{Registry: "get error", Original: ErrUnauthorized},
	}
	_, err = getErrRepo.GetManifest(ctx, "tag")
	if err == nil {
		t.Error("GetManifest should have returned an error")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("Expected ErrUnauthorized, got different error: %v", err)
	}

	// Test non-existent tag
	noTagRepo := &testMockRepository{
		name:      "no-tag-repo",
		manifests: map[string]*Manifest{},
	}
	_, err = noTagRepo.GetManifest(ctx, "non-existent")
	if err == nil {
		t.Error("GetManifest should have returned an error for non-existent tag")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got different error: %v", err)
	}
}
