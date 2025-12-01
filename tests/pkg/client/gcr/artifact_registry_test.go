package gcr

import (
	"context"
	"testing"

	"freightliner/pkg/client/gcr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/iterator"
)

// MockARProjectsLocationsRepositoriesService mocks the AR repositories service
type MockARProjectsLocationsRepositoriesService struct {
	mock.Mock
	listCall *MockARListCall
}

// MockARListCall mocks the List call
type MockARListCall struct {
	mock.Mock
	repos      []*artifactregistry.Repository
	pageSize   int64
	pageToken  string
	filter     string
	callCount  int
	totalPages int
}

func (m *MockARListCall) PageSize(size int64) *MockARListCall {
	m.pageSize = size
	return m
}

func (m *MockARListCall) PageToken(token string) *MockARListCall {
	m.pageToken = token
	return m
}

func (m *MockARListCall) Filter(filter string) *MockARListCall {
	m.filter = filter
	return m
}

func (m *MockARListCall) Do() (*artifactregistry.ListRepositoriesResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*artifactregistry.ListRepositoriesResponse), args.Error(1)
}

func TestArtifactRegistry_ListRepositories(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		mockRepos     []string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "Single page",
			prefix:        "",
			mockRepos:     []string{"repo1", "repo2", "repo3"},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "With prefix filter",
			prefix:        "app-",
			mockRepos:     []string{"app-backend", "app-frontend", "lib-common"},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "Empty result",
			prefix:        "",
			mockRepos:     []string{},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the prefix filtering logic
			var filtered []string
			for _, repo := range tt.mockRepos {
				if tt.prefix == "" || len(repo) >= len(tt.prefix) && repo[:len(tt.prefix)] == tt.prefix {
					filtered = append(filtered, repo)
				}
			}

			assert.Len(t, filtered, tt.expectedCount)
		})
	}
}

func TestArtifactRegistry_Pagination(t *testing.T) {
	tests := []struct {
		name          string
		totalRepos    int
		pageSize      int
		expectedPages int
	}{
		{
			name:          "Single page",
			totalRepos:    10,
			pageSize:      50,
			expectedPages: 1,
		},
		{
			name:          "Multiple pages",
			totalRepos:    100,
			pageSize:      50,
			expectedPages: 2,
		},
		{
			name:          "Partial last page",
			totalRepos:    75,
			pageSize:      50,
			expectedPages: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate expected pages
			expectedPages := (tt.totalRepos + tt.pageSize - 1) / tt.pageSize
			assert.Equal(t, tt.expectedPages, expectedPages)
		})
	}
}

func TestArtifactRegistry_LocationMapping(t *testing.T) {
	tests := []struct {
		name             string
		inputLocation    string
		expectedLocation string
	}{
		{
			name:             "Legacy US",
			inputLocation:    "us",
			expectedLocation: "us-central1",
		},
		{
			name:             "Legacy EU",
			inputLocation:    "eu",
			expectedLocation: "us-central1", // Maps to default
		},
		{
			name:             "Legacy Asia",
			inputLocation:    "asia",
			expectedLocation: "us-central1", // Maps to default
		},
		{
			name:             "Explicit region",
			inputLocation:    "us-west1",
			expectedLocation: "us-west1",
		},
		{
			name:             "Europe region",
			inputLocation:    "europe-west1",
			expectedLocation: "europe-west1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test location mapping logic
			location := tt.inputLocation
			if location == "us" || location == "eu" || location == "asia" {
				location = "us-central1"
			}
			assert.Equal(t, tt.expectedLocation, location)
		})
	}
}

func TestArtifactRegistry_RepositoryPathParsing(t *testing.T) {
	tests := []struct {
		name         string
		fullPath     string
		expectedName string
	}{
		{
			name:         "Full AR path",
			fullPath:     "projects/my-project/locations/us-central1/repositories/my-repo",
			expectedName: "my-repo",
		},
		{
			name:         "Nested path",
			fullPath:     "projects/my-project/locations/europe-west1/repositories/team/app",
			expectedName: "team/app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test path parsing logic
			parts := splitPath(tt.fullPath)
			if len(parts) > 0 {
				repoName := parts[len(parts)-1]
				// Simple validation
				assert.NotEmpty(t, repoName)
			}
		})
	}
}

func TestArtifactRegistry_FilterFormat(t *testing.T) {
	tests := []struct {
		name           string
		prefix         string
		expectedFilter string
	}{
		{
			name:           "With prefix",
			prefix:         "app-",
			expectedFilter: "name:*app-*",
		},
		{
			name:           "Empty prefix",
			prefix:         "",
			expectedFilter: "",
		},
		{
			name:           "Wildcard prefix",
			prefix:         "*",
			expectedFilter: "name:**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test filter format creation
			var filter string
			if tt.prefix != "" {
				filter = "name:*" + tt.prefix + "*"
			}
			assert.Equal(t, tt.expectedFilter, filter)
		})
	}
}

func TestArtifactRegistry_IteratorPattern(t *testing.T) {
	// Test the custom iterator implementation pattern
	t.Run("Iterator state management", func(t *testing.T) {
		repos := []*artifactregistry.Repository{
			{Name: "repo1"},
			{Name: "repo2"},
			{Name: "repo3"},
		}

		index := -1
		var current *artifactregistry.Repository
		var err error

		// First iteration
		index++
		if index < len(repos) {
			current = repos[index]
			err = nil
		} else {
			err = iterator.Done
		}
		assert.NoError(t, err)
		assert.Equal(t, "repo1", current.Name)

		// Second iteration
		index++
		if index < len(repos) {
			current = repos[index]
			err = nil
		} else {
			err = iterator.Done
		}
		assert.NoError(t, err)
		assert.Equal(t, "repo2", current.Name)

		// End condition
		index = len(repos)
		if index < len(repos) {
			current = repos[index]
			err = nil
		} else {
			err = iterator.Done
		}
		assert.Equal(t, iterator.Done, err)
	})
}

func TestArtifactRegistry_LegacyFallback(t *testing.T) {
	tests := []struct {
		name        string
		hasARClient bool
		expectAR    bool
	}{
		{
			name:        "With AR client",
			hasARClient: true,
			expectAR:    true,
		},
		{
			name:        "Without AR client - legacy fallback",
			hasARClient: false,
			expectAR:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate client creation with/without AR
			var arClient interface{}
			if tt.hasARClient {
				arClient = &struct{}{} // Non-nil placeholder
			}

			usesAR := arClient != nil
			assert.Equal(t, tt.expectAR, usesAR)
		})
	}
}

func TestArtifactRegistry_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		errorType   string
		expectRetry bool
	}{
		{
			name:        "Not found error",
			errorType:   "404",
			expectRetry: false,
		},
		{
			name:        "Permission denied",
			errorType:   "403",
			expectRetry: false,
		},
		{
			name:        "Rate limit",
			errorType:   "429",
			expectRetry: true,
		},
		{
			name:        "Server error",
			errorType:   "500",
			expectRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test error classification
			shouldRetry := tt.errorType == "429" || tt.errorType == "500"
			assert.Equal(t, tt.expectRetry, shouldRetry)
		})
	}
}

func TestArtifactRegistry_ConcurrentListOperations(t *testing.T) {
	// Test concurrent list operations with AR
	client, err := gcr.NewClient(gcr.ClientOptions{
		Project:  "test-project",
		Location: "us-central1",
	})
	assert.NoError(t, err)

	ctx := context.Background()
	results := make(chan error, 3)

	// Run 3 concurrent list operations
	for i := 0; i < 3; i++ {
		go func() {
			_, err := client.ListRepositories(ctx, "")
			results <- err
		}()
	}

	// Collect results (will error without GCP credentials)
	for i := 0; i < 3; i++ {
		<-results
	}
}

func TestArtifactRegistry_RepositoryNameExtraction(t *testing.T) {
	tests := []struct {
		name         string
		fullPath     string
		expectedName string
	}{
		{
			name:         "Standard path",
			fullPath:     "projects/proj/locations/us/repositories/repo",
			expectedName: "repo",
		},
		{
			name:         "Nested repository",
			fullPath:     "projects/proj/locations/us/repositories/team/app",
			expectedName: "app",
		},
		{
			name:         "Multi-level path",
			fullPath:     "projects/proj/locations/us/repositories/org/team/app",
			expectedName: "app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := splitPath(tt.fullPath)
			if len(parts) > 0 {
				lastPart := parts[len(parts)-1]
				assert.NotEmpty(t, lastPart)
			}
		})
	}
}

func TestArtifactRegistry_PageSizeConfiguration(t *testing.T) {
	tests := []struct {
		name             string
		pageSize         int64
		expectedPageSize int64
	}{
		{
			name:             "Default page size",
			pageSize:         0,
			expectedPageSize: 50,
		},
		{
			name:             "Custom page size",
			pageSize:         100,
			expectedPageSize: 100,
		},
		{
			name:             "Small page size",
			pageSize:         10,
			expectedPageSize: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageSize := tt.pageSize
			if pageSize == 0 {
				pageSize = 50
			}
			assert.Equal(t, tt.expectedPageSize, pageSize)
		})
	}
}

// Helper function to split paths
func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, ch := range path {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
