# System Patterns

## Architecture Overview

### Core Architecture Pattern
The Meta-MCP Server follows a modular, layered architecture designed for flexibility and testability, enhanced with Claude command orchestration:

```
┌─────────────────────────────────────────────────────────────────┐
│                    AI Coding Agent (Claude Code)                │
└─────────────────────┬───────────────────────────────────────────┘
                      │ MCP Client Connection
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Meta-MCP Server Core                       │
├─────────────────────────────────────────────────────────────────┤
│  Protocol Layer     │  Orchestration Layer  │  Integration Layer│
│  - MCP Handler      │  - Workflow Engine    │  - AI Integrator  │
│  - JSONRPC Codec    │  - Command Catalog    │  - Storage Manager│
│  - Message Router   │  - State Manager      │  - Config Manager │
└─────────────────────┬───────────────────────────────────────────┘
                      │ Multiple Connections
         ┌────────────┼────────────┐
         ▼            ▼            ▼
    MCP Servers (File, Git, Code Execution, etc.)
```

### Key Design Patterns

#### 0. Development Orchestration Pattern
- **Location:** `.claude/commands/`
- **Pattern:** Command-based development workflow with AI agent coordination
- **Components:**
  - `coordination/`: Swarm initialization and task orchestration
  - `automation/`: Automated agent spawning and workflow selection
  - `analysis/`: Performance and bottleneck detection
  - `github/`: Code review and repository management
  - `sparc/`: Specialized development modes

#### 1. Protocol Handler Pattern
- **Location:** `/internal/protocol/`
- **Pattern:** Layered protocol processing with clear separation of concerns
- **Components:**
  - `jsonrpc/`: Low-level JSON-RPC 2.0 implementation
  - `mcp/`: MCP-specific protocol handling
  - `router/`: Message routing and correlation
  - `handlers/`: Request/response processing

#### 2. Connection Management Pattern
- **Pattern:** Connection pooling with health monitoring
- **Key Features:**
  - Concurrent connection handling via goroutines
  - Transport abstraction (STDIO, HTTP/SSE)
  - Automatic reconnection with exponential backoff
  - Connection state management

#### 3. Command Aggregation Pattern
- **Pattern:** Dynamic catalog with conflict resolution
- **Implementation:**
  - Real-time discovery of available commands
  - Namespace isolation per server
  - Priority-based conflict resolution
  - Caching for performance

#### 4. Workflow Execution Pattern
- **Pattern:** State machine with persistence
- **Features:**
  - Sequential and parallel execution modes
  - Checkpoint-based recovery
  - Error handling with rollback
  - Progress monitoring

### Component Relationships

#### Protocol Layer
Handles all MCP communication:
```go
// Core interfaces
type Handler interface {
    HandleRequest(ctx context.Context, req *Request) (*Response, error)
}

type Router interface {
    Route(ctx context.Context, msg Message) error
    RegisterHandler(method string, handler Handler)
}
```

#### Orchestration Layer
Manages workflow execution:
```go
// Workflow interfaces
type Workflow interface {
    Execute(ctx context.Context) error
    GetState() WorkflowState
}

type Executor interface {
    RunWorkflow(ctx context.Context, workflow Workflow) error
}
```

#### Integration Layer
External system integration:
```go
// AI integration
type AIProvider interface {
    SuggestWorkflow(ctx context.Context, task string, catalog Catalog) (Workflow, error)
}

// Storage abstraction
type Storage interface {
    SaveState(ctx context.Context, id string, state interface{}) error
    LoadState(ctx context.Context, id string) (interface{}, error)
}
```

### Critical Implementation Paths

#### 1. Request Processing Flow
```
Client Request → JSONRPC Decode → Router → Handler → 
Command Execution → Response Assembly → JSONRPC Encode → Client Response
```

#### 2. Multi-Server Command Execution
```
Command Request → Catalog Lookup → Server Selection → 
Transport Selection (STDIO/HTTP) → Proxy Request → 
Aggregate Responses → Return Result
```

#### 3. AI-Assisted Workflow Generation
```
Task Description → Context Assembly → AI API Call → 
Response Parsing → Workflow Construction → Validation → 
Execution Plan
```

### Concurrency Patterns

#### 1. Connection Management
- One goroutine per MCP server connection
- Channel-based communication between connections
- Context-based cancellation for clean shutdown

#### 2. Request Handling
- Request-scoped goroutines with timeout
- Worker pool for CPU-intensive operations
- Rate limiting for external API calls

#### 3. State Management
- Read-write mutex for catalog updates
- Atomic operations for counters
- Channel-based event notifications

### Error Handling Strategy

#### 1. Error Types
- Protocol errors (MCP/JSONRPC violations)
- Connection errors (network, transport)
- Execution errors (command failures)
- System errors (resources, permissions)

#### 2. Error Propagation
- Wrapped errors with context
- Error codes mapping to MCP error codes
- Structured logging for debugging
- User-friendly error messages

#### 3. Recovery Mechanisms
- Automatic reconnection for connections
- Workflow checkpoint recovery
- Graceful degradation for AI failures
- Circuit breakers for external services

### Security Architecture

#### 1. Credential Management
- Environment variable loading only
- No persistent credential storage
- Memory-only credential handling
- Secure erasure on shutdown

#### 2. Access Control
- User consent prompts for sensitive operations
- Operation whitelisting/blacklisting
- Audit logging for all operations
- Rate limiting for resource protection

#### 3. Communication Security
- TLS 1.3 for all network communications
- Certificate validation
- Secure subprocess spawning
- Input sanitization and validation

### Performance Optimization

#### 1. Caching Strategy
- In-memory command catalog cache
- AI response caching with TTL
- Connection pooling
- Prepared statement equivalents

#### 2. Resource Management
- Bounded goroutine pools
- Memory limits for buffers
- Timeout enforcement
- Resource cleanup on context cancellation

#### 3. Scalability Considerations
- Horizontal scaling via load balancing
- Stateless core for easy replication
- Efficient message serialization
- Batch operation support

### Development Patterns

#### 1. Swarm-Based Development
- **Pattern:** Multi-agent coordination for complex tasks
- **Implementation:**
  - Hierarchical swarm topology
  - Specialized agents (coder, reviewer, tester)
  - Task orchestration with state management
  - Performance monitoring and metrics

#### 2. Memory-Driven Development
- **Pattern:** Persistent memory across sessions
- **Components:**
  - Memory bank for project context
  - Serena MCP for code navigation
  - Activity logging in `.claude-updates`
  - Knowledge synthesis and retrieval

#### 3. Command-Driven Workflows
- **Pattern:** Reusable command templates
- **Features:**
  - Pre-built workflows for common tasks
  - Parameterized commands
  - Hooks for customization
  - Integration with AI agents