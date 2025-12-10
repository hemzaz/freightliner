package sbom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
)

// SyftGenerator wraps the Syft CLI for SBOM generation
// This provides integration with the actual Syft tool if available
type SyftGenerator struct {
	syftPath string
	timeout  time.Duration
}

// NewSyftGenerator creates a new Syft-based generator
func NewSyftGenerator() (*SyftGenerator, error) {
	// Check if syft is installed
	syftPath, err := exec.LookPath("syft")
	if err != nil {
		return nil, fmt.Errorf("syft not found in PATH: %w", err)
	}

	return &SyftGenerator{
		syftPath: syftPath,
		timeout:  5 * time.Minute,
	}, nil
}

// SyftConfig configures Syft execution
type SyftConfig struct {
	// Output format
	Format SBOMFormat

	// Scope: all-layers, squashed
	Scope string

	// Additional catalogers to enable
	Catalogers []string

	// Exclude patterns
	Exclude []string

	// Registry authentication
	RegistryUser     string
	RegistryPassword string
	RegistryToken    string

	// Timeout for scan
	Timeout time.Duration
}

// GenerateWithSyft generates an SBOM using the Syft CLI
func (sg *SyftGenerator) GenerateWithSyft(ctx context.Context, ref name.Reference, config SyftConfig) (*SBOM, error) {
	// Build syft command
	args := sg.buildSyftArgs(ref, config)

	// Create command with timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = sg.timeout
	}

	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, sg.syftPath, args...)

	// Set environment variables for registry auth
	if config.RegistryUser != "" && config.RegistryPassword != "" {
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("SYFT_REGISTRY_AUTH_USERNAME=%s", config.RegistryUser),
			fmt.Sprintf("SYFT_REGISTRY_AUTH_PASSWORD=%s", config.RegistryPassword),
		)
	} else if config.RegistryToken != "" {
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("SYFT_REGISTRY_AUTH_TOKEN=%s", config.RegistryToken),
		)
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute syft
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("syft execution failed: %w, stderr: %s", err, stderr.String())
	}

	// Parse syft output
	sbom, err := sg.parseSyftOutput(stdout.Bytes(), ref.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse syft output: %w", err)
	}

	return sbom, nil
}

// buildSyftArgs builds command line arguments for syft
func (sg *SyftGenerator) buildSyftArgs(ref name.Reference, config SyftConfig) []string {
	args := []string{
		"scan",
		ref.String(),
	}

	// Output format
	format := config.Format
	if format == "" {
		format = FormatSyftJSON
	}
	args = append(args, "-o", string(format))

	// Scope
	if config.Scope != "" {
		args = append(args, "--scope", config.Scope)
	}

	// Catalogers
	if len(config.Catalogers) > 0 {
		args = append(args, "--catalogers", strings.Join(config.Catalogers, ","))
	}

	// Exclude patterns
	for _, pattern := range config.Exclude {
		args = append(args, "--exclude", pattern)
	}

	return args
}

// parseSyftOutput parses Syft JSON output into our SBOM structure
func (sg *SyftGenerator) parseSyftOutput(data []byte, imageRef string) (*SBOM, error) {
	var syftOutput struct {
		Artifacts []struct {
			Name      string   `json:"name"`
			Version   string   `json:"version"`
			Type      string   `json:"type"`
			Language  string   `json:"language"`
			Licenses  []string `json:"licenses"`
			PURL      string   `json:"purl"`
			CPE       []string `json:"cpes"`
			Locations []struct {
				Path      string `json:"path"`
				LayerID   string `json:"layerID"`
				LineRange string `json:"lineRange"`
			} `json:"locations"`
			Metadata map[string]interface{} `json:"metadata"`
		} `json:"artifacts"`
		ArtifactRelationships []struct {
			Parent string `json:"parent"`
			Child  string `json:"child"`
			Type   string `json:"type"`
		} `json:"artifactRelationships"`
		Source struct {
			Type   string `json:"type"`
			Target string `json:"target"`
		} `json:"source"`
		Descriptor struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"descriptor"`
	}

	if err := json.Unmarshal(data, &syftOutput); err != nil {
		return nil, fmt.Errorf("failed to unmarshal syft output: %w", err)
	}

	// Convert to our SBOM structure
	sbom := &SBOM{
		ImageRef:      imageRef,
		GeneratedAt:   time.Now(),
		Format:        FormatSyftJSON,
		Packages:      []Package{},
		Relationships: []Relationship{},
		Metadata: map[string]string{
			"generator":         syftOutput.Descriptor.Name,
			"generator_version": syftOutput.Descriptor.Version,
		},
	}

	// Convert artifacts to packages
	for _, artifact := range syftOutput.Artifacts {
		pkg := Package{
			Name:     artifact.Name,
			Version:  artifact.Version,
			Type:     artifact.Type,
			Language: artifact.Language,
			PURL:     artifact.PURL,
			Licenses: artifact.Licenses,
			Metadata: make(map[string]string),
		}

		// Add CPE if available
		if len(artifact.CPE) > 0 {
			pkg.CPE = artifact.CPE[0]
		}

		// Convert locations
		for _, loc := range artifact.Locations {
			pkg.Locations = append(pkg.Locations, Location{
				Path:      loc.Path,
				LayerID:   loc.LayerID,
				LineRange: loc.LineRange,
			})
		}

		// Convert metadata
		for k, v := range artifact.Metadata {
			if str, ok := v.(string); ok {
				pkg.Metadata[k] = str
			}
		}

		sbom.Packages = append(sbom.Packages, pkg)
	}

	// Convert relationships
	for _, rel := range syftOutput.ArtifactRelationships {
		sbom.Relationships = append(sbom.Relationships, Relationship{
			From: rel.Parent,
			To:   rel.Child,
			Type: rel.Type,
		})
	}

	return sbom, nil
}

// IsSyftInstalled checks if Syft is available
func IsSyftInstalled() bool {
	_, err := exec.LookPath("syft")
	return err == nil
}

// GetSyftVersion returns the installed Syft version
func GetSyftVersion() (string, error) {
	cmd := exec.Command("syft", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get syft version: %w", err)
	}

	// Parse version from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Version:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return strings.TrimSpace(string(output)), nil
}

// InstallSyft provides instructions for installing Syft
func InstallSyft() string {
	return `
Syft is not installed. To install Syft, use one of the following methods:

# macOS (Homebrew)
brew install syft

# Linux (curl)
curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin

# Docker
docker pull anchore/syft:latest

For more information, visit: https://github.com/anchore/syft
`
}

// SyftScanOptions provides additional scanning options
type SyftScanOptions struct {
	// Platform to scan (e.g., linux/amd64)
	Platform string

	// Include only specific package types
	OnlyPackageTypes []string

	// Generate attestation
	GenerateAttestation bool

	// SBOM quality
	SelectCatalogers []string
}

// AdvancedScan performs an advanced scan with custom options
func (sg *SyftGenerator) AdvancedScan(ctx context.Context, ref name.Reference, options SyftScanOptions) (*SBOM, error) {
	args := []string{"scan", ref.String(), "-o", "syft-json"}

	if options.Platform != "" {
		args = append(args, "--platform", options.Platform)
	}

	if len(options.OnlyPackageTypes) > 0 {
		args = append(args, "--catalogers", strings.Join(options.SelectCatalogers, ","))
	}

	cmd := exec.CommandContext(ctx, sg.syftPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("syft scan failed: %w, stderr: %s", err, stderr.String())
	}

	return sg.parseSyftOutput(stdout.Bytes(), ref.String())
}

// ComparePackages compares two SBOMs and returns differences
func ComparePackages(oldSBOM, newSBOM *SBOM) *PackageDiff {
	diff := &PackageDiff{
		Added:   []Package{},
		Removed: []Package{},
		Changed: []PackageChange{},
	}

	// Create maps for quick lookup
	oldPkgs := make(map[string]Package)
	for _, pkg := range oldSBOM.Packages {
		key := fmt.Sprintf("%s:%s", pkg.Name, pkg.Type)
		oldPkgs[key] = pkg
	}

	newPkgs := make(map[string]Package)
	for _, pkg := range newSBOM.Packages {
		key := fmt.Sprintf("%s:%s", pkg.Name, pkg.Type)
		newPkgs[key] = pkg
	}

	// Find added and changed packages
	for key, newPkg := range newPkgs {
		if oldPkg, exists := oldPkgs[key]; exists {
			if oldPkg.Version != newPkg.Version {
				diff.Changed = append(diff.Changed, PackageChange{
					Name:       newPkg.Name,
					Type:       newPkg.Type,
					OldVersion: oldPkg.Version,
					NewVersion: newPkg.Version,
				})
			}
		} else {
			diff.Added = append(diff.Added, newPkg)
		}
	}

	// Find removed packages
	for key, oldPkg := range oldPkgs {
		if _, exists := newPkgs[key]; !exists {
			diff.Removed = append(diff.Removed, oldPkg)
		}
	}

	return diff
}

// PackageDiff represents differences between two SBOMs
type PackageDiff struct {
	Added   []Package
	Removed []Package
	Changed []PackageChange
}

// PackageChange represents a changed package
type PackageChange struct {
	Name       string
	Type       string
	OldVersion string
	NewVersion string
}

// Summary returns a summary of the diff
func (pd *PackageDiff) Summary() string {
	return fmt.Sprintf("Added: %d, Removed: %d, Changed: %d",
		len(pd.Added), len(pd.Removed), len(pd.Changed))
}
