package load

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// K6ScriptTemplate defines the template for k6 test scripts
const K6ScriptTemplate = `import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics for container registry operations
const replicationFailureRate = new Rate('replication_failures');
const replicationDuration = new Trend('replication_duration');
const replicationThroughput = new Trend('replication_throughput_mbps');
const replicationCount = new Counter('replications_total');

// Test configuration
export let options = {
  stages: [
    { duration: '{{.RampUpTime}}', target: {{.VUs}} },
    { duration: '{{.Duration}}', target: {{.VUs}} },
    { duration: '{{.RampDownTime}}', target: 0 },
  ],
  thresholds: {
    'http_req_duration': ['p(95)<{{.P95Threshold}}ms'],
    'http_req_failed': ['rate<{{.FailureThreshold}}'],
    'replication_throughput_mbps': ['avg>{{.MinThroughput}}'],
    'replication_failures': ['rate<{{.MaxFailureRate}}'],
  },
};

// Container image data for realistic testing
const containerImages = [
{{range .Images}}  {
    repository: '{{.Repository}}',
    tag: '{{.Tag}}',
    sizeMB: {{.SizeMB}},
    layerCount: {{.LayerCount}},
    registry: '{{.Registry}}'
  },
{{end}}];

// Simulate container registry endpoints
const registryEndpoints = {
  'aws-ecr': 'https://{{.TestServerAddr}}/v2/aws-ecr',
  'gcp-gcr': 'https://{{.TestServerAddr}}/v2/gcp-gcr',
  'docker-hub': 'https://{{.TestServerAddr}}/v2/docker-hub'
};

// Authentication tokens (simulated)
const authTokens = {
  'aws-ecr': 'AWS4-HMAC-SHA256-simulated-token',
  'gcp-gcr': 'Bearer gcp-service-account-token',
  'docker-hub': 'Bearer docker-hub-token'
};

export default function() {
  // Select random container image for this iteration
  const image = containerImages[Math.floor(Math.random() * containerImages.length)];
  const sourceRegistry = image.registry;
  const destRegistry = getRandomDestinationRegistry(sourceRegistry);
  
  const testStartTime = Date.now();
  
  // Simulate container replication workflow
  const replicationResult = simulateContainerReplication(image, sourceRegistry, destRegistry);
  
  const testEndTime = Date.now();
  const durationMs = testEndTime - testStartTime;
  
  // Record metrics
  replicationCount.add(1);
  replicationDuration.add(durationMs);
  
  if (replicationResult.success) {
    const throughputMBps = (image.sizeMB * 1000) / durationMs; // MB/s
    replicationThroughput.add(throughputMBps);
    replicationFailureRate.add(0);
  } else {
    replicationFailureRate.add(1);
  }
  
  // Brief pause between operations (realistic pacing)
  sleep(Math.random() * 2 + 0.5); // 0.5-2.5 seconds
}

function simulateContainerReplication(image, sourceRegistry, destRegistry) {
  const sourceEndpoint = registryEndpoints[sourceRegistry];
  const destEndpoint = registryEndpoints[destRegistry];
  const sourceAuth = authTokens[sourceRegistry];
  const destAuth = authTokens[destRegistry];
  
  // Step 1: Check if image exists in destination (HEAD request)
  const checkResponse = http.head(` + "`${destEndpoint}/${image.repository}/manifests/${image.tag}`" + `, {
    headers: {
      'Authorization': destAuth,
      'Accept': 'application/vnd.docker.distribution.manifest.v2+json'
    }
  });
  
  // If image already exists, simulate skip
  if (checkResponse.status === 200) {
    return { success: true, skipped: true, reason: 'already_exists' };
  }
  
  // Step 2: Get manifest from source registry
  const manifestResponse = http.get(` + "`${sourceEndpoint}/${image.repository}/manifests/${image.tag}`" + `, {
    headers: {
      'Authorization': sourceAuth,
      'Accept': 'application/vnd.docker.distribution.manifest.v2+json'
    }
  });
  
  const manifestSuccess = check(manifestResponse, {
    'manifest retrieved successfully': (r) => r.status === 200,
    'manifest has valid content-type': (r) => r.headers['Content-Type'].includes('application/vnd.docker.distribution.manifest'),
  });
  
  if (!manifestSuccess) {
    return { success: false, reason: 'manifest_fetch_failed', status: manifestResponse.status };
  }
  
  // Step 3: Simulate layer transfers (most time-consuming part)
  const layerTransferResults = [];
  
  for (let i = 0; i < image.layerCount; i++) {
    const layerSize = Math.floor(image.sizeMB / image.layerCount); // Average layer size
    const layerDigest = ` + "`sha256:${generateRandomSHA256()}`" + `;
    
    // Check if layer exists in destination
    const layerCheckResponse = http.head(` + "`${destEndpoint}/${image.repository}/blobs/${layerDigest}`" + `, {
      headers: { 'Authorization': destAuth }
    });
    
    if (layerCheckResponse.status === 200) {
      // Layer already exists, skip
      layerTransferResults.push({ success: true, skipped: true, size: layerSize });
      continue;
    }
    
    // Simulate layer download from source
    const layerDownloadResponse = http.get(` + "`${sourceEndpoint}/${image.repository}/blobs/${layerDigest}`" + `, {
      headers: { 'Authorization': sourceAuth }
    });
    
    if (layerDownloadResponse.status !== 200) {
      layerTransferResults.push({ success: false, reason: 'download_failed', size: layerSize });
      continue;
    }
    
    // Simulate layer upload to destination (POST to initiate, PUT to upload)
    const uploadInitResponse = http.post(` + "`${destEndpoint}/${image.repository}/blobs/uploads/`" + `, null, {
      headers: { 'Authorization': destAuth }
    });
    
    if (uploadInitResponse.status !== 202) {
      layerTransferResults.push({ success: false, reason: 'upload_init_failed', size: layerSize });
      continue;
    }
    
    const uploadLocation = uploadInitResponse.headers['Location'];
    
    // Simulate streaming upload with progress
    const uploadResponse = http.put(` + "`${uploadLocation}&digest=${layerDigest}`" + `, 
      generateRandomData(layerSize), {
      headers: {
        'Authorization': destAuth,
        'Content-Type': 'application/octet-stream',
        'Content-Length': layerSize.toString()
      }
    });
    
    const layerSuccess = check(uploadResponse, {
      'layer uploaded successfully': (r) => r.status === 201,
    });
    
    layerTransferResults.push({ 
      success: layerSuccess, 
      size: layerSize,
      uploadTime: uploadResponse.timings.duration 
    });
    
    // Simulate network conditions and processing delay
    const processingDelay = Math.max(50, layerSize * {{.NetworkDelayFactor}}); // ms
    sleep(processingDelay / 1000);
  }
  
  // Step 4: Upload manifest to destination
  const manifestUploadResponse = http.put(` + "`${destEndpoint}/${image.repository}/manifests/${image.tag}`" + `, 
    manifestResponse.body, {
    headers: {
      'Authorization': destAuth,
      'Content-Type': 'application/vnd.docker.distribution.manifest.v2+json'
    }
  });
  
  const manifestUploadSuccess = check(manifestUploadResponse, {
    'manifest uploaded successfully': (r) => r.status === 201,
  });
  
  // Determine overall success
  const failedLayers = layerTransferResults.filter(r => !r.success).length;
  const overallSuccess = manifestUploadSuccess && failedLayers === 0;
  
  return {
    success: overallSuccess,
    layersTransferred: layerTransferResults.length,
    layersFailed: failedLayers,
    manifestUploaded: manifestUploadSuccess,
    totalSize: image.sizeMB
  };
}

function getRandomDestinationRegistry(sourceRegistry) {
  const registries = Object.keys(registryEndpoints).filter(r => r !== sourceRegistry);
  return registries[Math.floor(Math.random() * registries.length)];
}

function generateRandomSHA256() {
  return Array.from({length: 64}, () => Math.floor(Math.random() * 16).toString(16)).join('');
}

function generateRandomData(sizeMB) {
  // Generate random data to simulate layer content
  // For k6, we'll use a string of appropriate length
  const sizeBytes = sizeMB * 1024 * 1024;
  const chunkSize = 1024; // 1KB chunks
  const chunks = Math.ceil(sizeBytes / chunkSize);
  
  let data = '';
  for (let i = 0; i < chunks; i++) {
    data += 'x'.repeat(Math.min(chunkSize, sizeBytes - i * chunkSize));
  }
  
  return data;
}

// Network failure simulation for resilience testing
function simulateNetworkFailure() {
  const failureRate = {{.NetworkFailureRate}};
  return Math.random() < failureRate;
}

// Memory pressure simulation
function simulateMemoryPressure() {
  // Simulate memory usage patterns during large image processing
  const largeArray = new Array({{.MemoryPressureSize}}).fill('memory-pressure-simulation');
  // Allow garbage collection
  sleep(0.1);
  return largeArray.length;
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    '{{.OutputFile}}': JSON.stringify(data),
  };
}
`

// K6ScriptData contains data for k6 script template
type K6ScriptData struct {
	Scenario           ScenarioConfig
	VUs                int
	Duration           string
	RampUpTime         string
	RampDownTime       string
	P95Threshold       int64
	FailureThreshold   float64
	MinThroughput      float64
	MaxFailureRate     float64
	Images             []ContainerImage
	TestServerAddr     string
	NetworkDelayFactor float64
	NetworkFailureRate float64
	MemoryPressureSize int
	OutputFile         string
}

// generateK6Scripts creates k6 JavaScript test scripts for each scenario
func (bs *BenchmarkSuite) generateK6Scripts(scenarios []ScenarioConfig) error {
	// Ensure k6 scripts directory exists
	if err := os.MkdirAll(bs.k6Config.ScriptPath, 0755); err != nil {
		return fmt.Errorf("failed to create k6 scripts directory: %w", err)
	}

	tmpl, err := template.New("k6script").Parse(K6ScriptTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse k6 script template: %w", err)
	}

	for _, scenario := range scenarios {
		scriptData := K6ScriptData{
			Scenario:           scenario,
			VUs:                bs.k6Config.VUs,
			Duration:           bs.k6Config.Duration.String(),
			RampUpTime:         bs.k6Config.RampUpTime.String(),
			RampDownTime:       bs.k6Config.RampDownTime.String(),
			P95Threshold:       int64(scenario.ValidationCriteria.MaxP99LatencyMs),
			FailureThreshold:   scenario.ValidationCriteria.MaxFailureRate,
			MinThroughput:      scenario.ValidationCriteria.MinThroughputMBps,
			MaxFailureRate:     scenario.ValidationCriteria.MaxFailureRate,
			Images:             scenario.Images,
			TestServerAddr:     "localhost:8080", // Default test server
			NetworkDelayFactor: 2.0,              // 2ms per MB baseline
			NetworkFailureRate: scenario.NetworkConditions.PacketLossRate,
			MemoryPressureSize: 1000, // Array size for memory simulation
			OutputFile: filepath.Join(bs.resultsDir, fmt.Sprintf("k6_%s.json",
				strings.ReplaceAll(scenario.Name, " ", "_"))),
		}

		// Adjust network conditions based on scenario
		switch scenario.Scenario {
		case NetworkResilience:
			scriptData.NetworkDelayFactor = 5.0 // Higher delay for resilience testing
			scriptData.NetworkFailureRate = scenario.NetworkConditions.PacketLossRate
		case LargeImageStress:
			scriptData.MemoryPressureSize = 10000 // More memory pressure
			scriptData.NetworkDelayFactor = 1.5   // Optimized for large images
		case BurstReplication:
			scriptData.NetworkDelayFactor = 1.0 // Minimal delay for burst testing
		case SustainedThroughput:
			scriptData.NetworkDelayFactor = 1.2 // Balanced for sustained operations
		}

		scriptFile := filepath.Join(bs.k6Config.ScriptPath,
			fmt.Sprintf("%s.js", strings.ReplaceAll(scenario.Name, " ", "_")))

		file, err := os.Create(scriptFile)
		if err != nil {
			return fmt.Errorf("failed to create k6 script file %s: %w", scriptFile, err)
		}
		defer func() { _ = file.Close() }()

		if err := tmpl.Execute(file, scriptData); err != nil {
			return fmt.Errorf("failed to execute k6 script template for %s: %w", scenario.Name, err)
		}

		bs.logger.WithFields(map[string]interface{}{
			"scenario": scenario.Name,
			"file":     scriptFile,
			"images":   len(scenario.Images),
		}).Info("Generated k6 script")
	}

	// Generate a comprehensive test script that runs all scenarios
	if err := bs.generateComprehensiveK6Script(scenarios); err != nil {
		return fmt.Errorf("failed to generate comprehensive k6 script: %w", err)
	}

	return nil
}

// generateComprehensiveK6Script creates a single k6 script that runs all scenarios
func (bs *BenchmarkSuite) generateComprehensiveK6Script(scenarios []ScenarioConfig) error {
	const comprehensiveTemplate = `import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Comprehensive metrics across all scenarios
const overallFailureRate = new Rate('overall_failures');
const overallThroughput = new Trend('overall_throughput_mbps');
const scenarioResults = new Counter('scenario_completions');

export let options = {
  stages: [
    { duration: '1m', target: 20 },
    { duration: '5m', target: 50 },
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    'overall_throughput_mbps': ['avg>100'],
    'overall_failures': ['rate<0.01'],
    'http_req_duration': ['p(99)<5000'],
  },
};

const scenarios = [
{{range .Scenarios}}  {
    name: '{{.Name}}',
    images: {{.ImageCount}},
    expectedThroughput: {{.ExpectedThroughput}},
    maxFailureRate: {{.MaxFailureRate}}
  },
{{end}}];

export default function() {
  // Run each scenario proportionally
  const scenarioIndex = Math.floor(Math.random() * scenarios.length);
  const scenario = scenarios[scenarioIndex];
  
  group(scenario.name, function() {
    // Execute scenario-specific logic
    const result = executeScenario(scenario);
    
    if (result.success) {
      overallThroughput.add(result.throughput);
      overallFailureRate.add(0);
    } else {
      overallFailureRate.add(1);
    }
    
    scenarioResults.add(1);
  });
  
  sleep(1);
}

function executeScenario(scenario) {
  // Simplified scenario execution for comprehensive testing
  const startTime = Date.now();
  
  // Simulate realistic container registry operations
  const response = http.get('http://localhost:8080/replicate', {
    headers: {
      'X-Scenario': scenario.name,
      'Content-Type': 'application/json'
    }
  });
  
  const success = check(response, {
    'status is 200': (r) => r.status === 200,
  });
  
  const endTime = Date.now();
  const duration = endTime - startTime;
  const throughput = success ? (100 * 1000 / duration) : 0; // Simulated throughput
  
  return {
    success: success,
    throughput: throughput,
    duration: duration
  };
}
`

	type ComprehensiveScriptData struct {
		Scenarios []struct {
			Name               string
			ImageCount         int
			ExpectedThroughput float64
			MaxFailureRate     float64
		}
	}

	scriptData := ComprehensiveScriptData{}
	for _, scenario := range scenarios {
		scriptData.Scenarios = append(scriptData.Scenarios, struct {
			Name               string
			ImageCount         int
			ExpectedThroughput float64
			MaxFailureRate     float64
		}{
			Name:               scenario.Name,
			ImageCount:         len(scenario.Images),
			ExpectedThroughput: scenario.ExpectedThroughput,
			MaxFailureRate:     scenario.MaxFailureRate,
		})
	}

	tmpl, err := template.New("comprehensive").Parse(comprehensiveTemplate)
	if err != nil {
		return err
	}

	scriptFile := filepath.Join(bs.k6Config.ScriptPath, "comprehensive_test.js")
	file, err := os.Create(scriptFile)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	return tmpl.Execute(file, scriptData)
}

// K6ThresholdsConfig generates k6 thresholds configuration
func GenerateK6ThresholdsConfig(scenario ScenarioConfig, outputPath string) error {
	thresholds := map[string][]string{
		"http_req_duration":           {fmt.Sprintf("p(95)<%.0fms", float64(scenario.ValidationCriteria.MaxP99LatencyMs)*0.95)},
		"http_req_failed":             {fmt.Sprintf("rate<%.3f", scenario.ValidationCriteria.MaxFailureRate)},
		"replication_throughput_mbps": {fmt.Sprintf("avg>%.1f", scenario.ValidationCriteria.MinThroughputMBps)},
		"replication_failures":        {fmt.Sprintf("rate<%.3f", scenario.ValidationCriteria.MaxFailureRate)},
	}

	// Add scenario-specific thresholds
	switch scenario.Scenario {
	case HighVolumeReplication:
		thresholds["replications_total"] = []string{"count>1000"}
		thresholds["http_reqs"] = []string{"rate>100"} // requests per second
	case LargeImageStress:
		thresholds["replication_throughput_mbps"] = []string{fmt.Sprintf("avg>%.1f", scenario.ValidationCriteria.MinThroughputMBps)}
		thresholds["memory_usage_mb"] = []string{fmt.Sprintf("avg<%.0f", float64(scenario.ValidationCriteria.MaxMemoryUsageMB))}
	case NetworkResilience:
		thresholds["http_req_failed"] = []string{fmt.Sprintf("rate<%.3f", scenario.ValidationCriteria.MaxFailureRate*5)} // Higher tolerance for network issues
		thresholds["retry_success_rate"] = []string{"rate>0.95"}                                                         // 95% retry success
	case BurstReplication:
		thresholds["http_reqs"] = []string{"rate>200"} // High request rate for bursts
		thresholds["replication_throughput_mbps"] = []string{fmt.Sprintf("p(95)>%.1f", scenario.ValidationCriteria.MinThroughputMBps*1.1)}
	case SustainedThroughput:
		thresholds["replication_throughput_mbps"] = []string{
			fmt.Sprintf("avg>%.1f", scenario.ValidationCriteria.MinThroughputMBps),
			fmt.Sprintf("p(90)>%.1f", scenario.ValidationCriteria.MinThroughputMBps*0.9),
		}
		thresholds["throughput_variance"] = []string{"p(95)<10"} // Low variance for sustained performance
	}

	// Write thresholds to JSON file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	data := map[string]interface{}{
		"thresholds": thresholds,
		"options": map[string]interface{}{
			"scenarios": map[string]interface{}{
				scenario.Name: map[string]interface{}{
					"executor": "ramping-vus",
					"stages": []map[string]interface{}{
						{"duration": "30s", "target": scenario.ConcurrentWorkers / 2},
						{"duration": scenario.Duration.String(), "target": scenario.ConcurrentWorkers},
						{"duration": "30s", "target": 0},
					},
				},
			},
		},
	}

	return json.NewEncoder(file).Encode(data)
}
