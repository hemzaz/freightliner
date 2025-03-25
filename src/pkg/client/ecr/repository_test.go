package ecr

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"freightliner/src/pkg/client/common"
)

// Mock for ECR API for repository operations
type mockRepositoryECRAPI struct {
	mock.Mock
}

func (m *mockRepositoryECRAPI) DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.DescribeImagesOutput), args.Error(1)
}

func (m *mockRepositoryECRAPI) BatchGetImage(ctx context.Context, params *ecr.BatchGetImageInput, optFns ...func(*ecr.Options)) (*ecr.BatchGetImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.BatchGetImageOutput), args.Error(1)
}

func (m *mockRepositoryECRAPI) PutImage(ctx context.Context, params *ecr.PutImageInput, optFns ...func(*ecr.Options)) (*ecr.PutImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.PutImageOutput), args.Error(1)
}

func (m *mockRepositoryECRAPI) BatchDeleteImage(ctx context.Context, params *ecr.BatchDeleteImageInput, optFns ...func(*ecr.Options)) (*ecr.BatchDeleteImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.BatchDeleteImageOutput), args.Error(1)
}

func TestRepositoryListTags(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockRepositoryECRAPI)
		expected    []string
		expectedErr bool
	}{
		{
			name: "Successful list",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("DescribeImages", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.DescribeImagesOutput{
						ImageDetails: []types.ImageDetail{
							{ImageTags: []string{"tag1", "latest"}},
							{ImageTags: []string{"tag2"}},
							{ImageTags: nil}, // Image without tags (digest only)
						},
					}, nil)
			},
			expected:    []string{"tag1", "latest", "tag2"},
			expectedErr: false,
		},
		{
			name: "Empty list",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("DescribeImages", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.DescribeImagesOutput{
						ImageDetails: []types.ImageDetail{},
					}, nil)
			},
			expected:    []string{},
			expectedErr: false,
		},
		{
			name: "Only untagged images",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("DescribeImages", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.DescribeImagesOutput{
						ImageDetails: []types.ImageDetail{
							{ImageTags: nil},
							{ImageTags: nil},
						},
					}, nil)
			},
			expected:    []string{},
			expectedErr: false,
		},
		{
			name: "API error",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("DescribeImages", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.DescribeImagesOutput{}, errors.New("API error"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockECR := &mockRepositoryECRAPI{}
			tc.mockSetup(mockECR)

			repo := &Repository{
				name:      "test-repo",
				client:    mockECR,
				registry:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			}

			tags, err := repo.ListTags()
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expected, tags)
			}

			mockECR.AssertExpectations(t)
		})
	}
}

func TestRepositoryGetManifest(t *testing.T) {
	manifestBytes := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json"}`)
	
	tests := []struct {
		name           string
		tag            string
		mockSetup      func(*mockRepositoryECRAPI)
		expectedMediaType string
		expectedErr    bool
		expectedErrType error
	}{
		{
			name: "Successful get by tag",
			tag:  "latest",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("BatchGetImage", mock.Anything, &ecr.BatchGetImageInput{
					RepositoryName: aws.String("test-repo"),
					ImageIds: []types.ImageIdentifier{
						{ImageTag: aws.String("latest")},
					},
					AcceptedMediaTypes: []string{
						"application/vnd.docker.distribution.manifest.v2+json",
						"application/vnd.oci.image.manifest.v1+json",
						"application/vnd.docker.distribution.manifest.list.v2+json",
						"application/vnd.oci.image.index.v1+json",
					},
				}, mock.Anything).
					Return(&ecr.BatchGetImageOutput{
						Images: []types.Image{
							{
								ImageId: &types.ImageIdentifier{
									ImageTag: aws.String("latest"),
								},
								ImageManifest: aws.String(string(manifestBytes)),
							},
						},
					}, nil)
			},
			expectedMediaType: "application/vnd.docker.distribution.manifest.v2+json",
			expectedErr:       false,
		},
		{
			name: "Tag not found",
			tag:  "non-existent",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("BatchGetImage", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.BatchGetImageOutput{
						Images: []types.Image{},
						Failures: []types.ImageFailure{
							{
								FailureCode: aws.String("ImageNotFound"),
								FailureReason: aws.String("Requested image not found"),
								ImageId: &types.ImageIdentifier{
									ImageTag: aws.String("non-existent"),
								},
							},
						},
					}, nil)
			},
			expectedMediaType: "",
			expectedErr:       true,
			expectedErrType:   common.ErrNotFound,
		},
		{
			name: "API error",
			tag:  "latest",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("BatchGetImage", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.BatchGetImageOutput{}, errors.New("API error"))
			},
			expectedMediaType: "",
			expectedErr:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockECR := &mockRepositoryECRAPI{}
			tc.mockSetup(mockECR)

			repo := &Repository{
				name:     "test-repo",
				client:   mockECR,
				registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			}

			manifest, mediaType, err := repo.GetManifest(tc.tag)
			if tc.expectedErr {
				assert.Error(t, err)
				if tc.expectedErrType != nil {
					assert.True(t, errors.Is(err, tc.expectedErrType))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, manifestBytes, manifest)
				assert.Equal(t, tc.expectedMediaType, mediaType)
			}

			mockECR.AssertExpectations(t)
		})
	}
}

func TestRepositoryPutManifest(t *testing.T) {
	manifestBytes := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json"}`)
	
	tests := []struct {
		name        string
		tag         string
		manifest    []byte
		mediaType   string
		mockSetup   func(*mockRepositoryECRAPI)
		expectedErr bool
	}{
		{
			name:      "Successful put",
			tag:       "latest",
			manifest:  manifestBytes,
			mediaType: "application/vnd.docker.distribution.manifest.v2+json",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("PutImage", mock.Anything, &ecr.PutImageInput{
					RepositoryName: aws.String("test-repo"),
					ImageTag:       aws.String("latest"),
					ImageManifest:  aws.String(string(manifestBytes)),
				}, mock.Anything).
					Return(&ecr.PutImageOutput{
						Image: &types.Image{
							ImageId: &types.ImageIdentifier{
								ImageTag: aws.String("latest"),
							},
						},
					}, nil)
			},
			expectedErr: false,
		},
		{
			name:      "API error",
			tag:       "latest",
			manifest:  manifestBytes,
			mediaType: "application/vnd.docker.distribution.manifest.v2+json",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("PutImage", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.PutImageOutput{}, errors.New("API error"))
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockECR := &mockRepositoryECRAPI{}
			tc.mockSetup(mockECR)

			repo := &Repository{
				name:     "test-repo",
				client:   mockECR,
				registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			}

			err := repo.PutManifest(tc.tag, tc.manifest, tc.mediaType)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockECR.AssertExpectations(t)
		})
	}
}

func TestRepositoryDeleteManifest(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		mockSetup   func(*mockRepositoryECRAPI)
		expectedErr bool
		expectedErrType error
	}{
		{
			name: "Successful delete",
			tag:  "latest",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("BatchDeleteImage", mock.Anything, &ecr.BatchDeleteImageInput{
					RepositoryName: aws.String("test-repo"),
					ImageIds: []types.ImageIdentifier{
						{ImageTag: aws.String("latest")},
					},
				}, mock.Anything).
					Return(&ecr.BatchDeleteImageOutput{
						ImageIds: []types.ImageIdentifier{
							{ImageTag: aws.String("latest")},
						},
					}, nil)
			},
			expectedErr: false,
		},
		{
			name: "Tag not found",
			tag:  "non-existent",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("BatchDeleteImage", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.BatchDeleteImageOutput{
						Failures: []types.ImageFailure{
							{
								FailureCode: aws.String("ImageNotFound"),
								FailureReason: aws.String("Requested image not found"),
								ImageId: &types.ImageIdentifier{
									ImageTag: aws.String("non-existent"),
								},
							},
						},
					}, nil)
			},
			expectedErr: true,
			expectedErrType: common.ErrNotFound,
		},
		{
			name: "API error",
			tag:  "latest",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("BatchDeleteImage", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.BatchDeleteImageOutput{}, errors.New("API error"))
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockECR := &mockRepositoryECRAPI{}
			tc.mockSetup(mockECR)

			repo := &Repository{
				name:     "test-repo",
				client:   mockECR,
				registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			}

			err := repo.DeleteManifest(tc.tag)
			if tc.expectedErr {
				assert.Error(t, err)
				if tc.expectedErrType != nil {
					assert.True(t, errors.Is(err, tc.expectedErrType))
				}
			} else {
				assert.NoError(t, err)
			}

			mockECR.AssertExpectations(t)
		})
	}
}