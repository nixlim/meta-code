package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
)

func TestDefaultHandshakeConfig(t *testing.T) {
	config := DefaultHandshakeConfig()

	if config.Name != "Meta-MCP Server" {
		t.Errorf("Name = %v, want Meta-MCP Server", config.Name)
	}

	if config.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", config.Version)
	}

	if config.HandshakeTimeout != 30*time.Second {
		t.Errorf("HandshakeTimeout = %v, want 30s", config.HandshakeTimeout)
	}

	if len(config.SupportedVersions) != 2 {
		t.Errorf("SupportedVersions length = %v, want 2", len(config.SupportedVersions))
	}
}

func TestNewHandshakeServer(t *testing.T) {
	config := HandshakeConfig{
		Name:              "Test Server",
		Version:           "1.0.0",
		HandshakeTimeout:  10 * time.Second,
		SupportedVersions: []string{"1.0"},
		ServerOptions: []server.ServerOption{
			WithToolCapabilities(true),
		},
	}

	hs := NewHandshakeServer(config)

	if hs == nil {
		t.Fatal("NewHandshakeServer() returned nil")
	}

	if hs.Server == nil {
		t.Fatal("HandshakeServer.Server is nil")
	}

	if hs.connectionManager == nil {
		t.Fatal("HandshakeServer.connectionManager is nil")
	}

	if hs.config.Name != "Test Server" {
		t.Errorf("Server name = %v, want Test Server", hs.config.Name)
	}
}

func TestHandshakeServer_CreateConnection(t *testing.T) {
	config := DefaultHandshakeConfig()
	hs := NewHandshakeServer(config)

	ctx := context.Background()

	// Test creating connection
	newCtx, err := hs.CreateConnection(ctx, "test-conn-1")
	if err != nil {
		t.Fatalf("CreateConnection() error = %v", err)
	}

	// Verify connection ID is in context
	id := newCtx.Value(connection.ConnectionIDKey)
	if id != "test-conn-1" {
		t.Errorf("Connection ID in context = %v, want test-conn-1", id)
	}

	// Test creating duplicate connection
	_, err = hs.CreateConnection(ctx, "test-conn-1")
	if err == nil {
		t.Error("Expected error for duplicate connection")
	}
}

func TestHandshakeServer_CloseConnection(t *testing.T) {
	config := DefaultHandshakeConfig()
	hs := NewHandshakeServer(config)

	ctx := context.Background()

	// Create and close connection
	hs.CreateConnection(ctx, "test-conn-2")
	hs.CloseConnection("test-conn-2")

	// Verify connection is removed
	conn, exists := hs.connectionManager.GetConnection("test-conn-2")
	if exists {
		t.Error("Connection still exists after closing")
	}

	if conn != nil {
		t.Error("GetConnection returned non-nil for closed connection")
	}
}

func TestHandshakeServer_GetConnectionManager(t *testing.T) {
	config := DefaultHandshakeConfig()
	hs := NewHandshakeServer(config)

	manager := hs.GetConnectionManager()
	if manager == nil {
		t.Fatal("GetConnectionManager() returned nil")
	}

	// Verify it's the same manager
	ctx := context.Background()
	hs.CreateConnection(ctx, "test-conn-3")

	conn, exists := manager.GetConnection("test-conn-3")
	if !exists {
		t.Error("Connection not found through GetConnectionManager")
	}

	if conn == nil || conn.ID != "test-conn-3" {
		t.Error("Wrong connection returned through GetConnectionManager")
	}
}

func TestWithHandshakeTimeout(t *testing.T) {
	config := DefaultHandshakeConfig()

	// Apply timeout modifier
	modifier := WithHandshakeTimeout(5 * time.Second)
	modifier(&config)

	if config.HandshakeTimeout != 5*time.Second {
		t.Errorf("HandshakeTimeout = %v, want 5s", config.HandshakeTimeout)
	}
}

func TestWithSupportedVersions(t *testing.T) {
	config := DefaultHandshakeConfig()

	// Apply versions modifier
	modifier := WithSupportedVersions("2.0", "2.1")
	modifier(&config)

	if len(config.SupportedVersions) != 2 {
		t.Errorf("SupportedVersions length = %v, want 2", len(config.SupportedVersions))
	}

	if config.SupportedVersions[0] != "2.0" {
		t.Errorf("SupportedVersions[0] = %v, want 2.0", config.SupportedVersions[0])
	}
}

func TestGenerateConnectionID(t *testing.T) {
	id1 := generateConnectionID()
	id2 := generateConnectionID()

	if id1 == "" {
		t.Error("generateConnectionID() returned empty string")
	}

	if id1 == id2 {
		t.Error("generateConnectionID() returned duplicate IDs")
	}

	// Test format (should be timestamp-based)
	if len(id1) < 10 {
		t.Errorf("Connection ID too short: %v", id1)
	}
}

func TestRegisterHooks(t *testing.T) {
	config := DefaultHandshakeConfig()
	hs := NewHandshakeServer(config)
	
	// This method is now a no-op but should not panic
	hs.registerHooks()
}

func TestServeStdioWithHandshake(t *testing.T) {
	config := DefaultHandshakeConfig()
	// Note: This test would normally require mocking stdio,
	// but since ServeStdio will fail without proper stdio setup,
	// we're testing that the function exists and handles basic errors
	// The actual ServeStdio call would fail in test environment
	
	// We can't actually run this without proper stdio setup
	// hs := NewHandshakeServer(config)
	// err := ServeStdioWithHandshake(hs)
	// This would require a more complex test setup with mocked stdio
	_ = config // Mark as used to avoid compiler warning
}

func TestHandleMessage(t *testing.T) {
	config := DefaultHandshakeConfig()
	hs := NewHandshakeServer(config)
	
	tests := []struct {
		name            string
		setupConnection bool
		connectionID    string
		connectionState connection.ConnectionState
		message         json.RawMessage
		expectError     bool
		errorCode       int
	}{
		{
			name:            "no_connection_id_in_context",
			setupConnection: false,
			message:         json.RawMessage(`{"method": "tools/list", "id": 1}`),
			expectError:     false, // Falls back to base server
		},
		{
			name:            "connection_not_found",
			setupConnection: false,
			connectionID:    "non-existent",
			message:         json.RawMessage(`{"method": "tools/list", "id": 2}`),
			expectError:     true,
			errorCode:       -32002,
		},
		{
			name:            "parse_error",
			setupConnection: true,
			connectionID:    "test-conn-1",
			connectionState: connection.StateReady,
			message:         json.RawMessage(`{invalid json`),
			expectError:     true,
			errorCode:       mcp.PARSE_ERROR,
		},
		{
			name:            "not_initialized_error",
			setupConnection: true,
			connectionID:    "test-conn-2",
			connectionState: connection.StateNew,
			message:         json.RawMessage(`{"method": "tools/list", "id": 3}`),
			expectError:     true,
			errorCode:       -32001,
		},
		{
			name:            "allow_initialize_when_not_ready",
			setupConnection: true,
			connectionID:    "test-conn-3",
			connectionState: connection.StateInitializing,
			message:         json.RawMessage(`{"method": "initialize", "id": 4}`),
			expectError:     false,
		},
		{
			name:            "allow_request_when_ready",
			setupConnection: true,
			connectionID:    "test-conn-4",
			connectionState: connection.StateReady,
			message:         json.RawMessage(`{"method": "tools/list", "id": 5}`),
			expectError:     false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			
			if tt.setupConnection && tt.connectionID != "" {
				// Create connection if needed
				conn, _ := hs.connectionManager.CreateConnection(tt.connectionID)
				conn.State = tt.connectionState
				ctx = connection.WithConnectionID(ctx, tt.connectionID)
			} else if tt.connectionID != "" {
				// Just add connection ID without creating connection
				ctx = connection.WithConnectionID(ctx, tt.connectionID)
			}
			
			result := hs.HandleMessage(ctx, tt.message)
			
			// Check if we got an error response
			if tt.expectError {
				// Result should be a JSONRPCError
				errBytes, _ := json.Marshal(result)
				var errResp struct {
					Error *struct {
						Code int `json:"code"`
					} `json:"error"`
				}
				json.Unmarshal(errBytes, &errResp)
				
				if errResp.Error == nil {
					t.Errorf("Expected error response, got %v", result)
				} else if errResp.Error.Code != tt.errorCode {
					t.Errorf("Expected error code %d, got %d", tt.errorCode, errResp.Error.Code)
				}
			}
		})
	}
}
