package mcp_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpmock "github.com/meta-mcp/meta-mcp-server/internal/testing/mcp"
)

// TestConcurrentClientRequests tests concurrent requests to the mock client.
func TestConcurrentClientRequests(t *testing.T) {
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

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	ctx := context.Background()
	errors := make(chan error, numGoroutines*numRequestsPerGoroutine)

	// Launch concurrent requests
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < numRequestsPerGoroutine; j++ {
				// Rotate through different request types
				switch j % 3 {
				case 0:
					_, err := client.ListTools(ctx, mcp.ListToolsRequest{})
					if err != nil {
						errors <- fmt.Errorf("worker %d: ListTools failed: %v", workerID, err)
					}
				case 1:
					_, err := client.ListResources(ctx, mcp.ListResourcesRequest{})
					if err != nil {
						errors <- fmt.Errorf("worker %d: ListResources failed: %v", workerID, err)
					}
				case 2:
					_, err := client.ListPrompts(ctx, mcp.ListPromptsRequest{})
					if err != nil {
						errors <- fmt.Errorf("worker %d: ListPrompts failed: %v", workerID, err)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Verify call counts - exact distribution since we rotate through methods
	// With 50 goroutines * 10 requests = 500 total
	// Method 0 (ListTools): indices 0,3,6,9 = 4/10 * 500 = 200
	// Method 1 (ListResources): indices 1,4,7 = 3/10 * 500 = 150
	// Method 2 (ListPrompts): indices 2,5,8 = 3/10 * 500 = 150

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

// TestConcurrentServerHandshakes tests concurrent handshake handling.
func TestConcurrentServerHandshakes(t *testing.T) {
	server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
	defer server.Reset()

	ctx := context.Background()
	numClients := 20

	var wg sync.WaitGroup
	wg.Add(numClients)

	errors := make(chan error, numClients)
	successCount := int32(0)

	// Launch concurrent handshakes
	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			defer wg.Done()

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
				errors <- fmt.Errorf("client %d: handshake failed: %v", clientID, err)
				return
			}

			// Verify response
			if respMap, ok := result.(map[string]interface{}); ok {
				if _, hasVersion := respMap["protocolVersion"]; !hasVersion {
					errors <- fmt.Errorf("client %d: missing protocolVersion in response", clientID)
					return
				}
			}

			// Verify connection is ready
			conn, ok := server.GetConnectionManager().GetConnection(connID)
			if !ok || !conn.IsReady() {
				errors <- fmt.Errorf("client %d: connection not ready after handshake", clientID)
				return
			}

			atomic.AddInt32(&successCount, 1)
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Verify all handshakes succeeded
	if successCount != int32(numClients) {
		t.Errorf("Expected %d successful handshakes, got %d", numClients, successCount)
	}

	// Verify server tracked all requests
	if server.GetRequestCount("initialize") != numClients {
		t.Errorf("Expected %d initialize requests, got %d",
			numClients, server.GetRequestCount("initialize"))
	}
}

// TestConcurrentMixedOperations tests mixed concurrent operations.
func TestConcurrentMixedOperations(t *testing.T) {
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

	// Now run mixed operations concurrently
	var wg sync.WaitGroup
	operationsPerConnection := 20
	wg.Add(numConnections)

	errors := make(chan error, numConnections*operationsPerConnection)

	for i := 0; i < numConnections; i++ {
		go func(connIdx int) {
			defer wg.Done()

			connID := fmt.Sprintf("mixed-conn-%d", connIdx)

			for op := 0; op < operationsPerConnection; op++ {
				// Rotate through different operations
				switch op % 4 {
				case 0: // List tools
					_, err := server.SimulateClientMessage(ctx, connID, "tools/list", nil,
						fmt.Sprintf("tools-%d-%d", connIdx, op))
					// Expected to fail - tools not supported
					if err == nil {
						errors <- fmt.Errorf("conn %d op %d: tools/list should have failed",
							connIdx, op)
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
						errors <- fmt.Errorf("conn %d op %d: tools/call should have failed",
							connIdx, op)
					}

				case 2: // List resources
					_, err := server.SimulateClientMessage(ctx, connID, "resources/list", nil,
						fmt.Sprintf("resources-%d-%d", connIdx, op))
					// Expected to fail - resources not supported
					if err == nil {
						errors <- fmt.Errorf("conn %d op %d: resources/list should have failed",
							connIdx, op)
					}

				case 3: // Ping
					_, err := server.SimulateClientMessage(ctx, connID, "ping", nil,
						fmt.Sprintf("ping-%d-%d", connIdx, op))
					if err != nil {
						errors <- fmt.Errorf("conn %d op %d: ping failed: %v",
							connIdx, op, err)
					}
				}

				// Small random delay to increase contention
				time.Sleep(time.Microsecond * time.Duration(connIdx*10))
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Total errors: %d", errorCount)
	}

	// Verify all operations were recorded
	totalExpected := numConnections + (numConnections * operationsPerConnection)
	totalRequests := len(server.GetRequests())
	if totalRequests != totalExpected {
		t.Errorf("Expected %d total requests, got %d", totalExpected, totalRequests)
	}
}

// TestRaceConditions tests for race conditions using Go's race detector.
func TestRaceConditions(t *testing.T) {
	// This test will fail if run with -race and race conditions exist

	t.Run("ClientRaceConditions", func(t *testing.T) {
		client := mcpmock.NewMockClient()
		defer func() {
			if err := client.Close(); err != nil {
				t.Logf("Error closing client: %v", err)
			}
		}()

		client.SetResponse("Ping", nil)

		var wg sync.WaitGroup
		ctx := context.Background()

		// Multiple goroutines calling methods
		wg.Add(3)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if err := client.Ping(ctx); err != nil {
					// Ignore errors in race test
					continue
				}
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				client.GetCallCount("Ping")
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				client.GetCalls()
			}
		}()

		wg.Wait()
	})

	t.Run("ServerRaceConditions", func(t *testing.T) {
		server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
		defer server.Reset()

		var wg sync.WaitGroup
		ctx := context.Background()

		// Multiple goroutines accessing server
		wg.Add(3)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				connID := fmt.Sprintf("race-conn-%d", i)
				if _, err := server.SimulateClientMessage(ctx, connID, "initialize", map[string]interface{}{
					"protocolVersion": "1.0",
					"clientInfo": map[string]interface{}{
						"name":    "Race Test",
						"version": "1.0.0",
					},
				}, fmt.Sprintf("race-init-%d", i)); err != nil {
					// Ignore errors in race test
					continue
				}
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				server.GetRequestCount("initialize")
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				server.GetRequests()
			}
		}()

		wg.Wait()
	})
}

// TestConcurrentNotifications tests concurrent notification handling.
func TestConcurrentNotifications(t *testing.T) {
	client := mcpmock.NewMockClient()
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Error closing client: %v", err)
		}
	}()

	// Use atomic counter to track notifications
	var notificationCount int32
	var mu sync.Mutex
	receivedMethods := make(map[string]int)

	client.OnNotification(func(notification mcp.JSONRPCNotification) {
		atomic.AddInt32(&notificationCount, 1)

		mu.Lock()
		receivedMethods[notification.Method]++
		mu.Unlock()
	})

	// Send notifications from multiple goroutines
	var wg sync.WaitGroup
	numSenders := 10
	notificationsPerSender := 100

	wg.Add(numSenders)
	for i := 0; i < numSenders; i++ {
		go func(senderID int) {
			defer wg.Done()

			for j := 0; j < notificationsPerSender; j++ {
				method := fmt.Sprintf("notification.%d", j%5)
				client.SendNotification(mcp.JSONRPCNotification{
					JSONRPC: "2.0",
					Notification: mcp.Notification{
						Method: method,
					},
				})
			}
		}(i)
	}

	wg.Wait()

	// Verify all notifications were received
	expectedTotal := int32(numSenders * notificationsPerSender)
	if notificationCount != expectedTotal {
		t.Errorf("Expected %d notifications, got %d", expectedTotal, notificationCount)
	}

	// Verify distribution
	mu.Lock()
	for i := 0; i < 5; i++ {
		method := fmt.Sprintf("notification.%d", i)
		expected := (numSenders * notificationsPerSender) / 5
		if count, ok := receivedMethods[method]; !ok || count != expected {
			t.Errorf("Expected %d notifications for method %s, got %d",
				expected, method, count)
		}
	}
	mu.Unlock()
}

// BenchmarkConcurrentOperations benchmarks concurrent operations.
func BenchmarkConcurrentOperations(b *testing.B) {
	b.Run("ConcurrentClientRequests", func(b *testing.B) {
		client := mcpmock.NewMockClient()
		defer func() {
			if err := client.Close(); err != nil {
				b.Logf("Error closing client: %v", err)
			}
		}()

		client.SetResponse("Ping", nil)
		ctx := context.Background()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				err := client.Ping(ctx)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("ConcurrentCallTracking", func(b *testing.B) {
		client := mcpmock.NewMockClient()
		defer func() {
			if err := client.Close(); err != nil {
				b.Logf("Error closing client: %v", err)
			}
		}()

		ctx := context.Background()

		// Pre-populate some calls
		for i := 0; i < 1000; i++ {
			if err := client.Ping(ctx); err != nil {
				continue
			}
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = client.GetCallCount("Ping")
			}
		})
	})
}

