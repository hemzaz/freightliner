package replication

import (
	"context"
	"errors"
	"src/pkg/client/common"
	"src/pkg/copy"
	"src/pkg/metrics"
	"testing"
	"time"
)

// Mock registry client
type mockRegistryClient struct {
	repositories map[string]common.Repository
	listError    error
}

func (m *mockRegistryClient) GetRepository(name string) (common.Repository, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	repo, exists := m.repositories[name]
	if !exists {
		return nil, common.NewRegistryError("repository not found", common.ErrNotFound)
	}

	return repo, nil
}

func (m *mockRegistryClient) ListRepositories() ([]string, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	repos := make([]string, 0, len(m.repositories))
	for name := range m.repositories {
		repos = append(repos, name)
	}

	return repos, nil
}

// Mock repository
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
		return nil, "", common.NewRegistryError("manifest not found", common.ErrNotFound)
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
		return common.NewRegistryError("manifest not found", common.ErrNotFound)
	}

	_, exists := m.manifests[tag]
	if !exists {
		return common.NewRegistryError("manifest not found", common.ErrNotFound)
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

// Mock logger
type mockLogger struct {
	infoMessages  []string
	debugMessages []string
	warnMessages  []string
	errorMessages []string
	fields        []map[string]interface{}
}

func (m *mockLogger) Info(msg string, fields map[string]interface{}) {
	m.infoMessages = append(m.infoMessages, msg)
	m.fields = append(m.fields, fields)
}

func (m *mockLogger) Debug(msg string, fields map[string]interface{}) {
	m.debugMessages = append(m.debugMessages, msg)
	m.fields = append(m.fields, fields)
}

func (m *mockLogger) Warn(msg string, fields map[string]interface{}) {
	m.warnMessages = append(m.warnMessages, msg)
	m.fields = append(m.fields, fields)
}

func (m *mockLogger) Error(msg string, err error, fields map[string]interface{}) {
	m.errorMessages = append(m.errorMessages, msg)
	m.fields = append(m.fields, fields)
}

// Mock copier
type mockCopier struct {
	copyCount int
	copyError error
	lastSrc   string
	lastDest  string
	lastTag   string
}

func (m *mockCopier) CopyImage(ctx context.Context, sourceRepo common.Repository, destRepo common.Repository, tag string) error {
	m.copyCount++
	m.lastSrc = sourceRepo.GetRepositoryName()
	m.lastDest = destRepo.GetRepositoryName()
	m.lastTag = tag
	return m.copyError
}

func TestReconcileRepository(t *testing.T) {
	tests := []struct {
		name         string
		sourceRepo   mockRepository
		destRepo     mockRepository
		copyError    error
		expectCopies int
		expectErrors int
	}{
		{
			name: "New destination repository",
			sourceRepo: mockRepository{
				name: "source/repo",
				tags: []string{"latest", "v1.0", "v2.0"},
				manifests: map[string][]byte{
					"latest": []byte("manifest-latest"),
					"v1.0":   []byte("manifest-v1.0"),
					"v2.0":   []byte("manifest-v2.0"),
				},
			},
			destRepo: mockRepository{
				name:      "dest/repo",
				tags:      []string{},
				manifests: map[string][]byte{},
			},
			copyError:    nil,
			expectCopies: 3, // All 3 tags should be copied
			expectErrors: 0,
		},
		{
			name: "Partial sync needed",
			sourceRepo: mockRepository{
				name: "source/repo",
				tags: []string{"latest", "v1.0", "v2.0"},
				manifests: map[string][]byte{
					"latest": []byte("manifest-latest"),
					"v1.0":   []byte("manifest-v1.0"),
					"v2.0":   []byte("manifest-v2.0"),
				},
			},
			destRepo: mockRepository{
				name: "dest/repo",
				tags: []string{"latest", "v1.0"},
				manifests: map[string][]byte{
					"latest": []byte("manifest-latest"),
					"v1.0":   []byte("manifest-v1.0"),
				},
			},
			copyError:    nil,
			expectCopies: 1, // Only v2.0 should be copied
			expectErrors: 0,
		},
		{
			name: "Up to date, no copies needed",
			sourceRepo: mockRepository{
				name: "source/repo",
				tags: []string{"latest", "v1.0"},
				manifests: map[string][]byte{
					"latest": []byte("manifest-latest"),
					"v1.0":   []byte("manifest-v1.0"),
				},
			},
			destRepo: mockRepository{
				name: "dest/repo",
				tags: []string{"latest", "v1.0"},
				manifests: map[string][]byte{
					"latest": []byte("manifest-latest"),
					"v1.0":   []byte("manifest-v1.0"),
				},
			},
			copyError:    nil,
			expectCopies: 0, // Nothing should be copied
			expectErrors: 0,
		},
		{
			name: "Copy error",
			sourceRepo: mockRepository{
				name: "source/repo",
				tags: []string{"latest", "v1.0"},
				manifests: map[string][]byte{
					"latest": []byte("manifest-latest"),
					"v1.0":   []byte("manifest-v1.0"),
				},
			},
			destRepo: mockRepository{
				name:      "dest/repo",
				tags:      []string{},
				manifests: map[string][]byte{},
			},
			copyError:    errors.New("copy error"),
			expectCopies: 2, // 2 attempts
			expectErrors: 2, // 2 errors
		},
		{
			name: "List tags error",
			sourceRepo: mockRepository{
				name:      "source/repo",
				listError: errors.New("list error"),
			},
			destRepo: mockRepository{
				name: "dest/repo",
			},
			copyError:    nil,
			expectCopies: 0,
			expectErrors: 1, // Error listing tags
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock clients
			sourceClient := &mockRegistryClient{
				repositories: map[string]common.Repository{
					tt.sourceRepo.name: &tt.sourceRepo,
				},
			}

			destClient := &mockRegistryClient{
				repositories: map[string]common.Repository{
					tt.destRepo.name: &tt.destRepo,
				},
			}

			// Create mock logger
			logger := &mockLogger{}

			// Create mock copier
			copier := &mockCopier{
				copyError: tt.copyError,
			}

			// Create reconciler
			reconciler := NewReconciler(sourceClient, destClient, copier, metrics.NewNoopMetrics(), logger)

			// Run reconciliation
			ctx := context.Background()
			reconciler.ReconcileRepository(ctx, tt.sourceRepo.name, tt.destRepo.name)

			// Check results
			if copier.copyCount != tt.expectCopies {
				t.Errorf("Expected %d copy operations, got %d", tt.expectCopies, copier.copyCount)
			}

			if len(logger.errorMessages) != tt.expectErrors {
				t.Errorf("Expected %d error messages, got %d", tt.expectErrors, len(logger.errorMessages))
			}
		})
	}
}

func TestReconcileAllRepositories(t *testing.T) {
	// Create mock repositories
	sourceRepos := map[string]common.Repository{
		"project/repo1": &mockRepository{
			name: "project/repo1",
			tags: []string{"latest", "v1.0"},
			manifests: map[string][]byte{
				"latest": []byte("manifest-latest"),
				"v1.0":   []byte("manifest-v1.0"),
			},
		},
		"project/repo2": &mockRepository{
			name: "project/repo2",
			tags: []string{"latest"},
			manifests: map[string][]byte{
				"latest": []byte("manifest-latest"),
			},
		},
		"project/repo3": &mockRepository{
			name: "project/repo3",
			tags: []string{"v2.0"},
			manifests: map[string][]byte{
				"v2.0": []byte("manifest-v2.0"),
			},
		},
	}

	destRepos := map[string]common.Repository{
		"dest/repo1": &mockRepository{
			name:      "dest/repo1",
			tags:      []string{},
			manifests: map[string][]byte{},
		},
		"dest/repo2": &mockRepository{
			name:      "dest/repo2",
			tags:      []string{},
			manifests: map[string][]byte{},
		},
		"dest/repo3": &mockRepository{
			name:      "dest/repo3",
			tags:      []string{},
			manifests: map[string][]byte{},
		},
	}

	// Create mock clients
	sourceClient := &mockRegistryClient{
		repositories: sourceRepos,
	}

	destClient := &mockRegistryClient{
		repositories: destRepos,
	}

	// Create mock logger
	logger := &mockLogger{}

	// Create mock copier
	copier := &mockCopier{}

	// Create reconciler
	reconciler := NewReconciler(sourceClient, destClient, copier, metrics.NewNoopMetrics(), logger)

	// Create config with rules
	config := ReplicationConfig{
		MaxConcurrentReplications: 2, // Test concurrent replication
		Rules: []ReplicationRule{
			{
				SourceRepository:      "project/repo1",
				DestinationRepository: "dest/repo1",
				TagFilter:             "*",
			},
			{
				SourceRepository:      "project/repo2",
				DestinationRepository: "dest/repo2",
				TagFilter:             "latest",
			},
			{
				SourceRepository:      "project/repo3",
				DestinationRepository: "dest/repo3",
				TagFilter:             "v*",
			},
		},
	}

	// Run reconciliation
	ctx := context.Background()
	reconciler.ReconcileAllRepositories(ctx, config)

	// Wait a moment for goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Check results - should have copied 4 images total:
	// repo1: latest, v1.0 (2)
	// repo2: latest (1)
	// repo3: v2.0 (1)
	if copier.copyCount != 4 {
		t.Errorf("Expected 4 copy operations, got %d", copier.copyCount)
	}

	// Check destination repositories have correct tags
	for repoName, repo := range destRepos {
		mockRepo := repo.(*mockRepository)

		switch repoName {
		case "dest/repo1":
			if len(mockRepo.tags) != 2 {
				t.Errorf("Expected dest/repo1 to have 2 tags, got %d", len(mockRepo.tags))
			}
		case "dest/repo2":
			if len(mockRepo.tags) != 1 {
				t.Errorf("Expected dest/repo2 to have 1 tag, got %d", len(mockRepo.tags))
			}
			if len(mockRepo.tags) > 0 && mockRepo.tags[0] != "latest" {
				t.Errorf("Expected dest/repo2 to have tag 'latest', got %s", mockRepo.tags[0])
			}
		case "dest/repo3":
			if len(mockRepo.tags) != 1 {
				t.Errorf("Expected dest/repo3 to have 1 tag, got %d", len(mockRepo.tags))
			}
			if len(mockRepo.tags) > 0 && mockRepo.tags[0] != "v2.0" {
				t.Errorf("Expected dest/repo3 to have tag 'v2.0', got %s", mockRepo.tags[0])
			}
		}
	}
}
