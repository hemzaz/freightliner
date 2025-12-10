package load

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateK6ThresholdsConfig(t *testing.T) {
	scenario := CreateHighVolumeReplicationScenario()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "thresholds.json")

	err := GenerateK6ThresholdsConfig(scenario, outputPath)
	if err != nil {
		t.Fatalf("Failed to generate k6 thresholds config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Thresholds config file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read thresholds config: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "thresholds") {
		t.Error("Config should contain 'thresholds' key")
	}

	if !strings.Contains(contentStr, "options") {
		t.Error("Config should contain 'options' key")
	}
}

func TestGenerateK6ThresholdsConfig_HighVolumeScenario(t *testing.T) {
	scenario := CreateHighVolumeReplicationScenario()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "high_volume_thresholds.json")

	err := GenerateK6ThresholdsConfig(scenario, outputPath)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)

	// High volume scenario should have replications_total threshold
	if !strings.Contains(contentStr, "replications_total") {
		t.Error("High volume scenario should have replications_total threshold")
	}
}

func TestGenerateK6ThresholdsConfig_LargeImageScenario(t *testing.T) {
	scenario := CreateLargeImageStressScenario()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "large_image_thresholds.json")

	err := GenerateK6ThresholdsConfig(scenario, outputPath)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)

	// Large image scenario should have memory_usage_mb threshold
	if !strings.Contains(contentStr, "memory_usage_mb") {
		t.Error("Large image scenario should have memory_usage_mb threshold")
	}
}

func TestGenerateK6ThresholdsConfig_NetworkResilienceScenario(t *testing.T) {
	scenario := CreateNetworkResilienceScenario()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "resilience_thresholds.json")

	err := GenerateK6ThresholdsConfig(scenario, outputPath)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)

	// Network resilience should have retry_success_rate threshold
	if !strings.Contains(contentStr, "retry_success_rate") {
		t.Error("Network resilience scenario should have retry_success_rate threshold")
	}
}

func TestGenerateK6ThresholdsConfig_BurstScenario(t *testing.T) {
	scenario := CreateBurstReplicationScenario()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "burst_thresholds.json")

	err := GenerateK6ThresholdsConfig(scenario, outputPath)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)

	// Burst scenario should have high http_reqs rate
	if !strings.Contains(contentStr, "http_reqs") {
		t.Error("Burst scenario should have http_reqs threshold")
	}
}

func TestGenerateK6ThresholdsConfig_SustainedScenario(t *testing.T) {
	scenario := CreateSustainedThroughputScenario()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "sustained_thresholds.json")

	err := GenerateK6ThresholdsConfig(scenario, outputPath)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)

	// Sustained throughput should have throughput_variance threshold
	if !strings.Contains(contentStr, "throughput_variance") {
		t.Error("Sustained scenario should have throughput_variance threshold")
	}
}

func TestGenerateK6ThresholdsConfig_InvalidPath(t *testing.T) {
	scenario := CreateHighVolumeReplicationScenario()

	// Invalid path (directory doesn't exist)
	invalidPath := "/nonexistent/path/thresholds.json"

	err := GenerateK6ThresholdsConfig(scenario, invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestK6ScriptData_Structure(t *testing.T) {
	scenario := CreateHighVolumeReplicationScenario()

	data := K6ScriptData{
		Scenario:           scenario,
		VUs:                10,
		Duration:           "1m",
		RampUpTime:         "10s",
		RampDownTime:       "10s",
		P95Threshold:       5000,
		FailureThreshold:   0.01,
		MinThroughput:      100.0,
		MaxFailureRate:     0.01,
		Images:             scenario.Images,
		TestServerAddr:     "localhost:8080",
		NetworkDelayFactor: 2.0,
		NetworkFailureRate: 0.001,
		MemoryPressureSize: 1000,
		OutputFile:         "/tmp/output.json",
	}

	if data.VUs != 10 {
		t.Errorf("Expected 10 VUs, got %d", data.VUs)
	}

	if data.Duration != "1m" {
		t.Errorf("Expected duration '1m', got '%s'", data.Duration)
	}

	if len(data.Images) != len(scenario.Images) {
		t.Errorf("Expected %d images, got %d", len(scenario.Images), len(data.Images))
	}

	if data.MinThroughput != 100.0 {
		t.Errorf("Expected min throughput 100.0, got %f", data.MinThroughput)
	}
}

func TestK6ScriptTemplate_Variables(t *testing.T) {
	// Verify that K6ScriptTemplate contains expected variables
	requiredVars := []string{
		"{{.RampUpTime}}",
		"{{.Duration}}",
		"{{.VUs}}",
		"{{.P95Threshold}}",
		"{{.FailureThreshold}}",
		"{{.MinThroughput}}",
		"{{.MaxFailureRate}}",
		"{{.TestServerAddr}}",
		"{{.NetworkDelayFactor}}",
		"{{.NetworkFailureRate}}",
		"{{.MemoryPressureSize}}",
		"{{.OutputFile}}",
	}

	for _, varName := range requiredVars {
		if !strings.Contains(K6ScriptTemplate, varName) {
			t.Errorf("K6ScriptTemplate missing variable: %s", varName)
		}
	}
}

func TestK6ScriptTemplate_ImageLoop(t *testing.T) {
	// Verify template has image loop
	if !strings.Contains(K6ScriptTemplate, "{{range .Images}}") {
		t.Error("K6ScriptTemplate should contain image range loop")
	}

	if !strings.Contains(K6ScriptTemplate, "{{.Repository}}") {
		t.Error("K6ScriptTemplate should reference image repository")
	}

	if !strings.Contains(K6ScriptTemplate, "{{.Tag}}") {
		t.Error("K6ScriptTemplate should reference image tag")
	}

	if !strings.Contains(K6ScriptTemplate, "{{.SizeMB}}") {
		t.Error("K6ScriptTemplate should reference image size")
	}
}

func TestK6ScriptTemplate_Metrics(t *testing.T) {
	// Verify custom metrics are defined
	requiredMetrics := []string{
		"replicationFailureRate",
		"replicationDuration",
		"replicationThroughput",
		"replicationCount",
	}

	for _, metric := range requiredMetrics {
		if !strings.Contains(K6ScriptTemplate, metric) {
			t.Errorf("K6ScriptTemplate missing metric: %s", metric)
		}
	}
}

func TestK6ScriptTemplate_Thresholds(t *testing.T) {
	// Verify thresholds are defined
	if !strings.Contains(K6ScriptTemplate, "thresholds") {
		t.Error("K6ScriptTemplate should contain thresholds")
	}

	requiredThresholds := []string{
		"http_req_duration",
		"http_req_failed",
		"replication_throughput_mbps",
		"replication_failures",
	}

	for _, threshold := range requiredThresholds {
		if !strings.Contains(K6ScriptTemplate, threshold) {
			t.Errorf("K6ScriptTemplate missing threshold: %s", threshold)
		}
	}
}

func TestK6ScriptTemplate_Functions(t *testing.T) {
	// Verify required functions exist
	requiredFunctions := []string{
		"simulateContainerReplication",
		"getRandomDestinationRegistry",
		"generateRandomSHA256",
		"generateRandomData",
		"simulateNetworkFailure",
		"simulateMemoryPressure",
	}

	for _, fn := range requiredFunctions {
		if !strings.Contains(K6ScriptTemplate, fn) {
			t.Errorf("K6ScriptTemplate missing function: %s", fn)
		}
	}
}

func TestK6ScriptTemplate_RegistryEndpoints(t *testing.T) {
	// Verify registry endpoints are defined
	registries := []string{"aws-ecr", "gcp-gcr", "docker-hub"}

	for _, registry := range registries {
		if !strings.Contains(K6ScriptTemplate, registry) {
			t.Errorf("K6ScriptTemplate missing registry: %s", registry)
		}
	}
}

func TestK6ScriptTemplate_ContainerWorkflow(t *testing.T) {
	// Verify container replication workflow steps
	workflowSteps := []string{
		"Check if image exists in destination",
		"Get manifest from source registry",
		"Simulate layer transfers",
		"Upload manifest to destination",
	}

	for _, step := range workflowSteps {
		if !strings.Contains(K6ScriptTemplate, step) {
			t.Errorf("K6ScriptTemplate missing workflow step comment: %s", step)
		}
	}
}

func TestLoadTestConfig_Validation(t *testing.T) {
	config := LoadTestConfig{
		ConcurrentJobs:     10,
		RepositoriesPerJob: 5,
		TestDuration:       30 * time.Second,
		ErrorRate:          0.05,
		MetricsInterval:    5 * time.Second,
	}

	if config.ConcurrentJobs <= 0 {
		t.Error("ConcurrentJobs should be positive")
	}

	if config.RepositoriesPerJob <= 0 {
		t.Error("RepositoriesPerJob should be positive")
	}

	if config.TestDuration <= 0 {
		t.Error("TestDuration should be positive")
	}

	if config.ErrorRate < 0 || config.ErrorRate > 1 {
		t.Error("ErrorRate should be between 0 and 1")
	}

	if config.MetricsInterval <= 0 {
		t.Error("MetricsInterval should be positive")
	}
}

func TestScenarioConfig_AllFieldsPopulated(t *testing.T) {
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
			if scenario.Name == "" {
				t.Error("Name should not be empty")
			}
			if scenario.Description == "" {
				t.Error("Description should not be empty")
			}
			if scenario.Duration <= 0 {
				t.Error("Duration should be positive")
			}
			if scenario.ConcurrentWorkers <= 0 {
				t.Error("ConcurrentWorkers should be positive")
			}
			if scenario.ExpectedThroughput <= 0 {
				t.Error("ExpectedThroughput should be positive")
			}
			if len(scenario.Images) == 0 {
				t.Error("Images should not be empty")
			}
		})
	}
}
