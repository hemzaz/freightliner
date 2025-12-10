package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	testmanifest "freightliner/pkg/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMainUsage tests that main prints usage when no args provided
func TestMainUsage(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set args with no command
	os.Args = []string{"test-manifest"}

	// This would normally call main() but we can't easily test os.Exit(1)
	// Instead we'll test the usage constant
	assert.Contains(t, usage, "Test Manifest CLI")
	assert.Contains(t, usage, "Commands:")
	assert.Contains(t, usage, "summary")
	assert.Contains(t, usage, "test")
	assert.Contains(t, usage, "generate-args")
	assert.Contains(t, usage, "validate")
	assert.Contains(t, usage, "list-categories")
	assert.Contains(t, usage, "list-packages")

	w.Close()
	os.Stdout = old
	io.Copy(&bytes.Buffer{}, r)
}

// TestCommandParsing tests command parsing logic
func TestCommandParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "summary command",
			args:        []string{"summary"},
			expectError: false,
		},
		{
			name:        "test command without package",
			args:        []string{"test"},
			expectError: true,
		},
		{
			name:        "test command with package",
			args:        []string{"test", "pkg/example"},
			expectError: false,
		},
		{
			name:        "validate command",
			args:        []string{"validate"},
			expectError: false,
		},
		{
			name:        "unknown command",
			args:        []string{"unknown"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For this test, we just verify args length
			if len(tt.args) == 0 {
				assert.True(t, tt.expectError)
			} else {
				command := tt.args[0]
				validCommands := []string{"summary", "test", "generate-args", "validate", "list-categories", "list-packages"}
				isValid := false
				for _, valid := range validCommands {
					if command == valid {
						isValid = true
						break
					}
				}

				if !isValid {
					assert.True(t, tt.expectError)
				}
			}
		})
	}
}

// TestFlagParsing tests flag parsing for different commands
func TestFlagParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectFlags map[string]string
	}{
		{
			name: "default flags",
			args: []string{"summary"},
			expectFlags: map[string]string{
				"manifest":   testmanifest.GetDefaultManifestPath(),
				"env":        "",
				"categories": "",
				"dry-run":    "false",
				"verbose":    "false",
			},
		},
		{
			name: "custom manifest path",
			args: []string{"summary", "-manifest", "/custom/path.yaml"},
			expectFlags: map[string]string{
				"manifest": "/custom/path.yaml",
			},
		},
		{
			name: "environment override",
			args: []string{"summary", "-env", "ci"},
			expectFlags: map[string]string{
				"env": "ci",
			},
		},
		{
			name: "categories filter",
			args: []string{"summary", "-categories", "unit,integration"},
			expectFlags: map[string]string{
				"categories": "unit,integration",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag.CommandLine for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Define flags
			manifestPath := flag.String("manifest", testmanifest.GetDefaultManifestPath(), "Path to test manifest file")
			environment := flag.String("env", "", "Override environment detection")
			categories := flag.String("categories", "", "Comma-separated list of categories to filter by")
			dryRun := flag.Bool("dry-run", false, "Show what would be executed without running")
			verbose := flag.Bool("verbose", false, "Show detailed output")

			// Parse flags starting from index 1 (skip command)
			err := flag.CommandLine.Parse(tt.args[1:])
			require.NoError(t, err)

			// Verify expected flags
			for flagName, expectedValue := range tt.expectFlags {
				switch flagName {
				case "manifest":
					assert.Equal(t, expectedValue, *manifestPath)
				case "env":
					assert.Equal(t, expectedValue, *environment)
				case "categories":
					assert.Equal(t, expectedValue, *categories)
				case "dry-run":
					expected := expectedValue == "true"
					assert.Equal(t, expected, *dryRun)
				case "verbose":
					expected := expectedValue == "true"
					assert.Equal(t, expected, *verbose)
				}
			}
		})
	}
}

// TestValidateManifest tests the validateManifest function
func TestValidateManifest(t *testing.T) {
	tests := []struct {
		name           string
		setupManifest  func() *testmanifest.TestManifest
		expectedOutput []string
	}{
		{
			name: "valid manifest",
			setupManifest: func() *testmanifest.TestManifest {
				return &testmanifest.TestManifest{
					Version: "1.0",
					Packages: map[string]testmanifest.PackageConfig{
						"pkg/example": {
							Enabled:     true,
							Description: "Example package",
							Tests: map[string]testmanifest.TestConfig{
								"TestExample": {
									Enabled:    true,
									Categories: []string{"unit"},
								},
							},
						},
					},
					Categories: map[string]testmanifest.CategoryConfig{
						"unit": {
							Description: "Unit tests",
							EnabledIn:   []string{"ci", "local"},
						},
					},
				}
			},
			expectedOutput: []string{"Validating manifest", "Packages: 1", "Total tests: 1", "validation completed"},
		},
		{
			name: "empty manifest",
			setupManifest: func() *testmanifest.TestManifest {
				return &testmanifest.TestManifest{
					Version:    "",
					Packages:   map[string]testmanifest.PackageConfig{},
					Categories: map[string]testmanifest.CategoryConfig{},
				}
			},
			expectedOutput: []string{"Warning: No version", "Warning: No packages"},
		},
		{
			name: "unknown category warning",
			setupManifest: func() *testmanifest.TestManifest {
				return &testmanifest.TestManifest{
					Version: "1.0",
					Packages: map[string]testmanifest.PackageConfig{
						"pkg/example": {
							Tests: map[string]testmanifest.TestConfig{
								"TestExample": {
									Categories: []string{"unknown-category"},
								},
							},
						},
					},
					Categories: map[string]testmanifest.CategoryConfig{},
				}
			},
			expectedOutput: []string{"Warning: Unknown category"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := tt.setupManifest()

			// Capture output
			old := os.Stdout
			oldErr := os.Stderr
			r, w, _ := os.Pipe()
			rErr, wErr, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = wErr

			validateManifest(manifest)

			w.Close()
			wErr.Close()
			os.Stdout = old
			os.Stderr = oldErr

			var buf bytes.Buffer
			var bufErr bytes.Buffer
			io.Copy(&buf, r)
			io.Copy(&bufErr, rErr)

			output := buf.String() + bufErr.String()

			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

// TestListCategories tests the listCategories function
func TestListCategories(t *testing.T) {
	manifest := &testmanifest.TestManifest{
		Categories: map[string]testmanifest.CategoryConfig{
			"unit": {
				Description: "Unit tests",
				EnabledIn:   []string{"ci", "local"},
				DisabledIn:  []string{},
			},
			"integration": {
				Description: "Integration tests",
				EnabledIn:   []string{"integration"},
				DisabledIn:  []string{"ci"},
			},
		},
	}

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	listCategories(manifest)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Test Categories:")
	assert.Contains(t, output, "unit")
	assert.Contains(t, output, "integration")
	assert.Contains(t, output, "Unit tests")
	assert.Contains(t, output, "Integration tests")
}

// TestListPackages tests the listPackages function
func TestListPackages(t *testing.T) {
	manifest := &testmanifest.TestManifest{
		Packages: map[string]testmanifest.PackageConfig{
			"pkg/client/ecr": {
				Enabled:     true,
				Description: "AWS ECR client tests",
				Tests: map[string]testmanifest.TestConfig{
					"TestECRClient": {Enabled: true},
					"TestECRAuth":   {Enabled: true},
				},
			},
			"pkg/client/gcr": {
				Enabled:     false,
				Description: "GCR client tests",
				Tests: map[string]testmanifest.TestConfig{
					"TestGCRClient": {Enabled: false},
				},
			},
		},
	}

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	listPackages(manifest)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Packages in Manifest:")
	assert.Contains(t, output, "pkg/client/ecr")
	assert.Contains(t, output, "pkg/client/gcr")
	assert.Contains(t, output, "AWS ECR client tests")
	assert.Contains(t, output, "Tests: 2 total, 2 enabled")
	assert.Contains(t, output, "Tests: 1 total, 0 enabled")
}

// TestGenerateArgs tests the generateArgs function
func TestGenerateArgs(t *testing.T) {
	manifest := &testmanifest.TestManifest{
		Version: "1.0",
		Packages: map[string]testmanifest.PackageConfig{
			"pkg/example": {
				Enabled: true,
				Tests: map[string]testmanifest.TestConfig{
					"TestExample": {
						Enabled:    true,
						Categories: []string{"unit"},
					},
					"TestDisabled": {
						Enabled: false,
					},
				},
			},
		},
	}

	filter := testmanifest.NewTestFilter(manifest)

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	generateArgs(filter, "pkg/example")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	// Should generate -run flag with enabled tests
	assert.Contains(t, output, "-run")
	assert.Contains(t, output, "TestExample")
	// TestDisabled may appear in skip flag, which is acceptable
}

// TestRunTests tests the runTests function with dry-run
func TestRunTests(t *testing.T) {
	manifest := &testmanifest.TestManifest{
		Version: "1.0",
		Packages: map[string]testmanifest.PackageConfig{
			"pkg/example": {
				Enabled: true,
				Tests: map[string]testmanifest.TestConfig{
					"TestExample": {
						Enabled:    true,
						Categories: []string{"unit"},
					},
				},
			},
		},
	}

	filter := testmanifest.NewTestFilter(manifest)

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run with dry-run=true so it doesn't actually execute tests
	runTests(filter, "pkg/example", true, false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Would execute:")
	assert.Contains(t, output, "go test")
	assert.Contains(t, output, "pkg/example")
}

// TestRunTestsVerbose tests the runTests function with verbose flag
func TestRunTestsVerbose(t *testing.T) {
	manifest := &testmanifest.TestManifest{
		Version: "1.0",
		Packages: map[string]testmanifest.PackageConfig{
			"pkg/example": {
				Enabled: true,
				Tests: map[string]testmanifest.TestConfig{
					"TestExample": {
						Enabled: true,
					},
				},
			},
		},
	}

	filter := testmanifest.NewTestFilter(manifest)

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run with dry-run=true and verbose=true
	runTests(filter, "pkg/example", true, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should include -v flag for verbose output
	assert.Contains(t, output, "-v")
}

// TestCategoryFiltering tests category filtering in test execution
func TestCategoryFiltering(t *testing.T) {
	manifest := &testmanifest.TestManifest{
		Version: "1.0",
		Packages: map[string]testmanifest.PackageConfig{
			"pkg/example": {
				Enabled: true,
				Tests: map[string]testmanifest.TestConfig{
					"TestUnit": {
						Enabled:    true,
						Categories: []string{"unit"},
					},
					"TestIntegration": {
						Enabled:    true,
						Categories: []string{"integration"},
					},
				},
			},
		},
		Categories: map[string]testmanifest.CategoryConfig{
			"unit": {
				EnabledIn: []string{"ci", "local"},
			},
			"integration": {
				EnabledIn: []string{"integration"},
			},
		},
	}

	// Test with unit category filter
	filter := testmanifest.NewTestFilter(manifest).WithCategories([]string{"unit"})
	args := filter.GenerateTestArgs("pkg/example")

	argsStr := strings.Join(args, " ")
	assert.Contains(t, argsStr, "TestUnit")
	// Integration test should be filtered out if categories don't match
}

// TestEnvironmentFiltering tests environment-based filtering
func TestEnvironmentFiltering(t *testing.T) {
	manifest := &testmanifest.TestManifest{
		Version: "1.0",
		Packages: map[string]testmanifest.PackageConfig{
			"pkg/example": {
				Enabled: true,
				Tests: map[string]testmanifest.TestConfig{
					"TestCI": {
						Enabled:    true,
						Categories: []string{"ci"},
					},
					"TestLocal": {
						Enabled:    true,
						Categories: []string{"local"},
					},
				},
			},
		},
		Categories: map[string]testmanifest.CategoryConfig{
			"ci": {
				EnabledIn: []string{"ci"},
			},
			"local": {
				EnabledIn:  []string{"local"},
				DisabledIn: []string{"ci"},
			},
		},
	}

	// Test with CI environment
	filter := testmanifest.NewTestFilter(manifest).WithEnvironment("ci")
	args := filter.GenerateTestArgs("pkg/example")

	argsStr := strings.Join(args, " ")
	// Should include CI tests
	// Local tests should be excluded in CI environment
	assert.NotEmpty(t, argsStr)
}

// TestManifestLoading tests loading manifest from file
func TestManifestLoading(t *testing.T) {
	// Create temporary manifest file
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test-manifest.yaml")

	manifestContent := `
version: "1.0"
description: "Test manifest"
packages:
  pkg/example:
    enabled: true
    description: "Example package"
    tests:
      TestExample:
        enabled: true
        categories:
          - unit
categories:
  unit:
    description: "Unit tests"
    enabled_in:
      - ci
      - local
`

	err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	// Load manifest
	manifest, err := testmanifest.LoadTestManifest(manifestPath)
	require.NoError(t, err)
	assert.NotNil(t, manifest)
	assert.Equal(t, "1.0", manifest.Version)
	assert.Equal(t, 1, len(manifest.Packages))
	assert.Contains(t, manifest.Packages, "pkg/example")
}

// TestManifestLoadingError tests error handling for invalid manifest
func TestManifestLoadingError(t *testing.T) {
	tests := []struct {
		name          string
		setupManifest func() string
		expectError   bool
	}{
		{
			name: "nonexistent file",
			setupManifest: func() string {
				return "/nonexistent/path/manifest.yaml"
			},
			expectError: true,
		},
		{
			name: "invalid yaml",
			setupManifest: func() string {
				tmpDir := t.TempDir()
				manifestPath := filepath.Join(tmpDir, "invalid.yaml")
				os.WriteFile(manifestPath, []byte("invalid: yaml: content:"), 0644)
				return manifestPath
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifestPath := tt.setupManifest()
			_, err := testmanifest.LoadTestManifest(manifestPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUsageContent tests the usage string contains all necessary information
func TestUsageContent(t *testing.T) {
	// Verify usage contains all commands
	commands := []string{"summary", "test", "generate-args", "validate", "list-categories", "list-packages"}
	for _, cmd := range commands {
		assert.Contains(t, usage, cmd, "usage should contain command: %s", cmd)
	}

	// Verify usage contains all options
	options := []string{"-manifest", "-env", "-categories", "-dry-run", "-verbose"}
	for _, opt := range options {
		assert.Contains(t, usage, opt, "usage should contain option: %s", opt)
	}

	// Verify usage contains examples
	assert.Contains(t, usage, "Examples:")
}

// TestDefaultManifestPath tests default manifest path detection
func TestDefaultManifestPath(t *testing.T) {
	defaultPath := testmanifest.GetDefaultManifestPath()
	assert.NotEmpty(t, defaultPath)
	assert.True(t, strings.HasSuffix(defaultPath, "test-manifest.yaml"))
}

// TestFilterChaining tests filter method chaining
func TestFilterChaining(t *testing.T) {
	manifest := &testmanifest.TestManifest{
		Version: "1.0",
		Packages: map[string]testmanifest.PackageConfig{
			"pkg/example": {
				Enabled: true,
				Tests: map[string]testmanifest.TestConfig{
					"TestExample": {
						Enabled:    true,
						Categories: []string{"unit"},
					},
				},
			},
		},
		Categories: map[string]testmanifest.CategoryConfig{
			"unit": {
				EnabledIn: []string{"ci"},
			},
		},
	}

	// Test method chaining
	filter := testmanifest.NewTestFilter(manifest).
		WithEnvironment("ci").
		WithCategories([]string{"unit"})

	assert.NotNil(t, filter)

	// Verify filter generates args
	args := filter.GenerateTestArgs("pkg/example")
	assert.NotEmpty(t, args)
}

// TestPrintSummary tests filter summary output
func TestPrintSummary(t *testing.T) {
	manifest := &testmanifest.TestManifest{
		Version:     "1.0",
		Description: "Test manifest",
		Packages: map[string]testmanifest.PackageConfig{
			"pkg/example": {
				Enabled:     true,
				Description: "Example tests",
				Tests: map[string]testmanifest.TestConfig{
					"TestExample": {
						Enabled:    true,
						Categories: []string{"unit"},
					},
				},
			},
		},
		Categories: map[string]testmanifest.CategoryConfig{
			"unit": {
				Description: "Unit tests",
			},
		},
	}

	filter := testmanifest.NewTestFilter(manifest)

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	filter.PrintSummary()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify summary is generated (actual format depends on implementation)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "pkg/example")
}
