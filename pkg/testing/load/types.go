package load

import (
	"time"
)

// ScenarioConfig defines configuration for a load test scenario
type ScenarioConfig struct {
	Name               string
	Description        string
	Scenario           LoadTestScenario
	Duration           time.Duration
	ConcurrentWorkers  int
	ExpectedThroughput float64 // MB/s
	MemoryLimitMB      int64   // Maximum memory usage in MB
	MaxFailureRate     float64 // Maximum acceptable failure rate (0.0-1.0)
	Images             []ContainerImage
	NetworkConditions  NetworkConditions
	ValidationCriteria ValidationCriteria
}

// ContainerImage represents a container image for testing
type ContainerImage struct {
	Name   string
	Tag    string
	SizeMB int64
}

// NetworkConditions simulates various network conditions
type NetworkConditions struct {
	PacketLossRate       float64        // Percentage of packet loss (0.0-1.0)
	LatencyMs            int            // Network latency in milliseconds
	BandwidthLimitMBps   float64        // Bandwidth limit in MB/s
	ServiceInterruptions []Interruption // Planned service interruptions
}

// Interruption represents a planned service interruption
type Interruption struct {
	StartTime time.Duration // When to start the interruption (relative to test start)
	Duration  time.Duration // How long the interruption lasts
	Severity  string        // "partial" or "complete"
}

// ValidationCriteria defines success criteria for load tests
type ValidationCriteria struct {
	MinThroughputMBps   float64 // Minimum acceptable throughput
	MaxLatencyMs        int     // Maximum acceptable latency
	MaxP99LatencyMs     int     // Maximum acceptable P99 latency
	MaxMemoryUsageMB    int64   // Maximum memory usage
	MaxFailureRate      float64 // Maximum failure rate (0.0-1.0)
	RequiredSuccessRate float64 // Required success rate (0.0-1.0)
}

// LoadTestResults represents the results of a load test
type LoadTestResults struct {
	Scenario               string
	StartTime              time.Time
	EndTime                time.Time
	Duration               time.Duration
	TotalRepositories      int
	SuccessfulRepositories int
	FailedRepositories     int
	ProcessedImages        int
	FailedImages           int
	AverageThroughputMBps  float64
	PeakThroughputMBps     float64
	AverageLatencyMs       int
	P99LatencyMs           int64
	MemoryUsageMB          int64
	NetworkEfficiencyScore float64
	ConnectionReuseRate    float64
	FailureRate            float64
	ValidationPassed       bool
}

// LoadTestScenario represents different types of load test scenarios
type LoadTestScenario string

const (
	HighVolumeReplication LoadTestScenario = "high_volume_replication"
	LargeImageStress      LoadTestScenario = "large_image_stress"
	NetworkResilience     LoadTestScenario = "network_resilience"
	BurstReplication      LoadTestScenario = "burst_replication"
	SustainedThroughput   LoadTestScenario = "sustained_throughput"
	MixedContainerSizes   LoadTestScenario = "mixed_container_sizes"
)
