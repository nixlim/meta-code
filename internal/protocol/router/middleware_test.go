package router

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// testHandler is a simple handler for testing
type testHandler struct {
	called bool
	mu     sync.Mutex
}

func (h *testHandler) Handle(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	h.mu.Lock()
	h.called = true
	h.mu.Unlock()

	return &jsonrpc.Response{
		ID:     req.ID,
		Result: map[string]interface{}{"status": "ok"},
	}
}

func (h *testHandler) wasCalled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.called
}

func TestChain(t *testing.T) {
	// Test empty chain
	t.Run("EmptyChain", func(t *testing.T) {
		chain := NewChain()
		handler := &testHandler{}
		wrapped := chain.Then(handler)

		req := &jsonrpc.Request{
			ID:     "test-1",
			Method: "test.method",
		}

		resp := wrapped.Handle(context.Background(), req)

		if !handler.wasCalled() {
			t.Error("Expected handler to be called")
		}

		if resp.ID != req.ID {
			t.Errorf("Expected response ID %v, got %v", req.ID, resp.ID)
		}
	})

	// Test chain with middleware
	t.Run("ChainWithMiddleware", func(t *testing.T) {
		var order []string
		var mu sync.Mutex

		middleware1 := func(next Handler) Handler {
			return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
				mu.Lock()
				order = append(order, "m1-before")
				mu.Unlock()

				resp := next.Handle(ctx, req)

				mu.Lock()
				order = append(order, "m1-after")
				mu.Unlock()

				return resp
			})
		}

		middleware2 := func(next Handler) Handler {
			return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
				mu.Lock()
				order = append(order, "m2-before")
				mu.Unlock()

				resp := next.Handle(ctx, req)

				mu.Lock()
				order = append(order, "m2-after")
				mu.Unlock()

				return resp
			})
		}

		chain := NewChain(middleware1, middleware2)
		handler := HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			mu.Lock()
			order = append(order, "handler")
			mu.Unlock()

			return &jsonrpc.Response{ID: req.ID}
		})

		wrapped := chain.Then(handler)
		wrapped.Handle(context.Background(), &jsonrpc.Request{ID: "test"})

		// Verify execution order
		expected := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
		if len(order) != len(expected) {
			t.Fatalf("Expected %d calls, got %d", len(expected), len(order))
		}

		for i, v := range expected {
			if order[i] != v {
				t.Errorf("Expected order[%d] = %s, got %s", i, v, order[i])
			}
		}
	})

	// Test Append
	t.Run("Append", func(t *testing.T) {
		var called []string

		m1 := func(next Handler) Handler {
			return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
				called = append(called, "m1")
				return next.Handle(ctx, req)
			})
		}

		m2 := func(next Handler) Handler {
			return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
				called = append(called, "m2")
				return next.Handle(ctx, req)
			})
		}

		chain1 := NewChain(m1)
		chain2 := chain1.Append(m2)

		handler := HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			called = append(called, "handler")
			return &jsonrpc.Response{ID: req.ID}
		})

		wrapped := chain2.Then(handler)
		wrapped.Handle(context.Background(), &jsonrpc.Request{ID: "test"})

		// Verify both middleware were called
		if len(called) != 3 || called[0] != "m1" || called[1] != "m2" || called[2] != "handler" {
			t.Errorf("Expected [m1 m2 handler], got %v", called)
		}
	})
}

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	middleware := LoggingMiddleware(logger)
	handler := HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return &jsonrpc.Response{
			ID:     req.ID,
			Result: map[string]interface{}{"status": "ok"},
		}
	})

	// Create context with request context
	ctx := context.Background()
	rc := NewRequestContext("test-correlation-123")
	ctx = WithRequestContext(ctx, rc)

	wrapped := middleware(handler)
	req := &jsonrpc.Request{
		ID:     "test-123",
		Method: "test.method",
	}

	wrapped.Handle(ctx, req)

	logs := buf.String()

	// Verify request was logged
	if !strings.Contains(logs, "test-correlation-123") {
		t.Error("Expected correlation ID in logs")
	}

	if !strings.Contains(logs, "method=test.method") {
		t.Error("Expected method in logs")
	}

	if !strings.Contains(logs, "success=true") {
		t.Error("Expected success in logs")
	}
}

func TestMetricsMiddleware(t *testing.T) {
	metrics := NewRequestMetrics()
	middleware := MetricsMiddleware(metrics)

	handler := HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)

		if req.Method == "error.method" {
			return &jsonrpc.Response{
				ID:    req.ID,
				Error: jsonrpc.NewError(jsonrpc.ErrorCodeInternal, "test error", nil),
			}
		}

		return &jsonrpc.Response{
			ID:     req.ID,
			Result: map[string]interface{}{"status": "ok"},
		}
	})

	wrapped := middleware(handler)

	// Make some requests
	wrapped.Handle(context.Background(), &jsonrpc.Request{ID: "1", Method: "test.method1"})
	wrapped.Handle(context.Background(), &jsonrpc.Request{ID: "2", Method: "test.method1"})
	wrapped.Handle(context.Background(), &jsonrpc.Request{ID: "3", Method: "test.method2"})
	wrapped.Handle(context.Background(), &jsonrpc.Request{ID: "4", Method: "error.method"})

	// Verify metrics
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	if metrics.TotalRequests != 4 {
		t.Errorf("Expected 4 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.TotalErrors != 1 {
		t.Errorf("Expected 1 error, got %d", metrics.TotalErrors)
	}

	if metrics.MethodCounts["test.method1"] != 2 {
		t.Errorf("Expected 2 calls to test.method1, got %d", metrics.MethodCounts["test.method1"])
	}

	if metrics.TotalDuration < 40*time.Millisecond {
		t.Errorf("Expected total duration >= 40ms, got %v", metrics.TotalDuration)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	middleware := RecoveryMiddleware(logger)

	// Handler that panics
	handler := HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		panic("test panic")
	})

	wrapped := middleware(handler)
	req := &jsonrpc.Request{
		ID:     "test-123",
		Method: "panic.method",
	}

	resp := wrapped.Handle(context.Background(), req)

	// Verify error response
	if resp.Error == nil {
		t.Error("Expected error response")
	}

	if resp.Error.Code != jsonrpc.ErrorCodeInternal {
		t.Errorf("Expected internal error code, got %d", resp.Error.Code)
	}

	// Verify panic was logged
	if !strings.Contains(buf.String(), "Panic recovered: test panic") {
		t.Error("Expected panic to be logged")
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	middleware := TimeoutMiddleware(50 * time.Millisecond)

	t.Run("NoTimeout", func(t *testing.T) {
		handler := HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			return &jsonrpc.Response{
				ID:     req.ID,
				Result: map[string]interface{}{"status": "ok"},
			}
		})

		wrapped := middleware(handler)
		req := &jsonrpc.Request{ID: "test-1", Method: "fast.method"}

		resp := wrapped.Handle(context.Background(), req)

		if resp.Error != nil {
			t.Errorf("Expected no error, got %v", resp.Error)
		}
	})

	t.Run("WithTimeout", func(t *testing.T) {
		handler := HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			// Simulate slow processing
			time.Sleep(100 * time.Millisecond)

			return &jsonrpc.Response{
				ID:     req.ID,
				Result: map[string]interface{}{"status": "ok"},
			}
		})

		wrapped := middleware(handler)
		req := &jsonrpc.Request{ID: "test-2", Method: "slow.method"}

		resp := wrapped.Handle(context.Background(), req)

		if resp.Error == nil {
			t.Error("Expected timeout error")
		}

		if resp.Error.Code != jsonrpc.ErrorCodeTimeout {
			t.Errorf("Expected timeout error code, got %d", resp.Error.Code)
		}
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		handler := HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			// Simulate slow processing
			time.Sleep(30 * time.Millisecond)

			return &jsonrpc.Response{
				ID:     req.ID,
				Result: map[string]interface{}{"status": "ok"},
			}
		})

		// Create context with shorter timeout
		ctx := context.Background()
		rc := NewRequestContext("test-correlation")
		rc.Timeout = 20 * time.Millisecond
		ctx = WithRequestContext(ctx, rc)

		wrapped := middleware(handler)
		req := &jsonrpc.Request{ID: "test-3", Method: "context.timeout"}

		resp := wrapped.Handle(ctx, req)

		if resp.Error == nil {
			t.Error("Expected timeout error from context")
		}

		if resp.Error.Code != jsonrpc.ErrorCodeTimeout {
			t.Errorf("Expected timeout error code, got %d", resp.Error.Code)
		}
	})
}

func TestAuthMiddleware(t *testing.T) {
	// Simple auth function
	authFunc := func(ctx context.Context, method string) error {
		if strings.HasPrefix(method, "private.") {
			return errors.New("access denied")
		}
		return nil
	}

	middleware := AuthMiddleware(authFunc)
	handler := HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return &jsonrpc.Response{
			ID:     req.ID,
			Result: map[string]interface{}{"status": "ok"},
		}
	})

	wrapped := middleware(handler)

	t.Run("Authorized", func(t *testing.T) {
		req := &jsonrpc.Request{ID: "test-1", Method: "public.method"}
		resp := wrapped.Handle(context.Background(), req)

		if resp.Error != nil {
			t.Errorf("Expected no error, got %v", resp.Error)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req := &jsonrpc.Request{ID: "test-2", Method: "private.method"}
		resp := wrapped.Handle(context.Background(), req)

		if resp.Error == nil {
			t.Error("Expected authorization error")
		}

		if resp.Error.Code != jsonrpc.ErrorCodeUnauthorized {
			t.Errorf("Expected unauthorized error code, got %d", resp.Error.Code)
		}
	})
}

func TestContextEnrichmentMiddleware(t *testing.T) {
	middleware := ContextEnrichmentMiddleware()

	handler := HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		// Verify context was enriched
		rc, ok := GetRequestContext(ctx)
		if !ok {
			t.Error("Expected RequestContext in handler")
			return &jsonrpc.Response{ID: req.ID, Error: jsonrpc.NewError(-1, "no context", nil)}
		}

		method, ok := rc.GetMetadataString("method")
		if !ok || method != req.Method {
			t.Errorf("Expected method %s in context, got %s", req.Method, method)
		}

		return &jsonrpc.Response{
			ID:     req.ID,
			Result: map[string]interface{}{"status": "ok"},
		}
	})

	wrapped := middleware(handler)
	req := &jsonrpc.Request{
		ID:     "test-123",
		Method: "test.method",
	}

	wrapped.Handle(context.Background(), req)
}
