# Concurrent Testing Guide

This guide documents the concurrent testing utilities provided in `testutil/concurrent.go` and best practices for writing concurrent tests.

## Available Utilities

### 1. RunConcurrentTest
Basic concurrent test runner with panic recovery.

```go
func TestConcurrentOperations(t *testing.T) {
    testutil.RunConcurrentTest(t, 10, func(id int) {
        // Your test logic here
        // id is the goroutine identifier (0-9)
    })
}
```

**Features:**
- Automatic panic recovery and reporting through testing.T
- Ensures all goroutines complete before returning
- Uses t.Helper() for better error reporting

### 2. RunConcurrentTestWithDone
For tests that need explicit completion signaling.

```go
func TestConcurrentWithCompletion(t *testing.T) {
    testutil.RunConcurrentTestWithDone(t, 5, func(id int, done chan<- bool) {
        // Your test logic here
        defer func() { done <- true }()
        
        // Complex async operation
    })
}
```

**Features:**
- 30-second timeout protection
- Panic recovery with completion signaling
- Progress tracking

### 3. RunConcurrentTestWithErrors
For tests that need to collect errors from each goroutine.

```go
func TestConcurrentWithErrors(t *testing.T) {
    testutil.RunConcurrentTestWithErrors(t, 20, func(id int) error {
        // Your test logic here
        if err := someOperation(); err != nil {
            return fmt.Errorf("operation failed: %w", err)
        }
        return nil
    })
}
```

**Features:**
- Collects and reports all errors
- Wraps errors with goroutine ID
- Summary of total errors

### 4. RunConcurrentTestWithOptions
Most flexible option for complex scenarios.

```go
func TestComplexConcurrent(t *testing.T) {
    testutil.RunConcurrentTestWithOptions(t, testutil.ConcurrentTestOptions{
        NumGoroutines: 100,
        Timeout:       2 * time.Minute,
        Description:   "Load testing user API",
    }, func(id int) error {
        // Your test logic here
        return nil
    })
}
```

**Features:**
- Configurable timeout
- Test description for logging
- Performance timing
- Comprehensive error collection

### 5. AssertNoPanic
Ensures a function doesn't panic.

```go
func TestErrorHandling(t *testing.T) {
    testutil.AssertNoPanic(t, func() {
        // Code that should not panic
        riskyOperation()
    }, "risky operation should handle errors gracefully")
}
```

## Migration Guide

### Before (Raw WaitGroup Pattern)
```go
func TestConcurrent(t *testing.T) {
    var wg sync.WaitGroup
    numGoroutines := 50
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            // test logic
        }(i)
    }
    
    wg.Wait()
}
```

### After (Using testutil)
```go
func TestConcurrent(t *testing.T) {
    testutil.RunConcurrentTest(t, 50, func(id int) {
        // test logic
    })
}
```

## Best Practices

1. **Choose the Right Utility**
   - Use `RunConcurrentTest` for simple concurrent tests
   - Use `RunConcurrentTestWithErrors` when you need error aggregation
   - Use `RunConcurrentTestWithOptions` for complex scenarios with custom timeouts

2. **Error Handling**
   - Always check for errors in concurrent operations
   - Use atomic operations for shared counters
   - Prefer channels for communication between goroutines

3. **Resource Management**
   - Set appropriate timeouts to prevent hanging tests
   - Use buffered channels to prevent goroutine leaks
   - Clean up resources in defer blocks

4. **Testing Considerations**
   - Run with `-race` flag to detect race conditions
   - Use reasonable goroutine counts (10-100 for most tests)
   - Consider system resources when setting concurrency levels

## Example: Complete Test Suite

```go
package mypackage_test

import (
    "context"
    "fmt"
    "sync/atomic"
    "testing"
    "time"
    
    "github.com/meta-mcp/meta-mcp-server/test/testutil"
)

func TestConcurrentAPI(t *testing.T) {
    // Simple concurrent test
    t.Run("BasicOperations", func(t *testing.T) {
        var successCount int32
        
        testutil.RunConcurrentTest(t, 20, func(id int) {
            if err := performOperation(id); err == nil {
                atomic.AddInt32(&successCount, 1)
            }
        })
        
        if successCount != 20 {
            t.Errorf("Expected 20 successes, got %d", successCount)
        }
    })
    
    // Test with error collection
    t.Run("ErrorHandling", func(t *testing.T) {
        testutil.RunConcurrentTestWithErrors(t, 10, func(id int) error {
            if id%3 == 0 {
                return fmt.Errorf("simulated error for id %d", id)
            }
            return nil
        })
    })
    
    // Complex test with options
    t.Run("LoadTest", func(t *testing.T) {
        testutil.RunConcurrentTestWithOptions(t, testutil.ConcurrentTestOptions{
            NumGoroutines: 100,
            Timeout:       30 * time.Second,
            Description:   "High load concurrent test",
        }, func(id int) error {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            
            return performComplexOperation(ctx, id)
        })
    })
}
```

## Performance Considerations

- The utilities add minimal overhead compared to raw goroutine management
- Panic recovery adds safety without significant performance impact
- Error collection is optimized with buffered channels
- Timeout mechanisms use efficient select statements

## Future Improvements

- Integration with testing.B for concurrent benchmarks
- Metrics collection for performance analysis
- Progressive backoff for retry scenarios
- Context propagation helpers