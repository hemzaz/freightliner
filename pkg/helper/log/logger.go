package log

import (
	"fmt"
	"os"
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
)

// Logger provides structured logging capabilities
type Logger struct {
	level Level
}

// NewLogger creates a new logger with the specified level
func NewLogger(level Level) *Logger {
	return &Logger{level: level}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	if l.level <= DebugLevel {
		l.log("DEBUG", msg, fields)
	}
}

// Info logs an informational message
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	if l.level <= InfoLevel {
		l.log("INFO", msg, fields)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	if l.level <= WarnLevel {
		l.log("WARN", msg, fields)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	if l.level <= ErrorLevel {
		if fields == nil {
			fields = make(map[string]interface{})
		}
		if err != nil {
			fields["error"] = err.Error()
		}
		l.log("ERROR", msg, fields)
	}
}

// Fatal logs a fatal message and terminates the program
func (l *Logger) Fatal(msg string, err error, fields map[string]interface{}) {
	if l.level <= FatalLevel {
		if fields == nil {
			fields = make(map[string]interface{})
		}
		if err != nil {
			fields["error"] = err.Error()
		}
		l.log("FATAL", msg, fields)
		os.Exit(1)
	}
}

func (l *Logger) log(level string, msg string, fields map[string]interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("%s [%s] %s", timestamp, level, msg)
	if len(fields) > 0 {
		fmt.Print(" ")
		for k, v := range fields {
			fmt.Printf("%s=%v ", k, v)
		}
	}
	fmt.Println()
}
