# Response Fixtures

This directory contains fixture files for testing various response scenarios in the meta-mcp-server.

## Purpose
- Store expected response data for assertions
- Mock external service responses
- Test response parsing and validation logic

## Structure
- `success/` - Successful response examples
- `errors/` - Error response examples
- `partial/` - Partial or streaming response examples

## Usage Example
```go
expected := fixtures.LoadResponseFixture("success/context-created.json")
assert.Equal(t, expected, actual)
```