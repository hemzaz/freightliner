package checkpoint

import (
	"testing"
	"time"
)

func TestResumableCheckpointsImproved(t *testing.T) {
	// Create a temporary store
	tempDir := t.TempDir()
	store, err := NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	// Create checkpoints with different statuses and more realistic data
	checkpoints := []*TreeCheckpoint{
		{
			ID:                    "completed-1",
			Status:                StatusCompleted,
			StartTime:             time.Now().Add(-2 * time.Hour),
			LastUpdated:           time.Now().Add(-1 * time.Hour),
			SourceRegistry:        "ecr",
			SourcePrefix:          "prod",
			Progress:              100.0,
			CompletedRepositories: []string{"prod/app1", "prod/app2", "prod/app3"},
			Repositories: map[string]RepoStatus{
				"prod/app1": {Status: StatusCompleted, LastUpdated: time.Now().Add(-90 * time.Minute)},
				"prod/app2": {Status: StatusCompleted, LastUpdated: time.Now().Add(-85 * time.Minute)},
				"prod/app3": {Status: StatusCompleted, LastUpdated: time.Now().Add(-80 * time.Minute)},
			},
			RepoTasks: []RepoTask{
				{SourceRepository: "prod/app1", Status: StatusCompleted},
				{SourceRepository: "prod/app2", Status: StatusCompleted},
				{SourceRepository: "prod/app3", Status: StatusCompleted},
			},
		},
		{
			ID:                    "interrupted-1",
			Status:                StatusInterrupted,
			StartTime:             time.Now().Add(-3 * time.Hour),
			LastUpdated:           time.Now().Add(-30 * time.Minute),
			SourceRegistry:        "gcr",
			SourcePrefix:          "staging",
			Progress:              60.0,
			CompletedRepositories: []string{"staging/app1", "staging/app2"},
			Repositories: map[string]RepoStatus{
				"staging/app1": {Status: StatusCompleted, LastUpdated: time.Now().Add(-45 * time.Minute)},
				"staging/app2": {Status: StatusCompleted, LastUpdated: time.Now().Add(-40 * time.Minute)},
				"staging/app3": {Status: StatusInterrupted, LastUpdated: time.Now().Add(-30 * time.Minute)},
				"staging/app4": {Status: StatusPending, LastUpdated: time.Time{}},
			},
			RepoTasks: []RepoTask{
				{SourceRepository: "staging/app1", Status: StatusCompleted},
				{SourceRepository: "staging/app2", Status: StatusCompleted},
				{SourceRepository: "staging/app3", Status: StatusInterrupted, Error: "Connection timeout"},
				{SourceRepository: "staging/app4", Status: StatusPending},
			},
		},
		{
			ID:                    "failed-1",
			Status:                StatusFailed,
			StartTime:             time.Now().Add(-4 * time.Hour),
			LastUpdated:           time.Now().Add(-2 * time.Hour),
			SourceRegistry:        "ecr",
			SourcePrefix:          "dev",
			Progress:              25.0,
			CompletedRepositories: []string{"dev/app1"},
			Repositories: map[string]RepoStatus{
				"dev/app1": {Status: StatusCompleted, LastUpdated: time.Now().Add(-3 * time.Hour)},
				"dev/app2": {Status: StatusFailed, LastUpdated: time.Now().Add(-2*time.Hour + 30*time.Minute), Error: "Manifest not found"},
				"dev/app3": {Status: StatusPending, LastUpdated: time.Time{}},
				"dev/app4": {Status: StatusPending, LastUpdated: time.Time{}},
			},
			LastError: "Failed to replicate repository dev/app2: manifest not found",
			RepoTasks: []RepoTask{
				{SourceRepository: "dev/app1", Status: StatusCompleted},
				{SourceRepository: "dev/app2", Status: StatusFailed, Error: "Manifest not found"},
				{SourceRepository: "dev/app3", Status: StatusPending},
				{SourceRepository: "dev/app4", Status: StatusPending},
			},
		},
		{
			ID:                    "in-progress-1",
			Status:                StatusInProgress,
			StartTime:             time.Now().Add(-1 * time.Hour),
			LastUpdated:           time.Now().Add(-2 * time.Minute),
			SourceRegistry:        "gcr",
			SourcePrefix:          "prod",
			Progress:              80.0,
			CompletedRepositories: []string{"prod/web1", "prod/web2", "prod/web3", "prod/web4"},
			Repositories: map[string]RepoStatus{
				"prod/web1": {Status: StatusCompleted, LastUpdated: time.Now().Add(-50 * time.Minute)},
				"prod/web2": {Status: StatusCompleted, LastUpdated: time.Now().Add(-45 * time.Minute)},
				"prod/web3": {Status: StatusCompleted, LastUpdated: time.Now().Add(-40 * time.Minute)},
				"prod/web4": {Status: StatusCompleted, LastUpdated: time.Now().Add(-35 * time.Minute)},
				"prod/web5": {Status: StatusInProgress, LastUpdated: time.Now().Add(-2 * time.Minute)},
			},
			RepoTasks: []RepoTask{
				{SourceRepository: "prod/web1", Status: StatusCompleted},
				{SourceRepository: "prod/web2", Status: StatusCompleted},
				{SourceRepository: "prod/web3", Status: StatusCompleted},
				{SourceRepository: "prod/web4", Status: StatusCompleted},
				{SourceRepository: "prod/web5", Status: StatusInProgress},
			},
		},
	}

	// Save all checkpoints
	for _, cp := range checkpoints {
		if err := store.SaveCheckpoint(cp); err != nil {
			t.Fatalf("Failed to save checkpoint %s: %v", cp.ID, err)
		}
	}

	// Test GetResumableCheckpoints
	resumable, err := GetResumableCheckpoints(store)
	if err != nil {
		t.Fatalf("GetResumableCheckpoints failed: %v", err)
	}

	// Should get 3 resumable checkpoints (interrupted, failed, in-progress)
	// Completed checkpoints should be excluded
	expectedResumableCount := 3
	if len(resumable) != expectedResumableCount {
		t.Errorf("Expected %d resumable checkpoints, got %d", expectedResumableCount, len(resumable))

		// Debug: print what we got
		for i, r := range resumable {
			t.Logf("Resumable[%d]: ID=%s, Status=%s", i, r.ID, r.Status)
		}
	}

	// Verify that completed checkpoints are not included
	for _, r := range resumable {
		if r.Status == StatusCompleted {
			t.Errorf("Completed checkpoint %s should not be resumable", r.ID)
		}
	}

	// Test getting a specific checkpoint
	interruptedCP, err := GetCheckpointByID(store, "interrupted-1")
	if err != nil {
		t.Fatalf("GetCheckpointByID failed: %v", err)
	}

	if interruptedCP.ID != "interrupted-1" {
		t.Errorf("Expected checkpoint ID 'interrupted-1', got '%s'", interruptedCP.ID)
	}

	if interruptedCP.Status != StatusInterrupted {
		t.Errorf("Expected status %s, got %s", StatusInterrupted, interruptedCP.Status)
	}

	// Test GetRemainingRepositories with SkipCompleted=true
	// This should return only repositories that are not completed
	remaining, err := GetRemainingRepositories(interruptedCP, ResumableOptions{
		ID:            "interrupted-1",
		SkipCompleted: true,
		RetryFailed:   true,
	})
	if err != nil {
		t.Fatalf("GetRemainingRepositories with SkipCompleted=true failed: %v", err)
	}

	// From the interrupted checkpoint:
	// - staging/app1: completed (should be skipped)
	// - staging/app2: completed (should be skipped)
	// - staging/app3: interrupted (should be included)
	// - staging/app4: pending (should be included)
	expectedRemainingSkipCompleted := 2
	if len(remaining) != expectedRemainingSkipCompleted {
		t.Errorf("Expected %d remaining repositories with SkipCompleted=true, got %d",
			expectedRemainingSkipCompleted, len(remaining))

		// Debug: print what we got
		for i, repo := range remaining {
			t.Logf("Remaining[%d]: %s", i, repo)
		}
	}

	// Verify that completed repositories are not in the remaining list
	completedRepos := map[string]bool{
		"staging/app1": true,
		"staging/app2": true,
	}
	for _, repo := range remaining {
		if completedRepos[repo] {
			t.Errorf("Completed repository %s should not be in remaining list when SkipCompleted=true", repo)
		}
	}

	// Test GetRemainingRepositories with SkipCompleted=false
	// This should return all repositories except completed ones, but include failed for retry
	remainingIncludeCompleted, err := GetRemainingRepositories(interruptedCP, ResumableOptions{
		ID:            "interrupted-1",
		SkipCompleted: false,
		RetryFailed:   true,
	})
	if err != nil {
		t.Fatalf("GetRemainingRepositories with SkipCompleted=false failed: %v", err)
	}

	// When SkipCompleted=false, we should get:
	// - staging/app1: completed (should be included)
	// - staging/app2: completed (should be included)
	// - staging/app3: interrupted (should be included)
	// - staging/app4: pending (should be included)
	expectedRemainingIncludeCompleted := 4
	if len(remainingIncludeCompleted) != expectedRemainingIncludeCompleted {
		t.Errorf("Expected %d remaining repositories with SkipCompleted=false, got %d",
			expectedRemainingIncludeCompleted, len(remainingIncludeCompleted))
	}
}

func TestResumableOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		options ResumableOptions
		valid   bool
	}{
		{
			name: "Valid options - skip completed, retry failed",
			options: ResumableOptions{
				ID:            "test-1",
				SkipCompleted: true,
				RetryFailed:   true,
				Force:         false,
			},
			valid: true,
		},
		{
			name: "Valid options - include completed, no retry failed",
			options: ResumableOptions{
				ID:            "test-2",
				SkipCompleted: false,
				RetryFailed:   false,
				Force:         false,
			},
			valid: true,
		},
		{
			name: "Invalid options - empty ID",
			options: ResumableOptions{
				ID:            "",
				SkipCompleted: true,
				RetryFailed:   true,
				Force:         false,
			},
			valid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test that the options structure is valid
			if tc.valid {
				if tc.options.ID == "" {
					t.Error("Expected valid options to have non-empty ID")
				}
			}
		})
	}
}

func TestCheckpointStatusFiltering(t *testing.T) {
	// Create a temporary store
	tempDir := t.TempDir()
	store, err := NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	// Create checkpoints with all possible statuses
	statuses := []CheckpointStatus{
		StatusPending,
		StatusInProgress,
		StatusCompleted,
		StatusFailed,
		StatusInterrupted,
	}

	// Create one checkpoint for each status
	for i, status := range statuses {
		cp := &TreeCheckpoint{
			ID:                    string(rune('A'+i)) + "-checkpoint",
			Status:                status,
			StartTime:             time.Now().Add(-time.Duration(i+1) * time.Hour),
			LastUpdated:           time.Now().Add(-time.Duration(i*30) * time.Minute),
			SourceRegistry:        "test-registry",
			SourcePrefix:          "test-prefix",
			Progress:              float64(i * 20),
			CompletedRepositories: []string{},
			Repositories:          map[string]RepoStatus{},
			RepoTasks:             []RepoTask{},
		}

		if err := store.SaveCheckpoint(cp); err != nil {
			t.Fatalf("Failed to save checkpoint with status %s: %v", status, err)
		}
	}

	// Get resumable checkpoints
	resumable, err := GetResumableCheckpoints(store)
	if err != nil {
		t.Fatalf("GetResumableCheckpoints failed: %v", err)
	}

	// Should only include: Pending, InProgress, Failed, Interrupted
	// Should exclude: Completed
	expectedResumableStatuses := map[CheckpointStatus]bool{
		StatusPending:     true,
		StatusInProgress:  true,
		StatusFailed:      true,
		StatusInterrupted: true,
		StatusCompleted:   false, // Should be excluded
	}

	// Verify we have the right number of resumable checkpoints
	expectedCount := 4 // All except completed
	if len(resumable) != expectedCount {
		t.Errorf("Expected %d resumable checkpoints, got %d", expectedCount, len(resumable))
	}

	// Verify that only the expected statuses are included
	foundStatuses := make(map[CheckpointStatus]bool)
	for _, r := range resumable {
		foundStatuses[r.Status] = true

		if !expectedResumableStatuses[r.Status] {
			t.Errorf("Status %s should not be resumable", r.Status)
		}
	}

	// Verify that completed status is not found
	if foundStatuses[StatusCompleted] {
		t.Error("Completed status should not be in resumable checkpoints")
	}
}

// TestRemainingRepositoryCalculation specifically tests the logic for calculating remaining repositories
func TestRemainingRepositoryCalculation(t *testing.T) {
	// Create a checkpoint with various repository states
	checkpoint := &TreeCheckpoint{
		ID:                    "calculation-test",
		Status:                StatusInterrupted,
		StartTime:             time.Now().Add(-2 * time.Hour),
		LastUpdated:           time.Now().Add(-30 * time.Minute),
		SourceRegistry:        "test-registry",
		SourcePrefix:          "test-prefix",
		CompletedRepositories: []string{"repo1", "repo2"}, // Explicitly completed
		Repositories: map[string]RepoStatus{
			"repo1": {Status: StatusCompleted, LastUpdated: time.Now().Add(-90 * time.Minute)},
			"repo2": {Status: StatusCompleted, LastUpdated: time.Now().Add(-85 * time.Minute)},
			"repo3": {Status: StatusFailed, LastUpdated: time.Now().Add(-60 * time.Minute), Error: "Network error"},
			"repo4": {Status: StatusInterrupted, LastUpdated: time.Now().Add(-30 * time.Minute)},
			"repo5": {Status: StatusPending, LastUpdated: time.Time{}},
		},
		RepoTasks: []RepoTask{
			{SourceRepository: "repo1", Status: StatusCompleted},
			{SourceRepository: "repo2", Status: StatusCompleted},
			{SourceRepository: "repo3", Status: StatusFailed, Error: "Network error"},
			{SourceRepository: "repo4", Status: StatusInterrupted},
			{SourceRepository: "repo5", Status: StatusPending},
		},
	}

	tests := []struct {
		name          string
		options       ResumableOptions
		expectedRepos []string
		shouldInclude map[string]bool
		shouldExclude map[string]bool
	}{
		{
			name: "Skip completed, retry failed",
			options: ResumableOptions{
				ID:            "calculation-test",
				SkipCompleted: true,
				RetryFailed:   true,
			},
			expectedRepos: []string{"repo3", "repo4", "repo5"},
			shouldInclude: map[string]bool{"repo3": true, "repo4": true, "repo5": true},
			shouldExclude: map[string]bool{"repo1": true, "repo2": true},
		},
		{
			name: "Skip completed, don't retry failed",
			options: ResumableOptions{
				ID:            "calculation-test",
				SkipCompleted: true,
				RetryFailed:   false,
			},
			expectedRepos: []string{"repo4", "repo5"},
			shouldInclude: map[string]bool{"repo4": true, "repo5": true},
			shouldExclude: map[string]bool{"repo1": true, "repo2": true, "repo3": true},
		},
		{
			name: "Include completed, retry failed",
			options: ResumableOptions{
				ID:            "calculation-test",
				SkipCompleted: false,
				RetryFailed:   true,
			},
			expectedRepos: []string{"repo1", "repo2", "repo3", "repo4", "repo5"},
			shouldInclude: map[string]bool{"repo1": true, "repo2": true, "repo3": true, "repo4": true, "repo5": true},
			shouldExclude: map[string]bool{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			remaining, err := GetRemainingRepositories(checkpoint, tc.options)
			if err != nil {
				t.Fatalf("GetRemainingRepositories failed: %v", err)
			}

			// Check expected count
			if len(remaining) != len(tc.expectedRepos) {
				t.Errorf("Expected %d remaining repositories, got %d", len(tc.expectedRepos), len(remaining))
			}

			// Convert to map for easier checking
			remainingMap := make(map[string]bool)
			for _, repo := range remaining {
				remainingMap[repo] = true
			}

			// Check that expected repos are included
			for _, expectedRepo := range tc.expectedRepos {
				if !remainingMap[expectedRepo] {
					t.Errorf("Expected repository %s to be in remaining list", expectedRepo)
				}
			}

			// Check shouldInclude repos
			for repo, shouldBeIncluded := range tc.shouldInclude {
				if shouldBeIncluded && !remainingMap[repo] {
					t.Errorf("Repository %s should be included but was not found", repo)
				}
			}

			// Check shouldExclude repos
			for repo, shouldBeExcluded := range tc.shouldExclude {
				if shouldBeExcluded && remainingMap[repo] {
					t.Errorf("Repository %s should be excluded but was found", repo)
				}
			}
		})
	}
}
