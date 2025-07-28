// Test Manifest CLI tool for managing selective test execution
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"freightliner/pkg/testing"
)

const usage = `Test Manifest CLI - Selective Test Execution Tool

Usage:
  test-manifest [command] [options]

Commands:
  summary              Show test manifest summary
  test [package]       Run tests for specific package with manifest filtering
  generate-args [pkg]  Generate go test arguments for a package
  validate             Validate the test manifest file
  list-categories      List all test categories
  list-packages        List all packages in manifest

Options:
  -manifest string     Path to test manifest file (default: test-manifest.yaml)
  -env string          Override environment detection (ci|local|integration)
  -categories string   Comma-separated list of categories to filter by
  -dry-run            Show what would be executed without running tests
  -verbose            Show detailed output

Examples:
  test-manifest summary
  test-manifest test freightliner/pkg/client/gcr
  test-manifest test -categories unit,integration
  test-manifest generate-args freightliner/pkg/replication
  test-manifest validate -manifest custom-manifest.yaml
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	command := os.Args[1]

	// Common flags
	var (
		manifestPath = flag.String("manifest", testing.GetDefaultManifestPath(), "Path to test manifest file")
		environment  = flag.String("env", "", "Override environment detection")
		categories   = flag.String("categories", "", "Comma-separated list of categories to filter by")
		dryRun       = flag.Bool("dry-run", false, "Show what would be executed without running")
		verbose      = flag.Bool("verbose", false, "Show detailed output")
	)

	// Parse flags starting from the command
	if err := flag.CommandLine.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Load manifest
	manifest, err := testing.LoadTestManifest(*manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
		os.Exit(1)
	}

	// Create filter
	filter := testing.NewTestFilter(manifest)

	if *environment != "" {
		filter = filter.WithEnvironment(*environment)
	}

	if *categories != "" {
		categoryList := strings.Split(*categories, ",")
		for i, cat := range categoryList {
			categoryList[i] = strings.TrimSpace(cat)
		}
		filter = filter.WithCategories(categoryList)
	}

	// Execute command
	switch command {
	case "summary":
		filter.PrintSummary()

	case "test":
		if flag.NArg() < 1 {
			fmt.Fprintf(os.Stderr, "Error: package name required for test command\n")
			os.Exit(1)
		}
		packageName := flag.Arg(0)
		runTests(filter, packageName, *dryRun, *verbose)

	case "generate-args":
		if flag.NArg() < 1 {
			fmt.Fprintf(os.Stderr, "Error: package name required for generate-args command\n")
			os.Exit(1)
		}
		packageName := flag.Arg(0)
		generateArgs(filter, packageName)

	case "validate":
		validateManifest(manifest)

	case "list-categories":
		listCategories(manifest)

	case "list-packages":
		listPackages(manifest)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Print(usage)
		os.Exit(1)
	}
}

func runTests(filter *testing.TestFilter, packageName string, dryRun bool, verbose bool) {
	args := filter.GenerateTestArgs(packageName)

	// Build the go test command
	cmd := []string{"go", "test"}
	if verbose {
		cmd = append(cmd, "-v")
	}
	cmd = append(cmd, args...)
	cmd = append(cmd, packageName)

	if dryRun {
		fmt.Printf("Would execute: %s\n", strings.Join(cmd, " "))
		return
	}

	fmt.Printf("Running: %s\n", strings.Join(cmd, " "))

	// Execute the command
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error running tests: %v\n", err)
		os.Exit(1)
	}
}

func generateArgs(filter *testing.TestFilter, packageName string) {
	args := filter.GenerateTestArgs(packageName)
	fmt.Println(strings.Join(args, " "))
}

func validateManifest(manifest *testing.TestManifest) {
	fmt.Printf("Validating manifest version %s\n", manifest.Version)

	// Basic validation
	if manifest.Version == "" {
		fmt.Fprintf(os.Stderr, "Warning: No version specified in manifest\n")
	}

	if len(manifest.Packages) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: No packages defined in manifest\n")
	}

	packageCount := 0
	testCount := 0
	enabledTests := 0

	for packageName, pkgConfig := range manifest.Packages {
		packageCount++
		fmt.Printf("Package: %s (%d tests)\n", packageName, len(pkgConfig.Tests))

		for testName, testConfig := range pkgConfig.Tests {
			testCount++
			if testConfig.Enabled {
				enabledTests++
			}

			// Validate categories exist
			for _, category := range testConfig.Categories {
				if _, exists := manifest.Categories[category]; !exists {
					fmt.Fprintf(os.Stderr, "Warning: Unknown category '%s' in test %s:%s\n",
						category, packageName, testName)
				}
			}
		}
	}

	fmt.Printf("\nValidation Summary:\n")
	fmt.Printf("  Packages: %d\n", packageCount)
	fmt.Printf("  Total tests: %d\n", testCount)
	fmt.Printf("  Enabled tests: %d\n", enabledTests)
	fmt.Printf("  Disabled tests: %d\n", testCount-enabledTests)
	fmt.Printf("  Categories: %d\n", len(manifest.Categories))

	fmt.Println("✓ Manifest validation completed")
}

func listCategories(manifest *testing.TestManifest) {
	fmt.Println("Test Categories:")
	fmt.Println(strings.Repeat("=", 50))

	for categoryName, categoryConfig := range manifest.Categories {
		fmt.Printf("\n%s\n", categoryName)
		fmt.Printf("  Description: %s\n", categoryConfig.Description)
		fmt.Printf("  Enabled in: %v\n", categoryConfig.EnabledIn)
		fmt.Printf("  Disabled in: %v\n", categoryConfig.DisabledIn)
	}
}

func listPackages(manifest *testing.TestManifest) {
	fmt.Println("Packages in Manifest:")
	fmt.Println(strings.Repeat("=", 50))

	for packageName, pkgConfig := range manifest.Packages {
		enabledCount := 0
		for _, testConfig := range pkgConfig.Tests {
			if testConfig.Enabled {
				enabledCount++
			}
		}

		status := "✓"
		if !pkgConfig.Enabled {
			status = "✗"
		}

		fmt.Printf("\n%s %s\n", status, packageName)
		fmt.Printf("  Description: %s\n", pkgConfig.Description)
		fmt.Printf("  Tests: %d total, %d enabled\n", len(pkgConfig.Tests), enabledCount)
	}
}
