package copy

import (
	"bytes"
	"context"
	"io"
	"testing"

	"freightliner/pkg/helper/log"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// TestBlobLayerComprehensive tests all blobLayer methods
func TestBlobLayerComprehensive(t *testing.T) {
	testData := []byte("comprehensive test data for blob layer")
	hash, _ := v1.NewHash("sha256:comprehensive123")

	layer := &blobLayer{
		digestHash: hash,
		data:       testData,
	}

	t.Run("Digest", func(t *testing.T) {
		digest, err := layer.Digest()
		if err != nil {
			t.Fatalf("Digest() failed: %v", err)
		}
		if digest != hash {
			t.Errorf("Expected digest %v, got %v", hash, digest)
		}
	})

	t.Run("DiffID", func(t *testing.T) {
		diffID, err := layer.DiffID()
		if err != nil {
			t.Fatalf("DiffID() failed: %v", err)
		}
		if diffID != hash {
			t.Errorf("Expected diffID %v, got %v", hash, diffID)
		}
	})

	t.Run("Size", func(t *testing.T) {
		size, err := layer.Size()
		if err != nil {
			t.Fatalf("Size() failed: %v", err)
		}
		if size != int64(len(testData)) {
			t.Errorf("Expected size %d, got %d", len(testData), size)
		}
	})

	t.Run("MediaType", func(t *testing.T) {
		mt, err := layer.MediaType()
		if err != nil {
			t.Fatalf("MediaType() failed: %v", err)
		}
		if mt != types.DockerLayer {
			t.Errorf("Expected media type %v, got %v", types.DockerLayer, mt)
		}
	})

	t.Run("Compressed", func(t *testing.T) {
		compressed, err := layer.Compressed()
		if err != nil {
			t.Fatalf("Compressed() failed: %v", err)
		}
		defer compressed.Close()

		data, err := io.ReadAll(compressed)
		if err != nil {
			t.Fatalf("Reading compressed data failed: %v", err)
		}
		if !bytes.Equal(data, testData) {
			t.Error("Compressed data doesn't match original")
		}
	})

	t.Run("Uncompressed", func(t *testing.T) {
		uncompressed, err := layer.Uncompressed()
		if err != nil {
			t.Fatalf("Uncompressed() failed: %v", err)
		}
		defer uncompressed.Close()

		data, err := io.ReadAll(uncompressed)
		if err != nil {
			t.Fatalf("Reading uncompressed data failed: %v", err)
		}
		if !bytes.Equal(data, testData) {
			t.Error("Uncompressed data doesn't match original")
		}
	})
}

// TestStreamingBlobLayerComprehensive tests all streamingBlobLayer methods
func TestStreamingBlobLayerComprehensive(t *testing.T) {
	testData := []byte("streaming comprehensive test data")
	hash, _ := v1.NewHash("sha256:streaming456")

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	layer := &streamingBlobLayer{
		digestHash: hash,
		reader:     bytes.NewReader(testData),
		bufferMgr:  copier.bufferMgr,
		cachedSize: int64(len(testData)),
	}

	t.Run("Digest", func(t *testing.T) {
		digest, err := layer.Digest()
		if err != nil {
			t.Fatalf("Digest() failed: %v", err)
		}
		if digest != hash {
			t.Errorf("Expected digest %v, got %v", hash, digest)
		}
	})

	t.Run("DiffID", func(t *testing.T) {
		diffID, err := layer.DiffID()
		if err != nil {
			t.Fatalf("DiffID() failed: %v", err)
		}
		if diffID != hash {
			t.Errorf("Expected diffID %v, got %v", hash, diffID)
		}
	})

	t.Run("Size with cache", func(t *testing.T) {
		size, err := layer.Size()
		if err != nil {
			t.Fatalf("Size() failed: %v", err)
		}
		if size != int64(len(testData)) {
			t.Errorf("Expected size %d, got %d", len(testData), size)
		}
	})

	t.Run("Size without cache", func(t *testing.T) {
		layerNoCache := &streamingBlobLayer{
			digestHash: hash,
			reader:     bytes.NewReader(testData),
			bufferMgr:  copier.bufferMgr,
			cachedSize: 0,
		}

		size, err := layerNoCache.Size()
		if err != nil {
			t.Fatalf("Size() failed: %v", err)
		}
		// Should return default size
		if size <= 0 {
			t.Error("Expected positive default size")
		}
	})

	t.Run("MediaType", func(t *testing.T) {
		mt, err := layer.MediaType()
		if err != nil {
			t.Fatalf("MediaType() failed: %v", err)
		}
		if mt != types.DockerLayer {
			t.Errorf("Expected media type %v, got %v", types.DockerLayer, mt)
		}
	})

	t.Run("Compressed", func(t *testing.T) {
		compressed, err := layer.Compressed()
		if err != nil {
			t.Fatalf("Compressed() failed: %v", err)
		}
		defer compressed.Close()

		// Should be able to read
		buf := make([]byte, 10)
		n, _ := compressed.Read(buf)
		if n == 0 {
			t.Error("Expected to read data from compressed stream")
		}
	})

	t.Run("Uncompressed", func(t *testing.T) {
		// Reset reader
		layer.reader = bytes.NewReader(testData)

		uncompressed, err := layer.Uncompressed()
		if err != nil {
			t.Fatalf("Uncompressed() failed: %v", err)
		}
		defer uncompressed.Close()

		// Should be able to read
		buf := make([]byte, 10)
		n, _ := uncompressed.Read(buf)
		if n == 0 {
			t.Error("Expected to read data from uncompressed stream")
		}
	})
}

// TestManifestDescriptorComprehensive tests all manifest descriptor methods
func TestManifestDescriptorComprehensive(t *testing.T) {
	manifestData := []byte(`{"schemaVersion": 2, "config": {}}`)
	hash, _ := v1.NewHash("sha256:manifest789")

	tests := []struct {
		name      string
		mediaType types.MediaType
	}{
		{"Docker V2 Schema 2", types.DockerManifestSchema2},
		{"Docker V2 Schema 1", types.DockerManifestSchema1},
		{"OCI Manifest", types.OCIManifestSchema1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := &manifestDescriptor{
				mediaType: tt.mediaType,
				data:      manifestData,
				hash:      hash,
			}

			// Test MediaType
			mt, err := desc.MediaType()
			if err != nil {
				t.Fatalf("MediaType() failed: %v", err)
			}
			if mt != tt.mediaType {
				t.Errorf("Expected media type %v, got %v", tt.mediaType, mt)
			}

			// Test RawManifest
			raw, err := desc.RawManifest()
			if err != nil {
				t.Fatalf("RawManifest() failed: %v", err)
			}
			if !bytes.Equal(raw, manifestData) {
				t.Error("Raw manifest data mismatch")
			}

			// Test Digest
			digest, err := desc.Digest()
			if err != nil {
				t.Fatalf("Digest() failed: %v", err)
			}
			if digest != hash {
				t.Errorf("Expected digest %v, got %v", hash, digest)
			}
		})
	}
}

// TestOptimizedReadCloserComprehensive tests optimized read closer
func TestOptimizedReadCloserComprehensive(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	testData := []byte("comprehensive read closer test data")

	t.Run("Without buffer", func(t *testing.T) {
		reader := bytes.NewReader(testData)
		orc := &optimizedReadCloser{
			reader:    reader,
			bufferMgr: copier.bufferMgr,
			buffer:    nil,
		}

		// Read data
		buf := make([]byte, len(testData))
		n, err := orc.Read(buf)
		if err != nil && err != io.EOF {
			t.Fatalf("Read() failed: %v", err)
		}
		if n != len(testData) {
			t.Errorf("Expected to read %d bytes, got %d", len(testData), n)
		}

		// Close
		err = orc.Close()
		if err != nil {
			t.Fatalf("Close() failed: %v", err)
		}
	})

	t.Run("With buffer", func(t *testing.T) {
		reader := bytes.NewReader(testData)
		buffer := copier.bufferMgr.GetOptimalBuffer(1024, "test")

		orc := &optimizedReadCloser{
			reader:    reader,
			bufferMgr: copier.bufferMgr,
			buffer:    buffer,
		}

		// Close should release buffer
		err := orc.Close()
		if err != nil {
			t.Fatalf("Close() failed: %v", err)
		}

		if orc.buffer != nil {
			t.Error("Expected buffer to be nil after close")
		}
	})

	t.Run("With closable reader", func(t *testing.T) {
		// Create a reader that implements Close
		closableData := io.NopCloser(bytes.NewReader(testData))

		orc := &optimizedReadCloser{
			reader:    closableData,
			bufferMgr: copier.bufferMgr,
		}

		err := orc.Close()
		if err != nil {
			t.Fatalf("Close() with closable reader failed: %v", err)
		}
	})

	t.Run("Multiple close calls", func(t *testing.T) {
		reader := bytes.NewReader(testData)
		orc := &optimizedReadCloser{
			reader:    reader,
			bufferMgr: copier.bufferMgr,
		}

		// First close
		err := orc.Close()
		if err != nil {
			t.Fatalf("First close failed: %v", err)
		}

		// Second close should not panic
		err = orc.Close()
		if err != nil {
			t.Fatalf("Second close failed: %v", err)
		}
	})
}

// TestShouldCompressComprehensive tests compression decision logic
func TestShouldCompressComprehensive(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	tests := []struct {
		name     string
		size     int64
		expected bool
	}{
		{"zero size", 0, false},
		{"1 byte", 1, false},
		{"512 bytes", 512, false},
		{"1023 bytes", 1023, false},
		{"1024 bytes (threshold)", 1024, false},
		{"1025 bytes", 1025, true},
		{"10 KB", 10 * 1024, true},
		{"1 MB", 1024 * 1024, true},
		{"100 MB", 100 * 1024 * 1024, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := copier.shouldCompress(tt.size)
			if result != tt.expected {
				t.Errorf("shouldCompress(%d) = %v, expected %v", tt.size, result, tt.expected)
			}
		})
	}
}

// TestCopyBlobDeprecated tests the deprecated method returns error
func TestCopyBlobDeprecatedComprehensive(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	_, err := copier.copyBlob(ctx, "src", "dest", "gzip", false)
	if err == nil {
		t.Error("Expected error from deprecated method")
	}

	errorMsg := err.Error()
	if errorMsg != "copyBlob is deprecated, use transferBlob instead" {
		t.Errorf("Unexpected error message: %s", errorMsg)
	}
}

// TestEncryptBlobPassthrough tests encryption passthrough
func TestEncryptBlobPassthrough(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	testData := []byte("encryption test data")
	reader := io.NopCloser(bytes.NewReader(testData))

	result, err := copier.encryptBlob(ctx, reader, "test-registry")
	if err != nil {
		t.Fatalf("encryptBlob() failed: %v", err)
	}

	// Should pass through when no encryption manager
	if result != reader {
		t.Error("Expected reader to pass through")
	}

	// Verify we can still read from it
	data, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read from result: %v", err)
	}

	if !bytes.Equal(data, testData) {
		t.Error("Data mismatch after passthrough")
	}
}

// TestProcessManifestReturnsEmpty tests process manifest stub
func TestProcessManifestReturnsEmpty(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	result, err := copier.processManifest(ctx, nil, nil, nil, nil, nil, false, nil)
	if err != nil {
		t.Fatalf("processManifest() failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d bytes", len(result))
	}
}
