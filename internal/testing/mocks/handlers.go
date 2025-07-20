// Package mocks provides mock implementations for testing
package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// MockHandler implements a mock JSON-RPC handler for testing
type MockHandler struct {
	mu           sync.RWMutex
	method       string
	response     *jsonrpc.Response
	error        error
	callCount    int
	lastRequest  *jsonrpc.Request
	callHistory  []*jsonrpc.Request
	delay        time.Duration
	shouldPanic  bool
	panicMessage string
}

// NewMockHandler creates a new mock handler
func NewMockHandler(method string) *MockHandler {
	return &MockHandler{
		method:      method,
		callHistory: make([]*jsonrpc.Request, 0),
	}
}

// Handle implements the Handler interface
func (m *MockHandler) Handle(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callCount++
	m.lastRequest = req
	m.callHistory = append(m.callHistory, req)
	
	if m.shouldPanic {
		panic(m.panicMessage)
	}
	
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	
	if m.error != nil {
		return &jsonrpc.Response{
			ID: req.ID,
			Error: &jsonrpc.Error{
				Code:    -32000,
				Message: m.error.Error(),
			},
		}
	}
	
	if m.response != nil {
		// Clone the response and set the correct ID
		resp := *m.response
		resp.ID = req.ID
		return &resp
	}
	
	// Default success response
	return &jsonrpc.Response{
		ID:     req.ID,
		Result: map[string]interface{}{"status": "success", "method": req.Method},
	}
}

// SetResponse sets the response that the handler should return
func (m *MockHandler) SetResponse(response *jsonrpc.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.response = response
}

// SetError sets an error that the handler should return
func (m *MockHandler) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.error = err
}

// SetDelay sets a delay for the handler response
func (m *MockHandler) SetDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delay = delay
}

// SetPanic configures the handler to panic
func (m *MockHandler) SetPanic(shouldPanic bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldPanic = shouldPanic
	m.panicMessage = message
}

// GetCallCount returns the number of times the handler was called
func (m *MockHandler) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount
}

// GetLastRequest returns the last request received by the handler
func (m *MockHandler) GetLastRequest() *jsonrpc.Request {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastRequest
}

// GetCallHistory returns all requests received by the handler
func (m *MockHandler) GetCallHistory() []*jsonrpc.Request {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy to prevent race conditions
	history := make([]*jsonrpc.Request, len(m.callHistory))
	copy(history, m.callHistory)
	return history
}

// Reset resets the handler state
func (m *MockHandler) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callCount = 0
	m.lastRequest = nil
	m.callHistory = make([]*jsonrpc.Request, 0)
	m.response = nil
	m.error = nil
	m.delay = 0
	m.shouldPanic = false
	m.panicMessage = ""
}

// MockHandlerFunc is a function-based mock handler
type MockHandlerFunc func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response

// Handle implements the Handler interface
func (f MockHandlerFunc) Handle(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	return f(ctx, req)
}

// EchoHandler returns a handler that echoes the request parameters
func EchoHandler() MockHandlerFunc {
	return func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return &jsonrpc.Response{
			ID:     req.ID,
			Result: req.Params,
		}
	}
}

// ErrorHandler returns a handler that always returns an error
func ErrorHandler(code int, message string) MockHandlerFunc {
	return func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return &jsonrpc.Response{
			ID: req.ID,
			Error: &jsonrpc.Error{
				Code:    code,
				Message: message,
			},
		}
	}
}

// DelayHandler returns a handler that introduces a delay
func DelayHandler(delay time.Duration, result interface{}) MockHandlerFunc {
	return func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		time.Sleep(delay)
		return &jsonrpc.Response{
			ID:     req.ID,
			Result: result,
		}
	}
}

// CountingHandler returns a handler that counts calls
func CountingHandler() (*MockHandlerFunc, func() int) {
	var count int
	var mu sync.Mutex
	
	handler := MockHandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		mu.Lock()
		count++
		currentCount := count
		mu.Unlock()
		
		return &jsonrpc.Response{
			ID:     req.ID,
			Result: map[string]interface{}{"count": currentCount},
		}
	})
	
	getCount := func() int {
		mu.Lock()
		defer mu.Unlock()
		return count
	}
	
	return &handler, getCount
}

// ConditionalHandler returns different responses based on request content
type ConditionalHandler struct {
	conditions map[string]MockHandlerFunc
	defaultHandler MockHandlerFunc
}

// NewConditionalHandler creates a new conditional handler
func NewConditionalHandler() *ConditionalHandler {
	return &ConditionalHandler{
		conditions: make(map[string]MockHandlerFunc),
		defaultHandler: func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			return &jsonrpc.Response{
				ID: req.ID,
				Error: &jsonrpc.Error{
					Code:    -32601,
					Message: "Method not found",
				},
			}
		},
	}
}

// AddCondition adds a condition-based handler
func (c *ConditionalHandler) AddCondition(method string, handler MockHandlerFunc) {
	c.conditions[method] = handler
}

// SetDefault sets the default handler
func (c *ConditionalHandler) SetDefault(handler MockHandlerFunc) {
	c.defaultHandler = handler
}

// Handle implements the Handler interface
func (c *ConditionalHandler) Handle(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	if handler, exists := c.conditions[req.Method]; exists {
		return handler(ctx, req)
	}
	return c.defaultHandler(ctx, req)
}

// MockHandlerRegistry manages multiple mock handlers
type MockHandlerRegistry struct {
	handlers map[string]*MockHandler
	mu       sync.RWMutex
}

// NewMockHandlerRegistry creates a new handler registry
func NewMockHandlerRegistry() *MockHandlerRegistry {
	return &MockHandlerRegistry{
		handlers: make(map[string]*MockHandler),
	}
}

// Register registers a mock handler for a method
func (r *MockHandlerRegistry) Register(method string, handler *MockHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[method] = handler
}

// Get retrieves a handler for a method
func (r *MockHandlerRegistry) Get(method string) (*MockHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, exists := r.handlers[method]
	return handler, exists
}

// GetOrCreate retrieves or creates a handler for a method
func (r *MockHandlerRegistry) GetOrCreate(method string) *MockHandler {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if handler, exists := r.handlers[method]; exists {
		return handler
	}
	
	handler := NewMockHandler(method)
	r.handlers[method] = handler
	return handler
}

// Reset resets all handlers in the registry
func (r *MockHandlerRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for _, handler := range r.handlers {
		handler.Reset()
	}
}

// GetMethods returns all registered methods
func (r *MockHandlerRegistry) GetMethods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	methods := make([]string, 0, len(r.handlers))
	for method := range r.handlers {
		methods = append(methods, method)
	}
	return methods
}

// CreateTestHandlers creates a set of common test handlers
func CreateTestHandlers() map[string]MockHandlerFunc {
	return map[string]MockHandlerFunc{
		"echo":    EchoHandler(),
		"error":   ErrorHandler(-32000, "Test error"),
		"slow":    DelayHandler(100*time.Millisecond, "slow response"),
		"success": func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			return &jsonrpc.Response{
				ID:     req.ID,
				Result: "success",
			}
		},
		"ping": func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			return &jsonrpc.Response{
				ID:     req.ID,
				Result: "pong",
			}
		},
	}
}

// ValidateHandlerCall validates that a handler was called with expected parameters
func ValidateHandlerCall(handler *MockHandler, expectedMethod string, expectedParams interface{}) error {
	if handler.GetCallCount() == 0 {
		return fmt.Errorf("handler was not called")
	}
	
	lastReq := handler.GetLastRequest()
	if lastReq == nil {
		return fmt.Errorf("no request recorded")
	}
	
	if lastReq.Method != expectedMethod {
		return fmt.Errorf("expected method %s, got %s", expectedMethod, lastReq.Method)
	}
	
	// Additional parameter validation could be added here
	
	return nil
}
