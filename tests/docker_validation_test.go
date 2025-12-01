package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDockerfileDetection tests the Dockerfile detection logic
func TestDockerfileDetection(t *testing.T) {
	projectRoot := ".."

	t.Run("FindDockerfile", func(t *testing.T) {
		dockerfilePath := findDockerfile(projectRoot)

		if dockerfilePath == "" {
			t.Error("Expected to find a Dockerfile")
			return
		}

		t.Logf("Found Dockerfile: %s", dockerfilePath)

		// Verify file exists
		if _, err := os.Stat(dockerfilePath); err != nil {
			t.Errorf("Dockerfile path not accessible: %v", err)
		}

		// Verify it's a valid path
		if !filepath.IsAbs(dockerfilePath) && !strings.HasPrefix(dockerfilePath, ".") {
			t.Error("Dockerfile path should be absolute or relative")
		}
	})

	t.Run("DockerfileContent", func(t *testing.T) {
		dockerfilePath := findDockerfile(projectRoot)
		if dockerfilePath == "" {
			t.Skip("No Dockerfile found")
		}

		data, err := os.ReadFile(dockerfilePath)
		if err != nil {
			t.Fatalf("Failed to read Dockerfile: %v", err)
		}

		content := string(data)

		// Verify basic Dockerfile structure
		if !strings.Contains(content, "FROM") {
			t.Error("Dockerfile should contain FROM instruction")
		}

		// If it's Dockerfile.optimized, check for multi-stage
		if strings.Contains(dockerfilePath, "optimized") {
			stages := []string{"builder", "test", "build"}
			for _, stage := range stages {
				if !strings.Contains(content, "AS "+stage) && !strings.Contains(content, "as "+stage) {
					t.Errorf("Dockerfile.optimized should contain stage: %s", stage)
				}
			}
		}

		t.Log("✅ Dockerfile content validation passed")
	})
}

// findDockerfile mimics the helper function from pipeline_integration_test.go
func findDockerfile(projectRoot string) string {
	candidates := []string{
		filepath.Join(projectRoot, "Dockerfile.optimized"),
		filepath.Join(projectRoot, "Dockerfile"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			data, err := os.ReadFile(candidate)
			if err == nil {
				content := string(data)
				hasBuilder := strings.Contains(content, "AS builder") || strings.Contains(content, "as builder")
				hasTest := strings.Contains(content, "AS test") || strings.Contains(content, "as test")
				hasBuild := strings.Contains(content, "AS build") || strings.Contains(content, "as build")

				if strings.Contains(candidate, "optimized") && hasBuilder && hasTest && hasBuild {
					return candidate
				}
				if !strings.Contains(candidate, "optimized") {
					return candidate
				}
			}
		}
	}

	return ""
}
