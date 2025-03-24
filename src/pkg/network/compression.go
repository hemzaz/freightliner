package network

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
)

// CompressionType represents the type of compression to use
type CompressionType string

const (
	// NoCompression indicates no compression should be used
	NoCompression CompressionType = "none"

	// GzipCompression indicates gzip compression should be used
	GzipCompression CompressionType = "gzip"

	// ZlibCompression indicates zlib compression should be used
	ZlibCompression CompressionType = "zlib"
)

// CompressionLevel controls the tradeoff between speed and compression ratio
type CompressionLevel int

const (
	// DefaultCompression is the default compression level
	DefaultCompression CompressionLevel = -1

	// BestSpeed prioritizes speed over compression ratio
	BestSpeed CompressionLevel = 1

	// BestCompression prioritizes compression ratio over speed
	BestCompression CompressionLevel = 9
)

// CompressionOptions configures the compression behavior
type CompressionOptions struct {
	// Type is the compression algorithm to use
	Type CompressionType

	// Level is the compression level to use
	Level CompressionLevel

	// MinSize is the minimum size in bytes for a payload to be compressed
	// This avoids compressing small payloads where overhead may exceed savings
	MinSize int
}

// DefaultCompressionOptions returns sensible default compression options
func DefaultCompressionOptions() CompressionOptions {
	return CompressionOptions{
		Type:    GzipCompression,
		Level:   DefaultCompression,
		MinSize: 1024, // Only compress data larger than 1KB
	}
}

// Compress compresses the input data according to the compression options
func Compress(data []byte, opts CompressionOptions) ([]byte, error) {
	// Don't compress if the compression type is none or the data is too small
	if opts.Type == NoCompression || len(data) < opts.MinSize {
		return data, nil
	}

	var buf bytes.Buffer
	var compressor io.WriteCloser
	var err error

	// Create the appropriate compressor based on the type
	switch opts.Type {
	case GzipCompression:
		if opts.Level == DefaultCompression {
			compressor = gzip.NewWriter(&buf)
		} else {
			compressor, err = gzip.NewWriterLevel(&buf, int(opts.Level))
			if err != nil {
				return nil, fmt.Errorf("failed to create gzip compressor: %w", err)
			}
		}
	case ZlibCompression:
		compressor, err = zlib.NewWriterLevel(&buf, int(opts.Level))
		if err != nil {
			return nil, fmt.Errorf("failed to create zlib compressor: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", opts.Type)
	}

	// Write the data to the compressor
	if _, err := compressor.Write(data); err != nil {
		compressor.Close()
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	// Close the compressor to flush any remaining data
	if err := compressor.Close(); err != nil {
		return nil, fmt.Errorf("failed to close compressor: %w", err)
	}

	// If compression didn't reduce the size, return the original data
	if buf.Len() >= len(data) {
		return data, nil
	}

	return buf.Bytes(), nil
}

// Decompress decompresses the input data according to the compression type
func Decompress(data []byte, compressionType CompressionType) ([]byte, error) {
	// If no compression, return the original data
	if compressionType == NoCompression {
		return data, nil
	}

	var decompressor io.ReadCloser
	var err error

	// Create the appropriate decompressor based on the type
	switch compressionType {
	case GzipCompression:
		decompressor, err = gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip decompressor: %w", err)
		}
	case ZlibCompression:
		decompressor, err = zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create zlib decompressor: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", compressionType)
	}
	defer decompressor.Close()

	// Read all the decompressed data
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, decompressor); err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return buf.Bytes(), nil
}