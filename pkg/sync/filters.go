package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"time"

	"freightliner/pkg/manifest"
)

// TagFilter provides tag filtering capabilities
type TagFilter struct {
	// Specific tags to match
	tags map[string]bool

	// Regex pattern for tag matching
	regex *regexp.Regexp

	// Semver filter for semantic versioning
	semver *SemverFilter

	// AllTags matches all tags
	allTags bool

	// LatestN returns only the N most recent tags
	latestN int
}

// NewTagFilter creates a new tag filter from ImageSync configuration
func NewTagFilter(img ImageSync) (*TagFilter, error) {
	filter := &TagFilter{}

	// Specific tags
	if len(img.Tags) > 0 {
		filter.tags = make(map[string]bool)
		for _, tag := range img.Tags {
			filter.tags[tag] = true
		}
		return filter, nil
	}

	// Regex pattern
	if img.TagRegex != "" {
		regex, err := regexp.Compile(img.TagRegex)
		if err != nil {
			return nil, fmt.Errorf("invalid tag regex '%s': %w", img.TagRegex, err)
		}
		filter.regex = regex
		return filter, nil
	}

	// Semver constraint
	if img.SemverConstraint != "" {
		semver, err := NewSemverFilter(img.SemverConstraint)
		if err != nil {
			return nil, fmt.Errorf("invalid semver constraint: %w", err)
		}
		filter.semver = semver
		return filter, nil
	}

	// All tags
	if img.AllTags {
		filter.allTags = true
		return filter, nil
	}

	// Latest N tags
	if img.LatestN > 0 {
		filter.latestN = img.LatestN
		return filter, nil
	}

	return nil, fmt.Errorf("no filter criteria specified")
}

// Filter filters tags based on configured criteria
func (f *TagFilter) Filter(tags []string) []string {
	// Specific tags
	if f.tags != nil {
		var filtered []string
		for _, tag := range tags {
			if f.tags[tag] {
				filtered = append(filtered, tag)
			}
		}
		return filtered
	}

	// Regex pattern
	if f.regex != nil {
		var filtered []string
		for _, tag := range tags {
			if f.regex.MatchString(tag) {
				filtered = append(filtered, tag)
			}
		}
		return filtered
	}

	// Semver constraint
	if f.semver != nil {
		return f.semver.FilterAndSort(tags)
	}

	// All tags
	if f.allTags {
		return tags
	}

	// Latest N (requires metadata - for now just return first N)
	if f.latestN > 0 {
		if len(tags) <= f.latestN {
			return tags
		}
		return tags[:f.latestN]
	}

	return nil
}

// FilterWithMetadata filters tags with creation time metadata
func (f *TagFilter) FilterWithMetadata(tags []TagMetadata) []string {
	// Latest N with proper sorting by creation time
	if f.latestN > 0 {
		// Sort by creation time (descending)
		sorted := make([]TagMetadata, len(tags))
		copy(sorted, tags)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
		})

		// Take first N
		limit := f.latestN
		if len(sorted) < limit {
			limit = len(sorted)
		}

		result := make([]string, limit)
		for i := 0; i < limit; i++ {
			result[i] = sorted[i].Tag
		}
		return result
	}

	// For other filters, extract tag names and use regular Filter
	tagNames := make([]string, len(tags))
	for i, tm := range tags {
		tagNames[i] = tm.Tag
	}
	return f.Filter(tagNames)
}

// TagMetadata contains metadata about a tag
type TagMetadata struct {
	Tag         string
	Digest      string
	CreatedAt   time.Time
	Size        int64
	Platform    string
	Annotations map[string]string
}

// ApplyLimit applies a limit to filtered tags
func ApplyLimit(tags []string, limit int) []string {
	if limit <= 0 || len(tags) <= limit {
		return tags
	}
	return tags[:limit]
}

// ArchitectureFilterer provides architecture filtering capabilities
type ArchitectureFilterer interface {
	// GetManifest returns the manifest for a given repository and tag
	GetManifest(ctx context.Context, repository, tag string) ([]byte, string, error)

	// GetConfigBlob fetches a config blob by digest from the given repository
	// This is needed to determine architecture for single-arch manifests
	GetConfigBlob(ctx context.Context, repository, digest string) ([]byte, error)
}

// ApplyArchitectureFilter filters tags by architecture using manifest inspection
// This implementation queries the manifest for each tag to determine its architecture(s)
func ApplyArchitectureFilter(ctx context.Context, filterer ArchitectureFilterer, repository string, tags []string, architectures []string) ([]string, error) {
	if len(architectures) == 0 {
		return tags, nil
	}

	// Create a map for fast architecture lookup
	archMap := make(map[string]bool)
	for _, arch := range architectures {
		archMap[arch] = true
	}

	var filtered []string
	for _, tag := range tags {
		// Get manifest for this tag
		manifestData, mediaType, err := filterer.GetManifest(ctx, repository, tag)
		if err != nil {
			// Skip tags we can't fetch manifests for
			continue
		}

		// Check if this tag matches any desired architecture
		if hasMatchingArchitecture(ctx, filterer, repository, manifestData, mediaType, archMap) {
			filtered = append(filtered, tag)
		}
	}

	return filtered, nil
}

// hasMatchingArchitecture checks if a manifest contains any of the specified architectures
func hasMatchingArchitecture(ctx context.Context, filterer ArchitectureFilterer, repository string, manifestData []byte, mediaType string, desiredArchs map[string]bool) bool {
	switch mediaType {
	case "application/vnd.oci.image.index.v1+json",
		"application/vnd.docker.distribution.manifest.list.v2+json":
		// Multi-arch manifest (manifest list or OCI index)
		return checkMultiArchManifest(manifestData, desiredArchs)

	case "application/vnd.oci.image.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json":
		// Single-arch manifest - need to inspect config blob
		return checkSingleArchManifest(ctx, filterer, repository, manifestData, desiredArchs)

	default:
		// Unknown media type - assume it doesn't match
		return false
	}
}

// checkMultiArchManifest checks if a multi-arch manifest contains desired architectures
func checkMultiArchManifest(manifestData []byte, desiredArchs map[string]bool) bool {
	// Try to parse as OCI Image Index first
	var ociIndex manifest.OCIImageIndex
	if err := json.Unmarshal(manifestData, &ociIndex); err == nil {
		for _, desc := range ociIndex.Manifests {
			if desc.Platform != nil && desiredArchs[desc.Platform.Architecture] {
				return true
			}
		}
		return false
	}

	// Try to parse as Docker Manifest List
	var dockerList manifest.DockerManifestList
	if err := json.Unmarshal(manifestData, &dockerList); err == nil {
		for _, desc := range dockerList.Manifests {
			if desc.Platform != nil && desiredArchs[desc.Platform.Architecture] {
				return true
			}
		}
		return false
	}

	// Failed to parse as multi-arch manifest
	return false
}

// checkSingleArchManifest checks if a single-arch manifest matches desired architectures
// by fetching and parsing the config blob
func checkSingleArchManifest(ctx context.Context, filterer ArchitectureFilterer, repository string, manifestData []byte, desiredArchs map[string]bool) bool {
	// Try parsing as OCI Manifest
	var ociManifest manifest.OCIManifest
	if err := json.Unmarshal(manifestData, &ociManifest); err == nil {
		// Check if config descriptor has platform info (optional in OCI spec)
		if ociManifest.Config.Platform != nil {
			return desiredArchs[ociManifest.Config.Platform.Architecture]
		}

		// Fetch config blob to determine architecture
		if arch := fetchArchFromConfig(ctx, filterer, repository, ociManifest.Config.Digest); arch != "" {
			return desiredArchs[arch]
		}

		// Failed to determine architecture - be conservative and include it
		return true
	}

	// Try parsing as Docker V2 Manifest
	var dockerManifest manifest.DockerV2Schema2Manifest
	if err := json.Unmarshal(manifestData, &dockerManifest); err == nil {
		// Docker V2 manifests don't include architecture in manifest
		// Fetch config blob to determine architecture
		if arch := fetchArchFromConfig(ctx, filterer, repository, dockerManifest.Config.Digest); arch != "" {
			return desiredArchs[arch]
		}

		// Failed to determine architecture - be conservative and include it
		return true
	}

	// Unknown format - be conservative and include it
	return true
}

// fetchArchFromConfig fetches a config blob and extracts the architecture
func fetchArchFromConfig(ctx context.Context, filterer ArchitectureFilterer, repository, digest string) string {
	// Fetch the config blob
	configBlob, err := filterer.GetConfigBlob(ctx, repository, digest)
	if err != nil {
		// Failed to fetch config - can't determine architecture
		return ""
	}

	// Parse the config blob to extract architecture
	// Both OCI and Docker V2 configs have "architecture" field at root level
	var config struct {
		Architecture string `json:"architecture"`
	}

	if err := json.Unmarshal(configBlob, &config); err != nil {
		// Failed to parse config
		return ""
	}

	return config.Architecture
}

// CombineFilters combines multiple tag filters with AND logic
func CombineFilters(tags []string, filters ...*TagFilter) []string {
	result := tags
	for _, filter := range filters {
		if filter != nil {
			result = filter.Filter(result)
		}
	}
	return result
}

// ExcludeTags excludes specific tags from the list
func ExcludeTags(tags []string, exclude []string) []string {
	if len(exclude) == 0 {
		return tags
	}

	excludeMap := make(map[string]bool)
	for _, tag := range exclude {
		excludeMap[tag] = true
	}

	var filtered []string
	for _, tag := range tags {
		if !excludeMap[tag] {
			filtered = append(filtered, tag)
		}
	}

	return filtered
}

// FilterByPrefix filters tags by prefix
func FilterByPrefix(tags []string, prefix string) []string {
	if prefix == "" {
		return tags
	}

	var filtered []string
	for _, tag := range tags {
		if len(tag) >= len(prefix) && tag[:len(prefix)] == prefix {
			filtered = append(filtered, tag)
		}
	}

	return filtered
}

// FilterBySuffix filters tags by suffix
func FilterBySuffix(tags []string, suffix string) []string {
	if suffix == "" {
		return tags
	}

	var filtered []string
	for _, tag := range tags {
		if len(tag) >= len(suffix) && tag[len(tag)-len(suffix):] == suffix {
			filtered = append(filtered, tag)
		}
	}

	return filtered
}

// FilterByPattern filters tags by glob-like pattern (* wildcard)
func FilterByPattern(tags []string, pattern string) ([]string, error) {
	// Convert glob pattern to regex
	regexPattern := "^" + regexp.QuoteMeta(pattern) + "$"
	regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
	regexPattern = regexp.MustCompile(`\\\?`).ReplaceAllString(regexPattern, ".")

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern '%s': %w", pattern, err)
	}

	var filtered []string
	for _, tag := range tags {
		if regex.MatchString(tag) {
			filtered = append(filtered, tag)
		}
	}

	return filtered, nil
}

// DeduplicateTags removes duplicate tags while preserving order
func DeduplicateTags(tags []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	return result
}
