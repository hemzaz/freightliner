package tree

import (
	"fmt"
	
	"github.com/hemzaz/freightliner/src/pkg/tree/checkpoint"
)

// InitCheckpointStore initializes a checkpoint store
func InitCheckpointStore(dir string) (checkpoint.CheckpointStore, error) {
	return checkpoint.NewFileStore(dir)
}

// ListResumableCheckpoints returns a list of resumable checkpoints
func ListResumableCheckpoints(store checkpoint.CheckpointStore) ([]checkpoint.ResumableCheckpoint, error) {
	return checkpoint.GetResumableCheckpoints(store)
}