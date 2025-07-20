package errors

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// MCP-specific error codes in the reserved range (-32000 to -32099)
// These extend the JSON-RPC error codes for MCP protocol-specific errors
const (
	// Protocol-level errors (-32000 to -32019)
	ErrorCodeMCPProtocol         = -32000 // Generic MCP protocol error
	ErrorCodeMCPVersionMismatch  = -32001 // Protocol version mismatch
	ErrorCodeMCPCapabilityError  = -32002 // Capability negotiation error
	ErrorCodeMCPInitializeError  = -32003 // Initialization sequence error
	ErrorCodeMCPHandshakeTimeout = -32004 // Handshake timeout
	ErrorCodeMCPInvalidState     = -32005 // Invalid protocol state

	// Transport-level errors (-32020 to -32039)
	ErrorCodeMCPTransport        = -32020 // Generic transport error
	ErrorCodeMCPConnectionLost   = -32021 // Connection lost
	ErrorCodeMCPConnectionFailed = -32022 // Connection failed
	ErrorCodeMCPTransportTimeout = -32023 // Transport timeout
	ErrorCodeMCPMessageTooLarge  = -32024 // Message size exceeded
	ErrorCodeMCPEncodingError    = -32025 // Message encoding error

	// Handler-level errors (-32040 to -32059)
	ErrorCodeMCPHandler          = -32040 // Generic handler error
	ErrorCodeMCPToolNotFound     = -32041 // Tool not found
	ErrorCodeMCPToolError        = -32042 // Tool execution error
	ErrorCodeMCPResourceNotFound = -32043 // Resource not found
	ErrorCodeMCPResourceError    = -32044 // Resource access error
	ErrorCodeMCPPromptNotFound   = -32045 // Prompt not found
	ErrorCodeMCPPromptError      = -32046 // Prompt execution error

	// Security and authorization errors (-32060 to -32079)
	ErrorCodeMCPSecurity      = -32060 // Generic security error
	ErrorCodeMCPUnauthorized  = -32061 // Unauthorized access
	ErrorCodeMCPForbidden     = -32062 // Forbidden operation
	ErrorCodeMCPRateLimit     = -32063 // Rate limit exceeded
	ErrorCodeMCPQuotaExceeded = -32064 // Quota exceeded

	// System and resource errors (-32080 to -32099)
	ErrorCodeMCPSystem         = -32080 // Generic system error
	ErrorCodeMCPResourceLimit  = -32081 // Resource limit exceeded
	ErrorCodeMCPMemoryLimit    = -32082 // Memory limit exceeded
	ErrorCodeMCPDiskSpace      = -32083 // Disk space exceeded
	ErrorCodeMCPServiceUnavail = -32084 // Service unavailable
)

// MCPError represents an MCP-specific error that extends JSON-RPC errors
type MCPError struct {
	Code      int                    `json:"code"`
	Message   string                 `json:"message"`
	Data      interface{}            `json:"data,omitempty"`
	Category  string                 `json:"category,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Cause     error                  `json:"-"` // Original error, not serialized
	DebugInfo map[string]interface{} `json:"debugInfo,omitempty"`
	Sanitized bool                   `json:"-"` // Whether error has been sanitized
}

// Error implements the error interface
func (e *MCPError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("MCP %s error (%d): %s - caused by: %v", e.Category, e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("MCP %s error (%d): %s", e.Category, e.Code, e.Message)
}

// Unwrap returns the underlying cause for error chain support
func (e *MCPError) Unwrap() error {
	return e.Cause
}

// Is implements error comparison for errors.Is
func (e *MCPError) Is(target error) bool {
	if mcpErr, ok := target.(*MCPError); ok {
		return e.Code == mcpErr.Code
	}
	return false
}

// As implements error type assertion for errors.As
func (e *MCPError) As(target interface{}) bool {
	if mcpErr, ok := target.(**MCPError); ok {
		*mcpErr = e
		return true
	}
	if jsonrpcErr, ok := target.(**jsonrpc.Error); ok {
		*jsonrpcErr = &jsonrpc.Error{
			Code:    e.Code,
			Message: e.Message,
			Data:    e.Data,
		}
		return true
	}
	return false
}

// WithContext adds context information to the error
func (e *MCPError) WithContext(key string, value interface{}) *MCPError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDebugInfo adds debug information (only included in debug mode)
func (e *MCPError) WithDebugInfo(key string, value interface{}) *MCPError {
	if e.DebugInfo == nil {
		e.DebugInfo = make(map[string]interface{})
	}
	e.DebugInfo[key] = value
	return e
}

// ToJSONRPCError converts to a standard JSON-RPC error
func (e *MCPError) ToJSONRPCError() *jsonrpc.Error {
	return &jsonrpc.Error{
		Code:    e.Code,
		Message: e.Message,
		Data:    e.Data,
	}
}

// ToMCPError converts to mcp-go JSONRPCError format
func (e *MCPError) ToMCPError(id mcp.RequestId) mcp.JSONRPCError {
	return mcp.NewJSONRPCError(id, e.Code, e.Message, e.Data)
}

// GetCategory returns the error category based on error code
func GetCategory(code int) string {
	switch {
	case code >= -32019 && code <= -32000:
		return "protocol"
	case code >= -32039 && code <= -32020:
		return "transport"
	case code >= -32059 && code <= -32040:
		return "handler"
	case code >= -32079 && code <= -32060:
		return "security"
	case code >= -32099 && code <= -32080:
		return "system"
	default:
		return "unknown"
	}
}

// IsMCPError returns true if the error code is in the MCP range
func IsMCPError(code int) bool {
	return code >= -32099 && code <= -32000
}

// Error messages for MCP error codes
var mcpErrorMessages = map[int]string{
	// Protocol errors
	ErrorCodeMCPProtocol:         "MCP protocol error",
	ErrorCodeMCPVersionMismatch:  "Protocol version mismatch",
	ErrorCodeMCPCapabilityError:  "Capability negotiation error",
	ErrorCodeMCPInitializeError:  "Initialization sequence error",
	ErrorCodeMCPHandshakeTimeout: "Handshake timeout",
	ErrorCodeMCPInvalidState:     "Invalid protocol state",

	// Transport errors
	ErrorCodeMCPTransport:        "Transport error",
	ErrorCodeMCPConnectionLost:   "Connection lost",
	ErrorCodeMCPConnectionFailed: "Connection failed",
	ErrorCodeMCPTransportTimeout: "Transport timeout",
	ErrorCodeMCPMessageTooLarge:  "Message size exceeded",
	ErrorCodeMCPEncodingError:    "Message encoding error",

	// Handler errors
	ErrorCodeMCPHandler:          "Handler error",
	ErrorCodeMCPToolNotFound:     "Tool not found",
	ErrorCodeMCPToolError:        "Tool execution error",
	ErrorCodeMCPResourceNotFound: "Resource not found",
	ErrorCodeMCPResourceError:    "Resource access error",
	ErrorCodeMCPPromptNotFound:   "Prompt not found",
	ErrorCodeMCPPromptError:      "Prompt execution error",

	// Security errors
	ErrorCodeMCPSecurity:      "Security error",
	ErrorCodeMCPUnauthorized:  "Unauthorized access",
	ErrorCodeMCPForbidden:     "Forbidden operation",
	ErrorCodeMCPRateLimit:     "Rate limit exceeded",
	ErrorCodeMCPQuotaExceeded: "Quota exceeded",

	// System errors
	ErrorCodeMCPSystem:         "System error",
	ErrorCodeMCPResourceLimit:  "Resource limit exceeded",
	ErrorCodeMCPMemoryLimit:    "Memory limit exceeded",
	ErrorCodeMCPDiskSpace:      "Disk space exceeded",
	ErrorCodeMCPServiceUnavail: "Service unavailable",
}

// GetMCPErrorMessage returns the standard message for an MCP error code
func GetMCPErrorMessage(code int) string {
	if message, exists := mcpErrorMessages[code]; exists {
		return message
	}
	return "Unknown MCP error"
}

// Sanitize removes sensitive information from the error for production use
func (e *MCPError) Sanitize() *MCPError {
	if e.Sanitized {
		return e // Already sanitized
	}

	sanitized := &MCPError{
		Code:      e.Code,
		Message:   e.Message,
		Data:      nil, // Remove data to prevent leaks
		Category:  e.Category,
		Context:   make(map[string]interface{}),
		Sanitized: true,
	}

	// Copy only non-sensitive context
	sensitiveKeys := []string{
		"password", "token", "secret", "auth", "credential",
		"session", "cookie", "bearer", "api_key", "private",
		"access_key", "secret_key", "private_key", "public_key",
	}

	for k, v := range e.Context {
		isSensitive := false
		for _, sensitive := range sensitiveKeys {
			if strings.Contains(strings.ToLower(k), sensitive) {
				isSensitive = true
				break
			}
		}

		if !isSensitive {
			sanitized.Context[k] = v
		}
	}

	// Don't include debug info or cause in sanitized version
	return sanitized
}

// Clone creates a deep copy of the MCPError
func (e *MCPError) Clone() *MCPError {
	clone := &MCPError{
		Code:      e.Code,
		Message:   e.Message,
		Data:      e.Data,
		Category:  e.Category,
		Cause:     e.Cause,
		Sanitized: e.Sanitized,
	}

	// Deep copy context
	if e.Context != nil {
		clone.Context = make(map[string]interface{})
		for k, v := range e.Context {
			clone.Context[k] = v
		}
	}

	// Deep copy debug info
	if e.DebugInfo != nil {
		clone.DebugInfo = make(map[string]interface{})
		for k, v := range e.DebugInfo {
			clone.DebugInfo[k] = v
		}
	}

	return clone
}

// WithCause sets the underlying cause error
func (e *MCPError) WithCause(cause error) *MCPError {
	e.Cause = cause
	return e
}

// HasContext checks if a context key exists
func (e *MCPError) HasContext(key string) bool {
	_, exists := e.Context[key]
	return exists
}

// GetContext retrieves a context value
func (e *MCPError) GetContext(key string) (interface{}, bool) {
	value, exists := e.Context[key]
	return value, exists
}

// GetContextString retrieves a context value as a string
func (e *MCPError) GetContextString(key string) (string, bool) {
	if value, exists := e.Context[key]; exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// RemoveContext removes a context key
func (e *MCPError) RemoveContext(key string) *MCPError {
	delete(e.Context, key)
	return e
}

// ClearContext removes all context information
func (e *MCPError) ClearContext() *MCPError {
	e.Context = make(map[string]interface{})
	return e
}

// ClearDebugInfo removes all debug information
func (e *MCPError) ClearDebugInfo() *MCPError {
	e.DebugInfo = make(map[string]interface{})
	return e
}
