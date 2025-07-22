# Meta-MCP Server

A Model Context Protocol (MCP) server implementation providing tools and resources for AI assistants.

## Prerequisites

- Go 1.24.2 or later
- Git

## Installation

1. Clone the repository:
   ```bash
   git clone git@github.com:nixlim/meta-code.git
   cd meta-code
   ```

2. Build the server:
   ```bash
   go build -o meta-code ./cmd/server
   ```

## Quick Start

Start the MCP server:
```bash
./meta-code
```

The server will start and listen for MCP protocol messages via stdin/stdout.

## Testing

The Meta-MCP Server includes a comprehensive testing framework. See [docs/testing.md](docs/testing.md) for detailed testing guidelines.

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/protocol/jsonrpc

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### Test Coverage Goals

| Package | Target Coverage |
|---------|----------------|
| `jsonrpc` | 93.3% |
| `connection` | 87.0% |
| `router` | 82.1% |
| `mcp` | 78.5% |
| Overall | 70%+ |

### Environment Variables

Configure the server behavior with these optional environment variables:

- `ENVIRONMENT` or `ENV` or `GO_ENV`: Set to `development`, `staging`, or `production`
  - `development`: Pretty logging, debug mode enabled
  - `production`: JSON logging, info level (default)

### Example

```bash
# Run in development mode with pretty logging
ENVIRONMENT=development ./meta-code

# Run in production mode
ENVIRONMENT=production ./meta-code
```

### Verification

The server is running correctly when you see:
```
Starting Meta-MCP Server with handshake support...
Server configuration loaded
Starting stdio server
```

The server communicates using the MCP protocol over stdin/stdout and is designed to be used by MCP-compatible clients.
