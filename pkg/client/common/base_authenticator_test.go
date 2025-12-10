package common

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseAuthenticator(t *testing.T) {
	auth := NewBaseAuthenticator()
	assert.NotNil(t, auth)
	assert.Nil(t, auth.cachedAuth)
	assert.Equal(t, int64(0), auth.cachedExpiry)
}

func TestBaseAuthenticator_Authorization(t *testing.T) {
	auth := NewBaseAuthenticator()

	// Base implementation should return nil, nil
	authConfig, err := auth.Authorization()
	assert.NoError(t, err)
	assert.Nil(t, authConfig)
}

func TestBaseAuthenticator_SetCachedAuth(t *testing.T) {
	auth := NewBaseAuthenticator()
	mockAuth := &authn.Basic{Username: "user", Password: "pass"}
	expiry := time.Now().Add(1 * time.Hour).Unix()

	auth.SetCachedAuth(mockAuth, expiry)

	assert.Equal(t, mockAuth, auth.cachedAuth)
	assert.Equal(t, expiry, auth.cachedExpiry)
}

func TestBaseAuthenticator_GetCachedAuth(t *testing.T) {
	tests := []struct {
		name          string
		setupAuth     authn.Authenticator
		setupExpiry   int64
		currentTime   int64
		expectedValid bool
	}{
		{
			name:          "No cached auth",
			setupAuth:     nil,
			setupExpiry:   0,
			currentTime:   time.Now().Unix(),
			expectedValid: false,
		},
		{
			name:          "Valid cached auth",
			setupAuth:     &authn.Basic{Username: "user", Password: "pass"},
			setupExpiry:   time.Now().Add(1 * time.Hour).Unix(),
			currentTime:   time.Now().Unix(),
			expectedValid: true,
		},
		{
			name:          "Expired cached auth",
			setupAuth:     &authn.Basic{Username: "user", Password: "pass"},
			setupExpiry:   time.Now().Add(-1 * time.Hour).Unix(),
			currentTime:   time.Now().Unix(),
			expectedValid: false,
		},
		{
			name:          "No expiry set (valid)",
			setupAuth:     &authn.Basic{Username: "user", Password: "pass"},
			setupExpiry:   0,
			currentTime:   time.Now().Unix(),
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewBaseAuthenticator()
			if tt.setupAuth != nil {
				auth.SetCachedAuth(tt.setupAuth, tt.setupExpiry)
			}

			cachedAuth, valid := auth.GetCachedAuth(tt.currentTime)
			assert.Equal(t, tt.expectedValid, valid)
			if tt.expectedValid {
				assert.NotNil(t, cachedAuth)
			}
		})
	}
}

func TestBaseAuthenticator_ClearCachedAuth(t *testing.T) {
	auth := NewBaseAuthenticator()
	mockAuth := &authn.Basic{Username: "user", Password: "pass"}
	auth.SetCachedAuth(mockAuth, time.Now().Add(1*time.Hour).Unix())

	// Verify it's set
	assert.NotNil(t, auth.cachedAuth)
	assert.NotEqual(t, int64(0), auth.cachedExpiry)

	// Clear it
	auth.ClearCachedAuth()

	// Verify it's cleared
	assert.Nil(t, auth.cachedAuth)
	assert.Equal(t, int64(0), auth.cachedExpiry)
}

func TestTransportWithAuth(t *testing.T) {
	mockAuth := &authn.Basic{Username: "testuser", Password: "testpass"}
	mockRegistry, err := name.NewRegistry("example.com")
	require.NoError(t, err)

	transport := TransportWithAuth(nil, mockAuth, mockRegistry)
	assert.NotNil(t, transport)

	// Test with custom base transport
	customTransport := &http.Transport{}
	transport = TransportWithAuth(customTransport, mockAuth, mockRegistry)
	assert.NotNil(t, transport)

	// Verify it's an authnTransport
	authTransport, ok := transport.(*authnTransport)
	assert.True(t, ok)
	assert.Equal(t, mockAuth, authTransport.auth)
}

func TestAuthnTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		authConfig  *authn.AuthConfig
		shouldError bool
	}{
		{
			name: "Basic auth with username and password",
			authConfig: &authn.AuthConfig{
				Username: "user",
				Password: "pass",
			},
			shouldError: false,
		},
		{
			name: "Auth with encoded credentials",
			authConfig: &authn.AuthConfig{
				Auth: "dXNlcjpwYXNz", // base64 encoded "user:pass"
			},
			shouldError: false,
		},
		{
			name: "Auth with identity token",
			authConfig: &authn.AuthConfig{
				IdentityToken: "identity-token-12345",
			},
			shouldError: false,
		},
		{
			name: "Auth with registry token",
			authConfig: &authn.AuthConfig{
				RegistryToken: "registry-token-12345",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock authenticator
			mockAuth := &mockAuthenticator{authConfig: tt.authConfig}

			// Create a mock base transport
			mockBaseTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: 200,
					Header:     http.Header{},
				},
			}

			// Create auth transport
			transport := &authnTransport{
				inner:    mockBaseTransport,
				auth:     mockAuth,
				resource: nil,
			}

			// Create a test request
			req, err := http.NewRequest("GET", "https://example.com/v2/", nil)
			require.NoError(t, err)

			// Execute round trip
			resp, err := transport.RoundTrip(req)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, 200, resp.StatusCode)

				// Verify that appropriate auth header was added
				if tt.authConfig.Username != "" && tt.authConfig.Password != "" {
					// Basic auth should be set
					username, password, ok := req.BasicAuth()
					assert.True(t, ok)
					assert.Equal(t, tt.authConfig.Username, username)
					assert.Equal(t, tt.authConfig.Password, password)
				} else if tt.authConfig.Auth != "" {
					assert.Contains(t, req.Header.Get("Authorization"), "Basic")
				} else if tt.authConfig.IdentityToken != "" {
					assert.Contains(t, req.Header.Get("Authorization"), "Bearer "+tt.authConfig.IdentityToken)
				} else if tt.authConfig.RegistryToken != "" {
					assert.Contains(t, req.Header.Get("Authorization"), "Bearer "+tt.authConfig.RegistryToken)
				}
			}
		})
	}
}

// Mock types for testing

type mockAuthenticator struct {
	authConfig *authn.AuthConfig
	err        error
}

func (m *mockAuthenticator) Authorization() (*authn.AuthConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.authConfig, nil
}

type mockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}
