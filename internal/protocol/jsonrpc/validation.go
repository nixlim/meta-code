package jsonrpc

import (
	"strings"
)

// Validate validates a JSON-RPC request
func (r *Request) Validate() error {
	// Check version
	if r.Version != Version {
		return NewInvalidRequestError("jsonrpc field must be \"2.0\"")
	}

	// Check method
	if r.Method == "" {
		return NewInvalidRequestError("method field is required")
	}

	// Method names that begin with "rpc." are reserved for rpc-internal methods
	if strings.HasPrefix(r.Method, "rpc.") {
		return NewMethodNotFoundError(r.Method)
	}

	// Validate ID type (enhanced validation as per expert recommendation)
	switch r.ID.(type) {
	case string, float64, int, int32, int64, uint, uint32, uint64, nil:
		// Valid types - includes both JSON unmarshalled types and native Go types
	default:
		return NewInvalidRequestError("ID must be a string, number, or null")
	}

	return nil
}

// Validate validates a JSON-RPC response
func (r *Response) Validate() error {
	// Check version
	if r.Version != Version {
		return NewInvalidRequestError("jsonrpc field must be \"2.0\"")
	}

	// A response must have either result or error, but not both
	if r.Result != nil && r.Error != nil {
		return NewInvalidRequestError("response cannot have both result and error")
	}

	if r.Result == nil && r.Error == nil {
		return NewInvalidRequestError("response must have either result or error")
	}

	// Validate error if present
	if r.Error != nil {
		if err := r.Error.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates a JSON-RPC notification
func (n *Notification) Validate() error {
	// Check version
	if n.Version != Version {
		return NewInvalidRequestError("jsonrpc field must be \"2.0\"")
	}

	// Check method
	if n.Method == "" {
		return NewInvalidRequestError("method field is required")
	}

	// Method names that begin with "rpc." are reserved for rpc-internal methods
	if strings.HasPrefix(n.Method, "rpc.") {
		return NewMethodNotFoundError(n.Method)
	}

	return nil
}

// Validate validates a JSON-RPC error
func (e *Error) Validate() error {
	// Error code is required (zero is a valid error code)
	if e.Message == "" {
		return NewInvalidRequestError("error message is required")
	}

	return nil
}

// ValidateID checks if an ID is valid for JSON-RPC
func ValidateID(id any) bool {
	switch id.(type) {
	case string, float64, int, int32, int64, uint, uint32, uint64, nil:
		return true
	default:
		return false
	}
}

// ValidateMethod checks if a method name is valid for JSON-RPC
func ValidateMethod(method string) bool {
	if method == "" {
		return false
	}

	// Method names that begin with "rpc." are reserved
	if strings.HasPrefix(method, "rpc.") {
		return false
	}

	return true
}
