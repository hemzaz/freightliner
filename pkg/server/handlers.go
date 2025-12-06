package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// replicateHandler handles repository replication requests
func (s *Server) replicateHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req ReplicateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %s", err))
		return
	}

	// Validate request
	if err := s.validateReplicateRequest(&req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create source and destination paths
	source := fmt.Sprintf("%s/%s", req.SourceRegistry, req.SourceRepo)
	destination := fmt.Sprintf("%s/%s", req.DestRegistry, req.DestRepo)

	// Create replication job
	job := NewReplicateJob(source, destination, req.Tags, req.Force, req.DryRun, s.replicationSvc)

	// Add job to manager
	s.jobManager.AddJob(job)

	// Submit job to worker pool
	err := s.workerPool.Submit(job.GetID(), func(ctx context.Context) error {
		// Update job status
		job.SetStatus(JobStatusRunning)

		// Execute job
		err := job.Execute(ctx)

		// Job status and result are already updated by the Execute method
		return err
	})

	if err != nil {
		// Update job status if submission failed
		job.SetStatus(JobStatusFailed)
		job.SetError(fmt.Errorf("failed to submit job: %w", err))

		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to submit job")
		return
	}

	// Return job reference
	s.writeResponse(w, http.StatusAccepted, map[string]string{
		"job_id": job.GetID(),
		"status": string(job.GetStatus()),
	})
}

// replicateTreeHandler handles tree replication requests
func (s *Server) replicateTreeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req ReplicateTreeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %s", err))
		return
	}

	// Validate request
	if err := s.validateReplicateTreeRequest(&req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create source and destination paths
	source := fmt.Sprintf("%s/%s", req.SourceRegistry, req.SourceRepo)
	destination := fmt.Sprintf("%s/%s", req.DestRegistry, req.DestRepo)

	// Create options map
	options := map[string]interface{}{
		"excludeRepos":     req.ExcludeRepos,
		"excludeTags":      req.ExcludeTags,
		"includeTags":      req.IncludeTags,
		"force":            req.Force,
		"dryRun":           req.DryRun,
		"enableCheckpoint": req.EnableCheckpoint,
		"checkpointDir":    req.CheckpointDir,
		"skipCompleted":    true, // Default value
		"retryFailed":      true, // Default value
	}

	// Create replication job
	job := NewReplicateTreeJob(source, destination, options, s.treeReplicationSvc)

	// Add job to manager
	s.jobManager.AddJob(job)

	// Submit job to worker pool
	err := s.workerPool.Submit(job.GetID(), func(ctx context.Context) error {
		// Update job status
		job.SetStatus(JobStatusRunning)

		// Execute job
		err := job.Execute(ctx)

		// Job status and result are already updated by the Execute method
		return err
	})

	if err != nil {
		// Update job status if submission failed
		job.SetStatus(JobStatusFailed)
		job.SetError(fmt.Errorf("failed to submit job: %w", err))

		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to submit job")
		return
	}

	// Return job reference
	s.writeResponse(w, http.StatusAccepted, map[string]string{
		"job_id": job.GetID(),
		"status": string(job.GetStatus()),
	})
}

// listJobsHandler handles listing jobs
func (s *Server) listJobsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryValues := r.URL.Query()

	// Parse job type filter
	var jobType JobType
	if typeStr := queryValues.Get("type"); typeStr != "" {
		jobType = JobType(typeStr)
	}

	// Parse job status filter
	var jobStatus JobStatus
	if statusStr := queryValues.Get("status"); statusStr != "" {
		jobStatus = JobStatus(statusStr)
	}

	// Get jobs
	jobs := s.jobManager.ListJobs(jobType, jobStatus)

	// Convert jobs to JSON-friendly format
	result := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		// Convert job to JSON and parse it back to a map
		jsonData, err := job.ToJSON()
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"job_id": job.GetID(),
				"error":  err.Error(),
			}).Error("Failed to convert job to JSON", err)
			continue
		}

		var jobMap map[string]interface{}
		if err := json.Unmarshal(jsonData, &jobMap); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"job_id": job.GetID(),
				"error":  err.Error(),
			}).Error("Failed to parse job JSON", err)
			continue
		}

		result[i] = jobMap
	}

	// Return jobs
	s.writeResponse(w, http.StatusOK, map[string]interface{}{
		"jobs":  result,
		"count": len(result),
	})
}

// getJobHandler handles getting job details
func (s *Server) getJobHandler(w http.ResponseWriter, r *http.Request) {
	// Get job ID from path
	vars := mux.Vars(r)
	jobID := vars["id"]

	// Get job
	job, exists := s.jobManager.GetJob(jobID)
	if !exists {
		s.writeErrorResponse(w, http.StatusNotFound, "Job not found")
		return
	}

	// Convert job to JSON
	jsonData, err := job.ToJSON()
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"job_id": job.GetID(),
			"error":  err.Error(),
		}).Error("Failed to convert job to JSON", err)
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to convert job to JSON")
		return
	}

	// Parse job JSON to map
	var jobMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jobMap); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"job_id": job.GetID(),
			"error":  err.Error(),
		}).Error("Failed to parse job JSON", err)
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to parse job JSON")
		return
	}

	// Return job
	s.writeResponse(w, http.StatusOK, jobMap)
}

// listCheckpointsHandler handles listing checkpoints
func (s *Server) listCheckpointsHandler(w http.ResponseWriter, r *http.Request) {
	// Get checkpoints
	checkpoints, err := s.checkpointSvc.ListCheckpoints(r.Context())
	if err != nil {
		s.logger.Error("Failed to list checkpoints", err)
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list checkpoints")
		return
	}

	// Return checkpoints
	s.writeResponse(w, http.StatusOK, map[string]interface{}{
		"checkpoints": checkpoints,
		"count":       len(checkpoints),
	})
}

// getCheckpointHandler handles getting checkpoint details
func (s *Server) getCheckpointHandler(w http.ResponseWriter, r *http.Request) {
	// Get checkpoint ID from path
	vars := mux.Vars(r)
	checkpointID := vars["id"]

	// Get checkpoint
	checkpoint, err := s.checkpointSvc.GetCheckpoint(r.Context(), checkpointID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeErrorResponse(w, http.StatusNotFound, "Checkpoint not found")
		} else {
			s.logger.WithFields(map[string]interface{}{
				"checkpoint_id": checkpointID,
				"error":         err.Error(),
			}).Error("Failed to get checkpoint", err)
			s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get checkpoint")
		}
		return
	}

	// Return checkpoint
	s.writeResponse(w, http.StatusOK, checkpoint)
}

// deleteCheckpointHandler handles deleting a checkpoint
func (s *Server) deleteCheckpointHandler(w http.ResponseWriter, r *http.Request) {
	// Get checkpoint ID from path
	vars := mux.Vars(r)
	checkpointID := vars["id"]

	// Delete checkpoint
	err := s.checkpointSvc.DeleteCheckpoint(r.Context(), checkpointID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeErrorResponse(w, http.StatusNotFound, "Checkpoint not found")
		} else {
			s.logger.WithFields(map[string]interface{}{
				"checkpoint_id": checkpointID,
				"error":         err.Error(),
			}).Error("Failed to delete checkpoint", err)
			s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete checkpoint")
		}
		return
	}

	// Return success
	s.writeResponse(w, http.StatusOK, map[string]string{
		"id":      checkpointID,
		"status":  "deleted",
		"message": "Checkpoint deleted successfully",
	})
}
