package router

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewRequestContext(t *testing.T) {
	correlationID := "test-correlation-123"
	rc := NewRequestContext(correlationID)

	if rc.CorrelationID != correlationID {
		t.Errorf("Expected CorrelationID %s, got %s", correlationID, rc.CorrelationID)
	}

	if rc.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}

	if rc.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}
}

func TestWithAndGetRequestContext(t *testing.T) {
	ctx := context.Background()
	rc := NewRequestContext("test-123")

	// Test WithRequestContext
	ctxWithRC := WithRequestContext(ctx, rc)

	// Test GetRequestContext
	retrieved, ok := GetRequestContext(ctxWithRC)
	if !ok {
		t.Error("Expected to retrieve RequestContext")
	}

	if retrieved != rc {
		t.Error("Expected to retrieve the same RequestContext instance")
	}

	// Test with context that doesn't have RequestContext
	_, ok = GetRequestContext(ctx)
	if ok {
		t.Error("Expected GetRequestContext to return false for context without RequestContext")
	}
}

func TestRequestContextMetadata(t *testing.T) {
	rc := NewRequestContext("test-123")

	// Test SetMetadata and GetMetadata
	rc.SetMetadata("key1", "value1")
	rc.SetMetadata("key2", 42)

	val1, ok := rc.GetMetadata("key1")
	if !ok {
		t.Error("Expected to retrieve key1")
	}
	if val1 != "value1" {
		t.Errorf("Expected value1, got %v", val1)
	}

	val2, ok := rc.GetMetadata("key2")
	if !ok {
		t.Error("Expected to retrieve key2")
	}
	if val2 != 42 {
		t.Errorf("Expected 42, got %v", val2)
	}

	// Test non-existent key
	_, ok = rc.GetMetadata("nonexistent")
	if ok {
		t.Error("Expected GetMetadata to return false for non-existent key")
	}
}

func TestRequestContextGetMetadataString(t *testing.T) {
	rc := NewRequestContext("test-123")

	rc.SetMetadata("string", "value")
	rc.SetMetadata("number", 42)

	// Test string value
	val, ok := rc.GetMetadataString("string")
	if !ok {
		t.Error("Expected to retrieve string value")
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %s", val)
	}

	// Test non-string value
	_, ok = rc.GetMetadataString("number")
	if ok {
		t.Error("Expected GetMetadataString to return false for non-string value")
	}

	// Test non-existent key
	_, ok = rc.GetMetadataString("nonexistent")
	if ok {
		t.Error("Expected GetMetadataString to return false for non-existent key")
	}
}

func TestRequestContextConcurrentAccess(t *testing.T) {
	rc := NewRequestContext("test-123")

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", n)
			rc.SetMetadata(key, n)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", n)
			rc.GetMetadata(key)
		}(i)
	}

	wg.Wait()

	// Verify all values were set correctly
	for i := 0; i < iterations; i++ {
		key := fmt.Sprintf("key%d", i)
		val, ok := rc.GetMetadata(key)
		if !ok {
			t.Errorf("Expected to find %s", key)
		}
		if val != i {
			t.Errorf("Expected %d, got %v", i, val)
		}
	}
}

func TestRequestContextDuration(t *testing.T) {
	rc := NewRequestContext("test-123")

	// Sleep briefly
	time.Sleep(50 * time.Millisecond)

	duration := rc.Duration()
	if duration < 50*time.Millisecond {
		t.Errorf("Expected duration to be at least 50ms, got %v", duration)
	}
}

func TestRequestContextTimeout(t *testing.T) {
	// Test with no timeout
	rc1 := NewRequestContext("test-1")
	if rc1.IsTimedOut() {
		t.Error("Expected IsTimedOut to return false when no timeout is set")
	}

	// Test with timeout not exceeded
	rc2 := NewRequestContext("test-2")
	rc2.Timeout = 100 * time.Millisecond
	if rc2.IsTimedOut() {
		t.Error("Expected IsTimedOut to return false when timeout not exceeded")
	}

	// Test with timeout exceeded
	rc3 := NewRequestContext("test-3")
	rc3.Timeout = 10 * time.Millisecond
	time.Sleep(20 * time.Millisecond)
	if !rc3.IsTimedOut() {
		t.Error("Expected IsTimedOut to return true when timeout exceeded")
	}
}

func TestRequestContextWithTimeout(t *testing.T) {
	// Test with no timeout
	rc1 := NewRequestContext("test-1")
	ctx1, cancel1 := rc1.WithTimeout(context.Background())
	defer cancel1()

	select {
	case <-ctx1.Done():
		t.Error("Expected context without timeout to not be cancelled")
	case <-time.After(10 * time.Millisecond):
		// Expected
	}

	// Test with timeout
	rc2 := NewRequestContext("test-2")
	rc2.Timeout = 50 * time.Millisecond
	ctx2, cancel2 := rc2.WithTimeout(context.Background())
	defer cancel2()

	select {
	case <-ctx2.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected context with timeout to be cancelled")
	}
}

func TestRequestContextClone(t *testing.T) {
	rc := NewRequestContext("test-123")
	rc.SetMetadata("key1", "value1")
	rc.SetMetadata("key2", 42)
	rc.Timeout = 100 * time.Millisecond

	clone := rc.Clone()

	// Verify fields are copied
	if clone.CorrelationID != rc.CorrelationID {
		t.Error("Expected CorrelationID to be copied")
	}

	if clone.StartTime != rc.StartTime {
		t.Error("Expected StartTime to be copied")
	}

	if clone.Timeout != rc.Timeout {
		t.Error("Expected Timeout to be copied")
	}

	// Verify metadata is copied
	val1, _ := clone.GetMetadata("key1")
	if val1 != "value1" {
		t.Error("Expected metadata to be copied")
	}

	// Verify it's a deep copy (modifying clone doesn't affect original)
	clone.SetMetadata("key3", "value3")
	_, ok := rc.GetMetadata("key3")
	if ok {
		t.Error("Expected clone modifications to not affect original")
	}
}
