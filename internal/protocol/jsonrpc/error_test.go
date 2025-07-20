package jsonrpc

import (
	"encoding/json"
	"testing"
)

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"Parse Error", ErrorCodeParse, -32700},
		{"Invalid Request", ErrorCodeInvalidRequest, -32600},
		{"Method Not Found", ErrorCodeMethodNotFound, -32601},
		{"Invalid Params", ErrorCodeInvalidParams, -32602},
		{"Internal Error", ErrorCodeInternal, -32603},
		{"Server Error", ErrorCodeServerError, -32000},
		{"Not Implemented", ErrorCodeNotImplemented, -32001},
		{"Timeout", ErrorCodeTimeout, -32002},
		{"Resource Limit", ErrorCodeResourceLimit, -32003},
		{"Unauthorized", ErrorCodeUnauthorized, -32004},
		{"Forbidden", ErrorCodeForbidden, -32005},
		{"Not Found", ErrorCodeNotFound, -32006},
		{"Conflict", ErrorCodeConflict, -32007},
		{"Too Many Requests", ErrorCodeTooManyRequests, -32008},
		{"Bad Gateway", ErrorCodeBadGateway, -32009},
		{"Service Unavailable", ErrorCodeServiceUnavail, -32010},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("Error code %s = %d, want %d", tt.name, tt.code, tt.expected)
			}
		})
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
		data    any
	}{
		{
			name:    "simple error",
			code:    -32600,
			message: "Invalid Request",
			data:    nil,
		},
		{
			name:    "error with string data",
			code:    -32602,
			message: "Invalid params",
			data:    "missing required field 'id'",
		},
		{
			name:    "error with struct data",
			code:    -32602,
			message: "Invalid params",
			data:    struct{ Field string }{Field: "value"},
		},
		{
			name:    "error with map data",
			code:    -32603,
			message: "Internal error",
			data:    map[string]string{"reason": "database connection failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.code, tt.message, tt.data)
			
			if err.Code != tt.code {
				t.Errorf("NewError() code = %d, want %d", err.Code, tt.code)
			}
			if err.Message != tt.message {
				t.Errorf("NewError() message = %s, want %s", err.Message, tt.message)
			}
			// Special handling for map comparison
			if tt.name == "error with map data" {
				dataMap, ok := err.Data.(map[string]string)
				if !ok {
					t.Errorf("NewError() data type = %T, want map[string]string", err.Data)
				} else {
					expectedMap := tt.data.(map[string]string)
					if dataMap["reason"] != expectedMap["reason"] {
						t.Errorf("NewError() data map = %v, want %v", dataMap, expectedMap)
					}
				}
			} else if err.Data != tt.data {
				t.Errorf("NewError() data = %v, want %v", err.Data, tt.data)
			}
		})
	}
}

func TestNewStandardError(t *testing.T) {
	tests := []struct {
		name            string
		code            int
		data            any
		expectedMessage string
	}{
		{
			name:            "parse error",
			code:            ErrorCodeParse,
			data:            "unexpected EOF",
			expectedMessage: "Parse error",
		},
		{
			name:            "invalid request",
			code:            ErrorCodeInvalidRequest,
			data:            nil,
			expectedMessage: "Invalid Request",
		},
		{
			name:            "method not found",
			code:            ErrorCodeMethodNotFound,
			data:            "unknown_method",
			expectedMessage: "Method not found",
		},
		{
			name:            "unknown error code",
			code:            -99999,
			data:            nil,
			expectedMessage: "Unknown error",
		},
		{
			name:            "server error",
			code:            ErrorCodeServerError,
			data:            map[string]string{"detail": "internal failure"},
			expectedMessage: "Server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewStandardError(tt.code, tt.data)
			
			if err.Code != tt.code {
				t.Errorf("NewStandardError() code = %d, want %d", err.Code, tt.code)
			}
			if err.Message != tt.expectedMessage {
				t.Errorf("NewStandardError() message = %s, want %s", err.Message, tt.expectedMessage)
			}
			// Special handling for map comparison
			if tt.name == "server error" && tt.data != nil {
				dataMap, ok := err.Data.(map[string]string)
				if !ok {
					t.Errorf("NewStandardError() data type = %T, want map[string]string", err.Data)
				} else {
					expectedMap := tt.data.(map[string]string)
					if dataMap["detail"] != expectedMap["detail"] {
						t.Errorf("NewStandardError() data map = %v, want %v", dataMap, expectedMap)
					}
				}
			} else if err.Data != tt.data {
				t.Errorf("NewStandardError() data = %v, want %v", err.Data, tt.data)
			}
		})
	}
}

func TestErrorHelperFunctions(t *testing.T) {
	t.Run("NewParseError", func(t *testing.T) {
		err := NewParseError("invalid JSON")
		if err.Code != ErrorCodeParse {
			t.Errorf("NewParseError() code = %d, want %d", err.Code, ErrorCodeParse)
		}
		if err.Message != "Parse error" {
			t.Errorf("NewParseError() message = %s, want %s", err.Message, "Parse error")
		}
		if err.Data != "invalid JSON" {
			t.Errorf("NewParseError() data = %v, want %v", err.Data, "invalid JSON")
		}
	})

	t.Run("NewInvalidRequestError", func(t *testing.T) {
		err := NewInvalidRequestError("missing jsonrpc field")
		if err.Code != ErrorCodeInvalidRequest {
			t.Errorf("NewInvalidRequestError() code = %d, want %d", err.Code, ErrorCodeInvalidRequest)
		}
		if err.Message != "Invalid Request" {
			t.Errorf("NewInvalidRequestError() message = %s, want %s", err.Message, "Invalid Request")
		}
	})

	t.Run("NewMethodNotFoundError", func(t *testing.T) {
		err := NewMethodNotFoundError("unknown_method")
		if err.Code != ErrorCodeMethodNotFound {
			t.Errorf("NewMethodNotFoundError() code = %d, want %d", err.Code, ErrorCodeMethodNotFound)
		}
		if err.Message != "Method not found" {
			t.Errorf("NewMethodNotFoundError() message = %s, want %s", err.Message, "Method not found")
		}
		if err.Data != "unknown_method" {
			t.Errorf("NewMethodNotFoundError() data = %v, want %v", err.Data, "unknown_method")
		}
	})

	t.Run("NewInvalidParamsError", func(t *testing.T) {
		err := NewInvalidParamsError(map[string]string{"field": "missing"})
		if err.Code != ErrorCodeInvalidParams {
			t.Errorf("NewInvalidParamsError() code = %d, want %d", err.Code, ErrorCodeInvalidParams)
		}
		if err.Message != "Invalid params" {
			t.Errorf("NewInvalidParamsError() message = %s, want %s", err.Message, "Invalid params")
		}
	})

	t.Run("NewInternalError", func(t *testing.T) {
		err := NewInternalError("database error")
		if err.Code != ErrorCodeInternal {
			t.Errorf("NewInternalError() code = %d, want %d", err.Code, ErrorCodeInternal)
		}
		if err.Message != "Internal error" {
			t.Errorf("NewInternalError() message = %s, want %s", err.Message, "Internal error")
		}
	})
}

func TestErrorClassificationFunctions(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(int) bool
		testCases []struct {
			code     int
			expected bool
		}
	}{
		{
			name:     "IsStandardError",
			testFunc: IsStandardError,
			testCases: []struct {
				code     int
				expected bool
			}{
				{-32768, true},
				{-32700, true},
				{-32600, true},
				{-32000, true},
				{-31999, false},
				{-32769, false},
				{0, false},
				{100, false},
			},
		},
		{
			name:     "IsServerError",
			testFunc: IsServerError,
			testCases: []struct {
				code     int
				expected bool
			}{
				{-32099, true},
				{-32050, true},
				{-32000, true},
				{-32100, false},
				{-31999, false},
				{0, false},
			},
		},
		{
			name:     "IsApplicationError",
			testFunc: IsApplicationError,
			testCases: []struct {
				code     int
				expected bool
			}{
				{-32769, true},
				{-31999, true},
				{0, true},
				{100, true},
				{-100000, true},
				{-32768, false},
				{-32600, false},
				{-32000, false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, tc := range tt.testCases {
				result := tt.testFunc(tc.code)
				if result != tc.expected {
					t.Errorf("%s(%d) = %v, want %v", tt.name, tc.code, result, tc.expected)
				}
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "error without data",
			err:      &Error{Code: -32600, Message: "Invalid Request"},
			expected: "JSON-RPC error -32600: Invalid Request",
		},
		{
			name:     "error with string data",
			err:      &Error{Code: -32602, Message: "Invalid params", Data: "missing field"},
			expected: "JSON-RPC error -32602: Invalid params (data: missing field)",
		},
		{
			name:     "error with numeric data",
			err:      &Error{Code: -32603, Message: "Internal error", Data: 42},
			expected: "JSON-RPC error -32603: Internal error (data: 42)",
		},
		{
			name:     "error with complex data",
			err:      &Error{Code: -32000, Message: "Server error", Data: map[string]int{"code": 500}},
			expected: "JSON-RPC error -32000: Server error (data: map[code:500])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestError_ToResponse(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		id   any
	}{
		{
			name: "error response with string id",
			err:  NewInvalidRequestError("test"),
			id:   "req-123",
		},
		{
			name: "error response with numeric id",
			err:  NewMethodNotFoundError("unknown"),
			id:   42,
		},
		{
			name: "error response with nil id",
			err:  NewInternalError("server error"),
			id:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.err.ToResponse(tt.id)
			
			if resp.Version != Version {
				t.Errorf("ToResponse() version = %s, want %s", resp.Version, Version)
			}
			if resp.Error != tt.err {
				t.Errorf("ToResponse() error = %v, want %v", resp.Error, tt.err)
			}
			if resp.ID != tt.id {
				t.Errorf("ToResponse() id = %v, want %v", resp.ID, tt.id)
			}
			if resp.Result != nil {
				t.Errorf("ToResponse() result = %v, want nil", resp.Result)
			}
		})
	}
}

func TestValidateCode(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		wantErr bool
	}{
		// Standard error codes
		{"parse error", -32700, false},
		{"invalid request", -32600, false},
		{"method not found", -32601, false},
		{"invalid params", -32602, false},
		{"internal error", -32603, false},
		
		// Server error range
		{"server error start", -32000, false},
		{"server error middle", -32050, false},
		{"server error end", -32099, false},
		
		// Edge cases
		{"standard range start", -32768, false},
		{"standard range end", -32000, false},
		
		// Application-defined errors
		{"positive code", 100, false},
		{"zero code", 0, false},
		{"large negative", -100000, false},
		{"just outside standard range", -32769, false},
		{"just outside server range", -31999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCode(%d) error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
		})
	}
}

func TestErrorJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "error without data",
			err:  &Error{Code: -32600, Message: "Invalid Request"},
			expected: `{"code":-32600,"message":"Invalid Request"}`,
		},
		{
			name: "error with string data",
			err:  &Error{Code: -32602, Message: "Invalid params", Data: "missing field"},
			expected: `{"code":-32602,"message":"Invalid params","data":"missing field"}`,
		},
		{
			name: "error with object data",
			err:  &Error{Code: -32603, Message: "Internal error", Data: map[string]string{"detail": "db error"}},
			expected: `{"code":-32603,"message":"Internal error","data":{"detail":"db error"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.err)
			if err != nil {
				t.Fatalf("Failed to marshal error: %v", err)
			}
			
			if string(data) != tt.expected {
				t.Errorf("JSON marshal = %s, want %s", string(data), tt.expected)
			}
			
			// Test unmarshaling
			var unmarshaled Error
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal error: %v", err)
			}
			
			if unmarshaled.Code != tt.err.Code {
				t.Errorf("Unmarshaled code = %d, want %d", unmarshaled.Code, tt.err.Code)
			}
			if unmarshaled.Message != tt.err.Message {
				t.Errorf("Unmarshaled message = %s, want %s", unmarshaled.Message, tt.err.Message)
			}
		})
	}
}

func TestErrorResponseIntegration(t *testing.T) {
	// Test creating error responses using NewErrorResponse
	err := NewMethodNotFoundError("test_method")
	resp := NewErrorResponse(err, "req-123")
	
	// Marshal to JSON
	data, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal error response: %v", marshalErr)
	}
	
	expected := `{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found","data":"test_method"},"id":"req-123"}`
	if string(data) != expected {
		t.Errorf("Marshaled response = %s, want %s", string(data), expected)
	}
}

func TestErrorMessages(t *testing.T) {
	// Verify all standard error codes have messages
	standardCodes := []int{
		ErrorCodeParse,
		ErrorCodeInvalidRequest,
		ErrorCodeMethodNotFound,
		ErrorCodeInvalidParams,
		ErrorCodeInternal,
		ErrorCodeServerError,
		ErrorCodeNotImplemented,
		ErrorCodeTimeout,
		ErrorCodeResourceLimit,
		ErrorCodeUnauthorized,
		ErrorCodeForbidden,
		ErrorCodeNotFound,
		ErrorCodeConflict,
		ErrorCodeTooManyRequests,
		ErrorCodeBadGateway,
		ErrorCodeServiceUnavail,
	}
	
	for _, code := range standardCodes {
		if _, exists := errorMessages[code]; !exists {
			t.Errorf("Error code %d missing from errorMessages map", code)
		}
	}
	
	// Verify message content
	if msg := errorMessages[ErrorCodeParse]; msg != "Parse error" {
		t.Errorf("Parse error message = %q, want %q", msg, "Parse error")
	}
	if msg := errorMessages[ErrorCodeInvalidRequest]; msg != "Invalid Request" {
		t.Errorf("Invalid request message = %q, want %q", msg, "Invalid Request")
	}
}