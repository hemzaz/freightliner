package distributed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/replication"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

// RaftCoordinator manages distributed consensus for job coordination
type RaftCoordinator struct {
	raft          *raft.Raft
	fsm           *ReplicationFSM
	peers         []string
	logger        log.Logger
	dataDir       string
	bindAddr      string
	mu            sync.RWMutex
	shutdownCh    chan struct{}
	leaderNotifCh chan bool
}

// ReplicationFSM is the finite state machine for Raft
type ReplicationFSM struct {
	jobs        map[string]*JobState
	checkpoints map[string]*CheckpointState
	mu          sync.RWMutex
	logger      log.Logger
}

// JobState represents the state of a replication job
type JobState struct {
	ID         string                      `json:"id"`
	Rule       replication.ReplicationRule `json:"rule"`
	Status     string                      `json:"status"`
	StartTime  time.Time                   `json:"start_time"`
	UpdateTime time.Time                   `json:"update_time"`
	NodeID     string                      `json:"node_id"`
	RetryCount int                         `json:"retry_count"`
}

// CheckpointState represents checkpoint state
type CheckpointState struct {
	JobID          string            `json:"job_id"`
	CompletedTags  []string          `json:"completed_tags"`
	FailedTags     map[string]string `json:"failed_tags"`
	LastUpdateTime time.Time         `json:"last_update_time"`
}

// Command represents a state change command
type Command struct {
	Type       string           `json:"type"`
	JobID      string           `json:"job_id"`
	Job        *JobState        `json:"job,omitempty"`
	Checkpoint *CheckpointState `json:"checkpoint,omitempty"`
	Data       json.RawMessage  `json:"data,omitempty"`
}

// RaftConfig holds configuration for Raft coordinator
type RaftConfig struct {
	NodeID           string
	BindAddr         string
	DataDir          string
	Peers            []string
	Bootstrap        bool
	Logger           log.Logger
	HeartbeatTimeout time.Duration
	ElectionTimeout  time.Duration
}

// NewRaftCoordinator creates a new Raft coordinator
func NewRaftCoordinator(config RaftConfig) (*RaftCoordinator, error) {
	if config.Logger == nil {
		config.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Create data directory
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create data directory")
	}

	// Create FSM
	fsm := &ReplicationFSM{
		jobs:        make(map[string]*JobState),
		checkpoints: make(map[string]*CheckpointState),
		logger:      config.Logger,
	}

	coordinator := &RaftCoordinator{
		fsm:           fsm,
		peers:         config.Peers,
		logger:        config.Logger,
		dataDir:       config.DataDir,
		bindAddr:      config.BindAddr,
		shutdownCh:    make(chan struct{}),
		leaderNotifCh: make(chan bool, 1),
	}

	// Configure Raft
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(config.NodeID)

	// Set timeouts
	if config.HeartbeatTimeout > 0 {
		raftConfig.HeartbeatTimeout = config.HeartbeatTimeout
	} else {
		raftConfig.HeartbeatTimeout = 1 * time.Second
	}

	if config.ElectionTimeout > 0 {
		raftConfig.ElectionTimeout = config.ElectionTimeout
	} else {
		raftConfig.ElectionTimeout = 3 * time.Second
	}

	raftConfig.LeaderLeaseTimeout = 500 * time.Millisecond
	raftConfig.CommitTimeout = 500 * time.Millisecond
	raftConfig.LogLevel = "WARN"
	raftConfig.NotifyCh = coordinator.leaderNotifCh

	// Create transport
	addr, err := net.ResolveTCPAddr("tcp", config.BindAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve bind address")
	}

	transport, err := raft.NewTCPTransport(config.BindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transport")
	}

	// Create snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(config.DataDir, 2, os.Stderr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create snapshot store")
	}

	// Create log store
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(config.DataDir, "raft-log.db"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create log store")
	}

	// Create stable store
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(config.DataDir, "raft-stable.db"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create stable store")
	}

	// Create Raft instance
	ra, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create raft")
	}

	coordinator.raft = ra

	// Bootstrap cluster if needed
	if config.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(config.NodeID),
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	// Start leadership monitoring
	go coordinator.monitorLeadership()

	return coordinator, nil
}

// Apply implements the FSM interface
func (f *ReplicationFSM) Apply(l *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		f.logger.Error("Failed to unmarshal command", err)
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Type {
	case "create_job":
		if cmd.Job != nil {
			f.jobs[cmd.JobID] = cmd.Job
			f.logger.WithFields(map[string]interface{}{
				"job_id": cmd.JobID,
				"status": cmd.Job.Status,
			}).Info("Job created in FSM")
		}

	case "update_job":
		if cmd.Job != nil {
			f.jobs[cmd.JobID] = cmd.Job
			f.logger.WithFields(map[string]interface{}{
				"job_id": cmd.JobID,
				"status": cmd.Job.Status,
			}).Debug("Job updated in FSM")
		}

	case "complete_job":
		delete(f.jobs, cmd.JobID)
		f.logger.WithFields(map[string]interface{}{
			"job_id": cmd.JobID,
		}).Info("Job completed and removed from FSM")

	case "update_checkpoint":
		if cmd.Checkpoint != nil {
			f.checkpoints[cmd.JobID] = cmd.Checkpoint
			f.logger.WithFields(map[string]interface{}{
				"job_id": cmd.JobID,
				"tags":   len(cmd.Checkpoint.CompletedTags),
			}).Debug("Checkpoint updated in FSM")
		}

	case "delete_checkpoint":
		delete(f.checkpoints, cmd.JobID)
		f.logger.WithFields(map[string]interface{}{
			"job_id": cmd.JobID,
		}).Debug("Checkpoint deleted from FSM")

	default:
		f.logger.WithFields(map[string]interface{}{
			"type": cmd.Type,
		}).Warn("Unknown command type")
	}

	return nil
}

// Snapshot implements the FSM interface
func (f *ReplicationFSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Clone the state
	jobs := make(map[string]*JobState, len(f.jobs))
	for k, v := range f.jobs {
		jobCopy := *v
		jobs[k] = &jobCopy
	}

	checkpoints := make(map[string]*CheckpointState, len(f.checkpoints))
	for k, v := range f.checkpoints {
		cpCopy := *v
		checkpoints[k] = &cpCopy
	}

	return &fsmSnapshot{
		jobs:        jobs,
		checkpoints: checkpoints,
	}, nil
}

// Restore implements the FSM interface
func (f *ReplicationFSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var state struct {
		Jobs        map[string]*JobState        `json:"jobs"`
		Checkpoints map[string]*CheckpointState `json:"checkpoints"`
	}

	if err := json.NewDecoder(rc).Decode(&state); err != nil {
		return errors.Wrap(err, "failed to decode snapshot")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.jobs = state.Jobs
	f.checkpoints = state.Checkpoints

	f.logger.WithFields(map[string]interface{}{
		"jobs":        len(f.jobs),
		"checkpoints": len(f.checkpoints),
	}).Info("FSM state restored from snapshot")

	return nil
}

// fsmSnapshot implements the FSMSnapshot interface
type fsmSnapshot struct {
	jobs        map[string]*JobState
	checkpoints map[string]*CheckpointState
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		state := struct {
			Jobs        map[string]*JobState        `json:"jobs"`
			Checkpoints map[string]*CheckpointState `json:"checkpoints"`
		}{
			Jobs:        s.jobs,
			Checkpoints: s.checkpoints,
		}

		b, err := json.Marshal(state)
		if err != nil {
			return err
		}

		if _, err := sink.Write(b); err != nil {
			return err
		}

		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

func (s *fsmSnapshot) Release() {}

// CreateJob creates a new job in the cluster
func (rc *RaftCoordinator) CreateJob(ctx context.Context, job *JobState) error {
	if rc.raft.State() != raft.Leader {
		return errors.New("not the leader")
	}

	cmd := Command{
		Type:  "create_job",
		JobID: job.ID,
		Job:   job,
	}

	return rc.applyCommand(ctx, cmd)
}

// UpdateJob updates an existing job
func (rc *RaftCoordinator) UpdateJob(ctx context.Context, job *JobState) error {
	if rc.raft.State() != raft.Leader {
		return errors.New("not the leader")
	}

	cmd := Command{
		Type:  "update_job",
		JobID: job.ID,
		Job:   job,
	}

	return rc.applyCommand(ctx, cmd)
}

// CompleteJob marks a job as complete
func (rc *RaftCoordinator) CompleteJob(ctx context.Context, jobID string) error {
	if rc.raft.State() != raft.Leader {
		return errors.New("not the leader")
	}

	cmd := Command{
		Type:  "complete_job",
		JobID: jobID,
	}

	return rc.applyCommand(ctx, cmd)
}

// UpdateCheckpoint updates checkpoint state
func (rc *RaftCoordinator) UpdateCheckpoint(ctx context.Context, checkpoint *CheckpointState) error {
	if rc.raft.State() != raft.Leader {
		return errors.New("not the leader")
	}

	cmd := Command{
		Type:       "update_checkpoint",
		JobID:      checkpoint.JobID,
		Checkpoint: checkpoint,
	}

	return rc.applyCommand(ctx, cmd)
}

// GetJob retrieves a job from the FSM
func (rc *RaftCoordinator) GetJob(jobID string) (*JobState, bool) {
	rc.fsm.mu.RLock()
	defer rc.fsm.mu.RUnlock()

	job, exists := rc.fsm.jobs[jobID]
	return job, exists
}

// ListJobs returns all active jobs
func (rc *RaftCoordinator) ListJobs() []*JobState {
	rc.fsm.mu.RLock()
	defer rc.fsm.mu.RUnlock()

	jobs := make([]*JobState, 0, len(rc.fsm.jobs))
	for _, job := range rc.fsm.jobs {
		jobCopy := *job
		jobs = append(jobs, &jobCopy)
	}

	return jobs
}

// GetCheckpoint retrieves checkpoint state
func (rc *RaftCoordinator) GetCheckpoint(jobID string) (*CheckpointState, bool) {
	rc.fsm.mu.RLock()
	defer rc.fsm.mu.RUnlock()

	cp, exists := rc.fsm.checkpoints[jobID]
	return cp, exists
}

// IsLeader returns true if this node is the leader
func (rc *RaftCoordinator) IsLeader() bool {
	return rc.raft.State() == raft.Leader
}

// GetLeader returns the current leader address
func (rc *RaftCoordinator) GetLeader() string {
	return string(rc.raft.Leader())
}

// AddVoter adds a new voting node to the cluster
func (rc *RaftCoordinator) AddVoter(nodeID, address string, timeout time.Duration) error {
	if rc.raft.State() != raft.Leader {
		return errors.New("not the leader")
	}

	future := rc.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(address), 0, timeout)
	return future.Error()
}

// RemoveServer removes a node from the cluster
func (rc *RaftCoordinator) RemoveServer(nodeID string, timeout time.Duration) error {
	if rc.raft.State() != raft.Leader {
		return errors.New("not the leader")
	}

	future := rc.raft.RemoveServer(raft.ServerID(nodeID), 0, timeout)
	return future.Error()
}

// Stats returns Raft statistics
func (rc *RaftCoordinator) Stats() map[string]string {
	return rc.raft.Stats()
}

// Shutdown gracefully shuts down the coordinator
func (rc *RaftCoordinator) Shutdown() error {
	close(rc.shutdownCh)
	return rc.raft.Shutdown().Error()
}

// applyCommand applies a command to the Raft log
func (rc *RaftCoordinator) applyCommand(ctx context.Context, cmd Command) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return errors.Wrap(err, "failed to marshal command")
	}

	future := rc.raft.Apply(data, 10*time.Second)
	if err := future.Error(); err != nil {
		return errors.Wrap(err, "failed to apply command")
	}

	return nil
}

// monitorLeadership monitors leadership changes
func (rc *RaftCoordinator) monitorLeadership() {
	for {
		select {
		case isLeader := <-rc.leaderNotifCh:
			if isLeader {
				rc.logger.Info("This node became the leader")
			} else {
				rc.logger.Info("This node lost leadership")
			}
		case <-rc.shutdownCh:
			return
		}
	}
}

// WaitForLeader waits for a leader to be elected
func (rc *RaftCoordinator) WaitForLeader(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for leader")
		case <-ticker.C:
			if rc.GetLeader() != "" {
				return nil
			}
		}
	}
}
