package ecr

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"

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

// ECRServiceAPI interface for AWS ECR operations
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
	logger       *log.Logger
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
	Logger *log.Logger
}

// LegacyClientOptions is an alias for ClientOptions for backward compatibility
type LegacyClientOptions struct {
	// Region is the AWS region for ECR
	Region string

	// AccountID is the AWS account ID (optional)
	AccountID string

	// Internal STS client for testing
	stsClient STSServiceAPI
	Logger    *log.Logger
}

// STSServiceAPI is an interface for AWS STS API operations
type STSServiceAPI interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// NewClient creates a new ECR client
func NewClient(optsArg interface{}) (*Client, error) {
	var opts ClientOptions

	// Handle both ClientOptions and legacy Options
	switch o := optsArg.(type) {
	case ClientOptions:
		opts = o
	case LegacyClientOptions:
		opts = ClientOptions{
			Region:    o.Region,
			AccountID: o.AccountID,
			Logger:    o.Logger,
		}
		// If using legacy Options, we would handle stsClient here, but we've removed this
		// functionality to simplify the implementation
	default:
		return nil, errors.InvalidInputf("invalid options type")
	}
	if opts.Region == "" {
		return nil, errors.InvalidInputf("AWS region is required")
	}

	if opts.Logger == nil {
		opts.Logger = log.NewLogger(log.InfoLevel)
	}

	// Load AWS config
	var configOpts []func(*config.LoadOptions) error
	configOpts = append(configOpts, config.WithRegion(opts.Region))

	// Use profile if specified
	if opts.Profile != "" {
		configOpts = append(configOpts, config.WithSharedConfigProfile(opts.Profile))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), configOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load AWS config")
	}

	var ecrClient *awsecr.Client

	// If a role ARN is provided, assume the role and create an ECR client with the assumed role credentials
	if opts.RoleARN != "" {
		// Create an STS client to assume the role
		stsClient := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsClient, opts.RoleARN)

		// Create a new config with the role credentials
		roleCfg := aws.Config{
			Credentials: aws.NewCredentialsCache(provider),
			Region:      cfg.Region,
		}

		// Create an ECR client with the assumed role credentials
		ecrClient = awsecr.NewFromConfig(roleCfg)
	} else {
		// Use default credentials
		ecrClient = awsecr.NewFromConfig(cfg)
	}

	// Create authenticator for ECR
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
func (c *Client) GetRepository(ctx context.Context, repoName string) (common.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Create a proper repository reference
	registry := c.GetRegistryName()
	repository, err := name.NewRepository(fmt.Sprintf("%s/%s", registry, repoName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository reference")
	}

	return &Repository{
		client:     c,
		name:       repoName,
		repository: repository,
	}, nil
}

// CreateRepository creates a new repository in ECR
func (c *Client) CreateRepository(ctx context.Context, repoName string, tags map[string]string) (common.Repository, error) {
	if repoName == "" {
		return nil, errors.InvalidInputf("repository name cannot be empty")
	}

	// Convert tags to ECR tag format
	var ecrTags []ecrtypes.Tag
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

	c.logger.Info("Created ECR repository", map[string]interface{}{
		"repository": repoName,
		"registry":   registry,
		"arn":        *resp.Repository.RepositoryArn,
	})

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
	rt, err := transport.New(
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
