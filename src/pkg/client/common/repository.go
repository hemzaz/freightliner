package common

// Repository represents a container repository in a registry
type Repository interface {
	// GetRepositoryName returns the name of the repository
	GetRepositoryName() string

	// ListTags returns all tags for the repository
	ListTags() ([]string, error)
	
	// GetManifest returns the manifest for the given tag
	GetManifest(tag string) ([]byte, string, error)
	
	// PutManifest uploads a manifest with the given tag
	PutManifest(tag string, manifest []byte, mediaType string) error
	
	// DeleteManifest deletes the manifest for the given tag
	DeleteManifest(tag string) error
}
