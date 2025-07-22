# T02_S02: Unit Test Suite Implementation

## Task Overview

**ID**: T02_S02  
**Sprint**: S02 (M01 - Testing Framework)  
**Priority**: 2 - High  
**Complexity**: 4 - High  
**Estimated Effort**: 6-8 hours  

## Description

Create comprehensive unit tests for existing core components, focusing on packages with lower test coverage. Priority targets are the handlers package (42.5% coverage) and mcp package (60.9% coverage) to bring them up to 80%+ coverage standard.

## Current State Assessment

### Test Coverage Status
- **handlers**: 42.5% ❌ (Highest priority - missing validation_hooks.go tests entirely)
- **mcp**: 60.9% ⚠️ (Second priority - missing critical handshake and capability tests)
- **router**: 87.4% ✅ (Already meets target)
- **jsonrpc**: 93.5% ✅ (Excellent coverage)
- **connection**: 87.0% ✅ (Already meets target)

### Existing Test Patterns
1. **Table-driven tests**: Primary pattern (see jsonrpc/jsonrpc_test.go)
2. **Mock objects**: Simple struct-based mocks (see router/router_test.go)
3. **Naming convention**: `TestPackage_FunctionName` or `TestFunctionName`
4. **Test structure**: Arrange-Act-Assert with descriptive test names

## Objectives

1. **Increase handlers package coverage to 80%+**
   - Create tests for validation_hooks.go (currently 0% coverage)
   - Complete missing test cases in initialize_hooks_test.go

2. **Increase mcp package coverage to 80%+**
   - Add missing handshake lifecycle tests
   - Test error handling and edge cases
   - Cover capability checking functions

3. **Establish reusable test utilities**
   - Mock connection manager factory
   - Test context builders
   - Common assertion helpers

4. **Document testing patterns**
   - Update test documentation
   - Add examples for common scenarios

## Technical Guidance

### Priority 1: handlers/validation_hooks.go Tests

Create comprehensive tests for all functions:
- `CreateValidationHooks()` - Test BeforeAny hook behavior
- `CreateRequestValidator()` - Test middleware validation logic
- `isNotification()` - Test notification detection
- `CreateErrorHook()` - Test error logging behavior
- `CreateSuccessHook()` - Test success logging behavior

Key test scenarios:
```go
// Test structure example
func TestHandlers_CreateValidationHooks(t *testing.T) {
    tests := []struct {
        name           string
        method         mcp.MCPMethod
        connectionState connection.State
        hasConnectionID bool
        expectAllow     bool
    }{
        {
            name:            "allows_initialize_method",
            method:          mcp.MethodInitialize,
            connectionState: connection.StateNew,
            hasConnectionID: true,
            expectAllow:     true,
        },
        {
            name:            "allows_notification",
            method:          mcp.MethodPing,
            connectionState: connection.StateNew,
            hasConnectionID: true,
            expectAllow:     true,
        },
        {
            name:            "rejects_method_when_not_ready",
            method:          mcp.MethodToolsList,
            connectionState: connection.StateInitializing,
            hasConnectionID: true,
            expectAllow:     false,
        },
        // Add more test cases
    }
    // Implementation...
}
```

### Priority 2: mcp Package Tests

Focus areas needing coverage:
1. **handshake.go**:
   - `Shutdown()` method
   - `HandleRequest()` error cases
   - `HandleNotification()` method
   - Timeout handling in handshake process

2. **types.go**:
   - `HasToolCapability()`
   - `HasResourceCapability()`
   - `HasPromptCapability()`
   - `HasLoggingCapability()`
   - `GetCapabilityConfig()`

### Test Utilities to Create

```go
// test/testutil/connection.go
func NewMockConnectionManager() *connection.Manager {
    // Create pre-configured manager for tests
}

func NewTestConnection(state connection.State) *connection.Connection {
    // Create connection in specific state
}

// test/testutil/context.go
func ContextWithConnection(conn *connection.Connection) context.Context {
    // Create context with connection
}
```

### Mock Patterns to Follow

Based on existing patterns:
```go
// Simple struct-based mocks (following router pattern)
type mockConnectionManager struct {
    connections map[string]*connection.Connection
    returnError bool
}

func (m *mockConnectionManager) GetConnection(id string) (*connection.Connection, bool) {
    if m.returnError {
        return nil, false
    }
    conn, ok := m.connections[id]
    return conn, ok
}
```

## Components Requiring Tests

### handlers Package
1. **validation_hooks.go**:
   - All hook creation functions
   - Context extraction logic
   - State validation logic
   - Error response generation

2. **initialize_hooks_test.go** (expand):
   - Error cases for initialization
   - Version negotiation failures
   - Connection state transitions

### mcp Package
1. **handshake.go**:
   - Full lifecycle testing
   - Concurrent request handling
   - Error propagation
   - Timeout scenarios

2. **types.go**:
   - All capability checking functions
   - Edge cases for capability configs
   - Nil/empty capability handling

## Success Criteria

1. **Coverage Targets Met**:
   - handlers package: ≥80% coverage
   - mcp package: ≥80% coverage

2. **Test Quality**:
   - All critical paths tested
   - Error cases covered
   - Edge cases handled
   - Concurrent scenarios tested where applicable

3. **Maintainability**:
   - Clear test names describing scenarios
   - Reusable test utilities created
   - Consistent patterns across packages

## Testing Approach

1. **Start with validation_hooks.go**: Complete 0% coverage file first
2. **Follow existing patterns**: Use table-driven tests from jsonrpc package
3. **Create minimal mocks**: Follow router's simple mock pattern
4. **Test behavior, not implementation**: Focus on observable behavior
5. **Cover error paths**: Ensure all error conditions are tested

## Notes

- Validation hooks testing requires careful mock setup for connection states
- Consider integration test scenarios for complete handshake flow
- Some async router tests are failing - be aware when running full test suite
- Use `go test -v ./internal/protocol/handlers -run TestSpecificTest` for targeted testing

## References

- Existing test examples: `internal/protocol/jsonrpc/jsonrpc_test.go`
- Mock patterns: `internal/protocol/router/router_test.go`
- Connection testing: `internal/protocol/connection/state_test.go`