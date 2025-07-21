package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
	mcpserver "github.com/meta-mcp/meta-mcp-server/internal/protocol/mcp"
)

// MockServerConfig provides configuration options for the mock server.
type MockServerConfig struct {
	// Server identification
	Name    string
	Version string

	// Behavior configuration
	HandshakeTimeout  time.Duration
	SupportedVersions []string
	ServerOptions     []server.ServerOption

	// Response configuration
	ResponseDelay time.Duration
	ErrorRate     float64 // Probability of returning an error (0.0-1.0)

	// Custom handlers
	InitializeHandler func(ctx context.Context, req mcp.InitializeRequest) (*mcp.InitializeResult, error)
	ToolHandlers      map[string]func(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error)
	ResourceHandlers  map[string]func(ctx context.Context) (*mcp.Resource, error)
}

// DefaultMockServerConfig returns a default configuration for the mock server.
func DefaultMockServerConfig() MockServerConfig {
	return MockServerConfig{
		Name:              "Mock MCP Server",
		Version:           "1.0.0",
		HandshakeTimeout:  30 * time.Second,
		SupportedVersions: []string{"1.0", "0.1.0"},
		ResponseDelay:     0,
		ErrorRate:         0,
		ToolHandlers:      make(map[string]func(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error)),
		ResourceHandlers:  make(map[string]func(ctx context.Context) (*mcp.Resource, error)),
	}
}

// MockServer wraps a HandshakeServer with additional testing capabilities.
type MockServer struct {
	*mcpserver.HandshakeServer
	config MockServerConfig

	mu sync.RWMutex
	// Request tracking
	requests      []RequestRecord
	requestCounts map[string]int

	// State tracking
	connections map[string]*ConnectionState
}

// RequestRecord represents a recorded request to the server.
type RequestRecord struct {
	Method    string
	Params    interface{}
	RequestID interface{}
	Timestamp time.Time
	Response  interface{}
	Error     error
}

// ConnectionState tracks the state of a connection.
type ConnectionState struct {
	ID        string
	State     string
	StartTime time.Time
	LastSeen  time.Time
	Metadata  map[string]interface{}
}

// NewMockServer creates a new mock MCP server with the given configuration.
func NewMockServer(config MockServerConfig) *MockServer {
	// Create HandshakeConfig
	hsConfig := mcpserver.HandshakeConfig{
		Name:              config.Name,
		Version:           config.Version,
		HandshakeTimeout:  config.HandshakeTimeout,
		SupportedVersions: config.SupportedVersions,
		ServerOptions:     config.ServerOptions,
	}

	// Create base server
	hs := mcpserver.NewHandshakeServer(hsConfig)

	// Create mock server
	ms := &MockServer{
		HandshakeServer: hs,
		config:          config,
		requests:        make([]RequestRecord, 0),
		requestCounts:   make(map[string]int),
		connections:     make(map[string]*ConnectionState),
	}

	// Override handlers if custom ones are provided
	// This would require modifying the HandshakeServer to support custom handlers
	// For now, we'll track the default behavior

	return ms
}

// HandleRequest processes a JSON-RPC request and returns a response.
// This wraps the HandshakeServer's HandleMessage with additional tracking.
func (ms *MockServer) HandleRequest(ctx context.Context, connID string, request []byte) ([]byte, error) {
	// Ensure connection exists
	ctx, err := ms.CreateConnection(ctx, connID)
	if err != nil {
		// Connection might already exist
		ctx = ms.GetConnectionContext(ctx, connID)
	}

	// Parse request to track it
	var req jsonrpc.Request
	if err := json.Unmarshal(request, &req); err == nil {
		ms.recordRequest(req.Method, req.Params, req.ID)
	}

	// Apply configured delay
	if ms.config.ResponseDelay > 0 {
		time.Sleep(ms.config.ResponseDelay)
	}

	// Handle through base server
	response := ms.HandleMessage(ctx, request)

	// Convert response to bytes
	respBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	// Track response
	if req.Method != "" {
		ms.updateRequestRecord(req.ID, response, nil)
	}

	return respBytes, nil
}

// GetRequests returns all recorded requests.
func (ms *MockServer) GetRequests() []RequestRecord {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	requests := make([]RequestRecord, len(ms.requests))
	copy(requests, ms.requests)
	return requests
}

// GetRequestCount returns the number of times a method was called.
func (ms *MockServer) GetRequestCount(method string) int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.requestCounts[method]
}

// GetConnectionState returns the state of a specific connection.
func (ms *MockServer) GetConnectionState(connID string) (*ConnectionState, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	state, ok := ms.connections[connID]
	return state, ok
}

// Reset clears all recorded data.
func (ms *MockServer) Reset() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.requests = make([]RequestRecord, 0)
	ms.requestCounts = make(map[string]int)
	ms.connections = make(map[string]*ConnectionState)
}

// recordRequest records an incoming request.
func (ms *MockServer) recordRequest(method string, params interface{}, id interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	record := RequestRecord{
		Method:    method,
		Params:    params,
		RequestID: id,
		Timestamp: time.Now(),
	}

	ms.requests = append(ms.requests, record)
	ms.requestCounts[method]++
}

// updateRequestRecord updates a request record with the response.
func (ms *MockServer) updateRequestRecord(id interface{}, response interface{}, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find the request by ID
	for i := len(ms.requests) - 1; i >= 0; i-- {
		if ms.requests[i].RequestID == id {
			ms.requests[i].Response = response
			ms.requests[i].Error = err
			break
		}
	}
}

// GetConnectionContext gets or creates a context for a connection.
func (ms *MockServer) GetConnectionContext(ctx context.Context, connID string) context.Context {
	// Try to get existing connection
	if _, ok := ms.GetConnectionManager().GetConnection(connID); ok {
		// Connection exists, return the context we have
		return ctx
	}

	// Create new connection
	ctx, _ = ms.CreateConnection(ctx, connID)
	return ctx
}

// SimulateClientMessage simulates receiving a message from a client.
func (ms *MockServer) SimulateClientMessage(ctx context.Context, connID string, method string, params interface{}, id interface{}) (interface{}, error) {
	// Create JSON-RPC request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      id,
	}

	// Marshal to bytes
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Handle request
	respBytes, err := ms.HandleRequest(ctx, connID, reqBytes)
	if err != nil {
		return nil, err
	}

	// Unmarshal response
	var response jsonrpc.Response
	if err := json.Unmarshal(respBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("error response: %s", response.Error.Message)
	}

	return response.Result, nil
}

// TestScenario represents a predefined test scenario.
type TestScenario struct {
	Name        string
	Description string
	Steps       []TestStep
}

// TestStep represents a single step in a test scenario.
type TestStep struct {
	Action      string        // "request", "wait", "check"
	Method      string        // For request actions
	Params      interface{}   // For request actions
	ExpectError bool          // For request actions
	Duration    time.Duration // For wait actions
	Check       func() error  // For check actions
}

// RunScenario executes a test scenario against the mock server.
func (ms *MockServer) RunScenario(ctx context.Context, connID string, scenario TestScenario) error {
	for i, step := range scenario.Steps {
		switch step.Action {
		case "request":
			_, err := ms.SimulateClientMessage(ctx, connID, step.Method, step.Params, fmt.Sprintf("%s-step-%d", scenario.Name, i))
			if step.ExpectError && err == nil {
				return fmt.Errorf("step %d: expected error but got none", i)
			}
			if !step.ExpectError && err != nil {
				return fmt.Errorf("step %d: unexpected error: %w", i, err)
			}

		case "wait":
			time.Sleep(step.Duration)

		case "check":
			if step.Check != nil {
				if err := step.Check(); err != nil {
					return fmt.Errorf("step %d: check failed: %w", i, err)
				}
			}

		default:
			return fmt.Errorf("unknown action: %s", step.Action)
		}
	}

	return nil
}

// CommonScenarios provides predefined test scenarios.
var CommonScenarios = struct {
	BasicHandshake     TestScenario
	ToolDiscovery      TestScenario
	ResourceDiscovery  TestScenario
	ConcurrentRequests TestScenario
}{
	BasicHandshake: TestScenario{
		Name:        "basic-handshake",
		Description: "Basic initialization handshake",
		Steps: []TestStep{
			{
				Action: "request",
				Method: "initialize",
				Params: map[string]interface{}{
					"protocolVersion": "1.0",
					"clientInfo": map[string]interface{}{
						"name":    "Test Client",
						"version": "1.0.0",
					},
					"capabilities": map[string]interface{}{},
				},
				ExpectError: false,
			},
		},
	},
	ToolDiscovery: TestScenario{
		Name:        "tool-discovery",
		Description: "Discover available tools after handshake",
		Steps: []TestStep{
			{
				Action: "request",
				Method: "initialize",
				Params: map[string]interface{}{
					"protocolVersion": "1.0",
					"clientInfo": map[string]interface{}{
						"name":    "Test Client",
						"version": "1.0.0",
					},
					"capabilities": map[string]interface{}{},
				},
				ExpectError: false,
			},
			{
				Action:   "wait",
				Duration: 10 * time.Millisecond,
			},
			{
				Action:      "request",
				Method:      "tools/list",
				Params:      nil,
				ExpectError: false,
			},
		},
	},
}

// WaitForConnection waits for a connection to reach a specific state.
func (ms *MockServer) WaitForConnection(connID string, targetState string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if conn, ok := ms.GetConnectionManager().GetConnection(connID); ok {
			if conn.GetState().String() == targetState {
				return nil
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for connection %s to reach state %s", connID, targetState)
}
