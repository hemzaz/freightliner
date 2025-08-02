package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version,omitempty"`
	Uptime    string                 `json:"uptime,omitempty"`
	Checks    map[string]CheckResult `json:"checks,omitempty"`
	System    *SystemInfo            `json:"system,omitempty"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status    string        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	Error     string        `json:"error,omitempty"`
}

// SystemInfo contains system information
type SystemInfo struct {
	GoVersion    string      `json:"go_version"`
	OS           string      `json:"os"`
	Arch         string      `json:"arch"`
	NumCPU       int         `json:"num_cpu"`
	NumGoroutine int         `json:"num_goroutine"`
	MemoryUsage  *MemoryInfo `json:"memory,omitempty"`
}

// MemoryInfo contains memory usage information
type MemoryInfo struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

var (
	serverStartTime = time.Now()
	version         = "dev"
	buildTime       = "unknown"
	gitCommit       = "unknown"
)

// handleHealth handles basic health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(serverStartTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		// Log error but don't change response since headers are already written
		fmt.Printf("Failed to encode health response: %v\n", err)
	}
}

// handleReadiness handles readiness probe requests
func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]CheckResult)
	overallStatus := "ready"
	httpStatus := http.StatusOK

	// Check worker pool status
	if s.workerPool != nil {
		start := time.Now()
		if s.workerPool.IsHealthy() {
			checks["worker_pool"] = CheckResult{
				Status:    "healthy",
				Message:   "Worker pool is operational",
				Duration:  time.Since(start),
				Timestamp: time.Now(),
			}
		} else {
			checks["worker_pool"] = CheckResult{
				Status:    "unhealthy",
				Message:   "Worker pool is not operational",
				Duration:  time.Since(start),
				Timestamp: time.Now(),
				Error:     "Worker pool health check failed",
			}
			overallStatus = "not_ready"
			httpStatus = http.StatusServiceUnavailable
		}
	}

	// Check services
	checks["replication_service"] = s.checkServiceHealth("replication", s.replicationSvc != nil)
	checks["tree_replication_service"] = s.checkServiceHealth("tree_replication", s.treeReplicationSvc != nil)
	checks["checkpoint_service"] = s.checkServiceHealth("checkpoint", s.checkpointSvc != nil)

	// Check for any failed checks
	for _, check := range checks {
		if check.Status != "healthy" {
			overallStatus = "not_ready"
			httpStatus = http.StatusServiceUnavailable
			break
		}
	}

	health := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Uptime:    time.Since(serverStartTime).String(),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		// Log error but don't change response since headers are already written
		fmt.Printf("Failed to encode readiness response: %v\n", err)
	}
}

// handleLiveness handles liveness probe requests
func (s *Server) handleLiveness(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]CheckResult)
	overallStatus := "alive"
	httpStatus := http.StatusOK

	// Check if server context is still valid
	start := time.Now()
	select {
	case <-s.ctx.Done():
		checks["context"] = CheckResult{
			Status:    "unhealthy",
			Message:   "Server context is cancelled",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
			Error:     "Server is shutting down",
		}
		overallStatus = "not_alive"
		httpStatus = http.StatusServiceUnavailable
	default:
		checks["context"] = CheckResult{
			Status:    "healthy",
			Message:   "Server context is active",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	// Basic memory check
	start = time.Now()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Check if memory usage is reasonable (less than 1GB for now)
	if m.Alloc < 1<<30 { // 1GB
		checks["memory"] = CheckResult{
			Status:    "healthy",
			Message:   fmt.Sprintf("Memory usage: %d MB", m.Alloc/(1<<20)),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	} else {
		checks["memory"] = CheckResult{
			Status:    "warning",
			Message:   fmt.Sprintf("High memory usage: %d MB", m.Alloc/(1<<20)),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	health := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Uptime:    time.Since(serverStartTime).String(),
		Checks:    checks,
		System:    s.getSystemInfo(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		// Log error but don't change response since headers are already written
		fmt.Printf("Failed to encode liveness response: %v\n", err)
	}
}

// handleSystemInfo handles system information requests
func (s *Server) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"service":    "freightliner",
		"version":    version,
		"build_time": buildTime,
		"git_commit": gitCommit,
		"uptime":     time.Since(serverStartTime).String(),
		"system":     s.getSystemInfo(),
		"configuration": map[string]interface{}{
			"server_port":     s.cfg.Server.Port,
			"metrics_enabled": s.cfg.Metrics.Enabled,
			"worker_count":    s.cfg.Workers.ServeWorkers,
			"tls_enabled":     s.cfg.Server.TLSEnabled,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(info); err != nil {
		// Log error but don't change response since headers are already written
		fmt.Printf("Failed to encode system info response: %v\n", err)
	}
}

// handleVersion handles version information requests
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	versionInfo := map[string]string{
		"version":    version,
		"build_time": buildTime,
		"git_commit": gitCommit,
		"go_version": runtime.Version(),
		"os_arch":    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(versionInfo); err != nil {
		// Log error but don't change response since headers are already written
		fmt.Printf("Failed to encode version response: %v\n", err)
	}
}

// checkServiceHealth performs a basic service health check
func (s *Server) checkServiceHealth(serviceName string, serviceAvailable bool) CheckResult {
	start := time.Now()

	if serviceAvailable {
		return CheckResult{
			Status:    "healthy",
			Message:   fmt.Sprintf("%s service is available", serviceName),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	return CheckResult{
		Status:    "unhealthy",
		Message:   fmt.Sprintf("%s service is not available", serviceName),
		Duration:  time.Since(start),
		Timestamp: time.Now(),
		Error:     fmt.Sprintf("%s service is nil", serviceName),
	}
}

// getSystemInfo returns current system information
func (s *Server) getSystemInfo() *SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &SystemInfo{
		GoVersion:    runtime.Version(),
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		MemoryUsage: &MemoryInfo{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
		},
	}
}
