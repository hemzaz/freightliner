package common

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// testMockRepository is a mock implementation of the Repository interface for testing
type testMockRepository struct {
	name        string
	tags        []string
	manifests   map[string]*interfaces.Manifest
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

func (r *testMockRepository) ListTags(ctx context.Context) ([]string, error) {
	if r.listError != nil {
		return nil, r.listError
	}
	return r.tags, nil
}

func (r *testMockRepository) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	if r.getError != nil {
		return nil, r.getError
	}

	manifest, ok := r.manifests[tag]
	if !ok {
		return nil, errors.New("manifest not found")
	}

	return manifest, nil
}

func (r *testMockRepository) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	if r.putError != nil {
		return r.putError
	}

	if r.manifests == nil {
		r.manifests = make(map[string]*interfaces.Manifest)
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

	if r.manifests == nil {
		return errors.New("manifest not found")
	}

	_, ok := r.manifests[tag]
	if !ok {
		return errors.New("manifest not found")
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
	if r.getError != nil {
		return nil, r.getError
	}

	if r.layers == nil {
		return nil, errors.New("layer not found")
	}

	content, ok := r.layers[digest]
	if !ok {
		return nil, errors.New("layer not found")
	}

	return io.NopCloser(strings.NewReader(content)), nil
}

func (r *testMockRepository) GetImageReference(tag string) (name.Reference, error) {
	// Create a simple reference
	ref, err := name.NewTag("example.com/" + r.name + ":" + tag)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func (r *testMockRepository) GetRemoteOptions() ([]remote.Option, error) {
	return []remote.Option{}, nil
}

func (r *testMockRepository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	return nil, errors.New("not implemented in test mock")
}

func (r *testMockRepository) PutImage(ctx context.Context, tag string, img v1.Image) error {
	return errors.New("not implemented in test mock")
}

// testMockRegistry is a mock implementation of the RegistryClient interface for testing
type testMockRegistry struct {
	repositories map[string]interfaces.Repository
	listError    error
	getError     error
}

func (r *testMockRegistry) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	if r.listError != nil {
		return nil, r.listError
	}

	var result []string
	for name := range r.repositories {
		if prefix == "" || strings.HasPrefix(name, prefix) {
			result = append(result, name)
		}
	}

	return result, nil
}

func (r *testMockRegistry) GetRepository(ctx context.Context, name string) (interfaces.Repository, error) {
	if r.getError != nil {
		return nil, r.getError
	}

	repo, ok := r.repositories[name]
	if !ok {
		return nil, errors.New("repository not found")
	}

	return repo, nil
}

func (r *testMockRegistry) GetRegistryName() string {
	return "test-registry"
}

func TestMockRegistry(t *testing.T) {
	// Create test repositories
	repo1 := &testMockRepository{
		name: "repo1",
		tags: []string{"latest"},
	}
	repo2 := &testMockRepository{
		name: "repo2",
		tags: []string{"v1.0", "v2.0"},
	}

	registry := &testMockRegistry{
		repositories: map[string]interfaces.Repository{
			"repo1": repo1,
			"repo2": repo2,
		},
	}

	// Test GetRepository
	repo, err := registry.GetRepository(context.Background(), "repo1")
	if err != nil {
		t.Errorf("GetRepository returned unexpected error: %v", err)
	}
	if repo.GetName() != "repo1" {
		t.Errorf("Expected repository name to be 'repo1', got '%s'", repo.GetName())
	}

	// Test ListRepositories
	repos, err := registry.ListRepositories(context.Background(), "")
	if err != nil {
		t.Errorf("ListRepositories returned unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(repos))
	}

	// Test ListRepositories with prefix
	repos, err = registry.ListRepositories(context.Background(), "repo1")
	if err != nil {
		t.Errorf("ListRepositories returned unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(repos))
	}
	if repos[0] != "repo1" {
		t.Errorf("Expected repository name to be 'repo1', got '%s'", repos[0])
	}
}
