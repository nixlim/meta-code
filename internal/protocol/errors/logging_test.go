package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"Debug", LogLevelDebug, "DEBUG"},
		{"Info", LogLevelInfo, "INFO"},
		{"Warn", LogLevelWarn, "WARN"},
		{"Error", LogLevelError, "ERROR"},
		{"Fatal", LogLevelFatal, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewErrorLogger(t *testing.T) {
	errorLogger := NewErrorLogger(true, true)

	require.NotNil(t, errorLogger)
}

func TestErrorLogger_LogError(t *testing.T) {
	errorLogger := NewErrorLogger(false, false)

	err := errors.New("test error")
	// This should not panic
	errorLogger.LogError(nil, err, LogLevelError, "test operation failed")
}

func TestErrorLogger_LogMCPError(t *testing.T) {
	errorLogger := NewErrorLogger(false, false)

	mcpErr := NewProtocolError("invalid request", nil)
	mcpErr = mcpErr.WithContext("method", "test_method")
	mcpErr = mcpErr.WithContext("id", "123")

	// This should not panic
	errorLogger.LogMCPError(nil, mcpErr, LogLevelError, "MCP error occurred")
}

func TestErrorLogger_IsSensitiveKey(t *testing.T) {
	errorLogger := NewErrorLogger(false, true)

	tests := []struct {
		key       string
		sensitive bool
	}{
		{"password", true},
		{"token", true},
		{"secret", true},
		{"api_key", true},
		{"auth", true},
		{"credential", true},
		{"private", true},
		{"normal_field", false},
		{"user_name", false},
		{"method", false},
		{"id", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := errorLogger.isSensitiveKey(tt.key)
			assert.Equal(t, tt.sensitive, result)
		})
	}
}

func TestErrorLogger_SanitizeError(t *testing.T) {
	errorLogger := NewErrorLogger(false, true)

	// Create an error with sensitive context
	mcpErr := NewSecurityError("authentication failed", nil)
	mcpErr = mcpErr.WithContext("password", "secret123")
	mcpErr = mcpErr.WithContext("token", "abc123")
	mcpErr = mcpErr.WithContext("user_name", "john_doe")
	mcpErr = mcpErr.WithContext("method", "authenticate")
	mcpErr = mcpErr.WithContext("private_key", "rsa_key_data")

	sanitized := errorLogger.SanitizeError(mcpErr)

	// Check that it returns an error (the method should not panic)
	assert.NotNil(t, sanitized)
}

func TestErrorLogger_LogWithRecovery(t *testing.T) {
	errorLogger := NewErrorLogger(false, false)

	// This should not panic
	errorLogger.LogWithRecovery(nil, "test operation")
}

func TestErrorLogger_LogLevels(t *testing.T) {
	errorLogger := NewErrorLogger(false, false)

	testErr := errors.New("test error")

	// These should not panic
	errorLogger.Debug(nil, testErr, "debug message")
	errorLogger.Info(nil, testErr, "info message")
	errorLogger.Warn(nil, testErr, "warn message")
	errorLogger.Error(nil, testErr, "error message")
}

func TestSetAndGetDefaultLogger(t *testing.T) {
	errorLogger := NewErrorLogger(false, false)

	// Set default logger
	SetDefaultLogger(errorLogger)

	// Get default logger
	retrieved := GetDefaultLogger()
	assert.Equal(t, errorLogger, retrieved)
}

func TestGlobalLogFunctions(t *testing.T) {
	errorLogger := NewErrorLogger(false, false)
	SetDefaultLogger(errorLogger)

	// Test global LogError
	err := errors.New("global test error")
	LogError(nil, err, LogLevelError, "global error test")

	// Test global LogMCPError
	mcpErr := NewTransportError("connection failed", nil)
	LogMCPError(nil, mcpErr, LogLevelError, "global MCP error test")
}

func TestErrorLogger_WithConfiguration(t *testing.T) {
	// Test with different configurations
	debugLogger := NewErrorLogger(true, true)
	assert.NotNil(t, debugLogger)

	prodLogger := NewErrorLogger(false, false)
	assert.NotNil(t, prodLogger)
}
