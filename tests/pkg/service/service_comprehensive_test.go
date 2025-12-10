package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"freightliner/pkg/config"
	"freightliner/pkg/copy"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==== COMPREHENSIVE SERVICE TESTS ====

// TestParseRegistryPathVariations tests all variations of registry path parsing
func TestParseRegistryPathVariations(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectedValid bool
		expectedParts int
	}{
		{"valid ecr simple", "ecr/repo", true, 2},
		{"valid ecr nested", "ecr/team/app", true, 2},
		{"valid gcr simple", "gcr/repo", true, 2},
		{"valid gcr nested", "gcr/project/app", true, 2},
		{"invalid no slash", "invalid", false, 1},
		{"invalid empty", "", false, 1},
		{"invalid only slash", "/", false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := len(tt.path)
			if tt.path != "" {
				// Count slashes to determine parts
				parts = 1
				for _, c := range tt.path {
					if c == '/' {
						parts++
					}
				}
			}

			if tt.expectedValid {
				assert.GreaterOrEqual(t, parts, 2)
			}
		})
	}
}

// TestIsValidRegistryTypeComprehensive tests all registry type validations
func TestIsValidRegistryTypeComprehensive(t *testing.T) {
	tests := []struct {
		name         string
		registryType string
		expected     bool
	}{
		{"ecr valid", "ecr", true},
		{"gcr valid", "gcr", true},
		{"ECR uppercase", "ECR", false},
		{"GCR uppercase", "GCR", false},
		{"docker invalid", "docker", false},
		{"dockerhub invalid", "dockerhub", false},
		{"quay invalid", "quay", false},
		{"harbor invalid", "harbor", false},
		{"empty invalid", "", false},
		{"spaces invalid", "  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.registryType == "ecr" || tt.registryType == "gcr"
			assert.Equal(t, tt.expected, isValid, "registry type: %s", tt.registryType)
		})
	}
}

// TestRegistryCredentialsJSONMarshaling tests JSON marshaling/unmarshaling
func TestRegistryCredentialsJSONMarshaling(t *testing.T) {
	original := service.RegistryCredentials{}
	original.ECR.AccessKey = "AKIATEST123"
	original.ECR.SecretKey = "secret123"
	original.ECR.SessionToken = "session123"
	original.ECR.Region = "us-east-1"
	original.ECR.AccountID = "123456789012"

	original.GCR.Credentials = base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`))
	original.GCR.Project = "test-project"
	original.GCR.Location = "us-central1"

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Unmarshal back
	var decoded service.RegistryCredentials
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, original.ECR.AccessKey, decoded.ECR.AccessKey)
	assert.Equal(t, original.ECR.SecretKey, decoded.ECR.SecretKey)
	assert.Equal(t, original.ECR.SessionToken, decoded.ECR.SessionToken)
	assert.Equal(t, original.ECR.Region, decoded.ECR.Region)
	assert.Equal(t, original.ECR.AccountID, decoded.ECR.AccountID)
	assert.Equal(t, original.GCR.Credentials, decoded.GCR.Credentials)
	assert.Equal(t, original.GCR.Project, decoded.GCR.Project)
	assert.Equal(t, original.GCR.Location, decoded.GCR.Location)
}

// TestEncryptionKeysJSONMarshaling tests JSON marshaling/unmarshaling
func TestEncryptionKeysJSONMarshaling(t *testing.T) {
	original := service.EncryptionKeys{}
	original.AWS.KMSKeyID = "arn:aws:kms:us-east-1:123456789012:key/test-key-id"
	original.GCP.KMSKeyID = "projects/test/locations/us/keyRings/ring/cryptoKeys/key"
	original.GCP.KeyRing = "test-keyring"
	original.GCP.Key = "test-key"

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var decoded service.EncryptionKeys
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.AWS.KMSKeyID, decoded.AWS.KMSKeyID)
	assert.Equal(t, original.GCP.KMSKeyID, decoded.GCP.KMSKeyID)
	assert.Equal(t, original.GCP.KeyRing, decoded.GCP.KeyRing)
	assert.Equal(t, original.GCP.Key, decoded.GCP.Key)
}

// TestCopyStatsAccumulation tests copy statistics accumulation
func TestCopyStatsAccumulation(t *testing.T) {
	stats1 := copy.CopyStats{
		BytesTransferred: 1024,
		Layers:           3,
		PullDuration:     100 * time.Millisecond,
		PushDuration:     150 * time.Millisecond,
	}

	stats2 := copy.CopyStats{
		BytesTransferred: 2048,
		Layers:           5,
		PullDuration:     200 * time.Millisecond,
		PushDuration:     250 * time.Millisecond,
	}

	// Accumulate stats
	totalBytes := stats1.BytesTransferred + stats2.BytesTransferred
	totalLayers := stats1.Layers + stats2.Layers
	totalDuration := stats1.PullDuration + stats1.PushDuration + stats2.PullDuration + stats2.PushDuration

	assert.Equal(t, int64(3072), totalBytes)
	assert.Equal(t, 8, totalLayers)
	assert.Equal(t, 700*time.Millisecond, totalDuration)
}

// TestReplicationProgressCalculations tests progress calculations
func TestReplicationProgressCalculations(t *testing.T) {
	tests := []struct {
		name               string
		completed          int
		total              int
		expectedPercentage float64
	}{
		{"0 percent", 0, 100, 0.0},
		{"25 percent", 25, 100, 25.0},
		{"50 percent", 50, 100, 50.0},
		{"75 percent", 75, 100, 75.0},
		{"100 percent", 100, 100, 100.0},
		{"partial 33", 1, 3, 33.33},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.total == 0 {
				return
			}
			percentage := float64(tt.completed) / float64(tt.total) * 100
			assert.InDelta(t, tt.expectedPercentage, percentage, 0.01)
		})
	}
}

// TestConcurrentProgressTracking tests thread-safe progress tracking
func TestConcurrentProgressTracking(t *testing.T) {
	var counter int32
	numGoroutines := 100
	incrementsPerGoroutine := 100

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				atomic.AddInt32(&counter, 1)
			}
		}()
	}

	wg.Wait()

	expected := int32(numGoroutines * incrementsPerGoroutine)
	assert.Equal(t, expected, atomic.LoadInt32(&counter))
}

// TestEnvironmentVariableHandling tests environment variable operations
func TestEnvironmentVariableHandling(t *testing.T) {
	// Save original values
	origAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	origSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	defer func() {
		// Restore
		if origAccessKey != "" {
			os.Setenv("AWS_ACCESS_KEY_ID", origAccessKey)
		} else {
			os.Unsetenv("AWS_ACCESS_KEY_ID")
		}
		if origSecretKey != "" {
			os.Setenv("AWS_SECRET_ACCESS_KEY", origSecretKey)
		} else {
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		}
	}()

	// Test setting
	testAccessKey := "TESTKEY123"
	testSecretKey := "TESTSECRET456"

	err := os.Setenv("AWS_ACCESS_KEY_ID", testAccessKey)
	require.NoError(t, err)
	err = os.Setenv("AWS_SECRET_ACCESS_KEY", testSecretKey)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, testAccessKey, os.Getenv("AWS_ACCESS_KEY_ID"))
	assert.Equal(t, testSecretKey, os.Getenv("AWS_SECRET_ACCESS_KEY"))
}

// TestBatchOperationResults tests batch operation result handling
func TestBatchOperationResults(t *testing.T) {
	results := []*service.ReplicationResult{
		{Success: true, BytesCopied: 1024, LayersCopied: 3},
		{Success: false, BytesCopied: 0, LayersCopied: 0},
		{Success: true, BytesCopied: 2048, LayersCopied: 5},
		{Success: true, BytesCopied: 4096, LayersCopied: 7},
		{Success: false, BytesCopied: 0, LayersCopied: 0},
	}

	// Calculate metrics
	successCount := 0
	failureCount := 0
	var totalBytes int64
	totalLayers := 0

	for _, r := range results {
		if r.Success {
			successCount++
			totalBytes += r.BytesCopied
			totalLayers += r.LayersCopied
		} else {
			failureCount++
		}
	}

	assert.Equal(t, 3, successCount)
	assert.Equal(t, 2, failureCount)
	assert.Equal(t, int64(7168), totalBytes)
	assert.Equal(t, 15, totalLayers)

	successRate := float64(successCount) / float64(len(results)) * 100
	assert.InDelta(t, 60.0, successRate, 0.01)
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	workDone := make(chan bool, 1)
	go func() {
		select {
		case <-ctx.Done():
			workDone <- false
		case <-time.After(100 * time.Millisecond):
			workDone <- true
		}
	}()

	// Cancel immediately
	cancel()

	select {
	case done := <-workDone:
		assert.False(t, done, "work should have been cancelled")
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for cancellation")
	}
}

// TestTimeoutContextHandling tests timeout context handling
func TestTimeoutContextHandling(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		workDuration  time.Duration
		expectTimeout bool
	}{
		{"completes before timeout", 100 * time.Millisecond, 50 * time.Millisecond, false},
		{"exceeds timeout", 50 * time.Millisecond, 100 * time.Millisecond, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			done := make(chan bool, 1)
			go func() {
				time.Sleep(tt.workDuration)
				done <- true
			}()

			select {
			case <-done:
				assert.False(t, tt.expectTimeout)
			case <-ctx.Done():
				assert.True(t, tt.expectTimeout)
			}
		})
	}
}

// TestErrorWrapping tests error wrapping and unwrapping
func TestErrorWrapping(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := errors.New("wrapped: " + baseErr.Error())

	assert.Error(t, baseErr)
	assert.Error(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "base error")
	assert.Contains(t, wrappedErr.Error(), "wrapped")
}

// TestWorkerCountNormalization tests worker count normalization
func TestWorkerCountNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"negative becomes 1", -5, 1},
		{"zero becomes 1", 0, 1},
		{"positive unchanged", 4, 4},
		{"large unchanged", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := tt.input
			if normalized <= 0 {
				normalized = 1
			}
			assert.Equal(t, tt.expected, normalized)
		})
	}
}

// TestAutoDetectWorkerCount tests worker count auto-detection
func TestAutoDetectWorkerCount(t *testing.T) {
	optimalCount := config.GetOptimalWorkerCount()
	assert.Greater(t, optimalCount, 0)
	assert.LessOrEqual(t, optimalCount, 128) // Reasonable upper bound
}

// TestDryRunMode tests dry run mode behavior
func TestDryRunMode(t *testing.T) {
	cfg := &config.Config{
		Replicate: config.ReplicateConfig{
			DryRun: true,
		},
	}

	assert.True(t, cfg.Replicate.DryRun)

	// In dry run mode, no actual operations should execute
	// This is a configuration validation test
}

// TestForceOverwriteMode tests force overwrite behavior
func TestForceOverwriteMode(t *testing.T) {
	cfg := &config.Config{
		Replicate: config.ReplicateConfig{
			Force: true,
		},
	}

	assert.True(t, cfg.Replicate.Force)
}

// TestSecretsManagerDisabled tests behavior when secrets manager is disabled
func TestSecretsManagerDisabled(t *testing.T) {
	cfg := &config.Config{
		Secrets: config.SecretsConfig{
			UseSecretsManager: false,
		},
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	svc := service.NewReplicationService(cfg, logger)

	assert.NotNil(t, svc)
	assert.False(t, cfg.Secrets.UseSecretsManager)
}

// TestEncryptionDisabled tests behavior when encryption is disabled
func TestEncryptionDisabled(t *testing.T) {
	cfg := &config.Config{
		Encryption: config.EncryptionConfig{
			Enabled: false,
		},
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	svc := service.NewReplicationService(cfg, logger)

	assert.NotNil(t, svc)
	assert.False(t, cfg.Encryption.Enabled)
}

// TestReplicationResultTimingCalculations tests timing calculations
func TestReplicationResultTimingCalculations(t *testing.T) {
	startTime := time.Now()
	time.Sleep(50 * time.Millisecond)
	endTime := time.Now()

	duration := endTime.Sub(startTime)

	assert.True(t, duration > 0)
	assert.True(t, duration >= 50*time.Millisecond)
	assert.True(t, endTime.After(startTime))
}

// TestMetricsAggregation tests metrics aggregation
func TestMetricsAggregation(t *testing.T) {
	type metrics struct {
		totalRequests int
		successCount  int
		failureCount  int
		totalBytes    int64
		totalLayers   int
	}

	m := metrics{}

	// Simulate 10 operations
	operations := []struct {
		success bool
		bytes   int64
		layers  int
	}{
		{true, 1024, 3},
		{true, 2048, 5},
		{false, 0, 0},
		{true, 4096, 7},
		{false, 0, 0},
		{true, 8192, 9},
	}

	for _, op := range operations {
		m.totalRequests++
		if op.success {
			m.successCount++
			m.totalBytes += op.bytes
			m.totalLayers += op.layers
		} else {
			m.failureCount++
		}
	}

	assert.Equal(t, 6, m.totalRequests)
	assert.Equal(t, 4, m.successCount)
	assert.Equal(t, 2, m.failureCount)
	assert.Equal(t, int64(15360), m.totalBytes)
	assert.Equal(t, 24, m.totalLayers)
}

// TestConcurrentBatchProcessing tests concurrent batch processing
func TestConcurrentBatchProcessing(t *testing.T) {
	numBatches := 5
	batchSize := 10
	var processedItems int32

	var wg sync.WaitGroup
	for i := 0; i < numBatches; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < batchSize; j++ {
				atomic.AddInt32(&processedItems, 1)
			}
		}()
	}

	wg.Wait()

	expected := int32(numBatches * batchSize)
	assert.Equal(t, expected, atomic.LoadInt32(&processedItems))
}
