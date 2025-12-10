package sync

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// SemverFilter filters tags based on semantic versioning constraints
type SemverFilter struct {
	constraint *semver.Constraints
}

// NewSemverFilter creates a new semver filter from a constraint string
// Constraint examples:
//   - ">=1.2.3" - Greater than or equal to 1.2.3
//   - "^2.0.0" - Compatible with 2.0.0 (>=2.0.0, <3.0.0)
//   - "~1.2.3" - Approximately 1.2.3 (>=1.2.3, <1.3.0)
//   - "1.2.x" - Any patch version of 1.2
//   - ">=1.0.0 <2.0.0" - Range constraint
func NewSemverFilter(constraintStr string) (*SemverFilter, error) {
	constraint, err := semver.NewConstraint(constraintStr)
	if err != nil {
		return nil, fmt.Errorf("invalid semver constraint '%s': %w", constraintStr, err)
	}

	return &SemverFilter{
		constraint: constraint,
	}, nil
}

// Filter filters tags based on semver constraint
func (f *SemverFilter) Filter(tags []string) []string {
	var filtered []string

	for _, tag := range tags {
		// Try to parse as semver
		v, err := f.parseVersion(tag)
		if err != nil {
			// Not a valid semver tag, skip
			continue
		}

		// Check if version matches constraint
		if f.constraint.Check(v) {
			filtered = append(filtered, tag)
		}
	}

	return filtered
}

// FilterAndSort filters tags and sorts them by semantic version (descending)
func (f *SemverFilter) FilterAndSort(tags []string) []string {
	filtered := f.Filter(tags)

	// Parse versions for sorting
	type versionTag struct {
		tag     string
		version *semver.Version
	}

	var versions []versionTag
	for _, tag := range filtered {
		v, err := f.parseVersion(tag)
		if err != nil {
			continue
		}
		versions = append(versions, versionTag{tag: tag, version: v})
	}

	// Sort by version (descending - newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].version.GreaterThan(versions[j].version)
	})

	// Extract sorted tags
	sorted := make([]string, len(versions))
	for i, vt := range versions {
		sorted[i] = vt.tag
	}

	return sorted
}

// parseVersion parses a tag as a semantic version
// Handles common tag formats:
//   - "v1.2.3" -> 1.2.3
//   - "1.2.3" -> 1.2.3
//   - "v1.2.3-alpha" -> 1.2.3-alpha
//   - "release-1.2.3" -> 1.2.3
func (f *SemverFilter) parseVersion(tag string) (*semver.Version, error) {
	// Try direct parsing first
	v, err := semver.NewVersion(tag)
	if err == nil {
		return v, nil
	}

	// Strip common prefixes
	cleaned := tag
	tagLower := strings.ToLower(tag)
	for _, prefix := range []string{"v", "release-", "version-", "ver-"} {
		if strings.HasPrefix(tagLower, prefix) {
			cleaned = tag[len(prefix):]
			break
		}
	}

	// Try parsing cleaned version
	v, err = semver.NewVersion(cleaned)
	if err != nil {
		return nil, fmt.Errorf("not a valid semver: %s", tag)
	}

	return v, nil
}

// GetLatestVersion returns the latest version from a list of tags
func (f *SemverFilter) GetLatestVersion(tags []string) string {
	sorted := f.FilterAndSort(tags)
	if len(sorted) == 0 {
		return ""
	}
	return sorted[0]
}

// GetLatestN returns the N latest versions from a list of tags
func (f *SemverFilter) GetLatestN(tags []string, n int) []string {
	sorted := f.FilterAndSort(tags)
	if len(sorted) <= n {
		return sorted
	}
	return sorted[:n]
}

// ValidateSemverTags validates that tags are valid semantic versions
func ValidateSemverTags(tags []string) ([]string, []string) {
	var valid []string
	var invalid []string

	for _, tag := range tags {
		// Try common parsing patterns
		cleaned := tag
		for _, prefix := range []string{"v", "V", "release-", "version-", "ver-"} {
			if strings.HasPrefix(strings.ToLower(tag), strings.ToLower(prefix)) {
				cleaned = tag[len(prefix):]
				break
			}
		}

		_, err := semver.NewVersion(cleaned)
		if err == nil {
			valid = append(valid, tag)
		} else {
			invalid = append(invalid, tag)
		}
	}

	return valid, invalid
}

// SortSemverTags sorts tags by semantic version (descending)
func SortSemverTags(tags []string) []string {
	type versionTag struct {
		tag     string
		version *semver.Version
	}

	var versions []versionTag
	for _, tag := range tags {
		// Parse with common prefix stripping
		cleaned := tag
		for _, prefix := range []string{"v", "V", "release-", "version-", "ver-"} {
			if strings.HasPrefix(strings.ToLower(tag), strings.ToLower(prefix)) {
				cleaned = tag[len(prefix):]
				break
			}
		}

		v, err := semver.NewVersion(cleaned)
		if err != nil {
			continue
		}
		versions = append(versions, versionTag{tag: tag, version: v})
	}

	// Sort by version (descending)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].version.GreaterThan(versions[j].version)
	})

	// Extract sorted tags
	sorted := make([]string, len(versions))
	for i, vt := range versions {
		sorted[i] = vt.tag
	}

	return sorted
}

// GetMajorVersions groups tags by major version
func GetMajorVersions(tags []string) map[uint64][]string {
	groups := make(map[uint64][]string)

	for _, tag := range tags {
		// Parse with common prefix stripping
		cleaned := tag
		for _, prefix := range []string{"v", "V", "release-", "version-", "ver-"} {
			if strings.HasPrefix(strings.ToLower(tag), strings.ToLower(prefix)) {
				cleaned = tag[len(prefix):]
				break
			}
		}

		v, err := semver.NewVersion(cleaned)
		if err != nil {
			continue
		}

		major := v.Major()
		groups[major] = append(groups[major], tag)
	}

	return groups
}

// FilterByMajorVersion filters tags by major version
func FilterByMajorVersion(tags []string, major uint64) []string {
	var filtered []string

	for _, tag := range tags {
		// Parse with common prefix stripping
		cleaned := tag
		for _, prefix := range []string{"v", "V", "release-", "version-", "ver-"} {
			if strings.HasPrefix(strings.ToLower(tag), strings.ToLower(prefix)) {
				cleaned = tag[len(prefix):]
				break
			}
		}

		v, err := semver.NewVersion(cleaned)
		if err != nil {
			continue
		}

		if v.Major() == major {
			filtered = append(filtered, tag)
		}
	}

	return filtered
}
