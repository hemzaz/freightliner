package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Registry wraps Prometheus registry with application-specific metrics
type Registry struct {
	registry *prometheus.Registry

	// HTTP metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// Replication metrics
	replicationTotal       *prometheus.CounterVec
	replicationDuration    *prometheus.HistogramVec
	replicationBytesTotal  *prometheus.CounterVec
	replicationLayersTotal *prometheus.CounterVec
	replicationErrorsTotal *prometheus.CounterVec

	// Tag copy metrics
	tagCopyTotal      *prometheus.CounterVec
	tagCopyDuration   *prometheus.HistogramVec
	tagCopyBytesTotal *prometheus.CounterVec

	// Job metrics
	jobsTotal   *prometheus.CounterVec
	jobDuration *prometheus.HistogramVec
	jobsActive  prometheus.Gauge

	// Worker pool metrics
	workerPoolSize   prometheus.Gauge
	workerPoolActive prometheus.Gauge
	workerPoolQueued prometheus.Gauge

	// System metrics
	memoryUsage    prometheus.Gauge
	goroutineCount prometheus.Gauge
	panicTotal     *prometheus.CounterVec

	// Authentication metrics
	authFailuresTotal *prometheus.CounterVec
}

// NewRegistry creates a new metrics registry with all application metrics
func NewRegistry() *Registry {
	reg := prometheus.NewRegistry()

	// Create metrics
	r := &Registry{
		registry: reg,

		// HTTP metrics
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freightliner_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "freightliner_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		httpRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "freightliner_http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		// Replication metrics
		replicationTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freightliner_replications_total",
				Help: "Total number of replication operations",
			},
			[]string{"source_registry", "dest_registry", "status"},
		),
		replicationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "freightliner_replication_duration_seconds",
				Help:    "Replication operation duration in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
			},
			[]string{"source_registry", "dest_registry"},
		),
		replicationBytesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freightliner_replication_bytes_total",
				Help: "Total bytes transferred during replication",
			},
			[]string{"source_registry", "dest_registry"},
		),
		replicationLayersTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freightliner_replication_layers_total",
				Help: "Total layers transferred during replication",
			},
			[]string{"source_registry", "dest_registry"},
		),
		replicationErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freightliner_replication_errors_total",
				Help: "Total number of replication errors",
			},
			[]string{"source_registry", "dest_registry", "error_type"},
		),

		// Tag copy metrics
		tagCopyTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freightliner_tag_copy_total",
				Help: "Total number of tag copy operations",
			},
			[]string{"source_repo", "dest_repo", "status"},
		),
		tagCopyDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "freightliner_tag_copy_duration_seconds",
				Help:    "Tag copy operation duration in seconds",
				Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300},
			},
			[]string{"source_repo", "dest_repo"},
		),
		tagCopyBytesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freightliner_tag_copy_bytes_total",
				Help: "Total bytes transferred during tag copy",
			},
			[]string{"source_repo", "dest_repo"},
		),

		// Job metrics
		jobsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freightliner_jobs_total",
				Help: "Total number of jobs",
			},
			[]string{"type", "status"},
		),
		jobDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "freightliner_job_duration_seconds",
				Help:    "Job execution duration in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
			},
			[]string{"type"},
		),
		jobsActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "freightliner_jobs_active",
				Help: "Number of currently active jobs",
			},
		),

		// Worker pool metrics
		workerPoolSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "freightliner_worker_pool_size",
				Help: "Total number of workers in the pool",
			},
		),
		workerPoolActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "freightliner_worker_pool_active",
				Help: "Number of active workers",
			},
		),
		workerPoolQueued: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "freightliner_worker_pool_queued",
				Help: "Number of queued jobs waiting for workers",
			},
		),

		// System metrics
		memoryUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "freightliner_memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
		),
		goroutineCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "freightliner_goroutines_count",
				Help: "Current number of goroutines",
			},
		),
		panicTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freightliner_panics_total",
				Help: "Total number of panics",
			},
			[]string{"component"},
		),

		// Authentication metrics
		authFailuresTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freightliner_auth_failures_total",
				Help: "Total number of authentication failures",
			},
			[]string{"type"},
		),
	}

	// Register all metrics
	r.registerMetrics()

	return r
}

// registerMetrics registers all metrics with the Prometheus registry
func (r *Registry) registerMetrics() {
	metrics := []prometheus.Collector{
		r.httpRequestsTotal,
		r.httpRequestDuration,
		r.httpRequestsInFlight,
		r.replicationTotal,
		r.replicationDuration,
		r.replicationBytesTotal,
		r.replicationLayersTotal,
		r.replicationErrorsTotal,
		r.tagCopyTotal,
		r.tagCopyDuration,
		r.tagCopyBytesTotal,
		r.jobsTotal,
		r.jobDuration,
		r.jobsActive,
		r.workerPoolSize,
		r.workerPoolActive,
		r.workerPoolQueued,
		r.memoryUsage,
		r.goroutineCount,
		r.panicTotal,
		r.authFailuresTotal,
	}

	for _, metric := range metrics {
		r.registry.MustRegister(metric)
	}
}

// GetRegistry returns the underlying Prometheus registry
func (r *Registry) GetRegistry() *prometheus.Registry {
	return r.registry
}

// HTTP metrics methods
func (r *Registry) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	r.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	r.httpRequestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
}

func (r *Registry) IncHTTPRequestsInFlight() {
	r.httpRequestsInFlight.Inc()
}

func (r *Registry) DecHTTPRequestsInFlight() {
	r.httpRequestsInFlight.Dec()
}

// Replication metrics methods
func (r *Registry) RecordReplication(sourceRegistry, destRegistry, status string, duration time.Duration, bytes int64, layers int) {
	r.replicationTotal.WithLabelValues(sourceRegistry, destRegistry, status).Inc()
	r.replicationDuration.WithLabelValues(sourceRegistry, destRegistry).Observe(duration.Seconds())
	if bytes > 0 {
		r.replicationBytesTotal.WithLabelValues(sourceRegistry, destRegistry).Add(float64(bytes))
	}
	if layers > 0 {
		r.replicationLayersTotal.WithLabelValues(sourceRegistry, destRegistry).Add(float64(layers))
	}
}

func (r *Registry) RecordReplicationError(sourceRegistry, destRegistry, errorType string) {
	r.replicationErrorsTotal.WithLabelValues(sourceRegistry, destRegistry, errorType).Inc()
}

// Tag copy metrics methods
func (r *Registry) RecordTagCopy(sourceRepo, destRepo, status string, duration time.Duration, bytes int64) {
	r.tagCopyTotal.WithLabelValues(sourceRepo, destRepo, status).Inc()
	r.tagCopyDuration.WithLabelValues(sourceRepo, destRepo).Observe(duration.Seconds())
	if bytes > 0 {
		r.tagCopyBytesTotal.WithLabelValues(sourceRepo, destRepo).Add(float64(bytes))
	}
}

// Job metrics methods
func (r *Registry) RecordJob(jobType, status string, duration time.Duration) {
	r.jobsTotal.WithLabelValues(jobType, status).Inc()
	r.jobDuration.WithLabelValues(jobType).Observe(duration.Seconds())
}

func (r *Registry) SetJobsActive(count int) {
	r.jobsActive.Set(float64(count))
}

// Worker pool metrics methods
func (r *Registry) SetWorkerPoolSize(size int) {
	r.workerPoolSize.Set(float64(size))
}

func (r *Registry) SetWorkerPoolActive(active int) {
	r.workerPoolActive.Set(float64(active))
}

func (r *Registry) SetWorkerPoolQueued(queued int) {
	r.workerPoolQueued.Set(float64(queued))
}

// System metrics methods
func (r *Registry) SetMemoryUsage(bytes uint64) {
	r.memoryUsage.Set(float64(bytes))
}

func (r *Registry) SetGoroutineCount(count int) {
	r.goroutineCount.Set(float64(count))
}

func (r *Registry) RecordPanic(component string) {
	r.panicTotal.WithLabelValues(component).Inc()
}

// Authentication metrics methods
func (r *Registry) RecordAuthFailure(authType string) {
	r.authFailuresTotal.WithLabelValues(authType).Inc()
}
