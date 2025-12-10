package harbor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"freightliner/pkg/helper/errors"

	"github.com/google/go-containerregistry/pkg/authn"
)

// AuthType represents the type of authentication to use
type AuthType string

const (
	// AuthTypeBasic uses basic username/password authentication
	AuthTypeBasic AuthType = "basic"
	// AuthTypeRobot uses Harbor robot account authentication
	AuthTypeRobot AuthType = "robot"
	// AuthTypeBearer uses Bearer token authentication
	AuthTypeBearer AuthType = "bearer"
)

// AuthConfig contains authentication configuration for Harbor
type AuthConfig struct {
	// Type is the authentication type
	Type AuthType

	// Username for basic or robot authentication
	Username string

	// Password for basic or robot authentication
	Password string

	// Token for bearer authentication
	Token string

	// RobotName is the robot account name (typically robot$name)
	RobotName string

	// RobotToken is the robot account token
	RobotToken string

	// ProjectName is the Harbor project name
	ProjectName string

	// RegistryURL is the Harbor registry URL
	RegistryURL string
}

// TokenResponse represents a Harbor OAuth token response
type TokenResponse struct {
	Token        string `json:"token"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	IssuedAt     string `json:"issued_at"`
	RefreshToken string `json:"refresh_token"`
}

// HarborAuthenticator implements authn.Authenticator for Harbor
type HarborAuthenticator struct {
	config     *AuthConfig
	tokenCache *tokenCache
	httpClient *http.Client
	mu         sync.RWMutex
}

// tokenCache stores cached tokens with expiration
type tokenCache struct {
	token     string
	expiresAt time.Time
	mu        sync.RWMutex
}

// NewHarborAuthenticator creates a new Harbor authenticator
func NewHarborAuthenticator(config *AuthConfig) (*HarborAuthenticator, error) {
	if config == nil {
		return nil, errors.InvalidInputf("auth config is required")
	}

	if config.RegistryURL == "" {
		return nil, errors.InvalidInputf("registry URL is required")
	}

	// Validate authentication credentials based on type
	switch config.Type {
	case AuthTypeBasic:
		if config.Username == "" || config.Password == "" {
			return nil, errors.InvalidInputf("username and password required for basic auth")
		}
	case AuthTypeRobot:
		if config.RobotName == "" || config.RobotToken == "" {
			return nil, errors.InvalidInputf("robot name and token required for robot auth")
		}
		// Set username/password for robot account
		config.Username = config.RobotName
		config.Password = config.RobotToken
	case AuthTypeBearer:
		if config.Token == "" {
			return nil, errors.InvalidInputf("token required for bearer auth")
		}
	default:
		config.Type = AuthTypeBasic // Default to basic auth
	}

	return &HarborAuthenticator{
		config:     config,
		tokenCache: &tokenCache{},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Authorization returns the authorization configuration for Harbor
func (a *HarborAuthenticator) Authorization() (*authn.AuthConfig, error) {
	switch a.config.Type {
	case AuthTypeBasic, AuthTypeRobot:
		// For basic and robot auth, return credentials directly
		return &authn.AuthConfig{
			Username: a.config.Username,
			Password: a.config.Password,
		}, nil

	case AuthTypeBearer:
		// For bearer token, check cache first
		token, err := a.getAccessToken()
		if err != nil {
			return nil, err
		}

		return &authn.AuthConfig{
			Username:      a.config.Username,
			Password:      a.config.Password,
			IdentityToken: token,
		}, nil

	default:
		return nil, errors.InvalidInputf("unsupported auth type: %s", a.config.Type)
	}
}

// getAccessToken retrieves an access token, using cache if valid
func (a *HarborAuthenticator) getAccessToken() (string, error) {
	a.tokenCache.mu.RLock()
	// Check if cached token is still valid (with 5 minute buffer)
	if a.tokenCache.token != "" && time.Now().Add(5*time.Minute).Before(a.tokenCache.expiresAt) {
		token := a.tokenCache.token
		a.tokenCache.mu.RUnlock()
		return token, nil
	}
	a.tokenCache.mu.RUnlock()

	a.tokenCache.mu.Lock()
	defer a.tokenCache.mu.Unlock()

	// Double-check after acquiring write lock
	if a.tokenCache.token != "" && time.Now().Add(5*time.Minute).Before(a.tokenCache.expiresAt) {
		return a.tokenCache.token, nil
	}

	// Fetch new token
	token, expiresAt, err := a.fetchToken()
	if err != nil {
		return "", err
	}

	// Cache the token
	a.tokenCache.token = token
	a.tokenCache.expiresAt = expiresAt

	return token, nil
}

// fetchToken fetches a new access token from Harbor
func (a *HarborAuthenticator) fetchToken() (string, time.Time, error) {
	// Harbor token endpoint
	tokenURL := fmt.Sprintf("%s/service/token", a.config.RegistryURL)

	// Prepare request parameters
	params := url.Values{}
	params.Set("service", "harbor-registry")
	if a.config.ProjectName != "" {
		params.Set("scope", fmt.Sprintf("repository:%s:pull,push", a.config.ProjectName))
	}

	// Create request
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("%s?%s", tokenURL, params.Encode()),
		nil,
	)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to create token request")
	}

	// Add basic auth header
	if a.config.Username != "" && a.config.Password != "" {
		auth := base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", a.config.Username, a.config.Password)),
		)
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
	}

	// Execute request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to fetch token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", time.Time{}, errors.InvalidInputf("token request failed: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to parse token response")
	}

	// Determine which token to use
	token := tokenResp.Token
	if token == "" {
		token = tokenResp.AccessToken
	}

	// Calculate expiration
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	if tokenResp.ExpiresIn == 0 {
		// Default to 1 hour if not specified
		expiresAt = time.Now().Add(1 * time.Hour)
	}

	return token, expiresAt, nil
}

// RefreshToken forces a token refresh
func (a *HarborAuthenticator) RefreshToken() error {
	a.tokenCache.mu.Lock()
	defer a.tokenCache.mu.Unlock()

	// Clear cached token
	a.tokenCache.token = ""
	a.tokenCache.expiresAt = time.Time{}

	return nil
}

// GetRegistryURL returns the Harbor registry URL
func (a *HarborAuthenticator) GetRegistryURL() string {
	return a.config.RegistryURL
}

// AuthConfigOption is a functional option for AuthConfig
type AuthConfigOption func(*AuthConfig)

// WithBasicAuth configures basic authentication
func WithBasicAuth(username, password string) AuthConfigOption {
	return func(c *AuthConfig) {
		c.Type = AuthTypeBasic
		c.Username = username
		c.Password = password
	}
}

// WithRobotAccount configures robot account authentication
func WithRobotAccount(robotName, robotToken string) AuthConfigOption {
	return func(c *AuthConfig) {
		c.Type = AuthTypeRobot
		c.RobotName = robotName
		c.RobotToken = robotToken
		c.Username = robotName
		c.Password = robotToken
	}
}

// WithBearerToken configures bearer token authentication
func WithBearerToken(token string) AuthConfigOption {
	return func(c *AuthConfig) {
		c.Type = AuthTypeBearer
		c.Token = token
	}
}

// WithProject sets the Harbor project name
func WithProject(projectName string) AuthConfigOption {
	return func(c *AuthConfig) {
		c.ProjectName = projectName
	}
}

// NewAuthConfig creates a new AuthConfig with options
func NewAuthConfig(registryURL string, opts ...AuthConfigOption) *AuthConfig {
	config := &AuthConfig{
		RegistryURL: registryURL,
		Type:        AuthTypeBasic, // Default
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
