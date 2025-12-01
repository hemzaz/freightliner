// Package runtime provides interfaces and types for runtime security monitoring
// and threat detection in containerized environments.
package runtime

import (
	"context"
	"time"
)

// RuntimeMonitor defines the interface for runtime security monitoring systems.
// Implementations should provide real-time monitoring of container and host activities,
// detecting anomalous behavior and potential security threats.
type RuntimeMonitor interface {
	// Start begins the runtime monitoring process
	Start(ctx context.Context) error

	// Stop gracefully shuts down the runtime monitor
	Stop(ctx context.Context) error

	// GetStatus returns the current operational status of the monitor
	GetStatus() MonitorStatus

	// RegisterEventHandler registers a handler for security events
	RegisterEventHandler(handler EventHandler) error

	// UnregisterEventHandler removes a previously registered event handler
	UnregisterEventHandler(handlerID string) error

	// GetMetrics returns runtime monitoring metrics
	GetMetrics() (*MonitorMetrics, error)

	// Health performs a health check on the monitoring system
	Health(ctx context.Context) error
}

// PolicyEngine defines the interface for security policy management and enforcement.
// Implementations should evaluate events against configured policies and determine
// appropriate actions.
type PolicyEngine interface {
	// LoadPolicy loads a security policy from the provided configuration
	LoadPolicy(ctx context.Context, policy *Policy) error

	// UnloadPolicy removes a policy from the active policy set
	UnloadPolicy(ctx context.Context, policyID string) error

	// ListPolicies returns all currently loaded policies
	ListPolicies(ctx context.Context) ([]*Policy, error)

	// GetPolicy retrieves a specific policy by ID
	GetPolicy(ctx context.Context, policyID string) (*Policy, error)

	// ValidatePolicy checks if a policy is valid and can be loaded
	ValidatePolicy(policy *Policy) error

	// EvaluateEvent evaluates a security event against loaded policies
	EvaluateEvent(ctx context.Context, event *SecurityEvent) (*PolicyDecision, error)

	// UpdatePolicy modifies an existing policy
	UpdatePolicy(ctx context.Context, policy *Policy) error

	// EnablePolicy activates a policy for evaluation
	EnablePolicy(ctx context.Context, policyID string) error

	// DisablePolicy deactivates a policy without removing it
	DisablePolicy(ctx context.Context, policyID string) error
}

// AlertManager defines the interface for security alert management and routing.
// Implementations should handle alert generation, enrichment, deduplication,
// and routing to appropriate destinations.
type AlertManager interface {
	// SendAlert sends a security alert based on an event
	SendAlert(ctx context.Context, alert *Alert) error

	// GetAlert retrieves a specific alert by ID
	GetAlert(ctx context.Context, alertID string) (*Alert, error)

	// ListAlerts returns alerts based on filter criteria
	ListAlerts(ctx context.Context, filter *AlertFilter) ([]*Alert, error)

	// AcknowledgeAlert marks an alert as acknowledged by an operator
	AcknowledgeAlert(ctx context.Context, alertID string, acknowledgedBy string) error

	// ResolveAlert marks an alert as resolved
	ResolveAlert(ctx context.Context, alertID string, resolution *AlertResolution) error

	// ConfigureRouting configures alert routing rules
	ConfigureRouting(ctx context.Context, rules []*RoutingRule) error

	// GetAlertStatistics returns statistics about alerts
	GetAlertStatistics(ctx context.Context, timeRange TimeRange) (*AlertStatistics, error)

	// SupressAlert temporarily suppresses alerts matching criteria
	SuppressAlert(ctx context.Context, suppressionRule *SuppressionRule) error

	// RemoveSuppression removes an alert suppression rule
	RemoveSuppression(ctx context.Context, ruleID string) error
}

// EventHandler is a callback interface for processing security events
type EventHandler interface {
	// HandleEvent processes a security event
	HandleEvent(ctx context.Context, event *SecurityEvent) error

	// GetHandlerID returns the unique identifier for this handler
	GetHandlerID() string

	// GetHandlerMetadata returns metadata about the handler
	GetHandlerMetadata() HandlerMetadata
}

// MonitorStatus represents the operational status of a runtime monitor
type MonitorStatus struct {
	Status           string    `json:"status"`            // active, stopped, error
	StartTime        time.Time `json:"start_time"`        // When monitoring started
	EventsProcessed  int64     `json:"events_processed"`  // Total events processed
	AlertsGenerated  int64     `json:"alerts_generated"`  // Total alerts generated
	LastEventTime    time.Time `json:"last_event_time"`   // Time of last event
	HealthStatus     string    `json:"health_status"`     // healthy, degraded, unhealthy
	EngineVersion    string    `json:"engine_version"`    // Runtime engine version
	RulesLoaded      int       `json:"rules_loaded"`      // Number of active rules
	ErrorCount       int64     `json:"error_count"`       // Number of errors encountered
	ProcessingErrors []string  `json:"processing_errors"` // Recent error messages
}

// MonitorMetrics contains performance and operational metrics
type MonitorMetrics struct {
	EventsPerSecond     float64          `json:"events_per_second"`
	AlertsPerSecond     float64          `json:"alerts_per_second"`
	AverageLatency      time.Duration    `json:"average_latency_ms"`
	P95Latency          time.Duration    `json:"p95_latency_ms"`
	P99Latency          time.Duration    `json:"p99_latency_ms"`
	DroppedEvents       int64            `json:"dropped_events"`
	PolicyEvaluations   int64            `json:"policy_evaluations"`
	CPUUsagePercent     float64          `json:"cpu_usage_percent"`
	MemoryUsageMB       float64          `json:"memory_usage_mb"`
	EventTypeBreakdown  map[string]int64 `json:"event_type_breakdown"`
	SeverityBreakdown   map[string]int64 `json:"severity_breakdown"`
	ThroughputBySource  map[string]int64 `json:"throughput_by_source"`
	CollectionTimestamp time.Time        `json:"collection_timestamp"`
}

// PolicyDecision represents the result of policy evaluation
type PolicyDecision struct {
	Allow           bool              `json:"allow"`            // Whether the event is allowed
	MatchedPolicies []string          `json:"matched_policies"` // IDs of matched policies
	Actions         []string          `json:"actions"`          // Actions to take
	Severity        string            `json:"severity"`         // Severity level
	Reason          string            `json:"reason"`           // Reason for decision
	Metadata        map[string]string `json:"metadata"`         // Additional metadata
	GenerateAlert   bool              `json:"generate_alert"`   // Whether to generate an alert
}

// HandlerMetadata provides information about an event handler
type HandlerMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Priority    int               `json:"priority"`
	EventTypes  []string          `json:"event_types"`
	Tags        map[string]string `json:"tags"`
}

// AlertFilter defines criteria for filtering alerts
type AlertFilter struct {
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Severity     []string   `json:"severity,omitempty"`
	Status       []string   `json:"status,omitempty"`
	Source       []string   `json:"source,omitempty"`
	PolicyIDs    []string   `json:"policy_ids,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
	IncludeAcked bool       `json:"include_acknowledged"`
}

// AlertResolution contains information about how an alert was resolved
type AlertResolution struct {
	ResolvedBy       string            `json:"resolved_by"`
	ResolvedAt       time.Time         `json:"resolved_at"`
	Resolution       string            `json:"resolution"` // fixed, false_positive, accepted_risk, etc.
	Notes            string            `json:"notes"`
	RemediationSteps []string          `json:"remediation_steps,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// RoutingRule defines how alerts should be routed to destinations
type RoutingRule struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Enabled      bool              `json:"enabled"`
	Priority     int               `json:"priority"`
	Conditions   []Condition       `json:"conditions"`
	Destinations []Destination     `json:"destinations"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Condition defines a matching condition for routing rules
type Condition struct {
	Field    string   `json:"field"`    // severity, source, policy_id, etc.
	Operator string   `json:"operator"` // equals, contains, matches, in, etc.
	Values   []string `json:"values"`
}

// Destination defines where alerts should be sent
type Destination struct {
	Type     string            `json:"type"`   // slack, pagerduty, webhook, email, etc.
	Config   map[string]string `json:"config"` // Destination-specific configuration
	Enabled  bool              `json:"enabled"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SuppressionRule defines criteria for suppressing alerts
type SuppressionRule struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Enabled    bool              `json:"enabled"`
	Conditions []Condition       `json:"conditions"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    *time.Time        `json:"end_time,omitempty"` // nil means no expiration
	Reason     string            `json:"reason"`
	CreatedBy  string            `json:"created_by"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// AlertStatistics provides statistical information about alerts
type AlertStatistics struct {
	TimeRange           TimeRange        `json:"time_range"`
	TotalAlerts         int64            `json:"total_alerts"`
	AlertsBySeverity    map[string]int64 `json:"alerts_by_severity"`
	AlertsBySource      map[string]int64 `json:"alerts_by_source"`
	AlertsByStatus      map[string]int64 `json:"alerts_by_status"`
	AlertsByPolicy      map[string]int64 `json:"alerts_by_policy"`
	TopThreatIndicators []string         `json:"top_threat_indicators"`
	MeanTimeToAck       time.Duration    `json:"mean_time_to_ack"`
	MeanTimeToResolve   time.Duration    `json:"mean_time_to_resolve"`
	TrendData           []TrendDataPoint `json:"trend_data"`
}

// TimeRange defines a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// TrendDataPoint represents a point in a time series
type TrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int64     `json:"count"`
	Severity  string    `json:"severity,omitempty"`
}
