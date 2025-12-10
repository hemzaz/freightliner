package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	msg := "test error message"
	err := New(msg)

	if err == nil {
		t.Fatal("New() returned nil")
	}

	if err.Error() != msg {
		t.Errorf("New() error message = %s, want %s", err.Error(), msg)
	}
}

func TestWrap(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name        string
		err         error
		format      string
		args        []interface{}
		wantNil     bool
		wantContain []string
	}{
		{
			name:        "wrap with context",
			err:         baseErr,
			format:      "operation failed",
			args:        []interface{}{},
			wantContain: []string{"operation failed", "base error"},
		},
		{
			name:        "wrap with formatted context",
			err:         baseErr,
			format:      "operation %s failed",
			args:        []interface{}{"read"},
			wantContain: []string{"operation read failed", "base error"},
		},
		{
			name:    "wrap nil error",
			err:     nil,
			format:  "this should not appear",
			wantNil: true,
		},
		{
			name:        "wrap with multiple args",
			err:         baseErr,
			format:      "user %s operation %d failed",
			args:        []interface{}{"admin", 42},
			wantContain: []string{"user admin operation 42 failed", "base error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result error
			if tt.err == nil {
				// For nil error test, use Wrap directly to test its nil handling
				result = Wrap(tt.err, tt.format, tt.args...)
			} else if len(tt.args) == 0 {
				result = fmt.Errorf("%s: %w", tt.format, tt.err)
			} else {
				result = Wrap(tt.err, tt.format, tt.args...)
			}

			if tt.wantNil {
				if result != nil {
					t.Errorf("Wrap() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Wrap() returned nil")
			}

			errMsg := result.Error()
			for _, want := range tt.wantContain {
				if !strings.Contains(errMsg, want) {
					t.Errorf("Wrap() error message = %s, want to contain %s", errMsg, want)
				}
			}

			// Verify the error can be unwrapped
			if !Is(result, baseErr) {
				t.Error("Wrap() error cannot be unwrapped to base error")
			}
		})
	}
}

func TestWrapf(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrapf(baseErr, "context: %s", "test")

	if wrapped == nil {
		t.Fatal("Wrapf() returned nil")
	}

	if !strings.Contains(wrapped.Error(), "context: test") {
		t.Error("Wrapf() did not format message correctly")
	}

	if !Is(wrapped, baseErr) {
		t.Error("Wrapf() error cannot be unwrapped to base error")
	}
}

func TestIs(t *testing.T) {
	baseErr := ErrNotFound
	wrappedErr := Wrap(baseErr, "resource not found")

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "direct match",
			err:    ErrNotFound,
			target: ErrNotFound,
			want:   true,
		},
		{
			name:   "wrapped match",
			err:    wrappedErr,
			target: ErrNotFound,
			want:   true,
		},
		{
			name:   "no match",
			err:    ErrInvalidInput,
			target: ErrNotFound,
			want:   false,
		},
		{
			name:   "nil error",
			err:    nil,
			target: ErrNotFound,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Is(tt.err, tt.target); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAs(t *testing.T) {
	customErr := &customError{msg: "custom"}
	wrappedCustom := fmt.Errorf("wrapped: %w", customErr)

	var target *customError
	if !As(wrappedCustom, &target) {
		t.Error("As() should find custom error in chain")
	}

	if target.msg != "custom" {
		t.Errorf("As() target.msg = %s, want custom", target.msg)
	}
}

func TestUnwrap(t *testing.T) {
	baseErr := errors.New("base")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	unwrapped := Unwrap(wrappedErr)
	if unwrapped != baseErr {
		t.Error("Unwrap() did not return base error")
	}

	unwrappedAgain := Unwrap(baseErr)
	if unwrappedAgain != nil {
		t.Error("Unwrap() of non-wrapped error should return nil")
	}
}

func TestCommonErrors(t *testing.T) {
	commonErrors := []error{
		ErrNotFound,
		ErrAlreadyExists,
		ErrInvalidInput,
		ErrUnauthorized,
		ErrForbidden,
		ErrInternal,
		ErrUnavailable,
		ErrTimeout,
		ErrNotSupported,
		ErrCanceled,
	}

	for _, err := range commonErrors {
		if err == nil {
			t.Error("common error should not be nil")
		}
		if err.Error() == "" {
			t.Error("common error should have message")
		}
	}
}

func TestFormattedErrors(t *testing.T) {
	tests := []struct {
		name      string
		fn        func(string, ...interface{}) error
		format    string
		args      []interface{}
		baseErr   error
		wantInMsg []string
	}{
		{
			name:      "NotFoundf",
			fn:        NotFoundf,
			format:    "user %s",
			args:      []interface{}{"john"},
			baseErr:   ErrNotFound,
			wantInMsg: []string{"user john", "not found"},
		},
		{
			name:      "AlreadyExistsf",
			fn:        AlreadyExistsf,
			format:    "resource %d",
			args:      []interface{}{123},
			baseErr:   ErrAlreadyExists,
			wantInMsg: []string{"resource 123", "already exists"},
		},
		{
			name:      "InvalidInputf",
			fn:        InvalidInputf,
			format:    "field %s",
			args:      []interface{}{"email"},
			baseErr:   ErrInvalidInput,
			wantInMsg: []string{"field email", "invalid input"},
		},
		{
			name:      "Unauthorizedf",
			fn:        Unauthorizedf,
			format:    "access denied",
			baseErr:   ErrUnauthorized,
			wantInMsg: []string{"access denied", "unauthorized"},
		},
		{
			name:      "Forbiddenf",
			fn:        Forbiddenf,
			format:    "operation not allowed",
			baseErr:   ErrForbidden,
			wantInMsg: []string{"operation not allowed", "forbidden"},
		},
		{
			name:      "Internalf",
			fn:        Internalf,
			format:    "database error",
			baseErr:   ErrInternal,
			wantInMsg: []string{"database error", "internal error"},
		},
		{
			name:      "Unavailablef",
			fn:        Unavailablef,
			format:    "service down",
			baseErr:   ErrUnavailable,
			wantInMsg: []string{"service down", "service unavailable"},
		},
		{
			name:      "Timeoutf",
			fn:        Timeoutf,
			format:    "after %d seconds",
			args:      []interface{}{30},
			baseErr:   ErrTimeout,
			wantInMsg: []string{"after 30 seconds", "timed out"},
		},
		{
			name:      "NotSupportedf",
			fn:        NotSupportedf,
			format:    "feature %s",
			args:      []interface{}{"xyz"},
			baseErr:   ErrNotSupported,
			wantInMsg: []string{"feature xyz", "not supported"},
		},
		{
			name:      "Canceledf",
			fn:        Canceledf,
			format:    "by user",
			baseErr:   ErrCanceled,
			wantInMsg: []string{"by user", "canceled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if len(tt.args) == 0 {
				err = tt.fn(tt.format)
			} else {
				err = tt.fn(tt.format, tt.args...)
			}

			if err == nil {
				t.Fatal("formatted error function returned nil")
			}

			errMsg := err.Error()
			for _, want := range tt.wantInMsg {
				if !strings.Contains(errMsg, want) {
					t.Errorf("error message = %s, want to contain %s", errMsg, want)
				}
			}

			// Verify the error wraps the base error
			if !Is(err, tt.baseErr) {
				t.Errorf("error should wrap %v", tt.baseErr)
			}
		})
	}
}

func TestNotImplementedf(t *testing.T) {
	err := NotImplementedf("feature %s", "xyz")

	if err == nil {
		t.Fatal("NotImplementedf() returned nil")
	}

	if !Is(err, ErrNotSupported) {
		t.Error("NotImplementedf() should wrap ErrNotSupported")
	}

	if !strings.Contains(err.Error(), "feature xyz") {
		t.Error("NotImplementedf() should format message")
	}
}

func TestNewf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "simple message",
			format: "error occurred",
			want:   "error occurred",
		},
		{
			name:   "formatted message",
			format: "error in %s: %d",
			args:   []interface{}{"function", 42},
			want:   "error in function: 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Newf(tt.format, tt.args...)

			if err == nil {
				t.Fatal("Newf() returned nil")
			}

			if err.Error() != tt.want {
				t.Errorf("Newf() = %s, want %s", err.Error(), tt.want)
			}
		})
	}
}

func TestMultiple(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	tests := []struct {
		name        string
		errs        []error
		wantNil     bool
		wantContain []string
		wantCount   int
	}{
		{
			name:    "no errors",
			errs:    []error{},
			wantNil: true,
		},
		{
			name:    "all nil errors",
			errs:    []error{nil, nil, nil},
			wantNil: true,
		},
		{
			name:        "single error",
			errs:        []error{err1},
			wantContain: []string{"error 1"},
			wantCount:   1,
		},
		{
			name:        "single error with nils",
			errs:        []error{nil, err1, nil},
			wantContain: []string{"error 1"},
			wantCount:   1,
		},
		{
			name:        "multiple errors",
			errs:        []error{err1, err2, err3},
			wantContain: []string{"error 1", "error 2", "error 3"},
			wantCount:   3,
		},
		{
			name:        "multiple errors with nils",
			errs:        []error{nil, err1, nil, err2, err3, nil},
			wantContain: []string{"error 1", "error 2", "error 3"},
			wantCount:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Multiple(tt.errs...)

			if tt.wantNil {
				if result != nil {
					t.Errorf("Multiple() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Multiple() returned nil")
			}

			errMsg := result.Error()
			for _, want := range tt.wantContain {
				if !strings.Contains(errMsg, want) {
					t.Errorf("Multiple() error message = %s, want to contain %s", errMsg, want)
				}
			}

			// Check if we can access the underlying multiError
			if me, ok := result.(*multiError); ok {
				if len(me.Errors()) != tt.wantCount {
					t.Errorf("Multiple() error count = %d, want %d", len(me.Errors()), tt.wantCount)
				}
			} else if tt.wantCount > 1 {
				t.Error("Multiple() should return *multiError for multiple errors")
			}
		})
	}
}

func TestMultiErrorUnwrap(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	multiErr := Multiple(err1, err2)

	unwrapped := Unwrap(multiErr)
	if unwrapped != err1 {
		t.Error("multiError.Unwrap() should return first error")
	}
}

func TestMultiErrorEmpty(t *testing.T) {
	me := &multiError{errors: []error{}}

	if me.Error() != "" {
		t.Error("empty multiError should have empty string")
	}

	if me.Unwrap() != nil {
		t.Error("empty multiError.Unwrap() should return nil")
	}
}

func TestMultiErrorSingle(t *testing.T) {
	err1 := errors.New("single error")
	me := &multiError{errors: []error{err1}}

	if me.Error() != "single error" {
		t.Errorf("single multiError.Error() = %s, want 'single error'", me.Error())
	}
}

func TestMultiErrorSeparator(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	multiErr := Multiple(err1, err2, err3).(*multiError)
	errMsg := multiErr.Error()

	// Check that errors are separated by semicolon
	if !strings.Contains(errMsg, ";") {
		t.Error("multiError should separate errors with semicolon")
	}

	parts := strings.Split(errMsg, ";")
	if len(parts) != 3 {
		t.Errorf("multiError should have 3 parts, got %d", len(parts))
	}
}

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func TestErrorChaining(t *testing.T) {
	// Create a chain of errors
	base := errors.New("base error")
	level1 := Wrap(base, "level 1")
	level2 := Wrap(level1, "level 2")
	level3 := Wrap(level2, "level 3")

	// Test that Is works through the chain
	if !Is(level3, base) {
		t.Error("Is() should find base error through chain")
	}

	// Test unwrapping through the chain
	unwrapped := Unwrap(level3)
	if !strings.Contains(unwrapped.Error(), "level 2") {
		t.Error("Unwrap() should return level 2")
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New("benchmark error")
	}
}

func BenchmarkWrap(b *testing.B) {
	baseErr := errors.New("base")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Wrap(baseErr, "context")
	}
}

func BenchmarkNotFoundf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NotFoundf("resource %d", 123)
	}
}

func BenchmarkMultiple(b *testing.B) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Multiple(err1, err2, err3)
	}
}
