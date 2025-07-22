# Linting Configuration for meta-mcp-server

This document describes the linting setup and code quality tools configured for the meta-mcp-server project.

## Overview

The project uses [golangci-lint](https://golangci-lint.run/) as the primary code quality tool. It's a fast Go linters aggregator that runs multiple linters in parallel and uses caching for improved performance.

## Configuration

The linting configuration is defined in `.golangci.yml` at the project root. The configuration includes:

### Core Linters (Required by Specification)

1. **gofmt** - Ensures code is formatted according to Go standards
2. **revive** - Fast, configurable, extensible linter (replaces deprecated golint)
3. **govet** - Reports suspicious constructs
4. **errcheck** - Checks for unchecked errors
5. **ineffassign** - Detects unused variable assignments
6. **goconst** - Finds repeated strings that could be constants
7. **gocyclo** - Checks cyclomatic complexity (configured with min-complexity: 15)
8. **misspell** - Finds misspelled English words in comments

### Additional Linters

The configuration includes many additional linters for comprehensive code quality:

- **staticcheck** - Advanced static analysis
- **gosimple** - Simplifies code
- **unused** - Finds unused code
- **gosec** - Security analysis
- **bodyclose** - Checks HTTP response body closure
- **dupl** - Detects code duplication
- **funlen** - Limits function length (100 lines, 50 statements)
- **gocognit** - Cognitive complexity checker (min-complexity: 20)
- **gocritic** - Provides various diagnostics
- **gofumpt** - Stricter gofmt
- **goimports** - Manages imports and formatting
- And many more...

### Special Configurations

1. **Test Files Exception**: `errcheck` is excluded from test files (`*_test.go`)
2. **Complexity Limits**:
   - Cyclomatic complexity: 15
   - Cognitive complexity: 20
   - Function length: 100 lines / 50 statements
3. **Line Length**: Maximum 120 characters
4. **Import Grouping**: Local imports (`github.com/nixlim/meta-mcp-server`) are grouped separately

## Usage

### Running Linters

```bash
# Run all linters
make lint

# Run linters with auto-fix
make lint-fix

# Install/update golangci-lint
make lint-install
```

### Manual golangci-lint Commands

```bash
# Run with default configuration
golangci-lint run

# Run with auto-fix
golangci-lint run --fix

# Run on specific directories
golangci-lint run ./internal/...

# Run specific linters
golangci-lint run --enable-only gofmt,govet

# Show all available linters
golangci-lint linters
```

## Pre-commit Hook

A pre-commit hook is installed at `.git/hooks/pre-commit` that automatically runs golangci-lint on staged Go files before each commit.

### Installing the Hook

```bash
# The hook is already in .git/hooks/pre-commit
# To reinstall or update:
./scripts/install-hooks.sh
```

### How the Hook Works

1. Detects staged Go files
2. Creates a temporary directory with staged content
3. Runs golangci-lint on the staged files
4. Prevents commit if issues are found
5. Suggests running `golangci-lint run --fix` for auto-fixable issues

### Bypassing the Hook

In emergency situations, you can bypass the hook:

```bash
git commit --no-verify
```

**Note**: This is not recommended and should only be used in exceptional circumstances.

## CI/CD Integration

The Makefile includes CI-specific targets:

```bash
# CI lint with GitHub Actions format
make ci-lint

# CI test with coverage
make ci-test

# CI build verification
make ci-build
```

## Common Issues and Solutions

### 1. Import Formatting

**Issue**: Imports not grouped correctly
**Solution**: Run `make lint-fix` or `goimports -w .`

### 2. Cyclomatic Complexity

**Issue**: Function exceeds complexity limit (15)
**Solution**: Refactor the function into smaller, more focused functions

### 3. Error Checking

**Issue**: Unchecked errors reported by errcheck
**Solution**: Always handle errors appropriately:

```go
// Bad
result, _ := someFunction()

// Good
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

### 4. Cognitive Complexity

**Issue**: Function is too complex to understand
**Solution**: Simplify logic, extract helper functions, reduce nesting

### 5. Line Length

**Issue**: Line exceeds 120 characters
**Solution**: Break long lines appropriately:

```go
// Bad
veryLongFunctionNameWithManyParameters(firstParameter, secondParameter, thirdParameter, fourthParameter)

// Good
veryLongFunctionNameWithManyParameters(
    firstParameter,
    secondParameter,
    thirdParameter,
    fourthParameter,
)
```

## Best Practices

1. **Run linters before committing**: The pre-commit hook ensures this
2. **Fix issues immediately**: Don't let linting issues accumulate
3. **Use auto-fix when possible**: `make lint-fix` can resolve many issues
4. **Understand the warnings**: Don't just suppress them; understand why they exist
5. **Configure appropriately**: If a linter rule doesn't make sense for the project, configure it in `.golangci.yml`

## Suppressing Warnings

When necessary, you can suppress specific warnings:

```go
// Suppress for a single line
//nolint:errcheck
result, _ := someFunction()

// Suppress for a function
//nolint:gocyclo
func complexFunction() {
    // complex logic
}

// Suppress with reason (recommended)
//nolint:gocyclo // This complexity is necessary for performance
func optimizedFunction() {
    // complex but optimized logic
}
```

**Important**: Always document why you're suppressing a linter warning.

## Adding New Linters

To add a new linter:

1. Edit `.golangci.yml`
2. Add the linter name to the `enable` list
3. Configure any linter-specific settings in `linters-settings`
4. Test the configuration: `golangci-lint run`
5. Update this documentation

## Resources

- [golangci-lint Documentation](https://golangci-lint.run/)
- [Available Linters](https://golangci-lint.run/usage/linters/)
- [Configuration Reference](https://golangci-lint.run/usage/configuration/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)