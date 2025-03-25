package gcr

import (
	"context"
	"errors"
	"net/http"
	"src/pkg/client/common"
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock for Google Auth
type mockGoogleAuth struct {
	mock.Mock
}

func (m *mockGoogleAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// Mock for Google Container Registry Catalog API
type mockGoogleCatalog struct {
	mock.Mock
}

func (m *mockGoogleCatalog) Catalog(ctx context.Context, registry name.Registry, opt ...google.Option) ([]string, error) {
	args := m.Called(ctx, registry, opt)
	return args.Get(0).([]string), args.Error(1)
}

// Mock for Artifact Registry Client
type mockArtifactRegistryClient struct {
	mock.Mock
}

func (m *mockArtifactRegistryClient) ListRepositories(ctx context.Context, req interface{}, opts ...interface{}) *mockRepositoryIterator {
	args := m.Called(ctx, req, opts)
	return args.Get(0).(*mockRepositoryIterator)
}

// Mock repository struct with the same fields we need from the actual one
type mockRepository struct {
	Name string
}

// Mock for Repository Iterator
type mockRepositoryIterator struct {
	mock.Mock
	repos []*mockRepository
	index int
}

func (m *mockRepositoryIterator) Next() (*mockRepository, error) {
	if m.index >= len(m.repos) {
		return nil, iterator.Done
	}
	repo := m.repos[m.index]
	m.index++
	return repo, nil
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		registry    string
		expectedErr bool
	}{
		{
			name:        "Valid GCR registry",
			registry:    "gcr.io",
			expectedErr: false,
		},
		{
			name:        "Valid Artifact Registry",
			registry:    "us-central1-docker.pkg.dev",
			expectedErr: false,
		},
		{
			name:        "Invalid registry with path",
			registry:    "gcr.io/project/repo",
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(Options{
				Registry: tc.registry,
				// Use nil for auth to use default auth
			})

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tc.registry, client.registry.Name())
			}
		})
	}
}

func TestClientListRepositories(t *testing.T) {
	tests := []struct {
		name         string
		registry     string
		mockGCRSetup func(*mockGoogleCatalog)
		mockARSetup  func(*mockArtifactRegistryClient)
		expected     []string
		expectedErr  bool
	}{
		{
			name:     "GCR successful list",
			registry: "gcr.io",
			mockGCRSetup: func(mockCatalog *mockGoogleCatalog) {
				mockCatalog.On("Catalog", mock.Anything, mock.Anything, mock.Anything).
					Return([]string{"project/repo1", "project/repo2"}, nil)
			},
			mockARSetup: func(mockAR *mockArtifactRegistryClient) {
				// Should not be called
			},
			expected:    []string{"project/repo1", "project/repo2"},
			expectedErr: false,
		},
		{
			name:     "GCR error with AR fallback",
			registry: "gcr.io",
			mockGCRSetup: func(mockCatalog *mockGoogleCatalog) {
				mockCatalog.On("Catalog", mock.Anything, mock.Anything, mock.Anything).
					Return([]string{}, errors.New("GCR error"))
			},
			mockARSetup: func(mockAR *mockArtifactRegistryClient) {
				iterator := &mockRepositoryIterator{
					repos: []*artifactregistrypb.Repository{
						{
							Name: "projects/project/locations/us/repositories/repo3",
						},
						{
							Name: "projects/project/locations/us/repositories/repo4",
						},
					},
				}
				mockAR.On("ListRepositories", mock.Anything, mock.Anything, mock.Anything).
					Return(iterator)
			},
			expected:    []string{"project/repo3", "project/repo4"},
			expectedErr: false,
		},
		{
			name:     "Both GCR and AR error",
			registry: "gcr.io",
			mockGCRSetup: func(mockCatalog *mockGoogleCatalog) {
				mockCatalog.On("Catalog", mock.Anything, mock.Anything, mock.Anything).
					Return([]string{}, errors.New("GCR error"))
			},
			mockARSetup: func(mockAR *mockArtifactRegistryClient) {
				mockAR.On("ListRepositories", mock.Anything, mock.Anything, mock.Anything).
					Return(&mockRepositoryIterator{})
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCatalog := &mockGoogleCatalog{}
			mockAR := &mockArtifactRegistryClient{}
			tc.mockGCRSetup(mockCatalog)
			tc.mockARSetup(mockAR)

			reg, err := name.NewRegistry(tc.registry)
			assert.NoError(t, err)

			client := &Client{
				registry:            reg,
				catalogFunc:         mockCatalog.Catalog,
				artifactRegistryAPI: mockAR,
			}

			repos, err := client.ListRepositories()
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expected, repos)
			}

			mockCatalog.AssertExpectations(t)
			mockAR.AssertExpectations(t)
		})
	}
}

func TestClientGetRepository(t *testing.T) {
	tests := []struct {
		name             string
		registry         string
		repoName         string
		expectedRepoName string
		expectedErr      bool
		expectedErrType  error
	}{
		{
			name:             "Valid repository",
			registry:         "gcr.io",
			repoName:         "project/repo",
			expectedRepoName: "project/repo",
			expectedErr:      false,
		},
		{
			name:             "Empty repository name",
			registry:         "gcr.io",
			repoName:         "",
			expectedRepoName: "",
			expectedErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := name.NewRegistry(tc.registry)
			assert.NoError(t, err)

			client := &Client{
				registry: reg,
				keychain: authn.DefaultKeychain,
			}

			repo, err := client.GetRepository(tc.repoName)
			if tc.expectedErr {
				assert.Error(t, err)
				if tc.expectedErrType != nil {
					assert.True(t, errors.Is(err, tc.expectedErrType))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
				assert.Equal(t, tc.expectedRepoName, repo.GetRepositoryName())
			}
		})
	}
}

func TestParseGCRRepository(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		registry           string
		expectedRegistry   string
		expectedRepository string
		expectedErr        bool
	}{
		{
			name:               "Full GCR URI",
			input:              "gcr.io/project/repo-name",
			registry:           "gcr.io",
			expectedRegistry:   "gcr.io",
			expectedRepository: "project/repo-name",
			expectedErr:        false,
		},
		{
			name:               "Repository with suffix",
			input:              "gcr.io/project/repo-name:latest",
			registry:           "gcr.io",
			expectedRegistry:   "gcr.io",
			expectedRepository: "project/repo-name",
			expectedErr:        false,
		},
		{
			name:               "Repository with deep path",
			input:              "gcr.io/project/group/repo-name",
			registry:           "gcr.io",
			expectedRegistry:   "gcr.io",
			expectedRepository: "project/group/repo-name",
			expectedErr:        false,
		},
		{
			name:               "Simple repository (error expected)",
			input:              "repo-name",
			registry:           "gcr.io",
			expectedRegistry:   "",
			expectedRepository: "",
			expectedErr:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reg, _ := name.NewRegistry(tc.registry)

			registry, repository, err := parseGCRRepository(tc.input, reg)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRegistry, registry.String())
				assert.Equal(t, tc.expectedRepository, repository)
			}
		})
	}
}

func TestExtractProjectFromRepository(t *testing.T) {
	tests := []struct {
		name            string
		repository      string
		expectedProject string
		expectedErr     bool
	}{
		{
			name:            "Valid project/repo path",
			repository:      "project/repo",
			expectedProject: "project",
			expectedErr:     false,
		},
		{
			name:            "Valid project/path/repo path",
			repository:      "project/path/repo",
			expectedProject: "project",
			expectedErr:     false,
		},
		{
			name:            "Invalid repo without project",
			repository:      "repo",
			expectedProject: "",
			expectedErr:     true,
		},
		{
			name:            "Empty string",
			repository:      "",
			expectedProject: "",
			expectedErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			project, err := extractProjectFromRepository(tc.repository)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedProject, project)
			}
		})
	}
}

func TestIsGCRRegistry(t *testing.T) {
	tests := []struct {
		name     string
		registry string
		expected bool
	}{
		{
			name:     "Primary GCR domain",
			registry: "gcr.io",
			expected: true,
		},
		{
			name:     "Regional GCR domain",
			registry: "us.gcr.io",
			expected: true,
		},
		{
			name:     "European GCR domain",
			registry: "eu.gcr.io",
			expected: true,
		},
		{
			name:     "Artifact Registry",
			registry: "us-central1-docker.pkg.dev",
			expected: false,
		},
		{
			name:     "Docker Hub",
			registry: "docker.io",
			expected: false,
		},
		{
			name:     "Empty string",
			registry: "",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isGCRRegistry(tc.registry)
			assert.Equal(t, tc.expected, result)
		})
	}
}
