// Package testing provides test execution helpers
package testing

import (
	"os"
	"testing"
)

// ShouldSkip determines if a test should be skipped based on the manifest
// This function should be called at the beginning of tests that might be disabled
func ShouldSkip(t *testing.T, packageName, testName string) bool {
	// Only check manifest if SKIP_TEST_MANIFEST is not set
	if os.Getenv("SKIP_TEST_MANIFEST") != "" {
		return false
	}

	// Try to load manifest
	manifestPath := GetDefaultManifestPath()
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// No manifest file, don't skip
		return false
	}

	manifest, err := LoadTestManifest(manifestPath)
	if err != nil {
		// Error loading manifest, don't skip (but log warning in verbose mode)
		if testing.Verbose() {
			t.Logf("Warning: Failed to load test manifest: %v", err)
		}
		return false
	}

	filter := NewTestFilter(manifest)
	shouldRun, reason := filter.ShouldRunTest(packageName, testName)

	if !shouldRun {
		t.Skipf("Test disabled by manifest: %s", reason)
		return true
	}

	return false
}

// ShouldSkipSubtest determines if a subtest should be skipped
func ShouldSkipSubtest(t *testing.T, packageName, testName, subtestName string) bool {
	// Only check manifest if SKIP_TEST_MANIFEST is not set
	if os.Getenv("SKIP_TEST_MANIFEST") != "" {
		return false
	}

	// Try to load manifest
	manifestPath := GetDefaultManifestPath()
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return false
	}

	manifest, err := LoadTestManifest(manifestPath)
	if err != nil {
		if testing.Verbose() {
			t.Logf("Warning: Failed to load test manifest: %v", err)
		}
		return false
	}

	filter := NewTestFilter(manifest)
	shouldRun, reason := filter.ShouldRunSubtest(packageName, testName, subtestName)

	if !shouldRun {
		t.Skipf("Subtest disabled by manifest: %s", reason)
		return true
	}

	return false
}

// SkipIfExternalDeps skips the test if external dependencies are not available
// This is a convenience function for tests that require cloud services
func SkipIfExternalDeps(t *testing.T, packageName, testName string) {
	// Check common environment variables that indicate external deps should be skipped
	if os.Getenv("SKIP_EXTERNAL_DEPS") != "" ||
		os.Getenv("CI") != "" ||
		os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping test requiring external dependencies")
		return
	}

	// Also check manifest
	if ShouldSkip(t, packageName, testName) {
		// Skip already handled by ShouldSkip
		return
	}
}

// SkipIfFlaky skips the test if it's marked as flaky and we're in an environment
// that should skip flaky tests
func SkipIfFlaky(t *testing.T, packageName, testName string) {
	// Check if we should skip flaky tests
	if os.Getenv("SKIP_FLAKY_TESTS") != "" ||
		os.Getenv("CI") != "" {

		// Check if this test is marked as flaky in the manifest
		manifestPath := GetDefaultManifestPath()
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			return
		}

		manifest, err := LoadTestManifest(manifestPath)
		if err != nil {
			return
		}

		pkgConfig, exists := manifest.Packages[packageName]
		if !exists {
			return
		}

		testConfig, exists := pkgConfig.Tests[testName]
		if !exists {
			return
		}

		// Check if test has "flaky" category
		for _, category := range testConfig.Categories {
			if category == "flaky" {
				t.Skip("Skipping flaky test in current environment")
				return
			}
		}
	}
}

// RequireExternalDeps fails the test if external dependencies are not available
// Use this for tests that absolutely require external services
func RequireExternalDeps(t *testing.T, service string) {
	if os.Getenv("SKIP_EXTERNAL_DEPS") != "" {
		t.Fatalf("Test requires external dependency '%s' but SKIP_EXTERNAL_DEPS is set", service)
	}

	// You can extend this to check for specific service availability
	switch service {
	case "aws":
		if os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_PROFILE") == "" {
			t.Skip("Skipping test requiring AWS credentials")
		}
	case "gcp":
		if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GCLOUD_PROJECT") == "" {
			t.Skip("Skipping test requiring GCP credentials")
		}
	}
}
