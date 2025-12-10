package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

// StructuredLogger provides structured logging with JSON output
type StructuredLogger struct {
	level  Level
	writer io.Writer
	fields map[string]interface{}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    *CallerInfo            `json:"caller,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
}

// CallerInfo contains information about the calling code
type CallerInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(level Level) *StructuredLogger {
	return &StructuredLogger{
		level:  level,
		writer: os.Stdout,
		fields: make(map[string]interface{}),
	}
}

// NewStructuredLoggerWithWriter creates a structured logger with custom writer
func NewStructuredLoggerWithWriter(level Level, writer io.Writer) *StructuredLogger {
	return &StructuredLogger{
		level:  level,
		writer: writer,
		fields: make(map[string]interface{}),
	}
}

// WithField adds a field to the logger context
func (l *StructuredLogger) WithField(key string, value interface{}) *StructuredLogger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &StructuredLogger{
		level:  l.level,
		writer: l.writer,
		fields: newFields,
	}
}

// WithFields adds multiple fields to the logger context
func (l *StructuredLogger) WithFields(fields map[string]interface{}) *StructuredLogger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &StructuredLogger{
		level:  l.level,
		writer: l.writer,
		fields: newFields,
	}
}

// WithError adds an error to the logger context
func (l *StructuredLogger) WithError(err error) *StructuredLogger {
	if err == nil {
		return l
	}
	return l.WithField("error", err.Error())
}

// WithContext extracts tracing information from context
func (l *StructuredLogger) WithContext(ctx context.Context) *StructuredLogger {
	// Extract trace and span IDs from context if available
	// This would integrate with OpenTelemetry or similar tracing systems
	newLogger := l

	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		newLogger = newLogger.WithField("trace_id", traceID)
	}

	if spanID := getSpanIDFromContext(ctx); spanID != "" {
		newLogger = newLogger.WithField("span_id", spanID)
	}

	return newLogger
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(message string) {
	l.log(DebugLevel, message, nil, false)
}

// Info logs an info message
func (l *StructuredLogger) Info(message string) {
	l.log(InfoLevel, message, nil, false)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(message string) {
	l.log(WarnLevel, message, nil, false)
}

// Error logs an error message
func (l *StructuredLogger) Error(message string, err error) {
	l.log(ErrorLevel, message, err, false)
}

// Fatal logs a fatal message and exits
func (l *StructuredLogger) Fatal(message string, err error) {
	l.log(FatalLevel, message, err, true)
	os.Exit(1)
}

// Panic logs a panic message and panics
func (l *StructuredLogger) Panic(message string, err error) {
	l.log(PanicLevel, message, err, true)
	panic(message)
}

// log is the internal logging method
func (l *StructuredLogger) log(level Level, message string, err error, includeStack bool) {
	// Check if we should log at this level
	if level < l.level {
		return
	}

	// Create log entry
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     strings.ToLower(level.String()),
		Message:   message,
		Fields:    make(map[string]interface{}),
	}

	// Add fields
	for k, v := range l.fields {
		entry.Fields[k] = v
	}

	// Add error if present
	if err != nil {
		entry.Error = err.Error()
	}

	// Add caller information
	if caller := getCaller(3); caller != nil {
		entry.Caller = caller
	}

	// Add stack trace for errors and above
	if includeStack || level >= ErrorLevel {
		entry.Stack = getStackTrace()
	}

	// Get tracing information from fields
	if traceID, ok := entry.Fields["trace_id"].(string); ok {
		entry.TraceID = traceID
		delete(entry.Fields, "trace_id")
	}

	if spanID, ok := entry.Fields["span_id"].(string); ok {
		entry.SpanID = spanID
		delete(entry.Fields, "span_id")
	}

	// Remove fields if empty
	if len(entry.Fields) == 0 {
		entry.Fields = nil
	}

	// Marshal to JSON
	data, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		// Fallback to simple text output if JSON marshaling fails
		fallback := fmt.Sprintf("[%s] %s %s",
			entry.Timestamp,
			strings.ToUpper(entry.Level),
			entry.Message)
		if entry.Error != "" {
			fallback += fmt.Sprintf(" error=%s", entry.Error)
		}
		fallback += "\n"
		_, _ = l.writer.Write([]byte(fallback))
		return
	}

	// Write JSON log entry
	_, _ = l.writer.Write(data)
	_, _ = l.writer.Write([]byte("\n"))
}

// getCaller returns information about the calling function
func getCaller(skip int) *CallerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	// Get function name
	fn := runtime.FuncForPC(pc)
	var funcName string
	if fn != nil {
		funcName = fn.Name()
		// Remove package path, keep only the function name
		if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}
	}

	// Shorten file path
	if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
		file = file[lastSlash+1:]
	}

	return &CallerInfo{
		File:     file,
		Line:     line,
		Function: funcName,
	}
}

// getStackTrace returns a stack trace as a string
func getStackTrace() string {
	buf := make([]byte, 1024*8)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// getTraceIDFromContext extracts trace ID from context
// This is a placeholder - in a real implementation this would integrate
// with your tracing system (OpenTelemetry, Jaeger, etc.)
func getTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Example: extract from OpenTelemetry context
	// span := trace.SpanFromContext(ctx)
	// if span.SpanContext().IsValid() {
	//     return span.SpanContext().TraceID().String()
	// }

	return ""
}

// getSpanIDFromContext extracts span ID from context
// This is a placeholder - in a real implementation this would integrate
// with your tracing system (OpenTelemetry, Jaeger, etc.)
func getSpanIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Example: extract from OpenTelemetry context
	// span := trace.SpanFromContext(ctx)
	// if span.SpanContext().IsValid() {
	//     return span.SpanContext().SpanID().String()
	// }

	return ""
}
