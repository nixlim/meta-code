package router

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRequestManager(t *testing.T) {
	rm := NewRequestManager(ManagerConfig{
		MaxConcurrent: 5,
		MaxQueued:     10,
	})

	err := rm.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer rm.Shutdown(context.Background())

	t.Run("ExecuteSimple", func(t *testing.T) {
		var executed bool
		err := rm.Execute(context.Background(), "test-1", func(ctx context.Context) error {
			executed = true
			return nil
		})

		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}

		// Wait briefly for execution
		time.Sleep(10 * time.Millisecond)

		if !executed {
			t.Error("Function was not executed")
		}
	})

	t.Run("ExecuteWithError", func(t *testing.T) {
		expectedErr := errors.New("test error")
		err := rm.Execute(context.Background(), "test-2", func(ctx context.Context) error {
			return expectedErr
		})

		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}

		// Error from function doesn't bubble up, but metrics should reflect it
		time.Sleep(10 * time.Millisecond)
		metrics := rm.GetMetrics()

		if metrics.CompletedRequests < 1 {
			t.Error("Expected at least 1 completed request")
		}
	})

	t.Run("ConcurrentExecution", func(t *testing.T) {
		var counter int32
		numRequests := 20

		var wg sync.WaitGroup
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()

				err := rm.Execute(context.Background(), fmt.Sprintf("concurrent-%d", n), func(ctx context.Context) error {
					atomic.AddInt32(&counter, 1)
					time.Sleep(5 * time.Millisecond) // Simulate work
					return nil
				})

				if err != nil {
					t.Errorf("Execute failed for request %d: %v", n, err)
				}
			}(i)
		}

		wg.Wait()

		// Wait for all executions to complete
		time.Sleep(50 * time.Millisecond)

		if int(counter) != numRequests {
			t.Errorf("Expected %d executions, got %d", numRequests, counter)
		}

		metrics := rm.GetMetrics()
		if metrics.TotalRequests < int64(numRequests) {
			t.Errorf("Expected at least %d total requests, got %d", numRequests, metrics.TotalRequests)
		}
	})

	t.Run("CancelRequest", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Create request context with correlation ID
		rc := NewRequestContext("cancel-correlation")
		ctx = WithRequestContext(ctx, rc)

		var cancelled bool
		err := rm.Execute(ctx, "cancel-test", func(execCtx context.Context) error {
			select {
			case <-time.After(100 * time.Millisecond):
				return errors.New("not cancelled")
			case <-execCtx.Done():
				cancelled = true
				return execCtx.Err()
			}
		})

		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}

		// Give it time to start
		time.Sleep(10 * time.Millisecond)

		// Cancel the request
		err = rm.CancelRequest("cancel-test")
		if err != nil {
			t.Errorf("CancelRequest failed: %v", err)
		}

		// Wait for cancellation to take effect
		time.Sleep(20 * time.Millisecond)

		if !cancelled {
			t.Error("Request was not cancelled")
		}
	})

	t.Run("GetActiveRequest", func(t *testing.T) {
		ctx := context.Background()
		rc := NewRequestContext("active-correlation")
		rc.SetMetadata("method", "test.method")
		ctx = WithRequestContext(ctx, rc)

		err := rm.Execute(ctx, "active-test", func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		})

		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}

		// Give it time to start
		time.Sleep(10 * time.Millisecond)

		// Get active request
		activeReq, ok := rm.GetActiveRequest("active-test")
		if !ok {
			t.Fatal("Active request not found")
		}

		if activeReq.ID != "active-test" {
			t.Errorf("Expected ID active-test, got %s", activeReq.ID)
		}

		if activeReq.CorrelationID != "active-correlation" {
			t.Errorf("Expected correlation ID active-correlation, got %s", activeReq.CorrelationID)
		}

		if activeReq.Method != "test.method" {
			t.Errorf("Expected method test.method, got %s", activeReq.Method)
		}
	})

	t.Run("ListActiveRequests", func(t *testing.T) {
		// Start some long-running requests
		for i := 0; i < 3; i++ {
			err := rm.Execute(context.Background(), fmt.Sprintf("list-%d", i), func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				return nil
			})

			if err != nil {
				t.Errorf("Execute failed: %v", err)
			}
		}

		// Give them time to start
		time.Sleep(10 * time.Millisecond)

		activeRequests := rm.ListActiveRequests()
		if len(activeRequests) < 3 {
			t.Errorf("Expected at least 3 active requests, got %d", len(activeRequests))
		}
	})

	t.Run("Metrics", func(t *testing.T) {
		initialMetrics := rm.GetMetrics()

		// Execute some requests
		for i := 0; i < 5; i++ {
			rm.Execute(context.Background(), fmt.Sprintf("metrics-%d", i), func(ctx context.Context) error {
				return nil
			})
		}

		// Wait for completion
		time.Sleep(20 * time.Millisecond)

		finalMetrics := rm.GetMetrics()

		if finalMetrics.TotalRequests <= initialMetrics.TotalRequests {
			t.Error("Expected TotalRequests to increase")
		}

		if finalMetrics.CompletedRequests <= initialMetrics.CompletedRequests {
			t.Error("Expected CompletedRequests to increase")
		}
	})
}

func TestRequestManagerQueueing(t *testing.T) {
	// Small limits for testing
	rm := NewRequestManager(ManagerConfig{
		MaxConcurrent: 2,
		MaxQueued:     3,
	})

	err := rm.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer rm.Shutdown(context.Background())

	// Block all workers
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		err := rm.Execute(context.Background(), fmt.Sprintf("blocker-%d", i), func(ctx context.Context) error {
			wg.Done()
			time.Sleep(100 * time.Millisecond) // Block
			return nil
		})

		if err != nil {
			t.Fatalf("Failed to execute blocker: %v", err)
		}
	}

	// Wait for blockers to start
	wg.Wait()

	// Queue some requests
	for i := 0; i < 3; i++ {
		err := rm.Execute(context.Background(), fmt.Sprintf("queued-%d", i), func(ctx context.Context) error {
			return nil
		})

		if err != nil {
			t.Errorf("Failed to queue request %d: %v", i, err)
		}
	}

	// This should fail (queue full)
	err = rm.Execute(context.Background(), "overflow", func(ctx context.Context) error {
		return nil
	})

	if err == nil {
		t.Error("Expected error when queue is full")
	}

	metrics := rm.GetMetrics()
	if metrics.RejectedRequests < 1 {
		t.Error("Expected at least 1 rejected request")
	}
}

func TestRequestManagerShutdown(t *testing.T) {
	rm := NewRequestManager(ManagerConfig{
		MaxConcurrent: 5,
		MaxQueued:     10,
	})

	err := rm.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Start some long-running requests
	var started int32
	var cancelled int32

	for i := 0; i < 5; i++ {
		err := rm.Execute(context.Background(), fmt.Sprintf("shutdown-%d", i), func(ctx context.Context) error {
			atomic.AddInt32(&started, 1)

			select {
			case <-time.After(100 * time.Millisecond):
				return nil
			case <-ctx.Done():
				atomic.AddInt32(&cancelled, 1)
				return ctx.Err()
			}
		})

		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}
	}

	// Give them time to start
	time.Sleep(10 * time.Millisecond)

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = rm.Shutdown(shutdownCtx)
	if err != nil {
		t.Logf("Shutdown completed with: %v", err)
	}

	// Verify requests were cancelled
	if atomic.LoadInt32(&cancelled) == 0 {
		t.Error("Expected some requests to be cancelled")
	}

	// Try to execute after shutdown
	err = rm.Execute(context.Background(), "after-shutdown", func(ctx context.Context) error {
		return nil
	})

	if err == nil {
		t.Error("Expected error when executing after shutdown")
	}
}

func TestRequestManagerTimeout(t *testing.T) {
	rm := NewRequestManager(ManagerConfig{
		MaxConcurrent: 5,
		MaxQueued:     10,
	})

	err := rm.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer rm.Shutdown(context.Background())

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	rc := NewRequestContext("timeout-test")
	ctx = WithRequestContext(ctx, rc)

	var timedOut bool
	err = rm.Execute(ctx, "timeout-request", func(execCtx context.Context) error {
		select {
		case <-time.After(50 * time.Millisecond):
			return errors.New("did not timeout")
		case <-execCtx.Done():
			timedOut = true
			return execCtx.Err()
		}
	})

	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	// Wait for timeout
	time.Sleep(30 * time.Millisecond)

	if !timedOut {
		t.Error("Request did not timeout")
	}

	metrics := rm.GetMetrics()
	if metrics.TimeoutRequests < 1 {
		t.Error("Expected at least 1 timeout request in metrics")
	}
}
