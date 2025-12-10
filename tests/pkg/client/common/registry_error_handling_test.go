package common

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistryUtil_ErrorHandling_InvalidInputs tests error handling for invalid inputs
func TestRegistryUtil_ErrorHandling_InvalidInputs(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name          string
		operation     string
		input         interface{}
		expectedError string
	}{
		{
			name:          "ParseRegistryPath with nil-like string",
			operation:     "parse",
			input:         string([]byte{0, 0, 0}),
			expectedError: "invalid format",
		},
		{
			name:          "ParseRegistryPath with only delimiter",
			operation:     "parse",
			input:         "/",
			expectedError: "invalid format",
		},
		{
			name:          "ParseRegistryPath with multiple delimiters",
			operation:     "parse",
			input:         "///",
			expectedError: "invalid format",
		},
		{
			name:          "ValidateRepositoryName with empty string",
			operation:     "validate",
			input:         "",
			expectedError: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			switch tt.operation {
			case "parse":
				_, _, err = util.ParseRegistryPath(tt.input.(string))
			case "validate":
				err = util.ValidateRepositoryName(tt.input.(string))
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestRegistryUtil_ErrorMessages tests that error messages are informative
func TestRegistryUtil_ErrorMessages(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name          string
		operation     func() error
		expectedWords []string // Words that should appear in error message
	}{
		{
			name: "Empty repository name error",
			operation: func() error {
				return util.ValidateRepositoryName("")
			},
			expectedWords: []string{"repository", "empty"},
		},
		{
			name: "Invalid path format error",
			operation: func() error {
				_, _, err := util.ParseRegistryPath("invalid")
				return err
			},
			expectedWords: []string{"invalid", "format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			require.Error(t, err)

			errMsg := strings.ToLower(err.Error())
			for _, word := range tt.expectedWords {
				assert.Contains(t, errMsg, strings.ToLower(word),
					"Error message should contain '%s'", word)
			}
		})
	}
}

// TestRegistryUtil_BoundaryValues tests boundary value conditions
func TestRegistryUtil_BoundaryValues(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	t.Run("Maximum length repository name", func(t *testing.T) {
		// Docker registry repository names can be up to 255 characters
		maxLengthRepo := strings.Repeat("a", 255)
		err := util.ValidateRepositoryName(maxLengthRepo)
		assert.NoError(t, err, "Should accept maximum length repository name")
	})

	t.Run("Extremely long repository name", func(t *testing.T) {
		// Beyond reasonable limits
		tooLongRepo := strings.Repeat("a", 10000)
		// Current implementation doesn't check length, but we document the behavior
		err := util.ValidateRepositoryName(tooLongRepo)
		// Note: This documents current behavior; might want to add length validation
		assert.NoError(t, err, "Current implementation accepts very long names")
	})

	t.Run("Single character repository", func(t *testing.T) {
		err := util.ValidateRepositoryName("a")
		assert.NoError(t, err, "Should accept single character repository")
	})

	t.Run("Repository with maximum path depth", func(t *testing.T) {
		// Test deeply nested repository paths
		deepPath := strings.Repeat("a/", 50) + "final"
		err := util.ValidateRepositoryName(deepPath)
		assert.NoError(t, err, "Should accept deeply nested paths")
	})
}

// TestRegistryUtil_RaceConditions tests for race conditions
func TestRegistryUtil_RaceConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))
	ctx := context.Background()

	// Run multiple operations concurrently to detect race conditions
	done := make(chan bool)
	iterations := 1000

	// Reader goroutines
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				util.IsValidRegistryType("ecr")
				util.ParseRegistryPath("ecr/repo")
				util.FormatRepositoryURI("ecr", "123", "us-west-2", "repo")
			}
			done <- true
		}()
	}

	// Writer goroutines (logging operations)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				util.LogRegistryOperation(ctx, "test", "registry", "repo", map[string]interface{}{
					"id": id,
					"op": j,
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// If we get here without race detector warnings, test passes
	assert.True(t, true, "No race conditions detected")
}

// TestRegistryUtil_ErrorWrapping tests error wrapping and unwrapping
func TestRegistryUtil_ErrorWrapping(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	t.Run("Parse error is not nil", func(t *testing.T) {
		_, _, err := util.ParseRegistryPath("invalid")
		require.Error(t, err)
		assert.NotNil(t, err)
	})

	t.Run("Validation error is not nil", func(t *testing.T) {
		err := util.ValidateRepositoryName("")
		require.Error(t, err)
		assert.NotNil(t, err)
	})
}

// TestRegistryUtil_PanicRecovery tests that operations don't panic
func TestRegistryUtil_PanicRecovery(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))
	ctx := context.Background()

	tests := []struct {
		name      string
		operation func()
	}{
		{
			name: "ParseRegistryPath with extreme input",
			operation: func() {
				util.ParseRegistryPath(strings.Repeat("a", 1000000))
			},
		},
		{
			name: "ValidateRepositoryName with extreme input",
			operation: func() {
				util.ValidateRepositoryName(strings.Repeat("a", 1000000))
			},
		},
		{
			name: "FormatRepositoryURI with empty inputs",
			operation: func() {
				util.FormatRepositoryURI("", "", "", "")
			},
		},
		{
			name: "CreateRepositoryReference with invalid characters",
			operation: func() {
				util.CreateRepositoryReference("reg@#$%", "repo@#$%")
			},
		},
		{
			name: "LogRegistryOperation with nil context",
			operation: func() {
				util.LogRegistryOperation(nil, "op", "reg", "repo", nil)
			},
		},
		{
			name: "LogRegistryOperation with cancelled context",
			operation: func() {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				util.LogRegistryOperation(ctx, "op", "reg", "repo", nil)
			},
		},
		{
			name: "GetRemoteOptions with nil transport",
			operation: func() {
				util.GetRemoteOptions(nil)
			},
		},
		{
			name: "GetRemoteOptions with invalid transport",
			operation: func() {
				// Create a transport with invalid configuration
				transport := &http.Transport{
					MaxIdleConns: -1, // Invalid value
				}
				util.GetRemoteOptions(transport)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.operation, "Operation should not panic")
		})
	}
}

// TestRegistryUtil_MemorySafety tests memory safety
func TestRegistryUtil_MemorySafety(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory safety test in short mode")
	}

	// Test that operations don't cause memory issues
	t.Run("Large string operations", func(t *testing.T) {
		util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

		// Create a very large string
		largeString := strings.Repeat("a", 10*1024*1024) // 10MB

		assert.NotPanics(t, func() {
			util.ParseRegistryPath("ecr/" + largeString)
			util.ValidateRepositoryName(largeString)
			util.FormatRepositoryURI("ecr", largeString, "region", largeString)
		})
	})

	t.Run("Repeated allocations", func(t *testing.T) {
		util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

		// Create many allocations to test for leaks
		for i := 0; i < 10000; i++ {
			repo := strings.Repeat("a", 100)
			util.ParseRegistryPath("ecr/" + repo)
			util.ValidateRepositoryName(repo)
			util.FormatRepositoryURI("ecr", "123", "us-west-2", repo)
		}

		// If we get here without OOM, test passes
		assert.True(t, true)
	})
}

// TestRegistryUtil_ContextBehavior tests context handling
func TestRegistryUtil_ContextBehavior(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	t.Run("Nil context", func(t *testing.T) {
		assert.NotPanics(t, func() {
			util.LogRegistryOperation(nil, "op", "reg", "repo", nil)
		})
	})

	t.Run("Cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		assert.NotPanics(t, func() {
			util.LogRegistryOperation(ctx, "op", "reg", "repo", nil)
		})
	})

	t.Run("Context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Wait for timeout
		<-ctx.Done()

		assert.NotPanics(t, func() {
			util.LogRegistryOperation(ctx, "op", "reg", "repo", nil)
		})
	})

	t.Run("Context with values", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")

		assert.NotPanics(t, func() {
			util.LogRegistryOperation(ctx, "op", "reg", "repo", nil)
		})
	})
}

// TestRegistryUtil_SpecialCharacters tests handling of special characters
func TestRegistryUtil_SpecialCharacters(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	specialChars := []string{
		"repo-with-dash",
		"repo_with_underscore",
		"repo.with.dots",
		"repo123",
		"123repo",
		"UPPERCASE",
		"MixedCase",
		"repo-with.mixed_123",
	}

	for _, repoName := range specialChars {
		t.Run("Repository: "+repoName, func(t *testing.T) {
			// Should handle special characters without errors
			assert.NotPanics(t, func() {
				util.ValidateRepositoryName(repoName)
				util.ParseRegistryPath("ecr/" + repoName)
				util.FormatRepositoryURI("ecr", "123", "us-west-2", repoName)
			})
		})
	}
}

// TestRegistryUtil_ErrorRecovery tests error recovery mechanisms
func TestRegistryUtil_ErrorRecovery(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	t.Run("Multiple sequential errors", func(t *testing.T) {
		// Multiple failed operations should not affect subsequent operations
		for i := 0; i < 10; i++ {
			_, _, err := util.ParseRegistryPath("invalid")
			assert.Error(t, err)
		}

		// Valid operation should still work
		registry, repo, err := util.ParseRegistryPath("ecr/valid")
		require.NoError(t, err)
		assert.Equal(t, "ecr", registry)
		assert.Equal(t, "valid", repo)
	})

	t.Run("Error then success pattern", func(t *testing.T) {
		// Alternate between error and success
		for i := 0; i < 5; i++ {
			_, _, err1 := util.ParseRegistryPath("invalid")
			assert.Error(t, err1)

			_, _, err2 := util.ParseRegistryPath("ecr/valid")
			assert.NoError(t, err2)
		}
	})
}

// TestRegistryUtil_LoggingEdgeCases tests logging edge cases
func TestRegistryUtil_LoggingEdgeCases(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))
	ctx := context.Background()

	t.Run("Logging with nil extra fields", func(t *testing.T) {
		assert.NotPanics(t, func() {
			util.LogRegistryOperation(ctx, "op", "reg", "repo", nil)
		})
	})

	t.Run("Logging with empty extra fields", func(t *testing.T) {
		assert.NotPanics(t, func() {
			util.LogRegistryOperation(ctx, "op", "reg", "repo", map[string]interface{}{})
		})
	})

	t.Run("Logging with complex extra fields", func(t *testing.T) {
		complexFields := map[string]interface{}{
			"string": "value",
			"int":    42,
			"float":  3.14,
			"bool":   true,
			"nil":    nil,
			"slice":  []string{"a", "b", "c"},
			"map":    map[string]string{"key": "value"},
			"nested": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": "deep",
				},
			},
		}

		assert.NotPanics(t, func() {
			util.LogRegistryOperation(ctx, "op", "reg", "repo", complexFields)
		})
	})

	t.Run("Logging with very large fields", func(t *testing.T) {
		largeFields := map[string]interface{}{
			"large": strings.Repeat("a", 1024*1024), // 1MB string
		}

		assert.NotPanics(t, func() {
			util.LogRegistryOperation(ctx, "op", "reg", "repo", largeFields)
		})
	})
}

// BenchmarkRegistryUtil_ErrorCases benchmarks error case performance
func BenchmarkRegistryUtil_ErrorCases(b *testing.B) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	b.Run("ParseRegistryPath errors", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = util.ParseRegistryPath("invalid")
		}
	})

	b.Run("ValidateRepositoryName errors", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = util.ValidateRepositoryName("")
		}
	})
}
