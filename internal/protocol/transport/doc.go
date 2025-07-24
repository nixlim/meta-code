// Package transport provides transport implementations for the MCP protocol.
// It includes various transport mechanisms such as STDIO for subprocess communication
// and HTTP/SSE for network communication.
//
// The transport layer is responsible for:
//   - Sending and receiving MCP messages
//   - Managing connection lifecycle
//   - Handling transport-specific encoding/decoding
//   - Error handling and recovery
//
// Example usage:
//
//	// Create a new STDIO transport
//	transport, err := NewSTDIOTransport(cmd)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer transport.Close()
//
//	// Send a message
//	err = transport.Send(ctx, message)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Receive a message
//	msg, err := transport.Receive(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
package transport