// Package router provides a message routing system for JSON-RPC requests and notifications.
//
// The router package implements a thread-safe message dispatcher that routes incoming
// JSON-RPC requests and notifications to registered handlers based on method names.
//
// # Core Components
//
// The package provides the following core components:
//
//   - Router: The main routing component that manages handler registration and dispatch
//   - Handler: Interface for handling JSON-RPC requests
//   - NotificationHandler: Interface for handling JSON-RPC notifications
//   - HandlerFunc/NotificationHandlerFunc: Function types implementing the interfaces
//
// # Basic Usage
//
//	// Create a new router
//	router := router.New()
//
//	// Register a handler for a specific method
//	router.RegisterFunc("echo", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
//		return jsonrpc.NewResponse(req.Params, req.ID)
//	})
//
//	// Register a notification handler
//	router.RegisterNotificationFunc("log", func(ctx context.Context, notif *jsonrpc.Notification) {
//		fmt.Printf("Log: %v\n", notif.Params)
//	})
//
//	// Handle a request
//	request := jsonrpc.NewRequest("echo", "hello", 1)
//	response := router.Handle(context.Background(), request)
//
//	// Handle a notification
//	notification := jsonrpc.NewNotification("log", "debug message")
//	router.HandleNotification(context.Background(), notification)
//
// # Handler Registration
//
// Handlers can be registered in several ways:
//
//   - Register(method, handler): Register a Handler interface implementation
//   - RegisterFunc(method, handlerFunc): Register a function as a handler
//   - RegisterNotification(method, handler): Register a NotificationHandler interface
//   - RegisterNotificationFunc(method, handlerFunc): Register a function as notification handler
//
// # Default Handlers
//
// The router supports default handlers for unregistered methods:
//
//	// Set a default handler for unknown request methods
//	router.SetDefaultHandler(defaultHandler)
//
//	// Set a default handler for unknown notification methods
//	router.SetDefaultNotificationHandler(defaultNotificationHandler)
//
// If no default handler is set, unknown request methods return a JSON-RPC
// "method not found" error, and unknown notifications are silently ignored.
//
// # Thread Safety
//
// The Router is thread-safe and can be used concurrently from multiple goroutines.
// All registration and handling operations are protected by read-write mutexes.
//
// # Error Handling
//
// For requests:
//   - Unknown methods return JSON-RPC "method not found" errors
//   - Handler errors should be returned as JSON-RPC error responses
//   - Request IDs are always preserved in responses
//
// For notifications:
//   - Unknown methods are silently ignored (no response)
//   - Handler errors are not returned (notifications are fire-and-forget)
//
// # Management Operations
//
// The router provides several management operations:
//
//   - GetRegisteredMethods(): List all registered request methods
//   - GetRegisteredNotificationMethods(): List all registered notification methods
//   - HasMethod(method): Check if a request method is registered
//   - HasNotificationMethod(method): Check if a notification method is registered
//   - Unregister(method): Remove a request handler
//   - UnregisterNotification(method): Remove a notification handler
//   - Clear(): Remove all handlers
//   - GetStats(): Get router statistics
//
// # Statistics
//
// The GetStats() method returns router statistics:
//
//	stats := router.GetStats()
//	fmt.Printf("Registered methods: %d\n", stats.RegisteredMethods)
//	fmt.Printf("Has default handler: %v\n", stats.HasDefaultHandler)
//
// # Integration with MCP
//
// This router is designed to work with the MCP protocol types and can be used
// to route MCP-specific methods like "initialize", "resources/list", etc.
//
//	import "github.com/meta-mcp/meta-mcp-server/internal/protocol/mcp"
//
//	// Register MCP initialize handler
//	router.RegisterFunc(mcp.MethodInitialize, handleInitialize)
//
//	// Register MCP notification handlers
//	router.RegisterNotificationFunc(mcp.MethodInitialized, handleInitialized)
package router
