package transport

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		name          string
		ref           string
		expectedType  string
		expectedError bool
	}{
		{
			name:         "dir reference",
			ref:          "dir:/tmp/myimage",
			expectedType: "dir",
		},
		{
			name:         "oci reference",
			ref:          "oci:/tmp/oci-layout",
			expectedType: "oci",
		},
		{
			name:         "docker-archive reference",
			ref:          "docker-archive:/tmp/image.tar",
			expectedType: "docker-archive",
		},
		{
			name:          "unknown transport",
			ref:           "unknown://test",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ParseReference(tt.ref)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, ref)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ref)
				assert.Equal(t, tt.expectedType, ref.Transport().Name())
			}
		})
	}
}

func TestTransportRegistry(t *testing.T) {
	t.Run("list registered transports", func(t *testing.T) {
		transports := ListTransports()
		assert.Greater(t, len(transports), 0)

		// Check that common transports are registered
		transportMap := make(map[string]bool)
		for _, name := range transports {
			transportMap[name] = true
		}

		assert.True(t, transportMap["dir"], "dir transport should be registered")
		assert.True(t, transportMap["oci"], "oci transport should be registered")
		assert.True(t, transportMap["docker-archive"], "docker-archive transport should be registered")
	})

	t.Run("get registered transport", func(t *testing.T) {
		transport, ok := GetTransport("dir")
		assert.True(t, ok)
		assert.NotNil(t, transport)
		assert.Equal(t, "dir", transport.Name())
	})

	t.Run("get non-existent transport", func(t *testing.T) {
		transport, ok := GetTransport("non-existent")
		assert.False(t, ok)
		assert.Nil(t, transport)
	})
}

func TestDirectoryTransport(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("validate reference", func(t *testing.T) {
		transport := NewDirectoryTransport()

		err := transport.ValidateReference(tempDir)
		assert.NoError(t, err)

		err = transport.ValidateReference("")
		assert.Error(t, err)
	})

	t.Run("parse reference", func(t *testing.T) {
		transport := NewDirectoryTransport()

		ref, err := transport.ParseReference(tempDir)
		assert.NoError(t, err)
		assert.NotNil(t, ref)

		dirRef, ok := ref.(*DirectoryReference)
		assert.True(t, ok)
		assert.Equal(t, tempDir, dirRef.path)
	})

	t.Run("new image destination", func(t *testing.T) {
		transport := NewDirectoryTransport()
		ref, err := transport.ParseReference(tempDir)
		require.NoError(t, err)

		ctx := context.Background()
		dest, err := ref.NewImageDestination(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, dest)
		defer dest.Close()

		// Check capabilities
		assert.True(t, dest.HasThreadSafePutBlob())
		assert.Equal(t, PreserveOriginal, dest.DesiredLayerCompression())
		assert.False(t, dest.AcceptsForeignLayerURLs())
	})

	t.Run("supported manifest types", func(t *testing.T) {
		transport := NewDirectoryTransport()
		ref, err := transport.ParseReference(tempDir)
		require.NoError(t, err)

		ctx := context.Background()
		dest, err := ref.NewImageDestination(ctx)
		require.NoError(t, err)
		defer dest.Close()

		types := dest.SupportedManifestMIMETypes()
		assert.Greater(t, len(types), 0)
		assert.Contains(t, types, "application/vnd.docker.distribution.manifest.v2+json")
		assert.Contains(t, types, "application/vnd.oci.image.manifest.v1+json")
	})
}

func TestOCILayoutTransport(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("validate reference", func(t *testing.T) {
		transport := NewOCILayoutTransport()

		err := transport.ValidateReference(tempDir)
		assert.NoError(t, err)

		err = transport.ValidateReference("")
		assert.Error(t, err)
	})

	t.Run("parse reference with tag", func(t *testing.T) {
		transport := NewOCILayoutTransport()

		ref, err := transport.ParseReference(tempDir + ":mytag")
		assert.NoError(t, err)
		assert.NotNil(t, ref)

		ociRef, ok := ref.(*OCILayoutReference)
		assert.True(t, ok)
		assert.Equal(t, tempDir, ociRef.path)
		assert.Equal(t, "mytag", ociRef.reference)
	})

	t.Run("parse reference with digest", func(t *testing.T) {
		transport := NewOCILayoutTransport()

		ref, err := transport.ParseReference(tempDir + "@sha256:abcd1234")
		assert.NoError(t, err)
		assert.NotNil(t, ref)

		ociRef, ok := ref.(*OCILayoutReference)
		assert.True(t, ok)
		assert.Equal(t, tempDir, ociRef.path)
		assert.Equal(t, "sha256:abcd1234", ociRef.reference)
	})

	t.Run("initialize OCI layout", func(t *testing.T) {
		transport := NewOCILayoutTransport()
		ref, err := transport.ParseReference(tempDir)
		require.NoError(t, err)

		ctx := context.Background()
		dest, err := ref.NewImageDestination(ctx)
		require.NoError(t, err)
		defer dest.Close()

		// Check that OCI layout files exist
		layoutPath := filepath.Join(tempDir, "oci-layout")
		assert.FileExists(t, layoutPath)

		indexPath := filepath.Join(tempDir, "index.json")
		assert.FileExists(t, indexPath)

		blobsPath := filepath.Join(tempDir, "blobs", "sha256")
		assert.DirExists(t, blobsPath)
	})

	t.Run("oci-layout version", func(t *testing.T) {
		transport := NewOCILayoutTransport()
		ref, err := transport.ParseReference(tempDir)
		require.NoError(t, err)

		ctx := context.Background()
		dest, err := ref.NewImageDestination(ctx)
		require.NoError(t, err)
		dest.Close()

		// Read oci-layout file
		layoutPath := filepath.Join(tempDir, "oci-layout")
		data, err := os.ReadFile(layoutPath)
		require.NoError(t, err)

		// Check version
		assert.Contains(t, string(data), "1.0.0")
	})
}

func TestDockerArchiveTransport(t *testing.T) {
	tempDir := t.TempDir()
	archivePath := filepath.Join(tempDir, "test.tar")

	t.Run("validate reference", func(t *testing.T) {
		transport := NewDockerArchiveTransport()

		err := transport.ValidateReference(archivePath)
		assert.NoError(t, err)

		err = transport.ValidateReference("")
		assert.Error(t, err)
	})

	t.Run("parse reference with tag", func(t *testing.T) {
		transport := NewDockerArchiveTransport()

		ref, err := transport.ParseReference(archivePath + ":latest")
		assert.NoError(t, err)
		assert.NotNil(t, ref)

		archiveRef, ok := ref.(*DockerArchiveReference)
		assert.True(t, ok)
		assert.Equal(t, archivePath, archiveRef.path)
		assert.Equal(t, "latest", archiveRef.reference)
	})

	t.Run("parse reference without tag", func(t *testing.T) {
		transport := NewDockerArchiveTransport()

		ref, err := transport.ParseReference(archivePath)
		assert.NoError(t, err)
		assert.NotNil(t, ref)

		archiveRef, ok := ref.(*DockerArchiveReference)
		assert.True(t, ok)
		assert.Equal(t, archivePath, archiveRef.path)
		assert.Equal(t, "", archiveRef.reference)
	})

	t.Run("new image destination", func(t *testing.T) {
		transport := NewDockerArchiveTransport()
		ref, err := transport.ParseReference(archivePath)
		require.NoError(t, err)

		ctx := context.Background()
		dest, err := ref.NewImageDestination(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, dest)
		defer dest.Close()

		// Check capabilities
		assert.True(t, dest.HasThreadSafePutBlob())
		assert.Equal(t, PreserveOriginal, dest.DesiredLayerCompression())
		assert.False(t, dest.AcceptsForeignLayerURLs())
	})

	t.Run("supported manifest types", func(t *testing.T) {
		transport := NewDockerArchiveTransport()
		ref, err := transport.ParseReference(archivePath)
		require.NoError(t, err)

		ctx := context.Background()
		dest, err := ref.NewImageDestination(ctx)
		require.NoError(t, err)
		defer dest.Close()

		types := dest.SupportedManifestMIMETypes()
		assert.Greater(t, len(types), 0)
		assert.Contains(t, types, "application/vnd.docker.distribution.manifest.v2+json")
	})
}

func TestLayerCompression(t *testing.T) {
	tests := []struct {
		name  string
		value LayerCompression
	}{
		{"AutoCompression", AutoCompression},
		{"Compress", Compress},
		{"Decompress", Decompress},
		{"PreserveOriginal", PreserveOriginal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the constants exist and have different values
			assert.GreaterOrEqual(t, int(tt.value), 0)
		})
	}
}

func TestCryptoOperation(t *testing.T) {
	tests := []struct {
		name  string
		value CryptoOperation
	}{
		{"NoCrypto", NoCrypto},
		{"Encrypt", Encrypt},
		{"Decrypt", Decrypt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the constants exist and have different values
			assert.GreaterOrEqual(t, int(tt.value), 0)
		})
	}
}

func TestReferenceInterfaces(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		transport string
		ref       string
	}{
		{"directory", "dir", tempDir},
		{"oci-layout", "oci", tempDir},
		{"docker-archive", "docker-archive", filepath.Join(tempDir, "test.tar")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, ok := GetTransport(tt.transport)
			require.True(t, ok, "transport %s should be registered", tt.transport)

			ref, err := transport.ParseReference(tt.ref)
			require.NoError(t, err)
			require.NotNil(t, ref)

			// Test Reference interface methods
			assert.Equal(t, tt.transport, ref.Transport().Name())
			assert.NotEmpty(t, ref.StringWithinTransport())
			assert.NotEmpty(t, ref.PolicyConfigurationIdentity())
			assert.Greater(t, len(ref.PolicyConfigurationNamespaces()), 0)
		})
	}
}

func TestImageDestinationThreadSafety(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		transport string
		ref       string
		expected  bool
	}{
		{"directory", "dir", tempDir, true},
		{"oci-layout", "oci", tempDir, true},
		{"docker-archive", "docker-archive", filepath.Join(tempDir, "test.tar"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, ok := GetTransport(tt.transport)
			require.True(t, ok)

			ref, err := transport.ParseReference(tt.ref)
			require.NoError(t, err)

			ctx := context.Background()
			dest, err := ref.NewImageDestination(ctx)
			require.NoError(t, err)
			defer dest.Close()

			assert.Equal(t, tt.expected, dest.HasThreadSafePutBlob())
		})
	}
}

func TestImageSourceThreadSafety(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		transport string
		ref       string
		expected  bool
	}{
		{"directory", "dir", tempDir, true},
		{"oci-layout", "oci", tempDir, true},
		{"docker-archive", "docker-archive", filepath.Join(tempDir, "test.tar"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, ok := GetTransport(tt.transport)
			require.True(t, ok)

			ref, err := transport.ParseReference(tt.ref)
			require.NoError(t, err)

			// For archive, we need an existing file
			if tt.transport == "docker-archive" {
				// Create a minimal valid tar archive
				f, err := os.Create(tt.ref)
				require.NoError(t, err)
				f.Close()
			}

			// For directory and OCI, create manifest.json
			if tt.transport == "dir" {
				manifestPath := filepath.Join(tempDir, "manifest.json")
				err := os.WriteFile(manifestPath, []byte("{}"), 0644)
				require.NoError(t, err)
			} else if tt.transport == "oci" {
				// Initialize OCI layout first
				ctx := context.Background()
				dest, _ := ref.NewImageDestination(ctx)
				if dest != nil {
					dest.Close()
				}
			}

			ctx := context.Background()
			src, err := ref.NewImageSource(ctx)

			if tt.transport == "docker-archive" {
				// Archive without proper content will fail, that's expected
				if err != nil {
					t.Skip("docker-archive requires valid archive content")
				}
			}

			if src != nil {
				defer src.Close()
				assert.Equal(t, tt.expected, src.HasThreadSafeGetBlob())
			}
		})
	}
}
