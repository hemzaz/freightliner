package log

import (
	"context"
	"sync"
)

var (
	globalLogger Logger
	globalMutex  sync.RWMutex
)

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger Logger) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	globalMutex.RLock()
	defer globalMutex.RUnlock()

	if globalLogger == nil {
		return NewBasicLogger(InfoLevel)
	}

	return globalLogger
}

// Debug logs a debug message using the global logger
func Debug(message string) {
	GetGlobalLogger().Debug(message)
}

// Info logs an info message using the global logger
func Info(message string) {
	GetGlobalLogger().Info(message)
}

// Warn logs a warning message using the global logger
func Warn(message string) {
	GetGlobalLogger().Warn(message)
}

// Error logs an error message using the global logger
func Error(message string, err error) {
	GetGlobalLogger().Error(message, err)
}

// Fatal logs a fatal message using the global logger
func Fatal(message string, err error) {
	GetGlobalLogger().Fatal(message, err)
}

// Panic logs a panic message using the global logger
func Panic(message string, err error) {
	GetGlobalLogger().Panic(message, err)
}

// WithField adds a field to the global logger
func WithField(key string, value interface{}) Logger {
	return GetGlobalLogger().WithField(key, value)
}

// WithFields adds multiple fields to the global logger
func WithFields(fields map[string]interface{}) Logger {
	return GetGlobalLogger().WithFields(fields)
}

// WithError adds an error to the global logger
func WithError(err error) Logger {
	return GetGlobalLogger().WithError(err)
}

// WithContext adds context to the global logger
func WithContext(ctx context.Context) Logger {
	return GetGlobalLogger().WithContext(ctx)
}

func init() {
	// Initialize with a basic logger
	globalLogger = NewBasicLogger(InfoLevel)
}
