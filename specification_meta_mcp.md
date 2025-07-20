### System Overview

The proposed meta-MCP (Model Context Protocol) server is designed as an orchestrator that aggregates and manages multiple underlying MCP servers. It adheres to the MCP specification (an open JSON-RPC 2.0-based protocol for connecting LLMs to data sources and tools, as introduced by Anthropic). The meta-MCP acts as both an MCP client (to interact with connected MCP servers) and an MCP server (exposing an aggregated interface to external LLM hosts or applications). This enables intelligent, composable workflows for tasks like system design, code implementation, feature development, and refactoring.

Key features:
- **Connection to Multiple MCP Servers**: Connects to local MCP servers running on the user's machine (e.g., one for file system access, another for Git repositories, a third for database queries). Connections are configured via a JSON file in a known location (e.g., `~/.meta-mcp/config.json`), which users can amend to add servers.
- **Command Cataloging**: Dynamically builds and maintains a catalog of all available commands (e.g., resources, prompts, tools) from connected servers.
- **AI-Driven Suggestions**: Leverages external AI APIs (e.g., OpenAI's GPT models) to analyze user tasks and suggest optimal command sequences or workflows.
- **Hybrid Workflow Support**: Combines automatic AI suggestions with predefined rule-based workflows for deterministic tasks.
- **Task Focus**: Optimized for development workflows, such as generating system architectures, implementing features via code generation/execution, and refactoring code with context-aware changes.

The system follows the specified principles:
- **Local-First**: All components run locally on the user's machine as host processes, with no external dependencies beyond AI API calls. The meta-MCP server runs as a process on the host, started by an AI Coding Agent such as Claude Code (Anthropic's agentic coding tool that integrates into terminals for codebase-aware tasks, file editing, and command execution).
- **AI-Assisted**: External AI APIs handle intelligent suggestions; no local ML models to keep it lightweight.
- **Stateless Core**: The meta-MCP core processes requests without maintaining session state; workflow state (e.g., task progress, command history) is persisted to a local database.
- **Modular**: Components are designed as independent modules within the Go binary, each testable in isolation (e.g., via unit tests or integration mocks).
- **Secure**: API keys/credentials for external AI and MCP servers are loaded into memory at runtime (e.g., via environment variables or command-line flags) and never persisted; all communications use TLS even locally; user consent prompts for sensitive commands (e.g., file writes).

### Architecture Adjustments for Host-Based Execution

To accommodate the meta-MCP running as a host process started by an AI Coding Agent (e.g., Claude Code):
- **Startup Mechanism**: The meta-MCP is a standalone Go binary executed via command line (e.g., `./meta-mcp --port 8000 --config ~/.meta-mcp/config.json`). The AI Coding Agent (like Claude Code) can initiate this process programmatically (e.g., via shell commands in the terminal), allowing seamless integration into coding workflows. Once started, it listens for MCP requests and manages connections.
- **Configuration via JSON**: The meta-MCP reads a config file that lists MCP servers. Example JSON structure:
  ```json
  {
    "mcp_servers": [
      {
        "name": "file_mcp",
        "type": "stdio",
        "command": "/path/to/file-mcp-server",
        "args": ["--config", "/path/to/config"]
      },
      {
        "name": "git_mcp",
        "type": "url",
        "url": "http://localhost:9001/sse",
        "headers": {"Authorization": "Bearer token"}
      }
    ],
    "ai_api": {
      "provider": "openai",
      "endpoint": "https://api.openai.com/v1"
    },
    "storage_path": "~/.meta-mcp/storage.db"
  }
  ```
  Users amend this file to add servers, and meta-MCP reloads it dynamically (e.g., on startup or via signal handler like SIGHUP).
- **Connection Methods**: MCP supports STDIO (for local, spawned servers) and HTTP/SSE (for network-exposed servers). Meta-MCP handles both:
  - **STDIO**: For servers configured as STDIO, meta-MCP spawns subprocesses on the host using Go's `os/exec` package, managing stdin/stdout pipes for JSON-RPC communication.
  - **HTTP/SSE**: Direct HTTP client connections for URL-based servers.
- **Integration with AI Coding Agent**: Since started by agents like Claude Code, the meta-MCP can expose endpoints or logs that the agent monitors. For example, Claude Code could start meta-MCP, then use it as an MCP server for delegated tasks (e.g., querying codebases via aggregated commands).

The meta-MCP core routes requests intelligently, using the catalog to proxy or aggregate responses from underlying servers.

#### High-Level Components

| Component | Description | Technology Stack | Execution | Responsibilities |
|-----------|-------------|------------------|-----------|------------------|
| **Meta-MCP Core** | Central stateless server that implements the MCP protocol, catalogs commands, and orchestrates workflows. | Go (with JSON-RPC libraries like `gorilla/rpc` or custom implementation; MCP spec alignment via Go structs for messages). Uses Gin or Echo for HTTP handling if needed. | Host process (binary). | - Handle incoming MCP requests from LLM hosts.<br>- Load config JSON and connect to listed MCP servers.<br>- Catalog commands from connected MCP servers.<br>- Invoke AI for suggestions or apply rule-based workflows.<br>- Proxy/aggregate responses.<br>- Persist workflow state to storage. |
| **MCP Server Connectors** | Adapters for each connected MCP server. Each connector is a lightweight proxy that handles authentication and capability querying. | Same as core (Go). | Integrated into core binary (spawn subprocesses for STDIO). | - Establish connections based on config (spawn for STDIO, HTTP client for URL).<br>- Query and cache capabilities (e.g., available tools like `read_file`, `execute_tool`).<br>- Route specific commands to the target server. |
| **AI Integrator** | Module for calling external AI APIs to generate command suggestions. | Integrated into core; uses libraries like `go-openai` for OpenAI integration. | Integrated into core binary. | - Format task descriptions + catalog into prompts for AI.<br>- Parse AI responses into executable command sequences.<br>- Handle retries and error mapping. |
| **Workflow Engine** | Handles execution of AI-suggested or rule-based workflows. Supports chaining commands across servers. | Integrated into core; uses a simple state machine (e.g., via Go channels for async handling, but synchronous for statelessness). | Integrated into core binary. | - Execute sequences (e.g., read context from Server A, process with AI, write to Server B).<br>- Support rule-based templates (e.g., YAML-defined workflows).<br>- Monitor progress and cancel operations. |
| **Local Storage** | Persists workflow state, command catalog, and logs. | BoltDB, Redis or SQLite with Go drivers (e.g., `go.etcd.io/bbolt`). | Embedded file on host (e.g., `~/.meta-mcp/storage.db`). | - Store serialized workflow states (e.g., JSON blobs with task ID, steps, results).<br>- Cache command catalog to avoid frequent queries.<br>- No sensitive data (e.g., credentials). |
| **Underlying MCP Servers** | User-provided servers for dev tasks (extendable). | Implemented in any language (e.g., Python/TypeScript with official MCP SDKs); run directly on host. | Separate host processes (spawned or manual). | - Expose domain-specific commands (e.g., `list_files`, `read_file` for file access; `commit_changes` for Git).<br>- Configured to use HTTP/SSE or STDIO based on user setup. |

#### Data Flow
1. **Initialization**: The AI Coding Agent (e.g., Claude Code) starts the meta-MCP binary. The core loads the config JSON, establishes connections (spawn for STDIO, HTTP for URL), and queries capabilities using MCP's negotiation methods (e.g., `initialize` request).
2. **Catalog Building**: Aggregates features into a unified catalog (e.g., a JSON structure: `{ "server_id": "file_mcp", "commands": ["read_file", "write_file"], "descriptions": [...] }`). Stored in local storage.
3. **Task Reception**: An external LLM host (or the AI agent itself) sends a request to meta-MCP (as an MCP server) with a task (e.g., "Refactor the function in src/main.py to improve efficiency").
4. **Suggestion Generation**:
   - **AI-Driven**: Send prompt to external AI: "Given task: [task]. Available commands: [catalog]. Suggest a sequence of commands and servers to use."
   - **Rule-Based**: Match task to predefined rules (e.g., if "refactor" keyword, use workflow: read_file → analyze_code → write_file).
   - Hybrid: Use AI if no rule matches, or refine rules with AI.
5. **Workflow Execution**: Execute the sequence statelessly (e.g., proxy each command to the target MCP server via appropriate transport). Persist intermediate state (e.g., results) to storage.
6. **Response**: Aggregate results (e.g., updated code context) and return via MCP response format.
7. **Error/Security Handling**: If a command requires consent (e.g., write access), prompt user via client feature (e.g., `elicitation`). Errors are mapped to MCP error reports.

#### Protocol Extensions for Meta-MCP
- **Custom Methods**: Extend MCP with meta-commands like `get_catalog` (returns aggregated capabilities) and `suggest_workflow` (invokes AI/rule engine).
- **Message Formats**: Adhere to JSON-RPC 2.0. Example request/response unchanged from prior design.

### Implementation Details

#### Setup and Deployment
- **Binary Build and Run**: Compile the Go project to a binary (e.g., `go build -o meta-mcp`). Start via AI agent: `exec.Command("./meta-mcp", "--port", "8000")`.
- **Environment Variables**: For credentials (e.g., `export OPENAI_API_KEY=sk-...`), loaded at runtime.
- **Scalability**: Connections scale with Go's concurrency (goroutines for each connector and subprocess).
- **Testing**: Unit tests for core using Go's testing package (e.g., mock connections with `net/http/httptest` or `os/exec` mocks); integration tests simulate config JSON and dummy servers.

#### Feature Development Workflow Example
For a task like "Develop a new API endpoint":
1. Rule-based: Predefined workflow reads project files (FileMCP), generates code skeleton (CodeExecMCP), commits changes (GitMCP).
2. AI-Driven: AI suggests: "Use FileMCP.read_file on routes.py, then CodeExecMCP.execute_tool to add endpoint, finally GitMCP.commit."
3. Execution: Meta-MCP chains calls, persists state (e.g., generated code), and returns context to the LLM host.

#### Security Implementation
- **Credentials**: Loaded from env vars or flags at runtime; never written to disk.
- **Consent**: For tools/resources, implement MCP's `elicitation` to prompt users (e.g., "Approve file write?").
- **Auditing**: Log all commands to storage (anonymized).
- **Threat Mitigation**: Validate all inputs; sandbox spawned processes (e.g., via `syscall` limits in Go).

### Addressing the STDIO Connection Question
Most MCP servers prefer STDIO for local integrations, as it's simple and avoids network overhead. URL (HTTP/SSE) is used for remote or persistent servers.

Meta-MCP **can connect to several servers at once via STDIO**. In Go, this is achieved by spawning multiple subprocesses (using `os/exec.Cmd`)—one per STDIO-configured server—and managing their stdin/stdout pipes concurrently with goroutines. Each connector maintains a dedicated pipe for JSON-RPC messages, allowing parallel interactions without blocking. Since the meta-MCP runs directly on the host (started by the AI Coding Agent), spawning and managing these subprocesses is straightforward, with no container boundaries to bridge.

Limitations and Mitigations: Spawning multiple subprocesses increases resource use; implement timeouts and resource limits (e.g., via Go's `context` package). If a server is already running (not spawnable), users should configure it as URL in the JSON or log an error for fallback.

This design provides a flexible, secure foundation tailored for integration with AI Coding Agents like Claude Code. For prototyping, implement connectors with `os/exec` for STDIO and `net/http` for URL. Expand by adding support for OAuth in config (per MCP spec). If needed, reference the MCP spec for transport details.