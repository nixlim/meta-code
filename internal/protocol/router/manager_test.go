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
		numRequests := 15 // Max concurrent (5) + Max queued (10)

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

	t.Run("RejectedRequests", func(t *testing.T) {
		// Try to execute more than capacity
		var rejected int32
		for i := 0; i < 25; i++ {
			err := rm.Execute(context.Background(), fmt.Sprintf("overflow-%d", i), func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			})
			if err != nil {
				atomic.AddInt32(&rejected, 1)
			}
		}

		if rejected == 0 {
			t.Error("Expected some requests to be rejected")
		}

		metrics := rm.GetMetrics()
		if metrics.RejectedRequests == 0 {
			t.Error("Expected RejectedRequests metric to be > 0")
		}

		// Wait for queue to clear before next test
		for i := 0; i < 50; i++ {
			m := rm.GetMetrics()
			if m.ActiveRequests == 0 && m.QueuedRequests == 0 {
				break
			}
			time.Sleep(10 * time.Millisecond)
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
			time.Sleep(500 * time.Millisecond) // Block longer to prevent queue from draining
			return nil
		})

		if err != nil {
			t.Fatalf("Failed to execute blocker: %v", err)
		}
	}

	// Wait for blockers to start
	wg.Wait()

	// Give a small delay to ensure blockers are holding semaphores
	time.Sleep(10 * time.Millisecond)

	// Check metrics before queueing
	metrics := rm.GetMetrics()
	t.Logf("Before queueing - Active: %d, Queued: %d", metrics.ActiveRequests, metrics.QueuedRequests)

	// Queue requests to fill the queue (capacity 3)
	var queueErr error
	for i := 0; i < 5; i++ { // Try 5 to ensure we hit the limit
		err := rm.Execute(context.Background(), fmt.Sprintf("queued-%d", i), func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond) // Keep them in queue
			return nil
		})

		if err != nil {
			queueErr = err
			t.Logf("Request %d rejected as expected: %v", i, err)
			break
		}
	}

	// Check metrics after queueing
	metrics = rm.GetMetrics()
	t.Logf("After queueing - Active: %d, Queued: %d", metrics.ActiveRequests, metrics.QueuedRequests)

	if queueErr == nil {
		t.Error("Expected error when queue is full")
	}

	metrics = rm.GetMetrics()
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
	var completed int32
	var startedWG sync.WaitGroup
	blockChan := make(chan struct{})
	defer close(blockChan) // Ensure cleanup on test completion

	// Use a WaitGroup to track when all request goroutines complete
	var requestWG sync.WaitGroup
	requestWG.Add(5)

	startedWG.Add(5) // Wait for all 5 requests to start

	for i := 0; i < 5; i++ {
		err := rm.Execute(context.Background(), fmt.Sprintf("shutdown-%d", i), func(ctx context.Context) error {
			defer requestWG.Done()
			atomic.AddInt32(&started, 1)
			startedWG.Done() // Signal that this request has started

			// Use a ticker to periodically check for cancellation
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-blockChan: // Released by test
					atomic.AddInt32(&completed, 1)
					return nil
				case <-ticker.C:
					// Check if context is cancelled
					select {
					case <-ctx.Done():
						atomic.AddInt32(&cancelled, 1)
						return ctx.Err()
					default:
						// Context not cancelled, continue
					}
				case <-ctx.Done(): // Cancelled by shutdown
					atomic.AddInt32(&cancelled, 1)
					return ctx.Err()
				}
			}
		})

		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}
	}

	// Wait for all requests to actually start before shutdown
	startedWG.Wait()

	// Give a small delay to ensure goroutines are in their loops
	time.Sleep(50 * time.Millisecond)

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = rm.Shutdown(shutdownCtx)
	if err != nil {
		t.Logf("Shutdown completed with: %v", err)
	}

	// Wait for all request goroutines to complete
	done := make(chan struct{})
	go func() {
		requestWG.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines completed
	case <-time.After(500 * time.Millisecond):
		t.Error("Request goroutines did not complete in time")
	}

	// Verify requests were cancelled
	startedCount := atomic.LoadInt32(&started)
	cancelledCount := atomic.LoadInt32(&cancelled)
	completedCount := atomic.LoadInt32(&completed)

	t.Logf("Started: %d, Cancelled: %d, Completed: %d", startedCount, cancelledCount, completedCount)

	// We expect at least some requests to be cancelled
	if cancelledCount == 0 && completedCount == 0 {
		t.Errorf("Expected some requests to be cancelled or completed, but got cancelled: %d, completed: %d (started: %d)",
			cancelledCount, completedCount, startedCount)
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
