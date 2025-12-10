package ghcr

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
		envVars     map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "anonymous client",
			opts: ClientOptions{
				Token:    "",
				Username: "",
				Logger:   log.NewBasicLogger(log.InfoLevel),
			},
			wantErr: false,
		},
		{
			name: "authenticated client with token",
			opts: ClientOptions{
				Token:    "ghp_test_token",
				Username: "testuser",
				Logger:   log.NewBasicLogger(log.InfoLevel),
			},
			wantErr: false,
		},
		{
			name: "client with token from config",
			opts: ClientOptions{
				RegistryConfig: config.RegistryConfig{
					Auth: config.AuthConfig{
						Type:  config.AuthTypeToken,
						Token: "ghp_config_token",
					},
				},
				Logger: log.NewBasicLogger(log.InfoLevel),
			},
			wantErr: false,
		},
		{
			name: "client reads token from GITHUB_TOKEN env",
			opts: ClientOptions{
				Logger: log.NewBasicLogger(log.InfoLevel),
			},
			envVars: map[string]string{
				"GITHUB_TOKEN": "ghp_env_token",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables if provided
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			client, err := NewClient(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if client == nil {
					t.Error("NewClient() returned nil client")
				}
				if client.GetRegistryName() != GHCRRegistry {
					t.Errorf("GetRegistryName() = %v, want %v", client.GetRegistryName(), GHCRRegistry)
				}
			}
		})
	}
}

func TestNormalizeRepositoryName(t *testing.T) {
	client, _ := NewClient(ClientOptions{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple repository",
			input: "owner/repo",
			want:  "owner/repo",
		},
		{
			name:  "with ghcr.io prefix",
			input: "ghcr.io/owner/repo",
			want:  "owner/repo",
		},
		{
			name:  "uppercase to lowercase",
			input: "OWNER/REPO",
			want:  "owner/repo",
		},
		{
			name:  "with leading slash",
			input: "/owner/repo",
			want:  "owner/repo",
		},
		{
			name:  "organization repository",
			input: "myorg/myteam/myrepo",
			want:  "myorg/myteam/myrepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.normalizeRepositoryName(tt.input)
			if got != tt.want {
				t.Errorf("normalizeRepositoryName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRepository(t *testing.T) {
	client, err := NewClient(ClientOptions{
		Token:    "ghp_test_token",
		Username: "testuser",
		Logger:   log.NewBasicLogger(log.InfoLevel),
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
			repoName: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "valid org repository",
			repoName: "myorg/myteam/myrepo",
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

func TestAuthenticatorFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantAuth bool
	}{
		{
			name: "with GITHUB_TOKEN",
			envVars: map[string]string{
				"GITHUB_TOKEN": "ghp_token",
			},
			wantAuth: true,
		},
		{
			name: "with GH_TOKEN",
			envVars: map[string]string{
				"GH_TOKEN": "ghp_token",
			},
			wantAuth: true,
		},
		{
			name: "with GHCR_TOKEN",
			envVars: map[string]string{
				"GHCR_TOKEN": "ghp_token",
			},
			wantAuth: true,
		},
		{
			name:     "no environment variables",
			envVars:  map[string]string{},
			wantAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all possible env vars first
			os.Unsetenv("GITHUB_TOKEN")
			os.Unsetenv("GH_TOKEN")
			os.Unsetenv("GHCR_TOKEN")

			// Set environment variables if provided
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			auth := NewAuthenticatorFromEnv()
			if auth == nil {
				t.Error("NewAuthenticatorFromEnv() returned nil")
				return
			}

			if auth.IsAuthenticated() != tt.wantAuth {
				t.Errorf("IsAuthenticated() = %v, want %v", auth.IsAuthenticated(), tt.wantAuth)
			}
		})
	}
}
