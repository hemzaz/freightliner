package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
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
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	return &FileStore{
		directory: directory,
	}, nil
}

// SaveCheckpoint saves a checkpoint to a file
func (f *FileStore) SaveCheckpoint(checkpoint *TreeCheckpoint) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Update the last updated time
	checkpoint.LastUpdated = time.Now()

	// Marshal the checkpoint to JSON
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	// Create the checkpoint file path
	filePath := filepath.Join(f.directory, fmt.Sprintf("%s.json", checkpoint.ID))

	// Write the checkpoint to a temporary file first
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary checkpoint file: %w", err)
	}

	// Rename the temporary file to the actual file for atomic updates
	if err := os.Rename(tempPath, filePath); err != nil {
		return fmt.Errorf("failed to rename checkpoint file: %w", err)
	}

	return nil
}

// LoadCheckpoint loads a checkpoint from a file
func (f *FileStore) LoadCheckpoint(id string) (*TreeCheckpoint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	filePath := filepath.Join(f.directory, fmt.Sprintf("%s.json", id))

	// Read the checkpoint file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	// Unmarshal the checkpoint from JSON
	var checkpoint TreeCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// ListCheckpoints lists all checkpoints in the store
func (f *FileStore) ListCheckpoints() ([]*TreeCheckpoint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// List all JSON files in the checkpoint directory
	pattern := filepath.Join(f.directory, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoint files: %w", err)
	}

	var checkpoints []*TreeCheckpoint

	// Load each checkpoint file
	for _, match := range matches {
		// Skip temporary files
		if filepath.Ext(match) == ".tmp" {
			continue
		}

		// Read the checkpoint file
		data, err := os.ReadFile(match)
		if err != nil {
			continue // Skip files that can't be read
		}

		// Unmarshal the checkpoint from JSON
		var checkpoint TreeCheckpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			continue // Skip files that can't be unmarshaled
		}

		checkpoints = append(checkpoints, &checkpoint)
	}

	return checkpoints, nil
}

// DeleteCheckpoint deletes a checkpoint file
func (f *FileStore) DeleteCheckpoint(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	filePath := filepath.Join(f.directory, fmt.Sprintf("%s.json", id))

	// Delete the checkpoint file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete checkpoint file: %w", err)
	}

	return nil
}
