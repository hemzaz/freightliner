// Package testing provides test manifest functionality for selective test execution
package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// TestManifest represents the configuration for selective test execution
type TestManifest struct {
	Version     string                    `yaml:"version"`
	Description string                    `yaml:"description"`
	Global      GlobalConfig              `yaml:"global"`
	Packages    map[string]PackageConfig  `yaml:"packages"`
	Categories  map[string]CategoryConfig `yaml:"categories"`
	Environment EnvironmentDetection      `yaml:"environment_detection"`
	Reporting   ReportingConfig           `yaml:"reporting"`
	MakeTargets map[string]MakeTarget     `yaml:"make_targets"`
}

// GlobalConfig contains global test execution settings
type GlobalConfig struct {
	DefaultEnabled bool                         `yaml:"default_enabled"`
	Environments   map[string]EnvironmentConfig `yaml:"environments"`
}

// EnvironmentConfig defines behavior for different environments
type EnvironmentConfig struct {
	SkipExternalDeps bool `yaml:"skip_external_deps"`
	SkipFlakyTests   bool `yaml:"skip_flaky_tests"`
}

// PackageConfig defines test configuration for a package
type PackageConfig struct {
	Enabled     bool                  `yaml:"enabled"`
	Description string                `yaml:"description"`
	Tests       map[string]TestConfig `yaml:"tests"`
}

// TestConfig defines configuration for individual tests
type TestConfig struct {
	Enabled      bool     `yaml:"enabled"`
	Reason       string   `yaml:"reason"`
	Categories   []string `yaml:"categories"`
	SkipSubtests []string `yaml:"skip_subtests"`
}

// CategoryConfig defines test category settings
type CategoryConfig struct {
	Description string   `yaml:"description"`
	EnabledIn   []string `yaml:"enabled_in"`
	DisabledIn  []string `yaml:"disabled_in"`
}

// EnvironmentDetection defines how to detect different environments
type EnvironmentDetection struct {
	CIIndicators          []string `yaml:"ci_indicators"`
	IntegrationIndicators []string `yaml:"integration_indicators"`
	LocalIndicators       []string `yaml:"local_indicators"`
}

// ReportingConfig defines how test results should be reported
type ReportingConfig struct {
	ShowSkipped   bool   `yaml:"show_skipped"`
	ShowReasons   bool   `yaml:"show_reasons"`
	SummaryFormat string `yaml:"summary_format"`
}

// MakeTarget defines configuration for Make test targets
type MakeTarget struct {
	Description string   `yaml:"description"`
	Environment string   `yaml:"environment"`
	Categories  []string `yaml:"categories"`
}

// TestFilter represents a filter for test execution
type TestFilter struct {
	manifest    *TestManifest
	environment string
	categories  []string
}

// LoadTestManifest loads the test manifest from a YAML file
func LoadTestManifest(manifestPath string) (*TestManifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest TestManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}

	return &manifest, nil
}

// NewTestFilter creates a new test filter based on the manifest and current environment
func NewTestFilter(manifest *TestManifest) *TestFilter {
	environment := detectEnvironment(manifest.Environment)

	return &TestFilter{
		manifest:    manifest,
		environment: environment,
		categories:  []string{},
	}
}

// WithCategories sets specific categories to filter by
func (tf *TestFilter) WithCategories(categories []string) *TestFilter {
	tf.categories = categories
	return tf
}

// WithEnvironment overrides the detected environment
func (tf *TestFilter) WithEnvironment(env string) *TestFilter {
	tf.environment = env
	return tf
}

// ShouldRunTest determines if a specific test should be executed
func (tf *TestFilter) ShouldRunTest(packageName, testName string) (bool, string) {
	// Check if package exists in manifest
	pkgConfig, exists := tf.manifest.Packages[packageName]
	if !exists {
		// Package not in manifest, use global default
		return tf.manifest.Global.DefaultEnabled, "package not in manifest"
	}

	// Check if package is enabled
	if !pkgConfig.Enabled {
		return false, "package disabled"
	}

	// Check specific test configuration
	testConfig, exists := pkgConfig.Tests[testName]
	if !exists {
		// Test not specifically configured, use package default (enabled)
		return true, "test not specifically configured"
	}

	// Check if test is explicitly disabled
	if !testConfig.Enabled {
		return false, testConfig.Reason
	}

	// Check category-based filtering
	for _, category := range testConfig.Categories {
		if !tf.isCategoryEnabledInEnvironment(category) {
			return false, fmt.Sprintf("category '%s' disabled in environment '%s'", category, tf.environment)
		}
	}

	// Check if filtering by specific categories
	if len(tf.categories) > 0 {
		hasMatchingCategory := false
		for _, filterCategory := range tf.categories {
			for _, testCategory := range testConfig.Categories {
				if filterCategory == testCategory {
					hasMatchingCategory = true
					break
				}
			}
		}
		if !hasMatchingCategory {
			return false, "test doesn't match category filter"
		}
	}

	return true, "test enabled"
}

// ShouldRunSubtest determines if a specific subtest should be executed
func (tf *TestFilter) ShouldRunSubtest(packageName, testName, subtestName string) (bool, string) {
	// First check if main test should run
	shouldRun, reason := tf.ShouldRunTest(packageName, testName)
	if !shouldRun {
		return false, reason
	}

	// Check if subtest is specifically skipped
	pkgConfig := tf.manifest.Packages[packageName]
	testConfig := pkgConfig.Tests[testName]

	for _, skipSubtest := range testConfig.SkipSubtests {
		if skipSubtest == subtestName {
			return false, "subtest explicitly skipped"
		}
	}

	return true, "subtest enabled"
}

// GetPackageTests returns all tests for a package with their enabled status
func (tf *TestFilter) GetPackageTests(packageName string) map[string]bool {
	result := make(map[string]bool)

	pkgConfig, exists := tf.manifest.Packages[packageName]
	if !exists {
		return result
	}

	for testName := range pkgConfig.Tests {
		enabled, _ := tf.ShouldRunTest(packageName, testName)
		result[testName] = enabled
	}

	return result
}

// GenerateTestArgs generates Go test arguments based on the filter
func (tf *TestFilter) GenerateTestArgs(packageName string) []string {
	var args []string

	pkgConfig, exists := tf.manifest.Packages[packageName]
	if !exists {
		return args
	}

	var runPatterns []string
	var skipPatterns []string

	for testName, testConfig := range pkgConfig.Tests {
		enabled, _ := tf.ShouldRunTest(packageName, testName)

		if enabled {
			// Include this test
			runPatterns = append(runPatterns, fmt.Sprintf("^%s$", regexp.QuoteMeta(testName)))

			// But skip specific subtests if configured
			for _, skipSubtest := range testConfig.SkipSubtests {
				skipPatterns = append(skipPatterns, fmt.Sprintf("^%s/%s$",
					regexp.QuoteMeta(testName), regexp.QuoteMeta(skipSubtest)))
			}
		} else {
			// Skip this entire test
			skipPatterns = append(skipPatterns, fmt.Sprintf("^%s$", regexp.QuoteMeta(testName)))
		}
	}

	if len(runPatterns) > 0 {
		args = append(args, "-run", strings.Join(runPatterns, "|"))
	}

	if len(skipPatterns) > 0 {
		args = append(args, "-skip", strings.Join(skipPatterns, "|"))
	}

	return args
}

// PrintSummary prints a summary of test filtering decisions
func (tf *TestFilter) PrintSummary() {
	fmt.Printf("Test Manifest Summary (Environment: %s)\n", tf.environment)
	fmt.Println(strings.Repeat("=", 50))

	for packageName, pkgConfig := range tf.manifest.Packages {
		fmt.Printf("\nPackage: %s\n", packageName)
		fmt.Printf("  Description: %s\n", pkgConfig.Description)
		fmt.Printf("  Enabled: %v\n", pkgConfig.Enabled)

		if len(pkgConfig.Tests) > 0 {
			fmt.Println("  Tests:")
			for testName := range pkgConfig.Tests {
				enabled, reason := tf.ShouldRunTest(packageName, testName)
				status := "✓"
				if !enabled {
					status = "✗"
				}
				fmt.Printf("    %s %s", status, testName)
				if tf.manifest.Reporting.ShowReasons && reason != "" {
					fmt.Printf(" (%s)", reason)
				}
				fmt.Println()
			}
		}
	}

	// Print category summary
	fmt.Println("\nCategory Status:")
	for categoryName, categoryConfig := range tf.manifest.Categories {
		enabled := tf.isCategoryEnabledInEnvironment(categoryName)
		status := "✓"
		if !enabled {
			status = "✗"
		}
		fmt.Printf("  %s %s: %s\n", status, categoryName, categoryConfig.Description)
	}
}

// detectEnvironment determines the current environment based on environment variables
func detectEnvironment(envDetection EnvironmentDetection) string {
	// Check for CI environment
	for _, indicator := range envDetection.CIIndicators {
		if checkEnvIndicator(indicator) {
			return "ci"
		}
	}

	// Check for integration environment
	for _, indicator := range envDetection.IntegrationIndicators {
		if checkEnvIndicator(indicator) {
			return "integration"
		}
	}

	// Check for local environment
	for _, indicator := range envDetection.LocalIndicators {
		if checkEnvIndicator(indicator) {
			return "local"
		}
	}

	// Default to local
	return "local"
}

// checkEnvIndicator checks if an environment indicator is present
func checkEnvIndicator(indicator string) bool {
	if strings.Contains(indicator, "=") {
		parts := strings.SplitN(indicator, "=", 2)
		return os.Getenv(parts[0]) == parts[1]
	}
	return os.Getenv(indicator) != ""
}

// isCategoryEnabledInEnvironment checks if a category is enabled in the current environment
func (tf *TestFilter) isCategoryEnabledInEnvironment(category string) bool {
	categoryConfig, exists := tf.manifest.Categories[category]
	if !exists {
		return true // Unknown categories are enabled by default
	}

	// Check if explicitly disabled in this environment
	for _, disabledEnv := range categoryConfig.DisabledIn {
		if disabledEnv == tf.environment {
			return false
		}
	}

	// Check if explicitly enabled in this environment
	if len(categoryConfig.EnabledIn) > 0 {
		for _, enabledEnv := range categoryConfig.EnabledIn {
			if enabledEnv == tf.environment {
				return true
			}
		}
		return false // Not in enabled list
	}

	// Not explicitly disabled and no enabled list, so enabled by default
	return true
}

// GetDefaultManifestPath returns the default path for the test manifest
func GetDefaultManifestPath() string {
	// Try to find the manifest in common locations
	candidates := []string{
		"test-manifest.yaml",
		"test-manifest.yml",
		filepath.Join("test", "manifest.yaml"),
		filepath.Join("test", "manifest.yml"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return "test-manifest.yaml" // Default fallback
}
