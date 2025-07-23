package testutil

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// CreateTestInitializeRequest creates a test initialize request.
func CreateTestInitializeRequest(version string, clientName string) *mcp.InitializeRequest {
	return &mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: version,
			ClientInfo: mcp.Implementation{
				Name:    clientName,
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	}
}

// CreateTestInitializeResult creates a test initialize result.
func CreateTestInitializeResult(version string, serverName string) *mcp.InitializeResult {
	return &mcp.InitializeResult{
		ProtocolVersion: version,
		ServerInfo: mcp.Implementation{
			Name:    serverName,
			Version: "1.0.0",
		},
		Capabilities: mcp.ServerCapabilities{},
	}
}

// CreateTestCallToolRequest creates a test call tool request.
func CreateTestCallToolRequest(toolName string, args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}
}

// CreateTestReadResourceRequest creates a test read resource request.
func CreateTestReadResourceRequest(uri string) mcp.ReadResourceRequest {
	return mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: uri,
		},
	}
}

// CreateTestClientCapabilities creates test client capabilities with common configurations.
func CreateTestClientCapabilities(withRoots bool, withSampling bool) mcp.ClientCapabilities {
	caps := mcp.ClientCapabilities{}
	
	if withRoots {
		caps.Roots = &struct {
			ListChanged bool `json:"listChanged,omitempty"`
		}{
			ListChanged: true,
		}
	}
	
	if withSampling {
		caps.Sampling = &struct{}{}
	}
	
	return caps
}

// CreateTestServerCapabilities creates test server capabilities with common configurations.
func CreateTestServerCapabilities(withTools bool, withResources bool) mcp.ServerCapabilities {
	caps := mcp.ServerCapabilities{}
	
	if withTools {
		caps.Tools = &struct {
			ListChanged bool `json:"listChanged,omitempty"`
		}{
			ListChanged: true,
		}
	}
	
	if withResources {
		caps.Resources = &struct {
			Subscribe   bool `json:"subscribe,omitempty"`
			ListChanged bool `json:"listChanged,omitempty"`
		}{
			Subscribe:   true,
			ListChanged: true,
		}
	}
	
	return caps
}