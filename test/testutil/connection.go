// Package testutil provides common testing utilities for the meta-mcp-server project.
package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
)

// CreateTestManager creates a connection manager for testing with a default timeout.
func CreateTestManager() *connection.Manager {
	return connection.NewManager(10 * time.Second)
}

// CreateTestConnection creates a test connection with the given ID.
func CreateTestConnection(t *testing.T, manager *connection.Manager, id string) *connection.Connection {
	conn, err := manager.CreateConnection(id)
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	return conn
}

// CreateTestContext creates a context with a connection ID for testing.
func CreateTestContext(connID string) context.Context {
	return connection.WithConnectionID(context.Background(), connID)
}

// CreateTestManagerWithConnection creates a connection manager with a pre-configured connection.
// This is useful for tests that need a connection in a specific state.
func CreateTestManagerWithConnection(connID string, state connection.ConnectionState) *connection.Manager {
	manager := CreateTestManager()
	conn, _ := manager.CreateConnection(connID)
	conn.State = state
	return manager
}

// SetupTestConnection creates a manager and connection for testing.
func SetupTestConnection(t *testing.T, connID string) (*connection.Manager, *connection.Connection, context.Context) {
	manager := CreateTestManager()
	conn := CreateTestConnection(t, manager, connID)
	ctx := CreateTestContext(connID)
	return manager, conn, ctx
}

// StartHandshakeForTest starts handshake on a connection for testing.
func StartHandshakeForTest(t *testing.T, conn *connection.Connection) {
	if err := conn.StartHandshake(nil); err != nil {
		t.Fatalf("Failed to start handshake: %v", err)
	}
}

// CompleteHandshakeForTest completes handshake on a connection for testing.
func CompleteHandshakeForTest(t *testing.T, conn *connection.Connection, protocolVersion string) {
	if err := conn.CompleteHandshake(protocolVersion, nil); err != nil {
		t.Fatalf("Failed to complete handshake: %v", err)
	}
}