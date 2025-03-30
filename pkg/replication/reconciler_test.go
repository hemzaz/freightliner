package replication

import (
	"context"
	stderrors "errors"
	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/metrics"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Mock registry client
type mockRegistryClient struct {
	repositories map[string]common.Repository
	listError    error
}

func (m *mockRegistryClient) GetRepository(ctx context.Context, name string) (common.Repository, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	repo, exists := m.repositories[name]
	if !exists {
		return nil, errors.Wrap(errors.NotFoundf("repository not found"), name)
	}

	return repo, nil
}

func (m *mockRegistryClient) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	repos := make([]string, 0, len(m.repositories))
	for name := range m.repositories {
		repos = append(repos, name)
	}

	return repos, nil
}

func (m *mockRegistryClient) GetRegistryName() string {
	return "mock-registry"
}

// Mock repository
type mockRepository struct {
	name      string
	tags      []string
	manifests map[string]*common.Manifest
	listError error
	getError  error
	putError  error
	delError  error
}

func (m *mockRepository) GetRepositoryName() string {
	return m.name
}

// GetName is an alias for GetRepositoryName for backward compatibility
func (m *mockRepository) GetName() string {
	return m.GetRepositoryName()
}

func (m *mockRepository) ListTags() ([]string, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.tags, nil
}

func (m *mockRepository) GetManifest(ctx context.Context, tag string) (*common.Manifest, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	manifest, exists := m.manifests[tag]
	if !exists {
		return nil, common.NewRegistryError("manifest not found", common.ErrNotFound)
	}

	return manifest, nil
}

func (m *mockRepository) PutManifest(ctx context.Context, tag string, manifest *common.Manifest) error {
	if m.putError != nil {
		return m.putError
	}

	if m.manifests == nil {
		m.manifests = make(map[string]*common.Manifest)
	}

	m.manifests[tag] = manifest

	// Add tag to list if not present
	tagFound := false
	for _, t := range m.tags {
		if t == tag {
			tagFound = true
			break
		}
	}

	if !tagFound {
		m.tags = append(m.tags, tag)
	}

	return nil
}

func (m *mockRepository) DeleteManifest(ctx context.Context, tag string) error {
	if m.delError != nil {
		return m.delError
	}

	delete(m.manifests, tag)

	// Remove tag from list
	for i, t := range m.tags {
		if t == tag {
			m.tags = append(m.tags[:i], m.tags[i+1:]...)
			break
		}
	}

	return nil
}

func (m *mockRepository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock layer content")), nil
}

func (m *mockRepository) GetImageReference(tag string) (name.Reference, error) {
	return name.NewTag("example.com/repo:" + tag)
}

func (m *mockRepository) GetRemoteOptions() ([]remote.Option, error) {
	return []remote.Option{}, nil
}

// Mock copier
type mockCopier struct {
	copyError  error
	copiedTags []string
}

func (m *mockCopier) CopyTag(ctx context.Context, sourceRepo, destRepo common.Repository, tag string, dryRun bool) error {
	if m.copyError != nil {
		return m.copyError
	}

	// Append to copied tags for verification
	m.copiedTags = append(m.copiedTags, tag)
	return nil
}

// Mock logger
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Debug(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, "DEBUG: "+msg)
}

func (m *mockLogger) Info(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, "INFO: "+msg)
}

func (m *mockLogger) Warn(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, "WARN: "+msg)
}

func (m *mockLogger) Error(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, "ERROR: "+msg)
}

// Mock metrics
type mockMetrics struct {
	reconcileStart     int
	reconcileComplete  int
	repositoryComplete int
	tagCopyStart       int
	tagCopyComplete    int
	tagCopyError       int
}

func (m *mockMetrics) ReconcileStart(sourceRegistry, destRegistry string) {
	m.reconcileStart++
}

func (m *mockMetrics) ReconcileComplete(sourceRegistry, destRegistry string, duration time.Duration, repoCount, tagCount int) {
	m.reconcileComplete++
}

func (m *mockMetrics) RepositoryComplete(sourceRepo, destRepo string, duration time.Duration, tagCount int) {
	m.repositoryComplete++
}

func (m *mockMetrics) TagCopyStart(sourceRepo, destRepo, tag string) {
	m.tagCopyStart++
}

func (m *mockMetrics) TagCopyComplete(sourceRepo, destRepo, tag string, duration time.Duration, status metrics.TagCopyStatus) {
	m.tagCopyComplete++
}

func (m *mockMetrics) TagCopyError(sourceRepo, destRepo, tag string, err error) {
	m.tagCopyError++
}

// Tests for reconcile repo function
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
			name: "Destination Empty",
			sourceRepo: mockRepository{
				name: "test/repo",
				tags: []string{"latest", "v1.0", "v2.0"},
				manifests: map[string]*common.Manifest{
					"latest": {Content: []byte("manifest1"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
					"v1.0":   {Content: []byte("manifest2"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
					"v2.0":   {Content: []byte("manifest3"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
				},
			},
			destRepo: mockRepository{
				name:      "test/repo",
				tags:      []string{},
				manifests: map[string]*common.Manifest{},
			},
			copyError:    nil,
			expectCopies: 3,
			expectErrors: 0,
		},
		{
			name: "Destination Partial",
			sourceRepo: mockRepository{
				name: "test/repo",
				tags: []string{"latest", "v1.0", "v2.0"},
				manifests: map[string]*common.Manifest{
					"latest": {Content: []byte("manifest1"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
					"v1.0":   {Content: []byte("manifest2"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
					"v2.0":   {Content: []byte("manifest3"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
				},
			},
			destRepo: mockRepository{
				name: "test/repo",
				tags: []string{"latest"},
				manifests: map[string]*common.Manifest{
					"latest": {Content: []byte("manifest1"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
				},
			},
			copyError:    nil,
			expectCopies: 2,
			expectErrors: 0,
		},
		{
			name: "Destination Full",
			sourceRepo: mockRepository{
				name: "test/repo",
				tags: []string{"latest", "v1.0", "v2.0"},
				manifests: map[string]*common.Manifest{
					"latest": {Content: []byte("manifest1"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
					"v1.0":   {Content: []byte("manifest2"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
					"v2.0":   {Content: []byte("manifest3"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
				},
			},
			destRepo: mockRepository{
				name: "test/repo",
				tags: []string{"latest", "v1.0", "v2.0"},
				manifests: map[string]*common.Manifest{
					"latest": {Content: []byte("manifest1"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
					"v1.0":   {Content: []byte("manifest2"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
					"v2.0":   {Content: []byte("manifest3"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
				},
			},
			copyError:    nil,
			expectCopies: 0,
			expectErrors: 0,
		},
		{
			name: "Source Error",
			sourceRepo: mockRepository{
				name:      "test/repo",
				listError: stderrors.New("list error"),
			},
			destRepo: mockRepository{
				name: "test/repo",
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
			metrics := &mockMetrics{}
			reconciler := NewReconciler(logger, metrics)

			// Run reconcileRepository
			ctx := context.Background()
			result, err := reconciler.reconcileRepository(ctx, sourceClient, destClient, tt.sourceRepo.name, copier, false)

			// Check results
			if err != nil && tt.expectErrors == 0 {
				t.Errorf("Expected no errors, got: %v", err)
			}

			if err == nil && tt.expectErrors > 0 {
				t.Errorf("Expected errors, got none")
			}

			if len(copier.copiedTags) != tt.expectCopies {
				t.Errorf("Expected %d copies, got %d", tt.expectCopies, len(copier.copiedTags))
			}

			if result.Errors != tt.expectErrors {
				t.Errorf("Expected %d errors in result, got %d", tt.expectErrors, result.Errors)
			}

			if result.Copied != tt.expectCopies {
				t.Errorf("Expected %d copied in result, got %d", tt.expectCopies, result.Copied)
			}
		})
	}
}

// Tests for reconcile function
func TestReconcile(t *testing.T) {
	// Simple test case for now
	// Create mock repositories
	sourceRepo := mockRepository{
		name: "test/repo",
		tags: []string{"latest", "v1.0"},
		manifests: map[string]*common.Manifest{
			"latest": {Content: []byte("manifest1"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
			"v1.0":   {Content: []byte("manifest2"), MediaType: "application/vnd.docker.distribution.manifest.v2+json"},
		},
	}

	destRepo := mockRepository{
		name:      "test/repo",
		tags:      []string{},
		manifests: map[string]*common.Manifest{},
	}

	// Create mock clients
	sourceClient := &mockRegistryClient{
		repositories: map[string]common.Repository{
			sourceRepo.name: &sourceRepo,
		},
	}

	destClient := &mockRegistryClient{
		repositories: map[string]common.Repository{
			destRepo.name: &destRepo,
		},
	}

	// Create mock logger
	logger := &mockLogger{}

	// Create mock copier
	copier := &mockCopier{}

	// Create mock metrics
	metrics := &mockMetrics{}

	// Create reconciler
	reconciler := NewReconciler(logger, metrics)

	// Run reconcile
	ctx := context.Background()
	config := ReconcileConfig{
		SourceRegistry: "source-registry",
		DestRegistry:   "dest-registry",
		SourceClient:   sourceClient,
		DestClient:     destClient,
		Copier:         copier,
		DryRun:         false,
	}

	result, err := reconciler.Reconcile(ctx, config)

	// Check results
	if err != nil {
		t.Errorf("Reconcile returned error: %v", err)
	}

	if result.Repositories != 1 {
		t.Errorf("Expected 1 repository, got %d", result.Repositories)
	}

	if result.TagsCopied != 2 {
		t.Errorf("Expected 2 tags copied, got %d", result.TagsCopied)
	}

	if result.Errors != 0 {
		t.Errorf("Expected 0 errors, got %d", result.Errors)
	}

	// Verify metrics were recorded
	if metrics.reconcileStart != 1 {
		t.Errorf("Expected reconcileStart to be 1, got %d", metrics.reconcileStart)
	}

	if metrics.reconcileComplete != 1 {
		t.Errorf("Expected reconcileComplete to be 1, got %d", metrics.reconcileComplete)
	}

	if metrics.repositoryComplete != 1 {
		t.Errorf("Expected repositoryComplete to be 1, got %d", metrics.repositoryComplete)
	}

	if metrics.tagCopyStart != 2 {
		t.Errorf("Expected tagCopyStart to be 2, got %d", metrics.tagCopyStart)
	}

	if metrics.tagCopyComplete != 2 {
		t.Errorf("Expected tagCopyComplete to be 2, got %d", metrics.tagCopyComplete)
	}
}
