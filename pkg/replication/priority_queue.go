package replication

import (
	"container/heap"
	"context"
	"sync"

	"freightliner/pkg/helper/log"
)

// Priority levels for jobs
const (
	PriorityHigh   = 1
	PriorityMedium = 2
	PriorityLow    = 3
)

// PriorityJob wraps a WorkerJob with priority
type PriorityJob struct {
	Job      WorkerJob
	Priority int
	Index    int // Index in the heap
}

// PriorityQueue implements heap.Interface for priority-based job scheduling
type PriorityQueue struct {
	items []*PriorityJob
	mu    sync.RWMutex
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		items: make([]*PriorityJob, 0),
	}
	heap.Init(pq)
	return pq
}

// Len returns the number of items in the queue
func (pq *PriorityQueue) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items)
}

// Less compares two items (lower priority value = higher priority)
func (pq *PriorityQueue) Less(i, j int) bool {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return pq.items[i].Priority < pq.items[j].Priority
}

// Swap swaps two items in the queue
func (pq *PriorityQueue) Swap(i, j int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].Index = i
	pq.items[j].Index = j
}

// Push adds an item to the queue
func (pq *PriorityQueue) Push(x interface{}) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	n := len(pq.items)
	item := x.(*PriorityJob)
	item.Index = n
	pq.items = append(pq.items, item)
}

// Pop removes and returns the highest priority item
func (pq *PriorityQueue) Pop() interface{} {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	pq.items = old[0 : n-1]
	return item
}

// Enqueue adds a job to the priority queue
func (pq *PriorityQueue) Enqueue(job WorkerJob, priority int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	item := &PriorityJob{
		Job:      job,
		Priority: priority,
	}
	heap.Push(pq, item)
}

// Dequeue removes and returns the highest priority job
func (pq *PriorityQueue) Dequeue() *WorkerJob {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return nil
	}

	item := heap.Pop(pq).(*PriorityJob)
	return &item.Job
}

// Peek returns the highest priority job without removing it
func (pq *PriorityQueue) Peek() *WorkerJob {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	if len(pq.items) == 0 {
		return nil
	}

	return &pq.items[0].Job
}

// IsEmpty returns true if the queue is empty
func (pq *PriorityQueue) IsEmpty() bool {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items) == 0
}

// Clear removes all jobs from the queue
func (pq *PriorityQueue) Clear() {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.items = make([]*PriorityJob, 0)
	heap.Init(pq)
}

// PriorityWorkerPool extends WorkerPool with priority-based scheduling
type PriorityWorkerPool struct {
	*WorkerPool
	priorityQueue *PriorityQueue
	queueMu       sync.Mutex
}

// NewPriorityWorkerPool creates a new worker pool with priority support
func NewPriorityWorkerPool(workerCount int, logger log.Logger) *PriorityWorkerPool {
	basePool := NewWorkerPool(workerCount, logger)

	return &PriorityWorkerPool{
		WorkerPool:    basePool,
		priorityQueue: NewPriorityQueue(),
	}
}

// SubmitWithPriority submits a job with a specific priority
func (p *PriorityWorkerPool) SubmitWithPriority(jobID string, task TaskFunc, priority int) error {
	// Create a job
	job := WorkerJob{
		ID:       jobID,
		Task:     task,
		Priority: priority,
		Context:  context.Background(),
	}

	// Add to priority queue
	p.priorityQueue.Enqueue(job, priority)

	// Process the queue in a separate goroutine
	go p.processQueue()

	return nil
}

// processQueue processes jobs from the priority queue
func (p *PriorityWorkerPool) processQueue() {
	p.queueMu.Lock()
	defer p.queueMu.Unlock()

	// Dequeue the highest priority job
	job := p.priorityQueue.Dequeue()
	if job == nil {
		return
	}

	// Submit to the underlying worker pool
	select {
	case p.jobQueue <- *job:
		// Job submitted successfully
	default:
		// Queue is full, re-enqueue the job
		p.priorityQueue.Enqueue(*job, job.Priority)
	}
}

// GetQueueStats returns statistics about the priority queue
func (p *PriorityWorkerPool) GetQueueStats() map[string]int {
	p.queueMu.Lock()
	defer p.queueMu.Unlock()

	stats := map[string]int{
		"total":         p.priorityQueue.Len(),
		"high_priority": 0,
		"med_priority":  0,
		"low_priority":  0,
	}

	// Count jobs by priority
	p.priorityQueue.mu.RLock()
	defer p.priorityQueue.mu.RUnlock()

	for _, item := range p.priorityQueue.items {
		switch item.Priority {
		case PriorityHigh:
			stats["high_priority"]++
		case PriorityMedium:
			stats["med_priority"]++
		case PriorityLow:
			stats["low_priority"]++
		}
	}

	return stats
}
