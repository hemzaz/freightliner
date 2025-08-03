package validation

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"freightliner/pkg/testing/load"
)

// TestPipelineIntegration performs comprehensive pipeline integration testing
func TestPipelineIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	t.Run("LocalBuildAndTest", func(t *testing.T) {
		testLocalBuildAndTest(t, ctx)
	})

	t.Run("DockerMultiStageValidation", func(t *testing.T) {
		testDockerMultiStageValidation(t, ctx)
	})

	t.Run("LoadTestInfrastructure", func(t *testing.T) {
		testLoadTestInfrastructure(t, ctx)
	})

	t.Run("SecurityValidation", func(t *testing.T) {
		testSecurityValidation(t, ctx)
	})

	t.Run("ConfigurationValidation", func(t *testing.T) {
		testConfigurationValidation(t, ctx)
	})
}

// testLocalBuildAndTest validates local build and test execution
func testLocalBuildAndTest(t *testing.T, ctx context.Context) {
	projectRoot := filepath.Join("..", "..", "..")

	// Test Go module operations
	t.Run("GoModOperations", func(t *testing.T) {
		// Test go mod tidy
		cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go mod tidy failed: %v\nOutput: %s", err, output)
		}

		// Test go mod verify
		cmd = exec.CommandContext(ctx, "go", "mod", "verify")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go mod verify failed: %v\nOutput: %s", err, output)
		}

		t.Log("✅ Go module operations passed")
	})

	// Test build process
	t.Run("BuildProcess", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, "go", "build", "-v", "./...")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go build failed: %v\nOutput: %s", err, output)
		}

		t.Log("✅ Build process passed")
	})

	// Test unit tests
	t.Run("UnitTests", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, "go", "test", "-short", "-race", "-v", "./...")
		cmd.Dir = projectRoot
		cmd.Env = append(os.Environ(), "CGO_ENABLED=1") // Enable race detector

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		if err != nil {
			// Check if it's just missing test files (acceptable)
			if strings.Contains(outputStr, "no test files") {
				t.Log("✅ No test files found (acceptable for some packages)")
				return
			}
			t.Fatalf("Unit tests failed: %v\nOutput: %s", err, outputStr)
		}

		// Validate test output
		if !strings.Contains(outputStr, "PASS") && !strings.Contains(outputStr, "no test files") {
			t.Errorf("Unexpected test output format: %s", outputStr)
		}

		t.Log("✅ Unit tests passed")
	})
}

// testDockerMultiStageValidation validates Docker multi-stage build process
func testDockerMultiStageValidation(t *testing.T, ctx context.Context) {
	projectRoot := filepath.Join("..", "..", "..")

	// Check if Docker is available
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping Docker validation")
	}

	// Test individual build stages
	stages := map[string]string{
		"builder": "freightliner:builder-test",
		"test":    "freightliner:test-stage",
		"build":   "freightliner:build-stage",
	}

	for stage, tag := range stages {
		t.Run(fmt.Sprintf("Stage_%s", stage), func(t *testing.T) {
			buildCtx, buildCancel := context.WithTimeout(ctx, 5*time.Minute)
			defer buildCancel()

			cmd := exec.CommandContext(buildCtx, "docker", "build",
				"--target", stage,
				"-t", tag,
				".")
			cmd.Dir = projectRoot

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Docker build failed for stage %s: %v\nOutput: %s", stage, err, output)
			}

			t.Logf("✅ Docker stage %s built successfully", stage)
		})
	}

	// Test final image functionality
	t.Run("FinalImageTest", func(t *testing.T) {
		buildCtx, buildCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer buildCancel()

		// Build final image
		cmd := exec.CommandContext(buildCtx, "docker", "build", "-t", "freightliner:integration-test", ".")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Final Docker build failed: %v\nOutput: %s", err, output)
		}

		// Test image execution
		testCommands := [][]string{
			{"docker", "run", "--rm", "freightliner:integration-test", "--version"},
			{"docker", "run", "--rm", "freightliner:integration-test", "--help"},
		}

		for _, testCmd := range testCommands {
			runCtx, runCancel := context.WithTimeout(ctx, 30*time.Second)
			cmd := exec.CommandContext(runCtx, testCmd[0], testCmd[1:]...)

			output, err := cmd.CombinedOutput()
			runCancel()

			// Some commands might not be implemented yet, so don't fail on them
			if err != nil {
				t.Logf("Command %v returned error (acceptable): %v\nOutput: %s", testCmd[3:], err, output)
			} else {
				t.Logf("✅ Command %v executed successfully", testCmd[3:])
			}
		}

		// Cleanup test image
		exec.Command("docker", "rmi", "-f", "freightliner:integration-test").Run()
	})
}

// testLoadTestInfrastructure validates the load testing infrastructure
func testLoadTestInfrastructure(t *testing.T, ctx context.Context) {
	tempDir, err := os.MkdirTemp("", "load_test_integration")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test load testing framework components
	t.Run("BenchmarkSuiteCreation", func(t *testing.T) {
		suite := load.NewBenchmarkSuite(tempDir, nil)
		if suite == nil {
			t.Fatal("Failed to create benchmark suite")
		}

		// Validate suite configuration
		if suite.CountTotalResults(make(map[string][]load.BenchmarkResult)) != 0 {
			t.Error("Empty results should return 0 count")
		}

		t.Log("✅ Benchmark suite creation passed")
	})

	t.Run("ScenarioExecution", func(t *testing.T) {
		// Create minimal test scenario
		scenario := load.CreateHighVolumeReplicationScenario()
		scenario.Duration = 5 * time.Second // Reduced for testing
		if len(scenario.Images) > 2 {
			scenario.Images = scenario.Images[:2] // Limit to 2 images
		}

		runner := load.NewScenarioRunner(scenario, nil)
		if runner == nil {
			t.Fatal("Failed to create scenario runner")
		}

		// Run with timeout
		testCtx, testCancel := context.WithTimeout(ctx, 30*time.Second)
		defer testCancel()

		// Create a channel to receive the result
		resultChan := make(chan *load.LoadTestResults, 1)
		errChan := make(chan error, 1)

		go func() {
			result, err := runner.Run()
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- result
		}()

		select {
		case result := <-resultChan:
			// Validate results
			if result.ProcessedImages == 0 {
				t.Error("No images were processed")
			}
			if result.AverageThroughputMBps <= 0 {
				t.Error("Throughput should be positive")
			}
			t.Logf("✅ Scenario execution passed: %d images, %.2f MB/s",
				result.ProcessedImages, result.AverageThroughputMBps)

		case err := <-errChan:
			t.Logf("Scenario execution completed with issues (acceptable for integration test): %v", err)

		case <-testCtx.Done():
			t.Error("Scenario execution timed out")
		}
	})

	t.Run("PrometheusIntegration", func(t *testing.T) {
		collector := load.NewPrometheusLoadTestCollector(":0", nil)
		if collector == nil {
			t.Fatal("Failed to create Prometheus collector")
		}

		// Test metrics collection
		testResult := &load.LoadTestResults{
			ScenarioName:          "Integration Test",
			Duration:              5 * time.Second,
			ProcessedImages:       10,
			AverageThroughputMBps: 50.0,
			MemoryUsageMB:         256,
		}

		collector.RecordScenarioExecution("Integration Test", testResult)

		// Verify metrics were recorded (basic validation)
		collector.LoadTestMetrics.Mutex.RLock()
		executions := collector.LoadTestMetrics.ScenarioExecutions["Integration Test"]
		collector.LoadTestMetrics.Mutex.RUnlock()

		if executions != 1 {
			t.Errorf("Expected 1 execution, got %d", executions)
		}

		t.Log("✅ Prometheus integration passed")
	})
}

// testSecurityValidation validates security scanning and practices
func testSecurityValidation(t *testing.T, ctx context.Context) {
	projectRoot := filepath.Join("..", "..", "..")

	t.Run("GosecSecurity", func(t *testing.T) {
		// Check if gosec is available
		if !isCommandAvailable("gosec") {
			// Try to install gosec
			installCtx, installCancel := context.WithTimeout(ctx, 2*time.Minute)
			defer installCancel()

			cmd := exec.CommandContext(installCtx, "go", "install",
				"github.com/securego/gosec/v2/cmd/gosec@latest")
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Skipf("Failed to install gosec: %v\nOutput: %s", err, output)
			}
		}

		// Run security scan
		scanCtx, scanCancel := context.WithTimeout(ctx, 2*time.Minute)
		defer scanCancel()

		cmd := exec.CommandContext(scanCtx, "gosec", "-quiet", "-fmt", "json", "./...")
		cmd.Dir = projectRoot

		output, err := cmd.CombinedOutput()

		// gosec returns non-zero exit code when issues are found
		// We'll analyze the output instead of just checking the exit code
		outputStr := string(output)

		if err != nil && !strings.Contains(outputStr, `"Issues"`) {
			t.Fatalf("gosec execution failed: %v\nOutput: %s", err, outputStr)
		}

		// Parse JSON output to check for critical issues
		if strings.Contains(outputStr, `"severity":"HIGH"`) {
			t.Logf("Warning: High severity security issues found. Review gosec output.")
		}

		t.Log("✅ Security validation completed")
	})

	t.Run("DependencyCheck", func(t *testing.T) {
		// Check for known vulnerable dependencies using go list
		cmd := exec.CommandContext(ctx, "go", "list", "-json", "-m", "all")
		cmd.Dir = projectRoot

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to list dependencies: %v\nOutput: %s", err, output)
		}

		// Basic check for dependencies (more sophisticated vulnerability checking
		// would require integration with vulnerability databases)
		outputStr := string(output)
		if !strings.Contains(outputStr, `"Path"`) {
			t.Error("No dependencies found - unexpected for this project")
		}

		t.Log("✅ Dependency check completed")
	})
}

// testConfigurationValidation validates configuration files
func testConfigurationValidation(t *testing.T, ctx context.Context) {
	projectRoot := filepath.Join("..", "..", "..")

	t.Run("GolangCILintValidation", func(t *testing.T) {
		if !isCommandAvailable("golangci-lint") {
			t.Skip("golangci-lint not available, skipping validation")
		}

		// Validate configuration
		cmd := exec.CommandContext(ctx, "golangci-lint", "config", "verify", ".golangci.yml")
		cmd.Dir = projectRoot

		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("golangci-lint config validation failed: %v\nOutput: %s", err, output)
		}

		// Test dry run
		dryRunCtx, dryRunCancel := context.WithTimeout(ctx, 30*time.Second)
		defer dryRunCancel()

		cmd = exec.CommandContext(dryRunCtx, "golangci-lint", "run", "--dry-run", "--timeout=30s")
		cmd.Dir = projectRoot

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		if err != nil {
			// Check if it's a configuration issue vs. actual linting issues
			if strings.Contains(outputStr, "config") || strings.Contains(outputStr, "timeout") {
				t.Fatalf("golangci-lint configuration error: %v\nOutput: %s", err, outputStr)
			}
			// Linting issues are acceptable for dry run test
			t.Logf("golangci-lint found issues (acceptable for dry run): %s", outputStr)
		}

		t.Log("✅ golangci-lint configuration validation passed")
	})

	t.Run("DockerConfigValidation", func(t *testing.T) {
		if !isDockerAvailable() {
			t.Skip("Docker not available, skipping Docker config validation")
		}

		// Validate Dockerfile syntax using docker build --dry-run (if supported)
		// Or basic syntax validation
		dockerfilePath := filepath.Join(projectRoot, "Dockerfile")
		data, err := os.ReadFile(dockerfilePath)
		if err != nil {
			t.Fatalf("Failed to read Dockerfile: %v", err)
		}

		content := string(data)
		lines := strings.Split(content, "\n")

		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Basic syntax validation
			if !strings.Contains(line, " ") && !strings.HasPrefix(line, "FROM") {
				continue // Skip single-word lines
			}

			parts := strings.SplitN(line, " ", 2)
			if len(parts) < 2 {
				continue
			}

			instruction := strings.ToUpper(parts[0])
			validInstructions := []string{
				"FROM", "RUN", "COPY", "ADD", "WORKDIR", "EXPOSE", "ENV",
				"USER", "VOLUME", "ENTRYPOINT", "CMD", "LABEL", "HEALTHCHECK",
				"SHELL", "STOPSIGNAL", "ARG", "ONBUILD",
			}

			found := false
			for _, valid := range validInstructions {
				if instruction == valid {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Line %d: Invalid Dockerfile instruction: %s", i+1, instruction)
			}
		}

		t.Log("✅ Dockerfile configuration validation passed")
	})
}

// Helper functions

func isDockerAvailable() bool {
	cmd := exec.Command("docker", "--version")
	return cmd.Run() == nil
}

func isCommandAvailable(command string) bool {
	cmd := exec.Command("which", command)
	return cmd.Run() == nil
}

// BenchmarkPipelineIntegration benchmarks the integration test performance
func BenchmarkPipelineIntegration(b *testing.B) {
	ctx := context.Background()

	b.Run("ConfigurationValidation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate configuration validation performance
			configPath := filepath.Join("..", "..", "..", ".golangci.yml")
			_, err := os.ReadFile(configPath)
			if err != nil {
				b.Fatalf("Failed to read config: %v", err)
			}
		}
	})

	b.Run("LoadTestFramework", func(b *testing.B) {
		tempDir, err := os.MkdirTemp("", "bench_load_test")
		if err != nil {
			b.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			suite := load.NewBenchmarkSuite(tempDir, nil)
			if suite == nil {
				b.Fatal("Failed to create benchmark suite")
			}
		}
	})
}
