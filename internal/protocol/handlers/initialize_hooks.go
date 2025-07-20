// Package handlers provides MCP protocol handler implementations and hooks.
package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
)

// InitializeHooksConfig contains configuration for initialization hooks.
type InitializeHooksConfig struct {
	ConnectionManager *connection.Manager
	SupportedVersions []string
	ServerInfo        mcp.Implementation
}

// CreateInitializeHooks creates and returns initialization hooks for the MCP server.
func CreateInitializeHooks(config InitializeHooksConfig) (server.OnBeforeInitializeFunc, server.OnAfterInitializeFunc) {
	// Default supported versions if not specified
	if len(config.SupportedVersions) == 0 {
		config.SupportedVersions = []string{"1.0", "0.1.0"} // MCP protocol versions
	}

	// Before initialization hook
	beforeInit := func(ctx context.Context, id any, request *mcp.InitializeRequest) {
		log.Printf("[INIT] Before initialize hook triggered for request ID: %v", id)

		// Extract connection from context
		conn, ok := connection.ConnectionFromContext(ctx, config.ConnectionManager)
		if !ok {
			log.Printf("[INIT] Warning: No connection found in context for ID: %v", id)
			return
		}

		log.Printf("[INIT] Connection %s state before handshake: %s", conn.ID, conn.GetState())

		// Validate protocol version
		clientVersion := request.Params.ProtocolVersion
		if !isVersionSupported(clientVersion, config.SupportedVersions) {
			log.Printf("[INIT] Unsupported protocol version from client: %s", clientVersion)
			// Note: We can't return an error from this hook, so we'll handle it in the server response
		}

		// Log client info
		log.Printf("[INIT] Client info - Name: %s, Version: %s",
			request.Params.ClientInfo.Name,
			request.Params.ClientInfo.Version)

		// Log client capabilities
		logClientCapabilities(&request.Params.Capabilities)

		// Start handshake with timeout
		err := conn.StartHandshake(func() {
			log.Printf("[INIT] Handshake timeout for connection %s", conn.ID)
			config.ConnectionManager.RemoveConnection(conn.ID)
		})

		if err != nil {
			log.Printf("[INIT] Error starting handshake for connection %s: %v", conn.ID, err)
		}
	}

	// After initialization hook
	afterInit := func(ctx context.Context, id any, request *mcp.InitializeRequest, result *mcp.InitializeResult) {
		log.Printf("[INIT] After initialize hook triggered for request ID: %v", id)

		// Extract connection from context
		conn, ok := connection.ConnectionFromContext(ctx, config.ConnectionManager)
		if !ok {
			log.Printf("[INIT] Warning: No connection found in context after init for ID: %v", id)
			return
		}

		// Store client information
		clientInfo := make(map[string]interface{})
		clientInfo["name"] = request.Params.ClientInfo.Name
		clientInfo["version"] = request.Params.ClientInfo.Version

		// Store capabilities
		clientInfo["capabilities"] = request.Params.Capabilities

		// Complete handshake
		err := conn.CompleteHandshake(result.ProtocolVersion, clientInfo)
		if err != nil {
			log.Printf("[INIT] Error completing handshake for connection %s: %v", conn.ID, err)
			return
		}

		log.Printf("[INIT] Handshake completed successfully for connection %s", conn.ID)
		log.Printf("[INIT] Connection %s state after handshake: %s", conn.ID, conn.GetState())
		log.Printf("[INIT] Negotiated protocol version: %s", result.ProtocolVersion)

		// Log server capabilities that were sent
		logServerCapabilities(&result.Capabilities)
	}

	return beforeInit, afterInit
}

// isVersionSupported checks if the client version is supported by the server.
func isVersionSupported(clientVersion string, supportedVersions []string) bool {
	for _, v := range supportedVersions {
		if v == clientVersion {
			return true
		}
	}
	return false
}

// SelectProtocolVersion selects the best protocol version based on client and server support.
func SelectProtocolVersion(clientVersion string, supportedVersions []string) string {
	// If client version is supported, use it
	if isVersionSupported(clientVersion, supportedVersions) {
		return clientVersion
	}

	// Otherwise, return the first (highest priority) supported version
	if len(supportedVersions) > 0 {
		return supportedVersions[0]
	}

	// Fallback to a default version
	return "1.0"
}

// logClientCapabilities logs the client's capabilities for debugging.
func logClientCapabilities(caps *mcp.ClientCapabilities) {
	log.Printf("[INIT] Client capabilities:")

	if caps.Experimental != nil {
		log.Printf("[INIT]   - Experimental features: %+v", caps.Experimental)
	}

	if caps.Sampling != nil {
		log.Printf("[INIT]   - Sampling: %+v", caps.Sampling)
	}

	if caps.Roots != nil {
		log.Printf("[INIT]   - Roots:")
		if caps.Roots.ListChanged {
			log.Printf("[INIT]     - List changed notifications: enabled")
		}
	}
}

// logServerCapabilities logs the server's capabilities for debugging.
func logServerCapabilities(caps *mcp.ServerCapabilities) {
	log.Printf("[INIT] Server capabilities:")

	if caps.Experimental != nil {
		log.Printf("[INIT]   - Experimental features: %+v", caps.Experimental)
	}

	if caps.Logging != nil {
		log.Printf("[INIT]   - Logging: %+v", caps.Logging)
	}

	if caps.Prompts != nil {
		log.Printf("[INIT]   - Prompts:")
		if caps.Prompts.ListChanged {
			log.Printf("[INIT]     - List changed notifications: enabled")
		}
	}

	if caps.Resources != nil {
		log.Printf("[INIT]   - Resources:")
		if caps.Resources.Subscribe {
			log.Printf("[INIT]     - Subscribe: enabled")
		}
		if caps.Resources.ListChanged {
			log.Printf("[INIT]     - List changed notifications: enabled")
		}
	}

	if caps.Tools != nil {
		log.Printf("[INIT]   - Tools:")
		if caps.Tools.ListChanged {
			log.Printf("[INIT]     - List changed notifications: enabled")
		}
	}
}

// CreateInitializationError creates a proper JSON-RPC error for initialization failures.
func CreateInitializationError(message string) error {
	return fmt.Errorf("initialization failed: %s", message)
}
