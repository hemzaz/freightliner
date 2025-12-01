package runtime

import (
	"time"
)

// SecurityEvent represents a runtime security event detected by the monitoring system.
// Events are generated from system calls, network activity, file access, and other
// runtime behaviors that may indicate security threats.
type SecurityEvent struct {
	// Identification
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	EventType string    `json:"event_type"` // syscall, network, file, process, etc.
	Source    string    `json:"source"`     // container_id, host, service name, etc.

	// Severity and Classification
	Severity    Severity `json:"severity"`
	Priority    int      `json:"priority"`
	Category    string   `json:"category"`     // privilege_escalation, data_exfiltration, etc.
	ThreatScore float64  `json:"threat_score"` // 0.0 to 10.0

	// Event Details
	Description string            `json:"description"`
	RuleName    string            `json:"rule_name"`
	RuleID      string            `json:"rule_id"`
	Tags        []string          `json:"tags"`
	Indicators  []ThreatIndicator `json:"indicators"`

	// Context Information
	Container *ContainerContext `json:"container,omitempty"`
	Process   *ProcessContext   `json:"process,omitempty"`
	Network   *NetworkContext   `json:"network,omitempty"`
	File      *FileContext      `json:"file,omitempty"`
	User      *UserContext      `json:"user,omitempty"`

	// Metadata
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	RawEvent  string                 `json:"raw_event,omitempty"` // Original event data
	Processed bool                   `json:"processed"`
	Dismissed bool                   `json:"dismissed"`
}

// Policy represents a runtime security policy that defines detection rules and actions.
type Policy struct {
	// Identification
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Policy Configuration
	Enabled  bool     `json:"enabled"`
	Priority int      `json:"priority"`
	Scope    []string `json:"scope"` // cluster, namespace, workload, container
	Tags     []string `json:"tags"`
	Category string   `json:"category"`
	Severity Severity `json:"severity"`

	// Rules and Conditions
	Rules      []Rule            `json:"rules"`
	Conditions []PolicyCondition `json:"conditions"`
	Exceptions []Exception       `json:"exceptions,omitempty"`

	// Actions
	Actions     []Action          `json:"actions"`
	OnViolation OnViolationAction `json:"on_violation"`

	// Metadata
	Author     string            `json:"author"`
	Source     string            `json:"source"`               // custom, falco, apparmor, etc.
	Compliance []string          `json:"compliance,omitempty"` // CIS, PCI-DSS, etc.
	References []string          `json:"references,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Alert represents a security alert generated from a security event.
type Alert struct {
	// Identification
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	PolicyID  string    `json:"policy_id"`
	Timestamp time.Time `json:"timestamp"`

	// Alert Details
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	Status      string   `json:"status"` // new, acknowledged, investigating, resolved, dismissed
	Priority    int      `json:"priority"`

	// Classification
	Category         string   `json:"category"`
	Tags             []string `json:"tags"`
	ThreatIndicators []string `json:"threat_indicators"`
	MITRE_ATT        []string `json:"mitre_attack,omitempty"` // MITRE ATT&CK framework IDs

	// Context
	Source        string             `json:"source"`
	Destination   string             `json:"destination,omitempty"`
	Affected      []AffectedResource `json:"affected"`
	RelatedEvents []string           `json:"related_events,omitempty"`

	// Response
	RecommendedActions []string `json:"recommended_actions"`
	PlaybookID         string   `json:"playbook_id,omitempty"`
	AssignedTo         string   `json:"assigned_to,omitempty"`

	// Lifecycle
	AcknowledgedAt *time.Time       `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string           `json:"acknowledged_by,omitempty"`
	ResolvedAt     *time.Time       `json:"resolved_at,omitempty"`
	Resolution     *AlertResolution `json:"resolution,omitempty"`

	// Metadata
	Evidence map[string]interface{} `json:"evidence,omitempty"`
	Metadata map[string]string      `json:"metadata,omitempty"`
	Count    int                    `json:"count"` // Number of similar events aggregated
}

// ThreatIndicator represents an indicator of compromise or suspicious activity.
type ThreatIndicator struct {
	Type        string            `json:"type"` // ioc, behavior, anomaly, signature
	Value       string            `json:"value"`
	Confidence  float64           `json:"confidence"` // 0.0 to 1.0
	Severity    Severity          `json:"severity"`
	Description string            `json:"description"`
	FirstSeen   time.Time         `json:"first_seen"`
	LastSeen    time.Time         `json:"last_seen"`
	Count       int               `json:"count"`
	Source      string            `json:"source,omitempty"` // threat intel feed, ML model, etc.
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Severity represents the severity level of an event or alert
type Severity string

const (
	SeverityCritical      Severity = "critical"
	SeverityHigh          Severity = "high"
	SeverityMedium        Severity = "medium"
	SeverityLow           Severity = "low"
	SeverityInformational Severity = "informational"
)

// ContainerContext provides container-specific context for security events
type ContainerContext struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	ImageDigest string            `json:"image_digest,omitempty"`
	Namespace   string            `json:"namespace"`
	PodName     string            `json:"pod_name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Privileged  bool              `json:"privileged"`
	Runtime     string            `json:"runtime"` // docker, containerd, cri-o
}

// ProcessContext provides process-specific context for security events
type ProcessContext struct {
	PID         int               `json:"pid"`
	PPID        int               `json:"ppid"`
	Name        string            `json:"name"`
	Cmdline     string            `json:"cmdline"`
	Exe         string            `json:"exe"`
	UID         int               `json:"uid"`
	GID         int               `json:"gid"`
	User        string            `json:"user"`
	Group       string            `json:"group"`
	Cwd         string            `json:"cwd,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// NetworkContext provides network-specific context for security events
type NetworkContext struct {
	Protocol      string `json:"protocol"` // tcp, udp, icmp, etc.
	SourceIP      string `json:"source_ip"`
	SourcePort    int    `json:"source_port"`
	DestIP        string `json:"dest_ip"`
	DestPort      int    `json:"dest_port"`
	Direction     string `json:"direction"` // inbound, outbound
	BytesSent     int64  `json:"bytes_sent"`
	BytesReceived int64  `json:"bytes_received"`
	ConnectionID  string `json:"connection_id,omitempty"`
}

// FileContext provides file-specific context for security events
type FileContext struct {
	Path        string    `json:"path"`
	Mode        string    `json:"mode,omitempty"`
	Operation   string    `json:"operation"` // read, write, execute, delete
	Permissions string    `json:"permissions,omitempty"`
	Owner       string    `json:"owner,omitempty"`
	Group       string    `json:"group,omitempty"`
	Size        int64     `json:"size,omitempty"`
	Hash        *FileHash `json:"hash,omitempty"`
}

// FileHash contains file hash information
type FileHash struct {
	MD5    string `json:"md5,omitempty"`
	SHA1   string `json:"sha1,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

// UserContext provides user-specific context for security events
type UserContext struct {
	UID        int      `json:"uid"`
	Username   string   `json:"username"`
	GID        int      `json:"gid"`
	GroupName  string   `json:"group_name"`
	Groups     []string `json:"groups,omitempty"`
	SessionID  string   `json:"session_id,omitempty"`
	AuthMethod string   `json:"auth_method,omitempty"`
	Privileges []string `json:"privileges,omitempty"`
}

// Rule represents a single detection rule within a policy
type Rule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Condition   string            `json:"condition"` // Rule logic/expression
	Output      string            `json:"output"`    // Alert message template
	Priority    Severity          `json:"priority"`
	Tags        []string          `json:"tags"`
	Enabled     bool              `json:"enabled"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// PolicyCondition defines a condition that must be met for policy evaluation
type PolicyCondition struct {
	Field         string      `json:"field"`
	Operator      string      `json:"operator"` // equals, not_equals, contains, matches, etc.
	Value         interface{} `json:"value"`
	CaseSensitive bool        `json:"case_sensitive"`
}

// Exception defines an exception to a policy rule
type Exception struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Conditions  []PolicyCondition `json:"conditions"`
	Reason      string            `json:"reason"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	CreatedBy   string            `json:"created_by"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Action defines an action to take when a policy is violated
type Action struct {
	Type     string            `json:"type"`   // alert, block, log, quarantine, kill
	Target   string            `json:"target"` // container, process, network
	Config   map[string]string `json:"config"`
	Enabled  bool              `json:"enabled"`
	Priority int               `json:"priority"`
}

// OnViolationAction defines what happens when a policy is violated
type OnViolationAction struct {
	GenerateAlert bool     `json:"generate_alert"`
	BlockAction   bool     `json:"block_action"`
	LogEvent      bool     `json:"log_event"`
	Notify        []string `json:"notify,omitempty"` // Notification channels
	Quarantine    bool     `json:"quarantine"`
	Terminate     bool     `json:"terminate"`
	CustomActions []Action `json:"custom_actions,omitempty"`
}

// AffectedResource represents a resource affected by a security event
type AffectedResource struct {
	Type       string            `json:"type"` // container, pod, node, service
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace,omitempty"`
	Cluster    string            `json:"cluster,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Impact     string            `json:"impact"` // high, medium, low
	Remediated bool              `json:"remediated"`
}
