// Package gcr provides Google Container Registry client functionality.
package gcr

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	// LocationAsia represents the Asia GCR location
	LocationAsia = "asia"
	// LocationEU represents the European GCR location
	LocationEU = "eu"
	// LocationUS represents the US GCR location
	LocationUS = "us"
)

// Client implements the registry client interface for Google Container Registry
type Client struct {
	logger         log.Logger
	project        string
	location       string
	arClient       *artifactregistry.Service
	transportOpt   remote.Option
	googleAuthOpts []google.Option
}

// ClientOptions provides configuration for connecting to GCR
type ClientOptions struct {
	// Project is the GCP project ID
	Project string

	// Location is the GCR location (us, eu, asia)
	Location string

	// Logger is the logger to use
	Logger log.Logger

	// CredentialsFile is the path to a Google service account JSON key file
	CredentialsFile string
}

// GetRegistryName returns the registry hostname for this client
func (c *Client) GetRegistryName() string {
	if c.location == LocationUS {
		return "gcr.io"
	}
	if c.location == LocationEU {
		return "eu.gcr.io"
	}
	if c.location == LocationAsia {
		return "asia.gcr.io"
	}
	return fmt.Sprintf("%s-docker.pkg.dev", c.location)
}

// NewClient creates a new GCR client
func NewClient(opts ClientOptions) (*Client, error) {
	// Set default values
	if opts.Location == "" {
		opts.Location = "us"
	}

	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	var arOpts []option.ClientOption
	var googleOpts []google.Option
	var transportOpt remote.Option

	// If credentials file is provided, use it
	if opts.CredentialsFile != "" {
		arOpts = append(arOpts, option.WithCredentialsFile(opts.CredentialsFile))
		googleOpts = append(googleOpts, google.WithTransport(http.DefaultTransport))
		transportOpt = remote.WithAuth(&gcrCredentialHelper{
			credentialsFile: opts.CredentialsFile,
		})
	} else {
		// Use default authentication methods
		transportOpt = remote.WithAuthFromKeychain(google.Keychain)
	}

	// Create Artifact Registry client
	ctx := context.Background()
	arService, err := artifactregistry.NewService(ctx, arOpts...)
	if err != nil {
		opts.Logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Failed to create Artifact Registry client, some functionality may be limited")
		// We'll continue without the AR client - we can still use the GCR API
		arService = nil
	}

	return &Client{
		logger:         opts.Logger,
		project:        opts.Project,
		location:       opts.Location,
		arClient:       arService,
		transportOpt:   transportOpt,
		googleAuthOpts: googleOpts,
	}, nil
}

// GetRepository returns a repository by name
func (c *Client) GetRepository(ctx context.Context, repoName string) (interfaces.Repository, error) {
	// Input validation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Create the GCR repository reference
	registry := fmt.Sprintf("gcr.io/%s", c.project)
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", registry, repoName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Create and return the repository object
	repo := &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}

	return repo, nil
}

// CreateRepository creates a new repository in GCR - implements interfaces.RepositoryCreator
// Note: In GCR/Artifact Registry, repositories are created automatically when the first image is pushed
// This method essentially validates the repository name and returns a repository reference
// The tags parameter is accepted for interface compatibility but not used in GCR
func (c *Client) CreateRepository(ctx context.Context, repoName string, _ map[string]string) (interfaces.Repository, error) {
	// Input validation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	c.logger.WithFields(map[string]interface{}{
		"repository": repoName,
		"project":    c.project,
		"location":   c.location,
	}).Info("Creating repository in GCR")

	// Create the GCR repository reference
	registry := fmt.Sprintf("gcr.io/%s", c.project)
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", registry, repoName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// For GCR/Artifact Registry, repositories are automatically created when the first image is pushed
	// We don't need to make any API calls to "create" the repository
	// Just validate that the repository name is valid and return a repository object

	// Create and return the repository object
	repo := &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}

	c.logger.WithFields(map[string]interface{}{
		"repository": repoName,
		"registry":   registry,
	}).Info("Repository reference created (repository will be created automatically on first push)")

	return repo, nil
}

// ListRepositories lists all repositories in the registry with the given prefix
func (c *Client) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	// Input validation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Check if we have an AR client (preferred approach)
	if c.arClient != nil {
		return c.listRepositoriesViaAR(ctx, prefix)
	}

	// Fallback to direct GCR API
	return c.listRepositoriesViaGCR(ctx, prefix)
}

// listRepositoriesViaAR uses the Artifact Registry API to list repositories
func (c *Client) listRepositoriesViaAR(_ context.Context, prefix string) ([]string, error) {
	// Determine the location parameter
	location := c.location
	if location == LocationUS || location == LocationEU || location == LocationAsia {
		location = "us-central1" // Map legacy locations to GCP regions
	}

	// Create the parent parameter for the API call
	// Format: projects/{project}/locations/{location}
	parent := fmt.Sprintf("projects/%s/locations/%s", c.project, location)

	// List repositories in the project/location
	c.logger.WithFields(map[string]interface{}{
		"project":  c.project,
		"location": location,
		"prefix":   prefix,
	}).Debug("Listing repositories via Artifact Registry API")

	// Create request with filter if needed
	req := c.arClient.Projects.Locations.Repositories.List(parent)
	if prefix != "" {
		// Filter format: name:*{prefix}*
		req = req.Filter(fmt.Sprintf("name:*%s*", prefix))
	}

	// Use iterator pattern to list all repositories
	it := makeRepositoryIterator(req)
	repositories := make([]string, 0, 10) // Pre-allocate for common case

	for {
		repo, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to list repositories")
		}

		// Extract repository name from the full path
		// Format: projects/{project}/locations/{location}/repositories/{repository}
		parts := strings.Split(repo.Name, "/")
		if len(parts) > 0 {
			repoName := parts[len(parts)-1]

			// Apply prefix filtering manually in case AR API filter doesn't work as expected
			if prefix == "" || strings.HasPrefix(repoName, prefix) {
				repositories = append(repositories, repoName)
			}
		}
	}

	return repositories, nil
}

// listRepositoriesViaGCR uses the direct GCR API to list repositories
func (c *Client) listRepositoriesViaGCR(ctx context.Context, prefix string) ([]string, error) {
	// Determine registry path based on location
	var registryPath string
	if c.location == LocationUS {
		registryPath = fmt.Sprintf("gcr.io/%s", c.project)
	} else if c.location == LocationEU {
		registryPath = fmt.Sprintf("eu.gcr.io/%s", c.project)
	} else if c.location == LocationAsia {
		registryPath = fmt.Sprintf("asia.gcr.io/%s", c.project)
	} else {
		registryPath = fmt.Sprintf("%s-docker.pkg.dev/%s", c.location, c.project)
	}

	c.logger.WithFields(map[string]interface{}{
		"registry": registryPath,
		"prefix":   prefix,
	}).Debug("Listing repositories via GCR API")

	// Create a repository reference for the registry
	registry, err := name.NewRepository(registryPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create registry reference")
	}

	// Use the google.List function to list repositories
	tags, err := google.List(registry, c.googleAuthOpts...)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			// Registry might be empty or not exist yet
			return []string{}, nil
		}
		return nil, errors.Wrap(err, "failed to list repositories")
	}

	// Extract repository names from the registry catalog
	repoMap := make(map[string]bool) // Use map to deduplicate
	for repoName := range tags.Manifests {
		// Get repository part from full image name
		parts := strings.Split(repoName, "/")
		if len(parts) > 1 {
			// Skip the registry part and join the rest
			repo := strings.Join(parts[1:], "/")

			// Apply prefix filtering
			if prefix == "" || strings.HasPrefix(repo, prefix) {
				repoMap[repo] = true
			}
		}
	}

	// Convert map to slice
	repositories := make([]string, 0, len(repoMap))
	for repo := range repoMap {
		repositories = append(repositories, repo)
	}

	return repositories, nil
}

// GetTransport returns an authenticated HTTP transport for GCR
func (c *Client) GetTransport(repositoryName string) (http.RoundTripper, error) {
	// Create the GCR repository reference
	registry := fmt.Sprintf("gcr.io/%s", c.project)
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", registry, repositoryName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Get the authenticator from the Google keychain
	auth, err := google.Keychain.Resolve(repository.Registry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get authenticator")
	}

	// Create the transport
	rt, err := transport.NewWithContext(
		context.Background(),
		repository.Registry,
		auth,
		http.DefaultTransport,
		[]string{repository.Scope(transport.PushScope)},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transport")
	}

	return rt, nil
}

// gcrCredentialHelper implements the authn.Authenticator interface for GCR
type gcrCredentialHelper struct {
	credentialsFile string
}

// Authorization returns the Authorization header value to use in registry requests
func (h *gcrCredentialHelper) Authorization() (*authn.AuthConfig, error) {
	// Delegate to the Google Keychain
	registry, err := name.NewRegistry(name.DefaultRegistry)
	if err != nil {
		return nil, err
	}

	auth, err := google.Keychain.Resolve(registry)
	if err != nil {
		return nil, err
	}
	return auth.Authorization()
}

// repositoryIterator is a wrapper around the AR ListRepositoriesCall
type repositoryIterator struct {
	items     []*artifactregistry.Repository
	request   *artifactregistry.ProjectsLocationsRepositoriesListCall
	pageToken string
	pageSize  int64
	index     int
}

// makeRepositoryIterator creates a new repository iterator
func makeRepositoryIterator(request *artifactregistry.ProjectsLocationsRepositoriesListCall) *repositoryIterator {
	return &repositoryIterator{
		request:  request,
		pageSize: 50, // Default page size
		index:    -1,
	}
}

// Next returns the next repository
func (it *repositoryIterator) Next() (*artifactregistry.Repository, error) {
	it.index++

	// If we have items and haven't reached the end, return the next item
	if it.items != nil && it.index < len(it.items) {
		return it.items[it.index], nil
	}

	// Otherwise, fetch the next page
	it.index = 0
	it.items = nil

	// Set up the page request
	req := it.request
	if it.pageToken != "" {
		req = req.PageToken(it.pageToken)
	}
	if it.pageSize > 0 {
		req = req.PageSize(it.pageSize)
	}

	// Execute the request
	resp, err := req.Do()
	if err != nil {
		return nil, err
	}

	// Update state for next page
	it.pageToken = resp.NextPageToken
	it.items = resp.Repositories

	// Check if we're at the end
	if len(it.items) == 0 {
		return nil, iterator.Done
	}

	return it.items[0], nil
}
