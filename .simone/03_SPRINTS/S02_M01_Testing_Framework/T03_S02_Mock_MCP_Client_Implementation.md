# T03_S02: Mock MCP Client Implementation

## Description
Build a comprehensive mock MCP client using the mcp-go library types and interfaces to simulate realistic protocol interactions for testing. This task extends the existing MockClient foundation in `internal/testing/mcp/client.go` to provide full protocol compliance, configurable behaviors, and advanced testing capabilities including connection state management, concurrent client simulation, and handshake protocol support.

## Objectives
- [ ] Extend existing MockClient with full MCP client interface implementation
- [ ] Add configurable response patterns and error injection capabilities
- [ ] Implement connection state management matching server behavior
- [ ] Support concurrent mock client instances for load testing
- [ ] Create builder pattern for easy mock configuration
- [ ] Add comprehensive call tracking and verification methods
- [ ] Implement handshake simulation with timeout handling

## Technical Details

### Integration with mcp-go Library
Following the established patterns in `internal/protocol/mcp/types.go` and ADR001:
- Use mcp-go types directly (no custom type definitions) as per ADR001
- Leverage `github.com/mark3labs/mcp-go/mcp` for all protocol types
- Maintain compatibility with `server.MCPServer` interfaces
- See `.simone/05_ARCHITECTURE_DECISIONS/ADR001_mcp_go_library_integration.md` for architectural rationale

### Mock Client Architecture

#### 1. Enhanced MockClient Structure
```go
type MockClient struct {
    mu sync.RWMutex
    
    // Connection management
    connectionID    string
    connectionState ConnectionState
    handshakeTimer  *time.Timer
    
    // Protocol compliance
    protocolVersion string
    serverInfo      mcp.Implementation
    capabilities    mcp.ServerCapabilities
    
    // Existing fields...
    responses        map[string]interface{}
    errors          map[string]error
    delays          map[string]time.Duration
    
    // New: Response sequences for testing flows
    responseSequences map[string][]interface{}
    sequenceIndexes   map[string]int
}
```

#### 2. Connection State Integration
Match the server's connection state flow from `internal/protocol/connection/state.go`:
- StateNew → StateInitializing → StateReady → StateClosed
- Implement handshake timeout simulation
- Support state transition validation

#### 3. Builder Pattern Implementation
```go
type MockClientBuilder struct {
    client *MockClient
}

func NewMockClientBuilder() *MockClientBuilder {
    return &MockClientBuilder{
        client: NewMockClient(),
    }
}

func (b *MockClientBuilder) WithProtocolVersion(version string) *MockClientBuilder
func (b *MockClientBuilder) WithServerInfo(name, version string) *MockClientBuilder
func (b *MockClientBuilder) WithCapabilities(caps mcp.ServerCapabilities) *MockClientBuilder
func (b *MockClientBuilder) WithHandshakeTimeout(timeout time.Duration) *MockClientBuilder
func (b *MockClientBuilder) Build() *MockClient
```

### Key Methods to Implement

#### 1. Handshake Simulation
```go
// SimulateHandshake performs a realistic handshake sequence
func (m *MockClient) SimulateHandshake(ctx context.Context) error {
    // 1. Transition to Initializing state
    // 2. Send Initialize request
    // 3. Receive InitializeResult
    // 4. Send Initialized notification
    // 5. Transition to Ready state
}

// SimulateHandshakeTimeout triggers handshake timeout
func (m *MockClient) SimulateHandshakeTimeout() error
```

#### 2. Response Sequences
```go
// SetResponseSequence configures a sequence of responses for a method
func (m *MockClient) SetResponseSequence(method string, responses []interface{})

// Example: Simulating progressive resource loading
client.SetResponseSequence("ListResources", []interface{}{
    &mcp.ListResourcesResult{Resources: []mcp.Resource{resource1}},
    &mcp.ListResourcesResult{Resources: []mcp.Resource{resource1, resource2}},
    &mcp.ListResourcesResult{Resources: []mcp.Resource{resource1, resource2, resource3}},
})
```

#### 3. Advanced Call Verification
```go
// VerifyCallSequence checks if methods were called in specific order
func (m *MockClient) VerifyCallSequence(expectedSequence []string) error

// VerifyCallWithArgs verifies a call was made with specific arguments
func (m *MockClient) VerifyCallWithArgs(method string, argMatcher func(args interface{}) bool) bool

// WaitForCall blocks until a specific method is called (with timeout)
func (m *MockClient) WaitForCall(method string, timeout time.Duration) error
```

### Testing Scenarios to Support

1. **Protocol Compliance Testing**
   - Handshake success/failure scenarios
   - Protocol version negotiation
   - Capability advertisement

2. **Error Handling**
   - Connection timeouts
   - Invalid state transitions
   - Protocol errors (JSON-RPC error codes)

3. **Concurrent Operations**
   - Multiple clients connecting simultaneously
   - Parallel request handling
   - Resource contention scenarios

4. **State Management**
   - Connection lifecycle testing
   - State transition validation
   - Cleanup verification

### Implementation Approach

1. **Phase 1: Extend Core MockClient**
   - Add connection state fields
   - Implement state transition methods
   - Add handshake simulation

2. **Phase 2: Builder Pattern**
   - Create MockClientBuilder
   - Add fluent configuration methods
   - Support preset configurations

3. **Phase 3: Advanced Features**
   - Response sequences
   - Call verification methods
   - Concurrent client support

4. **Phase 4: Testing Utilities**
   - Helper functions for common scenarios
   - Assertion helpers
   - Mock server interaction patterns

### Example Usage
```go
// Create a mock client with builder
client := NewMockClientBuilder().
    WithProtocolVersion("1.0").
    WithServerInfo("Mock Server", "1.0.0").
    WithCapabilities(mcp.ServerCapabilities{
        Tools: &mcp.ToolsCapability{ListChanged: true},
        Resources: &mcp.ResourcesCapability{Subscribe: true},
    }).
    Build()

// Simulate successful handshake
err := client.SimulateHandshake(ctx)
require.NoError(t, err)
assert.True(t, client.IsReady())

// Configure response sequence
client.SetResponseSequence("CallTool", []interface{}{
    mcp.NewToolResultText("First call"),
    mcp.NewToolResultText("Second call"),
    mcp.NewToolResultError("Third call fails"),
})

// Verify call patterns
err = client.VerifyCallSequence([]string{"Initialize", "ListTools", "CallTool"})
require.NoError(t, err)
```

## Dependencies
- Existing MockClient in `internal/testing/mcp/client.go`
- Connection state types from `internal/protocol/connection/state.go`
- MCP types from `github.com/mark3labs/mcp-go/mcp`
- Testing patterns established in T02_S02

## Success Criteria
- [ ] MockClient supports all MCP client interface methods
- [ ] Connection state management matches server implementation
- [ ] Builder pattern provides easy configuration
- [ ] Response sequences enable complex scenario testing
- [ ] Call verification methods support precise assertions
- [ ] Concurrent client operations are thread-safe
- [ ] Documentation includes comprehensive examples
- [ ] Integration tests demonstrate realistic usage

## Related Tasks
- **T02_S02**: Establishes testing patterns this task will follow
- **T04_S02**: Will use this mock client for integration tests
- **T05_S02**: Coverage metrics will validate mock usage

## Notes
- Leverage existing MockClient foundation rather than creating from scratch
- Ensure thread safety for all operations
- Follow mcp-go library conventions for type usage
- Consider future extensibility for new MCP features

## Estimated Complexity
**Medium** - Builds on existing foundation but requires careful integration with connection state management and advanced testing features. The use of mcp-go library simplifies protocol compliance.