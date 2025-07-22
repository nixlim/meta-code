# Benchmarks

This directory contains benchmark tests for performance measurement and optimization.

## Purpose
- Measure performance of critical code paths
- Track performance regressions
- Identify optimization opportunities
- Compare different implementation approaches

## Structure
- `core/` - Core functionality benchmarks
- `handlers/` - Request handler benchmarks
- `memory/` - Memory usage benchmarks
- `concurrency/` - Concurrent operation benchmarks

## Running Benchmarks
```bash
# Run all benchmarks
go test -bench=. ./internal/testing/benchmarks/...

# Run specific benchmark
go test -bench=BenchmarkContextCreation ./internal/testing/benchmarks/core

# Run with memory profiling
go test -bench=. -benchmem ./internal/testing/benchmarks/...

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./internal/testing/benchmarks/...
```

## Writing Benchmarks
```go
func BenchmarkExample(b *testing.B) {
    // Setup
    setup()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Code to benchmark
        result := functionToBenchmark()
    }
}
```