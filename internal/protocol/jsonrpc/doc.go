// Package jsonrpc provides a complete JSON-RPC 2.0 implementation for the Meta-MCP server.
//
// This package implements the JSON-RPC 2.0 specification as defined at:
// https://www.jsonrpc.org/specification
//
// Features:
//   - Complete JSON-RPC 2.0 compliance
//   - Support for requests, responses, notifications, and errors
//   - Batch request/response handling
//   - Comprehensive validation
//   - Transport-agnostic design
//   - Enhanced parameter binding utilities
//
// Basic Usage:
//
//	// Create a request
//	req := jsonrpc.NewRequest("test_method", map[string]any{"key": "value"}, 1)
//
//	// Validate the request
//	if err := req.Validate(); err != nil {
//		// handle validation error
//	}
//
//	// Parse a JSON-RPC message
//	msg, err := jsonrpc.ParseMessage([]byte(`{"jsonrpc":"2.0","method":"test","id":1}`))
//	if err != nil {
//		// handle parse error
//	}
//
//	// Handle batch requests
//	messages, err := jsonrpc.Parse([]byte(`[{"jsonrpc":"2.0","method":"test1","id":1},{"jsonrpc":"2.0","method":"test2","id":2}]`))
//	if err != nil {
//		// handle parse error
//	}
//
//	// Bind parameters to a struct
//	type Params struct {
//		Name string `json:"name"`
//		Age  int    `json:"age"`
//	}
//	var params Params
//	if err := req.BindParams(&params); err != nil {
//		// handle binding error
//	}
//
// Error Handling:
//
// The package provides standard JSON-RPC error codes and utilities for creating
// appropriate error responses:
//
//	// Create standard errors
//	parseErr := jsonrpc.NewParseError("Invalid JSON")
//	methodErr := jsonrpc.NewMethodNotFoundError("unknown_method")
//	paramsErr := jsonrpc.NewInvalidParamsError("missing required parameter")
//
//	// Create custom errors
//	customErr := jsonrpc.NewError(-32001, "Custom error", "additional data")
//
// Transport Abstraction:
//
// The package defines interfaces for transport mechanisms, allowing the same
// JSON-RPC implementation to work over different transports (STDIO, HTTP, WebSocket, etc.):
//
//	type Transport interface {
//		Send(ctx context.Context, message Message) error
//		Receive(ctx context.Context) (Message, error)
//		// ... other methods
//	}
//
// This foundation serves as the base protocol layer for the Meta-MCP server,
// enabling communication with MCP clients and servers using the standard
// JSON-RPC 2.0 protocol.
package jsonrpc
