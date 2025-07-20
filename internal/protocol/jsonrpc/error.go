package jsonrpc

// Standard JSON-RPC 2.0 error codes
const (
	// Pre-defined errors
	ErrorCodeParse          = -32700 // Parse error - Invalid JSON was received by the server
	ErrorCodeInvalidRequest = -32600 // Invalid Request - The JSON sent is not a valid Request object
	ErrorCodeMethodNotFound = -32601 // Method not found - The method does not exist / is not available
	ErrorCodeInvalidParams  = -32602 // Invalid params - Invalid method parameter(s)
	ErrorCodeInternal       = -32603 // Internal error - Internal JSON-RPC error

	// Server error range: -32000 to -32099 (reserved for implementation-defined server-errors)
	ErrorCodeServerError     = -32000 // Generic server error
	ErrorCodeNotImplemented  = -32001 // Method not implemented
	ErrorCodeTimeout         = -32002 // Request timeout
	ErrorCodeResourceLimit   = -32003 // Resource limit exceeded
	ErrorCodeUnauthorized    = -32004 // Unauthorized access
	ErrorCodeForbidden       = -32005 // Forbidden operation
	ErrorCodeNotFound        = -32006 // Resource not found
	ErrorCodeConflict        = -32007 // Resource conflict
	ErrorCodeTooManyRequests = -32008 // Rate limit exceeded
	ErrorCodeBadGateway      = -32009 // Bad gateway
	ErrorCodeServiceUnavail  = -32010 // Service unavailable
)

// Error messages for standard error codes
var errorMessages = map[int]string{
	ErrorCodeParse:           "Parse error",
	ErrorCodeInvalidRequest:  "Invalid Request",
	ErrorCodeMethodNotFound:  "Method not found",
	ErrorCodeInvalidParams:   "Invalid params",
	ErrorCodeInternal:        "Internal error",
	ErrorCodeServerError:     "Server error",
	ErrorCodeNotImplemented:  "Method not implemented",
	ErrorCodeTimeout:         "Request timeout",
	ErrorCodeResourceLimit:   "Resource limit exceeded",
	ErrorCodeUnauthorized:    "Unauthorized access",
	ErrorCodeForbidden:       "Forbidden operation",
	ErrorCodeNotFound:        "Resource not found",
	ErrorCodeConflict:        "Resource conflict",
	ErrorCodeTooManyRequests: "Rate limit exceeded",
	ErrorCodeBadGateway:      "Bad gateway",
	ErrorCodeServiceUnavail:  "Service unavailable",
}

// NewError creates a new JSON-RPC error with the given code, message, and optional data
func NewError(code int, message string, data any) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewStandardError creates a new JSON-RPC error using a standard error code
func NewStandardError(code int, data any) *Error {
	message, exists := errorMessages[code]
	if !exists {
		message = "Unknown error"
	}
	return &Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewParseError creates a parse error
func NewParseError(data any) *Error {
	return NewStandardError(ErrorCodeParse, data)
}

// NewInvalidRequestError creates an invalid request error
func NewInvalidRequestError(data any) *Error {
	return NewStandardError(ErrorCodeInvalidRequest, data)
}

// NewMethodNotFoundError creates a method not found error
func NewMethodNotFoundError(method string) *Error {
	return NewError(ErrorCodeMethodNotFound, "Method not found", method)
}

// NewInvalidParamsError creates an invalid params error
func NewInvalidParamsError(data any) *Error {
	return NewStandardError(ErrorCodeInvalidParams, data)
}

// NewInternalError creates an internal error
func NewInternalError(data any) *Error {
	return NewStandardError(ErrorCodeInternal, data)
}

// IsStandardError returns true if the error code is a standard JSON-RPC error
func IsStandardError(code int) bool {
	return code >= -32768 && code <= -32000
}

// IsServerError returns true if the error code is in the server error range
func IsServerError(code int) bool {
	return code >= -32099 && code <= -32000
}

// IsApplicationError returns true if the error code is an application-defined error
func IsApplicationError(code int) bool {
	return code < -32768 || code > -32000
}
