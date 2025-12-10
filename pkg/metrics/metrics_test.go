package metrics

import (
	"testing"
	"time"
)

func TestNoopMetrics(t *testing.T) {
	// NoopMetrics should implement all methods without error
	metrics := &NoopMetrics{}

	// Test no panic on method calls
	metrics.ReplicationStarted("source", "dest")
	metrics.ReplicationCompleted(1*time.Second, 5, 1024)
	metrics.ReplicationFailed()

	// No assertions needed as these are no-op operations
}

func TestPrometheusMetricsReplicationStarted(t *testing.T) {
	metrics := NewPrometheusMetrics()

	// Track several replications
	metrics.ReplicationStarted("ecr.io/project/repo1", "gcr.io/project/repo1")
	metrics.ReplicationStarted("ecr.io/project/repo2", "gcr.io/project/repo2")
	metrics.ReplicationStarted("ecr.io/project/repo1", "gcr.io/project/repo3")
	metrics.ReplicationStarted("ecr.io/project/repo3", "gcr.io/project/repo1")

	// Test source repository counts
	sourceRepos := metrics.GetTopSourceRepositories(10)
	if len(sourceRepos) != 3 {
		t.Errorf("Expected 3 source repositories, got %d", len(sourceRepos))
	}

	// repo1 should have count 2
	for repo, count := range sourceRepos {
		if repo == "ecr.io/project/repo1" && count != 2 {
			t.Errorf("Expected repo1 to have count 2, got %d", count)
		}
	}

	// Test destination repository counts
	destRepos := metrics.GetTopDestinationRepositories(10)
	if len(destRepos) != 3 {
		t.Errorf("Expected 3 destination repositories, got %d", len(destRepos))
	}

	// repo1 should have count 2
	for repo, count := range destRepos {
		if repo == "gcr.io/project/repo1" && count != 2 {
			t.Errorf("Expected repo1 to have count 2, got %d", count)
		}
	}
}

func TestPrometheusMetricsReplicationCompleted(t *testing.T) {
	metrics := NewPrometheusMetrics()

	// Complete several replications
	metrics.ReplicationCompleted(1*time.Second, 5, 1024)
	metrics.ReplicationCompleted(2*time.Second, 3, 2048)
	metrics.ReplicationCompleted(3*time.Second, 7, 4096)

	// Test layers copied
	if layers := metrics.GetLayersCopied(); layers != 15 {
		t.Errorf("Expected 15 layers copied, got %d", layers)
	}

	// Test bytes copied
	if bytes := metrics.GetBytesCopied(); bytes != 7168 {
		t.Errorf("Expected 7168 bytes copied, got %d", bytes)
	}

	// Test average latency (1 + 2 + 3) / 3 = 2 seconds
	if latency := metrics.GetAverageLatency(); latency != 2*time.Second {
		t.Errorf("Expected average latency of 2s, got %v", latency)
	}
}

func TestPrometheusMetricsReplicationFailed(t *testing.T) {
	metrics := NewPrometheusMetrics()

	// Track replication starts
	metrics.ReplicationStarted("source", "dest")
	metrics.ReplicationStarted("source", "dest")
	metrics.ReplicationStarted("source", "dest")

	// Record some failures
	metrics.ReplicationFailed()
	metrics.ReplicationFailed()

	// Complete a successful replication
	metrics.ReplicationCompleted(1*time.Second, 5, 1024)

	// Test error count
	if errors := metrics.GetReplicationErrors(); errors != 2 {
		t.Errorf("Expected error count of 2, got %d", errors)
	}

	// Test replication count (based on ReplicationStarted calls)
	if count := metrics.GetReplicationCount(); count != 3 {
		t.Errorf("Expected replication count of 3, got %d", count)
	}
}

func TestPrometheusMetricsRepositoryCounts(t *testing.T) {
	metrics := NewPrometheusMetrics()

	// Record replications with different sources and destinations
	for i := 0; i < 5; i++ {
		metrics.ReplicationStarted("ecr.io/project/repo1", "gcr.io/project/dest1")
	}
	for i := 0; i < 3; i++ {
		metrics.ReplicationStarted("ecr.io/project/repo2", "gcr.io/project/dest2")
	}
	for i := 0; i < 7; i++ {
		metrics.ReplicationStarted("ecr.io/project/repo3", "gcr.io/project/dest3")
	}

	// Test source repositories
	sourceRepos := metrics.GetTopSourceRepositories(10)
	if len(sourceRepos) != 3 {
		t.Errorf("Expected 3 source repositories, got %d", len(sourceRepos))
	}

	if sourceRepos["ecr.io/project/repo1"] != 5 {
		t.Errorf("Expected repo1 to have count 5, got %d", sourceRepos["ecr.io/project/repo1"])
	}

	if sourceRepos["ecr.io/project/repo2"] != 3 {
		t.Errorf("Expected repo2 to have count 3, got %d", sourceRepos["ecr.io/project/repo2"])
	}

	if sourceRepos["ecr.io/project/repo3"] != 7 {
		t.Errorf("Expected repo3 to have count 7, got %d", sourceRepos["ecr.io/project/repo3"])
	}

	// Test destination repositories
	destRepos := metrics.GetTopDestinationRepositories(10)
	if len(destRepos) != 3 {
		t.Errorf("Expected 3 destination repositories, got %d", len(destRepos))
	}

	if destRepos["gcr.io/project/dest1"] != 5 {
		t.Errorf("Expected dest1 to have count 5, got %d", destRepos["gcr.io/project/dest1"])
	}

	if destRepos["gcr.io/project/dest2"] != 3 {
		t.Errorf("Expected dest2 to have count 3, got %d", destRepos["gcr.io/project/dest2"])
	}

	if destRepos["gcr.io/project/dest3"] != 7 {
		t.Errorf("Expected dest3 to have count 7, got %d", destRepos["gcr.io/project/dest3"])
	}
}

func TestPrometheusMetricsConcurrent(t *testing.T) {
	metrics := NewPrometheusMetrics()

	// Test that metrics are safe to use concurrently
	done := make(chan bool)

	// Simulate concurrent replications
	for i := 0; i < 5; i++ {
		go func(id int) {
			src := "ecr.io/project/repo"
			dest := "gcr.io/project/dest"

			metrics.ReplicationStarted(src, dest)

			// Simulate some work
			time.Sleep(10 * time.Millisecond)

			if id%2 == 0 {
				metrics.ReplicationCompleted(100*time.Millisecond, id, int64(id*1024))
			} else {
				metrics.ReplicationFailed()
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Check some metrics
	if count := metrics.GetReplicationCount(); count != 5 {
		t.Errorf("Expected replication count of 5, got %d", count)
	}

	if errors := metrics.GetReplicationErrors(); errors != 2 { // IDs 1, 3 fail
		t.Errorf("Expected error count of 2, got %d", errors)
	}

	// Check source repo count
	sourceRepos := metrics.GetTopSourceRepositories(10)
	if sourceRepos["ecr.io/project/repo"] != 5 {
		t.Errorf("Expected source repo count of 5, got %d", sourceRepos["ecr.io/project/repo"])
	}
}

func TestZeroValues(t *testing.T) {
	metrics := NewPrometheusMetrics()

	// Test initial zero values
	if metrics.GetReplicationCount() != 0 {
		t.Errorf("Expected initial replication count to be 0")
	}

	if metrics.GetReplicationErrors() != 0 {
		t.Errorf("Expected initial error count to be 0")
	}

	if metrics.GetLayersCopied() != 0 {
		t.Errorf("Expected initial layers count to be 0")
	}

	if metrics.GetBytesCopied() != 0 {
		t.Errorf("Expected initial bytes count to be 0")
	}

	if metrics.GetAverageLatency() != 0 {
		t.Errorf("Expected initial latency to be 0")
	}

	if len(metrics.GetTopSourceRepositories(10)) != 0 {
		t.Errorf("Expected empty top sources initially")
	}

	if len(metrics.GetTopDestinationRepositories(10)) != 0 {
		t.Errorf("Expected empty top destinations initially")
	}
}
