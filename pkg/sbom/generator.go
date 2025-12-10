package sbom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// SBOMFormat represents the output format for SBOM
type SBOMFormat string

const (
	FormatSPDX      SBOMFormat = "spdx"
	FormatCycloneDX SBOMFormat = "cyclonedx"
	FormatSyftJSON  SBOMFormat = "syft-json"
	FormatTable     SBOMFormat = "table"
)

// SBOM represents a Software Bill of Materials
type SBOM struct {
	// Metadata
	ImageRef    string            `json:"image_ref"`
	GeneratedAt time.Time         `json:"generated_at"`
	Format      SBOMFormat        `json:"format"`
	Metadata    map[string]string `json:"metadata"`

	// Packages discovered
	Packages []Package `json:"packages"`

	// Files cataloged
	Files []File `json:"files"`

	// Relationships between components
	Relationships []Relationship `json:"relationships"`

	// Secrets found (if enabled)
	Secrets []Secret `json:"secrets,omitempty"`
}

// Package represents a software package found in the image
type Package struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Type         string            `json:"type"` // deb, rpm, apk, npm, pip, go, etc.
	Language     string            `json:"language,omitempty"`
	PURL         string            `json:"purl"` // Package URL
	CPE          string            `json:"cpe,omitempty"`
	Licenses     []string          `json:"licenses,omitempty"`
	Locations    []Location        `json:"locations"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
}

// File represents a file in the image
type File struct {
	Path     string            `json:"path"`
	Size     int64             `json:"size"`
	Digest   string            `json:"digest"`
	MimeType string            `json:"mime_type,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Location represents where a package was found
type Location struct {
	Path      string `json:"path"`
	LayerID   string `json:"layer_id,omitempty"`
	LineRange string `json:"line_range,omitempty"`
}

// Relationship represents a relationship between components
type Relationship struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // contains, depends_on, etc.
}

// Secret represents a potential secret found in the image
type Secret struct {
	Type     string   `json:"type"` // api_key, password, token, etc.
	Location Location `json:"location"`
	Match    string   `json:"match"`
	Context  string   `json:"context"`
}

// GeneratorConfig configures the SBOM generator
type GeneratorConfig struct {
	// Output format
	Format SBOMFormat

	// Include file catalog
	IncludeFiles bool

	// Scan for secrets
	ScanSecrets bool

	// Include OS packages
	IncludeOSPackages bool

	// Include language-specific packages
	IncludeLanguagePackages bool

	// Supported package types
	PackageTypes []string

	// Output path (if writing to file)
	OutputPath string

	// Registry options
	RegistryOptions []remote.Option
}

// Generator generates SBOMs for container images
type Generator struct {
	config GeneratorConfig
}

// NewGenerator creates a new SBOM generator
func NewGenerator(config GeneratorConfig) *Generator {
	// Set defaults
	if config.Format == "" {
		config.Format = FormatSyftJSON
	}
	if len(config.PackageTypes) == 0 {
		config.PackageTypes = []string{"deb", "rpm", "apk", "npm", "pip", "go", "maven", "gem"}
	}

	return &Generator{
		config: config,
	}
}

// Generate creates an SBOM for the specified image
func (g *Generator) Generate(ctx context.Context, ref name.Reference) (*SBOM, error) {
	// Fetch image
	img, err := remote.Image(ref, g.config.RegistryOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	// Get image metadata
	configFile, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get image config: %w", err)
	}

	// Initialize SBOM
	sbom := &SBOM{
		ImageRef:    ref.String(),
		GeneratedAt: time.Now(),
		Format:      g.config.Format,
		Metadata: map[string]string{
			"architecture": configFile.Architecture,
			"os":           configFile.OS,
			"created":      configFile.Created.String(),
		},
		Packages:      []Package{},
		Files:         []File{},
		Relationships: []Relationship{},
	}

	// Extract packages from image
	if err := g.extractPackages(ctx, img, sbom); err != nil {
		return nil, fmt.Errorf("failed to extract packages: %w", err)
	}

	// Catalog files if enabled
	if g.config.IncludeFiles {
		if err := g.catalogFiles(ctx, img, sbom); err != nil {
			return nil, fmt.Errorf("failed to catalog files: %w", err)
		}
	}

	// Scan for secrets if enabled
	if g.config.ScanSecrets {
		if err := g.scanSecrets(ctx, img, sbom); err != nil {
			return nil, fmt.Errorf("failed to scan secrets: %w", err)
		}
	}

	return sbom, nil
}

// extractPackages extracts packages from the image
func (g *Generator) extractPackages(ctx context.Context, img v1.Image, sbom *SBOM) error {
	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("failed to get layers: %w", err)
	}

	for i, layer := range layers {
		layerDigest, err := layer.Digest()
		if err != nil {
			return fmt.Errorf("failed to get layer digest: %w", err)
		}

		// Extract layer to temporary directory
		tmpDir, err := os.MkdirTemp("", fmt.Sprintf("sbom-layer-%d-", i))
		if err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		// Uncompress layer
		rc, err := layer.Uncompressed()
		if err != nil {
			return fmt.Errorf("failed to uncompress layer: %w", err)
		}
		defer rc.Close()

		// Scan for packages in layer
		if g.config.IncludeOSPackages {
			g.scanDebianPackages(tmpDir, layerDigest.String(), sbom)
			g.scanRPMPackages(tmpDir, layerDigest.String(), sbom)
			g.scanAlpinePackages(tmpDir, layerDigest.String(), sbom)
		}

		if g.config.IncludeLanguagePackages {
			g.scanNPMPackages(tmpDir, layerDigest.String(), sbom)
			g.scanPythonPackages(tmpDir, layerDigest.String(), sbom)
			g.scanGoPackages(tmpDir, layerDigest.String(), sbom)
			g.scanMavenPackages(tmpDir, layerDigest.String(), sbom)
			g.scanRubyPackages(tmpDir, layerDigest.String(), sbom)
		}
	}

	return nil
}

// scanDebianPackages scans for Debian/Ubuntu packages
func (g *Generator) scanDebianPackages(rootDir, layerID string, sbom *SBOM) error {
	statusFile := filepath.Join(rootDir, "var/lib/dpkg/status")
	if _, err := os.Stat(statusFile); os.IsNotExist(err) {
		return nil // Not a Debian-based system
	}

	content, err := os.ReadFile(statusFile)
	if err != nil {
		return err
	}

	// Parse dpkg status file
	packages := g.parseDebianStatus(string(content), layerID)
	sbom.Packages = append(sbom.Packages, packages...)

	return nil
}

// parseDebianStatus parses Debian package status file
func (g *Generator) parseDebianStatus(content, layerID string) []Package {
	var packages []Package
	var currentPkg *Package

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if currentPkg != nil {
				packages = append(packages, *currentPkg)
				currentPkg = nil
			}
			continue
		}

		if currentPkg == nil {
			currentPkg = &Package{
				Type:      "deb",
				Locations: []Location{{LayerID: layerID}},
				Metadata:  make(map[string]string),
			}
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Package":
			currentPkg.Name = value
		case "Version":
			currentPkg.Version = value
		case "Architecture":
			currentPkg.Metadata["architecture"] = value
		}
	}

	if currentPkg != nil {
		packages = append(packages, *currentPkg)
	}

	return packages
}

// scanRPMPackages scans for RPM packages
func (g *Generator) scanRPMPackages(rootDir, layerID string, sbom *SBOM) error {
	rpmDir := filepath.Join(rootDir, "var/lib/rpm")
	if _, err := os.Stat(rpmDir); os.IsNotExist(err) {
		return nil // Not an RPM-based system
	}

	// In production, would use rpm database parsing
	// For now, placeholder implementation
	return nil
}

// scanAlpinePackages scans for Alpine APK packages
func (g *Generator) scanAlpinePackages(rootDir, layerID string, sbom *SBOM) error {
	installedFile := filepath.Join(rootDir, "lib/apk/db/installed")
	if _, err := os.Stat(installedFile); os.IsNotExist(err) {
		return nil // Not an Alpine system
	}

	content, err := os.ReadFile(installedFile)
	if err != nil {
		return err
	}

	packages := g.parseAlpineInstalled(string(content), layerID)
	sbom.Packages = append(sbom.Packages, packages...)

	return nil
}

// parseAlpineInstalled parses Alpine installed packages
func (g *Generator) parseAlpineInstalled(content, layerID string) []Package {
	var packages []Package
	var currentPkg *Package

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if currentPkg != nil {
				packages = append(packages, *currentPkg)
				currentPkg = nil
			}
			continue
		}

		if currentPkg == nil {
			currentPkg = &Package{
				Type:      "apk",
				Locations: []Location{{LayerID: layerID}},
				Metadata:  make(map[string]string),
			}
		}

		if strings.HasPrefix(line, "P:") {
			currentPkg.Name = strings.TrimPrefix(line, "P:")
		} else if strings.HasPrefix(line, "V:") {
			currentPkg.Version = strings.TrimPrefix(line, "V:")
		}
	}

	if currentPkg != nil {
		packages = append(packages, *currentPkg)
	}

	return packages
}

// scanNPMPackages scans for Node.js packages
func (g *Generator) scanNPMPackages(rootDir, layerID string, sbom *SBOM) error {
	// Search for package.json files
	var packageFiles []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "package.json" {
			packageFiles = append(packageFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, pkgFile := range packageFiles {
		pkg, err := g.parseNPMPackage(pkgFile, layerID)
		if err != nil {
			continue // Skip invalid package.json files
		}
		sbom.Packages = append(sbom.Packages, *pkg)
	}

	return nil
}

// parseNPMPackage parses a package.json file
func (g *Generator) parseNPMPackage(path, layerID string) (*Package, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkgJSON struct {
		Name         string            `json:"name"`
		Version      string            `json:"version"`
		License      string            `json:"license"`
		Dependencies map[string]string `json:"dependencies"`
	}

	if err := json.Unmarshal(content, &pkgJSON); err != nil {
		return nil, err
	}

	pkg := &Package{
		Name:     pkgJSON.Name,
		Version:  pkgJSON.Version,
		Type:     "npm",
		Language: "javascript",
		Locations: []Location{{
			Path:    path,
			LayerID: layerID,
		}},
		Metadata: make(map[string]string),
	}

	if pkgJSON.License != "" {
		pkg.Licenses = []string{pkgJSON.License}
	}

	// Add dependencies
	for dep := range pkgJSON.Dependencies {
		pkg.Dependencies = append(pkg.Dependencies, dep)
	}

	return pkg, nil
}

// scanPythonPackages scans for Python packages
func (g *Generator) scanPythonPackages(rootDir, layerID string, sbom *SBOM) error {
	// Search for requirements.txt, setup.py, etc.
	return nil // Placeholder
}

// scanGoPackages scans for Go modules
func (g *Generator) scanGoPackages(rootDir, layerID string, sbom *SBOM) error {
	// Search for go.mod files
	return nil // Placeholder
}

// scanMavenPackages scans for Maven packages
func (g *Generator) scanMavenPackages(rootDir, layerID string, sbom *SBOM) error {
	// Search for pom.xml files
	return nil // Placeholder
}

// scanRubyPackages scans for Ruby gems
func (g *Generator) scanRubyPackages(rootDir, layerID string, sbom *SBOM) error {
	// Search for Gemfile, *.gemspec
	return nil // Placeholder
}

// catalogFiles catalogs all files in the image
func (g *Generator) catalogFiles(ctx context.Context, img v1.Image, sbom *SBOM) error {
	// Placeholder - would walk through all layers and catalog files
	return nil
}

// scanSecrets scans for potential secrets in the image
func (g *Generator) scanSecrets(ctx context.Context, img v1.Image, sbom *SBOM) error {
	// Placeholder - would scan for patterns like API keys, passwords, etc.
	return nil
}

// Export exports the SBOM in the specified format
func (g *Generator) Export(sbom *SBOM, format SBOMFormat) ([]byte, error) {
	switch format {
	case FormatSyftJSON:
		return g.exportSyftJSON(sbom)
	case FormatSPDX:
		return g.exportSPDX(sbom)
	case FormatCycloneDX:
		return g.exportCycloneDX(sbom)
	case FormatTable:
		return g.exportTable(sbom)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// exportSyftJSON exports SBOM in Syft JSON format
func (g *Generator) exportSyftJSON(sbom *SBOM) ([]byte, error) {
	return json.MarshalIndent(sbom, "", "  ")
}

// exportSPDX exports SBOM in SPDX format
func (g *Generator) exportSPDX(sbom *SBOM) ([]byte, error) {
	// SPDX 2.3 format
	spdx := map[string]interface{}{
		"spdxVersion":       "SPDX-2.3",
		"dataLicense":       "CC0-1.0",
		"SPDXID":            "SPDXRef-DOCUMENT",
		"name":              sbom.ImageRef,
		"documentNamespace": fmt.Sprintf("https://sbom.freightliner/%s", sbom.ImageRef),
		"creationInfo": map[string]interface{}{
			"created": sbom.GeneratedAt.Format(time.RFC3339),
			"creators": []string{
				"Tool: freightliner-sbom",
			},
		},
		"packages": g.convertToSPDXPackages(sbom),
	}

	return json.MarshalIndent(spdx, "", "  ")
}

// exportCycloneDX exports SBOM in CycloneDX format
func (g *Generator) exportCycloneDX(sbom *SBOM) ([]byte, error) {
	// CycloneDX 1.4 format
	cdx := map[string]interface{}{
		"bomFormat":   "CycloneDX",
		"specVersion": "1.4",
		"version":     1,
		"metadata": map[string]interface{}{
			"timestamp": sbom.GeneratedAt.Format(time.RFC3339),
			"component": map[string]interface{}{
				"type": "container",
				"name": sbom.ImageRef,
			},
		},
		"components": g.convertToCycloneDXComponents(sbom),
	}

	return json.MarshalIndent(cdx, "", "  ")
}

// exportTable exports SBOM as a human-readable table
func (g *Generator) exportTable(sbom *SBOM) ([]byte, error) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("SBOM for: %s\n", sbom.ImageRef))
	builder.WriteString(fmt.Sprintf("Generated: %s\n\n", sbom.GeneratedAt.Format(time.RFC3339)))

	builder.WriteString("PACKAGES:\n")
	builder.WriteString(strings.Repeat("-", 80) + "\n")
	builder.WriteString(fmt.Sprintf("%-40s %-15s %-10s\n", "NAME", "VERSION", "TYPE"))
	builder.WriteString(strings.Repeat("-", 80) + "\n")

	for _, pkg := range sbom.Packages {
		builder.WriteString(fmt.Sprintf("%-40s %-15s %-10s\n", pkg.Name, pkg.Version, pkg.Type))
	}

	builder.WriteString(fmt.Sprintf("\nTotal Packages: %d\n", len(sbom.Packages)))

	return []byte(builder.String()), nil
}

// convertToSPDXPackages converts packages to SPDX format
func (g *Generator) convertToSPDXPackages(sbom *SBOM) []map[string]interface{} {
	var packages []map[string]interface{}
	for i, pkg := range sbom.Packages {
		packages = append(packages, map[string]interface{}{
			"SPDXID":           fmt.Sprintf("SPDXRef-Package-%d", i),
			"name":             pkg.Name,
			"versionInfo":      pkg.Version,
			"downloadLocation": "NOASSERTION",
			"filesAnalyzed":    false,
		})
	}
	return packages
}

// convertToCycloneDXComponents converts packages to CycloneDX format
func (g *Generator) convertToCycloneDXComponents(sbom *SBOM) []map[string]interface{} {
	var components []map[string]interface{}
	for _, pkg := range sbom.Packages {
		components = append(components, map[string]interface{}{
			"type":    "library",
			"name":    pkg.Name,
			"version": pkg.Version,
			"purl":    pkg.PURL,
		})
	}
	return components
}

// WriteTo writes the SBOM to a writer
func (g *Generator) WriteTo(sbom *SBOM, w io.Writer, format SBOMFormat) error {
	data, err := g.Export(sbom, format)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// WriteToFile writes the SBOM to a file
func (g *Generator) WriteToFile(sbom *SBOM, path string, format SBOMFormat) error {
	data, err := g.Export(sbom, format)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
