package manifest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/opencontainers/go-digest"
)

// Converter handles conversion between different manifest formats
type Converter struct {
	// PreserveAnnotations determines whether to preserve OCI annotations when converting
	PreserveAnnotations bool
	// PreserveLabels determines whether to preserve Docker config labels when converting
	PreserveLabels bool
	// StrictValidation enables strict validation during conversion
	StrictValidation bool
}

// NewConverter creates a new manifest converter with default settings
func NewConverter() *Converter {
	return &Converter{
		PreserveAnnotations: true,
		PreserveLabels:      true,
		StrictValidation:    false,
	}
}

// StandardManifest represents a normalized manifest structure
type StandardManifest struct {
	SchemaVersion int                  `json:"schemaVersion"`
	MediaType     string               `json:"mediaType"`
	Config        StandardDescriptor   `json:"config"`
	Layers        []StandardDescriptor `json:"layers"`
	Annotations   map[string]string    `json:"annotations,omitempty"`
	Subject       *StandardDescriptor  `json:"subject,omitempty"`
	ArtifactType  string               `json:"artifactType,omitempty"`
	Platform      *Platform            `json:"platform,omitempty"`
}

// StandardDescriptor represents a normalized descriptor
type StandardDescriptor struct {
	MediaType   string            `json:"mediaType"`
	Size        int64             `json:"size"`
	Digest      string            `json:"digest"`
	URLs        []string          `json:"urls,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Data        []byte            `json:"data,omitempty"`
	Platform    *Platform         `json:"platform,omitempty"`
}

// ManifestType represents the type of manifest
type ManifestType string

const (
	ManifestTypeDockerV2Schema1    ManifestType = "docker-v2-schema1"
	ManifestTypeDockerV2Schema2    ManifestType = "docker-v2-schema2"
	ManifestTypeDockerManifestList ManifestType = "docker-manifest-list"
	ManifestTypeOCIv1              ManifestType = "oci-v1"
	ManifestTypeOCIIndex           ManifestType = "oci-index"
	ManifestTypeUnknown            ManifestType = "unknown"
)

// DetectManifestType detects the type of manifest from raw bytes
func (c *Converter) DetectManifestType(manifestBytes []byte) (ManifestType, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(manifestBytes, &raw); err != nil {
		return ManifestTypeUnknown, fmt.Errorf("failed to parse manifest: %w", err)
	}

	mediaType, _ := raw["mediaType"].(string)
	schemaVersion, _ := raw["schemaVersion"].(float64)

	// Check for manifest list/index first
	if _, hasManifests := raw["manifests"]; hasManifests {
		switch mediaType {
		case string(types.DockerManifestList):
			return ManifestTypeDockerManifestList, nil
		case string(types.OCIImageIndex):
			return ManifestTypeOCIIndex, nil
		case "": // No media type, check schema version
			if schemaVersion == 2 {
				return ManifestTypeDockerManifestList, nil
			}
		}
	}

	// Check for single image manifest
	switch mediaType {
	case string(types.DockerManifestSchema2):
		return ManifestTypeDockerV2Schema2, nil
	case string(types.OCIManifestSchema1):
		return ManifestTypeOCIv1, nil
	case "application/vnd.docker.distribution.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v1+prettyjws":
		return ManifestTypeDockerV2Schema1, nil
	case "":
		// No media type specified, use schema version
		if schemaVersion == 1 {
			return ManifestTypeDockerV2Schema1, nil
		} else if schemaVersion == 2 {
			// Could be Docker v2 or OCI, check for config
			if _, hasConfig := raw["config"]; hasConfig {
				return ManifestTypeDockerV2Schema2, nil
			}
		}
	}

	return ManifestTypeUnknown, fmt.Errorf("unable to detect manifest type: mediaType=%s, schemaVersion=%.0f", mediaType, schemaVersion)
}

// DockerToOCI converts Docker v2 Schema 2 manifest to OCI v1 format
func (c *Converter) DockerToOCI(dockerManifest []byte) ([]byte, error) {
	manifestType, err := c.DetectManifestType(dockerManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to detect manifest type: %w", err)
	}

	switch manifestType {
	case ManifestTypeDockerV2Schema2:
		return c.dockerV2Schema2ToOCI(dockerManifest)
	case ManifestTypeDockerManifestList:
		return c.dockerManifestListToOCIIndex(dockerManifest)
	case ManifestTypeDockerV2Schema1:
		return nil, fmt.Errorf("Docker v2 Schema 1 is deprecated and not supported for conversion to OCI")
	case ManifestTypeOCIv1, ManifestTypeOCIIndex:
		// Already OCI format
		return dockerManifest, nil
	default:
		return nil, fmt.Errorf("unsupported manifest type for Docker to OCI conversion: %s", manifestType)
	}
}

// OCIToDocker converts OCI v1 manifest to Docker v2 Schema 2 format
func (c *Converter) OCIToDocker(ociManifest []byte) ([]byte, error) {
	manifestType, err := c.DetectManifestType(ociManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to detect manifest type: %w", err)
	}

	switch manifestType {
	case ManifestTypeOCIv1:
		return c.ociV1ToDockerV2Schema2(ociManifest)
	case ManifestTypeOCIIndex:
		return c.ociIndexToDockerManifestList(ociManifest)
	case ManifestTypeDockerV2Schema2, ManifestTypeDockerManifestList:
		// Already Docker format
		return ociManifest, nil
	default:
		return nil, fmt.Errorf("unsupported manifest type for OCI to Docker conversion: %s", manifestType)
	}
}

// Normalize converts any supported manifest format to a standard format
func (c *Converter) Normalize(manifestBytes []byte) (*StandardManifest, error) {
	manifestType, err := c.DetectManifestType(manifestBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to detect manifest type: %w", err)
	}

	switch manifestType {
	case ManifestTypeDockerV2Schema2:
		return c.normalizeDockerV2Schema2(manifestBytes)
	case ManifestTypeOCIv1:
		return c.normalizeOCIv1(manifestBytes)
	case ManifestTypeDockerV2Schema1:
		return c.normalizeDockerV2Schema1(manifestBytes)
	default:
		return nil, fmt.Errorf("manifest type %s cannot be normalized to standard format", manifestType)
	}
}

// dockerV2Schema2ToOCI converts Docker v2 Schema 2 to OCI v1
func (c *Converter) dockerV2Schema2ToOCI(dockerManifest []byte) ([]byte, error) {
	var docker DockerV2Schema2Manifest
	if err := json.Unmarshal(dockerManifest, &docker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Docker manifest: %w", err)
	}

	if c.StrictValidation {
		if err := docker.Validate(); err != nil {
			return nil, fmt.Errorf("invalid Docker manifest: %w", err)
		}
	}

	oci := OCIManifest{
		SchemaVersion: 2,
		MediaType:     string(types.OCIManifestSchema1),
		Config: OCIDescriptor{
			MediaType: c.convertDockerConfigMediaTypeToOCI(docker.Config.MediaType),
			Size:      docker.Config.Size,
			Digest:    docker.Config.Digest,
		},
		Layers: make([]OCIDescriptor, len(docker.Layers)),
	}

	// Convert layers
	for i, layer := range docker.Layers {
		oci.Layers[i] = OCIDescriptor{
			MediaType: c.convertDockerLayerMediaTypeToOCI(layer.MediaType),
			Size:      layer.Size,
			Digest:    layer.Digest,
			URLs:      layer.URLs,
		}
	}

	// Add annotations if preserving
	if c.PreserveAnnotations && docker.Annotations != nil {
		oci.Annotations = docker.Annotations
	}

	return json.Marshal(oci)
}

// ociV1ToDockerV2Schema2 converts OCI v1 to Docker v2 Schema 2
func (c *Converter) ociV1ToDockerV2Schema2(ociManifest []byte) ([]byte, error) {
	var oci OCIManifest
	if err := json.Unmarshal(ociManifest, &oci); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OCI manifest: %w", err)
	}

	if c.StrictValidation {
		if err := oci.Validate(); err != nil {
			return nil, fmt.Errorf("invalid OCI manifest: %w", err)
		}
	}

	docker := DockerV2Schema2Manifest{
		SchemaVersion: 2,
		MediaType:     string(types.DockerManifestSchema2),
		Config: DockerDescriptor{
			MediaType: c.convertOCIConfigMediaTypeToDocker(oci.Config.MediaType),
			Size:      oci.Config.Size,
			Digest:    oci.Config.Digest,
		},
		Layers: make([]DockerDescriptor, len(oci.Layers)),
	}

	// Convert layers
	for i, layer := range oci.Layers {
		docker.Layers[i] = DockerDescriptor{
			MediaType: c.convertOCILayerMediaTypeToDocker(layer.MediaType),
			Size:      layer.Size,
			Digest:    layer.Digest,
			URLs:      layer.URLs,
		}
	}

	// Preserve annotations if requested (Docker doesn't officially support them at manifest level)
	if c.PreserveAnnotations && oci.Annotations != nil {
		docker.Annotations = oci.Annotations
	}

	return json.Marshal(docker)
}

// convertDockerConfigMediaTypeToOCI converts Docker config media type to OCI
func (c *Converter) convertDockerConfigMediaTypeToOCI(dockerMediaType string) string {
	switch dockerMediaType {
	case string(types.DockerConfigJSON):
		return string(types.OCIConfigJSON)
	default:
		return dockerMediaType
	}
}

// convertDockerLayerMediaTypeToOCI converts Docker layer media type to OCI
func (c *Converter) convertDockerLayerMediaTypeToOCI(dockerMediaType string) string {
	switch dockerMediaType {
	case string(types.DockerLayer):
		// DockerLayer is already gzipped (application/vnd.docker.image.rootfs.diff.tar.gzip)
		return string(types.OCILayer) // OCILayer is also gzipped by default
	case string(types.DockerUncompressedLayer):
		return string(types.OCIUncompressedLayer)
	case string(types.DockerForeignLayer):
		return string(types.OCIRestrictedLayer)
	default:
		return dockerMediaType
	}
}

// convertOCIConfigMediaTypeToDocker converts OCI config media type to Docker
func (c *Converter) convertOCIConfigMediaTypeToDocker(ociMediaType string) string {
	switch ociMediaType {
	case string(types.OCIConfigJSON):
		return string(types.DockerConfigJSON)
	default:
		return ociMediaType
	}
}

// convertOCILayerMediaTypeToDocker converts OCI layer media type to Docker
func (c *Converter) convertOCILayerMediaTypeToDocker(ociMediaType string) string {
	switch ociMediaType {
	case string(types.OCILayer):
		return string(types.DockerLayer)
	case string(types.OCIUncompressedLayer):
		return string(types.DockerUncompressedLayer)
	case string(types.OCIRestrictedLayer):
		return string(types.DockerForeignLayer)
	default:
		// Handle OCI layer with compression suffix for backwards compatibility
		if strings.HasPrefix(ociMediaType, string(types.OCILayer)) {
			return string(types.DockerLayer)
		}
		return ociMediaType
	}
}

// normalizeDockerV2Schema2 normalizes Docker v2 Schema 2 to standard format
func (c *Converter) normalizeDockerV2Schema2(manifestBytes []byte) (*StandardManifest, error) {
	var docker DockerV2Schema2Manifest
	if err := json.Unmarshal(manifestBytes, &docker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Docker manifest: %w", err)
	}

	std := &StandardManifest{
		SchemaVersion: docker.SchemaVersion,
		MediaType:     docker.MediaType,
		Config: StandardDescriptor{
			MediaType: docker.Config.MediaType,
			Size:      docker.Config.Size,
			Digest:    docker.Config.Digest,
		},
		Layers:      make([]StandardDescriptor, len(docker.Layers)),
		Annotations: docker.Annotations,
	}

	for i, layer := range docker.Layers {
		std.Layers[i] = StandardDescriptor{
			MediaType: layer.MediaType,
			Size:      layer.Size,
			Digest:    layer.Digest,
			URLs:      layer.URLs,
		}
	}

	return std, nil
}

// normalizeOCIv1 normalizes OCI v1 to standard format
func (c *Converter) normalizeOCIv1(manifestBytes []byte) (*StandardManifest, error) {
	var oci OCIManifest
	if err := json.Unmarshal(manifestBytes, &oci); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OCI manifest: %w", err)
	}

	std := &StandardManifest{
		SchemaVersion: oci.SchemaVersion,
		MediaType:     oci.MediaType,
		Config: StandardDescriptor{
			MediaType: oci.Config.MediaType,
			Size:      oci.Config.Size,
			Digest:    oci.Config.Digest,
		},
		Layers:       make([]StandardDescriptor, len(oci.Layers)),
		Annotations:  oci.Annotations,
		ArtifactType: oci.ArtifactType,
	}

	if oci.Subject != nil {
		std.Subject = &StandardDescriptor{
			MediaType: oci.Subject.MediaType,
			Size:      oci.Subject.Size,
			Digest:    oci.Subject.Digest,
		}
	}

	for i, layer := range oci.Layers {
		std.Layers[i] = StandardDescriptor{
			MediaType:   layer.MediaType,
			Size:        layer.Size,
			Digest:      layer.Digest,
			URLs:        layer.URLs,
			Annotations: layer.Annotations,
		}
	}

	return std, nil
}

// normalizeDockerV2Schema1 normalizes Docker v2 Schema 1 to standard format
func (c *Converter) normalizeDockerV2Schema1(manifestBytes []byte) (*StandardManifest, error) {
	var docker DockerV2Schema1Manifest
	if err := json.Unmarshal(manifestBytes, &docker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Docker v2 Schema 1 manifest: %w", err)
	}

	// Docker v2 Schema 1 doesn't have explicit config or layers
	// We need to extract information from FSLayers and History
	std := &StandardManifest{
		SchemaVersion: docker.SchemaVersion,
		MediaType:     "application/vnd.docker.distribution.manifest.v1+json",
		Layers:        make([]StandardDescriptor, len(docker.FSLayers)),
	}

	for i, fsLayer := range docker.FSLayers {
		std.Layers[i] = StandardDescriptor{
			MediaType: string(types.DockerLayer),
			Digest:    fsLayer.BlobSum,
			// Size is not available in Schema 1
			Size: 0,
		}
	}

	return std, nil
}

// ValidateDigest validates that a digest matches the expected format
func ValidateDigest(digestStr string) error {
	_, err := digest.Parse(digestStr)
	return err
}

// ComputeDigest computes the digest of data
func ComputeDigest(data []byte) string {
	return digest.FromBytes(data).String()
}
