package gcr

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/assert"
)

func createTestClientExt() *Client {
	client, _ := NewClient(ClientOptions{
		Project:  "test-project",
		Location: "us",
		Logger:   log.NewBasicLogger(log.InfoLevel),
	})
	return client
}

func TestMockRemoteImage_Layers(t *testing.T) {
	img := mockRemoteImage{
		manifestBytes: []byte(`{"schemaVersion":2}`),
		mediaType:     types.DockerManifestSchema2,
	}

	layers, err := img.Layers()
	assert.NoError(t, err)
	assert.NotNil(t, layers)
	assert.Len(t, layers, 0) // Empty for mock
}

func TestMockRemoteImage_MediaType(t *testing.T) {
	expectedType := types.DockerManifestSchema2
	img := mockRemoteImage{
		manifestBytes: []byte(`{"schemaVersion":2}`),
		mediaType:     expectedType,
	}

	mediaType, err := img.MediaType()
	assert.NoError(t, err)
	assert.Equal(t, expectedType, mediaType)
}

func TestMockRemoteImage_Size(t *testing.T) {
	manifestBytes := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json"}`)
	img := mockRemoteImage{
		manifestBytes: manifestBytes,
		mediaType:     types.DockerManifestSchema2,
	}

	size, err := img.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(manifestBytes)), size)
}

func TestMockRemoteImage_ConfigName(t *testing.T) {
	manifestBytes := []byte(`{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size": 7023,
			"digest": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		}
	}`)

	img := mockRemoteImage{
		manifestBytes: manifestBytes,
		mediaType:     types.DockerManifestSchema2,
	}

	configName, err := img.ConfigName()
	assert.NoError(t, err)
	assert.NotEmpty(t, configName.String())
}

func TestMockRemoteImage_ConfigFile(t *testing.T) {
	img := mockRemoteImage{
		manifestBytes: []byte(`{"schemaVersion":2}`),
		mediaType:     types.DockerManifestSchema2,
	}

	config, err := img.ConfigFile()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "amd64", config.Architecture)
	assert.Equal(t, "linux", config.OS)
}

func TestMockRemoteImage_RawConfigFile(t *testing.T) {
	img := mockRemoteImage{
		manifestBytes: []byte(`{"schemaVersion":2}`),
		mediaType:     types.DockerManifestSchema2,
	}

	rawConfig, err := img.RawConfigFile()
	assert.NoError(t, err)
	assert.NotNil(t, rawConfig)

	// Verify it's valid JSON
	var config v1.ConfigFile
	err = json.Unmarshal(rawConfig, &config)
	assert.NoError(t, err)
}

func TestMockRemoteImage_Digest(t *testing.T) {
	manifestBytes := []byte(`{"schemaVersion":2}`)
	img := mockRemoteImage{
		manifestBytes: manifestBytes,
		mediaType:     types.DockerManifestSchema2,
	}

	digest, err := img.Digest()
	assert.NoError(t, err)
	assert.NotEmpty(t, digest.String())
}

func TestMockRemoteImage_Manifest(t *testing.T) {
	manifestBytes := []byte(`{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size": 7023,
			"digest": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		},
		"layers": []
	}`)

	img := mockRemoteImage{
		manifestBytes: manifestBytes,
		mediaType:     types.DockerManifestSchema2,
	}

	manifest, err := img.Manifest()
	assert.NoError(t, err)
	assert.NotNil(t, manifest)
	assert.Equal(t, int64(2), manifest.SchemaVersion)
}

func TestMockRemoteImage_RawManifest(t *testing.T) {
	manifestBytes := []byte(`{"schemaVersion":2}`)
	img := mockRemoteImage{
		manifestBytes: manifestBytes,
		mediaType:     types.DockerManifestSchema2,
	}

	rawManifest, err := img.RawManifest()
	assert.NoError(t, err)
	assert.Equal(t, manifestBytes, rawManifest)
}

func TestMockRemoteImage_LayerByDigest(t *testing.T) {
	img := mockRemoteImage{
		manifestBytes: []byte(`{"schemaVersion":2}`),
		mediaType:     types.DockerManifestSchema2,
	}

	hash, _ := v1.NewHash("sha256:abc123")
	layer, err := img.LayerByDigest(hash)
	assert.Error(t, err)
	assert.Nil(t, layer)
}

func TestMockRemoteImage_LayerByDiffID(t *testing.T) {
	img := mockRemoteImage{
		manifestBytes: []byte(`{"schemaVersion":2}`),
		mediaType:     types.DockerManifestSchema2,
	}

	hash, _ := v1.NewHash("sha256:abc123")
	layer, err := img.LayerByDiffID(hash)
	assert.Error(t, err)
	assert.Nil(t, layer)
}

func TestBytes2Reader(t *testing.T) {
	data := []byte("test data")
	reader := bytes2Reader(data)

	assert.NotNil(t, reader)

	// Verify we can read the data
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, reader)
	assert.NoError(t, err)
	assert.Equal(t, data, buf.Bytes())
}

func TestRepositoryExtended_PutManifestValidation(t *testing.T) {
	client := createTestClientExt()

	ctx := context.Background()
	repoInterface, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	repo, ok := repoInterface.(*Repository)
	assert.True(t, ok)

	// Test with nil manifest
	err = repo.PutManifest(ctx, "v1.0.0", nil)
	assert.Error(t, err)

	// Test with empty tag
	manifest := &interfaces.Manifest{
		Content:   []byte(`{"schemaVersion":2}`),
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:    "sha256:abc123",
	}
	err = repo.PutManifest(ctx, "", manifest)
	assert.Error(t, err)
}

func TestRepositoryExtended_GetLayerReaderValidation(t *testing.T) {
	client := createTestClientExt()

	ctx := context.Background()
	repoInterface, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	repo, ok := repoInterface.(*Repository)
	assert.True(t, ok)

	// Test with empty digest
	reader, err := repo.GetLayerReader(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, reader)
}

func TestRepositoryExtended_DeleteImageValidation(t *testing.T) {
	client := createTestClientExt()

	ctx := context.Background()
	repoInterface, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	repo, ok := repoInterface.(*Repository)
	assert.True(t, ok)

	// Test with empty tag
	err = repo.DeleteImage(ctx, "")
	assert.Error(t, err)
}

func TestRepositoryExtended_GetImageValidation(t *testing.T) {
	client := createTestClientExt()

	ctx := context.Background()
	repoInterface, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	repo, ok := repoInterface.(*Repository)
	assert.True(t, ok)

	// Test with empty tag
	_, err = repo.GetImage(ctx, "")
	assert.Error(t, err)
}

func TestRepositoryExtended_PutImageValidation(t *testing.T) {
	client := createTestClientExt()

	ctx := context.Background()
	repoInterface, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	repo, ok := repoInterface.(*Repository)
	assert.True(t, ok)

	// Test with empty tag
	err = repo.PutImage(ctx, "", nil)
	assert.Error(t, err)

	// Test with nil image
	err = repo.PutImage(ctx, "v1.0.0", nil)
	assert.Error(t, err)
}

func TestRepositoryExtended_PutLayerValidation(t *testing.T) {
	client := createTestClientExt()

	ctx := context.Background()
	repoInterface, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	repo, ok := repoInterface.(*Repository)
	assert.True(t, ok)

	// Test with nil layer
	err = repo.PutLayer(ctx, nil)
	assert.Error(t, err)
}

func TestRepositoryExtended_GetImageReferenceValidation(t *testing.T) {
	client := createTestClientExt()

	ctx := context.Background()
	repoInterface, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	repo, ok := repoInterface.(*Repository)
	assert.True(t, ok)

	// Test with empty tag
	_, err = repo.GetImageReference("")
	assert.Error(t, err)
}

func TestRepositoryExtended_GetRemoteOptions(t *testing.T) {
	client := createTestClientExt()

	ctx := context.Background()
	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	opts, err := repo.GetRemoteOptions()
	assert.NoError(t, err)
	assert.NotNil(t, opts)
	assert.Len(t, opts, 1)
}
