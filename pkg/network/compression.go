package network

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"

	"freightliner/pkg/helper/errors"
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

// CompressionOptions configures compression behavior
type CompressionOptions struct {
	// Type is the compression algorithm to use
	Type CompressionType

	// Level controls the compression level
	Level CompressionLevel

	// MinSize is the minimum size in bytes to apply compression
	MinSize int
}

// DefaultCompressionOptions returns sensible default options
func DefaultCompressionOptions() CompressionOptions {
	return CompressionOptions{
		Type:    GzipCompression,
		Level:   DefaultCompression,
		MinSize: 1024, // Don't compress tiny files
	}
}

// CompressorOptions configures a compressor
type CompressorOptions struct {
	// Type is the compression algorithm to use
	Type CompressionType

	// Level controls the compression level
	Level CompressionLevel
}

// DefaultCompressorOptions returns sensible default options
func DefaultCompressorOptions() CompressorOptions {
	return CompressorOptions{
		Type:  GzipCompression,
		Level: DefaultCompression,
	}
}

// NewCompressingWriter creates a new writer that compresses data
func NewCompressingWriter(w io.Writer, opts CompressorOptions) (io.WriteCloser, error) {
	if w == nil {
		return nil, errors.InvalidInputf("writer cannot be nil")
	}

	switch opts.Type {
	case NoCompression:
		// Just pass through
		return &nopWriteCloser{Writer: w}, nil
	case GzipCompression:
		return gzip.NewWriterLevel(w, int(opts.Level))
	case ZlibCompression:
		return zlib.NewWriterLevel(w, int(opts.Level))
	default:
		return nil, errors.InvalidInputf("unsupported compression type: %s", opts.Type)
	}
}

// NewDecompressingReader creates a new reader that decompresses data
func NewDecompressingReader(r io.Reader, compType CompressionType) (io.ReadCloser, error) {
	if r == nil {
		return nil, errors.InvalidInputf("reader cannot be nil")
	}

	switch compType {
	case NoCompression:
		// Just pass through
		return io.NopCloser(r), nil
	case GzipCompression:
		return gzip.NewReader(r)
	case ZlibCompression:
		return zlib.NewReader(r)
	default:
		return nil, errors.InvalidInputf("unsupported compression type: %s", compType)
	}
}

// Compress compresses the given data using the specified options
func Compress(data []byte, opts CompressionOptions) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.InvalidInputf("data cannot be empty")
	}

	// If no compression or data is smaller than minsize, just return the data
	if opts.Type == NoCompression || len(data) < opts.MinSize {
		return data, nil
	}

	var buf bytes.Buffer
	w, err := NewCompressingWriter(&buf, CompressorOptions{
		Type:  opts.Type,
		Level: opts.Level,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create compressing writer")
	}

	if _, err := w.Write(data); err != nil {
		return nil, errors.Wrap(err, "failed to compress data")
	}

	if err := w.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to finalize compression")
	}

	return buf.Bytes(), nil
}

// Decompress decompresses the given data using the specified compression type
func Decompress(data []byte, compType CompressionType) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.InvalidInputf("data cannot be empty")
	}

	// If no compression, just return the data
	if compType == NoCompression {
		return data, nil
	}

	r, err := NewDecompressingReader(bytes.NewReader(data), compType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create decompressing reader")
	}
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			// Log the close error but don't override the main error
		}
	}()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, errors.Wrap(err, "failed to decompress data")
	}

	return buf.Bytes(), nil
}

// ParseCompressionType parses a string into a CompressionType
func ParseCompressionType(s string) (CompressionType, error) {
	switch s {
	case "none":
		return NoCompression, nil
	case "gzip":
		return GzipCompression, nil
	case "zlib":
		return ZlibCompression, nil
	default:
		return "", errors.InvalidInputf("unsupported compression type: %s", s)
	}
}

// String implements the Stringer interface for CompressionType
func (c CompressionType) String() string {
	return string(c)
}

// nopWriteCloser is an io.WriteCloser that just wraps an io.Writer
type nopWriteCloser struct {
	io.Writer
}

// Close implements io.Closer
func (nopWriteCloser) Close() error {
	return nil
}
