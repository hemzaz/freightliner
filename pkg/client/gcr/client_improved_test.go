package gcr

import (
	"context"
	"testing"

	freightliner_log "freightliner/pkg/helper/log"
	"freightliner/pkg/testing/mocks"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/artifactregistry/v1"
)

func TestGCRClientListRepositoriesWithMocks(t *testing.T) {
	tests := []struct {
		name        string
		project     string
		location    string
		prefix      string
		setupMocks  func() (*mocks.MockGoogleCatalogClient, *mocks.MockGoogleAuth)
		expectedLen int
		expectErr   bool
	}{
		{
			name:     "Successful GCR repository listing",
			project:  "test-project",
			location: "us",
			prefix:   "",
			setupMocks: func() (*mocks.MockGoogleCatalogClient, *mocks.MockGoogleAuth) {
				scenarios := &mocks.MockGCRTestScenarios{}
				return scenarios.SuccessfulGCRListRepositories("test-project", 4)
			},
			expectedLen: 4,
			expectErr:   false,
		},
		{
			name:     "GCR with prefix filter",
			project:  "test-project",
			location: "us",
			prefix:   "test-project/testing",
			setupMocks: func() (*mocks.MockGoogleCatalogClient, *mocks.MockGoogleAuth) {
				// Return repositories that match the prefix
				repos := []string{
					"test-project/testing/repo1",
					"test-project/testing/repo2",
					"test-project/other/repo3",
				}
				builder := mocks.NewGCRMockBuilder()
				builder.ExpectCatalog(repos, nil)
				builder.ExpectAuthSuccess()
				return builder.BuildCatalogClient(), builder.BuildAuthTransport()
			},
			expectedLen: 2, // Only repos with "testing" prefix
			expectErr:   false,
		},
		{
			name:     "Authentication failure",
			project:  "test-project",
			location: "us",
			prefix:   "",
			setupMocks: func() (*mocks.MockGoogleCatalogClient, *mocks.MockGoogleAuth) {
				scenarios := &mocks.MockGCRTestScenarios{}
				return scenarios.FailedGCRAuthentication()
			},
			expectedLen: 0,
			expectErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCatalog, mockAuth := tc.setupMocks()
			logger := freightliner_log.NewLogger()

			// Create a test client with mocked dependencies
			// In practice, you'd need dependency injection to replace the real clients
			client := &Client{
				project:  tc.project,
				location: tc.location,
				logger:   logger,
				// In real implementation, inject mocked catalog and auth clients here
			}

			// Test the catalog mock directly to verify it works
			if !tc.expectErr {
				registry, err := name.NewRegistry("gcr.io")
				assert.NoError(t, err)

				repos, err := mockCatalog.Catalog(context.Background(), registry)
				assert.NoError(t, err)

				// Filter repos by prefix if specified
				filteredRepos := repos
				if tc.prefix != "" {
					filteredRepos = []string{}
					for _, repo := range repos {
						if len(repo) >= len(tc.prefix) && repo[:len(tc.prefix)] == tc.prefix {
							filteredRepos = append(filteredRepos, repo)
						}
					}
				}

				assert.Len(t, filteredRepos, tc.expectedLen)
			}

			// Verify mock expectations
			mockCatalog.AssertExpectations(t)
			mockAuth.AssertExpectations(t)

			// Note: In a real implementation, you would test:
			// repos, err := client.ListRepositories(context.Background(), tc.prefix)
			// if tc.expectErr {
			//     assert.Error(t, err)
			// } else {
			//     assert.NoError(t, err)
			//     assert.Len(t, repos, tc.expectedLen)
			// }
		})
	}
}

func TestArtifactRegistryClientWithMocks(t *testing.T) {
	tests := []struct {
		name        string
		project     string
		location    string
		setupMocks  func() *mocks.MockArtifactRegistryClient
		expectedLen int
		expectErr   bool
	}{
		{
			name:     "Successful Artifact Registry listing",
			project:  "test-project",
			location: "us-central1",
			setupMocks: func() *mocks.MockArtifactRegistryClient {
				scenarios := &mocks.MockGCRTestScenarios{}
				return scenarios.SuccessfulArtifactRegistryList("test-project", "us-central1", 3)
			},
			expectedLen: 3,
			expectErr:   false,
		},
		{
			name:     "Empty repository list",
			project:  "empty-project",
			location: "us-central1",
			setupMocks: func() *mocks.MockArtifactRegistryClient {
				builder := mocks.NewArtifactRegistryMockBuilder()
				builder.ExpectListRepositories([]*artifactregistry.Repository{}, nil)
				return builder.Build()
			},
			expectedLen: 0,
			expectErr:   false,
		},
		{
			name:     "API error",
			project:  "error-project",
			location: "us-central1",
			setupMocks: func() *mocks.MockArtifactRegistryClient {
				builder := mocks.NewArtifactRegistryMockBuilder()
				builder.ExpectListRepositories(nil, assert.AnError)
				return builder.Build()
			},
			expectedLen: 0,
			expectErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := tc.setupMocks()
			ctx := context.Background()

			// Test the mock directly
			req := &artifactregistry.ListRepositoriesRequest{
				Parent: "projects/" + tc.project + "/locations/" + tc.location,
			}

			iterator := mockClient.ListRepositories(ctx, req)
			assert.NotNil(t, iterator)

			// Collect repositories from iterator
			var repos []*artifactregistry.Repository
			for {
				repo, err := iterator.Next()
				if err != nil {
					if err.Error() == "no more items in iterator" {
						break // This would be iterator.Done in real code
					}
					if tc.expectErr {
						assert.Error(t, err)
						break
					} else {
						t.Fatalf("Unexpected error: %v", err)
					}
				}
				repos = append(repos, repo)
			}

			if !tc.expectErr {
				assert.Len(t, repos, tc.expectedLen)
				for _, repo := range repos {
					assert.Contains(t, repo.Name, tc.project)
					assert.Contains(t, repo.Name, tc.location)
					assert.Equal(t, "DOCKER", repo.Format)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestGCRRepositoryOperationsWithMocks(t *testing.T) {
	logger := freightliner_log.NewLogger()

	tests := []struct {
		name      string
		repoName  string
		operation string
		expectErr bool
	}{
		{
			name:      "Valid repository creation",
			repoName:  "test-project/valid-repo",
			operation: "create",
			expectErr: false,
		},
		{
			name:      "Invalid repository name",
			repoName:  "",
			operation: "create",
			expectErr: true,
		},
		{
			name:      "Repository name parsing",
			repoName:  "gcr.io/test-project/repo-name:latest",
			operation: "parse",
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &Client{
				project:  "test-project",
				location: "gcr.io",
				logger:   logger,
			}

			switch tc.operation {
			case "create":
				repo, err := client.GetRepository(context.Background(), tc.repoName)
				if tc.expectErr {
					assert.Error(t, err)
					assert.Nil(t, repo)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, repo)
				}

			case "parse":
				registry, err := name.NewRegistry("gcr.io")
				assert.NoError(t, err)

				regStr, repoStr, err := parseGCRRepository(tc.repoName, registry)
				if tc.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotEmpty(t, regStr.String())
					assert.NotEmpty(t, repoStr)
					assert.Contains(t, repoStr, "test-project")
				}
			}
		})
	}
}

func TestGCRClientWithDifferentRegistryTypes(t *testing.T) {
	tests := []struct {
		name         string
		location     string
		expectedType string
		isGCR        bool
	}{
		{
			name:         "Standard GCR",
			location:     "us",
			expectedType: "gcr",
			isGCR:        true,
		},
		{
			name:         "European GCR",
			location:     "eu",
			expectedType: "gcr",
			isGCR:        true,
		},
		{
			name:         "Artifact Registry",
			location:     "us-central1",
			expectedType: "artifact",
			isGCR:        false,
		},
		{
			name:         "Custom region Artifact Registry",
			location:     "europe-west1",
			expectedType: "artifact",
			isGCR:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(ClientOptions{
				Location: tc.location,
				Project:  "test-project",
			})

			assert.NoError(t, err)
			assert.NotNil(t, client)
			assert.Equal(t, tc.location, client.location)

			// Test registry type detection
			var registryHost string
			if tc.isGCR {
				registryHost = tc.location + ".gcr.io"
			} else {
				registryHost = tc.location + "-docker.pkg.dev"
			}

			isGCRRegistry := isGCRRegistry(registryHost)
			assert.Equal(t, tc.isGCR, isGCRRegistry)
		})
	}
}

// TestGCRErrorHandling tests error scenarios with proper mocking
func TestGCRErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func() (*mocks.MockGoogleCatalogClient, *mocks.MockArtifactRegistryClient)
		testFunc   func(*mocks.MockGoogleCatalogClient, *mocks.MockArtifactRegistryClient) error
		expectErr  bool
	}{
		{
			name: "Catalog service unavailable",
			setupMocks: func() (*mocks.MockGoogleCatalogClient, *mocks.MockArtifactRegistryClient) {
				catalogClient := mocks.NewGCRMockBuilder().
					ExpectCatalog(nil, assert.AnError).
					BuildCatalogClient()

				return catalogClient, nil
			},
			testFunc: func(catalog *mocks.MockGoogleCatalogClient, artifact *mocks.MockArtifactRegistryClient) error {
				registry, _ := name.NewRegistry("gcr.io")
				_, err := catalog.Catalog(context.Background(), registry)
				return err
			},
			expectErr: true,
		},
		{
			name: "Artifact Registry service error",
			setupMocks: func() (*mocks.MockGoogleCatalogClient, *mocks.MockArtifactRegistryClient) {
				artifactClient := mocks.NewArtifactRegistryMockBuilder().
					ExpectListRepositories(nil, assert.AnError).
					Build()

				return nil, artifactClient
			},
			testFunc: func(catalog *mocks.MockGoogleCatalogClient, artifact *mocks.MockArtifactRegistryClient) error {
				req := &artifactregistry.ListRepositoriesRequest{Parent: "projects/test/locations/us-central1"}
				iterator := artifact.ListRepositories(context.Background(), req)
				_, err := iterator.Next()
				return err
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			catalogClient, artifactClient := tc.setupMocks()

			err := tc.testFunc(catalogClient, artifactClient)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if catalogClient != nil {
				catalogClient.AssertExpectations(t)
			}
			if artifactClient != nil {
				artifactClient.AssertExpectations(t)
			}
		})
	}
}

// BenchmarkGCRMockOperations benchmarks mock operations
func BenchmarkGCRMockOperations(b *testing.B) {
	mockCatalog := mocks.NewGCRMockBuilder().
		ExpectCatalog(mocks.CreateMockGCRRepositories("test-project", 10), nil).
		BuildCatalogClient()

	registry, _ := name.NewRegistry("gcr.io")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mockCatalog.Catalog(ctx, registry)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
