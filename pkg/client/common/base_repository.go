package common

import (
	"context"
	"sync"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// BaseRepository implements common functionality for repositories
type BaseRepository struct {
	name       string
	repository name.Repository
	logger     *log.Logger

	// Cache for tags and images
	tagsMutex sync.RWMutex
	tags      []string

	imagesMutex sync.RWMutex
	images      map[string]v1.Image
}

// BaseRepositoryOptions provides options for creating a base repository
type BaseRepositoryOptions struct {
	Name       string
	Repository name.Repository
	Logger     *log.Logger
}

// NewBaseRepository creates a new base repository for registry operations
func NewBaseRepository(opts BaseRepositoryOptions) *BaseRepository {
	if opts.Logger == nil {
		opts.Logger = log.NewLogger(log.InfoLevel)
	}

	return &BaseRepository{
		name:       opts.Name,
		repository: opts.Repository,
		logger:     opts.Logger,
		images:     make(map[string]v1.Image),
	}
}

// GetName returns the repository name
func (r *BaseRepository) GetName() string {
	return r.name
}

// GetURI returns the fully qualified repository URI
func (r *BaseRepository) GetURI() string {
	return r.repository.String()
}

// ListTags returns all tags in the repository
func (r *BaseRepository) ListTags(ctx context.Context) ([]string, error) {
	// This is a placeholder that should be overridden by specific implementations
	return nil, errors.NotImplementedf("ListTags must be implemented by specific repository implementations")
}

// GetTag returns a tagged image from the repository
func (r *BaseRepository) GetTag(ctx context.Context, tagName string) (v1.Image, error) {
	if tagName == "" {
		return nil, errors.InvalidInputf("tag name cannot be empty")
	}

	// Check the cache first
	r.imagesMutex.RLock()
	img, ok := r.images[tagName]
	r.imagesMutex.RUnlock()

	if ok {
		return img, nil
	}

	// This is a placeholder that should be overridden by specific implementations
	return nil, errors.NotImplementedf("GetTag must be implemented by specific repository implementations")
}

// GetImage returns an image by digest
func (r *BaseRepository) GetImage(ctx context.Context, digest string) (v1.Image, error) {
	if digest == "" {
		return nil, errors.InvalidInputf("digest cannot be empty")
	}

	// This is a placeholder that should be overridden by specific implementations
	return nil, errors.NotImplementedf("GetImage must be implemented by specific repository implementations")
}

// DeleteTag removes a tag from the repository
func (r *BaseRepository) DeleteTag(ctx context.Context, tagName string) error {
	if tagName == "" {
		return errors.InvalidInputf("tag name cannot be empty")
	}

	// This is a placeholder that should be overridden by specific implementations
	return errors.NotImplementedf("DeleteTag must be implemented by specific repository implementations")
}

// PutImage adds an image to the repository
func (r *BaseRepository) PutImage(ctx context.Context, img v1.Image, tagName string) error {
	if img == nil {
		return errors.InvalidInputf("image cannot be nil")
	}

	if tagName == "" {
		return errors.InvalidInputf("tag name cannot be empty")
	}

	// This is a placeholder that should be overridden by specific implementations
	return errors.NotImplementedf("PutImage must be implemented by specific repository implementations")
}

// CreateTagReference creates a tag reference for the repository
func (r *BaseRepository) CreateTagReference(tagName string) (name.Tag, error) {
	if tagName == "" {
		return name.Tag{}, errors.InvalidInputf("tag name cannot be empty")
	}

	// Create a tag reference
	tag, err := name.NewTag(r.repository.String() + ":" + tagName)
	if err != nil {
		return name.Tag{}, errors.Wrap(err, "failed to create tag reference")
	}

	return tag, nil
}

// GetRemoteImage retrieves an image using remote options
func (r *BaseRepository) GetRemoteImage(ctx context.Context, ref name.Reference, options ...remote.Option) (v1.Image, error) {
	if ref == nil {
		return nil, errors.InvalidInputf("reference cannot be nil")
	}

	// Get the image
	img, err := remote.Image(ref, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get remote image")
	}

	return img, nil
}

// CacheImage adds an image to the cache
func (r *BaseRepository) CacheImage(tagName string, img v1.Image) {
	r.imagesMutex.Lock()
	r.images[tagName] = img
	r.imagesMutex.Unlock()
}

// ClearCache clears the image cache
func (r *BaseRepository) ClearCache() {
	r.imagesMutex.Lock()
	r.images = make(map[string]v1.Image)
	r.imagesMutex.Unlock()

	r.tagsMutex.Lock()
	r.tags = nil
	r.tagsMutex.Unlock()
}
