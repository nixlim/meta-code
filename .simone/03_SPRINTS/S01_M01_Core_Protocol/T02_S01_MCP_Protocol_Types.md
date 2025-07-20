# Task: MCP Protocol Types & Structures

## Task Metadata
- **Task ID**: T02_S01
- **Sprint**: S01
- **Status**: open
- **Complexity**: Medium
- **Dependencies**: T01_S01 (JSON-RPC Base Types)

## Task Details

### Title
MCP Protocol Types & Structures

### Description
Define MCP-specific protocol types and structures that build on top of the JSON-RPC foundation. This includes all MCP message types like Initialize, Initialized, and protocol constants according to the MCP specification.

### Goal/Objectives
- Define all MCP protocol message types
- Create protocol version constants and negotiation types
- Define capability structures for client/server negotiation
- Establish MCP-specific error codes
- Create type-safe method name constants

### Acceptance Criteria
- [ ] All MCP message types defined with proper JSON tags
- [ ] Protocol version negotiation structures in place
- [ ] Capability exchange types defined
- [ ] MCP error codes match specification
- [ ] Type safety for method names
- [ ] Documentation for all exported types

## Implementation

### Subtasks
- [ ] Define Initialize request structure
- [ ] Define Initialized response structure
- [ ] Create ProtocolVersion type with comparison methods
- [ ] Define ServerInfo and ClientInfo structures
- [ ] Create Capabilities structure for feature negotiation
- [ ] Define MCP-specific error codes as constants
- [ ] Create method name constants (e.g., MethodInitialize)
- [ ] Add validation methods for MCP messages

### Technical Guidance

#### Key interfaces and integration points:
- Build on top of JSON-RPC types from T01
- Define types in `internal/protocol/mcp/types.go`
- Use embedded structs to extend JSON-RPC base types
- Create constants in `internal/protocol/mcp/constants.go`

#### Existing patterns to follow:
- Use string enums with validation for protocol versions
- Embed JSON-RPC Request/Response for MCP messages
- Use const blocks for grouped constants
- Follow Go naming conventions (exported vs unexported)

#### Sample Code Structure:
```go
// internal/protocol/mcp/types.go
package mcp

import "github.com/nixlim/mcp-go-sdk/internal/protocol/jsonrpc"

// InitializeRequest represents the initialize message from client to server
type InitializeRequest struct {
    jsonrpc.Request
    Params InitializeParams `json:"params"`
}

// InitializeParams contains the parameters for initialization
type InitializeParams struct {
    ProtocolVersion string     `json:"protocolVersion"`
    ClientInfo      ClientInfo `json:"clientInfo"`
    Capabilities    Capabilities `json:"capabilities"`
}

// ProtocolVersion represents a semantic version for the MCP protocol
type ProtocolVersion string

// Compare returns -1, 0, or 1 if v is less than, equal to, or greater than other
func (v ProtocolVersion) Compare(other ProtocolVersion) int {
    // Implementation here
}
```

### Implementation Notes
1. Study the MCP specification to ensure all message types are covered
2. Use embedded structs to inherit JSON-RPC properties
3. Create validation methods as receivers on the types
4. Use string constants for method names to prevent typos
5. Consider creating a registry of supported methods
6. Define clear interfaces for extensibility
7. Ensure all types are JSON-serializable with proper tags

## Testing Requirements
- Unit tests for all type validation methods
- JSON marshaling/unmarshaling tests
- Protocol version comparison tests
- Capability negotiation tests
- Error code coverage tests

## Notes
- This task builds directly on the JSON-RPC foundation from T01_S01
- Pay special attention to the MCP specification for completeness
- Consider backward compatibility in protocol version design
- Ensure extensibility for future MCP protocol additions