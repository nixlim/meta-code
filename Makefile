# Meta MCP Server - Comprehensive Testing Makefile
# Generated for T01_S02 Testing Infrastructure Setup

# Variables
BINARY_NAME := meta-mcp-server
GO := go
PACKAGES := $(shell $(GO) list ./... 2>/dev/null || echo "")
TESTPKGS := $(shell $(GO) list ./... 2>/dev/null | grep -v /vendor/ | grep -v /examples/)
COVERAGE_DIR := coverage
PROFILE_DIR := profiles
BENCH_DIR := benchmarks

# Version and build info
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"
GOFLAGS := -v

# Coverage threshold
COVERAGE_THRESHOLD := 80

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

# Ensure required directories exist
$(shell mkdir -p $(COVERAGE_DIR) $(PROFILE_DIR) $(BENCH_DIR))

##@ Testing

.PHONY: test
test: ## Run all tests
	@echo "$(BLUE)Running all tests...$(NC)"
	@$(GO) test $(GOFLAGS) -race -coverprofile=$(COVERAGE_DIR)/coverage.out $(TESTPKGS)
	@echo "$(GREEN)✓ All tests completed$(NC)"

.PHONY: test-short
test-short: ## Run only short tests (exclude integration)
	@echo "$(BLUE)Running short tests...$(NC)"
	@$(GO) test $(GOFLAGS) -short -race $(TESTPKGS)
	@echo "$(GREEN)✓ Short tests completed$(NC)"

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo "$(BLUE)Running tests with verbose output...$(NC)"
	@$(GO) test $(GOFLAGS) -v -race $(TESTPKGS)

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "$(BLUE)Running tests with race detector...$(NC)"
	@$(GO) test $(GOFLAGS) -race $(TESTPKGS)
	@echo "$(GREEN)✓ Race detection completed$(NC)"

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "$(BLUE)Running unit tests...$(NC)"
	@$(GO) test $(GOFLAGS) -short -race -tags=unit $(TESTPKGS)
	@echo "$(GREEN)✓ Unit tests completed$(NC)"

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo "$(BLUE)Running integration tests...$(NC)"
	@$(GO) test $(GOFLAGS) -race -tags=integration -timeout=10m $(TESTPKGS)
	@echo "$(GREEN)✓ Integration tests completed$(NC)"

.PHONY: test-conformance
test-conformance: ## Run MCP conformance tests
	@echo "$(BLUE)Running MCP conformance tests...$(NC)"
	@$(GO) test $(GOFLAGS) -race -tags=conformance ./tests/conformance/...
	@echo "$(GREEN)✓ Conformance tests completed$(NC)"

##@ Coverage

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@$(GO) test $(GOFLAGS) -race -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out $(TESTPKGS)
	@$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep "total:" | awk '{print "Total Coverage: " $$3}'
	@echo "$(BLUE)Checking coverage threshold ($(COVERAGE_THRESHOLD)%)...$(NC)"
	@bash -c 'COVERAGE=$$($(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep "total:" | awk "{print \$$3}" | sed "s/%//"); \
		if [ $$(echo "$$COVERAGE < $(COVERAGE_THRESHOLD)" | bc) -eq 1 ]; then \
			echo "$(RED)✗ Coverage $$COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%$(NC)"; \
			exit 1; \
		else \
			echo "$(GREEN)✓ Coverage $$COVERAGE% meets threshold$(NC)"; \
		fi'

.PHONY: test-coverage-html
test-coverage-html: test-coverage ## Generate HTML coverage report
	@echo "$(BLUE)Generating HTML coverage report...$(NC)"
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)✓ Coverage report generated at $(COVERAGE_DIR)/coverage.html$(NC)"
	@if command -v open >/dev/null 2>&1; then \
		open $(COVERAGE_DIR)/coverage.html; \
	elif command -v xdg-open >/dev/null 2>&1; then \
		xdg-open $(COVERAGE_DIR)/coverage.html; \
	fi

.PHONY: coverage-by-package
coverage-by-package: ## Show coverage by package
	@echo "$(BLUE)Coverage by package:$(NC)"
	@$(GO) test $(GOFLAGS) -coverprofile=$(COVERAGE_DIR)/coverage.out $(TESTPKGS) >/dev/null 2>&1 || true
	@$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep -v "total:"

##@ Code Quality

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(BLUE)Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
		echo "$(GREEN)✓ Linting completed$(NC)"; \
	else \
		echo "$(YELLOW)⚠ golangci-lint not installed. Install with: make lint-install$(NC)"; \
		exit 1; \
	fi

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	@echo "$(BLUE)Running linters with auto-fix...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix --timeout=5m; \
		echo "$(GREEN)✓ Linting with fixes completed$(NC)"; \
	else \
		echo "$(YELLOW)⚠ golangci-lint not installed. Install with: make lint-install$(NC)"; \
		exit 1; \
	fi

.PHONY: lint-install
lint-install: ## Install or update golangci-lint
	@echo "$(BLUE)Installing/updating golangci-lint...$(NC)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.61.0
	@echo "$(GREEN)✓ golangci-lint installed successfully$(NC)"

.PHONY: fmt
fmt: ## Format code with gofmt
	@echo "$(BLUE)Formatting code...$(NC)"
	@gofmt -s -w .
	@echo "$(GREEN)✓ Code formatted$(NC)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	@$(GO) vet $(PACKAGES)
	@echo "$(GREEN)✓ go vet completed$(NC)"

.PHONY: mod-tidy
mod-tidy: ## Run go mod tidy
	@echo "$(BLUE)Tidying go modules...$(NC)"
	@$(GO) mod tidy
	@echo "$(GREEN)✓ go mod tidy completed$(NC)"

.PHONY: check
check: fmt vet lint mod-tidy ## Run all code checks
	@echo "$(GREEN)✓ All checks passed$(NC)"

##@ Performance

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	@$(GO) test -bench=. -benchmem -count=5 -run=^$$ $(TESTPKGS) | tee $(BENCH_DIR)/bench-$(VERSION).txt
	@echo "$(GREEN)✓ Benchmarks completed$(NC)"

.PHONY: benchmark-compare
benchmark-compare: ## Compare benchmark results
	@echo "$(BLUE)Comparing benchmarks...$(NC)"
	@if [ -f $(BENCH_DIR)/bench-baseline.txt ]; then \
		if command -v benchstat >/dev/null 2>&1; then \
			benchstat $(BENCH_DIR)/bench-baseline.txt $(BENCH_DIR)/bench-$(VERSION).txt; \
		else \
			echo "$(YELLOW)⚠ benchstat not installed. Install with: go install golang.org/x/perf/cmd/benchstat@latest$(NC)"; \
		fi; \
	else \
		echo "$(YELLOW)⚠ No baseline benchmark found. Run 'make benchmark-baseline' first$(NC)"; \
	fi

.PHONY: benchmark-baseline
benchmark-baseline: ## Save current benchmarks as baseline
	@echo "$(BLUE)Setting benchmark baseline...$(NC)"
	@$(GO) test -bench=. -benchmem -count=5 -run=^$$ $(TESTPKGS) | tee $(BENCH_DIR)/bench-baseline.txt
	@echo "$(GREEN)✓ Baseline benchmark saved$(NC)"

.PHONY: profile
profile: profile-cpu profile-mem ## Generate CPU and memory profiles

.PHONY: profile-cpu
profile-cpu: ## Generate CPU profile
	@echo "$(BLUE)Generating CPU profile...$(NC)"
	@$(GO) test -cpuprofile=$(PROFILE_DIR)/cpu.prof -bench=. -benchtime=10s -run=^$$ $(TESTPKGS)
	@echo "$(GREEN)✓ CPU profile generated at $(PROFILE_DIR)/cpu.prof$(NC)"
	@echo "View with: go tool pprof $(PROFILE_DIR)/cpu.prof"

.PHONY: profile-mem
profile-mem: ## Generate memory profile
	@echo "$(BLUE)Generating memory profile...$(NC)"
	@$(GO) test -memprofile=$(PROFILE_DIR)/mem.prof -bench=. -benchtime=10s -run=^$$ $(TESTPKGS)
	@echo "$(GREEN)✓ Memory profile generated at $(PROFILE_DIR)/mem.prof$(NC)"
	@echo "View with: go tool pprof $(PROFILE_DIR)/mem.prof"

##@ CI/CD

.PHONY: ci
ci: deps check test-coverage ## Run full CI pipeline
	@echo "$(GREEN)✓ CI pipeline completed successfully$(NC)"

.PHONY: ci-test
ci-test: ## Run tests in CI mode
	@echo "$(BLUE)Running CI tests...$(NC)"
	@$(GO) test $(GOFLAGS) -race -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out -json $(TESTPKGS) | tee test-results.json
	@echo "$(GREEN)✓ CI tests completed$(NC)"

.PHONY: pre-commit
pre-commit: fmt check test-short ## Run pre-commit checks
	@echo "$(GREEN)✓ Pre-commit checks passed$(NC)"

##@ Development

.PHONY: watch
watch: ## Run tests in watch mode (requires entr)
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" -not -path "./vendor/*" | entr -c make test-short; \
	else \
		echo "$(YELLOW)⚠ entr not installed. Install with: brew install entr (macOS) or apt-get install entr (Linux)$(NC)"; \
		exit 1; \
	fi

.PHONY: test-quick
test-quick: ## Run quick tests (no race detector)
	@echo "$(BLUE)Running quick tests...$(NC)"
	@$(GO) test -short $(TESTPKGS)
	@echo "$(GREEN)✓ Quick tests completed$(NC)"

.PHONY: test-pkg
test-pkg: ## Test specific package (use PKG=./path/to/package)
	@if [ -z "$(PKG)" ]; then \
		echo "$(RED)✗ Please specify a package: make test-pkg PKG=./path/to/package$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Testing package $(PKG)...$(NC)"
	@$(GO) test $(GOFLAGS) -race -v $(PKG)

.PHONY: test-func
test-func: ## Test specific function (use FUNC=TestName PKG=./path)
	@if [ -z "$(FUNC)" ] || [ -z "$(PKG)" ]; then \
		echo "$(RED)✗ Please specify both function and package: make test-func FUNC=TestName PKG=./path$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Testing function $(FUNC) in package $(PKG)...$(NC)"
	@$(GO) test $(GOFLAGS) -race -v -run=$(FUNC) $(PKG)

##@ Utilities

.PHONY: deps
deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	@$(GO) mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	@$(GO) get -u -t ./...
	@$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install golang.org/x/perf/cmd/benchstat@latest
	@$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)✓ Tools installed$(NC)"

.PHONY: clean
clean: ## Clean build and test artifacts
	@echo "$(BLUE)Cleaning artifacts...$(NC)"
	@rm -rf $(COVERAGE_DIR) $(PROFILE_DIR) $(BENCH_DIR)
	@rm -f test-results.json
	@$(GO) clean -testcache
	@echo "$(GREEN)✓ Cleanup completed$(NC)"

.PHONY: test-stats
test-stats: ## Show test statistics
	@echo "$(BLUE)Test Statistics:$(NC)"
	@echo "Total packages: $$(echo '$(TESTPKGS)' | wc -w | xargs)"
	@echo "Total test files: $$(find . -name "*_test.go" -not -path "./vendor/*" | wc -l | xargs)"
	@echo "Total test functions: $$(grep -r "^func Test" --include="*_test.go" . | wc -l | xargs)"
	@echo "Total benchmark functions: $$(grep -r "^func Benchmark" --include="*_test.go" . | wc -l | xargs)"

##@ Help

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BLUE)Meta MCP Server Testing Makefile$(NC)\n\nUsage:\n  make $(GREEN)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# Keep the Makefile updated
.PHONY: self-update
self-update: ## Update this Makefile from the template
	@echo "$(YELLOW)⚠ This would update the Makefile from a template (not implemented)$(NC)"