package handlers

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/meta-mcp/meta-mcp-server/internal/logging"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// ValidationHooksConfig contains configuration for validation hooks.
type ValidationHooksConfig struct {
	ConnectionManager *connection.Manager
}

// CreateValidationHooks creates hooks for validating requests based on connection state.
func CreateValidationHooks(config ValidationHooksConfig) server.BeforeAnyHookFunc {
	logger := logging.Default().WithComponent("validation")

	return func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		logger.WithFields(logging.LogFields{
			logging.FieldMethod: string(method),
			"id":                id,
		}).Debug(ctx, "BeforeAny hook triggered")

		// Always allow initialize method
		if method == mcp.MethodInitialize {
			logger.Debug(ctx, "Allowing initialize method")
			return
		}

		// Always allow notifications (they don't require responses)
		if isNotification(id) {
			logger.Debug(ctx, "Allowing notification (no ID)")
			return
		}

		// Extract connection from context
		conn, ok := connection.ConnectionFromContext(ctx, config.ConnectionManager)
		if !ok {
			logger.Warn(ctx, "No connection found in context for validation")
			// In a real implementation, we might want to reject the request here
			// For now, we'll log and continue
			return
		}

		// Check if connection is ready
		if !conn.IsReady() {
			state := conn.GetState()
			logger.WithFields(logging.LogFields{
				logging.FieldMethod:          string(method),
				logging.FieldConnectionID:    conn.ID,
				logging.FieldConnectionState: state.String(),
			}).Warn(ctx, "Rejecting method - connection not ready")

			// Note: The BeforeAny hook doesn't allow us to return an error directly.
			// In a real implementation, we would need to either:
			// 1. Store the error in context for the handler to check
			// 2. Use a different mechanism to reject the request
			// 3. Modify the mcp-go library to support error returns from this hook

			// For now, we'll log the rejection. The actual error response would need
			// to be handled by the request handler or through a custom middleware layer.
		} else {
			logger.WithFields(logging.LogFields{
				logging.FieldMethod:       string(method),
				logging.FieldConnectionID: conn.ID,
			}).Debug(ctx, "Allowing method - connection ready")
		}
	}
}

// CreateRequestValidator creates a middleware function that validates requests.
// This can be used in conjunction with the router to enforce handshake requirements.
func CreateRequestValidator(manager *connection.Manager) func(ctx context.Context, method string) error {
	return func(ctx context.Context, method string) error {
		// Always allow initialize and initialized
		if method == "initialize" || method == "initialized" {
			return nil
		}

		// Get connection from context
		conn, ok := connection.ConnectionFromContext(ctx, manager)
		if !ok {
			return &jsonrpc.Error{
				Code:    jsonrpc.ErrorCodeInvalidRequest,
				Message: "No connection context found",
			}
		}

		// Check if handshake is complete
		if !conn.IsReady() {
			return &jsonrpc.Error{
				Code:    -32011, // ErrorCodeServerNotInitialized
				Message: "Connection not initialized",
				Data: map[string]interface{}{
					"state":  conn.GetState().String(),
					"method": method,
				},
			}
		}

		return nil
	}
}

// isNotification checks if a message is a notification (has no ID).
func isNotification(id any) bool {
	return id == nil
}

// CreateErrorHook creates an error hook that logs validation errors.
func CreateErrorHook(config ValidationHooksConfig) server.OnErrorHookFunc {
	return func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		logger := logging.Default().WithComponent("validation")
		logger.WithFields(logging.LogFields{
			logging.FieldMethod: string(method),
			"id":                id,
		}).Error(ctx, err, "Error in method")

		// Log connection state for debugging
		if conn, ok := connection.ConnectionFromContext(ctx, config.ConnectionManager); ok {
			logger.WithFields(logging.LogFields{
				logging.FieldConnectionID:    conn.ID,
				logging.FieldConnectionState: conn.GetState().String(),
			}).Debug(ctx, "Connection state during error")
		}
	}
}

// CreateSuccessHook creates a success hook that logs successful operations.
func CreateSuccessHook(config ValidationHooksConfig) server.OnSuccessHookFunc {
	return func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		// Only log non-routine methods to reduce noise
		if method != mcp.MethodPing {
			logger := logging.Default().WithComponent("validation")
			logger.WithFields(logging.LogFields{
				logging.FieldMethod: string(method),
				"id":                id,
			}).Debug(ctx, "Success for method")
		}
	}
}
