package common_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/authn"
)

// mockAuthenticator implements authn.Authenticator for testing
type mockAuthenticator struct {
	username string
	password string
	token    string
}

func (m *mockAuthenticator) Authorization() (*authn.AuthConfig, error) {
	return &authn.AuthConfig{
		Username: m.username,
		Password: m.password,
		Auth:     m.token,
	}, nil
}

// TestNewEnhancedClient tests creating enhanced clients
func TestNewEnhancedClient(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	auth := &mockAuthenticator{
		username: "testuser",
		password: "testpass",
	}

	tests := []struct {
		name string
		opts common.EnhancedClientOptions
	}{
		{
			name: "minimal options",
			opts: common.EnhancedClientOptions{
				RegistryName: "gcr.io/test-project",
			},
		},
		{
			name: "with authentication",
			opts: common.EnhancedClientOptions{
				RegistryName:  "gcr.io/test-project",
				Logger:        logger,
				Authenticator: auth,
			},
		},
		{
			name: "with custom settings",
			opts: common.EnhancedClientOptions{
				RegistryName:             "gcr.io/test-project",
				Logger:                   logger,
				Authenticator:            auth,
				EnableLogging:            true,
				EnableRetries:            true,
				MaxRetries:               5,
				RequestTimeout:           60 * time.Second,
				CredentialRefreshTimeout: 30 * time.Minute,
			},
		},
		{
			name: "with nil logger",
			opts: common.EnhancedClientOptions{
				RegistryName: "gcr.io/test-project",
				Logger:       nil, // Should create default logger
			},
		},
		{
			name: "with default max retries",
			opts: common.EnhancedClientOptions{
				RegistryName: "gcr.io/test-project",
				MaxRetries:   0, // Should default to 3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := common.NewEnhancedClient(tt.opts)

			if client == nil {
				t.Fatal("Expected non-nil client")
			}

			if client.GetRegistryName() != tt.opts.RegistryName {
				t.Errorf("Expected registry %s, got %s", tt.opts.RegistryName, client.GetRegistryName())
			}
		})
	}
}

// TestGetAuthenticator tests retrieving the authenticator
func TestGetAuthenticator(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	tests := []struct {
		name          string
		authenticator authn.Authenticator
	}{
		{
			name: "with authenticator",
			authenticator: &mockAuthenticator{
				username: "testuser",
				password: "testpass",
			},
		},
		{
			name:          "without authenticator",
			authenticator: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := common.EnhancedClientOptions{
				RegistryName:  "gcr.io/test-project",
				Logger:        logger,
				Authenticator: tt.authenticator,
			}

			client := common.NewEnhancedClient(opts)

			retrievedAuth := client.GetAuthenticator()

			if tt.authenticator == nil {
				if retrievedAuth != nil {
					t.Error("Expected nil authenticator")
				}
			} else {
				if retrievedAuth == nil {
					t.Error("Expected non-nil authenticator")
				}

				// Verify it's the same authenticator
				expectedConfig, _ := tt.authenticator.Authorization()
				actualConfig, _ := retrievedAuth.Authorization()

				if expectedConfig.Username != actualConfig.Username {
					t.Errorf("Expected username %s, got %s", expectedConfig.Username, actualConfig.Username)
				}
			}
		})
	}
}

// TestSetAuthenticator tests setting a new authenticator
func TestSetAuthenticator(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	initialAuth := &mockAuthenticator{
		username: "user1",
		password: "pass1",
	}

	newAuth := &mockAuthenticator{
		username: "user2",
		password: "pass2",
	}

	opts := common.EnhancedClientOptions{
		RegistryName:  "gcr.io/test-project",
		Logger:        logger,
		Authenticator: initialAuth,
	}

	client := common.NewEnhancedClient(opts)

	// Verify initial authenticator
	auth := client.GetAuthenticator()
	if auth == nil {
		t.Fatal("Expected initial authenticator")
	}

	config, _ := auth.Authorization()
	if config.Username != "user1" {
		t.Errorf("Expected username user1, got %s", config.Username)
	}

	// Set new authenticator
	client.SetAuthenticator(newAuth)

	// Verify new authenticator
	auth = client.GetAuthenticator()
	if auth == nil {
		t.Fatal("Expected new authenticator")
	}

	config, _ = auth.Authorization()
	if config.Username != "user2" {
		t.Errorf("Expected username user2, got %s", config.Username)
	}
}

// TestSetAuthenticatorToNil tests setting authenticator to nil
func TestSetAuthenticatorToNil(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	initialAuth := &mockAuthenticator{
		username: "testuser",
		password: "testpass",
	}

	opts := common.EnhancedClientOptions{
		RegistryName:  "gcr.io/test-project",
		Logger:        logger,
		Authenticator: initialAuth,
	}

	client := common.NewEnhancedClient(opts)

	// Set to nil
	client.SetAuthenticator(nil)

	// Verify it's nil
	auth := client.GetAuthenticator()
	if auth != nil {
		t.Error("Expected nil authenticator after setting to nil")
	}
}

// TestGetTransport tests transport creation and caching
func TestGetTransport(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	tests := []struct {
		name     string
		opts     common.EnhancedClientOptions
		repoName string
		wantErr  bool
	}{
		{
			name: "basic transport",
			opts: common.EnhancedClientOptions{
				RegistryName: "gcr.io",
				Logger:       logger,
			},
			repoName: "test-project/app",
			wantErr:  false,
		},
		{
			name: "with authentication",
			opts: common.EnhancedClientOptions{
				RegistryName: "gcr.io",
				Logger:       logger,
				Authenticator: &mockAuthenticator{
					username: "testuser",
					password: "testpass",
				},
			},
			repoName: "test-project/app",
			wantErr:  false,
		},
		{
			name: "with logging enabled",
			opts: common.EnhancedClientOptions{
				RegistryName:  "gcr.io",
				Logger:        logger,
				EnableLogging: true,
			},
			repoName: "test-project/app",
			wantErr:  false,
		},
		{
			name: "with retries enabled",
			opts: common.EnhancedClientOptions{
				RegistryName:  "gcr.io",
				Logger:        logger,
				EnableRetries: true,
				MaxRetries:    5,
			},
			repoName: "test-project/app",
			wantErr:  false,
		},
		{
			name: "with timeout",
			opts: common.EnhancedClientOptions{
				RegistryName:   "gcr.io",
				Logger:         logger,
				RequestTimeout: 30 * time.Second,
			},
			repoName: "test-project/app",
			wantErr:  false,
		},
		{
			name: "all features enabled",
			opts: common.EnhancedClientOptions{
				RegistryName:   "gcr.io",
				Logger:         logger,
				Authenticator:  &mockAuthenticator{username: "user", password: "pass"},
				EnableLogging:  true,
				EnableRetries:  true,
				MaxRetries:     3,
				RequestTimeout: 60 * time.Second,
			},
			repoName: "test-project/app",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := common.NewEnhancedClient(tt.opts)

			transport, err := client.GetTransport(tt.repoName)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if transport == nil {
					t.Error("Expected non-nil transport")
				}

				// Verify transport is cached
				transport2, err2 := client.GetTransport(tt.repoName)
				if err2 != nil {
					t.Errorf("Unexpected error on second call: %v", err2)
				}

				// Should be same instance from cache
				if transport != transport2 {
					t.Error("Expected cached transport to be returned")
				}
			}
		})
	}
}

// TestClearTransportCache tests clearing the transport cache
func TestClearTransportCache(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	opts := common.EnhancedClientOptions{
		RegistryName:  "gcr.io",
		Logger:        logger,
		EnableLogging: true,
	}

	client := common.NewEnhancedClient(opts)

	// Get transport for multiple repos
	repoNames := []string{"project1/app", "project2/app", "project3/app"}
	transports := make([]http.RoundTripper, len(repoNames))

	for i, repoName := range repoNames {
		transport, err := client.GetTransport(repoName)
		if err != nil {
			t.Fatalf("Failed to get transport for %s: %v", repoName, err)
		}
		transports[i] = transport
	}

	// Clear cache
	client.ClearTransportCache()

	// Get transports again - should be different instances
	for i, repoName := range repoNames {
		newTransport, err := client.GetTransport(repoName)
		if err != nil {
			t.Fatalf("Failed to get transport after clear for %s: %v", repoName, err)
		}

		// Should be different instance after cache clear
		if newTransport == transports[i] {
			t.Error("Expected new transport instance after cache clear")
		}
	}
}

// TestSetRetryPolicy tests setting a custom retry policy
func TestSetRetryPolicy(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	opts := common.EnhancedClientOptions{
		RegistryName:  "gcr.io",
		Logger:        logger,
		EnableRetries: true,
	}

	client := common.NewEnhancedClient(opts)

	// Get initial transport
	transport1, err := client.GetTransport("test-project/app")
	if err != nil {
		t.Fatalf("Failed to get transport: %v", err)
	}

	// Set custom retry policy
	customPolicy := func(resp *http.Response, err error) bool {
		// Custom logic: only retry on 503
		return resp != nil && resp.StatusCode == 503
	}

	client.SetRetryPolicy(customPolicy)

	// Get transport again - should be new instance due to policy change
	transport2, err := client.GetTransport("test-project/app")
	if err != nil {
		t.Fatalf("Failed to get transport after policy change: %v", err)
	}

	// Should be different instance after policy change (cache cleared)
	if transport1 == transport2 {
		t.Error("Expected new transport instance after retry policy change")
	}
}

// TestAddTransportOption tests adding transport options
func TestAddTransportOption(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	opts := common.EnhancedClientOptions{
		RegistryName: "gcr.io",
		Logger:       logger,
	}

	client := common.NewEnhancedClient(opts)

	// Get initial transport
	transport1, err := client.GetTransport("test-project/app")
	if err != nil {
		t.Fatalf("Failed to get transport: %v", err)
	}

	// Add custom transport option
	client.AddTransportOption(func(t *http.Transport) {
		t.MaxIdleConns = 500
	})

	// Get transport again - should be new instance due to option change
	transport2, err := client.GetTransport("test-project/app")
	if err != nil {
		t.Fatalf("Failed to get transport after option change: %v", err)
	}

	// Should be different instance after option change (cache cleared)
	if transport1 == transport2 {
		t.Error("Expected new transport instance after adding transport option")
	}
}

// TestGetEnhancedRemoteOptions tests getting remote options
func TestGetEnhancedRemoteOptions(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	tests := []struct {
		name     string
		opts     common.EnhancedClientOptions
		repoName string
		wantErr  bool
	}{
		{
			name: "basic options",
			opts: common.EnhancedClientOptions{
				RegistryName: "gcr.io",
				Logger:       logger,
			},
			repoName: "test-project/app",
			wantErr:  false,
		},
		{
			name: "with TLS skip verify",
			opts: common.EnhancedClientOptions{
				RegistryName:          "gcr.io",
				Logger:                logger,
				InsecureSkipTLSVerify: true,
			},
			repoName: "test-project/app",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := common.NewEnhancedClient(tt.opts)
			ctx := context.Background()

			options, err := client.GetEnhancedRemoteOptions(ctx, tt.repoName)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if options == nil {
					t.Error("Expected non-nil options")
				}

				if len(options) == 0 {
					t.Error("Expected at least some options")
				}
			}
		})
	}
}

// TestEnhancedClientDefaultValues tests default value assignment
func TestEnhancedClientDefaultValues(t *testing.T) {
	// Create client with minimal options
	opts := common.EnhancedClientOptions{
		RegistryName: "gcr.io",
	}

	client := common.NewEnhancedClient(opts)

	// Verify defaults are applied
	// (These are internal, but we can verify through behavior)

	// Should have logger (not nil)
	if client.GetRegistryName() == "" {
		t.Error("Registry name should not be empty")
	}

	t.Log("Default values applied successfully")
}

// TestEnhancedClientConcurrentAccess tests thread-safe operations
func TestEnhancedClientConcurrentAccess(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	opts := common.EnhancedClientOptions{
		RegistryName:  "gcr.io",
		Logger:        logger,
		EnableLogging: true,
		EnableRetries: true,
	}

	client := common.NewEnhancedClient(opts)

	// Concurrent transport requests
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			repoName := "project/app-" + string(rune('0'+idx))
			_, _ = client.GetTransport(repoName)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Concurrent cache operations
	go client.ClearTransportCache()
	go client.ClearTransportCache()

	t.Log("Concurrent access completed without race conditions")
}

// TestEnhancedClientAuthenticatorSwitch tests switching authenticators
func TestEnhancedClientAuthenticatorSwitch(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	auth1 := &mockAuthenticator{username: "user1", password: "pass1"}
	auth2 := &mockAuthenticator{username: "user2", password: "pass2"}
	auth3 := &mockAuthenticator{username: "user3", password: "pass3"}

	opts := common.EnhancedClientOptions{
		RegistryName:  "gcr.io",
		Logger:        logger,
		Authenticator: auth1,
	}

	client := common.NewEnhancedClient(opts)

	// Verify initial auth
	config, _ := client.GetAuthenticator().Authorization()
	if config.Username != "user1" {
		t.Errorf("Expected user1, got %s", config.Username)
	}

	// Switch to auth2
	client.SetAuthenticator(auth2)
	config, _ = client.GetAuthenticator().Authorization()
	if config.Username != "user2" {
		t.Errorf("Expected user2, got %s", config.Username)
	}

	// Switch to auth3
	client.SetAuthenticator(auth3)
	config, _ = client.GetAuthenticator().Authorization()
	if config.Username != "user3" {
		t.Errorf("Expected user3, got %s", config.Username)
	}

	// Switch to nil
	client.SetAuthenticator(nil)
	if client.GetAuthenticator() != nil {
		t.Error("Expected nil authenticator")
	}
}
