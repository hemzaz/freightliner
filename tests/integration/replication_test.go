//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"freightliner/pkg/copy"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"
	"freightliner/pkg/tree"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndReplication tests the complete replication workflow
func TestEndToEndReplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	logger := log.NewLogger()

	// Test with mock registries
	t.Run("MockRegistryReplication", func(t *testing.T) {
		// Setup mock source and destination
		sourceClient := NewMockRegistryClient("source.registry.io", logger)
		destClient := NewMockRegistryClient("dest.registry.io", logger)

		// Populate source with test data
		sourceClient.repositories = []string{
			"app/backend",
			"app/frontend",
			"app/database",
		}

		// Setup tags for each repository
		for _, repo := range sourceClient.repositories {
			sourceClient.tags[repo] = []string{"v1.0.0", "v1.1.0", "latest"}
		}

		// Create copier and tree replicator
		copier := copy.NewCopier(logger)
		replicator := tree.NewTreeReplicator(logger, copier, tree.TreeReplicatorOptions{
			WorkerCount:         3,
			EnableCheckpointing: false,
			DryRun:              false,
		})

		// Execute replication
		result, err := replicator.ReplicateTree(ctx, tree.ReplicateTreeOptions{
			SourceClient:   sourceClient,
			DestClient:     destClient,
			SourcePrefix:   "app/",
			DestPrefix:     "backup/",
			ForceOverwrite: false,
		})

		// Validate results
		require.NoError(t, err, "Replication should succeed")
		assert.Equal(t, 3, result.Repositories, "Should process 3 repositories")
		assert.Greater(t, result.ImagesReplicated.Load(), int64(0), "Should replicate at least one image")
		assert.False(t, result.Interrupted, "Replication should not be interrupted")
		assert.Greater(t, result.Duration, time.Duration(0), "Duration should be positive")
	})

	t.Run("ReplicationWithFiltering", func(t *testing.T) {
		sourceClient := NewMockRegistryClient("source.registry.io", logger)
		destClient := NewMockRegistryClient("dest.registry.io", logger)

		// Populate source
		sourceClient.repositories = []string{
			"prod/service-a",
			"prod/service-b",
			"test/service-c",
		}

		for _, repo := range sourceClient.repositories {
			sourceClient.tags[repo] = []string{"v1.0.0", "v2.0.0", "dev", "latest"}
		}

		copier := copy.NewCopier(logger)
		replicator := tree.NewTreeReplicator(logger, copier, tree.TreeReplicatorOptions{
			WorkerCount:         2,
			ExcludeRepositories: []string{"test/*"},
			ExcludeTags:         []string{"dev"},
			IncludeTags:         []string{"v*"},
			EnableCheckpointing: false,
		})

		result, err := replicator.ReplicateTree(ctx, tree.ReplicateTreeOptions{
			SourceClient: sourceClient,
			DestClient:   destClient,
			SourcePrefix: "prod/",
			DestPrefix:   "backup/",
		})

		require.NoError(t, err)
		assert.Equal(t, 2, result.Repositories, "Should only process prod repositories")
	})

	t.Run("ReplicationWithCheckpointing", func(t *testing.T) {
		// Create temporary directory for checkpoints
		tempDir, err := os.MkdirTemp("", "checkpoint_test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		sourceClient := NewMockRegistryClient("source.registry.io", logger)
		destClient := NewMockRegistryClient("dest.registry.io", logger)

		sourceClient.repositories = []string{"app/service"}
		sourceClient.tags["app/service"] = []string{"v1.0.0", "v1.1.0"}

		copier := copy.NewCopier(logger)
		replicator := tree.NewTreeReplicator(logger, copier, tree.TreeReplicatorOptions{
			WorkerCount:         1,
			EnableCheckpointing: true,
			CheckpointDirectory: tempDir,
		})

		result, err := replicator.ReplicateTree(ctx, tree.ReplicateTreeOptions{
			SourceClient: sourceClient,
			DestClient:   destClient,
			SourcePrefix: "app/",
			DestPrefix:   "backup/",
		})

		require.NoError(t, err)
		assert.NotEmpty(t, result.CheckpointID, "Should have checkpoint ID")
		assert.False(t, result.Interrupted)
	})
}

// TestMultiRegistryReplication tests replication across different registry types
func TestMultiRegistryReplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	logger := log.NewLogger()

	testCases := []struct {
		name         string
		sourceType   string
		destType     string
		repositories int
		tags         int
	}{
		{
			name:         "ECR to GCR",
			sourceType:   "ecr",
			destType:     "gcr",
			repositories: 2,
			tags:         3,
		},
		{
			name:         "GCR to Docker Hub",
			sourceType:   "gcr",
			destType:     "dockerhub",
			repositories: 1,
			tags:         5,
		},
		{
			name:         "Docker Hub to ECR",
			sourceType:   "dockerhub",
			destType:     "ecr",
			repositories: 3,
			tags:         2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sourceClient := NewMockRegistryClient(fmt.Sprintf("source-%s.io", tc.sourceType), logger)
			destClient := NewMockRegistryClient(fmt.Sprintf("dest-%s.io", tc.destType), logger)

			// Setup test data
			for i := 0; i < tc.repositories; i++ {
				repo := fmt.Sprintf("app/service-%d", i)
				sourceClient.repositories = append(sourceClient.repositories, repo)

				tags := make([]string, tc.tags)
				for j := 0; j < tc.tags; j++ {
					tags[j] = fmt.Sprintf("v1.%d.0", j)
				}
				sourceClient.tags[repo] = tags
			}

			copier := copy.NewCopier(logger)
			replicator := tree.NewTreeReplicator(logger, copier, tree.TreeReplicatorOptions{
				WorkerCount: 3,
			})

			result, err := replicator.ReplicateTree(ctx, tree.ReplicateTreeOptions{
				SourceClient: sourceClient,
				DestClient:   destClient,
				SourcePrefix: "app/",
				DestPrefix:   "mirror/",
			})

			require.NoError(t, err)
			assert.Equal(t, tc.repositories, result.Repositories)
			expectedImages := int64(tc.repositories * tc.tags)
			assert.Equal(t, expectedImages, result.ImagesReplicated.Load())
		})
	}
}

// TestErrorScenarios tests various error conditions
func TestErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	logger := log.NewLogger()

	t.Run("SourceRegistryUnreachable", func(t *testing.T) {
		sourceClient := NewMockRegistryClient("unreachable.registry.io", logger)
		sourceClient.simulateError = true
		destClient := NewMockRegistryClient("dest.registry.io", logger)

		copier := copy.NewCopier(logger)
		replicator := tree.NewTreeReplicator(logger, copier, tree.TreeReplicatorOptions{
			WorkerCount: 1,
		})

		result, err := replicator.ReplicateTree(ctx, tree.ReplicateTreeOptions{
			SourceClient: sourceClient,
			DestClient:   destClient,
			SourcePrefix: "app/",
			DestPrefix:   "backup/",
		})

		assert.Error(t, err, "Should fail with unreachable source")
		assert.Equal(t, 0, result.Repositories)
	})

	t.Run("PartialFailure", func(t *testing.T) {
		sourceClient := NewMockRegistryClient("source.registry.io", logger)
		destClient := NewMockRegistryClient("dest.registry.io", logger)

		sourceClient.repositories = []string{"app/good", "app/bad", "app/ok"}
		for _, repo := range sourceClient.repositories {
			sourceClient.tags[repo] = []string{"v1.0.0"}
		}

		// Simulate failure for specific repository
		sourceClient.failOnRepo = "app/bad"

		copier := copy.NewCopier(logger)
		replicator := tree.NewTreeReplicator(logger, copier, tree.TreeReplicatorOptions{
			WorkerCount: 2,
		})

		result, err := replicator.ReplicateTree(ctx, tree.ReplicateTreeOptions{
			SourceClient: sourceClient,
			DestClient:   destClient,
			SourcePrefix: "app/",
			DestPrefix:   "backup/",
		})

		// Should complete despite one failure
		assert.NoError(t, err)
		assert.Equal(t, 3, result.Repositories)
		assert.Greater(t, result.ImagesFailed.Load(), int64(0), "Should have failed images")
		assert.Greater(t, result.ImagesReplicated.Load(), int64(0), "Should have successful images")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		cancelCtx, cancelFunc := context.WithCancel(context.Background())

		sourceClient := NewMockRegistryClient("source.registry.io", logger)
		destClient := NewMockRegistryClient("dest.registry.io", logger)

		sourceClient.repositories = []string{"app/service"}
		sourceClient.tags["app/service"] = []string{"v1.0.0"}
		sourceClient.delayDuration = 500 * time.Millisecond

		copier := copy.NewCopier(logger)
		replicator := tree.NewTreeReplicator(logger, copier, tree.TreeReplicatorOptions{
			WorkerCount: 1,
		})

		// Cancel context shortly after starting
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancelFunc()
		}()

		result, err := replicator.ReplicateTree(cancelCtx, tree.ReplicateTreeOptions{
			SourceClient: sourceClient,
			DestClient:   destClient,
			SourcePrefix: "app/",
			DestPrefix:   "backup/",
		})

		assert.Error(t, err, "Should fail due to cancellation")
		assert.True(t, result.Interrupted, "Should be marked as interrupted")
	})
}

// BenchmarkReplication benchmarks the replication process
func BenchmarkReplication(b *testing.B) {
	logger := log.NewLogger()
	ctx := context.Background()

	benchmarks := []struct {
		name         string
		repositories int
		tags         int
		workers      int
	}{
		{"SmallWorkload", 5, 3, 2},
		{"MediumWorkload", 20, 5, 5},
		{"LargeWorkload", 50, 10, 10},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			sourceClient := NewMockRegistryClient("source.registry.io", logger)
			destClient := NewMockRegistryClient("dest.registry.io", logger)

			// Setup test data
			for i := 0; i < bm.repositories; i++ {
				repo := fmt.Sprintf("app/service-%d", i)
				sourceClient.repositories = append(sourceClient.repositories, repo)

				tags := make([]string, bm.tags)
				for j := 0; j < bm.tags; j++ {
					tags[j] = fmt.Sprintf("v%d.0.0", j)
				}
				sourceClient.tags[repo] = tags
			}

			copier := copy.NewCopier(logger)
			replicator := tree.NewTreeReplicator(logger, copier, tree.TreeReplicatorOptions{
				WorkerCount: bm.workers,
			})

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := replicator.ReplicateTree(ctx, tree.ReplicateTreeOptions{
					SourceClient: sourceClient,
					DestClient:   destClient,
					SourcePrefix: "app/",
					DestPrefix:   "backup/",
				})
				if err != nil {
					b.Fatalf("Replication failed: %v", err)
				}
			}
		})
	}
}

// MockRegistryClient implements interfaces.RegistryClient for testing
type MockRegistryClient struct {
	registryName  string
	repositories  []string
	tags          map[string][]string
	logger        log.Logger
	simulateError bool
	failOnRepo    string
	delayDuration time.Duration
}

func NewMockRegistryClient(registryName string, logger log.Logger) *MockRegistryClient {
	return &MockRegistryClient{
		registryName: registryName,
		tags:         make(map[string][]string),
		logger:       logger,
	}
}

func (m *MockRegistryClient) GetRegistryName() string {
	return m.registryName
}

func (m *MockRegistryClient) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	if m.simulateError {
		return nil, fmt.Errorf("simulated registry error")
	}

	time.Sleep(m.delayDuration)

	filtered := []string{}
	for _, repo := range m.repositories {
		if prefix == "" || len(repo) >= len(prefix) && repo[:len(prefix)] == prefix {
			filtered = append(filtered, repo)
		}
	}

	return filtered, nil
}

func (m *MockRegistryClient) GetRepository(ctx context.Context, repoName string) (interfaces.Repository, error) {
	if m.simulateError || repoName == m.failOnRepo {
		return nil, fmt.Errorf("simulated repository error")
	}

	time.Sleep(m.delayDuration)

	return &MockRepository{
		name:   repoName,
		tags:   m.tags[repoName],
		logger: m.logger,
	}, nil
}

// MockRepository implements interfaces.Repository for testing
type MockRepository struct {
	name   string
	tags   []string
	logger log.Logger
}

func (m *MockRepository) GetName() string {
	return m.name
}

func (m *MockRepository) GetRepositoryName() string {
	return m.name
}

func (m *MockRepository) ListTags(ctx context.Context) ([]string, error) {
	return m.tags, nil
}

func (m *MockRepository) GetImageReference(tag string) (name.Reference, error) {
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", m.name, tag))
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func (m *MockRepository) GetRemoteOptions() ([]remote.Option, error) {
	return []remote.Option{}, nil
}

func (m *MockRepository) DeleteManifest(ctx context.Context, digest string) error {
	// Mock implementation - not used in integration tests
	return nil
}

func (m *MockRepository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	// Mock implementation - not used in integration tests
	return nil, nil
}

func (m *MockRepository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	// Mock implementation - returns mock layer data for testing
	return io.NopCloser(strings.NewReader("mock layer data")), nil
}

func (m *MockRepository) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	// Mock implementation - returns mock manifest for testing
	return &interfaces.Manifest{
		Content:   []byte(`{"schemaVersion": 2, "mediaType": "application/vnd.docker.distribution.manifest.v2+json"}`),
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:    "sha256:mock-digest",
	}, nil
}

func (m *MockRepository) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	// Mock implementation - not used in integration tests
	return nil
}
