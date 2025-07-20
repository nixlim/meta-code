package mcp

// Re-export commonly used constants from mcp-go
const (
	// Protocol version constants - using mcp-go defaults
	ProtocolVersionLatest  = "2024-11-05"
	ProtocolVersionMinimum = "2024-11-05"
)

// Method name constants for MCP protocol
const (
	// Core protocol methods
	MethodInitialize  = "initialize"
	MethodInitialized = "initialized"
	MethodShutdown    = "shutdown"
	MethodExit        = "exit"

	// Resource methods
	MethodListResources = "resources/list"
	MethodReadResource  = "resources/read"
	MethodSubscribe     = "resources/subscribe"
	MethodUnsubscribe   = "resources/unsubscribe"

	// Tool methods
	MethodListTools = "tools/list"
	MethodCallTool  = "tools/call"

	// Prompt methods
	MethodListPrompts = "prompts/list"
	MethodGetPrompt   = "prompts/get"

	// Logging methods
	MethodSetLogLevel = "logging/setLevel"

	// Notification methods
	MethodNotificationCancelled        = "notifications/cancelled"
	MethodNotificationProgress         = "notifications/progress"
	MethodNotificationResourcesChanged = "notifications/resources/list_changed"
	MethodNotificationToolsChanged     = "notifications/tools/list_changed"
	MethodNotificationPromptsChanged   = "notifications/prompts/list_changed"
)

// MCP-specific error codes (extending JSON-RPC error codes)
const (
	// ErrorCodeInvalidRequest represents an invalid MCP request
	ErrorCodeInvalidRequest = -32600

	// ErrorCodeMethodNotFound represents a method not found error
	ErrorCodeMethodNotFound = -32601

	// ErrorCodeInvalidParams represents invalid parameters error
	ErrorCodeInvalidParams = -32602

	// ErrorCodeInternalError represents an internal error
	ErrorCodeInternalError = -32603

	// MCP-specific error codes (range: -32000 to -32099)
	ErrorCodeResourceNotFound    = -32001
	ErrorCodeResourceUnavailable = -32002
	ErrorCodeToolNotFound        = -32003
	ErrorCodeToolExecutionError  = -32004
	ErrorCodePromptNotFound      = -32005
	ErrorCodeInvalidCapability   = -32006
	ErrorCodeProtocolMismatch    = -32007
	ErrorCodeUnauthorized        = -32008
	ErrorCodeRateLimited         = -32009
	ErrorCodeTimeout             = -32010
)

// Error messages for MCP-specific error codes
var MCPErrorMessages = map[int]string{
	ErrorCodeResourceNotFound:    "Resource not found",
	ErrorCodeResourceUnavailable: "Resource unavailable",
	ErrorCodeToolNotFound:        "Tool not found",
	ErrorCodeToolExecutionError:  "Tool execution error",
	ErrorCodePromptNotFound:      "Prompt not found",
	ErrorCodeInvalidCapability:   "Invalid capability",
	ErrorCodeProtocolMismatch:    "Protocol version mismatch",
	ErrorCodeUnauthorized:        "Unauthorized access",
	ErrorCodeRateLimited:         "Rate limit exceeded",
	ErrorCodeTimeout:             "Request timeout",
}

// Capability constants
const (
	// Server capabilities
	CapabilityResources = "resources"
	CapabilityTools     = "tools"
	CapabilityPrompts   = "prompts"
	CapabilityLogging   = "logging"

	// Client capabilities
	CapabilityRoots        = "roots"
	CapabilitySampling     = "sampling"
	CapabilityExperimental = "experimental"
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)
