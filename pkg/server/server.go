package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/replication"
	"freightliner/pkg/service"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents a replication server
type Server struct {
	ctx                context.Context
	cancel             context.CancelFunc
	logger             log.Logger
	cfg                *config.Config
	router             *mux.Router
	httpServer         *http.Server
	workerPool         *replication.WorkerPool
	replicationSvc     service.ReplicationService
	treeReplicationSvc *service.TreeReplicationService
	checkpointSvc      *service.CheckpointService
	jobManager         *JobManager
	metricsRegistry    *MetricsRegistry
}

// NewServer creates a new server instance
func NewServer(ctx context.Context, cfg *config.Config,
	logger log.Logger, replicationSvc service.ReplicationService,
	treeReplicationSvc *service.TreeReplicationService,
	checkpointSvc *service.CheckpointService) (*Server, error) {

	// Create a context with cancellation
	serverCtx, cancel := context.WithCancel(ctx)

	// Create router
	router := mux.NewRouter()

	// Create worker pool
	workerCount := cfg.Workers.ServeWorkers
	if workerCount == 0 && cfg.Workers.AutoDetect {
		workerCount = config.GetOptimalWorkerCount()
		logger.WithFields(map[string]interface{}{
			"workers": workerCount,
		}).Info("Auto-detected worker count for server mode")
	}

	// Create worker pool
	workerPool := replication.NewWorkerPool(workerCount, logger)

	// Create job manager
	jobManager := NewJobManager()

	// Create server
	server := &Server{
		ctx:                serverCtx,
		cancel:             cancel,
		logger:             logger,
		cfg:                cfg,
		router:             router,
		workerPool:         workerPool,
		replicationSvc:     replicationSvc,
		treeReplicationSvc: treeReplicationSvc,
		checkpointSvc:      checkpointSvc,
		jobManager:         jobManager,
		metricsRegistry:    NewMetricsRegistry(),
	}

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Register endpoints
	server.registerEndpoints()

	return server, nil
}

// Start starts the server
func (s *Server) Start() error {
	// Start worker pool
	s.workerPool.Start()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	go func() {
		s.logger.WithFields(map[string]interface{}{
			"address": s.httpServer.Addr,
			"tls":     s.cfg.Server.TLSEnabled,
		}).Info("Starting HTTP server")

		var err error
		if s.cfg.Server.TLSEnabled {
			err = s.httpServer.ListenAndServeTLS(s.cfg.Server.TLSCertFile, s.cfg.Server.TLSKeyFile)
		} else {
			err = s.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", err)
			// Signal shutdown if not already shutting down
			select {
			case <-s.ctx.Done():
				// Already shutting down
			default:
				s.cancel() // Cancel the context to signal shutdown
			}
		}
	}()

	// Wait for context cancellation or signal
	select {
	case <-s.ctx.Done():
		s.logger.Info("Server context canceled")
	case sig := <-sigChan:
		s.logger.WithFields(map[string]interface{}{
			"signal": sig.String(),
		}).Info("Received signal")
		s.cancel() // Cancel the context to signal shutdown
	}

	// Start graceful shutdown
	s.logger.Info("Shutting down server")

	// Create a context with timeout for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("HTTP server shutdown error", err)
	}

	// Stop worker pool
	s.workerPool.Stop()

	s.logger.Info("Server shutdown complete")
	return nil
}

// registerEndpoints registers server endpoints
func (s *Server) registerEndpoints() {
	// Health check endpoint
	s.router.HandleFunc(s.cfg.Server.HealthCheckPath, s.healthCheckHandler).Methods("GET")

	// Metrics endpoint
	s.router.Handle(s.cfg.Server.MetricsPath, promhttp.Handler()).Methods("GET")

	// API endpoints
	apiRouter := s.router.PathPrefix("/api/v1").Subrouter()

	// Add API key authentication middleware if enabled
	if s.cfg.Server.APIKeyAuth {
		apiRouter.Use(s.apiKeyMiddleware)
	}

	// CORS middleware for all API endpoints
	apiRouter.Use(s.corsMiddleware)

	// Register specific API endpoints
	apiRouter.HandleFunc("/replicate", s.replicateHandler).Methods("POST")
	apiRouter.HandleFunc("/replicate-tree", s.replicateTreeHandler).Methods("POST")
	apiRouter.HandleFunc("/jobs", s.listJobsHandler).Methods("GET")
	apiRouter.HandleFunc("/jobs/{id}", s.getJobHandler).Methods("GET")
	apiRouter.HandleFunc("/checkpoints", s.listCheckpointsHandler).Methods("GET")
	apiRouter.HandleFunc("/checkpoints/{id}", s.getCheckpointHandler).Methods("GET")
	apiRouter.HandleFunc("/checkpoints/{id}", s.deleteCheckpointHandler).Methods("DELETE")
}

// healthCheckHandler handles health check requests
func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"healthy"}`))
}

// apiKeyMiddleware validates the API key
func (s *Server) apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from request
		apiKey := r.Header.Get("X-API-Key")

		// Validate API key
		if apiKey != s.cfg.Server.APIKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Invalid API key"}`))
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// writeResponse writes a JSON response
func (s *Server) writeResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			s.logger.Error("Failed to encode response", err)
		}
	}
}

// writeErrorResponse writes an error response
func (s *Server) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode error response", err)
	}
}
