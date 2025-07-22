# Error Fixtures

This directory contains fixture files for testing error scenarios and error handling.

## Purpose
- Store various error response formats
- Test error handling and recovery mechanisms
- Validate error message formatting and codes

## Structure
- `validation/` - Validation error examples
- `system/` - System error examples
- `protocol/` - Protocol error examples

## Usage Example
```go
errorFixture := fixtures.LoadErrorFixture("validation/missing-field.json")
```