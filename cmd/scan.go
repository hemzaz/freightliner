package cmd

import (
	"fmt"
	"os"

	"freightliner/pkg/vulnerability"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"
)

var (
	scanFailOn        string
	scanOnlyFixed     bool
	scanIgnoreUnfixed bool
	scanOutputFormat  string
	scanOutput        string
	scanUseGrype      bool
	scanScope         string
	scanPlatform      string
	scanExclude       []string
	scanDBUpdate      bool
	scanPolicyPath    string
)

// newScanCmd creates a new scan command
func newScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan [image]",
		Short: "Scan container images for vulnerabilities",
		Long: `Scan container images for known security vulnerabilities using CVE databases.
Supports multiple output formats and severity-based filtering.

Examples:
  # Scan image and fail on high severity vulnerabilities
  freightliner scan docker.io/library/nginx:latest --fail-on high

  # Scan with only fixed vulnerabilities
  freightliner scan gcr.io/myproject/app:v1.0.0 --only-fixed

  # Scan using Grype (if installed)
  freightliner scan myregistry.com/app:latest --use-grype

  # Generate SARIF report for GitHub
  freightliner scan nginx:latest --format sarif --output results.sarif

  # Update vulnerability database before scanning
  freightliner scan alpine:latest --db-update`,
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
				"image":   imageRef,
				"fail_on": scanFailOn,
			}).Info("Scanning image for vulnerabilities")

			var report *vulnerability.Report

			// Check if we should use Grype
			if scanUseGrype {
				if !vulnerability.IsGrypeInstalled() {
					fmt.Println("Error: Grype is not installed")
					fmt.Println(vulnerability.InstallGrype())
					os.Exit(1)
				}

				grypeScanner, err := vulnerability.NewGrypeScanner("")
				if err != nil {
					logger.Error("Failed to initialize Grype scanner", err)
					fmt.Printf("Error: %s\n", err)
					os.Exit(1)
				}

				report, err = grypeScanner.ScanWithGrype(ctx, ref, vulnerability.GrypeConfig{
					FailOn:        vulnerability.Severity(scanFailOn),
					OnlyFixed:     scanOnlyFixed,
					IgnoreUnfixed: scanIgnoreUnfixed,
					OutputFormat:  "json",
					Scope:         scanScope,
					Platform:      scanPlatform,
					Exclude:       scanExclude,
					DBUpdateCheck: scanDBUpdate,
				})
				if err != nil {
					logger.Error("Failed to scan image with Grype", err)
					fmt.Printf("Error: %s\n", err)
					os.Exit(1)
				}
			} else {
				// Use built-in scanner
				scanConfig := vulnerability.ScanConfig{
					FailOnSeverity: vulnerability.Severity(scanFailOn),
					IgnoreUnfixed:  scanIgnoreUnfixed,
					OnlyFixed:      scanOnlyFixed,
					Platform:       scanPlatform,
					OutputFormat:   scanOutputFormat,
					AutoUpdateDB:   scanDBUpdate,
				}

				// Load policy if specified
				if scanPolicyPath != "" {
					// Policy loading would be implemented here
					logger.WithFields(map[string]interface{}{
						"policy": scanPolicyPath,
					}).Info("Loading scan policy")
				}

				scanner, err := vulnerability.NewScanner(scanConfig)
				if err != nil {
					logger.Error("Failed to initialize scanner", err)
					fmt.Printf("Error: %s\n", err)
					os.Exit(1)
				}

				report, err = scanner.Scan(ctx, ref)
				if err != nil {
					logger.Error("Failed to scan image", err)
					fmt.Printf("Error: %s\n", err)
					os.Exit(1)
				}
			}

			// Export report in specified format
			var output []byte
			switch scanOutputFormat {
			case "json":
				output, err = report.ExportJSON()
			case "sarif":
				output, err = report.ExportSARIF()
			case "table", "":
				output = []byte(report.ExportTable())
			default:
				fmt.Printf("Error: unsupported output format: %s\n", scanOutputFormat)
				os.Exit(1)
			}

			if err != nil {
				logger.Error("Failed to export report", err)
				fmt.Printf("Error: %s\n", err)
				os.Exit(1)
			}

			// Write output
			if scanOutput != "" {
				if err := os.WriteFile(scanOutput, output, 0644); err != nil {
					logger.Error("Failed to write report to file", err)
					fmt.Printf("Error: %s\n", err)
					os.Exit(1)
				}
				fmt.Printf("Scan report written to: %s\n", scanOutput)
			} else {
				fmt.Println(string(output))
			}

			logger.WithFields(map[string]interface{}{
				"total_vulnerabilities": report.Summary.TotalVulnerabilities,
				"critical":              report.Summary.Critical,
				"high":                  report.Summary.High,
				"medium":                report.Summary.Medium,
				"low":                   report.Summary.Low,
			}).Info("Vulnerability scan complete")

			// Check if we should fail based on severity
			if scanFailOn != "" {
				failSeverity := vulnerability.Severity(scanFailOn)
				shouldFail := false

				switch failSeverity {
				case vulnerability.SeverityCritical:
					shouldFail = report.Summary.Critical > 0
				case vulnerability.SeverityHigh:
					shouldFail = report.Summary.Critical > 0 || report.Summary.High > 0
				case vulnerability.SeverityMedium:
					shouldFail = report.Summary.Critical > 0 || report.Summary.High > 0 || report.Summary.Medium > 0
				case vulnerability.SeverityLow:
					shouldFail = report.Summary.TotalVulnerabilities > 0
				}

				if shouldFail {
					fmt.Printf("\nScan failed: vulnerabilities found at or above '%s' severity level\n", scanFailOn)
					os.Exit(1)
				}
			}

			// Check policy result if available
			if report.PolicyResult != nil && !report.PolicyResult.Passed {
				fmt.Printf("\nPolicy evaluation failed: %d violations\n", report.PolicyResult.FailureCount)
				os.Exit(1)
			}

			fmt.Println("\nScan completed successfully")
		},
	}

	// Add flags
	cmd.Flags().StringVar(&scanFailOn, "fail-on", "",
		"Fail if vulnerabilities are found at or above this severity (critical, high, medium, low)")
	cmd.Flags().BoolVar(&scanOnlyFixed, "only-fixed", false,
		"Only report vulnerabilities that have fixes available")
	cmd.Flags().BoolVar(&scanIgnoreUnfixed, "ignore-unfixed", false,
		"Ignore vulnerabilities without fixes")
	cmd.Flags().StringVarP(&scanOutputFormat, "format", "f", "table",
		"Output format (json, table, sarif)")
	cmd.Flags().StringVarP(&scanOutput, "output", "o", "",
		"Output file path (default: stdout)")
	cmd.Flags().BoolVar(&scanUseGrype, "use-grype", false,
		"Use Grype CLI for vulnerability scanning (must be installed)")
	cmd.Flags().StringVar(&scanScope, "scope", "squashed",
		"Scope of the scan (all-layers, squashed)")
	cmd.Flags().StringVar(&scanPlatform, "platform", "",
		"Platform to scan (e.g., linux/amd64)")
	cmd.Flags().StringArrayVar(&scanExclude, "exclude", []string{},
		"Paths to exclude from scan")
	cmd.Flags().BoolVar(&scanDBUpdate, "db-update", false,
		"Update vulnerability database before scanning")
	cmd.Flags().StringVar(&scanPolicyPath, "policy", "",
		"Path to vulnerability policy file")

	return cmd
}
