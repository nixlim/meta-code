# Test Utilities

This package provides common testing utilities for the meta-mcp-server project.

## Available Utilities

### Connection Testing (`connection.go`)
- `CreateTestManager()` - Creates a connection manager with default timeout
- `CreateTestConnection()` - Creates a test connection with error handling
- `CreateTestContext()` - Creates a context with connection ID
- `SetupTestConnection()` - One-liner to set up manager, connection, and context
- `StartHandshakeForTest()` - Starts handshake with error handling
- `CompleteHandshakeForTest()` - Completes handshake with error handling

### MCP Testing (`mcp.go`)
- `CreateTestInitializeRequest()` - Creates initialize request with defaults
- `CreateTestInitializeResult()` - Creates initialize result with defaults
- `CreateTestCallToolRequest()` - Creates tool call request
- `CreateTestReadResourceRequest()` - Creates resource read request
- `CreateTestClientCapabilities()` - Creates client capabilities with options
- `CreateTestServerCapabilities()` - Creates server capabilities with options

### Concurrent Testing (`concurrent.go`)
- `RunConcurrentTest()` - Runs test function concurrently with WaitGroup
- `RunConcurrentTestWithDone()` - Runs concurrent tests with done channel
- `AssertNoPanic()` - Ensures function doesn't panic

## Usage Example

```go
import (
    "testing"
    "github.com/meta-mcp/meta-mcp-server/test/testutil"
)

func TestMyFeature(t *testing.T) {
    // Set up connection for testing
    manager, conn, ctx := testutil.SetupTestConnection(t, "test-conn-1")
    
    // Start handshake
    testutil.StartHandshakeForTest(t, conn)
    
    // Create test request
    request := testutil.CreateTestInitializeRequest("1.0", "Test Client")
    
    // Test your feature...
}
```

## Test Patterns

These utilities follow the patterns established in the existing test suites:
- Table-driven tests from jsonrpc package
- Simple struct-based test helpers from router package
- Concurrent testing patterns from handlers package