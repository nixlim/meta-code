package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// TestNotificationConformance tests notification message conformance
func (suite *ConformanceTestSuite) TestNotificationConformance(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name        string
		method      string
		params      json.RawMessage
		shouldPass  bool
		description string
	}{
		// Valid notifications
		{
			name:        "initialized_notification",
			method:      "initialized",
			params:      nil,
			shouldPass:  true,
			description: "Valid initialized notification without params",
		},
		{
			name:        "cancelled_notification",
			method:      "cancelled",
			params:      json.RawMessage(`{"requestId": "123"}`),
			shouldPass:  true,
			description: "Valid cancelled notification with request ID",
		},
		{
			name:        "progress_notification",
			method:      "progress",
			params:      json.RawMessage(`{"token": "abc", "progress": 50, "total": 100}`),
			shouldPass:  true,
			description: "Valid progress notification with progress info",
		},
		{
			name:        "resources_updated",
			method:      "resources/updated",
			params:      json.RawMessage(`{"uri": "file:///test.txt"}`),
			shouldPass:  true,
			description: "Valid resources/updated notification",
		},
		{
			name:        "resources_list_changed",
			method:      "resources/listChanged",
			params:      nil,
			shouldPass:  true,
			description: "Valid resources/listChanged notification",
		},
		{
			name:        "tools_list_changed",
			method:      "tools/listChanged",
			params:      nil,
			shouldPass:  true,
			description: "Valid tools/listChanged notification",
		},
		{
			name:        "prompts_list_changed",
			method:      "prompts/listChanged",
			params:      nil,
			shouldPass:  true,
			description: "Valid prompts/listChanged notification",
		},
		{
			name:        "message_notification",
			method:      "message",
			params:      json.RawMessage(`{"level": "info", "logger": "test", "data": "Test message"}`),
			shouldPass:  true,
			description: "Valid message notification with log data",
		},
		// Invalid notifications
		{
			name:        "invalid_method",
			method:      "invalid/notification",
			params:      nil,
			shouldPass:  false,
			description: "Notification with invalid method name",
		},
		{
			name:        "empty_method",
			method:      "",
			params:      nil,
			shouldPass:  false,
			description: "Notification with empty method name",
		},
		{
			name:        "malformed_params",
			method:      "progress",
			params:      json.RawMessage(`{invalid json}`),
			shouldPass:  false,
			description: "Notification with malformed parameters",
		},
		// Edge cases
		{
			name:        "notification_with_empty_params",
			method:      "progress",
			params:      json.RawMessage(`{}`),
			shouldPass:  true,
			description: "Valid notification with empty params object",
		},
		{
			name:        "notification_with_array_params",
			method:      "message",
			params:      json.RawMessage(`["info", "test", "message"]`),
			shouldPass:  false,  // MCP spec requires object params, not arrays
			description: "Invalid notification with array parameters (MCP requires object params)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := suite.validator.ValidateNotification(ctx, tt.method, tt.params)
			passed := (err == nil) == tt.shouldPass
			
			result := TestResult{
				TestName:    fmt.Sprintf("notification_%s", tt.name),
				Category:    "Notifications",
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
	
	// Test that notifications don't have ID field
	t.Run("NotificationNoID", func(t *testing.T) {
		notificationWithID := json.RawMessage(`{
			"jsonrpc": "2.0",
			"method": "progress",
			"params": {"progress": 50},
			"id": "123"
		}`)
		
		// This should fail because notifications shouldn't have an ID
		err := suite.validator.ValidateMessage(ctx, "notification", notificationWithID)
		passed := err != nil // We expect an error
		
		result := TestResult{
			TestName:    "notification_with_id_rejected",
			Category:    "Notifications",
			Description: "Notification with ID field should be rejected",
			Passed:      passed,
		}
		
		if err != nil {
			result.Details = fmt.Sprintf("Correctly rejected notification with ID: %v", err)
		} else {
			result.Error = "Failed to reject notification with ID field"
		}
		
		suite.recordResult(result)
		
		if !passed {
			t.Error("Notification with ID field should be rejected")
		}
	})
	
	// Test notification batching
	t.Run("NotificationBatching", func(t *testing.T) {
		// Test that we can validate multiple notifications in sequence
		notifications := []struct {
			method string
			params json.RawMessage
		}{
			{"progress", json.RawMessage(`{"progress": 25}`)},
			{"progress", json.RawMessage(`{"progress": 50}`)},
			{"progress", json.RawMessage(`{"progress": 75}`)},
			{"progress", json.RawMessage(`{"progress": 100}`)},
		}
		
		allPassed := true
		for i, n := range notifications {
			err := suite.validator.ValidateNotification(ctx, n.method, n.params)
			if err != nil {
				allPassed = false
				t.Errorf("Notification %d failed: %v", i, err)
			}
		}
		
		result := TestResult{
			TestName:    "notification_sequence_validation",
			Category:    "Notifications",
			Description: "Validate sequence of progress notifications",
			Passed:      allPassed,
		}
		
		if allPassed {
			result.Details = "Successfully validated sequence of 4 progress notifications"
		} else {
			result.Error = "Failed to validate notification sequence"
		}
		
		suite.recordResult(result)
	})
}