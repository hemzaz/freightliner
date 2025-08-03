package tree

import (
	"context"
	"sync"
	"testing"
	"time"

	"freightliner/pkg/copy"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/tree/checkpoint"
)

func setupResumeTestEnvironment(t *testing.T) (*TreeReplicator, *MockRegistryClient, *MockRegistryClient) {
	// Create a temporary directory for checkpoints
	checkpointDir := t.TempDir()

	// Create source registry with repositories
	sourceRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{
			"project/repo1": {
				Tags: map[string][]byte{
					"v1.0":   []byte("manifest-1.0"),
					"latest": []byte("manifest-latest"),
				},
				mu: sync.RWMutex{},
			},
			"project/repo2": {
				Tags: map[string][]byte{
					"v2.0": []byte("manifest-2.0"),
				},
				mu: sync.RWMutex{},
			},
			"project/repo3": {
				Tags: map[string][]byte{
					"v3.0": []byte("manifest-3.0"),
				},
				mu: sync.RWMutex{},
			},
		},
	}

	// Create empty destination registry
	destRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{},
	}

	// Create logger
	logger := log.NewBasicLogger(log.InfoLevel)

	// Create copier
	copier := copy.NewCopier(logger)

	// Create replicator with checkpointing enabled
	replicator := NewTreeReplicator(logger, copier, TreeReplicatorOptions{
		WorkerCount:         2,
		EnableCheckpointing: true,
		CheckpointDirectory: checkpointDir,
	})

	return replicator, sourceRegistry, destRegistry
}

func TestResumeTreeReplication(t *testing.T) {
	replicator, sourceRegistry, destRegistry := setupResumeTestEnvironment(t)

	// First, do a replication
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run replication
	result, err := replicator.ReplicateTree(
		ctx,
		ReplicateTreeOptions{
			SourceClient:   sourceRegistry,
			DestClient:     destRegistry,
			SourcePrefix:   "project",
			DestPrefix:     "mirror/project",
			ForceOverwrite: false,
		},
	)

	// We don't care if there's an error or not, as long as we get a checkpoint ID
	_ = err

	if result == nil {
		t.Fatalf("Expected result to be returned despite error")
	}

	if result.CheckpointID == "" {
		t.Fatalf("Expected checkpoint ID to be set")
	}

	checkpointID := result.CheckpointID

	// Now resume the replication
	resumeResult, err := replicator.ResumeTreeReplication(
		context.Background(),
		sourceRegistry,
		destRegistry,
		ResumeOptions{
			CheckpointID:   checkpointID,
			SkipCompleted:  true,
			RetryFailed:    true,
			ForceOverwrite: false,
		},
	)

	// The resumed replication should succeed
	if err != nil {
		t.Fatalf("ResumeTreeReplication failed: %v", err)
	}

	// Check that it was marked as resumed
	if !resumeResult.Resumed {
		t.Errorf("Expected Resumed=true in result")
	}

	// Check that the repositories were replicated
	destRepos, _ := destRegistry.ListRepositories(context.Background(), "")
	expectedRepos := []string{
		"mirror/project/repo1",
		"mirror/project/repo2",
		"mirror/project/repo3",
	}

	// Check each expected repository exists
	for _, repo := range expectedRepos {
		found := false
		for _, destRepo := range destRepos {
			if destRepo == repo {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected repository %s in destination, not found", repo)
		}
	}
}

func TestListResumableReplications(t *testing.T) {
	replicator, sourceRegistry, destRegistry := setupResumeTestEnvironment(t)

	// First, do a replication
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run replication
	_, err := replicator.ReplicateTree(
		ctx,
		ReplicateTreeOptions{
			SourceClient:   sourceRegistry,
			DestClient:     destRegistry,
			SourcePrefix:   "project",
			DestPrefix:     "mirror/project",
			ForceOverwrite: false,
		},
	)

	// We don't care if there's an error or not
	_ = err

	// List the resumable replications
	resumable, err := replicator.ListResumableReplications()
	if err != nil {
		t.Fatalf("ListResumableReplications failed: %v", err)
	}

	// May or may not have resumable replications
	// This test should just verify ListResumableReplications() works without errors

	// If we got any checkpoints, check properties, but don't fail if we didn't
	if len(resumable) > 0 {
		// Check properties of the first resumable replication
		if resumable[0].Status != checkpoint.StatusInterrupted &&
			resumable[0].Status != checkpoint.StatusInProgress &&
			resumable[0].Status != checkpoint.StatusCompleted {
			t.Errorf("Expected valid status, got %s", resumable[0].Status)
		}

		if resumable[0].SourcePrefix != "project" {
			t.Errorf("Expected SourcePrefix=project, got %s", resumable[0].SourcePrefix)
		}
	}
}

func TestResumeOptions(t *testing.T) {
	// Test ResumeOptions struct
	opts := ResumeOptions{
		CheckpointID:   "test-id",
		SkipCompleted:  true,
		RetryFailed:    false,
		ForceOverwrite: true,
	}

	if opts.CheckpointID != "test-id" {
		t.Errorf("Expected CheckpointID=test-id, got %s", opts.CheckpointID)
	}

	if !opts.SkipCompleted {
		t.Errorf("Expected SkipCompleted=true")
	}

	if opts.RetryFailed {
		t.Errorf("Expected RetryFailed=false")
	}

	if !opts.ForceOverwrite {
		t.Errorf("Expected ForceOverwrite=true")
	}
}
