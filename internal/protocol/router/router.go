package router

import (
	"context"
	"sync"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// Handler defines the interface for handling JSON-RPC requests
type Handler interface {
	// Handle processes a JSON-RPC request and returns a response
	Handle(ctx context.Context, request *jsonrpc.Request) *jsonrpc.Response
}

// HandlerFunc is a function type that implements Handler
type HandlerFunc func(ctx context.Context, request *jsonrpc.Request) *jsonrpc.Response

// Handle implements the Handler interface
func (f HandlerFunc) Handle(ctx context.Context, request *jsonrpc.Request) *jsonrpc.Response {
	return f(ctx, request)
}

// NotificationHandler defines the interface for handling JSON-RPC notifications
type NotificationHandler interface {
	// HandleNotification processes a JSON-RPC notification
	HandleNotification(ctx context.Context, notification *jsonrpc.Notification)
}

// NotificationHandlerFunc is a function type that implements NotificationHandler
type NotificationHandlerFunc func(ctx context.Context, notification *jsonrpc.Notification)

// HandleNotification implements the NotificationHandler interface
func (f NotificationHandlerFunc) HandleNotification(ctx context.Context, notification *jsonrpc.Notification) {
	f(ctx, notification)
}

// Router provides message routing for JSON-RPC requests and notifications
type Router struct {
	mu                         sync.RWMutex
	handlers                   map[string]Handler
	notificationHandlers       map[string]NotificationHandler
	defaultHandler             Handler
	defaultNotificationHandler NotificationHandler
}

// New creates a new Router instance
func New() *Router {
	return &Router{
		handlers:             make(map[string]Handler),
		notificationHandlers: make(map[string]NotificationHandler),
	}
}

// Register registers a handler for the specified method
func (r *Router) Register(method string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[method] = handler
}

// RegisterFunc registers a handler function for the specified method
func (r *Router) RegisterFunc(method string, handlerFunc HandlerFunc) {
	r.Register(method, handlerFunc)
}

// RegisterNotification registers a notification handler for the specified method
func (r *Router) RegisterNotification(method string, handler NotificationHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notificationHandlers[method] = handler
}

// RegisterNotificationFunc registers a notification handler function for the specified method
func (r *Router) RegisterNotificationFunc(method string, handlerFunc NotificationHandlerFunc) {
	r.RegisterNotification(method, handlerFunc)
}

// SetDefaultHandler sets a default handler for unregistered methods
func (r *Router) SetDefaultHandler(handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultHandler = handler
}

// SetDefaultNotificationHandler sets a default handler for unregistered notification methods
func (r *Router) SetDefaultNotificationHandler(handler NotificationHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultNotificationHandler = handler
}

// Handle routes a JSON-RPC request to the appropriate handler
func (r *Router) Handle(ctx context.Context, request *jsonrpc.Request) *jsonrpc.Response {
	r.mu.RLock()
	handler, exists := r.handlers[request.Method]
	defaultHandler := r.defaultHandler
	r.mu.RUnlock()

	if exists {
		return handler.Handle(ctx, request)
	}

	if defaultHandler != nil {
		return defaultHandler.Handle(ctx, request)
	}

	// Return method not found error
	return jsonrpc.NewErrorResponse(
		jsonrpc.NewMethodNotFoundError(request.Method),
		request.ID,
	)
}

// HandleNotification routes a JSON-RPC notification to the appropriate handler
func (r *Router) HandleNotification(ctx context.Context, notification *jsonrpc.Notification) {
	r.mu.RLock()
	handler, exists := r.notificationHandlers[notification.Method]
	defaultHandler := r.defaultNotificationHandler
	r.mu.RUnlock()

	if exists {
		handler.HandleNotification(ctx, notification)
		return
	}

	if defaultHandler != nil {
		defaultHandler.HandleNotification(ctx, notification)
		return
	}

	// Notifications don't return responses, so we silently ignore unknown methods
}

// GetRegisteredMethods returns a list of all registered method names
func (r *Router) GetRegisteredMethods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	methods := make([]string, 0, len(r.handlers))
	for method := range r.handlers {
		methods = append(methods, method)
	}
	return methods
}

// GetRegisteredNotificationMethods returns a list of all registered notification method names
func (r *Router) GetRegisteredNotificationMethods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	methods := make([]string, 0, len(r.notificationHandlers))
	for method := range r.notificationHandlers {
		methods = append(methods, method)
	}
	return methods
}

// HasMethod checks if a method is registered
func (r *Router) HasMethod(method string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.handlers[method]
	return exists
}

// HasNotificationMethod checks if a notification method is registered
func (r *Router) HasNotificationMethod(method string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.notificationHandlers[method]
	return exists
}

// Unregister removes a handler for the specified method
func (r *Router) Unregister(method string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.handlers, method)
}

// UnregisterNotification removes a notification handler for the specified method
func (r *Router) UnregisterNotification(method string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.notificationHandlers, method)
}

// Clear removes all registered handlers
func (r *Router) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers = make(map[string]Handler)
	r.notificationHandlers = make(map[string]NotificationHandler)
	r.defaultHandler = nil
	r.defaultNotificationHandler = nil
}

// Stats returns statistics about the router
type Stats struct {
	RegisteredMethods             int
	RegisteredNotificationMethods int
	HasDefaultHandler             bool
	HasDefaultNotificationHandler bool
}

// GetStats returns router statistics
func (r *Router) GetStats() Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return Stats{
		RegisteredMethods:             len(r.handlers),
		RegisteredNotificationMethods: len(r.notificationHandlers),
		HasDefaultHandler:             r.defaultHandler != nil,
		HasDefaultNotificationHandler: r.defaultNotificationHandler != nil,
	}
}
