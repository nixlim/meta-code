package errors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		category ErrorCategory
		expected string
	}{
		{"ProtocolError", ProtocolError, "protocol"},
		{"TransportError", TransportError, "transport"},
		{"HandlerError", HandlerError, "handler"},
		{"SecurityError", SecurityError, "security"},
		{"SystemError", SystemError, "system"},
		{"UnknownCategory", ErrorCategory(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.category.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewErrorLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	
	errorLogger := NewErrorLogger(logger)
	
	require.NotNil(t, errorLogger)
	assert.Equal(t, logger, errorLogger.logger)
	assert.True(t, errorLogger.logStackTrace)
	assert.True(t, errorLogger.logContext)
	assert.True(t, errorLogger.sanitizeOutput)
}

func TestErrorLogger_LogError(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	errorLogger := NewErrorLogger(logger)
	
	err := errors.New("test error")
	errorLogger.LogError(err, "test operation failed")
	
	output := buf.String()
	assert.Contains(t, output, "test operation failed")
	assert.Contains(t, output, "test error")
}

func TestErrorLogger_LogMCPError(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	errorLogger := NewErrorLogger(logger)
	
	mcpErr := NewProtocolError("invalid request")
	mcpErr = mcpErr.WithContext(map[string]interface{}{
		"method": "test_method",
		"id":     "123",
	})
	
	errorLogger.LogMCPError(mcpErr, "MCP error occurred")
	
	output := buf.String()
	assert.Contains(t, output, "MCP error occurred")
	assert.Contains(t, output, "invalid request")
	assert.Contains(t, output, "protocol")
	assert.Contains(t, output, "-32600")
}

func TestAddMCPErrorFields(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	errorLogger := NewErrorLogger(logger)
	
	mcpErr := NewHandlerError("tool failed")
	mcpErr = mcpErr.WithContext(map[string]interface{}{
		"tool_name": "test_tool",
		"params":    map[string]interface{}{"arg1": "value1"},
	})
	
	errorLogger.LogMCPError(mcpErr, "Tool execution failed")
	
	output := buf.String()
	assert.Contains(t, output, "handler")
	assert.Contains(t, output, "-32800")
	assert.Contains(t, output, "tool_name")
}

func TestIsSensitiveKey(t *testing.T) {
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
			result := isSensitiveKey(tt.key)
			assert.Equal(t, tt.sensitive, result)
		})
	}
}

func TestSanitizeError(t *testing.T) {
	// Create an error with sensitive context
	mcpErr := NewSecurityError("authentication failed")
	mcpErr = mcpErr.WithContext(map[string]interface{}{
		"password":    "secret123",
		"token":       "abc123",
		"user_name":   "john_doe",
		"method":      "authenticate",
		"private_key": "rsa_key_data",
	})
	
	sanitized := SanitizeError(mcpErr)
	
	// Check that sensitive fields are removed
	contextStr := fmt.Sprintf("%v", sanitized.GetContext())
	assert.NotContains(t, contextStr, "secret123")
	assert.NotContains(t, contextStr, "abc123")
	assert.NotContains(t, contextStr, "rsa_key_data")
	
	// Check that non-sensitive fields are preserved
	assert.Contains(t, contextStr, "john_doe")
	assert.Contains(t, contextStr, "authenticate")
}

func TestLogWithRecovery(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	errorLogger := NewErrorLogger(logger)
	
	// Test normal execution
	result := errorLogger.LogWithRecovery(func() interface{} {
		return "success"
	})
	
	assert.Equal(t, "success", result)
	
	// Test panic recovery
	result = errorLogger.LogWithRecovery(func() interface{} {
		panic("test panic")
	})
	
	assert.Nil(t, result)
	output := buf.String()
	assert.Contains(t, output, "panic")
	assert.Contains(t, output, "test panic")
}

func TestErrorLogger_LogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	errorLogger := NewErrorLogger(logger)
	
	tests := []struct {
		name   string
		logFn  func(string, ...interface{})
		level  string
		format string
		args   []interface{}
	}{
		{"Debug", errorLogger.Debug, "DEBUG", "debug message: %s", []interface{}{"test"}},
		{"Info", errorLogger.Info, "INFO", "info message: %s", []interface{}{"test"}},
		{"Warn", errorLogger.Warn, "WARN", "warn message: %s", []interface{}{"test"}},
		{"Error", errorLogger.Error, "ERROR", "error message: %s", []interface{}{"test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFn(tt.format, tt.args...)
			
			output := buf.String()
			assert.Contains(t, output, tt.level)
			assert.Contains(t, output, "test")
		})
	}
}

func TestErrorLogger_Fatal(t *testing.T) {
	// Note: We can't easily test Fatal since it calls os.Exit
	// This test just ensures the method exists and can be called
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	errorLogger := NewErrorLogger(logger)
	
	// We can't actually call Fatal in tests, but we can verify the method exists
	assert.NotNil(t, errorLogger.Fatal)
}

func TestSetAndGetDefaultLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	errorLogger := NewErrorLogger(logger)
	
	// Set default logger
	SetDefaultLogger(errorLogger)
	
	// Get default logger
	retrieved := GetDefaultLogger()
	assert.Equal(t, errorLogger, retrieved)
}

func TestGlobalLogFunctions(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	errorLogger := NewErrorLogger(logger)
	SetDefaultLogger(errorLogger)
	
	// Test global LogError
	err := errors.New("global test error")
	LogError(err, "global error test")
	
	output := buf.String()
	assert.Contains(t, output, "global error test")
	assert.Contains(t, output, "global test error")
	
	// Test global LogMCPError
	buf.Reset()
	mcpErr := NewTransportError("connection failed")
	LogMCPError(mcpErr, "global MCP error test")
	
	output = buf.String()
	assert.Contains(t, output, "global MCP error test")
	assert.Contains(t, output, "connection failed")
	assert.Contains(t, output, "transport")
}

func TestErrorLogger_WithConfiguration(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	
	// Test with different configurations
	errorLogger := NewErrorLogger(logger)
	errorLogger.logStackTrace = false
	errorLogger.logContext = false
	errorLogger.sanitizeOutput = false
	
	mcpErr := NewProtocolError("test error")
	mcpErr = mcpErr.WithContext(map[string]interface{}{
		"password": "secret",
		"method":   "test",
	})
	
	errorLogger.LogMCPError(mcpErr, "configured test")
	
	output := buf.String()
	assert.Contains(t, output, "configured test")
	assert.Contains(t, output, "test error")
}

func TestErrorLogger_ContextSanitization(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	errorLogger := NewErrorLogger(logger)
	errorLogger.sanitizeOutput = true
	
	mcpErr := NewSecurityError("auth failed")
	mcpErr = mcpErr.WithContext(map[string]interface{}{
		"password":  "secret123",
		"user_name": "john",
		"token":     "abc123",
	})
	
	errorLogger.LogMCPError(mcpErr, "sanitization test")
	
	output := buf.String()
	assert.Contains(t, output, "sanitization test")
	assert.Contains(t, output, "john") // non-sensitive should remain
	assert.NotContains(t, output, "secret123") // sensitive should be removed
	assert.NotContains(t, output, "abc123")    // sensitive should be removed
}

func TestSanitizeError_WithNilContext(t *testing.T) {
	mcpErr := NewProtocolError("test error")
	// Don't add any context
	
	sanitized := SanitizeError(mcpErr)
	assert.NotNil(t, sanitized)
	assert.Equal(t, "test error", sanitized.Error())
}

func TestSanitizeError_WithNonMCPError(t *testing.T) {
	regularErr := errors.New("regular error")
	
	sanitized := SanitizeError(regularErr)
	assert.Equal(t, regularErr, sanitized) // Should return the same error
}

func TestLogWithRecovery_ReturnsCorrectValue(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	errorLogger := NewErrorLogger(logger)
	
	// Test with different return types
	stringResult := errorLogger.LogWithRecovery(func() interface{} {
		return "test string"
	})
	assert.Equal(t, "test string", stringResult)
	
	intResult := errorLogger.LogWithRecovery(func() interface{} {
		return 42
	})
	assert.Equal(t, 42, intResult)
	
	nilResult := errorLogger.LogWithRecovery(func() interface{} {
		return nil
	})
	assert.Nil(t, nilResult)
}
