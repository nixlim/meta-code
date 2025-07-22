package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// TestRequestResponseConformance tests request/response pattern conformance
func (suite *ConformanceTestSuite) TestRequestResponseConformance(t *testing.T) {
	ctx := context.Background()

	// Test various request formats
	t.Run("RequestFormats", func(t *testing.T) {
		tests := []struct {
			name        string
			method      string
			params      json.RawMessage
			shouldPass  bool
			description string
		}{
			{
				name:        "tools_list_no_params",
				method:      "tools/list",
				params:      nil,
				shouldPass:  true,
				description: "tools/list request without parameters",
			},
			{
				name:        "tools_call_with_params",
				method:      "tools/call",
				params:      json.RawMessage(`{"name": "echo", "arguments": {"message": "test"}}`),
				shouldPass:  true,
				description: "tools/call request with valid parameters",
			},
			{
				name:        "resources_read_with_uri",
				method:      "resources/read",
				params:      json.RawMessage(`{"uri": "file:///test.txt"}`),
				shouldPass:  true,
				description: "resources/read request with URI parameter",
			},
			{
				name:        "invalid_method_name",
				method:      "invalid/method/name",
				params:      nil,
				shouldPass:  false,
				description: "Request with invalid method name",
			},
			{
				name:        "empty_method_name",
				method:      "",
				params:      nil,
				shouldPass:  false,
				description: "Request with empty method name",
			},
			{
				name:        "malformed_params",
				method:      "tools/call",
				params:      json.RawMessage(`{invalid json}`),
				shouldPass:  false,
				description: "Request with malformed parameters",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := suite.validator.ValidateRequest(ctx, tt.method, tt.params)
				passed := (err == nil) == tt.shouldPass

				result := TestResult{
					TestName:    fmt.Sprintf("request_%s", tt.name),
					Category:    "RequestResponse",
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

	// Test various response formats
	t.Run("ResponseFormats", func(t *testing.T) {
		tests := []struct {
			name        string
			result      json.RawMessage
			error       json.RawMessage
			shouldPass  bool
			description string
		}{
			{
				name:        "tools_list_result",
				result:      json.RawMessage(`{"tools": [{"name": "echo", "description": "Echo tool"}]}`),
				error:       nil,
				shouldPass:  true,
				description: "Valid tools/list response",
			},
			{
				name:        "empty_result",
				result:      json.RawMessage(`{}`),
				error:       nil,
				shouldPass:  true,
				description: "Valid response with empty result object",
			},
			{
				name:        "null_result",
				result:      json.RawMessage(`null`),
				error:       nil,
				shouldPass:  true,
				description: "Valid response with null result",
			},
			{
				name:        "error_response",
				result:      nil,
				error:       json.RawMessage(`{"code": -32601, "message": "Method not found"}`),
				shouldPass:  true,
				description: "Valid error response",
			},
			{
				name:        "error_with_data",
				result:      nil,
				error:       json.RawMessage(`{"code": -32602, "message": "Invalid params", "data": {"param": "name", "reason": "required"}}`),
				shouldPass:  true,
				description: "Valid error response with additional data",
			},
			{
				name:        "neither_result_nor_error",
				result:      nil,
				error:       nil,
				shouldPass:  false,
				description: "Invalid response missing both result and error",
			},
			{
				name:        "error_missing_code",
				result:      nil,
				error:       json.RawMessage(`{"message": "Something went wrong"}`),
				shouldPass:  false,
				description: "Invalid error response missing code",
			},
			{
				name:        "error_missing_message",
				result:      nil,
				error:       json.RawMessage(`{"code": -32603}`),
				shouldPass:  false,
				description: "Invalid error response missing message",
			},
			{
				name:        "error_invalid_code_type",
				result:      nil,
				error:       json.RawMessage(`{"code": "not-a-number", "message": "Error"}`),
				shouldPass:  false,
				description: "Invalid error response with non-integer code",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := suite.validator.ValidateResponse(ctx, tt.result, tt.error)
				passed := (err == nil) == tt.shouldPass

				result := TestResult{
					TestName:    fmt.Sprintf("response_%s", tt.name),
					Category:    "RequestResponse",
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

	// Test ID correlation
	t.Run("IDCorrelation", func(t *testing.T) {
		tests := []struct {
			name        string
			request     json.RawMessage
			response    json.RawMessage
			shouldPass  bool
			description string
		}{
			{
				name: "matching_string_ids",
				request: json.RawMessage(`{
					"jsonrpc": "2.0",
					"method": "tools/list",
					"id": "abc123"
				}`),
				response: json.RawMessage(`{
					"jsonrpc": "2.0",
					"result": {"tools": []},
					"id": "abc123"
				}`),
				shouldPass:  true,
				description: "Request and response with matching string IDs",
			},
			{
				name: "matching_number_ids",
				request: json.RawMessage(`{
					"jsonrpc": "2.0",
					"method": "tools/list",
					"id": 42
				}`),
				response: json.RawMessage(`{
					"jsonrpc": "2.0",
					"result": {"tools": []},
					"id": 42
				}`),
				shouldPass:  true,
				description: "Request and response with matching number IDs",
			},
			{
				name: "null_id",
				request: json.RawMessage(`{
					"jsonrpc": "2.0",
					"method": "tools/list",
					"id": null
				}`),
				response: json.RawMessage(`{
					"jsonrpc": "2.0",
					"result": {"tools": []},
					"id": null
				}`),
				shouldPass:  true,
				description: "Request and response with null IDs",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Validate request
				reqErr := suite.validator.ValidateMessage(ctx, "request", tt.request)
				// Validate response
				respErr := suite.validator.ValidateMessage(ctx, "response", tt.response)

				passed := (reqErr == nil && respErr == nil) == tt.shouldPass

				result := TestResult{
					TestName:    fmt.Sprintf("id_correlation_%s", tt.name),
					Category:    "RequestResponse",
					Description: tt.description,
					Passed:      passed,
				}

				if reqErr != nil || respErr != nil {
					if !tt.shouldPass {
						result.Details = fmt.Sprintf("Expected validation failure - req: %v, resp: %v", reqErr, respErr)
					} else {
						result.Error = fmt.Sprintf("req: %v, resp: %v", reqErr, respErr)
					}
				}

				suite.recordResult(result)

				if !passed {
					t.Errorf("%s: expected shouldPass=%v, got reqErr=%v, respErr=%v",
						tt.name, tt.shouldPass, reqErr, respErr)
				}
			})
		}
	})
}
