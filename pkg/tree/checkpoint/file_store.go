package checkpoint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"freightliner/pkg/helper/errors"
)

// FileStore implements the CheckpointStore interface using the filesystem
type FileStore struct {
	// Directory where checkpoint files are stored
	directory string

	// Mutex for concurrent access
	mu sync.Mutex
}

// NewFileStore creates a new file-based checkpoint store
func NewFileStore(directory string) (*FileStore, error) {
	// Expand HOME directory if present
	if strings.HasPrefix(directory, "${HOME}") || strings.HasPrefix(directory, "$HOME") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get user home directory")
		}
		directory = strings.Replace(directory, "${HOME}", home, 1)
		directory = strings.Replace(directory, "$HOME", home, 1)
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, errors.Wrap(err, "failed to create checkpoint directory")
	}

	return &FileStore{
		directory: directory,
	}, nil
}

// SaveCheckpoint saves a checkpoint to the store
func (s *FileStore) SaveCheckpoint(checkpoint *TreeCheckpoint) error {
	// Validate input before locking to fail fast
	if checkpoint == nil {
		return errors.InvalidInputf("checkpoint cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update the checkpoint timestamp
	checkpoint.LastUpdated = time.Now()

	// Serialize the checkpoint to JSON
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to serialize checkpoint")
	}

	// Write to file
	filename := filepath.Join(s.directory, checkpoint.ID+".json")

	// Create or overwrite the file
	err = os.WriteFile(filename, data, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to write checkpoint file")
	}

	return nil
}

// LoadCheckpoint retrieves a checkpoint from the store
// This is an alias for GetCheckpoint to satisfy the interface
func (s *FileStore) LoadCheckpoint(id string) (*TreeCheckpoint, error) {
	return s.GetCheckpoint(id)
}

// GetCheckpoint retrieves a checkpoint from the store
func (s *FileStore) GetCheckpoint(id string) (*TreeCheckpoint, error) {
	// Validate input before locking to fail fast
	if id == "" {
		return nil, errors.InvalidInputf("checkpoint ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Read the checkpoint file
	filename := filepath.Join(s.directory, id+".json")

	data, err := os.ReadFile(filename) // #nosec G304 - filename is constructed from validated directory and ID
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NotFoundf("checkpoint not found: %s", id)
		}
		return nil, errors.Wrap(err, "failed to read checkpoint file")
	}

	// Deserialize the checkpoint
	var checkpoint TreeCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, errors.Wrap(err, "failed to deserialize checkpoint")
	}

	return &checkpoint, nil
}

// CheckpointExists checks if a checkpoint with the given ID exists
func (s *FileStore) CheckpointExists(id string) (bool, error) {
	// Validate input before locking to fail fast
	if id == "" {
		return false, errors.InvalidInputf("checkpoint ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the checkpoint file exists
	filename := filepath.Join(s.directory, id+".json")
	_, err := os.Stat(filename)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, errors.Wrap(err, "failed to check if checkpoint exists")
}

// ListCheckpoints returns a list of all checkpoints in the store
func (s *FileStore) ListCheckpoints() ([]*TreeCheckpoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// List all JSON files in the directory
	pattern := filepath.Join(s.directory, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list checkpoint files")
	}

	var checkpoints []*TreeCheckpoint

	for _, filename := range matches {
		// Read and deserialize each checkpoint
		data, err := os.ReadFile(filename) // #nosec G304 - filename is constructed from validated directory and ID // #nosec G304 - filename is constructed from validated directory and ID
		if err != nil {
			continue // Skip files that can't be read
		}

		var checkpoint TreeCheckpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			continue // Skip files that can't be deserialized
		}

		checkpoints = append(checkpoints, &checkpoint)
	}

	return checkpoints, nil
}

// DeleteCheckpoint deletes a checkpoint from the store
func (s *FileStore) DeleteCheckpoint(id string) error {
	// Validate input before locking to fail fast
	if id == "" {
		return errors.InvalidInputf("checkpoint ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete the checkpoint file
	filename := filepath.Join(s.directory, id+".json")

	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return errors.NotFoundf("checkpoint not found: %s", id)
		}
		return errors.Wrap(err, "failed to delete checkpoint file")
	}

	return nil
}

// PruneCheckpoints deletes checkpoints older than the given duration
func (s *FileStore) PruneCheckpoints(olderThan time.Duration) (int, error) {
	// Validate input before locking to fail fast
	if olderThan <= 0 {
		return 0, errors.InvalidInputf("duration must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// List all checkpoints
	checkpoints, err := s.listCheckpointsUnlocked()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-olderThan)
	deleted := 0

	// Delete checkpoints older than the cutoff
	for _, checkpoint := range checkpoints {
		if checkpoint.LastUpdated.Before(cutoff) {
			filename := filepath.Join(s.directory, checkpoint.ID+".json")
			if err := os.Remove(filename); err == nil {
				deleted++
			}
		}
	}

	return deleted, nil
}

// listCheckpointsUnlocked is an helper helper that doesn't lock the mutex
func (s *FileStore) listCheckpointsUnlocked() ([]*TreeCheckpoint, error) {
	// List all JSON files in the directory
	pattern := filepath.Join(s.directory, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list checkpoint files")
	}

	var checkpoints []*TreeCheckpoint

	for _, filename := range matches {
		// Read and deserialize each checkpoint
		data, err := os.ReadFile(filename) // #nosec G304 - filename is constructed from validated directory and ID // #nosec G304 - filename is constructed from validated directory and ID
		if err != nil {
			continue // Skip files that can't be read
		}

		var checkpoint TreeCheckpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			continue // Skip files that can't be deserialized
		}

		checkpoints = append(checkpoints, &checkpoint)
	}

	return checkpoints, nil
}

// GetDirectory returns the directory where checkpoints are stored
func (s *FileStore) GetDirectory() string {
	return s.directory
}
