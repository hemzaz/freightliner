# Go Vet Guidelines

This document explains how to use the `go vet` tools configured for the Freightliner project.

## Overview

Go vet examines Go source code and reports suspicious constructs, such as:
- Printf calls whose arguments do not align with the format string
- Method signatures that differ from the interface type declarations
- Unreachable code
- Suspicious assignments
- Misuse of unsafe pointers
- And many other common mistakes

The Freightliner project has enhanced the standard `go vet` with additional checks to catch more potential issues.

## Running Go Vet

### Using the Vet Script

The easiest way to run the checks is using the provided script:

```bash
# Run on the entire codebase
./scripts/vet.sh

# Run on specific packages
./scripts/vet.sh ./pkg/client/...

# Run with verbose output
./scripts/vet.sh --verbose

# Run showing all diagnostics (not just errors)
./scripts/vet.sh --all
```

### Using the Makefile

For convenience, you can also use the Make target:

```bash
# Run go vet on the entire codebase
make vet
```

### Using Go Vet Directly

If you prefer to use go vet directly:

```bash
# Run standard go vet
go vet ./...

# Run with shadow variable checking
go vet -vettool=$(which shadow) ./...

# Run with interface checking
go vet -vettool=$(which interfacer) ./...
```

## Understanding Go Vet Output

Go vet output includes:

1. **Package location**: Path to the package with the issue
2. **File name**: File containing the issue
3. **Line number**: Line where the issue was found
4. **Issue description**: What the problem is

Example:
```
pkg/client/ecr/client.go:120:6: this value of err is never used
```

## Common Issues Detected by Go Vet

### Printf Format Strings

```go
// Issue
fmt.Printf("Value: %s", 42)  // %s used with int value

// Fix
fmt.Printf("Value: %d", 42)  // Correct format specifier
```

### Unreachable Code

```go
// Issue
func example() {
    return
    fmt.Println("This will never execute")  // Unreachable
}

// Fix
func example() {
    fmt.Println("This will execute")
    return
}
```

### Interface Implementation Errors

```go
// Issue
type ReadCloser interface {
    Read(p []byte) (n int, err error)
    Close() error
}

type MyReader struct{}

func (r *MyReader) Read(data []byte) (n int, err error) {
    // Implementation
    return 0, nil
}

func (r *MyReader) Close() {  // Missing error return value
    // Implementation
}

// Fix
func (r *MyReader) Close() error {  // Correct signature
    // Implementation
    return nil
}
```

### Composite Literals

```go
// Issue
type Point struct {
    X, Y int
}

p := Point{1, 2, 3}  // Too many values

// Fix
p := Point{1, 2}
// or
p := Point{X: 1, Y: 2}
```

### Shadow Variables

```go
// Issue
func example() {
    x := 1
    if true {
        x := 2  // Shadows outer x
        fmt.Println(x)
    }
    fmt.Println(x)  // Still 1, might be unexpected
}

// Fix
func example() {
    x := 1
    if true {
        y := 2  // Use a different name
        fmt.Println(y)
    }
    fmt.Println(x)
}
```

## Additional Checks

The Freightliner `go vet` setup includes additional checks beyond the standard ones:

1. **Shadow variable detection**: Identifies variables that may inadvertently shadow other variables
2. **Interface checking**: Suggests interfaces when method sets match

## Fixing Go Vet Issues

Most go vet issues are actual bugs or potential problems that should be fixed. Here are strategies for addressing common issues:

1. **Format string issues**: Ensure format specifiers match argument types
2. **Interface issues**: Make sure method signatures exactly match the interface definitions
3. **Unreachable code**: Remove or make the code reachable
4. **Shadow variables**: Rename variables to avoid shadowing
5. **Suspicious assignments**: Verify the logic and fix any mistakes

## IDE Integration

### Visual Studio Code

VS Code's Go extension runs `go vet` by default when files are saved.

### GoLand

GoLand integrates `go vet` into its inspection system. Enable it in Settings → Go → Inspections.

### Vim/Neovim

For vim-go plugin, add this to your configuration:

```vim
let g:go_metalinter_autosave_enabled = ['vet']
let g:go_metalinter_enabled = ['vet']
```

## CI Integration

Go vet is integrated into our CI pipeline to catch issues early:

1. Standard `go vet` runs on all code
2. Shadow checking catches variable shadowing issues
3. Interface checking suggests interface optimizations

A pull request will fail if any of these checks identify issues, ensuring code quality before merge.
