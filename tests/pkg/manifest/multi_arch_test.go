package manifest

import (
	"runtime"
	"testing"

	"freightliner/pkg/manifest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformValidation(t *testing.T) {
	tests := []struct {
		name     string
		platform manifest.Platform
		wantErr  bool
	}{
		{
			name: "valid linux/amd64",
			platform: manifest.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			wantErr: false,
		},
		{
			name: "valid with variant",
			platform: manifest.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			wantErr: false,
		},
		{
			name: "empty OS",
			platform: manifest.Platform{
				Architecture: "amd64",
			},
			wantErr: true,
		},
		{
			name: "empty architecture",
			platform: manifest.Platform{
				OS: "linux",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.platform.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlatformMatches(t *testing.T) {
	tests := []struct {
		name     string
		p1       manifest.Platform
		p2       manifest.Platform
		expected bool
	}{
		{
			name: "exact match",
			p1: manifest.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			p2: manifest.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			expected: true,
		},
		{
			name: "different OS",
			p1: manifest.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			p2: manifest.Platform{
				OS:           "windows",
				Architecture: "amd64",
			},
			expected: false,
		},
		{
			name: "different architecture",
			p1: manifest.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			p2: manifest.Platform{
				OS:           "linux",
				Architecture: "arm64",
			},
			expected: false,
		},
		{
			name: "variant match",
			p1: manifest.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			p2: manifest.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			expected: true,
		},
		{
			name: "variant mismatch",
			p1: manifest.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			p2: manifest.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v6",
			},
			expected: false,
		},
		{
			name: "one variant empty",
			p1: manifest.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			p2: manifest.Platform{
				OS:           "linux",
				Architecture: "arm",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.p1.Matches(&tt.p2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlatformMatchesRuntime(t *testing.T) {
	currentPlatform := manifest.Platform{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	assert.True(t, currentPlatform.MatchesRuntime())

	differentPlatform := manifest.Platform{
		OS:           "nonexistent",
		Architecture: "nonexistent",
	}

	assert.False(t, differentPlatform.MatchesRuntime())
}

func TestPlatformString(t *testing.T) {
	tests := []struct {
		name     string
		platform manifest.Platform
		expected string
	}{
		{
			name: "basic platform",
			platform: manifest.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			expected: "linux/amd64",
		},
		{
			name: "with variant",
			platform: manifest.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			expected: "linux/arm/v7",
		},
		{
			name: "with OS version",
			platform: manifest.Platform{
				OS:           "windows",
				Architecture: "amd64",
				OSVersion:    "10.0.19041",
			},
			expected: "windows/amd64:10.0.19041",
		},
		{
			name: "with variant and OS version",
			platform: manifest.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
				OSVersion:    "5.4.0",
			},
			expected: "linux/arm/v7:5.4.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.platform.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMultiArchBuilder(t *testing.T) {
	builder := manifest.NewMultiArchBuilder(manifest.ManifestTypeOCIIndex)

	platform1 := manifest.PlatformManifest{
		MediaType: "application/vnd.oci.image.manifest.v1+json",
		Digest:    "sha256:amd64digest",
		Size:      1234,
		Platform: manifest.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
	}

	platform2 := manifest.PlatformManifest{
		MediaType: "application/vnd.oci.image.manifest.v1+json",
		Digest:    "sha256:arm64digest",
		Size:      5678,
		Platform: manifest.Platform{
			OS:           "linux",
			Architecture: "arm64",
		},
	}

	err := builder.AddPlatformManifest(platform1)
	require.NoError(t, err)

	err = builder.AddPlatformManifest(platform2)
	require.NoError(t, err)

	builder.AddAnnotation("org.opencontainers.image.created", "2024-01-01T00:00:00Z")

	multiArch, err := builder.Build()
	require.NoError(t, err)

	assert.Equal(t, 2, multiArch.SchemaVersion)
	assert.Equal(t, "application/vnd.oci.image.index.v1+json", multiArch.MediaType)
	assert.Len(t, multiArch.Manifests, 2)
	assert.NotNil(t, multiArch.Annotations)
	assert.Contains(t, multiArch.Annotations, "org.opencontainers.image.created")
}

func TestMultiArchBuilderDuplicatePlatform(t *testing.T) {
	builder := manifest.NewMultiArchBuilder(manifest.ManifestTypeDockerManifestList)

	platform1 := manifest.PlatformManifest{
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:    "sha256:digest1",
		Size:      1234,
		Platform: manifest.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
	}

	platform2 := manifest.PlatformManifest{
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:    "sha256:digest2",
		Size:      5678,
		Platform: manifest.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
	}

	err := builder.AddPlatformManifest(platform1)
	require.NoError(t, err)

	err = builder.AddPlatformManifest(platform2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGetManifestForPlatform(t *testing.T) {
	multiArch := &manifest.MultiArchManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests: []manifest.PlatformManifest{
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:amd64digest",
				Size:      1234,
				Platform: manifest.Platform{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:arm64digest",
				Size:      5678,
				Platform: manifest.Platform{
					OS:           "linux",
					Architecture: "arm64",
				},
			},
		},
	}

	t.Run("found platform", func(t *testing.T) {
		pm, err := multiArch.GetManifestForPlatform("linux", "amd64")
		require.NoError(t, err)
		assert.Equal(t, "sha256:amd64digest", pm.Digest)
		assert.Equal(t, "amd64", pm.Platform.Architecture)
	})

	t.Run("not found platform", func(t *testing.T) {
		_, err := multiArch.GetManifestForPlatform("windows", "amd64")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no manifest found")
	})
}

func TestGetManifestForCurrentPlatform(t *testing.T) {
	multiArch := &manifest.MultiArchManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests: []manifest.PlatformManifest{
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:currentdigest",
				Size:      1234,
				Platform: manifest.Platform{
					OS:           runtime.GOOS,
					Architecture: runtime.GOARCH,
				},
			},
		},
	}

	pm, err := multiArch.GetManifestForCurrentPlatform()
	require.NoError(t, err)
	assert.Equal(t, "sha256:currentdigest", pm.Digest)
	assert.Equal(t, runtime.GOOS, pm.Platform.OS)
	assert.Equal(t, runtime.GOARCH, pm.Platform.Architecture)
}

func TestGetPlatforms(t *testing.T) {
	multiArch := &manifest.MultiArchManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests: []manifest.PlatformManifest{
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:digest1",
				Size:      1234,
				Platform: manifest.Platform{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:digest2",
				Size:      5678,
				Platform: manifest.Platform{
					OS:           "linux",
					Architecture: "arm64",
				},
			},
		},
	}

	platforms := multiArch.GetPlatforms()
	assert.Len(t, platforms, 2)
	assert.Equal(t, "linux", platforms[0].OS)
	assert.Equal(t, "amd64", platforms[0].Architecture)
	assert.Equal(t, "linux", platforms[1].OS)
	assert.Equal(t, "arm64", platforms[1].Architecture)
}

func TestHasPlatform(t *testing.T) {
	multiArch := &manifest.MultiArchManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests: []manifest.PlatformManifest{
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:amd64digest",
				Size:      1234,
				Platform: manifest.Platform{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
		},
	}

	assert.True(t, multiArch.HasPlatform("linux", "amd64"))
	assert.False(t, multiArch.HasPlatform("windows", "amd64"))
}

func TestAddManifest(t *testing.T) {
	multiArch := &manifest.MultiArchManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests:     []manifest.PlatformManifest{},
	}

	newManifest := manifest.PlatformManifest{
		MediaType: "application/vnd.oci.image.manifest.v1+json",
		Digest:    "sha256:newdigest",
		Size:      1234,
		Platform: manifest.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
	}

	err := multiArch.AddManifest(newManifest)
	require.NoError(t, err)
	assert.Len(t, multiArch.Manifests, 1)

	// Try to add duplicate
	err = multiArch.AddManifest(newManifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRemoveManifest(t *testing.T) {
	multiArch := &manifest.MultiArchManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests: []manifest.PlatformManifest{
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:digest1",
				Size:      1234,
				Platform: manifest.Platform{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:digest2",
				Size:      5678,
				Platform: manifest.Platform{
					OS:           "linux",
					Architecture: "arm64",
				},
			},
		},
	}

	err := multiArch.RemoveManifest("sha256:digest1")
	require.NoError(t, err)
	assert.Len(t, multiArch.Manifests, 1)
	assert.Equal(t, "sha256:digest2", multiArch.Manifests[0].Digest)

	err = multiArch.RemoveManifest("sha256:nonexistent")
	assert.Error(t, err)
}

func TestRemovePlatform(t *testing.T) {
	multiArch := &manifest.MultiArchManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests: []manifest.PlatformManifest{
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:digest1",
				Size:      1234,
				Platform: manifest.Platform{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:digest2",
				Size:      5678,
				Platform: manifest.Platform{
					OS:           "linux",
					Architecture: "arm64",
				},
			},
		},
	}

	err := multiArch.RemovePlatform("linux", "amd64")
	require.NoError(t, err)
	assert.Len(t, multiArch.Manifests, 1)
	assert.Equal(t, "arm64", multiArch.Manifests[0].Platform.Architecture)

	err = multiArch.RemovePlatform("windows", "amd64")
	assert.Error(t, err)
}

func TestCreateMultiArch(t *testing.T) {
	manifests := []manifest.PlatformManifest{
		{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Digest:    "sha256:digest1",
			Size:      1234,
			Platform: manifest.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
		},
		{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Digest:    "sha256:digest2",
			Size:      5678,
			Platform: manifest.Platform{
				OS:           "linux",
				Architecture: "arm64",
			},
		},
	}

	multiArch, err := manifest.CreateMultiArch(manifests, manifest.ManifestTypeOCIIndex)
	require.NoError(t, err)

	assert.Equal(t, 2, multiArch.SchemaVersion)
	assert.Equal(t, "application/vnd.oci.image.index.v1+json", multiArch.MediaType)
	assert.Len(t, multiArch.Manifests, 2)
}

func TestMultiArchValidation(t *testing.T) {
	t.Run("valid manifest", func(t *testing.T) {
		multiArch := &manifest.MultiArchManifest{
			SchemaVersion: 2,
			MediaType:     "application/vnd.oci.image.index.v1+json",
			Manifests: []manifest.PlatformManifest{
				{
					MediaType: "application/vnd.oci.image.manifest.v1+json",
					Digest:    "sha256:validdigest",
					Size:      1234,
					Platform: manifest.Platform{
						OS:           "linux",
						Architecture: "amd64",
					},
				},
			},
		}

		err := multiArch.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid schema version", func(t *testing.T) {
		multiArch := &manifest.MultiArchManifest{
			SchemaVersion: 1,
			MediaType:     "application/vnd.oci.image.index.v1+json",
			Manifests:     []manifest.PlatformManifest{},
		}

		err := multiArch.Validate()
		assert.Error(t, err)
	})

	t.Run("no manifests", func(t *testing.T) {
		multiArch := &manifest.MultiArchManifest{
			SchemaVersion: 2,
			MediaType:     "application/vnd.oci.image.index.v1+json",
			Manifests:     []manifest.PlatformManifest{},
		}

		err := multiArch.Validate()
		assert.Error(t, err)
	})
}
