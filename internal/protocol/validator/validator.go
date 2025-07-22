// Package validator provides JSON schema validation for MCP protocol messages
package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/schemas"
	"github.com/xeipuuv/gojsonschema"
)

// Validator defines the interface for MCP message validation
type Validator interface {
	// ValidateMessage validates a raw JSON message against the MCP schema
	ValidateMessage(ctx context.Context, messageType string, message json.RawMessage) error

	// ValidateRequest validates an MCP request message
	ValidateRequest(ctx context.Context, method string, params json.RawMessage) error

	// ValidateResponse validates an MCP response message
	ValidateResponse(ctx context.Context, result json.RawMessage, error json.RawMessage) error

	// ValidateNotification validates an MCP notification message
	ValidateNotification(ctx context.Context, method string, params json.RawMessage) error

	// IsEnabled returns whether validation is enabled
	IsEnabled() bool
}

// SchemaValidator implements the Validator interface using JSON schema validation
type SchemaValidator struct {
	enabled bool
	schemas map[string]*gojsonschema.Schema // Compiled schemas keyed by message type
}

// ValidationError represents a schema validation error with details
type ValidationError struct {
	Field        string `json:"field,omitempty"`
	Value        string `json:"value,omitempty"`
	Message      string `json:"message"`
	SchemaPath   string `json:"schemaPath,omitempty"`
	InstancePath string `json:"instancePath,omitempty"`
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error at field '%s': %s", e.Field, e.Message)
	}
	return e.Message
}

// MultiValidationError represents multiple validation errors
type MultiValidationError struct {
	Errors []ValidationError `json:"errors"`
}

func (e *MultiValidationError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple validation errors (%d errors)", len(e.Errors))
}

// MessageType represents the type of MCP message
type MessageType string

const (
	MessageTypeInitialize   MessageType = "initialize"
	MessageTypeInitialized  MessageType = "initialized"
	MessageTypeRequest      MessageType = "request"
	MessageTypeResponse     MessageType = "response"
	MessageTypeNotification MessageType = "notification"
	MessageTypeError        MessageType = "error"
)

// Config holds validator configuration
type Config struct {
	// Enabled determines if validation is active
	Enabled bool

	// SchemaDir is the directory containing schema files
	SchemaDir string

	// CacheSchemas determines if compiled schemas should be cached
	CacheSchemas bool

	// StrictMode enables strict validation (fail on unknown fields)
	StrictMode bool
}

// New creates a new schema validator with the given configuration
func New(config Config) (*SchemaValidator, error) {
	validator := &SchemaValidator{
		enabled: config.Enabled,
		schemas: make(map[string]*gojsonschema.Schema),
	}

	if config.Enabled {
		// Load and compile schemas
		if err := validator.loadSchemas(config); err != nil {
			return nil, fmt.Errorf("failed to load schemas: %w", err)
		}
	}

	return validator, nil
}

// loadSchemas loads and compiles JSON schemas from files or embedded resources
func (v *SchemaValidator) loadSchemas(config Config) error {
	// Define schema mappings
	schemaMap := map[string]string{
		"jsonrpc":      schemas.JSONRPCSchema,
		"initialize":   schemas.MCPInitializeSchema,
		"initialized":  schemas.MCPInitializedSchema,
		"request":      schemas.MCPRequestSchema,
		"response":     schemas.MCPResponseSchema,
		"notification": schemas.MCPNotificationSchema,
		"error":        schemas.MCPErrorSchema,
	}

	// Compile each schema
	for messageType, schemaJSON := range schemaMap {
		loader := gojsonschema.NewStringLoader(schemaJSON)
		schema, err := gojsonschema.NewSchema(loader)
		if err != nil {
			return fmt.Errorf("failed to compile schema for %s: %w", messageType, err)
		}
		v.schemas[messageType] = schema
	}

	return nil
}

// ValidateMessage validates a raw JSON message against the MCP schema
func (v *SchemaValidator) ValidateMessage(ctx context.Context, messageType string, message json.RawMessage) error {
	if !v.enabled {
		return nil
	}

	schema, ok := v.schemas[messageType]
	if !ok {
		return fmt.Errorf("unknown message type: %s", messageType)
	}

	// Validate the message
	documentLoader := gojsonschema.NewBytesLoader(message)
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		return v.formatValidationErrors(result.Errors())
	}

	return nil
}

// ValidateRequest validates an MCP request message
func (v *SchemaValidator) ValidateRequest(ctx context.Context, method string, params json.RawMessage) error {
	if !v.enabled {
		return nil
	}

	// Build a complete request message for validation
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      "validation-test",
	}

	if params != nil && len(params) > 0 && strings.TrimSpace(string(params)) != "" {
		var p interface{}
		if err := json.Unmarshal(params, &p); err != nil {
			return fmt.Errorf("invalid params: %w", err)
		}
		request["params"] = p
	}

	message, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	return v.ValidateMessage(ctx, "request", message)
}

// ValidateResponse validates an MCP response message
func (v *SchemaValidator) ValidateResponse(ctx context.Context, result json.RawMessage, errMsg json.RawMessage) error {
	if !v.enabled {
		return nil
	}

	// Build a complete response message for validation
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "validation-test",
	}

	if errMsg != nil && len(errMsg) > 0 {
		var e interface{}
		if err := json.Unmarshal(errMsg, &e); err != nil {
			return fmt.Errorf("invalid error: %w", err)
		}
		response["error"] = e
	} else if result != nil && len(result) > 0 {
		var r interface{}
		if err := json.Unmarshal(result, &r); err != nil {
			return fmt.Errorf("invalid result: %w", err)
		}
		response["result"] = r
	} else {
		return fmt.Errorf("response must have either result or error")
	}

	message, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	return v.ValidateMessage(ctx, "response", message)
}

// ValidateNotification validates an MCP notification message
func (v *SchemaValidator) ValidateNotification(ctx context.Context, method string, params json.RawMessage) error {
	if !v.enabled {
		return nil
	}

	// Build a complete notification message for validation
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
	}

	if params != nil && len(params) > 0 && strings.TrimSpace(string(params)) != "" {
		var p interface{}
		if err := json.Unmarshal(params, &p); err != nil {
			return fmt.Errorf("invalid params: %w", err)
		}
		notification["params"] = p
	}

	message, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	return v.ValidateMessage(ctx, "notification", message)
}

// IsEnabled returns whether validation is enabled
func (v *SchemaValidator) IsEnabled() bool {
	return v.enabled
}

// formatValidationErrors converts gojsonschema errors to our ValidationError format
func (v *SchemaValidator) formatValidationErrors(errs []gojsonschema.ResultError) error {
	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		err := errs[0]
		return &ValidationError{
			Field:        err.Field(),
			Value:        fmt.Sprintf("%v", err.Value()),
			Message:      err.Description(),
			SchemaPath:   err.Type(),
			InstancePath: err.Context().String(),
		}
	}

	// Multiple errors
	multiErr := &MultiValidationError{
		Errors: make([]ValidationError, len(errs)),
	}

	for i, err := range errs {
		multiErr.Errors[i] = ValidationError{
			Field:        err.Field(),
			Value:        fmt.Sprintf("%v", err.Value()),
			Message:      err.Description(),
			SchemaPath:   err.Type(),
			InstancePath: err.Context().String(),
		}
	}

	return multiErr
}
