// Package schemas contains embedded MCP JSON schema definitions
package schemas

import _ "embed"

// MCP Protocol Schema Definitions
// Based on MCP specification: https://spec.modelcontextprotocol.io/

//go:embed jsonrpc.json
var JSONRPCSchema string

//go:embed mcp-initialize.json
var MCPInitializeSchema string

//go:embed mcp-initialized.json
var MCPInitializedSchema string

//go:embed mcp-request.json
var MCPRequestSchema string

//go:embed mcp-response.json
var MCPResponseSchema string

//go:embed mcp-notification.json
var MCPNotificationSchema string

//go:embed mcp-error.json
var MCPErrorSchema string

// GetSchema returns the schema for a given message type
func GetSchema(messageType string) (string, bool) {
	schemas := map[string]string{
		"jsonrpc":      JSONRPCSchema,
		"initialize":   MCPInitializeSchema,
		"initialized":  MCPInitializedSchema,
		"request":      MCPRequestSchema,
		"response":     MCPResponseSchema,
		"notification": MCPNotificationSchema,
		"error":        MCPErrorSchema,
	}

	schema, ok := schemas[messageType]
	return schema, ok
}
