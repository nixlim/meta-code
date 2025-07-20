package jsonrpc

import (
	"encoding/json"
	"testing"
)

func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request *Request
		wantErr bool
	}{
		{
			name: "valid request",
			request: &Request{
				Version: "2.0",
				Method:  "test_method",
				Params:  map[string]any{"key": "value"},
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "valid request with string ID",
			request: &Request{
				Version: "2.0",
				Method:  "test_method",
				Params:  []any{"param1", "param2"},
				ID:      "test-id",
			},
			wantErr: false,
		},
		{
			name: "valid request with nil ID (notification)",
			request: &Request{
				Version: "2.0",
				Method:  "test_method",
				Params:  nil,
				ID:      nil,
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			request: &Request{
				Version: "1.0",
				Method:  "test_method",
				ID:      1,
			},
			wantErr: true,
		},
		{
			name: "empty method",
			request: &Request{
				Version: "2.0",
				Method:  "",
				ID:      1,
			},
			wantErr: true,
		},
		{
			name: "reserved method name",
			request: &Request{
				Version: "2.0",
				Method:  "rpc.test",
				ID:      1,
			},
			wantErr: true,
		},
		{
			name: "invalid ID type",
			request: &Request{
				Version: "2.0",
				Method:  "test_method",
				ID:      []string{"invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Request.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResponseValidation(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
		wantErr  bool
	}{
		{
			name: "valid response with result",
			response: &Response{
				Version: "2.0",
				Result:  "success",
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "valid response with error",
			response: &Response{
				Version: "2.0",
				Error:   NewInternalError("test error"),
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "invalid - both result and error",
			response: &Response{
				Version: "2.0",
				Result:  "success",
				Error:   NewInternalError("test error"),
				ID:      1,
			},
			wantErr: true,
		},
		{
			name: "invalid - neither result nor error",
			response: &Response{
				Version: "2.0",
				ID:      1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.response.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Response.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		msgType string
	}{
		{
			name:    "valid request",
			input:   `{"jsonrpc":"2.0","method":"test","params":{"key":"value"},"id":1}`,
			wantErr: false,
			msgType: "*jsonrpc.Request",
		},
		{
			name:    "valid notification",
			input:   `{"jsonrpc":"2.0","method":"test","params":{"key":"value"}}`,
			wantErr: false,
			msgType: "*jsonrpc.Notification",
		},
		{
			name:    "valid response",
			input:   `{"jsonrpc":"2.0","result":"success","id":1}`,
			wantErr: false,
			msgType: "*jsonrpc.Response",
		},
		{
			name:    "invalid JSON",
			input:   `{"jsonrpc":"2.0","method":"test"`,
			wantErr: true,
		},
		{
			name:    "missing jsonrpc field",
			input:   `{"method":"test","id":1}`,
			wantErr: true,
		},
		{
			name:    "wrong jsonrpc version",
			input:   `{"jsonrpc":"1.0","method":"test","id":1}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseMessage([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && msg != nil {
				msgTypeName := getTypeName(msg)
				if msgTypeName != tt.msgType {
					t.Errorf("ParseMessage() got type %v, want %v", msgTypeName, tt.msgType)
				}
			}
		})
	}
}

func TestParseBatch(t *testing.T) {
	batchInput := `[
		{"jsonrpc":"2.0","method":"test1","id":1},
		{"jsonrpc":"2.0","method":"test2"},
		{"jsonrpc":"2.0","result":"success","id":2}
	]`

	messages, err := Parse([]byte(batchInput))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(messages) != 3 {
		t.Errorf("Parse() got %d messages, want 3", len(messages))
	}

	// Check message types
	if _, ok := messages[0].(*Request); !ok {
		t.Errorf("First message should be Request, got %T", messages[0])
	}
	if _, ok := messages[1].(*Notification); !ok {
		t.Errorf("Second message should be Notification, got %T", messages[1])
	}
	if _, ok := messages[2].(*Response); !ok {
		t.Errorf("Third message should be Response, got %T", messages[2])
	}
}

func TestBindParams(t *testing.T) {
	type TestParams struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	req := &Request{
		Version: "2.0",
		Method:  "test",
		Params:  map[string]any{"name": "test", "value": 42},
		ID:      1,
	}

	var params TestParams
	err := req.BindParams(&params)
	if err != nil {
		t.Fatalf("BindParams() error = %v", err)
	}

	if params.Name != "test" || params.Value != 42 {
		t.Errorf("BindParams() got %+v, want {Name:test Value:42}", params)
	}
}

func getTypeName(v any) string {
	switch v.(type) {
	case *Request:
		return "*jsonrpc.Request"
	case *Response:
		return "*jsonrpc.Response"
	case *Notification:
		return "*jsonrpc.Notification"
	default:
		return "unknown"
	}
}

// Test missing functions for better coverage
func TestErrorMethods(t *testing.T) {
	// Test Error.Error() method
	err := &Error{Code: -32600, Message: "Invalid Request", Data: "test data"}
	expected := "JSON-RPC error -32600: Invalid Request (data: test data)"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	// Test Error.Error() without data
	err2 := &Error{Code: -32601, Message: "Method not found"}
	expected2 := "JSON-RPC error -32601: Method not found"
	if err2.Error() != expected2 {
		t.Errorf("Error() = %q, want %q", err2.Error(), expected2)
	}
}

func TestRequestMethods(t *testing.T) {
	// Test IsNotification
	req := &Request{Version: "2.0", Method: "test", ID: nil}
	if !req.IsNotification() {
		t.Error("IsNotification() should return true for request with nil ID")
	}

	req.ID = "test-id"
	if req.IsNotification() {
		t.Error("IsNotification() should return false for request with ID")
	}
}

func TestErrorHelpers(t *testing.T) {
	// Test NewInvalidParamsError
	err := NewInvalidParamsError("invalid params")
	if err.Code != ErrorCodeInvalidParams {
		t.Errorf("NewInvalidParamsError() code = %d, want %d", err.Code, ErrorCodeInvalidParams)
	}

	// Test IsStandardError
	if !IsStandardError(-32600) {
		t.Error("IsStandardError(-32600) should return true")
	}
	if IsStandardError(-1000) {
		t.Error("IsStandardError(-1000) should return false")
	}

	// Test IsServerError
	if !IsServerError(-32000) {
		t.Error("IsServerError(-32000) should return true")
	}
	if IsServerError(-32600) {
		t.Error("IsServerError(-32600) should return false")
	}

	// Test IsApplicationError
	if !IsApplicationError(-1000) {
		t.Error("IsApplicationError(-1000) should return true")
	}
	if IsApplicationError(-32000) {
		t.Error("IsApplicationError(-32000) should return false")
	}
}

func TestValidationHelpers(t *testing.T) {
	// Test ValidateID
	validIDs := []any{"string", 123, 123.45, nil}
	for _, id := range validIDs {
		if !ValidateID(id) {
			t.Errorf("ValidateID(%v) should return true", id)
		}
	}

	invalidIDs := []any{[]string{"array"}, map[string]any{"object": true}}
	for _, id := range invalidIDs {
		if ValidateID(id) {
			t.Errorf("ValidateID(%v) should return false", id)
		}
	}

	// Test ValidateMethod
	if !ValidateMethod("valid_method") {
		t.Error("ValidateMethod('valid_method') should return true")
	}
	if ValidateMethod("") {
		t.Error("ValidateMethod('') should return false")
	}
	if ValidateMethod("rpc.reserved") {
		t.Error("ValidateMethod('rpc.reserved') should return false")
	}
}

func TestMarshalBatch(t *testing.T) {
	// Test empty batch
	_, err := MarshalBatch([]Message{})
	if err == nil {
		t.Error("MarshalBatch([]) should return error")
	}

	// Test single message batch
	req := NewRequest("test", nil, 1)
	data, err := MarshalBatch([]Message{req})
	if err != nil {
		t.Errorf("MarshalBatch([single]) failed: %v", err)
	}

	// Should not be wrapped in array for single message
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Failed to parse single message result: %v", err)
	}

	// Test multiple messages batch
	req2 := NewRequest("test2", nil, 2)
	data, err = MarshalBatch([]Message{req, req2})
	if err != nil {
		t.Errorf("MarshalBatch([multiple]) failed: %v", err)
	}

	// Should be wrapped in array for multiple messages
	var parsedArray []map[string]any
	if err := json.Unmarshal(data, &parsedArray); err != nil {
		t.Errorf("Failed to parse batch result: %v", err)
	}
	if len(parsedArray) != 2 {
		t.Errorf("Expected 2 messages in batch, got %d", len(parsedArray))
	}
}

func TestParseEdgeCases(t *testing.T) {
	// Test empty input
	_, err := Parse([]byte(""))
	if err == nil {
		t.Error("Parse('') should return error")
	}

	// Test whitespace only
	_, err = Parse([]byte("   \n\t  "))
	if err == nil {
		t.Error("Parse(whitespace) should return error")
	}

	// Test invalid JSON structure
	_, err = Parse([]byte("invalid"))
	if err == nil {
		t.Error("Parse('invalid') should return error")
	}

	// Test valid single object
	validJSON := `{"jsonrpc":"2.0","method":"test","id":1}`
	messages, err := Parse([]byte(validJSON))
	if err != nil {
		t.Errorf("Parse(valid) failed: %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestNotificationValidation(t *testing.T) {
	// Test valid notification
	notif := &Notification{Version: "2.0", Method: "test"}
	if err := notif.Validate(); err != nil {
		t.Errorf("Valid notification failed validation: %v", err)
	}

	// Test notification with reserved method
	notif.Method = "rpc.test"
	if err := notif.Validate(); err == nil {
		t.Error("Notification with reserved method should fail validation")
	}

	// Test notification with empty method
	notif.Method = ""
	if err := notif.Validate(); err == nil {
		t.Error("Notification with empty method should fail validation")
	}
}

func TestBindParamsEdgeCases(t *testing.T) {
	// Test BindParams with nil params
	req := &Request{Params: nil}
	var target map[string]any
	if err := req.BindParams(&target); err != nil {
		t.Errorf("BindParams with nil params should not error: %v", err)
	}

	// Test BindParams with marshal error (circular reference)
	type circular struct {
		Self *circular `json:"self"`
	}
	c := &circular{}
	c.Self = c
	req.Params = c

	if err := req.BindParams(&target); err == nil {
		t.Error("BindParams with circular reference should error")
	}
}

func TestErrorValidation(t *testing.T) {
	// Test valid error
	err := &Error{Code: -32600, Message: "Invalid Request"}
	if validationErr := err.Validate(); validationErr != nil {
		t.Errorf("Valid error failed validation: %v", validationErr)
	}

	// Test error with empty message
	err.Message = ""
	if validationErr := err.Validate(); validationErr == nil {
		t.Error("Error with empty message should fail validation")
	}
}

func TestNewStandardErrorEdgeCase(t *testing.T) {
	// Test with unknown error code
	err := NewStandardError(9999, "test data")
	if err.Message != "Unknown error" {
		t.Errorf("NewStandardError with unknown code should use 'Unknown error', got %q", err.Message)
	}
}

func TestResponseValidationEdgeCases(t *testing.T) {
	// Test response with invalid version
	resp := &Response{Version: "1.0", Result: "test", ID: 1}
	if err := resp.Validate(); err == nil {
		t.Error("Response with invalid version should fail validation")
	}

	// Test response with invalid error
	resp = &Response{
		Version: "2.0",
		Error:   &Error{Code: -32600, Message: ""}, // Empty message
		ID:      1,
	}
	if err := resp.Validate(); err == nil {
		t.Error("Response with invalid error should fail validation")
	}
}

func TestNotificationValidationEdgeCases(t *testing.T) {
	// Test notification with invalid version
	notif := &Notification{Version: "1.0", Method: "test"}
	if err := notif.Validate(); err == nil {
		t.Error("Notification with invalid version should fail validation")
	}
}

func TestParseMessageEdgeCases(t *testing.T) {
	// Test message with invalid jsonrpc field type
	invalidVersionJSON := `{"jsonrpc":2.0,"method":"test","id":1}`
	_, err := ParseMessage([]byte(invalidVersionJSON))
	if err == nil {
		t.Error("ParseMessage with invalid jsonrpc field type should fail")
	}

	// Test response with invalid format during unmarshaling
	invalidResponseJSON := `{"jsonrpc":"2.0","result":"test","id":1,"invalid":}`
	_, err = ParseMessage([]byte(invalidResponseJSON))
	if err == nil {
		t.Error("ParseMessage with invalid response format should fail")
	}
}

func TestBindParamsUnmarshalError(t *testing.T) {
	// Test BindParams with unmarshal error
	req := &Request{Params: "invalid for struct"}
	var target struct {
		Number int `json:"number"`
	}
	if err := req.BindParams(&target); err == nil {
		t.Error("BindParams with unmarshal error should fail")
	}
}

func TestParseBatchEdgeCases(t *testing.T) {
	// Test batch with parse error in one message
	batchWithError := `[
		{"jsonrpc":"2.0","method":"test","id":1},
		{"jsonrpc":"1.0","method":"test","id":2}
	]`
	messages, err := Parse([]byte(batchWithError))
	if err != nil {
		t.Errorf("Parse batch with error should not fail: %v", err)
	}
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
	// Second message should be an error response
	if resp, ok := messages[1].(*Response); !ok || resp.Error == nil {
		t.Error("Second message should be an error response")
	}
}

func TestValidateIDTypes(t *testing.T) {
	// Test all valid ID types for Request validation
	validIDs := []any{
		"string-id",
		int(123),
		int32(123),
		int64(123),
		uint(123),
		uint32(123),
		uint64(123),
		float64(123.45),
		nil,
	}

	for _, id := range validIDs {
		req := &Request{Version: "2.0", Method: "test", ID: id}
		if err := req.Validate(); err != nil {
			t.Errorf("Request with ID type %T should be valid: %v", id, err)
		}
	}
}

func TestParseMessageCompleteEdgeCases(t *testing.T) {
	// Test notification parsing with invalid format during unmarshaling
	invalidNotifJSON := `{"jsonrpc":"2.0","method":"test","params":}`
	_, err := ParseMessage([]byte(invalidNotifJSON))
	if err == nil {
		t.Error("ParseMessage with invalid notification format should fail")
	}

	// Test request parsing with invalid format during unmarshaling
	invalidReqJSON := `{"jsonrpc":"2.0","method":"test","id":1,"params":}`
	_, err = ParseMessage([]byte(invalidReqJSON))
	if err == nil {
		t.Error("ParseMessage with invalid request format should fail")
	}
}

func TestParseBatchCompleteEdgeCases(t *testing.T) {
	// Test batch with invalid JSON array format
	invalidBatchJSON := `[{"jsonrpc":"2.0","method":"test","id":1},`
	_, err := Parse([]byte(invalidBatchJSON))
	if err == nil {
		t.Error("Parse with invalid batch JSON should fail")
	}
}

func TestResponseValidationComplete(t *testing.T) {
	// Test response validation with invalid version in error validation path
	resp := &Response{
		Version: "2.0",
		Error:   &Error{Code: -32600, Message: "test"},
		ID:      1,
	}
	// Force error validation by making error invalid
	resp.Error.Message = ""
	if err := resp.Validate(); err == nil {
		t.Error("Response with invalid error should fail validation")
	}
}

func TestNotificationValidationComplete(t *testing.T) {
	// Test notification with invalid version in validation
	notif := &Notification{Version: "1.0", Method: "test"}
	if err := notif.Validate(); err == nil {
		t.Error("Notification with wrong version should fail")
	}
}

// Benchmarks for parser performance
func BenchmarkParseRequest(b *testing.B) {
	requestJSON := []byte(`{"jsonrpc":"2.0","method":"test_method","params":{"key":"value"},"id":1}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseMessage(requestJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseResponse(b *testing.B) {
	responseJSON := []byte(`{"jsonrpc":"2.0","result":{"success":true},"id":1}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseMessage(responseJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalRequest(b *testing.B) {
	req := NewRequest("test_method", map[string]any{"key": "value"}, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalResponse(b *testing.B) {
	resp := NewResponse(map[string]any{"success": true}, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(resp)
		if err != nil {
			b.Fatal(err)
		}
	}
}
