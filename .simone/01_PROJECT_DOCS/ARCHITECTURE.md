# Meta-MCP Server Architecture

## Project Overview

The Meta-MCP Server is an orchestration platform that aggregates multiple Model Context Protocol (MCP) servers to create intelligent, AI-driven development workflows. It acts as both an MCP client (connecting to multiple MCP servers) and an MCP server (exposing a unified interface to AI coding agents).

### Key Value Proposition
- **Workflow Orchestration**: Combines multiple MCP servers into cohesive workflows
- **AI-Powered Intelligence**: Leverages external AI APIs for workflow suggestions
- **Local-First Security**: All operations run locally with secure credential management
- **Developer Productivity**: 40-60% reduction in routine development workflow setup time

## Technical Architecture

### Core Technologies
- **Language**: Go (for performance, concurrency, and easy deployment)
- **Protocol**: MCP (Model Context Protocol) - JSON-RPC 2.0 based
- **MCP Implementation**: github.com/mark3labs/mcp-go v0.34.0 (standardized library)
- **Database**: Local SQLite for workflow state persistence
- **AI Integration**: External AI APIs (OpenAI, Anthropic, etc.)

### Recent Architectural Decision (2025-07-20)
**MCP Library Integration**: Adopted github.com/mark3labs/mcp-go as the foundation for MCP protocol implementation instead of custom implementation. This provides:
- Battle-tested, specification-compliant MCP implementation
- Automatic compliance with protocol updates
- Reduced development time and maintenance burden
- Focus on core business logic rather than protocol details

### System Components

#### 1. Core Server
- Go binary executable that implements MCP server protocol
- Listens on configurable port (default: 8000)
- Manages lifecycle of connected MCP servers
- Handles JSON-RPC message routing

#### 2. Connection Layer
- **STDIO Connector**: Spawns and manages subprocess MCP servers
- **HTTP/SSE Connector**: Connects to URL-based MCP servers
- **Connection Pool**: Manages multiple concurrent connections
- **Health Monitoring**: Tracks server availability and performance

#### 3. Command Catalog
- Dynamic discovery of available commands from connected servers
- Caches capabilities for performance
- Provides unified command interface
- Handles command translation and routing

#### 4. AI Integration Layer
- Interfaces with external AI APIs (OpenAI, Anthropic, etc.)
- Manages API keys securely (memory-only, never persisted)
- Implements retry logic and fallback strategies
- Provides workflow suggestion capabilities

#### 5. Workflow Engine
- Executes command sequences based on AI suggestions
- Manages workflow state in local SQLite database
- Implements rollback and error recovery
- Supports both automatic and manual workflows

#### 6. Security Layer
- TLS for all network communications
- Secure credential management (environment variables/CLI flags)
- User consent prompts for sensitive operations
- Input validation and sandboxing

### Data Flow

```
AI Coding Agent (e.g., Claude Code)
        ↓
Meta-MCP Server (JSON-RPC)
        ↓
┌─────────────────────────────┐
│   Connection Manager         │
├──────────┬──────────────────┤
│  STDIO   │   HTTP/SSE       │
└──────────┴──────────────────┘
        ↓
Connected MCP Servers
(File System, Git, Database, etc.)
```

## Design Principles

### 1. Local-First
- All components run on user's machine
- No external dependencies beyond AI API calls
- Data never leaves user's control

### 2. Stateless Core
- Core server processes requests without session state
- Workflow state persisted to local database
- Easy horizontal scaling if needed

### 3. Modular Architecture
- Independent, testable components
- Clear separation of concerns
- Plugin-style architecture for extensibility

### 4. AI-Assisted
- External AI for intelligence, no local ML models
- Lightweight footprint
- Flexible AI provider support

### 5. Security by Design
- Zero-trust approach to external connections
- Minimal attack surface
- Comprehensive audit logging

## Key Technical Decisions

### Why Go?
- Excellent concurrency support for managing multiple MCP connections
- Single binary deployment simplifies distribution
- Strong standard library for networking and subprocess management
- Good performance characteristics for I/O-bound workloads

### Why SQLite for State?
- Zero configuration database
- File-based storage aligns with local-first principle
- Sufficient performance for workflow state management
- Easy backup and portability

### Why External AI APIs?
- Avoids large local model requirements
- Access to state-of-the-art capabilities
- Reduces resource requirements on user machines
- Allows flexible provider selection

## Integration Points

### MCP Server Discovery
- Configuration via `~/.meta-mcp/config.json`
- Dynamic server registration/deregistration
- Health check endpoints for availability

### AI Coding Agent Integration
- Started via command line by agents like Claude Code
- Exposes MCP server interface on local port
- Provides status and monitoring endpoints

### External AI Services
- Configurable AI provider endpoints
- Support for multiple providers (OpenAI, Anthropic, etc.)
- Fallback strategies for availability

## Performance Considerations

### Concurrency Model
- Goroutines for each MCP server connection
- Non-blocking I/O for all network operations
- Connection pooling for HTTP-based servers

### Resource Management
- Configurable limits on subprocess count
- Memory limits for command output buffering
- Timeout management for all operations

### Caching Strategy
- Command catalog cached with TTL
- AI suggestion cache for repeated queries
- Connection state caching for performance

## Deployment Architecture

### Local Development
```
meta-mcp --port 8000 --config ~/.meta-mcp/config.json
```

### Production Usage
- Systemd service for Linux/macOS
- Windows Service for Windows
- Auto-start on system boot (optional)
- Graceful shutdown handling

## Future Extensibility

### Plugin System
- Dynamic loading of workflow templates
- Custom command processors
- Third-party integrations

### Multi-User Support
- Team workflow sharing
- Centralized configuration management
- Usage analytics and optimization

### Advanced AI Features
- Learning from user patterns
- Predictive workflow suggestions
- Multi-modal input support

## Technical Constraints

### System Requirements
- Go 1.24+ for development
- 100MB disk space for binary and database
- 256MB RAM minimum, 512MB recommended
- Network access for AI API calls

### Scalability Limits
- Designed for single-user workstation use
- Can handle 10-20 concurrent MCP server connections
- Workflow database can store millions of records

### Security Boundaries
- No remote access capabilities
- All operations require local authentication
- Sensitive operations require explicit consent