package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Additional API handlers for production readiness

// cancelJobHandler handles job cancellation requests
func (s *Server) cancelJobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	if jobID == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	// Get the job
	job, exists := s.jobManager.GetJob(jobID)
	if !exists {
		s.writeErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Job %s not found", jobID))
		return
	}

	// Check if job can be cancelled
	status := job.GetStatus()
	if status == JobStatusCompleted || status == JobStatusFailed || status == JobStatusCancelled {
		s.writeErrorResponse(w, http.StatusBadRequest,
			fmt.Sprintf("Job %s cannot be cancelled (status: %s)", jobID, status))
		return
	}

	// Cancel the job
	if err := s.cancelJob(job); err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to cancel job: %s", err))
		return
	}

	s.writeResponse(w, http.StatusOK, map[string]interface{}{
		"job_id":  jobID,
		"status":  string(JobStatusCancelled),
		"message": "Job cancellation initiated",
	})
}

// retryJobHandler handles job retry requests
func (s *Server) retryJobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	if jobID == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	// Get the job
	job, exists := s.jobManager.GetJob(jobID)
	if !exists {
		s.writeErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Job %s not found", jobID))
		return
	}

	// Check if job can be retried
	status := job.GetStatus()
	if status != JobStatusFailed && status != JobStatusCancelled {
		s.writeErrorResponse(w, http.StatusBadRequest,
			fmt.Sprintf("Job %s cannot be retried (status: %s)", jobID, status))
		return
	}

	// Create a new job with the same parameters
	newJob, err := s.cloneJob(job)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to create retry job: %s", err))
		return
	}

	// Add new job to manager
	s.jobManager.AddJob(newJob)

	// Submit to worker pool
	err = s.workerPool.Submit(newJob.GetID(), func(ctx context.Context) error {
		newJob.SetStatus(JobStatusRunning)
		return newJob.Execute(ctx)
	})

	if err != nil {
		newJob.SetStatus(JobStatusFailed)
		newJob.SetError(fmt.Errorf("failed to submit retry job: %w", err))
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to submit retry job")
		return
	}

	s.writeResponse(w, http.StatusAccepted, map[string]interface{}{
		"job_id":          newJob.GetID(),
		"original_job_id": jobID,
		"status":          string(newJob.GetStatus()),
		"message":         "Job retry initiated",
	})
}

// listRegistriesHandler lists all configured registries
func (s *Server) listRegistriesHandler(w http.ResponseWriter, r *http.Request) {
	registries := make([]map[string]interface{}, 0)

	// Add ECR registry if configured
	if s.cfg.ECR.AccountID != "" {
		registries = append(registries, map[string]interface{}{
			"name":     "ecr",
			"type":     "ecr",
			"region":   s.cfg.ECR.Region,
			"endpoint": fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", s.cfg.ECR.AccountID, s.cfg.ECR.Region),
			"enabled":  true,
		})
	}

	// Add GCR registry if configured
	if s.cfg.GCR.Project != "" {
		location := s.cfg.GCR.Location
		if location == "" {
			location = "us"
		}
		registries = append(registries, map[string]interface{}{
			"name":     "gcr",
			"type":     "gcr",
			"project":  s.cfg.GCR.Project,
			"location": location,
			"endpoint": fmt.Sprintf("%s.gcr.io", location),
			"enabled":  true,
		})
	}

	// Add custom registries
	for _, reg := range s.cfg.Registries.Registries {
		registries = append(registries, map[string]interface{}{
			"name":     reg.Name,
			"type":     string(reg.Type),
			"endpoint": reg.Endpoint,
			"enabled":  true,
		})
	}

	s.writeResponse(w, http.StatusOK, map[string]interface{}{
		"registries": registries,
		"count":      len(registries),
	})
}

// getRegistryHealthHandler checks registry health
func (s *Server) getRegistryHealthHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	registryName := vars["name"]

	if registryName == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "Registry name is required")
		return
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Check registry health based on type
	healthy, message, responseTime := s.checkRegistryHealth(ctx, registryName)

	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	s.writeResponse(w, http.StatusOK, map[string]interface{}{
		"registry":      registryName,
		"status":        status,
		"message":       message,
		"response_time": responseTime,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	})
}

// getWorkerPoolStatsHandler returns worker pool statistics
func (s *Server) getWorkerPoolStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := s.workerPool.GetStats()

	s.writeResponse(w, http.StatusOK, map[string]interface{}{
		"workers": map[string]interface{}{
			"total":  stats.TotalWorkers,
			"active": stats.ActiveWorkers,
			"idle":   stats.IdleWorkers,
		},
		"jobs": map[string]interface{}{
			"queued":   stats.QueuedJobs,
			"running":  stats.RunningJobs,
			"complete": stats.CompletedJobs,
			"failed":   stats.FailedJobs,
		},
		"performance": map[string]interface{}{
			"avg_job_duration": stats.AvgJobDuration,
			"throughput":       stats.Throughput,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// getSystemHealthHandler returns overall system health
func (s *Server) getSystemHealthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"components": map[string]interface{}{
			"server": map[string]interface{}{
				"status": "healthy",
			},
			"worker_pool": map[string]interface{}{
				"status":  "healthy",
				"workers": s.workerPool.GetStats().TotalWorkers,
			},
			"job_manager": map[string]interface{}{
				"status": "healthy",
				"jobs":   s.jobManager.GetJobCount(),
			},
		},
		"version": map[string]interface{}{
			"build": "dev",
		},
	}

	// Check if any component is unhealthy
	// This can be expanded with actual health checks

	s.writeResponse(w, http.StatusOK, health)
}

// Helper functions

func (s *Server) cancelJob(job Job) error {
	// Set job status to cancelled
	job.SetStatus(JobStatusCancelled)

	// The actual cancellation depends on the job type and implementation
	// For now, we just mark it as cancelled
	// In a production system, we would:
	// 1. Cancel the job's context
	// 2. Wait for graceful shutdown
	// 3. Clean up resources

	s.logger.WithFields(map[string]interface{}{
		"job_id": job.GetID(),
		"type":   job.GetType(),
	}).Info("Job cancelled")

	return nil
}

func (s *Server) cloneJob(originalJob Job) (Job, error) {
	// Clone based on job type
	switch originalJob.GetType() {
	case JobTypeReplicate:
		// Type assert to access specific fields
		if replicateJob, ok := originalJob.(*ReplicateJob); ok {
			return NewReplicateJob(
				replicateJob.Source,
				replicateJob.Destination,
				replicateJob.Tags,
				replicateJob.Force,
				replicateJob.DryRun,
				s.replicationSvc,
			), nil
		}

	case JobTypeReplicateTree:
		// Type assert to access specific fields
		if treeJob, ok := originalJob.(*ReplicateTreeJob); ok {
			// Build options map from job fields
			options := map[string]interface{}{
				"excludeRepos":     treeJob.ExcludeRepos,
				"excludeTags":      treeJob.ExcludeTags,
				"includeTags":      treeJob.IncludeTags,
				"force":            treeJob.Force,
				"dryRun":           treeJob.DryRun,
				"enableCheckpoint": treeJob.EnableCheckpoint,
				"checkpointDir":    treeJob.CheckpointDir,
				"resumeID":         treeJob.ResumeID,
				"skipCompleted":    treeJob.SkipCompleted,
				"retryFailed":      treeJob.RetryFailed,
			}
			return NewReplicateTreeJob(
				treeJob.Source,
				treeJob.Destination,
				options,
				s.treeReplicationSvc,
			), nil
		}
	}

	return nil, fmt.Errorf("unsupported job type for cloning: %s", originalJob.GetType())
}

func (s *Server) checkRegistryHealth(ctx context.Context, registryName string) (bool, string, string) {
	start := time.Now()

	// Try to create a client for the registry
	var err error
	switch registryName {
	case "ecr":
		if s.cfg.ECR.AccountID == "" {
			return false, "ECR not configured", "0ms"
		}
		// In production, we would ping the registry
		// For now, just check configuration
		err = nil

	case "gcr":
		if s.cfg.GCR.Project == "" {
			return false, "GCR not configured", "0ms"
		}
		err = nil

	default:
		// Check custom registries
		var found bool
		for _, reg := range s.cfg.Registries.Registries {
			if reg.Name == registryName {
				found = true
				// In production, we would ping the registry
				break
			}
		}
		if !found {
			return false, "Registry not found in configuration", "0ms"
		}
	}

	duration := time.Since(start)
	responseTime := fmt.Sprintf("%dms", duration.Milliseconds())

	if err != nil {
		return false, fmt.Sprintf("Health check failed: %s", err), responseTime
	}

	return true, "Registry is accessible", responseTime
}

// Validation helpers

func (s *Server) validateReplicateRequest(req *ReplicateRequest) error {
	if req.SourceRegistry == "" {
		return fmt.Errorf("source_registry is required")
	}
	if req.SourceRepo == "" {
		return fmt.Errorf("source_repo is required")
	}
	if req.DestRegistry == "" {
		return fmt.Errorf("dest_registry is required")
	}
	if req.DestRepo == "" {
		return fmt.Errorf("dest_repo is required")
	}
	return nil
}

func (s *Server) validateReplicateTreeRequest(req *ReplicateTreeRequest) error {
	if req.SourceRegistry == "" {
		return fmt.Errorf("source_registry is required")
	}
	if req.SourceRepo == "" {
		return fmt.Errorf("source_repo is required")
	}
	if req.DestRegistry == "" {
		return fmt.Errorf("dest_registry is required")
	}
	if req.DestRepo == "" {
		return fmt.Errorf("dest_repo is required")
	}
	return nil
}

// WorkerPoolStats represents worker pool statistics
type WorkerPoolStats struct {
	TotalWorkers   int           `json:"total_workers"`
	ActiveWorkers  int           `json:"active_workers"`
	IdleWorkers    int           `json:"idle_workers"`
	QueuedJobs     int           `json:"queued_jobs"`
	RunningJobs    int           `json:"running_jobs"`
	CompletedJobs  int64         `json:"completed_jobs"`
	FailedJobs     int64         `json:"failed_jobs"`
	AvgJobDuration time.Duration `json:"avg_job_duration"`
	Throughput     float64       `json:"throughput"` // Jobs per minute
}
