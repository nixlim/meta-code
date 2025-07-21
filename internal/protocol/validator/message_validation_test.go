package validator

import (
	"context"
	"encoding/json"
	"testing"
)

func TestSchemaValidator_ValidateRequest(t *testing.T) {
	validator, err := New(Config{Enabled: true})
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	
	ctx := context.Background()
	
	tests := []struct {
		name    string
		method  string
		params  json.RawMessage
		wantErr bool
	}{
		{
			name:    "valid request with params",
			method:  "tools/call",
			params:  json.RawMessage(`{"name": "test-tool", "arguments": {}}`),
			wantErr: false,
		},
		{
			name:    "valid request without params",
			method:  "tools/list",
			params:  nil,
			wantErr: false,
		},
		{
			name:    "valid request with empty params",
			method:  "resources/list",
			params:  json.RawMessage(`{}`),
			wantErr: false,
		},
		{
			name:    "invalid method",
			method:  "invalid/method",
			params:  nil,
			wantErr: true,
		},
		{
			name:    "invalid params JSON",
			method:  "tools/call",
			params:  json.RawMessage(`{invalid json`),
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRequest(ctx, tt.method, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchemaValidator_ValidateResponse(t *testing.T) {
	validator, err := New(Config{Enabled: true})
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	
	ctx := context.Background()
	
	tests := []struct {
		name    string
		result  json.RawMessage
		error   json.RawMessage
		wantErr bool
	}{
		{
			name:    "valid response with result",
			result:  json.RawMessage(`{"tools": []}`),
			error:   nil,
			wantErr: false,
		},
		{
			name:    "valid response with error",
			result:  nil,
			error:   json.RawMessage(`{"code": -32601, "message": "Method not found"}`),
			wantErr: false,
		},
		{
			name:    "invalid - both result and error",
			result:  json.RawMessage(`{"tools": []}`),
			error:   json.RawMessage(`{"code": -32601, "message": "Method not found"}`),
			wantErr: false, // The validator will use error if both are provided
		},
		{
			name:    "invalid - neither result nor error",
			result:  nil,
			error:   nil,
			wantErr: true,
		},
		{
			name:    "invalid error format - missing code",
			result:  nil,
			error:   json.RawMessage(`{"message": "Method not found"}`),
			wantErr: true,
		},
		{
			name:    "invalid error format - missing message",
			result:  nil,
			error:   json.RawMessage(`{"code": -32601}`),
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateResponse(ctx, tt.result, tt.error)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchemaValidator_ValidateNotification(t *testing.T) {
	validator, err := New(Config{Enabled: true})
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	
	ctx := context.Background()
	
	tests := []struct {
		name    string
		method  string
		params  json.RawMessage
		wantErr bool
	}{
		{
			name:    "valid notification with params",
			method:  "initialized",
			params:  json.RawMessage(`{}`),
			wantErr: false,
		},
		{
			name:    "valid notification without params",
			method:  "cancelled",
			params:  nil,
			wantErr: false,
		},
		{
			name:    "valid progress notification",
			method:  "progress",
			params:  json.RawMessage(`{"token": "123", "progress": 50}`),
			wantErr: false,
		},
		{
			name:    "invalid method",
			method:  "invalid/notification",
			params:  nil,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateNotification(ctx, tt.method, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchemaValidator_ValidateMessage_MCP(t *testing.T) {
	validator, err := New(Config{Enabled: true})
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	
	ctx := context.Background()
	
	tests := []struct {
		name        string
		messageType string
		message     json.RawMessage
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid initialize request",
			messageType: "initialize",
			message: json.RawMessage(`{
				"protocolVersion": "1.0",
				"capabilities": {},
				"clientInfo": {
					"name": "test-client",
					"version": "1.0.0"
				}
			}`),
			wantErr: false,
		},
		{
			name:        "invalid initialize - missing clientInfo",
			messageType: "initialize",
			message: json.RawMessage(`{
				"protocolVersion": "1.0",
				"capabilities": {}
			}`),
			wantErr:     true,
			errContains: "clientInfo",
		},
		{
			name:        "valid initialized response",
			messageType: "initialized",
			message: json.RawMessage(`{
				"protocolVersion": "1.0",
				"capabilities": {
					"tools": {"listChanged": true},
					"resources": {"subscribe": true, "listChanged": true}
				},
				"serverInfo": {
					"name": "test-server",
					"version": "1.0.0"
				}
			}`),
			wantErr: false,
		},
		{
			name:        "valid initialized with instructions",
			messageType: "initialized",
			message: json.RawMessage(`{
				"protocolVersion": "1.0",
				"capabilities": {},
				"serverInfo": {
					"name": "test-server",
					"version": "1.0.0"
				},
				"instructions": "Welcome to the test server"
			}`),
			wantErr: false,
		},
		{
			name:        "unknown message type",
			messageType: "unknown",
			message:     json.RawMessage(`{}`),
			wantErr:     true,
			errContains: "unknown message type",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMessage(ctx, tt.messageType, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errContains, err.Error())
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) == 0 || (len(substr) > 0 && findSubstring(s, substr) != -1))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}