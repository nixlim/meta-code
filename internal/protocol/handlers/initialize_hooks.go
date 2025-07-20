// Package handlers provides MCP protocol handler implementations and hooks.
package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/meta-mcp/meta-mcp-server/internal/logging"
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

	logger := logging.Default().WithComponent("init")
	
	// Store request data for use in afterInit
	var requestData struct {
		mu sync.Mutex
		requests map[any]*mcp.InitializeRequest
	}
	requestData.requests = make(map[any]*mcp.InitializeRequest)

	// Before initialization hook
	beforeInit := func(ctx context.Context, id any, request *mcp.InitializeRequest) {
		logger.WithField("request_id", id).Debug(ctx, "Before initialize hook triggered")
		
		// Store request for afterInit
		requestData.mu.Lock()
		requestData.requests[id] = request
		requestData.mu.Unlock()

		// Get connection from context
		connID, ok := connection.GetConnectionID(ctx)
		if !ok {
			logger.WithField("request_id", id).Warn(ctx, "No connection found in context")
			return
		}

		conn, exists := config.ConnectionManager.GetConnection(connID)
		if !exists {
			logger.WithField("request_id", id).Warn(ctx, "Connection not found")
			return
		}
		
		logger.WithFields(logging.LogFields{
			logging.FieldConnectionID: conn.ID,
			logging.FieldConnectionState: conn.GetState().String(),
		}).Debug(ctx, "Connection state before handshake")

		// Validate protocol version
		clientVersion := request.Params.ProtocolVersion
		if !isVersionSupported(clientVersion, config.SupportedVersions) {
			logger.WithField(logging.FieldProtocolVersion, clientVersion).
				Error(ctx, nil, "Unsupported protocol version from client")
			// In a real implementation, we'd reject the request here
			return
		}

		logger.WithFields(logging.LogFields{
			logging.FieldClientName: request.Params.ClientInfo.Name,
			logging.FieldVersion: request.Params.ClientInfo.Version,
		}).Info(ctx, "Client info")

		// Start handshake
		timeoutCallback := func() {
			logger.WithField(logging.FieldConnectionID, conn.ID).
				Warn(ctx, "Handshake timeout")
		}

		if err := conn.StartHandshake(timeoutCallback); err != nil {
			logger.WithField(logging.FieldConnectionID, conn.ID).
				Error(ctx, err, "Error starting handshake")
		}
	}

	// After initialization hook
	afterInit := func(ctx context.Context, id any, message *mcp.InitializeRequest, result *mcp.InitializeResult) {
		logger.WithField("request_id", id).Debug(ctx, "After initialize hook triggered")

		// Get connection from context
		connID, ok := connection.GetConnectionID(ctx)
		if !ok {
			logger.WithField("request_id", id).Warn(ctx, "No connection found in context after init")
			return
		}

		conn, exists := config.ConnectionManager.GetConnection(connID)
		if !exists {
			return
		}

		// Clean up stored request
		requestData.mu.Lock()
		delete(requestData.requests, id)
		requestData.mu.Unlock()
		
		// Prepare client info for handshake completion
		clientInfo := make(map[string]interface{})
		if message != nil {
			clientInfo["name"] = message.Params.ClientInfo.Name
			clientInfo["version"] = message.Params.ClientInfo.Version
		}

		// Complete handshake
		if err := conn.CompleteHandshake(result.ProtocolVersion, clientInfo); err != nil {
			logger.WithField(logging.FieldConnectionID, conn.ID).
				Error(ctx, err, "Error completing handshake")
			return
		}

		logger.WithFields(logging.LogFields{
			logging.FieldConnectionID: conn.ID,
			logging.FieldConnectionState: conn.GetState().String(),
			logging.FieldProtocolVersion: result.ProtocolVersion,
		}).Info(ctx, "Handshake completed successfully")

		// Log capabilities if needed for debugging
		if message != nil {
			logClientCapabilities(ctx, logger, &message.Params.Capabilities)
		}
		logServerCapabilities(ctx, logger, &result.Capabilities)
	}

	return beforeInit, afterInit
}

// isVersionSupported checks if the client version is supported.
func isVersionSupported(clientVersion string, supportedVersions []string) bool {
	for _, v := range supportedVersions {
		if v == clientVersion {
			return true
		}
	}
	return false
}

// validateVersionCompatibility ensures client and server can communicate.
func validateVersionCompatibility(clientVersion string, supportedVersions []string) error {
	if !isVersionSupported(clientVersion, supportedVersions) {
		return fmt.Errorf("unsupported protocol version: %s (supported: %v)",
			clientVersion, supportedVersions)
	}
	return nil
}

// logClientCapabilities logs detailed client capabilities for debugging.
func logClientCapabilities(ctx context.Context, logger *logging.Logger, caps *mcp.ClientCapabilities) {
	if caps == nil {
		return
	}

	logger.Debug(ctx, "Client capabilities:")

	if caps.Experimental != nil {
		logger.WithField("experimental", caps.Experimental).Debug(ctx, "  - Experimental features")
	}

	if caps.Sampling != nil {
		logger.WithField("sampling", caps.Sampling).Debug(ctx, "  - Sampling")
	}

	if caps.Roots != nil {
		logger.Debug(ctx, "  - Roots:")
		if caps.Roots.ListChanged {
			logger.Debug(ctx, "    - List changed notifications: enabled")
		}
	}
}

// logServerCapabilities logs detailed server capabilities for debugging.
func logServerCapabilities(ctx context.Context, logger *logging.Logger, caps *mcp.ServerCapabilities) {
	if caps == nil {
		return
	}

	logger.Debug(ctx, "Server capabilities:")

	if caps.Experimental != nil {
		logger.WithField("experimental", caps.Experimental).Debug(ctx, "  - Experimental features")
	}

	if caps.Logging != nil {
		logger.WithField("logging", caps.Logging).Debug(ctx, "  - Logging")
	}

	if caps.Prompts != nil {
		logger.Debug(ctx, "  - Prompts:")
		if caps.Prompts.ListChanged {
			logger.Debug(ctx, "    - List changed notifications: enabled")
		}
	}

	if caps.Resources != nil {
		logger.Debug(ctx, "  - Resources:")
		if caps.Resources.Subscribe {
			logger.Debug(ctx, "    - Subscribe: enabled")
		}
		if caps.Resources.ListChanged {
			logger.Debug(ctx, "    - List changed notifications: enabled")
		}
	}

	if caps.Tools != nil {
		logger.Debug(ctx, "  - Tools:")
		if caps.Tools.ListChanged {
			logger.Debug(ctx, "    - List changed notifications: enabled")
		}
	}
}