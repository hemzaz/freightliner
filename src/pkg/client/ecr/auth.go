package ecr

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

// ECRAuthenticator implements the go-containerregistry authn.Authenticator interface for ECR
type ECRAuthenticator struct {
	client   *ecr.Client
	registry string
	region   string
}

// NewECRAuthenticator creates a new authenticator for ECR
func NewECRAuthenticator(client *ecr.Client, region string) *ECRAuthenticator {
	registryID := "" // Empty string means use the default registry for the credentials
	registry := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", registryID, region)

	return &ECRAuthenticator{
		client:   client,
		registry: registry,
		region:   region,
	}
}

// Authorization returns the authorization for ECR
func (a *ECRAuthenticator) Authorization() (*authn.AuthConfig, error) {
	ctx := context.Background()

	// Get authorization token from ECR
	output, err := a.client.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ECR authorization token: %w", err)
	}

	if len(output.AuthorizationData) == 0 {
		return nil, fmt.Errorf("no ECR authorization data returned")
	}

	// Get the authorization token
	authData := output.AuthorizationData[0]
	authToken := aws.ToString(authData.AuthorizationToken)

	// Decode the authorization token
	decodedToken, err := base64.StdEncoding.DecodeString(authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ECR authorization token: %w", err)
	}

	// The token is in the format "username:password"
	parts := strings.SplitN(string(decodedToken), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ECR authorization token format")
	}

	return &authn.AuthConfig{
		Username: parts[0],
		Password: parts[1],
	}, nil
}

// ECRKeychain implements the go-containerregistry authn.Keychain interface for ECR
type ECRKeychain struct {
	client    *ecr.Client
	region    string
	accountID string
}

// NewECRKeychain creates a new keychain for ECR
func NewECRKeychain(client *ecr.Client, region, accountID string) *ECRKeychain {
	return &ECRKeychain{
		client:    client,
		region:    region,
		accountID: accountID,
	}
}

// Resolve returns an authenticator for the given resource or an error
func (k *ECRKeychain) Resolve(target authn.Resource) (authn.Authenticator, error) {
	registry := target.RegistryStr()

	// Check if this is an ECR registry
	if !isECRRegistry(registry, k.region) {
		return authn.Anonymous, nil
	}

	return NewECRAuthenticator(k.client, k.region), nil
}

// isECRRegistry checks if the registry is an ECR registry in the expected region
func isECRRegistry(registry, region string) bool {
	return strings.HasSuffix(registry, fmt.Sprintf(".dkr.ecr.%s.amazonaws.com", region))
}

// ParseECRRepository parses an ECR repository name into registry and repository components
func ParseECRRepository(repoName string, region string) (name.Registry, name.Repository, error) {
	// Handle repo name with or without the registry prefix
	if !strings.Contains(repoName, ".dkr.ecr.") {
		// Assume it's a bare repository name, construct the full name with registry
		registryName := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", "", region)
		fullRepo := fmt.Sprintf("%s/%s", registryName, repoName)
		repo, err := name.NewRepository(fullRepo)
		if err != nil {
			return name.Registry{}, name.Repository{}, fmt.Errorf("failed to parse ECR repository name: %w", err)
		}

		return repo.Registry, repo, nil
	}

	// If it already has the registry, just parse it
	repo, err := name.NewRepository(repoName)
	if err != nil {
		return name.Registry{}, name.Repository{}, fmt.Errorf("failed to parse ECR repository name: %w", err)
	}

	return repo.Registry, repo, nil
}
