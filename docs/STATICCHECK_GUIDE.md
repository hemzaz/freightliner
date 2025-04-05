# Staticcheck Guidelines

This document explains how to use the staticcheck static analysis tool configured for the Freightliner project.

## Overview

[Staticcheck](https://staticcheck.io/) is a state-of-the-art static analysis tool for Go code. It finds bugs, performance issues, suspicious constructs, and enforces style rules. It goes beyond the capabilities of `go vet` by applying hundreds of additional checks.

The Freightliner project has configured staticcheck for optimal use with our codebase.

## Running Staticcheck

### Using the Staticcheck Script

The easiest way to run staticcheck is using the provided script:

```bash
# Run on the entire codebase
./scripts/staticcheck.sh

# Run on specific packages
./scripts/staticcheck.sh ./pkg/client/...

# Run with verbose output (includes explanations)
./scripts/staticcheck.sh --verbose

# Run with specific checks
./scripts/staticcheck.sh --checks=ST1000,ST1005
```

### Using the Makefile

For convenience, you can also use the Make target:

```bash
# Run staticcheck on the entire codebase
make staticcheck
```

### Using Staticcheck Directly

If you prefer to use staticcheck directly:

```bash
# Run staticcheck with project configuration
staticcheck -f=text ./...

# Run with detailed explanations
staticcheck -f=text -explain ./...
```

## Understanding Staticcheck Output

Staticcheck output includes:

1. **File location**: Path to the file with the issue
2. **Line number**: Line where the issue was found
3. **Check ID**: Identifier for the specific check (e.g., ST1005)
4. **Issue description**: What the problem is

Example:
```
pkg/client/ecr/client.go:120:6: this value of err is never used (SA4006)
```

## Staticcheck Check Categories

Staticcheck groups checks into several categories:

1. **S1xxx: Code correctness** - Bugs and incorrect behavior
2. **S2xxx: Performance** - Inefficient code constructs
3. **S3xxx: Style** - Code style issues
4. **S4xxx: Simplifications** - Code that can be simplified
5. **S5xxx: Edge cases** - Uncommon issues and pitfalls
6. **S6xxx: Unused code** - Dead code and other unused elements 
7. **ST1xxx: Standard library usage** - Common issues with the Go standard library
8. **SA1xxx-SA9xxx: Various static analysis checks** - A range of issues

## Common Issues Detected by Staticcheck

### Unused Code (SA4xxx)

```go
// Issue
func example() {
    result, err := someFunc()  // err is defined but not used
    fmt.Println(result)
}

// Fix
func example() {
    result, _ := someFunc()  // Use _ to explicitly ignore the error
    fmt.Println(result)
    // Or properly handle the error:
    // result, err := someFunc()
    // if err != nil {
    //     return fmt.Errorf("failed to get result: %w", err)
    // }
}
```

### Incorrect Error Checking (ST1005)

```go
// Issue
if err != nil {
    return fmt.Errorf("Something failed: %s", err)  // Error strings shouldn't be capitalized
}

// Fix
if err != nil {
    return fmt.Errorf("something failed: %w", err)  // Lowercase first letter and use %w
}
```

### Incorrect Context Usage (SA1012, SA1019)

```go
// Issue
ctx := context.TODO()  // Using the wrong context type
http.Get("https://example.com")  // Not using context with HTTP requests

// Fix
ctx := context.Background()
req, err := http.NewRequestWithContext(ctx, "GET", "https://example.com", nil)
```

### Redundant Code (S1xxx)

```go
// Issue
if x == true {  // Redundant comparison
    return "yes"
}

// Fix
if x {
    return "yes"
}
```

## Configuration

The project uses a `staticcheck.conf` file to configure the tool's behavior. This file:

1. Enables all checks by default
2. Disables specific checks that aren't appropriate for our codebase
3. Configures initialisms that should be all-uppercase
4. Sets up file exclusion patterns

## Suppressing Issues

In rare cases, you may need to suppress a staticcheck issue:

### Line-Level Suppression

```go
//lint:ignore SA4006 This error variable is intentionally unused
_, err := someFunc()
```

### File-Level Suppression

```go
//lint:file-ignore U1000 This file contains intentionally unused code
package example
```

### Function-Level Suppression

```go
//lint:ignore SA1019 We need to use a deprecated API for backward compatibility
func legacyCallUsingDeprecatedAPI() {
    // ...
}
```

## IDE Integration

### Visual Studio Code

VS Code's Go extension supports staticcheck.

Add to your settings.json:
```json
{
  "go.lintTool": "staticcheck",
  "go.lintFlags": ["-checks=all"]
}
```

### GoLand

Configure "Go → Go Modules → Preferences → Tools → File Watchers" to use staticcheck with the project's configuration file.

### Vim/Neovim

For vim-go plugin:

```vim
let g:go_metalinter_enabled = ['staticcheck']
let g:go_metalinter_command = "staticcheck"
```

## CI Integration

Staticcheck is integrated into our CI pipeline:

1. CI runs staticcheck on all code
2. Failed checks will cause the build to fail
3. The staticcheck.conf file ensures consistent configuration

This ensures code quality is maintained across the project.

## Resources

1. [Staticcheck Documentation](https://staticcheck.io/docs/)
2. [Available Checks](https://staticcheck.io/docs/checks)
3. [Configuration Documentation](https://staticcheck.io/docs/configuration)
