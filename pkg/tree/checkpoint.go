package tree

import (
	"context"
	"fmt"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/tree/checkpoint"

	"github.com/google/uuid"
)

// InitCheckpointStore initializes a checkpoint store
func InitCheckpointStore(dir string) (checkpoint.CheckpointStore, error) {
	return checkpoint.NewFileStore(dir)
}

// ListResumableCheckpoints returns a list of resumable checkpoints
func ListResumableCheckpoints(store checkpoint.CheckpointStore) ([]checkpoint.ResumableCheckpoint, error) {
	return checkpoint.GetResumableCheckpoints(store)
}

// CheckpointManager handles checkpoint operations for tree replication
type CheckpointManager struct {
	store  checkpoint.CheckpointStore
	logger log.Logger

	// Current checkpoint being used
	current *checkpoint.TreeCheckpoint

	// Options for resumption
	resumeOpts checkpoint.ResumableOptions

	// Whether checkpointing is enabled
	enabled bool
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(store checkpoint.CheckpointStore, logger log.Logger) *CheckpointManager {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &CheckpointManager{
		store:   store,
		logger:  logger,
		enabled: true,
	}
}

// StartNewCheckpoint creates a new checkpoint for a tree replication
func (m *CheckpointManager) StartNewCheckpoint(sourceReg, sourcePrefix, destReg, destPrefix string) (string, error) {
	if !m.enabled {
		return "", nil
	}

	// Generate a unique ID for this checkpoint
	id := uuid.New().String()

	// Create a new checkpoint
	cp := &checkpoint.TreeCheckpoint{
		ID:             id,
		StartTime:      time.Now(),
		LastUpdated:    time.Now(),
		SourceRegistry: sourceReg,
		SourcePrefix:   sourcePrefix,
		DestRegistry:   destReg,
		DestPrefix:     destPrefix,
		Status:         checkpoint.StatusInProgress,
		Repositories:   make(map[string]checkpoint.RepoStatus),
		Progress:       0.0,
	}

	// Save the checkpoint
	if err := m.store.SaveCheckpoint(cp); err != nil {
		return "", errors.Wrap(err, "failed to save initial checkpoint")
	}

	m.current = cp
	return id, nil
}

// ResumeCheckpoint loads a checkpoint for resumption
func (m *CheckpointManager) ResumeCheckpoint(id string, opts checkpoint.ResumableOptions) error {
	if !m.enabled {
		return nil
	}

	// Load the checkpoint
	cp, err := m.store.LoadCheckpoint(id)
	if err != nil {
		return errors.Wrap(err, "failed to load checkpoint")
	}

	// Update checkpoint status
	cp.Status = checkpoint.StatusInProgress
	cp.LastUpdated = time.Now()

	// Save the updated checkpoint
	if err := m.store.SaveCheckpoint(cp); err != nil {
		return errors.Wrap(err, "failed to update resumed checkpoint")
	}

	m.current = cp
	m.resumeOpts = opts
	return nil
}

// GetRemainingRepositories returns repositories that need to be processed based on resume options
func (m *CheckpointManager) GetRemainingRepositories() ([]string, error) {
	if !m.enabled || m.current == nil {
		return nil, errors.InvalidInputf("no active checkpoint")
	}

	return checkpoint.GetRemainingRepositories(m.current, m.resumeOpts)
}

// AddRepository adds a repository to be tracked in the checkpoint
func (m *CheckpointManager) AddRepository(ctx context.Context, sourceRepo, destRepo string) error {
	if !m.enabled || m.current == nil {
		return nil
	}

	m.current.Repositories[sourceRepo] = checkpoint.RepoStatus{
		Status:      checkpoint.StatusPending,
		SourceRepo:  sourceRepo,
		DestRepo:    destRepo,
		LastUpdated: time.Now(),
	}

	return m.updateCheckpoint(ctx)
}

// UpdateRepositoryStatus updates the status of a repository
func (m *CheckpointManager) UpdateRepositoryStatus(ctx context.Context, repo string, status checkpoint.Status, errMsg string) error {
	if !m.enabled || m.current == nil {
		return nil
	}

	// Check if the repository exists
	repoStatus, exists := m.current.Repositories[repo]
	if !exists {
		return errors.NotFoundf("repository not found in checkpoint: %s", repo)
	}

	// Update status
	repoStatus.Status = status
	repoStatus.LastUpdated = time.Now()

	if errMsg != "" {
		repoStatus.Error = errMsg
	}

	// If the repository is completed, add to completed list
	if status == checkpoint.StatusCompleted {
		m.current.CompletedRepositories = append(m.current.CompletedRepositories, repo)
	}

	// Update the repository status
	m.current.Repositories[repo] = repoStatus

	// Calculate progress
	totalRepos := len(m.current.Repositories)
	if totalRepos > 0 {
		completedRepos := len(m.current.CompletedRepositories)
		m.current.Progress = float64(completedRepos) / float64(totalRepos) * 100.0
	}

	return m.updateCheckpoint(ctx)
}

// CompleteCheckpoint marks the checkpoint as completed
func (m *CheckpointManager) CompleteCheckpoint(ctx context.Context) error {
	if !m.enabled || m.current == nil {
		return nil
	}

	m.current.Status = checkpoint.StatusCompleted
	m.current.LastUpdated = time.Now()
	m.current.Progress = 100.0

	return m.updateCheckpoint(ctx)
}

// FailCheckpoint marks the checkpoint as failed
func (m *CheckpointManager) FailCheckpoint(ctx context.Context, reason string) error {
	if !m.enabled || m.current == nil {
		return nil
	}

	m.current.Status = checkpoint.StatusFailed
	m.current.LastError = reason
	m.current.LastUpdated = time.Now()

	return m.updateCheckpoint(ctx)
}

// InterruptCheckpoint marks the checkpoint as interrupted
func (m *CheckpointManager) InterruptCheckpoint(ctx context.Context) error {
	if !m.enabled || m.current == nil {
		return nil
	}

	m.current.Status = checkpoint.StatusInterrupted
	m.current.LastUpdated = time.Now()

	return m.updateCheckpoint(ctx)
}

// updateCheckpoint saves the current checkpoint to storage
func (m *CheckpointManager) updateCheckpoint(ctx context.Context) error {
	if ctx.Err() != nil {
		// If context is cancelled, mark as interrupted
		m.current.Status = checkpoint.StatusInterrupted
	}

	m.current.LastUpdated = time.Now()
	if err := m.store.SaveCheckpoint(m.current); err != nil {
		m.logger.WithFields(map[string]interface{}{
			"checkpoint_id": m.current.ID,
		}).Error("Failed to save checkpoint", err)
		return errors.Wrap(err, "failed to save checkpoint")
	}

	return nil
}

// GetCheckpoint returns the current checkpoint
func (m *CheckpointManager) GetCheckpoint() *checkpoint.TreeCheckpoint {
	return m.current
}

// DeleteCheckpoint deletes a checkpoint by ID
func (m *CheckpointManager) DeleteCheckpoint(id string) error {
	if id == "" {
		return errors.InvalidInputf("checkpoint ID cannot be empty")
	}

	return m.store.DeleteCheckpoint(id)
}

// EnableCheckpointing enables or disables checkpointing
func (m *CheckpointManager) EnableCheckpointing(enabled bool) {
	m.enabled = enabled
}

// IsCheckpointingEnabled returns whether checkpointing is enabled
func (m *CheckpointManager) IsCheckpointingEnabled() bool {
	return m.enabled
}

// GetCheckpointSummary returns a human-readable summary of a checkpoint
func GetCheckpointSummary(cp *checkpoint.TreeCheckpoint) string {
	if cp == nil {
		return "No checkpoint"
	}

	// Count different statuses
	var completed, pending, inProgress, failed, skipped int
	for _, status := range cp.Repositories {
		switch status.Status {
		case checkpoint.StatusCompleted:
			completed++
		case checkpoint.StatusPending:
			pending++
		case checkpoint.StatusInProgress:
			inProgress++
		case checkpoint.StatusFailed:
			failed++
		case checkpoint.StatusSkipped:
			skipped++
		}
	}

	// Format duration
	duration := time.Since(cp.StartTime)

	return fmt.Sprintf(
		"Checkpoint %s: %s [%.1f%%]\n"+
			"Source: %s/%s → Destination: %s/%s\n"+
			"Started: %s (duration: %s)\n"+
			"Repositories: %d total, %d completed, %d pending, %d in progress, %d failed, %d skipped",
		cp.ID,
		cp.Status,
		cp.Progress,
		cp.SourceRegistry,
		cp.SourcePrefix,
		cp.DestRegistry,
		cp.DestPrefix,
		cp.StartTime.Format(time.RFC3339),
		duration.Round(time.Second),
		len(cp.Repositories),
		completed,
		pending,
		inProgress,
		failed,
		skipped,
	)
}
