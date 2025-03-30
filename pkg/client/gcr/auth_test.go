package gcr

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

// Mock for OAuth2 TokenSource
type mockTokenSource struct {
	mock.Mock
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	args := m.Called()
	return args.Get(0).(*oauth2.Token), args.Error(1)
}

// Mock HTTP Client
type mockHTTPClient struct {
	mock.Mock
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// Mock Resource that implements authn.Resource
type resourceMock struct {
	registry string
}

func (r resourceMock) RegistryStr() string {
	return r.registry
}

func (r resourceMock) String() string {
	return r.registry
}

func TestGCRAuthenticatorAuthorization(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mockTokenSource)
		expectedAuth *authn.AuthConfig
		expectedErr  bool
	}{
		{
			name: "Successful authentication",
			mockSetup: func(mockTS *mockTokenSource) {
				mockTS.On("Token").Return(&oauth2.Token{
					AccessToken: "gcp-token-123",
				}, nil)
			},
			expectedAuth: &authn.AuthConfig{
				Username: "oauth2accesstoken",
				Password: "gcp-token-123",
			},
			expectedErr: false,
		},
		{
			name: "Token error",
			mockSetup: func(mockTS *mockTokenSource) {
				mockTS.On("Token").Return(&oauth2.Token{}, errors.New("token error"))
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTS := &mockTokenSource{}
			tc.mockSetup(mockTS)

			auth := &GCRAuthenticator{
				ts: mockTS,
			}

			config, err := auth.Authorization()
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAuth.Username, config.Username)
				assert.Equal(t, tc.expectedAuth.Password, config.Password)
			}

			mockTS.AssertExpectations(t)
		})
	}
}

func TestGCRKeychainResolve(t *testing.T) {
	tests := []struct {
		name         string
		resource     resourceMock
		mockSetup    func(*mockTokenSource)
		expectedAuth *authn.AuthConfig
		expectedErr  bool
	}{
		{
			name: "Successful GCR resolution",
			resource: resourceMock{
				registry: "gcr.io",
			},
			mockSetup: func(mockTS *mockTokenSource) {
				mockTS.On("Token").Return(&oauth2.Token{
					AccessToken: "gcp-token-123",
				}, nil)
			},
			expectedAuth: &authn.AuthConfig{
				Username: "oauth2accesstoken",
				Password: "gcp-token-123",
			},
			expectedErr: false,
		},
		{
			name: "Token error",
			resource: resourceMock{
				registry: "gcr.io",
			},
			mockSetup: func(mockTS *mockTokenSource) {
				mockTS.On("Token").Return(&oauth2.Token{}, errors.New("token error"))
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
		{
			name: "Non-GCR registry",
			resource: resourceMock{
				registry: "docker.io",
			},
			mockSetup: func(mockTS *mockTokenSource) {
				// Should not be called
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTS := &mockTokenSource{}
			tc.mockSetup(mockTS)

			keychain := &GCRKeychain{
				ts: mockTS,
			}

			authenticator, err := keychain.Resolve(tc.resource)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Get the auth config from the authenticator
				config, err := authenticator.Authorization()
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAuth.Username, config.Username)
				assert.Equal(t, tc.expectedAuth.Password, config.Password)
			}

			mockTS.AssertExpectations(t)
		})
	}
}

func TestGCRTransport(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockTokenSource, *mockHTTPClient)
		expectedErr bool
	}{
		{
			name: "Successful transport",
			mockSetup: func(mockTS *mockTokenSource, mockClient *mockHTTPClient) {
				mockTS.On("Token").Return(&oauth2.Token{
					AccessToken: "gcp-token-123",
				}, nil)

				// Create a response for the HTTP client
				resp := &http.Response{
					StatusCode: http.StatusOK,
				}

				mockClient.On("Do", mock.Anything).Return(resp, nil)
			},
			expectedErr: false,
		},
		{
			name: "Token error",
			mockSetup: func(mockTS *mockTokenSource, mockClient *mockHTTPClient) {
				mockTS.On("Token").Return(&oauth2.Token{}, errors.New("token error"))

				// Client's Do method should not be called
			},
			expectedErr: true,
		},
		{
			name: "HTTP error",
			mockSetup: func(mockTS *mockTokenSource, mockClient *mockHTTPClient) {
				mockTS.On("Token").Return(&oauth2.Token{
					AccessToken: "gcp-token-123",
				}, nil)

				mockClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("http error"))
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTS := &mockTokenSource{}
			mockClient := &mockHTTPClient{}
			tc.mockSetup(mockTS, mockClient)

			// Create a GCR transport with the mock token source
			transport := &gcrTransport{
				base: http.DefaultTransport,
				src:  mockTS,
			}

			// Create a test request
			req, _ := http.NewRequest("GET", "https://gcr.io/v2/", nil)

			// Call the RoundTrip method on the transport
			resp, err := transport.RoundTrip(req)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}

			mockTS.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestNewAuthenticator(t *testing.T) {
	// Test with valid credentials file
	auth, err := NewAuthenticator(Options{
		CredentialsFile: "", // Will use default authentication
	})
	assert.NoError(t, err)
	assert.NotNil(t, auth)

	// Test with specific token source
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"})
	auth, err = NewAuthenticator(Options{
		TokenSource: ts,
	})
	assert.NoError(t, err)
	assert.NotNil(t, auth)
}

func TestIsGCPPath(t *testing.T) {
	tests := []struct {
		name       string
		repository string
		expected   bool
	}{
		{
			name:       "Valid GCP path with project/repo",
			repository: "project/repo",
			expected:   true,
		},
		{
			name:       "Valid GCP path with project/path/repo",
			repository: "project/path/repo",
			expected:   true,
		},
		{
			name:       "Invalid path with just repo",
			repository: "repo",
			expected:   false,
		},
		{
			name:       "Empty string",
			repository: "",
			expected:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isGCPPath(tc.repository)
			assert.Equal(t, tc.expected, result)
		})
	}
}
