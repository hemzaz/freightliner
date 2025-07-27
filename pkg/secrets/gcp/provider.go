package gcp

import (
	"context"
	"encoding/json"
	"fmt"

	"freightliner/pkg/helper/log"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Provider implements credential storage and retrieval using Google Secret Manager
type Provider struct {
	client  *secretmanager.Client
	logger  *log.Logger
	project string
}

// ProviderOptions contains configuration for the Google Secret Manager provider
type ProviderOptions struct {
	Project         string
	CredentialsFile string
	Logger          *log.Logger
}

// NewProvider creates a new Google Secret Manager provider
func NewProvider(ctx context.Context, opts ProviderOptions) (*Provider, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if opts.Project == "" {
		return nil, fmt.Errorf("project is required")
	}

	var clientOpts []option.ClientOption
	if opts.CredentialsFile != "" {
		clientOpts = append(clientOpts, option.WithCredentialsFile(opts.CredentialsFile))
	}

	client, err := secretmanager.NewClient(ctx, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Secret Manager client: %w", err)
	}

	return &Provider{
		client:  client,
		logger:  opts.Logger,
		project: opts.Project,
	}, nil
}

// buildSecretName creates a fully-qualified secret name
func (p *Provider) buildSecretName(secretName string) string {
	return fmt.Sprintf("projects/%s/secrets/%s", p.project, secretName)
}

// buildSecretVersionName creates a fully-qualified secret version name
func (p *Provider) buildSecretVersionName(secretName string) string {
	return fmt.Sprintf("%s/versions/latest", p.buildSecretName(secretName))
}

// GetSecret retrieves a secret value by name
func (p *Provider) GetSecret(ctx context.Context, secretName string) (string, error) {
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: p.buildSecretVersionName(secretName),
	}

	result, err := p.client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %w", secretName, err)
	}

	return string(result.Payload.Data), nil
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

// CreateSecret creates a new secret
func (p *Provider) CreateSecret(ctx context.Context, secretName string) error {
	createRequest := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", p.project),
		SecretId: secretName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	_, err := p.client.CreateSecret(ctx, createRequest)
	if err != nil {
		return fmt.Errorf("failed to create secret %s: %w", secretName, err)
	}

	return nil
}

// secretExists checks if a secret already exists
func (p *Provider) secretExists(ctx context.Context, secretName string) (bool, error) {
	request := &secretmanagerpb.GetSecretRequest{
		Name: p.buildSecretName(secretName),
	}

	_, err := p.client.GetSecret(ctx, request)
	if err != nil {
		// If the error is "NotFound", the secret doesn't exist
		if err.Error() == "rpc error: code = NotFound desc = Secret not found" {
			return false, nil
		}
		return false, fmt.Errorf("error checking if secret exists: %w", err)
	}
	return true, nil
}

// PutSecret creates or updates a secret value
func (p *Provider) PutSecret(ctx context.Context, secretName, secretValue string) error {
	// Check if the secret exists, create it if it doesn't
	exists, err := p.secretExists(ctx, secretName)
	if err != nil {
		return err
	}

	if !exists {
		if createErr := p.CreateSecret(ctx, secretName); createErr != nil {
			return createErr
		}
	}

	// Add a new version to the secret with the provided value
	addVersionRequest := &secretmanagerpb.AddSecretVersionRequest{
		Parent: p.buildSecretName(secretName),
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue),
		},
	}

	_, err = p.client.AddSecretVersion(ctx, addVersionRequest)
	if err != nil {
		return fmt.Errorf("failed to add secret version: %w", err)
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

// DeleteSecret deletes a secret and all its versions
func (p *Provider) DeleteSecret(ctx context.Context, secretName string) error {
	deleteRequest := &secretmanagerpb.DeleteSecretRequest{
		Name: p.buildSecretName(secretName),
	}

	err := p.client.DeleteSecret(ctx, deleteRequest)
	if err != nil {
		return fmt.Errorf("failed to delete secret %s: %w", secretName, err)
	}

	return nil
}

// ListSecrets lists all secrets in the project
func (p *Provider) ListSecrets(ctx context.Context) ([]string, error) {
	parent := fmt.Sprintf("projects/%s", p.project)
	request := &secretmanagerpb.ListSecretsRequest{
		Parent: parent,
	}

	it := p.client.ListSecrets(ctx, request)
	var secretNames []string

	for {
		secret, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}

		// Extract the secret name from the full resource name
		// Format: projects/PROJECT_ID/secrets/SECRET_ID
		secretNames = append(secretNames, secret.Name)
	}

	return secretNames, nil
}
