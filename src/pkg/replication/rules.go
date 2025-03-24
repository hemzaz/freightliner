package replication

import (
	"path/filepath"
	"strings"
)

// MatchPattern checks if a string matches a pattern (supporting wildcards)
func MatchPattern(pattern, str string) bool {
	// If the pattern has no wildcards, do a direct comparison
	if !strings.Contains(pattern, "*") {
		return pattern == str
	}
	
	// Use filepath.Match for wildcard matching
	matched, err := filepath.Match(pattern, str)
	if err != nil {
		// Invalid pattern
		return false
	}
	
	return matched
}

// ShouldReplicate determines if an image should be replicated based on a rule
func ShouldReplicate(rule ReplicationRule, repository, tag string) bool {
	// Check if the repository matches the source repository pattern
	if !MatchPattern(rule.SourceRepository, repository) {
		return false
	}
	
	// Check if the tag matches the tag filter
	if rule.TagFilter != "" && !MatchPattern(rule.TagFilter, tag) {
		return false
	}
	
	return true
}

// GetDestinationRepository determines the destination repository based on a rule
func GetDestinationRepository(rule ReplicationRule, sourceRepository string) string {
	// If the source and destination repository patterns are the same,
	// use the source repository as is
	if rule.SourceRepository == rule.DestinationRepository {
		return sourceRepository
	}
	
	// If the source repository pattern has a wildcard and the destination doesn't,
	// this is a many-to-one mapping, use the destination as is
	if strings.Contains(rule.SourceRepository, "*") && 
	   !strings.Contains(rule.DestinationRepository, "*") {
		return rule.DestinationRepository
	}
	
	// If both patterns have wildcards, try to substitute
	if strings.Contains(rule.SourceRepository, "*") && 
	   strings.Contains(rule.DestinationRepository, "*") {
		// Find what matched the wildcard in the source
		parts := strings.Split(rule.SourceRepository, "*")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]
			
			if strings.HasPrefix(sourceRepository, prefix) && 
			   strings.HasSuffix(sourceRepository, suffix) {
				// Extract the middle part that matched the wildcard
				middle := sourceRepository[len(prefix):len(sourceRepository)-len(suffix)]
				
				// Replace the wildcard in the destination pattern
				return strings.Replace(rule.DestinationRepository, "*", middle, 1)
			}
		}
	}
	
	// If no special case applies, just use the destination as is
	return rule.DestinationRepository
}
