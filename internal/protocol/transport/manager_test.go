package transport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// TestManager tests basic manager functionality
func TestManager(t *testing.T) {
	manager := NewManager()

	// Test adding a connection
	config := &ConnectionConfig{
		Type:    ConnectionTypeSTDIO,
		Command: "cat",
		Args:    []string{},
	}

	err := manager.AddConnection("test1", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Test getting connection
	transport, exists := manager.GetConnection("test1")
	if !exists {
		t.Error("Connection should exist")
	}
	if transport == nil {
		t.Error("Transport should not be nil")
	}

	// Test listing connections
	ids := manager.ListConnections()
	if len(ids) != 1 || ids[0] != "test1" {
		t.Errorf("Expected [test1], got %v", ids)
	}

	// Test connection info
	info, exists := manager.GetConnectionInfo("test1")
	if !exists {
		t.Error("Connection info should exist")
	}
	if info.Type != ConnectionTypeSTDIO {
		t.Errorf("Expected STDIO type, got %v", info.Type)
	}
	if !info.Connected {
		t.Error("Connection should be connected")
	}

	// Cleanup
	manager.Close()
}

// TestManagerDuplicateConnection tests adding duplicate connections
func TestManagerDuplicateConnection(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := &ConnectionConfig{
		Type:    ConnectionTypeSTDIO,
		Command: "cat",
	}

	// Add first connection
	err := manager.AddConnection("test1", config)
	if err != nil {
		t.Fatalf("Failed to add first connection: %v", err)
	}

	// Try to add duplicate
	err = manager.AddConnection("test1", config)
	if err == nil {
		t.Error("Should fail when adding duplicate connection")
	}
}

// TestManagerRemoveConnection tests removing connections
func TestManagerRemoveConnection(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := &ConnectionConfig{
		Type:    ConnectionTypeSTDIO,
		Command: "cat",
	}

	// Add connection
	err := manager.AddConnection("test1", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Remove connection
	err = manager.RemoveConnection("test1")
	if err != nil {
		t.Fatalf("Failed to remove connection: %v", err)
	}

	// Verify it's gone
	_, exists := manager.GetConnection("test1")
	if exists {
		t.Error("Connection should not exist after removal")
	}

	// Try to remove non-existent connection
	err = manager.RemoveConnection("test1")
	if err == nil {
		t.Error("Should fail when removing non-existent connection")
	}
}

// TestManagerBroadcast tests broadcasting to multiple connections
func TestManagerBroadcast(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// Add multiple connections
	for i := 1; i <= 3; i++ {
		config := &ConnectionConfig{
			Type:    ConnectionTypeSTDIO,
			Command: "cat",
		}
		err := manager.AddConnection(t.Name()+string(rune(i)), config)
		if err != nil {
			t.Fatalf("Failed to add connection %d: %v", i, err)
		}
	}

	// Broadcast a message
	ctx := context.Background()
	notification := &jsonrpc.Notification{
		Version: "2.0",
		Method:  "broadcast_test",
		Params:  json.RawMessage(`{"message": "Hello all!"}`),
	}

	err := manager.Broadcast(ctx, notification)
	if err != nil {
		t.Fatalf("Broadcast failed: %v", err)
	}
}

// TestManagerHealthCheck tests health checking
func TestManagerHealthCheck(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// Add a connection
	config := &ConnectionConfig{
		Type:    ConnectionTypeSTDIO,
		Command: "cat",
	}
	err := manager.AddConnection("test1", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Check health
	health := manager.HealthCheck()
	if len(health) != 1 {
		t.Errorf("Expected 1 health status, got %d", len(health))
	}

	status, exists := health["test1"]
	if !exists {
		t.Error("Health status should exist for test1")
	}
	if !status.Connected {
		t.Error("Connection should be healthy")
	}
	if status.ProcessID == 0 {
		t.Error("Process ID should not be 0")
	}
}

// TestManagerRestartConnection tests restarting a connection
func TestManagerRestartConnection(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// Add a connection
	config := &ConnectionConfig{
		Type:    ConnectionTypeSTDIO,
		Command: "cat",
	}
	err := manager.AddConnection("test1", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Get original process ID
	info1, _ := manager.GetConnectionInfo("test1")
	originalPID := info1.ProcessID

	// Restart connection
	err = manager.RestartConnection("test1")
	if err != nil {
		t.Fatalf("Failed to restart connection: %v", err)
	}

	// Verify new process ID
	info2, _ := manager.GetConnectionInfo("test1")
	if info2.ProcessID == originalPID {
		t.Error("Process ID should change after restart")
	}
	if !info2.Connected {
		t.Error("Connection should be connected after restart")
	}
}

// TestManagerInvalidConnectionType tests invalid connection type
func TestManagerInvalidConnectionType(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := &ConnectionConfig{
		Type: ConnectionType("invalid"),
	}

	err := manager.AddConnection("test1", config)
	if err == nil {
		t.Error("Should fail with invalid connection type")
	}
}

// TestManagerHTTPNotImplemented tests HTTP transport (not yet implemented)
func TestManagerHTTPNotImplemented(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := &ConnectionConfig{
		Type: ConnectionTypeHTTP,
		URL:  "http://localhost:8080",
	}

	err := manager.AddConnection("test1", config)
	if err == nil {
		t.Error("HTTP transport should not be implemented yet")
	}
}

// TestManagerEmptyCommand tests STDIO with empty command
func TestManagerEmptyCommand(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	config := &ConnectionConfig{
		Type:    ConnectionTypeSTDIO,
		Command: "", // Empty command
	}

	err := manager.AddConnection("test1", config)
	if err == nil {
		t.Error("Should fail with empty command")
	}
}

// TestManagerConcurrentOperations tests concurrent manager operations
func TestManagerConcurrentOperations(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// Concurrent adds
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			config := &ConnectionConfig{
				Type:    ConnectionTypeSTDIO,
				Command: "cat",
			}
			err := manager.AddConnection(t.Name()+string(rune(id)), config)
			if err != nil {
				t.Errorf("Failed to add connection %d: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all adds
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all connections exist
	ids := manager.ListConnections()
	if len(ids) != 5 {
		t.Errorf("Expected 5 connections, got %d", len(ids))
	}
}

// TestManagerCloseAll tests closing all connections
func TestManagerCloseAll(t *testing.T) {
	manager := NewManager()

	// Add multiple connections
	for i := 1; i <= 3; i++ {
		config := &ConnectionConfig{
			Type:    ConnectionTypeSTDIO,
			Command: "cat",
		}
		err := manager.AddConnection(t.Name()+string(rune(i)), config)
		if err != nil {
			t.Fatalf("Failed to add connection %d: %v", i, err)
		}
	}

	// Close all
	err := manager.Close()
	if err != nil {
		t.Fatalf("Failed to close all connections: %v", err)
	}

	// Verify all connections are gone
	ids := manager.ListConnections()
	if len(ids) != 0 {
		t.Errorf("Expected 0 connections after close, got %d", len(ids))
	}
}

// TestManagerBroadcastWithDisconnected tests broadcast with some disconnected transports
func TestManagerBroadcastWithDisconnected(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// Add a connection that will exit quickly
	config1 := &ConnectionConfig{
		Type:    ConnectionTypeSTDIO,
		Command: "sh",
		Args:    []string{"-c", "exit 0"},
	}
	err := manager.AddConnection("dying", config1)
	if err != nil {
		t.Fatalf("Failed to add dying connection: %v", err)
	}

	// Add a normal connection
	config2 := &ConnectionConfig{
		Type:    ConnectionTypeSTDIO,
		Command: "cat",
	}
	err = manager.AddConnection("alive", config2)
	if err != nil {
		t.Fatalf("Failed to add alive connection: %v", err)
	}

	// Wait for first connection to die
	time.Sleep(100 * time.Millisecond)

	// Broadcast should still work (only sends to connected transports)
	ctx := context.Background()
	notification := &jsonrpc.Notification{
		Version: "2.0",
		Method:  "test",
	}

	err = manager.Broadcast(ctx, notification)
	if err != nil {
		t.Fatalf("Broadcast should succeed even with disconnected transports: %v", err)
	}
}