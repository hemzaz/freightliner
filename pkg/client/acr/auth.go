package acr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"freightliner/pkg/helper/errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/go-containerregistry/pkg/authn"
)

// AuthConfig contains authentication configuration for ACR
type AuthConfig struct {
	// TenantID is the Azure AD tenant ID
	TenantID string

	// ClientID is the service principal client ID
	ClientID string

	// ClientSecret is the service principal client secret
	ClientSecret string

	// UseManagedIdentity enables managed identity authentication
	UseManagedIdentity bool

	// SubscriptionID is the Azure subscription ID
	SubscriptionID string

	// ResourceGroup is the Azure resource group
	ResourceGroup string

	// RegistryName is the ACR registry name
	RegistryName string
}

// TokenResponse represents an ACR access token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// ACRAuthenticator implements authn.Authenticator for Azure Container Registry
type ACRAuthenticator struct {
	config      *AuthConfig
	credential  azcore.TokenCredential
	tokenCache  *tokenCache
	registryURL string
	mu          sync.RWMutex
}

// tokenCache stores cached tokens with expiration
type tokenCache struct {
	token        string
	expiresAt    time.Time
	refreshToken string
	mu           sync.RWMutex
}

// NewACRAuthenticator creates a new ACR authenticator
func NewACRAuthenticator(config *AuthConfig) (*ACRAuthenticator, error) {
	if config.RegistryName == "" {
		return nil, errors.InvalidInputf("registry name is required")
	}

	var credential azcore.TokenCredential
	var err error

	// Use managed identity if configured
	if config.UseManagedIdentity {
		credential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create managed identity credential")
		}
	} else if config.ClientID != "" && config.ClientSecret != "" && config.TenantID != "" {
		// Use service principal authentication
		credential, err = azidentity.NewClientSecretCredential(
			config.TenantID,
			config.ClientID,
			config.ClientSecret,
			nil,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create service principal credential")
		}
	} else {
		// Try default Azure credential chain
		credential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create default Azure credential")
		}
	}

	registryURL := fmt.Sprintf("https://%s.azurecr.io", config.RegistryName)

	return &ACRAuthenticator{
		config:      config,
		credential:  credential,
		tokenCache:  &tokenCache{},
		registryURL: registryURL,
	}, nil
}

// Authorization returns the authorization configuration for ACR
func (a *ACRAuthenticator) Authorization() (*authn.AuthConfig, error) {
	token, err := a.getAccessToken()
	if err != nil {
		return nil, err
	}

	return &authn.AuthConfig{
		Username: "00000000-0000-0000-0000-000000000000", // ACR uses a special GUID for username
		Password: token,
	}, nil
}

// getAccessToken retrieves an access token, using cache if valid
func (a *ACRAuthenticator) getAccessToken() (string, error) {
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

	// Get AAD token first
	ctx := context.Background()
	aadToken, err := a.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to get AAD token")
	}

	// Exchange AAD token for ACR access token
	acrToken, expiresAt, err := a.exchangeAADForACRToken(aadToken.Token)
	if err != nil {
		return "", err
	}

	// Cache the token
	a.tokenCache.token = acrToken
	a.tokenCache.expiresAt = expiresAt

	return acrToken, nil
}

// exchangeAADForACRToken exchanges an AAD token for an ACR access token
func (a *ACRAuthenticator) exchangeAADForACRToken(aadToken string) (string, time.Time, error) {
	// ACR token exchange endpoint
	exchangeURL := fmt.Sprintf("%s/oauth2/exchange", a.registryURL)

	// Prepare the request
	data := url.Values{}
	data.Set("grant_type", "access_token")
	data.Set("service", fmt.Sprintf("%s.azurecr.io", a.config.RegistryName))
	data.Set("access_token", aadToken)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		exchangeURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to create token exchange request")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to exchange AAD token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", time.Time{}, errors.InvalidInputf("token exchange failed: %s - %s", resp.Status, string(body))
	}

	// Parse the response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to parse token response")
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return tokenResp.AccessToken, expiresAt, nil
}

// RefreshToken refreshes the cached token
func (a *ACRAuthenticator) RefreshToken() error {
	a.tokenCache.mu.Lock()
	defer a.tokenCache.mu.Unlock()

	// Clear cached token to force refresh
	a.tokenCache.token = ""
	a.tokenCache.expiresAt = time.Time{}

	return nil
}

// GetRegistryURL returns the ACR registry URL
func (a *ACRAuthenticator) GetRegistryURL() string {
	return a.registryURL
}

// NewAuthConfigFromEnvironment creates an AuthConfig from environment variables
// Expected variables: AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, ACR_REGISTRY_NAME
func NewAuthConfigFromEnvironment() (*AuthConfig, error) {
	// This would typically read from environment variables
	// Left as a placeholder for implementation
	return nil, errors.NotImplementedf("environment-based auth config not yet implemented")
}

// AuthConfigOption is a functional option for AuthConfig
type AuthConfigOption func(*AuthConfig)

// WithManagedIdentity enables managed identity authentication
func WithManagedIdentity() AuthConfigOption {
	return func(c *AuthConfig) {
		c.UseManagedIdentity = true
	}
}

// WithServicePrincipal configures service principal authentication
func WithServicePrincipal(tenantID, clientID, clientSecret string) AuthConfigOption {
	return func(c *AuthConfig) {
		c.TenantID = tenantID
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		c.UseManagedIdentity = false
	}
}

// WithSubscription sets the Azure subscription ID
func WithSubscription(subscriptionID string) AuthConfigOption {
	return func(c *AuthConfig) {
		c.SubscriptionID = subscriptionID
	}
}

// NewAuthConfig creates a new AuthConfig with options
func NewAuthConfig(registryName string, opts ...AuthConfigOption) *AuthConfig {
	config := &AuthConfig{
		RegistryName: registryName,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
