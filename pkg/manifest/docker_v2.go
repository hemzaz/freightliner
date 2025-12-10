package manifest

import (
	"encoding/json"
	"fmt"

	"github.com/opencontainers/go-digest"
)

// DockerV2Schema2Manifest represents a Docker v2 Schema 2 manifest
type DockerV2Schema2Manifest struct {
	SchemaVersion int                `json:"schemaVersion"`
	MediaType     string             `json:"mediaType"`
	Config        DockerDescriptor   `json:"config"`
	Layers        []DockerDescriptor `json:"layers"`
	Annotations   map[string]string  `json:"annotations,omitempty"`
}

// DockerDescriptor represents a Docker content descriptor
type DockerDescriptor struct {
	MediaType string   `json:"mediaType"`
	Size      int64    `json:"size"`
	Digest    string   `json:"digest"`
	URLs      []string `json:"urls,omitempty"`
}

// DockerV2Schema1Manifest represents a Docker v2 Schema 1 manifest (deprecated)
type DockerV2Schema1Manifest struct {
	SchemaVersion int                        `json:"schemaVersion"`
	Name          string                     `json:"name"`
	Tag           string                     `json:"tag"`
	Architecture  string                     `json:"architecture"`
	FSLayers      []DockerV2Schema1FSLayer   `json:"fsLayers"`
	History       []DockerV2Schema1History   `json:"history"`
	Signatures    []DockerV2Schema1Signature `json:"signatures,omitempty"`
}

// DockerV2Schema1FSLayer represents a filesystem layer in Docker v2 Schema 1
type DockerV2Schema1FSLayer struct {
	BlobSum string `json:"blobSum"`
}

// DockerV2Schema1History represents history entry in Docker v2 Schema 1
type DockerV2Schema1History struct {
	V1Compatibility string `json:"v1Compatibility"`
}

// DockerV2Schema1Signature represents a signature in Docker v2 Schema 1
type DockerV2Schema1Signature struct {
	Header    DockerV2Schema1JWSHeader `json:"header"`
	Signature string                   `json:"signature"`
	Protected string                   `json:"protected"`
}

// DockerV2Schema1JWSHeader represents a JWS header in Docker v2 Schema 1
type DockerV2Schema1JWSHeader struct {
	JWK       DockerV2Schema1JWK `json:"jwk"`
	Algorithm string             `json:"alg"`
}

// DockerV2Schema1JWK represents a JSON Web Key in Docker v2 Schema 1
type DockerV2Schema1JWK struct {
	CRV string `json:"crv"`
	KID string `json:"kid"`
	KTY string `json:"kty"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// DockerManifestList represents a Docker manifest list (multi-platform)
type DockerManifestList struct {
	SchemaVersion int                            `json:"schemaVersion"`
	MediaType     string                         `json:"mediaType"`
	Manifests     []DockerManifestListDescriptor `json:"manifests"`
	Annotations   map[string]string              `json:"annotations,omitempty"`
}

// DockerManifestListDescriptor represents a manifest descriptor in a manifest list
type DockerManifestListDescriptor struct {
	MediaType   string            `json:"mediaType"`
	Size        int64             `json:"size"`
	Digest      string            `json:"digest"`
	Platform    *Platform         `json:"platform,omitempty"`
	URLs        []string          `json:"urls,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// NewDockerV2Schema2Manifest creates a new Docker v2 Schema 2 manifest
func NewDockerV2Schema2Manifest(config DockerDescriptor, layers []DockerDescriptor) *DockerV2Schema2Manifest {
	return &DockerV2Schema2Manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.docker.distribution.manifest.v2+json",
		Config:        config,
		Layers:        layers,
	}
}

// NewDockerManifestList creates a new Docker manifest list
func NewDockerManifestList(manifests []DockerManifestListDescriptor) *DockerManifestList {
	return &DockerManifestList{
		SchemaVersion: 2,
		MediaType:     "application/vnd.docker.distribution.manifest.list.v2+json",
		Manifests:     manifests,
	}
}

// Validate validates a Docker v2 Schema 2 manifest
func (m *DockerV2Schema2Manifest) Validate() error {
	if m.SchemaVersion != 2 {
		return fmt.Errorf("invalid schema version: expected 2, got %d", m.SchemaVersion)
	}

	if m.MediaType != "application/vnd.docker.distribution.manifest.v2+json" {
		return fmt.Errorf("invalid media type: %s", m.MediaType)
	}

	if err := m.Config.Validate(); err != nil {
		return fmt.Errorf("invalid config descriptor: %w", err)
	}

	if len(m.Layers) == 0 {
		return fmt.Errorf("manifest must have at least one layer")
	}

	for i, layer := range m.Layers {
		if err := layer.Validate(); err != nil {
			return fmt.Errorf("invalid layer %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates a Docker descriptor
func (d *DockerDescriptor) Validate() error {
	if d.MediaType == "" {
		return fmt.Errorf("media type cannot be empty")
	}

	if d.Size < 0 {
		return fmt.Errorf("size cannot be negative: %d", d.Size)
	}

	if d.Digest == "" {
		return fmt.Errorf("digest cannot be empty")
	}

	if _, err := digest.Parse(d.Digest); err != nil {
		return fmt.Errorf("invalid digest format: %w", err)
	}

	return nil
}

// Validate validates a Docker manifest list
func (m *DockerManifestList) Validate() error {
	if m.SchemaVersion != 2 {
		return fmt.Errorf("invalid schema version: expected 2, got %d", m.SchemaVersion)
	}

	if m.MediaType != "application/vnd.docker.distribution.manifest.list.v2+json" {
		return fmt.Errorf("invalid media type: %s", m.MediaType)
	}

	if len(m.Manifests) == 0 {
		return fmt.Errorf("manifest list must have at least one manifest")
	}

	for i, manifest := range m.Manifests {
		if err := manifest.Validate(); err != nil {
			return fmt.Errorf("invalid manifest %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates a Docker manifest list descriptor
func (d *DockerManifestListDescriptor) Validate() error {
	if d.MediaType == "" {
		return fmt.Errorf("media type cannot be empty")
	}

	if d.Size < 0 {
		return fmt.Errorf("size cannot be negative: %d", d.Size)
	}

	if d.Digest == "" {
		return fmt.Errorf("digest cannot be empty")
	}

	if _, err := digest.Parse(d.Digest); err != nil {
		return fmt.Errorf("invalid digest format: %w", err)
	}

	if d.Platform != nil {
		if err := d.Platform.Validate(); err != nil {
			return fmt.Errorf("invalid platform: %w", err)
		}
	}

	return nil
}

// Marshal marshals the Docker v2 Schema 2 manifest to JSON
func (m *DockerV2Schema2Manifest) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal unmarshals JSON into a Docker v2 Schema 2 manifest
func (m *DockerV2Schema2Manifest) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

// Marshal marshals the Docker manifest list to JSON
func (m *DockerManifestList) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal unmarshals JSON into a Docker manifest list
func (m *DockerManifestList) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

// GetConfigDigest returns the digest of the config
func (m *DockerV2Schema2Manifest) GetConfigDigest() string {
	return m.Config.Digest
}

// GetLayerDigests returns the digests of all layers
func (m *DockerV2Schema2Manifest) GetLayerDigests() []string {
	digests := make([]string, len(m.Layers))
	for i, layer := range m.Layers {
		digests[i] = layer.Digest
	}
	return digests
}

// GetTotalSize returns the total size of config and all layers
func (m *DockerV2Schema2Manifest) GetTotalSize() int64 {
	total := m.Config.Size
	for _, layer := range m.Layers {
		total += layer.Size
	}
	return total
}

// FindManifestByPlatform finds a manifest in the list matching the platform
func (m *DockerManifestList) FindManifestByPlatform(os, arch string) (*DockerManifestListDescriptor, error) {
	for _, manifest := range m.Manifests {
		if manifest.Platform != nil {
			if manifest.Platform.OS == os && manifest.Platform.Architecture == arch {
				return &manifest, nil
			}
		}
	}
	return nil, fmt.Errorf("no manifest found for platform %s/%s", os, arch)
}

// GetPlatforms returns all unique platforms in the manifest list
func (m *DockerManifestList) GetPlatforms() []Platform {
	platforms := make([]Platform, 0, len(m.Manifests))
	for _, manifest := range m.Manifests {
		if manifest.Platform != nil {
			platforms = append(platforms, *manifest.Platform)
		}
	}
	return platforms
}

// AddManifest adds a manifest to the manifest list
func (m *DockerManifestList) AddManifest(descriptor DockerManifestListDescriptor) {
	m.Manifests = append(m.Manifests, descriptor)
}

// RemoveManifest removes a manifest from the list by digest
func (m *DockerManifestList) RemoveManifest(digestStr string) error {
	for i, manifest := range m.Manifests {
		if manifest.Digest == digestStr {
			m.Manifests = append(m.Manifests[:i], m.Manifests[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("manifest with digest %s not found", digestStr)
}
