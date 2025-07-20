package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
)

func TestCreateInitializeHooks(t *testing.T) {
	manager := connection.NewManager(10 * time.Second)
	config := InitializeHooksConfig{
		ConnectionManager: manager,
		SupportedVersions: []string{"1.0", "0.1.0"},
		ServerInfo: mcp.Implementation{
			Name:    "Test Server",
			Version: "1.0.0",
		},
	}

	beforeHook, afterHook := CreateInitializeHooks(config)

	if beforeHook == nil {
		t.Error("CreateInitializeHooks() returned nil beforeHook")
	}

	if afterHook == nil {
		t.Error("CreateInitializeHooks() returned nil afterHook")
	}
}

func TestBeforeInitializeHook(t *testing.T) {
	manager := connection.NewManager(10 * time.Second)
	conn, _ := manager.CreateConnection("test-init-1")

	config := InitializeHooksConfig{
		ConnectionManager: manager,
		SupportedVersions: []string{"1.0"},
	}

	beforeHook, _ := CreateInitializeHooks(config)

	// Create context with connection
	ctx := connection.WithConnectionID(context.Background(), "test-init-1")

	// Create initialize request
	request := &mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "1.0",
			ClientInfo: mcp.Implementation{
				Name:    "Test Client",
				Version: "1.0.0",
			},
		},
	}

	// Call hook
	beforeHook(ctx, "req-1", request)

	// Verify connection state changed
	if conn.GetState() != connection.StateInitializing {
		t.Errorf("Connection state = %v, want StateInitializing", conn.GetState())
	}
}

func TestAfterInitializeHook(t *testing.T) {
	manager := connection.NewManager(10 * time.Second)
	conn, _ := manager.CreateConnection("test-init-2")

	// Start handshake first
	conn.StartHandshake(nil)

	config := InitializeHooksConfig{
		ConnectionManager: manager,
		SupportedVersions: []string{"1.0"},
	}

	_, afterHook := CreateInitializeHooks(config)

	// Create context with connection
	ctx := connection.WithConnectionID(context.Background(), "test-init-2")

	// Create initialize request and result
	request := &mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "1.0",
			ClientInfo: mcp.Implementation{
				Name:    "Test Client",
				Version: "1.0.0",
			},
		},
	}

	result := &mcp.InitializeResult{
		ProtocolVersion: "1.0",
		ServerInfo: mcp.Implementation{
			Name:    "Test Server",
			Version: "1.0.0",
		},
	}

	// Call hook
	afterHook(ctx, "req-1", request, result)

	// Verify connection state changed to ready
	if conn.GetState() != connection.StateReady {
		t.Errorf("Connection state = %v, want StateReady", conn.GetState())
	}

	// Verify protocol version stored
	if conn.ProtocolVersion != "1.0" {
		t.Errorf("ProtocolVersion = %v, want 1.0", conn.ProtocolVersion)
	}
}

func TestIsVersionSupported(t *testing.T) {
	supportedVersions := []string{"1.0", "0.1.0", "2.0"}

	tests := []struct {
		version string
		want    bool
	}{
		{"1.0", true},
		{"0.1.0", true},
		{"2.0", true},
		{"3.0", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := isVersionSupported(tt.version, supportedVersions); got != tt.want {
				t.Errorf("isVersionSupported(%v) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestSelectProtocolVersion(t *testing.T) {
	supportedVersions := []string{"2.0", "1.0", "0.1.0"}

	tests := []struct {
		clientVersion string
		want          string
	}{
		{"1.0", "1.0"}, // Exact match
		{"2.0", "2.0"}, // Exact match
		{"3.0", "2.0"}, // Not supported, use highest
		{"", "2.0"},    // Empty, use highest
	}

	for _, tt := range tests {
		t.Run(tt.clientVersion, func(t *testing.T) {
			if got := SelectProtocolVersion(tt.clientVersion, supportedVersions); got != tt.want {
				t.Errorf("SelectProtocolVersion(%v) = %v, want %v", tt.clientVersion, got, tt.want)
			}
		})
	}
}

func TestSelectProtocolVersion_EmptySupported(t *testing.T) {
	got := SelectProtocolVersion("any", []string{})
	if got != "1.0" {
		t.Errorf("SelectProtocolVersion with empty supported = %v, want 1.0", got)
	}
}
