// Package helpers provides test builders for creating test data
package helpers

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// RequestBuilder builds JSON-RPC request messages for testing
type RequestBuilder struct {
	jsonrpc string
	method  string
	params  interface{}
	id      interface{}
	t       *testing.T
}

// NewRequestBuilder creates a new request builder
func NewRequestBuilder(t *testing.T) *RequestBuilder {
	return &RequestBuilder{
		jsonrpc: "2.0",
		t:       t,
	}
}

// WithMethod sets the method name
func (rb *RequestBuilder) WithMethod(method string) *RequestBuilder {
	rb.method = method
	return rb
}

// WithParams sets the parameters
func (rb *RequestBuilder) WithParams(params interface{}) *RequestBuilder {
	rb.params = params
	return rb
}

// WithID sets the request ID
func (rb *RequestBuilder) WithID(id interface{}) *RequestBuilder {
	rb.id = id
	return rb
}

// WithJSONRPC sets the JSON-RPC version
func (rb *RequestBuilder) WithJSONRPC(version string) *RequestBuilder {
	rb.jsonrpc = version
	return rb
}

// Build creates the JSON-RPC request
func (rb *RequestBuilder) Build() map[string]interface{} {
	req := map[string]interface{}{
		"jsonrpc": rb.jsonrpc,
		"method":  rb.method,
		"id":      rb.id,
	}

	if rb.params != nil {
		req["params"] = rb.params
	}

	return req
}

// BuildJSON creates the JSON-RPC request as JSON bytes
func (rb *RequestBuilder) BuildJSON() []byte {
	rb.t.Helper()

	req := rb.Build()
	data, err := json.Marshal(req)
	require.NoError(rb.t, err, "Failed to marshal request")

	return data
}

// BuildString creates the JSON-RPC request as a JSON string
func (rb *RequestBuilder) BuildString() string {
	return string(rb.BuildJSON())
}

// ResponseBuilder builds JSON-RPC response messages for testing
type ResponseBuilder struct {
	jsonrpc string
	result  interface{}
	error   *ErrorObject
	id      interface{}
	t       *testing.T
}

// ErrorObject represents a JSON-RPC error
type ErrorObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder(t *testing.T) *ResponseBuilder {
	return &ResponseBuilder{
		jsonrpc: "2.0",
		t:       t,
	}
}

// WithResult sets the result
func (rb *ResponseBuilder) WithResult(result interface{}) *ResponseBuilder {
	rb.result = result
	rb.error = nil // Clear error if setting result
	return rb
}

// WithError sets the error
func (rb *ResponseBuilder) WithError(code int, message string, data interface{}) *ResponseBuilder {
	rb.error = &ErrorObject{
		Code:    code,
		Message: message,
		Data:    data,
	}
	rb.result = nil // Clear result if setting error
	return rb
}

// WithID sets the response ID
func (rb *ResponseBuilder) WithID(id interface{}) *ResponseBuilder {
	rb.id = id
	return rb
}

// WithJSONRPC sets the JSON-RPC version
func (rb *ResponseBuilder) WithJSONRPC(version string) *ResponseBuilder {
	rb.jsonrpc = version
	return rb
}

// Build creates the JSON-RPC response
func (rb *ResponseBuilder) Build() map[string]interface{} {
	resp := map[string]interface{}{
		"jsonrpc": rb.jsonrpc,
		"id":      rb.id,
	}

	if rb.error != nil {
		resp["error"] = rb.error
	} else {
		resp["result"] = rb.result
	}

	return resp
}

// BuildJSON creates the JSON-RPC response as JSON bytes
func (rb *ResponseBuilder) BuildJSON() []byte {
	rb.t.Helper()

	resp := rb.Build()
	data, err := json.Marshal(resp)
	require.NoError(rb.t, err, "Failed to marshal response")

	return data
}

// BuildString creates the JSON-RPC response as a JSON string
func (rb *ResponseBuilder) BuildString() string {
	return string(rb.BuildJSON())
}

// MCPMessageBuilder builds MCP protocol messages for testing
type MCPMessageBuilder struct {
	t       *testing.T
	msgType string
	content map[string]interface{}
}

// NewMCPMessageBuilder creates a new MCP message builder
func NewMCPMessageBuilder(t *testing.T) *MCPMessageBuilder {
	return &MCPMessageBuilder{
		t:       t,
		content: make(map[string]interface{}),
	}
}

// Initialize creates an MCP initialize request
func (mb *MCPMessageBuilder) Initialize(clientName, clientVersion, protocolVersion string) *MCPMessageBuilder {
	mb.msgType = "initialize"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": protocolVersion,
			"clientInfo": map[string]interface{}{
				"name":    clientName,
				"version": clientVersion,
			},
			"capabilities": map[string]interface{}{},
		},
		"id": 1,
	}
	return mb
}

// InitializeWithCapabilities creates an MCP initialize request with capabilities
func (mb *MCPMessageBuilder) InitializeWithCapabilities(clientName, clientVersion, protocolVersion string, capabilities map[string]interface{}) *MCPMessageBuilder {
	mb.Initialize(clientName, clientVersion, protocolVersion)
	if params, ok := mb.content["params"].(map[string]interface{}); ok {
		params["capabilities"] = capabilities
	}
	return mb
}

// InitializeResponse creates an MCP initialize response
func (mb *MCPMessageBuilder) InitializeResponse(serverName, serverVersion, protocolVersion string) *MCPMessageBuilder {
	mb.msgType = "initialize_response"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"protocolVersion": protocolVersion,
			"serverInfo": map[string]interface{}{
				"name":    serverName,
				"version": serverVersion,
			},
			"capabilities": map[string]interface{}{},
		},
		"id": 1,
	}
	return mb
}

// InitializedNotification creates an MCP initialized notification
func (mb *MCPMessageBuilder) InitializedNotification() *MCPMessageBuilder {
	mb.msgType = "initialized"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "initialized",
		"params":  map[string]interface{}{},
	}
	return mb
}

// ToolsList creates a tools/list request
func (mb *MCPMessageBuilder) ToolsList(id interface{}) *MCPMessageBuilder {
	mb.msgType = "tools_list"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/list",
		"params":  map[string]interface{}{},
		"id":      id,
	}
	return mb
}

// ToolsListResponse creates a tools/list response
func (mb *MCPMessageBuilder) ToolsListResponse(tools []map[string]interface{}, id interface{}) *MCPMessageBuilder {
	mb.msgType = "tools_list_response"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"tools": tools,
		},
		"id": id,
	}
	return mb
}

// ToolCall creates a tools/call request
func (mb *MCPMessageBuilder) ToolCall(name string, arguments interface{}, id interface{}) *MCPMessageBuilder {
	mb.msgType = "tool_call"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
		"id": id,
	}
	return mb
}

// ToolCallResponse creates a tools/call response
func (mb *MCPMessageBuilder) ToolCallResponse(content interface{}, isError bool, id interface{}) *MCPMessageBuilder {
	mb.msgType = "tool_call_response"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"content": content,
			"isError": isError,
		},
		"id": id,
	}
	return mb
}

// PromptsList creates a prompts/list request
func (mb *MCPMessageBuilder) PromptsList(id interface{}) *MCPMessageBuilder {
	mb.msgType = "prompts_list"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "prompts/list",
		"params":  map[string]interface{}{},
		"id":      id,
	}
	return mb
}

// PromptsListResponse creates a prompts/list response
func (mb *MCPMessageBuilder) PromptsListResponse(prompts []map[string]interface{}, id interface{}) *MCPMessageBuilder {
	mb.msgType = "prompts_list_response"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"prompts": prompts,
		},
		"id": id,
	}
	return mb
}

// ResourcesList creates a resources/list request
func (mb *MCPMessageBuilder) ResourcesList(id interface{}) *MCPMessageBuilder {
	mb.msgType = "resources_list"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "resources/list",
		"params":  map[string]interface{}{},
		"id":      id,
	}
	return mb
}

// ResourcesListResponse creates a resources/list response
func (mb *MCPMessageBuilder) ResourcesListResponse(resources []map[string]interface{}, id interface{}) *MCPMessageBuilder {
	mb.msgType = "resources_list_response"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"resources": resources,
		},
		"id": id,
	}
	return mb
}

// Notification creates a generic notification
func (mb *MCPMessageBuilder) Notification(method string, params interface{}) *MCPMessageBuilder {
	mb.msgType = "notification"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	return mb
}

// ErrorResponse creates an error response
func (mb *MCPMessageBuilder) ErrorResponse(code int, message string, data interface{}, id interface{}) *MCPMessageBuilder {
	mb.msgType = "error_response"
	mb.content = map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
			"data":    data,
		},
		"id": id,
	}
	return mb
}

// WithField adds a custom field to the message
func (mb *MCPMessageBuilder) WithField(key string, value interface{}) *MCPMessageBuilder {
	mb.content[key] = value
	return mb
}

// WithParam adds a parameter to the params field
func (mb *MCPMessageBuilder) WithParam(key string, value interface{}) *MCPMessageBuilder {
	if params, ok := mb.content["params"].(map[string]interface{}); ok {
		params[key] = value
	}
	return mb
}

// Build returns the message as a map
func (mb *MCPMessageBuilder) Build() map[string]interface{} {
	return mb.content
}

// BuildJSON returns the message as JSON bytes
func (mb *MCPMessageBuilder) BuildJSON() []byte {
	mb.t.Helper()

	data, err := json.Marshal(mb.content)
	require.NoError(mb.t, err, "Failed to marshal MCP message")

	return data
}

// BuildString returns the message as a JSON string
func (mb *MCPMessageBuilder) BuildString() string {
	return string(mb.BuildJSON())
}

// TestDataBuilder builds complex test data structures
type TestDataBuilder struct {
	t    *testing.T
	data map[string]interface{}
}

// NewTestDataBuilder creates a new test data builder
func NewTestDataBuilder(t *testing.T) *TestDataBuilder {
	return &TestDataBuilder{
		t:    t,
		data: make(map[string]interface{}),
	}
}

// WithField adds a field to the data
func (tb *TestDataBuilder) WithField(key string, value interface{}) *TestDataBuilder {
	tb.data[key] = value
	return tb
}

// WithTimestamp adds a timestamp field
func (tb *TestDataBuilder) WithTimestamp(key string, t time.Time) *TestDataBuilder {
	tb.data[key] = t.Format(time.RFC3339)
	return tb
}

// WithNow adds a current timestamp field
func (tb *TestDataBuilder) WithNow(key string) *TestDataBuilder {
	return tb.WithTimestamp(key, time.Now())
}

// WithArray adds an array field
func (tb *TestDataBuilder) WithArray(key string, values ...interface{}) *TestDataBuilder {
	tb.data[key] = values
	return tb
}

// WithObject adds a nested object field
func (tb *TestDataBuilder) WithObject(key string, obj map[string]interface{}) *TestDataBuilder {
	tb.data[key] = obj
	return tb
}

// WithObjectBuilder adds a nested object using another builder
func (tb *TestDataBuilder) WithObjectBuilder(key string, builder func() map[string]interface{}) *TestDataBuilder {
	tb.data[key] = builder()
	return tb
}

// Build returns the built data
func (tb *TestDataBuilder) Build() map[string]interface{} {
	return tb.data
}

// BuildJSON returns the data as JSON bytes
func (tb *TestDataBuilder) BuildJSON() []byte {
	tb.t.Helper()

	data, err := json.Marshal(tb.data)
	require.NoError(tb.t, err, "Failed to marshal test data")

	return data
}

// BuildString returns the data as a JSON string
func (tb *TestDataBuilder) BuildString() string {
	return string(tb.BuildJSON())
}

// SequenceBuilder builds sequences of messages for testing
type SequenceBuilder struct {
	t        *testing.T
	messages []interface{}
}

// NewSequenceBuilder creates a new sequence builder
func NewSequenceBuilder(t *testing.T) *SequenceBuilder {
	return &SequenceBuilder{
		t:        t,
		messages: make([]interface{}, 0),
	}
}

// AddMessage adds a message to the sequence
func (sb *SequenceBuilder) AddMessage(msg interface{}) *SequenceBuilder {
	sb.messages = append(sb.messages, msg)
	return sb
}

// AddRequest adds a request to the sequence
func (sb *SequenceBuilder) AddRequest(method string, params interface{}, id interface{}) *SequenceBuilder {
	req := NewRequestBuilder(sb.t).
		WithMethod(method).
		WithParams(params).
		WithID(id).
		Build()
	return sb.AddMessage(req)
}

// AddResponse adds a response to the sequence
func (sb *SequenceBuilder) AddResponse(result interface{}, id interface{}) *SequenceBuilder {
	resp := NewResponseBuilder(sb.t).
		WithResult(result).
		WithID(id).
		Build()
	return sb.AddMessage(resp)
}

// AddError adds an error response to the sequence
func (sb *SequenceBuilder) AddError(code int, message string, id interface{}) *SequenceBuilder {
	resp := NewResponseBuilder(sb.t).
		WithError(code, message, nil).
		WithID(id).
		Build()
	return sb.AddMessage(resp)
}

// AddDelay adds a delay marker to the sequence
func (sb *SequenceBuilder) AddDelay(duration time.Duration) *SequenceBuilder {
	return sb.AddMessage(fmt.Sprintf("DELAY:%v", duration))
}

// Build returns the message sequence
func (sb *SequenceBuilder) Build() []interface{} {
	return sb.messages
}

// BuildJSON returns each message as JSON bytes
func (sb *SequenceBuilder) BuildJSON() [][]byte {
	sb.t.Helper()

	result := make([][]byte, 0, len(sb.messages))

	for _, msg := range sb.messages {
		if str, ok := msg.(string); ok && strings.HasPrefix(str, "DELAY:") {
			// Skip delay markers
			continue
		}

		data, err := json.Marshal(msg)
		require.NoError(sb.t, err, "Failed to marshal message in sequence")
		result = append(result, data)
	}

	return result
}
