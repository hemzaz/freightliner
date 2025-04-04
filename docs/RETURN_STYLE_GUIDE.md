# Standardized Return Style Guide

This document outlines the standard pattern for function returns in the Freightliner codebase.

## Direct Returns

The Freightliner codebase uses **direct returns** rather than named returns. This promotes clarity and readability by making the control flow explicit.

### Preferred Style (Direct Returns)

```go
// Correct: Using direct returns
func ParseRegistryPath(path string) (string, string, error) {
    parts := strings.SplitN(path, "/", 2)
    if len(parts) != 2 {
        return "", "", errors.InvalidInputf("invalid format. Use [registry]/[repository]")
    }
    return parts[0], parts[1], nil
}
```

### Avoid (Named Returns)

```go
// Avoid: Using named returns
func ParseRegistryPath(path string) (registry, repo string, err error) {
    parts := strings.SplitN(path, "/", 2)
    if len(parts) != 2 {
        return "", "", errors.InvalidInputf("invalid format. Use [registry]/[repository]")
    }
    return parts[0], parts[1], nil
}
```

## Guidelines

1. **Always use direct returns** instead of named returns.
2. **Be explicit about all return values** at each return statement.
3. **Document return values** in function comments when the purpose isn't obvious from the function name.
4. **Use descriptive variable names** in the function signature to make the purpose of each return value clear.

## Exception

The only exception to this rule is when dealing with very complex functions where named returns significantly improve readability or aid in defer patterns. Such exceptions should be rare and must be documented with a comment explaining why named returns were chosen.

## Examples

### Simple Functions

```go
// Good
func IsValid(input string) bool {
    return len(input) > 0
}

// Good
func GetConfig(name string) (*Config, error) {
    if name == "" {
        return nil, errors.InvalidInputf("name cannot be empty")
    }
    // ...
    return config, nil
}
```

### Functions with Multiple Returns

```go
// Good
func SplitPath(path string) (string, string, error) {
    if path == "" {
        return "", "", errors.InvalidInputf("path cannot be empty")
    }
    // ...
    return dir, file, nil
}
```
