package generic

import (
	"context"
	"os"
	"testing"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		opts        ClientOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "valid anonymous client",
			opts: ClientOptions{
				RegistryConfig: config.RegistryConfig{
					Endpoint: "https://registry.example.com",
					Auth: config.AuthConfig{
						Type: config.AuthTypeAnonymous,
					},
				},
				RegistryName: "example",
				Logger:       log.NewBasicLogger(log.InfoLevel),
			},
			wantErr: false,
		},
		{
			name: "valid basic auth client",
			opts: ClientOptions{
				RegistryConfig: config.RegistryConfig{
					Endpoint: "https://registry.example.com",
					Auth: config.AuthConfig{
						Type:     config.AuthTypeBasic,
						Username: "testuser",
						Password: "testpass",
					},
				},
				RegistryName: "example",
				Logger:       log.NewBasicLogger(log.InfoLevel),
			},
			wantErr: false,
		},
		{
			name: "valid token auth client",
			opts: ClientOptions{
				RegistryConfig: config.RegistryConfig{
					Endpoint: "https://registry.example.com",
					Auth: config.AuthConfig{
						Type:  config.AuthTypeToken,
						Token: "test-token",
					},
				},
				RegistryName: "example",
				Logger:       log.NewBasicLogger(log.InfoLevel),
			},
			wantErr: false,
		},
		{
			name: "insecure registry",
			opts: ClientOptions{
				RegistryConfig: config.RegistryConfig{
					Endpoint: "http://insecure-registry.local",
					Insecure: true,
					Auth: config.AuthConfig{
						Type: config.AuthTypeAnonymous,
					},
				},
				RegistryName: "insecure",
				Logger:       log.NewBasicLogger(log.InfoLevel),
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			opts: ClientOptions{
				RegistryConfig: config.RegistryConfig{
					Endpoint: "",
				},
				Logger: log.NewBasicLogger(log.InfoLevel),
			},
			wantErr:     true,
			errContains: "endpoint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestNormalizeRegistryURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "with https prefix",
			input: "https://registry.example.com",
			want:  "registry.example.com",
		},
		{
			name:  "with http prefix",
			input: "http://registry.example.com",
			want:  "registry.example.com",
		},
		{
			name:  "with trailing slash",
			input: "registry.example.com/",
			want:  "registry.example.com",
		},
		{
			name:  "no prefix or suffix",
			input: "registry.example.com",
			want:  "registry.example.com",
		},
		{
			name:  "with port",
			input: "https://registry.example.com:5000",
			want:  "registry.example.com:5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRegistryURL(tt.input)
			if got != tt.want {
				t.Errorf("normalizeRegistryURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandEnvVars(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_VAR", "testvalue")
	os.Setenv("REGISTRY_USER", "admin")
	defer os.Unsetenv("TEST_VAR")
	defer os.Unsetenv("REGISTRY_USER")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no env vars",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "single env var",
			input: "${TEST_VAR}",
			want:  "testvalue",
		},
		{
			name:  "env var in string",
			input: "user=${REGISTRY_USER}",
			want:  "user=admin",
		},
		{
			name:  "multiple env vars",
			input: "${REGISTRY_USER}:${TEST_VAR}",
			want:  "admin:testvalue",
		},
		{
			name:  "undefined env var",
			input: "${UNDEFINED_VAR}",
			want:  "",
		},
		{
			name:  "malformed env var",
			input: "${INCOMPLETE",
			want:  "${INCOMPLETE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandEnvVars(tt.input)
			if got != tt.want {
				t.Errorf("expandEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRepository(t *testing.T) {
	client, err := NewClient(ClientOptions{
		RegistryConfig: config.RegistryConfig{
			Endpoint: "https://registry.example.com",
			Auth: config.AuthConfig{
				Type: config.AuthTypeAnonymous,
			},
		},
		RegistryName: "example",
		Logger:       log.NewBasicLogger(log.InfoLevel),
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name     string
		repoName string
		wantErr  bool
	}{
		{
			name:     "valid repository",
			repoName: "myrepo",
			wantErr:  false,
		},
		{
			name:     "valid nested repository",
			repoName: "team/myrepo",
			wantErr:  false,
		},
		{
			name:     "empty repository name",
			repoName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := client.GetRepository(context.Background(), tt.repoName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && repo == nil {
				t.Error("GetRepository() returned nil repository")
			}
		})
	}
}

func TestCreateAuthenticator(t *testing.T) {
	tests := []struct {
		name    string
		config  config.RegistryConfig
		wantErr bool
	}{
		{
			name: "anonymous auth",
			config: config.RegistryConfig{
				Auth: config.AuthConfig{
					Type: config.AuthTypeAnonymous,
				},
			},
			wantErr: false,
		},
		{
			name: "basic auth",
			config: config.RegistryConfig{
				Auth: config.AuthConfig{
					Type:     config.AuthTypeBasic,
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: false,
		},
		{
			name: "token auth",
			config: config.RegistryConfig{
				Auth: config.AuthConfig{
					Type:  config.AuthTypeToken,
					Token: "token123",
				},
			},
			wantErr: false,
		},
		{
			name: "basic auth missing password",
			config: config.RegistryConfig{
				Auth: config.AuthConfig{
					Type:     config.AuthTypeBasic,
					Username: "user",
				},
			},
			wantErr: true,
		},
		{
			name: "token auth missing token",
			config: config.RegistryConfig{
				Auth: config.AuthConfig{
					Type: config.AuthTypeToken,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := createAuthenticator(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("createAuthenticator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && auth == nil {
				t.Error("createAuthenticator() returned nil authenticator")
			}
		})
	}
}
