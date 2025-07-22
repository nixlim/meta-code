# Test Scenarios

This directory contains complete end-to-end test scenarios that simulate real-world usage patterns.

## Purpose
- Test complete workflows and user journeys
- Validate integration between components
- Test complex multi-step operations
- Ensure system behavior under realistic conditions

## Structure
- `basic/` - Basic usage scenarios
- `advanced/` - Complex multi-step scenarios
- `error/` - Error handling scenarios
- `performance/` - Performance-critical scenarios
- `concurrent/` - Concurrent operation scenarios

## Scenario Format
Each scenario should include:
1. **Description** - What the scenario tests
2. **Prerequisites** - Required setup
3. **Steps** - Detailed test steps
4. **Expected Results** - What should happen
5. **Validation** - How to verify success

## Example Scenario
```go
// Scenario: Create and manage multiple contexts
func TestMultiContextScenario(t *testing.T) {
    // 1. Create primary context
    primaryCtx := createContext("primary", "project")
    
    // 2. Create secondary context with reference
    secondaryCtx := createContext("secondary", "module")
    linkContexts(primaryCtx, secondaryCtx)
    
    // 3. Perform operations on both
    updateContext(primaryCtx, "status", "active")
    updateContext(secondaryCtx, "parent", primaryCtx.ID)
    
    // 4. Validate relationships
    assert.Equal(t, "active", primaryCtx.Status)
    assert.Equal(t, primaryCtx.ID, secondaryCtx.Parent)
    
    // 5. Cleanup
    deleteContext(secondaryCtx)
    deleteContext(primaryCtx)
}
```

## Running Scenarios
```bash
# Run all scenarios
go test ./internal/testing/scenarios/...

# Run specific scenario set
go test ./internal/testing/scenarios/basic

# Run with verbose output
go test -v ./internal/testing/scenarios/...
```