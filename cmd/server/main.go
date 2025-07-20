package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/mcp"
)

func main() {
	// Create a new MCP server using the mcp-go library
	server := mcp.NewServer(
		"Meta-MCP Server",
		"1.0.0",
		mcp.WithToolCapabilities(true),
		mcp.WithResourceCapabilities(true, true),
		mcp.WithRecovery(),
	)

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

	// Start the server using stdio transport
	log.Println("Starting Meta-MCP Server...")
	if err := mcp.ServeStdio(server); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
