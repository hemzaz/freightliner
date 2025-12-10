package replication

import (
	"path/filepath"
	"regexp"
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
	// Case 1: Source and destination repository patterns are the same
	if rule.SourceRepository == rule.DestinationRepository {
		return sourceRepository
	}

	// Case 2: Source has wildcards and destination has substitution patterns
	if isWildcardPattern(rule.SourceRepository) && isSubstitutionPattern(rule.DestinationRepository) {
		substitutedRepo := substituteWildcard(rule.SourceRepository, rule.DestinationRepository, sourceRepository)
		if substitutedRepo != "" {
			return substitutedRepo
		}
	}

	// Case 3: Many-to-one mapping (source has wildcard, destination doesn't have wildcards or substitutions)
	if isWildcardPattern(rule.SourceRepository) && !isWildcardPattern(rule.DestinationRepository) && !isSubstitutionPattern(rule.DestinationRepository) {
		return rule.DestinationRepository
	}

	// Case 4: Both patterns have wildcards, try to substitute
	if isWildcardPattern(rule.SourceRepository) && isWildcardPattern(rule.DestinationRepository) {
		substitutedRepo := substituteWildcard(rule.SourceRepository, rule.DestinationRepository, sourceRepository)
		if substitutedRepo != "" {
			return substitutedRepo
		}
	}

	// Default case: Just use the destination as is
	return rule.DestinationRepository
}

// isWildcardPattern checks if a pattern contains a wildcard
func isWildcardPattern(pattern string) bool {
	return strings.Contains(pattern, "*")
}

// isSubstitutionPattern checks if a pattern contains substitution placeholders ($1, $2, etc.)
func isSubstitutionPattern(pattern string) bool {
	// Check for $1, $2, $3, etc.
	for i := 1; i <= 9; i++ {
		if strings.Contains(pattern, "$"+string(rune('0'+i))) {
			return true
		}
	}
	return false
}

// substituteWildcard extracts the parts matching wildcards in the source pattern
// and substitutes them into the destination pattern using $1, $2, etc.
func substituteWildcard(sourcePattern, destPattern, sourceString string) string {
	// Convert shell-style wildcards to regex pattern
	regexPattern := strings.ReplaceAll(sourcePattern, "*", "([^/]+)")
	regexPattern = "^" + regexPattern + "$"

	// Compile the regex
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		// Fallback to simple wildcard replacement for invalid regex
		return fallbackWildcardSubstitution(sourcePattern, destPattern, sourceString)
	}

	// Find matches
	matches := re.FindStringSubmatch(sourceString)
	if len(matches) < 2 {
		// No capture groups found, fallback to simple substitution
		return fallbackWildcardSubstitution(sourcePattern, destPattern, sourceString)
	}

	// Substitute capture groups in destination pattern
	result := destPattern
	for i := 1; i < len(matches); i++ {
		placeholder := "$" + string(rune('0'+i))
		result = strings.ReplaceAll(result, placeholder, matches[i])
	}

	// Also handle simple * replacement for backward compatibility
	if strings.Contains(result, "*") && len(matches) >= 2 {
		result = strings.ReplaceAll(result, "*", matches[1])
	}

	return result
}

// fallbackWildcardSubstitution provides simple single-wildcard substitution for backward compatibility
func fallbackWildcardSubstitution(sourcePattern, destPattern, sourceString string) string {
	// Split the source pattern at the wildcard
	parts := strings.Split(sourcePattern, "*")
	if len(parts) != 2 {
		// We only handle simple patterns with a single wildcard
		return ""
	}

	prefix := parts[0]
	suffix := parts[1]

	// Check if the source string matches the pattern
	if !strings.HasPrefix(sourceString, prefix) || !strings.HasSuffix(sourceString, suffix) {
		return ""
	}

	// Extract the middle part that matched the wildcard
	middle := sourceString[len(prefix) : len(sourceString)-len(suffix)]

	// Replace the wildcard in the destination pattern
	return strings.Replace(destPattern, "*", middle, 1)
}
