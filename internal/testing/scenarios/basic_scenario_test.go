package scenarios

import (
	"sync"
	"testing"
	"time"
)

// TestBasicContextLifecycle tests the complete lifecycle of a context
func TestBasicContextLifecycle(t *testing.T) {
	// Scenario: Create, Read, Update, Delete a context

	// Step 1: Create a new context
	t.Run("CreateContext", func(t *testing.T) {
		// Create context with basic information
		contextData := map[string]interface{}{
			"name": "test-context-lifecycle",
			"type": "project",
			"metadata": map[string]string{
				"description": "Testing context lifecycle",
				"owner":       "test-suite",
			},
		}

		// Simulate context creation
		contextID := createTestContext(t, contextData)
		if contextID == "" {
			t.Fatal("Failed to create context")
		}
	})

	// Step 2: Read the created context
	t.Run("ReadContext", func(t *testing.T) {
		// Retrieve context by ID
		context := getTestContext(t, "test-context-lifecycle")
		if context == nil {
			t.Fatal("Failed to retrieve context")
		}

		// Validate context data
		validateContextData(t, context)
	})

	// Step 3: Update the context
	t.Run("UpdateContext", func(t *testing.T) {
		// Update context metadata
		updates := map[string]interface{}{
			"metadata": map[string]string{
				"status":  "active",
				"version": "1.0.1",
			},
		}

		// Apply updates
		success := updateTestContext(t, "test-context-lifecycle", updates)
		if !success {
			t.Fatal("Failed to update context")
		}
	})

	// Step 4: Delete the context
	t.Run("DeleteContext", func(t *testing.T) {
		// Delete context
		success := deleteTestContext(t, "test-context-lifecycle")
		if !success {
			t.Fatal("Failed to delete context")
		}

		// Verify deletion
		context := getTestContext(t, "test-context-lifecycle")
		if context != nil {
			t.Fatal("Context still exists after deletion")
		}
	})
}

// TestConcurrentContextOperations tests concurrent access to contexts
func TestConcurrentContextOperations(t *testing.T) {
	// Scenario: Multiple goroutines operating on contexts simultaneously

	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	// Launch concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Each goroutine creates and manages its own context
			contextName := "concurrent-context-" + string(rune(id))

			// Create
			createTestContext(t, map[string]interface{}{
				"name": contextName,
				"type": "concurrent-test",
			})

			// Read
			getTestContext(t, contextName)

			// Update
			updateTestContext(t, contextName, map[string]interface{}{
				"status": "processed",
			})

			// Delete
			deleteTestContext(t, contextName)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// Simple in-memory context store for testing
var (
	testContextStore = make(map[string]map[string]interface{})
	testContextMutex sync.Mutex
)

// Helper functions with basic implementation
func createTestContext(t *testing.T, data map[string]interface{}) string {
	testContextMutex.Lock()
	defer testContextMutex.Unlock()
	
	name, ok := data["name"].(string)
	if !ok {
		name = "ctx-" + time.Now().Format("20060102150405")
	}
	
	testContextStore[name] = data
	return name
}

func getTestContext(t *testing.T, name string) map[string]interface{} {
	testContextMutex.Lock()
	defer testContextMutex.Unlock()
	
	context, exists := testContextStore[name]
	if !exists {
		return nil
	}
	
	// Return a copy to prevent external modifications
	result := make(map[string]interface{})
	for k, v := range context {
		result[k] = v
	}
	return result
}

func updateTestContext(t *testing.T, name string, updates map[string]interface{}) bool {
	testContextMutex.Lock()
	defer testContextMutex.Unlock()
	
	context, exists := testContextStore[name]
	if !exists {
		return false
	}
	
	// Apply updates
	for k, v := range updates {
		context[k] = v
	}
	return true
}

func deleteTestContext(t *testing.T, name string) bool {
	testContextMutex.Lock()
	defer testContextMutex.Unlock()
	
	_, exists := testContextStore[name]
	if !exists {
		return false
	}
	
	delete(testContextStore, name)
	return true
}

func validateContextData(t *testing.T, context map[string]interface{}) {
	// Basic validation
	if context == nil {
		t.Fatal("Context is nil")
	}
	if _, ok := context["name"]; !ok {
		t.Fatal("Context missing required 'name' field")
	}
}
