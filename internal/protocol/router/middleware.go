package router

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// Middleware is a function that wraps a Handler to provide additional functionality
type Middleware func(Handler) Handler

// Chain represents a chain of middleware
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
	}
}

// Append adds middleware to the chain
func (c *Chain) Append(middlewares ...Middleware) *Chain {
	newMiddlewares := make([]Middleware, 0, len(c.middlewares)+len(middlewares))
	newMiddlewares = append(newMiddlewares, c.middlewares...)
	newMiddlewares = append(newMiddlewares, middlewares...)

	return &Chain{
		middlewares: newMiddlewares,
	}
}

// Then creates a new Handler by wrapping the final handler with all middleware
// Middleware is applied in reverse order, so the first middleware in the chain
// is the outermost layer
func (c *Chain) Then(final Handler) Handler {
	if final == nil {
		panic("chain: final handler cannot be nil")
	}

	// Apply middleware in reverse order
	handler := final
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}

	return handler
}

// ThenFunc is a convenience method that wraps a HandlerFunc
func (c *Chain) ThenFunc(final HandlerFunc) Handler {
	return c.Then(final)
}

// Common middleware implementations

// LoggingMiddleware logs request and response details
func LoggingMiddleware(logger *log.Logger) Middleware {
	if logger == nil {
		logger = log.Default()
	}

	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			start := time.Now()

			// Extract correlation ID if available
			correlationID := "unknown"
			if rc, ok := GetRequestContext(ctx); ok {
				correlationID = rc.CorrelationID
			}

			// Log request
			logger.Printf("[%s] Request: method=%s id=%v", correlationID, req.Method, req.ID)

			// Call next handler
			resp := next.Handle(ctx, req)

			// Log response
			duration := time.Since(start)
			if resp.Error != nil {
				logger.Printf("[%s] Response: id=%v error=%v duration=%v",
					correlationID, resp.ID, resp.Error, duration)
			} else {
				logger.Printf("[%s] Response: id=%v success=true duration=%v",
					correlationID, resp.ID, duration)
			}

			return resp
		})
	}
}

// MetricsMiddleware collects request metrics
type RequestMetrics struct {
	TotalRequests int64
	TotalErrors   int64
	MethodCounts  map[string]int64
	TotalDuration time.Duration
	mu            sync.RWMutex
}

func NewRequestMetrics() *RequestMetrics {
	return &RequestMetrics{
		MethodCounts: make(map[string]int64),
	}
}

func MetricsMiddleware(metrics *RequestMetrics) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			start := time.Now()

			// Call next handler
			resp := next.Handle(ctx, req)

			// Update metrics
			duration := time.Since(start)
			metrics.mu.Lock()
			metrics.TotalRequests++
			metrics.MethodCounts[req.Method]++
			metrics.TotalDuration += duration
			if resp.Error != nil {
				metrics.TotalErrors++
			}
			metrics.mu.Unlock()

			// Store duration in context if available
			if rc, ok := GetRequestContext(ctx); ok {
				rc.SetMetadata("duration", duration)
			}

			return resp
		})
	}
}

// RecoveryMiddleware recovers from panics and returns an error response
func RecoveryMiddleware(logger *log.Logger) Middleware {
	if logger == nil {
		logger = log.Default()
	}

	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) (resp *jsonrpc.Response) {
			defer func() {
				if r := recover(); r != nil {
					// Extract correlation ID if available
					correlationID := "unknown"
					if rc, ok := GetRequestContext(ctx); ok {
						correlationID = rc.CorrelationID
					}

					// Log panic
					logger.Printf("[%s] Panic recovered: %v", correlationID, r)

					// Return error response
					resp = &jsonrpc.Response{
						ID: req.ID,
						Error: jsonrpc.NewError(
							jsonrpc.ErrorCodeInternal,
							"Internal server error",
							fmt.Sprintf("panic: %v", r),
						),
					}
				}
			}()

			return next.Handle(ctx, req)
		})
	}
}

// TimeoutMiddleware enforces request timeouts
func TimeoutMiddleware(defaultTimeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			// Check if RequestContext has a timeout
			timeout := defaultTimeout
			if rc, ok := GetRequestContext(ctx); ok && rc.Timeout > 0 {
				timeout = rc.Timeout
			}

			// No timeout specified
			if timeout <= 0 {
				return next.Handle(ctx, req)
			}

			// Create timeout context
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Channel for response
			type result struct {
				resp *jsonrpc.Response
			}
			resultChan := make(chan result, 1)

			// Execute handler in goroutine
			go func() {
				resp := next.Handle(timeoutCtx, req)
				resultChan <- result{resp: resp}
			}()

			// Wait for response or timeout
			select {
			case res := <-resultChan:
				return res.resp
			case <-timeoutCtx.Done():
				return &jsonrpc.Response{
					ID: req.ID,
					Error: jsonrpc.NewError(
						jsonrpc.ErrorCodeTimeout,
						"Request timeout",
						fmt.Sprintf("timeout after %v", timeout),
					),
				}
			}
		})
	}
}

// AuthMiddleware provides basic authentication checking
type AuthFunc func(ctx context.Context, method string) error

func AuthMiddleware(authFunc AuthFunc) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			// Check authorization
			if err := authFunc(ctx, req.Method); err != nil {
				return &jsonrpc.Response{
					ID: req.ID,
					Error: jsonrpc.NewError(
						jsonrpc.ErrorCodeUnauthorized,
						"Unauthorized",
						err.Error(),
					),
				}
			}

			return next.Handle(ctx, req)
		})
	}
}

// ContextEnrichmentMiddleware adds request information to the context
func ContextEnrichmentMiddleware() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			// Get or create RequestContext
			rc, ok := GetRequestContext(ctx)
			if !ok {
				// Create new context if not present
				rc = NewRequestContext(req.ID.(string))
				ctx = WithRequestContext(ctx, rc)
			}

			// Add request metadata
			rc.SetMetadata("method", req.Method)
			rc.SetMetadata("request_id", req.ID)

			return next.Handle(ctx, req)
		})
	}
}
