# MCP Protocol Fixtures

This directory contains fixture files specific to the MCP (Model Context Protocol) implementation.

## Purpose
- Store MCP-specific message formats and protocol data
- Test MCP protocol compliance
- Validate MCP message handling and routing

## Structure
- `messages/` - MCP message examples
- `protocols/` - Protocol negotiation fixtures
- `capabilities/` - Capability declaration fixtures

## Usage Example
```go
mcpMessage := fixtures.LoadMCPFixture("messages/tool-call.json")
```