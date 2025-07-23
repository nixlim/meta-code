package handlers

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/meta-mcp/meta-mcp-server/internal/logging"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
	"github.com/meta-mcp/meta-mcp-server/test/testutil"
)

func TestCreateInitializeHooks(t *testing.T) {
	manager := testutil.CreateTestManager()
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
	manager := testutil.CreateTestManager()
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
	manager := testutil.CreateTestManager()
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

func TestValidateVersionCompatibility(t *testing.T) {
	supportedVersions := []string{"1.0", "0.1.0", "2.0"}

	tests := []struct {
		name              string
		clientVersion     string
		supportedVersions []string
		wantErr           bool
	}{
		{
			name:              "supported_version",
			clientVersion:     "1.0",
			supportedVersions: supportedVersions,
			wantErr:           false,
		},
		{
			name:              "unsupported_version",
			clientVersion:     "3.0",
			supportedVersions: supportedVersions,
			wantErr:           true,
		},
		{
			name:              "empty_version",
			clientVersion:     "",
			supportedVersions: supportedVersions,
			wantErr:           true,
		},
		{
			name:              "no_supported_versions",
			clientVersion:     "1.0",
			supportedVersions: []string{},
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVersionCompatibility(tt.clientVersion, tt.supportedVersions)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVersionCompatibility() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogClientCapabilities(t *testing.T) {
	ctx := context.Background()
	logger := logging.Default().WithComponent("test")

	tests := []struct {
		name string
		caps *mcp.ClientCapabilities
	}{
		{
			name: "nil_capabilities",
			caps: nil,
		},
		{
			name: "empty_capabilities",
			caps: &mcp.ClientCapabilities{},
		},
		{
			name: "full_capabilities",
			caps: &mcp.ClientCapabilities{
				Experimental: map[string]interface{}{"feature": "value"},
				Sampling:     &struct{}{},
				Roots: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{
					ListChanged: true,
				},
			},
		},
		{
			name: "partial_capabilities",
			caps: &mcp.ClientCapabilities{
				Roots: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{
					ListChanged: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function only logs, so we just verify it doesn't panic
			logClientCapabilities(ctx, logger, tt.caps)
		})
	}
}

func TestLogServerCapabilities(t *testing.T) {
	ctx := context.Background()
	logger := logging.Default().WithComponent("test")

	tests := []struct {
		name string
		caps *mcp.ServerCapabilities
	}{
		{
			name: "nil_capabilities",
			caps: nil,
		},
		{
			name: "empty_capabilities",
			caps: &mcp.ServerCapabilities{},
		},
		{
			name: "full_capabilities",
			caps: &mcp.ServerCapabilities{
				Experimental: map[string]interface{}{"feature": "value"},
				Logging:      &struct{}{},
				Prompts: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{
					ListChanged: true,
				},
			},
		},
		{
			name: "partial_capabilities",
			caps: &mcp.ServerCapabilities{
				Prompts: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{
					ListChanged: false,
				},
			},
		},
		{
			name: "with_resources",
			caps: &mcp.ServerCapabilities{
				Resources: &struct {
					Subscribe   bool `json:"subscribe,omitempty"`
					ListChanged bool `json:"listChanged,omitempty"`
				}{
					Subscribe:   true,
					ListChanged: true,
				},
			},
		},
		{
			name: "with_tools",
			caps: &mcp.ServerCapabilities{
				Tools: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{
					ListChanged: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function only logs, so we just verify it doesn't panic
			logServerCapabilities(ctx, logger, tt.caps)
		})
	}
}

// Test error cases and edge scenarios for CreateInitializeHooks
func TestCreateInitializeHooksEdgeCases(t *testing.T) {
	t.Run("no_connection_in_context", func(t *testing.T) {
		manager := testutil.CreateTestManager()
		config := InitializeHooksConfig{
			ConnectionManager: manager,
			SupportedVersions: []string{"1.0"},
		}

		beforeHook, _ := CreateInitializeHooks(config)
		
		// Call without connection in context
		ctx := context.Background()
		request := &mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: "1.0",
			},
		}

		// Should not panic, just log warning
		beforeHook(ctx, "req-1", request)
	})

	t.Run("unsupported_version", func(t *testing.T) {
		manager := testutil.CreateTestManager()
		conn, _ := manager.CreateConnection("test-unsupported")
		
		config := InitializeHooksConfig{
			ConnectionManager: manager,
			SupportedVersions: []string{"1.0"},
		}

		beforeHook, _ := CreateInitializeHooks(config)
		
		ctx := connection.WithConnectionID(context.Background(), "test-unsupported")
		request := &mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: "99.0", // Unsupported version
			},
		}

		// Should handle gracefully
		beforeHook(ctx, "req-1", request)
		
		// Connection should remain in New state since handshake wasn't started
		if conn.GetState() != connection.StateNew {
			t.Errorf("Expected StateNew (handshake not started for unsupported version), got %v", conn.GetState())
		}
	})

	t.Run("connection_not_found", func(t *testing.T) {
		manager := testutil.CreateTestManager()
		config := InitializeHooksConfig{
			ConnectionManager: manager,
			SupportedVersions: []string{"1.0"},
		}

		_, afterHook := CreateInitializeHooks(config)
		
		// Context with non-existent connection ID
		ctx := connection.WithConnectionID(context.Background(), "non-existent")
		request := &mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: "1.0",
			},
		}
		result := &mcp.InitializeResult{
			ProtocolVersion: "1.0",
		}

		// Should handle gracefully
		afterHook(ctx, "req-1", request, result)
	})

	t.Run("with_client_capabilities", func(t *testing.T) {
		manager := testutil.CreateTestManager()
		conn, _ := manager.CreateConnection("test-caps")
		
		config := InitializeHooksConfig{
			ConnectionManager: manager,
			SupportedVersions: []string{"1.0"},
		}

		beforeHook, _ := CreateInitializeHooks(config)
		
		ctx := connection.WithConnectionID(context.Background(), "test-caps")
		request := &mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: "1.0",
				Capabilities: mcp.ClientCapabilities{
					Experimental: map[string]interface{}{"test": true},
					Roots: &struct {
						ListChanged bool `json:"listChanged,omitempty"`
					}{
						ListChanged: true,
					},
				},
			},
		}

		// Should log capabilities
		beforeHook(ctx, "req-1", request)
		
		if conn.GetState() != connection.StateInitializing {
			t.Errorf("Expected StateInitializing, got %v", conn.GetState())
		}
	})

	t.Run("with_server_capabilities", func(t *testing.T) {
		manager := testutil.CreateTestManager()
		conn, _ := manager.CreateConnection("test-server-caps")
		conn.StartHandshake(nil)
		
		config := InitializeHooksConfig{
			ConnectionManager: manager,
			SupportedVersions: []string{"1.0"},
			ServerInfo: mcp.Implementation{
				Name:    "Test Server",
				Version: "1.0.0",
			},
		}

		_, afterHook := CreateInitializeHooks(config)
		
		ctx := connection.WithConnectionID(context.Background(), "test-server-caps")
		request := &mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: "1.0",
			},
		}
		
		// Create capabilities for the result
		capabilities := &mcp.ServerCapabilities{
			Tools: &struct {
				ListChanged bool `json:"listChanged,omitempty"`
			}{
				ListChanged: true,
			},
			Resources: &struct {
				Subscribe   bool `json:"subscribe,omitempty"`
				ListChanged bool `json:"listChanged,omitempty"`
			}{
				Subscribe: true,
			},
		}
		
		result := &mcp.InitializeResult{
			ProtocolVersion: "1.0",
			Capabilities: *capabilities,
		}

		// Should log server capabilities
		afterHook(ctx, "req-1", request, result)
		
		if conn.GetState() != connection.StateReady {
			t.Errorf("Expected StateReady, got %v", conn.GetState())
		}
	})
}

// Test concurrent access
func TestCreateInitializeHooksConcurrency(t *testing.T) {
	manager := testutil.CreateTestManager()
	config := InitializeHooksConfig{
		ConnectionManager: manager,
		SupportedVersions: []string{"1.0"},
		ServerInfo: mcp.Implementation{
			Name:    "Test Server",
			Version: "1.0.0",
		},
	}

	beforeHook, afterHook := CreateInitializeHooks(config)

	// Create multiple connections
	for i := 0; i < 10; i++ {
		manager.CreateConnection(string(rune('0' + i)))
	}

	// Test concurrent calls
	done := make(chan bool, 20)

	// Before hook calls
	for i := 0; i < 10; i++ {
		go func(id int) {
			ctx := connection.WithConnectionID(context.Background(), string(rune('0'+id)))
			request := &mcp.InitializeRequest{
				Params: mcp.InitializeParams{
					ProtocolVersion: "1.0",
				},
			}
			beforeHook(ctx, id, request)
			done <- true
		}(i % 10)
	}

	// After hook calls
	for i := 0; i < 10; i++ {
		go func(id int) {
			ctx := connection.WithConnectionID(context.Background(), string(rune('0'+id)))
			request := &mcp.InitializeRequest{
				Params: mcp.InitializeParams{
					ProtocolVersion: "1.0",
				},
			}
			result := &mcp.InitializeResult{
				ProtocolVersion: "1.0",
			}
			afterHook(ctx, id, request, result)
			done <- true
		}(i % 10)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

// Commented out - SelectProtocolVersion function was removed in favor of isVersionSupported
/*
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
*/
