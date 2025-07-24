# Tech Context

## Technology Stack

### Core Language: Go 1.24.2
- **Why Go:** Excellent concurrency support, strong standard library, easy deployment
- **Key Features Used:**
  - Goroutines for concurrent server management
  - Channels for inter-component communication
  - Context package for cancellation and timeouts
  - Interfaces for clean abstraction

### Development Environment
- **Platform:** Darwin (macOS) 24.5.0
- **Build Tool:** Go modules with Makefile automation
- **Version Control:** Git with structured commit messages
- **IDE Support:** IntelliJ IDEA integration

## Key Dependencies

### Protocol Implementation
- **JSON-RPC 2.0:** Custom implementation in `/internal/protocol/jsonrpc/`
  - Full spec compliance with batch support
  - Streaming parser for efficiency
  - Comprehensive error handling

### Testing Framework
- **Standard Library:** `testing` package for unit tests
- **Race Detection:** Built-in Go race detector
- **Coverage Tools:** Native Go coverage with HTML reports
- **Benchmarking:** Go benchmark framework

### External Communications
- **HTTP Client:** Standard `net/http` for API calls
- **TLS:** Native Go crypto/tls for secure communications
- **Process Management:** `os/exec` for subprocess spawning

### Storage Solutions
- **Configuration:** JSON files with schema validation
- **State Persistence:** BoltDB (planned) for embedded storage
- **Caching:** In-memory maps with sync primitives

### Development Tools Integration
- **Zen MCP:** Debug, analysis, and code review automation
- **Serena MCP:** Code navigation and memory management
- **Claude Flow:** Swarm orchestration and workflow automation
- **Browser Tools MCP:** UI testing and debugging support

## Development Setup

### Project Structure
```
meta-code/
├── cmd/
│   ├── server/        # Main server binary
│   └── meta-mcp/      # CLI tool (planned)
├── internal/
│   ├── protocol/      # MCP protocol implementation
│   ├── logging/       # Structured logging
│   └── testing/       # Test utilities
├── pkg/               # Public packages (future)
└── test/              # Integration tests
```

### Build Commands
```bash
# Build server
go build -o meta-code ./cmd/server

# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Generate coverage
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Environment Configuration
```bash
# Development mode
ENVIRONMENT=development ./meta-code

# Production mode
ENVIRONMENT=production ./meta-code

# With AI API key
OPENAI_API_KEY=sk-... ./meta-code
```

## Technical Constraints

### Performance Requirements
- Response time: <2 seconds for command routing
- Memory usage: <500MB under normal operation
- Concurrent connections: Support 50+ MCP servers
- Startup time: <5 seconds to operational state

### Compatibility Requirements
- MCP Protocol: Full specification compliance
- Go Version: 1.24.2 or later
- Platform Support: Linux, macOS, Windows
- Architecture: amd64, arm64

### Security Requirements
- TLS 1.3 minimum for network communications
- No credential persistence to disk
- Secure subprocess isolation
- Input validation and sanitization

## Development Patterns

### Code Style
- Standard Go formatting (`gofmt`)
- Meaningful variable names
- Comprehensive error handling
- Interface-based design

### Testing Approach
- Table-driven tests for comprehensive coverage
- Builder pattern for test data
- Separate test packages for isolation
- Integration tests for real scenarios

### Error Handling
```go
// Consistent error wrapping
if err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}

// Typed errors for different scenarios
type ValidationError struct {
    Field string
    Reason string
}
```

### Logging Strategy
- Structured logging with context
- Log levels: Debug, Info, Warn, Error
- Pretty printing in development
- JSON format in production

## Tool Usage Patterns

### Linting and Quality
```bash
# Run linters
golangci-lint run

# Format code
go fmt ./...

# Vet code
go vet ./...
```

### Debugging
- Delve debugger support
- Comprehensive logging
- Stack traces on panic
- Memory profiling with pprof

### Documentation
- Godoc comments for all public APIs
- README files for each package
- Architecture decision records
- Code examples in comments

## Deployment Considerations

### Binary Distribution
- Single static binary
- Cross-compilation support
- Version embedding
- Update mechanism (planned)

### Configuration Management
- JSON configuration files
- Environment variable overrides
- Hot reload support (SIGHUP)
- Configuration validation

### Monitoring and Observability
- Health check endpoints
- Metrics collection (planned)
- Distributed tracing (future)
- Log aggregation support