# Import Organization Guidelines

This document outlines the standard pattern for organizing imports in Go files in the Freightliner codebase.

## Overview

Consistent organization of imports improves code readability and maintainability. The Freightliner project uses the `goimports` tool to automatically organize imports according to Go conventions.

## Import Groups

Imports should be organized into the following groups, separated by a blank line:

1. **Standard Library Packages**
   - Built-in Go packages (e.g., `fmt`, `context`, `io`)

2. **Third-Party Packages**
   - External dependencies (e.g., `github.com/aws/aws-sdk-go-v2`, `github.com/spf13/cobra`)

3. **Internal Project Packages**
   - Freightliner packages (e.g., `freightliner/pkg/client`, `freightliner/pkg/config`)

## Example

```go
package example

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"

	"freightliner/pkg/client/common"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"
)
```

## Automated Organization

The Freightliner project includes several tools to automatically organize imports:

1. **Script**: `scripts/organize_imports.sh`
   - Run this script to organize imports across the entire codebase
   - Usage: `./scripts/organize_imports.sh` or `make imports`

2. **Git Pre-Commit Hook**: `.git/hooks/pre-commit`
   - Automatically organizes imports when committing changes
   - Install with: `make hooks`

3. **Make Target**: `make imports`
   - Runs the organize_imports.sh script

## Configuration

The import organization is configured by:

1. `.goimportsignore` file: Specifies patterns for files that should be ignored by goimports
2. `scripts/organize_imports.sh`: Configures goimports with the `-local freightliner` flag, which ensures project imports are grouped together

## Developer Workflow

Developers should follow these steps to ensure consistent import organization:

1. Install the required tools:
   ```
   make setup
   ```

2. Install the Git pre-commit hook:
   ```
   make hooks
   ```

3. Run import organization manually when needed:
   ```
   make imports
   ```

4. Run all quality checks before submitting changes:
   ```
   make check
   ```

## IDE Integration

Most major IDEs support automatic import organization with goimports:

### Visual Studio Code
Add the following to your settings.json:
```json
{
  "go.formatTool": "goimports",
  "go.formatFlags": [
    "-local", "freightliner"
  ],
  "editor.formatOnSave": true
}
```

### GoLand
Configure "Go > Go Modules > Preferences > Tools > File Watchers" to use goimports with the `-local freightliner` flag.

### Vim/Neovim
Configure vim-go plugin with:
```vim
let g:go_fmt_command = "goimports"
let g:go_fmt_options = "-local freightliner"
```
