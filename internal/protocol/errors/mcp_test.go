package errors

import (
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestMCPError_Creation(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		message  string
		data     interface{}
		expected string
	}{
		{
			name:     "protocol error",
			code:     ErrorCodeMCPProtocol,
			message:  "test protocol error",
			data:     nil,
			expected: "protocol",
		},
		{
			name:     "transport error",
			code:     ErrorCodeMCPTransport,
			message:  "test transport error",
			data:     nil,
			expected: "transport",
		},
		{
			name:     "handler error",
			code:     ErrorCodeMCPHandler,
			message:  "test handler error",
			data:     nil,
			expected: "handler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewMCPError(tt.code, tt.message, tt.data)

			if err.Code != tt.code {
				t.Errorf("Expected code %d, got %d", tt.code, err.Code)
			}

			if err.Message != tt.message {
				t.Errorf("Expected message %q, got %q", tt.message, err.Message)
			}

			if err.Category != tt.expected {
				t.Errorf("Expected category %q, got %q", tt.expected, err.Category)
			}
		})
	}
}

func TestMCPError_ErrorInterface(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewMCPError(ErrorCodeMCPProtocol, "test error", nil)
	err.Cause = cause

	expected := "MCP protocol error (-32000): test error - caused by: underlying error"
	if err.Error() != expected {
		t.Errorf("Expected error string %q, got %q", expected, err.Error())
	}
}

func TestMCPError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewMCPError(ErrorCodeMCPProtocol, "test error", nil)
	err.Cause = cause

	if err.Unwrap() != cause {
		t.Errorf("Expected unwrap to return cause error")
	}
}

func TestMCPError_Is(t *testing.T) {
	err1 := NewMCPError(ErrorCodeMCPProtocol, "test error", nil)
	err2 := NewMCPError(ErrorCodeMCPProtocol, "different message", nil)
	err3 := NewMCPError(ErrorCodeMCPTransport, "test error", nil)

	if !errors.Is(err1, err2) {
		t.Errorf("Expected errors with same code to be equal")
	}

	if errors.Is(err1, err3) {
		t.Errorf("Expected errors with different codes to not be equal")
	}
}

func TestMCPError_As(t *testing.T) {
	err := NewMCPError(ErrorCodeMCPProtocol, "test error", nil)

	var mcpErr *MCPError
	if !errors.As(err, &mcpErr) {
		t.Errorf("Expected errors.As to work with MCPError")
	}

	if mcpErr.Code != ErrorCodeMCPProtocol {
		t.Errorf("Expected extracted error to have correct code")
	}
}

func TestMCPError_Context(t *testing.T) {
	err := NewMCPError(ErrorCodeMCPProtocol, "test error", nil)

	// Test adding context
	err.WithContext("key1", "value1")
	err.WithContext("key2", 42)

	// Test retrieving context
	if value, exists := err.GetContext("key1"); !exists || value != "value1" {
		t.Errorf("Expected context key1 to have value 'value1'")
	}

	if value, exists := err.GetContext("key2"); !exists || value != 42 {
		t.Errorf("Expected context key2 to have value 42")
	}

	// Test string context
	if value, exists := err.GetContextString("key1"); !exists || value != "value1" {
		t.Errorf("Expected context string key1 to have value 'value1'")
	}

	// Test non-existent key
	if _, exists := err.GetContext("nonexistent"); exists {
		t.Errorf("Expected non-existent key to not exist")
	}
}

func TestMCPError_DebugInfo(t *testing.T) {
	err := NewMCPError(ErrorCodeMCPProtocol, "test error", nil)

	err.WithDebugInfo("debug_key", "debug_value")

	if value, exists := err.DebugInfo["debug_key"]; !exists || value != "debug_value" {
		t.Errorf("Expected debug info to be set correctly")
	}
}

func TestMCPError_Sanitize(t *testing.T) {
	err := NewMCPError(ErrorCodeMCPProtocol, "test error", nil)
	err.WithContext("safe_key", "safe_value")
	err.WithContext("password", "secret123")
	err.WithContext("api_key", "key123")
	err.WithDebugInfo("debug_info", "sensitive_debug")
	err.Cause = errors.New("underlying cause")

	sanitized := err.Sanitize()

	// Check that safe context is preserved
	if value, exists := sanitized.GetContext("safe_key"); !exists || value != "safe_value" {
		t.Errorf("Expected safe context to be preserved")
	}

	// Check that sensitive context is removed
	if _, exists := sanitized.GetContext("password"); exists {
		t.Errorf("Expected sensitive context to be removed")
	}

	if _, exists := sanitized.GetContext("api_key"); exists {
		t.Errorf("Expected sensitive context to be removed")
	}

	// Check that debug info is removed
	if len(sanitized.DebugInfo) > 0 {
		t.Errorf("Expected debug info to be removed")
	}

	// Check that cause is removed
	if sanitized.Cause != nil {
		t.Errorf("Expected cause to be removed")
	}

	// Check that sanitized flag is set
	if !sanitized.Sanitized {
		t.Errorf("Expected sanitized flag to be set")
	}
}

func TestMCPError_Clone(t *testing.T) {
	err := NewMCPError(ErrorCodeMCPProtocol, "test error", nil)
	err.WithContext("key", "value")
	err.WithDebugInfo("debug", "info")
	err.Cause = errors.New("cause")

	clone := err.Clone()

	// Check that clone has same values
	if clone.Code != err.Code {
		t.Errorf("Expected clone to have same code")
	}

	if clone.Message != err.Message {
		t.Errorf("Expected clone to have same message")
	}

	if clone.Category != err.Category {
		t.Errorf("Expected clone to have same category")
	}

	// Check that context is copied
	if value, exists := clone.GetContext("key"); !exists || value != "value" {
		t.Errorf("Expected clone to have same context")
	}

	// Check that modifying clone doesn't affect original
	clone.WithContext("new_key", "new_value")
	if _, exists := err.GetContext("new_key"); exists {
		t.Errorf("Expected original to not be affected by clone modification")
	}
}

func TestMCPError_ToMCPError(t *testing.T) {
	err := NewMCPError(ErrorCodeMCPProtocol, "test error", "test data")
	requestId := mcp.NewRequestId("test-id")

	mcpErr := err.ToMCPError(requestId)

	if mcpErr.Error.Code != err.Code {
		t.Errorf("Expected MCP error to have same code")
	}

	if mcpErr.Error.Message != err.Message {
		t.Errorf("Expected MCP error to have same message")
	}
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{ErrorCodeMCPProtocol, "protocol"},
		{ErrorCodeMCPTransport, "transport"},
		{ErrorCodeMCPHandler, "handler"},
		{ErrorCodeMCPSecurity, "security"},
		{ErrorCodeMCPSystem, "system"},
		{-99999, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			category := GetCategory(tt.code)
			if category != tt.expected {
				t.Errorf("Expected category %q for code %d, got %q", tt.expected, tt.code, category)
			}
		})
	}
}

func TestIsMCPError(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{ErrorCodeMCPProtocol, true},
		{ErrorCodeMCPSystem, true},
		{-32100, false}, // Outside MCP range
		{-31999, false}, // Outside MCP range
		{-32700, false}, // JSON-RPC error
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := IsMCPError(tt.code)
			if result != tt.expected {
				t.Errorf("Expected IsMCPError(%d) to be %v, got %v", tt.code, tt.expected, result)
			}
		})
	}
}

func TestGetMCPErrorMessage(t *testing.T) {
	// Test known error code
	message := GetMCPErrorMessage(ErrorCodeMCPProtocol)
	if message != "MCP protocol error" {
		t.Errorf("Expected known error message, got %q", message)
	}

	// Test unknown error code
	message = GetMCPErrorMessage(-99999)
	if message != "Unknown MCP error" {
		t.Errorf("Expected unknown error message, got %q", message)
	}
}

func TestMCPError_ToJSONRPCError(t *testing.T) {
	mcpErr := NewProtocolError("invalid request", nil)
	mcpErr = mcpErr.WithContext("method", "test_method")
	mcpErr = mcpErr.WithContext("id", "123")

	jsonrpcErr := mcpErr.ToJSONRPCError()

	assert.Equal(t, ErrorCodeMCPProtocol, jsonrpcErr.Code)
	assert.Equal(t, "invalid request", jsonrpcErr.Message)
	// Data might be nil, so just check that the conversion worked
	assert.NotNil(t, jsonrpcErr)
}

func TestMCPError_WithCause(t *testing.T) {
	originalErr := NewProtocolError("original error", nil)
	causeErr := NewTransportError("cause error", nil)

	result := originalErr.WithCause(causeErr)

	assert.Equal(t, originalErr, result) // Should return the same instance
	assert.Equal(t, causeErr, result.Cause)
}

func TestMCPError_HasContext(t *testing.T) {
	mcpErr := NewProtocolError("test error", nil)

	// Initially should not have context
	assert.False(t, mcpErr.HasContext("key"))

	// After adding context should have context
	mcpErr = mcpErr.WithContext("key", "value")
	assert.True(t, mcpErr.HasContext("key"))
}

func TestMCPError_RemoveContext(t *testing.T) {
	mcpErr := NewProtocolError("test error", nil)
	mcpErr = mcpErr.WithContext("key1", "value1")
	mcpErr = mcpErr.WithContext("key2", "value2")

	mcpErr.RemoveContext("key1")

	_, exists1 := mcpErr.GetContext("key1")
	_, exists2 := mcpErr.GetContext("key2")
	assert.False(t, exists1)
	assert.True(t, exists2)
}

func TestMCPError_ClearContext(t *testing.T) {
	mcpErr := NewProtocolError("test error", nil)
	mcpErr = mcpErr.WithContext("key1", "value1")
	mcpErr = mcpErr.WithContext("key2", "value2")

	mcpErr.ClearContext()

	assert.False(t, mcpErr.HasContext("key1"))
	assert.False(t, mcpErr.HasContext("key2"))
}

func TestMCPError_ClearDebugInfo(t *testing.T) {
	mcpErr := NewProtocolError("test error", nil)
	mcpErr = mcpErr.WithDebugInfo("stack", "trace info")
	mcpErr = mcpErr.WithDebugInfo("line", 123)

	mcpErr.ClearDebugInfo()

	assert.Empty(t, mcpErr.DebugInfo)
}
