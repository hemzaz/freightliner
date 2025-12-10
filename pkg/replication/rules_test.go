package replication

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		value   string
		want    bool
	}{
		{
			name:    "Exact match",
			pattern: "repository",
			value:   "repository",
			want:    true,
		},
		{
			name:    "Simple wildcard",
			pattern: "*",
			value:   "any-value",
			want:    true,
		},
		{
			name:    "Prefix wildcard",
			pattern: "prefix-*",
			value:   "prefix-suffix",
			want:    true,
		},
		{
			name:    "Prefix wildcard - no match",
			pattern: "prefix-*",
			value:   "different-suffix",
			want:    false,
		},
		{
			name:    "Suffix wildcard",
			pattern: "*-suffix",
			value:   "prefix-suffix",
			want:    true,
		},
		{
			name:    "Suffix wildcard - no match",
			pattern: "*-suffix",
			value:   "prefix-different",
			want:    false,
		},
		{
			name:    "Middle wildcard",
			pattern: "pre*post",
			value:   "pre-middle-post",
			want:    true,
		},
		{
			name:    "Middle wildcard - no match",
			pattern: "pre*post",
			value:   "pre-different",
			want:    false,
		},
		{
			name:    "Multiple wildcards",
			pattern: "*part*of*string*",
			value:   "this-part-is-of-a-string-value",
			want:    true,
		},
		{
			name:    "Multiple wildcards - no match",
			pattern: "*part*of*string*",
			value:   "this-does-not-contain-required-parts",
			want:    false,
		},
		{
			name:    "Empty pattern",
			pattern: "",
			value:   "any-value",
			want:    false,
		},
		{
			name:    "Empty value",
			pattern: "pattern",
			value:   "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchPattern(tt.pattern, tt.value); got != tt.want {
				t.Errorf("MatchPattern(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
			}
		})
	}
}

func TestShouldReplicate(t *testing.T) {
	tests := []struct {
		name            string
		rule            ReplicationRule
		repoName        string
		tagName         string
		expectReplicate bool
	}{
		{
			name: "Exact match repo and tag",
			rule: ReplicationRule{
				SourceRepository:      "source/repo",
				DestinationRepository: "dest/repo",
				TagFilter:             "latest",
			},
			repoName:        "source/repo",
			tagName:         "latest",
			expectReplicate: true,
		},
		{
			name: "Wildcard repo match",
			rule: ReplicationRule{
				SourceRepository:      "source/*",
				DestinationRepository: "dest/repo",
				TagFilter:             "latest",
			},
			repoName:        "source/any-repo",
			tagName:         "latest",
			expectReplicate: true,
		},
		{
			name: "Wildcard tag match",
			rule: ReplicationRule{
				SourceRepository:      "source/repo",
				DestinationRepository: "dest/repo",
				TagFilter:             "v*",
			},
			repoName:        "source/repo",
			tagName:         "v1.0.0",
			expectReplicate: true,
		},
		{
			name: "Repo doesn't match",
			rule: ReplicationRule{
				SourceRepository:      "source/repo",
				DestinationRepository: "dest/repo",
				TagFilter:             "latest",
			},
			repoName:        "different/repo",
			tagName:         "latest",
			expectReplicate: false,
		},
		{
			name: "Tag doesn't match",
			rule: ReplicationRule{
				SourceRepository:      "source/repo",
				DestinationRepository: "dest/repo",
				TagFilter:             "latest",
			},
			repoName:        "source/repo",
			tagName:         "v1.0.0",
			expectReplicate: false,
		},
		{
			name: "Empty tag filter matches all tags",
			rule: ReplicationRule{
				SourceRepository:      "source/repo",
				DestinationRepository: "dest/repo",
				TagFilter:             "",
			},
			repoName:        "source/repo",
			tagName:         "any-tag",
			expectReplicate: true,
		},
		{
			name: "Multiple wildcards in repo and tag",
			rule: ReplicationRule{
				SourceRepository:      "source/**/nested/*",
				DestinationRepository: "dest/repo",
				TagFilter:             "*-stable-*",
			},
			repoName:        "source/middle/nested/repo",
			tagName:         "v1-stable-release",
			expectReplicate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ReplicationConfig{
				Rules: []ReplicationRule{tt.rule},
			}

			if got := config.ShouldReplicate(tt.repoName, tt.tagName); got != tt.expectReplicate {
				t.Errorf("ShouldReplicate(%q, %q) = %v, want %v", tt.repoName, tt.tagName, got, tt.expectReplicate)
			}
		})
	}
}

func TestGetDestinationRepository(t *testing.T) {
	tests := []struct {
		name         string
		rules        []ReplicationRule
		sourceRepo   string
		expectedRepo string
		expectFound  bool
	}{
		{
			name: "Exact match",
			rules: []ReplicationRule{
				{
					SourceRepository:      "source/repo",
					DestinationRepository: "dest/repo",
				},
			},
			sourceRepo:   "source/repo",
			expectedRepo: "dest/repo",
			expectFound:  true,
		},
		{
			name: "Wildcard match with substitution",
			rules: []ReplicationRule{
				{
					SourceRepository:      "source/*",
					DestinationRepository: "dest/$1",
				},
			},
			sourceRepo:   "source/app",
			expectedRepo: "dest/app",
			expectFound:  true,
		},
		{
			name: "Complex wildcard with multiple captures",
			rules: []ReplicationRule{
				{
					SourceRepository:      "source/*/group/*",
					DestinationRepository: "dest/$2/$1",
				},
			},
			sourceRepo:   "source/project/group/app",
			expectedRepo: "dest/app/project",
			expectFound:  true,
		},
		{
			name: "No matching rule",
			rules: []ReplicationRule{
				{
					SourceRepository:      "source/repo",
					DestinationRepository: "dest/repo",
				},
			},
			sourceRepo:   "other/repo",
			expectedRepo: "",
			expectFound:  false,
		},
		{
			name: "First matching rule is used",
			rules: []ReplicationRule{
				{
					SourceRepository:      "source/*",
					DestinationRepository: "first/$1",
				},
				{
					SourceRepository:      "source/app",
					DestinationRepository: "second/app",
				},
			},
			sourceRepo:   "source/app",
			expectedRepo: "first/app", // First rule matches
			expectFound:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ReplicationConfig{
				Rules: tt.rules,
			}

			repo, found := config.GetDestinationRepository(tt.sourceRepo)
			if found != tt.expectFound {
				t.Errorf("GetDestinationRepository(%q) found = %v, want %v", tt.sourceRepo, found, tt.expectFound)
			}

			if repo != tt.expectedRepo {
				t.Errorf("GetDestinationRepository(%q) = %q, want %q", tt.sourceRepo, repo, tt.expectedRepo)
			}
		})
	}
}

func TestRuleScheduling(t *testing.T) {
	// Placeholder for future schedule-related tests
	tests := []struct {
		name      string
		rule      ReplicationRule
		expectNow bool
	}{
		{
			name: "Empty schedule - run now",
			rule: ReplicationRule{
				Schedule: "",
			},
			expectNow: true,
		},
		{
			name: "Daily schedule - not now",
			rule: ReplicationRule{
				Schedule: "0 0 * * *", // midnight every day
			},
			expectNow: false,
		},
		{
			name: "Every minute - run now",
			rule: ReplicationRule{
				Schedule: "* * * * *", // every minute
			},
			expectNow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldRunNow := tt.rule.Schedule == "" || tt.rule.Schedule == "* * * * *"

			if shouldRunNow != tt.expectNow {
				t.Errorf("Rule with schedule %q, expectNow = %v, want %v",
					tt.rule.Schedule, shouldRunNow, tt.expectNow)
			}
		})
	}
}
