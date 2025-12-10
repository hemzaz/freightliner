package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/tree/checkpoint"
)

// CheckpointService handles checkpoint operations
type CheckpointService struct {
	cfg    *config.Config
	logger log.Logger
	store  checkpoint.CheckpointStore
}

// CheckpointInfo represents checkpoint information
type CheckpointInfo struct {
	ID                    string           `json:"id"`
	CreatedAt             time.Time        `json:"created_at"`
	Source                string           `json:"source"`
	Destination           string           `json:"destination"`
	Status                string           `json:"status"`
	TotalRepositories     int              `json:"total_repositories"`
	CompletedRepositories int              `json:"completed_repositories"`
	FailedRepositories    int              `json:"failed_repositories"`
	TotalTagsCopied       int              `json:"total_tags_copied"`
	TotalTagsSkipped      int              `json:"total_tags_skipped"`
	TotalErrors           int              `json:"total_errors"`
	TotalBytesTransferred int64            `json:"total_bytes_transferred"`
	Repositories          []RepositoryInfo `json:"repositories,omitempty"`
}

// RepositoryInfo represents repository information within a checkpoint
type RepositoryInfo struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	TagsCopied  int    `json:"tags_copied"`
	TagsSkipped int    `json:"tags_skipped"`
	Errors      int    `json:"errors"`
}

// NewCheckpointService creates a new checkpoint service
func NewCheckpointService(cfg *config.Config, logger log.Logger) *CheckpointService {
	return &CheckpointService{
		cfg:    cfg,
		logger: logger,
	}
}

// initStore initializes the checkpoint store
func (s *CheckpointService) initStore(ctx context.Context) error {
	if s.store != nil {
		return nil
	}

	// Expand checkpoint directory path
	dir := config.ExpandHomeDir(s.cfg.Checkpoint.Directory)

	s.logger.WithFields(map[string]interface{}{
		"directory": dir,
	}).Debug("Initializing checkpoint store")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.Wrap(err, "failed to create checkpoint directory")
	}

	// Check directory permissions (should be owner-only)
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return errors.Wrap(err, "failed to stat checkpoint directory")
	}

	// On Unix-like systems, check if permissions are 0700 (owner-only)
	if fileInfo.Mode().Perm() != 0700 {
		s.logger.WithFields(map[string]interface{}{
			"directory":   dir,
			"permissions": fileInfo.Mode().String(),
			"recommended": "0700",
		}).Warn("Checkpoint directory has insecure permissions, fixing")

		// Fix permissions
		if chmodErr := os.Chmod(dir, 0700); chmodErr != nil { // #nosec G302 - directory needs executable bit for access
			return errors.Wrap(chmodErr, "failed to fix checkpoint directory permissions")
		}
	}

	// Initialize store
	store, err := checkpoint.NewFileStore(dir)
	if err != nil {
		return errors.Wrap(err, "failed to initialize checkpoint store")
	}

	s.store = store
	return nil
}

// ListCheckpoints lists all checkpoints
func (s *CheckpointService) ListCheckpoints(ctx context.Context) ([]CheckpointInfo, error) {
	if err := s.initStore(ctx); err != nil {
		return nil, err
	}

	// Get all checkpoints
	checkpoints, err := s.store.ListCheckpoints()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list checkpoints")
	}

	s.logger.WithFields(map[string]interface{}{
		"count": len(checkpoints),
	}).Debug("Listed checkpoints")

	// Create result array
	result := make([]CheckpointInfo, 0, len(checkpoints))

	// Get checkpoint info for each checkpoint
	for _, cp := range checkpoints {
		// Convert to checkpoint info
		info := s.convertCheckpointToInfo(cp)
		result = append(result, info)
	}

	return result, nil
}

// GetCheckpoint retrieves a specific checkpoint
func (s *CheckpointService) GetCheckpoint(ctx context.Context, id string) (*CheckpointInfo, error) {
	if err := s.initStore(ctx); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, errors.InvalidInputf("checkpoint ID is required")
	}

	s.logger.WithFields(map[string]interface{}{
		"id": id,
	}).Debug("Loading checkpoint")

	// Load checkpoint from store
	cp, err := s.store.LoadCheckpoint(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load checkpoint")
	}

	// Convert to checkpoint info
	info := s.convertCheckpointToInfo(cp)

	return &info, nil
}

// DeleteCheckpoint deletes a specific checkpoint
func (s *CheckpointService) DeleteCheckpoint(ctx context.Context, id string) error {
	if err := s.initStore(ctx); err != nil {
		return err
	}

	if id == "" {
		return errors.InvalidInputf("checkpoint ID is required")
	}

	s.logger.WithFields(map[string]interface{}{
		"id": id,
	}).Debug("Deleting checkpoint")

	// Delete checkpoint from store
	if err := s.store.DeleteCheckpoint(id); err != nil {
		return errors.Wrap(err, "failed to delete checkpoint")
	}

	s.logger.WithFields(map[string]interface{}{
		"id": id,
	}).Info("Checkpoint deleted")

	return nil
}

// ExportCheckpoint exports a checkpoint to a file
func (s *CheckpointService) ExportCheckpoint(ctx context.Context, id string, filePath string) error {
	if err := s.initStore(ctx); err != nil {
		return err
	}

	if id == "" {
		return errors.InvalidInputf("checkpoint ID is required")
	}

	if filePath == "" {
		return errors.InvalidInputf("export file path is required")
	}

	s.logger.WithFields(map[string]interface{}{
		"id":   id,
		"path": filePath,
	}).Debug("Exporting checkpoint")

	// Load checkpoint from store
	cp, err := s.store.LoadCheckpoint(id)
	if err != nil {
		return errors.Wrap(err, "failed to load checkpoint")
	}

	// Convert to checkpoint info
	info := s.convertCheckpointToInfo(cp)

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if mkdirErr := os.MkdirAll(dir, 0750); mkdirErr != nil {
		return errors.Wrap(mkdirErr, "failed to create directory for export file")
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to create export file")
	}
	defer func() {
		_ = file.Close()
	}()

	// Marshal to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print
	if err := encoder.Encode(info); err != nil {
		return errors.Wrap(err, "failed to export checkpoint to JSON")
	}

	s.logger.WithFields(map[string]interface{}{
		"id":   id,
		"path": filePath,
	}).Info("Checkpoint exported")

	return nil
}

// ImportCheckpoint imports a checkpoint from a file
func (s *CheckpointService) ImportCheckpoint(ctx context.Context, filePath string) (*CheckpointInfo, error) {
	if err := s.initStore(ctx); err != nil {
		return nil, err
	}

	if filePath == "" {
		return nil, errors.InvalidInputf("import file path is required")
	}

	s.logger.WithFields(map[string]interface{}{
		"path": filePath,
	}).Debug("Importing checkpoint")

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open import file")
	}
	defer func() {
		_ = file.Close()
	}()

	// Unmarshal from JSON
	var info CheckpointInfo
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&info); err != nil {
		return nil, errors.Wrap(err, "failed to parse checkpoint from JSON")
	}

	// Convert to checkpoint and save
	cp := s.convertInfoToCheckpoint(info)
	if err := s.store.SaveCheckpoint(cp); err != nil {
		return nil, errors.Wrap(err, "failed to save imported checkpoint")
	}

	s.logger.WithFields(map[string]interface{}{
		"id":   info.ID,
		"path": filePath,
	}).Info("Checkpoint imported")

	return &info, nil
}

// convertCheckpointToInfo converts a checkpoint to checkpoint info
func (s *CheckpointService) convertCheckpointToInfo(cp *checkpoint.TreeCheckpoint) CheckpointInfo {
	// Create repository info
	var repositories []RepositoryInfo
	if cp.Repositories != nil {
		repositories = make([]RepositoryInfo, 0, len(cp.Repositories))
		for name, repo := range cp.Repositories {
			repositories = append(repositories, RepositoryInfo{
				Name:        name,
				Status:      string(repo.Status),
				TagsCopied:  0, // Not available in TreeCheckpoint
				TagsSkipped: 0, // Not available in TreeCheckpoint
				Errors:      0, // Not available in TreeCheckpoint
			})
		}
	}

	// Calculate completed and failed repositories
	completedRepos := len(cp.CompletedRepositories)
	failedRepos := 0
	for _, repo := range cp.Repositories {
		if repo.Status == checkpoint.StatusFailed {
			failedRepos++
		}
	}

	// Create checkpoint info
	return CheckpointInfo{
		ID:                    cp.ID,
		CreatedAt:             cp.StartTime,
		Source:                cp.SourceRegistry + "/" + cp.SourcePrefix,
		Destination:           cp.DestRegistry + "/" + cp.DestPrefix,
		Status:                string(cp.Status),
		TotalRepositories:     len(cp.Repositories),
		CompletedRepositories: completedRepos,
		FailedRepositories:    failedRepos,
		TotalTagsCopied:       0, // Not available in TreeCheckpoint
		TotalTagsSkipped:      0, // Not available in TreeCheckpoint
		TotalErrors:           0, // Not available in TreeCheckpoint
		TotalBytesTransferred: 0, // Not available in TreeCheckpoint
		Repositories:          repositories,
	}
}

// convertInfoToCheckpoint converts checkpoint info to a checkpoint
func (s *CheckpointService) convertInfoToCheckpoint(info CheckpointInfo) *checkpoint.TreeCheckpoint {
	// Create repository records
	repositories := make(map[string]checkpoint.RepoStatus, len(info.Repositories))
	for _, repo := range info.Repositories {
		repositories[repo.Name] = checkpoint.RepoStatus{
			Status:      checkpoint.Status(repo.Status),
			SourceRepo:  repo.Name,
			DestRepo:    repo.Name, // We don't have separate source/dest in the info
			LastUpdated: time.Now(),
		}
	}

	// Create completed repositories list
	completedRepos := make([]string, 0, info.CompletedRepositories)
	for _, repo := range info.Repositories {
		if repo.Status == string(checkpoint.StatusCompleted) {
			completedRepos = append(completedRepos, repo.Name)
		}
	}

	// Split source and destination into registry and prefix
	sourceRegistry, sourcePrefix := splitPath(info.Source)
	destRegistry, destPrefix := splitPath(info.Destination)

	// Create checkpoint
	return &checkpoint.TreeCheckpoint{
		ID:                    info.ID,
		StartTime:             info.CreatedAt,
		LastUpdated:           time.Now(),
		SourceRegistry:        sourceRegistry,
		SourcePrefix:          sourcePrefix,
		DestRegistry:          destRegistry,
		DestPrefix:            destPrefix,
		Status:                checkpoint.Status(info.Status),
		Repositories:          repositories,
		CompletedRepositories: completedRepos,
		Progress:              float64(info.CompletedRepositories) / float64(info.TotalRepositories) * 100.0,
	}
}

// splitPath splits a path into registry and prefix
func splitPath(path string) (registry, prefix string) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return path, ""
}

// VerifyCheckpoint verifies that a checkpoint exists and is valid
func (s *CheckpointService) VerifyCheckpoint(ctx context.Context, id string) (bool, error) {
	if err := s.initStore(ctx); err != nil {
		return false, err
	}

	if id == "" {
		return false, errors.InvalidInputf("checkpoint ID is required")
	}

	s.logger.WithFields(map[string]interface{}{
		"id": id,
	}).Debug("Verifying checkpoint")

	// Check if checkpoint exists
	exists, err := s.store.CheckpointExists(id)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if checkpoint exists")
	}

	if !exists {
		return false, nil
	}

	// Load checkpoint to verify it's valid
	_, err = s.store.LoadCheckpoint(id)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		}).Warn("Checkpoint exists but is invalid")
		return false, errors.Wrap(err, "checkpoint exists but is invalid")
	}

	return true, nil
}

// GetRemainingRepositories returns the repositories that still need to be processed
func (s *CheckpointService) GetRemainingRepositories(ctx context.Context, id string, skipCompleted, retryFailed bool) ([]string, error) {
	if err := s.initStore(ctx); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, errors.InvalidInputf("checkpoint ID is required")
	}

	s.logger.WithFields(map[string]interface{}{
		"id":             id,
		"skip_completed": skipCompleted,
		"retry_failed":   retryFailed,
	}).Debug("Getting remaining repositories")

	// Load checkpoint
	cp, err := s.store.LoadCheckpoint(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load checkpoint")
	}

	// Create list of remaining repos
	var remaining []string

	// Add repositories that aren't completed
	for repoName, repoStatus := range cp.Repositories {
		// Skip completed repositories if requested
		if skipCompleted && repoStatus.Status == checkpoint.StatusCompleted {
			continue
		}

		// Skip failed repositories unless retry is enabled
		if !retryFailed && repoStatus.Status == checkpoint.StatusFailed {
			continue
		}

		remaining = append(remaining, repoName)
	}

	s.logger.WithFields(map[string]interface{}{
		"id":        id,
		"count":     len(remaining),
		"total":     len(cp.Repositories),
		"completed": len(cp.CompletedRepositories),
	}).Info("Found remaining repositories")

	return remaining, nil
}
