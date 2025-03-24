package gcr

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	
	"github.com/elad/freightliner/internal/log"
	"github.com/elad/freightliner/pkg/client/common"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"google.golang.org/api/iterator"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// Client implements the registry client interface for Google GCR
type Client struct {
	project       string
	location      string
	keychain      authn.Keychain
	transportOpt  remote.Option
	logger        *log.Logger
	arClient      *artifactregistry.Service // Artifact Registry client for listing repositories
}

// ClientOptions contains options for the GCR client
type ClientOptions struct {
	Project  string
	Location string // Optional, defaults to "us"
	Logger   *log.Logger
}

// NewClient creates a new GCR client
func NewClient(opts ClientOptions) (*Client, error) {
	ctx := context.Background()
	
	// Default location to "us" if not provided
	location := opts.Location
	if location == "" {
		location = "us"
	}
	
	// Create keychain for GCR authentication
	keychain, err := NewGCRKeychain()
	if err != nil {
		return nil, fmt.Errorf("failed to create GCR keychain: %w", err)
	}
	
	// Create transport option for remote.* operations
	transportOpt := remote.WithAuthFromKeychain(keychain)
	
	// Create Artifact Registry client for listing repositories
	arClient, err := artifactregistry.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Artifact Registry client: %w", err)
	}
	
	return &Client{
		project:      opts.Project,
		location:     location,
		keychain:     keychain,
		transportOpt: transportOpt,
		logger:       opts.Logger,
		arClient:     arClient,
	}, nil
}

// GetRepository returns a repository interface for the given repository name
func (c *Client) GetRepository(name string) (common.Repository, error) {
	// Parse repository name
	_, repo, err := ParseGCRRepository(name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repository name: %w", err)
	}
	
	// Unlike ECR, GCR/AR doesn't need explicit creation - repositories are created on first push
	
	return &Repository{
		client:     c,
		name:       name,
		repository: repo,
	}, nil
}

// ListRepositories returns a list of all repositories in the registry
func (c *Client) ListRepositories() ([]string, error) {
	ctx := context.Background()
	var repositories []string
	
	// First try using the Google Container Registry approach
	repos, err := google.List(fmt.Sprintf("gcr.io/%s", c.project), c.transportOpt)
	if err != nil {
		c.logger.Warn("Failed to list GCR repositories, trying Artifact Registry", map[string]interface{}{
			"error": err.Error(),
		})
		
		// Fall back to Artifact Registry API
		parent := fmt.Sprintf("projects/%s/locations/%s", c.project, c.location)
		it := c.arClient.Projects.Locations.Repositories.List(parent).Context(ctx)
		
		for {
			repo, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to list repositories: %w", err)
			}
			
			// Extract repository name from full path
			// Format: projects/PROJECT/locations/LOCATION/repositories/REPO
			parts := strings.Split(repo.Name, "/")
			if len(parts) > 0 {
				repositories = append(repositories, parts[len(parts)-1])
			}
		}
	} else {
		// Process GCR repos
		for _, r := range repos {
			// Extract just the repository name
			parts := strings.Split(r, "/")
			if len(parts) > 1 {
				repositories = append(repositories, strings.Join(parts[1:], "/"))
			}
		}
	}
	
	return repositories, nil
}

// GetTransport returns a transport for the given repository
func (c *Client) GetTransport(repo string) (http.RoundTripper, error) {
	// Parse repository name
	_, repository, err := ParseGCRRepository(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repository name: %w", err)
	}
	
	// Get authenticator for this repository
	auth, err := c.keychain.Resolve(repository)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticator: %w", err)
	}
	
	// Create transport
	return transport.NewWithContext(
		context.Background(),
		repository.Registry,
		auth,
		transport.WithContext(context.Background()),
	)
}
