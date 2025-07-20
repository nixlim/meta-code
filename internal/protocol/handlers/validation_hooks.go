package handlers

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// ValidationHooksConfig contains configuration for validation hooks.
type ValidationHooksConfig struct {
	ConnectionManager *connection.Manager
}

// CreateValidationHooks creates hooks for validating requests based on connection state.
func CreateValidationHooks(config ValidationHooksConfig) server.BeforeAnyHookFunc {
	return func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		log.Printf("[VALIDATION] BeforeAny hook triggered - Method: %s, ID: %v", method, id)

		// Always allow initialize method
		if method == mcp.MethodInitialize {
			log.Printf("[VALIDATION] Allowing initialize method")
			return
		}

		// Always allow notifications (they don't require responses)
		if isNotification(id) {
			log.Printf("[VALIDATION] Allowing notification (no ID)")
			return
		}

		// Extract connection from context
		conn, ok := connection.ConnectionFromContext(ctx, config.ConnectionManager)
		if !ok {
			log.Printf("[VALIDATION] Warning: No connection found in context for validation")
			// In a real implementation, we might want to reject the request here
			// For now, we'll log and continue
			return
		}

		// Check if connection is ready
		if !conn.IsReady() {
			state := conn.GetState()
			log.Printf("[VALIDATION] Rejecting method %s - connection %s not ready (state: %s)",
				method, conn.ID, state)

			// Note: The BeforeAny hook doesn't allow us to return an error directly.
			// In a real implementation, we would need to either:
			// 1. Store the error in context for the handler to check
			// 2. Use a different mechanism to reject the request
			// 3. Modify the mcp-go library to support error returns from this hook

			// For now, we'll log the rejection. The actual error response would need
			// to be handled by the request handler or through a custom middleware layer.
		} else {
			log.Printf("[VALIDATION] Allowing method %s - connection %s is ready", method, conn.ID)
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
				Code:    -32002, // Custom error code for "not initialized"
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
		log.Printf("[VALIDATION] Error in method %s (ID: %v): %v", method, id, err)

		// Log connection state for debugging
		if conn, ok := connection.ConnectionFromContext(ctx, config.ConnectionManager); ok {
			log.Printf("[VALIDATION] Connection %s state during error: %s",
				conn.ID, conn.GetState())
		}
	}
}

// CreateSuccessHook creates a success hook that logs successful operations.
func CreateSuccessHook(config ValidationHooksConfig) server.OnSuccessHookFunc {
	return func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		// Only log non-routine methods to reduce noise
		if method != mcp.MethodPing {
			log.Printf("[VALIDATION] Success for method %s (ID: %v)", method, id)
		}
	}
}
