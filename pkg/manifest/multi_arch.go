package manifest

import (
	"encoding/json"
	"fmt"
	"runtime"
)

// Platform represents a platform (OS/Architecture) for multi-arch images
type Platform struct {
	OS           string   `json:"os"`
	Architecture string   `json:"architecture"`
	Variant      string   `json:"variant,omitempty"`
	OSVersion    string   `json:"os.version,omitempty"`
	OSFeatures   []string `json:"os.features,omitempty"`
}

// MultiArchManifest represents a multi-architecture manifest (either Docker or OCI)
type MultiArchManifest struct {
	SchemaVersion int                `json:"schemaVersion"`
	MediaType     string             `json:"mediaType"`
	Manifests     []PlatformManifest `json:"manifests"`
	Annotations   map[string]string  `json:"annotations,omitempty"`
}

// PlatformManifest represents a manifest for a specific platform
type PlatformManifest struct {
	MediaType   string            `json:"mediaType"`
	Digest      string            `json:"digest"`
	Size        int64             `json:"size"`
	Platform    Platform          `json:"platform"`
	URLs        []string          `json:"urls,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// MultiArchBuilder helps build multi-architecture manifests
type MultiArchBuilder struct {
	manifests   []PlatformManifest
	annotations map[string]string
	format      ManifestType
}

// NewMultiArchBuilder creates a new multi-architecture manifest builder
func NewMultiArchBuilder(format ManifestType) *MultiArchBuilder {
	return &MultiArchBuilder{
		manifests:   make([]PlatformManifest, 0),
		annotations: make(map[string]string),
		format:      format,
	}
}

// AddPlatformManifest adds a manifest for a specific platform
func (b *MultiArchBuilder) AddPlatformManifest(manifest PlatformManifest) error {
	if err := manifest.Platform.Validate(); err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}

	// Check for duplicate platforms
	for _, existing := range b.manifests {
		if existing.Platform.Matches(&manifest.Platform) {
			return fmt.Errorf("platform %s/%s already exists", manifest.Platform.OS, manifest.Platform.Architecture)
		}
	}

	b.manifests = append(b.manifests, manifest)
	return nil
}

// AddAnnotation adds an annotation to the multi-arch manifest
func (b *MultiArchBuilder) AddAnnotation(key, value string) {
	b.annotations[key] = value
}

// Build builds the multi-architecture manifest
func (b *MultiArchBuilder) Build() (*MultiArchManifest, error) {
	if len(b.manifests) == 0 {
		return nil, fmt.Errorf("at least one platform manifest is required")
	}

	manifest := &MultiArchManifest{
		SchemaVersion: 2,
		Manifests:     b.manifests,
		Annotations:   b.annotations,
	}

	switch b.format {
	case ManifestTypeDockerManifestList:
		manifest.MediaType = "application/vnd.docker.distribution.manifest.list.v2+json"
	case ManifestTypeOCIIndex:
		manifest.MediaType = "application/vnd.oci.image.index.v1+json"
	default:
		return nil, fmt.Errorf("unsupported multi-arch format: %s", b.format)
	}

	return manifest, nil
}

// Validate validates a platform
func (p *Platform) Validate() error {
	if p.OS == "" {
		return fmt.Errorf("OS cannot be empty")
	}

	if p.Architecture == "" {
		return fmt.Errorf("architecture cannot be empty")
	}

	// Validate common OS values
	validOS := map[string]bool{
		"linux":   true,
		"windows": true,
		"darwin":  true,
		"freebsd": true,
		"netbsd":  true,
		"openbsd": true,
		"solaris": true,
		"aix":     true,
	}

	if !validOS[p.OS] {
		// Not a hard error, just a warning case
		// Allow unknown OS values for future compatibility
	}

	// Validate common architecture values
	validArch := map[string]bool{
		"amd64":    true,
		"386":      true,
		"arm":      true,
		"arm64":    true,
		"ppc64le":  true,
		"s390x":    true,
		"mips64le": true,
		"riscv64":  true,
	}

	if !validArch[p.Architecture] {
		// Not a hard error, just a warning case
		// Allow unknown architectures for future compatibility
	}

	return nil
}

// Matches checks if this platform matches another platform
func (p *Platform) Matches(other *Platform) bool {
	if p.OS != other.OS {
		return false
	}

	if p.Architecture != other.Architecture {
		return false
	}

	// Variant is optional but must match if both are specified
	if p.Variant != "" && other.Variant != "" && p.Variant != other.Variant {
		return false
	}

	return true
}

// MatchesRuntime checks if this platform matches the current runtime
func (p *Platform) MatchesRuntime() bool {
	if p.OS != runtime.GOOS {
		return false
	}

	if p.Architecture != runtime.GOARCH {
		return false
	}

	return true
}

// String returns a string representation of the platform
func (p *Platform) String() string {
	s := fmt.Sprintf("%s/%s", p.OS, p.Architecture)
	if p.Variant != "" {
		s += "/" + p.Variant
	}
	if p.OSVersion != "" {
		s += ":" + p.OSVersion
	}
	return s
}

// GetManifestForPlatform returns the manifest matching the specified platform
func (m *MultiArchManifest) GetManifestForPlatform(os, arch string) (*PlatformManifest, error) {
	targetPlatform := Platform{
		OS:           os,
		Architecture: arch,
	}

	for _, manifest := range m.Manifests {
		if manifest.Platform.Matches(&targetPlatform) {
			return &manifest, nil
		}
	}

	return nil, fmt.Errorf("no manifest found for platform %s/%s", os, arch)
}

// GetManifestForCurrentPlatform returns the manifest matching the current runtime platform
func (m *MultiArchManifest) GetManifestForCurrentPlatform() (*PlatformManifest, error) {
	return m.GetManifestForPlatform(runtime.GOOS, runtime.GOARCH)
}

// GetPlatforms returns all platforms in the multi-arch manifest
func (m *MultiArchManifest) GetPlatforms() []Platform {
	platforms := make([]Platform, len(m.Manifests))
	for i, manifest := range m.Manifests {
		platforms[i] = manifest.Platform
	}
	return platforms
}

// HasPlatform checks if the multi-arch manifest has a specific platform
func (m *MultiArchManifest) HasPlatform(os, arch string) bool {
	_, err := m.GetManifestForPlatform(os, arch)
	return err == nil
}

// AddManifest adds a manifest to the multi-arch manifest
func (m *MultiArchManifest) AddManifest(manifest PlatformManifest) error {
	if err := manifest.Platform.Validate(); err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}

	// Check for duplicates
	for _, existing := range m.Manifests {
		if existing.Platform.Matches(&manifest.Platform) {
			return fmt.Errorf("platform %s already exists", manifest.Platform.String())
		}
	}

	m.Manifests = append(m.Manifests, manifest)
	return nil
}

// RemoveManifest removes a manifest by digest
func (m *MultiArchManifest) RemoveManifest(digest string) error {
	for i, manifest := range m.Manifests {
		if manifest.Digest == digest {
			m.Manifests = append(m.Manifests[:i], m.Manifests[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("manifest with digest %s not found", digest)
}

// RemovePlatform removes manifests for a specific platform
func (m *MultiArchManifest) RemovePlatform(os, arch string) error {
	targetPlatform := Platform{
		OS:           os,
		Architecture: arch,
	}

	found := false
	newManifests := make([]PlatformManifest, 0, len(m.Manifests))
	for _, manifest := range m.Manifests {
		if manifest.Platform.Matches(&targetPlatform) {
			found = true
			continue
		}
		newManifests = append(newManifests, manifest)
	}

	if !found {
		return fmt.Errorf("no manifest found for platform %s/%s", os, arch)
	}

	m.Manifests = newManifests
	return nil
}

// Marshal marshals the multi-arch manifest to JSON
func (m *MultiArchManifest) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal unmarshals JSON into a multi-arch manifest
func (m *MultiArchManifest) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

// Validate validates the multi-arch manifest
func (m *MultiArchManifest) Validate() error {
	if m.SchemaVersion != 2 {
		return fmt.Errorf("invalid schema version: expected 2, got %d", m.SchemaVersion)
	}

	if m.MediaType == "" {
		return fmt.Errorf("media type cannot be empty")
	}

	if len(m.Manifests) == 0 {
		return fmt.Errorf("at least one manifest is required")
	}

	for i, manifest := range m.Manifests {
		if err := manifest.Validate(); err != nil {
			return fmt.Errorf("invalid manifest %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates a platform manifest
func (pm *PlatformManifest) Validate() error {
	if pm.MediaType == "" {
		return fmt.Errorf("media type cannot be empty")
	}

	if pm.Digest == "" {
		return fmt.Errorf("digest cannot be empty")
	}

	if pm.Size < 0 {
		return fmt.Errorf("size cannot be negative: %d", pm.Size)
	}

	if err := pm.Platform.Validate(); err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}

	return nil
}

// CreateMultiArch creates a multi-architecture manifest list
func CreateMultiArch(manifests []PlatformManifest, format ManifestType) (*MultiArchManifest, error) {
	builder := NewMultiArchBuilder(format)

	for _, manifest := range manifests {
		if err := builder.AddPlatformManifest(manifest); err != nil {
			return nil, err
		}
	}

	return builder.Build()
}

// ConvertDockerManifestListToMultiArch converts a Docker manifest list to MultiArchManifest
func ConvertDockerManifestListToMultiArch(dockerList *DockerManifestList) *MultiArchManifest {
	manifests := make([]PlatformManifest, len(dockerList.Manifests))
	for i, dm := range dockerList.Manifests {
		manifests[i] = PlatformManifest{
			MediaType:   dm.MediaType,
			Digest:      dm.Digest,
			Size:        dm.Size,
			URLs:        dm.URLs,
			Annotations: dm.Annotations,
		}
		if dm.Platform != nil {
			manifests[i].Platform = *dm.Platform
		}
	}

	return &MultiArchManifest{
		SchemaVersion: dockerList.SchemaVersion,
		MediaType:     dockerList.MediaType,
		Manifests:     manifests,
		Annotations:   dockerList.Annotations,
	}
}

// ConvertOCIImageIndexToMultiArch converts an OCI image index to MultiArchManifest
func ConvertOCIImageIndexToMultiArch(ociIndex *OCIImageIndex) *MultiArchManifest {
	manifests := make([]PlatformManifest, len(ociIndex.Manifests))
	for i, om := range ociIndex.Manifests {
		manifests[i] = PlatformManifest{
			MediaType:   om.MediaType,
			Digest:      om.Digest,
			Size:        om.Size,
			URLs:        om.URLs,
			Annotations: om.Annotations,
		}
		if om.Platform != nil {
			manifests[i].Platform = *om.Platform
		}
	}

	return &MultiArchManifest{
		SchemaVersion: ociIndex.SchemaVersion,
		MediaType:     ociIndex.MediaType,
		Manifests:     manifests,
		Annotations:   ociIndex.Annotations,
	}
}

// dockerManifestListToOCIIndex converts Docker manifest list to OCI image index
func (c *Converter) dockerManifestListToOCIIndex(dockerManifest []byte) ([]byte, error) {
	var dockerList DockerManifestList
	if err := json.Unmarshal(dockerManifest, &dockerList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Docker manifest list: %w", err)
	}

	ociIndex := OCIImageIndex{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests:     make([]OCIDescriptor, len(dockerList.Manifests)),
		Annotations:   dockerList.Annotations,
	}

	for i, manifest := range dockerList.Manifests {
		ociIndex.Manifests[i] = OCIDescriptor{
			MediaType:   c.convertDockerManifestMediaTypeToOCI(manifest.MediaType),
			Size:        manifest.Size,
			Digest:      manifest.Digest,
			URLs:        manifest.URLs,
			Annotations: manifest.Annotations,
			Platform:    manifest.Platform,
		}
	}

	return json.Marshal(ociIndex)
}

// ociIndexToDockerManifestList converts OCI image index to Docker manifest list
func (c *Converter) ociIndexToDockerManifestList(ociManifest []byte) ([]byte, error) {
	var ociIndex OCIImageIndex
	if err := json.Unmarshal(ociManifest, &ociIndex); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OCI image index: %w", err)
	}

	dockerList := DockerManifestList{
		SchemaVersion: 2,
		MediaType:     "application/vnd.docker.distribution.manifest.list.v2+json",
		Manifests:     make([]DockerManifestListDescriptor, len(ociIndex.Manifests)),
		Annotations:   ociIndex.Annotations,
	}

	for i, manifest := range ociIndex.Manifests {
		dockerList.Manifests[i] = DockerManifestListDescriptor{
			MediaType:   c.convertOCIManifestMediaTypeToDocker(manifest.MediaType),
			Size:        manifest.Size,
			Digest:      manifest.Digest,
			URLs:        manifest.URLs,
			Annotations: manifest.Annotations,
			Platform:    manifest.Platform,
		}
	}

	return json.Marshal(dockerList)
}

// convertDockerManifestMediaTypeToOCI converts Docker manifest media type to OCI
func (c *Converter) convertDockerManifestMediaTypeToOCI(dockerMediaType string) string {
	switch dockerMediaType {
	case "application/vnd.docker.distribution.manifest.v2+json":
		return "application/vnd.oci.image.manifest.v1+json"
	default:
		return dockerMediaType
	}
}

// convertOCIManifestMediaTypeToDocker converts OCI manifest media type to Docker
func (c *Converter) convertOCIManifestMediaTypeToDocker(ociMediaType string) string {
	switch ociMediaType {
	case "application/vnd.oci.image.manifest.v1+json":
		return "application/vnd.docker.distribution.manifest.v2+json"
	default:
		return ociMediaType
	}
}
