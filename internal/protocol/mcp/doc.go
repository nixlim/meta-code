// Package mcp provides types and utilities for the Model Context Protocol (MCP).
//
// The MCP package builds on top of the JSON-RPC 2.0 foundation to provide
// MCP-specific message types, protocol negotiation, and capability exchange.
//
// # Protocol Overview
//
// MCP is a JSON-RPC 2.0-based protocol that enables communication between
// LLM applications and data sources/tools. The protocol defines:
//
//   - Initialize/Initialized handshake for protocol negotiation
//   - Capability exchange between client and server
//   - Resource, tool, and prompt management
//   - Logging and notification mechanisms
//
// # Core Message Types
//
// The package provides the following core message types:
//
//   - InitializeRequest: Client's initialization request
//   - InitializeResponse: Server's initialization response
//   - InitializedNotification: Client's confirmation of initialization
//
// # Protocol Versions
//
// MCP uses date-based versioning (e.g., "2024-11-05"). The ProtocolVersion
// type provides comparison and validation methods for version handling.
//
// # Capabilities
//
// The Capabilities type defines what features are supported by clients
// and servers:
//
//   - Server capabilities: resources, tools, prompts, logging
//   - Client capabilities: roots, sampling, experimental features
//
// # Error Handling
//
// MCP-specific error codes extend the JSON-RPC error code space:
//
//   - Resource errors: -32001 to -32002
//   - Tool errors: -32003 to -32004
//   - Prompt errors: -32005
//   - Protocol errors: -32006 to -32010
//
// # Usage Example
//
//	// Create an initialize request
//	params := mcp.InitializeParams{
//		ProtocolVersion: mcp.ProtocolVersionLatest,
//		ClientInfo: mcp.ClientInfo{
//			Name:    "my-client",
//			Version: "1.0.0",
//		},
//		Capabilities: mcp.Capabilities{
//			Resources: &mcp.ResourcesCapability{
//				Subscribe: true,
//			},
//		},
//	}
//
//	request := mcp.NewInitializeRequest(params, "req-1")
//
//	// Validate the request
//	if err := request.Validate(); err != nil {
//		log.Fatal(err)
//	}
//
//	// Create a response
//	result := mcp.InitializeResult{
//		ProtocolVersion: mcp.ProtocolVersionLatest,
//		ServerInfo: mcp.ServerInfo{
//			Name:    "my-server",
//			Version: "1.0.0",
//		},
//		Capabilities: mcp.Capabilities{
//			Resources: &mcp.ResourcesCapability{
//				Subscribe:   true,
//				ListChanged: true,
//			},
//		},
//	}
//
//	response := mcp.NewInitializeResponse(result, "req-1")
//
// # Method Constants
//
// The package provides constants for all MCP method names to ensure
// type safety and prevent typos:
//
//   - MethodInitialize: "initialize"
//   - MethodListResources: "resources/list"
//   - MethodCallTool: "tools/call"
//   - etc.
//
// # Validation
//
// All message types provide Validate() methods that check:
//
//   - Protocol version validity
//   - Required field presence
//   - Field format correctness
//   - JSON-RPC compliance
//
// # Compatibility
//
// The IsCompatible function checks protocol version compatibility
// between clients and servers. Currently, exact version matching
// is required for date-based versions.
package mcp
