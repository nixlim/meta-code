package logging

import (
	"context"
	"log"
)

// StdLogAdapter wraps our Logger to implement the standard log.Logger interface
type StdLogAdapter struct {
	logger *Logger
	ctx    context.Context
}

// NewStdLogAdapter creates a new adapter that implements log.Logger interface
func NewStdLogAdapter(logger *Logger, ctx context.Context) *log.Logger {
	if ctx == nil {
		ctx = context.Background()
	}
	adapter := &StdLogAdapter{
		logger: logger,
		ctx:    ctx,
	}
	return log.New(adapter, "", 0)
}

// Write implements io.Writer interface for StdLogAdapter
func (a *StdLogAdapter) Write(p []byte) (n int, err error) {
	// Remove trailing newline if present
	message := string(p)
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[:len(message)-1]
	}
	
	// Log as info level by default
	a.logger.Info(a.ctx, message)
	
	return len(p), nil
}