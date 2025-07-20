package router

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

var (
	// ErrCorrelationNotFound is returned when a correlation ID is not found
	ErrCorrelationNotFound = errors.New("correlation ID not found")

	// ErrCorrelationTimeout is returned when waiting for a response times out
	ErrCorrelationTimeout = errors.New("correlation timeout")
)

// responseChannel holds a response and any error
type responseChannel struct {
	response chan *jsonrpc.Response
	error    chan error
	closed   bool
	mu       sync.Mutex
}

// safeClose safely closes the channels if not already closed
func (rc *responseChannel) safeClose() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if !rc.closed {
		close(rc.response)
		close(rc.error)
		rc.closed = true
	}
}

// CorrelationTracker manages request/response correlation for async operations
type CorrelationTracker struct {
	// pending maps correlation IDs to response channels
	pending sync.Map

	// cleanupInterval specifies how often to clean up expired entries
	cleanupInterval time.Duration

	// done signals shutdown
	done chan struct{}

	// wg tracks cleanup goroutine
	wg sync.WaitGroup
}

// NewCorrelationTracker creates a new CorrelationTracker
func NewCorrelationTracker() *CorrelationTracker {
	ct := &CorrelationTracker{
		cleanupInterval: 30 * time.Second,
		done:            make(chan struct{}),
	}

	// Start cleanup goroutine
	ct.wg.Add(1)
	go ct.cleanupLoop()

	return ct
}

// GenerateCorrelationID creates a new unique correlation ID
func (ct *CorrelationTracker) GenerateCorrelationID() string {
	return uuid.New().String()
}

// Register registers a new correlation ID and returns channels for the response
func (ct *CorrelationTracker) Register(correlationID string) (<-chan *jsonrpc.Response, <-chan error) {
	respChan := &responseChannel{
		response: make(chan *jsonrpc.Response, 1),
		error:    make(chan error, 1),
	}

	ct.pending.Store(correlationID, respChan)

	return respChan.response, respChan.error
}

// GetSendChannels returns send channels for direct worker access
func (ct *CorrelationTracker) GetSendChannels(correlationID string) (chan<- *jsonrpc.Response, chan<- error, bool) {
	value, ok := ct.pending.Load(correlationID)
	if !ok {
		return nil, nil, false
	}

	respChan := value.(*responseChannel)
	return respChan.response, respChan.error, true
}

// Complete completes a correlation with a response
func (ct *CorrelationTracker) Complete(correlationID string, response *jsonrpc.Response) error {
	value, ok := ct.pending.Load(correlationID)
	if !ok {
		return ErrCorrelationNotFound
	}

	respChan := value.(*responseChannel)

	// Use mutex to coordinate with safeClose()
	respChan.mu.Lock()
	defer respChan.mu.Unlock()

	if respChan.closed {
		return errors.New("response channel already closed")
	}

	select {
	case respChan.response <- response:
		// ✅ FIXED: Don't close channels here - let consumer handle cleanup
	default:
		// Channel already closed or full
		return errors.New("response channel blocked")
	}

	return nil
}

// CompleteWithError completes a correlation with an error
func (ct *CorrelationTracker) CompleteWithError(correlationID string, err error) error {
	value, ok := ct.pending.Load(correlationID)
	if !ok {
		return ErrCorrelationNotFound
	}

	respChan := value.(*responseChannel)

	// Use mutex to coordinate with safeClose()
	respChan.mu.Lock()
	defer respChan.mu.Unlock()

	if respChan.closed {
		return errors.New("error channel already closed")
	}

	select {
	case respChan.error <- err:
		// ✅ FIXED: Don't close channels here - let consumer handle cleanup
	default:
		// Channel already closed or full
		return errors.New("error channel blocked")
	}

	return nil
}

// Cancel cancels a pending correlation
func (ct *CorrelationTracker) Cancel(correlationID string) {
	value, ok := ct.pending.LoadAndDelete(correlationID)
	if !ok {
		return
	}

	respChan := value.(*responseChannel)
	respChan.safeClose()
}

// WaitForResponse waits for a response with the given correlation ID
func (ct *CorrelationTracker) WaitForResponse(correlationID string, timeout time.Duration) (*jsonrpc.Response, error) {
	value, ok := ct.pending.Load(correlationID)
	if !ok {
		return nil, ErrCorrelationNotFound
	}

	respChan := value.(*responseChannel)

	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case response := <-respChan.response:
			respChan.safeClose()             // ✅ FIXED: Close channels after consuming response
			ct.pending.Delete(correlationID) // ✅ FIXED: Delete after consuming response
			return response, nil
		case err := <-respChan.error:
			respChan.safeClose()             // ✅ FIXED: Close channels after consuming error
			ct.pending.Delete(correlationID) // ✅ FIXED: Delete after consuming error
			return nil, err
		case <-timer.C:
			ct.Cancel(correlationID) // Cancel already handles deletion and closing
			return nil, ErrCorrelationTimeout
		}
	}

	// No timeout, wait indefinitely
	select {
	case response := <-respChan.response:
		respChan.safeClose()             // ✅ FIXED: Close channels after consuming response
		ct.pending.Delete(correlationID) // ✅ FIXED: Delete after consuming response
		return response, nil
	case err := <-respChan.error:
		respChan.safeClose()             // ✅ FIXED: Close channels after consuming error
		ct.pending.Delete(correlationID) // ✅ FIXED: Delete after consuming error
		return nil, err
	}
}

// cleanupLoop periodically cleans up leaked correlations
func (ct *CorrelationTracker) cleanupLoop() {
	defer ct.wg.Done()

	ticker := time.NewTicker(ct.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ct.cleanup()
		case <-ct.done:
			return
		}
	}
}

// cleanup removes expired entries
func (ct *CorrelationTracker) cleanup() {
	// In a production system, we would track creation time
	// and remove entries older than a threshold
	// For now, this is a placeholder
}

// Shutdown gracefully shuts down the correlation tracker
func (ct *CorrelationTracker) Shutdown() {
	close(ct.done)
	ct.wg.Wait()

	// Cancel all pending correlations
	ct.pending.Range(func(key, value interface{}) bool {
		correlationID := key.(string)
		ct.Cancel(correlationID)
		return true
	})
}

// Stats returns statistics about the correlation tracker
type CorrelationStats struct {
	PendingCount int
}

// Stats returns current statistics
func (ct *CorrelationTracker) Stats() CorrelationStats {
	count := 0
	ct.pending.Range(func(_, _ interface{}) bool {
		count++
		return true
	})

	return CorrelationStats{
		PendingCount: count,
	}
}
