package resilience

import (
	"context"
	"fmt"

	"freightliner/pkg/helper/log"
)

// FallbackFunc is a fallback operation
type FallbackFunc func(ctx context.Context) error

// DegradationPolicy manages graceful degradation strategies
type DegradationPolicy struct {
	name      string
	primary   func(ctx context.Context) error
	fallbacks []Fallback
	logger    log.Logger
}

// Fallback represents a fallback operation
type Fallback struct {
	// Name of the fallback
	Name string
	// Try is the fallback operation to attempt
	Try FallbackFunc
	// Condition determines if this fallback should be attempted
	Condition func(error) bool
}

// NewDegradationPolicy creates a new degradation policy
func NewDegradationPolicy(name string, primary func(ctx context.Context) error, logger log.Logger) *DegradationPolicy {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &DegradationPolicy{
		name:      name,
		primary:   primary,
		fallbacks: make([]Fallback, 0),
		logger:    logger,
	}
}

// AddFallback adds a fallback to the policy
func (d *DegradationPolicy) AddFallback(fallback Fallback) *DegradationPolicy {
	d.fallbacks = append(d.fallbacks, fallback)
	return d
}

// AddSimpleFallback adds a simple fallback without conditions
func (d *DegradationPolicy) AddSimpleFallback(name string, fn FallbackFunc) *DegradationPolicy {
	return d.AddFallback(Fallback{
		Name: name,
		Try:  fn,
		Condition: func(error) bool {
			return true // Always try this fallback
		},
	})
}

// Execute attempts the primary operation and falls back on failure
func (d *DegradationPolicy) Execute(ctx context.Context) error {
	// Try primary operation
	d.logger.WithFields(map[string]interface{}{
		"policy": d.name,
	}).Debug("Attempting primary operation")

	err := d.primary(ctx)
	if err == nil {
		return nil
	}

	d.logger.WithError(err).WithFields(map[string]interface{}{
		"policy": d.name,
	}).Warn("Primary operation failed, attempting fallbacks")

	// Try fallbacks in order
	for i, fallback := range d.fallbacks {
		// Check if this fallback should be attempted
		if fallback.Condition != nil && !fallback.Condition(err) {
			d.logger.WithFields(map[string]interface{}{
				"policy":   d.name,
				"fallback": fallback.Name,
			}).Debug("Skipping fallback (condition not met)")
			continue
		}

		d.logger.WithFields(map[string]interface{}{
			"policy":   d.name,
			"fallback": fallback.Name,
			"index":    i,
		}).Info("Attempting fallback")

		fallbackErr := fallback.Try(ctx)
		if fallbackErr == nil {
			d.logger.WithFields(map[string]interface{}{
				"policy":   d.name,
				"fallback": fallback.Name,
			}).Info("Fallback succeeded")
			return nil
		}

		d.logger.WithError(fallbackErr).WithFields(map[string]interface{}{
			"policy":   d.name,
			"fallback": fallback.Name,
		}).Warn("Fallback failed")
	}

	// All fallbacks failed
	return fmt.Errorf("all operations failed for policy '%s': %w", d.name, err)
}

// ExecuteWithResult attempts operations that return a result
func ExecuteWithResult[T any](
	ctx context.Context,
	policy *DegradationPolicy,
	primary func(ctx context.Context) (T, error),
	fallbacks []FallbackWithResult[T],
	logger log.Logger,
) (T, error) {
	var zero T

	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Try primary operation
	logger.WithFields(map[string]interface{}{
		"policy": policy.name,
	}).Debug("Attempting primary operation")

	result, err := primary(ctx)
	if err == nil {
		return result, nil
	}

	logger.WithError(err).WithFields(map[string]interface{}{
		"policy": policy.name,
	}).Warn("Primary operation failed, attempting fallbacks")

	// Try fallbacks in order
	for i, fallback := range fallbacks {
		// Check if this fallback should be attempted
		if fallback.Condition != nil && !fallback.Condition(err) {
			logger.WithFields(map[string]interface{}{
				"policy":   policy.name,
				"fallback": fallback.Name,
			}).Debug("Skipping fallback (condition not met)")
			continue
		}

		logger.WithFields(map[string]interface{}{
			"policy":   policy.name,
			"fallback": fallback.Name,
			"index":    i,
		}).Info("Attempting fallback")

		fallbackResult, fallbackErr := fallback.Try(ctx)
		if fallbackErr == nil {
			logger.WithFields(map[string]interface{}{
				"policy":   policy.name,
				"fallback": fallback.Name,
			}).Info("Fallback succeeded")
			return fallbackResult, nil
		}

		logger.WithError(fallbackErr).WithFields(map[string]interface{}{
			"policy":   policy.name,
			"fallback": fallback.Name,
		}).Warn("Fallback failed")
	}

	// All fallbacks failed
	return zero, fmt.Errorf("all operations failed for policy '%s': %w", policy.name, err)
}

// FallbackWithResult is a fallback that returns a result
type FallbackWithResult[T any] struct {
	// Name of the fallback
	Name string
	// Try is the fallback operation to attempt
	Try func(ctx context.Context) (T, error)
	// Condition determines if this fallback should be attempted
	Condition func(error) bool
}

// DegradationManager manages multiple degradation policies
type DegradationManager struct {
	policies map[string]*DegradationPolicy
	logger   log.Logger
}

// NewDegradationManager creates a new degradation manager
func NewDegradationManager(logger log.Logger) *DegradationManager {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &DegradationManager{
		policies: make(map[string]*DegradationPolicy),
		logger:   logger,
	}
}

// RegisterPolicy registers a new degradation policy
func (m *DegradationManager) RegisterPolicy(policy *DegradationPolicy) {
	m.policies[policy.name] = policy
}

// Execute executes a policy by name
func (m *DegradationManager) Execute(ctx context.Context, policyName string) error {
	policy, exists := m.policies[policyName]
	if !exists {
		return fmt.Errorf("degradation policy '%s' not found", policyName)
	}
	return policy.Execute(ctx)
}

// GetPolicy retrieves a policy by name
func (m *DegradationManager) GetPolicy(name string) (*DegradationPolicy, bool) {
	policy, exists := m.policies[name]
	return policy, exists
}

// Common degradation strategies

// NetworkProtocolFallback creates a fallback for HTTP protocol versions
func NetworkProtocolFallback(
	tryHTTP3 func(ctx context.Context) error,
	tryHTTP2 func(ctx context.Context) error,
	tryHTTP1 func(ctx context.Context) error,
	logger log.Logger,
) *DegradationPolicy {
	policy := NewDegradationPolicy("network-protocol", tryHTTP3, logger)

	policy.AddSimpleFallback("http2", tryHTTP2)
	policy.AddSimpleFallback("http1.1", tryHTTP1)

	return policy
}

// RegistryMirrorFallback creates a fallback for registry mirrors
func RegistryMirrorFallback(
	tryPrimary func(ctx context.Context) error,
	mirrors []func(ctx context.Context) error,
	logger log.Logger,
) *DegradationPolicy {
	policy := NewDegradationPolicy("registry-mirror", tryPrimary, logger)

	for i, mirror := range mirrors {
		name := fmt.Sprintf("mirror-%d", i+1)
		policy.AddSimpleFallback(name, mirror)
	}

	return policy
}

// SyncStrategyFallback creates a fallback for sync strategies
func SyncStrategyFallback(
	tryFullSync func(ctx context.Context) error,
	tryIncrementalSync func(ctx context.Context) error,
	tryManifestOnly func(ctx context.Context) error,
	logger log.Logger,
) *DegradationPolicy {
	policy := NewDegradationPolicy("sync-strategy", tryFullSync, logger)

	policy.AddSimpleFallback("incremental", tryIncrementalSync)
	policy.AddSimpleFallback("manifest-only", tryManifestOnly)

	return policy
}
