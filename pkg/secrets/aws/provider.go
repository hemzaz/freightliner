package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"freightliner/pkg/helper/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Provider implements credential storage and retrieval using AWS Secrets Manager
type Provider struct {
	client *secretsmanager.Client
	logger *log.Logger
	region string
}

// ProviderOptions contains configuration for the AWS Secrets Manager provider
type ProviderOptions struct {
	Region string
	Logger *log.Logger
}

// NewProvider creates a new AWS Secrets Manager provider
func NewProvider(ctx context.Context, opts ProviderOptions) (*Provider, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(opts.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	return &Provider{
		client: client,
		logger: opts.Logger,
		region: opts.Region,
	}, nil
}

// GetSecret retrieves a secret value by name
func (p *Provider) GetSecret(ctx context.Context, secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := p.client.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get secret value for '%s': %w", secretName, err)
	}

	// AWS returns either SecretString or SecretBinary
	var secretValue string
	if result.SecretString != nil {
		secretValue = *result.SecretString
	} else if result.SecretBinary != nil {
		// If the secret is binary, we'll need to decode it
		secretValue = string(result.SecretBinary)
	} else {
		return "", fmt.Errorf("retrieved secret has no value")
	}

	return secretValue, nil
}

// GetJSONSecret retrieves a JSON-formatted secret and unmarshal it into the provided struct
func (p *Provider) GetJSONSecret(ctx context.Context, secretName string, v interface{}) error {
	secretValue, err := p.GetSecret(ctx, secretName)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(secretValue), v); err != nil {
		return fmt.Errorf("failed to unmarshal secret JSON: %w", err)
	}

	return nil
}

// PutSecret creates or updates a secret value
func (p *Provider) PutSecret(ctx context.Context, secretName, secretValue string) error {
	input := &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretName),
		SecretString: aws.String(secretValue),
	}

	_, err := p.client.PutSecretValue(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put secret value for '%s': %w", secretName, err)
	}

	return nil
}

// PutJSONSecret marshals a struct to JSON and stores it as a secret
func (p *Provider) PutJSONSecret(ctx context.Context, secretName string, v interface{}) error {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal secret to JSON: %w", err)
	}

	return p.PutSecret(ctx, secretName, string(jsonBytes))
}

// DeleteSecret deletes a secret
func (p *Provider) DeleteSecret(ctx context.Context, secretName string) error {
	input := &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretName),
		ForceDeleteWithoutRecovery: aws.Bool(false), // Use recovery window
	}

	_, err := p.client.DeleteSecret(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete secret '%s': %w", secretName, err)
	}

	return nil
}

// ListSecrets lists all secrets with an optional filter
func (p *Provider) ListSecrets(ctx context.Context, filter string) ([]string, error) {
	// Create a list request without filters
	input := &secretsmanager.ListSecretsInput{}

	// We'll manually filter the results
	result, err := p.client.ListSecrets(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var secretNames []string
	for _, entry := range result.SecretList {
		if entry.Name != nil {
			// Apply filter if provided
			name := *entry.Name
			if filter == "" || (filter != "" && strings.Contains(name, filter)) {
				secretNames = append(secretNames, name)
			}
		}
	}

	// Handle pagination if necessary
	for result.NextToken != nil {
		input.NextToken = result.NextToken
		result, err = p.client.ListSecrets(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets (pagination): %w", err)
		}

		for _, entry := range result.SecretList {
			if entry.Name != nil {
				// Apply filter if provided
				name := *entry.Name
				if filter == "" || (filter != "" && strings.Contains(name, filter)) {
					secretNames = append(secretNames, name)
				}
			}
		}
	}

	return secretNames, nil
}
