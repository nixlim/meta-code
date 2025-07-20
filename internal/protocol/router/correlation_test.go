package router

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

func TestCorrelationTrackerGenerateID(t *testing.T) {
	ct := NewCorrelationTracker()
	defer ct.Shutdown()

	// Generate multiple IDs and ensure uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := ct.GenerateCorrelationID()
		if ids[id] {
			t.Errorf("Duplicate correlation ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestCorrelationTrackerRegisterAndComplete(t *testing.T) {
	ct := NewCorrelationTracker()
	defer ct.Shutdown()

	correlationID := ct.GenerateCorrelationID()

	// Register correlation
	_, _ = ct.Register(correlationID)

	// Complete with response
	expectedResponse := &jsonrpc.Response{
		ID:     "test-123",
		Result: map[string]interface{}{"status": "ok"},
	}

	err := ct.Complete(correlationID, expectedResponse)
	if err != nil {
		t.Errorf("Failed to complete correlation: %v", err)
	}

	// ✅ FIXED: Use WaitForResponse to properly consume the response and trigger cleanup
	response, err := ct.WaitForResponse(correlationID, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForResponse failed: %v", err)
	}

	if response.ID != expectedResponse.ID {
		t.Errorf("Expected response ID %s, got %s", expectedResponse.ID, response.ID)
	}

	// ✅ FIXED: After WaitForResponse, correlation should be deleted
	// Verify correlation is deleted
	_, err = ct.WaitForResponse(correlationID, 10*time.Millisecond)
	if err == nil {
		t.Error("Expected correlation to be deleted after WaitForResponse")
	}
}

func TestCorrelationTrackerCompleteWithError(t *testing.T) {
	ct := NewCorrelationTracker()
	defer ct.Shutdown()

	correlationID := ct.GenerateCorrelationID()

	// Register correlation
	_, _ = ct.Register(correlationID)

	// Complete with error
	expectedErr := jsonrpc.NewError(jsonrpc.ErrorCodeInternal, "test error", nil)

	err := ct.CompleteWithError(correlationID, expectedErr)
	if err != nil {
		t.Errorf("Failed to complete correlation with error: %v", err)
	}

	// ✅ FIXED: Use WaitForResponse to properly consume the error and trigger cleanup
	_, err = ct.WaitForResponse(correlationID, 100*time.Millisecond)
	if err == nil {
		t.Error("Expected WaitForResponse to return an error")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// ✅ FIXED: After WaitForResponse, correlation should be deleted
	// Verify correlation is deleted
	_, err = ct.WaitForResponse(correlationID, 10*time.Millisecond)
	if err == nil {
		t.Error("Expected correlation to be deleted after WaitForResponse")
	}
}

func TestCorrelationTrackerCancel(t *testing.T) {
	ct := NewCorrelationTracker()
	defer ct.Shutdown()

	correlationID := ct.GenerateCorrelationID()

	// Register correlation
	respChan, errChan := ct.Register(correlationID)

	// Cancel correlation
	ct.Cancel(correlationID)

	// Verify channels are closed
	select {
	case _, ok := <-respChan:
		if ok {
			t.Error("Expected response channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Response channel not closed")
	}

	select {
	case _, ok := <-errChan:
		if ok {
			t.Error("Expected error channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Error channel not closed")
	}
}

func TestCorrelationTrackerWaitForResponse(t *testing.T) {
	ct := NewCorrelationTracker()
	defer ct.Shutdown()

	t.Run("Success", func(t *testing.T) {
		correlationID := ct.GenerateCorrelationID()
		ct.Register(correlationID)

		expectedResponse := &jsonrpc.Response{
			ID:     "test-123",
			Result: map[string]interface{}{"status": "ok"},
		}

		// Complete in background
		go func() {
			time.Sleep(10 * time.Millisecond)
			ct.Complete(correlationID, expectedResponse)
		}()

		// Wait for response
		response, err := ct.WaitForResponse(correlationID, 100*time.Millisecond)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if response.ID != expectedResponse.ID {
			t.Errorf("Expected response ID %s, got %s", expectedResponse.ID, response.ID)
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		correlationID := ct.GenerateCorrelationID()
		ct.Register(correlationID)

		// Wait for response with short timeout
		_, err := ct.WaitForResponse(correlationID, 10*time.Millisecond)
		if err != ErrCorrelationTimeout {
			t.Errorf("Expected ErrCorrelationTimeout, got %v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		// Try to wait for non-existent correlation
		_, err := ct.WaitForResponse("nonexistent", 10*time.Millisecond)
		if err != ErrCorrelationNotFound {
			t.Errorf("Expected ErrCorrelationNotFound, got %v", err)
		}
	})
}

func TestCorrelationTrackerConcurrentOperations(t *testing.T) {
	ct := NewCorrelationTracker()
	defer ct.Shutdown()

	var wg sync.WaitGroup
	numOperations := 100

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			correlationID := ct.GenerateCorrelationID()
			ct.Register(correlationID)

			// Randomly complete or cancel
			if n%2 == 0 {
				response := &jsonrpc.Response{
					ID: fmt.Sprintf("test-%d", n),
				}
				ct.Complete(correlationID, response)
			} else {
				ct.Cancel(correlationID)
			}
		}(i)
	}

	wg.Wait()

	// Verify no panics and tracker is still functional
	testID := ct.GenerateCorrelationID()
	_, _ = ct.Register(testID)
	err := ct.Complete(testID, &jsonrpc.Response{ID: "final"})
	if err != nil {
		t.Errorf("Tracker not functional after concurrent operations: %v", err)
	}
}

func TestCorrelationTrackerStats(t *testing.T) {
	ct := NewCorrelationTracker()
	defer ct.Shutdown()

	// Initially empty
	stats := ct.Stats()
	if stats.PendingCount != 0 {
		t.Errorf("Expected 0 pending, got %d", stats.PendingCount)
	}

	// Register some correlations
	for i := 0; i < 5; i++ {
		ct.Register(ct.GenerateCorrelationID())
	}

	stats = ct.Stats()
	if stats.PendingCount != 5 {
		t.Errorf("Expected 5 pending, got %d", stats.PendingCount)
	}

	// Complete one
	ct.Complete(ct.GenerateCorrelationID(), &jsonrpc.Response{})

	// Stats should still show 5 (completed one wasn't registered)
	stats = ct.Stats()
	if stats.PendingCount != 5 {
		t.Errorf("Expected 5 pending after completing unregistered ID, got %d", stats.PendingCount)
	}
}

func TestCorrelationTrackerShutdown(t *testing.T) {
	ct := NewCorrelationTracker()

	// Register some correlations
	var channels [](<-chan *jsonrpc.Response)
	for i := 0; i < 5; i++ {
		respChan, _ := ct.Register(ct.GenerateCorrelationID())
		channels = append(channels, respChan)
	}

	// Shutdown
	ct.Shutdown()

	// Verify all channels are closed
	for i, ch := range channels {
		select {
		case _, ok := <-ch:
			if ok {
				t.Errorf("Expected channel %d to be closed after shutdown", i)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Channel %d not closed after shutdown", i)
		}
	}

	// Verify stats show 0 pending
	stats := ct.Stats()
	if stats.PendingCount != 0 {
		t.Errorf("Expected 0 pending after shutdown, got %d", stats.PendingCount)
	}
}
