package util

import (
	"sync"
)

// MutexUnlocker is a helper for deferred mutex unlocking
// It ensures that mutexes are unlocked even in error paths
type MutexUnlocker struct {
	mutex *sync.Mutex
	done  bool
}

// NewMutexUnlocker creates a new unlocker and locks the mutex
func NewMutexUnlocker(mutex *sync.Mutex) *MutexUnlocker {
	mutex.Lock()
	return &MutexUnlocker{
		mutex: mutex,
		done:  false,
	}
}

// Unlock unlocks the mutex if it hasn't already been unlocked
func (u *MutexUnlocker) Unlock() {
	if !u.done {
		u.mutex.Unlock()
		u.done = true
	}
}

// RWMutexReadUnlocker is a helper for deferred RWMutex read unlocking
type RWMutexReadUnlocker struct {
	mutex *sync.RWMutex
	done  bool
}

// NewRWMutexReadUnlocker creates a new read unlocker and locks the mutex for reading
func NewRWMutexReadUnlocker(mutex *sync.RWMutex) *RWMutexReadUnlocker {
	mutex.RLock()
	return &RWMutexReadUnlocker{
		mutex: mutex,
		done:  false,
	}
}

// Unlock unlocks the mutex if it hasn't already been unlocked
func (u *RWMutexReadUnlocker) Unlock() {
	if !u.done {
		u.mutex.RUnlock()
		u.done = true
	}
}

// RWMutexWriteUnlocker is a helper for deferred RWMutex write unlocking
type RWMutexWriteUnlocker struct {
	mutex *sync.RWMutex
	done  bool
}

// NewRWMutexWriteUnlocker creates a new write unlocker and locks the mutex for writing
func NewRWMutexWriteUnlocker(mutex *sync.RWMutex) *RWMutexWriteUnlocker {
	mutex.Lock()
	return &RWMutexWriteUnlocker{
		mutex: mutex,
		done:  false,
	}
}

// Unlock unlocks the mutex if it hasn't already been unlocked
func (u *RWMutexWriteUnlocker) Unlock() {
	if !u.done {
		u.mutex.Unlock()
		u.done = true
	}
}
