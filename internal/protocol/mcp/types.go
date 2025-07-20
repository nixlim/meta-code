package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the mcp-go server with additional functionality
type Server struct {
	*server.MCPServer
}

// NewServer creates a new MCP server using mcp-go
func NewServer(name, version string, options ...server.ServerOption) *Server {
	mcpServer := server.NewMCPServer(name, version, options...)

	return &Server{
		MCPServer: mcpServer,
	}
}

// Type aliases for convenience
type (
	Tool                 = mcp.Tool
	Resource             = mcp.Resource
	CallToolRequest      = mcp.CallToolRequest
	CallToolResult       = mcp.CallToolResult
	ReadResourceRequest  = mcp.ReadResourceRequest
	ResourceContents     = mcp.ResourceContents
	TextResourceContents = mcp.TextResourceContents
	ToolHandlerFunc      = server.ToolHandlerFunc
	ResourceHandlerFunc  = server.ResourceHandlerFunc
)

// Tool creation helpers that wrap mcp-go functions
func NewTool(name string, options ...mcp.ToolOption) mcp.Tool {
	return mcp.NewTool(name, options...)
}

// Resource creation helpers
func NewResource(uri, name string, options ...mcp.ResourceOption) mcp.Resource {
	return mcp.NewResource(uri, name, options...)
}

// Result creation helpers
func NewToolResultText(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

func NewToolResultError(message string) *mcp.CallToolResult {
	return mcp.NewToolResultError(message)
}

// Tool option helpers
func WithDescription(desc string) mcp.ToolOption {
	return mcp.WithDescription(desc)
}

func WithString(name string, options ...mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithString(name, options...)
}

func WithNumber(name string, options ...mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithNumber(name, options...)
}

func Required() mcp.PropertyOption {
	return mcp.Required()
}

func Description(desc string) mcp.PropertyOption {
	return mcp.Description(desc)
}

// Server methods that integrate with mcp-go
func (s *Server) AddTool(tool mcp.Tool, handler ToolHandlerFunc) {
	s.MCPServer.AddTool(tool, handler)
}

func (s *Server) AddResource(resource mcp.Resource, handler ResourceHandlerFunc) {
	s.MCPServer.AddResource(resource, handler)
}

// ServeStdio starts the server using stdio transport
func ServeStdio(s *Server, opts ...server.StdioOption) error {
	return server.ServeStdio(s.MCPServer, opts...)
}

// Server option helpers
func WithToolCapabilities(listChanged bool) server.ServerOption {
	return server.WithToolCapabilities(listChanged)
}

func WithResourceCapabilities(subscribe, listChanged bool) server.ServerOption {
	return server.WithResourceCapabilities(subscribe, listChanged)
}

func WithRecovery() server.ServerOption {
	return server.WithRecovery()
}

// Example server creation function
func NewExampleServer() *Server {
	return NewServer(
		"Meta-MCP Server",
		"1.0.0",
		WithToolCapabilities(true),
		WithResourceCapabilities(true, true),
		WithRecovery(),
	)
}

// Utility functions for common operations
func CreateEchoTool() mcp.Tool {
	return NewTool("echo",
		WithDescription("Echo back the input message"),
		WithString("message",
			Required(),
			Description("Message to echo back"),
		),
	)
}

func EchoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message, err := request.RequireString("message")
	if err != nil {
		return NewToolResultError(fmt.Sprintf("Invalid message parameter: %v", err)), nil
	}

	return NewToolResultText(fmt.Sprintf("Echo: %s", message)), nil
}

// Additional utility functions can be added here as needed
// The mcp-go library handles most of the protocol details automatically
