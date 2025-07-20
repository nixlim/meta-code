# Testing Framework Documentation

This document provides comprehensive documentation for the Meta-MCP Server testing framework, including test utilities, fixtures, patterns, and best practices.

## Overview

The testing framework provides a comprehensive suite of utilities, fixtures, and patterns for testing the MCP protocol implementation. It includes:

- **Test Helpers**: Reusable utilities for common test operations
- **Test Fixtures**: Pre-defined test data for various message types
- **Mock Implementations**: Mock handlers and components for testing
- **Testing Patterns**: Established patterns for table-driven tests and benchmarks

## Directory Structure

```
internal/testing/
├── README.md           # This documentation
├── helpers/            # Test helper utilities
│   ├── helpers.go      # Core helper functions
│   ├── assertions.go   # Custom assertion helpers
│   └── setup.go        # Test setup and teardown utilities
├── fixtures/           # Test data fixtures
│   ├── jsonrpc/        # JSON-RPC message fixtures
│   ├── mcp/            # MCP protocol fixtures
│   └── errors/         # Error response fixtures
└── mocks/              # Mock implementations
    └── handlers.go     # Mock handlers for testing
```

## Test Helpers

### Core Helpers (`helpers/helpers.go`)

The core helpers provide utilities for common testing operations:

```go
// CreateTestRequest creates a test JSON-RPC request
func CreateTestRequest(method string, params interface{}, id interface{}) *jsonrpc.Request

// CreateTestNotification creates a test JSON-RPC notification
func CreateTestNotification(method string, params interface{}) *jsonrpc.Notification

// CreateTestResponse creates a test JSON-RPC response
func CreateTestResponse(result interface{}, id interface{}) *jsonrpc.Response

// CreateTestErrorResponse creates a test JSON-RPC error response
func CreateTestErrorResponse(code int, message string, id interface{}) *jsonrpc.Response
```

### Assertions (`helpers/assertions.go`)

Custom assertion helpers for common test validations:

```go
// AssertValidJSONRPC validates that a message conforms to JSON-RPC 2.0
func AssertValidJSONRPC(t *testing.T, message interface{})

// AssertErrorCode validates that an error response has the expected code
func AssertErrorCode(t *testing.T, response *jsonrpc.Response, expectedCode int)

// AssertResponseID validates that a response has the expected ID
func AssertResponseID(t *testing.T, response *jsonrpc.Response, expectedID interface{})
```

### Setup Utilities (`helpers/setup.go`)

Test setup and teardown utilities:

```go
// SetupTestRouter creates a router with common test handlers
func SetupTestRouter() *router.Router

// SetupTestAsyncRouter creates an async router for testing
func SetupTestAsyncRouter(workers int, queueSize int) *router.AsyncRouter

// CleanupTestRouter properly shuts down test routers
func CleanupTestRouter(ar *router.AsyncRouter)
```

## Test Fixtures

### JSON-RPC Fixtures (`fixtures/jsonrpc/`)

Pre-defined JSON-RPC messages for testing:

- `valid_request.json` - Valid JSON-RPC 2.0 request
- `valid_notification.json` - Valid JSON-RPC 2.0 notification
- `valid_response.json` - Valid JSON-RPC 2.0 response
- `error_response.json` - JSON-RPC 2.0 error response
- `invalid_version.json` - Invalid version for negative testing

### MCP Protocol Fixtures (`fixtures/mcp/`)

MCP-specific protocol messages:

- `initialize_request.json` - MCP initialize request
- `initialize_response.json` - MCP initialize response
- `initialized_notification.json` - MCP initialized notification
- `resources_list_request.json` - Resources list request
- `tools_list_request.json` - Tools list request

### Error Fixtures (`fixtures/errors/`)

Common error responses for testing error handling:

- `method_not_found.json` - Method not found error
- `invalid_params.json` - Invalid parameters error
- `internal_error.json` - Internal server error
- `parse_error.json` - JSON parse error

## Mock Implementations

### Mock Handlers (`mocks/handlers.go`)

Mock implementations for testing:

```go
// MockHandler provides a configurable mock handler
type MockHandler struct {
    Response *jsonrpc.Response
    Error    error
    Delay    time.Duration
}

// MockNotificationHandler provides a mock notification handler
type MockNotificationHandler struct {
    Called       bool
    LastMethod   string
    LastParams   interface{}
}
```

## Testing Patterns

### Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestMessageValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        errCode int
    }{
        {
            name:    "valid request",
            input:   `{"jsonrpc":"2.0","method":"test","id":1}`,
            wantErr: false,
        },
        {
            name:    "invalid version",
            input:   `{"jsonrpc":"1.0","method":"test","id":1}`,
            wantErr: true,
            errCode: jsonrpc.ErrorCodeInvalidRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Benchmark Tests

Performance benchmarks for critical paths:

```go
func BenchmarkRouterHandle(b *testing.B) {
    router := helpers.SetupTestRouter()
    request := helpers.CreateTestRequest("test.method", nil, 1)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        response := router.Handle(context.Background(), request)
        if response.Error != nil {
            b.Fatal("Unexpected error")
        }
    }
}
```

### Concurrent Tests

Testing concurrent access and race conditions:

```go
func TestConcurrentAccess(t *testing.T) {
    router := helpers.SetupTestRouter()
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            // Concurrent test logic
        }(i)
    }
    
    wg.Wait()
}
```

## Best Practices

### Test Organization

1. **Group Related Tests**: Use subtests to group related test cases
2. **Clear Test Names**: Use descriptive names that explain what is being tested
3. **Isolated Tests**: Ensure tests don't depend on each other
4. **Cleanup**: Always clean up resources in tests

### Error Testing

1. **Test Error Paths**: Include negative test cases for error conditions
2. **Validate Error Codes**: Check that errors have the correct JSON-RPC error codes
3. **Test Edge Cases**: Include boundary conditions and edge cases

### Performance Testing

1. **Benchmark Critical Paths**: Focus on performance-critical code
2. **Concurrent Benchmarks**: Test concurrent access patterns
3. **Memory Benchmarks**: Use `b.ReportAllocs()` to track allocations
4. **Realistic Data**: Use realistic data sizes in benchmarks

### Coverage Guidelines

- **Target**: Aim for 70%+ test coverage
- **Focus**: Prioritize critical paths and error handling
- **Quality**: Coverage percentage is less important than test quality
- **Documentation**: Document any intentionally untested code

## Running Tests

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkRouterHandle ./internal/protocol/router

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./...
```

### Race Detection

```bash
# Run tests with race detection
go test -race ./...
```

## Integration with CI/CD

The testing framework integrates with continuous integration:

1. **Automated Testing**: All tests run on every commit
2. **Coverage Reporting**: Coverage reports are generated and tracked
3. **Performance Monitoring**: Benchmark results are tracked over time
4. **Quality Gates**: Tests must pass before merging

## Contributing

When adding new tests:

1. Follow the established patterns and conventions
2. Add appropriate fixtures for new message types
3. Include both positive and negative test cases
4. Add benchmarks for performance-critical code
5. Update this documentation for new utilities or patterns
