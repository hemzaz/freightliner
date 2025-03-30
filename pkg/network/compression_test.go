package network

import (
	"bytes"
	"testing"
)

func TestCompression(t *testing.T) {
	testCases := []struct {
		name           string
		data           []byte
		opts           CompressionOptions
		shouldCompress bool
	}{
		{
			name: "No compression when type is none",
			data: []byte("test data that is longer than the min size threshold"),
			opts: CompressionOptions{
				Type:    NoCompression,
				Level:   DefaultCompression,
				MinSize: 10,
			},
			shouldCompress: false,
		},
		{
			name: "No compression when data is too small",
			data: []byte("small"),
			opts: CompressionOptions{
				Type:    GzipCompression,
				Level:   DefaultCompression,
				MinSize: 10,
			},
			shouldCompress: false,
		},
		{
			name: "Gzip compression",
			data: bytes.Repeat([]byte("a"), 1000), // Highly compressible data
			opts: CompressionOptions{
				Type:    GzipCompression,
				Level:   DefaultCompression,
				MinSize: 10,
			},
			shouldCompress: true,
		},
		{
			name: "Zlib compression",
			data: bytes.Repeat([]byte("b"), 1000), // Highly compressible data
			opts: CompressionOptions{
				Type:    ZlibCompression,
				Level:   DefaultCompression,
				MinSize: 10,
			},
			shouldCompress: true,
		},
		{
			name: "Best compression level",
			data: bytes.Repeat([]byte("c"), 1000),
			opts: CompressionOptions{
				Type:    GzipCompression,
				Level:   BestCompression,
				MinSize: 10,
			},
			shouldCompress: true,
		},
		{
			name: "Best speed level",
			data: bytes.Repeat([]byte("d"), 1000),
			opts: CompressionOptions{
				Type:    GzipCompression,
				Level:   BestSpeed,
				MinSize: 10,
			},
			shouldCompress: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			compressed, err := Compress(tc.data, tc.opts)
			if err != nil {
				t.Fatalf("Compress() error = %v", err)
			}

			// If we don't expect compression, the output should be the same as input
			if !tc.shouldCompress {
				if !bytes.Equal(compressed, tc.data) {
					t.Errorf("Compress() should return original data, but didn't")
				}
				return
			}

			// For compressible data, check that it's smaller
			if len(compressed) >= len(tc.data) {
				t.Errorf("Compressed data isn't smaller: original=%d, compressed=%d", len(tc.data), len(compressed))
			}

			// Verify we can decompress it back
			decompressed, err := Decompress(compressed, tc.opts.Type)
			if err != nil {
				t.Fatalf("Decompress() error = %v", err)
			}

			// Check if decompressed data matches original
			if !bytes.Equal(decompressed, tc.data) {
				t.Errorf("Decompress() output doesn't match original input")
			}
		})
	}
}

func TestDefaultCompressionOptions(t *testing.T) {
	opts := DefaultCompressionOptions()

	if opts.Type != GzipCompression {
		t.Errorf("Default compression type = %v, want %v", opts.Type, GzipCompression)
	}

	if opts.Level != DefaultCompression {
		t.Errorf("Default compression level = %v, want %v", opts.Level, DefaultCompression)
	}

	if opts.MinSize <= 0 {
		t.Errorf("Default min size should be > 0, got %v", opts.MinSize)
	}
}

func TestDecompress(t *testing.T) {
	// Test decompress with no compression
	original := []byte("test data")
	decompressed, err := Decompress(original, NoCompression)
	if err != nil {
		t.Fatalf("Decompress() error = %v", err)
	}
	if !bytes.Equal(decompressed, original) {
		t.Errorf("Decompress() with NoCompression should return original data")
	}

	// Test with unsupported compression type
	_, err = Decompress([]byte("test"), CompressionType("unknown"))
	if err == nil {
		t.Errorf("Decompress() with unknown compression type should return error")
	}
}
