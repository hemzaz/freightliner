# Concurrency Patterns in Freightliner

This document outlines the concurrency patterns and best practices used in the Freightliner codebase to ensure thread safety and prevent deadlocks or race conditions.

## Mutex Usage Guidelines

### 1. Input Validation Before Lock Acquisition

To prevent potential deadlocks, always validate input parameters before acquiring locks:

```go
// Bad pattern
func (s *Store) GetItem(id string) (*Item, error) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    if id == "" {
        return nil, errors.New("id cannot be empty")
    }
    
    // ...rest of the function
}

// Good pattern
func (s *Store) GetItem(id string) (*Item, error) {
    // Validate input before locking to fail fast
    if id == "" {
        return nil, errors.New("id cannot be empty")
    }
    
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    // ...rest of the function
}
```

### 2. Using Safe Unlock Helpers

For complex functions with multiple return paths, use the mutex helper utilities in `pkg/helper/util/mutex_helpers.go`:

```go
func (s *Store) ComplexOperation() error {
    // Create a mutex unlocker
    unlocker := util.NewMutexUnlocker(&s.mutex)
    defer unlocker.Unlock()
    
    // Function with multiple return paths
    if condition1 {
        return errors.New("condition 1 failed")
    }
    
    if condition2 {
        return errors.New("condition 2 failed")
    }
    
    // Do work...
    return nil
}
```

### 3. Lock Granularity

Use fine-grained locking to minimize contention:

```go
// Bad pattern - locking entire function
func (s *Service) Process() {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    // Long operation A
    // Long operation B
    // Long operation C
}

// Good pattern - lock only critical sections
func (s *Service) Process() {
    // Do non-critical work
    resultA := s.operationA()
    
    // Lock only when needed
    s.mutex.Lock()
    s.sharedState = resultA
    s.mutex.Unlock()
    
    // More non-critical work
    resultB := s.operationB(resultA)
    
    // Lock again only when needed
    s.mutex.Lock()
    s.sharedState.Append(resultB)
    s.mutex.Unlock()
}
```

### 4. Read-Write Mutex Usage

Prefer `sync.RWMutex` when reads are more common than writes:

```go
type Cache struct {
    data map[string]interface{}
    mu   sync.RWMutex
}

// Multiple concurrent readers are allowed
func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    value, exists := c.data[key]
    return value, exists
}

// Writers get exclusive access
func (c *Cache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.data[key] = value
}
```

### 5. Avoiding Nested Locks

Avoid acquiring locks while already holding another lock:

```go
// Dangerous pattern - can cause deadlocks
func (s *Service) TransferBetween(from, to *Resource) error {
    from.mutex.Lock()
    defer from.mutex.Unlock()
    
    to.mutex.Lock() // <-- Potential deadlock if another thread locks in reverse order
    defer to.mutex.Unlock()
    
    // Transfer logic
    return nil
}

// Better pattern - consistent lock ordering
func (s *Service) TransferBetween(from, to *Resource) error {
    // Sort the resources by ID to ensure consistent lock order
    first, second := orderResources(from, to)
    
    first.mutex.Lock()
    defer first.mutex.Unlock()
    
    second.mutex.Lock()
    defer second.mutex.Unlock()
    
    // Transfer logic
    return nil
}
```

## Worker Pool Pattern

The worker pool pattern is used for executing concurrent tasks with controlled parallelism:

```go
type WorkerPool struct {
    tasks   chan Task
    results chan Result
    wg      sync.WaitGroup
    ctx     context.Context
    cancel  context.CancelFunc
}

func (wp *WorkerPool) Start(numWorkers int) {
    for i := 0; i < numWorkers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()
    
    for {
        select {
        case task, ok := <-wp.tasks:
            if !ok {
                return
            }
            
            result := task.Execute()
            wp.results <- result
            
        case <-wp.ctx.Done():
            return
        }
    }
}
```

## Context Usage for Cancellation

Use context for propagating cancellation signals:

```go
func (s *Service) ProcessWithTimeout(timeout time.Duration) error {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    return s.processWithContext(ctx)
}

func (s *Service) processWithContext(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue processing
    }
    
    // Check context periodically during long operations
    for i := 0; i < iterations; i++ {
        if ctx.Err() != nil {
            return ctx.Err()
        }
        
        // Do work...
    }
    
    return nil
}
```

## Recommended Practices

1. **Validate inputs before acquiring locks**
2. **Use defer for unlocking mutexes**
3. **Keep critical sections small**
4. **Use RWMutex when reads outnumber writes**
5. **Follow consistent lock ordering to prevent deadlocks**
6. **Use context for propagating cancellation**
7. **Consider using higher-level synchronization primitives when appropriate**
8. **Document thread safety guarantees for packages and types**