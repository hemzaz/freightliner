package ecr

import (
	"os"
	"strings"
	"testing"

	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
)

// Note: Mock types have been moved to pkg/testing/mocks for reuse
// Use mocks.MockECRClient and mocks.MockSTSClient for testing

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		region      string
		accountID   string
		registry    string
		expectedErr bool
	}{
		{
			name:        "With explicit account ID",
			region:      "us-west-2",
			accountID:   "123456789012",
			registry:    "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			expectedErr: false,
		},
		{
			name:        "Auto-detect account ID requires AWS credentials",
			region:      "us-west-2",
			accountID:   "",
			registry:    "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			expectedErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Skip the test unless explicitly enabled via environment variable
			// These tests require real AWS credentials
			if os.Getenv("ENABLE_ECR_INTEGRATION_TESTS") != "true" {
				t.Skip("Skipping ECR integration test. Set ENABLE_ECR_INTEGRATION_TESTS=true to run.")
			}

			client, err := NewClient(ClientOptions{
				Region:    tc.region,
				AccountID: tc.accountID,
				Logger:    log.NewLogger(),
			})

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if tc.registry != "" {
					assert.Equal(t, tc.registry, client.GetRegistryName())
				}
				assert.Equal(t, tc.region, client.region)
			}
		})
	}
}

// parseECRRepository parses ECR repository references for testing
func parseECRRepository(input string) (string, string) {
	// Split by '/' to separate registry from repository
	parts := strings.Split(input, "/")

	// If no '/' or only one part, it's just a repository name
	if len(parts) == 1 {
		return "", input
	}

	// Check if first part looks like an ECR registry
	if strings.Contains(parts[0], "dkr.ecr") && strings.Contains(parts[0], ".amazonaws.com") {
		// It's a full ECR URI
		registry := parts[0]
		repository := strings.Join(parts[1:], "/")
		return registry, repository
	}

	// Otherwise it's just a repository path
	return "", input
}

func TestParseECRRepository(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedRegistry   string
		expectedRepository string
	}{
		{
			name:               "Full ECR URI",
			input:              "123456789012.dkr.ecr.us-west-2.amazonaws.com/repo-name",
			expectedRegistry:   "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			expectedRepository: "repo-name",
		},
		{
			name:               "Simple repository name",
			input:              "repo-name",
			expectedRegistry:   "",
			expectedRepository: "repo-name",
		},
		{
			name:               "Repository with path",
			input:              "org/repo-name",
			expectedRegistry:   "",
			expectedRepository: "org/repo-name",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			registry, repository := parseECRRepository(tc.input)
			assert.Equal(t, tc.expectedRegistry, registry)
			assert.Equal(t, tc.expectedRepository, repository)
		})
	}
}
