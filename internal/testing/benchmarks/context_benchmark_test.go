package benchmarks

import (
	"fmt"
	"sync"
	"testing"
)

// BenchmarkContextCreation measures the performance of context creation
func BenchmarkContextCreation(b *testing.B) {
	// Setup test data
	contextData := map[string]interface{}{
		"name": "benchmark-context",
		"type": "test",
		"metadata": map[string]string{
			"created_by": "benchmark",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate context creation
		_ = createContext(contextData)
	}
}

// BenchmarkContextRetrieval measures the performance of context retrieval
func BenchmarkContextRetrieval(b *testing.B) {
	// Setup: create contexts
	contexts := make([]string, 100)
	for i := 0; i < 100; i++ {
		contexts[i] = createTestContext(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate context retrieval
		_ = getContext(contexts[i%100])
	}
}

// BenchmarkConcurrentContextOperations measures concurrent context operations
func BenchmarkConcurrentContextOperations(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate concurrent operations
			performContextOperation()
		}
	})
}

// Simple in-memory context store for benchmarking
var (
	benchContextStore = make(map[string]interface{})
	benchContextMutex sync.RWMutex
	contextCounter    int
)

// Helper functions with actual implementations
func createContext(data map[string]interface{}) string {
	benchContextMutex.Lock()
	defer benchContextMutex.Unlock()
	
	contextCounter++
	id := fmt.Sprintf("ctx-%d", contextCounter)
	benchContextStore[id] = data
	return id
}

func createTestContext(id int) string {
	benchContextMutex.Lock()
	defer benchContextMutex.Unlock()
	
	ctxID := fmt.Sprintf("ctx-test-%d", id)
	benchContextStore[ctxID] = map[string]interface{}{
		"id":   id,
		"type": "benchmark",
	}
	return ctxID
}

func getContext(id string) interface{} {
	benchContextMutex.RLock()
	defer benchContextMutex.RUnlock()
	
	return benchContextStore[id]
}

func performContextOperation() {
	// Simulate a context operation with some work
	benchContextMutex.Lock()
	defer benchContextMutex.Unlock()
	
	// Create a context
	id := fmt.Sprintf("op-ctx-%d", contextCounter)
	contextCounter++
	benchContextStore[id] = map[string]interface{}{
		"operation": "concurrent",
	}
	
	// Read it back
	_ = benchContextStore[id]
	
	// Delete it
	delete(benchContextStore, id)
}
