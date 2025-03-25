package gcr

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"freightliner/src/pkg/client/common"
)

// Mock for Google Tags
type mockGoogleTags struct {
	mock.Mock
}

func (m *mockGoogleTags) List(ctx context.Context, repo name.Repository, opts ...google.Option) ([]string, error) {
	args := m.Called(ctx, repo, opts)
	return args.Get(0).([]string), args.Error(1)
}

// Mock for Remote operations
type mockRemote struct {
	mock.Mock
}

func (m *mockRemote) Get(ref name.Reference, opts ...remote.Option) (*remote.Descriptor, error) {
	args := m.Called(ref, opts)
	if d, ok := args.Get(0).(*remote.Descriptor); ok {
		return d, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRemote) Head(ref name.Reference, opts ...remote.Option) (*v1.Descriptor, error) {
	args := m.Called(ref, opts)
	if d, ok := args.Get(0).(*v1.Descriptor); ok {
		return d, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRemote) Image(ref name.Reference, opts ...remote.Option) (v1.Image, error) {
	args := m.Called(ref, opts)
	if img, ok := args.Get(0).(v1.Image); ok {
		return img, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRemote) Write(ref name.Reference, img v1.Image, opts ...remote.Option) error {
	args := m.Called(ref, img, opts)
	return args.Error(0)
}

func (m *mockRemote) Delete(ref name.Reference, opts ...remote.Option) error {
	args := m.Called(ref, opts)
	return args.Error(0)
}

// Mock for v1.Image
type mockImage struct {
	mock.Mock
}

func (m *mockImage) Layers() ([]v1.Layer, error) {
	args := m.Called()
	return args.Get(0).([]v1.Layer), args.Error(1)
}

func (m *mockImage) MediaType() (types.MediaType, error) {
	args := m.Called()
	return args.Get(0).(types.MediaType), args.Error(1)
}

func (m *mockImage) Size() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockImage) ConfigName() (v1.Hash, error) {
	args := m.Called()
	return args.Get(0).(v1.Hash), args.Error(1)
}

func (m *mockImage) ConfigFile() (*v1.ConfigFile, error) {
	args := m.Called()
	return args.Get(0).(*v1.ConfigFile), args.Error(1)
}

func (m *mockImage) RawConfigFile() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockImage) Digest() (v1.Hash, error) {
	args := m.Called()
	return args.Get(0).(v1.Hash), args.Error(1)
}

func (m *mockImage) Manifest() (*v1.Manifest, error) {
	args := m.Called()
	return args.Get(0).(*v1.Manifest), args.Error(1)
}

func (m *mockImage) RawManifest() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockImage) LayerByDigest(digest v1.Hash) (v1.Layer, error) {
	args := m.Called(digest)
	return args.Get(0).(v1.Layer), args.Error(1)
}

func (m *mockImage) LayerByDiffID(diffID v1.Hash) (v1.Layer, error) {
	args := m.Called(diffID)
	return args.Get(0).(v1.Layer), args.Error(1)
}

func TestRepositoryListTags(t *testing.T) {
	tests := []struct {
		name        string
		repoName    string
		mockSetup   func(*mockGoogleTags)
		expected    []string
		expectedErr bool
	}{
		{
			name:     "Successful list",
			repoName: "project/repo",
			mockSetup: func(mockTags *mockGoogleTags) {
				mockTags.On("List", mock.Anything, mock.Anything, mock.Anything).
					Return([]string{"latest", "v1.0", "v2.0"}, nil)
			},
			expected:    []string{"latest", "v1.0", "v2.0"},
			expectedErr: false,
		},
		{
			name:     "Empty list",
			repoName: "project/repo",
			mockSetup: func(mockTags *mockGoogleTags) {
				mockTags.On("List", mock.Anything, mock.Anything, mock.Anything).
					Return([]string{}, nil)
			},
			expected:    []string{},
			expectedErr: false,
		},
		{
			name:     "API error",
			repoName: "project/repo",
			mockSetup: func(mockTags *mockGoogleTags) {
				mockTags.On("List", mock.Anything, mock.Anything, mock.Anything).
					Return([]string{}, errors.New("API error"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTags := &mockGoogleTags{}
			tc.mockSetup(mockTags)

			reg, _ := name.NewRegistry("gcr.io")
			repo, _ := name.NewRepository("gcr.io/" + tc.repoName)

			repository := &Repository{
				name:        tc.repoName,
				ref:         repo,
				registry:    reg,
				keychain:    authn.DefaultKeychain,
				tagsFunc:    mockTags.List,
			}

			tags, err := repository.ListTags()
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expected, tags)
			}

			mockTags.AssertExpectations(t)
		})
	}
}

func TestRepositoryGetManifest(t *testing.T) {
	manifestBytes := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json"}`)
	
	tests := []struct {
		name              string
		tag               string
		mockSetup         func(*mockRemote, *mockImage)
		expectedManifest  []byte
		expectedMediaType string
		expectedErr       bool
		expectedErrType   error
	}{
		{
			name: "Successful get by tag",
			tag:  "latest",
			mockSetup: func(mockRem *mockRemote, mockImg *mockImage) {
				img := &mockImage{}
				img.On("RawManifest").Return(manifestBytes, nil)
				img.On("MediaType").Return(types.MediaType("application/vnd.docker.distribution.manifest.v2+json"), nil)
				
				mockRem.On("Image", mock.Anything, mock.Anything).Return(img, nil)
			},
			expectedManifest:  manifestBytes,
			expectedMediaType: "application/vnd.docker.distribution.manifest.v2+json",
			expectedErr:       false,
		},
		{
			name: "Image not found",
			tag:  "non-existent",
			mockSetup: func(mockRem *mockRemote, mockImg *mockImage) {
				mockRem.On("Image", mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedManifest:  nil,
			expectedMediaType: "",
			expectedErr:       true,
		},
		{
			name: "Manifest error",
			tag:  "latest",
			mockSetup: func(mockRem *mockRemote, mockImg *mockImage) {
				img := &mockImage{}
				img.On("RawManifest").Return(nil, errors.New("manifest error"))
				
				mockRem.On("Image", mock.Anything, mock.Anything).Return(img, nil)
			},
			expectedManifest:  nil,
			expectedMediaType: "",
			expectedErr:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRem := &mockRemote{}
			mockImg := &mockImage{}
			tc.mockSetup(mockRem, mockImg)

			reg, _ := name.NewRegistry("gcr.io")
			repo, _ := name.NewRepository("gcr.io/project/repo")

			repository := &Repository{
				name:        "project/repo",
				ref:         repo,
				registry:    reg,
				keychain:    authn.DefaultKeychain,
				remoteFunc:  mockRem,
			}

			manifest, mediaType, err := repository.GetManifest(tc.tag)
			if tc.expectedErr {
				assert.Error(t, err)
				if tc.expectedErrType != nil {
					assert.True(t, errors.Is(err, tc.expectedErrType))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedManifest, manifest)
				assert.Equal(t, tc.expectedMediaType, mediaType)
			}
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
		mockSetup   func(*mockRemote)
		expectedErr bool
	}{
		{
			name:      "Successful put",
			tag:       "latest",
			manifest:  manifestBytes,
			mediaType: "application/vnd.docker.distribution.manifest.v2+json",
			mockSetup: func(mockRem *mockRemote) {
				mockRem.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:      "Write error",
			tag:       "latest",
			manifest:  manifestBytes,
			mediaType: "application/vnd.docker.distribution.manifest.v2+json",
			mockSetup: func(mockRem *mockRemote) {
				mockRem.On("Write", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("write error"))
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRem := &mockRemote{}
			tc.mockSetup(mockRem)

			reg, _ := name.NewRegistry("gcr.io")
			repo, _ := name.NewRepository("gcr.io/project/repo")

			repository := &Repository{
				name:       "project/repo",
				ref:        repo,
				registry:   reg,
				keychain:   authn.DefaultKeychain,
				remoteFunc: mockRem,
			}

			err := repository.PutManifest(tc.tag, tc.manifest, tc.mediaType)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRem.AssertExpectations(t)
		})
	}
}

func TestRepositoryDeleteManifest(t *testing.T) {
	tests := []struct {
		name           string
		tag            string
		mockSetup      func(*mockRemote)
		expectedErr    bool
		expectedErrType error
	}{
		{
			name: "Successful delete",
			tag:  "latest",
			mockSetup: func(mockRem *mockRemote) {
				mockRem.On("Delete", mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Delete error",
			tag:  "latest",
			mockSetup: func(mockRem *mockRemote) {
				mockRem.On("Delete", mock.Anything, mock.Anything).
					Return(errors.New("delete error"))
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRem := &mockRemote{}
			tc.mockSetup(mockRem)

			reg, _ := name.NewRegistry("gcr.io")
			repo, _ := name.NewRepository("gcr.io/project/repo")

			repository := &Repository{
				name:       "project/repo",
				ref:        repo,
				registry:   reg,
				keychain:   authn.DefaultKeychain,
				remoteFunc: mockRem,
			}

			err := repository.DeleteManifest(tc.tag)
			if tc.expectedErr {
				assert.Error(t, err)
				if tc.expectedErrType != nil {
					assert.True(t, errors.Is(err, tc.expectedErrType))
				}
			} else {
				assert.NoError(t, err)
			}

			mockRem.AssertExpectations(t)
		})
	}
}

func TestStaticImage(t *testing.T) {
	manifestBytes := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json"}`)
	
	img := newStaticImage(manifestBytes, "application/vnd.docker.distribution.manifest.v2+json")
	
	// Test RawManifest
	rawManifest, err := img.RawManifest()
	assert.NoError(t, err)
	assert.Equal(t, manifestBytes, rawManifest)
	
	// Test MediaType
	mediaType, err := img.MediaType()
	assert.NoError(t, err)
	assert.Equal(t, types.MediaType("application/vnd.docker.distribution.manifest.v2+json"), mediaType)
	
	// Test Digest
	digest, err := img.Digest()
	assert.NoError(t, err)
	assert.NotEmpty(t, digest.String())
	
	// Test Size
	size, err := img.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(manifestBytes)), size)
	
	// Test ConfigFile (this should error in our implementation)
	_, err = img.ConfigFile()
	assert.Error(t, err)
}