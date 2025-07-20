// Package logging provides a centralized logging solution using zerolog
package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger and provides additional functionality
type Logger struct {
	logger    zerolog.Logger
	debugMode bool
	sanitize  bool
}

// LogLevel represents the severity level for logging
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Config holds logger configuration
type Config struct {
	// Output writer (defaults to os.Stderr)
	Output io.Writer
	// Level sets the minimum log level
	Level LogLevel
	// DebugMode enables debug logging and caller information
	DebugMode bool
	// Sanitize enables sanitization of sensitive data
	Sanitize bool
	// Pretty enables human-readable console output (for development)
	Pretty bool
}

// New creates a new Logger instance with the given configuration
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	// Convert our LogLevel to zerolog.Level
	var zlLevel zerolog.Level
	switch cfg.Level {
	case LogLevelDebug:
		zlLevel = zerolog.DebugLevel
	case LogLevelInfo:
		zlLevel = zerolog.InfoLevel
	case LogLevelWarn:
		zlLevel = zerolog.WarnLevel
	case LogLevelError:
		zlLevel = zerolog.ErrorLevel
	case LogLevelFatal:
		zlLevel = zerolog.FatalLevel
	default:
		zlLevel = zerolog.InfoLevel
	}

	// Configure zerolog
	zerolog.SetGlobalLevel(zlLevel)
	
	var zl zerolog.Logger
	if cfg.Pretty {
		// Development mode with pretty console output
		zl = zerolog.New(zerolog.ConsoleWriter{
			Out:        cfg.Output,
			TimeFormat: time.RFC3339,
		})
	} else {
		// Production mode with JSON output
		zl = zerolog.New(cfg.Output)
	}

	// Add timestamp to all logs
	zl = zl.With().Timestamp().Logger()

	// Add caller information if in debug mode
	if cfg.DebugMode {
		zl = zl.With().Caller().Logger()
	}

	return &Logger{
		logger:    zl,
		debugMode: cfg.DebugMode,
		sanitize:  cfg.Sanitize,
	}
}

// WithContext returns a new Logger with context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	newLogger := *l
	
	// Extract correlation ID if present
	if corrID := extractCorrelationID(ctx); corrID != "" {
		newLogger.logger = l.logger.With().Str("correlation_id", corrID).Logger()
	}
	
	// Extract other context values as needed
	// This can be extended based on your context keys
	
	return &newLogger
}

// WithCorrelationID returns a new Logger with the specified correlation ID
func (l *Logger) WithCorrelationID(id string) *Logger {
	newLogger := *l
	newLogger.logger = l.logger.With().Str("correlation_id", id).Logger()
	return &newLogger
}

// WithField returns a new Logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := *l
	newLogger.logger = l.logger.With().Interface(key, value).Logger()
	return &newLogger
}

// WithFields returns a new Logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := *l
	event := l.logger.With()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	newLogger.logger = event.Logger()
	return &newLogger
}

// WithComponent returns a new Logger with a component field
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithField(FieldComponent, component)
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, msg string) {
	l.WithContext(ctx).logger.Debug().Msg(msg)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, msg string) {
	l.WithContext(ctx).logger.Info().Msg(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, msg string) {
	l.WithContext(ctx).logger.Warn().Msg(msg)
}

// Error logs an error message with an error
func (l *Logger) Error(ctx context.Context, err error, msg string) {
	event := l.WithContext(ctx).logger.Error()
	if err != nil {
		event = event.Err(err)
		// Add error type for better debugging
		event = event.Str("error_type", fmt.Sprintf("%T", err))
	}
	event.Msg(msg)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(ctx context.Context, err error, msg string) {
	event := l.WithContext(ctx).logger.Fatal()
	if err != nil {
		event = event.Err(err)
		event = event.Str("error_type", fmt.Sprintf("%T", err))
	}
	event.Msg(msg)
}

// LogError logs an error with automatic caller information
func (l *Logger) LogError(ctx context.Context, err error, level LogLevel, message string) {
	if err == nil {
		return
	}

	// Get caller information if in debug mode
	var callerFile string
	var callerLine int
	var callerFunc string
	
	if l.debugMode {
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			callerFile = file
			callerLine = line
			if fn := runtime.FuncForPC(pc); fn != nil {
				callerFunc = fn.Name()
			}
		}
	}

	// Create event based on level
	var event *zerolog.Event
	logger := l.WithContext(ctx).logger

	switch level {
	case LogLevelDebug:
		event = logger.Debug()
	case LogLevelInfo:
		event = logger.Info()
	case LogLevelWarn:
		event = logger.Warn()
	case LogLevelError:
		event = logger.Error()
	case LogLevelFatal:
		event = logger.Fatal()
	default:
		event = logger.Error()
	}

	// Add error information
	event = event.Err(err).
		Str("error_type", fmt.Sprintf("%T", err))

	// Add caller information if available
	if l.debugMode && callerFile != "" {
		event = event.
			Str("caller_file", callerFile).
			Int("caller_line", callerLine)
		if callerFunc != "" {
			event = event.Str("caller_func", callerFunc)
		}
	}

	event.Msg(message)
}

// LogWithRecovery executes a function and logs any panic that occurs
func (l *Logger) LogWithRecovery(ctx context.Context, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			l.WithContext(ctx).logger.Error().
				Interface("panic", r).
				Str("stack", string(debug.Stack())).
				Msg("Panic recovered")
		}
	}()
	fn()
}

// GetZerolog returns the underlying zerolog.Logger for advanced usage
// This should be used sparingly to maintain abstraction
func (l *Logger) GetZerolog() zerolog.Logger {
	return l.logger
}

// Global logger instance
var defaultLogger *Logger

func init() {
	// Initialize with a basic configuration
	defaultLogger = New(Config{
		Level:     LogLevelInfo,
		DebugMode: false,
		Sanitize:  true,
		Pretty:    false,
	})
}

// SetDefault sets the default global logger
func SetDefault(logger *Logger) {
	defaultLogger = logger
}

// Default returns the default global logger
func Default() *Logger {
	return defaultLogger
}

// Convenience functions using the default logger

// Debug logs a debug message using the default logger
func Debug(ctx context.Context, msg string) {
	defaultLogger.Debug(ctx, msg)
}

// Info logs an info message using the default logger
func Info(ctx context.Context, msg string) {
	defaultLogger.Info(ctx, msg)
}

// Warn logs a warning message using the default logger
func Warn(ctx context.Context, msg string) {
	defaultLogger.Warn(ctx, msg)
}

// Error logs an error message using the default logger
func Error(ctx context.Context, err error, msg string) {
	defaultLogger.Error(ctx, err, msg)
}

// Fatal logs a fatal message using the default logger and exits
func Fatal(ctx context.Context, err error, msg string) {
	defaultLogger.Fatal(ctx, err, msg)
}


// Import debug for stack traces
var debug struct {
	Stack func() []byte
}

func init() {
	debug.Stack = func() []byte {
		buf := make([]byte, 1024*1024)
		n := runtime.Stack(buf, false)
		return buf[:n]
	}
}