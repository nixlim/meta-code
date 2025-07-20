package errors

import (
	"fmt"
)

// NewMCPError creates a new MCP error with the given code and message
func NewMCPError(code int, message string, data interface{}) *MCPError {
	category := GetCategory(code)

	// Use standard message if none provided
	if message == "" {
		message = GetMCPErrorMessage(code)
	}

	return &MCPError{
		Code:     code,
		Message:  message,
		Data:     data,
		Category: category,
		Context:  make(map[string]interface{}),
	}
}

// NewMCPErrorf creates a new MCP error with a formatted message
func NewMCPErrorf(code int, format string, args ...interface{}) *MCPError {
	message := fmt.Sprintf(format, args...)
	return NewMCPError(code, message, nil)
}

// Protocol Error Factories

// NewProtocolError creates a generic protocol error
func NewProtocolError(message string, data interface{}) *MCPError {
	return NewMCPError(ErrorCodeMCPProtocol, message, data)
}

// NewVersionMismatchError creates a version mismatch error
func NewVersionMismatchError(clientVersion, serverVersion string) *MCPError {
	err := NewMCPError(ErrorCodeMCPVersionMismatch,
		fmt.Sprintf("Protocol version mismatch: client=%s, server=%s", clientVersion, serverVersion),
		nil)
	err.WithContext("client_version", clientVersion)
	err.WithContext("server_version", serverVersion)
	return err
}

// NewCapabilityError creates a capability negotiation error
func NewCapabilityError(capability string, message string) *MCPError {
	err := NewMCPError(ErrorCodeMCPCapabilityError, message, nil)
	err.WithContext("capability", capability)
	return err
}

// NewInitializeError creates an initialization error
func NewInitializeError(message string, data interface{}) *MCPError {
	return NewMCPError(ErrorCodeMCPInitializeError, message, data)
}

// NewHandshakeTimeoutError creates a handshake timeout error
func NewHandshakeTimeoutError(timeout string) *MCPError {
	err := NewMCPError(ErrorCodeMCPHandshakeTimeout, "Handshake timeout", nil)
	err.WithContext("timeout", timeout)
	return err
}

// NewInvalidStateError creates an invalid state error
func NewInvalidStateError(currentState, expectedState string) *MCPError {
	err := NewMCPError(ErrorCodeMCPInvalidState,
		fmt.Sprintf("Invalid protocol state: current=%s, expected=%s", currentState, expectedState),
		nil)
	err.WithContext("current_state", currentState)
	err.WithContext("expected_state", expectedState)
	return err
}

// Transport Error Factories

// NewTransportError creates a generic transport error
func NewTransportError(message string, data interface{}) *MCPError {
	return NewMCPError(ErrorCodeMCPTransport, message, data)
}

// NewConnectionLostError creates a connection lost error
func NewConnectionLostError(reason string) *MCPError {
	err := NewMCPError(ErrorCodeMCPConnectionLost, "Connection lost", nil)
	if reason != "" {
		err.WithContext("reason", reason)
	}
	return err
}

// NewConnectionFailedError creates a connection failed error
func NewConnectionFailedError(address string, reason error) *MCPError {
	err := NewMCPError(ErrorCodeMCPConnectionFailed, "Connection failed", nil)
	err.WithContext("address", address)
	if reason != nil {
		err.Cause = reason
		err.WithContext("reason", reason.Error())
	}
	return err
}

// NewTransportTimeoutError creates a transport timeout error
func NewTransportTimeoutError(operation string, timeout string) *MCPError {
	err := NewMCPError(ErrorCodeMCPTransportTimeout,
		fmt.Sprintf("Transport timeout during %s", operation), nil)
	err.WithContext("operation", operation)
	err.WithContext("timeout", timeout)
	return err
}

// NewMessageTooLargeError creates a message too large error
func NewMessageTooLargeError(size, maxSize int64) *MCPError {
	err := NewMCPError(ErrorCodeMCPMessageTooLarge,
		fmt.Sprintf("Message size %d exceeds maximum %d", size, maxSize), nil)
	err.WithContext("message_size", size)
	err.WithContext("max_size", maxSize)
	return err
}

// NewEncodingError creates a message encoding error
func NewEncodingError(format string, cause error) *MCPError {
	err := NewMCPError(ErrorCodeMCPEncodingError,
		fmt.Sprintf("Message encoding error: %s", format), nil)
	err.WithContext("format", format)
	if cause != nil {
		err.Cause = cause
	}
	return err
}

// Handler Error Factories

// NewHandlerError creates a generic handler error
func NewHandlerError(message string, data interface{}) *MCPError {
	return NewMCPError(ErrorCodeMCPHandler, message, data)
}

// NewToolNotFoundError creates a tool not found error
func NewToolNotFoundError(toolName string) *MCPError {
	err := NewMCPError(ErrorCodeMCPToolNotFound,
		fmt.Sprintf("Tool not found: %s", toolName), nil)
	err.WithContext("tool_name", toolName)
	return err
}

// NewToolError creates a tool execution error
func NewToolError(toolName string, cause error) *MCPError {
	err := NewMCPError(ErrorCodeMCPToolError,
		fmt.Sprintf("Tool execution error: %s", toolName), nil)
	err.WithContext("tool_name", toolName)
	if cause != nil {
		err.Cause = cause
	}
	return err
}

// NewResourceNotFoundError creates a resource not found error
func NewResourceNotFoundError(resourceURI string) *MCPError {
	err := NewMCPError(ErrorCodeMCPResourceNotFound,
		fmt.Sprintf("Resource not found: %s", resourceURI), nil)
	err.WithContext("resource_uri", resourceURI)
	return err
}

// NewResourceError creates a resource access error
func NewResourceError(resourceURI string, cause error) *MCPError {
	err := NewMCPError(ErrorCodeMCPResourceError,
		fmt.Sprintf("Resource access error: %s", resourceURI), nil)
	err.WithContext("resource_uri", resourceURI)
	if cause != nil {
		err.Cause = cause
	}
	return err
}

// NewPromptNotFoundError creates a prompt not found error
func NewPromptNotFoundError(promptName string) *MCPError {
	err := NewMCPError(ErrorCodeMCPPromptNotFound,
		fmt.Sprintf("Prompt not found: %s", promptName), nil)
	err.WithContext("prompt_name", promptName)
	return err
}

// NewPromptError creates a prompt execution error
func NewPromptError(promptName string, cause error) *MCPError {
	err := NewMCPError(ErrorCodeMCPPromptError,
		fmt.Sprintf("Prompt execution error: %s", promptName), nil)
	err.WithContext("prompt_name", promptName)
	if cause != nil {
		err.Cause = cause
	}
	return err
}

// Security Error Factories

// NewSecurityError creates a generic security error
func NewSecurityError(message string, data interface{}) *MCPError {
	return NewMCPError(ErrorCodeMCPSecurity, message, data)
}

// NewUnauthorizedError creates an unauthorized access error
func NewUnauthorizedError(resource string) *MCPError {
	err := NewMCPError(ErrorCodeMCPUnauthorized, "Unauthorized access", nil)
	if resource != "" {
		err.WithContext("resource", resource)
	}
	return err
}

// NewForbiddenError creates a forbidden operation error
func NewForbiddenError(operation string) *MCPError {
	err := NewMCPError(ErrorCodeMCPForbidden, "Forbidden operation", nil)
	if operation != "" {
		err.WithContext("operation", operation)
	}
	return err
}

// NewRateLimitError creates a rate limit exceeded error
func NewRateLimitError(limit int, window string) *MCPError {
	err := NewMCPError(ErrorCodeMCPRateLimit, "Rate limit exceeded", nil)
	err.WithContext("limit", limit)
	err.WithContext("window", window)
	return err
}

// NewQuotaExceededError creates a quota exceeded error
func NewQuotaExceededError(quotaType string, used, limit int64) *MCPError {
	err := NewMCPError(ErrorCodeMCPQuotaExceeded,
		fmt.Sprintf("Quota exceeded for %s: %d/%d", quotaType, used, limit), nil)
	err.WithContext("quota_type", quotaType)
	err.WithContext("used", used)
	err.WithContext("limit", limit)
	return err
}

// System Error Factories

// NewSystemError creates a generic system error
func NewSystemError(message string, data interface{}) *MCPError {
	return NewMCPError(ErrorCodeMCPSystem, message, data)
}

// NewResourceLimitError creates a resource limit exceeded error
func NewResourceLimitError(resource string, used, limit int64) *MCPError {
	err := NewMCPError(ErrorCodeMCPResourceLimit,
		fmt.Sprintf("Resource limit exceeded for %s: %d/%d", resource, used, limit), nil)
	err.WithContext("resource", resource)
	err.WithContext("used", used)
	err.WithContext("limit", limit)
	return err
}

// NewMemoryLimitError creates a memory limit exceeded error
func NewMemoryLimitError(used, limit int64) *MCPError {
	err := NewMCPError(ErrorCodeMCPMemoryLimit,
		fmt.Sprintf("Memory limit exceeded: %d/%d bytes", used, limit), nil)
	err.WithContext("used_bytes", used)
	err.WithContext("limit_bytes", limit)
	return err
}

// NewDiskSpaceError creates a disk space exceeded error
func NewDiskSpaceError(path string, used, available int64) *MCPError {
	err := NewMCPError(ErrorCodeMCPDiskSpace,
		fmt.Sprintf("Disk space exceeded on %s: %d bytes used, %d available", path, used, available), nil)
	err.WithContext("path", path)
	err.WithContext("used_bytes", used)
	err.WithContext("available_bytes", available)
	return err
}

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(service string, reason string) *MCPError {
	err := NewMCPError(ErrorCodeMCPServiceUnavail,
		fmt.Sprintf("Service unavailable: %s", service), nil)
	err.WithContext("service", service)
	if reason != "" {
		err.WithContext("reason", reason)
	}
	return err
}
