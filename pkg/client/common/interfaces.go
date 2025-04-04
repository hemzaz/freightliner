package common

import (
	"fmt"
	"strings"

	"freightliner/pkg/helper/errors"
)

// FormRegistryPath creates a properly formatted registry path
func FormRegistryPath(registry, name string) string {
	return fmt.Sprintf("%s/%s", registry, name)
}

// ParseRegistryPath parses a registry path into registry type and repository name
func ParseRegistryPath(path string) (string, string, error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return "", "", errors.InvalidInputf("invalid format. Use [registry]/[repository]")
	}
	return parts[0], parts[1], nil
}

// Note: The interfaces previously defined here have been moved to the packages that use them:
// - RegistryClient, Repository, RegistryAuthenticator have moved to pkg/service/interfaces.go
// - Repository, ManifestAccessor, LayerAccessor have moved to pkg/copy/interfaces.go
//
// This aligns with the Dependency Inversion Principle by defining interfaces where they are used,
// not where they are implemented.
