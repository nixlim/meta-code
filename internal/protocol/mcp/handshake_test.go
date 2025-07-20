package mcp

import (
	"context"
	"testing"
	"time"

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
