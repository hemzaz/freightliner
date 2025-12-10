// Package quay provides Quay.io Registry client functionality.
package quay

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

// Client implements the registry client interface for Quay.io
type Client struct {
	registryURL  string
	apiURL       string
	auth         *QuayAuthenticator
	logger       log.Logger
	transportOpt remote.Option
	httpClient   *http.Client
}

// ClientOptions provides configuration for connecting to Quay
type ClientOptions struct {
	// RegistryURL is the Quay registry URL (default: quay.io)
	RegistryURL string

	// APIVersion is the Quay API version (default: v1)
	APIVersion string

	// AuthConfig contains authentication configuration
	AuthConfig *AuthConfig

	// Logger is the logger to use
	Logger log.Logger

	// Username for basic authentication
	Username string

	// Password for basic authentication
	Password string

	// RobotUsername for robot account authentication
	RobotUsername string

	// RobotToken for robot account authentication
	RobotToken string

	// OAuth2Token for OAuth2 authentication
	OAuth2Token string

	// Organization is the Quay organization name
	Organization string

	// Insecure allows insecure connections (for testing)
	Insecure bool
}

// NewClient creates a new Quay client
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Set default registry URL
	registryURL := opts.RegistryURL
	if registryURL == "" {
		registryURL = "quay.io"
	}

	// Normalize registry URL
	registryURL = strings.TrimPrefix(registryURL, "https://")
	registryURL = strings.TrimPrefix(registryURL, "http://")
	registryURL = strings.TrimSuffix(registryURL, "/")

	// Create auth config if not provided
	var authConfig *AuthConfig
	if opts.AuthConfig != nil {
		authConfig = opts.AuthConfig
	} else {
		authConfig = &AuthConfig{
			RegistryURL:  registryURL,
			Organization: opts.Organization,
		}

		// Configure authentication based on provided credentials
		if opts.RobotUsername != "" && opts.RobotToken != "" {
			authConfig.Type = AuthTypeRobot
			authConfig.RobotUsername = opts.RobotUsername
			authConfig.RobotToken = opts.RobotToken
			authConfig.Username = opts.RobotUsername
			authConfig.Password = opts.RobotToken
		} else if opts.OAuth2Token != "" {
			authConfig.Type = AuthTypeOAuth2
			authConfig.OAuth2Token = opts.OAuth2Token
		} else if opts.Username != "" && opts.Password != "" {
			authConfig.Type = AuthTypeBasic
			authConfig.Username = opts.Username
			authConfig.Password = opts.Password
		} else {
			return nil, errors.InvalidInputf("authentication credentials required")
		}
	}

	// Create authenticator
	auth, err := NewQuayAuthenticator(authConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Quay authenticator")
	}

	// Determine API URL
	apiVersion := opts.APIVersion
	if apiVersion == "" {
		apiVersion = "v1"
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

// GetRegistryName returns the Quay registry endpoint
func (c *Client) GetRegistryName() string {
	return c.registryURL
}

// ListRepositories lists all repositories in the registry
func (c *Client) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	// Use Quay API to list repositories
	apiURL := fmt.Sprintf("%s/repository", c.apiURL)

	// Add query parameters
	params := url.Values{}
	if prefix != "" {
		params.Set("namespace", prefix)
	}
	params.Set("public", "false") // Include private repositories

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

	if authConfig.IdentityToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authConfig.IdentityToken))
	} else if authConfig.Username != "" && authConfig.Password != "" {
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
	var result struct {
		Repositories []struct {
			Namespace   string `json:"namespace"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"repositories"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "failed to parse repositories response")
	}

	// Extract repository names
	var repositories []string
	for _, repo := range result.Repositories {
		repoName := fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)
		if prefix == "" || strings.HasPrefix(repoName, prefix) {
			repositories = append(repositories, repoName)
		}
	}

	return repositories, nil
}

// GetRepository returns a repository by name
func (c *Client) GetRepository(ctx context.Context, repoName string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Create a proper repository reference
	repoPath := fmt.Sprintf("%s/%s", c.registryURL, repoName)
	repository, err := name.NewRepository(repoPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	return &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}, nil
}

// CreateRepository creates a new repository in Quay
func (c *Client) CreateRepository(ctx context.Context, repoName string, tags map[string]string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	c.logger.WithFields(map[string]interface{}{
		"repository": repoName,
		"registry":   c.registryURL,
	}).Info("Creating repository in Quay")

	// Parse repository name to get namespace and name
	parts := strings.SplitN(repoName, "/", 2)
	if len(parts) != 2 {
		return nil, errors.InvalidInputf("repository name must be in format: namespace/name")
	}

	namespace := parts[0]
	repoBaseName := parts[1]

	// Create repository via API
	apiURL := fmt.Sprintf("%s/repository", c.apiURL)

	reqBody := map[string]interface{}{
		"namespace":   namespace,
		"repository":  repoBaseName,
		"visibility":  "private",
		"description": tags["description"],
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository request")
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication
	authConfig, err := c.auth.Authorization()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get authorization")
	}

	if authConfig.IdentityToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authConfig.IdentityToken))
	} else if authConfig.Username != "" && authConfig.Password != "" {
		req.SetBasicAuth(authConfig.Username, authConfig.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.InvalidInputf("create repository failed: %s - %s", resp.Status, string(body))
	}

	// Create the repository reference
	repoPath := fmt.Sprintf("%s/%s", c.registryURL, repoName)
	repository, err := name.NewRepository(repoPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	c.logger.WithFields(map[string]interface{}{
		"repository": repoName,
		"registry":   c.registryURL,
	}).Info("Successfully created repository in Quay")

	return &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}, nil
}

// GetTransport returns an authenticated HTTP transport for Quay
func (c *Client) GetTransport(repositoryName string) (http.RoundTripper, error) {
	// Create a proper repository reference
	repoPath := fmt.Sprintf("%s/%s", c.registryURL, repositoryName)
	repository, err := name.NewRepository(repoPath)
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
		return nil, errors.Wrap(err, "failed to create Quay transport")
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

// GetAPIURL returns the Quay API URL
func (c *Client) GetAPIURL() string {
	return c.apiURL
}
