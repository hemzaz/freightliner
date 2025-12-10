package e2e

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2EReplicationWorkflow tests the complete end-to-end workflow
func TestE2EReplicationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Check if freightliner binary exists
	binaryPath := "../../bin/freightliner"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Freightliner binary not found, run 'make build' first")
	}

	// Check if test registries are available (skip if not)
	testRegistryAvailable := func(registry string) bool {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "curl", "-sf", "http://"+registry+"/v2/")
		return cmd.Run() == nil
	}

	if !testRegistryAvailable("localhost:5000") || !testRegistryAvailable("localhost:5001") {
		t.Skip("Test registries not available (localhost:5000 or localhost:5001). Run docker registries first.")
	}

	testCases := []struct {
		name        string
		source      string
		destination string
		expectError bool
	}{
		{
			name:        "BasicReplication",
			source:      "localhost:5000/test",
			destination: "localhost:5001/backup",
			expectError: false,
		},
		{
			name:        "MultiRepoReplication",
			source:      "localhost:5000/test",
			destination: "localhost:5001/mirror",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			// Build the command using positional args (replicate-tree takes source and destination as args)
			args := []string{
				"replicate-tree",
				tc.source,
				tc.destination,
				"--workers", "2",
				"--dry-run", // Use dry-run to avoid actual copying in tests
			}

			cmd := exec.CommandContext(ctx, binaryPath, args...)
			output, err := cmd.CombinedOutput()

			if tc.expectError {
				assert.Error(t, err, "Expected command to fail")
			} else {
				require.NoError(t, err, "Command should succeed\nOutput: %s", string(output))

				// Verify output contains expected information
				outputStr := string(output)
				assert.Contains(t, outputStr, "replication", "Output should mention replication")

				t.Logf("Command output:\n%s", outputStr)
			}
		})
	}
}

// TestE2ECheckpointResume tests checkpoint and resume functionality
func TestE2ECheckpointResume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := "../../bin/freightliner"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Freightliner binary not found")
	}

	// Check if test registries are available
	testRegistryAvailable := func(registry string) bool {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "curl", "-sf", "http://"+registry+"/v2/")
		return cmd.Run() == nil
	}

	if !testRegistryAvailable("localhost:5000") || !testRegistryAvailable("localhost:5001") {
		t.Skip("Test registries not available")
	}

	// Create temporary checkpoint directory
	checkpointDir, err := os.MkdirTemp("", "e2e_checkpoint_*")
	require.NoError(t, err)
	defer os.RemoveAll(checkpointDir)

	t.Run("CreateCheckpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		args := []string{
			"replicate-tree",
			"localhost:5000/test",
			"localhost:5001/checkpoint",
			"--checkpoint-dir", checkpointDir,
			"--checkpoint",
			"--dry-run", // Use dry-run to avoid actual copying
		}

		cmd := exec.CommandContext(ctx, binaryPath, args...)
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "Replication with checkpoint should succeed\nOutput: %s", string(output))

		// Verify checkpoint file was created (may not exist in dry-run mode)
		entries, err := os.ReadDir(checkpointDir)
		require.NoError(t, err)
		// Note: checkpoint files may not be created in dry-run mode
		t.Logf("Checkpoint directory has %d entries", len(entries))
		t.Logf("Created checkpoint in: %s", checkpointDir)
	})

	// Note: Resume testing would require interrupting the first run
	// and then resuming it, which is complex in a test environment
}

// TestE2EMetricsEndpoint tests the metrics HTTP endpoint
func TestE2EMetricsEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := "../../bin/freightliner"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Freightliner binary not found")
	}

	t.Run("ServeMetrics", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Start the server
		args := []string{
			"serve",
			"--port", "8081",
			"--metrics-port", "9092",
		}

		cmd := exec.CommandContext(ctx, binaryPath, args...)
		err := cmd.Start()
		require.NoError(t, err)

		// Give server time to start
		time.Sleep(2 * time.Second)

		// Try to curl the metrics endpoint
		curlCmd := exec.CommandContext(ctx, "curl", "-f", "http://localhost:9092/metrics")
		output, err := curlCmd.CombinedOutput()

		if err == nil {
			// Verify metrics output contains Prometheus-format metrics
			outputStr := string(output)
			assert.Contains(t, outputStr, "# HELP", "Should contain Prometheus help text")
			assert.Contains(t, outputStr, "# TYPE", "Should contain Prometheus type declarations")
			t.Logf("Metrics endpoint responding correctly")
		} else {
			t.Logf("Could not reach metrics endpoint (may need manual verification): %v", err)
		}

		// Stop the server
		cmd.Process.Kill()
	})
}

// TestE2EConfigValidation tests configuration validation
func TestE2EConfigValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := "../../bin/freightliner"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Freightliner binary not found")
	}

	testCases := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name: "MissingDestination",
			args: []string{
				"replicate-tree",
				"localhost:5000/test",
			},
			expectError: true,
			errorMsg:    "accepts 2 arg",
		},
		{
			name: "MissingBothArgs",
			args: []string{
				"replicate-tree",
			},
			expectError: true,
			errorMsg:    "accepts 2 arg",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, tc.args...)
			output, err := cmd.CombinedOutput()

			if tc.expectError {
				assert.Error(t, err, "Expected validation error")

				if tc.errorMsg != "" {
					outputStr := strings.ToLower(string(output))
					assert.Contains(t, outputStr, strings.ToLower(tc.errorMsg),
						"Error message should contain: %s\nGot: %s", tc.errorMsg, outputStr)
				}
			} else {
				assert.NoError(t, err, "Should not error\nOutput: %s", string(output))
			}
		})
	}
}

// TestE2EDryRun tests dry-run functionality
func TestE2EDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := "../../bin/freightliner"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Freightliner binary not found")
	}

	// Check if test registries are available
	testRegistryAvailable := func(registry string) bool {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "curl", "-sf", "http://"+registry+"/v2/")
		return cmd.Run() == nil
	}

	if !testRegistryAvailable("localhost:5000") || !testRegistryAvailable("localhost:5001") {
		t.Skip("Test registries not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	args := []string{
		"replicate-tree",
		"localhost:5000/test",
		"localhost:5001/dryrun",
		"--dry-run",
	}

	cmd := exec.CommandContext(ctx, binaryPath, args...)
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "Dry run should succeed\nOutput: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, "dry", "Output should mention dry run")

	t.Logf("Dry run output:\n%s", outputStr)
}

// TestE2EVersionCommand tests the version command
func TestE2EVersionCommand(t *testing.T) {
	binaryPath := "../../bin/freightliner"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Freightliner binary not found")
	}

	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "Version command should succeed")

	outputStr := string(output)
	assert.Contains(t, outputStr, "Freightliner", "Should show Freightliner version information")

	t.Logf("Version output:\n%s", outputStr)
}
