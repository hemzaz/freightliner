// Package ecr provides AWS Elastic Container Registry client functionality.
package ecr

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	awsauth "github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// ECRServiceAPI is an interface for AWS ECR operations
// This interface is defined in the same package as its implementation because
// it represents an adapter for a third-party service and is only used internally.
// It helps with testing by allowing us to mock the AWS ECR API.
type ECRServiceAPI interface {
	ListImages(ctx context.Context, params *awsecr.ListImagesInput, optFns ...func(*awsecr.Options)) (*awsecr.ListImagesOutput, error)
	BatchGetImage(ctx context.Context, params *awsecr.BatchGetImageInput, optFns ...func(*awsecr.Options)) (*awsecr.BatchGetImageOutput, error)
	PutImage(ctx context.Context, params *awsecr.PutImageInput, optFns ...func(*awsecr.Options)) (*awsecr.PutImageOutput, error)
	BatchDeleteImage(ctx context.Context, params *awsecr.BatchDeleteImageInput, optFns ...func(*awsecr.Options)) (*awsecr.BatchDeleteImageOutput, error)
	DescribeImages(ctx context.Context, params *awsecr.DescribeImagesInput, optFns ...func(*awsecr.Options)) (*awsecr.DescribeImagesOutput, error)
	DescribeRepositories(ctx context.Context, params *awsecr.DescribeRepositoriesInput, optFns ...func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error)
	CreateRepository(ctx context.Context, params *awsecr.CreateRepositoryInput, optFns ...func(*awsecr.Options)) (*awsecr.CreateRepositoryOutput, error)
	GetAuthorizationToken(ctx context.Context, params *awsecr.GetAuthorizationTokenInput, optFns ...func(*awsecr.Options)) (*awsecr.GetAuthorizationTokenOutput, error)
}

// Client implements the registry client interface for AWS ECR
type Client struct {
	ecr          ECRServiceAPI
	region       string
	accountID    string
	logger       log.Logger
	transportOpt remote.Option
}

// ClientOptions provides configuration for connecting to ECR
type ClientOptions struct {
	// Region is the AWS region for ECR
	Region string

	// AccountID is the AWS account ID (optional, uses default credentials if empty)
	AccountID string

	// Profile is the AWS profile to use (optional)
	Profile string

	// RoleARN is an optional IAM role to assume for ECR operations
	RoleARN string

	// CredentialsFile is the path to AWS credentials file (optional)
	CredentialsFile string

	// Logger is the logger to use
	Logger log.Logger
}

// LegacyClientOptions is an alias for ClientOptions for backward compatibility
type LegacyClientOptions struct {
	// Region is the AWS region for ECR
	Region string

	// AccountID is the AWS account ID (optional)
	AccountID string

	// Internal STS client for testing
	stsClient STSServiceAPI
	Logger    log.Logger
}

// STSServiceAPI is an interface for AWS STS API operations
type STSServiceAPI interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// normalizeClientOptions normalizes different option types into a standard ClientOptions
func normalizeClientOptions(optsArg interface{}) (ClientOptions, error) {
	switch o := optsArg.(type) {
	case ClientOptions:
		return o, nil
	case LegacyClientOptions:
		return ClientOptions{
			Region:    o.Region,
			AccountID: o.AccountID,
			Logger:    o.Logger,
		}, nil
	default:
		return ClientOptions{}, errors.InvalidInputf("invalid options type")
	}
}

// validateClientOptions validates the client options
func validateClientOptions(opts *ClientOptions) error {
	if opts.Region == "" {
		return errors.InvalidInputf("AWS region is required")
	}

	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	return nil
}

// createAWSConfig creates an AWS SDK config based on the provided options
func createAWSConfig(ctx context.Context, opts *ClientOptions) (aws.Config, error) {
	var configOpts []func(*config.LoadOptions) error
	configOpts = append(configOpts, config.WithRegion(opts.Region))

	// Use profile if specified
	if opts.Profile != "" {
		configOpts = append(configOpts, config.WithSharedConfigProfile(opts.Profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return aws.Config{}, errors.Wrap(err, "failed to load AWS config")
	}

	return cfg, nil
}

// createECRClient creates an ECR client with the provided AWS config, optionally assuming a role
func createECRClient(cfg aws.Config, roleARN string) (*awsecr.Client, error) {
	if roleARN == "" {
		// Use default credentials
		return awsecr.NewFromConfig(cfg), nil
	}

	// Create an STS client to assume the role
	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)

	// Create a new config with the role credentials
	roleCfg := aws.Config{
		Credentials: aws.NewCredentialsCache(provider),
		Region:      cfg.Region,
	}

	// Create an ECR client with the assumed role credentials
	return awsecr.NewFromConfig(roleCfg), nil
}

// NewClient creates a new ECR client
func NewClient(optsArg interface{}) (*Client, error) {
	// Normalize and validate options
	opts, err := normalizeClientOptions(optsArg)
	if err != nil {
		return nil, err
	}

	if validateErr := validateClientOptions(&opts); validateErr != nil {
		return nil, validateErr
	}

	// Create AWS config
	cfg, err := createAWSConfig(context.Background(), &opts)
	if err != nil {
		return nil, err
	}

	// Create ECR client
	ecrClient, err := createECRClient(cfg, opts.RoleARN)
	if err != nil {
		return nil, err
	}

	// Create authenticator
	auth := NewECRAuthenticator(ecrClient, opts.Region)

	// Create client
	return &Client{
		ecr:          ecrClient,
		region:       opts.Region,
		accountID:    opts.AccountID,
		logger:       opts.Logger,
		transportOpt: remote.WithAuth(auth),
	}, nil
}

// GetRegistryName returns the ECR registry endpoint
func (c *Client) GetRegistryName() string {
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", c.accountID, c.region)
}

// ListRepositories lists all repositories in the registry
func (c *Client) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	var repositories []string
	var nextToken *string

	for {
		input := &awsecr.DescribeRepositoriesInput{
			NextToken: nextToken,
		}

		// Apply account ID if specified
		if c.accountID != "" {
			input.RegistryId = &c.accountID
		}

		// Call the ECR API
		resp, err := c.ecr.DescribeRepositories(ctx, input)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list ECR repositories")
		}

		// Process the response
		for _, repo := range resp.Repositories {
			repoName := *repo.RepositoryName
			if prefix == "" || strings.HasPrefix(repoName, prefix) {
				repositories = append(repositories, repoName)
			}
		}

		// Continue pagination if more results
		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}

	return repositories, nil
}

// GetRepository returns a repository by name
func (c *Client) GetRepository(ctx context.Context, repoName string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Create a proper repository reference
	registry := c.GetRegistryName()
	fullRepoPath := fmt.Sprintf("%s/%s", registry, repoName)

	// Debug logging
	c.logger.WithFields(map[string]interface{}{
		"repository_name": repoName,
		"registry":        registry,
		"account_id":      c.accountID,
		"region":          c.region,
		"full_path":       fullRepoPath,
	}).Debug("Getting ECR repository")

	repository, err := name.NewRepository(fullRepoPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Verify repository exists by calling DescribeRepositories
	input := &awsecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repoName},
	}

	if c.accountID != "" {
		input.RegistryId = aws.String(c.accountID)
	}

	c.logger.WithFields(map[string]interface{}{
		"repository_name": repoName,
		"registry_id":     c.accountID,
	}).Debug("Verifying repository exists in ECR")

	resp, err := c.ecr.DescribeRepositories(ctx, input)
	if err != nil {
		c.logger.WithFields(map[string]interface{}{
			"repository_name": repoName,
			"registry_id":     c.accountID,
			"error":           err.Error(),
		}).Warn("Failed to verify repository existence")
		return nil, errors.Wrap(err, "repository does not exist or cannot be accessed")
	}

	if len(resp.Repositories) == 0 {
		c.logger.WithFields(map[string]interface{}{
			"repository_name": repoName,
			"registry_id":     c.accountID,
		}).Warn("Repository not found in DescribeRepositories response")
		return nil, errors.NotFoundf("repository not found: %s", repoName)
	}

	c.logger.WithFields(map[string]interface{}{
		"repository_name": repoName,
		"repository_arn":  *resp.Repositories[0].RepositoryArn,
	}).Info("Repository verified successfully")

	return &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}, nil
}

// CreateRepository creates a new repository in ECR
func (c *Client) CreateRepository(ctx context.Context, repoName string, tags map[string]string) (interfaces.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Convert tags to ECR tag format
	ecrTags := make([]ecrtypes.Tag, 0, len(tags))
	for k, v := range tags {
		key, value := k, v
		ecrTags = append(ecrTags, ecrtypes.Tag{
			Key:   &key,
			Value: &value,
		})
	}

	// Call the ECR API to create the repository
	input := &awsecr.CreateRepositoryInput{
		RepositoryName: aws.String(repoName),
		Tags:           ecrTags,
	}

	if c.accountID != "" {
		input.RegistryId = aws.String(c.accountID)
	}

	resp, err := c.ecr.CreateRepository(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ECR repository")
	}

	// Return a Repository interface
	registry := c.GetRegistryName()
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", registry, repoName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	c.logger.WithFields(map[string]interface{}{
		"repository": repoName,
		"registry":   registry,
		"arn":        *resp.Repository.RepositoryArn,
	}).Info("Created ECR repository")

	return &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}, nil
}

// GetTransport returns an authenticated HTTP transport for ECR
func (c *Client) GetTransport(repositoryName string) (http.RoundTripper, error) {
	// Create a proper repository reference
	registry := c.GetRegistryName()
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", registry, repositoryName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	// Create a new authenticator for ECR
	auth := NewECRAuthenticator(c.ecr, c.region)

	// Create transport with authentication
	rt, err := transport.NewWithContext(
		context.Background(),
		repository.Registry,
		auth,
		http.DefaultTransport,
		[]string{repository.Scope(transport.PushScope)},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ECR transport")
	}

	return rt, nil
}

// GetRemoteOptions returns options for the go-containerregistry remote package
func (c *Client) GetRemoteOptions() []remote.Option {
	return []remote.Option{
		c.transportOpt,
	}
}

// GetDefaultCredentialHelper returns the default credential helper for ECR
func GetDefaultCredentialHelper() *awsauth.ECRHelper {
	return &awsauth.ECRHelper{}
}
