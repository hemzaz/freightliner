package checkpoint

import (
	"testing"
	"time"
)

func TestResumableCheckpoints(t *testing.T) {
	// Create a temporary store
	tempDir := t.TempDir()
	store, err := NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	// Create checkpoints with different statuses
	checkpoints := []*TreeCheckpoint{
		{
			ID:                    "completed",
			Status:                StatusCompleted,
			StartTime:             time.Now().Add(-time.Hour),
			LastUpdated:           time.Now().Add(-30 * time.Minute),
			SourceRegistry:        "ecr",
			SourcePrefix:          "prod",
			Progress:              100.0,
			RepoTasks:             []RepoTask{},
			CompletedRepositories: []string{"repo1", "repo2"},
			Repositories:          map[string]RepoStatus{},
		},
		{
			ID:                    "interrupted",
			Status:                StatusInterrupted,
			StartTime:             time.Now().Add(-time.Hour),
			LastUpdated:           time.Now().Add(-30 * time.Minute),
			SourceRegistry:        "ecr",
			SourcePrefix:          "staging",
			Progress:              50.0,
			CompletedRepositories: []string{"repo1"},
			Repositories:          map[string]RepoStatus{},
			RepoTasks: []RepoTask{
				{
					SourceRepository: "repo1",
					Status:           StatusCompleted,
				},
				{
					SourceRepository: "repo2",
					Status:           StatusInterrupted,
				},
			},
		},
		{
			ID:                    "failed",
			Status:                StatusFailed,
			StartTime:             time.Now().Add(-time.Hour),
			LastUpdated:           time.Now().Add(-30 * time.Minute),
			SourceRegistry:        "ecr",
			SourcePrefix:          "dev",
			Progress:              33.3,
			CompletedRepositories: []string{"repo1"},
			Repositories:          map[string]RepoStatus{},
			LastError:             "Failed to replicate repository repo2",
			RepoTasks: []RepoTask{
				{
					SourceRepository: "repo1",
					Status:           StatusCompleted,
				},
				{
					SourceRepository: "repo2",
					Status:           StatusFailed,
					Error:            "Error copying manifest",
				},
				{
					SourceRepository: "repo3",
					Status:           StatusPending,
				},
			},
		},
		{
			ID:                    "in-progress",
			Status:                StatusInProgress,
			StartTime:             time.Now().Add(-time.Hour),
			LastUpdated:           time.Now().Add(-5 * time.Minute),
			SourceRegistry:        "gcr",
			SourcePrefix:          "prod",
			Progress:              75.0,
			CompletedRepositories: []string{"repo1", "repo2", "repo3"},
			Repositories:          map[string]RepoStatus{},
			RepoTasks: []RepoTask{
				{
					SourceRepository: "repo1",
					Status:           StatusCompleted,
				},
				{
					SourceRepository: "repo2",
					Status:           StatusCompleted,
				},
				{
					SourceRepository: "repo3",
					Status:           StatusCompleted,
				},
				{
					SourceRepository: "repo4",
					Status:           StatusInProgress,
				},
			},
		},
	}

	// Save all checkpoints
	for _, cp := range checkpoints {
		saveErr := store.SaveCheckpoint(cp)
		if saveErr != nil {
			t.Fatalf("Failed to save checkpoint %s: %v", cp.ID, saveErr)
		}
	}

	// Get resumable checkpoints
	resumable, err := GetResumableCheckpoints(store)
	if err != nil {
		t.Fatalf("GetResumableCheckpoints failed: %v", err)
	}

	// Should get 3 resumable checkpoints (interrupted, failed, in-progress)
	if len(resumable) != 3 {
		t.Errorf("Expected 3 resumable checkpoints, got %d", len(resumable))
	}

	// Check that completed checkpoints are not included
	for _, r := range resumable {
		if r.ID == "completed" {
			t.Errorf("Completed checkpoint should not be resumable")
		}
	}

	// Test getting a specific checkpoint
	cp, err := GetCheckpointByID(store, "interrupted")
	if err != nil {
		t.Fatalf("GetCheckpointByID failed: %v", err)
	}

	if cp.ID != "interrupted" {
		t.Errorf("Expected checkpoint ID 'interrupted', got '%s'", cp.ID)
	}

	if cp.Status != StatusInterrupted {
		t.Errorf("Expected status %s, got %s", StatusInterrupted, cp.Status)
	}

	// Test getting remaining repositories with skip completed
	remaining, err := GetRemainingRepositories(cp, ResumableOptions{
		ID:            "interrupted",
		SkipCompleted: true,
		RetryFailed:   true,
	})
	if err != nil {
		t.Fatalf("GetRemainingRepositories failed: %v", err)
	}

	if len(remaining) != 1 {
		t.Errorf("Expected 1 remaining repository with SkipCompleted=true, got %d", len(remaining))
	}

	// Test getting remaining repositories without skip completed
	remaining, err = GetRemainingRepositories(cp, ResumableOptions{
		ID:            "interrupted",
		SkipCompleted: false,
		RetryFailed:   true,
	})
	if err != nil {
		t.Fatalf("GetRemainingRepositories failed: %v", err)
	}

	if len(remaining) != 2 {
		t.Errorf("Expected 2 remaining repositories with SkipCompleted=false, got %d", len(remaining))
	}
}

func TestResumableCheckpointStructs(t *testing.T) {
	// Test ResumableCheckpoint fields
	rc := ResumableCheckpoint{
		ID:                    "test",
		SourceRegistry:        "ecr",
		SourcePrefix:          "prod",
		DestRegistry:          "gcr",
		DestPrefix:            "mirror",
		Status:                StatusInterrupted,
		Progress:              50.0,
		LastUpdated:           time.Now(),
		TotalRepositories:     10,
		CompletedRepositories: 5,
		FailedRepositories:    1,
		Duration:              30 * time.Minute,
	}

	if rc.Progress != 50.0 {
		t.Errorf("Expected Progress=50.0, got %f", rc.Progress)
	}

	if rc.CompletedRepositories != 5 {
		t.Errorf("Expected CompletedRepositories=5, got %d", rc.CompletedRepositories)
	}

	// Test ResumableOptions fields
	ro := ResumableOptions{
		ID:            "test",
		SkipCompleted: true,
		RetryFailed:   false,
		Force:         true,
	}

	if !ro.SkipCompleted {
		t.Errorf("Expected SkipCompleted=true")
	}

	if ro.RetryFailed {
		t.Errorf("Expected RetryFailed=false")
	}
}
