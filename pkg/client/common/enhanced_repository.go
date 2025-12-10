package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// EnhancedRepositoryOptions provides options for creating an enhanced repository
type EnhancedRepositoryOptions struct {
	Name       string
	Repository name.Repository
	Logger     log.Logger
	Client     interface{}

	// Cache configuration
	CacheSize          int
	CacheExpiration    time.Duration
	EnableSummaryCache bool
}

// ImageSummary contains the basic metadata for an image
type ImageSummary struct {
	Digest     string    `json:"digest"`
	Size       int64     `json:"size"`
	MediaType  string    `json:"mediaType"`
	Tags       []string  `json:"tags,omitempty"`
	Created    time.Time `json:"created,omitempty"`
	Layers     int       `json:"layers,omitempty"`
	Compressed bool      `json:"compressed,omitempty"`
}

// EnhancedRepository extends the base repository with additional functionality
type EnhancedRepository struct {
	*BaseRepository

	client          interface{} // This can be cast to the specific client type
	cacheExpiration time.Duration

	// Summary cache
	imageSummaryCache     map[string]ImageSummary
	enableSummaryCache    bool
	summaryCacheTimestamp time.Time
}

// NewEnhancedRepository creates a new enhanced repository
func NewEnhancedRepository(opts EnhancedRepositoryOptions) *EnhancedRepository {
	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Create base repository
	baseRepo := NewBaseRepository(BaseRepositoryOptions{
		Name:       opts.Name,
		Repository: opts.Repository,
		Logger:     opts.Logger,
	})

	// Set default cache expiration
	if opts.CacheExpiration == 0 {
		opts.CacheExpiration = 5 * time.Minute
	}

	// Create enhanced repository
	return &EnhancedRepository{
		BaseRepository:     baseRepo,
		client:             opts.Client,
		cacheExpiration:    opts.CacheExpiration,
		imageSummaryCache:  make(map[string]ImageSummary),
		enableSummaryCache: opts.EnableSummaryCache,
	}
}

// RefreshTags refreshes the tag list
func (r *EnhancedRepository) RefreshTags(ctx context.Context) error {
	r.logger.WithField("repository", r.GetName()).Debug("Refreshing tags for repository")

	// Clear the tags cache to force a fresh fetch
	func() {
		r.tagsMutex.Lock()
		defer r.tagsMutex.Unlock()
		r.tags = nil
	}()

	// Fetch fresh tags using the base implementation
	tags, err := r.ListTags(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh tags")
	}

	r.logger.WithFields(map[string]interface{}{
		"repository": r.GetName(),
		"tag_count":  len(tags),
	}).Info("Successfully refreshed tags")

	return nil
}

// RefreshImages refreshes the image cache
func (r *EnhancedRepository) RefreshImages(ctx context.Context, tags []string) error {
	if len(tags) == 0 {
		// If no tags specified, get all tags
		allTags, err := r.ListTags(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to list tags for refresh")
		}
		tags = allTags
	}

	r.logger.WithFields(map[string]interface{}{
		"repository": r.GetName(),
		"tag_count":  len(tags),
	}).Debug("Refreshing image cache")

	// Clear existing image cache
	r.ClearCache()

	// Pre-fetch images for the specified tags
	for i, tag := range tags {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		r.logger.WithFields(map[string]interface{}{
			"repository": r.GetName(),
			"tag":        tag,
			"progress":   fmt.Sprintf("%d/%d", i+1, len(tags)),
		}).Debug("Refreshing image")

		// Fetch the image (this will cache it)
		_, err := r.GetTag(ctx, tag)
		if err != nil {
			r.logger.WithFields(map[string]interface{}{
				"repository": r.GetName(),
				"tag":        tag,
				"error":      err.Error(),
			}).Warn("Failed to refresh image")
			// Continue with other images rather than failing completely
			continue
		}
	}

	r.logger.WithFields(map[string]interface{}{
		"repository": r.GetName(),
		"tag_count":  len(tags),
	}).Info("Successfully refreshed image cache")

	return nil
}

// GetImageSummary returns a summary of image information
func (r *EnhancedRepository) GetImageSummary(ctx context.Context, reference string) (ImageSummary, error) {
	// Check cache first if enabled
	if r.enableSummaryCache {
		if summary, ok := r.imageSummaryCache[reference]; ok {
			// Check if cache is fresh
			if time.Since(r.summaryCacheTimestamp) < r.cacheExpiration {
				return summary, nil
			}
		}
	}

	// Create tag reference
	_, err := r.CreateTagReference(reference)
	if err != nil {
		return ImageSummary{}, err
	}

	// Get the image
	img, err := r.GetTag(ctx, reference)
	if err != nil {
		return ImageSummary{}, err
	}

	// Get image digest
	digest, err := img.Digest()
	if err != nil {
		return ImageSummary{}, errors.Wrap(err, "failed to get image digest")
	}

	// Get manifest and config
	manifest, err := img.Manifest()
	if err != nil {
		return ImageSummary{}, errors.Wrap(err, "failed to get image manifest")
	}

	configFile, err := img.ConfigFile()
	if err != nil {
		r.logger.WithFields(map[string]interface{}{
			"tag":   reference,
			"error": err.Error(),
		}).Warn("Failed to get image config file")
	}

	// Calculate size
	var size int64
	for _, layer := range manifest.Layers {
		size += layer.Size
	}

	// Create summary
	summary := ImageSummary{
		Digest:    digest.String(),
		Size:      size,
		MediaType: string(manifest.MediaType),
		Layers:    len(manifest.Layers),
	}

	// Add creation time if available
	if configFile != nil && !configFile.Created.IsZero() {
		summary.Created = configFile.Created.Time
	}

	// Store in cache if enabled
	if r.enableSummaryCache {
		r.imageSummaryCache[reference] = summary
		r.summaryCacheTimestamp = time.Now()
	}

	return summary, nil
}

// ListImageManifests lists all manifests in the repository
func (r *EnhancedRepository) ListImageManifests(ctx context.Context) ([]string, error) {
	r.logger.WithField("repository", r.GetName()).Debug("Listing image manifests for repository")

	// Get all tags first
	tags, err := r.ListTags(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list tags")
	}

	manifests := make([]string, 0, len(tags))

	// For each tag, get the image and extract manifest digest
	for _, tag := range tags {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Get the image for this tag
		img, err := r.GetTag(ctx, tag)
		if err != nil {
			r.logger.WithFields(map[string]interface{}{
				"repository": r.GetName(),
				"tag":        tag,
				"error":      err.Error(),
			}).Warn("Failed to get image for manifest listing")
			continue
		}

		// Get the manifest digest
		digest, err := img.Digest()
		if err != nil {
			r.logger.WithFields(map[string]interface{}{
				"repository": r.GetName(),
				"tag":        tag,
				"error":      err.Error(),
			}).Warn("Failed to get digest for manifest listing")
			continue
		}

		manifestDigest := digest.String()

		// Check if we already have this manifest (avoid duplicates)
		found := false
		for _, existing := range manifests {
			if existing == manifestDigest {
				found = true
				break
			}
		}

		if !found {
			manifests = append(manifests, manifestDigest)
		}
	}

	r.logger.WithFields(map[string]interface{}{
		"repository":     r.GetName(),
		"tag_count":      len(tags),
		"manifest_count": len(manifests),
	}).Debug("Successfully listed image manifests")

	return manifests, nil
}

// CopyTag copies a tag from another repository
// CopyTag copies a tag from another repository
func (r *EnhancedRepository) CopyTag(ctx context.Context, sourceRepo interface{}, sourceTag, destTag string) error {
	// Convert sourceRepo to the expected interface
	sourceRepository, ok := sourceRepo.(interface {
		GetTag(ctx context.Context, tag string) (v1.Image, error)
	})
	if !ok {
		return errors.InvalidInputf("source repository is not compatible")
	}
	// Get the source image
	sourceImage, err := sourceRepository.GetTag(ctx, sourceTag)
	if err != nil {
		return errors.Wrapf(err, "failed to get source image: %s", sourceTag)
	}

	// Put the image in this repository with the new tag
	return r.PutImage(ctx, sourceImage, destTag)
}

// ExportImage exports an image to a writer
func (r *EnhancedRepository) ExportImage(ctx context.Context, tagName string, format string, writer io.Writer) error {
	// Get the image
	img, err := r.GetTag(ctx, tagName)
	if err != nil {
		return errors.Wrapf(err, "failed to get image: %s", tagName)
	}

	switch format {
	case "json":
		// Export manifest and config as JSON
		manifest, err := img.Manifest()
		if err != nil {
			return errors.Wrap(err, "failed to get image manifest")
		}

		config, err := img.ConfigFile()
		if err != nil {
			return errors.Wrap(err, "failed to get image config")
		}

		exportData := struct {
			Manifest *v1.Manifest   `json:"manifest"`
			Config   *v1.ConfigFile `json:"config"`
		}{
			Manifest: manifest,
			Config:   config,
		}

		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(exportData)

	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// CompareTags compares two tags and returns the differences
func (r *EnhancedRepository) CompareTags(ctx context.Context, tag1, tag2 string) (map[string]interface{}, error) {
	// Get both images
	img1, err := r.GetTag(ctx, tag1)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get first image: %s", tag1)
	}

	img2, err := r.GetTag(ctx, tag2)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get second image: %s", tag2)
	}

	// Get digests
	digest1, err := img1.Digest()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get first image digest")
	}

	digest2, err := img2.Digest()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get second image digest")
	}

	// Quick check: same digest means same image
	if digest1.String() == digest2.String() {
		return map[string]interface{}{
			"identical": true,
			"digest":    digest1.String(),
		}, nil
	}

	// Get manifests
	manifest1, err := img1.Manifest()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get first image manifest")
	}

	manifest2, err := img2.Manifest()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get second image manifest")
	}

	// Compare layer counts
	layerCount1 := len(manifest1.Layers)
	layerCount2 := len(manifest2.Layers)

	// Compare configs
	config1, _ := img1.ConfigFile()
	config2, _ := img2.ConfigFile()

	// Build the diff result
	diff := map[string]interface{}{
		"identical":    false,
		"digest1":      digest1.String(),
		"digest2":      digest2.String(),
		"mediaType1":   manifest1.MediaType,
		"mediaType2":   manifest2.MediaType,
		"layerCount1":  layerCount1,
		"layerCount2":  layerCount2,
		"size1":        calculateSize(manifest1.Layers),
		"size2":        calculateSize(manifest2.Layers),
		"layersMatch":  compareLayerDigests(manifest1.Layers, manifest2.Layers),
		"configsMatch": config1 != nil && config2 != nil && compareConfigs(config1, config2),
	}

	return diff, nil
}

// Helper functions

// calculateSize calculates the total size of layers
func calculateSize(layers []v1.Descriptor) int64 {
	var size int64
	for _, layer := range layers {
		size += layer.Size
	}
	return size
}

// compareLayerDigests compares layer digests between two images
func compareLayerDigests(layers1, layers2 []v1.Descriptor) bool {
	if len(layers1) != len(layers2) {
		return false
	}

	for i, layer1 := range layers1 {
		if layer1.Digest.String() != layers2[i].Digest.String() {
			return false
		}
	}

	return true
}

// compareConfigs compares image configs
func compareConfigs(config1, config2 *v1.ConfigFile) bool {
	// This is a simple comparison - in reality, you might want to compare
	// specific fields like environment variables, command, etc.
	return config1.Architecture == config2.Architecture &&
		config1.OS == config2.OS &&
		config1.OSVersion == config2.OSVersion
}
