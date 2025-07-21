package mcp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpmock "github.com/meta-mcp/meta-mcp-server/internal/testing/mcp"
)

// TestClientRequestResponse tests basic request/response message flows.
func TestClientRequestResponse(t *testing.T) {
	client := mcpmock.NewMockClient()
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Error closing client: %v", err)
		}
	}()

	ctx := context.Background()

	// Test ListTools
	t.Run("ListTools", func(t *testing.T) {
		// Configure response
		expectedTools := []mcp.Tool{
			{
				Name:        "echo",
				Description: "Echoes input",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"message": map[string]interface{}{
							"type":        "string",
							"description": "Message to echo",
						},
					},
					Required: []string{"message"},
				},
			},
			{
				Name:        "calculator",
				Description: "Basic math operations",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"operation": map[string]interface{}{
							"type": "string",
							"enum": []string{"add", "subtract", "multiply", "divide"},
						},
						"a": map[string]interface{}{
							"type": "number",
						},
						"b": map[string]interface{}{
							"type": "number",
						},
					},
					Required: []string{"operation", "a", "b"},
				},
			},
		}

		client.SetResponse("ListTools", &mcp.ListToolsResult{
			Tools: expectedTools,
		})

		// Call ListTools
		result, err := client.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			t.Fatalf("ListTools failed: %v", err)
		}

		// Verify response
		if len(result.Tools) != 2 {
			t.Errorf("Expected 2 tools, got %d", len(result.Tools))
		}

		if result.Tools[0].Name != "echo" {
			t.Errorf("Expected first tool to be 'echo', got '%s'", result.Tools[0].Name)
		}

		// Verify call was tracked
		if client.GetCallCount("ListTools") != 1 {
			t.Errorf("Expected 1 ListTools call, got %d", client.GetCallCount("ListTools"))
		}
	})

	// Test CallTool
	t.Run("CallTool", func(t *testing.T) {
		// Configure response
		client.SetResponse("CallTool", &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Hello, World!"),
			},
		})

		// Call tool
		result, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "echo",
				Arguments: map[string]interface{}{
					"message": "Hello, World!",
				},
			},
		})

		if err != nil {
			t.Fatalf("CallTool failed: %v", err)
		}

		// Verify response
		if len(result.Content) != 1 {
			t.Errorf("Expected 1 content item, got %d", len(result.Content))
		}

		// Verify call arguments were tracked
		calls := client.GetCallsForMethod("CallTool")
		if len(calls) != 1 {
			t.Fatalf("Expected 1 CallTool call, got %d", len(calls))
		}

		callReq, ok := calls[0].Args.(mcp.CallToolRequest)
		if !ok {
			t.Fatalf("Expected CallToolRequest, got %T", calls[0].Args)
		}

		if callReq.Params.Name != "echo" {
			t.Errorf("Expected tool name 'echo', got '%s'", callReq.Params.Name)
		}
	})

	// Test ListResources
	t.Run("ListResources", func(t *testing.T) {
		// Configure response
		expectedResources := []mcp.Resource{
			{
				URI:         "file:///config.json",
				Name:        "Configuration",
				Description: "Application configuration",
				MIMEType:    "application/json",
			},
			{
				URI:         "file:///data.csv",
				Name:        "Data",
				Description: "Sample data",
				MIMEType:    "text/csv",
			},
		}

		client.SetResponse("ListResources", &mcp.ListResourcesResult{
			Resources: expectedResources,
		})

		// Call ListResources
		result, err := client.ListResources(ctx, mcp.ListResourcesRequest{})
		if err != nil {
			t.Fatalf("ListResources failed: %v", err)
		}

		// Verify response
		if len(result.Resources) != 2 {
			t.Errorf("Expected 2 resources, got %d", len(result.Resources))
		}

		if result.Resources[0].URI != "file:///config.json" {
			t.Errorf("Expected first resource URI to be 'file:///config.json', got '%s'", result.Resources[0].URI)
		}
	})

	// Test ReadResource
	t.Run("ReadResource", func(t *testing.T) {
		// Configure response
		// Configure response
		textContent := mcp.TextResourceContents{
			Text: `{"debug": true, "port": 8080}`,
		}
		textContent.URI = "file:///config.json"
		textContent.MIMEType = "application/json"

		client.SetResponse("ReadResource", &mcp.ReadResourceResult{
			Contents: []mcp.ResourceContents{
				textContent,
			},
		})

		// Read resource
		result, err := client.ReadResource(ctx, mcp.ReadResourceRequest{
			Params: mcp.ReadResourceParams{
				URI: "file:///config.json",
			},
		})

		if err != nil {
			t.Fatalf("ReadResource failed: %v", err)
		}

		// Verify response
		if len(result.Contents) != 1 {
			t.Errorf("Expected 1 content item, got %d", len(result.Contents))
		}

		// Type assert to check text content
		if textContent, ok := result.Contents[0].(mcp.TextResourceContents); ok {
			if textContent.Text != `{"debug": true, "port": 8080}` {
				t.Errorf("Unexpected content: %s", textContent.Text)
			}
		} else {
			t.Error("Expected TextResourceContents type")
		}
	})
}

// TestClientErrorHandling tests error scenarios.
func TestClientErrorHandling(t *testing.T) {
	client := mcpmock.NewMockClient()
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Error closing client: %v", err)
		}
	}()

	ctx := context.Background()

	// Test method-specific error
	t.Run("MethodError", func(t *testing.T) {
		expectedErr := errors.New("tool not found")
		client.SetError("CallTool", expectedErr)

		_, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "nonexistent",
			},
		})

		if err == nil {
			t.Error("Expected error, got none")
		}

		if err.Error() != expectedErr.Error() {
			t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
		}

		// Verify error was recorded
		calls := client.GetCallsForMethod("CallTool")
		if len(calls) != 1 {
			t.Fatalf("Expected 1 call, got %d", len(calls))
		}

		if calls[0].Error == nil {
			t.Error("Expected error to be recorded in call")
		}
	})

	// Test client closed error
	t.Run("ClientClosed", func(t *testing.T) {
		// Close client
		err := client.Close()
		if err != nil {
			t.Fatalf("Failed to close client: %v", err)
		}

		// Try to use closed client
		_, err = client.ListTools(ctx, mcp.ListToolsRequest{})
		if err == nil {
			t.Error("Expected error when using closed client")
		}

		if err.Error() != "client is closed" {
			t.Errorf("Expected 'client is closed' error, got '%v'", err)
		}
	})
}

// TestClientDelay tests delay simulation.
func TestClientDelay(t *testing.T) {
	client := mcpmock.NewMockClient()
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Error closing client: %v", err)
		}
	}()

	ctx := context.Background()

	// Set method-specific delay
	client.SetDelay("Ping", 50*time.Millisecond)

	// Measure ping time
	start := time.Now()
	err := client.Ping(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Verify delay was applied
	if elapsed < 50*time.Millisecond {
		t.Errorf("Expected delay of at least 50ms, got %v", elapsed)
	}

	// Test default delay
	client.SetDefaultDelay(30 * time.Millisecond)

	// Call method without specific delay
	start = time.Now()
	_, err = client.ListTools(ctx, mcp.ListToolsRequest{})
	elapsed = time.Since(start)

	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	// Verify default delay was applied
	if elapsed < 30*time.Millisecond {
		t.Errorf("Expected default delay of at least 30ms, got %v", elapsed)
	}
}

// TestClientNotifications tests notification handling.
func TestClientNotifications(t *testing.T) {
	client := mcpmock.NewMockClient()
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Error closing client: %v", err)
		}
	}()

	// Track received notifications
	var receivedNotifications []mcp.JSONRPCNotification
	client.OnNotification(func(notification mcp.JSONRPCNotification) {
		receivedNotifications = append(receivedNotifications, notification)
	})

	// Send notifications
	notifications := []mcp.JSONRPCNotification{
		{
			JSONRPC: "2.0",
			Notification: mcp.Notification{
				Method: "resources/updated",
			},
		},
		{
			JSONRPC: "2.0",
			Notification: mcp.Notification{
				Method: "tools/added",
			},
		},
		{
			JSONRPC: "2.0",
			Notification: mcp.Notification{
				Method: "log",
			},
		},
	}

	for _, notif := range notifications {
		client.SendNotification(notif)
	}

	// Verify all notifications were received
	if len(receivedNotifications) != 3 {
		t.Errorf("Expected 3 notifications, got %d", len(receivedNotifications))
	}

	// Verify notification content
	for i, notif := range receivedNotifications {
		if notif.Method != notifications[i].Method {
			t.Errorf("Notification %d: expected method '%s', got '%s'",
				i, notifications[i].Method, notif.Method)
		}
	}
}

// TestClientCallTracking tests call tracking functionality.
func TestClientCallTracking(t *testing.T) {
	client := mcpmock.NewMockClient()
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Error closing client: %v", err)
		}
	}()

	ctx := context.Background()

	// Make various calls
	_, err := client.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "1.0",
			ClientInfo: mcp.Implementation{
				Name:    "Test",
				Version: "1.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		t.Logf("Initialize error (expected): %v", err)
	}

	if err := client.Ping(ctx); err != nil {
		t.Logf("Ping error (expected): %v", err)
	}
	if err := client.Ping(ctx); err != nil {
		t.Logf("Ping error (expected): %v", err)
	}

	if _, err := client.ListTools(ctx, mcp.ListToolsRequest{}); err != nil {
		t.Logf("ListTools error (expected): %v", err)
	}

	// Verify call counts
	if client.GetCallCount("Initialize") != 1 {
		t.Errorf("Expected 1 Initialize call, got %d", client.GetCallCount("Initialize"))
	}

	if client.GetCallCount("Ping") != 2 {
		t.Errorf("Expected 2 Ping calls, got %d", client.GetCallCount("Ping"))
	}

	if client.GetCallCount("ListTools") != 1 {
		t.Errorf("Expected 1 ListTools call, got %d", client.GetCallCount("ListTools"))
	}

	// Verify total calls
	allCalls := client.GetCalls()
	if len(allCalls) != 4 {
		t.Errorf("Expected 4 total calls, got %d", len(allCalls))
	}

	// Verify call order
	expectedOrder := []string{"Initialize", "Ping", "Ping", "ListTools"}
	for i, call := range allCalls {
		if call.Method != expectedOrder[i] {
			t.Errorf("Call %d: expected method '%s', got '%s'",
				i, expectedOrder[i], call.Method)
		}
	}

	// Test Reset
	client.Reset()

	if client.GetCallCount("Ping") != 0 {
		t.Error("Expected call count to be 0 after reset")
	}

	if len(client.GetCalls()) != 0 {
		t.Error("Expected no calls after reset")
	}

	if client.IsInitialized() {
		t.Error("Expected client to not be initialized after reset")
	}
}

// TestClientPagination tests pagination support.
func TestClientPagination(t *testing.T) {
	client := mcpmock.NewMockClient()
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Error closing client: %v", err)
		}
	}()

	ctx := context.Background()

	// Configure paginated response
	page1Tools := []mcp.Tool{
		{Name: "tool1", Description: "Tool 1"},
		{Name: "tool2", Description: "Tool 2"},
	}

	// Configure paginated response without NextCursor field
	client.SetResponse("ListToolsByPage", &mcp.ListToolsResult{
		Tools: page1Tools,
		// Note: mcp-go doesn't have NextCursor field in ListToolsResult
	})

	// Get first page
	result, err := client.ListToolsByPage(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListToolsByPage failed: %v", err)
	}

	if len(result.Tools) != 2 {
		t.Errorf("Expected 2 tools on page 1, got %d", len(result.Tools))
	}

	// Note: mcp-go doesn't support cursor-based pagination in the same way
	// This is a limitation of the current library version

	// Configure page 2
	page2Tools := []mcp.Tool{
		{Name: "tool3", Description: "Tool 3"},
	}

	client.SetResponse("ListToolsByPage", &mcp.ListToolsResult{
		Tools: page2Tools,
	})

	// Get second page (without cursor support)
	result, err = client.ListToolsByPage(ctx, mcp.ListToolsRequest{})

	if err != nil {
		t.Fatalf("ListToolsByPage page 2 failed: %v", err)
	}

	if len(result.Tools) != 1 {
		t.Errorf("Expected 1 tool on page 2, got %d", len(result.Tools))
	}
}

// BenchmarkClientOperations benchmarks various client operations.
func BenchmarkClientOperations(b *testing.B) {
	client := mcpmock.NewMockClient()
	defer func() {
		if err := client.Close(); err != nil {
			b.Logf("Error closing client: %v", err)
		}
	}()

	ctx := context.Background()

	// Configure responses
	client.SetResponse("ListTools", &mcp.ListToolsResult{
		Tools: []mcp.Tool{
			{Name: "tool1", Description: "Tool 1"},
			{Name: "tool2", Description: "Tool 2"},
		},
	})

	b.Run("ListTools", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := client.ListTools(ctx, mcp.ListToolsRequest{})
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Ping", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := client.Ping(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CallTracking", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := client.Ping(ctx); err != nil {
				b.Fatal(err)
			}
			_ = client.GetCallCount("Ping")
		}
	})
}
