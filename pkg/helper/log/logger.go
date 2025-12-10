package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Level represents a logging level
type Level int

const (
	// DebugLevel is for detailed debugging information
	DebugLevel Level = iota
	// InfoLevel is for general operational information
	InfoLevel
	// WarnLevel is for warning messages
	WarnLevel
	// ErrorLevel is for error messages
	ErrorLevel
	// FatalLevel is for fatal errors that should terminate the program
	FatalLevel
	// PanicLevel is for panic messages
	PanicLevel
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case PanicLevel:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a logger instance
type Logger interface {
	Debug(message string, fields ...map[string]interface{})
	Info(message string, fields ...map[string]interface{})
	Warn(message string, fields ...map[string]interface{})
	Error(message string, err error, fields ...map[string]interface{})
	Fatal(message string, err error, fields ...map[string]interface{})
	Panic(message string, err error, fields ...map[string]interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
	WithContext(ctx context.Context) Logger
}

// BasicLogger provides basic logging capabilities
type BasicLogger struct {
	level  Level
	writer io.Writer
	fields map[string]interface{}
}

// NewBasicLogger creates a new logger with the specified level
func NewBasicLogger(level Level) Logger {
	return &BasicLogger{
		level:  level,
		writer: os.Stdout,
		fields: make(map[string]interface{}),
	}
}

// NewBasicLoggerWithWriter creates a logger with custom writer
func NewBasicLoggerWithWriter(level Level, writer io.Writer) Logger {
	return &BasicLogger{
		level:  level,
		writer: writer,
		fields: make(map[string]interface{}),
	}
}

// WithField adds a field to the logger
func (l *BasicLogger) WithField(key string, value interface{}) Logger {
	newLogger := &BasicLogger{
		level:  l.level,
		writer: l.writer,
		fields: make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value

	return newLogger
}

// WithFields adds multiple fields to the logger
func (l *BasicLogger) WithFields(fields map[string]interface{}) Logger {
	newLogger := &BasicLogger{
		level:  l.level,
		writer: l.writer,
		fields: make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithError adds an error to the logger
func (l *BasicLogger) WithError(err error) Logger {
	if err == nil {
		return l
	}

	newLogger := &BasicLogger{
		level:  l.level,
		writer: l.writer,
		fields: make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add error field
	newLogger.fields["error"] = err.Error()

	return newLogger
}

// WithContext adds context information to the logger
func (l *BasicLogger) WithContext(ctx context.Context) Logger {
	return l // BasicLogger doesn't support context - use StructuredLogger for context support
}

// Debug logs a debug message
func (l *BasicLogger) Debug(message string, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.logWithFields(DebugLevel, message, nil, fieldMap)
}

// Info logs an info message
func (l *BasicLogger) Info(message string, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.logWithFields(InfoLevel, message, nil, fieldMap)
}

// Warn logs a warning message
func (l *BasicLogger) Warn(message string, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.logWithFields(WarnLevel, message, nil, fieldMap)
}

// Error logs an error message
func (l *BasicLogger) Error(message string, err error, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.logWithFields(ErrorLevel, message, err, fieldMap)
}

// Fatal logs a fatal message and exits
func (l *BasicLogger) Fatal(message string, err error, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.logWithFields(FatalLevel, message, err, fieldMap)
	os.Exit(1)
}

// Panic logs a panic message and panics
func (l *BasicLogger) Panic(message string, err error, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.logWithFields(PanicLevel, message, err, fieldMap)
	panic(message)
}

// logWithFields is the internal logging method with field support
func (l *BasicLogger) logWithFields(level Level, message string, err error, fields map[string]interface{}) {
	// Check if we should log at this level
	if level < l.level {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	levelStr := strings.ToUpper(level.String())

	// Build the log line
	logLine := fmt.Sprintf("[%s] %s: %s", timestamp, levelStr, message)

	// Add error if present
	if err != nil {
		logLine += fmt.Sprintf(" error=\"%v\"", err)
	}

	// Add fields if present
	for k, v := range fields {
		logLine += fmt.Sprintf(" %s=%v", k, v)
	}

	// Add logger's own fields
	for k, v := range l.fields {
		logLine += fmt.Sprintf(" %s=%v", k, v)
	}

	logLine += "\n"

	_, _ = l.writer.Write([]byte(logLine))
}

// log is the internal logging method (kept for backward compatibility)
func (l *BasicLogger) log(level Level, message string, err error) {
	l.logWithFields(level, message, err, nil)
}

// NewLogger creates a new logger with INFO level by default
func NewLogger() Logger {
	return NewBasicLogger(InfoLevel)
}

// NewLoggerWithLevel creates a new logger with specified level
func NewLoggerWithLevel(level Level) Logger {
	return NewBasicLogger(level)
}

// ParseLevel parses a string level and returns the corresponding Level
func ParseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	case "panic":
		return PanicLevel
	default:
		return InfoLevel
	}
}
