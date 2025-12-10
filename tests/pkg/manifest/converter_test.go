package manifest

import (
	"encoding/json"
	"testing"

	"freightliner/pkg/manifest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectManifestType(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		expected manifest.ManifestType
		wantErr  bool
	}{
		{
			name: "Docker v2 Schema 2",
			manifest: `{
				"schemaVersion": 2,
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"config": {
					"mediaType": "application/vnd.docker.container.image.v1+json",
					"size": 1234,
					"digest": "sha256:abc123"
				},
				"layers": []
			}`,
			expected: manifest.ManifestTypeDockerV2Schema2,
			wantErr:  false,
		},
		{
			name: "OCI v1",
			manifest: `{
				"schemaVersion": 2,
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"config": {
					"mediaType": "application/vnd.oci.image.config.v1+json",
					"size": 1234,
					"digest": "sha256:abc123"
				},
				"layers": []
			}`,
			expected: manifest.ManifestTypeOCIv1,
			wantErr:  false,
		},
		{
			name: "Docker Manifest List",
			manifest: `{
				"schemaVersion": 2,
				"mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
				"manifests": []
			}`,
			expected: manifest.ManifestTypeDockerManifestList,
			wantErr:  false,
		},
		{
			name: "OCI Image Index",
			manifest: `{
				"schemaVersion": 2,
				"mediaType": "application/vnd.oci.image.index.v1+json",
				"manifests": []
			}`,
			expected: manifest.ManifestTypeOCIIndex,
			wantErr:  false,
		},
		{
			name:     "Invalid JSON",
			manifest: `{invalid json}`,
			expected: manifest.ManifestTypeUnknown,
			wantErr:  true,
		},
	}

	converter := manifest.NewConverter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifestType, err := converter.DetectManifestType([]byte(tt.manifest))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, manifestType)
			}
		})
	}
}

func TestDockerToOCI(t *testing.T) {
	dockerManifest := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size": 1234,
			"digest": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
		},
		"layers": [
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size": 5678,
				"digest": "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
			}
		]
	}`

	converter := manifest.NewConverter()
	ociBytes, err := converter.DockerToOCI([]byte(dockerManifest))
	require.NoError(t, err)

	var ociManifest manifest.OCIManifest
	err = json.Unmarshal(ociBytes, &ociManifest)
	require.NoError(t, err)

	assert.Equal(t, 2, ociManifest.SchemaVersion)
	assert.Equal(t, "application/vnd.oci.image.manifest.v1+json", ociManifest.MediaType)
	assert.Equal(t, "application/vnd.oci.image.config.v1+json", ociManifest.Config.MediaType)
	assert.Equal(t, "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", ociManifest.Config.Digest)
	assert.Len(t, ociManifest.Layers, 1)
}

func TestOCIToDocker(t *testing.T) {
	ociManifest := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.manifest.v1+json",
		"config": {
			"mediaType": "application/vnd.oci.image.config.v1+json",
			"size": 1234,
			"digest": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
		},
		"layers": [
			{
				"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
				"size": 5678,
				"digest": "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
			}
		]
	}`

	converter := manifest.NewConverter()
	dockerBytes, err := converter.OCIToDocker([]byte(ociManifest))
	require.NoError(t, err)

	var dockerManifest manifest.DockerV2Schema2Manifest
	err = json.Unmarshal(dockerBytes, &dockerManifest)
	require.NoError(t, err)

	assert.Equal(t, 2, dockerManifest.SchemaVersion)
	assert.Equal(t, "application/vnd.docker.distribution.manifest.v2+json", dockerManifest.MediaType)
	assert.Equal(t, "application/vnd.docker.container.image.v1+json", dockerManifest.Config.MediaType)
	assert.Equal(t, "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", dockerManifest.Config.Digest)
	assert.Len(t, dockerManifest.Layers, 1)
}

func TestNormalize(t *testing.T) {
	dockerManifest := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size": 1234,
			"digest": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
		},
		"layers": [
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size": 5678,
				"digest": "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
			}
		],
		"annotations": {
			"key": "value"
		}
	}`

	converter := manifest.NewConverter()
	normalized, err := converter.Normalize([]byte(dockerManifest))
	require.NoError(t, err)

	assert.Equal(t, 2, normalized.SchemaVersion)
	assert.Equal(t, "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", normalized.Config.Digest)
	assert.Len(t, normalized.Layers, 1)
	assert.Equal(t, "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b", normalized.Layers[0].Digest)
	assert.NotNil(t, normalized.Annotations)
	assert.Equal(t, "value", normalized.Annotations["key"])
}

func TestConvertDockerManifestListToOCIIndex(t *testing.T) {
	dockerList := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
		"manifests": [
			{
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"size": 1234,
				"digest": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
				"platform": {
					"architecture": "amd64",
					"os": "linux"
				}
			},
			{
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"size": 5678,
				"digest": "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b",
				"platform": {
					"architecture": "arm64",
					"os": "linux"
				}
			}
		]
	}`

	converter := manifest.NewConverter()
	ociBytes, err := converter.DockerToOCI([]byte(dockerList))
	require.NoError(t, err)

	var ociIndex manifest.OCIImageIndex
	err = json.Unmarshal(ociBytes, &ociIndex)
	require.NoError(t, err)

	assert.Equal(t, 2, ociIndex.SchemaVersion)
	assert.Equal(t, "application/vnd.oci.image.index.v1+json", ociIndex.MediaType)
	assert.Len(t, ociIndex.Manifests, 2)
	assert.NotNil(t, ociIndex.Manifests[0].Platform)
	assert.Equal(t, "amd64", ociIndex.Manifests[0].Platform.Architecture)
}

func TestConverterWithStrictValidation(t *testing.T) {
	invalidManifest := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size": -1,
			"digest": "invalid-digest"
		},
		"layers": []
	}`

	converter := manifest.NewConverter()
	converter.StrictValidation = true

	_, err := converter.DockerToOCI([]byte(invalidManifest))
	assert.Error(t, err)
}

func TestPreserveAnnotations(t *testing.T) {
	dockerManifest := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size": 1234,
			"digest": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
		},
		"layers": [
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size": 5678,
				"digest": "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
			}
		],
		"annotations": {
			"org.opencontainers.image.created": "2024-01-01T00:00:00Z"
		}
	}`

	t.Run("with preservation", func(t *testing.T) {
		converter := manifest.NewConverter()
		converter.PreserveAnnotations = true

		ociBytes, err := converter.DockerToOCI([]byte(dockerManifest))
		require.NoError(t, err)

		var ociManifest manifest.OCIManifest
		err = json.Unmarshal(ociBytes, &ociManifest)
		require.NoError(t, err)

		assert.NotNil(t, ociManifest.Annotations)
		assert.Contains(t, ociManifest.Annotations, "org.opencontainers.image.created")
	})

	t.Run("without preservation", func(t *testing.T) {
		converter := manifest.NewConverter()
		converter.PreserveAnnotations = false

		ociBytes, err := converter.DockerToOCI([]byte(dockerManifest))
		require.NoError(t, err)

		var ociManifest manifest.OCIManifest
		err = json.Unmarshal(ociBytes, &ociManifest)
		require.NoError(t, err)

		assert.Nil(t, ociManifest.Annotations)
	})
}

func TestValidateDigest(t *testing.T) {
	tests := []struct {
		name    string
		digest  string
		wantErr bool
	}{
		{
			name:    "valid sha256",
			digest:  "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
			wantErr: false,
		},
		{
			name:    "valid sha512",
			digest:  "sha512:ee26b0dd4af7e749aa1a8ee3c10ae9923f618980772e473f8819a5d4940e0db27ac185f8a0e1d5f84f88bc887fd67b143732c304cc5fa9ad8e6f57f50028a8ff",
			wantErr: false,
		},
		{
			name:    "invalid format",
			digest:  "invalid-digest",
			wantErr: true,
		},
		{
			name:    "empty digest",
			digest:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manifest.ValidateDigest(tt.digest)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestComputeDigest(t *testing.T) {
	data := []byte("test data")
	digest := manifest.ComputeDigest(data)

	assert.NotEmpty(t, digest)
	assert.Contains(t, digest, "sha256:")

	// Computing the same data should produce the same digest
	digest2 := manifest.ComputeDigest(data)
	assert.Equal(t, digest, digest2)

	// Different data should produce different digest
	differentData := []byte("different data")
	digest3 := manifest.ComputeDigest(differentData)
	assert.NotEqual(t, digest, digest3)
}

func TestBidirectionalConversion(t *testing.T) {
	// Start with Docker manifest
	dockerManifest := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size": 1234,
			"digest": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
		},
		"layers": [
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size": 5678,
				"digest": "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
			}
		]
	}`

	converter := manifest.NewConverter()

	// Convert Docker to OCI
	ociBytes, err := converter.DockerToOCI([]byte(dockerManifest))
	require.NoError(t, err)

	// Convert OCI back to Docker
	dockerBytes, err := converter.OCIToDocker(ociBytes)
	require.NoError(t, err)

	// Parse both and compare important fields
	var original, roundtrip manifest.DockerV2Schema2Manifest
	err = json.Unmarshal([]byte(dockerManifest), &original)
	require.NoError(t, err)
	err = json.Unmarshal(dockerBytes, &roundtrip)
	require.NoError(t, err)

	assert.Equal(t, original.Config.Digest, roundtrip.Config.Digest)
	assert.Equal(t, original.Config.Size, roundtrip.Config.Size)
	assert.Len(t, roundtrip.Layers, len(original.Layers))
	assert.Equal(t, original.Layers[0].Digest, roundtrip.Layers[0].Digest)
}
