package replication

// ReplicationRule defines a rule for image replication
type ReplicationRule struct {
	// SourceRegistry is the source registry (e.g., "ecr" or "gcr")
	SourceRegistry string

	// SourceRepository is the source repository pattern (supports wildcards)
	SourceRepository string

	// DestinationRegistry is the destination registry (e.g., "ecr" or "gcr")
	DestinationRegistry string

	// DestinationRepository is the destination repository pattern (supports wildcards)
	DestinationRepository string

	// TagFilter is a pattern to filter which tags to replicate (supports wildcards)
	TagFilter string

	// Schedule is a cron expression for scheduled replication (empty for manual only)
	Schedule string

	// IncludeTags is a list of tag patterns to include (supports wildcards)
	IncludeTags []string

	// ExcludeTags is a list of tag patterns to exclude (supports wildcards)
	ExcludeTags []string

	// ForceOverwrite controls whether existing images should be overwritten
	ForceOverwrite bool
}

// ReplicationConfig holds the configuration for replication
type ReplicationConfig struct {
	// Rules contains all replication rules
	Rules []ReplicationRule

	// MaxConcurrentReplications is the maximum number of concurrent replications
	MaxConcurrentReplications int

	// RetryCount is the number of times to retry failed operations
	RetryCount int
}

// ShouldReplicate determines if an image should be replicated based on the config rules
func (c *ReplicationConfig) ShouldReplicate(repository, tag string) bool {
	// Check if any rule matches
	for _, rule := range c.Rules {
		if ShouldReplicate(rule, repository, tag) {
			return true
		}
	}
	return false
}

// GetDestinationRepository returns the destination repository based on the source repository
func (c *ReplicationConfig) GetDestinationRepository(sourceRepo string) (string, bool) {
	// Check all rules for a matching source repository
	for _, rule := range c.Rules {
		if MatchPattern(rule.SourceRepository, sourceRepo) {
			// Perform wildcard substitution to get the actual destination
			return GetDestinationRepository(rule, sourceRepo), true
		}
	}
	return "", false
}
