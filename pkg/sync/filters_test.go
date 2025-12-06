package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTagFilter(t *testing.T) {
	tests := []struct {
		name        string
		image       ImageSync
		expectError bool
		filterType  string
	}{
		{
			name:       "specific tags",
			image:      ImageSync{Repository: "nginx", Tags: []string{"latest", "1.21"}},
			filterType: "tags",
		},
		{
			name:       "regex",
			image:      ImageSync{Repository: "nginx", TagRegex: "^1\\..*"},
			filterType: "regex",
		},
		{
			name:       "semver",
			image:      ImageSync{Repository: "nginx", SemverConstraint: ">=1.20.0"},
			filterType: "semver",
		},
		{
			name:       "all tags",
			image:      ImageSync{Repository: "nginx", AllTags: true},
			filterType: "all",
		},
		{
			name:       "latest N",
			image:      ImageSync{Repository: "nginx", LatestN: 5},
			filterType: "latest_n",
		},
		{
			name:        "no filter",
			image:       ImageSync{Repository: "nginx"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewTagFilter(tt.image)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, filter)
			}
		})
	}
}

func TestTagFilter_Filter_SpecificTags(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		Tags:       []string{"latest", "1.21", "1.22"},
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	availableTags := []string{"latest", "1.20", "1.21", "1.22", "1.23"}
	filtered := filter.Filter(availableTags)

	expected := []string{"latest", "1.21", "1.22"}
	assert.ElementsMatch(t, expected, filtered)
}

func TestTagFilter_Filter_Regex(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		TagRegex:   "^1\\.2[0-9]$",
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	availableTags := []string{"latest", "1.20", "1.21", "1.22", "1.30", "2.0"}
	filtered := filter.Filter(availableTags)

	expected := []string{"1.20", "1.21", "1.22"}
	assert.ElementsMatch(t, expected, filtered)
}

func TestTagFilter_Filter_Semver(t *testing.T) {
	image := ImageSync{
		Repository:       "nginx",
		SemverConstraint: ">=1.21.0 <1.23.0",
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	availableTags := []string{"1.20.0", "1.21.0", "1.22.0", "1.23.0", "1.24.0"}
	filtered := filter.Filter(availableTags)

	expected := []string{"1.22.0", "1.21.0"} // Sorted descending
	assert.Equal(t, expected, filtered)
}

func TestTagFilter_Filter_AllTags(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		AllTags:    true,
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	availableTags := []string{"latest", "1.20", "1.21", "1.22"}
	filtered := filter.Filter(availableTags)

	assert.Equal(t, availableTags, filtered)
}

func TestTagFilter_Filter_LatestN(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		LatestN:    3,
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	availableTags := []string{"1.20", "1.21", "1.22", "1.23", "1.24"}
	filtered := filter.Filter(availableTags)

	// Should return first 3 (without metadata, just slice first N)
	assert.Len(t, filtered, 3)
	assert.Equal(t, []string{"1.20", "1.21", "1.22"}, filtered)
}

func TestApplyLimit(t *testing.T) {
	tags := []string{"tag1", "tag2", "tag3", "tag4", "tag5"}

	tests := []struct {
		name     string
		limit    int
		expected []string
	}{
		{"no limit", 0, tags},
		{"no limit negative", -1, tags},
		{"limit smaller", 3, []string{"tag1", "tag2", "tag3"}},
		{"limit equal", 5, tags},
		{"limit larger", 10, tags},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyLimit(tags, tt.limit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExcludeTags(t *testing.T) {
	tags := []string{"latest", "1.20", "1.21", "1.22", "develop"}

	tests := []struct {
		name     string
		exclude  []string
		expected []string
	}{
		{
			name:     "no exclusions",
			exclude:  []string{},
			expected: tags,
		},
		{
			name:     "exclude one",
			exclude:  []string{"develop"},
			expected: []string{"latest", "1.20", "1.21", "1.22"},
		},
		{
			name:     "exclude multiple",
			exclude:  []string{"latest", "develop"},
			expected: []string{"1.20", "1.21", "1.22"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExcludeTags(tags, tt.exclude)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestFilterByPrefix(t *testing.T) {
	tags := []string{"v1.20", "v1.21", "1.22", "latest", "v2.0"}

	result := FilterByPrefix(tags, "v")
	expected := []string{"v1.20", "v1.21", "v2.0"}
	assert.ElementsMatch(t, expected, result)
}

func TestFilterBySuffix(t *testing.T) {
	tags := []string{"1.20-alpine", "1.21-alpine", "1.22-debian", "latest"}

	result := FilterBySuffix(tags, "-alpine")
	expected := []string{"1.20-alpine", "1.21-alpine"}
	assert.ElementsMatch(t, expected, result)
}

func TestFilterByPattern(t *testing.T) {
	tags := []string{"1.20.0", "1.21.0", "2.0.0", "latest", "1.20.1"}

	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "wildcard at end",
			pattern:  "1.20.*",
			expected: []string{"1.20.0", "1.20.1"},
		},
		{
			name:     "wildcard in middle",
			pattern:  "1.*.0",
			expected: []string{"1.20.0", "1.21.0"},
		},
		{
			name:     "exact match",
			pattern:  "latest",
			expected: []string{"latest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterByPattern(tags, tt.pattern)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestDeduplicateTags(t *testing.T) {
	tags := []string{"latest", "1.20", "latest", "1.21", "1.20", "1.22"}

	result := DeduplicateTags(tags)

	expected := []string{"latest", "1.20", "1.21", "1.22"}
	assert.Equal(t, expected, result)
}

func TestCombineFilters(t *testing.T) {
	availableTags := []string{"v1.20.0", "v1.21.0", "v1.22.0", "v2.0.0", "latest"}

	// Create filter 1: prefix "v"
	filter1Image := ImageSync{Repository: "nginx", AllTags: true}
	_, err := NewTagFilter(filter1Image)
	require.NoError(t, err)

	// Apply prefix separately
	prefixFiltered := FilterByPrefix(availableTags, "v")

	// Create filter 2: semver constraint
	filter2Image := ImageSync{Repository: "nginx", SemverConstraint: "^1.20.0"}
	filter2, err := NewTagFilter(filter2Image)
	require.NoError(t, err)

	// Combine filters
	result := CombineFilters(prefixFiltered, filter2)

	// Should have v1.20.0, v1.21.0, v1.22.0 (sorted desc)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "v1.22.0")
	assert.Contains(t, result, "v1.21.0")
	assert.Contains(t, result, "v1.20.0")
}

func TestTagFilter_InvalidRegex(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		TagRegex:   "[invalid(regex",
	}

	_, err := NewTagFilter(image)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tag regex")
}

func TestTagFilter_InvalidSemver(t *testing.T) {
	image := ImageSync{
		Repository:       "nginx",
		SemverConstraint: "invalid constraint",
	}

	_, err := NewTagFilter(image)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid semver constraint")
}
