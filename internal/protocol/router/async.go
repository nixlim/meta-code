package router

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

var (
	// ErrRouterShutdown is returned when operations are attempted on a shutdown router
	ErrRouterShutdown = errors.New("router is shutdown")

	// ErrQueueFull is returned when the request queue is full
	ErrQueueFull = errors.New("request queue is full")
)

// asyncRequest represents an async request being processed
type asyncRequest struct {
	ctx           context.Context
	request       *jsonrpc.Request
	correlationID string
	responseChan  chan<- *jsonrpc.Response
}

// AsyncRouter provides asynchronous request handling with correlation
type AsyncRouter struct {
	// Embedded router for handler registration
	*Router

	// Correlation tracker
	tracker *CorrelationTracker

	// Worker configuration
	workers   int
	queueSize int

	// Request handling
	requestChan chan asyncRequest

	// Middleware chain
	middleware *Chain

	// Lifecycle management
	shutdown chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
	running  bool
}

// AsyncRouterConfig holds configuration for AsyncRouter
type AsyncRouterConfig struct {
	Router     *Router
	Workers    int
	QueueSize  int
	Middleware []Middleware
}

// NewAsyncRouter creates a new AsyncRouter with the given configuration
func NewAsyncRouter(config AsyncRouterConfig) *AsyncRouter {
	if config.Router == nil {
		config.Router = New()
	}

	if config.Workers <= 0 {
		config.Workers = 10 // Default workers
	}

	if config.QueueSize <= 0 {
		config.QueueSize = 100 // Default queue size
	}

	ar := &AsyncRouter{
		Router:      config.Router,
		tracker:     NewCorrelationTracker(),
		workers:     config.Workers,
		queueSize:   config.QueueSize,
		requestChan: make(chan asyncRequest, config.QueueSize),
		middleware:  NewChain(config.Middleware...),
		shutdown:    make(chan struct{}),
	}

	return ar
}

// Start starts the async router workers
func (ar *AsyncRouter) Start() error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if ar.running {
		return errors.New("router already running")
	}

	ar.running = true

	// Start workers
	for i := 0; i < ar.workers; i++ {
		ar.wg.Add(1)
		go ar.worker(i)
	}

	return nil
}

// worker processes requests from the queue
func (ar *AsyncRouter) worker(id int) {
	defer ar.wg.Done()

	for {
		select {
		case req := <-ar.requestChan:
			ar.processRequest(req)
		case <-ar.shutdown:
			// Drain remaining requests with timeout
			timeout := time.NewTimer(5 * time.Second)
			defer timeout.Stop()

			for {
				select {
				case req := <-ar.requestChan:
					ar.processRequest(req)
				case <-timeout.C:
					return
				default:
					return
				}
			}
		}
	}
}

// processRequest handles a single request
func (ar *AsyncRouter) processRequest(asyncReq asyncRequest) {
	// Build the handler chain with middleware
	var handler Handler = ar.Router
	if ar.middleware != nil && len(ar.middleware.middlewares) > 0 {
		handler = ar.middleware.Then(ar.Router)
	}

	// Handle the request
	response := handler.Handle(asyncReq.ctx, asyncReq.request)

	// Send response
	select {
	case asyncReq.responseChan <- response:
		// Response sent successfully
	case <-asyncReq.ctx.Done():
		// Context cancelled, response discarded
	}
}

// HandleAsync handles a request asynchronously and returns a correlation ID
func (ar *AsyncRouter) HandleAsync(ctx context.Context, request *jsonrpc.Request) (string, error) {
	ar.mu.RLock()
	if !ar.running {
		ar.mu.RUnlock()
		return "", ErrRouterShutdown
	}
	ar.mu.RUnlock()

	// Generate correlation ID
	correlationID := ar.tracker.GenerateCorrelationID()

	// Create or get request context
	rc, ok := GetRequestContext(ctx)
	if !ok {
		rc = NewRequestContext(correlationID)
		ctx = WithRequestContext(ctx, rc)
	} else {
		rc.CorrelationID = correlationID
	}

	// Create response channel
	responseChan := make(chan *jsonrpc.Response, 1)

	// Create async request
	asyncReq := asyncRequest{
		ctx:           ctx,
		request:       request,
		correlationID: correlationID,
		responseChan:  responseChan,
	}

	// Register for correlation tracking BEFORE queuing
	ar.tracker.Register(correlationID)

	// Start goroutine to wait for response BEFORE queuing
	go func() {
		defer func() {
			// If context has a cancel function in metadata, call it
			if rc, ok := GetRequestContext(ctx); ok {
				if cancelFn, ok := rc.GetMetadata("_cancel"); ok {
					if cancel, ok := cancelFn.(context.CancelFunc); ok {
						cancel()
					}
				}
			}
		}()

		select {
		case response := <-responseChan:
			ar.tracker.Complete(correlationID, response)
		case <-ctx.Done():
			ar.tracker.CompleteWithError(correlationID, ctx.Err())
		}
	}()

	// Try to queue request AFTER setting up response handling
	select {
	case ar.requestChan <- asyncReq:
		// Request queued successfully
		return correlationID, nil
	default:
		// Queue full - clean up
		ar.tracker.Cancel(correlationID)
		close(responseChan)
		return "", ErrQueueFull
	}
}

// HandleAsyncWithTimeout handles a request asynchronously with a timeout
func (ar *AsyncRouter) HandleAsyncWithTimeout(ctx context.Context, request *jsonrpc.Request, timeout time.Duration) (string, error) {
	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)

	// Get or create request context
	rc, ok := GetRequestContext(timeoutCtx)
	if !ok {
		rc = NewRequestContext("")
		rc.Timeout = timeout
		timeoutCtx = WithRequestContext(timeoutCtx, rc)
	} else if rc.Timeout == 0 {
		rc.Timeout = timeout
	}

	// Store cancel function in metadata so it can be called after request completes
	rc.SetMetadata("_cancel", cancel)

	return ar.HandleAsync(timeoutCtx, request)
}

// GetResponse waits for a response with the given correlation ID
func (ar *AsyncRouter) GetResponse(correlationID string, timeout time.Duration) (*jsonrpc.Response, error) {
	return ar.tracker.WaitForResponse(correlationID, timeout)
}

// HandleAsyncWithCallback handles a request asynchronously and calls the callback with the response
func (ar *AsyncRouter) HandleAsyncWithCallback(ctx context.Context, request *jsonrpc.Request, callback func(*jsonrpc.Response, error)) error {
	correlationID, err := ar.HandleAsync(ctx, request)
	if err != nil {
		return err
	}

	// Get timeout from context if available
	timeout := 30 * time.Second // Default timeout
	if rc, ok := GetRequestContext(ctx); ok && rc.Timeout > 0 {
		timeout = rc.Timeout
	}

	// Wait for response in background
	go func() {
		response, err := ar.GetResponse(correlationID, timeout)
		callback(response, err)
	}()

	return nil
}

// Handle implements the Handler interface for synchronous compatibility
func (ar *AsyncRouter) Handle(ctx context.Context, request *jsonrpc.Request) *jsonrpc.Response {
	// Get timeout from context
	timeout := 30 * time.Second // Default timeout
	if rc, ok := GetRequestContext(ctx); ok && rc.Timeout > 0 {
		timeout = rc.Timeout
	}

	// Handle asynchronously
	correlationID, err := ar.HandleAsync(ctx, request)
	if err != nil {
		return &jsonrpc.Response{
			ID: request.ID,
			Error: jsonrpc.NewError(
				jsonrpc.ErrorCodeInternal,
				"Failed to process request",
				err.Error(),
			),
		}
	}

	// Wait for response
	response, err := ar.GetResponse(correlationID, timeout)
	if err != nil {
		return &jsonrpc.Response{
			ID: request.ID,
			Error: jsonrpc.NewError(
				jsonrpc.ErrorCodeTimeout,
				"Request timeout",
				err.Error(),
			),
		}
	}

	return response
}

// Shutdown gracefully shuts down the async router
func (ar *AsyncRouter) Shutdown(ctx context.Context) error {
	ar.mu.Lock()
	if !ar.running {
		ar.mu.Unlock()
		return nil
	}

	ar.running = false
	ar.mu.Unlock()

	// Signal shutdown
	close(ar.shutdown)

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		ar.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Workers finished
	case <-ctx.Done():
		return ctx.Err()
	}

	// Shutdown correlation tracker
	ar.tracker.Shutdown()

	return nil
}

// Stats returns statistics about the async router
type AsyncRouterStats struct {
	QueuedRequests  int
	PendingRequests int
	Workers         int
	Running         bool
}

// Stats returns current statistics
func (ar *AsyncRouter) Stats() AsyncRouterStats {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	trackerStats := ar.tracker.Stats()

	return AsyncRouterStats{
		QueuedRequests:  len(ar.requestChan),
		PendingRequests: trackerStats.PendingCount,
		Workers:         ar.workers,
		Running:         ar.running,
	}
}
