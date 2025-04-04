# Method Organization Guidelines

This document outlines the standard pattern for organizing methods within types in the Freightliner codebase.

## Overview

Consistent organization of methods within types improves code readability, maintainability, and makes it easier for developers to navigate and understand the codebase. The following guidelines should be followed when organizing methods within types.

## Method Ordering

Methods within a type should be organized in the following order:

1. **Constructors and Factory Methods**
   - Functions that create new instances of the type (e.g., `New*`, `Create*`)
   - Static factory methods

2. **Core Interface Methods**
   - Methods that implement the primary interface of the type
   - Methods are ordered based on their logical sequence of use

3. **Public Methods**
   - Other public methods not part of the core interface
   - Grouped by related functionality
   - Arranged alphabetically within each group

4. **Internal/Unexported Methods**
   - Helper methods used internally by the type
   - Utility methods specific to the type
   - Arranged alphabetically within logical groups

5. **Overridden Methods**
   - Methods that override behavior from embedded types
   - Implementation of standard interfaces (e.g., `fmt.Stringer`, `json.Marshaler`)

## Example

```go
// Repository represents a container registry repository
type Repository struct {
    // fields...
}

// ==========================================
// Constructors and Factory Methods
// ==========================================

// NewRepository creates a new repository instance
func NewRepository(name string) *Repository {
    // implementation...
}

// ==========================================
// Core Interface Methods (e.g., common.Repository interface)
// ==========================================

// GetName returns the repository name
func (r *Repository) GetName() string {
    // implementation...
}

// ListTags lists all tags in the repository
func (r *Repository) ListTags(ctx context.Context) ([]string, error) {
    // implementation...
}

// ==========================================
// Public Methods (not part of core interface)
// ==========================================

// DeleteTag removes a tag from the repository
func (r *Repository) DeleteTag(ctx context.Context, tag string) error {
    // implementation...
}

// UpdateMetadata updates the repository metadata
func (r *Repository) UpdateMetadata(ctx context.Context, metadata map[string]string) error {
    // implementation...
}

// ==========================================
// Internal/Unexported Methods
// ==========================================

// buildTagReference creates a reference for a tag
func (r *Repository) buildTagReference(tag string) (name.Tag, error) {
    // implementation...
}

// validateTag checks if a tag name is valid
func (r *Repository) validateTag(tag string) error {
    // implementation...
}

// ==========================================
// Overridden Methods
// ==========================================

// String implements the fmt.Stringer interface
func (r *Repository) String() string {
    // implementation...
}
```

## Code Comments

Types and methods should have accompanying documentation comments following Go's standard documentation patterns:

```go
// Repository represents a container registry repository.
// It provides methods for managing images and tags.
type Repository struct {
    // fields...
}

// GetName returns the repository name.
func (r *Repository) GetName() string {
    // implementation...
}
```

## Interface Implementations

For clarity, it's recommended to add comments indicating which methods are part of interface implementations:

```go
// ListTags lists all tags in the repository.
// Implements common.Repository.ListTags.
func (r *Repository) ListTags(ctx context.Context) ([]string, error) {
    // implementation...
}
```

## File Organization

In addition to method organization within types, files should have a consistent organization:

1. Package declaration and package documentation
2. Imports (grouped by standard library, third-party, and internal)
3. Constants and variables
4. Interface definitions
5. Type definitions
6. Global functions
7. Methods for each type (following the method organization guidelines)

## Guidelines Application

These guidelines should be applied to new code and during refactoring of existing code. While it may not be practical to reorganize all existing code immediately, aim to follow these patterns in new code and when making substantial changes to existing types.
