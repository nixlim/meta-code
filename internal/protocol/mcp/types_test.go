package mcp

import (
	"encoding/json"
	"testing"
)

func TestProtocolVersion_Compare(t *testing.T) {
	tests := []struct {
		name     string
		v1       ProtocolVersion
		v2       ProtocolVersion
		expected int
	}{
		{
			name:     "equal versions",
			v1:       "2024-11-05",
			v2:       "2024-11-05",
			expected: 0,
		},
		{
			name:     "v1 less than v2",
			v1:       "2024-11-04",
			v2:       "2024-11-05",
			expected: -1,
		},
		{
			name:     "v1 greater than v2",
			v1:       "2024-11-06",
			v2:       "2024-11-05",
			expected: 1,
		},
		{
			name:     "different years",
			v1:       "2023-12-31",
			v2:       "2024-01-01",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Compare(tt.v2)
			if result != tt.expected {
				t.Errorf("Compare() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestProtocolVersion_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		version ProtocolVersion
		valid   bool
	}{
		{
			name:    "valid version",
			version: "2024-11-05",
			valid:   true,
		},
		{
			name:    "invalid format - no dashes",
			version: "20241105",
			valid:   false,
		},
		{
			name:    "invalid format - too many parts",
			version: "2024-11-05-01",
			valid:   false,
		},
		{
			name:    "invalid year - too early",
			version: "2023-11-05",
			valid:   false,
		},
		{
			name:    "invalid year - too late",
			version: "2031-11-05",
			valid:   false,
		},
		{
			name:    "invalid month",
			version: "2024-13-05",
			valid:   false,
		},
		{
			name:    "invalid day",
			version: "2024-11-32",
			valid:   false,
		},
		{
			name:    "non-numeric parts",
			version: "abcd-ef-gh",
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.IsValid()
			if result != tt.valid {
				t.Errorf("IsValid() = %v, want %v", result, tt.valid)
			}
		})
	}
}

func TestInitializeRequest_Validate(t *testing.T) {
	validParams := InitializeParams{
		ProtocolVersion: "2024-11-05",
		ClientInfo: ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
		Capabilities: Capabilities{},
	}

	tests := []struct {
		name    string
		request *InitializeRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &InitializeRequest{
				Request: NewInitializeRequest(validParams, 1).Request,
				Params:  validParams,
			},
			wantErr: false,
		},
		{
			name: "invalid protocol version",
			request: &InitializeRequest{
				Request: NewInitializeRequest(InitializeParams{
					ProtocolVersion: "invalid",
					ClientInfo:      validParams.ClientInfo,
					Capabilities:    validParams.Capabilities,
				}, 1).Request,
				Params: InitializeParams{
					ProtocolVersion: "invalid",
					ClientInfo:      validParams.ClientInfo,
					Capabilities:    validParams.Capabilities,
				},
			},
			wantErr: true,
		},
		{
			name: "missing client name",
			request: &InitializeRequest{
				Request: NewInitializeRequest(InitializeParams{
					ProtocolVersion: validParams.ProtocolVersion,
					ClientInfo: ClientInfo{
						Name:    "",
						Version: "1.0.0",
					},
					Capabilities: validParams.Capabilities,
				}, 1).Request,
				Params: InitializeParams{
					ProtocolVersion: validParams.ProtocolVersion,
					ClientInfo: ClientInfo{
						Name:    "",
						Version: "1.0.0",
					},
					Capabilities: validParams.Capabilities,
				},
			},
			wantErr: true,
		},
		{
			name: "missing client version",
			request: &InitializeRequest{
				Request: NewInitializeRequest(InitializeParams{
					ProtocolVersion: validParams.ProtocolVersion,
					ClientInfo: ClientInfo{
						Name:    "test-client",
						Version: "",
					},
					Capabilities: validParams.Capabilities,
				}, 1).Request,
				Params: InitializeParams{
					ProtocolVersion: validParams.ProtocolVersion,
					ClientInfo: ClientInfo{
						Name:    "test-client",
						Version: "",
					},
					Capabilities: validParams.Capabilities,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInitializeResponse_Validate(t *testing.T) {
	validResult := InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo: ServerInfo{
			Name:    "test-server",
			Version: "1.0.0",
		},
		Capabilities: Capabilities{},
	}

	tests := []struct {
		name     string
		response *InitializeResponse
		wantErr  bool
	}{
		{
			name: "valid response",
			response: &InitializeResponse{
				Response: NewInitializeResponse(validResult, 1).Response,
				Result:   &validResult,
			},
			wantErr: false,
		},
		{
			name: "nil result",
			response: &InitializeResponse{
				Response: NewInitializeResponse(validResult, 1).Response,
				Result:   nil,
			},
			wantErr: true,
		},
		{
			name: "invalid protocol version",
			response: &InitializeResponse{
				Response: NewInitializeResponse(InitializeResult{
					ProtocolVersion: "invalid",
					ServerInfo:      validResult.ServerInfo,
					Capabilities:    validResult.Capabilities,
				}, 1).Response,
				Result: &InitializeResult{
					ProtocolVersion: "invalid",
					ServerInfo:      validResult.ServerInfo,
					Capabilities:    validResult.Capabilities,
				},
			},
			wantErr: true,
		},
		{
			name: "missing server name",
			response: &InitializeResponse{
				Response: NewInitializeResponse(InitializeResult{
					ProtocolVersion: validResult.ProtocolVersion,
					ServerInfo: ServerInfo{
						Name:    "",
						Version: "1.0.0",
					},
					Capabilities: validResult.Capabilities,
				}, 1).Response,
				Result: &InitializeResult{
					ProtocolVersion: validResult.ProtocolVersion,
					ServerInfo: ServerInfo{
						Name:    "",
						Version: "1.0.0",
					},
					Capabilities: validResult.Capabilities,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.response.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSON_Marshaling(t *testing.T) {
	// Test InitializeRequest marshaling
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		ClientInfo: ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
		Capabilities: Capabilities{
			Resources: &ResourcesCapability{
				Subscribe:   true,
				ListChanged: true,
			},
		},
	}

	req := NewInitializeRequest(params, "test-id")
	
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal InitializeRequest: %v", err)
	}

	var unmarshaled InitializeRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal InitializeRequest: %v", err)
	}

	if unmarshaled.Params.ProtocolVersion != params.ProtocolVersion {
		t.Errorf("ProtocolVersion mismatch: got %s, want %s", 
			unmarshaled.Params.ProtocolVersion, params.ProtocolVersion)
	}

	if unmarshaled.Params.ClientInfo.Name != params.ClientInfo.Name {
		t.Errorf("ClientInfo.Name mismatch: got %s, want %s",
			unmarshaled.Params.ClientInfo.Name, params.ClientInfo.Name)
	}
}

func TestMCPError(t *testing.T) {
	// Test NewMCPError with known error code
	err := NewMCPError(ErrorCodeResourceNotFound, "test-resource")
	if err.Code != ErrorCodeResourceNotFound {
		t.Errorf("Expected error code %d, got %d", ErrorCodeResourceNotFound, err.Code)
	}
	if err.Message != "Resource not found" {
		t.Errorf("Expected message 'Resource not found', got %s", err.Message)
	}

	// Test NewMCPError with unknown error code
	unknownErr := NewMCPError(99999, "test-data")
	if unknownErr.Message != "Unknown MCP error" {
		t.Errorf("Expected message 'Unknown MCP error', got %s", unknownErr.Message)
	}

	// Test specific error constructors
	resourceErr := NewResourceNotFoundError("my-resource")
	if resourceErr.Code != ErrorCodeResourceNotFound {
		t.Errorf("Expected resource error code %d, got %d", ErrorCodeResourceNotFound, resourceErr.Code)
	}

	toolErr := NewToolNotFoundError("my-tool")
	if toolErr.Code != ErrorCodeToolNotFound {
		t.Errorf("Expected tool error code %d, got %d", ErrorCodeToolNotFound, toolErr.Code)
	}

	promptErr := NewPromptNotFoundError("my-prompt")
	if promptErr.Code != ErrorCodePromptNotFound {
		t.Errorf("Expected prompt error code %d, got %d", ErrorCodePromptNotFound, promptErr.Code)
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		name           string
		clientVersion  ProtocolVersion
		serverVersion  ProtocolVersion
		expectedCompat bool
	}{
		{
			name:           "same versions",
			clientVersion:  "2024-11-05",
			serverVersion:  "2024-11-05",
			expectedCompat: true,
		},
		{
			name:           "different versions",
			clientVersion:  "2024-11-04",
			serverVersion:  "2024-11-05",
			expectedCompat: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCompatible(tt.clientVersion, tt.serverVersion)
			if result != tt.expectedCompat {
				t.Errorf("IsCompatible() = %v, want %v", result, tt.expectedCompat)
			}
		})
	}
}

func TestIsValidMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		valid  bool
	}{
		{
			name:   "valid method - initialize",
			method: MethodInitialize,
			valid:  true,
		},
		{
			name:   "valid method - list resources",
			method: MethodListResources,
			valid:  true,
		},
		{
			name:   "invalid method",
			method: "invalid/method",
			valid:  false,
		},
		{
			name:   "empty method",
			method: "",
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidMethod(tt.method)
			if result != tt.valid {
				t.Errorf("IsValidMethod() = %v, want %v", result, tt.valid)
			}
		})
	}
}

func TestGetSupportedMethods(t *testing.T) {
	methods := GetSupportedMethods()

	// Check that we have a reasonable number of methods
	if len(methods) < 10 {
		t.Errorf("Expected at least 10 supported methods, got %d", len(methods))
	}

	// Check that initialize method is included
	found := false
	for _, method := range methods {
		if method == MethodInitialize {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected MethodInitialize to be in supported methods")
	}
}

func TestInitializedNotification(t *testing.T) {
	notif := NewInitializedNotification()

	if err := notif.Validate(); err != nil {
		t.Errorf("InitializedNotification validation failed: %v", err)
	}

	if notif.Method != MethodInitialized {
		t.Errorf("Expected method %s, got %s", MethodInitialized, notif.Method)
	}
}
