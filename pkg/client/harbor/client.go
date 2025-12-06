// Package harbor provides Harbor Registry client functionality.
package harbor

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// Client implements the registry client interface for Harbor
type Client struct {
	registryURL  string
	apiURL       string
	auth         *HarborAuthenticator
	logger       log.Logger
	transportOpt remote.Option
	httpClient   *http.Client
}

// ClientOptions provides configuration for connecting to Harbor
type ClientOptions struct {
	// RegistryURL is the Harbor registry URL (e.g., harbor.example.com)
	RegistryURL string

	// APIVersion is the Harbor API version (default: v2.0)
	APIVersion string

	// AuthConfig contains authentication configuration
	AuthConfig *AuthConfig

	// Logger is the logger to use
	Logger log.Logger

	// Username for basic authentication
	Username string

	// Password for basic authentication
	Password string

	// RobotName for robot account authentication
	RobotName string

	// RobotToken for robot account authentication
	RobotToken string

	// ProjectName is the default Harbor project
	ProjectName string

	// Insecure allows insecure connections (for testing)
	Insecure bool
}

// NewClient creates a new Harbor client
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.RegistryURL == "" {
		return nil, errors.InvalidInputf("registry URL is required")
	}

	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Normalize registry URL
	registryURL := strings.TrimPrefix(opts.RegistryURL, "https://")
	registryURL = strings.TrimPrefix(registryURL, "http://")
	registryURL = strings.TrimSuffix(registryURL, "/")

	// Create auth config if not provided
	var authConfig *AuthConfig
	if opts.AuthConfig != nil {
		authConfig = opts.AuthConfig
	} else {
		authConfig = &AuthConfig{
			RegistryURL: registryURL,
			ProjectName: opts.ProjectName,
		}

		// Configure authentication based on provided credentials
		if opts.RobotName != "" && opts.RobotToken != "" {
			authConfig.Type = AuthTypeRobot
			authConfig.RobotName = opts.RobotName
			authConfig.RobotToken = opts.RobotToken
			authConfig.Username = opts.RobotName
			authConfig.Password = opts.RobotToken
		} else if opts.Username != "" && opts.Password != "" {
			authConfig.Type = AuthTypeBasic
			authConfig.Username = opts.Username
			authConfig.Password = opts.Password
		} else {
			return nil, errors.InvalidInputf("authentication credentials required")
		}
	}

	// Create authenticator
	auth, err := NewHarborAuthenticator(authConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Harbor authenticator")
	}

	// Determine API URL
	apiVersion := opts.APIVersion
	if apiVersion == "" {
		apiVersion = "v2.0"
	}
	apiURL := fmt.Sprintf("https://%s/api/%s", registryURL, apiVersion)

	return &Client{
		registryURL:  registryURL,
		apiURL:       apiURL,
		auth:         auth,
		logger:       opts.Logger,
		transportOpt: remote.WithAuth(auth),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: opts.Insecure,
				},
			},
		},
	}, nil
}

// GetRegistryName returns the Harbor registry endpoint
func (c *Client) GetRegistryName() string {
	return c.registryURL
}

// ListRepositories lists all repositories in the registry
func (c *Client) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	// Use Harbor API to list repositories
	apiURL := fmt.Sprintf("%s/repositories", c.apiURL)

	// Add query parameters
	params := url.Values{}
	if prefix != "" {
		params.Set("q", fmt.Sprintf("name~%s", prefix))
	}
	params.Set("page_size", "100") // Adjust as needed

	if len(params) > 0 {
		apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create list request")
	}

	// Add authentication
	authConfig, err := c.auth.Authorization()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get authorization")
	}

	if authConfig.Username != "" && authConfig.Password != "" {
		req.SetBasicAuth(authConfig.Username, authConfig.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list repositories")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.InvalidInputf("list repositories failed: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var repos []struct {
		Name      string `json:"name"`
		ProjectID int    `json:"project_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, errors.Wrap(err, "failed to parse repositories response")
	}

	// Extract repository names
	var repositories []string
	for _, repo := range repos {
		repositories = append(repositories, repo.Name)
	}

	return repositories, nil
}

// GetRepository returns a repository by name
func (c *Client) GetRepository(ctx context.Context, repoName string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Create a proper repository reference
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", c.registryURL, repoName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	return &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}, nil
}

// CreateRepository creates a new repository in Harbor
// Note: In Harbor, repositories are typically created as part of a project
func (c *Client) CreateRepository(ctx context.Context, repoName string, tags map[string]string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	c.logger.WithFields(map[string]interface{}{
		"repository": repoName,
		"registry":   c.registryURL,
	}).Info("Creating repository reference (Harbor auto-creates on first push)")

	// Create the repository reference
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", c.registryURL, repoName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Note: Harbor repositories are created automatically on first push
	// Projects must exist before pushing

	return &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}, nil
}

// GetTransport returns an authenticated HTTP transport for Harbor
func (c *Client) GetTransport(repositoryName string) (http.RoundTripper, error) {
	// Create a proper repository reference
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", c.registryURL, repositoryName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Create transport with authentication
	rt, err := transport.NewWithContext(
		context.Background(),
		repository.Registry,
		c.auth,
		http.DefaultTransport,
		[]string{repository.Scope(transport.PushScope)},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Harbor transport")
	}

	return rt, nil
}

// GetRemoteOptions returns options for the go-containerregistry remote package
func (c *Client) GetRemoteOptions() []remote.Option {
	return []remote.Option{
		c.transportOpt,
	}
}

// RefreshAuth refreshes the authentication token
func (c *Client) RefreshAuth() error {
	return c.auth.RefreshToken()
}

// GetAPIURL returns the Harbor API URL
func (c *Client) GetAPIURL() string {
	return c.apiURL
}
