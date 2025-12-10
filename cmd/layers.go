package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"freightliner/pkg/transport"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	layersFormat string
)

// newLayersCmd creates the layers command
func newLayersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "layers IMAGE",
		Short: "Display image layer information",
		Long: `Show detailed information about image layers.

Displays layer digests, sizes, and media types for a container image.
This is useful for understanding the layer structure and identifying
large layers.

Examples:
  # Display layers as table
  freightliner layers docker://nginx:latest

  # Display layers as JSON
  freightliner layers --format json docker://nginx:latest

  # Display layers as YAML
  freightliner layers --format yaml docker://nginx:latest`,
		Args: cobra.ExactArgs(1),
		RunE: runLayers,
	}

	cmd.Flags().StringVar(&layersFormat, "format", "table", "Output format (table, json, yaml)")

	return cmd
}

// LayersResult represents the layers command output
type LayersResult struct {
	Image      string            `json:"image" yaml:"image"`
	TotalSize  int64             `json:"totalSize" yaml:"totalSize"`
	LayerCount int               `json:"layerCount" yaml:"layerCount"`
	Layers     []LayerDetailInfo `json:"layers" yaml:"layers"`
}

// LayerDetailInfo represents detailed layer information for display
type LayerDetailInfo struct {
	Digest    string `json:"digest" yaml:"digest"`
	Size      int64  `json:"size" yaml:"size"`
	MediaType string `json:"mediaType" yaml:"mediaType"`
}

// runLayers executes the layers command
func runLayers(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger, ctx, cancel := setupCommand(ctx)
	defer cancel()

	source := args[0]

	logger.WithFields(map[string]interface{}{
		"image":  source,
		"format": layersFormat,
	}).Info("Retrieving layer information")

	// Parse reference
	ref, err := transport.ParseReference(source)
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}

	// Create image
	image, err := ref.NewImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	defer image.Close()

	// Get layer infos from image
	layerInfos := image.LayerInfos()

	// Build result
	result := LayersResult{
		Image:      source,
		LayerCount: len(layerInfos),
		Layers:     make([]LayerDetailInfo, 0, len(layerInfos)),
	}

	// Extract layer information
	for _, layerInfo := range layerInfos {
		detail := LayerDetailInfo{
			Digest:    layerInfo.Digest,
			Size:      layerInfo.Size,
			MediaType: layerInfo.MediaType,
		}

		result.Layers = append(result.Layers, detail)
		result.TotalSize += layerInfo.Size
	}

	// Output result
	return outputLayersResult(&result, layersFormat)
}

// outputLayersResult outputs the layers result
func outputLayersResult(result *LayersResult, format string) error {
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

		fmt.Fprintf(w, "Image:\t%s\n", result.Image)
		fmt.Fprintf(w, "Total Size:\t%s (%d bytes)\n", formatBytes(result.TotalSize), result.TotalSize)
		fmt.Fprintf(w, "Layer Count:\t%d\n", result.LayerCount)
		fmt.Fprintf(w, "\n")

		fmt.Fprintf(w, "Layer\tDigest\tSize\n")
		fmt.Fprintf(w, "-----\t------\t----\n")

		for i, layer := range result.Layers {
			fmt.Fprintf(w, "[%d]\t%s\t%s\n", i, layer.Digest, formatBytes(layer.Size))
		}

		return nil

	case "simple":
		// Simple format: one digest per line
		for _, layer := range result.Layers {
			fmt.Println(layer.Digest)
		}
		return nil

	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json, yaml, simple)", format)
	}
}
