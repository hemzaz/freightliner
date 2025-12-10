package testing

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
)

// OptimizedTestRunner provides enhanced test execution with timeout management
type OptimizedTestRunner struct {
	logger         log.Logger
	maxConcurrency int
	defaultTimeout time.Duration
	retryCount     int
	healthChecks   map[string]HealthChecker
	mu             sync.RWMutex
}

// HealthChecker defines interface for service health validation
type HealthChecker interface {
	IsHealthy(ctx context.Context) error
	WaitForReady(ctx context.Context, timeout time.Duration) error
	GetServiceName() string
}

// TestExecutionResult contains test execution outcomes
type TestExecutionResult struct {
	Package       string
	Success       bool
	Duration      time.Duration
	Output        string
	Error         error
	RetryAttempts int
	TimedOut      bool
}

// TestSuiteResult aggregates results from multiple test executions
type TestSuiteResult struct {
	TotalTests    int
	PassedTests   int
	FailedTests   int
	TimedOutTests int
	TotalDuration time.Duration
	Results       []TestExecutionResult
}

// NewOptimizedTestRunner creates a new test runner with optimization features
func NewOptimizedTestRunner(logger log.Logger) *OptimizedTestRunner {
	if logger == nil {
		logger = log.NewLogger()
	}

	return &OptimizedTestRunner{
		logger:         logger,
		maxConcurrency: 4, // Reasonable default for most CI environments
		defaultTimeout: 5 * time.Minute,
		retryCount:     2,
		healthChecks:   make(map[string]HealthChecker),
	}
}

// WithConcurrency sets the maximum number of concurrent test executions
func (otr *OptimizedTestRunner) WithConcurrency(max int) *OptimizedTestRunner {
	otr.maxConcurrency = max
	return otr
}

// WithTimeout sets the default timeout for test execution
func (otr *OptimizedTestRunner) WithTimeout(timeout time.Duration) *OptimizedTestRunner {
	otr.defaultTimeout = timeout
	return otr
}

// WithRetryCount sets the number of retry attempts for flaky tests
func (otr *OptimizedTestRunner) WithRetryCount(count int) *OptimizedTestRunner {
	otr.retryCount = count
	return otr
}

// AddHealthCheck registers a health checker for a service dependency
func (otr *OptimizedTestRunner) AddHealthCheck(checker HealthChecker) {
	otr.mu.Lock()
	defer otr.mu.Unlock()
	otr.healthChecks[checker.GetServiceName()] = checker
}

// RunIntegrationTests executes integration tests with optimizations
func (otr *OptimizedTestRunner) RunIntegrationTests(ctx context.Context, packages []string) (*TestSuiteResult, error) {
	otr.logger.Info(fmt.Sprintf("Starting optimized integration test execution packages=%d concurrency=%d timeout=%v",
		len(packages), otr.maxConcurrency, otr.defaultTimeout))

	// Wait for all service dependencies to be ready
	if err := otr.waitForDependencies(ctx); err != nil {
		return nil, fmt.Errorf("service dependencies not ready: %w", err)
	}

	// Execute tests with controlled concurrency
	results := make([]TestExecutionResult, 0, len(packages))
	semaphore := make(chan struct{}, otr.maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	startTime := time.Now()

	for _, pkg := range packages {
		wg.Add(1)
		go func(packagePath string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := otr.executeTestPackage(ctx, packagePath)

			mu.Lock()
			results = append(results, result)
			mu.Unlock()

			if result.Success {
				otr.logger.Info(fmt.Sprintf("✓ %s completed in %v", packagePath, result.Duration))
			} else {
				otr.logger.Error(fmt.Sprintf("✗ %s failed after %v", packagePath, result.Duration), result.Error)
			}
		}(pkg)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Compile results
	suite := &TestSuiteResult{
		TotalTests:    len(results),
		TotalDuration: totalDuration,
		Results:       results,
	}

	for _, result := range results {
		if result.Success {
			suite.PassedTests++
		} else {
			suite.FailedTests++
			if result.TimedOut {
				suite.TimedOutTests++
			}
		}
	}

	otr.logger.Info(fmt.Sprintf("Test suite completed: %d/%d passed, %d timed out, duration: %v",
		suite.PassedTests, suite.TotalTests, suite.TimedOutTests, suite.TotalDuration))

	return suite, nil
}

// executeTestPackage runs tests for a single package with retry logic
func (otr *OptimizedTestRunner) executeTestPackage(ctx context.Context, packagePath string) TestExecutionResult {
	result := TestExecutionResult{
		Package: packagePath,
	}

	for attempt := 0; attempt <= otr.retryCount; attempt++ {
		if attempt > 0 {
			otr.logger.Warn(fmt.Sprintf("Retrying %s (attempt %d/%d)", packagePath, attempt+1, otr.retryCount+1))

			// Exponential backoff
			backoff := time.Duration(attempt) * 2 * time.Second
			select {
			case <-ctx.Done():
				result.Error = ctx.Err()
				return result
			case <-time.After(backoff):
				// Continue with retry
			}
		}

		attemptResult := otr.runSingleTest(ctx, packagePath)
		result.Duration = attemptResult.Duration
		result.Output = attemptResult.Output
		result.Error = attemptResult.Error
		result.TimedOut = attemptResult.TimedOut
		result.RetryAttempts = attempt

		if attemptResult.Success {
			result.Success = true
			return result
		}

		// Don't retry on context cancellation or certain errors
		if attemptResult.Error == context.DeadlineExceeded ||
			attemptResult.Error == context.Canceled ||
			strings.Contains(attemptResult.Output, "build failed") {
			break
		}
	}

	return result
}

// runSingleTest executes a single test run
func (otr *OptimizedTestRunner) runSingleTest(ctx context.Context, packagePath string) TestExecutionResult {
	// Create context with timeout
	testCtx, cancel := context.WithTimeout(ctx, otr.defaultTimeout)
	defer cancel()

	// Build test command
	args := []string{
		"test",
		"-v",
		"-timeout=" + otr.defaultTimeout.String(),
		"-run", "Integration",
		packagePath,
	}

	cmd := exec.CommandContext(testCtx, "go", args...)
	startTime := time.Now()

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	result := TestExecutionResult{
		Package:  packagePath,
		Duration: duration,
		Output:   string(output),
	}

	if err != nil {
		result.Error = err
		// Check if timeout occurred
		if testCtx.Err() == context.DeadlineExceeded {
			result.TimedOut = true
		}
	} else {
		result.Success = true
	}

	return result
}

// waitForDependencies ensures all registered services are healthy
func (otr *OptimizedTestRunner) waitForDependencies(ctx context.Context) error {
	otr.mu.RLock()
	checkers := make([]HealthChecker, 0, len(otr.healthChecks))
	for _, checker := range otr.healthChecks {
		checkers = append(checkers, checker)
	}
	otr.mu.RUnlock()

	if len(checkers) == 0 {
		return nil
	}

	otr.logger.Info(fmt.Sprintf("Waiting for %d service dependencies", len(checkers)))

	// Wait for all services concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, len(checkers))

	for _, checker := range checkers {
		wg.Add(1)
		go func(hc HealthChecker) {
			defer wg.Done()
			if err := hc.WaitForReady(ctx, 30*time.Second); err != nil {
				errChan <- fmt.Errorf("service %s not ready: %w", hc.GetServiceName(), err)
			}
		}(checker)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	otr.logger.Info("All service dependencies are ready")
	return nil
}

// GetOptimizedPackageList returns a filtered list of packages for testing
func (otr *OptimizedTestRunner) GetOptimizedPackageList() []string {
	// Focus on packages most likely to have integration tests
	return []string{
		"./pkg/testing/load/...",
		"./pkg/tree/...",
		"./pkg/replication/...",
		"./pkg/network/...",
		"./pkg/client/...",
	}
}

// RegistryHealthChecker implements health checking for Docker registry
type RegistryHealthChecker struct {
	registryURL string
	logger      log.Logger
}

// NewRegistryHealthChecker creates a health checker for Docker registry
func NewRegistryHealthChecker(registryURL string, logger log.Logger) *RegistryHealthChecker {
	return &RegistryHealthChecker{
		registryURL: registryURL,
		logger:      logger,
	}
}

func (rhc *RegistryHealthChecker) GetServiceName() string {
	return "registry"
}

func (rhc *RegistryHealthChecker) IsHealthy(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "curl", "-f", "-s", rhc.registryURL+"/v2/")
	return cmd.Run()
}

func (rhc *RegistryHealthChecker) WaitForReady(ctx context.Context, timeout time.Duration) error {
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-checkCtx.Done():
			return fmt.Errorf("registry health check timed out after %v", timeout)
		case <-ticker.C:
			if err := rhc.IsHealthy(checkCtx); err == nil {
				rhc.logger.Info("Registry is healthy")
				return nil
			}
		}
	}
}
