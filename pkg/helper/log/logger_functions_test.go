package log

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestLevelStringValues(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{PanicLevel, "PANIC"},
		{Level(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %v, want %v", tt.level, got, tt.want)
		}
	}
}

func TestNewLoggerVariants(t *testing.T) {
	l1 := NewLogger()
	if l1 == nil {
		t.Error("NewLogger() returned nil")
	}

	l2 := NewLoggerWithLevel(WarnLevel)
	if l2 == nil {
		t.Error("NewLoggerWithLevel() returned nil")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  Level
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"error", ErrorLevel},
		{"fatal", FatalLevel},
		{"panic", PanicLevel},
		{"invalid", InfoLevel}, // default
		{"", InfoLevel},        // default
	}
	for _, tt := range tests {
		if got := ParseLevel(tt.input); got != tt.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestBasicLoggerWithField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewBasicLoggerWithWriter(InfoLevel, &buf)

	logger.WithField("key", "value").Info("test")
	output := buf.String()
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected field in output: %s", output)
	}
}

func TestBasicLoggerWithError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewBasicLoggerWithWriter(InfoLevel, &buf)

	testErr := errors.New("test error")
	logger.WithError(testErr).Info("test")
	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Errorf("Expected error in output: %s", output)
	}

	// Test nil error
	buf.Reset()
	logger.WithError(nil).Info("test2")
	output = buf.String()
	if !strings.Contains(output, "test2") {
		t.Error("Expected message even with nil error")
	}
}

func TestBasicLoggerWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewBasicLoggerWithWriter(InfoLevel, &buf)

	ctx := context.Background()
	logger.WithContext(ctx).Info("test")
	if !strings.Contains(buf.String(), "test") {
		t.Error("Expected message")
	}
}

func TestBasicLoggerLogMethodsWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewBasicLoggerWithWriter(DebugLevel, &buf)

	fields := map[string]interface{}{"key": "value"}

	logger.Debug("debug", fields)
	if !strings.Contains(buf.String(), "debug") || !strings.Contains(buf.String(), "key=value") {
		t.Error("Debug with fields failed")
	}

	buf.Reset()
	logger.Info("info", fields)
	if !strings.Contains(buf.String(), "info") {
		t.Error("Info with fields failed")
	}

	buf.Reset()
	logger.Warn("warn", fields)
	if !strings.Contains(buf.String(), "warn") {
		t.Error("Warn with fields failed")
	}

	buf.Reset()
	logger.Error("error", nil, fields)
	if !strings.Contains(buf.String(), "error") {
		t.Error("Error with fields failed")
	}
}
