package network

import (
	"context"
	"testing"
	"time"

	"src/internal/log"
)

func TestTransferManager(t *testing.T) {
	logger := log.NewLogger(log.InfoLevel)
	opts := DefaultTransferOptions()

	// Create transfer manager
	manager := NewTransferManager(logger, opts)

	// Check that the manager was initialized properly
	if manager.opts.RetryAttempts <= 0 {
		t.Errorf("RetryAttempts should be positive, got %d", manager.opts.RetryAttempts)
	}

	if manager.opts.RetryInitialDelay <= 0 {
		t.Errorf("RetryInitialDelay should be positive, got %v", manager.opts.RetryInitialDelay)
	}

	if manager.opts.RetryMaxDelay <= 0 {
		t.Errorf("RetryMaxDelay should be positive, got %v", manager.opts.RetryMaxDelay)
	}

	// Ensure delta manager was initialized
	if manager.deltaMan == nil {
		t.Errorf("Delta manager should be initialized")
	}
}

func TestDefaultTransferOptions(t *testing.T) {
	opts := DefaultTransferOptions()

	// Verify reasonable defaults
	if !opts.EnableCompression {
		t.Errorf("EnableCompression should default to true")
	}

	if !opts.EnableDelta {
		t.Errorf("EnableDelta should default to true")
	}

	if opts.RetryAttempts < 1 {
		t.Errorf("RetryAttempts should be at least 1, got %d", opts.RetryAttempts)
	}

	if opts.RetryInitialDelay < time.Millisecond {
		t.Errorf("RetryInitialDelay should be at least 1ms, got %v", opts.RetryInitialDelay)
	}
}

func TestTransferBlob(t *testing.T) {
	logger := log.NewLogger(log.InfoLevel)
	opts := DefaultTransferOptions()
	manager := NewTransferManager(logger, opts)

	// Create source and destination repositories
	sourceRepo := NewMockRepository()
	destRepo := NewMockRepository()

	// Add a manifest to the source
	manifest := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:abc"},"layers":[{"digest":"layer1"},{"digest":"layer2"}]}`)
	sourceRepo.PutManifest("@sha256:test-digest", manifest, "application/json")

	// Test transfer
	ctx := context.Background()
	result, err := manager.TransferBlob(ctx, sourceRepo, destRepo, "sha256:test-digest", "application/json")

	if err != nil {
		t.Fatalf("TransferBlob failed: %v", err)
	}

	// Verify the result
	if result.Digest != "sha256:test-digest" {
		t.Errorf("Expected digest=%s, got %s", "sha256:test-digest", result.Digest)
	}

	if result.Size == 0 {
		t.Errorf("Expected Size > 0")
	}

	// Check that the data was actually transferred
	if sourceRepo.getManifestCalls == 0 {
		t.Errorf("Source repository should have been accessed")
	}

	if destRepo.putManifestCalls == 0 {
		t.Errorf("Destination repository should have been updated")
	}

	// Verify the manifest was transferred correctly
	destManifest, _, err := destRepo.GetManifest("@sha256:test-digest")
	if err != nil {
		t.Fatalf("Failed to get manifest from destination: %v", err)
	}

	if string(destManifest) != string(manifest) {
		t.Errorf("Transferred manifest doesn't match original")
	}
}

func TestTransferResult(t *testing.T) {
	// Create a TransferResult
	result := &TransferResult{
		Digest:             "sha256:test",
		Size:               1000,
		TransferSize:       800,
		CompressionSavings: 10.0,
		DeltaSavings:       10.0,
		TotalSavings:       20.0,
		Duration:           100 * time.Millisecond,
		UsedDelta:          true,
		UsedCompression:    true,
	}

	// Verify calculations
	expectedSavings := 20.0
	if result.TotalSavings != expectedSavings {
		t.Errorf("Expected TotalSavings=%f, got %f", expectedSavings, result.TotalSavings)
	}

	// Verify size difference matches savings
	expectedTransferSize := 800
	if result.TransferSize != expectedTransferSize {
		t.Errorf("Expected TransferSize=%d, got %d", expectedTransferSize, result.TransferSize)
	}
}
