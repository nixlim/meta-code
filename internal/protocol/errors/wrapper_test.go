package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")

	wrappedErr := WrapError(originalErr, ErrorCodeMCPProtocol, "wrapped message")

	if wrappedErr == nil {
		t.Fatal("Expected wrapped error to not be nil")
	}

	if wrappedErr.Code != ErrorCodeMCPProtocol {
		t.Errorf("Expected code %d, got %d", ErrorCodeMCPProtocol, wrappedErr.Code)
	}

	if wrappedErr.Message != "wrapped message" {
		t.Errorf("Expected message 'wrapped message', got %q", wrappedErr.Message)
	}

	if wrappedErr.Cause != originalErr {
		t.Errorf("Expected cause to be original error")
	}

	if wrappedErr.Category != "protocol" {
		t.Errorf("Expected category 'protocol', got %q", wrappedErr.Category)
	}
}

func TestWrapError_NilError(t *testing.T) {
	wrappedErr := WrapError(nil, ErrorCodeMCPProtocol, "message")

	if wrappedErr != nil {
		t.Errorf("Expected nil when wrapping nil error")
	}
}

func TestWrapErrorf(t *testing.T) {
	originalErr := errors.New("original error")

	wrappedErr := WrapErrorf(originalErr, ErrorCodeMCPProtocol, "wrapped %s with %d", "error", 42)

	if wrappedErr == nil {
		t.Fatal("Expected wrapped error to not be nil")
	}

	expectedMessage := "wrapped error with 42"
	if wrappedErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, wrappedErr.Message)
	}
}

func TestWrapWithContext(t *testing.T) {
	originalErr := errors.New("original error")
	context := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	wrappedErr := WrapWithContext(originalErr, ErrorCodeMCPProtocol, "wrapped message", context)

	if wrappedErr == nil {
		t.Fatal("Expected wrapped error to not be nil")
	}

	// Check context is preserved
	if value, exists := wrappedErr.GetContext("key1"); !exists || value != "value1" {
		t.Errorf("Expected context key1 to be preserved")
	}

	if value, exists := wrappedErr.GetContext("key2"); !exists || value != 42 {
		t.Errorf("Expected context key2 to be preserved")
	}
}

func TestChainError(t *testing.T) {
	primaryErr := errors.New("primary error")
	secondaryErr := errors.New("secondary error")

	chainedErr := ChainError(primaryErr, secondaryErr, ErrorCodeMCPProtocol, "chained error")

	if chainedErr == nil {
		t.Fatal("Expected chained error to not be nil")
	}

	if chainedErr.Cause != primaryErr {
		t.Errorf("Expected cause to be primary error")
	}

	// Check that secondary error is in context
	if value, exists := chainedErr.GetContext("secondary_error"); !exists || value != secondaryErr.Error() {
		t.Errorf("Expected secondary error to be in context")
	}
}

func TestChainError_NilErrors(t *testing.T) {
	chainedErr := ChainError(nil, nil, ErrorCodeMCPProtocol, "message")

	if chainedErr != nil {
		t.Errorf("Expected nil when chaining nil errors")
	}
}

func TestUnwrapAll(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := fmt.Errorf("error 2: %w", err1)
	err3 := fmt.Errorf("error 3: %w", err2)

	allErrors := UnwrapAll(err3)

	expectedCount := 3
	if len(allErrors) != expectedCount {
		t.Errorf("Expected %d errors, got %d", expectedCount, len(allErrors))
	}

	if allErrors[0] != err3 {
		t.Errorf("Expected first error to be err3")
	}

	if allErrors[2] != err1 {
		t.Errorf("Expected last error to be err1")
	}
}

func TestFindMCPError(t *testing.T) {
	originalErr := errors.New("original error")
	mcpErr := WrapError(originalErr, ErrorCodeMCPProtocol, "mcp error")
	wrappedErr := fmt.Errorf("wrapped: %w", mcpErr)

	foundErr := FindMCPError(wrappedErr)

	if foundErr == nil {
		t.Fatal("Expected to find MCP error in chain")
	}

	if foundErr.Code != ErrorCodeMCPProtocol {
		t.Errorf("Expected found error to have correct code")
	}
}

func TestFindMCPError_NotFound(t *testing.T) {
	regularErr := errors.New("regular error")

	foundErr := FindMCPError(regularErr)

	if foundErr != nil {
		t.Errorf("Expected not to find MCP error in regular error")
	}
}

func TestFindErrorCode(t *testing.T) {
	originalErr := errors.New("original error")
	mcpErr := WrapError(originalErr, ErrorCodeMCPProtocol, "mcp error")
	wrappedErr := fmt.Errorf("wrapped: %w", mcpErr)

	found := FindErrorCode(wrappedErr, ErrorCodeMCPProtocol)
	if !found {
		t.Errorf("Expected to find error code in chain")
	}

	notFound := FindErrorCode(wrappedErr, ErrorCodeMCPTransport)
	if notFound {
		t.Errorf("Expected not to find different error code in chain")
	}
}

func TestIsTemporary(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected bool
	}{
		{"timeout error", ErrorCodeMCPTransportTimeout, true},
		{"connection lost", ErrorCodeMCPConnectionLost, true},
		{"rate limit", ErrorCodeMCPRateLimit, true},
		{"protocol error", ErrorCodeMCPProtocol, false},
		{"tool not found", ErrorCodeMCPToolNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewMCPError(tt.code, "test error", nil)
			result := IsTemporary(err)

			if result != tt.expected {
				t.Errorf("Expected IsTemporary to be %v for code %d", tt.expected, tt.code)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected bool
	}{
		{"timeout error", ErrorCodeMCPTransportTimeout, true},
		{"connection failed", ErrorCodeMCPConnectionFailed, true},
		{"service unavailable", ErrorCodeMCPServiceUnavail, true},
		{"unauthorized", ErrorCodeMCPUnauthorized, false},
		{"tool not found", ErrorCodeMCPToolNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewMCPError(tt.code, "test error", nil)
			result := IsRetryable(err)

			if result != tt.expected {
				t.Errorf("Expected IsRetryable to be %v for code %d", tt.expected, tt.code)
			}
		})
	}
}

func TestIsFatal(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected bool
	}{
		{"version mismatch", ErrorCodeMCPVersionMismatch, true},
		{"unauthorized", ErrorCodeMCPUnauthorized, true},
		{"tool not found", ErrorCodeMCPToolNotFound, true},
		{"timeout error", ErrorCodeMCPTransportTimeout, false},
		{"rate limit", ErrorCodeMCPRateLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewMCPError(tt.code, "test error", nil)
			result := IsFatal(err)

			if result != tt.expected {
				t.Errorf("Expected IsFatal to be %v for code %d", tt.expected, tt.code)
			}
		})
	}
}

func TestAggregateError(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	aggErr := NewAggregateError([]error{err1, err2, err3}, ErrorCodeMCPHandler, "multiple errors")

	if aggErr == nil {
		t.Fatal("Expected aggregate error to not be nil")
	}

	if len(aggErr.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(aggErr.Errors))
	}

	if aggErr.Code != ErrorCodeMCPHandler {
		t.Errorf("Expected code %d, got %d", ErrorCodeMCPHandler, aggErr.Code)
	}

	expectedMessage := "multiple errors (3 errors)"
	if aggErr.Error() != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, aggErr.Error())
	}
}

func TestAggregateError_EmptyErrors(t *testing.T) {
	aggErr := NewAggregateError([]error{}, ErrorCodeMCPHandler, "no errors")

	if aggErr != nil {
		t.Errorf("Expected nil aggregate error for empty error list")
	}
}

func TestAggregateError_NilErrors(t *testing.T) {
	aggErr := NewAggregateError([]error{nil, nil}, ErrorCodeMCPHandler, "nil errors")

	if aggErr != nil {
		t.Errorf("Expected nil aggregate error for nil error list")
	}
}

func TestAggregateError_ToMCPError(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	aggErr := NewAggregateError([]error{err1, err2}, ErrorCodeMCPHandler, "multiple errors")
	mcpErr := aggErr.ToMCPError()

	if mcpErr == nil {
		t.Fatal("Expected MCP error to not be nil")
	}

	if mcpErr.Cause != err1 {
		t.Errorf("Expected cause to be first error")
	}

	// Check that error count is in context
	if value, exists := mcpErr.GetContext("error_count"); !exists || value != 2 {
		t.Errorf("Expected error count to be in context")
	}

	// Check that additional errors are in context
	if _, exists := mcpErr.GetContext("additional_errors"); !exists {
		t.Errorf("Expected additional errors to be in context")
	}
}
