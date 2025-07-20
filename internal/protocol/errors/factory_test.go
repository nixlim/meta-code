package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProtocolError(t *testing.T) {
	err := NewProtocolError("invalid request format", nil)

	assert.Contains(t, err.Error(), "invalid request format")
	assert.Equal(t, ErrorCodeMCPProtocol, err.Code)
}

func TestNewVersionMismatchError(t *testing.T) {
	err := NewVersionMismatchError("2025-03-26", "2024-01-01")

	assert.Contains(t, err.Error(), "2025-03-26")
	assert.Contains(t, err.Error(), "2024-01-01")
	assert.Equal(t, ErrorCodeMCPVersionMismatch, err.Code)
}

func TestNewCapabilityError(t *testing.T) {
	err := NewCapabilityError("tools", "Tool capability not supported")

	assert.Contains(t, err.Error(), "Tool capability not supported")
	assert.Equal(t, ErrorCodeMCPCapabilityError, err.Code)
}

func TestNewInitializeError(t *testing.T) {
	err := NewInitializeError("missing client info", nil)

	assert.Contains(t, err.Error(), "missing client info")
	assert.Equal(t, ErrorCodeMCPInitializeError, err.Code)
}

func TestNewHandshakeTimeoutError(t *testing.T) {
	err := NewHandshakeTimeoutError("30s")

	assert.Contains(t, err.Error(), "Handshake")
	assert.Contains(t, err.Error(), "timeout")
	assert.Equal(t, ErrorCodeMCPHandshakeTimeout, err.Code)
}

func TestNewInvalidStateError(t *testing.T) {
	err := NewInvalidStateError("not_initialized", "ready")

	assert.Contains(t, err.Error(), "not_initialized")
	assert.Contains(t, err.Error(), "ready")
	assert.Equal(t, ErrorCodeMCPInvalidState, err.Code)
}

func TestNewTransportError(t *testing.T) {
	err := NewTransportError("connection failed", nil)

	assert.Contains(t, err.Error(), "connection failed")
	assert.Equal(t, ErrorCodeMCPTransport, err.Code)
}

func TestNewConnectionLostError(t *testing.T) {
	err := NewConnectionLostError("network timeout")

	assert.Contains(t, err.Error(), "Connection lost")
	assert.Equal(t, ErrorCodeMCPConnectionLost, err.Code)
}

func TestNewConnectionFailedError(t *testing.T) {
	err := NewConnectionFailedError("tcp://localhost:8080", nil)

	assert.Contains(t, err.Error(), "Connection failed")
	assert.Equal(t, ErrorCodeMCPConnectionFailed, err.Code)
}

func TestNewTransportTimeoutError(t *testing.T) {
	err := NewTransportTimeoutError("read", "5s")

	assert.Contains(t, err.Error(), "timeout")
	assert.Equal(t, ErrorCodeMCPTransportTimeout, err.Code)
}

func TestNewMessageTooLargeError(t *testing.T) {
	err := NewMessageTooLargeError(1024, 512)

	assert.Contains(t, err.Error(), "1024")
	assert.Contains(t, err.Error(), "512")
	assert.Equal(t, ErrorCodeMCPMessageTooLarge, err.Code)
}

func TestNewEncodingError(t *testing.T) {
	err := NewEncodingError("json", nil)

	assert.Contains(t, err.Error(), "json")
	assert.Equal(t, ErrorCodeMCPEncodingError, err.Code)
}

func TestNewHandlerError(t *testing.T) {
	err := NewHandlerError("handler execution failed", nil)

	assert.Contains(t, err.Error(), "handler execution failed")
	assert.Equal(t, ErrorCodeMCPHandler, err.Code)
}

func TestNewToolNotFoundError(t *testing.T) {
	err := NewToolNotFoundError("test_tool")

	assert.Contains(t, err.Error(), "test_tool")
	assert.Equal(t, ErrorCodeMCPToolNotFound, err.Code)
}

func TestNewToolError(t *testing.T) {
	err := NewToolError("test_tool", nil)

	assert.Contains(t, err.Error(), "test_tool")
	assert.Equal(t, ErrorCodeMCPToolError, err.Code)
}

func TestNewResourceNotFoundError(t *testing.T) {
	err := NewResourceNotFoundError("test_resource")

	assert.Contains(t, err.Error(), "test_resource")
	assert.Equal(t, ErrorCodeMCPResourceNotFound, err.Code)
}

func TestNewResourceError(t *testing.T) {
	err := NewResourceError("test_resource", nil)

	assert.Contains(t, err.Error(), "test_resource")
	assert.Equal(t, ErrorCodeMCPResourceError, err.Code)
}

func TestNewPromptNotFoundError(t *testing.T) {
	err := NewPromptNotFoundError("test_prompt")

	assert.Contains(t, err.Error(), "test_prompt")
	assert.Equal(t, ErrorCodeMCPPromptNotFound, err.Code)
}

func TestNewPromptError(t *testing.T) {
	err := NewPromptError("test_prompt", nil)

	assert.Contains(t, err.Error(), "test_prompt")
	assert.Equal(t, ErrorCodeMCPPromptError, err.Code)
}

func TestNewSecurityError(t *testing.T) {
	err := NewSecurityError("access denied", nil)

	assert.Contains(t, err.Error(), "access denied")
	assert.Equal(t, ErrorCodeMCPSecurity, err.Code)
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("protected_resource")

	assert.Contains(t, err.Error(), "Unauthorized")
	assert.Equal(t, ErrorCodeMCPUnauthorized, err.Code)
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("delete_operation")

	assert.Contains(t, err.Error(), "Forbidden")
	assert.Equal(t, ErrorCodeMCPForbidden, err.Code)
}

func TestNewRateLimitError(t *testing.T) {
	err := NewRateLimitError(100, "1m")

	assert.Contains(t, err.Error(), "Rate limit")
	assert.Equal(t, ErrorCodeMCPRateLimit, err.Code)
}

func TestNewQuotaExceededError(t *testing.T) {
	err := NewQuotaExceededError("API calls", 1000, 500)

	assert.Contains(t, err.Error(), "API calls")
	assert.Contains(t, err.Error(), "1000")
	assert.Contains(t, err.Error(), "500")
	assert.Equal(t, ErrorCodeMCPQuotaExceeded, err.Code)
}

func TestNewSystemError(t *testing.T) {
	err := NewSystemError("system overload", nil)

	assert.Contains(t, err.Error(), "system overload")
	assert.Equal(t, ErrorCodeMCPSystem, err.Code)
}

func TestNewResourceLimitError(t *testing.T) {
	err := NewResourceLimitError("memory", 1024, 512)

	assert.Contains(t, err.Error(), "memory")
	assert.Contains(t, err.Error(), "1024")
	assert.Contains(t, err.Error(), "512")
	assert.Equal(t, ErrorCodeMCPResourceLimit, err.Code)
}

func TestNewMemoryLimitError(t *testing.T) {
	err := NewMemoryLimitError(1024, 512)

	assert.Contains(t, err.Error(), "1024")
	assert.Contains(t, err.Error(), "512")
	assert.Equal(t, ErrorCodeMCPMemoryLimit, err.Code)
}

func TestNewDiskSpaceError(t *testing.T) {
	err := NewDiskSpaceError("/tmp", 1024, 512)

	assert.Contains(t, err.Error(), "/tmp")
	assert.Contains(t, err.Error(), "1024")
	assert.Contains(t, err.Error(), "512")
	assert.Equal(t, ErrorCodeMCPDiskSpace, err.Code)
}

func TestNewServiceUnavailableError(t *testing.T) {
	err := NewServiceUnavailableError("database", "maintenance mode")

	assert.Contains(t, err.Error(), "Service unavailable")
	assert.Equal(t, ErrorCodeMCPServiceUnavail, err.Code)
}

func TestNewMCPErrorf(t *testing.T) {
	err := NewMCPErrorf(ErrorCodeMCPProtocol, "test error with %s", "formatting")

	assert.Contains(t, err.Error(), "test error with formatting")
	assert.Equal(t, ErrorCodeMCPProtocol, err.Code)
}

func TestFactoryFunctionsReturnCorrectTypes(t *testing.T) {
	tests := []struct {
		name    string
		factory func() *MCPError
		code    int
	}{
		{"ProtocolError", func() *MCPError { return NewProtocolError("test", nil) }, ErrorCodeMCPProtocol},
		{"TransportError", func() *MCPError { return NewTransportError("test", nil) }, ErrorCodeMCPTransport},
		{"HandlerError", func() *MCPError { return NewHandlerError("test", nil) }, ErrorCodeMCPHandler},
		{"SecurityError", func() *MCPError { return NewSecurityError("test", nil) }, ErrorCodeMCPSecurity},
		{"SystemError", func() *MCPError { return NewSystemError("test", nil) }, ErrorCodeMCPSystem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.factory()
			require.NotNil(t, err)
			assert.Equal(t, tt.code, err.Code)
			assert.Implements(t, (*error)(nil), err)
		})
	}
}
