# Linting Guidelines

This document explains how to use the linting tools configured for the Freightliner project.

## Overview

Freightliner uses [golangci-lint](https://golangci-lint.run/) to enforce code quality standards. The tool is configured with specific rules in the `.golangci.yml` file at the root of the repository.

## Running the Linter

### Using the Lint Script

The easiest way to run the linter is using the provided script:

```bash
# Run on the entire codebase
./scripts/lint.sh

# Run on specific packages
./scripts/lint.sh ./pkg/client/...

# Run with faster settings (fewer linters)
./scripts/lint.sh --fast

# Run with auto-fix for supported issues
./scripts/lint.sh --fix

# Run with a custom timeout
./scripts/lint.sh --timeout=5m
```

### Using the Makefile

For convenience, you can also use the Make target:

```bash
# Run linting on the entire codebase
make lint
```

### Using golangci-lint Directly

If you prefer to use golangci-lint directly:

```bash
# Run with the project configuration
golangci-lint run --config=.golangci.yml
```

## Understanding Linter Output

The linter output includes:

1. **File location**: Path to the file with the issue
2. **Line number**: Line where the issue was found
3. **Linter name**: Which linter detected the issue
4. **Issue description**: What the problem is
5. **Suggested fix**: For some issues, a suggested fix is provided

Example:
```
pkg/client/ecr/client.go:120:6: ineffectual assignment to err (ineffassign)
```

## Fixing Common Linting Issues

### Unused Variables

```go
// Issue
func example() error {
    result, err := someFunction()  // err is unused
    return nil
}

// Fix
func example() error {
    result, err := someFunction()
    if err != nil {
        return err
    }
    _ = result  // Use blank identifier if you need to discard the value
    return nil
}
```

### Missing Error Checks

```go
// Issue
func example() {
    file, _ := os.Open("file.txt")  // Error not checked
    defer file.Close()
}

// Fix
func example() error {
    file, err := os.Open("file.txt")
    if err != nil {
        return err
    }
    defer file.Close()
    return nil
}
```

### Unused Imports

```go
// Issue
import (
    "fmt"
    "time"  // time is imported but not used
)

// Fix - Remove the unused import
import (
    "fmt"
)
```

### Ineffective Assignments

```go
// Issue
func example() {
    err := process()
    err = cleanup()  // Original error is lost
}

// Fix
func example() error {
    if err := process(); err != nil {
        return err
    }
    return cleanup()
}
```

## Ignoring Linting Issues

In rare cases, you may need to ignore a linting issue. This should be used sparingly and with appropriate justification.

### Using Line Comments

```go
// Ignore a specific issue on a line
var example interface{} // nolint:structcheck // Used for reflection

// Ignore multiple issues
func example() { // nolint:unparam,deadcode // Used for testing
    // ...
}
```

### Using Block Comments

```go
//nolint
func legacyCode() {
    // This entire function is excluded from linting
}
```

### Using File-Level Directives

```go
//nolint:unparam // At the top of the file to ignore a specific linter
package example
```

## IDE Integration

### Visual Studio Code

Add the following to your settings.json:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": [
    "--fast"
  ]
}
```

### GoLand

Configure "Go → Go Modules → Preferences → Tools → Golang Linter" to use golangci-lint with the project's configuration file.

### Vim/Neovim

For vim-go plugin, add this to your configuration:

```vim
let g:go_metalinter_command = "golangci-lint"
let g:go_metalinter_autosave = 1
```

## Configured Linters

The project uses the following linters, configured in `.golangci.yml`:

| Linter | Description |
|--------|-------------|
| bodyclose | Checks whether HTTP response bodies are closed |
| deadcode | Finds unused code |
| dupl | Code clone detection |
| errcheck | Checks for unchecked errors |
| gocyclo | Checks function cyclomatic complexity |
| gofmt | Verifies code is properly formatted |
| goimports | Checks import formatting and grouping |
| golint | Enforces Go style |
| gosec | Inspects source code for security issues |
| gosimple | Simplifies code |
| govet | Examines Go source code for suspicious constructs |
| ineffassign | Detects ineffectual assignments |
| misspell | Finds commonly misspelled English words |
| staticcheck | Advanced static analysis |
| structcheck | Finds unused struct fields |
| typecheck | Type checking |
| unconvert | Removes unnecessary conversions |
| unparam | Reports unused function parameters |
| unused | Checks for unused constants, variables, functions, etc. |

## Custom Linter Configuration

See the `.golangci.yml` file for the complete configuration, including:

1. Specific settings for each linter
2. Rules for excluding certain issues
3. Special handling for test files and generated code
4. Performance settings
