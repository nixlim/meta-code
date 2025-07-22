# Task: Test Coverage Configuration (T05_S02)

## Summary
Set up comprehensive test coverage reporting infrastructure, integrate coverage checks with CI/CD pipeline, establish and monitor coverage targets across all packages, and create automated coverage tracking mechanisms.

## Objective
Establish a robust test coverage measurement and reporting system that provides visibility into code coverage metrics, enforces minimum coverage thresholds, and integrates seamlessly with the development workflow. This includes setting up coverage profiling, HTML report generation, package-specific monitoring, and establishing coverage gates for quality assurance.

## Acceptance Criteria
- [ ] Configure go test coverage profiling with appropriate flags and options
- [ ] Set up HTML coverage report generation with automatic opening in browser
- [ ] Establish package-specific coverage targets (minimum 70% overall)
- [ ] Create coverage threshold enforcement mechanisms
- [ ] Implement coverage trend tracking and reporting
- [ ] Set up CI/CD integration for coverage checks (when CI/CD is established)
- [ ] Create coverage badge generation for README
- [ ] Document coverage workflows and best practices

## Technical Guidance

### Current State Analysis
Based on CLAUDE.md, the project already has:
- Basic coverage commands documented (`go test -cover ./...`, `go test -coverprofile=coverage.out ./...`)
- Existing coverage.out file in project root
- Current coverage status tracked:
  - `jsonrpc`: 93.3% ✅
  - `connection`: 87.0% ✅  
  - `router`: 86.5% ⚠️ (some async tests failing)
  - `mcp`: 63.9% ⚠️
  - `handlers`: 42.2% ❌

### Required Implementation

#### 1. Enhanced Coverage Commands in Makefile

Add sophisticated coverage targets to the Makefile:

```makefile
# Coverage configuration
COVERAGE_THRESHOLD := 70
COVERAGE_PROFILE := coverage.out
COVERAGE_HTML := coverage.html
COVERAGE_MODE := atomic

# Basic coverage commands
coverage:
	@echo "Running tests with coverage..."
	@go test -covermode=$(COVERAGE_MODE) -coverprofile=$(COVERAGE_PROFILE) ./...
	@go tool cover -func=$(COVERAGE_PROFILE) | grep "total:" | awk '{print "Total Coverage: " $$3}'

coverage-html: coverage
	@echo "Generating HTML coverage report..."
	@go tool cover -html=$(COVERAGE_PROFILE) -o=$(COVERAGE_HTML)
	@echo "Opening coverage report in browser..."
	@open $(COVERAGE_HTML) 2>/dev/null || xdg-open $(COVERAGE_HTML) 2>/dev/null || echo "Please open $(COVERAGE_HTML) manually"

# Package-specific coverage
coverage-package:
	@if [ -z "$(PKG)" ]; then echo "Usage: make coverage-package PKG=./internal/protocol/jsonrpc"; exit 1; fi
	@go test -covermode=$(COVERAGE_MODE) -coverprofile=$(COVERAGE_PROFILE) $(PKG)
	@go tool cover -func=$(COVERAGE_PROFILE) | grep "total:"

# Coverage with threshold checking
coverage-check: coverage
	@echo "Checking coverage threshold ($(COVERAGE_THRESHOLD)%)..."
	@bash -c 'COV=$$(go tool cover -func=$(COVERAGE_PROFILE) | grep "total:" | awk "{print int(\$$3)}"); \
	if [ $$COV -lt $(COVERAGE_THRESHOLD) ]; then \
		echo "Coverage $$COV% is below threshold $(COVERAGE_THRESHOLD)%"; exit 1; \
	else \
		echo "Coverage $$COV% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi'

# Detailed coverage report by package
coverage-report:
	@echo "Package Coverage Report:"
	@echo "======================="
	@for pkg in $$(go list ./... | grep -v /vendor/); do \
		go test -covermode=$(COVERAGE_MODE) -coverprofile=temp.out $$pkg 2>/dev/null; \
		if [ -f temp.out ]; then \
			COV=$$(go tool cover -func=temp.out | grep "total:" | awk '{print $$3}'); \
			printf "%-60s %s\n" $$pkg $$COV; \
		fi; \
	done
	@rm -f temp.out

# Coverage excluding test files
coverage-no-tests:
	@go test -covermode=$(COVERAGE_MODE) -coverprofile=$(COVERAGE_PROFILE) -coverpkg=./... ./...
	@go tool cover -func=$(COVERAGE_PROFILE)

# Coverage diff (requires git)
coverage-diff:
	@echo "Calculating coverage difference from main branch..."
	@git diff main --name-only | grep "\.go$$" | grep -v "_test\.go$$" | xargs -I {} dirname {} | sort -u | xargs go test -cover
```

#### 2. Coverage Monitoring Script

Create `scripts/coverage-monitor.sh`:

```bash
#!/bin/bash
set -e

# Configuration
THRESHOLD_CRITICAL=70
THRESHOLD_WARNING=80
THRESHOLD_GOOD=90

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Function to get package coverage
get_coverage() {
    local pkg=$1
    local cov=$(go test -cover $pkg 2>/dev/null | grep -o '[0-9]*\.[0-9]*%' | sed 's/%//')
    echo ${cov:-0}
}

# Monitor all packages
echo "Coverage Monitoring Report"
echo "========================="
echo

failed_packages=()
warning_packages=()

for pkg in $(go list ./... | grep -v /vendor/); do
    coverage=$(get_coverage $pkg)
    coverage_int=${coverage%.*}
    
    if [ -z "$coverage_int" ] || [ "$coverage_int" -eq 0 ]; then
        coverage_int=0
    fi
    
    # Determine status
    if [ $coverage_int -lt $THRESHOLD_CRITICAL ]; then
        status="${RED}❌ CRITICAL${NC}"
        failed_packages+=("$pkg: $coverage%")
    elif [ $coverage_int -lt $THRESHOLD_WARNING ]; then
        status="${YELLOW}⚠️  WARNING${NC}"
        warning_packages+=("$pkg: $coverage%")
    elif [ $coverage_int -lt $THRESHOLD_GOOD ]; then
        status="${GREEN}✓ GOOD${NC}"
    else
        status="${GREEN}✅ EXCELLENT${NC}"
    fi
    
    printf "%-60s %6s%% %s\n" "$pkg" "$coverage" "$status"
done

# Summary
echo
echo "Summary:"
echo "--------"
if [ ${#failed_packages[@]} -gt 0 ]; then
    echo -e "${RED}Failed Packages (< ${THRESHOLD_CRITICAL}%):${NC}"
    for pkg in "${failed_packages[@]}"; do
        echo "  - $pkg"
    done
fi

if [ ${#warning_packages[@]} -gt 0 ]; then
    echo -e "${YELLOW}Warning Packages (< ${THRESHOLD_WARNING}%):${NC}"
    for pkg in "${warning_packages[@]}"; do
        echo "  - $pkg"
    done
fi

# Exit with error if any package is below critical threshold
if [ ${#failed_packages[@]} -gt 0 ]; then
    exit 1
fi
```

#### 3. Coverage Configuration File

Create `.coverage.yml`:

```yaml
# Test coverage configuration
coverage:
  # Global threshold
  threshold: 70
  
  # Package-specific thresholds
  packages:
    - path: ./internal/protocol/jsonrpc
      threshold: 90
    - path: ./internal/protocol/connection
      threshold: 85
    - path: ./internal/protocol/router
      threshold: 80
    - path: ./internal/protocol/mcp
      threshold: 70
    - path: ./internal/protocol/handlers
      threshold: 70
    - path: ./internal/logging
      threshold: 80
    
  # Files to exclude from coverage
  exclude:
    - "**/*_test.go"
    - "**/mocks/**"
    - "**/testdata/**"
    - "**/examples/**"
    - "cmd/**"
    
  # Coverage modes
  mode: atomic  # Options: set, count, atomic
  
  # Report formats
  reports:
    - func    # Function coverage
    - html    # HTML report
    - json    # JSON output for tools
```

#### 4. GitHub Actions Integration (future)

Create `.github/workflows/coverage.yml` template:

```yaml
name: Coverage Check

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  coverage:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24.2'
    
    - name: Run tests with coverage
      run: make coverage-check
    
    - name: Generate coverage report
      run: make coverage-report > coverage-summary.txt
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        fail_ci_if_error: true
    
    - name: Comment PR with coverage
      if: github.event_name == 'pull_request'
      uses: actions/github-script@v6
      with:
        script: |
          const fs = require('fs');
          const coverage = fs.readFileSync('coverage-summary.txt', 'utf8');
          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body: '## Coverage Report\n\n```\n' + coverage + '\n```'
          });
```

#### 5. Coverage Badge Generation

Add to Makefile:

```makefile
coverage-badge:
	@COV=$$(go tool cover -func=$(COVERAGE_PROFILE) | grep "total:" | awk '{print int($$3)}'); \
	if [ $$COV -ge 90 ]; then COLOR="brightgreen"; \
	elif [ $$COV -ge 80 ]; then COLOR="green"; \
	elif [ $$COV -ge 70 ]; then COLOR="yellow"; \
	elif [ $$COV -ge 60 ]; then COLOR="orange"; \
	else COLOR="red"; fi; \
	echo "[![Coverage](https://img.shields.io/badge/coverage-$$COV%25-$$COLOR)]()"
```

### Key Integration Points

1. **Makefile Integration**: All coverage commands should be accessible via make targets
2. **Existing Coverage Data**: Build upon the current coverage.out file
3. **Package Structure**: Respect the existing package organization
4. **Testing Infrastructure**: Integrate with T01_S02 testing setup
5. **Documentation**: Update testing.md with coverage guidelines

### Implementation Notes

1. **Phase 1**: Add coverage targets to Makefile
2. **Phase 2**: Create coverage monitoring script
3. **Phase 3**: Set up coverage configuration file
4. **Phase 4**: Implement threshold enforcement
5. **Phase 5**: Prepare CI/CD templates for future use
6. **Phase 6**: Document coverage workflows

### Coverage Best Practices

1. **Incremental Coverage**: Focus on increasing coverage for critical paths first
2. **Meaningful Tests**: Avoid coverage-driven testing; write meaningful tests
3. **Edge Cases**: Prioritize edge case coverage over happy path
4. **Integration Coverage**: Don't ignore integration test coverage
5. **Regular Monitoring**: Run coverage reports as part of development workflow

## Complexity Assessment

**Estimated Complexity**: Medium (4)

**Factors**:
- Go's built-in coverage tooling simplifies implementation
- Makefile targets are straightforward to implement
- Threshold enforcement requires some bash scripting
- Package-specific monitoring adds moderate complexity
- CI/CD preparation is template-based (low complexity)
- Good foundation exists with current coverage tracking

## References

- Go coverage documentation: https://go.dev/blog/cover
- Current coverage status in CLAUDE.md
- Makefile being created in T01_S02
- Testing infrastructure from T01_S02
- Go test coverage tutorial: https://go.dev/doc/tutorial/add-a-test
- Codecov documentation: https://docs.codecov.com/docs