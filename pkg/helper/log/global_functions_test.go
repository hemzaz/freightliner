package log

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestGlobalLoggerSetGet(t *testing.T) {
	// Save original
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	var buf bytes.Buffer
	custom := NewBasicLoggerWithWriter(DebugLevel, &buf)
	SetGlobalLogger(custom)

	retrieved := GetGlobalLogger()
	if retrieved == nil {
		t.Error("Expected non-nil logger")
	}

	// Test it works
	Info("test")
	if !strings.Contains(buf.String(), "test") {
		t.Error("Expected message in output")
	}
}

func TestGlobalDebug(t *testing.T) {
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	var buf bytes.Buffer
	SetGlobalLogger(NewBasicLoggerWithWriter(DebugLevel, &buf))
	Debug("debug msg")
	if !strings.Contains(buf.String(), "debug msg") {
		t.Error("Expected debug message")
	}
}

func TestGlobalInfo(t *testing.T) {
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	var buf bytes.Buffer
	SetGlobalLogger(NewBasicLoggerWithWriter(InfoLevel, &buf))
	Info("info msg")
	if !strings.Contains(buf.String(), "info msg") {
		t.Error("Expected info message")
	}
}

func TestGlobalWarn(t *testing.T) {
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	var buf bytes.Buffer
	SetGlobalLogger(NewBasicLoggerWithWriter(WarnLevel, &buf))
	Warn("warn msg")
	if !strings.Contains(buf.String(), "warn msg") {
		t.Error("Expected warn message")
	}
}

func TestGlobalError(t *testing.T) {
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	var buf bytes.Buffer
	SetGlobalLogger(NewBasicLoggerWithWriter(ErrorLevel, &buf))
	Error("error msg", errors.New("test"))
	if !strings.Contains(buf.String(), "error msg") {
		t.Error("Expected error message")
	}
}

func TestGlobalWithField(t *testing.T) {
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	var buf bytes.Buffer
	SetGlobalLogger(NewBasicLoggerWithWriter(InfoLevel, &buf))
	WithField("key", "val").Info("msg")
	output := buf.String()
	if !strings.Contains(output, "msg") || !strings.Contains(output, "key=val") {
		t.Errorf("Expected message and field, got: %s", output)
	}
}

func TestGlobalWithFields(t *testing.T) {
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	var buf bytes.Buffer
	SetGlobalLogger(NewBasicLoggerWithWriter(InfoLevel, &buf))
	WithFields(map[string]interface{}{"a": 1, "b": 2}).Info("msg")
	output := buf.String()
	if !strings.Contains(output, "a=1") || !strings.Contains(output, "b=2") {
		t.Errorf("Expected fields, got: %s", output)
	}
}

func TestGlobalWithError(t *testing.T) {
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	var buf bytes.Buffer
	SetGlobalLogger(NewBasicLoggerWithWriter(InfoLevel, &buf))
	WithError(errors.New("err")).Info("msg")
	output := buf.String()
	if !strings.Contains(output, "err") {
		t.Errorf("Expected error in output, got: %s", output)
	}
}

func TestGlobalWithContext(t *testing.T) {
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	var buf bytes.Buffer
	SetGlobalLogger(NewBasicLoggerWithWriter(InfoLevel, &buf))
	ctx := context.Background()
	WithContext(ctx).Info("msg")
	if !strings.Contains(buf.String(), "msg") {
		t.Error("Expected message")
	}
}

func TestGlobalNilLogger(t *testing.T) {
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	SetGlobalLogger(nil)
	logger := GetGlobalLogger()
	if logger == nil {
		t.Error("Expected default logger when nil")
	}
}

func TestGlobalConcurrent(t *testing.T) {
	orig := GetGlobalLogger()
	defer SetGlobalLogger(orig)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			SetGlobalLogger(NewBasicLogger(InfoLevel))
		}()
		go func() {
			defer wg.Done()
			_ = GetGlobalLogger()
		}()
	}
	wg.Wait()
}
