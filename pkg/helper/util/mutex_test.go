package util

import (
	"sync"
	"testing"
)

func TestMutexUnlocker(t *testing.T) {
	var mu sync.Mutex
	unlocker := NewMutexUnlocker(&mu)
	if unlocker == nil {
		t.Fatal("Expected non-nil unlocker")
	}

	// Should be locked
	locked := make(chan bool, 1)
	go func() {
		locked <- mu.TryLock()
	}()
	if <-locked {
		mu.Unlock()
		t.Error("Expected mutex to be locked")
	}

	unlocker.Unlock()

	// Should be unlocked now
	if !mu.TryLock() {
		t.Error("Expected mutex to be unlocked")
	}
	mu.Unlock()

	// Multiple unlocks should be safe
	unlocker.Unlock()
	unlocker.Unlock()
}

func TestRWMutexReadUnlocker(t *testing.T) {
	var rwmu sync.RWMutex
	unlocker := NewRWMutexReadUnlocker(&rwmu)
	if unlocker == nil {
		t.Fatal("Expected non-nil unlocker")
	}

	// Should allow another read lock
	rwmu.RLock()
	rwmu.RUnlock()

	unlocker.Unlock()

	// Should allow write lock now
	if !rwmu.TryLock() {
		t.Error("Expected to acquire write lock")
	}
	rwmu.Unlock()

	// Multiple unlocks should be safe
	unlocker.Unlock()
}

func TestRWMutexWriteUnlocker(t *testing.T) {
	var rwmu sync.RWMutex
	unlocker := NewRWMutexWriteUnlocker(&rwmu)
	if unlocker == nil {
		t.Fatal("Expected non-nil unlocker")
	}

	// Should not allow read lock
	locked := make(chan bool, 1)
	go func() {
		locked <- rwmu.TryRLock()
	}()
	if <-locked {
		rwmu.RUnlock()
		t.Error("Expected read lock to fail")
	}

	unlocker.Unlock()

	// Should allow read lock now
	rwmu.RLock()
	rwmu.RUnlock()

	// Multiple unlocks should be safe
	unlocker.Unlock()
}
