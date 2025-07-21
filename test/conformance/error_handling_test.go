package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// TestErrorHandlingConformance tests error response conformance
func (suite *ConformanceTestSuite) TestErrorHandlingConformance(t *testing.T) {
	ctx := context.Background()
	
	// Test standard error codes
	t.Run("StandardErrorCodes", func(t *testing.T) {
		tests := []struct {
			name        string
			errorObj    json.RawMessage
			shouldPass  bool
			description string
		}{
			// JSON-RPC standard errors
			{
				name:        "parse_error",
				errorObj:    json.RawMessage(`{"code": -32700, "message": "Parse error"}`),
				shouldPass:  true,
				description: "Valid parse error (-32700)",
			},
			{
				name:        "invalid_request",
				errorObj:    json.RawMessage(`{"code": -32600, "message": "Invalid Request"}`),
				shouldPass:  true,
				description: "Valid invalid request error (-32600)",
			},
			{
				name:        "method_not_found",
				errorObj:    json.RawMessage(`{"code": -32601, "message": "Method not found"}`),
				shouldPass:  true,
				description: "Valid method not found error (-32601)",
			},
			{
				name:        "invalid_params",
				errorObj:    json.RawMessage(`{"code": -32602, "message": "Invalid params"}`),
				shouldPass:  true,
				description: "Valid invalid params error (-32602)",
			},
			{
				name:        "internal_error",
				errorObj:    json.RawMessage(`{"code": -32603, "message": "Internal error"}`),
				shouldPass:  true,
				description: "Valid internal error (-32603)",
			},
			// MCP-specific errors
			{
				name:        "resource_not_found",
				errorObj:    json.RawMessage(`{"code": -32001, "message": "Resource not found"}`),
				shouldPass:  true,
				description: "Valid resource not found error (-32001)",
			},
			{
				name:        "resource_error",
				errorObj:    json.RawMessage(`{"code": -32002, "message": "Resource error"}`),
				shouldPass:  true,
				description: "Valid resource error (-32002)",
			},
			{
				name:        "tool_not_found",
				errorObj:    json.RawMessage(`{"code": -32003, "message": "Tool not found"}`),
				shouldPass:  true,
				description: "Valid tool not found error (-32003)",
			},
			{
				name:        "tool_error",
				errorObj:    json.RawMessage(`{"code": -32004, "message": "Tool error"}`),
				shouldPass:  true,
				description: "Valid tool error (-32004)",
			},
			{
				name:        "prompt_not_found",
				errorObj:    json.RawMessage(`{"code": -32005, "message": "Prompt not found"}`),
				shouldPass:  true,
				description: "Valid prompt not found error (-32005)",
			},
			{
				name:        "prompt_error",
				errorObj:    json.RawMessage(`{"code": -32006, "message": "Prompt error"}`),
				shouldPass:  true,
				description: "Valid prompt error (-32006)",
			},
			// Invalid error formats
			{
				name:        "missing_code",
				errorObj:    json.RawMessage(`{"message": "Something went wrong"}`),
				shouldPass:  false,
				description: "Invalid error missing code field",
			},
			{
				name:        "missing_message",
				errorObj:    json.RawMessage(`{"code": -32600}`),
				shouldPass:  false,
				description: "Invalid error missing message field",
			},
			{
				name:        "code_as_string",
				errorObj:    json.RawMessage(`{"code": "-32600", "message": "Invalid Request"}`),
				shouldPass:  false,
				description: "Invalid error with code as string instead of number",
			},
			{
				name:        "message_as_number",
				errorObj:    json.RawMessage(`{"code": -32600, "message": 123}`),
				shouldPass:  false,
				description: "Invalid error with message as number instead of string",
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := suite.validator.ValidateResponse(ctx, nil, tt.errorObj)
				passed := (err == nil) == tt.shouldPass
				
				result := TestResult{
					TestName:    fmt.Sprintf("error_code_%s", tt.name),
					Category:    "ErrorHandling",
					Description: tt.description,
					Passed:      passed,
				}
				
				if err != nil && !tt.shouldPass {
					result.Details = fmt.Sprintf("Expected validation failure: %v", err)
				} else if err != nil && tt.shouldPass {
					result.Error = err.Error()
				}
				
				suite.recordResult(result)
				
				if !passed {
					t.Errorf("%s: expected shouldPass=%v, got error=%v", tt.name, tt.shouldPass, err)
				}
			})
		}
	})
	
	// Test error data field
	t.Run("ErrorDataField", func(t *testing.T) {
		tests := []struct {
			name        string
			errorObj    json.RawMessage
			shouldPass  bool
			description string
		}{
			{
				name: "error_with_string_data",
				errorObj: json.RawMessage(`{
					"code": -32602,
					"message": "Invalid params",
					"data": "Parameter 'name' is required"
				}`),
				shouldPass:  true,
				description: "Valid error with string data field",
			},
			{
				name: "error_with_object_data",
				errorObj: json.RawMessage(`{
					"code": -32602,
					"message": "Invalid params",
					"data": {
						"param": "name",
						"reason": "required",
						"received": null
					}
				}`),
				shouldPass:  true,
				description: "Valid error with object data field",
			},
			{
				name: "error_with_array_data",
				errorObj: json.RawMessage(`{
					"code": -32602,
					"message": "Multiple validation errors",
					"data": [
						{"field": "name", "error": "required"},
						{"field": "age", "error": "must be positive"}
					]
				}`),
				shouldPass:  true,
				description: "Valid error with array data field",
			},
			{
				name: "error_without_data",
				errorObj: json.RawMessage(`{
					"code": -32601,
					"message": "Method not found"
				}`),
				shouldPass:  true,
				description: "Valid error without optional data field",
			},
			{
				name: "error_with_null_data",
				errorObj: json.RawMessage(`{
					"code": -32603,
					"message": "Internal error",
					"data": null
				}`),
				shouldPass:  true,
				description: "Valid error with null data field",
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := suite.validator.ValidateResponse(ctx, nil, tt.errorObj)
				passed := (err == nil) == tt.shouldPass
				
				result := TestResult{
					TestName:    fmt.Sprintf("error_data_%s", tt.name),
					Category:    "ErrorHandling",
					Description: tt.description,
					Passed:      passed,
				}
				
				if err != nil && !tt.shouldPass {
					result.Details = fmt.Sprintf("Expected validation failure: %v", err)
				} else if err != nil && tt.shouldPass {
					result.Error = err.Error()
				}
				
				suite.recordResult(result)
				
				if !passed {
					t.Errorf("%s: expected shouldPass=%v, got error=%v", tt.name, tt.shouldPass, err)
				}
			})
		}
	})
	
	// Test error response structure
	t.Run("ErrorResponseStructure", func(t *testing.T) {
		tests := []struct {
			name        string
			response    json.RawMessage
			shouldPass  bool
			description string
		}{
			{
				name: "valid_error_response",
				response: json.RawMessage(`{
					"jsonrpc": "2.0",
					"error": {
						"code": -32601,
						"message": "Method not found"
					},
					"id": "123"
				}`),
				shouldPass:  true,
				description: "Valid error response structure",
			},
			{
				name: "error_response_without_id",
				response: json.RawMessage(`{
					"jsonrpc": "2.0",
					"error": {
						"code": -32601,
						"message": "Method not found"
					}
				}`),
				shouldPass:  false,
				description: "Invalid error response missing ID",
			},
			{
				name: "error_response_with_result",
				response: json.RawMessage(`{
					"jsonrpc": "2.0",
					"error": {
						"code": -32601,
						"message": "Method not found"
					},
					"result": null,
					"id": "123"
				}`),
				shouldPass:  false,
				description: "Invalid error response with both error and result",
			},
			{
				name: "error_response_null_id",
				response: json.RawMessage(`{
					"jsonrpc": "2.0",
					"error": {
						"code": -32601,
						"message": "Method not found"
					},
					"id": null
				}`),
				shouldPass:  true,
				description: "Valid error response with null ID",
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := suite.validator.ValidateMessage(ctx, "response", tt.response)
				passed := (err == nil) == tt.shouldPass
				
				result := TestResult{
					TestName:    fmt.Sprintf("error_structure_%s", tt.name),
					Category:    "ErrorHandling",
					Description: tt.description,
					Passed:      passed,
				}
				
				if err != nil && !tt.shouldPass {
					result.Details = fmt.Sprintf("Expected validation failure: %v", err)
				} else if err != nil && tt.shouldPass {
					result.Error = err.Error()
				}
				
				suite.recordResult(result)
				
				if !passed {
					t.Errorf("%s: expected shouldPass=%v, got error=%v", tt.name, tt.shouldPass, err)
				}
			})
		}
	})
}