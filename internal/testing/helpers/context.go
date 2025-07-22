// Package helpers provides test context utilities
package helpers

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestContext provides enhanced context management for tests
type TestContext struct {
	ctx      context.Context
	cancel   context.CancelFunc
	t        *testing.T
	values   map[string]interface{}
	valuesMu sync.RWMutex
	metadata map[string]string
	metaMu   sync.RWMutex
	events   []ContextEvent
	eventsMu sync.RWMutex
}

// ContextEvent represents an event in the test context
type ContextEvent struct {
	Type      string
	Timestamp time.Time
	Message   string
	Data      interface{}
}

// NewTestContext creates a new test context
func NewTestContext(t *testing.T) *TestContext {
	ctx, cancel := context.WithCancel(context.Background())

	tc := &TestContext{
		ctx:      ctx,
		cancel:   cancel,
		t:        t,
		values:   make(map[string]interface{}),
		metadata: make(map[string]string),
		events:   make([]ContextEvent, 0),
	}

	// Register cleanup
	t.Cleanup(func() {
		tc.Cancel()
	})

	return tc
}

// WithTimeout creates a test context with timeout
func (tc *TestContext) WithTimeout(timeout time.Duration) *TestContext {
	ctx, cancel := context.WithTimeout(tc.ctx, timeout)

	newTC := &TestContext{
		ctx:      ctx,
		cancel:   cancel,
		t:        tc.t,
		values:   make(map[string]interface{}),
		metadata: make(map[string]string),
		events:   make([]ContextEvent, 0),
	}

	// Copy parent values
	tc.valuesMu.RLock()
	for k, v := range tc.values {
		newTC.values[k] = v
	}
	tc.valuesMu.RUnlock()

	// Copy parent metadata
	tc.metaMu.RLock()
	for k, v := range tc.metadata {
		newTC.metadata[k] = v
	}
	tc.metaMu.RUnlock()

	return newTC
}

// WithDeadline creates a test context with deadline
func (tc *TestContext) WithDeadline(deadline time.Time) *TestContext {
	ctx, cancel := context.WithDeadline(tc.ctx, deadline)

	newTC := &TestContext{
		ctx:      ctx,
		cancel:   cancel,
		t:        tc.t,
		values:   make(map[string]interface{}),
		metadata: make(map[string]string),
		events:   make([]ContextEvent, 0),
	}

	// Copy parent values and metadata
	tc.copyTo(newTC)

	return newTC
}

// Context returns the underlying context
func (tc *TestContext) Context() context.Context {
	return tc.ctx
}

// Cancel cancels the context
func (tc *TestContext) Cancel() {
	tc.cancel()
	tc.logEvent("context_cancelled", "Context cancelled", nil)
}

// Done returns the context's done channel
func (tc *TestContext) Done() <-chan struct{} {
	return tc.ctx.Done()
}

// Err returns the context's error
func (tc *TestContext) Err() error {
	return tc.ctx.Err()
}

// Deadline returns the context's deadline
func (tc *TestContext) Deadline() (deadline time.Time, ok bool) {
	return tc.ctx.Deadline()
}

// Set sets a value in the test context
func (tc *TestContext) Set(key string, value interface{}) {
	tc.valuesMu.Lock()
	defer tc.valuesMu.Unlock()

	tc.values[key] = value
	tc.logEvent("value_set", fmt.Sprintf("Set %s", key), map[string]interface{}{
		"key":   key,
		"value": value,
	})
}

// Get retrieves a value from the test context
func (tc *TestContext) Get(key string) (interface{}, bool) {
	tc.valuesMu.RLock()
	defer tc.valuesMu.RUnlock()

	value, ok := tc.values[key]
	return value, ok
}

// MustGet retrieves a value or fails the test
func (tc *TestContext) MustGet(key string) interface{} {
	tc.t.Helper()

	value, ok := tc.Get(key)
	require.True(tc.t, ok, "Key not found in context: %s", key)

	return value
}

// SetMetadata sets metadata in the test context
func (tc *TestContext) SetMetadata(key, value string) {
	tc.metaMu.Lock()
	defer tc.metaMu.Unlock()

	tc.metadata[key] = value
	tc.logEvent("metadata_set", fmt.Sprintf("Set metadata %s", key), map[string]string{
		"key":   key,
		"value": value,
	})
}

// GetMetadata retrieves metadata from the test context
func (tc *TestContext) GetMetadata(key string) (string, bool) {
	tc.metaMu.RLock()
	defer tc.metaMu.RUnlock()

	value, ok := tc.metadata[key]
	return value, ok
}

// AllMetadata returns all metadata
func (tc *TestContext) AllMetadata() map[string]string {
	tc.metaMu.RLock()
	defer tc.metaMu.RUnlock()

	result := make(map[string]string)
	for k, v := range tc.metadata {
		result[k] = v
	}
	return result
}

// LogEvent logs an event in the test context
func (tc *TestContext) LogEvent(eventType, message string, data interface{}) {
	tc.logEvent(eventType, message, data)
}

// logEvent internal event logging
func (tc *TestContext) logEvent(eventType, message string, data interface{}) {
	tc.eventsMu.Lock()
	defer tc.eventsMu.Unlock()

	event := ContextEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Message:   message,
		Data:      data,
	}

	tc.events = append(tc.events, event)
}

// GetEvents returns all logged events
func (tc *TestContext) GetEvents() []ContextEvent {
	tc.eventsMu.RLock()
	defer tc.eventsMu.RUnlock()

	result := make([]ContextEvent, len(tc.events))
	copy(result, tc.events)
	return result
}

// GetEventsByType returns events of a specific type
func (tc *TestContext) GetEventsByType(eventType string) []ContextEvent {
	tc.eventsMu.RLock()
	defer tc.eventsMu.RUnlock()

	result := make([]ContextEvent, 0)
	for _, event := range tc.events {
		if event.Type == eventType {
			result = append(result, event)
		}
	}
	return result
}

// Run runs a function with the test context
func (tc *TestContext) Run(fn func(context.Context)) {
	tc.t.Helper()

	tc.logEvent("run_start", "Starting context run", nil)
	defer tc.logEvent("run_end", "Completed context run", nil)

	fn(tc.ctx)
}

// RunWithTimeout runs a function with timeout
func (tc *TestContext) RunWithTimeout(timeout time.Duration, fn func(context.Context)) {
	tc.t.Helper()

	ctx, cancel := context.WithTimeout(tc.ctx, timeout)
	defer cancel()

	tc.logEvent("run_with_timeout_start", fmt.Sprintf("Starting with timeout %v", timeout), nil)
	defer tc.logEvent("run_with_timeout_end", "Completed timeout run", nil)

	done := make(chan struct{})

	go func() {
		defer close(done)
		fn(ctx)
	}()

	select {
	case <-done:
		// Function completed
	case <-ctx.Done():
		tc.t.Fatalf("Function timed out after %v", timeout)
	}
}

// copyTo copies values and metadata to another test context
func (tc *TestContext) copyTo(other *TestContext) {
	tc.valuesMu.RLock()
	for k, v := range tc.values {
		other.values[k] = v
	}
	tc.valuesMu.RUnlock()

	tc.metaMu.RLock()
	for k, v := range tc.metadata {
		other.metadata[k] = v
	}
	tc.metaMu.RUnlock()
}

// ContextBuilder builds test contexts with predefined values
type ContextBuilder struct {
	t        *testing.T
	values   map[string]interface{}
	metadata map[string]string
	timeout  time.Duration
	deadline time.Time
}

// NewContextBuilder creates a new context builder
func NewContextBuilder(t *testing.T) *ContextBuilder {
	return &ContextBuilder{
		t:        t,
		values:   make(map[string]interface{}),
		metadata: make(map[string]string),
	}
}

// WithValue adds a value to the context
func (cb *ContextBuilder) WithValue(key string, value interface{}) *ContextBuilder {
	cb.values[key] = value
	return cb
}

// WithMetadata adds metadata to the context
func (cb *ContextBuilder) WithMetadata(key, value string) *ContextBuilder {
	cb.metadata[key] = value
	return cb
}

// WithTimeout sets a timeout for the context
func (cb *ContextBuilder) WithTimeout(timeout time.Duration) *ContextBuilder {
	cb.timeout = timeout
	cb.deadline = time.Time{} // Clear deadline if timeout is set
	return cb
}

// WithDeadline sets a deadline for the context
func (cb *ContextBuilder) WithDeadline(deadline time.Time) *ContextBuilder {
	cb.deadline = deadline
	cb.timeout = 0 // Clear timeout if deadline is set
	return cb
}

// Build creates the test context
func (cb *ContextBuilder) Build() *TestContext {
	tc := NewTestContext(cb.t)

	// Apply timeout or deadline
	if cb.timeout > 0 {
		tc = tc.WithTimeout(cb.timeout)
	} else if !cb.deadline.IsZero() {
		tc = tc.WithDeadline(cb.deadline)
	}

	// Set values
	for k, v := range cb.values {
		tc.Set(k, v)
	}

	// Set metadata
	for k, v := range cb.metadata {
		tc.SetMetadata(k, v)
	}

	return tc
}

// ContextPool manages a pool of test contexts
type ContextPool struct {
	contexts []*TestContext
	mu       sync.Mutex
	t        *testing.T
}

// NewContextPool creates a new context pool
func NewContextPool(t *testing.T) *ContextPool {
	pool := &ContextPool{
		contexts: make([]*TestContext, 0),
		t:        t,
	}

	// Cleanup all contexts when test ends
	t.Cleanup(func() {
		pool.CancelAll()
	})

	return pool
}

// Create creates a new context in the pool
func (cp *ContextPool) Create() *TestContext {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	tc := NewTestContext(cp.t)
	cp.contexts = append(cp.contexts, tc)

	return tc
}

// CreateWithTimeout creates a new context with timeout in the pool
func (cp *ContextPool) CreateWithTimeout(timeout time.Duration) *TestContext {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	tc := NewTestContext(cp.t).WithTimeout(timeout)
	cp.contexts = append(cp.contexts, tc)

	return tc
}

// CancelAll cancels all contexts in the pool
func (cp *ContextPool) CancelAll() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for _, tc := range cp.contexts {
		tc.Cancel()
	}
}

// Size returns the number of contexts in the pool
func (cp *ContextPool) Size() int {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	return len(cp.contexts)
}

// ContextScope provides scoped context management
type ContextScope struct {
	parent *TestContext
	child  *TestContext
	t      *testing.T
}

// NewContextScope creates a new context scope
func NewContextScope(parent *TestContext, timeout time.Duration) *ContextScope {
	child := parent.WithTimeout(timeout)

	return &ContextScope{
		parent: parent,
		child:  child,
		t:      parent.t,
	}
}

// Enter enters the scope and returns the child context
func (cs *ContextScope) Enter() *TestContext {
	cs.child.logEvent("scope_enter", "Entering context scope", nil)
	return cs.child
}

// Exit exits the scope and cancels the child context
func (cs *ContextScope) Exit() {
	cs.child.logEvent("scope_exit", "Exiting context scope", nil)
	cs.child.Cancel()
}

// Run runs a function within the scope
func (cs *ContextScope) Run(fn func(*TestContext)) {
	cs.t.Helper()

	ctx := cs.Enter()
	defer cs.Exit()

	fn(ctx)
}

// Global context functions for convenience

// WithTestContext runs a function with a test context
func WithTestContext(t *testing.T, fn func(*TestContext)) {
	t.Helper()

	tc := NewTestContext(t)
	fn(tc)
}

// WithTestContextTimeout runs a function with a timeout context
func WithTestContextTimeout(t *testing.T, timeout time.Duration, fn func(*TestContext)) {
	t.Helper()

	tc := NewTestContext(t).WithTimeout(timeout)
	fn(tc)
}

// RunInContext runs a function with a basic context with timeout
func RunInContext(t *testing.T, timeout time.Duration, fn func(context.Context)) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		defer close(done)
		fn(ctx)
	}()

	select {
	case <-done:
		// Function completed
	case <-ctx.Done():
		t.Fatalf("Function timed out after %v", timeout)
	}
}

// MustCompleteWithin ensures a function completes within a timeout
func MustCompleteWithin(t *testing.T, timeout time.Duration, fn func()) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
		// Function completed
	case <-time.After(timeout):
		t.Fatalf("Function did not complete within %v", timeout)
	}
}

// ContextValue represents a typed context value
type ContextValue[T any] struct {
	key string
}

// NewContextValue creates a new typed context value
func NewContextValue[T any](key string) ContextValue[T] {
	return ContextValue[T]{key: key}
}

// Set sets the value in the context
func (cv ContextValue[T]) Set(tc *TestContext, value T) {
	tc.Set(cv.key, value)
}

// Get retrieves the value from the context
func (cv ContextValue[T]) Get(tc *TestContext) (T, bool) {
	var zero T

	value, ok := tc.Get(cv.key)
	if !ok {
		return zero, false
	}

	typed, ok := value.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

// MustGet retrieves the value or fails the test
func (cv ContextValue[T]) MustGet(tc *TestContext) T {
	tc.t.Helper()

	value, ok := cv.Get(tc)
	require.True(tc.t, ok, "Context value not found or wrong type: %s", cv.key)

	return value
}
