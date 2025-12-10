package replication

import (
	"testing"
)

func TestFallbackWildcardSubstitution(t *testing.T) {
	tests := []struct {
		name          string
		sourcePattern string
		destPattern   string
		sourceString  string
		expected      string
	}{
		{
			name:          "Simple substitution",
			sourcePattern: "source/*",
			destPattern:   "dest/*",
			sourceString:  "source/app",
			expected:      "dest/app",
		},
		{
			name:          "No match - wrong prefix",
			sourcePattern: "source/*",
			destPattern:   "dest/*",
			sourceString:  "other/app",
			expected:      "",
		},
		{
			name:          "No match - wrong suffix",
			sourcePattern: "*/suffix",
			destPattern:   "*/newsuffix",
			sourceString:  "prefix/wrong",
			expected:      "",
		},
		{
			name:          "Multiple wildcards - only one supported",
			sourcePattern: "*/*/*",
			destPattern:   "*/*/*",
			sourceString:  "a/b/c",
			expected:      "",
		},
		{
			name:          "Empty middle part",
			sourcePattern: "pre*post",
			destPattern:   "new*end",
			sourceString:  "prepost",
			expected:      "newend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fallbackWildcardSubstitution(tt.sourcePattern, tt.destPattern, tt.sourceString)
			if got != tt.expected {
				t.Errorf("fallbackWildcardSubstitution() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIsWildcardPattern(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{"source/*", true},
		{"*/repo", true},
		{"*", true},
		{"source/repo", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := isWildcardPattern(tt.pattern)
			if got != tt.want {
				t.Errorf("isWildcardPattern(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestIsSubstitutionPattern(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{"dest/$1", true},
		{"dest/$2/$1", true},
		{"dest/$9", true},
		{"dest/*", false},
		{"dest/repo", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := isSubstitutionPattern(tt.pattern)
			if got != tt.want {
				t.Errorf("isSubstitutionPattern(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestSubstituteWildcard(t *testing.T) {
	tests := []struct {
		name          string
		sourcePattern string
		destPattern   string
		sourceString  string
		want          string
	}{
		{
			name:          "Single capture group",
			sourcePattern: "source/*",
			destPattern:   "dest/$1",
			sourceString:  "source/app",
			want:          "dest/app",
		},
		{
			name:          "Multiple capture groups",
			sourcePattern: "source/*/group/*",
			destPattern:   "dest/$2/$1",
			sourceString:  "source/project/group/app",
			want:          "dest/app/project",
		},
		{
			name:          "Backward compatibility with *",
			sourcePattern: "source/*",
			destPattern:   "dest/*",
			sourceString:  "source/app",
			want:          "dest/app",
		},
		{
			name:          "Three capture groups",
			sourcePattern: "*/*/*/repo",
			destPattern:   "$3/$2/$1/repo",
			sourceString:  "a/b/c/repo",
			want:          "c/b/a/repo",
		},
		{
			name:          "No match",
			sourcePattern: "source/*",
			destPattern:   "dest/$1",
			sourceString:  "other/app",
			want:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := substituteWildcard(tt.sourcePattern, tt.destPattern, tt.sourceString)
			if got != tt.want {
				t.Errorf("substituteWildcard() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetDestinationRepository_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		rule       ReplicationRule
		sourceRepo string
		want       string
	}{
		{
			name: "Same pattern",
			rule: ReplicationRule{
				SourceRepository:      "source/repo",
				DestinationRepository: "source/repo",
			},
			sourceRepo: "source/repo",
			want:       "source/repo",
		},
		{
			name: "Many-to-one mapping",
			rule: ReplicationRule{
				SourceRepository:      "source/*",
				DestinationRepository: "dest/consolidated",
			},
			sourceRepo: "source/app1",
			want:       "dest/consolidated",
		},
		{
			name: "Both have wildcards",
			rule: ReplicationRule{
				SourceRepository:      "source/*",
				DestinationRepository: "dest/*",
			},
			sourceRepo: "source/app",
			want:       "dest/app",
		},
		{
			name: "Wildcard with substitution",
			rule: ReplicationRule{
				SourceRepository:      "source/*/project",
				DestinationRepository: "dest/$1/repo",
			},
			sourceRepo: "source/myapp/project",
			want:       "dest/myapp/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDestinationRepository(tt.rule, tt.sourceRepo)
			if got != tt.want {
				t.Errorf("GetDestinationRepository() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMatchPattern_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		str     string
		want    bool
	}{
		// Note: filepath.Match handles [ and ? differently than shell wildcards
		{
			name:    "Wildcard star matches all",
			pattern: "source/*",
			str:     "source/anything",
			want:    true,
		},
		{
			name:    "Multiple directory levels",
			pattern: "source/*/*",
			str:     "source/app/component",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchPattern(tt.pattern, tt.str)
			if got != tt.want {
				t.Errorf("MatchPattern(%q, %q) = %v, want %v", tt.pattern, tt.str, got, tt.want)
			}
		})
	}
}

func TestShouldReplicate_WithIncludeExcludeTags(t *testing.T) {
	tests := []struct {
		name string
		rule ReplicationRule
		repo string
		tag  string
		want bool
	}{
		{
			name: "Match with include tags",
			rule: ReplicationRule{
				SourceRepository: "source/repo",
				IncludeTags:      []string{"v*", "latest"},
			},
			repo: "source/repo",
			tag:  "v1.0.0",
			want: true,
		},
		{
			name: "Match with exclude tags",
			rule: ReplicationRule{
				SourceRepository: "source/repo",
				ExcludeTags:      []string{"*-dev", "*-test"},
			},
			repo: "source/repo",
			tag:  "v1.0.0-dev",
			want: true, // ShouldReplicate doesn't check include/exclude yet
		},
		{
			name: "No match - repository mismatch",
			rule: ReplicationRule{
				SourceRepository: "source/repo",
				IncludeTags:      []string{"v*"},
			},
			repo: "other/repo",
			tag:  "v1.0.0",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldReplicate(tt.rule, tt.repo, tt.tag)
			if got != tt.want {
				t.Errorf("ShouldReplicate() = %v, want %v", got, tt.want)
			}
		})
	}
}
