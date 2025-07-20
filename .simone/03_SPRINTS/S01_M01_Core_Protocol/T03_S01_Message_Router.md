# Task: Basic Message Router

**Task ID:** T03_S01  
**Sprint:** S01  
**Status:** open  
**Complexity:** Medium  
**Title:** Basic Message Router Implementation

## Description
Implement a basic message routing system that dispatches incoming JSON-RPC requests to appropriate handlers. This router will provide the foundation for synchronous request processing with a simple, extensible handler registration mechanism.

## Goal/Objectives
- Create a simple message router/dispatcher system
- Implement handler registration mechanism
- Support synchronous request processing
- Map methods to handlers efficiently
- Provide proper error handling for unknown methods

## Acceptance Criteria
- [ ] Router can register handlers for specific methods
- [ ] Incoming requests are dispatched to correct handlers
- [ ] Request IDs are preserved in responses
- [ ] Unknown methods return proper JSON-RPC errors
- [ ] Thread-safe handler registration
- [ ] Clean handler interface for easy implementation

## Subtasks
- [ ] Design Router interface with Register and Handle methods
- [ ] Implement handler registration with method mapping
- [ ] Create Handler interface for request processing
- [ ] Implement synchronous request dispatch
- [ ] Add error handler for unknown methods
- [ ] Create unit tests for router functionality
- [ ] Document handler implementation patterns

## Technical Guidance

### Key interfaces and integration points:
- Router in `internal/protocol/router/router.go`
- Handler interface in `internal/protocol/router/handler.go`
- Integrate with JSON-RPC types from T01
- Support MCP message types from T02

### Existing patterns to follow:
- Use sync.Map for thread-safe handler registry
- Define Handler as interface{ Handle(params) (result, error) }
- Use function types for simple handlers
- Return structured errors for proper JSON-RPC responses

## Implementation Notes
1. Start with a simple synchronous router design
2. Use method names as keys for handler lookup
3. Consider using reflect for flexible parameter handling
4. Use defer/recover for panic protection in handlers
5. Design for testability with dependency injection
6. Keep the implementation simple and focused
7. Prepare interfaces for future async extensions

## Complexity Analysis
This task has been assessed as **Medium complexity** due to:
- **Focused Scope**: Only synchronous routing, no async complexity
- **Clear Patterns**: Well-established handler registration patterns
- **Limited Dependencies**: Only requires T01 and T02 types
- **Straightforward Testing**: Easy to mock and test synchronous flows

## File Structure
```
internal/protocol/router/
├── router.go      # Core router implementation
├── handler.go     # Handler interface definition
└── registry.go    # Handler registry management
```

## Testing Requirements
- Unit tests for handler registration
- Tests for method dispatch
- Error handling tests
- Thread safety tests
- Mock handler implementations

## References
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- Go sync.Map documentation
- Handler pattern best practices