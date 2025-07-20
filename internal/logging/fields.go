package logging

import "fmt"

// Standard field names for consistent logging across the application
const (
	// Request/Response fields
	FieldCorrelationID = "correlation_id"
	FieldRequestID     = "request_id"
	FieldMethod        = "method"
	FieldPath          = "path"
	FieldStatusCode    = "status_code"
	FieldDuration      = "duration_ms"
	FieldResponseTime  = "response_time"

	// Error fields
	FieldError        = "error"
	FieldErrorCode    = "error_code"
	FieldErrorType    = "error_type"
	FieldErrorMessage = "error_message"
	FieldStackTrace   = "stack_trace"
	FieldCause        = "cause"

	// Context fields
	FieldUserID      = "user_id"
	FieldSessionID   = "session_id"
	FieldClientID    = "client_id"
	FieldComponent   = "component"
	FieldService     = "service"
	FieldVersion     = "version"
	FieldEnvironment = "environment"

	// MCP specific fields
	FieldProtocolVersion = "protocol_version"
	FieldServerName      = "server_name"
	FieldClientName      = "client_name"
	FieldCapabilities    = "capabilities"
	FieldHandshakeState  = "handshake_state"
	FieldConnectionID    = "connection_id"
	FieldConnectionState = "connection_state"

	// Performance fields
	FieldMemoryUsage = "memory_usage_bytes"
	FieldCPUUsage    = "cpu_usage_percent"
	FieldGoroutines  = "goroutines"
	FieldQueueSize   = "queue_size"
	FieldWorkerCount = "worker_count"

	// Metadata fields
	FieldTimestamp = "timestamp"
	FieldHostname  = "hostname"
	FieldPID       = "pid"
	FieldCaller    = "caller"
	FieldFunction  = "function"
	FieldFile      = "file"
	FieldLine      = "line"
)

// LogFields provides a fluent interface for building log fields
type LogFields map[string]interface{}

// NewLogFields creates a new LogFields instance
func NewLogFields() LogFields {
	return make(LogFields)
}

// With adds a field to the log fields
func (f LogFields) With(key string, value interface{}) LogFields {
	f[key] = value
	return f
}

// WithError adds error-related fields
func (f LogFields) WithError(err error) LogFields {
	if err == nil {
		return f
	}

	f[FieldError] = err.Error()
	f[FieldErrorType] = fmt.Sprintf("%T", err)

	// Check if error has additional properties we can extract
	type causer interface {
		Cause() error
	}
	if cause, ok := err.(causer); ok {
		if c := cause.Cause(); c != nil {
			f[FieldCause] = c.Error()
		}
	}

	return f
}

// WithRequest adds request-related fields
func (f LogFields) WithRequest(method string, path string, correlationID string) LogFields {
	if method != "" {
		f[FieldMethod] = method
	}
	if path != "" {
		f[FieldPath] = path
	}
	if correlationID != "" {
		f[FieldCorrelationID] = correlationID
	}
	return f
}

// WithResponse adds response-related fields
func (f LogFields) WithResponse(statusCode int, duration int64) LogFields {
	if statusCode > 0 {
		f[FieldStatusCode] = statusCode
	}
	if duration > 0 {
		f[FieldDuration] = duration
	}
	return f
}

// WithUser adds user-related fields
func (f LogFields) WithUser(userID string, sessionID string) LogFields {
	if userID != "" {
		f[FieldUserID] = userID
	}
	if sessionID != "" {
		f[FieldSessionID] = sessionID
	}
	return f
}

// WithComponent adds component information
func (f LogFields) WithComponent(component string) LogFields {
	if component != "" {
		f[FieldComponent] = component
	}
	return f
}

// WithConnection adds connection-related fields
func (f LogFields) WithConnection(connectionID string, state string) LogFields {
	if connectionID != "" {
		f[FieldConnectionID] = connectionID
	}
	if state != "" {
		f[FieldConnectionState] = state
	}
	return f
}

// ToMap returns the fields as a map
func (f LogFields) ToMap() map[string]interface{} {
	return map[string]interface{}(f)
}

// StandardFields returns commonly used field combinations
type StandardFields struct{}

// Fields returns a StandardFields instance for accessing field builders
func Fields() StandardFields {
	return StandardFields{}
}

// Request creates fields for a request
func (StandardFields) Request(method, correlationID string) LogFields {
	return NewLogFields().
		With(FieldMethod, method).
		With(FieldCorrelationID, correlationID)
}

// Response creates fields for a response
func (StandardFields) Response(correlationID string, duration int64, err error) LogFields {
	fields := NewLogFields().
		With(FieldCorrelationID, correlationID).
		With(FieldDuration, duration)

	if err != nil {
		fields.WithError(err)
	}

	return fields
}

// Connection creates fields for connection events
func (StandardFields) Connection(connectionID, state string) LogFields {
	return NewLogFields().
		With(FieldConnectionID, connectionID).
		With(FieldConnectionState, state)
}

// Handshake creates fields for handshake events
func (StandardFields) Handshake(connectionID, clientName, protocolVersion string) LogFields {
	return NewLogFields().
		With(FieldConnectionID, connectionID).
		With(FieldClientName, clientName).
		With(FieldProtocolVersion, protocolVersion)
}
