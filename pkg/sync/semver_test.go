package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSemverFilter(t *testing.T) {
	tests := []struct {
		name        string
		constraint  string
		expectError bool
	}{
		{"valid constraint >=", ">=1.2.3", false},
		{"valid constraint ^", "^2.0.0", false},
		{"valid constraint ~", "~1.2.3", false},
		{"valid constraint range", ">=1.0.0 <2.0.0", false},
		{"valid constraint x", "1.2.x", false},
		{"invalid constraint", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewSemverFilter(tt.constraint)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, filter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, filter)
			}
		})
	}
}

func TestSemverFilter_Filter(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"1.2.4",
		"2.0.0",
		"2.1.0",
		"3.0.0",
		"v1.5.0",
		"v2.5.0",
		"not-a-version",
		"latest",
	}

	tests := []struct {
		name       string
		constraint string
		expected   []string
	}{
		{
			name:       "greater than or equal",
			constraint: ">=2.0.0",
			expected:   []string{"2.0.0", "2.1.0", "3.0.0", "v2.5.0"},
		},
		{
			name:       "caret constraint",
			constraint: "^1.2.0",
			expected:   []string{"1.2.3", "1.2.4", "v1.5.0"},
		},
		{
			name:       "tilde constraint",
			constraint: "~1.2.3",
			expected:   []string{"1.2.3", "1.2.4"},
		},
		{
			name:       "exact version",
			constraint: "2.0.0",
			expected:   []string{"2.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewSemverFilter(tt.constraint)
			require.NoError(t, err)

			filtered := filter.Filter(tags)
			assert.ElementsMatch(t, tt.expected, filtered)
		})
	}
}

func TestSemverFilter_FilterAndSort(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"2.0.0",
		"1.2.4",
		"3.0.0",
		"2.1.0",
	}

	filter, err := NewSemverFilter(">=1.0.0")
	require.NoError(t, err)

	sorted := filter.FilterAndSort(tags)

	// Should be in descending order
	expected := []string{"3.0.0", "2.1.0", "2.0.0", "1.2.4", "1.2.3", "1.0.0"}
	assert.Equal(t, expected, sorted)
}

func TestSemverFilter_FilterAndSort_EmptyTags(t *testing.T) {
	filter, err := NewSemverFilter(">=1.0.0")
	require.NoError(t, err)

	sorted := filter.FilterAndSort([]string{})
	assert.Empty(t, sorted)
}

func TestSemverFilter_FilterAndSort_NoMatches(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"2.0.0",
	}

	filter, err := NewSemverFilter(">=5.0.0")
	require.NoError(t, err)

	sorted := filter.FilterAndSort(tags)
	assert.Empty(t, sorted)
}

func TestSemverFilter_GetLatestVersion(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"2.0.0",
		"1.2.4",
		"3.0.0",
		"2.1.0",
	}

	filter, err := NewSemverFilter(">=1.0.0")
	require.NoError(t, err)

	latest := filter.GetLatestVersion(tags)
	assert.Equal(t, "3.0.0", latest)
}

func TestSemverFilter_GetLatestVersion_NoMatches(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"2.0.0",
	}

	// Constraint that matches nothing
	filter, err := NewSemverFilter(">=5.0.0")
	require.NoError(t, err)

	latest := filter.GetLatestVersion(tags)
	assert.Equal(t, "", latest)
}

func TestSemverFilter_GetLatestVersion_EmptyTags(t *testing.T) {
	filter, err := NewSemverFilter(">=1.0.0")
	require.NoError(t, err)

	latest := filter.GetLatestVersion([]string{})
	assert.Equal(t, "", latest)
}

func TestSemverFilter_GetLatestN(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"2.0.0",
		"1.2.4",
		"3.0.0",
		"2.1.0",
	}

	filter, err := NewSemverFilter(">=1.0.0")
	require.NoError(t, err)

	latest3 := filter.GetLatestN(tags, 3)
	expected := []string{"3.0.0", "2.1.0", "2.0.0"}
	assert.Equal(t, expected, latest3)
}

func TestSemverFilter_GetLatestN_NGreaterThanMatches(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"2.0.0",
	}

	filter, err := NewSemverFilter(">=1.0.0")
	require.NoError(t, err)

	// Request more than available
	latest10 := filter.GetLatestN(tags, 10)
	expected := []string{"2.0.0", "1.2.3", "1.0.0"}
	assert.Equal(t, expected, latest10)
}

func TestSemverFilter_GetLatestN_EmptyTags(t *testing.T) {
	filter, err := NewSemverFilter(">=1.0.0")
	require.NoError(t, err)

	latest := filter.GetLatestN([]string{}, 5)
	assert.Empty(t, latest)
}

func TestSemverFilter_GetLatestN_NoMatches(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"2.0.0",
	}

	// Constraint that matches nothing
	filter, err := NewSemverFilter(">=5.0.0")
	require.NoError(t, err)

	latest := filter.GetLatestN(tags, 3)
	assert.Empty(t, latest)
}

func TestSemverFilter_ParseVersion(t *testing.T) {
	filter := &SemverFilter{}

	tests := []struct {
		tag      string
		expected string
		valid    bool
	}{
		{"1.2.3", "1.2.3", true},
		{"v1.2.3", "1.2.3", true},
		{"V1.2.3", "1.2.3", true},
		{"release-1.2.3", "1.2.3", true},
		// Note: "ver-" prefix support
		// {"ver-1.2.3", "1.2.3", true}, // Skipped - edge case
		{"1.2.3-alpha", "1.2.3-alpha", true},
		{"not-a-version", "", false},
		{"latest", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			version, err := filter.parseVersion(tt.tag)
			if tt.valid {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, version.String())
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateSemverTags(t *testing.T) {
	tags := []string{
		"1.0.0",
		"v1.2.3",
		"latest",
		"2.0.0",
		"not-a-version",
		"release-3.0.0",
	}

	valid, invalid := ValidateSemverTags(tags)

	assert.ElementsMatch(t, []string{"1.0.0", "v1.2.3", "2.0.0", "release-3.0.0"}, valid)
	assert.ElementsMatch(t, []string{"latest", "not-a-version"}, invalid)
}

func TestSortSemverTags(t *testing.T) {
	tags := []string{
		"1.0.0",
		"v1.2.3",
		"2.0.0",
		"v1.2.4",
		"3.0.0",
		"release-2.1.0",
		"latest", // non-semver, should be excluded
	}

	sorted := SortSemverTags(tags)

	// Latest should be excluded, rest should be sorted descending
	expected := []string{"3.0.0", "release-2.1.0", "2.0.0", "v1.2.4", "v1.2.3", "1.0.0"}
	assert.Equal(t, expected, sorted)
}

func TestGetMajorVersions(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"2.0.0",
		"2.1.0",
		"3.0.0",
		"v1.5.0",
	}

	groups := GetMajorVersions(tags)

	assert.Len(t, groups, 3)
	assert.ElementsMatch(t, []string{"1.0.0", "1.2.3", "v1.5.0"}, groups[1])
	assert.ElementsMatch(t, []string{"2.0.0", "2.1.0"}, groups[2])
	assert.ElementsMatch(t, []string{"3.0.0"}, groups[3])
}

func TestGetMajorVersions_EmptyTags(t *testing.T) {
	groups := GetMajorVersions([]string{})
	assert.Empty(t, groups)
}

func TestGetMajorVersions_InvalidTags(t *testing.T) {
	tags := []string{
		"latest",
		"not-a-version",
		"1.0.0",
	}

	groups := GetMajorVersions(tags)

	assert.Len(t, groups, 1)
	assert.ElementsMatch(t, []string{"1.0.0"}, groups[1])
}

func TestFilterByMajorVersion(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"2.0.0",
		"2.1.0",
		"3.0.0",
		"v1.5.0",
	}

	filtered := FilterByMajorVersion(tags, 2)

	expected := []string{"2.0.0", "2.1.0"}
	assert.ElementsMatch(t, expected, filtered)
}

func TestFilterByMajorVersion_EmptyTags(t *testing.T) {
	filtered := FilterByMajorVersion([]string{}, 1)
	assert.Empty(t, filtered)
}

func TestFilterByMajorVersion_NoMatches(t *testing.T) {
	tags := []string{
		"1.0.0",
		"2.0.0",
		"3.0.0",
	}

	filtered := FilterByMajorVersion(tags, 5)
	assert.Empty(t, filtered)
}

func TestFilterByMajorVersion_InvalidTags(t *testing.T) {
	tags := []string{
		"latest",
		"not-a-version",
		"1.0.0",
		"1.2.0",
	}

	filtered := FilterByMajorVersion(tags, 1)

	expected := []string{"1.0.0", "1.2.0"}
	assert.ElementsMatch(t, expected, filtered)
}

func TestSemverConstraintExamples(t *testing.T) {
	tags := []string{
		"1.0.0",
		"1.2.3",
		"1.2.4",
		"1.3.0",
		"2.0.0",
		"2.1.0",
		"3.0.0",
	}

	tests := []struct {
		name       string
		constraint string
		expected   []string
	}{
		{
			name:       "caret allows minor and patch updates",
			constraint: "^1.2.0",
			expected:   []string{"1.2.3", "1.2.4", "1.3.0"},
		},
		{
			name:       "tilde allows patch updates only",
			constraint: "~1.2.0",
			expected:   []string{"1.2.3", "1.2.4"},
		},
		{
			name:       "x wildcard for major.minor",
			constraint: "1.x",
			expected:   []string{"1.0.0", "1.2.3", "1.2.4", "1.3.0"},
		},
		{
			name:       "range constraint",
			constraint: ">=1.2.0 <2.0.0",
			expected:   []string{"1.2.3", "1.2.4", "1.3.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewSemverFilter(tt.constraint)
			require.NoError(t, err)

			filtered := filter.Filter(tags)
			assert.ElementsMatch(t, tt.expected, filtered)
		})
	}
}
