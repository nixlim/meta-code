# MCP Protocol Handshake Implementation

This package implements the Initialize/Initialized handshake for the Model Context Protocol (MCP), building on top of the mcp-go library.

## Overview

The implementation consists of several key components:

1. **Connection State Management** (`connection/state.go`)
   - Tracks connection states: New → Initializing → Ready → Closed
   - Thread-safe state transitions
   - Handshake timeout enforcement
   - Connection metadata storage

2. **Initialization Hooks** (`handlers/initialize_hooks.go`)
   - `OnBeforeInitialize`: Validates protocol version, starts handshake
   - `OnAfterInitialize`: Completes handshake, stores client info
   - Protocol version negotiation logic
   - Comprehensive logging for debugging

3. **Validation Hooks** (`handlers/validation_hooks.go`)
   - `BeforeAny`: Ensures handshake completion before allowing other methods
   - Returns proper JSON-RPC errors for protocol violations
   - Allows notifications and initialize method without handshake

4. **Enhanced Server** (`mcp/handshake.go`)
   - `HandshakeServer`: Wraps mcp-go server with handshake support
   - Automatic hook registration
   - Connection lifecycle management
   - Configurable handshake timeout

## Usage

```go
// Configure the handshake-enabled server
config := mcp.HandshakeConfig{
    Name:              "My MCP Server",
    Version:           "1.0.0",
    HandshakeTimeout:  30 * time.Second,
    SupportedVersions: []string{"1.0", "0.1.0"},
    ServerOptions: []server.ServerOption{
        mcp.WithToolCapabilities(true),
        mcp.WithResourceCapabilities(true, true),
    },
}

// Create server with handshake support
server := mcp.NewHandshakeServer(config)

// Add tools and resources as usual
server.AddTool(myTool, myHandler)

// Start server with handshake support
mcp.ServeStdioWithHandshake(server)
```

## Handshake Flow

1. **Client connects** - Connection created in "New" state
2. **Client sends Initialize** - State transitions to "Initializing", timeout starts
3. **Server validates** - Protocol version negotiation occurs
4. **Server responds with Initialized** - State transitions to "Ready"
5. **Normal operations begin** - Other methods now allowed

## Key Features

- **State Management**: Finite state machine ensures proper handshake flow
- **Timeout Protection**: Configurable timeout prevents hanging connections
- **Version Negotiation**: Supports multiple protocol versions
- **Single Handshake**: Uses sync.Once to prevent multiple handshakes
- **Comprehensive Logging**: All handshake steps logged for debugging
- **Thread Safety**: All operations are thread-safe for concurrent connections

## Testing

The implementation includes comprehensive tests:
- Unit tests for connection state management
- Hook behavior tests
- Integration tests for full handshake flow
- Timeout scenario tests
- Concurrent connection tests

## Integration with mcp-go

The implementation leverages mcp-go's hook system:
- Uses `server.WithHooks()` to register custom hooks
- Implements standard hook interfaces
- Maintains compatibility with mcp-go's message handling