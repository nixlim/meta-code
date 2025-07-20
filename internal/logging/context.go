package logging

import (
	"context"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Context keys for logging
const (
	// CorrelationIDKey is the context key for correlation IDs
	CorrelationIDKey contextKey = "correlation_id"

	// RequestIDKey is the context key for request IDs
	RequestIDKey contextKey = "request_id"

	// UserIDKey is the context key for user IDs
	UserIDKey contextKey = "user_id"

	// SessionIDKey is the context key for session IDs
	SessionIDKey contextKey = "session_id"

	// ComponentKey is the context key for component names
	ComponentKey contextKey = "component"

	// MethodKey is the context key for method names
	MethodKey contextKey = "method"
)

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithSessionID adds a session ID to the context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

// WithComponent adds a component name to the context
func WithComponent(ctx context.Context, component string) context.Context {
	return context.WithValue(ctx, ComponentKey, component)
}

// WithMethod adds a method name to the context
func WithMethod(ctx context.Context, method string) context.Context {
	return context.WithValue(ctx, MethodKey, method)
}

// extractCorrelationID extracts the correlation ID from the context
func extractCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Check for correlation ID in our context
	if corrID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return corrID
	}

	// Also check for RouterContext which might have correlation ID
	// This ensures compatibility with existing router context
	if rc := extractRouterContext(ctx); rc != nil && rc.CorrelationID != "" {
		return rc.CorrelationID
	}

	return ""
}

// extractRequestID extracts the request ID from the context
func extractRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}

	return ""
}

// extractAllContextFields extracts all logging-relevant fields from the context
func extractAllContextFields(ctx context.Context) map[string]interface{} {
	if ctx == nil {
		return nil
	}

	fields := make(map[string]interface{})

	// Extract standard fields
	if corrID := extractCorrelationID(ctx); corrID != "" {
		fields["correlation_id"] = corrID
	}

	if reqID := extractRequestID(ctx); reqID != "" {
		fields["request_id"] = reqID
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		fields["user_id"] = userID
	}

	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok && sessionID != "" {
		fields["session_id"] = sessionID
	}

	if component, ok := ctx.Value(ComponentKey).(string); ok && component != "" {
		fields["component"] = component
	}

	if method, ok := ctx.Value(MethodKey).(string); ok && method != "" {
		fields["method"] = method
	}

	// Extract RouterContext fields if present
	if rc := extractRouterContext(ctx); rc != nil {
		if rc.Method != "" && fields["method"] == nil {
			fields["method"] = rc.Method
		}
		if rc.CorrelationID != "" && fields["correlation_id"] == nil {
			fields["correlation_id"] = rc.CorrelationID
		}
		// Add metadata if present
		for k, v := range rc.Metadata {
			if _, exists := fields[k]; !exists {
				fields[k] = v
			}
		}
	}

	return fields
}

// routerContextKey is the key used by the router package for RequestContext
type routerContextKey struct{}

// RouterContext represents the router's RequestContext structure
// This is a minimal representation to avoid circular dependencies
type RouterContext struct {
	CorrelationID string
	Method        string
	Metadata      map[string]interface{}
}

// extractRouterContext attempts to extract RouterContext from context
func extractRouterContext(ctx context.Context) *RouterContext {
	if ctx == nil {
		return nil
	}

	// Try to extract router context
	if val := ctx.Value(routerContextKey{}); val != nil {
		// Use type assertion to check if it matches our expected structure
		if rc, ok := val.(*RouterContext); ok {
			return rc
		}

		// If not exact match, try to extract fields using reflection
		// This is a fallback for when we can't import the actual RouterContext type
		// In production, you might want to use an interface instead
	}

	return nil
}

// ContextLogger returns a logger with all context fields pre-populated
func ContextLogger(ctx context.Context, logger *Logger) *Logger {
	fields := extractAllContextFields(ctx)
	if len(fields) == 0 {
		return logger
	}

	return logger.WithFields(fields)
}
