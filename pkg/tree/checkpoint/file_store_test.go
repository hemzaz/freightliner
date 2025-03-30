package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileStore(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "checkpoint-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the file store
	store, err := NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	// Create a test checkpoint
	checkpoint := &TreeCheckpoint{
		ID:                    "test-id",
		StartTime:             time.Now().Add(-time.Hour),
		LastUpdated:           time.Now(),
		SourceRegistry:        "ecr",
		SourcePrefix:          "prod",
		DestRegistry:          "gcr",
		DestPrefix:            "mirror",
		Status:                StatusInProgress,
		Progress:              50.0,
		Repositories:          make(map[string]RepoStatus),
		CompletedRepositories: []string{"repo1"},
		RepoTasks: []RepoTask{
			{
				SourceRegistry:   "ecr",
				SourceRepository: "repo1",
				DestRegistry:     "gcr",
				DestRepository:   "repo1",
				Status:           StatusCompleted,
				LastUpdated:      time.Now(),
			},
			{
				SourceRegistry:   "ecr",
				SourceRepository: "repo2",
				DestRegistry:     "gcr",
				DestRepository:   "repo2",
				Status:           StatusInProgress,
				LastUpdated:      time.Now(),
			},
			{
				SourceRegistry:   "ecr",
				SourceRepository: "repo3",
				DestRegistry:     "gcr",
				DestRepository:   "repo3",
				Status:           StatusPending,
				LastUpdated:      time.Now(),
			},
		},
	}

	// Test SaveCheckpoint
	err = store.SaveCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// Verify the file was created
	filePath := filepath.Join(tempDir, "test-id.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Checkpoint file was not created at %s", filePath)
	}

	// Test LoadCheckpoint
	loaded, err := store.LoadCheckpoint("test-id")
	if err != nil {
		t.Fatalf("LoadCheckpoint failed: %v", err)
	}

	// Verify checkpoint contents
	if loaded.ID != checkpoint.ID {
		t.Errorf("Expected ID=%s, got %s", checkpoint.ID, loaded.ID)
	}

	if loaded.SourceRegistry != checkpoint.SourceRegistry {
		t.Errorf("Expected SourceRegistry=%s, got %s", checkpoint.SourceRegistry, loaded.SourceRegistry)
	}

	if loaded.Progress != checkpoint.Progress {
		t.Errorf("Expected Progress=%f, got %f", checkpoint.Progress, loaded.Progress)
	}

	if len(loaded.Repositories) != len(checkpoint.Repositories) {
		t.Errorf("Expected %d repositories, got %d", len(checkpoint.Repositories), len(loaded.Repositories))
	}

	if len(loaded.RepoTasks) != len(checkpoint.RepoTasks) {
		t.Errorf("Expected %d repo tasks, got %d", len(checkpoint.RepoTasks), len(loaded.RepoTasks))
	}

	// Test ListCheckpoints
	checkpoints, err := store.ListCheckpoints()
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}

	if len(checkpoints) != 1 {
		t.Errorf("Expected 1 checkpoint, got %d", len(checkpoints))
	}

	// Test DeleteCheckpoint
	err = store.DeleteCheckpoint("test-id")
	if err != nil {
		t.Fatalf("DeleteCheckpoint failed: %v", err)
	}

	// Verify the file was deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("Checkpoint file was not deleted at %s", filePath)
	}

	// Test loading a non-existent checkpoint
	_, err = store.LoadCheckpoint("non-existent")
	if err == nil {
		t.Errorf("Expected error when loading non-existent checkpoint")
	}
}

func TestFileStoreConcurrency(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "checkpoint-test-concurrency")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the file store
	store, err := NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	// Create a test checkpoint
	checkpoint := &TreeCheckpoint{
		ID:             "test-concurrent",
		StartTime:      time.Now(),
		LastUpdated:    time.Now(),
		SourceRegistry: "ecr",
		SourcePrefix:   "prod",
		DestRegistry:   "gcr",
		DestPrefix:     "mirror",
		Status:         StatusInProgress,
		Progress:       0.0,
		Repositories:   make(map[string]RepoStatus),
	}

	// Save the initial checkpoint
	err = store.SaveCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// Simulate concurrent updates
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			// Update the checkpoint
			cp, err := store.LoadCheckpoint("test-concurrent")
			if err != nil {
				t.Errorf("LoadCheckpoint failed in goroutine %d: %v", idx, err)
				done <- true
				return
			}

			// Modify the checkpoint
			cp.Progress = float64((idx + 1) * 20)
			cp.LastUpdated = time.Now()

			// Save the updated checkpoint
			err = store.SaveCheckpoint(cp)
			if err != nil {
				t.Errorf("SaveCheckpoint failed in goroutine %d: %v", idx, err)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 5; i++ {
		<-done
	}

	// Load the final checkpoint
	final, err := store.LoadCheckpoint("test-concurrent")
	if err != nil {
		t.Fatalf("LoadCheckpoint failed: %v", err)
	}

	// The progress should be set by one of the goroutines
	if final.Progress == 0.0 {
		t.Errorf("Expected progress to be updated, still 0.0")
	}
}
