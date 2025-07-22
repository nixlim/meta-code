# Task: Integration Test Framework

## Task Metadata
- **Task ID**: T04_S02
- **Sprint**: S02
- **Status**: open
- **Created**: 2025-07-21
- **Complexity**: High
- **Dependencies**: T03_S02 (Mock MCP Client), T09_S01 (Integration Testing)

## Description
Create a comprehensive end-to-end testing framework for testing complete protocol flows, from initial connection establishment through handshake, request/response cycles, notifications, and connection teardown. This framework should enable testing of realistic MCP server scenarios with proper lifecycle management, concurrent operations, and error conditions.

## Goal/Objectives
- Establish robust test server setup and teardown infrastructure
- Create comprehensive protocol flow testing capabilities
- Enable concurrent connection and request testing
- Implement proper test lifecycle management
- Support both synchronous and asynchronous testing patterns
- Provide utilities for complex multi-step scenarios

## Acceptance Criteria
- [ ] Test server can be started and stopped cleanly in tests
- [ ] Full handshake flow can be tested end-to-end
- [ ] Multiple concurrent connections can be tested
- [ ] Request/response flows work with real protocol messages
- [ ] Notification delivery can be verified
- [ ] Error scenarios properly propagate through the stack
- [ ] Test cleanup prevents resource leaks
- [ ] Tests can run in parallel without interference
- [ ] Performance benchmarks included for key flows

## Subtasks
- [ ] Create test server setup utilities in `test/integration/framework/`
- [ ] Implement connection lifecycle test helpers
- [ ] Build protocol flow test scenarios
- [ ] Create concurrent testing utilities
- [ ] Implement test assertion helpers for MCP
- [ ] Add performance benchmarking support
- [ ] Create test data builders for complex scenarios
- [ ] Document testing patterns and best practices

## Technical Guidance

### Test Server Architecture
```go
// test/integration/framework/server.go
type TestServer struct {
    *mcp.HandshakeServer
    transport   Transport
    connections sync.Map
    logger      *logging.Logger
}

func NewTestServer(config TestServerConfig) *TestServer
func (ts *TestServer) Start() error
func (ts *TestServer) Stop() error
func (ts *TestServer) WaitForConnection(timeout time.Duration) (*Connection, error)
```

### Key Components to Implement:

1. **Test Transport Layer**
   - In-memory transport for fast testing
   - Network transport for realistic scenarios
   - Configurable latency and error injection

2. **Connection Management**
   - Track all active connections
   - Enable inspection of connection state
   - Support graceful and forced disconnects

3. **Message Flow Testing**
   ```go
   // Example test pattern
   func TestCompleteProtocolFlow(t *testing.T) {
       server := framework.NewTestServer(config)
       defer server.Stop()
       
       client := server.NewTestClient()
       defer client.Close()
       
       // Test handshake
       require.NoError(t, client.Connect())
       require.NoError(t, client.Initialize())
       
       // Test request/response
       resp := client.CallTool("echo", params)
       assert.Equal(t, expected, resp)
       
       // Test notifications
       client.Subscribe("resourceUpdated")
       notification := client.WaitForNotification()
       assert.NotNil(t, notification)
   }
   ```

4. **Scenario-Based Testing**
   - Pre-defined scenarios for common flows
   - Composable test steps
   - State verification between steps

5. **Concurrent Testing Support**
   ```go
   func TestConcurrentConnections(t *testing.T) {
       server := framework.NewTestServer(config)
       defer server.Stop()
       
       // Launch multiple clients
       clients := framework.LaunchClients(server, 100)
       
       // Perform concurrent operations
       results := framework.ConcurrentExecute(clients, func(c *TestClient) error {
           return c.CallTool("process", data)
       })
       
       // Verify results
       framework.AssertAllSuccessful(t, results)
   }
   ```

### Integration Points:
- Leverage mock client from T03_S02
- Use existing HandshakeServer from `internal/protocol/mcp`
- Integrate with router and connection managers
- Support all transport types (stdio, HTTP/SSE, WebSocket)

### Error Scenario Testing:
1. Connection failures during handshake
2. Timeout during initialization
3. Malformed messages
4. Protocol version mismatches
5. Server shutdown during active connections
6. Network interruptions
7. Resource exhaustion

### Performance Testing:
```go
func BenchmarkProtocolFlow(b *testing.B) {
    server := framework.NewTestServer(config)
    defer server.Stop()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        client := server.NewTestClient()
        defer client.Close()
        
        for pb.Next() {
            // Measure full flow performance
            client.Connect()
            client.Initialize()
            client.CallTool("echo", params)
            client.Disconnect()
        }
    })
}
```

## Implementation Notes
1. Build on existing integration tests in `test/integration/mcp/`
2. Reuse mock client/server from `internal/testing/mcp/`
3. Consider using testcontainers for network transport tests
4. Implement proper cleanup with `t.Cleanup()` hooks
5. Use subtests for better test organization
6. Add race detection support (`go test -race`)
7. Include goroutine leak detection
8. Support debugging with verbose logging options
9. Create test fixtures for common scenarios
10. Enable test isolation for parallel execution

## Test Categories to Cover:
1. **Lifecycle Tests**: Connection, handshake, teardown
2. **Protocol Tests**: All MCP message types
3. **State Tests**: Connection state transitions
4. **Concurrency Tests**: Multiple clients, parallel requests
5. **Error Tests**: All failure modes
6. **Performance Tests**: Throughput, latency, scalability
7. **Stress Tests**: High load, resource limits
8. **Integration Tests**: Full stack with real transports

## Example Test Structure:
```
test/integration/
├── framework/
│   ├── server.go         # Test server utilities
│   ├── client.go         # Test client helpers
│   ├── transport.go      # Transport implementations
│   ├── assertions.go     # MCP-specific assertions
│   ├── scenarios.go      # Pre-built test scenarios
│   └── benchmarks.go     # Performance test utilities
├── lifecycle_test.go     # Connection lifecycle tests
├── protocol_test.go      # Protocol flow tests
├── concurrent_test.go    # Concurrency tests
├── error_test.go         # Error scenario tests
└── benchmark_test.go     # Performance benchmarks
```

## Progress Tracking
- [ ] Task started
- [ ] Test server framework designed
- [ ] Basic server lifecycle implemented
- [ ] Connection management added
- [ ] Protocol flow tests created
- [ ] Concurrent testing utilities built
- [ ] Error scenarios implemented
- [ ] Performance benchmarks added
- [ ] Documentation complete
- [ ] Code review passed
- [ ] Task completed

## Output Log
[Date/Time]: Task created

## Notes
- Focus on realistic testing scenarios that mirror production usage
- Ensure tests are deterministic and reproducible
- Consider security testing scenarios (auth, permissions)
- Plan for future transport implementations (HTTP/SSE, WebSocket)
- Create examples showing proper test patterns
- Consider integration with CI/CD pipeline (GitHub Actions)