package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestGolangCILintConfig validates the golangci-lint configuration file
func TestGolangCILintConfig(t *testing.T) {
	configPath := filepath.Join("..", "..", "..", ".golangci.yml")

	// Test file exists and is readable
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read golangci-lint config: %v", err)
	}

	// Test YAML syntax validation
	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Invalid YAML syntax in .golangci.yml: %v", err)
	}

	// Test required sections exist
	requiredSections := []string{"run", "linters", "issues"}
	for _, section := range requiredSections {
		if _, exists := config[section]; !exists {
			t.Errorf("Missing required section in .golangci.yml: %s", section)
		}
	}

	// Test run configuration
	if run, ok := config["run"].(map[string]interface{}); ok {
		// Validate timeout setting
		if timeout, exists := run["timeout"]; exists {
			if timeoutStr, ok := timeout.(string); ok {
				if !strings.HasSuffix(timeoutStr, "m") && !strings.HasSuffix(timeoutStr, "s") {
					t.Errorf("Invalid timeout format: %v", timeout)
				}
			}
		}

		// Validate Go version if specified
		if goVersion, exists := run["go"]; exists {
			if goVersionStr, ok := goVersion.(string); ok {
				// Check if Go version format is valid (e.g., "1.23.4")
				parts := strings.Split(goVersionStr, ".")
				if len(parts) != 3 {
					t.Errorf("Invalid Go version format: %v", goVersion)
				}
			}
		}

		// Validate concurrency setting
		if concurrency, exists := run["concurrency"]; exists {
			if concurrencyVal, ok := concurrency.(int); ok {
				if concurrencyVal <= 0 || concurrencyVal > 16 {
					t.Errorf("Invalid concurrency value: %d (should be 1-16)", concurrencyVal)
				}
			}
		}
	}

	// Test linters configuration
	if linters, ok := config["linters"].(map[string]interface{}); ok {
		if enable, exists := linters["enable"]; exists {
			if enableList, ok := enable.([]interface{}); ok {
				// Check for essential linters
				essentialLinters := []string{"errcheck", "gosimple", "govet", "staticcheck"}
				enabledLinters := make(map[string]bool)

				for _, linter := range enableList {
					if linterStr, ok := linter.(string); ok {
						enabledLinters[linterStr] = true
					}
				}

				for _, essential := range essentialLinters {
					if !enabledLinters[essential] {
						t.Errorf("Essential linter not enabled: %s", essential)
					}
				}
			}
		}
	}

	// Test issues configuration
	if issues, ok := config["issues"].(map[string]interface{}); ok {
		// Validate max-issues-per-linter
		if maxIssues, exists := issues["max-issues-per-linter"]; exists {
			if maxIssuesVal, ok := maxIssues.(int); ok {
				if maxIssuesVal <= 0 {
					t.Errorf("Invalid max-issues-per-linter: %d", maxIssuesVal)
				}
			}
		}

		// Validate exclude-rules structure
		if excludeRules, exists := issues["exclude-rules"]; exists {
			if rules, ok := excludeRules.([]interface{}); ok {
				for i, rule := range rules {
					if ruleMap, ok := rule.(map[string]interface{}); ok {
						// Each rule should have either path or linters
						if _, hasPath := ruleMap["path"]; !hasPath {
							if _, hasLinters := ruleMap["linters"]; !hasLinters {
								t.Errorf("Exclude rule %d missing both 'path' and 'linters'", i)
							}
						}
					}
				}
			}
		}
	}

	t.Logf("✅ golangci-lint configuration validation passed")
}

// TestDockerfileValidation validates the Dockerfile structure and security practices
func TestDockerfileValidation(t *testing.T) {
	dockerfilePath := filepath.Join("..", "..", "..", "Dockerfile")

	data, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("Failed to read Dockerfile: %v", err)
	}

	content := string(data)

	// Test multi-stage build stages exist
	requiredStages := []string{"builder", "test", "build"}
	for _, stage := range requiredStages {
		// Check for stage definition patterns
		stagePatterns := []string{
			fmt.Sprintf("FROM golang:1.24.5-alpine AS %s", stage),
			fmt.Sprintf("FROM builder AS %s", stage),
		}

		found := false
		for _, pattern := range stagePatterns {
			if strings.Contains(content, pattern) {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Missing required build stage: %s", stage)
		}
	}

	// Test security best practices
	securityChecks := map[string]string{
		"Non-root user":     "USER 1001:1001",
		"Health check":      "HEALTHCHECK",
		"Alpine base image": "FROM alpine:",
		"Cache mounts":      "RUN --mount=type=cache",
	}

	for checkName, pattern := range securityChecks {
		if !strings.Contains(content, pattern) {
			t.Errorf("Dockerfile security check failed: %s (missing: %s)", checkName, pattern)
		}
	}

	// Test for security anti-patterns
	antiPatterns := map[string]string{
		"Root user in final stage": "USER root",
		"Unnecessary packages":     "apk add.*curl.*wget",
		"Secrets in build args":    "ARG.*PASSWORD|ARG.*SECRET|ARG.*TOKEN",
	}

	for checkName, pattern := range antiPatterns {
		// Use case-insensitive search for security patterns
		if strings.Contains(strings.ToUpper(content), strings.ToUpper(pattern)) {
			t.Errorf("Dockerfile security anti-pattern detected: %s", checkName)
		}
	}

	// Test Dockerfile syntax patterns
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Check for proper instruction casing
		if strings.HasPrefix(line, "from ") || strings.HasPrefix(line, "run ") ||
			strings.HasPrefix(line, "copy ") || strings.HasPrefix(line, "workdir ") {
			t.Errorf("Line %d: Dockerfile instructions should be uppercase: %s", i+1, line)
		}
	}

	t.Logf("✅ Dockerfile validation passed")
}

// TestGitHubActionsWorkflowValidation validates GitHub Actions workflow files
func TestGitHubActionsWorkflowValidation(t *testing.T) {
	workflowsDir := filepath.Join("..", "..", "..", ".github", "workflows")

	entries, err := os.ReadDir(workflowsDir)
	if err != nil {
		t.Fatalf("Failed to read workflows directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yml") {
			t.Run(entry.Name(), func(t *testing.T) {
				validateWorkflowFile(t, filepath.Join(workflowsDir, entry.Name()))
			})
		}
	}
}

func validateWorkflowFile(t *testing.T, workflowPath string) {
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read workflow file %s: %v", workflowPath, err)
	}

	var workflow map[string]interface{}
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatalf("Invalid YAML syntax in %s: %v", workflowPath, err)
	}

	// Test required fields
	requiredFields := []string{"name", "on", "jobs"}
	for _, field := range requiredFields {
		if _, exists := workflow[field]; !exists {
			t.Errorf("Missing required field in %s: %s", workflowPath, field)
		}
	}

	// Validate jobs structure
	if jobs, ok := workflow["jobs"].(map[string]interface{}); ok {
		for jobName, jobConfig := range jobs {
			if job, ok := jobConfig.(map[string]interface{}); ok {
				// Each job should have runs-on
				if _, hasRunsOn := job["runs-on"]; !hasRunsOn {
					t.Errorf("Job %s missing 'runs-on' in %s", jobName, workflowPath)
				}

				// Each job should have steps
				if _, hasSteps := job["steps"]; !hasSteps {
					t.Errorf("Job %s missing 'steps' in %s", jobName, workflowPath)
				}

				// Check for timeout configuration
				if steps, ok := job["steps"].([]interface{}); ok {
					hasTimeout := false
					if timeout, hasJobTimeout := job["timeout-minutes"]; hasJobTimeout {
						if timeoutVal, ok := timeout.(int); ok && timeoutVal > 0 {
							hasTimeout = true
						}
					}

					// If no job timeout, check for reasonable number of steps
					if !hasTimeout && len(steps) > 10 {
						t.Logf("Warning: Job %s has %d steps without timeout in %s", jobName, len(steps), workflowPath)
					}
				}
			}
		}
	}

	// Check for security best practices
	content := string(data)

	// Check for pinned action versions
	actionLines := strings.Split(content, "\n")
	for i, line := range actionLines {
		if strings.Contains(line, "uses:") && !strings.Contains(line, "#") {
			// Extract action reference
			parts := strings.Split(line, "uses:")
			if len(parts) > 1 {
				action := strings.TrimSpace(parts[1])
				// Check if action is pinned to a specific version (not @main or @master)
				if strings.HasSuffix(action, "@main") || strings.HasSuffix(action, "@master") {
					t.Logf("Warning: Unpinned action version on line %d in %s: %s", i+1, workflowPath, action)
				}
			}
		}
	}

	t.Logf("✅ Workflow validation passed for %s", filepath.Base(workflowPath))
}

// TestProjectStructure validates the overall project structure
func TestProjectStructure(t *testing.T) {
	projectRoot := filepath.Join("..", "..", "..")

	// Required directories and files
	requiredPaths := []string{
		".github/workflows",
		"pkg",
		"cmd",
		"go.mod",
		"go.sum",
		"Dockerfile",
		".golangci.yml",
		".gitignore",
	}

	for _, path := range requiredPaths {
		fullPath := filepath.Join(projectRoot, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Required path does not exist: %s", path)
		}
	}

	// Check for security files
	securityFiles := []string{
		".github/dependabot.yml",
		"scripts/security-validation.sh",
	}

	missingSecurityFiles := []string{}
	for _, file := range securityFiles {
		fullPath := filepath.Join(projectRoot, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			missingSecurityFiles = append(missingSecurityFiles, file)
		}
	}

	if len(missingSecurityFiles) > 0 {
		t.Logf("Warning: Missing recommended security files: %v", missingSecurityFiles)
	}

	// Validate Go module structure
	goModPath := filepath.Join(projectRoot, "go.mod")
	goModData, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	goModContent := string(goModData)
	if !strings.Contains(goModContent, "module freightliner") {
		t.Error("go.mod should contain 'module freightliner'")
	}

	// Check for Go version specification
	if !strings.Contains(goModContent, "go 1.") {
		t.Error("go.mod should specify Go version")
	}

	t.Logf("✅ Project structure validation passed")
}

// BenchmarkConfigValidation benchmarks the configuration validation performance
func BenchmarkConfigValidation(b *testing.B) {
	configPath := filepath.Join("..", "..", "..", ".golangci.yml")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := os.ReadFile(configPath)
		if err != nil {
			b.Fatalf("Failed to read config: %v", err)
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(data, &config); err != nil {
			b.Fatalf("Failed to parse YAML: %v", err)
		}
	}
}
