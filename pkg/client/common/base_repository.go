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
	logger     log.Logger

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
	Logger     log.Logger
}

// NewBaseRepository creates a new base repository for registry operations
func NewBaseRepository(opts BaseRepositoryOptions) *BaseRepository {
	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
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
	// Check cache first
	r.tagsMutex.RLock()
	if r.tags != nil {
		cachedTags := make([]string, len(r.tags))
		copy(cachedTags, r.tags)
		r.tagsMutex.RUnlock()
		return cachedTags, nil
	}
	r.tagsMutex.RUnlock()

	r.logger.WithFields(map[string]interface{}{
		"repository": r.name,
	}).Debug("Listing tags for repository")

	// List tags using go-containerregistry
	tags, err := remote.List(r.repository)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list tags from repository")
	}

	// Cache the results
	r.tagsMutex.Lock()
	r.tags = make([]string, len(tags))
	copy(r.tags, tags)
	r.tagsMutex.Unlock()

	r.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag_count":  len(tags),
	}).Debug("Successfully listed tags")

	return tags, nil
}

// GetTag returns a tagged image from the repository
func (r *BaseRepository) GetTag(ctx context.Context, tagName string) (v1.Image, error) {
	if tagName == "" {
		return nil, errors.InvalidInputf("tag name cannot be empty")
	}

	// Check the cache first
	var img v1.Image
	var ok bool
	func() {
		r.imagesMutex.RLock()
		defer r.imagesMutex.RUnlock()
		img, ok = r.images[tagName]
	}()

	if ok {
		return img, nil
	}

	// Create tag reference
	tagRef, err := r.CreateTagReference(tagName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tag reference")
	}

	r.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tagName,
	}).Debug("Getting tagged image")

	// Get the image using remote.Image
	img, err = remote.Image(tagRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get image from registry")
	}

	// Cache the image
	r.CacheImage(tagName, img)

	r.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tagName,
	}).Debug("Successfully retrieved tagged image")

	return img, nil
}

// GetImage returns an image by digest
func (r *BaseRepository) GetImage(ctx context.Context, digest string) (v1.Image, error) {
	if digest == "" {
		return nil, errors.InvalidInputf("digest cannot be empty")
	}

	r.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"digest":     digest,
	}).Debug("Getting image by digest")

	// Create digest reference
	digestRef, err := name.NewDigest(r.repository.String() + "@" + digest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digest reference")
	}

	// Get the image using remote.Image
	img, err := remote.Image(digestRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get image by digest from registry")
	}

	r.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"digest":     digest,
	}).Debug("Successfully retrieved image by digest")

	return img, nil
}

// DeleteTag removes a tag from the repository
func (r *BaseRepository) DeleteTag(ctx context.Context, tagName string) error {
	if tagName == "" {
		return errors.InvalidInputf("tag name cannot be empty")
	}

	r.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tagName,
	}).Debug("Deleting tag from repository")

	// Create tag reference
	tagRef, err := r.CreateTagReference(tagName)
	if err != nil {
		return errors.Wrap(err, "failed to create tag reference")
	}

	// Delete the tag using remote.Delete
	err = remote.Delete(tagRef)
	if err != nil {
		return errors.Wrap(err, "failed to delete tag from registry")
	}

	// Remove from cache
	func() {
		r.imagesMutex.Lock()
		defer r.imagesMutex.Unlock()
		delete(r.images, tagName)
	}()

	// Clear tags cache to force refresh
	func() {
		r.tagsMutex.Lock()
		defer r.tagsMutex.Unlock()
		r.tags = nil
	}()

	r.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tagName,
	}).Info("Successfully deleted tag")

	return nil
}

// PutImage adds an image to the repository
func (r *BaseRepository) PutImage(ctx context.Context, img v1.Image, tagName string) error {
	if img == nil {
		return errors.InvalidInputf("image cannot be nil")
	}

	if tagName == "" {
		return errors.InvalidInputf("tag name cannot be empty")
	}

	r.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tagName,
	}).Debug("Putting image to repository")

	// Create tag reference
	tagRef, err := r.CreateTagReference(tagName)
	if err != nil {
		return errors.Wrap(err, "failed to create tag reference")
	}

	// Push the image using remote.Write
	err = remote.Write(tagRef, img)
	if err != nil {
		return errors.Wrap(err, "failed to push image to registry")
	}

	// Cache the image
	r.CacheImage(tagName, img)

	// Clear tags cache to force refresh
	func() {
		r.tagsMutex.Lock()
		defer r.tagsMutex.Unlock()
		r.tags = nil
	}()

	r.logger.WithFields(map[string]interface{}{
		"repository": r.name,
		"tag":        tagName,
	}).Info("Successfully pushed image")

	return nil
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
	func() {
		r.imagesMutex.Lock()
		defer r.imagesMutex.Unlock()
		r.images[tagName] = img
	}()
}

// ClearCache clears the image cache
func (r *BaseRepository) ClearCache() {
	func() {
		r.imagesMutex.Lock()
		defer r.imagesMutex.Unlock()
		r.images = make(map[string]v1.Image)
	}()

	func() {
		r.tagsMutex.Lock()
		defer r.tagsMutex.Unlock()
		r.tags = nil
	}()
}
