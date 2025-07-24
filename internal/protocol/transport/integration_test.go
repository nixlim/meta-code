package transport_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/transport"
)

// TestSTDIOTransportIntegration tests STDIO transport with a real subprocess
func TestSTDIOTransportIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a test MCP server subprocess
	serverScript := createMockMCPServer(t)
	defer os.Remove(serverScript)

	cmd := exec.Command("python3", serverScript)
	
	transport, err := transport.NewSTDIOTransport(cmd)
	if err != nil {
		t.Fatalf("Failed to create STDIO transport: %v", err)
	}
	defer transport.Close()

	ctx := context.Background()

	// Test 1: Send initialize request
	initRequest := &jsonrpc.Request{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
		Params: json.RawMessage(`{
			"protocolVersion": "2024-11-05",
			"capabilities": {
				"tools": {},
				"prompts": {}
			},
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			}
		}`),
	}

	err = transport.Send(ctx, initRequest)
	if err != nil {
		t.Fatalf("Failed to send initialize request: %v", err)
	}

	// Receive response
	msg, err := transport.Receive(ctx)
	if err != nil {
		t.Fatalf("Failed to receive response: %v", err)
	}

	response, ok := msg.(*jsonrpc.Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", msg)
	}

	if response.Error != nil {
		t.Fatalf("Initialize returned error: %v", response.Error)
	}

	// Test 2: Send a notification
	notification := &jsonrpc.Notification{
		Version: "2.0",
		Method:  "initialized",
	}

	err = transport.Send(ctx, notification)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	// Test 3: Send a custom method
	customRequest := &jsonrpc.Request{
		Version: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "echo",
		Params:  json.RawMessage(`{"message": "Hello, MCP!"}`),
	}

	err = transport.Send(ctx, customRequest)
	if err != nil {
		t.Fatalf("Failed to send custom request: %v", err)
	}

	// Receive echo response
	msg, err = transport.Receive(ctx)
	if err != nil {
		t.Fatalf("Failed to receive echo response: %v", err)
	}

	response, ok = msg.(*jsonrpc.Response)
	if !ok {
		t.Fatalf("Expected Response for echo, got %T", msg)
	}

	// Verify echo response contains our message
	if response.Result != nil {
		// Marshal the result to JSON first
		resultBytes, err := json.Marshal(response.Result)
		if err == nil {
			var result map[string]interface{}
			if err := json.Unmarshal(resultBytes, &result); err == nil {
				if result["message"] != "Hello, MCP!" {
					t.Errorf("Echo response mismatch: %v", result)
				}
			}
		}
	}
}

// TestSTDIOTransportBatchIntegration tests batch operations with a subprocess
func TestSTDIOTransportBatchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a test MCP server subprocess that supports batch
	serverScript := createMockMCPServer(t)
	defer os.Remove(serverScript)

	cmd := exec.Command("python3", serverScript)
	
	transport, err := transport.NewSTDIOTransport(cmd)
	if err != nil {
		t.Fatalf("Failed to create STDIO transport: %v", err)
	}
	defer transport.Close()

	ctx := context.Background()

	// Send batch of requests
	batch := []jsonrpc.Message{
		&jsonrpc.Request{
			Version: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "echo",
			Params:  json.RawMessage(`{"message": "First"}`),
		},
		&jsonrpc.Request{
			Version: "2.0",
			ID:      json.RawMessage(`2`),
			Method:  "echo",
			Params:  json.RawMessage(`{"message": "Second"}`),
		},
		&jsonrpc.Notification{
			Version: "2.0",
			Method:  "log",
			Params:  json.RawMessage(`{"level": "info", "message": "Batch test"}`),
		},
	}

	err = transport.SendBatch(ctx, batch)
	if err != nil {
		t.Fatalf("Failed to send batch: %v", err)
	}

	// Receive batch response
	responses, err := transport.ReceiveBatch(ctx)
	if err != nil {
		t.Fatalf("Failed to receive batch response: %v", err)
	}

	// Should get 2 responses (notifications don't get responses)
	if len(responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(responses))
	}
}

// TestSTDIOTransportErrorHandling tests error scenarios
func TestSTDIOTransportErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test 1: Subprocess that exits immediately
	cmd := exec.Command("sh", "-c", "exit 1")
	
	transport, err := transport.NewSTDIOTransport(cmd)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}
	defer transport.Close()

	// Wait for process to exit
	time.Sleep(100 * time.Millisecond)

	// Should not be connected
	if transport.IsConnected() {
		t.Error("Transport should detect process exit")
	}

	// Operations should fail
	ctx := context.Background()
	err = transport.Send(ctx, &jsonrpc.Request{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "test",
	})
	if err == nil {
		t.Error("Send should fail on disconnected transport")
	}
}

// createMockMCPServer creates a simple Python MCP server for testing
func createMockMCPServer(t *testing.T) string {
	t.Helper()

	script := `#!/usr/bin/env python3
import json
import sys

def read_message():
    """Read a JSON-RPC message from stdin."""
    line = sys.stdin.readline()
    if not line:
        return None
    return json.loads(line)

def write_message(msg):
    """Write a JSON-RPC message to stdout."""
    json.dump(msg, sys.stdout)
    sys.stdout.write('\n')
    sys.stdout.flush()

def handle_request(request):
    """Handle a JSON-RPC request."""
    method = request.get('method')
    params = request.get('params', {})
    req_id = request.get('id')
    
    if method == 'initialize':
        return {
            'jsonrpc': '2.0',
            'id': req_id,
            'result': {
                'protocolVersion': '2024-11-05',
                'capabilities': {
                    'tools': {},
                    'prompts': {}
                },
                'serverInfo': {
                    'name': 'mock-mcp-server',
                    'version': '1.0.0'
                }
            }
        }
    elif method == 'echo':
        return {
            'jsonrpc': '2.0',
            'id': req_id,
            'result': params
        }
    elif method == 'initialized':
        # Notification, no response
        return None
    elif method == 'log':
        # Notification, no response
        sys.stderr.write(f"Log: {params}\n")
        return None
    else:
        return {
            'jsonrpc': '2.0',
            'id': req_id,
            'error': {
                'code': -32601,
                'message': 'Method not found'
            }
        }

def main():
    """Main server loop."""
    while True:
        try:
            msg = read_message()
            if msg is None:
                break
                
            # Handle batch requests
            if isinstance(msg, list):
                responses = []
                for req in msg:
                    resp = handle_request(req)
                    if resp is not None:
                        responses.append(resp)
                if responses:
                    write_message(responses)
            else:
                # Single request
                response = handle_request(msg)
                if response is not None:
                    write_message(response)
                    
        except Exception as e:
            sys.stderr.write(f"Error: {e}\n")
            break

if __name__ == '__main__':
    main()
`

	file, err := os.CreateTemp("", "mock_mcp_server_*.py")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := file.WriteString(script); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	if err := file.Chmod(0755); err != nil {
		t.Fatalf("Failed to chmod script: %v", err)
	}

	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	return file.Name()
}