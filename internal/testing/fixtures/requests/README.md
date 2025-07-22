# Request Fixtures

This directory contains fixture files for testing various types of requests in the meta-mcp-server.

## Purpose
- Store mock request data for unit and integration tests
- Provide consistent test data across different test suites
- Enable testing of edge cases and error scenarios

## Structure
- `valid/` - Valid request examples
- `invalid/` - Invalid request examples for error testing
- `edge-cases/` - Edge case scenarios

## Usage Example
```go
fixture := fixtures.LoadRequestFixture("valid/create-context.json")
```