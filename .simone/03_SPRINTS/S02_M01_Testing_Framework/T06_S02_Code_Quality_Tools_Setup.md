# Task: Code Quality Tools Setup (T06_S02)

## Summary
Configure comprehensive code quality enforcement tools including golangci-lint configuration, pre-commit hooks, code formatting standards, and integrate quality checks into the development workflow to ensure consistent code quality across the codebase.

## Objective
Establish automated code quality enforcement mechanisms that catch issues early, enforce consistent coding standards, and integrate seamlessly with the development workflow. This includes configuring golangci-lint with project-specific rules, setting up pre-commit hooks for automatic validation, and creating code quality monitoring dashboards.

## Acceptance Criteria
- [ ] Create comprehensive `.golangci.yml` configuration with project-specific linting rules
- [ ] Set up pre-commit hooks for automatic code quality checks
- [ ] Configure go fmt, go vet, and additional quality tools
- [ ] Integrate linting with Makefile targets from T01_S02
- [ ] Create custom linting rules for project-specific patterns
- [ ] Set up code complexity monitoring and reporting
- [ ] Establish security vulnerability scanning with govulncheck
- [ ] Document code quality standards and enforcement policies
- [ ] Create CI/CD templates for automated quality checks

## Technical Guidance

### Current State Analysis
Based on project analysis:
- golangci-lint is already installed at `/usr/local/bin/golangci-lint` (per CLAUDE.md)
- No `.golangci.yml` configuration file exists yet
- No pre-commit hooks are currently configured (only sample hooks exist)
- Code formatting commands are documented but not automated
- `.claude/default-rules.md` references golangci-lint configuration

### Required Implementation

#### 1. Comprehensive golangci-lint Configuration

Create `.golangci.yml` at project root:

```yaml
# golangci-lint configuration for Meta-MCP Server
run:
  # Timeout for analysis
  timeout: 5m
  
  # Include test files
  tests: true
  
  # Skip vendor, third party, etc.
  skip-dirs:
    - vendor
    - third_party
    - testdata
    - examples
    
  # Skip generated files
  skip-files:
    - ".*\\.pb\\.go$"
    - ".*\\.gen\\.go$"
    - "mock_.*\\.go$"

# Output configuration
output:
  # Format of output: colored-line-number|line-number|json|tab|checkstyle|code-climate
  format: colored-line-number
  
  # Print lines of code with issue
  print-issued-lines: true
  
  # Print linter name in the end of issue text
  print-linter-name: true
  
  # Make output unique by line
  uniq-by-line: true

# Linter settings
linters-settings:
  # Error checking
  errcheck:
    # Report about not checking of errors in type assertions
    check-type-assertions: true
    
    # Report about assignment of errors to blank identifier
    check-blank: true
    
    # Exclude certain error return patterns
    exclude-functions:
      - io/ioutil.ReadFile
      - io.Copy(*bytes.Buffer)
      - io.Copy(os.Stdout)

  # Go fmt
  gofmt:
    # Simplify code
    simplify: true

  # Go imports
  goimports:
    # Put local imports after 3rd party
    local-prefixes: github.com/nixlim/meta-mcp-server

  # Cyclomatic complexity
  gocyclo:
    # Minimal code complexity to report
    min-complexity: 15

  # Cognitive complexity
  gocognit:
    # Minimal cognitive complexity to report
    min-complexity: 20

  # Line length
  lll:
    # Max line length
    line-length: 120
    # Tab width in spaces
    tab-width: 1

  # Magic numbers
  gomnd:
    settings:
      mnd:
        # Don't include the "operation" and "assign" checks
        checks: [argument, case, condition, return]
        # Ignore common numbers
        ignored-numbers: [0, 1, 2, 10, 100]
        # Ignore common time numbers
        ignored-functions: [time.Duration, time.Sleep]

  # Misspell
  misspell:
    # Locale to use
    locale: US

  # Unused parameters
  unparam:
    # Report unused function parameters
    check-exported: true

  # Security
  gosec:
    # Which checks to run
    includes:
      - G101 # Look for hard coded credentials
      - G102 # Bind to all interfaces
      - G103 # Audit the use of unsafe block
      - G104 # Audit errors not checked
      - G106 # Audit the use of ssh.InsecureIgnoreHostKey
      - G107 # Url provided to HTTP request as taint input
      - G108 # Profiling endpoint automatically exposed on /debug/pprof
      - G201 # SQL query construction using format string
      - G202 # SQL query construction using string concatenation
      - G203 # Use of unescaped data in HTML templates
      - G204 # Audit use of command execution
      - G301 # Poor file permissions used when creating a directory
      - G302 # Poor file permissions used with chmod
      - G303 # Creating tempfile using a predictable path
      - G304 # File path provided as taint input
      - G305 # File traversal when extracting zip/tar archive
      - G401 # Detect the usage of DES, RC4, MD5 or SHA1
      - G402 # Look for bad TLS connection settings
      - G403 # Ensure minimum RSA key length of 2048 bits
      - G404 # Insecure random number source (rand)
      - G501 # Import blocklist: crypto/md5
      - G502 # Import blocklist: crypto/des
      - G503 # Import blocklist: crypto/rc4
      - G504 # Import blocklist: net/http/cgi
      - G505 # Import blocklist: crypto/sha1

  # Revive (more configurable than golint)
  revive:
    # Confidence level
    confidence: 0.8
    rules:
      - name: blank-imports
      - name: context-as-argument
      - name: context-keys-type
      - name: dot-imports
      - name: error-return
      - name: error-strings
      - name: error-naming
      - name: exported
      - name: if-return
      - name: increment-decrement
      - name: var-naming
      - name: var-declaration
      - name: package-comments
      - name: range
      - name: receiver-naming
      - name: time-naming
      - name: unexported-return
      - name: indent-error-flow
      - name: errorf
      - name: empty-block
      - name: superfluous-else
      - name: unreachable-code
      - name: redefines-builtin-id

  # Static analysis
  staticcheck:
    # Use Go 1.24
    go: "1.24"
    # Check tests too
    checks: ["all", "-ST1000", "-ST1003", "-ST1016", "-ST1020", "-ST1021", "-ST1022"]

  # Style check
  stylecheck:
    # Use Go 1.24
    go: "1.24"
    # Check tests too
    checks: ["all", "-ST1000", "-ST1003", "-ST1016", "-ST1020", "-ST1021", "-ST1022"]

# Linters configuration
linters:
  # Disable all linters
  disable-all: true
  
  # Enable specific linters
  enable:
    # Basic
    - gofmt          # Gofmt checks whether code was gofmt-ed
    - goimports      # Check import statements are formatted
    - govet          # Go vet examines Go source code
    - errcheck       # Checking for unchecked errors
    - staticcheck    # Staticcheck is go vet on steroids
    - ineffassign    # Detects when assignments to existing variables are not used
    - typecheck      # Like the front-end of a Go compiler, parses and type-checks code
    
    # Code quality
    - revive         # Fast, configurable, extensible, flexible, and beautiful linter
    - gosimple       # Simplifies code
    - goconst        # Finds repeated strings that could be replaced by a constant
    - gocyclo        # Computes cyclomatic complexity
    - gocognit       # Computes cognitive complexity
    - maintidx       # Measures the maintainability index
    - funlen         # Tool for detection of long functions
    - lll            # Reports long lines
    
    # Error handling
    - errorlint      # Find code that will cause problems with error wrapping
    - wrapcheck      # Checks that errors returned from external packages are wrapped
    
    # Style
    - stylecheck     # Stylecheck is a replacement for golint
    - misspell       # Finds commonly misspelled English words
    - gofumpt        # Gofumpt checks whether code was gofumpt-ed
    - whitespace     # Detects leading and trailing whitespace
    - unconvert      # Removes unnecessary type conversions
    - gci            # Controls package import order and makes it deterministic
    
    # Performance
    - prealloc       # Finds slice declarations that could potentially be pre-allocated
    - bodyclose      # Checks whether HTTP response body is closed successfully
    
    # Security
    - gosec          # Inspects source code for security problems
    
    # Bugs
    - asciicheck     # Checks that code doesn't contain non-ASCII identifiers
    - bidichk        # Checks for dangerous unicode character sequences
    - durationcheck  # Check for two durations multiplied together
    - exportloopref  # Checks for pointers to enclosing loop variables
    - makezero       # Finds slice declarations with non-zero initial length
    - nilerr         # Finds the code that returns nil even if it checks that the error is not nil
    
    # Testing
    - testpackage    # Makes you use a separate _test package
    - tparallel      # Detects inappropriate usage of t.Parallel()

# Issues configuration
issues:
  # Maximum issues count per one linter
  max-issues-per-linter: 50
  
  # Maximum count of issues with the same text
  max-same-issues: 3
  
  # Show only new issues created after git revision
  new: false
  
  # Exclude certain issues
  exclude-rules:
    # Exclude certain linters for test files
    - path: _test\.go
      linters:
        - errcheck
        - dupl
        - gosec
        - funlen
        - gocognit
        - gocyclo
        
    # Exclude certain patterns
    - linters:
        - lll
      source: "^//go:generate "
      
    # Exclude table-driven test long lines
    - path: _test\.go
      linters:
        - lll
      source: "^\\s*\\{.*\\},?$"
      
    # Allow init functions in main packages
    - path: cmd/
      linters:
        - gochecknoinits
        
    # Allow global variables in config packages
    - path: internal/config/
      linters:
        - gochecknoglobals

# Severity rules
severity:
  # Default severity
  default-severity: warning
  
  # Specific severities
  rules:
    - linters:
        - gosec
      severity: error
    - linters:
        - errcheck
      severity: error
    - linters:
        - staticcheck
      severity: error
```

#### 2. Pre-commit Hooks Setup

Create `.pre-commit-config.yaml`:

```yaml
# Pre-commit hook configuration
repos:
  # Go formatting and linting
  - repo: local
    hooks:
      - id: go-fmt
        name: Go Format
        entry: bash -c 'go fmt ./...'
        language: system
        files: '\.go$'
        pass_filenames: false
        
      - id: go-vet
        name: Go Vet
        entry: bash -c 'go vet ./...'
        language: system
        files: '\.go$'
        pass_filenames: false
        
      - id: go-lint
        name: Go Lint
        entry: bash -c 'golangci-lint run ./...'
        language: system
        files: '\.go$'
        pass_filenames: false
        
      - id: go-mod-tidy
        name: Go Mod Tidy
        entry: bash -c 'go mod tidy && git diff --exit-code go.mod go.sum'
        language: system
        files: 'go\.mod|go\.sum$'
        pass_filenames: false
        
      - id: go-test
        name: Go Test
        entry: bash -c 'go test -short ./...'
        language: system
        files: '\.go$'
        pass_filenames: false
        
      - id: go-security
        name: Go Security Check
        entry: bash -c 'govulncheck ./...'
        language: system
        files: '\.go$'
        pass_filenames: false

  # General file checks
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-json
      - id: check-added-large-files
        args: ['--maxkb=500']
      - id: check-case-conflict
      - id: check-merge-conflict
      - id: detect-private-key
```

Create installation script `scripts/setup-pre-commit.sh`:

```bash
#!/bin/bash
set -e

echo "Setting up pre-commit hooks..."

# Check if pre-commit is installed
if ! command -v pre-commit &> /dev/null; then
    echo "Installing pre-commit..."
    pip install pre-commit || pip3 install pre-commit
fi

# Install the pre-commit hooks
pre-commit install

# Run pre-commit on all files to verify setup
echo "Running pre-commit checks..."
pre-commit run --all-files || true

echo "Pre-commit hooks installed successfully!"
```

#### 3. Custom Git Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Custom pre-commit hook for additional checks

set -e

echo "Running pre-commit checks..."

# Format check
echo "Checking code formatting..."
if ! go fmt ./... | grep -q .; then
    echo "✓ Code formatting OK"
else
    echo "✗ Code needs formatting. Run 'make fmt'"
    exit 1
fi

# Vet check
echo "Running go vet..."
if go vet ./... 2>&1 | grep -q .; then
    echo "✗ Go vet found issues"
    exit 1
else
    echo "✓ Go vet passed"
fi

# Lint check
echo "Running golangci-lint..."
if golangci-lint run ./... > /dev/null 2>&1; then
    echo "✓ Linting passed"
else
    echo "✗ Linting failed. Run 'make lint' for details"
    exit 1
fi

# Quick test
echo "Running quick tests..."
if go test -short ./... > /dev/null 2>&1; then
    echo "✓ Quick tests passed"
else
    echo "✗ Tests failed. Run 'make test' for details"
    exit 1
fi

echo "All pre-commit checks passed!"
```

#### 4. Makefile Integration

Add to Makefile (building on T01_S02):

```makefile
# Code quality targets
.PHONY: lint fmt vet sec mod-tidy quality

# Linting with golangci-lint
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

# Lint and fix issues automatically where possible
lint-fix:
	@echo "Running golangci-lint with fixes..."
	@golangci-lint run --fix ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofumpt -w .

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Security vulnerability check
sec:
	@echo "Running security checks..."
	@govulncheck ./...

# Tidy dependencies
mod-tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@go mod verify

# Run all quality checks
quality: fmt vet lint sec
	@echo "All quality checks passed!"

# Install quality tools
install-tools:
	@echo "Installing code quality tools..."
	@go install mvdan.cc/gofumpt@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
	@echo "Tools installed successfully!"

# Pre-commit setup
setup-pre-commit:
	@echo "Setting up pre-commit hooks..."
	@chmod +x scripts/setup-pre-commit.sh
	@./scripts/setup-pre-commit.sh

# Code complexity report
complexity:
	@echo "Generating complexity report..."
	@gocyclo -top 10 -avg ./...
	@echo ""
	@echo "Cognitive complexity:"
	@gocognit -top 10 -avg ./...

# Full quality report
quality-report: quality complexity coverage-report
	@echo "Full quality report generated!"
```

#### 5. Code Quality Monitoring Script

Create `scripts/code-quality-monitor.sh`:

```bash
#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Code Quality Report ===${NC}"
echo "Generated at: $(date)"
echo

# 1. Formatting Check
echo -e "${BLUE}1. Code Formatting${NC}"
if go fmt ./... | grep -q .; then
    echo -e "${RED}✗ Code needs formatting${NC}"
    ISSUES+=("formatting")
else
    echo -e "${GREEN}✓ Code is properly formatted${NC}"
fi
echo

# 2. Linting Results
echo -e "${BLUE}2. Linting Results${NC}"
LINT_OUTPUT=$(golangci-lint run ./... 2>&1 || true)
LINT_COUNT=$(echo "$LINT_OUTPUT" | grep -c "^.*\.go:" || true)
if [ $LINT_COUNT -eq 0 ]; then
    echo -e "${GREEN}✓ No linting issues found${NC}"
else
    echo -e "${YELLOW}⚠ Found $LINT_COUNT linting issues${NC}"
    echo "$LINT_OUTPUT" | head -10
    echo "... (showing first 10 issues)"
    ISSUES+=("linting")
fi
echo

# 3. Cyclomatic Complexity
echo -e "${BLUE}3. Cyclomatic Complexity${NC}"
echo "Top 5 most complex functions:"
gocyclo -top 5 ./... | while read line; do
    COMPLEXITY=$(echo $line | awk '{print $1}')
    if [ $COMPLEXITY -gt 15 ]; then
        echo -e "${RED}$line${NC}"
    elif [ $COMPLEXITY -gt 10 ]; then
        echo -e "${YELLOW}$line${NC}"
    else
        echo -e "${GREEN}$line${NC}"
    fi
done
echo

# 4. Security Vulnerabilities
echo -e "${BLUE}4. Security Scan${NC}"
if govulncheck ./... > /dev/null 2>&1; then
    echo -e "${GREEN}✓ No known vulnerabilities found${NC}"
else
    echo -e "${RED}✗ Security vulnerabilities detected${NC}"
    govulncheck ./... 2>&1 | head -20
    ISSUES+=("security")
fi
echo

# 5. Dependencies
echo -e "${BLUE}5. Dependencies${NC}"
OUTDATED=$(go list -u -m all 2>/dev/null | grep -c '\[' || true)
if [ $OUTDATED -gt 0 ]; then
    echo -e "${YELLOW}⚠ $OUTDATED outdated dependencies${NC}"
else
    echo -e "${GREEN}✓ All dependencies up to date${NC}"
fi
echo

# 6. Test Coverage Summary
echo -e "${BLUE}6. Test Coverage${NC}"
go test -cover ./... | grep -E "coverage:|FAIL" | while read line; do
    if echo $line | grep -q "FAIL"; then
        echo -e "${RED}$line${NC}"
    else
        COV=$(echo $line | grep -oE '[0-9]+\.[0-9]+%' | sed 's/%//')
        if (( $(echo "$COV < 70" | bc -l) )); then
            echo -e "${RED}$line${NC}"
        elif (( $(echo "$COV < 80" | bc -l) )); then
            echo -e "${YELLOW}$line${NC}"
        else
            echo -e "${GREEN}$line${NC}"
        fi
    fi
done
echo

# Summary
echo -e "${BLUE}=== Summary ===${NC}"
if [ ${#ISSUES[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ All quality checks passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Issues found in: ${ISSUES[*]}${NC}"
    exit 1
fi
```

#### 6. CI/CD Templates

Create `.github/workflows/quality.yml`:

```yaml
name: Code Quality

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  quality:
    name: Code Quality Checks
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.24.2'
    
    - name: Install tools
      run: make install-tools
    
    - name: Check formatting
      run: |
        if [ -n "$(go fmt ./...)" ]; then
          echo "Code needs formatting. Run 'go fmt ./...'"
          exit 1
        fi
    
    - name: Run go vet
      run: make vet
    
    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: v1.55.2
        args: --timeout=5m
    
    - name: Run security scan
      run: make sec
    
    - name: Check dependencies
      run: |
        go mod tidy
        git diff --exit-code go.mod go.sum
    
    - name: Generate quality report
      run: |
        chmod +x scripts/code-quality-monitor.sh
        ./scripts/code-quality-monitor.sh | tee quality-report.txt
    
    - name: Upload quality report
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: quality-report
        path: quality-report.txt
```

### Implementation Notes

1. **Phase 1**: Create `.golangci.yml` configuration
2. **Phase 2**: Set up pre-commit hooks and scripts
3. **Phase 3**: Integrate with Makefile from T01_S02
4. **Phase 4**: Create monitoring and reporting scripts
5. **Phase 5**: Prepare CI/CD templates
6. **Phase 6**: Document standards and workflows

### Key Integration Points

1. **Makefile Integration**: All quality commands accessible via make
2. **Pre-commit Hooks**: Automatic validation before commits
3. **Existing Tools**: Leverage golangci-lint at `/usr/local/bin/golangci-lint`
4. **Testing Infrastructure**: Coordinate with T01_S02 test setup
5. **Coverage Integration**: Work with T05_S02 coverage configuration

### Best Practices

1. **Progressive Enhancement**: Start with basic checks, add more over time
2. **Fast Feedback**: Pre-commit hooks should be fast (use -short tests)
3. **Clear Reporting**: Quality issues should be clearly reported
4. **Automated Fixes**: Use tools that can auto-fix where possible
5. **Customization**: Rules should match project needs, not be generic

## Complexity Assessment

**Estimated Complexity**: Medium-High (6)

**Factors**:
- Comprehensive golangci-lint configuration requires careful tuning
- Pre-commit hook setup involves multiple integration points
- Custom linting rules need testing and validation
- Monitoring scripts require bash scripting expertise
- CI/CD templates need to work across environments
- Integration with existing tools adds coordination complexity

## References

- golangci-lint documentation: https://golangci-lint.run/
- Pre-commit framework: https://pre-commit.com/
- Go code review comments: https://github.com/golang/go/wiki/CodeReviewComments
- Effective Go: https://golang.org/doc/effective_go
- `.claude/default-rules.md` for existing code standards
- CLAUDE.md for project-specific requirements