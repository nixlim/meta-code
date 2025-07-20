package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMCPErrorf(t *testing.T) {
	err := NewMCPErrorf(ProtocolError, "test error with %s", "formatting")
	
	assert.Equal(t, "test error with formatting", err.Error())
	assert.Equal(t, ProtocolError, err.Category)
	assert.Equal(t, -32600, err.Code)
}

func TestNewProtocolError(t *testing.T) {
	err := NewProtocolError("invalid request format")
	
	assert.Equal(t, "invalid request format", err.Error())
	assert.Equal(t, ProtocolError, err.Category)
	assert.Equal(t, -32600, err.Code)
}

func TestNewVersionMismatchError(t *testing.T) {
	err := NewVersionMismatchError("2025-03-26", "2024-01-01")
	
	assert.Contains(t, err.Error(), "2025-03-26")
	assert.Contains(t, err.Error(), "2024-01-01")
	assert.Equal(t, ProtocolError, err.Category)
	assert.Equal(t, -32601, err.Code)
}

func TestNewCapabilityError(t *testing.T) {
	err := NewCapabilityError("tools", "Tool capability not supported")
	
	assert.Contains(t, err.Error(), "tools")
	assert.Contains(t, err.Error(), "Tool capability not supported")
	assert.Equal(t, ProtocolError, err.Category)
	assert.Equal(t, -32602, err.Code)
}

func TestNewInitializeError(t *testing.T) {
	err := NewInitializeError("missing client info")
	
	assert.Contains(t, err.Error(), "missing client info")
	assert.Equal(t, ProtocolError, err.Category)
	assert.Equal(t, -32603, err.Code)
}

func TestNewHandshakeTimeoutError(t *testing.T) {
	err := NewHandshakeTimeoutError()
	
	assert.Contains(t, err.Error(), "handshake")
	assert.Contains(t, err.Error(), "timeout")
	assert.Equal(t, ProtocolError, err.Category)
	assert.Equal(t, -32604, err.Code)
}

func TestNewInvalidStateError(t *testing.T) {
	err := NewInvalidStateError("not_initialized", "ready")
	
	assert.Contains(t, err.Error(), "not_initialized")
	assert.Contains(t, err.Error(), "ready")
	assert.Equal(t, ProtocolError, err.Category)
	assert.Equal(t, -32605, err.Code)
}

func TestNewTransportError(t *testing.T) {
	err := NewTransportError("connection failed")
	
	assert.Contains(t, err.Error(), "connection failed")
	assert.Equal(t, TransportError, err.Category)
	assert.Equal(t, -32700, err.Code)
}

func TestNewConnectionLostError(t *testing.T) {
	err := NewConnectionLostError()
	
	assert.Contains(t, err.Error(), "connection")
	assert.Contains(t, err.Error(), "lost")
	assert.Equal(t, TransportError, err.Category)
	assert.Equal(t, -32701, err.Code)
}

func TestNewConnectionFailedError(t *testing.T) {
	err := NewConnectionFailedError("tcp://localhost:8080")
	
	assert.Contains(t, err.Error(), "tcp://localhost:8080")
	assert.Equal(t, TransportError, err.Category)
	assert.Equal(t, -32702, err.Code)
}

func TestNewTransportTimeoutError(t *testing.T) {
	err := NewTransportTimeoutError("5s")
	
	assert.Contains(t, err.Error(), "5s")
	assert.Equal(t, TransportError, err.Category)
	assert.Equal(t, -32703, err.Code)
}

func TestNewMessageTooLargeError(t *testing.T) {
	err := NewMessageTooLargeError(1024, 512)
	
	assert.Contains(t, err.Error(), "1024")
	assert.Contains(t, err.Error(), "512")
	assert.Equal(t, TransportError, err.Category)
	assert.Equal(t, -32704, err.Code)
}

func TestNewEncodingError(t *testing.T) {
	err := NewEncodingError("json", "invalid character")
	
	assert.Contains(t, err.Error(), "json")
	assert.Contains(t, err.Error(), "invalid character")
	assert.Equal(t, TransportError, err.Category)
	assert.Equal(t, -32705, err.Code)
}

func TestNewHandlerError(t *testing.T) {
	err := NewHandlerError("handler execution failed")
	
	assert.Contains(t, err.Error(), "handler execution failed")
	assert.Equal(t, HandlerError, err.Category)
	assert.Equal(t, -32800, err.Code)
}

func TestNewToolNotFoundError(t *testing.T) {
	err := NewToolNotFoundError("test_tool")
	
	assert.Contains(t, err.Error(), "test_tool")
	assert.Equal(t, HandlerError, err.Category)
	assert.Equal(t, -32801, err.Code)
}

func TestNewToolError(t *testing.T) {
	err := NewToolError("test_tool", "execution failed")
	
	assert.Contains(t, err.Error(), "test_tool")
	assert.Contains(t, err.Error(), "execution failed")
	assert.Equal(t, HandlerError, err.Category)
	assert.Equal(t, -32802, err.Code)
}

func TestNewResourceNotFoundError(t *testing.T) {
	err := NewResourceNotFoundError("test_resource")
	
	assert.Contains(t, err.Error(), "test_resource")
	assert.Equal(t, HandlerError, err.Category)
	assert.Equal(t, -32803, err.Code)
}

func TestNewResourceError(t *testing.T) {
	err := NewResourceError("test_resource", "access denied")
	
	assert.Contains(t, err.Error(), "test_resource")
	assert.Contains(t, err.Error(), "access denied")
	assert.Equal(t, HandlerError, err.Category)
	assert.Equal(t, -32804, err.Code)
}

func TestNewPromptNotFoundError(t *testing.T) {
	err := NewPromptNotFoundError("test_prompt")
	
	assert.Contains(t, err.Error(), "test_prompt")
	assert.Equal(t, HandlerError, err.Category)
	assert.Equal(t, -32805, err.Code)
}

func TestNewPromptError(t *testing.T) {
	err := NewPromptError("test_prompt", "validation failed")
	
	assert.Contains(t, err.Error(), "test_prompt")
	assert.Contains(t, err.Error(), "validation failed")
	assert.Equal(t, HandlerError, err.Category)
	assert.Equal(t, -32806, err.Code)
}

func TestNewSecurityError(t *testing.T) {
	err := NewSecurityError("access denied")
	
	assert.Contains(t, err.Error(), "access denied")
	assert.Equal(t, SecurityError, err.Category)
	assert.Equal(t, -32900, err.Code)
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("invalid token")
	
	assert.Contains(t, err.Error(), "invalid token")
	assert.Equal(t, SecurityError, err.Category)
	assert.Equal(t, -32901, err.Code)
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("insufficient permissions")
	
	assert.Contains(t, err.Error(), "insufficient permissions")
	assert.Equal(t, SecurityError, err.Category)
	assert.Equal(t, -32902, err.Code)
}

func TestNewRateLimitError(t *testing.T) {
	err := NewRateLimitError("100", "1m")
	
	assert.Contains(t, err.Error(), "100")
	assert.Contains(t, err.Error(), "1m")
	assert.Equal(t, SecurityError, err.Category)
	assert.Equal(t, -32903, err.Code)
}

func TestNewQuotaExceededError(t *testing.T) {
	err := NewQuotaExceededError("API calls", "1000")
	
	assert.Contains(t, err.Error(), "API calls")
	assert.Contains(t, err.Error(), "1000")
	assert.Equal(t, SecurityError, err.Category)
	assert.Equal(t, -32904, err.Code)
}

func TestNewSystemError(t *testing.T) {
	err := NewSystemError("system overload")
	
	assert.Contains(t, err.Error(), "system overload")
	assert.Equal(t, SystemError, err.Category)
	assert.Equal(t, -33000, err.Code)
}

func TestNewResourceLimitError(t *testing.T) {
	err := NewResourceLimitError("memory", "1GB")
	
	assert.Contains(t, err.Error(), "memory")
	assert.Contains(t, err.Error(), "1GB")
	assert.Equal(t, SystemError, err.Category)
	assert.Equal(t, -33001, err.Code)
}

func TestNewMemoryLimitError(t *testing.T) {
	err := NewMemoryLimitError("1GB", "2GB")
	
	assert.Contains(t, err.Error(), "1GB")
	assert.Contains(t, err.Error(), "2GB")
	assert.Equal(t, SystemError, err.Category)
	assert.Equal(t, -33002, err.Code)
}

func TestNewDiskSpaceError(t *testing.T) {
	err := NewDiskSpaceError("100MB", "/tmp")
	
	assert.Contains(t, err.Error(), "100MB")
	assert.Contains(t, err.Error(), "/tmp")
	assert.Equal(t, SystemError, err.Category)
	assert.Equal(t, -33003, err.Code)
}

func TestNewServiceUnavailableError(t *testing.T) {
	err := NewServiceUnavailableError("maintenance mode")
	
	assert.Contains(t, err.Error(), "maintenance mode")
	assert.Equal(t, SystemError, err.Category)
	assert.Equal(t, -33004, err.Code)
}

func TestFactoryFunctionsReturnCorrectTypes(t *testing.T) {
	tests := []struct {
		name     string
		factory  func() *MCPError
		category ErrorCategory
		code     int
	}{
		{"ProtocolError", func() *MCPError { return NewProtocolError("test") }, ProtocolError, -32600},
		{"TransportError", func() *MCPError { return NewTransportError("test") }, TransportError, -32700},
		{"HandlerError", func() *MCPError { return NewHandlerError("test") }, HandlerError, -32800},
		{"SecurityError", func() *MCPError { return NewSecurityError("test") }, SecurityError, -32900},
		{"SystemError", func() *MCPError { return NewSystemError("test") }, SystemError, -33000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.factory()
			require.NotNil(t, err)
			assert.Equal(t, tt.category, err.Category)
			assert.Equal(t, tt.code, err.Code)
			assert.Implements(t, (*error)(nil), err)
		})
	}
}
