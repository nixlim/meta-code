package mcp_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpmock "github.com/meta-mcp/meta-mcp-server/internal/testing/mcp"
	"github.com/meta-mcp/meta-mcp-server/test/testutil"
)

// TestConcurrentClientRequests_Refactored demonstrates using standardized utilities
func TestConcurrentClientRequests_Refactored(t *testing.T) {
	client := mcpmock.NewMockClient()
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Error closing client: %v", err)
		}
	}()

	// Configure responses
	client.SetResponse("ListTools", &mcp.ListToolsResult{
		Tools: []mcp.Tool{{Name: "test", Description: "Test tool"}},
	})
	client.SetResponse("ListResources", &mcp.ListResourcesResult{
		Resources: []mcp.Resource{{URI: "test://resource", Name: "Test"}},
	})
	client.SetResponse("ListPrompts", &mcp.ListPromptsResult{
		Prompts: []mcp.Prompt{{Name: "test", Description: "Test prompt"}},
	})

	// Number of concurrent goroutines
	numGoroutines := 50
	numRequestsPerGoroutine := 10
	ctx := context.Background()

	// Use standardized concurrent test utility with error collection
	testutil.RunConcurrentTestWithErrors(t, numGoroutines, func(workerID int) error {
		for j := 0; j < numRequestsPerGoroutine; j++ {
			// Rotate through different request types
			switch j % 3 {
			case 0:
				_, err := client.ListTools(ctx, mcp.ListToolsRequest{})
				if err != nil {
					return fmt.Errorf("ListTools failed: %v", err)
				}
			case 1:
				_, err := client.ListResources(ctx, mcp.ListResourcesRequest{})
				if err != nil {
					return fmt.Errorf("ListResources failed: %v", err)
				}
			case 2:
				_, err := client.ListPrompts(ctx, mcp.ListPromptsRequest{})
				if err != nil {
					return fmt.Errorf("ListPrompts failed: %v", err)
				}
			}
		}
		return nil
	})

	// Verify call counts - exact distribution since we rotate through methods
	toolsCalls := client.GetCallCount("ListTools")
	if toolsCalls != 200 {
		t.Errorf("Expected 200 ListTools calls, got %d", toolsCalls)
	}

	resourcesCalls := client.GetCallCount("ListResources")
	if resourcesCalls != 150 {
		t.Errorf("Expected 150 ListResources calls, got %d", resourcesCalls)
	}

	promptsCalls := client.GetCallCount("ListPrompts")
	if promptsCalls != 150 {
		t.Errorf("Expected 150 ListPrompts calls, got %d", promptsCalls)
	}

	// Verify total calls
	totalCalls := len(client.GetCalls())
	expectedTotal := numGoroutines * numRequestsPerGoroutine
	if totalCalls != expectedTotal {
		t.Errorf("Expected %d total calls, got %d", expectedTotal, totalCalls)
	}
}

// TestConcurrentServerHandshakes_Refactored with standardized utilities
func TestConcurrentServerHandshakes_Refactored(t *testing.T) {
	server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
	defer server.Reset()

	ctx := context.Background()
	numClients := 20

	// Use standardized concurrent test with options for better control
	opts := testutil.ConcurrentTestOptions{
		NumGoroutines: numClients,
		Timeout:       30 * time.Second,
		Description:   "Concurrent handshake test",
	}

	successCount := atomic.Int32{}

	testutil.RunConcurrentTestWithOptions(t, opts, func(clientID int) error {
		connID := fmt.Sprintf("concurrent-client-%d", clientID)

		result, err := server.SimulateClientMessage(ctx, connID, "initialize", map[string]interface{}{
			"protocolVersion": "1.0",
			"clientInfo": map[string]interface{}{
				"name":    fmt.Sprintf("Client %d", clientID),
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{},
		}, fmt.Sprintf("init-%d", clientID))

		if err != nil {
			return fmt.Errorf("handshake failed: %v", err)
		}

		// Verify response
		if respMap, ok := result.(map[string]interface{}); ok {
			if _, hasVersion := respMap["protocolVersion"]; !hasVersion {
				return fmt.Errorf("missing protocolVersion in response")
			}
		}

		// Verify connection is ready
		conn, ok := server.GetConnectionManager().GetConnection(connID)
		if !ok || !conn.IsReady() {
			return fmt.Errorf("connection not ready after handshake")
		}

		successCount.Add(1)
		return nil
	})

	// Verify all handshakes succeeded
	if successCount.Load() != int32(numClients) {
		t.Errorf("Expected %d successful handshakes, got %d", numClients, successCount.Load())
	}

	// Verify server tracked all requests
	if server.GetRequestCount("initialize") != numClients {
		t.Errorf("Expected %d initialize requests, got %d",
			numClients, server.GetRequestCount("initialize"))
	}
}

// TestConcurrentMixedOperations_Refactored demonstrates complex concurrent scenarios
func TestConcurrentMixedOperations_Refactored(t *testing.T) {
	server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
	defer server.Reset()

	ctx := context.Background()

	// First, establish connections
	numConnections := 10
	for i := 0; i < numConnections; i++ {
		connID := fmt.Sprintf("mixed-conn-%d", i)
		_, err := server.SimulateClientMessage(ctx, connID, "initialize", map[string]interface{}{
			"protocolVersion": "1.0",
			"clientInfo": map[string]interface{}{
				"name":    fmt.Sprintf("Client %d", i),
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{},
		}, fmt.Sprintf("init-%d", i))

		if err != nil {
			t.Fatalf("Failed to initialize connection %d: %v", i, err)
		}
	}

	// Use standardized utility for mixed operations
	operationsPerConnection := 20
	
	testutil.RunConcurrentTestWithErrors(t, numConnections, func(connIdx int) error {
		connID := fmt.Sprintf("mixed-conn-%d", connIdx)

		for op := 0; op < operationsPerConnection; op++ {
			// Rotate through different operations
			switch op % 4 {
			case 0: // List tools
				_, err := server.SimulateClientMessage(ctx, connID, "tools/list", nil,
					fmt.Sprintf("tools-%d-%d", connIdx, op))
				// Expected to fail - tools not supported
				if err == nil {
					return fmt.Errorf("tools/list should have failed")
				}

			case 1: // Call tool
				_, err := server.SimulateClientMessage(ctx, connID, "tools/call",
					map[string]interface{}{
						"name": "echo",
						"arguments": map[string]interface{}{
							"message": fmt.Sprintf("Hello from %d-%d", connIdx, op),
						},
					}, fmt.Sprintf("call-%d-%d", connIdx, op))
				// Expected to fail - tools not supported
				if err == nil {
					return fmt.Errorf("tools/call should have failed")
				}

			case 2: // List resources
				_, err := server.SimulateClientMessage(ctx, connID, "resources/list", nil,
					fmt.Sprintf("resources-%d-%d", connIdx, op))
				// Expected to fail - resources not supported
				if err == nil {
					return fmt.Errorf("resources/list should have failed")
				}

			case 3: // Ping
				_, err := server.SimulateClientMessage(ctx, connID, "ping", nil,
					fmt.Sprintf("ping-%d-%d", connIdx, op))
				if err != nil {
					return fmt.Errorf("ping failed: %v", err)
				}
			}

			// Small random delay to increase contention
			time.Sleep(time.Microsecond * time.Duration(connIdx*10))
		}
		return nil
	})

	// Verify all operations were recorded
	totalExpected := numConnections + (numConnections * operationsPerConnection)
	totalRequests := len(server.GetRequests())
	if totalRequests != totalExpected {
		t.Errorf("Expected %d total requests, got %d", totalExpected, totalRequests)
	}
}