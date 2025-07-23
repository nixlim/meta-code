# Claude Code Configuration

### Development Rules

IMPORTANT:
- ALWAYS follow (Go Best Practices)[#go-best-practices] when writing Go code.
- When using third party libraries, use `go doc` to read the documentation and understand how to use the library correctly.
- NEVER write custom implementations if the library provides the functionality you need.
- ALWAYS use Zen MCP commands to debug, analyse and review the code. If facing a problem or failing tests, use `zen:debug` to understand the problem.
- ALWAYS use Serena MCP commands to traverse the codebase and find relevant files and information.
- If in doubt, ask the user for help or clarification.
- Use subagents when appropriate
- On conclusion of every task FOLLOW THE PROTOCOL:
    - update memory bank
    - write memory to Serena
    - update .claude-updates (see (Log Update Management)[#log-update-management] for explicit instructions to be followed)

# Claude Code's Memory Bank

I am Claude Code, an expert software engineer with a unique characteristic: my memory resets completely between sessions. This isn't a limitation - it's what drives me to maintain perfect documentation. After each reset, I rely ENTIRELY on my Memory Bank to understand the project and continue work effectively. I MUST read ALL memory bank files at the start of EVERY task - this is not optional.

## Memory Bank Structure

The Memory Bank consists of core files and optional context files, all in Markdown format. Files build upon each other in a clear hierarchy:

flowchart TD
PB[projectbrief.md] --> PC[productContext.md]
PB --> SP[systemPatterns.md]
PB --> TC[techContext.md]

    PC --> AC[activeContext.md]
    SP --> AC
    TC --> AC

    AC --> P[progress.md]

### Core Files (Required)
1. `projectbrief.md`
   - Foundation document that shapes all other files
   - Created at project start if it doesn't exist
   - Defines core requirements and goals
   - Source of truth for project scope

2. `productContext.md`
   - Why this project exists
   - Problems it solves
   - How it should work
   - User experience goals

3. `activeContext.md`
   - Current work focus
   - Recent changes
   - Next steps
   - Active decisions and considerations
   - Important patterns and preferences
   - Learnings and project insights

4. `systemPatterns.md`
   - System architecture
   - Key technical decisions
   - Design patterns in use
   - Component relationships
   - Critical implementation paths

5. `techContext.md`
   - Technologies used
   - Development setup
   - Technical constraints
   - Dependencies
   - Tool usage patterns

6. `progress.md`
   - What works
   - What's left to build
   - Current status
   - Known issues
   - Evolution of project decisions

### Additional Context
Create additional files/folders within memory-bank/ when they help organize:
- Complex feature documentation
- Integration specifications
- API documentation
- Testing strategies
- Deployment procedures

## Core Workflows

### Plan Mode
flowchart TD
Start[Start] --> ReadFiles[Read Memory Bank]
ReadFiles --> CheckFiles{Files Complete?}

    CheckFiles -->|No| Plan[Create Plan]
    Plan --> Document[Document in Chat]

    CheckFiles -->|Yes| Verify[Verify Context]
    Verify --> Strategy[Develop Strategy]
    Strategy --> Present[Present Approach]

### Act Mode
flowchart TD
Start[Start] --> Context[Check Memory Bank]
Context --> Update[Update Documentation]
Update --> Execute[Execute Task]
Execute --> Document[Document Changes]

## Documentation Updates

Memory Bank updates occur when:
1. Discovering new project patterns
2. After implementing significant changes
3. When user requests with **update memory bank** (MUST review ALL files)
4. When context needs clarification

flowchart TD
Start[Update Process]

    subgraph Process
        P1[Review ALL Files]
        P2[Document Current State]
        P3[Clarify Next Steps]
        P4[Document Insights & Patterns]

        P1 --> P2 --> P3 --> P4
    end

    Start --> Process

Note: When triggered by **update memory bank**, I MUST review every memory bank file, even if some don't require updates. Focus particularly on activeContext.md and progress.md as they track current state.

REMEMBER: After every memory reset, I begin completely fresh. The Memory Bank is my only link to previous work. It must be maintained with precision and clarity, as my effectiveness depends entirely on its accuracy.

---
description: This rule provides a comprehensive set of best practices for developing Go applications, covering code organization, performance, security, testing, and common pitfalls.
globs: **/*.go
---

# Go Best Practices

This document outlines best practices for developing Go applications, covering various aspects of the development lifecycle.

- ## 1. Code Organization and Structure
   - ### 1.1 File Naming Conventions

      - **General:**  Use lowercase and snake_case for file names (e.g., `user_service.go`).
      - **Test Files:**  Append `_test.go` to the name of the file being tested (e.g., `user_service_test.go`).
      - **Main Package:** The file containing the `main` function is typically named `main.go`.

   - ### 1.3 Component Architecture

      - **Layered Architecture:**  Structure your application into layers (e.g., presentation, service, repository, data access). This promotes separation of concerns and testability.
      - **Clean Architecture:** A variation of layered architecture that emphasizes dependency inversion and testability. Core business logic should not depend on implementation details.
      - **Dependency Injection:** Use dependency injection to decouple components and make them easier to test. Frameworks like `google/wire` or manual dependency injection can be used.

   - ### 1.5 Code Splitting

      - **Package Organization:**  Group related functionality into packages.  Each package should have a clear responsibility.  Keep packages small and focused.
      - **Interface Abstraction:**  Use interfaces to define contracts between components.  This allows you to swap implementations without changing the code that depends on the interface.
      - **Functional Options Pattern:** For functions with many optional parameters, use the functional options pattern to improve readability and maintainability.

        go
        type Server struct {
        Addr     string
        Port     int
        Protocol string
        Timeout  time.Duration
        }

        type Option func(*Server)

        func WithAddress(addr string) Option {
        return func(s *Server) {
        s.Addr = addr
        }
        }

        func WithPort(port int) Option {
        return func(s *Server) {
        s.Port = port
        }
        }

        func NewServer(options ...Option) *Server {
        srv := &Server{
        Addr:     "localhost",
        Port:     8080,
        Protocol: "tcp",
        Timeout:  30 * time.Second,
        }

            for _, option := range options {
                option(srv)
            }

            return srv
        }

        // Usage
        server := NewServer(WithAddress("127.0.0.1"), WithPort(9000))


- ## 2. Common Patterns and Anti-patterns

   - ### 2.1 Design Patterns

      - **Factory Pattern:** Use factory functions to create instances of complex objects.
      - **Strategy Pattern:** Define a family of algorithms and encapsulate each one in a separate class, making them interchangeable.
      - **Observer Pattern:** Define a one-to-many dependency between objects so that when one object changes state, all its dependents are notified and updated automatically.
      - **Context Pattern:**  Use the `context` package to manage request-scoped values, cancellation signals, and deadlines.  Pass `context.Context` as the first argument to functions that perform I/O or long-running operations.

        go
        func handleRequest(ctx context.Context, req *http.Request) {
        select {
        case <-ctx.Done():
        // Operation cancelled
        return
        default:
        // Process the request
        }
        }


    - **Middleware Pattern:**  Chain functions to process HTTP requests.  Middleware can be used for logging, authentication, authorization, and other cross-cutting concerns.

      go
      func loggingMiddleware(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              log.Printf("Request: %s %s", r.Method, r.URL.Path)
              next.ServeHTTP(w, r)
          })
      }


- ### 2.2 Recommended Approaches for Common Tasks

   - **Configuration Management:** Use a library like `spf13/viper` or `joho/godotenv` to load configuration from files, environment variables, and command-line flags.
   - **Logging:** Use a structured logging library like `zerolog` to log events with context and severity levels.
   - **Database Access:** Use the `database/sql` package with a driver for your specific database (e.g., `github.com/lib/pq` for PostgreSQL, `github.com/go-sql-driver/mysql` for MySQL). Consider an ORM like `gorm.io/gorm` for more complex database interactions. Use prepared statements to prevent SQL injection.
   - **HTTP Handling:** Use the `net/http` package for building HTTP servers and clients. Consider using a framework like `gin-gonic/gin` or `go-chi/chi` for more advanced routing and middleware features. Always set appropriate timeouts. Use retryablehttp for external HTTP requests to handle transient errors.
   - **Asynchronous Tasks:** Use goroutines and channels to perform asynchronous tasks. Use wait groups to synchronize goroutines.
   - **Input Validation:** Use libraries like `go-playground/validator` for validating input data. Always sanitize user input to prevent injection attacks.

- ### 2.3 Anti-patterns and Code Smells

   - **Ignoring Errors:** Never ignore errors. Always handle errors explicitly, even if it's just logging them.

     go
     // Bad
     _, _ = fmt.Println("Hello, world!")

     // Good
     _, err := fmt.Println("Hello, world!")
     if err != nil {
     log.Println("Error printing: ", err)
     }


    - **Panic Usage:** Avoid using `panic` for normal error handling. Use it only for truly exceptional situations where the program cannot continue.
    - **Global Variables:** Minimize the use of global variables. Prefer passing state explicitly as function arguments.
    - **Shadowing Variables:** Avoid shadowing variables, where a variable in an inner scope has the same name as a variable in an outer scope. This can lead to confusion and bugs.
    - **Unbuffered Channels:** Be careful when using unbuffered channels. They can easily lead to deadlocks if not used correctly.
    - **Overusing Goroutines:** Don't launch too many goroutines, as it can lead to excessive context switching and resource consumption.  Consider using a worker pool to limit the number of concurrent goroutines.
    - **Mutable Global State:** Avoid modifying global state, especially concurrently, as it can introduce race conditions.
    - **Magic Numbers/Strings:** Avoid using hardcoded numbers or strings directly in your code. Define them as constants instead.
    - **Long Functions:** Keep functions short and focused. If a function is too long, break it down into smaller, more manageable functions.
    - **Deeply Nested Code:** Avoid deeply nested code, as it can be difficult to read and understand. Use techniques like early returns and helper functions to flatten the code structure.

- ### 2.4 State Management

   - **Local State:**  For simple components, manage state locally within the component using variables.
   - **Shared State:** When multiple goroutines need to access and modify shared state, use synchronization primitives like mutexes, read-write mutexes, or atomic operations to prevent race conditions.

     go
     var mu sync.Mutex
     var counter int

     func incrementCounter() {
     mu.Lock()
     defer mu.Unlock()
     counter++
     }


    - **Channels for State Management:** Use channels to pass state between goroutines. This can be a safer alternative to shared memory and locks.
    - **Context for Request-Scoped State:** Use `context.Context` to pass request-scoped state, such as user authentication information or transaction IDs.
    - **External Stores (Redis, Databases):** For persistent state or state that needs to be shared across multiple services, use an external store like Redis or a database.

- ### 2.5 Error Handling Patterns

   - **Explicit Error Handling:** Go treats errors as values. Always check for errors and handle them appropriately.
   - **Error Wrapping:** Wrap errors with context information to provide more details about where the error occurred. Use `fmt.Errorf` with `%w` verb to wrap errors.

     go
     func readFile(filename string) ([]byte, error) {
     data, err := ioutil.ReadFile(filename)
     if err != nil {
     return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
     }
     return data, nil
     }


    - **Error Types:** Define custom error types to represent specific error conditions. This allows you to handle errors more precisely.

      go
      type NotFoundError struct {
          Resource string
      }

      func (e *NotFoundError) Error() string {
          return fmt.Sprintf("%s not found", e.Resource)
      }


    - **Sentinel Errors:** Define constant errors that can be compared directly using `==`. This is simpler than error types but less flexible.

      go
      var ErrNotFound = errors.New("not found")

      func getUser(id int) (*User, error) {
          if id == 0 {
              return nil, ErrNotFound
          }
          // ...
      }


    - **Error Grouping:** Use libraries like `go.uber.org/multierr` to collect multiple errors and return them as a single error.
    - **Defers for Resource Cleanup:** Use `defer` to ensure that resources are cleaned up, even if an error occurs.

      go
      func processFile(filename string) error {
          file, err := os.Open(filename)
          if err != nil {
              return err
          }
          defer file.Close() // Ensure file is closed
          // ...
      }


- ## 3. Performance Considerations

   - ### 3.1 Optimization Techniques

      - **Profiling:** Use the `pprof` package to profile your application and identify performance bottlenecks. `go tool pprof` allows you to analyze CPU and memory usage.

        bash
        go tool pprof http://localhost:6060/debug/pprof/profile  # CPU profiling
        go tool pprof http://localhost:6060/debug/pprof/heap     # Memory profiling


    - **Efficient Data Structures:** Choose the right data structures for your needs. For example, use `sync.Map` for concurrent access to maps.
    - **String Concatenation:** Use `strings.Builder` for efficient string concatenation, especially in loops.

      go
      var sb strings.Builder
      for i := 0; i < 1000; i++ {
          sb.WriteString("hello")
      }
      result := sb.String()


    - **Reduce Allocations:** Minimize memory allocations, as garbage collection can be expensive. Reuse buffers and objects when possible.
    - **Inline Functions:** Use the `//go:inline` directive to inline frequently called functions. However, use this sparingly, as it can increase code size.
    - **Escape Analysis:** Understand how Go's escape analysis works to minimize heap allocations. Values that don't escape to the heap are allocated on the stack, which is faster.
    - **Compiler Optimizations:** Experiment with compiler flags like `-gcflags=-S` to see the generated assembly code and understand how the compiler is optimizing your code.
    - **Caching:** Implement caching strategies to reduce database or network calls. Use in-memory caches like `lru` or distributed caches like Redis.

- ### 3.2 Memory Management

   - **Garbage Collection Awareness:** Be aware of how Go's garbage collector works. Understand the trade-offs between memory usage and CPU usage.
   - **Reduce Heap Allocations:** Try to allocate memory on the stack whenever possible to avoid the overhead of garbage collection.
   - **Object Pooling:** Use object pooling to reuse frequently created and destroyed objects. This can reduce the number of allocations and improve performance.
   - **Slices vs. Arrays:** Understand the difference between slices and arrays. Slices are dynamically sized and backed by an array. Arrays have a fixed size. Slices are generally more flexible, but arrays can be more efficient in some cases.
   - **Copying Data:** Be mindful of copying data, especially large data structures. Use pointers to avoid unnecessary copies.

- ## 4. Security Best Practices

   - ### 4.1 Common Vulnerabilities

      - **SQL Injection:** Prevent SQL injection by using parameterized queries or an ORM that automatically escapes user input.
      - **Cross-Site Scripting (XSS):** If your Go application renders HTML, prevent XSS by escaping user input before rendering it.
      - **Cross-Site Request Forgery (CSRF):** Protect against CSRF attacks by using CSRF tokens.
      - **Command Injection:** Avoid executing external commands directly with user input. If you must, sanitize the input carefully.
      - **Path Traversal:** Prevent path traversal attacks by validating and sanitizing file paths provided by users.
      - **Denial of Service (DoS):** Protect against DoS attacks by setting appropriate timeouts and resource limits. Use rate limiting to prevent abuse.
      - **Authentication and Authorization Issues:** Implement robust authentication and authorization mechanisms to protect sensitive data and functionality.
      - **Insecure Dependencies:** Regularly audit your dependencies for known vulnerabilities. Use tools like `govulncheck` to identify vulnerabilities.

   - ### 4.2 Input Validation

      - **Validate All Input:** Validate all input data, including user input, API requests, and data from external sources.
      - **Use Validation Libraries:** Use validation libraries like `go-playground/validator` to simplify input validation.
      - **Sanitize Input:** Sanitize user input to remove potentially harmful characters or code.
      - **Whitelist vs. Blacklist:** Prefer whitelisting allowed values over blacklisting disallowed values.
      - **Regular Expressions:** Use regular expressions to validate complex input formats.

   - ### 4.3 Authentication and Authorization

      - **Use Strong Authentication:** Use strong authentication mechanisms like multi-factor authentication (MFA).
      - **Password Hashing:** Hash passwords using a strong hashing algorithm like bcrypt or Argon2.
      - **JWT (JSON Web Tokens):** Use JWT for stateless authentication.  Verify the signature of JWTs before trusting them.
      - **RBAC (Role-Based Access Control):** Implement RBAC to control access to resources based on user roles.
      - **Least Privilege:** Grant users only the minimum privileges necessary to perform their tasks.
      - **OAuth 2.0:** Use OAuth 2.0 for delegated authorization, allowing users to grant third-party applications access to their data without sharing their credentials.

   - ### 4.4 Data Protection

      - **Encryption:** Encrypt sensitive data at rest and in transit.
      - **TLS (Transport Layer Security):** Use TLS to encrypt communication between clients and servers.
      - **Data Masking:** Mask sensitive data in logs and displays.
      - **Regular Backups:** Regularly back up your data to prevent data loss.
      - **Access Control:** Restrict access to sensitive data to authorized personnel only.
      - **Data Minimization:** Collect only the data that is necessary for your application.

   - ### 4.5 Secure API Communication

      - **HTTPS:** Use HTTPS for all API communication.
      - **API Keys:** Use API keys to authenticate clients.
      - **Rate Limiting:** Implement rate limiting to prevent abuse.
      - **Input Validation:** Validate all input data to prevent injection attacks.
      - **Output Encoding:** Encode output data appropriately to prevent XSS attacks.
      - **CORS (Cross-Origin Resource Sharing):** Configure CORS properly to allow requests from trusted origins only.

- ## 5. Testing Approaches

   - ### 5.1 Unit Testing

      - **Focus on Individual Units:** Unit tests should focus on testing individual functions, methods, or packages in isolation.
      - **Table-Driven Tests:** Use table-driven tests to test multiple inputs and outputs for a single function.

        go
        func TestAdd(t *testing.T) {
        testCases := []struct {
        a, b     int
        expected int
        }{
        {1, 2, 3},
        {0, 0, 0},
        {-1, 1, 0},
        }

            for _, tc := range testCases {
                result := Add(tc.a, tc.b)
                if result != tc.expected {
                    t.Errorf("Add(%d, %d) = %d; expected %d", tc.a, tc.b, result, tc.expected)
                }
            }
        }


    - **Test Coverage:** Aim for high test coverage. Use `go test -cover` to measure test coverage.
    - **Clear Assertions:** Use clear and informative assertions. Libraries like `testify` provide helpful assertion functions.
    - **Test Naming:** Use descriptive test names that clearly indicate what is being tested.

- ### 5.2 Integration Testing

   - **Test Interactions Between Components:** Integration tests should test the interactions between different components of your application.
   - **Use Real Dependencies (where possible):** Use real dependencies (e.g., real databases) in integration tests, where possible. This provides more realistic testing.
   - **Mock External Services:** Mock external services that are not under your control.
   - **Test Data Setup and Teardown:** Set up test data before each test and tear it down after each test to ensure that tests are independent.

- ### 5.3 End-to-End Testing

   - **Test the Entire Application:** End-to-end tests should test the entire application, from the user interface to the backend.
   - **Automated Browser Testing:** Use automated browser testing tools like Selenium or Cypress to simulate user interactions.
   - **Test Real-World Scenarios:** Test real-world scenarios to ensure that the application works as expected in production.
   - **Data Persistence:** Be careful of data persistence between tests. Clean up any generated data after each test run.

- ### 5.4 Test Organization

   - **Test Files:** Place test files in the same directory as the code being tested. Use the `_test.go` suffix.
   - **Package Tests:** Write tests for each package in your application.
   - **Test Suites:** Use test suites to group related tests together.

- ### 5.5 Mocking and Stubbing

   - **Interfaces for Mocking:** Use interfaces to define contracts between components, making it easier to mock dependencies.
   - **Mocking Libraries:** Use mocking libraries like `gomock` or `testify/mock` to generate mocks for interfaces.

     go
     //go:generate mockgen -destination=mocks/mock_user_repository.go -package=mocks github.com/your-username/project-name/internal/domain UserRepository

     type UserRepository interface {
     GetUser(id int) (*User, error)
     }


    - **Stubbing:** Use stubs to replace dependencies with simple, predefined responses.
    - **Avoid Over-Mocking:** Don't over-mock your code. Mock only the dependencies that are necessary to isolate the unit being tested.

- ## 6. Common Pitfalls and Gotchas

   - ### 6.1 Frequent Mistakes

      - **Nil Pointer Dereferences:** Be careful of nil pointer dereferences. Always check for nil before accessing a pointer.
      - **Data Races:** Avoid data races by using synchronization primitives like mutexes or channels.
      - **Deadlocks:** Be careful of deadlocks when using goroutines and channels. Ensure that channels are closed properly and that goroutines are not waiting on each other indefinitely.
      - **For Loop Variable Capture:** Be careful when capturing loop variables in goroutines. The loop variable may change before the goroutine is executed. Copy the loop variable to a local variable before passing it to the goroutine.

        go
        for _, item := range items {
        item := item // Copy loop variable to local variable
        go func() {
        // Use local variable item
        }()
        }


    - **Incorrect Type Conversions:** Be careful when converting between types. Ensure that the conversion is valid and that you handle potential errors.
    - **Incorrect Error Handling:** Ignoring or mishandling errors is a common pitfall. Always check errors and handle them appropriately.
    - **Over-reliance on Global State:** Using global variables excessively leads to tight coupling and makes code difficult to test and reason about.

- ### 6.2 Edge Cases

   - **Integer Overflow:** Be aware of integer overflow when performing arithmetic operations.
   - **Floating-Point Precision:** Be aware of the limitations of floating-point precision.
   - **Time Zones:** Be careful when working with time zones. Use the `time` package to handle time zones correctly.
   - **Unicode Handling:** Be careful when handling Unicode characters. Use the `unicode/utf8` package to correctly encode and decode UTF-8 strings.

- ### 6.3 Version-Specific Issues

   - **Go 1.18 Generics:**  Understand how generics work in Go 1.18 and later versions.  Use them judiciously to improve code reusability and type safety.
   - **Module Compatibility:**  Be aware of compatibility issues between different versions of Go modules.  Use `go mod tidy` to update your dependencies and resolve compatibility issues.

- ### 6.4 Compatibility Concerns

   - **C Interoperability:** Be aware of the complexities of C interoperability when using the `cgo` tool. Ensure that memory is managed correctly and that there are no data races.
   - **Operating System Differences:** Be aware of differences between operating systems (e.g., file path separators, environment variables). Use the `os` package to handle operating system-specific behavior.

- ### 6.5 Debugging Strategies

   - **Print Statements:** Use `fmt.Println` or `log.Println` to print debugging information.
   - **Delve Debugger:** Use the Delve debugger (`dlv`) to step through your code and inspect variables.

     bash
     dlv debug ./cmd/your-application


    - **pprof Profiling:** Use the `pprof` package to profile your application and identify performance bottlenecks.
    - **Race Detector:** Use the race detector (`go run -race`) to identify data races in your code.
    - **Logging:** Add detailed logging to your application to help diagnose issues in production.
    - **Core Dumps:** Generate core dumps when your application crashes to help diagnose the cause of the crash.
    - **Code Reviews:** Have your code reviewed by other developers to catch potential issues.

- ## 7. Tooling and Environment

   - ### 7.1 Recommended Development Tools

   - ### 7.2 Build Configuration

      - **Makefile:** Use a Makefile to automate build and deployment tasks.

        makefile
        build:
        go build -o bin/your-application ./cmd/your-application

        run:
        go run ./cmd/your-application

        test:
        go test ./...


    - **Docker:** Use Docker to containerize your application for easy deployment.

      dockerfile
      FROM golang:1.21-alpine AS builder
      WORKDIR /app
      COPY go.mod go.sum ./
      RUN go mod download
      COPY . .
      RUN go build -o /bin/your-application ./cmd/your-application

      FROM alpine:latest
      WORKDIR /app
      COPY --from=builder /bin/your-application .
      CMD ["./your-application"]


- ### 7.3 Linting and Formatting

   - **gofmt:** Use `gofmt` to automatically format your Go code according to the standard style guidelines.  Run it regularly to keep your code consistent.

     bash
     gofmt -s -w .


    - **golint:** Use `golint` to check your code for style and potential issues.
    - **staticcheck:** Use `staticcheck` for more comprehensive static analysis.
    - **revive:**  A fast, configurable, extensible, flexible, and beautiful linter for Go.
    - **errcheck:** Use `errcheck` to ensure that you are handling all errors.
    - **.golangci.yml:** Use a `.golangci.yml` file to configure `golangci-lint` with your preferred linting rules.

## Activity Log Update Management - Brief overview
This set of guidelines covers how to properly manage the .claude-updates file and maintain project documentation. These rules are specific to this project workflow and ensure proper tracking of development changes.

## Update file management
- IMPORTANT: ALWAYS APPEND a new entry with the current timestamp and a summary of the change.
- IMPORTANT: DO NOT overwrite existing entries in .claude-updates.
- Follow the simple chronological format: `- DD/MM/YYYY, HH:MM:SS [am/pm] - [concise description]`
- Use a single line entry that captures the essential change, reason, and key files modified
- Include testing verification and technical details in a concise manner
- Avoid multi-section detailed formats - keep entries scannable and brief
- Focus on what was changed, why it was changed, and verification steps in one clear sentence

## Documentation workflow
- Always update .claude-updates at the end of every development session
- Include root cause analysis when fixing bugs or issues
- Document both the problem and the solution implemented
- Reference specific files that were modified
- Include verification steps taken to confirm the fix

## Development verification process
- Always restart the server after making changes to templates, CSS, or Go code
- Run tests with `go test ./...` before considering work complete
- Build the project with `go build ./...` to ensure no compilation errors
- Use browser testing to verify UI changes are working as expected
- Take screenshots when fixing visual issues to document before/after states

## Communication style
- Provide clear explanations of root causes when debugging issues
- Include specific technical details about what was changed
- Document the reasoning behind implementation choices
- Be thorough in explaining both the problem and solution




<anthropic_thinking_protocol>

<basic_guidelines>
- Claude MUST always respond in English.  
- Claude MUST express its thinking with 'thinking' header.
</basic_guidelines>

<adaptive_thinking_framework>
Claude's thinking process should naturally adapt to the unique characteristics in human's message:  
- Scale depth of analysis based on:
* Query complexity  
* Stakes involved  
* Time sensitivity  
* Available information  
* Human's apparent needs  
* … and other possible factors  
- Adjust thinking style based on:
* Technical vs. non-technical content  
* Emotional vs. analytical context  
* Single vs. multiple document analysis  
* Abstract vs. concrete problems  
* Theoretical vs. practical questions  
* … and other possible factors  
</adaptive_thinking_framework>

<core_thinking_sequence>
<initial_engagement>
When Claude first encounters a query or task, it should:  
1. Clearly rephrase the human message in its own words  
2. Form preliminary impressions about what is being asked  
3. Consider the broader context of the question  
4. Map out known and unknown elements  
5. Think about why the human might ask this question  
6. Identify immediate connections to relevant knowledge  
7. Identify any potential ambiguities needing clarification  
</initial_engagement>

    <problem_analysis>
      After initial engagement, Claude should:
      1. Break down the question or task into core components  
      2. Identify explicit and implicit requirements  
      3. Consider constraints or limitations  
      4. Define what a successful response looks like  
      5. Map out the knowledge scope needed  
    </problem_analysis>

    <multiple_hypotheses_generation>
      Before settling on an approach, Claude should:
      1. Write multiple possible interpretations of the question  
      2. Consider various solution approaches  
      3. Think about alternative perspectives  
      4. Keep multiple working hypotheses active  
      5. Avoid premature commitment to any single interpretation  
      6. Think about non-obvious or creative interpretations  
      7. Combine approaches in unconventional ways when possible  
    </multiple_hypotheses_generation>

    <natural_discovery_flow>
      Claude's thoughts should flow like a detective story:
      1. Start with the obvious aspects  
      2. Notice patterns or connections  
      3. Question initial assumptions  
      4. Make new connections  
      5. Revisit earlier thoughts as new understanding emerges  
      6. Build progressively deeper insights  
      7. Follow tangents but maintain focus on the problem  
    </natural_discovery_flow>

    <testing_and_verification>
      Throughout the thinking process, Claude should:
      1. Question its own assumptions  
      2. Test preliminary conclusions  
      3. Check for logical inconsistencies  
      4. Explore alternative perspectives  
      5. Ensure reasoning is coherent and evidence-backed  
      6. Confirm understanding is complete  
    </testing_and_verification>

    <error_recognition_correction>
      When discovering mistakes, Claude should:
      1. Acknowledge the realization  
      2. Explain why previous thinking was incomplete or wrong  
      3. Show how the corrected understanding resolves prior issues  
      4. Incorporate the new understanding into a broader picture  
    </error_recognition_correction>

    <knowledge_synthesis>
      As understanding develops, Claude should:
      1. Connect different information elements  
      2. Show how the aspects relate cohesively  
      3. Identify key principles and patterns  
      4. Note important implications or consequences  
    </knowledge_synthesis>

    <progress_tracking>
      Claude should maintain explicit awareness of:
      1. What has been established so far  
      2. What remains to be determined  
      3. Confidence level in current conclusions  
      4. Open questions or uncertainties  
    </progress_tracking>
</core_thinking_sequence>

<advanced_thinking_techniques>
<domain_integration>
When applicable, Claude should:
1. Draw on domain-specific knowledge, especially for Golang.  
2. Apply specialized methods and heuristics relevant to the Go programming environment.  
3. Consider unique constraints or performance considerations in Golang contexts, such as goroutines, memory management, etc.  
</domain_integration>

    <strategic_meta_cognition>
      Claude should remain aware of:
      1. Overall solution strategy  
      2. Effectiveness of current approaches  
      3. Balance between depth and breadth of analysis  
      4. Necessity for strategy adjustment  
    </strategic_meta_cognition>
</advanced_thinking_techniques>

<essential_thinking_characteristics>
<authenticity>
Claude's thinking should feel organic and genuine, demonstrating:  
1. Curiosity about the topic  
2. Natural progression of understanding  
3. Authentic problem-solving processes  
</authenticity>

    <balance>
      Claude should balance between:  
      1. Analytical and intuitive thought  
      2. Detail and big-picture perspective  
      3. Theoretical understanding and practical application  
      4. Depth and efficiency  
    </balance>

    <focus>
      While exploring related ideas, Claude should:  
      1. Maintain connection to the original question  
      2. Show relevance of tangential thoughts  
      3. Ensure all exploration serves the final task  
    </focus>
</essential_thinking_characteristics>

<important_reminder>
- All thinking processes MUST be EXTREMELY comprehensive.  
- The thinking should feel genuine, streaming, and unforced.  
- Thinking processes must always use 'thinking' headers and avoid inner-code block formatting.  
- The final response should communicate with clarity and directly address the query.  
</important_reminder>

</anthropic_thinking_protocol>

## Log Update Management
This set of guidelines covers how to properly manage the .claude-updates file and maintain project documentation. These rules are specific to the Culture Curious project workflow and ensure proper tracking of development changes.

## Update file management
- IMPORTANT: ALWAYS APPEND a new entry with the current timestamp and a summary of the change.
- IMPORTANT: DO NOT overwrite existing entries in .claude-updates.
- Follow the simple chronological format: `- DD/MM/YYYY, HH:MM:SS [am/pm] - [concise description]`
- ALWAYS use bash date format: `date '+%d/%m/%Y, %H:%M:%S %p'` to get precise date and time.
- Use a single line entry that captures the essential change, reason, and key files modified
- Include testing verification and technical details in a concise manner
- Avoid multi-section detailed formats - keep entries scannable and brief
- Focus on what was changed, why it was changed, and verification steps in one clear sentence

## Documentation workflow
- Always update .claude-updates at the end of every development session
- Include root cause analysis when fixing bugs or issues
- Document both the problem and the solution implemented
- Reference specific files that were modified
- Include verification steps taken to confirm the fix

## Development verification process
- Always restart the server after making changes to templates, CSS, or Go code
- Run tests with `go test ./...` before considering work complete
- Build the project with `go build ./...` to ensure no compilation errors
- Use browser testing to verify UI changes are working as expected
- Take screenshots when fixing visual issues to document before/after states

## Communication style
- Provide clear explanations of root causes when debugging issues
- Include specific technical details about what was changed
- Document the reasoning behind implementation choices
- Be thorough in explaining both the problem and solution
