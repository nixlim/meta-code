package router

import (
	"context"
	"sync"
	"time"
)

// contextKey is a type used for context keys to avoid collisions
type contextKey string

const (
	// requestContextKey is the context key for storing RequestContext
	requestContextKey contextKey = "request-context"
)

// RequestContext holds request-scoped data for async processing
type RequestContext struct {
	// CorrelationID uniquely identifies a request across async boundaries
	CorrelationID string

	// Metadata stores arbitrary request-scoped data
	Metadata map[string]interface{}

	// StartTime marks when the request was received
	StartTime time.Time

	// Timeout specifies the maximum duration for request processing
	Timeout time.Duration

	// mu protects concurrent access to Metadata
	mu sync.RWMutex
}

// NewRequestContext creates a new RequestContext with the given correlation ID
func NewRequestContext(correlationID string) *RequestContext {
	return &RequestContext{
		CorrelationID: correlationID,
		Metadata:      make(map[string]interface{}),
		StartTime:     time.Now(),
	}
}

// WithRequestContext returns a new context with the RequestContext attached
func WithRequestContext(ctx context.Context, rc *RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey, rc)
}

// GetRequestContext retrieves the RequestContext from the context
func GetRequestContext(ctx context.Context) (*RequestContext, bool) {
	rc, ok := ctx.Value(requestContextKey).(*RequestContext)
	return rc, ok
}

// SetMetadata sets a metadata value in the RequestContext
func (rc *RequestContext) SetMetadata(key string, value interface{}) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.Metadata[key] = value
}

// GetMetadata retrieves a metadata value from the RequestContext
func (rc *RequestContext) GetMetadata(key string) (interface{}, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	value, ok := rc.Metadata[key]
	return value, ok
}

// GetMetadataString retrieves a metadata value as a string
func (rc *RequestContext) GetMetadataString(key string) (string, bool) {
	value, ok := rc.GetMetadata(key)
	if !ok {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// Duration returns how long the request has been processing
func (rc *RequestContext) Duration() time.Duration {
	return time.Since(rc.StartTime)
}

// IsTimedOut returns true if the request has exceeded its timeout
func (rc *RequestContext) IsTimedOut() bool {
	if rc.Timeout == 0 {
		return false
	}
	return rc.Duration() > rc.Timeout
}

// WithTimeout returns a new context with the specified timeout
func (rc *RequestContext) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if rc.Timeout == 0 {
		// No timeout specified, return a no-op cancel function
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, rc.Timeout)
}

// Clone creates a deep copy of the RequestContext
func (rc *RequestContext) Clone() *RequestContext {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	metadata := make(map[string]interface{}, len(rc.Metadata))
	for k, v := range rc.Metadata {
		metadata[k] = v
	}

	return &RequestContext{
		CorrelationID: rc.CorrelationID,
		Metadata:      metadata,
		StartTime:     rc.StartTime,
		Timeout:       rc.Timeout,
	}
}
