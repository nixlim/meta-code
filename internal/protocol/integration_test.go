// Package protocol_test provides integration tests for the MCP protocol implementation.
package protocol_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
	mcpserver "github.com/meta-mcp/meta-mcp-server/internal/protocol/mcp"
)

// TestHandshakeIntegration tests the full handshake flow with a mock client.
func TestHandshakeIntegration(t *testing.T) {
	// Create handshake-enabled server
	config := mcpserver.HandshakeConfig{
		Name:              "Test Server",
		Version:           "1.0.0",
		HandshakeTimeout:  5 * time.Second,
		SupportedVersions: []string{"1.0", "0.1.0"},
		ServerOptions: []server.ServerOption{
			mcpserver.WithToolCapabilities(true),
		},
	}

	hs := mcpserver.NewHandshakeServer(config)

	// Create a connection
	ctx := context.Background()
	connID := "test-integration-1"
	ctx, err := hs.CreateConnection(ctx, connID)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}

	// Simulate initialize request as JSON-RPC
	initRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "init-1",
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "1.0",
			"clientInfo": map[string]interface{}{
				"name":    "Test Client",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{},
		},
	}

	// Marshal request
	reqBytes, err := json.Marshal(initRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Process request through server's HandleMessage
	response := hs.HandleMessage(ctx, reqBytes)

	// Verify response
	if response == nil {
		t.Fatal("HandleMessage returned nil response")
	}

	// Convert response to JSON
	respBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Check response type
	var respMap map[string]interface{}
	if err := json.Unmarshal(respBytes, &respMap); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify it's a successful response (has result, not error)
	if _, hasResult := respMap["result"]; !hasResult {
		t.Errorf("Response missing 'result' field: %+v", respMap)
	}

	if _, hasError := respMap["error"]; hasError {
		t.Errorf("Response contains unexpected error: %+v", respMap)
	}

	// Verify connection is ready
	connManager := hs.GetConnectionManager()
	conn, ok := connManager.GetConnection(connID)
	if !ok {
		t.Fatal("Connection not found after handshake")
	}

	if !conn.IsReady() {
		t.Errorf("Connection not ready after handshake, state: %s", conn.GetState())
	}

	// Clean up
	hs.CloseConnection(connID)
}

// TestPreHandshakeRejection tests that requests before handshake are rejected.
func TestPreHandshakeRejection(t *testing.T) {
	// Create server
	config := mcpserver.DefaultHandshakeConfig()
	hs := mcpserver.NewHandshakeServer(config)

	// Create connection but don't initialize
	ctx := context.Background()
	connID := "test-rejection-1"
	ctx, err := hs.CreateConnection(ctx, connID)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}
	defer hs.CloseConnection(connID)

	// Try to call a method before handshake
	toolRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "tool-1",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "echo",
			"arguments": map[string]interface{}{
				"message": "test",
			},
		},
	}

	reqBytes, err := json.Marshal(toolRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Process request
	response := hs.HandleMessage(ctx, reqBytes)

	// Convert response to JSON
	respBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Verify it was rejected
	var respMap map[string]interface{}
	if err := json.Unmarshal(respBytes, &respMap); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should have error, not result
	if _, hasError := respMap["error"]; !hasError {
		t.Errorf("Expected error response for pre-handshake request, got: %+v", respMap)
	}

	if _, hasResult := respMap["result"]; hasResult {
		t.Errorf("Unexpected result in error response: %+v", respMap)
	}
}

// TestHandshakeTimeout tests that handshake timeout works correctly.
func TestHandshakeTimeout(t *testing.T) {
	// Create server with short timeout
	config := mcpserver.HandshakeConfig{
		Name:              "Timeout Test Server",
		Version:           "1.0.0",
		HandshakeTimeout:  100 * time.Millisecond,
		SupportedVersions: []string{"1.0"},
	}

	hs := mcpserver.NewHandshakeServer(config)

	// Create connection
	ctx := context.Background()
	connID := "test-timeout-1"
	ctx, err := hs.CreateConnection(ctx, connID)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}
	defer hs.CloseConnection(connID)

	// Get connection and start handshake manually
	connManager := hs.GetConnectionManager()
	conn, _ := connManager.GetConnection(connID)

	// Start handshake (this happens in BeforeInitialize hook normally)
	conn.StartHandshake(nil)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Verify connection is closed
	if conn.GetState() != connection.StateClosed {
		t.Errorf("Connection should be closed after timeout, but state is: %s", conn.GetState())
	}
}
