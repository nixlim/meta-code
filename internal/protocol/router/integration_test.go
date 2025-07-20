package router

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// TestAsyncRouterIntegration tests the complete async router with all features
func TestAsyncRouterIntegration(t *testing.T) {
	// Create base router with various handlers
	baseRouter := New()

	// Echo handler - returns params
	baseRouter.RegisterFunc("echo", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return &jsonrpc.Response{
			ID:     req.ID,
			Result: req.Params,
		}
	})

	// Slow handler - simulates processing
	baseRouter.RegisterFunc("slow", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		select {
		case <-time.After(100 * time.Millisecond):
			return &jsonrpc.Response{
				ID:     req.ID,
				Result: map[string]interface{}{"processed": true},
			}
		case <-ctx.Done():
			return &jsonrpc.Response{
				ID:    req.ID,
				Error: jsonrpc.NewError(jsonrpc.ErrorCodeTimeout, "cancelled", nil),
			}
		}
	})

	// Context-aware handler
	baseRouter.RegisterFunc("context", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		rc, ok := GetRequestContext(ctx)
		if !ok {
			return &jsonrpc.Response{
				ID:    req.ID,
				Error: jsonrpc.NewError(jsonrpc.ErrorCodeInternal, "no request context", nil),
			}
		}

		return &jsonrpc.Response{
			ID: req.ID,
			Result: map[string]interface{}{
				"correlationID": rc.CorrelationID,
				"metadata":      rc.Metadata,
			},
		}
	})

	// Setup logging
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", log.LstdFlags)

	// Setup metrics
	metrics := NewRequestMetrics()

	// Create async router with middleware
	ar := NewAsyncRouter(AsyncRouterConfig{
		Router:    baseRouter,
		Workers:   10,
		QueueSize: 50,
		Middleware: []Middleware{
			ContextEnrichmentMiddleware(),
			LoggingMiddleware(logger),
			MetricsMiddleware(metrics),
			RecoveryMiddleware(logger),
			TimeoutMiddleware(200 * time.Millisecond),
		},
	})

	// Start router
	err := ar.Start()
	if err != nil {
		t.Fatalf("Failed to start router: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		ar.Shutdown(ctx)
	}()

	t.Run("BasicAsyncFlow", func(t *testing.T) {
		req := &jsonrpc.Request{
			ID:     "basic-1",
			Method: "echo",
			Params: map[string]interface{}{"message": "hello async"},
		}

		correlationID, err := ar.HandleAsync(context.Background(), req)
		if err != nil {
			t.Fatalf("HandleAsync failed: %v", err)
		}

		resp, err := ar.GetResponse(correlationID, 1*time.Second)
		if err != nil {
			t.Fatalf("GetResponse failed: %v", err)
		}

		if resp.Error != nil {
			t.Errorf("Unexpected error: %v", resp.Error)
		}

		// Check logs contain correlation ID
		if !bytes.Contains(logBuf.Bytes(), []byte(correlationID)) {
			t.Error("Expected correlation ID in logs")
		}

		// Check metrics
		if metrics.TotalRequests < 1 {
			t.Error("Expected metrics to be updated")
		}
	})

	t.Run("ContextPropagation", func(t *testing.T) {
		ctx := context.Background()
		rc := NewRequestContext("test-context-123")
		rc.SetMetadata("user", "testuser")
		rc.SetMetadata("requestType", "test")
		ctx = WithRequestContext(ctx, rc)

		req := &jsonrpc.Request{
			ID:     "context-1",
			Method: "context",
		}

		correlationID, err := ar.HandleAsync(ctx, req)
		if err != nil {
			t.Fatalf("HandleAsync failed: %v", err)
		}

		resp, err := ar.GetResponse(correlationID, 1*time.Second)
		if err != nil {
			t.Fatalf("GetResponse failed: %v", err)
		}

		result := resp.Result.(map[string]interface{})
		metadata := result["metadata"].(map[string]interface{})

		if metadata["user"] != "testuser" {
			t.Error("Expected user metadata to be propagated")
		}

		// Should have method from enrichment middleware
		if metadata["method"] != "context" {
			t.Error("Expected method to be added by middleware")
		}
	})

	t.Run("TimeoutHandling", func(t *testing.T) {
		// Request with short timeout
		ctx := context.Background()
		rc := NewRequestContext("timeout-test")
		rc.Timeout = 50 * time.Millisecond
		ctx = WithRequestContext(ctx, rc)

		req := &jsonrpc.Request{
			ID:     "timeout-1",
			Method: "slow", // Takes 100ms
		}

		correlationID, err := ar.HandleAsync(ctx, req)
		if err != nil {
			t.Fatalf("HandleAsync failed: %v", err)
		}

		resp, err := ar.GetResponse(correlationID, 200*time.Millisecond)
		if err != nil {
			t.Fatalf("GetResponse failed: %v", err)
		}

		if resp.Error == nil {
			t.Error("Expected timeout error")
		}

		if resp.Error.Code != jsonrpc.ErrorCodeTimeout {
			t.Errorf("Expected timeout error code, got %d", resp.Error.Code)
		}
	})

	t.Run("ConcurrentRequests", func(t *testing.T) {
		numRequests := 100
		var wg sync.WaitGroup
		var successCount int32
		var errorCount int32

		start := time.Now()

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()

				req := &jsonrpc.Request{
					ID:     fmt.Sprintf("concurrent-%d", n),
					Method: "echo",
					Params: map[string]interface{}{"num": n},
				}

				correlationID, err := ar.HandleAsync(context.Background(), req)
				if err != nil {
					atomic.AddInt32(&errorCount, 1)
					return
				}

				resp, err := ar.GetResponse(correlationID, 2*time.Second)
				if err != nil {
					atomic.AddInt32(&errorCount, 1)
					return
				}

				if resp.Error == nil {
					atomic.AddInt32(&successCount, 1)
				} else {
					atomic.AddInt32(&errorCount, 1)
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		t.Logf("Processed %d requests in %v", numRequests, duration)
		t.Logf("Success: %d, Errors: %d", successCount, errorCount)

		if int(successCount) < numRequests*90/100 {
			t.Errorf("Expected at least 90%% success rate, got %d%%", int(successCount)*100/numRequests)
		}

		// Check final metrics
		finalMetrics := metrics
		t.Logf("Total requests: %d, Total errors: %d",
			finalMetrics.TotalRequests, finalMetrics.TotalErrors)

		if finalMetrics.TotalRequests < int64(numRequests) {
			t.Errorf("Expected at least %d total requests in metrics", numRequests)
		}
	})

	t.Run("CallbackPattern", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		var gotResult map[string]interface{}
		var gotError error

		req := &jsonrpc.Request{
			ID:     "callback-1",
			Method: "echo",
			Params: map[string]interface{}{"callback": true},
		}

		err := ar.HandleAsyncWithCallback(context.Background(), req, func(resp *jsonrpc.Response, err error) {
			defer wg.Done()

			if err != nil {
				gotError = err
				return
			}

			if resp.Error != nil {
				gotError = fmt.Errorf("response error: %v", resp.Error)
				return
			}

			gotResult = resp.Result.(map[string]interface{})
		})

		if err != nil {
			t.Fatalf("HandleAsyncWithCallback failed: %v", err)
		}

		wg.Wait()

		if gotError != nil {
			t.Errorf("Callback received error: %v", gotError)
		}

		if gotResult["callback"] != true {
			t.Error("Expected callback result")
		}
	})

	t.Run("RouterStats", func(t *testing.T) {
		stats := ar.Stats()

		if !stats.Running {
			t.Error("Expected router to be running")
		}

		if stats.Workers != 10 {
			t.Errorf("Expected 10 workers, got %d", stats.Workers)
		}

		t.Logf("Router stats: Queued=%d, Pending=%d",
			stats.QueuedRequests, stats.PendingRequests)
	})
}

// TestAsyncRouterStressTest performs a stress test
func TestAsyncRouterStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	baseRouter := New()

	// Variable delay handler
	baseRouter.RegisterFunc("work", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		delay := time.Duration(req.Params.(map[string]interface{})["delay"].(float64)) * time.Millisecond

		select {
		case <-time.After(delay):
			return &jsonrpc.Response{
				ID:     req.ID,
				Result: map[string]interface{}{"completed": true},
			}
		case <-ctx.Done():
			return &jsonrpc.Response{
				ID:    req.ID,
				Error: jsonrpc.NewError(jsonrpc.ErrorCodeTimeout, "cancelled", nil),
			}
		}
	})

	ar := NewAsyncRouter(AsyncRouterConfig{
		Router:    baseRouter,
		Workers:   20,
		QueueSize: 200,
	})

	err := ar.Start()
	if err != nil {
		t.Fatalf("Failed to start router: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ar.Shutdown(ctx)
	}()

	// Run stress test
	numRequests := 1000
	var wg sync.WaitGroup
	var successCount int32
	var timeoutCount int32
	var errorCount int32

	start := time.Now()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			// Random delay between 1-100ms
			delay := float64(1 + n%100)

			req := &jsonrpc.Request{
				ID:     fmt.Sprintf("stress-%d", n),
				Method: "work",
				Params: map[string]interface{}{"delay": delay},
			}

			// Random timeout between 50-150ms
			timeout := time.Duration(50+n%100) * time.Millisecond

			correlationID, err := ar.HandleAsyncWithTimeout(context.Background(), req, timeout)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				return
			}

			resp, err := ar.GetResponse(correlationID, timeout+50*time.Millisecond)
			if err != nil {
				if err == ErrCorrelationTimeout {
					atomic.AddInt32(&timeoutCount, 1)
				} else {
					atomic.AddInt32(&errorCount, 1)
				}
				return
			}

			if resp.Error != nil {
				if resp.Error.Code == jsonrpc.ErrorCodeTimeout {
					atomic.AddInt32(&timeoutCount, 1)
				} else {
					atomic.AddInt32(&errorCount, 1)
				}
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)

		// Stagger request starts
		if i%50 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	wg.Wait()
	duration := time.Since(start)

	t.Logf("Stress test completed in %v", duration)
	t.Logf("Requests: %d, Success: %d, Timeouts: %d, Errors: %d",
		numRequests, successCount, timeoutCount, errorCount)

	total := successCount + timeoutCount + errorCount
	if int(total) != numRequests {
		t.Errorf("Expected %d total outcomes, got %d", numRequests, total)
	}

	// At least 50% should succeed (depends on random delays vs timeouts)
	if int(successCount) < numRequests/2 {
		t.Errorf("Expected at least 50%% success rate, got %d%%", int(successCount)*100/numRequests)
	}
}
