package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestLoggerCreation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   struct {
			hasTimestamp bool
			isJSON       bool
		}
	}{
		{
			name: "production config",
			config: Config{
				Output:    &bytes.Buffer{},
				Level:     LogLevelInfo,
				DebugMode: false,
				Sanitize:  true,
				Pretty:    false,
			},
			want: struct {
				hasTimestamp bool
				isJSON       bool
			}{
				hasTimestamp: true,
				isJSON:       true,
			},
		},
		{
			name: "development config",
			config: Config{
				Output:    &bytes.Buffer{},
				Level:     LogLevelDebug,
				DebugMode: true,
				Sanitize:  false,
				Pretty:    true,
			},
			want: struct {
				hasTimestamp bool
				isJSON       bool
			}{
				hasTimestamp: true,
				isJSON:       false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tt.config.Output = buf
			
			logger := New(tt.config)
			logger.Info(context.Background(), "test message")
			
			output := buf.String()
			
			// Check if output contains timestamp
			if tt.want.hasTimestamp && !tt.want.isJSON {
				// Console output has different timestamp format
				if !strings.Contains(output, "test message") {
					t.Error("Expected message in output")
				}
			} else if tt.want.hasTimestamp && tt.want.isJSON {
				if !strings.Contains(output, "time") {
					t.Error("Expected timestamp in JSON output")
				}
			}
			
			// Check if output is JSON
			if tt.want.isJSON {
				var jsonData map[string]interface{}
				if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
					t.Errorf("Expected JSON output, got error: %v", err)
				}
			}
		})
	}
}

func TestLoggerWithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Output:    buf,
		Level:     LogLevelDebug,
		DebugMode: false,
		Pretty:    false,
	})
	
	// Create context with correlation ID
	ctx := WithCorrelationID(context.Background(), "test-correlation-123")
	
	// Log with context
	logger.Info(ctx, "test with correlation")
	
	// Check output contains correlation ID
	output := buf.String()
	if !strings.Contains(output, "test-correlation-123") {
		t.Error("Expected correlation ID in output")
	}
}

func TestLoggerWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Output: buf,
		Level:  LogLevelDebug,
		Pretty: false,
	})
	
	// Create logger with fields
	fieldLogger := logger.
		WithField("component", "test").
		WithField("version", "1.0.0")
	
	fieldLogger.Info(context.Background(), "test with fields")
	
	// Parse JSON output
	var jsonData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	// Check fields
	if jsonData["component"] != "test" {
		t.Error("Expected component field")
	}
	if jsonData["version"] != "1.0.0" {
		t.Error("Expected version field")
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		logLevel    LogLevel
		shouldLog   map[LogLevel]bool
	}{
		{
			logLevel: LogLevelError,
			shouldLog: map[LogLevel]bool{
				LogLevelDebug: false,
				LogLevelInfo:  false,
				LogLevelWarn:  false,
				LogLevelError: true,
				// Fatal cannot be tested as it exits the program
			},
		},
		{
			logLevel: LogLevelDebug,
			shouldLog: map[LogLevel]bool{
				LogLevelDebug: true,
				LogLevelInfo:  true,
				LogLevelWarn:  true,
				LogLevelError: true,
				// Fatal cannot be tested as it exits the program
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.logLevel.String(), func(t *testing.T) {
			for level, shouldLog := range tt.shouldLog {
				buf := &bytes.Buffer{}
				logger := New(Config{
					Output: buf,
					Level:  tt.logLevel,
					Pretty: false,
				})
				
				// Log at the test level
				switch level {
				case LogLevelDebug:
					logger.Debug(context.Background(), "debug message")
				case LogLevelInfo:
					logger.Info(context.Background(), "info message")
				case LogLevelWarn:
					logger.Warn(context.Background(), "warn message")
				case LogLevelError:
					logger.Error(context.Background(), nil, "error message")
				}
				
				hasOutput := buf.Len() > 0
				if hasOutput != shouldLog {
					t.Errorf("Level %s: expected output=%v, got output=%v", level.String(), shouldLog, hasOutput)
				}
			}
		})
	}
}

func TestErrorLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Output:    buf,
		Level:     LogLevelDebug,
		DebugMode: true,
		Pretty:    false,
	})
	
	// Create a test error
	testErr := &testError{msg: "test error", code: 123}
	
	logger.Error(context.Background(), testErr, "operation failed")
	
	// Parse output
	var jsonData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	// Check error fields
	if jsonData["error"] != "test error" {
		t.Error("Expected error message")
	}
	if !strings.Contains(jsonData["error_type"].(string), "testError") {
		t.Error("Expected error type")
	}
}

// testError is a simple error type for testing
type testError struct {
	msg  string
	code int
}

func (e *testError) Error() string {
	return e.msg
}

func TestStandardFieldBuilders(t *testing.T) {
	fields := Fields()
	
	// Test request fields
	reqFields := fields.Request("GET", "corr-123")
	if reqFields[FieldMethod] != "GET" {
		t.Error("Expected method field")
	}
	if reqFields[FieldCorrelationID] != "corr-123" {
		t.Error("Expected correlation ID field")
	}
	
	// Test response fields
	respFields := fields.Response("corr-123", 150, nil)
	if respFields[FieldCorrelationID] != "corr-123" {
		t.Error("Expected correlation ID field")
	}
	if respFields[FieldDuration] != int64(150) {
		t.Error("Expected duration field")
	}
	
	// Test connection fields
	connFields := fields.Connection("conn-456", "ready")
	if connFields[FieldConnectionID] != "conn-456" {
		t.Error("Expected connection ID field")
	}
	if connFields[FieldConnectionState] != "ready" {
		t.Error("Expected connection state field")
	}
}