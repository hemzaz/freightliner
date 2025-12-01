package replication

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
)

// MockReplicationService implements ReplicationService for testing
type MockReplicationService struct {
	replicateCalls atomic.Int32
	shouldFail     bool
}

func (m *MockReplicationService) ReplicateRepository(ctx context.Context, rule ReplicationRule) error {
	m.replicateCalls.Add(1)
	if m.shouldFail {
		return &replicationError{"mock replication failure"}
	}
	return nil
}

type replicationError struct {
	msg string
}

func (e *replicationError) Error() string {
	return e.msg
}

func TestNewScheduler(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		RegistryProviders:  nil,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	if scheduler == nil {
		t.Fatal("Expected non-nil scheduler")
	}

	if scheduler.logger == nil {
		t.Error("Expected non-nil logger")
	}

	if scheduler.workerPool == nil {
		t.Error("Expected non-nil worker pool")
	}
}

func TestScheduler_AddJob_Success(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	rule := ReplicationRule{
		SourceRegistry:        "source-registry",
		SourceRepository:      "source/repo",
		DestinationRegistry:   "dest-registry",
		DestinationRepository: "dest/repo",
		Schedule:              "*/5 * * * * *", // Every 5 seconds
	}

	err := scheduler.AddJob(rule)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify job was added
	scheduler.mutex.RLock()
	jobCount := len(scheduler.jobs)
	scheduler.mutex.RUnlock()

	if jobCount != 1 {
		t.Errorf("Expected 1 job, got %d", jobCount)
	}
}

func TestScheduler_AddJob_ImmediateExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing-sensitive test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	rule := ReplicationRule{
		SourceRegistry:        "source-registry",
		SourceRepository:      "source/repo",
		DestinationRegistry:   "dest-registry",
		DestinationRepository: "dest/repo",
		Schedule:              "@now",
	}

	err := scheduler.AddJob(rule)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Manually trigger job check
	scheduler.checkJobs()

	// Give scheduler time to process
	time.Sleep(2 * time.Second)

	// Verify job was processed
	if mockSvc.replicateCalls.Load() < 1 {
		t.Error("Expected at least 1 replication call")
	}
}

func TestScheduler_AddJob_ValidationErrors(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	tests := []struct {
		name string
		rule ReplicationRule
		want string
	}{
		{
			name: "Empty source registry",
			rule: ReplicationRule{
				SourceRegistry:        "",
				SourceRepository:      "source/repo",
				DestinationRegistry:   "dest",
				DestinationRepository: "dest/repo",
				Schedule:              "* * * * * *",
			},
			want: "source registry cannot be empty",
		},
		{
			name: "Empty source repository",
			rule: ReplicationRule{
				SourceRegistry:        "source",
				SourceRepository:      "",
				DestinationRegistry:   "dest",
				DestinationRepository: "dest/repo",
				Schedule:              "* * * * * *",
			},
			want: "source repository cannot be empty",
		},
		{
			name: "Empty destination registry",
			rule: ReplicationRule{
				SourceRegistry:        "source",
				SourceRepository:      "source/repo",
				DestinationRegistry:   "",
				DestinationRepository: "dest/repo",
				Schedule:              "* * * * * *",
			},
			want: "destination registry cannot be empty",
		},
		{
			name: "Empty destination repository",
			rule: ReplicationRule{
				SourceRegistry:        "source",
				SourceRepository:      "source/repo",
				DestinationRegistry:   "dest",
				DestinationRepository: "",
				Schedule:              "* * * * * *",
			},
			want: "destination repository cannot be empty",
		},
		{
			name: "Invalid cron expression",
			rule: ReplicationRule{
				SourceRegistry:        "source",
				SourceRepository:      "source/repo",
				DestinationRegistry:   "dest",
				DestinationRepository: "dest/repo",
				Schedule:              "invalid cron",
			},
			want: "invalid cron expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scheduler.AddJob(tt.rule)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
		})
	}
}

func TestScheduler_AddJob_NoSchedule(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	rule := ReplicationRule{
		SourceRegistry:        "source-registry",
		SourceRepository:      "source/repo",
		DestinationRegistry:   "dest-registry",
		DestinationRepository: "dest/repo",
		Schedule:              "", // No schedule
	}

	err := scheduler.AddJob(rule)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify job was not added
	scheduler.mutex.RLock()
	jobCount := len(scheduler.jobs)
	scheduler.mutex.RUnlock()

	if jobCount != 0 {
		t.Errorf("Expected 0 jobs, got %d", jobCount)
	}
}

func TestScheduler_RemoveJob_Success(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	rule := ReplicationRule{
		SourceRegistry:        "source-registry",
		SourceRepository:      "source/repo",
		DestinationRegistry:   "dest-registry",
		DestinationRepository: "dest/repo",
		Schedule:              "* * * * * *",
	}

	// Add job
	err := scheduler.AddJob(rule)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Remove job
	err = scheduler.RemoveJob(rule)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify job was removed
	scheduler.mutex.RLock()
	jobCount := len(scheduler.jobs)
	scheduler.mutex.RUnlock()

	if jobCount != 0 {
		t.Errorf("Expected 0 jobs after removal, got %d", jobCount)
	}
}

func TestScheduler_RemoveJob_NotFound(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	rule := ReplicationRule{
		SourceRegistry:        "source-registry",
		SourceRepository:      "source/repo",
		DestinationRegistry:   "dest-registry",
		DestinationRepository: "dest/repo",
		Schedule:              "* * * * * *",
	}

	// Try to remove non-existent job
	err := scheduler.RemoveJob(rule)
	if err == nil {
		t.Fatal("Expected error for non-existent job, got nil")
	}
}

func TestScheduler_RemoveJob_ValidationErrors(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	tests := []struct {
		name string
		rule ReplicationRule
	}{
		{
			name: "Empty source registry",
			rule: ReplicationRule{
				SourceRegistry:        "",
				SourceRepository:      "source/repo",
				DestinationRegistry:   "dest",
				DestinationRepository: "dest/repo",
			},
		},
		{
			name: "Empty source repository",
			rule: ReplicationRule{
				SourceRegistry:        "source",
				SourceRepository:      "",
				DestinationRegistry:   "dest",
				DestinationRepository: "dest/repo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scheduler.RemoveJob(tt.rule)
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}
		})
	}
}

func TestScheduler_Stop(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)

	// Stop scheduler
	err := scheduler.Stop()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Try stopping again (should return error)
	err = scheduler.Stop()
	if err == nil {
		t.Error("Expected error when stopping already stopped scheduler")
	}
}

func TestScheduler_SubmitJob_WithoutWorkerPool(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         nil, // No worker pool
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	rule := ReplicationRule{
		SourceRegistry:        "source-registry",
		SourceRepository:      "source/repo",
		DestinationRegistry:   "dest-registry",
		DestinationRepository: "dest/repo",
		Schedule:              "@now",
	}

	err := scheduler.AddJob(rule)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Give time for submission attempt
	time.Sleep(2 * time.Second)

	// No replication should have occurred due to nil worker pool
	if mockSvc.replicateCalls.Load() > 0 {
		t.Error("Expected no replication calls with nil worker pool")
	}
}

func TestScheduler_SubmitJob_WithoutReplicationService(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: nil, // No replication service
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	rule := ReplicationRule{
		SourceRegistry:        "source-registry",
		SourceRepository:      "source/repo",
		DestinationRegistry:   "dest-registry",
		DestinationRepository: "dest/repo",
		Schedule:              "@now",
	}

	err := scheduler.AddJob(rule)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Give time for job processing
	time.Sleep(2 * time.Second)

	// Job should have been processed but failed due to nil replication service
}

func TestScheduler_JobExecution_OneTimeSchedule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scheduler execution test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	rule := ReplicationRule{
		SourceRegistry:        "source-registry",
		SourceRepository:      "source/repo",
		DestinationRegistry:   "dest-registry",
		DestinationRepository: "dest/repo",
		Schedule:              "@once",
	}

	err := scheduler.AddJob(rule)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Manually trigger job check
	scheduler.checkJobs()

	// Give time for job execution
	time.Sleep(2 * time.Second)

	// Verify job was executed
	if mockSvc.replicateCalls.Load() < 1 {
		t.Error("Expected at least 1 replication call")
	}
}

func TestScheduler_JobExecution_WithError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scheduler error test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	pool := NewWorkerPool(5, logger)
	pool.Start()
	defer pool.Stop()

	mockSvc := &MockReplicationService{shouldFail: true}

	opts := SchedulerOptions{
		Logger:             logger,
		WorkerPool:         pool,
		ReplicationService: mockSvc,
	}

	scheduler := NewScheduler(opts)
	defer scheduler.Stop()

	rule := ReplicationRule{
		SourceRegistry:        "source-registry",
		SourceRepository:      "source/repo",
		DestinationRegistry:   "dest-registry",
		DestinationRepository: "dest/repo",
		Schedule:              "@now",
	}

	err := scheduler.AddJob(rule)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Give time for job execution
	time.Sleep(2 * time.Second)

	// Verify job was attempted
	if mockSvc.replicateCalls.Load() < 1 {
		t.Error("Expected at least 1 replication attempt")
	}
}
