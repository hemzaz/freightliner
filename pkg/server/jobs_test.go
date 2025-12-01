package server

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"freightliner/pkg/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJobManager tests the job manager functionality
func TestJobManager(t *testing.T) {
	manager := NewJobManager()
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.jobs)
}

// TestJobManagerAddJob tests adding jobs
func TestJobManagerAddJob(t *testing.T) {
	manager := NewJobManager()

	job := NewReplicateJob("ecr/source", "gcr/dest", []string{"latest"}, false, false, &mockReplicationService{})
	manager.AddJob(job)

	retrieved, exists := manager.GetJob(job.GetID())
	assert.True(t, exists)
	assert.Equal(t, job.GetID(), retrieved.GetID())
}

// TestJobManagerGetJob tests job retrieval
func TestJobManagerGetJob(t *testing.T) {
	manager := NewJobManager()

	job := NewReplicateJob("ecr/source", "gcr/dest", []string{"latest"}, false, false, &mockReplicationService{})
	manager.AddJob(job)

	// Test existing job
	retrieved, exists := manager.GetJob(job.GetID())
	assert.True(t, exists)
	assert.NotNil(t, retrieved)

	// Test non-existent job
	_, exists = manager.GetJob("non-existent")
	assert.False(t, exists)
}

// TestJobManagerListJobs tests job listing
func TestJobManagerListJobs(t *testing.T) {
	manager := NewJobManager()

	// Add multiple jobs
	job1 := NewReplicateJob("ecr/repo1", "gcr/repo1", []string{"latest"}, false, false, &mockReplicationService{})
	job2 := NewReplicateJob("ecr/repo2", "gcr/repo2", []string{"v1.0"}, false, false, &mockReplicationService{})
	job3 := NewReplicateTreeJob("ecr/test", "gcr/test", map[string]interface{}{}, &service.TreeReplicationService{})

	manager.AddJob(job1)
	manager.AddJob(job2)
	manager.AddJob(job3)

	// List all jobs
	allJobs := manager.ListJobs("", "")
	assert.Len(t, allJobs, 3)

	// Filter by type
	replicateJobs := manager.ListJobs(JobTypeReplicate, "")
	assert.Len(t, replicateJobs, 2)

	treeJobs := manager.ListJobs(JobTypeReplicateTree, "")
	assert.Len(t, treeJobs, 1)

	// Filter by status
	pendingJobs := manager.ListJobs("", JobStatusPending)
	assert.Len(t, pendingJobs, 3)

	// Combined filter
	pendingReplicateJobs := manager.ListJobs(JobTypeReplicate, JobStatusPending)
	assert.Len(t, pendingReplicateJobs, 2)
}

// TestJobManagerUpdateJob tests job updates
func TestJobManagerUpdateJob(t *testing.T) {
	manager := NewJobManager()

	job := NewReplicateJob("ecr/source", "gcr/dest", []string{"latest"}, false, false, &mockReplicationService{})
	manager.AddJob(job)

	// Update job
	manager.UpdateJob(job.GetID(), JobStatusCompleted, map[string]int{"tags": 5}, nil)

	retrieved, _ := manager.GetJob(job.GetID())
	assert.Equal(t, JobStatusCompleted, retrieved.GetStatus())
	assert.NotNil(t, retrieved.GetResult())
	assert.NotZero(t, retrieved.GetEndTime())
}

// TestBaseJob tests the base job functionality
func TestBaseJob(t *testing.T) {
	job := NewBaseJob(JobTypeReplicate, "ecr/source", "gcr/dest")

	assert.NotEmpty(t, job.GetID())
	assert.Equal(t, JobTypeReplicate, job.GetType())
	assert.Equal(t, "ecr/source", job.GetSource())
	assert.Equal(t, "gcr/dest", job.GetDestination())
	assert.Equal(t, JobStatusPending, job.GetStatus())
	assert.NotZero(t, job.GetStartTime())
}

// TestBaseJobSetters tests base job setters
func TestBaseJobSetters(t *testing.T) {
	job := NewBaseJob(JobTypeReplicate, "ecr/source", "gcr/dest")

	// Test SetStatus
	job.SetStatus(JobStatusRunning)
	assert.Equal(t, JobStatusRunning, job.GetStatus())

	// Test SetResult
	result := map[string]int{"count": 42}
	job.SetResult(result)
	assert.Equal(t, result, job.GetResult())

	// Test SetError
	err := assert.AnError
	job.SetError(err)
	assert.Equal(t, err, job.GetError())
	assert.Equal(t, err.Error(), job.ErrorMsg)

	// Test SetEndTime
	endTime := time.Now()
	job.SetEndTime(endTime)
	assert.Equal(t, endTime, job.GetEndTime())
}

// TestBaseJobToJSON tests JSON serialization
func TestBaseJobToJSON(t *testing.T) {
	job := NewBaseJob(JobTypeReplicate, "ecr/source", "gcr/dest")
	job.SetStatus(JobStatusCompleted)
	job.SetResult(map[string]string{"status": "ok"})

	jsonData, err := job.ToJSON()
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, job.GetID(), decoded["id"])
	assert.Equal(t, string(JobTypeReplicate), decoded["type"])
	assert.Equal(t, "ecr/source", decoded["source"])
	assert.Equal(t, "gcr/dest", decoded["destination"])
}

// TestReplicateJob tests replicate job functionality
func TestReplicateJob(t *testing.T) {
	mockSvc := &mockReplicationService{}
	job := NewReplicateJob("ecr/source", "gcr/dest", []string{"latest", "v1.0"}, true, false, mockSvc)

	assert.Equal(t, JobTypeReplicate, job.GetType())
	assert.Len(t, job.Tags, 2)
	assert.True(t, job.Force)
	assert.False(t, job.DryRun)
}

// TestReplicateJobExecute tests job execution
func TestReplicateJobExecute(t *testing.T) {
	mockSvc := &mockReplicationService{}
	job := NewReplicateJob("ecr/source", "gcr/dest", []string{"latest"}, false, false, mockSvc)

	ctx := context.Background()
	err := job.Execute(ctx)

	require.NoError(t, err)
	assert.Equal(t, JobStatusCompleted, job.GetStatus())
	assert.NotNil(t, job.GetResult())
	assert.NotZero(t, job.GetEndTime())
}

// TestReplicateTreeJob tests tree replication job
func TestReplicateTreeJob(t *testing.T) {
	options := map[string]interface{}{
		"excludeRepos":     []string{"old-repo"},
		"excludeTags":      []string{"alpha"},
		"includeTags":      []string{"latest"},
		"force":            true,
		"dryRun":           false,
		"enableCheckpoint": true,
		"checkpointDir":    "/tmp/checkpoints",
		"resumeID":         "resume-123",
		"skipCompleted":    true,
		"retryFailed":      true,
	}

	job := NewReplicateTreeJob("ecr/source", "gcr/dest", options, &service.TreeReplicationService{})

	assert.Equal(t, JobTypeReplicateTree, job.GetType())
	assert.Len(t, job.ExcludeRepos, 1)
	assert.Len(t, job.ExcludeTags, 1)
	assert.Len(t, job.IncludeTags, 1)
	assert.True(t, job.Force)
	assert.False(t, job.DryRun)
	assert.True(t, job.EnableCheckpoint)
	assert.Equal(t, "/tmp/checkpoints", job.CheckpointDir)
	assert.Equal(t, "resume-123", job.ResumeID)
	assert.True(t, job.SkipCompleted)
	assert.True(t, job.RetryFailed)
}

// TestReplicateTreeJobExecute tests tree job execution
func TestReplicateTreeJobExecute(t *testing.T) {
	t.Skip("Skipping tree job execution test - requires full service initialization with registry credentials")

	// This test requires a properly initialized TreeReplicationService with:
	// - Valid configuration (cfg.TreeReplicate with all fields)
	// - Logger instance
	// - ReplicationService dependency
	// - Registry credentials for source and destination
	// These dependencies are better tested in integration tests rather than unit tests
}

// TestJobTypes tests job type constants
func TestJobTypes(t *testing.T) {
	assert.Equal(t, JobType("replicate"), JobTypeReplicate)
	assert.Equal(t, JobType("replicate-tree"), JobTypeReplicateTree)
	assert.Equal(t, JobType("checkpoint"), JobTypeCheckpoint)
}

// TestJobStatuses tests job status constants
func TestJobStatuses(t *testing.T) {
	assert.Equal(t, JobStatus("pending"), JobStatusPending)
	assert.Equal(t, JobStatus("running"), JobStatusRunning)
	assert.Equal(t, JobStatus("completed"), JobStatusCompleted)
	assert.Equal(t, JobStatus("failed"), JobStatusFailed)
	assert.Equal(t, JobStatus("canceled"), JobStatusCanceled)
}

// TestJobConcurrentAccess tests concurrent job manager access
func TestJobConcurrentAccess(t *testing.T) {
	manager := NewJobManager()

	// Add jobs concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			job := NewReplicateJob("ecr/source", "gcr/dest", []string{"latest"}, false, false, &mockReplicationService{})
			manager.AddJob(job)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all jobs were added
	jobs := manager.ListJobs("", "")
	assert.Len(t, jobs, 10)
}

// TestJobOptionsExtraction tests options extraction for tree jobs
func TestJobOptionsExtraction(t *testing.T) {
	options := map[string]interface{}{
		"excludeRepos": []string{"repo1", "repo2"},
		"excludeTags":  []string{"tag1"},
		"includeTags":  []string{"tag2", "tag3"},
		"force":        true,
		"dryRun":       true,
	}

	job := NewReplicateTreeJob("ecr/source", "gcr/dest", options, &service.TreeReplicationService{})

	assert.Equal(t, []string{"repo1", "repo2"}, job.ExcludeRepos)
	assert.Equal(t, []string{"tag1"}, job.ExcludeTags)
	assert.Equal(t, []string{"tag2", "tag3"}, job.IncludeTags)
	assert.True(t, job.Force)
	assert.True(t, job.DryRun)
}

// TestJobWithMissingOptions tests job creation with missing options
func TestJobWithMissingOptions(t *testing.T) {
	// Empty options map
	job := NewReplicateTreeJob("ecr/source", "gcr/dest", map[string]interface{}{}, &service.TreeReplicationService{})

	assert.Empty(t, job.ExcludeRepos)
	assert.Empty(t, job.ExcludeTags)
	assert.Empty(t, job.IncludeTags)
	assert.False(t, job.Force)
	assert.False(t, job.DryRun)
	assert.False(t, job.EnableCheckpoint)
}
