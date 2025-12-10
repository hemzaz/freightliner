package quay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	// AuthTypeRobot uses Quay robot account authentication
	AuthTypeRobot AuthType = "robot"
	// AuthTypeOAuth2 uses OAuth2 token authentication
	AuthTypeOAuth2 AuthType = "oauth2"
)

// AuthConfig contains authentication configuration for Quay
type AuthConfig struct {
	// Type is the authentication type
	Type AuthType

	// Username for basic authentication
	Username string

	// Password for basic authentication
	Password string

	// RobotUsername is the robot account username (namespace+robotname)
	RobotUsername string

	// RobotToken is the robot account token
	RobotToken string

	// OAuth2Token is the OAuth2 access token
	OAuth2Token string

	// Organization is the Quay organization name
	Organization string

	// RegistryURL is the Quay registry URL (default: quay.io)
	RegistryURL string

	// TeamName is the team name for team-based permissions
	TeamName string
}

// TokenResponse represents a Quay OAuth token response
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	IssuedAt  string `json:"issued_at"`
}

// QuayAuthenticator implements authn.Authenticator for Quay.io
type QuayAuthenticator struct {
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

// NewQuayAuthenticator creates a new Quay authenticator
func NewQuayAuthenticator(config *AuthConfig) (*QuayAuthenticator, error) {
	if config == nil {
		return nil, errors.InvalidInputf("auth config is required")
	}

	// Set default registry URL if not provided
	if config.RegistryURL == "" {
		config.RegistryURL = "quay.io"
	}

	// Validate authentication credentials based on type
	switch config.Type {
	case AuthTypeBasic:
		if config.Username == "" || config.Password == "" {
			return nil, errors.InvalidInputf("username and password required for basic auth")
		}
	case AuthTypeRobot:
		if config.RobotUsername == "" || config.RobotToken == "" {
			return nil, errors.InvalidInputf("robot username and token required for robot auth")
		}
		// Set username/password for robot account
		config.Username = config.RobotUsername
		config.Password = config.RobotToken
	case AuthTypeOAuth2:
		if config.OAuth2Token == "" {
			return nil, errors.InvalidInputf("OAuth2 token required for OAuth2 auth")
		}
	default:
		config.Type = AuthTypeBasic // Default to basic auth
	}

	return &QuayAuthenticator{
		config:     config,
		tokenCache: &tokenCache{},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Authorization returns the authorization configuration for Quay
func (a *QuayAuthenticator) Authorization() (*authn.AuthConfig, error) {
	switch a.config.Type {
	case AuthTypeBasic, AuthTypeRobot:
		// For basic and robot auth, return credentials directly
		return &authn.AuthConfig{
			Username: a.config.Username,
			Password: a.config.Password,
		}, nil

	case AuthTypeOAuth2:
		// For OAuth2, use the token as password with special username
		return &authn.AuthConfig{
			Username:      "$oauthtoken",
			Password:      a.config.OAuth2Token,
			IdentityToken: a.config.OAuth2Token,
		}, nil

	default:
		return nil, errors.InvalidInputf("unsupported auth type: %s", a.config.Type)
	}
}

// getAccessToken retrieves an access token, using cache if valid
func (a *QuayAuthenticator) getAccessToken() (string, error) {
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

// fetchToken fetches a new access token from Quay
func (a *QuayAuthenticator) fetchToken() (string, time.Time, error) {
	// Quay token endpoint
	tokenURL := fmt.Sprintf("https://%s/v2/auth", a.config.RegistryURL)

	// Prepare query parameters
	params := fmt.Sprintf("?service=%s", a.config.RegistryURL)
	if a.config.Organization != "" {
		params += fmt.Sprintf("&scope=repository:%s:pull,push", a.config.Organization)
	}

	// Create request
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		tokenURL+params,
		nil,
	)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to create token request")
	}

	// Add authentication header
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

	// Calculate expiration
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	if tokenResp.ExpiresIn == 0 {
		// Default to 1 hour if not specified
		expiresAt = time.Now().Add(1 * time.Hour)
	}

	return tokenResp.Token, expiresAt, nil
}

// RefreshToken forces a token refresh
func (a *QuayAuthenticator) RefreshToken() error {
	a.tokenCache.mu.Lock()
	defer a.tokenCache.mu.Unlock()

	// Clear cached token
	a.tokenCache.token = ""
	a.tokenCache.expiresAt = time.Time{}

	return nil
}

// GetRegistryURL returns the Quay registry URL
func (a *QuayAuthenticator) GetRegistryURL() string {
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
func WithRobotAccount(robotUsername, robotToken string) AuthConfigOption {
	return func(c *AuthConfig) {
		c.Type = AuthTypeRobot
		c.RobotUsername = robotUsername
		c.RobotToken = robotToken
		c.Username = robotUsername
		c.Password = robotToken
	}
}

// WithOAuth2Token configures OAuth2 token authentication
func WithOAuth2Token(token string) AuthConfigOption {
	return func(c *AuthConfig) {
		c.Type = AuthTypeOAuth2
		c.OAuth2Token = token
	}
}

// WithOrganization sets the Quay organization
func WithOrganization(organization string) AuthConfigOption {
	return func(c *AuthConfig) {
		c.Organization = organization
	}
}

// WithTeam sets the team name for team-based permissions
func WithTeam(teamName string) AuthConfigOption {
	return func(c *AuthConfig) {
		c.TeamName = teamName
	}
}

// NewAuthConfig creates a new AuthConfig with options
func NewAuthConfig(registryURL string, opts ...AuthConfigOption) *AuthConfig {
	if registryURL == "" {
		registryURL = "quay.io"
	}

	config := &AuthConfig{
		RegistryURL: registryURL,
		Type:        AuthTypeBasic, // Default
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
