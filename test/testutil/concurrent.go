package testutil

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// RunConcurrentTest runs a test function concurrently with the specified number of goroutines.
// It ensures all goroutines complete before returning and reports any panics through the testing framework.
func RunConcurrentTest(t *testing.T, numGoroutines int, testFunc func(id int)) {
	t.Helper() // Mark this as a test helper function
	
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	// Channel to collect panics from goroutines
	panicChan := make(chan interface{}, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					panicChan <- r
				}
				wg.Done()
			}()
			testFunc(id)
		}(i)
	}
	
	wg.Wait()
	close(panicChan)
	
	// Report any panics through the testing framework
	for panic := range panicChan {
		t.Errorf("Goroutine panicked: %v", panic)
	}
}

// RunConcurrentTestWithDone runs concurrent tests that signal completion via a channel.
// It provides better control over goroutine completion and supports timeout handling.
func RunConcurrentTestWithDone(t *testing.T, numGoroutines int, testFunc func(id int, done chan<- bool)) {
	t.Helper() // Mark this as a test helper function
	
	done := make(chan bool, numGoroutines)
	panicChan := make(chan interface{}, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					panicChan <- r
					done <- false // Signal completion even on panic
				}
			}()
			testFunc(id, done)
		}(i)
	}
	
	// Wait for all goroutines to complete with timeout
	timeout := time.After(30 * time.Second) // Reasonable timeout for concurrent tests
	completed := 0
	
	for completed < numGoroutines {
		select {
		case <-done:
			completed++
		case panic := <-panicChan:
			t.Errorf("Goroutine panicked: %v", panic)
		case <-timeout:
			t.Fatalf("Concurrent test timed out after 30s (completed %d/%d goroutines)", 
				completed, numGoroutines)
		}
	}
}

// RunConcurrentTestWithErrors runs concurrent tests and collects errors from each goroutine.
// It's useful when test functions need to report multiple errors or perform assertions.
func RunConcurrentTestWithErrors(t *testing.T, numGoroutines int, testFunc func(id int) error) {
	t.Helper()
	
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	// Buffered channel to collect errors
	errorChan := make(chan error, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			
			// Catch panics and convert to errors
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()
			
			if err := testFunc(id); err != nil {
				errorChan <- fmt.Errorf("goroutine %d: %w", id, err)
			}
		}(i)
	}
	
	wg.Wait()
	close(errorChan)
	
	// Report all errors
	errorCount := 0
	for err := range errorChan {
		t.Error(err)
		errorCount++
	}
	
	if errorCount > 0 {
		t.Errorf("Total errors in concurrent test: %d", errorCount)
	}
}

// AssertNoPanic ensures a function doesn't panic.
// It's useful for testing error handling and recovery mechanisms.
func AssertNoPanic(t *testing.T, f func(), msg string) {
	t.Helper()
	
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: function panicked with: %v", msg, r)
		}
	}()
	f()
}

// ConcurrentTestOptions provides configuration for concurrent tests.
type ConcurrentTestOptions struct {
	NumGoroutines int
	Timeout       time.Duration
	Description   string
}

// RunConcurrentTestWithOptions runs a concurrent test with configurable options.
// This provides the most flexibility for complex concurrent testing scenarios.
func RunConcurrentTestWithOptions(t *testing.T, opts ConcurrentTestOptions, testFunc func(id int) error) {
	t.Helper()
	
	if opts.NumGoroutines <= 0 {
		opts.NumGoroutines = 10 // Default
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second // Default
	}
	
	t.Logf("Running concurrent test: %s (goroutines: %d, timeout: %v)", 
		opts.Description, opts.NumGoroutines, opts.Timeout)
	
	var wg sync.WaitGroup
	wg.Add(opts.NumGoroutines)
	
	errorChan := make(chan error, opts.NumGoroutines)
	done := make(chan struct{})
	
	startTime := time.Now()
	
	for i := 0; i < opts.NumGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()
			
			if err := testFunc(id); err != nil {
				errorChan <- err
			}
		}(i)
	}
	
	// Wait for completion in a separate goroutine
	go func() {
		wg.Wait()
		close(done)
	}()
	
	// Wait with timeout
	select {
	case <-done:
		t.Logf("Concurrent test completed in %v", time.Since(startTime))
	case <-time.After(opts.Timeout):
		t.Fatalf("Concurrent test '%s' timed out after %v", opts.Description, opts.Timeout)
	}
	
	close(errorChan)
	
	// Collect and report errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
		t.Error(err)
	}
	
	if len(errors) > 0 {
		t.Errorf("Concurrent test '%s' failed with %d errors", opts.Description, len(errors))
	}
}