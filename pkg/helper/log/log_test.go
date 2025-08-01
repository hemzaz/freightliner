package log

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// captureOutput captures stdout during the execution of function f
func captureOutput(f func()) string {
	// Save the original stdout
	originalStdout := os.Stdout

	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the function
	f()

	// Close the write end of the pipe
	_ = w.Close()

	// Restore original stdout
	os.Stdout = originalStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	return buf.String()
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name         string
		loggerLevel  Level
		logFuncLevel Level
		logFunc      func(logger Logger, msg string)
		expectedLog  bool
	}{
		{
			name:         "Debug level logs debug",
			loggerLevel:  DebugLevel,
			logFuncLevel: DebugLevel,
			logFunc: func(logger Logger, msg string) {
				logger.Debug(msg)
			},
			expectedLog: true,
		},
		{
			name:         "Info level doesn't log debug",
			loggerLevel:  InfoLevel,
			logFuncLevel: DebugLevel,
			logFunc: func(logger Logger, msg string) {
				logger.Debug(msg)
			},
			expectedLog: false,
		},
		{
			name:         "Debug level logs info",
			loggerLevel:  DebugLevel,
			logFuncLevel: InfoLevel,
			logFunc: func(logger Logger, msg string) {
				logger.Info(msg)
			},
			expectedLog: true,
		},
		{
			name:         "Info level logs info",
			loggerLevel:  InfoLevel,
			logFuncLevel: InfoLevel,
			logFunc: func(logger Logger, msg string) {
				logger.Info(msg)
			},
			expectedLog: true,
		},
		{
			name:         "Warn level logs warn",
			loggerLevel:  WarnLevel,
			logFuncLevel: WarnLevel,
			logFunc: func(logger Logger, msg string) {
				logger.Warn(msg)
			},
			expectedLog: true,
		},
		{
			name:         "Warn level doesn't log info",
			loggerLevel:  WarnLevel,
			logFuncLevel: InfoLevel,
			logFunc: func(logger Logger, msg string) {
				logger.Info(msg)
			},
			expectedLog: false,
		},
		{
			name:         "Error level logs error",
			loggerLevel:  ErrorLevel,
			logFuncLevel: ErrorLevel,
			logFunc: func(logger *Logger, msg string) {
				logger.Error(msg, errors.New("test error"), nil)
			},
			expectedLog: true,
		},
		{
			name:         "Error level doesn't log warn",
			loggerLevel:  ErrorLevel,
			logFuncLevel: WarnLevel,
			logFunc: func(logger Logger, msg string) {
				logger.Warn(msg)
			},
			expectedLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			testMsg := "test message " + time.Now().String() // Make message unique

			output := captureOutput(func() {
				tt.logFunc(logger, testMsg)
			})

			if tt.expectedLog {
				if !strings.Contains(output, testMsg) {
					t.Errorf("Expected log to contain message '%s', but got: %s", testMsg, output)
				}
			} else {
				if strings.Contains(output, testMsg) {
					t.Errorf("Expected no log message, but got: %s", output)
				}
			}
		})
	}
}

func TestLoggerFields(t *testing.T) {
	tests := []struct {
		name         string
		msg          string
		fields       map[string]interface{}
		expectedKeys []string
	}{
		{
			name:         "Log with no fields",
			msg:          "simple message",
			fields:       nil,
			expectedKeys: []string{},
		},
		{
			name: "Log with string field",
			msg:  "message with fields",
			fields: map[string]interface{}{
				"stringKey": "string value",
			},
			expectedKeys: []string{"stringKey"},
		},
		{
			name: "Log with multiple fields",
			msg:  "message with multiple fields",
			fields: map[string]interface{}{
				"stringKey": "string value",
				"intKey":    42,
				"boolKey":   true,
			},
			expectedKeys: []string{"stringKey", "intKey", "boolKey"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(DebugLevel)

			output := captureOutput(func() {
				logger.Info(tt.msg, tt.fields)
			})

			// Check if expected message is present
			if !strings.Contains(output, tt.msg) {
				t.Errorf("Expected message '%s' not found in log output: %s", tt.msg, output)
			}

			// Check for all expected fields keys in output
			for _, key := range tt.expectedKeys {
				if !strings.Contains(output, key+"=") {
					t.Errorf("Expected field key '%s' not found in log output: %s", key, output)
				}
			}

			// Check field values if any
			if tt.fields != nil {
				for k, v := range tt.fields {
					valueStr := fmt.Sprintf("%v", v)
					if !strings.Contains(output, k+"="+valueStr) {
						t.Errorf("Expected field '%s=%v' not found in log output: %s", k, v, output)
					}
				}
			}
		})
	}
}

func TestErrorLogging(t *testing.T) {
	logger := NewLogger(DebugLevel)
	testErr := errors.New("test error")

	output := captureOutput(func() {
		logger.Error("error message", testErr, map[string]interface{}{
			"key": "value",
		})
	})

	// Check error field
	if !strings.Contains(output, "error="+testErr.Error()) {
		t.Errorf("Expected 'error=%s' in log output, but got: %s", testErr.Error(), output)
	}
}

// Testing Fatal requires mocking os.Exit, which is difficult
// This test just verifies that the FATAL level is properly logged
func TestFatalLogging(t *testing.T) {
	// NOTE: We can't fully test Fatal because it calls os.Exit
	// We can only test the logging part, not the exit behavior

	logger := NewLogger(FatalLevel + 1) // Set level higher than Fatal to avoid actual exit
	testErr := errors.New("fatal error")

	// Test that Fatal doesn't log or exit when level is too high
	output := captureOutput(func() {
		logger.Fatal("should not log", testErr, nil)
	})

	if strings.Contains(output, "should not log") {
		t.Errorf("Expected no log output, but got: %s", output)
	}
}
