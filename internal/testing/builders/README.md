# Test Builders

This directory contains builder patterns and utilities for constructing test data and objects.

## Purpose
- Provide fluent interfaces for creating test objects
- Reduce boilerplate in test setup
- Ensure consistent test data creation
- Support complex object construction with defaults

## Structure
- `request_builder.go` - Build request objects for testing
- `response_builder.go` - Build response objects for testing
- `context_builder.go` - Build context objects for testing
- `fixture_builder.go` - Build fixture data for testing

## Usage Example
```go
// Using request builder
request := builders.NewRequest().
    WithMethod("context/create").
    WithParam("name", "test-context").
    WithParam("type", "project").
    Build()

// Using context builder
context := builders.NewContext().
    WithName("test-context").
    WithMetadata("key", "value").
    WithConfig(configObj).
    Build()

// Using response builder
response := builders.NewResponse().
    WithSuccess(true).
    WithResult(resultData).
    WithID("test-001").
    Build()
```

## Benefits
- Readable test setup
- Easy modification of test data
- Default values for common scenarios
- Type-safe construction