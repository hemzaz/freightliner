package log

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestNewStructuredLoggers(t *testing.T) {
	l1 := NewStructuredLogger(InfoLevel)
	if l1 == nil {
		t.Error("NewStructuredLogger() returned nil")
	}

	var buf bytes.Buffer
	l2 := NewStructuredLoggerWithWriter(DebugLevel, &buf)
	if l2 == nil {
		t.Error("NewStructuredLoggerWithWriter() returned nil")
	}
}

func TestStructuredLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLoggerWithWriter(DebugLevel, &buf)

	// Test Debug
	logger.Debug("debug message")
	var entry1 map[string]interface{}
	json.Unmarshal(buf.Bytes(), &entry1)
	if entry1["level"] != "debug" {
		t.Error("Expected debug level")
	}

	// Test Info
	buf.Reset()
	logger.Info("info message")
	var entry2 map[string]interface{}
	json.Unmarshal(buf.Bytes(), &entry2)
	if entry2["level"] != "info" {
		t.Error("Expected info level")
	}

	// Test Warn
	buf.Reset()
	logger.Warn("warn message")
	var entry3 map[string]interface{}
	json.Unmarshal(buf.Bytes(), &entry3)
	if entry3["level"] != "warn" {
		t.Error("Expected warn level")
	}

	// Test Error
	buf.Reset()
	logger.Error("error message", errors.New("test"))
	var entry4 map[string]interface{}
	json.Unmarshal(buf.Bytes(), &entry4)
	if entry4["level"] != "error" || entry4["error"] != "test" {
		t.Error("Expected error level with error field")
	}
}

func TestStructuredLoggerWithField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLoggerWithWriter(InfoLevel, &buf)

	logger.WithField("key", "value").Info("test")
	var entry map[string]interface{}
	json.Unmarshal(buf.Bytes(), &entry)

	fields, ok := entry["fields"].(map[string]interface{})
	if !ok || fields["key"] != "value" {
		t.Error("Expected field in JSON output")
	}
}

func TestStructuredLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLoggerWithWriter(InfoLevel, &buf)

	logger.WithFields(map[string]interface{}{
		"a": 1,
		"b": "test",
	}).Info("test")

	var entry map[string]interface{}
	json.Unmarshal(buf.Bytes(), &entry)

	fields, ok := entry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields in JSON")
	}
	if fields["a"] != float64(1) || fields["b"] != "test" {
		t.Error("Expected both fields")
	}
}

func TestStructuredLoggerWithError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLoggerWithWriter(InfoLevel, &buf)

	testErr := errors.New("test error")
	logger.WithError(testErr).Info("test")

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Error("Expected error in output")
	}

	// Test nil error
	buf.Reset()
	logger.WithError(nil).Info("test2")
	if buf.Len() == 0 {
		t.Error("Expected output even with nil error")
	}
}

func TestStructuredLoggerWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLoggerWithWriter(InfoLevel, &buf)

	ctx := context.Background()
	logger.WithContext(ctx).Info("test")

	if buf.Len() == 0 {
		t.Error("Expected output")
	}
}

func TestStructuredLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLoggerWithWriter(WarnLevel, &buf)

	// Debug should be filtered
	logger.Debug("debug")
	if buf.Len() > 0 {
		t.Error("Expected DEBUG to be filtered")
	}

	// Info should be filtered
	buf.Reset()
	logger.Info("info")
	if buf.Len() > 0 {
		t.Error("Expected INFO to be filtered")
	}

	// Warn should log
	buf.Reset()
	logger.Warn("warn")
	if buf.Len() == 0 {
		t.Error("Expected WARN to log")
	}

	// Error should log
	buf.Reset()
	logger.Error("error", nil)
	if buf.Len() == 0 {
		t.Error("Expected ERROR to log")
	}
}
