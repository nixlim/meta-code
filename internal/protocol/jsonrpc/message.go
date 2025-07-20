package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// Version represents the JSON-RPC version
const Version = "2.0"

// Request represents a JSON-RPC 2.0 request message
type Request struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
	ID      any    `json:"id,omitempty"`
}

// Response represents a JSON-RPC 2.0 response message
type Response struct {
	Version string `json:"jsonrpc"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
	ID      any    `json:"id"`
}

// Notification represents a JSON-RPC 2.0 notification message
type Notification struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// Error represents a JSON-RPC 2.0 error object
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("JSON-RPC error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// NewRequest creates a new JSON-RPC request
func NewRequest(method string, params any, id any) *Request {
	return &Request{
		Version: Version,
		Method:  method,
		Params:  params,
		ID:      id,
	}
}

// NewNotification creates a new JSON-RPC notification
func NewNotification(method string, params any) *Notification {
	return &Notification{
		Version: Version,
		Method:  method,
		Params:  params,
	}
}

// NewResponse creates a new JSON-RPC response with result
func NewResponse(result any, id any) *Response {
	return &Response{
		Version: Version,
		Result:  result,
		ID:      id,
	}
}

// NewErrorResponse creates a new JSON-RPC response with error
func NewErrorResponse(err *Error, id any) *Response {
	return &Response{
		Version: Version,
		Error:   err,
		ID:      id,
	}
}

// IsRequest returns true if this is a request (has ID and is not a notification)
func (r *Request) IsRequest() bool {
	return r.ID != nil
}

// IsNotification returns true if this is a notification (no ID)
func (r *Request) IsNotification() bool {
	return r.ID == nil
}

// BindParams unmarshals the params from a request into a given struct.
// This simplifies handling of named or positional parameters.
func (r *Request) BindParams(v any) error {
	if r.Params == nil {
		// No params, nothing to bind
		return nil
	}

	// Re-marshal and unmarshal to convert from any to specific struct
	paramsBytes, err := json.Marshal(r.Params)
	if err != nil {
		return NewError(ErrorCodeInternal, "Failed to re-marshal params", err.Error())
	}

	if err := json.Unmarshal(paramsBytes, v); err != nil {
		return NewError(ErrorCodeInvalidParams, "Failed to bind params to target", err.Error())
	}

	return nil
}

// HasResult returns true if the response contains a result
func (r *Response) HasResult() bool {
	return r.Error == nil && r.Result != nil
}

// HasError returns true if the response contains an error
func (r *Response) HasError() bool {
	return r.Error != nil
}
