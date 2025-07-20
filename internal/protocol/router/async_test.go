package router

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

func TestAsyncRouter(t *testing.T) {
	// Create base router with test handler
	baseRouter := New()
	baseRouter.RegisterFunc("test.echo", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return &jsonrpc.Response{
			ID:     req.ID,
			Result: req.Params,
		}
	})

	baseRouter.RegisterFunc("test.slow", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		time.Sleep(50 * time.Millisecond)
		return &jsonrpc.Response{
			ID:     req.ID,
			Result: map[string]interface{}{"status": "slow"},
		}
	})

	baseRouter.RegisterFunc("test.error", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return &jsonrpc.Response{
			ID:    req.ID,
			Error: jsonrpc.NewError(jsonrpc.ErrorCodeInternal, "test error", nil),
		}
	})

	// Create async router
	ar := NewAsyncRouter(AsyncRouterConfig{
		Router:    baseRouter,
		Workers:   5,
		QueueSize: 100, // Increased for concurrent tests
	})

	// Start router
	err := ar.Start()
	if err != nil {
		t.Fatalf("Failed to start router: %v", err)
	}
	defer ar.Shutdown(context.Background())

	t.Run("HandleAsync", func(t *testing.T) {
		req := &jsonrpc.Request{
			ID:     "test-1",
			Method: "test.echo",
			Params: map[string]interface{}{"message": "hello"},
		}

		correlationID, err := ar.HandleAsync(context.Background(), req)
		if err != nil {
			t.Fatalf("HandleAsync failed: %v", err)
		}

		if correlationID == "" {
			t.Error("Expected non-empty correlation ID")
		}

		// Get response
		resp, err := ar.GetResponse(correlationID, 1*time.Second)
		if err != nil {
			t.Fatalf("GetResponse failed: %v", err)
		}

		if resp == nil {
			t.Fatal("GetResponse returned nil response")
		}

		if resp.ID != req.ID {
			t.Errorf("Expected response ID %v, got %v", req.ID, resp.ID)
		}

		if resp.Error != nil {
			t.Errorf("Unexpected error: %v", resp.Error)
		}
	})

	t.Run("HandleAsyncWithTimeout", func(t *testing.T) {
		req := &jsonrpc.Request{
			ID:     "test-2",
			Method: "test.slow",
			Params: nil,
		}

		// Short timeout
		correlationID, err := ar.HandleAsyncWithTimeout(context.Background(), req, 10*time.Millisecond)
		if err != nil {
			t.Fatalf("HandleAsyncWithTimeout failed: %v", err)
		}

		// Should timeout
		_, err = ar.GetResponse(correlationID, 30*time.Millisecond)
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})

	t.Run("HandleAsyncWithCallback", func(t *testing.T) {
		req := &jsonrpc.Request{
			ID:     "test-3",
			Method: "test.echo",
			Params: map[string]interface{}{"test": true},
		}

		var wg sync.WaitGroup
		wg.Add(1)

		var gotResponse *jsonrpc.Response
		var gotError error

		err := ar.HandleAsyncWithCallback(context.Background(), req, func(resp *jsonrpc.Response, err error) {
			defer wg.Done()
			gotResponse = resp
			gotError = err
		})

		if err != nil {
			t.Fatalf("HandleAsyncWithCallback failed: %v", err)
		}

		// Wait for callback
		wg.Wait()

		if gotError != nil {
			t.Errorf("Callback received error: %v", gotError)
		}

		if gotResponse == nil {
			t.Fatal("Callback did not receive response")
		}

		if gotResponse.ID != req.ID {
			t.Errorf("Expected response ID %v, got %v", req.ID, gotResponse.ID)
		}
	})

	t.Run("ConcurrentRequests", func(t *testing.T) {
		numRequests := 50
		var wg sync.WaitGroup
		var successCount int32

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()

				req := &jsonrpc.Request{
					ID:     fmt.Sprintf("concurrent-%d", n),
					Method: "test.echo",
					Params: map[string]interface{}{"num": n},
				}

				correlationID, err := ar.HandleAsync(context.Background(), req)
				if err != nil {
					t.Errorf("Request %d failed: %v", n, err)
					return
				}

				resp, err := ar.GetResponse(correlationID, 1*time.Second)
				if err != nil {
					t.Errorf("GetResponse %d failed: %v", n, err)
					return
				}

				if resp == nil {
					t.Errorf("GetResponse %d returned nil response", n)
					return
				}

				if resp.Error == nil {
					atomic.AddInt32(&successCount, 1)
				}
			}(i)
		}

		wg.Wait()

		if int(successCount) != numRequests {
			t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
		}
	})

	t.Run("SynchronousHandle", func(t *testing.T) {
		req := &jsonrpc.Request{
			ID:     "sync-1",
			Method: "test.echo",
			Params: map[string]interface{}{"sync": true},
		}

		resp := ar.Handle(context.Background(), req)

		if resp.Error != nil {
			t.Errorf("Unexpected error: %v", resp.Error)
		}

		if resp.ID != req.ID {
			t.Errorf("Expected response ID %v, got %v", req.ID, resp.ID)
		}
	})

	t.Run("Stats", func(t *testing.T) {
		stats := ar.Stats()

		if !stats.Running {
			t.Error("Expected router to be running")
		}

		if stats.Workers != 5 {
			t.Errorf("Expected 5 workers, got %d", stats.Workers)
		}
	})
}

func TestAsyncRouterShutdown(t *testing.T) {
	baseRouter := New()
	baseRouter.RegisterFunc("test.sleep", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		select {
		case <-time.After(100 * time.Millisecond):
			return &jsonrpc.Response{ID: req.ID, Result: "completed"}
		case <-ctx.Done():
			return &jsonrpc.Response{
				ID:    req.ID,
				Error: jsonrpc.NewError(jsonrpc.ErrorCodeInternal, "cancelled", nil),
			}
		}
	})

	ar := NewAsyncRouter(AsyncRouterConfig{
		Router:    baseRouter,
		Workers:   3,
		QueueSize: 10,
	})

	err := ar.Start()
	if err != nil {
		t.Fatalf("Failed to start router: %v", err)
	}

	// Queue some requests
	var correlationIDs []string
	for i := 0; i < 5; i++ {
		req := &jsonrpc.Request{
			ID:     fmt.Sprintf("shutdown-%d", i),
			Method: "test.sleep",
		}

		correlationID, err := ar.HandleAsync(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to handle request: %v", err)
		}
		correlationIDs = append(correlationIDs, correlationID)
	}

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = ar.Shutdown(shutdownCtx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify router is stopped
	stats := ar.Stats()
	if stats.Running {
		t.Error("Expected router to be stopped")
	}

	// Try to handle new request (should fail)
	_, err = ar.HandleAsync(context.Background(), &jsonrpc.Request{ID: "after-shutdown"})
	if err != ErrRouterShutdown {
		t.Errorf("Expected ErrRouterShutdown, got %v", err)
	}
}

func TestAsyncRouterQueueFull(t *testing.T) {
	baseRouter := New()
	baseRouter.RegisterFunc("test.block", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		// Block until context is cancelled
		<-ctx.Done()
		return &jsonrpc.Response{ID: req.ID}
	})

	// Small queue for testing
	ar := NewAsyncRouter(AsyncRouterConfig{
		Router:    baseRouter,
		Workers:   1,
		QueueSize: 2,
	})

	err := ar.Start()
	if err != nil {
		t.Fatalf("Failed to start router: %v", err)
	}
	defer ar.Shutdown(context.Background())

	// Fill up the queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// First request blocks the worker
	_, err = ar.HandleAsync(ctx, &jsonrpc.Request{ID: "1", Method: "test.block"})
	if err != nil {
		t.Fatalf("Failed to handle first request: %v", err)
	}

	// Give the first request time to be picked up by worker
	time.Sleep(10 * time.Millisecond)

	// Fill the queue (worker is blocked, so these go to queue)
	for i := 2; i <= 3; i++ {
		_, err = ar.HandleAsync(ctx, &jsonrpc.Request{ID: fmt.Sprintf("%d", i), Method: "test.block"})
		if err != nil {
			t.Fatalf("Failed to handle request %d: %v", i, err)
		}
	}

	// Queue should be full now (1 in worker, 2 in queue)
	_, err = ar.HandleAsync(ctx, &jsonrpc.Request{ID: "4", Method: "test.block"})
	if err != ErrQueueFull {
		t.Errorf("Expected ErrQueueFull, got %v", err)
	}
}

func TestAsyncRouterWithMiddleware(t *testing.T) {
	baseRouter := New()
	baseRouter.RegisterFunc("test.method", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return &jsonrpc.Response{
			ID:     req.ID,
			Result: map[string]interface{}{"original": true},
		}
	})

	var middlewareCalled bool
	testMiddleware := func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			middlewareCalled = true

			// Modify response
			resp := next.Handle(ctx, req)
			if result, ok := resp.Result.(map[string]interface{}); ok {
				result["middleware"] = true
			}

			return resp
		})
	}

	ar := NewAsyncRouter(AsyncRouterConfig{
		Router:     baseRouter,
		Workers:    2,
		QueueSize:  10,
		Middleware: []Middleware{testMiddleware},
	})

	err := ar.Start()
	if err != nil {
		t.Fatalf("Failed to start router: %v", err)
	}
	defer ar.Shutdown(context.Background())

	req := &jsonrpc.Request{
		ID:     "middleware-test",
		Method: "test.method",
	}

	correlationID, err := ar.HandleAsync(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleAsync failed: %v", err)
	}

	resp, err := ar.GetResponse(correlationID, 1*time.Second)
	if err != nil {
		t.Fatalf("GetResponse failed: %v", err)
	}

	if resp == nil {
		t.Fatal("GetResponse returned nil response")
	}

	if !middlewareCalled {
		t.Error("Middleware was not called")
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}

	if !result["original"].(bool) || !result["middleware"].(bool) {
		t.Error("Expected both original and middleware flags to be true")
	}
}

// Benchmarks for async router performance
func BenchmarkAsyncRouterHandleAsync(b *testing.B) {
	baseRouter := New()
	baseRouter.RegisterFunc("test.method", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return jsonrpc.NewResponse(map[string]interface{}{"success": true}, req.ID)
	})

	ar := NewAsyncRouter(AsyncRouterConfig{
		Router:    baseRouter,
		Workers:   4,
		QueueSize: 1000,
	})

	err := ar.Start()
	if err != nil {
		b.Fatalf("Failed to start router: %v", err)
	}
	defer ar.Shutdown(context.Background())

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := jsonrpc.NewRequest("test.method", map[string]interface{}{"key": "value"}, i)
		correlationID, err := ar.HandleAsync(ctx, req)
		if err != nil {
			b.Fatal("HandleAsync failed:", err)
		}

		_, err = ar.GetResponse(correlationID, 1*time.Second)
		if err != nil {
			b.Fatal("GetResponse failed:", err)
		}
	}
}

func BenchmarkAsyncRouterConcurrentRequests(b *testing.B) {
	baseRouter := New()
	baseRouter.RegisterFunc("test.method", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return jsonrpc.NewResponse(map[string]interface{}{"success": true}, req.ID)
	})

	ar := NewAsyncRouter(AsyncRouterConfig{
		Router:    baseRouter,
		Workers:   8,
		QueueSize: 10000,
	})

	err := ar.Start()
	if err != nil {
		b.Fatalf("Failed to start router: %v", err)
	}
	defer ar.Shutdown(context.Background())

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := jsonrpc.NewRequest("test.method", map[string]interface{}{"id": i}, i)
			correlationID, err := ar.HandleAsync(ctx, req)
			if err != nil {
				b.Fatal("HandleAsync failed:", err)
			}

			_, err = ar.GetResponse(correlationID, 1*time.Second)
			if err != nil {
				b.Fatal("GetResponse failed:", err)
			}
			i++
		}
	})
}

func BenchmarkAsyncRouterSynchronousHandle(b *testing.B) {
	baseRouter := New()
	baseRouter.RegisterFunc("test.method", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return jsonrpc.NewResponse(map[string]interface{}{"success": true}, req.ID)
	})

	ar := NewAsyncRouter(AsyncRouterConfig{
		Router:    baseRouter,
		Workers:   4,
		QueueSize: 1000,
	})

	err := ar.Start()
	if err != nil {
		b.Fatalf("Failed to start router: %v", err)
	}
	defer ar.Shutdown(context.Background())

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := jsonrpc.NewRequest("test.method", map[string]interface{}{"key": "value"}, i)
		response := ar.Handle(ctx, req)
		if response.Error != nil {
			b.Fatal("Handle failed:", response.Error)
		}
	}
}
