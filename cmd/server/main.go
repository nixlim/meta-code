package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/meta-mcp/meta-mcp-server/internal/logging"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/mcp"
)

func main() {
	// Initialize logger based on environment
	logConfig := logging.ConfigFromEnv()
	logger := logging.New(logConfig)
	logging.SetDefault(logger)
	
	// Create context with component information
	ctx := logging.WithComponent(context.Background(), "main")
	
	// Configure the handshake-enabled server
	config := mcp.HandshakeConfig{
		Name:              "Meta-MCP Server",
		Version:           "1.0.0",
		HandshakeTimeout:  30 * time.Second,
		SupportedVersions: []string{"1.0", "0.1.0"},
		ServerOptions: []server.ServerOption{
			mcp.WithToolCapabilities(true),
			mcp.WithResourceCapabilities(true, true),
			mcp.WithRecovery(),
		},
	}

	// Create a new handshake-enabled MCP server
	server := mcp.NewHandshakeServer(config)

	// Add an echo tool
	echoTool := mcp.CreateEchoTool()
	server.AddTool(echoTool, mcp.EchoHandler)

	// Add a calculator tool
	calculatorTool := mcp.NewTool("calculate",
		mcp.WithDescription("Perform basic arithmetic operations"),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The operation to perform (add, subtract, multiply, divide)"),
		),
		mcp.WithNumber("x",
			mcp.Required(),
			mcp.Description("First number"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("Second number"),
		),
	)

	server.AddTool(calculatorTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get operation parameter
		operation, err := request.RequireString("operation")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid operation: %v", err)), nil
		}

		// Get x parameter
		x, err := request.RequireFloat("x")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid x parameter: %v", err)), nil
		}

		// Get y parameter
		y, err := request.RequireFloat("y")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid y parameter: %v", err)), nil
		}

		// Perform calculation
		var result float64
		switch operation {
		case "add":
			result = x + y
		case "subtract":
			result = x - y
		case "multiply":
			result = x * y
		case "divide":
			if y == 0 {
				return mcp.NewToolResultError("Cannot divide by zero"), nil
			}
			result = x / y
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unknown operation: %s", operation)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
	})

	// Add a simple resource
	readmeResource := mcp.NewResource(
		"file://README.md",
		"Project README",
	)

	server.AddResource(readmeResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Read the README file
		content, err := os.ReadFile("README.md")
		if err != nil {
			// Return a default message if README doesn't exist
			content = []byte("# Meta-MCP Server\n\nA Model Context Protocol server implementation using mcp-go.")
		}

		// Create ResourceContents using the struct directly
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/markdown",
				Text:     string(content),
			},
		}, nil
	})

	// Start the server using stdio transport with handshake support
	logger.Info(ctx, "Starting Meta-MCP Server with handshake support...")
	logger.WithFields(logging.LogFields{
		"server_name": config.Name,
		"version": config.Version,
		"handshake_timeout": config.HandshakeTimeout,
	}).Info(ctx, "Server configuration loaded")
	
	if err := mcp.ServeStdioWithHandshake(server); err != nil {
		logger.Fatal(ctx, err, "Server error")
	}
}
