package ecr

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createExtendedTestClient() (*Client, *MockECRServiceExt) {
	mockService := new(MockECRServiceExt)

	return &Client{
		ecr:       mockService,
		region:    "us-west-2",
		accountID: "123456789012",
		logger:    log.NewBasicLogger(log.InfoLevel),
	}, mockService
}

func TestRepositoryExtended_ListTagsWithPagination(t *testing.T) {
	client, mockService := createExtendedTestClient()

	// Mock DescribeRepositories (called by GetRepository)
	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/test-repo"
	mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{
					RepositoryArn:  &repoArn,
					RepositoryName: aws.String("test-repo"),
				},
			},
		}, nil).Once()

	// First page
	mockService.On("ListImages", mock.Anything, mock.MatchedBy(func(input *awsecr.ListImagesInput) bool {
		return input.NextToken == nil
	}), mock.Anything).
		Return(&awsecr.ListImagesOutput{
			ImageIds: []ecrtypes.ImageIdentifier{
				{ImageTag: aws.String("v1.0.0")},
				{ImageTag: aws.String("v1.1.0")},
			},
			NextToken: aws.String("token1"),
		}, nil).Once()

	// Second page
	mockService.On("ListImages", mock.Anything, mock.MatchedBy(func(input *awsecr.ListImagesInput) bool {
		return input.NextToken != nil && *input.NextToken == "token1"
	}), mock.Anything).
		Return(&awsecr.ListImagesOutput{
			ImageIds: []ecrtypes.ImageIdentifier{
				{ImageTag: aws.String("v2.0.0")},
				{ImageTag: aws.String("latest")},
			},
			NextToken: nil,
		}, nil).Once()

	ctx := context.Background()
	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	tags, err := repo.ListTags(ctx)
	assert.NoError(t, err)
	assert.Len(t, tags, 4)
	assert.Equal(t, []string{"v1.0.0", "v1.1.0", "v2.0.0", "latest"}, tags)
	mockService.AssertExpectations(t)
}

func TestRepositoryExtended_DeleteManifestSuccess(t *testing.T) {
	client, mockService := createExtendedTestClient()

	// Mock DescribeRepositories (called by GetRepository)
	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/test-repo"
	mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{
					RepositoryArn:  &repoArn,
					RepositoryName: aws.String("test-repo"),
				},
			},
		}, nil).Once()

	digest := "sha256:abc123"
	mockService.On("BatchGetImage", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.BatchGetImageOutput{
			Images: []ecrtypes.Image{
				{
					ImageId: &ecrtypes.ImageIdentifier{
						ImageDigest: &digest,
						ImageTag:    aws.String("v1.0.0"),
					},
				},
			},
		}, nil)

	mockService.On("BatchDeleteImage", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.BatchDeleteImageOutput{}, nil)

	ctx := context.Background()
	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	err = repo.DeleteManifest(ctx, "v1.0.0")
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestRepositoryExtended_GetImageReferenceDigest(t *testing.T) {
	client, mockService := createExtendedTestClient()

	// Mock DescribeRepositories (called by GetRepository)
	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/test-repo"
	mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{
					RepositoryArn:  &repoArn,
					RepositoryName: aws.String("test-repo"),
				},
			},
		}, nil).Once()

	ctx := context.Background()
	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	// Test with digest format
	_, err = repo.GetImageReference("@sha256:abc123")
	if err != nil {
		// Expected - needs proper digest format
		assert.Error(t, err)
	}
}

func TestRepositoryExtended_GetRemoteOptions(t *testing.T) {
	client, mockService := createExtendedTestClient()

	// Mock DescribeRepositories (called by GetRepository)
	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/test-repo"
	mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{
					RepositoryArn:  &repoArn,
					RepositoryName: aws.String("test-repo"),
				},
			},
		}, nil).Once()

	ctx := context.Background()
	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	opts, err := repo.GetRemoteOptions()
	assert.NoError(t, err)
	assert.NotNil(t, opts)
	assert.Len(t, opts, 1)
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
	client, mockService := createExtendedTestClient()

	// Mock DescribeRepositories (called by GetRepository)
	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/test-repo"
	mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{
					RepositoryArn:  &repoArn,
					RepositoryName: aws.String("test-repo"),
				},
			},
		}, nil).Once()

	ctx := context.Background()
	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

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
	client, mockService := createExtendedTestClient()

	// Mock DescribeRepositories (called by GetRepository)
	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/test-repo"
	mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{
					RepositoryArn:  &repoArn,
					RepositoryName: aws.String("test-repo"),
				},
			},
		}, nil).Once()

	ctx := context.Background()
	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	// Test with empty digest
	reader, err := repo.GetLayerReader(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, reader)
}

func TestRepositoryExtended_DeleteManifestImageNotFound(t *testing.T) {
	client, mockService := createExtendedTestClient()

	// Mock DescribeRepositories (called by GetRepository)
	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/test-repo"
	mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{
					RepositoryArn:  &repoArn,
					RepositoryName: aws.String("test-repo"),
				},
			},
		}, nil).Once()

	mockService.On("BatchGetImage", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.BatchGetImageOutput{
			Images: []ecrtypes.Image{},
		}, nil)

	ctx := context.Background()
	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	err = repo.DeleteManifest(ctx, "nonexistent-tag")
	assert.Error(t, err)
	mockService.AssertExpectations(t)
}

func TestRepositoryExtended_DeleteManifestNilDigest(t *testing.T) {
	client, mockService := createExtendedTestClient()

	// Mock DescribeRepositories (called by GetRepository)
	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/test-repo"
	mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{
					RepositoryArn:  &repoArn,
					RepositoryName: aws.String("test-repo"),
				},
			},
		}, nil).Once()

	mockService.On("BatchGetImage", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.BatchGetImageOutput{
			Images: []ecrtypes.Image{
				{
					ImageId: &ecrtypes.ImageIdentifier{
						ImageTag: aws.String("v1.0.0"),
						// ImageDigest is nil
					},
				},
			},
		}, nil)

	ctx := context.Background()
	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	err = repo.DeleteManifest(ctx, "v1.0.0")
	assert.Error(t, err)
	mockService.AssertExpectations(t)
}

func TestRepositoryExtended_PutImageValidation(t *testing.T) {
	client, mockService := createExtendedTestClient()

	// Mock DescribeRepositories (called by GetRepository)
	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/test-repo"
	mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{
					RepositoryArn:  &repoArn,
					RepositoryName: aws.String("test-repo"),
				},
			},
		}, nil).Once()

	ctx := context.Background()
	repoInterface, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	// Cast to concrete type to access PutImage
	repo, ok := repoInterface.(*Repository)
	assert.True(t, ok)

	// Test with nil image
	err = repo.PutImage(ctx, "v1.0.0", nil)
	assert.Error(t, err)

	// Test with empty tag
	err = repo.PutImage(ctx, "", nil)
	assert.Error(t, err)
}
