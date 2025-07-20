package mcp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// ProtocolVersion represents a semantic version for the MCP protocol
type ProtocolVersion string

// Compare returns -1, 0, or 1 if v is less than, equal to, or greater than other
func (v ProtocolVersion) Compare(other ProtocolVersion) int {
	// Simple string comparison for date-based versions like "2024-11-05"
	if string(v) < string(other) {
		return -1
	}
	if string(v) > string(other) {
		return 1
	}
	return 0
}

// IsValid checks if the protocol version is valid
func (v ProtocolVersion) IsValid() bool {
	// Check if it's a valid date format YYYY-MM-DD
	parts := strings.Split(string(v), "-")
	if len(parts) != 3 {
		return false
	}
	
	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 2024 || year > 2030 {
		return false
	}
	
	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return false
	}
	
	day, err := strconv.Atoi(parts[2])
	if err != nil || day < 1 || day > 31 {
		return false
	}
	
	return true
}

// String returns the string representation of the protocol version
func (v ProtocolVersion) String() string {
	return string(v)
}

// ClientInfo represents information about the MCP client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo represents information about the MCP server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Implementation represents an implementation detail
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Capabilities represents the capabilities supported by client or server
type Capabilities struct {
	// Server capabilities
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
	
	// Client capabilities
	Roots        *RootsCapability        `json:"roots,omitempty"`
	Sampling     *SamplingCapability     `json:"sampling,omitempty"`
	Experimental *ExperimentalCapability `json:"experimental,omitempty"`
}

// ResourcesCapability represents server's resource capabilities
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCapability represents server's tool capabilities
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability represents server's prompt capabilities
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability represents server's logging capabilities
type LoggingCapability struct{}

// RootsCapability represents client's roots capability
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability represents client's sampling capability
type SamplingCapability struct{}

// ExperimentalCapability represents experimental capabilities
type ExperimentalCapability map[string]interface{}

// InitializeParams contains the parameters for initialization
type InitializeParams struct {
	ProtocolVersion ProtocolVersion `json:"protocolVersion"`
	Capabilities    Capabilities    `json:"capabilities"`
	ClientInfo      ClientInfo      `json:"clientInfo"`
}

// InitializeRequest represents the initialize message from client to server
type InitializeRequest struct {
	jsonrpc.Request
	Params InitializeParams `json:"params"`
}

// InitializeResult contains the result of initialization
type InitializeResult struct {
	ProtocolVersion ProtocolVersion `json:"protocolVersion"`
	Capabilities    Capabilities    `json:"capabilities"`
	ServerInfo      ServerInfo      `json:"serverInfo"`
	Instructions    string          `json:"instructions,omitempty"`
}

// InitializeResponse represents the initialize response from server to client
type InitializeResponse struct {
	jsonrpc.Response
	Result *InitializeResult `json:"result,omitempty"`
}

// InitializedNotification represents the initialized notification from client to server
type InitializedNotification struct {
	jsonrpc.Notification
}

// NewInitializeRequest creates a new initialize request
func NewInitializeRequest(params InitializeParams, id interface{}) *InitializeRequest {
	return &InitializeRequest{
		Request: *jsonrpc.NewRequest(MethodInitialize, params, id),
		Params:  params,
	}
}

// NewInitializeResponse creates a new initialize response
func NewInitializeResponse(result InitializeResult, id interface{}) *InitializeResponse {
	return &InitializeResponse{
		Response: *jsonrpc.NewResponse(result, id),
		Result:   &result,
	}
}

// NewInitializedNotification creates a new initialized notification
func NewInitializedNotification() *InitializedNotification {
	return &InitializedNotification{
		Notification: *jsonrpc.NewNotification(MethodInitialized, nil),
	}
}

// Validate validates the initialize request
func (r *InitializeRequest) Validate() error {
	if err := r.Request.Validate(); err != nil {
		return err
	}
	
	if !r.Params.ProtocolVersion.IsValid() {
		return fmt.Errorf("invalid protocol version: %s", r.Params.ProtocolVersion)
	}
	
	if r.Params.ClientInfo.Name == "" {
		return fmt.Errorf("client name is required")
	}
	
	if r.Params.ClientInfo.Version == "" {
		return fmt.Errorf("client version is required")
	}
	
	return nil
}

// Validate validates the initialize response
func (r *InitializeResponse) Validate() error {
	if err := r.Response.Validate(); err != nil {
		return err
	}
	
	if r.Result == nil {
		return fmt.Errorf("initialize result is required")
	}
	
	if !r.Result.ProtocolVersion.IsValid() {
		return fmt.Errorf("invalid protocol version: %s", r.Result.ProtocolVersion)
	}
	
	if r.Result.ServerInfo.Name == "" {
		return fmt.Errorf("server name is required")
	}
	
	if r.Result.ServerInfo.Version == "" {
		return fmt.Errorf("server version is required")
	}
	
	return nil
}

// Validate validates the initialized notification
func (n *InitializedNotification) Validate() error {
	return n.Notification.Validate()
}

// MCPError represents an MCP-specific error
type MCPError struct {
	*jsonrpc.Error
}

// NewMCPError creates a new MCP error with the given code and data
func NewMCPError(code int, data interface{}) *MCPError {
	message, exists := MCPErrorMessages[code]
	if !exists {
		message = "Unknown MCP error"
	}

	return &MCPError{
		Error: &jsonrpc.Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// NewResourceNotFoundError creates a resource not found error
func NewResourceNotFoundError(resource string) *MCPError {
	return NewMCPError(ErrorCodeResourceNotFound, resource)
}

// NewToolNotFoundError creates a tool not found error
func NewToolNotFoundError(tool string) *MCPError {
	return NewMCPError(ErrorCodeToolNotFound, tool)
}

// NewPromptNotFoundError creates a prompt not found error
func NewPromptNotFoundError(prompt string) *MCPError {
	return NewMCPError(ErrorCodePromptNotFound, prompt)
}

// NewProtocolMismatchError creates a protocol mismatch error
func NewProtocolMismatchError(clientVersion, serverVersion ProtocolVersion) *MCPError {
	return NewMCPError(ErrorCodeProtocolMismatch, map[string]string{
		"clientVersion": string(clientVersion),
		"serverVersion": string(serverVersion),
	})
}

// IsCompatible checks if two protocol versions are compatible
func IsCompatible(clientVersion, serverVersion ProtocolVersion) bool {
	// For now, require exact match for date-based versions
	return clientVersion.Compare(serverVersion) == 0
}

// GetSupportedMethods returns a list of all supported MCP methods
func GetSupportedMethods() []string {
	return []string{
		MethodInitialize,
		MethodInitialized,
		MethodShutdown,
		MethodExit,
		MethodListResources,
		MethodReadResource,
		MethodSubscribe,
		MethodUnsubscribe,
		MethodListTools,
		MethodCallTool,
		MethodListPrompts,
		MethodGetPrompt,
		MethodSetLogLevel,
		MethodNotificationCancelled,
		MethodNotificationProgress,
		MethodNotificationResourcesChanged,
		MethodNotificationToolsChanged,
		MethodNotificationPromptsChanged,
	}
}

// IsValidMethod checks if a method name is a valid MCP method
func IsValidMethod(method string) bool {
	supportedMethods := GetSupportedMethods()
	for _, supported := range supportedMethods {
		if method == supported {
			return true
		}
	}
	return false
}
