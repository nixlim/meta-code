# Transport Package

The transport package provides transport implementations for the MCP (Model Context Protocol) protocol. It includes support for various transport mechanisms to enable communication between MCP clients and servers.

## Overview

This package implements the transport layer for MCP, supporting:
- **STDIO Transport**: For subprocess-based MCP servers communicating via standard input/output
- **HTTP/SSE Transport**: (Planned) For network-based MCP servers using HTTP and Server-Sent Events
- **Connection Management**: Managing multiple transport connections simultaneously

## Architecture

### Core Components

1. **Transport Interface** (`jsonrpc.Transport`):
   - `Send()`: Send a single message
   - `Receive()`: Receive a single message
   - `SendBatch()`: Send multiple messages as a batch
   - `ReceiveBatch()`: Receive multiple messages as a batch
   - `Close()`: Close the transport connection
   - `IsConnected()`: Check connection status

2. **STDIO Transport** (`STDIOTransport`):
   - Implements communication with subprocess MCP servers
   - Handles process lifecycle management
   - Monitors stderr for debugging
   - Thread-safe for concurrent operations

3. **Manager** (`Manager`):
   - Manages multiple transport connections
   - Supports different transport types
   - Provides health checking and monitoring
   - Enables broadcast messaging to all connections

## Usage

### STDIO Transport

```go
// Create a subprocess command
cmd := exec.Command("mcp-server", "--stdio")

// Create STDIO transport
transport, err := transport.NewSTDIOTransport(cmd)
if err != nil {
    log.Fatal(err)
}
defer transport.Close()

// Send a message
ctx := context.Background()
request := jsonrpc.NewRequest("initialize", params, 1)
err = transport.Send(ctx, request)

// Receive a response
response, err := transport.Receive(ctx)
```

### Connection Manager

```go
// Create a manager
manager := transport.NewManager()
defer manager.Close()

// Add a connection
config := &transport.ConnectionConfig{
    Type:    transport.ConnectionTypeSTDIO,
    Command: "mcp-server",
    Args:    []string{"--stdio"},
}
err := manager.AddConnection("server1", config)

// Get a connection
transport, exists := manager.GetConnection("server1")

// Broadcast to all connections
notification := jsonrpc.NewNotification("event", data)
err = manager.Broadcast(ctx, notification)

// Health check
status := manager.HealthCheck()
```

## Features

### Process Management
- Automatic process lifecycle management
- Graceful shutdown with timeout
- Process health monitoring
- Stderr capture for debugging

### Concurrency Safety
- Thread-safe message sending and receiving
- Protected concurrent writes
- Safe connection state management

### Error Handling
- Comprehensive error reporting
- Process exit detection
- Connection state tracking
- Timeout support for all operations

### Codec Support
- JSON encoding/decoding
- Batch message support
- Message type detection
- Error message handling

## Testing

The package includes comprehensive tests:
- Unit tests for all components
- Integration tests with real subprocesses
- Concurrent operation tests
- Error handling scenarios
- Performance benchmarks

Run tests:
```bash
go test ./internal/protocol/transport/...
```

Run with coverage:
```bash
go test -cover ./internal/protocol/transport/...
```

## Future Enhancements

1. **HTTP/SSE Transport**:
   - WebSocket support
   - HTTP long polling
   - Server-Sent Events
   - TLS configuration

2. **Connection Pooling**:
   - Reusable connections
   - Connection limits
   - Load balancing

3. **Monitoring**:
   - Metrics collection
   - Performance tracking
   - Connection statistics

## Dependencies

- Standard library only (no external dependencies)
- Uses `os/exec` for subprocess management
- Uses `bufio` for efficient I/O
- Uses `sync` for concurrency control