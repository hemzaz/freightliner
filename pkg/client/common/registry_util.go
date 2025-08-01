package common

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// RegistryUtil provides common functionality for registry clients
type RegistryUtil struct {
	logger log.Logger
}

// NewRegistryUtil creates a new registry utility instance
func NewRegistryUtil(logger log.Logger) *RegistryUtil {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &RegistryUtil{
		logger: logger,
	}
}

// ParseRegistryPath parses a registry path into registry type and repository name
func (u *RegistryUtil) ParseRegistryPath(path string) (string, string, error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return "", "", errors.InvalidInputf("invalid format. Use [registry]/[repository]")
	}
	return parts[0], parts[1], nil
}

// ValidateRepositoryName checks if a repository name is valid
func (u *RegistryUtil) ValidateRepositoryName(repoName string) error {
	if repoName == "" {
		return errors.InvalidInputf("repository name cannot be empty")
	}

	// Check for invalid characters or patterns in the repository name
	// For simplicity, we're only checking empty names here, but real implementations
	// should check for valid characters and patterns based on registry specifications

	return nil
}

// CreateRepositoryReference creates a repository reference for a given registry and repository name
func (u *RegistryUtil) CreateRepositoryReference(registry, repoName string) (name.Repository, error) {
	if err := u.ValidateRepositoryName(repoName); err != nil {
		return name.Repository{}, err
	}

	repoPath := fmt.Sprintf("%s/%s", registry, repoName)
	repository, err := name.NewRepository(repoPath)
	if err != nil {
		return name.Repository{}, errors.Wrap(err, "failed to create repository reference")
	}

	return repository, nil
}

// GetRemoteOptions returns the basic remote options for a registry
func (u *RegistryUtil) GetRemoteOptions(transport http.RoundTripper) []remote.Option {
	var options []remote.Option

	if transport != nil {
		options = append(options, remote.WithTransport(transport))
	}

	return options
}

// IsValidRegistryType checks if a registry type is supported
func (u *RegistryUtil) IsValidRegistryType(registryType string) bool {
	validTypes := map[string]bool{
		"ecr": true,
		"gcr": true,
	}

	return validTypes[registryType]
}

// FormatRepositoryURI formats a repository URI based on registry type
func (u *RegistryUtil) FormatRepositoryURI(registryType, accountID, region, repoName string) string {
	switch registryType {
	case "ecr":
		return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", accountID, region, repoName)
	case "gcr":
		return fmt.Sprintf("gcr.io/%s/%s", accountID, repoName)
	default:
		return fmt.Sprintf("%s/%s", registryType, repoName)
	}
}

// LogRegistryOperation logs registry operations with consistent format
func (u *RegistryUtil) LogRegistryOperation(ctx context.Context, operation, registry, repository string, extraFields map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":  operation,
		"registry":   registry,
		"repository": repository,
	}

	// Add any extra fields
	for k, v := range extraFields {
		fields[k] = v
	}

	u.logger.WithFields(fields).Info(fmt.Sprintf("Registry operation: %s", operation))
}
