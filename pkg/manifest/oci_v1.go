package manifest

import (
	"encoding/json"
	"fmt"

	"github.com/opencontainers/go-digest"
)

// OCIManifest represents an OCI Image Manifest v1
type OCIManifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType,omitempty"`
	Config        OCIDescriptor     `json:"config"`
	Layers        []OCIDescriptor   `json:"layers"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Subject       *OCIDescriptor    `json:"subject,omitempty"`
	ArtifactType  string            `json:"artifactType,omitempty"`
}

// OCIDescriptor represents an OCI content descriptor
type OCIDescriptor struct {
	MediaType   string            `json:"mediaType"`
	Size        int64             `json:"size"`
	Digest      string            `json:"digest"`
	URLs        []string          `json:"urls,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Data        []byte            `json:"data,omitempty"`
	Platform    *Platform         `json:"platform,omitempty"`
}

// OCIImageIndex represents an OCI Image Index (multi-platform)
type OCIImageIndex struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType,omitempty"`
	Manifests     []OCIDescriptor   `json:"manifests"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Subject       *OCIDescriptor    `json:"subject,omitempty"`
	ArtifactType  string            `json:"artifactType,omitempty"`
}

// OCIImageConfig represents an OCI image configuration
type OCIImageConfig struct {
	Created      string             `json:"created,omitempty"`
	Author       string             `json:"author,omitempty"`
	Architecture string             `json:"architecture"`
	OS           string             `json:"os"`
	OSVersion    string             `json:"os.version,omitempty"`
	OSFeatures   []string           `json:"os.features,omitempty"`
	Variant      string             `json:"variant,omitempty"`
	Config       OCIImageConfigData `json:"config,omitempty"`
	RootFS       OCIRootFS          `json:"rootfs"`
	History      []OCIHistoryEntry  `json:"history,omitempty"`
}

// OCIImageConfigData represents the execution parameters
type OCIImageConfigData struct {
	User         string              `json:"User,omitempty"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts,omitempty"`
	Env          []string            `json:"Env,omitempty"`
	Entrypoint   []string            `json:"Entrypoint,omitempty"`
	Cmd          []string            `json:"Cmd,omitempty"`
	Volumes      map[string]struct{} `json:"Volumes,omitempty"`
	WorkingDir   string              `json:"WorkingDir,omitempty"`
	Labels       map[string]string   `json:"Labels,omitempty"`
	StopSignal   string              `json:"StopSignal,omitempty"`
}

// OCIRootFS represents the root filesystem
type OCIRootFS struct {
	Type    string   `json:"type"`
	DiffIDs []string `json:"diff_ids"`
}

// OCIHistoryEntry represents a history entry
type OCIHistoryEntry struct {
	Created    string `json:"created,omitempty"`
	CreatedBy  string `json:"created_by,omitempty"`
	Author     string `json:"author,omitempty"`
	Comment    string `json:"comment,omitempty"`
	EmptyLayer bool   `json:"empty_layer,omitempty"`
}

// NewOCIManifest creates a new OCI manifest
func NewOCIManifest(config OCIDescriptor, layers []OCIDescriptor) *OCIManifest {
	return &OCIManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config:        config,
		Layers:        layers,
	}
}

// NewOCIImageIndex creates a new OCI image index
func NewOCIImageIndex(manifests []OCIDescriptor) *OCIImageIndex {
	return &OCIImageIndex{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests:     manifests,
	}
}

// Validate validates an OCI manifest
func (m *OCIManifest) Validate() error {
	if m.SchemaVersion != 2 {
		return fmt.Errorf("invalid schema version: expected 2, got %d", m.SchemaVersion)
	}

	// MediaType is optional but should be valid if present
	if m.MediaType != "" && m.MediaType != "application/vnd.oci.image.manifest.v1+json" {
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

	if m.Subject != nil {
		if err := m.Subject.Validate(); err != nil {
			return fmt.Errorf("invalid subject descriptor: %w", err)
		}
	}

	return nil
}

// Validate validates an OCI descriptor
func (d *OCIDescriptor) Validate() error {
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

// Validate validates an OCI image index
func (i *OCIImageIndex) Validate() error {
	if i.SchemaVersion != 2 {
		return fmt.Errorf("invalid schema version: expected 2, got %d", i.SchemaVersion)
	}

	// MediaType is optional but should be valid if present
	if i.MediaType != "" && i.MediaType != "application/vnd.oci.image.index.v1+json" {
		return fmt.Errorf("invalid media type: %s", i.MediaType)
	}

	if len(i.Manifests) == 0 {
		return fmt.Errorf("image index must have at least one manifest")
	}

	for idx, manifest := range i.Manifests {
		if err := manifest.Validate(); err != nil {
			return fmt.Errorf("invalid manifest %d: %w", idx, err)
		}
	}

	if i.Subject != nil {
		if err := i.Subject.Validate(); err != nil {
			return fmt.Errorf("invalid subject descriptor: %w", err)
		}
	}

	return nil
}

// Marshal marshals the OCI manifest to JSON
func (m *OCIManifest) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal unmarshals JSON into an OCI manifest
func (m *OCIManifest) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

// Marshal marshals the OCI image index to JSON
func (i *OCIImageIndex) Marshal() ([]byte, error) {
	return json.Marshal(i)
}

// Unmarshal unmarshals JSON into an OCI image index
func (i *OCIImageIndex) Unmarshal(data []byte) error {
	return json.Unmarshal(data, i)
}

// GetConfigDigest returns the digest of the config
func (m *OCIManifest) GetConfigDigest() string {
	return m.Config.Digest
}

// GetLayerDigests returns the digests of all layers
func (m *OCIManifest) GetLayerDigests() []string {
	digests := make([]string, len(m.Layers))
	for i, layer := range m.Layers {
		digests[i] = layer.Digest
	}
	return digests
}

// GetTotalSize returns the total size of config and all layers
func (m *OCIManifest) GetTotalSize() int64 {
	total := m.Config.Size
	for _, layer := range m.Layers {
		total += layer.Size
	}
	return total
}

// FindManifestByPlatform finds a manifest in the index matching the platform
func (i *OCIImageIndex) FindManifestByPlatform(os, arch string) (*OCIDescriptor, error) {
	for _, manifest := range i.Manifests {
		if manifest.Platform != nil {
			if manifest.Platform.OS == os && manifest.Platform.Architecture == arch {
				return &manifest, nil
			}
		}
	}
	return nil, fmt.Errorf("no manifest found for platform %s/%s", os, arch)
}

// GetPlatforms returns all unique platforms in the image index
func (i *OCIImageIndex) GetPlatforms() []Platform {
	platforms := make([]Platform, 0, len(i.Manifests))
	for _, manifest := range i.Manifests {
		if manifest.Platform != nil {
			platforms = append(platforms, *manifest.Platform)
		}
	}
	return platforms
}

// AddManifest adds a manifest to the image index
func (i *OCIImageIndex) AddManifest(descriptor OCIDescriptor) {
	i.Manifests = append(i.Manifests, descriptor)
}

// RemoveManifest removes a manifest from the index by digest
func (i *OCIImageIndex) RemoveManifest(digestStr string) error {
	for idx, manifest := range i.Manifests {
		if manifest.Digest == digestStr {
			i.Manifests = append(i.Manifests[:idx], i.Manifests[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("manifest with digest %s not found", digestStr)
}

// SetAnnotation sets an annotation on the manifest
func (m *OCIManifest) SetAnnotation(key, value string) {
	if m.Annotations == nil {
		m.Annotations = make(map[string]string)
	}
	m.Annotations[key] = value
}

// GetAnnotation gets an annotation from the manifest
func (m *OCIManifest) GetAnnotation(key string) (string, bool) {
	if m.Annotations == nil {
		return "", false
	}
	val, ok := m.Annotations[key]
	return val, ok
}

// SetAnnotation sets an annotation on the image index
func (i *OCIImageIndex) SetAnnotation(key, value string) {
	if i.Annotations == nil {
		i.Annotations = make(map[string]string)
	}
	i.Annotations[key] = value
}

// GetAnnotation gets an annotation from the image index
func (i *OCIImageIndex) GetAnnotation(key string) (string, bool) {
	if i.Annotations == nil {
		return "", false
	}
	val, ok := i.Annotations[key]
	return val, ok
}
