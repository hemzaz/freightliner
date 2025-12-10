package log

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
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
			logFunc: func(logger Logger, msg string) {
				logger.Error(msg, errors.New("test error"))
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
			// Use a buffer to capture output instead of pipe
			var output bytes.Buffer
			logger := NewBasicLoggerWithWriter(tt.loggerLevel, &output)
			testMsg := "test message"

			tt.logFunc(logger, testMsg)

			logOutput := output.String()
			if tt.expectedLog {
				if !strings.Contains(logOutput, testMsg) {
					t.Errorf("Expected log to contain message '%s', but got: %s", testMsg, logOutput)
				}
			} else {
				if strings.Contains(logOutput, testMsg) {
					t.Errorf("Expected no log message, but got: %s", logOutput)
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
			var output bytes.Buffer
			logger := NewBasicLoggerWithWriter(InfoLevel, &output)

			logger.WithFields(tt.fields).Info(tt.msg)

			logOutput := output.String()

			// Check if expected message is present
			if !strings.Contains(logOutput, tt.msg) {
				t.Errorf("Expected message '%s' not found in log output: %s", tt.msg, logOutput)
			}

			// Check for all expected fields keys in output
			for _, key := range tt.expectedKeys {
				if !strings.Contains(logOutput, key+"=") {
					t.Errorf("Expected field key '%s' not found in log output: %s", key, logOutput)
				}
			}

			// Check field values if any
			if tt.fields != nil {
				for k, v := range tt.fields {
					valueStr := fmt.Sprintf("%v", v)
					if !strings.Contains(logOutput, k+"="+valueStr) {
						t.Errorf("Expected field '%s=%v' not found in log output: %s", k, v, logOutput)
					}
				}
			}
		})
	}
}

func TestErrorLogging(t *testing.T) {
	var output bytes.Buffer
	logger := NewBasicLoggerWithWriter(InfoLevel, &output)
	testErr := errors.New("test error")

	logger.WithFields(map[string]interface{}{
		"key": "value",
	}).Error("error message", testErr)

	logOutput := output.String()

	// Check error field
	if !strings.Contains(logOutput, "error=\""+testErr.Error()+"\"") {
		t.Errorf("Expected 'error=\"%s\"' in log output, but got: %s", testErr.Error(), logOutput)
	}
}

// Testing Fatal logging format without triggering os.Exit
func TestFatalLogging(t *testing.T) {
	// We test the format by using a higher log level to prevent the actual Fatal behavior
	var output bytes.Buffer
	logger := NewBasicLoggerWithWriter(DebugLevel, &output)
	testErr := errors.New("fatal error")

	// Test fatal logging format by directly calling the internal log method
	// Since we can't safely test Fatal (it calls os.Exit), we verify the format works
	basicLogger := logger.(*BasicLogger)
	basicLogger.logWithFields(FatalLevel, "fatal message", testErr, nil)

	logOutput := output.String()
	if !strings.Contains(logOutput, "FATAL") || !strings.Contains(logOutput, "fatal message") {
		t.Errorf("Expected FATAL level log with message, but got: %s", logOutput)
	}
}
