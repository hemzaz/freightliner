// Package errors provides standardized error handling utilities for the freightliner application.
// It wraps around the standard errors package and fmt.Errorf to provide consistent error handling patterns.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// Common error types that can be used across the application
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInternal      = errors.New("internal error")
	ErrUnavailable   = errors.New("service unavailable")
	ErrTimeout       = errors.New("operation timed out")
	ErrNotSupported  = errors.New("not supported")
	ErrCanceled      = errors.New("operation canceled")
)

// New creates a new error with the given message.
// This is a direct wrapper around errors.New.
func New(message string) error {
	return errors.New(message)
}

// Wrap wraps an error with additional context using fmt.Errorf and the %w verb.
// If err is nil, Wrap returns nil.
func Wrap(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	// If no args are provided, just add the message
	if len(args) == 0 {
		return fmt.Errorf("%s: %w", format, err)
	}

	// Format the message and then append the error
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// Wrapf wraps an error with a formatted message.
// This is the same as Wrap but makes the formatting more explicit in the function name.
func Wrapf(err error, format string, args ...interface{}) error {
	return Wrap(err, format, args...)
}

// Is reports whether any error in err's tree matches target.
// This is a direct wrapper around errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's tree that matches target, and if one is found, sets
// target to that error value and returns true. Otherwise, it returns false.
// This is a direct wrapper around errors.As.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Unwrap returns the result of calling the Unwrap method on err, if err implements Unwrap.
// Otherwise, Unwrap returns nil.
// This is a direct wrapper around errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Common helper function for creating formatted errors with a base error
func formatError(baseError error, format string, args ...interface{}) error {
	if len(args) == 0 {
		return fmt.Errorf("%s: %w", format, baseError)
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), baseError)
}

// NotFoundf returns an error indicating that the requested resource was not found.
func NotFoundf(format string, args ...interface{}) error {
	return formatError(ErrNotFound, format, args...)
}

// AlreadyExistsf returns an error indicating that the resource already exists.
func AlreadyExistsf(format string, args ...interface{}) error {
	return formatError(ErrAlreadyExists, format, args...)
}

// InvalidInputf returns an error indicating that the input was invalid.
func InvalidInputf(format string, args ...interface{}) error {
	return formatError(ErrInvalidInput, format, args...)
}

// Unauthorizedf returns an error indicating that the user is not authorized.
func Unauthorizedf(format string, args ...interface{}) error {
	return formatError(ErrUnauthorized, format, args...)
}

// Forbiddenf returns an error indicating that the action is forbidden.
func Forbiddenf(format string, args ...interface{}) error {
	return formatError(ErrForbidden, format, args...)
}

// Internalf returns an error indicating an internal error.
func Internalf(format string, args ...interface{}) error {
	return formatError(ErrInternal, format, args...)
}

// Unavailablef returns an error indicating that a service is unavailable.
func Unavailablef(format string, args ...interface{}) error {
	return formatError(ErrUnavailable, format, args...)
}

// Timeoutf returns an error indicating that an operation timed out.
func Timeoutf(format string, args ...interface{}) error {
	return formatError(ErrTimeout, format, args...)
}

// NotSupportedf returns an error indicating that the functionality is not supported.
func NotSupportedf(format string, args ...interface{}) error {
	return formatError(ErrNotSupported, format, args...)
}

// Canceledf returns an error indicating that an operation was canceled.
func Canceledf(format string, args ...interface{}) error {
	return formatError(ErrCanceled, format, args...)
}

// NotImplementedf returns an error indicating that the functionality is not implemented.
// This is an alias for NotSupportedf for backward compatibility.
func NotImplementedf(format string, args ...interface{}) error {
	return NotSupportedf(format, args...)
}

// Newf creates a new error with a formatted message.
// This is equivalent to fmt.Errorf without any wrapped errors.
func Newf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// Multiple combines multiple errors into a single error.
// If no errors are provided, returns nil.
// If only one error is provided, returns that error.
// If multiple errors are provided, returns an error that contains all of them.
func Multiple(errs ...error) error {
	// Filter out nil errors
	validErrors := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			validErrors = append(validErrors, err)
		}
	}

	// Return based on number of valid errors
	switch len(validErrors) {
	case 0:
		return nil
	case 1:
		return validErrors[0]
	default:
		return &multiError{errors: validErrors}
	}
}

// multiError is an error that wraps multiple errors
type multiError struct {
	errors []error
}

// Error returns a string representation of all errors
func (me *multiError) Error() string {
	if len(me.errors) == 0 {
		return ""
	}

	if len(me.errors) == 1 {
		return me.errors[0].Error()
	}

	messages := make([]string, len(me.errors))
	for i, err := range me.errors {
		messages[i] = err.Error()
	}

	return strings.Join(messages, "; ")
}

// Unwrap returns the first error for error unwrapping
func (me *multiError) Unwrap() error {
	if len(me.errors) == 0 {
		return nil
	}
	return me.errors[0]
}

// Errors returns all wrapped errors
func (me *multiError) Errors() []error {
	return me.errors
}
