package cmd

import (
	"fmt"
	"os"

	"freightliner/pkg/sbom"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var (
	sbomFormat       string
	sbomOutput       string
	sbomIncludeFiles bool
	sbomScanSecrets  bool
	sbomUseSyft      bool
	sbomScope        string
	sbomExclude      []string
)

// newSBOMCmd creates a new sbom command
func newSBOMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sbom [image]",
		Short: "Generate Software Bill of Materials (SBOM) for container images",
		Long: `Generate a comprehensive Software Bill of Materials (SBOM) for container images.
Supports multiple output formats including SPDX, CycloneDX, and Syft JSON.

Examples:
  # Generate SBOM in SPDX format
  freightliner sbom docker.io/library/alpine:latest --format spdx

  # Generate SBOM using Syft (if installed)
  freightliner sbom gcr.io/myproject/myapp:v1.0.0 --use-syft

  # Save SBOM to file
  freightliner sbom nginx:latest --format cyclonedx --output sbom.json

  # Include file catalog and secret scanning
  freightliner sbom myregistry.com/app:latest --include-files --scan-secrets`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger and context
			logger, ctx, cancel := setupCommand(cmd.Context())
			defer cancel()

			// Parse image reference
			imageRef := args[0]
			ref, err := name.ParseReference(imageRef)
			if err != nil {
				logger.Error("Failed to parse image reference", err)
				fmt.Printf("Error: invalid image reference: %s\n", err)
				os.Exit(1)
			}

			logger.WithFields(map[string]interface{}{
				"image":  imageRef,
				"format": sbomFormat,
			}).Info("Generating SBOM")

			var imageSBOM *sbom.SBOM

			// Check if we should use Syft
			if sbomUseSyft {
				if !sbom.IsSyftInstalled() {
					fmt.Println("Error: Syft is not installed")
					fmt.Println(sbom.InstallSyft())
					os.Exit(1)
				}

				syftGen, err := sbom.NewSyftGenerator()
				if err != nil {
					logger.Error("Failed to initialize Syft generator", err)
					fmt.Printf("Error: %s\n", err)
					os.Exit(1)
				}

				imageSBOM, err = syftGen.GenerateWithSyft(ctx, ref, sbom.SyftConfig{
					Format:  sbom.SBOMFormat(sbomFormat),
					Scope:   sbomScope,
					Exclude: sbomExclude,
				})
				if err != nil {
					logger.Error("Failed to generate SBOM with Syft", err)
					fmt.Printf("Error: %s\n", err)
					os.Exit(1)
				}
			} else {
				// Use built-in generator
				generator := sbom.NewGenerator(sbom.GeneratorConfig{
					Format:                  sbom.SBOMFormat(sbomFormat),
					IncludeFiles:            sbomIncludeFiles,
					ScanSecrets:             sbomScanSecrets,
					IncludeOSPackages:       true,
					IncludeLanguagePackages: true,
					RegistryOptions:         buildRegistryOptions(),
				})

				imageSBOM, err = generator.Generate(ctx, ref)
				if err != nil {
					logger.Error("Failed to generate SBOM", err)
					fmt.Printf("Error: %s\n", err)
					os.Exit(1)
				}

				// Export in specified format
				data, err := generator.Export(imageSBOM, sbom.SBOMFormat(sbomFormat))
				if err != nil {
					logger.Error("Failed to export SBOM", err)
					fmt.Printf("Error: %s\n", err)
					os.Exit(1)
				}

				// Write output
				if sbomOutput != "" {
					if err := os.WriteFile(sbomOutput, data, 0644); err != nil {
						logger.Error("Failed to write SBOM to file", err)
						fmt.Printf("Error: %s\n", err)
						os.Exit(1)
					}
					fmt.Printf("SBOM written to: %s\n", sbomOutput)
				} else {
					fmt.Println(string(data))
				}

				logger.WithFields(map[string]interface{}{
					"packages": len(imageSBOM.Packages),
					"files":    len(imageSBOM.Files),
				}).Info("SBOM generation complete")

				return
			}

			// Handle Syft-generated SBOM
			if sbomOutput != "" {
				generator := sbom.NewGenerator(sbom.GeneratorConfig{})
				if err := generator.WriteToFile(imageSBOM, sbomOutput, sbom.SBOMFormat(sbomFormat)); err != nil {
					logger.Error("Failed to write SBOM to file", err)
					fmt.Printf("Error: %s\n", err)
					os.Exit(1)
				}
				fmt.Printf("SBOM written to: %s\n", sbomOutput)
			}

			logger.WithFields(map[string]interface{}{
				"packages": len(imageSBOM.Packages),
			}).Info("SBOM generation complete")

			fmt.Printf("\nSBOM generated successfully\n")
			fmt.Printf("Total packages: %d\n", len(imageSBOM.Packages))
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&sbomFormat, "format", "f", "syft-json",
		"Output format (spdx, cyclonedx, syft-json, table)")
	cmd.Flags().StringVarP(&sbomOutput, "output", "o", "",
		"Output file path (default: stdout)")
	cmd.Flags().BoolVar(&sbomIncludeFiles, "include-files", false,
		"Include file catalog in SBOM")
	cmd.Flags().BoolVar(&sbomScanSecrets, "scan-secrets", false,
		"Scan for potential secrets in the image")
	cmd.Flags().BoolVar(&sbomUseSyft, "use-syft", false,
		"Use Syft CLI for SBOM generation (must be installed)")
	cmd.Flags().StringVar(&sbomScope, "scope", "squashed",
		"Scope of the scan (all-layers, squashed)")
	cmd.Flags().StringArrayVar(&sbomExclude, "exclude", []string{},
		"Paths to exclude from scan")

	return cmd
}

// buildRegistryOptions builds registry options from config
func buildRegistryOptions() []remote.Option {
	var opts []remote.Option

	// Registry auth is handled by the client factory
	// This is a placeholder for additional registry options
	// Future: add timeout, retry, TLS config, etc.

	return opts
}
