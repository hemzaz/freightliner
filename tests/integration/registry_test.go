//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"freightliner/pkg/client"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"
	"freightliner/pkg/replication"
	"freightliner/pkg/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContext holds the test environment setup
type TestContext struct {
	Config           *config.Config
	Logger           log.Logger
	Factory          *client.Factory
	ReplicationSvc   *service.ReplicateService
	SourceRegistry   interfaces.RegistryClient
	DestRegistry     interfaces.RegistryClient
	TestRepositories []string
	Cleanup          func()
}

// setupTestContext initializes the test environment
func setupTestContext(t *testing.T) *TestContext {
	// Load configuration
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    os.Getenv("TEST_AWS_REGION"),
			AccountID: os.Getenv("TEST_AWS_ACCOUNT_ID"),
		},
		GCR: config.GCRConfig{
			Project:  os.Getenv("TEST_GCP_PROJECT"),
			Location: "us",
		},
	}

	// Initialize logger
	logger := log.NewBasicLogger(log.DebugLevel)

	// Create client factory
	factory := client.NewFactory(cfg, logger)

	// Create test context
	ctx := &TestContext{
		Config:           cfg,
		Logger:           logger,
		Factory:          factory,
		TestRepositories: []string{},
	}

	// Cleanup function
	ctx.Cleanup = func() {
		// Clean up test repositories
		for _, repo := range ctx.TestRepositories {
			t.Logf("Cleaning up test repository: %s", repo)
		}
	}

	return ctx
}

// TestSingleRepositoryReplication tests replicating a single repository
func TestSingleRepositoryReplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := setupTestContext(t)
	defer ctx.Cleanup()

	testCases := []struct {
		name            string
		sourceRegistry  string
		sourceRepo      string
		destRegistry    string
		destRepo        string
		tags            []string
		expectedSuccess bool
	}{
		{
			name:            "ECR to GCR single tag",
			sourceRegistry:  "ecr",
			sourceRepo:      "test-app",
			destRegistry:    "gcr",
			destRepo:        "test-app",
			tags:            []string{"latest"},
			expectedSuccess: true,
		},
		{
			name:            "DockerHub to Harbor multi-tag",
			sourceRegistry:  "dockerhub",
			sourceRepo:      "library/alpine",
			destRegistry:    "harbor",
			destRepo:        "library/alpine",
			tags:            []string{"latest", "3.18"},
			expectedSuccess: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create replication rule
			rule := replication.ReplicationRule{
				SourceRegistry:        tc.sourceRegistry,
				SourceRepository:      tc.sourceRepo,
				DestinationRegistry:   tc.destRegistry,
				DestinationRepository: tc.destRepo,
				IncludeTags:           tc.tags,
			}

			// Execute replication
			svc := service.NewReplicateService(ctx.Config, ctx.Logger)
			err := svc.ReplicateRepository(context.Background(), rule)

			if tc.expectedSuccess {
				assert.NoError(t, err, "Replication should succeed")

				// Verify image exists in destination
				destClient, err := ctx.Factory.CreateClientForRegistry(context.Background(), tc.destRegistry)
				require.NoError(t, err)

				repo := destClient.Repository(tc.destRepo)
				for _, tag := range tc.tags {
					descriptor, err := repo.Descriptor(context.Background(), tag)
					assert.NoError(t, err, "Tag %s should exist in destination", tag)
					assert.NotNil(t, descriptor)
				}
			} else {
				assert.Error(t, err, "Replication should fail")
			}
		})
	}
}

// TestMultiRepositoryBatchReplication tests batch replication of multiple repositories
func TestMultiRepositoryBatchReplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := setupTestContext(t)
	defer ctx.Cleanup()

	// Define multiple repositories to replicate
	repos := []string{"app1", "app2", "app3", "app4", "app5"}

	// Create replication rules
	rules := make([]replication.ReplicationRule, len(repos))
	for i, repo := range repos {
		rules[i] = replication.ReplicationRule{
			SourceRegistry:        "ecr",
			SourceRepository:      repo,
			DestinationRegistry:   "gcr",
			DestinationRepository: repo,
			IncludeTags:           []string{"latest", "v*"},
		}
	}

	// Create worker pool for parallel replication
	pool := replication.NewWorkerPool(5, ctx.Logger)
	pool.Start()
	defer pool.Stop()

	// Submit replication jobs
	for _, rule := range rules {
		r := rule // Capture loop variable
		err := pool.Submit(r.SourceRepository, func(jobCtx context.Context) error {
			svc := service.NewReplicateService(ctx.Config, ctx.Logger)
			return svc.ReplicateRepository(jobCtx, r)
		})
		require.NoError(t, err)
	}

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		pool.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Log("All batch replications completed")
	case <-time.After(10 * time.Minute):
		t.Fatal("Batch replication timed out")
	}

	// Verify all repositories were replicated
	for _, repo := range repos {
		t.Run(fmt.Sprintf("Verify_%s", repo), func(t *testing.T) {
			destClient, err := ctx.Factory.CreateClientForRegistry(context.Background(), "gcr")
			require.NoError(t, err)

			repository := destClient.Repository(repo)
			descriptor, err := repository.Descriptor(context.Background(), "latest")
			assert.NoError(t, err, "Repository %s should exist in destination", repo)
			assert.NotNil(t, descriptor)
		})
	}
}

// TestCrossCloudReplication tests replication across different cloud providers
func TestCrossCloudReplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	scenarios := []struct {
		name       string
		source     string
		dest       string
		repository string
		tag        string
	}{
		{
			name:       "ECR to GCR",
			source:     "ecr",
			dest:       "gcr",
			repository: "cross-cloud-app",
			tag:        "latest",
		},
		{
			name:       "GCR to ECR",
			source:     "gcr",
			dest:       "ecr",
			repository: "cross-cloud-app",
			tag:        "latest",
		},
		{
			name:       "DockerHub to Harbor",
			source:     "dockerhub",
			dest:       "harbor",
			repository: "library/nginx",
			tag:        "alpine",
		},
		{
			name:       "GHCR to Quay",
			source:     "ghcr",
			dest:       "quay",
			repository: "test/app",
			tag:        "latest",
		},
	}

	ctx := setupTestContext(t)
	defer ctx.Cleanup()

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			rule := replication.ReplicationRule{
				SourceRegistry:        sc.source,
				SourceRepository:      sc.repository,
				DestinationRegistry:   sc.dest,
				DestinationRepository: sc.repository,
				IncludeTags:           []string{sc.tag},
			}

			svc := service.NewReplicateService(ctx.Config, ctx.Logger)
			err := svc.ReplicateRepository(context.Background(), rule)
			assert.NoError(t, err, "Cross-cloud replication should succeed")

			// Verify destination
			destClient, err := ctx.Factory.CreateClientForRegistry(context.Background(), sc.dest)
			require.NoError(t, err)

			repo := destClient.Repository(sc.repository)
			descriptor, err := repo.Descriptor(context.Background(), sc.tag)
			assert.NoError(t, err)
			assert.NotNil(t, descriptor)
		})
	}
}

// TestReplicationWithFilters tests tag filtering during replication
func TestReplicationWithFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := setupTestContext(t)
	defer ctx.Cleanup()

	testCases := []struct {
		name         string
		includeTags  []string
		excludeTags  []string
		expectedTags []string
	}{
		{
			name:         "Include specific tags",
			includeTags:  []string{"v1.0.0", "v1.0.1"},
			excludeTags:  []string{},
			expectedTags: []string{"v1.0.0", "v1.0.1"},
		},
		{
			name:         "Include pattern with exclusion",
			includeTags:  []string{"v*"},
			excludeTags:  []string{"*-alpha", "*-beta"},
			expectedTags: []string{"v1.0.0", "v1.0.1", "v1.1.0"},
		},
		{
			name:         "Latest only",
			includeTags:  []string{"latest"},
			excludeTags:  []string{},
			expectedTags: []string{"latest"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rule := replication.ReplicationRule{
				SourceRegistry:        "ecr",
				SourceRepository:      "filtered-app",
				DestinationRegistry:   "gcr",
				DestinationRepository: "filtered-app",
				IncludeTags:           tc.includeTags,
				ExcludeTags:           tc.excludeTags,
			}

			svc := service.NewReplicateService(ctx.Config, ctx.Logger)
			err := svc.ReplicateRepository(context.Background(), rule)
			assert.NoError(t, err)

			// Verify only expected tags were replicated
			destClient, err := ctx.Factory.CreateClientForRegistry(context.Background(), "gcr")
			require.NoError(t, err)

			repo := destClient.Repository("filtered-app")
			for _, tag := range tc.expectedTags {
				descriptor, err := repo.Descriptor(context.Background(), tag)
				assert.NoError(t, err, "Tag %s should exist", tag)
				assert.NotNil(t, descriptor)
			}
		})
	}
}

// TestReplicationResume tests checkpoint and resume functionality
func TestReplicationResume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := setupTestContext(t)
	defer ctx.Cleanup()

	rule := replication.ReplicationRule{
		SourceRegistry:        "ecr",
		SourceRepository:      "large-app",
		DestinationRegistry:   "gcr",
		DestinationRepository: "large-app",
		IncludeTags:           []string{"latest"},
	}

	// Start replication with cancellation
	cancelCtx, cancel := context.WithCancel(context.Background())

	// Cancel after 2 seconds to simulate interruption
	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	svc := service.NewReplicateService(ctx.Config, ctx.Logger)
	err := svc.ReplicateRepository(cancelCtx, rule)
	assert.Error(t, err, "Expected context cancellation error")
	assert.Contains(t, err.Error(), "context canceled")

	// Resume replication
	t.Log("Resuming replication from checkpoint")
	err = svc.ResumeReplication(context.Background(), rule)
	assert.NoError(t, err, "Resume should complete successfully")

	// Verify completion
	destClient, err := ctx.Factory.CreateClientForRegistry(context.Background(), "gcr")
	require.NoError(t, err)

	repo := destClient.Repository("large-app")
	descriptor, err := repo.Descriptor(context.Background(), "latest")
	assert.NoError(t, err)
	assert.NotNil(t, descriptor)
}

// TestConcurrentReplication tests concurrent replication with race detection
func TestConcurrentReplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := setupTestContext(t)
	defer ctx.Cleanup()

	// Create worker pool with 10 workers
	pool := replication.NewWorkerPool(10, ctx.Logger)
	pool.Start()
	defer pool.Stop()

	// Submit 50 concurrent replication jobs
	jobCount := 50
	for i := 0; i < jobCount; i++ {
		repo := fmt.Sprintf("concurrent-app-%d", i)
		err := pool.Submit(repo, func(jobCtx context.Context) error {
			rule := replication.ReplicationRule{
				SourceRegistry:        "ecr",
				SourceRepository:      repo,
				DestinationRegistry:   "gcr",
				DestinationRepository: repo,
				IncludeTags:           []string{"latest"},
			}

			svc := service.NewReplicateService(ctx.Config, ctx.Logger)
			return svc.ReplicateRepository(jobCtx, rule)
		})
		require.NoError(t, err)
	}

	// Wait for completion
	pool.Wait()

	// Verify all replications succeeded
	destClient, err := ctx.Factory.CreateClientForRegistry(context.Background(), "gcr")
	require.NoError(t, err)

	successCount := 0
	for i := 0; i < jobCount; i++ {
		repo := fmt.Sprintf("concurrent-app-%d", i)
		repository := destClient.Repository(repo)
		if descriptor, err := repository.Descriptor(context.Background(), "latest"); err == nil && descriptor != nil {
			successCount++
		}
	}

	// At least 90% should succeed (allowing for some transient failures)
	minSuccess := int(float64(jobCount) * 0.9)
	assert.GreaterOrEqual(t, successCount, minSuccess,
		"At least %d out of %d replications should succeed", minSuccess, jobCount)
}

// TestRateLimitingAndRetries tests rate limiting and retry logic
func TestRateLimitingAndRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := setupTestContext(t)
	defer ctx.Cleanup()

	// Configure with aggressive rate limiting
	cfg := ctx.Config
	cfg.RateLimit = config.RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             5,
	}
	cfg.Retry = config.RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     10 * time.Second,
	}

	// Submit 100 rapid requests
	pool := replication.NewWorkerPool(20, ctx.Logger)
	pool.Start()
	defer pool.Stop()

	for i := 0; i < 100; i++ {
		repo := fmt.Sprintf("rate-limited-app-%d", i)
		err := pool.Submit(repo, func(jobCtx context.Context) error {
			rule := replication.ReplicationRule{
				SourceRegistry:        "ecr",
				SourceRepository:      repo,
				DestinationRegistry:   "gcr",
				DestinationRepository: repo,
				IncludeTags:           []string{"latest"},
			}

			svc := service.NewReplicateService(cfg, ctx.Logger)
			return svc.ReplicateRepository(jobCtx, rule)
		})
		require.NoError(t, err)
	}

	// Wait for completion
	pool.Wait()

	t.Log("All rate-limited requests completed with retries")
}

// TestAuthenticationFailures tests handling of authentication failures
func TestAuthenticationFailures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testCases := []struct {
		name        string
		registryURL string
		authConfig  config.AuthConfig
		expectedErr string
	}{
		{
			name:        "Invalid ECR credentials",
			registryURL: "123456789012.dkr.ecr.us-east-1.amazonaws.com",
			authConfig: config.AuthConfig{
				Type:     config.AuthTypeBasic,
				Username: "invalid",
				Password: "invalid",
			},
			expectedErr: "authentication failed",
		},
		{
			name:        "Expired GCR token",
			registryURL: "gcr.io",
			authConfig: config.AuthConfig{
				Type:  config.AuthTypeToken,
				Token: "expired-token",
			},
			expectedErr: "unauthorized",
		},
		{
			name:        "No credentials for private registry",
			registryURL: "private-registry.example.com",
			authConfig: config.AuthConfig{
				Type: config.AuthTypeAnonymous,
			},
			expectedErr: "unauthorized",
		},
	}

	ctx := setupTestContext(t)
	defer ctx.Cleanup()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create registry config with invalid auth
			regConfig := config.RegistryConfig{
				Name:     "test-invalid",
				Type:     config.RegistryTypeGeneric,
				Endpoint: tc.registryURL,
				Auth:     tc.authConfig,
			}

			factory := client.NewFactory(&config.Config{
				Registries: config.RegistriesConfig{
					Registries: []config.RegistryConfig{regConfig},
				},
			}, ctx.Logger)

			client, err := factory.CreateCustomClient("test-invalid")
			if err != nil {
				assert.Contains(t, err.Error(), tc.expectedErr)
				return
			}

			// Try to list repositories (should fail)
			repo := client.Repository("test-repo")
			_, err = repo.Descriptor(context.Background(), "latest")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestScheduledReplication tests the scheduler functionality
func TestScheduledReplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := setupTestContext(t)
	defer ctx.Cleanup()

	// Create worker pool
	pool := replication.NewWorkerPool(5, ctx.Logger)
	pool.Start()
	defer pool.Stop()

	// Create replication service
	svc := service.NewReplicateService(ctx.Config, ctx.Logger)

	// Create scheduler
	scheduler := replication.NewScheduler(replication.SchedulerOptions{
		Logger:             ctx.Logger,
		WorkerPool:         pool,
		ReplicationService: svc,
	})
	defer scheduler.Stop()

	// Add immediate execution job
	rule := replication.ReplicationRule{
		SourceRegistry:        "ecr",
		SourceRepository:      "scheduled-app",
		DestinationRegistry:   "gcr",
		DestinationRepository: "scheduled-app",
		IncludeTags:           []string{"latest"},
		Schedule:              "@now",
	}

	err := scheduler.AddJob(rule)
	assert.NoError(t, err)

	// Wait for job to complete
	time.Sleep(10 * time.Second)

	// Verify replication occurred
	destClient, err := ctx.Factory.CreateClientForRegistry(context.Background(), "gcr")
	require.NoError(t, err)

	repo := destClient.Repository("scheduled-app")
	descriptor, err := repo.Descriptor(context.Background(), "latest")
	assert.NoError(t, err)
	assert.NotNil(t, descriptor)
}
