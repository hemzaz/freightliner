package sync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// FilterWithMetadata Tests (Phase 9g)
// ============================================================================

func TestTagFilter_FilterWithMetadata_LatestN(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		LatestN:    3,
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	// Create metadata with different creation times
	now := time.Now()
	tagsWithMetadata := []TagMetadata{
		{Tag: "1.20", CreatedAt: now.Add(-5 * time.Hour)},
		{Tag: "1.21", CreatedAt: now.Add(-3 * time.Hour)},
		{Tag: "1.22", CreatedAt: now.Add(-1 * time.Hour)}, // Newest
		{Tag: "1.19", CreatedAt: now.Add(-10 * time.Hour)},
		{Tag: "1.18", CreatedAt: now.Add(-15 * time.Hour)},
	}

	result := filter.FilterWithMetadata(tagsWithMetadata)

	// Should return the 3 newest tags sorted by creation time (descending)
	expected := []string{"1.22", "1.21", "1.20"}
	assert.Equal(t, expected, result)
}

func TestTagFilter_FilterWithMetadata_LatestN_LessThanN(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		LatestN:    5,
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	now := time.Now()
	tagsWithMetadata := []TagMetadata{
		{Tag: "1.20", CreatedAt: now.Add(-2 * time.Hour)},
		{Tag: "1.21", CreatedAt: now.Add(-1 * time.Hour)},
	}

	result := filter.FilterWithMetadata(tagsWithMetadata)

	// Should return all 2 tags when N=5 but only 2 available
	assert.Len(t, result, 2)
	assert.ElementsMatch(t, []string{"1.21", "1.20"}, result)
}

func TestTagFilter_FilterWithMetadata_LatestN_Empty(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		LatestN:    3,
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	result := filter.FilterWithMetadata([]TagMetadata{})

	assert.Empty(t, result)
}

func TestTagFilter_FilterWithMetadata_SpecificTags(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		Tags:       []string{"latest", "1.21"},
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	now := time.Now()
	tagsWithMetadata := []TagMetadata{
		{Tag: "latest", CreatedAt: now.Add(-1 * time.Hour)},
		{Tag: "1.20", CreatedAt: now.Add(-2 * time.Hour)},
		{Tag: "1.21", CreatedAt: now.Add(-3 * time.Hour)},
		{Tag: "1.22", CreatedAt: now.Add(-4 * time.Hour)},
	}

	result := filter.FilterWithMetadata(tagsWithMetadata)

	// Should extract tag names and use regular Filter
	expected := []string{"latest", "1.21"}
	assert.ElementsMatch(t, expected, result)
}

func TestTagFilter_FilterWithMetadata_Regex(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		TagRegex:   "^1\\.2[0-9]$",
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	now := time.Now()
	tagsWithMetadata := []TagMetadata{
		{Tag: "latest", CreatedAt: now},
		{Tag: "1.20", CreatedAt: now.Add(-1 * time.Hour)},
		{Tag: "1.21", CreatedAt: now.Add(-2 * time.Hour)},
		{Tag: "1.30", CreatedAt: now.Add(-3 * time.Hour)},
	}

	result := filter.FilterWithMetadata(tagsWithMetadata)

	expected := []string{"1.20", "1.21"}
	assert.ElementsMatch(t, expected, result)
}

func TestTagFilter_FilterWithMetadata_Semver(t *testing.T) {
	image := ImageSync{
		Repository:       "nginx",
		SemverConstraint: ">=1.21.0 <1.23.0",
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	now := time.Now()
	tagsWithMetadata := []TagMetadata{
		{Tag: "1.20.0", CreatedAt: now.Add(-5 * time.Hour)},
		{Tag: "1.21.0", CreatedAt: now.Add(-4 * time.Hour)},
		{Tag: "1.22.0", CreatedAt: now.Add(-3 * time.Hour)},
		{Tag: "1.23.0", CreatedAt: now.Add(-2 * time.Hour)},
	}

	result := filter.FilterWithMetadata(tagsWithMetadata)

	// Semver filter should return in descending order
	expected := []string{"1.22.0", "1.21.0"}
	assert.Equal(t, expected, result)
}

func TestTagFilter_FilterWithMetadata_AllTags(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		AllTags:    true,
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	now := time.Now()
	tagsWithMetadata := []TagMetadata{
		{Tag: "latest", CreatedAt: now},
		{Tag: "1.20", CreatedAt: now.Add(-1 * time.Hour)},
		{Tag: "1.21", CreatedAt: now.Add(-2 * time.Hour)},
	}

	result := filter.FilterWithMetadata(tagsWithMetadata)

	// Should return all tags
	expected := []string{"latest", "1.20", "1.21"}
	assert.ElementsMatch(t, expected, result)
}

func TestTagFilter_FilterWithMetadata_WithFullMetadata(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		LatestN:    2,
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	now := time.Now()
	tagsWithMetadata := []TagMetadata{
		{
			Tag:       "1.20",
			Digest:    "sha256:abc123",
			CreatedAt: now.Add(-3 * time.Hour),
			Size:      1024000,
			Platform:  "linux/amd64",
			Annotations: map[string]string{
				"org.opencontainers.image.version": "1.20",
			},
		},
		{
			Tag:       "1.21",
			Digest:    "sha256:def456",
			CreatedAt: now.Add(-1 * time.Hour),
			Size:      1025000,
			Platform:  "linux/arm64",
			Annotations: map[string]string{
				"org.opencontainers.image.version": "1.21",
			},
		},
	}

	result := filter.FilterWithMetadata(tagsWithMetadata)

	// Should return the 2 newest (all tags in this case)
	expected := []string{"1.21", "1.20"}
	assert.Equal(t, expected, result)
}

func TestTagFilter_FilterWithMetadata_SameCreationTime(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		LatestN:    2,
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	now := time.Now()
	// All tags have the same creation time
	tagsWithMetadata := []TagMetadata{
		{Tag: "1.20", CreatedAt: now},
		{Tag: "1.21", CreatedAt: now},
		{Tag: "1.22", CreatedAt: now},
	}

	result := filter.FilterWithMetadata(tagsWithMetadata)

	// Should return first N from the sorted list
	assert.Len(t, result, 2)
	// Order may vary for same timestamp, but should get 2 tags
	assert.Subset(t, []string{"1.20", "1.21", "1.22"}, result)
}

func TestTagFilter_FilterWithMetadata_ZeroLatestN(t *testing.T) {
	// When LatestN is 0, should not use the latestN path
	image := ImageSync{
		Repository: "nginx",
		AllTags:    true, // Use AllTags instead
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	now := time.Now()
	tagsWithMetadata := []TagMetadata{
		{Tag: "1.20", CreatedAt: now.Add(-2 * time.Hour)},
		{Tag: "1.21", CreatedAt: now.Add(-1 * time.Hour)},
	}

	result := filter.FilterWithMetadata(tagsWithMetadata)

	// Should use regular Filter path and return all
	assert.ElementsMatch(t, []string{"1.20", "1.21"}, result)
}

func TestTagFilter_FilterWithMetadata_ComplexMetadata(t *testing.T) {
	image := ImageSync{
		Repository: "nginx",
		LatestN:    3,
	}

	filter, err := NewTagFilter(image)
	require.NoError(t, err)

	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tagsWithMetadata := []TagMetadata{
		{
			Tag:       "v1.0.0",
			Digest:    "sha256:hash1",
			CreatedAt: baseTime,
			Size:      1000000,
			Platform:  "linux/amd64",
		},
		{
			Tag:       "v1.1.0",
			Digest:    "sha256:hash2",
			CreatedAt: baseTime.Add(24 * time.Hour),
			Size:      1100000,
			Platform:  "linux/amd64",
		},
		{
			Tag:       "v1.2.0",
			Digest:    "sha256:hash3",
			CreatedAt: baseTime.Add(48 * time.Hour),
			Size:      1200000,
			Platform:  "linux/arm64",
		},
		{
			Tag:       "v2.0.0",
			Digest:    "sha256:hash4",
			CreatedAt: baseTime.Add(72 * time.Hour), // Newest
			Size:      2000000,
			Platform:  "linux/amd64",
		},
		{
			Tag:       "v0.9.0",
			Digest:    "sha256:hash0",
			CreatedAt: baseTime.Add(-24 * time.Hour), // Oldest
			Size:      900000,
			Platform:  "linux/amd64",
		},
	}

	result := filter.FilterWithMetadata(tagsWithMetadata)

	// Should return the 3 newest by creation time
	expected := []string{"v2.0.0", "v1.2.0", "v1.1.0"}
	assert.Equal(t, expected, result)
}
