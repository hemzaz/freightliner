package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"freightliner/pkg/config"
	"freightliner/pkg/formatting"
	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	inspectShowConfig bool
	inspectRaw        bool
	inspectFormat     string
)

// ImageInspectResult represents the complete inspection result
type ImageInspectResult struct {
	Name         string            `json:"name" yaml:"name"`
	Digest       string            `json:"digest" yaml:"digest"`
	MediaType    string            `json:"mediaType" yaml:"mediaType"`
	Size         int64             `json:"size" yaml:"size"`
	Config       *v1.ConfigFile    `json:"config,omitempty" yaml:"config,omitempty"`
	Manifest     interface{}       `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	Layers       []LayerInfo       `json:"layers" yaml:"layers"`
	RepoTags     []string          `json:"repoTags,omitempty" yaml:"repoTags,omitempty"`
	Architecture string            `json:"architecture" yaml:"architecture"`
	OS           string            `json:"os" yaml:"os"`
	Created      string            `json:"created" yaml:"created"`
	Author       string            `json:"author,omitempty" yaml:"author,omitempty"`
	Env          []string          `json:"env,omitempty" yaml:"env,omitempty"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// LayerInfo represents information about an image layer
type LayerInfo struct {
	Digest    string `json:"digest" yaml:"digest"`
	Size      int64  `json:"size" yaml:"size"`
	MediaType string `json:"mediaType" yaml:"mediaType"`
}

// newInspectCmd creates the inspect command
func newInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect SOURCE",
		Short: "Inspect image manifest and metadata without pulling",
		Long: `Inspect a container image's manifest, configuration, and metadata without downloading the image.

SOURCE format: [TRANSPORT://]IMAGE[:TAG|@DIGEST]

Supported transports:
  docker://     Docker registry (default)
  docker-daemon: Docker daemon
  oci:          OCI layout directory

Examples:
  # Inspect image from Docker Hub
  freightliner inspect docker://nginx:latest

  # Inspect with authentication
  freightliner inspect docker://registry.io/private/image:v1.0

  # Show raw manifest
  freightliner inspect --raw docker://nginx:latest

  # Show config in JSON format
  freightliner inspect --config --format json docker://nginx:latest

  # Inspect from local Docker daemon
  freightliner inspect docker-daemon:nginx:latest
`,
		Args: cobra.ExactArgs(1),
		RunE: runInspect,
	}

	cmd.Flags().BoolVar(&inspectShowConfig, "config", false, "Show container configuration")
	cmd.Flags().BoolVar(&inspectRaw, "raw", false, "Show raw manifest JSON")
	cmd.Flags().StringVar(&inspectFormat, "format", "table", "Output format: table, json, yaml, or Go template")

	return cmd
}

// runInspect executes the inspect command
func runInspect(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger, ctx, cancel := setupCommand(ctx)
	defer cancel()

	source := args[0]

	// Parse the image reference
	transport, imageRef, err := parseImageReference(source)
	if err != nil {
		return fmt.Errorf("failed to parse image reference: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"image":     imageRef,
		"transport": transport,
	}).Info("Inspecting image")

	// Perform inspection based on transport
	var result *ImageInspectResult
	switch transport {
	case "docker", "":
		result, err = inspectDockerImage(ctx, logger, imageRef)
	case "docker-daemon":
		result, err = inspectDockerDaemonImage(ctx, logger, imageRef)
	case "oci":
		result, err = inspectOCIImage(ctx, logger, imageRef)
	default:
		return fmt.Errorf("unsupported transport: %s", transport)
	}

	if err != nil {
		return fmt.Errorf("failed to inspect image: %w", err)
	}

	// Output results based on format
	return outputInspectResult(result, inspectFormat, inspectRaw, inspectShowConfig)
}

// inspectDockerImage inspects an image from a Docker registry
func inspectDockerImage(ctx context.Context, logger log.Logger, imageRef string) (*ImageInspectResult, error) {
	// Parse the reference
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference: %w", err)
	}

	// Get authentication
	auth, err := getAuthForRegistry(ref.Context().RegistryStr())
	if err != nil {
		logger.WithFields(map[string]interface{}{"error": err.Error()}).Warn("Using anonymous authentication")
		auth = authn.Anonymous
	}

	// Fetch the descriptor
	desc, err := remote.Get(ref, remote.WithAuth(auth), remote.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get image descriptor: %w", err)
	}

	result := &ImageInspectResult{
		Name:      imageRef,
		Digest:    desc.Digest.String(),
		MediaType: string(desc.MediaType),
		Size:      desc.Size,
	}

	// Try to get the image
	img, err := desc.Image()
	if err != nil {
		logger.WithFields(map[string]interface{}{"error": err.Error()}).Warn("Could not parse as image")
		// Try as manifest index
		idx, err := desc.ImageIndex()
		if err != nil {
			return result, nil // Return what we have
		}
		return inspectImageIndex(idx, result)
	}

	// Get config
	configFile, err := img.ConfigFile()
	if err != nil {
		logger.WithFields(map[string]interface{}{"error": err.Error()}).Warn("Could not get config file")
	} else {
		result.Config = configFile
		result.Architecture = configFile.Architecture
		result.OS = configFile.OS
		result.Created = configFile.Created.String()
		result.Author = configFile.Author
		if configFile.Config.Env != nil {
			result.Env = configFile.Config.Env
		}
		if configFile.Config.Labels != nil {
			result.Labels = configFile.Config.Labels
		}
	}

	// Get manifest
	manifest, err := img.Manifest()
	if err != nil {
		logger.WithFields(map[string]interface{}{"error": err.Error()}).Warn("Could not get manifest")
	} else {
		result.Manifest = manifest
	}

	// Get layers
	layers, err := img.Layers()
	if err != nil {
		logger.WithFields(map[string]interface{}{"error": err.Error()}).Warn("Could not get layers")
	} else {
		for _, layer := range layers {
			digest, _ := layer.Digest()
			size, _ := layer.Size()
			mediaType, _ := layer.MediaType()

			result.Layers = append(result.Layers, LayerInfo{
				Digest:    digest.String(),
				Size:      size,
				MediaType: string(mediaType),
			})
		}
	}

	return result, nil
}

// inspectDockerDaemonImage inspects an image from the local Docker daemon
func inspectDockerDaemonImage(ctx context.Context, logger log.Logger, imageRef string) (*ImageInspectResult, error) {
	// This would require Docker daemon integration
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("docker-daemon transport not yet implemented")
}

// inspectOCIImage inspects an image from an OCI layout directory
func inspectOCIImage(ctx context.Context, logger log.Logger, imageRef string) (*ImageInspectResult, error) {
	// This would require OCI layout parsing
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("oci transport not yet implemented")
}

// inspectImageIndex inspects a manifest index (multi-platform image)
func inspectImageIndex(idx v1.ImageIndex, result *ImageInspectResult) (*ImageInspectResult, error) {
	manifest, err := idx.IndexManifest()
	if err != nil {
		return result, fmt.Errorf("failed to get index manifest: %w", err)
	}

	result.Manifest = manifest

	// Extract platform information from manifests
	var platforms []string
	for _, m := range manifest.Manifests {
		if m.Platform != nil {
			platform := fmt.Sprintf("%s/%s", m.Platform.OS, m.Platform.Architecture)
			if m.Platform.Variant != "" {
				platform += "/" + m.Platform.Variant
			}
			platforms = append(platforms, platform)
		}
	}

	result.RepoTags = platforms

	return result, nil
}

// outputInspectResult outputs the inspection result in the specified format
func outputInspectResult(result *ImageInspectResult, format string, raw bool, showConfig bool) error {
	if raw {
		// Output raw manifest
		if result.Manifest != nil {
			data, err := json.MarshalIndent(result.Manifest, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal manifest: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}
		return fmt.Errorf("no manifest available")
	}

	if showConfig && result.Config == nil {
		return fmt.Errorf("no config available for this image")
	}

	// Check if format is a Go template
	if formatting.IsTemplateFormat(format) {
		formatter, err := formatting.NewTemplateFormatter(format)
		if err != nil {
			return fmt.Errorf("invalid template: %w", err)
		}
		if err := formatter.Format(os.Stdout, result); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		fmt.Println() // Add newline after template output
		return nil
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		fmt.Println(string(data))

	case "yaml":
		data, err := yaml.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		fmt.Print(string(data))

	case "table":
		outputInspectTable(result, showConfig)

	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json, yaml, or Go template starting with {{)", format)
	}

	return nil
}

// outputInspectTable outputs the inspection result as a formatted table
func outputInspectTable(result *ImageInspectResult, showConfig bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "Name:\t%s\n", result.Name)
	fmt.Fprintf(w, "Digest:\t%s\n", result.Digest)
	fmt.Fprintf(w, "MediaType:\t%s\n", result.MediaType)
	fmt.Fprintf(w, "Size:\t%d bytes\n", result.Size)

	if result.Architecture != "" {
		fmt.Fprintf(w, "Architecture:\t%s\n", result.Architecture)
	}
	if result.OS != "" {
		fmt.Fprintf(w, "OS:\t%s\n", result.OS)
	}
	if result.Created != "" {
		fmt.Fprintf(w, "Created:\t%s\n", result.Created)
	}
	if result.Author != "" {
		fmt.Fprintf(w, "Author:\t%s\n", result.Author)
	}

	// Layers
	if len(result.Layers) > 0 {
		fmt.Fprintf(w, "\nLayers:\t%d\n", len(result.Layers))
		for i, layer := range result.Layers {
			fmt.Fprintf(w, "  [%d]:\t%s (%d bytes)\n", i, layer.Digest, layer.Size)
		}
	}

	// Environment variables
	if len(result.Env) > 0 {
		fmt.Fprintf(w, "\nEnvironment:\n")
		for _, env := range result.Env {
			fmt.Fprintf(w, "  %s\n", env)
		}
	}

	// Labels
	if len(result.Labels) > 0 {
		fmt.Fprintf(w, "\nLabels:\n")
		for key, value := range result.Labels {
			fmt.Fprintf(w, "  %s:\t%s\n", key, value)
		}
	}

	// RepoTags (for multi-platform images)
	if len(result.RepoTags) > 0 {
		fmt.Fprintf(w, "\nPlatforms:\n")
		for _, tag := range result.RepoTags {
			fmt.Fprintf(w, "  %s\n", tag)
		}
	}

	// Config (if requested)
	if showConfig && result.Config != nil {
		fmt.Fprintf(w, "\nConfiguration:\n")
		configData, _ := json.MarshalIndent(result.Config, "  ", "  ")
		fmt.Fprintf(w, "%s\n", string(configData))
	}
}

// getAuthForRegistry returns authentication for a given registry
func getAuthForRegistry(registry string) (authn.Authenticator, error) {
	// Try to use registry configuration if available
	if cfg != nil && len(cfg.Registries.Registries) > 0 {
		// Find matching registry in the loaded config
		for _, r := range cfg.Registries.Registries {
			host, _ := r.GetRegistryHost()
			if host == registry {
				return getAuthenticatorFromConfig(&r)
			}
		}
	}

	// Fall back to anonymous authentication
	return authn.Anonymous, nil
}

// getAuthenticatorFromConfig creates an authenticator from registry config
func getAuthenticatorFromConfig(r *config.RegistryConfig) (authn.Authenticator, error) {
	switch r.Auth.Type {
	case config.AuthTypeBasic:
		return &authn.Basic{
			Username: r.Auth.Username,
			Password: r.Auth.Password,
		}, nil
	case config.AuthTypeToken:
		return &authn.Bearer{
			Token: r.Auth.Token,
		}, nil
	case config.AuthTypeAnonymous:
		return authn.Anonymous, nil
	default:
		// Default to anonymous if no specific auth type
		return authn.Anonymous, nil
	}
}

// parseImageReference parses an image reference and extracts the transport
func parseImageReference(ref string) (transport string, imageRef string, err error) {
	parts := strings.SplitN(ref, "://", 2)
	if len(parts) == 1 {
		// No transport specified, default to docker
		return "docker", ref, nil
	}
	return parts[0], parts[1], nil
}
