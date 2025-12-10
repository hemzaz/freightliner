package load

import (
	"testing"
	"time"

	"freightliner/pkg/helper/log"
)

func TestCreateHighVolumeReplicationScenario(t *testing.T) {
	scenario := CreateHighVolumeReplicationScenario()

	if scenario.Name == "" {
		t.Error("Expected non-empty scenario name")
	}

	if scenario.Scenario != HighVolumeReplication {
		t.Errorf("Expected HighVolumeReplication, got %v", scenario.Scenario)
	}

	if len(scenario.Images) < 1200 {
		t.Errorf("Expected at least 1200 images, got %d", len(scenario.Images))
	}

	if scenario.Duration != 2*time.Hour {
		t.Errorf("Expected 2 hour duration, got %v", scenario.Duration)
	}

	if scenario.ConcurrentWorkers != 50 {
		t.Errorf("Expected 50 concurrent workers, got %d", scenario.ConcurrentWorkers)
	}

	if scenario.ExpectedThroughput != 125.0 {
		t.Errorf("Expected throughput 125.0, got %f", scenario.ExpectedThroughput)
	}

	// Validate images
	for i, img := range scenario.Images {
		if img.Repository == "" {
			t.Errorf("Image %d has empty repository", i)
		}
		if img.Tag == "" {
			t.Errorf("Image %d has empty tag", i)
		}
		if img.SizeMB <= 0 {
			t.Errorf("Image %d has invalid size: %d", i, img.SizeMB)
		}
		if img.LayerCount <= 0 {
			t.Errorf("Image %d has invalid layer count: %d", i, img.LayerCount)
		}
		if img.Registry == "" {
			t.Errorf("Image %d has empty registry", i)
		}
	}

	// Validate criteria
	if scenario.ValidationCriteria.MinThroughputMBps <= 0 {
		t.Error("Expected positive min throughput")
	}
	if scenario.ValidationCriteria.MaxMemoryUsageMB <= 0 {
		t.Error("Expected positive max memory")
	}
}

func TestCreateLargeImageStressScenario(t *testing.T) {
	scenario := CreateLargeImageStressScenario()

	if scenario.Scenario != LargeImageStress {
		t.Errorf("Expected LargeImageStress, got %v", scenario.Scenario)
	}

	if len(scenario.Images) != 50 {
		t.Errorf("Expected 50 images, got %d", len(scenario.Images))
	}

	// Verify all images are large (>= 5GB)
	for i, img := range scenario.Images {
		if img.SizeMB < 5120 {
			t.Errorf("Image %d is not large enough: %d MB (expected >= 5120 MB)", i, img.SizeMB)
		}
		if img.LayerCount < 25 {
			t.Errorf("Image %d has too few layers: %d (expected >= 25)", i, img.LayerCount)
		}
	}

	if scenario.ConcurrentWorkers != 10 {
		t.Errorf("Expected 10 concurrent workers for large images, got %d", scenario.ConcurrentWorkers)
	}

	if scenario.ExpectedThroughput != 140.0 {
		t.Errorf("Expected throughput 140.0, got %f", scenario.ExpectedThroughput)
	}
}

func TestCreateNetworkResilienceScenario(t *testing.T) {
	scenario := CreateNetworkResilienceScenario()

	if scenario.Scenario != NetworkResilience {
		t.Errorf("Expected NetworkResilience, got %v", scenario.Scenario)
	}

	if len(scenario.Images) != 200 {
		t.Errorf("Expected 200 images, got %d", len(scenario.Images))
	}

	// Check network conditions
	if scenario.NetworkConditions.PacketLossRate != 0.05 {
		t.Errorf("Expected 5%% packet loss, got %f", scenario.NetworkConditions.PacketLossRate)
	}

	if scenario.NetworkConditions.LatencyMs != 100 {
		t.Errorf("Expected 100ms latency, got %d", scenario.NetworkConditions.LatencyMs)
	}

	// Check interruptions
	if len(scenario.NetworkConditions.ServiceInterruptions) != 3 {
		t.Errorf("Expected 3 interruptions, got %d", len(scenario.NetworkConditions.ServiceInterruptions))
	}

	// Verify higher failure tolerance for poor network conditions
	if scenario.MaxFailureRate != 0.05 {
		t.Errorf("Expected 5%% max failure rate, got %f", scenario.MaxFailureRate)
	}
}

func TestCreateBurstReplicationScenario(t *testing.T) {
	scenario := CreateBurstReplicationScenario()

	if scenario.Scenario != BurstReplication {
		t.Errorf("Expected BurstReplication, got %v", scenario.Scenario)
	}

	if len(scenario.Images) != 500 {
		t.Errorf("Expected 500 images, got %d", len(scenario.Images))
	}

	if scenario.ConcurrentWorkers != 100 {
		t.Errorf("Expected 100 concurrent workers for burst, got %d", scenario.ConcurrentWorkers)
	}

	if scenario.Duration != 30*time.Minute {
		t.Errorf("Expected 30 minute duration, got %v", scenario.Duration)
	}
}

func TestCreateSustainedThroughputScenario(t *testing.T) {
	scenario := CreateSustainedThroughputScenario()

	if scenario.Scenario != SustainedThroughput {
		t.Errorf("Expected SustainedThroughput, got %v", scenario.Scenario)
	}

	if len(scenario.Images) != 800 {
		t.Errorf("Expected 800 images, got %d", len(scenario.Images))
	}

	if scenario.Duration != 4*time.Hour {
		t.Errorf("Expected 4 hour duration for sustained test, got %v", scenario.Duration)
	}

	// Verify optimized sizes for sustained throughput (500MB-1.5GB)
	for i, img := range scenario.Images {
		if img.SizeMB < 500 || img.SizeMB > 1500 {
			t.Errorf("Image %d size %d MB outside optimal range (500-1500 MB)", i, img.SizeMB)
		}
	}

	// Higher efficiency requirements
	if scenario.ValidationCriteria.MinConnectionReuse != 0.90 {
		t.Errorf("Expected 90%% min connection reuse, got %f", scenario.ValidationCriteria.MinConnectionReuse)
	}
}

func TestCreateMixedContainerSizesScenario(t *testing.T) {
	scenario := CreateMixedContainerSizesScenario()

	if scenario.Scenario != MixedContainerSizes {
		t.Errorf("Expected MixedContainerSizes, got %v", scenario.Scenario)
	}

	if len(scenario.Images) != 600 {
		t.Errorf("Expected 600 images, got %d", len(scenario.Images))
	}

	// Count images in different size categories
	small := 0  // 10-100MB
	medium := 0 // 100-500MB
	large := 0  // 500MB-2GB
	xlarge := 0 // 2GB+

	for _, img := range scenario.Images {
		switch {
		case img.SizeMB >= 2500:
			xlarge++
		case img.SizeMB >= 600:
			large++
		case img.SizeMB >= 150:
			medium++
		default:
			small++
		}
	}

	// Verify distribution roughly matches expected (40%, 35%, 20%, 5%)
	// Allow some variance due to rounding
	totalImages := len(scenario.Images)
	if small < int(float64(totalImages)*0.35) || small > int(float64(totalImages)*0.45) {
		t.Errorf("Small images: %d (%.1f%%), expected ~40%%", small, float64(small)/float64(totalImages)*100)
	}

	t.Logf("Size distribution: small=%d (%.1f%%), medium=%d (%.1f%%), large=%d (%.1f%%), xlarge=%d (%.1f%%)",
		small, float64(small)/float64(totalImages)*100,
		medium, float64(medium)/float64(totalImages)*100,
		large, float64(large)/float64(totalImages)*100,
		xlarge, float64(xlarge)/float64(totalImages)*100)
}

func TestNewScenarioRunner(t *testing.T) {
	scenario := CreateHighVolumeReplicationScenario()
	runner := NewScenarioRunner(scenario, nil)

	if runner == nil {
		t.Fatal("Expected non-nil runner")
	}

	if runner.logger == nil {
		t.Error("Expected non-nil logger")
	}

	if runner.metrics == nil {
		t.Error("Expected non-nil metrics")
	}

	if runner.ctx == nil {
		t.Error("Expected non-nil context")
	}

	if runner.cancel == nil {
		t.Error("Expected non-nil cancel function")
	}
}

func TestConnectionStats_GetConnectionReuseRate(t *testing.T) {
	stats := &ConnectionStats{}

	// Initially should be 0
	rate := stats.GetConnectionReuseRate()
	if rate != 0 {
		t.Errorf("Expected 0 initial reuse rate, got %f", rate)
	}

	// Add some connections
	stats.TotalConnections.Store(100)
	stats.ReuseConnections.Store(80)

	rate = stats.GetConnectionReuseRate()
	if rate != 0.8 {
		t.Errorf("Expected 0.8 reuse rate, got %f", rate)
	}

	// Add more
	stats.TotalConnections.Store(200)
	stats.ReuseConnections.Store(180)

	rate = stats.GetConnectionReuseRate()
	if rate != 0.9 {
		t.Errorf("Expected 0.9 reuse rate, got %f", rate)
	}
}

func TestScenarioConfig_ValidationCriteria(t *testing.T) {
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
			// All scenarios should have positive validation criteria
			if scenario.ValidationCriteria.MinThroughputMBps <= 0 {
				t.Error("MinThroughputMBps must be positive")
			}
			if scenario.ValidationCriteria.MaxMemoryUsageMB <= 0 {
				t.Error("MaxMemoryUsageMB must be positive")
			}
			if scenario.ValidationCriteria.MaxFailureRate < 0 || scenario.ValidationCriteria.MaxFailureRate > 1 {
				t.Errorf("MaxFailureRate must be between 0 and 1, got %f", scenario.ValidationCriteria.MaxFailureRate)
			}
			if scenario.ValidationCriteria.MaxP99LatencyMs <= 0 {
				t.Error("MaxP99LatencyMs must be positive")
			}
			if scenario.ValidationCriteria.MinConnectionReuse < 0 || scenario.ValidationCriteria.MinConnectionReuse > 1 {
				t.Errorf("MinConnectionReuse must be between 0 and 1, got %f", scenario.ValidationCriteria.MinConnectionReuse)
			}
		})
	}
}

func TestScenarioConfig_NetworkConditions(t *testing.T) {
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
			// Validate network conditions
			if scenario.NetworkConditions.PacketLossRate < 0 || scenario.NetworkConditions.PacketLossRate > 1 {
				t.Errorf("PacketLossRate must be between 0 and 1, got %f", scenario.NetworkConditions.PacketLossRate)
			}
			if scenario.NetworkConditions.LatencyMs < 0 {
				t.Errorf("LatencyMs must be non-negative, got %d", scenario.NetworkConditions.LatencyMs)
			}
			if scenario.NetworkConditions.BandwidthLimitMBps <= 0 {
				t.Errorf("BandwidthLimitMBps must be positive, got %f", scenario.NetworkConditions.BandwidthLimitMBps)
			}
		})
	}
}

func TestScenarioTypes(t *testing.T) {
	types := []LoadTestScenario{
		HighVolumeReplication,
		LargeImageStress,
		NetworkResilience,
		BurstReplication,
		SustainedThroughput,
		MixedContainerSizes,
	}

	seen := make(map[LoadTestScenario]bool)
	for _, typ := range types {
		if seen[typ] {
			t.Errorf("Duplicate scenario type: %v", typ)
		}
		seen[typ] = true

		if string(typ) == "" {
			t.Errorf("Empty scenario type string")
		}
	}
}

func TestContainerImage_Validation(t *testing.T) {
	scenarios := []ScenarioConfig{
		CreateHighVolumeReplicationScenario(),
		CreateLargeImageStressScenario(),
		CreateNetworkResilienceScenario(),
	}

	for _, scenario := range scenarios {
		for i, img := range scenario.Images {
			if img.Repository == "" {
				t.Errorf("Scenario %s: Image %d missing repository", scenario.Name, i)
			}
			if img.Tag == "" {
				t.Errorf("Scenario %s: Image %d missing tag", scenario.Name, i)
			}
			if img.SizeMB <= 0 {
				t.Errorf("Scenario %s: Image %d has invalid size %d", scenario.Name, i, img.SizeMB)
			}
			if img.LayerCount <= 0 {
				t.Errorf("Scenario %s: Image %d has invalid layer count %d", scenario.Name, i, img.LayerCount)
			}
			if img.Registry == "" {
				t.Errorf("Scenario %s: Image %d missing registry", scenario.Name, i)
			}
		}
	}
}

func TestInterruption_Validation(t *testing.T) {
	scenario := CreateNetworkResilienceScenario()

	for i, intr := range scenario.NetworkConditions.ServiceInterruptions {
		if intr.StartTime < 0 {
			t.Errorf("Interruption %d has negative start time", i)
		}
		if intr.Duration <= 0 {
			t.Errorf("Interruption %d has non-positive duration", i)
		}
		if intr.Severity != "partial" && intr.Severity != "complete" {
			t.Errorf("Interruption %d has invalid severity: %s", i, intr.Severity)
		}
	}
}

func TestScenarioRunner_WithCustomLogger(t *testing.T) {
	scenario := CreateHighVolumeReplicationScenario()
	logger := log.NewBasicLogger(log.DebugLevel)

	runner := NewScenarioRunner(scenario, logger)

	if runner.logger != logger {
		t.Error("Expected custom logger to be used")
	}
}

func TestLoadTestScenario_Duration(t *testing.T) {
	tests := []struct {
		name             string
		scenario         ScenarioConfig
		expectedDuration time.Duration
	}{
		{
			name:             "High Volume",
			scenario:         CreateHighVolumeReplicationScenario(),
			expectedDuration: 2 * time.Hour,
		},
		{
			name:             "Large Image",
			scenario:         CreateLargeImageStressScenario(),
			expectedDuration: 90 * time.Minute,
		},
		{
			name:             "Network Resilience",
			scenario:         CreateNetworkResilienceScenario(),
			expectedDuration: 60 * time.Minute,
		},
		{
			name:             "Burst",
			scenario:         CreateBurstReplicationScenario(),
			expectedDuration: 30 * time.Minute,
		},
		{
			name:             "Sustained",
			scenario:         CreateSustainedThroughputScenario(),
			expectedDuration: 4 * time.Hour,
		},
		{
			name:             "Mixed Sizes",
			scenario:         CreateMixedContainerSizesScenario(),
			expectedDuration: 90 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.scenario.Duration != tt.expectedDuration {
				t.Errorf("Expected duration %v, got %v", tt.expectedDuration, tt.scenario.Duration)
			}
		})
	}
}
