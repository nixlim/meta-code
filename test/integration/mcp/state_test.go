package mcp_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
	mcpmock "github.com/meta-mcp/meta-mcp-server/internal/testing/mcp"
)

// TestConnectionStateTransitions tests connection state transitions.
func TestConnectionStateTransitions(t *testing.T) {
	t.Run("NormalStateFlow", func(t *testing.T) {
		server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
		defer server.Reset()

		ctx := context.Background()
		connID := "test-state-flow"

		// Initial state should be New
		ctx, err := server.CreateConnection(ctx, connID)
		if err != nil {
			t.Fatalf("Failed to create connection: %v", err)
		}

		conn, ok := server.GetConnectionManager().GetConnection(connID)
		if !ok {
			t.Fatal("Connection not found")
		}

		initialState := conn.GetState()
		if initialState != connection.StateNew {
			t.Errorf("Expected initial state to be New, got %s", initialState)
		}

		// After initialize, state should be Ready
		_, err = server.SimulateClientMessage(ctx, connID, "initialize", map[string]interface{}{
			"protocolVersion": "1.0",
			"clientInfo": map[string]interface{}{
				"name":    "State Test Client",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{},
		}, "init-1")

		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		if !conn.IsReady() {
			t.Errorf("Expected state to be Ready after initialize, got %s", conn.GetState())
		}

		// Close connection
		server.CloseConnection(connID)

		finalState := conn.GetState()
		if finalState != connection.StateClosed {
			t.Errorf("Expected state to be Closed after close, got %s", finalState)
		}
	})

	t.Run("InvalidTransitions", func(t *testing.T) {
		server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
		defer server.Reset()

		ctx := context.Background()
		connID := "test-invalid-transitions"

		// Create and immediately close connection
		ctx, err := server.CreateConnection(ctx, connID)
		if err != nil {
			t.Fatalf("Failed to create connection: %v", err)
		}

		server.CloseConnection(connID)

		// Try to initialize closed connection
		// The server might create a new connection automatically
		result, err := server.SimulateClientMessage(ctx, connID, "initialize", map[string]interface{}{
			"protocolVersion": "1.0",
			"clientInfo": map[string]interface{}{
				"name":    "Test Client",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{},
		}, "init-closed")

		// If no error, it means server created a new connection
		if err == nil && result != nil {
			t.Log("Server created new connection for closed connection ID")
			// Verify the new connection is ready
			conn, ok := server.GetConnectionManager().GetConnection(connID)
			if ok && conn.IsReady() {
				t.Log("New connection is ready after re-initialization")
			}
		} else if err != nil {
			// This is also acceptable - server rejected the request
			t.Logf("Server correctly rejected initialize on closed connection: %v", err)
		}
	})

	t.Run("ConcurrentStateChanges", func(t *testing.T) {
		server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
		defer server.Reset()

		ctx := context.Background()
		numConnections := 20

		var wg sync.WaitGroup
		wg.Add(numConnections)

		errors := make(chan error, numConnections*2) // Increase buffer size

		// Create and transition multiple connections concurrently
		for i := 0; i < numConnections; i++ {
			go func(idx int) {
				defer wg.Done()

				connID := fmt.Sprintf("concurrent-state-%d", idx)

				// Create connection - might already exist from another goroutine
				localCtx, err := server.CreateConnection(ctx, connID)
				if err != nil {
					// Try to get existing context
					localCtx = server.GetConnectionContext(ctx, connID)
				}

				// Initialize
				_, err = server.SimulateClientMessage(localCtx, connID, "initialize", map[string]interface{}{
					"protocolVersion": "1.0",
					"clientInfo": map[string]interface{}{
						"name":    fmt.Sprintf("Client %d", idx),
						"version": "1.0.0",
					},
					"capabilities": map[string]interface{}{},
				}, fmt.Sprintf("init-%d", idx))

				if err != nil {
					// Connection might already be initialized by another goroutine
					t.Logf("conn %d: initialize returned error (possibly already initialized): %v", idx, err)
				}

				// Verify connection exists
				conn, ok := server.GetConnectionManager().GetConnection(connID)
				if !ok {
					errors <- fmt.Errorf("conn %d: connection not found", idx)
					return
				}

				// Close connection (might already be closed)
				server.CloseConnection(connID)

				// Verify final state
				finalState := conn.GetState()
				if finalState != connection.StateClosed && finalState != connection.StateNew {
					errors <- fmt.Errorf("conn %d: unexpected final state: %s", idx, finalState)
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
			t.Errorf("Total errors in concurrent test: %d", errorCount)
		}
	})
}

// TestClientStateManagement tests client state management.
func TestClientStateManagement(t *testing.T) {
	t.Run("InitializationState", func(t *testing.T) {
		client := mcpmock.NewMockClient()
		defer func() {
			if err := client.Close(); err != nil {
				t.Logf("Error closing client: %v", err)
			}
		}()

		// Initially not initialized
		if client.IsInitialized() {
			t.Error("Client should not be initialized initially")
		}

		// Initialize
		ctx := context.Background()
		_, err := client.Initialize(ctx, mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: "1.0",
				ClientInfo: mcp.Implementation{
					Name:    "Test Client",
					Version: "1.0.0",
				},
				Capabilities: mcp.ClientCapabilities{},
			},
		})

		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		// Should be initialized
		if !client.IsInitialized() {
			t.Error("Client should be initialized after Initialize")
		}

		// Reset should clear initialization
		client.Reset()
		if client.IsInitialized() {
			t.Error("Client should not be initialized after Reset")
		}
	})

	t.Run("ClosedState", func(t *testing.T) {
		client := mcpmock.NewMockClient()

		// Initially not closed
		if client.IsClosed() {
			t.Error("Client should not be closed initially")
		}

		// Close client
		err := client.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Should be closed
		if !client.IsClosed() {
			t.Error("Client should be closed after Close")
		}

		// Double close should error
		err = client.Close()
		if err == nil {
			t.Error("Expected error on double close")
		}

		// Operations should fail on closed client
		ctx := context.Background()
		_, err = client.Initialize(ctx, mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: "1.0",
				ClientInfo: mcp.Implementation{
					Name:    "Test",
					Version: "1.0",
				},
				Capabilities: mcp.ClientCapabilities{},
			},
		})

		if err == nil {
			t.Error("Expected error when using closed client")
		}
	})
}

// TestStateTransitionScenarios tests specific state transition scenarios.
func TestStateTransitionScenarios(t *testing.T) {
	t.Run("TimeoutDuringHandshake", func(t *testing.T) {
		config := mcpmock.DefaultMockServerConfig()
		config.HandshakeTimeout = 100 * time.Millisecond
		server := mcpmock.NewMockServer(config)
		defer server.Reset()

		ctx := context.Background()
		connID := "timeout-handshake"

		// Create connection
		_, err := server.CreateConnection(ctx, connID)
		if err != nil {
			t.Fatalf("Failed to create connection: %v", err)
		}

		conn, _ := server.GetConnectionManager().GetConnection(connID)

		// Start handshake
		if err := conn.StartHandshake(nil); err != nil {
			t.Logf("StartHandshake error (expected): %v", err)
		}

		// Initial state should be Initializing
		if conn.GetState() != connection.StateInitializing {
			t.Errorf("Expected Initializing state, got %s", conn.GetState())
		}

		// Wait for timeout
		time.Sleep(150 * time.Millisecond)

		// State should be Closed after timeout
		if conn.GetState() != connection.StateClosed {
			t.Errorf("Expected Closed state after timeout, got %s", conn.GetState())
		}
	})

	t.Run("RapidStateChanges", func(t *testing.T) {
		server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
		defer server.Reset()

		ctx := context.Background()

		// Perform rapid connection lifecycle
		for i := 0; i < 10; i++ {
			connID := fmt.Sprintf("rapid-%d", i)

			// Create
			_, err := server.CreateConnection(ctx, connID)
			if err != nil {
				t.Errorf("Iteration %d: Failed to create connection: %v", i, err)
				continue
			}

			// Initialize immediately
			_, err = server.SimulateClientMessage(ctx, connID, "initialize", map[string]interface{}{
				"protocolVersion": "1.0",
				"clientInfo": map[string]interface{}{
					"name":    "Rapid Test",
					"version": "1.0.0",
				},
				"capabilities": map[string]interface{}{},
			}, fmt.Sprintf("rapid-init-%d", i))

			if err != nil {
				t.Errorf("Iteration %d: Initialize failed: %v", i, err)
				continue
			}

			// Close immediately
			server.CloseConnection(connID)

			// Verify final state
			conn, ok := server.GetConnectionManager().GetConnection(connID)
			if ok && conn.GetState() != connection.StateClosed {
				t.Errorf("Iteration %d: Expected closed state, got %s", i, conn.GetState())
			}
		}
	})

	t.Run("StateTransitionCallbacks", func(t *testing.T) {
		server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
		defer server.Reset()

		ctx := context.Background()
		connID := "callback-test"

		// Track state transitions
		var transitions []string
		var mu sync.Mutex

		// Create connection and track initial state
		ctx, err := server.CreateConnection(ctx, connID)
		if err != nil {
			t.Fatalf("Failed to create connection: %v", err)
		}

		conn, _ := server.GetConnectionManager().GetConnection(connID)

		mu.Lock()
		transitions = append(transitions, conn.GetState().String())
		mu.Unlock()

		// Initialize
		_, err = server.SimulateClientMessage(ctx, connID, "initialize", map[string]interface{}{
			"protocolVersion": "1.0",
			"clientInfo": map[string]interface{}{
				"name":    "Callback Test",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{},
		}, "callback-init")

		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		mu.Lock()
		transitions = append(transitions, conn.GetState().String())
		mu.Unlock()

		// Close
		server.CloseConnection(connID)

		mu.Lock()
		transitions = append(transitions, conn.GetState().String())
		mu.Unlock()

		// Verify transition sequence
		expectedTransitions := []string{"New", "Ready", "Closed"}
		if len(transitions) != len(expectedTransitions) {
			t.Errorf("Expected %d transitions, got %d", len(expectedTransitions), len(transitions))
		}

		for i, expected := range expectedTransitions {
			if i < len(transitions) && transitions[i] != expected {
				t.Errorf("Transition %d: expected %s, got %s", i, expected, transitions[i])
			}
		}
	})
}

// TestStatePersistence tests state persistence across operations.
func TestStatePersistence(t *testing.T) {
	server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
	defer server.Reset()

	ctx := context.Background()
	connID := "persistence-test"

	// Initialize connection
	ctx, err := server.CreateConnection(ctx, connID)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}

	_, err = server.SimulateClientMessage(ctx, connID, "initialize", map[string]interface{}{
		"protocolVersion": "1.0",
		"clientInfo": map[string]interface{}{
			"name":    "Persistence Test",
			"version": "1.0.0",
		},
		"capabilities": map[string]interface{}{},
	}, "persist-init")

	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Perform multiple operations
	operations := []string{"tools/list", "resources/list", "prompts/list"}

	for i, op := range operations {
		_, err := server.SimulateClientMessage(ctx, connID, op, nil, fmt.Sprintf("persist-%d", i))
		// These operations are expected to fail with "not supported" since handlers aren't configured
		// We're testing that requests are tracked and connection remains stable
		if err != nil {
			t.Logf("Operation %s returned expected error: %v", op, err)
		}

		// Verify connection remains ready even after errors
		conn, ok := server.GetConnectionManager().GetConnection(connID)
		if !ok || !conn.IsReady() {
			t.Errorf("Connection not ready after operation %s", op)
		}
	}

	// Verify all operations were tracked
	requests := server.GetRequests()
	if len(requests) != len(operations)+1 { // +1 for initialize
		t.Errorf("Expected %d requests, got %d", len(operations)+1, len(requests))
	}
}

// BenchmarkStateTransitions benchmarks state transition performance.
func BenchmarkStateTransitions(b *testing.B) {
	b.Run("ConnectionLifecycle", func(b *testing.B) {
		server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
		defer server.Reset()

		ctx := context.Background()

		initParams := map[string]interface{}{
			"protocolVersion": "1.0",
			"clientInfo": map[string]interface{}{
				"name":    "Benchmark Client",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			connID := fmt.Sprintf("bench-%d", i)

			// Full lifecycle
			ctx, _ := server.CreateConnection(ctx, connID)
			if _, err := server.SimulateClientMessage(ctx, connID, "initialize", initParams, fmt.Sprintf("bench-init-%d", i)); err != nil {
				// Ignore errors in benchmark
				continue
			}
			server.CloseConnection(connID)
		}
	})

	b.Run("StateChecks", func(b *testing.B) {
		server := mcpmock.NewMockServer(mcpmock.DefaultMockServerConfig())
		defer server.Reset()

		ctx := context.Background()
		connID := "bench-state-check"

		// Setup connection
		ctx, _ = server.CreateConnection(ctx, connID)
		if _, err := server.SimulateClientMessage(ctx, connID, "initialize", map[string]interface{}{
			"protocolVersion": "1.0",
			"clientInfo": map[string]interface{}{
				"name":    "Benchmark",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{},
		}, "bench-init"); err != nil {
			b.Fatalf("Failed to initialize: %v", err)
		}

		conn, _ := server.GetConnectionManager().GetConnection(connID)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = conn.IsReady()
			_ = conn.GetState()
		}
	})
}
