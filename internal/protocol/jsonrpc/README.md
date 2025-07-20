# JSON-RPC 2.0 Foundation

This package provides a complete JSON-RPC 2.0 implementation for the Meta-MCP server, serving as the foundational protocol layer for the Model Context Protocol (MCP) system.

## Features

- ✅ **Complete JSON-RPC 2.0 Compliance** - Implements the full JSON-RPC 2.0 specification
- ✅ **Batch Request Support** - Handles both single and batch JSON-RPC requests
- ✅ **Enhanced Validation** - Comprehensive validation for all message types
- ✅ **Parameter Binding** - Easy parameter binding with `BindParams()` method
- ✅ **Transport Agnostic** - Clean interfaces for STDIO, HTTP, and other transports
- ✅ **Standard Error Codes** - All standard JSON-RPC error codes included
- ✅ **Go Best Practices** - Follows Go idioms and error handling patterns

## Message Types

### Request
```go
type Request struct {
    Version string `json:"jsonrpc"`
    Method  string `json:"method"`
    Params  any    `json:"params,omitempty"`
    ID      any    `json:"id,omitempty"`
}
```

### Response
```go
type Response struct {
    Version string `json:"jsonrpc"`
    Result  any    `json:"result,omitempty"`
    Error   *Error `json:"error,omitempty"`
    ID      any    `json:"id"`
}
```

### Notification
```go
type Notification struct {
    Version string `json:"jsonrpc"`
    Method  string `json:"method"`
    Params  any    `json:"params,omitempty"`
}
```

### Error
```go
type Error struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}
```

## Usage Examples

### Creating Messages

```go
// Create a request
req := jsonrpc.NewRequest("get_user", map[string]any{"id": 123}, 1)

// Create a notification
notif := jsonrpc.NewNotification("user_updated", map[string]any{"id": 123})

// Create a response
resp := jsonrpc.NewResponse(map[string]any{"name": "John"}, 1)

// Create an error response
err := jsonrpc.NewMethodNotFoundError("unknown_method")
errResp := jsonrpc.NewErrorResponse(err, 1)
```

### Parsing Messages

```go
// Parse a single message
msg, err := jsonrpc.ParseMessage([]byte(`{"jsonrpc":"2.0","method":"test","id":1}`))

// Parse batch messages
messages, err := jsonrpc.Parse([]byte(`[
    {"jsonrpc":"2.0","method":"test1","id":1},
    {"jsonrpc":"2.0","method":"test2","id":2}
]`))
```

### Parameter Binding

```go
type UserParams struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

var params UserParams
err := request.BindParams(&params)
```

## Error Codes

The package includes all standard JSON-RPC 2.0 error codes:

| Code | Constant | Description |
|------|----------|-------------|
| -32700 | `ErrorCodeParse` | Parse error |
| -32600 | `ErrorCodeInvalidRequest` | Invalid Request |
| -32601 | `ErrorCodeMethodNotFound` | Method not found |
| -32602 | `ErrorCodeInvalidParams` | Invalid params |
| -32603 | `ErrorCodeInternal` | Internal error |
| -32000 to -32099 | Server errors | Implementation-defined |

## Transport Interfaces

The package defines clean interfaces for transport abstraction:

```go
type Transport interface {
    Send(ctx context.Context, message Message) error
    Receive(ctx context.Context) (Message, error)
    SendBatch(ctx context.Context, messages []Message) error
    ReceiveBatch(ctx context.Context) ([]Message, error)
    Close() error
    IsConnected() bool
}
```

## Testing

Run the test suite:

```bash
go test ./internal/protocol/jsonrpc -v
```

Run examples:

```bash
go test ./internal/protocol/jsonrpc -run Example
```

## Architecture

This JSON-RPC foundation is designed to support the Meta-MCP server's requirements:

1. **MCP Protocol Base** - Serves as the underlying protocol for MCP communication
2. **Multi-Transport** - Supports both STDIO and HTTP/SSE transports
3. **Extensible** - Clean interfaces allow for MCP-specific extensions
4. **Performant** - Efficient parsing and validation
5. **Reliable** - Comprehensive error handling and validation

## Next Steps

This foundation enables the implementation of:

- MCP client and server components
- Transport implementations (STDIO, HTTP/SSE)
- Connection management
- Method routing and handling
- Capability negotiation

The JSON-RPC foundation is now ready to support the complete Meta-MCP server implementation.
