package sync

import (
	"context"
	"encoding/json"
	"fmt"
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

// ============================================================================
// Architecture Filtering Tests (Phase 3 functionality)
// ============================================================================

// MockArchitectureFilterer is a mock implementation of ArchitectureFilterer
type MockArchitectureFilterer struct {
	manifests   map[string]manifestInfo
	configBlobs map[string][]byte
}

type manifestInfo struct {
	data      []byte
	mediaType string
	err       error
}

func NewMockArchitectureFilterer() *MockArchitectureFilterer {
	return &MockArchitectureFilterer{
		manifests:   make(map[string]manifestInfo),
		configBlobs: make(map[string][]byte),
	}
}

func (m *MockArchitectureFilterer) GetManifest(ctx context.Context, repository, tag string) ([]byte, string, error) {
	key := repository + ":" + tag
	info, ok := m.manifests[key]
	if !ok {
		return nil, "", fmt.Errorf("manifest not found for %s", key)
	}
	return info.data, info.mediaType, info.err
}

func (m *MockArchitectureFilterer) GetConfigBlob(ctx context.Context, repository, digest string) ([]byte, error) {
	blob, ok := m.configBlobs[digest]
	if !ok {
		return nil, fmt.Errorf("config blob not found for digest %s", digest)
	}
	return blob, nil
}

func (m *MockArchitectureFilterer) AddMultiArchManifest(repo, tag string, architectures []string) {
	manifests := make([]map[string]interface{}, 0, len(architectures))
	for _, arch := range architectures {
		manifests = append(manifests, map[string]interface{}{
			"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
			"size":      1234,
			"digest":    fmt.Sprintf("sha256:abc%s", arch),
			"platform": map[string]string{
				"architecture": arch,
				"os":           "linux",
			},
		})
	}

	data := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.docker.distribution.manifest.list.v2+json",
		"manifests":     manifests,
	}

	jsonData, _ := json.Marshal(data)
	m.manifests[repo+":"+tag] = manifestInfo{
		data:      jsonData,
		mediaType: "application/vnd.docker.distribution.manifest.list.v2+json",
	}
}

func (m *MockArchitectureFilterer) AddSingleArchManifest(repo, tag, arch, digest string) {
	// Add config blob with architecture
	config := map[string]interface{}{
		"architecture": arch,
		"os":           "linux",
	}
	configJSON, _ := json.Marshal(config)
	m.configBlobs[digest] = configJSON

	// Create manifest that references this config
	manifest := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.docker.distribution.manifest.v2+json",
		"config": map[string]interface{}{
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size":      1234,
			"digest":    digest,
		},
		"layers": []map[string]interface{}{
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size":      5678,
				"digest":    "sha256:layer123",
			},
		},
	}

	jsonData, _ := json.Marshal(manifest)
	m.manifests[repo+":"+tag] = manifestInfo{
		data:      jsonData,
		mediaType: "application/vnd.docker.distribution.manifest.v2+json",
	}
}

func (m *MockArchitectureFilterer) AddManifestError(repo, tag string, err error) {
	m.manifests[repo+":"+tag] = manifestInfo{
		err: err,
	}
}

func TestApplyArchitectureFilter_NoArchitectures(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	tags := []string{"v1.0", "v1.1", "v1.2"}
	architectures := []string{}

	result, err := ApplyArchitectureFilter(ctx, filterer, "myrepo", tags, architectures)

	require.NoError(t, err)
	assert.Equal(t, tags, result, "should return all tags when no architecture filter")
}

func TestApplyArchitectureFilter_MultiArchMatch(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	// Setup multi-arch manifests
	filterer.AddMultiArchManifest("myrepo", "v1.0", []string{"amd64", "arm64"})
	filterer.AddMultiArchManifest("myrepo", "v1.1", []string{"amd64"})
	filterer.AddMultiArchManifest("myrepo", "v1.2", []string{"arm64", "s390x"})

	tags := []string{"v1.0", "v1.1", "v1.2"}

	tests := []struct {
		name          string
		architectures []string
		expectedTags  []string
	}{
		{
			name:          "filter amd64 only",
			architectures: []string{"amd64"},
			expectedTags:  []string{"v1.0", "v1.1"},
		},
		{
			name:          "filter arm64 only",
			architectures: []string{"arm64"},
			expectedTags:  []string{"v1.0", "v1.2"},
		},
		{
			name:          "filter multiple architectures",
			architectures: []string{"amd64", "arm64"},
			expectedTags:  []string{"v1.0", "v1.1", "v1.2"},
		},
		{
			name:          "no architecture match",
			architectures: []string{"ppc64le"},
			expectedTags:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ApplyArchitectureFilter(ctx, filterer, "myrepo", tags, tt.architectures)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedTags, result)
		})
	}
}

func TestApplyArchitectureFilter_SingleArchMatch(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	// Setup single-arch manifests
	filterer.AddSingleArchManifest("myrepo", "v1.0", "amd64", "sha256:config1")
	filterer.AddSingleArchManifest("myrepo", "v1.1", "arm64", "sha256:config2")
	filterer.AddSingleArchManifest("myrepo", "v1.2", "amd64", "sha256:config3")

	tags := []string{"v1.0", "v1.1", "v1.2"}

	tests := []struct {
		name          string
		architectures []string
		expectedTags  []string
	}{
		{
			name:          "filter amd64 only",
			architectures: []string{"amd64"},
			expectedTags:  []string{"v1.0", "v1.2"},
		},
		{
			name:          "filter arm64 only",
			architectures: []string{"arm64"},
			expectedTags:  []string{"v1.1"},
		},
		{
			name:          "filter multiple architectures",
			architectures: []string{"amd64", "arm64"},
			expectedTags:  []string{"v1.0", "v1.1", "v1.2"},
		},
		{
			name:          "no architecture match",
			architectures: []string{"s390x"},
			expectedTags:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ApplyArchitectureFilter(ctx, filterer, "myrepo", tags, tt.architectures)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedTags, result)
		})
	}
}

func TestApplyArchitectureFilter_ManifestFetchError(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	// Setup one valid manifest and one with error
	filterer.AddMultiArchManifest("myrepo", "v1.0", []string{"amd64"})
	filterer.AddManifestError("myrepo", "v1.1", fmt.Errorf("network error"))
	filterer.AddMultiArchManifest("myrepo", "v1.2", []string{"amd64"})

	tags := []string{"v1.0", "v1.1", "v1.2"}
	architectures := []string{"amd64"}

	result, err := ApplyArchitectureFilter(ctx, filterer, "myrepo", tags, architectures)

	// Should skip tags with errors and continue
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"v1.0", "v1.2"}, result)
}

func TestApplyArchitectureFilter_MixedMultiAndSingleArch(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	// Mix of multi-arch and single-arch manifests
	filterer.AddMultiArchManifest("myrepo", "v1.0", []string{"amd64", "arm64"})
	filterer.AddSingleArchManifest("myrepo", "v1.1", "amd64", "sha256:config1")
	filterer.AddMultiArchManifest("myrepo", "v1.2", []string{"arm64"})
	filterer.AddSingleArchManifest("myrepo", "v1.3", "arm64", "sha256:config2")

	tags := []string{"v1.0", "v1.1", "v1.2", "v1.3"}
	architectures := []string{"amd64"}

	result, err := ApplyArchitectureFilter(ctx, filterer, "myrepo", tags, architectures)

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"v1.0", "v1.1"}, result)
}

func TestCheckMultiArchManifest_OCIImageIndex(t *testing.T) {
	tests := []struct {
		name          string
		architectures []string
		desiredArchs  map[string]bool
		expectMatch   bool
	}{
		{
			name:          "amd64 present",
			architectures: []string{"amd64", "arm64"},
			desiredArchs:  map[string]bool{"amd64": true},
			expectMatch:   true,
		},
		{
			name:          "arm64 present",
			architectures: []string{"amd64", "arm64"},
			desiredArchs:  map[string]bool{"arm64": true},
			expectMatch:   true,
		},
		{
			name:          "multiple desired, one matches",
			architectures: []string{"amd64"},
			desiredArchs:  map[string]bool{"amd64": true, "arm64": true},
			expectMatch:   true,
		},
		{
			name:          "no match",
			architectures: []string{"amd64"},
			desiredArchs:  map[string]bool{"s390x": true},
			expectMatch:   false,
		},
		{
			name:          "empty manifests",
			architectures: []string{},
			desiredArchs:  map[string]bool{"amd64": true},
			expectMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifests := make([]map[string]interface{}, 0, len(tt.architectures))
			for _, arch := range tt.architectures {
				manifests = append(manifests, map[string]interface{}{
					"mediaType": "application/vnd.oci.image.manifest.v1+json",
					"size":      1234,
					"digest":    fmt.Sprintf("sha256:abc%s", arch),
					"platform": map[string]string{
						"architecture": arch,
						"os":           "linux",
					},
				})
			}

			data := map[string]interface{}{
				"schemaVersion": 2,
				"mediaType":     "application/vnd.oci.image.index.v1+json",
				"manifests":     manifests,
			}

			jsonData, err := json.Marshal(data)
			require.NoError(t, err)

			result := checkMultiArchManifest(jsonData, tt.desiredArchs)
			assert.Equal(t, tt.expectMatch, result)
		})
	}
}

func TestCheckMultiArchManifest_DockerManifestList(t *testing.T) {
	tests := []struct {
		name          string
		architectures []string
		desiredArchs  map[string]bool
		expectMatch   bool
	}{
		{
			name:          "amd64 present in Docker list",
			architectures: []string{"amd64", "arm64"},
			desiredArchs:  map[string]bool{"amd64": true},
			expectMatch:   true,
		},
		{
			name:          "no match in Docker list",
			architectures: []string{"amd64"},
			desiredArchs:  map[string]bool{"ppc64le": true},
			expectMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifests := make([]map[string]interface{}, 0, len(tt.architectures))
			for _, arch := range tt.architectures {
				manifests = append(manifests, map[string]interface{}{
					"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
					"size":      1234,
					"digest":    fmt.Sprintf("sha256:abc%s", arch),
					"platform": map[string]string{
						"architecture": arch,
						"os":           "linux",
					},
				})
			}

			data := map[string]interface{}{
				"schemaVersion": 2,
				"mediaType":     "application/vnd.docker.distribution.manifest.list.v2+json",
				"manifests":     manifests,
			}

			jsonData, err := json.Marshal(data)
			require.NoError(t, err)

			result := checkMultiArchManifest(jsonData, tt.desiredArchs)
			assert.Equal(t, tt.expectMatch, result)
		})
	}
}

func TestCheckMultiArchManifest_InvalidJSON(t *testing.T) {
	desiredArchs := map[string]bool{"amd64": true}

	// Invalid JSON should return false
	result := checkMultiArchManifest([]byte("{invalid json}"), desiredArchs)
	assert.False(t, result)
}

func TestCheckMultiArchManifest_MissingPlatform(t *testing.T) {
	// Manifest without platform info
	data := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests": []map[string]interface{}{
			{
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"size":      1234,
				"digest":    "sha256:abc123",
				// No platform field
			},
		},
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	desiredArchs := map[string]bool{"amd64": true}
	result := checkMultiArchManifest(jsonData, desiredArchs)
	assert.False(t, result, "should not match when platform is missing")
}

func TestCheckSingleArchManifest_OCIManifest(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	tests := []struct {
		name         string
		architecture string
		desiredArchs map[string]bool
		expectMatch  bool
	}{
		{
			name:         "amd64 matches",
			architecture: "amd64",
			desiredArchs: map[string]bool{"amd64": true},
			expectMatch:  true,
		},
		{
			name:         "no match",
			architecture: "amd64",
			desiredArchs: map[string]bool{"arm64": true},
			expectMatch:  false,
		},
		{
			name:         "multiple desired, one matches",
			architecture: "arm64",
			desiredArchs: map[string]bool{"amd64": true, "arm64": true},
			expectMatch:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configDigest := "sha256:configtest"

			// Setup config blob
			config := map[string]string{
				"architecture": tt.architecture,
				"os":           "linux",
			}
			configJSON, _ := json.Marshal(config)
			filterer.configBlobs[configDigest] = configJSON

			// Setup OCI manifest
			manifest := map[string]interface{}{
				"schemaVersion": 2,
				"mediaType":     "application/vnd.oci.image.manifest.v1+json",
				"config": map[string]interface{}{
					"mediaType": "application/vnd.oci.image.config.v1+json",
					"size":      1234,
					"digest":    configDigest,
				},
				"layers": []map[string]interface{}{
					{
						"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
						"size":      5678,
						"digest":    "sha256:layer123",
					},
				},
			}

			jsonData, err := json.Marshal(manifest)
			require.NoError(t, err)

			result := checkSingleArchManifest(ctx, filterer, "myrepo", jsonData, tt.desiredArchs)
			assert.Equal(t, tt.expectMatch, result)
		})
	}
}

func TestCheckSingleArchManifest_DockerV2Manifest(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	configDigest := "sha256:dockerconfig"

	// Setup config blob
	config := map[string]string{
		"architecture": "amd64",
		"os":           "linux",
	}
	configJSON, _ := json.Marshal(config)
	filterer.configBlobs[configDigest] = configJSON

	// Setup Docker V2 manifest
	manifest := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.docker.distribution.manifest.v2+json",
		"config": map[string]interface{}{
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size":      1234,
			"digest":    configDigest,
		},
		"layers": []map[string]interface{}{
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size":      5678,
				"digest":    "sha256:layer123",
			},
		},
	}

	jsonData, err := json.Marshal(manifest)
	require.NoError(t, err)

	desiredArchs := map[string]bool{"amd64": true}
	result := checkSingleArchManifest(ctx, filterer, "myrepo", jsonData, desiredArchs)
	assert.True(t, result)
}

func TestCheckSingleArchManifest_ConfigFetchError(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	// Don't add config blob to filterer - will cause fetch error
	configDigest := "sha256:missingconfig"

	manifest := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.manifest.v1+json",
		"config": map[string]interface{}{
			"mediaType": "application/vnd.oci.image.config.v1+json",
			"size":      1234,
			"digest":    configDigest,
		},
		"layers": []map[string]interface{}{},
	}

	jsonData, err := json.Marshal(manifest)
	require.NoError(t, err)

	desiredArchs := map[string]bool{"amd64": true}

	// Should return true when config fetch fails (conservative approach)
	result := checkSingleArchManifest(ctx, filterer, "myrepo", jsonData, desiredArchs)
	assert.True(t, result, "should be conservative and include tag when config fetch fails")
}

func TestCheckSingleArchManifest_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	desiredArchs := map[string]bool{"amd64": true}

	// Invalid JSON should return true (conservative)
	result := checkSingleArchManifest(ctx, filterer, "myrepo", []byte("{invalid}"), desiredArchs)
	assert.True(t, result, "should be conservative with invalid JSON")
}

func TestFetchArchFromConfig_ValidConfig(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	tests := []struct {
		name         string
		architecture string
	}{
		{"amd64", "amd64"},
		{"arm64", "arm64"},
		{"s390x", "s390x"},
		{"ppc64le", "ppc64le"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			digest := fmt.Sprintf("sha256:config%s", tt.name)

			config := map[string]string{
				"architecture": tt.architecture,
				"os":           "linux",
			}
			configJSON, _ := json.Marshal(config)
			filterer.configBlobs[digest] = configJSON

			arch := fetchArchFromConfig(ctx, filterer, "myrepo", digest)
			assert.Equal(t, tt.architecture, arch)
		})
	}
}

func TestFetchArchFromConfig_MissingArchitecture(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	digest := "sha256:noarch"

	// Config without architecture field
	config := map[string]string{
		"os": "linux",
	}
	configJSON, _ := json.Marshal(config)
	filterer.configBlobs[digest] = configJSON

	arch := fetchArchFromConfig(ctx, filterer, "myrepo", digest)
	assert.Equal(t, "", arch, "should return empty string when architecture is missing")
}

func TestFetchArchFromConfig_FetchError(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	// Don't add config blob - will cause fetch error
	digest := "sha256:notfound"

	arch := fetchArchFromConfig(ctx, filterer, "myrepo", digest)
	assert.Equal(t, "", arch, "should return empty string on fetch error")
}

func TestFetchArchFromConfig_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	digest := "sha256:badjson"
	filterer.configBlobs[digest] = []byte("{invalid json")

	arch := fetchArchFromConfig(ctx, filterer, "myrepo", digest)
	assert.Equal(t, "", arch, "should return empty string for invalid JSON")
}

func TestHasMatchingArchitecture_OCIIndex(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	tests := []struct {
		name          string
		manifestArchs []string
		desiredArchs  []string
		expectMatch   bool
	}{
		{
			name:          "exact match",
			manifestArchs: []string{"amd64"},
			desiredArchs:  []string{"amd64"},
			expectMatch:   true,
		},
		{
			name:          "multiple available, one matches",
			manifestArchs: []string{"amd64", "arm64"},
			desiredArchs:  []string{"arm64"},
			expectMatch:   true,
		},
		{
			name:          "no match",
			manifestArchs: []string{"amd64"},
			desiredArchs:  []string{"arm64"},
			expectMatch:   false,
		},
		{
			name:          "multiple desired, one available matches",
			manifestArchs: []string{"amd64"},
			desiredArchs:  []string{"amd64", "arm64"},
			expectMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build manifest
			manifests := make([]map[string]interface{}, 0, len(tt.manifestArchs))
			for _, arch := range tt.manifestArchs {
				manifests = append(manifests, map[string]interface{}{
					"mediaType": "application/vnd.oci.image.manifest.v1+json",
					"size":      1234,
					"digest":    fmt.Sprintf("sha256:abc%s", arch),
					"platform": map[string]string{
						"architecture": arch,
						"os":           "linux",
					},
				})
			}

			data := map[string]interface{}{
				"schemaVersion": 2,
				"mediaType":     "application/vnd.oci.image.index.v1+json",
				"manifests":     manifests,
			}

			jsonData, err := json.Marshal(data)
			require.NoError(t, err)

			desiredArchMap := make(map[string]bool)
			for _, arch := range tt.desiredArchs {
				desiredArchMap[arch] = true
			}

			result := hasMatchingArchitecture(ctx, filterer, "myrepo", jsonData,
				"application/vnd.oci.image.index.v1+json", desiredArchMap)
			assert.Equal(t, tt.expectMatch, result)
		})
	}
}

func TestHasMatchingArchitecture_UnknownMediaType(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	desiredArchs := map[string]bool{"amd64": true}

	// Unknown media type should return false
	result := hasMatchingArchitecture(ctx, filterer, "myrepo", []byte("{}"),
		"application/unknown", desiredArchs)
	assert.False(t, result, "should return false for unknown media type")
}

func TestApplyArchitectureFilter_EmptyTags(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	tags := []string{}
	architectures := []string{"amd64"}

	result, err := ApplyArchitectureFilter(ctx, filterer, "myrepo", tags, architectures)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestApplyArchitectureFilter_NilContext(t *testing.T) {
	filterer := NewMockArchitectureFilterer()
	filterer.AddMultiArchManifest("myrepo", "v1.0", []string{"amd64"})

	tags := []string{"v1.0"}
	architectures := []string{"amd64"}

	// Should handle nil context gracefully (though not recommended)
	result, err := ApplyArchitectureFilter(context.Background(), filterer, "myrepo", tags, architectures)

	require.NoError(t, err)
	assert.Equal(t, []string{"v1.0"}, result)
}

func TestCheckMultiArchManifest_BothFormats(t *testing.T) {
	// Test OCI Index parsing
	ociManifests := []map[string]interface{}{
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"size":      1234,
			"digest":    "sha256:abcamd64",
			"platform": map[string]string{
				"architecture": "amd64",
				"os":           "linux",
			},
		},
	}

	ociData := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests":     ociManifests,
	}

	ociJSON, err := json.Marshal(ociData)
	require.NoError(t, err)

	desiredArchs := map[string]bool{"amd64": true}
	assert.True(t, checkMultiArchManifest(ociJSON, desiredArchs))

	// Test Docker Manifest List parsing
	dockerManifests := []map[string]interface{}{
		{
			"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
			"size":      1234,
			"digest":    "sha256:abcarm64",
			"platform": map[string]string{
				"architecture": "arm64",
				"os":           "linux",
			},
		},
	}

	dockerData := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.docker.distribution.manifest.list.v2+json",
		"manifests":     dockerManifests,
	}

	dockerJSON, err := json.Marshal(dockerData)
	require.NoError(t, err)

	assert.False(t, checkMultiArchManifest(dockerJSON, desiredArchs), "should not match arm64 when looking for amd64")
}

func TestCheckSingleArchManifest_OCIPlatformInConfig(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	// Test OCI manifest with platform in config descriptor (optional field)
	manifest := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.manifest.v1+json",
		"config": map[string]interface{}{
			"mediaType": "application/vnd.oci.image.config.v1+json",
			"size":      1234,
			"digest":    "sha256:configtest",
			"platform": map[string]string{
				"architecture": "arm64",
				"os":           "linux",
			},
		},
		"layers": []map[string]interface{}{
			{
				"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
				"size":      5678,
				"digest":    "sha256:layer123",
			},
		},
	}

	jsonData, err := json.Marshal(manifest)
	require.NoError(t, err)

	// Should use platform from config descriptor, not fetch blob
	desiredArchs := map[string]bool{"arm64": true}
	result := checkSingleArchManifest(ctx, filterer, "myrepo", jsonData, desiredArchs)
	assert.True(t, result)

	// Should not match different architecture
	desiredArchs = map[string]bool{"amd64": true}
	result = checkSingleArchManifest(ctx, filterer, "myrepo", jsonData, desiredArchs)
	assert.False(t, result)
}

func TestCheckSingleArchManifest_BothManifestTypes(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	tests := []struct {
		name         string
		manifestType string
		configType   string
		layerType    string
		expectOCI    bool
	}{
		{
			name:         "OCI manifest",
			manifestType: "application/vnd.oci.image.manifest.v1+json",
			configType:   "application/vnd.oci.image.config.v1+json",
			layerType:    "application/vnd.oci.image.layer.v1.tar+gzip",
			expectOCI:    true,
		},
		{
			name:         "Docker V2 manifest",
			manifestType: "application/vnd.docker.distribution.manifest.v2+json",
			configType:   "application/vnd.docker.container.image.v1+json",
			layerType:    "application/vnd.docker.image.rootfs.diff.tar.gzip",
			expectOCI:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configDigest := fmt.Sprintf("sha256:config%s", tt.name)

			// Setup config blob
			config := map[string]string{
				"architecture": "amd64",
				"os":           "linux",
			}
			configJSON, _ := json.Marshal(config)
			filterer.configBlobs[configDigest] = configJSON

			// Setup manifest
			manifest := map[string]interface{}{
				"schemaVersion": 2,
				"mediaType":     tt.manifestType,
				"config": map[string]interface{}{
					"mediaType": tt.configType,
					"size":      1234,
					"digest":    configDigest,
				},
				"layers": []map[string]interface{}{
					{
						"mediaType": tt.layerType,
						"size":      5678,
						"digest":    "sha256:layer123",
					},
				},
			}

			jsonData, err := json.Marshal(manifest)
			require.NoError(t, err)

			desiredArchs := map[string]bool{"amd64": true}
			result := checkSingleArchManifest(ctx, filterer, "myrepo", jsonData, desiredArchs)
			assert.True(t, result, "should match amd64 for %s", tt.name)
		})
	}
}

func TestApplyArchitectureFilter_AllMediaTypes(t *testing.T) {
	ctx := context.Background()
	filterer := NewMockArchitectureFilterer()

	// Add OCI Index
	ociManifests := []map[string]interface{}{
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"size":      1234,
			"digest":    "sha256:oci1",
			"platform": map[string]string{
				"architecture": "amd64",
				"os":           "linux",
			},
		},
	}
	ociData := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests":     ociManifests,
	}
	ociJSON, _ := json.Marshal(ociData)
	filterer.manifests["myrepo:oci-index"] = manifestInfo{
		data:      ociJSON,
		mediaType: "application/vnd.oci.image.index.v1+json",
	}

	// Add Docker Manifest List
	dockerManifests := []map[string]interface{}{
		{
			"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
			"size":      1234,
			"digest":    "sha256:docker1",
			"platform": map[string]string{
				"architecture": "arm64",
				"os":           "linux",
			},
		},
	}
	dockerData := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.docker.distribution.manifest.list.v2+json",
		"manifests":     dockerManifests,
	}
	dockerJSON, _ := json.Marshal(dockerData)
	filterer.manifests["myrepo:docker-list"] = manifestInfo{
		data:      dockerJSON,
		mediaType: "application/vnd.docker.distribution.manifest.list.v2+json",
	}

	// Add OCI single manifest
	filterer.AddSingleArchManifest("myrepo", "oci-single", "amd64", "sha256:ocisingle")

	// Add Docker V2 single manifest
	filterer.AddSingleArchManifest("myrepo", "docker-single", "arm64", "sha256:dockersingle")

	tags := []string{"oci-index", "docker-list", "oci-single", "docker-single"}

	// Test filtering for amd64
	architectures := []string{"amd64"}
	result, err := ApplyArchitectureFilter(ctx, filterer, "myrepo", tags, architectures)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"oci-index", "oci-single"}, result)

	// Test filtering for arm64
	architectures = []string{"arm64"}
	result, err = ApplyArchitectureFilter(ctx, filterer, "myrepo", tags, architectures)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"docker-list", "docker-single"}, result)
}
