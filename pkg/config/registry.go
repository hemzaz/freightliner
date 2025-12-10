package config

import (
	"fmt"
	"net/url"
	"strings"
)

// RegistryType represents the type of container registry
type RegistryType string

const (
	// RegistryTypeECR represents AWS Elastic Container Registry
	RegistryTypeECR RegistryType = "ecr"
	// RegistryTypeGCR represents Google Container Registry
	RegistryTypeGCR RegistryType = "gcr"
	// RegistryTypeDockerHub represents Docker Hub
	RegistryTypeDockerHub RegistryType = "dockerhub"
	// RegistryTypeHarbor represents Harbor registry
	RegistryTypeHarbor RegistryType = "harbor"
	// RegistryTypeQuay represents Quay.io registry
	RegistryTypeQuay RegistryType = "quay"
	// RegistryTypeGitLab represents GitLab Container Registry
	RegistryTypeGitLab RegistryType = "gitlab"
	// RegistryTypeGitHub represents GitHub Container Registry
	RegistryTypeGitHub RegistryType = "github"
	// RegistryTypeAzure represents Azure Container Registry
	RegistryTypeAzure RegistryType = "azure"
	// RegistryTypeGeneric represents a generic OCI-compliant registry
	RegistryTypeGeneric RegistryType = "generic"
)

// AuthType represents the authentication method for a registry
type AuthType string

const (
	// AuthTypeBasic represents basic username/password authentication
	AuthTypeBasic AuthType = "basic"
	// AuthTypeToken represents bearer token authentication
	AuthTypeToken AuthType = "token"
	// AuthTypeAWS represents AWS IAM authentication (for ECR)
	AuthTypeAWS AuthType = "aws"
	// AuthTypeGCP represents GCP service account authentication (for GCR)
	AuthTypeGCP AuthType = "gcp"
	// AuthTypeOAuth represents OAuth2 authentication
	AuthTypeOAuth AuthType = "oauth"
	// AuthTypeAnonymous represents anonymous (no authentication)
	AuthTypeAnonymous AuthType = "anonymous"
)

// RegistryConfig represents configuration for a single container registry
type RegistryConfig struct {
	// Name is a unique identifier for this registry configuration
	Name string `yaml:"name" json:"name"`

	// Type is the registry type (ecr, gcr, dockerhub, harbor, etc.)
	Type RegistryType `yaml:"type" json:"type"`

	// Endpoint is the registry endpoint URL (e.g., "https://registry.hub.docker.com")
	// For cloud registries, this can be auto-generated from region/project
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`

	// Region is the cloud region (for AWS ECR, GCR, Azure ACR)
	Region string `yaml:"region,omitempty" json:"region,omitempty"`

	// Project is the project/account identifier (for GCR, GitLab, etc.)
	Project string `yaml:"project,omitempty" json:"project,omitempty"`

	// AccountID is the account identifier (for AWS ECR)
	AccountID string `yaml:"account_id,omitempty" json:"account_id,omitempty"`

	// Auth contains authentication configuration
	Auth AuthConfig `yaml:"auth" json:"auth"`

	// TLS contains TLS configuration
	TLS TLSConfig `yaml:"tls,omitempty" json:"tls,omitempty"`

	// Insecure allows insecure connections (skip TLS verification)
	Insecure bool `yaml:"insecure,omitempty" json:"insecure,omitempty"`

	// Timeout is the connection timeout in seconds (default: 30)
	Timeout int `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// RetryAttempts is the number of retry attempts for failed operations
	RetryAttempts int `yaml:"retry_attempts,omitempty" json:"retry_attempts,omitempty"`

	// Metadata contains additional registry-specific metadata
	Metadata map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// AuthConfig represents authentication configuration for a registry
type AuthConfig struct {
	// Type is the authentication type (basic, token, aws, gcp, oauth, anonymous)
	Type AuthType `yaml:"type" json:"type"`

	// Username for basic authentication
	Username string `yaml:"username,omitempty" json:"username,omitempty"`

	// Password for basic authentication
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Token for bearer token authentication
	Token string `yaml:"token,omitempty" json:"token,omitempty"`

	// CredentialsFile is the path to a credentials file (for GCP service accounts, AWS credentials)
	CredentialsFile string `yaml:"credentials_file,omitempty" json:"credentials_file,omitempty"`

	// UseSecretsManager indicates whether to fetch credentials from secrets manager
	UseSecretsManager bool `yaml:"use_secrets_manager,omitempty" json:"use_secrets_manager,omitempty"`

	// SecretName is the name of the secret in the secrets manager
	SecretName string `yaml:"secret_name,omitempty" json:"secret_name,omitempty"`

	// Profile is the AWS profile to use (for AWS authentication)
	Profile string `yaml:"profile,omitempty" json:"profile,omitempty"`

	// RoleARN is the AWS IAM role ARN to assume (for AWS authentication)
	RoleARN string `yaml:"role_arn,omitempty" json:"role_arn,omitempty"`
}

// TLSConfig represents TLS configuration for registry connections
type TLSConfig struct {
	// CertFile is the path to the client certificate file
	CertFile string `yaml:"cert_file,omitempty" json:"cert_file,omitempty"`

	// KeyFile is the path to the client key file
	KeyFile string `yaml:"key_file,omitempty" json:"key_file,omitempty"`

	// CAFile is the path to the CA certificate file
	CAFile string `yaml:"ca_file,omitempty" json:"ca_file,omitempty"`

	// InsecureSkipVerify skips TLS certificate verification
	InsecureSkipVerify bool `yaml:"insecure_skip_verify,omitempty" json:"insecure_skip_verify,omitempty"`
}

// RegistriesConfig represents configuration for multiple registries
type RegistriesConfig struct {
	// DefaultSource is the default source registry name
	DefaultSource string `yaml:"default_source,omitempty" json:"default_source,omitempty"`

	// DefaultDestination is the default destination registry name
	DefaultDestination string `yaml:"default_destination,omitempty" json:"default_destination,omitempty"`

	// Registries is a list of configured registries
	Registries []RegistryConfig `yaml:"registries" json:"registries"`
}

// Validate validates the registry configuration
func (r *RegistryConfig) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("registry name is required")
	}

	if r.Type == "" {
		return fmt.Errorf("registry type is required for registry %s", r.Name)
	}

	// Validate registry-specific requirements
	switch r.Type {
	case RegistryTypeECR:
		if r.Region == "" {
			return fmt.Errorf("region is required for ECR registry %s", r.Name)
		}
	case RegistryTypeGCR:
		if r.Project == "" {
			return fmt.Errorf("project is required for GCR registry %s", r.Name)
		}
	case RegistryTypeDockerHub, RegistryTypeHarbor, RegistryTypeQuay, RegistryTypeGeneric:
		if r.Endpoint == "" {
			// Set default endpoints for known registries
			r.Endpoint = r.GetDefaultEndpoint()
		}
	}

	// Validate authentication configuration
	if err := r.Auth.Validate(r.Type); err != nil {
		return fmt.Errorf("invalid auth config for registry %s: %w", r.Name, err)
	}

	return nil
}

// Validate validates the authentication configuration
func (a *AuthConfig) Validate(registryType RegistryType) error {
	if a.Type == "" {
		// Set default auth type based on registry type
		switch registryType {
		case RegistryTypeECR:
			a.Type = AuthTypeAWS
		case RegistryTypeGCR:
			a.Type = AuthTypeGCP
		case RegistryTypeDockerHub, RegistryTypeHarbor, RegistryTypeQuay:
			if a.Username != "" || a.Password != "" {
				a.Type = AuthTypeBasic
			} else {
				a.Type = AuthTypeAnonymous
			}
		default:
			a.Type = AuthTypeBasic
		}
	}

	// Validate required fields based on auth type
	switch a.Type {
	case AuthTypeBasic:
		if !a.UseSecretsManager && a.Username == "" {
			return fmt.Errorf("username is required for basic authentication")
		}
	case AuthTypeToken:
		if !a.UseSecretsManager && a.Token == "" {
			return fmt.Errorf("token is required for token authentication")
		}
	case AuthTypeAWS:
		// AWS credentials are typically from environment or IAM role
		// No validation needed
	case AuthTypeGCP:
		// GCP credentials can be from environment or credentials file
		// No strict validation needed
	}

	return nil
}

// GetDefaultEndpoint returns the default endpoint for known registry types
func (r *RegistryConfig) GetDefaultEndpoint() string {
	switch r.Type {
	case RegistryTypeDockerHub:
		return "https://registry-1.docker.io"
	case RegistryTypeQuay:
		return "https://quay.io"
	case RegistryTypeGitHub:
		return "https://ghcr.io"
	case RegistryTypeECR:
		if r.Region != "" && r.AccountID != "" {
			return fmt.Sprintf("https://%s.dkr.ecr.%s.amazonaws.com", r.AccountID, r.Region)
		}
	case RegistryTypeGCR:
		location := r.Region
		if location == "" {
			location = "us"
		}
		return fmt.Sprintf("https://%s.gcr.io", location)
	case RegistryTypeAzure:
		if r.Project != "" {
			return fmt.Sprintf("https://%s.azurecr.io", r.Project)
		}
	}
	return r.Endpoint
}

// GetRegistryHost returns the registry host (without protocol)
func (r *RegistryConfig) GetRegistryHost() (string, error) {
	endpoint := r.Endpoint
	if endpoint == "" {
		endpoint = r.GetDefaultEndpoint()
	}

	// Parse the endpoint URL
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint URL: %w", err)
	}

	return u.Host, nil
}

// GetImageReference returns the full image reference for a given repository and tag
func (r *RegistryConfig) GetImageReference(repository, tag string) (string, error) {
	host, err := r.GetRegistryHost()
	if err != nil {
		return "", err
	}

	// Remove leading slash from repository if present
	repository = strings.TrimPrefix(repository, "/")

	// Construct the image reference
	if tag == "" {
		return fmt.Sprintf("%s/%s", host, repository), nil
	}
	return fmt.Sprintf("%s/%s:%s", host, repository, tag), nil
}

// GetByName finds a registry configuration by name
func (rc *RegistriesConfig) GetByName(name string) (*RegistryConfig, error) {
	for i := range rc.Registries {
		if rc.Registries[i].Name == name {
			return &rc.Registries[i], nil
		}
	}
	return nil, fmt.Errorf("registry %s not found in configuration", name)
}

// GetByType returns all registries of a specific type
func (rc *RegistriesConfig) GetByType(registryType RegistryType) []RegistryConfig {
	var registries []RegistryConfig
	for _, r := range rc.Registries {
		if r.Type == registryType {
			registries = append(registries, r)
		}
	}
	return registries
}

// Validate validates all registry configurations
func (rc *RegistriesConfig) Validate() error {
	if len(rc.Registries) == 0 {
		return fmt.Errorf("no registries configured")
	}

	// Check for duplicate names
	names := make(map[string]bool)
	for i := range rc.Registries {
		if names[rc.Registries[i].Name] {
			return fmt.Errorf("duplicate registry name: %s", rc.Registries[i].Name)
		}
		names[rc.Registries[i].Name] = true

		// Validate each registry
		if err := rc.Registries[i].Validate(); err != nil {
			return err
		}
	}

	// Validate default registries exist
	if rc.DefaultSource != "" {
		if _, err := rc.GetByName(rc.DefaultSource); err != nil {
			return fmt.Errorf("default source registry not found: %w", err)
		}
	}

	if rc.DefaultDestination != "" {
		if _, err := rc.GetByName(rc.DefaultDestination); err != nil {
			return fmt.Errorf("default destination registry not found: %w", err)
		}
	}

	return nil
}
