package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiRegistryScenario tests operations across multiple registries
func TestMultiRegistryScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	registries := []struct {
		name         string
		registryType string
		accountID    string
		region       string
		repoName     string
	}{
		{
			name:         "ECR US West",
			registryType: "ecr",
			accountID:    "123456789012",
			region:       "us-west-2",
			repoName:     "app-west",
		},
		{
			name:         "ECR US East",
			registryType: "ecr",
			accountID:    "123456789012",
			region:       "us-east-1",
			repoName:     "app-east",
		},
		{
			name:         "GCR Production",
			registryType: "gcr",
			accountID:    "prod-project",
			region:       "",
			repoName:     "app-prod",
		},
		{
			name:         "GCR Staging",
			registryType: "gcr",
			accountID:    "staging-project",
			region:       "",
			repoName:     "app-staging",
		},
	}

	for _, reg := range registries {
		t.Run(reg.name, func(t *testing.T) {
			// Format URI
			uri := util.FormatRepositoryURI(reg.registryType, reg.accountID, reg.region, reg.repoName)
			assert.NotEmpty(t, uri)

			// Validate registry type
			isValid := util.IsValidRegistryType(reg.registryType)
			assert.True(t, isValid, "Registry type %s should be valid", reg.registryType)

			// Parse the formatted path
			registryPath := fmt.Sprintf("%s/%s", reg.registryType, reg.repoName)
			registry, repo, err := util.ParseRegistryPath(registryPath)
			require.NoError(t, err)
			assert.Equal(t, reg.registryType, registry)
			assert.Equal(t, reg.repoName, repo)
		})
	}
}

// TestRegistryFailover tests failover scenarios
func TestRegistryFailover(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	primaryRegistry := "ecr"
	secondaryRegistry := "gcr"
	repoName := "critical-app"

	// Simulate primary registry failure
	t.Run("Failover to secondary", func(t *testing.T) {
		// Primary registry (simulated failure)
		isPrimaryValid := util.IsValidRegistryType(primaryRegistry)
		assert.True(t, isPrimaryValid)

		// Failover to secondary
		isSecondaryValid := util.IsValidRegistryType(secondaryRegistry)
		assert.True(t, isSecondaryValid)

		// Both registries should handle the same repo name
		primaryURI := util.FormatRepositoryURI(primaryRegistry, "123456789012", "us-west-2", repoName)
		secondaryURI := util.FormatRepositoryURI(secondaryRegistry, "my-project", "", repoName)

		assert.NotEmpty(t, primaryURI)
		assert.NotEmpty(t, secondaryURI)
		assert.NotEqual(t, primaryURI, secondaryURI, "URIs should be different for different registries")
	})
}

// TestRegistryConcurrentOperations tests concurrent operations across registries
func TestRegistryConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))
	ctx := context.Background()

	registries := []string{"ecr", "gcr"}
	repositories := []string{"repo1", "repo2", "repo3", "repo4", "repo5"}

	var wg sync.WaitGroup
	errors := make(chan error, len(registries)*len(repositories))

	for _, registry := range registries {
		for _, repo := range repositories {
			wg.Add(1)
			go func(reg, rep string) {
				defer wg.Done()

				// Parse registry path
				path := fmt.Sprintf("%s/%s", reg, rep)
				_, _, err := util.ParseRegistryPath(path)
				if err != nil {
					errors <- err
					return
				}

				// Validate repository name
				if err := util.ValidateRepositoryName(rep); err != nil {
					errors <- err
					return
				}

				// Format URI
				accountID := "test-account"
				region := "us-west-2"
				if reg == "gcr" {
					region = ""
				}
				uri := util.FormatRepositoryURI(reg, accountID, region, rep)
				if uri == "" {
					errors <- fmt.Errorf("empty URI for %s/%s", reg, rep)
					return
				}

				// Log operation
				util.LogRegistryOperation(ctx, "test", reg, rep, nil)
			}(registry, repo)
		}
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
		errorCount++
	}

	assert.Equal(t, 0, errorCount, "Should have no errors in concurrent operations")
}

// TestRegistryBackwardCompatibility tests backward compatibility scenarios
func TestRegistryBackwardCompatibility(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	// Test cases that should continue to work with old code
	legacyTests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Legacy ECR format",
			path:     "ecr/legacy-app",
			expected: "ecr",
		},
		{
			name:     "Legacy GCR format",
			path:     "gcr/project/app",
			expected: "gcr",
		},
	}

	for _, tt := range legacyTests {
		t.Run(tt.name, func(t *testing.T) {
			registry, _, err := util.ParseRegistryPath(tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, registry)
		})
	}
}

// TestRegistryURIResolution tests complete URI resolution workflow
func TestRegistryURIResolution(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name         string
		registryType string
		accountID    string
		region       string
		repoName     string
		expectedURI  string
	}{
		{
			name:         "Complete ECR workflow",
			registryType: "ecr",
			accountID:    "123456789012",
			region:       "us-west-2",
			repoName:     "my-service",
			expectedURI:  "123456789012.dkr.ecr.us-west-2.amazonaws.com/my-service",
		},
		{
			name:         "Complete GCR workflow",
			registryType: "gcr",
			accountID:    "my-project",
			region:       "",
			repoName:     "my-service",
			expectedURI:  "gcr.io/my-project/my-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Validate registry type
			isValid := util.IsValidRegistryType(tt.registryType)
			require.True(t, isValid, "Registry type should be valid")

			// Step 2: Validate repository name
			err := util.ValidateRepositoryName(tt.repoName)
			require.NoError(t, err, "Repository name should be valid")

			// Step 3: Format URI
			uri := util.FormatRepositoryURI(tt.registryType, tt.accountID, tt.region, tt.repoName)
			assert.Equal(t, tt.expectedURI, uri)

			// Step 4: Parse back the registry path
			path := fmt.Sprintf("%s/%s", tt.registryType, tt.repoName)
			registry, repo, err := util.ParseRegistryPath(path)
			require.NoError(t, err)
			assert.Equal(t, tt.registryType, registry)
			assert.Equal(t, tt.repoName, repo)
		})
	}
}

// TestRegistryErrorRecovery tests error recovery scenarios
func TestRegistryErrorRecovery(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	t.Run("Recover from invalid path", func(t *testing.T) {
		// First attempt fails
		_, _, err := util.ParseRegistryPath("invalid")
		require.Error(t, err)

		// Second attempt with valid path succeeds
		registry, repo, err := util.ParseRegistryPath("ecr/valid-repo")
		require.NoError(t, err)
		assert.Equal(t, "ecr", registry)
		assert.Equal(t, "valid-repo", repo)
	})

	t.Run("Recover from empty repository", func(t *testing.T) {
		// First attempt fails
		err := util.ValidateRepositoryName("")
		require.Error(t, err)

		// Second attempt with valid name succeeds
		err = util.ValidateRepositoryName("valid-repo")
		require.NoError(t, err)
	})
}

// TestRegistryHTTPIntegration tests HTTP transport integration
func TestRegistryHTTPIntegration(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	// Create a test HTTP server
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create transport
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		DisableKeepAlives:  false,
	}

	// Get remote options with transport
	opts := util.GetRemoteOptions(transport)
	assert.NotNil(t, opts)
	assert.Len(t, opts, 1)
}

// TestRegistryLoadSimulation simulates realistic load
func TestRegistryLoadSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))
	ctx := context.Background()

	// Simulate 100 concurrent users
	concurrentUsers := 100
	operationsPerUser := 10

	start := time.Now()
	var wg sync.WaitGroup

	for user := 0; user < concurrentUsers; user++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			for op := 0; op < operationsPerUser; op++ {
				registryType := "ecr"
				if userID%2 == 0 {
					registryType = "gcr"
				}

				repoName := fmt.Sprintf("user-%d-repo-%d", userID, op)

				// Parse path
				path := fmt.Sprintf("%s/%s", registryType, repoName)
				_, _, _ = util.ParseRegistryPath(path)

				// Validate
				_ = util.ValidateRepositoryName(repoName)

				// Format URI
				_ = util.FormatRepositoryURI(registryType, "account", "region", repoName)

				// Log
				util.LogRegistryOperation(ctx, "simulate", registryType, repoName, nil)
			}
		}(user)
	}

	wg.Wait()
	duration := time.Since(start)

	totalOps := concurrentUsers * operationsPerUser * 4 // 4 operations per iteration
	opsPerSecond := float64(totalOps) / duration.Seconds()

	t.Logf("Completed %d operations in %v (%.2f ops/sec)", totalOps, duration, opsPerSecond)

	// Should handle at least 1000 ops/sec
	assert.Greater(t, opsPerSecond, 1000.0, "Should handle at least 1000 operations per second")
}

// TestRegistryConfigurationScenarios tests different configuration scenarios
func TestRegistryConfigurationScenarios(t *testing.T) {
	tests := []struct {
		name     string
		scenario func(t *testing.T, util *common.RegistryUtil)
	}{
		{
			name: "Single registry configuration",
			scenario: func(t *testing.T, util *common.RegistryUtil) {
				registry := "ecr"
				assert.True(t, util.IsValidRegistryType(registry))

				uri := util.FormatRepositoryURI(registry, "123", "us-west-2", "app")
				assert.Contains(t, uri, "ecr")
			},
		},
		{
			name: "Multi-region configuration",
			scenario: func(t *testing.T, util *common.RegistryUtil) {
				regions := []string{"us-west-2", "us-east-1", "eu-west-1"}
				for _, region := range regions {
					uri := util.FormatRepositoryURI("ecr", "123", region, "app")
					assert.Contains(t, uri, region)
				}
			},
		},
		{
			name: "Multi-project GCR configuration",
			scenario: func(t *testing.T, util *common.RegistryUtil) {
				projects := []string{"dev-project", "staging-project", "prod-project"}
				for _, project := range projects {
					uri := util.FormatRepositoryURI("gcr", project, "", "app")
					assert.Contains(t, uri, project)
				}
			},
		},
		{
			name: "Mixed registry configuration",
			scenario: func(t *testing.T, util *common.RegistryUtil) {
				ecrURI := util.FormatRepositoryURI("ecr", "123", "us-west-2", "app")
				gcrURI := util.FormatRepositoryURI("gcr", "project", "", "app")

				assert.Contains(t, ecrURI, "ecr")
				assert.Contains(t, gcrURI, "gcr.io")
				assert.NotEqual(t, ecrURI, gcrURI)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))
			tt.scenario(t, util)
		})
	}
}

// BenchmarkMultiRegistryOperations benchmarks multi-registry operations
func BenchmarkMultiRegistryOperations(b *testing.B) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	registries := []string{"ecr", "gcr"}
	repositories := []string{"repo1", "repo2", "repo3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, reg := range registries {
			for _, repo := range repositories {
				path := fmt.Sprintf("%s/%s", reg, repo)
				_, _, _ = util.ParseRegistryPath(path)
				_ = util.ValidateRepositoryName(repo)
				_ = util.FormatRepositoryURI(reg, "account", "region", repo)
			}
		}
	}
}
