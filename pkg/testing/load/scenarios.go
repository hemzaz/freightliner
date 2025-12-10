package load

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/metrics"
)

// LoadTestScenario defines different types of load testing scenarios
type LoadTestScenario string

const (
	// High-volume container replication with 1000+ repositories
	HighVolumeReplication LoadTestScenario = "high_volume_replication"

	// Large image stress test with 50 repositories containing 5GB+ images
	LargeImageStress LoadTestScenario = "large_image_stress"

	// Network resilience testing with intermittent failures
	NetworkResilience LoadTestScenario = "network_resilience"

	// Burst replication loads for sudden high-volume transfers
	BurstReplication LoadTestScenario = "burst_replication"

	// Sustained high-throughput operations for continuous replication
	SustainedThroughput LoadTestScenario = "sustained_throughput"

	// Mixed container size distributions for realistic production patterns
	MixedContainerSizes LoadTestScenario = "mixed_container_sizes"
)

// ContainerImage represents a container image for load testing
type ContainerImage struct {
	Repository string
	Tag        string
	SizeMB     int64  // Size in megabytes
	LayerCount int    // Number of layers
	Registry   string // Source registry (aws-ecr, gcp-gcr, etc.)
}

// ScenarioConfig defines configuration for specific load test scenarios
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
	MinThroughputMBps  float64 // Minimum sustained throughput
	MaxMemoryUsageMB   int64   // Maximum memory usage
	MaxFailureRate     float64 // Maximum failure rate
	MaxP99LatencyMs    int64   // Maximum 99th percentile latency
	MinConnectionReuse float64 // Minimum connection reuse rate
}

// ScenarioRunner executes specific load test scenarios
type ScenarioRunner struct {
	config           ScenarioConfig
	metrics          *LoadTestMetrics
	logger           log.Logger
	metricsCollector metrics.MetricsCollector

	// Performance tracking
	throughputMBps  atomic.Int64 // Current throughput in KB/s (multiply by 1000 for MB/s)
	memoryUsageMB   atomic.Int64 // Current memory usage in MB
	connectionStats ConnectionStats

	// Scenario control
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	errorRate atomic.Int64 // Error rate per 10000 (for precision)
}

// ConnectionStats tracks connection pool efficiency
type ConnectionStats struct {
	TotalConnections  atomic.Int64
	ReuseConnections  atomic.Int64
	NewConnections    atomic.Int64
	FailedConnections atomic.Int64
}

// GetConnectionReuseRate returns the connection reuse rate
func (cs *ConnectionStats) GetConnectionReuseRate() float64 {
	total := cs.TotalConnections.Load()
	if total == 0 {
		return 0
	}
	reused := cs.ReuseConnections.Load()
	return float64(reused) / float64(total)
}

// NewScenarioRunner creates a new scenario runner
func NewScenarioRunner(config ScenarioConfig, logger log.Logger) *ScenarioRunner {
	if logger == nil {
		logger = log.NewLoggerWithLevel(log.InfoLevel)
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)

	return &ScenarioRunner{
		config:           config,
		metrics:          NewLoadTestMetrics(),
		logger:           logger,
		metricsCollector: metrics.NewPrometheusMetrics(),
		ctx:              ctx,
		cancel:           cancel,
	}
}

// CreateHighVolumeReplicationScenario creates a high-volume replication scenario
func CreateHighVolumeReplicationScenario() ScenarioConfig {
	// Generate 1000+ realistic container images
	images := make([]ContainerImage, 1200)
	registries := []string{"aws-ecr", "gcp-gcr", "docker-hub"}

	for i := 0; i < 1200; i++ {
		sizeVariants := []int64{10, 25, 50, 100, 200, 500, 1024, 2048} // Various sizes in MB
		layerVariants := []int{3, 5, 8, 12, 15, 20}

		images[i] = ContainerImage{
			Repository: fmt.Sprintf("org-%d/app-%d", i/50, i%50),
			Tag:        fmt.Sprintf("v1.%d.%d", i/100, i%100),
			SizeMB:     sizeVariants[i%len(sizeVariants)],
			LayerCount: layerVariants[i%len(layerVariants)],
			Registry:   registries[i%len(registries)],
		}
	}

	return ScenarioConfig{
		Name:               "High Volume Container Replication",
		Description:        "Validates 1000+ repository replication with mixed sizes and concurrent multi-cloud operations",
		Scenario:           HighVolumeReplication,
		Duration:           2 * time.Hour,
		ConcurrentWorkers:  50,
		ExpectedThroughput: 125.0, // Target 125 MB/s
		MemoryLimitMB:      1024,  // 1GB limit
		MaxFailureRate:     0.01,  // 1% failure rate
		Images:             images,
		NetworkConditions: NetworkConditions{
			PacketLossRate:     0.001, // 0.1% packet loss
			LatencyMs:          50,
			BandwidthLimitMBps: 150.0,
		},
		ValidationCriteria: ValidationCriteria{
			MinThroughputMBps:  100.0, // Minimum 100 MB/s
			MaxMemoryUsageMB:   1000,  // Max 1GB
			MaxFailureRate:     0.01,  // Max 1% failures
			MaxP99LatencyMs:    5000,  // Max 5s p99 latency
			MinConnectionReuse: 0.80,  // Min 80% connection reuse
		},
	}
}

// CreateLargeImageStressScenario creates a large image stress testing scenario
func CreateLargeImageStressScenario() ScenarioConfig {
	// Generate 50 repositories with 5GB+ images
	images := make([]ContainerImage, 50)
	registries := []string{"aws-ecr", "gcp-gcr"}

	for i := 0; i < 50; i++ {
		// Large container sizes: 5GB to 10GB
		largeSizes := []int64{5120, 6144, 7168, 8192, 9216, 10240} // 5GB to 10GB in MB
		layerCounts := []int{25, 30, 35, 40, 45, 50}               // More layers for large images

		images[i] = ContainerImage{
			Repository: fmt.Sprintf("big-data/processing-%d", i),
			Tag:        fmt.Sprintf("v2.%d", i),
			SizeMB:     largeSizes[i%len(largeSizes)],
			LayerCount: layerCounts[i%len(layerCounts)],
			Registry:   registries[i%len(registries)],
		}
	}

	return ScenarioConfig{
		Name:               "Large Image Stress Test",
		Description:        "Validates memory efficiency and network saturation with 50 repositories containing 5GB+ images",
		Scenario:           LargeImageStress,
		Duration:           90 * time.Minute,
		ConcurrentWorkers:  10,    // Lower concurrency due to large images
		ExpectedThroughput: 140.0, // Higher throughput expected with large images
		MemoryLimitMB:      1024,  // 1GB limit despite large images
		MaxFailureRate:     0.005, // 0.5% failure rate
		Images:             images,
		NetworkConditions: NetworkConditions{
			PacketLossRate:     0.002, // 0.2% packet loss
			LatencyMs:          30,
			BandwidthLimitMBps: 150.0,
		},
		ValidationCriteria: ValidationCriteria{
			MinThroughputMBps:  120.0, // Higher minimum for large images
			MaxMemoryUsageMB:   1000,  // Strict memory limit
			MaxFailureRate:     0.005, // Lower failure tolerance
			MaxP99LatencyMs:    10000, // 10s for large images
			MinConnectionReuse: 0.85,  // Higher reuse expected
		},
	}
}

// CreateNetworkResilienceScenario creates a network resilience testing scenario
func CreateNetworkResilienceScenario() ScenarioConfig {
	// Mixed size containers for resilience testing
	images := make([]ContainerImage, 200)
	registries := []string{"aws-ecr", "gcp-gcr", "docker-hub"}

	for i := 0; i < 200; i++ {
		mixedSizes := []int64{50, 150, 300, 800, 1500, 2500} // Various sizes
		images[i] = ContainerImage{
			Repository: fmt.Sprintf("resilient/app-%d", i),
			Tag:        fmt.Sprintf("v1.%d", i%10),
			SizeMB:     mixedSizes[i%len(mixedSizes)],
			LayerCount: 5 + (i % 15), // 5-20 layers
			Registry:   registries[i%len(registries)],
		}
	}

	// Simulate various network interruptions
	interruptions := []Interruption{
		{StartTime: 10 * time.Minute, Duration: 30 * time.Second, Severity: "partial"},
		{StartTime: 25 * time.Minute, Duration: 60 * time.Second, Severity: "complete"},
		{StartTime: 45 * time.Minute, Duration: 45 * time.Second, Severity: "partial"},
	}

	return ScenarioConfig{
		Name:               "Network Resilience Test",
		Description:        "Validates retry mechanisms and resilience with 5% packet loss and service interruptions",
		Scenario:           NetworkResilience,
		Duration:           60 * time.Minute,
		ConcurrentWorkers:  25,
		ExpectedThroughput: 90.0, // Lower due to network conditions
		MemoryLimitMB:      800,
		MaxFailureRate:     0.05, // 5% acceptable with poor conditions
		Images:             images,
		NetworkConditions: NetworkConditions{
			PacketLossRate:       0.05, // 5% packet loss
			LatencyMs:            100,  // High latency
			BandwidthLimitMBps:   120.0,
			ServiceInterruptions: interruptions,
		},
		ValidationCriteria: ValidationCriteria{
			MinThroughputMBps:  70.0, // Lower expectation due to conditions
			MaxMemoryUsageMB:   800,
			MaxFailureRate:     0.01,  // Still expect <1% final failure rate
			MaxP99LatencyMs:    15000, // 15s with retries
			MinConnectionReuse: 0.75,  // Lower due to connection drops
		},
	}
}

// CreateBurstReplicationScenario creates a burst replication scenario
func CreateBurstReplicationScenario() ScenarioConfig {
	// Create images that will be submitted in bursts
	images := make([]ContainerImage, 500)
	for i := 0; i < 500; i++ {
		images[i] = ContainerImage{
			Repository: fmt.Sprintf("burst/repo-%d", i),
			Tag:        "latest",
			SizeMB:     200 + int64(i%1000), // 200MB-1.2GB range
			LayerCount: 8 + (i % 10),        // 8-18 layers
			Registry:   []string{"aws-ecr", "gcp-gcr"}[i%2],
		}
	}

	return ScenarioConfig{
		Name:               "Burst Replication Load",
		Description:        "Validates system response to sudden high-volume transfer bursts",
		Scenario:           BurstReplication,
		Duration:           30 * time.Minute,
		ConcurrentWorkers:  100, // High concurrency for burst
		ExpectedThroughput: 130.0,
		MemoryLimitMB:      1200,
		MaxFailureRate:     0.02, // 2% acceptable during bursts
		Images:             images,
		NetworkConditions: NetworkConditions{
			PacketLossRate:     0.001,
			LatencyMs:          40,
			BandwidthLimitMBps: 150.0,
		},
		ValidationCriteria: ValidationCriteria{
			MinThroughputMBps:  110.0,
			MaxMemoryUsageMB:   1200,
			MaxFailureRate:     0.02,
			MaxP99LatencyMs:    3000,
			MinConnectionReuse: 0.70, // Lower due to burst nature
		},
	}
}

// CreateSustainedThroughputScenario creates a sustained throughput scenario
func CreateSustainedThroughputScenario() ScenarioConfig {
	// Optimized mix for sustained throughput
	images := make([]ContainerImage, 800)
	for i := 0; i < 800; i++ {
		// Optimal sizes for sustained throughput
		optimalSizes := []int64{500, 750, 1000, 1250, 1500} // 500MB-1.5GB
		images[i] = ContainerImage{
			Repository: fmt.Sprintf("sustained/app-%d", i/10),
			Tag:        fmt.Sprintf("v%d.%d", i/100, i%10),
			SizeMB:     optimalSizes[i%len(optimalSizes)],
			LayerCount: 10 + (i % 8), // 10-18 layers
			Registry:   []string{"aws-ecr", "gcp-gcr"}[i%2],
		}
	}

	return ScenarioConfig{
		Name:               "Sustained High-Throughput",
		Description:        "Validates continuous replication at 100-150 MB/s for extended periods",
		Scenario:           SustainedThroughput,
		Duration:           4 * time.Hour, // Extended duration
		ConcurrentWorkers:  40,
		ExpectedThroughput: 135.0,
		MemoryLimitMB:      900,
		MaxFailureRate:     0.005, // Very low tolerance for sustained operations
		Images:             images,
		NetworkConditions: NetworkConditions{
			PacketLossRate:     0.0005, // Very low packet loss
			LatencyMs:          25,     // Low latency
			BandwidthLimitMBps: 160.0,  // High bandwidth ceiling
		},
		ValidationCriteria: ValidationCriteria{
			MinThroughputMBps:  120.0, // High sustained minimum
			MaxMemoryUsageMB:   900,
			MaxFailureRate:     0.005,
			MaxP99LatencyMs:    2000, // Low latency requirement
			MinConnectionReuse: 0.90, // High efficiency expected
		},
	}
}

// CreateMixedContainerSizesScenario creates a realistic mixed container sizes scenario
func CreateMixedContainerSizesScenario() ScenarioConfig {
	// Realistic distribution based on production patterns
	images := make([]ContainerImage, 600)

	// Production-like size distribution:
	// 40% small (10-100MB), 35% medium (100-500MB), 20% large (500MB-2GB), 5% extra-large (2GB+)
	sizeDistribution := []struct {
		percentage int
		sizeRange  []int64
	}{
		{40, []int64{10, 25, 50, 75, 100}},         // Small
		{35, []int64{150, 200, 300, 400, 500}},     // Medium
		{20, []int64{600, 800, 1200, 1600, 2000}},  // Large
		{5, []int64{2500, 3000, 4000, 5000, 6000}}, // Extra-large
	}

	imageIndex := 0
	for _, dist := range sizeDistribution {
		count := (600 * dist.percentage) / 100
		for i := 0; i < count && imageIndex < 600; i++ {
			size := dist.sizeRange[i%len(dist.sizeRange)]
			layerCount := 5 + int(size/200) // More layers for larger images
			if layerCount > 50 {
				layerCount = 50
			}

			images[imageIndex] = ContainerImage{
				Repository: fmt.Sprintf("mixed/category-%d/app-%d",
					imageIndex/(600/4), imageIndex%(600/4)),
				Tag:        fmt.Sprintf("v1.%d", imageIndex%20),
				SizeMB:     size,
				LayerCount: layerCount,
				Registry:   []string{"aws-ecr", "gcp-gcr", "docker-hub"}[imageIndex%3],
			}
			imageIndex++
		}
	}

	return ScenarioConfig{
		Name:               "Mixed Container Sizes",
		Description:        "Validates performance with realistic production-like container size distributions",
		Scenario:           MixedContainerSizes,
		Duration:           90 * time.Minute,
		ConcurrentWorkers:  35,
		ExpectedThroughput: 115.0,
		MemoryLimitMB:      1000,
		MaxFailureRate:     0.01,
		Images:             images,
		NetworkConditions: NetworkConditions{
			PacketLossRate:     0.002, // 0.2% packet loss
			LatencyMs:          45,
			BandwidthLimitMBps: 140.0,
		},
		ValidationCriteria: ValidationCriteria{
			MinThroughputMBps:  100.0,
			MaxMemoryUsageMB:   1000,
			MaxFailureRate:     0.01,
			MaxP99LatencyMs:    4000,
			MinConnectionReuse: 0.80,
		},
	}
}

// Run executes the load test scenario
func (sr *ScenarioRunner) Run() (*LoadTestResults, error) {
	sr.logger.Info("Starting load test scenario", map[string]interface{}{
		"scenario":            sr.config.Name,
		"duration":            sr.config.Duration.String(),
		"concurrent_workers":  sr.config.ConcurrentWorkers,
		"expected_throughput": sr.config.ExpectedThroughput,
		"total_images":        len(sr.config.Images),
	})

	// Initialize metrics collection
	sr.metrics.StartTime = time.Now()

	// Start scenario-specific workers
	switch sr.config.Scenario {
	case HighVolumeReplication:
		return sr.runHighVolumeReplication()
	case LargeImageStress:
		return sr.runLargeImageStress()
	case NetworkResilience:
		return sr.runNetworkResilience()
	case BurstReplication:
		return sr.runBurstReplication()
	case SustainedThroughput:
		return sr.runSustainedThroughput()
	case MixedContainerSizes:
		return sr.runMixedContainerSizes()
	default:
		return nil, fmt.Errorf("unknown scenario: %s", sr.config.Scenario)
	}
}

// LoadTestResults contains the results of a load test scenario
type LoadTestResults struct {
	ScenarioName          string
	Duration              time.Duration
	TotalImages           int64
	ProcessedImages       int64
	FailedImages          int64
	AverageThroughputMBps float64
	PeakThroughputMBps    float64
	MemoryUsageMB         int64
	ConnectionReuseRate   float64
	FailureRate           float64
	P99LatencyMs          int64
	ValidationPassed      bool
	ValidationErrors      []string
	DetailedMetrics       map[string]interface{}
}

// Helper methods for specific scenario implementations will be added in separate files
// to keep this file focused on scenario definitions and configuration
