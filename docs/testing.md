# Meta-MCP Server Testing Guide

This comprehensive guide covers all aspects of testing the Meta-MCP Server, including testing patterns, best practices, tools, and strategies for achieving high-quality test coverage.

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test Organization](#test-organization)
3. [Testing Patterns](#testing-patterns)
4. [Mock Usage Guidelines](#mock-usage-guidelines)
5. [Fixture Organization](#fixture-organization)
6. [Integration Test Patterns](#integration-test-patterns)
7. [Benchmarking Guidelines](#benchmarking-guidelines)
8. [Coverage Targets and Strategies](#coverage-targets-and-strategies)
9. [Running Tests](#running-tests)
10. [Debugging Tests](#debugging-tests)
11. [CI/CD Integration](#cicd-integration)

## Testing Philosophy

Our testing approach follows these core principles:

1. **Test Behavior, Not Implementation**: Focus on what the code does, not how it does it
2. **Comprehensive Coverage**: Aim for high coverage while prioritizing critical paths
3. **Fast and Reliable**: Tests should run quickly and produce consistent results
4. **Clear and Maintainable**: Tests serve as documentation and should be easy to understand
5. **Isolated and Independent**: Each test should run in isolation without dependencies

## Test Organization

### Directory Structure

```
meta-mcp-server/
├── internal/
│   ├── protocol/           # Protocol implementation
│   │   ├── *_test.go      # Unit tests alongside implementation
│   │   └── testdata/      # Test data specific to protocol
│   └── testing/           # Shared testing utilities
│       ├── helpers/       # Test helper functions
│       ├── fixtures/      # Shared test fixtures
│       └── mocks/         # Mock implementations
├── test/
│   ├── integration/       # Integration tests
│   ├── conformance/       # MCP conformance tests
│   └── e2e/              # End-to-end tests
└── docs/
    └── testing.md        # This file
```

### Test Types

1. **Unit Tests** (`*_test.go`): Test individual functions and methods
2. **Integration Tests** (`test/integration/`): Test component interactions
3. **Conformance Tests** (`test/conformance/`): Verify MCP protocol compliance
4. **End-to-End Tests** (`test/e2e/`): Test complete workflows
5. **Benchmark Tests** (`*_test.go` with `Benchmark*`): Performance testing

## Testing Patterns

### Table-Driven Tests

Table-driven tests are the preferred pattern for comprehensive test coverage:

```go
func TestJSONRPCValidation(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        wantErr     bool
        errCode     int
        description string
    }{
        {
            name:        "valid_request",
            input:       `{"jsonrpc":"2.0","method":"test","id":1}`,
            wantErr:     false,
            description: "Standard JSON-RPC 2.0 request should validate",
        },
        {
            name:        "missing_version",
            input:       `{"method":"test","id":1}`,
            wantErr:     true,
            errCode:     jsonrpc.ErrorCodeInvalidRequest,
            description: "Request without jsonrpc field should fail",
        },
        {
            name:        "invalid_version",
            input:       `{"jsonrpc":"1.0","method":"test","id":1}`,
            wantErr:     true,
            errCode:     jsonrpc.ErrorCodeInvalidRequest,
            description: "Non-2.0 version should be rejected",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var msg jsonrpc.Message
            err := json.Unmarshal([]byte(tt.input), &msg)
            
            if tt.wantErr {
                require.Error(t, err, tt.description)
                if tt.errCode != 0 {
                    var rpcErr *jsonrpc.Error
                    require.ErrorAs(t, err, &rpcErr)
                    assert.Equal(t, tt.errCode, rpcErr.Code)
                }
            } else {
                require.NoError(t, err, tt.description)
            }
        })
    }
}
```

### Subtests for Complex Scenarios

Use subtests to organize related test cases:

```go
func TestRouterHandling(t *testing.T) {
    router := setupTestRouter(t)
    
    t.Run("request_handling", func(t *testing.T) {
        t.Run("successful_request", func(t *testing.T) {
            // Test successful request handling
        })
        
        t.Run("method_not_found", func(t *testing.T) {
            // Test unknown method handling
        })
        
        t.Run("handler_error", func(t *testing.T) {
            // Test handler error propagation
        })
    })
    
    t.Run("notification_handling", func(t *testing.T) {
        // Test notification-specific behavior
    })
}
```

### Setup and Teardown Patterns

Use proper setup and cleanup:

```go
func TestWithCleanup(t *testing.T) {
    // Setup
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    server := httptest.NewServer(handler)
    defer server.Close()
    
    client := NewTestClient(server.URL)
    defer client.Close()
    
    // Test logic here
}
```

### Testing Concurrent Code

Ensure thread safety with concurrent tests:

```go
func TestConcurrentAccess(t *testing.T) {
    manager := NewConnectionManager()
    
    const numGoroutines = 100
    var wg sync.WaitGroup
    wg.Add(numGoroutines)
    
    errors := make(chan error, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        go func(id int) {
            defer wg.Done()
            
            conn := &MockConnection{ID: fmt.Sprintf("conn-%d", id)}
            if err := manager.Add(conn); err != nil {
                errors <- err
                return
            }
            
            // Simulate work
            time.Sleep(time.Millisecond)
            
            if err := manager.Remove(conn.ID); err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for any errors
    for err := range errors {
        t.Errorf("Concurrent operation failed: %v", err)
    }
}
```

## Mock Usage Guidelines

### When to Use Mocks

1. **External Dependencies**: Mock external services, databases, or APIs
2. **Complex Interfaces**: Mock complex interfaces to test in isolation
3. **Error Injection**: Use mocks to simulate error conditions
4. **Performance**: Replace slow operations with fast mocks

### Mock Implementation Pattern

```go
// Define the interface
type MessageHandler interface {
    Handle(ctx context.Context, msg *jsonrpc.Request) (*jsonrpc.Response, error)
}

// Create a mock implementation
type MockHandler struct {
    HandleFunc func(ctx context.Context, msg *jsonrpc.Request) (*jsonrpc.Response, error)
    calls      []HandleCall
    mu         sync.Mutex
}

type HandleCall struct {
    Ctx context.Context
    Msg *jsonrpc.Request
}

func (m *MockHandler) Handle(ctx context.Context, msg *jsonrpc.Request) (*jsonrpc.Response, error) {
    m.mu.Lock()
    m.calls = append(m.calls, HandleCall{Ctx: ctx, Msg: msg})
    m.mu.Unlock()
    
    if m.HandleFunc != nil {
        return m.HandleFunc(ctx, msg)
    }
    return nil, errors.New("HandleFunc not set")
}

func (m *MockHandler) CallCount() int {
    m.mu.Lock()
    defer m.mu.Unlock()
    return len(m.calls)
}
```

### Using Mocks in Tests

```go
func TestClientWithMockHandler(t *testing.T) {
    mockHandler := &MockHandler{
        HandleFunc: func(ctx context.Context, msg *jsonrpc.Request) (*jsonrpc.Response, error) {
            // Return predetermined response
            return &jsonrpc.Response{
                ID:     msg.ID,
                Result: json.RawMessage(`{"status":"ok"}`),
            }, nil
        },
    }
    
    client := NewClient(mockHandler)
    result, err := client.Call(context.Background(), "test.method", nil)
    
    require.NoError(t, err)
    assert.Equal(t, 1, mockHandler.CallCount())
    assert.Equal(t, `{"status":"ok"}`, string(result))
}
```

## Fixture Organization

### Fixture Directory Structure

```
internal/testing/fixtures/
├── jsonrpc/
│   ├── requests/
│   │   ├── valid_request.json
│   │   ├── invalid_version.json
│   │   └── batch_request.json
│   ├── responses/
│   │   ├── success_response.json
│   │   ├── error_response.json
│   │   └── batch_response.json
│   └── notifications/
│       └── valid_notification.json
├── mcp/
│   ├── initialize/
│   │   ├── request.json
│   │   └── response.json
│   ├── tools/
│   │   ├── list_request.json
│   │   └── call_request.json
│   └── resources/
│       └── list_response.json
└── errors/
    ├── parse_error.json
    ├── invalid_request.json
    └── method_not_found.json
```

### Loading Fixtures

```go
// Helper function to load fixtures
func LoadFixture(t *testing.T, path string) []byte {
    t.Helper()
    
    data, err := os.ReadFile(filepath.Join("testdata", path))
    require.NoError(t, err, "Failed to load fixture: %s", path)
    
    return data
}

// Type-specific fixture loaders
func LoadRequestFixture(t *testing.T, name string) *jsonrpc.Request {
    t.Helper()
    
    data := LoadFixture(t, filepath.Join("jsonrpc/requests", name))
    
    var req jsonrpc.Request
    err := json.Unmarshal(data, &req)
    require.NoError(t, err, "Failed to unmarshal request fixture: %s", name)
    
    return &req
}

// Usage in tests
func TestWithFixture(t *testing.T) {
    req := LoadRequestFixture(t, "valid_request.json")
    
    // Test with loaded fixture
    resp := handler.Handle(context.Background(), req)
    assert.NotNil(t, resp)
}
```

### Fixture Validation

Ensure fixtures remain valid:

```go
func TestFixtureValidity(t *testing.T) {
    // Validate all JSON-RPC fixtures
    fixtures, err := filepath.Glob("testdata/jsonrpc/**/*.json")
    require.NoError(t, err)
    
    for _, fixture := range fixtures {
        t.Run(fixture, func(t *testing.T) {
            data, err := os.ReadFile(fixture)
            require.NoError(t, err)
            
            var msg jsonrpc.Message
            err = json.Unmarshal(data, &msg)
            require.NoError(t, err, "Invalid JSON-RPC message in fixture")
        })
    }
}
```

## Integration Test Patterns

### Test Server Setup

```go
func setupTestServer(t *testing.T) *TestServer {
    t.Helper()
    
    // Create server with test configuration
    server := &TestServer{
        handlers: make(map[string]HandlerFunc),
        logger:   testLogger(t),
    }
    
    // Register test handlers
    server.Register("test.echo", echoHandler)
    server.Register("test.error", errorHandler)
    
    // Start HTTP server
    httpServer := httptest.NewServer(server)
    t.Cleanup(func() {
        httpServer.Close()
    })
    
    server.URL = httpServer.URL
    return server
}
```

### Integration Test Structure

```go
func TestCompleteWorkflow(t *testing.T) {
    // Setup
    server := setupTestServer(t)
    client := NewClient(server.URL)
    defer client.Close()
    
    ctx := context.Background()
    
    // Test workflow
    t.Run("initialize", func(t *testing.T) {
        result, err := client.Initialize(ctx, InitializeParams{
            ProtocolVersion: "1.0.0",
            ClientInfo: ClientInfo{
                Name:    "test-client",
                Version: "1.0.0",
            },
        })
        require.NoError(t, err)
        assert.Equal(t, "1.0.0", result.ProtocolVersion)
    })
    
    t.Run("list_tools", func(t *testing.T) {
        tools, err := client.ListTools(ctx)
        require.NoError(t, err)
        assert.NotEmpty(t, tools)
    })
    
    t.Run("call_tool", func(t *testing.T) {
        result, err := client.CallTool(ctx, "test.echo", map[string]interface{}{
            "message": "hello",
        })
        require.NoError(t, err)
        assert.Equal(t, "hello", result["message"])
    })
}
```

### Testing Error Scenarios

```go
func TestErrorHandling(t *testing.T) {
    server := setupTestServer(t)
    client := NewClient(server.URL)
    defer client.Close()
    
    tests := []struct {
        name      string
        setup     func()
        operation func() error
        wantErr   string
    }{
        {
            name: "network_timeout",
            setup: func() {
                client.SetTimeout(1 * time.Millisecond)
                server.SetDelay(100 * time.Millisecond)
            },
            operation: func() error {
                _, err := client.ListTools(context.Background())
                return err
            },
            wantErr: "context deadline exceeded",
        },
        {
            name: "server_error",
            setup: func() {
                server.ForceError(errors.New("internal error"))
            },
            operation: func() error {
                _, err := client.ListTools(context.Background())
                return err
            },
            wantErr: "internal error",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.setup != nil {
                tt.setup()
            }
            
            err := tt.operation()
            require.Error(t, err)
            assert.Contains(t, err.Error(), tt.wantErr)
        })
    }
}
```

## Benchmarking Guidelines

### Basic Benchmark Structure

```go
func BenchmarkJSONRPCParsing(b *testing.B) {
    message := []byte(`{"jsonrpc":"2.0","method":"test","params":{"key":"value"},"id":1}`)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var msg jsonrpc.Message
        if err := json.Unmarshal(message, &msg); err != nil {
            b.Fatal(err)
        }
    }
}
```

### Benchmark Variations

```go
func BenchmarkRouterPerformance(b *testing.B) {
    // Test different message sizes
    sizes := []int{100, 1000, 10000}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
            router := setupBenchmarkRouter()
            msg := generateMessage(size)
            
            b.ResetTimer()
            b.ReportAllocs()
            
            for i := 0; i < b.N; i++ {
                _ = router.Handle(context.Background(), msg)
            }
        })
    }
}
```

### Concurrent Benchmarks

```go
func BenchmarkConcurrentRequests(b *testing.B) {
    router := setupBenchmarkRouter()
    
    b.RunParallel(func(pb *testing.PB) {
        msg := createBenchmarkMessage()
        ctx := context.Background()
        
        for pb.Next() {
            _ = router.Handle(ctx, msg)
        }
    })
}
```

### Memory Allocation Benchmarks

```go
func BenchmarkMemoryAllocation(b *testing.B) {
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        // Operation to benchmark
        msg := &jsonrpc.Request{
            Version: "2.0",
            Method:  "test",
            ID:      json.RawMessage(`1`),
        }
        
        // Force heap allocation
        _ = msg.String()
    }
}
```

## Coverage Targets and Strategies

### Coverage Goals

Based on CLAUDE.md, our coverage targets are:

| Package | Target | Priority |
|---------|--------|----------|
| `jsonrpc` | 93.3% | Critical |
| `connection` | 87.0% | Critical |
| `router` | 82.1% | High |
| `mcp` | 78.5% | High |
| `errors` | 75.0% | Medium |
| Overall | 70%+ | Required |

### Coverage Strategy

1. **Focus on Critical Paths**: Prioritize core functionality
2. **Error Handling**: Ensure all error paths are tested
3. **Edge Cases**: Test boundary conditions
4. **Integration Points**: Test component interactions

### Measuring Coverage

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# View coverage by function
go tool cover -func=coverage.out

# Check specific package coverage
go test -cover ./internal/protocol/jsonrpc
```

### Improving Coverage

```go
// Use build tags to exclude hard-to-test code
// +build !coverage

func difficultToTestFunction() {
    // Code that's hard to test in unit tests
}
```

```go
// Alternative implementation for tests
// +build coverage

func difficultToTestFunction() {
    // Simplified version for testing
}
```

### Coverage Reports

Generate detailed coverage reports:

```makefile
coverage-report:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep -E '^total:|jsonrpc|connection|router|mcp|errors'
	@echo ""
	@echo "Detailed HTML report: coverage.html"
	@go tool cover -html=coverage.out -o coverage.html
```

## Running Tests

### Command Line Options

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/protocol/jsonrpc

# Run specific test
go test -run TestJSONRPCValidation ./internal/protocol/jsonrpc

# Run tests with race detection
go test -race ./...

# Run short tests only
go test -short ./...

# Run tests with timeout
go test -timeout 30s ./...

# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkJSONRPCParsing ./...

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./...

# Run tests in parallel
go test -parallel 4 ./...
```

### Using the Makefile

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Run all checks (lint, test, etc.)
make check

# Run CI pipeline
make ci
```

### Test Tags

Use build tags to control test execution:

```go
// +build integration

func TestIntegration(t *testing.T) {
    // Integration test that requires external resources
}
```

```bash
# Run integration tests
go test -tags=integration ./...

# Skip integration tests
go test -tags=!integration ./...
```

## Debugging Tests

### Verbose Output

```go
func TestWithLogging(t *testing.T) {
    if testing.Verbose() {
        // Enable debug logging
        logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
        logger.Printf("Starting test with verbose logging")
    }
    
    // Test logic
}
```

### Using t.Log for Debug Information

```go
func TestWithDebugInfo(t *testing.T) {
    t.Log("Starting test")
    
    result, err := performOperation()
    t.Logf("Operation result: %+v, error: %v", result, err)
    
    require.NoError(t, err)
}
```

### Debugging Flaky Tests

```go
func TestFlakyOperation(t *testing.T) {
    const maxRetries = 3
    
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        if i > 0 {
            t.Logf("Retry %d/%d after error: %v", i, maxRetries-1, lastErr)
            time.Sleep(100 * time.Millisecond)
        }
        
        err := flakyOperation()
        if err == nil {
            return // Success
        }
        lastErr = err
    }
    
    t.Fatalf("Operation failed after %d retries: %v", maxRetries, lastErr)
}
```

### Capturing Goroutine Dumps

```go
func TestGoroutineLeaks(t *testing.T) {
    before := runtime.NumGoroutine()
    
    // Run test
    runTest()
    
    // Give goroutines time to clean up
    time.Sleep(100 * time.Millisecond)
    
    after := runtime.NumGoroutine()
    if after > before {
        buf := make([]byte, 1<<20)
        runtime.Stack(buf, true)
        t.Fatalf("Goroutine leak detected: %d -> %d\n%s", before, after, buf)
    }
}
```

### Using Delve Debugger

```bash
# Debug a specific test
dlv test ./internal/protocol/jsonrpc -- -test.run TestJSONRPCValidation

# Set breakpoints and inspect
(dlv) break jsonrpc.go:42
(dlv) continue
(dlv) print msg
(dlv) stack
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run linters
      run: make lint
    
    - name: Run tests
      run: make test-coverage
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
    
    - name: Run benchmarks
      run: make bench-compare
```

### Pre-commit Hooks

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Run tests
echo "Running tests..."
make test-short || exit 1

# Run linters
echo "Running linters..."
make lint || exit 1

# Check formatting
echo "Checking formatting..."
make fmt-check || exit 1

echo "All checks passed!"
```

### Test Result Reporting

```makefile
# Generate JUnit XML for CI systems
test-ci:
	go install github.com/jstemmer/go-junit-report/v2@latest
	go test -v ./... 2>&1 | go-junit-report -set-exit-code > test-report.xml
```

## Best Practices Summary

1. **Write Tests First**: TDD helps design better APIs
2. **Keep Tests Simple**: Each test should verify one behavior
3. **Use Descriptive Names**: Test names should explain what they test
4. **Avoid Test Interdependence**: Tests should not rely on execution order
5. **Mock External Dependencies**: Keep tests fast and reliable
6. **Test Edge Cases**: Don't just test the happy path
7. **Maintain Test Code**: Refactor tests as you refactor code
8. **Document Complex Tests**: Add comments explaining non-obvious test logic
9. **Use Helper Functions**: DRY principle applies to tests too
10. **Review Test Failures**: Failed tests should clearly indicate what went wrong

## Conclusion

This testing guide provides a comprehensive framework for testing the Meta-MCP Server. By following these patterns and practices, we ensure our code is reliable, maintainable, and meets the high-quality standards expected of MCP protocol implementations.

Remember: Tests are not just about catching bugs—they're about building confidence in our code and enabling fearless refactoring. Write tests that you'll thank yourself for later!