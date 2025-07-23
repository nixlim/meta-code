package mcp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/meta-mcp/meta-mcp-server/internal/logging"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/handlers"
)

// HandshakeConfig contains configuration for the handshake-enabled server.
type HandshakeConfig struct {
	Name              string
	Version           string
	HandshakeTimeout  time.Duration
	SupportedVersions []string
	ServerOptions     []server.ServerOption
}

// DefaultHandshakeConfig returns a default configuration.
func DefaultHandshakeConfig() HandshakeConfig {
	return HandshakeConfig{
		Name:              "Meta-MCP Server",
		Version:           "1.0.0",
		HandshakeTimeout:  30 * time.Second,
		SupportedVersions: []string{"1.0", "0.1.0"},
	}
}

// HandshakeServer extends Server with connection management and handshake capabilities.
type HandshakeServer struct {
	*Server
	connectionManager *connection.Manager
	config            HandshakeConfig
}

// NewHandshakeServer creates a new MCP server with handshake support.
func NewHandshakeServer(config HandshakeConfig) *HandshakeServer {
	// Create connection manager
	connManager := connection.NewManager(config.HandshakeTimeout)

	// Create handshake server instance first (needed for hooks)
	hs := &HandshakeServer{
		connectionManager: connManager,
		config:            config,
	}

	// Create hooks
	hooks := hs.createHooks()

	// Append WithHooks to server options
	options := append(config.ServerOptions, server.WithHooks(hooks))

	// Create base server with hooks
	baseServer := NewServer(config.Name, config.Version, options...)
	hs.Server = baseServer

	return hs
}

// createHooks creates and configures all hooks for handshake management.
func (hs *HandshakeServer) createHooks() *server.Hooks {
	hooks := &server.Hooks{}

	logger := logging.Default().WithComponent("handshake")
	logger.Debug(context.Background(), "Creating handshake hooks...")

	// Create initialization hooks
	beforeInit, afterInit := handlers.CreateInitializeHooks(handlers.InitializeHooksConfig{
		ConnectionManager: hs.connectionManager,
		SupportedVersions: hs.config.SupportedVersions,
		ServerInfo: mcp.Implementation{
			Name:    hs.config.Name,
			Version: hs.config.Version,
		},
	})

	// Create validation hooks
	beforeAny := handlers.CreateValidationHooks(handlers.ValidationHooksConfig{
		ConnectionManager: hs.connectionManager,
	})

	// Create error and success hooks
	errorHook := handlers.CreateErrorHook(handlers.ValidationHooksConfig{
		ConnectionManager: hs.connectionManager,
	})

	successHook := handlers.CreateSuccessHook(handlers.ValidationHooksConfig{
		ConnectionManager: hs.connectionManager,
	})

	// Register all hooks
	hooks.AddBeforeInitialize(beforeInit)
	hooks.AddAfterInitialize(afterInit)
	hooks.AddBeforeAny(beforeAny)
	hooks.AddOnError(errorHook)
	hooks.AddOnSuccess(successHook)

	logger.Debug(context.Background(), "Hooks registered successfully")

	return hooks
}

// registerHooks sets up all the necessary hooks for handshake management.
func (hs *HandshakeServer) registerHooks() {
	// This method is no longer needed as we pass hooks during server creation
	logger := logging.Default().WithComponent("handshake")
	logger.Debug(context.Background(), "Hooks configured during server creation")
}

// CreateConnection creates a new connection and returns a context with the connection ID.
func (hs *HandshakeServer) CreateConnection(ctx context.Context, connectionID string) (context.Context, error) {
	// Create connection in manager
	conn, err := hs.connectionManager.CreateConnection(connectionID)
	if err != nil {
		return ctx, err
	}

	logger := logging.Default().WithComponent("handshake")
	logger.WithFields(logging.LogFields{
		logging.FieldConnectionID: connectionID,
		"timeout":                 conn.HandshakeTimeout,
	}).Debug(ctx, "Created connection")

	// Add connection ID to context
	ctx = connection.WithConnectionID(ctx, connectionID)

	return ctx, nil
}

// CloseConnection closes a connection and cleans up resources.
func (hs *HandshakeServer) CloseConnection(connectionID string) {
	logger := logging.Default().WithComponent("handshake")
	logger.WithField(logging.FieldConnectionID, connectionID).Debug(context.Background(), "Closing connection")
	hs.connectionManager.RemoveConnection(connectionID)
}

// GetConnectionManager returns the connection manager for external use.
func (hs *HandshakeServer) GetConnectionManager() *connection.Manager {
	return hs.connectionManager
}

// ServeStdioWithHandshake starts the server with stdio transport and handshake support.
func ServeStdioWithHandshake(hs *HandshakeServer, opts ...server.StdioOption) error {
	// Generate a connection ID for stdio transport
	connectionID := "stdio-" + generateConnectionID()

	// Create connection context
	ctx := context.Background()
	ctx, err := hs.CreateConnection(ctx, connectionID)
	if err != nil {
		return err
	}

	// Ensure connection is cleaned up on exit
	defer hs.CloseConnection(connectionID)

	logger := logging.Default().WithComponent("handshake")
	logger.WithField(logging.FieldConnectionID, connectionID).Info(ctx, "Starting stdio server")

	// Start the server
	// Note: We need to pass the context with connection ID to the server
	// This might require modification of mcp-go or a custom stdio implementation
	return ServeStdio(hs.Server, opts...)
}

// HandleMessage processes a JSON-RPC message with handshake validation.
// This method enables request interception for pre-handshake validation.
func (hs *HandshakeServer) HandleMessage(ctx context.Context, message json.RawMessage) mcp.JSONRPCMessage {
	// Extract connection ID from context
	connID, ok := connection.GetConnectionID(ctx)
	if !ok {
		// No connection ID means no handshake validation
		logger := logging.Default().WithComponent("handshake")
		logger.Warn(ctx, "No connection ID in context, proceeding without validation")
		// Fall back to base server handling
		return hs.Server.HandleMessage(ctx, message)
	}

	// Get connection to check handshake state
	conn, exists := hs.connectionManager.GetConnection(connID)
	if !exists {
		logger := logging.Default().WithComponent("handshake")
		logger.WithField(logging.FieldConnectionID, connID).Error(ctx, nil, "Connection not found")
		// Return error response
		return mcp.NewJSONRPCError(mcp.RequestId{}, -32002, "Connection not found", nil)
	}

	// Parse the request to check method
	var req struct {
		Method string        `json:"method"`
		ID     mcp.RequestId `json:"id,omitempty"`
	}
	if err := json.Unmarshal(message, &req); err != nil {
		logger := logging.Default().WithComponent("handshake")
		logger.Error(ctx, err, "Error parsing request")
		// Return parse error
		return mcp.NewJSONRPCError(mcp.RequestId{}, mcp.PARSE_ERROR, "Parse error", nil)
	}

	// Check if connection is ready for non-initialize requests
	if req.Method != "initialize" && !conn.IsReady() {
		logger := logging.Default().WithComponent("handshake")
		logger.WithFields(logging.LogFields{
			logging.FieldMethod:          req.Method,
			logging.FieldConnectionID:    connID,
			logging.FieldConnectionState: "not_initialized",
		}).Warn(ctx, "Rejecting request - connection not initialized")
		// Return not initialized error with custom code
		return mcp.NewJSONRPCError(req.ID, ErrorCodeServerNotInitialized, "Not initialized",
			"Initialize handshake must be completed before other requests")
	}

	// Delegate to base server for actual handling
	return hs.Server.HandleMessage(ctx, message)
}

// generateConnectionID generates a unique connection ID.
func generateConnectionID() string {
	// Use timestamp with nanoseconds for uniqueness
	// In production, consider using UUID or similar
	return time.Now().Format("20060102-150405.000000000")
}

// WithHandshakeTimeout creates a server option for handshake timeout.
// Note: This is a placeholder - actual implementation depends on mcp-go's extensibility.
func WithHandshakeTimeout(timeout time.Duration) func(*HandshakeConfig) {
	return func(config *HandshakeConfig) {
		config.HandshakeTimeout = timeout
	}
}

// WithSupportedVersions sets the supported protocol versions.
func WithSupportedVersions(versions ...string) func(*HandshakeConfig) {
	return func(config *HandshakeConfig) {
		config.SupportedVersions = versions
	}
}
