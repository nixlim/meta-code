// Package mcp provides mock MCP client and server utilities for testing.
package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// MockClient implements the MCPClient interface for testing purposes.
// It provides configurable responses, error injection, and call tracking.
type MockClient struct {
	mu sync.RWMutex

	// Configuration
	responses        map[string]interface{}        // Method-specific responses
	errors           map[string]error              // Method-specific errors
	delays           map[string]time.Duration      // Method-specific delays
	defaultDelay     time.Duration                 // Default delay for all methods
	notificationFunc func(mcp.JSONRPCNotification) // Notification handler

	// Call tracking
	calls      []CallRecord
	callCounts map[string]int

	// State
	initialized bool
	closed      bool
}

// CallRecord represents a single method call made to the mock client.
type CallRecord struct {
	Method    string
	Args      interface{}
	Timestamp time.Time
	Error     error
}

// NewMockClient creates a new mock MCP client with default configuration.
func NewMockClient() *MockClient {
	return &MockClient{
		responses:    make(map[string]interface{}),
		errors:       make(map[string]error),
		delays:       make(map[string]time.Duration),
		callCounts:   make(map[string]int),
		calls:        make([]CallRecord, 0),
		defaultDelay: 0,
	}
}

// SetResponse configures the response for a specific method.
func (m *MockClient) SetResponse(method string, response interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[method] = response
}

// SetError configures an error response for a specific method.
func (m *MockClient) SetError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method] = err
}

// SetDelay configures a delay for a specific method.
func (m *MockClient) SetDelay(method string, delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delays[method] = delay
}

// SetDefaultDelay sets a default delay for all methods.
func (m *MockClient) SetDefaultDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultDelay = delay
}

// GetCallCount returns the number of times a method was called.
func (m *MockClient) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCounts[method]
}

// GetCalls returns all recorded calls.
func (m *MockClient) GetCalls() []CallRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	calls := make([]CallRecord, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetCallsForMethod returns all calls for a specific method.
func (m *MockClient) GetCallsForMethod(method string) []CallRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var methodCalls []CallRecord
	for _, call := range m.calls {
		if call.Method == method {
			methodCalls = append(methodCalls, call)
		}
	}
	return methodCalls
}

// Reset clears all call records and counts.
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = make([]CallRecord, 0)
	m.callCounts = make(map[string]int)
	m.initialized = false
	m.closed = false
}

// recordCall records a method call.
func (m *MockClient) recordCall(method string, args interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if closed
	if m.closed {
		return fmt.Errorf("client is closed")
	}

	// Apply delay
	delay := m.defaultDelay
	if methodDelay, ok := m.delays[method]; ok {
		delay = methodDelay
	}
	if delay > 0 {
		time.Sleep(delay)
	}

	// Record the call
	call := CallRecord{
		Method:    method,
		Args:      args,
		Timestamp: time.Now(),
	}

	// Check for configured error
	if err, ok := m.errors[method]; ok {
		call.Error = err
		m.calls = append(m.calls, call)
		m.callCounts[method]++
		return err
	}

	m.calls = append(m.calls, call)
	m.callCounts[method]++
	return nil
}

// Initialize implements MCPClient.Initialize
func (m *MockClient) Initialize(ctx context.Context, request mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	if err := m.recordCall("Initialize", request); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.initialized = true
	m.mu.Unlock()

	// Return configured response or default
	if resp, ok := m.responses["Initialize"]; ok {
		if result, ok := resp.(*mcp.InitializeResult); ok {
			return result, nil
		}
	}

	// Default response
	return &mcp.InitializeResult{
		ProtocolVersion: "1.0",
		ServerInfo: mcp.Implementation{
			Name:    "Mock MCP Server",
			Version: "1.0.0",
		},
		Capabilities: mcp.ServerCapabilities{},
	}, nil
}

// Ping implements MCPClient.Ping
func (m *MockClient) Ping(ctx context.Context) error {
	return m.recordCall("Ping", nil)
}

// ListResourcesByPage implements MCPClient.ListResourcesByPage
func (m *MockClient) ListResourcesByPage(ctx context.Context, request mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	if err := m.recordCall("ListResourcesByPage", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["ListResourcesByPage"]; ok {
		if result, ok := resp.(*mcp.ListResourcesResult); ok {
			return result, nil
		}
	}

	return &mcp.ListResourcesResult{
		Resources: []mcp.Resource{},
	}, nil
}

// ListResources implements MCPClient.ListResources
func (m *MockClient) ListResources(ctx context.Context, request mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	if err := m.recordCall("ListResources", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["ListResources"]; ok {
		if result, ok := resp.(*mcp.ListResourcesResult); ok {
			return result, nil
		}
	}

	return &mcp.ListResourcesResult{
		Resources: []mcp.Resource{},
	}, nil
}

// ListResourceTemplatesByPage implements MCPClient.ListResourceTemplatesByPage
func (m *MockClient) ListResourceTemplatesByPage(ctx context.Context, request mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	if err := m.recordCall("ListResourceTemplatesByPage", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["ListResourceTemplatesByPage"]; ok {
		if result, ok := resp.(*mcp.ListResourceTemplatesResult); ok {
			return result, nil
		}
	}

	return &mcp.ListResourceTemplatesResult{
		ResourceTemplates: []mcp.ResourceTemplate{},
	}, nil
}

// ListResourceTemplates implements MCPClient.ListResourceTemplates
func (m *MockClient) ListResourceTemplates(ctx context.Context, request mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	if err := m.recordCall("ListResourceTemplates", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["ListResourceTemplates"]; ok {
		if result, ok := resp.(*mcp.ListResourceTemplatesResult); ok {
			return result, nil
		}
	}

	return &mcp.ListResourceTemplatesResult{
		ResourceTemplates: []mcp.ResourceTemplate{},
	}, nil
}

// ReadResource implements MCPClient.ReadResource
func (m *MockClient) ReadResource(ctx context.Context, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if err := m.recordCall("ReadResource", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["ReadResource"]; ok {
		if result, ok := resp.(*mcp.ReadResourceResult); ok {
			return result, nil
		}
	}

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContents{},
	}, nil
}

// Subscribe implements MCPClient.Subscribe
func (m *MockClient) Subscribe(ctx context.Context, request mcp.SubscribeRequest) error {
	return m.recordCall("Subscribe", request)
}

// Unsubscribe implements MCPClient.Unsubscribe
func (m *MockClient) Unsubscribe(ctx context.Context, request mcp.UnsubscribeRequest) error {
	return m.recordCall("Unsubscribe", request)
}

// ListPromptsByPage implements MCPClient.ListPromptsByPage
func (m *MockClient) ListPromptsByPage(ctx context.Context, request mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	if err := m.recordCall("ListPromptsByPage", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["ListPromptsByPage"]; ok {
		if result, ok := resp.(*mcp.ListPromptsResult); ok {
			return result, nil
		}
	}

	return &mcp.ListPromptsResult{
		Prompts: []mcp.Prompt{},
	}, nil
}

// ListPrompts implements MCPClient.ListPrompts
func (m *MockClient) ListPrompts(ctx context.Context, request mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	if err := m.recordCall("ListPrompts", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["ListPrompts"]; ok {
		if result, ok := resp.(*mcp.ListPromptsResult); ok {
			return result, nil
		}
	}

	return &mcp.ListPromptsResult{
		Prompts: []mcp.Prompt{},
	}, nil
}

// GetPrompt implements MCPClient.GetPrompt
func (m *MockClient) GetPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	if err := m.recordCall("GetPrompt", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["GetPrompt"]; ok {
		if result, ok := resp.(*mcp.GetPromptResult); ok {
			return result, nil
		}
	}

	return &mcp.GetPromptResult{}, nil
}

// ListToolsByPage implements MCPClient.ListToolsByPage
func (m *MockClient) ListToolsByPage(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	if err := m.recordCall("ListToolsByPage", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["ListToolsByPage"]; ok {
		if result, ok := resp.(*mcp.ListToolsResult); ok {
			return result, nil
		}
	}

	return &mcp.ListToolsResult{
		Tools: []mcp.Tool{},
	}, nil
}

// ListTools implements MCPClient.ListTools
func (m *MockClient) ListTools(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	if err := m.recordCall("ListTools", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["ListTools"]; ok {
		if result, ok := resp.(*mcp.ListToolsResult); ok {
			return result, nil
		}
	}

	return &mcp.ListToolsResult{
		Tools: []mcp.Tool{},
	}, nil
}

// CallTool implements MCPClient.CallTool
func (m *MockClient) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := m.recordCall("CallTool", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["CallTool"]; ok {
		if result, ok := resp.(*mcp.CallToolResult); ok {
			return result, nil
		}
	}

	return &mcp.CallToolResult{}, nil
}

// SetLevel implements MCPClient.SetLevel
func (m *MockClient) SetLevel(ctx context.Context, request mcp.SetLevelRequest) error {
	return m.recordCall("SetLevel", request)
}

// Complete implements MCPClient.Complete
func (m *MockClient) Complete(ctx context.Context, request mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	if err := m.recordCall("Complete", request); err != nil {
		return nil, err
	}

	if resp, ok := m.responses["Complete"]; ok {
		if result, ok := resp.(*mcp.CompleteResult); ok {
			return result, nil
		}
	}

	result := &mcp.CompleteResult{}
	result.Completion.Values = []string{}
	return result, nil
}

// Close implements MCPClient.Close
func (m *MockClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("client already closed")
	}

	m.closed = true
	return nil
}

// OnNotification implements MCPClient.OnNotification
func (m *MockClient) OnNotification(handler func(notification mcp.JSONRPCNotification)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notificationFunc = handler
}

// SendNotification simulates sending a notification to the registered handler.
func (m *MockClient) SendNotification(notification mcp.JSONRPCNotification) {
	m.mu.RLock()
	handler := m.notificationFunc
	m.mu.RUnlock()

	if handler != nil {
		handler(notification)
	}
}

// IsInitialized returns whether the client has been initialized.
func (m *MockClient) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// IsClosed returns whether the client has been closed.
func (m *MockClient) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closed
}
