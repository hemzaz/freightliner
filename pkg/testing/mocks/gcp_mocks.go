package mocks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/iterator"
)

// ErrIteratorDone is returned when an iterator has no more items
var ErrIteratorDone = errors.New("no more items in iterator")

// MockGoogleAuth implements a mock HTTP transport for Google authentication
type MockGoogleAuth struct {
	mock.Mock
}

func (m *MockGoogleAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockGoogleCatalogClient implements a mock for Google Container Registry Catalog API
type MockGoogleCatalogClient struct {
	mock.Mock
}

func (m *MockGoogleCatalogClient) Catalog(ctx context.Context, registry name.Registry, opts ...google.Option) ([]string, error) {
	args := m.Called(ctx, registry, opts)
	return args.Get(0).([]string), args.Error(1)
}

// MockArtifactRegistryClient implements a mock Artifact Registry client
type MockArtifactRegistryClient struct {
	mock.Mock
}

func (m *MockArtifactRegistryClient) ListRepositories(ctx context.Context, parent string) *MockRepositoryIterator {
	args := m.Called(ctx, parent)
	return args.Get(0).(*MockRepositoryIterator)
}

func (m *MockArtifactRegistryClient) GetRepository(ctx context.Context, name string) (*artifactregistry.Repository, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*artifactregistry.Repository), args.Error(1)
}

func (m *MockArtifactRegistryClient) CreateRepository(ctx context.Context, parent string, repository *artifactregistry.Repository) (*artifactregistry.Operation, error) {
	args := m.Called(ctx, parent, repository)
	return args.Get(0).(*artifactregistry.Operation), args.Error(1)
}

// MockRepositoryIterator implements a mock repository iterator
type MockRepositoryIterator struct {
	mock.Mock
	repos []*artifactregistry.Repository
	index int
	err   error
}

func NewMockRepositoryIterator(repos []*artifactregistry.Repository, err error) *MockRepositoryIterator {
	return &MockRepositoryIterator{
		repos: repos,
		index: 0,
		err:   err,
	}
}

func (m *MockRepositoryIterator) Next() (*artifactregistry.Repository, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.index >= len(m.repos) {
		return nil, ErrIteratorDone
	}

	repo := m.repos[m.index]
	m.index++
	return repo, nil
}

func (m *MockRepositoryIterator) PageInfo() *iterator.PageInfo {
	// Return a mock PageInfo - not used in our tests
	return &iterator.PageInfo{}
}

// Helper functions to create mock responses

// CreateMockGCRRepositories creates sample GCR repositories for testing
func CreateMockGCRRepositories(projectID string, count int) []string {
	repos := make([]string, count)

	for i := 0; i < count; i++ {
		repos[i] = fmt.Sprintf("%s/test-repo-%d", projectID, i+1)
	}

	return repos
}

// CreateMockArtifactRepositories creates sample Artifact Registry repositories
func CreateMockArtifactRepositories(projectID, location string, count int) []*artifactregistry.Repository {
	repos := make([]*artifactregistry.Repository, count)

	for i := 0; i < count; i++ {
		repoID := fmt.Sprintf("test-repo-%d", i+1)
		repos[i] = &artifactregistry.Repository{
			Name:        fmt.Sprintf("projects/%s/locations/%s/repositories/%s", projectID, location, repoID),
			Format:      "DOCKER",
			Description: fmt.Sprintf("Test repository %d", i+1),
			CreateTime:  time.Now().Format(time.RFC3339),
			UpdateTime:  time.Now().Format(time.RFC3339),
		}
	}

	return repos
}

// CreateMockHTTPResponse creates a mock HTTP response for authentication testing
func CreateMockHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       &MockReadCloser{content: body},
		Header:     make(http.Header),
	}
}

// MockReadCloser implements io.ReadCloser for mock HTTP responses
type MockReadCloser struct {
	content string
	pos     int
}

func (m *MockReadCloser) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.content) {
		return 0, fmt.Errorf("EOF")
	}

	n = copy(p, m.content[m.pos:])
	m.pos += n
	return n, nil
}

func (m *MockReadCloser) Close() error {
	return nil
}

// GCRMockBuilder provides a fluent interface for setting up GCR mock expectations
type GCRMockBuilder struct {
	catalogClient *MockGoogleCatalogClient
	authTransport *MockGoogleAuth
}

// NewGCRMockBuilder creates a new GCR mock builder
func NewGCRMockBuilder() *GCRMockBuilder {
	return &GCRMockBuilder{
		catalogClient: &MockGoogleCatalogClient{},
		authTransport: &MockGoogleAuth{},
	}
}

// ExpectCatalog sets up expectations for Catalog calls
func (b *GCRMockBuilder) ExpectCatalog(repos []string, err error) *GCRMockBuilder {
	b.catalogClient.On("Catalog", mock.Anything, mock.Anything, mock.Anything).Return(repos, err)
	return b
}

// ExpectAuthSuccess sets up expectations for successful authentication
func (b *GCRMockBuilder) ExpectAuthSuccess() *GCRMockBuilder {
	response := CreateMockHTTPResponse(200, `{"access_token": "mock-token", "token_type": "Bearer"}`)
	b.authTransport.On("RoundTrip", mock.Anything).Return(response, nil)
	return b
}

// ExpectAuthFailure sets up expectations for failed authentication
func (b *GCRMockBuilder) ExpectAuthFailure(err error) *GCRMockBuilder {
	b.authTransport.On("RoundTrip", mock.Anything).Return((*http.Response)(nil), err)
	return b
}

// BuildCatalogClient returns the configured mock catalog client
func (b *GCRMockBuilder) BuildCatalogClient() *MockGoogleCatalogClient {
	return b.catalogClient
}

// BuildAuthTransport returns the configured mock auth transport
func (b *GCRMockBuilder) BuildAuthTransport() *MockGoogleAuth {
	return b.authTransport
}

// ArtifactRegistryMockBuilder provides a fluent interface for Artifact Registry mocks
type ArtifactRegistryMockBuilder struct {
	client *MockArtifactRegistryClient
}

// NewArtifactRegistryMockBuilder creates a new Artifact Registry mock builder
func NewArtifactRegistryMockBuilder() *ArtifactRegistryMockBuilder {
	return &ArtifactRegistryMockBuilder{
		client: &MockArtifactRegistryClient{},
	}
}

// ExpectListRepositories sets up expectations for ListRepositories calls
func (b *ArtifactRegistryMockBuilder) ExpectListRepositories(repos []*artifactregistry.Repository, err error) *ArtifactRegistryMockBuilder {
	iterator := NewMockRepositoryIterator(repos, err)
	b.client.On("ListRepositories", mock.Anything, mock.Anything).Return(iterator)
	return b
}

// ExpectGetRepository sets up expectations for GetRepository calls
func (b *ArtifactRegistryMockBuilder) ExpectGetRepository(repo *artifactregistry.Repository, err error) *ArtifactRegistryMockBuilder {
	b.client.On("GetRepository", mock.Anything, mock.Anything).Return(repo, err)
	return b
}

// Build returns the configured mock client
func (b *ArtifactRegistryMockBuilder) Build() *MockArtifactRegistryClient {
	return b.client
}

// MockGCRTestScenarios provides common test scenarios for GCR testing
type MockGCRTestScenarios struct{}

// SuccessfulGCRListRepositories returns mocks configured for successful repository listing
func (s *MockGCRTestScenarios) SuccessfulGCRListRepositories(projectID string, count int) (*MockGoogleCatalogClient, *MockGoogleAuth) {
	repos := CreateMockGCRRepositories(projectID, count)

	builder := NewGCRMockBuilder()
	builder.ExpectCatalog(repos, nil)
	builder.ExpectAuthSuccess()

	return builder.BuildCatalogClient(), builder.BuildAuthTransport()
}

// FailedGCRAuthentication returns mocks configured for authentication failure
func (s *MockGCRTestScenarios) FailedGCRAuthentication() (*MockGoogleCatalogClient, *MockGoogleAuth) {
	builder := NewGCRMockBuilder()
	builder.ExpectAuthFailure(fmt.Errorf("authentication failed"))

	return builder.BuildCatalogClient(), builder.BuildAuthTransport()
}

// SuccessfulArtifactRegistryList returns mocks for successful Artifact Registry listing
func (s *MockGCRTestScenarios) SuccessfulArtifactRegistryList(projectID, location string, count int) *MockArtifactRegistryClient {
	repos := CreateMockArtifactRepositories(projectID, location, count)

	builder := NewArtifactRegistryMockBuilder()
	builder.ExpectListRepositories(repos, nil)

	return builder.Build()
}
