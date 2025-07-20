// Package errors provides MCP-specific error handling extensions for the Meta-MCP Server.
//
// This package extends the JSON-RPC error handling foundation with MCP protocol-specific
// error codes, error wrapping utilities, structured logging, and context management.
// It builds upon the mcp-go library integration while providing additional functionality
// for debugging, error context, and production-safe error reporting.
//
// # Error Categories
//
// MCP errors are organized into categories based on their error code ranges:
//
//   - Protocol errors (-32000 to -32019): Version mismatches, capability errors, handshake issues
//   - Transport errors (-32020 to -32039): Connection issues, timeouts, encoding problems
//   - Handler errors (-32040 to -32059): Tool/resource/prompt execution errors
//   - Security errors (-32060 to -32079): Authentication, authorization, rate limiting
//   - System errors (-32080 to -32099): Resource limits, system unavailability
//
// # Basic Usage
//
// Creating MCP errors:
//
//	// Using factory functions
//	err := NewToolNotFoundError("echo")
//	err := NewConnectionFailedError("localhost:8080", cause)
//
//	// Using generic constructor
//	err := NewMCPError(ErrorCodeMCPProtocol, "Custom message", nil)
//
// Adding context and debug information:
//
//	err := NewToolError("calculator", cause)
//	err.WithContext("operation", "divide")
//	err.WithContext("arguments", map[string]interface{}{"a": 10, "b": 0})
//	err.WithDebugInfo("stack_trace", debug.Stack())
//
// # Error Wrapping
//
// The package supports Go 1.13+ error wrapping patterns:
//
//	originalErr := errors.New("connection refused")
//	mcpErr := WrapError(originalErr, ErrorCodeMCPConnectionFailed, "Failed to connect to server")
//
//	// Check error chain
//	if errors.Is(mcpErr, originalErr) {
//		// Handle specific error
//	}
//
//	// Extract MCP error from chain
//	if mcpErr := FindMCPError(err); mcpErr != nil {
//		log.Printf("MCP error code: %d", mcpErr.Code)
//	}
//
// # Structured Logging
//
// The package provides structured logging with automatic context extraction:
//
//	logger := NewErrorLogger(debugMode, sanitize)
//	logger.LogMCPError(ctx, mcpErr, LogLevelError, "Tool execution failed")
//
//	// Convenience methods
//	logger.Error(ctx, err, "Operation failed")
//	logger.Debug(ctx, err, "Debug information")
//
// # Error Sanitization
//
// For production environments, errors can be sanitized to remove sensitive information:
//
//	sanitized := err.Sanitize()
//	// Sensitive context keys are removed, debug info is cleared
//
//	// Logger can automatically sanitize
//	logger := NewErrorLogger(false, true) // sanitize=true
//
// # Integration with mcp-go
//
// MCP errors can be converted to mcp-go format for protocol compliance:
//
//	mcpErr := NewToolError("calculator", cause)
//	jsonrpcErr := mcpErr.ToMCPError(requestId)
//	// Send as JSON-RPC error response
//
// # Error Classification
//
// The package provides utilities for error classification:
//
//	if IsTemporary(err) {
//		// Retry the operation
//	}
//
//	if IsRetryable(err) {
//		// Safe to retry with backoff
//	}
//
//	if IsFatal(err) {
//		// Don't retry, handle gracefully
//	}
//
// # Aggregate Errors
//
// For operations that can produce multiple errors:
//
//	var errors []error
//	// ... collect errors ...
//
//	if len(errors) > 0 {
//		aggErr := NewAggregateError(errors, ErrorCodeMCPHandler, "Multiple tool failures")
//		mcpErr := aggErr.ToMCPError()
//	}
//
// # Best Practices
//
//   - Use factory functions for common error types
//   - Add relevant context information for debugging
//   - Use error wrapping to preserve error chains
//   - Sanitize errors in production environments
//   - Log errors with appropriate severity levels
//   - Check error types using errors.Is and errors.As
//   - Use structured logging for better observability
//
// # Security Considerations
//
// The package automatically identifies and sanitizes sensitive information:
//
//   - Context keys containing "password", "token", "secret", etc. are redacted
//   - Debug information is only included in debug mode
//   - Error data can be cleared during sanitization
//   - Cause chains are not included in sanitized errors
//
// This ensures that sensitive information doesn't leak through error messages
// in production environments while maintaining rich debugging capabilities
// during development.
package errors
