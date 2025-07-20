package router

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// RequestManager manages concurrent request processing with resource limits
type RequestManager struct {
	// Configuration
	maxConcurrent int
	maxQueued     int

	// Concurrency control
	semaphore chan struct{}
	queue     chan func()

	// Active request tracking
	activeRequests sync.Map // requestID -> *ActiveRequest
	requestCount   int64

	// Metrics
	metrics *ManagerMetrics

	// Lifecycle
	shutdown chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
	running  bool
}

// ActiveRequest represents an active request being processed
type ActiveRequest struct {
	ID            string
	CorrelationID string
	Method        string
	StartTime     time.Time
	Context       context.Context
	Cancel        context.CancelFunc
}

// ManagerMetrics tracks request manager statistics
type ManagerMetrics struct {
	TotalRequests     int64
	ActiveRequests    int64
	QueuedRequests    int64
	RejectedRequests  int64
	CompletedRequests int64
	TimeoutRequests   int64
	MaxQueueDepth     int64
	MaxActiveDuration time.Duration
	mu                sync.RWMutex
}

// ManagerConfig holds configuration for RequestManager
type ManagerConfig struct {
	MaxConcurrent int
	MaxQueued     int
}

// NewRequestManager creates a new RequestManager
func NewRequestManager(config ManagerConfig) *RequestManager {
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = 100
	}

	if config.MaxQueued <= 0 {
		config.MaxQueued = 1000
	}

	rm := &RequestManager{
		maxConcurrent: config.MaxConcurrent,
		maxQueued:     config.MaxQueued,
		semaphore:     make(chan struct{}, config.MaxConcurrent),
		queue:         make(chan func(), config.MaxQueued),
		metrics:       &ManagerMetrics{},
		shutdown:      make(chan struct{}),
	}

	return rm
}

// Start starts the request manager
func (rm *RequestManager) Start() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.running {
		return errors.New("manager already running")
	}

	rm.running = true

	// Start queue processor
	rm.wg.Add(1)
	go rm.processQueue()

	// Start metrics collector
	rm.wg.Add(1)
	go rm.collectMetrics()

	return nil
}

// processQueue processes queued requests
func (rm *RequestManager) processQueue() {
	defer rm.wg.Done()

	for {
		select {
		case fn := <-rm.queue:
			// Acquire semaphore
			select {
			case rm.semaphore <- struct{}{}:
				// Update metrics
				atomic.AddInt64(&rm.metrics.QueuedRequests, -1)
				atomic.AddInt64(&rm.metrics.ActiveRequests, 1)

				// Execute function
				go func() {
					defer func() {
						atomic.AddInt64(&rm.metrics.ActiveRequests, -1)
						<-rm.semaphore // Release semaphore
					}()
					fn()
				}()
			case <-rm.shutdown:
				return
			}
		case <-rm.shutdown:
			return
		}
	}
}

// collectMetrics periodically collects metrics
func (rm *RequestManager) collectMetrics() {
	defer rm.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rm.updateMetrics()
		case <-rm.shutdown:
			return
		}
	}
}

// updateMetrics updates manager metrics
func (rm *RequestManager) updateMetrics() {
	rm.metrics.mu.Lock()
	defer rm.metrics.mu.Unlock()

	// Calculate max duration of active requests
	var maxDuration time.Duration

	rm.activeRequests.Range(func(_, value interface{}) bool {
		req := value.(*ActiveRequest)
		duration := time.Since(req.StartTime)
		if duration > maxDuration {
			maxDuration = duration
		}
		return true
	})
	rm.metrics.MaxActiveDuration = maxDuration

	// Update max queue depth
	queueDepth := int64(len(rm.queue))
	if queueDepth > rm.metrics.MaxQueueDepth {
		rm.metrics.MaxQueueDepth = queueDepth
	}
}

// Execute executes a request with concurrency control
func (rm *RequestManager) Execute(ctx context.Context, requestID string, fn func(context.Context) error) error {
	rm.mu.RLock()
	if !rm.running {
		rm.mu.RUnlock()
		return errors.New("manager not running")
	}
	rm.mu.RUnlock()

	// Update metrics
	atomic.AddInt64(&rm.metrics.TotalRequests, 1)

	// Create cancellable context
	execCtx, cancel := context.WithCancel(ctx)

	// Get request context for metadata
	var correlationID, method string
	if rc, ok := GetRequestContext(ctx); ok {
		correlationID = rc.CorrelationID
		if m, ok := rc.GetMetadataString("method"); ok {
			method = m
		}
	}

	// Create active request
	activeReq := &ActiveRequest{
		ID:            requestID,
		CorrelationID: correlationID,
		Method:        method,
		StartTime:     time.Now(),
		Context:       execCtx,
		Cancel:        cancel,
	}

	// Track active request
	rm.activeRequests.Store(requestID, activeReq)

	// Create execution function
	execFn := func() {
		defer func() {
			rm.activeRequests.Delete(requestID)
			atomic.AddInt64(&rm.metrics.CompletedRequests, 1)
			cancel() // Ensure context is cancelled
		}()

		// Check if context already cancelled
		select {
		case <-execCtx.Done():
			atomic.AddInt64(&rm.metrics.TimeoutRequests, 1)
			return
		default:
		}

		// Execute the function
		if err := fn(execCtx); err != nil {
			// Error handling could be enhanced here
			if errors.Is(err, context.DeadlineExceeded) {
				atomic.AddInt64(&rm.metrics.TimeoutRequests, 1)
			}
		}
	}

	// Try direct execution
	select {
	case rm.semaphore <- struct{}{}:
		// Got semaphore, execute directly
		atomic.AddInt64(&rm.metrics.ActiveRequests, 1)
		go func() {
			defer func() {
				atomic.AddInt64(&rm.metrics.ActiveRequests, -1)
				<-rm.semaphore
			}()
			execFn()
		}()
		return nil

	default:
		// Try to queue
		select {
		case rm.queue <- execFn:
			atomic.AddInt64(&rm.metrics.QueuedRequests, 1)
			return nil

		default:
			// Queue full
			atomic.AddInt64(&rm.metrics.RejectedRequests, 1)
			cancel()
			return errors.New("request queue full")
		}
	}
}

// CancelRequest cancels an active request
func (rm *RequestManager) CancelRequest(requestID string) error {
	value, ok := rm.activeRequests.Load(requestID)
	if !ok {
		return errors.New("request not found")
	}

	activeReq := value.(*ActiveRequest)
	activeReq.Cancel()

	return nil
}

// GetActiveRequest retrieves information about an active request
func (rm *RequestManager) GetActiveRequest(requestID string) (*ActiveRequest, bool) {
	value, ok := rm.activeRequests.Load(requestID)
	if !ok {
		return nil, false
	}

	return value.(*ActiveRequest), true
}

// ListActiveRequests returns a list of all active requests
func (rm *RequestManager) ListActiveRequests() []*ActiveRequest {
	var requests []*ActiveRequest

	rm.activeRequests.Range(func(_, value interface{}) bool {
		req := value.(*ActiveRequest)
		requests = append(requests, req)
		return true
	})

	return requests
}

// Shutdown gracefully shuts down the request manager
func (rm *RequestManager) Shutdown(ctx context.Context) error {
	rm.mu.Lock()
	if !rm.running {
		rm.mu.Unlock()
		return nil
	}

	rm.running = false
	rm.mu.Unlock()

	// Signal shutdown
	close(rm.shutdown)

	// Cancel all active requests
	rm.activeRequests.Range(func(key, value interface{}) bool {
		activeReq := value.(*ActiveRequest)
		activeReq.Cancel()
		return true
	})

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		rm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetMetrics returns a copy of the current metrics
func (rm *RequestManager) GetMetrics() ManagerMetrics {
	rm.metrics.mu.RLock()
	defer rm.metrics.mu.RUnlock()

	return ManagerMetrics{
		TotalRequests:     atomic.LoadInt64(&rm.metrics.TotalRequests),
		ActiveRequests:    atomic.LoadInt64(&rm.metrics.ActiveRequests),
		QueuedRequests:    atomic.LoadInt64(&rm.metrics.QueuedRequests),
		RejectedRequests:  atomic.LoadInt64(&rm.metrics.RejectedRequests),
		CompletedRequests: atomic.LoadInt64(&rm.metrics.CompletedRequests),
		TimeoutRequests:   atomic.LoadInt64(&rm.metrics.TimeoutRequests),
		MaxQueueDepth:     rm.metrics.MaxQueueDepth,
		MaxActiveDuration: rm.metrics.MaxActiveDuration,
	}
}
