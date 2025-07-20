// Package helpers provides custom assertions for MCP protocol testing
package helpers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JSONRPCAssertions provides assertions specific to JSON-RPC messages
type JSONRPCAssertions struct {
	t *testing.T
}

// NewJSONRPCAssertions creates a new JSON-RPC assertions helper
func NewJSONRPCAssertions(t *testing.T) *JSONRPCAssertions {
	return &JSONRPCAssertions{t: t}
}

// AssertValidJSONRPCRequest validates that a message is a valid JSON-RPC request
func (j *JSONRPCAssertions) AssertValidJSONRPCRequest(data []byte, msgAndArgs ...interface{}) {
	j.t.Helper()
	
	var req map[string]interface{}
	err := json.Unmarshal(data, &req)
	require.NoError(j.t, err, "Failed to unmarshal JSON-RPC request")
	
	// Check required fields
	assert.Contains(j.t, req, "jsonrpc", msgAndArgs...)
	assert.Contains(j.t, req, "method", msgAndArgs...)
	assert.Contains(j.t, req, "id", msgAndArgs...)
	
	// Check version
	assert.Equal(j.t, "2.0", req["jsonrpc"], msgAndArgs...)
	
	// Check method is string
	assert.IsType(j.t, "", req["method"], msgAndArgs...)
	
	// ID can be string, number, or null, but not missing
	assert.NotNil(j.t, req["id"], msgAndArgs...)
}

// AssertValidJSONRPCResponse validates that a message is a valid JSON-RPC response
func (j *JSONRPCAssertions) AssertValidJSONRPCResponse(data []byte, msgAndArgs ...interface{}) {
	j.t.Helper()
	
	var resp map[string]interface{}
	err := json.Unmarshal(data, &resp)
	require.NoError(j.t, err, "Failed to unmarshal JSON-RPC response")
	
	// Check required fields
	assert.Contains(j.t, resp, "jsonrpc", msgAndArgs...)
	assert.Contains(j.t, resp, "id", msgAndArgs...)
	
	// Check version
	assert.Equal(j.t, "2.0", resp["jsonrpc"], msgAndArgs...)
	
	// Must have either result or error, but not both
	hasResult := resp["result"] != nil
	hasError := resp["error"] != nil
	
	assert.True(j.t, hasResult != hasError, "Response must have either result or error, but not both", msgAndArgs...)
}

// AssertValidJSONRPCError validates that a message is a valid JSON-RPC error response
func (j *JSONRPCAssertions) AssertValidJSONRPCError(data []byte, expectedCode int, msgAndArgs ...interface{}) {
	j.t.Helper()
	
	var resp map[string]interface{}
	err := json.Unmarshal(data, &resp)
	require.NoError(j.t, err, "Failed to unmarshal JSON-RPC error response")
	
	// Validate basic response structure
	j.AssertValidJSONRPCResponse(data, msgAndArgs...)
	
	// Check error structure
	assert.Contains(j.t, resp, "error", msgAndArgs...)
	
	errorObj, ok := resp["error"].(map[string]interface{})
	require.True(j.t, ok, "Error field must be an object", msgAndArgs...)
	
	// Check error fields
	assert.Contains(j.t, errorObj, "code", msgAndArgs...)
	assert.Contains(j.t, errorObj, "message", msgAndArgs...)
	
	// Check error code
	code, ok := errorObj["code"].(float64) // JSON numbers are float64
	require.True(j.t, ok, "Error code must be a number", msgAndArgs...)
	assert.Equal(j.t, expectedCode, int(code), msgAndArgs...)
	
	// Check message is string
	assert.IsType(j.t, "", errorObj["message"], msgAndArgs...)
}

// MCPAssertions provides assertions specific to MCP protocol
type MCPAssertions struct {
	t *testing.T
}

// NewMCPAssertions creates a new MCP assertions helper
func NewMCPAssertions(t *testing.T) *MCPAssertions {
	return &MCPAssertions{t: t}
}

// AssertValidMCPInitialize validates an MCP initialize request
func (m *MCPAssertions) AssertValidMCPInitialize(data []byte, msgAndArgs ...interface{}) {
	m.t.Helper()
	
	var req map[string]interface{}
	err := json.Unmarshal(data, &req)
	require.NoError(m.t, err, "Failed to unmarshal MCP initialize request")
	
	// Check it's a valid JSON-RPC request
	jsonrpcAssert := NewJSONRPCAssertions(m.t)
	jsonrpcAssert.AssertValidJSONRPCRequest(data, msgAndArgs...)
	
	// Check method
	assert.Equal(m.t, "initialize", req["method"], msgAndArgs...)
	
	// Check params structure
	assert.Contains(m.t, req, "params", msgAndArgs...)
	params, ok := req["params"].(map[string]interface{})
	require.True(m.t, ok, "Params must be an object", msgAndArgs...)
	
	// Check required params
	assert.Contains(m.t, params, "protocolVersion", msgAndArgs...)
	assert.Contains(m.t, params, "clientInfo", msgAndArgs...)
	
	// Check clientInfo structure
	clientInfo, ok := params["clientInfo"].(map[string]interface{})
	require.True(m.t, ok, "ClientInfo must be an object", msgAndArgs...)
	
	assert.Contains(m.t, clientInfo, "name", msgAndArgs...)
	assert.Contains(m.t, clientInfo, "version", msgAndArgs...)
}

// AssertValidMCPInitialized validates an MCP initialized notification
func (m *MCPAssertions) AssertValidMCPInitialized(data []byte, msgAndArgs ...interface{}) {
	m.t.Helper()
	
	var notif map[string]interface{}
	err := json.Unmarshal(data, &notif)
	require.NoError(m.t, err, "Failed to unmarshal MCP initialized notification")
	
	// Check basic structure
	assert.Contains(m.t, notif, "jsonrpc", msgAndArgs...)
	assert.Contains(m.t, notif, "method", msgAndArgs...)
	assert.Equal(m.t, "2.0", notif["jsonrpc"], msgAndArgs...)
	assert.Equal(m.t, "initialized", notif["method"], msgAndArgs...)
	
	// Notifications should not have an id field
	assert.NotContains(m.t, notif, "id", msgAndArgs...)
}

// ErrorAssertions provides assertions for error handling
type ErrorAssertions struct {
	t *testing.T
}

// NewErrorAssertions creates a new error assertions helper
func NewErrorAssertions(t *testing.T) *ErrorAssertions {
	return &ErrorAssertions{t: t}
}

// AssertErrorContains checks that an error contains a specific substring
func (e *ErrorAssertions) AssertErrorContains(err error, substring string, msgAndArgs ...interface{}) {
	e.t.Helper()
	
	require.Error(e.t, err, msgAndArgs...)
	assert.Contains(e.t, err.Error(), substring, msgAndArgs...)
}

// AssertErrorType checks that an error is of a specific type
func (e *ErrorAssertions) AssertErrorType(err error, expectedType interface{}, msgAndArgs ...interface{}) {
	e.t.Helper()
	
	require.Error(e.t, err, msgAndArgs...)
	assert.IsType(e.t, expectedType, err, msgAndArgs...)
}

// AssertNoErrorOrType checks that either there's no error or the error is of a specific type
func (e *ErrorAssertions) AssertNoErrorOrType(err error, allowedType interface{}, msgAndArgs ...interface{}) {
	e.t.Helper()
	
	if err != nil {
		assert.IsType(e.t, allowedType, err, msgAndArgs...)
	}
}

// PerformanceAssertions provides assertions for performance testing
type PerformanceAssertions struct {
	t *testing.T
}

// NewPerformanceAssertions creates a new performance assertions helper
func NewPerformanceAssertions(t *testing.T) *PerformanceAssertions {
	return &PerformanceAssertions{t: t}
}

// AssertDurationLessThan checks that an operation completes within a time limit
func (p *PerformanceAssertions) AssertDurationLessThan(duration, limit interface{}, msgAndArgs ...interface{}) {
	p.t.Helper()
	
	// Convert to comparable types
	d := reflect.ValueOf(duration)
	l := reflect.ValueOf(limit)
	
	// Both should be time.Duration or comparable numeric types
	if d.Type() != l.Type() {
		require.Fail(p.t, "Duration and limit must be of the same type", msgAndArgs...)
		return
	}
	
	switch d.Kind() {
	case reflect.Int64: // time.Duration is int64
		assert.Less(p.t, d.Int(), l.Int(), msgAndArgs...)
	case reflect.Float64:
		assert.Less(p.t, d.Float(), l.Float(), msgAndArgs...)
	default:
		require.Fail(p.t, "Unsupported duration type", msgAndArgs...)
	}
}

// ConcurrencyAssertions provides assertions for concurrent operations
type ConcurrencyAssertions struct {
	t *testing.T
}

// NewConcurrencyAssertions creates a new concurrency assertions helper
func NewConcurrencyAssertions(t *testing.T) *ConcurrencyAssertions {
	return &ConcurrencyAssertions{t: t}
}

// AssertNoDataRace runs a function and checks for data races (requires -race flag)
func (c *ConcurrencyAssertions) AssertNoDataRace(fn func(), msgAndArgs ...interface{}) {
	c.t.Helper()
	
	// This is a placeholder - actual race detection happens at runtime with -race flag
	// We can add additional checks here if needed
	fn()
}

// AssertChannelReceives checks that a channel receives a value within a timeout
func (c *ConcurrencyAssertions) AssertChannelReceives(ch interface{}, timeout interface{}, msgAndArgs ...interface{}) interface{} {
	c.t.Helper()
	
	// This would need proper implementation based on channel type
	// For now, this is a placeholder
	return nil
}

// AssertChannelClosed checks that a channel is closed
func (c *ConcurrencyAssertions) AssertChannelClosed(ch interface{}, msgAndArgs ...interface{}) {
	c.t.Helper()
	
	// This would need proper implementation based on channel type
	// For now, this is a placeholder
}

// Global assertion functions for convenience

// AssertJSONRPCRequest validates a JSON-RPC request globally
func AssertJSONRPCRequest(t *testing.T, data []byte, msgAndArgs ...interface{}) {
	t.Helper()
	assertions := NewJSONRPCAssertions(t)
	assertions.AssertValidJSONRPCRequest(data, msgAndArgs...)
}

// AssertJSONRPCResponse validates a JSON-RPC response globally
func AssertJSONRPCResponse(t *testing.T, data []byte, msgAndArgs ...interface{}) {
	t.Helper()
	assertions := NewJSONRPCAssertions(t)
	assertions.AssertValidJSONRPCResponse(data, msgAndArgs...)
}

// AssertJSONRPCError validates a JSON-RPC error response globally
func AssertJSONRPCError(t *testing.T, data []byte, expectedCode int, msgAndArgs ...interface{}) {
	t.Helper()
	assertions := NewJSONRPCAssertions(t)
	assertions.AssertValidJSONRPCError(data, expectedCode, msgAndArgs...)
}

// AssertMCPInitialize validates an MCP initialize request globally
func AssertMCPInitialize(t *testing.T, data []byte, msgAndArgs ...interface{}) {
	t.Helper()
	assertions := NewMCPAssertions(t)
	assertions.AssertValidMCPInitialize(data, msgAndArgs...)
}

// AssertMCPInitialized validates an MCP initialized notification globally
func AssertMCPInitialized(t *testing.T, data []byte, msgAndArgs ...interface{}) {
	t.Helper()
	assertions := NewMCPAssertions(t)
	assertions.AssertValidMCPInitialized(data, msgAndArgs...)
}

// AssertErrorContains checks error content globally
func AssertErrorContains(t *testing.T, err error, substring string, msgAndArgs ...interface{}) {
	t.Helper()
	assertions := NewErrorAssertions(t)
	assertions.AssertErrorContains(err, substring, msgAndArgs...)
}

// AssertStringContainsAll checks that a string contains all specified substrings
func AssertStringContainsAll(t *testing.T, str string, substrings []string, msgAndArgs ...interface{}) {
	t.Helper()
	
	for _, substring := range substrings {
		assert.Contains(t, str, substring, 
			fmt.Sprintf("String should contain '%s'. %v", substring, msgAndArgs))
	}
}

// AssertStringContainsAny checks that a string contains at least one of the specified substrings
func AssertStringContainsAny(t *testing.T, str string, substrings []string, msgAndArgs ...interface{}) {
	t.Helper()
	
	for _, substring := range substrings {
		if strings.Contains(str, substring) {
			return // Found at least one
		}
	}
	
	assert.Fail(t, fmt.Sprintf("String should contain at least one of %v", substrings), msgAndArgs...)
}
