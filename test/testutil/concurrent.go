package testutil

import (
	"sync"
	"testing"
)

// RunConcurrentTest runs a test function concurrently with the specified number of goroutines.
func RunConcurrentTest(t *testing.T, numGoroutines int, testFunc func(id int)) {
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			testFunc(id)
		}(i)
	}
	
	wg.Wait()
}

// RunConcurrentTestWithDone runs concurrent tests that signal completion via a channel.
func RunConcurrentTestWithDone(t *testing.T, numGoroutines int, testFunc func(id int, done chan<- bool)) {
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			testFunc(id, done)
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// AssertNoPanic ensures a function doesn't panic.
func AssertNoPanic(t *testing.T, f func(), msg string) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: function panicked with: %v", msg, r)
		}
	}()
	f()
}