# Integration Testing Framework

This document describes the integration testing framework for the Meta-MCP Server, including the mock client, test utilities, and comprehensive test scenarios.

## Overview

The integration testing framework provides end-to-end testing of the MCP protocol implementation, covering:
- Protocol handshake flows
- Request/response message handling
- Notification delivery
- Error scenarios and recovery
- Concurrent request processing
- State transitions
- Timeout and cancellation handling

## Architecture

### Mock MCP Client (`internal/testing/mcp/client.go`)

The mock client implements the full MCPClient interface with configurable behavior:

```go
type MockClient struct {
    responses      map[string]interface{}
    notifications  []NotificationMessage
    errors         map[string]error
    delay          time.Duration
    mu             sync.RWMutex
}
```

Key features:
- Configurable responses for each method
- Error injection capabilities
- Request delay simulation
- Thread-safe operation tracking
- Full protocol compliance

### Mock Server Utilities (`internal/testing/mcp/server.go`)

Test server utilities provide:
- HTTP/WebSocket test servers
- Scenario-based test configurations
- Protocol message validation
- Connection lifecycle management

## Test Structure

### 1. Handshake Tests (`test/integration/mcp/handshake_test.go`)

Tests the complete MCP initialization handshake:
- Client sends Initialize request
- Server validates and responds with Initialized
- Connection transitions to Ready state
- Timeout protection for hanging connections

Example:
```go
func TestHandshakeIntegration(t *testing.T) {
    server := setupTestServer(t)
    client := mcp.NewMockClient()
    
    // Perform handshake
    result, err := client.Initialize(ctx, server.URL)
    require.NoError(t, err)
    require.Equal(t, "1.0.0", result.ProtocolVersion)
}
```

### 2. Client Operation Tests (`test/integration/mcp/client_test.go`)

Comprehensive testing of all client operations:
- Tool listing and execution
- Resource discovery and retrieval
- Prompt listing and execution
- Subscription management

### 3. Concurrent Request Tests (`test/integration/mcp/concurrent_test.go`)

Stress testing with parallel operations:
- Multiple concurrent tool calls
- Parallel resource fetches
- Race condition detection
- Goroutine leak verification

### 4. Error Scenario Tests (`test/integration/mcp/error_test.go`)

Error handling and recovery:
- Network failures
- Protocol violations
- Invalid responses
- Timeout scenarios
- Graceful degradation

### 5. State Transition Tests (`test/integration/mcp/state_test.go`)

Connection state management:
- New → Initializing → Ready flow
- Error state transitions
- Connection cleanup
- Reconnection scenarios

## Usage Patterns

### Basic Test Setup

```go
func setupTest(t *testing.T) (*MockClient, *httptest.Server) {
    client := NewMockClient()
    server := NewMockServer(client)
    return client, server
}
```

### Configuring Mock Responses

```go
client.SetResponse("tools/list", ListToolsResult{
    Tools: []Tool{
        {Name: "test-tool", Description: "Test tool"},
    },
})
```

### Error Injection

```go
client.SetError("tools/call", errors.New("tool execution failed"))
```

### Delay Simulation

```go
client.SetDelay(100 * time.Millisecond) // Simulate network latency
```

## Test Scenarios

### Scenario 1: Complete Workflow Test

```go
func TestCompleteWorkflow(t *testing.T) {
    // 1. Initialize connection
    // 2. List available tools
    // 3. Execute tool with parameters
    // 4. Handle notifications
    // 5. Clean shutdown
}
```

### Scenario 2: Failure Recovery

```go
func TestFailureRecovery(t *testing.T) {
    // 1. Simulate network failure
    // 2. Verify error handling
    // 3. Attempt reconnection
    // 4. Resume operations
}
```

### Scenario 3: Concurrent Load

```go
func TestConcurrentLoad(t *testing.T) {
    // 1. Launch multiple goroutines
    // 2. Execute parallel requests
    // 3. Verify ordering guarantees
    // 4. Check resource cleanup
}
```

## Running Integration Tests

```bash
# Run all integration tests
go test ./test/integration/... -v

# Run specific test suite
go test ./test/integration/mcp -run TestHandshake -v

# Run with race detection
go test ./test/integration/... -race

# Run with coverage
go test ./test/integration/... -cover

# Skip integration tests (for quick unit test runs)
go test ./... -short
```

## Best Practices

1. **Test Isolation**: Each test should be completely independent
2. **Resource Cleanup**: Always defer cleanup in test setup
3. **Timeout Protection**: Use context with timeout for all operations
4. **Error Assertions**: Check both error presence and type
5. **Goroutine Hygiene**: Verify no goroutine leaks after tests

## Debugging Failed Tests

### Enable Verbose Logging

```go
func TestWithLogging(t *testing.T) {
    if testing.Verbose() {
        // Enable debug logging
    }
}
```

### Capture Protocol Messages

```go
client.OnMessage(func(msg Message) {
    t.Logf("Message: %+v", msg)
})
```

### Check Goroutine Leaks

```go
func TestNoLeaks(t *testing.T) {
    defer goleak.VerifyNone(t)
    // Test code
}
```

## Future Enhancements

1. **Network Simulation**: Add packet loss and jitter simulation
2. **Performance Benchmarks**: Include throughput and latency tests
3. **Chaos Testing**: Random failure injection
4. **Protocol Fuzzing**: Test with malformed messages
5. **Load Testing**: High-volume concurrent operations

## Troubleshooting

### Common Issues

1. **Test Timeouts**: Increase context timeout for slow systems
2. **Port Conflicts**: Use dynamic port allocation
3. **Race Conditions**: Run with `-race` flag
4. **Flaky Tests**: Add retry logic for network operations

### Debug Commands

```bash
# Run single test with verbose output
go test -v -run TestSpecificTest ./test/integration/mcp

# Generate CPU profile
go test -cpuprofile=cpu.prof ./test/integration/...

# Analyze test coverage gaps
go test -coverprofile=coverage.out ./test/integration/...
go tool cover -html=coverage.out
```