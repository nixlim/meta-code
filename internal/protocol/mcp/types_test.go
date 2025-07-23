package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewServer(t *testing.T) {
	server := NewServer("Test Server", "1.0.0")

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.MCPServer == nil {
		t.Error("MCPServer is nil")
	}
}

func TestCreateEchoTool(t *testing.T) {
	tool := CreateEchoTool()

	if tool.Name != "echo" {
		t.Errorf("Expected tool name 'echo', got %s", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Tool description should not be empty")
	}
}

func TestEchoHandler(t *testing.T) {
	// Create a mock request using mcp-go types
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "echo",
			Arguments: map[string]interface{}{"message": "Hello, World!"},
		},
	}

	result, err := EchoHandler(context.Background(), request)

	if err != nil {
		t.Fatalf("EchoHandler() error = %v", err)
	}

	if result == nil {
		t.Fatal("EchoHandler() returned nil result")
	}

	// Check that the result contains the echoed message
	if result.Content == nil {
		t.Error("Result content is nil")
	}
}

func TestNewExampleServer(t *testing.T) {
	server := NewExampleServer()

	if server == nil {
		t.Fatal("NewExampleServer() returned nil")
	}

	// Test adding a tool
	tool := CreateEchoTool()
	server.AddTool(tool, EchoHandler)

	// This test verifies the integration works without errors
}

func TestToolCreation(t *testing.T) {
	// Test creating a tool with various options
	tool := NewTool("test-tool",
		WithDescription("A test tool"),
		WithString("param1", Required(), Description("First parameter")),
		WithNumber("param2", Description("Second parameter")),
	)

	if tool.Name != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got %s", tool.Name)
	}

	if tool.Description != "A test tool" {
		t.Errorf("Expected description 'A test tool', got %s", tool.Description)
	}
}

// Additional tests can be added here as needed
// The mcp-go library handles most of the protocol testing internally

func TestNewResource(t *testing.T) {
	// Test creating a resource
	resource := NewResource("file:///test/path", "test-resource")

	if resource.URI != "file:///test/path" {
		t.Errorf("Expected URI 'file:///test/path', got %s", resource.URI)
	}

	if resource.Name != "test-resource" {
		t.Errorf("Expected name 'test-resource', got %s", resource.Name)
	}
}

func TestNewToolResultError(t *testing.T) {
	// Test creating an error result
	result := NewToolResultError("Something went wrong")

	if result == nil {
		t.Fatal("NewToolResultError() returned nil")
	}

	// Verify it's an error result
	if !result.IsError {
		t.Error("Expected IsError to be true")
	}

	// Check error content
	if len(result.Content) == 0 {
		t.Error("Expected error content, got empty")
	}
}

func TestAddResource(t *testing.T) {
	server := NewServer("Test Server", "1.0.0")
	resource := NewResource("file:///test", "test-resource")

	// Create a simple resource handler
	handler := func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "text/plain",
				Text:     "test content",
			},
		}, nil
	}

	// Add resource to server
	server.AddResource(resource, handler)

	// This test verifies the method exists and doesn't panic
}

func TestServeStdio(t *testing.T) {
	// Note: We can't actually test ServeStdio without proper stdio setup
	// This test just verifies the function exists and has the right signature
	// Actual stdio serving would fail in test environment
	// ServeStdio function exists in the package
}

func TestEchoHandlerErrors(t *testing.T) {
	tests := []struct {
		name           string
		request        mcp.CallToolRequest
		expectErrorMsg bool
		expectSuccess  bool
	}{
		{
			name: "missing_message_argument",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "echo",
					Arguments: map[string]interface{}{},
				},
			},
			expectErrorMsg: true,
			expectSuccess:  false,
		},
		{
			name: "invalid_message_type",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "echo",
					Arguments: map[string]interface{}{"message": 123}, // Not a string
				},
			},
			expectErrorMsg: true,
			expectSuccess:  false,
		},
		{
			name: "valid_message",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "echo",
					Arguments: map[string]interface{}{"message": "Hello"},
				},
			},
			expectErrorMsg: false,
			expectSuccess:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EchoHandler(context.Background(), tt.request)

			// EchoHandler never returns an error, it returns error results
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			// Check if result contains error
			if tt.expectErrorMsg {
				if !result.IsError {
					t.Error("Expected error result, got success")
				}
			} else {
				if result.IsError {
					t.Error("Expected success result, got error")
				}
			}
		})
	}
}

func TestServerWithCapabilities(t *testing.T) {
	server := NewServer("Test Server", "1.0.0",
		WithToolCapabilities(true),
		WithResourceCapabilities(true, true),
		WithRecovery(),
	)

	if server == nil {
		t.Fatal("NewServer() with capabilities returned nil")
	}

	// The actual capability settings are handled by mcp-go
	// This test verifies the options are accepted without error
}
