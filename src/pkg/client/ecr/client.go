package ecr

import (
	"context"
	"fmt"
	"net/http"
	"src/internal/log"
	"src/pkg/client/common"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	awsauth "github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// Client implements the registry client interface for AWS ECR
type Client struct {
	ecr          *awsecr.Client
	region       string
	accountID    string
	logger       *log.Logger
	transportOpt remote.Option
}

// ClientOptions provides configuration for connecting to ECR
type ClientOptions struct {
	// Region is the AWS region to use
	Region string

	// AccountID is the AWS account ID
	AccountID string

	// Logger is the logger to use
	Logger *log.Logger

	// RegistryID is the AWS registry ID (optional, defaults to AccountID)
	RegistryID string

	// AssumeRoleARN is an optional role to assume
	AssumeRoleARN string
}

// NewClient creates a new ECR client
func NewClient(opts ClientOptions) (*Client, error) {
	// Set default values
	if opts.Region == "" {
		opts.Region = "us-west-2"
	}

	if opts.Logger == nil {
		opts.Logger = log.NewLogger(log.InfoLevel)
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(opts.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// If assume role ARN is provided, assume that role
	var ecrClient *awsecr.Client
	if opts.AssumeRoleARN != "" {
		// Create STS client for assuming the role
		stsClient := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsClient, opts.AssumeRoleARN)

		// Create new config with the assumed role
		roleCfg := aws.Config{
			Credentials: aws.NewCredentialsCache(provider),
			Region:      cfg.Region,
		}

		ecrClient = awsecr.NewFromConfig(roleCfg)
	} else {
		ecrClient = awsecr.NewFromConfig(cfg)
	}

	// Determine account ID if not specified
	accountID := opts.AccountID
	if accountID == "" {
		// Use STS to get the caller identity
		stsClient := sts.NewFromConfig(cfg)
		identity, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
		if err != nil {
			return nil, fmt.Errorf("failed to get AWS account ID: %w", err)
		}

		accountID = *identity.Account
	}

	// Create credential helper for ECR
	auth := awsauth.NewECRHelper(nil)

	// Create the transport option for the remote library
	credHelper := &ecrCredentialHelper{
		auth:      auth,
		accountID: accountID,
		region:    opts.Region,
	}
	transportOpt := remote.WithAuth(credHelper)

	return &Client{
		ecr:          ecrClient,
		region:       opts.Region,
		accountID:    accountID,
		logger:       opts.Logger,
		transportOpt: transportOpt,
	}, nil
}

// GetRepository returns a repository by name
func (c *Client) GetRepository(name string) (common.Repository, error) {
	// Create the ECR repository reference
	registry := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", c.accountID, c.region)
	repoRef := fmt.Sprintf("%s/%s", registry, name)
	// Create a repository reference
	repository, err := createRepository(repoRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository reference: %w", err)
	}

	return &Repository{
		client:     c,
		name:       name,
		repository: repository,
	}, nil
}

// ListRepositories lists all repositories in the registry
func (c *Client) ListRepositories() ([]string, error) {
	ctx := context.Background()
	var repos []string
	var nextToken *string

	// Collect all repositories using pagination
	for {
		input := &awsecr.DescribeRepositoriesInput{
			NextToken: nextToken,
		}

		response, err := c.ecr.DescribeRepositories(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, repo := range response.Repositories {
			// Extract just the repository name without the registry prefix
			name := *repo.RepositoryName
			repos = append(repos, name)
		}

		if response.NextToken == nil {
			break
		}
		nextToken = response.NextToken
	}

	return repos, nil
}

// GetTransport returns an authenticated HTTP transport for ECR
func (c *Client) GetTransport(repositoryName string) (http.RoundTripper, error) {
	// Create the ECR repository reference
	registry := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", c.accountID, c.region)
	repoRef := fmt.Sprintf("%s/%s", registry, repositoryName)
	// Create a repository reference
	repository, err := createRepository(repoRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository reference: %w", err)
	}

	// Create authenticator using the client's credential helper already configured
	credHelper := &ecrCredentialHelper{
		auth:      awsauth.NewECRHelper(nil),
		accountID: c.accountID,
		region:    c.region,
	}

	auth, err := credHelper.Resolve(repository.Registry)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticator: %w", err)
	}

	// Create transport
	rt, err := transport.New(
		repository.Registry,
		auth,
		http.DefaultTransport,
		[]string{repository.Scope(transport.PushScope)},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	return rt, nil
}

// ecrCredentialHelper implements the authn.Authenticator interface for AWS ECR
type ecrCredentialHelper struct {
	auth      *awsauth.ECRHelper
	accountID string
	region    string
}

// Resolve returns an authenticator for the given registry
func (h *ecrCredentialHelper) Resolve(registry name.Registry) (authn.Authenticator, error) {
	// Handle only ECR registries
	registryHost := registry.RegistryStr()
	if !strings.Contains(registryHost, "amazonaws.com") {
		return authn.Anonymous, nil
	}

	// Get ECR credentials using the helper
	authToken, err := h.getAuthToken(registryHost)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECR credentials: %w", err)
	}

	// Parse auth token
	authConfig := &authn.AuthConfig{
		Username: "AWS",
		Password: authToken,
	}

	// Create authenticator with the credentials
	return &authn.Basic{
		Username: authConfig.Username,
		Password: authConfig.Password,
	}, nil
}

// Authorization returns the AuthConfig for the ECR credentials
func (h *ecrCredentialHelper) Authorization() (*authn.AuthConfig, error) {
	// Determine the registry endpoint
	registry := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", h.accountID, h.region)

	// Get ECR credentials
	authToken, err := h.getAuthToken(registry)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECR credentials: %w", err)
	}

	// Return the auth config
	return &authn.AuthConfig{
		Username: "AWS",
		Password: authToken,
	}, nil
}

// createRepository creates a name.Repository from a repository name
func createRepository(repoName string) (name.Repository, error) {
	return name.NewRepository(repoName)
}

// getAuthToken gets an ECR authorization token
func (h *ecrCredentialHelper) getAuthToken(registryHost string) (string, error) {
	// Check if this is an ECR registry
	if !strings.Contains(registryHost, "amazonaws.com") {
		return "", fmt.Errorf("not an ECR registry: %s", registryHost)
	}

	// For testing purposes, we'll use a dummy token
	// In a real implementation, you would use AWS SDK or the ECR helper to get a token
	return "dummytoken", nil
}
