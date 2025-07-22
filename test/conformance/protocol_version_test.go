package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// TestProtocolVersionConformance tests protocol version negotiation conformance
func (suite *ConformanceTestSuite) TestProtocolVersionConformance(t *testing.T) {
	ctx := context.Background()

	// Test various protocol versions in initialize
	t.Run("InitializeVersions", func(t *testing.T) {
		tests := []struct {
			name        string
			version     string
			shouldPass  bool
			description string
		}{
			{
				name:        "version_1_0",
				version:     "1.0",
				shouldPass:  true,
				description: "Valid protocol version 1.0",
			},
			{
				name:        "version_0_1_0",
				version:     "0.1.0",
				shouldPass:  true,
				description: "Valid protocol version 0.1.0",
			},
			{
				name:        "version_2_0",
				version:     "2.0",
				shouldPass:  true,
				description: "Valid protocol version 2.0",
			},
			{
				name:        "empty_version",
				version:     "",
				shouldPass:  true, // Empty string is technically valid as a string
				description: "Empty protocol version string",
			},
			{
				name:        "semantic_version",
				version:     "1.0.0-beta.1",
				shouldPass:  true,
				description: "Semantic version with pre-release",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				message := json.RawMessage(fmt.Sprintf(`{
					"protocolVersion": "%s",
					"capabilities": {},
					"clientInfo": {
						"name": "test-client",
						"version": "1.0.0"
					}
				}`, tt.version))

				err := suite.validator.ValidateMessage(ctx, "initialize", message)
				passed := (err == nil) == tt.shouldPass

				result := TestResult{
					TestName:    fmt.Sprintf("initialize_version_%s", tt.name),
					Category:    "ProtocolVersion",
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

	// Test version field type validation
	t.Run("VersionFieldTypes", func(t *testing.T) {
		tests := []struct {
			name        string
			message     json.RawMessage
			shouldPass  bool
			description string
		}{
			{
				name: "version_as_number",
				message: json.RawMessage(`{
					"protocolVersion": 1.0,
					"capabilities": {},
					"clientInfo": {
						"name": "test-client",
						"version": "1.0.0"
					}
				}`),
				shouldPass:  false,
				description: "Invalid protocol version as number instead of string",
			},
			{
				name: "version_as_null",
				message: json.RawMessage(`{
					"protocolVersion": null,
					"capabilities": {},
					"clientInfo": {
						"name": "test-client",
						"version": "1.0.0"
					}
				}`),
				shouldPass:  false,
				description: "Invalid protocol version as null",
			},
			{
				name: "version_as_object",
				message: json.RawMessage(`{
					"protocolVersion": {"major": 1, "minor": 0},
					"capabilities": {},
					"clientInfo": {
						"name": "test-client",
						"version": "1.0.0"
					}
				}`),
				shouldPass:  false,
				description: "Invalid protocol version as object",
			},
			{
				name: "missing_version",
				message: json.RawMessage(`{
					"capabilities": {},
					"clientInfo": {
						"name": "test-client",
						"version": "1.0.0"
					}
				}`),
				shouldPass:  false,
				description: "Missing required protocolVersion field",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := suite.validator.ValidateMessage(ctx, "initialize", tt.message)
				passed := (err == nil) == tt.shouldPass

				result := TestResult{
					TestName:    fmt.Sprintf("version_type_%s", tt.name),
					Category:    "ProtocolVersion",
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

	// Test initialized response versions
	t.Run("InitializedVersions", func(t *testing.T) {
		tests := []struct {
			name        string
			message     json.RawMessage
			shouldPass  bool
			description string
		}{
			{
				name: "matching_version",
				message: json.RawMessage(`{
					"protocolVersion": "1.0",
					"capabilities": {},
					"serverInfo": {
						"name": "test-server",
						"version": "1.0.0"
					}
				}`),
				shouldPass:  true,
				description: "Valid initialized with matching protocol version",
			},
			{
				name: "different_version",
				message: json.RawMessage(`{
					"protocolVersion": "0.1.0",
					"capabilities": {},
					"serverInfo": {
						"name": "test-server",
						"version": "1.0.0"
					}
				}`),
				shouldPass:  true,
				description: "Valid initialized with different protocol version (negotiation)",
			},
			{
				name: "server_version_info",
				message: json.RawMessage(`{
					"protocolVersion": "1.0",
					"capabilities": {},
					"serverInfo": {
						"name": "test-server",
						"version": "2.5.0-alpha.1+build.123"
					}
				}`),
				shouldPass:  true,
				description: "Valid initialized with complex server version",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := suite.validator.ValidateMessage(ctx, "initialized", tt.message)
				passed := (err == nil) == tt.shouldPass

				result := TestResult{
					TestName:    fmt.Sprintf("initialized_%s", tt.name),
					Category:    "ProtocolVersion",
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

	// Test client/server info version fields
	t.Run("ClientServerInfo", func(t *testing.T) {
		tests := []struct {
			name        string
			messageType string
			message     json.RawMessage
			shouldPass  bool
			description string
		}{
			{
				name:        "client_info_valid",
				messageType: "initialize",
				message: json.RawMessage(`{
					"protocolVersion": "1.0",
					"capabilities": {},
					"clientInfo": {
						"name": "My MCP Client",
						"version": "1.2.3"
					}
				}`),
				shouldPass:  true,
				description: "Valid client info with name and version",
			},
			{
				name:        "client_info_missing_version",
				messageType: "initialize",
				message: json.RawMessage(`{
					"protocolVersion": "1.0",
					"capabilities": {},
					"clientInfo": {
						"name": "My MCP Client"
					}
				}`),
				shouldPass:  false,
				description: "Invalid client info missing version",
			},
			{
				name:        "client_info_missing_name",
				messageType: "initialize",
				message: json.RawMessage(`{
					"protocolVersion": "1.0",
					"capabilities": {},
					"clientInfo": {
						"version": "1.2.3"
					}
				}`),
				shouldPass:  false,
				description: "Invalid client info missing name",
			},
			{
				name:        "server_info_valid",
				messageType: "initialized",
				message: json.RawMessage(`{
					"protocolVersion": "1.0",
					"capabilities": {},
					"serverInfo": {
						"name": "My MCP Server",
						"version": "2.0.0"
					}
				}`),
				shouldPass:  true,
				description: "Valid server info with name and version",
			},
			{
				name:        "server_info_extra_fields",
				messageType: "initialized",
				message: json.RawMessage(`{
					"protocolVersion": "1.0",
					"capabilities": {},
					"serverInfo": {
						"name": "My MCP Server",
						"version": "2.0.0",
						"extraField": "should be rejected"
					}
				}`),
				shouldPass:  false,
				description: "Invalid server info with extra fields",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := suite.validator.ValidateMessage(ctx, tt.messageType, tt.message)
				passed := (err == nil) == tt.shouldPass

				result := TestResult{
					TestName:    fmt.Sprintf("info_%s", tt.name),
					Category:    "ProtocolVersion",
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
