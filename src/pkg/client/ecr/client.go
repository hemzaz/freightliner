package ecr

import (
	"context"
	"fmt"
	"net/http"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/elad/freightliner/internal/log"
	"github.com/elad/freightliner/pkg/client/common"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// Client implements the registry client interface for AWS ECR
type Client struct {
	ecrClient    *ecr.Client
	region       string
	accountID    string
	keychain     authn.Keychain
	transportOpt remote.Option
	logger       *log.Logger
}

// ClientOptions contains options for the ECR client
type ClientOptions struct {
	AccountID string
	Region    string
	Logger    *log.Logger
}

// NewClient creates a new ECR client
func NewClient(opts ClientOptions) (*Client, error) {
	ctx := context.Background()
	
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(opts.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	// Create ECR client
	ecrClient := ecr.NewFromConfig(cfg)
	
	// Create keychain
	keychain := NewECRKeychain(ecrClient, opts.Region, opts.AccountID)
	
	// Create transport option for remote.* operations
	transportOpt := remote.WithAuthFromKeychain(keychain)
	
	return &Client{
		ecrClient:    ecrClient,
		region:       opts.Region,
		accountID:    opts.AccountID,
		keychain:     keychain,
		transportOpt: transportOpt,
		logger:       opts.Logger,
	}, nil
}

// GetRepository returns a repository interface for the given repository name
func (c *Client) GetRepository(name string) (common.Repository, error) {
	// Parse repository name
	_, repo, err := ParseECRRepository(name, c.region)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repository name: %w", err)
	}
	
	// Check if repository exists
	ctx := context.Background()
	_, err = c.ecrClient.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repo.RepositoryStr()},
	})
	
	// If repository doesn't exist, create it
	if err != nil {
		c.logger.Info("Repository doesn't exist, creating it", map[string]interface{}{
			"repository": name,
		})
		
		_, err = c.ecrClient.CreateRepository(ctx, &ecr.CreateRepositoryInput{
			RepositoryName: aws.String(repo.RepositoryStr()),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create repository: %w", err)
		}
	}
	
	return &Repository{
		client:     c,
		name:       name,
		repository: repo,
	}, nil
}

// ListRepositories returns a list of all repositories in the registry
func (c *Client) ListRepositories() ([]string, error) {
	ctx := context.Background()
	var repositories []string
	var nextToken *string
	
	for {
		// Call the API to list repositories
		resp, err := c.ecrClient.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}
		
		// Add repositories to the list
		for _, repo := range resp.Repositories {
			repositories = append(repositories, aws.ToString(repo.RepositoryName))
		}
		
		// Check if there are more repositories
		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}
	
	return repositories, nil
}

// GetTransport returns a transport for the given repository
func (c *Client) GetTransport(repo string) (http.RoundTripper, error) {
	// Parse repository name
	_, repository, err := ParseECRRepository(repo, c.region)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repository name: %w", err)
	}
	
	// Get authenticator for this repository
	auth, err := c.keychain.Resolve(repository)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticator: %w", err)
	}
	
	// Create transport
	return transport.NewWithContext(
		context.Background(),
		repository.Registry,
		auth,
		transport.WithContext(context.Background()),
	)
}
