package common

// RegistryClient defines the common interface for interacting with container registries
type RegistryClient interface {
	// GetRepository returns a repository interface for the given repository name
	GetRepository(name string) (Repository, error)
	
	// ListRepositories returns a list of all repositories in the registry
	ListRepositories() ([]string, error)
}
