package cmd

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"text/tabwriter"

	"freightliner/pkg/transport"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	manifestFormat    string
	manifestAlgorithm string
)

// newManifestCmd creates the manifest command
func newManifestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Manifest operations",
		Long: `Operations on container image manifests.

Available subcommands:
  digest    - Calculate manifest digest
  inspect   - Inspect manifest contents`,
	}

	cmd.AddCommand(newManifestDigestCmd())
	cmd.AddCommand(newManifestInspectCmd())

	return cmd
}

// newManifestDigestCmd creates the manifest digest command
func newManifestDigestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "digest IMAGE",
		Short: "Calculate manifest digest",
		Long: `Calculate and display the manifest digest for an image.

The digest is calculated using the specified algorithm (default: sha256).

Examples:
  # Calculate sha256 digest
  freightliner manifest digest docker://nginx:latest

  # Calculate sha512 digest
  freightliner manifest digest --algorithm sha512 docker://nginx:latest

  # Calculate digest for OCI image
  freightliner manifest digest oci:/tmp/image`,
		Args: cobra.ExactArgs(1),
		RunE: runManifestDigest,
	}

	cmd.Flags().StringVar(&manifestAlgorithm, "algorithm", "sha256", "Digest algorithm (sha256, sha512)")

	return cmd
}

// newManifestInspectCmd creates the manifest inspect command
func newManifestInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect IMAGE",
		Short: "Inspect manifest contents",
		Long: `Display detailed manifest information.

Shows the complete manifest structure including schema version, media type,
config information, and layer details.

Examples:
  # Inspect manifest as table
  freightliner manifest inspect docker://nginx:latest

  # Inspect manifest as JSON
  freightliner manifest inspect --format json docker://nginx:latest

  # Inspect manifest as YAML
  freightliner manifest inspect --format yaml docker://nginx:latest`,
		Args: cobra.ExactArgs(1),
		RunE: runManifestInspect,
	}

	cmd.Flags().StringVar(&manifestFormat, "format", "table", "Output format (table, json, yaml)")

	return cmd
}

// runManifestDigest executes the manifest digest command
func runManifestDigest(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger, ctx, cancel := setupCommand(ctx)
	defer cancel()

	source := args[0]

	logger.WithFields(map[string]interface{}{
		"image":     source,
		"algorithm": manifestAlgorithm,
	}).Info("Calculating manifest digest")

	// Parse reference
	ref, err := transport.ParseReference(source)
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}

	// Create image source
	imageSource, err := ref.NewImageSource(ctx)
	if err != nil {
		return fmt.Errorf("failed to create image source: %w", err)
	}
	defer imageSource.Close()

	// Get manifest
	manifest, _, err := imageSource.GetManifest(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}

	// Calculate digest
	var h hash.Hash
	switch manifestAlgorithm {
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return fmt.Errorf("unsupported algorithm: %s (supported: sha256, sha512)", manifestAlgorithm)
	}

	_, err = h.Write(manifest)
	if err != nil {
		return fmt.Errorf("failed to calculate digest: %w", err)
	}

	digest := fmt.Sprintf("%s:%x", manifestAlgorithm, h.Sum(nil))

	// Output digest
	fmt.Println(digest)

	logger.WithFields(map[string]interface{}{
		"image":  source,
		"digest": digest,
	}).Info("Calculated manifest digest")

	return nil
}

// ManifestInspectResult represents manifest inspection output
type ManifestInspectResult struct {
	SchemaVersion int                    `json:"schemaVersion" yaml:"schemaVersion"`
	MediaType     string                 `json:"mediaType" yaml:"mediaType"`
	Config        ManifestConfigInfo     `json:"config" yaml:"config"`
	Layers        []ManifestLayerInfo    `json:"layers" yaml:"layers"`
	Annotations   map[string]string      `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Raw           map[string]interface{} `json:"-" yaml:"-"`
}

// ManifestConfigInfo represents config information in manifest
type ManifestConfigInfo struct {
	MediaType string `json:"mediaType" yaml:"mediaType"`
	Size      int64  `json:"size" yaml:"size"`
	Digest    string `json:"digest" yaml:"digest"`
}

// ManifestLayerInfo represents layer information in manifest
type ManifestLayerInfo struct {
	MediaType string `json:"mediaType" yaml:"mediaType"`
	Size      int64  `json:"size" yaml:"size"`
	Digest    string `json:"digest" yaml:"digest"`
}

// runManifestInspect executes the manifest inspect command
func runManifestInspect(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger, ctx, cancel := setupCommand(ctx)
	defer cancel()

	source := args[0]

	logger.WithFields(map[string]interface{}{
		"image":  source,
		"format": manifestFormat,
	}).Info("Inspecting manifest")

	// Parse reference
	ref, err := transport.ParseReference(source)
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}

	// Create image source
	imageSource, err := ref.NewImageSource(ctx)
	if err != nil {
		return fmt.Errorf("failed to create image source: %w", err)
	}
	defer imageSource.Close()

	// Get manifest
	manifestBytes, mediaType, err := imageSource.GetManifest(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}

	// Parse manifest JSON
	var rawManifest map[string]interface{}
	if err := json.Unmarshal(manifestBytes, &rawManifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Build result
	result := ManifestInspectResult{
		MediaType: mediaType,
		Raw:       rawManifest,
	}

	// Extract schema version
	if sv, ok := rawManifest["schemaVersion"].(float64); ok {
		result.SchemaVersion = int(sv)
	}

	// Extract config
	if configMap, ok := rawManifest["config"].(map[string]interface{}); ok {
		result.Config = ManifestConfigInfo{
			MediaType: getStringField(configMap, "mediaType"),
			Digest:    getStringField(configMap, "digest"),
		}
		if size, ok := configMap["size"].(float64); ok {
			result.Config.Size = int64(size)
		}
	}

	// Extract layers
	if layersArray, ok := rawManifest["layers"].([]interface{}); ok {
		for _, layerInterface := range layersArray {
			if layerMap, ok := layerInterface.(map[string]interface{}); ok {
				layer := ManifestLayerInfo{
					MediaType: getStringField(layerMap, "mediaType"),
					Digest:    getStringField(layerMap, "digest"),
				}
				if size, ok := layerMap["size"].(float64); ok {
					layer.Size = int64(size)
				}
				result.Layers = append(result.Layers, layer)
			}
		}
	}

	// Extract annotations if present
	if annotations, ok := rawManifest["annotations"].(map[string]interface{}); ok {
		result.Annotations = make(map[string]string)
		for k, v := range annotations {
			if str, ok := v.(string); ok {
				result.Annotations[k] = str
			}
		}
	}

	// Output result
	return outputManifestInspectResult(&result, manifestFormat)
}

// outputManifestInspectResult outputs the manifest inspection result
func outputManifestInspectResult(result *ManifestInspectResult, format string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)

	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(result)

	case "table":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		defer w.Flush()

		fmt.Fprintf(w, "Schema Version:\t%d\n", result.SchemaVersion)
		fmt.Fprintf(w, "Media Type:\t%s\n", result.MediaType)
		fmt.Fprintf(w, "\n")

		fmt.Fprintf(w, "Config:\n")
		fmt.Fprintf(w, "  Digest:\t%s\n", result.Config.Digest)
		fmt.Fprintf(w, "  Size:\t%s\n", formatBytesManifest(result.Config.Size))
		fmt.Fprintf(w, "  Media Type:\t%s\n", result.Config.MediaType)
		fmt.Fprintf(w, "\n")

		fmt.Fprintf(w, "Layers:\n")
		for i, layer := range result.Layers {
			fmt.Fprintf(w, "  [%d] %s\t(%s)\n", i, layer.Digest, formatBytesManifest(layer.Size))
		}

		if len(result.Annotations) > 0 {
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "Annotations:\n")
			for k, v := range result.Annotations {
				fmt.Fprintf(w, "  %s:\t%s\n", k, v)
			}
		}

		return nil

	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", format)
	}
}

// getStringField safely extracts a string field from a map
func getStringField(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// formatBytesManifest formats bytes to human-readable string for manifest output
func formatBytesManifest(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
