# Code Formatting and Linting Standards

This document outlines the automated code formatting and linting standards used in the Freightliner project.

## Overview

Freightliner enforces consistent code formatting and quality through automated tools. These tools are configured to run both locally during development and as part of CI/CD pipelines.

## Tools

The following tools are used for code formatting and linting:

1. **gofmt**: Standard Go code formatter
2. **goimports**: Adds import management to gofmt
3. **golangci-lint**: Multi-linter tool that runs multiple static analysis tools

## Configuration

### golangci-lint

The project uses a `.golangci.yml` configuration file at the root of the repository. This configuration:

- Enables specific linters appropriate for our project
- Configures linter-specific settings
- Excludes certain files (e.g., generated code, test files) from some checks
- Sets reasonable timeouts and performance settings

### CI/CD Integration

The formatting and linting checks are integrated into our CI/CD pipeline via GitHub Actions:

1. The workflow runs on all pull requests and pushes to main branches
2. It verifies that code is properly formatted with gofmt
3. It checks that imports are properly organized with goimports
4. It runs golangci-lint to perform comprehensive static analysis

## Running Locally

### Prerequisites

Install the required tools:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install goimports
go install golang.org/x/tools/cmd/goimports@latest
```

### Using Make Targets

The Makefile provides convenient targets for running the formatting and linting tools:

```bash
# Format code with gofmt
make fmt

# Organize imports
make imports

# Run linters
make lint

# Run all quality checks (formatting, imports, linting, and tests)
make check
```

### Pre-commit Hooks

Git pre-commit hooks are configured to automatically check formatting and linting before committing changes:

```bash
# Install the pre-commit hooks
make hooks
```

## Editor Integration

### Visual Studio Code

Add the following to your settings.json:

```json
{
  "go.formatTool": "goimports",
  "go.formatFlags": ["-local", "freightliner"],
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "editor.formatOnSave": true
}
```

### GoLand

1. Enable "Go > Go Modules > Preferences > Tools > File Watchers" to use goimports and gofmt
2. Configure "Go > Go Modules > Preferences > Tools > Golang Linter" to use golangci-lint

### Vim/Neovim

For vim-go plugin, add the following to your configuration:

```vim
let g:go_fmt_command = "goimports"
let g:go_fmt_options = "-local freightliner"
let g:go_metalinter_command = "golangci-lint"
let g:go_metalinter_enabled = ['vet', 'golint', 'errcheck']
let g:go_metalinter_autosave = 1
```

## Linter Rules

The following key linters are enabled:

1. **gofmt/goimports**: Ensures consistent code formatting and import organization
2. **golint/revive**: Enforces Go style conventions
3. **govet**: Checks for common errors and suspicious constructs
4. **errcheck**: Ensures errors are properly checked
5. **gosec**: Identifies security issues
6. **staticcheck**: Performs extensive static analysis
7. **unused**: Identifies unused code

## Handling Linter Warnings

When a linter reports an issue:

1. **Fix the issue**: Most issues should be fixed to maintain code quality
2. **Justify exceptions**: If a linter warning must be ignored:
   - Add a `//nolint` comment with a justification
   - Example: `//nolint:errcheck // Intentionally ignoring this error as it cannot happen in this context`

3. **Update configuration**: For systemic exceptions, update the `.golangci.yml` file

## Resources

- [golangci-lint Documentation](https://golangci-lint.run/)
- [Effective Go - Formatting](https://golang.org/doc/effective_go#formatting)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
