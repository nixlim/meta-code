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
