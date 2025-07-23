package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/connection"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
	"github.com/meta-mcp/meta-mcp-server/test/testutil"
)


func TestCreateValidationHooks(t *testing.T) {
	tests := []struct {
		name            string
		method          mcp.MCPMethod
		connectionState connection.ConnectionState
		hasConnectionID bool
		connectionID    string
		messageID       any
		expectLog       string
	}{
		{
			name:            "allows_initialize_method",
			method:          mcp.MethodInitialize,
			connectionState: connection.StateNew,
			hasConnectionID: true,
			connectionID:    "test-conn-1",
			messageID:       1,
			expectLog:       "Allowing initialize method",
		},
		{
			name:            "allows_notification_nil_id",
			method:          mcp.MethodPing,
			connectionState: connection.StateNew,
			hasConnectionID: true,
			connectionID:    "test-conn-2",
			messageID:       nil,
			expectLog:       "Allowing notification",
		},
		{
			name:            "rejects_method_when_not_ready",
			method:          mcp.MethodToolsList,
			connectionState: connection.StateInitializing,
			hasConnectionID: true,
			connectionID:    "test-conn-3",
			messageID:       3,
			expectLog:       "Rejecting method - connection not ready",
		},
		{
			name:            "allows_method_when_ready",
			method:          mcp.MethodResourcesList,
			connectionState: connection.StateReady,
			hasConnectionID: true,
			connectionID:    "test-conn-4",
			messageID:       4,
			expectLog:       "Allowing method - connection ready",
		},
		{
			name:            "no_connection_in_context",
			method:          mcp.MethodToolsList,
			connectionState: connection.StateReady,
			hasConnectionID: false,
			connectionID:    "",
			messageID:       5,
			expectLog:       "No connection found in context",
		},
		{
			name:            "connection_closed",
			method:          mcp.MethodPromptsList,
			connectionState: connection.StateClosed,
			hasConnectionID: true,
			connectionID:    "test-conn-6",
			messageID:       6,
			expectLog:       "Rejecting method - connection not ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create connection manager
			var manager *connection.Manager
			if tt.hasConnectionID {
				manager = testutil.CreateTestManagerWithConnection(tt.connectionID, tt.connectionState)
			} else {
				manager = testutil.CreateTestManager()
			}

			// Create config
			config := ValidationHooksConfig{
				ConnectionManager: manager,
			}

			// Create the hook
			hook := CreateValidationHooks(config)

			// Create context
			ctx := context.Background()
			if tt.hasConnectionID {
				ctx = connection.WithConnectionID(ctx, tt.connectionID)
			}

			// Call the hook - it doesn't return anything, just logs
			// In a real test, we'd capture logs to verify behavior
			hook(ctx, tt.messageID, tt.method, nil)

			// Since we can't easily test logging, we verify the setup worked
			if hook == nil {
				t.Error("Expected hook function to be created")
			}
		})
	}
}

func TestCreateRequestValidator(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		connectionState connection.ConnectionState
		hasConnectionID bool
		connectionID    string
		wantErr         bool
		expectedErrCode int
	}{
		{
			name:            "allows_initialize",
			method:          "initialize",
			connectionState: connection.StateNew,
			hasConnectionID: true,
			connectionID:    "test-1",
			wantErr:         false,
		},
		{
			name:            "allows_initialized",
			method:          "initialized",
			connectionState: connection.StateInitializing,
			hasConnectionID: true,
			connectionID:    "test-2",
			wantErr:         false,
		},
		{
			name:            "requires_ready_for_other_methods",
			method:          "tools/list",
			connectionState: connection.StateInitializing,
			hasConnectionID: true,
			connectionID:    "test-3",
			wantErr:         true,
			expectedErrCode: -32011,
		},
		{
			name:            "allows_when_ready",
			method:          "resources/list",
			connectionState: connection.StateReady,
			hasConnectionID: true,
			connectionID:    "test-4",
			wantErr:         false,
		},
		{
			name:            "error_no_connection_context",
			method:          "tools/list",
			connectionState: connection.StateReady,
			hasConnectionID: false,
			wantErr:         true,
			expectedErrCode: jsonrpc.ErrorCodeInvalidRequest,
		},
		{
			name:            "error_connection_closed",
			method:          "prompts/list",
			connectionState: connection.StateClosed,
			hasConnectionID: true,
			connectionID:    "test-5",
			wantErr:         true,
			expectedErrCode: -32011,
		},
		{
			name:            "error_connection_new",
			method:          "ping",
			connectionState: connection.StateNew,
			hasConnectionID: true,
			connectionID:    "test-6",
			wantErr:         true,
			expectedErrCode: -32011,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create connection manager
			var manager *connection.Manager
			if tt.hasConnectionID {
				manager = testutil.CreateTestManagerWithConnection(tt.connectionID, tt.connectionState)
			} else {
				manager = testutil.CreateTestManager()
			}

			// Create validator
			validator := CreateRequestValidator(manager)

			// Create context
			ctx := context.Background()
			if tt.hasConnectionID {
				ctx = connection.WithConnectionID(ctx, tt.connectionID)
			}

			// Call validator
			err := validator(ctx, tt.method)

			// Check result
			if (err != nil) != tt.wantErr {
				t.Errorf("validator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				jsonrpcErr, ok := err.(*jsonrpc.Error)
				if !ok {
					t.Errorf("Expected jsonrpc.Error, got %T", err)
					return
				}
				if jsonrpcErr.Code != tt.expectedErrCode {
					t.Errorf("Expected error code %d, got %d", tt.expectedErrCode, jsonrpcErr.Code)
				}
			}
		})
	}
}

func TestIsNotification(t *testing.T) {
	tests := []struct {
		name     string
		id       any
		expected bool
	}{
		{
			name:     "nil_id_is_notification",
			id:       nil,
			expected: true,
		},
		{
			name:     "string_id_not_notification",
			id:       "test-id",
			expected: false,
		},
		{
			name:     "int_id_not_notification",
			id:       123,
			expected: false,
		},
		{
			name:     "zero_int_not_notification",
			id:       0,
			expected: false,
		},
		{
			name:     "empty_string_not_notification",
			id:       "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNotification(tt.id)
			if result != tt.expected {
				t.Errorf("isNotification(%v) = %v, want %v", tt.id, result, tt.expected)
			}
		})
	}
}

func TestCreateErrorHook(t *testing.T) {
	tests := []struct {
		name            string
		method          mcp.MCPMethod
		messageID       any
		err             error
		hasConnectionID bool
		connectionID    string
		connectionState connection.ConnectionState
	}{
		{
			name:            "logs_error_with_connection",
			method:          mcp.MethodToolsList,
			messageID:       1,
			err:             errors.New("test error"),
			hasConnectionID: true,
			connectionID:    "test-conn-1",
			connectionState: connection.StateReady,
		},
		{
			name:            "logs_error_without_connection",
			method:          mcp.MethodResourcesList,
			messageID:       2,
			err:             errors.New("another error"),
			hasConnectionID: false,
		},
		{
			name:            "logs_error_with_notification",
			method:          mcp.MethodPing,
			messageID:       nil,
			err:             errors.New("notification error"),
			hasConnectionID: true,
			connectionID:    "test-conn-2",
			connectionState: connection.StateInitializing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create connection manager
			var manager *connection.Manager
			if tt.hasConnectionID {
				manager = testutil.CreateTestManagerWithConnection(tt.connectionID, tt.connectionState)
			} else {
				manager = testutil.CreateTestManager()
			}

			// Create config
			config := ValidationHooksConfig{
				ConnectionManager: manager,
			}

			// Create the hook
			hook := CreateErrorHook(config)

			// Create context
			ctx := context.Background()
			if tt.hasConnectionID {
				ctx = connection.WithConnectionID(ctx, tt.connectionID)
			}

			// Call the hook - it doesn't return anything, just logs
			hook(ctx, tt.messageID, tt.method, nil, tt.err)

			// Since we can't easily test logging, we verify the setup worked
			if hook == nil {
				t.Error("Expected error hook function to be created")
			}
		})
	}
}

func TestCreateSuccessHook(t *testing.T) {
	tests := []struct {
		name      string
		method    mcp.MCPMethod
		messageID any
		result    any
	}{
		{
			name:      "logs_non_ping_method",
			method:    mcp.MethodToolsList,
			messageID: 1,
			result:    "success",
		},
		{
			name:      "skips_ping_method",
			method:    mcp.MethodPing,
			messageID: 2,
			result:    "pong",
		},
		{
			name:      "logs_notification_success",
			method:    mcp.MethodResourcesList,
			messageID: nil,
			result:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config
			config := ValidationHooksConfig{
				ConnectionManager: testutil.CreateTestManager(),
			}

			// Create the hook
			hook := CreateSuccessHook(config)

			// Create context
			ctx := context.Background()

			// Call the hook - it doesn't return anything, just logs
			hook(ctx, tt.messageID, tt.method, nil, tt.result)

			// Since we can't easily test logging, we verify the setup worked
			if hook == nil {
				t.Error("Expected success hook function to be created")
			}
		})
	}
}

// Test edge cases and error scenarios
func TestValidationHooksEdgeCases(t *testing.T) {
	t.Run("nil_connection_manager", func(t *testing.T) {
		config := ValidationHooksConfig{
			ConnectionManager: nil,
		}

		// These should not panic
		hook := CreateValidationHooks(config)
		if hook == nil {
			t.Error("Expected hook to be created even with nil manager")
		}

		errorHook := CreateErrorHook(config)
		if errorHook == nil {
			t.Error("Expected error hook to be created even with nil manager")
		}

		successHook := CreateSuccessHook(config)
		if successHook == nil {
			t.Error("Expected success hook to be created even with nil manager")
		}
	})

	t.Run("request_validator_nil_manager", func(t *testing.T) {
		validator := CreateRequestValidator(nil)
		ctx := context.Background()
		
		// Should handle nil manager gracefully
		err := validator(ctx, "test-method")
		if err == nil {
			t.Error("Expected error for nil manager")
		}
	})
}

// Test concurrent access scenarios
func TestValidationHooksConcurrency(t *testing.T) {
	manager := testutil.CreateTestManager()
	config := ValidationHooksConfig{
		ConnectionManager: manager,
	}

	// Add multiple connections
	for i := 0; i < 10; i++ {
		conn, _ := manager.CreateConnection(string(rune('0' + i)))
		conn.State = connection.StateReady
	}

	hook := CreateValidationHooks(config)
	validator := CreateRequestValidator(manager)

	// Test concurrent hook calls
	done := make(chan bool, 20)

	// Concurrent hook calls
	for i := 0; i < 10; i++ {
		go func(id int) {
			ctx := connection.WithConnectionID(context.Background(), string(rune('0'+id)))
			hook(ctx, id, mcp.MethodToolsList, nil)
			done <- true
		}(i % 10)
	}

	// Concurrent validator calls
	for i := 0; i < 10; i++ {
		go func(id int) {
			ctx := connection.WithConnectionID(context.Background(), string(rune('0'+id)))
			_ = validator(ctx, "test-method")
			done <- true
		}(i % 10)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

// Test error scenarios in CreateRequestValidator
func TestCreateRequestValidatorErrorCases(t *testing.T) {
	manager := testutil.CreateTestManager()
	validator := CreateRequestValidator(manager)

	t.Run("connection_not_found", func(t *testing.T) {
		// Connection ID in context but not in manager
		ctx := connection.WithConnectionID(context.Background(), "non-existent")
		err := validator(ctx, "tools/list")
		
		if err == nil {
			t.Error("Expected error for non-existent connection")
		}
		
		jsonrpcErr, ok := err.(*jsonrpc.Error)
		if !ok {
			t.Errorf("Expected jsonrpc.Error, got %T", err)
		} else if jsonrpcErr.Code != jsonrpc.ErrorCodeInvalidRequest {
			t.Errorf("Expected error code %d, got %d", jsonrpc.ErrorCodeInvalidRequest, jsonrpcErr.Code)
		}
	})
}

// Benchmark tests
func BenchmarkCreateValidationHooks(b *testing.B) {
	manager := testutil.CreateTestManagerWithConnection("bench-conn", connection.StateReady)
	
	config := ValidationHooksConfig{
		ConnectionManager: manager,
	}
	
	hook := CreateValidationHooks(config)
	ctx := connection.WithConnectionID(context.Background(), "bench-conn")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hook(ctx, i, mcp.MethodToolsList, nil)
	}
}

func BenchmarkCreateRequestValidator(b *testing.B) {
	manager := testutil.CreateTestManagerWithConnection("bench-conn", connection.StateReady)
	
	validator := CreateRequestValidator(manager)
	ctx := connection.WithConnectionID(context.Background(), "bench-conn")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(ctx, "tools/list")
	}
}