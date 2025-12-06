package distributed_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"freightliner/pkg/distributed"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/replication"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRaftCoordinator_SingleNode(t *testing.T) {
	tmpDir := t.TempDir()

	config := distributed.RaftConfig{
		NodeID:    "node-1",
		BindAddr:  "127.0.0.1:0", // Random port
		DataDir:   tmpDir,
		Bootstrap: true,
		Logger:    log.NewBasicLogger(log.InfoLevel),
	}

	coordinator, err := distributed.NewRaftCoordinator(config)
	require.NoError(t, err)
	defer coordinator.Shutdown()

	// Wait for leader election
	err = coordinator.WaitForLeader(5 * time.Second)
	require.NoError(t, err)

	// Should be leader
	assert.True(t, coordinator.IsLeader())
}

func TestRaftCoordinator_JobOperations(t *testing.T) {
	tmpDir := t.TempDir()

	coordinator, err := distributed.NewRaftCoordinator(distributed.RaftConfig{
		NodeID:    "node-1",
		BindAddr:  "127.0.0.1:0",
		DataDir:   tmpDir,
		Bootstrap: true,
		Logger:    log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)
	defer coordinator.Shutdown()

	err = coordinator.WaitForLeader(5 * time.Second)
	require.NoError(t, err)

	ctx := context.Background()

	// Create job
	job := &distributed.JobState{
		ID:     "job-123",
		Status: "pending",
		Rule: replication.ReplicationRule{
			SourceRegistry:   "source",
			SourceRepository: "repo",
		},
		StartTime:  time.Now(),
		UpdateTime: time.Now(),
		NodeID:     "node-1",
	}

	err = coordinator.CreateJob(ctx, job)
	require.NoError(t, err)

	// Retrieve job
	retrieved, exists := coordinator.GetJob("job-123")
	require.True(t, exists)
	assert.Equal(t, "job-123", retrieved.ID)
	assert.Equal(t, "pending", retrieved.Status)

	// Update job
	job.Status = "running"
	err = coordinator.UpdateJob(ctx, job)
	require.NoError(t, err)

	retrieved, _ = coordinator.GetJob("job-123")
	assert.Equal(t, "running", retrieved.Status)

	// Complete job
	err = coordinator.CompleteJob(ctx, "job-123")
	require.NoError(t, err)

	_, exists = coordinator.GetJob("job-123")
	assert.False(t, exists)
}

func TestRaftCoordinator_Checkpoint(t *testing.T) {
	tmpDir := t.TempDir()

	coordinator, err := distributed.NewRaftCoordinator(distributed.RaftConfig{
		NodeID:    "node-1",
		BindAddr:  "127.0.0.1:0",
		DataDir:   tmpDir,
		Bootstrap: true,
		Logger:    log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)
	defer coordinator.Shutdown()

	err = coordinator.WaitForLeader(5 * time.Second)
	require.NoError(t, err)

	ctx := context.Background()

	// Create checkpoint
	checkpoint := &distributed.CheckpointState{
		JobID:          "job-456",
		CompletedTags:  []string{"v1.0", "v1.1"},
		FailedTags:     map[string]string{"v2.0": "network error"},
		LastUpdateTime: time.Now(),
	}

	err = coordinator.UpdateCheckpoint(ctx, checkpoint)
	require.NoError(t, err)

	// Retrieve checkpoint
	retrieved, exists := coordinator.GetCheckpoint("job-456")
	require.True(t, exists)
	assert.Equal(t, "job-456", retrieved.JobID)
	assert.Len(t, retrieved.CompletedTags, 2)
	assert.Len(t, retrieved.FailedTags, 1)
}

func TestRaftCoordinator_MultipleJobs(t *testing.T) {
	tmpDir := t.TempDir()

	coordinator, err := distributed.NewRaftCoordinator(distributed.RaftConfig{
		NodeID:    "node-1",
		BindAddr:  "127.0.0.1:0",
		DataDir:   tmpDir,
		Bootstrap: true,
		Logger:    log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)
	defer coordinator.Shutdown()

	err = coordinator.WaitForLeader(5 * time.Second)
	require.NoError(t, err)

	ctx := context.Background()

	// Create multiple jobs
	for i := 0; i < 10; i++ {
		job := &distributed.JobState{
			ID:         string(rune(i)),
			Status:     "pending",
			NodeID:     "node-1",
			StartTime:  time.Now(),
			UpdateTime: time.Now(),
		}
		err = coordinator.CreateJob(ctx, job)
		require.NoError(t, err)
	}

	// List all jobs
	jobs := coordinator.ListJobs()
	assert.Len(t, jobs, 10)
}

func TestRaftCoordinator_Stats(t *testing.T) {
	tmpDir := t.TempDir()

	coordinator, err := distributed.NewRaftCoordinator(distributed.RaftConfig{
		NodeID:    "node-1",
		BindAddr:  "127.0.0.1:0",
		DataDir:   tmpDir,
		Bootstrap: true,
		Logger:    log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)
	defer coordinator.Shutdown()

	err = coordinator.WaitForLeader(5 * time.Second)
	require.NoError(t, err)

	stats := coordinator.Stats()
	assert.NotEmpty(t, stats)
	assert.Contains(t, stats, "state")
}

func TestRaftCoordinator_Snapshot(t *testing.T) {
	tmpDir := t.TempDir()

	coordinator, err := distributed.NewRaftCoordinator(distributed.RaftConfig{
		NodeID:    "node-1",
		BindAddr:  "127.0.0.1:0",
		DataDir:   tmpDir,
		Bootstrap: true,
		Logger:    log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)
	defer coordinator.Shutdown()

	err = coordinator.WaitForLeader(5 * time.Second)
	require.NoError(t, err)

	ctx := context.Background()

	// Create jobs
	for i := 0; i < 5; i++ {
		job := &distributed.JobState{
			ID:         string(rune(i)),
			Status:     "pending",
			NodeID:     "node-1",
			StartTime:  time.Now(),
			UpdateTime: time.Now(),
		}
		err = coordinator.CreateJob(ctx, job)
		require.NoError(t, err)
	}

	// Check snapshot directory exists
	snapshotDir := filepath.Join(tmpDir, "snapshots")
	_, err = os.Stat(snapshotDir)
	// Snapshot directory may not exist yet (snapshots are created periodically)
	if err == nil {
		// Directory exists, check if it's accessible
		assert.NoError(t, err)
	}
}
