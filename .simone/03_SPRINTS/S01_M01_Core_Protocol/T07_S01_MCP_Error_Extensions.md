# Task: MCP Error Extensions

## Task Metadata
- **Task ID**: T07_S01
- **Sprint**: S01
- **Status**: open
- **Complexity**: Medium
- **Dependencies**: T06 (JSON-RPC Error Handling), T02 (MCP Protocol Types)

## Title
MCP Error Extensions

## Description
Extend the JSON-RPC error handling with MCP-specific error codes, error wrapping utilities, and error logging infrastructure. This task builds upon the JSON-RPC foundation to provide rich error context and debugging capabilities specific to the MCP protocol.

## Goal/Objectives
- Define MCP-specific error codes and types
- Implement error wrapping with context preservation
- Create error logging and debugging utilities
- Provide helpful error context without leaking sensitive data

## Acceptance Criteria
- [ ] MCP error codes defined in reserved range (-32000 to -32099)
- [ ] Error wrapping preserves original error chain
- [ ] Context can be added to errors without losing information
- [ ] Structured logging support for errors
- [ ] Debug mode provides additional error details
- [ ] Sensitive information is sanitized in error messages
- [ ] Error context helps with debugging

## Subtasks
- [ ] Define MCP error code constants and categories
- [ ] Create MCP-specific error types
- [ ] Implement error wrapping utilities using fmt.Errorf("%w")
- [ ] Add context attachment to errors
- [ ] Create error factory functions for MCP errors
- [ ] Implement error logging with structured fields
- [ ] Add debug data support for development mode
- [ ] Create error sanitization for production

## Technical Guidance

### Key interfaces and integration points:
- MCP errors in `internal/protocol/errors/mcp.go`
- Error wrapper in `internal/protocol/errors/wrapper.go`
- Logger interface in `internal/protocol/errors/logging.go`
- Extends JSON-RPC errors from T06
- Integrates with MCP types from T02

### Existing patterns to follow:
- Use Go 1.13+ error wrapping with %w
- Implement errors.Is and errors.As support
- Use structured logging fields
- Create error chains for context
- Follow MCP error code conventions

## Implementation Notes
1. Define MCP error categories (protocol, transport, handler)
2. Use negative numbers -32000 to -32099 for MCP errors
3. Implement Unwrap() for error chain support
4. Add fields for structured error context
5. Create WithContext() method for adding details
6. Use log levels appropriately (error, warn, debug)
7. Sanitize errors based on environment
8. Consider error aggregation for batch operations

## Complexity Analysis
This task has been assessed as **Medium complexity** due to:
- **Clear Boundaries**: Well-defined scope building on JSON-RPC foundation
- **Standard Patterns**: Uses established Go error handling patterns
- **Focused Feature Set**: Limited to MCP-specific error extensions
- **Moderate Integration**: Builds on existing error types without major complexity

## File Structure
```
internal/protocol/errors/
├── mcp.go        # MCP-specific error types
├── wrapper.go    # Error wrapping utilities
├── logging.go    # Error logging infrastructure
└── factory.go    # MCP error factory functions
```

## Critical Success Factors
1. **Developer Experience**: Easy to add context to errors
2. **Debugging Support**: Rich error information in development
3. **Production Safety**: No sensitive data leakage in errors
4. **Performance**: Minimal overhead for error handling
5. **Maintainability**: Clear error categories and patterns

## Additional Notes
This task extends the JSON-RPC error foundation with MCP-specific needs. It should focus on practical error handling improvements while maintaining the simplicity of the base error system. The implementation should make debugging easier without compromising security.