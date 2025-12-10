package server

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"freightliner/pkg/service"

	"github.com/google/uuid"
)

// JobType defines the type of a replication job
type JobType string

const (
	// JobTypeReplicate is a single repository replication job
	JobTypeReplicate JobType = "replicate"

	// JobTypeReplicateTree is a tree replication job
	JobTypeReplicateTree JobType = "replicate-tree"

	// JobTypeCheckpoint is a checkpoint operation job
	JobTypeCheckpoint JobType = "checkpoint"
)

// JobStatus defines the status of a job
type JobStatus string

const (
	// JobStatusPending indicates a job is pending execution
	JobStatusPending JobStatus = "pending"

	// JobStatusRunning indicates a job is currently running
	JobStatusRunning JobStatus = "running"

	// JobStatusCompleted indicates a job has completed successfully
	JobStatusCompleted JobStatus = "completed"

	// JobStatusFailed indicates a job has failed
	JobStatusFailed JobStatus = "failed"

	// JobStatusCanceled indicates a job was canceled
	JobStatusCanceled JobStatus = "canceled"

	// JobStatusCancelled indicates a job was cancelled (British spelling)
	JobStatusCancelled JobStatus = "cancelled"
)

// JobManager manages job execution and tracking
type JobManager struct {
	jobs      map[string]Job
	jobsMutex sync.RWMutex
}

// NewJobManager creates a new job manager
func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]Job),
	}
}

// AddJob adds a job to the manager
func (m *JobManager) AddJob(job Job) {
	m.jobsMutex.Lock()
	defer m.jobsMutex.Unlock()

	m.jobs[job.GetID()] = job
}

// GetJob returns a job by ID
func (m *JobManager) GetJob(id string) (Job, bool) {
	m.jobsMutex.RLock()
	defer m.jobsMutex.RUnlock()

	job, exists := m.jobs[id]
	return job, exists
}

// ListJobs returns all jobs, optionally filtered by type and status
func (m *JobManager) ListJobs(jobType JobType, status JobStatus) []Job {
	m.jobsMutex.RLock()
	defer m.jobsMutex.RUnlock()

	var result []Job
	for _, job := range m.jobs {
		// Filter by type if specified
		if jobType != "" && job.GetType() != jobType {
			continue
		}

		// Filter by status if specified
		if status != "" && job.GetStatus() != status {
			continue
		}

		result = append(result, job)
	}

	return result
}

// GetJobCount returns the total number of jobs
func (m *JobManager) GetJobCount() int {
	m.jobsMutex.RLock()
	defer m.jobsMutex.RUnlock()

	return len(m.jobs)
}

// UpdateJob updates a job's status and result
func (m *JobManager) UpdateJob(id string, status JobStatus, result interface{}, err error) {
	m.jobsMutex.Lock()
	defer m.jobsMutex.Unlock()

	job, exists := m.jobs[id]
	if !exists {
		return
	}

	// Update job status and result
	job.SetStatus(status)
	if result != nil {
		job.SetResult(result)
	}
	if err != nil {
		job.SetError(err)
	}

	// If the job is completed or failed, set the end time
	if status == JobStatusCompleted || status == JobStatusFailed || status == JobStatusCanceled {
		job.SetEndTime(time.Now())
	}
}

// Job represents a replication job
type Job interface {
	// GetID returns the job ID
	GetID() string

	// GetType returns the job type
	GetType() JobType

	// GetStatus returns the job status
	GetStatus() JobStatus

	// GetSource returns the source for the job
	GetSource() string

	// GetDestination returns the destination for the job
	GetDestination() string

	// GetStartTime returns when the job started
	GetStartTime() time.Time

	// GetEndTime returns when the job ended
	GetEndTime() time.Time

	// GetResult returns the job result
	GetResult() interface{}

	// GetError returns the job error
	GetError() error

	// Execute executes the job
	Execute(ctx context.Context) error

	// ToJSON returns the job as JSON
	ToJSON() ([]byte, error)

	// SetStatus sets the job status
	SetStatus(status JobStatus)

	// SetResult sets the job result
	SetResult(result interface{})

	// SetError sets the job error
	SetError(err error)

	// SetEndTime sets when the job ended
	SetEndTime(time time.Time)
}

// BaseJob provides common functionality for all jobs
type BaseJob struct {
	ID          string      `json:"id"`
	Type        JobType     `json:"type"`
	Source      string      `json:"source"`
	Destination string      `json:"destination"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time,omitempty"`
	Status      JobStatus   `json:"status"`
	ErrorMsg    string      `json:"error,omitempty"`
	ResultData  interface{} `json:"result,omitempty"`

	// Internal fields not serialized to JSON
	error error `json:"-"`
}

// NewBaseJob creates a base job
func NewBaseJob(jobType JobType, source, destination string) *BaseJob {
	return &BaseJob{
		ID:          uuid.New().String(),
		Type:        jobType,
		Source:      source,
		Destination: destination,
		StartTime:   time.Now(),
		Status:      JobStatusPending,
	}
}

// GetID returns the job ID
func (j *BaseJob) GetID() string {
	return j.ID
}

// GetType returns the job type
func (j *BaseJob) GetType() JobType {
	return j.Type
}

// GetStatus returns the job status
func (j *BaseJob) GetStatus() JobStatus {
	return j.Status
}

// GetSource returns the source for the job
func (j *BaseJob) GetSource() string {
	return j.Source
}

// GetDestination returns the destination for the job
func (j *BaseJob) GetDestination() string {
	return j.Destination
}

// GetStartTime returns when the job started
func (j *BaseJob) GetStartTime() time.Time {
	return j.StartTime
}

// GetEndTime returns when the job ended
func (j *BaseJob) GetEndTime() time.Time {
	return j.EndTime
}

// GetResult returns the job result
func (j *BaseJob) GetResult() interface{} {
	return j.ResultData
}

// GetError returns the job error
func (j *BaseJob) GetError() error {
	return j.error
}

// SetStatus sets the job status
func (j *BaseJob) SetStatus(status JobStatus) {
	j.Status = status
}

// SetResult sets the job result
func (j *BaseJob) SetResult(result interface{}) {
	j.ResultData = result
}

// SetError sets the job error
func (j *BaseJob) SetError(err error) {
	j.error = err
	if err != nil {
		j.ErrorMsg = err.Error()
	} else {
		j.ErrorMsg = ""
	}
}

// SetEndTime sets when the job ended
func (j *BaseJob) SetEndTime(time time.Time) {
	j.EndTime = time
}

// ToJSON returns the job as JSON
func (j *BaseJob) ToJSON() ([]byte, error) {
	return json.Marshal(j)
}

// ReplicateJob represents a single repository replication job
type ReplicateJob struct {
	*BaseJob
	Tags   []string `json:"tags,omitempty"`
	Force  bool     `json:"force"`
	DryRun bool     `json:"dry_run"`
	svc    service.ReplicationService
}

// NewReplicateJob creates a new replicate job
func NewReplicateJob(source, destination string, tags []string, force, dryRun bool, svc service.ReplicationService) *ReplicateJob {
	return &ReplicateJob{
		BaseJob: NewBaseJob(JobTypeReplicate, source, destination),
		Tags:    tags,
		Force:   force,
		DryRun:  dryRun,
		svc:     svc,
	}
}

// Execute executes the job
func (j *ReplicateJob) Execute(ctx context.Context) error {
	// Update status to running
	j.Status = JobStatusRunning

	// Execute replication
	result, err := j.svc.ReplicateRepository(ctx, j.Source, j.Destination)

	// Handle result and error
	if err != nil {
		j.Status = JobStatusFailed
		j.SetError(err)
		return err
	}

	// Update result and status
	j.Status = JobStatusCompleted
	j.ResultData = result
	j.EndTime = time.Now()

	return nil
}

// ReplicateTreeJob represents a tree replication job
type ReplicateTreeJob struct {
	*BaseJob
	ExcludeRepos     []string `json:"exclude_repos,omitempty"`
	ExcludeTags      []string `json:"exclude_tags,omitempty"`
	IncludeTags      []string `json:"include_tags,omitempty"`
	Force            bool     `json:"force"`
	DryRun           bool     `json:"dry_run"`
	EnableCheckpoint bool     `json:"enable_checkpoint"`
	CheckpointDir    string   `json:"checkpoint_dir,omitempty"`
	ResumeID         string   `json:"resume_id,omitempty"`
	SkipCompleted    bool     `json:"skip_completed"`
	RetryFailed      bool     `json:"retry_failed"`
	svc              *service.TreeReplicationService
}

// NewReplicateTreeJob creates a new replicate tree job
func NewReplicateTreeJob(source, destination string, options map[string]interface{}, svc *service.TreeReplicationService) *ReplicateTreeJob {
	job := &ReplicateTreeJob{
		BaseJob: NewBaseJob(JobTypeReplicateTree, source, destination),
		svc:     svc,
	}

	// Extract options
	if excludeRepos, ok := options["excludeRepos"].([]string); ok {
		job.ExcludeRepos = excludeRepos
	}

	if excludeTags, ok := options["excludeTags"].([]string); ok {
		job.ExcludeTags = excludeTags
	}

	if includeTags, ok := options["includeTags"].([]string); ok {
		job.IncludeTags = includeTags
	}

	if force, ok := options["force"].(bool); ok {
		job.Force = force
	}

	if dryRun, ok := options["dryRun"].(bool); ok {
		job.DryRun = dryRun
	}

	if enableCheckpoint, ok := options["enableCheckpoint"].(bool); ok {
		job.EnableCheckpoint = enableCheckpoint
	}

	if checkpointDir, ok := options["checkpointDir"].(string); ok {
		job.CheckpointDir = checkpointDir
	}

	if resumeID, ok := options["resumeID"].(string); ok {
		job.ResumeID = resumeID
	}

	if skipCompleted, ok := options["skipCompleted"].(bool); ok {
		job.SkipCompleted = skipCompleted
	}

	if retryFailed, ok := options["retryFailed"].(bool); ok {
		job.RetryFailed = retryFailed
	}

	return job
}

// Execute executes the job
func (j *ReplicateTreeJob) Execute(ctx context.Context) error {
	// Update status to running
	j.Status = JobStatusRunning

	// Execute replication
	result, err := j.svc.ReplicateTree(ctx, j.Source, j.Destination)

	// Handle result and error
	if err != nil {
		j.Status = JobStatusFailed
		j.SetError(err)
		return err
	}

	// Update result and status
	j.Status = JobStatusCompleted
	j.ResultData = result
	j.EndTime = time.Now()

	return nil
}
