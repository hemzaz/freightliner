package tree

import (
	"context"
	"testing"
	"time"

	"github.com/hemzaz/freightliner/src/internal/log"
	"github.com/hemzaz/freightliner/src/pkg/client/common"
	"github.com/hemzaz/freightliner/src/pkg/copy"
)

// MockRegistryClient is a mock implementation of the common.RegistryClient interface
type MockRegistryClient struct {
	Repositories map[string]*MockRepository
}

// ListRepositories returns all repositories in the registry
func (m *MockRegistryClient) ListRepositories() ([]string, error) {
	var repos []string
	for repo := range m.Repositories {
		repos = append(repos, repo)
	}
	return repos, nil
}

// GetRepository returns a repository interface for the given repository name
func (m *MockRegistryClient) GetRepository(name string) (common.Repository, error) {
	repo, ok := m.Repositories[name]
	if !ok {
		// Create a new empty repository if it doesn't exist
		repo = &MockRepository{
			Tags: make(map[string][]byte),
		}
		m.Repositories[name] = repo
	}
	return repo, nil
}

// MockRepository is a mock implementation of the common.Repository interface
type MockRepository struct {
	Tags map[string][]byte // map of tag -> manifest
}

// ListTags returns all tags for the repository
func (m *MockRepository) ListTags() ([]string, error) {
	var tags []string
	for tag := range m.Tags {
		tags = append(tags, tag)
	}
	return tags, nil
}

// GetManifest returns the manifest for the given tag
func (m *MockRepository) GetManifest(tag string) ([]byte, string, error) {
	manifest, ok := m.Tags[tag]
	if !ok {
		// Return empty manifest for non-existent tags
		// In a real implementation, this would return an error
		return []byte{}, "application/vnd.docker.distribution.manifest.v2+json", nil
	}
	return manifest, "application/vnd.docker.distribution.manifest.v2+json", nil
}

// PutManifest uploads a manifest with the given tag
func (m *MockRepository) PutManifest(tag string, manifest []byte, mediaType string) error {
	m.Tags[tag] = manifest
	return nil
}

// DeleteManifest deletes the manifest for the given tag
func (m *MockRepository) DeleteManifest(tag string) error {
	delete(m.Tags, tag)
	return nil
}

// MockMetrics is a mock implementation of the metrics.Metrics interface
type MockMetrics struct {
	StartCount    int
	CompleteCount int
	FailCount     int
}

func (m *MockMetrics) ReplicationStarted(source, destination string) {
	m.StartCount++
}

func (m *MockMetrics) ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64) {
	m.CompleteCount++
}

func (m *MockMetrics) ReplicationFailed() {
	m.FailCount++
}

func TestReplicateTree(t *testing.T) {
	// Create source registry with multiple repositories and tags
	sourceRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{
			"project-a/service-1": {
				Tags: map[string][]byte{
					"v1.0": []byte("manifest-1.0"),
					"v1.1": []byte("manifest-1.1"),
					"latest": []byte("manifest-latest"),
				},
			},
			"project-a/service-2": {
				Tags: map[string][]byte{
					"v2.0": []byte("manifest-2.0"),
					"latest": []byte("manifest-latest"),
				},
			},
			"project-b/service-3": {
				Tags: map[string][]byte{
					"v3.0": []byte("manifest-3.0"),
					"v3.1": []byte("manifest-3.1"),
					"latest": []byte("manifest-latest"),
				},
			},
		},
	}

	// Create destination registry (empty)
	destRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{},
	}

	// Create logger
	logger := log.NewLogger(log.InfoLevel)

	// Create copier
	copier := copy.NewCopier(logger)

	// Create tree replicator
	treeReplicator := NewTreeReplicator(logger, copier, TreeReplicatorOptions{
		WorkerCount: 2,
		ExcludeRepositories: []string{},
		ExcludeTags: []string{},
		IncludeTags: []string{},
		DryRun: false,
	})

	// Create mock metrics
	metrics := &MockMetrics{}
	treeReplicator.WithMetrics(metrics)

	// Replicate the tree
	result, err := treeReplicator.ReplicateTree(
		context.Background(),
		sourceRegistry,
		destRegistry,
		"",
		"",
		false,
	)

	// Check for errors
	if err != nil {
		t.Fatalf("ReplicateTree failed: %v", err)
	}

	// Check the results
	if result.Repositories != 3 {
		t.Errorf("Expected 3 repositories, got %d", result.Repositories)
	}

	if result.ImagesReplicated != 8 {
		t.Errorf("Expected 8 images replicated, got %d", result.ImagesReplicated)
	}

	// Check that repositories were created in destination
	destRepos, _ := destRegistry.ListRepositories()
	if len(destRepos) != 3 {
		t.Errorf("Expected 3 repositories in destination, got %d", len(destRepos))
	}

	// Check that tags were copied
	for srcRepoName, srcRepo := range sourceRegistry.Repositories {
		destRepo, _ := destRegistry.GetRepository(srcRepoName)
		
		srcTags, _ := srcRepo.ListTags()
		destTags, _ := destRepo.ListTags()
		
		if len(srcTags) != len(destTags) {
			t.Errorf("Repository %s: expected %d tags, got %d", srcRepoName, len(srcTags), len(destTags))
		}
		
		// Check that each manifest was copied correctly
		for tag, srcManifest := range srcRepo.Tags {
			if destRepo, ok := destRegistry.Repositories[srcRepoName]; ok {
				destManifest, ok := destRepo.Tags[tag]
				if !ok {
					t.Errorf("Repository %s: tag %s not found in destination", srcRepoName, tag)
				} else if string(srcManifest) != string(destManifest) {
					t.Errorf("Repository %s: tag %s manifest differs", srcRepoName, tag)
				}
			}
		}
	}
}

func TestReplicateTreeWithPrefixes(t *testing.T) {
	// Create source registry with multiple repositories and tags
	sourceRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{
			"project-a/service-1": {
				Tags: map[string][]byte{
					"v1.0": []byte("manifest-1.0"),
					"latest": []byte("manifest-latest"),
				},
			},
			"project-a/service-2": {
				Tags: map[string][]byte{
					"v2.0": []byte("manifest-2.0"),
				},
			},
			"project-b/service-3": {
				Tags: map[string][]byte{
					"v3.0": []byte("manifest-3.0"),
				},
			},
		},
	}

	// Create destination registry (empty)
	destRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{},
	}

	// Create logger
	logger := log.NewLogger(log.InfoLevel)

	// Create copier
	copier := copy.NewCopier(logger)

	// Create tree replicator
	treeReplicator := NewTreeReplicator(logger, copier, TreeReplicatorOptions{
		WorkerCount: 2,
		ExcludeRepositories: []string{},
		ExcludeTags: []string{},
		IncludeTags: []string{},
		DryRun: false,
	})

	// Replicate only project-a to mirror/project-a
	result, err := treeReplicator.ReplicateTree(
		context.Background(),
		sourceRegistry,
		destRegistry,
		"project-a",
		"mirror/project-a",
		false,
	)

	// Check for errors
	if err != nil {
		t.Fatalf("ReplicateTree failed: %v", err)
	}

	// Check the results
	if result.Repositories != 2 {
		t.Errorf("Expected 2 repositories, got %d", result.Repositories)
	}

	if result.ImagesReplicated != 3 {
		t.Errorf("Expected 3 images replicated, got %d", result.ImagesReplicated)
	}

	// Check that repositories were created in destination with correct names
	if _, ok := destRegistry.Repositories["mirror/project-a/service-1"]; !ok {
		t.Errorf("Expected repository mirror/project-a/service-1 to exist")
	}

	if _, ok := destRegistry.Repositories["mirror/project-a/service-2"]; !ok {
		t.Errorf("Expected repository mirror/project-a/service-2 to exist")
	}

	// Ensure project-b was not replicated
	if _, ok := destRegistry.Repositories["project-b/service-3"]; ok {
		t.Errorf("Repository project-b/service-3 should not have been replicated")
	}
}

func TestReplicateTreeWithFilters(t *testing.T) {
	// Create source registry with multiple repositories and tags
	sourceRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{
			"project-a/service-1": {
				Tags: map[string][]byte{
					"v1.0": []byte("manifest-1.0"),
					"v1.1": []byte("manifest-1.1"),
					"latest": []byte("manifest-latest"),
					"dev": []byte("manifest-dev"),
				},
			},
			"project-a/service-2": {
				Tags: map[string][]byte{
					"v2.0": []byte("manifest-2.0"),
					"latest": []byte("manifest-latest"),
					"dev": []byte("manifest-dev"),
				},
			},
			"project-b/service-3": {
				Tags: map[string][]byte{
					"v3.0": []byte("manifest-3.0"),
				},
			},
		},
	}

	// Create destination registry (empty)
	destRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{},
	}

	// Create logger
	logger := log.NewLogger(log.InfoLevel)

	// Create copier
	copier := copy.NewCopier(logger)

	// Create tree replicator with filters
	treeReplicator := NewTreeReplicator(logger, copier, TreeReplicatorOptions{
		WorkerCount: 2,
		ExcludeRepositories: []string{"*service-3"},
		ExcludeTags: []string{"dev"},
		IncludeTags: []string{"v*", "latest"},
		DryRun: false,
	})

	// Replicate the tree
	result, err := treeReplicator.ReplicateTree(
		context.Background(),
		sourceRegistry,
		destRegistry,
		"",
		"",
		false,
	)

	// Check for errors
	if err != nil {
		t.Fatalf("ReplicateTree failed: %v", err)
	}

	// Check the results - should exclude project-b/service-3 and dev tags
	if result.Repositories != 2 {
		t.Errorf("Expected 2 repositories, got %d", result.Repositories)
	}

	// Check that dev tags were excluded
	for srcRepoName, repo := range destRegistry.Repositories {
		tags, _ := repo.ListTags()
		for _, tag := range tags {
			if tag == "dev" {
				t.Errorf("Repository %s: tag 'dev' should have been excluded", srcRepoName)
			}
		}
	}

	// Check that project-b/service-3 was excluded
	if _, ok := destRegistry.Repositories["project-b/service-3"]; ok {
		t.Errorf("Repository project-b/service-3 should have been excluded")
	}
}

func TestMatchPattern(t *testing.T) {
	testCases := []struct {
		pattern string
		str     string
		match   bool
	}{
		// Exact matches
		{"foo", "foo", true},
		{"foo", "bar", false},
		
		// Wildcard patterns
		{"*", "anything", true},
		{"foo*", "foobar", true},
		{"foo*", "barfoo", false},
		{"*foo", "barfoo", true},
		{"*foo", "foobar", false},
		{"*foo*", "barfoobaz", true},
		{"v*", "v1.0", true},
		{"v*", "1.0", false},
		
		// Complex patterns
		{"v?.?", "v1.0", true},
		{"v?.?", "v12.0", false},
		{"project-*/service-?", "project-a/service-1", true},
		{"project-*/service-?", "other/service-1", false},
	}
	
	for _, tc := range testCases {
		result := matchPattern(tc.pattern, tc.str)
		if result != tc.match {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tc.pattern, tc.str, result, tc.match)
		}
	}
}