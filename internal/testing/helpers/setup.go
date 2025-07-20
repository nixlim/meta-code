// Package helpers provides test setup and teardown utilities
package helpers

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestSetup provides common test setup and teardown functionality
type TestSetup struct {
	t        *testing.T
	cleanups []func()
	mu       sync.Mutex
}

// NewSetup creates a new TestSetup instance
func NewSetup(t *testing.T) *TestSetup {
	setup := &TestSetup{
		t:        t,
		cleanups: make([]func(), 0),
	}

	// Register cleanup to run all cleanup functions
	t.Cleanup(func() {
		setup.runCleanups()
	})

	return setup
}

// AddCleanup adds a cleanup function to be run when the test completes
func (s *TestSetup) AddCleanup(cleanup func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanups = append(s.cleanups, cleanup)
}

// runCleanups runs all registered cleanup functions in reverse order
func (s *TestSetup) runCleanups() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Run cleanups in reverse order (LIFO)
	for i := len(s.cleanups) - 1; i >= 0; i-- {
		func() {
			defer func() {
				if r := recover(); r != nil {
					s.t.Errorf("Cleanup function panicked: %v", r)
				}
			}()
			s.cleanups[i]()
		}()
	}
}

// WithTimeout creates a context with timeout for testing
func (s *TestSetup) WithTimeout(timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	s.AddCleanup(cancel)
	return ctx
}

// WithCancel creates a cancellable context for testing
func (s *TestSetup) WithCancel() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	s.AddCleanup(cancel)
	return ctx, cancel
}

// SetupParallelTest configures a test for parallel execution
func SetupParallelTest(t *testing.T) {
	t.Helper()
	t.Parallel()
}

// SetupSubTest creates a subtest with proper setup
func SetupSubTest(t *testing.T, name string, testFunc func(*testing.T)) {
	t.Helper()
	t.Run(name, func(st *testing.T) {
		st.Helper()
		testFunc(st)
	})
}

// TestEnvironment provides a controlled test environment
type TestEnvironment struct {
	setup   *TestSetup
	started bool
	mu      sync.Mutex
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	return &TestEnvironment{
		setup: NewSetup(t),
	}
}

// Start initializes the test environment
func (e *TestEnvironment) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.started {
		return
	}

	e.started = true

	// Add any global setup here
	e.setup.AddCleanup(func() {
		e.started = false
	})
}

// AddCleanup adds a cleanup function to the environment
func (e *TestEnvironment) AddCleanup(cleanup func()) {
	e.setup.AddCleanup(cleanup)
}

// WithTimeout creates a context with timeout
func (e *TestEnvironment) WithTimeout(timeout time.Duration) context.Context {
	return e.setup.WithTimeout(timeout)
}

// WithCancel creates a cancellable context
func (e *TestEnvironment) WithCancel() (context.Context, context.CancelFunc) {
	return e.setup.WithCancel()
}

// MockTimer provides controllable time for testing
type MockTimer struct {
	currentTime time.Time
	mu          sync.RWMutex
}

// NewMockTimer creates a new mock timer starting at the given time
func NewMockTimer(startTime time.Time) *MockTimer {
	return &MockTimer{
		currentTime: startTime,
	}
}

// Now returns the current mock time
func (m *MockTimer) Now() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTime
}

// Advance advances the mock time by the given duration
func (m *MockTimer) Advance(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentTime = m.currentTime.Add(d)
}

// Set sets the mock time to the given time
func (m *MockTimer) Set(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentTime = t
}

// TestTimeout provides timeout utilities for tests
type TestTimeout struct {
	defaultTimeout time.Duration
}

// NewTestTimeout creates a new test timeout helper
func NewTestTimeout(defaultTimeout time.Duration) *TestTimeout {
	return &TestTimeout{
		defaultTimeout: defaultTimeout,
	}
}

// WithTimeout runs a function with a timeout
func (tt *TestTimeout) WithTimeout(t *testing.T, fn func()) {
	t.Helper()
	tt.WithTimeoutDuration(t, tt.defaultTimeout, fn)
}

// WithTimeoutDuration runs a function with a specific timeout
func (tt *TestTimeout) WithTimeoutDuration(t *testing.T, timeout time.Duration, fn func()) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
		// Function completed successfully
	case <-time.After(timeout):
		t.Fatalf("Test timed out after %v", timeout)
	}
}

// WaitGroup provides a testing-friendly wait group
type WaitGroup struct {
	wg sync.WaitGroup
	t  *testing.T
}

// NewWaitGroup creates a new test wait group
func NewWaitGroup(t *testing.T) *WaitGroup {
	return &WaitGroup{t: t}
}

// Add adds delta to the wait group counter
func (wg *WaitGroup) Add(delta int) {
	wg.wg.Add(delta)
}

// Done decrements the wait group counter
func (wg *WaitGroup) Done() {
	wg.wg.Done()
}

// Wait waits for the wait group counter to reach zero
func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}

// WaitWithTimeout waits for the wait group with a timeout
func (wg *WaitGroup) WaitWithTimeout(timeout time.Duration) {
	done := make(chan struct{})

	go func() {
		defer close(done)
		wg.wg.Wait()
	}()

	select {
	case <-done:
		// Wait group completed
	case <-time.After(timeout):
		wg.t.Fatalf("WaitGroup timed out after %v", timeout)
	}
}

// Barrier provides a synchronization barrier for tests
type Barrier struct {
	n     int
	count int
	ch    chan struct{}
	mu    sync.Mutex
	reset chan struct{}
}

// NewBarrier creates a new barrier for n goroutines
func NewBarrier(n int) *Barrier {
	return &Barrier{
		n:     n,
		ch:    make(chan struct{}),
		reset: make(chan struct{}),
	}
}

// Wait waits for all n goroutines to reach the barrier
func (b *Barrier) Wait() {
	b.mu.Lock()
	b.count++
	if b.count == b.n {
		close(b.ch)
		b.mu.Unlock()
		return
	}
	ch := b.ch
	b.mu.Unlock()

	<-ch
}

// Reset resets the barrier for reuse
func (b *Barrier) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.count > 0 {
		b.count = 0
		b.ch = make(chan struct{})
	}
}

// Eventually repeatedly calls a function until it succeeds or times out
func Eventually(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration, msgAndArgs ...interface{}) {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}

	require.Fail(t, "Condition never became true", msgAndArgs...)
}

// Never ensures a condition never becomes true within a timeout period
func Never(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration, msgAndArgs ...interface{}) {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			require.Fail(t, "Condition became true when it should not have", msgAndArgs...)
			return
		}
		time.Sleep(interval)
	}
}
