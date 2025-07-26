package gcr

import (
	"strings"

	"freightliner/pkg/helper/errors"

	"github.com/google/go-containerregistry/pkg/name"
)

// parseGCRRepository parses a GCR repository URI and returns the registry and repository name
func parseGCRRepository(repoURI string, defaultRegistry name.Registry) (name.Registry, string, error) {
	// Check if the URI includes a registry
	if !strings.Contains(repoURI, "/") {
		return name.Registry{}, "", errors.InvalidInputf("invalid GCR repository format: %s", repoURI)
	}

	// If the URI starts with a registry, parse it
	var registry name.Registry
	var repository string
	var err error

	// Check if this is a full URI with registry (e.g., gcr.io/project/repo)
	if strings.Contains(repoURI, ".") && strings.Contains(repoURI, "/") {
		// Extract the registry and repository parts
		parts := strings.SplitN(repoURI, "/", 2)
		if len(parts) != 2 {
			return name.Registry{}, "", errors.InvalidInputf("invalid GCR repository format: %s", repoURI)
		}

		// Parse the registry
		registry, err = name.NewRegistry(parts[0])
		if err != nil {
			return name.Registry{}, "", errors.Wrap(err, "failed to parse registry")
		}

		// The rest is the repository
		repository = parts[1]

		// If the repository has a tag or digest, remove it
		if strings.Contains(repository, ":") {
			repository = strings.Split(repository, ":")[0]
		}
		if strings.Contains(repository, "@") {
			repository = strings.Split(repository, "@")[0]
		}
	} else {
		// Use the default registry and the full string as repository
		registry = defaultRegistry
		repository = repoURI
	}

	return registry, repository, nil
}

// isGCRRegistry returns true if the registry is a GCR registry
func isGCRRegistry(registry string) bool {
	if registry == "" {
		return false
	}

	return registry == "gcr.io" || strings.HasSuffix(registry, ".gcr.io")
}

// extractProjectFromRepository extracts the GCP project ID from a repository path
func extractProjectFromRepository(repository string) (string, error) {
	if repository == "" {
		return "", errors.InvalidInputf("repository path cannot be empty")
	}

	parts := strings.SplitN(repository, "/", 2)
	if len(parts) < 2 {
		return "", errors.InvalidInputf("invalid repository path format, expected project/repo: %s", repository)
	}

	return parts[0], nil
}
