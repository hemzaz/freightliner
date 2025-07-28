package util

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	debugMessages []string
	warnMessages  []string
	fields        []map[string]interface{}
}

func (m *MockLogger) Debug(msg string, fields map[string]interface{}) {
	m.debugMessages = append(m.debugMessages, msg)
	m.fields = append(m.fields, fields)
}

func (m *MockLogger) Warn(msg string, fields map[string]interface{}) {
	m.warnMessages = append(m.warnMessages, msg)
	m.fields = append(m.fields, fields)
}

func TestDefaultRetryOptions(t *testing.T) {
	opts := DefaultRetryOptions()

	if opts.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries to be 5, got %d", opts.MaxRetries)
	}

	if opts.InitialWait != 1*time.Second {
		t.Errorf("Expected InitialWait to be 1s, got %v", opts.InitialWait)
	}

	if opts.MaxWait != 60*time.Second {
		t.Errorf("Expected MaxWait to be 60s, got %v", opts.MaxWait)
	}

	if opts.Factor != 2.0 {
		t.Errorf("Expected Factor to be 2.0, got %f", opts.Factor)
	}
}

func TestRetryWithContextSuccess(t *testing.T) {
	called := 0
	successOnAttempt := 1

	fn := func() error {
		called++
		if called >= successOnAttempt {
			return nil
		}
		return errors.New("temporary error")
	}

	opts := RetryOptions{
		MaxRetries:  3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Factor:      2.0,
		Retryable:   func(err error) bool { return true },
	}

	ctx := context.Background()
	err := RetryWithContext(ctx, fn, opts)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if called != successOnAttempt {
		t.Errorf("Expected function to be called %d times, got %d", successOnAttempt, called)
	}
}

func TestRetryWithContextMaxRetries(t *testing.T) {
	called := 0
	maxRetries := 3

	fn := func() error {
		called++
		return errors.New("persistent error")
	}

	opts := RetryOptions{
		MaxRetries:  maxRetries,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Factor:      2.0,
		Retryable:   func(err error) bool { return true },
	}

	ctx := context.Background()
	err := RetryWithContext(ctx, fn, opts)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Function called once initially, then maxRetries times
	expectedCalls := maxRetries + 1
	if called != expectedCalls {
		t.Errorf("Expected function to be called %d times, got %d", expectedCalls, called)
	}
}

func TestRetryWithContextCancellation(t *testing.T) {
	called := 0

	fn := func() error {
		called++
		time.Sleep(50 * time.Millisecond) // Ensure operation takes some time
		return errors.New("temporary error")
	}

	opts := RetryOptions{
		MaxRetries:  5,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Factor:      2.0,
		Retryable:   func(err error) bool { return true },
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := RetryWithContext(ctx, fn, opts)

	if err == nil {
		t.Errorf("Expected context cancellation error, got nil")
	} else if err.Error() != "retry aborted by context cancellation" && err.Error() != "context deadline exceeded" {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}

	// Should have called the function at least once
	if called < 1 {
		t.Errorf("Expected function to be called at least once, got %d", called)
	}

	// But should not have completed all retries
	if called > 3 {
		t.Errorf("Expected fewer calls due to cancellation, got %d", called)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	called := 0
	successOnAttempt := 2

	fn := func() error {
		called++
		if called >= successOnAttempt {
			return nil
		}
		return errors.New("temporary error")
	}

	ctx := context.Background()
	err := RetryWithBackoff(ctx, 2, 10*time.Millisecond, 100*time.Millisecond, fn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if called != successOnAttempt {
		t.Errorf("Expected function to be called %d times, got %d", successOnAttempt, called)
	}
}

func TestRetryWithBackoffFailure(t *testing.T) {
	called := 0
	maxRetries := 2

	fn := func() error {
		called++
		return errors.New("persistent error")
	}

	ctx := context.Background()
	err := RetryWithBackoff(ctx, maxRetries, 10*time.Millisecond, 100*time.Millisecond, fn)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Function called once initially, then maxRetries times
	expectedCalls := maxRetries + 1
	if called != expectedCalls {
		t.Errorf("Expected function to be called %d times, got %d", expectedCalls, called)
	}
}

func TestRetryWithLogger(t *testing.T) {
	called := 0
	successOnAttempt := 3

	fn := func() error {
		called++
		if called >= successOnAttempt {
			return nil
		}
		return errors.New("temporary error")
	}

	logger := &MockLogger{
		debugMessages: []string{},
		warnMessages:  []string{},
		fields:        []map[string]interface{}{},
	}

	ctx := context.Background()
	err := RetryWithBackoffAndLogger(ctx, 3, 10*time.Millisecond, 100*time.Millisecond, logger, "test operation", fn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if called != successOnAttempt {
		t.Errorf("Expected function to be called %d times, got %d", successOnAttempt, called)
	}

	// Should have debug logs for eventual success
	if len(logger.debugMessages) < 1 {
		t.Error("Expected debug messages for eventual success")
	}

	// Should have warning messages for each failed attempt
	expectedWarnMsgs := successOnAttempt - 1
	if len(logger.warnMessages) != expectedWarnMsgs {
		t.Errorf("Expected %d warning messages, got %d", expectedWarnMsgs, len(logger.warnMessages))
	}
}
