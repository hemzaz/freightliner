package ecr

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"freightliner/pkg/helper/errors"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

// ECRAPI defines the interface for ECR API operations needed by this package
type ECRAPI interface {
	GetAuthorizationToken(ctx context.Context, params *ecr.GetAuthorizationTokenInput, optFns ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error)
}

// ECRAuthenticator implements the go-containerregistry authn.Authenticator interface for ECR
type ECRAuthenticator struct {
	client   ECRAPI
	registry string
	region   string
}

// NewECRAuthenticator creates a new authenticator for ECR
func NewECRAuthenticator(client ECRAPI, region string) *ECRAuthenticator {
	registryID := "" // Empty string means use the default registry for the credentials
	registry := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", registryID, region)

	return &ECRAuthenticator{
		client:   client,
		registry: registry,
		region:   region,
	}
}

// Authorization returns the authorization header for ECR
func (a *ECRAuthenticator) Authorization() (*authn.AuthConfig, error) {
	// Get authorization token from ECR
	input := &ecr.GetAuthorizationTokenInput{}
	resp, err := a.client.GetAuthorizationToken(context.Background(), input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ECR authorization token")
	}

	if len(resp.AuthorizationData) == 0 {
		return nil, errors.InvalidInputf("no authorization data returned from ECR")
	}

	// The token is base64 encoded and in the format "username:password"
	authData := resp.AuthorizationData[0]
	token, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode ECR authorization token")
	}

	parts := strings.SplitN(string(token), ":", 2)
	if len(parts) != 2 {
		return nil, errors.InvalidInputf("invalid ECR authorization token format")
	}

	return &authn.AuthConfig{
		Username: parts[0],
		Password: parts[1],
	}, nil
}

// NewECRClientForRegion creates a new ECR client for the given region
func NewECRClientForRegion(region string) (ECRAPI, error) {
	// Create AWS SDK config for the target region
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load AWS config for region %s", region)
	}

	// Create ECR client using the region-specific config
	return ecr.NewFromConfig(cfg), nil
}

// RegistryAuthenticator creates an authenticator for a specific registry
func (a *ECRAuthenticator) RegistryAuthenticator(registry string) (authn.Authenticator, error) {
	// Check if the registry is an ECR registry
	if !strings.Contains(registry, ".dkr.ecr.") || !strings.Contains(registry, ".amazonaws.com") {
		return nil, errors.InvalidInputf("not an ECR registry: %s", registry)
	}

	// Extract the region from the registry
	regionStart := strings.Index(registry, ".dkr.ecr.") + 9
	regionEnd := strings.Index(registry[regionStart:], ".") + regionStart
	if regionStart == -1 || regionEnd <= regionStart {
		return nil, errors.InvalidInputf("invalid ECR registry format: %s", registry)
	}

	region := registry[regionStart:regionEnd]
	if region != a.region {
		// Create a new AWS SDK client for the target region
		ecrClient, err := NewECRClientForRegion(region)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create ECR client for region %s", region)
		}

		// Create a new authenticator for the cross-region registry
		crossRegionAuth := NewECRAuthenticator(ecrClient, region)
		return crossRegionAuth, nil
	}

	// For the same region, we can use this authenticator
	return a, nil
}

// GetECRRegistry returns the ECR registry URL for the given account and region
func GetECRRegistry(accountID, region string) string {
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", accountID, region)
}

// GetECRRepository returns a Repository object for an ECR repository
// The repo parameter is the repository name without the registry prefix
func (a *ECRAuthenticator) GetECRRepository(repo string) (name.Repository, error) {
	registry := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", "", a.region)
	repoName := fmt.Sprintf("%s/%s", registry, repo)

	repository, err := name.NewRepository(repoName)
	if err != nil {
		return name.Repository{}, errors.Wrap(err, "failed to create ECR repository reference")
	}

	return repository, nil
}

// isECRRegistry returns true if the given registry is an ECR registry
func isECRRegistry(registry string) bool {
	if registry == "" {
		return false
	}

	// Check for standard ECR registry format
	if strings.Contains(registry, ".dkr.ecr.") && strings.Contains(registry, ".amazonaws.com") {
		return true
	}

	// Check for public ECR registry format
	if registry == "public.ecr.aws" {
		return true
	}

	return false
}

// CredentialHelper interface for ECR credential helpers
type CredentialHelper interface {
	Get(serverURL string) (ClientAuth, error)
}

// ClientAuth holds authentication credentials for clients
type ClientAuth struct {
	Username string
	Password string
}

// ECRKeychain implements the go-containerregistry authn.Keychain interface for ECR
type ECRKeychain struct {
	helper CredentialHelper
}

// Resolve returns an authenticator for the given resource
func (k *ECRKeychain) Resolve(resource authn.Resource) (authn.Authenticator, error) {
	registry := resource.RegistryStr()

	// Check if this is an ECR registry
	if !isECRRegistry(registry) {
		return nil, errors.InvalidInputf("not an ECR registry: %s", registry)
	}

	// Get credentials from the helper
	serverURL := fmt.Sprintf("https://%s", registry)
	auth, err := k.helper.Get(serverURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ECR credentials")
	}

	// Create an authenticator with the credentials
	return &staticAuthenticator{
		auth: &authn.AuthConfig{
			Username: auth.Username,
			Password: auth.Password,
		},
	}, nil
}

// staticAuthenticator is a simple authn.Authenticator that returns static credentials
type staticAuthenticator struct {
	auth *authn.AuthConfig
}

// Authorization returns the static auth config
func (a *staticAuthenticator) Authorization() (*authn.AuthConfig, error) {
	return a.auth, nil
}
