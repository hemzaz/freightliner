// Package transport provides support for multiple image transport types
// Compatible with Skopeo transport formats: docker://, dir:, oci:, docker-archive:, oci-archive:
package transport

import (
	"context"
	"fmt"
	"io"
)

// Transport represents a container image transport mechanism
type Transport interface {
	// Name returns the transport name (e.g., "docker", "dir", "oci")
	Name() string

	// ValidateReference validates a transport-specific reference
	ValidateReference(ref string) error

	// ParseReference parses a transport-specific reference
	ParseReference(ref string) (Reference, error)
}

// Reference represents a transport-specific image reference
type Reference interface {
	// Transport returns the transport for this reference
	Transport() Transport

	// StringWithinTransport returns the reference string within the transport
	StringWithinTransport() string

	// DockerReference returns equivalent docker:// reference if possible
	DockerReference() string

	// PolicyConfigurationIdentity returns identity for policy configuration
	PolicyConfigurationIdentity() string

	// PolicyConfigurationNamespaces returns namespaces for policy configuration
	PolicyConfigurationNamespaces() []string

	// NewImage returns an Image for this reference
	NewImage(ctx context.Context) (Image, error)

	// NewImageSource returns an ImageSource for reading
	NewImageSource(ctx context.Context) (ImageSource, error)

	// NewImageDestination returns an ImageDestination for writing
	NewImageDestination(ctx context.Context) (ImageDestination, error)

	// DeleteImage deletes the image at this reference
	DeleteImage(ctx context.Context) error
}

// Image represents a container image
type Image interface {
	// Reference returns the image reference
	Reference() Reference

	// Close releases resources
	Close() error

	// Manifest returns the image manifest
	Manifest(ctx context.Context) ([]byte, string, error)

	// Inspect returns image metadata
	Inspect(ctx context.Context) (*ImageInspectInfo, error)

	// LayerInfos returns information about image layers
	LayerInfos() []LayerInfo

	// Size returns the image size in bytes
	Size() (int64, error)
}

// ImageSource represents a source for reading images
type ImageSource interface {
	// Reference returns the image reference
	Reference() Reference

	// Close releases resources
	Close() error

	// GetManifest returns the image manifest
	GetManifest(ctx context.Context, instanceDigest *string) ([]byte, string, error)

	// GetBlob returns a blob (layer or config)
	GetBlob(ctx context.Context, info LayerInfo, cache BlobInfoCache) (io.ReadCloser, int64, error)

	// HasThreadSafeGetBlob returns true if GetBlob can be called concurrently
	HasThreadSafeGetBlob() bool

	// GetSignatures returns image signatures
	GetSignatures(ctx context.Context, instanceDigest *string) ([][]byte, error)

	// LayerInfosForCopy returns layer infos optimized for copying
	LayerInfosForCopy(ctx context.Context) ([]LayerInfo, error)
}

// ImageDestination represents a destination for writing images
type ImageDestination interface {
	// Reference returns the image reference
	Reference() Reference

	// Close releases resources
	Close() error

	// SupportedManifestMIMETypes returns supported manifest MIME types
	SupportedManifestMIMETypes() []string

	// SupportsSignatures returns true if signatures are supported
	SupportsSignatures(ctx context.Context) error

	// DesiredLayerCompression returns desired layer compression
	DesiredLayerCompression() LayerCompression

	// AcceptsForeignLayerURLs returns true if foreign layer URLs are accepted
	AcceptsForeignLayerURLs() bool

	// MustMatchRuntimeOS returns true if runtime OS must match
	MustMatchRuntimeOS() bool

	// IgnoresEmbeddedDockerReference returns true if embedded docker reference is ignored
	IgnoresEmbeddedDockerReference() bool

	// HasThreadSafePutBlob returns true if PutBlob can be called concurrently
	HasThreadSafePutBlob() bool

	// PutBlob writes a blob (layer or config)
	PutBlob(ctx context.Context, stream io.Reader, inputInfo LayerInfo, cache BlobInfoCache, isConfig bool) (LayerInfo, error)

	// TryReusingBlob checks if a blob can be reused
	TryReusingBlob(ctx context.Context, info LayerInfo, cache BlobInfoCache, canSubstitute bool) (bool, LayerInfo, error)

	// PutManifest writes the image manifest
	PutManifest(ctx context.Context, manifest []byte, instanceDigest *string) error

	// PutSignatures writes image signatures
	PutSignatures(ctx context.Context, signatures [][]byte, instanceDigest *string) error

	// Commit commits the image
	Commit(ctx context.Context, unparsedToplevel interface{}) error
}

// LayerInfo contains information about an image layer
type LayerInfo struct {
	// Digest is the layer digest
	Digest string

	// Size is the layer size in bytes
	Size int64

	// MediaType is the layer media type
	MediaType string

	// URLs are URLs where the layer can be downloaded
	URLs []string

	// Annotations are layer annotations
	Annotations map[string]string

	// CompressionOperation describes compression applied to layer
	CompressionOperation LayerCompression

	// CompressionAlgorithm is the compression algorithm
	CompressionAlgorithm string

	// CryptoOperation describes encryption/decryption operation
	CryptoOperation CryptoOperation
}

// LayerCompression represents layer compression type
type LayerCompression int

const (
	// AutoCompression automatically selects compression
	AutoCompression LayerCompression = iota
	// Compress compresses the layer
	Compress
	// Decompress decompresses the layer
	Decompress
	// PreserveOriginal preserves original compression
	PreserveOriginal
)

// CryptoOperation represents encryption/decryption operation
type CryptoOperation int

const (
	// NoCrypto no encryption/decryption
	NoCrypto CryptoOperation = iota
	// Encrypt encrypts the layer
	Encrypt
	// Decrypt decrypts the layer
	Decrypt
)

// BlobInfoCache caches blob information
type BlobInfoCache interface {
	// UncompressedDigest returns the uncompressed digest for a blob
	UncompressedDigest(digest string) string

	// RecordDigestUncompressedPair records a digest pair
	RecordDigestUncompressedPair(anyDigest, uncompressed string)

	// RecordKnownLocation records a known blob location
	RecordKnownLocation(transport Transport, scope string, digest string, location string)

	// CandidateLocations returns candidate locations for a blob
	CandidateLocations(transport Transport, scope string, digest string, canSubstitute bool) []string
}

// ImageInspectInfo contains image inspection information
type ImageInspectInfo struct {
	// Name is the image name
	Name string `json:"Name"`

	// Digest is the image digest
	Digest string `json:"Digest"`

	// RepoTags are repository tags
	RepoTags []string `json:"RepoTags"`

	// Created is the creation timestamp
	Created string `json:"Created"`

	// DockerVersion is the Docker version
	DockerVersion string `json:"DockerVersion"`

	// Labels are image labels
	Labels map[string]string `json:"Labels"`

	// Architecture is the image architecture
	Architecture string `json:"Architecture"`

	// Os is the operating system
	Os string `json:"Os"`

	// Layers are layer digests
	Layers []string `json:"Layers"`

	// LayersData contains detailed layer information
	LayersData []LayerData `json:"LayersData,omitempty"`

	// Env are environment variables
	Env []string `json:"Env,omitempty"`
}

// LayerData contains detailed layer information
type LayerData struct {
	// Digest is the layer digest
	Digest string `json:"Digest"`

	// Size is the layer size
	Size int64 `json:"Size"`

	// MediaType is the media type
	MediaType string `json:"MediaType"`

	// MIMEType is the MIME type
	MIMEType string `json:"MIMEType"`
}

// TransportRegistry manages available transports
type TransportRegistry struct {
	transports map[string]Transport
}

// NewTransportRegistry creates a new transport registry
func NewTransportRegistry() *TransportRegistry {
	return &TransportRegistry{
		transports: make(map[string]Transport),
	}
}

// Register registers a transport
func (r *TransportRegistry) Register(t Transport) {
	r.transports[t.Name()] = t
}

// Get returns a transport by name
func (r *TransportRegistry) Get(name string) (Transport, bool) {
	t, ok := r.transports[name]
	return t, ok
}

// List returns all registered transports
func (r *TransportRegistry) List() []string {
	names := make([]string, 0, len(r.transports))
	for name := range r.transports {
		names = append(names, name)
	}
	return names
}

// Global transport registry
var defaultRegistry = NewTransportRegistry()

// RegisterTransport registers a transport globally
func RegisterTransport(t Transport) {
	defaultRegistry.Register(t)
}

// GetTransport returns a transport by name
func GetTransport(name string) (Transport, bool) {
	return defaultRegistry.Get(name)
}

// ListTransports returns all registered transport names
func ListTransports() []string {
	return defaultRegistry.List()
}

// ParseReference parses a transport-prefixed reference
// Format: <transport>:<reference>
// Examples:
//   - docker://docker.io/library/nginx:latest
//   - dir:/path/to/directory
//   - oci:/path/to/oci/layout
//   - docker-archive:/path/to/image.tar
func ParseReference(ref string) (Reference, error) {
	// Find transport separator
	colonIdx := -1
	for i := 0; i < len(ref); i++ {
		if ref[i] == ':' {
			colonIdx = i
			break
		}
	}

	if colonIdx == -1 || colonIdx == 0 {
		// No transport specified, assume docker://
		colonIdx = 0
		ref = "docker:" + ref
	}

	transportName := ref[:colonIdx]
	refWithinTransport := ref[colonIdx+1:]

	// Skip // if present
	if len(refWithinTransport) >= 2 && refWithinTransport[:2] == "//" {
		refWithinTransport = refWithinTransport[2:]
	}

	transport, ok := GetTransport(transportName)
	if !ok {
		// Unknown transport, assume docker://
		transport, ok = GetTransport("docker")
		if !ok {
			// Docker transport not registered, return error
			return nil, fmt.Errorf("unknown transport: %s (and docker transport not available)", transportName)
		}
		refWithinTransport = ref
	}

	return transport.ParseReference(refWithinTransport)
}
