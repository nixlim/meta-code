# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Workflow

Code Style: refer to .claude/default-rules.md

When using third party libraries, use `go doc` to read the documentation and understand how to use the library correctly.
Avoid writing custom implementations if the library provides the functionality you need.

## Project Overview

Meta-MCP Server is an orchestration platform that aggregates multiple Model Context Protocol (MCP) servers to create intelligent, AI-driven development workflows. It acts as both an MCP client (connecting to multiple MCP servers) and an MCP server (exposing a unified interface to AI coding agents).

**Current Status**: M01 Sprint S01 - MCP Protocol Foundation (90% complete, 9/10 tasks done)
**Next Task**: T10_S01 - Protocol Conformance

## Architecture

The codebase follows hexagonal architecture with clear component boundaries:

### Core Protocol Stack
- **JSON-RPC 2.0 Foundation** (`internal/protocol/jsonrpc/`) - Complete JSON-RPC 2.0 implementation with 93.3% test coverage
- **MCP Protocol Layer** (`internal/protocol/mcp/`) - MCP-specific types and handshake implementation using github.com/mark3labs/mcp-go v0.34.0
- **Message Router** (`internal/protocol/router/`) - Thread-safe request routing with async support, middleware, and correlation tracking
- **Connection Management** (`internal/protocol/connection/`) - Connection state management (New → Initializing → Ready → Closed)

### Key Architectural Decisions
- **mcp-go Integration** (2025-07-20): Uses github.com/mark3labs/mcp-go library instead of custom MCP implementation for standardization and reduced maintenance burden
- **Transport Agnostic**: Clean interfaces support both STDIO and HTTP/SSE transports
- **Async Request Handling**: Worker pool-based async router with correlation tracking and graceful shutdown
- **Handshake Protocol**: Full MCP Initialize/Initialized handshake with timeout protection and state validation

## Development Commands

### Building and Running
```bash
# Build the server
go build -o server ./cmd/server

# Run the server
./server

# Install dependencies
go mod tidy
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/protocol/jsonrpc -v

# Run integration tests
go test -run TestHandshakeIntegration ./...

# Skip long-running tests
go test -short ./...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Run linter (golangci-lint available at /usr/local/bin/golangci-lint)
golangci-lint run ./...

# Vet code
go vet ./...
```

### Current Test Coverage Status
- `jsonrpc`: 93.3% ✅
- `connection`: 87.0% ✅  
- `router`: 86.5% ⚠️ (some async tests failing)
- `mcp`: 63.9% ⚠️
- `handlers`: 42.2% ❌

## Simone Project Management System

This project uses the Simone task management system with Claude Code slash commands:

### Key Commands
- `/simone:do_task T06_S01` - Execute specific task
- `/simone:do_task` - Execute next open task in current sprint
- `/simone:commit T06_S01` - Commit task changes to git
- `/simone:code_review` - Run comprehensive code review

### Task Structure
Tasks are located in `.simone/03_SPRINTS/S01_M01_Core_Protocol/` with format:
- `T[ID]_S01_[Name].md` - Open tasks
- `T[ID]_S01_[Name]_COMPLETED.md` - Completed tasks

### Development Workflow
1. Use `/simone:do_task` to get next task
2. Follow 8-step process: scope analysis → task execution → code review → completion
3. Code review runs in parallel subagents with automated linting/type-checking
4. Use `/simone:commit` to create logical git commits
5. Tasks automatically update project manifest and sprint progress

## Component Integration Patterns

### MCP Server Setup
```go
// Use HandshakeServer for proper MCP protocol compliance
config := mcp.HandshakeConfig{
    Name:              "Meta-MCP Server", 
    Version:           "1.0.0",
    HandshakeTimeout:  30 * time.Second,
    SupportedVersions: []string{"1.0", "0.1.0"},
    ServerOptions: []server.ServerOption{
        mcp.WithToolCapabilities(true),
        mcp.WithResourceCapabilities(true, true),
    },
}
server := mcp.NewHandshakeServer(config)
```

### Router Usage
```go
// Create router with async support
router := router.New()
asyncRouter := router.NewAsyncRouter(router, 4) // 4 workers

// Register handlers
router.RegisterFunc("echo", echoHandler)
router.RegisterNotificationFunc("log", logHandler)
```

### Connection State Management
- All connections go through handshake flow before allowing requests
- Use connection context to track state and metadata
- Timeout protection prevents hanging connections

## Common Issues

### Router Test Failures
Some async router tests are currently failing due to concurrent execution issues. When working on router components, run tests individually:
```bash
go test -v ./internal/protocol/router -run TestSpecificTest
```

### MCP Integration
Always prefer mcp-go library functions over custom implementations. Check `go doc github.com/mark3labs/mcp-go` before implementing custom functionality.

### Pre-commit Validation
Before committing, ensure:
1. All tests pass: `go test ./...`
2. Code is formatted: `go fmt ./...` 
3. No linting errors: `golangci-lint run ./...`

## Project Context

- **Go Version**: 1.24.2
- **Main Dependency**: github.com/mark3labs/mcp-go v0.34.0
- **Architecture**: Local-first with external AI API integration
- **Transport Support**: STDIO (current), HTTP/SSE (planned)
- **Protocol**: MCP over JSON-RPC 2.0

The project is currently in M01 (MVP Foundation) focusing on core protocol implementation before moving to connection orchestration, AI integration, and production features in subsequent milestones.