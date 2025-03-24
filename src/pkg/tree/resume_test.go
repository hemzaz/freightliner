package tree

import (
	"context"
	"testing"
	"time"

	"github.com/hemzaz/freightliner/src/internal/log"
	"github.com/hemzaz/freightliner/src/pkg/copy"
	"github.com/hemzaz/freightliner/src/pkg/tree/checkpoint"
)

func setupResumeTestEnvironment(t *testing.T) (*TreeReplicator, *MockRegistryClient, *MockRegistryClient, string) {
	// Create a temporary directory for checkpoints
	checkpointDir := t.TempDir()
	
	// Create source registry with repositories
	sourceRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{
			"project/repo1": {
				Tags: map[string][]byte{
					"v1.0": []byte("manifest-1.0"),
					"latest": []byte("manifest-latest"),
				},
			},
			"project/repo2": {
				Tags: map[string][]byte{
					"v2.0": []byte("manifest-2.0"),
				},
			},
			"project/repo3": {
				Tags: map[string][]byte{
					"v3.0": []byte("manifest-3.0"),
				},
			},
		},
	}
	
	// Create empty destination registry
	destRegistry := &MockRegistryClient{
		Repositories: map[string]*MockRepository{},
	}
	
	// Create logger
	logger := log.NewLogger(log.InfoLevel)
	
	// Create copier
	copier := copy.NewCopier(logger)
	
	// Create replicator with checkpointing enabled
	replicator := NewTreeReplicator(logger, copier, TreeReplicatorOptions{
		WorkerCount:       2,
		EnableCheckpoints: true,
		CheckpointDir:     checkpointDir,
	})
	
	return replicator, sourceRegistry, destRegistry, checkpointDir
}

func TestResumeTreeReplication(t *testing.T) {
	replicator, sourceRegistry, destRegistry, checkpointDir := setupResumeTestEnvironment(t)
	
	// First, do a partial replication that gets interrupted
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// Run replication with a timeout to simulate interruption
	result, err := replicator.ReplicateTree(
		ctx,
		sourceRegistry,
		destRegistry,
		"project",
		"mirror/project",
		false,
	)
	
	// The replication should be interrupted
	if err == nil {
		t.Fatalf("Expected error due to context timeout")
	}
	
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
			CheckpointID:    checkpointID,
			SkipCompleted:   true,
			RetryFailed:     true,
			ForceOverwrite:  false,
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
	destRepos, _ := destRegistry.ListRepositories()
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
	replicator, sourceRegistry, destRegistry, _ := setupResumeTestEnvironment(t)
	
	// First, do a partial replication that gets interrupted
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// Run replication with a timeout to simulate interruption
	_, err := replicator.ReplicateTree(
		ctx,
		sourceRegistry,
		destRegistry,
		"project",
		"mirror/project",
		false,
	)
	
	// The replication should be interrupted
	if err == nil {
		t.Fatalf("Expected error due to context timeout")
	}
	
	// List the resumable replications
	resumable, err := replicator.ListResumableReplications()
	if err != nil {
		t.Fatalf("ListResumableReplications failed: %v", err)
	}
	
	// Should have at least one resumable replication
	if len(resumable) == 0 {
		t.Errorf("Expected at least one resumable replication")
	}
	
	// Check properties of the first resumable replication
	if resumable[0].Status != checkpoint.StatusInterrupted && 
	   resumable[0].Status != checkpoint.StatusInProgress {
		t.Errorf("Expected status Interrupted or InProgress, got %s", resumable[0].Status)
	}
	
	if resumable[0].SourcePrefix != "project" {
		t.Errorf("Expected SourcePrefix=project, got %s", resumable[0].SourcePrefix)
	}
}

func TestResumeOptions(t *testing.T) {
	// Test ResumeOptions struct
	opts := ResumeOptions{
		CheckpointID:    "test-id",
		SkipCompleted:   true,
		RetryFailed:     false,
		ForceOverwrite:  true,
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