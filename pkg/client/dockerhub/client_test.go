package dockerhub

import (
	"context"
	"testing"

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
			name: "anonymous client",
			opts: ClientOptions{
				Username: "",
				Password: "",
				Logger:   log.NewBasicLogger(log.InfoLevel),
			},
			wantErr: false,
		},
		{
			name: "authenticated client",
			opts: ClientOptions{
				Username: "testuser",
				Password: "testpass",
				Logger:   log.NewBasicLogger(log.InfoLevel),
			},
			wantErr: false,
		},
		{
			name: "client with custom retry config",
			opts: ClientOptions{
				Username: "testuser",
				Password: "testpass",
				Logger:   log.NewBasicLogger(log.InfoLevel),
				RetryConfig: &RetryConfig{
					MaxRetries:               5,
					EnableExponentialBackoff: true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if client == nil {
					t.Error("NewClient() returned nil client")
				}
				if client.GetRegistryName() != DockerHubRegistry {
					t.Errorf("GetRegistryName() = %v, want %v", client.GetRegistryName(), DockerHubRegistry)
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
			name:  "official image",
			input: "nginx",
			want:  "library/nginx",
		},
		{
			name:  "user repository",
			input: "myuser/myrepo",
			want:  "myuser/myrepo",
		},
		{
			name:  "with docker.io prefix",
			input: "docker.io/myuser/myrepo",
			want:  "myuser/myrepo",
		},
		{
			name:  "with registry-1.docker.io prefix",
			input: "registry-1.docker.io/library/nginx",
			want:  "library/nginx",
		},
		{
			name:  "with leading slash",
			input: "/nginx",
			want:  "library/nginx",
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

func TestShouldRetry(t *testing.T) {
	client, _ := NewClient(ClientOptions{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})

	tests := []struct {
		name   string
		errMsg string
		want   bool
	}{
		{
			name:   "rate limit error",
			errMsg: "429 Too Many Requests",
			want:   true,
		},
		{
			name:   "timeout error",
			errMsg: "context deadline exceeded timeout",
			want:   true,
		},
		{
			name:   "server error",
			errMsg: "500 Internal Server Error",
			want:   true,
		},
		{
			name:   "not found error",
			errMsg: "404 Not Found",
			want:   false,
		},
		{
			name:   "unauthorized error",
			errMsg: "401 Unauthorized",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &testError{msg: tt.errMsg}
			got := client.shouldRetry(err)
			if got != tt.want {
				t.Errorf("shouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRepository(t *testing.T) {
	client, err := NewClient(ClientOptions{
		Logger: log.NewBasicLogger(log.InfoLevel),
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
			name:     "valid official image",
			repoName: "nginx",
			wantErr:  false,
		},
		{
			name:     "valid user repository",
			repoName: "myuser/myrepo",
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

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
