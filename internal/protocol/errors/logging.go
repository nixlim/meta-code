package errors

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/meta-mcp/meta-mcp-server/internal/logging"
)

// LogLevel represents the severity level for error logging
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ErrorLogger provides structured logging for MCP errors
type ErrorLogger struct {
	logger    *logging.Logger
	debugMode bool
	sanitize  bool
}

// NewErrorLogger creates a new error logger
func NewErrorLogger(debugMode bool, sanitize bool) *ErrorLogger {
	// Create a logger configuration based on debug mode
	cfg := logging.Config{
		Level:     logging.LogLevelInfo,
		DebugMode: debugMode,
		Sanitize:  sanitize,
		Pretty:    false, // Use JSON for error logging
	}

	if debugMode {
		cfg.Level = logging.LogLevelDebug
	}

	logger := logging.New(cfg)

	return &ErrorLogger{
		logger:    logger,
		debugMode: debugMode,
		sanitize:  sanitize,
	}
}

// LogError logs an MCP error with structured fields
func (el *ErrorLogger) LogError(ctx context.Context, err error, level LogLevel, message string) {
	if err == nil {
		return
	}

	// Convert LogLevel to logging.LogLevel
	logLevel := logging.LogLevel(level)

	// Build a logger with error context
	logger := el.logger.WithContext(ctx)

	// Add MCP error fields if available
	if mcpErr := FindMCPError(err); mcpErr != nil {
		logger = el.addMCPErrorFields(logger, mcpErr)
	}

	// Add caller information if in debug mode
	if el.debugMode {
		if pc, file, line, ok := runtime.Caller(1); ok {
			logger = logger.
				WithField("caller_file", file).
				WithField("caller_line", line)
			if fn := runtime.FuncForPC(pc); fn != nil {
				logger = logger.WithField("caller_func", fn.Name())
			}
		}
	}

	// Log using the internal LogError method which handles all log levels
	logger.LogError(ctx, err, logLevel, message)
}

// LogMCPError logs an MCP error with full context
func (el *ErrorLogger) LogMCPError(ctx context.Context, mcpErr *MCPError, level LogLevel, message string) {
	if mcpErr == nil {
		return
	}

	// Use the MCPError as the base error
	el.LogError(ctx, mcpErr, level, message)
}

// addMCPErrorFields adds MCP-specific fields to the logger
func (el *ErrorLogger) addMCPErrorFields(logger *logging.Logger, mcpErr *MCPError) *logging.Logger {
	logger = logger.
		WithField("error_code", mcpErr.Code).
		WithField("error_category", mcpErr.Category).
		WithField("error_message", mcpErr.Message)

	// Add context if available
	if len(mcpErr.Context) > 0 {
		contextMap := make(map[string]interface{})
		for k, v := range mcpErr.Context {
			if el.sanitize && el.isSensitiveKey(k) {
				contextMap[k] = "[REDACTED]"
			} else {
				contextMap[k] = v
			}
		}
		logger = logger.WithField("context", contextMap)
	}

	// Add debug info if in debug mode and available
	if el.debugMode && len(mcpErr.DebugInfo) > 0 {
		logger = logger.WithField("debug_info", mcpErr.DebugInfo)
	}

	// Add cause chain if available
	if mcpErr.Cause != nil {
		logger = logger.WithField("cause", mcpErr.Cause.Error())

		// Add full error chain in debug mode
		if el.debugMode {
			chain := UnwrapAll(mcpErr.Cause)
			if len(chain) > 1 {
				chainStrs := make([]string, len(chain))
				for i, chainErr := range chain {
					chainStrs[i] = chainErr.Error()
				}
				logger = logger.WithField("error_chain", chainStrs)
			}
		}
	}

	return logger
}

// isSensitiveKey checks if a context key contains sensitive information
func (el *ErrorLogger) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"password", "token", "secret", "key", "auth", "credential",
		"session", "cookie", "bearer", "api_key", "private",
	}

	keyLower := strings.ToLower(key)
	for _, sensitive := range sensitiveKeys {
		if strings.Contains(keyLower, sensitive) {
			return true
		}
	}

	return false
}

// SanitizeError removes sensitive information from an error for production logging
func (el *ErrorLogger) SanitizeError(err error) error {
	if err == nil {
		return nil
	}

	mcpErr := FindMCPError(err)
	if mcpErr == nil {
		// For non-MCP errors, just return a generic message
		return fmt.Errorf("internal error occurred")
	}

	// Create a sanitized copy
	sanitized := &MCPError{
		Code:      mcpErr.Code,
		Message:   mcpErr.Message,
		Data:      nil, // Remove data to prevent leaks
		Category:  mcpErr.Category,
		Context:   make(map[string]interface{}),
		Sanitized: true,
	}

	// Copy only non-sensitive context
	for k, v := range mcpErr.Context {
		if !el.isSensitiveKey(k) {
			sanitized.Context[k] = v
		}
	}

	// Don't include debug info or cause in sanitized version
	return sanitized
}

// LogWithRecovery logs an error and recovers from panics
func (el *ErrorLogger) LogWithRecovery(ctx context.Context, operation string) {
	if r := recover(); r != nil {
		var err error
		if e, ok := r.(error); ok {
			err = e
		} else {
			err = fmt.Errorf("panic: %v", r)
		}

		// Create a system error for the panic
		panicErr := NewSystemError(fmt.Sprintf("Panic during %s", operation), nil)
		panicErr.Cause = err
		panicErr.WithContext("operation", operation)

		if el.debugMode {
			// Add stack trace in debug mode
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			panicErr.WithDebugInfo("stack_trace", string(buf[:n]))
		}

		el.LogMCPError(ctx, panicErr, LogLevelError, "Recovered from panic")
	}
}

// Convenience methods for different log levels

// Debug logs an error at debug level
func (el *ErrorLogger) Debug(ctx context.Context, err error, message string) {
	el.LogError(ctx, err, LogLevelDebug, message)
}

// Info logs an error at info level
func (el *ErrorLogger) Info(ctx context.Context, err error, message string) {
	el.LogError(ctx, err, LogLevelInfo, message)
}

// Warn logs an error at warn level
func (el *ErrorLogger) Warn(ctx context.Context, err error, message string) {
	el.LogError(ctx, err, LogLevelWarn, message)
}

// Error logs an error at error level
func (el *ErrorLogger) Error(ctx context.Context, err error, message string) {
	el.LogError(ctx, err, LogLevelError, message)
}

// Fatal logs an error at fatal level
func (el *ErrorLogger) Fatal(ctx context.Context, err error, message string) {
	el.LogError(ctx, err, LogLevelFatal, message)
}

// Global logger instance (can be configured)
var defaultLogger = NewErrorLogger(false, true)

// SetDefaultLogger sets the global default logger
func SetDefaultLogger(logger *ErrorLogger) {
	defaultLogger = logger
}

// GetDefaultLogger returns the global default logger
func GetDefaultLogger() *ErrorLogger {
	return defaultLogger
}

// Convenience functions using the default logger

// LogError logs an error using the default logger
func LogError(ctx context.Context, err error, level LogLevel, message string) {
	defaultLogger.LogError(ctx, err, level, message)
}

// LogMCPError logs an MCP error using the default logger
func LogMCPError(ctx context.Context, mcpErr *MCPError, level LogLevel, message string) {
	defaultLogger.LogMCPError(ctx, mcpErr, level, message)
}
