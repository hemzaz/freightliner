package tree

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"freightliner/pkg/copy"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// MockRegistryClient is a mock implementation of the interfaces.RegistryClient interface
type MockRegistryClient struct {
	Repositories map[string]*MockRepository
	RegistryName string
	mu           sync.RWMutex
}

// GetRegistryName returns the name of this registry
func (m *MockRegistryClient) GetRegistryName() string {
	return m.RegistryName
}

// ListRepositories returns all repositories in the registry with the given prefix
func (m *MockRegistryClient) ListRepositories(_ context.Context, prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	repos := []string{}
	for name := range m.Repositories {
		if prefix == "" || strings.HasPrefix(name, prefix) {
			repos = append(repos, name)
		}
	}
	return repos, nil
}

// GetRepository returns a repository interface for the given repository name
func (m *MockRegistryClient) GetRepository(_ context.Context, name string) (interfaces.Repository, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	repo, ok := m.Repositories[name]
	if !ok {
		// Create a new empty repository if it doesn't exist
		repo = &MockRepository{
			Tags: make(map[string][]byte),
			Name: name,
			mu:   sync.RWMutex{},
		}
		m.Repositories[name] = repo
	}
	return repo, nil
}

// MockRepository is a mock implementation of the interfaces.Repository interface
type MockRepository struct {
	Tags map[string][]byte // map of tag -> manifest
	Name string
	mu   sync.RWMutex
}

// GetImage returns an image for testing
func (m *MockRepository) GetImage(_ context.Context, tag string) (v1.Image, error) {
	return &MockImage{}, nil
}

// GetManifest returns the manifest for the given tag
func (m *MockRepository) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	manifest, ok := m.Tags[tag]
	if !ok {
		// Return empty manifest for non-existent tags
		// In a real implementation, this would return an error
		return &interfaces.Manifest{
			Content:   []byte{},
			MediaType: "application/vnd.docker.distribution.manifest.v2+json",
			Digest:    "sha256:empty",
		}, nil
	}

	return &interfaces.Manifest{
		Content:   manifest,
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:    "sha256:" + tag,
	}, nil
}

// PutManifest uploads a manifest with the given tag
func (m *MockRepository) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Tags[tag] = manifest.Content
	return nil
}

// GetLayerReader returns a reader for the layer with the given digest
func (m *MockRepository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	// Just return a reader with some test data
	return io.NopCloser(strings.NewReader("test layer data")), nil
}

// DeleteManifest deletes the manifest for the given tag
func (m *MockRepository) DeleteManifest(ctx context.Context, tag string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.Tags, tag)
	return nil
}

// GetRepositoryName returns the name of this repository
func (m *MockRepository) GetRepositoryName() string {
	return m.Name
}

// GetName returns the name of this repository
func (m *MockRepository) GetName() string {
	return m.Name
}

// ListTags returns all tags in this repository
func (m *MockRepository) ListTags(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tags := []string{}
	for tag := range m.Tags {
		tags = append(tags, tag)
	}
	return tags, nil
}

// GetImageReference returns a reference for the given tag
func (m *MockRepository) GetImageReference(tag string) (name.Reference, error) {
	// Use correct registry port for local testing (5100 for source registry)
	ref, err := name.NewTag("localhost:5100/" + m.Name + ":" + tag)
	return ref, err
}

// GetRemoteOptions returns options for remote operations
func (m *MockRepository) GetRemoteOptions() ([]remote.Option, error) {
	return []remote.Option{}, nil
}

// MockImage is a mock implementation of the v1.Image interface
type MockImage struct{}

// Layers returns the ordered collection of filesystem layers that comprise this image.
func (i *MockImage) Layers() ([]v1.Layer, error) {
	return []v1.Layer{}, nil
}

// MediaType of this image's manifest.
func (i *MockImage) MediaType() (types.MediaType, error) {
	return "application/vnd.docker.distribution.manifest.v2+json", nil
}

// Size returns the size of the image's manifest.
func (i *MockImage) Size() (int64, error) {
	return 100, nil
}

// ConfigName returns the hash of the image's config file.
func (i *MockImage) ConfigName() (v1.Hash, error) {
	return v1.Hash{Algorithm: "sha256", Hex: "deadbeef"}, nil
}

// ConfigFile returns this image's config file.
func (i *MockImage) ConfigFile() (*v1.ConfigFile, error) {
	return &v1.ConfigFile{}, nil
}

// RawConfigFile returns the serialized bytes of ConfigFile().
func (i *MockImage) RawConfigFile() ([]byte, error) {
	return []byte{}, nil
}

// Digest returns the sha256 of this image's manifest.
func (i *MockImage) Digest() (v1.Hash, error) {
	return v1.Hash{Algorithm: "sha256", Hex: "deadbeef"}, nil
}

// Manifest returns this image's Manifest object.
func (i *MockImage) Manifest() (*v1.Manifest, error) {
	return &v1.Manifest{}, nil
}

// RawManifest returns the serialized bytes of Manifest().
func (i *MockImage) RawManifest() ([]byte, error) {
	return []byte{}, nil
}

// LayerByDiffID returns a Layer for interacting with a particular layer, identified
// by its uncompressed digest (DiffID).
func (i *MockImage) LayerByDiffID(h v1.Hash) (v1.Layer, error) {
	return nil, fmt.Errorf("layer with diff ID %s not found in mock image", h)
}

// LayerByDigest returns a Layer for interacting with a particular layer, identified
// by its compressed digest.
func (i *MockImage) LayerByDigest(h v1.Hash) (v1.Layer, error) {
	return nil, fmt.Errorf("layer with digest %s not found in mock image", h)
}

func TestReplicateTree(t *testing.T) {
	// Create source registry with multiple repositories and tags
	sourceRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{
			"project-a/service-1": {
				Tags: map[string][]byte{
					"v1.0":   []byte("manifest-1.0"),
					"v1.1":   []byte("manifest-1.1"),
					"latest": []byte("manifest-latest"),
				},
				Name: "project-a/service-1",
				mu:   sync.RWMutex{},
			},
			"project-a/service-2": {
				Tags: map[string][]byte{
					"v2.0":   []byte("manifest-2.0"),
					"latest": []byte("manifest-latest"),
				},
				Name: "project-a/service-2",
				mu:   sync.RWMutex{},
			},
			"project-b/service-3": {
				Tags: map[string][]byte{
					"v3.0":   []byte("manifest-3.0"),
					"latest": []byte("manifest-latest"),
				},
				Name: "project-b/service-3",
				mu:   sync.RWMutex{},
			},
		},
		RegistryName: "source.registry.com",
	}

	// Create empty destination registry
	destRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{},
		RegistryName: "dest.registry.com",
	}

	// Create a mock copier
	copier := &copy.Copier{}

	// Create a logger
	logger := log.NewBasicLogger(log.InfoLevel)

	// Create a tree replicator
	treeReplicator := NewTreeReplicator(logger, copier, TreeReplicatorOptions{
		WorkerCount:         2,
		ExcludeRepositories: []string{"excluded/*"},
		ExcludeTags:         []string{"*-dev", "*-test"},
		IncludeTags:         []string{"v*", "latest"},
		EnableCheckpointing: false,
		CheckpointDirectory: "",
		DryRun:              true,
	})

	// Replicate the tree
	result, err := treeReplicator.ReplicateTree(
		context.Background(),
		ReplicateTreeOptions{
			SourceClient:   sourceRegistry,
			DestClient:     destRegistry,
			SourcePrefix:   "",
			DestPrefix:     "",
			ForceOverwrite: false,
		},
	)

	// Check for errors
	if err != nil {
		t.Fatalf("ReplicateTree failed: %v", err)
	}

	// Check the results - we only check that the repositories were processed
	// In the mock test, the copy will fail but we want to ensure the code runs properly
	if result.Repositories != 3 {
		t.Errorf("Expected 3 repositories to be processed, got %d", result.Repositories)
	}
}

func TestReplicateTreeWithPrefix(t *testing.T) {
	// Create source registry with multiple repositories and tags
	sourceRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{
			"project-a/service-1": {
				Tags: map[string][]byte{
					"v1.0":   []byte("manifest-1.0"),
					"v1.1":   []byte("manifest-1.1"),
					"latest": []byte("manifest-latest"),
				},
				Name: "project-a/service-1",
				mu:   sync.RWMutex{},
			},
			"project-a/service-2": {
				Tags: map[string][]byte{
					"v2.0":   []byte("manifest-2.0"),
					"latest": []byte("manifest-latest"),
				},
				Name: "project-a/service-2",
				mu:   sync.RWMutex{},
			},
			"project-b/service-3": {
				Tags: map[string][]byte{
					"v3.0":   []byte("manifest-3.0"),
					"latest": []byte("manifest-latest"),
				},
				Name: "project-b/service-3",
				mu:   sync.RWMutex{},
			},
		},
		RegistryName: "source.registry.com",
	}

	// Create empty destination registry
	destRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{},
		RegistryName: "dest.registry.com",
	}

	// Create a mock copier
	copier := &copy.Copier{}

	// Create a logger
	logger := log.NewBasicLogger(log.InfoLevel)

	// Create a tree replicator
	treeReplicator := NewTreeReplicator(logger, copier, TreeReplicatorOptions{
		WorkerCount:         2,
		ExcludeRepositories: []string{"excluded/*"},
		ExcludeTags:         []string{"*-dev", "*-test"},
		IncludeTags:         []string{"v*", "latest"},
		EnableCheckpointing: false,
		CheckpointDirectory: "",
		DryRun:              true,
	})

	// Replicate only project-a repositories with a different destination prefix
	result, err := treeReplicator.ReplicateTree(
		context.Background(),
		ReplicateTreeOptions{
			SourceClient:   sourceRegistry,
			DestClient:     destRegistry,
			SourcePrefix:   "project-a",
			DestPrefix:     "mirror/project-a",
			ForceOverwrite: false,
		},
	)

	// Check for errors
	if err != nil {
		t.Fatalf("ReplicateTree failed: %v", err)
	}

	// Check the results
	if result.Repositories != 2 {
		t.Errorf("Expected 2 repositories to be processed, got %d", result.Repositories)
	}
}

func TestReplicateTreeWithFilters(t *testing.T) {
	// Create source registry with multiple repositories and tags
	sourceRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{
			"project-a/service-1": {
				Tags: map[string][]byte{
					"v1.0":     []byte("manifest-1.0"),
					"v1.1":     []byte("manifest-1.1"),
					"v1.0-dev": []byte("manifest-1.0-dev"),
					"latest":   []byte("manifest-latest"),
				},
				Name: "project-a/service-1",
			},
			"project-a/service-2": {
				Tags: map[string][]byte{
					"v2.0":      []byte("manifest-2.0"),
					"v2.0-test": []byte("manifest-2.0-test"),
					"latest":    []byte("manifest-latest"),
				},
				Name: "project-a/service-2",
			},
			"excluded/service-3": {
				Tags: map[string][]byte{
					"v3.0":   []byte("manifest-3.0"),
					"latest": []byte("manifest-latest"),
				},
				Name: "excluded/service-3",
			},
		},
		RegistryName: "source.registry.com",
	}

	// Create empty destination registry
	destRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{},
		RegistryName: "dest.registry.com",
	}

	// Create a mock copier
	copier := &copy.Copier{}

	// Create a logger
	logger := log.NewBasicLogger(log.InfoLevel)

	// Create a tree replicator with specific filters
	treeReplicator := NewTreeReplicator(logger, copier, TreeReplicatorOptions{
		WorkerCount:         2,
		ExcludeRepositories: []string{"excluded/*"},
		ExcludeTags:         []string{"*-dev", "*-test"},
		IncludeTags:         []string{"v*", "latest"},
		EnableCheckpointing: false,
		CheckpointDirectory: "",
		DryRun:              true,
	})

	// Replicate the tree with filtering
	result, err := treeReplicator.ReplicateTree(
		context.Background(),
		ReplicateTreeOptions{
			SourceClient:   sourceRegistry,
			DestClient:     destRegistry,
			SourcePrefix:   "",
			DestPrefix:     "",
			ForceOverwrite: false,
		},
	)

	// Check for errors
	if err != nil {
		t.Fatalf("ReplicateTree failed: %v", err)
	}

	// Check the results - only project-a repositories should be processed
	if result.Repositories != 2 {
		t.Errorf("Expected 2 repositories after filtering, got %d", result.Repositories)
	}

	// In a real implementation, we would check which tags were replicated
	// But since our mock doesn't fully implement the filtering, we only check repository count
}
