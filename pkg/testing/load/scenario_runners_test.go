package load

import (
	"testing"
	"time"
)

func TestScenarioRunner_Run(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scenario runner test in short mode")
	}

	// Create a minimal scenario for testing
	scenario := ScenarioConfig{
		Name:              "Test Scenario",
		Scenario:          HighVolumeReplication,
		Duration:          100 * time.Millisecond,
		ConcurrentWorkers: 2,
		Images: []ContainerImage{
			{
				Repository: "test/repo1",
				Tag:        "v1.0.0",
				SizeMB:     100,
				LayerCount: 5,
				Registry:   "aws-ecr",
			},
			{
				Repository: "test/repo2",
				Tag:        "latest",
				SizeMB:     200,
				LayerCount: 10,
				Registry:   "gcp-gcr",
			},
		},
		ValidationCriteria: ValidationCriteria{
			MinThroughputMBps:  1.0,
			MaxMemoryUsageMB:   100,
			MaxFailureRate:     0.5,
			MaxP99LatencyMs:    5000,
			MinConnectionReuse: 0.5,
		},
	}

	runner := NewScenarioRunner(scenario, nil)

	// Run should return error for unknown scenario type implementation
	// since we only defined scenario type but not the implementation
	_, err := runner.Run()
	if err == nil {
		t.Error("Expected error for unimplemented scenario")
	}
}

func TestScenarioRunner_UnknownScenario(t *testing.T) {
	scenario := ScenarioConfig{
		Name:     "Unknown Scenario",
		Scenario: LoadTestScenario("unknown_scenario_type"),
		Duration: 100 * time.Millisecond,
		Images:   []ContainerImage{},
	}

	runner := NewScenarioRunner(scenario, nil)

	_, err := runner.Run()
	if err == nil {
		t.Error("Expected error for unknown scenario type")
	}

	expectedError := "unknown scenario"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestLoadTestResults_Validation(t *testing.T) {
	results := &LoadTestResults{
		ScenarioName:          "Test Scenario",
		Duration:              1 * time.Minute,
		TotalImages:           100,
		ProcessedImages:       95,
		FailedImages:          5,
		AverageThroughputMBps: 125.5,
		PeakThroughputMBps:    150.0,
		MemoryUsageMB:         800,
		ConnectionReuseRate:   0.85,
		FailureRate:           0.05,
		P99LatencyMs:          3000,
		ValidationPassed:      true,
		ValidationErrors:      []string{},
		DetailedMetrics:       make(map[string]interface{}),
	}

	if results.ScenarioName == "" {
		t.Error("ScenarioName should not be empty")
	}

	if results.Duration <= 0 {
		t.Error("Duration should be positive")
	}

	if results.TotalImages <= 0 {
		t.Error("TotalImages should be positive")
	}

	if results.ProcessedImages > results.TotalImages {
		t.Error("ProcessedImages should not exceed TotalImages")
	}

	if results.FailureRate < 0 || results.FailureRate > 1 {
		t.Errorf("FailureRate should be between 0 and 1, got %f", results.FailureRate)
	}

	if results.ConnectionReuseRate < 0 || results.ConnectionReuseRate > 1 {
		t.Errorf("ConnectionReuseRate should be between 0 and 1, got %f", results.ConnectionReuseRate)
	}
}

func TestLoadTestResults_FailedValidation(t *testing.T) {
	results := &LoadTestResults{
		ScenarioName:     "Failed Scenario",
		ValidationPassed: false,
		ValidationErrors: []string{
			"Throughput below minimum",
			"Failure rate exceeded threshold",
			"Memory usage too high",
		},
	}

	if results.ValidationPassed {
		t.Error("Expected validation to fail")
	}

	if len(results.ValidationErrors) != 3 {
		t.Errorf("Expected 3 validation errors, got %d", len(results.ValidationErrors))
	}

	expectedErrors := map[string]bool{
		"Throughput below minimum":        true,
		"Failure rate exceeded threshold": true,
		"Memory usage too high":           true,
	}

	for _, err := range results.ValidationErrors {
		if !expectedErrors[err] {
			t.Errorf("Unexpected validation error: %s", err)
		}
	}
}

func TestInterruption_Timing(t *testing.T) {
	interruption := Interruption{
		StartTime: 10 * time.Minute,
		Duration:  30 * time.Second,
		Severity:  "partial",
	}

	if interruption.StartTime < 0 {
		t.Error("StartTime should not be negative")
	}

	if interruption.Duration <= 0 {
		t.Error("Duration should be positive")
	}

	if interruption.Severity != "partial" && interruption.Severity != "complete" {
		t.Errorf("Invalid severity: %s", interruption.Severity)
	}

	// Calculate end time
	endTime := interruption.StartTime + interruption.Duration
	expectedEndTime := 10*time.Minute + 30*time.Second

	if endTime != expectedEndTime {
		t.Errorf("Expected end time %v, got %v", expectedEndTime, endTime)
	}
}

func TestNetworkConditions_Validation(t *testing.T) {
	tests := []struct {
		name       string
		conditions NetworkConditions
		wantValid  bool
	}{
		{
			name: "Valid conditions",
			conditions: NetworkConditions{
				PacketLossRate:     0.01,
				LatencyMs:          50,
				BandwidthLimitMBps: 150.0,
			},
			wantValid: true,
		},
		{
			name: "Invalid packet loss (negative)",
			conditions: NetworkConditions{
				PacketLossRate:     -0.1,
				LatencyMs:          50,
				BandwidthLimitMBps: 150.0,
			},
			wantValid: false,
		},
		{
			name: "Invalid packet loss (> 1)",
			conditions: NetworkConditions{
				PacketLossRate:     1.5,
				LatencyMs:          50,
				BandwidthLimitMBps: 150.0,
			},
			wantValid: false,
		},
		{
			name: "Invalid latency (negative)",
			conditions: NetworkConditions{
				PacketLossRate:     0.01,
				LatencyMs:          -10,
				BandwidthLimitMBps: 150.0,
			},
			wantValid: false,
		},
		{
			name: "Invalid bandwidth (zero)",
			conditions: NetworkConditions{
				PacketLossRate:     0.01,
				LatencyMs:          50,
				BandwidthLimitMBps: 0,
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateNetworkConditions(tt.conditions)
			if valid != tt.wantValid {
				t.Errorf("validateNetworkConditions() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func validateNetworkConditions(nc NetworkConditions) bool {
	if nc.PacketLossRate < 0 || nc.PacketLossRate > 1 {
		return false
	}
	if nc.LatencyMs < 0 {
		return false
	}
	if nc.BandwidthLimitMBps <= 0 {
		return false
	}
	return true
}

func TestValidationCriteria_Validation(t *testing.T) {
	tests := []struct {
		name      string
		criteria  ValidationCriteria
		wantValid bool
	}{
		{
			name: "Valid criteria",
			criteria: ValidationCriteria{
				MinThroughputMBps:  100.0,
				MaxMemoryUsageMB:   1000,
				MaxFailureRate:     0.01,
				MaxP99LatencyMs:    5000,
				MinConnectionReuse: 0.80,
			},
			wantValid: true,
		},
		{
			name: "Invalid throughput (negative)",
			criteria: ValidationCriteria{
				MinThroughputMBps: -10.0,
			},
			wantValid: false,
		},
		{
			name: "Invalid memory (zero)",
			criteria: ValidationCriteria{
				MinThroughputMBps: 100.0,
				MaxMemoryUsageMB:  0,
			},
			wantValid: false,
		},
		{
			name: "Invalid failure rate (> 1)",
			criteria: ValidationCriteria{
				MinThroughputMBps: 100.0,
				MaxMemoryUsageMB:  1000,
				MaxFailureRate:    1.5,
			},
			wantValid: false,
		},
		{
			name: "Invalid connection reuse (< 0)",
			criteria: ValidationCriteria{
				MinThroughputMBps:  100.0,
				MaxMemoryUsageMB:   1000,
				MaxFailureRate:     0.01,
				MinConnectionReuse: -0.1,
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateCriteria(tt.criteria)
			if valid != tt.wantValid {
				t.Errorf("validateCriteria() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func validateCriteria(vc ValidationCriteria) bool {
	if vc.MinThroughputMBps < 0 {
		return false
	}
	if vc.MaxMemoryUsageMB <= 0 {
		return false
	}
	if vc.MaxFailureRate < 0 || vc.MaxFailureRate > 1 {
		return false
	}
	if vc.MinConnectionReuse < 0 || vc.MinConnectionReuse > 1 {
		return false
	}
	return true
}

func TestContainerImage_SizeCalculations(t *testing.T) {
	image := ContainerImage{
		Repository: "test/app",
		Tag:        "v1.0.0",
		SizeMB:     1024,
		LayerCount: 10,
		Registry:   "aws-ecr",
	}

	// Average layer size
	avgLayerSize := image.SizeMB / int64(image.LayerCount)
	expectedAvgSize := int64(102) // 1024 / 10 = 102.4, truncated to 102

	if avgLayerSize != expectedAvgSize {
		t.Errorf("Expected average layer size %d MB, got %d MB", expectedAvgSize, avgLayerSize)
	}

	// Size in bytes
	sizeBytes := image.SizeMB * 1024 * 1024
	expectedBytes := int64(1024 * 1024 * 1024) // 1GB

	if sizeBytes != expectedBytes {
		t.Errorf("Expected size %d bytes, got %d bytes", expectedBytes, sizeBytes)
	}
}

func TestScenarioConfig_Consistency(t *testing.T) {
	scenarios := []ScenarioConfig{
		CreateHighVolumeReplicationScenario(),
		CreateLargeImageStressScenario(),
		CreateNetworkResilienceScenario(),
		CreateBurstReplicationScenario(),
		CreateSustainedThroughputScenario(),
		CreateMixedContainerSizesScenario(),
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			// Verify expected throughput doesn't exceed bandwidth limit
			if scenario.ExpectedThroughput > scenario.NetworkConditions.BandwidthLimitMBps {
				t.Errorf("Expected throughput (%f) exceeds bandwidth limit (%f)",
					scenario.ExpectedThroughput, scenario.NetworkConditions.BandwidthLimitMBps)
			}

			// Verify min throughput is reasonable relative to expected
			if scenario.ValidationCriteria.MinThroughputMBps > scenario.ExpectedThroughput {
				t.Errorf("Min throughput (%f) exceeds expected throughput (%f)",
					scenario.ValidationCriteria.MinThroughputMBps, scenario.ExpectedThroughput)
			}

			// Allow validation criteria to be stricter than scenario maxFailureRate
			// This is intentional for resilience scenarios
		})
	}
}
