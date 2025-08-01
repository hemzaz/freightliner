package common

import (
	"context"
	"io"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// CompositeRepository demonstrates how to implement the new segregated interfaces
// using composition patterns. This shows how the interface architecture improvements
// enable better separation of concerns and testability.
type CompositeRepository struct {
	*BaseRepository
	logger log.Logger

	// Composed behaviors - each can be implemented independently
	reader           interfaces.Reader
	writer           interfaces.Writer
	imageProvider    interfaces.ImageProvider
	metadataProvider interfaces.MetadataProvider
	contentProvider  interfaces.ContentProvider
	contentManager   interfaces.ContentManager
}

// CompositeRepositoryOptions provides options for creating a composite repository
type CompositeRepositoryOptions struct {
	BaseRepository  *BaseRepository
	Logger          log.Logger
	EnableCaching   bool
	EnableStreaming bool
	Timeout         time.Duration
}

// NewCompositeRepository creates a new composite repository with segregated interface implementations
func NewCompositeRepository(opts CompositeRepositoryOptions) *CompositeRepository {
	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	repo := &CompositeRepository{
		BaseRepository: opts.BaseRepository,
		logger:         opts.Logger,
	}

	// Initialize composed behaviors
	repo.reader = &repositoryReader{base: repo}
	repo.writer = &repositoryWriter{base: repo}
	repo.imageProvider = &repositoryImageProvider{base: repo}
	repo.metadataProvider = &repositoryMetadataProvider{base: repo}
	repo.contentProvider = &repositoryContentProvider{base: repo}
	repo.contentManager = &repositoryContentManager{base: repo}

	return repo
}

// ===== COMPOSITION INTERFACE IMPLEMENTATIONS =====

// AsReader returns a read-only view of the repository
func (r *CompositeRepository) AsReader() interfaces.Reader {
	return r.reader
}

// AsWriter returns a write-only view of the repository
func (r *CompositeRepository) AsWriter() interfaces.Writer {
	return r.writer
}

// AsImageProvider returns an image provider view
func (r *CompositeRepository) AsImageProvider() interfaces.ImageProvider {
	return r.imageProvider
}

// AsMetadataProvider returns a metadata provider view
func (r *CompositeRepository) AsMetadataProvider() interfaces.MetadataProvider {
	return r.metadataProvider
}

// AsContentProvider returns a content provider view
func (r *CompositeRepository) AsContentProvider() interfaces.ContentProvider {
	return r.contentProvider
}

// AsContentManager returns a content manager view
func (r *CompositeRepository) AsContentManager() interfaces.ContentManager {
	return r.contentManager
}

// ===== SEGREGATED INTERFACE IMPLEMENTATIONS =====

// repositoryReader implements the Reader interface
type repositoryReader struct {
	base *CompositeRepository
}

func (r *repositoryReader) GetName() string {
	return r.base.GetName()
}

func (r *repositoryReader) GetRepositoryName() string {
	return r.base.GetName()
}

func (r *repositoryReader) ListTags(ctx context.Context) ([]string, error) {
	return r.base.ListTags(ctx)
}

func (r *repositoryReader) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	// Implement manifest retrieval with proper context handling
	r.base.logger.WithFields(map[string]interface{}{
		"repository": r.base.GetName(),
		"tag":        tag,
	}).Debug("Getting manifest")

	// Use context timeout for the operation
	if deadline, ok := ctx.Deadline(); ok {
		r.base.logger.WithFields(map[string]interface{}{
			"deadline": deadline,
			"timeout":  time.Until(deadline),
		}).Debug("Operation has timeout")
	}

	// Implementation would call go-containerregistry APIs
	// This is a placeholder for the actual implementation
	return &interfaces.Manifest{
		Content:   []byte("placeholder-manifest"),
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:    "sha256:placeholder",
	}, nil
}

func (r *repositoryReader) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	// Implement layer reader with proper context handling
	r.base.logger.WithFields(map[string]interface{}{
		"repository": r.base.GetName(),
		"digest":     digest,
	}).Debug("Getting layer reader")

	// Implementation would return actual layer reader
	// This is a placeholder
	return io.NopCloser(io.LimitReader(nil, 0)), nil
}

// repositoryWriter implements the Writer interface
type repositoryWriter struct {
	base *CompositeRepository
}

func (r *repositoryWriter) GetName() string {
	return r.base.GetName()
}

func (r *repositoryWriter) GetRepositoryName() string {
	return r.base.GetName()
}

func (r *repositoryWriter) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	// Delegate to the reader capability
	return r.base.reader.GetManifest(ctx, tag)
}

func (r *repositoryWriter) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	r.base.logger.WithFields(map[string]interface{}{
		"repository": r.base.GetName(),
		"tag":        tag,
		"digest":     manifest.Digest,
	}).Info("Putting manifest")

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Implementation would upload manifest
	return nil
}

func (r *repositoryWriter) DeleteManifest(ctx context.Context, tag string) error {
	r.base.logger.WithFields(map[string]interface{}{
		"repository": r.base.GetName(),
		"tag":        tag,
	}).Info("Deleting manifest")

	// Implementation would delete manifest
	return nil
}

// repositoryImageProvider implements the ImageProvider interface
type repositoryImageProvider struct {
	base *CompositeRepository
}

func (r *repositoryImageProvider) GetName() string {
	return r.base.GetName()
}

func (r *repositoryImageProvider) GetRepositoryName() string {
	return r.base.GetName()
}

func (r *repositoryImageProvider) GetImageReference(tag string) (name.Reference, error) {
	// Create image reference
	return name.ParseReference(r.base.GetURI() + ":" + tag)
}

func (r *repositoryImageProvider) GetRemoteOptions() ([]remote.Option, error) {
	// Return remote options for accessing this repository
	return []remote.Option{}, nil
}

func (r *repositoryImageProvider) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	r.base.logger.WithFields(map[string]interface{}{
		"repository": r.base.GetName(),
		"tag":        tag,
	}).Debug("Getting image")

	// Implementation would return actual image
	// This is a placeholder
	return nil, errors.NotImplementedf("GetImage not yet implemented")
}

// repositoryMetadataProvider implements the MetadataProvider interface
type repositoryMetadataProvider struct {
	base *CompositeRepository
}

func (r *repositoryMetadataProvider) GetName() string {
	return r.base.GetName()
}

func (r *repositoryMetadataProvider) GetRepositoryName() string {
	return r.base.GetName()
}

func (r *repositoryMetadataProvider) ListTags(ctx context.Context) ([]string, error) {
	return r.base.ListTags(ctx)
}

// repositoryContentProvider implements the ContentProvider interface
type repositoryContentProvider struct {
	base *CompositeRepository
}

func (r *repositoryContentProvider) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	// Delegate to reader implementation
	reader := &repositoryReader{base: r.base}
	return reader.GetManifest(ctx, tag)
}

func (r *repositoryContentProvider) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	// Delegate to reader implementation
	reader := &repositoryReader{base: r.base}
	return reader.GetLayerReader(ctx, digest)
}

// repositoryContentManager implements the ContentManager interface
type repositoryContentManager struct {
	base *CompositeRepository
}

func (r *repositoryContentManager) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	reader := &repositoryReader{base: r.base}
	return reader.GetManifest(ctx, tag)
}

func (r *repositoryContentManager) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	writer := &repositoryWriter{base: r.base}
	return writer.PutManifest(ctx, tag, manifest)
}

func (r *repositoryContentManager) DeleteManifest(ctx context.Context, tag string) error {
	writer := &repositoryWriter{base: r.base}
	return writer.DeleteManifest(ctx, tag)
}

func (r *repositoryContentManager) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	reader := &repositoryReader{base: r.base}
	return reader.GetLayerReader(ctx, digest)
}

// ===== INTERFACE COMPLIANCE VALIDATION =====

// Compile-time interface compliance checks
var (
	_ interfaces.RepositoryComposer = (*CompositeRepository)(nil)
	_ interfaces.Reader             = (*repositoryReader)(nil)
	_ interfaces.Writer             = (*repositoryWriter)(nil)
	_ interfaces.ImageProvider      = (*repositoryImageProvider)(nil)
	_ interfaces.MetadataProvider   = (*repositoryMetadataProvider)(nil)
	_ interfaces.ContentProvider    = (*repositoryContentProvider)(nil)
	_ interfaces.ContentManager     = (*repositoryContentManager)(nil)

	// Composition interface compliance
	_ interfaces.ReadWriteRepository = (interfaces.ReadWriteRepository)(nil)
	_ interfaces.ImageRepository     = (interfaces.ImageRepository)(nil)
	_ interfaces.FullRepository      = (interfaces.FullRepository)(nil)
)
